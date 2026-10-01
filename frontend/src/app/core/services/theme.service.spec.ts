import { TestBed } from '@angular/core/testing';

import {
  aplicarTemaInicial,
  TEMAS,
  TEMAS_DE_FABRICA,
  TEMAS_FIJOS,
  ThemeService,
} from './theme.service';

/** Un sistema que dice claro u oscuro, para poder cambiar de opinión a mitad de la prueba. */
function sistema(empiezaOscuro: boolean): {
  cambiarA: (oscuro: boolean) => void;
  restaurar: () => void;
} {
  const oyentes = new Set<(evento: MediaQueryListEvent) => void>();
  const original = window.matchMedia;

  let oscuro = empiezaOscuro;
  window.matchMedia = ((consulta: string) =>
    ({
      matches: oscuro,
      media: consulta,
      addEventListener: (_tipo: string, oyente: (evento: MediaQueryListEvent) => void) =>
        oyentes.add(oyente),
      removeEventListener: (_tipo: string, oyente: (evento: MediaQueryListEvent) => void) =>
        oyentes.delete(oyente),
    }) as unknown as MediaQueryList) as typeof window.matchMedia;

  return {
    cambiarA: (nuevo: boolean) => {
      oscuro = nuevo;
      for (const oyente of oyentes) {
        oyente({ matches: nuevo } as MediaQueryListEvent);
      }
    },
    restaurar: () => {
      window.matchMedia = original;
    },
  };
}

describe('ThemeService', () => {
  let sistemaFalso: ReturnType<typeof sistema>;

  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
    TestBed.configureTestingModule({});
  });

  afterEach(() => {
    sistemaFalso?.restaurar();
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  it('de fábrica no hay nada elegido: el tema lo decide el sistema', () => {
    sistemaFalso = sistema(false);
    const tema = TestBed.inject(ThemeService);

    expect(tema.loDecideElSistema()).toBe(true);
    expect(tema.tema()).toBe('claro');
    // Sin atributo no se fuerza nada: manda el sistema operativo.
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  });

  it('con el sistema en oscuro, se ve oscuro sin que nadie haya elegido', () => {
    sistemaFalso = sistema(true);
    const tema = TestBed.inject(ThemeService);

    expect(tema.loDecideElSistema()).toBe(true);
    // Lo que enseña el conmutador es **lo que se ve**, no una tercera opción que no existe.
    expect(tema.tema()).toBe('oscuro');
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  });

  it('si el sistema cambia y nadie ha elegido, la aplicación cambia con él', () => {
    sistemaFalso = sistema(false);
    const tema = TestBed.inject(ThemeService);
    expect(tema.tema()).toBe('claro');

    sistemaFalso.cambiarA(true);
    expect(tema.tema()).toBe('oscuro');

    sistemaFalso.cambiarA(false);
    expect(tema.tema()).toBe('claro');
  });

  it('elegir un tema lo aplica en `html`, lo recuerda y deja de seguir al sistema', () => {
    sistemaFalso = sistema(false);
    const tema = TestBed.inject(ThemeService);

    tema.cambiar('oscuro');
    expect(document.documentElement.getAttribute('data-theme')).toBe('oscuro');
    expect(localStorage.getItem('catalina-support.tema')).toBe('oscuro');
    expect(tema.loDecideElSistema()).toBe(false);

    tema.cambiar('claro');
    expect(document.documentElement.getAttribute('data-theme')).toBe('claro');
    expect(localStorage.getItem('catalina-support.tema')).toBe('claro');
  });

  it('lo elegido manda sobre el sistema, y el sistema ya no cambia nada', () => {
    sistemaFalso = sistema(true);
    const tema = TestBed.inject(ThemeService);

    tema.cambiar('claro');
    expect(tema.tema()).toBe('claro');

    sistemaFalso.cambiarA(false);
    sistemaFalso.cambiarA(true);
    expect(tema.tema()).toBe('claro');
  });

  it('recuerda lo elegido al volver', () => {
    sistemaFalso = sistema(false);
    TestBed.inject(ThemeService).cambiar('oscuro');
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({});

    expect(TestBed.inject(ThemeService).tema()).toBe('oscuro');
    expect(document.documentElement.getAttribute('data-theme')).toBe('oscuro');
  });

  it('los ocho temas se aplican, y cada uno deja su atributo', () => {
    sistemaFalso = sistema(false);
    const tema = TestBed.inject(ThemeService);

    expect(TEMAS).toHaveLength(8);
    expect(TEMAS_DE_FABRICA).toEqual(['claro', 'oscuro']);
    expect(TEMAS_FIJOS).toHaveLength(6);
    // Ninguno se llama «automático»: seguir al sistema es no haber elegido.
    expect(TEMAS).not.toContain('sistema');

    for (const cualquiera of TEMAS) {
      tema.cambiar(cualquiera);
      expect(document.documentElement.getAttribute('data-theme')).toBe(cualquiera);
      expect(tema.tema()).toBe(cualquiera);
      expect(tema.loDecideElSistema()).toBe(false);
    }
  });

  it('no se cree un tema guardado que no existe', () => {
    sistemaFalso = sistema(false);
    localStorage.setItem('catalina-support.tema', 'sistema');

    const tema = TestBed.inject(ThemeService);
    expect(tema.loDecideElSistema()).toBe(true);
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  });

  it('el tema se aplica antes de arrancar, para no ver el parpadeo del tema equivocado', () => {
    localStorage.setItem('catalina-support.tema', 'oscuro');
    document.documentElement.removeAttribute('data-theme');

    aplicarTemaInicial();

    expect(document.documentElement.getAttribute('data-theme')).toBe('oscuro');
  });

  it('sin nada elegido, no fuerza ningún tema antes de arrancar', () => {
    aplicarTemaInicial();

    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  });

  it('un navegador sin `matchMedia` se queda en claro en vez de no arrancar', () => {
    const original = window.matchMedia;
    // @ts-expect-error: se simula un navegador que no la tiene.
    delete window.matchMedia;

    const tema = TestBed.inject(ThemeService);
    expect(tema.tema()).toBe('claro');
    expect(tema.loDecideElSistema()).toBe(true);

    window.matchMedia = original;
  });
});
