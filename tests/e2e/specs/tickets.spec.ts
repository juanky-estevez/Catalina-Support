import { readFileSync } from 'node:fs';

import { expect, test, type APIRequestContext, type Locator, type Page } from '@playwright/test';

import {
  abrirMenuSiEsCajon,
  asignarA,
  buscarEnLaLista,
  filtrarEnLaLista,
  crearCuentaLista,
  entrar,
  elegirEstado,
  entrarComo,
  FABRICA,
  salir,
  tokenDe,
  tokenDeFabrica,
} from '../ayudas';

/**
 * Las pantallas de tickets: la bandeja, el alta y el detalle
 * (`docs/interfaz-y-experiencia.md`, secciones 3.3, 3.4 y 3.7).
 *
 * Son recorridos completos y **encadenados**: el mismo ticket pasa por el usuario que lo abre, Soporte
 * que pregunta y escala, y Desarrollo que lo devuelve. Es el caso que ninguna prueba de unidad puede
 * contar, porque vive en los cuatro papeles a la vez.
 *
 * Los datos se montan por la API (las cuentas y, cuando hace falta, el ticket) porque lo que se está
 * probando aquí es la pantalla, no el alta: el alta tiene sus propias pruebas en el backend.
 */

/**
 * Los archivos de prueba, **de verdad**: una captura y un PDF que el navegador sabe abrir.
 *
 * Se leen del disco y no se inventan con bytes sueltos porque hay cosas que sólo se pueden comprobar
 * con un archivo real: que la imagen **se vea** (un `naturalWidth` mayor que cero) y que el visor
 * enseñe el PDF en vez de bajárselo.
 */
const CAPTURA = readFileSync('archivos/captura.png');
const INFORME = readFileSync('archivos/informe.pdf');

/** El mismo PNG, en base64, para lo que hay que construir dentro de la página: pegar y arrastrar. */
const CAPTURA_EN_BASE64 = CAPTURA.toString('base64');

/** El número que el detalle enseña en su título: es el ticket que se acaba de abrir. */
async function numeroEnPantalla(page: Page): Promise<string> {
  const titulo = page.getByRole('heading', { level: 1 });
  await expect(titulo).toBeVisible();

  const numero = (await titulo.textContent())?.trim() ?? '';
  expect(numero, 'la cabecera del ticket tiene que llevar su número').toMatch(/^(INT-)?[A-Z0-9]+-\d{4}-\d+/);

  return numero;
}

/**
 * Crea un ticket por la API y devuelve su número y su asunto.
 *
 * **El asunto lleva una marca única**: en desarrollo los tickets de cada pasada se acumulan —no existe
 * borrar un ticket—, y dos tickets con el mismo asunto hacen que cualquier búsqueda por texto
 * encuentre dos y la prueba falle por su cuenta, no por la pantalla.
 */
async function ticketDePrueba(
  request: APIRequestContext,
  token: string,
  asunto: string,
  descripcion = 'Creado por las pruebas de interfaz.',
): Promise<{ numero: string; asunto: string }> {
  const unico = `${asunto} ${Date.now().toString(36)}`;

  // **Todo ticket nace con categoría** (docs/modules/tickets.md, decisión 65): sin ella, la API lo
  // rechaza con `tickets.category.required`, así que el ayudante pide el catálogo y usa la que se
  // llame «General» —la que trae la migración— o, si no está, la primera activa.
  const catalogo = await request.get('/api/tickets/categories', {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(catalogo.status(), await catalogo.text()).toBe(200);

  const { categories } = (await catalogo.json()) as { categories: { id: number; name: string }[] };
  const categoria = categories.find((c) => c.name === 'General') ?? categories[0]!;

  const respuesta = await request.post('/api/tickets', {
    headers: { Authorization: `Bearer ${token}` },
    data: { subject: unico, description: descripcion, categoryId: categoria.id },
  });

  expect(respuesta.status(), await respuesta.text()).toBe(201);
  const cuerpo = (await respuesta.json()) as { ticket: { number: string } };

  return { numero: cuerpo.ticket.number, asunto: unico };
}

/** El detalle de un ticket, tal y como lo devuelve la API: es donde se mira lo que se ha guardado. */
async function detalleDe(
  request: APIRequestContext,
  token: string,
  numero: string,
): Promise<{
  ticket: {
    number: string;
    /** Los dos campos del motor de IA, con su estado y el idioma global (`docs/modules/ai.md`). */
    insights?: {
      motivo: { state: string; text?: string; language?: string };
      ultimaAccion: { state: string; text?: string; language?: string };
    };
  };
  comments: { id: number; body: string }[];
  attachments: { id: number; filename: string; commentId?: number | null }[];
}> {
  const respuesta = await request.get(`/api/tickets/${numero}`, {
    headers: { Authorization: `Bearer ${token}` },
  });

  expect(respuesta.status(), await respuesta.text()).toBe(200);

  return (await respuesta.json()) as never;
}

/**
 * Pega un archivo de verdad, **con un evento de pegado del navegador**: es lo que recibe la pantalla
 * cuando alguien pulsa `Ctrl+V` con una captura en el portapapeles.
 *
 * Se construye un `DataTransfer` con un `File` dentro y se dispara el evento sobre el campo que se
 * diga, que es exactamente lo que hace el navegador. No se llama a ningún método de la pantalla: si la
 * pantalla no busca los archivos donde el navegador los pone, esto no los encuentra.
 */
async function pegarUnArchivo(page: Page, etiquetaDelCampo: string, nombre: string): Promise<void> {
  const campo = page.getByLabel(etiquetaDelCampo);

  await campo.evaluate((nodo, datos) => {
    const transferencia = new DataTransfer();
    transferencia.items.add(
      new File([Uint8Array.from(atob(datos.base64), (letra) => letra.charCodeAt(0))], datos.nombre, {
        type: 'image/png',
      }),
    );

    nodo.dispatchEvent(
      new ClipboardEvent('paste', { clipboardData: transferencia, bubbles: true, cancelable: true }),
    );
  }, { base64: CAPTURA_EN_BASE64, nombre });
}

/**
 * Y lo suelta encima, con el evento del arrastre: es la tercera puerta, y la que usa quien tiene la
 * ventana del explorador al lado (`docs/modules/tickets.md`, decisión 43).
 */
async function soltarUnArchivo(page: Page, etiquetaDelCampo: string, nombre: string): Promise<void> {
  const campo = page.getByLabel(etiquetaDelCampo);

  await campo.evaluate((nodo, datos) => {
    const transferencia = new DataTransfer();
    transferencia.items.add(
      new File([Uint8Array.from(atob(datos.base64), (letra) => letra.charCodeAt(0))], datos.nombre, {
        type: 'image/png',
      }),
    );

    nodo.dispatchEvent(
      new DragEvent('drop', { dataTransfer: transferencia, bubbles: true, cancelable: true }),
    );
  }, { base64: CAPTURA_EN_BASE64, nombre });
}

/** El ancho de una imagen ya pintada: es lo que dice que **se ve de verdad** y no es un hueco. */
async function anchoDeLaImagen(imagen: Locator): Promise<{ natural: number; ancho: number }> {
  return imagen.evaluate((nodo) => ({
    natural: (nodo as HTMLImageElement).naturalWidth,
    ancho: nodo.getBoundingClientRect().width,
  }));
}

/** El cuerpo pintado de un mensaje de la conversación: el texto con formato, con sus adjuntos dentro. */
function cuerpoDe(page: Page, indice = -1): Locator {
  return page.locator('div.cuerpo-con-adjuntos').nth(indice);
}

/**
 * Lo que hay dentro del cuerpo, **en el orden del documento**: los trozos de texto y los adjuntos que
 * el texto nombra. Es lo que permite comprobar que una imagen quedó **entre** dos párrafos y no sólo
 * que aparezca.
 */
async function ordenDelCuerpo(
  cuerpo: Locator,
): Promise<{ orden: string[]; texto: string; adjuntos: string[] }> {
  return cuerpo.evaluate((nodo) => {
    const recorrido: string[] = [];

    const andar = (elemento: Element): void => {
      for (const hijo of Array.from(elemento.childNodes)) {
        if (hijo.nodeType === Node.TEXT_NODE) {
          const texto = (hijo.nodeValue ?? '').trim();
          if (texto) {
            recorrido.push(`texto:${texto}`);
          }
          continue;
        }

        if (hijo.nodeType !== Node.ELEMENT_NODE) {
          continue;
        }

        const etiqueta = (hijo as Element).tagName.toLowerCase();
        if (etiqueta === 'img' || etiqueta === 'video') {
          recorrido.push(`adjunto:${(hijo as Element).getAttribute('data-adjunto')}`);
          continue;
        }
        if (etiqueta === 'a' && (hijo as Element).hasAttribute('data-adjunto')) {
          recorrido.push(`adjunto:${(hijo as Element).getAttribute('data-adjunto')}`);
          continue;
        }

        andar(hijo as Element);
      }
    };

    andar(nodo);

    return {
      orden: recorrido,
      texto: recorrido.filter((parte) => parte.startsWith('texto:')).join('|'),
      adjuntos: recorrido.filter((parte) => parte.startsWith('adjunto:')),
    };
  }) as Promise<{ orden: string[]; texto: string; adjuntos: string[] }>;
}

// El botón de copiar usa el portapapeles del navegador: sin permiso, el navegador lo bloquea, y la
// pantalla lo dice en vez de callarse. Aquí se le da permiso para poder probar el camino bueno.
test.use({ permissions: ['clipboard-read', 'clipboard-write'] });

test.describe('El recorrido de un ticket', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se pueden crear cuentas');
  });

  test('Soporte revisa una mejora de IA antes de aplicarla y el usuario no recibe el botón', async ({
    page,
    request,
  }) => {
    const solicitante = await crearCuentaLista(request, { name: 'Lucía', lastName: 'Solicitante' });
    const soporte = await crearCuentaLista(request, { role: 'soporte', name: 'Mateo', lastName: 'Soporte' });
    const tokenSolicitante = await tokenDe(request, solicitante.email);
    const { numero } = await ticketDePrueba(
      request,
      tokenSolicitante,
      'Ayuda de redacción',
      '<p>La conexión se interrumpe cada pocos minutos.</p>',
    );

    await entrarComo(page, soporte.email);
    await page.goto(`/tickets/${numero}`);
    const editor = page.getByLabel('Escribe un comentario');
    await editor.fill('hola lucía mandame el error que aparece');

    const abrir = page.getByRole('button', { name: 'Mejorar con IA' });
    await expect(abrir).toBeVisible();
    await abrir.click();
    const dialogo = page.getByRole('dialog', { name: 'Mejorar la redacción con IA' });
    await expect(dialogo.getByLabel('Borrador')).toHaveValue('hola lucía mandame el error que aparece');
    await dialogo.getByLabel('Tonalidad').selectOption('empathetic');
    await dialogo.getByRole('button', { name: 'Mejorar', exact: true }).click();

    await expect(dialogo.getByLabel('Borrador')).toHaveValue('connection ok');
    // La propuesta permanece en el modal hasta que la persona la acepta.
    await expect(editor).toHaveText('hola lucía mandame el error que aparece');
    await dialogo.getByRole('button', { name: 'Usar este texto' }).click();
    await expect(editor).toHaveText('connection ok');

    // Aplicarla sólo cambia el borrador: todavía no existe ningún comentario nuevo en la API.
    expect((await detalleDe(request, tokenSolicitante, numero)).comments).toHaveLength(0);

    await salir(page);
    await entrarComo(page, solicitante.email);
    await page.goto(`/tickets/${numero}`);
    await expect(page.getByRole('button', { name: 'Mejorar con IA' })).toHaveCount(0);
  });

  test('el usuario abre un ticket, lo sigue, lo cierra y lo reabre', async ({ page, request }, info) => {
    const ana = await crearCuentaLista(request);

    await entrarComo(page, ana.email);

    // Un usuario entra en **sus tickets**, no en una pantalla de inicio: ya no hay inicio.
    await expect(page).toHaveURL(/\/tickets$/);

    // **El nombre del módulo, los filtros y «Nuevo ticket» van en la misma fila** (decisión del
    // responsable, 2026-09-28): juntarlos deja un renglón más de tabla a la vista. Se mide por su
    // centro: si estuvieran en dos renglones, el segundo estaría más abajo.
    const titulo = page.getByRole('heading', { name: 'Mis tickets' });
    const buscar = page.getByRole('button', { name: 'Buscar', exact: true });
    const nuevo = page.getByRole('button', { name: 'Nuevo ticket' }).first();

    await expect(titulo).toBeVisible();
    await expect(buscar).toBeVisible();
    await expect(nuevo).toBeVisible();

    // En un ancho de PC los tres comparten renglón; en móvil se reparten solos, y eso también es lo
    // que se pidió (que quepan, no que se solapen).
    if (info.project.name === 'pc') {
      const centros = await Promise.all(
        [titulo, buscar, nuevo].map(async (elemento) => {
          const caja = await elemento.boundingBox();

          return (caja?.y ?? 0) + (caja?.height ?? 0) / 2;
        }),
      );

      expect(
        Math.max(...centros) - Math.min(...centros),
        'el título, los filtros y «Nuevo ticket» tienen que estar en la misma fila',
      ).toBeLessThan(8);
    }

    // El alta es una pantalla, y lleva su botón desde la bandeja.
    await page.getByRole('button', { name: 'Nuevo ticket' }).first().click();
    await expect(page).toHaveURL(/\/tickets\/new$/);

    const marca = Date.now().toString(36);
    await page.getByLabel('Asunto', { exact: true }).fill(`No puedo entrar ${marca}`);
    await page.getByLabel('Descripción').fill('Desde esta mañana me da error al entrar.');

    // **La categoría es obligatoria** (docs/modules/tickets.md, decisión 65): sin elegirla, el alta no
    // crea el ticket. Se elige «General», que es la que trae la migración y siempre está.
    await page.getByLabel('Categoría').selectOption({ label: 'General' });

    // Un adjunto de verdad, por el campo de archivo: es el camino que sigue una persona. **Va dentro
    // del texto**, en el sitio donde está el cursor (docs/modules/tickets.md, sección 2.3).
    await page.locator('input[type=file]').setInputFiles({
      name: 'captura.png',
      mimeType: 'image/png',
      buffer: CAPTURA,
    });

    // Y se ve mientras se escribe: es lo que pidió el responsable.
    const enElEditor = page.getByLabel('Descripción').locator('img[data-adjunto="captura.png"]');
    await expect(enElEditor).toBeVisible();
    expect(
      (await anchoDeLaImagen(enElEditor)).natural,
      'la vista previa se ve al escribir',
    ).toBeGreaterThan(0);

    await page.getByRole('button', { name: 'Crear el ticket' }).click();

    // Al crearlo se va al ticket, que es lo que quiere ver quien acaba de abrirlo.
    await expect(page).toHaveURL(/\/tickets\/[A-Z0-9-]+$/);
    const numero = await numeroEnPantalla(page);

    // El estado que lee el usuario no es el del backend: es lenguaje llano.
    await expect(page.getByText('Recibido').first()).toBeVisible();
    await expect(page.getByText('escalado')).toHaveCount(0);

    // Y el adjunto está **dentro de la descripción**, a la vista y con su tamaño.
    const enLaDescripcion = cuerpoDe(page, 0).locator('img[data-adjunto="captura.png"]');
    await expect(enLaDescripcion).toBeVisible();
    expect((await anchoDeLaImagen(enLaDescripcion)).natural).toBeGreaterThan(0);

    // El usuario comenta: la conversación crece y sigue siendo suya.
    await page.getByLabel('Escribe un comentario').fill('¿Hay novedades?');
    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('¿Hay novedades?')).toBeVisible();

    // **Cerrar se elige por su estado y se confirma en la ventana** (decisiones 80 y 81), y **sin decir
    // por qué no se cierra**: se prueba primero a confirmar sin el comentario, y el ticket sigue abierto.
    await elegirEstado(page, numero, 'cerrado');
    await page.getByRole('button', { name: 'Confirmar', exact: true }).click();
    await expect(page.getByRole('dialog'), 'la ventana sigue abierta sin el comentario').toBeVisible();
    await expect(page.getByText('Este ticket está cerrado')).toHaveCount(0);

    await page.getByLabel('Comentario del cierre').fill('Lo cierro yo: ya me lo arreglaron.');
    await page.getByRole('button', { name: 'Confirmar', exact: true }).click();

    await expect(page.getByText('Cerrado').first()).toBeVisible();
    await expect(
      page.getByText('Lo cierro yo: ya me lo arreglaron.'),
      'el comentario del cierre queda en la conversación',
    ).toBeVisible();

    // Un ticket cerrado no admite comentarios: se explica por qué, y **reabrir se elige en el
    // desplegable**, que es donde viven las acciones del ticket.
    await expect(page.getByLabel('Escribe un comentario')).toHaveCount(0);
    await expect(page.getByText('Este ticket está cerrado')).toBeVisible();

    // **Reabrir también se confirma en su ventana** (decisión 80), y explica que se reabre.
    await elegirEstado(page, numero, 'en progreso');
    await page.getByRole('button', { name: 'Confirmar', exact: true }).click();
    await expect(page.getByLabel('Escribe un comentario')).toBeVisible();
    await expect(page.getByText('En curso').first()).toBeVisible();

    // Y el número se copia, que es lo que se dice en voz alta. **El botón es sólo un icono** y
    // confirma con un visto verde unos segundos (decisión del responsable, 2026-09-27): sin texto que
    // ocupe sitio, y con el nombre accesible cambiando, que es lo que oye quien no ve el color.
    const copiar = page.getByRole('button', { name: 'Copiar el número' });
    await expect(copiar.locator('svg[data-icono="copiar"]')).toBeVisible();

    // **Y no se confunde con el lápiz de editar** (decisión del responsable, 2026-09-28): antes eran dos
    // caracteres —`⧉` y `✏`— que a 16 px se parecían, y el segundo salía como emoji de colores según la
    // fuente. Se comprueba que **los dos dibujos son distintos**, no sólo que hay un icono.
    const editar = page.getByRole('button', { name: 'Editar el asunto' });
    const dibujoDeCopiar = await copiar.locator('svg').innerHTML();
    const dibujoDeEditar = await editar.locator('svg').innerHTML();

    expect(dibujoDeCopiar, 'el de copiar y el de editar no pueden ser el mismo dibujo').not.toBe(
      dibujoDeEditar,
    );

    await copiar.click();

    const copiado = page.getByRole('button', { name: 'Copiado' });
    await expect(copiado.locator('svg[data-icono="visto"]')).toBeVisible();
    await expect(copiado).toHaveClass(/text-exito/);

    // **Y el verde es el del tema**, comprobado en lo que el navegador pinta y no en el nombre de la
    // clase: se pregunta por el color de una sonda pintada con `var(--exito)` y se compara. Así vale
    // también con los otros siete temas.
    //
    // Se espera a que la transición termine: el botón tiene `transition`, así que la primera lectura
    // coge el color a medio camino entre el de antes y el verde (lo cazó esta prueba: dio `#1b5f38`
    // cuando el del tema es `#1b6b3a`).
    const verdeDelTema = await page.evaluate(() => {
      const sonda = document.createElement('span');
      sonda.style.color = 'var(--exito)';
      document.body.appendChild(sonda);
      const color = getComputedStyle(sonda).color;
      sonda.remove();

      return color;
    });
    await expect
      .poll(async () => copiado.evaluate((nodo) => getComputedStyle(nodo).color), {
        message: 'el visto no llega al verde del tema',
      })
      .toBe(verdeDelTema);

    // Y a los dos segundos vuelve solo a como estaba.
    await expect(
      page.getByRole('button', { name: 'Copiar el número' }).locator('svg[data-icono="copiar"]'),
      'a los dos segundos vuelve el icono de copiar',
    ).toBeVisible({ timeout: 5000 });

    expect(numero).toMatch(/^[A-Z0-9]+-\d{4}-\d+$/);
  });

  test('Soporte pregunta al usuario y escala; Desarrollo lo devuelve a su bandeja', async ({
    page,
    request,
  }) => {
    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const desarrollo = await crearCuentaLista(request, { role: 'desarrollo' });
    const ana = await crearCuentaLista(request);

    await entrarComo(page, soporte.email);

    // Soporte entra en la bandeja, y ve el ticket que ha abierto el usuario.
    await expect(page).toHaveURL(/\/tickets$/);
    // La bandeja de Soporte es **«Mis tickets»**: lo suyo, no todo lo que hay (decisión 52).
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();

    // El ticket lo abre el usuario, que es quien abre tickets: Soporte lo atiende.
    const { numero } = await ticketDePrueba(request, await tokenDe(request, ana.email), 'Escalado y devuelto');

    // **Se busca en «Tickets principales»**, que es la lista de todo lo que hay (decisión 52): el
    // ticket es del usuario y puede no ser suyo —el turno se lo dará a un técnico u otro—, así que
    // buscarlo en «Mis tickets» sería depender de a quién le tocó. Se busca por su número, que es como
    // se encuentra un ticket: en desarrollo hay muchos y el nuevo cae en la última página.
    await page.goto('/tickets/main');
    await expect(page.getByRole('heading', { name: 'Tickets principales' })).toBeVisible();
    await buscarEnLaLista(page, numero);
    await expect(page.getByRole('link', { name: numero })).toBeVisible();

    // Entra al detalle desde la lista.
    await page.getByRole('link', { name: numero }).first().click();
    await expect(page).toHaveURL(new RegExp(`/tickets/${numero}$`));

    // Primero se empieza a trabajarlo: nadie pregunta sin haberlo mirado, y la tabla de transiciones
    // no deja pasar de `nuevo` a `en espera`.
    // **Todo estado se confirma en su ventana** (decisión 80), también empezar.
    await elegirEstado(page, numero, 'en progreso');
    await page.getByRole('button', { name: 'Confirmar', exact: true }).click();
    await expect(page.getByText('En progreso').first()).toBeVisible();

    // **«En espera» sólo se confirma** (decisión 88): antes pedía el texto —la pregunta al usuario— y
    // ahora no pide ninguno. La ventana sigue diciendo lo que va a pasar.
    await elegirEstado(page, numero, 'en espera');
    await expect(page.getByRole('dialog')).toContainText('El ticket queda esperando la respuesta');
    await expect(page.getByLabel('La pregunta')).toHaveCount(0);
    await page.getByRole('button', { name: 'Confirmar', exact: true }).click();

    await expect(page.getByText('En espera').first()).toBeVisible();

    // Y escala: sin motivo no se escala, y el motivo es lo que se pide.
    await elegirEstado(page, numero, 'escalado');
    await expect(page.getByText('Sin el motivo')).toBeVisible();
    await page.getByLabel('Motivo').fill('Probado en local y en el navegador: falla el servidor de sesiones.');
    await page.getByRole('button', { name: 'Confirmar', exact: true }).click();

    await expect(page.getByText('Escalado').first()).toBeVisible();

    // Con el interno creado aparece **el conmutador de la vista**, y Soporte entra por el principal
    // (decisión del responsable, 2026-09-27). Se cambia a «Los dos» para ver el interno, que es lo que
    // hacía antes por defecto.
    const vista = page.getByRole('group', { name: 'Qué quieres ver' });
    await expect(vista).toBeVisible();
    await expect(page.getByText('Principal', { exact: true })).toBeVisible();

    // **Y el número y el estado salen una sola vez**: los pintaba la cabecera de la página y la del
    // propio ticket, y el responsable los vio dos veces en `CS-2026-0017` (decisión 74).
    await expect(page.getByRole('heading', { level: 1, name: numero })).toHaveCount(1);
    await expect(page.getByRole('heading', { name: numero })).toHaveCount(1);

    // **Y el botón de volver dice sólo «Volver»** (decisión del responsable, 2026-09-28).
    await expect(page.getByRole('button', { name: 'Volver', exact: true })).toBeVisible();

    await vista.getByRole('button', { name: 'Los dos' }).click();

    // **El número del interno sale en dos sitios, y los dos valen**: en la cabecera de la página —al
    // lado del principal— y en la cabecera de su columna (decisión del responsable, 2026-09-28).
    await expect(page.getByText(`INT-${numero}`).first()).toBeVisible();

    // **Y con los dos tickets a la vista hay un control de acciones por columna** (decisión 78), cada
    // uno con el suyo: desde un solo control de arriba no se sabría a cuál se le cambia el estado.
    await expect(page.locator(`#acciones-${numero}`)).toBeVisible();
    await expect(page.locator(`#acciones-INT-${numero}`)).toBeVisible();

    // Desarrollo entra en **lo suyo**, que son los internos, y el interno recién creado todavía no
    // es suyo: el reparto por turnos se lo habrá dado a otro desarrollador. Está en **Tickets
    // internos**, que es la lista de todo lo que hay (decisión 52).
    await entrarComo(page, desarrollo.email);
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();

    await page.goto('/tickets/internal');
    await expect(page.getByRole('heading', { name: 'Tickets internos' })).toBeVisible();
    await buscarEnLaLista(page, `INT-${numero}`);
    await expect(page.getByRole('link', { name: `INT-${numero}` })).toBeVisible();
    // En los internos no está el principal: son listas distintas, y esa es la gracia.
    await expect(page.getByRole('link', { name: numero, exact: true })).toHaveCount(0);

    await page.getByRole('link', { name: `INT-${numero}` }).first().click();
    await expect(page.getByText('Motivo del escalado:')).toBeVisible();

    // **Se lo asigna desde el propio ticket** (decisión 50): el atajo de la lista ya no existe, y
    // desde aquí se ve a quién se le pasa. Y con eso el interno pasa a estar en lo suyo.
    await asignarA(page, desarrollo.id);
    await expect(page.getByText('Ticket asignado.')).toBeVisible();

    await page.goto('/tickets');
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();
    await buscarEnLaLista(page, `INT-${numero}`);
    await expect(page.getByRole('link', { name: `INT-${numero}` })).toBeVisible();

    await page.getByRole('link', { name: `INT-${numero}` }).first().click();
    await expect(page.getByText('Motivo del escalado:')).toBeVisible();
    // **Acotado a la línea de tiempo**, y no a la página entera: desde el 2026-09-27 el motivo del
    // escalado se enseña también en la ficha del interno, y en esta prueba el texto del motivo y el de
    // la descripción son el mismo, así que buscar en toda la pantalla encuentra dos.
    await expect(page.getByRole('main').getByText('Probado en local y en el navegador').first()).toBeVisible();

    // **«No es un cambio de código» es «Cerrado» en un interno** (decisión 80): devuelve el caso a
    // Soporte y lo cierra, y el porqué va en el comentario obligatorio del cierre.
    await elegirEstado(page, `INT-${numero}`, 'cerrado');
    await page.getByLabel('Comentario del cierre').fill(
      'Es una duda de uso: se resuelve explicándole cómo entrar.',
    );
    await page.getByRole('button', { name: 'Confirmar', exact: true }).click();

    await expect(page.getByText('Cerrado').first()).toBeVisible();

    // Y el caso ha vuelto a Soporte: el principal ya no está escalado.
    await entrarComo(page, soporte.email);
    await page.goto(`/tickets/${numero}`);
    await expect(page.getByText('En progreso').first()).toBeVisible();
  });
});

