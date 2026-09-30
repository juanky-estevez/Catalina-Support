import { expect, test } from '@playwright/test';

import {
  BUZON,
  cuentaDePrueba,
  entrar,
  esperarEnlaceDelCorreo,
  ASUNTO_DEL_ALTA,
  FABRICA,
  tokenGuardado,
  vaciarBuzon,
} from '../ayudas';

/**
 * El camino local entero, con el correo de verdad: **alta, enlace, contraseña y entrada**.
 *
 * Es el recorrido que no se puede dar por bueno con pruebas de unidad: el enlace tiene que salir del
 * buzón de pruebas, el navegador tiene que abrirlo con su token en el fragmento, y la contraseña
 * tiene que servir para entrar después.
 */
test.describe('El enlace del correo', () => {
  test('el alta manda un enlace que sirve para establecer la contraseña y entrar', async ({
    page,
    request,
  }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede dar de alta nada');

    await vaciarBuzon(request);

    // El alta se pide a la API porque la pantalla de cuentas todavía no existe (le toca al módulo
    // `users`): lo que aquí se prueba es el enlace y la pantalla, no el formulario de alta.
    const entrarComoAdmin = await request.post('/api/auth/login', {
      data: { email: FABRICA.email, password: FABRICA.password },
    });
    expect(entrarComoAdmin.ok(), 'No se pudo entrar como administrador').toBeTruthy();
    const { token } = (await entrarComoAdmin.json()) as { token: string };

    const cuenta = cuentaDePrueba();
    const alta = await request.post('/api/users', {
      headers: { Authorization: `Bearer ${token}` },
      data: { ...cuenta, role: 'usuario', origin: 'local', language: 'es' },
    });
    expect(alta.status(), await alta.text()).toBe(201);

    // El correo llega al buzón de pruebas, y de ahí sale el enlace: es lo que haría una persona.
    const enlace = await esperarEnlaceDelCorreo(
      request,
      cuenta.email,
      'Establece tu contraseña de Catalina Support',
    );
    expect(enlace).toContain('#token=');

    // La pantalla lee el token del fragmento, lo borra de la dirección y pide la contraseña.
    await page.goto(enlace);
    await expect(page).toHaveURL(/\/set-password$/);
    await expect(page.getByRole('heading', { name: 'Establece tu contraseña' })).toBeVisible();

    // El token no se queda escrito en la barra de direcciones.
    expect(page.url()).not.toContain('token');

    await page.getByLabel('Contraseña nueva').fill('una-contraseña-larga');
    await page.getByLabel('Repite la contraseña').fill('una-contraseña-larga');
    await page.getByRole('button', { name: 'Guardar la contraseña' }).click();

    await expect(page.getByText('Tu contraseña ya está puesta. Entra con ella.')).toBeVisible();

    // Y con esa contraseña se entra de verdad.
    await entrar(page, cuenta.email, 'una-contraseña-larga');
    await expect(page).toHaveURL(/\/tickets$/);
    // Un usuario entra en sus tickets, y esta persona acaba de entrar: no tiene ninguno todavía.
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();
    await expect(page.getByText('Todavía no hay ningún ticket.')).toBeVisible();
    expect(await tokenGuardado(page)).toBeTruthy();

    // Con el mismo enlace no se puede volver a establecer: un enlace, un uso.
    await page.goto(enlace);
    await expect(page.getByRole('heading', { name: 'Establece tu contraseña' })).toBeVisible();
    await page.getByLabel('Contraseña nueva').fill('otra-contraseña-larga');
    await page.getByLabel('Repite la contraseña').fill('otra-contraseña-larga');
    await page.getByRole('button', { name: 'Guardar la contraseña' }).click();
    await expect(page.getByText(/ya no vale/)).toBeVisible();
  });

  test('las dos contraseñas tienen que coincidir', async ({ page }) => {
    await page.goto('/set-password#token=un-token-cualquiera-de-cuarenta-y-tres-caracter');

    await page.getByLabel('Contraseña nueva').fill('una-contraseña-larga');
    await page.getByLabel('Repite la contraseña').fill('otra-distinta-larga');
    await page.getByRole('button', { name: 'Guardar la contraseña' }).click();

    await expect(page.getByText('Las dos contraseñas no son iguales.')).toBeVisible();
  });

  test('sin token en la dirección, la pantalla lo dice', async ({ page }) => {
    await page.goto('/set-password');

    await expect(
      page.getByText('Este enlace no trae la información necesaria. Pide otro desde la pantalla de entrada.'),
    ).toBeVisible();
  });

  test('el buzón de pruebas está accesible desde las pruebas', async ({ request }) => {
    // Sin esto, el resto de casos de este archivo no podrían leer ningún correo, y fallarían con un
    // mensaje que no explica nada.
    const respuesta = await request.get(`${BUZON}/api/v1/messages`);
    expect(respuesta.ok(), `El buzón no responde en ${BUZON}`).toBeTruthy();
  });

  /**
   * **La política de contraseñas y el enlace, en el mismo caso**, porque van juntos: la contraseña
   * tiene que llegar a **8 caracteres**, y **una que no llega no puede gastar el enlace**.
   *
   * Lo segundo se descubrió probando lo primero (2026-09-25): el enlace se gastaba antes de comprobar
   * la contraseña, así que quien escribía una corta —y leía «al menos 8 caracteres»— se quedaba sin
   * enlace y tenía que pedir otro. Un enlace de un solo uso que se gasta cuando la contraseña no vale
   * deja a la persona en un bucle.
   */
  test('una contraseña corta se rechaza y no gasta el enlace; una de ocho vale', async ({
    page,
    request,
  }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede dar de alta nada');

    await vaciarBuzon(request);

    const entrarComoAdmin = await request.post('/api/auth/login', {
      data: { email: FABRICA.email, password: FABRICA.password },
    });
    expect(entrarComoAdmin.ok(), 'No se pudo entrar como administrador').toBeTruthy();
    const { token } = (await entrarComoAdmin.json()) as { token: string };

    const cuenta = cuentaDePrueba();
    const alta = await request.post('/api/users', {
      headers: { Authorization: `Bearer ${token}` },
      data: { ...cuenta, role: 'usuario', origin: 'local', language: 'es' },
    });
    expect(alta.status(), await alta.text()).toBe(201);

    const enlace = await esperarEnlaceDelCorreo(request, cuenta.email, ASUNTO_DEL_ALTA);
    const delEnlace = enlace.split('#token=')[1];

    // Siete caracteres: se rechaza, y la pantalla lo dice con el mínimo que toca.
    await page.goto(`/set-password#token=${delEnlace}`);
    await page.getByLabel('Contraseña nueva').fill('1234567');
    await page.getByLabel('Repite la contraseña').fill('1234567');
    await page.getByRole('button', { name: 'Guardar la contraseña' }).click();

    await expect(page.getByText('La contraseña tiene que tener al menos 8 caracteres.')).toBeVisible();
    await expect(page).toHaveURL(/set-password/);

    // **Y el enlace sigue sirviendo**: se establece la contraseña con ocho caracteres, en la misma
    // pantalla y sin pedir otro correo. Es lo que antes no pasaba: el enlace se gastaba con el
    // intento corto y había que empezar de nuevo.
    await page.getByLabel('Contraseña nueva').fill('12345678');
    await page.getByLabel('Repite la contraseña').fill('12345678');
    await page.getByRole('button', { name: 'Guardar la contraseña' }).click();

    await expect(page.getByText('Tu contraseña ya está puesta. Entra con ella.')).toBeVisible();

    // Y esa contraseña entra de verdad.
    await entrar(page, cuenta.email, '12345678');
    await expect(page).toHaveURL(/\/tickets$/);
    expect(await tokenGuardado(page)).toBeTruthy();
  });
});
