# Usuarios y permisos

> **Estado:** aprobado
> **Última actualización:** 2026-09-22
>
> Aprobado por el responsable el 2026-09-22, con las siete decisiones de la antigua sección 10
> confirmadas tal como estaban propuestas. Desbloquea `tickets.md` y `flujos.md`, y con ellos los
> módulos `auth` y `users`, que son los primeros que se pueden implementar. **No habilita escribir
> código por sí solo** (Regla 0 de `AGENTS.md`).
>
> **Enmendado el 2026-09-22** en dos cosas, al escribir `docs/tickets.md`:
>
> 1. Se añadió a la matriz la acción de **cerrar el ticket interno** (Desarrollo y Soporte): se
>    decidió que su cierre es manual, no consecuencia del cierre del principal.
> 2. Se añadió la acción de **configurar el prefijo de la numeración** (sólo Administrador), y la
>    sección 6 ya no dice que el prefijo sea configuración del servidor: se cambia desde la
>    aplicación.
> 3. Se añadió la **sección 8**, la cuenta de administrador de fábrica, que entra escribiendo
>    `admin` y cuya contraseña viene de `ADMIN_PASSWORD` y se aplica en cada arranque. La sección 4
>    recoge que es la única cuenta que no se identifica por su correo.
> 8. Al empezar `docs/autenticación.md`: **la sesión deja de ir en cookie** y pasa a un token en
>    `localStorage` enviado en la cabecera `Authorization`. Se elige así para que el navegador no
>    mande nada por su cuenta, a cambio de que un XSS pueda leer el token; las compensaciones y la
>    regla de no añadir scripts de terceros quedan escritas en la sección 6.
> 7. Al documentar el módulo de usuarios: el papel `administrador` pasa a **ver los tickets en sólo
>    lectura**, incluidos los internos, y su cuenta de fábrica no existe como fila.
> 6. Al documentar el módulo de usuarios: se añadió a la matriz el **cambio de papel** (Administrador),
>    se quitó la **baja** del ciclo de vida —sólo se desactiva—, y la cuenta de fábrica dejó de ser una
>    fila: **vive en la configuración**.
> 5. Al empezar `docs/autenticación.md` se cambió la **duración de la sesión a 10 horas** y se
>    escribió lo que implica un token sin estado: **la sesión no se puede revocar en el servidor**, así
>    que cerrar sesión borra el token del navegador pero no lo invalida. La cuenta sí se lee en cada
>    petición, así que desactivar a alguien **bloquea al instante**.
> 4. Al repasar `docs/tickets.md` se añadió a la matriz el **reseteo de la contraseña**, que pueden
>    lanzar **Soporte y los administradores**, y la sección 7 recoge las tres formas de tocar una
>    contraseña: el alta, el cambio propio y el reseteo (que envía correo y nunca asigna una
>    contraseña a mano).

## 1. Alcance de este documento

Fija **quién puede qué**, **cómo se entra**, **cómo se gestionan las cuentas** y las **reglas de
convivencia** entre los tres métodos de acceso. No define el modelo de datos del ticket
(`tickets.md`) ni los flujos paso a paso (`flujos.md`).

Es una **propuesta**: no habilita escribir código hasta que esté aprobada (Regla 0 de `AGENTS.md`).
La sección 11 lista las decisiones que propuse yo, todas confirmadas.

## 2. Los cuatro papeles

Vienen de `docs/propósito-y-alcance.md`. Aquí sólo se añade cómo entra cada uno:

| Papel | Qué es | Cómo entra |
| --- | --- | --- |
| **Usuario** | Quien necesita atención. Crea tickets y responde. | Cualquiera de los tres caminos |
| **Soporte Técnico** | Primer contacto. Atiende, escala y gestiona altas. | Cualquiera de los tres caminos |
| **Desarrollo** | Segundo nivel. Resuelve lo escalado. | Cualquiera de los tres caminos |
| **Administrador** | Gestiona cuentas. No atiende tickets. | Cualquiera de los tres caminos |

