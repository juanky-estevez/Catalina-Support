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

  // **La zona horaria de la instalación viaja en la marca**: es la que decide cómo se leen todas las
  // fechas, y la primera lista de tickets ya la necesita. Sin marca se queda vacía, y quien formatea
  // cae a la del navegador.
  it('la zona horaria de la instalación se lee de la marca', async () => {
    expect(marca.zonaHoraria()).toBe('');

    const carga = marca.cargar();
    http.expectOne('/api/settings/brand').flush(marcaConNombre('Catalina Support', '1.0.0', 'Europe/Madrid'));
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
