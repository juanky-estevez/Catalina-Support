import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

/** Un hueco del logo, tal y como lo cuenta la configuración. */
export interface HuecoDeLogo {
  readonly filled: boolean;
  readonly fileName?: string;
  readonly size?: number;
  readonly width?: number;
  readonly height?: number;
  readonly updatedAt?: string;
}

/** La configuración del directorio, tal y como viaja por la API. */
export interface Directorio {
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
  /** Si hay contraseña guardada. **El valor no llega nunca.** */
  readonly passwordSet: boolean;
}

/** La configuración de Keycloak, con la misma regla para el secreto. */
export interface Keycloak {
  readonly issuer: string;
  readonly internalIssuer?: string;
  readonly clientId: string;
  readonly redirectUri: string;
  readonly secretSet: boolean;
}

/**
 * La configuración de la instalación, tal y como la cuenta `GET /api/settings`
 * (`docs/modules/settings.md`).
 *
 * Es **un documento entero**, no un parche: se lee de una vez y se guarda entero con un `PUT`, así
 * que cada tarjeta de la pantalla manda lo suyo sobre lo que ya había.
 */
export interface Configuracion {
  /**
   * El nombre de la instalación: lo que se lee donde antes decía «Catalina Support».
   *
   * Va con **la marca** y no con la numeración: es cómo se llama esta instalación, y se enseña junto
   * al logo en la pantalla de entrada y en el menú.
   */
  readonly name: string;
  /**
   * Cómo se entra en la instalación: el método y las dos configuraciones, **sin sus secretos**.
   *
   * Lo que llega de la contraseña de la cuenta de servicio y del secreto del cliente son dos
   * booleanos —`passwordSet` y `secretSet`—, y nunca el valor: se guardan para hablar con el
   * directorio y no salen por la API.
   */
  readonly entryMethod: string;
  readonly directory: Directorio;
  readonly keycloak: Keycloak;
  readonly language: string;
  readonly primaryColor: string;
  /**
   * La zona horaria de la instalación, en nombre IANA (`America/Guayaquil`).
   *
   * Es la que decide cómo se leen todas las fechas, aquí y en los correos: las fechas guardadas siguen
   * en UTC, así que cambiarla no mueve ningún ticket, sólo cambia la hora a la que se lee.
   */
  readonly timeZone: string;
  /**
   * La dirección pública de la instalación: la base de los enlaces que salen en los correos y la
   * vuelta de Keycloak. La aplicación se navega en relativo, así que sirve cualquier dirección con la
   * que se llegue a ella.
   */
  readonly publicAppUrl: string;
  readonly numberPrefix: string;
  readonly mainAssignment: string;
  readonly mainNotification: string;
  readonly internalAssignment: string;
  readonly internalNotification: string;
  readonly updatedAt: string;
  /**
   * El motor de IA: su dirección y su modelo.
   *
   * Vacíos quieren decir «esta instalación no tiene motor»: la mesa de ayuda funciona entera sin él y
   * los dos resúmenes del ticket se quedan sin texto. Son la fuente; el entorno (`AI_URL`,
   * `AI_MODEL`) queda de respaldo (`docs/modules/ai.md`).
   */
  readonly aiUrl: string;
  readonly aiModel: string;
  readonly brand: { readonly light: HuecoDeLogo; readonly dark: HuecoDeLogo };
}

/** Lo que se ha escrito del directorio, sin los campos que la API no acepta de vuelta. */
export type DirectorioEscrito = Partial<Directorio> & { bindPassword?: string };

/** Lo mismo con Keycloak: la configuración escrita con su secreto, que sólo viaja hacia el backend. */
export type KeycloakEscrito = Partial<Keycloak> & { clientSecret?: string };

/**
 * La configuración de la instalación: **un único servicio que centraliza sus llamadas HTTP**.
 *
 * Vive en `core/services` porque la configuración **es de la instalación y no de un módulo**: es la
 * excepción que recoge `docs/arquitectura.md`, sección 4. Sigue el patrón de la casa
 * (`SetupService`): standalone, con las rutas **relativas** y **sólo** a `/api/settings/**`, y un
 * método por endpoint.
 *
 * La pantalla se queda con el estado y la interacción —qué se está escribiendo, qué tarjeta pide
 * guardar y qué mensaje se enseña—, y pide aquí cada petición. Los secretos (la contraseña del
 * directorio y el secreto de Keycloak) **sólo viajan hacia el backend**: nunca vuelven.
 *
 * **La marca no está aquí**: leer el nombre, subir o restablecer el logo y previsualizar el color
 * los sirve `BrandService`, que es de quien es el recurso (`/api/settings/brand`). Aquí queda el
 * resto de la configuración de la instalación.
 */
@Injectable({ providedIn: 'root' })
export class SettingsService {
  private readonly http = inject(HttpClient);

  /** Lee la configuración entera de la instalación. */
  async cargar(): Promise<Configuracion> {
    return firstValueFrom(this.http.get<Configuracion>('/api/settings'));
  }

  /**
   * Guarda la configuración: **es el mismo `PUT` para todas las tarjetas**.
   *
   * El documento viaja entero —no es un parche—, así que lo que cambia entre una tarjeta y otra es
   * qué campos se han tocado antes de llamar aquí, y eso lo decide la pantalla, que es quien tiene lo
   * que se está escribiendo.
   */
  async guardar(configuracion: Configuracion): Promise<Configuracion> {
    return firstValueFrom(this.http.put<Configuracion>('/api/settings', configuracion));
  }

  /**
   * Prueba la conexión con el directorio **con lo que hay en pantalla**, antes de guardarlo.
   *
   * Es lo que evita guardar una configuración que no funciona y quedarse sin poder entrar: la prueba
   * la contesta el módulo `auth`, que es el que sabe hablar con un directorio.
   */
  async probarDirectorio(directorio: DirectorioEscrito): Promise<void> {
    await firstValueFrom(this.http.post('/api/settings/directory/test', directorio));
  }

  /** Lo mismo con Keycloak: se lee el reino y se comprueba que dice dónde está su pantalla. */
  async probarKeycloak(keycloak: KeycloakEscrito): Promise<void> {
    await firstValueFrom(this.http.post('/api/settings/keycloak/test', keycloak));
  }

  /**
   * Prueba el motor de IA **con lo que hay en pantalla**: se le pregunta a su comprobación de salud.
   *
   * Si la dirección va vacía, el backend prueba **la guardada**, que es lo que permite comprobar el
   * motor ya configurado sin volver a escribirlo.
   */
  async probarMotor(url: string): Promise<void> {
    await firstValueFrom(this.http.post('/api/settings/ai/test', { url }));
  }
}
