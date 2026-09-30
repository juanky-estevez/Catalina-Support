import { Component, computed, inject, input, output, signal } from '@angular/core';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { Area } from '../../shared/components/area';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Dialogo } from '../../shared/components/dialogo';
import { Selector, type OpcionSelector } from '../../shared/components/selector';
import { claveDelError } from '../../shared/errores';
import {
  ESTADOS_DEL_INTERNO,
  ESTADOS_DEL_PRINCIPAL,
  etiquetaDeEstado,
  interpolar,
} from './etiquetas';
import { TicketsService, type Ticket } from './tickets.service';

/**
 * Las acciones que **mueven el ticket**.
 *
 * Siguen existiendo, pero **ya no se enseñan**: desde la corrección del responsable del 2026-09-29
 * (decisión 80) el desplegable presenta **estados**, no acciones, y estas sólo son la tabla de
 * permisos de la que salen los estados alcanzables. Así las reglas —qué puede cada papel y en cada
 * estado— viven en un solo sitio y no se duplican.
 *
 * No están las que no son de estado —asignar, comentar, editar—, que viven donde se usan.
 */
export type AccionDeEstado =
  | 'empezar'
  | 'preguntarAlUsuario'
  | 'pedirAlgo'
  | 'escalar'
  | 'devolver'
  | 'resolver'
  | 'cerrar'
  | 'reabrir';

/** Lo que hace falta saber del ticket y de quien mira para decidir **qué se puede hacer** ahora. */
export interface ContextoDeAcciones {
  readonly ticket: Ticket;
  readonly esSoporte: boolean;
  readonly esDesarrollo: boolean;
  readonly esUsuario: boolean;
  readonly esSuyo: boolean;
  readonly soloLectura: boolean;
}

/**
 * Las acciones que **se pueden hacer ahora mismo**, en el orden en que se ofrecen.
 *
 * Son exactamente las reglas que ya tenían los botones sueltos —`puedeEmpezar()`,
 * `puedePreguntarAlUsuario()`, `puedeEscalar()`, `puedeResolver()`, `puedeCerrar()`,
 * `puedeReabrir()`…— y **ninguna más**. La interfaz no ofrece lo que el backend rechazaría
 * (`docs/interfaz-y-experiencia.md`, sección 4.3).
 *
 * Vive aquí, y no en la vista, para que **los dos sitios donde aparece el control** —la fila de
 * arriba con un solo ticket y la cabecera de cada columna con los dos— usen las mismas reglas sin
 * duplicarlas (decisión 78).
 */
export function accionesDisponibles(contexto: ContextoDeAcciones): readonly AccionDeEstado[] {
  const { ticket, esSoporte, esDesarrollo, esUsuario, esSuyo, soloLectura } = contexto;
  const estado = ticket.state;
  const interno = ticket.internal;
  const cerrado = estado === 'cerrado';
  const enMarcha = estado === 'en progreso' || estado === 'en espera';
  const acciones: AccionDeEstado[] = [];

  // Empezar a trabajarlo: `nuevo → en progreso`, y en el interno también la respuesta de Soporte.
  if (
    interno
      ? (esDesarrollo && estado === 'nuevo') || (esSoporte && estado === 'en espera')
      : esSoporte && estado === 'nuevo'
  ) {
    acciones.push('empezar');
  }

  // Pedir algo: al usuario, desde Soporte; a Soporte, desde Desarrollo.
  if (!interno && esSoporte && estado === 'en progreso') {
    acciones.push('preguntarAlUsuario');
  }
  if (interno && esDesarrollo && estado === 'en progreso') {
    acciones.push('pedirAlgo');
  }

  // Escalar a Desarrollo, y su vuelta —«no es un cambio de código»—, que sólo existen en un lado.
  if (!interno && esSoporte && !cerrado) {
    acciones.push('escalar');
  }
  if (interno && (esDesarrollo || esSoporte) && !cerrado && estado !== 'resuelto') {
    acciones.push('devolver');
  }

  if (interno ? esDesarrollo && enMarcha : esSoporte && enMarcha) {
    acciones.push('resolver');
  }

  // Cerrar sin resolver se puede desde cualquier estado abierto, y también desde `resuelto`: lo que
  // no se puede es cerrar un ticket que está en manos de Desarrollo.
  const puedeCerrar =
    !cerrado &&
    !soloLectura &&
    (interno
      ? esDesarrollo || esSoporte
      : estado !== 'escalado' && (esSoporte || (esUsuario && esSuyo)));
  if (puedeCerrar) {
    acciones.push('cerrar');
  }

  if (cerrado && !interno && (esSoporte || (esUsuario && esSuyo))) {
    acciones.push('reabrir');
  }

  return acciones;
}

