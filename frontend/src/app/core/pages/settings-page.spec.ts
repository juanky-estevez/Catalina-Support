import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { type ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { TranslationService } from '../i18n/translation.service';
import { SettingsPage } from './settings-page';

/**
 * **La región horaria y la dirección pública de la instalación.** Lo que se prueba aquí es lo que la
 * pantalla enseña y lo que manda al guardar: que el buscador filtre la lista de zonas, que la elegida
 * viaje en el `PUT`, que el aviso de que no hay https salga y se calle cuando toca, y que al cargar se
 * vea lo guardado. El aspecto y el teclado, en un navegador, los cubre la capa de Playwright.
 */
const CONFIGURACION = {
  name: 'Ayuntamiento de Ejemplo',
  entryMethod: 'local',
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
  language: 'es',
  primaryColor: '#1d4ed8',
  timeZone: 'UTC',
  publicAppUrl: 'https://soporte.demo.com',
  numberPrefix: 'CS',
  mainAssignment: 'por_turnos',
  mainNotification: 'al_asignado',
  internalAssignment: 'ninguna',
  internalNotification: 'a_nadie',
  updatedAt: '2026-09-29T00:00:00Z',
  brand: { light: { filled: false }, dark: { filled: false } },
};

/** La marca que se contesta cuando la pantalla la relee al guardar. */
const MARCA = {
  name: CONFIGURACION.name,
  version: '1.0.1',
  timeZone: 'America/Guayaquil',
  primaryColor: '#1d4ed8',
  colors: {
    light: '#1d4ed8',
    dark: '#1d4ed8',
    onLight: '#ffffff',
    onDark: '#0d0f12',
    isDefault: true,
  },
  logo: { light: { filled: false }, dark: { filled: false } },
  logoVersion: '',
};

describe('SettingsPage: la región horaria y la dirección pública', () => {
  let http: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();

    await TestBed.configureTestingModule({
      imports: [SettingsPage],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    }).compileComponents();

    // La pantalla se prueba en español, que es el idioma con el que se lee en el documento.
    TestBed.inject(TranslationService).cambiar('es');
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    localStorage.clear();
  });

  /** Deja correr las promesas: el backend de prueba contesta en el acto. */
  function esperar(): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve));
  }

  /** Monta la pantalla y responde lo primero que pide: la configuración. */
  async function montar(
    semilla: typeof CONFIGURACION = CONFIGURACION,
  ): Promise<ComponentFixture<SettingsPage>> {
    const fixture = TestBed.createComponent(SettingsPage);
    fixture.detectChanges();
    http.expectOne('/api/settings').flush(semilla);
    await esperar();
    fixture.detectChanges();
    return fixture;
  }

  function opciones(fixture: ComponentFixture<SettingsPage>): HTMLElement[] {
    return Array.from(fixture.nativeElement.querySelectorAll('[role="option"]')) as HTMLElement[];
  }

  function botonDeGuardar(fixture: ComponentFixture<SettingsPage>): HTMLButtonElement {
    const botones = Array.from(
      fixture.nativeElement.querySelectorAll('button'),
    ) as HTMLButtonElement[];
    // **La región y la dirección tienen su propio botón** (decisión del responsable, 2026-09-29): se
    // prefiere el suyo, que es el que está encendido cuando lo que ha cambiado es la zona o la
    // dirección. El de la numeración sigue valiendo para lo suyo.
    const boton =
      botones.find((candidato) =>
        candidato.textContent?.includes('Guardar la región y la dirección'),
      ) ?? botones.find((candidato) => candidato.textContent?.includes('Guardar la numeración'));

    if (!boton) {
      throw new Error('No se encontró ningún botón de guardar');
    }

    return boton;
  }

  /** El texto del aviso de que la dirección no va cifrada, o vacío si no está. */
  function avisoInseguro(fixture: ComponentFixture<SettingsPage>): string {
    const avisos = Array.from(
      fixture.nativeElement.querySelectorAll('app-aviso'),
    ) as HTMLElement[];
    return (
      avisos.find((aviso) => aviso.textContent?.includes('sin cifrar'))?.textContent?.trim() ?? ''
    );
  }

  /** Escribe en un campo como lo haría alguien: valor y evento `input`. */
  function escribir(fixture: ComponentFixture<SettingsPage>, selector: string, valor: string): void {
    const campo = fixture.nativeElement.querySelector(selector) as HTMLInputElement;
    campo.value = valor;
    campo.dispatchEvent(new Event('input'));
    fixture.detectChanges();
  }

  it('el buscador filtra la lista de zonas', async () => {
    const fixture = await montar();
    expect(opciones(fixture).length).toBeGreaterThan(1);

    escribir(fixture, '#zona-busqueda', 'guayaquil');

    const filtradas = opciones(fixture);
    expect(filtradas.length).toBe(1);
    expect(filtradas[0].textContent).toContain('America/Guayaquil');
  });

  it('elegir una zona la manda en el PUT', async () => {
    const fixture = await montar();

    escribir(fixture, '#zona-busqueda', 'guayaquil');
    opciones(fixture)[0].click();
    fixture.detectChanges();

    // Elegir una zona distinta de la guardada ya pide guardar.
    const guardar = botonDeGuardar(fixture);
    expect(guardar.disabled).toBe(false);

    guardar.click();
    await esperar();

    const peticion = http.expectOne(
      (candidata) => candidata.method === 'PUT' && candidata.url === '/api/settings',
    );
    const cuerpo = peticion.request.body as { timeZone: string; publicAppUrl: string };
    expect(cuerpo.timeZone).toBe('America/Guayaquil');
    expect(cuerpo.publicAppUrl).toBe(CONFIGURACION.publicAppUrl);
    peticion.flush({ ...CONFIGURACION, timeZone: 'America/Guayaquil' });
    await esperar();

    // Al guardar, la pantalla relee la marca: se contesta para no dejar la petición colgada.
    http.expectOne('/api/settings/brand').flush(MARCA);
    await esperar();
  });

  it('el aviso de seguridad sale con http y no sale con https ni con el campo vacío', async () => {
    const fixture = await montar();

    // La dirección guardada es https: no hay nada que avisar.
    expect(avisoInseguro(fixture)).toBe('');

    escribir(fixture, '#direccion-publica', '');
    expect(avisoInseguro(fixture)).toBe('');

    escribir(fixture, '#direccion-publica', 'http://soporte.local:8080');
    expect(avisoInseguro(fixture)).toContain('sin cifrar');
    // **No bloquea nada**: se puede guardar igual.
    expect(botonDeGuardar(fixture).disabled).toBe(false);

    escribir(fixture, '#direccion-publica', 'https://soporte.demo.com');
    expect(avisoInseguro(fixture)).toBe('');
  });

  it('al cargar se enseñan la zona y la dirección guardadas', async () => {
    const fixture = await montar({
      ...CONFIGURACION,
      timeZone: 'America/Guayaquil',
      publicAppUrl: 'http://soporte.local:8080',
    });

    const elegida = fixture.nativeElement.querySelector(
      '[role="option"][aria-selected="true"]',
    ) as HTMLElement | null;
    expect(elegida?.textContent).toContain('America/Guayaquil');

    const campo = fixture.nativeElement.querySelector('#direccion-publica') as HTMLInputElement;
    expect(campo.value).toBe('http://soporte.local:8080');

    // La zona elegida se explica sola: la hora que es en ella y su desfase.
    const texto = fixture.nativeElement.textContent as string;
    expect(texto).toContain('Ahora son las');
    expect(texto).toMatch(/GMT[+-]\d/);
  });
});
