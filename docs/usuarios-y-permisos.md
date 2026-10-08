# Usuarios y permisos

> **Estado:** as-built
> **Última actualización:** 2026-10-08
>
> **Enmienda implementada y verificada el 2026-10-08.** Únicamente Soporte y Desarrollo pueden
> mejorar con IA los borradores de los editores que ya tienen permitidos. El backend lo exige y la
> interfaz lo oculta al Usuario y al Administrador; no concede permisos nuevos sobre tickets.
>
> **Enmendado el 2026-10-07**, conforme a la propuesta aprobada del 2026-10-06: el idioma dejó de
> pertenecer a las cuentas. Es único para toda la instalación y sólo el Administrador lo cambia
> desde Configuración. Datos, API, sesión y pantallas ya cumplen esta regla.
>
> **Enmendado el 2026-09-29 (segunda vez)**: **la contraseña sólo se cambia en las cuentas locales**
> (decisión del responsable). El botón de mandar el enlace pasa a llamarse **«Cambiar contraseña»** —es lo
> que hace— y **sólo se le enseña a las cuentas con origen `local`**: con Active Directory o Keycloak la
> contraseña **la comprueba el directorio**, así que ofrecerlo sería una promesa falsa. **Y lo garantiza el
> servidor**, no sólo la pantalla: pedir el enlace de una cuenta de directorio responde **422** con
> `users.password.notLocal`. Encaja con una regla que ya estaba en la base: **una cuenta de directorio no
> puede tener contraseña** (`users_directory_has_no_password`).
>
> **Enmendado el 2026-09-29**: **el catálogo lo mantiene sólo el Administrador** —crear, renombrar y
> retirar, categorías y etiquetas— (decisión 83 de `docs/modules/tickets.md`, decisión del responsable).
> **Soporte y Desarrollo lo usan** —clasifican tickets y filtran— pero **no lo cambian**. Se corrigen así
> las dos filas de la matriz y **la decisión 73**, que era propuesta mía y queda decidida al revés.
>
> **Enmendado el 2026-09-27 (tercera vez)**: las **etiquetas también se mantienen** —crear y renombrar,
> Soporte y Administrador; **retirar, sólo el Administrador**—, con las dos filas nuevas de la matriz
> (decisión 73 de `docs/modules/tickets.md`). Esta fue la regla implementada inicialmente; la
> enmienda del 2026-09-29 que aparece justo arriba la sustituyó por la decisión 83 y cerró el asunto:
> hoy sólo el Administrador mantiene todo el catálogo.
>
> **Enmendado el 2026-09-27 (segunda vez)**, al entrar **las categorías y las etiquetas**: la matriz
> estrena tres filas —elegirlas, mantener el catálogo y **retirar una categoría, que es sólo del
> Administrador**—. **Todo ticket nace con categoría**: es obligatoria y la elige quien lo abre.
>
> **Enmendado el 2026-09-27**, al entrar **las menciones y los observadores**: la matriz estrena las dos
> filas de **etiquetar** y **quitar a un observador** (Soporte y Desarrollo), y la regla 3 nueva dice lo
> que más se puede confundir: **ser observador no es ser responsable**. **El asignado sigue siendo
> opcional**, como estaba.
>
> **Enmendado el 2026-09-26**, al repartir las listas de tickets: la fila «ver sus propios tickets»
> dice ya **qué es lo suyo** —lo asignado, lo abierto por uno y donde ha comentado (regla 2)—, y
> Soporte y Desarrollo estrenan **dos listas donde ven todo** (Tickets principales y Tickets internos)
> además de su bandeja. **No cambia ningún permiso**: lo que cambia es dónde se ve lo que cada papel
> ya podía ver (`docs/modules/tickets.md`, decisiones 51 y 52).
>
> **Enmendado el 2026-09-25**, al bajar el responsable **la longitud mínima de las
> contraseñas de 12 a 8 caracteres** (sección 7), que es la única regla de la política.
>
> **Enmendado el 2026-09-25**, al pedir el responsable que **la instalación entre por un método a la
> vez** (sección 5): los tres caminos de la sección 4 **ya no conviven**. El que está puesto es el
> único que se atiende, y con el método en `ad` o en `keycloak` **las cuentas locales —Soporte y
> Desarrollo incluidas— no pueden entrar**; la cuenta de fábrica sí, siempre, porque es la única
> puerta que no se puede cerrar (sección 8). La enmienda está contada entera en
> `docs/modules/settings.md`, sección 5.8, y en `docs/modules/auth.md`, secciones 5.0 y 5.5.
>
> **Enmendado el 2026-09-24**, al empezar a implementar `tickets`: la matriz estrena la fila de
> **reabrir el ticket interno** (Soporte), que es la mitad que faltaba de «cerrar el interno»: la
> transición existe en `docs/modules/tickets.md`, sección 3.2 —`cerrado` → `en progreso`, y equivale a
> volver a escalar— y aquí no se veía.
>
> Aprobado por el responsable el 2026-09-22, con las siete decisiones de la antigua sección 10
> confirmadas tal como estaban propuestas. Desbloquea `modules/tickets.md` y `flujos.md`, y con ellos los
> módulos `auth` y `users`, que son los primeros que se pueden implementar. **No habilita escribir
> código por sí solo** (Regla 0 de `AGENTS.md`).
>
> **Enmendado el 2026-09-22** en dos cosas, al escribir `docs/modules/tickets.md`:
>
> 1. Se añadió a la matriz la acción de **cerrar el ticket interno** (Desarrollo y Soporte): se
>    decidió que su cierre es manual, no consecuencia del cierre del principal.
> 2. Se añadió la acción de **configurar el prefijo de la numeración** (sólo Administrador), y la
>    sección 6 ya no dice que el prefijo sea configuración del servidor: se cambia desde la
>    aplicación.
> 3. Se añadió la **sección 8**, la cuenta de administrador de fábrica, que entra escribiendo
>    `admin` y cuya contraseña viene de `ADMIN_PASSWORD` y se aplica en cada arranque. La sección 4
>    recoge que es la única cuenta que no se identifica por su correo.
> 8. Al empezar `docs/modules/auth.md`: **la sesión deja de ir en cookie** y pasa a un token en
>    `localStorage` enviado en la cabecera `Authorization`. Se elige así para que el navegador no
>    mande nada por su cuenta, a cambio de que un XSS pueda leer el token; las compensaciones y la
>    regla de no añadir scripts de terceros quedan escritas en la sección 6.
> 7. Al documentar el módulo de usuarios: el papel `administrador` pasa a **ver los tickets en sólo
>    lectura**, incluidos los internos, y su cuenta de fábrica no existe como fila.
> 6. Al documentar el módulo de usuarios: se añadió a la matriz el **cambio de papel** (Administrador),
>    se quitó la **baja** del ciclo de vida —sólo se desactiva—, y la cuenta de fábrica dejó de ser una
>    fila: **vive en la configuración**.
> 5. Al empezar `docs/modules/auth.md` se cambió la **duración de la sesión a 10 horas** y se
>    escribió lo que implica un token sin estado: **la sesión no se puede revocar en el servidor**, así
>    que cerrar sesión borra el token del navegador pero no lo invalida. La cuenta sí se lee en cada
>    petición, así que desactivar a alguien **bloquea al instante**.
> 4. Al repasar `docs/modules/tickets.md` se añadió a la matriz el **reseteo de la contraseña**, que pueden
>    lanzar **Soporte y los administradores**, y la sección 7 recoge las tres formas de tocar una
>    contraseña: el alta, el cambio propio y el reseteo (que envía correo y nunca asigna una
>    contraseña a mano).

