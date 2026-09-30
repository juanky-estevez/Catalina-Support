import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';

import { TranslationService } from '../i18n/translation.service';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Tarjeta } from '../../shared/components/tarjeta';

/**
 * Pantalla «sin permiso»: la quinta del armazón.
 *
 * Existe porque llegar a algo que no te toca tiene que **decirse**, no quedarse en blanco
 * (`docs/interfaz-y-experiencia.md`, sección 9). Se llega aquí desde una guarda —cuando el papel
 * no da— o desde una pantalla que recibe un 403 del backend.
 */
@Component({
  selector: 'app-forbidden-page',
  imports: [Aviso, Boton, Tarjeta],
  template: `
    <main class="flex min-h-dvh items-center justify-center p-4">
      <div class="w-full max-w-sm">
        <app-tarjeta [titulo]="t().sinPermiso.titulo">
          <div class="flex flex-col gap-4">
            <app-aviso forma="atencion" [texto]="t().sinPermiso.explicacion" />
            <app-boton [ancho]="true" [texto]="t().sinPermiso.volver" (pulsado)="volver()" />
          </div>
        </app-tarjeta>
      </div>
    </main>
  `,
})
export class ForbiddenPage {
  private readonly textos = inject(TranslationService);
  private readonly router = inject(Router);

  protected t() {
    return this.textos.textos();
  }

  protected volver(): void {
    void this.router.navigate(['/']);
  }
}
