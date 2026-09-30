import { Component, ElementRef, input, model, viewChild } from '@angular/core';

/** Algo que se inserta en el texto: un marcador, por ejemplo. */
export interface Insercion {
  readonly etiqueta: string;
  readonly valor: string;
  /** Lo que dice al pasar por encima: para qué sirve ese marcador. */
  readonly ayuda?: string;
}

/** Los rótulos del editor. Llegan de fuera: un componente de `shared` no sabe en qué idioma se lee. */
export interface TextosDelEditor {
  /** El nombre accesible del grupo de botones: «Formato del texto». */
  readonly barra: string;
  readonly negrita: string;
  readonly cursiva: string;
  readonly lista: string;
  readonly enlace: string;
  readonly vistaPrevia: string;
}

/**
 * Editor de texto con formato: el del inventario de componentes propios
 * (`docs/interfaz-y-experiencia.md`, sección 6.3).
 *
 * **Es un campo de texto con botones que envuelven lo seleccionado**, y no un editor de los que
 * pintan el resultado mientras se escribe (decisión 1 de `docs/modules/mail.md`). Hay tres razones, y
 * las tres son del producto: lo que se guarda es **HTML** y conviene poder mirarlo; un contenido
 * editable trae consigo un mundo de etiquetas pegadas de cualquier sitio; y **acepta HTML escrito a
 * mano**, porque quien administra la instalación sabe lo que hace.
 *
 * **La vista previa no la pinta este componente con reglas propias**: el HTML llega ya renderizado
 * desde el backend, con los mismos datos de ejemplo que el correo de prueba. Es lo que hace que lo
 * que se ve aquí sea lo que va a salir.
 *
 * Los colores salen del tema, como en todo.
 */
@Component({
  selector: 'app-editor',
  template: `
    <div class="flex flex-col gap-2">
      <div class="flex flex-wrap items-center gap-2">
        <label [for]="identificador()" class="text-sm font-medium text-texto">{{ etiqueta() }}</label>

        <!--
          Los botones van en grupo y con su nombre: un lector de pantalla tiene que poder decir que
          son el formato del texto, y qué hace cada uno.
        -->
        <div
          class="ml-auto flex flex-wrap gap-1"
          role="group"
          [attr.aria-label]="textos().barra"
        >
          @for (boton of botones(); track boton.texto) {
            <button
              type="button"
              class="inline-flex min-h-9 items-center rounded-md border border-borde-fuerte px-2 text-sm text-texto hover:bg-superficie-suave"
              [attr.aria-label]="boton.texto"
              [attr.title]="boton.texto"
              (click)="boton.aplicar()"
            >
              <span aria-hidden="true" [class]="boton.clase">{{ boton.muestra }}</span>
            </button>
          }
        </div>
      </div>

      <textarea
        #area
        [id]="identificador()"
        [rows]="filas()"
        [value]="valor()"
        (input)="alEscribir($event)"
        class="rounded-md border border-borde-fuerte bg-superficie px-3 py-2 font-mono text-sm text-texto"
      ></textarea>

      @if (ayuda()) {
        <p class="text-sm text-apagado">{{ ayuda() }}</p>
      }

      <!--
        Lo que se inserta en el punto del cursor: los marcadores del correo, que es lo que hace que
        nadie tenga que memorizar cómo se escriben. Se insertan desde aquí, y no desde fuera, porque
        **quien tiene el campo es quien sabe dónde está el cursor**.
      -->
      @if (inserciones().length) {
        <div class="flex flex-col gap-1">
          <span class="text-xs font-medium text-apagado">{{ etiquetaInserciones() }}</span>

          <div class="flex flex-wrap gap-1">
            @for (insercion of inserciones(); track insercion.valor) {
              <button
                type="button"
                class="inline-flex min-h-8 items-center rounded-md border border-borde-fuerte px-2 font-mono text-xs text-texto hover:bg-superficie-suave"
                [attr.title]="insercion.ayuda || insercion.etiqueta"
                (click)="insertar(insercion.valor)"
              >
                {{ insercion.etiqueta }}
              </button>
            }
          </div>
        </div>
      }

      <!-- La vista previa: el HTML que ha renderizado el backend, con sus datos de ejemplo. -->
      <div class="rounded-md border border-borde bg-superficie p-3">
        <p class="mb-2 text-xs font-medium uppercase text-apagado">{{ textos().vistaPrevia }}</p>

        @if (vistaPrevia()) {
          <div class="text-sm text-texto" [innerHTML]="vistaPrevia()"></div>
        } @else {
          <p class="text-sm text-apagado">{{ sinVistaPrevia() }}</p>
        }
      </div>
    </div>
  `,
})
export class Editor {
  readonly identificador = input.required<string>();
  readonly etiqueta = input.required<string>();
  readonly ayuda = input('');
  readonly filas = input(10);
  readonly textos = input.required<TextosDelEditor>();
  /** Lo que dice cuando todavía no hay vista previa que enseñar. */
  readonly sinVistaPrevia = input('');
  /** El HTML ya renderizado. Si viene vacío, se enseña el aviso de arriba. */
  readonly vistaPrevia = input('');
  /** Lo que se puede insertar en el punto del cursor. */
  readonly inserciones = input<readonly Insercion[]>([]);
  readonly etiquetaInserciones = input('');

