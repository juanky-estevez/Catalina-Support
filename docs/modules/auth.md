# auth

> **Estado:** as-built
> **Última actualización:** 2026-10-07
>
> **Enmendado el 2026-10-07**, conforme a la propuesta aprobada del 2026-10-06: identidad, sesión y
> respuestas de cuenta ya no llevan idioma personal. Los tres correos de cuenta consultan el idioma
> global al enviarse, sin cambiar ninguno de los tres caminos de entrada.
>
> **Enmendado el 2026-10-03**, conforme a `docs/prueba-local.md` aprobado: Keycloak distingue emisor
> público e `internalIssuer` opcional. Descubrimiento, token y UserInfo usan la dirección interna;
> autorización sigue siendo pública. Se exige issuer anunciado igual al configurado y endpoints del
> mismo origen y reino, y se rechazan redirecciones HTTP. La configuración se lee en cada intento.
> El compose local anuncia `http://127.0.0.1:11006/sso` por defecto; nginx no es necesario.
>
> **Enmendado el 2026-10-01**: **los dos servicios de pruebas dejan `dev.yml`**. OpenLDAP y Keycloak
> viven ahora cada uno en **su propio archivo** —`active-directory.yml` y `keycloak.yml`—, con **su
> propio comando**, y **el perfil `auth` desaparece**. Los dos entran en la red
> **`catalina-support-dev`**, que **posee `dev.yml`** (nombre fijo, sin `external`), así que **primero
> se levanta el entorno** y después el servicio; el backend los alcanza por su nombre de servicio
> (`ldap`, `keycloak`). Se ponen al día la sección 11 y la decisión 1. Los puertos, los nombres de
> contenedor y las credenciales de pruebas no cambian.
>
> **Enmendado el 2026-09-25**, al pedir el responsable que **el directorio y Keycloak se
> configuren desde la pantalla** y que **la instalación entre por un método a la vez**: los datos de
> los dos caminos **salen del entorno y viven en la base** —`directory_settings` y
> `keycloak_settings`, `docs/modules/settings.md` sección 5.8—, `auth` los lee **en cada intento** por
> una interfaz que declara él mismo, y el método que está puesto decide por dónde se entra: los otros
> dos quedan apagados, con **la cuenta de fábrica siempre dentro** por su propia puerta. Las
> variables `LDAP_*` y `OIDC_*` **desaparecen**, y con esto quedan enmendadas las decisiones 28 y 34 y
> la sección 5.4. **Consecuencia dicha y aceptada**: con el método en `ad` o en `keycloak`, las
> cuentas locales —Soporte y Desarrollo incluidas— no entran (sección 5.5).
>
> Aprobado por el responsable el 2026-09-23, tras repasarlo en forma de preguntas. Fija
> el módulo `auth`: el token de sesión, los tres caminos de entrada, las contraseñas y sus
> enlaces, la cuenta de fábrica, los correos de cuenta, los endpoints, las seis pantallas del armazón (sección 9)
y los contenedores con los que se prueba.
>
> Se implementa junto con el otro: no se puede tener cuentas sin poder entrar, ni entrar sin cuentas.
>
> **Enmendado el 2026-09-23** (al empezar a implementar, sección 13, paso 1): la tabla de la
> sección 10 no tenía ninguna clave para el **token de sesión** —sólo las tenía para los enlaces de
> correo—, así que el middleware no podía explicar por qué rechaza una petición. Se añaden
> `auth.session.invalid` y `auth.session.expired`, se añade `auth.forbidden` para el 403 de los
> permisos (que no tenía clave en ningún documento: `docs/usuarios-y-permisos.md`, sección 10, dice
> dónde se comprueba el permiso, pero no con qué clave se contesta), y se fija el código HTTP con el
> que viaja cada clave. Las decisiones las aprobó el responsable el 2026-09-23.
>
> **Estado de la implementación (2026-09-23): el camino local está hecho y verificado.** Existen el
> inicio de sesión (cuenta local y cuenta de fábrica), `me`, `logout`, olvido, reseteo y cambio de
> contraseña, con sus enlaces y sus tres correos, probados de extremo a extremo contra desarrollo.
> **Las seis pantallas del armazón están hechas** y se entra desde el navegador (paso 4).
>
> **Estado (2026-09-25): los tres caminos están hechos y verificados** y **este documento está
> entero**. El local, el de AD y el de Keycloak, los tres contra sus servicios de verdad (el
> directorio de pruebas y el reino de pruebas viven en el repositorio, en `config/ldap/` y
> `config/keycloak/`), con los cinco casos de la sección 5.4 comprobados uno a uno en cada camino de
> directorio y con la cuenta que queda en la base. Del camino de Keycloak, además, se ha comprobado
> en un navegador de verdad el recorrido completo: el botón, la pantalla de Keycloak, la vuelta, el
> token en el fragmento y el fragmento borrado.
>
> **Enmendado el 2026-09-25**, al implementar el camino de Keycloak. Se fijan las
> cuatro claves nuevas de error (sección 10) y **cómo llega el navegador a Keycloak en desarrollo**:
> por el mismo dominio que la aplicación y por un camino (`/sso/`), que es lo que hace que **el
> emisor sea uno solo** y no dos direcciones que puedan discrepar (sección 11). Las decisiones
> 36 a 40 recogen lo que se decidió al implementarlo.
>
> **Enmendado el 2026-09-25**, al implementar el camino de AD, con una **corrección del responsable**:
> le llevé el vínculo como una pregunta con dos opciones y él contestó lo que de verdad quiere —
> «con AD o con Keycloak quisiera que los usuarios puedan ingresar directamente, **sin el proceso de
> alta manual** por un administrador o técnico»—. Se aplica: **nadie da de alta a nadie** que esté en
> el directorio; entra y su cuenta nace sola (sección 5.4). Lo que un administrador sigue decidiendo
> es el **papel**, porque eso no lo dice ningún directorio. Queda anotado como corrección en el
> registro (31 a 35), junto con el disparador del vínculo y lo que la pantalla de entrada no cambia.
>
> **Enmendado el 2026-09-23**, al escribir el camino local: se fijaron el campo de entrada,
> lo que devuelve la entrada, la ruta del enlace y dónde viaja su token. **El responsable corrigió una
> propuesta**: yo propuse la ruta en español y él decidió que **las rutas del frontend van siempre en
> inglés**, aunque el idioma del producto sea español. Se aplica lo que dijo y queda anotado como
> corrección en el registro (21).
>
> **Enmendado el 2026-09-23**, al escribir el armazón: se fija **de dónde sale el idioma
> de la interfaz** y se dice **cuáles son las seis pantallas**, que hasta ahora eran un número.
> **El responsable corrigió otra propuesta**: yo propuse entrar en español por defecto y él decidió
> que **se toma el idioma del navegador y, si no es ninguno de los dos, se entra en inglés**, con el
> conmutador siempre a la vista. Queda anotado como corrección en el registro (24 a 26).

## 1. Alcance de este documento

Cuenta **cómo se construye el módulo `auth`**: la tabla de tokens de enlace, el token de sesión, las
contraseñas, los tres caminos de entrada con su mecánica, la cuenta de fábrica, los correos de
cuenta, los endpoints, las pantallas del armazón y los contenedores con los que se prueba.

**No repite** lo que ya está decidido en otro sitio:

| Qué | Dónde está |
| --- | --- |
| Qué puede hacer cada papel | `docs/usuarios-y-permisos.md`, sección 3 |
| Las reglas de convivencia de los tres accesos | `docs/usuarios-y-permisos.md`, sección 5 |
| Cómo viaja la sesión y qué implica | `docs/usuarios-y-permisos.md`, sección 6 |
| La política de contraseñas y sus tres formas de cambiarlas | `docs/usuarios-y-permisos.md`, sección 7 |
| La cuenta de fábrica y por qué no está en la base | `docs/usuarios-y-permisos.md`, sección 8 |
| La tabla de cuentas | `docs/modules/users.md`, sección 2 |
| Las pantallas y el lenguaje visual | `docs/interfaz-y-experiencia.md` |

Está **aprobado**, así que habilita escribir código (Regla 0). La sección 12
recoge lo que he propuesto yo.

