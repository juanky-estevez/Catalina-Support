import { Component, computed, effect, inject, input, output, signal } from '@angular/core';

import { TranslationService } from '../../core/i18n/translation.service';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Dialogo } from '../../shared/components/dialogo';
import { claveDelError } from '../../shared/errores';
import { TicketsService, type Etiqueta } from './tickets.service';

/**
 * El modal que abre el botón de etiquetas de la ficha (decisión 82 de `docs/modules/tickets.md`).
 *
 * **Antes se etiquetaba con un campo y sus sugerencias**, y cada añadir y cada quitar se guardaba al
 * momento (decisión 74). Con el catálogo entero eso era incómodo —había que traer las diez más usadas y
 * la recién creada no salía—, así que ahora se etiqueta de una vez: se pide **todo el catálogo** con
 * cuántos tickets lleva cada una (`this.tickets.etiquetas('', true)`, decisión 68-quinquies), se
 * **filtra por nombre** mientras se escribe, se marcan y desmarcan **varias a la vez** —las que el
 * ticket ya lleva salen marcadas— y **«Guardar» manda la lista entera en un solo `PATCH`** (el backend
 * reemplaza las del ticket con lo que le llegue). «Cancelar» no cambia nada.
 *
 * **El `PATCH` no se hace aquí**: lo hace la ficha con su `editar`, que es quien ya sabe enseñar el
 * resultado y el error donde se avisa. Este componente sólo recoge lo que hay que guardar.
 *
 * La ventana es la del proyecto (`app-dialogo`, un `<dialog>` nativo): atrapa el foco, se cierra con
 * Escape y con su ✕. El buscador es `app-campo` —con su etiqueta de verdad— y cada casilla lleva el
 * nombre de su etiqueta, así que buscar, marcar y guardar se puede hacer **sin ratón**.
 */
@Component({
  selector: 'app-modal-etiquetas',
  imports: [Aviso, Boton, Campo, Dialogo],
  templateUrl: './modal-etiquetas.html',
})
export class ModalEtiquetas {
  /** Si se enseña. Al ponerse en verdadero, se trae el catálogo entero. */
  readonly abierto = input(false);
  /** El número del ticket, que sólo se usa para no repetir identificadores en la página. */
  readonly numero = input.required<string>();
  /** Las etiquetas que el ticket lleva ahora mismo: son las que salen marcadas. */
  readonly etiquetas = input<readonly string[]>([]);
  /** Mientras el `PATCH` va en camino: el botón de guardar lo dice y no se puede mandar dos veces. */
  readonly trabajando = input(false);

  /** Se ha cerrado sin guardar (Cancelar, Escape o la ✕). */
  readonly cerrado = output<void>();
  /** Lo que hay que guardar: la lista entera de etiquetas marcadas. */
  readonly guardar = output<readonly string[]>();

  private readonly tickets = inject(TicketsService);
  private readonly textos = inject(TranslationService);

  /** El catálogo entero, tal como lo devuelve la API: cada etiqueta con su cuenta de tickets. */
  protected readonly catalogo = signal<readonly Etiqueta[]>([]);
  /** Lo que se escribe en el buscador: filtra por nombre, sin ir al servidor. */
  protected readonly buscando = signal('');
  /** Las que están marcadas ahora mismo, que es lo que se manda al guardar. */
  protected readonly marcadas = signal<ReadonlySet<string>>(new Set());
  protected readonly cargando = signal(false);
  protected readonly error = signal('');

  /** Evita volver a pedir el catálogo en cada ciclo: sólo se pide al abrirse. */
  private estabaAbierto = false;

  constructor() {
    effect(() => {
      const abierto = this.abierto();

      if (abierto && !this.estabaAbierto) {
        this.estabaAbierto = true;
        void this.traerElCatalogo();
      } else if (!abierto) {
        this.estabaAbierto = false;
      }
    });
  }

  protected t() {
    return this.textos.textos();
  }

  /** El catálogo filtrado **por nombre** con lo que se escribe. Vacío, se enseña entero. */
  protected readonly filtradas = computed<readonly Etiqueta[]>(() => {
    const texto = this.buscando().trim().toLowerCase();
    if (!texto) {
      return this.catalogo();
    }

    return this.catalogo().filter((etiqueta) => etiqueta.tag.toLowerCase().includes(texto));
  });

  /** Marca o desmarca una etiqueta: se pueden tocar varias antes de guardar. */
  protected alternar(tag: string, marcada: boolean): void {
    this.marcadas.update((actuales) => {
      const copia = new Set(actuales);
      if (marcada) {
        copia.add(tag);
      } else {
        copia.delete(tag);
      }

      return copia;
    });
  }

  /** Manda la lista entera de las marcadas: la ficha la guarda en un solo `PATCH`. */
  protected guardarLasMarcadas(): void {
    this.guardar.emit([...this.marcadas()]);
  }

  protected cerrar(): void {
    this.cerrado.emit();
  }

  private async traerElCatalogo(): Promise<void> {
    this.buscando.set('');
    this.error.set('');
    this.cargando.set(true);
    // Las que el ticket ya lleva salen marcadas: lo que se mande al guardar es lo que queda.
    this.marcadas.set(new Set(this.etiquetas()));

    try {
      const respuesta = await this.tickets.etiquetas('', true);
      this.catalogo.set(this.conLasQueLlevaElTicket(respuesta.tags));
    } catch (error) {
      this.error.set(this.textos.error(claveDelError(error)));
    } finally {
      this.cargando.set(false);
    }
  }

  /**
   * El catálogo entero **más las etiquetas del ticket que no estén en él**.
   *
   * El backend reemplaza las etiquetas del ticket con lo que llegue, así que una que el ticket lleve y
   * el catálogo no enseñe —no debería pasar: un ticket no lleva una etiqueta que no existe— se
   * perdería al guardar. Se añade a la lista, sin cuenta, para que salga marcada y se pueda conservar.
   */
  private conLasQueLlevaElTicket(catalogo: readonly Etiqueta[]): readonly Etiqueta[] {
    const vistas = new Set(catalogo.map((etiqueta) => etiqueta.tag));
    const queFaltan = this.etiquetas()
      .filter((tag) => !vistas.has(tag))
      .map((tag) => ({ tag, tickets: 0 }));

    return queFaltan.length ? [...catalogo, ...queFaltan] : catalogo;
  }
}
