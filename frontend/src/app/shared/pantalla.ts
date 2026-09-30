/**
 * Si la pantalla es ancha, medido con la consulta del navegador.
 *
 * Lo usan el armazón —para decidir si el menú se puede plegar— y la lista de usuarios, que en móvil se
 * pinta con tarjetas y no con una tabla (`docs/interfaz-y-experiencia.md`, secciones 3.1 y 7).
 *
 * **Se decide con una consulta y no con clases de CSS** para que la pantalla que no toca **no exista
 * en el documento**: una tabla y unas tarjetas ocultas a la vez duplican los identificadores, y la
 * etiqueta de un campo acaba apuntando al campo escondido. Es el mismo fallo que ya apareció con el
 * `id` del componente `campo`.
 *
 * De fábrica se responde que sí: sin navegador (en una prueba de unidad) se asume pantalla ancha, que
 * es lo que enseña todo.
 */
export function consultaDeAncho(minimo: number): MediaQueryList | null {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return null;
  }

  return window.matchMedia(`(min-width: ${minimo}px)`);
}

/** Si la pantalla llega a ese ancho ahora mismo. */
export function esPantallaAncha(minimo: number): boolean {
  return consultaDeAncho(minimo)?.matches ?? true;
}
