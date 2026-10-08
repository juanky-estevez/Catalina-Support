import { Component, computed, inject, input } from '@angular/core';

import { TranslationService } from '../i18n/translation.service';
import { ThemeService, TEMAS_DE_FABRICA, type Tema } from '../services/theme.service';
import { Selector, type OpcionSelector } from '../../shared/components/selector';

/**
 * Los controles que cada persona puede tocar: **el idioma y el tema**.
 *
 * Son del armazón y no de ninguna pantalla: los usan la **entrada** —en fila, antes de entrar— y el
 * **menú lateral** —en columna, en su zona de controles— (`docs/interfaz-y-experiencia.md`, secciones
 * 3.1 y 6.2). Están juntos porque son lo mismo: ajustes de quien mira, no de lo que mira.
 */
@Component({
  selector: 'app-controles',
  imports: [Selector],
  template: `
    <!--
      **Los dos, del ancho de lo que tengan al lado** (decisión del responsable, 2026-09-28): en la
      entrada ocupaban lo que medía su texto —idioma 104 px y tema 140, medidos— y no cuadraban ni con
      el formulario ni entre ellos. Ahora la fila ocupa el ancho entero y **los dos se reparten a
      medias**, que es como se pide «simétrico».
    -->
    <div class="flex w-full items-stretch gap-3" [class.flex-col]="vertical()">
      <!--
        **El idioma y el tema son el mismo control** (decisión del responsable, 2026-09-26): los dos son
        desplegables, ocupan el mismo ancho y ninguno lleva la etiqueta a la vista —el valor ya dice qué
        es—, aunque los dos la conservan para quien navega con lector de pantalla.
      -->
      <app-selector
        class="flex-1"
        identificador="tema"
        [etiqueta]="t().tema.etiqueta"
        [etiquetaVisible]="false"
        [anchoCompleto]="true"
        [opciones]="opcionesDeTema()"
        [valor]="tema.tema()"
        (valorChange)="cambiarTema($event)"
      />
    </div>
  `,
})
export class Controles {
  /** En el menú lateral van en columna; en la pantalla de entrada, en fila y centrados. */
  readonly vertical = input(false);

  protected readonly textos = inject(TranslationService);
  protected readonly tema = inject(ThemeService);

  protected t() {
    return this.textos.textos();
  }

  /**
   * Los ocho temas, agrupados en claros y oscuros.
   *
   * **«Automático» no está en la lista** (decisión del responsable, 2026-09-23): seguir al sistema es
   * no haber elegido, y mientras nadie elija el selector enseña el tema que se está viendo.
   *
   * Los dos primeros de cada grupo son **los de fábrica**, los que llevan el color institucional que
   * fija el administrador.
   */
  protected readonly opcionesDeTema = computed<readonly OpcionSelector[]>(() => {
    const nombres: Record<Tema, string> = {
      claro: this.t().tema.claro,
      oscuro: this.t().tema.oscuro,
      papel: this.t().tema.papel,
      niebla: this.t().tema.niebla,
      contraste: this.t().tema.contraste,
      grafito: this.t().tema.grafito,
      noche: this.t().tema.noche,
      sepia: this.t().tema.sepia,
    };

    // El grupo de cada tema es su claridad, y los de fábrica van primero en el suyo.
    const claridad: Record<Tema, string> = {
      claro: 'claros',
      papel: 'claros',
      niebla: 'claros',
      contraste: 'claros',
      oscuro: 'oscuros',
      grafito: 'oscuros',
      noche: 'oscuros',
      sepia: 'oscuros',
    };

    const orden: Tema[] = [
      ...TEMAS_DE_FABRICA,
      'papel',
      'niebla',
      'contraste',
      'grafito',
      'noche',
      'sepia',
    ];

    return orden.map((tema) => ({
      valor: tema,
      etiqueta: nombres[tema],
      grupo: claridad[tema] === 'claros' ? this.t().tema.claros : this.t().tema.oscuros,
    }));
  });

  protected cambiarTema(valor: string): void {
    this.tema.cambiar(valor as Tema);
  }
}