/**
 * **A qué estado lleva cada acción**: el mapeo que fija la decisión 80, y que existe para que
 * traducir acciones a estados no pierda nada.
 *
 * «Empezar» es `en progreso`; «Preguntar al usuario» y «Necesito algo de Soporte» son los dos `en
 * espera` —cada uno en su tipo de ticket—; «Escalar» es `escalado`; «Resolver» es `resuelto`;
 * «Cerrar» es `cerrado`; «Reabrir» es `en progreso` desde `cerrado`; y «No es un cambio de código»
 * es `cerrado` en un interno, que es la devolución a Soporte (regla 8 de la sección 3.3).
 */
const ESTADO_DE_LA_ACCION: Readonly<Record<AccionDeEstado, string>> = {
  empezar: 'en progreso',
  preguntarAlUsuario: 'en espera',
  pedirAlgo: 'en espera',
  escalar: 'escalado',
  devolver: 'cerrado',
  resolver: 'resuelto',
  cerrar: 'cerrado',
  reabrir: 'en progreso',
};

/**
 * **Los estados a los que el ticket puede pasar ahora**, que es lo que enseña el desplegable
 * (decisión 80).
 *
 * No se recalcula ninguna regla: se traducen las acciones disponibles con el mapeo de arriba y se
 * quitan los repetidos —«Devolver» y «Cerrar» llevan los dos a `cerrado` en un interno—. El orden es
 * el canónico del tipo de ticket, para que la lista se lea igual siempre y no dependa del orden en
 * que se hayan juntado las acciones.
 */
export function estadosDisponibles(contexto: ContextoDeAcciones): readonly string[] {
  const alcanzables = new Set(
    accionesDisponibles(contexto).map((accion) => ESTADO_DE_LA_ACCION[accion]),
  );
  const orden = contexto.ticket.internal ? ESTADOS_DEL_INTERNO : ESTADOS_DEL_PRINCIPAL;

  return orden.filter((estado) => alcanzables.has(estado));
}

/** Lo que dice la ventana de confirmación de un estado: a cuál se pasa, por qué y qué pide. */
interface VentanaDeEstado {
  /** El título: «¿Pasar el ticket a «…»?», con el nombre del estado interpolado. */
  readonly titulo: string;
  /** La explicación de lo que va a pasar, con las ayudas que ya existían. */
  readonly explicacion: string;
  /** Una segunda explicación, cuando el cambio arrastra algo más —la devolución a Soporte—. */
  readonly consecuencia: string;
  /** El texto que pide, si lo pide: el motivo, la pregunta, el comentario del cierre… */
  readonly campo: string;
  /** Si ese texto es obligatorio para poder confirmar. */
  readonly obligatorio: boolean;
}

/**
 * El control para **cambiar de estado el ticket**, en un solo desplegable.
 *
 * La corrección del responsable del 2026-09-29 (decisión 80) cambió lo que enseña: **no es una lista
 * de acciones, es la lista de estados** a los que el ticket puede pasar ahora, con **el actual como
 * primera opción**, marcado y sin poder elegirlo. Vale igual para los principales y para los internos,
 * y cada elección **se confirma en una sola ventana**, que dice a cuál se pasa y, si esa transición
 * necesita un texto, lo pide ahí mismo —el motivo al escalar, qué se ha hecho al resolver, la pregunta
 * o lo que se necesita de Soporte al ponerse en espera y el comentario al cerrar, que es obligatorio
 * (decisión 81)—.
 *
 * El componente es el mismo en los dos sitios donde vive —la fila de arriba de la página cuando se ve
 * un solo ticket y la cabecera de cada columna cuando se ven los dos—, porque recibe el ticket y
 * **dentro** tiene el desplegable y su ventana. Así no hay dos copias de la misma lógica ni dos
 * maneras de mover un ticket. Y **al elegir, el desplegable vuelve a su sitio**: es un control de
 * mando, no un campo con valor.
 */
@Component({
  selector: 'app-acciones-del-ticket',
  imports: [Area, Aviso, Boton, Dialogo, Selector],
  templateUrl: './acciones-del-ticket.html',
})
export class AccionesDelTicket {
  /** El ticket al que se le cambia el estado. Basta con el ticket: el número y el estado vienen aquí. */
  readonly ticket = input.required<Ticket>();

