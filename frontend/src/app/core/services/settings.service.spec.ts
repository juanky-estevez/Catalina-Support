import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { SettingsService, type Configuracion } from './settings.service';

/**
 * La configuración que contesta el backend. Es **un documento entero**: lo que se prueba aquí es que
 * el servicio llame a la ruta que toca con el cuerpo que toca y devuelva lo que recibe, no lo que la
 * pantalla hace con ello (docs/modules/settings.md).
 */
const CONFIGURACION: Configuracion = {
  name: 'Ayuntamiento de Ejemplo',
  entryMethod: 'local',
  directory: {
    host: 'ldap.ejemplo.com',
    port: '389',
    useTls: false,
    bindDn: 'cn=servicio,dc=ejemplo,dc=com',
    searchBase: 'dc=ejemplo,dc=com',
    userFilter: '(uid={0})',
    attrEmail: 'mail',
    attrName: 'givenName',
    attrLastName: 'sn',
    attrId: 'uid',
    passwordSet: true,
  },
  keycloak: {
    issuer: 'https://keycloak.ejemplo.com/realms/catalina',
    clientId: 'catalina-support',
    redirectUri: 'https://soporte.demo.com/login',
    secretSet: true,
  },
  language: 'es',
  primaryColor: '#1d4ed8',
  timeZone: 'America/Guayaquil',
  publicAppUrl: 'https://soporte.demo.com',
  numberPrefix: 'CS',
  mainAssignment: 'por_turnos',
  mainNotification: 'al_asignado',
  internalAssignment: 'ninguna',
  internalNotification: 'a_nadie',
  updatedAt: '2026-09-29T00:00:00Z',
  aiUrl: 'http://catalina_support_ai:8080',
  aiModel: 'qwen2.5-1.5b-instruct',
  brand: { light: { filled: false }, dark: { filled: false } },
};

describe('SettingsService', () => {
  let ajustes: SettingsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    ajustes = TestBed.inject(SettingsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  // **Listar**: la pantalla lee el documento entero de una vez, y el servicio devuelve lo que llega.
  it('cargar pide la configuración y devuelve lo que recibe', async () => {
    const carga = ajustes.cargar();

    const peticion = http.expectOne('/api/settings');
    expect(peticion.request.method).toBe('GET');
    peticion.flush(CONFIGURACION);

    await expect(carga).resolves.toEqual(CONFIGURACION);
  });

  // **Guardar cada bloque es el mismo `PUT`**: el documento viaja entero y el servicio lo manda tal
  // cual lo compone la pantalla.
  it('guardar manda el documento entero por PUT y devuelve lo que recibe', async () => {
    const nueva = { ...CONFIGURACION, name: 'Otro nombre' };
    const guardado = ajustes.guardar(nueva);

    const peticion = http.expectOne('/api/settings');
    expect(peticion.request.method).toBe('PUT');
    expect(peticion.request.body).toEqual(nueva);
    peticion.flush(nueva);

    await expect(guardado).resolves.toEqual(nueva);
  });

  // **Las tres pruebas de conexión**: no guardan nada y contestan sin cuerpo; lo que importa es que
  // cada una vaya a su ruta con lo que hay en pantalla.
  it('probarDirectorio manda lo escrito a su ruta', async () => {
    const escrito = {
      host: 'ldap.ejemplo.com',
      searchBase: 'dc=ejemplo,dc=com',
      bindPassword: 'x',
    };
    const prueba = ajustes.probarDirectorio(escrito);

    const peticion = http.expectOne('/api/settings/directory/test');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual(escrito);
    peticion.flush(null);

    await prueba;
  });

  it('probarKeycloak manda lo escrito a su ruta', async () => {
    const escrito = { issuer: 'https://keycloak.ejemplo.com/realms/catalina', clientSecret: 'x' };
    const prueba = ajustes.probarKeycloak(escrito);

    const peticion = http.expectOne('/api/settings/keycloak/test');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual(escrito);
    peticion.flush(null);

    await prueba;
  });

  it('probarMotor manda la dirección del motor a su ruta', async () => {
    const prueba = ajustes.probarMotor('http://catalina_support_ai:8080');

    const peticion = http.expectOne('/api/settings/ai/test');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual({ url: 'http://catalina_support_ai:8080' });
    peticion.flush(null);

    await prueba;
  });
});
