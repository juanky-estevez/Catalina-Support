# Usuarios: el módulo

> **Estado:** as-built
> **Última actualización:** 2026-10-07
>
> **Enmendado el 2026-10-07**, conforme a la propuesta aprobada del 2026-10-06: `users.language`
> desaparece de la migración, el modelo, las API y las pantallas. El perfil conserva nombre y
> apellidos; los correos usan el idioma global. Backend y frontend quedaron verificados.
>
> **Pasa a as-built el 2026-09-26**: el módulo **está entero** —backend, las tres pantallas y las tres
> acciones que preguntan al directorio, con las cinco enmiendas de abajo—, así que este documento
> describe lo que el código hace. La cabecera se quedó en `aprobado` por un descuido de forma, no
> porque faltara nada.
>
> **Enmendado el 2026-09-25**: las tres acciones que dependían del directorio —las que
> el responsable decidió ese mismo día y quedaron sin hacer— **están hechas**, con el alcance que él
> fijó: **reactivar una cuenta de AD pregunta al directorio** si sigue conociendo a esa persona, y
> **la de Keycloak no se reactiva a mano** (se cuenta que vuelve sola al entrar, porque preguntarlo
> pediría una cuenta de servicio en el reino); y **el alta y el cambio de origen hacia el directorio
> no se hacen a mano**, con un mensaje que dice eso en vez de hablar del directorio. Con esto, **este
> documento no tiene nada pendiente**.
>
> Aprobado por el responsable el 2026-09-23, tras repasarlo en forma de preguntas. Fija
> el módulo `users`: la tabla de cuentas, los endpoints, las reglas del alta, el cambio de
> papel, la desactivación y el perfil propio.
>
> Se implementa junto con el otro: no se puede tener cuentas sin poder entrar, ni entrar sin cuentas.
>
> **Enmendado el 2026-09-25**, al terminar el camino de AD de `auth`: el punto 4 de la sección 5 decía
> que sus tres rechazos se quitarían «en el mismo cambio que traiga la comprobación», y **la
> comprobación ya está**. El responsable decidió las tres el 2026-09-25: **reactivar una cuenta de
> directorio pregunta al directorio** (si lo encuentra, se reactiva), y **el alta a mano y el cambio
> de origen hacia `ad` no se hacen**, con un mensaje que diga eso en vez de hablar del directorio.
> Las tres quedan **decididas y sin hacer**, para implementarlas juntas en cuanto se retome el módulo
> (decisión del responsable: primero se cierra Keycloak). Con esto **queda cerrada la duda** de la
> sección 7 sobre si el aviso `users.directoryMayReturn` se podía ejercitar: se ejercita, y hay prueba
> de interfaz en `docs/ambientes.md`, sección 9.3.
>
> **Enmendado el 2026-09-24**, al terminar el módulo: la sección 9 estrena **`users.selfDeactivation`**
> (nadie se desactiva a sí mismo), la sección 8 fija que **la cuenta de fábrica no tiene perfil** y
> responde 404, y la sección 5 recoge que **todo lo que exige preguntarle al directorio** —el alta de
> cuentas de directorio, su reactivación y el cambio de origen hacia `ad` o `keycloak`— se rechaza
> mientras ese camino no exista. Las tres decisiones son del responsable, del 2026-09-24.
>
> **Enmendado el 2026-09-24**, al construir las pantallas, con cinco decisiones del
> responsable que salen de un hallazgo: **Soporte puede cambiar el nombre y los apellidos de otra
> cuenta, pero no ve la ficha de nadie** —decisión 2 de este documento—, así que ese permiso no tenía
> dónde ejercerse. Queda así: **Soporte los edita en línea, desde la propia lista** (decisión 15);
> **el alta es un diálogo** sobre la lista (16); **el origen se ofrece con sus tres valores, y `ad` y
> `keycloak` desactivados** con su nota (17); **la lista es tabla en PC y tarjetas en móvil, con chips
> de filtro** (18); y **la ficha y el perfil son dos rutas**, `/users/{id}` y `/profile` (19). Las
> cinco están escritas donde les toca, en `docs/interfaz-y-experiencia.md`, sección 3.6: **este
> documento cuenta el módulo, no cómo se ve**. Lo que sí se sostiene aquí es que **la decisión 2 no
> cambia** —Soporte sigue sin ver la ficha de nadie— y que su límite de edición pasa a ejercerse en la
> lista.
>
> **Estado de la implementación (2026-09-24): el módulo está entero**, backend y frontend. El alta, el
> reenvío del enlace, **la lista** con filtros y paginación, **la ficha**, **los cambios** (con el
> límite de Soporte), **el perfil propio**, **desactivar**, **reactivar** y **el cambio de origen**,
> probados de extremo a extremo, más **sus tres pantallas** —la lista, la ficha y el perfil—, que fija
> `docs/interfaz-y-experiencia.md`, sección 3.6. Lo que depende del directorio se rechaza mientras ese
> camino no exista (sección 5, punto 4). **Ese camino ya existe y ese punto está hecho** (quinta
> enmienda, más arriba).
>
> **Enmendado el 2026-09-23**, al implementar el camino local de `auth`: la tabla de errores no tenía
> **código HTTP** —igual que le pasaba a `mail.md`— ni ninguna clave para un origen o un idioma
> inventados. Se añadieron `users.origin.unknown` y la hoy retirada `users.language.unknown`, con sus códigos, aprobados
> por el responsable: 404 lo que no existe, 409 el correo repetido, 422 lo que no vale y **403
> `users.role.notAllowed`**, que es un permiso y no una errata.