El papel es **uno por cuenta** y no se cambia desde la pantalla: se decide al dar de alta (sección 9).

## 3. La matriz: qué puede hacer cada papel

| Acción | Usuario | Soporte | Desarrollo | Administrador |
| --- | :---: | :---: | :---: | :---: |
| Ver sus propios tickets | ✅ | ✅ | ✅ | — |
| Ver **todos** los tickets principales | — | ✅ | ✅ (sólo lectura, regla 2) | ✅ (sólo lectura) |
| Ver los tickets **internos** | — | ✅ | ✅ | ✅ (sólo lectura) |
| Crear un ticket propio | ✅ | ✅ | ✅ | — |
| Crear un ticket **en nombre de otro usuario** | — | ✅ | — | — |
| Comentar en el ticket principal | ✅ (el suyo) | ✅ | — | — |
| Comentar en el ticket interno | — | ✅ | ✅ | — |
| Asignar responsable | — | ✅ | ✅ (los internos) | — |
| Poner `en progreso` / `en espera` | — | ✅ | ✅ (los internos) | — |
| **Escalar** a Desarrollo | — | ✅ | — | — |
| **Resolver** el ticket interno | — | — | ✅ | — |
| **Cerrar** el ticket interno | — | ✅ | ✅ | — |
| Marcar el principal como `resuelto` **sin escalar** | — | ✅ | — | — |
| **Cerrar y reabrir** el principal | ✅ (el suyo) | ✅ | — | — |
| Ver la lista de usuarios (nombre y correo) | — | ✅ | — | ✅ |
| Dar de alta un usuario | — | ✅ (sólo rol `usuario`) | — | ✅ (cualquier rol) |
| Desactivar o reactivar una cuenta | — | ✅ | — | ✅ |
| Cambiar el **origen** de una cuenta | — | — | — | ✅ |
| Cambiar el **papel** de una cuenta | — | — | — | ✅ |
| Configurar el **prefijo** de la numeración de tickets | — | — | — | ✅ |
| **Resetear la contraseña** de un usuario | — | ✅ | — | ✅ |

Tres reglas que la matriz no puede expresar y que van aparte:

1. **El usuario sólo ve lo suyo.** Ni tickets de otros, ni la lista de usuarios, ni el ticket interno
   de su propia incidencia: ve su principal y nada más.
2. **Desarrollo no escribe en el ticket principal.** Lo lee para tener contexto, pero todo lo que
   tenga que decir va en el interno, y es Soporte quien decide qué traslada al usuario.
3. **Soporte crea usuarios, pero sólo con rol `usuario`.** Si Soporte pudiera crear un
   `administrador`, el papel más alto dejaría de estar protegido. Repartir `soporte`, `desarrollo` y
   `administrador` es cosa de un Administrador.

El cierre del ticket interno es manual a propósito: lo dan por terminado Desarrollo o Soporte, y
**cerrarlo no cambia el estado del principal**. Re-escalar el principal lo devuelve a `en progreso`.
El detalle está en `docs/tickets.md`, sección 3.2.

## 4. Los tres caminos de entrada

| | Correo y contraseña | Active Directory | Keycloak |
| --- | --- | --- | --- |
| **Alta** | La hacen Administrador o Soporte | Igual, validando que existe en el directorio; o **automática en el primer acceso** | Igual que AD |
| **Primer acceso** | El usuario recibe un correo para establecer su contraseña | Se crea la cuenta con rol `usuario` | Igual que AD |
| **Contraseña** | La guarda la aplicación | **La gestiona el dominio**: la aplicación no interviene | La gestiona Keycloak |
| **Identificador** | El correo | Correo + identificador del directorio | Correo + `sub` de Keycloak |

La única cuenta que **no** se identifica por su correo es la de fábrica, que se entra escribiendo
`admin` (sección 8). El formulario de entrada dice «correo o usuario» por ella.

Notas:

