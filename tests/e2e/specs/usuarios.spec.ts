import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

import {
  ASUNTO_DEL_ALTA,
  CONTRASENA_DE_PRUEBA,
  abrirMenuSiEsCajon,
  crearCuentaLista,
  entrar,
  entrarComo,
  esperarEnlaceDelCorreo,
  FABRICA,
  salir,
  tokenDe,
  tokenDeFabrica,
} from '../ayudas';

/**
 * Las pantallas de usuarios: la lista, la ficha y el perfil
 * (`docs/interfaz-y-experiencia.md`, sección 3.6).
 *
 * Son recorridos completos: una cuenta se da de alta de verdad, el correo se lee del buzón de pruebas
 * y la persona entra con la contraseña que establece desde su enlace. Nada se da por hecho por
 * detrás, que es lo que hace que esta capa encuentre lo que las pruebas de unidad no ven.
 */

/**
 * La fila de una cuenta: en PC es una fila de la tabla y en móvil una tarjeta.
 *
 * **Se busca por el correo y no por la posición**: la lista se pagina y se filtra, así que el número
 * de fila no dice nada.
 */
function filaDe(page: Page, email: string) {
  return page.locator('li, tr').filter({ hasText: email }).first();
}

test.describe('La lista de usuarios', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede entrar');

    await entrar(page, FABRICA.email, FABRICA.password);
    // Un Administrador entra directamente en la lista (sección 4.4).
    await expect(page).toHaveURL(/\/users$/);
  });

  test('la lista enseña las cuentas, con sus filtros a la vista y sin esconder nada', async ({
    page,
    request,
  }) => {
    const ana = await crearCuentaLista(request);

    await page.goto('/users');
    await expect(page.getByRole('heading', { name: 'Usuarios' })).toBeVisible();

    // Los filtros son chips a la vista, no un formulario escondido detrás de un botón.
    await expect(page.getByRole('group', { name: 'Papel' })).toBeVisible();
    await expect(page.getByRole('group', { name: 'Origen' })).toBeVisible();
    await expect(page.getByRole('group', { name: 'Estado' })).toBeVisible();

    // La búsqueda acota por correo, y la cuenta está, con su papel, su origen y su estado escritos:
    // nunca sólo un color.
    await page.getByLabel('Buscar por nombre o correo').fill(ana.email);

    const fila = filaDe(page, ana.email);
    await expect(fila).toBeVisible();
    await expect(fila).toContainText('Usuario');
    await expect(fila).toContainText('Local');
    await expect(fila).toContainText('Activa');

    // Y una búsqueda que no encuentra nada lo dice, con la salida a mano.
    await page.getByLabel('Buscar por nombre o correo').fill('no-existe-esta-cuenta');
    await expect(page.getByText('No hay ninguna cuenta que coincida con esos filtros.')).toBeVisible();

    // Quitar los filtros los deja como estaban: los chips vuelven a «Todos» y la búsqueda se vacía.
    // **No se comprueba que la cuenta se vea**: en desarrollo hay cuentas de muchas pasadas —no existe
    // dar de baja— y una recién creada puede caer en la página siguiente. Lo que se prueba aquí es el
    // filtro, así que se busca otra vez.
    await page.getByRole('button', { name: 'Quitar los filtros' }).first().click();
    await expect(
      page.getByRole('group', { name: 'Estado' }).getByRole('button', { name: 'Todos' }),
    ).toHaveAttribute('aria-pressed', 'true');
    await expect(page.getByLabel('Buscar por nombre o correo')).toHaveValue('');

    await page.getByLabel('Buscar por nombre o correo').fill(ana.email);
    await expect(filaDe(page, ana.email)).toBeVisible();

    // El filtro de estado la esconde cuando se piden las desactivadas, que es lo que hace útil el
    // filtro: mirar sólo lo que está apagado.
    await page.getByRole('group', { name: 'Estado' }).getByRole('button', { name: 'Desactivadas' }).click();
    await expect(filaDe(page, ana.email)).toHaveCount(0);

    await page.getByRole('group', { name: 'Estado' }).getByRole('button', { name: 'Activas' }).click();
    await expect(filaDe(page, ana.email)).toBeVisible();
  });

  test('el alta desde la pantalla crea la cuenta, manda el enlace y la persona entra', async ({
    page,
    request,
  }) => {
    await page.getByRole('button', { name: 'Nueva cuenta' }).click();

    const alta = page.getByRole('dialog');
    await expect(alta).toBeVisible();

    const marca = Date.now().toString(36);
    const correo = `e2e-alta-${marca}@demo.com`;

    await alta.getByLabel('Nombre').fill('Marta');
    await alta.getByLabel('Apellidos').fill('Ruiz');
    await alta.getByLabel('Correo electrónico').fill(correo);
    // El papel se elige en el alta: es lo que reparte un Administrador.
    await alta.getByLabel('Papel').selectOption('soporte');
    await alta.getByRole('button', { name: 'Crear la cuenta' }).click();

    // El diálogo se cierra y se dice lo que ha pasado.
    await expect(alta).toBeHidden();
    await expect(page.getByText('Cuenta creada. Se le ha mandado el enlace')).toBeVisible();

    // La cuenta está en la lista: se busca por su correo, que es lo que haría quien la acaba de dar
    // de alta (y así la prueba no depende de en qué página haya caído por el orden de los apellidos).
    await page.getByLabel('Buscar por nombre o correo').fill(correo);
    await expect(filaDe(page, correo)).toContainText('Soporte técnico');

    // El enlace llegó de verdad al buzón de pruebas: no se da por hecho.
    const enlace = await esperarEnlaceDelCorreo(request, correo, ASUNTO_DEL_ALTA);

    // Y con él, la persona nueva entra: es el recorrido entero, del alta a la primera entrada.
    await request.post('/api/auth/password/reset', {
      data: { token: enlace.split('#token=')[1], password: CONTRASENA_DE_PRUEBA },
    });

    await entrarComo(page, correo);
    // Entra como Soporte, así que va a su bandeja, que es donde trabaja.
    await expect(page).toHaveURL(/\/tickets$/);
  });

  test('desactivar pregunta, deja a la persona fuera y reactivar la devuelve', async ({
    page,
    request,
  }) => {
    const ana = await crearCuentaLista(request);
    const token = await tokenDeFabrica(request);

    // Se busca por su correo antes de mirar la fila: en desarrollo hay cuentas de muchas pasadas —no
    // existe dar de baja— y una recién creada puede caer en la página siguiente.
    await page.goto('/users');
    await page.getByLabel('Buscar por nombre o correo').fill(ana.email);
    await filaDe(page, ana.email).getByRole('link', { name: 'Abrir la ficha' }).click();

    // La ficha es de la cuenta que se ha pedido, y va sólo un Administrador.
    await expect(page).toHaveURL(new RegExp(`/users/${ana.id}$`));
    await expect(page.getByRole('heading', { name: 'Prueba Automática' })).toBeVisible();

    // Desactivar **no se hace de un clic**: primero dice lo que va a pasar.
    await page.getByRole('button', { name: 'Desactivar' }).click();
    await expect(page.getByText('¿Desactivar esta cuenta?')).toBeVisible();
    await page.getByRole('button', { name: 'Sí, desactivar' }).click();

    await expect(page.getByText('Cuenta desactivada. Ya no puede entrar.')).toBeVisible();
    await expect(page.getByText('Desactivada').first()).toBeVisible();

    // Y de verdad está fuera: la cuenta existe y la contraseña es la buena, y aun así no entra.
    await salir(page);
    await entrar(page, ana.email, ana.password);
    await expect(page.getByText('Tu cuenta está desactivada')).toBeVisible();

    // Reactivarla la devuelve: se hace por la API, porque la pantalla ya se ha probado.
    const reactivar = await request.post(`/api/users/${ana.id}/activate`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(reactivar.status(), await reactivar.text()).toBe(200);

    await entrar(page, ana.email, ana.password);
    await expect(page).toHaveURL(/\/tickets$/);
  });
});

