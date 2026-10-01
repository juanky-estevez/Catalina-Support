import { expect, type APIRequestContext, type Page } from '@playwright/test';

/**
 * Lo que comparten las pruebas que necesitan un correo o una cuenta.
 *
 * Es la parte que hace que estas pruebas valgan la pena: en vez de dar por hecho que el enlace
 * llegó, **se lee del buzón de pruebas**, que es exactamente lo que haría una persona.
 */

/** La dirección del buzón de pruebas, por dentro de la red de los contenedores. */
export const BUZON = process.env['MAILPIT_URL'] ?? 'http://mail:8025';

/** La cuenta de fábrica: la única que existe siempre, sin crear nada. */
export const FABRICA = {
  email: 'admin',
  password: process.env['ADMIN_PASSWORD'] ?? '',
};

interface Mensaje {
  ID: string;
  Subject: string;
  To: { Address: string }[];
}

/**
 * Espera a que llegue un correo a esa dirección y devuelve su HTML.
 *
 * Espera de verdad, en lugar de dormir: el envío va en segundo plano a propósito, así que el
 * correo puede tardar un poco en aparecer.
 */
export async function esperarMensaje(
  peticion: APIRequestContext,
  destinatario: string,
  asunto: string,
): Promise<string> {
  await expect
    .poll(
      async () => {
        const respuesta = await peticion.get(`${BUZON}/api/v1/messages?limit=50`);
        const cuerpo = (await respuesta.json()) as { messages: Mensaje[] };
        return cuerpo.messages.some(
          (mensaje) => mensaje.To[0]?.Address === destinatario && mensaje.Subject === asunto,
        );
      },
      { message: `No llegó a ${destinatario} ningún correo con el asunto «${asunto}»`, timeout: 15_000 },
    )
    .toBe(true);

  const respuesta = await peticion.get(`${BUZON}/api/v1/messages?limit=50`);
  const cuerpo = (await respuesta.json()) as { messages: Mensaje[] };
  const mensaje = cuerpo.messages.find(
    (candidato) => candidato.To[0]?.Address === destinatario && candidato.Subject === asunto,
  );

  const detalle = await peticion.get(`${BUZON}/api/v1/message/${mensaje!.ID}`);
  const correo = (await detalle.json()) as { HTML: string; Text: string };

  return correo.HTML;
}

/** Lo mismo, pero devolviendo el enlace que trae el correo. */
export async function esperarEnlaceDelCorreo(
  peticion: APIRequestContext,
  destinatario: string,
  asunto: string,
): Promise<string> {
  const html = await esperarMensaje(peticion, destinatario, asunto);
  const enlace = /href="([^"]+)"/.exec(html)?.[1];
  expect(enlace, `El correo «${asunto}» no traía ningún enlace`).toBeTruthy();

  return enlace!;
}

/** Vacía el buzón, para que una prueba no lea el correo de la anterior. */
export async function vaciarBuzon(peticion: APIRequestContext): Promise<void> {
  await peticion.delete(`${BUZON}/api/v1/messages`);
}

/** La cuenta con la que se prueba el camino local, distinta en cada ejecución. */
export function cuentaDePrueba(): { name: string; lastName: string; email: string } {
  const marca = Date.now().toString(36);
  return { name: 'Prueba', lastName: 'Automática', email: `e2e-${marca}@demo.com` };
}

/** La contraseña que se le pone a las cuentas de prueba: larga, y la misma siempre. */
export const CONTRASENA_DE_PRUEBA = 'una-contraseña-larga';

/** El asunto del correo del alta: es el que trae el enlace, y el que hay que buscar. */
export const ASUNTO_DEL_ALTA = 'Establece tu contraseña de Catalina Support';

/** Una cuenta de prueba ya terminada: existe, tiene contraseña y se puede entrar con ella. */
export interface CuentaListaDePrueba {
  readonly id: number;
  readonly name: string;
  readonly lastName: string;
  readonly email: string;
  readonly password: string;
}

/** Entra por la API con una cuenta y devuelve su token, para montar los datos de una prueba. */
export async function tokenDe(
  peticion: APIRequestContext,
  email: string,
  password = CONTRASENA_DE_PRUEBA,
): Promise<string> {
  const respuesta = await peticion.post('/api/auth/login', { data: { email, password } });
  const cuerpo = (await respuesta.json()) as { token?: string };
  expect(cuerpo.token, await respuesta.text()).toBeTruthy();

  return cuerpo.token!;
}

