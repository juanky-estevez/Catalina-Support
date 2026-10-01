import { Component, OnInit, computed, inject, input, signal } from '@angular/core';
import { Router } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Conmutador, type OpcionConmutador } from '../../shared/components/conmutador';
import { Estado } from '../../shared/components/estado';
import { claveDelError } from '../../shared/errores';
import { AccionesDelTicket } from './acciones-del-ticket';
import { etiquetaDeEstado } from './etiquetas';
import { TicketVista, type Responsables } from './ticket-vista';
import { TicketsService, type Categoria, type DetalleDeTicket } from './tickets.service';

/** Las tres formas de mirar un detalle con interno. */
type FormaDeVer = 'principal' | 'losdos' | 'interno';

/**
 * El detalle de un ticket: **la pieza central**, y una sola pantalla para los cuatro papeles.
 *
 * Aquí vive **la vista doble** (`docs/interfaz-y-experiencia.md`, sección 5): cuando el ticket tiene
 * interno y quien mira puede verlo, aparece el conmutador de tres posiciones —Principal · Los dos ·
 * Interno— y por defecto se ven los dos, porque quien está ahí quiere las dos cosas. **Si no hay
 * interno no aparece nada**: no hay nada que comparar.
 *
 * El usuario **nunca ve el interno** —ni el conmutador, ni la columna—: lo que ve es su principal con
 * lo que Soporte le haya escrito (`docs/usuarios-y-permisos.md`, regla 1 de la sección 3).
 */
@Component({
  selector: 'app-ticket-page',
  imports: [AccionesDelTicket, Aviso, Boton, Conmutador, Estado, TicketVista],
  templateUrl: './ticket-page.html',
})
export class TicketPage implements OnInit {
  /** El número que viene en la dirección: `ACME-2026-0042` o `INT-ACME-2026-0042`. */
  readonly numero = input.required<string>();

  private readonly tickets = inject(TicketsService);
  private readonly textos = inject(TranslationService);
  private readonly sesion = inject(SessionService);
  private readonly router = inject(Router);

  /** El principal, si lo hay: es el que se pide o el padre de un interno. */
  protected readonly principal = signal<DetalleDeTicket | null>(null);
  protected readonly interno = signal<DetalleDeTicket | null>(null);

