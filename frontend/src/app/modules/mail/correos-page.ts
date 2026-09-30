import { Icono } from '../../shared/components/icono';
import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { Aviso, type MensajeDePantalla } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Dialogo } from '../../shared/components/dialogo';
import { Editor, type Insercion, type TextosDelEditor } from '../../shared/components/editor';
import { claveDelError } from '../../shared/errores';
import { MailService, type Plantilla, type VistaPrevia } from './mail.service';

/** Los dos idiomas de la instalación, en el orden en que se pintan las columnas. */
const IDIOMAS = ['es', 'en'] as const;

type Idioma = (typeof IDIOMAS)[number];

/** Lo que se está escribiendo en un idioma, todavía sin guardar. */
interface Borrador {
  subject: string;
  body: string;
}

/** Lo que se espera después de la última tecla antes de pedir la vista previa. */
const ESPERA_DE_VISTA_PREVIA = 700;

/**
 * El editor de los correos: la pantalla del módulo `mail`.
 *
 * **Configuración lleva hasta aquí**, y no edita los correos ella misma: llamar a `/api/mail/**` desde
 * `settings` rompería la regla dura de modularidad (`docs/modules/mail.md`, decisión 15).
 *
 * Los diez correos a la izquierda, y el elegido a la derecha **con sus dos idiomas al lado** —en PC—
 * o apilados en móvil: es la misma carta contada dos veces, y verlas juntas es lo que permite
 * comprobar que dicen lo mismo (`docs/interfaz-y-experiencia.md`, sección 3.8).
 *
 * **La vista previa la renderiza el backend** con los datos de ejemplo de la prueba, y se pide con el
 * borrador que se está escribiendo: así avisa del marcador inventado **mientras se escribe**, y no
 * cuando el correo ya ha salido.
 */
@Component({
  selector: 'app-correos-page',
  imports: [RouterLink, Aviso, Boton, Campo, Dialogo, Editor, Icono],
  templateUrl: './correos-page.html',
})
export class CorreosPage {
  private readonly mail = inject(MailService);
  private readonly textos = inject(TranslationService);

  protected readonly plantillas = signal<readonly Plantilla[]>([]);
  protected readonly elegida = signal('ticket.created');
  protected readonly cargando = signal(true);
  protected readonly trabajando = signal(false);
  protected readonly mensaje = signal<MensajeDePantalla | null>(null);

  /** Lo que se está escribiendo, por idioma. */
  protected readonly borradores = signal<Record<string, Borrador>>({
    es: { subject: '', body: '' },
    en: { subject: '', body: '' },
  });

  /** La vista previa de cada idioma, ya renderizada por el backend. */
  protected readonly vistas = signal<Record<string, VistaPrevia | null>>({ es: null, en: null });

  /** El idioma que está esperando confirmación para volver al texto de fábrica. */
  protected readonly restaurando = signal<string>('');

  private temporizadores: Record<string, ReturnType<typeof setTimeout> | null> = { es: null, en: null };

  protected readonly idiomas = IDIOMAS;

  /** Los diez correos, sin repetir: las plantillas vienen en dos idiomas. */
  protected readonly correos = computed(() => {
    const vistos = new Set<string>();

    return this.plantillas().filter((plantilla) => {
      if (vistos.has(plantilla.key)) {
        return false;
      }
      vistos.add(plantilla.key);
      return true;
    });
  });

  /** Los marcadores de ese correo, que son los mismos en los dos idiomas. */
  protected readonly marcadores = computed<readonly Insercion[]>(() => {
    const plantilla = this.plantillaDe('es');
    if (!plantilla?.markers) {
      return [];
    }

    return plantilla.markers.map((marcador) => ({
      etiqueta: `{{${marcador.name}}}`,
      valor: `{{${marcador.name}}}`,
      ayuda: marcador.essential
        ? `${marcador.name} · ${this.t().correos.marcadorImprescindible}`
        : marcador.name,
    }));
  });

  /** Si esa plantilla está editada: lo dice el backend, comparando el texto. */
  protected readonly editada = computed(() => this.plantillaDe('es')?.edited === true);

  constructor() {
    void this.cargar();
  }

  protected t() {
    return this.textos.textos();
  }

  /** Cómo se llama un correo en la interfaz: «Ticket nuevo, al técnico». */
  protected nombre(key: string): string {
    const nombres = this.t().plantillas as Record<string, string>;
    return nombres[key] ?? key;
  }

  /** El nombre del idioma, en su propio idioma. */
  protected idioma(idioma: string): string {
    return idioma === 'en' ? 'English' : 'Español';
  }

  protected marcador(idioma: string): string {
    return this.t().correos[idioma === 'en' ? 'idiomaEn' : 'idiomaEs'];
  }

  protected asuntoDe(idioma: string): string {
    return this.borradores()[idioma]?.subject ?? '';
  }

  protected cuerpoDe(idioma: string): string {
    return this.borradores()[idioma]?.body ?? '';
  }

  protected vistaDe(idioma: string): VistaPrevia | null {
    return this.vistas()[idioma] ?? null;
  }

  /** Lo que le falta a ese idioma: los marcadores imprescindibles que no están en su texto. */
  protected falta(idioma: string): string {
    const plantilla = this.plantillaDe(idioma);
    if (!plantilla?.missing?.length) {
      return '';
    }

    return plantilla.missing.join(', ');
  }

