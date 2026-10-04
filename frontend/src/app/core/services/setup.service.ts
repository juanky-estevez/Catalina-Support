import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

/**
 * La configuración del directorio tal y como la ve el asistente: **sin la contraseña**, con
 * `passwordSet` diciendo si hay una guardada.
 */
export interface DirectorioDeInstalacion {
  readonly host: string;
  readonly port: string;
  readonly useTls: boolean;
  readonly bindDn: string;
  readonly searchBase: string;
  readonly userFilter: string;
  readonly attrEmail: string;
  readonly attrName: string;
  readonly attrLastName: string;
  readonly attrId: string;
  readonly passwordSet: boolean;
}

/** Lo mismo con Keycloak: el secreto no sale de aquí, sólo si hay uno puesto. */
export interface KeycloakDeInstalacion {
  readonly issuer: string;
  readonly internalIssuer?: string;
  readonly clientId: string;
  readonly redirectUri: string;
  readonly secretSet: boolean;
}

/** El correo saliente, **sin su contraseña**: en su lugar dice si hay una guardada. */
export interface CorreoDeInstalacion {
  readonly host: string;
  readonly port: string;
  readonly secure: boolean;
  readonly user: string;
  readonly fromName: string;
  readonly fromEmail: string;
  readonly passwordSet: boolean;
}

/**
 * El estado de la instalación tal y como lo cuenta `GET /api/setup`
 * (`docs/primer-arranque.md`, sección 2).
 *
 * **Sin secretos**: de la contraseña del directorio, del secreto de Keycloak y de la contraseña del
 * correo sólo llega si hay una puesta, nunca el valor. Y **sin la contraseña de la cuenta de
 * fábrica**: vive en `ADMIN_PASSWORD`, del entorno, y el asistente sólo cuenta de dónde sale.
 */
export interface EstadoDeInstalacion {
  readonly installed: boolean;
  readonly name: string;
  readonly language: string;
  readonly entryMethod: string;
  readonly timeZone: string;
  readonly publicAppUrl: string;
  readonly directory: DirectorioDeInstalacion;
  readonly keycloak: KeycloakDeInstalacion;
  readonly mail: CorreoDeInstalacion;
  readonly aiAvailable: boolean;
}

/**
 * Lo que se da por sabido si `GET /api/setup` **no contesta**.
 *
 * **Un fallo de red no puede dejar la aplicación sin puerta**: si la instalación no contesta, se
 * sigue por el camino de siempre —la entrada—, que es el que existía antes de que hubiera
 * asistente. Por eso aquí `installed` es `true`: sin respuesta, nadie va a `/setup`.
 */
const SIN_RESPUESTA: EstadoDeInstalacion = {
  installed: true,
  name: '',
  language: 'es',
  entryMethod: 'local',
  timeZone: 'UTC',
  publicAppUrl: '',
  directory: {
    host: '',
    port: '',
    useTls: false,
    bindDn: '',
    searchBase: '',
    userFilter: '',
    attrEmail: '',
    attrName: '',
    attrLastName: '',
    attrId: '',
    passwordSet: false,
  },
  keycloak: { issuer: '', clientId: '', redirectUri: '', secretSet: false },
  mail: {
    host: '',
    port: '',
    secure: false,
    user: '',
    fromName: '',
    fromEmail: '',
    passwordSet: false,
  },
  aiAvailable: false,
};

/** El paso 1 del asistente: la instalación. Los cuatro pasos comparten forma y van uno a uno. */
export interface Paso1 {
  readonly name: string;
  readonly language: string;
}

export interface Paso2 {
  readonly entryMethod: string;
  readonly directory: Partial<DirectorioDeInstalacion> & { bindPassword?: string };
  readonly keycloak: Partial<KeycloakDeInstalacion> & { clientSecret?: string };
}

export interface Paso3 {
  readonly timeZone: string;
  readonly publicAppUrl: string;
}

export interface CorreoEscrito {
  readonly host: string;
  readonly port: string;
  readonly secure: boolean;
  readonly user: string;
  /** Vacío quiere decir **«no la cambies»**: la guardada no sale nunca por la API. */
  readonly password: string;
  readonly fromName: string;
  readonly fromEmail: string;
}

export interface Paso4 {
  readonly mail: CorreoEscrito;
}

/**
 * El estado de la instalación, **preguntado una vez y recordado**.
 *
 * Es lo que sostiene la guarda en las dos direcciones: quien llega a cualquier pantalla con la
 * instalación sin sellar va a `/setup`, y quien llega a `/setup` con la instalación sellada va a la
 * entrada (`docs/primer-arranque.md`, secciones 2 y 6).
 *
 * Vive en `core` y no en un módulo porque **es de la instalación entera**: lo necesitan la guarda y
 * la pantalla del primer arranque antes de que exista ningún módulo con sesión.
 */
