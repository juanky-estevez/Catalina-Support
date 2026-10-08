import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Tarjeta } from '../../shared/components/tarjeta';
import { claveDelError } from '../../shared/errores';
import { etiquetaDePapel } from './etiquetas';
import { UsersService, type CambiosDePerfil } from './users.service';

/**
 * El perfil propio: tu nombre y tus apellidos.
 *
 * Es de **cualquiera que haya entrado**, y se abre desde tu nombre en el menú
 * (`docs/interfaz-y-experiencia.md`, sección 3.6). Se cambia con `PATCH /api/users/me`, que acepta
 * esos dos campos y ninguno más: ni el correo, ni el papel, ni el origen, que son decisiones de otro
 * (`docs/modules/users.md`, sección 8).
 *
 * **La cuenta de fábrica no tiene perfil**: no está en la tabla de cuentas y el endpoint responde 404.
 * La guarda de la ruta no la deja llegar hasta aquí, y su nombre en el menú no lleva a ninguna parte.
 */
@Component({
  selector: 'app-profile-page',
  imports: [RouterLink, Aviso, Boton, Campo, Tarjeta],
  templateUrl: './profile-page.html',
})
export class ProfilePage {
  private readonly users = inject(UsersService);
  private readonly textos = inject(TranslationService);
  private readonly sesion = inject(SessionService);

  protected readonly guardando = signal(false);
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);

  protected readonly nombre = signal(this.sesion.usuario()?.name ?? '');
  protected readonly apellidos = signal(this.sesion.usuario()?.lastName ?? '');

  /** Lo que se ha cambiado: es lo único que se manda. */
  protected readonly cambios = computed<CambiosDePerfil>(() => {
    const quien = this.sesion.usuario();
    const cambios: { name?: string; lastName?: string } = {};

    if (quien && this.nombre().trim() !== quien.name) {
      cambios.name = this.nombre().trim();
    }
    if (quien && this.apellidos().trim() !== quien.lastName) {
      cambios.lastName = this.apellidos().trim();
    }

    return cambios;
  });

  protected readonly hayCambios = computed(() => Object.keys(this.cambios()).length > 0);

  protected t() {
    return this.textos.textos();
  }

  /** El correo y el papel se enseñan, pero no se tocan: se leen de la sesión, que es la que manda. */
  protected correo(): string {
    return this.sesion.usuario()?.email ?? '';
  }

  protected papel(): string {
    const quien = this.sesion.usuario();
    return quien ? etiquetaDePapel(this.t(), quien.role) : '';
  }

  protected async guardar(): Promise<void> {
    if (!this.hayCambios()) {
      return;
    }

    this.guardando.set(true);
    this.mensaje.set(null);

    try {
      const respuesta = await this.users.cambiarPerfil(this.cambios());
      // La sesión adopta la cuenta: el menú enseña el nombre nuevo de inmediato.
      this.sesion.actualizar(respuesta.user);
      this.nombre.set(respuesta.user.name);
      this.apellidos.set(respuesta.user.lastName);
      this.mensaje.set({ forma: 'exito', texto: this.t().perfil.guardado });
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.guardando.set(false);
    }
  }
}
