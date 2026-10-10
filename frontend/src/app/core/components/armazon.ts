import { Component, computed, inject, signal } from '@angular/core';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { Controles } from './controles';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Dialogo } from '../../shared/components/dialogo';
import { Icono } from '../../shared/components/icono';
import { Logo } from './logo';
import { TranslationService } from '../i18n/translation.service';
import { claveDeEntrada, entradasPara, type EntradaDeMenu } from '../navigation';
import { BrandService } from '../services/brand.service';
import { SessionService } from '../services/session.service';
import { consultaDeAncho } from '../../shared/pantalla';
import type { Textos } from '../i18n/es';

/** Dónde se recuerda si el menú está plegado, como el tema. */
const CLAVE_PLEGADO = 'catalina-support.menu-plegado';

/** A partir de aquí se considera PC, que es donde el menú se puede plegar (sección 3.1). */
const ANCHO_DE_PC = 1200;

/**
 * El armazón de la aplicación: el menú lateral y el contenido.
 *
 * **No hay barra superior** (`docs/interfaz-y-experiencia.md`, sección 3.1): el menú lo lleva todo y
 * el contenido se queda con todo el ancho. El menú tiene tres zonas —el producto arriba, las
 * opciones del papel en medio y los controles abajo—, y **los controles no cuentan** como opciones:
 * el tema, el idioma y salir son iguales para todos.
 *
 * En móvil hay una tira superior mínima con el botón de menú, porque sin ella no habría dónde poner
 * el disparador; el menú se abre como cajón sobre el contenido.
 */
@Component({
  selector: 'app-armazon',
  imports: [RouterOutlet, RouterLink, RouterLinkActive, Logo, Icono, Controles, Boton, Dialogo],
  templateUrl: './armazon.html',
})
export class Armazon {
  private readonly sesion = inject(SessionService);
  private readonly marca = inject(BrandService);
  private readonly textos = inject(TranslationService);
  private readonly router = inject(Router);

  protected readonly usuario = this.sesion.usuario;

  /**
   * El nombre de la instalación, que es el que se lee arriba junto al logo.
   *
   * **No es lo mismo que `nombre`**, que es el de la persona que ha entrado y va en los controles.
   */
  protected readonly nombreDeLaInstalacion = this.marca.nombre;

  /** La versión del sistema, con su `v`: se lee en la fila de salir, alineada a la derecha. */
  protected readonly versionDelSistema = this.marca.versionDelSistema;

  /** Licencia y fuente correspondiente de esta compilación. */
  protected readonly fuente = this.marca.fuente;

  /** Si el menú está recogido: la tira de iconos. Es una elección, y se recuerda. */
  private readonly plegado = signal(leerPlegado());

  /** Si la pantalla es ancha, que es donde el menú va fijo y se puede plegar. */
  private readonly pantallaAncha = signal(consultaAncha()?.matches ?? true);

  /**
   * Si el menú se enseña plegado **de verdad**.
   *
   * El plegado es cosa de PC y tablet, donde el menú va fijo al lado del contenido. En móvil el menú
   * es un cajón que se abre encima: ahí no se pliega, porque **no hay nada que ganar** —tapa el
   * contenido mientras está abierto— y un cajón de iconos no se entiende.
   */
  protected readonly plegadoEfectivo = computed(() => this.pantallaAncha() && this.plegado());

  /** Si el cajón está abierto: sólo tiene sentido en móvil y tablet. */
  protected readonly cajonAbierto = signal(false);

  /** Si la pantalla es ancha: lo decide la plantilla para enseñar el botón de plegar. */
  protected readonly esAncha = this.pantallaAncha.asReadonly();

  constructor() {
    // Si alguien cambia el tamaño de la ventana o gira el móvil, el menú se coloca como toca.
    consultaAncha()?.addEventListener('change', (evento) => this.pantallaAncha.set(evento.matches));
  }

