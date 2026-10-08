import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

import { abrirMenuSiEsCajon, entrar, FABRICA } from '../ayudas';

/**
 * La marca de la instalación: el logo.
 *
 * Se prueba **en el navegador y con la API de verdad**: se sube un logo propio, se comprueba que la
 * pantalla de entrada lo enseña de verdad —que la imagen carga, no sólo que hay un `<img>`—, y se
 * vuelve al de fábrica. Dejar la instalación como estaba forma parte de la prueba.
 */
test.describe('La marca', () => {
  /** Sube un logo pequeño por la API, como haría la pantalla de Configuración. */
  async function subirLogo(peticion: APIRequestContext, token: string, svg: string, variante: string) {
    const respuesta = await peticion.post(`/api/settings/brand/logo?variant=${variante}`, {
      headers: { Authorization: `Bearer ${token}` },
      multipart: {
        logo: { name: 'logo.svg', mimeType: 'image/svg+xml', buffer: Buffer.from(svg) },
      },
    });

    expect(respuesta.ok(), await respuesta.text()).toBeTruthy();
  }

  /** Quita el logo propio de una variante. */
  async function quitarLogo(peticion: APIRequestContext, token: string, variante: string) {
    const respuesta = await peticion.delete(`/api/settings/brand/logo?variant=${variante}`, {
      headers: { Authorization: `Bearer ${token}` },
    });

    expect(respuesta.ok(), await respuesta.text()).toBeTruthy();
  }

  async function entrarComoAdmin(peticion: APIRequestContext): Promise<string> {
    const respuesta = await peticion.post('/api/auth/login', {
      data: { email: FABRICA.email, password: FABRICA.password },
    });
    expect(respuesta.ok(), 'No se pudo entrar como administrador').toBeTruthy();

    const { token } = (await respuesta.json()) as { token: string };
    return token;
  }

  /** Espera a que la imagen del logo esté cargada y devuelve lo que el navegador ve. */
  async function logoPintado(page: Page) {
    const imagen = page.locator('app-logo img');
    await expect(imagen).toBeVisible();
    await expect
      .poll(() => imagen.evaluate((nodo) => (nodo as HTMLImageElement).naturalWidth))
      .toBeGreaterThan(0);

    return imagen.evaluate((nodo) => {
      const img = nodo as HTMLImageElement;
      const marco = img.parentElement!;
      return {
        src: img.getAttribute('src') ?? '',
        ancho: img.naturalWidth,
        alto: img.naturalHeight,
        // El recuadro mide 128 px en móvil y 160 de tablet para arriba, y la imagen va dentro con su
        // relleno (docs/interfaz-y-experiencia.md, sección 6.4).
        altoDelRecuadro: marco.getBoundingClientRect().height,
        altoEnPantalla: img.getBoundingClientRect().height,
        marcoConBorde: getComputedStyle(marco).borderTopWidth,
      };
    });
  }

  test('el selector de tema no enseña la palabra «Tema», pero la sigue teniendo', async ({ page }) => {
    await page.goto('/login');

    // La etiqueta **sigue en la página** —es el nombre accesible del control—, pero no se ve: se
    // comprueba que ocupa lo que ocupa un texto oculto, no que no exista.
    const etiqueta = page.locator('label[for="tema"]');
    await expect(etiqueta).toHaveCount(1);
    await expect(etiqueta).toHaveClass(/sr-only/);

    const ancho = await etiqueta.evaluate((nodo) => nodo.getBoundingClientRect().width);
    expect(ancho, 'la etiqueta del tema no debería verse').toBeLessThanOrEqual(2);

    // Y el control sigue teniendo nombre: quien usa un lector de pantalla sabe de qué es.
    await expect(page.getByLabel('Tema')).toBeVisible();
  });

  /**
   * **La versión del sistema** (decisión del responsable, 2026-09-25): se lee en la fila de salir del
   * menú lateral, alineada a la derecha, y **también en el pie de la pantalla de entrada**, que es
   * donde sirve para decir qué versión se está mirando al reportar un fallo.
   *
   * Se comprueba contra lo que dice la API, no contra un número escrito en la prueba: **el día que se
   * cierre una versión, esto sigue valiendo** sin tocar el caso.
   */
  test('la versión del sistema se lee en la entrada y en el menú', async ({ page, request }, info) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede entrar al menú');

    const marca = (await (await request.get('/api/settings/brand')).json()) as { version?: string };
    expect(marca.version, 'la marca pública tiene que decir la versión').toBeTruthy();
    const esperada = `v${marca.version}`;

    // En el pie de la pantalla de entrada, antes de entrar.
    await page.goto('/login');
    const enLaEntrada = page.getByText(esperada, { exact: true });
    await expect(enLaEntrada).toBeVisible();

    // Y es un texto, no un enlace: no lleva a ningún sitio.
    await expect(page.getByRole('link', { name: esperada })).toHaveCount(0);

    // En el menú lateral, en la fila de salir. En móvil el menú es un cajón y hay que abrirlo, que es
    // como lo haría una persona.
    await entrar(page, FABRICA.email, FABRICA.password);
    await abrirMenuSiEsCajon(page);

    const enElMenu = page.getByText(esperada, { exact: true });
    await expect(enElMenu).toBeVisible();

    // **Alineada a la derecha**: empieza más a la derecha que el botón de salir, y acaba dentro del
    // menú. Es lo que la distingue de un control más.
    const salir = await page.getByRole('button', { name: 'Salir' }).boundingBox();
    const version = await enElMenu.boundingBox();
    const menu = await page.locator('aside').boundingBox();

    expect(version!.x).toBeGreaterThan(salir!.x + salir!.width - 1);
    expect(version!.x + version!.width).toBeLessThanOrEqual(menu!.x + menu!.width);

    // Y se lee: no puede ir con el color apagado tan flojo que no se vea.
    const color = await enElMenu.evaluate((nodo) => getComputedStyle(nodo).color);
    const fondo = await page.locator('aside').evaluate((nodo) => getComputedStyle(nodo).backgroundColor);
    expect(color).not.toBe(fondo);

    // **Plegado no se enseña**: un número suelto al lado de un icono no dice nada. Sólo en PC, que es
    // donde el menú se pliega: en móvil es un cajón y no hay nada que plegar.
    if (info.project.name === 'pc') {
      await page.getByRole('button', { name: 'Plegar el menú' }).click();
      await expect(enElMenu).toHaveCount(0);
    }
  });

  /**
   * **Los dos logos de fábrica, uno por tema** (decisión del responsable, 2026-09-25): un logo hecho
   * para un fondo claro no se lee sobre uno oscuro, así que la aplicación trae los dos y enseña el que
   * toca. Se comprueba **cambiando el tema y mirando qué archivo carga de verdad**: si el de fábrica
   * del tema oscuro no existiera, `naturalWidth` sería 0.
   */
  test('el logo de fábrica cambia con el tema', async ({ page, request }) => {
    // Sin logo propio: si la instalación tuviera uno, no se vería ninguno de los dos de fábrica. Se
    // quita el de los dos huecos y **se deja como estaba** al terminar.
    if (FABRICA.password) {
      const token = await entrarComoAdmin(request);
      await quitarLogo(request, token, 'claro');
      await quitarLogo(request, token, 'oscuro');
    }

    await page.goto('/login');

    // Con un tema claro, el de fábrica claro, y cargado de verdad.
    await page.getByLabel('Tema').selectOption({ label: 'Claro' });
    const claro = await logoPintado(page);
    expect(claro.src).toContain('logo-catalina-support-light');
    expect(claro.ancho, 'el logo de fábrica claro tiene que cargar').toBeGreaterThan(0);

    // Y con uno oscuro, el otro.
    await page.getByLabel('Tema').selectOption({ label: 'Oscuro' });
    const oscuro = await logoPintado(page);
    expect(oscuro.src).toContain('logo-catalina-support-dark');
    expect(oscuro.ancho, 'el logo de fábrica oscuro tiene que cargar').toBeGreaterThan(0);

    // Son dos archivos distintos, no el mismo con otro nombre.
    expect(oscuro.src).not.toBe(claro.src);
  });

  test('la pantalla de entrada enseña el logo, en un recuadro con su borde', async ({ page }) => {
    await page.goto('/login');

    const logo = await logoPintado(page);

    // La imagen cargada de verdad: si el archivo diera 404, `naturalWidth` sería 0.
    expect(logo.ancho).toBeGreaterThan(0);
    // Bastante grande —es la primera cosa que se ve— y sin pasarse del hueco.
    expect(logo.altoDelRecuadro).toBeGreaterThanOrEqual(128);
    expect(logo.altoDelRecuadro).toBeLessThanOrEqual(160);
    expect(logo.altoEnPantalla).toBeGreaterThanOrEqual(100);
    // Y va en el recuadro del tema, no suelto.
    expect(logo.marcoConBorde).not.toBe('0px');
  });

  test('el logo propio sustituye al de fábrica, y al quitarlo se vuelve', async ({ page, request }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede probar la subida');

    const token = await entrarComoAdmin(request);
    // La variante es `claro` o `oscuro`: el nombre del archivo lo pone la aplicación.
    const propio = 'claro';

    try {
      await quitarLogo(request, token, propio);

      // Sin logo propio: el de fábrica, que viene con la aplicación.
      await page.goto('/login');
      const deFabrica = await logoPintado(page);
      expect(deFabrica.src).toContain('logo-catalina-support');

      // Se sube uno propio: la pantalla lo enseña sin recargar la instalación.
      await subirLogo(
        request,
        token,
        '<svg xmlns="http://www.w3.org/2000/svg" width="200" height="80"><rect width="200" height="80" fill="#6a3fb0"/></svg>',
        propio,
      );

      await page.reload();
      const subido = await logoPintado(page);
      expect(subido.src).toContain('/api/settings/brand/logo');
      expect(subido.src).toContain('theme=claro');
      // Y el sello de la versión: es lo que impide que el navegador siga enseñando el viejo.
      expect(subido.src).toContain('v=');
      // Las medidas del SVG que se subió.
      expect(subido.ancho).toBe(200);
      expect(subido.alto).toBe(80);
    } finally {
      // La instalación se deja como estaba: la prueba no puede dejar su logo puesto.
      await quitarLogo(request, token, propio);
    }

    await page.reload();
    const devuelta = await logoPintado(page);
    expect(devuelta.src).toContain('logo-catalina-support');
  });

  test('el color institucional de la instalación se aplica a los dos temas de fábrica', async ({
    page,
    request,
  }) => {
    test.skip(!FABRICA.password, 'Falta ADMIN_PASSWORD: sin ella no se puede probar el color');

    const token = await entrarComoAdmin(request);
    const leerConfiguracion = await request.get('/api/settings', {
      headers: { Authorization: `Bearer ${token}` },
    });
    const original = (await leerConfiguracion.json()) as Record<string, string>;

    try {
      // Un color que no se lee por sí solo sobre blanco: el backend tiene que oscurecerlo para el
      // tema claro y dejarlo tal cual para el oscuro.
      const cambio = await request.put('/api/settings', {
        headers: { Authorization: `Bearer ${token}` },
        data: { ...original, primaryColor: '#ffff00' },
      });
      expect(cambio.ok(), await cambio.text()).toBeTruthy();

      await page.goto('/login');
      const claro = await page.evaluate(() =>
        getComputedStyle(document.documentElement).getPropertyValue('--institucional-claro').trim(),
      );
      const oscuro = await page.evaluate(() =>
        getComputedStyle(document.documentElement).getPropertyValue('--institucional-oscuro').trim(),
      );

      expect(claro, 'el color del tema claro debería haberse oscurecido').not.toBe('#ffff00');
      expect(oscuro).toBe('#ffff00');

      // Y sigue leyéndose: el texto del botón sobre su color.
      const boton = await page.locator('button[type="submit"]').evaluate((nodo) => {
        const estilo = getComputedStyle(nodo);
        return { color: estilo.color, fondo: estilo.backgroundColor };
      });
      expect(boton.color).not.toBe(boton.fondo);
    } finally {
      await request.put('/api/settings', {
        headers: { Authorization: `Bearer ${token}` },
        data: { ...original, primaryColor: original['primaryColor'] },
      });
    }
  });
});