## 2. La tabla de tokens de enlace

Los enlaces de alta y de recuperación **no son tokens firmados**: son valores aleatorios que se
guardan, con su huella, en una tabla del módulo.

**`password_tokens`**

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `bigserial` | |
| `user_id` | `bigint` | A quién pertenece |
| `purpose` | `text` | `alta` o `recuperacion` |
| `token_hash` | `text` | **La huella**, nunca el token |
| `expires_at` | `timestamptz` | Alta: 24 horas. Recuperación: 1 hora |
| `used_at` | `timestamptz` | Cuándo se usó. Nulo mientras no se haya usado |
| `created_by_id` | `bigint` | Quién lo pidió. **Nulo si lo pidió la propia persona** |
| `created_at` | `timestamptz` | |

Decisiones que van con la tabla:

- **Se guarda el `sha256` del token, no el token.** Si alguien lee la base no puede reconstruir
  ningún enlace; y comparar es tan simple como calcular la huella de lo que llegó.
- **Un enlace, un uso.** Al usarse se marca `used_at` y deja de valer.
- **El token son 32 bytes aleatorios** en base64url. No se firma ni se deriva de nada: no hace falta,
  porque vive en una tabla.
- **Pedir uno nuevo borra los anteriores de esa cuenta y ese propósito.** Si alguien pide tres
  correos seguidos, sólo vale el último; guardar los otros dos no aportaría nada, porque ya no
  pueden usarse. Queda en el log que se pidieron.
- **`created_by_id` distingue quién lo pidió**: si es nulo, lo pidió la persona desde «he olvidado mi
  contraseña»; si tiene valor, lo lanzó Soporte o un administrador. Es la única diferencia entre los
  dos casos, porque el correo es el mismo.

## 3. El token de sesión

- **Firmado con HMAC-SHA256** usando `TOKEN_SECRET`, una variable de entorno **obligatoria en
  producción**: el backend no arranca sin ella, como con `ADMIN_PASSWORD` y `PUBLIC_APP_URL`.
- **Contenido**: `sub` —el identificador de la cuenta, o `admin` para la de fábrica—, `iat` y `exp`.
  **Nada más**: ni el papel, ni el nombre, ni el idioma. Todo eso se lee de la cuenta en cada
  petición (`docs/usuarios-y-permisos.md`, sección 6).
- **Duración de 10 horas**, sin renovación deslizante: el `exp` se pone al firmar y no se toca.
- **Viaja en `Authorization: Bearer …`**, y se guarda en `localStorage` del navegador.
- **Dónde vive el código**: firmar y validar es de `backend/shared/auth`, **no** del módulo. La razón
  es la regla de modularidad: el middleware que protege las rutas vive en `shared/middleware` y
  **`shared` no puede importar de `modules`**. Así que la mecánica del token es transversal y el
  módulo `auth` la usa como la usa el middleware.
- El middleware, en cada petición protegida: valida la firma y la fecha, saca el `sub`, **lee la
  cuenta** y deja en el contexto quién es, qué papel tiene y en qué idioma lee. Si la cuenta está
  desactivada o no existe, responde **401**.
- **El middleware no conoce la tabla de cuentas**: vive en `shared/middleware` y `shared` no puede
  importar de `modules`, así que **recibe una función que carga la identidad** y `main.go` la conecta
  con el servicio de `users`. Es el mismo motivo por el que la mecánica del token está en `shared`:
  mantener la regla de modularidad sin duplicar código.
- **Lo que esto no puede hacer, y conviene saberlo**: como no hay tabla de sesiones, **cambiar la
  contraseña no invalida los tokens ya emitidos**. Quien tuviera uno seguirá dentro hasta que
  caduque, hasta 10 horas. Es el precio de no tener sesiones guardadas, y por eso la duración es de
  horas y no de días.

## 4. Las contraseñas

- **Hash con `bcrypt`**, con coste 12. Nunca se guarda ni se registra una contraseña en claro, ni en
  el log ni en un correo.
- **Política**: mínimo **8 caracteres** —bajado de 12 por el responsable el 2026-09-25—, sin
  caducidad y sin preguntas de seguridad (`docs/usuarios-y-permisos.md`, sección 7). **Nada más**: ni
  mayúsculas, ni números, ni símbolos obligatorios. Cada regla añadida empeora la contraseña que la
  gente acaba eligiendo.
- **La contraseña se comprueba antes de gastar el enlace** (2026-09-25): un enlace de un solo uso que
  se gasta cuando la contraseña no vale deja a la persona en un bucle —pide otro y vuelve a empezar—,
  y el error que acaba de leer ya le decía qué tenía que cambiar. Se comprueba el texto primero
  —y se cifra— y sólo después se gasta el enlace, con el que se busca la cuenta. **Lo encontró una
  prueba que empezó siendo de la política de contraseñas**: al bajar el mínimo de 12 a 8 se probó el
  camino entero, y con una contraseña corta el enlace se perdía.
- **La contraseña de fábrica se compara en tiempo constante** contra `ADMIN_PASSWORD`
  (`subtle.ConstantTimeCompare`), porque no hay hash que comparar: la variable es la fuente de
  verdad y así se evita que el tiempo de respuesta delate coincidencias parciales.
- **Cambiar la contraseña propia exige la actual**, excepto cuando se llega desde un enlace de
  recuperación o de alta, que es justamente el caso en el que no se la sabe.
- **Nunca se dice una contraseña por teléfono ni por correo** (`docs/usuarios-y-permisos.md`,
  sección 7): los dos caminos terminan en un enlace, y el usuario elige.

## 5. Los tres caminos de entrada

### 5.0 Por dónde se entra: un método a la vez

**La instalación entra por un método a la vez**, y lo elige un administrador desde Configuración: sin
eso, tener un directorio configurado significaría que las contraseñas locales dejan de valer sin que
nadie lo haya decidido. Lo que está puesto decide **por dónde se entra**, y los otros dos caminos no
se atienden aunque estén configurados (`docs/modules/settings.md`, sección 5.8).

| Método puesto | Por dónde se entra | Qué se ofrece en la pantalla de entrada |
| --- | --- | --- |
| `local` | Sección 5.1: la contraseña de la base | El formulario y el enlace de recuperarla |
| `ad` | Sección 5.2: el directorio de la organización | **El mismo formulario**, sin el enlace de recuperarla |
| `keycloak` | Sección 5.3: la vuelta por el navegador | El botón que lleva al reino, y **ningún formulario** |

1. **La cuenta de fábrica entra siempre, y va primero**: su correo es `admin` y su contraseña vive en
   la configuración. Es la única puerta que no se puede cerrar —si no, elegir mal el método dejaría la
   instalación sin nadie que lo cambiara— y **con el método en `keycloak` no hay formulario**, así que
   su puerta es un enlace discreto («Entrar como administrador») que lo saca.
2. Después, el método que esté puesto. Sin configuración de la instalación —una instalación recién
   puesta, que todavía no tiene su fila— el método es `local`.
3. **Una cuenta es de un camino y no de dos**: una cuenta de Keycloak no entra por el directorio,
   aunque el método puesto sea el de AD, y una cuenta de AD no entra con una contraseña local.

**Y con el método en `ad` o en `keycloak`, las cuentas locales no entran.** No es un efecto colateral
que se pueda evitar: es lo que quiere decir «un método a la vez», y por eso está dicho aquí y en
`docs/usuarios-y-permisos.md`, sección 5. Volver a `local` las devuelve todas.

### 5.1 Correo y contraseña (local)

**Es el camino de fábrica**, y con el método en `local` es el único que se atiende.

1. Se busca la cuenta por correo —y por usuario, para la de fábrica—, sin distinguir mayúsculas.
   **Lo que llega es un solo campo, `email`**, que además admite `admin` para la cuenta de fábrica: la
   pantalla pide «correo electrónico» y no obliga a nadie a saber que la cuenta de fábrica es distinta.
2. Si es `ad` o `keycloak`, **no se compara nada aquí**: se resuelve por su camino (5.2 y 5.3).
3. Se compara la contraseña con su `bcrypt` y, si todo va bien, se firma el token, se actualiza
   `last_login_at` y se devuelve.

