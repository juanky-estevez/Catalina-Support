import { HttpClient } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { TranslationService } from '../i18n/translation.service';
import { BrandService, logoDeFabrica, type Marca } from '../services/brand.service';
import { SessionService } from '../services/session.service';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Campo } from '../../shared/components/campo';
import { Selector, type OpcionSelector } from '../../shared/components/selector';
import { Tarjeta } from '../../shared/components/tarjeta';
import { claveDelError } from '../../shared/errores';
import { interpolar } from '../../shared/textos';

/**
 * La lista corta de zonas horarias, **sólo si el navegador no sabe darlas**.
 *
 * `Intl.supportedValuesOf('timeZone')` existe en los navegadores modernos y trae el catálogo IANA
 * entero, que es lo que se quiere. Un navegador que no lo traiga —o que lo anuncie y falle— se queda
 * con esta docena de zonas habituales, elegidas para que las horas de la casa estén: `America/Guayaquil`
 * y `Etc/GMT+5`, que es su desfase. **La pantalla no se queda sin poder configurarse** por eso.
 */
/** Los desfases fijos, que el catálogo del navegador no ofrece: `Etc/GMT+5` es UTC-5. */
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

/**
 * Las zonas que ofrece el navegador, o la lista corta si no las sabe dar.
 *
 * Se comprueba que exista **y que devuelva algo**: hay entornos que exponen la función y contestan
 * vacío, y una lista vacía dejaría el campo sin nada que elegir. En cualquier duda, la lista corta.
 */
function zonasDelNavegador(): readonly string[] {
  try {
    const soportado = (
      Intl as unknown as { supportedValuesOf?: (clave: string) => string[] }
    ).supportedValuesOf;
    const zonas = soportado?.('timeZone');
    if (zonas && zonas.length > 0) {
      // **Los desfases fijos se añaden a mano**: `Intl.supportedValuesOf` sólo trae zonas canónicas y
      // deja fuera `Etc/GMT±N`, que es como se expresa «UTC-5» sin horario de verano. Se ofrecen los
      // doce de cada lado, y la pantalla enseña el desfase de cada uno para no tener que adivinarlo.
      return [...zonas, ...DESFASES_FIJOS];
    }
  } catch {
    // Un navegador que la anuncia pero falla cae a la lista corta: esto no puede romper la pantalla.
  }

  return ZONAS_DE_RESPALDO;
}

/**
 * Un texto de zona listo para buscar: en minúsculas y con los separadores del nombre (`/`, `_`, `-`)
 * convertidos en espacios, para que «new york» encuentre `America/New_York`.
 */