/**
 * Lo que cambia de un papel a otro: lo que ve cada uno y lo que la pantalla **no** le ofrece.
 */
test.describe('Los usuarios, por papel', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede entrar');
  });

  test('Soporte edita el nombre en la lista, y no tiene ficha de nadie', async ({
    page,
    request,
  }) => {
    const soporte = await crearCuentaLista(request, { role: 'soporte' });
    const ana = await crearCuentaLista(request);

    await entrarComo(page, soporte.email);

    // Soporte entra en su bandeja, y la lista de usuarios está en el menú.
    await expect(page).toHaveURL(/\/tickets$/);
    await page.goto('/users');

    // Se busca por su correo: la lista se pagina, y la cuenta recién creada puede no estar en la
    // primera página cuando en desarrollo se acumulan las de todas las pasadas.
    await page.getByLabel('Buscar por nombre o correo').fill(ana.email);

    const fila = filaDe(page, ana.email);
    await expect(fila).toBeVisible();

    // No hay ficha que abrir: Soporte no ve la ficha de nadie (users.md, decisión 2).
    await expect(fila.getByRole('link', { name: 'Abrir la ficha' })).toHaveCount(0);

    // Y edita el nombre donde está, que es el único sitio donde ese permiso se puede ejercer.
    await fila.getByRole('button', { name: 'Editar' }).click();
    await fila.getByLabel('Nombre').fill('Ana María');
    await fila.getByLabel('Apellidos').fill('Pérez Gómez');
    await fila.getByRole('button', { name: 'Guardar' }).click();

    await expect(page.getByText('Cambios guardados.')).toBeVisible();

    // **La fila se relee**, y con margen: en una pasada completa (con el motor de IA y la base
    // trabajando) el repintado de la lista tardó más de lo normal y esta comprobación salió roja una vez,
    // pasando en solitario y en la pasada anterior. Se le da tiempo en vez de darlo por fallo.
    await expect(filaDe(page, ana.email)).toContainText('Ana María Pérez Gómez', { timeout: 20_000 });

    // Se recarga: el nombre está guardado de verdad, no sólo pintado. Se busca por correo, porque al
    // cambiarle los apellidos la cuenta **cambia de sitio en la lista** —el orden es por apellidos y
    // nombre— y puede caer en otra página.
    await page.reload();
    await page.getByLabel('Buscar por nombre o correo').fill(ana.email);
    await expect(filaDe(page, ana.email)).toContainText('Ana María Pérez Gómez');

    // Y si escribe la dirección de la ficha a mano, se le dice que no.
    await page.goto(`/users/${ana.id}`);
    await expect(page).toHaveURL(/\/forbidden$/);
  });

  test('el perfil propio cambia el nombre y el idioma, y no deja tocar el correo', async ({
    page,
    request,
  }) => {
    const ana = await crearCuentaLista(request);

    await entrarComo(page, ana.email);
    // Un usuario entra en sus tickets.
    await expect(page).toHaveURL(/\/tickets$/);

    // El perfil se abre desde tu nombre, en la zona de controles del menú.
    await abrirMenuSiEsCajon(page);
    await page.getByRole('link', { name: 'Mi perfil' }).click();
    await expect(page).toHaveURL(/\/profile$/);
    await expect(page.getByRole('heading', { name: 'Mi perfil' })).toBeVisible();

    // El correo y el papel se ven, pero no hay campo: son decisiones de un Administrador.
    await expect(page.getByText(ana.email)).toBeVisible();
    await expect(page.getByLabel('Correo electrónico')).toHaveCount(0);

    await page.getByLabel('Nombre').fill('Ana María');
    await page.getByRole('button', { name: 'Guardar' }).click();
    await expect(page.getByText('Tu perfil está guardado.')).toBeVisible();

    // El menú enseña el nombre nuevo sin volver a entrar.
    await expect(page.locator('aside').getByText('Ana María')).toBeVisible();

    // Y el idioma de la cuenta es el que manda: al cambiarlo, la interfaz cambia con él, porque es
    // también el idioma de sus correos (docs/modules/users.md, sección 8).
    await page.getByLabel('Idioma de los correos').selectOption('en');
    await page.getByRole('button', { name: 'Guardar' }).click();

    await expect(page.getByRole('heading', { name: 'My profile' })).toBeVisible();
    await expect(page.locator('aside').getByText('My profile')).toBeVisible();
  });

  test('nadie se desactiva a sí mismo, y la cuenta de fábrica no tiene perfil', async ({
    page,
    request,
  }) => {
    const jefa = await crearCuentaLista(request, { role: 'administrador' });

    await entrarComo(page, jefa.email);
    await expect(page).toHaveURL(/\/users$/);

    // Su propia ficha: se abre, y **no** ofrece desactivarse (users.md, sección 7).
    await page.goto(`/users/${jefa.id}`);
    await expect(page.getByRole('heading', { name: 'Prueba Automática' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Desactivar' })).toHaveCount(0);
    // Lo que sí puede: **cambiarse la contraseña a sí misma** —el botón se llama por lo que hace desde
    // el 2026-09-29—, que no deja a nadie fuera. Y sale porque su cuenta es **local**: en una de
    // directorio no se enseña (la contraseña la comprueba el directorio).
    await expect(page.getByRole('button', { name: 'Cambiar contraseña' })).toBeVisible();

    // La cuenta de fábrica no está en la tabla: su perfil no existe, y la pantalla no se le enseña.
    await salir(page);
    await entrar(page, FABRICA.email, FABRICA.password);
    await page.goto('/profile');
    await expect(page).toHaveURL(/\/forbidden$/);

    // Y en el menú su nombre no lleva a ninguna parte: no se promete un perfil que no hay.
    await abrirMenuSiEsCajon(page);
    await expect(page.locator('aside').getByRole('link', { name: 'Mi perfil' })).toHaveCount(0);
  });

  test('mandar el enlace pregunta antes, y la confirmación sale centrada', async ({
    page,
    request,
  }, info) => {
    // **La cuenta que se mira tiene contraseña**, así que el enlace es el de **recuperación**: la que
    // tenga dejará de valer. Es lo que hay que decir antes de mandarlo (decisión del responsable,
    // 2026-09-28), y por eso se pregunta.
    //
    // La ficha de otra cuenta la ve **el Administrador** (el reparto de la sección 8 de
    // `docs/usuarios-y-permisos.md`), así que aquí se entra con la cuenta de fábrica.
    const cuenta = await crearCuentaLista(request, { name: 'Usuario', lastName: 'Uno' });

    await entrar(page, FABRICA.email, FABRICA.password);
    await page.goto(`/users/${cuenta.id}`);
    await expect(page.getByRole("heading", { level: 1, name: "Usuario Uno" })).toBeVisible();

    // El de la ficha es el primero del documento; el del modal sale después, dentro de la ventana.
    const elDeLaFicha = page.getByRole('button', { name: 'Cambiar contraseña' }).first();

    await elDeLaFicha.click();

    const ventana = page.getByRole('dialog');
    await expect(ventana).toBeVisible();
    await expect(ventana).toContainText('contraseña nueva');

    // **Y la ventana sale centrada**: el `<dialog>` nativo se centra con `margin: auto` de la hoja del
    // navegador, y la preparación de Tailwind pone `margin: 0` a todo, así que salía en la esquina
    // —medido: `x: 0, y: 0`—. Se mide aquí, que es donde el responsable lo vio.
    if (info.project.name === 'pc') {
      const caja = (await ventana.boundingBox())!;
      const vista = page.viewportSize()!;
      const centroDeLaVentana = caja.x + caja.width / 2;
      const centroDeLaPantalla = vista.width / 2;

      expect(
        Math.abs(centroDeLaVentana - centroDeLaPantalla),
        'la ventana tiene que salir centrada en la pantalla',
      ).toBeLessThan(8);
    }

    // Cancelar no manda nada.
    await ventana.getByRole('button', { name: 'Cancelar' }).click();
    await expect(page.getByText('Enlace mandado otra vez')).toHaveCount(0);

    // Y confirmar sí, con su aviso de que caduca.
    await elDeLaFicha.click();
    await ventana.getByRole('button', { name: 'Cambiar contraseña' }).click();
    await expect(page.getByText('Enlace mandado otra vez. Caduca en 24 horas.')).toBeVisible();
  });

});
