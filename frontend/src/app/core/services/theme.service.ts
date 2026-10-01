import { Injectable, computed, signal } from '@angular/core';

/**
 * Los temas que se pueden elegir.
 *
 * Los dos primeros son **los de fábrica**, los personalizables: llevan el color institucional que
 * fija el administrador, y son los que se usan mientras nadie elige (siguiendo al sistema). Los seis
 * siguientes son **fijos**, cada uno con su propia paleta y su propio acento
 * (`docs/interfaz-y-experiencia.md`, sección 6.2).
 */
export type Tema =
  'claro' | 'oscuro' | 'papel' | 'niebla' | 'contraste' | 'grafito' | 'noche' | 'sepia';

/** Los de fábrica, que son los que se usan cuando nadie ha elegido. */
export const TEMAS_DE_FABRICA: readonly Tema[] = ['claro', 'oscuro'];

/** Los seis fijos, cada uno con su paleta. */
export const TEMAS_FIJOS: readonly Tema[] = [
  'papel',
  'niebla',
  'contraste',
  'grafito',
  'noche',
  'sepia',
];

/** Todos, en el orden en que se ofrecen: primero los de fábrica. */
export const TEMAS: readonly Tema[] = [...TEMAS_DE_FABRICA, ...TEMAS_FIJOS];

/** Los temas oscuros. El logo se elige por la claridad del tema, no por su nombre. */
export const TEMAS_OSCUROS: readonly Tema[] = ['oscuro', 'grafito', 'noche', 'sepia'];

/** Dónde se recuerda lo que alguien eligió. Si no hay nada, se sigue al sistema. */
const CLAVE_GUARDADA = 'catalina-support.tema';

/** El atributo de `<html>` que decide el tema; sin él, mandan los de fábrica según el sistema. */
const ATRIBUTO = 'data-theme';

/**
 * El tema: se sigue al sistema hasta que alguien elige.
 *
 * Los colores no viven aquí: viven en `styles.css`, un bloque de variables por tema. Este servicio
 * decide **cuál** se usa —con un atributo en `<html>`— y lo recuerda en el navegador.
 *
 * Vive en `core` porque el tema es del armazón, no de ningún módulo.
 */
@Injectable({ providedIn: 'root' })
export class ThemeService {
  /** Lo que esa persona eligió, o nada si todavía no ha elegido. */
  private readonly eleccion = signal<Tema | null>(leerTemaGuardado());

  /** Si el sistema pide oscuro. Se sigue en vivo: el sistema puede cambiar con la página abierta. */
  private readonly sistemaOscuro = signal(consultaOscuro()?.matches ?? false);

  /**
   * El tema que se está viendo.
   *
   * Es lo que enseña el conmutador: si nadie ha elegido nada, enseña **lo que se ve** —claro u
   * oscuro según el sistema—, no una tercera opción que no existe.
   */
  readonly tema = computed<Tema>(
    () => this.eleccion() ?? (this.sistemaOscuro() ? 'oscuro' : 'claro'),
  );

  /** Si todavía no ha elegido nadie: el tema lo están decidiendo los dos de fábrica y el sistema. */
  readonly loDecideElSistema = computed(() => this.eleccion() === null);

  /** Si el tema que se está viendo es oscuro. Es lo que decide qué logo se enseña. */
  readonly esOscuro = computed(() => TEMAS_OSCUROS.includes(this.tema()));

  constructor() {
    aplicarAtributo(this.eleccion());

    // Si el sistema cambia de claro a oscuro —o al revés— y nadie ha elegido, la página cambia con él.
    consultaOscuro()?.addEventListener('change', (evento) =>
      this.sistemaOscuro.set(evento.matches),
    );
  }

  /** Elige un tema, y lo recuerda en este navegador. */
  cambiar(tema: Tema): void {
    this.eleccion.set(tema);
    aplicarAtributo(tema);

    try {
      localStorage.setItem(CLAVE_GUARDADA, tema);
    } catch {
      // Sin almacenamiento (modo privado, por ejemplo) el tema se cambia igual: sólo no se recuerda.
    }
  }
}

/**
 * Pone el tema guardado **antes de arrancar la aplicación**, para que nadie vea un parpadeo del tema
 * equivocado al cargar.
 *
 * Se llama desde `main.ts` y no desde un `script` incrustado en el `index.html` **a propósito**: la
 * CSP de producción es `script-src 'self'`, sin `'unsafe-inline'`, así que un script incrustado no
 * funcionaría allí (`docs/interfaz-y-experiencia.md`, sección 6.2).
 */
export function aplicarTemaInicial(): void {
  aplicarAtributo(leerTemaGuardado());
}

/** Lee lo que esa persona eligió, y no se cree un valor que no conocemos. */
function leerTemaGuardado(): Tema | null {
  try {
    const guardado = localStorage.getItem(CLAVE_GUARDADA);
    if (guardado && (TEMAS as readonly string[]).includes(guardado)) {
      return guardado as Tema;
    }
  } catch {
    // Sin almacenamiento se sigue al sistema, que es lo de fábrica.
  }

  return null;
}

/** Sin elección no se fuerza nada: el atributo fuera y deciden los de fábrica según el sistema. */
function aplicarAtributo(tema: Tema | null): void {
  if (tema === null) {
    document.documentElement.removeAttribute(ATRIBUTO);
    return;
  }

  document.documentElement.setAttribute(ATRIBUTO, tema);
}

/**
 * La consulta de «el sistema pide oscuro», o nada si el navegador no la tiene.
 *
 * Devuelve nada en lugar de fallar porque esto se ejecuta al construir el servicio: un navegador sin
 * `matchMedia` (o un entorno de pruebas sin ella) se queda en claro, que es lo de fábrica, en vez de
 * dejar la aplicación sin arrancar.
 */
function consultaOscuro(): MediaQueryList | null {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return null;
  }

  return window.matchMedia('(prefers-color-scheme: dark)');
}
