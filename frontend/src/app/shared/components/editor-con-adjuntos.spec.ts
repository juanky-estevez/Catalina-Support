import {
  ATRIBUTO_DE_LA_MENCION,
  CLASE_DE_LO_QUE_FALTA,
  direccionAdmitida,
  esImagen,
  esVideo,
  extensionAdmitida,
  htmlGuardable,
  colocacionDeLasSugerencias,
  mencionEscrita,
  textoComoHtml,
} from './editor-con-adjuntos';

/**
 * Lo que el editor **saca para guardar**, que es la pieza que más se puede torcer de todo esto: el
 * texto que se manda tiene que llevar las referencias a los adjuntos por su nombre y **ninguna
 * dirección**, y sólo los atributos de la lista blanca del backend
 * (`docs/modules/tickets.md`, sección 2.3). Si se cuela el `blob:` de la vista previa o una clase, el
 * backend rechaza el comentario entero con `tickets.body.notAllowed`.
 *
 * Se prueba sobre nodos de verdad, con el HTML montado en un contenedor: es exactamente lo que hay en
 * el lienzo cuando alguien escribe.
 */
function guardable(html: string): string {
  const contenedor = document.createElement('div');
  contenedor.innerHTML = html;

  return htmlGuardable(contenedor);
}

describe('Las extensiones que se admiten', () => {
  it('admite las cuatro de vídeo, que entraron el 2026-09-26', () => {
    expect(['x.mp4', 'x.webm', 'x.mov', 'x.avi'].every(extensionAdmitida)).toBe(true);
  });

  it('admite lo de siempre: imágenes, PDF, Office, texto y comprimidos', () => {
    const admitidas = [
      'a.png',
      'a.jpg',
      'a.jpeg',
      'a.gif',
      'a.webp',
      'a.pdf',
      'a.docx',
      'a.xlsx',
      'a.txt',
      'a.zip',
    ];

    expect(admitidas.every(extensionAdmitida)).toBe(true);
  });

  it('admite el texto y el código, que entraron el 2026-09-26', () => {
    // Es lo que se manda cuando el caso es una consulta que falla o un despliegue que no arranca
    // (decisión 54). **No se pintan nunca**: se descargan.
    const admitidas = [
      'consulta.sql',
      'datos.json',
      'config.xml',
      'docker-compose.yml',
      'parametros.yaml',
      'php.ini',
      'nginx.conf',
      'despliegue.sh',
      'script.py',
      'app.js',
      'app.ts',
      'Consulta.java',
      '.htaccess',
      'copia.bak',
      'registro.log',
      'paquete.tar',
      'paquete.gz',
      'paquete.7z',
    ];

    expect(admitidas.every(extensionAdmitida)).toBe(true);
  });

  it('lo que no se admite sigue sin admitirse', () => {
    expect(extensionAdmitida('dibujo.svg')).toBe(false);
    expect(extensionAdmitida('instalador.exe')).toBe(false);
    expect(extensionAdmitida('pagina.html')).toBe(false);
    expect(extensionAdmitida('sin-extension')).toBe(false);
  });

  it('reconoce lo que se ve en línea', () => {
    expect(esImagen('captura.PNG')).toBe(true);
    expect(esImagen('informe.pdf')).toBe(false);
    expect(esVideo('grabacion.mov')).toBe(true);
    expect(esVideo('captura.png')).toBe(false);
  });
});

describe('Las direcciones de los enlaces', () => {
  it('admite http, https y mailto', () => {
    expect(direccionAdmitida('https://ejemplo.com')).toBe(true);
    expect(direccionAdmitida('http://ejemplo.com')).toBe(true);
    expect(direccionAdmitida('mailto:soporte@ejemplo.com')).toBe(true);
  });

  it('no admite lo que ejecuta código ni las direcciones provisionales', () => {
    expect(direccionAdmitida('javascript:alert(1)')).toBe(false);
    expect(direccionAdmitida('data:text/html,<b>hola</b>')).toBe(false);
    expect(direccionAdmitida('blob:http://localhost/1234')).toBe(false);
    expect(direccionAdmitida('')).toBe(false);
  });
});

