import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { TranslationService } from '../i18n/translation.service';

/** Quién ha entrado, tal y como lo cuenta el backend. */
export interface Usuario {
  readonly id: number;
  readonly name: string;
  readonly lastName: string;
  readonly email: string;
  readonly role: string;
  readonly origin: string;
  readonly language: string;
  readonly factory: boolean;
}

interface RespuestaEntrada {
  token: string;
  expiresAt: string;
  user: Usuario;
}

interface RespuestaYo {
  user: Usuario;
}

/**
 * Cómo se entra en esta instalación. Lo dice el backend, que es quien lo tiene.
 *
 * Va el **método** —`local`, `ad` o `keycloak`— y cuál de los tres caminos está abierto. Los tres
 * booleanos siguen ahí porque la pantalla de entrada no tiene por qué conocer los nombres: con el
 * método en AD el formulario es el mismo de siempre, y con el de Keycloak no hay formulario
 * (docs/modules/settings.md, sección 5.8).
 */
export interface CaminosDeEntrada {
  readonly method: string;
  readonly local: boolean;
  readonly ad: boolean;
  readonly keycloak: boolean;
}

/**
 * Lo que la propia persona puede cambiar de su cuenta.
 *
 * Es lo que devuelve `PATCH /api/users/me`, y es lo único que la sesión adopta: el identificador y el
 * de fábrica siguen siendo los que ya sabía.
 */
export type DatosDeLaCuenta = Pick<
  Usuario,
  'name' | 'lastName' | 'email' | 'role' | 'origin' | 'language'
>;

/** Dónde vive el token: en `localStorage`, porque la sesión viaja en una cabecera explícita. */
const CLAVE_TOKEN = 'catalina-support.token';

/**
 * La sesión: el token, quién ha entrado y con qué papel.
 *
 * Vive en `core` y no en el módulo `auth` porque la necesitan todos: la guarda de rutas, el
 * interceptor y el armazón. Es la excepción ya prevista para la sesión
 * (docs/arquitectura.md, sección 4, regla 5; docs/modules/auth.md, sección 9).
 *
 * **El token se lee siempre del almacenamiento, no se guarda en memoria**: así, cerrar sesión en
 * una pestaña deja sin sesión a las demás en cuanto pidan algo.
 */
@Injectable({ providedIn: 'root' })
export class SessionService {
  private readonly http = inject(HttpClient);
  private readonly textos = inject(TranslationService);

  /**
   * El token, como señal.
   *
   * **Tiene que ser una señal y no una lectura suelta**: si `hayToken` se calculara leyendo el
   * almacenamiento sin depender de nada, se quedaría con el primer valor para siempre —los
   * `computed` sólo se recalculan cuando cambia algo de lo que dependen—, y la guarda seguiría
   * diciendo «no hay sesión» después de haber entrado.
   */
  private readonly tokenActual = signal<string | null>(leerTokenDelAlmacenamiento());

  private readonly usuarioActual = signal<Usuario | null>(null);
  private readonly comprobada = signal(false);
  private readonly caducada = signal(false);

  /**
   * Los caminos de entrada de esta instalación, que se preguntan una vez.
   *
   * Los necesita la pantalla de entrada para **no ofrecer un botón que lleva a un error**: si esta
   * instalación no tiene Keycloak, no hay botón (docs/modules/auth.md, decisión 27). Se pregunta sin
   * sesión y sin bloquear a nadie: mientras no llegue la respuesta, no hay botón, que es lo mismo que
   * decir «esta instalación no lo tiene».
   */
  private readonly caminosActuales = signal<CaminosDeEntrada>({
    method: 'local',
    local: true,
    ad: false,
    keycloak: false,
  });

  /** Quién ha entrado, o nada si no ha entrado nadie. */
  readonly usuario = this.usuarioActual.asReadonly();

  /** Si ya se ha preguntado al backend quién es: es lo que evita enseñar la pantalla de entrada
   *  durante medio segundo a alguien que sí tiene sesión. */
  readonly comprobadaLaSesion = this.comprobada.asReadonly();

  /** Si hay token. No dice que valga: eso lo dice el backend. */
  readonly hayToken = computed(() => this.tokenActual() !== null);

  /** Si la sesión se acabó sin que nadie la cerrara: es lo que hace que la pantalla de entrada
   *  avise en vez de aparecer como si nada. */
  readonly sesionCaducada = this.caducada.asReadonly();

  readonly esAdministrador = computed(() => this.usuarioActual()?.role === 'administrador');
  readonly esSoporte = computed(() => {
    const papel = this.usuarioActual()?.role;
    return papel === 'soporte' || papel === 'administrador';
  });

  /**
   * El token de ahora mismo, leído **del almacenamiento cada vez**, no de memoria.
   *
   * Es lo que hace que cerrar sesión en una pestaña deje sin sesión a las demás en cuanto pidan
   * algo. Si lo que hay guardado no es lo que esta pestaña creía, se actualiza la señal: así la
   * interfaz se entera de un cambio hecho desde fuera.
   */
  token(): string | null {
    const guardado = leerTokenDelAlmacenamiento();
    if (guardado !== this.tokenActual()) {
      this.tokenActual.set(guardado);
    }
    return guardado;
  }

