# settings

> **Estado:** as-built
> **Última actualización:** 2026-09-27
>
> **Enmendado el 2026-09-27**, a petición del responsable: en «Método de autenticación», **cada método enseña
> sólo su configuración** —«Cuentas de la aplicación» no enseña ninguna, el directorio sale con la
> organización y el reino con Keycloak—, en vez de las dos a la vez (decisión 10-bis).
>
> **Enmendado el 2026-09-26**, al reemplazar el responsable **los archivos de la marca de fábrica**:
> el **favicon** y **dos logos**, uno para los temas claros y otro para los oscuros. Eso convierte la
> marca de fábrica en **dos archivos por tema** —antes era uno solo, que se usaba en los ocho—, así
> que la sección 5.1, la 5.6 y la decisión 4 cambian. **Los archivos de fábrica no se configuran desde
> la pantalla**: se reemplazan en `frontend/public/`, y eso lo hace quien mantiene la aplicación
> (decisión del responsable, 2026-09-26).
>
> **Enmendado el 2026-09-25**, al pedir el responsable que **se vea la versión del
> sistema**: es una constante del backend que viaja en la marca pública y se lee en el menú lateral y
> en el pie de la pantalla de entrada (sección 5.9, decisión 12). **Eligió el pie de la entrada, que no
> era mi recomendación**: yo proponía sólo el menú.
>
> **Enmendado el 2026-09-25**, al pedir el responsable que **el directorio de la organización y
> Keycloak se configuren desde la pantalla** en lugar de por variables de entorno, y que **la
> instalación entre por un método a la vez**: entran en la lista cerrada de lo configurable (sección
> 2), estrenan **dos tablas de una fila** —`directory_settings` y `keycloak_settings`— y su sección
> (5.8). **Los secretos se guardan en la base y no salen nunca por la API**, y las variables
> `LDAP_*` y `OIDC_*` **desaparecen del entorno** (decisión 10). Se enmienda con esto
> `docs/usuarios-y-permisos.md` (sección 5) y `docs/modules/auth.md` (decisión 28), que describían los
> tres caminos como coexistentes.
>
> **Enmendado el 2026-09-25**, al pedir el responsable que **se pueda configurar el nombre de la
> institución, la empresa o el equipo** en lugar del texto «Catalina Support»: entra en la lista
> cerrada de lo configurable (sección 2), estrena columna propia —en `v1.0.0.sql`, junto al resto del
> esquema— y su sección (5.7), que dice dónde se lee y qué **no** cambia. **El
> responsable corrigió dos cosas de mi propuesta**: los correos se quedan como están y el remitente
> sigue siendo el de la variable del entorno (decisión 9).
>
> Escrito al pedir el responsable que **el logo de la instalación se pueda reemplazar** desde la
> pantalla de Configuración: el logo, el idioma, el color institucional, el prefijo de la numeración y
> el reparto de los tickets **son la configuración de la instalación**, y su dueño es este módulo, que
> hasta ahora estaba sin documento.
>
> **Implementado el 2026-09-24** (la marca y el color institucional, que es lo que se aprobó para
> este tramo): las dos tablas, los seis endpoints, la validación del logo por contenido, el endpoint
> público con versión y el aislamiento del SVG, el logo en la pantalla de entrada y el color
> institucional aplicado a los dos temas de fábrica.
>
> **Y el armazón y la pantalla de Configuración, también** (2026-09-24): el menú lateral con el logo
> en su zona de producto, y la pantalla con **la marca** —los dos huecos con su vista previa, subir y
> volver al de fábrica— y **el color institucional** con **vista previa pedida al backend** antes de
> guardar. El prefijo, el reparto y el idioma llegarán con sus consumidores.
>
> **Aprobado por el responsable el 2026-09-24**, tras contestar las siete preguntas del repaso: el
> límite del logo (1 MB y 2000 píxeles), qué hace el idioma de la instalación (no manda sobre la
> interfaz), que el color institucional sólo afecta a los dos temas de fábrica, los dos huecos de logo
> opcionales, **la enmienda a `docs/modules/tickets.md`** (`ticket_number_settings` pasa a
> `ticket_settings`), que el logo no va en los correos y que su endpoint es público.

## 1. Alcance de este documento

Cuenta **qué es configurable en una instalación de Catalina Support**, dónde vive esa configuración,
cómo la leen los demás módulos, y **la marca**: el logo de la instalación, cómo se sube, dónde se
guarda, cuál se enseña en cada tema y cómo se vuelve al de fábrica.

**No repite** lo que ya está decidido en otro sitio:

| Qué | Dónde está |
| --- | --- |
| Los temas y el color primario | `docs/interfaz-y-experiencia.md`, sección 6.2 |
| El prefijo de la numeración y el reparto de los tickets | `docs/modules/tickets.md`, secciones 2.2 y 3.3 |
| Las plantillas de correo | `docs/modules/mail.md` (son del módulo `mail`, no de aquí) |
| Quién puede tocar la configuración | `docs/usuarios-y-permisos.md`, sección 3: **sólo un Administrador** |
| Dónde está la pantalla | `docs/interfaz-y-experiencia.md`, sección 5 |
| Los tres caminos de entrada, y cómo entra cada uno | `docs/modules/auth.md`, secciones 5.1 a 5.4. **Aquí sólo se dice dónde se configuran** |

## 2. Qué es configurable, y qué no

**La lista es cerrada.** Lo que no está aquí no se configura desde la aplicación: o es del servidor
(las variables de entorno), o es de cada persona, o no existe.

