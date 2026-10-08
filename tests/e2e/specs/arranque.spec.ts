import { expect, test } from '@playwright/test';

/**
 * El caso más básico de todos: la aplicación abre.
 *
 * Existe porque un frontend que no arranca pasa todas las pruebas de unidad y no sirve para nada:
 * esto mira lo que ve una persona en el navegador, incluidos **los errores de la consola**, que son
 * los que delatan un fallo de arranque que no se ve en pantalla.
 */
test.describe('La aplicación abre', () => {
  test('la raíz lleva a la pantalla de entrada y se ve el formulario', async ({ page }) => {
    const errores: string[] = [];
    page.on('console', (mensaje) => {
      if (mensaje.type() === 'error') {
        errores.push(mensaje.text());
      }
    });
    page.on('pageerror', (error) => errores.push(error.message));

    await page.goto('/');

    // Sin sesión, la raíz tiene que terminar en la pantalla de entrada.
    await expect(page).toHaveURL(/\/login$/);

    // Y la pantalla tiene que estar pintada, no en blanco: título, los dos campos y el botón.
    await expect(page.getByRole('heading', { name: 'Entrar' })).toBeVisible();
    await expect(page.getByLabel('Correo electrónico')).toBeVisible();
    await expect(page.getByLabel('Contraseña')).toBeVisible();
    // El nombre va exacto: la pantalla de entrada tiene «Entrar» y, si la instalación lo tiene,
    // «Entrar con Keycloak», y sin `exact` el localizador encuentra los dos.
    await expect(page.getByRole('button', { name: 'Entrar', exact: true })).toBeVisible();

    expect(errores, `La consola del navegador tiene errores:\n${errores.join('\n')}`).toEqual([]);
  });

  test('ninguna pantalla repite un id', async ({ page }) => {
    // Un `id` repetido deja la etiqueta apuntando al elemento equivocado y el campo sin nombre
    // accesible: se ve en el navegador y no en las pruebas de unidad (pasó con `app-campo`).
    for (const ruta of ['/login', '/forgot-password', '/set-password', '/forbidden']) {
      await page.goto(ruta);
      await page.waitForLoadState('networkidle');

      const repetidos = await page.evaluate(() => {
        const cuenta = new Map<string, number>();
        for (const elemento of document.querySelectorAll('[id]')) {
          const id = elemento.id;
          cuenta.set(id, (cuenta.get(id) ?? 0) + 1);
        }
        return [...cuenta.entries()].filter(([, cuantos]) => cuantos > 1).map(([id]) => id);
      });

      expect(repetidos, `En ${ruta} hay ids repetidos: ${repetidos.join(', ')}`).toEqual([]);
    }
  });

  test('los campos tienen nombre accesible y se pueden usar con el teclado', async ({ page }) => {
    await page.goto('/login');

    // Es lo que hace que un lector de pantalla diga qué se está escribiendo.
    await expect(page.getByLabel('Correo electrónico')).toBeVisible();
    await expect(page.getByLabel('Contraseña')).toBeVisible();

    // Y se llega a ellos con el tabulador, sin tocar el ratón.
    await page.keyboard.press('Tab');
    await expect(page.getByLabel('Correo electrónico')).toBeFocused();
  });

  test('el CSS llega y se aplica de verdad', async ({ page }) => {
    // Esta es la comprobación que faltaba: la página puede tener todo el texto, los campos y los
    // botones en su sitio y aun así **salir sin estilos**, porque el CSS no llegue. Pasó: el
    // servidor de desarrollo llevaba un día levantado y servía la hoja de Tailwind sin ninguna
    // utilidad generada. Las pruebas de estructura no lo ven; mirar el color, sí.
    const fallos: string[] = [];
    page.on('requestfailed', (peticion) => fallos.push(peticion.url()));

    await page.goto('/login');

    const aspecto = await page.evaluate(() => {
      const boton = document.querySelector('button[type="submit"]');
      const campo = document.querySelector('input[type="email"]');
      const fondo = getComputedStyle(document.body).backgroundColor;

      return {
        colorBoton: boton ? getComputedStyle(boton).backgroundColor : 'sin botón',
        bordeCampo: campo ? getComputedStyle(campo).borderTopWidth : 'sin campo',
        fondoPagina: fondo,
        hojas: document.styleSheets.length,
        // Si Tailwind no generó utilidades, ninguna regla menciona una clase de las que usamos.
        tieneUtilidades: [...document.styleSheets].some((hoja) => {
          try {
            return [...hoja.cssRules].some((regla) => regla.cssText.includes('.min-h-dvh'));
          } catch {
            return false;
          }
        }),
      };
    });

    expect(fallos, `Hay peticiones que fallaron: ${fallos.join(', ')}`).toEqual([]);
    expect(aspecto.hojas, 'No se cargó ninguna hoja de estilos').toBeGreaterThan(0);
    expect(aspecto.tieneUtilidades, 'Tailwind no generó ninguna utilidad: la hoja llega sin clases').toBe(true);

    // Y se nota en lo que se ve: el botón tiene color y el campo tiene borde.
    expect(aspecto.colorBoton).not.toBe('rgba(0, 0, 0, 0)');
    expect(aspecto.colorBoton).not.toBe('transparent');
    expect(aspecto.bordeCampo).not.toBe('0px');
    expect(aspecto.fondoPagina).not.toBe('rgba(0, 0, 0, 0)');
  });

  test('el idioma de arranque es el global aunque cambie el navegador', async ({ browser }) => {
    for (const locale of ['es-CO', 'en-GB', 'fr-FR']) {
      const context = await browser.newContext({ locale });
      const pagina = await context.newPage();
      await pagina.goto('/login');
      await expect(pagina.getByRole('heading', { name: 'Entrar' })).toBeVisible();
      await context.close();
    }
  });

  test('la entrada no ofrece un selector de idioma personal', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: 'Entrar' })).toBeVisible();
    await expect(page.getByLabel('Cambiar idioma')).toHaveCount(0);
  });

  test('enseña la pantalla de olvido y la de contraseña', async ({ page }) => {
    await page.goto('/forgot-password');
    await expect(page.getByRole('heading', { name: 'Recuperar la contraseña' })).toBeVisible();

    await page.goto('/set-password');
    await expect(page.getByRole('heading', { name: 'Establece tu contraseña' })).toBeVisible();
  });

  test('las pantallas de dentro piden entrar: sin sesión no se llega', async ({ page }) => {
    // «Sin permiso» es una pantalla de dentro: sin sesión no se sabe el papel de nadie, así que lo
    // que toca es pedir que se entre. Con sesión y sin permiso lo prueba `armazon.spec.ts`.
    for (const ruta of ['/forbidden', '/settings', '/change-password']) {
      await page.goto(ruta);
      await expect(page, `Desde ${ruta} debería pedir entrar`).toHaveURL(/\/login$/);
    }
  });
});