/** Entra por la API con la cuenta de fábrica y devuelve su token, para montar los datos de una prueba. */
export async function tokenDeFabrica(peticion: APIRequestContext): Promise<string> {
  const respuesta = await peticion.post('/api/auth/login', {
    data: { email: FABRICA.email, password: FABRICA.password },
  });

  const cuerpo = (await respuesta.json()) as { token?: string };
  expect(cuerpo.token, await respuesta.text()).toBeTruthy();

  return cuerpo.token!;
}

/**
 * Deja una cuenta lista para entrar con ella: **dada de alta y con su contraseña puesta**.
 *
 * Se hace por la API y **leyendo el correo del buzón de pruebas**, que es el mismo camino que sigue
 * una persona: no se inventa una contraseña por detrás ni se toca la base de datos.
 */
export async function crearCuentaLista(
  peticion: APIRequestContext,
  opciones: {
    role?: string;
    language?: string;
    token?: string;
    /**
     * El nombre y los apellidos se pueden fijar desde la prueba.
     *
     * Hace falta cuando la prueba tiene que **reconocer a esa persona en pantalla** entre las demás
     * —el buscador de a quién etiquetar enseña nombres, y todas las cuentas de prueba se llaman igual—:
     * con un nombre propio, el botón de esa persona es el único que se llama así.
     */
    name?: string;
    lastName?: string;
  } = {},
): Promise<CuentaListaDePrueba> {
  const token = opciones.token ?? (await tokenDeFabrica(peticion));
  const cuenta = { ...cuentaDePrueba(), name: opciones.name ?? '', lastName: opciones.lastName ?? '' };
  cuenta.name = opciones.name ?? 'Prueba';
  cuenta.lastName = opciones.lastName ?? 'Automática';

  const alta = await peticion.post('/api/users', {
    headers: { Authorization: `Bearer ${token}` },
    data: {
      ...cuenta,
      role: opciones.role ?? 'usuario',
      origin: 'local',
      language: opciones.language ?? 'es',
    },
  });
  expect(alta.status(), await alta.text()).toBe(201);

  const cuerpo = (await alta.json()) as { user: { id: number } };

  const enlace = await esperarEnlaceDelCorreo(peticion, cuenta.email, ASUNTO_DEL_ALTA);
  const conPassword = await peticion.post('/api/auth/password/reset', {
    data: { token: enlace.split('#token=')[1], password: CONTRASENA_DE_PRUEBA },
  });
  expect(conPassword.status(), await conPassword.text()).toBe(200);

  return { ...cuenta, id: cuerpo.user.id, password: CONTRASENA_DE_PRUEBA };
}

/** Entra con esa cuenta por la pantalla de entrada. */
export async function entrar(page: Page, email: string, password: string): Promise<void> {
  await page.goto('/login');

  // **Con el método en Keycloak no hay formulario**, y la cuenta de fábrica tiene su puerta discreta:
  // es la única que entra siempre, sea cual sea el método (docs/usuarios-y-permisos.md, sección 8).
  //
  // **Se espera a que una de las dos cosas esté**, y no se pregunta nada más llegar: la pantalla pide
  // sus caminos al backend al montarse, así que preguntar antes de que pinte devolvía «no hay
  // formulario» con el formulario por aparecer, y el clic se quedaba esperando un botón que en una
  // instalación `local` no existe. Es la misma carrera que ya apareció con el cajón del menú.
  const correo = page.getByLabel('Correo electrónico');
  const puerta = page.getByRole('button', { name: 'Entrar como administrador' });
  await expect(correo.or(puerta).first()).toBeVisible();

  if (!(await correo.isVisible())) {
    await puerta.click();
  }

  await correo.fill(email);
  await page.getByLabel('Contraseña', { exact: true }).fill(password);
  // **El nombre va exacto**: el botón de la pantalla de entrada se llama «Entrar» y el de Keycloak
  // «Entrar con Keycloak», y sin `exact` el localizador encuentra los dos y el clic no sabe a cuál ir.
  await page.getByRole('button', { name: 'Entrar', exact: true }).click();
}

/** Sale por donde sale una persona: el botón del menú lateral. */
export async function salir(page: Page): Promise<void> {
  await abrirMenuSiEsCajon(page);
  await page.getByRole('button', { name: 'Salir' }).click();

  // **Salir se confirma** (decisión del responsable, 2026-09-29): el botón abre su ventana y hay que
  // confirmar. El «Salir» de dentro de la ventana es el único que cierra la sesión.
  await page.getByRole('dialog').getByRole('button', { name: 'Salir' }).click();
  await expect(page).toHaveURL(/\/login$/);
}

