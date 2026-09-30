import { Component, inject } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { SinServidor } from './core/components/sin-servidor';
import { AvailabilityService } from './core/services/availability.service';

/**
 * El armazón de la aplicación.
 *
 * Aquí **no hay ninguna regla de negocio**: es el sitio donde se pone el aviso de que el servidor
 * no responde, que puede pasar en cualquier pantalla y por eso no es de ninguna
 * (`docs/interfaz-y-experiencia.md`, sección 9). El menú lateral y los temas llegan con las
 * pantallas de producto.
 */
@Component({
  selector: 'app-root',
  imports: [RouterOutlet, SinServidor],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  private readonly disponibilidad = inject(AvailabilityService);

  protected readonly sinServidor = this.disponibilidad.sinServidor;
}
