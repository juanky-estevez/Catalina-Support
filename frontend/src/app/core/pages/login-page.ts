import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import { Controles } from '../components/controles';
import { Logo } from '../components/logo';
import { TranslationService } from '../../core/i18n/translation.service';
import { BrandService } from '../../core/services/brand.service';
import { SessionService } from '../../core/services/session.service';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Tarjeta } from '../../shared/components/tarjeta';

/**
 * Pantalla de entrada: correo y contraseña, «he olvidado mi contraseña» y el conmutador de idioma.
 *
 * Es una de las seis pantallas del armazón y vive en `core` porque la sesión no es de ningún
 * módulo (docs/modules/auth.md, sección 9).
 */
@Component({
  selector: 'app-login-page',
  imports: [FormsModule, RouterLink, Aviso, Boton, Campo, Tarjeta, Controles, Logo],
  template: `
    <main class="flex min-h-dvh items-center justify-center p-4">
      <div class="flex w-full max-w-sm flex-col items-center gap-6">
        <!--
          La marca de la instalación, arriba del formulario (interfaz-y-experiencia, 6.4).

          **El ancho se pide en el elemento, no sólo dentro** (decisión del responsable, 2026-09-28): esta
          columna centra a sus hijos, así que un componente encoge hasta lo que mida su
          contenido —medido: el logo 136 px y el formulario 329— aunque por dentro ocupe el 100 %. El ancho
          se pide aquí, que es donde se decide, y así el logo, la tarjeta y los controles cuadran.
        -->
        <app-logo class="w-full" tamano="entrada" />

        <!-- El ancho también aquí: si no, la tarjeta encoge a lo que mida su contenido y queda más
             estrecha que el logo y los controles (medido: 329 contra 380). -->
        <app-tarjeta class="w-full" [titulo]="t().entrada.titulo" [descripcion]="nombre()">
          <form class="flex flex-col gap-4" (ngSubmit)="entrar()">
            <!--
              Con el método en AD **el formulario es el mismo** —correo y contraseña— y lo que cambia
              es quién lo contesta, así que se dice de dónde es la contraseña que se espera. Con el
              método en Keycloak no hay formulario: se entra desde su pantalla.
            -->
            @if (porDirectorio()) {
              <app-aviso forma="informacion" [texto]="t().entrada.porDirectorio" />
            } @else if (!hayFormulario()) {
              <app-aviso forma="informacion" [texto]="t().entrada.soloKeycloak" />
            }

            @if (hayFormulario()) {
              <app-campo
                identificador="correo"
                tipo="email"
                autocomplete="username"
                [etiqueta]="t().entrada.correo"
                [ayuda]="t().entrada.correoAyuda"
                [(valor)]="email"
              />

              <app-campo
                identificador="contrasena"
                tipo="password"
                autocomplete="current-password"
                [etiqueta]="t().entrada.contrasena"
                [(valor)]="password"
              />
            }

            @if (mensajeError(); as error) {
              <app-aviso forma="error" [texto]="error" />
            }

            <!--
              Quien llega aquí desde /setup con la instalación ya terminada tiene que saber por qué
              no ve el asistente, en vez de encontrarse la entrada sin explicación
              (docs/primer-arranque.md, sección 6).
            -->
            @if (yaInstalada()) {
              <app-aviso forma="informacion" [texto]="t().instalacion.terminada" />
            }

            @if (sesionCaducada()) {
              <app-aviso forma="informacion" [texto]="t().sesion.caducada" />
            }

            @if (hayFormulario()) {
              <app-boton
                tipo="submit"
                [ancho]="true"
                [texto]="t().entrada.entrar"
                [textoTrabajando]="t().entrada.entrando"
                [trabajando]="entrando()"
              />
            }

            <!--
              El enlace de recuperar la contraseña **sólo con el método local**: con AD o Keycloak la
              contraseña no es de aquí, y un enlace que manda un correo para cambiarla sería mentira
              (docs/modules/auth.md, sección 5.2).
            -->
            @if (caminos().local) {
              <a routerLink="/forgot-password" class="text-center text-sm text-primario underline">
                {{ t().entrada.olvidada }}
              </a>
            }

            <!--
              El botón de Keycloak **sólo si esta instalación entra por ahí**: lo dice el backend, y
              ofrecer un botón que lleva a un error es peor que no ofrecerlo (decisión 27). No es un
              envío del formulario: es irse a Keycloak, así que es un enlace con pinta de botón
              secundario y con la navegación del navegador, que es la que tiene que llevarse las
              cookies de Keycloak.
            -->
            @if (caminos().keycloak) {
              <div class="flex flex-col gap-4 border-t border-borde pt-4">
                <p class="text-center text-xs text-apagado">{{ t().entrada.oCon }}</p>
                <app-boton
                  [ancho]="true"
                  forma="secundario"
                  [texto]="t().entrada.conKeycloak"
                  (pulsado)="entrarConKeycloak()"
                />
              </div>
            }

            <!--
              **La puerta de la cuenta de fábrica.** Con el método en Keycloak no hay formulario, y sin
              esta salida una instalación a la que se le elige mal el método se quedaría sin nadie que
              pudiera volver a cambiarlo. Es discreta a propósito: no se le ofrece a quien entra todos
              los días, pero está donde hay que buscarla (docs/usuarios-y-permisos.md, sección 8).
            -->
            @if (!caminos().local && !caminos().ad && !puertaDeFabrica()) {
              <button
                type="button"
                class="text-center text-xs text-apagado underline"
                (click)="puertaDeFabrica.set(true)"
              >
                {{ t().entrada.comoAdministrador }}
              </button>
            }

            @if (puertaDeFabrica()) {
              <app-aviso forma="informacion" [texto]="t().entrada.cuentaDeFabrica" />
            }
          </form>
        </app-tarjeta>

        <div class="mt-4 flex w-full flex-col items-center gap-3">
          <app-controles class="w-full" />

          <!--
            **La versión del sistema, en el pie de la pantalla de entrada** (decisión del responsable,
            2026-09-25): es donde alguien que no puede entrar mira para decir qué versión está usando,
            y esta pantalla es lo único que ve. No es un enlace: se lee y ya.
          -->
          @if (versionDelSistema(); as version) {
            <p class="text-xs text-apagado">{{ version }}</p>
          }
          @if (fuente(); as fuente) {
            <a
              class="text-xs text-apagado underline hover:text-texto"
              [href]="fuente.url"
              target="_blank"
              rel="noopener noreferrer"
            >{{ t().entrada.codigoFuente }} · {{ fuente.licencia }}</a>
          }
        </div>
      </div>
    </main>
  `,
})
export class LoginPage implements OnInit {
  protected readonly textos = inject(TranslationService);
  private readonly sesion = inject(SessionService);
  private readonly marca = inject(BrandService);
  private readonly router = inject(Router);
  private readonly ruta = inject(ActivatedRoute);