describe('El HTML que se guarda', () => {
  it('deja la imagen con su referencia y le quita la dirección y la clase', () => {
    const guardado = guardable(
      '<p>Mira:</p><img data-adjunto="captura.png" src="blob:http://localhost/1234" class="max-w-full rounded-md border border-borde" alt="captura.png">',
    );

    expect(guardado).toBe('<p>Mira:</p><img data-adjunto="captura.png">');
  });

  it('deja el vídeo con su referencia, sin `src` y **sin `controls`**', () => {
    // `controls` lo pone la pantalla al pintar: el atributo no está en la lista blanca del backend.
    const guardado = guardable(
      '<video data-adjunto="grabacion.mp4" src="blob:http://localhost/5678" controls class="max-w-full"></video>',
    );

    expect(guardado).toBe('<video data-adjunto="grabacion.mp4"></video>');
  });

  it('deja el enlace de un adjunto por su nombre', () => {
    const guardado = guardable(
      '<a data-adjunto="informe.pdf" href="blob:http://localhost/9999" class="text-primario underline">informe.pdf</a>',
    );

    expect(guardado).toBe('<a data-adjunto="informe.pdf">informe.pdf</a>');
  });

  it('conserva un enlace de verdad, y se queda con su texto si la dirección no vale', () => {
    expect(guardable('<a href="https://ejemplo.com">Ejemplo</a>')).toBe(
      '<a href="https://ejemplo.com">Ejemplo</a>',
    );
    expect(guardable('<a href="javascript:alert(1)">Pincha</a>')).toBe('Pincha');
  });

  it('no deja pasar ni el estilo ni una clase ni un atributo que empiece por `on`', () => {
    const guardado = guardable('<p class="cuerpo" style="color: red" onclick="alert(1)">Hola</p>');

    expect(guardado).toBe('<p>Hola</p>');
  });

  it('tira un `<script>` con lo que lleve dentro, y no sólo la etiqueta', () => {
    expect(guardable('<p>Hola</p><script>alert(1)</script>')).toBe('<p>Hola</p>');
  });

  it('desenvuelve lo que no está en la lista: el texto se queda y la etiqueta se va', () => {
    expect(guardable('<span>Hola</span>')).toBe('Hola');
    expect(guardable('<font color="red">Hola</font>')).toBe('Hola');
  });

  it('guarda el formato con los nombres de la lista blanca', () => {
    expect(guardable('<b>Negrita</b><i>Cursiva</i><strike>Tachado</strike>')).toBe(
      '<strong>Negrita</strong><em>Cursiva</em><s>Tachado</s>',
    );
    expect(guardable('<div>Uno</div><div>Dos</div>')).toBe('<p>Uno</p><p>Dos</p>');
    expect(guardable('<ul><li>Uno</li></ul><ol><li>Dos</li></ol><br>')).toBe(
      '<ul><li>Uno</li></ul><ol><li>Dos</li></ol><br>',
    );
  });

  it('escapa lo que el HTML se come', () => {
    expect(guardable('a < b & c')).toBe('a &lt; b &amp; c');
  });

  it('no guarda la marca de un adjunto que no subió: es de la pantalla, no del texto', () => {
    expect(
      guardable(`<p>Hola</p><span class="${CLASE_DE_LO_QUE_FALTA}">No subió: x.png</span>`),
    ).toBe('<p>Hola</p>');
  });
});

describe('El pegado de texto', () => {
  it('respeta los saltos de línea y escapa las etiquetas', () => {
    expect(textoComoHtml('Uno\nDos <tres> & cuatro')).toBe('Uno<br>Dos &lt;tres&gt; &amp; cuatro');
  });
});

/* --------------------------------------------------------------------------------------------
 * La lista blanca del backend, comprobada sobre lo que sale del editor
 * ------------------------------------------------------------------------------------------ */

/** Las etiquetas que el saneador del backend admite. **Ninguna más**. */
const ETIQUETAS = [
  'p',
  'br',
  'strong',
  'b',
  'em',
  'i',
  'u',
  's',
  'strike',
  'ul',
  'ol',
  'li',
  'a',
  'img',
  'video',
];

/** Y los atributos que admite, por etiqueta: `img` y `video` sólo su referencia, y `a` una de las dos. */
function atributosAdmitidos(etiqueta: string, atributo: string, valor: string): boolean {
  if (etiqueta === 'img' || etiqueta === 'video') {
    return atributo === 'data-adjunto' && valor.trim() !== '';
  }

  if (etiqueta === 'a') {
    if (atributo === 'data-adjunto') {
      return valor.trim() !== '';
    }

    return (
      atributo === 'href' &&
      (valor.startsWith('http://') || valor.startsWith('https://') || valor.startsWith('mailto:'))
    );
  }

  return false;
}

/**
 * Un texto del lienzo con todo lo que puede colarse, para comprobar que lo que sale pasa la puerta del
 * backend: es un espejo de las reglas de `backend/modules/tickets/services/body.go`, y está aquí
 * porque **un `class` de la vista previa haría que el comentario entero se rechazara** con
 * `tickets.body.notAllowed`.
 */