/**
 * Entra con una cuenta local, saliendo antes de la que hubiera.
 *
 * Con sesión puesta la pantalla de entrada no se enseña, así que primero hay que salir: es lo que
 * haría una persona que va a entrar con otra cuenta.
 */
export async function entrarComo(
  page: Page,
  email: string,
  password = CONTRASENA_DE_PRUEBA,
): Promise<void> {
  const salirBoton = page.getByRole('button', { name: 'Salir' });

  if ((await salirBoton.count()) > 0) {
    await salir(page);
  }

  await entrar(page, email, password);

  // Entrar es navegar: se espera a estar dentro antes de tocar nada del armazón.
  await expect(page).not.toHaveURL(/\/login$/);
}

/** El token que el frontend guarda en el navegador, leído como lo leería el interceptor. */
export async function tokenGuardado(page: Page): Promise<string | null> {
  return page.evaluate(() => localStorage.getItem('catalina-support.token'));
}

/**
 * El botón de salir, el perfil y los controles viven en el menú lateral, que **en móvil es un cajón**:
 * hay que abrirlo para llegar a ellos, igual que haría una persona.
 *
 * Se decide mirando si el botón de abrir está a la vista, y no por el ancho del proyecto: así la misma
 * prueba vale para PC y para móvil sin preguntar en qué proyecto corre.
 *
 * **Se espera a que el cajón termine de entrar.** Entra deslizándose (200 ms) y, mientras se mueve,
 * pulsar algo de dentro es una carrera: el punto del clic se calcula sobre una caja que se está
 * desplazando, y el clic acaba donde ya no está el botón. Playwright lo cuenta como «otro elemento
 * intercepta el clic», que es un mensaje que despista y no dice nada del problema de verdad.
 */
export async function abrirMenuSiEsCajon(page: Page): Promise<void> {
  const abrir = page.getByRole('button', { name: 'Abrir el menú' });
  const menu = page.locator('aside');

  // **Se espera a que el armazón esté montado antes de decidir nada.** `isVisible()` no espera:
  // llamarlo nada más navegar devolvía «no está a la vista» con el menú todavía por pintar, así que
  // el cajón no se abría y el clic siguiente se quedaba esperando —treinta segundos— a un botón que
  // existía pero estaba fuera de la pantalla. Pasó en dos ejecuciones seguidas de la suite entera y
  // no en una prueba suelta, que es justo la firma de una carrera.
  await expect(menu).toBeAttached();
  await expect(menu.getByRole('navigation', { name: 'Opciones' })).toBeAttached();

  if (!(await abrir.isVisible())) {
    return;
  }

  await abrir.click();

  await expect
    .poll(async () => (await menu.boundingBox())?.x, {
      message: 'El cajón del menú no ha terminado de abrirse',
      timeout: 5_000,
    })
    .toBe(0);
}

/**
 * El método de entrada que tiene puesta la instalación: `local`, `ad` o `keycloak`.
 *
 * Lo dice el endpoint público de los caminos de entrada, que es el mismo que lee la pantalla.
 */
export async function metodoDeEntrada(peticion: APIRequestContext): Promise<string> {
  const caminos = (await (await peticion.get('/api/auth/methods')).json()) as { method?: string };

  return caminos.method ?? 'local';
}

/**
 * Busca en una lista de tickets: **abre el modal, escribe y aplica**.
 *
 * Es como se busca desde el 2026-09-27 (decisión del responsable): los criterios viven en **un modal**
 * que abre el botón «Buscar», y en la lista queda sólo el resumen de lo filtrado. Antes había tres
 * filas de filtros a la vista, y la de categorías se salía de la pantalla en cuanto el catálogo
 * crecía.
 *
 * Los dos botones se llaman «Buscar» —el que abre y el que aplica—, así que el de dentro se busca
 * **dentro del diálogo**: es lo que hace una persona, que ve uno antes y otro después.
 */
export async function buscarEnLaLista(page: Page, texto: string): Promise<void> {
  await page.getByRole('button', { name: 'Buscar', exact: true }).click();

  const modal = page.getByRole('dialog');

  await modal.getByLabel('Buscar por número, asunto o texto').fill(texto);
  await modal.getByRole('button', { name: 'Buscar', exact: true }).click();
}

