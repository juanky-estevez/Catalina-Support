# Usuarios: el módulo

> **Estado:** propuesta
> **Última actualización:** 2026-09-22

## 1. Alcance de este documento

Cuenta **cómo se construye el módulo `users`**: la tabla de cuentas, los endpoints con sus formas de
petición y respuesta, las reglas del alta, el cambio de papel, la desactivación y el perfil propio.

**No repite** lo que ya está decidido en otro sitio:

| Qué | Dónde está |
| --- | --- |
| Qué puede hacer cada papel | `docs/usuarios-y-permisos.md` (la matriz) |
| Cómo se entra, sesión y contraseñas | `docs/usuarios-y-permisos.md` y, con detalle, `docs/autenticación.md` |
| Las pantallas del módulo | `docs/interfaz-y-experiencia.md` |
| La lista de endpoints, en resumen | `docs/tickets.md`, sección 5 |

Es una **propuesta**: no habilita escribir código hasta que esté aprobada (Regla 0). La sección 11
recoge lo que he propuesto yo.

## 2. La tabla de cuentas

Una sola tabla, `users`, con nombres de columna en inglés y **valores en español**, que es el mismo
criterio que ya siguen los estados de los tickets.

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `bigserial` | |
| `name` | `text` | Nombre de pila. Obligatorio |
| `last_name` | `text` | Apellidos, en un solo campo. Obligatorio |
| `email` | `text` | Obligatorio y **único sin distinguir mayúsculas** |
| `password_hash` | `text` | **Nulo** en las cuentas de directorio: ahí la contraseña no es nuestra |
| `role` | `text` | `usuario`, `soporte`, `desarrollo` o `administrador` |
| `origin` | `text` | `local`, `ad` o `keycloak` |
| `external_id` | `text` | El identificador del directorio (`sub` de Keycloak o GUID/cuenta de AD). Nulo en las locales |
| `language` | `text` | `es` o `en`. Es lo que decide **en qué idioma se le escriben los correos** |
| `is_active` | `boolean` | Verdadero al nacer |
| `last_login_at` | `timestamptz` | Cuándo entró por última vez |
| `created_at`, `updated_at` | `timestamptz` | |

**Restricciones** (en la base, no sólo en el código):

- `email` **único sobre `lower(email)`**: `Ana@…` y `ana@…` son la misma persona.
- `role` y `origin` sólo admiten sus valores; una errata en un papel no puede llegar a la base.
- **`password_hash` va y viene con el origen**: obligatorio si el origen es `local`, y **nulo** si es
  `ad` o `keycloak`. Así la regla «si el origen es de directorio, no hay contraseña local» no depende
  de que el código se acuerde.
- **`external_id` único por origen** cuando no es nulo: dos cuentas no pueden apuntar al mismo
  usuario del directorio.

**No hay `deleted_at`**: la baja no existe (`docs/usuarios-y-permisos.md`, sección 9). Una cuenta se
desactiva y sigue siendo legible por los tickets que tenga.

**El administrador de fábrica no está aquí.** Es la única cuenta que no tiene fila: vive en la
configuración (`docs/usuarios-y-permisos.md`, sección 8). Por eso esta tabla **no necesita una
columna de usuario**: todo el mundo entra con su correo.

## 3. Cómo se identifica a alguien

- El **correo** es el identificador de las personas, y es también con lo que se casa una cuenta local
  con la del directorio.
- El **administrador de fábrica** se identifica con `admin`, y no es una fila de esta tabla.
- Nadie más tiene usuario ni alias: si dos personas comparten correo, no son dos personas.

## 4. Los endpoints

Estos son los del módulo, y son los únicos: la lista de `auth` vive en `docs/autenticación.md`, y
`docs/tickets.md` se queda con los de tickets.

| Método y ruta | Qué hace | Quién |
| --- | --- | --- |
| `GET /api/users` | Lista paginada, con filtros por papel, origen y estado, y búsqueda por nombre o correo | Soporte y Administrador |
| `POST /api/users` | Da de alta una cuenta | Soporte (sólo `usuario`) y Administrador |
| `GET /api/users/{id}` | El detalle de una cuenta | Administrador |
| `PATCH /api/users/{id}` | Cambia nombre, apellidos, papel, correo, idioma o estado | Administrador; Soporte sólo el nombre y los apellidos |
| `PATCH /api/users/me` | Cambia **tu** nombre, tus apellidos y tu idioma | Cualquier cuenta autenticada |
| `POST /api/users/{id}/deactivate` | Desactiva | Soporte y Administrador |
| `POST /api/users/{id}/activate` | Reactiva. En cuentas de directorio, sólo si el directorio dice que está activa | Soporte y Administrador |
| `POST /api/users/{id}/origin` | Cambia el origen de la cuenta | Administrador |
| `POST /api/users/{id}/reset-password` | Lanza un reseteo: el usuario recibe el correo | Soporte y Administrador |

- **Las acciones que cambian de estado tienen su propia ruta** (`/deactivate`, `/activate`) en vez de
  ser un `PATCH` con un campo: así no se puede desactivar a alguien por accidente al editar su
  nombre, y queda claro en el registro de peticiones qué pasó.
- **`/origin` y `/reset-password` no existirían si no fueran peligrosas**: son las dos únicas
  acciones que pueden dejar a alguien sin poder entrar, y por eso tienen su ruta y su permiso.
- El **detalle** de una cuenta sólo lo ve un Administrador: Soporte necesita la lista para buscar a
  quien reporta, no la ficha de nadie.