## 1. Alcance de este documento

Cuenta **cómo se construye el módulo `users`**: la tabla de cuentas, los endpoints con sus formas de
petición y respuesta, las reglas del alta, el cambio de papel, la desactivación y el perfil propio.

**No repite** lo que ya está decidido en otro sitio:

| Qué | Dónde está |
| --- | --- |
| Qué puede hacer cada papel | `docs/usuarios-y-permisos.md` (la matriz) |
| Cómo se entra, sesión y contraseñas | `docs/usuarios-y-permisos.md` y, con detalle, `docs/modules/auth.md` |
| Las pantallas del módulo | `docs/interfaz-y-experiencia.md` |
| La lista de endpoints, en resumen | `docs/modules/tickets.md`, sección 5 |

Está **aprobado**, así que habilita escribir código (Regla 0). La sección 11
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

Estos son los del módulo, y son los únicos: la lista de `auth` vive en `docs/modules/auth.md`, y
`docs/modules/tickets.md` se queda con los de tickets.

| Método y ruta | Qué hace | Quién |
| --- | --- | --- |
| `GET /api/users` | Lista paginada, con filtros por papel, origen y estado, y búsqueda por nombre o correo | Soporte y Administrador |
| `POST /api/users` | Da de alta una cuenta | Soporte (sólo `usuario`) y Administrador |
| `GET /api/users/{id}` | El detalle de una cuenta | Administrador |
| `PATCH /api/users/{id}` | Cambia nombre, apellidos, papel, correo, idioma o estado | Administrador; Soporte sólo el nombre y los apellidos |
| `PATCH /api/users/me` | Cambia **tu** nombre, tus apellidos y tu idioma | Cualquier cuenta autenticada |
| `POST /api/users/{id}/deactivate` | Desactiva | Soporte y Administrador |
| `POST /api/users/{id}/activate` | Reactiva. En una cuenta de AD, sólo si el directorio la sigue conociendo; en una de Keycloak, no se hace desde aquí | Soporte y Administrador |
| `POST /api/users/{id}/origin` | Cambia el origen de la cuenta | Administrador |
| `POST /api/users/{id}/reset-password` | Lanza un reseteo: el usuario recibe el correo | Soporte y Administrador |

- **Las acciones que cambian de estado tienen su propia ruta** (`/deactivate`, `/activate`) en vez de
  ser un `PATCH` con un campo: así no se puede desactivar a alguien por accidente al editar su
  nombre, y queda claro en el registro de peticiones qué pasó. Por eso el `PATCH` **no** cambia el
  estado: si le llega, responde `users.stateHasItsOwnAction` y no toca nada. La tabla de arriba nombra
  el estado entre lo que cambia el `PATCH`; manda el motivo, y el estado va por sus acciones.
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
   - `ad` o `keycloak`: **no se dan de alta a mano**, y se contesta `users.origin.byDirectory`. Quien
     está en el directorio **entra solo** con su cuenta y la suya se crea, se vincula o se pone al día
     en ese primer acceso (`docs/modules/auth.md`, sección 5.4): crearla aquí antes de tiempo sería
     una cuenta con la contraseña de otro sitio y con unos datos que envejecen solos. **La pantalla ya
     no ofrece esos dos orígenes** (decisión 17), así que esto es la red que recoge a quien llame a la
     API a mano.
