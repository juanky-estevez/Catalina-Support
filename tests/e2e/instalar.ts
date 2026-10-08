import { chromium, devices, expect, request } from '@playwright/test';

/** Sólo actúa en la instancia exclusiva de tests.yml; no toca la instalación del usuario. */
export async function instalar(): Promise<void> {
  const baseURL = process.env['BASE_URL'] ?? 'http://frontend.localhost:11001';
  const api = await request.newContext({ baseURL });
  await expect
    .poll(
      async () => {
        try {
          return (await api.get('/api/health', { timeout: 3000 })).status();
        } catch {
          return 0;
        }
      },
      { timeout: 180_000 },
    )
    .toBe(200);
  const estado = await (await api.get('/api/setup')).json();
  if (estado.installed)
    throw new Error(
      'La instancia de pruebas ya está sellada. Ejecuta down -v sobre tests.yml antes de repetir.',
    );
  const browser = await chromium.launch({ args: ['--host-resolver-rules=MAP frontend.localhost frontend'] });
  try {
    // Recorremos los cinco pasos en PC y móvil antes de sellar, sin cuentas ni tickets. La IA usa
    // el servidor compatible falso de tests.yml: no descarga modelos ni consume un proveedor.
    for (const movil of [false, true]) {
      const context = await browser.newContext({
        ...(movil ? devices['Pixel 7'] : {}),
        locale: 'es-ES',
      });
      const page = await context.newPage();
      await page.goto(baseURL);
      await expect(page).toHaveURL(/\/setup$/);
      if (!movil) {
        // Una base nueva empieza en inglés aunque el navegador de esta prueba prefiera español.
        await expect(page.locator('html')).toHaveAttribute('lang', 'en');
        await expect(page.locator('#idioma-instalacion')).toHaveValue('en');
        await expect(page.getByRole('button', { name: 'Next', exact: true })).toBeVisible();

        // El selector traduce el asistente entero en el acto, antes de guardar el paso.
        await page.locator('#idioma-instalacion').selectOption('es');
        await expect(page.locator('html')).toHaveAttribute('lang', 'es');
        await expect(page.getByRole('button', { name: 'Siguiente', exact: true })).toBeVisible();
      } else {
        // Al reanudar manda el idioma guardado por el primer recorrido, no el del navegador.
        await expect(page.locator('html')).toHaveAttribute('lang', 'es');
        await expect(page.locator('#idioma-instalacion')).toHaveValue('es');
      }
      await page.locator('#nombre-instalacion').fill('Catalina Support');
      await page
        .getByRole('button', { name: 'Siguiente', exact: true })
        .click();
      await page.locator('#metodo-entrada').selectOption('local');
      await page
        .getByRole('button', { name: 'Siguiente', exact: true })
        .click();
      await expect(page.locator('#zona-seleccionada')).toHaveValue(
        movil ? 'America/Guayaquil' : 'UTC',
      );
      await page.locator('#zona-busqueda').fill('America/Guayaquil');
      await page
        .getByRole('option')
        .filter({ hasText: 'America/Guayaquil' })
        .click();
      await expect(page.locator('#zona-seleccionada')).toHaveValue('America/Guayaquil');
      await expect(page.locator('#zona-seleccionada')).toHaveAttribute('readonly', '');
      await page.locator('#direccion-publica').fill(baseURL);
      await page
        .getByRole('button', { name: 'Siguiente', exact: true })
        .click();
      await page.locator('#correo-host').fill('mail');
      await page.locator('#correo-puerto').fill('1025');
      await page.locator('#correo-remitente-nombre').fill('Catalina Support');
      await page
        .locator('#correo-remitente-correo')
        .fill('no-responder@catalina-support.local');
      const prueba = page.waitForResponse((r) =>
        r.url().endsWith('/api/setup/mail/test'),
      );
      await page
        .getByRole('button', { name: 'Probar la conexión', exact: true })
        .click();
      expect((await prueba).status()).toBe(200);
      const toast = page.locator('app-toast');
      await expect(toast).toContainText(
        'El servidor de correo ha contestado y la autenticación ha funcionado.',
      );
      expect(await toast.evaluate((elemento) => getComputedStyle(elemento).position)).toBe('fixed');
      const caja = await toast.boundingBox();
      const ventana = page.viewportSize();
      expect(caja).not.toBeNull();
      expect(ventana).not.toBeNull();
      expect(caja!.x).toBeGreaterThanOrEqual(0);
      expect(caja!.y).toBeGreaterThanOrEqual(0);
      expect(caja!.x + caja!.width).toBeLessThanOrEqual(ventana!.width);
      await toast.getByRole('button', { name: 'Cerrar' }).click();
      await expect(toast).toHaveCount(0);
      await page
        .getByRole('button', { name: 'Siguiente', exact: true })
        .click();
      await page.locator('#ia-modo-setup').selectOption(movil ? 'remote' : 'provider');
      if (!movil) {
        await page.locator('#ia-proveedor-setup').selectOption('openai-compatible');
        await expect(page.getByText(/costos del proveedor|provider charges/)).toBeVisible();
      }
      await page.locator('#ia-direccion-setup').fill('http://ai-test:8080');
      await page.locator('#ia-modelo-setup').fill('modelo-de-prueba');
      await page.locator('#ia-auth-setup').selectOption('none');
      await page.locator('input[type="checkbox"]').check();
      const pruebaIA = page.waitForResponse((r) =>
        r.url().endsWith('/api/setup/ai'),
      );
      await page
        .getByRole('button', { name: 'Siguiente', exact: true })
        .click();
      const respuestaIA = await pruebaIA;
      expect(respuestaIA.status(), await respuestaIA.text()).toBe(200);
      const resumen = page.locator('app-tarjeta').filter({
        has: page.getByRole('heading', { name: 'Resumen' }),
      });
      await expect(resumen).toContainText(movil ? 'En otro servidor' : 'Proveedor externo');
      await expect(resumen).toContainText('openai-compatible');
      await expect(resumen).toContainText('modelo-de-prueba');
      await expect(resumen).toContainText('No requerida');
      if (!movil) {
        await expect(resumen).toContainText('Privacidad y posible costo');
        await expect(resumen).toContainText('Aceptados');
      }
      await expect(
        page.getByRole('button', { name: /Terminar/ }),
      ).toBeVisible();
      await page.screenshot({
        path: `/resultados/asistente-${movil ? 'movil' : 'pc'}.png`,
        fullPage: true,
      });
      if (movil) {
        await page.getByRole('button', { name: /Terminar/ }).click();
        await expect(page).toHaveURL(/\/login$/);
      }
      await context.close();
    }
    const entrada = await api.post('/api/auth/login', {
      data: { email: 'admin', password: process.env['ADMIN_PASSWORD'] },
    });
    expect(entrada.status()).toBe(200);
    const { token } = await entrada.json();
    const headers = { Authorization: `Bearer ${token}` };
    const respuesta = await api.get('/api/settings', { headers });
    const ajustes = await respuesta.json();
    const guardado = await api.put('/api/settings', {
      headers,
      data: {
        ...ajustes,
        directory: {
          host: 'ldap',
          port: '389',
          useTls: false,
          bindDn: 'cn=admin,dc=ejemplo,dc=com',
          bindPassword: 'admin-directorio',
          searchBase: 'ou=personas,dc=ejemplo,dc=com',
          userFilter: '(mail=%s)',
          attrEmail: 'mail',
          attrName: 'givenName',
          attrLastName: 'sn',
          attrId: 'uid',
        },
        keycloak: {
          issuer: 'http://keycloak:8080/sso/realms/catalina-support',
          internalIssuer: '',
          clientId: 'catalina-support',
          clientSecret: 'el-secreto-de-desarrollo',
          redirectUri: `${baseURL}/api/auth/keycloak/callback`,
        },
      },
    });
    expect(guardado.status(), await guardado.text()).toBe(200);
    const categoria = await api.post('/api/tickets/categories', {
      headers,
      data: { name: 'Impresoras' },
    });
    expect(categoria.status(), await categoria.text()).toBe(201);
    const etiqueta = await api.post('/api/tickets/tags', {
      headers,
      data: { tag: 'correo' },
    });
    expect(etiqueta.status(), await etiqueta.text()).toBe(201);
    console.log(
      'Instalación vacía verificada en PC y móvil, sellada y preparada para la suite aislada.',
    );
  } finally {
    await browser.close();
    await api.dispose();
  }
}
