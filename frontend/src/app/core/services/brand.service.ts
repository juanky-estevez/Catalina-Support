import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

// **Sólo el tipo**: lo que devuelven subir y restablecer el logo es el documento de configuración
// entero, porque es lo que contesta el backend. No hay dependencia de ejecución con el otro servicio.
import { type Configuracion } from './settings.service';
import { TranslationService } from '../i18n/translation.service';

/** La marca de la instalación, tal y como la cuenta el backend (docs/modules/settings.md). */
export interface Marca {
  /**
   * El nombre de la instalación: la institución, la empresa o el equipo.
   *
   * **Es el texto que sustituye a «Catalina Support»** en la pantalla de entrada, en el menú lateral y
   * en la pestaña del navegador. Lo configura un Administrador desde Configuración, y viaja con la
   * marca porque hace falta **antes de que nadie entre**.
   */
  readonly name: string;
  /**
   * La versión del sistema, **sin la `v`**: la `v` la pone la interfaz al enseñarla.
   *
   * Es un dato del software, no de la institución, así que no se configura: la escribe el backend en
   * su código y viaja aquí porque **la marca es lo que la aplicación pide al arrancar**, y así el menú
   * y la pantalla de entrada no necesitan una segunda llamada
   * (`docs/modules/settings.md`, sección 5.9).
   */
  readonly version: string;
  readonly license?: string;
  readonly sourceUrl?: string;
  readonly development?: boolean;
  readonly language?: string;
  readonly settingsVersion?: number;
  /**
   * La zona horaria de la instalación, **en nombre IANA** (`America/Guayaquil`).
   *
   * **Es la que decide cómo se leen todas las fechas**, en la interfaz y en los correos: las fechas
   * guardadas siguen en UTC, y cambiar la zona no mueve ningún ticket, sólo cambia la hora a la que
   * se leen. Viaja en la marca porque **la marca es lo que la aplicación pide al arrancar**, y la
   * primera lista de tickets ya necesita la zona.
   */
  readonly timeZone: string;
  readonly primaryColor: string;
  readonly colors: {
    readonly light: string;
    readonly dark: string;
    readonly onLight: string;
    readonly onDark: string;
    readonly isDefault: boolean;
  };
  readonly logo: {
    readonly light: { readonly filled: boolean };
    readonly dark: { readonly filled: boolean };
  };
  readonly logoVersion: string;
}

/**
 * Los dos logos de fábrica: los archivos que trae la aplicación, y a los que se vuelve siempre que se
 * pueda (`docs/modules/settings.md`, secciones 5.1 y 5.6).
 *
 * **Son dos, uno por tema**, igual que los huecos del logo propio: un logo hecho para un fondo claro
 * no se lee sobre uno oscuro, y al revés. Los dos viven en `public/`, así que se cambian reemplazando
 * el archivo —son de la aplicación, no de la instalación, y no se configuran desde la pantalla—.
 */
export const LOGO_DE_FABRICA_CLARO = 'logo-catalina-support-light.png';
export const LOGO_DE_FABRICA_OSCURO = 'logo-catalina-support-dark.png';

/** El logo de fábrica que le toca a ese tema. */
export function logoDeFabrica(oscuro: boolean): string {
  return oscuro ? LOGO_DE_FABRICA_OSCURO : LOGO_DE_FABRICA_CLARO;
}

/**
 * El nombre de fábrica: el que trae la aplicación y al que se vuelve si se deja el campo vacío.
 *
 * Está aquí, y no sólo en el diccionario de textos, porque **la pestaña del navegador tiene que
 * llamarse algo** aunque el backend todavía no haya contestado o no conteste nunca: sin esto, una
 * instalación que no responde dejaría la pestaña en blanco.
 */
export const NOMBRE_DE_FABRICA = 'Catalina Support';

