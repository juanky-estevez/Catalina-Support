import { Component, effect, input, output } from '@angular/core';

import { Aviso, type MensajeDePantalla } from './aviso';

/**
 * Un resultado transitorio que permanece visible aunque la acción ocurra al final de una pantalla.
 *
 * Sólo éxito e información se retiran solos. Los fallos necesitan que quien los lee pueda conservar
 * el texto mientras corrige lo que pasó (`docs/interfaz-y-experiencia.md`, sección 15).
 */
@Component({
  selector: 'app-toast',
  imports: [Aviso],
  host: {
    class:
      'fixed inset-x-4 top-4 z-50 block sm:left-auto sm:right-4 sm:w-full sm:max-w-md',
  },
  template: `
    <app-aviso
      [forma]="mensaje().forma"
      [texto]="mensaje().texto"
      [etiquetaCerrar]="etiquetaCerrar()"
      (cerrado)="cerrar()"
    />
  `,
})
export class Toast {
  readonly mensaje = input.required<MensajeDePantalla>();
  readonly etiquetaCerrar = input.required<string>();
  readonly cerrado = output<void>();

  constructor() {
    /** Cada mensaje nuevo sustituye al anterior y, si corresponde, estrena sus cinco segundos. */
    effect((limpiar) => {
      const forma = this.mensaje().forma;
      if (forma !== 'exito' && forma !== 'informacion') {
        return;
      }

      const temporizador = window.setTimeout(() => this.cerrar(), 5_000);
      limpiar(() => window.clearTimeout(temporizador));
    });
  }

  protected cerrar(): void {
    this.cerrado.emit();
  }
}