/**
 * El aspecto, medido sobre lo que el navegador pinta.
 *
 * Una pantalla sin estilos pasa todas las pruebas de estructura: aquí se comprueba que el estado
 * lleve su color y su borde, que las tarjetas tengan el borde del tema, que la ficha esté **a la
 * derecha en PC y debajo en móvil** (sección 3.4) y que la tabla de la bandeja **no exista** en móvil,
 * donde se usan tarjetas (sección 7).
 */
test.describe('El aspecto del ticket', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se pueden crear cuentas');
    void page;
  });

  test('el estado lleva su color, la ficha se coloca donde toca y en móvil no hay tabla', async ({
    page,
    request,
  }, info) => {
    const ana = await crearCuentaLista(request);
    const { numero } = await ticketDePrueba(request, await tokenDe(request, ana.email), 'Aspecto');

    // La consola y las peticiones: una pantalla que pide algo que no existe se ve mal aunque el
    // texto esté.
    const fallos: string[] = [];
    page.on('requestfailed', (peticion) => fallos.push(peticion.url()));
    page.on('response', (respuesta) => {
      if (respuesta.status() >= 400) {
        fallos.push(`${respuesta.status()} ${respuesta.url()}`);
      }
    });

    await entrarComo(page, ana.email);

    // La bandeja en PC es una tabla; en móvil, tarjetas.
    if (info.project.name === 'pc') {
      await expect(page.locator('table')).toBeVisible();
      await expect(page.locator('table thead th').first()).toBeVisible();
    } else {
      await expect(page.locator('table')).toHaveCount(0);
      await expect(page.locator('li').first()).toBeVisible();
    }

    await page.goto(`/tickets/${numero}`);

    // La tarjeta de la descripción: con el borde y el fondo del tema, no transparente. El cuerpo del
    // texto va dentro, así que la tarjeta es su padre.
    const tarjeta = cuerpoDe(page, 0).locator('..');
    const estiloTarjeta = await tarjeta.evaluate((nodo) => {
      const estilo = getComputedStyle(nodo);
      return {
        borde: estilo.borderTopWidth,
        colorBorde: estilo.borderTopColor,
        fondo: estilo.backgroundColor,
      };
    });

    expect(estiloTarjeta.borde).toBe('1px');
    expect(estiloTarjeta.fondo).not.toBe('rgba(0, 0, 0, 0)');
    expect(estiloTarjeta.colorBorde).not.toBe('rgba(0, 0, 0, 0)');

    // El estado: con su texto y con su color, nunca sólo color.
    const estado = page.locator('app-estado span').first();
    await expect(estado).toHaveText('Recibido');

    const estiloEstado = await estado.evaluate((nodo) => {
      const estilo = getComputedStyle(nodo);
      return { borde: estilo.borderTopWidth, colorBorde: estilo.borderTopColor, color: estilo.color };
    });

    expect(estiloEstado.borde).toBe('1px');
    expect(estiloEstado.colorBorde).not.toBe('rgba(0, 0, 0, 0)');
    expect(estiloEstado.color).not.toBe('rgb(0, 0, 0)');

    // Y las dos columnas donde dice el documento: la ficha a la derecha en PC, debajo en móvil.
    const descripcion = await tarjeta.boundingBox();
    const ficha = await page.getByRole('heading', { name: 'La ficha' }).boundingBox();

    expect(descripcion).not.toBeNull();
    expect(ficha).not.toBeNull();

    if (info.project.name === 'pc') {
      expect(ficha!.x).toBeGreaterThan(descripcion!.x + descripcion!.width - 1);
    } else {
      expect(ficha!.y).toBeGreaterThan(descripcion!.y + descripcion!.height - 1);
    }

    // Y **el editor con formato se ve**: su barra, sus botones y su lienzo con el borde del campo. Una
    // pantalla sin estilos pasaría lo de arriba y fallaría aquí.
    const caja = page.getByLabel('Escribe un comentario');
    await expect(caja).toBeVisible();

    const barra = page.getByRole('group', { name: 'Formato del texto' });
    await expect(barra).toBeVisible();
    await expect(barra.getByRole('button', { name: 'Negrita' })).toBeVisible();
    await expect(barra.getByRole('button', { name: 'Adjuntar' })).toBeVisible();

    const estiloDelBoton = await barra
      .getByRole('button', { name: 'Negrita' })
      .evaluate((nodo) => {
        const estilo = getComputedStyle(nodo);
        return { borde: estilo.borderTopWidth, colorBorde: estilo.borderTopColor, alto: estilo.height };
      });

    expect(estiloDelBoton.borde).toBe('1px');
    expect(estiloDelBoton.colorBorde).not.toBe('rgba(0, 0, 0, 0)');
    expect(parseFloat(estiloDelBoton.alto), 'el botón tiene que poder pulsarse').toBeGreaterThanOrEqual(32);

    const estiloDelLienzo = await caja.evaluate((nodo) => {
      const estilo = getComputedStyle(nodo);
      return { borde: estilo.borderTopWidth, fondo: estilo.backgroundColor };
    });

    expect(estiloDelLienzo.borde).toBe('1px');
    expect(estiloDelLienzo.fondo).not.toBe('rgba(0, 0, 0, 0)');

    // Y **nada se sale de la pantalla**: la barra del editor, con sus botones y el de adjuntar, se
    // reparte en varias líneas en un móvil en vez de empujar la página a lo ancho.
    const desborde = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
    );
    expect(desborde, 'la pantalla no puede desbordarse a lo ancho').toBeLessThanOrEqual(1);

    expect(fallos, `peticiones que fallaron: ${fallos.join(', ')}`).toEqual([]);
  });
});

