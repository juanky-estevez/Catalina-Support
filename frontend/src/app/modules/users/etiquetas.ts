import type { Textos } from '../../core/i18n/es';
import type { Idioma } from '../../core/i18n/translation.service';
import { opcionesConZona } from '../../shared/fechas';
import type { OpcionSelector } from '../../shared/components/selector';
import type { Cuenta } from './users.service';

/**
 * Las etiquetas y las listas cerradas del módulo `users`.
 *
 * Los papeles y los orígenes son **valores cerrados**, los mismos que valida el backend
 * (`docs/modules/users.md`, secciones 2 y 9). Están aquí una sola vez para que las tres pantallas
 * —la lista, la ficha y el perfil— digan lo mismo, y para que añadir un valor no sea buscar en tres
 * plantillas.
 */

/** Los cuatro papeles, en el orden en que se ofrecen. */
export const PAPELES = ['usuario', 'soporte', 'desarrollo', 'administrador'] as const;

/** Los tres orígenes de una cuenta. */
export const ORIGENES = ['local', 'ad', 'keycloak'] as const;

/**
 * Los orígenes que **hoy no se pueden elegir**.
 *
 * El alta de una cuenta de directorio y el cambio de origen hacia él exigen preguntarle al directorio,
 * y ese camino todavía no existe: el backend los rechaza con `users.directory.notFound`
 * (`docs/modules/users.md`, sección 5, punto 4). En la pantalla **se ven y salen desactivados**, con
 * su nota, en vez de ofrecerse y fallar después (`docs/interfaz-y-experiencia.md`, sección 3.6).
 *
 * El día que llegue el camino del directorio, esto se queda vacío y no hay que tocar ninguna pantalla.
 */
export const ORIGENES_DE_DIRECTORIO: readonly string[] = ['ad', 'keycloak'];

/** Cómo se llama un papel en la interfaz. */
export function etiquetaDePapel(textos: Textos, papel: string): string {
  const nombre = textos.papeles[papel as keyof Textos['papeles']];
  return nombre ?? papel;
}

/** Cómo se llama un origen en la interfaz. */
export function etiquetaDeOrigen(textos: Textos, origen: string): string {
  const nombre = textos.origenes[origen as keyof Textos['origenes']];
  return nombre ?? origen;
}

/** Los papeles que se pueden repartir: Soporte sólo crea usuarios, y eso lo decide la pantalla. */
export function opcionesDePapel(
  textos: Textos,
  papeles: readonly string[],
): readonly OpcionSelector[] {
  return papeles.map((papel) => ({
    valor: papel,
    etiqueta: etiquetaDePapel(textos, papel),
    grupo: textos.usuarios.papel,
  }));
}

/** Los tres orígenes, con los de directorio **desactivados** mientras no exista su camino. */
export function opcionesDeOrigen(textos: Textos): readonly OpcionSelector[] {
  return ORIGENES.map((origen) => ({
    valor: origen,
    etiqueta: etiquetaDeOrigen(textos, origen),
    grupo: textos.usuarios.origen,
    deshabilitado: ORIGENES_DE_DIRECTORIO.includes(origen),
  }));
}

/**
 * Si esa cuenta **se reactiva sola** al entrar por su camino.
 *
 * Es el caso de Keycloak: su camino no tiene forma de preguntar «¿sigues conociendo a esta persona?»
 * sin una cuenta de servicio en el reino, así que la reactivación a mano se rechaza —y se dice por
 * qué—, y el acceso vuelve solo en cuanto la persona entra (docs/modules/users.md, sección 5, punto
 * 4). **La interfaz no ofrece lo que no se puede hacer**: en vez del botón, se cuenta lo que pasa.
 *
 * En AD sí se ofrece: allí el botón pregunta al directorio antes de devolver el acceso.
 */
export function seReactivaSola(cuenta: Cuenta): boolean {
  return !cuenta.isActive && cuenta.origin === 'keycloak';
}

/**
 * Una fecha en algo que se lee: «24/09/2026, 16:30».
 *
 * Se escribe en el idioma de quien mira, que es lo que hace que el orden de los números sea el que
 * espera. **El año va entero y los números con dos cifras**: el formato corto del navegador deja
 * «24/9/26», que en una lista de cuentas se lee peor y se confunde con un día de otro mes. Si no hay
 * fecha —nunca ha entrado— se responde vacío, y la pantalla pone su texto.
 *
 * **La hora es la de la instalación**, que llega en `zona` (la marca pública): la última entrada se
 * lee a la hora que es donde está la mesa de ayuda, y no a la de quien mira. Si `zona` no llega o
 * `Intl` no la conoce, se cae a la del navegador sin lanzar (`shared/fechas.ts`).
 */
export function fechaCorta(iso: string | undefined, idioma: Idioma, zona?: string): string {
  if (!iso) {
    return '';
  }

  const momento = new Date(iso);
  if (Number.isNaN(momento.getTime())) {
    return '';
  }

  return new Intl.DateTimeFormat(
    idioma === 'en' ? 'en-GB' : 'es-ES',
    opcionesConZona(
      {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      },
      zona,
    ),
  ).format(momento);
}
