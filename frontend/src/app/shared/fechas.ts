/**
 * Las fechas de la instalación.
 *
 * La instalación tiene **una zona horaria configurable** —un nombre IANA, como `America/Guayaquil`— y
 * es la que decide cómo se leen todas las fechas, en la interfaz y en los correos. **Las fechas
 * guardadas siguen en UTC**: cambiar la zona no mueve ningún ticket, sólo cambia la hora a la que se
 * lee. Sin esto, la misma actuación se vería a horas distintas según dónde estuviera quien mira, que
 * es justo lo que se arregla.
 *
 * Vive en `shared` porque lo usan los dos módulos que formatean fechas —`tickets` y `users`—, y la
 * regla de qué hacer cuando la zona no vale tiene que ser la misma en los dos.
 */

/**
 * La zona que se le puede pasar a `Intl`, o `undefined` para que use **la del navegador**.
 *
 * **La marca es pública y no puede romper la pantalla**: si el campo no llegara —`undefined`, o una
 * cadena vacía— o si `Intl` no conociera el nombre, se devuelve `undefined` y se lee con la zona del
 * navegador, que es lo que se hacía antes de que existiera este campo. Lanzar aquí dejaría la lista
 * de tickets sin pintar, y por un dato que sólo decide a qué hora se lee algo.
 */
export function zonaSegura(zona: string | undefined): string | undefined {
  const limpia = zona?.trim();
  if (!limpia) {
    return undefined;
  }

  try {
    // Es la única forma fiable de saber si `Intl` conoce el nombre: con una zona que no reconoce,
    // `Intl.DateTimeFormat` lanza `RangeError`.
    new Intl.DateTimeFormat('es', { timeZone: limpia });
    return limpia;
  } catch {
    return undefined;
  }
}

/**
 * Las opciones de `Intl.DateTimeFormat` con la zona de la instalación puesta, o **sin `timeZone`** si
 * la zona no vale —y entonces se lee con la del navegador—.
 *
 * El formato no cambia: sólo se le añade la zona.
 */
export function opcionesConZona(
  opciones: Intl.DateTimeFormatOptions,
  zona: string | undefined,
): Intl.DateTimeFormatOptions {
  const valida = zonaSegura(zona);
  return valida ? { ...opciones, timeZone: valida } : opciones;
}
