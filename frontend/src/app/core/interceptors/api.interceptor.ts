import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, tap, throwError } from 'rxjs';

import { AvailabilityService } from '../services/availability.service';
import { SessionService } from '../services/session.service';

/**
 * Todo lo que se le hace a cada petición, en un solo sitio.
 *
 * Tres cosas, y ninguna en las pantallas:
 *
 * 1. **Pone la cabecera `Authorization`** con el token. El navegador no manda nada por su cuenta:
 *    la sesión viaja en una cabecera explícita (docs/usuarios-y-permisos.md, sección 6).
 * 2. **Atiende el 401**: borra la sesión y lleva a la pantalla de entrada. Sólo cuando la petición
 *    llevaba token —si no, el 401 es «esas credenciales no valen» y lo tiene que contar la propia
 *    pantalla, que es donde está el formulario—.
 * 3. **Marca si el servidor no está**: una petición sin respuesta o un 502/503 son el estado
 *    «servidor caído» del armazón.
 *
 * Está registrado como interceptor de función, que es la forma moderna de Angular: sin clases y
 * sin `NgModule` (docs/arquitectura.md, sección 6).
 */
/**
 * Las peticiones que pueden contestar 401 sin que la sesión tenga nada que ver.
 *
 * Entrar, pedir el enlace y establecer la contraseña responden 401 cuando lo que no vale son las
 * credenciales o el enlace del correo —«esas credenciales no son», «ese enlace ya no vale»—, y eso
 * lo tiene que contar la pantalla que lo pidió. **Un 401 de aquí no puede echar a nadie de su
 * sesión**: si alguien con sesión abre un enlace caducado, lo que pasa es que el enlace caducó, no
 * que su sesión se haya muerto.
 */
const RUTAS_QUE_NO_ECHAN_A_NADIE = [
  '/api/auth/login',
  '/api/auth/password/forgot',
  '/api/auth/password/reset',
];

export const apiInterceptor: HttpInterceptorFn = (peticion, siguiente) => {
  const sesion = inject(SessionService);
  const router = inject(Router);
  const disponibilidad = inject(AvailabilityService);

  const token = sesion.token();
  const conSesion = token !== null;
  const puedeEchar = !RUTAS_QUE_NO_ECHAN_A_NADIE.some((ruta) => peticion.url.includes(ruta));

  const conCabecera = conSesion
    ? peticion.clone({ setHeaders: { Authorization: `Bearer ${token}` } })
    : peticion;

  return siguiente(conCabecera).pipe(
    // Una respuesta buena es la noticia de que el servidor está: quita el aviso, si lo había.
    tap(() => disponibilidad.disponible()),
    catchError((error: unknown) => {
      if (error instanceof HttpErrorResponse) {
        // Sin respuesta del servidor (cortado) o un 502/503 de nginx: no hay con quién hablar.
        if (error.status === 0 || error.status === 502 || error.status === 503) {
          disponibilidad.noDisponible();
        }

        if (error.status === 401 && conSesion && puedeEchar) {
          // Caducó o la cuenta se desactivó: se borra la sesión y se lleva a la entrada, que es
          // quien explica lo que ha pasado.
          sesion.olvidar(true);
          void router.navigate(['/login']);
        }

        // Un 403 no se lleva a ninguna parte por su cuenta: quien lo pidió sabe si era una
        // pantalla entera (y enseña la de «sin permiso») o una acción suelta (y lo cuenta al lado).
      }

      return throwError(() => error);
    }),
  );
};