/**
 * Pone un criterio del modal de búsqueda y aplica: es para los chips (tipo, estado, categoría…).
 *
 * `grupo` es el nombre del grupo del conmutador —«Tipo», «Estado», «Categoría»— y `opcion` el del
 * botón que se pulsa dentro.
 */
export async function filtrarEnLaLista(page: Page, grupo: string, opcion: string): Promise<void> {
  await page.getByRole('button', { name: 'Buscar', exact: true }).click();

  const modal = page.getByRole('dialog');

  await modal.getByRole('group', { name: grupo }).getByRole('button', { name: opcion }).click();
  await modal.getByRole('button', { name: 'Buscar', exact: true }).click();
}

/**
 * Pone el método de entrada de la instalación, **como lo haría un administrador desde Configuración**.
 *
 * Estas pruebas lo necesitan porque **la instalación entra por un método a la vez**: el camino de AD y
 * el de Keycloak no se pueden probar con el método en `local`, y por eso cada uno lo pone al empezar y
 * **lo devuelve a `local` al terminar**.
 *
 * Se guarda con la cuenta de fábrica, que **entra siempre**, sea cual sea el método: es lo que hace que
 * una prueba que falle a medias no deje el entorno sin puerta.
 */
export async function ponerElMetodo(peticion: APIRequestContext, metodo: string): Promise<void> {
  const token = await tokenDeFabrica(peticion);

  const ajustes = await peticion.get('/api/settings', { headers: { Authorization: `Bearer ${token}` } });
  expect(ajustes.status(), await ajustes.text()).toBe(200);

  const configuracion = (await ajustes.json()) as Record<string, unknown>;

  const guardado = await peticion.put('/api/settings', {
    headers: { Authorization: `Bearer ${token}` },
    data: { ...configuracion, entryMethod: metodo },
  });
  expect(guardado.status(), await guardado.text()).toBe(200);
}

/**
 * ¿Está el servicio de pruebas levantado? Se pregunta **por el botón de Configuración**, que es lo que
 * pregunta una persona: se manda la configuración guardada —sin la contraseña y sin el secreto, que no
 * salen nunca por la API— y se mira si contesta que sí.
 *
 * Es lo que permite saltarse las pruebas de los caminos de directorio cuando el perfil `auth` no está
 * levantado, **sin depender del método que esté puesto**, y las del motor de IA cuando no está su
 * contenedor.
 */
export async function respondeAlProbar(
  peticion: APIRequestContext,
  camino: 'directory' | 'keycloak' | 'ai',
): Promise<boolean> {
  const token = await tokenDeFabrica(peticion);

  const ajustes = await peticion.get('/api/settings', { headers: { Authorization: `Bearer ${token}` } });
  if (ajustes.status() !== 200) {
    return false;
  }

  const configuracion = (await ajustes.json()) as Record<string, unknown>;

  // El motor de IA no tiene su configuración anidada: su dirección vive en la raíz y su prueba la
  // recibe como `url`. Los otros dos caminos mandan su objeto tal cual (docs/modules/ai.md).
  const cuerpo = camino === 'ai' ? { url: configuracion['aiUrl'] } : configuracion[camino];

  const prueba = await peticion.post(`/api/settings/${camino}/test`, {
    headers: { Authorization: `Bearer ${token}` },
    data: cuerpo,
  });

  return prueba.status() === 200;
}

/**
 * Elige un estado en el desplegable del ticket (decisiones 80 y 81).
 *
 * El control **cambia el estado**, no ejecuta acciones sueltas: sus opciones son los estados a los que
 * el ticket puede pasar —`en progreso`, `en espera`, `escalado`, `resuelto`, `cerrado`— y cada uno abre
 * su ventana de confirmación. El control es uno por ticket: `#acciones-<número>`.
 */
export async function elegirEstado(page: Page, numero: string, estado: string): Promise<void> {
  await page.locator(`#acciones-${numero}`).selectOption(estado);
}

/**
 * Reasigna el ticket a una cuenta (decisión 50): **el botón que está junto al responsable abre el modal**
 * y la elección se hace dentro (decisión del responsable, 2026-09-29). El campo «Asignar a» ya no vive
 * suelto en la ficha.
 */
export async function asignarA(page: Page, cuenta: number): Promise<void> {
  await page.getByRole('button', { name: 'Asignar a', exact: true }).click();

  const ventana = page.getByRole('dialog');
  await ventana.getByLabel('Asignar a').selectOption(String(cuenta));
  await ventana.getByRole('button', { name: 'Asignar', exact: true }).click();
}
