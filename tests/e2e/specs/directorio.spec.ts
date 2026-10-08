import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

import {
  ASUNTO_DEL_ALTA,
  CONTRASENA_DE_PRUEBA,
  FABRICA,
  entrar,
  entrarComo,
  esperarEnlaceDelCorreo,
  ponerElMetodo,
  respondeAlProbar,
  salir,
  tokenDeFabrica,
  tokenGuardado,
} from '../ayudas';

/**
 * El camino de AD, desde el navegador (`docs/modules/auth.md`, secciones 5.2 y 5.4).
 *
 * Lo que se prueba aquí es lo que pidió el responsable: **una persona del directorio entra
 * directamente, sin que nadie le dé de alta nada**. Nadie crea su cuenta, nadie le manda un enlace y
 * nadie le pone una contraseña: entra con la de la empresa y la cuenta aparece —o se pone al día—
 * sola.
 *
 * **Estas pruebas necesitan el directorio levantado**, que la instancia aislada levanta con
 * el perfil `directory` de `tests.yml`:
 *
 * ```bash
 * docker compose -f tests.yml --profile directory down -v
 * docker compose -f tests.yml --profile directory up -d --build database backend frontend mail ldap keycloak
 * docker compose -f tests.yml run --rm e2e
 * ```
 *
 * Sin él se saltan en vez de fallar: la suite entera tiene que poder correr sin el directorio, que es
 * lo que hace una instalación que no tenga AD (`docs/modules/auth.md`, decisión 28).
 *
 * **Las tres personas están en `config/ldap/01-personas.ldif`** y sus cuentas no se crean a mano:
 * este es el único sitio donde la contraseña de una prueba vive fuera de la base de datos, y es
 * porque vive en el directorio.
 */
const ANA = { email: 'ana.directorio@demo.com', password: 'una-contraseña-larga' };
const LUIS = { email: 'luis.directorio@demo.com', password: 'otra-contraseña-larga' };
const MARTA = { email: 'marta.directorio@demo.com', password: 'la-de-marta-larga' };

/**
 * ¿Responde el directorio? Se pregunta **por el botón de «Probar la conexión» de Configuración**, que
 * es lo que pregunta una persona antes de guardarlo: conecta con la cuenta de servicio y no entra con
 * nadie, así que la pregunta no da de alta a nadie ni gasta una contraseña.
 *
 * Es lo que permite saltarse el caso cuando el directorio no está —o está a medio arrancar— en vez de
 * fallar por algo que no es del producto, **y ya no depende del método de entrada**.
 */
async function respondeElDirectorio(request: APIRequestContext): Promise<boolean> {
  return respondeAlProbar(request, 'directory');
}

/** Las cuentas que la lista de usuarios enseña con ese correo, vistas por la API. */
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

