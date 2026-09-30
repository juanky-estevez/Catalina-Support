import { expect, test, type Page } from '@playwright/test';

/**
 * El tema, probado **mirando los colores**: que siga al sistema, que el selector lo cambie de verdad,
 * que lo elegido se recuerde, y —lo que más importa— que **los ocho temas se lean**.
 *
 * La última es la que hace que esto valga: el contraste se mide **sobre lo que el navegador pinta**,
 * no sobre la tabla de colores. Se mide así a propósito: en los temas de fábrica los colores se
 * declaran con `light-dark()`, y leer la variable devuelve el texto `light-dark(…)` **sin resolver**,
 * porque las variables CSS se heredan tal cual. Lo que decide si algo se lee es el color pintado.
 *
 * **«Automático» no se muestra ni se elige** (decisión del responsable, 2026-09-23): seguir al
 * sistema es no haber elegido, y el selector enseña el tema que se está viendo.
 */
test.describe('El tema', () => {
  /** Los ocho, tal y como se ven en el selector: agrupados, y los de fábrica primero en su grupo. */
  const TEMAS_POR_GRUPO: Record<string, string[]> = {
    Claros: ['Claro', 'Papel', 'Niebla', 'Alto contraste'],
    Oscuros: ['Oscuro', 'Grafito', 'Noche', 'Sepia'],
  };

  const TEMAS = Object.values(TEMAS_POR_GRUPO).flat();

  /**
   * Elige un tema y **espera a que los colores dejen de moverse**.
   *
   * Los botones llevan `transition`, así que medir justo después de cambiar de tema da el color **a
   * mitad de la animación**, que no es el de ninguna paleta. Se espera a que dos medidas seguidas
   * coincidan, que es cuando el navegador ha terminado.
   */
  async function elegirTema(page: Page, nombre: string): Promise<void> {
    await page.getByLabel('Tema').selectOption({ label: nombre });
    await expect(page.getByLabel('Tema')).toHaveValue(/\S/);

    await expect
      .poll(
        async () => {
          const antes = await colorDelBoton(page);
          await page.waitForTimeout(120);
          const despues = await colorDelBoton(page);
          return antes === despues;
        },
        { message: `Los colores del tema «${nombre}» no se estabilizan` },
      )
      .toBe(true);
  }

  /** El fondo del botón principal, que es el que lleva la transición. */
  async function colorDelBoton(page: Page): Promise<string> {
    return page.evaluate(
      () => getComputedStyle(document.querySelector('button[type="submit"]')!).backgroundColor,
    );
  }

  /** El color de fondo de la página, tal y como lo pinta el navegador. */
  async function fondo(page: Page): Promise<string> {
    return page.evaluate(() => getComputedStyle(document.body).backgroundColor);
  }

  test('no hay ninguna opción «automático», y están los ocho temas', async ({ page }) => {
    await page.goto('/login');

    const selector = page.getByLabel('Tema');
    await expect(selector).toBeVisible();

    // Los ocho, agrupados en claros y oscuros, y **los de fábrica primero en cada grupo**.
    const etiquetas = (await selector.locator('option').allTextContents()).map((t) => t.trim());
    expect(etiquetas).toEqual(TEMAS);

    const grupos = await selector.locator('optgroup').evaluateAll((nodos) =>
      nodos.map((nodo) => ({
        nombre: (nodo as HTMLOptGroupElement).label.trim(),
        opciones: [...nodo.querySelectorAll('option')].map((opcion) => opcion.textContent!.trim()),
      })),
    );
    expect(grupos).toEqual(
      Object.entries(TEMAS_POR_GRUPO).map(([nombre, opciones]) => ({ nombre, opciones })),
    );

    // Y ninguna opción que ofrezca «automático»: eso es no haber elegido, no una opción.
    expect(etiquetas.join(' ')).not.toMatch(/automátic|automatic/i);
  });

  test('de fábrica se ve según el sistema, y el selector enseña lo que se ve', async ({ browser }) => {
    const claro = await browser.newContext({ colorScheme: 'light' });
    const paginaClara = await claro.newPage();
    await paginaClara.goto('/login');
    const fondoClaro = await fondo(paginaClara);
    // Sin atributo en `html`: deciden los de fábrica según el sistema.
    expect(await paginaClara.evaluate(() => document.documentElement.hasAttribute('data-theme'))).toBe(
      false,
    );
    await expect(paginaClara.getByLabel('Tema')).toHaveValue('claro');
    await claro.close();

    const oscuro = await browser.newContext({ colorScheme: 'dark' });
    const paginaOscura = await oscuro.newPage();
    await paginaOscura.goto('/login');
    const fondoOscuro = await fondo(paginaOscura);
    expect(await paginaOscura.evaluate(() => document.documentElement.hasAttribute('data-theme'))).toBe(
      false,
    );
    await expect(paginaOscura.getByLabel('Tema')).toHaveValue('oscuro');
    await oscuro.close();

    expect(fondoClaro).not.toBe(fondoOscuro);
    expect(tono(fondoOscuro)).toBeLessThan(80);
    expect(tono(fondoClaro)).toBeGreaterThan(180);
  });

  test('el selector cambia el tema y lo recuerda al recargar', async ({ page }) => {
    await page.goto('/login');

    await elegirTema(page, 'Sepia');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'sepia');
    expect(tono(await fondo(page))).toBeLessThan(80);

    await elegirTema(page, 'Niebla');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'niebla');
    expect(tono(await fondo(page))).toBeGreaterThan(180);

    // Se recuerda: quien eligió Grafito no vuelve al tema de su sistema al recargar.
    await elegirTema(page, 'Grafito');
    await page.reload();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'grafito');
  });

  test('cada uno de los ocho temas cambia de verdad el aspecto', async ({ page }) => {
    await page.goto('/login');

    const fondos = new Set<string>();
    for (const nombre of TEMAS) {
      await elegirTema(page, nombre);
      fondos.add(await fondo(page));
    }

    // Ocho temas con el mismo fondo serían ocho nombres y un solo tema.
    expect(fondos.size, `Fondos repetidos entre los ocho temas: ${[...fondos].join(', ')}`).toBe(8);
  });

  test('los ocho temas se leen: el contraste medido sobre lo que se pinta', async ({ page }) => {
    await page.goto('/login');

    for (const nombre of TEMAS) {
      await elegirTema(page, nombre);

      // Se mide lo que el navegador pinta: el fondo del cuerpo, el texto de verdad, el campo, el
      // botón y los tres avisos, creados al vuelo con las clases del inventario.
      const medidas = await page.evaluate(() => {
        const pintar = (clases: string) => {
          const caja = document.createElement('div');
          caja.className = clases;
          caja.textContent = 'medida';
          document.body.appendChild(caja);
          const estilo = getComputedStyle(caja);
          const resultado = { color: estilo.color, fondo: estilo.backgroundColor };
          caja.remove();
          return resultado;
        };

        const cuerpo = getComputedStyle(document.body);
        const tarjeta = document.querySelector('section');
        const campo = document.querySelector('input[type="email"]');
        const boton = document.querySelector('button[type="submit"]');

        return {
          fondo: cuerpo.backgroundColor,
          texto: cuerpo.color,
          tarjeta: tarjeta ? getComputedStyle(tarjeta).backgroundColor : cuerpo.backgroundColor,
          apagado: pintar('text-apagado').color,
          acento: pintar('text-primario').color,
          botonTexto: boton ? getComputedStyle(boton).color : cuerpo.color,
          botonFondo: boton ? getComputedStyle(boton).backgroundColor : cuerpo.backgroundColor,
          campoBorde: campo ? getComputedStyle(campo).borderTopColor : cuerpo.color,
          campoFondo: campo ? getComputedStyle(campo).backgroundColor : cuerpo.backgroundColor,
          peligro: pintar('text-peligro bg-peligro-suave'),
          exito: pintar('text-exito bg-exito-suave'),
          aviso: pintar('text-aviso bg-aviso-suave'),
        };
      });

      const fallos: string[] = [];
      const exigir = (que: string, razon: number, minimo: number) => {
        if (razon < minimo) {
          fallos.push(`${que}: ${razon.toFixed(2)} (mínimo ${minimo})`);
        }
      };

      exigir('texto sobre fondo', contraste(medidas.texto, medidas.fondo), 4.5);
      exigir('texto sobre la tarjeta', contraste(medidas.texto, medidas.tarjeta), 4.5);
      exigir('texto apagado sobre fondo', contraste(medidas.apagado, medidas.fondo), 4.5);
      exigir('acento sobre fondo', contraste(medidas.acento, medidas.fondo), 4.5);
      exigir('texto del botón sobre su fondo', contraste(medidas.botonTexto, medidas.botonFondo), 4.5);
      exigir('borde del campo contra su fondo', contraste(medidas.campoBorde, medidas.campoFondo), 3);
      exigir('peligro sobre su fondo', contraste(medidas.peligro.color, medidas.peligro.fondo), 4.5);
      exigir('éxito sobre su fondo', contraste(medidas.exito.color, medidas.exito.fondo), 4.5);
      exigir('aviso sobre su fondo', contraste(medidas.aviso.color, medidas.aviso.fondo), 4.5);

      expect(fallos, `El tema «${nombre}» no se lee:\n${fallos.join('\n')}`).toEqual([]);
    }
  });

  test('se sigue al sistema hasta que alguien toca, y el sistema cambia la página en vivo', async ({
    page,
  }) => {
    await page.goto('/login');
    await expect(page.getByLabel('Tema')).toHaveValue('claro');

    // El sistema pasa a oscuro con la página abierta: la aplicación cambia con él, sin recargar.
    await page.emulateMedia({ colorScheme: 'dark' });
    await expect(page.getByLabel('Tema')).toHaveValue('oscuro');
    expect(tono(await fondo(page))).toBeLessThan(80);
  });

  test('el selector se puede usar con el teclado', async ({ page }) => {
    await page.goto('/login');

    await page.getByLabel('Tema').focus();
    await page.keyboard.press('ArrowDown');
    await page.keyboard.press('Enter');

    // Con el teclado se cambia igual que con el ratón: no es un adorno con aspecto de control.
    const elegido = await page.getByLabel('Tema').inputValue();
    await expect(page.locator('html')).toHaveAttribute('data-theme', elegido);
  });
});

/** La claridad de un color `rgb(r, g, b)`, de 0 a 255. */
function tono(color: string): number {
  const numeros = componentes(color);
  return (numeros[0]! + numeros[1]! + numeros[2]!) / 3;
}

/** El contraste entre dos colores pintados, como lo define la norma (WCAG 2.1). */
function contraste(uno: string, otro: string): number {
  const luminancia = (color: string) => {
    const [r, g, b] = componentes(color).map((valor) => {
      const canal = valor / 255;
      return canal <= 0.03928 ? canal / 12.92 : ((canal + 0.055) / 1.055) ** 2.4;
    });

    return 0.2126 * r! + 0.7152 * g! + 0.0722 * b!;
  };

  const a = luminancia(uno);
  const b = luminancia(otro);

  return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
}

/** Los tres números de un color pintado: `rgb(...)` o `rgba(...)`. */
function componentes(color: string): number[] {
  const numeros = color.match(/[\d.]+/g)?.map(Number) ?? [];
  if (numeros.length < 3) {
    throw new Error(`No se pudo leer el color pintado: ${color}`);
  }

  return numeros.slice(0, 3);
}
