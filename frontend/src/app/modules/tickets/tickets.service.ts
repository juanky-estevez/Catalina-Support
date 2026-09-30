import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

/** Una cuenta, reducida a lo que se enseña dentro de un ticket. */
export interface Persona {
  readonly id: number;
  readonly name: string;
  readonly lastName: string;
  readonly email?: string;
  readonly role: string;
}

/**
 * Una categoría del catálogo: el **qué es** del ticket.
 *
 * Va entera —con su nombre y su estado— porque la pantalla la enseña, la elige y, si es de quien
 * mantiene el catálogo, la edita. `tickets` es **cuántos tickets la usan**, que es lo que hace falta
 * para retirarla con criterio y para enseñarlo en «Categorías y etiquetas»
 * (`docs/modules/tickets.md`, sección 2.3.2).
 */
export interface Categoria {
  readonly id: number;
  readonly name: string;
  /** Si se sigue ofreciendo al crear un ticket. **Retirar es desactivar** (decisión 66). */
  readonly active: boolean;
  readonly tickets: number;
}

/**
 * La categoría **como viaja dentro de un ticket**: su identificador y su nombre.
 *
 * No lleva el recuento de tickets —eso es del catálogo, que es quien lo cuenta— y sí lleva `active`,
 * que es lo que permite saber si la categoría del ticket está retirada sin pedir el catálogo entero.
 */
export interface CategoriaDelTicket {
  readonly id: number;
  readonly name: string;
  readonly active?: boolean;
}

/**
 * Una etiqueta que ya existe y cuántos tickets la llevan.
 *
 * No hay catálogo de etiquetas: esto es lo que se ha usado, y se enseña para sugerirlo mientras se
 * escribe (`docs/modules/tickets.md`, sección 2.3.2).
 */
export interface Etiqueta {
  readonly tag: string;
  readonly tickets: number;
}

/** Un ticket, principal o interno (`docs/modules/tickets.md`, sección 2.1). */
/**
 * Un cambio de estado del ticket.
 *
 * **El estado se admite en `state` y en `to`** —el backend acepta los dos, y lo dice en su propio
 * comentario—: `state` es lo que mandaba la pantalla desde el principio y `to` es el nombre que fija la
 * tabla de endpoints del documento. `comment` es **obligatorio al cerrar** (decisión 81) y opcional en
 * los demás estados.
 */
export interface CambioDeEstado {
  readonly state?: string;
  readonly to?: string;
  readonly comment?: string;
}

export interface Ticket {
  readonly number: string;
  readonly internal: boolean;
  readonly subject: string;
  readonly description: string;
  readonly state: string;
  /**
   * La **categoría** del caso: el *qué es* del ticket. Es obligatoria y la hereda el interno al
   * leerse, así que aquí siempre viene —el backend no puede dejar un ticket sin clasificar—, pero se
   * declara opcional por si una respuesta vieja no la trae (`docs/modules/tickets.md`, decisión 67).
   */
  readonly category?: CategoriaDelTicket;
  /** Las **etiquetas** del caso, ya normalizadas: minúsculas y con guiones (`red-wifi`). */
  readonly tags?: readonly string[];
  /** El motivo del escalado. Sólo en los internos. */
  readonly reason?: string;
  /** El número del principal, en los internos. */
  readonly parent?: string;
  /** El número del interno, en los principales que lo tienen. */
  readonly child?: string;
  readonly requester?: Persona;
  readonly createdBy?: Persona;
  readonly assignee?: Persona;
  readonly subjectEditedAt?: string;
  readonly descriptionEditedAt?: string;
  readonly resolvedAt?: string;
  readonly closedAt?: string;
  readonly createdAt: string;
  readonly updatedAt: string;

  /**
   * Los dos campos que redacta el motor de IA: el **motivo** y la **última acción**
   * (`docs/modules/ai.md`). Viajan siempre, aunque estén vacíos: el estado dice si se están
   * escribiendo, si el motor no está o si fallaron.
   */
  readonly insights?: Insights;
}

/**
 * Alguien que **observa** el ticket: la cuenta, quién la etiquetó y cuándo.
 *
 * La cuenta viene entera —y no sólo su nombre— porque es lo que hace falta para pintarla y para
 * quitarla: el aspa manda **el identificador de la cuenta**, que es de quien se habla.
 */
export interface Observador {
  readonly id: number;
  readonly account: Persona;
  readonly addedBy?: Persona;
  readonly createdAt?: string;
}

/** Los dos resúmenes de un ticket. */
export interface Insights {
  readonly motivo: Resumen;
  readonly ultimaAccion: Resumen;
}

