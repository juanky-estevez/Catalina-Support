import { Component, computed, inject, signal } from '@angular/core';
import { NgTemplateOutlet } from '@angular/common';
import { RouterLink } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { BrandService } from '../../core/services/brand.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Conmutador, type OpcionConmutador } from '../../shared/components/conmutador';
import { Dialogo } from '../../shared/components/dialogo';
import { Selector } from '../../shared/components/selector';
import { claveDelError } from '../../shared/errores';
import { interpolar } from '../../shared/textos';
import { consultaDeAncho, esPantallaAncha } from '../../shared/pantalla';
import {
  ORIGENES,
  PAPELES,
  fechaCorta,
  opcionesDeOrigen,
  opcionesDePapel,
  etiquetaDeOrigen,
  etiquetaDePapel,
  seReactivaSola,
} from './etiquetas';
import {
  FILTROS_VACIOS,
  UsersService,
  type AltaDeCuenta,
  type Cuenta,
  type FiltrosDeCuentas,
} from './users.service';

/** A partir de aquí la lista se pinta con la tabla: por debajo, con tarjetas (sección 7). */
const ANCHO_DE_TABLA = 768;

/** Lo que se espera antes de buscar: escribir un correo entero es una búsqueda, no seis. */
const ESPERA_DE_BUSQUEDA = 300;

/**
 * La lista de cuentas: el trabajo del día de un Administrador y la búsqueda de Soporte.
 *
 * Enseña **tabla en PC y tarjetas en móvil** —una tabla de seis columnas no se lee en un teléfono—,
 * con **chips de filtro** encima (papel, origen y estado) y la búsqueda por nombre o correo
 * (`docs/interfaz-y-experiencia.md`, sección 3.6).
 *
 * **Soporte edita aquí el nombre y los apellidos, en línea**: no ve la ficha de nadie
 * (`docs/modules/users.md`, decisión 2), así que la lista es el único sitio donde ese permiso se puede
 * ejercer (decisión 15).
 */
@Component({
  selector: 'app-users-page',
  imports: [RouterLink, NgTemplateOutlet, Aviso, Boton, Campo, Conmutador, Dialogo, Selector],
  templateUrl: './users-page.html',
})
export class UsersPage {
  private readonly users = inject(UsersService);
  private readonly textos = inject(TranslationService);
  private readonly marca = inject(BrandService);
  private readonly sesion = inject(SessionService);

  protected readonly cuentas = signal<readonly Cuenta[]>([]);
  protected readonly total = signal(0);
  protected readonly porPagina = signal(0);
  protected readonly cargando = signal(true);