function paraBuscar(texto: string): string {
  return texto.toLowerCase().replace(/_/g, ' ').replace(/\//g, ' ').replace(/-/g, ' ');
}

/** Un hueco del logo, tal y como lo cuenta la configuración. */
interface HuecoDeLogo {
  readonly filled: boolean;
  readonly fileName?: string;
  readonly size?: number;
  readonly width?: number;
  readonly height?: number;
  readonly updatedAt?: string;
}

interface Configuracion {
  /**
   * El nombre de la instalación: lo que se lee donde antes decía «Catalina Support».
   *
   * Va con **la marca** y no con la numeración: es cómo se llama esta instalación, y se enseña junto
   * al logo en la pantalla de entrada y en el menú (docs/modules/settings.md).
   */
  readonly name: string;
  /**
   * Cómo se entra en la instalación: el método y las dos configuraciones, **sin sus secretos**.
   *
   * Lo que llega de la contraseña de la cuenta de servicio y del secreto del cliente son dos
   * booleanos —`passwordSet` y `secretSet`—, y nunca el valor: se guardan para hablar con el
   * directorio y no salen por la API (docs/modules/settings.md, sección 5.8).
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
  readonly brand: { readonly light: HuecoDeLogo; readonly dark: HuecoDeLogo };
}

/** La configuración del directorio, tal y como viaja por la API. */
interface Directorio {
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
interface Keycloak {
  readonly issuer: string;
  readonly clientId: string;
  readonly redirectUri: string;
  readonly secretSet: boolean;
}

/**
 * La pantalla de Configuración: la marca de la instalación.
 *
 * Vive en `core` porque **es de la instalación y no de un módulo**: es la excepción que recoge
 * `docs/arquitectura.md`, sección 4. La ve sólo un Administrador, y lo comprueba el backend en cada
 * petición: la guarda de la ruta es comodidad, no seguridad.
 *
 * Tiene **el nombre de la instalación** —lo que se lee donde antes decía «Catalina Support»—, **el
 * logo** con sus dos huecos, **el color institucional** con vista previa antes de guardar, **el
 * idioma** y **el prefijo y el reparto** de los tickets (`docs/modules/settings.md`).
 */
@Component({
  selector: 'app-settings-page',
  imports: [RouterLink, Aviso, Boton, Campo, Selector, Tarjeta],
  templateUrl: './settings-page.html',
})
export class SettingsPage {
  private readonly http = inject(HttpClient);
  private readonly textos = inject(TranslationService);
  private readonly marca = inject(BrandService);
  private readonly sesion = inject(SessionService);

  protected readonly configuracion = signal<Configuracion | null>(null);
  protected readonly cargando = signal(true);
  protected readonly guardando = signal(false);
  protected readonly mensaje = signal<{ forma: 'exito' | 'error'; texto: string } | null>(null);

  /** El nombre que se está escribiendo, antes de guardarlo. */
  protected readonly nombreElegido = signal('');

  /**
   * **Si hay algo que guardar en la instalación**: el nombre o el idioma. La tarjeta lleva los dos
   * (decisión del responsable, 2026-09-29) y tiene un solo botón, así que el botón se enciende con
   * cualquiera de los dos.
   */
  protected readonly cambioDeInstalacionPendiente = computed(() => {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return false;
    }

    return (
      this.nombreElegido().trim() !== configuracion.name || this.idioma() !== configuracion.language
    );
  });

  /** El color que se está eligiendo, y **cómo quedaría de verdad**, resuelto por el backend. */
  protected readonly colorElegido = signal('#1d4ed8');
  protected readonly colorResuelto = signal<Marca['colors'] | null>(null);

  /** El archivo elegido para cada hueco, antes de subirlo. */
  protected readonly archivoClaro = signal<File | null>(null);
  protected readonly archivoOscuro = signal<File | null>(null);

  protected readonly cambioDeColorPendiente = computed(
    () => this.colorElegido().toLowerCase() !== (this.configuracion()?.primaryColor ?? '').toLowerCase(),
  );

  // --- Cómo se entra: el método y las dos configuraciones ---
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

  /**
   * Si el directorio está configurado, para poder ofrecer el método de AD.
   *
   * Sale de **lo que se está escribiendo**, no de lo guardado: así, en cuanto se rellenan el servidor
   * y la base de búsqueda, la opción deja de estar apagada sin tener que guardar antes.
   */
  protected readonly directorioConfigurado = computed(
    () => this.dirServidor().trim() !== '' && this.dirBase().trim() !== '',
  );

  /** Lo mismo con Keycloak: sin emisor, cliente y vuelta no hay camino. */
  protected readonly keycloakConfigurado = computed(
    () =>
      this.kcEmisor().trim() !== '' && this.kcCliente().trim() !== '' && this.kcVuelta().trim() !== '',
  );

  /** Si hay algo que guardar en cómo se entra. */
  protected readonly cambioDeEntradaPendiente = computed(() => {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return false;
    }

    // La contraseña y el secreto cuentan **sólo si se han escrito**: vacío quiere decir «no lo
    // cambies», y por eso dejarlos vacíos no marca el cambio como pendiente.
    return (
      this.metodo() !== configuracion.entryMethod ||
      this.dirServidor().trim() !== configuracion.directory.host ||
      this.dirPuerto().trim() !== configuracion.directory.port ||
      this.dirTls() !== configuracion.directory.useTls ||
      this.dirCuenta().trim() !== configuracion.directory.bindDn ||
      this.dirBase().trim() !== configuracion.directory.searchBase ||
      this.dirFiltro().trim() !== configuracion.directory.userFilter ||
      this.dirCorreo().trim() !== configuracion.directory.attrEmail ||
      this.dirNombre().trim() !== configuracion.directory.attrName ||
      this.dirApellidos().trim() !== configuracion.directory.attrLastName ||
      this.dirId().trim() !== configuracion.directory.attrId ||
      this.dirContrasena() !== '' ||
      this.kcEmisor().trim() !== configuracion.keycloak.issuer ||
      this.kcCliente().trim() !== configuracion.keycloak.clientId ||
      this.kcVuelta().trim() !== configuracion.keycloak.redirectUri ||
      this.kcSecreto() !== ''
    );
  });

  /** El idioma de la instalación y la numeración */
  protected readonly idioma = signal('es');
  protected readonly prefijo = signal('');
  protected readonly asignacionPrincipal = signal('por_turnos');
  protected readonly avisoPrincipal = signal('al_asignado');
  protected readonly asignacionInterno = signal('ninguna');
  protected readonly avisoInterno = signal('a_nadie');

  /**
   * La zona horaria de la instalación.
   *
   * **Se elige de la lista, nunca se escribe**: el buscador sólo filtra, y lo que se guarda es el
   * nombre IANA de la opción pulsada.
   */
  protected readonly zonaElegida = signal('');
  /** Lo que se escribe en el buscador de zonas. No se guarda: es una lupa, no el valor. */
  protected readonly busquedaZona = signal('');
  /** La dirección pública que se está escribiendo, antes de guardarla. */
  protected readonly direccionPublica = signal('');

  /** Las zonas que se ofrecen: las del navegador, o la lista corta si no las sabe dar. */
  protected readonly zonas = signal<readonly string[]>(zonasDelNavegador());

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
   * La hora que es ahora mismo en la zona elegida, con su desfase respecto a UTC.
   *
   * El desfase sale de `Intl.DateTimeFormat` con `timeZoneName: 'shortOffset'` —así el horario de
   * verano de quien lo tenga se respeta solo, porque lo resuelve el navegador— y la hora con el
   * mismo formateador. **Si la zona no vale, no se enseña nada**: `Intl` lanza y aquí se recoge, que
   * una zona rara en la base no puede dejar la pantalla a medias.
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

  /**
   * Si la dirección está puesta y **no va cifrada**.
   *
   * El aviso **no bloquea nada**: se puede guardar igual —para probarlo en local, por ejemplo—, pero
   * quien configura la instalación tiene que saber que así la contraseña y la sesión viajan sin
   * cifrar. Vacío no avisa: no hay nada que decir todavía.
   */
  protected readonly direccionInsegura = computed(() => {
    const direccion = this.direccionPublica().trim();
    return direccion !== '' && !direccion.toLowerCase().startsWith('https://');
  });

  /** Elige una zona de la lista. Se limpia el buscador para que la elegida se vea en su sitio. */
  protected elegirZona(zona: string): void {
    this.zonaElegida.set(zona);
    this.busquedaZona.set('');
  }

  /** Si hay algo que guardar en la numeración o el idioma: sin cambios, el botón está apagado. */
  protected readonly cambioDeIdiomaPendiente = computed(
    () => this.idioma() !== (this.configuracion()?.language ?? ''),
  );

  /** Si hay algo que guardar en la numeración: sin cambios, el botón está apagado. */
  protected readonly cambioDeNumeracionPendiente = computed(() => {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return false;
    }

    return (
      this.prefijo().trim() !== configuracion.numberPrefix ||
      this.asignacionPrincipal() !== configuracion.mainAssignment ||
      this.avisoPrincipal() !== configuracion.mainNotification ||
      this.asignacionInterno() !== configuracion.internalAssignment ||
      this.avisoInterno() !== configuracion.internalNotification
    );
  });

  /** **Si hay algo que guardar en la región y la dirección**: es la otra tarjeta, con su botón. */
  protected readonly cambioDeRegionPendiente = computed(() => {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return false;
    }

    return (
      this.zonaElegida().trim() !== configuracion.timeZone ||
      this.direccionPublica().trim() !== configuracion.publicAppUrl
    );
  });

  /** La dirección del logo propio que se está enseñando, con su sello para no ver el viejo. */
  protected readonly versionDeLaMarca = this.marca.version;

  constructor() {
    void this.cargar();
  }

  protected t() {
    return this.textos.textos();
  }

  /**
   * La dirección de la vista previa de un hueco: la de la marca, o el logo de fábrica de ese tema.
   *
   * **El de fábrica que toca al hueco**, y no siempre el mismo: si la instalación no tiene logo
   * propio, lo que se ve en el hueco oscuro es el de fábrica oscuro.
   */
  protected vistaPrevia(oscuro: boolean): string {
    if (!this.marca.hayLogoPropio()) {
      return logoDeFabrica(oscuro);
    }

    return this.marca.urlDelLogo(oscuro);
  }

  protected hayLogoPropio(oscuro: boolean): boolean {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return false;
    }

    return oscuro ? configuracion.brand.dark.filled : configuracion.brand.light.filled;
  }

  /** Un tamaño en algo que se lee: «203 KB», «1,2 MB». */
  protected peso(bytes: number | undefined): string {
    if (!bytes) {
      return '';
    }
    return bytes < 1024 * 1024
      ? `${Math.round(bytes / 1024)} KB`
      : `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  }

  /** Elige un archivo para un hueco. No se sube hasta que se pide: primero se ve el nombre. */
  protected elegirArchivo(oscuro: boolean, evento: Event): void {
    const entrada = evento.target as HTMLInputElement;
    const archivo = entrada.files?.[0] ?? null;

    if (oscuro) {
      this.archivoOscuro.set(archivo);
    } else {
      this.archivoClaro.set(archivo);
    }

    this.mensaje.set(null);
  }

  /** Sube el archivo elegido a su hueco. */
  protected async subir(oscuro: boolean): Promise<void> {
    const archivo = oscuro ? this.archivoOscuro() : this.archivoClaro();
    if (!archivo) {
      return;
    }

    await this.pedir(
      () => {
        const datos = new FormData();
        datos.append('logo', archivo, archivo.name);

        return firstValueFrom(
          this.http.post<Configuracion>(
            `/api/settings/brand/logo?variant=${oscuro ? 'oscuro' : 'claro'}`,
            datos,
          ),
        );
      },
      this.t().configuracion.logoSubido,
      () => {
        if (oscuro) {
          this.archivoOscuro.set(null);
        } else {
          this.archivoClaro.set(null);
        }
      },
    );
  }

  /** Quita el logo propio de un hueco y lo devuelve al de fábrica. */
  protected async volverAlDeFabrica(oscuro: boolean): Promise<void> {
    await this.pedir(
      () =>
        firstValueFrom(
          this.http.delete<Configuracion>(
            `/api/settings/brand/logo?variant=${oscuro ? 'oscuro' : 'claro'}`,
          ),
        ),
      this.t().configuracion.logoQuitado,
    );
  }

  /**
   * Pide al backend cómo quedaría ese color, **sin guardarlo**.
   *
   * La vista previa la resuelve el backend a propósito: la regla de qué se lee y qué no es una sola,
   * y tener una copia en el frontend es la forma segura de que un día digan cosas distintas.
   */
  protected async previsualizarColor(valor: string): Promise<void> {
    this.colorElegido.set(valor);

    if (!/^#[0-9a-fA-F]{6}$/.test(valor)) {
      this.colorResuelto.set(null);
      return;
    }

    try {
      const marca = await firstValueFrom(
        this.http.get<Marca>(`/api/settings/brand?color=${encodeURIComponent(valor)}`),
      );
      this.colorResuelto.set(marca.colors);
      this.aplicarVistaPrevia(marca.colors);
    } catch {
      // Si la vista previa no llega, se enseña el color elegido tal cual: la pantalla no se rompe.
      this.colorResuelto.set(null);
    }
  }

  /** Guarda la configuración entera, como manda un `PUT`. */
  protected async guardarColor(): Promise<void> {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return;
    }

    await this.pedir(
      () =>
        firstValueFrom(
          this.http.put<Configuracion>('/api/settings', {
            ...configuracion,
            primaryColor: this.colorElegido(),
          }),
        ),
      this.t().configuracion.marcaGuardada,
      () => {
        this.colorResuelto.set(null);
        // Y la marca se relee, para que el cambio se aplique en toda la aplicación ahora mismo.
        void this.marca.cargar();
      },
    );
  }

  /**
   * Guarda el nombre de la instalación.
   *
   * Es un `PUT`: va la configuración entera con el nombre nuevo. **Se vuelve a leer la marca** al
   * guardar, que es lo que hace que el nombre nuevo se vea en el menú, en la pantalla de entrada y en
   * la pestaña del navegador **sin recargar** (decisión del responsable, 2026-09-25).
   *
   * Si el campo se deja vacío, el backend guarda el nombre de fábrica: es lo que dice la ayuda del
   * campo, y por eso no se avisa de nada raro.
   */
  protected async guardarNombre(): Promise<void> {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return;
    }

    await this.pedir(
      () =>
        firstValueFrom(
          this.http.put<Configuracion>('/api/settings', {
            ...configuracion,
            name: this.nombreElegido().trim(),
            // **El idioma va con el nombre** (decisión del responsable, 2026-09-29): los guarda el mismo
            // botón, y sin esto se mandaría el que estaba cargado y no el que se acaba de elegir.
            language: this.idioma(),
          }),
        ),
      this.t().configuracion.instalacionGuardada,
    );
  }

  /**
   * Guarda el prefijo y el reparto.
   *
   * Es un `PUT`: va la configuración entera. El prefijo **no toca los números ya emitidos** —cada
   * ticket guarda el suyo— y eso se avisa antes de guardarlo, que es la duda que tiene cualquiera.
   */
  /**
   * Guarda la región y la dirección. El `PUT` manda **la configuración entera** —es un documento,
   * no un parche—, así que es el mismo guardado que el de la numeración: lo que cambia es qué botón
   * lo pide y qué se ha tocado.
   */
  protected async guardarRegion(): Promise<void> {
    await this.guardarLaConfiguracion(this.t().configuracion.regionGuardada);
  }

  protected async guardarNumeracion(): Promise<void> {
    await this.guardarLaConfiguracion(this.t().configuracion.numeracionGuardada);
  }

  /**
   * Guarda lo que pida la tarjeta que lo pide. El `PUT` manda **la configuración entera** —es un
   * documento, no un parche—, así que sólo cambia **el mensaje con el que se confirma**, que es lo que
   * lee quien acaba de pulsar.
   */
  private async guardarLaConfiguracion(mensaje: string): Promise<void> {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return;
    }

    await this.pedir(
      () =>
        firstValueFrom(
          this.http.put<Configuracion>('/api/settings', {
            ...configuracion,
            primaryColor: this.colorElegido(),
            language: this.idioma(),
            timeZone: this.zonaElegida().trim(),
            publicAppUrl: this.direccionPublica().trim(),
            numberPrefix: this.prefijo().trim(),
            mainAssignment: this.asignacionPrincipal(),
            mainNotification: this.avisoPrincipal(),
            internalAssignment: this.asignacionInterno(),
            internalNotification: this.avisoInterno(),
          }),
        ),
      mensaje,
    );
  }

  /** Los dos idiomas de la instalación. */
  protected opcionesDeIdioma(): readonly OpcionSelector[] {
    return [
      { valor: 'es', etiqueta: this.t().idioma.es, grupo: this.t().configuracion.idioma },
      { valor: 'en', etiqueta: this.t().idioma.en, grupo: this.t().configuracion.idioma },
    ];
  }

  /** Las opciones del reparto: repartir o no repartir, y nada más. */
  protected opcionesDeAsignacion(): readonly OpcionSelector[] {
    return [
      {
        valor: 'por_turnos',
        etiqueta: this.t().configuracion.asignacionPorTurnos,
        grupo: this.t().configuracion.asignacion,
      },
      {
        valor: 'ninguna',
        etiqueta: this.t().configuracion.asignacionNinguna,
        grupo: this.t().configuracion.asignacion,
      },
    ];
  }

  /**
   * Las opciones del aviso.
   *
   * **Avisar «al asignado» sin repartir no tiene sentido** —no hay a quién avisar—, así que esa opción
   * sale desactivada: la base tampoco la acepta, y la pantalla no ofrece algo que va a fallar
   * (`docs/modules/settings.md`, sección 5).
   */
  protected opcionesDeAviso(asignacion: string): readonly OpcionSelector[] {
    const reparte = asignacion === 'por_turnos';

    return [
      {
        valor: 'a_nadie',
        etiqueta: this.t().configuracion.avisoANadie,
        grupo: this.t().configuracion.aviso,
      },
      {
        valor: 'a_todo_el_equipo',
        etiqueta: this.t().configuracion.avisoATodoElEquipo,
        grupo: this.t().configuracion.aviso,
      },
      {
        valor: 'al_asignado',
        etiqueta: this.t().configuracion.avisoAlAsignado,
        grupo: this.t().configuracion.aviso,
        deshabilitado: !reparte,
      },
    ];
  }

  /** Cambiar el reparto arrastra el aviso: sin reparto, no se puede avisar al asignado. */
  protected cambiarAsignacion(tipo: 'principal' | 'interno', valor: string): void {
    if (tipo === 'principal') {
      this.asignacionPrincipal.set(valor);
      if (valor === 'ninguna' && this.avisoPrincipal() === 'al_asignado') {
        this.avisoPrincipal.set('a_nadie');
      }
      return;
    }

    this.asignacionInterno.set(valor);
    if (valor === 'ninguna' && this.avisoInterno() === 'al_asignado') {
      this.avisoInterno.set('a_nadie');
    }
  }

  /**
   * Los tres métodos de entrada.
   *
   * Un método que no está configurado **se enseña pero no se puede elegir**: elegirlo dejaría la
   * instalación sin puerta para todo el mundo menos la cuenta de fábrica, y el backend lo rechaza de
   * todos modos (`docs/modules/settings.md`, sección 5.8). Vale tanto lo guardado como lo que se está
   * escribiendo ahora mismo en la tarjeta.
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
        deshabilitado: !this.directorioConfigurado(),
      },
      {
        valor: 'keycloak',
        etiqueta: this.t().configuracion.metodoKeycloak,
        grupo: this.t().configuracion.metodo,
        deshabilitado: !this.keycloakConfigurado(),
      },
    ];
  }

  /** La explicación del método que está elegido, para que se lea qué significa. */
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

  /**
   * Guarda cómo se entra: el método y las dos configuraciones.
   *
   * Va en el mismo `PUT` que lo demás —la configuración es una— y **es lo que hace que el cambio valga
   * sin reiniciar nada**: el camino de entrada lee la base en cada intento
   * (`docs/modules/settings.md`, sección 5.8).
   */
  protected async guardarEntrada(): Promise<void> {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return;
    }

    await this.pedir(
      () =>
        firstValueFrom(
          this.http.put<Configuracion>('/api/settings', {
            ...configuracion,
            entryMethod: this.metodo(),
            directory: { ...configuracion.directory, ...this.directorioEscrito() },
            keycloak: { ...configuracion.keycloak, ...this.keycloakEscrito() },
          }),
        ),
      this.t().configuracion.entradaGuardada,
    );
  }

  /**
   * Prueba la conexión con el directorio **con lo que hay en pantalla**, antes de guardarlo.
   *
   * Es lo que evita guardar una configuración que no funciona y quedarse sin poder entrar: la prueba
   * la contesta el módulo `auth`, que es el que sabe hablar con un directorio
   * (`docs/modules/settings.md`, sección 5.8).
   */
  protected async probarDirectorio(): Promise<void> {
    await this.probar(
      '/api/settings/directory/test',
      this.directorioEscrito(),
      this.t().configuracion.directorioOk,
    );
  }

  /** Lo mismo con Keycloak: se lee el reino y se comprueba que dice dónde está su pantalla. */
  protected async probarKeycloak(): Promise<void> {
    await this.probar(
      '/api/settings/keycloak/test',
      this.keycloakEscrito(),
      this.t().configuracion.keycloakOk,
    );
  }

  /** Lo que se ha escrito del directorio, sin los campos que la API no acepta de vuelta. */
  private directorioEscrito(): Partial<Directorio> & { bindPassword?: string } {
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
      // Vacío quiere decir «no la cambies» y no «bórrala»: es lo mismo que hace el backend al guardar.
      bindPassword: this.dirContrasena(),
    };
  }

  private keycloakEscrito(): Partial<Keycloak> & { clientSecret?: string } {
    return {
      issuer: this.kcEmisor().trim(),
      clientId: this.kcCliente().trim(),
      redirectUri: this.kcVuelta().trim(),
      clientSecret: this.kcSecreto(),
    };
  }

  /** Una prueba de conexión: no guarda nada y cuenta lo que ha pasado en el aviso de arriba. */
  private async probar(url: string, cuerpo: object, exito: string): Promise<void> {
    this.guardando.set(true);
    this.mensaje.set(null);

    try {
      await firstValueFrom(this.http.post(url, cuerpo));
      this.mensaje.set({ forma: 'exito', texto: exito });
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textoDelError(error) });
    } finally {
      this.guardando.set(false);
    }
  }

  /** Deshace la vista previa: se vuelve al color que hay guardado. */
  protected descartarColor(): void {
    const configuracion = this.configuracion();
    if (!configuracion) {
      return;
    }

    this.colorElegido.set(configuracion.primaryColor);
    this.colorResuelto.set(null);
    this.aplicarVistaPrevia({
      light: configuracion.primaryColor,
      dark: configuracion.primaryColor,
      onLight: '#ffffff',
      onDark: '#0d0f12',
      isDefault: false,
    });
  }

  private async cargar(): Promise<void> {
    this.cargando.set(true);
    try {
      const configuracion = await firstValueFrom(this.http.get<Configuracion>('/api/settings'));
      this.configuracion.set(configuracion);
      this.ponerEntrada(configuracion);
      this.nombreElegido.set(configuracion.name);
      this.colorElegido.set(configuracion.primaryColor);
      this.ponerNumeracion(configuracion);
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textoDelError(error) });
    } finally {
      this.cargando.set(false);
    }
  }

  /** Hace una petición que devuelve la configuración, y deja la pantalla al día. */
  private async pedir(
    peticion: () => Promise<Configuracion>,
    exito: string,
    despues?: () => void,
  ): Promise<void> {
    this.guardando.set(true);
    this.mensaje.set(null);

    try {
      const configuracion = await peticion();
      this.configuracion.set(configuracion);
      this.ponerEntrada(configuracion);
      this.nombreElegido.set(configuracion.name);
      this.colorElegido.set(configuracion.primaryColor);
      this.ponerNumeracion(configuracion);
      despues?.();
      this.mensaje.set({ forma: 'exito', texto: exito });
      void this.marca.cargar();
    } catch (error) {
      this.mensaje.set({ forma: 'error', texto: this.textoDelError(error) });
    } finally {
      this.guardando.set(false);
    }
  }

  /** Deja los campos de la numeración con lo que hay guardado. */
  private ponerNumeracion(configuracion: Configuracion): void {
    this.idioma.set(configuracion.language);
    this.zonaElegida.set(configuracion.timeZone);
    this.direccionPublica.set(configuracion.publicAppUrl);
    this.prefijo.set(configuracion.numberPrefix);
    this.asignacionPrincipal.set(configuracion.mainAssignment);
    this.avisoPrincipal.set(configuracion.mainNotification);
    this.asignacionInterno.set(configuracion.internalAssignment);
    this.avisoInterno.set(configuracion.internalNotification);
  }

  /**
   * Deja los campos de cómo se entra con lo que hay guardado.
   *
   * **La contraseña y el secreto se dejan vacíos a propósito**: lo que llega es «hay uno puesto», nunca
   * el valor. Escribir en esos campos es cambiarlo; dejarlos vacíos es no tocarlo.
   */
  private ponerEntrada(configuracion: Configuracion): void {
    this.metodo.set(configuracion.entryMethod);
    this.dirServidor.set(configuracion.directory.host);
    this.dirPuerto.set(configuracion.directory.port);
    this.dirTls.set(configuracion.directory.useTls);
    this.dirCuenta.set(configuracion.directory.bindDn);
    this.dirContrasena.set('');
    this.dirBase.set(configuracion.directory.searchBase);
    this.dirFiltro.set(configuracion.directory.userFilter);
    this.dirCorreo.set(configuracion.directory.attrEmail);
    this.dirNombre.set(configuracion.directory.attrName);
    this.dirApellidos.set(configuracion.directory.attrLastName);
    this.dirId.set(configuracion.directory.attrId);
    this.kcEmisor.set(configuracion.keycloak.issuer);
    this.kcCliente.set(configuracion.keycloak.clientId);
    this.kcSecreto.set('');
    this.kcVuelta.set(configuracion.keycloak.redirectUri);
  }

  /** La vista previa se hace con las variables del tema: es exactamente lo que se va a ver. */
  private aplicarVistaPrevia(colors: Marca['colors']): void {
    const raiz = document.documentElement.style;
    raiz.setProperty('--institucional-claro', colors.light);
    raiz.setProperty('--institucional-oscuro', colors.dark);
    raiz.setProperty('--institucional-texto-claro', colors.onLight);
    raiz.setProperty('--institucional-texto-oscuro', colors.onDark);
  }

  private textoDelError(error: unknown): string {
    return this.textos.error(claveDelError(error));
  }
}