test.describe('La bandeja, por papel', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede entrar');
  });

  test('el Administrador mira los tickets: ve los dos tipos y no puede tocar nada', async ({
    page,
    request,
  }) => {
    // El ticket lo abre otra persona: el Administrador **no crea tickets** —lo dice la matriz— y por
    // eso la prueba no puede montarlo con su cuenta.
    const quienPide = await crearCuentaLista(request);

    await entrarComo(page, FABRICA.email, FABRICA.password);

    // El Administrador aterriza en los usuarios, y desde el menú se va a la bandeja.
    await expect(page).toHaveURL(/\/users$/);
    await page.goto('/tickets');
    await expect(page.getByRole('heading', { name: 'Bandeja' })).toBeVisible();

    // El chip de tipo es suyo —es el único papel que ve los dos—, y desde el 2026-09-27 vive **dentro
    // del modal de búsqueda**: se comprueba abriéndolo.
    await page.getByRole('button', { name: 'Buscar', exact: true }).click();
    await expect(page.getByRole('dialog').getByRole('group', { name: 'Tipo' })).toBeVisible();
    // La ventana se cierra con su ✕, que se llama «Cerrar» y no «Cancelar» (decisión 77): antes se
    // llamaba igual que el botón de cancelar de dentro de la ventana.
    await page.getByRole('dialog').getByRole('button', { name: 'Cerrar' }).click();

    const { numero, asunto } = await ticketDePrueba(
      request,
      await tokenDe(request, quienPide.email),
      'Un ticket para mirar',
    );

    await buscarEnLaLista(page, numero);
    await page.getByRole('link', { name: numero }).click();

    // Lee el ticket entero y **no se le ofrece nada**: ni comentar, ni mover, ni asignar.
    await expect(page.getByRole('heading', { level: 1, name: numero })).toBeVisible();
    await expect(page.getByText(asunto)).toBeVisible();
    await expect(page.getByText('No hay nada que hacer aquí')).toBeVisible();
    await expect(page.getByLabel('Escribe un comentario')).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Asignar', exact: true })).toHaveCount(0);

    // Y el chip de tipo lleva a los internos, que ahora vive dentro del modal de búsqueda.
    await page.goto('/tickets');
    await filtrarEnLaLista(page, 'Tipo', 'Internos');
    await expect(page.getByRole('heading', { name: 'Bandeja' })).toBeVisible();
  });
});

/**
 * **Los adjuntos de un comentario, con sus tres puertas**: pegar, arrastrar y elegir el archivo, y
 * **dentro del texto**, en el sitio donde está el cursor (`docs/modules/tickets.md`, sección 2.3 y
 * decisiones 41 a 49).
 *
 * El pegado y el arrastre se disparan **con eventos de verdad**, con su `DataTransfer` y su `File`
 * dentro, que es lo que recibe la pantalla cuando alguien pega una captura o la suelta encima. No se
 * llama a ningún método por dentro: si la pantalla no busca los archivos donde el navegador los pone,
 * esto falla.
 *
 * Las imágenes que se adjuntan son **PNG de verdad**, y por eso se puede comprobar que se ven
 * (`naturalWidth`): con bytes inventados la prueba pasaría sin que se viera nada.
 */
/**
 * **Las tres listas** (decisiones del responsable, 2026-09-26): «Mis tickets» es lo mío —lo asignado,
 * lo abierto por mí y aquello donde he comentado—, y las dos listas del «todo» son para Soporte y
 * Desarrollo.
 *
 * Es el caso que cuenta lo que ninguna prueba de unidad puede: que dos técnicos vean cosas distintas
 * de la misma base, y que **la reasignación se haga dentro del ticket** y mueva el ticket de una
 * bandeja a la otra.
 */