## 1. Alcance de este documento

Fija **quién puede qué**, **cómo se entra**, **cómo se gestionan las cuentas** y las **reglas de
convivencia** entre los tres métodos de acceso. No define el modelo de datos del ticket
(`modules/tickets.md`) ni los flujos paso a paso (`flujos.md`).

Está **aprobado**, así que habilita escribir código (Regla 0 de `AGENTS.md`).
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
| Ver sus propios tickets —para un técnico, los suyos, que son los que tiene asignados, los que abrió y donde ha comentado— | ✅ | ✅ | ✅ | — |
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
| **Reabrir** el ticket interno *(es volver a escalar)* | — | ✅ | — | — |
| Marcar el principal como `resuelto` **sin escalar** | — | ✅ | — | — |
| **Cerrar y reabrir** el principal | ✅ (el suyo) | ✅ | — | — |
| **Elegir la categoría** del ticket | ✅ (el suyo) | ✅ | — | — |
| **Poner, cambiar y quitar las etiquetas** del ticket | — | ✅ (desde el principal) | ✅ (desde el interno) | — |
| **Mantener el catálogo de categorías** (crear, renombrar y retirar) | — | — | — | ✅ |
| **Mantener las etiquetas** (crear, renombrar y retirar, en todos los tickets) | — | — | — | ✅ |
| **Retirar una etiqueta** del catálogo y de los tickets | — | — | — | ✅ |
| **Retirar** una categoría del catálogo | — | — | — | ✅ |
| **Etiquetar** a otro técnico o desarrollador (y con eso hacerlo observador) | — | ✅ | ✅ | — |
| **Quitar** a un observador | — | ✅ | ✅ | — |
| Ver la lista de usuarios (nombre y correo) | — | ✅ | — | ✅ |
| Dar de alta un usuario | — | ✅ (sólo rol `usuario`) | — | ✅ (cualquier rol) |
| Desactivar o reactivar una cuenta | — | ✅ | — | ✅ |
| Cambiar el **origen** de una cuenta | — | — | — | ✅ |
| Cambiar el **papel** de una cuenta | — | — | — | ✅ |
| Configurar la **instalación** (numeración, reparto, avisos, apariencia y textos de los correos) | — | — | — | ✅ |
| **Resetear la contraseña** de un usuario | — | ✅ | — | ✅ |