4. **Lo que le preguntamos al directorio**, y por qué no se le pregunta lo mismo en los dos
   caminos:

   | Acción | Qué se hace |
   | --- | --- |
   | **Reactivar** una cuenta de **AD** | **Pregunta al directorio** si sigue conociendo a esa persona, con la misma búsqueda con la que se entra. Si lo encuentra, se reactiva; si no aparece, `users.directory.notFound`. **Devolverle el acceso a quien ya no está allí sería devolverle un acceso que no puede usar**, porque su contraseña es de allí |
   | **Reactivar** una cuenta de **Keycloak** | **No se reactiva a mano**, y se dice por qué con `users.directory.activatesItself`: su camino **entra solo y al entrar se reactiva**. Preguntarle a Keycloak si esa persona sigue allí pediría una **cuenta de servicio con permiso de lectura dentro del reino**, y eso no se le pide a quien lo administra por un botón que se arregla solo (decisión del responsable, 2026-09-25) |
   | **Dar de alta** una cuenta `ad` o `keycloak` | **No se hace a mano** (`users.origin.byDirectory`): quien está en el directorio entra solo y su cuenta nace, se vincula o se pone al día en ese acceso |
   | **Cambiar el origen** de una cuenta a `ad` o `keycloak` | **No se hace a mano** (`users.origin.byDirectory`), por lo mismo: el origen de una cuenta de directorio lo pone el directorio, al entrar esa persona por su camino |

   El **cambio de origen hacia `local`** sí funciona: quita el vínculo con el directorio y, si la
   cuenta se queda sin contraseña, manda el enlace de alta.

   **Si no se puede comprobar** —el directorio no responde, o esta instalación no tiene ninguno
   **configurado**—, se contesta `users.directory.unavailable` (503) y el motivo queda en el log: el
   fallo no es de quien pide la reactivación.

   **Quién pregunta al directorio es `auth`** (enmienda del 2026-09-25): este módulo declara la
   pregunta que necesita —«¿el directorio sigue conociendo a esta persona?»— y la contesta `auth`, que
   es el módulo que habla el protocolo y **el que tiene la configuración guardada**. Así la
   comprobación se hace con la misma cuenta de servicio y el mismo filtro con los que se entra, y no
   hay dos copias de la configuración del directorio. La conexión se hace en `main.go`
   (`docs/modules/settings.md`, sección 5.8).

5. **A una cuenta desactivada no se le reenvía el enlace** (`users.accountInactive`): establecería
   una contraseña para seguir sin poder entrar. Primero se reactiva.
6. **Quién puede dar de alta qué**: Soporte sólo crea cuentas con papel `usuario`; repartir `soporte`,
   `desarrollo` o `administrador` es cosa de un Administrador
   (`docs/usuarios-y-permisos.md`, sección 3).
7. **El alta y el correo los hacen dos módulos.** `users` crea la cuenta; el enlace y el correo los
   emite `auth`, que es quien sabe de contraseñas y de tokens. `users` no toca las tablas de `auth`:
   se lo pide a su servicio.
8. **Si el correo de alta no sale**, la cuenta existe y se puede lanzar otro:
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
- **Reactivar una cuenta de AD comprueba antes con el directorio**: si ya no está allí, no se
  reactiva, porque su contraseña es de allí y no podría entrar. **La de Keycloak no se reactiva a
  mano**: su camino la reactiva sola al entrar, y la pantalla lo cuenta en vez de ofrecer el botón
  (punto 4 de la sección 5).
- **Nadie se desactiva a sí mismo** (`users.selfDeactivation`): desactivarse es quedarse fuera al
  instante y quien lo hace no puede volver, porque hace falta que otro le reactive. La pantalla
  tampoco ofrece ese botón sobre uno mismo (decisión del responsable, 2026-09-24).
- **Desactivar bloquea al instante**: la siguiente petición de esa cuenta ya no vale, aunque su
  token siga siendo válido hasta caducar. Lo que no hace es borrar el token del navegador de esa
  persona: quien intente algo recibirá un aviso y saldrá a la pantalla de entrada.

## 8. El perfil propio

- Cualquiera puede cambiar **su nombre, sus apellidos y su idioma** desde `PATCH /api/users/me`. Nada
  más: ni su correo, ni su papel, ni su origen, que son decisiones de otro.
