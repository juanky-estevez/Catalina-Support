import { Component, OnInit, inject, signal } from '@angular/core';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Dialogo } from '../../shared/components/dialogo';
import { Tarjeta } from '../../shared/components/tarjeta';
import { claveDelError } from '../../shared/errores';
import {
  TOPE_DE_ETIQUETA,
  interpolar,
  normalizarAlEscribir,
  normalizarEtiqueta,
} from './etiquetas';
import { TicketsService, type Categoria, type Etiqueta } from './tickets.service';

/**
 * **Categorías y etiquetas**: el catálogo con el que se clasifica cada ticket.
 *
 * Es una pantalla de `tickets` y **no una sección de Configuración** (decisión 64): son datos de los
 * tickets, y un módulo del frontend sólo habla con su propia API (`docs/arquitectura.md`, sección 4).
 *
 * **Dos mitades con el mismo peso** (decisión 72): las categorías y las etiquetas, cada una con su
 * tarjeta, su descripción, su cuenta de tickets y sus botones. Antes las etiquetas eran una lista de
 * sólo lectura y ahí es donde se echaba en falta poder mantenerlas.
 *
 * - **Soporte y el Administrador** crean y renombran, en las dos mitades.
 * - **Sólo el Administrador** retira una categoría o la vuelve a poner (decisión 64) y **retira una
 *   etiqueta** (decisión 73): el botón no se le enseña a Soporte, porque la interfaz no ofrece lo que
 *   no se puede hacer.
 * - **Retirar una categoría es desactivarla** (decisión 66): deja de ofrecerse al crear, y los tickets
 *   que la tienen la conservan. **Retirar una etiqueta es quitarla de sus tickets**, y por eso
 *   **pregunta antes** a cuántos afecta (decisión 72).
 *
 * **La clave de una etiqueta es su nombre**, no un identificador: para renombrar y retirar se manda el
 * `tag` que se está viendo, y el servicio lo codifica en la ruta. Por eso el `track` del `@for` y los
 * identificadores de los campos salen del propio `tag`.
 */
@Component({
  selector: 'app-categorias-page',
  imports: [Aviso, Boton, Campo, Dialogo, Tarjeta],
  templateUrl: './categorias-page.html',
})
export class CategoriasPage implements OnInit {
  private readonly tickets = inject(TicketsService);
  private readonly textos = inject(TranslationService);
  private readonly sesion = inject(SessionService);