**Un solo mensaje para no acertar y no existir**: `auth.invalidCredentials` para la contraseña
equivocada, para la cuenta que no existe y para la que aún no tiene contraseña establecida. Decir
«esa cuenta no existe» es regalar la lista de quién trabaja aquí.

**Excepción a propósito**: si la cuenta existe y está **desactivada**, el mensaje lo dice
(`auth.accountInactive`), porque esa persona no puede arreglarlo sola y necesita llamar a Soporte.
Es información que se le da a quien **ya ha demostrado saber la contraseña**… salvo en el caso de la
cuenta de directorio, donde no hay contraseña que demostrar: allí se dice lo mismo que en
`auth.invalidCredentials`.

### 5.2 Active Directory

**La pantalla de entrada no cambia para AD**: es el mismo campo de correo y contraseña, y quien
decide por dónde validar es el backend, que es el único que sabe de qué es cada cuenta
(sección 5.1).

**Y nadie da de alta a nadie**: quien está en el directorio **entra directamente**, con la contraseña
que ya tiene, y su cuenta nace en ese primer acceso con lo que el directorio dice de él
(sección 5.4). No hay un alta previa de un administrador ni de Soporte, ni hay que avisar a nadie
para empezar a usarlo: es la corrección del responsable del 2026-09-25 y la razón de ser de este
camino. Lo que sí sigue decidiendo una persona es el **papel** de esa cuenta —`usuario` de entrada,
y otro si le toca—, porque el directorio no sabe nada de eso.


**Cómo se valida**: con una **cuenta de servicio** de sólo lectura. Se busca a la persona y se
valida contra el resultado, porque en un directorio real las unidades organizativas no siguen un
patrón predecible y construir el nombre completo a mano falla el día que alguien cambia de
departamento.

1. El backend se conecta al directorio **con la cuenta de servicio**.
2. Busca a la persona por su correo con el filtro configurado.
3. Si no aparece, `auth.invalidCredentials`.
4. Si aparece, **se conecta con su nombre completo y su contraseña**. Ese intento es la validación: si
   el directorio lo acepta, es quien dice ser.
5. Lee sus atributos y **actualiza o crea la cuenta** (5.4).

**Dónde se configura**: en la pantalla de Configuración, **no en el entorno** (enmienda del
2026-09-25). Los datos viven en la tabla `directory_settings`, los cambia un administrador desde ahí y
`auth` **los lee en cada intento**, así que el cambio vale sin reiniciar nada
(`docs/modules/settings.md`, sección 5.8). Son estos:

| Dato | Para qué |
| --- | --- |
| `host`, `port` | El directorio. **`host` vacío es «esta instalación no tiene directorio»** |
| `use_tls` | Cifrar la conexión. **Sí en producción**: por ahí viajan credenciales |
| `bind_dn`, `bind_password` | La cuenta de servicio. **El secreto se guarda en la base y no se devuelve nunca por la API** |
| `search_base` | Dónde buscar |
| `user_filter` | El filtro, con `%s` donde va el correo |
| `attr_email`, `attr_name`, `attr_last_name`, `attr_id` | Qué atributo es cada dato |

- **El camino de AD se ofrece si el método puesto es `ad`** (sección 5.0), y **no se puede elegir ese
  método sin que el directorio esté configurado**: el backend lo rechaza con
  `settings.method.notConfigured` antes de dejar la instalación sin puerta.
- **Cada método enseña su configuración y sólo la suya** (decisión del responsable, 2026-09-27): en
  Configuración, «Cuentas de la aplicación» no enseña ningún panel, el directorio sale al elegir la
  organización y el reino al elegir Keycloak (`docs/interfaz-y-experiencia.md`, sección 3.8).
- **La configuración se prueba antes de guardarla**: el botón «Probar la conexión» de Configuración
  pide a `auth` que abra la conexión con la cuenta de servicio y la cierre. **No valida la contraseña
  de nadie** —eso sólo se sabe cuando alguien entra— y no toca la base.

**Si el directorio no responde**, el error es propio (`auth.directory.unavailable`) y la pantalla
dice que no se puede entrar **por un problema del directorio**, no que la contraseña esté mal. Son
dos cosas distintas y confundirlas manda a la gente a buscar una contraseña que no ha perdido.

### 5.3 Keycloak

**El callback lo recibe el backend**, no el frontend. Así hay **una sola sesión** que mantener —la
nuestra— y el frontend no sabe que Keycloak existe.

1. El navegador va a `GET /api/auth/keycloak/start`.
2. El backend redirige a Keycloak con un **`state` firmado y con caducidad de 5 minutos**. No se
   guarda en ningún sitio: se firma con `TOKEN_SECRET` y se valida al volver. Es la forma de
   protegerse sin cookies.
3. Keycloak autentica y vuelve a `GET /api/auth/keycloak/callback?code=…&state=…`.
4. El backend valida el `state`, canjea el código por el token de Keycloak y lee el `sub`, el correo
   y el nombre.
5. Actualiza o crea la cuenta (5.4), firma **nuestro** token y redirige al frontend con él **en el
   fragmento de la dirección** (`#token=…`), que no viaja al servidor ni queda en los registros.
6. El frontend lo guarda y **borra el fragmento** de la barra de direcciones.

**Dónde se configura**: igual que el directorio, **en la pantalla y no en el entorno** (enmienda del
2026-09-25): el emisor, el cliente, su secreto y la vuelta viven en la tabla `keycloak_settings`
(`docs/modules/settings.md`, sección 5.8). **Los cuatro van juntos**: sin emisor, sin cliente o sin
vuelta, la configuración está a medias y se rechaza al guardarla (`settings.keycloak.incomplete`), y
**el camino se ofrece si el método puesto es `keycloak`** (sección 5.0). El secreto se guarda en la
base y no se devuelve nunca por la API.

**Cómo está hecho** (as-built, y lo que se decidió al hacerlo):

- **Se habla OIDC con la biblioteca estándar**: se pide el **documento de descubrimiento** del reino
  (`/.well-known/openid-configuration`), se canjea el código contra su `token_endpoint` y se lee a la
  persona en su `userinfo_endpoint`. **No se añade ninguna dependencia** para esto, y el documento se
  pide **una vez y se recuerda**: cambia cuando cambia el reino.
- **El `state` es un valor aleatorio con la fecha, firmado con `TOKEN_SECRET`** (HMAC-SHA256). No se
  guarda en ningún sitio: lo que impide que una vuelta vieja sirva es **su caducidad de cinco
  minutos**, y como la fecha va dentro de lo firmado, estirarla invalida la firma. Un `state` de otro
  secreto, uno retocado o uno inventado no valen, y **sin `state` válido no se canjea nada**.
- **El canje es una petición del servidor a Keycloak**, con el secreto del cliente. Un 4xx del canje es
  «esa vuelta no vale» (`auth.oidc.rejected`); un fallo de red o un 5xx es «Keycloak no responde»
  (`auth.oidc.unavailable`). Son cosas distintas y se contestan distinto.
- **De Keycloak se leen tres cosas**: el `sub`, que es el **identificador externo** y manda sobre el
  correo; el correo, sin el cual no se crea ninguna cuenta —el correo es con lo que se entra y con lo
  que se identifican los tickets—; y el nombre y los apellidos, con el nombre de usuario como última
  salida para que una cuenta nunca quede sin nombre.
- **La vuelta deja a la persona en la pantalla de entrada**, con el token en el fragmento
  (`/login#token=…`) o con **la clave del fallo** (`/login#error=auth.directory…`): es una navegación
  de un navegador en mitad de un recorrido, así que hasta lo que sale mal tiene que acabar en un sitio
  donde se pueda contar qué ha pasado. La pantalla **lee el fragmento, lo usa y lo borra** con
  `replaceState`, para que el token no se quede en la barra de direcciones ni en el historial.
- **Lo que no se hace**: no hay cierre de sesión en Keycloak (nuestra sesión es la nuestra y se sale
  por donde se sale siempre), no se usa el flujo de contraseña directa —que está apagado en el
  cliente— y **no se registran ni el código ni los tokens**: al log sólo va la clave del fallo.