- **El límite de Soporte es para editar a otros**: Soporte sólo cambia el nombre y los apellidos de
  otra cuenta, pero **de la suya cambia lo mismo que cualquiera**, su idioma incluido. Aplicarle ese
  límite a su propio perfil le dejaría sin poder cambiar el idioma de sus correos.
- **Cambiar el idioma cambia el de los correos que recibe**: es el único ajuste personal que el
  servidor necesita saber, y por eso vive en la cuenta y no en el navegador. El tema, en cambio, se
  queda en el navegador, porque al servidor le da igual.
- El perfil se abre desde el menú lateral, con los demás controles
  (`docs/interfaz-y-experiencia.md`, sección 3.1).
- **La cuenta de fábrica no tiene perfil**: no está en la tabla de cuentas, así que no hay fila que
  cambiar y se responde **404 `users.notFound`** (decisión del responsable, 2026-09-24). Su nombre y
  su contraseña viven en la configuración de la instalación, y la pantalla del perfil no se le
  enseña.

## 9. Los errores del módulo

Con claves, que el frontend traduce (`docs/interfaz-y-experiencia.md`, sección 8):

| Clave | Cuándo | Código |
| --- | --- | --- |
| `users.notFound` | La cuenta no existe | **404** |
| `users.email.duplicated` | El correo ya está en uso | **409** |
| `users.accountInactive` | Se reenvía el enlace de una cuenta desactivada: primero hay que reactivarla | **409** |
| `users.selfDeactivation` | Alguien intenta desactivar **su propia** cuenta | **403** |
| `users.stateHasItsOwnAction` | Se intenta cambiar el estado por el `PATCH` general, en vez de con `/deactivate` o `/activate` | **422** |
| `users.email.invalid` | El correo no tiene forma de correo | **422** |
| `users.name.required` | Falta el nombre o los apellidos | **422** |
| `users.origin.unknown` | El origen no es `local`, `ad` ni `keycloak` | **422** |
| `users.directory.notFound` | El directorio **ya no conoce** a esa persona, así que no se le devuelve el acceso | **422** |
| `users.origin.byDirectory` | Se intenta dar de alta una cuenta de directorio, o cambiarle el origen a `ad` o `keycloak`, **a mano** | **422** |
| `users.directory.activatesItself` | Se intenta reactivar a mano una cuenta de **Keycloak**, que se reactiva sola al entrar por su camino | **422** |
| `users.directory.unavailable` | No se pudo comprobar con el directorio, o esta instalación no tiene ninguno conectado | **503** |
| `users.origin.inUse` | No se puede cambiar el origen de una cuenta de directorio activa | **422** |
| `users.role.notAllowed` | Soporte intenta crear un papel que no le toca; **el papel no es ninguno de los cuatro**; o alguien intenta cambiar un dato que no le toca (el correo o el papel de otro, o los suyos propios) | **403** |

- **`users.role.notAllowed` es un 403 y no un 422** porque es un permiso: quien lo recibe ha pedido
  algo que no le corresponde, y no ha escrito mal nada.
- **Un papel inventado usa la misma clave que un papel que no toca**: en los dos casos el papel no se
  puede poner, y separarlos añadiría una clave que no cambia lo que hace ninguna pantalla.

## 10. Los correos del módulo

`users` **no manda correos**: se los pide a `auth`, que emite el enlace, y `auth` se los pide a
`mail`, que tiene el texto y envía. Los dos que puede provocar son el del alta y el reseteo, y los
dos llevan el mismo enlace. El texto va en el idioma global de la instalación (`docs/modules/mail.md`).