  protected readonly categorias = signal<readonly Categoria[]>([]);
  protected readonly etiquetas = signal<readonly Etiqueta[]>([]);
  protected readonly cargando = signal(true);
  protected readonly trabajando = signal(false);
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);

  /** El alta de una categoría es un diálogo, como el de una cuenta. */
  protected readonly mostrandoAlta = signal(false);
  protected readonly nombreNueva = signal('');

  /** El alta de una etiqueta es el mismo diálogo, con un solo campo: su nombre. */
  protected readonly mostrandoAltaDeEtiqueta = signal(false);
  protected readonly etiquetaNueva = signal('');

  /** El renombrado se hace en la fila, donde está la categoría. */
  protected readonly renombrando = signal<number | null>(null);
  protected readonly nombreEditado = signal('');

  /** El renombrado de una etiqueta también, pero su clave es el nombre que se está viendo. */
  protected readonly renombrandoEtiqueta = signal<string | null>(null);
  protected readonly nombreEditadoDeEtiqueta = signal('');

  /** La etiqueta que espera confirmación para retirarse: retirarla toca a todos sus tickets. */
  protected readonly retirandoEtiqueta = signal<Etiqueta | null>(null);

  /** El tope de una etiqueta, para el `maxlength` del campo. Lo comprueba también el backend. */
  protected readonly topeDeEtiqueta = TOPE_DE_ETIQUETA;

  /** Si quien mira es Administrador: es el único que retira y vuelve a poner. */
  protected readonly esAdministrador = this.sesion.esAdministrador;

  ngOnInit(): void {
    void this.cargar();
  }

  protected t() {
    return this.textos.textos();
  }

  // ------------------------------------------------------------------ el alta de una categoría

  protected abrirAlta(): void {
    this.mensaje.set(null);
    this.nombreNueva.set('');
    this.mostrandoAlta.set(true);
  }

  protected cerrarAlta(): void {
    this.mostrandoAlta.set(false);
  }

  protected async crear(): Promise<void> {
    const nombre = this.nombreNueva().trim();
    if (!nombre) {
      this.mensaje.set({
        forma: 'error',
        texto: this.textos.error('tickets.category.name.required'),
      });
      return;
    }

    await this.actuar(async () => {
      await this.tickets.crearCategoria(nombre);
      this.mostrandoAlta.set(false);
      await this.cargar();
      return { forma: 'exito', texto: this.t().tickets.categoriaCreada };
    });
  }

  // ------------------------------------------------------------------ el alta de una etiqueta

  protected abrirAltaDeEtiqueta(): void {
    this.mensaje.set(null);
    this.etiquetaNueva.set('');
    this.mostrandoAltaDeEtiqueta.set(true);
  }

  protected cerrarAltaDeEtiqueta(): void {
    this.mostrandoAltaDeEtiqueta.set(false);
  }

  /**
   * Se normaliza **mientras se teclea**, como en el alta de un ticket, para que quien escribe vea lo
   * que va a quedar. No se repite la regla: vive en `normalizarAlEscribir` (`etiquetas.ts`).
   */
  protected alEscribirEtiqueta(valor: string): void {
    this.etiquetaNueva.set(normalizarAlEscribir(valor));
  }

  protected async crearLaEtiqueta(): Promise<void> {
    // La forma final la da `normalizarEtiqueta` al mandarla: el campo enseña el guion del final
    // mientras se escribe, y aquí se recorta (docs/modules/tickets.md, decisión 68).
    const tag = normalizarEtiqueta(this.etiquetaNueva());
    if (!tag) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error('tickets.tag.required') });
      return;
    }

    await this.actuar(async () => {
      await this.tickets.crearEtiqueta(tag);
      this.mostrandoAltaDeEtiqueta.set(false);
      await this.cargar();
      return { forma: 'exito', texto: this.t().tickets.etiquetaCreada };
    });
  }

  // ------------------------------------------------------------------ el renombrado de una categoría

  protected empezarRenombrado(categoria: Categoria): void {
    this.mensaje.set(null);
    this.renombrando.set(categoria.id);
    this.nombreEditado.set(categoria.name);
  }

  protected cancelarRenombrado(): void {
    this.renombrando.set(null);
  }

  protected async guardarNombre(categoria: Categoria): Promise<void> {
    const nombre = this.nombreEditado().trim();
    if (!nombre) {
      this.mensaje.set({
        forma: 'error',
        texto: this.textos.error('tickets.category.name.required'),
      });
      return;
    }

    await this.actuar(async () => {
      await this.tickets.renombrarCategoria(categoria.id, nombre);
      this.renombrando.set(null);
      await this.cargar();
      return { forma: 'exito', texto: this.t().tickets.categoriaRenombrada };
    });
  }

  // ------------------------------------------------------------------ el renombrado de una etiqueta

  protected empezarRenombradoDeEtiqueta(etiqueta: Etiqueta): void {
    this.mensaje.set(null);
    this.renombrandoEtiqueta.set(etiqueta.tag);
    this.nombreEditadoDeEtiqueta.set(etiqueta.tag);
  }

  protected cancelarRenombradoDeEtiqueta(): void {
    this.renombrandoEtiqueta.set(null);
  }

  protected async guardarNombreDeEtiqueta(etiqueta: Etiqueta): Promise<void> {
    const tag = normalizarEtiqueta(this.nombreEditadoDeEtiqueta());
    if (!tag) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error('tickets.tag.required') });
      return;
    }

    await this.actuar(async () => {
      // El cambio vale para **todos los tickets que la llevan** (decisión 72): por eso se manda la
      // etiqueta que se está viendo y no un identificador de una fila.
      await this.tickets.renombrarEtiqueta(etiqueta.tag, tag);
      this.renombrandoEtiqueta.set(null);
      await this.cargar();
      return { forma: 'exito', texto: this.t().tickets.etiquetaRenombrada };
    });
  }

  // ------------------------------------------------------------------ retirar y volver a poner

  protected async retirar(categoria: Categoria): Promise<void> {
    await this.actuar(async () => {
      await this.tickets.cambiarEstadoDeCategoria(categoria.id, false);
      await this.cargar();
      return { forma: 'exito', texto: this.t().tickets.categoriaRetiradaAviso };
    });
  }

  protected async reactivar(categoria: Categoria): Promise<void> {
    await this.actuar(async () => {
      await this.tickets.cambiarEstadoDeCategoria(categoria.id, true);
      await this.cargar();
      return { forma: 'exito', texto: this.t().tickets.categoriaReactivadaAviso };
    });
  }

  // ------------------------------------------------------------------ retirar una etiqueta

  protected abrirRetiradaDeEtiqueta(etiqueta: Etiqueta): void {
    this.mensaje.set(null);
    this.retirandoEtiqueta.set(etiqueta);
  }

  protected cerrarRetiradaDeEtiqueta(): void {
    this.retirandoEtiqueta.set(null);
  }

  /**
   * El aviso que se enseña antes de retirar: dice **a cuántos tickets afecta**, porque quitarla la
   * quita de todos ellos (decisión 72). Es la misma pieza que usa el detalle al cerrar un ticket.
   */
  protected avisoDeRetiradaDeEtiqueta(): string {
    const etiqueta = this.retirandoEtiqueta();
    if (!etiqueta) {
      return '';
    }

    return interpolar(this.t().tickets.avisoDeRetiradaDeEtiqueta, {
      etiqueta: etiqueta.tag,
      n: String(etiqueta.tickets),
    });
  }

  protected async retirarLaEtiqueta(): Promise<void> {
    const etiqueta = this.retirandoEtiqueta();
    if (!etiqueta) {
      return;
    }

    await this.actuar(async () => {
      await this.tickets.retirarEtiqueta(etiqueta.tag);
      this.retirandoEtiqueta.set(null);
      await this.cargar();
      return { forma: 'exito', texto: this.t().tickets.etiquetaRetirada };
    });
  }

  // ------------------------------------------------------------------ el trabajo sucio

  /** Trae el catálogo y las etiquetas. Es lo que se vuelve a pedir tras cada cambio. */
  private async cargar(): Promise<void> {
    try {
      // **El catálogo entero de etiquetas**, no las diez más usadas: aquí se mantienen, y una recién
      // creada que no lleva ningún ticket tiene que verse.
      const [catalogo, usadas] = await Promise.all([
        this.tickets.categorias(),
        this.tickets.etiquetas('', true),
      ]);

      this.categorias.set(catalogo.categories);
      this.etiquetas.set(usadas.tags);
    } catch (error) {
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
}
