import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { TranslationService } from '../i18n/translation.service';
import { SessionService } from './session.service';

const USUARIO = {
  id: 7,
  name: 'Ana',
  lastName: 'Pérez',
  email: 'ana@ejemplo.com',
  role: 'usuario',
  origin: 'local',
  factory: false,
};

describe('SessionService', () => {
  let sesion: SessionService;
  let http: HttpTestingController;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    sesion = TestBed.inject(SessionService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    localStorage.clear();
  });

  it('guarda el token al entrar y se acuerda de quién es', async () => {
    const entrada = sesion.entrar('ana@ejemplo.com', 'una-contraseña-larga');

    const peticion = http.expectOne('/api/auth/login');
    expect(peticion.request.body).toEqual({
      email: 'ana@ejemplo.com',
      password: 'una-contraseña-larga',
    });
    peticion.flush({ token: 'un-token', expiresAt: '2026-09-24T04:00:00Z', user: USUARIO });
    await entrada;

    expect(sesion.hayToken()).toBe(true);
    expect(localStorage.getItem('catalina-support.token')).toBe('un-token');
    expect(sesion.usuario()?.name).toBe('Ana');
    // Entrar no cambia el idioma global por el valor antiguo de una cuenta.
    expect(TestBed.inject(TranslationService).idioma()).toBe('en');
  });

  it('no deja token si la entrada falla', async () => {
    const entrada = sesion.entrar('ana@ejemplo.com', 'mal');
    http
      .expectOne('/api/auth/login')
      .flush({ error: 'auth.invalidCredentials' }, { status: 401, statusText: 'Unauthorized' });

    await expect(entrada).rejects.toBeTruthy();
    expect(sesion.hayToken()).toBe(false);
    expect(sesion.usuario()).toBeNull();
  });

  it('al salir borra la sesión y avisa al backend', async () => {
    localStorage.setItem('catalina-support.token', 'un-token');

    const salida = sesion.salir();
    http.expectOne('/api/auth/logout').flush({ status: 'ok' });
    await salida;

    expect(sesion.hayToken()).toBe(false);
    expect(sesion.usuario()).toBeNull();
    // Salir a propósito no es «se te ha caducado la sesión»: no hay nada que explicar.
    expect(sesion.sesionCaducada()).toBe(false);
  });

  it('sale igual aunque el backend no conteste: la sesión es del navegador', async () => {
    localStorage.setItem('catalina-support.token', 'un-token');

    const salida = sesion.salir();
    http
      .expectOne('/api/auth/logout')
      .flush(null, { status: 503, statusText: 'Service Unavailable' });
    await salida;

    expect(sesion.hayToken()).toBe(false);
  });

  it('comprobar() sin token no pregunta nada y da la sesión por comprobada', async () => {
    await expect(sesion.comprobar()).resolves.toBeNull();

    expect(sesion.comprobadaLaSesion()).toBe(true);
    expect(sesion.usuario()).toBeNull();
  });

  it('comprobar() con un token que ya no vale olvida la sesión', async () => {
    localStorage.setItem('catalina-support.token', 'caducado');

    const comprobacion = sesion.comprobar();
    http
      .expectOne('/api/auth/me')
      .flush({ error: 'auth.session.invalid' }, { status: 401, statusText: 'Unauthorized' });

    await expect(comprobacion).resolves.toBeNull();
    expect(sesion.hayToken()).toBe(false);
    expect(sesion.sesionCaducada()).toBe(true);
  });

  it('preguntar si hay sesión antes de entrar no deja la respuesta pegada', async () => {
    // Este fue un fallo de verdad: `hayToken` se calculaba leyendo el almacenamiento sin depender
    // de nada, así que se quedaba con el primer «no» para siempre y, después de entrar bien, la
    // guarda seguía devolviendo a la pantalla de entrada.
    expect(sesion.hayToken()).toBe(false);

    const entrada = sesion.entrar('ana@ejemplo.com', 'una-contraseña-larga');
    http
      .expectOne('/api/auth/login')
      .flush({ token: 'un-token', expiresAt: '2026-09-24T04:00:00Z', user: USUARIO });
    await entrada;

    expect(sesion.hayToken()).toBe(true);
  });

  it('el token se lee del almacenamiento cada vez, no se guarda en memoria', () => {
    expect(sesion.token()).toBeNull();

    // Otra pestaña (o cualquiera) escribe el token: esta pestaña lo ve sin avisarle.
    localStorage.setItem('catalina-support.token', 'puesto-desde-fuera');
    expect(sesion.token()).toBe('puesto-desde-fuera');
    expect(sesion.hayToken()).toBe(true);
  });

  it('el perfil adoptado conserva lo que el módulo de usuarios no sabe: el identificador y la fábrica', async () => {
    // El módulo `users` sólo devuelve cuentas de la tabla, y la de fábrica no está en ella: si la
    // sesión adoptara su respuesta entera, esa cuenta dejaría de reconocerse como de fábrica.
    localStorage.setItem('catalina-support.token', 'un-token');
    const comprobacion = sesion.comprobar();
    http.expectOne('/api/auth/me').flush({ user: { ...USUARIO, factory: true, id: 1 } });
    await comprobacion;

    sesion.actualizar({
      name: 'Administrador',
      lastName: 'de fábrica',
      email: '',
      role: 'administrador',
      origin: 'local',
    });

    expect(sesion.usuario()?.id).toBe(1);
    expect(sesion.usuario()?.factory).toBe(true);
    expect(sesion.usuario()?.name).toBe('Administrador');
    // Y el idioma nuevo manda dentro de la aplicación desde ese momento.
    expect(TestBed.inject(TranslationService).idioma()).toBe('en');
  });
  // Los caminos de entrada: la pantalla de entrada sólo ofrece el botón de Keycloak si esta
  // instalación lo tiene, y lo dice el backend (docs/modules/auth.md, decisión 27).
  describe('los caminos de entrada', () => {
    it('pregunta al backend qué caminos hay', async () => {
      expect(sesion.caminos().keycloak).toBe(false);

      const carga = sesion.cargarCaminos();
      http
        .expectOne('/api/auth/methods')
        .flush({ method: 'ad', local: false, ad: true, keycloak: false });
      await carga;

      // **Un método a la vez**: llega el nombre del que está puesto y cuál de los tres caminos está
      // abierto. La pantalla no tiene por qué conocer los nombres: lo que mira es qué camino ofrecer
      // (`docs/modules/settings.md`, sección 5.8).
      expect(sesion.caminos()).toEqual({ method: 'ad', local: false, ad: true, keycloak: false });
    });

    it('sin respuesta se queda con el camino local: no ofrece lo que no puede comprobar', async () => {
      const carga = sesion.cargarCaminos();
      http
        .expectOne('/api/auth/methods')
        .flush({ error: 'error interno' }, { status: 500, statusText: 'Server Error' });
      await carga;

      expect(sesion.caminos().keycloak).toBe(false);
      expect(sesion.caminos().local).toBe(true);
      expect(sesion.caminos().method).toBe('local');
    });

    // La vuelta de Keycloak trae el token en el fragmento: la pantalla lo lee y lo adopta aquí, sin
    // pasar por el formulario (docs/modules/auth.md, decisiones 13 y 30).
    it('adopta el token que trae la vuelta de Keycloak', () => {
      sesion.adoptarToken('el-token-de-la-vuelta');

      expect(sesion.hayToken()).toBe(true);
      expect(localStorage.getItem('catalina-support.token')).toBe('el-token-de-la-vuelta');
    });
  });
});