const LIENZO_SUCIO = [
  '<p class="cuerpo" style="color: red" onclick="alert(1)">Hola</p>',
  '<div>Un renglón</div>',
  '<span>Sin etiqueta</span><font color="red">Vieja</font>',
  '<b>Negrita</b><i>Cursiva</i><strike>Tachado</strike><u>Subrayado</u>',
  '<ul><li>Uno</li></ul><ol><li>Dos</li></ol><br>',
  '<img data-adjunto="captura.png" src="blob:https://ejemplo/1" class="max-w-full rounded-md border" alt="captura.png">',
  '<img src="https://fuera.example/contador.gif" onerror="alert(1)">',
  '<video data-adjunto="grabacion.mp4" src="blob:https://ejemplo/2" controls class="max-w-full"></video>',
  '<video src="blob:https://ejemplo/3"></video>',
  '<a data-adjunto="informe.pdf" href="blob:https://ejemplo/4" class="text-primario underline">informe.pdf</a>',
  '<a href="https://ejemplo.com">Un enlace</a>',
  '<a href="javascript:alert(1)">Malo</a>',
  '<a>Sin dirección ni adjunto</a>',
  '<script>alert(1)</script>',
  '<iframe src="https://fuera.example"></iframe>',
  '<img data-adjunto="captura.png">',
].join('');

describe('Lo que sale del editor pasa la lista blanca del backend', () => {
  const guardado = guardable(LIENZO_SUCIO);
  const contenedor = document.createElement('div');
  contenedor.innerHTML = guardado;

  it('no deja ninguna etiqueta que no esté en la lista', () => {
    const etiquetas = Array.from(contenedor.querySelectorAll('*')).map((nodo) =>
      nodo.tagName.toLowerCase(),
    );

    expect(etiquetas.filter((etiqueta) => !ETIQUETAS.includes(etiqueta))).toEqual([]);
  });

  it('no deja ningún atributo que el backend no admita, ni uno', () => {
    const sobran: string[] = [];

    for (const elemento of Array.from(contenedor.querySelectorAll('*'))) {
      const etiqueta = elemento.tagName.toLowerCase();

      for (const atributo of Array.from(elemento.attributes)) {
        if (!atributosAdmitidos(etiqueta, atributo.name, atributo.value)) {
          sobran.push(`${etiqueta}[${atributo.name}="${atributo.value}"]`);
        }
      }
    }

    expect(sobran, 'un atributo de más rechaza el comentario entero').toEqual([]);
  });

  it('deja la imagen, el vídeo y el PDF con su referencia, y sin direcciones', () => {
    expect(guardado).toContain('<img data-adjunto="captura.png">');
    expect(guardado).toContain('<video data-adjunto="grabacion.mp4"></video>');
    expect(guardado).toContain('<a data-adjunto="informe.pdf">informe.pdf</a>');
    expect(guardado).not.toContain('blob:');
    expect(guardado).not.toContain('src=');
  });

  it('no emite una imagen, un vídeo ni un enlace sin lo que el backend exige', () => {
    // Sin `data-adjunto`, la imagen y el vídeo **no salen**; y un enlace sin dirección válida se queda
    // con su texto, que es lo escrito, y pierde la etiqueta.
    expect(guardado).not.toContain('fuera.example');
    expect(guardado).not.toContain('javascript:');
    expect(guardado).toContain('Malo');
    expect(guardado).toContain('Sin dirección ni adjunto');
    expect(guardado).toContain('<a href="https://ejemplo.com">Un enlace</a>');

    const enlacesVacios = Array.from(contenedor.querySelectorAll('a')).filter(
      (enlace) => !enlace.hasAttribute('href') && !enlace.hasAttribute('data-adjunto'),
    );

    expect(enlacesVacios).toEqual([]);
  });

  it('deja el formato y los renglones, que es lo que sí admite', () => {
    expect(guardado).toContain('<p>Hola</p>');
    expect(guardado).toContain('<p>Un renglón</p>');
    expect(guardado).toContain(
      '<strong>Negrita</strong><em>Cursiva</em><s>Tachado</s><u>Subrayado</u>',
    );
    expect(guardado).toContain('<ul><li>Uno</li></ul><ol><li>Dos</li></ol><br>');
    expect(guardado).toContain('Sin etiqueta');
    expect(guardado).toContain('Vieja');
  });
});

