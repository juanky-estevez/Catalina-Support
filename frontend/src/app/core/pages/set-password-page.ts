import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Tarjeta } from '../../shared/components/tarjeta';

/**
 * Pantalla para establecer la contraseña: la del enlace del correo, tanto el de alta como el de
 * recuperación. En los dos casos se llega sin saber la contraseña y se sale con una puesta.
 *
 * **El token llega en el fragmento de la dirección** (`#token=…`), que el navegador no manda al
 * servidor: no acaba en el registro de accesos de nginx. Aquí se lee una vez y **se borra de la
 * barra de direcciones**, para que no se quede a la vista ni en el historial
 * (docs/modules/auth.md, decisión 22).
 */
@Component({
  selector: 'app-set-password-page',
  imports: [FormsModule, RouterLink, Aviso, Boton, Campo, Tarjeta],
  template: `
    <main class="flex min-h-dvh items-center justify-center p-4">
      <div class="w-full max-w-sm">
        <app-tarjeta [titulo]="t().establecer.titulo" [descripcion]="t().establecer.explicacion">
          @if (hecha()) {
            <div class="flex flex-col gap-4">
              <app-aviso forma="exito" [texto]="t().establecer.hecha" />
              <a routerLink="/login" class="text-center text-sm text-primario underline">
                {{ t().entrada.titulo }}
              </a>
            </div>
          } @else if (!hayToken()) {
            <app-aviso forma="atencion" [texto]="t().establecer.sinToken" />
          } @else {
            <form class="flex flex-col gap-4" (ngSubmit)="guardar()">
              <app-campo
                identificador="nueva"
                tipo="password"
                autocomplete="new-password"
                [etiqueta]="t().establecer.nueva"
                [ayuda]="t().establecer.ayuda"
                [(valor)]="password"
              />

              <app-campo
                identificador="repetir"
                tipo="password"
                autocomplete="new-password"
                [etiqueta]="t().establecer.repetir"
                [(valor)]="repetir"
              />

              @if (mensajeError(); as error) {
                <app-aviso forma="error" [texto]="error" />
              }

              <app-boton
                tipo="submit"
                [ancho]="true"
                [texto]="t().establecer.guardar"
                [textoTrabajando]="t().establecer.guardando"
                [trabajando]="guardando()"
              />
            </form>
          }
        </app-tarjeta>
      </div>
    </main>
  `,
})
export class SetPasswordPage {
  private readonly textos = inject(TranslationService);
  private readonly sesion = inject(SessionService);
  private readonly router = inject(Router);

  private readonly token = signal('');

  protected readonly password = signal('');
  protected readonly repetir = signal('');
  protected readonly guardando = signal(false);
  protected readonly hecha = signal(false);
  protected readonly mensajeError = signal('');

  protected readonly hayToken = computed(() => this.token() !== '');

  constructor() {
    this.token.set(leerTokenDelFragmento());
  }

  protected t() {
    return this.textos.textos();
  }

  protected async guardar(): Promise<void> {
    this.mensajeError.set('');

    // Lo que se comprueba aquí es sólo lo que el backend no puede saber: si las dos coinciden.
    // **El mínimo de caracteres no se copia aquí**: lo dice el backend con su clave
    // (`auth.password.tooShort`), y tenerlo en dos sitios es la forma segura de que un día digan
    // cosas distintas.
    if (this.password() !== this.repetir()) {
      this.mensajeError.set(this.t().establecer.noCoinciden);
      return;
    }

    this.guardando.set(true);
    try {
      await this.sesion.establecerContrasena(this.token(), this.password());
      // El token ya está gastado: se quita de la dirección por si alguien recarga la pantalla.
      await this.router.navigate(['/set-password'], { replaceUrl: true });
      this.hecha.set(true);
    } catch (error) {
      this.mensajeError.set(this.textoDelError(error));
    } finally {
      this.guardando.set(false);
    }
  }

  private textoDelError(error: unknown): string {
    if (error instanceof HttpErrorResponse && typeof error.error?.error === 'string') {
      return this.textos.error(error.error.error);
    }
    return this.textos.error('error interno');
  }
}

/**
 * Saca el token del fragmento de la dirección y **lo borra**.
 *
 * Se hace con `history.replaceState` y no con el enrutador porque el fragmento no es una ruta: es
 * un dato que se lee una vez y no debería quedarse escrito en la barra de direcciones.
 */
function leerTokenDelFragmento(): string {
  const fragmento = window.location.hash.startsWith('#') ? window.location.hash.slice(1) : '';
  if (!fragmento) {
    return '';
  }

  const parametros = new URLSearchParams(fragmento);
  const token = parametros.get('token')?.trim() ?? '';

  if (token) {
    window.history.replaceState(null, '', window.location.pathname + window.location.search);
  }

  return token;
}