  /**
   * La confirmación de mandar el enlace: `null` es «no se está preguntando», y el valor es la cuenta
   * por la que se pregunta. Para una que ya tiene contraseña es **restablecerla** (decisión del
   * responsable, 2026-09-28).
   */
  protected readonly mandandoElEnlace = signal<Cuenta | null>(null);
  protected readonly trabajando = signal(false);
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);

  /** Los filtros que se están aplicando: es lo que se manda al backend. */
  protected readonly filtros = signal<FiltrosDeCuentas>(FILTROS_VACIOS);

  /** Lo que hay escrito en el buscador, que llega con retardo a los filtros. */
  protected readonly buscador = signal('');

  /** La cuenta que se está editando en línea, si hay alguna. */
  protected readonly editando = signal<number | null>(null);
  protected readonly nombreEditado = signal('');
  protected readonly apellidosEditados = signal('');

  /** El alta, que es un diálogo sobre la lista. */
  protected readonly mostrandoAlta = signal(false);
  protected readonly alta = signal<AltaDeCuenta>({
    name: '',
    lastName: '',
    email: '',
    role: 'usuario',
    origin: 'local',
  });

  private temporizador: ReturnType<typeof setTimeout> | null = null;

  /**
   * Si la pantalla es ancha: decide si se pinta la tabla o las tarjetas.
   *
   * Se decide con una consulta y no con clases de CSS para que **la que no toca no exista**, y no
   * queden dos campos con el mismo identificador, que es lo que deja un campo sin nombre accesible.
   */
  protected readonly esAncha = signal(esPantallaAncha(ANCHO_DE_TABLA));

  /** Si quien mira es Administrador: es quien abre fichas y reparte papeles. */
  protected readonly esAdministrador = this.sesion.esAdministrador;

  /** Si quien mira es Soporte y sólo Soporte: edita en línea, porque no tiene ficha. */
  protected readonly esSoporte = computed(() => this.sesion.usuario()?.role === 'soporte');

  /** Cuántas páginas hay, para saber si hay siguiente. */
  protected readonly paginas = computed(() =>
    Math.max(1, Math.ceil(this.total() / (this.porPagina() || 1))),
  );

  /** Los papeles que puede repartir quien está dando de alta: Soporte, sólo `usuario`. */
  protected readonly papelesDelAlta = computed<readonly string[]>(() =>
    this.esAdministrador() ? PAPELES : ['usuario'],
  );

  constructor() {
    consultaDeAncho(ANCHO_DE_TABLA)?.addEventListener('change', (evento) =>
      this.esAncha.set(evento.matches),
    );
    void this.cargar();
  }

  protected t() {
    return this.textos.textos();
  }

  // ------------------------------------------------------------------ la lista

  /** Los chips del papel: todos, y los cuatro. */
  protected opcionesDePapel(): readonly OpcionConmutador[] {
    return [
      { valor: '', etiqueta: this.t().usuarios.todos },
      ...PAPELES.map((papel) => ({
        valor: papel,
        etiqueta: etiquetaDePapel(this.t(), papel),
      })),
    ];
  }

  /** Los chips del origen. */
  protected opcionesDeOrigen(): readonly OpcionConmutador[] {
    return [
      { valor: '', etiqueta: this.t().usuarios.todos },
      ...ORIGENES.map((origen) => ({
        valor: origen,
        etiqueta: etiquetaDeOrigen(this.t(), origen),
      })),
    ];
  }

  /** Los chips del estado: todas, activas o desactivadas. */
  protected opcionesDeEstado(): readonly OpcionConmutador[] {
    return [
      { valor: '', etiqueta: this.t().usuarios.todos },
      { valor: 'true', etiqueta: this.t().usuarios.activas },
      { valor: 'false', etiqueta: this.t().usuarios.desactivadas },
    ];
  }

  /** Lee el papel de una cuenta como se llama en la interfaz. */
  protected papel(cuenta: Cuenta): string {
    return etiquetaDePapel(this.t(), cuenta.role);
  }

  /** Lee el origen de una cuenta como se llama en la interfaz. */
  /**
   * Si la contraseña de esa cuenta es nuestra (decisión del responsable, 2026-09-29): **sólo en las
   * locales**. Con Active Directory o Keycloak la comprueba el directorio, así que cambiar la contraseña
   * desde aquí no serviría de nada.
   */
  protected esLocal(cuenta: Cuenta): boolean {
    return cuenta.origin === 'local';
  }

  protected origen(cuenta: Cuenta): string {
    return etiquetaDeOrigen(this.t(), cuenta.origin);
  }

  /** Cuándo entró por última vez, o el texto que dice que nunca. */
  protected ultimaEntrada(cuenta: Cuenta): string {
    return (
      fechaCorta(cuenta.lastLoginAt, this.textos.idioma(), this.marca.zonaHoraria()) ||
      this.t().usuarios.nunca
    );
  }

  /** El nombre completo, que es como se conoce a una persona. */
  protected nombreCompleto(cuenta: Cuenta): string {
    return `${cuenta.name} ${cuenta.lastName}`.trim();
  }

  /**
   * Si esa cuenta es la de quien está mirando.
   *
   * Sirve para **no ofrecerle el botón de desactivarse a sí mismo**: desactivarse es quedarse fuera al
   * instante y sin poder volver, porque hace falta que otro le reactive
   * (`docs/modules/users.md`, sección 7). El backend también lo rechaza, con 403.
   */
  /** Si esa cuenta se reactiva sola al entrar por su camino (Keycloak). */
  protected readonly seReactivaSola = seReactivaSola;

  protected esMiCuenta(cuenta: Cuenta): boolean {
    return cuenta.id === this.sesion.usuario()?.id;
  }

  /** Un filtro cambia, y se vuelve a la primera página: lo que se busca es desde el principio. */
  protected filtrar(clave: 'role' | 'origin' | 'active', valor: string): void {
    this.filtros.update((filtros) => ({ ...filtros, [clave]: valor, page: 1 }));
    void this.cargar();
  }

  /** La búsqueda llega con retardo: se espera a que se termine de escribir. */
  protected buscar(valor: string): void {
    this.buscador.set(valor);

    if (this.temporizador) {
      clearTimeout(this.temporizador);
    }

    this.temporizador = setTimeout(() => {
      this.filtros.update((filtros) => ({ ...filtros, q: valor, page: 1 }));
      void this.cargar();
    }, ESPERA_DE_BUSQUEDA);
  }

  /** Quita todos los filtros: es la salida cuando una búsqueda no encuentra nada. */
  protected limpiarFiltros(): void {
    this.buscador.set('');
    this.filtros.set(FILTROS_VACIOS);
    void this.cargar();
  }

  /** Si hay algún filtro puesto: es lo que decide si se ofrece quitarlos. */
  protected hayFiltros(): boolean {
    const filtros = this.filtros();
    return Boolean(filtros.role || filtros.origin || filtros.active || filtros.q);
  }

  protected irAPagina(pagina: number): void {
    if (pagina < 1 || pagina > this.paginas()) {
      return;
    }

    this.filtros.update((filtros) => ({ ...filtros, page: pagina }));
    void this.cargar();
  }

  // ------------------------------------------------------ editar en línea (Soporte)

  /** Empieza a editar el nombre y los apellidos de una fila, donde está. */
  protected empezarEdicion(cuenta: Cuenta): void {
    this.mensaje.set(null);
    this.editando.set(cuenta.id);
    this.nombreEditado.set(cuenta.name);
    this.apellidosEditados.set(cuenta.lastName);
  }

  protected cancelarEdicion(): void {
    this.editando.set(null);
  }

  /** Guarda el nombre y los apellidos. La cuenta se cambia en la lista, sin volver a pedirla. */
  protected async guardarNombre(cuenta: Cuenta): Promise<void> {
    const nombre = this.nombreEditado().trim();
    const apellidos = this.apellidosEditados().trim();

    if (!nombre || !apellidos) {
      this.mensaje.set({ forma: 'error', texto: this.t().errores['users.name.required'] });
      return;
    }

    await this.actuar(async () => {
      const respuesta = await this.users.cambiar(cuenta.id, { name: nombre, lastName: apellidos });
      this.reemplazar(respuesta.user);
      this.editando.set(null);
      return { forma: 'exito', texto: this.t().usuarios.guardada };
    });
  }

  // ------------------------------------------------------------------ las acciones

  protected async desactivar(cuenta: Cuenta): Promise<void> {
    await this.actuar(async () => {
      const respuesta = await this.users.desactivar(cuenta.id);
      this.reemplazar(respuesta.user);
      return {
        forma: respuesta.warnings?.length ? 'atencion' : 'exito',
        texto: this.conAvisos(this.t().usuarios.desactivadaHecha, respuesta.warnings),
      };
    });
  }

  protected async reactivar(cuenta: Cuenta): Promise<void> {
    await this.actuar(async () => {
      const respuesta = await this.users.activar(cuenta.id);
      this.reemplazar(respuesta.user);
      return { forma: 'exito', texto: this.t().usuarios.reactivadaHecha };
    });
  }

  /** Pregunta antes de mandarlo, diciendo cuál de los dos enlaces es. */
  protected preguntarPorElEnlace(cuenta: Cuenta): void {
    this.mensaje.set(null);
    this.mandandoElEnlace.set(cuenta);
  }

  protected avisoDelEnlace(cuenta: Cuenta): string {
    const clave = cuenta.hasPassword
      ? this.t().usuarios.mandarEnlaceRecuperacion
      : this.t().usuarios.mandarEnlaceAlta;

    return `${interpolar(this.t().usuarios.mandarEnlacePregunta, {
      nombre: `${cuenta.name} ${cuenta.lastName}`.trim() || cuenta.email,
    })} ${clave}`;
  }

  protected cerrarLaConfirmacion(): void {
    this.mandandoElEnlace.set(null);
  }

  protected async mandarEnlace(cuenta: Cuenta): Promise<void> {
    await this.actuar(async () => {
      const respuesta = await this.users.mandarEnlace(cuenta.id);
      this.reemplazar(respuesta.user);
      return { forma: 'exito', texto: this.t().usuarios.enlaceMandado };
    });
  }

  // ------------------------------------------------------------------ el alta

  protected abrirAlta(): void {
    this.mensaje.set(null);
    this.alta.set({
      name: '',
      lastName: '',
      email: '',
      // Soporte sólo puede crear usuarios, y un Administrador empieza por el papel más común.
      role: 'usuario',
      origin: 'local',
    });
    this.mostrandoAlta.set(true);
  }

  protected cerrarAlta(): void {
    this.mostrandoAlta.set(false);
  }

  protected campoDelAlta(campo: 'name' | 'lastName' | 'email', valor: string): void {
    this.alta.update((alta) => ({ ...alta, [campo]: valor }));
  }

  protected campoDelAltaElegido(campo: 'role' | 'origin', valor: string): void {
    this.alta.update((alta) => ({ ...alta, [campo]: valor }));
  }

  protected opcionesDePapelDelAlta() {
    return opcionesDePapel(this.t(), this.papelesDelAlta());
  }

  protected opcionesDeOrigenDelAlta() {
    return opcionesDeOrigen(this.t());
  }


  /** Da de alta la cuenta y, si sale bien, cierra el diálogo y refresca la lista. */
  protected async crear(): Promise<void> {
    const alta = this.alta();
    if (!alta.name.trim() || !alta.lastName.trim() || !alta.email.trim()) {
      this.mensaje.set({ forma: 'error', texto: this.t().usuarios.obligatorios });
      return;
    }

    this.trabajando.set(true);
    this.mensaje.set(null);

    try {
      const respuesta = await this.users.crear(alta);
      this.mostrandoAlta.set(false);
      this.filtros.set(FILTROS_VACIOS);
      this.buscador.set('');
      await this.cargar();

      // `invited` en falso quiere decir que la cuenta existe pero el enlace no salió: se dice, en vez
      // de dejar a quien da de alta creyendo que todo fue bien (docs/modules/users.md, sección 5).
      const sinEnlace = respuesta.invited === false && respuesta.user.origin === 'local';
      this.mensaje.set({
        forma: sinEnlace ? 'atencion' : 'exito',
        texto: sinEnlace ? this.t().usuarios.creadaSinEnlace : this.t().usuarios.creada,
      });
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.trabajando.set(false);
    }
  }

  // ------------------------------------------------------------------ el trabajo sucio

  private async cargar(): Promise<void> {
    this.cargando.set(true);

    try {
      const pagina = await this.users.listar(this.filtros());
      this.cuentas.set(pagina.users);
      this.total.set(pagina.total);
      this.porPagina.set(pagina.perPage);
    } catch (error) {
      this.cuentas.set([]);
      this.total.set(0);
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.cargando.set(false);
    }
  }

  /** Hace una acción y deja la pantalla avisada de lo que pasó, bien o mal. */
  private async actuar(accion: () => Promise<MensajeDePantalla>): Promise<void> {
    this.trabajando.set(true);
    this.mensaje.set(null);

    try {
      this.mensaje.set(await accion());
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.trabajando.set(false);
    }
  }

  /** La cuenta que ha cambiado, en su sitio de la lista: sin volver a pedir la página entera. */
  private reemplazar(cuenta: Cuenta): void {
    this.cuentas.update((cuentas) =>
      cuentas.map((actual) => (actual.id === cuenta.id ? cuenta : actual)),
    );
  }

  /** Un aviso del backend, traducido, detrás del texto de lo que sí se ha hecho. */
  private conAvisos(texto: string, avisos: readonly string[] | undefined): string {
    if (!avisos?.length) {
      return texto;
    }

    return `${texto} ${avisos.map((clave) => this.textos.error(clave)).join(' ')}`;
  }
}
