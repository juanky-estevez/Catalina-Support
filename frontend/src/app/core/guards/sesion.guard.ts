import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { SessionService } from '../services/session.service';
import { SetupService } from '../services/setup.service';

/**
 * Sin sesión no se entra: a la pantalla de entrada.
 *
 * La guarda **no decide permisos**, sólo si hay alguien dentro: ocultar un botón no es una medida
 * de seguridad, y la decisión de qué puede cada papel se toma en el backend
 * (docs/usuarios-y-permisos.md, sección 10).
 *
 * Antes de decidir, se comprueba la sesión contra el backend: un token vencido en el navegador no
 * es una sesión, y dejar pasar a alguien para que el backend lo rechace en la primera llamada es
 * enseñarle una pantalla vacía un segundo antes de echarlo.
 */
export const sesionActiva: CanActivateFn = async (_route, state) => {
  const sesion = inject(SessionService);
  const setup = inject(SetupService);
  const router = inject(Router);

  if (!sesion.hayToken()) {
    return router.createUrlTree(['/login']);
  }

  if (!sesion.comprobadaLaSesion()) {
    const usuario = await sesion.comprobar();
    if (!usuario) {
      return router.createUrlTree(['/login']);
    }
  }

  const usuario = sesion.usuario();
  if ((await setup.estadoDeInstalacion()).aiRequired && state.url !== '/settings' && state.url !== '/forbidden') {
    return usuario?.role === 'administrador'
      ? router.createUrlTree(['/settings'], { queryParams: { ai: 'required' } })
      : router.createUrlTree(['/forbidden'], { queryParams: { ai: 'required' } });
  }

  return true;
};

/**
 * Sólo para quien ha entrado **y** tiene uno de esos papeles.
 *
 * Es comodidad, no seguridad: quien llegue por su cuenta se encuentra con el 403 del backend.
 */
export const conPapel = (...papeles: readonly string[]): CanActivateFn => {
  return async () => {
    const sesion = inject(SessionService);
    const router = inject(Router);

    if (!sesion.hayToken()) {
      return router.createUrlTree(['/login']);
    }

    const usuario = sesion.usuario() ?? (await sesion.comprobar());
    if (!usuario) {
      return router.createUrlTree(['/login']);
    }

    return papeles.includes(usuario.role) ? true : router.createUrlTree(['/forbidden']);
  };
};

/**
 * Dónde entra cada papel: **la raíz no es una pantalla, es un reparto**.
 *
 * El Administrador aterriza en la lista de usuarios, que es lo que va a tocar
 * (`docs/interfaz-y-experiencia.md`, sección 4.4), y los otros tres en su bandeja. Antes había una
 * pantalla de inicio provisional; desapareció con las pantallas de producto, que era lo que le
 * faltaba a los cuatro papeles.
 */
export const inicioSegunPapel: CanActivateFn = async () => {
  const sesion = inject(SessionService);
  const router = inject(Router);

  if (!sesion.hayToken()) {
    return router.createUrlTree(['/login']);
  }

  const usuario = sesion.usuario() ?? (await sesion.comprobar());
  if (!usuario) {
    return router.createUrlTree(['/login']);
  }

  // El Administrador a los usuarios; el usuario, Soporte y Desarrollo, a su bandeja.
  return router.createUrlTree([usuario.role === 'administrador' ? '/users' : '/tickets']);
};

/**
 * Sólo para quien ha entrado **y tiene cuenta en la tabla**.
 *
 * **La cuenta de fábrica no tiene perfil** (`docs/modules/users.md`, sección 8): no está en la tabla
 * de cuentas, así que no hay fila que cambiar y el backend responde 404. Aquí se le manda a la
 * pantalla que explica que eso no es suyo, en vez de enseñarle un formulario que no puede guardar.
 */
export const conCuentaPropia: CanActivateFn = async () => {
  const sesion = inject(SessionService);
  const router = inject(Router);

  if (!sesion.hayToken()) {
    return router.createUrlTree(['/login']);
  }

  const usuario = sesion.usuario() ?? (await sesion.comprobar());
  if (!usuario) {
    return router.createUrlTree(['/login']);
  }

  return usuario.factory ? router.createUrlTree(['/forbidden']) : true;
};