| Qué | Valores | De fábrica |
| --- | --- | --- |
| **El nombre de la instalación** | texto, **hasta 60 caracteres**; vacío devuelve el de fábrica | `Catalina Support` |
| **Idioma de la instalación** | `es` o `en` | `es` |
| **Color primario institucional** | un color, en hexadecimal | el azul institucional |
| **La marca**: el logo | hasta **dos** archivos: uno para los temas claros y otro para los oscuros, **los dos opcionales** | ninguno: se usan los dos logos de fábrica, uno por tema |
| **El prefijo de la numeración** | mayúsculas y dígitos, de 2 a 8 caracteres | `CS` |
| **La asignación de los tickets nuevos** | `ninguna` o `por_turnos`, **por cada tipo de ticket** | principales `por_turnos`; internos `ninguna` |
| **El aviso de ticket nuevo** | `a_nadie`, `a_todo_el_equipo` o `al_asignado`, **por cada tipo de ticket** | principales `al_asignado`; internos `a_nadie` |
| **El método de entrada** | `local`, `ad` o `keycloak`. **Uno a la vez** | `local` |
| **El directorio de la organización** | servidor, puerto, TLS, cuenta de servicio **con su contraseña**, dónde se busca, el filtro y cuatro atributos | vacío: no hay camino de AD |
| **Keycloak** | emisor del reino, cliente, **su secreto** y la dirección de vuelta | vacío: no hay camino de Keycloak |
| **La región horaria** | un **nombre de zona IANA** (`America/Guayaquil`, `Etc/GMT+5`…), de la lista de zonas con buscador. Decide **cómo se leen** las fechas —interfaz y correos—; las guardadas siguen en UTC (decisión 15) | `UTC` |
| **La dirección pública** | `http` o `https`, host y **puerto** opcional; **`localhost` vale**. Es la base de los enlaces de los correos y de la vuelta de Keycloak (decisión 14) | vacío, y entonces se usa `PUBLIC_APP_URL` |

- **El prefijo y el reparto ya estaban decididos** en `docs/modules/tickets.md`; aquí sólo se dice
  **dónde viven**, que es lo que allí faltaba.
- **Los valores viajan en inglés y sin espacios** (`por_turnos`, `al_asignado`) aunque la interfaz
  los enseñe en español: es un valor de un contrato, no un texto, y los textos los pone el
  diccionario del frontend (`docs/interfaz-y-experiencia.md`, sección 8).
- **El idioma de la instalación es el de las cuentas nuevas** y el que usa la instalación para lo
  suyo. **No manda sobre la interfaz de nadie**: quien ha entrado lee en el idioma de su cuenta, y
  antes de entrar manda el navegador —y si no pide español ni inglés, se entra en inglés
  (`docs/interfaz-y-experiencia.md`, principio 6)—. Es lo que evita que una instalación en español le
  cambie el idioma a quien lee en inglés.
- **El color institucional sólo cambia los dos temas de fábrica** (Claro y Oscuro), que son los
  personalizables. Los seis temas fijos llevan su propio acento, aprobado el 2026-09-23
  (`docs/interfaz-y-experiencia.md`, sección 6.2).
- **El directorio y Keycloak dejan de ser variables de entorno** (decisión del responsable,
  2026-09-25): en una instalación que se entrega, configurar un AD pidiendo que se edite un archivo en
  el servidor y se reinicie el contenedor no es configurar. **Sus secretos se guardan en la base** —es
  el único sitio donde un administrador los puede cambiar sin entrar por SSH— y **se enseñan una sola
  vez: al escribirlos**. Lo que devuelve la API es si hay uno puesto, nunca el valor: eso lo fija la
  sección 5.8 y lo comprueba una prueba.
- **El método de entrada se elige, y los otros dos quedan apagados.** No se elige uno que no esté
  configurado: eso dejaría la instalación sin puerta para todo el mundo menos la cuenta de fábrica, y
  el backend lo rechaza con su clave (sección 7).

**Lo que NO se configura aquí**, para que este módulo no acabe siendo el cajón de todo:

- **El servidor de correo**: sus datos son variables de entorno, no configuración de la aplicación.
- **Las plantillas de los correos**: son del módulo `mail`, y su editor vive en la misma pantalla.
- **Los papeles y sus permisos**: los cuatro papeles son fijos (`docs/propósito-y-alcance.md`).
- **El tema de cada persona**: lo elige cada uno, y se recuerda en su navegador; lo que fija la
  instalación es **el color primario**, que sólo aparece en los dos temas personalizables.
- **Los estados de los tickets, los límites de los adjuntos y la política de contraseñas**: son
  reglas del producto, no configuración.

## 3. Dónde vive: cuatro tablas de una fila

Como ya se decidió para la numeración, **nada de «ajustes» con clave y valor**: una tabla así acaba
siendo el cajón donde entra todo y pierde los tipos y las restricciones. Cada tabla tiene **una sola
fila**, con columnas de verdad.

**`installation_settings`** — la configuración propia de la instalación:

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `integer` | clave primaria con `CHECK (id = 1)`: es una sola fila |
| `installation_name` | `text` | el nombre de la institución, la empresa o el equipo. **No admite vacío** y no pasa de 60 caracteres |
| `language` | `text` | `es` o `en` |
| `primary_color` | `text` | en hexadecimal, `#rrggbb` |
| `logo_light` | `text` | **el nombre del archivo** del logo para los temas claros, o nulo |
| `logo_dark` | `text` | el del tema oscuro, o nulo |
| `updated_at` | `timestamptz` | |
| `updated_by_id` | `bigint` | quién la cambió; nulo si fue la siembra |

- **En la tabla se guarda el nombre del archivo, no el archivo.** El logo es un archivo y vive en el
  disco (sección 5): meterlo en la base engordaría cada copia de seguridad y cada consulta.
- **Ninguna de las dos variantes del logo es obligatoria**: se puede subir sólo una, y la aplicación
  se arregla con esa (sección 5).

**`directory_settings`** — el directorio de la organización, para el camino de AD:

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `integer` | una sola fila, como las demás |
| `host`, `port` | `text` | el servidor y su puerto; **`host` vacío es «esta instalación no tiene directorio»** |
| `use_tls` | `boolean` | por ahí viajan credenciales: en producción, sí |
| `bind_dn`, `bind_password` | `text` | la cuenta de servicio con la que se busca a la gente |
| `search_base`, `user_filter` | `text` | dónde se busca, y el filtro con `%s` donde va el correo |
| `attr_email`, `attr_name`, `attr_last_name`, `attr_id` | `text` | de qué atributos salen el correo, el nombre, los apellidos y el identificador |
| `updated_at`, `updated_by_id` | | |