/**
 * La marca: el nombre, el logo de la institución, su color y la vista previa de ese color.
 *
 * **Es de la instalación, no de ninguna persona**, y hace falta **antes de que nadie entre** —la
 * pantalla de entrada lo enseña—, así que su endpoint es público
 * (`docs/modules/settings.md`, sección 5.4).
 *
 * **Aquí vive la marca entera**: leerla, subir y restablecer sus dos logos y previsualizar el color.
 * Es de quien es el recurso (`/api/settings/brand`), y Configuración lo usa para eso; el resto de la
 * configuración lo sirve `SettingsService`.
 *
 * Vive en `core` porque es armazón: lo usan la entrada, el menú lateral y Configuración.
 */
@Injectable({ providedIn: 'root' })
export class BrandService {
  private readonly http = inject(HttpClient);
  private readonly textos = inject(TranslationService);

  private readonly marcaActual = signal<Marca | null>(null);

  /** La marca leída, o nada si todavía no ha llegado (o si la instalación no responde). */
  readonly marca = this.marcaActual.asReadonly();

  /**
   * El nombre de la instalación, listo para enseñar.
   *
   * Si todavía no ha llegado la marca —o la instalación no responde— devuelve **el de fábrica**, que
   * es lo mismo que hace el backend cuando el nombre está vacío.
   */
  readonly nombre = computed(() => this.marcaActual()?.name?.trim() || NOMBRE_DE_FABRICA);

  /** Si la instalación tiene un logo propio, en cualquiera de los dos huecos. */
  readonly hayLogoPropio = computed(() => {
    const marca = this.marcaActual();
    return marca !== null && (marca.logo.light.filled || marca.logo.dark.filled);
  });

  /** El sello de la marca: cambia al reemplazar un logo, y es lo que evita enseñar el viejo. */
  readonly version = computed(() => this.marcaActual()?.logoVersion ?? '');

  /**
   * La versión del sistema, lista para enseñar, **con su `v`**.
   *
   * Vacía mientras no haya llegado la marca —o si la instalación no responde—, y entonces no se
   * enseña nada: **inventarse un número sería peor que no decir ninguno**, y esto es justo lo que se
   * mira cuando algo va mal.
   */
  readonly versionDelSistema = computed(() => {
    const version = this.marcaActual()?.version?.trim();
    if (!version) {
      return '';
    }
    return this.marcaActual()?.development ? `v${version} · dev` : `v${version}`;
  });

  /** Fuente correspondiente y licencia, únicamente cuando el backend entrega una URL HTTPS válida. */
  readonly fuente = computed(() => {
    const marca = this.marcaActual();
    const url = marca?.sourceUrl?.trim() ?? '';
    const licencia = marca?.license?.trim() ?? '';
    if (!licencia || !esURLHTTPS(url)) {
      return null;
    }
    return { url, licencia };
  });

  /**
   * La zona horaria de la instalación, tal cual, lista para `Intl`.
   *
   * Vacía mientras no haya llegado la marca —o si la instalación no responde—, y **entonces las
   * fechas se leen con la del navegador**, que es lo que se hacía antes: quien formatea se encarga
   * de caer ahí, así que la pantalla nunca se queda sin pintar. Si la zona llegara con espacios de
   * más, se recortan; si llegara un nombre que `Intl` no conoce, el formateador también cae a la del
   * navegador (`shared/fechas.ts`).
   */
  readonly zonaHoraria = computed(() => this.marcaActual()?.timeZone?.trim() ?? '');

  /**
   * Lee la marca y la aplica.
   *
   * **Si falla, no pasa nada**: se queda el color de fábrica y el logo de fábrica, y la aplicación
   * funciona igual. La marca es importante, pero no es una función crítica.
   */
  async cargar(): Promise<void> {
    try {
      const marca = await firstValueFrom(this.http.get<Marca>('/api/settings/brand'));
      this.marcaActual.set(marca);
      this.aplicarColor(marca);
      if (marca.language) {
        this.textos.adoptarGlobal(marca.language, marca.settingsVersion ?? 0);
      }
    } catch {
      // Sin marca se sigue con lo de fábrica: ni un aviso al usuario ni un error en pantalla.
    }
  }