test.describe('El camino de AD', () => {
  /**
   * **La instalación entra por un método a la vez**, así que este camino hay que ponerlo: con el método
   * en `local` —que es como queda el entorno de ejemplo— el directorio no se mira siquiera.
   *
   * Se pone una vez para toda la suite y **se deja en `local` al terminar**, pase lo que pase: el resto
   * de las pruebas entra con cuentas locales, y dejarlo en `ad` las dejaría sin puerta. Se devuelve con
   * la cuenta de fábrica, que entra siempre.
   */
  test.beforeAll(async ({ request }) => {
    test.skip(
      !(await respondeElDirectorio(request)),
      'LDAP no responde: incluye el perfil directory de tests.yml para probar AD',
    );
    await ponerElMetodo(request, 'ad');
  });

  test.afterAll(async ({ request }) => {
    await ponerElMetodo(request, 'local');
  });

  test.beforeEach(async ({ page, request }) => {
    // Y la pantalla de entrada lo dice: con el método en AD el formulario es el mismo, pero se cuenta
    // de dónde es la contraseña que se espera (docs/modules/settings.md, sección 5.8).
    await page.goto('/login');
    await expect(
      page.getByText('Esta instalación entra con las cuentas de la organización'),
    ).toBeVisible();
  });

  /**
   * El alta automática: una persona del directorio **sin cuenta ninguna** entra con su contraseña de
   * la empresa y llega a sus tickets. Nadie la ha dado de alta.
   */
  test('una persona del directorio entra sin que nadie le dé de alta nada', async ({
    page,
    request,
  }, info) => {
    // Se prueba una vez, y no en los dos anchos: lo que comprueba —que la cuenta nace sola— no
    // depende del tamaño de la pantalla, y los dos proyectos comparten la misma base de datos.
    test.skip(info.project.name !== 'pc', 'El alta automática no depende del ancho: se prueba en PC');

    const token = await tokenDeFabrica(request);
    const antes = await cuentasConEseCorreo(request, LUIS.email, token);

    // Si ya hay cuenta, es de una ejecución anterior y este caso no puede distinguir el alta del
    // vínculo: se dice qué hay que limpiar en vez de dar por bueno lo que no se ha probado.
    test.skip(
      antes.length > 0,
      `Queda la cuenta de ${LUIS.email} de otra ejecución: limpia los datos de desarrollo ` +
        '(docs/ambientes.md, sección 9.3) y vuelve a correr',
    );

    await entrar(page, LUIS.email, LUIS.password);

    // Dentro: la raíz reparte, y un `usuario` va a su bandeja (docs/interfaz-y-experiencia, 4.4).
    await expect(page).toHaveURL(/\/tickets$/);
    await expect(page.getByRole('heading', { name: 'Mis tickets' })).toBeVisible();

    // Y la cuenta existe, con el papel de quien entra por primera vez y **sin contraseña local**:
    // su contraseña es la del directorio (docs/modules/auth.md, sección 5.4).
    const despues = await cuentasConEseCorreo(request, LUIS.email, token);
    expect(despues).toHaveLength(1);
    expect(despues[0]!.origin).toBe('ad');
    expect(despues[0]!.role).toBe('usuario');
    expect(despues[0]!.isActive).toBe(true);
  });

  /**
   * Quien ya tiene cuenta de directorio vuelve a entrar, y **el directorio manda en sus datos**: si
   * allí cambia el apellido, aquí se entera solo (sección 5.4).
   */
  test('una cuenta que ya es del directorio vuelve a entrar', async ({ page, request }) => {
    await entrar(page, ANA.email, ANA.password);
    await expect(page).toHaveURL(/\/tickets$/);

    // El nombre que se enseña en el menú es el del directorio.
    await expect(page.getByText('Ana Pérez').first()).toBeVisible();
  });

  /** Con la contraseña equivocada, lo dice el directorio y no se entra. */
  test('la contraseña equivocada la rechaza el directorio', async ({ page, request }) => {
    await entrar(page, ANA.email, 'la-que-no-es');

    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByText('El directorio ha rechazado esas credenciales.')).toBeVisible();
    expect(await tokenGuardado(page)).toBeNull();
  });

  /**
   * **El vínculo**: una cuenta que nació local, con el correo de alguien que está en el directorio,
   * pasa a ser del directorio al entrar por su camino, y desde ese momento **su contraseña local ya
   * no sirve** (sección 5.4).
   *
   * La cuenta local la monta la propia prueba, por la API y leyendo el correo del buzón: es el mismo
   * camino que sigue un alta de verdad.
   */
  test('una cuenta local con el correo de alguien del directorio se vincula al entrar', async ({
    page,
    request,
  }, info) => {
    test.skip(info.project.name !== 'pc', 'El vínculo no depende del ancho: se prueba en PC');

    const token = await tokenDeFabrica(request);
    const antes = await cuentasConEseCorreo(request, MARTA.email, token);

    test.skip(
      antes.length > 0,
      `Queda la cuenta de ${MARTA.email} de otra ejecución: limpia los datos de desarrollo ` +
        '(docs/ambientes.md, sección 9.3) y vuelve a correr',
    );

    // Una cuenta local con ese correo, con su contraseña puesta desde el enlace del correo.
    const alta = await request.post('/api/users', {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        name: 'Marta',
        lastName: 'Local',
        email: MARTA.email,
        role: 'usuario',
        origin: 'local',
      },
    });
    expect(alta.status(), await alta.text()).toBe(201);

    const enlace = await esperarEnlaceDelCorreo(request, MARTA.email, ASUNTO_DEL_ALTA);
    const conPassword = await request.post('/api/auth/password/reset', {
      data: { token: enlace.split('#token=')[1], password: CONTRASENA_DE_PRUEBA },
    });
    expect(conPassword.status(), await conPassword.text()).toBe(200);

    // Entra con la contraseña del directorio, que no es la suya local.
    await entrar(page, MARTA.email, MARTA.password);
    await expect(page).toHaveURL(/\/tickets$/);

    const despues = await cuentasConEseCorreo(request, MARTA.email, token);
    expect(despues[0]!.origin).toBe('ad');

    // Y su contraseña local ya no sirve: la cuenta es del directorio y allí manda su contraseña.
    await salir(page);
    await entrar(page, MARTA.email, CONTRASENA_DE_PRUEBA);
    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByText('El directorio ha rechazado esas credenciales.')).toBeVisible();
  });

  /**
   * **Reactivar una cuenta de AD pregunta al directorio**, que es lo que el documento pedía desde el
   * principio: devolverle el acceso a quien ya no está allí sería devolverle un acceso que no puede
   * usar, porque su contraseña es de allí (sección 5.4 de `auth.md` y punto 4 de la sección 5 de
   * `users.md`).
   *
   * Se comprueba en los dos sitios: la ficha ofrece el botón —que es lo que hace el administrador— y
   * el resultado se lee en la cuenta, no en lo que la pantalla dice.
   */
  test('una cuenta de AD desactivada se reactiva preguntando al directorio', async ({
    page,
    request,
  }, info) => {
    test.skip(info.project.name !== 'pc', 'No depende del ancho: se prueba en PC');

    // La cuenta tiene que existir: entra por AD (o vuelve a entrar).
    await entrar(page, ANA.email, ANA.password);
    await expect(page).not.toHaveURL(/\/login$/);
    await salir(page);

    const token = await tokenDeFabrica(request);
    const cuentas = await cuentasConEseCorreo(request, ANA.email, token);
    expect(cuentas).toHaveLength(1);
    const id = cuentas[0]!.id;

    // Se desactiva, como haría un administrador.
    const baja = await request.post(`/api/users/${id}/deactivate`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(baja.status(), await baja.text()).toBe(200);

    // Y en su ficha **sí se ofrece reactivarla**: en AD se puede comprobar con el directorio.
    await entrarComo(page, FABRICA.email, FABRICA.password);
    await page.goto(`/users/${id}`);
    await expect(page.getByRole('button', { name: 'Reactivar' })).toBeVisible();

    await page.getByRole('button', { name: 'Reactivar' }).click();
    await expect(page.getByText('Cuenta reactivada. Puede entrar otra vez.')).toBeVisible();

    // Y la cuenta está activa de verdad, no sólo en la pantalla.
    const despues = await cuentasConEseCorreo(request, ANA.email, token);
    expect(despues[0]!.isActive).toBe(true);
  });
});
