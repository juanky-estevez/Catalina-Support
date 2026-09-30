import {
  Component,
  ElementRef,
  computed,
  effect,
  input,
  model,
  signal,
  viewChild,
} from '@angular/core';

/** El tope por archivo, que es el mismo del backend y el que deja margen bajo el de nginx. */
export const MAX_ARCHIVO_BYTES = 25 << 20;

/**
 * Las extensiones que se admiten, **la misma lista cerrada que el backend**
 * (`docs/modules/tickets.md`, sección 2.3).
 *
 * Está aquí repetida a propósito, y con la autoridad en el backend: la copia sirve para decir «ese
 * archivo no vale» **en el momento de elegirlo**, sin subir 3 MB para que te lo digan después. Si
 * mañana cambia la lista, cambia en los dos sitios, y el error del backend sigue estando ahí como red
 * de seguridad.
 *
 * Los **cuatro formatos de vídeo** entraron el 2026-09-26 (decisión 47): `mp4` y `webm`, que cualquier
 * navegador reproduce, y también `mov` y `avi`, que muchos no saben reproducir —y entonces queda el
 * botón de descargar del visor, que es la consecuencia dicha y aceptada—.
 */
const ADMITIDAS = [
  'pdf',
  'png',
  'jpg',
  'jpeg',
  'gif',
  'webp',
  'mp4',
  'webm',
  'mov',
  'avi',
  'doc',
  'docx',
  'xls',
  'xlsx',
  'ppt',
  'pptx',
  'odt',
  'ods',
  'odp',
  'txt',
  'csv',
  'log',
  'md',
  'css',
  // **Texto y código** (decisión 54, del 2026-09-26): se descargan siempre, nunca se pintan.
  'sql',
  'json',
  'xml',
  'yml',
  'yaml',
  'ini',
  'conf',
  'cnf',
  'sh',
  'bash',
  'bat',
  'ps1',
  'py',
  'js',
  'ts',
  'java',
  'php',
  'go',
  'cs',
  'rb',
  'pl',
  'kt',
  'rs',
  'swift',
  'c',
  'h',
  'cpp',
  'hpp',
  'vue',
  'jsx',
  'tsx',
  'htaccess',
  'env',
  'properties',
  'diff',
  'patch',
  'bak',
  'old',
  'zip',
  'rar',
  'tar',
  'gz',
  'tgz',
  '7z',
];

/** Las que se ven en línea dentro del texto. */
const IMAGENES = ['png', 'jpg', 'jpeg', 'gif', 'webp'];

/** Y las que se ven con un reproductor. */
const VIDEOS = ['mp4', 'webm', 'mov', 'avi'];

/** El tipo de archivo, en minúsculas y sin el punto. */
export function extensionDe(nombre: string): string {
  const partes = nombre.toLowerCase().split('.');

  return partes.length > 1 ? partes[partes.length - 1] : '';
}

/** Si esa extensión está en la lista cerrada. */
export function extensionAdmitida(nombre: string): boolean {
  const partes = nombre.toLowerCase().split('.');
  if (partes.length < 2) {
    return false;
  }

  return ADMITIDAS.includes(partes[partes.length - 1]);
}

/** Si ese archivo se ve en línea: una imagen. */
export function esImagen(nombre: string): boolean {
  return IMAGENES.includes(extensionDe(nombre));
}

/** Si ese archivo se ve en línea: un vídeo, con su reproductor. */
export function esVideo(nombre: string): boolean {
  return VIDEOS.includes(extensionDe(nombre));
}

/**
 * Si la dirección de un enlace se puede guardar: **tiene que empezar por `http`, `https` o `mailto`**
 * (docs/modules/tickets.md, sección 2.3).
 *
 * Es la misma regla que aplica el backend al guardar, y aquí sirve para decirlo antes: un enlace que
 * empiece por `javascript:` no se pone y se explica por qué, en vez de rechazar el comentario entero
 * por algo que la persona no ha escrito.
 */
export function direccionAdmitida(destino: string): boolean {
  const valor = destino.trim().toLowerCase();

  return valor.startsWith('http://') || valor.startsWith('https://') || valor.startsWith('mailto:');
}

/* --------------------------------------------------------------------------------------------
 * Lo que se saca del lienzo para guardar
 * ------------------------------------------------------------------------------------------ */

/**
 * Los elementos que **no entran nunca** en el texto, ni ellos ni lo que lleven dentro: código que se
 * ejecuta, estilos y formularios. El backend los rechaza igual —su lista blanca es la única puerta—,
 * pero aquí se quitan antes de mandar nada, porque un texto que no se puede guardar es peor que uno
 * que se manda sin el trozo que no valía.
 */
const FUERA = [
  'script',
  'style',
  'iframe',
  'object',
  'embed',
  'form',
  'input',
  'textarea',
  'select',
  'button',
  'link',
  'meta',
];

/**
 * Cómo se llama cada etiqueta al guardar. Lo que no esté aquí **se desenvuelve**: se queda su texto y
 * se va la etiqueta, que es lo que hace que un `<font>` o un `<span>` pegados de cualquier sitio no
 * tiren el comentario entero.
 *
 * `div` se guarda como `p` porque es lo que el navegador pone al separar renglones dentro de un
 * contenido editable, y `b`/`i`/`strike`/`del` se guardan con su nombre moderno: el formato copiado de
 * otro sitio llega con los nombres viejos y el backend admite los dos, pero guardar uno solo hace que
 * el texto guardado se lea igual siempre.
 */
const COMO_SE_LLAMA: Record<string, string> = {
  div: 'p',
  b: 'strong',
  i: 'em',
  strike: 's',
  del: 's',
};

/** Las etiquetas que se guardan tal cual, sin atributos, cuando están en la lista blanca. */
const TAL_CUAL = ['p', 'br', 'strong', 'em', 'u', 's', 'ul', 'ol', 'li'];

/** La marca que se le pone en pantalla a un adjunto que no se puede enseñar. No se guarda nunca. */
export const CLASE_DE_LO_QUE_FALTA = 'adjunto-que-falta';

/** Y la clase de la vista previa dentro del lienzo. Tampoco se guarda: es sólo cómo se ve al escribir. */
/**
 * Cómo se ve un adjunto **mientras se escribe**: con el mismo tope que al leerlo, porque lo que se ve
 * escribiendo tiene que ser lo que va a salir (decisión 55, del 2026-09-26: 480 × 360 como máximo, y el
 * clic abre el visor grande).
 */
