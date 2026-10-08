import { Component, computed, inject, input } from '@angular/core';

import { TranslationService } from '../i18n/translation.service';
import { BrandService } from '../services/brand.service';
import { ThemeService } from '../services/theme.service';

/** Dónde se está enseñando el logo, que es lo que decide su tamaño (interfaz-y-experiencia, 6.4). */
export type TamanoLogo = 'entrada' | 'menu' | 'cabecera';

/**
 * El logo de la instalación.
 *
 * **Sirve para cualquier logo**, y eso se resuelve en cómo se enseña, no pidiendo un archivo
 * concreto (`docs/interfaz-y-experiencia.md`, sección 6.4):
 *
 * - **No se recorta y no se estira**: se enseña entero, con su proporción, dentro del alto máximo.
 * - **Va en un recuadro de esquinas redondeadas y con un borde del tema**, así un logo con fondo
 *   opaco se ve como una pieza deliberada y uno con transparencia se integra en el fondo.
 * - **Elige el hueco que toca al tema que se está viendo**, y si ese hueco está vacío usa el otro; y
 *   si no hay ninguno propio, el de fábrica.
 */
@Component({
  selector: 'app-logo',
  template: `
    <span [class]="marco()">
      <img [src]="direccion()" [alt]="alternativo()" [class]="imagen()" />
    </span>
  `,
})
export class Logo {
  private readonly marca = inject(BrandService);
  private readonly tema = inject(ThemeService);
  private readonly textos = inject(TranslationService);

  /** Dónde se enseña: en la entrada, o en el menú lateral. */
  readonly tamano = input<TamanoLogo>('entrada');

  protected readonly direccion = computed(() => this.marca.urlDelLogo(this.tema.esOscuro()));

  /** El texto alternativo es el nombre del producto: quien no ve la imagen tiene que saber qué es. */
  protected readonly alternativo = computed(() => this.textos.textos().entrada.descripcion);

  protected marco(): string {
    const base =
      'inline-flex items-center justify-center overflow-hidden rounded-md border border-borde bg-superficie';

    if (this.tamano() === 'menu') {
      return `${base} h-10 w-10 p-1`;
    }

    // En el primer arranque acompaña al título y a su explicación: 80 px ocupan el mismo bloque
    // visual que esas tres líneas sin convertirlo en el logo protagonista de la pantalla de entrada.
    if (this.tamano() === 'cabecera') {
      return `${base} h-20 w-20 p-2`;
    }

    // En la entrada es la primera cosa que se ve, así que va grande: 128 px de alto en móvil y 160 en
    // pantallas de tablet para arriba (docs/interfaz-y-experiencia.md, sección 6.4).
    //
    // **Y del ancho del formulario** (decisión del responsable, 2026-09-28): tenía tope propio —288 px y
    // 320 px—, así que el recuadro del logo quedaba **más estrecho que la tarjeta del formulario** y la
    // entrada se veía desalineada. Medido: 136 px de logo contra 329 de formulario. Es `w-full` y el
    // contenedor decide: los dos son el mismo ancho, con el logo entero dentro.
    return `${base} h-32 w-full px-4 py-3 sm:h-40`;
  }

  protected imagen(): string {
    return this.tamano() === 'menu' || this.tamano() === 'cabecera'
      ? 'h-full w-full object-contain'
      : 'max-h-full max-w-full object-contain';
  }
}