- **El correo es el identificador de la persona**: con él se entra, se reciben los avisos y se casa
  la cuenta local con la del directorio. Es único en todo el sistema.
- De AD se guarda además el **identificador del directorio** (el GUID o la cuenta) y de Keycloak el
  **`sub`**. Son los que mandan una vez vinculada la cuenta (regla 3 de la sección 5).
- AD se consulta con **LDAP** usando las credenciales que escribe el usuario: no hay otra forma de
  saber quién es sin preguntarle al dominio. Keycloak, por OIDC.

## 5. Reglas de convivencia

Son las que evitan que la misma persona acabe con dos cuentas, o sin ninguna.

1. **El directorio manda.** Si la persona existe y está activa en AD o Keycloak, entra. La
   configuración institucional tiene más peso que el estado local de la cuenta.
2. **Se vincula por correo, nunca se duplica.** Si al entrar por primera vez por AD o Keycloak ya
   existe una cuenta con ese correo, el acceso se vincula a ella: conserva sus roles, sus tickets y
   su historial.
3. **Una vez vinculada, manda el identificador del directorio**, no el correo. Si en el directorio le
   cambian el correo, la cuenta sigue siendo la misma y el correo se actualiza.
4. **Un origen por cuenta** (`local`, `ad`, `keycloak`), y lo fija el primer camino que se use. Si la
   persona existe en los dos directorios, entra por el que use primero y ahí se queda.
5. **Si el origen es `ad` o `keycloak`, no hay contraseña local**, aunque la cuenta haya tenido una
   antes. El origen manda.
6. **La desactivación local no bloquea a una cuenta de directorio.** Si un usuario desactivado aquí
   sigue activo en AD o Keycloak, **vuelve a entrar y la cuenta se reactiva sola**. Para quitarle el
   acceso de verdad hay que desactivarlo **en el directorio**. Esta regla no aplica a las cuentas
   `local`, que no tienen directorio al que preguntar: ahí desactivar sí bloquea, y reactivar es una
   acción explícita.
7. **Cambiar el origen es una acción de un Administrador**, explícita y registrada. Nunca ocurre por
   entrar desde otro sitio.

## 6. La sesión

- **El token viaja en una cabecera `Authorization: Bearer …`**, no en una cookie. El frontend lo
  guarda y lo adjunta en cada petición con un interceptor; **el navegador no manda nada por su
  cuenta**, que es lo que se buscaba: todo lo que viaja, viaja porque nuestro código lo pone.
- **Dónde se guarda: `localStorage`.** Es una decisión **consciente y contraria** a la del proyecto
  hermano Calibyou, donde el token en `localStorage` quedó registrado como hallazgo de seguridad
  porque **un fallo de XSS puede leerlo**. Se acepta a cambio de no tener nada automático en el
  navegador, y se compensa con lo que ya está decidido: política de seguridad estricta en producción
  (sin scripts en línea ni de terceros), ningún contenido de usuario se pinta como HTML —los
  comentarios son texto y Angular los escapa— y los `svg` se fuerzan a descarga.
- **La regla que va con eso**: **no se añaden scripts de terceros a la aplicación**. Cualquier
  script que se cargue puede leer el token; una analítica, un chat de soporte o una fuente con
  JavaScript abrirían por la puerta lo que acabamos de aceptar por la ventana.
- **Duración de 10 horas**, sin renovación deslizante: se entra una vez y aguanta la jornada larga.
  Es lo simple, y lo que evita ventanas de sesión abiertas para siempre.
- **Dentro del token va el identificador de quien ha entrado**, y nada más: **la cuenta se lee de la
  base en cada petición**, que es lo que permite saber su papel y si sigue activa. Cuesta una
  consulta indexada por petición y a cambio **desactivar y cambiar de papel surten efecto al
  instante**.
