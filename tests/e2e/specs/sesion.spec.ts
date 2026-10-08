import { expect, test } from '@playwright/test';

import { abrirMenuSiEsCajon, cuentaDePrueba, entrar, FABRICA, tokenGuardado } from '../ayudas';

/**
 * El primer caso de la tabla de `docs/ambientes.md`, sección 8.3: **entrar y salir**.
 *
 * Es un recorrido completo y no veinte comprobaciones de detalle: entra una persona de verdad en un
 * navegador de verdad, contra el entorno de desarrollo, y se comprueba lo que pasa por debajo
 * —la cabecera de la sesión, `GET /api/auth/me` y el token guardado—.
 */
test.describe('Entrar y salir', () => {
  test('salir pregunta antes, y cancelar deja la sesión puesta', async ({ page }) => {
    // **Salir se confirma** (decisión del responsable, 2026-09-29): se pierde la sesión y hay que volver
    // a entrar —y con el directorio, volver a pasar por allí—, así que el botón del menú pregunta.
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no hay cuenta con la que entrar');

    await entrar(page, FABRICA.email, FABRICA.password);
    await expect(page).toHaveURL(/\/users$/);

    await abrirMenuSiEsCajon(page);
    await page.getByRole('button', { name: 'Salir' }).click();

    const ventana = page.getByRole('dialog');
    await expect(ventana).toBeVisible();
    await expect(ventana).toContainText('Tendrás que volver a entrar');

    // Cancelar no sale: se sigue dentro, y la ventana se cierra.
    await ventana.getByRole('button', { name: 'Cancelar' }).click();
    await expect(ventana).toHaveCount(0);
    await expect(page).toHaveURL(/\/users$/);
    await expect(page.getByRole('heading', { name: 'Usuarios' })).toBeVisible();

    // Y confirmar sí: se vuelve a la entrada.
    await abrirMenuSiEsCajon(page);
    await page.getByRole('button', { name: 'Salir' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Salir' }).click();
    await expect(page).toHaveURL(/\/login$/);
  });

  test('la entrada está cuadrada: el logo, el formulario y los controles, del mismo ancho', async ({
    page,
  }) => {
    // **Petición del responsable, 2026-09-28**: el recuadro del logo no medía lo mismo que el formulario,
    // y el idioma y el tema medían cada uno lo suyo —104 y 140 px—. Aquí se mide, y no se mira: es la
    // única forma de que no se vuelva a torcer, porque una pantalla descentrada pasa cualquier prueba de
    // texto.
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    const ancho = async (selector: string) => {
      const caja = (await page.locator(selector).first().boundingBox())!;

      return Math.round(caja.width);
    };

    const logo = await ancho('app-logo > span');
    const tarjeta = await ancho('app-tarjeta');
    const controles = await ancho('app-controles > div');

    expect(logo, 'el logo tiene que medir lo mismo que el formulario').toBe(tarjeta);
    expect(controles, 'los controles también').toBe(tarjeta);

    // Sólo queda el tema como preferencia del navegador; el idioma es global.
    await expect(page.locator('#idioma')).toHaveCount(0);
    expect(await ancho('#tema')).toBeLessThanOrEqual(controles);
  });

  test('la cuenta de fábrica entra, se la reconoce y sale', async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no hay cuenta con la que entrar');

    // Al entrar, el frontend pregunta quién es con la cabecera de la sesión puesta.
    const quienSoy = page.waitForRequest(
      (peticion) =>
        peticion.url().endsWith('/api/auth/me') &&
        (peticion.headers()['authorization'] ?? '').startsWith('Bearer '),
    );

    await entrar(page, FABRICA.email, FABRICA.password);

    // Se entra de verdad: la raíz deja de ser la pantalla de entrada y aparece **lo que le toca al
    // papel**. Un Administrador entra en la lista de usuarios, que es donde trabaja
    // (docs/interfaz-y-experiencia.md, sección 4.4).
    await expect(page).toHaveURL(/\/users$/);
    await expect(page.getByRole('heading', { name: 'Usuarios' })).toBeVisible();

    await quienSoy;
    expect(await tokenGuardado(page)).toBeTruthy();

    // Con sesión, la pantalla de entrada ya no tiene nada que hacer: lleva a donde toca entrar.
    await page.goto('/login');
    await expect(page).toHaveURL(/\/users$/);

    // Y desde dentro se llega a cambiar la contraseña, que es de `auth` y no del módulo de usuarios.
    await page.goto('/change-password');
    await expect(page).toHaveURL(/\/change-password$/);
    await expect(page.getByRole('heading', { name: 'Cambiar mi contraseña' })).toBeVisible();

    // Se sale desde el menú, que es donde está el botón, y **salir se confirma** (2026-09-29).
    await page.goto('/');
    await abrirMenuSiEsCajon(page);
    await page.getByRole('button', { name: 'Salir' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Salir' }).click();
    await expect(page).toHaveURL(/\/login$/);
    expect(await tokenGuardado(page)).toBeNull();

    // Al volver a la raíz, sin sesión, la guarda manda otra vez a la entrada.
    await page.goto('/');
    await expect(page).toHaveURL(/\/login$/);
  });

  test('con la contraseña equivocada no se entra y se dice por qué', async ({ page }) => {
    await entrar(page, FABRICA.email, 'esta-no-es-la-contraseña');

    // El mensaje es el traducido, no la clave del backend.
    await expect(page.getByText('El correo o la contraseña no son correctos.')).toBeVisible();
    await expect(page).toHaveURL(/\/login$/);
    expect(await tokenGuardado(page)).toBeNull();
  });

  test('sin sesión no se llega a cambiar la contraseña', async ({ page }) => {
    await page.goto('/change-password');
    await expect(page).toHaveURL(/\/login$/);
  });

  test('el aviso de «he olvidado mi contraseña» no dice si la cuenta existe', async ({ page }) => {
    await page.goto('/forgot-password');
    await page.getByLabel('Correo electrónico').fill('quien-sea@demo.com');
    await page.getByRole('button', { name: 'Mandar el enlace' }).click();

    await expect(page.getByText('Si esa dirección tiene cuenta, te ha llegado un enlace.')).toBeVisible();
  });

  test('el saludo depende de la cuenta, no del navegador', async ({ page }) => {
    // Una cuenta cualquiera no entra si no existe: aquí lo que se comprueba es que el error es el
    // mismo que con una contraseña mal puesta, sin decir si la cuenta existe.
    const cuenta = cuentaDePrueba();
    await entrar(page, cuenta.email, 'una-contraseña-larga');

    await expect(page.getByText('El correo o la contraseña no son correctos.')).toBeVisible();
  });
});
