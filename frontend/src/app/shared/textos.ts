/**
 * Rellenar una plantilla con sus datos: `{nombre}` se cambia por lo que se le pasa.
 *
 * Vive en `shared` y no en un módulo porque **lo usan dos**: los tickets (los avisos de la bandeja, la
 * retirada de una etiqueta, el alta) y los usuarios (la confirmación de mandar el enlace). Un módulo no
 * importa de otro —es la regla dura de modularidad—, así que lo común va aquí.
 *
 * Si una clave no viene en los datos, se deja **vacío** y no `undefined`: una pantalla no debe enseñar
 * la palabra «undefined» por un hueco que faltaba.
 */
export function interpolar(plantilla: string, datos: Record<string, string>): string {
  return plantilla.replace(/\{(\w+)\}/g, (_, clave: string) => datos[clave] ?? '');
}