test.describe('Las listas de tickets', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se pueden crear cuentas');
    void page;
  });

  test('«Mis tickets» es lo que es mío, y lo que hay lo enseñan las dos listas del todo', async ({
    page,
    request,
  }) => {
    // **Un minuto**: el caso da de alta tres cuentas leyendo sus correos del buzón y busca cuatro
    // veces abriendo el modal, y con los treinta segundos de siempre se quedaba sin tiempo en la
    // última aserción (lo cazó la propia suite, no una persona).
    test.setTimeout(60_000);

    const soporteA = await crearCuentaLista(request, { role: 'soporte' });
    const soporteB = await crearCuentaLista(request, { role: 'soporte' });
    const quienPide = await crearCuentaLista(request);

    const { numero } = await ticketDePrueba(
      request,
      await tokenDe(request, quienPide.email),
      'Un ticket para repartir',
    );

    // **El ticket está en la lista de todo**, que es donde lo encuentra cualquiera del equipo. Se busca
    // por el número, que es como se busca un ticket.
    //
    // **No se afirma de quién es todavía**: el reparto por turnos se lo asigna a un técnico activo al
    // crearlo, y cuál de ellos depende del turno —con las cuentas que dejan las pruebas, cualquiera—,
    // así que «un ticket nuevo no es de nadie» no es verdad y dar por hecho lo contrario hacía que esta
    // prueba fallara según a quién le tocara el turno (pasó el 2026-09-27). Lo que sí es fijo es el
    // resto: se lo queda A, luego B se lo lleva, y a A se le va de su bandeja.
    await entrarComo(page, soporteA.email);
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();

    await page.goto('/tickets/main');
    await expect(page.getByRole('heading', { name: 'Tickets principales' })).toBeVisible();
    await buscarEnLaLista(page, numero);
    await expect(page.getByRole('link', { name: numero })).toBeVisible();

    // **La reasignación, dentro del ticket**: A se lo queda desde su ficha, que es donde se ve a quién
    // se le pasa. El atajo «Asignármelo» de la lista ya no existe (decisión 50). Se elige **por su
    // identificador** y no por su nombre: las cuentas de prueba se llaman todas igual.
    await page.getByRole('link', { name: numero }).click();
    await expect(page.getByRole('heading', { level: 1, name: numero })).toBeVisible();
    await asignarA(page, soporteA.id);
    await expect(page.getByText('Ticket asignado.')).toBeVisible();

    // Y ahora sí está en lo suyo: «lo mío» es lo que tengo asignado.
    await page.goto('/tickets');
    await buscarEnLaLista(page, numero);
    await expect(page.getByRole('link', { name: numero })).toBeVisible();

    // **Y el ejemplo del responsable**: un técnico se lo pasa a otro técnico. B entra por la lista de
    // todo, se lo asigna desde el ticket, y el ticket cambia de bandeja.
    await entrarComo(page, soporteB.email);
    await page.goto('/tickets/main');
    await buscarEnLaLista(page, numero);
    await page.getByRole('link', { name: numero }).click();
    await expect(page.getByRole('heading', { level: 1, name: numero })).toBeVisible();
    await asignarA(page, soporteB.id);
    await expect(page.getByText('Ticket asignado.')).toBeVisible();

    await page.goto('/tickets');
    await buscarEnLaLista(page, numero);
    await expect(page.getByRole('link', { name: numero })).toBeVisible();

    // Y a A se le ha ido de su bandeja: no lo tiene asignado, no lo abrió y no ha comentado en él.
    //
    // **Se navega a la lista después de entrar**, en vez de fiarse de dónde deja la entrada: en móvil,
    // rellenar la búsqueda antes de que la pantalla sea la de A encuentra la lista de la persona
    // anterior, y la prueba falla por una carrera y no por el producto (le pasó a esta misma prueba).
    await entrarComo(page, soporteA.email);
    await page.goto('/tickets');
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();
    await buscarEnLaLista(page, numero);
    await expect(page.getByRole('link', { name: numero })).toHaveCount(0);
  });

  test('el chip de «Mis tickets» lleva a los internos, y las dos listas del todo están en el menú', async ({
    page,
    request,
  }) => {
    const soporte = await crearCuentaLista(request, { role: 'soporte' });

    await entrarComo(page, soporte.email);

    // El chip de tipo lo lleva Soporte desde el 2026-09-26, arranca en los principales y desde el
    // 2026-09-27 se cambia **dentro del modal de búsqueda**.
    await page.getByRole('button', { name: 'Buscar', exact: true }).click();
    const chip = page.getByRole('dialog').getByRole('group', { name: 'Tipo' });
    await expect(chip.getByRole('button', { name: 'Principales' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );

    await chip.getByRole('button', { name: 'Internos' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Buscar', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();

    // Las dos listas del «todo» son suyas, y desde ellas no se crea un ticket: el botón vive en la
    // bandeja.
    await page.goto('/tickets/internal');
    await expect(page.getByRole('heading', { name: 'Tickets internos' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Nuevo ticket' })).toHaveCount(0);

    await page.goto('/tickets/main');
    await expect(page.getByRole('heading', { name: 'Tickets principales' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Nuevo ticket' })).toHaveCount(0);

    // Y la bandeja sí lo ofrece, que es donde se crea (dos veces: arriba y en la lista vacía).
    await page.goto('/tickets');
    await expect(page.getByRole('button', { name: 'Nuevo ticket' }).first()).toBeVisible();
  });

  test('el usuario no tiene chip de tipo ni ve las listas del todo', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);

    await entrarComo(page, ana.email);
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();

    // Sin chip: sus tickets son principales, y un filtro de una sola opción es ruido.
    await expect(page.getByRole('group', { name: 'Tipo' })).toHaveCount(0);

    await abrirMenuSiEsCajon(page);
    const menu = page.locator('aside');
    await expect(menu.getByRole('link', { name: 'Tickets principales' })).toHaveCount(0);
    await expect(menu.getByRole('link', { name: 'Tickets internos' })).toHaveCount(0);

    // Y si escribe la dirección a mano, la guarda lo lleva a «sin permiso».
    await page.goto('/tickets/main');
    await expect(page).toHaveURL(/\/forbidden$/);
  });
});

test.describe('Los adjuntos en los comentarios', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se pueden crear cuentas');
  });

  /**
   * **Las tres puertas meten la imagen donde está el cursor**, y se ve mientras se escribe.
   *
   * Es el caso que pidió el responsable: la vista previa se ve **también al editar**, y no como un
   * texto que representa al archivo (decisión 46).
   */
  test('las tres puertas meten la imagen en el texto, y se ve al escribir', async ({
    page,
    request,
  }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Las tres puertas');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    const caja = page.getByLabel('Escribe un comentario');
    await caja.fill('Primero.');

    // Puerta 1: el botón de adjuntar, que es un campo de archivo de verdad.
    await page.locator('input[type=file]').setInputFiles('archivos/captura.png');

    // Puerta 2: pegar, con el cursor en el texto.
    await pegarUnArchivo(page, 'Escribe un comentario', 'captura-pegada.png');

    // Puerta 3: arrastrar y soltar encima.
    await soltarUnArchivo(page, 'Escribe un comentario', 'captura-soltada.png');

    await caja.click();
    await page.keyboard.press('End');
    await page.keyboard.type(' Y despues.');

    // **Se ven de verdad mientras se escribe**, las tres, con su tamaño.
    await expect(caja.locator('img[data-adjunto]')).toHaveCount(3);
    for (const nombre of ['captura.png', 'captura-pegada.png', 'captura-soltada.png']) {
      const imagen = caja.locator(`img[data-adjunto="${nombre}"]`);
      await expect(imagen, `la imagen de la puerta «${nombre}» se ve al escribir`).toBeVisible();

      const medida = await anchoDeLaImagen(imagen);
      expect(medida.natural, `«${nombre}» no se ha pintado`).toBeGreaterThan(0);
      expect(medida.ancho).toBeGreaterThan(0);
    }

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    // Y después de guardar, en la conversación: las tres en línea, con su tamaño.
    //
    // **Con un tope de 480 × 360** (decisión 55, del 2026-09-26): una captura a pantalla completa
    // ocupaba todo el ancho de la conversación. Se **mide** lo que el navegador pinta, que es lo único
    // que distingue «tiene tope» de «casi lo tiene».
    const cuerpo = cuerpoDe(page);
    await expect(cuerpo.locator('img[data-adjunto]')).toHaveCount(3);
    for (const nombre of ['captura.png', 'captura-pegada.png', 'captura-soltada.png']) {
      const imagen = cuerpo.locator(`img[data-adjunto="${nombre}"]`);
      await expect(imagen).toBeVisible();
      expect((await anchoDeLaImagen(imagen)).natural).toBeGreaterThan(0);

      const pintada = await imagen.evaluate((nodo) => {
        const caja = nodo.getBoundingClientRect();
        return { ancho: caja.width, alto: caja.height };
      });

      expect(pintada.ancho, `${nombre} no puede pasar de 480 px`).toBeLessThanOrEqual(480);
      expect(pintada.alto, `${nombre} no puede pasar de 360 px`).toBeLessThanOrEqual(360);
    }

    // **Lo que se guarda son referencias por nombre, y ninguna dirección**: si viajara el `blob:` de la
    // vista previa, el backend rechazaría el comentario entero (`tickets.body.notAllowed`).
    const detalle = await detalleDe(request, suyo, numero);
    const guardado = detalle.comments[detalle.comments.length - 1].body;

    for (const nombre of ['captura.png', 'captura-pegada.png', 'captura-soltada.png']) {
      expect(guardado, `falta la referencia a «${nombre}»`).toContain(`data-adjunto="${nombre}"`);
    }
    expect(guardado, 'el texto guardado no puede llevar direcciones').not.toContain('src=');
    expect(guardado).not.toContain('blob:');
    expect(guardado, 'el texto guardado no puede llevar clases').not.toContain('class=');

    // Y la imagen se pulsa para verla a tamaño completo: el visor, con sus dos botones (decisión 48).
    await cuerpo.locator('img[data-adjunto="captura.png"]').click();

    const visor = page.locator('dialog[open]');
    await expect(visor).toBeVisible();
    await expect(visor.getByRole('button', { name: 'Descargar' })).toBeVisible();
    await expect(visor.getByRole('button', { name: 'Abrir en una pestaña' })).toBeVisible();

    const pestana = page.waitForEvent('popup');
    await visor.getByRole('button', { name: 'Abrir en una pestaña' }).click();
    expect((await pestana).url(), 'la pestaña lleva el archivo').toMatch(/^blob:/);
  });

  /**
   * **Un comentario puede ser sólo un archivo** (decisión 41): «mira esto» con la captura y sin
   * escribir nada. Lo que no se puede es no mandar nada, y eso se dice **antes** de enviarlo, que es
   * donde se sostiene la regla (los archivos suben después del comentario, así que el backend no puede
   * comprobarla).
   */
  test('un comentario puede ser sólo un archivo, y sin nada no se envía', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Sólo el archivo');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    // Sin texto, con su PDF: el archivo entra en el texto como su enlace, con su nombre.
    await page.locator('input[type=file]').setInputFiles('archivos/informe.pdf');
    await expect(page.getByLabel('Escribe un comentario').locator('a[data-adjunto="informe.pdf"]')).toHaveText(
      'informe.pdf',
    );

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    // El comentario existe **sin texto propio**: lo que lleva dentro es la referencia al archivo.
    const detalle = await detalleDe(request, suyo, numero);
    const comentario = detalle.comments[detalle.comments.length - 1];

    expect(comentario.body).toContain('data-adjunto="informe.pdf"');
    expect(
      detalle.attachments.some((adjunto) => adjunto.commentId === comentario.id),
      'el archivo cuelga de ese comentario',
    ).toBe(true);

    // Y en pantalla **no queda ningún cuerpo vacío**: el que no tiene texto se lee por sus adjuntos,
    // sin un hueco que parezca un fallo.
    const vacios = await page
      .locator('div.cuerpo-con-adjuntos')
      .evaluateAll((nodos) => nodos.filter((nodo) => (nodo.textContent ?? '').trim() === '').length);

    expect(vacios, 'un comentario sin texto no puede dejar un cuerpo vacío').toBe(0);

    // Y sin nada —ni texto ni archivos— no se envía.
    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Escribe algo o adjunta un archivo antes de mandarlo.')).toBeVisible();
  });

  /** **La imagen queda entre los dos párrafos**, en el sitio donde se puso, y no al final. */
  test('una imagen entre dos párrafos se queda entre los dos', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'La imagen en medio');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    const caja = page.getByLabel('Escribe un comentario');
    await caja.click();

    // Tres renglones con el cursor donde toca: se escribe el primero, se abre el segundo, **ahí se
    // pega la imagen** y se abre el tercero para el texto de abajo.
    await page.keyboard.type('Antes de la imagen.');
    await page.keyboard.press('Enter');
    await pegarUnArchivo(page, 'Escribe un comentario', 'en-medio.png');
    await page.keyboard.press('Enter');
    await page.keyboard.type('Despues de la imagen.');

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    // **El orden se mira en el DOM**, que es lo que hace que la imagen esté en su sitio y no al final:
    // primero el texto de arriba, después el adjunto y después el texto de abajo, uno detrás de otro.
    const orden = await ordenDelCuerpo(cuerpoDe(page));

    expect(orden.orden, 'la imagen va entre los dos párrafos').toEqual([
      'texto:Antes de la imagen.',
      'adjunto:en-medio.png',
      'texto:Despues de la imagen.',
    ]);
  });

  /**
   * Un PDF y un Word: **uno abre el visor y el otro descarga**, que es la diferencia que fija la
   * decisión 46.
   */
  test('el PDF abre el visor y el Word descarga', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Un PDF y un Word');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    await page.getByLabel('Escribe un comentario').fill('Van los dos documentos.');
    await page.locator('input[type=file]').setInputFiles({
      name: 'informe.pdf',
      mimeType: 'application/pdf',
      buffer: INFORME,
    });
    await page.locator('input[type=file]').setInputFiles({
      name: 'acta.docx',
      mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      buffer: Buffer.from('un docx de prueba'),
    });

    // Dentro del texto, los dos se ven como **su enlace**, con su nombre.
    const caja = page.getByLabel('Escribe un comentario');
    await expect(caja.locator('a[data-adjunto="informe.pdf"]')).toHaveText('informe.pdf');
    await expect(caja.locator('a[data-adjunto="acta.docx"]')).toHaveText('acta.docx');

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    const cuerpo = cuerpoDe(page);

    // El PDF: se pulsa y **se abre el visor**, con sus dos botones, y **Descargar lo descarga**.
    await cuerpo.locator('a[data-adjunto="informe.pdf"]').click();

    const visor = page.locator('dialog[open]');
    await expect(visor).toBeVisible();
    await expect(visor.getByRole('button', { name: 'Descargar' })).toBeVisible();
    await expect(visor.getByRole('button', { name: 'Abrir en una pestaña' })).toBeVisible();

    const descargaDelVisor = page.waitForEvent('download');
    await visor.getByRole('button', { name: 'Descargar' }).click();
    expect((await descargaDelVisor).suggestedFilename()).toBe('informe.pdf');

    await visor.getByRole('button', { name: 'Cerrar la vista previa' }).click();
    await expect(page.locator('dialog[open]')).toHaveCount(0);

    // El Word: se pulsa y **descarga**, sin visor de por medio.
    const descarga = page.waitForEvent('download');
    await cuerpo.locator('a[data-adjunto="acta.docx"]').click();
    expect((await descarga).suggestedFilename()).toBe('acta.docx');
    await expect(page.locator('dialog[open]')).toHaveCount(0);

    // Y nada de esto ha dejado direcciones en el texto guardado.
    const detalle = await detalleDe(request, suyo, numero);
    expect(detalle.comments[detalle.comments.length - 1].body).not.toContain('blob:');
  });

  /**
   * **El texto que queda guardado, clavado**: es el contrato con el backend, que sólo admite sus
   * etiquetas y sus atributos y **rechaza el comentario entero** (`422 tickets.body.notAllowed`) si se
   * cuela cualquier otra cosa —una clase, un `style`, una dirección—.
   *
   * Se afirma el HTML exacto a propósito: es el ejemplo que enseña el documento, y cualquier atributo
   * de más que se colara mañana haría fallar esto antes que la pantalla.
   */
  test('lo que se guarda es el texto con sus referencias, y nada más', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'El texto que se guarda');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    await page.getByLabel('Escribe un comentario').fill('Mira lo que me sale, con el informe.');
    await page.locator('input[type=file]').setInputFiles('archivos/captura.png');
    await page.locator('input[type=file]').setInputFiles({
      name: 'informe.pdf',
      mimeType: 'application/pdf',
      buffer: INFORME,
    });

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    const detalle = await detalleDe(request, suyo, numero);

    expect(detalle.comments[detalle.comments.length - 1].body).toBe(
      'Mira lo que me sale, con el informe.' +
        '<img data-adjunto="captura.png">' +
        '<a data-adjunto="informe.pdf">informe.pdf</a>',
    );

    // Y al leer se ve lo mismo: la imagen en línea y el PDF como enlace, los dos dentro del texto.
    const cuerpo = cuerpoDe(page);
    await expect(cuerpo.locator('img[data-adjunto="captura.png"]')).toBeVisible();
    await expect(cuerpo.locator('a[data-adjunto="informe.pdf"]')).toHaveText('informe.pdf');
  });

  /**
   * Un vídeo **se ve como una miniatura y se reproduce en el visor** (decisión 55, del 2026-09-26), y
   * sus cuatro formatos están admitidos (decisión 47).
   *
   * Antes mandaba su reproductor dentro de la conversación, y el visor sólo salía cuando el navegador
   * no sabía reproducirlo. Ahora el clic lo abre siempre: es donde se ve grande y se descarga.
   *
   * Lo que se guarda sigue siendo la referencia **sin `controls`**: ese atributo no está en la lista
   * blanca del backend, así que si viajara, el texto entero se rechazaría.
   */
  test('un vídeo se ve como una miniatura y se reproduce en el visor', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Un vídeo dentro');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    await page.getByLabel('Escribe un comentario').fill('Va el vídeo.');
    await page.locator('input[type=file]').setInputFiles({
      name: 'grabacion.mp4',
      mimeType: 'video/mp4',
      buffer: Buffer.from('un mp4 de prueba'),
    });

    // Al escribir: el reproductor, con sus controles y su archivo de verdad. **Aquí sí**: escribiendo
    // se comprueba lo que se adjunta, y el visor grande no existe en este cuadro.
    const enElEditor = page.getByLabel('Escribe un comentario').locator('video[data-adjunto="grabacion.mp4"]');
    await expect(enElEditor).toBeVisible();
    expect(
      await enElEditor.evaluate((nodo) => ({
        controls: (nodo as HTMLVideoElement).controls,
        archivo: (nodo as HTMLVideoElement).src.startsWith('blob:'),
      })),
    ).toEqual({ controls: true, archivo: true });

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    // Y lo guardado es la referencia, sin `controls` y sin dirección.
    const detalle = await detalleDe(request, suyo, numero);
    const guardado = detalle.comments[detalle.comments.length - 1].body;

    expect(guardado).toContain('<video data-adjunto="grabacion.mp4">');
    expect(guardado, '`controls` no está en la lista blanca del backend').not.toContain('controls');
    expect(guardado).not.toContain('src=');

    // En la conversación: **sin reproductor**, con el tope de tamaño y dentro del botón que lo abre.
    const enLaConversacion = cuerpoDe(page).locator('video[data-adjunto="grabacion.mp4"]');
    await expect(enLaConversacion).toBeVisible();
    const pintado = await enLaConversacion.evaluate((nodo) => {
      const caja = nodo.getBoundingClientRect();
      return {
        controls: (nodo as HTMLVideoElement).controls,
        ancho: caja.width,
        alto: caja.height,
      };
    });

    expect(pintado.controls, 'el reproductor no va en la conversación').toBe(false);
    expect(pintado.ancho, 'el vídeo no puede pasar de 480 px').toBeLessThanOrEqual(480);
    expect(pintado.alto, 'el vídeo no puede pasar de 360 px').toBeLessThanOrEqual(360);

    // Y **el clic abre el visor grande**, con el reproductor y los dos botones (decisiones 48 y 55).
    await page.getByRole('button', { name: 'Ver el vídeo en grande: grabacion.mp4' }).click();

    const visor = page.locator('dialog[open]');
    await expect(visor).toBeVisible();
    await expect(visor.locator('video')).toBeVisible();
    await expect(visor.getByRole('button', { name: 'Descargar' })).toBeVisible();
    await expect(visor.getByRole('button', { name: 'Abrir en una pestaña' })).toBeVisible();
  });

  /**
   * **Una captura grande no se come la conversación** (decisión 55): se ve con el tope de 480 × 360 y
   * entera al pulsarla.
   *
   * La imagen de esta prueba mide **900 × 700**, más que el tope a propósito: con la captura de las
   * otras pruebas —480 × 300, justo el tope— la comprobación pasaría sin que hubiera tope ninguno.
   */
  test('una captura más grande que el tope se ve reducida, y entera en el visor', async ({
    page,
    request,
  }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Una captura enorme');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    await page.getByLabel('Escribe un comentario').fill('Mira el pantallazo.');
    await page.locator('input[type=file]').setInputFiles('archivos/captura-grande.png');
    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    const imagen = cuerpoDe(page).locator('img[data-adjunto="captura-grande.png"]');
    await expect(imagen).toBeVisible();

    const medida = await imagen.evaluate((nodo) => {
      const caja = nodo.getBoundingClientRect();
      return {
        natural: (nodo as HTMLImageElement).naturalWidth,
        ancho: caja.width,
        alto: caja.height,
      };
    });

    // El archivo es más grande de verdad —no se ha recortado ni se ha subido otra cosa—…
    expect(medida.natural).toBe(900);

    // …y lo que se pinta cabe en el tope: 900 × 700 no cabe en 480 × 360, así que se reduce. **Cuál
    // de los dos topes manda depende del ancho de la pantalla** —en PC el alto (360) y en móvil el
    // ancho de la conversación—, así que se comprueban los dos límites y **la proporción**, que es lo
    // que distingue «reducida sin deformarse» de «estrujada».
    expect(medida.ancho).toBeLessThanOrEqual(480);
    expect(medida.alto).toBeLessThanOrEqual(360);
    expect(medida.ancho, 'tiene que encogerse: el archivo mide 900').toBeLessThan(900);
    expect(medida.ancho / medida.alto).toBeCloseTo(900 / 700, 1);

    // **Y no desborda**: el tope es «lo que quepa, y como mucho 480», no «480 y que la página se
    // ensanche». Con `max-w-[480px]` a secas la imagen de 480 se salía de un móvil de 412 px y el
    // documento entero se iba a 530: lo encontró esta capa, porque el clic en un botón de la misma
    // tarjeta empezó a fallar. Aquí se comprueba en los dos anchos.
    const desborde = await page.evaluate(() => ({
      documento: document.documentElement.scrollWidth,
      ventana: window.innerWidth,
    }));

    expect(desborde.documento, 'la conversación no puede tener scroll horizontal').toBeLessThanOrEqual(
      desborde.ventana,
    );

    // Y al pulsarla, el visor la enseña entera, con sus dos botones.
    await imagen.click();

    const visor = page.locator('dialog[open]');
    await expect(visor).toBeVisible();
    await expect(visor.locator('img')).toBeVisible();
    await expect(visor.getByRole('button', { name: 'Descargar' })).toBeVisible();
  });

  /**
   * **El texto y el código también se adjuntan** (decisión 54, del 2026-09-26): un `.sql`, un `.json` o
   * un `.log` es lo que se manda cuando el caso es una consulta que falla o un despliegue que no
   * arranca, y hasta ese día la lista de extensiones no los tenía —`css`, `md` y `log` sí, `sql` no—.
   *
   * Se comprueba lo que importa: que **se cogen** (el editor no los rechaza), que **se guardan** como
   * cualquier adjunto, que al pulsarlos **se descargan** —nunca se pintan en línea— y que la lista
   * sigue siendo cerrada: el `svg` se rechaza igual que antes.
   */
  test('un .sql se adjunta y se descarga, y un .svg se sigue rechazando', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Con la consulta dentro');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    const caja = page.getByLabel('Escribe un comentario');
    await caja.fill('Va la consulta que falla.');

    await page.locator('input[type=file]').setInputFiles({
      name: 'consulta.sql',
      mimeType: 'application/sql',
      buffer: Buffer.from('select * from tickets where 1 = 1;'),
    });

    // Se coge: aparece en el texto como enlace con su nombre, y **sin el aviso de rechazo**.
    await expect(caja.locator('a[data-adjunto="consulta.sql"]')).toHaveText('consulta.sql');
    await expect(page.getByText('Estos archivos no se han añadido')).toHaveCount(0);

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    // Y se sube y se guarda: está entre los adjuntos del comentario.
    const detalle = await detalleDe(request, suyo, numero);
    expect(detalle.comments[detalle.comments.length - 1].body).toContain('data-adjunto="consulta.sql"');
    expect(detalle.attachments.map((adjunto) => adjunto.filename)).toContain('consulta.sql');

    // Al pulsarlo **se descarga**, sin visor: no es imagen, ni vídeo, ni PDF.
    const descarga = page.waitForEvent('download');
    await cuerpoDe(page).locator('a[data-adjunto="consulta.sql"]').click();
    expect((await descarga).suggestedFilename()).toBe('consulta.sql');
    await expect(page.locator('dialog[open]')).toHaveCount(0);

    // **Y la lista sigue cerrada**: el `svg` no entra, y se dice antes de subirlo.
    await caja.fill('Y ahora un dibujo.');
    await page.locator('input[type=file]').setInputFiles({
      name: 'dibujo.svg',
      mimeType: 'image/svg+xml',
      buffer: Buffer.from('<svg xmlns="http://www.w3.org/2000/svg"></svg>'),
    });

    await expect(page.getByText('Estos archivos no se han añadido')).toBeVisible();
    await expect(page.getByText('dibujo.svg')).toBeVisible();
  });

  /**
   * **Un adjunto que el texto no nombra sigue en su lista, al final** (decisión 49): es lo que hace que
   * los tickets y comentarios que ya existen —cuyo texto no referencia nada— se vean igual que antes.
   */
  test('un adjunto que el texto no nombra sigue en la lista del final', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Uno fuera del texto');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    await page.getByLabel('Escribe un comentario').fill('Con la captura dentro.');
    await page.locator('input[type=file]').setInputFiles('archivos/captura.png');
    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    // Y otro archivo que el texto no nombra, colgado del mismo comentario: es lo que pasa con todo lo
    // que se subió antes de este cambio.
    const detalle = await detalleDe(request, suyo, numero);
    const comentario = detalle.comments[detalle.comments.length - 1];

    const subida = await request.post(`/api/tickets/${numero}/attachments`, {
      headers: { Authorization: `Bearer ${suyo}` },
      multipart: {
        commentId: String(comentario.id),
        file: {
          name: 'fuera-del-texto.txt',
          mimeType: 'text/plain',
          buffer: Buffer.from('un archivo que el texto no nombra'),
        },
      },
    });
    expect(subida.status(), await subida.text()).toBe(201);

    await page.reload();

    // El que el texto nombra está **dentro**, con su imagen…
    const cuerpo = cuerpoDe(page);
    await expect(cuerpo.locator('img[data-adjunto="captura.png"]')).toBeVisible();

    // …y el que no nombra, **en su lista al final**, con su nombre y sus botones.
    const enLaLista = page.locator('app-adjunto').filter({ hasText: 'fuera-del-texto.txt' });
    await expect(enLaLista).toBeVisible();
    await expect(enLaLista.getByRole('button', { name: 'Descargar' })).toBeVisible();

    // La imagen de dentro no se repite en la lista: lo que el texto ya enseña no se enseña dos veces.
    await expect(page.locator('app-adjunto').filter({ hasText: 'captura.png' })).toHaveCount(0);
  });

  /**
   * **Al editar un comentario con una imagen, la vista previa se ve también en el editor**: es lo que
   * pidió el responsable, y lo que hace que el texto no sea un montón de etiquetas.
   */
  test('al editar un comentario, la imagen se ve en el editor', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Editar con imagen');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    await page.getByLabel('Escribe un comentario').fill('Con la captura.');
    await page.locator('input[type=file]').setInputFiles('archivos/captura.png');
    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    // **Exacto**: desde el 2026-09-28 la ficha tiene «Editar el asunto» y «Editar la descripción», y
    // sin `exact` este `first()` cogía el lápiz de la descripción en vez del botón del comentario.
    await page.getByRole('button', { name: 'Editar', exact: true }).first().click();

    // **El editor del comentario que se está editando**, no otro: desde el 2026-09-28 la ficha tiene
    // **su propio editor** para la descripción (se edita donde se muestra), así que en la página hay
    // dos, y el suyo lleva el identificador del comentario.
    // El identificador va **en el propio lienzo** (que es el `textbox`), no en un padre.
    const editor = page.locator('[id^="editar-comentario-"][role="textbox"]');
    const imagen = editor.locator('img[data-adjunto="captura.png"]');
    await expect(imagen).toBeVisible();
    expect(
      (await anchoDeLaImagen(imagen)).natural,
      'la vista previa se ve también al editar',
    ).toBeGreaterThan(0);

    // Y guardar sin tocar nada deja el texto como estaba, con su referencia y sin direcciones.
    // **El primero de los dos**: desde el 2026-09-27 la ficha tiene su propio «Guardar» para la categoría
    // y las etiquetas, y el de la conversación va antes en la pantalla —la línea de tiempo está a la
    // izquierda y la ficha a la derecha—, así que «el» botón de guardar ya no es uno solo.
    await page.getByRole('button', { name: 'Guardar' }).first().click();
    await expect(page.getByText('Comentario guardado.')).toBeVisible();

    const detalle = await detalleDe(request, suyo, numero);
    const guardado = detalle.comments[detalle.comments.length - 1].body;

    expect(guardado).toContain('data-adjunto="captura.png"');
    expect(guardado).not.toContain('blob:');
  });

  /**
   * Pegar texto **no trae su formato**: el HTML de un correo o de Word traería etiquetas que no están
   * en la lista blanca, y el comentario se rechazaría entero por algo que la persona no ha escrito
   * (docs/modules/tickets.md, sección 2.3 y `docs/interfaz-y-experiencia.md`, sección 6.5).
   */
  test('pegar texto con formato entra como texto plano', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Pegar sin formato');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    const caja = page.getByLabel('Escribe un comentario');
    await caja.click();

    await caja.evaluate((nodo) => {
      const transferencia = new DataTransfer();
      transferencia.setData('text/html', '<b>Hola</b> <span style="color: red">mundo</span>');
      transferencia.setData('text/plain', 'Hola mundo\nsegunda linea');

      nodo.dispatchEvent(
        new ClipboardEvent('paste', { clipboardData: transferencia, bubbles: true, cancelable: true }),
      );
    });

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    const detalle = await detalleDe(request, suyo, numero);
    const guardado = detalle.comments[detalle.comments.length - 1].body;

    expect(guardado).toContain('Hola mundo');
    expect(guardado, 'el formato del original no entra').not.toContain('<b>');
    expect(guardado).not.toContain('style=');

    // Los saltos de línea se respetan: es lo que se ve al pegar.
    expect(guardado).toContain('<br>');
    await expect(page.getByText('segunda linea')).toBeVisible();
  });

  /**
   * **Si un archivo no sube después de publicar el comentario** (decisión 44): el comentario **se
   * queda** —está escrito y lo ha visto quien lo escriba—, se avisa de **qué archivo** falta y el
   * botón lo sube **al mismo comentario**, sin volver a escribirlo. Y su hueco se ve **en el texto**,
   * marcado, que es lo que añadió la decisión 49.
   *
   * El fallo se provoca **cortando la subida en el navegador**, que es lo único que hace que esto se
   * pueda probar de verdad: con el servidor sano, una subida no falla nunca sola.
   */
  test('si un archivo no sube, se avisa y se puede reintentar sin repetir el comentario', async ({
    page,
    request,
  }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);
    const { numero } = await ticketDePrueba(request, suyo, 'Una subida que falla');

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    // La primera subida se corta; a partir de ahí, todo pasa.
    let cortadas = 0;
    await page.route('**/api/tickets/*/attachments', async (ruta) => {
      if (ruta.request().method() === 'POST' && cortadas === 0) {
        cortadas++;
        await ruta.abort('failed');
        return;
      }

      await ruta.continue();
    });

    await page.getByLabel('Escribe un comentario').fill('Va con captura.');
    await page.locator('input[type=file]').setInputFiles({
      name: 'no-subio.png',
      mimeType: 'image/png',
      buffer: CAPTURA,
    });
    await page.getByRole('button', { name: 'Comentar' }).click();

    // El aviso dice cuál falta, y el comentario está publicado **una sola vez**.
    await expect(
      page.getByText('El comentario se ha publicado, pero no se han podido subir estos archivos:'),
    ).toBeVisible();
    await expect(page.getByText('no-subio.png').first()).toBeVisible();
    await expect(page.getByText('Va con captura.')).toHaveCount(1);
    await expect(page.getByLabel('Escribe un comentario')).toHaveText('');

    // **Y el hueco del que falta se ve en el texto**, marcado: se nombra el archivo y no subió.
    const marca = cuerpoDe(page).locator('[data-adjunto-que-falta="no-subio.png"]');
    await expect(marca).toBeVisible();
    await expect(marca).toContainText('no-subio.png');

    // El botón lo sube al mismo comentario: el aviso se va, el hueco se convierte en la imagen y el
    // comentario sigue estando una sola vez.
    await page.getByRole('button', { name: 'Reintentar la subida' }).click();

    await expect(page.getByText('Ya están subidos todos los archivos del comentario.')).toBeVisible();
    await expect(
      page.getByText('El comentario se ha publicado, pero no se han podido subir estos archivos:'),
    ).toHaveCount(0);
    await expect(cuerpoDe(page).locator('[data-adjunto-que-falta]')).toHaveCount(0);

    const enElTexto = cuerpoDe(page).locator('img[data-adjunto="no-subio.png"]');
    await expect(enElTexto).toBeVisible();
    expect((await anchoDeLaImagen(enElTexto)).natural).toBeGreaterThan(0);
    await expect(page.getByText('Va con captura.')).toHaveCount(1);
  });

  /** El alta también entiende el pegado: es donde más capturas se adjuntan (sección 4.1). */
  test('el alta acepta una captura pegada', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);

    await entrarComo(page, ana.email);
    await page.goto('/tickets/new');

    const marca = Date.now().toString(36);
    await page.getByLabel('Asunto', { exact: true }).fill(`Pegada ${marca}`);
    await page.getByLabel('Descripción').fill('Va la captura pegada.');

    // La categoría es obligatoria: sin ella, el alta no crea el ticket.
    await page.getByLabel('Categoría').selectOption({ label: 'General' });

    await pegarUnArchivo(page, 'Descripción', 'captura-del-alta.png');
    await expect(
      page.getByLabel('Descripción').locator('img[data-adjunto="captura-del-alta.png"]'),
    ).toBeVisible();

    await page.getByRole('button', { name: 'Crear el ticket' }).click();
    await expect(page).toHaveURL(/\/tickets\/[A-Z0-9-]+$/);

    // El alta tiene un paso más —crear, subir y volver a guardar la descripción—, y **no se nota**: la
    // imagen está dentro de la descripción, con su tamaño.
    const imagen = cuerpoDe(page, 0).locator('img[data-adjunto="captura-del-alta.png"]');
    await expect(imagen).toBeVisible();
    expect((await anchoDeLaImagen(imagen)).natural).toBeGreaterThan(0);
  });

  /**
   * **La búsqueda de la bandeja no encuentra por una etiqueta**: la descripción de un ticket es HTML, y
   * buscar «img» sobre el HTML crudo encontraría todos los que llevan una imagen
   * (`docs/modules/tickets.md`, sección 2.3). El backend quita las etiquetas antes de comparar, y esto
   * lo comprueba **por la API**, que es donde está la consulta.
   */
  test('la bandeja no encuentra un ticket por una etiqueta de su texto', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const suyo = await tokenDe(request, ana.email);

    // La descripción lleva una imagen, y su texto no dice la palabra «img» por ninguna parte.
    const { numero, asunto } = await ticketDePrueba(
      request,
      suyo,
      'Con una imagen dentro',
      '<p>Mira lo que me sale:</p><img data-adjunto="captura.png">',
    );

    const busqueda = await request.get('/api/tickets', {
      headers: { Authorization: `Bearer ${suyo}` },
      params: { q: 'img' },
    });

    expect(busqueda.status(), await busqueda.text()).toBe(200);
    const pagina = (await busqueda.json()) as { tickets: { number: string }[] };

    expect(
      pagina.tickets.some((ticket) => ticket.number === numero),
      'buscar «img» no puede encontrar un ticket por la etiqueta de su imagen',
    ).toBe(false);

    // Y buscar su texto de verdad sí lo encuentra: la prueba no está diciendo que la búsqueda no
    // funcione, sino que **no mira dentro de las etiquetas**.
    const porTexto = await request.get('/api/tickets', {
      headers: { Authorization: `Bearer ${suyo}` },
      params: { q: 'Mira lo que me sale' },
    });
    const encontrados = ((await porTexto.json()) as { tickets: { number: string }[] }).tickets;

    expect(encontrados.some((ticket) => ticket.number === numero)).toBe(true);
    expect(asunto).toContain('Con una imagen dentro');

    // Y por la pantalla, lo mismo: la bandeja no lo enseña al buscar «img».
    await entrarComo(page, ana.email);
    await page.goto('/tickets');
    await buscarEnLaLista(page, 'img');
    await expect(page.getByRole('link', { name: numero })).toHaveCount(0);
  });
});

