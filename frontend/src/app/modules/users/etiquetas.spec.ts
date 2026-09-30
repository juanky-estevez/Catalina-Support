import { ES } from '../../core/i18n/es';
import {
  ORIGENES_DE_DIRECTORIO,
  etiquetaDeIdioma,
  etiquetaDeOrigen,
  etiquetaDePapel,
  fechaCorta,
  opcionesDeIdioma,
  opcionesDeOrigen,
  opcionesDePapel,
  seReactivaSola,
} from './etiquetas';
import type { Cuenta } from './users.service';

describe('las etiquetas del módulo users', () => {
  it('un papel se lee como se llama en la interfaz, no como lo llama el backend', () => {
    expect(etiquetaDePapel(ES, 'soporte')).toBe('Soporte técnico');
    expect(etiquetaDePapel(ES, 'administrador')).toBe('Administrador');
  });

  it('un valor que no conocemos se enseña tal cual, en vez de desaparecer', () => {
    // Es lo que evita que una cuenta se quede sin papel a la vista si el backend estrena uno: se lee
    // el valor, aunque no esté traducido, y se ve que hay algo que no cuadra.
    expect(etiquetaDePapel(ES, 'superusuario')).toBe('superusuario');
    expect(etiquetaDeOrigen(ES, 'otro')).toBe('otro');
  });

  it('los idiomas se leen en su propio idioma: no se traducen', () => {
    expect(etiquetaDeIdioma('es')).toBe('Español');
    expect(etiquetaDeIdioma('en')).toBe('English');
  });

  it('los orígenes de directorio salen desactivados, con su nota en la pantalla', () => {
    // Mientras no exista el camino del directorio, el backend los rechaza con
    // `users.directory.notFound`: la pantalla los enseña, pero no deja elegirlos
    // (docs/interfaz-y-experiencia.md, sección 3.6).
    const opciones = opcionesDeOrigen(ES);

    expect(opciones.map((opcion) => opcion.valor)).toEqual(['local', 'ad', 'keycloak']);

    for (const opcion of opciones) {
      expect(opcion.deshabilitado).toBe(ORIGENES_DE_DIRECTORIO.includes(opcion.valor));
    }

    expect(opciones.find((opcion) => opcion.valor === 'local')?.deshabilitado).toBe(false);
  });

  it('Soporte no puede repartir papeles: sólo se le ofrece el suyo de alta', () => {
    expect(opcionesDePapel(ES, ['usuario']).map((opcion) => opcion.valor)).toEqual(['usuario']);
    expect(opcionesDePapel(ES, ['usuario', 'soporte', 'desarrollo', 'administrador'])).toHaveLength(
      4,
    );
  });

  it('los dos idiomas se ofrecen, cada uno en el suyo', () => {
    expect(opcionesDeIdioma(ES).map((opcion) => opcion.etiqueta)).toEqual(['Español', 'English']);
  });
});

describe('fechaCorta', () => {
  it('escribe la fecha como se lee, no como la manda el backend', () => {
    const escrita = fechaCorta('2026-09-24T16:30:00Z', 'es');

    // El backend manda ISO; la pantalla no enseña ISO. Y el año va entero: «24/9/26» se confunde.
    expect(escrita).toMatch(/\d{2}\/\d{2}\/\d{4}/);
    expect(escrita).toContain('2026');
    expect(escrita).not.toContain('T');
    expect(escrita).not.toContain('Z');
  });

  it('quien no ha entrado nunca no tiene fecha, y la pantalla pone su texto', () => {
    expect(fechaCorta(undefined, 'es')).toBe('');
    expect(fechaCorta('', 'es')).toBe('');
    expect(fechaCorta('esto-no-es-una-fecha', 'es')).toBe('');
  });

  // **La zona de la instalación manda**: la instalación tiene una zona configurable —un nombre IANA—
  // y es la que decide cómo se leen todas las fechas. La última entrada se lee a la hora que es donde
  // está la mesa de ayuda, y no a la de quien mira.
  it('una marca en UTC se lee con el día que es en la zona de la instalación', () => {
    // `2026-09-30T02:00:00Z` es **el 29 a las 21:00** en Guayaquil: el día cambia, que es justo el
    // fallo que esto arregla.
    expect(fechaCorta('2026-09-30T02:00:00Z', 'es', 'America/Guayaquil')).toBe('29/09/2026, 21:00');
    expect(fechaCorta('2026-09-30T02:00:00Z', 'en', 'America/Guayaquil')).toBe('29/09/2026, 21:00');
  });

  it('la zona de la instalación manda sobre la del navegador de quien mira', () => {
    // La misma marca leída en dos zonas de la instalación: el navegador sólo puede estar en una, así
    // que al menos una de las dos lecturas demuestra que se usa la zona que se le pasa y no la suya.
    expect(fechaCorta('2026-09-30T02:00:00Z', 'es', 'America/Guayaquil')).toBe('29/09/2026, 21:00');
    expect(fechaCorta('2026-09-30T02:00:00Z', 'es', 'Pacific/Kiritimati')).toBe('30/09/2026, 16:00');
  });

  // La marca es pública: si la zona no llega —o llega un nombre que `Intl` no conoce— se lee con la
  // del navegador, que es lo que se hacía antes, en vez de dejar la lista sin pintar.
  it('sin zona se cae a la del navegador, y no revienta', () => {
    const iso = '2026-09-30T02:00:00Z';
    const enElNavegador = new Intl.DateTimeFormat('es-ES', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(iso));

    expect(fechaCorta(iso, 'es')).toBe(enElNavegador);
    expect(fechaCorta(iso, 'es', '')).toBe(enElNavegador);
    expect(fechaCorta(iso, 'es', '   ')).toBe(enElNavegador);
    expect(fechaCorta(iso, 'es', 'Zona/Que-No-Existe')).toBe(enElNavegador);
  });

  // **La interfaz no ofrece lo que no se puede hacer**: la reactivación a mano de una cuenta de
  // Keycloak se rechaza —su camino no tiene forma de comprobar quién sigue allí sin una cuenta de
  // servicio en el reino—, así que en su ficha se cuenta lo que pasa en vez de ofrecer el botón
  // (docs/modules/users.md, sección 5, punto 4).
  it('una cuenta de Keycloak desactivada se reactiva sola, y las demás no', () => {
    const cuenta = (origin: string, isActive: boolean): Cuenta =>
      ({
        id: 1,
        name: 'Ana',
        lastName: 'Pérez',
        email: 'ana@ejemplo.com',
        role: 'usuario',
        origin,
        language: 'es',
        isActive,
        hasPassword: false,
      }) as Cuenta;

    expect(seReactivaSola(cuenta('keycloak', false))).toBe(true);
    // En AD sí se ofrece el botón: allí se pregunta al directorio antes de devolver el acceso.
    expect(seReactivaSola(cuenta('ad', false))).toBe(false);
    expect(seReactivaSola(cuenta('local', false))).toBe(false);
    // Y una cuenta activa no se reactiva: ya lo está.
    expect(seReactivaSola(cuenta('keycloak', true))).toBe(false);
  });
});
