import { Component, input } from '@angular/core';

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
 * ve igual para todo el mundo (sección 8). El aviso de error se anuncia solo
 * (`role="alert"`), que es lo que hace que un lector de pantalla lo lea sin que nadie lo busque.
 */
@Component({
  selector: 'app-aviso',
  template: `
    <div [class]="clases()" [attr.role]="forma() === 'error' ? 'alert' : 'status'">
      <span aria-hidden="true">{{ icono() }}</span>
      <span>{{ texto() }}</span>
    </div>
  `,
})
export class Aviso {
  readonly texto = input.required<string>();
  readonly forma = input<FormaDeAviso>('informacion');

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
