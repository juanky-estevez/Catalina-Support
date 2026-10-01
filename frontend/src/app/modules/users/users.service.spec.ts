import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { UsersService } from './users.service';

const CUENTA = {
  id: 7,
  name: 'Ana',
  lastName: 'Pérez',
  email: 'ana@ejemplo.com',
  role: 'usuario',
  origin: 'local',
  language: 'es',
  isActive: true,
  hasPassword: true,
};

/**
 * El servicio del módulo, que es **lo único que llama a `/api/users/**`**
 * (`docs/arquitectura.md`, sección 4): aquí se comprueba que cada cosa va a su ruta y que los filtros
 * viajan como el backend los espera.
 */
describe('UsersService', () => {
  let users: UsersService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    users = TestBed.inject(UsersService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
  });

  it('la lista lleva sus filtros y su página, y no manda los que están vacíos', async () => {
    const peticion = users.listar({
      role: '',
      origin: 'local',
      active: 'false',
      q: '  ana ',
      page: 2,
    });

    const pidiendo = http.expectOne((peticion) => peticion.url === '/api/users');
    expect(pidiendo.request.method).toBe('GET');
    expect(pidiendo.request.params.get('page')).toBe('2');
    expect(pidiendo.request.params.get('origin')).toBe('local');
    expect(pidiendo.request.params.get('active')).toBe('false');
    // La búsqueda va sin los espacios de fuera: es lo que se escribe al buscar.
    expect(pidiendo.request.params.get('q')).toBe('ana');
    // Un filtro vacío no viaja: el backend entiende «sin filtrar» por su ausencia.
    expect(pidiendo.request.params.has('role')).toBe(false);

    pidiendo.flush({ users: [CUENTA], total: 1, page: 2, perPage: 25 });
    await expect(peticion).resolves.toEqual({ users: [CUENTA], total: 1, page: 2, perPage: 25 });
  });

  it('cada acción va a su ruta: las de estado tienen la suya, no un `PATCH`', async () => {
    // Es la decisión 14 de `users.md`: desactivar no puede pasar por el `PATCH` general.
    void users.desactivar(7);
    http.expectOne({ url: '/api/users/7/deactivate', method: 'POST' }).flush({ user: CUENTA });

    void users.activar(7);
    http.expectOne({ url: '/api/users/7/activate', method: 'POST' }).flush({ user: CUENTA });

    void users.mandarEnlace(7);
    http.expectOne({ url: '/api/users/7/reset-password', method: 'POST' }).flush({ user: CUENTA });

    void users.cambiarOrigen(7, 'local');
    const origen = http.expectOne({ url: '/api/users/7/origin', method: 'POST' });
    expect(origen.request.body).toEqual({ origin: 'local', externalId: '' });
    origen.flush({ user: CUENTA });
  });

  it('la ficha y los cambios van a la cuenta que toca, y el perfil propio a `/me`', async () => {
    void users.ficha(7);
    http.expectOne({ url: '/api/users/7', method: 'GET' }).flush({ user: CUENTA });

    void users.cambiar(7, { name: 'Ana', lastName: 'Pérez Gómez' });
    const cambios = http.expectOne({ url: '/api/users/7', method: 'PATCH' });
    expect(cambios.request.body).toEqual({ name: 'Ana', lastName: 'Pérez Gómez' });
    cambios.flush({ user: CUENTA });

    // **El perfil propio se cambia en `/me`**, y sólo con lo que puede cambiar cualquiera: ni el
    // correo ni el papel viajan de aquí (docs/modules/users.md, sección 8).
    void users.cambiarPerfil({ language: 'en' });
    const perfil = http.expectOne({ url: '/api/users/me', method: 'PATCH' });
    expect(perfil.request.body).toEqual({ language: 'en' });
    perfil.flush({ user: CUENTA });
  });

  it('el alta manda lo que el backend espera, y el aviso del enlace vuelve con ella', async () => {
    const alta = users.crear({
      name: 'Ana',
      lastName: 'Pérez',
      email: 'ana@ejemplo.com',
      role: 'usuario',
      origin: 'local',
      language: 'es',
    });

    const peticion = http.expectOne({ url: '/api/users', method: 'POST' });
    expect(peticion.request.body).toEqual({
      name: 'Ana',
      lastName: 'Pérez',
      email: 'ana@ejemplo.com',
      role: 'usuario',
      origin: 'local',
      language: 'es',
    });

    // `invited` es lo que distingue «cuenta creada y enlace mandado» de «cuenta creada sin enlace».
    peticion.flush({ user: CUENTA, invited: true });
    await expect(alta).resolves.toEqual({ user: CUENTA, invited: true });
  });
});