/**
 * **Los dos campos que redacta el motor de IA**: «Motivo» y «Última acción», en la lista y en la ficha
 * (`docs/modules/ai.md`).
 *
 * El motor tarda de verdad —medido, entre 12 y 24 segundos por campo—, así que la prueba de que los
 * textos llegan **espera**, y si el motor no está configurado lo dice y se salta: la aplicación tiene
 * que funcionar sin él, y eso también se comprueba aquí.
 */
test.describe('Los resúmenes del motor de IA', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se pueden crear cuentas');
    void page;
  });

  test('la lista lleva sus columnas, con los dos nombres nuevos', async ({ page, request }, info) => {
    test.skip(info.project.name !== 'pc', 'La tabla es de PC: en móvil se ven tarjetas');
    const ana = await crearCuentaLista(request);
    await ticketDePrueba(request, await tokenDe(request, ana.email), 'Para mirar las columnas');

    await entrarComo(page, ana.email);
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();

    // Los nombres que pidió el responsable el 2026-09-27: «Número» es «Ticket» y «Última
    // actualización» es «Fecha actualización», y en medio están los dos campos del motor. **Y
    // «Clasificación» es del 2026-09-29** (decisión 86): la categoría y las etiquetas, que antes iban
    // como chips debajo del número.
    const cabeceras = page.locator('table thead th');
    await expect(cabeceras).toHaveText([
      'Ticket',
      'Asunto',
      'Motivo',
      'Última acción',
      'Clasificación',
      'Estado',
      'Solicitante',
      'Responsable',
      'Fecha actualización',
      'Abrir',
    ]);

    // **Y la columna del número va sola** (decisión 86): sin los chips de la clasificación, que ahora
    // tienen la suya. Se comprueba en una fila de verdad, con un ticket clasificado.
    const renglon = page.locator('table tbody tr').first();
    await expect(renglon.locator('td').first()).toHaveText(/^\s*[A-Z]+-\d{4}-\d+\s*$/);
  });

  test('el motivo y la última acción se redactan, y se pueden volver a pedir', async ({
    page,
    request,
  }, info) => {
    // **Sólo en PC**: es el único caso que espera al motor de verdad, y el motor es **uno** para toda
    // la instalación y atiende de uno en uno. Repetirlo en los dos anchos multiplicaba la cola y hacía
    // fallar la espera por saturación, no por un fallo.
    test.skip(info.project.name !== 'pc', 'Lo que se prueba aquí no depende del ancho');

    // **Cinco minutos**: el motor tarda entre 12 y 24 segundos por campo y aquí se esperan los dos, con
    // la cola compartida. El tope de por defecto son 30 s, que no llegan ni para el primero.
    test.setTimeout(330_000);
    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const ana = await crearCuentaLista(request);
    const { numero } = await ticketDePrueba(
      request,
      await tokenDe(request, ana.email),
      'No puedo entrar a la aplicación',
      '<p>Desde esta mañana me sale usuario o contraseña incorrectos. He reiniciado el equipo y sigue igual.</p>',
    );

    // ¿Está el motor puesto en esta instalación? Lo dice el estado de los dos campos.
    const suyo = await tokenDe(request, ana.email);
    const primero = await detalleDe(request, suyo, numero);
    const estado = primero.ticket.insights?.motivo.state ?? '';

    test.skip(
      estado === 'sin_motor',
      'El motor de IA no está configurado en esta instalación: los campos quedan sin texto, que es lo que se quiere cuando no hay motor',
    );

    // **El enganche es inmediato**: nada más crear el ticket, los dos campos están «pendiente». Eso no
    // depende del motor y es lo que prueba que el alta no espera a nadie (decisión 2).
    await expect
      .poll(
        async () => {
          const detalle = await detalleDe(request, suyo, numero);
          const motivo = detalle.ticket.insights?.motivo.state ?? '';
          const ultima = detalle.ticket.insights?.ultimaAccion.state ?? '';

          return [motivo, ultima].includes('pendiente') || [motivo, ultima].every((e) => e !== '');
        },
        { timeout: 30_000, message: 'los dos campos deberían quedar pendientes al crear el ticket' },
      )
      .toBe(true);

    // Y después, el motor tarda: se espera a que lleguen a un estado terminal (hasta cinco minutos).
    //
    // **Se admite `error`**: es un modelo de 1,5B, y una de cada tantas respuestas llega a medias —el
    // módulo lo reintenta tres veces y después lo dice—. Si eso pasa, la prueba comprueba lo que sí
    // tiene que cumplirse siempre: que la pantalla lo cuenta en castellano y que no se inventa nada.
    //
    // **Y se admite que siga esperando, saltándose el resto**: el motor es **uno** y atiende de uno en
    // uno, y una pasada completa de esta capa crea decenas de tickets que encolan dos campos cada uno
    // —al terminar esta pasada había **143 campos en cola**—, así que los dos de este ticket pueden
    // tardar horas. Lo que la prueba garantiza siempre es el enganche —los dos campos quedan
    // `pendiente` al crear el ticket, sin que el alta espere a nadie—, y los textos se comprueban
    // cuando el motor ha llegado; cuando no, se dice y se salta, en vez de fallar por una cola.
    const llegoElMotor = await expect
      .poll(
        async () => {
          const detalle = await detalleDe(request, suyo, numero);
          const motivo = detalle.ticket.insights?.motivo.state ?? '';
          const ultima = detalle.ticket.insights?.ultimaAccion.state ?? '';

          return [motivo, ultima].every((estado) => estado === 'listo' || estado === 'error');
        },
        { timeout: 300_000, message: 'el motor no terminó de redactar los dos campos' },
      )
      .toBe(true)
      .then(
        () => true,
        () => false,
      );

    // **La pantalla se comprueba antes de esperar al motor**, porque lo que no depende de él tiene que
    // quedar comprobado siempre: los dos campos se enseñan con su aviso de que se están escribiendo, el
    // botón de volver a pedirlos está y funciona, y la lista enseña lo mismo.
    await entrarComo(page, soporte.email);
    await page.goto('/tickets/main');
    await buscarEnLaLista(page, numero);

    const fila = page.getByRole('row').filter({ hasText: numero });

    // **El aviso de que se está escribiendo, sólo si de verdad lo está**: el motor puede haber
    // contestado ya —es rápido cuando no tiene cola— y entonces la fila enseña el texto. Lo que se
    // comprueba siempre es que el alta **no espera** a nadie: los dos campos quedan `pendiente` (eso se
    // mira antes, por la API).
    const alLlegar = await detalleDe(request, suyo, numero);

    if (alLlegar.ticket.insights?.motivo.state === 'pendiente') {
      await expect(fila).toContainText('Generando…');
    } else {
      await expect(fila).toContainText(alLlegar.ticket.insights?.motivo.text ?? '');
    }

    await page.getByRole('link', { name: numero }).click();
    await expect(page.getByRole('heading', { name: 'De qué va el ticket' })).toBeVisible();

    const regenerar = page.getByRole('button', { name: 'Volver a resumir' });
    await expect(regenerar).toBeVisible();
    await regenerar.click();
    await expect(page.getByText('Se han pedido otra vez los dos resúmenes.')).toBeVisible();

    // Y si el motor ha llegado, sus dos textos: los mismos que dice la API, en el idioma de quien mira,
    // en la lista y en la ficha.
    if (!llegoElMotor) {
      test.skip(true, 'El motor todavía tenía cola: los textos ya se comprueban cuando llega');
    }

    const listo = await detalleDe(request, suyo, numero);

    // **Los textos se comprueban sólo cuando están** (docs/modules/ai.md, decisiones 2 y 8): un campo en
    // `error` no tiene texto, y uno `pendiente` tampoco —volver a pedir los resúmenes los deja así hasta
    // que el motor vuelve a contestar—. Dar por hecho que están hacía fallar esta prueba por su culpa y
    // no por el producto, y pasó dos veces: con el modelo fallando un campo, y justo después de pulsar
    // «Volver a resumir», que es lo que la prueba hace más arriba.
    const estados = [listo.ticket.insights?.motivo.state, listo.ticket.insights?.ultimaAccion.state];
    const fallo = estados.includes('error');

    if (estados.includes('pendiente')) {
      test.skip(true, 'El motor todavía no había contestado a los resúmenes pedidos otra vez');
    }

    if (!fallo) {
      expect(listo.ticket.insights?.motivo.text?.length ?? 0).toBeGreaterThan(10);
      expect(listo.ticket.insights?.ultimaAccion.text?.length ?? 0).toBeGreaterThan(10);
      expect(listo.ticket.insights?.motivo.language).toBe('es');
      expect(listo.ticket.insights?.ultimaAccion.language).toBe('es');
    }

    await page.goto('/tickets/main');
    await buscarEnLaLista(page, numero);
    await expect(fila).toContainText(listo.ticket.insights?.motivo.text ?? '');
    await expect(fila).toContainText(listo.ticket.insights?.ultimaAccion.text ?? '');

    await page.getByRole('link', { name: numero }).click();
    const textoMotivo = listo.ticket.insights?.motivo.text ?? '';
    const textoUltima = listo.ticket.insights?.ultimaAccion.text ?? '';
    if (textoMotivo === textoUltima) {
      // El doble de pruebas responde el mismo texto artificial para ambos encargos. La ficha debe
      // mostrarlo una vez en cada campo; un locator estricto sin acotar sería ambiguo.
      await expect(page.getByText(textoMotivo, { exact: true })).toHaveCount(2);
    } else {
      await expect(page.getByText(textoMotivo, { exact: true })).toBeVisible();
      await expect(page.getByText(textoUltima, { exact: true })).toBeVisible();
    }

    // Si alguno falló, la pantalla lo dice en castellano y no enseña ninguna clave (sección 8 de
    // `docs/interfaz-y-experiencia.md`).
    if (fallo) {
      await expect(page.getByText('No se pudo resumir').first()).toBeVisible();
    }
  });

  test('el «Motivo» de un interno es el del escalado, y su última acción la redacta el motor', async ({
    page,
    request,
  }) => {
    test.setTimeout(200_000);
    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const ana = await crearCuentaLista(request);
    const { numero } = await ticketDePrueba(
      request,
      await tokenDe(request, ana.email),
      'Un caso para escalar',
    );

    // Se escala desde la pantalla, que es como lo hace Soporte.
    await entrarComo(page, soporte.email);
    await page.goto(`/tickets/${numero}`);
    await elegirEstado(page, numero, 'escalado');
    await page.getByLabel('Motivo').fill('Es un fallo del servidor de sesiones, no de la cuenta.');
    await page.getByRole('button', { name: 'Confirmar', exact: true }).click();
    await expect(page.getByText('Escalado').first()).toBeVisible();

    const interno = `INT-${numero}`;

    // En la lista de internos, el motivo es **el del escalado** —no se le pide al motor— y la última
    // acción sí la redacta el motor (decisión 5).
    await page.goto('/tickets/internal');
    await buscarEnLaLista(page, interno);

    // Se mira el renglón de ese interno, que en PC es una fila y en móvil una tarjeta.
    const renglon = page.locator('tr, li').filter({ hasText: interno }).last();
    await expect(renglon).toContainText('Es un fallo del servidor de sesiones, no de la cuenta.');

    // Y **no se le pide el motivo al motor**: al volver a pedir los resúmenes de un interno —que es lo
    // que hace el botón, que en la vista doble aparece **una vez por cada ficha**—, sólo se encola la
    // última acción. El motivo sigue siendo el del escalado, así que su estado llega vacío (decisión 5).
    //
    // Se pide por la API y no pulsando el botón: en la pantalla del interno hay **dos fichas** —la del
    // principal y la suya—, y las dos traen su botón, así que pulsar «el» botón es ambiguo por
    // construcción. El botón ya se prueba en el caso del principal, donde hay uno solo.
    const suyo = await tokenDe(request, soporte.email);
    const pedido = await request.post(`/api/tickets/${interno}/insights`, {
      headers: { Authorization: `Bearer ${suyo}` },
      data: {},
    });
    expect(pedido.status(), await pedido.text()).toBe(200);

    const trasPedir = await detalleDe(request, suyo, interno);

    expect(
      trasPedir.ticket.insights?.motivo.state ?? '',
      'a un interno no se le pide el motivo: es el del escalado',
    ).toBe('');
    expect(trasPedir.ticket.insights?.ultimaAccion.state ?? '').not.toBe('');
  });
});