**`keycloak_settings`** — el reino, para el camino de Keycloak:

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `integer` | una sola fila |
| `issuer` | `text` | la dirección del reino, **la que ve el navegador**; vacío es «no hay este camino» |
| `client_id`, `client_secret` | `text` | el cliente confidencial del reino y su secreto |
| `redirect_uri` | `text` | a dónde vuelve Keycloak: la ruta de vuelta de `auth` |
| `updated_at`, `updated_by_id` | | |

- **Los valores de fábrica están puestos donde se puede**, para no obligar a escribirlo todo: el
  puerto `389`, el filtro `(mail=%s)`, los atributos `mail`, `givenName`, `sn` y `objectGUID` y el
  directorio vacío. Lo que no se puede adivinar —el servidor, la base de búsqueda, el reino— lo pone
  quien administra.
- **Un camino configurado a medias se rechaza al guardar** (`settings.directory.incomplete`,
  `settings.keycloak.incomplete`): un directorio sin base de búsqueda no busca a nadie, y un reino sin
  vuelta no puede volver.

**`ticket_settings`** — lo que configura el comportamiento de los tickets:

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `integer` | una sola fila, como la anterior |
| `number_prefix` | `text` | `^[A-Z0-9]{2,8}$` |
| `main_assignment` | `text` | `ninguna` o `por_turnos` |
| `main_notification` | `text` | `a_nadie`, `a_todo_el_equipo` o `al_asignado` |
| `internal_assignment` | `text` | igual que la de los principales |
| `internal_notification` | `text` | igual que la de los principales |
| `updated_at`, `updated_by_id` | | |

> **Hallazgo, enmendado en `docs/modules/tickets.md`** (aprobado el 2026-09-24). Ese documento describe
> el reparto y el aviso por tipo de ticket (sección 3.3) y dice que el prefijo vive aquí, pero
> **nombraba una tabla `ticket_number_settings` que sólo recogía el prefijo**: el reparto no cabía en
> ninguna parte. Esa tabla pasa a llamarse **`ticket_settings`** y recoge las cinco cosas, que es lo
> que se describe arriba. La enmienda está anotada en el registro de `tickets.md` y su tabla de la
> migración quedó corregida.

**Una restricción que no se puede olvidar**: con `main_assignment = ninguna`, el aviso `al_asignado`
no tiene sentido —no hay a quién avisar— y la base lo rechaza
(`CHECK (main_assignment = 'por_turnos' OR main_notification <> 'al_asignado')`), igual que lo hace la
pantalla, que no ofrece esa combinación.

**La migración** (`backend/migrations/v1.0.0.sql`) crea las cuatro tablas y **siembra su única fila**
con los valores de fábrica de la sección 2, para que la aplicación arranque sin pasar por la pantalla
de configuración. **El nombre de la instalación es una columna más de esa misma tabla**, y las
cuentas y los tickets con los que se trabaja en desarrollo van en `backend/migrations/v1.0.0_dev.sql`,
el archivo de datos de ejemplo que se aplica sólo en desarrollo con `scripts/dev-seed.sh`
(`docs/ambientes.md`, secciones 3.3 y 5).

## 4. Cómo la lee quien la necesita

**Nadie lee las tablas de otro módulo** (regla de modularidad, `docs/arquitectura.md`, sección 4). Lo
que hay es **una función de servicio**, que es lo que ya decidió `docs/modules/tickets.md`:

| Quién | Qué le pide | Para qué |
| --- | --- | --- |
| `tickets` | El prefijo y el reparto de ese tipo de ticket | Numerar un ticket y repartirlo al crearlo |
| `users` | El idioma de la instalación | El idioma con el que nace una cuenta que no lo elige |
| `auth` | **`Access()`: el método de entrada y las dos configuraciones, con sus secretos** | Entrar: decidir por dónde se entra y hablar con el directorio o con el reino |
| El frontend (`core`) | Toda la configuración | La pantalla de Configuración, el logo, el color primario y los campos de los dos caminos |

- **Sin caché**: es una fila que se lee por su clave primaria, y una caché que se quede desfasada
  haría que un cambio de prefijo no se notara. Cuando se guarda, se guarda, y la siguiente lectura lo
  ve. **Y con cómo se entra es obligatorio que sea así**: cambiar el método o el directorio tiene que
  valer **en el intento siguiente**, sin reiniciar nada.
- **El módulo `settings` no avisa a nadie**: los demás leen cuando lo necesitan. No hay eventos ni
  suscripciones que mantener.
- **Lo que `auth` necesita lo declara `auth`** —una interfaz con un solo método— y lo cumple este
  módulo; lo que este módulo necesita para probar una conexión lo declara este módulo y lo cumple
  `auth`, que es quien sabe hablar con un directorio y con un reino. Las dos direcciones se conectan en
  `main.go`, que es el único sitio donde los módulos se conocen (`docs/arquitectura.md`, sección 4).

## 5. La marca: el logo y el nombre

**El logo es de la instalación, no de una persona.** Se enseña en la pantalla de entrada, en el menú
lateral y en la pantalla de Configuración, y **un Administrador puede reemplazarlo** por el de su
institución.

### 5.1 Dos huecos, los dos opcionales

Un logo sirve para un fondo claro o para uno oscuro, casi nunca para los dos: uno con la tinta oscura
**no se lee sobre un fondo oscuro**, y al revés. Por eso hay **dos huecos** —y por eso **la marca de
fábrica también son dos archivos**, uno por tema:

| Hueco | Para qué sirve | ¿Obligatorio? |
| --- | --- | --- |
| **Claro** | El logo que se ve en los **temas claros** (Papel, Niebla, Alto contraste, Claro) | No |
| **Oscuro** | El que se ve en los **temas oscuros** (Grafito, Noche, Sepia, Oscuro) | No |

**Cuál se enseña**, en este orden:

1. El del hueco que toca al tema que se está viendo.
2. Si ese hueco está vacío, **el del otro hueco** (mejor un logo pensado para el otro fondo que
   ninguno).
3. Si no hay ninguno, **el logo de fábrica de ese tema**, que trae la aplicación.

Así, quien sube un solo logo lo ve en los ocho temas, y quien sube los dos ve el que corresponde. **Y
si no ha subido ninguno, ve el de fábrica del tema que tenga puesto**, que es lo que hace que la
pantalla de entrada se lea con el tema oscuro igual que con el claro.

