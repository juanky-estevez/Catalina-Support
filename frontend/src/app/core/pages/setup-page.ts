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
import { Toast } from '../../shared/components/toast';
import { claveDelError, memoriaDelError } from '../../shared/errores';
import { interpolar } from '../../shared/textos';
import type { CatalogoLocalDeIA } from '../services/settings.service';

/** Los cinco pasos del asistente, en orden. */
type Paso = 1 | 2 | 3 | 4 | 5;

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
    const soportado = (Intl as unknown as { supportedValuesOf?: (clave: string) => string[] })
      .supportedValuesOf;
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
 * Cinco pasos —instalación, entrada, ubicación, correo e IA— más un resumen final. Cada
 * paso **se guarda al avanzar**, así que cerrar el navegador a medias no pierde lo hecho: al volver,
 * el `GET` trae lo guardado y los campos salen rellenos
 * (`docs/primer-arranque.md`, secciones 3 y 6).
 *
 * Vive en `core` porque **es de la instalación entera y ocurre antes de que exista ningún módulo**.
 * No lleva el armazón del menú —no hay nada que navegar todavía— y se enseña con la marca de fábrica.
 */
@Component({
  selector: 'app-setup-page',
  imports: [Aviso, Boton, Campo, Logo, Selector, Tarjeta, Toast],
  templateUrl: './setup-page.html',
})
export class SetupPage {
  protected readonly interpolar = interpolar;
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
  protected readonly idioma = signal('en');

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
  protected readonly kcInterno = signal('');
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

  // --- Paso 5 · la IA obligatoria ---
  protected readonly iaModo = signal<'local' | 'remote' | 'provider'>('local');
  protected readonly iaProveedor = signal('local');
  protected readonly iaDireccion = signal('http://ai:8080');
  protected readonly iaModelo = signal('qwen2.5-1.5b-instruct');
  protected readonly iaAutenticacion = signal<'none' | 'bearer' | 'header' | 'basic'>('none');
  protected readonly iaCabecera = signal('');
  protected readonly iaCredencial = signal('');
  protected readonly iaPrivacidad = signal(false);
  protected readonly catalogoIA = signal<CatalogoLocalDeIA | null>(null);
  protected readonly licenciasAceptadas = signal<ReadonlySet<string>>(new Set());

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

  constructor() {
    // Una instalación nueva empieza siempre en inglés, aunque este navegador recuerde otro idioma.
    // Al cargar un asistente empezado, ponerEstado aplica inmediatamente el valor ya guardado.
    this.textos.cambiar('en');
    void this.cargar();
  }

  protected t() {
    return this.textos.textos();
  }

  /** Los cinco pasos, para la lista de avance. */
  protected pasos(): readonly { numero: Paso; titulo: string }[] {
    const t = this.t().instalacion;
    return [
      { numero: 1, titulo: t.paso1 },
      { numero: 2, titulo: t.paso2 },
      { numero: 3, titulo: t.paso3 },
      { numero: 4, titulo: t.paso4 },
      { numero: 5, titulo: t.paso5 },
    ];
  }

  /** «Paso 2 de 4». */
  protected pasoDeCuatro(): string {
    return interpolar(this.t().instalacion.pasoDeCuatro, {
      paso: String(this.paso()),
      total: '5',
    });
  }

  /** Los dos idiomas de la instalación. */
  protected opcionesDeIdioma(): readonly OpcionSelector[] {
    return [
      { valor: 'es', etiqueta: this.t().idioma.es, grupo: this.t().configuracion.idioma },
      { valor: 'en', etiqueta: this.t().idioma.en, grupo: this.t().configuracion.idioma },
    ];
  }

