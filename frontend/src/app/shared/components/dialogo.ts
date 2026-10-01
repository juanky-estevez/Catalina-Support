import { Component, ElementRef, effect, input, output, viewChild } from '@angular/core';

/**
 * Diálogo: el del inventario de componentes propios
 * (`docs/interfaz-y-experiencia.md`, sección 6.3).
 *
 * Es un **`<dialog>` nativo**, y no un `<div>` con aspecto de ventana, porque el navegador ya trae lo
 * que un diálogo tiene que hacer: se cierra con **Escape**, atrapa el foco mientras está abierto y lo
 * **devuelve a donde estaba** al cerrarse (sección 8). Construir eso a mano es la forma segura de
 * hacerlo mal.
 *
 * El contenido lo pone quien lo usa: aquí sólo está la ventana, su título y su botón de cerrar.
 */
@Component({
  selector: 'app-dialogo',
  template: `
    <dialog #ventana [attr.aria-label]="titulo()" (close)="cerrado.emit()" [class]="clases()">
      <div class="flex items-start gap-4 border-b border-borde px-5 py-4">
        <h2 class="flex-1 text-base font-semibold text-texto">{{ titulo() }}</h2>

        <button
          type="button"
          class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-borde-fuerte text-apagado hover:text-texto"
          [attr.aria-label]="etiquetaCerrar()"
          (click)="cerrar()"
        >
          <span aria-hidden="true">✕</span>
        </button>
      </div>

      <div class="px-5 py-4">
        <ng-content />
      </div>
    </dialog>
  `,
})
export class Dialogo {
  readonly titulo = input.required<string>();
  readonly etiquetaCerrar = input('Cerrar');
  /** Si se enseña. Al ponerse en falso, se cierra. */
  readonly abierto = input(false);

  /**
   * Lo ancho que es. `normal` es una ventana de formulario —el alta de una cuenta, una acción del
   * ticket—, y **`grande` es el visor de un adjunto** (decisión 48): una imagen o un PDF necesitan
   * sitio para verse, y el ancho de una ventana de formulario no lo es.
   */
  readonly tamano = input<'normal' | 'grande'>('normal');

  readonly cerrado = output<void>();

  // No es `required`: la ventana no existe hasta que la vista está creada, y el efecto de abajo se
  // ejecuta también antes de eso.
  private readonly ventana = viewChild<ElementRef<HTMLDialogElement>>('ventana');

  constructor() {
    effect(() => {
      const elemento = this.ventana()?.nativeElement;
      if (!elemento) {
        return;
      }

      if (this.abierto() && !elemento.open) {
        elemento.showModal();
      } else if (!this.abierto() && elemento.open) {
        elemento.close();
      }
    });
  }

  /** Cierra la ventana: el navegador avisa por `close`, y ese aviso es el que sale hacia fuera. */
  protected cerrar(): void {
    this.ventana()?.nativeElement.close();
  }

  protected clases(): string {
    // **`m-auto` es obligatorio, y no es cosmético** (lo reportó el responsable el 2026-09-28): el
    // `<dialog>` nativo se centra con `margin: auto`, que viene en la hoja del navegador, y **la
    // preparación de Tailwind pone `margin: 0` a todo** —`*, ::before, ::after`—, así que la ventana se
    // iba a la esquina. Medido: salía en `x: 0, y: 0`. Devolver el `auto` es lo que la centra, y
    // `max-h` con desplazamiento propio es lo que evita que una ventana alta se salga por abajo.
    const base =
      'm-auto max-h-[calc(100dvh-2rem)] max-w-[calc(100vw-2rem)] overflow-auto rounded-lg border border-borde bg-superficie p-0 text-texto backdrop:bg-black/40';
    const anchos: Record<string, string> = {
      normal: 'w-[min(32rem,calc(100vw-2rem))]',
      grande: 'w-[min(72rem,calc(100vw-2rem))]',
    };

    return [base, anchos[this.tamano()] ?? anchos['normal']].join(' ');
  }
}
