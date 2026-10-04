import { request } from '@playwright/test';

import { FABRICA } from './ayudas';

/**
 * **Al terminar la pasada, las cuentas de prueba se apagan.**
 *
 * Cada caso da de alta las cuentas que necesita —es lo que hace que pruebe el camino de verdad y no
 * una suposición—, y hasta ahora **se quedaban encendidas en el entorno de desarrollo**: medido el
 * 2026-09-28, seiscientas personas activas de soporte y desarrollo **llamadas todas igual**, y el
 * buscador de a quién etiquetar y el desplegable de «Asignar a» convertidos en una lista inservible
 * para quien mira el entorno con los ojos (lo reportó el responsable).
 *
 * Se apagan —**no se borran**, que es la regla de las cuentas— pidiendo **por la API**, que es lo que
 * usa la propia aplicación: el contenedor de las pruebas no habla con la base ni conoce sus tablas.
 * Al apagarlas salen del filtro de activas, así que basta con pedir las que quedan y repetir hasta que
 * no haya ninguna.
 *
 * Los **tickets** de prueba no se tocan: un ticket no se borra (decisión del producto), y el entorno
 * se deja como nuevo con `./scripts/dev-seed.sh`, que borra los tickets y las cuentas de prueba y
 * vuelve a poner las once de ejemplo (`docs/ambientes.md`, sección 9).
 */
export default async function limpiar(): Promise<void> {
  const base = process.env['BASE_URL'] ?? 'http://frontend.localhost:11001';
  const peticion = await request.newContext({ baseURL: base });

  try {
    const entrada = await peticion.post('/api/auth/login', {
      data: { email: FABRICA.email, password: FABRICA.password },
    });

    const { token } = (await entrada.json()) as { token?: string };

    if (!token) {
      // Sin la cuenta de fábrica no se puede limpiar. Se dice y no se tumba la pasada: el resultado de
      // las pruebas ya está dado.
      console.warn('No se ha podido entrar con la cuenta de fábrica: no se apagan las cuentas de prueba.');

      return;
    }

    const cabecera = { Authorization: `Bearer ${token}` };

    // Cada vuelta apaga una tanda; como las apagadas salen del filtro, el bucle termina solo. El tope de
    // vueltas está para no quedarse dando vueltas si algo va mal.
    for (let vueltas = 0; vueltas < 40; vueltas++) {
      const listado = await peticion.get('/api/users?q=e2e-&active=true&perPage=50', {
        headers: cabecera,
      });

      const { users } = (await listado.json()) as { users?: { id: number }[] };

      if (!users?.length) {
        return;
      }

      await Promise.all(
        users.map((cuenta) =>
          peticion.post(`/api/users/${cuenta.id}/deactivate`, { headers: cabecera, data: {} }),
        ),
      );
    }
  } finally {
    await peticion.dispose();
  }
}