- **Cerrar sesión borra el token del navegador**, y con eso basta porque **el token no se puede
  revocar en el servidor**: si alguien se hubiera copiado el token, le seguiría sirviendo hasta que
  caduque. Es el precio de no tener una tabla de sesiones, y está escrito aquí para que nadie lo dé
  por hecho.
- **El interceptor lee el token en cada petición**, no lo guarda en memoria: así, cerrar sesión en
  una pestaña deja sin sesión a las demás en cuanto vuelvan a pedir algo, en vez de dejarlas
  trabajando con una sesión que ya no existe.
- **No hay nada que proteger contra CSRF**: una petición de otro sitio no puede poner la cabecera
  `Authorization`, porque no tiene el token. Es la ventaja del token explícito, y por eso desaparece
  la comprobación de `Origin` que se había previsto.
- El secreto de firma es una **variable de entorno** (`TOKEN_SECRET`), nunca un valor en el código ni
  en git. **El único ajuste que se cambia desde la aplicación es el prefijo de la numeración de los
  tickets**, y lo cambia un Administrador (está en la matriz y en `docs/tickets.md`, sección 2.2).

## 7. Contraseñas, cambios de contraseña y correos de cuenta

- **Longitud mínima 12, sin caducidad forzada y sin preguntas de seguridad.** La caducidad empuja a
  la gente a inventarse variantes de la misma contraseña, y las preguntas de seguridad son un camino
  de recuperación más débil que el correo. La **credencial de fábrica** (sección 8) es la única
  excepción, y por una razón práctica: si le aplicara, el backend no arrancaría en desarrollo.
- Se guardan **con hash** (`bcrypt`), nunca en claro y nunca cifradas de forma reversible.
- **Alta**: se envía un correo con un **enlace de un solo uso** para que el usuario establezca su
  contraseña. **Caduca a las 24 horas.** Sin ese paso la cuenta no sirve para entrar, así que el alta
  y la contraseña son la misma cosa.
- **Cambiarla cuando se quiera**: un usuario `local` puede cambiar su contraseña estando dentro, sin
  tener que simular que la ha olvidado. Con AD o Keycloak no: la gestiona el dominio.
- **Olvidarla**: el mismo enlace del alta, con **1 hora** de caducidad, porque ese correo lo puede
  pedir cualquiera y conviene que la ventana sea corta.
- **Soporte o un administrador pueden lanzar un reseteo** por alguien que no puede entrar. Soporte
  es quien atiende al usuario que llama diciendo que no entra, así que es quien mejor puede
  lanzarlo. El usuario recibe el mismo correo: **nadie le asigna una contraseña ni se la dice por
  teléfono**, que es como se filtran.
- Si el origen es AD o Keycloak, **no se envía nada**: la contraseña es del dominio (regla 5).
- Los correos de cuenta (alta, cambio y recuperación) son **cosa distinta de los avisos de ticket**:
  esos son **siete** y están listados en `docs/propósito-y-alcance.md`. Aquí no se cuenta ninguno.
- Ni contraseñas, ni tokens, ni enlaces de recuperación se escriben en los logs.

## 8. La cuenta de administrador de fábrica

Una instalación nueva no tiene a nadie que pueda dar de alta usuarios, así que nace con **una sola
cuenta**. Es la que configura el sistema la primera vez: dar de alta a Soporte Técnico, a Desarrollo
y al resto de la gente.

| | |
| --- | --- |
| **Identificador** | `admin`. Constante en el código: no es variable de entorno, porque nunca cambia |
| **Contraseña** | La variable de entorno `ADMIN_PASSWORD` |
| **Papel** | `administrador` |
| **Correo** | **Ninguno.** Es la única cuenta a la que no se le exige, y da igual: no necesita avisos |
| **Dónde vive** | **En la configuración, no en la base de datos**: no hay ninguna fila que crear |

Detalles que importan:

- **En desarrollo** la contraseña vale `admin`, y está en `config/env/dev.env`, que se versiona con
  el resto de credenciales del contenedor local.