describe('las menciones dentro del texto', () => {
  // La mención se guarda **por identificador y no por nombre** (docs/modules/tickets.md, decisión 58):
  // el nombre se lee, pero el dato es el identificador, y es lo que hace que un cambio de apellidos no
  // rompa la observación del ticket.
  const enUnContenedor = (html: string): HTMLElement => {
    const contenedor = document.createElement('div');
    contenedor.innerHTML = html;

    return contenedor;
  };

  it('el identificador de la persona se conserva, y el nombre que se lee también', () => {
    const contenedor = enUnContenedor(
      `<p>Mira esto, <span ${ATRIBUTO_DE_LA_MENCION}="12">María Pérez</span></p>`,
    );

    expect(htmlGuardable(contenedor)).toBe(
      `<p>Mira esto, <span ${ATRIBUTO_DE_LA_MENCION}="12">María Pérez</span></p>`,
    );
  });

  it('las clases y lo que se pegue de más se caen, y el nombre se queda', () => {
    // Es el caso de copiar un comentario de otra pantalla: llega con `class`, con `style` y hasta con
    // un `data-` que no toca. Nada de eso está en la lista blanca, así que se va; el nombre se queda.
    const contenedor = enUnContenedor(
      `<span ${ATRIBUTO_DE_LA_MENCION}="12" class="text-primario" style="color:red">María</span>`,
    );

    expect(htmlGuardable(contenedor)).toBe(`<span ${ATRIBUTO_DE_LA_MENCION}="12">María</span>`);
  });

  it('una mención sin identificador, o con uno que no lo es, se desenvuelve', () => {
    // El backend la rechazaría entera (`tickets.body.notAllowed`), así que aquí se quita la etiqueta y
    // se queda el texto: es lo mismo que se hace con cualquier etiqueta que no esté en la lista.
    expect(htmlGuardable(enUnContenedor('<span>María</span>'))).toBe('María');
    expect(
      htmlGuardable(enUnContenedor(`<span ${ATRIBUTO_DE_LA_MENCION}="abc">María</span>`)),
    ).toBe('María');
    expect(htmlGuardable(enUnContenedor(`<span ${ATRIBUTO_DE_LA_MENCION}="0">María</span>`))).toBe(
      'María',
    );
  });
});

describe('la mención que se escribe con arroba', () => {
  // Es la regla que decide cuándo sale la lista de a quién etiquetar mientras se teclea
  // (docs/interfaz-y-experiencia.md, decisión 74).
  it('con una arroba a medias, devuelve lo que se lleva escrito', () => {
    expect(mencionEscrita('Por favor @sop')).toBe('sop');
    expect(mencionEscrita('Por favor @')).toBe('');
    expect(mencionEscrita('@soporte2@demo.com')).toBe('soporte2@demo.com');
  });

  it('sin arroba, o con la mención ya terminada, no hay nada a medias', () => {
    expect(mencionEscrita('Por favor, míralo')).toBeNull();
    expect(mencionEscrita('Por favor @soporte2 ayúdame')).toBeNull();
    expect(mencionEscrita('')).toBeNull();
  });

  it('el correo entero vale como búsqueda, y una arroba detrás de otra no encuentra a nadie', () => {
    // El correo entero, que es como lo escribe el responsable: la arroba de la mención y la del correo.
    expect(mencionEscrita('@soporte2@demo.com')).toBe('soporte2@demo.com');
    expect(mencionEscrita('Por favor @soporte2@demo.com')).toBe('soporte2@demo.com');

    // Dos arrobas seguidas: la búsqueda es la segunda, y no encuentra a nadie (la lista no sale).
    expect(mencionEscrita('@@')).toBe('@');
  });
});

describe('dónde se pone la lista de a quién etiquetar', () => {
  // Lo que se pidió el 2026-09-29 (decisión 89): **junto al cursor**, no encima del bloque del comentario.
  const ventana = { ancho: 1200, alto: 800 };
  const cursor = { izquierda: 300, arriba: 400, abajo: 420 };

  it('debajo del cursor, cuando cabe', () => {
    const sitio = colocacionDeLasSugerencias(cursor, 200, ventana);

    expect(sitio.arriba).toBe(424);
    expect(sitio.izquierda).toBe(300);
    expect(sitio.maxAlto).toBe(200);
  });

  it('encima, cuando no cabe debajo', () => {
    const abajo = { izquierda: 300, arriba: 700, abajo: 720 };
    const sitio = colocacionDeLasSugerencias(abajo, 200, ventana);

    expect(sitio.arriba).toBe(496);
  });

  it('y nunca se sale por la derecha', () => {
    const aLaDerecha = { izquierda: 1150, arriba: 400, abajo: 420 };
    const sitio = colocacionDeLasSugerencias(aLaDerecha, 200, ventana);

    expect(sitio.izquierda).toBe(1200 - 320 - 8);
  });
});
