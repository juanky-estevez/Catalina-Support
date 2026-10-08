import { Component, input, output } from '@angular/core';

import { Icono } from './icono';

/** Las cuatro formas de un aviso. */
export type FormaDeAviso = 'informacion' | 'exito' | 'error' | 'atencion';

/**
 * Lo que una pantalla enseña cuando algo ha pasado: un aviso y su forma.
 *
 * Es un tipo y no un par de campos sueltos porque **todo lo que contesta una acción se enseña igual**
 * —el alta, un cambio guardado, un error del backend o un aviso de algo que sí se hizo—, y así ninguna
 * pantalla se inventa su propia manera de decirlo.
 */
export interface MensajeDePantalla {
  readonly forma: FormaDeAviso;
  readonly texto: string;
}

/**
 * Aviso: éxito, error o información (`docs/interfaz-y-experiencia.md`, sección 6.3).
 *
 * **Nunca dice nada sólo con color**: cada forma lleva su icono y su texto, porque el color no se
 * ve igual para todo el mundo (sección 8). Error y atención se anuncian solos con `role="alert"`;
 * éxito e información usan `role="status"`.
 */
@Component({
  selector: 'app-aviso',
  imports: [Icono],
  template: `
    <div [class]="clases()" [attr.role]="rol()">
      <span aria-hidden="true">{{ icono() }}</span>
      <span class="min-w-0 flex-1">{{ texto() }}</span>
      @if (etiquetaCerrar(); as etiqueta) {
        <button
          type="button"
          class="-m-1 ml-auto inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md hover:bg-fondo focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primario"
          [attr.aria-label]="etiqueta"
          (click)="cerrado.emit()"
        >
          <app-icono nombre="cerrar" />
        </button>
      }
    </div>
  `,
})
export class Aviso {
  readonly texto = input.required<string>();
  readonly forma = input<FormaDeAviso>('informacion');
  readonly etiquetaCerrar = input<string | null>(null);
  readonly cerrado = output<void>();

  protected rol(): 'alert' | 'status' {
    return this.forma() === 'error' || this.forma() === 'atencion' ? 'alert' : 'status';
  }

  protected icono(): string {
    const iconos: Record<string, string> = {
      informacion: 'ℹ',
      exito: '✓',
      error: '✕',
      atencion: '!',
    };
    return iconos[this.forma()] ?? 'ℹ';
  }

  protected clases(): string {
    const base = 'flex items-start gap-2 rounded-md border px-3 py-2 text-sm';

    const formas: Record<string, string> = {
      informacion: 'border-borde bg-superficie-suave text-texto',
      exito: 'border-exito bg-exito-suave text-exito',
      error: 'border-peligro bg-peligro-suave text-peligro',
      atencion: 'border-aviso bg-aviso-suave text-aviso',
    };

    return [base, formas[this.forma()]].join(' ');
  }
}