/** Un resumen del motor: su estado, sus dos idiomas y la clave del error si no se pudo escribir. */
export interface Resumen {
  /** `pendiente`, `listo`, `error` o `sin_motor`. Vacío: de un ticket que nunca ha pedido resumen. */
  readonly state: string;
  readonly es?: string;
  readonly en?: string;
  readonly errorKey?: string;
}

/** Un comentario de la conversación. */
export interface Comentario {
  readonly id: number;
  readonly author?: Persona;
  readonly body: string;
  readonly edited: boolean;
  readonly deleted: boolean;
  readonly createdAt: string;
}

/** Un adjunto. El archivo no viaja aquí: esto es lo que la pantalla enseña. */
export interface Adjunto {
  readonly id: number;
  readonly filename: string;
  readonly contentType: string;
  readonly size: number;
  readonly commentId?: number;
  readonly uploadedBy?: Persona;
  readonly createdAt: string;
}

/** Lo que hizo el sistema, para la línea de tiempo. */
export interface EntradaDeHistorial {
  readonly id: number;
  readonly event: string;
  readonly fromState?: string;
  readonly toState?: string;
  readonly detail?: string;
  /** Nulo es «lo hizo el sistema». */
  readonly actor?: Persona;
  readonly createdAt: string;
}

/** Un ticket con todo lo que cuelga de él. */
export interface DetalleDeTicket {
  readonly ticket: Ticket;
  readonly comments: readonly Comentario[];
  readonly attachments: readonly Adjunto[];
  readonly history: readonly EntradaDeHistorial[];
  /**
   * Quiénes **observan** el ticket: los técnicos y desarrolladores a los que se ha etiquetado
   * (`docs/modules/tickets.md`, sección 2.3.1). **No es lo mismo que el responsable**: el ticket tiene
   * a lo sumo un responsable y muchos observadores.
   *
   * Va en la **raíz de la ficha**, al lado de los comentarios y del historial —que es lo que cuelga del
   * ticket—, y **no** en las listas paginadas: ahí sería una consulta por página para un dato que no se
   * usa.
   */
  readonly observers?: readonly Observador[];
}

/** Una página de la bandeja. */
export interface PaginaDeTickets {
  readonly tickets: readonly Ticket[];
  readonly total: number;
  readonly page: number;
  readonly perPage: number;
}

/** Los filtros de la bandeja. Vacío quiere decir «sin filtrar». */
export interface FiltrosDeBandeja {
  /** `principal`, `interno` o vacío: los dos, que es lo que ve el Administrador. */
  readonly type: string;
  readonly state: string;
  readonly q: string;
  readonly page: number;
  /**
   * **La categoría** por la que se filtra, con su identificador en texto (los filtros viajan como
   * cadenas). Vacío es «todas» (`docs/interfaz-y-experiencia.md`, sección 3.3).
   */
  readonly category?: string;
  /** **La etiqueta** por la que se filtra, tal y como se escribe. Vacío es «todas». */
  readonly tag?: string;
  /**
   * **Sólo lo mío**: lo asignado a quien mira, lo que abrió y aquello donde ha comentado
   * (`docs/modules/tickets.md`, decisión 51). Es lo que hace que la bandeja sea su bandeja.
   */
  readonly mine?: boolean;
  /**
   * **Qué parte de «lo mío»**: `assigned` (lo que tengo asignado), `watching` (lo que observo) o vacío
   * (todo lo mío). Sólo matiza cuando la lista es «lo mío» (`mine`), que es como lo sirve el backend
   * (`docs/modules/tickets.md`, decisión 62).
   */
  readonly view?: string;
}

/** Los filtros de partida. */
export const FILTROS_VACIOS: FiltrosDeBandeja = { type: '', state: '', q: '', page: 1 };

/** Lo que hace falta para dar de alta un ticket. */
export interface AltaDeTicket {
  readonly subject: string;
  readonly description: string;
  /**
   * **La categoría**, obligatoria: no hay ticket sin clasificar (`docs/modules/tickets.md`, decisión
   * 65). Se manda su identificador, no su nombre, para que renombrar la categoría no toque los tickets.
   */
  readonly categoryId: number;
  /** Las etiquetas del ticket, ya normalizadas por la pantalla. Opcionales. */
  readonly tags?: readonly string[];
  /** Sólo Soporte: el ticket es de otra persona, y se identifica por su correo. */
  readonly requesterEmail?: string;
}

/** Los tipos de ticket, para el filtro. */
export const TIPO_PRINCIPAL = 'principal';
export const TIPO_INTERNO = 'interno';