export const CLASE_DE_VISTA_PREVIA =
  'max-h-[360px] max-w-[min(100%,480px)] rounded-md border border-borde';

/** El nombre de un archivo, listo para ir dentro de un atributo. */
function escaparAtributo(valor: string): string {
  return valor
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

/** Y el texto, listo para ir entre dos etiquetas. */
function escaparTexto(valor: string): string {
  return valor.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

/** Texto plano convertido en HTML: se respetan los saltos de línea, y nada más. */
export function textoComoHtml(texto: string): string {
  return texto
    .split(/\r\n|\r|\n/)
    .map((linea) => escaparTexto(linea))
    .join('<br>');
}

/**
 * El HTML del lienzo, **listo para guardar**: sin direcciones y sólo con los atributos de la lista
 * blanca del backend (docs/modules/tickets.md, sección 2.3).
 *
 * Es la trampa más fácil de pisar de todo esto. Mientras se escribe, la imagen y el vídeo llevan una
 * dirección `blob:` del navegador y la vista previa lleva sus clases; si eso llegara a la base, la
 * dirección no significaría nada mañana —y la clase no está en la lista blanca—, así que el backend
 * rechazaría el comentario entero con `tickets.body.notAllowed`. Aquí se quitan las dos cosas: lo que
 * sale de esta función es texto con `data-adjunto` y nada más.
 */
export function htmlGuardable(raiz: HTMLElement): string {
  return Array.from(raiz.childNodes)
    .map((nodo) => htmlDelNodo(nodo))
    .join('');
}

function htmlDelNodo(nodo: Node): string {
  if (nodo.nodeType === Node.TEXT_NODE) {
    return escaparTexto(nodo.nodeValue ?? '');
  }
  if (nodo.nodeType !== Node.ELEMENT_NODE) {
    return '';
  }

  const elemento = nodo as HTMLElement;
  const etiqueta = elemento.tagName.toLowerCase();

  if (FUERA.includes(etiqueta)) {
    return '';
  }

  // La marca de un adjunto que no se puede enseñar es de la pantalla, no del texto: fuera.
  if (elemento.classList.contains(CLASE_DE_LO_QUE_FALTA)) {
    return '';
  }

  const hijos = Array.from(elemento.childNodes)
    .map((hijo) => htmlDelNodo(hijo))
    .join('');

  if (etiqueta === 'img') {
    const adjunto = elemento.getAttribute('data-adjunto');
    return adjunto ? `<img data-adjunto="${escaparAtributo(adjunto)}">` : '';
  }

  if (etiqueta === 'video') {
    const adjunto = elemento.getAttribute('data-adjunto');
    // **Sin `controls`**: el reproductor lo pone la pantalla al pintarlo, y el atributo no está en la
    // lista blanca del backend (`backend/modules/tickets/services/body.go`).
    return adjunto ? `<video data-adjunto="${escaparAtributo(adjunto)}">${hijos}</video>` : '';
  }

  if (etiqueta === 'a') {
    const adjunto = elemento.getAttribute('data-adjunto');
    if (adjunto) {
      return `<a data-adjunto="${escaparAtributo(adjunto)}">${hijos}</a>`;
    }

    const destino = (elemento.getAttribute('href') ?? '').trim();
    // Una dirección que no vale se lleva el enlace, no el texto: lo escrito se queda.
    return direccionAdmitida(destino)
      ? `<a href="${escaparAtributo(destino)}">${hijos}</a>`
      : hijos;
  }

  // **Una mención se guarda tal cual**: el identificador es el dato y el nombre que se lee es lo que
  // hay dentro. Un `data-mencion` que no sea un identificador se desenvuelve —el texto se queda y la
  // etiqueta se va—, que es lo mismo que se hace con cualquier etiqueta que no esté en la lista.
  if (etiqueta === 'span') {
    const mencion = (elemento.getAttribute(ATRIBUTO_DE_LA_MENCION) ?? '').trim();

    return /^[1-9][0-9]{0,17}$/.test(mencion)
      ? `<span ${ATRIBUTO_DE_LA_MENCION}="${escaparAtributo(mencion)}">${hijos}</span>`
      : hijos;
  }

  const nombre = COMO_SE_LLAMA[etiqueta] ?? etiqueta;
  if (TAL_CUAL.includes(nombre)) {
    return nombre === 'br' ? '<br>' : `<${nombre}>${hijos}</${nombre}>`;
  }

  // Lo que no está en la lista se desenvuelve: el texto se queda y la etiqueta se va.
  return hijos;
}

/* --------------------------------------------------------------------------------------------
 * El componente
 * ------------------------------------------------------------------------------------------ */

/** Los rótulos del editor. Llegan de fuera: un componente de `shared` no sabe en qué idioma se lee. */
/**
 * Una persona a la que se puede etiquetar: su identificador y cómo se llama.
 *
 * **El identificador es lo que se guarda** y el nombre es lo que se lee (`docs/modules/tickets.md`,
 * decisión 58): los nombres se repiten y cambian, y una mención guardada por nombre se rompe sola.
 */
export interface PersonaDelEditor {
  readonly id: number;
  readonly nombre: string;
  /**
   * El correo, **para poder buscar por él** y para distinguir a dos personas que se llaman igual.
   *
   * Es opcional porque quien no lo tenga sigue funcionando; se enseña al lado del nombre en la lista de
   * a quién etiquetar, que es donde hace falta (`docs/interfaz-y-experiencia.md`, decisión 74).
   */
  readonly email?: string;
}

/**
 * Lo que se está escribiendo detrás de una arroba, si se está escribiendo una mención.
 *
 * Se le pasa **el texto que hay antes del cursor** y devuelve lo escrito detrás de la última `@`, o
 * `null` si no hay ninguna mención a medias. Va aparte —y es pura— para poder probarla sin navegador:
 * es la regla que decide cuándo sale la lista de a quién etiquetar mientras se teclea (decisión 74).
 */
export function mencionEscrita(textoAntesDelCursor: string): string | null {
  // **La mención es la última palabra, y empieza por arroba**: se busca la arroba del principio de esa
  // palabra y lo que venga detrás es la búsqueda. Así vale **escribir el correo entero**
  // (`@soporte2@demo.com`), que es lo que pidió el responsable (decisión 74) —por eso no se corta en la
  // segunda arroba—, y en cuanto hay un espacio, la mención ya no está a medias.
  const ultima = textoAntesDelCursor.match(/(?:^|\s)@(\S*)$/);

  return ultima ? ultima[1] : null;
}

/**
 * Dónde se pone la lista de a quién etiquetar, **junto al cursor** (decisión 89).
 *
 * Es una función pura para poder probarla sin navegador: recibe **dónde está el cursor** en la pantalla,
 * **cuánto puede medir la lista** y **el alto de la ventana**, y devuelve dónde ponerla.
 *
 * Se pone **debajo de la línea que se escribe**; si ahí no cabe, **encima**; y se **aprieta contra el
 * borde** para que nunca se salga por la derecha. Antes salía encima del bloque del comentario, donde el
 * ojo no la encuentra: lo reportó el responsable.
 */
export function colocacionDeLasSugerencias(
  cursor: { izquierda: number; arriba: number; abajo: number },
  altoDeLaLista: number,
  ventana: { ancho: number; alto: number },
  anchoDeLaLista = 320,
): { izquierda: number; arriba: number; maxAlto: number } {
  const margen = 8;
  const cabeDebajo = cursor.abajo + altoDeLaLista + margen <= ventana.alto;
  const arriba = cabeDebajo ? cursor.abajo + 4 : Math.max(margen, cursor.arriba - altoDeLaLista - 4);

  // Y nunca se sale por la derecha: si no cabe, se aprieta contra el borde.
  const izquierda = Math.max(margen, Math.min(cursor.izquierda, ventana.ancho - anchoDeLaLista - margen));

  return { izquierda, arriba, maxAlto: altoDeLaLista };
}

/** Cuánto puede medir de alto la lista de a quién etiquetar antes de empezar a desplazarse. */
const ALTO_DE_LA_LISTA = 200;

/** El atributo de una mención: a quién nombra, por su identificador (decisión 58). */
export const ATRIBUTO_DE_LA_MENCION = 'data-mencion';

export interface TextosDelEditorConAdjuntos {
  /** El nombre accesible del grupo de botones: «Formato del texto». */
  readonly barra: string;
  readonly negrita: string;
  readonly cursiva: string;
  readonly subrayado: string;
  readonly tachado: string;
  readonly listaVinietas: string;
  readonly listaNumerada: string;
  readonly enlace: string;
  readonly adjuntar: string;
  /** Etiquetar a otro técnico o desarrollador: el nombre entra en el texto y la persona pasa a observar. */
  readonly etiquetar: string;
  readonly etiquetarA: string;
  readonly sinPersonas: string;
  /** Lo que se pide para poner un enlace, y lo que se dice si la dirección no vale. */
  readonly direccion: string;
  readonly direccionPoner: string;
  readonly direccionCancelar: string;
  readonly direccionInvalida: string;
  /** Lo que se dice cuando un archivo no vale, con la lista de los que no valen detrás. */
  readonly avisoDeRechazo: string;
  /** Y lo que se le pone encima a un adjunto que no se puede enseñar. */
  readonly adjuntoNoSubio: string;
}

/**
 * Editor con adjuntos dentro: el del inventario de componentes propios
 * (`docs/interfaz-y-experiencia.md`, sección 6.5), y el cuadro de escribir de un ticket y de un
 * comentario.
 *
 * **Un cuadro de texto normal no puede llevar una imagen dentro**, y lo que se pide es que los
 * archivos vivan en el sitio donde se ponen y que **se vean de verdad mientras se escribe**
 * (docs/modules/tickets.md, secciones 2.3 y 4, decisiones 45 a 47). Por eso el lienzo es un
 * `contenteditable` con una barra de botones que actúan sobre lo seleccionado.
 *
 * **Se usa `document.execCommand`**, que está obsoleto y sigue siendo lo que hay: la alternativa es
 * una biblioteca de editor completo, y este producto **no añade bibliotecas de componentes**
 * (`AGENTS.md`). Lo que hace falta de un editor de verdad —formato sobre la selección, insertar donde
 * está el cursor, pegar sin formato— está en esa API desde siempre, funciona en todos los navegadores
 * que se soportan y no trae cien kilobytes de dependencias. El día que desaparezca, lo que hay que
 * sustituir es este archivo y nada más.
 *
 * **Las tres puertas para adjuntar meten el archivo donde está el cursor**: pegar (`Ctrl+V`),
 * arrastrarlo encima y el botón. Las tres comprueban lo mismo —la lista cerrada de extensiones y el
 * tope de peso— y avisan igual de lo que no vale.
 *
 * **Lo que se guarda no lleva direcciones**: `html()` devuelve el texto con la referencia por nombre
 * (`data-adjunto`) y sólo con los atributos que el backend admite. La dirección `blob:` que se ve
 * mientras se escribe es del navegador y se cae al guardar.
 *
 * **El componente no sube nada.** Devuelve el texto y los archivos nuevos —los que están dentro del
 * texto y no están ya en el ticket—, y quien lo usa los sube: vive en `shared`, que no puede depender
 * de ningún módulo.
 */
@Component({
  selector: 'app-editor-con-adjuntos',
  template: `
    <div class="flex flex-col gap-2">
      <div class="flex flex-wrap items-center gap-2">
        <!--
          La etiqueta es la de verdad y apunta al lienzo con su atributo for: es lo que hace que el
          cuadro de escribir se pueda localizar y anunciar como un campo, y no como un adorno
          (sección 8).
        -->
        <label
          [id]="identificador() + '-etiqueta'"
          [for]="identificador()"
          class="text-sm font-medium text-texto"
        >
          {{ etiqueta() }}
        </label>

        <!--
          Los botones van en grupo y con su nombre: un lector de pantalla tiene que poder decir que son
          el formato del texto y qué hace cada uno.
        -->
        <div class="ml-auto flex flex-wrap gap-1" role="group" [attr.aria-label]="textos().barra">
          @for (boton of botones(); track boton.texto) {
            <button
              type="button"
              class="inline-flex min-h-9 items-center rounded-md border border-borde-fuerte px-2 text-sm text-texto hover:bg-superficie-suave"
              [attr.aria-label]="boton.texto"
              [attr.title]="boton.texto"
              (mousedown)="noQuitarElFoco($event)"
              (click)="boton.aplicar()"
            >
              <span aria-hidden="true" [class]="boton.clase">{{ boton.muestra }}</span>
            </button>
          }

          <!--
            **Etiquetar**: mete el nombre de un técnico o un desarrollador dentro del comentario y lo
            convierte en observador del ticket (docs/modules/tickets.md, decisión 58). Sólo se ofrece a
            quien puede hacerlo —Soporte y Desarrollo— y sólo si hay alguien a quien llamar.
          -->
          @if (puedeEtiquetar() && personas().length) {
            <button
              type="button"
              class="inline-flex min-h-9 items-center gap-1 rounded-md border border-borde-fuerte px-2 text-sm text-texto hover:bg-superficie-suave"
              [attr.aria-expanded]="eligiendoPersona()"
              [attr.title]="textos().etiquetar"
              (mousedown)="noQuitarElFoco($event)"
              (click)="alternarPersonas()"
            >
              <span aria-hidden="true">@</span>
              {{ textos().etiquetar }}
            </button>
          }

          <!--
            El botón de adjuntar es la tercera puerta, con el campo de archivo dentro y escondido: en
            móvil esto abre la cámara o los archivos, que es lo que se espera. El campo es el que se
            pulsa, así que funciona con teclado.
          -->
          <label
            class="inline-flex min-h-9 cursor-pointer items-center gap-1 rounded-md border border-borde-fuerte bg-superficie px-2 text-sm text-texto hover:bg-superficie-suave"
          >
            <span aria-hidden="true">📎</span>
            {{ textos().adjuntar }}
            <input #campo type="file" multiple class="sr-only" (change)="elegirArchivos($event)" />
          </label>
        </div>
      </div>

      <!-- El enlace se pide con su dirección, y se comprueba antes de ponerlo. -->
      @if (pidiendoEnlace()) {
        <div
          class="flex flex-wrap items-end gap-2 rounded-md border border-borde bg-superficie-suave p-2"
        >
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-texto" [for]="identificador() + '-direccion'">
              {{ textos().direccion }}
            </label>
            <input
              [id]="identificador() + '-direccion'"
              type="text"
              [value]="direccion()"
              (input)="alEscribirLaDireccion($event)"
              class="min-h-9 rounded-md border border-borde-fuerte bg-superficie px-2 text-sm text-texto"
            />
          </div>

          <button
            type="button"
            class="inline-flex min-h-9 items-center rounded-md border border-borde-fuerte bg-superficie px-2 text-sm text-texto hover:bg-superficie-suave"
            (click)="ponerElEnlace()"
          >
            {{ textos().direccionPoner }}
          </button>

          <button
            type="button"
            class="inline-flex min-h-9 items-center rounded-md border border-borde-fuerte bg-superficie px-2 text-sm text-apagado hover:text-texto"
            (click)="cerrarElEnlace()"
          >
            {{ textos().direccionCancelar }}
          </button>

          @if (enlaceInvalido()) {
            <p class="text-sm text-peligro" role="alert">{{ textos().direccionInvalida }}</p>
          }
        </div>
      }

      <!--
        El buscador de personas: los técnicos y los desarrolladores activos, en botones. Se elige uno y su
        nombre entra donde está el cursor; **la lista entera y no un desplegable**, porque son pocas
        personas y verlas todas de un vistazo es lo que hace rápido llamar a alguien.
      -->
      @if (eligiendoPersona()) {
        <div
          class="flex flex-col gap-2 rounded-md border border-borde bg-superficie-suave p-2"
          [attr.aria-label]="textos().etiquetarA"
          role="group"
        >
          <span class="text-xs font-medium text-texto">{{ textos().etiquetarA }}</span>

          @if (personas().length) {
            <div class="flex flex-wrap gap-1">
              @for (persona of personas(); track persona.id) {
                <button
                  type="button"
                  class="inline-flex min-h-9 items-center rounded-md border border-borde-fuerte bg-superficie px-2 text-sm text-texto hover:bg-superficie-suave"
                  (mousedown)="noQuitarElFoco($event)"
                  (click)="etiquetarA(persona)"
                >
                  {{ persona.nombre }}
                </button>
              }
            </div>
          } @else {
            <p class="text-sm text-apagado">{{ textos().sinPersonas }}</p>
          }

          <button
            type="button"
            class="inline-flex min-h-9 w-fit items-center rounded-md border border-borde-fuerte px-2 text-sm text-apagado hover:text-texto"
            (click)="alternarPersonas()"
          >
            {{ textos().direccionCancelar }}
          </button>
        </div>
      }

      <!--
        **Las sugerencias que salen al escribir una arroba** (decisión del responsable, 2026-09-28): se filtran
        por lo que se lleva escrito —vale el nombre o el correo— y se enseña **el correo al lado**, que
        es lo que permite distinguir a dos personas que se llaman igual. Se elige con el ratón o con el
        teclado (el botón es un botón de verdad), y **Escape** cierra la lista sin tocar lo escrito.
      -->
      @if (
        puedeEtiquetar() && mencionAMedias() !== null && sugerenciasDeMencion().length
      ) {
        <div
          class="fixed z-50 flex max-w-[20rem] flex-wrap gap-1 overflow-y-auto rounded-md border border-borde-fuerte bg-superficie p-2 shadow-lg"
          role="listbox"
          [style.left.px]="posicionDeLasSugerencias()?.izquierda"
          [style.top.px]="posicionDeLasSugerencias()?.arriba"
          [style.max-height.px]="posicionDeLasSugerencias()?.maxAlto"
        >
          @for (persona of sugerenciasDeMencion(); track persona.id) {
            <button
              type="button"
              role="option"
              class="inline-flex flex-col items-start rounded-md border border-borde-fuerte bg-superficie px-2 py-1 text-left text-sm text-texto hover:bg-superficie-suave"
              (mousedown)="noQuitarElFoco($event)"
              (click)="etiquetarDesdeElTexto(persona)"
            >
              <span>{{ persona.nombre }}</span>

              @if (persona.email) {
                <span class="text-xs text-apagado">{{ persona.email }}</span>
              }
            </button>
          }
        </div>
      }

      <!--
        El lienzo. Es un campo de verdad para quien lo lee con un lector de pantalla (su rol y sus
        atributos aria), y por eso las pruebas de interfaz lo encuentran por su etiqueta como
        cualquier otro campo.
      -->
      <div
        #lienzo
        [id]="identificador()"
        [attr.contenteditable]="deshabilitado() ? 'false' : 'true'"
        role="textbox"
        aria-multiline="true"
        [attr.aria-labelledby]="identificador() + '-etiqueta'"
        [attr.aria-describedby]="ayuda() ? identificador() + '-ayuda' : null"
        [attr.aria-disabled]="deshabilitado() ? 'true' : null"
        (input)="alEscribir(); alEscribirEnElLienzo()"
        (keydown.escape)="cerrarLasSugerencias()"
        (paste)="pegar($event)"
        (dragover)="arrastrando($event)"
        (dragleave)="saliendo()"
        (drop)="soltar($event)"
        (blur)="alSalir()"
        [class]="clasesDelLienzo()"
      ></div>

      @if (error()) {
        <p class="text-sm text-peligro" role="alert">{{ error() }}</p>
      } @else if (ayuda()) {
        <p [id]="identificador() + '-ayuda'" class="text-sm text-apagado">{{ ayuda() }}</p>
      }
    </div>
  `,
})
export class EditorConAdjuntos {
  readonly identificador = input.required<string>();
  readonly etiqueta = input.required<string>();
  readonly ayuda = input('');
  readonly textos = input.required<TextosDelEditorConAdjuntos>();
  readonly deshabilitado = input(false);

  /**
   * Las direcciones de los adjuntos que ya están en el ticket, **por nombre de archivo**: lo que ya
   * subió alguna vez y hay que volver a ver al editar (la pantalla las pide con la sesión puesta y las
   * cachea, `docs/modules/tickets.md`, sección 2.3).
   */
  readonly direcciones = input<ReadonlyMap<string, string>>(new Map());

  /**
   * A quién se puede etiquetar: técnicos y desarrolladores activos, la misma lista que la de
   * responsables (`GET /api/tickets/assignees`). Vacía quiere decir que no se ofrece etiquetar.
   */
  readonly personas = input<readonly PersonaDelEditor[]>([]);

  /** Si esta persona puede etiquetar: **Soporte y Desarrollo**, y nadie más (decisión 59). */
  readonly puedeEtiquetar = input(false);

  /** Si el buscador de personas está abierto. */
  protected readonly eligiendoPersona = signal(false);

  /**
   * Lo que se está escribiendo detrás de la `@`, si se está escribiendo una.
   *
   * `null` quiere decir «no hay ninguna mención a medias»: es lo que enciende y apaga la lista de
   * sugerencias mientras se teclea (decisión 74). Se guarda **sin la arroba**.
   */
  protected readonly mencionAMedias = signal<string | null>(null);

  /**
   * Dónde se enseña la lista de sugerencias, **junto al cursor** (decisión 89): se calcula con
   * `colocacionDeLasSugerencias` cada vez que se teclea, porque el cursor se mueve.
   */
  protected readonly posicionDeLasSugerencias = signal<{
    izquierda: number;
    arriba: number;
    maxAlto: number;
  } | null>(null);

  /** La gente que encaja con lo que se lleva escrito detrás de la `@`, por nombre o por correo. */
  protected readonly sugerenciasDeMencion = computed<readonly PersonaDelEditor[]>(() => {
    const escrito = (this.mencionAMedias() ?? '').trim().toLowerCase();
    const gente = this.personas();

    const encajan = escrito
      ? gente.filter(
          (persona) =>
            persona.nombre.toLowerCase().includes(escrito) ||
            (persona.email ?? '').toLowerCase().includes(escrito),
        )
      : gente;

    // **Lo que empieza por lo escrito va primero** (y por el nombre antes que por el correo): buscando
    // «Arroba», lo que se quiere es quien se llama así, no quien lo lleva en medio. Se ordena en vez de
    // sólo filtrar porque la lista tiene tope, y con muchos nombres parecidos el que se busca puede
    // quedar fuera de los primeros —le pasó a la prueba de la arroba, con cuentas de otras pasadas—.
    const ordenadas = escrito ? [...encajan].sort((a, b) => peso(a) - peso(b)) : encajan;

    // **Un tope**: la lista se enseña mientras se escribe, no es un buscador aparte.
    return ordenadas.slice(0, 8);

    function peso(persona: PersonaDelEditor): number {
      const nombre = persona.nombre.toLowerCase();
      const correo = (persona.email ?? '').toLowerCase();

      if (nombre.startsWith(escrito)) return 0;
      if (correo.startsWith(escrito)) return 1;

      return 2;
    }
  });

  /** El texto que se está escribiendo, **ya listo para guardar**, en dos direcciones. */
  readonly valor = model<string>('');

  /** Los archivos nuevos que hay dentro del texto: los que hay que subir al enviar. */
  readonly archivos = signal<readonly File[]>([]);

  protected readonly error = signal('');
  protected readonly pidiendoEnlace = signal(false);
  protected readonly direccion = signal('');
  protected readonly enlaceInvalido = signal(false);

  private readonly activo = signal(false);
  private readonly lienzo = viewChild<ElementRef<HTMLElement>>('lienzo');

  /** Los archivos que se han adjuntado aquí, por nombre. */
  private readonly nuevos = new Map<string, File>();

  /** Y su dirección provisional mientras se escribe, también por nombre. */
  private readonly provisionales = new Map<string, string>();

  /** Lo último que se ha pintado en el lienzo, para no volver a pintarlo al oír el propio cambio. */
  private ultimoPintado = '';

  /** La selección de cuando se pulsó «Enlace»: el campo de la dirección se lleva el foco. */
  private rangoGuardado: Range | null = null;

  constructor() {
    // Los renglones se separan con `<p>`, que es lo que admite la lista blanca: el navegador, si no se
    // le dice otra cosa, mete `<div>`, y aunque al guardar se convierte, es mejor que el lienzo tenga
    // desde el principio la forma que se guarda.
    if (typeof document.execCommand === 'function') {
      document.execCommand('defaultParagraphSeparator', false, 'p');
      // Y el formato se pone con etiquetas, no con `style` en línea, que no entra en la lista blanca.
      document.execCommand('styleWithCSS', false, 'false');
    }

    effect(() => {
      const valor = this.valor();
      // Se lee también aquí para que la vista previa se resuelva en cuanto llegue la dirección.
      this.direcciones();

      const elemento = this.lienzo()?.nativeElement;
      if (!elemento) {
        return;
      }

      if (valor !== this.ultimoPintado) {
        this.ultimoPintado = valor;
        elemento.innerHTML = valor ?? '';
        this.apuntarLosArchivosDeDentro(elemento);
      }

      this.resolver(elemento);
    });
  }

  // ------------------------------------------------------------------ lo que devuelve

  /** El texto del lienzo, listo para guardar: sin direcciones y sólo con lo que el backend admite. */
  html(): string {
    const elemento = this.lienzo()?.nativeElement;

    return elemento ? htmlGuardable(elemento) : '';
  }

  /** Lo que dice el texto, sin etiquetas: es lo que decide si un comentario está vacío. */
  textoPlano(): string {
    return (this.lienzo()?.nativeElement.textContent ?? '').trim();
  }

  /** Vacía el lienzo y olvida sus archivos: lo que se hace después de enviar el comentario. */
  limpiar(): void {
    const elemento = this.lienzo()?.nativeElement;

    this.revocarProvisionales();
    this.nuevos.clear();
    this.archivos.set([]);
    this.error.set('');
    this.ultimoPintado = '';

    if (elemento) {
      elemento.innerHTML = '';
    }

    this.valor.set('');
  }

  // ------------------------------------------------------------------ los adjuntos

  /**
   * Recoge lo que se ha pegado: **un archivo, o texto**.
   *
   * Lo que se pega con formato —el HTML de un correo, lo de Word— **se pega como texto plano**: traería
   * etiquetas que no están en la lista blanca, y el comentario se rechazaría entero por algo que la
   * persona no ha escrito. Los saltos de línea se respetan, que es lo que se ve al pegar.
   *
   * Un **archivo** pegado (una captura, lo que se copia del explorador) sí entra, y entra como
   * adjunto, en el sitio donde está el cursor.
   */
  protected pegar(evento: ClipboardEvent): void {
    const archivos = Array.from(evento.clipboardData?.files ?? []);
    if (archivos.length) {
      evento.preventDefault();
      this.meter(archivos);
      return;
    }

    const texto = evento.clipboardData?.getData('text/plain') ?? '';
    if (!texto) {
      return;
    }

    evento.preventDefault();
    this.insertarHtml(textoComoHtml(texto));
    this.actualizar();
  }

  protected arrastrando(evento: DragEvent): void {
    // Sin esto, el navegador abre el archivo en una pestaña en vez de dejarlo caer aquí.
    evento.preventDefault();
    this.activo.set(true);
  }

  protected saliendo(): void {
    this.activo.set(false);
  }

  protected soltar(evento: DragEvent): void {
    evento.preventDefault();
    this.activo.set(false);

    this.meter(Array.from(evento.dataTransfer?.files ?? []));
  }

  protected elegirArchivos(evento: Event): void {
    const campo = evento.target as HTMLInputElement;
    this.meter(Array.from(campo.files ?? []));
    // El campo se vacía para que elegir el mismo archivo dos veces vuelva a disparar el cambio.
    campo.value = '';
  }

  /**
   * Mete los archivos que valgan **donde está el cursor**, cada uno con su vista previa.
   *
   * Lo que no vale se dice con el mismo aviso de siempre y no entra: la lista cerrada de extensiones y
   * el tope de peso son los mismos en las tres puertas (decisión 43).
   */
  private meter(nuevos: readonly File[]): void {
    const rechazados: string[] = [];
    const admitidos: File[] = [];

    for (const archivo of nuevos) {
      if (!extensionAdmitida(archivo.name) || archivo.size > MAX_ARCHIVO_BYTES) {
        rechazados.push(archivo.name);
        continue;
      }

      admitidos.push(archivo);
    }

    this.error.set(
      rechazados.length ? `${this.textos().avisoDeRechazo} ${rechazados.join(', ')}` : '',
    );

    if (!admitidos.length) {
      return;
    }

    // Se pone el cursor al final si el foco no está dentro: es lo que hace que el botón de adjuntar
    // meta el archivo en el texto aunque el foco esté en el propio botón.
    this.prepararElCursor();

    for (const archivo of admitidos) {
      const url = URL.createObjectURL(archivo);

      this.nuevos.set(archivo.name, archivo);
      this.provisionales.set(archivo.name, url);

      this.insertarHtml(this.referencia(archivo.name, url));
    }

    this.actualizar();
  }

  /** Cómo se ve un adjunto dentro del lienzo, con su dirección provisional. */
  private referencia(nombre: string, url: string): string {
    const escapado = escaparAtributo(nombre);

    if (esImagen(nombre)) {
      return `<img data-adjunto="${escapado}" src="${url}" alt="${escapado}" class="${CLASE_DE_VISTA_PREVIA}">`;
    }

    if (esVideo(nombre)) {
      // **Con controles aquí**: escribiendo se comprueba lo que se adjunta, y el visor grande no
      // existe en este cuadro; al leerlo, el vídeo va con su botón de reproducir y el visor.
      return `<video data-adjunto="${escapado}" src="${url}" controls class="${CLASE_DE_VISTA_PREVIA}"></video>`;
    }

    return `<a data-adjunto="${escapado}" href="${url}">${escaparTexto(nombre)}</a>`;
  }

  /**
   * Resuelve cada referencia a su dirección: la provisional de lo que se acaba de adjuntar, o la que
   * la pantalla ha traído para un adjunto que ya está en el ticket.
   *
   * Lo que no se puede resolver **no se queda mudo**: se marca, porque un archivo que se nombró en el
   * texto y no está es lo que se ve cuando una subida falla (decisión 44).
   */
  private resolver(elemento: HTMLElement): void {
    for (const adjunto of Array.from(elemento.querySelectorAll<HTMLElement>('[data-adjunto]'))) {
      const nombre = adjunto.getAttribute('data-adjunto') ?? '';
      const url = this.provisionales.get(nombre) ?? this.direcciones().get(nombre);

      if (!url) {
        if (!adjunto.classList.contains(CLASE_DE_LO_QUE_FALTA)) {
          adjunto.replaceWith(this.marcaDeLoQueFalta(nombre));
        }
        continue;
      }

      if (adjunto.tagName === 'IMG' || adjunto.tagName === 'VIDEO') {
        adjunto.setAttribute('src', url);
        adjunto.className = CLASE_DE_VISTA_PREVIA;
        if (adjunto.tagName === 'IMG') {
          adjunto.setAttribute('alt', nombre);
        } else {
          adjunto.setAttribute('controls', '');
        }
        continue;
      }

      adjunto.setAttribute('href', url);
      adjunto.className = 'text-primario underline';
    }
  }

  /** El recuadro que dice que ese archivo no está, donde estaba. */
  private marcaDeLoQueFalta(nombre: string): HTMLElement {
    const marca = document.createElement('span');

    marca.className = `${CLASE_DE_LO_QUE_FALTA} mx-1 inline-flex items-center gap-1 rounded-md border border-aviso bg-aviso-suave px-2 py-0.5 text-xs text-aviso`;
    marca.textContent = `${this.textos().adjuntoNoSubio}: ${nombre || '—'}`;

    return marca;
  }

  /** Los archivos nuevos que están dentro del texto ahora mismo, en el orden en que aparecen. */
  private apuntarLosArchivosDeDentro(elemento: HTMLElement): void {
    const dentro: File[] = [];

    for (const adjunto of Array.from(elemento.querySelectorAll<HTMLElement>('[data-adjunto]'))) {
      const archivo = this.nuevos.get(adjunto.getAttribute('data-adjunto') ?? '');
      if (archivo) {
        dentro.push(archivo);
      }
    }

    this.archivos.set(dentro);
  }

  // ------------------------------------------------------------------ el texto

  protected alEscribir(): void {
    this.actualizar();
  }

  /** Lo que ha cambiado el texto: se saca el HTML guardable y se revisan los archivos de dentro. */
  private actualizar(): void {
    const elemento = this.lienzo()?.nativeElement;
    if (!elemento) {
      return;
    }

    const html = htmlGuardable(elemento);
    this.ultimoPintado = html;
    this.valor.set(html);
    this.apuntarLosArchivosDeDentro(elemento);
  }

  protected alSalir(): void {
    this.activo.set(false);
  }

  // ------------------------------------------------------------------ la barra

  /** Los botones, con lo que hacen. Se arma en el componente para no repetir la plantilla. */
  protected botones(): readonly {
    texto: string;
    muestra: string;
    clase: string;
    aplicar: () => void;
  }[] {
    return [
      {
        texto: this.textos().negrita,
        muestra: 'N',
        clase: 'font-bold',
        aplicar: () => this.formato('bold'),
      },
      {
        texto: this.textos().cursiva,
        muestra: 'C',
        clase: 'italic',
        aplicar: () => this.formato('italic'),
      },
      {
        texto: this.textos().subrayado,
        muestra: 'S',
        clase: 'underline',
        aplicar: () => this.formato('underline'),
      },
      {
        texto: this.textos().tachado,
        muestra: 'T',
        clase: 'line-through',
        aplicar: () => this.formato('strikeThrough'),
      },
      {
        texto: this.textos().listaVinietas,
        muestra: '≡',
        clase: '',
        aplicar: () => this.formato('insertUnorderedList'),
      },
      {
        texto: this.textos().listaNumerada,
        muestra: '1.',
        clase: '',
        aplicar: () => this.formato('insertOrderedList'),
      },
      { texto: this.textos().enlace, muestra: '↗', clase: '', aplicar: () => this.abrirElEnlace() },
    ];
  }

  /**
   * Los botones **no le quitan el foco al lienzo**: lo que hay seleccionado es sobre lo que actúan, y
   * un botón que se lleva la selección al pulsarlo no sirve para nada.
   */
  /**
   * Mira lo que se lleva escrito detrás del cursor: si es una `@` con lo que sea detrás —sin espacios—,
   * abre la lista de sugerencias filtrada por eso (decisión 74). Si no, la cierra.
   *
   * Se lee del texto del lienzo y no de un campo aparte porque la mención **va dentro del comentario**:
   * es lo que se está escribiendo, y lo que hay que sustituir al elegir a alguien.
   */
  protected alEscribirEnElLienzo(): void {
    const rango = this.rangoActual();
    const nodo = rango?.startContainer;

    if (!rango || !nodo || nodo.nodeType !== Node.TEXT_NODE) {
      this.mencionAMedias.set(null);

      return;
    }

    const escrito = mencionEscrita((nodo.textContent ?? '').slice(0, rango.startOffset));

    this.mencionAMedias.set(escrito);

    // **Y la lista sale donde está el cursor** (decisión 89), no encima del bloque del comentario.
    this.posicionDeLasSugerencias.set(escrito === null ? null : this.posicionJuntoAlCursor(rango));
  }

  /** Dónde poner la lista: el rectángulo del cursor, pasado por el ayudante puro. */
  private posicionJuntoAlCursor(rango: Range): { izquierda: number; arriba: number; maxAlto: number } | null {
    const caja = rango.getBoundingClientRect();
    if (!caja.width && !caja.height) {
      // Un cursor sin caja (raro, pero pasa) no da dónde ponerla: mejor no enseñarla mal puesta.
      return null;
    }

    return colocacionDeLasSugerencias(
      { izquierda: caja.left, arriba: caja.top, abajo: caja.bottom },
      ALTO_DE_LA_LISTA,
      { ancho: window.innerWidth, alto: window.innerHeight },
    );
  }

  /** Cierra la lista de sugerencias sin elegir a nadie; lo escrito se queda como está. */
  protected cerrarLasSugerencias(): void {
    this.mencionAMedias.set(null);
    this.posicionDeLasSugerencias.set(null);
  }

  /**
   * Mete la mención **en el sitio de la `@` que se está escribiendo**: borra la arroba y lo tecleado
   * detrás, y deja el nombre. Así lo que queda en el comentario es la mención, y no `@sop` más el nombre.
   */
  protected etiquetarDesdeElTexto(persona: PersonaDelEditor): void {
    const escrito = this.mencionAMedias() ?? '';
    const rango = this.rangoActual();

    if (rango && rango.startContainer.nodeType === Node.TEXT_NODE) {
      // Se borra la `@` y lo que se haya escrito detrás, que es lo que se está sustituyendo.
      const borrado = document.createRange();
      const nodo = rango.startContainer;
      const desde = Math.max(0, rango.startOffset - escrito.length - 1);

      borrado.setStart(nodo, desde);
      borrado.setEnd(nodo, rango.startOffset);
      borrado.deleteContents();

      const seleccion = document.getSelection();
      seleccion?.removeAllRanges();
      seleccion?.addRange(borrado);
    }

    this.etiquetarA(persona);
  }

  /** Abre o cierra el buscador de personas. */
  protected alternarPersonas(): void {
    this.eligiendoPersona.update((abierto) => !abierto);
  }

  /**
   * Mete la mención donde está el cursor.
   *
   * Lo que se escribe es **el identificador y el nombre**: el identificador es el dato —es lo que se
   * guarda y a lo que apunta la observación— y el nombre es lo que se lee. Detrás va un espacio, para
   * poder seguir escribiendo sin pegarse a la mención.
   */
  protected etiquetarA(persona: PersonaDelEditor): void {
    const nombre = escaparTexto(persona.nombre);

    this.insertarHtml(`<span ${ATRIBUTO_DE_LA_MENCION}="${persona.id}">${nombre}</span>&nbsp;`);
    this.eligiendoPersona.set(false);
    this.lienzo()?.nativeElement.focus();
  }

  protected noQuitarElFoco(evento: Event): void {
    evento.preventDefault();
  }

  /** Aplica un comando de formato sobre lo seleccionado. */
  private formato(orden: string): void {
    const elemento = this.lienzo()?.nativeElement;
    if (!elemento) {
      return;
    }

    if (typeof document.execCommand !== 'function') {
      return;
    }

    // El formato se pone con etiquetas y no con un `style` en línea, que no entra en la lista blanca.
    document.execCommand('styleWithCSS', false, 'false');
    elemento.focus();
    document.execCommand(orden, false);
    this.actualizar();
  }

  private abrirElEnlace(): void {
    this.rangoGuardado = this.rangoActual();
    this.direccion.set('');
    this.enlaceInvalido.set(false);
    this.pidiendoEnlace.set(true);
  }

  protected alEscribirLaDireccion(evento: Event): void {
    this.direccion.set((evento.target as HTMLInputElement).value);
    this.enlaceInvalido.set(false);
  }

  protected cerrarElEnlace(): void {
    this.pidiendoEnlace.set(false);
    this.enlaceInvalido.set(false);
    this.rangoGuardado = null;
  }

  /**
   * Pone el enlace sobre lo seleccionado, o lo inserta con su dirección si no había nada seleccionado.
   *
   * **La dirección se comprueba antes**: lo que no empiece por `http`, `https` o `mailto` no se pone y
   * se dice por qué, en vez de rechazar el comentario entero al guardarlo.
   */
  protected ponerElEnlace(): void {
    const destino = this.direccion().trim();

    if (!direccionAdmitida(destino)) {
      this.enlaceInvalido.set(true);
      return;
    }

    this.restaurarElRango();
    this.insertarHtml(`<a href="${escaparAtributo(destino)}">${escaparTexto(destino)}</a>`, true);
    this.cerrarElEnlace();
    this.actualizar();
  }

  // ------------------------------------------------------------------ el cursor

  /** Pone el cursor al final si el foco no está dentro del lienzo: el archivo cae donde toca. */
  private prepararElCursor(): void {
    const elemento = this.lienzo()?.nativeElement;
    if (!elemento) {
      return;
    }

    const seleccion = document.getSelection();
    if (seleccion && seleccion.rangeCount && elemento.contains(seleccion.anchorNode)) {
      return;
    }

    elemento.focus();

    const rango = document.createRange();
    rango.selectNodeContents(elemento);
    rango.collapse(false);

    seleccion?.removeAllRanges();
    seleccion?.addRange(rango);
  }

  /** El rango que hay puesto, o uno al final del lienzo. */
  private rangoActual(): Range | null {
    const elemento = this.lienzo()?.nativeElement;
    if (!elemento) {
      return null;
    }

    const seleccion = document.getSelection();
    if (seleccion && seleccion.rangeCount && elemento.contains(seleccion.anchorNode)) {
      return seleccion.getRangeAt(0).cloneRange();
    }

    const rango = document.createRange();
    rango.selectNodeContents(elemento);
    rango.collapse(false);

    return rango;
  }

  private restaurarElRango(): void {
    const rango = this.rangoGuardado;
    if (!rango) {
      return;
    }

    const seleccion = document.getSelection();
    seleccion?.removeAllRanges();
    seleccion?.addRange(rango);
  }

  /**
   * Escribe HTML donde está el cursor.
   *
   * Es `execCommand('insertHTML')` cuando el navegador lo tiene —que es lo que inserta en el sitio del
   * cursor respetando la selección—, y a mano cuando no lo tiene (jsdom, que es con lo que corren las
   * pruebas de unidad, no lo implementa).
   */
  private insertarHtml(html: string, sobreLaSeleccion = false): void {
    const elemento = this.lienzo()?.nativeElement;
    if (!elemento) {
      return;
    }

    this.prepararElCursor();

    if (typeof document.execCommand === 'function') {
      document.execCommand('insertHTML', false, html);
      return;
    }

    const rango = this.rangoActual();
    if (!rango) {
      return;
    }

    if (sobreLaSeleccion && !rango.collapsed) {
      rango.deleteContents();
    }

    const nodo = rango.createContextualFragment(html);
    const ultimo = nodo.lastChild;
    rango.insertNode(nodo);

    if (ultimo) {
      const despues = document.createRange();
      despues.setStartAfter(ultimo);
      despues.collapse(true);

      const seleccion = document.getSelection();
      seleccion?.removeAllRanges();
      seleccion?.addRange(despues);
    }
  }

  private revocarProvisionales(): void {
    for (const url of this.provisionales.values()) {
      URL.revokeObjectURL(url);
    }

    this.provisionales.clear();
  }

  /** El lienzo, con el estilo del campo y el del arrastre encima. */
  protected clasesDelLienzo(): string {
    const base = [
      'min-h-24 rounded-md border px-3 py-2 text-sm text-texto whitespace-pre-line',
      '[&_img]:max-w-full [&_img]:rounded-md [&_video]:max-w-full [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5',
      this.activo() ? 'border-primario bg-superficie-suave' : 'border-borde-fuerte bg-superficie',
      this.deshabilitado() ? 'opacity-60' : '',
    ];

    return base.join(' ');
  }
}