  protected textosDelEditor(): TextosDelEditor {
    return {
      barra: this.t().correos.barraDeFormato,
      negrita: this.t().correos.negrita,
      cursiva: this.t().correos.cursiva,
      lista: this.t().correos.lista,
      enlace: this.t().correos.enlace,
      vistaPrevia: this.t().correos.vistaPrevia,
    };
  }

  /** Cambia de correo: se traen sus dos idiomas y sus dos vistas previas. */
  protected async elegir(key: string): Promise<void> {
    this.elegida.set(key);
    this.mensaje.set(null);
    this.ponerBorradores();
    await this.pedirVistas();
  }

  /** Escribe en un idioma, y pide la vista previa cuando se deja de teclear. */
  protected escribir(idioma: string, campo: 'subject' | 'body', valor: string): void {
    this.borradores.update((borradores) => ({
      ...borradores,
      [idioma]: { ...borradores[idioma], [campo]: valor },
    }));

    if (this.temporizadores[idioma]) {
      clearTimeout(this.temporizadores[idioma]!);
    }

    this.temporizadores[idioma] = setTimeout(() => void this.pedirVista(idioma), ESPERA_DE_VISTA_PREVIA);
  }

  protected async guardar(idioma: string): Promise<void> {
    const borrador = this.borradores()[idioma];
    if (!borrador) {
      return;
    }

    this.trabajando.set(true);
    this.mensaje.set(null);

    try {
      const respuesta = await this.mail.guardar(
        this.elegida(),
        idioma,
        borrador.subject,
        borrador.body,
      );

      this.ponerPlantilla(respuesta.template, respuesta.missing);
      this.mensaje.set({ forma: 'exito', texto: this.t().correos.guardado });
      await this.pedirVista(idioma);
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.trabajando.set(false);
    }
  }

  /** Volver al de fábrica se pregunta antes: el texto que hay ahora se pierde. */
  protected async restaurar(): Promise<void> {
    const idioma = this.restaurando();
    if (!idioma) {
      return;
    }

    this.restaurando.set('');
    this.trabajando.set(true);
    this.mensaje.set(null);

    try {
      const respuesta = await this.mail.restaurar(this.elegida(), idioma);
      this.ponerPlantilla(respuesta.template, respuesta.missing);
      this.mensaje.set({ forma: 'exito', texto: this.t().correos.restaurarHecho });
      await this.pedirVista(idioma);
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.trabajando.set(false);
    }
  }

  /** Manda una prueba **a tu propio correo**: es lo único que admite el endpoint. */
  protected async enviarPrueba(idioma: string): Promise<void> {
    this.trabajando.set(true);
    this.mensaje.set(null);

    try {
      const respuesta = await this.mail.enviarPrueba(this.elegida(), idioma);
      this.mensaje.set({
        forma: 'exito',
        texto: this.t().correos.pruebaEnviada.replace('{correo}', respuesta.sentTo),
      });
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.trabajando.set(false);
    }
  }

  private plantillaDe(idioma: string): Plantilla | undefined {
    return this.plantillas().find(
      (plantilla) => plantilla.key === this.elegida() && plantilla.language === idioma,
    );
  }

  private ponerPlantilla(plantilla: Plantilla, missing: readonly string[] | undefined): void {
    this.plantillas.update((plantillas) =>
      plantillas.map((actual) =>
        actual.key === plantilla.key && actual.language === plantilla.language
          ? { ...plantilla, missing: missing ?? plantilla.missing }
          : actual,
      ),
    );

    this.borradores.update((borradores) => ({
      ...borradores,
      [plantilla.language]: { subject: plantilla.subject, body: plantilla.body },
    }));
  }

  private ponerBorradores(): void {
    const borradores: Record<string, Borrador> = {};

    for (const idioma of IDIOMAS) {
      const plantilla = this.plantillaDe(idioma);
      borradores[idioma] = {
        subject: plantilla?.subject ?? '',
        body: plantilla?.body ?? '',
      };
    }

    this.borradores.set(borradores);
  }

  private async pedirVistas(): Promise<void> {
    for (const idioma of IDIOMAS) {
      await this.pedirVista(idioma);
    }
  }

  /**
   * Pide la vista previa del borrador que hay ahora.
   *
   * Si el borrador usa un marcador que ese correo no admite, el backend contesta con su clave y aquí
   * se enseña el aviso: **es el mismo aviso que se vería al guardar, pero antes de guardar**.
   */
  private async pedirVista(idioma: string): Promise<void> {
    const borrador = this.borradores()[idioma];
    if (!borrador) {
      return;
    }

    try {
      const vista = await this.mail.previsualizar(
        this.elegida(),
        idioma,
        borrador.subject,
        borrador.body,
      );

      this.vistas.update((vistas) => ({ ...vistas, [idioma]: vista }));
    } catch (error) {
      this.vistas.update((vistas) => ({ ...vistas, [idioma]: null }));
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    }
  }

  private async cargar(): Promise<void> {
    this.cargando.set(true);

    try {
      this.plantillas.set(await this.mail.listar());
      this.ponerBorradores();
      await this.pedirVistas();
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.cargando.set(false);
    }
  }
}