/**
 * El único servicio del módulo `tickets` del frontend.
 *
 * **Todas las llamadas del módulo pasan por aquí**, y todas van a `/api/tickets/**`: es la regla dura
 * de modularidad (`docs/arquitectura.md`, sección 4). Los componentes no usan `HttpClient`
 * directamente, así que la forma de la API se cambia en un solo sitio.
 *
 * Los errores **no se traducen aquí**: viajan como claves y la pantalla las traduce
 * (`docs/interfaz-y-experiencia.md`, sección 8).
 */
@Injectable({ providedIn: 'root' })
export class TicketsService {
  private readonly http = inject(HttpClient);

  /** La bandeja, con sus filtros y su paginación. */
  listar(filtros: FiltrosDeBandeja): Promise<PaginaDeTickets> {
    let parametros = new HttpParams().set('page', String(filtros.page));

    if (filtros.type) {
      parametros = parametros.set('type', filtros.type);
    }
    if (filtros.state) {
      parametros = parametros.set('state', filtros.state);
    }
    if (filtros.q.trim()) {
      parametros = parametros.set('q', filtros.q.trim());
    }
    // **Los dos filtros nuevos viajan de verdad**, como la vista: cuando un filtro se queda en la
    // pantalla y no en la petición, la lista enseña lo de siempre y nadie se entera
    // (`docs/interfaz-y-experiencia.md`, sección 3.3).
    if (filtros.category) {
      parametros = parametros.set('category', filtros.category);
    }
    if (filtros.tag?.trim()) {
      parametros = parametros.set('tag', filtros.tag.trim());
    }
    if (filtros.mine) {
      parametros = parametros.set('mine', '1');
    }
    // **La vista viaja de verdad**: sin esto el chip de «Asignados · Observo · Todos» cambiaba el
    // filtro en la pantalla y pedía lo mismo de siempre. Lo encontró la prueba de interfaz.
    if (filtros.view) {
      parametros = parametros.set('view', filtros.view);
    }

    return firstValueFrom(this.http.get<PaginaDeTickets>('/api/tickets', { params: parametros }));
  }

  /**
   * Vuelve a pedirle al motor los dos resúmenes del ticket.
   *
   * Va al módulo `tickets` y no a `ai`: la pantalla sólo puede hablar con su propia API
   * (`docs/modules/ai.md`, sección 5).
   */
  regenerarResumenes(numero: string): Promise<{ ticket: Ticket }> {
    return firstValueFrom(
      this.http.post<{ ticket: Ticket }>(`/api/tickets/${numero}/insights`, {}),
    );
  }

  /**
   * Añade a alguien como observador **a mano**, sin necesidad de etiquetarlo en un comentario
   * (decisión 74).
   *
   * Lo hacen Soporte y Desarrollo —los dos equipos— y sólo valen técnicos y desarrolladores activos,
   * que es la misma regla que la del etiquetado. Quien se añade así **queda igual** que quien llegó por
   * una mención: todos son observadores.
   */
  anadirObservador(numero: string, cuenta: number): Promise<{ ticket: Ticket }> {
    return firstValueFrom(
      this.http.post<{ ticket: Ticket }>(`/api/tickets/${numero}/observers`, { accountId: cuenta }),
    );
  }

  /**
   * Quita a alguien de los observadores.
   *
   * Lo puede hacer **cualquier técnico o desarrollador**, no sólo quien lo etiquetó (decisión 63), y
   * **quitar manda**: si el comentario que lo menciona se vuelve a guardar, no vuelve.
   */
  quitarObservador(numero: string, cuenta: number): Promise<{ status: string }> {
    // **Contesta `{"status":"ok"}`**, no la ficha: la pantalla recarga por su cuenta al terminar. Antes
    // estaba tipado como si devolviera el ticket, y eso era mentira (lo encontró el que hizo el
    // `POST` de al lado, comparando con el código).
    return firstValueFrom(
      this.http.delete<{ status: string }>(`/api/tickets/${numero}/observers/${cuenta}`),
    );
  }

  /** Crea un ticket. `requesterId` es de Soporte: lo crea en nombre de otra persona. */
  crear(alta: AltaDeTicket): Promise<{ ticket: Ticket }> {
    return firstValueFrom(this.http.post<{ ticket: Ticket }>('/api/tickets', alta));
  }

