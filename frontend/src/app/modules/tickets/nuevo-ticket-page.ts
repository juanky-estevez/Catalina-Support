import { Component, OnInit, computed, inject, signal, viewChild } from '@angular/core';
import { Router, RouterLink } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import {
  EditorConAdjuntos,
  type TextosDelEditorConAdjuntos,
} from '../../shared/components/editor-con-adjuntos';
import { Selector, type OpcionSelector } from '../../shared/components/selector';
import { Icono } from '../../shared/components/icono';
import { Tarjeta } from '../../shared/components/tarjeta';
import { claveDelError } from '../../shared/errores';
import { interpolar, opcionesDeCategoria } from './etiquetas';
import { TicketsService, type Categoria } from './tickets.service';

/**
 * El alta de un ticket: **asunto, categoría y descripción, con los adjuntos dentro**.
 *
 * Es una pantalla y no un diálogo, al revés que el alta de una cuenta: aquí se escribe un texto largo
 * y se adjuntan archivos, y eso necesita sitio (`docs/interfaz-y-experiencia.md`, sección 3.7).
 *
 * **Soporte puede crear el ticket en nombre de otra persona** —es el caso de quien llama por
 * teléfono— y por eso su formulario lleva un campo con **el correo del solicitante**: el correo es el
 * identificador de las personas en este producto, y es lo que tiene a mano
 * (`docs/modules/tickets.md`, sección 5).
 *
 * **El alta estrena la clasificación** (sección 2.3.2): la **categoría es obligatoria** —no hay ticket
 * sin clasificar, y si el catálogo no tiene ninguna activa no se puede crear y se dice—. **Etiquetas no
 * se piden** (decisión 82): son de Soporte y Desarrollo, que las ponen en la ficha del ticket; quien
 * abre el ticket elige su categoría y ya.
 *
 * **El alta tiene un paso más que un comentario** (sección 2.3): el número no existe hasta que el
 * ticket se crea, y el adjunto cuelga de un ticket, así que primero se crea el ticket con su
 * descripción, después se suben sus archivos y después **se guarda la descripción otra vez** con las
 * referencias definitivas. La pantalla no lo enseña: la persona ve un solo botón de crear.
 *
 * Si algún archivo no sube, el ticket ya está creado y se dice: es la salida honrada, en vez de dejar
 * creyendo que no se creó nada.
 */
@Component({
  selector: 'app-nuevo-ticket-page',
  imports: [RouterLink, Aviso, Boton, Campo, EditorConAdjuntos, Selector, Tarjeta, Icono],
  templateUrl: './nuevo-ticket-page.html',
})
export class NuevoTicketPage implements OnInit {
  private readonly tickets = inject(TicketsService);
  private readonly textos = inject(TranslationService);
  private readonly sesion = inject(SessionService);
  private readonly router = inject(Router);

  protected readonly asunto = signal('');
  protected readonly descripcion = signal('');
  protected readonly solicitante = signal('');

  /** El catálogo y la categoría elegida. La categoría es obligatoria. */
  protected readonly categorias = signal<readonly Categoria[]>([]);
  protected readonly categoriaElegida = signal('');
  protected readonly cargandoCategorias = signal(true);

  /** El editor de la descripción, que es quien tiene el texto y sus archivos. */
  private readonly elDeLaDescripcion = viewChild<EditorConAdjuntos>('elDeLaDescripcion');

