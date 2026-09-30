import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Tarjeta } from '../../shared/components/tarjeta';

/**
 * Pantalla de olvido: se pide el correo y se confirma que se ha mandado.
 *
 * **Nunca dice si esa cuenta existe**: el backend responde lo mismo en los dos casos, y la pantalla
 * tampoco lo insinúa (docs/modules/auth.md, sección 8). Quien lee el mensaje no puede deducir si
 * trabaja aquí alguien con ese correo.
 */
@Component({
  selector: 'app-forgot-password-page',
  imports: [FormsModule, RouterLink, Aviso, Boton, Campo, Tarjeta],
  template: `
    <main class="flex min-h-dvh items-center justify-center p-4">
      <div class="w-full max-w-sm">
        <app-tarjeta [titulo]="t().olvido.titulo" [descripcion]="t().olvido.explicacion">
          @if (enviado()) {
            <div class="flex flex-col gap-4">
              <app-aviso forma="exito" [texto]="t().olvido.enviado" />
              <a routerLink="/login" class="text-center text-sm text-primario underline">
                {{ t().olvido.volver }}
              </a>
            </div>
          } @else {
            <form class="flex flex-col gap-4" (ngSubmit)="mandar()">
              <app-campo
                identificador="correo-olvido"
                tipo="email"
                autocomplete="username"
                [etiqueta]="t().olvido.correo"
                [(valor)]="email"
              />

              @if (mensajeError(); as error) {
                <app-aviso forma="error" [texto]="error" />
              }

              <app-boton
                tipo="submit"
                [ancho]="true"
                [texto]="t().olvido.enviar"
                [textoTrabajando]="t().olvido.enviando"
                [trabajando]="mandando()"
              />

              <a routerLink="/login" class="text-center text-sm text-primario underline">
                {{ t().olvido.volver }}
              </a>
            </form>
          }
        </app-tarjeta>
      </div>
    </main>
  `,
})
export class ForgotPasswordPage {
  private readonly textos = inject(TranslationService);
  private readonly sesion = inject(SessionService);

  protected readonly email = signal('');
  protected readonly mandando = signal(false);
  protected readonly enviado = signal(false);
  protected readonly mensajeError = signal('');

  protected t() {
    return this.textos.textos();
  }

  protected async mandar(): Promise<void> {
    this.mensajeError.set('');

    if (!this.email().trim()) {
      this.mensajeError.set(this.t().entrada.obligatorios);
      return;
    }

    this.mandando.set(true);
    try {
      await this.sesion.pedirEnlace(this.email().trim());
      this.enviado.set(true);
    } catch (error) {
      // Un correo que no sale no se cuenta aquí: si el backend ha podido pedirlo, la respuesta es
      // la misma a propósito. Sólo se enseña el fallo cuando la petición en sí no llegó.
      if (error instanceof HttpErrorResponse && error.status >= 400 && error.status < 500) {
        this.enviado.set(true);
        return;
      }
      this.mensajeError.set(this.textos.error('error interno'));
    } finally {
      this.mandando.set(false);
    }
  }
}
