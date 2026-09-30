import { Injectable, computed, signal } from '@angular/core';

import { EN } from './en';
import { ES, type Textos } from './es';

/** Los dos idiomas de la instalación. Son un valor cerrado, como los papeles. */
export type Idioma = 'es' | 'en';

const IDIOMAS: readonly Idioma[] = ['es', 'en'];

/** Dónde se recuerda lo que alguien eligió en el conmutador. */
const CLAVE_GUARDADA = 'catalina-support.idioma';

/**
 * Decide en qué idioma se entra (docs/modules/auth.md, decisión 25):
 *
 * 1. Si esa persona ya eligió uno en el conmutador, manda su elección.
 * 2. Si no, el primero de **los que pide su navegador** que sea de los nuestros: `es-419` es
 *    español y `en-GB` es inglés.
 * 3. **Si no pide ninguno de los dos, se entra en inglés**, que es el idioma que más gente puede
 *    leer de los dos que hay. Es una corrección del responsable: yo había propuesto español.
 *
 * Es una función suelta y no un método para poder probarla sin montar nada.
 */
export function idiomaElegido(preferidos: readonly string[], guardado: string | null): Idioma {
  const recordado = guardado?.toLowerCase().split('-')[0];
  if (recordado && esIdioma(recordado)) {
    return recordado;
  }

  for (const preferido of preferidos) {
    const base = preferido.toLowerCase().split('-')[0];
    if (base && esIdioma(base)) {
      return base;
    }
  }

  return 'en';
}

function esIdioma(valor: string): valor is Idioma {
  return (IDIOMAS as readonly string[]).includes(valor);
}

/**
 * Los textos de la interfaz, en el idioma que toque.
 *
 * Es armazón, no un módulo: lo usan todas las pantallas, y el idioma de la instalación y el de
 * cada cuenta no son de nadie en concreto (docs/arquitectura.md, sección 4, regla 5).
 */
@Injectable({ providedIn: 'root' })
export class TranslationService {
  private readonly idiomaActual = signal<Idioma>(leerIdiomaInicial());

  /** El idioma de ahora mismo. */
  readonly idioma = this.idiomaActual.asReadonly();

  /** Los textos de ahora mismo. En las plantillas se usa `t().pantalla.texto`. */
  readonly textos = computed<Textos>(() => (this.idiomaActual() === 'es' ? ES : EN));

  constructor() {
    // Que el navegador lo sepa: es lo que hace que los textos se partan bien y que los lectores
    // de pantalla lean en el idioma correcto.
    document.documentElement.lang = this.idiomaActual();
  }

  /** Cambia el idioma y lo recuerda en este navegador. */
  cambiar(idioma: Idioma): void {
    this.idiomaActual.set(idioma);
    document.documentElement.lang = idioma;

    try {
      localStorage.setItem(CLAVE_GUARDADA, idioma);
    } catch {
      // Sin almacenamiento (modo privado, por ejemplo) se cambia igual: sólo no se recuerda.
    }
  }

  /**
   * Adopta el idioma de la cuenta que acaba de entrar.
   *
   * Dentro de la aplicación manda el idioma de la cuenta, porque es el que decide también en qué
   * idioma se le escriben los correos (docs/modules/auth.md, sección 7).
   */
  usarIdiomaDeCuenta(idioma: string): void {
    const base = idioma?.toLowerCase().split('-')[0];
    if (base && esIdioma(base)) {
      this.cambiar(base);
    }
  }

  /**
   * El texto de una clave de error del backend.
   *
   * Si llega una clave que no conocemos —porque el backend es más nuevo que este frontend— se
   * responde el texto genérico: una pantalla nunca enseña una clave.
   */
  error(clave: string): string {
    const errores: Partial<Record<string, string>> = this.textos().errores;
    return errores[clave] ?? this.textos().errores['error interno'];
  }
}

function leerIdiomaInicial(): Idioma {
  let guardado: string | null = null;
  try {
    guardado = localStorage.getItem(CLAVE_GUARDADA);
  } catch {
    guardado = null;
  }

  const preferidos = navigator.languages?.length ? navigator.languages : [navigator.language];
  return idiomaElegido(preferidos, guardado);
}
