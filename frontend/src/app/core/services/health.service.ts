import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

/** Estado de la comprobación de salud del backend. */
export type EstadoServicio = 'comprobando' | 'ok' | 'error';

/**
 * Comprueba que el backend responde y que su base de datos contesta.
 *
 * No pertenece a ningún módulo: es una comprobación operativa del armazón, y sirve para dos cosas:
 * ver de un vistazo que la cadena navegador → nginx → backend → PostgreSQL está entera, y dar al
 * aviso de «servidor caído» un botón de reintentar que no obliga a recargar la página ni a perder
 * lo que se estuviera escribiendo (`docs/modules/auth.md`, sección 9).
 */
@Injectable({ providedIn: 'root' })
export class HealthService {
  private readonly http = inject(HttpClient);

  private readonly estadoActual = signal<EstadoServicio>('comprobando');

  /** Lo que dijo la última comprobación. */
  readonly estado = this.estadoActual.asReadonly();

  /** Si hay una comprobación en marcha: es lo que pone el botón en «trabajando». */
  readonly comprobando = computed(() => this.estadoActual() === 'comprobando');

  /**
   * Pregunta al backend cómo está y devuelve el resultado.
   *
   * Devuelve el estado en lugar de dejarlo sólo en la señal porque quien llama casi siempre tiene
   * que decidir algo con la respuesta (quitar el aviso, seguir esperando).
   */
  async comprobar(): Promise<EstadoServicio> {
    this.estadoActual.set('comprobando');

    try {
      const respuesta = await firstValueFrom(
        this.http.get<{ status: string; database: string }>('/api/health'),
      );
      const sano = respuesta.status === 'ok' && respuesta.database === 'ok';
      this.estadoActual.set(sano ? 'ok' : 'error');
    } catch {
      this.estadoActual.set('error');
    }

    return this.estadoActual();
  }
}
