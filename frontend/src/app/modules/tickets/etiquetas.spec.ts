import { ES } from '../../core/i18n/es';
import {
  ESTADOS_DEL_INTERNO,
  ESTADOS_DEL_PRINCIPAL,
  etiquetaDeEstado,
  etiquetasDeTexto,
  etiquetasUnicas,
  fechaHora,
  interpolar,
  nombreDe,
  normalizarAlEscribir,
  normalizarEtiqueta,
  opcionesDeCategoria,
  opcionesDeEstado,
  peso,
  puedeEditarElTicket,
  puedeEtiquetarElTicket,
  sePuedePrevisualizar,
  textoDelResumen,
} from './etiquetas';

describe('las etiquetas del módulo tickets', () => {
  it('el usuario lee sus estados, y **no lee jerga interna**', () => {
    // Para el usuario, `escalado` es «En curso»: no tiene por qué saber que ha pasado a Desarrollo
    // (docs/interfaz-y-experiencia.md, sección 3.5).
    expect(etiquetaDeEstado(ES, 'nuevo', true)).toBe('Recibido');
    expect(etiquetaDeEstado(ES, 'escalado', true)).toBe('En curso');
    expect(etiquetaDeEstado(ES, 'en espera', true)).toBe('Esperamos tu respuesta');
    expect(etiquetaDeEstado(ES, 'escalado', true)).not.toContain('escal');
  });

  it('quien trabaja con tickets ve el estado de verdad', () => {
    expect(etiquetaDeEstado(ES, 'escalado', false)).toBe('Escalado');
    expect(etiquetaDeEstado(ES, 'en progreso', false)).toBe('En progreso');
  });

  it('un estado que no conocemos se enseña tal cual, en vez de desaparecer', () => {
    expect(etiquetaDeEstado(ES, 'cancelado', false)).toBe('cancelado');
  });

  it('el interno no tiene `escalado` entre sus estados: el interno **es** la escalación', () => {
    expect(ESTADOS_DEL_PRINCIPAL).toContain('escalado');
    expect(ESTADOS_DEL_INTERNO).not.toContain('escalado');
    expect(ESTADOS_DEL_INTERNO.length).toBe(ESTADOS_DEL_PRINCIPAL.length - 1);
  });

  it('los chips del estado llevan «todos» delante y los del tipo de ticket', () => {
    const delPrincipal = opcionesDeEstado(ES, false);

    expect(delPrincipal[0].valor).toBe('');
    expect(delPrincipal.map((opcion) => opcion.valor)).toContain('escalado');

    const delInterno = opcionesDeEstado(ES, true);
    expect(delInterno.map((opcion) => opcion.valor)).not.toContain('escalado');
  });

  it('sólo se previsualizan las imágenes y los PDF: lo demás se descarga', () => {
    for (const nombre of ['captura.png', 'foto.JPG', 'informe.pdf', 'dibujo.webp']) {
      expect(sePuedePrevisualizar(nombre)).toBe(true);
    }

    for (const nombre of ['hoja.xlsx', 'notas.txt', 'comprimido.zip', 'dibujo.svg']) {
      expect(sePuedePrevisualizar(nombre)).toBe(false);
    }
  });

  it('el peso se lee en algo que se entiende', () => {
    expect(peso(300)).toBe('300 B');
    expect(peso(2048)).toBe('2 KB');
    expect(peso(2 * 1024 * 1024)).toBe('2.0 MB');
  });
});

describe('las frases y las fechas', () => {
  it('una frase se rellena con sus datos, sin montarla a trozos', () => {
    expect(interpolar('{quien} lo pasó a «{estado}»', { quien: 'Ana', estado: 'En curso' })).toBe(
      'Ana lo pasó a «En curso»',
    );
  });

  it('un hueco sin dato se queda vacío, no con llaves', () => {
    expect(interpolar('{quien} lo cerró', {})).toBe(' lo cerró');
  });

  it('la fecha va con su hora, el año entero y en el idioma de quien mira', () => {
    const escrita = fechaHora('2026-09-24T16:30:00Z', 'es');

    expect(escrita).toMatch(/\d{2}\/\d{2}\/\d{4}/);
    expect(escrita).not.toContain('T');
  });

  // **La zona de la instalación manda**: la instalación tiene una zona configurable —un nombre IANA—
  // y es la que decide cómo se leen todas las fechas. Las guardadas siguen en UTC; lo que cambia es
  // la hora a la que se leen, y no la de quien mira.
  it('una marca en UTC se lee con el día que es en la zona de la instalación', () => {
    // `2026-09-30T02:00:00Z` es **el 29 a las 21:00** en Guayaquil: el día cambia, que es justo el
    // fallo que esto arregla.
    expect(fechaHora('2026-09-30T02:00:00Z', 'es', 'America/Guayaquil')).toBe('29/09/2026, 21:00');
    expect(fechaHora('2026-09-30T02:00:00Z', 'en', 'America/Guayaquil')).toBe('29/09/2026, 21:00');
  });

  it('la zona de la instalación manda sobre la del navegador de quien mira', () => {
    // La misma marca leída en dos zonas de la instalación: el navegador sólo puede estar en una, así
    // que al menos una de las dos lecturas demuestra que se usa la zona que se le pasa y no la suya.
    expect(fechaHora('2026-09-30T02:00:00Z', 'es', 'America/Guayaquil')).toBe('29/09/2026, 21:00');
    expect(fechaHora('2026-09-30T02:00:00Z', 'es', 'Pacific/Kiritimati')).toBe('30/09/2026, 16:00');
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

    expect(fechaHora(iso, 'es')).toBe(enElNavegador);
    expect(fechaHora(iso, 'es', '')).toBe(enElNavegador);
    expect(fechaHora(iso, 'es', '   ')).toBe(enElNavegador);
    expect(fechaHora(iso, 'es', 'Zona/Que-No-Existe')).toBe(enElNavegador);
  });

  it('sin fecha no se inventa nada', () => {
    expect(fechaHora(undefined, 'es')).toBe('');
    expect(fechaHora('esto no es una fecha', 'es')).toBe('');
  });

  it('el nombre completo se arma con lo que hay', () => {
    expect(nombreDe({ name: 'Ana', lastName: 'Pérez' })).toBe('Ana Pérez');
    expect(nombreDe(undefined)).toBe('');
  });
});