- **Y una consecuencia de lo anterior que hay que conocer**: como no se cierra la sesión en Keycloak,
  **su sesión sigue viva después de salir de la aplicación**. Al volver a pulsar el botón, Keycloak
  reconoce a la persona y **no le pregunta nada**: vuelve sola con un código nuevo y entra. Es el
  comportamiento normal de un SSO, y la prueba de interfaz lo contempla como un caso más, no como un
  fallo.
- **Una cosa que se puede añadir y no se ha añadido**: PKCE. Con un cliente confidencial y el canje
  en el backend, el secreto ya protege el canje; PKCE protegería además contra la interceptación del
  código. Queda **propuesto**, no hecho, y si se hace, el `code_verifier` viajaría dentro del propio
  `state` firmado, que ya existe para eso.

### 5.4 Lo que pasa con la cuenta, en los dos casos

Cuando alguien entra por AD o por Keycloak, y siguiendo las reglas de convivencia ya aprobadas:

| Situación | Qué se hace |
| --- | --- |
| No existe ninguna cuenta con ese correo | Se **crea** con papel `usuario`, origen del directorio y su identificador externo |
| Existe una cuenta local con ese correo | Se **vincula**: pasa a ser del directorio y **su contraseña local deja de servir** |
| Existe ya una cuenta de ese directorio | Se **actualiza**: nombre, apellidos y correo, que es lo que manda el directorio |
| Estaba desactivada y el directorio la tiene activa | Se **reactiva** (regla 6 de la convivencia). En un directorio no hay un «está activa» que se pueda leer sin más: al entrar, lo que manda es que el directorio **acepte sus credenciales**, y eso es lo que se comprueba |
| El correo cambió en el directorio | La cuenta **sigue siendo la misma**: manda el identificador externo, no el correo |

**El directorio manda también en los datos.** Nombre, apellidos y correo se actualizan en cada
entrada: si alguien cambia de apellido en la empresa, aquí se entera solo y nadie tiene que avisar a
Soporte.

**Cuándo se mira al directorio, que es lo que decide cuál de las cinco filas se aplica.** El orden
importa, y desde la enmienda del 2026-09-25 **el método que está puesto es lo primero que decide**:

1. **La cuenta de fábrica** (`admin`), que no está en la tabla: se compara con `ADMIN_PASSWORD`. Va
   antes que nada, con cualquier método.
2. Con el método en **`local`**, la contraseña se compara con la de la base y **no se pregunta al
   directorio**: no se manda a la red algo que ya se sabe en casa. Una cuenta de directorio **no tiene
   contraseña aquí**, así que con este método no entra.
3. Con el método en **`ad`**, **se pregunta al directorio siempre**, sin mirar antes la contraseña
   local: es el camino de la instalación. Si el directorio acepta esas credenciales y su correo era el
   de una cuenta local, esa cuenta pasa a ser del directorio en ese mismo momento y su contraseña
   local deja de servir —es el **vínculo**—. Nadie tiene que avisar a Soporte, ni Soporte tiene que
   tocar la cuenta.
4. Con el método en **`keycloak`**, este endpoint no entra a nadie que no sea la cuenta de fábrica: el
   camino es la vuelta por el navegador (sección 5.3).

**Una cuenta desactivada aquí y activa en el directorio vuelve a entrar** (regla 6). Vale también
para el vínculo: si la cuenta local estaba desactivada, al vincularse se reactiva, porque a partir de
ahí quien manda sobre su acceso es el directorio. Al desactivar una cuenta de directorio la interfaz
ya avisa de que puede volver (`users.directoryMayReturn`), y esto es exactamente eso.

**Si el directorio no responde o rechaza las credenciales**, la respuesta depende de **si se sabe de
dónde es esa persona**, y eso se sabe si hay una cuenta de directorio con ese correo:

| Quién entra | Qué se contesta |
| --- | --- |
| Una cuenta que **ya es de AD**, y el directorio no responde | `auth.directory.unavailable` (503): se sabe de dónde es esa persona, así que se dice que el problema es del directorio y no de su contraseña |
| Una cuenta que **ya es de AD**, y el directorio rechaza esas credenciales | `auth.directory.rejected` (401): es su contraseña, y la contraseña es de allí |
| Un correo que **no tiene cuenta**, o una cuenta local | `auth.invalidCredentials` (401), y el fallo queda en el log. Aquí **no se sabe** si esa persona es del directorio, y contarle a un desconocido cómo está la red de la casa no ayuda a nadie |

**Y la pantalla no cambia por nada de esto**: los dos mensajes viajan como claves y se leen traducidos
en el mismo formulario de siempre (sección 5.2).

### 5.5 Lo que queda fuera al elegir un método

Elegir un método **apaga los otros dos**, y hay que decirlo entero porque tiene consecuencias que se
notan al día siguiente:

| Con el método en… | Quién puede entrar | Quién se queda fuera |
| --- | --- | --- |
| `local` | Las cuentas de la base, con su contraseña | Quien sólo tenga cuenta en el directorio o en el reino |
| `ad` | Quien esté en el directorio | **Las cuentas locales, Soporte y Desarrollo incluidas**; y quien sólo tenga cuenta en Keycloak |
| `keycloak` | Quien tenga cuenta en el reino | **Las cuentas locales, Soporte y Desarrollo incluidas**; y quien sólo tenga cuenta de AD |

- **Nadie pierde su cuenta**: lo que se apaga es la puerta, no la cuenta. Al volver a `local`, todo el
  mundo entra otra vez con lo que tenía.
- **A una cuenta local no se le puede mandar el enlace de recuperación mientras el método no sea
  `local`**: el enlace se establece por un endpoint público, pero entrar con esa contraseña nueva no
  serviría de nada. Es una consecuencia a conocer, no un fallo.
- **La cuenta de fábrica es la excepción y por eso está**: es la que puede volver a poner el método
  donde estaba sin tocar la base a mano.

## 6. La cuenta de fábrica

`admin` **no es una fila**: vive en la configuración
(`docs/usuarios-y-permisos.md`, sección 8). Aquí va su mecánica:

- Al entrar con `admin`, **no se busca en la tabla**: se compara la contraseña con `ADMIN_PASSWORD`
  en tiempo constante. Si coincide, se firma un token con `sub = "admin"`.
- El middleware reconoce ese `sub` y le da la identidad de fábrica: papel `administrador`, lectura de
  todo y escritura sólo en la configuración. No lee ninguna cuenta porque no hay ninguna.
- **No puede cambiar su contraseña desde la aplicación ni recibir correos**: no hay dónde guardarla
  ni correo al que escribir. Cambiarla es editar la variable y reiniciar.
- Es la única cuenta que no aparece nunca como autora de nada.

## 7. Los correos de cuenta

**`auth` no envía correos: se los pide a `mail`**, que es quien tiene los textos. Aquí sólo se dice
cuándo se piden y qué llevan; el texto, los marcadores y el envío son de `docs/modules/mail.md`.

Tres, y ninguno más. Son **cosa distinta de los siete avisos de ticket**
(`docs/propósito-y-alcance.md`):

| Correo | Cuándo | Caducidad del enlace |
| --- | --- | --- |
| **Alta** | Al crear una cuenta `local` | 24 horas |
| **Recuperación** | Al pedirla la persona, o al lanzarla Soporte o un administrador | 1 hora |
| **Aviso de cambio** | Cuando una contraseña cambia: al completar el alta, al usar un enlace, al cambiarla desde dentro | No lleva enlace |

- **En el idioma global de la instalación**: `mail` elige la plantilla de ese idioma.
- **HTML, con una versión de texto automática**, y las plantillas de `mail` con sus marcadores.
- **El texto de la caducidad lo pone `auth`**, no `mail`: las plantillas llevan el marcador
  `{{caducidad}}` («este enlace caduca en …») y quien sabe si son 24 horas o 1 es este módulo, que es
  el que emite el enlace. Por eso `auth` pasa `24 horas` o `1 hora`… en el idioma global, que es
  el mismo en el que va la plantilla: `24 hours`, `1 hour`.
- **La dirección desde la que se cambió sale de `X-Real-IP`**, que es la cabecera con la que nginx
  pasa la dirección real de quien llama: dentro del contenedor, `RemoteAddr` es la del proxy.