/**
 * **Etiquetar a alguien y los observadores** (`docs/modules/tickets.md`, decisiones 58 a 63).
 *
 * Lo que se prueba aquí es lo que no se ve en el backend: que el nombre entra **dentro del comentario y
 * resaltado**, que la ficha lista a los observadores —que **no son el responsable**—, que quien observa
 * ve el ticket en su chip «Observo», que **el usuario no tiene el botón** y que quitar a alguien deja su
 * rastro en la línea de tiempo.
 */
test.describe('Las menciones y los observadores', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se pueden crear cuentas');
    void page;
  });

  test('etiquetar a un desarrollador lo hace observador, y se le puede quitar', async ({
    page,
    request,
  }) => {
    // **Un minuto**: el caso da de alta tres cuentas leyendo sus correos, escribe un comentario con una
    // mención, entra con otra persona y vuelve, y con el modal de búsqueda por medio se quedaba sin los
    // treinta segundos de siempre.
    test.setTimeout(60_000);

    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    // **Con nombre propio**: el buscador de a quién etiquetar enseña nombres y todas las cuentas de
    // prueba se llaman igual, así que una sin nombre propio no se puede ni elegir ni comprobar.
    // El apellido lleva la marca del momento: si no, cada pasada de esta prueba deja otra persona con
    // el mismo nombre y el botón deja de ser único.
    const desarrollo = await crearCuentaLista(request, {
      role: 'desarrollo',
      name: 'Etiquetada',
      lastName: `Desarrollo ${Date.now().toString(36)}`,
    });
    const ana = await crearCuentaLista(request);
    const { numero } = await ticketDePrueba(
      request,
      await tokenDe(request, ana.email),
      'Para etiquetar a un compañero',
    );

    await entrarComo(page, soporte.email);
    await page.goto(`/tickets/${numero}`);

    // **El botón de etiquetar está** para Soporte, y abre la lista de personas.
    await page.getByRole('button', { name: 'Etiquetar' }).click();
    const lista = page.getByRole('group', { name: 'A quién llamas' });
    await expect(lista).toBeVisible();
    await lista
      .getByRole('button', { name: `${desarrollo.name} ${desarrollo.lastName}`, exact: true })
      .click();

    // El nombre queda **dentro del comentario**, como una mención de verdad.
    const mencion = page.locator('[data-mencion]').last();
    await expect(mencion).toHaveText(`${desarrollo.name} ${desarrollo.lastName}`);

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    // Y en la ficha aparece **en «Observadores»**, que es una línea aparte de la del responsable: el
    // ticket lo lleva otra persona —el reparto lo asigna al crearlo— y el observador no lo sustituye.
    await expect(page.getByText('Observadores:')).toBeVisible();
    await expect(
      page.getByText(`${desarrollo.name} ${desarrollo.lastName}`, { exact: true }).last(),
    ).toBeVisible();
    await expect(page.getByText('Responsable:')).toBeVisible();

    // Quien observa lo tiene en **su** «Observo», no entre los asignados. El chip de vista vive ahora
    // **dentro del modal de búsqueda** (es uno de los criterios), así que se elige ahí.
    await entrarComo(page, desarrollo.email);
    await page.goto('/tickets');
    await filtrarEnLaLista(page, 'Vista', 'Observo');
    await expect(page.getByRole('link', { name: numero })).toBeVisible();

    await filtrarEnLaLista(page, 'Vista', 'Asignados');
    await expect(page.getByRole('link', { name: numero })).toHaveCount(0);

    // Y quitarlo lo puede hacer **cualquier** técnico o desarrollador: aquí lo quita Soporte, que no
    // fue quien lo etiquetó... y lo fue. Se quita desde la ficha, y **queda en el historial**.
    await entrarComo(page, soporte.email);
    await page.goto(`/tickets/${numero}`);
    await page
      .getByRole('button', { name: `Quitar de los observadores: ${desarrollo.name} ${desarrollo.lastName}` })
      .click();
    await expect(page.getByText('ya no sigue el ticket.')).toBeVisible();

    await expect(page.getByText('Nadie más sigue este ticket.')).toBeVisible();
    await expect(page.getByText('de los observadores')).toBeVisible();

    // **Y no vuelve** al volver a guardar el mismo comentario: quitar manda sobre lo escrito.
    const suyo = await tokenDe(request, soporte.email);
    const detalles = await request.get(`/api/tickets/${numero}`, {
      headers: { Authorization: `Bearer ${suyo}` },
    });
    const antes = (await detalles.json()) as { comments: { id: number; body: string }[] };
    const conMencion = antes.comments.find((comentario) => comentario.body.includes('data-mencion'));

    if (conMencion) {
      await request.post(`/api/tickets/${numero}/comments/${conMencion.id}`, {
        headers: { Authorization: `Bearer ${suyo}` },
        data: { body: conMencion.body },
      });
    }

    await page.reload();
    await expect(page.getByText('Nadie más sigue este ticket.')).toBeVisible();
  });

  test('el usuario no tiene el botón de etiquetar', async ({ page, request }) => {
    const ana = await crearCuentaLista(request);
    const { numero } = await ticketDePrueba(
      request,
      await tokenDe(request, ana.email),
      'Un ticket de usuario, sin etiquetas',
    );

    await entrarComo(page, ana.email);
    await page.goto(`/tickets/${numero}`);

    await expect(page.getByLabel('Escribe un comentario')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Etiquetar' })).toHaveCount(0);
    // Y la ficha le dice que nadie más lo sigue, sin ofrecerle nada.
    await expect(page.getByText('Nadie más sigue este ticket.')).toBeVisible();
  });
});

