import { Component, computed, inject, signal } from '@angular/core';
import { Router } from '@angular/router';

import { TranslationService } from '../i18n/translation.service';
import {
  SetupService,
  type CorreoEscrito,
  type DirectorioDeInstalacion,
  type EstadoDeInstalacion,
  type KeycloakDeInstalacion,
} from '../services/setup.service';
import { Aviso, type FormaDeAviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Logo } from '../components/logo';
import { Selector, type OpcionSelector } from '../../shared/components/selector';
import { Tarjeta } from '../../shared/components/tarjeta';
import { claveDelError } from '../../shared/errores';
import { interpolar } from '../../shared/textos';

/** Los cuatro pasos del asistente, en orden. */
type Paso = 1 | 2 | 3 | 4;

/** Los mismos útiles de zona horaria que usa Configuración: la lista y el buscador. */
const DESFASES_FIJOS: readonly string[] = Array.from({ length: 12 }, (_, i) => i + 1).flatMap(
  (horas) => [`Etc/GMT+${horas}`, `Etc/GMT-${horas}`],
);

const ZONAS_DE_RESPALDO: readonly string[] = [
  'UTC',
  'Etc/GMT+5',
  'America/Guayaquil',
  'America/Bogota',
  'America/Lima',
  'America/Mexico_City',
  'America/New_York',
  'America/Chicago',
  'America/Los_Angeles',
  'America/Sao_Paulo',
  'Europe/London',
  'Europe/Madrid',
  'Asia/Tokyo',
  'Australia/Sydney',
];

/** Las zonas que ofrece el navegador, o la lista corta si no las sabe dar. */
function zonasDelNavegador(): readonly string[] {
  try {
    const soportado = (
      Intl as unknown as { supportedValuesOf?: (clave: string) => string[] }
    ).supportedValuesOf;
    const zonas = soportado?.('timeZone');
    if (zonas && zonas.length > 0) {
      return [...zonas, ...DESFASES_FIJOS];
    }
  } catch {
    // Un navegador que la anuncia pero falla cae a la lista corta: esto no puede romper la pantalla.
  }

  return ZONAS_DE_RESPALDO;
}