  protected readonly email = signal('');
  protected readonly password = signal('');
  protected readonly entrando = signal(false);
  protected readonly mensajeError = signal('');
  protected readonly sesionCaducada = this.sesion.sesionCaducada;
  protected readonly caminos = this.sesion.caminos;

  /**
   * Si se ha llegado aquí **desde `/setup` con la instalación ya terminada**.
   *
   * Lo pone la guarda del primer arranque con `?installed=1`: es lo que explica por qué esta
   * instalación no vuelve a enseñar el asistente (`docs/primer-arranque.md`, sección 6).
   */
  protected readonly yaInstalada = signal(false);

  /**
   * Si se ha pedido la puerta de la cuenta de fábrica.
   *
   * **Es la única puerta que no se puede cerrar**: con el método en Keycloak no hay formulario, y sin
   * esta salida una instalación a la que se le elige mal el método se quedaría sin nadie que pudiera
   * volver a cambiarlo. La cuenta de fábrica entra siempre, y su contraseña vive en la configuración
   * (docs/usuarios-y-permisos.md, sección 8).
   */
  protected readonly puertaDeFabrica = signal(false);

  /**
   * Si hay formulario de correo y contraseña.
   *
   * Lo hay con el método local y **también con el de AD**: es el mismo formulario, y lo que cambia es
   * quién contesta. Con el de Keycloak no lo hay —se entra desde su pantalla—, y ahí aparece sólo si
   * se pide la puerta de la cuenta de fábrica.
   */
  protected readonly hayFormulario = computed(
    () => this.caminos().local || this.caminos().ad || this.puertaDeFabrica(),
  );

  /** Si la contraseña que se espera es la de la organización y no la de aquí. */
  protected readonly porDirectorio = computed(() => this.caminos().ad);

  /**
   * El nombre de la instalación, que es lo que se lee bajo el título.
   *
   * Sale de la marca y **no del diccionario de textos**: es de la instalación, no del idioma, y una
   * empresa que se llama de una manera se llama igual en los dos idiomas
   * (docs/modules/settings.md).
   */
  protected readonly nombre = this.marca.nombre;

  /** La versión del sistema, con su `v`, en el pie de la pantalla. */
  protected readonly versionDelSistema = this.marca.versionDelSistema;
  protected readonly fuente = this.marca.fuente;

