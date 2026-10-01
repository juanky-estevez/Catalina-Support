import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

import {
  ASUNTO_DEL_ALTA,
  CONTRASENA_DE_PRUEBA,
  FABRICA,
  entrarComo,
  esperarEnlaceDelCorreo,
  ponerElMetodo,
  respondeAlProbar,
  salir,
  tokenDeFabrica,
  tokenGuardado,
} from '../ayudas';

/**
 * El camino de Keycloak, desde el navegador (`docs/modules/auth.md`, secciones 5.3 y 5.4).
 *
 * Lo que se prueba es lo que se le pidió a este camino: **una persona con cuenta en Keycloak entra
 * directamente, sin que nadie le dé de alta nada** —es la misma corrección del responsable que en
 * AD— y **el navegador no ve el token en ninguna parte salvo en el fragmento**, que es lo que no
 * viaja al servidor.
 *
 * **Estas pruebas necesitan Keycloak levantado**, que en desarrollo vive en su propio archivo
 * (`keycloak.yml`) y entra en la red que crea `dev.yml`:
 *
 * ```bash
 * docker compose -f dev.yml up -d
 * docker compose -f keycloak.yml up -d
 * docker compose -f dev.yml run --rm e2e
 * ```
 *
 * Sin él se saltan en vez de fallar: la suite entera tiene que poder correr en una instalación que no
 * tenga Keycloak (`docs/modules/auth.md`, decisión 28).
 *
 * Las tres personas están en `config/keycloak/realm-catalina-support.json`, que es el reino de
 * pruebas del repositorio: aquí la contraseña de una prueba vive en el reino, y no en la base de
 * datos, porque es de allí de donde sale.
 */
const SARA = { email: 'sara.keycloak@demo.com', password: 'la-de-keycloak-larga' };
const NURIA = { email: 'nuria.keycloak@demo.com', password: 'la-de-nuria-larga' };
const PABLO = { email: 'pablo.keycloak@demo.com', password: 'la-de-pablo-larga' };

/**
 * ¿Responde Keycloak? Se pregunta **por el botón de «Probar la conexión» de Configuración**, que es lo
 * que pregunta una persona: lee el documento del reino y no entra con nadie.
 *
 * Así la pregunta **no depende del método que esté puesto**, que es lo que hace falta para poder
 * saltarse estas pruebas cuando Keycloak no está levantado.
 */
async function hayKeycloak(request: APIRequestContext): Promise<boolean> {
  return respondeAlProbar(request, 'keycloak');
}

