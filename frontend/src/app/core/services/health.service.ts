import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';

/** Estado de la comprobación de salud del backend. */
export type EstadoServicio = 'comprobando' | 'ok' | 'error';

/**
 * Comprueba que el backend responde y que su base de datos contesta.
 *
 * No pertenece a ningún módulo: es una comprobación operativa del armazón, y sirve para
 * ver de un vistazo que la cadena navegador → nginx → backend → PostgreSQL está entera.
 */
@Injectable({ providedIn: 'root' })
export class HealthService {
  private readonly http = inject(HttpClient);

  readonly estado = signal<EstadoServicio>('comprobando');

  comprobar(): void {
    this.estado.set('comprobando');

    this.http.get<{ status: string; database: string }>('/api/health').subscribe({
      next: (respuesta) => {
        const sano = respuesta.status === 'ok' && respuesta.database === 'ok';
        this.estado.set(sano ? 'ok' : 'error');
      },
      error: () => this.estado.set('error'),
    });
  }
}
