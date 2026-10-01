import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { routes } from '../../app.routes';
import { AvailabilityService } from '../services/availability.service';
import { SessionService } from '../services/session.service';
import { apiInterceptor } from './api.interceptor';

describe('apiInterceptor', () => {
  let http: HttpClient;
  let control: HttpTestingController;
  let sesion: SessionService;
  let disponibilidad: AvailabilityService;
  let router: Router;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([apiInterceptor])),
        provideHttpClientTesting(),
        provideRouter(routes),
      ],
    });

    http = TestBed.inject(HttpClient);
    control = TestBed.inject(HttpTestingController);
    sesion = TestBed.inject(SessionService);
    disponibilidad = TestBed.inject(AvailabilityService);
    router = TestBed.inject(Router);
  });

  afterEach(() => {
    control.verify();
    localStorage.clear();
  });

  it('pone la cabecera de la sesión cuando hay token', async () => {
    localStorage.setItem('catalina-support.token', 'un-token');

    const llamada = firstValueFrom(http.get('/api/users'));
    const peticion = control.expectOne('/api/users');
    expect(peticion.request.headers.get('Authorization')).toBe('Bearer un-token');
    peticion.flush({});
    await llamada;
  });

  it('no pone ninguna cabecera si no hay sesión', async () => {
    const llamada = firstValueFrom(http.get('/api/health'));
    const peticion = control.expectOne('/api/health');
    expect(peticion.request.headers.has('Authorization')).toBe(false);
    peticion.flush({});
    await llamada;
  });

  it('con un 401 de una petición con sesión, la olvida y lleva a la entrada', async () => {
    localStorage.setItem('catalina-support.token', 'caducado');
    const navegar = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    const llamada = firstValueFrom(http.get('/api/auth/me'));
    control
      .expectOne('/api/auth/me')
      .flush({ error: 'auth.session.invalid' }, { status: 401, statusText: 'Unauthorized' });
    await expect(llamada).rejects.toBeTruthy();

    expect(sesion.hayToken()).toBe(false);
    expect(sesion.sesionCaducada()).toBe(true);
    expect(navegar).toHaveBeenCalledWith(['/login']);
  });

  it('con un 401 al entrar, no borra nada ni lleva a ninguna parte', async () => {
    const navegar = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    // Entrar no lleva token: el 401 es «esas credenciales no valen», y eso lo cuenta la pantalla.
    const llamada = firstValueFrom(
      http.post('/api/auth/login', { email: 'a@b.com', password: 'x' }),
    );
    control
      .expectOne('/api/auth/login')
      .flush({ error: 'auth.invalidCredentials' }, { status: 401, statusText: 'Unauthorized' });
    await expect(llamada).rejects.toBeTruthy();

    expect(navegar).not.toHaveBeenCalled();
    expect(sesion.sesionCaducada()).toBe(false);
  });

  it('un 401 al establecer la contraseña no echa a nadie de su sesión', async () => {
    localStorage.setItem('catalina-support.token', 'una-sesion-viva');
    const navegar = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    // Alguien con sesión abre un enlace del correo que ya se usó: lo que ha caducado es el enlace,
    // no su sesión, y la pantalla tiene que poder contarlo.
    const llamada = firstValueFrom(
      http.post('/api/auth/password/reset', { token: 'gastado', password: 'una-contraseña-larga' }),
    );
    control
      .expectOne('/api/auth/password/reset')
      .flush({ error: 'auth.token.used' }, { status: 401, statusText: 'Unauthorized' });
    await expect(llamada).rejects.toBeTruthy();

    expect(sesion.hayToken()).toBe(true);
    expect(navegar).not.toHaveBeenCalled();
  });

  it('un 401 al entrar tampoco, aunque haya sesión de otra cuenta', async () => {
    localStorage.setItem('catalina-support.token', 'una-sesion-viva');
    const navegar = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    const llamada = firstValueFrom(
      http.post('/api/auth/password/forgot', { email: 'quien@sea.com' }),
    );
    control
      .expectOne('/api/auth/password/forgot')
      .flush({ error: 'error interno' }, { status: 401, statusText: 'Unauthorized' });
    await expect(llamada).rejects.toBeTruthy();

    expect(sesion.hayToken()).toBe(true);
    expect(navegar).not.toHaveBeenCalled();
  });

  it('marca «servidor caído» cuando no hay respuesta o el servidor no puede atender', async () => {
    const primera = firstValueFrom(http.get('/api/health'));
    control.expectOne('/api/health').error(new ProgressEvent('error'));
    await expect(primera).rejects.toBeTruthy();
    expect(disponibilidad.sinServidor()).toBe(true);

    // Y se quita en cuanto el backend vuelve a contestar.
    const segunda = firstValueFrom(http.get('/api/health'));
    control.expectOne('/api/health').flush({ status: 'ok', database: 'ok' });
    await segunda;
    expect(disponibilidad.sinServidor()).toBe(false);
  });
});