  /**
   * Quién puede ser responsable, por tipo de ticket.
   *
   * Lo da **este módulo** y no `users`: la pantalla sólo puede hablar con su propia API
   * (`docs/arquitectura.md`, sección 4).
   */
  responsables(): Promise<{ main: readonly Persona[]; internal: readonly Persona[] }> {
    return firstValueFrom(
      this.http.get<{ main: readonly Persona[]; internal: readonly Persona[] }>(
        '/api/tickets/assignees',
      ),
    );
  }

  /** La ficha de un ticket, con su conversación, sus adjuntos y su historial. */
  ficha(numero: string): Promise<DetalleDeTicket> {
    return firstValueFrom(this.http.get<DetalleDeTicket>(`/api/tickets/${numero}`));
  }

  // ------------------------------------------------------------------ el catálogo

  /**
   * El catálogo de categorías.
   *
   * Las ve **cualquiera que haya entrado**: el alta las necesita para poder clasificar
   * (`docs/modules/tickets.md`, sección 5). Las retiradas las devuelve el backend a quien mantiene el
   * catálogo, y aquí no se distingue: la pantalla decide qué ofrece en cada sitio.
   */
  categorias(): Promise<{ categories: readonly Categoria[] }> {
    return firstValueFrom(
      this.http.get<{ categories: readonly Categoria[] }>('/api/tickets/categories'),
    );
  }

  /** Crea una categoría. La mantienen Soporte y el Administrador. */
  crearCategoria(name: string): Promise<{ category: Categoria }> {
    return firstValueFrom(
      this.http.post<{ category: Categoria }>('/api/tickets/categories', { name }),
    );
  }

  /** Renombra una categoría. La mantienen Soporte y el Administrador. */
  renombrarCategoria(id: number, name: string): Promise<{ category: Categoria }> {
    return firstValueFrom(
      this.http.patch<{ category: Categoria }>(`/api/tickets/categories/${id}`, { name }),
    );
  }

  /**
   * Retira una categoría o la vuelve a poner.
   *
   * Va por su propia ruta y no por el `PATCH`, como el estado de una cuenta: **retirar es cosa sólo
   * del Administrador** y mezclarlo con el nombre dejaría que Soporte lo hiciera de rebote
   * (`docs/modules/tickets.md`, decisión 64).
   */
  cambiarEstadoDeCategoria(id: number, active: boolean): Promise<{ category: Categoria }> {
    return firstValueFrom(
      this.http.post<{ category: Categoria }>(`/api/tickets/categories/${id}/state`, { active }),
    );
  }

  /**
   * Las etiquetas que ya existen.
   *
   * **Dos usos con dos listados** (decisión del responsable, 2026-09-27): el alta y la ficha las piden
   * **para sugerirlas** mientras se escriben —sin `q`, las diez más usadas; con `q`, las que empiezan por
   * lo que se escribe—, y **la pantalla del catálogo las pide enteras** (`catalogoCompleto`), porque si no
   * **una etiqueta recién creada que no lleva ningún ticket no se vería**, que es justo lo que se pidió
   * poder mantener. Sin el listado completo, el tope de diez la dejaba fuera.
   */
  etiquetas(q = '', catalogoCompleto = false): Promise<{ tags: readonly Etiqueta[] }> {
    let parametros = new HttpParams();
    if (q.trim()) {
      parametros = parametros.set('q', q.trim());
    }
    if (catalogoCompleto) {
      parametros = parametros.set('all', '1');
    }

    return firstValueFrom(
      this.http.get<{ tags: readonly Etiqueta[] }>('/api/tickets/tags', { params: parametros }),
    );
  }

  /**
   * Crea una etiqueta en el catálogo.
   *
   * **No hace falta para usarla** —escribirla en un ticket la crea—, pero sí para dejarla puesta y
   * corregirla después (`docs/modules/tickets.md`, decisión 72).
   */
  crearEtiqueta(tag: string): Promise<{ tag: Etiqueta }> {
    return firstValueFrom(this.http.post<{ tag: Etiqueta }>('/api/tickets/tags', { tag }));
  }

  /**
   * Renombra una etiqueta **en todos los tickets que la llevan**.
   *
   * Es lo que la hace mantenible: una etiqueta vive en muchos tickets, y cambiar una no puede cambiar
   * sólo uno (decisión 72). La clave de la ruta es la etiqueta que hay que cambiar, no un identificador:
   * es su nombre normalizado, que es único.
   */
  renombrarEtiqueta(tag: string, nuevo: string): Promise<{ tag: Etiqueta }> {
    return firstValueFrom(
      this.http.patch<{ tag: Etiqueta }>(`/api/tickets/tags/${encodeURIComponent(tag)}`, {
        tag: nuevo,
      }),
    );
  }