  /** El HTML que se está editando, en dos direcciones. */
  readonly valor = model<string>('');

  private readonly area = viewChild<ElementRef<HTMLTextAreaElement>>('area');

  /** Los cuatro botones, con lo que hacen. Se arma en el componente para no repetir la plantilla. */
  protected botones(): readonly {
    texto: string;
    muestra: string;
    clase: string;
    aplicar: () => void;
  }[] {
    return [
      { texto: this.textos().negrita, muestra: 'N', clase: 'font-bold', aplicar: () => this.envolver('<strong>', '</strong>') },
      { texto: this.textos().cursiva, muestra: 'C', clase: 'italic', aplicar: () => this.envolver('<em>', '</em>') },
      { texto: this.textos().lista, muestra: '≡', clase: '', aplicar: () => this.lista() },
      { texto: this.textos().enlace, muestra: '↗', clase: '', aplicar: () => this.enlace() },
    ];
  }

  protected alEscribir(evento: Event): void {
    this.valor.set((evento.target as HTMLTextAreaElement).value);
  }

  /**
   * Escribe algo donde está el cursor, o al final si el campo no tiene el foco.
   *
   * Es lo que usan los marcadores: pulsar `{{numero}}` y que aparezca donde se está escribiendo, en
   * vez de copiarlo y pegarlo.
   */
  insertar(texto: string): void {
    const area = this.area()?.nativeElement;
    if (!area) {
      return;
    }

    const inicio = area.selectionStart ?? area.value.length;
    const fin = area.selectionEnd ?? area.value.length;

    area.value = area.value.slice(0, inicio) + texto + area.value.slice(fin);
    area.focus();
    area.setSelectionRange(inicio + texto.length, inicio + texto.length);

    this.valor.set(area.value);
  }

  /** Envuelve lo seleccionado entre dos etiquetas, y deja seleccionado el texto de dentro. */
  private envolver(antes: string, despues: string): void {
    const area = this.area()?.nativeElement;
    if (!area) {
      return;
    }

    const inicio = area.selectionStart;
    const fin = area.selectionEnd;
    const texto = area.value;
    const seleccionado = texto.slice(inicio, fin);

    area.value = texto.slice(0, inicio) + antes + seleccionado + despues + texto.slice(fin);
    area.focus();
    area.setSelectionRange(inicio + antes.length, inicio + antes.length + seleccionado.length);

    this.valor.set(area.value);
  }

  /**
   * Convierte lo seleccionado en una lista: **una línea, un elemento**.
   *
   * Es lo que espera quien selecciona tres líneas y pulsa el botón de la lista; si no hay nada
   * seleccionado, deja un elemento vacío para escribir dentro.
   */
  private lista(): void {
    const area = this.area()?.nativeElement;
    if (!area) {
      return;
    }

    const inicio = area.selectionStart;
    const fin = area.selectionEnd;
    const texto = area.value;
    const seleccionado = texto.slice(inicio, fin);

    const lineas = seleccionado
      .split('\n')
      .map((linea) => linea.trim())
      .filter((linea) => linea !== '');

    const elementos = (lineas.length ? lineas : ['']).map((linea) => `<li>${linea}</li>`).join('');
    const nuevo = `<ul>${elementos}</ul>`;

    area.value = texto.slice(0, inicio) + nuevo + texto.slice(fin);
    area.focus();
    area.setSelectionRange(inicio + nuevo.length, inicio + nuevo.length);

    this.valor.set(area.value);
  }

  /**
   * Pone un enlace sobre lo seleccionado y deja el cursor **dentro de las comillas de la dirección**:
   * es lo siguiente que hay que escribir, y así no hay que buscarlo.
   */
  private enlace(): void {
    const area = this.area()?.nativeElement;
    if (!area) {
      return;
    }

    const inicio = area.selectionStart;
    const fin = area.selectionEnd;
    const texto = area.value;
    const seleccionado = texto.slice(inicio, fin);

    const antes = '<a href="';
    const nuevo = `${antes}">${seleccionado}</a>`;

    area.value = texto.slice(0, inicio) + nuevo + texto.slice(fin);
    area.focus();
    // El cursor va entre las comillas, que es donde falta la dirección.
    area.setSelectionRange(inicio + antes.length, inicio + antes.length);

    this.valor.set(area.value);
  }
}