**Los dos logos de fábrica son archivos de la aplicación**, no de la instalación: viven en
`frontend/public/` (`logo-catalina-support-light.png` y `logo-catalina-support-dark.png`) y **no se
configuran desde la pantalla**. Cambiarlos es reemplazar el archivo y volver a construir el frontend,
y eso lo hace quien mantiene la aplicación, no un administrador.

### 5.2 Qué se puede subir

| Qué | Regla |
| --- | --- |
| **Formatos** | `png`, `jpg`, `jpeg`, `webp` y `svg` |
| **Peso máximo** | **1 MB** |
| **Lado máximo** | **2000 píxeles** de ancho y de alto (el SVG no tiene píxeles: se mide su `viewBox`) |
| **Nombre** | Da igual el que traiga el archivo: **el nombre lo pone la aplicación** |

- **Se valida el contenido, no la extensión**: se miran los primeros bytes del archivo. Un `.png` que
  por dentro es otra cosa se rechaza.
- **El SVG se sirve aislado**: se descarga con `Content-Security-Policy: sandbox` y
  `X-Content-Type-Options: nosniff`, para que **no pueda ejecutar nada** si alguien abre su dirección
  directamente. Dentro de la aplicación se usa como imagen, donde un SVG no puede hacer nada.
- **No se recorta ni se estira nada**: el logo se enseña entero, con su proporción, dentro de un
  recuadro de esquinas redondeadas y con un borde del tema. Un logo con fondo opaco —como el de
  fábrica— se ve como una pieza, y uno con transparencia se integra; los dos se ven bien sin tocar el
  archivo. **Ni el fondo ni la forma se dan por supuestos**, porque el sistema tiene que servir para
  cualquier logo.

### 5.3 Dónde se guarda

- **En el disco**, en `_files/brand/`, que es la carpeta que ya existe para los archivos (los
  adjuntos de los tickets) y que **entra en la copia de seguridad** (`docs/ambientes.md`).
- **El nombre lo genera la aplicación** (`logo-claro-20260924T181500.png`): así no hay dos archivos
  con el mismo nombre, no se puede colar una ruta, y se sabe de cuándo es cada uno.
- **El nuevo se guarda antes de borrar el viejo**: si algo falla a mitad, se queda el logo anterior y
  no una instalación sin logo.
- Si el archivo de la tabla ya no está en el disco —alguien lo borró a mano—, la aplicación **se
  comporta como si ese hueco estuviera vacío** y enseña el que toque. No falla: la marca no es una
  función crítica.

### 5.4 Cómo se sirve

El logo se ve **en la pantalla de entrada, antes de que nadie haya entrado**, así que su endpoint es
**público**: no lleva sesión. Enseña el logo de la institución y nada más.

**Se sirve con versión**: la dirección lleva el sello de la última vez que se cambió
(`/api/settings/brand/logo?theme=claro&v=20260924T181500`). Sin eso, un navegador que ya tenía el logo
viejo en su caché seguiría enseñándolo después de reemplazarlo, y el administrador creería que el
cambio no funcionó. Con el sello, la respuesta se puede guardar para siempre; sin él, se pide cada
vez.

### 5.5 Reemplazarlo, y volver al de fábrica

- **Subir** un logo de una variante **reemplaza** el que hubiera en ese hueco, sin preguntar y sin
  borrar el otro hueco.
- **Quitar** el logo propio de una variante devuelve ese hueco al logo de fábrica.
- **Nada de esto toca el logo de fábrica**: el archivo que trae la aplicación no se sobrescribe nunca,
  así que siempre se puede volver a él.
- La pantalla enseña **cómo está cada hueco**: con su miniatura, el nombre del archivo, su peso y
  cuándo se cambió, o «el de fábrica» si está vacío.

### 5.6 Dónde se enseña, y dónde no

| Dónde | Cómo |
| --- | --- |
| **Pantalla de entrada** | Arriba del formulario, centrado, con una altura máxima de 128 px en móvil y 160 px de tablet para arriba |
| **Menú lateral** | Arriba, junto al nombre del producto, con una altura máxima de 32 px |
| **Configuración** | La marca: los dos huecos con su vista previa, subir y volver al de fábrica |

- **En los correos no va**: `docs/modules/mail.md` decidió que los correos son HTML **sin imágenes**,
  y un logo por correo obligaría a servirlo desde fuera o a adjuntarlo en cada envío.
- **El icono del navegador (favicon) es un archivo de la aplicación** (`frontend/public/favicon.ico`),
  **no una imagen configurable**: un favicon tiene que ser un `.ico` con varios tamaños dentro, y eso
  es cosa de quien mantiene la aplicación, no de la pantalla de Configuración. Se cambia reemplazando
  el archivo, como los dos logos de fábrica.

### 5.7 El nombre de la instalación

**Es cómo se llama esta instalación**, y sustituye al texto «Catalina Support» que traía la
aplicación: la institución, la empresa o el equipo. Se configura en la pantalla de Configuración,
**junto al logo**, que es como se pidió: personalizar la marca es el logo **y** el nombre.

| Dónde se lee | Cómo |
| --- | --- |
| **Pantalla de entrada** | Debajo del título, donde está el formulario: es lo primero que se lee al llegar |
| **Menú lateral** | Arriba, junto al logo |
| **Pestaña del navegador** | Es el título de la pestaña, en todas las pantallas |

- **Viaja en la marca pública** (`GET /api/settings/brand`), igual que el logo: las tres cosas se
  leen **antes de que nadie haya entrado**, así que no pueden depender de una sesión.
- **Se cuenta en caracteres, no en bytes** («Ayuntamiento de Ávila» son 20): lo que se mide es lo que
  se ve, y el tope de 60 está puesto donde se enseña, que es un menú estrecho.
- **Vacío devuelve el de fábrica**, y el propio campo lo dice: es la forma de deshacer el cambio sin
  quedarse sin nombre, que la base tampoco admite.
- **No cambia los correos** (decisión del responsable, 2026-09-25): los asuntos y los textos siguen
  como están, y **el remitente sigue saliendo de `SMTP_FROM_NAME`**, que es una variable del entorno.
  Personalizar los correos es otra cosa —el editor de `mail` está para eso— y no se mezcla con esto.