/**
 * **Las categorías y las etiquetas** (`docs/modules/tickets.md`, decisiones 64 a 68).
 *
 * Lo que se prueba aquí es lo que no se ve en el backend: que **el alta no deja crear sin categoría**,
 * que las etiquetas **se normalizan mientras se escriben** —minúsculas, espacios a guiones, sin
 * acentos—, que el listado las enseña como fichas, que se **filtra y se busca** por las dos, y que el
 * catálogo lo mantiene quien toca: **Soporte crea y renombra, y retirar es sólo del Administrador**.
 */
test.describe('Las categorías y las etiquetas', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se pueden crear cuentas');
    void page;
  });

  test('«Categorías y etiquetas»: lo mantiene el Administrador, y Soporte lo ve sin poder tocarlo', async ({
    page,
    request,
  }) => {
    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const marca = Date.now().toString(36);
    const nombre = `Categoria ${marca}`;
    const renombrada = `Renombrada ${marca}`;

    // **Quien mantiene el catálogo es el Administrador** (decisión 83), así que aquí se entra con la
    // cuenta de fábrica. Soporte entra después, a comprobar que ve el catálogo y no lo puede tocar.
    await entrarComo(page, FABRICA.email, FABRICA.password);
    await abrirMenuSiEsCajon(page);
    await page.getByRole('link', { name: 'Categorías y etiquetas' }).click();
    await expect(page).toHaveURL(/\/tickets\/categories$/);
    await expect(page.getByRole('heading', { name: 'Categorías y etiquetas' })).toBeVisible();

    // **Soporte crea** una categoría.
    // **El Administrador crea** la categoría.
    await page.getByRole('button', { name: 'Nueva categoría' }).click();
    await page.locator('#nueva-categoria').fill(nombre);
    await page.getByRole('button', { name: 'Crear', exact: true }).click();
    await expect(page.getByText('Categoría creada.')).toBeVisible();

    // El catálogo es una lista: cada categoría es un renglón con su nombre, su cuenta y sus botones.
    const fila = page.locator('li').filter({ hasText: nombre }).last();
    await expect(fila).toBeVisible();

    // **Y la renombra**, en la fila.
    await fila.getByRole('button', { name: 'Renombrar' }).click();
    await page.getByRole('textbox', { name: 'Nombre' }).last().fill(renombrada);
    await page.getByRole('button', { name: 'Guardar', exact: true }).click();
    await expect(page.getByText('Categoría renombrada.')).toBeVisible();
    await expect(page.locator('li').filter({ hasText: renombrada })).toBeVisible();

    // **Y la retira**: deja de ofrecerse al crear, y los tickets que la tienen la conservan.
    const suya = page.locator('li').filter({ hasText: renombrada }).last();
    await suya.getByRole('button', { name: 'Retirar' }).click();
    await expect(page.getByText('Categoría retirada. Los tickets que la tienen la conservan.')).toBeVisible();
    await expect(page.locator('li').filter({ hasText: renombrada })).toContainText('Retirada');

    // Se deja como estaba: vuelta a poner, que el catálogo no se queda sin ella por una prueba.
    await page
      .locator('li')
      .filter({ hasText: renombrada })
      .last()
      .getByRole('button', { name: 'Volver a poner' })
      .click();
    await expect(page.getByText('Categoría puesta otra vez.')).toBeVisible();

    // **Y Soporte entra a ver el catálogo**: lo ve —la pantalla es suya para consultarlo— pero **sin
    // ningún botón de mantenerlo** (decisión 83): ni crear, ni renombrar, ni retirar.
    await salir(page);
    await entrarComo(page, soporte.email);
    await page.goto('/tickets/categories');

    await expect(page.getByRole('heading', { name: 'Categorías y etiquetas' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Nueva categoría' })).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Nueva etiqueta' })).toHaveCount(0);

    const laDeSoporte = page.locator('li').filter({ hasText: renombrada }).last();
    await expect(laDeSoporte).toBeVisible();
    await expect(laDeSoporte.getByRole('button', { name: 'Renombrar' })).toHaveCount(0);
    await expect(laDeSoporte.getByRole('button', { name: 'Retirar' })).toHaveCount(0);
    await expect(laDeSoporte.getByRole('button', { name: 'Volver a poner' })).toHaveCount(0);
  });

  test('Desarrollo etiqueta desde el interno, y las etiquetas quedan en el principal', async ({
    page,
    request,
  }) => {
    test.setTimeout(60_000);

    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const desarrollo = await crearCuentaLista(request, { role: 'desarrollo' });
    const suyo = await tokenDe(request, soporte.email);
    const suyoDeDesarrollo = await tokenDe(request, desarrollo.email);

    const { numero } = await ticketDePrueba(request, suyo, 'Para etiquetar desde el interno', '<p>Prueba.</p>');

    // Se escala para tener el interno, que es donde trabaja Desarrollo.
    await entrarComo(page, soporte.email);
    await page.goto(`/tickets/${numero}`);
    await elegirEstado(page, numero, 'escalado');
    await page.getByLabel('Motivo').fill('Es un fallo del servidor de sesiones, no de la cuenta.');
    await page.getByRole('button', { name: 'Confirmar', exact: true }).click();
    await expect(page.getByText('Escalado').first()).toBeVisible();

    // **Desarrollo entra por su vista, que es el interno** (decisión del 2026-09-27), así que la
    // comprobación del principal se hace **cambiando a esa vista**, no dando por hecho lo que se ve.
    await salir(page);
    await entrarComo(page, desarrollo.email);
    await page.goto(`/tickets/${numero}`);

    const vista = page.getByRole('group', { name: 'Qué quieres ver' });
    await vista.getByRole('button', { name: 'Principal' }).click();
    await expect(page.getByText('La ficha')).toBeVisible();

    // **En el principal no tiene el control** (Desarrollo no escribe el principal).
    await expect(page.getByRole('button', { name: 'Editar las etiquetas' })).toHaveCount(0);

    // **Y en el interno sí**: las etiquetas son suyas también (decisión 85).
    await vista.getByRole('button', { name: 'Interno' }).click();

    // **Y la cabecera enseña primero el ticket de su equipo** (decisión 87): para Desarrollo el título es
    // el **interno**, y el principal va al lado como referencia.
    await expect(page.getByRole('heading', { level: 1, name: `INT-${numero}` })).toBeVisible();
    await expect(page.getByText(numero, { exact: true })).toBeVisible();
    await expect(page.getByLabel('Categoría')).toHaveCount(0);
    await page.getByRole('button', { name: 'Editar las etiquetas' }).click();

    const ventana = page.getByRole('dialog');
    await expect(ventana).toBeVisible();
    await ventana.getByLabel('Buscar por etiqueta').fill('correo');
    await ventana.getByRole('checkbox', { name: 'correo' }).check();
    await ventana.getByRole('button', { name: 'Guardar', exact: true }).click();
    await expect(page.getByText('Etiquetas guardadas.')).toBeVisible();

    // **Y quedan en el principal**, no en el interno: la etiqueta es del ticket principal (decisión 67).
    const principal = await detalleDe(request, suyo, numero);
    expect(principal.ticket.tags, 'la etiqueta la lleva el principal').toContain('correo');

    const interno = await detalleDe(request, suyoDeDesarrollo, `INT-${numero}`);
    expect(interno.ticket.tags, 'y el interno la hereda').toContain('correo');
  });

  test('se etiqueta escribiendo la arroba, y se ve el correo para no confundir a nadie', async ({
    page,
    request,
  }) => {
    test.setTimeout(60_000);

    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const desarrollo = await crearCuentaLista(request, {
      role: 'desarrollo',
      name: 'Arroba',
      lastName: `Desarrollo ${Date.now().toString(36)}`,
    });
    const ana = await crearCuentaLista(request);
    const { numero } = await ticketDePrueba(
      request,
      await tokenDe(request, ana.email),
      'Para etiquetar escribiendo',
    );

    await entrarComo(page, soporte.email);
    await page.goto(`/tickets/${numero}`);

    // Se escribe el comentario y, **dentro del texto**, la arroba con la parte del correo que es única
    // —la marca del momento—: es como lo pidió el responsable («@soporte2@demo.com»), y además el
    // nombre se repite entre cuentas de otras pasadas, así que la marca es lo que distingue a esta.
    const marca = desarrollo.email!.split('@')[0]!.replace('e2e-', '');

    const comentario = page.getByLabel('Escribe un comentario');
    await comentario.click();
    await page.keyboard.type(`Por favor @${marca}`);

    // Sale la lista filtrada, **con el correo al lado**: es lo que distingue a dos personas que se
    // llaman igual (`docs/interfaz-y-experiencia.md`, decisión 74).
    const sugerencia = page.getByRole('option').filter({ hasText: desarrollo.email! });
    await expect(sugerencia).toBeVisible();
    await expect(sugerencia).toContainText(`${desarrollo.name} ${desarrollo.lastName}`);

    // **Y sale junto al cursor** (decisión 89): dentro del comentario que se está escribiendo, no encima
    // del bloque ni al final de la página. Es lo que reportó el responsable: no la encontraba.
    const lista = page.getByRole('listbox');
    const cajaDeLaLista = (await lista.boundingBox())!;
    const cajaDelComentario = (await page.getByLabel('Escribe un comentario').boundingBox())!;

    expect(
      cajaDeLaLista.y,
      'la lista no puede quedarse por encima del comentario',
    ).toBeGreaterThanOrEqual(cajaDelComentario.y);
    expect(
      cajaDeLaLista.y - cajaDelComentario.y,
      'y tiene que estar cerca de donde se escribe',
    ).toBeLessThan(cajaDelComentario.height + 40);

    await sugerencia.click();

    // **La mención queda en el texto con el nombre**, y lo que se había escrito (`@Arroba`) se ha ido.
    const mencion = page.locator('[data-mencion]').last();
    await expect(mencion).toHaveText(`${desarrollo.name} ${desarrollo.lastName}`);
    await expect(page.getByLabel('Escribe un comentario')).not.toContainText(`@${marca}`);

    // **Y se ve como una mención** (decisión del responsable, 2026-09-29): con su arroba —que la pone el
    // estilo, para que las menciones de antes se vean igual— y con el color del tema, no como texto
    // normal. Se comprueba en lo que el navegador pinta, así que vale con los ocho temas.
    const comoSeVe = await mencion.evaluate((nodo) => ({
      arroba: getComputedStyle(nodo, '::before').content,
      color: getComputedStyle(nodo).color,
      fondo: getComputedStyle(nodo).backgroundColor,
      primario: getComputedStyle(document.documentElement).getPropertyValue('--primario').trim(),
    }));

    expect(comoSeVe.arroba, 'la mención lleva su arroba').toContain('@');
    expect(comoSeVe.color, 'y no se ve como el texto de al lado').not.toBe(
      await page.getByLabel('Escribe un comentario').evaluate((nodo) => getComputedStyle(nodo).color),
    );
    expect(comoSeVe.fondo, 'con su fondo, que la separa del texto').not.toBe('rgba(0, 0, 0, 0)');

    await page.getByRole('button', { name: 'Comentar' }).click();
    await expect(page.getByText('Comentario escrito.')).toBeVisible();

    // Y esa persona queda observando el ticket, como si se hubiera etiquetado con el botón.
    const suyo = await tokenDe(request, soporte.email);
    const detalle = await detalleDe(request, suyo, numero);
    expect(
      detalle.observers?.map((o) => o.account.id),
      'quien se etiqueta escribiendo tiene que quedar de observador',
    ).toContain(desarrollo.id);
  });

  test('el ticket se edita donde se muestra: asunto, descripción, categoría, etiquetas y observadores', async ({
    page,
    request,
  }) => {
    test.setTimeout(60_000);

    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const otro = await crearCuentaLista(request, {
      role: 'desarrollo',
      name: 'Observa',
      lastName: `Desarrollo ${Date.now().toString(36)}`,
    });
    const suyo = await tokenDe(request, soporte.email);

    const catalogo = (await request
      .get('/api/tickets/categories', { headers: { Authorization: `Bearer ${suyo}` } })
      .then((r) => r.json())) as { categories: { id: number; name: string }[] };
    const primera = catalogo.categories[0]!;
    const otraCategoria = catalogo.categories.find((c) => c.id !== primera.id)!;

    const { numero } = await ticketDePrueba(request, suyo, 'Un asunto que se cambia', '<p>Descripción original.</p>');

    // **El nombre de la etiqueta lo estrena el Administrador** en el catálogo (decisión 84): Soporte y
    // Desarrollo eligen de las que hay.
    const estrenada = await request.post('/api/tickets/tags', {
      headers: { Authorization: `Bearer ${await tokenDeFabrica(request)}` },
      data: { tag: 'prueba-de-sitio' },
    });
    expect([201, 409], await estrenada.text()).toContain(estrenada.status());

    await entrarComo(page, soporte.email);
    await page.goto(`/tickets/${numero}`);

    // **El asunto, con su lápiz, en el sitio**: se cambia y se guarda sin salir de la pantalla.
    await page.getByRole('button', { name: 'Editar el asunto' }).click();
    await page.getByLabel('Asunto', { exact: true }).fill('Un asunto ya cambiado');
    await page.getByRole('button', { name: 'Guardar', exact: true }).click();
    await expect(page.getByText('Ticket actualizado.')).toBeVisible();
    await expect(page.getByText('Un asunto ya cambiado')).toBeVisible();

    // **La categoría, con su desplegable en su propia línea**: se elige otra y se guarda sola.
    await page.getByLabel('Categoría').selectOption(String(otraCategoria.id));
    await expect(page.getByText('Categoría cambiada.')).toBeVisible();

    // **Las etiquetas, en su modal** (decisión 82): se abre, se busca, se marcan las que van y se
    // guarda. **La etiqueta tiene que existir en el catálogo** (decisión 84), así que ésta la ha
    // estrenado el Administrador en el catálogo antes de esta comprobación.
    await page.getByRole('button', { name: 'Editar las etiquetas' }).click();

    const ventanaDeEtiquetas = page.getByRole('dialog');
    await expect(ventanaDeEtiquetas).toBeVisible();

    await ventanaDeEtiquetas.getByLabel('Buscar por etiqueta').fill('prueba-de-sitio');
    await ventanaDeEtiquetas.getByRole('checkbox', { name: 'prueba-de-sitio' }).check();
    await ventanaDeEtiquetas.getByRole('button', { name: 'Guardar', exact: true }).click();
    await expect(page.getByText('Etiquetas guardadas.')).toBeVisible();
    await expect(page.getByText('prueba-de-sitio').first()).toBeVisible();

    // **Y los observadores**, con su «Añadir» y su buscador.
    await page.getByRole('button', { name: 'Añadir un observador' }).click();
    await page.getByRole('button', { name: `${otro.name} ${otro.lastName}`, exact: true }).click();
    await expect(page.getByText('Observador añadido.')).toBeVisible();

    // Lo que ha quedado guardado, que es lo que de verdad importa.
    const detalle = await detalleDe(request, suyo, numero);
    expect(detalle.ticket.subject).toBe('Un asunto ya cambiado');
    expect(detalle.ticket.category?.id).toBe(otraCategoria.id);
    expect(detalle.ticket.tags).toContain('prueba-de-sitio');
    expect(detalle.observers?.map((o) => o.account.id)).toContain(otro.id);

    // **Y se puede quitar** desde donde se ve: el aspa del observador y la de la etiqueta.
    await page
      .getByRole('button', { name: `Quitar de los observadores: ${otro.name} ${otro.lastName}` })
      .click();
    await expect(page.getByText('ya no sigue el ticket.')).toBeVisible();

    const sinObservador = await detalleDe(request, suyo, numero);
    expect(sinObservador.observers?.map((o) => o.account.id) ?? []).not.toContain(otro.id);

    // **Desarrollo no escribe el principal**: no se le enseña ningún lápiz, ni el desplegable de la
    // categoría, ni el campo de las etiquetas. **Los observadores sí**: seguir un ticket es de los dos
    // equipos, y quitarlos también (decisión 63), así que ahí sí tiene sus botones.
    await entrarComo(page, otro.email);
    await page.goto(`/tickets/${numero}`);
    await expect(page.getByRole('heading', { level: 1, name: numero })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Editar el asunto' })).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Editar la descripción' })).toHaveCount(0);
    await expect(page.getByLabel('Categoría')).toHaveCount(0);
    await expect(page.getByLabel('Añadir una etiqueta')).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Añadir un observador' })).toBeVisible();
  });

  test('las etiquetas se crean, se renombran y se retiran, y valen para todos sus tickets', async ({
    page,
    request,
  }) => {
    // **El catálogo lo mantiene sólo el Administrador** (decisión 83), así que este caso entra con la
    // cuenta de fábrica; Soporte sólo lo usa.
    const deFabrica = await tokenDeFabrica(request);
    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const suyo = await tokenDe(request, soporte.email);
    const marca = Date.now().toString(36);
    const nacida = `encargo-${marca}`;
    const renombrada = `renombrada-${marca}`;

    // Un ticket con la etiqueta puesta, para que el renombrado tenga a quién alcanzar y la retirada a
    // quién quitársela. **La etiqueta la pone el Administrador** desde el 2026-09-29 (decisión 84): el
    // catálogo se cura y **sólo él introduce nombres nuevos**; Soporte y Desarrollo eligen de los que
    // hay, así que el mismo `PATCH` con una etiqueta inventada le responde 422.
    const catalogo = (await request
      .get('/api/tickets/categories', { headers: { Authorization: `Bearer ${suyo}` } })
      .then((r) => r.json())) as { categories: { id: number }[] };
    const { numero } = await ticketDePrueba(request, suyo, 'Con etiqueta que se renombra', '<p>Prueba.</p>');

    const inventada = await request.patch(`/api/tickets/${numero}`, {
      headers: { Authorization: `Bearer ${suyo}` },
      data: { categoryId: catalogo.categories[0]!.id, tags: [nacida] },
    });
    expect(
      inventada.status(),
      'Soporte no puede estrenar una etiqueta: el catálogo lo cura el Administrador',
    ).toBe(422);
    expect(await inventada.text()).toContain('tickets.etiqueta.desconocida');

    // **El nombre nuevo lo estrena el Administrador** en el catálogo (el Administrador no toca tickets,
    // así que no puede hacerlo etiquetando), y a partir de ahí Soporte la usa como cualquier otra.
    const estrenada = await request.post('/api/tickets/tags', {
      headers: { Authorization: `Bearer ${deFabrica}` },
      data: { tag: nacida },
    });
    expect(estrenada.status(), await estrenada.text()).toBe(201);

    const conLaEtiqueta = await request.patch(`/api/tickets/${numero}`, {
      headers: { Authorization: `Bearer ${suyo}` },
      data: { categoryId: catalogo.categories[0]!.id, tags: [nacida] },
    });
    expect(conLaEtiqueta.status(), await conLaEtiqueta.text()).toBe(200);

    // **Soporte entra primero y NO puede mantener el catálogo** (decisión 83): la pantalla la ve —para
    // usarla y ver la cuenta de cada etiqueta— pero **sin los botones de crear, renombrar ni retirar**.
    await entrarComo(page, soporte.email);
    await page.goto('/tickets/categories');
    await expect(page.getByRole('heading', { name: 'Categorías y etiquetas' })).toBeVisible();

    const suyaDeSoporte = page.getByRole('heading', { name: 'Etiquetas' }).locator('../..');
    await expect(suyaDeSoporte.getByText(nacida)).toBeVisible();
    await expect(suyaDeSoporte.getByRole('button', { name: 'Renombrar' })).toHaveCount(0);
    await expect(suyaDeSoporte.getByRole('button', { name: 'Retirar' })).toHaveCount(0);
    await expect(
      page.getByRole('button', { name: 'Nueva etiqueta' }),
      'crear tampoco: el catálogo es del Administrador',
    ).toHaveCount(0);

    // Y entra el Administrador, que sí lo mantiene.
    await salir(page);
    await entrar(page, FABRICA.email, FABRICA.password);
    await page.goto('/tickets/categories');

    // **La mitad de las etiquetas enseña la que lleva el ticket**, con su cuenta.
    const mitadDeEtiquetas = page.getByRole('heading', { name: 'Etiquetas' }).locator('../..');
    await expect(mitadDeEtiquetas.getByText(nacida)).toBeVisible();
    await expect(mitadDeEtiquetas.getByText(nacida).locator('..')).toContainText('1');

    // **El Administrador la renombra**, y el cambio tiene que valer para el ticket.
    await mitadDeEtiquetas
      .locator('li')
      .filter({ hasText: nacida })
      .getByRole('button', { name: 'Renombrar' })
      .click();
    await page.getByRole('textbox', { name: 'Etiqueta' }).last().fill(renombrada);
    await page.getByRole('button', { name: 'Guardar', exact: true }).click();
    await expect(page.getByText('Etiqueta renombrada en todos los tickets que la llevan.')).toBeVisible();

    const trasRenombrar = await detalleDe(request, suyo, numero);
    expect(trasRenombrar.ticket.tags, 'el renombrado vale para los tickets').toContain(renombrada);
    expect(trasRenombrar.ticket.tags).not.toContain(nacida);

    const suya = page.getByRole('heading', { name: 'Etiquetas' }).locator('../..');
    await suya.locator('li').filter({ hasText: renombrada }).getByRole('button', { name: 'Retirar' }).click();

    // **Pregunta antes, y dice a cuántos toca**: retirar una etiqueta la quita de sus tickets.
    await expect(page.getByText(`Se va a retirar «${renombrada}»`)).toBeVisible();
    await page.getByRole('button', { name: 'Retirar', exact: true }).last().click();
    await expect(page.getByText('Etiqueta retirada y quitada de sus tickets.')).toBeVisible();

    const trasRetirar = await detalleDe(request, suyo, numero);
    expect(trasRetirar.ticket.tags ?? [], 'la etiqueta se va de sus tickets').not.toContain(renombrada);
  });

  test('el alta no deja crear sin categoría, y no pide etiquetas', async ({
    page,
    request,
  }) => {
    const ana = await crearCuentaLista(request);
    const categoria = (await request.get('/api/tickets/categories', {
      headers: { Authorization: `Bearer ${await tokenDe(request, ana.email)}` },
    }).then((r) => r.json())) as { categories: { id: number; name: string }[] };
    const elegida = categoria.categories[0]!;

    await entrarComo(page, ana.email);
    await page.goto('/tickets/new');

    // **La categoría es obligatoria**: el desplegable arranca en «Elige una categoría» y no vale.
    const selector = page.getByLabel('Categoría');
    await expect(selector).toBeVisible();
    await expect(selector).toHaveValue('');

    await page.getByLabel('Asunto').fill('Con categoría');
    await page.getByLabel('Descripción').fill('Un ticket clasificado.');
    await selector.selectOption(String(elegida.id));

    // **El alta no pide etiquetas** (decisión 82): etiquetar es de Soporte y Desarrollo, y el campo no
    // está —ni el de las sugerencias—, así que quien abre el ticket sólo elige la categoría.
    await expect(page.getByLabel('Etiquetas')).toHaveCount(0);
    await expect(page.getByLabel('Añadir una etiqueta')).toHaveCount(0);

    await page.getByRole('button', { name: 'Crear el ticket' }).click();
    await expect(page.getByRole('heading', { level: 1 })).toContainText('CS-');

    // Y en la ficha están la categoría y las etiquetas, tal y como las guardó el backend. **La
    // categoría es un desplegable** desde el 2026-09-28 (se cambia donde se muestra), así que lo que se
    // comprueba es su valor, no un texto.
    await expect(page.getByLabel('Categoría')).toHaveValue(String(elegida.id));

    // **Y el usuario no ve etiquetas en su ticket** (decisión 82): ni las del ticket ni el control para
    // ponerlas. Ve la categoría, que es lo que él eligió.
    await expect(page.getByText('Etiquetas')).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Etiquetas' })).toHaveCount(0);
  });

  test('se filtra por categoría y por etiqueta, y la búsqueda las encuentra', async ({
    page,
    request,
  }, info) => {
    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const suyo = await tokenDe(request, soporte.email);

    const catalogo = (await request
      .get('/api/tickets/categories', { headers: { Authorization: `Bearer ${suyo}` } })
      .then((r) => r.json())) as { categories: { id: number; name: string }[] };
    const categoria = catalogo.categories.find((c) => c.name === 'Red') ?? catalogo.categories[0]!;
    const marca = Date.now().toString(36);
    const etiqueta = `prueba-${marca}`;

    // **El nombre de la etiqueta lo estrena el Administrador** en el catálogo (decisión 84), y quien la
    // pone en el ticket es Soporte: el catálogo se cura, pero etiquetar es de los dos equipos.
    const estrenada = await request.post('/api/tickets/tags', {
      headers: { Authorization: `Bearer ${await tokenDeFabrica(request)}` },
      data: { tag: etiqueta },
    });
    expect(estrenada.status(), await estrenada.text()).toBe(201);

    const { numero } = await ticketDePrueba(request, suyo, 'Clasificado y etiquetado', '<p>Prueba.</p>');
    const clasificado = await request.patch(`/api/tickets/${numero}`, {
      headers: { Authorization: `Bearer ${suyo}` },
      data: { categoryId: categoria.id, tags: [etiqueta] },
    });
    expect(clasificado.status(), await clasificado.text()).toBe(200);

    await entrarComo(page, soporte.email);
    await page.goto('/tickets/main');
    await buscarEnLaLista(page, numero);

    // **Los chips en el renglón**, debajo del número: la categoría y la etiqueta.
    const renglon = page.locator('tr, li').filter({ hasText: numero }).last();
    await expect(renglon).toContainText(categoria.name);
    await expect(renglon).toContainText(etiqueta);

    // **Y se filtra por las dos**, que es para lo que se pidieron. Desde el 2026-09-27 los criterios
    // viven **en el modal** (decisión 69): se abre, se ponen los dos —el chip de la categoría y el texto—
    // y se aplica. Al no haber nada repintándose por detrás mientras el modal está abierto, el chip se
    // puede pulsar **en los dos anchos**, que es lo que no se podía con la fila de filtros de antes.
    const abrirLaBusqueda = () => page.getByRole('button', { name: 'Buscar', exact: true }).click();
    const elModal = () => page.getByRole('dialog');
    const aplicar = () =>
      elModal().getByRole('button', { name: 'Buscar', exact: true }).click();

    await page.goto('/tickets/main');
    await abrirLaBusqueda();
    await elModal()
      .getByRole('group', { name: 'Categoría' })
      .getByRole('button', { name: categoria.name, exact: true })
      .click();
    await elModal().getByLabel('Buscar por número, asunto o texto').fill(numero);
    await aplicar();

    await expect(page.getByRole('link', { name: numero })).toBeVisible();

    // **Con otra categoría y el mismo número, no sale**: es el filtro el que lo deja fuera, no la
    // búsqueda. (El modal se abre con lo que ya estaba filtrado, así que sólo se cambia el chip.)
    await abrirLaBusqueda();
    await elModal()
      .getByRole('group', { name: 'Categoría' })
      .getByRole('button', { name: 'Impresoras', exact: true })
      .click();
    await aplicar();
    await expect(page.getByRole('link', { name: numero })).toHaveCount(0);

    // **El filtro de etiqueta**: se quita la categoría y la búsqueda, y se deja sólo la etiqueta —que es
    // única de esta prueba—, así que el ticket sale por ella y no por otra cosa.
    await abrirLaBusqueda();
    await elModal()
      .getByRole('group', { name: 'Categoría' })
      .getByRole('button', { name: 'Todas', exact: true })
      .click();
    await elModal().getByLabel('Buscar por número, asunto o texto').fill('');
    await elModal().getByLabel('Etiqueta').fill(etiqueta);
    await aplicar();
    await expect(page.getByRole('link', { name: numero })).toBeVisible();

    // **Y la búsqueda de texto la encuentra**: sin ningún filtro puesto, buscar la etiqueta la trae.
    await abrirLaBusqueda();
    await elModal().getByLabel('Etiqueta').fill('');
    await elModal().getByLabel('Buscar por número, asunto o texto').fill(etiqueta);
    await aplicar();
    await expect(page.getByRole('link', { name: numero })).toBeVisible();

    // Y la búsqueda también llega **por el nombre de la categoría**, que es la otra mitad de la decisión
    // 68. Se comprueba contra la API porque por pantalla «Red» devuelve treinta y el nuestro puede caer
    // en otra página; lo que se mide es que la búsqueda lo alcanza, no en qué página lo pone.
    const porCategoria = await request.get('/api/tickets', {
      headers: { Authorization: `Bearer ${suyo}` },
      params: { q: categoria.name, perPage: 100 },
    });
    const buscados = (await porCategoria.json()) as { tickets: { number: string }[] };
    expect(buscados.tickets.some((t) => t.number === numero)).toBe(true);
  });
});