## 11. Lo que se decidió al repasar este documento

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **`last_login_at`** | Sí: saber cuándo entró alguien por última vez es lo que permite decidir con criterio a quién desactivar |
| 2 | **El detalle de una cuenta** | Sólo Administrador. Soporte ve la lista, que es lo que necesita para buscar a quien reporta |
| 3 | **Un correo, una cuenta** | Sí, y la comparación **ignora las mayúsculas** |
| 4 | **Las claves que faltaban** | `users.origin.unknown`, y un papel inventado usa `users.role.notAllowed`. La antigua clave de idioma desapareció con la preferencia personal |
| 5 | **El código de cada clave** | 404, 409, 422 y 403, cada uno donde toca |
| 11 | **El perfil de la cuenta de fábrica** | Responde **404 `users.notFound`** y no se le enseña la pantalla: no está en la tabla de cuentas |
| 12 | **Nadie se desactiva a sí mismo** | **403 `users.selfDeactivation`**: desactivarse es quedarse fuera sin poder volver |
| 14 | **El estado no se cambia por el `PATCH`** | El documento se contradecía: la tabla del `PATCH` incluye el estado, pero el mismo documento explica que desactivar tiene su propia ruta para que no pase por accidente. Manda el motivo: el `PATCH` responde `users.stateHasItsOwnAction` |
| 13 | **Cambiar el origen, hoy** | Se comprueba lo que se puede (no a una cuenta de directorio activa) y **pasar a `ad` o `keycloak` se rechaza** mientras no exista la consulta al directorio; pasar a `local` sí funciona |
| 6 | **Reenviar el enlace a una cuenta desactivada** | No: 409 `users.accountInactive`, porque sería prometer una entrada que no va a ocurrir |
| 7 | **El alta de cuentas de directorio, por ahora** | Rechazada con `users.directory.notFound` mientras no exista la comprobación contra el directorio. **El 2026-09-25 el responsable decidió cómo queda cuando exista** (y ya existe): no se crean a mano y el mensaje lo dice (sección 5, punto 4) |
| 20 | **Reactivar una cuenta de directorio** | **Pregunta al directorio** si lo sigue conociendo: si lo encuentra, se reactiva; si no, `users.directory.notFound`. Decidido el 2026-09-25 y **hecho ese mismo día**, con un alcance que se fijó al implementarlo: **en Keycloak no se pregunta** —pediría una cuenta de servicio en el reino— y su reactivación a mano se rechaza diciendo que vuelve sola al entrar |
| 21 | **El mensaje del alta a mano y del cambio de origen** | Dejan de hablar del directorio —que sí conoce a esa persona— y dicen lo que pasa: **esas cuentas no se crean ni se cambian a mano, las pone el directorio al entrar**. Decidido el 2026-09-25 y **hecho** con una sola clave, `users.origin.byDirectory`, porque las dos situaciones se cuentan con la misma frase |
| 8 | **La dirección de correo en el log** | Sí, por decisión del responsable: es lo que permite investigar un «no me ha llegado». Nunca contraseñas, tokens ni el contenido del correo |
| 15 | **Dónde edita Soporte** | **En línea, desde la lista**, porque no ve la ficha de nadie (decisión 2). El hallazgo: sin ficha, su permiso de cambiar nombre y apellidos no tenía dónde ejercerse |
| 16 | **El alta** | **Un diálogo sobre la lista**, no una pantalla: son seis campos, y al guardar se vuelve a donde se estaba |
| 17 | **El origen en la pantalla** | Los tres, con `ad` y `keycloak` **desactivados y con su nota**: se ve que están previstos y no se ofrece algo que hoy va a fallar |
| 18 | **La lista** | **Tabla en PC y tarjetas en móvil**, con chips de filtro (papel, origen, estado y búsqueda) y paginación. Orden fijo, por apellidos y nombre |
| 19 | **La ficha y el perfil** | **Dos rutas**: `/users/{id}`, que sólo abre un Administrador, y `/profile`, que se abre desde tu nombre en el menú. La cuenta de fábrica no tiene perfil |

## 12. Qué habilita este documento

Con `docs/modules/users.md` aprobado, el siguiente es **`docs/modules/auth.md`**, que ya tiene tomadas
sus decisiones (token sin estado de 10 horas, AD por búsqueda con cuenta de servicio, enlaces con
huella y un solo uso, contenedores de pruebas para el directorio). Después, `users` y `auth` se
implementan juntos: no se puede tener cuentas sin poder entrar, ni entrar sin cuentas.

## 13. Idioma global, sin preferencia personal

La migración elimina `users.language`. Las DTO de alta, lista, detalle, sesión y perfil no
aceptarlo y devolverlo; desaparece `users.language.unknown`. AD y Keycloak ya no asignan el idioma de
la instalación al crear una cuenta. `PATCH /api/users/me` cambia sólo nombre y apellidos, y
`PATCH /api/users/{id}` tampoco admite idioma.

Las pantallas de lista, alta, ficha y perfil retiran campos, etiquetas y filtros de idioma. Fechas y
textos se formatean con el idioma global entregado por `settings`. Los correos de alta y cambio de
contraseña usan también ese valor global, sin leer la cuenta.

No cambian papeles, orígenes, permisos, directorios, desactivación ni contraseñas. El responsable
confirmó el alcance el 2026-10-06. El repaso no añadió excepciones personales ni migración de
valores. Se implementó y verificó el 2026-10-07.
