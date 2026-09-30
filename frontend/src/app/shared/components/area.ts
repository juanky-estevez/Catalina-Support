import { Component, input, model } from '@angular/core';

/**
 * Área de texto: el del inventario de componentes propios
 * (`docs/interfaz-y-experiencia.md`, sección 6.3).
 *
 * Es para lo que se escribe en varios renglones —la descripción de un ticket, un comentario, el motivo
 * de un escalado—, y va con **etiqueta de verdad**: un texto de ejemplo que desaparece al escribir no
 * dice qué se está escribiendo, y esta es la caja donde se cuenta lo que pasa (sección 8).
 *
 * Como el campo de texto, **el borde es el fuerte** (el que llega a 3:1 de contraste): es lo que hace
 * visible que ahí se escribe.
 */
@Component({
  selector: 'app-area',
  template: `
    <div class="flex flex-col gap-1">
      <label [for]="identificador()" class="text-sm font-medium text-texto">{{ etiqueta() }}</label>

      <textarea
        [id]="identificador()"
        [rows]="filas()"
        [attr.aria-invalid]="error() ? 'true' : null"
        [attr.aria-describedby]="descritoPor()"
        [disabled]="deshabilitado()"
        [value]="valor()"
        (input)="alEscribir($event)"
        class="rounded-md border border-borde-fuerte bg-superficie px-3 py-2 text-sm text-texto"
      ></textarea>

      @if (error()) {
        <p [id]="identificador() + '-error'" class="text-sm text-peligro">{{ error() }}</p>
      } @else if (ayuda()) {
        <p [id]="identificador() + '-ayuda'" class="text-sm text-apagado">{{ ayuda() }}</p>
      }
    </div>
  `,
})
export class Area {
  readonly identificador = input.required<string>();
  readonly etiqueta = input.required<string>();
  readonly ayuda = input('');
  readonly error = input('');
  readonly filas = input(5);
  readonly deshabilitado = input(false);

  /** El texto, en dos direcciones: la pantalla lo lee y lo escribe. */
  readonly valor = model<string>('');

  protected alEscribir(evento: Event): void {
    this.valor.set((evento.target as HTMLTextAreaElement).value);
  }

  protected descritoPor(): string | null {
    if (this.error()) {
      return `${this.identificador()}-error`;
    }

    return this.ayuda() ? `${this.identificador()}-ayuda` : null;
  }
}