  /** Las entradas que le tocan al papel de quien ha entrado. */
  protected readonly entradas = computed(() => entradasPara(this.usuario()?.role ?? ''));

  protected readonly nombre = computed(() => {
    const quien = this.usuario();
    if (!quien) {
      return '';
    }
    return `${quien.name} ${quien.lastName}`.trim();
  });

  /**
   * Si es la cuenta de fábrica.
   *
   * **No tiene perfil** —no está en la tabla de cuentas—, así que su nombre no lleva a ninguna parte:
   * ofrecerle un enlace que responde 404 sería prometer algo que no existe
   * (`docs/modules/users.md`, sección 8).
   */
  protected readonly esDeFabrica = computed(() => this.usuario()?.factory === true);

  protected t() {
    return this.textos.textos();
  }

  /** El papel de quien ha entrado, como se llama en la interfaz. */
  protected papel(): string {
    const papel = this.usuario()?.role as keyof Textos['papeles'] | undefined;

    return papel ? this.t().papeles[papel] : '';
  }

  /**
   * La etiqueta de una entrada del menú, en el idioma de quien lee.
   *
   * Algunas entradas se llaman distinto según el papel —la bandeja es «Mis tickets», «Bandeja» o «Mi
   * bandeja»—, así que la clave se resuelve con el papel de quien ha entrado.
   */
  protected etiquetaDe(entrada: EntradaDeMenu): string {
    return this.t().menu[claveDeEntrada(entrada, this.usuario()?.role ?? '')];
  }

  /** Pliega o despliega el menú, y lo recuerda. */
  protected alternarPlegado(): void {
    const nuevo = !this.plegado();
    this.plegado.set(nuevo);

    try {
      localStorage.setItem(CLAVE_PLEGADO, nuevo ? 'si' : 'no');
    } catch {
      // Sin almacenamiento el menú se pliega igual: sólo no se recuerda.
    }
  }

  protected abrirCajon(): void {
    this.cajonAbierto.set(true);
  }

  protected cerrarCajon(): void {
    this.cajonAbierto.set(false);
  }

  /**
   * **Salir se confirma** (decisión del responsable, 2026-09-29): el botón del menú abre una ventana
   * antes de cerrar la sesión. Salir sin querer cuesta volver a entrar —y con el directorio, volver a
   * pasar por allí—, así que se pregunta, con el mismo diálogo que el resto de las acciones que
   * preguntan.
   */
  protected readonly confirmandoLaSalida = signal(false);

  protected preguntarPorLaSalida(): void {
    this.confirmandoLaSalida.set(true);
    // El cajón, si está abierto en móvil, se cierra: la ventana manda.
    this.cajonAbierto.set(false);
  }

  protected cerrarLaConfirmacion(): void {
    this.confirmandoLaSalida.set(false);
  }

  /** Cierra la sesión de verdad, ya confirmada. */
  protected async salir(): Promise<void> {
    this.confirmandoLaSalida.set(false);
    await this.sesion.salir();
    await this.router.navigate(['/login']);
  }
}

/**
 * Si la pantalla es ancha: es el mismo corte que usa la plantilla para el menú fijo (`lg`, 1024 px).
 *
 * Está en `shared` porque la lista de usuarios lo necesita para lo mismo: decidir si se pinta la tabla
 * o las tarjetas.
 */
function consultaAncha(): MediaQueryList | null {
  return consultaDeAncho(ANCHO_DE_PC);
}

/**
 * Si el menú estaba plegado.
 *
 * De fábrica, **plegado en tablet y desplegado en PC**: la tablet tiene menos ancho y el contenido
 * agradece el sitio. Quien lo cambie, lo cambia para siempre.
 */
function leerPlegado(): boolean {
  let guardado: string | null = null;
  try {
    guardado = localStorage.getItem(CLAVE_PLEGADO);
  } catch {
    guardado = null;
  }

  if (guardado !== null) {
    return guardado === 'si';
  }

  return typeof window !== 'undefined' && window.innerWidth < ANCHO_DE_PC;
}
