import { Component, type OnDestroy, OnInit, computed, inject, input, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { BrandService } from '../../core/services/brand.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Conmutador, type OpcionConmutador } from '../../shared/components/conmutador';
import { Dialogo } from '../../shared/components/dialogo';
import { Estado } from '../../shared/components/estado';
import { claveDelError } from '../../shared/errores';
import { consultaDeAncho, esPantallaAncha } from '../../shared/pantalla';
import {
  etiquetaDeEstado,
  fechaHora,
  interpolar,
  nombreDe,
  normalizarEtiqueta,
  opcionesDeEstado,
  textoDelResumen,
} from './etiquetas';
import {
  FILTROS_VACIOS,
  TIPO_INTERNO,
  TIPO_PRINCIPAL,
  TicketsService,
  type Categoria,
  type FiltrosDeBandeja,
  type PaginaDeTickets,
  type Ticket,
} from './tickets.service';

/** A partir de aquí la bandeja se pinta con la tabla: por debajo, con tarjetas (sección 7). */
const ANCHO_DE_TABLA = 768;

/**
 * Cuánto se espera entre relecturas y cuántas se hacen como mucho mientras el motor siga pendiente.
 *
 * El motor tarda, medido, **entre 12 y 24 segundos por campo**, y más si tiene cola (`docs/modules/ai.md`):
 * con ocho intentos de ocho segundos se le da algo más de un minuto, el mismo margen que usa la ficha.
 * **No es un sondeo eterno**: al llegar al tope se para (`docs/interfaz-y-experiencia.md`, enmienda del
 * 2026-09-29).
 */
const ESPERA_DEL_MOTOR = 8000;
const INTENTOS_DEL_MOTOR = 8;

/** Lo que se espera antes de buscar: escribir un número entero es una búsqueda, no seis. */

/**
 * Las listas de tickets: **una sola pantalla con cuatro usos**.
 *
 * Es la misma tabla con distinto contenido, como manda `docs/modules/tickets.md`, sección 8:
 *
 * - **`mios`** (`/tickets`): **lo mío** —lo asignado, lo abierto por mí y aquello donde he
 *   comentado—, que es la bandeja del usuario, de Soporte y de Desarrollo. Soporte y Desarrollo
 *   llevan **chip de tipo** —cada uno entra por el suyo (decisión 52)—, y el usuario no: sus tickets
 *   son principales.
 * - **`principal`** (`/tickets/main`) y **`interno`** (`/tickets/internal`): **todo lo que hay**, para
 *   Soporte y Desarrollo, con el tipo fijo por la pantalla misma.
 * - **El Administrador** usa `/tickets` como siempre: los dos tipos, con su chip y **de sólo
 *   lectura** (`docs/interfaz-y-experiencia.md`, sección 4.4).
 *
 * Lo que lleva más tiempo esperando va primero, y lo cerrado no estorba: el orden lo decide el
 * backend, que es quien tiene los datos (sección 3.3).
 */
@Component({
  selector: 'app-bandeja-page',
  imports: [RouterLink, Aviso, Boton, Campo, Conmutador, Dialogo, Estado],
  templateUrl: './bandeja-page.html',
})
export class BandejaPage implements OnInit, OnDestroy {
  private readonly tickets = inject(TicketsService);
  private readonly textos = inject(TranslationService);
  private readonly marca = inject(BrandService);
  private readonly sesion = inject(SessionService);
  private readonly router = inject(Router);

  protected readonly pagina = signal<PaginaDeTickets>({
    tickets: [],
    total: 0,
    page: 1,
    perPage: 0,
  });

  protected readonly filtros = signal<FiltrosDeBandeja>(FILTROS_VACIOS);

  /**
   * Si el modal de búsqueda está abierto.
   *
   * **Los criterios viven en un modal** (decisión del responsable, 2026-09-27): eran tres filas de
   * filtros para cuatro criterios —y la de categorías se salía de la pantalla cuando el catálogo
   * crecía—, y así la lista se ve limpia y los criterios caben todos, los de hoy y los de mañana.
   */
  protected readonly buscando = signal(false);

  /**
   * Lo que se está escribiendo en el modal, **antes de aplicarlo**.
   *
   * Se edita aquí y no en los filtros de verdad: dentro de un modal, filtrar en vivo es confuso —la
   * lista de detrás iría cambiando mientras se rellenan cinco campos—, así que lo aplica el botón
   * «Buscar». Al abrir el modal, el borrador arranca con lo que ya está filtrado.
   */
  protected readonly borrador = signal<FiltrosDeBandeja>(FILTROS_VACIOS);

  /** El catálogo, para el chip de categoría. Lo ven todos los que han entrado. */
  protected readonly categorias = signal<readonly Categoria[]>([]);
  protected readonly cargando = signal(true);
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);

  /** Las relecturas que lleva hechas la vigilancia del motor y la que tiene programada, si hay alguna. */
  private intentosDelMotor = 0;
  private temporizadorDelMotor: ReturnType<typeof setTimeout> | null = null;

  /**
   * Cuál de las cuatro listas es. Lo trae la ruta (`data.listado`), así que la pantalla es la misma
   * para todas y lo que cambia es esto.
   */
  readonly listado = input<'mios' | 'principal' | 'interno'>('mios');

  /** Tabla en PC y tarjetas en móvil: se elige con una consulta para que no existan las dos a la vez. */
  protected readonly esAncha = signal(esPantallaAncha(ANCHO_DE_TABLA));

  protected readonly papel = computed(() => this.sesion.usuario()?.role ?? '');
  protected readonly esUsuario = computed(() => this.papel() === 'usuario');
  protected readonly esSoporte = computed(() => this.papel() === 'soporte');
  protected readonly esDesarrollo = computed(() => this.papel() === 'desarrollo');
  protected readonly esAdministrador = this.sesion.esAdministrador;

  /**
   * Quién puede crear un ticket desde aquí: el usuario, Soporte y Desarrollo, **y sólo en la
   * bandeja** —«Mis tickets»—, que es donde vive el botón desde que salió del menú
   * (decisión 53). Las dos listas del «todo» son para mirar y trabajar lo que hay, no para crear.
   */
  protected readonly puedeCrear = computed(
    () =>
      !this.esListaDelTodo() &&
      !this.esAdministrador() &&
      (this.esUsuario() || this.esSoporte() || this.esDesarrollo()),
  );

  /** Si la bandeja enseña internos, que es lo que decide los chips de estado y el nombre de las cosas. */
  protected readonly mirandoInternos = computed(() => this.filtros().type === TIPO_INTERNO);

  protected readonly paginas = computed(() =>
    Math.max(1, Math.ceil(this.pagina().total / (this.pagina().perPage || 1))),
  );

  /** Si esta lista lleva el chip de tipo: «Mis tickets» de Soporte y Desarrollo, y la del Administrador. */
  /**
   * El chip de vista lo lleva **«Mis tickets» para Soporte y Desarrollo**: es donde se mezclan los
   * asignados, los que uno abrió, donde ha comentado y los que observa (decisión 62). El usuario no lo
   * necesita —todo lo suyo es suyo— y el Administrador mira la bandeja entera.
   */
  protected readonly conChipDeVista = computed(
    () => this.listado() === 'mios' && (this.esSoporte() || this.esDesarrollo()),
  );

  /**
   * Con qué vista arranca «Mis tickets»: **Asignados** para Soporte y Desarrollo —lo que hay que
   * atender— y sin vista para los demás, que no tienen el chip.
   */
  private vistaInicial(): { view?: string } {
    return this.conChipDeVista() ? { view: 'assigned' } : {};
  }

  /** Las tres vistas, con **Asignados** primero, que es lo que hay que atender. */
  protected opcionesDeVista(): readonly OpcionConmutador[] {
    return [
      { valor: 'assigned', etiqueta: this.t().tickets.vistaAsignados },
      { valor: 'watching', etiqueta: this.t().tickets.vistaObservo },
      { valor: '', etiqueta: this.t().tickets.vistaTodos },
    ];
  }

  protected readonly conChipDeTipo = computed(
    () =>
      this.esAdministrador() ||
      (this.listado() === 'mios' && (this.esSoporte() || this.esDesarrollo())),
  );

  /** Si la lista enseña **todo** (las dos del «todo»), que es lo que decide el botón de crear. */
  protected readonly esListaDelTodo = computed(
    () => this.listado() !== 'mios' && !this.esAdministrador(),
  );

  constructor() {
    consultaDeAncho(ANCHO_DE_TABLA)?.addEventListener('change', (evento) =>
      this.esAncha.set(evento.matches),
    );
  }

  /**
   * La lista arranca aquí y no en el constructor: **cuál de las cuatro es lo trae la ruta**, y el
   * enrutador pone ese dato justo antes de `ngOnInit`. En el constructor la lista sería siempre «lo
   * mío», y una de las dos del «todo» pediría el tipo equivocado.
   */
  ngOnInit(): void {
    // El tipo de ticket: el que fija la lista cuando es una del «todo», y si no el de siempre —el
    // usuario y Soporte, los principales; Desarrollo, los internos; el Administrador, los principales,
    // que puede cambiar con su chip—.
    this.filtros.set({ ...FILTROS_VACIOS, ...this.vistaInicial(), type: this.tipoInicial() });
    void this.cargar();
    void this.cargarCategorias();
  }

  protected t() {
    return this.textos.textos();
  }

  /** El título de la pantalla, que es como se llama esta lista. */
  protected titulo(): string {
    if (this.esAdministrador()) {
      return this.t().menu.bandeja;
    }
    if (this.listado() === 'principal') {
      return this.t().menu.ticketsPrincipales;
    }
    if (this.listado() === 'interno') {
      return this.t().menu.ticketsInternos;
    }

    return this.t().menu.misTickets;
  }

  /** Los chips del estado: los del tipo de ticket que se está mirando. */
  protected opcionesDeEstado(): readonly OpcionConmutador[] {
    return opcionesDeEstado(this.t(), this.mirandoInternos());
  }

  /** El chip de tipo: lo lleva quien puede tener de los dos —Soporte, Desarrollo y el Administrador—. */
  protected opcionesDeTipo(): readonly OpcionConmutador[] {
    return [
      { valor: TIPO_PRINCIPAL, etiqueta: this.t().tickets.tipoPrincipal },
      { valor: TIPO_INTERNO, etiqueta: this.t().tickets.tipoInterno },
      { valor: '', etiqueta: this.t().tickets.tipoTodos },
    ];
  }

  /**
   * El chip de categoría: todas y las del catálogo.
   *
   * Se enseña **tal y como lo devuelve el backend** —activas para todos, y las retiradas para quien
   * mantiene el catálogo—: una categoría retirada sigue clasificando tickets, así que poder filtrar por
   * ella es lo que hace falta para encontrarlos.
   */
  protected opcionesDeCategoria(): readonly OpcionConmutador[] {
    return [
      { valor: '', etiqueta: this.t().tickets.todasLasCategorias },
      ...this.categorias().map((categoria) => ({
        valor: String(categoria.id),
        etiqueta: categoria.name,
      })),
    ];
  }

  /** El estado de un ticket, como lo lee quien mira. */
  protected estado(ticket: Ticket): string {
    return etiquetaDeEstado(this.t(), ticket.state, this.esUsuario());
  }

  protected solicitante(ticket: Ticket): string {
    return nombreDe(ticket.requester) || '—';
  }

  protected responsable(ticket: Ticket): string {
    return nombreDe(ticket.assignee) || this.t().tickets.sinResponsable;
  }

  /**
   * El motivo del ticket.
   *
   * **En un interno es el del escalado**, que ya está escrito y no se le pide al motor
   * (`docs/modules/ai.md`, decisión 5); en un principal, lo que redactó el motor, en el idioma de
   * quien mira.
   */
  protected motivo(ticket: Ticket): string {
    if (ticket.internal && ticket.reason) {
      return ticket.reason;
    }

    return textoDelResumen(ticket.insights?.motivo, this.textos.idioma(), this.t());
  }

  /** Y la última acción. */
  protected ultimaAccion(ticket: Ticket): string {
    return textoDelResumen(ticket.insights?.ultimaAccion, this.textos.idioma(), this.t());
  }

  protected actualizado(ticket: Ticket): string {
    return fechaHora(ticket.updatedAt, this.textos.idioma(), this.marca.zonaHoraria());
  }

  protected hayFiltros(): boolean {
    const filtros = this.filtros();
    return Boolean(
      filtros.state ||
      filtros.q ||
      filtros.category ||
      filtros.tag ||
      (this.conChipDeTipo() && filtros.type !== this.tipoInicial()),
    );
  }

  protected filtrar(clave: 'type' | 'state' | 'view' | 'category', valor: string): void {
    this.filtros.update((filtros) => ({
      ...filtros,
      [clave]: valor,
      // Al cambiar de vista, el tipo se reajusta: «Observo» mira los dos.
      ...(clave === 'view' ? { type: this.tipoInicial(valor) } : {}),
      page: 1,
    }));
    void this.cargar();
  }

  /** Abre el modal con lo que ya está filtrado, para corregirlo en vez de empezar de cero. */
  protected abrirBusqueda(): void {
    this.borrador.set({ ...this.filtros() });
    this.buscando.set(true);
  }

  /** Cierra el modal **sin aplicar**: lo que se haya escrito en él se descarta. */
  protected cerrarBusqueda(): void {
    this.buscando.set(false);
  }

  /** Aplica lo del modal: es el botón «Buscar». */
  protected aplicarBusqueda(): void {
    this.filtros.set({ ...this.borrador(), page: 1 });
    this.buscando.set(false);
    void this.cargar();
  }

  /** Lo que se escribe en el modal: **no filtra todavía**, sólo rellena el borrador. */
  protected escribirEnElBorrador(
    clave: 'q' | 'tag' | 'state' | 'type' | 'view' | 'category',
    valor: string,
  ): void {
    // La etiqueta se normaliza **al aplicarla**, que es como está guardada: minúsculas y guiones.
    const limpio = clave === 'tag' ? normalizarEtiqueta(valor) : valor;

    this.borrador.update((borrador) => ({
      ...borrador,
      [clave]: limpio,
      // **Al cambiar de vista, el tipo se reajusta**: mirar lo que uno observa cruza los dos tipos, y el
      // «Mis tickets» de Desarrollo arranca en internos —que es su trabajo—, así que «Observo» tiene que
      // soltar ese filtro o no vería lo que sigue del otro lado. Es la misma regla que tenía el chip
      // cuando estaba a la vista: al moverlo al modal se perdió, y lo cazó la prueba de las menciones.
      ...(clave === 'view' ? { type: this.tipoInicial(limpio) } : {}),
    }));
  }

  /** Cuántos criterios están puestos: es lo que dice el resumen de la cabecera. */
  protected cuantosFiltros(): number {
    return this.clavesDeLosFiltros().length;
  }

  /** El resumen de la cabecera: «3 filtros · Buscar, Estado, Categoría». */
  protected resumenDeFiltros(): string {
    const cuantos = this.cuantosFiltros();
    const etiqueta =
      cuantos === 1
        ? this.t().tickets.resumenUnFiltro
        : interpolar(this.t().tickets.resumenFiltros, { n: String(cuantos) });

    return `${etiqueta} · ${this.clavesDeLosFiltros().join(', ')}`;
  }

  /**
   * Los criterios que están puestos, **por su nombre**: es lo que enseña el resumen de la lista, y lo
   * que hace que se entienda por qué salen pocos tickets sin tener que abrir el modal.
   */
  protected clavesDeLosFiltros(): readonly string[] {
    const filtros = this.filtros();
    const puestos: string[] = [];

    if (filtros.q) {
      puestos.push(this.t().tickets.buscarEtiqueta);
    }
    if (this.conChipDeVista() && filtros.view) {
      puestos.push(this.t().tickets.vista);
    }
    if (this.conChipDeTipo() && filtros.type !== this.tipoInicial()) {
      puestos.push(this.t().tickets.filtroTipo);
    }
    if (filtros.state) {
      puestos.push(this.t().tickets.filtroEstado);
    }
    if (filtros.category) {
      puestos.push(this.t().tickets.filtroCategoria);
    }
    if (filtros.tag) {
      puestos.push(this.t().tickets.filtroEtiqueta);
    }

    return puestos;
  }

  /** Quita todos los filtros y recarga: es el botón del resumen, y el del modal. */
  protected limpiarFiltros(): void {
    this.borrador.set({ ...FILTROS_VACIOS, type: this.tipoInicial() });
    this.filtros.set({ ...FILTROS_VACIOS, type: this.tipoInicial() });
    this.buscando.set(false);
    void this.cargar();
  }

  protected irAPagina(pagina: number): void {
    if (pagina < 1 || pagina > this.paginas()) {
      return;
    }

    this.filtros.update((filtros) => ({ ...filtros, page: pagina }));
    void this.cargar();
  }

  protected irANuevo(): void {
    void this.router.navigate(['/tickets/new']);
  }

  /**
   * El tipo con el que arranca la lista.
   *
   * **Mirar lo que uno observa cruza los dos tipos** (2026-09-27): a un desarrollador lo pueden
   * etiquetar en un principal —lo hace Soporte— y su «Mis tickets» arranca en internos, que es su
   * trabajo; si «Observo» se quedara con ese filtro, no vería la mitad de lo que sigue. El
   * responsable pidió que etiquetar valga **en los dos tipos**, y esto es lo que hace que se cumpla en
   * la pantalla.
   */
  private tipoInicial(vista = this.filtros().view): string {
    if (vista === 'watching' && this.conChipDeVista()) {
      return '';
    }

    if (this.listado() === 'principal') {
      return TIPO_PRINCIPAL;
    }
    if (this.listado() === 'interno') {
      return TIPO_INTERNO;
    }
    if (this.esDesarrollo() && !this.esAdministrador()) {
      return TIPO_INTERNO;
    }

    return TIPO_PRINCIPAL;
  }

  /** Los filtros de la lista, con «lo mío» cuando la lista es la bandeja. */
  private filtrosDeLaLista(): FiltrosDeBandeja {
    const filtros = this.filtros();

    // **El Administrador nunca pide «lo mío»**: no tiene tickets propios, y su bandeja es la de
    // siempre, con los dos tipos (decisión 52).
    return this.listado() === 'mios' && !this.esAdministrador()
      ? { ...filtros, mine: true }
      : filtros;
  }

  /**
   * Trae el catálogo para el chip de categoría.
   *
   * Si no llega, la lista se sigue viendo y sólo falta un filtro: no es un fallo de la pantalla, así
   * que no se saca un aviso por esto.
   */
  private async cargarCategorias(): Promise<void> {
    try {
      this.categorias.set((await this.tickets.categorias()).categories);
    } catch {
      this.categorias.set([]);
    }
  }

  private async cargar(): Promise<void> {
    await this.leerLaLista();
    this.vigilarElMotor();
  }

  /** La petición de la lista, sin tocar la vigilancia del motor: la comparten la carga y la relectura. */
  private async leerLaLista(): Promise<void> {
    this.cargando.set(true);

    try {
      this.pagina.set(await this.tickets.listar(this.filtrosDeLaLista()));
    } catch (error) {
      this.pagina.set({ tickets: [], total: 0, page: 1, perPage: 0 });
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.cargando.set(false);
    }
  }

  /**
   * **El motor se refresca solo** (enmienda del 2026-09-29): después de cualquier acción que recargue
   * la lista, si algún ticket se queda con el motivo o la última acción en `pendiente`, la pantalla
   * **vuelve a leer sola** hasta que dejen de estarlo.
   *
   * Es el mismo mecanismo que la ficha (`ticket-vista.ts`, `esperarAlMotor`) adaptado a una lista con
   * varios tickets: **se repite cada `ESPERA_DEL_MOTOR`** y **se para al llegar a `INTENTOS_DEL_MOTOR`**.
   * No se comparte con la ficha porque aquel archivo es de otro compañero y no se puede tocar.
   */
  private vigilarElMotor(): void {
    // Cada carga de verdad —filtro, página o búsqueda— estrena el presupuesto entero.
    this.intentosDelMotor = 0;
    this.programarRelectura();
  }

  /** Programa la siguiente relectura, si aún queda presupuesto y algo sigue pendiente. */
  private programarRelectura(): void {
    if (this.temporizadorDelMotor !== null) {
      clearTimeout(this.temporizadorDelMotor);
      this.temporizadorDelMotor = null;
    }

    if (!this.hayResumenesPendientes() || this.intentosDelMotor >= INTENTOS_DEL_MOTOR) {
      return;
    }

    this.temporizadorDelMotor = setTimeout(() => {
      this.temporizadorDelMotor = null;
      this.intentosDelMotor += 1;
      void this.leerLaLista().then(() => this.programarRelectura());
    }, ESPERA_DEL_MOTOR);
  }

  /** Si algún ticket de la página sigue con el motor pendiente en cualquiera de los dos campos. */
  private hayResumenesPendientes(): boolean {
    return this.pagina().tickets.some(
      (ticket) =>
        ticket.insights?.motivo.state === 'pendiente' ||
        ticket.insights?.ultimaAccion.state === 'pendiente',
    );
  }

  ngOnDestroy(): void {
    if (this.temporizadorDelMotor !== null) {
      clearTimeout(this.temporizadorDelMotor);
      this.temporizadorDelMotor = null;
    }
  }
}
