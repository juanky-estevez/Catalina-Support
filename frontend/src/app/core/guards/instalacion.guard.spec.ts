import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { UrlTree, provideRouter, type CanActivateFn } from '@angular/router';

import { SetupService, type EstadoDeInstalacion } from '../services/setup.service';
import { instalacionComprobada, instalacionSinSellar } from './instalacion.guard';

/**
 * **Las dos guardas del primer arranque**, que son lo que decide si se ve el asistente o la entrada.
 *
 * Lo que se prueba aquí es la regla entera, en las dos direcciones y con el caso que más importa: un
 * fallo de red **no puede dejar la aplicación sin puerta**. Esa es la prueba que evita que un backend
 * caído convierta la aplicación en una pantalla de instalación que no lleva a ninguna parte
 * (`docs/primer-arranque.md`, secciones 2 y 6).
 */
describe('Las guardas de la instalación', () => {
  let http: HttpTestingController;

  /** Un estado de instalación con lo mínimo: lo que decide la guarda es `installed`. */
  function estado(installed: boolean): EstadoDeInstalacion {
    return {
      installed,
      aiRequired: false,
      name: 'Ayuntamiento de Ejemplo',
      language: 'es',
      entryMethod: 'local',
      timeZone: 'UTC',
      publicAppUrl: '',
      directory: {
        host: '',
        port: '',
        useTls: false,
        bindDn: '',
        searchBase: '',
        userFilter: '',
        attrEmail: '',
        attrName: '',
        attrLastName: '',
        attrId: '',
        passwordSet: false,
      },
      keycloak: { issuer: '', clientId: '', redirectUri: '', secretSet: false },
      mail: {
        host: '',
        port: '',
        secure: false,
        user: '',
        fromName: '',
        fromEmail: '',
        passwordSet: false,
      },
      aiAvailable: false,
    };
  }

  beforeEach(async () => {
    localStorage.clear();

    await TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    }).compileComponents();

    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    localStorage.clear();
  });

  /** Ejecuta una guarda como la ejecutaría el router: dentro del contexto de inyección. */
  function ejecutar(guarda: CanActivateFn): Promise<boolean | UrlTree> {
    return TestBed.runInInjectionContext(() =>
      (guarda as unknown as () => Promise<boolean | UrlTree>)(),
    );
  }

  it('sin la instalación sellada, cualquier pantalla lleva al asistente', async () => {
    const resultado = ejecutar(instalacionComprobada);

    http.expectOne('/api/setup').flush(estado(false));

    const destino = await resultado;
    expect(destino).toBeInstanceOf(UrlTree);
    expect(String(destino)).toBe('/setup');
  });

  it('con la instalación sellada, la aplicación sigue por donde iba', async () => {
    const resultado = ejecutar(instalacionComprobada);

    http.expectOne('/api/setup').flush(estado(true));

    expect(await resultado).toBe(true);
  });

  it('con la instalación sellada, /setup lleva a la entrada', async () => {
    const resultado = ejecutar(instalacionSinSellar);

    http.expectOne('/api/setup').flush(estado(true));

    const destino = await resultado;
    expect(destino).toBeInstanceOf(UrlTree);
    // Y lleva el aviso de que ya está terminada, para que se sepa por qué no se ve el asistente.
    expect(String(destino)).toContain('/login');
    expect((destino as UrlTree).queryParams['installed']).toBe('1');
  });

  it('sin la instalación sellada, /setup se puede ver', async () => {
    const resultado = ejecutar(instalacionSinSellar);

    http.expectOne('/api/setup').flush(estado(false));

    expect(await resultado).toBe(true);
  });

  it('si el estado no contesta, no se bloquea la aplicación: se sigue por la entrada', async () => {
    const resultado = ejecutar(instalacionComprobada);

    http.expectOne('/api/setup').error(new ProgressEvent('error'));

    // **Un fallo de red no puede dejar la aplicación sin puerta**: se pasa, y la entrada es el
    // camino de siempre. Lo contrario sería convertir un corte en una pantalla de instalación.
    expect(await resultado).toBe(true);
  });

  it('si el estado no contesta, /setup también lleva a la entrada', async () => {
    const resultado = ejecutar(instalacionSinSellar);

    http.expectOne('/api/setup').error(new ProgressEvent('error'));

    expect(String(await resultado)).toContain('/login');
  });

  it('el estado se pregunta una sola vez, aunque la guarda pase varias veces', async () => {
    const setup = TestBed.inject(SetupService);

    const primera = ejecutar(instalacionComprobada);
    const segunda = ejecutar(instalacionSinSellar);

    // Las dos llamadas comparten la misma petición: una sola vez por arranque.
    http.expectOne('/api/setup').flush(estado(true));

    expect(await primera).toBe(true);
    expect(String(await segunda)).toContain('/login');
    http.expectNone('/api/setup');
    expect(await setup.estaInstalada()).toBe(true);
    http.expectNone('/api/setup');
  });
});
