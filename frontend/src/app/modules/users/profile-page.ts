import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Selector } from '../../shared/components/selector';
import { Tarjeta } from '../../shared/components/tarjeta';
import { claveDelError } from '../../shared/errores';
import { etiquetaDeIdioma, etiquetaDePapel, opcionesDeIdioma } from './etiquetas';
import { UsersService, type CambiosDePerfil } from './users.service';

/**
 * El perfil propio: tu nombre, tus apellidos y el idioma con el que se te escribe.
 *
 * Es de **cualquiera que haya entrado**, y se abre desde tu nombre en el menú
 * (`docs/interfaz-y-experiencia.md`, sección 3.6). Se cambia con `PATCH /api/users/me`, que acepta
 * esos tres campos y ninguno más: ni el correo, ni el papel, ni el origen, que son decisiones de otro
 * (`docs/modules/users.md`, sección 8).
 *
 * **El límite de Soporte es para editar a otros**: de su propia cuenta cambia lo mismo que cualquiera,
 * su idioma incluido. Aplicarle aquí el límite le dejaría sin poder cambiar el idioma de sus correos.
 *
 * **La cuenta de fábrica no tiene perfil**: no está en la tabla de cuentas y el endpoint responde 404.
 * La guarda de la ruta no la deja llegar hasta aquí, y su nombre en el menú no lleva a ninguna parte.
 */
@Component({
  selector: 'app-profile-page',
  imports: [RouterLink, Aviso, Boton, Campo, Selector, Tarjeta],
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
  protected readonly idioma = signal(this.sesion.usuario()?.language ?? this.textos.idioma());

  /** Lo que se ha cambiado: es lo único que se manda. */
  protected readonly cambios = computed<CambiosDePerfil>(() => {
    const quien = this.sesion.usuario();
    const cambios: { name?: string; lastName?: string; language?: string } = {};

    if (quien && this.nombre().trim() !== quien.name) {
      cambios.name = this.nombre().trim();
    }
    if (quien && this.apellidos().trim() !== quien.lastName) {
      cambios.lastName = this.apellidos().trim();
    }
    if (quien && this.idioma() !== quien.language) {
      cambios.language = this.idioma();
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

  protected opcionesDeIdioma() {
    return opcionesDeIdioma(this.t());
  }

  /** Cómo se llama el idioma elegido, en su propio idioma. */
  protected etiquetaDeIdioma(idioma: string): string {
    return etiquetaDeIdioma(idioma);
  }

  protected async guardar(): Promise<void> {
    if (!this.hayCambios()) {
      return;
    }

    this.guardando.set(true);
    this.mensaje.set(null);

    try {
      const respuesta = await this.users.cambiarPerfil(this.cambios());
      // La sesión adopta la cuenta: el menú enseña el nombre nuevo, y el idioma nuevo manda ya.
      this.sesion.actualizar(respuesta.user);
      this.nombre.set(respuesta.user.name);
      this.apellidos.set(respuesta.user.lastName);
      this.idioma.set(respuesta.user.language);
      this.mensaje.set({ forma: 'exito', texto: this.t().perfil.guardado });
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.guardando.set(false);
    }
  }
}