  protected readonly cargando = signal(true);
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);
  protected readonly responsables = signal<Responsables>({ main: [], internal: [] });
  /** El catálogo de categorías, que la ficha necesita para poder cambiar la clasificación. */
  protected readonly categorias = signal<readonly Categoria[]>([]);
  protected readonly forma = signal<FormaDeVer>('principal');
  protected readonly copiado = signal(false);

  protected readonly esUsuario = computed(() => this.sesion.usuario()?.role === 'usuario');

  /** Desarrollo entra por el interno: es el ticket con el que trabaja. */
  protected readonly esDesarrollo = computed(() => this.sesion.usuario()?.role === 'desarrollo');

  /** Si hay interno y quien mira puede verlo: sólo entonces existe la vista doble. */
  protected readonly hayInterno = computed(() => !this.esUsuario() && this.interno() !== null);

  /**
   * La vista con la que se abre el ticket.
   *
   * **Y si se ha pedido el interno, se enseña el interno** aunque seas Soporte: es lo que espera
   * cualquiera que pulse el enlace de un `INT-…` (decisión del responsable, 2026-09-27).
   */
  private vistaDeEntrada(): FormaDeVer {
    if (!this.hayInterno()) {
      return 'principal';
    }

    // Lo que se ha pedido manda: si la dirección era la de un `INT-…`, se enseña el interno.
    if (this.interno() !== null && this.principal() !== null && this.pedidoEraElInterno) {
      return 'interno';
    }

    return this.esDesarrollo() ? 'interno' : 'principal';
  }

  /** Si el número que se abrió era el del interno y no el del principal. */
  private pedidoEraElInterno = false;

  protected readonly opcionesDeVista = computed<readonly OpcionConmutador[]>(() => [
    { valor: 'principal', etiqueta: this.t().tickets.vistaPrincipal },
    { valor: 'losdos', etiqueta: this.t().tickets.vistaDoble },
    { valor: 'interno', etiqueta: this.t().tickets.vistaInterna },
  ]);

  /**
   * El ticket que se está mirando cuando **sólo se ve uno**, que es el que lleva el desplegable de
   * acciones en la fila de arriba (decisión 78).
   *
   * Con el conmutador en «Los dos» devuelve nulo a propósito: ahí no hay un solo ticket, y cada
   * columna lleva su propio desplegable en su cabecera, porque cada uno tiene su estado y desde un
   * control de arriba no se sabría a cuál se le cambia.
   */
  protected readonly detalleUnico = computed<DetalleDeTicket | null>(() => {
    switch (this.forma()) {
      case 'interno':
        return this.interno();
      case 'losdos':
        return null;
      default:
        return this.principal();
    }
  });

  ngOnInit(): void {
    void this.cargar();
  }

  protected t() {
    return this.textos.textos();
  }

  /** El número que se enseña arriba: el que se ha pedido. */
  /**
   * **El número con el que trabaja quien mira** (decisión del responsable, 2026-09-29): Desarrollo entra
   * por el interno, así que el interno es su número; Soporte y el Administrador, por el principal. El
   * otro va al lado, como referencia y en apagado.
   */
  protected numeroDeArriba(): string {
    return this.numeroDeMiEquipo() ?? this.numeroDeReferencia() ?? this.numero();
  }

  /** El número del otro ticket, el que no es el de quien mira: la referencia. */
  protected numeroDeReferencia(): string | null {
    return this.esDesarrollo()
      ? (this.principal()?.ticket.number ?? null)
      : (this.interno()?.ticket.number ?? null);
  }

  /** El número del ticket del equipo de quien mira, si el ticket tiene los dos lados. */
  private numeroDeMiEquipo(): string | null {
    if (this.esDesarrollo()) {
      return this.interno()?.ticket.number ?? null;
    }

    return this.principal()?.ticket.number ?? null;
  }

  protected estadoDeArriba(): { estado: string; texto: string } | null {
    const detalle = this.principal() ?? this.interno();
    if (!detalle) {
      return null;
    }

    // **Con el mismo lenguaje que el resto de la pantalla**: si el usuario lee «Recibido» abajo, no
    // puede leer «Nuevo» arriba (docs/interfaz-y-experiencia.md, sección 3.5). Lo encontró la prueba
    // de aspecto, que buscaba «Recibido» y encontró «Nuevo».
    return {
      estado: detalle.ticket.state,
      texto: etiquetaDeEstado(this.t(), detalle.ticket.state, this.esUsuario()),
    };
  }

  /**
   * Copia el número: es lo que se dice en voz alta y lo que se pega en un correo.
   *
   * Si el navegador no deja usar el portapapeles, **se dice**: un botón que no hace nada y no lo
   * cuenta es peor que no tenerlo, porque quien lo pulsa se queda creyendo que ya lo tiene copiado.
   */
  protected async copiar(): Promise<void> {
    try {
      await navigator.clipboard.writeText(this.numeroDeArriba());
      this.mensaje.set(null);
      this.copiado.set(true);
      setTimeout(() => this.copiado.set(false), 2000);
    } catch {
      this.mensaje.set({ forma: 'atencion', texto: this.t().tickets.copiarNoSePudo });
    }
  }

  protected irALaBandeja(): void {
    void this.router.navigate(['/tickets']);
  }

  /** Vuelve a leer el ticket y su interno, después de cualquier cambio. */
  protected async recargar(): Promise<void> {
    await this.cargar(false);
  }

  private async cargar(conCargando = true): Promise<void> {
    if (conCargando) {
      this.cargando.set(true);
    }

    try {
      const pedido = await this.tickets.ficha(this.numero());

      if (pedido.ticket.internal) {
        // Se ha abierto el interno: su principal es el otro lado de la vista doble.
        this.pedidoEraElInterno = true;
        this.interno.set(pedido);
        this.principal.set(
          pedido.ticket.parent ? await this.tickets.ficha(pedido.ticket.parent) : null,
        );
      } else {
        this.principal.set(pedido);
        this.interno.set(
          pedido.ticket.child ? await this.tickets.ficha(pedido.ticket.child) : null,
        );
      }

      // **La vista por defecto es la del papel** (decisión del responsable, 2026-09-27): Soporte entra
      // por el principal y Desarrollo por el interno, que es el ticket con el que cada uno trabaja. Los
      // dos pueden cambiar a «Los dos» o al otro con el conmutador.
      this.forma.set(this.vistaDeEntrada());

      await this.cargarResponsables();
      await this.cargarCategorias();
    } catch (error) {
      this.principal.set(null);
      this.interno.set(null);
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.cargando.set(false);
    }
  }

  /** Quién puede ser responsable: lo dice el módulo de tickets, que es la API de esta pantalla. */
  private async cargarResponsables(): Promise<void> {
    if (this.esUsuario()) {
      return;
    }

    try {
      this.responsables.set(await this.tickets.responsables());
    } catch {
      // Sin la lista no se puede asignar, pero el ticket se sigue viendo: no es un fallo de la pantalla.
    }
  }

  /**
   * El catálogo de categorías, que la ficha necesita para poder cambiar la clasificación.
   *
   * Si no llega, el ticket se sigue viendo con la categoría que trae: sólo se pierde poder cambiarla.
   */
  private async cargarCategorias(): Promise<void> {
    try {
      this.categorias.set((await this.tickets.categorias()).categories);
    } catch {
      this.categorias.set([]);
    }
  }
}
