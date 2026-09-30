import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

/** Un marcador que ese correo admite. */
export interface Marcador {
  readonly name: string;
  /** Si es de los que conviene que estén: el enlace, el número. */
  readonly essential: boolean;
}

/** Una plantilla de correo, tal y como la cuenta el backend (`docs/modules/mail.md`, sección 3). */
export interface Plantilla {
  readonly key: string;
  readonly language: string;
  readonly subject: string;
  readonly body: string;
  /** Si su texto es distinto del de fábrica. */
  readonly edited: boolean;
  readonly updatedAt?: string;
  readonly markers?: readonly Marcador[];
  /** Los marcadores imprescindibles que **no** están en el texto de hoy. */
  readonly missing?: readonly string[];
}

/** El correo ya renderizado, con sus datos de ejemplo. */
export interface VistaPrevia {
  readonly subject: string;
  readonly body: string;
  readonly text: string;
}

/** Lo que devuelven guardar y restaurar: la plantilla, y lo que le falta si le falta algo. */
interface RespuestaDePlantilla {
  readonly template: Plantilla;
  readonly missing?: readonly string[];
}

/**
 * El único servicio del módulo `mail` del frontend.
 *
 * **Todas las llamadas del módulo pasan por aquí**, y todas van a `/api/mail/**`: es la regla dura de
 * modularidad (`docs/arquitectura.md`, sección 4). Por eso la pantalla del editor vive en este módulo
 * y Configuración sólo enlaza a ella (`docs/modules/mail.md`, decisión 15).
 */
@Injectable({ providedIn: 'root' })
export class MailService {
  private readonly http = inject(HttpClient);

  /** Las veinte plantillas: diez correos en dos idiomas. */
  async listar(): Promise<readonly Plantilla[]> {
    const respuesta = await firstValueFrom(
      this.http.get<{ templates: readonly Plantilla[] }>('/api/mail/templates'),
    );

    return respuesta.templates;
  }

  /** Guarda el texto de una plantilla. El backend valida los marcadores. */
  guardar(key: string, language: string, subject: string, body: string): Promise<RespuestaDePlantilla> {
    return firstValueFrom(
      this.http.put<RespuestaDePlantilla>(`/api/mail/templates/${key}/${language}`, { subject, body }),
    );
  }

  /** Devuelve la plantilla a su texto de fábrica. */
  restaurar(key: string, language: string): Promise<RespuestaDePlantilla> {
    return firstValueFrom(
      this.http.post<RespuestaDePlantilla>(`/api/mail/templates/${key}/${language}/reset`, {}),
    );
  }

  /**
   * El correo ya renderizado, con datos de ejemplo.
   *
   * Se manda **el borrador** que se está escribiendo, y no lo guardado: la vista previa sirve para
   * decidir antes de guardar, y además avisa del marcador inventado en el momento de escribirlo.
   */
  previsualizar(key: string, language: string, subject: string, body: string): Promise<VistaPrevia> {
    return firstValueFrom(
      this.http.post<VistaPrevia>(`/api/mail/templates/${key}/${language}/preview`, { subject, body }),
    );
  }

  /** Manda una prueba **al correo de quien la pide**, que es lo único que admite el endpoint. */
  enviarPrueba(key: string, language: string): Promise<{ sentTo: string }> {
    return firstValueFrom(
      this.http.post<{ sentTo: string }>(`/api/mail/templates/${key}/${language}/test`, {}),
    );
  }
}
