import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';

import { idiomaElegido, TranslationService } from './translation.service';

describe('idiomaElegido', () => {
  it('manda lo que esa persona eligió antes', () => {
    expect(idiomaElegido(['en-US'], 'es')).toBe('es');
    expect(idiomaElegido(['es-ES'], 'en')).toBe('en');
  });

  it('si no ha elegido nada, usa lo que pide el navegador', () => {
    expect(idiomaElegido(['es-ES'], null)).toBe('es');
    expect(idiomaElegido(['en-GB'], null)).toBe('en');
    // Las variantes son el mismo idioma: es-MX no es un idioma distinto de es.
    expect(idiomaElegido(['es-419'], null)).toBe('es');
    expect(idiomaElegido(['EN-us'], null)).toBe('en');
  });

  it('entre varios que pide el navegador, coge el primero de los nuestros', () => {
    expect(idiomaElegido(['fr-FR', 'es-ES'], null)).toBe('es');
    expect(idiomaElegido(['de', 'en-US', 'es'], null)).toBe('en');
  });

  it('si el navegador no pide ninguno de los dos, entra en inglés', () => {
    // Corrección del responsable el 2026-09-23: yo había propuesto español.
    expect(idiomaElegido(['fr-FR'], null)).toBe('en');
    expect(idiomaElegido(['de', 'it', 'pt-BR'], null)).toBe('en');
    expect(idiomaElegido([], null)).toBe('en');
  });

  it('no se cree un idioma guardado que no existe', () => {
    expect(idiomaElegido(['es-ES'], 'klingon')).toBe('es');
  });
});

describe('TranslationService', () => {
  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
  });

  afterEach(() => {
    TestBed.inject(HttpTestingController).verify();
    localStorage.clear();
  });

  it('traduce una clave de error del backend', () => {
    const textos = TestBed.inject(TranslationService);
    textos.cambiar('es');

    expect(textos.error('auth.invalidCredentials')).toContain('correo o la contraseña');
  });

  /**
   * El texto que rechaza el saneador del backend (`422 tickets.body.notAllowed`): dice **qué hacer**,
   * no sólo que no vale. Lo que la persona tiene delante es un comentario que no se ha guardado, y lo
   * que puede arreglar es el formato de lo que había pegado.
   */
  it('traduce el rechazo del texto, en los dos idiomas', () => {
    const textos = TestBed.inject(TranslationService);

    textos.cambiar('es');
    expect(textos.error('tickets.body.notAllowed')).toContain('Quita el formato');

    textos.cambiar('en');
    expect(textos.error('tickets.body.notAllowed')).toContain('Remove the formatting');
  });

  it('no enseña nunca una clave: si no la conoce, responde el texto genérico', () => {
    const textos = TestBed.inject(TranslationService);
    textos.cambiar('es');

    // Una clave que el backend mande y este frontend no conozca todavía.
    const respuesta = textos.error('algo.que.no.existe');

    expect(respuesta).not.toContain('algo.que.no.existe');
    expect(respuesta).toBe(textos.textos().errores['error interno']);
  });

  it('cambia de idioma y lo recuerda', () => {
    const textos = TestBed.inject(TranslationService);

    textos.cambiar('en');
    expect(textos.textos().entrada.entrar).toBe('Sign in');
    expect(localStorage.getItem('catalina-support.idioma')).toBe('en');

    textos.cambiar('es');
    expect(textos.textos().entrada.entrar).toBe('Entrar');
    expect(document.documentElement.lang).toBe('es');
  });

  it('adopta el idioma de la cuenta que entra, y no se deja engañar por uno que no existe', () => {
    const textos = TestBed.inject(TranslationService);

    textos.cambiar('es');
    textos.usarIdiomaDeCuenta('en-GB');
    expect(textos.idioma()).toBe('en');

    textos.usarIdiomaDeCuenta('fr');
    expect(textos.idioma()).toBe('en');
  });
});