Cinco reglas que la matriz no puede expresar y que van aparte:

1. **El usuario sólo ve lo suyo.** Ni tickets de otros, ni la lista de usuarios, ni el ticket interno
   de su propia incidencia: ve su principal y nada más.
2. **«Lo mío» vale para los tres papeles, y significa lo mismo** (decisión del responsable,
   2026-09-26): lo que tengo asignado, lo que abrí yo y aquello donde he comentado. Con eso, Soporte
   y Desarrollo tienen **su** bandeja —**Mis tickets**— además de las listas donde lo ven todo
   —**Tickets principales** y **Tickets internos**—, que es dónde está el trabajo que no es suyo. El
   Administrador no tiene tickets propios: su bandeja es la de siempre, con los dos tipos y de sólo
   lectura (`docs/modules/tickets.md`, decisión 52).
3. **Ser observador no es ser responsable.** Un ticket tiene **como mucho un asignado** —el que
   atiende— y **muchos observadores** —los que lo siguen, porque alguien los etiquetó—. **Observar no
   da permisos**: ni escribe quien no escribía, ni mueve estados quien no los movía; sólo hace que el
   ticket esté en su bandeja (`docs/modules/tickets.md`, decisión 60).
4. **Desarrollo no escribe en el ticket principal.** Lo lee para tener contexto, pero todo lo que
   tenga que decir va en el interno, y es Soporte quien decide qué traslada al usuario.
5. **Soporte crea usuarios, pero sólo con rol `usuario`.** Si Soporte pudiera crear un
   `administrador`, el papel más alto dejaría de estar protegido. Repartir `soporte`, `desarrollo` y
   `administrador` es cosa de un Administrador.

El cierre del ticket interno es manual a propósito: lo dan por terminado Desarrollo o Soporte, y
**cerrarlo no cambia el estado del principal**. Re-escalar el principal lo devuelve a `en progreso`.
El detalle está en `docs/modules/tickets.md`, sección 3.2.

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
- **Los tres caminos no están abiertos a la vez** (enmienda del 2026-09-25): la instalación entra por
  **uno solo**, el que esté puesto, y se cambia desde Configuración sin reiniciar nada
  (`docs/modules/settings.md`, sección 5.8). Dónde se configuran los dos caminos de directorio —en la
  pantalla, no en el entorno— está también ahí.

## 5. Reglas de convivencia

Son las que evitan que la misma persona acabe con dos cuentas, o sin ninguna.

0. **La instalación entra por un método a la vez** (enmienda del 2026-09-25). Un administrador elige
   `local`, `ad` o `keycloak` desde Configuración, y **los otros dos quedan apagados**: no es que no
   funcionen, es que la instalación no los ofrece. Cambiarlo vale en la entrada siguiente, sin
   reiniciar nada.
   - **Con el método en `ad` o en `keycloak`, las cuentas locales no entran** —Soporte y Desarrollo
     incluidas—, porque su contraseña es de aquí y el camino de la instalación es otro. **Nadie pierde
     su cuenta**: lo que se apaga es la puerta, y al volver a `local` todo el mundo entra otra vez.
   - **La cuenta de fábrica entra siempre**, sea cual sea el método (sección 8): es la que puede
     devolver el método a donde estaba sin tocar la base a mano.
   - **No se puede elegir un método que no esté configurado**: la instalación no se queda sin puerta
     por un descuido, porque el backend lo rechaza antes de guardarlo.
