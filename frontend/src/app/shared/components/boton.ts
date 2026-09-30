import { Component, input, output } from '@angular/core';

import { Icono, type NombreDeIcono } from './icono';

/**
 * Botón: el del inventario de componentes propios (docs/interfaz-y-experiencia.md, sección 6.3).
 *
 * Cuatro formas —la principal, la secundaria, la peligrosa y **la de la confirmación**, que es verde y
 * dice «ha salido bien» sin sacar un aviso aparte— y un estado de «trabajando» que **deshabilita** el
 * botón: es lo que evita que alguien mande dos veces lo mismo por dar dos veces al botón.
 *
 * **Puede llevar un icono, y puede ser sólo un icono** (`soloIcono`), que entonces se queda cuadrado y
 * **necesita `etiqueta`**: un botón sin texto y sin nombre accesible es un botón que nadie que use un
 * lector de pantalla puede pulsar.
 *
 * Es un botón de verdad (`<button>`), no un enlace con aspecto de botón: así funciona con teclado,
 * con lector de pantalla y con el Enter sin que haya que añadir nada.
 *
 * Los colores salen del tema (clases que leen las variables CSS), nunca escritos aquí.
 */
@Component({
  selector: 'app-boton',
  imports: [Icono],
  template: `
    <button
      [type]="tipo()"
      [disabled]="deshabilitado() || trabajando()"
      [class]="clases()"
      [attr.aria-label]="etiqueta() || null"
      (click)="pulsado.emit()"
    >
      @if (nombreDeIcono(); as icono) {
        <span [class]="claseDelIcono()">
          <app-icono [nombre]="icono" />
        </span>
      }

      @if (!soloIcono()) {
        {{ trabajando() ? textoTrabajando() : texto() }}
      }
    </button>
  `,
})
export class Boton {
  /**
   * El texto del botón. **Puede ir vacío** cuando el botón es sólo un icono: entonces el nombre
   * accesible tiene que venir en `etiqueta`, o el botón no se puede anunciar.
   */
  readonly texto = input<string>('');
  /**
   * Un icono dibujado antes del texto, o **el único contenido** si `soloIcono`.
   *
   * Es **uno de los iconos del proyecto** (`app-icono`), no un carácter ni una imagen: va oculto a quien
   * lee con lector de pantalla —el nombre lo da el texto o la `etiqueta`— y hereda el color del botón,
   * así que sirve con los ocho temas. Los ejemplos vivos son los del ticket: copiar el número, el visto
   * verde que lo confirma, y el lápiz de editar.
   *
   * **Antes era un carácter suelto** y se cambió el 2026-09-28 (decisión 76): `⧉` y `✏` no se
   * distinguían, y dependían de la fuente de cada equipo.
   */
  readonly nombreDeIcono = input<NombreDeIcono | null>(null);
  /** Sin texto: el botón es cuadrado y el icono manda. Necesita `etiqueta`. */
  readonly soloIcono = input(false);
  /** El nombre accesible de un botón que no lleva texto (`aria-label`). */
  readonly etiqueta = input<string>('');
  /** El color del icono, cuando no es el del texto del botón. */
  readonly claseDelIcono = input<string>('');
  readonly textoTrabajando = input<string>('');
  readonly tipo = input<'button' | 'submit'>('button');
  readonly forma = input<'principal' | 'secundario' | 'peligro' | 'exito'>('principal');
  readonly ancho = input(false);
  readonly trabajando = input(false);
  readonly deshabilitado = input(false);

  readonly pulsado = output<void>();

  protected clases(): string {
    const base =
      'inline-flex items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition ' +
      'disabled:cursor-not-allowed disabled:opacity-60 min-h-11';

    const formas: Record<string, string> = {
      principal: 'bg-primario text-primario-texto hover:brightness-110',
      secundario: 'bg-superficie-suave text-texto border border-borde hover:bg-superficie',
      peligro: 'bg-peligro text-primario-texto hover:brightness-110',
      // **La forma de la confirmación**: la que dice «ha salido bien» sin sacar un aviso aparte. Es
      // verde y sale del tema, como todo. Se usa mientras dura la confirmación.
      exito: 'bg-exito-suave text-exito border border-exito',
    };

    return [
      base,
      formas[this.forma()],
      this.ancho() ? 'w-full' : '',
      // Un botón de sólo icono no lleva el relleno de lado del que lleva texto: se queda cuadrado.
      // **Un botón de sólo icono es cuadrado y discreto** (decisión del responsable, 2026-09-29): medía
      // 44 px y los lápices de la ficha pesaban demasiado. 36 px sigue por encima de lo que se puede
      // pulsar —la prueba de interfaz exige 32— y deja el icono con aire.
      this.soloIcono() ? 'min-w-9 px-1.5 py-1.5' : '',
    ].join(' ');
  }
}
