import { Component, input } from '@angular/core';

/**
 * Tarjeta: el contenedor de la ficha y, en móvil, de cada fila de una lista
 * (`docs/interfaz-y-experiencia.md`, sección 6.3).
 *
 * Su título es un encabezado de verdad, **y su nivel lo elige la pantalla**: en las pantallas de la
 * sesión la tarjeta *es* la pantalla y su título va como `h1`; en las pantallas de producto hay un
 * título de página (`Usuarios`, `Mi perfil`) y las tarjetas van debajo, así que el suyo es un `h2`.
 * Una página con tres `h1` no se navega con un lector de pantalla: lo encontró Playwright, mirando
 * los encabezados de verdad (sección 8).
 */
@Component({
  selector: 'app-tarjeta',
  template: `
    <section class="rounded-lg border border-borde bg-superficie p-6">
      @if (titulo()) {
        @if (nivel() === 'h1') {
          <h1 class="mb-1 text-xl font-semibold text-texto">{{ titulo() }}</h1>
        } @else {
          <h2 class="mb-1 text-lg font-semibold text-texto">{{ titulo() }}</h2>
        }
      }
      @if (descripcion()) {
        <p class="mb-4 text-sm text-apagado">{{ descripcion() }}</p>
      }
      <ng-content />
    </section>
  `,
})
export class Tarjeta {
  readonly titulo = input<string>('');
  readonly descripcion = input<string>('');
  /**
   * El nivel del título: `h1` cuando la tarjeta es la pantalla entera, `h2` cuando va debajo del
   * título de una pantalla de producto.
   */
  readonly nivel = input<'h1' | 'h2'>('h1');
}