1. **El directorio manda**, **cuando el directorio es el camino de la instalación**. Si la persona
   existe y está activa en AD o Keycloak, entra. La configuración institucional tiene más peso que el
   estado local de la cuenta. Y al revés: con el método en `local`, una cuenta de directorio no entra,
   porque aquí no hay contraseña suya que comparar.
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
  tickets**, y lo cambia un Administrador (está en la matriz y en `docs/modules/tickets.md`, sección 2.2).

## 7. Contraseñas, cambios de contraseña y correos de cuenta

- **Longitud mínima 8, sin caducidad forzada y sin preguntas de seguridad.** La caducidad empuja a
  la gente a inventarse variantes de la misma contraseña, y las preguntas de seguridad son un camino
  de recuperación más débil que el correo. **El mínimo eran 12 y el responsable lo bajó a 8 el
  2026-09-25**: en una mesa de ayuda interna la contraseña no es lo que protege el sistema —lo que lo
  protege es quién tiene cuenta, y una cuenta se desactiva en un clic—, y un mínimo alto sólo consigue
  que la gente escriba la misma frase con un número detrás. **La credencial de fábrica** (sección 8)
  sigue siendo la única excepción, y por una razón práctica: si le aplicara, el backend no arrancaría
  en desarrollo.
- **Es la única regla**: ni mayúsculas, ni números, ni símbolos obligatorios. Una regla de composición
  no añade entropía de verdad y sí añade contraseñas apuntadas en un papel.
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
- Los correos de cuenta —**el alta, la recuperación y el aviso de que la contraseña ha cambiado**—
  son **cosa distinta de los avisos de ticket**: esos son **siete** y están listados en
  `docs/propósito-y-alcance.md`. El detalle de los tres está en `docs/modules/auth.md`, sección 7.
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
  exigirle un mínimo impediría arrancar en desarrollo.
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
- **Entra siempre, sea cual sea el método de entrada** (enmienda del 2026-09-25), y por eso va lo
  primero que se comprueba al entrar. Es lo que hace que elegir mal el método no deje la instalación
  sin salida: con el método en `keycloak` la pantalla de entrada no tiene formulario, así que **su
  puerta es un enlace discreto** («Entrar como administrador») que lo saca
  (`docs/modules/auth.md`, secciones 5.0 y 9).

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

1. `modules/tickets.md` — modelo de datos, transiciones y lista cerrada de módulos, que ya puede decir quién
   ve y quién mueve cada cosa.
2. `flujos.md` — los flujos paso a paso, con los avisos por correo.

Con `modules/tickets.md` y `flujos.md` aprobados, **`auth` y `users` son los primeros módulos que se pueden
implementar**, porque no dependen de ninguna decisión pendiente del producto.

## 12. El idioma no es un permiso ni una preferencia personal

- Todas las cuentas ven la aplicación y reciben correos en el idioma global de la instalación.
- Crear, sincronizar o editar una cuenta no recibe ni devuelve idioma. Cambiar el perfil propio queda
  limitado a nombre y apellidos.
- El selector desaparece de entrada, menú, perfil, alta y ficha. Sólo el Administrador cambia el
  idioma global, y la cuenta de fábrica puede hacerlo cuando una actualización exige configurar IA.
- La columna personal existente se retira del modelo y de las API; no se copia el idioma global en
  cada cuenta.
- El idioma global no modifica papeles, permisos, método de entrada, sesiones ni ciclo de vida.

Esta propuesta sustituye únicamente las reglas que atribuían idioma y correo a cada cuenta. El
responsable la confirmó el 2026-10-06. El repaso confirmó que las sesiones abiertas adoptan el idioma
global en su siguiente petición, sin cambiar permisos ni cerrar sesión. La regla se implementó y
verificó el 2026-10-07.

## 13. Implementación: permiso para mejorar un borrador con IA

| Acción | Usuario | Soporte | Desarrollo | Administrador |
| --- | --- | --- | --- | --- |
| Mejorar con IA una descripción o comentario que ya puede escribir | No | Sí | Sí | No |

El permiso exige las dos condiciones: papel admitido y permiso vigente sobre el ticket y el editor.
No permite ver un ticket, conversación o campo que la cuenta no pudiera consultar antes. El backend
lo comprueba en cada solicitud y responde 403 cuando falla; el frontend sólo evita ofrecer una acción
que será rechazada. La cuenta de fábrica conserva las capacidades del Administrador y no obtiene
este permiso operativo.
