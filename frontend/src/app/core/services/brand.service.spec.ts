import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { Title } from '@angular/platform-browser';

import {
  BrandService,
  LOGO_DE_FABRICA_CLARO,
  LOGO_DE_FABRICA_OSCURO,
  NOMBRE_DE_FABRICA,
} from './brand.service';
import { TituloDeLaInstalacion } from './title.strategy';

/** La marca que contesta el backend, con el nombre que se le ponga. */
function marcaConNombre(name: string, version = '1.0.0', timeZone = 'America/Guayaquil') {
  return {
    name,
    version,
    license: 'AGPL-3.0-only',
    sourceUrl: 'https://github.com/juanky-estevez/Catalina-Support',
    development: false,
    timeZone,
    primaryColor: '#1d4ed8',
    colors: {
      light: '#1d4ed8',
      dark: '#93b4f5',
      onLight: '#ffffff',
      onDark: '#101828',
      isDefault: false,
    },
    logo: { light: { filled: false }, dark: { filled: false } },
    logoVersion: '',
  };
}

/**
 * Lo que contesta el backend al tocar el logo: **el documento de configuración entero**, no sólo la
 * marca. El servicio lo devuelve tal cual, porque es lo que la pantalla necesita para refrescarse.
 */
const CONFIGURACION_CON_LOGO = {
  name: 'Ayuntamiento de Ejemplo',
  brand: { light: { filled: true }, dark: { filled: false } },
};