  /** Aplica el idioma elegido a toda la aplicación mientras se completa el asistente. */
  protected cambiarIdioma(idioma: string): void {
    const elegido = idioma === 'es' ? 'es' : 'en';
    this.idioma.set(elegido);
    this.textos.cambiar(elegido);
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

  protected opcionesDeModoIA(): readonly OpcionSelector[] {
    const t = this.t().configuracion;
    return [
      { valor: 'local', etiqueta: t.iaLocal, grupo: t.motorDeIA },
      { valor: 'remote', etiqueta: t.iaRemota, grupo: t.motorDeIA },
      { valor: 'provider', etiqueta: t.iaProveedor, grupo: t.motorDeIA },
    ];
  }

  protected cambiarModoIA(valor: string): void {
    const modo = valor as 'local' | 'remote' | 'provider';
    this.iaModo.set(modo);
    this.iaProveedor.set(modo === 'local' ? 'local' : modo === 'remote' ? 'openai-compatible' : 'openai');
    if (modo === 'local') {
      this.iaDireccion.set('http://ai:8080');
      this.iaModelo.set('qwen2.5-1.5b-instruct');
      this.iaAutenticacion.set('none');
      this.iaPrivacidad.set(false);
    } else {
      this.iaDireccion.set('');
      this.iaAutenticacion.set('bearer');
    }
  }

  protected opcionesDeProveedorIA(): readonly OpcionSelector[] {
    return ['openai', 'claude', 'deepseek', 'openai-compatible'].map((valor) => ({
      valor,
      etiqueta: valor === 'openai-compatible' ? this.t().configuracion.iaCompatible : valor,
      grupo: this.t().configuracion.iaProveedor,
    }));
  }

  protected opcionesDeModeloLocal(): readonly OpcionSelector[] {
    return [
      { valor: 'qwen2.5-1.5b-instruct', etiqueta: 'Qwen2.5 1.5B · RAM 2 GiB · disk ~1.1 GB', grupo: this.t().configuracion.iaLocal },
      { valor: 'qwen2.5-3b-instruct', etiqueta: 'Qwen2.5 3B · RAM 4 GiB · disk ~2.0 GB', grupo: this.t().configuracion.iaLocal },
      { valor: 'qwen2.5-7b-instruct', etiqueta: 'Qwen2.5 7B · RAM 6 GiB · disk ~4.7 GB', grupo: this.t().configuracion.iaLocal },
    ];
  }

  protected opcionesDeAuthIA(): readonly OpcionSelector[] {
    const t = this.t().configuracion;
    return [
      { valor: 'none', etiqueta: t.iaAuth_none, grupo: t.iaAutenticacion },
      { valor: 'bearer', etiqueta: t.iaAuth_bearer, grupo: t.iaAutenticacion },
      { valor: 'header', etiqueta: t.iaAuth_header, grupo: t.iaAutenticacion },
      { valor: 'basic', etiqueta: t.iaAuth_basic, grupo: t.iaAutenticacion },
    ];
  }
  protected formatoBytes(bytes: number): string { return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GiB`; }
  protected aceptarLicencia(id: string, accepted: boolean): void { const next=new Set(this.licenciasAceptadas());accepted?next.add(id):next.delete(id);this.licenciasAceptadas.set(next); }
  protected async cargarCatalogoIA(): Promise<void> { await this.accionModelo(async()=>{this.catalogoIA.set(await this.setup.modelosLocales());}); }
  protected async descargarModelo(id: string): Promise<void> { await this.accionModelo(async()=>{await this.setup.descargarModelo(id,this.licenciasAceptadas().has(id));this.catalogoIA.set(await this.setup.modelosLocales());}); }
  protected async activarModelo(id: string): Promise<void> { await this.accionModelo(async()=>{await this.setup.activarModelo(id);this.iaModelo.set(id);this.iaDireccion.set('http://ai:8080');this.catalogoIA.set(await this.setup.modelosLocales());}); }
  protected async eliminarModelo(id: string): Promise<void> { await this.accionModelo(async()=>{await this.setup.eliminarModelo(id);this.catalogoIA.set(await this.setup.modelosLocales());}); }
  private async accionModelo(action:()=>Promise<void>):Promise<void>{this.guardando.set(true);this.mensaje.set(null);try{await action();}catch(error){this.mensaje.set({forma:'error',texto:this.textoDelError(error)});}finally{this.guardando.set(false);}}

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

  /** Guarda el paso actual y avanza. Después del quinto se abre el resumen. */
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

    if (this.paso() === 5) {
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

  /**
   * Prueba la conexión del paso 2 **con lo que hay en pantalla**, antes de guardarlo.
   *
   * Prueba lo que corresponda al método elegido —el directorio con AD y el reino con Keycloak— y **no
   * guarda nada**: es lo que evita terminar la instalación con una puerta que no funciona. La contesta
   * el mismo módulo que la prueba de Configuración (`docs/primer-arranque.md`, sección 3).
   */
  protected async probarEntrada(): Promise<void> {
    const t = this.t().configuracion;
    await this.probar(
      () =>
        this.setup.probarEntrada({
          entryMethod: this.metodo(),
          directory: this.directorioEscrito(),
          keycloak: this.keycloakEscrito(),
        }),
      this.metodo() === 'keycloak' ? t.keycloakOk : t.directorioOk,
    );
  }

  /**
   * Prueba la conexión del correo del paso 4 **con lo que hay en pantalla**, antes de guardarlo.
   *
   * Comprueba la conexión y la autenticación y **no manda ningún correo**: en el asistente no hay
   * destinatario (`docs/primer-arranque.md`, sección 3).
   */
  protected async probarCorreo(): Promise<void> {
    await this.probar(
      () => this.setup.probarCorreo({ mail: this.correoEscrito() }),
      this.t().instalacion.correoOk,
    );
  }

  /** Una prueba de conexión: no guarda nada y cuenta lo que ha pasado en el aviso de arriba. */
  private async probar(peticion: () => Promise<unknown>, exito: string): Promise<void> {
    this.guardando.set(true);
    this.mensaje.set(null);

    try {
      await peticion();
      this.mensaje.set({ forma: 'exito', texto: exito });
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
            : paso === 4
              ? this.setup.guardarCorreo({ mail: this.correoEscrito() })
              : this.setup.guardarIA({
                  ai: {
                    mode: this.iaModo(), provider: this.iaProveedor(),
                    baseUrl: this.iaDireccion().trim(), model: this.iaModelo().trim(),
                    authType: this.iaAutenticacion(), authHeader: this.iaCabecera().trim(),
                    credential: this.iaCredencial(),
                    privacyConfirmed: this.iaModo() === 'local' || this.iaPrivacidad(),
                    language: this.idioma(),
                  },
                }));

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
      internalIssuer: this.kcInterno().trim(),
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
      this.cambiarIdioma(estado.language || 'en');
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
    this.kcInterno.set(estado.keycloak.internalIssuer ?? '');
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
    const ia = estado.ai;
    if (ia?.mode) this.iaModo.set(ia.mode);
    this.iaProveedor.set(ia?.provider || 'local');
    this.iaDireccion.set(ia?.baseUrl || 'http://ai:8080');
    this.iaModelo.set(ia?.model || 'qwen2.5-1.5b-instruct');
    this.iaAutenticacion.set(ia?.authType || 'none');
    this.iaCabecera.set(ia?.authHeader || '');
    this.iaCredencial.set('');
    this.iaPrivacidad.set(ia?.privacyConfirmed ?? false);
  }

  private textoDelError(error: unknown): string {
    const key = claveDelError(error);
    const memory = memoriaDelError(error);
    if (key === 'settings.ai.insufficientMemory' && memory) {
      return interpolar(this.textos.error(key), {
        requerida: this.formatoBytes(memory.requiredBytes), disponible: this.formatoBytes(memory.availableBytes),
      });
    }
    return this.textos.error(key);
  }
}