  /**
   * Si ya hay sesión, esta pantalla no tiene nada que hacer: se vuelve al inicio.
   *
   * Pasa al recargar justo después de entrar, o al abrir un enlace a `/login` teniendo sesión:
   * enseñar el formulario ahí es pedir dos veces lo mismo.
   */
  async ngOnInit(): Promise<void> {
    // Quien llega desde `/setup` con la instalación ya sellada trae la marca en la dirección: se lee
    // aquí y se le cuenta, en vez de dejarle delante de la entrada sin saber qué ha pasado.
    this.yaInstalada.set(this.ruta.snapshot.queryParamMap.get('installed') === '1');

    // La vuelta de Keycloak llega aquí, con lo que haya pasado **en el fragmento**: el token si se ha
    // entrado, y la clave del fallo si no. Se lee antes que nada y se borra de la dirección en el
    // mismo momento, para que no quede en el historial ni al recargar (decisiones 13 y 30).
    const vuelta = leerLaVuelta();
    if (vuelta) {
      this.aplicarLaVuelta(vuelta);
    }

    // Y los caminos de entrada, que es lo que decide si se ofrece el botón de Keycloak.
    if (!this.caminos().keycloak) {
      void this.sesion.cargarCaminos();
    }

    if (!this.sesion.hayToken()) {
      // Con un token recién adoptado sí hay sesión: el backend dirá quién es y con qué papel.
      if (vuelta?.token) {
        await this.sesion.comprobar();
        await this.router.navigate(['/']);
      }
      return;
    }

    if (await this.sesion.comprobar()) {
      await this.router.navigate(['/']);
    }
  }

  /**
   * Aplica lo que traía la vuelta de Keycloak: el token, o el motivo por el que no se ha entrado.
   *
   * **No se enseña la clave, se traduce** (`textos.error`), y si la clave es de las que la pantalla de
   * entrada no conoce sale el mensaje genérico en vez de un texto raro.
   */
  private aplicarLaVuelta(vuelta: VueltaDeKeycloak): void {
    if (vuelta.token) {
      this.sesion.adoptarToken(vuelta.token);
      return;
    }

    this.mensajeError.set(this.textos.error(vuelta.error ?? 'error interno'));
  }

  /**
   * Entra por Keycloak.
   *
   * **No es una llamada del servicio HTTP**: es llevar el navegador a la ruta que redirige a Keycloak,
   * porque la autenticación ocurre allí y hay que irse con la navegación de verdad.
   */
  protected entrarConKeycloak(): void {
    window.location.href = '/api/auth/keycloak/start';
  }

  /** Los textos de ahora mismo, para no repetir `textos.textos()` en cada línea. */
  protected t() {
    return this.textos.textos();
  }

  protected async entrar(): Promise<void> {
    this.mensajeError.set('');

    if (!this.email().trim() || !this.password()) {
      this.mensajeError.set(this.t().entrada.obligatorios);
      return;
    }

    this.entrando.set(true);
    try {
      await this.sesion.entrar(this.email().trim(), this.password());
      // Con la sesión puesta, a la raíz: allí decide la guarda a dónde puede ir cada papel.
      await this.router.navigate(['/']);
    } catch (error) {
      this.mensajeError.set(this.textoDelError(error));
    } finally {
      this.entrando.set(false);
    }
  }

  /**
   * Un fallo al entrar **no borra nada ni lleva a ninguna parte**: se cuenta en la propia pantalla,
   * que es donde está el formulario. El interceptor sólo actúa si la petición llevaba sesión.
   */
  private textoDelError(error: unknown): string {
    if (error instanceof HttpErrorResponse && typeof error.error?.error === 'string') {
      return this.textos.error(error.error.error);
    }
    return this.textos.error('error interno');
  }
}

/**
 * Lo que puede traer la vuelta de Keycloak en el fragmento de la dirección.
 *
 * `token` cuando se ha entrado y `error` cuando no, con la clave que traduce la pantalla. Los dos van
 * en el fragmento y no en la consulta: **el fragmento no viaja al servidor**, así que el token no
 * queda en los registros de nginx (docs/modules/auth.md, decisión 13).
 */
interface VueltaDeKeycloak {
  token?: string;
  error?: string;
}

/**
 * Lee la vuelta de Keycloak del fragmento y **lo borra de la dirección**.
 *
 * Se borra siempre, hasta cuando no hay nada que leer, y con `replaceState` para no dejar un paso de
 * más en el historial: lo que se ha leído ya está en memoria, y volver atrás no tiene que volver a
 * aplicarlo.
 */
function leerLaVuelta(): VueltaDeKeycloak | null {
  const fragmento = window.location.hash.replace(/^#/, '');
  if (!fragmento) {
    return null;
  }

  const parametros = new URLSearchParams(fragmento);
  const vuelta: VueltaDeKeycloak = {};
  if (parametros.get('token')) {
    vuelta.token = parametros.get('token')!;
  }
  if (parametros.get('error')) {
    vuelta.error = parametros.get('error')!;
  }

  window.history.replaceState(null, '', window.location.pathname + window.location.search);

  return vuelta.token || vuelta.error ? vuelta : null;
}
