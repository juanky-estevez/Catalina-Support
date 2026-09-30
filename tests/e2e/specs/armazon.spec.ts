import { expect, test, type TestInfo } from '@playwright/test';

import { cuentaDePrueba, entrar, esperarEnlaceDelCorreo, FABRICA, vaciarBuzon } from '../ayudas';

/**
 * El armazón: el menú lateral y lo que hay dentro.
 *
 * Se prueba **con la sesión puesta**, que es cuando existe el armazón: sin sesión, las pantallas de
 * la sesión van a pantalla completa y no hay menú.
 */
/**
 * Hay cosas del armazón que sólo tienen sentido en un sitio: el menú **se pliega en PC** y en móvil
 * **es un cajón** que hay que abrir. Cada prueba dice dónde corre, en vez de dar por hecho que el
 * navegador es grande.
 */
const esPc = (info: TestInfo) => info.project.name === 'pc';

/**
 * En móvil el menú es un cajón: para mirar lo que hay dentro hay que abrirlo, como haría una persona.
 */
async function abrirCajonSiHaceFalta(page: import('@playwright/test').Page, info: TestInfo) {
  if (!esPc(info)) {
    await page.getByRole('button', { name: 'Abrir el menú' }).click();
  }
}

test.describe('El armazón', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede entrar');
    await entrar(page, FABRICA.email, FABRICA.password);
    // Un Administrador entra en la lista de usuarios: el inicio es de quien no tiene pantallas.
    await expect(page).toHaveURL(/\/users$/);
  });

  test('el menú tiene las tres zonas, con el logo arriba y los controles abajo', async ({
    page,
    request,
  }, info) => {
    await abrirCajonSiHaceFalta(page, info);
    const menu = page.locator('aside');

    // Zona 1: el producto, con el logo de la instalación y **su nombre**. El nombre no se escribe
    // aquí: se lee de la instalación, porque es configurable y la prueba no tiene por qué saber cómo
    // se llama hoy (docs/modules/settings.md).
    await expect(menu.locator('app-logo img')).toBeVisible();

    const marca = (await (await request.get('/api/settings/brand')).json()) as { name: string };
    await expect(menu.getByText(marca.name)).toBeVisible();

    // Zona 2: las opciones, con su navegación etiquetada. **«Nuevo ticket» ya no está** (decisión
    // del responsable, 2026-09-26): se confundía con un módulo y vive en el botón de la bandeja.
    const opciones = menu.getByRole('navigation', { name: 'Opciones' });
    await expect(opciones).toBeVisible();
    await expect(opciones.getByRole('link', { name: 'Nuevo ticket' })).toHaveCount(0);

    // Zona 3: los controles, con el nombre de quien ha entrado, el tema y el idioma.
    await expect(menu.getByText('Administrador').first()).toBeVisible();
    await expect(menu.getByLabel('Tema')).toBeVisible();
    await expect(menu.getByLabel('Cambiar idioma')).toBeVisible();
    await expect(menu.getByRole('button', { name: 'Salir' })).toBeVisible();
  });

  /**
   * **El idioma y el tema son el mismo control** (decisión del responsable, 2026-09-26): los dos son
   * desplegables, van uno debajo del otro y **miden lo mismo**. Se mide el ancho de verdad, que es lo
   * único que distingue «son iguales» de «casi».
   */
  test('el desplegable del idioma ocupa el mismo ancho que el del tema', async ({ page }, info) => {
    await abrirCajonSiHaceFalta(page, info);
    const menu = page.locator('aside');

    const anchoDe = (etiqueta: string) =>
      menu.getByLabel(etiqueta).evaluate((nodo) => nodo.getBoundingClientRect().width);

    const anchoTema = await anchoDe('Tema');
    const anchoIdioma = await anchoDe('Cambiar idioma');

    expect(anchoIdioma).toBeGreaterThan(0);
    expect(anchoIdioma).toBe(anchoTema);

    // Y **ninguno de los dos ocupa toda la columna por su cuenta**: los dos se estiran con el menú,
    // que es lo que los hace iguales al desplegarlo o al plegarlo.
    const anchoMenu = await menu.evaluate((nodo) => nodo.getBoundingClientRect().width);
    expect(anchoIdioma).toBeLessThanOrEqual(anchoMenu);
  });

  test('el menú se pliega, se queda en iconos y lo recuerda', async ({ page }, info) => {
    test.skip(!esPc(info), 'El menú se pliega en PC, que es donde hay sitio');

    const menu = page.locator('aside');

    // **El ancho de antes de plegar**, que es lo que se comprueba al final que vuelve. Se mide en vez
    // de preguntar por la clase: el ancho exacto es una decisión de diseño que cambia —eran 16 rem y
    // son 15—, y una prueba que dice `w-64` se rompe cada vez que alguien lo ajusta.
    const anchoDesplegado = await menu.evaluate((nodo) => nodo.getBoundingClientRect().width);
    expect(anchoDesplegado).toBeGreaterThan(200);

    // **Los nombres van exactos**: «Desplegar el menú» contiene «plegar el menú», así que sin `exact`
    // el localizador del botón de desplegar encuentra también el de plegar, y el error que sale
    // —«resolved to 2 elements»— habla de otra cosa.
    await menu.getByRole('button', { name: 'Plegar el menú', exact: true }).click();

    // Plegado: sin el nombre del producto a la vista, pero con las opciones (en iconos) y el enlace
    // de texto para quien no ve la pantalla.
    await expect(menu.getByRole('link', { name: 'Configuración' })).toBeAttached();

    // **Se espera a que el plegado esté puesto**, no sólo a haber pulsado: el ancho lo decide Angular
    // al repintar, y medir antes de que repinte es medir el menú de antes.
    await expect(menu).toHaveClass(/w-16/);
    const anchoPlegado = await menu.evaluate((nodo) => nodo.getBoundingClientRect().width);

    await page.reload();
    const trasRecargar = page.locator('aside');
    await expect(trasRecargar).toHaveClass(/w-16/);
    const anchoTrasRecargar = await trasRecargar.evaluate((nodo) => nodo.getBoundingClientRect().width);
    expect(anchoTrasRecargar).toBe(anchoPlegado);
    expect(anchoPlegado).toBeLessThan(100);

    // Y se despliega otra vez, desde el botón de la cabecera (plegado hay dos: el de la cabecera y el
    // de abajo, y los dos despliegan, pero el de la cabecera es el que se ve primero).
    await trasRecargar.getByRole('button', { name: 'Desplegar el menú', exact: true }).first().click();

    // **Y vuelve a ser el menú de antes**: el mismo ancho con el que empezó, medido.
    await expect
      .poll(async () => (await trasRecargar.boundingBox())?.width, {
        message: 'el menú no volvió a su ancho',
      })
      .toBe(anchoDesplegado);
    await expect(trasRecargar.getByRole('button', { name: 'Plegar el menú', exact: true })).toBeVisible();
  });

  test('el menú lleva a Configuración y marca dónde estás', async ({ page }, info) => {
    test.skip(!esPc(info), 'En móvil el enlace está en el cajón: eso se prueba en su caso');

    await page.getByRole('link', { name: 'Configuración' }).click();

    await expect(page).toHaveURL(/\/settings$/);
    await expect(page.getByRole('heading', { name: 'Configuración' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Configuración' })).toHaveClass(/bg-superficie-suave/);
  });

  test('la pantalla de entrada no lleva el menú', async ({ page }, info) => {
    test.skip(!esPc(info), 'En móvil el botón de salir está en el cajón');

    // Salir se confirma en su ventana (decisión del responsable, 2026-09-29).
    await page.getByRole('button', { name: 'Salir' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Salir' }).click();

    await expect(page).toHaveURL(/\/login$/);
    await expect(page.locator('aside')).toHaveCount(0);
    await expect(page.locator('header')).toHaveCount(0);
  });

  test('en móvil el menú es un cajón: se abre, se ve y se cierra', async ({ page }, info) => {
    test.skip(esPc(info), 'En PC el menú está siempre a la vista');

    // La tira de arriba sólo existe en pantallas pequeñas.
    const tira = page.locator('header');
    await expect(tira).toBeVisible();

    const menu = page.locator('aside');
    await expect(menu).not.toBeInViewport();

    await page.getByRole('button', { name: 'Abrir el menú' }).click();
    await expect(menu).toBeInViewport();
    await expect(menu.getByRole('link', { name: 'Configuración' })).toBeVisible();

    await menu.getByRole('button', { name: 'Cerrar el menú' }).click();
    await expect(menu).not.toBeInViewport();
  });
});

/**
 * Los permisos, que es lo que decide qué ve cada papel: el menú no enseña lo que no se puede hacer y
 * el backend lo rechaza aunque alguien escriba la dirección a mano.
 */
test.describe('El menú, por papel', () => {
  test('un usuario no ve Configuración, y si entra a mano se le dice que no', async ({
    page,
    request,
  }, info) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede dar de alta nada');

    await vaciarBuzon(request);

    const admin = await request.post('/api/auth/login', {
      data: { email: FABRICA.email, password: FABRICA.password },
    });
    const { token } = (await admin.json()) as { token: string };

    // Una cuenta de usuario, creada y activada como en la vida real: por el correo.
    const cuenta = cuentaDePrueba();
    const alta = await request.post('/api/users', {
      headers: { Authorization: `Bearer ${token}` },
      data: { ...cuenta, role: 'usuario', origin: 'local', language: 'es' },
    });
    expect(alta.status(), await alta.text()).toBe(201);

    const enlace = await esperarEnlaceDelCorreo(
      request,
      cuenta.email,
      'Establece tu contraseña de Catalina Support',
    );
    await request.post('/api/auth/password/reset', {
      data: { token: enlace.split('#token=')[1], password: 'una-contraseña-larga' },
    });

    await entrar(page, cuenta.email, 'una-contraseña-larga');
    // Un usuario entra en **sus tickets**: el inicio provisional desapareció con las pantallas de
    // producto, que era lo que le faltaba.
    await expect(page).toHaveURL(/\/tickets$/);
    await abrirCajonSiHaceFalta(page, info);

    // El menú no le enseña lo que no puede hacer…
    await expect(page.locator('aside').getByRole('link', { name: 'Configuración' })).toHaveCount(0);
    // …y enseña su papel, no el de otro.
    await expect(page.locator('aside').getByText('Usuario')).toBeVisible();

    // Y si escribe la dirección a mano, la guarda lo lleva a «sin permiso».
    await page.goto('/settings');
    await expect(page).toHaveURL(/\/forbidden$/);
    await expect(page.getByRole('heading', { name: 'No puedes ver esta pantalla' })).toBeVisible();
  });
});
