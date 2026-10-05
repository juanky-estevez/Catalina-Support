import { expect, test } from '@playwright/test';

import { crearCuentaLista, entrarComo, esperarMensaje, vaciarBuzon } from '../ayudas';

/**
 * El editor de los correos: la pantalla del módulo `mail` a la que Configuración enlaza
 * (`docs/modules/mail.md`, decisión 15, y `docs/interfaz-y-experiencia.md`, sección 3.8).
 *
 * El recorrido es el de verdad: se entra como Administrador, se llega desde Configuración, se cambia
 * el texto, **se ve la vista previa que renderiza el backend** y se manda una prueba que se lee en el
 * buzón. Y se deja el texto como estaba, porque las plantillas son de la instalación y no de la
 * prueba.
 */

/** La marca única de esta pasada: en desarrollo, las plantillas se comparten entre ejecuciones. */
const marca = () => Date.now().toString(36);

test.describe('El editor de los correos', () => {
  test.use({ permissions: ['clipboard-read', 'clipboard-write'] });

  test('el Administrador llega desde Configuración, edita, ve la vista previa y se manda una prueba', async ({
    page,
    request,
  }, info) => {
    // **Una cuenta de administrador y no la de fábrica**: la prueba se manda al correo de quien la
    // pide, y la cuenta de fábrica es la única sin correo (responde `mail.test.noEmail`).
    const jefa = await crearCuentaLista(request, { role: 'administrador' });
    await vaciarBuzon(request);

    await entrarComo(page, jefa.email);

    // Se llega desde Configuración, que es donde vive el enlace: la pantalla es de otro módulo.
    await page.goto('/settings');
    await page.getByRole('link', { name: 'Editar los correos' }).click();
    await expect(page).toHaveURL(/\/mail$/);
    await expect(page.getByRole('heading', { name: 'Los correos' })).toBeVisible();

    // **Y la vuelta a Configuración es el botón de atrás** (decisión del responsable, 2026-09-29): la
    // flecha a la izquierda, como en las listas de tickets, con su nombre accesible.
    const volver = page.getByRole('link', { name: 'Volver' });
    const titulo = page.getByRole('heading', { name: 'Los correos' });
    await expect(volver).toBeVisible();
    expect(
      await volver.evaluate((enlace) => enlace.parentElement?.firstElementChild === enlace),
    ).toBe(true);
    const [cajaDeVolver, cajaDelTitulo] = await Promise.all([
      volver.boundingBox(),
      titulo.boundingBox(),
    ]);
    expect(cajaDeVolver).not.toBeNull();
    expect(cajaDelTitulo).not.toBeNull();
    expect(cajaDeVolver!.x).toBeLessThan(cajaDelTitulo!.x);
    await volver.click();
    await expect(page).toHaveURL(/\/settings$/);
    await page.getByRole('link', { name: 'Editar los correos' }).click();
    await expect(page).toHaveURL(/\/mail$/);

    // Los **once** correos, y los dos idiomas del elegido. Eran diez hasta el 2026-09-27, que entró el
    // aviso de que te han etiquetado (`docs/modules/mail.md`, sección 5.1).
    await expect(page.getByRole('navigation', { name: 'Qué correo' }).getByRole('button')).toHaveCount(11);

    const texto = marca();
    const asunto = `Aviso de prueba ${texto}: {{numero}}`;
    const cuerpo = `<p>Hola {{nombre}}, esto es una prueba ${texto}.</p>`;

    // Se elige el correo de «Ticket resuelto» y se escriben las dos cosas.
    await page.getByRole('button', { name: 'Ticket resuelto, al usuario' }).click();
    await page.getByLabel('Asunto').first().fill(asunto);
    await page.getByLabel('Cuerpo').first().fill(cuerpo);

    // Los marcadores se insertan **al pulsarlos**, en el punto donde está el cursor.
    await page.getByRole('button', { name: '{{enlace}}' }).first().click();

    // La vista previa la renderiza el backend: enseña el texto con los datos de ejemplo ya puestos.
    await expect(page.getByText(`esto es una prueba ${texto}`)).toBeVisible();
    await expect(page.getByText('Ana', { exact: false }).first()).toBeVisible();

    await page.getByRole('button', { name: 'Guardar' }).first().click();
    await expect(page.getByText('Guardado. La vista previa ya enseña el texto nuevo.')).toBeVisible();

    // Y la prueba sale de verdad: se lee en el buzón, con el asunto de prueba delante.
    await page.getByRole('button', { name: 'Enviarme una prueba' }).first().click();
    await expect(page.getByText(`Prueba enviada a ${jefa.email}`)).toBeVisible();

    const correo = await esperarMensaje(request, jefa.email, `[prueba] Aviso de prueba ${texto}: ACME-2026-0042`);
    expect(correo).toContain(`esto es una prueba ${texto}`);
    // El enlace del marcador va compuesto por el backend, no escrito en el texto.
    expect(correo).toContain('http');

    // Y se deja como estaba: las plantillas son de la instalación, no de la prueba.
    await page.getByRole('button', { name: 'Volver al de fábrica' }).first().click();
    await expect(page.getByText('¿Volver al texto de fábrica?')).toBeVisible();
    await page.getByRole('button', { name: 'Sí, volver' }).click();
    await expect(page.getByText('Este correo ha vuelto al texto de fábrica.')).toBeVisible();

    // Las dos columnas: al lado en PC y apiladas en móvil, que es lo que dice la sección 3.8.
    const espanol = await page.getByRole('region', { name: 'Español' }).boundingBox();
    const ingles = await page.getByRole('region', { name: 'English' }).boundingBox();

    expect(espanol).not.toBeNull();
    expect(ingles).not.toBeNull();

    if (info.project.name === 'pc') {
      expect(ingles!.x).toBeGreaterThan(espanol!.x + espanol!.width - 1);
    } else {
      expect(ingles!.y).toBeGreaterThan(espanol!.y + espanol!.height - 1);
    }
  });

  test('un marcador que ese correo no admite se avisa **antes** de guardar', async ({
    page,
    request,
  }) => {
    const jefa = await crearCuentaLista(request, { role: 'administrador' });
    await entrarComo(page, jefa.email);
    await page.goto('/mail');

    await page.getByRole('button', { name: 'Ticket resuelto, al usuario' }).click();

    // El marcador que no existe: la vista previa lo dice con la clave traducida, sin haber guardado
    // nada (docs/modules/mail.md, sección 4).
    await page.getByLabel('Cuerpo').first().fill('<p>Hola {{inventado}}</p>');

    await expect(page.getByText('El texto usa un marcador que ese correo no admite.')).toBeVisible();

    // Y no se ha guardado: el texto del backend sigue siendo el que estaba.
    const plantilla = await request.get('/api/mail/templates', {
      headers: { Authorization: `Bearer ${await tokenDe(request, jefa.email)}` },
    });
    const { templates } = (await plantilla.json()) as {
      templates: readonly { key: string; language: string; body: string }[];
    };
    const original = templates.find((t) => t.key === 'ticket.resolved' && t.language === 'es');

    expect(original?.body).not.toContain('inventado');
  });
});

/** El token de una cuenta, para comprobar por la API lo que la pantalla ha hecho. */
async function tokenDe(
  request: Parameters<typeof crearCuentaLista>[0],
  email: string,
): Promise<string> {
  const respuesta = await request.post('/api/auth/login', {
    data: { email, password: 'una-contraseña-larga' },
  });
  const cuerpo = (await respuesta.json()) as { token: string };

  return cuerpo.token;
}
