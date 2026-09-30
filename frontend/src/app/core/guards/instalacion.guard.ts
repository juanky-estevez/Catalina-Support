import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { SetupService } from '../services/setup.service';

/**
 * Las dos guardas del primer arranque, y son **la pieza más importante de todo el asistente**.
 *
 * La instalación tiene **un sello**: mientras no esté puesto, la aplicación no tiene puerta —no hay
 * ninguna cuenta con la que entrar, ni configuración que enseñar— y lo único que tiene sentido
 * hacer es configurarla. En cuanto el sello está puesto, el asistente ya no vuelve: la
 * configuración se cambia desde Configuración, no repitiendo el primer arranque
 * (`docs/primer-arranque.md`, secciones 2 y 6).
 *
 * Las dos preguntan **lo mismo** —si la instalación está sellada— a través de `SetupService`, que lo
 * pregunta al backend **una sola vez y lo recuerda**. Así una navegación normal no dispara una
 * petición por pantalla.
 */

/**
 * Sin la instalación sellada, **cualquier pantalla de la aplicación lleva al asistente**.
 *
 * Va en la ruta padre de todo: la entrada, olvidar la contraseña, establecerla y el armazón entero.
 * No se deja ni la pantalla de entrada, porque antes de configurar la puerta no hay a quién dejar
 * entrar.
 */
export const instalacionComprobada: CanActivateFn = async () => {
  const setup = inject(SetupService);
  const router = inject(Router);

  return (await setup.estaInstalada()) ? true : router.createUrlTree(['/setup']);
};

/**
 * Con la instalación ya sellada, **`/setup` lleva a la entrada**.
 *
 * Se entra por la derecha y se deja el aviso de que ya está terminada, para que quien llegue con la
 * dirección a mano sepa por qué no ve el asistente en vez de encontrarse una pantalla de entrada
 * sin explicación.
 */
export const instalacionSinSellar: CanActivateFn = async () => {
  const setup = inject(SetupService);
  const router = inject(Router);

  return (await setup.estaInstalada())
    ? router.createUrlTree(['/login'], { queryParams: { installed: '1' } })
    : true;
};