- **No existe como fila**: es la única cuenta que no está en la tabla de cuentas. Se entra con ella
  porque la aplicación reconoce `admin` y **compara la contraseña con la variable de entorno**, no con
  nada guardado en la base.
- **En producción es obligatoria**: el backend **no arranca** sin `ADMIN_PASSWORD`, y su valor vive
  sólo en `config/env/prod.env`, que no se versiona. **La contraseña de producción no se escribe en
  esta documentación**, que va al repositorio.
- **No le aplica la política de contraseñas** de la sección 7: es una credencial de fábrica, y
  exigirle doce caracteres impediría arrancar en desarrollo.
- **La contraseña sólo se cambia en la configuración.** No se puede cambiar desde la aplicación
  —no hay dónde guardarla— ni recuperar por correo —no tiene—. Cambiarla es editar la variable y
  reiniciar, y así **nunca se pierde el acceso a la instalación**. Nada de «he olvidado la
  contraseña del administrador».
- **No la siembra nadie**: no hay nada que sembrar. El resto de cuentas las da de alta una persona,
  y esta existe desde antes del primer arranque.
- **Al no ser una fila, no puede aparecer en un ticket** como autor, solicitante ni responsable: los
  tickets apuntan a cuentas que existen. Su trabajo es dejar el sistema en marcha, no atender.
- **Ve todo en sólo lectura**, tickets incluidos: puede comprobar qué está pasando, pero no comentar,
  ni mover estados, ni asignar. El papel `administrador` es el más alto en **configuración**
  (usuarios, numeración y apariencia) y **sólo lectura** en todo lo demás.

## 9. Ciclo de vida de una cuenta

| Momento | Cómo funciona |
| --- | --- |
| **Alta** | Administrador o Soporte desde la aplicación (Soporte sólo con rol `usuario`), o automática en el primer acceso por AD o Keycloak. La cuenta de fábrica (sección 8) es la excepción: existe desde el primer arranque |
| **Primer acceso** | Por enlace del correo si es local; directa si viene del directorio |
| **Cambio de correo** | Lo cambia un Administrador. Si la cuenta es de directorio, manda el directorio |
| **Desactivación** | La hacen Administrador o Soporte. Bloquea **al instante**: la cuenta se comprueba en cada petición, así que la siguiente que haga ya no vale. En cuentas de directorio no bloquea si allí sigue activa (regla 6) |
| **Reactivación** | Automática si el directorio dice que está activa; explícita (Administrador o Soporte) en cuentas `local` |

Por eso la tabla de cuentas **no lleva `deleted_at`**: sería una columna que nadie usaría.
| **Baja** | **No existe**: no hay acción de dar de baja. Una cuenta desactivada sigue siendo legible por los tickets que tenga, y con eso basta |

## 10. Dónde se comprueba el permiso

- **En el backend, siempre.** La autorización vive en `backend/shared/authz` y se aplica con
  middlewares sobre las rutas; ningún controlador decide permisos por su cuenta.
- **En el frontend, sólo para no enseñar lo que no se puede hacer.** Ocultar un botón no es una
  medida de seguridad: la decisión se toma en el backend.
- Módulos implicados: **`auth`** (entrar, salir, contraseñas) y **`users`** (cuentas, roles,
  desactivación). En el frontend, la sesión vive en `core` —guardas, interceptor y cierre de sesión—,
  porque es armazón y no una pantalla de módulo; es la excepción ya prevista en
  `docs/arquitectura.md`.

## 11. Qué habilita este documento

Este documento, ya aprobado, fija los permisos, el acceso y el ciclo de vida de las cuentas, y
desbloquea:

1. `tickets.md` — modelo de datos, transiciones y lista cerrada de módulos, que ya puede decir quién
   ve y quién mueve cada cosa.
2. `flujos.md` — los flujos paso a paso, con los avisos por correo.

Con `tickets.md` y `flujos.md` aprobados, **`auth` y `users` son los primeros módulos que se pueden
implementar**, porque no dependen de ninguna decisión pendiente del producto.
