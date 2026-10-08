import {
  Component,
  OnDestroy,
  computed,
  effect,
  inject,
  input,
  output,
  signal,
  viewChild,
} from '@angular/core';
import {
  DomSanitizer,
  type SafeHtml,
  type SafeResourceUrl,
  type SafeUrl,
} from '@angular/platform-browser';

import { TranslationService } from '../../core/i18n/translation.service';
import { BrandService } from '../../core/services/brand.service';
import { SessionService } from '../../core/services/session.service';
import { Adjunto, type AdjuntoDeTicket } from '../../shared/components/adjunto';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Dialogo } from '../../shared/components/dialogo';
import {
  EditorConAdjuntos,
  esImagen,
  esVideo,
  textoComoHtml,
  type PersonaDelEditor,
  type TextosDelEditorConAdjuntos,
} from '../../shared/components/editor-con-adjuntos';
import { Estado } from '../../shared/components/estado';
import { Selector, type OpcionSelector } from '../../shared/components/selector';
import { Tarjeta } from '../../shared/components/tarjeta';
import { claveDelError } from '../../shared/errores';
import {
  AccionesDelTicket,
  accionesDisponibles,
  type ContextoDeAcciones,
} from './acciones-del-ticket';
import {
  etiquetaDeEstado,
  fechaHora,
  interpolar,
  nombreDe,
  opcionesDeCategoria,
  puedeEditarElTicket,
  puedeEtiquetarElTicket,
  sePuedePrevisualizar,
  textoDelResumen,
} from './etiquetas';
import { ModalEtiquetas } from './modal-etiquetas';
import { MejorarRedaccion } from './mejorar-redaccion';
import {
  TicketsService,
  type Adjunto as AdjuntoDeLaApi,
  type Categoria,
  type Comentario,
  type DetalleDeTicket,
  type EntradaDeHistorial,
  type Observador,
  type Persona,
  type Resumen,
} from './tickets.service';

/** Quién puede ser responsable de cada tipo de ticket. */
export interface Responsables {
  readonly main: readonly Persona[];
  readonly internal: readonly Persona[];
}

/**
 * Cómo se ve dentro de la conversación lo que el texto nombra: la imagen y el vídeo **con un tope de
 * 480 × 360** —una captura a pantalla completa no puede comerse la conversación— y **el clic abre el
 * visor grande** para verlos enteros y descargarlos (docs/modules/tickets.md, decisiones 46 y 55).
 *
 * **El ancho es `min(100%, 480px)` y no `480px`**, y esto no es un detalle: `max-w-[480px]` sólo
 * limita, así que en un móvil de 412 px de ancho la imagen de 480 **se salía de la pantalla** y
 * ensanchaba el documento entero. Lo encontró la capa de Playwright: el clic en un botón de la misma
 * tarjeta empezó a fallar porque la página tenía scroll horizontal.
 */
/**
 * Cómo se pinta una mención: el nombre de una persona, marcado.
 *
 * Los colores salen del tema, como todo (`AGENTS.md`), y no lleva `@` delante porque el nombre ya se
 * lee solo y el símbolo se confundiría con el botón del editor.
 */
const CLASE_DE_LA_MENCION =
  'rounded bg-superficie-suave px-1 font-medium text-primario ring-1 ring-borde';

const CLASE_DE_LO_QUE_SE_VE =
  'max-h-[360px] max-w-[min(100%,480px)] rounded-md border border-borde';

/** El marco de un vídeo: el tope de arriba y, encima, el botón de reproducir. */
const CLASE_DEL_VIDEO = 'relative inline-flex max-w-full';

/** El botón de reproducir del centro, que es lo que dice que eso se pulsa para verlo. */
const CLASE_DEL_BOTON_DE_REPRODUCIR =
  'absolute inset-0 flex items-center justify-center rounded-md bg-black/30 text-white transition hover:bg-black/40';

/**
 * **Cuántas veces se vuelve a leer la ficha esperando al motor** (quinta enmienda del 2026-09-29,
 * punto 5): ocho relecturas cada `CADA_RECARGA_DEL_MOTOR`, unos 64 s. El motor tarda entre 12 y 24
 * segundos por campo y más si tiene cola, así que los cuatro intentos de antes (32 s) se quedaban
 * cortos y el campo parecía colgado.
 */
const INTENTOS_DEL_MOTOR = 8;

/** Cada cuánto se relee la ficha mientras el motor escribe. */
const CADA_RECARGA_DEL_MOTOR = 8000;

/** Un renglón de la línea de tiempo: un comentario o algo que hizo el sistema. */
interface Renglon {
  readonly clave: string;
  readonly fecha: string;
  readonly comentario?: Comentario;
  readonly historial?: EntradaDeHistorial;
}

/**
 * La única acción que sigue viviendo en la conversación: borrar un comentario.
 *
 * Las acciones que **mueven el ticket** —empezar, preguntar, escalar, resolver, cerrar, reabrir— se
 * fueron al desplegable de `app-acciones-del-ticket` (decisión 78), que es el mismo en los dos sitios
 * donde vive. Borrar un comentario es de un comentario, no del ticket, así que se queda aquí.
 */
type Accion = 'borrarComentario';

interface Pendiente {
  readonly accion: Accion;
  readonly comentarioID?: number;
}

/**
 * La vista de un ticket: **la conversación a la izquierda y la ficha con las acciones a la derecha**
 * (`docs/interfaz-y-experiencia.md`, sección 3.4), en móvil una detrás de otra.
 *
 * Es la misma pieza para el principal y para el interno: el detalle de un ticket interno **es el mismo
 * detalle**, con su número y su conversación (sección 3.7). Lo que cambia es qué se puede hacer
 * dentro, y eso depende del papel y del estado: **la interfaz no ofrece lo que no se puede hacer**
 * (sección 4.3), en vez de ofrecerlo y que el backend lo rechace.
 */
@Component({
  selector: 'app-ticket-vista',
  imports: [
    AccionesDelTicket,
    Adjunto,
    Aviso,
    Boton,
    Campo,
    Dialogo,
    EditorConAdjuntos,
    Estado,
    ModalEtiquetas,
    MejorarRedaccion,
    Selector,
    Tarjeta,
  ],
  templateUrl: './ticket-vista.html',
})
export class TicketVista implements OnDestroy {
  readonly detalle = input.required<DetalleDeTicket>();

  /**
   * **Sólo la conversación**, sin la ficha.
   *
   * Es lo que se usa en la vista doble (decisión del responsable, 2026-09-27): con las dos
   * conversaciones y las dos fichas, la pantalla se ve saturada; en «Los dos» se lee lo que se ha dicho
   * y, para actuar o mirar la ficha, se elige ese ticket con el conmutador.
   */
  readonly soloLaConversacion = input(false);

  /**
   * Si esta vista pinta **su propia cabecera** con el número y el estado.
   *
   * Sólo en la vista doble: con un ticket solo, la cabecera de la página ya los pinta, y tenerlos dos
   * veces era lo que el responsable vio en `CS-2026-0017` (decisión 74).
   */
  readonly conCabecera = input(false);
  readonly responsables = input<Responsables>({ main: [], internal: [] });
  /**
   * El catálogo de categorías, que es lo que necesita la ficha para poder cambiarla. Lo carga el
   * detalle y lo pasa a las dos mitades de la vista doble: la categoría es del principal y **el interno
   * la hereda** (decisión 67), así que ambas pantallas enseñan la misma.
   */
  readonly categorias = input<readonly Categoria[]>([]);
  /** Lo que se acaba de hacer en este ticket: la pantalla vuelve a leerlo. */
  readonly cambio = output<void>();

