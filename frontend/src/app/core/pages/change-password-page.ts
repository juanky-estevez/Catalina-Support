import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Tarjeta } from '../../shared/components/tarjeta';

/**
 * Pantalla para cambiar la contraseña propia, desde dentro.
 *
 * Exige la actual, que es lo que impide que alguien que se siente delante de un equipo desatendido
 * cambie la contraseña y se quede con la cuenta (docs/modules/auth.md, sección 4).
 */
@Component({
  selector: 'app-change-password-page',
  imports: [FormsModule, Aviso, Boton, Campo, Tarjeta],
  template: `
    <main class="flex min-h-dvh items-center justify-center p-4">
      <div class="w-full max-w-sm">
        <app-tarjeta [titulo]="t().cambio.titulo">
          <form class="flex flex-col gap-4" (ngSubmit)="cambiar()">
            <app-campo
              identificador="actual"
              tipo="password"
              autocomplete="current-password"
              [etiqueta]="t().cambio.actual"
              [(valor)]="actual"
            />

            <app-campo
              identificador="nueva-cambio"
              tipo="password"
              autocomplete="new-password"
              [etiqueta]="t().cambio.nueva"
              [ayuda]="t().cambio.ayuda"
              [(valor)]="nueva"
            />

            <app-campo
              identificador="repetir-cambio"
              tipo="password"
              autocomplete="new-password"
              [etiqueta]="t().cambio.repetir"
              [(valor)]="repetir"
            />

            @if (mensajeError(); as error) {
              <app-aviso forma="error" [texto]="error" />
            }

            @if (hecha()) {
              <app-aviso forma="exito" [texto]="t().cambio.hecha" />
            }

            <app-boton
              tipo="submit"
              [ancho]="true"
              [texto]="t().cambio.guardar"
              [textoTrabajando]="t().cambio.guardando"
              [trabajando]="cambiando()"
            />
          </form>
        </app-tarjeta>
      </div>
    </main>
  `,
})
export class ChangePasswordPage {
  private readonly textos = inject(TranslationService);
  private readonly sesion = inject(SessionService);

  protected readonly actual = signal('');
  protected readonly nueva = signal('');
  protected readonly repetir = signal('');
  protected readonly cambiando = signal(false);
  protected readonly hecha = signal(false);
  protected readonly mensajeError = signal('');

  protected t() {
    return this.textos.textos();
  }

  protected async cambiar(): Promise<void> {
    this.mensajeError.set('');
    this.hecha.set(false);

    // Lo único que se comprueba aquí es lo que el backend no puede saber: que las dos nuevas
    // coincidan. El mínimo lo dice él, con su clave.
    if (this.nueva() !== this.repetir()) {
      this.mensajeError.set(this.t().cambio.noCoinciden);
      return;
    }

    this.cambiando.set(true);
    try {
      await this.sesion.cambiarContrasena(this.actual(), this.nueva());
      this.hecha.set(true);
      this.actual.set('');
      this.nueva.set('');
      this.repetir.set('');
    } catch (error) {
      this.mensajeError.set(this.textoDelError(error));
    } finally {
      this.cambiando.set(false);
    }
  }

  private textoDelError(error: unknown): string {
    if (error instanceof HttpErrorResponse && typeof error.error?.error === 'string') {
      return this.textos.error(error.error.error);
    }
    return this.textos.error('error interno');
  }
}