/** Las cuentas que la lista enseña con ese correo, vistas por la API de administración. */
async function cuentasConEseCorreo(
  request: APIRequestContext,
  email: string,
  token: string,
): Promise<{ id: number; origin: string; role: string; isActive: boolean }[]> {
  const respuesta = await request.get(`/api/users?q=${encodeURIComponent(email)}&perPage=100`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const cuerpo = (await respuesta.json()) as {
    users: { id: number; email: string; origin: string; role: string; isActive: boolean }[];
  };

  return cuerpo.users.filter((cuenta) => cuenta.email === email);
}

/**
 * Entra por Keycloak como lo haría una persona: desde la pantalla de entrada, por su botón.
 *
 * **Keycloak puede no preguntar nada**, y no es un fallo: si esa persona ya tiene sesión allí —que es
 * lo normal cuando acaba de entrar en la misma prueba—, Keycloak vuelve sola con el código y la
 * aplicación la deja dentro sin enseñar su formulario. Eso es el SSO de Keycloak, y **no hay cierre de
 * sesión en Keycloak** (`docs/modules/auth.md`, sección 5.3), así que la prueba tiene que aceptar las
 * dos cosas: el formulario, o estar ya dentro.
 */
async function entrarPorKeycloak(page: Page, persona: { email: string; password: string }): Promise<void> {
  await page.goto('/login');

  const boton = page.getByRole('button', { name: 'Entrar con Keycloak' });
  await expect(boton).toBeVisible();
  await boton.click();

  // O el formulario de Keycloak —que es suyo y está en su dirección—, o la aplicación, ya dentro.
  const formulario = page.locator('#username');
  const dentro = page.getByRole('heading', { name: 'Mis tickets' });
  await expect(formulario.or(dentro).first()).toBeVisible();

  if (!(await formulario.isVisible())) {
    return;
  }

  await expect(page).toHaveURL(/\/sso\/realms\/catalina-support\/protocol\/openid-connect\/auth/);

  await formulario.fill(persona.email);
  await page.locator('#password').fill(persona.password);
  await page.locator('#kc-login').click();
}

test.describe('El camino de Keycloak', () => {
  /**
   * **La instalación entra por un método a la vez**: con el método en `local` —que es como queda el
   * entorno de ejemplo— el botón de Keycloak no se enseña siquiera. Así que este camino se pone, y se
   * devuelve el método a `local` al terminar, pase lo que pase.
   */
  test.beforeAll(async ({ request }) => {
    test.skip(
      !(await hayKeycloak(request)),
      'Keycloak no responde: levántalo con `docker compose -f keycloak.yml up -d`',
    );

    await ponerElMetodo(request, 'keycloak');
  });

  test.afterAll(async ({ request }) => {
    await ponerElMetodo(request, 'local');
  });

  /**
   * Con el método en Keycloak **no hay formulario**: se entra desde su pantalla, y por eso lo único
   * que la pantalla de entrada ofrece es el botón. Es la otra mitad de «un método a la vez».
   */
  test('la pantalla de entrada no tiene formulario cuando el método es Keycloak', async ({ page }) => {
    await page.goto('/login');

    await expect(
      page.getByText('Esta instalación entra con Keycloak: se entra desde su pantalla, no desde aquí.'),
    ).toBeVisible();
    await expect(page.getByLabel('Correo electrónico')).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Entrar', exact: true })).toHaveCount(0);
    // Y no se ofrece «he olvidado mi contraseña»: la contraseña no es de aquí.
    await expect(page.getByRole('link', { name: 'He olvidado mi contraseña' })).toHaveCount(0);
  });

  /**
   * El botón está **sólo si la instalación tiene ese camino**, y lleva de verdad a Keycloak: es lo
   * que dice la decisión 27, y lo que impide ofrecer un botón que lleva a un error.
   */
  test('la pantalla de entrada ofrece Keycloak y lleva allí', async ({ page }) => {
    await page.goto('/login');

    await expect(page.getByRole('button', { name: 'Entrar con Keycloak' })).toBeVisible();

    await page.getByRole('button', { name: 'Entrar con Keycloak' }).click();
    await expect(page).toHaveURL(/\/sso\/realms\/catalina-support\/protocol\/openid-connect\/auth/);

    // Y allí está su formulario, que es suyo y no nuestro.
    await expect(page.locator('#username')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
  });

  /**
   * El alta automática, igual que en AD: alguien del directorio de Keycloak entra con su cuenta de la
   * organización y **su cuenta nace sola**, con papel `usuario` y sin contraseña nuestra.
   */
  test('una persona de Keycloak entra sin que nadie le dé de alta nada', async ({ page, request }, info) => {
    // Se prueba una vez: el camino no depende del ancho, y los dos proyectos comparten la base.
    test.skip(info.project.name !== 'pc', 'El alta automática no depende del ancho: se prueba en PC');

    const token = await tokenDeFabrica(request);
    const antes = await cuentasConEseCorreo(request, SARA.email, token);
    test.skip(
      antes.length > 0,
      `Queda la cuenta de ${SARA.email} de otra ejecución: limpia los datos de desarrollo ` +
        '(docs/ambientes.md, sección 9.3) y vuelve a correr',
    );

    await entrarPorKeycloak(page, SARA);

    // Vuelve a la aplicación, ya dentro: un `usuario` va a sus tickets.
    await expect(page).toHaveURL(/\/tickets$/);
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();

    // **El fragmento se borra en cuanto se usa**: el token no se queda en la barra de direcciones.
    expect(page.url()).not.toContain('#token=');

    // Y quedó la sesión guardada, que es lo que hace que lo anterior sea entrar de verdad.
    expect(await tokenGuardado(page)).toBeTruthy();

    // La cuenta existe con lo que dijo Keycloak: origen `keycloak` y el papel de quien entra solo.
    const despues = await cuentasConEseCorreo(request, SARA.email, token);
    expect(despues).toHaveLength(1);
    expect(despues[0]!.origin).toBe('keycloak');
    expect(despues[0]!.role).toBe('usuario');
    expect(despues[0]!.isActive).toBe(true);

    // El nombre del menú es el del directorio, que es quien manda en los datos.
    await expect(page.getByText('Sara Núñez').first()).toBeVisible();

    await salir(page);
  });

  /** Quien ya es de Keycloak vuelve a entrar, y sin dejar cuentas de más por el camino. */
  test('una cuenta que ya es de Keycloak vuelve a entrar', async ({ page, request }, info) => {
    test.skip(info.project.name !== 'pc', 'No depende del ancho: se prueba en PC');

    await entrarPorKeycloak(page, PABLO);

    await expect(page).toHaveURL(/\/tickets$/);
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();

    const token = await tokenDeFabrica(request);
    const cuentas = await cuentasConEseCorreo(request, PABLO.email, token);
    expect(cuentas).toHaveLength(1);
    expect(cuentas[0]!.origin).toBe('keycloak');

    await salir(page);
  });

  /**
   * **El vínculo**, igual que en AD: una cuenta que nació local, con el correo de alguien que está en
   * Keycloak, pasa a ser de Keycloak al entrar por su camino, y **su contraseña local ya no sirve**
   * (sección 5.4).
   */
  test('una cuenta local con el correo de alguien de Keycloak se vincula al entrar', async ({
    page,
    request,
  }, info) => {
    test.skip(info.project.name !== 'pc', 'El vínculo no depende del ancho: se prueba en PC');

    const token = await tokenDeFabrica(request);
    const antes = await cuentasConEseCorreo(request, NURIA.email, token);
    test.skip(
      antes.length > 0,
      `Queda la cuenta de ${NURIA.email} de otra ejecución: limpia los datos de desarrollo ` +
        '(docs/ambientes.md, sección 9.3) y vuelve a correr',
    );

    // Una cuenta local con ese correo, con su contraseña puesta desde el enlace del correo.
    const alta = await request.post('/api/users', {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        name: 'Nuria',
        lastName: 'Local',
        email: NURIA.email,
        role: 'usuario',
        origin: 'local',
        language: 'es',
      },
    });
    expect(alta.status(), await alta.text()).toBe(201);

    const enlace = await esperarEnlaceDelCorreo(request, NURIA.email, ASUNTO_DEL_ALTA);
    const conPassword = await request.post('/api/auth/password/reset', {
      data: { token: enlace.split('#token=')[1], password: CONTRASENA_DE_PRUEBA },
    });
    expect(conPassword.status(), await conPassword.text()).toBe(200);

    // Entra por Keycloak, con la contraseña de allí, que no es la suya local.
    await entrarPorKeycloak(page, NURIA);
    await expect(page).toHaveURL(/\/tickets$/);

    const despues = await cuentasConEseCorreo(request, NURIA.email, token);
    expect(despues[0]!.origin).toBe('keycloak');

    // Y su contraseña local ya no sirve: la cuenta es de Keycloak. Con el método en Keycloak no hay
    // formulario que rellenar —se entra desde su pantalla—, así que se comprueba por la API, que es
    // donde está la regla.
    await salir(page);
    const intento = await request.post('/api/auth/login', {
      data: { email: NURIA.email, password: CONTRASENA_DE_PRUEBA },
    });
    expect(((await intento.json()) as { error?: string }).error).toBe('auth.invalidCredentials');
  });

  /**
   * Una vuelta que no vale se cuenta en la pantalla de entrada y **el fragmento se borra**: si
   * alguien abre a mano una dirección con `#error=…`, no se queda ahí para siempre.
   */
  test('una vuelta que no vale se cuenta y no deja el fragmento', async ({ page }) => {
    await page.goto('/login#error=auth.oidc.state');

    await expect(page.getByText('La vuelta del acceso no vale o ha caducado. Prueba a entrar otra vez.')).toBeVisible();
    expect(page.url()).not.toContain('#error=');
  });

  /**
   * **Una cuenta de Keycloak no se reactiva a mano, y se dice por qué.** Su camino no tiene forma de
   * preguntar «¿sigues conociendo a esta persona?» sin una cuenta de servicio en el reino, así que la
   * interfaz **no ofrece el botón** y explica lo que pasa: esa cuenta entra sola y al entrar se
   * reactiva (docs/modules/users.md, sección 5, punto 4).
   *
   * Lo que se comprueba no es el mensaje: es que el acceso **vuelve de verdad** al entrar, que es lo
   * que se le ha contado a quien mira la ficha.
   */
  test('una cuenta de Keycloak desactivada no se reactiva a mano, y vuelve al entrar', async ({
    page,
    request,
  }, info) => {
    test.skip(info.project.name !== 'pc', 'No depende del ancho: se prueba en PC');

    // La cuenta tiene que existir: entra por Keycloak (o vuelve a entrar).
    await entrarPorKeycloak(page, PABLO);
    await expect(page).toHaveURL(/\/tickets$/);
    await salir(page);

    const token = await tokenDeFabrica(request);
    const cuentas = await cuentasConEseCorreo(request, PABLO.email, token);
    expect(cuentas).toHaveLength(1);
    const id = cuentas[0]!.id;

    const baja = await request.post(`/api/users/${id}/deactivate`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(baja.status(), await baja.text()).toBe(200);

    // En su ficha no hay botón de reactivar, y se cuenta lo que pasa.
    await entrarComo(page, FABRICA.email, FABRICA.password);
    await page.goto(`/users/${id}`);
    await expect(page.getByRole('button', { name: 'Reactivar' })).toHaveCount(0);
    await expect(
      page.getByText(
        'Es una cuenta de Keycloak: se reactivará sola cuando esa persona entre por su camino, sin que nadie haga nada aquí.',
      ),
    ).toBeVisible();

    // Y por la API se rechaza con su motivo, por si alguien la pide a mano.
    const intento = await request.post(`/api/users/${id}/activate`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(intento.status(), await intento.text()).toBe(422);
    expect(((await intento.json()) as { error: string }).error).toBe('users.directory.activatesItself');

    // Lo que se le ha contado es verdad: al entrar por su camino, la cuenta vuelve sola. Hay que
    // salir antes, porque la sesión que hay ahora es la del administrador y con sesión la pantalla
    // de entrada no se enseña.
    await salir(page);
    await entrarPorKeycloak(page, PABLO);
    await expect(page).toHaveURL(/\/tickets$/);

    const despues = await cuentasConEseCorreo(request, PABLO.email, token);
    expect(despues[0]!.isActive).toBe(true);

    await salir(page);
  });
});