@Injectable({ providedIn: 'root' })
export class SetupService {
  private readonly http = inject(HttpClient);

  private readonly estadoActual = signal<EstadoDeInstalacion | null>(null);

  /** La petición en curso, para que dos llamadas seguidas no pidan el estado dos veces. */
  private enCurso: Promise<EstadoDeInstalacion> | null = null;

  /** El estado que se conoce, o `null` mientras no se haya preguntado. */
  readonly estado = this.estadoActual.asReadonly();

  /**
   * El estado de la instalación, **preguntado como mucho una vez por arranque**.
   *
   * El resultado se guarda en la señal; mientras la primera petición está en vuelo, las demás
   * llamadas esperan a esa misma. Si la petición falla, se recuerda el estado «sin respuesta», que
   * es el que lleva a la entrada y no bloquea nada.
   */
  async estadoDeInstalacion(): Promise<EstadoDeInstalacion> {
    const conocido = this.estadoActual();
    if (conocido) {
      return conocido;
    }

    if (!this.enCurso) {
      this.enCurso = this.pedirEstado().finally(() => {
        this.enCurso = null;
      });
    }

    return this.enCurso;
  }

  /** Si la instalación ya está sellada. Es lo que deciden las dos guardas. */
  async estaInstalada(): Promise<boolean> {
    return (await this.estadoDeInstalacion()).installed;
  }

  /** Guarda el paso 1: el nombre y el idioma. */
  async guardarInstalacion(paso: Paso1): Promise<EstadoDeInstalacion> {
    return this.guardar('/api/setup/installation', paso);
  }

  /** Guarda el paso 2: cómo se entra y sus dos configuraciones. */
  async guardarEntrada(paso: Paso2): Promise<EstadoDeInstalacion> {
    return this.guardar('/api/setup/entry', paso);
  }

  /** Guarda el paso 3: la región horaria y la dirección pública. */
  async guardarUbicacion(paso: Paso3): Promise<EstadoDeInstalacion> {
    return this.guardar('/api/setup/location', paso);
  }

  /**
   * Guarda el paso 4: el correo saliente.
   *
   * **Se puede dejar sin poner**: si el host va vacío, el backend no guarda nada y deja seguir. La
   * instalación funciona igual, sólo que no sale ningún correo.
   */
  async guardarCorreo(paso: Paso4): Promise<EstadoDeInstalacion> {
    return this.guardar('/api/setup/mail', paso);
  }

  /** Sella la instalación: es lo que hace que el asistente no vuelva a aparecer. */
  async terminar(): Promise<EstadoDeInstalacion> {
    const estado = await firstValueFrom(
      this.http.post<EstadoDeInstalacion>('/api/setup/finish', {}),
    );

    return this.adoptar(estado);
  }

  /**
   * Prueba la conexión del paso 2 —el directorio o Keycloak, según el método— **con lo que hay
   * escrito en pantalla**, antes de guardarlo.
   *
   * No guarda nada y no devuelve estado: contesta si la prueba ha ido bien o falla con su clave,
   * igual que las pruebas de Configuración (`docs/primer-arranque.md`, sección 3).
   */
  async probarEntrada(paso: Paso2): Promise<void> {
    await firstValueFrom(this.http.post('/api/setup/entry/test', paso));
  }

  /**
   * Prueba la conexión del correo saliente del paso 4 **con lo que hay escrito en pantalla**, antes
   * de guardarlo.
   *
   * Comprueba la conexión y la autenticación y **no manda ningún correo**: en el asistente todavía no
   * hay destinatario (`docs/primer-arranque.md`, sección 3).
   */
  async probarCorreo(paso: Paso4): Promise<void> {
    await firstValueFrom(this.http.post('/api/setup/mail/test', paso));
  }

  /** Una petición que devuelve el estado nuevo: el asistente avanza con lo que le contestan. */
  private async guardar(url: string, cuerpo: object): Promise<EstadoDeInstalacion> {
    const estado = await firstValueFrom(this.http.post<EstadoDeInstalacion>(url, cuerpo));

    return this.adoptar(estado);
  }

  /** Se guarda el estado y se devuelve, para que la pantalla lo use sin volver a leerlo. */
  private adoptar(estado: EstadoDeInstalacion): EstadoDeInstalacion {
    this.estadoActual.set(estado);

    return estado;
  }

  /** Pide el estado y, si no contesta, deja el de «sin respuesta». */
  private async pedirEstado(): Promise<EstadoDeInstalacion> {
    try {
      return this.adoptar(await firstValueFrom(this.http.get<EstadoDeInstalacion>('/api/setup')));
    } catch {
      return this.adoptar(SIN_RESPUESTA);
    }
  }
}