- **El aviso de cambio dice cuándo y desde dónde** se cambió, y termina con «si no has sido tú,
  avisa a Soporte». Es la única forma de que alguien se entere de que han entrado en su cuenta.
- Se envían **en segundo plano** y **un fallo de correo no tumba la acción**
  (`docs/arquitectura.md`, sección 9). Si el aviso de cambio no sale, la contraseña ya está cambiada.

## 8. Los endpoints

| Método y ruta | Qué hace | Quién |
| --- | --- | --- |
| `POST /api/auth/login` | Entra con correo y contraseña. Devuelve el token y los datos de la sesión | Cualquiera |
| `GET /api/auth/methods` | **Público.** Cómo se entra en esta instalación: `{"method": "local", "local": true, "ad": false, "keycloak": false}`. Sin ningún dato de nadie | Cualquiera |
| `GET /api/auth/keycloak/start` | Empieza el acceso por Keycloak: **redirige** a su pantalla de entrada con un `state` firmado | Cualquiera |
| `GET /api/auth/keycloak/callback` | Vuelve de Keycloak: **redirige** a la pantalla de entrada con el token —o con la clave del fallo— en el fragmento | Keycloak (el navegador) |
| `POST /api/auth/logout` | Cierra sesión. Como el token no se puede revocar, aquí sólo se registra | Autenticado |
| `GET /api/auth/me` | Quién soy: identificador, nombre, apellidos, correo, papel, origen e idioma | Autenticado |
| `POST /api/auth/password/forgot` | Manda el enlace de recuperación. **Responde igual exista o no la cuenta** | Cualquiera |
| `POST /api/auth/password/reset` | Establece la contraseña con el token del enlace | Cualquiera con enlace |
| `POST /api/auth/password/change` | Cambia la contraseña propia, con la actual | Autenticado, cuenta `local` |

- **`forgot` responde siempre lo mismo**, aunque el correo no exista: si dijera «esa cuenta no
  existe», el endpoint serviría para averiguar quién trabaja aquí.
- **`logout` no invalida nada** (no hay tabla de sesiones) pero se registra en el log: es la forma de
  saber cuándo alguien dijo que se iba.
- **`me` es lo que usa el frontend al arrancar** para saber si la sesión sigue viva y con qué papel.

Lo que devuelven, cuando sale bien:

| Endpoint | Devuelve |
| --- | --- |
| `POST /api/auth/login` | `{"token": "…", "expiresAt": "…", "user": { … }}` |
| `GET /api/auth/keycloak/start` | **302** a Keycloak, con `client_id`, `redirect_uri`, `response_type=code`, `scope` y `state`. Sin Keycloak configurado, **404 `auth.oidc.notConfigured`** |
| `GET /api/auth/keycloak/callback` | **302** a `/login`, con `#token=…` si se ha entrado y `#error=<clave>` si no. **Nunca responde con el token en el cuerpo**: viaja en el fragmento, que no llega al servidor |
| `POST /api/auth/logout` | `{"status": "ok"}`: no hay nada que revocar |
| `GET /api/auth/methods` | `{"method": "local", "local": true, "ad": false, "keycloak": false}`. Dice **el método que está puesto y cuál de los tres caminos está abierto** —uno solo, y la cuenta de fábrica entra igual—. **Mira la configuración, no la salud**: dice `ad: true` con el directorio configurado aunque no responda en ese momento, y lo que pasa entonces se cuenta al entrar |
| `GET /api/auth/me` | `{"user": { … }}`, **el mismo objeto** que devuelve la entrada |
| `POST /api/auth/password/forgot` | `{"status": "ok"}` **exista o no la cuenta** |
| `POST /api/auth/password/reset` | `{"status": "ok"}` |
| `POST /api/auth/password/change` | `{"status": "ok"}` |

Y el objeto `user`, que es el mismo en los tres sitios donde aparece:

| Campo | Qué es |
| --- | --- |
| `id`, `name`, `lastName`, `email` | Quién es |
| `role`, `origin`, `language` | Su papel, de dónde viene su cuenta y en qué idioma lee |
| `factory` | Si es la cuenta de fábrica, que no está en la base |

## 9. El frontend: el armazón

Todo esto vive en `frontend/src/app/core`, que es la excepción ya prevista para la sesión.

| Pieza | Qué hace |
| --- | --- |
| **Servicio de sesión** | Guarda el token, expone la identidad con señales y dice si hay sesión |
| **Interceptor** | Pone la cabecera `Authorization` en cada petición y **atiende el 401**: borra el token y lleva a la pantalla de entrada |
| **Guarda de ruta** | Sin sesión, no se entra: a la pantalla de entrada |
| **Pantalla de entrada** | **Lo que ofrece depende del método puesto** (sección 5.0): el formulario de correo y contraseña con «he olvidado mi contraseña» (`local`), el mismo formulario sin ese enlace y con un aviso de que la contraseña es la de la organización (`ad`), o sólo el botón que lleva al reino (`keycloak`), con la puerta discreta de la cuenta de fábrica. **Abajo, en el pie, se lee la versión del sistema** (`docs/modules/settings.md`, sección 5.9) |
| **Pantalla de olvido** | Pide el correo y confirma que se ha mandado, **sin decir si existía** |
| **Pantalla de establecer contraseña** | Es la del enlace: llega con el token en la dirección y pide la nueva dos veces |
| **Pantalla de cambiar contraseña** | Para quien ya está dentro: actual, nueva y repetir |

- **El interceptor lee el token en cada petición**, no lo guarda en memoria: así, cerrar sesión en una
  pestaña deja sin sesión a las demás en cuanto pidan algo.
- **Un 401 no se maneja en cada pantalla**: se maneja en un sitio, y la pantalla de entrada explica
  que la sesión ha caducado.
- **El servicio de sesión pregunta cómo se entra** (`GET /api/auth/methods`) y **adopta el token que
  trae la vuelta de Keycloak**, que la pantalla de entrada le pasa después de leerlo del fragmento.
  Con esa respuesta la pantalla decide qué pintar: **el formulario si el camino abierto es el local o
  el de AD**, y **el botón de Keycloak sólo si el camino abierto es el suyo** —ofrecer un botón que
  lleva a un error es peor que no ofrecerlo—. Si el backend no contesta, se queda con el camino local,
  que es el que siempre existe.
- **Y con el método en `keycloak`, la pantalla de entrada enseña la puerta de la cuenta de fábrica**:
  un enlace discreto al pie que saca el formulario de correo y contraseña. Sin él, una instalación a la
  que se le elige mal el método se quedaría sin nadie que pudiera volver a cambiarlo (sección 5.0).
- **El botón de Keycloak no es un envío del formulario**: es llevar el navegador a
  `/api/auth/keycloak/start`, porque quien autentica es Keycloak y hay que irse con la navegación de
  verdad. Se llama igual en el inventario —un botón secundario, con su texto en los dos idiomas— y
  como no envía nada, no puede dispararse con el Enter del formulario.
- El token se guarda en `localStorage` (`docs/usuarios-y-permisos.md`, sección 6) y **no se enseña
  en ninguna pantalla**.
- Los textos de estas seis pantallas van en **español y en inglés**, como todo lo demás, y **el
  idioma de arranque sale del navegador**: si el navegador pide `es` o `en` —o cualquiera de sus
  variantes, como `es-MX`— se entra en ese idioma; **si pide cualquier otro, se entra en inglés**,
  que es el segundo de la instalación y el que más gente puede leer. El conmutador está a la vista en
  la pantalla de entrada, y lo que alguien elija se recuerda en su navegador. Dentro de la
  aplicación manda el idioma global de la instalación, que también decide el idioma de los correos
  (sección 7).
- **Los textos viven en `core/i18n/es.ts` y `core/i18n/en.ts`**, con el mismo tipo compartido: **una
  clave que falte en un idioma no compila**. Es la única forma de que la regla «cada clave necesita
  su texto en los dos idiomas» no dependa de que alguien se acuerde, y por eso los textos de error
  van en el mismo diccionario que los de las pantallas.

**Las seis pantallas del armazón**, que son estas y ninguna más:

| # | Pantalla | Qué es |
| --- | --- | --- |
| 1 | Entrada | **Lo que ofrece depende del método puesto** (sección 5.0): el formulario con «he olvidado mi contraseña», el mismo formulario con el aviso de que la contraseña es la de la organización, o sólo el botón de Keycloak con la puerta de la cuenta de fábrica |
| 2 | Olvido | Pide el correo y confirma que se ha mandado, **sin decir si existía** |
| 3 | Establecer contraseña | La del enlace: llega con el token en la dirección y pide la nueva dos veces |
| 4 | Cambiar contraseña | Para quien ya está dentro: actual, nueva y repetir |
| 5 | **Sin permiso** | Alguien llega a algo que su papel no alcanza: se le dice, no se le enseña en blanco |
| 6 | **Servidor caído** | La aplicación no puede hablar con su backend: un aviso claro, sin tecnicismos |

- **Las dos últimas no son «pantallas de auth»**, son estados del armazón
  (`docs/interfaz-y-experiencia.md`, sección 9): viven en `core` y las usan todos los módulos. La 5
  la provoca un 403 y también una guarda; la 6, cualquier respuesta que diga que el servidor no está
  (o ninguna respuesta, que es lo que pasa cuando se cae del todo).

Las rutas de las cuatro primeras, **en inglés y en minúsculas** (corregido por el responsable el
2026-09-23: las rutas van siempre en inglés aunque el idioma del producto sea español):

| Pantalla | Ruta |
| --- | --- |
| Entrada | `/login` |
| Olvido | `/forgot-password` |
| Establecer contraseña (el enlace del correo) | `/set-password` |
| Cambiar contraseña (desde dentro) | `/change-password` |
| Sin permiso | `/forbidden` |
| Servidor caído | Sin ruta propia: es un aviso que se pone **encima** de lo que hubiera en pantalla, porque puede pasar en cualquier momento y en cualquier pantalla |

- **El enlace del correo apunta a `/set-password` con el token en el fragmento**:
  `…/set-password#token=…`. El fragmento **no se manda al servidor**, así que el token no acaba en el
  registro de accesos de nginx ni en ningún log, y la pantalla lo borra de la barra de direcciones al
  usarlo —igual que con la vuelta de Keycloak (decisión 13)—.
- **Una sola pantalla sirve para los dos enlaces**, el de alta y el de recuperación: en los dos casos
  se llega sin saber la contraseña y se sale con una puesta. Cambia el texto, no el mecanismo.

## 10. Las claves de error

| Clave | Cuándo |
| --- | --- |
| `auth.invalidCredentials` | Contraseña equivocada, cuenta que no existe o cuenta sin contraseña establecida |
| `auth.accountInactive` | La cuenta existe y está desactivada (cuentas locales) |
| `auth.directory.unavailable` | El directorio no responde |
| `auth.directory.rejected` | El directorio ha rechazado las credenciales |
| `auth.token.expired` | El enlace del correo ha caducado |
| `auth.token.used` | El enlace ya se usó, o lo sustituyó otro más nuevo |
| `auth.password.tooShort` | Menos de 8 caracteres. **No gasta el enlace**: se puede volver a intentar con el mismo |
| `auth.password.wrong` | La contraseña actual no es la que se ha escrito |
| `auth.password.notLocal` | Intento de cambiar la contraseña de una cuenta de directorio |
| `auth.oidc.state` | El `state` del acceso por Keycloak no vale o ha caducado |
| `auth.oidc.notConfigured` | Esta instalación no tiene Keycloak, y se ha pedido su camino a mano |
| `auth.oidc.rejected` | Keycloak no ha aceptado la vuelta: código caducado, ya usado, de otra instalación, o alguien que canceló en su pantalla |
| `auth.oidc.unavailable` | Keycloak no responde: no se pudo leer su reino, ni canjear el código, ni saber quién es |
| `auth.oidc.noEmail` | La cuenta de Keycloak no trae correo, y sin correo no se puede crear la cuenta aquí |
| `auth.session.invalid` | Falta la cabecera `Authorization`, el token no vale o su cuenta ya no existe o está desactivada |
| `auth.session.expired` | El token de sesión ha pasado de sus 10 horas |
| `auth.forbidden` | La cuenta ha entrado bien, pero su papel no llega para esa acción |

Y con qué código HTTP viaja cada una, para que el frontend no tenga que adivinarlo:

| Clave | Código |
| --- | --- |
| `auth.session.invalid`, `auth.session.expired` | **401** |
| `auth.forbidden` | **403** |
| `auth.invalidCredentials`, `auth.accountInactive`, `auth.directory.rejected`, `auth.password.wrong`, `auth.password.notLocal`, `auth.password.tooShort`, `auth.token.expired`, `auth.token.used` | **401** |
| `auth.directory.unavailable`, `auth.oidc.unavailable` | **503**, porque el fallo es del servidor y no de quien escribe la contraseña |
| `auth.oidc.notConfigured` | **404**: en esta instalación ese camino no existe |
| `auth.oidc.state`, `auth.oidc.rejected`, `auth.oidc.noEmail` | **En la vuelta de Keycloak no hay código HTTP que valga**: el navegador tiene que acabar en algún sitio, así que viajan en el fragmento de la dirección y la pantalla de entrada las traduce |

- **`auth.forbidden` es transversal**: la comprobación vive en `backend/shared/authz`
  (`docs/usuarios-y-permisos.md`, sección 10) y la usan todos los módulos, pero la clave es una sola
  y el frontend la traduce una vez. Es la única clave de esta tabla que no nace en el módulo `auth`.
- **Si la lectura de la cuenta falla** —la base de datos no responde—, la petición no se rechaza por
  sesión: se registra con `go-logs` y se responde **503 con `error interno`**, igual que el
  middleware de pánico. Un fallo del servidor no es una sesión inválida.

**`auth.session.invalid` cubre los tres casos a propósito** (sin cabecera, firma que no vale y
cuenta que ya no está o está desactivada): el frontend hace lo mismo en los tres —borrar el token y
llevar a la pantalla de entrada—, y separarlos sólo añadiría claves que no cambian ninguna pantalla.

## 11. Con qué se prueba

Con **dos servicios de pruebas**, cada uno en **su propio archivo** y **opcional** —el entorno no los
levanta—: `active-directory.yml` (OpenLDAP) y `keycloak.yml` (Keycloak). **Primero se levanta el
entorno**, que es quien crea la red del desarrollo, y después el servicio que se quiera:

```bash
docker compose -f dev.yml up -d                  # el entorno (crea la red `catalina-support-dev`)
docker compose -f active-directory.yml up -d     # el directorio de pruebas (AD/LDAP)
docker compose -f keycloak.yml up -d             # Keycloak
```

| Servicio | Qué es | Qué se prueba con él |
| --- | --- | --- |
| **OpenLDAP** | Un directorio con **tres** personas de prueba y su configuración **en el repositorio** (`config/ldap/01-personas.ldif`) | El camino de AD: buscar, validar y leer atributos. **Está hecho y verificado** |
| **Keycloak** | Con **un reino importado desde el repositorio** (`config/keycloak/realm-catalina-support.json`) y sus tres personas de prueba | El camino de OIDC de principio a fin. **Está hecho y verificado** |

Las tres personas de `config/ldap/01-personas.ldif` están puestas a propósito para que cada caso de la
sección 5.4 se pueda provocar a mano: una que **no tiene cuenta** en la aplicación (el alta
automática), otra que no se usa en ninguna prueba y una tercera **cuyo correo es el de una cuenta
local** (el vínculo).

**Aviso que hay que tener presente**: OpenLDAP **no es Active Directory**. Se habla con él de la
misma manera —buscar, conectar, validar—, así que el código se prueba de verdad, pero **lo
específico de AD no queda cubierto**: ni sus referencias entre dominios, ni sus atributos
particulares, ni sus reglas de contraseña. Cuando haya un AD real al que apuntar, se prueba contra él
y se anota lo que salga.

**El contenedor del directorio no se para y se arranca**: sus scripts de arranque no son idempotentes
—al volver a arrancarlo intenta copiar los LDIF a una carpeta que ya no está y muere—, así que un
`stop` deja el camino de AD caído con el contenedor en `Exited`. Se levanta otra vez recreándolo, y
como el directorio no guarda nada en un volumen, **vuelve al contenido del LDIF del repositorio**:

