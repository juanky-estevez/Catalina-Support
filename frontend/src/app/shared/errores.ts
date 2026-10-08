import { HttpErrorResponse } from '@angular/common/http';

/**
 * La clave de error que ha mandado el backend.
 *
 * Los errores viajan **como claves** —`{"error": "users.email.duplicated"}`— y la pantalla las traduce
 * con `TranslationService.error` (`docs/interfaz-y-experiencia.md`, sección 8). Esto es lo único que
 * hay que hacer antes: sacar la clave de la respuesta.
 *
 * Si lo que llegó no es una clave nuestra —un corte de red, un 502 de nginx, un `500` sin cuerpo— se
 * responde la clave genérica: **una pantalla nunca se queda sin decir nada**, y tampoco enseña un
 * texto técnico.
 */
export function claveDelError(error: unknown): string {
  if (error instanceof HttpErrorResponse && typeof error.error?.error === 'string') {
    return error.error.error;
  }

  return 'error interno';
}

export function memoriaDelError(error: unknown): { requiredBytes: number; availableBytes: number } | null {
  if (!(error instanceof HttpErrorResponse)) return null;
  const requiredBytes = error.error?.requiredBytes;
  const availableBytes = error.error?.availableBytes;
  if (typeof requiredBytes !== 'number' || requiredBytes < 0 ||
      typeof availableBytes !== 'number' || availableBytes < 0) return null;
  return { requiredBytes, availableBytes };
}
