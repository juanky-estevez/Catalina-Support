import { Component, input } from '@angular/core';

/**
 * Etiqueta de estado: el estado de un ticket, con su color **y su texto**
 * (`docs/interfaz-y-experiencia.md`, sección 6.3).
 *
 * **Nunca dice nada sólo con el color** (sección 8): el color ayuda a barrer una lista, y el texto es
 * lo que dice qué es. Quien no distinga los colores —o quien use un lector de pantalla— tiene que
 * poder leer el estado igual.
 *
 * Los colores salen del tema, como todos: aquí no hay ningún color escrito a mano.
 */
@Component({
  selector: 'app-estado',
  template: `
    <span [class]="clases()">{{ texto() }}</span>
  `,
})
export class Estado {
  /** El estado de verdad, que es lo que decide el color: `nuevo`, `en espera`… */
  readonly estado = input.required<string>();
  /** Lo que se lee, que no siempre es el estado de verdad: el usuario lee «En curso», no «escalado». */
  readonly texto = input.required<string>();

  protected clases(): string {
    const base =
      'inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium whitespace-nowrap';

    const formas: Record<string, string> = {
      nuevo: 'border-primario text-primario',
      'en progreso': 'border-aviso bg-aviso-suave text-aviso',
      'en espera': 'border-borde bg-superficie-suave text-apagado',
      escalado: 'border-peligro bg-peligro-suave text-peligro',
      resuelto: 'border-exito bg-exito-suave text-exito',
      cerrado: 'border-borde bg-superficie text-apagado',
    };

    return [base, formas[this.estado()] ?? 'border-borde bg-superficie-suave text-apagado'].join(' ');
  }
}
