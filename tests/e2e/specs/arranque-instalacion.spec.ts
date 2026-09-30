import { expect, test } from '@playwright/test';

import { FABRICA, tokenDeFabrica } from '../ayudas';

/**
 * El primer arranque **en una instalación ya configurada**.
 *
 * El contenedor de las pruebas no habla con la base y no se puede quitar el sello, así que **el
 * camino de instalación de verdad no se puede recorrer aquí**: eso se prueba a mano. Lo que sí se
 * puede —y es lo que más importa de la guarda— es comprobar que una instalación terminada **no
 * vuelve a enseñar el asistente** y que su API contesta que está instalada
 * (`docs/primer-arranque.md`, secciones 2, 6 y 7).
 */
test.describe('El primer arranque, con la instalación ya terminada', () => {
  test('el estado dice que está instalada y /setup lleva a la entrada, con su aviso', async ({
    page,
    request,
  }) => {
    // Primero, lo que dice el backend: esta instalación está sellada.
    const respuesta = await request.get('/api/setup');
    expect(respuesta.status(), await respuesta.text()).toBe(200);

    const estado = (await respuesta.json()) as { installed: boolean };
    expect(estado.installed).toBe(true);

    // Y en el navegador: /setup no se ve, se va a la entrada.
    await page.goto('/setup');
    await expect(page).toHaveURL(/\/login/);

    // Se explica por qué, en vez de dejar a quien llegó con la dirección a mano sin saber qué pasó.
    await expect(page.getByText('Esta instalación ya está terminada')).toBeVisible();

    // Y la entrada es la entrada: su título y su formulario, no una pantalla a medias.
    await expect(page.getByRole('heading', { name: 'Entrar' })).toBeVisible();
    await expect(page.getByLabel('Correo electrónico')).toBeVisible();
  });

  test('la cuenta de fábrica sigue siendo la puerta de una instalación terminada', async ({
    request,
  }) => {
    // Sin `ADMIN_PASSWORD` no hay puerta que probar: no se inventa nada.
    test.skip(!FABRICA.password, 'No hay ADMIN_PASSWORD en el entorno de las pruebas');

    // Es lo que permite volver a cambiar el método si algún día se elige mal: la instalación sigue
    // teniendo una puerta que no depende de la configuración guardada.
    const token = await tokenDeFabrica(request);
    expect(token.length).toBeGreaterThan(0);
  });
});
