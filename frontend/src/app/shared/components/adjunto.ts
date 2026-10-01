import { Component, input, output } from '@angular/core';

/** Un adjunto, tal y como lo cuenta la API (`docs/modules/tickets.md`, sección 2.3). */
export interface AdjuntoDeTicket {
  readonly id: number;
  readonly filename: string;
  readonly contentType: string;
  readonly size: number;
  readonly commentId?: number;
  readonly uploadedBy?: { name: string; lastName: string };
  readonly createdAt: string;
}

/**
 * Adjunto: el del inventario de componentes propios, **con su vista previa**
 * (`docs/interfaz-y-experiencia.md`, sección 6.3).
 *
 * Enseña **el nombre y el peso antes de descargar** (sección 8): nadie tiene que bajarse 20 MB para
 * saber qué era. Y ofrece **ver** lo que se puede ver —imágenes y PDF—, porque abrir una captura
 * dentro del ticket es lo que se hace el 90 % de las veces.
 *
 * **No descarga nada por su cuenta.** Guardar un archivo o enseñarlo son cosas de la pantalla, y este
 * componente vive en `shared`, que no puede depender de ningún módulo: emite lo que se ha pedido y
 * quien lo usa se encarga.
 */
@Component({
  selector: 'app-adjunto',
  template: `
    <div
      class="flex flex-wrap items-center gap-2 rounded-md border border-borde bg-superficie px-2 py-1.5"
    >
      <span class="truncate text-sm text-texto" [attr.title]="adjunto().filename">
        {{ adjunto().filename }}
      </span>

      <span class="text-xs text-apagado">{{ peso() }}</span>

      @if (sePuedeVer()) {
        <button
          type="button"
          class="ml-auto inline-flex min-h-9 items-center rounded-md border border-borde-fuerte px-2 text-sm text-texto hover:bg-superficie-suave"
          (click)="ver.emit(adjunto())"
        >
          {{ verTexto() }}
        </button>
      }

      <button
        type="button"
        class="inline-flex min-h-9 items-center rounded-md border border-borde-fuerte px-2 text-sm text-texto hover:bg-superficie-suave"
        [class.ml-auto]="!sePuedeVer()"
        (click)="descargar.emit(adjunto())"
      >
        {{ descargarTexto() }}
      </button>
    </div>
  `,
})
export class Adjunto {
  readonly adjunto = input.required<AdjuntoDeTicket>();
  /** Si se puede ver dentro del ticket: lo dice la pantalla, que conoce la lista de tipos. */
  readonly previsualizable = input(false);
  readonly verTexto = input.required<string>();
  readonly descargarTexto = input.required<string>();

  readonly ver = output<AdjuntoDeTicket>();
  readonly descargar = output<AdjuntoDeTicket>();

  protected sePuedeVer(): boolean {
    return this.previsualizable();
  }

  /** El peso en algo que se lee: «203 KB», «1,2 MB». */
  protected peso(): string {
    const bytes = this.adjunto().size;

    if (bytes < 1024) {
      return `${bytes} B`;
    }
    if (bytes < 1024 * 1024) {
      return `${Math.round(bytes / 1024)} KB`;
    }

    return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  }
}
