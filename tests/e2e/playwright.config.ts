import { defineConfig, devices } from '@playwright/test';

/**
 * Pruebas de interfaz contra el entorno de desarrollo.
 *
 * **Corren contra desarrollo, no contra un servidor de mentira**: la dirección sale de `BASE_URL`, y
 * por defecto es la de desarrollo a través de nginx, que es exactamente lo que usa una persona
 * (docs/ambientes.md, sección 9.3). Así lo que se prueba incluye nginx, el backend y la base de
 * datos, no sólo el navegador.
 *
 * Se ejecutan **dentro de un contenedor** (`e2e` en `tests.yml`, con los navegadores ya instalados):
 * la máquina no tiene Node ni debe tenerlo (AGENTS.md, reglas para agentes).
 */
export default defineConfig({
  testDir: './specs',
  outputDir:
    process.env['ISOLATED_TESTS'] === '1'
      ? '/resultados/test-results'
      : 'test-results',

  // Lo que se hace una vez antes de empezar: **devolver la instalación a cuentas locales**, para que
  // una pasada cortada a medias no deje entrando por AD a la siguiente (ver `preparar.ts`).
  globalSetup: './preparar.ts',
  // Y al terminar, las cuentas de prueba se apagan (docs/ambientes.md, sección 9).
  globalTeardown: './limpiar.ts',

  // Un recorrido completo vale más que veinte pruebas de detalle: no se paraleliza de más para que
  // dos casos no se pisen los datos (los dos escriben en la misma base desechable).
  fullyParallel: false,
  workers: 1,

  // Si alguien se dejó un `test.only`, en integración continua tiene que fallar, no pasar en
  // silencio con el resto sin ejecutar.
  forbidOnly: !!process.env['CI'],

  retries: 0,

  // Desarrollo recarga en caliente y el backend vive en la misma máquina: 5 s por aserción es
  // justo cuando el servidor de Angular está reconstruyendo.
  expect: { timeout: 10_000 },

  reporter: [
    ['list'],
    [
      'html',
      {
        open: 'never',
        outputFolder:
          process.env['ISOLATED_TESTS'] === '1'
            ? '/resultados/informe'
            : 'informe',
      },
    ],
  ],

  // Las capturas y los vídeos de lo que falla son la mitad del valor de esta capa: sin ellos, un
  // fallo en un navegador que no ves es un mensaje de error y nada más.
  use: {
    // El navegador del usuario considera localhost seguro. En Docker usamos el alias frontend;
    // habilitamos esa misma capacidad sólo para este origen del navegador aislado, sin simular clipboard.
    launchOptions:
      process.env['ISOLATED_TESTS'] === '1'
        ? {
            args: [
              '--host-resolver-rules=MAP frontend.localhost frontend',
            ],
          }
        : {},
    baseURL: process.env['BASE_URL'] ?? 'http://frontend.localhost:11001',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    locale: 'es-ES',
    timezoneId: 'America/Bogota',
    // El producto tiene que funcionar en los tres anchos (docs/interfaz-y-experiencia.md, sección 7).
    viewport: { width: 1280, height: 800 },
  },

  projects: [
    {
      name: 'pc',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      // El mismo recorrido en un móvil: la aplicación tiene que **funcionar** en los tres anchos, no
      // sólo caber (docs/interfaz-y-experiencia.md, sección 7). En PC manda el ratón; aquí, el dedo.
      name: 'movil',
      use: { ...devices['Pixel 7'] },
    },
  ],
});