  /**
   * La dirección del logo que toca a ese tema.
   *
   * Si no hay logo propio, devuelve **el de fábrica de ese tema**, que viene con la aplicación: así
   * la pantalla siempre tiene algo que enseñar, y además se lee sobre el fondo que hay debajo.
   */
  urlDelLogo(oscuro: boolean): string {
    const marca = this.marcaActual();
    if (marca === null || !this.hayLogoPropio()) {
      return logoDeFabrica(oscuro);
    }

    const tema = oscuro ? 'oscuro' : 'claro';

    // El sello va en la dirección: sin él, un navegador con el logo viejo guardado seguiría
    // enseñándolo después de reemplazarlo, y parecería que el cambio no funcionó.
    const version = marca.logoVersion ? `&v=${encodeURIComponent(marca.logoVersion)}` : '';

    return `/api/settings/brand/logo?theme=${tema}${version}`;
  }

  /**
   * Sube el logo elegido a su hueco —claro u oscuro— y devuelve la configuración nueva.
   *
   * **El logo es de la marca**, así que vive aquí, con el resto del recurso
   * (`/api/settings/brand/logo`). El hueco viaja como `variant`, que es como lo llama el backend
   * (`oscuro` o `claro`).
   *
   * Lo que devuelve es el documento de configuración entero —es lo que contesta el backend—, y por
   * eso el tipo se importa de `settings.service`: es la única pieza que cruza, y sólo como tipo.
   */
  async subirLogo(oscuro: boolean, archivo: File): Promise<Configuracion> {
    const datos = new FormData();
    datos.append('logo', archivo, archivo.name);

    return firstValueFrom(
      this.http.post<Configuracion>(
        `/api/settings/brand/logo?variant=${oscuro ? 'oscuro' : 'claro'}`,
        datos,
      ),
    );
  }

  /** Quita el logo propio de un hueco y lo devuelve al de fábrica. */
  async restablecerLogo(oscuro: boolean): Promise<Configuracion> {
    return firstValueFrom(
      this.http.delete<Configuracion>(
        `/api/settings/brand/logo?variant=${oscuro ? 'oscuro' : 'claro'}`,
      ),
    );
  }

  /**
   * Pide al backend cómo quedaría ese color, **sin guardarlo**.
   *
   * La vista previa la resuelve el backend a propósito: la regla de qué se lee y qué no es una sola,
   * y tener una copia en el frontend es la forma segura de que un día digan cosas distintas. Devuelve
   * **la marca entera**, que es lo que la pantalla lee para pintar la vista previa.
   */
  async previsualizarColor(valor: string): Promise<Marca> {
    return firstValueFrom(
      this.http.get<Marca>(`/api/settings/brand?color=${encodeURIComponent(valor)}`),
    );
  }

  /**
   * Vuelve a leer la marca: es lo que hace que el nombre y el logo nuevos se vean **sin recargar**,
   * en cuanto el administrador guarda.
   */
  async recargar(): Promise<void> {
    await this.cargar();
  }

  /**
   * Pone el color institucional en las variables del tema.
   *
   * El backend manda cuatro valores —claro, oscuro y el texto que va encima de cada uno— porque el
   * administrador elige **un** color y hay que resolverlo para los dos temas sin romper el contraste.
   */
  private aplicarColor(marca: Marca): void {
    const raiz = document.documentElement.style;

    raiz.setProperty('--institucional-claro', marca.colors.light);
    raiz.setProperty('--institucional-oscuro', marca.colors.dark);
    raiz.setProperty('--institucional-texto-claro', marca.colors.onLight);
    raiz.setProperty('--institucional-texto-oscuro', marca.colors.onDark);
  }
}

function esURLHTTPS(valor: string): boolean {
  try {
    return new URL(valor).protocol === 'https:';
  } catch {
    return false;
  }
}
