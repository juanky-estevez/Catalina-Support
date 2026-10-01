import { Component, OnInit, computed, inject, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { BrandService } from '../../core/services/brand.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Icono } from '../../shared/components/icono';
import { Dialogo } from '../../shared/components/dialogo';
import { interpolar } from '../../shared/textos';
import { Campo } from '../../shared/components/campo';
import { Selector } from '../../shared/components/selector';
import { Tarjeta } from '../../shared/components/tarjeta';
import { claveDelError } from '../../shared/errores';
import {
  PAPELES,
  fechaCorta,
  opcionesDeIdioma,
  opcionesDeOrigen,
  opcionesDePapel,
  seReactivaSola,
  etiquetaDeIdioma,
  etiquetaDeOrigen,
  etiquetaDePapel,
} from './etiquetas';
import { UsersService, type CambiosDeCuenta, type Cuenta } from './users.service';

/**
 * La ficha de una cuenta: todo lo que hay que saber de ella y todo lo que se puede hacer con ella.
 *
 * **Sólo la abre un Administrador** (`docs/modules/users.md`, sección 4, decisión 2): Soporte ve la
 * lista, que es lo que necesita para buscar a quien reporta, y edita los nombres desde allí.
 *
 * Las dos acciones que pueden dejar a alguien sin poder entrar —desactivar y cambiar el origen— **dicen
 * lo que va a pasar antes de hacerlo** (`docs/interfaz-y-experiencia.md`, sección 4.4): desactivar
 * pregunta, y el cambio de origen lleva su explicación al lado.
 */
@Component({
  selector: 'app-user-page',
  imports: [Dialogo, RouterLink, Aviso, Boton, Campo, Icono, Selector, Tarjeta],
  templateUrl: './user-page.html',
})
export class UserPage implements OnInit {
  /** El identificador de la cuenta, que llega de la dirección (`/users/{id}`). */
  readonly id = input.required<string>();

  private readonly users = inject(UsersService);
  private readonly textos = inject(TranslationService);
  private readonly marca = inject(BrandService);
  private readonly sesion = inject(SessionService);

  protected readonly cuenta = signal<Cuenta | null>(null);
  protected readonly cargando = signal(true);
  /**
   * La confirmación de mandar el enlace: `null` es «no se está preguntando», y un booleano dice si esa
   * cuenta **ya tenía contraseña** —entonces el enlace es el de recuperación, no el de alta—.
   */
  protected readonly mandandoElEnlace = signal<boolean | null>(null);
  protected readonly trabajando = signal(false);
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);

  // Los campos del formulario: se editan aquí y se guardan de una vez.
  protected readonly nombre = signal('');
  protected readonly apellidos = signal('');
  protected readonly correo = signal('');
  protected readonly papel = signal('usuario');
  protected readonly idioma = signal('es');

  /** El origen elegido, que se guarda con su propia acción: no es un campo del formulario. */
  protected readonly origenElegido = signal('local');

  /** Si se está preguntando antes de desactivar. */
  protected readonly preguntandoDesactivar = signal(false);

  /** Lo que se ha cambiado respecto a lo guardado: es lo único que se manda. */
  /** Si esa cuenta se reactiva sola al entrar por su camino (Keycloak). */
  protected readonly seReactivaSola = seReactivaSola;

  protected readonly cambios = computed<CambiosDeCuenta>(() => {
    const cuenta = this.cuenta();
    if (!cuenta) {
      return {};
    }

    const cambios: {
      name?: string;
      lastName?: string;
      email?: string;
      role?: string;
      language?: string;
    } = {};

    if (this.nombre().trim() !== cuenta.name) {
      cambios.name = this.nombre().trim();
    }
    if (this.apellidos().trim() !== cuenta.lastName) {
      cambios.lastName = this.apellidos().trim();
    }
    if (this.correo().trim().toLowerCase() !== cuenta.email.toLowerCase()) {
      cambios.email = this.correo().trim();
    }
    if (this.papel() !== cuenta.role) {
      cambios.role = this.papel();
    }
    if (this.idioma() !== cuenta.language) {
      cambios.language = this.idioma();
    }

    return cambios;
  });

  /** Si hay algo que guardar: sin cambios, el botón está apagado. */
  protected readonly hayCambios = computed(() => Object.keys(this.cambios()).length > 0);

  /** Si el origen elegido es distinto del que tiene: es lo que enciende su botón. */
  protected readonly cambiaElOrigen = computed(
    () => this.origenElegido() !== this.cuenta()?.origin,
  );

  /**
   * Si esa cuenta es la de quien mira.
   *
   * La ficha **no ofrece desactivar tu propia cuenta**: es quedarse fuera al instante y sin poder
   * volver, porque hace falta que otro te reactive (`docs/modules/users.md`, sección 7).
   */
  protected readonly esMiCuenta = computed(() => this.cuenta()?.id === this.sesion.usuario()?.id);

  ngOnInit(): void {
    void this.cargar();
  }

  protected t() {
    return this.textos.textos();
  }

  protected nombreCompleto(cuenta: Cuenta): string {
    return `${cuenta.name} ${cuenta.lastName}`.trim();
  }

  protected etiquetaDePapel(papel: string): string {
    return etiquetaDePapel(this.t(), papel);
  }

  protected etiquetaDeOrigen(origen: string): string {
    return etiquetaDeOrigen(this.t(), origen);
  }

  protected etiquetaDeIdioma(idioma: string): string {
    return etiquetaDeIdioma(idioma);
  }

  protected ultimaEntrada(cuenta: Cuenta): string {
    return (
      fechaCorta(cuenta.lastLoginAt, this.textos.idioma(), this.marca.zonaHoraria()) ||
      this.t().usuarios.nunca
    );
  }

  protected opcionesDePapel() {
    return opcionesDePapel(this.t(), PAPELES);
  }

  protected opcionesDeOrigen() {
    return opcionesDeOrigen(this.t());
  }

  protected opcionesDeIdioma() {
    return opcionesDeIdioma(this.t());
  }

  /** Guarda los datos de la cuenta. El estado y el origen **no** van por aquí. */
  protected async guardar(): Promise<void> {
    const cuenta = this.cuenta();
    if (!cuenta || !this.hayCambios()) {
      return;
    }

    await this.actuar(async () => {
      const respuesta = await this.users.cambiar(cuenta.id, this.cambios());
      this.poner(respuesta.user);
      return { forma: 'exito', texto: this.t().usuarios.guardada };
    });
  }

  /** Desactiva la cuenta, después de preguntar. */
  protected async desactivar(): Promise<void> {
    const cuenta = this.cuenta();
    if (!cuenta) {
      return;
    }

    await this.actuar(async () => {
      const respuesta = await this.users.desactivar(cuenta.id);
      this.poner(respuesta.user);
      this.preguntandoDesactivar.set(false);

      if (!respuesta.warnings?.length) {
        return { forma: 'exito', texto: this.t().usuarios.desactivadaHecha };
      }

      // El aviso de una cuenta de directorio: se ha desactivado aquí, pero allí sigue activa y
      // volverá a entrar sola (docs/modules/users.md, sección 7).
      const avisos = respuesta.warnings.map((clave) => this.textos.error(clave)).join(' ');
      return { forma: 'atencion', texto: `${this.t().usuarios.desactivadaHecha} ${avisos}` };
    });
  }

  /** Reactiva la cuenta. */
  protected async reactivar(): Promise<void> {
    const cuenta = this.cuenta();
    if (!cuenta) {
      return;
    }

    await this.actuar(async () => {
      const respuesta = await this.users.activar(cuenta.id);
      this.poner(respuesta.user);
      return { forma: 'exito', texto: this.t().usuarios.reactivadaHecha };
    });
  }

  /** Manda otra vez el enlace de contraseña: es la salida para quien no recibió el primero. */
  /**
   * **Pregunta antes de mandar el enlace** (decisión del responsable, 2026-09-28): para una cuenta que
   * ya tiene contraseña, este botón es **restablecerla** —se le manda un enlace de recuperación y la
   * que tenía deja de valer—, así que no puede salir por accidente. Se usa el diálogo de siempre, no un
   * `confirm()` del navegador.
   */
  protected preguntarPorElEnlace(): void {
    this.mensaje.set(null);
    this.mandandoElEnlace.set(this.cuenta()?.hasPassword ?? false);
  }

  /** El aviso de la confirmación: dice **cuál de los dos enlaces** es, que caducan distinto. */
  protected avisoDelEnlace(): string {
    const cuenta = this.cuenta();
    if (!cuenta) {
      return '';
    }

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

  /**
   * Si la contraseña de esa cuenta es nuestra (decisión del responsable, 2026-09-29): **sólo en las
   * locales**. Con Active Directory o Keycloak la comprueba el directorio, así que cambiar la contraseña
   * desde aquí no serviría de nada.
   */
  protected esLocal(cuenta: Cuenta): boolean {
    return cuenta.origin === 'local';
  }

  protected async mandarEnlace(): Promise<void> {
    const cuenta = this.cuenta();
    if (!cuenta) {
      return;
    }

    await this.actuar(async () => {
      const respuesta = await this.users.mandarEnlace(cuenta.id);
      this.poner(respuesta.user);
      return { forma: 'exito', texto: this.t().usuarios.enlaceMandado };
    });
  }

  /** Cambia el origen de la cuenta: la otra acción peligrosa del módulo. */
  protected async guardarOrigen(): Promise<void> {
    const cuenta = this.cuenta();
    if (!cuenta || !this.cambiaElOrigen()) {
      return;
    }

    await this.actuar(async () => {
      const respuesta = await this.users.cambiarOrigen(cuenta.id, this.origenElegido());
      this.poner(respuesta.user);
      return { forma: 'exito', texto: this.t().usuarios.origenGuardado };
    });
  }

  private async cargar(): Promise<void> {
    const id = Number(this.id());
    if (!Number.isInteger(id) || id <= 0) {
      this.cargando.set(false);
      this.mensaje.set({ forma: 'error', texto: this.textos.error('users.notFound') });
      return;
    }

    this.cargando.set(true);

    try {
      const respuesta = await this.users.ficha(id);
      this.poner(respuesta.user);
    } catch (error) {
      this.cuenta.set(null);
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.cargando.set(false);
    }
  }

  /** Deja la ficha con lo que ha contestado el backend: la cuenta y sus campos. */
  private poner(cuenta: Cuenta): void {
    this.cuenta.set(cuenta);
    this.nombre.set(cuenta.name);
    this.apellidos.set(cuenta.lastName);
    this.correo.set(cuenta.email);
    this.papel.set(cuenta.role);
    this.idioma.set(cuenta.language);
    this.origenElegido.set(cuenta.origin);
  }

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
}