  /** Entra con correo y contraseña. Si no valen, el error lo cuenta la pantalla. */
  async entrar(email: string, password: string): Promise<void> {
    const respuesta = await firstValueFrom(
      this.http.post<RespuestaEntrada>('/api/auth/login', { email, password }),
    );

    this.guardarToken(respuesta.token);
    this.recibirUsuario(respuesta.user);
    this.caducada.set(false);
  }

  /**
   * Pregunta al backend quién es, que es lo que hace el frontend al arrancar.
   *
   * Si el token ya no vale, el interceptor lo borra y esta llamada falla: la guarda manda a la
   * pantalla de entrada.
   */
  async comprobar(): Promise<Usuario | null> {
    if (!this.token()) {
      this.comprobada.set(true);
      return null;
    }

    try {
      const respuesta = await firstValueFrom(this.http.get<RespuestaYo>('/api/auth/me'));
      this.recibirUsuario(respuesta.user);
      return respuesta.user;
    } catch {
      // El interceptor ya ha borrado la sesión y ha marcado que caducó.
      this.olvidar(true);
      return null;
    } finally {
      this.comprobada.set(true);
    }
  }

  /** Sale. El token no se puede revocar, así que esto es borrarlo aquí y avisar al backend. */
  async salir(): Promise<void> {
    try {
      await firstValueFrom(this.http.post('/api/auth/logout', {}));
    } catch {
      // Aunque el backend no conteste, aquí se sale: la sesión es del navegador.
    }
    this.olvidar();
  }

  /** Los caminos de entrada que tiene esta instalación. */
  readonly caminos = this.caminosActuales.asReadonly();

  /**
   * Pregunta al backend qué caminos de entrada hay.
   *
   * Es público y no lleva ningún dato de nadie; si falla, se queda con lo que ya sabía —el camino
   * local— y la pantalla no ofrece lo que no puede comprobar.
   */
  async cargarCaminos(): Promise<void> {
    try {
      const caminos = await firstValueFrom(this.http.get<CaminosDeEntrada>('/api/auth/methods'));
      this.caminosActuales.set(caminos);
    } catch {
      // Sin respuesta se queda el camino local, que es el que siempre existe.
    }
  }

  /**
   * Adopta el token que trae la vuelta de Keycloak.
   *
   * Keycloak devuelve al navegador a la pantalla de entrada **con el token en el fragmento**, que no
   * viaja al servidor; la pantalla lo lee, lo guarda aquí y borra el fragmento de la dirección
   * (docs/modules/auth.md, decisiones 13 y 30).
   */
  adoptarToken(token: string): void {
    this.guardarToken(token);
    this.caducada.set(false);
  }

  /** Pide el enlace de recuperación. La respuesta es la misma exista o no la cuenta. */
  pedirEnlace(email: string): Promise<unknown> {
    return firstValueFrom(this.http.post('/api/auth/password/forgot', { email }));
  }

  /**
   * Adopta los datos de la cuenta que acaban de cambiar.
   *
   * Es lo que hace que el nombre y el idioma del menú lateral sean los nuevos en cuanto alguien cambia
   * su perfil, sin volver a entrar (`docs/modules/users.md`, sección 8).
   *
   * Se queda con los campos que puede cambiar la propia persona: **el identificador y el de fábrica no
   * los dice el módulo `users`**, que sólo devuelve cuentas de la tabla —y la de fábrica no está en
   * ella—, así que los pone la sesión, que es quien los sabe.
   */
  actualizar(cuenta: DatosDeLaCuenta): void {
    const actual = this.usuarioActual();
    if (!actual) {
      return;
    }

    this.recibirUsuario({ ...actual, ...cuenta });
  }

  /** Establece la contraseña con el token del enlace. */
  establecerContrasena(token: string, password: string): Promise<unknown> {
    return firstValueFrom(this.http.post('/api/auth/password/reset', { token, password }));
  }

  /** Cambia la contraseña propia, con la actual. */
  cambiarContrasena(currentPassword: string, password: string): Promise<unknown> {
    return firstValueFrom(this.http.post('/api/auth/password/change', { currentPassword, password }));
  }

  /**
   * Borra la sesión de este navegador.
   *
   * `caducada` distingue las dos formas de quedarse sin sesión: salir a propósito —y entonces no
   * hay nada que explicar— o que el backend diga que ese token ya no vale.
   */
  olvidar(caducada = false): void {
    escribirToken(null);
    this.tokenActual.set(null);
    this.usuarioActual.set(null);
    this.comprobada.set(true);
    this.caducada.set(caducada);
  }

  private guardarToken(token: string): void {
    escribirToken(token);
    this.tokenActual.set(token);
  }

  private recibirUsuario(usuario: Usuario): void {
    this.usuarioActual.set(usuario);
    // El idioma de la cuenta manda dentro de la aplicación, y es también el de sus correos.
    this.textos.usarIdiomaDeCuenta(usuario.language);
  }
}

function leerTokenDelAlmacenamiento(): string | null {
  try {
    return localStorage.getItem(CLAVE_TOKEN);
  } catch {
    return null;
  }
}

function escribirToken(token: string | null): void {
  try {
    if (token === null) {
      localStorage.removeItem(CLAVE_TOKEN);
    } else {
      localStorage.setItem(CLAVE_TOKEN, token);
    }
  } catch {
    // Sin almacenamiento (modo privado, por ejemplo) la sesión no dura, pero la aplicación sigue.
  }
}
