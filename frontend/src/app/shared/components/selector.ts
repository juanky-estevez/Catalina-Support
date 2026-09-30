import { Component, computed, input, model } from '@angular/core';

/** Una opción del selector: lo que se guarda, lo que se lee y a qué grupo pertenece. */
export interface OpcionSelector {
  readonly valor: string;
  readonly etiqueta: string;
  /** El grupo con el que se agrupa en la lista: «Claros», «Oscuros», «Roles». */
  readonly grupo: string;
  /**
   * Si la opción se enseña pero no se puede elegir.
   *
   * Es para lo que está previsto y todavía no funciona —hoy, el alta de cuentas de directorio—: así se
   * ve que existe y **no se ofrece algo que va a fallar**
   * (`docs/interfaz-y-experiencia.md`, sección 3.6).
   */
  readonly deshabilitado?: boolean;
}

/** Un grupo ya armado, para pintarlo. */
interface GrupoSelector {
  readonly nombre: string;
  readonly opciones: readonly OpcionSelector[];
}

/**
 * Selector: el del inventario de componentes propios
 * (`docs/interfaz-y-experiencia.md`, sección 6.3).
 *
 * Es un `<select>` de verdad, con su etiqueta: funciona con teclado, con lector de pantalla y en
 * móvil sin construir nada, y allí sale el selector nativo del sistema, que es el que la gente
 * conoce. Se usa cuando las opciones son demasiadas para un conmutador segmentado —el tema tiene
 * ocho— y para los filtros.
 *
 * **Las opciones se agrupan** (`<optgroup>`): ocho temas en una lista plana se leen peor que
 * separados en claros y oscuros.
 *
 * Los colores salen del tema, y el borde es **el fuerte**: es lo que hace visible que ahí hay un
 * control.
 */
@Component({
  selector: 'app-selector',
  template: `
    <div class="flex flex-col gap-1">
      @if (etiqueta()) {
        <!--
          La etiqueta se queda aunque no se vea: es el **nombre accesible** del control, y sin ella
          un lector de pantalla anunciaría un desplegable sin decir de qué. Cuando no se enseña, se
          oculta a la vista y sigue estando para quien la necesita.
        -->
        <label
          [for]="identificador()"
          [class]="etiquetaVisible() ? 'text-sm font-medium text-texto' : 'sr-only'"
        >
          {{ etiqueta() }}
        </label>
      }

      <select
        [id]="identificador()"
        [disabled]="deshabilitado()"
        [attr.aria-describedby]="ayuda() ? identificador() + '-ayuda' : null"
        (change)="elegir($event)"
        [class]="
          'rounded-md border border-borde-fuerte bg-superficie px-2 py-1.5 text-sm text-texto min-h-9' +
          (anchoCompleto() ? ' w-full' : '')
        "
      >
        @for (grupo of grupos(); track grupo.nombre) {
          <optgroup [label]="grupo.nombre">
            @for (opcion of grupo.opciones; track opcion.valor) {
              <!-- Se marca la elegida en la propia opción: es lo que funciona cuando las opciones se
                   pintan después de asignar el valor. -->
              <option
                [value]="opcion.valor"
                [selected]="opcion.valor === valor()"
                [disabled]="opcion.deshabilitado ?? false"
              >
                {{ opcion.etiqueta }}
              </option>
            }
          </optgroup>
        }
      </select>

      @if (ayuda()) {
        <p [id]="identificador() + '-ayuda'" class="text-sm text-apagado">{{ ayuda() }}</p>
      }
    </div>
  `,
})
export class Selector {
  readonly identificador = input.required<string>();
  readonly etiqueta = input<string>('');
  /** Si la etiqueta se enseña. En falso se queda para quien usa un lector de pantalla, y no se ve. */
  readonly etiquetaVisible = input(true);
  readonly opciones = input.required<readonly OpcionSelector[]>();
  readonly deshabilitado = input(false);
  /** Un texto de ayuda debajo, atado al control: es lo que explica una opción que no se puede elegir. */
  readonly ayuda = input<string>('');
  /**
   * Si el desplegable ocupa todo el ancho de donde está.
   *
   * Lo usan los controles del menú lateral —**el idioma y el tema**, que van uno debajo del otro y
   * tienen que medir lo mismo** (decisión del responsable, 2026-09-26)—. En el resto de sitios mide lo
   * que mide su contenido, que es lo que quieren un filtro o el selector de responsables.
   */
  readonly anchoCompleto = input(false);

  /** La opción elegida. Se puede leer y escribir: `[(valor)]="tema"`. */
  readonly valor = model.required<string>();

  /** Las opciones agrupadas, **en el orden en que llegan**: quien las pasa decide qué va primero. */
  protected readonly grupos = computed<readonly GrupoSelector[]>(() => {
    const grupos = new Map<string, OpcionSelector[]>();

    for (const opcion of this.opciones()) {
      const existente = grupos.get(opcion.grupo);
      if (existente) {
        existente.push(opcion);
      } else {
        grupos.set(opcion.grupo, [opcion]);
      }
    }

    return [...grupos.entries()].map(([nombre, opciones]) => ({ nombre, opciones }));
  });

  protected elegir(evento: Event): void {
    this.valor.set((evento.target as HTMLSelectElement).value);
  }
}
