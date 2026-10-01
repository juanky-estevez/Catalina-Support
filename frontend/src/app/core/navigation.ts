import type { Textos } from './i18n/es';

/** Un papel del producto, tal y como lo llama el backend. */
export type Papel = 'usuario' | 'soporte' | 'desarrollo' | 'administrador';

/** Las claves de las entradas del menú, que son las del diccionario de textos. */
export type ClaveDeMenu = keyof Textos['menu'];

/**
 * Una entrada del menú lateral.
 *
 * `papeles` vacío quiere decir «para todo el mundo que haya entrado».
 */
export interface EntradaDeMenu {
  readonly ruta: string;
  readonly clave: ClaveDeMenu;
  /**
   * Cómo se llama la entrada para un papel, cuando no es lo mismo para todos.
   *
   * La bandeja es **una sola pantalla con dos nombres** —«Mis tickets» para el usuario, Soporte y
   * Desarrollo, y «Bandeja» para el Administrador, que no tiene tickets propios—, que es como la llama
   * `docs/interfaz-y-experiencia.md`, sección 3.2. Los demás papeles usan el nombre de `clave`.
   */
  readonly clavePorPapel?: Partial<Record<Papel, ClaveDeMenu>>;
  /** Qué icono lleva cuando el menú está plegado: los dibuja el propio componente. */
  readonly icono:
    'tickets' | 'principales' | 'internos' | 'categorias' | 'usuarios' | 'configuracion';
  readonly papeles: readonly Papel[];
}

/**
 * Las entradas del menú.
 *
 * **La lista es la que se puede usar hoy**, y es la de `docs/interfaz-y-experiencia.md`, sección 3.2:
 * **Mis tickets**, **Tickets principales**, **Tickets internos**, Usuarios y Configuración.
 *
 * **«Nuevo ticket» ya no está** (decisión del responsable, 2026-09-26): se confundía con un módulo, y
 * crear un ticket es una acción de la bandeja, no un sitio. El botón vive dentro de ella y la ruta
 * `/tickets/new` sigue existiendo.
 *
 * **Las dos listas del «todo» son para Soporte y Desarrollo**, que son los dos papeles que pueden
 * verlo todo: la Bandeja pasó a ser «lo mío» —lo asignado, lo abierto por uno y donde ha comentado— y
 * sin estas dos pantallas el trabajo que no es de nadie no tendría dónde verse (decisión 52).
 */
export const ENTRADAS: readonly EntradaDeMenu[] = [
  {
    ruta: '/tickets',
    clave: 'misTickets',
    clavePorPapel: { administrador: 'bandeja' },
    icono: 'tickets',
    papeles: ['usuario', 'soporte', 'desarrollo', 'administrador'],
  },
  {
    ruta: '/tickets/main',
    clave: 'ticketsPrincipales',
    icono: 'principales',
    papeles: ['soporte', 'desarrollo'],
  },
  {
    ruta: '/tickets/internal',
    clave: 'ticketsInternos',
    icono: 'internos',
    papeles: ['soporte', 'desarrollo'],
  },
  {
    // **«Categorías y etiquetas» es una pantalla de `tickets`** (decisión 64): el catálogo son datos
    // de los tickets y su sitio es este módulo, no Configuración. La mantiene Soporte, que crea y
    // renombra, y **retirar es sólo del Administrador** —que también la ve— porque dejar el catálogo
    // sin ninguna activa rompería «no hay ticket sin categoría».
    ruta: '/tickets/categories',
    clave: 'categorias',
    icono: 'categorias',
    papeles: ['soporte', 'administrador'],
  },
  { ruta: '/users', clave: 'usuarios', icono: 'usuarios', papeles: ['soporte', 'administrador'] },
  { ruta: '/settings', clave: 'configuracion', icono: 'configuracion', papeles: ['administrador'] },
];

/** Cómo se llama una entrada para ese papel: unas se llaman distinto según quién mire. */
export function claveDeEntrada(entrada: EntradaDeMenu, papel: string): ClaveDeMenu {
  return entrada.clavePorPapel?.[papel as Papel] ?? entrada.clave;
}

/** Las entradas que le tocan a ese papel. */
export function entradasPara(papel: string): readonly EntradaDeMenu[] {
  return ENTRADAS.filter(
    (entrada) => entrada.papeles.length === 0 || entrada.papeles.includes(papel as Papel),
  );
}
