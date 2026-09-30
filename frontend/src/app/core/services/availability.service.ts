import { Injectable, signal } from '@angular/core';

/**
 * Si el backend está respondiendo.
 *
 * Existe porque «servidor caído» es una de las seis pantallas del armazón y no es de nadie en
 * concreto: cualquier petición, de cualquier módulo, puede descubrir que no hay con quién hablar
 * (docs/modules/auth.md, sección 9). Lo levanta y lo baja el interceptor, que es el único que ve
 * todas las respuestas.
 */
@Injectable({ providedIn: 'root' })
export class AvailabilityService {
  private readonly caido = signal(false);

  /** Si el servidor no está respondiendo. */
  readonly sinServidor = this.caido.asReadonly();

  /** El backend ha contestado: se quita el aviso. */
  disponible(): void {
    if (this.caido()) {
      this.caido.set(false);
    }
  }

  /** No hay respuesta, o la que hay dice que el servidor no puede atender. */
  noDisponible(): void {
    if (!this.caido()) {
      this.caido.set(true);
    }
  }
}