  /** Lo que se acaba de hacer: quien lo puso vuelve a leer el ticket. */
  readonly cambio = output<void>();

  private readonly tickets = inject(TicketsService);
  private readonly textos = inject(TranslationService);
  private readonly sesion = inject(SessionService);

  protected readonly trabajando = signal(false);

  /** El aviso del resultado, que se pinta **junto al control**, no dentro de la ventana. */
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);

  /**
   * El aviso de que falta un texto obligatorio, que se pinta **dentro de la ventana**.
   *
   * Va aparte de `mensaje` porque el ✕ y «Cancelar» de la ventana cierran y **no pueden borrar el
   * resultado** de un cambio que ya se hizo: al cerrarse la ventana después de confirmar, el aviso de
   * éxito tiene que seguir viéndose junto al control.
   */
  protected readonly errorDeLaVentana = signal<string | null>(null);

  /** El estado que está esperando su ventana de confirmación. `null` es que no hay ninguna abierta. */
  protected readonly pendiente = signal<string | null>(null);
  protected readonly texto = signal('');

  /**
   * Cuántas veces se ha vuelto a la opción del estado actual.
   *
   * No es un dato de negocio: es lo que hace que **el desplegable vuelva a su sitio** después de cada
   * elección. El `[selected]` de las opciones sólo se reescribe cuando la expresión cambia, así que con
   * el mismo valor el control se quedaría enseñando el último estado elegido; cambiando el valor de la
   * opción actual, Angular lo devuelve a ella. Y se cambia el valor, no se recrea el control, para no
   * perder el foco.
   */
  private readonly generacion = signal(0);

  protected t() {
    return this.textos.textos();
  }

  protected readonly esUsuario = computed(() => this.sesion.usuario()?.role === 'usuario');
  protected readonly esSoporte = computed(() => this.sesion.usuario()?.role === 'soporte');
  protected readonly esDesarrollo = computed(() => this.sesion.usuario()?.role === 'desarrollo');
  /** El Administrador mira los tickets: no mueve estados (sección 4.4). */
  protected readonly soloLectura = this.sesion.esAdministrador;
  protected readonly esSuyo = computed(
    () => this.ticket().requester?.id === this.sesion.usuario()?.id,
  );

  /** Lo que decide **qué se puede hacer**, sin repetir ninguna regla en la vista. */
  private readonly contexto = computed<ContextoDeAcciones>(() => ({
    ticket: this.ticket(),
    esSoporte: this.esSoporte(),
    esDesarrollo: this.esDesarrollo(),
    esUsuario: this.esUsuario(),
    esSuyo: this.esSuyo(),
    soloLectura: this.soloLectura(),
  }));

  /** **Sólo los estados alcanzables**: ni los que ya pasaron ni los que este papel no puede mover. */
  protected readonly estados = computed<readonly string[]>(() =>
    estadosDisponibles(this.contexto()),
  );

  /**
   * El valor de **la opción del estado actual**, que es la primera y no se puede elegir.
   *
   * Lleva la cuenta de la generación para que **cambie en cada elección**: es lo que obliga al
   * desplegable a volver a su sitio en vez de quedarse con el último estado elegido (ver `generacion`).
   * Elegirlo no hace nada: no es un destino, es dónde está el ticket.
   */
  protected readonly valorDelEstadoActual = computed(() => `actual:${this.generacion()}`);

  /**
   * Las opciones del desplegable: **el estado actual primero**, marcado y sin poder elegirlo, y
   * detrás **sólo los estados alcanzables**. Si no hay ninguno, el componente no enseña nada.
   */
  protected readonly opciones = computed<readonly OpcionSelector[]>(() => {
    const grupo = this.etiquetaDelControl();

    return [
      {
        valor: this.valorDelEstadoActual(),
        etiqueta: this.etiquetaDe(this.ticket().state),
        grupo,
        deshabilitado: true,
      },
      ...this.estados().map((estado) => ({
        valor: estado,
        etiqueta: this.etiquetaDe(estado),
        grupo,
      })),
    ];
  });

  /**
   * El nombre accesible del desplegable.
   *
   * **La etiqueta accesible** es `tickets.estadoDelTicket` («Estado del ticket»): el control cambia el
  estado, y su nombre tiene que decirlo (`accionesDelTicket` era de cuando elegía acciones).
   * que era el nombre del control de acciones; para un control de estados lo natural sería algo como
   * «Estado del ticket». Mientras el responsable confirma la clave nueva, se reutiliza ésta, que es la
   * que ya estaba puesta (`AGENTS.md`: los textos no se escriben dentro del componente).
   */
  protected etiquetaDelControl(): string {
    return this.t().tickets.estadoDelTicket;
  }

  /** El nombre de un estado, con el mismo lenguaje que el resto de la pantalla y del papel. */
  protected etiquetaDe(estado: string): string {
    return etiquetaDeEstado(this.t(), estado, this.esUsuario());
  }

  /**
   * Al elegir un estado: **el desplegable vuelve a su sitio** y se abre **una sola ventana** —la que
   * confirma el cambio y, si hace falta, pide el texto—. Todo pasa por aquí, también `en progreso`:
   * ya no hay cambios que se hagan sin confirmar.
   */
  protected elegir(valor: string): void {
    this.mensaje.set(null);
    this.errorDeLaVentana.set(null);

    // El estado actual no es un destino: si se elige, no pasa nada.
    if (!valor || valor.startsWith('actual:')) {
      return;
    }

    // **Vuelve a su sitio**: es un control de mando, no un campo con valor. Se cambia el valor de la
    // opción actual para que Angular lo devuelva a ella.
    this.generacion.update((cuantas) => cuantas + 1);

    this.texto.set('');
    this.pendiente.set(valor);
  }

  protected cerrarDialogo(): void {
    this.pendiente.set(null);
    this.texto.set('');
    // Sólo se borra lo de la ventana: el resultado de un cambio ya hecho se queda donde se lee.
    this.errorDeLaVentana.set(null);
  }

  /**
   * Lo que dice la ventana del estado que está esperando: a cuál se pasa, con qué ayuda se explica y
   * qué texto pide.
   *
   * **Los textos viven en el diccionario** y se reutilizan los que ya había para cada acción
   * (`AGENTS.md`). El porqué de cada uno:
   *
   * · `escalado` pide el motivo, con la ayuda de escalar.
   * · `resuelto` pide qué se ha hecho, con la ayuda de resolver.
   * · `en espera` en un principal es **preguntarle al usuario**: pide la pregunta.
   * · `en espera` en un interno es **devolver el caso a Soporte sin cerrarlo** (regla 2): pide qué se
   *   necesita y lo explica la ayuda de devolver.
   * · `cerrado` pide **siempre** el comentario (decisión 81), y en un interno que no estaba resuelto
   *   además **devuelve el caso a Soporte** (regla 8): por eso lleva la segunda explicación.
   * · `en progreso` desde `cerrado` es **reabrirlo**; desde `nuevo` o `en espera` sólo confirma.
   */
  protected readonly ventana = computed<VentanaDeEstado>(() => {
    const estado = this.pendiente();
    const textos = this.t().tickets;
    const desde = this.ticket().state;
    const interno = this.ticket().internal;

    if (!estado) {
      return { titulo: '', explicacion: '', consecuencia: '', campo: '', obligatorio: false };
    }

    const titulo = interpolar(textos.pasarA, { estado: this.etiquetaDe(estado) });

    switch (estado) {
      case 'escalado':
        return {
          titulo,
          explicacion: textos.escalarAyuda,
          consecuencia: '',
          campo: textos.motivo,
          obligatorio: true,
        };

      case 'resuelto':
        return {
          titulo,
          explicacion: textos.resolverAyuda,
          consecuencia: '',
          campo: textos.queSeHizo,
          obligatorio: true,
        };

      case 'en espera':
        // **«En espera» sólo se confirma** (decisión del responsable, 2026-09-29, decisión 88): antes
        // pedía el texto —la pregunta al usuario, o qué necesitas de Soporte— y ahora no pide nada. La
        // ventana sigue **diciendo lo que va a pasar**, que es lo que hace falta para decidir.
        // **El único estado que pide texto es «Cerrado»** (el comentario del cierre, decisión 81).
        return interno
          ? {
              titulo,
              explicacion: textos.devolverConsecuencia,
              consecuencia: '',
              campo: '',
              obligatorio: false,
            }
          : {
              titulo,
              explicacion: textos.preguntarConsecuencia,
              consecuencia: '',
              campo: '',
              obligatorio: false,
            };

      case 'cerrado':
        return {
          titulo,
          // Cerrar siempre pide el porqué, y queda en la conversación (decisión 81).
          explicacion: textos.cierreAyuda,
          // Cerrar un interno que no estaba resuelto es la devolución a Soporte (regla 8): se dice,
          // porque el cambio no es sólo cerrar.
          consecuencia: interno && desde !== 'resuelto' ? textos.devolverConsecuencia : '',
          campo: textos.comentarioDelCierre,
          obligatorio: true,
        };

      case 'en progreso':
        return {
          titulo,
          // La reapertura y el empezar explican lo suyo: los dos dejan el ticket en progreso, y no
          // significan lo mismo.
          explicacion: desde === 'cerrado' ? textos.reabrirAyuda : textos.empezarConsecuencia,
          consecuencia: '',
          campo: '',
          obligatorio: false,
        };

      default:
        return { titulo, explicacion: '', consecuencia: '', campo: '', obligatorio: false };
    }
  });

  /**
   * Hace el cambio que estaba esperando la ventana.
   *
   * **Nada se manda sin el texto obligatorio**: al cerrar, sin comentario, se enseña el mismo aviso
   * que traduce el `422` del backend (`tickets.cierre.sinComentario`), y el cambio no sale. Es la
   * misma regla que la categoría obligatoria: lo que no puede faltar se comprueba donde se escribe.
   */
  protected async confirmar(): Promise<void> {
    const estado = this.pendiente();
    if (!estado) {
      return;
    }

    const numero = this.ticket().number;
    const texto = this.texto().trim();

    if (this.ventana().obligatorio && !texto) {
      this.errorDeLaVentana.set(
        estado === 'cerrado'
          ? this.textos.error('tickets.cierre.sinComentario')
          : this.t().tickets.comentarioVacio,
      );
      return;
    }

    this.cerrarDialogo();

    await this.actuar(async () => {
      switch (estado) {
        // Escalar tiene su propia puerta en el backend —crea o reabre el interno y exige motivo—, así
        // que no va por el cambio de estado, aunque sea el estado `escalado` (decisión 80).
        case 'escalado':
          await this.tickets.escalar(numero, texto);
          return { forma: 'exito', texto: this.t().tickets.escaladoAviso };

        // Reabrir un principal cerrado también tiene su acción propia en el backend: `cerrado → en
        // progreso` no se hace por el cambio de estado.
        case 'en progreso':
          if (this.ticket().state === 'cerrado') {
            await this.tickets.reabrir(numero);
            return { forma: 'exito', texto: this.t().tickets.reabiertoAviso };
          }

          await this.tickets.moverEstado(numero, { to: 'en progreso' });
          return { forma: 'exito', texto: this.t().tickets.estadoCambiado };

        // El texto va **en el mismo cambio de estado**, y no en un comentario aparte: es lo que pide
        // el endpoint y lo que hace que cerrar guarde el porqué como un comentario de la conversación.
        // Al cerrar va **siempre**; en los demás, sólo si se ha escrito.
        case 'resuelto':
          await this.tickets.moverEstado(numero, this.cambioA(estado, texto));
          return { forma: 'exito', texto: this.t().tickets.resueltoAviso };

        case 'en espera':
          await this.tickets.moverEstado(numero, this.cambioA(estado, texto));
          return { forma: 'exito', texto: this.t().tickets.estadoCambiado };

        case 'cerrado':
          await this.tickets.moverEstado(numero, this.cambioA(estado, texto));
          return { forma: 'exito', texto: this.t().tickets.cerradoAviso };

        default:
          return null;
      }
    });
  }

  /**
   * El cuerpo del cambio de estado: `{ to }` y, cuando hay algo que decir, `{ to, comment }`.
   *
   * El `comment` va también en los estados que no lo piden si se ha escrito —hoy no puede pasar,
   * porque esos no enseñan campo, pero es la regla— y **al cerrar va siempre**, que es lo que exige
   * el backend (decisión 81).
   */
  private cambioA(estado: string, texto: string): { to: string; comment?: string } {
    if (texto || estado === 'cerrado') {
      return { to: estado, comment: texto };
    }

    return { to: estado };
  }

  /** Hace el cambio y deja la pantalla avisada de lo que pasó, bien o mal. */
  private async actuar(accion: () => Promise<MensajeDePantalla | null>): Promise<void> {
    this.trabajando.set(true);
    this.mensaje.set(null);

    try {
      this.mensaje.set(await accion());
      this.cambio.emit();
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.trabajando.set(false);
    }
  }
}