- **El título de la pestaña no lleva el de la ruta**: todas las rutas traían el mismo texto, así que
  no decían en qué pantalla estabas. Lo pone la instalación, en un solo sitio (`TitleStrategy`).

### 5.8 Método de autenticación: el método y las dos configuraciones

**La instalación entra por un método a la vez**, y eso se elige aquí: `local` —las cuentas viven en
esta base—, `ad` —la contraseña es la del directorio de la organización— o `keycloak` —se entra desde
la pantalla del reino—. Los otros dos **quedan apagados**, aunque estén configurados: no es que no
funcionen, es que la instalación no los ofrece.

| Método | Qué se ofrece en la pantalla de entrada | Quién puede entrar |
| --- | --- | --- |
| `local` | El formulario de correo y contraseña, y el enlace de recuperarla | Las cuentas de la base |
| `ad` | **El mismo formulario** —lo que cambia es quién contesta—, sin el enlace de recuperarla | Quien esté en el directorio |
| `keycloak` | El botón que lleva al reino, y **ningún formulario** | Quien tenga cuenta en el reino |

- **La cuenta de fábrica entra siempre**, sea cual sea el método: es la única puerta que no se puede
  cerrar, porque sin ella elegir mal el método dejaría la instalación sin nadie que lo cambiara. Con
  el método en Keycloak no hay formulario, así que **su puerta es un enlace discreto** en la pantalla
  de entrada («Entrar como administrador») que lo saca. Con los otros dos métodos, el mismo formulario
  vale para ella.
- **Con el método en `ad` o en `keycloak`, las cuentas locales —Soporte y Desarrollo incluidas— no
  pueden entrar.** Es la consecuencia de «uno a la vez» y está contada donde toca
  (`docs/usuarios-y-permisos.md`, sección 5): una instalación que pasa a AD se queda con el directorio
  como única puerta, y volver atrás es volver a elegir `local`.
- **No se puede elegir un método que no esté configurado**: el backend lo rechaza
  (`settings.method.notConfigured`, 422). Es lo que impide dejar la instalación con la cuenta de
  fábrica como única puerta por un descuido.
- **Los dos caminos se prueban antes de guardarlos**, con su botón: la prueba deja al módulo `auth`
  conectarse con la cuenta de servicio —sin validar la contraseña de nadie— o leer el documento del
  reino, y **no toca la base**. Es lo que evita guardar una configuración con la que nadie puede
  entrar y descubrirlo al día siguiente.
- **Los secretos no salen nunca por la API.** Lo que devuelve `GET /api/settings` es
  `directory.passwordSet` y `keycloak.secretSet`: dos booleanos que dicen si hay uno puesto. Al
  guardar, **vacío quiere decir «no lo cambies»**, no «bórralo»: la pantalla deja el campo en blanco y
  sólo escribe quien quiere cambiarlo.
- **El secreto está en la base, y eso tiene una consecuencia que hay que decir**: va dentro de la copia
  de seguridad de la base (`docs/ambientes.md`, sección 6), así que **una copia de la base es un
  secreto más**. Es el precio de poder cambiarlo desde la pantalla, y es lo que decidió el responsable
  el 2026-09-25.
- **Cambiar el método o el directorio vale en el intento siguiente**, sin reiniciar nada: el camino de
  entrada lee la base en cada intento (sección 4). Es lo que hace que esto sea configuración y no
  despliegue.
- **El botón de Configuración se llama «Probar la conexión» y no «Guardar y probar»**: probar no
  guarda nada, y guardar no prueba nada. Se puede probar con lo que hay en pantalla y luego decidir.

### 5.9 La versión del sistema

**La instalación dice qué versión es**, y se lee en dos sitios: **en la fila de salir del menú
lateral, alineada a la derecha**, y **en el pie de la pantalla de entrada**
(`docs/interfaz-y-experiencia.md`, secciones 3.1 y 6.4).

- **No es configurable, y por eso no está en la lista cerrada de la sección 2**: es un dato del
  software, no de la institución. **Se escribe en un solo sitio del backend** —`shared/version`, una
  constante— y se cambia cuando se cierra una versión, junto con la etiqueta de git.
- **Va en la marca pública** (`GET /api/settings/brand`), que es lo que la aplicación pide al
  arrancar: así el menú y la pantalla de entrada la tienen **sin una segunda llamada** y sin sesión,
  porque la entrada se pinta antes de que nadie entre. **La versión de un producto no es un secreto**,
  y no dice nada de nadie.
- **Se enseña como `v1.0.0`**: la `v` la pone la interfaz, y el valor que guarda y publica el backend
  es `1.0.0`.
- **Sin fecha de compilación**: cambiaría en cada compilación, y dos instalaciones con el mismo código
  dirían cosas distintas. Lo que hace falta saber —qué versión está puesta— lo dice el número.
- **La versión de la aplicación y la del esquema de la base son la misma cosa**: `1.0.0` es la versión
  de este repositorio, que es la que nombra la migración `v1.0.0.sql`. Cuando haya una `v1.1.0.sql`,
  habrá una versión `1.1.0` que decir.

### 5.10 La región horaria y la dirección pública

**Son de la instalación entera, no de la numeración**, así que viven en **su propia tarjeta** de la
pantalla, con **su propio botón** (decisiones 14 y 15). El `PUT` manda **la configuración entera** —es
un documento, no un parche—, así que lo único que cambia es qué botón lo pide y con qué mensaje se
confirma.

**La región** se elige de **una lista con buscador**, no se escribe:

- La lista es **el catálogo de zonas del navegador** (`Intl.supportedValuesOf('timeZone')`), **más los
  desfases fijos** `Etc/GMT±1…12`, que ese catálogo deja fuera y son la forma de decir «UTC-5» sin
  horario de verano. Si el navegador no trae la función —o contesta vacío—, cae a **una lista corta
  escrita a mano** con las zonas habituales: la pantalla nunca se queda sin nada que elegir.
- **Debajo se explica la elegida**: la hora que es en ella **ahora mismo** y su desfase respecto a
  UTC. Se calculan con `Intl.DateTimeFormat` (`timeZoneName: 'shortOffset'`) dentro de un `try`, así
  que una zona que no valga **no rompe la pantalla**: no se enseña nada.