  protected readonly trabajando = signal(false);
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);

  /** El número del ticket recién creado, cuando los adjuntos no han subido todos. */
  protected readonly creado = signal('');

  /** Sólo Soporte crea un ticket para otra persona. */
  protected readonly puedeElegirSolicitante = computed(
    () => this.sesion.usuario()?.role === 'soporte',
  );

  /** Las categorías que se ofrecen: **sólo las activas** —las retiradas no se ofrecen al crear—. */
  protected readonly categoriasActivas = computed(() =>
    this.categorias().filter((categoria) => categoria.active),
  );

  /** Si se puede crear un ticket: **sin ninguna categoría activa, no** (decisión 65). */
  protected readonly puedeCrearTicket = computed(() => this.categoriasActivas().length > 0);

  ngOnInit(): void {
    void this.cargarCategorias();
  }

  protected t() {
    return this.textos.textos();
  }

  /** Los textos del editor, que viven en el diccionario como todos. */
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

  /**
   * Las opciones del selector de categoría.
   *
   * La primera **no se puede elegir** y está vacía: así el selector arranca diciendo «Elige una
   * categoría» y la obligación se ve, en vez de venir una puesta que nadie ha decidido.
   */
  protected opcionesDeCategoria(): readonly OpcionSelector[] {
    const textos = this.t();

    return opcionesDeCategoria(textos, this.categoriasActivas(), {
      valor: '',
      etiqueta: textos.tickets.elegirCategoria,
      grupo: textos.tickets.categoria,
      deshabilitado: true,
    });
  }

  // ------------------------------------------------------------------ el alta

  /**
   * Da de alta el ticket, con lo que lleva dentro la descripción.
   *
   * Los tres pasos, en el orden que obliga el destino de los archivos: crear, subir y **volver a
   * guardar la descripción** para que quede con sus referencias definitivas
   * (`docs/modules/tickets.md`, sección 2.3).
   */
  protected async crear(): Promise<void> {
    const asunto = this.asunto().trim();
    const categoria = Number(this.categoriaElegida());
    const editor = this.elDeLaDescripcion();
    const descripcion = editor?.html() ?? '';
    const archivos = editor?.archivos() ?? [];

    if (!asunto || (!editor?.textoPlano() && !archivos.length)) {
      this.mensaje.set({ forma: 'error', texto: this.t().tickets.obligatorios });
      return;
    }

    // **La categoría es obligatoria** y el backend también la exige: se comprueba aquí para decirlo en
    // castellano y no gastar una petición que va a volver con una clave.
    if (!categoria) {
      this.mensaje.set({
        forma: 'error',
        texto: this.puedeCrearTicket()
          ? this.textos.error('tickets.category.required')
          : this.t().tickets.sinCategoriasActivas,
      });
      return;
    }

    this.trabajando.set(true);
    this.mensaje.set(null);

    try {
      const respuesta = await this.tickets.crear({
        subject: asunto,
        description: descripcion,
        categoryId: categoria,
        requesterEmail: this.puedeElegirSolicitante() ? this.solicitante().trim() : undefined,
      });

      const numero = respuesta.ticket.number;
      const fallidos = await this.subirAdjuntos(numero, archivos);

      // Si algún adjunto no subió, se queda en la pantalla diciéndolo: el ticket ya existe, y quien lo
      // creó tiene que enterarse antes de irse.
      if (fallidos.length) {
        this.creado.set(numero);
        this.mensaje.set({
          forma: 'atencion',
          texto: `${interpolar(this.t().tickets.creadoConAdjuntos, { numero })} ${fallidos.join(', ')}`,
        });
        return;
      }

      // **No hay que volver a guardar la descripción**, y es la gracia de que el texto referencie los
      // adjuntos por su nombre: el ticket se creó con las referencias ya puestas, y el adjunto que se
      // acaba de subir se encuentra por ese nombre. Antes de esto la descripción se guardaba dos veces
      // y el segundo guardado no cambiaba nada.
      // Y si todo fue bien, se va al ticket: es lo que quiere ver quien acaba de crearlo.
      await this.router.navigate(['/tickets', numero]);
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.trabajando.set(false);
    }
  }

  /** Trae el catálogo. Sin él no se puede clasificar, y se dice en vez de dejar crear a ciegas. */
  private async cargarCategorias(): Promise<void> {
    try {
      this.categorias.set((await this.tickets.categorias()).categories);
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.cargandoCategorias.set(false);
    }
  }

  /** Sube los adjuntos y devuelve los nombres de los que no han subido. */
  private async subirAdjuntos(numero: string, archivos: readonly File[]): Promise<string[]> {
    const fallidos: string[] = [];

    for (const archivo of archivos) {
      try {
        await this.tickets.adjuntar(numero, archivo);
      } catch {
        fallidos.push(archivo.name);
      }
    }

    return fallidos;
  }
}