## 5. Las reglas del alta

1. **Obligatorios**: nombre, apellidos, correo, papel y origen. Los dos primeros porque son personas
   y quien da de alta las conoce; el correo, porque es con lo que entra.
2. **El correo se valida** en forma y en unicidad, sin distinguir mayúsculas. Si ya existe, el error
   lo dice con claridad: es un alta, no un misterio.
3. **Según el origen**:
   - `local`: se genera el enlace de alta y se le manda por correo. La cuenta **no puede entrar hasta
     que establece la contraseña**, porque no tiene ninguna.
   - `ad` o `keycloak`: **se comprueba que la persona existe en el directorio** antes de crear nada.
     Si el directorio no la conoce, no se crea la cuenta. No se manda ningún correo: su contraseña es
     del dominio.
4. **Quién puede dar de alta qué**: Soporte sólo crea cuentas con papel `usuario`; repartir `soporte`,
   `desarrollo` o `administrador` es cosa de un Administrador
   (`docs/usuarios-y-permisos.md`, sección 3).
5. **El alta y el correo los hacen dos módulos.** `users` crea la cuenta; el enlace y el correo los
   emite `auth`, que es quien sabe de contraseñas y de tokens. `users` no toca las tablas de `auth`:
   se lo pide a su servicio.
6. **Si el correo de alta no sale**, la cuenta existe y se puede lanzar otro:
   `POST /api/users/{id}/reset-password` manda el mismo enlace. Es la salida para quien no recibió el
   primero, y no hace falta borrar y volver a crear nada.

## 6. Cambiar el papel

- Lo hace un **Administrador**, y es una acción explícita con su propio registro.
- **Cambia al instante**: el papel no viaja dentro del token, se lee de la cuenta en cada petición.
- Los cuatro papeles siguen siendo los mismos: lo que cambia es cuál tiene una persona. No se crean
  papeles nuevos ni se editan permisos desde la aplicación.

## 7. Desactivar y reactivar

- Lo hacen **Soporte y Administrador**. Desactivar no borra nada: la cuenta deja de poder entrar y
  los tickets que tenga siguen contando su historia con su nombre.
- **A una cuenta de directorio no se le quita el acceso desde aquí** (regla 6 de
  `docs/usuarios-y-permisos.md`): si sigue activa en AD o Keycloak, vuelve a entrar y se reactiva
  sola. La interfaz lo dice al desactivar, para que nadie se quede con la sensación de haberlo hecho.
- **Reactivar una cuenta de directorio comprueba primero con el directorio**: si allí está
  desactivada, no se reactiva aquí.
- **Desactivar bloquea al instante**: la siguiente petición de esa cuenta ya no vale, aunque su
  token siga siendo válido hasta caducar. Lo que no hace es borrar el token del navegador de esa
  persona: quien intente algo recibirá un aviso y saldrá a la pantalla de entrada.

## 8. El perfil propio

- Cualquiera puede cambiar **su nombre, sus apellidos y su idioma** desde `PATCH /api/users/me`. Nada
  más: ni su correo, ni su papel, ni su origen, que son decisiones de otro.
- **Cambiar el idioma cambia el de los correos que recibe**: es el único ajuste personal que el
  servidor necesita saber, y por eso vive en la cuenta y no en el navegador. El tema, en cambio, se
  queda en el navegador, porque al servidor le da igual.
- El perfil se abre desde el menú lateral, con los demás controles
  (`docs/interfaz-y-experiencia.md`, sección 3.1).

## 9. Los errores del módulo

Con claves, que el frontend traduce (`docs/interfaz-y-experiencia.md`, sección 8):

| Clave | Cuándo |
| --- | --- |
| `users.email.duplicated` | El correo ya está en uso |
| `users.email.invalid` | El correo no tiene forma de correo |
| `users.name.required` | Falta el nombre o los apellidos |
| `users.role.notAllowed` | Soporte intenta crear un papel que no le toca |
| `users.directory.notFound` | Se intenta dar de alta a alguien que el directorio no conoce |
| `users.origin.inUse` | No se puede cambiar el origen de una cuenta de directorio activa |
| `users.notFound` | La cuenta no existe |

## 10. Los correos del módulo

`users` **no manda correos**: los pide a `auth`. Los dos que puede provocar son el del alta y el del
reseteo, y los dos llevan el mismo enlace. El texto va en el idioma de la cuenta
(`docs/arquitectura.md`, sección 9).

## 11. Decisiones que hay que confirmar

| # | Decisión | Propuesta |
| --- | --- | --- |
| 1 | **`last_login_at`** | Sí: saber cuándo entró alguien por última vez es lo que permite decidir con criterio a quién desactivar |
| 2 | **El detalle de una cuenta** | Sólo Administrador. Soporte ve la lista (nombre, apellidos y correo), que es lo que necesita para buscar a quien reporta |
| 3 | **Un correo, una cuenta** | Sí, y la comparación ignora mayúsculas |

Lo del token y la pantalla de puesta en marcha ya está decidido y aplicado arriba y en
`docs/interfaz-y-experiencia.md`.

## 12. Qué habilita este documento

Con `docs/usuarios.md` aprobado, el siguiente es **`docs/autenticación.md`**, que ya tiene tomadas
sus decisiones (token sin estado de 10 horas, AD por búsqueda con cuenta de servicio, enlaces con
huella y un solo uso, contenedores de pruebas para el directorio). Después, `users` y `auth` se
implementan juntos: no se puede tener cuentas sin poder entrar, ni entrar sin cuentas.