- Se guarda **el nombre de la zona**, y el servidor lo comprueba con la base de zonas que va
  **incrustada en el binario** (`time/tzdata`): una zona que no exista se rechaza con
  `settings.timeZone.unknown`.
- **La lee quien tiene que escribir una fecha**: la interfaz, por la marca pública (`timeZone`), y los
  correos, por lo que el módulo de correo le pide a este. **Las fechas guardadas siguen en UTC**, así
  que cambiarla no mueve ningún ticket; y los correos que llevan fecha —el aviso de contraseña
  cambiada— y la vista previa del editor se escriben en ella.

**La dirección pública** se escribe con su esquema, su host y **su puerto si hace falta**, y
**`localhost` vale**: es lo que permite probar la instalación en local y usarla en serio sin tocar el
servidor. El servidor la valida —`http` o `https`, con host— y rechaza lo demás con
`settings.publicUrl.invalid`; la base tiene su propio tope para lo mismo. Es **la base de los enlaces
que salen en los correos** y de **la vuelta de Keycloak**, y **si no hay ninguna** —ni en la
configuración ni en `PUBLIC_APP_URL`, que queda de respaldo— el enlace **falla con su clave**
(`auth.publicUrl.missing`) en vez de mandar un enlace roto.

**Y si no es https, la pantalla lo avisa** (`configuracion.direccionInsegura`): la contraseña y la
sesión viajan sin cifrar. **No bloquea nada** —lo eligió así el responsable—, porque la instalación
tiene que poder probarse en local; el aviso sale en Configuración y no en la pantalla de entrada.

### 5.11 El orden de la pantalla

**Cada tarjeta lleva lo que su título dice** (decisión del responsable, 2026-09-30). La pantalla tenía
dos incoherencias que se veían al usarla: **«La numeración y el reparto» sólo tenía dentro el idioma
de la instalación**, y **«Región horaria y dirección pública» tenía dentro el prefijo y todo el
reparto**. Se ordenó así, de lo que la instalación **es** a cómo **entra**, **trabaja** y **avisa**:

| # | Tarjeta | Qué lleva |
| --- | --- | --- |
| 1 | **La instalación** | El nombre y **el idioma** —las dos cosas son *qué es* esta instalación, y el idioma es el de las cuentas nuevas y el de los correos—. Un botón: **Guardar la instalación** |
| 2 | **La marca** | **El logo y el color institucional**, que son la identidad de la institución y estaban en dos tarjetas sin motivo. El logo se guarda al subirlo; el botón **Guardar la marca** es para el color |
| 3 | **Método de autenticación** | El método y sus dos configuraciones, con sus pruebas de conexión (sección 5.8) |
| 4 | **La numeración y el reparto** | El prefijo y el reparto de los dos tipos de ticket (sección 5.8) |
| 5 | **Región horaria y dirección pública** | La zona horaria y la dirección (sección 5.10) |
| 6 | **Los correos** | El enlace al editor, que es del módulo `mail` |

**El tope está en una prueba de interfaz**: mira los títulos en orden y comprueba que el nombre y el
idioma están juntos, que el logo y el color están juntos, que el prefijo y el reparto están en la
numeración y que la región no lleva la numeración dentro. Es lo que impide que se vuelva a
descolocar al añadir un campo.

### 5.12 El sello de instalación, y el asistente de primer arranque

**La instalación se configura antes de tener puerta, y se hace una sola vez**
(`docs/primer-arranque.md`). Lo que lo sostiene es **un sello**: `installation_settings.installed_at`,
**nulo** mientras nadie haya terminado el asistente.

- **Nulo es «sin instalar»**: la aplicación lleva a `/setup` y la API acepta configurar.
- **Puesto es «instalada»**: `/setup` lleva a la entrada y **toda** la API del asistente contesta
  **409** `setup.alreadyInstalled`. Eso es lo que impide reescribir la configuración de una instalación
  en marcha, y es la única escritura de la aplicación **sin una cuenta detrás**: lo que la protege no es
  un permiso, es el sello.
- **El sello se pone al terminar** y se guarda con `installed_at = now()`.
- **Y no aparece en producción por actualizar**: el relleno de la migración sella las filas que ya
  existían, y sólo lo hace la primera vez que la columna nace (ver el bloque del relleno en
  `v1.0.0.sql`).

**Los cuatro pasos** guardan de uno en uno —la instalación, cómo se entra, dónde está y el correo—, y
cada uno valida **sólo lo suyo**, porque el resto todavía no está puesto. El estado que devuelven
—`GET /api/setup` y las respuestas de cada paso— es lo que permite **seguir donde se dejó**, y **no
lleva ningún secreto**: de las contraseñas dice sólo si hay una puesta.

**El correo saliente vive en esta tabla** (sección 5 del documento del asistente) y el módulo `mail` lo
lee en cada envío a través de la interfaz que él declara. **No hay respaldo en el entorno**: las
variables `SMTP_*` se retiraron, como se retiraron las del directorio y Keycloak (corrección del
responsable, 2026-09-30).

## 6. Los endpoints

| Método y ruta | Qué hace | Quién |
| --- | --- | --- |
| `GET /api/settings` | La configuración de la instalación, con el estado de la marca | Administrador |
| `PUT /api/settings` | La cambia. El prefijo **no toca los números ya emitidos** | Administrador |
| `POST /api/settings/brand/logo?variant=claro\|oscuro` | Sube o reemplaza el logo de ese hueco (`multipart`) | Administrador |
| `DELETE /api/settings/brand/logo?variant=claro\|oscuro` | Quita el propio y vuelve al de fábrica | Administrador |
| `GET /api/settings/brand/logo?theme=claro\|oscuro` | **El archivo que toca** para ese tema. **Público** | Cualquiera |
| `GET /api/settings/brand?color=%23rrggbb` | **Vista previa**: los mismos colores resueltos, pero **sin guardar nada**. Es lo que permite enseñar cómo va a quedar antes de cambiar el color de la casa | Cualquiera |
| `GET /api/settings/brand` | **Público.** El **nombre de la instalación**, **la versión del sistema**, el color ya resuelto y si hay logo propio: es lo que la aplicación lee para pintarse antes de que nadie entre | Cualquiera |
| `POST /api/settings/directory/test` | **Prueba** la configuración del directorio que llega en el cuerpo, **sin guardarla**: conecta con la cuenta de servicio y la cierra. No valida la contraseña de nadie | Administrador |
| `POST /api/settings/keycloak/test` | **Prueba** la configuración de Keycloak que llega: lee el documento del reino y comprueba que dice dónde está su pantalla de entrada | Administrador |

