import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { HealthService } from './core/services/health.service';

/**
 * Armazón de la aplicación: cabecera, contenido y pie.
 *
 * El contenido lo pone cada módulo a través del enrutador; aquí no hay ninguna regla de
 * negocio (docs/arquitectura.md, sección 6).
 */
@Component({
  selector: 'app-root',
  imports: [RouterOutlet],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App implements OnInit {
  private readonly health = inject(HealthService);

  protected readonly title = signal('Catalina Support');
  protected readonly estadoServicio = this.health.estado;

  ngOnInit(): void {
    this.health.comprobar();
  }
}