  private readonly tickets = inject(TicketsService);
  private readonly textos = inject(TranslationService);
  private readonly marca = inject(BrandService);
  private readonly sesion = inject(SessionService);
  private readonly saneador = inject(DomSanitizer);

  /** El texto del comentario que se está escribiendo, ya listo para guardar. */
  protected readonly comentario = signal('');

  /** La caja de escribir del comentario, que es quien tiene el texto y sus archivos. */
  private readonly elDelComentario = viewChild<EditorConAdjuntos>('elDelComentario');

  /** El editor de la descripción, cuando se está editando el ticket. */
  private readonly elDeLaDescripcion = viewChild<EditorConAdjuntos>('elDeLaDescripcion');

  /** Y el del comentario que se está editando: sólo se edita uno a la vez. */
  private readonly elDeLaEdicion = viewChild<EditorConAdjuntos>('elDeLaEdicion');
  private readonly modalDeRedaccion = viewChild<MejorarRedaccion>('modalDeRedaccion');
  private readonly destinoDeRedaccion = signal<'description' | 'comment' | 'editedComment'>('comment');

  /**
   * **Los archivos que no han subido**, con el comentario al que iban.
   *
   * El comentario ya está publicado cuando esto se llena, así que no se borra: se avisa de qué falta y
   * se puede reintentar, que es lo que evita acabar con el comentario escrito dos veces
   * (docs/modules/tickets.md, decisión 44).
   */
  protected readonly adjuntosQueFaltan = signal<
    readonly { archivo: File; comentarioID: number | undefined }[]
  >([]);

