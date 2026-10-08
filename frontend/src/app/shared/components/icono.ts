import { Component, input } from '@angular/core';

/**
 * Los iconos del armazón y los del ticket, dibujados aquí: **no se añaden bibliotecas de iconos**.
 *
 * Los del ticket entraron el 2026-09-28 (decisión 76) porque los caracteres que había antes —`⧉` para
 * copiar y `✏` para editar— **no se distinguían**: el primero es un cuadrado con otro detrás, que a
 * 16 px se lee como un cuadrado, y el segundo sale como emoji de colores según la fuente del equipo.
 */
export type NombreDeIcono =
  | 'tickets'
  | 'principales'
  | 'internos'
  | 'categorias'
  | 'usuarios'
  | 'configuracion'
  | 'menu'
  | 'plegar'
  | 'salir'
  | 'cerrar'
  // Los del ticket (decisión 76): copiar el número, confirmar que se ha copiado y editar.
  | 'copiar'
  | 'visto'
  | 'editar'
  // El destello identifica la ayuda de redacción con IA sin depender de una biblioteca externa.
  | 'ia'
  // **La flecha de volver** (quinta enmienda del 2026-09-29): una flecha hacia la izquierda, cuadrada,
  // para el botón de volver del ticket y de la ficha de una cuenta.
  | 'volver';

/**
 * Un icono del armazón.
 *
 * Son **svg propios y mínimos**, con `aria-hidden`: el icono no dice nada que no diga el texto que
 * lleva al lado, y un lector de pantalla no tiene por qué anunciarlo. No se añade ninguna biblioteca
 * de iconos, por la misma razón por la que no se añaden de componentes.
 */
@Component({
  selector: 'app-icono',
  template: `
    <svg
      aria-hidden="true"
      [attr.data-icono]="nombre()"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="1.7"
      stroke-linecap="round"
      stroke-linejoin="round"
      class="h-5 w-5 shrink-0"
    >
      @switch (nombre()) {
        @case ('tickets') {
          <path d="M4 6h16v4a2 2 0 0 0 0 4v4H4v-4a2 2 0 0 0 0-4z" />
          <path d="M10 6v12" />
        }
        @case ('principales') {
          <path d="M4 5h16v4a2 2 0 0 0 0 4v4H4v-4a2 2 0 0 0 0-4z" />
          <path d="M8 9.5h8M8 13.5h5" />
        }
        @case ('internos') {
          <path d="M5 5h14v4a2 2 0 0 0 0 4v4H5v-4a2 2 0 0 0 0-4z" />
          <path d="M12 8.5v6M9.5 11l2.5-2.5 2.5 2.5" />
        }
        @case ('categorias') {
          <path d="M4 5h7v7H4zM13 5h7v4h-7zM13 11h7v8h-7zM4 14h7v5H4z" />
        }
        @case ('usuarios') {
          <circle cx="9" cy="8" r="3.2" />
          <path d="M3.5 20c0-3.3 2.5-5.5 5.5-5.5s5.5 2.2 5.5 5.5" />
          <path d="M16 5.5a3 3 0 0 1 0 5.8" />
          <path d="M17.5 14.8c2 .6 3.1 2.4 3.1 4.4" />
        }
        @case ('configuracion') {
          <circle cx="12" cy="12" r="3" />
          <path
            d="M12 2.5v3M12 18.5v3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M2.5 12h3M18.5 12h3M4.9 19.1 7 17M17 7l2.1-2.1"
          />
        }
        @case ('menu') {
          <path d="M4 6h16M4 12h16M4 18h16" />
        }
        @case ('plegar') {
          <path d="M15 5l-7 7 7 7" />
        }
        @case ('salir') {
          <path d="M10 4H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h4" />
          <path d="M16 8l4 4-4 4M20 12H9" />
        }
        @case ('cerrar') {
          <path d="M6 6l12 12M18 6L6 18" />
        }
        @case ('copiar') {
          <!-- **Dos hojas superpuestas**, el icono de copiar de toda la vida: se distingue del lápiz por
               la forma, no por el trazo, que es lo que fallaba con el carácter de antes (⧉). -->
          <rect x="9" y="9" width="11" height="12" rx="2" />
          <path d="M6 15H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v1" />
        }
        @case ('visto') {
          <path d="M5 12.5 10 17.5 19 7" />
        }
        @case ('editar') {
          <path d="M4 20h4l10-10a2.5 2.5 0 0 0-3.5-3.5L4.5 16.5z" />
          <path d="M13.5 6.5 17.5 10.5" />
        }
        @case ('ia') {
          <path d="M12 3l1.1 3.4L16.5 7.5l-3.4 1.1L12 12l-1.1-3.4-3.4-1.1 3.4-1.1z" />
          <path d="M18 13l.8 2.2L21 16l-2.2.8L18 19l-.8-2.2L15 16l2.2-.8z" />
          <path d="M6 14l.7 1.8 1.8.7-1.8.7L6 19l-.7-1.8-1.8-.7 1.8-.7z" />
        }
        @case ('volver') {
          <!-- **Una flecha hacia la izquierda**: el mismo trazo de los demás svg del proyecto, con su
               varilla, para que se lea como «volver» y no como el galón de plegar el menú. -->
          <path d="M20 12H4" />
          <path d="M10 6l-6 6 6 6" />
        }
      }
    </svg>
  `,
})
export class Icono {
  readonly nombre = input.required<NombreDeIcono>();
}
