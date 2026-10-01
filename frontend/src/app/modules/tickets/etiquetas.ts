export { interpolar } from '../../shared/textos';

import type { Textos } from '../../core/i18n/es';
import { opcionesConZona } from '../../shared/fechas';
import type { Categoria, Resumen } from './tickets.service';
import type { Idioma } from '../../core/i18n/translation.service';
import type { OpcionConmutador } from '../../shared/components/conmutador';
import type { OpcionSelector } from '../../shared/components/selector';

/**
 * Las etiquetas (los rótulos) del módulo `tickets`, **y las etiquetas de un ticket**.
 *
 * Los estados son **valores cerrados**, los mismos que valida el backend
 * (`docs/modules/tickets.md`, sección 2.4), y aquí están una sola vez para que la bandeja, el detalle
 * y el historial digan lo mismo.
 *
 * El nombre se comparte a propósito: en este módulo «etiqueta» es tanto el rótulo de un estado como
 * la etiqueta libre de un ticket (`red-wifi`). La segunda vive aquí porque es **una regla de la
 * pantalla** —se normaliza mientras se teclea— y tiene que poder probarse sola
 * (`docs/modules/tickets.md`, sección 2.3.2).
 */

/** Los seis estados del principal. */
export const ESTADOS_DEL_PRINCIPAL = [
  'nuevo',
  'en progreso',
  'en espera',
  'escalado',
  'resuelto',
  'cerrado',
] as const;

/** Los cinco del interno: el interno **es** la escalación, así que no tiene ese estado. */
export const ESTADOS_DEL_INTERNO = [
  'nuevo',
  'en progreso',
  'en espera',
  'resuelto',
  'cerrado',
] as const;

/** Los tipos de adjunto que se pueden ver dentro del ticket. Lo demás se descarga siempre. */
const PREVISUALIZABLES = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'pdf'];

/**
 * Cómo se llama un estado.
 *
 * **El usuario no lee jerga interna** (`docs/interfaz-y-experiencia.md`, sección 3.5): para él
 * `escalado` es «En curso», porque no tiene por qué saber que ha pasado a Desarrollo. Soporte,
 * Desarrollo y el Administrador ven el estado de verdad, que es lo que necesitan para trabajar.
 */
export function etiquetaDeEstado(textos: Textos, estado: string, esUsuario: boolean): string {
  const estados = textos.estados as Record<string, string>;

  if (esUsuario) {
    const suyos = textos.estadosDelUsuario as Record<string, string>;
    return suyos[estado] ?? estados[estado] ?? estado;
  }

  return estados[estado] ?? estado;
}

/** Los chips del estado: todos, y los que le tocan a ese tipo de ticket. */
export function opcionesDeEstado(textos: Textos, interno: boolean): readonly OpcionConmutador[] {
  const estados = interno ? ESTADOS_DEL_INTERNO : ESTADOS_DEL_PRINCIPAL;

  return [
    { valor: '', etiqueta: textos.tickets.todosLosEstados },
    ...estados.map((estado) => ({
      valor: estado,
      etiqueta: etiquetaDeEstado(textos, estado, false),
    })),
  ];
}

/** El tipo de archivo, en minúsculas y sin el punto. */
export function extensionDe(nombre: string): string {
  const partes = nombre.toLowerCase().split('.');
  return partes.length > 1 ? partes[partes.length - 1] : '';
}

/** Si ese adjunto se puede ver dentro del ticket: imágenes y PDF. */
export function sePuedePrevisualizar(nombre: string): boolean {
  return PREVISUALIZABLES.includes(extensionDe(nombre));
}