/** Un texto de zona listo para buscar: sin separadores, para que «new york» encuentre su zona. */
function paraBuscar(texto: string): string {
  return texto.toLowerCase().replace(/_/g, ' ').replace(/\//g, ' ').replace(/-/g, ' ');
}

/** Lo que contesta una acción: el aviso con su forma, como en el resto de la aplicación. */
interface Mensaje {
  readonly forma: FormaDeAviso;
  readonly texto: string;
}

/**
 * La vista de primer arranque: **lo que se pide antes de que la instalación tenga puerta**.
 *
 * Cuatro pasos —la instalación, cómo se entra, dónde está y el correo— más un resumen final. Cada
 * paso **se guarda al avanzar**, así que cerrar el navegador a medias no pierde lo hecho: al volver,
 * el `GET` trae lo guardado y los campos salen rellenos
 * (`docs/primer-arranque.md`, secciones 3 y 6).
 *
 * Vive en `core` porque **es de la instalación entera y ocurre antes de que exista ningún módulo**.
 * No lleva el armazón del menú —no hay nada que navegar todavía— y se enseña con la marca de fábrica.
 */
@Component({
  selector: 'app-setup-page',
  imports: [Aviso, Boton, Campo, Logo, Selector, Tarjeta],
  templateUrl: './setup-page.html',
})
export class SetupPage {
  private readonly setup = inject(SetupService);
  private readonly textos = inject(TranslationService);
  private readonly router = inject(Router);

  /** El paso que se está rellenando, y si ya se está en el resumen final. */
  protected readonly paso = signal<Paso>(1);
  protected readonly enResumen = signal(false);

  protected readonly cargando = signal(true);
  protected readonly guardando = signal(false);
  protected readonly mensaje = signal<Mensaje | null>(null);

  /** El estado con lo guardado: es lo que rellena los campos al volver a entrar. */
  protected readonly estado = signal<EstadoDeInstalacion | null>(null);

  // --- Paso 1 · la instalación ---
  protected readonly nombre = signal('');
  protected readonly idioma = signal('es');

  // --- Paso 2 · cómo se entra ---
  protected readonly metodo = signal('local');
  protected readonly dirServidor = signal('');
  protected readonly dirPuerto = signal('');
  protected readonly dirTls = signal(false);
  protected readonly dirCuenta = signal('');
  protected readonly dirContrasena = signal('');
  protected readonly dirBase = signal('');
  protected readonly dirFiltro = signal('');
  protected readonly dirCorreo = signal('');
  protected readonly dirNombre = signal('');
  protected readonly dirApellidos = signal('');
  protected readonly dirId = signal('');
  protected readonly kcEmisor = signal('');
  protected readonly kcCliente = signal('');
  protected readonly kcSecreto = signal('');
  protected readonly kcVuelta = signal('');

  // --- Paso 3 · dónde está ---
  protected readonly zonaElegida = signal('UTC');
  protected readonly busquedaZona = signal('');
  protected readonly direccionPublica = signal('');
  protected readonly zonas = signal<readonly string[]>(zonasDelNavegador());

  // --- Paso 4 · el correo ---
  protected readonly correoHost = signal('');
  protected readonly correoPuerto = signal('');
  protected readonly correoSecure = signal(false);
  protected readonly correoUsuario = signal('');
  protected readonly correoContrasena = signal('');
  protected readonly correoNombre = signal('');
  protected readonly correoCorreo = signal('');

  /** Las zonas que pasan el buscador. Sin texto se enseña la lista entera, dentro de su tope. */
  protected readonly zonasFiltradas = computed(() => {
    const texto = this.busquedaZona().trim();
    if (texto === '') {
      return this.zonas();
    }

    const buscado = paraBuscar(texto);
    return this.zonas().filter((zona) => paraBuscar(zona).includes(buscado));
  });

  /**
   * La hora que es ahora en la zona elegida, con su desfase. **Si la zona no vale, no se enseña
   * nada**: una zona rara no puede dejar la pantalla a medias.
   */
  protected readonly horaDeLaZona = computed(() => {
    const zona = this.zonaElegida().trim();
    if (zona === '') {
      return '';
    }

    try {
      const ahora = new Date();
      const idioma = this.textos.idioma();
      const hora = new Intl.DateTimeFormat(idioma, {
        hour: '2-digit',
        minute: '2-digit',
        timeZone: zona,
      }).format(ahora);
      const desfase =
        new Intl.DateTimeFormat(idioma, { timeZoneName: 'shortOffset', timeZone: zona })
          .formatToParts(ahora)
          .find((parte) => parte.type === 'timeZoneName')?.value ?? '';

      return desfase === ''
        ? ''
        : interpolar(this.t().configuracion.horaDeLaZona, { hora, desfase });
    } catch {
      return '';
    }
  });

  /** Si la dirección está puesta y **no va cifrada**: se avisa, sin bloquear nada. */
  protected readonly direccionInsegura = computed(() => {
    const direccion = this.direccionPublica().trim();
    return direccion !== '' && !direccion.toLowerCase().startsWith('https://');
  });

  /** Si el correo saliente tiene servidor: es lo que decide qué se dice en el resumen. */
  protected readonly correoPuesto = computed(() => this.correoHost().trim() !== '');

  constructor() {
    void this.cargar();
  }

  protected t() {
    return this.textos.textos();
  }

  /** Los cuatro pasos, para la lista de avance. */
  protected pasos(): readonly { numero: Paso; titulo: string }[] {
    const t = this.t().instalacion;
    return [
      { numero: 1, titulo: t.paso1 },
      { numero: 2, titulo: t.paso2 },
      { numero: 3, titulo: t.paso3 },
      { numero: 4, titulo: t.paso4 },
    ];
  }

  /** «Paso 2 de 4». */
  protected pasoDeCuatro(): string {
    return interpolar(this.t().instalacion.pasoDeCuatro, {
      paso: String(this.paso()),
      total: '4',
    });
  }

  /** Los dos idiomas de la instalación. */
  protected opcionesDeIdioma(): readonly OpcionSelector[] {
    return [
      { valor: 'es', etiqueta: this.t().idioma.es, grupo: this.t().configuracion.idioma },
      { valor: 'en', etiqueta: this.t().idioma.en, grupo: this.t().configuracion.idioma },
    ];
  }

  /**
   * Los tres métodos, **todos elegibles**.
   *
   * En Configuración un método que no está configurado se enseña apagado, porque lo que hay puesto
   * es uno y cambiarlo a ciegas dejaría la instalación sin puerta. Aquí es al revés: se está
   * configurando desde cero, y hay que poder elegir AD o Keycloak para rellenar sus campos después.
   */
  protected opcionesDeMetodo(): readonly OpcionSelector[] {
    return [
      {
        valor: 'local',
        etiqueta: this.t().configuracion.metodoLocal,
        grupo: this.t().configuracion.metodo,
      },
      {
        valor: 'ad',
        etiqueta: this.t().configuracion.metodoAD,
        grupo: this.t().configuracion.metodo,
      },
      {
        valor: 'keycloak',
        etiqueta: this.t().configuracion.metodoKeycloak,
        grupo: this.t().configuracion.metodo,
      },
    ];
  }

  /** La explicación del método elegido. */
  protected ayudaDelMetodo(): string {
    switch (this.metodo()) {
      case 'ad':
        return this.t().configuracion.metodoADAyuda;
      case 'keycloak':
        return this.t().configuracion.metodoKeycloakAyuda;
      default:
        return this.t().configuracion.metodoLocalAyuda;
    }
  }

  /** El nombre del método, para el resumen. */
  protected nombreDelMetodo(): string {
    switch (this.metodo()) {
      case 'ad':
        return this.t().configuracion.metodoAD;
      case 'keycloak':
        return this.t().configuracion.metodoKeycloak;
      default:
        return this.t().configuracion.metodoLocal;
    }
  }

  /** El nombre del idioma, para el resumen. */
  protected nombreDelIdioma(): string {
    return this.idioma() === 'en' ? this.t().idioma.en : this.t().idioma.es;
  }

  /** Elige una zona de la lista y limpia el buscador. */
  protected elegirZona(zona: string): void {
    this.zonaElegida.set(zona);
    this.busquedaZona.set('');
  }

  /** Guarda el paso actual y avanza. Al cuarto, el siguiente sitio es el resumen. */
  protected async siguiente(): Promise<void> {
    const guardado = await this.guardarPaso(this.paso());
    if (!guardado) {
      return;
    }

    // Si al guardar resulta que la instalación ya estaba sellada —alguien la terminó desde otro
    // sitio—, no se sigue configurando: se va a la entrada.
    if (this.estado()?.installed) {
      await this.router.navigate(['/login']);
      return;
    }

    if (this.paso() === 4) {
      this.enResumen.set(true);
      return;
    }

    this.paso.set((this.paso() + 1) as Paso);
  }

  /** Vuelve atrás sin guardar: lo escrito se queda donde está. */
  protected anterior(): void {
    if (this.enResumen()) {
      this.enResumen.set(false);
      return;
    }

    if (this.paso() > 1) {
      this.paso.set((this.paso() - 1) as Paso);
    }
  }

  /** Sella la instalación y lleva a la entrada, que es donde se entra ya de verdad. */
  protected async terminar(): Promise<void> {
    this.guardando.set(true);
    this.mensaje.set(null);

    try {
      await this.setup.terminar();
      await this.router.navigate(['/login']);
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.guardando.set(false);
    }
  }

  /** Carga el estado —ya preguntado por la guarda— y rellena los campos con lo guardado. */
  private async cargar(): Promise<void> {
    this.cargando.set(true);

    try {
      this.ponerEstado(await this.setup.estadoDeInstalacion());
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
    } finally {
      this.cargando.set(false);
    }
  }

  /** Guarda un paso y deja el estado nuevo puesto. Devuelve si salió bien. */
  private async guardarPaso(paso: Paso): Promise<boolean> {
    this.guardando.set(true);
    this.mensaje.set(null);

    try {
      const estado = await (paso === 1
        ? this.setup.guardarInstalacion({
            name: this.nombre().trim(),
            language: this.idioma(),
          })
        : paso === 2
          ? this.setup.guardarEntrada({
              entryMethod: this.metodo(),
              directory: this.directorioEscrito(),
              keycloak: this.keycloakEscrito(),
            })
          : paso === 3
            ? this.setup.guardarUbicacion({
                timeZone: this.zonaElegida().trim(),
                publicAppUrl: this.direccionPublica().trim(),
              })
            : this.setup.guardarCorreo({ mail: this.correoEscrito() }));

      this.ponerEstado(estado, false);
      return true;
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textos.error(claveDelError(error)) });
      return false;
    } finally {
      this.guardando.set(false);
    }
  }

  /** Lo que se ha escrito del directorio, sin los campos que la API no acepta de vuelta. */
  private directorioEscrito(): Partial<DirectorioDeInstalacion> & { bindPassword?: string } {
    return {
      host: this.dirServidor().trim(),
      port: this.dirPuerto().trim(),
      useTls: this.dirTls(),
      bindDn: this.dirCuenta().trim(),
      searchBase: this.dirBase().trim(),
      userFilter: this.dirFiltro().trim(),
      attrEmail: this.dirCorreo().trim(),
      attrName: this.dirNombre().trim(),
      attrLastName: this.dirApellidos().trim(),
      attrId: this.dirId().trim(),
      // Vacío quiere decir «no la cambies»: la contraseña guardada no sale nunca por la API.
      bindPassword: this.dirContrasena(),
    };
  }

  private keycloakEscrito(): Partial<KeycloakDeInstalacion> & { clientSecret?: string } {
    return {
      issuer: this.kcEmisor().trim(),
      clientId: this.kcCliente().trim(),
      redirectUri: this.kcVuelta().trim(),
      clientSecret: this.kcSecreto(),
    };
  }

  private correoEscrito(): CorreoEscrito {
    return {
      host: this.correoHost().trim(),
      port: this.correoPuerto().trim(),
      secure: this.correoSecure(),
      user: this.correoUsuario().trim(),
      password: this.correoContrasena(),
      fromName: this.correoNombre().trim(),
      fromEmail: this.correoCorreo().trim(),
    };
  }

  /**
   * Deja los campos con lo que hay guardado.
   *
   * **Las contraseñas se dejan vacías a propósito**: lo que llega es «hay una puesta», nunca el
   * valor. Escribir en esos campos es cambiarla; dejarlos vacíos es no tocarla.
   *
   * Con `conservarEscrito` en falso —lo normal al guardar— los campos se rellenan con lo que
   * contestó el backend, que es la verdad de lo guardado.
   */
  private ponerEstado(estado: EstadoDeInstalacion, conservarEscrito = true): void {
    this.estado.set(estado);

    if (conservarEscrito) {
      this.nombre.set(estado.name);
      this.idioma.set(estado.language || 'es');
    }

    this.metodo.set(estado.entryMethod || 'local');
    this.dirServidor.set(estado.directory.host);
    this.dirPuerto.set(estado.directory.port);
    this.dirTls.set(estado.directory.useTls);
    this.dirCuenta.set(estado.directory.bindDn);
    this.dirContrasena.set('');
    this.dirBase.set(estado.directory.searchBase);
    this.dirFiltro.set(estado.directory.userFilter);
    this.dirCorreo.set(estado.directory.attrEmail);
    this.dirNombre.set(estado.directory.attrName);
    this.dirApellidos.set(estado.directory.attrLastName);
    this.dirId.set(estado.directory.attrId);
    this.kcEmisor.set(estado.keycloak.issuer);
    this.kcCliente.set(estado.keycloak.clientId);
    this.kcSecreto.set('');
    this.kcVuelta.set(estado.keycloak.redirectUri);
    this.zonaElegida.set(estado.timeZone || 'UTC');
    this.direccionPublica.set(estado.publicAppUrl);
    this.correoHost.set(estado.mail.host);
    this.correoPuerto.set(estado.mail.port);
    this.correoSecure.set(estado.mail.secure);
    this.correoUsuario.set(estado.mail.user);
    this.correoContrasena.set('');
    this.correoNombre.set(estado.mail.fromName);
    this.correoCorreo.set(estado.mail.fromEmail);
  }
}
