import { Component, computed, inject } from '@angular/core';

import { TranslationService } from '../i18n/translation.service';
import { BrandService } from '../services/brand.service';
import { AvailabilityService } from '../services/availability.service';
import { HealthService } from '../services/health.service';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Tarjeta } from '../../shared/components/tarjeta';

/**
 * Aviso de «servidor caído»: la sexta pantalla del armazón.
 *
 * **No tiene ruta propia**: se pone encima de lo que hubiera en pantalla, porque puede pasar en
 * cualquier momento y en cualquier pantalla (docs/modules/auth.md, sección 9). Lo levanta el
 * interceptor cuando una petición se queda sin respuesta o el servidor contesta que no puede
 * atender.
 *
 * «Reintentar» vuelve a preguntar por la salud del backend: si contesta, el aviso se quita y se
 * sigue donde se estaba, sin recargar nada ni perder lo escrito.
 */
@Component({
  selector: 'app-sin-servidor',
  imports: [Aviso, Boton, Tarjeta],
  template: `
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-fondo/90 p-4">
      <div class="w-full max-w-sm">
        <app-tarjeta [titulo]="t().servidor.titulo">
          <div class="flex flex-col gap-4">
            <app-aviso forma="error" [texto]="explicacion()" />

            <app-boton
              [ancho]="true"
              [texto]="t().servidor.reintentar"
              [textoTrabajando]="t().comun.cargando"
              [trabajando]="comprobando()"
              (pulsado)="reintentar()"
            />
          </div>
        </app-tarjeta>
      </div>
    </div>
  `,
})
export class SinServidor {
  private readonly textos = inject(TranslationService);
  private readonly marca = inject(BrandService);
  private readonly salud = inject(HealthService);
  private readonly disponibilidad = inject(AvailabilityService);

  /**
   * El aviso, con el nombre de la instalación dentro.
   *
   * Aquí el nombre puede no saberse —esta pantalla sale justo cuando el servidor no contesta—, así
   * que si no ha llegado se usa el de fábrica: es lo que hace `BrandService.nombre`, y por eso el
   * texto no se queda con un hueco.
   */
  protected readonly explicacion = computed(() =>
    this.textos.textos().servidor.explicacion.replace('{instalacion}', this.marca.nombre()),
  );

  protected readonly comprobando = this.salud.comprobando;

  protected t() {
    return this.textos.textos();
  }

  protected async reintentar(): Promise<void> {
    await this.salud.comprobar();
    if (this.salud.estado() === 'ok') {
      this.disponibilidad.disponible();
    }
  }
}
