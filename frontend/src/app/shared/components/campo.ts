import { Component, input, model, output } from '@angular/core';

/**
 * Campo de texto: etiqueta de verdad, ayuda y error debajo.
 *
 * **La etiqueta no es un texto de ejemplo que desaparece al escribir** (`<label>` de verdad, atado
 * al campo): es lo que hace que un lector de pantalla diga qué se está escribiendo, y lo que
 * permite repasar el formulario sin perder el sitio (docs/interfaz-y-experiencia.md, sección 8).
 *
 * El error se anuncia con `aria-describedby` y `aria-invalid`, para que quien no ve la pantalla se
 * entere de que algo va mal sin salir del campo.
 *
 * **El borde es el fuerte** (el que llega a 3:1 de contraste): es lo que hace visible que ahí se
 * escribe, y la norma de contraste lo exige para lo que identifica un control. El borde claro se
 * queda para lo que es decoración —tarjetas, separadores—, y así la aplicación no parece una
 * cuadrícula.
 *
 * **Se llama `identificador` y no `id` a propósito**: Angular deja los atributos estáticos en el
 * elemento anfitrión, así que un `id` aquí aparecía dos veces en la página —en el anfitrión y en el
 * campo—, y la etiqueta acababa apuntando al anfitrión. El campo se quedaba **sin nombre
 * accesible**: lo encontró Playwright, no las pruebas de unidad, porque sólo se ve en un navegador
 * de verdad.
 */
@Component({
  selector: 'app-campo',
  template: `
    <div class="flex flex-col gap-1">
      <label [for]="identificador()" class="text-sm font-medium text-texto">{{ etiqueta() }}</label>

      <input
        [id]="identificador()"
        [type]="tipo()"
        [autocomplete]="autocomplete()"
        [attr.placeholder]="ejemplo() || null"
        [attr.maxlength]="maximo() > 0 ? maximo() : null"
        [attr.aria-invalid]="error() ? 'true' : null"
        [attr.aria-describedby]="descritoPor()"
        [disabled]="deshabilitado()"
        [value]="valor()"
        (input)="alEscribir($event)"
        (blur)="alSalir.emit()"
        class="rounded-md border border-borde-fuerte bg-superficie px-3 py-2 text-sm text-texto min-h-11"
      />

      @if (error()) {
        <p [id]="identificador() + '-error'" class="text-sm text-peligro">{{ error() }}</p>
      } @else if (ayuda()) {
        <p [id]="identificador() + '-ayuda'" class="text-sm text-apagado">{{ ayuda() }}</p>
      }
    </div>
  `,
})
export class Campo {
  readonly identificador = input.required<string>();
  readonly etiqueta = input.required<string>();
  readonly tipo = input<'text' | 'email' | 'password'>('text');
  /** Lo que el navegador puede rellenar solo. `current-password` y `new-password` son los que
   *  hacen que un gestor de contraseñas proponga guardar (o no) lo que toca. */
  readonly autocomplete = input<string>('off');
  readonly ayuda = input<string>('');
  readonly error = input<string>('');
  readonly ejemplo = input<string>('');
  /**
   * El tope de caracteres, si lo hay.
   *
   * Va en el propio campo (`maxlength`) para que el navegador no deje escribir de más: es más
   * amable que dejar escribir y contestar con un error. El backend lo comprueba igual, porque un
   * tope que sólo vive en el navegador no es un tope.
   */
  readonly maximo = input(0);
  readonly deshabilitado = input(false);

  /** El valor del campo, en dos direcciones: la pantalla lo lee y lo escribe. */
  readonly valor = model<string>('');

  readonly alSalir = output<void>();

  protected alEscribir(evento: Event): void {
    this.valor.set((evento.target as HTMLInputElement).value);
  }

  protected descritoPor(): string | null {
    if (this.error()) {
      return `${this.identificador()}-error`;
    }
    return this.ayuda() ? `${this.identificador()}-ayuda` : null;
  }
}