  /**
   * Retira una etiqueta **y la quita de sus tickets**.
   *
   * Sólo la puede retirar el Administrador, como retirar una categoría (decisión 73).
   */
  retirarEtiqueta(tag: string): Promise<{ status: string }> {
    return firstValueFrom(
      this.http.delete<{ status: string }>(`/api/tickets/tags/${encodeURIComponent(tag)}`),
    );
  }

  /**
   * Cambia el asunto, la descripción, **la categoría o las etiquetas** del principal.
   *
   * Es el mismo `PATCH` para las cuatro cosas porque es la misma edición del ticket
   * (`docs/modules/tickets.md`, sección 5): la ficha la usa para la clasificación y el formulario de
   * edición para el texto.
   */
  editar(
    numero: string,
    cambios: {
      subject?: string;
      description?: string;
      categoryId?: number;
      tags?: readonly string[];
    },
  ): Promise<{ ticket: Ticket }> {
    return firstValueFrom(this.http.patch<{ ticket: Ticket }>(`/api/tickets/${numero}`, cambios));
  }

  /** Pone responsable. */
  asignar(numero: string, assigneeId: number): Promise<{ ticket: Ticket }> {
    return firstValueFrom(
      this.http.post<{ ticket: Ticket }>(`/api/tickets/${numero}/assign`, { assigneeId }),
    );
  }

  /**
   * Mueve el estado: `en progreso`, `en espera`, `resuelto` o `cerrado`.
   *
   * **Al cerrar, el comentario es obligatorio** (decisión 81): sin él el backend responde **422**, y con
   * él **queda como un comentario de la conversación**, que es lo que se lee para saber por qué se
   * cerró. En los demás estados es opcional.
   *
   * **`escalado` y reabrir no van por aquí**: tienen su propia acción (`escalar` y `reabrir`), porque el
   * backend no los admite como un movimiento de estado.
   */
  moverEstado(numero: string, cambio: CambioDeEstado): Promise<{ ticket: Ticket }> {
    return firstValueFrom(
      this.http.post<{ ticket: Ticket }>(`/api/tickets/${numero}/state`, cambio),
    );
  }

  /** Escala a Desarrollo. El motivo es obligatorio. */
  escalar(numero: string, reason: string): Promise<{ ticket: Ticket }> {
    return firstValueFrom(
      this.http.post<{ ticket: Ticket }>(`/api/tickets/${numero}/escalate`, { reason }),
    );
  }

  /** Reabre un ticket cerrado. */
  reabrir(numero: string): Promise<{ ticket: Ticket }> {
    return firstValueFrom(this.http.post<{ ticket: Ticket }>(`/api/tickets/${numero}/reopen`, {}));
  }

  /** Escribe un comentario. */
  comentar(numero: string, body: string): Promise<{ comment: Comentario }> {
    return firstValueFrom(
      this.http.post<{ comment: Comentario }>(`/api/tickets/${numero}/comments`, { body }),
    );
  }

  /** Cambia el texto de un comentario: sólo su autor. */
  editarComentario(numero: string, id: number, body: string): Promise<{ comment: Comentario }> {
    return firstValueFrom(
      this.http.patch<{ comment: Comentario }>(`/api/tickets/${numero}/comments/${id}`, { body }),
    );
  }

  /** Borra un comentario: se vacía el texto y queda la marca. */
  borrarComentario(numero: string, id: number): Promise<unknown> {
    return firstValueFrom(this.http.delete(`/api/tickets/${numero}/comments/${id}`));
  }

  /** Sube un adjunto, con el comentario al que pertenece si lo hay. */
  adjuntar(numero: string, archivo: File, commentId?: number): Promise<{ attachment: Adjunto }> {
    const datos = new FormData();
    datos.append('file', archivo, archivo.name);
    if (commentId) {
      datos.append('commentId', String(commentId));
    }

    return firstValueFrom(
      this.http.post<{ attachment: Adjunto }>(`/api/tickets/${numero}/attachments`, datos),
    );
  }

  /**
   * Se trae el archivo de un adjunto.
   *
   * **No se puede enlazar a la dirección del adjunto**: la descarga pasa por la comprobación de
   * permisos y necesita la cabecera de la sesión, y un `<a href>` no la lleva. Así que el archivo se
   * pide como datos y la pantalla decide si lo enseña o lo guarda.
   */
  descargar(numero: string, id: number): Promise<Blob> {
    return firstValueFrom(
      this.http.get(`/api/tickets/${numero}/attachments/${id}`, { responseType: 'blob' }),
    );
  }
}
