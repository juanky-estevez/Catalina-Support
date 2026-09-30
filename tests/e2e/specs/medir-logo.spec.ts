import { test } from '@playwright/test';

// Herramienta: mide el logo (tamaño, esquinas, color medio de los bordes) dibujándolo en un canvas.
test.skip(!process.env['DIAGNOSTICO'], 'Herramienta: se ejecuta con DIAGNOSTICO=1');

test('medir el logo', async ({ page }) => {
  await page.goto('/login');

  const datos = await page.evaluate(async () => {
    const imagen = new Image();
    // El de fábrica de los temas claros: es el que se ve en la pantalla de entrada de fábrica.
    imagen.src = '/logo-catalina-support-light.png';
    await imagen.decode();

    const lienzo = document.createElement('canvas');
    lienzo.width = imagen.naturalWidth;
    lienzo.height = imagen.naturalHeight;
    const ctx = lienzo.getContext('2d')!;
    ctx.drawImage(imagen, 0, 0);

    const pixel = (x: number, y: number) => {
      const d = ctx.getImageData(x, y, 1, 1).data;
      return `rgb(${d[0]}, ${d[1]}, ${d[2]})`;
    };

    const ancho = lienzo.width;
    const alto = lienzo.height;

    // Media de una franja de 8 píxeles pegada a cada borde.
    const mediaBorde = (lado: 'arriba' | 'abajo' | 'izquierda' | 'derecha') => {
      const datos = ctx.getImageData(
        lado === 'izquierda' ? 0 : lado === 'derecha' ? ancho - 8 : 0,
        lado === 'arriba' ? 0 : lado === 'abajo' ? alto - 8 : 0,
        lado === 'izquierda' || lado === 'derecha' ? 8 : ancho,
        lado === 'izquierda' || lado === 'derecha' ? alto : 8,
      ).data;

      let r = 0, g = 0, b = 0;
      for (let i = 0; i < datos.length; i += 4) {
        r += datos[i]!; g += datos[i + 1]!; b += datos[i + 2]!;
      }
      const n = datos.length / 4;
      return `rgb(${Math.round(r / n)}, ${Math.round(g / n)}, ${Math.round(b / n)})`;
    };

    const todos = ctx.getImageData(0, 0, ancho, alto).data;
    let r = 0, g = 0, b = 0;
    for (let i = 0; i < todos.length; i += 4) { r += todos[i]!; g += todos[i + 1]!; b += todos[i + 2]!; }
    const n = todos.length / 4;

    return {
      ancho,
      alto,
      esquinaSupIzq: pixel(2, 2),
      esquinaSupDer: pixel(ancho - 3, 2),
      esquinaInfIzq: pixel(2, alto - 3),
      esquinaInfDer: pixel(ancho - 3, alto - 3),
      centro: pixel(Math.floor(ancho / 2), Math.floor(alto / 2)),
      bordeArriba: mediaBorde('arriba'),
      bordeAbajo: mediaBorde('abajo'),
      bordeIzquierda: mediaBorde('izquierda'),
      bordeDerecha: mediaBorde('derecha'),
      mediaGeneral: `rgb(${Math.round(r / n)}, ${Math.round(g / n)}, ${Math.round(b / n)})`,
    };
  });

  console.log('=== EL LOGO ===');
  for (const [clave, valor] of Object.entries(datos)) {
    console.log(`  ${clave}: ${valor}`);
  }
});