describe('el texto de un resumen del motor de IA', () => {
  // La pantalla **nunca enseña una clave ni un hueco mudo**: cada estado se dice con palabras
  // (docs/modules/ai.md, decisiones 2 y 8), y el texto sale en el idioma de quien mira, que es por lo
  // que los dos idiomas viajan en la misma respuesta.
  const listo = { state: 'listo', es: 'No puede entrar', en: 'Cannot sign in' };

  it('con el resumen escrito, se lee en el idioma de quien mira', () => {
    expect(textoDelResumen(listo, 'es', ES)).toBe('No puede entrar');
    expect(textoDelResumen(listo, 'en', ES)).toBe('Cannot sign in');
  });

  it('mientras el motor escribe, lo dice', () => {
    expect(textoDelResumen({ state: 'pendiente' }, 'es', ES)).toBe('Generando…');
  });

  it('sin motor, lo dice; y si falló, también', () => {
    expect(textoDelResumen({ state: 'sin_motor' }, 'es', ES)).toBe('Sin motor de IA');
    expect(textoDelResumen({ state: 'error', errorKey: 'ai.invalid' }, 'es', ES)).toBe(
      'No se pudo resumir',
    );
  });

  it('un ticket que nunca ha pedido resumen, y uno con el texto vacío, dicen «—»', () => {
    expect(textoDelResumen(undefined, 'es', ES)).toBe('—');
    expect(textoDelResumen({ state: 'listo', es: '   ' }, 'es', ES)).toBe('—');
  });
});

describe('la normalización de una etiqueta', () => {
  // Es la regla de la decisión 68: minúsculas, espacios a guiones y sin acentos. El backend la vuelve
  // a aplicar, y por eso la pantalla puede enseñarla según se teclea: las dos puertas acaban igual.
  it('pasa a minúsculas y cambia los espacios por guiones', () => {
    expect(normalizarEtiqueta('Red Wifi')).toBe('red-wifi');
    expect(normalizarEtiqueta('LICENCIA   OFFICE')).toBe('licencia-office');
  });

  it('quita los acentos y la eñe', () => {
    expect(normalizarEtiqueta('Sesión Caída')).toBe('sesion-caida');
    expect(normalizarEtiqueta('Año Nuevo')).toBe('ano-nuevo');
  });

  it('lo que no sea letra, número o guion se cae', () => {
    expect(normalizarEtiqueta('red_wifi!')).toBe('redwifi');
    expect(normalizarEtiqueta('a@b#c')).toBe('abc');
  });

  it('junta los guiones repetidos y quita los de los extremos', () => {
    // `red--wifi-` y `red-wifi` son la misma etiqueta: dejar dos formas es justo lo que la
    // normalización evita.
    expect(normalizarEtiqueta('red--wifi-')).toBe('red-wifi');
    expect(normalizarEtiqueta('  Red  ')).toBe('red');
  });

  it('al teclear se conserva el guion del final, o la segunda palabra se pegaría a la primera', () => {
    // Si al teclear se recortara, escribir «red wifi» daría «red» y luego «redwifi». La forma final la
    // da `normalizarEtiqueta` al añadir la ficha.
    expect(normalizarAlEscribir('Red ')).toBe('red-');
    expect(normalizarAlEscribir('Red W')).toBe('red-w');
  });

  it('un texto se parte en sus etiquetas por espacios y comas', () => {
    // En el campo las palabras ya han pasado por `normalizarAlEscribir`, así que llegan unidas; esto es
    // para lo que se pega o se escribe de un tirón.
    expect(etiquetasDeTexto('Red Wifi, licencia-office')).toEqual(['red', 'wifi', 'licencia-office']);
    expect(etiquetasDeTexto('red-wifi, vpn')).toEqual(['red-wifi', 'vpn']);
    expect(etiquetasDeTexto('  ,  ')).toEqual([]);
  });

  it('las etiquetas repetidas no se añaden dos veces', () => {
    // La base tiene `UNIQUE (ticket_id, tag)`: dejarlas entrar sería pedir un error al backend.
    expect(etiquetasUnicas(['red-wifi'], ['Red Wifi', 'vpn', 'vpn'])).toEqual(['red-wifi', 'vpn']);
  });
});

