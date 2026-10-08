import { Component, input, model } from '@angular/core';

/** Una opción del conmutador: lo que se guarda y lo que se lee. */
export interface OpcionConmutador {
  readonly valor: string;
  readonly etiqueta: string;
}

/**
 * Conmutador segmentado: el del inventario de componentes propios
 * (`docs/interfaz-y-experiencia.md`, sección 6.3). Es el que usa la vista doble, los filtros y —hoy—
 * el idioma y el tema.
 *
 * **Es un grupo de botones, no una lista de pestañas**: se navega con el tabulador, se activa con
 * Enter o con la barra espaciadora, y cada opción dice si está puesta con `aria-pressed`, que es lo
 * que hace que un lector de pantalla anuncie «Idioma: Español, pulsado».
 *
 * Los colores salen del tema: la opción elegida usa el color de acento y las demás el texto apagado,
 * **y además lleva un borde**, para que no sea sólo el color lo que dice cuál está puesta
 * (sección 8).
 */
@Component({
  selector: 'app-conmutador',
  template: `
    <div
      role="group"
      [attr.aria-label]="etiqueta()"
      class="inline-flex max-w-full flex-wrap rounded-md border border-borde bg-superficie-suave p-0.5"
    >
      @for (opcion of opciones(); track opcion.valor) {
        <button
          type="button"
          [attr.aria-pressed]="opcion.valor === valor()"
          [class]="clases(opcion)"
          (click)="elegir(opcion.valor)"
        >
          {{ opcion.etiqueta }}
        </button>
      }
    </div>
  `,
})
export class Conmutador {
  /** Para qué es: «Idioma», «Tema». Es lo que se anuncia al entrar en el grupo. */
  readonly etiqueta = input.required<string>();
  readonly opciones = input.required<readonly OpcionConmutador[]>();

  /** La opción puesta. Se puede leer y escribir: `[(valor)]="tema"`. */
  readonly valor = model.required<string>();

  protected elegir(valor: string): void {
    this.valor.set(valor);
  }

  protected clases(opcion: OpcionConmutador): string {
    const base =
      'rounded px-3 py-1.5 text-sm min-h-9 transition ' +
      'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foco';

    return opcion.valor === this.valor()
      ? `${base} bg-superficie text-texto font-medium shadow-sm ring-1 ring-borde-fuerte`
      : `${base} text-apagado hover:text-texto`;
  }
}
