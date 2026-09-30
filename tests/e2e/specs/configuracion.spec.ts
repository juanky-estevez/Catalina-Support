import { expect, test, type Page } from '@playwright/test';

import {
  entrar,
  entrarComo,
  FABRICA,
  metodoDeEntrada,
  ponerElMetodo,
  respondeAlProbar,
  salir,
  tokenDeFabrica,
} from '../ayudas';

/**
 * Deja el nombre de la instalación en el de fábrica.
 *
 * **Partir de un estado conocido es lo que hace repetible una prueba que cambia un ajuste de la
 * instalación**: si otra ejecución se quedó a medias con el nombre cambiado —o lo dejó cambiado—, el
 * campo ya tendría ese valor, el botón estaría apagado por no haber nada que guardar, y la prueba
 * fallaría por una razón que no tiene nada que ver con lo que prueba.
 */
async function partirDelNombreDeFabrica(page: Page): Promise<void> {
  // **El nombre va exacto**: la pantalla tiene además «Atributo del nombre» en la configuración del
  // directorio, y sin `exact` el localizador encuentra los dos.
  await page.getByLabel('Nombre', { exact: true }).fill('');

  const guardar = page.getByRole('button', { name: 'Guardar la instalación' });
  if (!(await guardar.isEnabled())) {
    return;
  }

  await guardar.click();
  await expect(page.getByText('La instalación se ha guardado.')).toBeVisible();
}

/**
 * La pantalla de Configuración: la marca y el color institucional.
 *
 * Se prueba **usándola**: se elige un archivo en el campo, se sube, se ve la vista previa, se vuelve
 * al de fábrica y se cambia el color con su vista previa. Dejar la instalación como estaba forma
 * parte de la prueba.
 */