  protected readonly trabajando = signal(false);
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);

  protected readonly editando = signal<number | null>(null);
  protected readonly textoEditado = signal('');

  /**
   * **El asunto y la descripción se editan donde se leen** (decisión 74): cada uno tiene su lápiz y su
   * propio estado, porque son dos ediciones independientes y sólo se hace una a la vez. Antes vivían
   * las dos en un diálogo con un solo «Editar», y por eso no se encontraba.
   */
  protected readonly editandoAsunto = signal(false);
  protected readonly asuntoEditado = signal('');
  protected readonly editandoDescripcion = signal(false);
  protected readonly descripcionEditada = signal('');

  /**
   * El buscador de personas para **añadir un observador desde su lista** (decisión 74): se abre con
   * «Añadir» y deja elegir a quien todavía no observa, filtrando por nombre o por correo.
   */
  protected readonly eligiendoObservador = signal(false);
  protected readonly buscandoObservador = signal('');

  /** El comentario que espera confirmación para borrarse: la única ventana que queda aquí. */
  protected readonly pendiente = signal<Pendiente | null>(null);
  protected readonly elegido = signal('');

  /**
   * Si está abierto el **modal de reasignación** (quinta enmienda del 2026-09-29, punto 2), que abre
   * el botón que va junto al nombre del responsable.
   */
  protected readonly mostrandoAsignacion = signal(false);

  /**
   * La categoría que se está enseñando en la ficha.
   *
   * Sólo hace falta para el desplegable: el ticket se recarga entero con cada cambio, así que la
   * categoría de verdad es siempre la del detalle. **Las etiquetas ya no se editan aquí una a una**:
   * viven en su modal y se guardan **de una vez** (decisión 82), así que la fuente de lo que se enseña
   * sigue siendo lo que tiene el ticket.
   */
  protected readonly categoriaElegida = signal('');

  /**
   * Si está abierto el **modal de etiquetas** (decisión 82), que es donde se marcan y desmarcan todas
   * y se guardan de una vez. Vive aquí porque es la ficha quien manda el `PATCH` y avisa del resultado.
   */
  protected readonly mostrandoEtiquetas = signal(false);

  /**
   * Las direcciones de los adjuntos ya subidos, **por identificador**.
   *
   * Un adjunto **no se puede enlazar con su dirección de la API**: la descarga pasa por la comprobación
   * de permisos y necesita la cabecera de la sesión, y el navegador no la manda al pintar un `<img>`.
   * Así que el archivo se pide con `HttpClient`, se envuelve en una dirección `blob:` del navegador y
   * **esa** es la que se pinta. Se cachea por identificador para no pedir dos veces lo mismo, y se
   * libera al destruir la pantalla (sección 2.3).
   */
  private readonly direcciones = signal<ReadonlyMap<number, string>>(new Map());

  /** Las mismas direcciones por nombre de archivo, que es como las nombra el texto. */
  protected readonly direccionesPorNombre = computed<ReadonlyMap<string, string>>(() => {
    const porNombre = new Map<string, string>();

    for (const adjunto of this.detalle().attachments) {
      const url = this.direcciones().get(adjunto.id);
      if (url) {
        porNombre.set(adjunto.filename, url);
      }
    }

    return porNombre;
  });

  /** Los que se están pidiendo ahora mismo, para no pedir el mismo dos veces. */
  private readonly pidiendo = new Set<number>();

  /** Las direcciones que ha creado esta pantalla, para liberarlas al salir. */
  private readonly creadas = new Set<string>();

  /**
   * La vista previa de un adjunto.
   *
   * Las direcciones van saneadas a mano (`blob:` es lo que el navegador da para un archivo que se
   * acaba de traer, y Angular lo bloquea por defecto): la CSP de producción ya admite `blob:` en
   * `img-src` y `frame-src`, así que enseñarlo es exactamente lo que se quiere.
   */
  protected readonly vistaPrevia = signal<{
    adjunto: AdjuntoDeLaApi;
    nombre: string;
    url: string;
    imagen: SafeUrl;
    marco: SafeResourceUrl;
    pdf: boolean;
    video: boolean;
  } | null>(null);

  private readonly idioma = this.textos.idioma;

  constructor() {
    // **Los adjuntos que el texto nombra se piden en cuanto se sabe qué texto hay.** La dirección de un
    // adjunto no se puede enlazar desde la API —hace falta la cabecera de la sesión—, así que hay que
    // traerla: sin esto, el texto se quedaría con su hueco en vez de con su imagen.
    effect(() => {
      this.detalle();
      this.pedirLosQueNombraElTexto();
    });

    // **La categoría del desplegable se pone al día con el ticket**: quien guarda, recarga, y esto deja
    // el selector con lo que de verdad se ha guardado, sin inventar estado.
    effect(() => {
      const ticket = this.detalle().ticket;
      this.categoriaElegida.set(ticket.category ? String(ticket.category.id) : '');
    });

    // **El motor escribe en segundo plano: la ficha se vuelve a leer sola mientras quede algo
    // pendiente** (quinta enmienda del 2026-09-29, punto 5). Al ser un efecto sobre el detalle, cubre
    // cualquier recarga —comentar, editar, mover el estado, reabrir— sin repetir la llamada en cada
    // acción, y se para en cuanto los dos campos están `listo`.
    effect(() => {
      const detalle = this.detalle();
      const pendientes = this.resumenesPendientes();

      if (!pendientes) {
        this.pararLaEsperaDelMotor();
        return;
      }

      // Otro cálculo —el ticket se ha movido— abre una espera nueva y devuelve el margen entero.
      if (this.relecturaDe !== detalle.ticket.updatedAt) {
        this.relecturaDe = detalle.ticket.updatedAt;
        this.intentosDelMotor = 0;
      }

      this.programarLaRelecturaDelMotor();
    });
  }

  protected readonly ticket = computed(() => this.detalle().ticket);
  protected readonly papel = computed(() => this.sesion.usuario()?.role ?? '');
  protected readonly esUsuario = computed(() => this.papel() === 'usuario');
  protected readonly esSoporte = computed(() => this.papel() === 'soporte');
  protected readonly esDesarrollo = computed(() => this.papel() === 'desarrollo');
  /** El Administrador mira los tickets: no comenta, no mueve estados y no asigna (sección 4.4). */
  protected readonly soloLectura = this.sesion.esAdministrador;
  protected readonly esSuyo = computed(
    () => this.ticket().requester?.id === this.sesion.usuario()?.id,
  );
  protected readonly cerrado = computed(() => this.ticket().state === 'cerrado');

  /** La conversación con el historial intercalado, en orden: **una sola línea de tiempo**. */
  protected readonly linea = computed<Renglon[]>(() => {
    const detalle = this.detalle();

    const renglones: Renglon[] = [
      ...detalle.comments.map((comentario) => ({
        clave: `c-${comentario.id}`,
        fecha: comentario.createdAt,
        comentario,
      })),
      ...detalle.history.map((entrada) => ({
        clave: `h-${entrada.id}`,
        fecha: entrada.createdAt,
        historial: entrada,
      })),
    ];

    return renglones.sort((uno, otro) => uno.fecha.localeCompare(otro.fecha));
  });

  /**
   * Los adjuntos que van con la descripción inicial y **el texto no nombra**.
   *
   * Los que sí nombra se ven dentro del texto, en su sitio; los demás siguen en su lista al final,
   * como hasta ahora (decisión 49). Es lo que hace que los tickets y comentarios que ya existen —cuyo
   * texto no referencia nada— se vean exactamente igual que antes.
   */
  protected readonly adjuntosIniciales = computed(() =>
    this.adjuntosSueltos(
      this.detalle().attachments.filter((adjunto) => !adjunto.commentId),
      this.ticket().description,
    ),
  );

  /**
   * El texto de la descripción y el de cada comentario, **pintados con formato y con sus adjuntos en
   * su sitio**: la imagen y el vídeo en línea, el PDF y lo demás como enlaces.
   *
   * Se calcula de una vez por conversación y no en la plantilla: montar el HTML nodo a nodo es trabajo,
   * y una llamada desde la plantilla lo haría en cada comprobación de cambios.
   */
  protected readonly cuerpos = computed<ReadonlyMap<number, SafeHtml>>(() => {
    const cuerpos = new Map<number, SafeHtml>();

    for (const comentario of this.detalle().comments) {
      cuerpos.set(comentario.id, this.pintar(comentario.body));
    }

    return cuerpos;
  });

  /** Y el de la descripción, que es el primer mensaje del ticket. */
  protected readonly cuerpoDeLaDescripcion = computed(() => this.pintar(this.ticket().description));

  protected readonly listaDeResponsables = computed<readonly OpcionSelector[]>(() => {
    const personas = this.ticket().internal
      ? this.responsables().internal
      : this.responsables().main;

    return personas.map((persona) => ({
      valor: String(persona.id),
      etiqueta: nombreDe(persona),
      grupo: this.t().tickets.responsable,
    }));
  });

  protected t() {
    return this.textos.textos();
  }

  /**
   * El **motivo**: lo que redactó el motor, o —si es un interno— **el motivo del escalado**, que es
   * suyo y no se genera (docs/modules/ai.md, decisión 5).
   */
  protected motivo(): string {
    const ticket = this.ticket();

    if (ticket.internal && ticket.reason) {
      return ticket.reason;
    }

    return textoDelResumen(ticket.insights?.motivo, this.textos.idioma(), this.t());
  }

  /** Y la **última acción**, que siempre redacta el motor. */
  protected ultimaAccion(): string {
    return textoDelResumen(this.ticket().insights?.ultimaAccion, this.textos.idioma(), this.t());
  }

  /**
   * A quién se puede etiquetar: **técnicos y desarrolladores activos** (decisión 59), que son los
   * mismos que pueden ser responsables. Se juntan los de los dos tipos de ticket y se quitan los
   * repetidos: quien puede ver un principal puede ver un interno.
   */
  protected readonly personasParaEtiquetar = computed<readonly PersonaDelEditor[]>(() => {
    const personas = new Map<number, PersonaDelEditor>();

    for (const persona of [...this.responsables().main, ...this.responsables().internal]) {
      if (!personas.has(persona.id)) {
        // **Con el correo**, para poder buscar por él al etiquetar escribiendo la arroba: con dos
        // personas que se llaman igual, el correo es lo único que las distingue (decisión 74).
        personas.set(persona.id, {
          id: persona.id,
          nombre: nombreDe(persona),
          email: persona.email,
        });
      }
    }

    return [...personas.values()];
  });

  /** Etiquetar es de **Soporte y Desarrollo** (decisión 59): el usuario y el Administrador, no. */
  protected readonly puedeEtiquetar = computed(
    () => !this.soloLectura() && (this.esSoporte() || this.esDesarrollo()),
  );

  /** Los que observan este ticket, que vienen con la ficha. */
  protected readonly observadores = computed(() => this.detalle().observers ?? []);

  /** Y quitar a uno: **cualquier técnico o desarrollador** (decisión 63). */
  protected async quitarObservador(observador: Observador): Promise<void> {
    await this.actuar(async () => {
      await this.tickets.quitarObservador(this.ticket().number, observador.account.id);

      return {
        forma: 'exito',
        texto: interpolar(this.t().tickets.observadorQuitado, {
          quien: this.nombreDe(observador.account),
        }),
      };
    });
  }

  /**
   * A quién se puede añadir como observador: **técnicos y desarrolladores activos** que todavía no
   * observan, que son los mismos que pueden ser responsables (decisión 74). Se juntan los de los dos
   * tipos de ticket y se quitan los repetidos, como en la lista de a quién etiquetar.
   */
  protected readonly personasParaObservar = computed<readonly Persona[]>(() => {
    const yaObservan = new Set(this.observadores().map((observador) => observador.account.id));
    const personas = new Map<number, Persona>();

    for (const persona of [...this.responsables().main, ...this.responsables().internal]) {
      if (!yaObservan.has(persona.id) && !personas.has(persona.id)) {
        personas.set(persona.id, persona);
      }
    }

    return [...personas.values()];
  });

  /** Y las que encajan con lo que se escribe en el buscador: **por nombre o por correo**. */
  protected readonly personasQueEncuadran = computed<readonly Persona[]>(() => {
    const escrito = this.buscandoObservador().trim().toLowerCase();

    if (!escrito) {
      return this.personasParaObservar();
    }

    return this.personasParaObservar().filter(
      (persona) =>
        nombreDe(persona).toLowerCase().includes(escrito) ||
        (persona.email ?? '').toLowerCase().includes(escrito),
    );
  });

  /** Abre o cierra el buscador de personas del «Añadir». */
  protected alternarBuscadorDeObservadores(): void {
    this.buscandoObservador.set('');
    this.mensaje.set(null);
    this.eligiendoObservador.update((abierto) => !abierto);
  }

  /**
   * Añade a alguien como observador, **y se guarda al momento** (decisión 74).
   *
   * Quien se añade así **queda igual que quien llegó por una mención**: todos son observadores y la
   * lista no distingue de dónde vino cada uno (`docs/interfaz-y-experiencia.md`, sección 3.3).
   */
  protected async anadirObservador(persona: Persona): Promise<void> {
    this.eligiendoObservador.set(false);
    this.buscandoObservador.set('');

    await this.actuar(async () => {
      await this.tickets.anadirObservador(this.ticket().number, persona.id);
      return { forma: 'exito', texto: this.t().tickets.observadorAnadido };
    });
  }

  /** Quién puede volver a pedir los resúmenes: Soporte y Desarrollo, que son los que lo trabajan. */
  protected puedeRegenerar(): boolean {
    return !this.soloLectura() && (this.esSoporte() || this.esDesarrollo());
  }

  /** Volver a pedirle al motor los dos campos. */
  protected async regenerar(): Promise<void> {
    await this.actuar(async () => {
      await this.tickets.regenerarResumenes(this.ticket().number);

      // **La espera no se arranca aquí**: la enciende el efecto que mira la ficha en cuanto llega con
      // algún campo en «pendiente», y así vale igual para regenerar que para comentar o mover el
      // ticket (quinta enmienda del 2026-09-29, punto 5).
      return { forma: 'exito', texto: this.t().tickets.resumenesPedidos };
    });
  }

  /**
   * **Los resúmenes del motor se refrescan solos** (quinta enmienda del 2026-09-29, punto 5).
   *
   * El backend los vuelve a pedir con cada movimiento del ticket, así que la ficha puede llegar con
   * alguno de los dos campos en `pendiente`. Antes sólo se volvía a leer al pulsar «Volver a
   * resumir», y el campo se quedaba en «Generando…» hasta recargar a mano: lo reportó el responsable
   * como un cuelgue. Mientras haya algo pendiente se relee la ficha cada `CADA_RECARGA_DEL_MOTOR` y
   * **se para en cuanto los dos están `listo`**, con un tope de intentos para no sondear para siempre.
   */
  private readonly resumenesPendientes = computed(() => {
    const insights = this.ticket().insights;

    return this.estaPendiente(insights?.motivo) || this.estaPendiente(insights?.ultimaAccion);
  });

  /** La fecha de la última actualización con la que se empezó a esperar: si cambia, es otro cálculo. */
  private relecturaDe: string | null = null;

  /** Las relecturas que lleva esta espera. Cada cálculo nuevo —otra `updatedAt`— reinicia la cuenta. */
  private intentosDelMotor = 0;

  /** El temporizador de la próxima relectura, para no solapar dos esperas. */
  private temporizadorDelMotor: ReturnType<typeof setTimeout> | null = null;

  /** Si el motor sigue escribiendo este campo. */
  private estaPendiente(resumen: Resumen | undefined): boolean {
    return resumen?.state === 'pendiente';
  }

  /** Programa la siguiente relectura de la ficha, si toca y queda margen. */
  private programarLaRelecturaDelMotor(): void {
    if (this.temporizadorDelMotor !== null || this.intentosDelMotor >= INTENTOS_DEL_MOTOR) {
      return;
    }

    this.temporizadorDelMotor = setTimeout(() => {
      this.temporizadorDelMotor = null;
      this.intentosDelMotor += 1;
      // El `cambio` pide la ficha al padre; el efecto vuelve a entrar con el detalle nuevo y, si
      // sigue pendiente, programa la siguiente. Si ya está listo, para.
      this.cambio.emit();
    }, CADA_RECARGA_DEL_MOTOR);
  }

  /** Para la espera y deja la cuenta a cero: ya no hay nada pendiente, o la pantalla se va. */
  private pararLaEsperaDelMotor(): void {
    if (this.temporizadorDelMotor !== null) {
      clearTimeout(this.temporizadorDelMotor);
      this.temporizadorDelMotor = null;
    }

    this.intentosDelMotor = 0;
    this.relecturaDe = null;
  }

  // ------------------------------------------------------------------ qué se puede hacer

  protected puedeComentar(): boolean {
    if (this.soloLectura() || this.cerrado()) {
      return false;
    }

    if (this.ticket().internal) {
      return this.esSoporte() || this.esDesarrollo();
    }

    return this.esSoporte() || (this.esUsuario() && this.esSuyo());
  }

  protected puedeEditar(): boolean {
    return puedeEditarElTicket({
      soloLectura: this.soloLectura(),
      cerrado: this.cerrado(),
      interno: this.ticket().internal,
      esSoporte: this.esSoporte(),
      esUsuario: this.esUsuario(),
      esSuyo: this.esSuyo(),
    });
  }

  protected puedeMejorarConIA(): boolean {
    return this.detalle().capabilities?.aiWriting === true;
  }

  protected abrirMejora(
    destino: 'description' | 'comment' | 'editedComment',
    editor: EditorConAdjuntos | undefined,
  ): void {
    const texto = editor?.textoPlano() ?? '';
    if (!texto) return;
    this.destinoDeRedaccion.set(destino);
    this.modalDeRedaccion()?.abrirCon(texto, destino === 'description' ? 'description' : 'comment');
  }

  protected aplicarMejora(texto: string): void {
    const html = textoComoHtml(texto);
    switch (this.destinoDeRedaccion()) {
      case 'description': this.descripcionEditada.set(html); break;
      case 'editedComment': this.textoEditado.set(html); break;
      default: this.comentario.set(html);
    }
  }

  /**
   * Si esta persona puede abrir **el modal de etiquetas**.
   *
   * No es `puedeEditar()`: **las etiquetas las ponen los dos equipos** (decisión 85), pero cada uno
   * desde su ticket —Soporte en el principal y Desarrollo en el interno—, y las etiquetas que se ven
   * son siempre las del principal, que el interno hereda (decisión 67). La regla vive en
   * `puedeEtiquetarElTicket`, para no repetirla.
   */
  protected puedeAbrirEtiquetas(): boolean {
    return puedeEtiquetarElTicket({
      soloLectura: this.soloLectura(),
      cerrado: this.cerrado(),
      interno: this.ticket().internal,
      esSoporte: this.esSoporte(),
      esDesarrollo: this.esDesarrollo(),
    });
  }

  protected puedeAsignar(): boolean {
    if (this.soloLectura() || this.cerrado()) {
      return false;
    }

    return this.ticket().internal ? this.esSoporte() || this.esDesarrollo() : this.esSoporte();
  }

  protected puedeTocarComentario(comentario: Comentario): boolean {
    return (
      !this.soloLectura() &&
      !this.cerrado() &&
      !comentario.deleted &&
      comentario.author?.id === this.sesion.usuario()?.id
    );
  }

  /**
   * Lo que hace falta para saber qué acciones se pueden hacer, **sin repetir las reglas**: las decide
   * `accionesDisponibles`, que es la misma que usa el desplegable. Aquí sólo se usa para el aviso de
   * «no hay nada que hacer», que se queda en la tarjeta de acciones.
   */
  private readonly contextoDeAcciones = computed<ContextoDeAcciones>(() => ({
    ticket: this.ticket(),
    esSoporte: this.esSoporte(),
    esDesarrollo: this.esDesarrollo(),
    esUsuario: this.esUsuario(),
    esSuyo: this.esSuyo(),
    soloLectura: this.soloLectura(),
  }));

  /**
   * Si queda algo que hacer con este ticket.
   *
   * Cuenta lo que sigue viviendo en la ficha —asignar, editar— y **las acciones de estado**, que ahora
   * están en el desplegable: sin contarlas, un ticket cerrado al que sólo le queda reabrir diría «no
   * hay nada que hacer» con la acción delante (decisión 78).
   */
  protected hayAcciones(): boolean {
    return (
      this.puedeAsignar() ||
      this.puedeEditar() ||
      accionesDisponibles(this.contextoDeAcciones()).length > 0
    );
  }

  // ------------------------------------------------------------------ la conversación

  protected estado(estado: string): string {
    return etiquetaDeEstado(this.t(), estado, this.esUsuario());
  }

  protected cuando(iso: string): string {
    return fechaHora(iso, this.idioma(), this.marca.zonaHoraria());
  }

  protected nombreDe(persona: { name: string; lastName: string } | undefined): string {
    return nombreDe(persona);
  }

  /** La frase de una entrada del historial: el texto entero vive en el diccionario. */
  protected frase(entrada: EntradaDeHistorial): string {
    const textos = this.t();

    if (!entrada.actor) {
      if (entrada.event === 'estado') {
        return interpolar(textos.historial.estadoSistema, {
          estado: this.estado(entrada.toState ?? ''),
        });
      }

      return textos.historial.creadoSistema;
    }

    const quien = this.nombreDe(entrada.actor);
    const frases: Record<string, string> = {
      creado: textos.historial.creado,
      editado: textos.historial.editado,
      asignado: textos.historial.asignado,
      estado: textos.historial.estado,
      escalado: textos.historial.escalado,
      resuelto: textos.historial.resuelto,
      cerrado: textos.historial.cerrado,
      reabierto: textos.historial.reabierto,
      observador: textos.historial.observador,
      observador_anadido: textos.historial.observador_anadido,
    };

    return interpolar(frases[entrada.event] ?? textos.historial.creado, {
      quien,
      detalle: entrada.detail ?? '',
      estado: this.estado(entrada.toState ?? ''),
    });
  }

  /** El texto de un comentario, ya pintado con formato y con sus adjuntos en su sitio. */
  protected cuerpoDe(comentario: Comentario): SafeHtml {
    return this.cuerpos().get(comentario.id) ?? '';
  }

  /**
   * Los adjuntos de un comentario que **su texto no nombra**, que son los que van en la lista del
   * final (decisión 49).
   */
  protected adjuntosDe(comentario: Comentario): readonly AdjuntoDeLaApi[] {
    const suyos = this.detalle().attachments.filter(
      (adjunto) => adjunto.commentId === comentario.id,
    );

    return this.adjuntosSueltos(suyos, comentario.body);
  }

  protected sePuedeVer(adjunto: AdjuntoDeTicket): boolean {
    return sePuedePrevisualizar(adjunto.filename);
  }

  /** Los textos del editor con formato del cuadro de escribir, que viven en el diccionario. */
  protected textosDelEditor(): TextosDelEditorConAdjuntos {
    const textos = this.t();

    return {
      barra: textos.editor.barra,
      negrita: textos.editor.negrita,
      cursiva: textos.editor.cursiva,
      subrayado: textos.editor.subrayado,
      tachado: textos.editor.tachado,
      listaVinietas: textos.editor.listaVinietas,
      listaNumerada: textos.editor.listaNumerada,
      enlace: textos.editor.enlace,
      adjuntar: textos.editor.adjuntar,
      etiquetar: textos.tickets.etiquetar,
      etiquetarA: textos.tickets.etiquetarA,
      sinPersonas: textos.tickets.sinPersonas,
      direccion: textos.editor.direccion,
      direccionPoner: textos.editor.direccionPoner,
      direccionCancelar: textos.editor.direccionCancelar,
      direccionInvalida: textos.editor.direccionInvalida,
      avisoDeRechazo: textos.tickets.archivoRechazado,
      adjuntoNoSubio: textos.tickets.adjuntoNoSubio,
    };
  }

  /** Los nombres de lo que falta, para poder decir cuál es. */
  protected nombresQueFaltan(): string {
    return this.adjuntosQueFaltan()
      .map((pendiente) => pendiente.archivo.name)
      .join(', ');
  }

  /**
   * Escribe un comentario, con sus adjuntos si los hay.
   *
   * **El comentario puede ir sin texto** si lleva archivos: «mira esto» con la captura es lo que pasa
   * todos los días, y obligar a escribir una frase para mandar una captura es ruido (decisión 41). Lo
   * que no se puede es no mandar nada: eso se dice antes de enviar.
   *
   * **Primero el comentario y luego sus archivos**, porque el adjunto necesita su `commentId`. Como el
   * texto referencia cada archivo **por su nombre**, no hay que volver a guardarlo después de subirlos.
   * Y si uno falla, el comentario **se queda** —está escrito y lo ha visto quien lo escriba—, la caja
   * se vacía y queda el aviso con el botón de reintentar (decisión 44). El hueco del que falta se ve
   * además **en el texto**, marcado, que es lo que la decisión 49 añadió.
   */
  protected async comentar(): Promise<void> {
    const editor = this.elDelComentario();

    if (!editor) {
      return;
    }

    const texto = editor.html();
    const archivos = editor.archivos();

    if (!editor.textoPlano() && !archivos.length) {
      this.mensaje.set({ forma: 'error', texto: this.t().tickets.comentarioVacio });
      return;
    }

    await this.actuar(async () => {
      const respuesta = await this.tickets.comentar(this.ticket().number, texto);
      const fallidos = await this.subirLosDelComentario(respuesta.comment.id, archivos);

      editor.limpiar();
      this.comentario.set('');
      this.adjuntosQueFaltan.set(fallidos);

      // Con archivos que no han subido **no se dice nada arriba**: lo dice el aviso que está junto a la
      // caja del comentario, con el nombre del archivo y su botón. Dos mensajes para lo mismo sobran.
      if (fallidos.length) {
        return null;
      }

      return { forma: 'exito', texto: this.t().tickets.comentarioEscrito };
    });
  }

  /**
   * Vuelve a subir lo que no subió, **al mismo comentario**.
   *
   * No se escribe nada de nuevo: el comentario ya está en la conversación, y lo único que falta es el
   * archivo. Cuando no queda ninguno, el aviso desaparece.
   */
  protected async reintentarAdjuntos(): Promise<void> {
    const pendientes = this.adjuntosQueFaltan();
    if (!pendientes.length) {
      return;
    }

    await this.actuar(async () => {
      const fallidos: { archivo: File; comentarioID: number | undefined }[] = [];

      for (const pendiente of pendientes) {
        const [restante] = await this.subirLosDelComentario(pendiente.comentarioID, [
          pendiente.archivo,
        ]);
        if (restante) {
          fallidos.push(restante);
        }
      }

      this.adjuntosQueFaltan.set(fallidos);

      if (fallidos.length) {
        return null;
      }

      return { forma: 'exito', texto: this.t().tickets.adjuntosSubidos };
    });
  }

  /**
   * Sube esos archivos a ese comentario y devuelve **los que no han subido**.
   *
   * Sin comentario, el archivo cuelga del ticket: es lo que pasa con los de la descripción.
   *
   * No se lanza el error hacia arriba a propósito: un archivo que no sube **no puede tumbar el
   * comentario**, que ya está publicado. Lo que se pierde, si acaso, es el archivo, y eso se cuenta.
   */
  private async subirLosDelComentario(
    comentarioID: number | undefined,
    archivos: readonly File[],
  ): Promise<{ archivo: File; comentarioID: number | undefined }[]> {
    const fallidos: { archivo: File; comentarioID: number | undefined }[] = [];

    for (const archivo of archivos) {
      try {
        await this.tickets.adjuntar(this.ticket().number, archivo, comentarioID);
      } catch {
        fallidos.push({ archivo, comentarioID });
      }
    }

    return fallidos;
  }

  /** El lápiz del asunto: lo convierte en un campo **en el sitio**, con lo que hay ahora puesto. */
  protected empezarEdicionDelAsunto(): void {
    this.mensaje.set(null);
    this.asuntoEditado.set(this.ticket().subject);
    this.editandoAsunto.set(true);
  }

  /**
   * Y el de la descripción, que se edita con el mismo editor con adjuntos de los comentarios, **donde
   * se lee** (decisión 74): el texto y sus archivos se cambian ahí, no en un diálogo.
   */
  protected empezarEdicionDeLaDescripcion(): void {
    this.mensaje.set(null);
    this.descripcionEditada.set(this.ticket().description);
    this.editandoDescripcion.set(true);
  }

  /**
   * Guarda el asunto. El campo se cierra antes de mandarlo —el asunto nuevo se lee en la cabecera en
   * cuanto llega la recarga—, y si el `PATCH` falla, el aviso de error sale arriba.
   */
  protected async guardarAsunto(): Promise<void> {
    const asunto = this.asuntoEditado().trim();

    if (!asunto) {
      this.mensaje.set({ forma: 'error', texto: this.t().tickets.obligatorios });
      return;
    }

    this.editandoAsunto.set(false);

    await this.actuar(async () => {
      await this.tickets.editar(this.ticket().number, { subject: asunto });
      return { forma: 'exito', texto: this.t().tickets.ticketActualizado };
    });
  }

  /**
   * Guarda la descripción, **con su formato**, y sube los archivos nuevos que lleve.
   *
   * La descripción se escribe en un editor con adjuntos, así que al guardarla puede traer archivos que
   * todavía no están subidos: se suben al ticket —sin comentario, que es la descripción— y el texto
   * sigue nombrando cada uno por su nombre. Deja la marca de que se editó, y eso lo pone el backend.
   */
  protected async guardarDescripcion(): Promise<void> {
    const editor = this.elDeLaDescripcion();
    const descripcion = editor?.html() ?? '';
    const archivos = editor?.archivos() ?? [];

    if (!editor || (!editor.textoPlano() && !archivos.length)) {
      this.mensaje.set({ forma: 'error', texto: this.t().tickets.obligatorios });
      return;
    }

    this.editandoDescripcion.set(false);

    await this.actuar(async () => {
      await this.tickets.editar(this.ticket().number, { description: descripcion });

      const fallidos = await this.subirLosDelComentario(undefined, archivos);
      if (fallidos.length) {
        return {
          forma: 'atencion',
          texto: `${this.t().tickets.adjuntosQueFaltan} ${this.nombresDe(fallidos.map((fallido) => fallido.archivo))}`,
        };
      }

      return { forma: 'exito', texto: this.t().tickets.descripcionActualizada };
    });
  }

  protected empezarEdicion(comentario: Comentario): void {
    this.editando.set(comentario.id);
    this.textoEditado.set(comentario.body);
  }

  /**
   * Guarda el texto de un comentario, **con su formato**, y sube los archivos que se hayan añadido al
   * editarlo.
   *
   * **Editar sigue exigiendo texto** (decisión 42): dejar en blanco un comentario que ya se escribió no
   * es lo mismo que nacer sin texto, y lo que se quiere en ese caso es borrarlo.
   */
  protected async guardarEdicion(comentario: Comentario): Promise<void> {
    const editor = this.elDeLaEdicion();

    if (!editor || !editor.textoPlano()) {
      this.mensaje.set({ forma: 'error', texto: this.t().tickets.comentarioVacio });
      return;
    }

    const texto = editor.html();
    const archivos = editor.archivos();

    await this.actuar(async () => {
      await this.tickets.editarComentario(this.ticket().number, comentario.id, texto);

      const fallidos = await this.subirLosDelComentario(comentario.id, archivos);
      if (fallidos.length) {
        // El comentario ya está guardado y lo que falta es el archivo: se dice cuál y se puede
        // reintentar, igual que cuando se publica uno nuevo (decisión 44).
        this.adjuntosQueFaltan.set(fallidos);
        this.editando.set(null);
        return null;
      }

      this.editando.set(null);
      return { forma: 'exito', texto: this.t().tickets.comentarioGuardado };
    });
  }

  // ------------------------------------------------------------------ las acciones

  /**
   * Abre la ventana de una acción de la conversación: hoy, sólo borrar un comentario.
   *
   * Las acciones del ticket ya no pasan por aquí —las lleva `app-acciones-del-ticket`—, pero borrar un
   * comentario es de un comentario y no del ticket, así que se queda con la conversación.
   */
  protected abrir(accion: Accion, comentarioID?: number): void {
    this.mensaje.set(null);
    this.pendiente.set({ accion, comentarioID });
  }

  protected cerrarDialogo(): void {
    this.pendiente.set(null);
  }

  /** El título, la explicación y el botón de la acción que está esperando. */
  protected textoDeLaAccionSegun(): { titulo: string; ayuda: string; boton: string } {
    const textos = this.t().tickets;

    switch (this.pendiente()?.accion) {
      case 'borrarComentario':
        return {
          titulo: textos.borrarPregunta,
          ayuda: textos.borrarConsecuencia,
          boton: textos.siBorrar,
        };
      default:
        return { titulo: '', ayuda: '', boton: '' };
    }
  }

  /** Hace lo que estaba esperando la ventana. */
  protected async confirmar(): Promise<void> {
    const pendiente = this.pendiente();
    if (!pendiente) {
      return;
    }

    const numero = this.ticket().number;
    this.cerrarDialogo();

    await this.actuar(async () => {
      // El texto del comentario se vacía de verdad y queda la marca; los adjuntos se quedan.
      if (pendiente.comentarioID) {
        await this.tickets.borrarComentario(numero, pendiente.comentarioID);
      }
      return { forma: 'exito', texto: this.t().tickets.comentarioBorradoAviso };
    });
  }

  /**
   * Un cambio hecho desde el desplegable de acciones: la pantalla vuelve a leer el ticket.
   *
   * El desplegable vive en la cabecera de la columna cuando se ven los dos, y el aviso de lo que pasó
   * también: por eso el aviso que enseña es el suyo y esto sólo pide la recarga.
   */
  protected avisarDelCambio(): void {
    this.cambio.emit();
  }

  /**
   * Abre el **modal de reasignación** con el responsable actual ya elegido: cambiarlo es un solo
   * paso, y quien no quiera cambiar nada cancela (quinta enmienda del 2026-09-29, punto 2).
   */
  protected abrirAsignacion(): void {
    this.mensaje.set(null);
    const responsable = this.ticket().assignee;
    this.elegido.set(responsable ? String(responsable.id) : '');
    this.mostrandoAsignacion.set(true);
  }

  /** Cierra el modal sin asignar. */
  protected cerrarAsignacion(): void {
    this.mostrandoAsignacion.set(false);
  }

  /** Asigna el ticket a quien se haya elegido en la lista. */
  protected async asignar(): Promise<void> {
    const persona = Number(this.elegido());
    if (!persona) {
      return;
    }

    const guardado = await this.actuar(async () => {
      await this.tickets.asignar(this.ticket().number, persona);
      this.elegido.set('');
      return { forma: 'exito', texto: this.t().tickets.asignado };
    });

    // **Sólo se cierra si salió bien**: si el backend lo rechaza, la ventana se queda abierta para
    // poder elegir a otra persona sin volver a abrirla.
    if (guardado) {
      this.mostrandoAsignacion.set(false);
    }
  }

  // ------------------------------------------------------------------ la clasificación

  /**
   * Las categorías que se pueden elegir en la ficha: las **activas**, más la que el ticket ya tiene
   * aunque esté retirada —si no, un ticket clasificado con una categoría retirada se quedaría sin poder
   * enseñar la suya—.
   */
  protected readonly opcionesDeCategoriaDeLaFicha = computed<readonly OpcionSelector[]>(() => {
    const actual = this.ticket().category;
    const categorias = this.categorias().filter(
      (categoria) => categoria.active || categoria.id === actual?.id,
    );

    return opcionesDeCategoria(this.t(), categorias);
  });

  /** Las etiquetas del ticket ahora mismo: es la lista que se enseña y la que se manda. */
  protected readonly etiquetas = computed<readonly string[]>(() => this.ticket().tags ?? []);

  /**
   * **La categoría se cambia en su propia línea y se guarda al elegir** (decisión 74): el desplegable es
   * la confirmación, así que no lleva botón de guardar.
   *
   * Se manda **sólo la categoría** —el backend deja las etiquetas como estaban cuando no vienen en el
   * `PATCH`—, y así un cambio de categoría no puede pisar las etiquetas de otro.
   */
  protected async cambiarCategoria(valor: string): Promise<void> {
    this.categoriaElegida.set(valor);

    const categoria = Number(valor);
    if (!categoria) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error('tickets.category.required') });
      return;
    }

    await this.actuar(async () => {
      await this.tickets.editar(this.ticket().number, { categoryId: categoria });
      return { forma: 'exito', texto: this.t().tickets.categoriaCambiada };
    });
  }

  /**
   * Abre el **modal de etiquetas** (decisión 82): todas las del catálogo, con su cuenta, un buscador y
   * casillas para marcar y desmarcar varias a la vez.
   */
  protected abrirEtiquetas(): void {
    this.mostrandoEtiquetas.set(true);
  }

  /**
   * Guarda **la lista entera de etiquetas marcadas en un solo `PATCH`**.
   *
   * El backend **reemplaza** las etiquetas del ticket con lo que le llegue (`docs/modules/tickets.md`,
   * sección 5), así que basta con mandar la lista que sale del modal. Si el guardado sale bien se cierra
   * la ventana; si falla, se queda abierta —no se pierde lo marcado— y el error se avisa en la ficha,
   * que es donde ya se avisa.
   */
  protected async guardarEtiquetas(etiquetas: readonly string[]): Promise<void> {
    const guardado = await this.actuar(async () => {
      await this.tickets.editar(this.ticket().number, { tags: etiquetas });
      return { forma: 'exito', texto: this.t().tickets.etiquetasCambiadas };
    });

    if (guardado) {
      this.mostrandoEtiquetas.set(false);
    }
  }

  // ------------------------------------------------------------------ los adjuntos

  /**
   * La dirección de un adjunto, **pidiéndola si no está**: el archivo se trae con `HttpClient` —la
   * descarga pasa por la comprobación de permisos y necesita la cabecera de la sesión— y se envuelve en
   * una dirección `blob:` del navegador. Se cachea por identificador y se libera al destruir la
   * pantalla (decisión 2.3).
   */
  private async direccionDe(adjunto: AdjuntoDeLaApi): Promise<string> {
    const guardada = this.direcciones().get(adjunto.id);
    if (guardada) {
      return guardada;
    }

    const datos = await this.tickets.descargar(this.ticket().number, adjunto.id);
    const url = URL.createObjectURL(datos);

    this.creadas.add(url);
    this.direcciones.update((direcciones) => new Map(direcciones).set(adjunto.id, url));

    return url;
  }

  /** Se trae los adjuntos que el texto nombra y todavía no tienen dirección, sin repetir los pedidos. */
  private pedirLosQueNombraElTexto(): void {
    const detalle = this.detalle();
    const nombres = new Set([
      ...this.nombresEnElTexto(detalle.ticket.description),
      ...detalle.comments.flatMap((comentario) => this.nombresEnElTexto(comentario.body)),
    ]);

    for (const adjunto of detalle.attachments) {
      if (
        !nombres.has(adjunto.filename) ||
        this.direcciones().has(adjunto.id) ||
        this.pidiendo.has(adjunto.id)
      ) {
        continue;
      }

      this.pidiendo.add(adjunto.id);

      void this.direccionDe(adjunto)
        .catch(() => undefined)
        .finally(() => this.pidiendo.delete(adjunto.id));
    }
  }

  /** Los nombres de archivo que un texto nombra con `data-adjunto`. */
  private nombresEnElTexto(html: string): readonly string[] {
    if (!html || !html.includes('data-adjunto')) {
      return [];
    }

    const contenedor = document.createElement('div');
    contenedor.innerHTML = html;

    return Array.from(contenedor.querySelectorAll('[data-adjunto]'))
      .map((elemento) => elemento.getAttribute('data-adjunto') ?? '')
      .filter((nombre) => nombre !== '');
  }

  /**
   * Los adjuntos de esa lista **que el texto no nombra**: los que siguen enseñándose en su lista al
   * final, con sus botones (decisión 49).
   */
  private adjuntosSueltos(
    adjuntos: readonly AdjuntoDeLaApi[],
    texto: string,
  ): readonly AdjuntoDeLaApi[] {
    const nombrados = new Set(this.nombresEnElTexto(texto));

    return adjuntos.filter((adjunto) => !nombrados.has(adjunto.filename));
  }

  /** El adjunto de ese ticket que se llama así. */
  private adjuntoPorNombre(nombre: string): AdjuntoDeLaApi | undefined {
    return this.detalle().attachments.find((adjunto) => adjunto.filename === nombre);
  }

  /**
   * Pinta el texto con formato y **con cada adjunto resuelto en su sitio**: la imagen y el vídeo en
   * línea, el PDF y lo demás como enlaces.
   *
   * Se monta nodo a nodo en vez de con un `[innerHTML]` a secas porque hay que **resolver** cada
   * `data-adjunto` con la lista de adjuntos del ticket que ya se ha leído, y porque lo que no se puede
   * resolver —un archivo que se nombró y no subió— se marca en su hueco (decisión 44).
   *
   * Los enlaces **no pueden ser `<a href>` normales**: la descarga necesita la sesión, así que el clic
   * se intercepta en el contenedor (`alPulsarElTexto`).
   */
  private pintar(html: string): SafeHtml {
    const contenedor = document.createElement('div');
    contenedor.innerHTML = html ?? '';

    for (const elemento of Array.from(contenedor.querySelectorAll<HTMLElement>('[data-adjunto]'))) {
      const nombre = elemento.getAttribute('data-adjunto') ?? '';
      const adjunto = this.adjuntoPorNombre(nombre);
      const url = adjunto ? this.direccionesPorNombre().get(nombre) : undefined;

      if (!adjunto) {
        elemento.replaceWith(this.marcaDeLoQueFalta(nombre));
        continue;
      }

      if (!url) {
        // Todavía no está: el archivo se está trayendo. Se dice, en vez de dejar un hueco mudo.
        elemento.replaceWith(this.textoDeEspera(nombre));
        continue;
      }

      if (elemento.tagName === 'IMG' || elemento.tagName === 'VIDEO') {
        const esVideo = elemento.tagName === 'VIDEO';

        elemento.setAttribute('src', esVideo ? `${url}#t=0.1` : url);
        elemento.setAttribute('class', CLASE_DE_LO_QUE_SE_VE);
        // **El nombre va también en el marco**, que es lo que se pulsa: el clic vale igual sobre el
        // archivo, sobre el hueco del vídeo y sobre el botón de reproducir.

        if (esVideo) {
          // **Sin controles**: el vídeo se ve como una miniatura y **se reproduce en el visor**, que
          // es donde se ve grande y se descarga (decisión 55). El primer fotograma se pide con el
          // fragmento `#t=0.1`, que es lo que hace que se pinte uno en vez de un rectángulo negro.
          elemento.setAttribute('preload', 'metadata');
        } else {
          elemento.setAttribute('alt', nombre);
        }

        // **Primero el marco en su sitio, y después el medio dentro**: al revés, mover el medio lo saca
        // del contenedor y el `replaceWith` se queda sin nada que sustituir (y el adjunto desaparece
        // del texto pintado).
        const marco = this.elMarcoDeLoQueSeVe(nombre, esVideo);
        elemento.replaceWith(marco);
        marco.appendChild(elemento);

        continue;
      }

      elemento.setAttribute('href', url);
      elemento.setAttribute('class', 'text-primario underline');
    }

    // **Las menciones se pintan como lo que son**: el nombre de una persona, con su fondo, para que al
    // leer el comentario se vea de un vistazo a quién se llamó (docs/modules/tickets.md, decisión 58).
    // El identificador no se enseña: está en el atributo, que es de donde sale la lista de
    // observadores, y a quien lee lo que le sirve es el nombre.
    for (const mencion of Array.from(contenedor.querySelectorAll<HTMLElement>('[data-mencion]'))) {
      mencion.setAttribute('class', CLASE_DE_LA_MENCION);
    }

    return this.saneador.bypassSecurityTrustHtml(contenedor.innerHTML);
  }

  /**
   * El marco de lo que se ve: **la imagen o la miniatura del vídeo dentro de un botón**, que es lo que
   * abre el visor.
   *
   * Es un `button` de verdad —y no un `div` con un `(click)`— porque así se llega con el tabulador y se
   * pulsa con Enter, y lleva su nombre accesible: una imagen que se pulsa y no dice que se pulsa es una
   * imagen que nadie descubre con un lector de pantalla (sección 8).
   */
  private elMarcoDeLoQueSeVe(nombre: string, esVideo: boolean): HTMLElement {
    const marco = document.createElement('button');
    marco.setAttribute('type', 'button');
    marco.setAttribute('data-adjunto', nombre);
    marco.setAttribute('class', esVideo ? CLASE_DEL_VIDEO : 'inline-flex max-w-full');
    marco.setAttribute(
      'aria-label',
      `${esVideo ? this.t().tickets.verVideo : this.t().tickets.verImagen}: ${nombre}`,
    );

    if (esVideo) {
      const boton = document.createElement('span');
      boton.setAttribute('class', CLASE_DEL_BOTON_DE_REPRODUCIR);
      boton.setAttribute('aria-hidden', 'true');
      boton.textContent = '▶';
      marco.appendChild(boton);
    }

    return marco;
  }

  /** El recuadro de un adjunto que el texto nombra y no está: no subió (decisión 44). */
  private marcaDeLoQueFalta(nombre: string): HTMLElement {
    const marca = document.createElement('span');

    marca.setAttribute('data-adjunto-que-falta', nombre);
    marca.className =
      'mx-1 inline-flex items-center gap-1 rounded-md border border-aviso bg-aviso-suave px-2 py-0.5 text-xs text-aviso';
    marca.textContent = `⚠ ${this.t().tickets.adjuntoNoSubio}: ${nombre || '—'}`;

    return marca;
  }

  /** Y lo que se enseña mientras el archivo se está trayendo. */
  private textoDeEspera(nombre: string): HTMLElement {
    const espera = document.createElement('span');

    espera.setAttribute('data-adjunto-trayendose', nombre);
    espera.className = 'mx-1 text-xs text-apagado';
    espera.textContent = this.t().tickets.trayendoAdjunto;

    return espera;
  }

  /**
   * El clic dentro de un texto con formato: **una imagen abre el visor**, un PDF también, y un Word o
   * un Excel **se descarga**.
   *
   * Se intercepta aquí y no se deja a un `<a href>`, porque la dirección del adjunto necesita la
   * cabecera de la sesión y el navegador no la manda al seguir un enlace.
   */
  protected alPulsarElTexto(evento: Event): void {
    const destino = (evento.target as HTMLElement | null)?.closest('[data-adjunto]');
    if (!destino) {
      return;
    }

    const adjunto = this.adjuntoPorNombre(destino.getAttribute('data-adjunto') ?? '');
    if (!adjunto) {
      return;
    }

    evento.preventDefault();

    // **La imagen, el vídeo y el PDF se abren en el visor** —grande, con Descargar y Abrir en una
    // pestaña—, y todo lo demás se descarga (decisiones 48 y 55). El vídeo entra aquí desde el
    // 2026-09-26: antes mandaba su reproductor y el visor sólo salía cuando el navegador no sabía
    // reproducirlo.
    const esVideo = destino.tagName === 'VIDEO' || Boolean(destino.querySelector('video'));

    if (destino.tagName === 'IMG' || esImagen(adjunto.filename) || this.esPdf(adjunto) || esVideo) {
      void this.traer(adjunto, true);
      return;
    }

    void this.traer(adjunto, false);
  }

  private esPdf(adjunto: AdjuntoDeTicket): boolean {
    return (
      adjunto.contentType === 'application/pdf' || adjunto.filename.toLowerCase().endsWith('.pdf')
    );
  }

  /**
   * Trae el archivo y lo guarda, o lo enseña en el visor.
   *
   * La dirección se guarda en la caché de la pantalla y **no se libera al cerrar el visor**: la misma
   * imagen puede estar pintada dentro del texto, y quitarle la dirección la dejaría rota. Se liberan
   * todas juntas al destruir la pantalla.
   */
  protected async traer(adjunto: AdjuntoDeTicket, paraVer: boolean): Promise<void> {
    try {
      const url = await this.direccionDe(adjunto as AdjuntoDeLaApi);

      if (paraVer) {
        this.vistaPrevia.set({
          adjunto: adjunto as AdjuntoDeLaApi,
          nombre: adjunto.filename,
          url,
          imagen: this.saneador.bypassSecurityTrustUrl(url),
          marco: this.saneador.bypassSecurityTrustResourceUrl(url),
          pdf: this.esPdf(adjunto),
          video: esVideo(adjunto.filename),
        });
        return;
      }

      const enlace = document.createElement('a');
      enlace.href = url;
      enlace.download = adjunto.filename;
      enlace.click();
    } catch {
      this.mensaje.set({ forma: 'error', texto: this.t().tickets.archivoNoTraido });
    }
  }

  /**
   * Abre el archivo del visor en una pestaña: es la forma de verlo a tamaño completo (decisión 48).
   *
   * Se hace con un enlace de verdad —`target="_blank"` y `rel="noopener"`—, y no con `window.open`: es
   * lo mismo que hace el navegador cuando alguien pulsa «abrir en una pestaña», y con `noopener` la
   * pestaña nueva no puede tocar la de la aplicación.
   */
  protected abrirEnPestana(): void {
    const vista = this.vistaPrevia();
    if (!vista) {
      return;
    }

    const enlace = document.createElement('a');
    enlace.href = vista.url;
    enlace.target = '_blank';
    enlace.rel = 'noopener';
    enlace.click();
  }

  /** Descarga lo que está enseñando el visor. */
  protected descargarDelVisor(): void {
    const vista = this.vistaPrevia();
    if (vista) {
      void this.traer(vista.adjunto, false);
    }
  }

  protected cerrarVistaPrevia(): void {
    this.vistaPrevia.set(null);
  }

  /** Las direcciones `blob:` que ha creado esta pantalla se liberan al salir (sección 2.3). */
  ngOnDestroy(): void {
    // Y la espera del motor se corta al salir: el `setTimeout` no debe sobrevivir a la pantalla.
    this.pararLaEsperaDelMotor();

    for (const url of this.creadas) {
      URL.revokeObjectURL(url);
    }

    this.creadas.clear();
  }

  /** Los nombres de unos archivos, para poder decir cuáles no han subido. */
  private nombresDe(archivos: readonly File[]): string {
    return archivos.map((archivo) => archivo.name).join(', ');
  }

  /** Hace la acción y deja la pantalla avisada de lo que pasó, bien o mal. */
  /**
   * Hace una acción y deja su mensaje en pantalla.
   *
   * El mensaje puede ser **nulo**: hay acciones que ya avisan donde ha pasado —el comentario publicado
   * con un archivo que no subió lo dice junto a su caja, con su botón de reintentar—, y repetirlo
   * arriba es ruido y dos sitios donde leer lo mismo.
   *
   * Devuelve **si salió bien**, que es lo que necesita el modal de etiquetas para cerrarse: si el guardado
   * falló, se queda abierto para no perder lo marcado.
   */
  private async actuar(accion: () => Promise<MensajeDePantalla | null>): Promise<boolean> {
    this.trabajando.set(true);
    this.mensaje.set(null);

    try {
      this.mensaje.set(await accion());
      this.cambio.emit();
      return true;
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
      return false;
    } finally {
      this.trabajando.set(false);
    }
  }
}
