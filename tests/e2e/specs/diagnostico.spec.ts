import { test } from '@playwright/test';

/**
 * Diagnóstico: vuelca lo que hay de verdad en la pantalla de entrada.
 *
 * No es una prueba de nada, es la herramienta para mirar lo que el navegador tiene delante cuando
 * algo no se ve como debería: el HTML pintado, los errores de la consola y una captura.
 */
// Es una herramienta, no una prueba: se ejecuta a propósito cuando algo no se ve como debería.
test.skip(!process.env['DIAGNOSTICO'], 'Herramienta: se ejecuta con DIAGNOSTICO=1');

test('volcado de la pantalla de entrada', async ({ page }) => {
  const errores: string[] = [];
  page.on('console', (mensaje) => errores.push(`${mensaje.type()}: ${mensaje.text()}`));
  page.on('pageerror', (error) => errores.push(`pageerror: ${error.message}`));
  page.on('requestfailed', (peticion) =>
    errores.push(`requestfailed: ${peticion.url()} ${peticion.failure()?.errorText}`),
  );

  await page.goto('/');
  await page.waitForTimeout(1500);

  console.log('=== dirección ===', page.url());

  const formulario = await page.evaluate(() => {
    const form = document.querySelector('form');
    return form ? form.outerHTML : '(no hay ningún formulario en la página)';
  });

  console.log('=== formulario ===\n' + formulario);
  console.log('=== errores y peticiones fallidas ===\n' + (errores.join('\n') || '(ninguno)'));

  await page.screenshot({ path: 'test-results/diagnostico-login.png', fullPage: true });
});
