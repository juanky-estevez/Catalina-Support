import { Injectable, effect, inject } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { TitleStrategy } from '@angular/router';

import { BrandService } from './brand.service';

/**
 * El título de la pestaña del navegador: **el nombre de la instalación**.
 *
 * Las rutas traían un título escrito a mano —«Catalina Support» en todas—, que es el nombre de
 * fábrica: una instalación que se llama de otra manera no puede llevar el nombre de otra empresa en
 * la pestaña. Aquí se sustituye por el nombre de verdad, que viene con la marca
 * (`docs/modules/settings.md`).
 *
 * **No se pierde nada por no usar el título de la ruta**: todas las rutas llevan el mismo, así que
 * el título nunca decía en qué pantalla estabas. Si algún día una pantalla quiere su propio título,
 * este es el sitio donde decidirlo.
 *
 * El `effect` es lo que hace que la pestaña cambie **en cuanto se guarda el nombre nuevo** en
 * Configuración, sin recargar; y mientras la marca no ha llegado —o si la instalación no contesta—
 * el nombre es el de fábrica, así que la pestaña nunca se queda sin nada.
 */
@Injectable({ providedIn: 'root' })
export class TituloDeLaInstalacion extends TitleStrategy {
  private readonly titulo = inject(Title);
  private readonly marca = inject(BrandService);

  constructor() {
    super();

    effect(() => this.titulo.setTitle(this.marca.nombre()));
  }

  override updateTitle(): void {
    this.titulo.setTitle(this.marca.nombre());
  }
}