- **`GET` y `PUT /api/settings` ya estaban decididos** en `docs/modules/tickets.md`, sección 5, y son
  de administrador. Aquí se les añade el estado de la marca en la respuesta.
- **El logo no viaja dentro de `PUT /api/settings`**: es un archivo, va por `multipart` en su propia
  ruta, como los adjuntos de los tickets. **El nombre sí**: es texto, y va con el resto de la
  configuración.
- **`GET .../logo` sin nada subido responde 404**, y el frontend enseña el logo de fábrica. No es un
  error: es lo normal en una instalación nueva.
- **`PUT /api/settings` lleva también el método y las dos configuraciones**, y `GET` las devuelve
  **sin los secretos**. Los dos endpoints de prueba van aparte y **por `POST` con la configuración en
  el cuerpo**, porque prueban **lo que hay en pantalla** y no lo que hay guardado: si leyeran la base,
  no servirían para lo único que sirven.
- **Las dos pruebas contestan `{"status":"ok"}`** o su clave de error, que es lo que la pantalla
  necesita para decir qué ha pasado sin que el backend cuente sus detalles: el motivo —no contesta, el
  bind no vale, el reino no existe— queda en el log, que es donde lo lee quien administra.

## 7. Las claves de error

| Clave | Cuándo | Código |
| --- | --- | --- |
| `settings.name.tooLong` | El nombre de la instalación pasa de 60 caracteres | **422** |
| `settings.language.unknown` | El idioma no es `es` ni `en` | **422** |
| `settings.primaryColor.invalid` | El color no es un hexadecimal de seis dígitos | **422** |
| `settings.numberPrefix.invalid` | El prefijo no cumple `^[A-Z0-9]{2,8}$` | **422** |
| `settings.assignment.unknown` | El valor de asignación no es `ninguna` ni `por_turnos` | **422** |
| `settings.notification.unknown` | El valor de aviso no es ninguno de los tres | **422** |
| `settings.notification.withoutAssignment` | Se pide aviso `al_asignado` con asignación `ninguna` | **422** |
| `settings.logo.format` | El archivo no es `png`, `jpg`, `jpeg`, `webp` ni `svg` | **422** |
| `settings.logo.invalid` | El contenido no es una imagen válida, o pasa del lado máximo | **422** |
| `settings.logo.tooBig` | Pesa más de 1 MB | **413** |
| `settings.logo.variantUnknown` | La variante pedida no es `claro` ni `oscuro` | **404** |
| `settings.method.unknown` | El método de entrada no es `local`, `ad` ni `keycloak` | **422** |
| `settings.method.notConfigured` | Se elige un método que todavía no está configurado | **422** |
| `settings.directory.incomplete` | Al directorio le falta el servidor o la base de búsqueda | **422** |
| `settings.keycloak.incomplete` | Al reino le falta el emisor, el cliente o la vuelta | **422** |
| `settings.directory.unreachable` | La prueba de la conexión con el directorio ha fallado | **422** |
| `settings.keycloak.unreachable` | La prueba con el reino ha fallado | **422** |