```bash
docker compose -f active-directory.yml up -d --force-recreate ldap
```

Eso es cómodo para probar (el directorio siempre está como dice el repositorio) y hay que tenerlo
presente: **lo que se cambie a mano dentro del contenedor se pierde al recrearlo**. Las pruebas de
interfaz del camino de AD se saltan solas si el directorio no responde
(`tests/e2e/specs/directorio.spec.ts`), así que la suite entera se puede correr sin él.

**Y las pruebas de los dos caminos ponen el método y lo devuelven** (enmienda del 2026-09-25): desde
que la instalación entra por uno solo (sección 5.0), probar el de AD o el de Keycloak exige elegirlo
primero. Cada suite lo pone al empezar y **lo deja en `local` al terminar, pase lo que pase**, con la
cuenta de fábrica, que entra siempre. Se pregunta si el servicio responde **por el botón de «Probar la
conexión» de Configuración**, que es lo que pregunta una persona y no depende del método puesto.

**Las tres personas del reino de Keycloak** están puestas por los mismos motivos que las del
directorio: una que **no tiene cuenta** aquí (el alta automática), otra que vuelve a entrar y una
tercera **cuyo correo es el de una cuenta local** (el vínculo). El reino trae también el **cliente**
con el que entra la aplicación: confidencial —con secreto, porque el canje lo hace el backend—, con
el flujo de código encendido, **el de contraseña directa apagado** y la única dirección de vuelta que
se admite.

**Cómo llega el navegador a Keycloak.** En local se publica sólo en `127.0.0.1:11006`,
con emisor `http://127.0.0.1:11006/sso/realms/catalina-support`. El backend conecta mediante
`internalIssuer = http://keycloak:8080/sso/realms/catalina-support`; el emisor anunciado conserva
la identidad pública. La autorización es pública y sólo el tráfico del backend utiliza la dirección
interna. El cliente permite vueltas exactas a `http://127.0.0.1:11001/api/auth/keycloak/callback`
y `http://frontend.localhost:11001/api/auth/keycloak/callback` para la suite aislada, sin comodines.
No se necesita nginx. `KEYCLOAK_PUBLIC_URL` permite configurar el origen anunciado.

El perfil `directory` de `tests.yml` levanta sus propios LDAP y Keycloak, sin publicar puertos
ni usar los servicios del usuario. Allí navegador y backend alcanzan `http://keycloak:8080/sso`,
por lo que la dirección interna puede permanecer vacía. Los tres usuarios del reino se conservan.

**Una consulta más, que no es entrar pero vive aquí**: «¿sigues conociendo a esta persona?». La
declara `users`, que es quien la usa —para no devolverle el acceso a una cuenta de AD que ya no está
en el directorio—, y la cumple el cliente del directorio de este módulo con **la misma cuenta de
servicio y el mismo filtro** con los que se entra: la pregunta «¿sigue estando?» tiene que contestarse
igual que «¿puede entrar?» (`docs/modules/users.md`, sección 5, punto 4). **No la ofrece el camino de
Keycloak**: preguntarlo pediría una cuenta de servicio en el reino, y por eso la reactivación a mano
de esas cuentas se rechaza diciendo que vuelven solas al entrar.