/** Un peso en algo que se lee: «203 KB», «1,2 MB». */
export function peso(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${Math.round(bytes / 1024)} KB`;
  }

  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

/**
 * Una fecha con su hora: «24/09/2026, 16:30».
 *
 * Se escribe en el idioma de quien mira, con el año entero: el formato corto del navegador deja
 * «24/9/26», que en una lista de tickets se confunde con un día de otro mes.
 *
 * **La hora es la de la instalación**, que llega en `zona` (la marca pública): un ticket guardado en
 * UTC se lee a la hora que es donde está la mesa de ayuda, y no a la de quien mira. Si `zona` no
 * llega o `Intl` no la conoce, se cae a la del navegador sin lanzar (`shared/fechas.ts`).
 */
export function fechaHora(iso: string | undefined, idioma: Idioma, zona?: string): string {
  if (!iso) {
    return '';
  }

  const momento = new Date(iso);
  if (Number.isNaN(momento.getTime())) {
    return '';
  }

  return new Intl.DateTimeFormat(
    idioma === 'en' ? 'en-GB' : 'es-ES',
    opcionesConZona(
      {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      },
      zona,
    ),
  ).format(momento);
}

/** El nombre completo de una persona, que es como se la conoce. */
export function nombreDe(persona: { name: string; lastName: string } | undefined): string {
  if (!persona) {
    return '';
  }

  return `${persona.name} ${persona.lastName}`.trim();
}

/**
 * Rellena una frase con sus datos: `{quien} cambió el estado a {estado}`.
 *
 * Las frases del historial llevan nombres y estados dentro, y traducirlas no puede ser montar el
 * texto a trozos en el código: en inglés el orden cambia. Así la frase entera vive en el diccionario y
 * aquí sólo se sustituyen los huecos.
 */
/**
 * El texto de uno de los dos resúmenes del motor de IA, en el idioma de quien mira.
 *
 * **La pantalla nunca enseña una clave ni un hueco mudo**: si el motor está escribiendo, si no está o
 * si falló, se dice con palabras (`docs/modules/ai.md`, decisiones 2 y 8). Un resumen vacío dice
 * «—», que es lo que se enseña cuando todavía no hay nada que contar.
 */
export function textoDelResumen(
  resumen: Resumen | undefined,
  idioma: Idioma,
  textos: Textos,
): string {
  if (!resumen) {
    return textos.tickets.resumenVacio;
  }

  switch (resumen.state) {
    case 'pendiente':
      return textos.tickets.resumenPendiente;
    case 'sin_motor':
      return textos.tickets.resumenSinMotor;
    case 'error':
      return textos.tickets.resumenError;
    case 'listo': {
      const texto = (idioma === 'en' ? resumen.en : resumen.es) ?? '';

      return texto.trim() || textos.tickets.resumenVacio;
    }
    default:
      return textos.tickets.resumenVacio;
  }
}

/**
 * Quién puede editar el ticket —el asunto, la descripción **y la categoría**—.
 *
 * Es la regla que la ficha reutiliza para el texto y la categoría (y por la que se eligió el patrón de
 * la asignación): **el solicitante en el suyo y Soporte**; **Desarrollo no escribe en el principal** y
 * **el Administrador no toca** (`docs/interfaz-y-experiencia.md`, sección 3.4). Un ticket **cerrado no
 * se edita** —primero se reabre— y el interno no se toca: su asunto, su descripción y su categoría son
 * los del principal.
 *
 * **Las etiquetas ya no van por aquí**: tienen su propia regla (`puedeEtiquetarElTicket`), porque las
 * ponen los dos equipos y cada uno desde su ticket (decisión 85).
 */
export function puedeEditarElTicket(opciones: {
  readonly soloLectura: boolean;
  readonly cerrado: boolean;
  readonly interno: boolean;
  readonly esSoporte: boolean;
  readonly esUsuario: boolean;
  readonly esSuyo: boolean;
}): boolean {
  if (opciones.soloLectura || opciones.cerrado || opciones.interno) {
    return false;
  }

  return opciones.esSoporte || (opciones.esUsuario && opciones.esSuyo);
}

/**
 * Quién puede **etiquetar** un ticket.
 *
 * Las etiquetas son **del principal** y el interno las hereda (decisión 67), y las ponen **los dos
 * equipos** (decisión 85): **Soporte desde el principal** y **Desarrollo desde el interno**, cada uno
 * desde el ticket en el que trabaja. El usuario no las pone ni las ve (decisión 82) y el Administrador
 * mira sin botones. Un ticket **cerrado no se etiqueta**: primero se reabre.
 */
export function puedeEtiquetarElTicket(opciones: {
  readonly soloLectura: boolean;
  readonly cerrado: boolean;
  readonly interno: boolean;
  readonly esSoporte: boolean;
  readonly esDesarrollo: boolean;
}): boolean {
  if (opciones.soloLectura || opciones.cerrado) {
    return false;
  }

  return opciones.interno ? opciones.esDesarrollo : opciones.esSoporte;
}

/** Los papeles que pueden trabajar en un ticket, para la lista de responsables. */
export function opcionesDeResponsable(
  textos: Textos,
  personas: readonly { id: number; label: string }[],
): readonly OpcionSelector[] {
  return personas.map((persona) => ({
    valor: String(persona.id),
    etiqueta: persona.label,
    grupo: textos.tickets.responsable,
  }));
}

// ------------------------------------------------------------------ las etiquetas de un ticket

/**
 * El tope de una etiqueta, en caracteres: **32**, y lo comprueba el backend con `tickets.tag.tooLong`.
 *
 * La pantalla lo aplica en el propio campo (`maxlength`), que es más amable que dejar escribir y
 * contestar con un error; el backend lo vuelve a comprobar, porque un tope que sólo vive en el
 * navegador no es un tope.
 */
export const TOPE_DE_ETIQUETA = 32;

/**
 * Normaliza una etiqueta como la escribe la gente.
 *
 * **En minúsculas, con guiones en lugar de espacios y sin acentos** (decisión 68): `Red Wifi` acaba en
 * `red-wifi`, y lo que no sea letra sin acento, número o guion se cae. Se hace **mientras se teclea**,
 * para que quien escribe vea lo que va a quedar y nadie mande `Wifi_Red` y `wifi-red` como dos cosas
 * distintas. **El backend vuelve a normalizar** —es su trabajo, y es lo que garantiza que dos puertas
 * acaben en el mismo sitio—; esta función existe para que la persona lo vea antes de mandarlo.
 *
 * Los guiones repetidos se juntan y los de los extremos se quitan: `red--wifi-` y `red-wifi` son la
 * misma etiqueta, y dejar dos formas de escribirla es justo lo que la normalización evita.
 */
export function normalizarEtiqueta(texto: string): string {
  return limpiarEtiqueta(texto)
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '');
}

/**
 * La misma limpieza, **pero sin juntar ni recortar los guiones**: es la que se aplica al teclear.
 *
 * Si al teclear se recortara el guion del final, escribir `red wifi` daría `red` —el espacio se
 * convierte en guion y el guion se cae— y la segunda palabra se pegaría a la primera: `redwifi`. Con
 * esta variante el campo enseña `red-wifi` mientras se escribe, y la forma final la da
 * `normalizarEtiqueta` al añadir la ficha.
 */
export function normalizarAlEscribir(texto: string): string {
  return limpiarEtiqueta(texto);
}

/** La limpieza común: sin acentos, en minúsculas, espacios a guiones y fuera lo que no vale. */
function limpiarEtiqueta(texto: string): string {
  return (
    texto
      .normalize('NFD')
      // Los acentos llegan como marcas sueltas después de descomponer: se quitan, no se sustituyen.
      .replace(/[\u0300-\u036f]/g, '')
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')
  );
}

/**
 * Las etiquetas que hay dentro de un texto escrito de un tirón.
 *
 * El campo de etiquetas se cierra con **Intro o coma**, y también admite pegar varias separadas por
 * espacios: esto es lo que convierte lo escrito en la lista de etiquetas de verdad.
 */
export function etiquetasDeTexto(texto: string): readonly string[] {
  return texto
    .split(/[\s,]+/)
    .map(normalizarEtiqueta)
    .filter(Boolean);
}

/**
 * Añade etiquetas a las que ya hay, **sin repetir** y normalizando lo que llegue.
 *
 * Es lo que necesitan el alta y la ficha: pulsar una sugerencia o escribir la misma dos veces no puede
 * dejar dos fichas iguales, y la base tampoco lo admitiría (`UNIQUE (ticket_id, tag)`).
 */
export function etiquetasUnicas(
  actuales: readonly string[],
  nuevas: readonly string[],
): readonly string[] {
  const resultado = [...actuales];
  const vistas = new Set(actuales);

  for (const etiqueta of nuevas) {
    const limpia = normalizarEtiqueta(etiqueta);
    if (limpia && !vistas.has(limpia)) {
      vistas.add(limpia);
      resultado.push(limpia);
    }
  }

  return resultado;
}

/**
 * Las opciones del selector de categoría, con la primera que se quiera poner delante.
 *
 * El alta le pone delante una opción vacía que **no se puede elegir**, para que la categoría sea
 * obligatoria de verdad: hasta que alguien elija, el selector dice «Elige una categoría».
 */
export function opcionesDeCategoria(
  textos: Textos,
  categorias: readonly Categoria[],
  primera?: OpcionSelector,
): readonly OpcionSelector[] {
  return [
    ...(primera ? [primera] : []),
    ...categorias.map((categoria) => ({
      valor: String(categoria.id),
      etiqueta: categoria.name,
      grupo: textos.tickets.categoria,
    })),
  ];
}