describe('BrandService', () => {
  let marca: BrandService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting(), TituloDeLaInstalacion],
    });

    marca = TestBed.inject(BrandService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  // **El nombre de la instalación sustituye a «Catalina Support»**: es lo que se lee en la pantalla
  // de entrada, en el menú lateral y en la pestaña (docs/modules/settings.md).
  it('el nombre es el que ha configurado la instalación', async () => {
    expect(marca.nombre()).toBe(NOMBRE_DE_FABRICA);

    const carga = marca.cargar();
    http.expectOne('/api/settings/brand').flush(marcaConNombre('Ayuntamiento de Ávila'));
    await carga;

    expect(marca.nombre()).toBe('Ayuntamiento de Ávila');
  });

  // Sin backend —o con una instalación que no contesta— la aplicación sigue funcionando y **la
  // pestaña no se queda sin nombre**: se enseña el de fábrica.
  it('sin marca se queda el nombre de fábrica', async () => {
    const carga = marca.cargar();
    http
      .expectOne('/api/settings/brand')
      .flush({ error: 'error interno' }, { status: 500, statusText: 'Server Error' });
    await carga;

    expect(marca.nombre()).toBe(NOMBRE_DE_FABRICA);
    expect(marca.hayLogoPropio()).toBe(false);
    expect(marca.urlDelLogo(false)).toBe(LOGO_DE_FABRICA_CLARO);
    expect(marca.urlDelLogo(true)).toBe(LOGO_DE_FABRICA_OSCURO);
  });

  // Un nombre en blanco —que el backend no admite, pero que puede llegar de una base tocada a mano—
  // no deja la pestaña en blanco: vuelve el de fábrica.
  it('un nombre en blanco no deja la pestaña vacía', async () => {
    const carga = marca.cargar();
    http.expectOne('/api/settings/brand').flush(marcaConNombre('   '));
    await carga;

    expect(marca.nombre()).toBe(NOMBRE_DE_FABRICA);
  });

  // **La versión del sistema**, que es lo que se lee en la fila de salir del menú y en el pie de la
  // pantalla de entrada (docs/modules/settings.md, sección 5.9). La `v` la pone la interfaz.
  it('la versión del sistema se enseña con su «v»', async () => {
    expect(marca.versionDelSistema()).toBe('');

    const carga = marca.cargar();
    http.expectOne('/api/settings/brand').flush(marcaConNombre('Catalina Support', '2.3.1'));
    await carga;

    expect(marca.versionDelSistema()).toBe('v2.3.1');
  });

  // Y sin marca —o si viene vacía— **no se enseña ningún número**: inventarse uno sería peor que no
  // decir ninguno, y esto es justo lo que se mira cuando algo va mal.
  it('sin versión no se enseña ningún número', async () => {
    const carga = marca.cargar();
    http.expectOne('/api/settings/brand').flush(marcaConNombre('Catalina Support', '  '));
    await carga;

    expect(marca.versionDelSistema()).toBe('');

    const otra = marca.recargar();
    http
      .expectOne('/api/settings/brand')
      .flush({ error: 'error interno' }, { status: 500, statusText: 'Server Error' });
    await otra;

    expect(marca.versionDelSistema()).toBe('');
  });

  it('identifica una compilación de desarrollo y expone su fuente HTTPS', async () => {
    const carga = marca.cargar();
    http.expectOne('/api/settings/brand').flush({
      ...marcaConNombre('Catalina Support'),
      development: true,
    });
    await carga;

    expect(marca.versionDelSistema()).toBe('v1.0.0 · dev');
    expect(marca.fuente()).toEqual({
      url: 'https://github.com/juanky-estevez/Catalina-Support',
      licencia: 'AGPL-3.0-only',
    });
  });

  it('no inventa un enlace si la fuente falta o no usa HTTPS', async () => {
    const carga = marca.cargar();
    http.expectOne('/api/settings/brand').flush({
      ...marcaConNombre('Catalina Support'),
      sourceUrl: 'javascript:alert(1)',
    });
    await carga;

    expect(marca.fuente()).toBeNull();
  });

  // **La zona horaria de la instalación viaja en la marca**: es la que decide cómo se leen todas las
  // fechas, y la primera lista de tickets ya la necesita. Sin marca se queda vacía, y quien formatea
  // cae a la del navegador.
  it('la zona horaria de la instalación se lee de la marca', async () => {
    expect(marca.zonaHoraria()).toBe('');

    const carga = marca.cargar();
    http
      .expectOne('/api/settings/brand')
      .flush(marcaConNombre('Catalina Support', '1.0.0', 'Europe/Madrid'));
    await carga;

    expect(marca.zonaHoraria()).toBe('Europe/Madrid');
  });

  it('una zona en blanco no deja la marca sin zona utilizable', async () => {
    const carga = marca.cargar();
    http.expectOne('/api/settings/brand').flush(marcaConNombre('Catalina Support', '1.0.0', '  '));
    await carga;

    // Vacía: quien formatea decide, y lee con la del navegador (shared/fechas.ts).
    expect(marca.zonaHoraria()).toBe('');
  });

  // **El logo es de la marca**: se sube a su hueco (`variant`) en un `FormData`, y lo que contesta el
  // backend es el documento de configuración entero (docs/modules/settings.md).
  it('subirLogo sube el archivo a su hueco por POST', async () => {
    const archivo = new File(['logo'], 'marca.png', { type: 'image/png' });
    const subida = marca.subirLogo(true, archivo);

    const peticion = http.expectOne('/api/settings/brand/logo?variant=oscuro');
    expect(peticion.request.method).toBe('POST');
    const datos = peticion.request.body as FormData;
    expect(datos).toBeInstanceOf(FormData);
    expect((datos.get('logo') as File).name).toBe('marca.png');
    peticion.flush(CONFIGURACION_CON_LOGO);

    await expect(subida).resolves.toEqual(CONFIGURACION_CON_LOGO);
  });

  it('subirLogo distingue el hueco claro del oscuro', async () => {
    const archivo = new File(['logo'], 'marca.png', { type: 'image/png' });
    const subida = marca.subirLogo(false, archivo);

    const peticion = http.expectOne('/api/settings/brand/logo?variant=claro');
    peticion.flush(CONFIGURACION_CON_LOGO);

    await subida;
  });

  // **Restablecer el logo**: el mismo hueco, pero por `DELETE`.
  it('restablecerLogo quita el logo propio de su hueco por DELETE', async () => {
    const restablecido = marca.restablecerLogo(true);

    const peticion = http.expectOne('/api/settings/brand/logo?variant=oscuro');
    expect(peticion.request.method).toBe('DELETE');
    peticion.flush(CONFIGURACION_CON_LOGO);

    await expect(restablecido).resolves.toEqual(CONFIGURACION_CON_LOGO);
  });

  // **La vista previa del color**: se pregunta al backend sin guardarlo, el color viaja escapado y
  // vuelve **la marca entera**, que es lo que la pantalla lee.
  it('previsualizarColor pide la marca con el color y devuelve lo que recibe', async () => {
    const vista = marca.previsualizarColor('#1d4ed8');

    const peticion = http.expectOne('/api/settings/brand?color=%231d4ed8');
    expect(peticion.request.method).toBe('GET');
    const esperada = marcaConNombre('Catalina Support');
    peticion.flush(esperada);

    await expect(vista).resolves.toEqual(esperada);
  });
});

describe('TituloDeLaInstalacion', () => {
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting(), TituloDeLaInstalacion],
    });

    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  // La pestaña se llama como la instalación, y **cambia al guardar el nombre nuevo** sin recargar.
  it('la pestaña lleva el nombre de la instalación, y cambia al guardarlo', async () => {
    const titulo = TestBed.inject(Title);
    const marca = TestBed.inject(BrandService);

    TestBed.inject(TituloDeLaInstalacion).updateTitle();
    expect(titulo.getTitle()).toBe(NOMBRE_DE_FABRICA);

    const carga = marca.cargar();
    http.expectOne('/api/settings/brand').flush(marcaConNombre('Mesa de ayuda de Acme'));
    await carga;
    TestBed.flushEffects();

    expect(titulo.getTitle()).toBe('Mesa de ayuda de Acme');
  });
});