test.describe('Configuración', () => {
  test.beforeEach(async ({ page }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede entrar');

    await entrar(page, FABRICA.email, FABRICA.password);
    await page.goto('/settings');
    await expect(page.getByRole('heading', { name: 'Configuración' })).toBeVisible();
  });

  test('enseña el nombre, la marca y el color institucional', async ({ page }) => {
    // La instalación: el nombre, que es lo que se lee en toda la aplicación, y **el idioma**, que va
    // con él desde el 2026-09-30.
    await expect(page.getByRole('heading', { name: 'La instalación' })).toBeVisible();
    await expect(page.getByLabel('Nombre', { exact: true })).toBeVisible();
    await expect(page.getByLabel('El idioma de la instalación')).toBeVisible();

    // La marca: los dos huecos, cada uno con su campo de archivo etiquetado.
    await expect(page.getByRole('heading', { name: 'La marca' })).toBeVisible();
    await expect(page.getByText('Logo para los temas claros')).toBeVisible();
    await expect(page.getByText('Logo para los temas oscuros')).toBeVisible();
    await expect(page.getByLabel('Elegir un archivo')).toHaveCount(2);

    // **El color va dentro de la marca** desde el 2026-09-30: son la misma cosa, la identidad de la
    // institución, y estaban en dos tarjetas sin motivo.
    await expect(page.getByLabel('Color', { exact: true })).toBeVisible();
    await expect(page.getByText('Vista previa')).toBeVisible();

    // Y lo que sí tiene su tarjeta, con su aviso: el prefijo y el reparto.
    await expect(page.getByRole('heading', { name: 'La numeración y el reparto' })).toBeVisible();
    await expect(page.getByLabel('Prefijo de los números')).toBeVisible();
    await expect(page.getByText(/no cambia los números ya emitidos/)).toBeVisible();
  });

  test('se sube un logo desde el formulario y se ve en la vista previa', async ({ page }) => {
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="240" height="120">
      <rect width="240" height="120" fill="#6a3fb0"/></svg>`;

    await page
      .getByLabel('Elegir un archivo')
      .first()
      .setInputFiles({ name: 'logo.svg', mimeType: 'image/svg+xml', buffer: Buffer.from(svg) });

    // El botón está en cada hueco, así que se busca el hueco **por su título**: el título es su hijo
    // directo, así que su padre es el hueco y no la tarjeta que los contiene a los dos.
    const huecoClaro = page
      .getByRole('heading', { name: 'Logo para los temas claros' })
      .locator('..');
    await huecoClaro.getByRole('button', { name: 'Subir el logo' }).click();
    await expect(page.getByText('Logo subido. Ya se está usando.')).toBeVisible();

    // La vista previa del hueco claro enseña lo que se acaba de subir, de verdad.
    const previa = huecoClaro.locator('img');
    await expect
      .poll(() => previa.evaluate((nodo) => (nodo as HTMLImageElement).naturalWidth))
      .toBe(240);

    // Y el logo aparece también en el menú, sin recargar nada.
    await expect
      .poll(() =>
        page
          .locator('aside app-logo img')
          .evaluate((nodo) => (nodo as HTMLImageElement).currentSrc.includes('/api/settings/brand/logo')),
      )
      .toBe(true);

    // Se vuelve al de fábrica y se comprueba que de verdad vuelve.
    await huecoClaro.getByRole('button', { name: 'Volver al de fábrica' }).click();
    await expect(page.getByText('Ese hueco ha vuelto al logo de fábrica.')).toBeVisible();
    await expect(page.getByText('Se está usando el logo de fábrica.').first()).toBeVisible();
  });

  test('el color tiene vista previa antes de guardar, y se puede descartar', async ({ page }) => {
    const color = page.getByLabel('Color');

    // Un amarillo puro: no se lee sobre blanco, así que el backend tiene que oscurecerlo.
    await color.evaluate((nodo) => {
      const entrada = nodo as HTMLInputElement;
      entrada.value = '#ffff00';
      entrada.dispatchEvent(new Event('input', { bubbles: true }));
    });

    // La vista previa dice cómo queda de verdad, ya resuelto: el amarillo no se lee sobre blanco, así
    // que en los temas claros se oscurece; y en los oscuros se queda tal cual, que sí se lee.
    const resumen = page.getByText(/En los temas claros queda/);
    await expect(resumen).toBeVisible();

    const texto = (await resumen.textContent()) ?? '';
    const claro = /claros queda: (#[0-9a-f]{6})/i.exec(texto)?.[1];
    const oscuro = /oscuros: (#[0-9a-f]{6})/i.exec(texto)?.[1];

    expect(claro, `El color de los temas claros debería haberse oscurecido (${texto})`).not.toBe('#ffff00');
    expect(oscuro).toBe('#ffff00');

    // Y los botones de guardar y descartar sólo están cuando hay algo que guardar.
    await expect(page.getByRole('button', { name: 'Guardar la marca' })).toBeVisible();
    await page.getByRole('button', { name: 'Descartar' }).click();
    await expect(page.getByRole('button', { name: 'Guardar la marca' })).toHaveCount(0);
  });

  /**
   * **El nombre de la instalación**: lo que sustituye a «Catalina Support» en la pantalla de entrada,
   * en el menú lateral y en la pestaña del navegador (decisión del responsable, 2026-09-25).
   *
   * Se comprueba en los tres sitios y **se deja la instalación como estaba** al terminar, vaciando el
   * campo: es lo que hace el resto de casos de este archivo, y además prueba que vaciar el nombre
   * devuelve el de fábrica.
   */
  test('el nombre de la instalación se ve en la aplicación y vuelve al de fábrica', async ({ page }) => {
    await partirDelNombreDeFabrica(page);

    await page.getByLabel('Nombre', { exact: true }).fill('Mesa de ayuda de Acme');

    await page.getByRole('button', { name: 'Guardar la instalación' }).click();
    await expect(page.getByText('La instalación se ha guardado.')).toBeVisible();

    // En el menú lateral, que es donde se lee junto al logo.
    await expect(page.getByRole('complementary').getByText('Mesa de ayuda de Acme')).toBeVisible();

    // Y en la pestaña del navegador.
    await expect(page).toHaveTitle('Mesa de ayuda de Acme');

    // En la pantalla de entrada, que se pinta antes de que nadie entre. Se sale por donde sale una
    // persona —el menú, que en móvil es un cajón—, no con un clic a mano.
    await salir(page);
    await expect(page.getByText('Mesa de ayuda de Acme')).toBeVisible();
    await expect(page).toHaveTitle('Mesa de ayuda de Acme');

    // Y se deja como estaba: vacío es «vuelve el nombre de fábrica».
    await entrar(page, FABRICA.email, FABRICA.password);
    await page.goto('/settings');
    await partirDelNombreDeFabrica(page);
    await expect(page.getByRole('complementary').getByText('Catalina Support')).toBeVisible();
    await expect(page).toHaveTitle('Catalina Support');
  });

  /**
   * **Cómo se entra**: el método que está puesto y las dos configuraciones, con su botón de probar la
   * conexión (decisión del responsable, 2026-09-25).
   *
   * Lo que se prueba es lo que hace que esto sirva de algo: que la configuración se cambia **desde la
   * pantalla**, que **los secretos no salen** —la pantalla dice si hay uno puesto y nunca lo enseña— y
   * que el método se puede poner y volver a dejar donde estaba.
   */
  test('el método de entrada se cambia desde Configuración', async ({ page, request }) => {
    // Se parte del método local, que es como queda el entorno de ejemplo: si otra ejecución se quedó a
    // medias, esto lo devuelve a un estado conocido antes de empezar.
    await ponerElMetodo(request, 'local');
    await page.reload();

    const metodo = page.getByLabel('Método de entrada');
    await expect(metodo).toHaveValue('local');

    // **Con «Cuentas de la aplicación» no se enseña ninguna de las dos configuraciones** (decisión del
    // responsable, 2026-09-27): cada panel sale al elegir su método, porque el que no se usa no es
    // trabajo que toque hacer.
    await expect(page.getByLabel('Servidor')).toHaveCount(0);
    await expect(page.getByLabel('Emisor del reino')).toHaveCount(0);

    // Se elige AD y aparece **su** configuración —la deja puesta el seeder—, sin su secreto.
    await metodo.selectOption('ad');
    await expect(page.getByLabel('Servidor')).toHaveValue('ldap');
    // **El secreto no llega nunca**: el campo está vacío y lo que se cuenta es que hay uno guardado.
    await expect(page.getByLabel('Contraseña de la cuenta de servicio')).toHaveValue('');
    await expect(page.getByText('Hay una contraseña guardada. Déjalo vacío para no cambiarla.')).toBeVisible();
    // Y la de Keycloak sigue sin verse: cada método enseña la suya y ninguna más.
    await expect(page.getByLabel('Emisor del reino')).toHaveCount(0);

    // Con Keycloak, al revés: sale su panel y se va el del directorio.
    await metodo.selectOption('keycloak');
    await expect(page.getByLabel('Emisor del reino')).toHaveValue(
      'https://dev-catalina-support.calibyou.com/sso/realms/catalina-support',
    );
    await expect(page.getByLabel('Servidor')).toHaveCount(0);

    // Y se vuelve a AD para guardar: vale desde la próxima entrada, sin reiniciar nada.
    await metodo.selectOption('ad');
    await expect(
      page.getByText('Se entra con la cuenta de la organización, contra el directorio.'),
    ).toBeVisible();
    await page.getByRole('button', { name: 'Guardar cómo se entra' }).click();
    await expect(page.getByText('Guardado. Vale desde la próxima entrada, sin reiniciar nada.')).toBeVisible();

    expect(await metodoDeEntrada(request)).toBe('ad');

    // Y se deja como estaba: el resto de las pruebas entra con cuentas locales.
    await ponerElMetodo(request, 'local');
    await page.reload();
    await expect(metodo).toHaveValue('local');
    expect(await metodoDeEntrada(request)).toBe('local');
  });

  /**
   * Los dos botones de «Probar la conexión» prueban **lo que hay en pantalla**, antes de guardarlo: es
   * lo que evita quedarse sin poder entrar por guardar algo que no funciona.
   *
   * Necesitan los dos servicios de pruebas levantados, así que se saltan si no lo están, en vez de
   * fallar por algo que no es del producto.
   */
  test('las dos conexiones se prueban desde la pantalla', async ({ page, request }) => {
    test.skip(
      !(await respondeAlProbar(request, 'directory')),
      'El directorio no responde: levántalo con `--profile auth`',
    );
    test.skip(
      !(await respondeAlProbar(request, 'keycloak')),
      'Keycloak no responde: levántalo con `--profile auth`',
    );

    // **Cada botón sale con su panel**: se elige el método, se prueba, y se cambia. Antes estaban los
    // dos a la vez, y desde el 2026-09-27 cada configuración se enseña sólo con su método.
    const metodo = page.getByLabel('Método de entrada');

    await metodo.selectOption('ad');
    await page.getByRole('button', { name: 'Probar la conexión' }).click();
    await expect(page.getByText('El directorio ha contestado y la cuenta de servicio entra.')).toBeVisible();

    await metodo.selectOption('keycloak');
    await page.getByRole('button', { name: 'Probar la conexión' }).click();
    await expect(
      page.getByText('El reino ha contestado y dice dónde está su pantalla de entrada.'),
    ).toBeVisible();

    // Y lo que no funciona se dice: con el servidor cambiado por uno que no existe, la prueba falla.
    await metodo.selectOption('ad');
    await page.getByLabel('Servidor').fill('no-existe-de-verdad');
    await page.getByRole('button', { name: 'Probar la conexión' }).click();
    await expect(page.getByText('No se ha podido conectar con el directorio')).toBeVisible();

    // Se deja como estaba, sin guardar nada: la prueba de conexión **no toca la base**. Tras recargar
    // vuelve el método guardado —local—, así que no hay ningún panel y **no se ha guardado nada**:
    // para ver que el directorio sigue con sus valores hay que volver a elegir AD.
    await page.reload();
    await expect(page.getByLabel('Método de entrada')).toHaveValue('local');
    await expect(page.getByLabel('Servidor')).toHaveCount(0);

    await page.getByLabel('Método de entrada').selectOption('ad');
    await expect(page.getByLabel('Servidor')).toHaveValue('ldap');
  });
  test('la región horaria y la dirección pública se configuran, y avisan si no hay https', async ({
    page,
    request,
  }) => {
    test.setTimeout(120_000);

    // **Se parte de una dirección distinta de la que se va a escribir**, para que el campo tenga algo
    // que cambiar: el botón de guardar está apagado mientras no haya cambios, y una prueba que dependa
    // de lo que hubiera guardado antes deja de ser una prueba. Se deja apuntada la de verdad, para
    // devolverla al final.
    const peticion = await tokenDeFabrica(request);
    const antes = (await (
      await request.get('/api/settings', { headers: { Authorization: `Bearer ${peticion}` } })
    ).json()) as Record<string, unknown>;
    await request.put('/api/settings', {
      headers: { Authorization: `Bearer ${peticion}` },
      data: { ...antes, publicAppUrl: 'https://antes.example.org' },
    });

    // Se entra con la cuenta de fábrica, que es la única que ve Configuración. `entrarComo` **sale
    // antes** de la sesión que traiga el navegador: la suite guarda el estado de una cuenta.
    await entrarComo(page, FABRICA.email, FABRICA.password);
    await page.goto('/settings');

    const guardar = page.getByRole('button', { name: 'Guardar la región y la dirección' });
    const direccion = page.getByLabel('Dirección pública');

    // **El aviso de que no hay https** (decisión 13): sale con `http://` y **no bloquea nada**, que es
    // lo que permite probar la instalación en local.
    await direccion.fill('http://soporte.local:8080');
    await expect(page.getByText('no está sirviendo por https')).toBeVisible();
    await guardar.click();
    await expect(page.getByText('La región y la dirección se han guardado.')).toBeVisible();

    // Con `https://` desaparece: es la forma en la que la sesión viaja cifrada.
    await direccion.fill('https://soporte.example.org');
    await expect(page.getByText('no está sirviendo por https')).toHaveCount(0);

    // **Y la zona horaria se elige de la lista y llega a la marca pública**, que es de donde la lee
    // toda la interfaz para pintar sus fechas (decisión 15).
    await page.getByLabel('Buscar una ciudad o una zona').fill('guayaquil');
    await page.getByRole('option', { name: 'America/Guayaquil', exact: true }).click();
    await guardar.click();
    await expect(page.getByText('La región y la dirección se han guardado.')).toBeVisible();

    const marca = await request.get('/api/settings/brand');
    expect(((await marca.json()) as { timeZone: string }).timeZone).toBe('America/Guayaquil');

    // Y la hora de la zona elegida se explica sola, con su desfase.
    await expect(page.getByText(/Ahora son las .*GMT-5/)).toBeVisible();

    // Se devuelve la dirección que había. Va aquí, y no en un `finally`, porque al final de una prueba
    // fallida el contexto ya está cerrado y el `PUT` no llega: entonces la deja el seeder.
    await request.put('/api/settings', {
      headers: { Authorization: `Bearer ${peticion}` },
      data: { ...antes, publicAppUrl: 'https://dev-catalina-support.calibyou.com' },
    });
  });
  test('las tarjetas de Configuración están en su orden y cada cosa en la suya', async ({ page }) => {
    // **La pantalla se ordenó el 2026-09-30** (decisión del responsable): cada tarjeta lleva lo que su
    // título dice. Antes, «La numeración y el reparto» tenía dentro el idioma, y el prefijo y el reparto
    // vivían en «Región horaria y dirección pública». Esta prueba es el tope que impide que se vuelva a
    // descolocar: mira los títulos en orden y dónde está cada campo.
    await entrarComo(page, FABRICA.email, FABRICA.password);
    await page.goto('/settings');

    const tarjetas = page.locator('app-tarjeta');
    await expect(tarjetas).toHaveCount(6);

    // Los títulos, en orden: de lo que la instalación **es** a cómo se entra, cómo trabaja y cómo avisa.
    for (const [i, titulo] of [
      'La instalación',
      'La marca',
      'Método de autenticación',
      'La numeración y el reparto',
      'Región horaria y dirección pública',
      'Los correos',
    ].entries()) {
      await expect(tarjetas.nth(i).getByRole('heading').first()).toHaveText(titulo);
    }

    // **El nombre y el idioma, juntos**: los dos son lo que la instalación es.
    await expect(tarjetas.nth(0).getByLabel('Nombre')).toBeVisible();
    await expect(tarjetas.nth(0).getByLabel('El idioma de la instalación')).toBeVisible();

    // **El logo y el color, juntos**: los dos son la marca.
    await expect(tarjetas.nth(1).locator('input[type="color"]')).toBeVisible();
    await expect(tarjetas.nth(1).getByText('Logo para los temas claros')).toBeVisible();

    // **El prefijo y el reparto, en la numeración**, que es lo que dice su título.
    await expect(tarjetas.nth(3).getByLabel('Prefijo de los números')).toBeVisible();
    await expect(tarjetas.nth(3).getByLabel('Reparto').first()).toBeVisible();

    // **Y la región y la dirección van juntas**, sin la numeración dentro.
    await expect(tarjetas.nth(4).getByLabel('Buscar una ciudad o una zona')).toBeVisible();
    await expect(tarjetas.nth(4).getByLabel('Dirección pública')).toBeVisible();
    await expect(tarjetas.nth(4).getByLabel('Prefijo de los números')).toHaveCount(0);
  });
});
