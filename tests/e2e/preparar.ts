import { request } from '@playwright/test';
import { FABRICA, ponerElMetodo } from './ayudas';

/**
 * Lo que se hace **una vez, antes de la primera prueba**.
 *
 * Y es una sola cosa, pero importante: **dejar la instalación entrando por cuentas de la aplicación**.
 * Casi todos los casos entran con cuentas de prueba locales, y la instalación **entra por un método a
 * la vez** (`docs/usuarios-y-permisos.md`): si una pasada anterior se cortó a medias con el método en
 * `ad` o en `keycloak`, la siguiente empieza con cincuenta casos fallando por «credenciales
 * incorrectas», que no dice nada de lo que de verdad pasó. Pasó el 2026-09-27, y costó una hora de
 * confusión.
 *
 * Se hace con la **cuenta de fábrica**, que entra siempre sea cual sea el método: es la puerta que
 * permite volver a cambiarlo, y por eso es la única que puede hacer esto.
 *
 * Si no hay `ADMIN_PASSWORD` no se hace nada: sin ella no se puede entrar ni cambiar nada, y los casos
 * que la necesitan se saltan por su cuenta.
 */
export default async function preparar(): Promise<void> {
  if (!FABRICA.password) {
    return;
  }

  const peticion = await request.newContext({
    baseURL: process.env['BASE_URL'] ?? 'https://dev.catalina-support.example.com',
  });

  try {
    await ponerElMetodo(peticion, 'local');
    console.log('preparación: la instalación entra por cuentas de la aplicación');
  } catch (error) {
    // **No se tumba la pasada por esto**: puede que se esté probando contra una instalación que no es
    // la de desarrollo, y entonces lo que falla es la preparación, no el producto. Se dice y se sigue.
    console.warn('preparación: no se ha podido dejar el método en local:', error);
  } finally {
    await peticion.dispose();
  }
}