## 8. Lo que se decidió al repasar este documento

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **El límite del logo** | **1 MB y 2000 píxeles de lado**. Un logo de más de 1 MB no aporta nada a 160 px de alto, y uno de más de 2000 px sólo engorda |
| 2 | **Qué hace el idioma de la instalación** | Es el de **las cuentas nuevas** y el de la instalación; **no manda sobre la interfaz** de nadie |
| 3 | **El color institucional** | El de fábrica es el azul que ya está puesto, y **sólo afecta a los dos temas de fábrica**: los seis fijos llevan su acento |
| 4 | **Dos huecos de logo, los dos opcionales** | Sí, y si sólo hay uno se usa en los ocho temas. **La marca de fábrica son dos archivos, uno por tema** (decisión del responsable, 2026-09-26): los que trae la aplicación, reemplazables pero **no configurables desde la pantalla** |
| 5 | **`ticket_number_settings` pasa a `ticket_settings`** | La enmienda a `docs/modules/tickets.md`: el reparto no cabía en ninguna tabla |
| 6 | **El logo no va en los correos** | Ya estaba decidido en `docs/modules/mail.md` (HTML sin imágenes) |
| 7 | **El endpoint del logo es público** | Tiene que serlo: la entrada lo enseña antes de entrar. Sólo devuelve la marca de la institución |
| 8 | **Qué entra ahora en la pantalla de Configuración** | **La marca y el color institucional**, que son lo que ya tiene quién lo use. El idioma llega con el alta de cuentas, y el prefijo y el reparto con los tickets |
| 9 | **El nombre de la instalación** | Se configura, y sustituye a «Catalina Support» en la aplicación. **Decisión del responsable, 2026-09-25**, con dos correcciones suyas al proponerlo: **los correos no se tocan** —ni sus asuntos ni su texto— y **el remitente sigue saliendo de `SMTP_FROM_NAME`**, que es del entorno. El tope de 60 caracteres y que vacío devuelva el de fábrica son propuestas mías, aplicadas y contadas aquí (sección 5.7) |
| 10-bis | **Cada método enseña sólo su configuración** | **Decisión del responsable, 2026-09-27**: en «Método de autenticación», el panel del directorio sale **sólo** con el método de la organización y el de Keycloak **sólo** con Keycloak; con «Cuentas de la aplicación» no se enseña ninguno de los dos. **No cambia lo que se guarda** —el `PUT` sigue mandando las dos configuraciones y los secretos guardados no se tocan— ni el aviso de «método sin configurar», que sigue a la vista del método elegido: cambia **lo que se enseña**, que era enseñar trabajo que no tocaba |
| 10 | **El directorio y Keycloak se configuran desde la pantalla** | **Decisión del responsable, 2026-09-25**: los dos caminos salen del entorno y entran en la lista cerrada de lo configurable, con **dos tablas de una fila** y su sección (5.8). **Los secretos se guardan en la base y no se devuelven nunca por la API** —sólo `passwordSet` y `secretSet`—, con un botón de «Probar la conexión» que no guarda nada. Consecuencia aceptada y **confirmada por el responsable el 2026-09-25**: **una copia de la base es un secreto más**, con los permisos de la máquina y sin salir de ella (`docs/ambientes.md`, sección 6) |
| 15 | **La región de la instalación decide cómo se leen las fechas** | **Decisión del responsable, 2026-09-29**: la instalación tiene **una zona horaria** —se elige de una **lista de zonas (IANA) con buscador**, y la pantalla enseña la hora que es en ella y su desfase— y **decide cómo se leen todas las fechas: las de la interfaz y las que van dentro de los correos**. **Las fechas guardadas no se tocan**: siguen en UTC, así que **cambiar la zona no mueve ningún ticket**, sólo cambia cómo se lee. Y no es un adorno: hasta hoy las fechas se pintaban con la zona **del navegador** de quien mira, así que la misma actuación se leía a horas distintas según dónde estuviera. Se guarda **el nombre de la zona**, no un desfase: así el horario de verano de quien lo tenga se respeta solo. La marca pública la lleva (`timeZone`), porque la entrada se pinta antes de entrar |
| 14 | **La dirección pública: los enlaces de los correos y la vuelta de Keycloak** | **Decisión del responsable, 2026-09-29**: la instalación tiene **una dirección pública** —esquema (`http` o `https`), host y puerto, **también `localhost`**— y es **la base de los enlaces que salen en los correos** (alta, restablecer la contraseña, los avisos) **y la dirección de vuelta de Keycloak**. La aplicación se sigue navegando **en relativo**, así que funciona por cualquier dirección con la que se llegue a ella: es lo que permite probarla en local y usarla en serio sin tocar el servidor. Hasta hoy eso era **una variable de entorno** (`PUBLIC_APP_URL`), obligatoria en producción; pasa a vivir en la base, como el directorio y Keycloak, y la variable queda **como respaldo** para una instalación que ya la tenga puesta |
| 13 | **El aviso de que no hay https, en Configuración** | **Decisión del responsable, 2026-09-29**: si la dirección pública **no es https**, la pantalla de Configuración lo **avisa** —la sesión y la contraseña viajan sin cifrar— **y no bloquea nada**: la instalación funciona igual. El responsable eligió que el aviso salga **sólo en Configuración** (y no también en la pantalla de entrada, que era la propuesta): quien decide cómo se expone es quien la configura |
| 12 | **La versión del sistema** | **Decisión del responsable, 2026-09-25**, con una **elección suya que no era mi recomendación**: es **una constante del backend** que viaja **en la marca pública**, y se lee **en el menú lateral —en la fila de salir, a la derecha— y también en el pie de la pantalla de entrada**. Yo proponía sólo el menú; él decidió que también se vea antes de entrar, que es donde sirve para decir qué versión se está mirando. **Sólo el número, con su `v`, sin fecha de compilación** |
| 11 | **Probar una conexión que falla responde 422** | **Confirmado por el responsable el 2026-09-25**, tal como se implementó: lo que se está probando son **los datos que hay en pantalla**, así que un fallo es «esa configuración no sirve» y no «el servicio está caído». El motivo —no contesta, el bind no vale, el reino no existe— queda en el log |
| 11 | **Un método de entrada a la vez** | **Decisión del responsable, 2026-09-25**, al preguntarle si los tres caminos convivían: **uno a la vez**, cambiable después, y **la cuenta de fábrica siempre puede entrar** —su puerta discreta en la pantalla de entrada es la que aprobó—. Consecuencia dicha y aceptada: con el método en `ad` o en `keycloak` las cuentas locales no entran, y eso enmienda la sección 5 de `docs/usuarios-y-permisos.md` |

## 9. El color institucional, resuelto

El administrador elige **un** color, pero la aplicación necesita cuatro valores, porque el mismo color
no se lee igual sobre un fondo claro que sobre uno oscuro. Los calcula el backend y los manda ya
resueltos:

| Valor | Qué es |
| --- | --- |
| `colors.light` | El color para los temas claros: el elegido, **oscurecido** si no llega a 4,5:1 sobre blanco |
| `colors.dark` | El de los temas oscuros: el elegido, **aclarado** si no llega a 4,5:1 sobre el fondo oscuro |
| `colors.onLight`, `colors.onDark` | El color del texto del botón: blanco o casi negro, **el que más contraste dé** |

- **Se mueve sólo la claridad**, no el tono: el color de la institución se sigue pareciendo a lo que
  eligió quien lo eligió.
- Los cuatro valores salen en el endpoint público de la marca, y la aplicación los pone en variables
  CSS. **Los seis temas fijos no los usan**: llevan su propio acento.
- Está comprobado con colores difíciles —blanco, negro, amarillo, rojo— en las pruebas del módulo:
  con cualquiera de ellos, los cuatro valores se leen.
- **La pantalla pide la vista previa al backend**, con el color elegido y sin guardarlo: resolverlo en
  el frontend sería tener la misma regla en dos sitios, y un día dirían cosas distintas. La vista
  previa enseña **lo que se va a ver de verdad**, incluido el ajuste si el color no se lee.

## 10. Qué habilita este documento

Con `docs/modules/settings.md` aprobado se puede hacer:

1. **La marca completa**: la tabla, la subida del logo con su validación, el endpoint público con
   versión, el aislamiento del SVG, y volver al de fábrica.
2. **El logo en la pantalla de entrada y en el menú lateral** —el menú llega con el armazón de la
   aplicación—, y en la pantalla de Configuración, con sus dos huecos.
3. **El prefijo y el reparto leídos desde donde viven**, que es lo que `tickets` necesita para
   numerar y repartir cuando llegue su módulo.
4. **Método de autenticación, configurable desde la pantalla**: el método, el directorio de la organización y
   Keycloak, con sus dos pruebas de conexión y sus secretos guardados sin salir por la API. Es lo que
   permite entregar la instalación sin pedir que nadie edite un archivo en el servidor.