describe('las opciones del selector de categoría', () => {
  it('llevan la categoría delante y su nombre como etiqueta', () => {
    const opciones = opcionesDeCategoria(ES, [
      { id: 1, name: 'General', active: true, tickets: 3 },
      { id: 2, name: 'Red', active: true, tickets: 0 },
    ]);

    expect(opciones).toEqual([
      { valor: '1', etiqueta: 'General', grupo: 'Categoría' },
      { valor: '2', etiqueta: 'Red', grupo: 'Categoría' },
    ]);
  });

  it('el alta le pone delante una opción vacía que no se puede elegir', () => {
    // Es lo que hace que la categoría sea obligatoria de verdad: arranca diciendo «Elige una».
    const opciones = opcionesDeCategoria(ES, [], {
      valor: '',
      etiqueta: 'Elige una categoría',
      grupo: 'Categoría',
      deshabilitado: true,
    });

    expect(opciones[0].deshabilitado).toBe(true);
  });
});

describe('quién puede editar el texto y la categoría de un ticket', () => {
  // Es la regla de la ficha (docs/interfaz-y-experiencia.md, sección 3.4), y de ella depende que el
  // selector de categoría y los lápices del texto se le ofrezcan a quien de verdad puede. **Las
  // etiquetas tienen su propia regla**, más abajo: las ponen los dos equipos (decisión 85).
  const base = {
    soloLectura: false,
    cerrado: false,
    interno: false,
    esSoporte: false,
    esUsuario: false,
    esSuyo: false,
  };

  it('Soporte puede, y el solicitante en el suyo', () => {
    expect(puedeEditarElTicket({ ...base, esSoporte: true })).toBe(true);
    expect(puedeEditarElTicket({ ...base, esUsuario: true, esSuyo: true })).toBe(true);
  });

  it('un usuario no edita el ticket de otra persona', () => {
    expect(puedeEditarElTicket({ ...base, esUsuario: true, esSuyo: false })).toBe(false);
  });

  it('el Administrador no toca, ni Desarrollo escribe en el principal', () => {
    expect(puedeEditarElTicket({ ...base, esSoporte: true, soloLectura: true })).toBe(false);
    // Desarrollo no es Soporte ni el solicitante: no le toca.
    expect(puedeEditarElTicket(base)).toBe(false);
  });

  it('un ticket cerrado no se edita, y el interno no se toca', () => {
    expect(puedeEditarElTicket({ ...base, esSoporte: true, cerrado: true })).toBe(false);
    expect(puedeEditarElTicket({ ...base, esSoporte: true, interno: true })).toBe(false);
  });
});

describe('quién puede etiquetar un ticket', () => {
  // Decisión 85: las etiquetas son del principal y las ponen los dos equipos, **cada uno desde el
  // ticket en el que trabaja** —Soporte en el principal y Desarrollo en el interno—. No es la regla de
  // editar: por eso tiene la suya.
  const base = {
    soloLectura: false,
    cerrado: false,
    interno: false,
    esSoporte: false,
    esDesarrollo: false,
  };

  it('Soporte etiqueta el principal, y Desarrollo el interno', () => {
    expect(puedeEtiquetarElTicket({ ...base, esSoporte: true })).toBe(true);
    expect(puedeEtiquetarElTicket({ ...base, interno: true, esDesarrollo: true })).toBe(true);
  });

  it('cada equipo en el ticket del otro no: Desarrollo no toca el principal ni Soporte el interno', () => {
    expect(puedeEtiquetarElTicket({ ...base, esDesarrollo: true })).toBe(false);
    expect(puedeEtiquetarElTicket({ ...base, interno: true, esSoporte: true })).toBe(false);
  });

  it('el Administrador mira y un cerrado no se etiqueta', () => {
    expect(puedeEtiquetarElTicket({ ...base, esSoporte: true, soloLectura: true })).toBe(false);
    expect(puedeEtiquetarElTicket({ ...base, esSoporte: true, cerrado: true })).toBe(false);
    expect(puedeEtiquetarElTicket({ ...base, interno: true, esDesarrollo: true, cerrado: true })).toBe(
      false,
    );
  });
});