## 12. Lo que se decidió al repasar este documento

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **Los directorios de prueba** | OpenLDAP y Keycloak, cada uno en su archivo (`active-directory.yml` y `keycloak.yml`), opcionales y con su propio comando. **Cambiado el 2026-10-01**: antes eran servicios de `dev.yml` detrás del perfil `auth`, que desaparece |
| 2 | **La vuelta de Keycloak** | La recibe **el backend**, que entrega nuestro token al frontend |
| 3 | **Los datos de la cuenta** | El directorio **actualiza nombre, apellidos y correo** en cada entrada |
| 4 | **Intentos fallidos** | **Sin bloqueo de cuenta**: nginx limita y cada fallo va al log |
| 5 | **Auditoría de accesos** | **Sólo en los logs**, sin tabla |
| 6 | **Enlaces** | Pedir uno nuevo **borra los anteriores** |
| 7 | **Aviso de cambio** | Sí: cuando una contraseña cambia, su dueño recibe un correo |
| 8 | **El token de sesión** | Sin estado, 10 horas, `sub` con el identificador, la cuenta se lee en cada petición |
| 9 | **El transporte** | Cabecera `Authorization`, token en `localStorage`, sin cookies |
| 10 | **El coste de `bcrypt`** | 12 |
| 11 | **La mecánica del token** | Vive en `shared/auth`: el middleware no puede importar de `modules` |
| 12 | **`auth.accountInactive`** | Sólo en cuentas locales: en las de directorio no hay contraseña que demostrar |
| 13 | **El token de Keycloak** | Vuelve **en el fragmento** de la dirección, y el frontend lo borra al guardarlo |
| 14 | **El log de los intentos** | Cuenta, resultado, IP y fecha. **Nunca la contraseña**, ni siquiera la equivocada |
| 15 | **Las claves del token de sesión** | `auth.session.invalid` y `auth.session.expired`, aparte de las de los enlaces, que se quedan como estaban |
| 16 | **El código de cada error** | 401 en todos salvo `auth.directory.unavailable`, que es 503 |
| 17 | **La clave del 403** | `auth.forbidden`, una sola para todos los módulos, comprobada en `shared/authz` |
| 18 | **Fallo al leer la cuenta** | No es un 401: 503 con `error interno`, porque el fallo es del servidor |
| 19 | **El campo de entrada** | Uno solo, `email`, que también admite `admin` para la cuenta de fábrica |
| 20 | **Lo que devuelve la entrada** | `token`, `expiresAt` y un objeto `user`; `me` devuelve el mismo objeto `user`, para no tener dos formas de describir a alguien |
| 21 | **La ruta del enlace** | `/set-password`. **Corrección del responsable**: propuse la ruta en español y decidió que **las rutas del frontend van siempre en inglés**, aunque el idioma del producto sea español |
| 22 | **El token del enlace** | Viaja en el **fragmento** (`#token=…`), que el navegador no manda al servidor, y la pantalla lo borra de la dirección al usarlo |
| 23 | **Una sola pantalla para los dos enlaces** | Alta y recuperación llegan a `/set-password`, y la caducidad la escribe `auth` en el idioma global |
| 24 | **Las seis pantallas** | Las cuatro de rutas más **sin permiso** (`/forbidden`) y **servidor caído** (un aviso encima, sin ruta) |
| 25 | **El idioma de arranque** | **Corrección del responsable**: propuse español por defecto y decidió que se toma **el del navegador** y, si no es `es` ni `en`, se entra **en inglés**, con conmutador a la vista |
| 26 | **Dónde viven los textos** | `core/i18n/es.ts` y `core/i18n/en.ts` con un tipo compartido: **una clave que falte no compila**, y los textos de error van en el mismo diccionario |
| 27 | **Cómo sabe la pantalla qué caminos hay** | **Un endpoint público**, `GET /api/auth/methods`, que dice si esta instalación tiene AD y Keycloak. El botón de Keycloak se enseña **sólo si su camino existe**: ofrecer un botón que lleva a un error es peor que no ofrecerlo |
| 28 | **Sin directorio configurado no hay camino de AD** | **Enmendada el 2026-09-25**: ya no es «sin `LDAP_HOST`», que era una variable del entorno, sino **sin directorio configurado en la base** —y además el método tiene que ser `ad`—. Sigue sin haber un interruptor aparte: una instalación sin directorio no tiene que desactivar nada |
| 29 | **El camino de AD primero, y Keycloak después** | Se implementan por separado, cada uno con su contenedor de pruebas y su verificación: así un fallo se achaca a un camino y no a dos |
| 30 | **El botón de Keycloak, en la pantalla de entrada** | La vuelta de Keycloak entrega el token **a la misma pantalla donde empezó todo** (`/login#token=…`), que lo guarda y borra el fragmento (decisión 13) |
| 31 | **Entrar sin alta manual** | **Corrección del responsable** (2026-09-25), al llevarle el vínculo como pregunta: «con AD o con Keycloak quisiera que los usuarios puedan ingresar directamente, sin el proceso de alta manual por un administrador o técnico». Quien está en el directorio entra y su cuenta nace, se vincula o se pone al día sola. Lo que un administrador sigue repartiendo es el **papel**: el directorio no lo dice |
| 32 | **Cuándo se dispara el vínculo** | Cuando la contraseña local de una cuenta **no vale** y el directorio conoce ese correo: se vincula en esa misma entrada y la contraseña local deja de servir. Es lo que hace que la fila «existe una cuenta local con ese correo» se cumpla sin que nadie avise a Soporte |
| 33 | **La pantalla de entrada cambia poco por AD** | **Enmendada el 2026-09-25**: el formulario **es el mismo** y no hay botón propio, pero ahora sí lleva un aviso —«esta instalación entra con las cuentas de la organización»— y **se quita el enlace de «he olvidado mi contraseña»**, porque con el método en AD esa contraseña no es de aquí y el enlace mandaría un correo que no sirve para entrar. **Con Keycloak la pantalla sí cambia del todo**: no hay formulario |
| 34 | **`methods` no sondea el directorio** | **Enmendada el 2026-09-25**: dice el método que está puesto, no la salud de nadie. Sondear el directorio en cada carga de la pantalla costaría una conexión por visita y un directorio lento retrasaría la entrada; el aviso honesto llega al entrar, que es cuando importa, y **la prueba de la conexión está en Configuración**, que es donde se configura |
| 36 | **El emisor es el que ve el navegador** | En desarrollo, Decisión histórica, sustituida por la enmienda local del 2026-10-03: Keycloak se publicaba en un camino del mismo dominio que la aplicación (`/sso/`) y el backend usa ese mismo emisor. Se gana que haya **una sola dirección** y que el `iss` cuadre siempre; se pierde una ida y vuelta de red del backend a su propio dominio en cada canje, que es irrelevante al lado de tener dos direcciones que puedan discrepar |
| 37 | **El `state` no se guarda en ningún sitio** | Es un valor aleatorio con su fecha, firmado con `TOKEN_SECRET`: lo que impide que una vuelta vieja sirva es su caducidad de cinco minutos, y la fecha va dentro de lo firmado. Sin cookies y sin tabla de estados pendientes, que es la misma decisión que con la sesión |
| 38 | **La vuelta deja a la persona en la pantalla de entrada, también cuando falla** | Con el token (`#token=…`) o con la clave del fallo (`#error=…`). Es una navegación de un navegador: hasta lo que sale mal tiene que acabar en un sitio donde se pueda contar qué ha pasado. La pantalla lee el fragmento y **lo borra con `replaceState`** |
| 39 | **Cuatro claves de error nuevas** | `auth.oidc.notConfigured` (404: esta instalación no tiene ese camino), `auth.oidc.rejected` (la vuelta no vale: código caducado, usado, de otra instalación, o alguien que canceló), `auth.oidc.unavailable` (503: Keycloak no responde) y `auth.oidc.noEmail` (la cuenta del directorio no trae correo, y sin correo no hay cuenta aquí). `auth.oidc.state`, que ya estaba, es la quinta |
| 40 | **PKCE, propuesto y no hecho** | Con un cliente confidencial y el canje en el backend, el secreto ya protege el canje; PKCE protegería además contra la interceptación del código. **Queda propuesto**: si se hace, el `code_verifier` viajaría dentro del `state` firmado, que ya existe. No se ha hecho porque no estaba en el documento aprobado y no cambia nada de lo que se ve |
| 44 | **Con el método en AD no se ofrece recuperar la contraseña** | **Confirmado por el responsable el 2026-09-25**: con el método en AD el enlace desaparece, porque esa contraseña no es de aquí y un correo con un enlace para establecerla sería mentira —entrar con ella no serviría—. Con el método local se queda como estaba, y con el de Keycloak tampoco se ofrece |
| 45 | **La puerta de la cuenta de fábrica es discreta** | **Confirmado por el responsable el 2026-09-25**: un enlace de texto al pie de la pantalla de entrada, que saca el formulario cuando el método no lo tiene. No se le ofrece a quien entra todos los días, pero está donde hay que buscarla, y es lo que impide que elegir mal el método cierre la instalación |
| 41 | **Un método de entrada a la vez** | **Decisión del responsable, 2026-09-25**: los tres caminos **no conviven**; el que está puesto es el único que se atiende, y se puede cambiar cuando se quiera. **La cuenta de fábrica entra siempre**, por su cuenta y con su puerta discreta en la pantalla de entrada cuando el método no tiene formulario. Consecuencia dicha y aceptada: con el método en `ad` o en `keycloak` **las cuentas locales no entran**, y eso enmienda la sección 5 de `docs/usuarios-y-permisos.md`, que describía los tres caminos conviviendo |
| 42 | **El directorio y el reino se leen en cada intento** | **Consecuencia de la decisión anterior, y su precio**: cambiar el método o el directorio vale en el intento siguiente, sin reiniciar nada, a cambio de tres lecturas por clave primaria en cada entrada. El cliente de Keycloak se construye por intento con la configuración vigente; el documento se recuerda únicamente durante ese intento |
| 43 | **Una cuenta es de un camino, y no de dos** | Una cuenta de Keycloak **no entra por el directorio** aunque el método puesto sea el de AD, y una cuenta de AD no entra con una contraseña local con el método en `local`. Lo que cambia el método es **por dónde se entra**, no la naturaleza de cada cuenta |
| 35 | **Una cuenta desactivada aquí vuelve por el directorio** | Vale también para el vínculo: si la cuenta local estaba desactivada, al vincularse se reactiva. Es la regla 6, y es lo que hace cierto el aviso `users.directoryMayReturn`. Con esto **queda cerrada la duda que quedaba abierta** en `docs/modules/users.md` sobre si ese aviso se podía ejercitar: se ejercita, y hay prueba de interfaz |

## 13. Qué habilita este documento

Con `docs/modules/auth.md` aprobado, **`auth` y `users` ya se pueden implementar**, y conviene
hacerlos juntos: no se puede tener cuentas sin poder entrar, ni entrar sin cuentas. El orden que
propongo es el que hace que cada paso se pueda probar solo:

0. **`mail`**, que es de quien depende `auth` para mandar el correo de alta: sin él, el alta no se
   puede completar. Va primero.
1. `shared/auth` (firma y validación) y el middleware, con sus pruebas. **Hecho.**
2. La tabla `password_tokens` y las cuentas, en `v1.0.0.sql`. **Hecho.**
3. El camino local completo: alta, enlace, entrada, cambio y reseteo. **Hecho y verificado.**
4. El armazón del frontend: servicio, interceptor, guarda y las seis pantallas. **Hecho.**
5. Los contenedores de pruebas y el camino de AD. **Hecho y verificado** (2026-09-25).
6. El camino de Keycloak. **Hecho y verificado** (2026-09-25). **Con esto, este documento no tiene
   nada pendiente.**

## 14. Autenticación sin idioma personal

La identidad compartida, las respuestas de entrada y `me`, y el estado de sesión dejan de incluir
`language`. El frontend obtiene el idioma de la marca pública y no vuelve a cambiarlo al recibir una
cuenta. La pantalla de entrada retira su selector; se pinta directamente en el idioma global.

Las respuestas autenticadas incluyen las cabeceras `X-Catalina-Language` y
`X-Catalina-Settings-Version`. El interceptor ya existente las observa y, cuando la versión cambia,
actualiza la señal global de idioma. Esto no cambia el token, no cierra la sesión y no añade sondeo:
una pestaña abierta adopta el cambio en su siguiente petición.

Los correos de alta, olvido y cambio de contraseña consultan el idioma global al enviarse. El texto
de caducidad se forma con ese mismo idioma. No cambian token, duración, contraseñas, cuenta de
fábrica, AD, Keycloak, permisos ni rutas. La propuesta fue confirmada por el responsable el
2026-10-06 y en el repaso eligió actualizar las sesiones abiertas con las cabeceras de idioma y
versión, sin cerrar sesión. Se implementó y verificó el 2026-10-07.
