# Documentación de Catalina-Support

> **Estado:** as-built
> **Última actualización:** 2026-10-01
>
> **Actualizado el 2026-10-01 (séptima vez)**: **los datos de ejemplo se siembran con un contenedor de
> un solo uso**, `docker compose -f dev.yml run --rm seed`, **igual en Linux, macOS y Windows** —deja
> fuera el problema del guion de bash, que en Windows no se ejecuta, y con él la copia de los
> adjuntos—. El servicio vive en `dev.yml` **con perfil propio**, así que **`up -d` no lo arranca**, y
> **los ejemplos son opcionales** (el esquema sigue siendo obligatorio). El contenedor corre **el guion
> de siempre, `scripts/dev-seed.sh`**, que también sigue valiendo en Linux y macOS fuera del
> contenedor. Quedan enmendados `ambientes.md` (**18 enmiendas**: secciones 3.1, 3.3, 5 y 9.3) y
> `arquitectura.md` (**6 enmiendas**: sección 10.3), y el `README.md` quita el aviso de Windows y el de
> los adjuntos. Comprobado el 2026-10-01: **dos pasadas seguidas** del comando («Listo: 11 cuentas, 25
> tickets y 5 adjuntos» las dos veces), 11 cuentas y 25 tickets en la base, **5 adjuntos en `_files/`**
> y `up -d` sin arrancar el servicio.
>
> **Actualizado el 2026-10-01 (sexta vez)**: **el servidor de desarrollo de Angular reenvía `/api` al
> backend**, así que **se entra en `http://127.0.0.1:11001` con `admin`/`admin` sin nginx** —el
> recorrido de quien se descarga el proyecto y sólo tiene Docker—. Se añade `frontend/proxy.conf.json`
> y se declara en `serve.options.proxyConfig` de `frontend/angular.json`, con destino
> **`http://backend:11002`** (el servicio dentro de la red del entorno) y **sin reescribir la ruta**.
> Es **sólo de desarrollo**: el `build` no la ve y **en producción sigue reenviando nginx**. El
> `README.md` quita el aviso de «pendiente de decisión». Enmendados: `arquitectura.md` (**5
> enmiendas**: la sección 9, la 10, la 11 y la fila de la 13, y nginx deja de ser necesario para
> entrar) y `ambientes.md` (**17 enmiendas**: sección 3.1 y 9.3). Comprobado el 2026-10-01 por la
> 11001: `GET /api/auth/methods` y `GET /api/health` → **200 JSON**, `POST /api/auth/login` con
> `admin`/`admin` → **200** con token, y Playwright contra `http://127.0.0.1:11001` → **14 casos en
> verde** (entrar y salir con la cuenta de fábrica).
>
> **Actualizado el 2026-10-01 (quinta vez)**: **el esquema y los ejemplos se pueden aplicar sin `bash`**.
> Los pasos de local del `README.md` quedan como un recorrido para quien llega de fuera —clonar,
> levantar, aplicar el esquema, entrar con `admin`/`admin`—, con el porqué de cada paso, y **los dos
> comandos de Docker que valen igual en PowerShell, CMD y bash**. `./scripts/dev-seed.sh` queda dicho
> como **lo que es, de Linux y macOS** (y como lo que es: **un extra de contenido, no un requisito para
> entrar**). `ambientes.md` pasa a **16 enmiendas** (su sección 3.2 estrena los dos comandos y las
> secciones 3.1, 3.3 y 5 separan las dos vías) y `arquitectura.md` a **4** (su sección 10.3, que
> empezaba por `up -d` sin el esquema, ya lo trae). Comprobado el 2026-10-01 sobre una base nueva
> (`catalina_support_prueba`, borrada después): el esquema deja 18 tablas, `GET /api/auth/methods` y
> `POST /api/auth/login` con `admin`/`admin` responden **200 sin datos de ejemplo**, y el mismo archivo
> aplicado dos veces seguidas da **0 errores**.
>
> **Actualizado el 2026-10-01 (cuarta vez)**: el `README.md` deja claro, en los pasos de local, que
> **en una base nueva `./scripts/dev-seed.sh` es lo que crea el esquema** —no sólo los ejemplos—, que
> **hasta que no se pasa la aplicación no tiene tablas** y que los síntomas son
> `relación "installation_settings" does not exist` y `relación "ai_insights" does not exist`; y
> **reordena los pasos** (entorno, después esquema y datos, y sólo entonces abrir). `ambientes.md`
> pasa a **15 enmiendas**: sus secciones 3.1, 3.3 y 5 dicen qué aplica el esquema en desarrollo y
> **cuándo** —antes de abrir la aplicación— y apuntan al `README.md` en vez de repetirlo.
>
> **Actualizado el 2026-10-01 (tercera vez)**: **el directorio de pruebas y Keycloak dejan `dev.yml`**
> y pasan a **`active-directory.yml`** y **`keycloak.yml`**, cada uno con **su propio comando**, y **el
> perfil `auth` desaparece**. La suite de los caminos de directorio pasa a **dos pasos** —levantar cada
> servicio con su archivo y correr la suite normal—, y los dos entran en la red
> **`catalina-support-dev`**, que **posee `dev.yml`** con un nombre fijo (sin `external`). Quedan
> enmendados `arquitectura.md` (**3 enmiendas**), `ambientes.md` (**14**) y `modules/auth.md` (**7**).
> Y el `README.md` dice ya **la contraseña de fábrica de desarrollo** —la de `config/env/dev.env`, que
> se versiona—, dejando claro que **la de producción es `ADMIN_PASSWORD` y no se escribe**.
>
> **Actualizado el 2026-10-01 (segunda vez)**: el `README.md` estrena **la tabla de las once cuentas
> de ejemplo de desarrollo** —nombre, correo, rol y contraseña—, y `ambientes.md` pasa a **13
> enmiendas** al **dejar de repetir esa lista** y apuntar a la tabla del `README.md` (sección 3.3),
> para que el dato viva en un solo sitio.
>
> **Actualizado el 2026-10-01**: los **tres elementos opcionales** —el motor de IA, el directorio de
> pruebas y Keycloak— quedan con **un comando de Docker cada uno** y dicho que son opcionales:
> `ambientes.md` pasa a **12 enmiendas** (sección 9.3: el directorio y Keycloak comparten el perfil
> `auth` pero **pueden levantarse por separado** nombrando el servicio, sin activar el perfil
> —comprobado—, y el comando de siempre `--profile auth up -d` sigue levantando los dos),
> `modules/ai.md` a **2 enmiendas** (se corrigen los recursos del motor: **1500m de memoria**, **sin
> tope de CPU** y **~1,44 GiB medidos**, donde decía «950 MB / 1,09 GB» y «2 CPU»; y se explica el
> **aviso esperado del volumen `ai_modelos`**) y `arquitectura.md` a **2 enmiendas** (la fila del motor
> medido, en la sección 13, con las mismas cifras: **1500 MiB de tope**, **~1,44 GiB —el 98% del
> tope—** y **sin tope de CPU**).
>
> **Actualizado el 2026-09-30 (tercera vez)**: los dos documentos de módulo afectados por las pruebas
> del asistente quedan al día: `modules/mail.md` pasa a **6 enmiendas** —estrena **`Probar`**, que
> conecta y autentica sin mandar ningún correo— y `modules/settings.md` a **8 enmiendas** —declara
> **`ProberDeCorreo`**, que el asistente reutiliza para el paso 4—.
>
> **Actualizado el 2026-09-30 (segunda vez)**: **el asistente de primer arranque prueba lo que pide**
> —sus dos endpoints públicos `POST /api/setup/entry/test` y `POST /api/setup/mail/test`, con el mismo
> candado del sello, y los botones de los pasos 2 y 4—, y la prueba del correo **conecta y autentica
> sin mandar ningún correo**. `primer-arranque.md` pasa a **1 enmienda**
> (`docs/primer-arranque.md`, sección 3.1).
>
> **Actualizado el 2026-09-30**: **el motor de IA se configura desde la pantalla de Configuración**
> —su dirección y su modelo, con su botón de «Probar la conexión»—, con el entorno como respaldo
> (`docs/modules/settings.md`, decisión 16, y `docs/modules/ai.md`, decisión 21). `modules/ai.md`
> pasa a **as-built**.
>
> **Actualizado el 2026-09-27**: nace `modules/ai.md`, el sexto módulo, con los dos campos que
> redacta el motor de IA en la lista de tickets.
>
> **Actualizado el 2026-09-26 (tercera vez)**: los adjuntos de texto y código —un `.sql` no se podía
> mandar— y **la imagen y el vídeo con tope de 480 × 360** que se abren en el visor al pulsarlos
> (`docs/modules/tickets.md`, decisiones 54 y 55).
>
> **Actualizado el 2026-09-26 (segunda vez)**: las listas de tickets pasan a ser **Mis tickets** (lo
> mío), **Tickets principales** y **Tickets internos**, y se quitan del menú «Nuevo ticket» y de la
> bandeja «Asignármelo» (decisiones del responsable; `docs/modules/tickets.md`, decisiones 50 a 53).
>
> **Actualizado el 2026-09-26**: se fija **cómo se cuenta una enmienda** (sección «El registro de
> enmiendas y cómo se cuenta») y se ponen los números de la tabla con esa regla, porque hasta hoy
> unos cuadraban y otros no; y `modules/mail.md` y `modules/users.md` pasan a **as-built**, que era
> lo que les tocaba desde que su código existe.

Índice de la documentación del proyecto. La **Regla 0** de `AGENTS.md` es obligatoria: no se
escribe código sin un documento aprobado antes.

## Documentos

| Documento | Estado | Cubre |
| --- | --- | --- |
| `README.md` (este archivo) | as-built | Índice, convenciones y estados |
| `arquitectura.md` | **as-built** (6 enmiendas) | Stack, versiones, forma del repositorio, contenedores y regla de modularidad (frontend ↔ backend). Su sección 13 dice qué está construido, qué está sin verificar y qué falta. Enmendada el 2026-09-30 al poner al día esa sección (la pantalla de Configuración entera —con el motor de IA, la región y la dirección y el correo—, el primer arranque, el motor aparte y los recuentos); el despliegue a producción queda **aparcado**. Y el 2026-10-01: la fila del **motor de IA medido** se corrige con lo medido —**1500 MiB de tope**, **~1,44 GiB en marcha (el 98% del tope)** y **sin tope de CPU**—, y se quita el «1,09 GB con una entrada de 2 034 piezas», que no se pudo comprobar. Enmendada otra vez el 2026-10-01: **el directorio de pruebas y Keycloak salen de `dev.yml`** a `active-directory.yml` y `keycloak.yml`, con la red `catalina-support-dev` que posee `dev.yml`, y sus tablas y comandos se ponen al día. Y otra vez el 2026-10-01: **la sección 10.3 estrena el paso del esquema**, porque su lista empezaba por `docker compose -f dev.yml up -d` y quien la siguiera caía en `relación "installation_settings" does not exist`; quedan sus dos comandos de Docker (PowerShell, CMD y bash), los de los ejemplos y el aviso de que el guion es de Linux y macOS. Y otra vez el 2026-10-01: **el servidor de desarrollo reenvía `/api` al backend** con `frontend/proxy.conf.json` declarado en `serve.options.proxyConfig` (destino `http://backend:11002`, sólo de desarrollo: el `build` no cambia y en producción reenvía nginx), y nginx pasa a **no ser necesario para entrar** (secciones 9, 10, 11 y 13). Y otra vez el 2026-10-01: **los datos de ejemplo se siembran con `docker compose -f dev.yml run --rm seed`** —contenedor de un solo uso, con perfil propio para que `up -d` no lo arranque, corriendo el guion `scripts/dev-seed.sh`— y su sección 10.3 cambia su bloque de ejemplos por ese comando |
| `propósito-y-alcance.md` | **as-built** (3 enmiendas) | Qué problema resuelve la mesa de ayuda, los dos equipos y los cuatro papeles, el modelo de tickets (principal e interno, numeración y estados) y qué queda fuera. Sin decisiones abiertas. Las tres enmiendas son del 2026-09-22, antes de aprobarlo: el detalle del acceso, el repaso (el sexto aviso, el límite de 25 MB, las marcas de editado y eliminado) y la **regla 5** que cambió al escribir `flujos.md` |
| `usuarios-y-permisos.md` | **as-built** (10 enmiendas) | La matriz papel × acción, los tres caminos de entrada (correo, AD, Keycloak), las reglas de convivencia, la sesión, el ciclo de vida de las cuentas y la cuenta de administrador de fábrica. Enmendado el 2026-09-25 con **la instalación entrando por un método a la vez** (sección 5) y con que **la cuenta de fábrica entra siempre** (sección 8) |
| `flujos.md` | **as-built** (1 enmienda) | Los seis recorridos paso a paso (alta, triaje, escalado, trabajo de Desarrollo, cierre y reapertura), con los ocho avisos por correo |
| `modules/settings.md` | **as-built** (8 enmiendas) | El módulo `settings`: qué es configurable —**el nombre de la instalación**, **el método de entrada, el directorio y Keycloak**, idioma, color institucional, **la marca**, prefijo de numeración, reparto y **el motor de IA**—, dónde vive (cuatro tablas de una fila), cómo lo leen los demás módulos y **el logo**: dos huecos opcionales, formatos, validación, endpoint público y vuelta al de fábrica. Enmendado el 2026-09-25 con **el nombre de la instalación**, con **cómo se entra** (sección 5.8) —los dos caminos de directorio salen del entorno y se configuran desde la pantalla, **con sus secretos guardados y sin salir por la API**— con **la versión del sistema** (sección 5.9), que viaja en la marca pública, y con **los dos logos de fábrica, uno por tema**, más el favicon propio (2026-09-26). Enmendado el 2026-09-30 con **el motor de IA** (sección 5.13) y con **`ProberDeCorreo`**, la interfaz con la que el módulo prueba el correo saliente y que **reutiliza el asistente** para el paso 4 |
| `modules/mail.md` | **as-built** (6 enmiendas) | El módulo `mail`: la tabla de plantillas, los marcadores, **los veintidós textos por defecto** (once correos en dos idiomas), el envío en HTML con versión de texto, **las claves de error con su código y lo que devuelve cada endpoint** (sección 8) y el editor. **Está entero**: backend y el editor del frontend. Enmendado el 2026-09-30: estrena **`Probar`**, que **conecta y autentica sin mandar ningún correo** (`MAIL FROM`, `RCPT TO` y `DATA` no se usan) |
| `modules/auth.md` | **as-built** (7 enmiendas) | El módulo `auth`: la tabla de tokens de enlace, el token de sesión, las contraseñas, los tres caminos de entrada, los correos de cuenta, los endpoints, las seis pantallas del armazón y los contenedores de pruebas. Enmendado tres veces el 2026-09-23 al implementarlo (las claves del token de sesión y del 403, que no tenían ninguna; el campo de entrada, la respuesta, la ruta del enlace y dónde viaja su token; y el paso 1 de la sección 13), dos veces el 2026-09-25 al terminar **el camino de AD** —con la **corrección del responsable**: quien está en el directorio entra directamente, sin alta manual— y y **el de Keycloak**: las cuatro claves de error nuevas, cómo llega el navegador a Keycloak en desarrollo (por el mismo dominio, para que el emisor sea uno solo) y cómo vuelve el token en el fragmento. **Está entero**: los tres caminos de entrada, hechos y verificados. La sexta enmienda, el 2026-09-25, es **un método de entrada a la vez** (sección 5.0): el directorio y el reino se leen de la base en cada intento, los otros dos caminos quedan apagados, y **la cuenta de fábrica entra siempre** por su puerta. La séptima, el 2026-10-01: **el directorio de pruebas y Keycloak dejan `dev.yml`** y pasan a `active-directory.yml` y `keycloak.yml`, con **su propio comando** y la red `catalina-support-dev` que posee `dev.yml`; **el perfil `auth` desaparece** (sección 11 y decisión 1) |
| `modules/users.md` | **as-built** (5 enmiendas) | El módulo `users`: la tabla de cuentas, los endpoints, las reglas del alta, el cambio de papel, la desactivación y el perfil propio. Enmendado el 2026-09-23 (los códigos de cada clave, las claves que faltaban y el reenvío a cuentas desactivadas) y dos veces el 2026-09-24 (nadie se desactiva a sí mismo y la cuenta de fábrica no tiene perfil; y **las cinco decisiones de sus pantallas**), y dos veces el 2026-09-25: al implementar **las tres acciones que dependían del directorio** —reactivar pregunta al directorio en AD y no se hace a mano en Keycloak, y el alta y el cambio de origen hacia el directorio no se hacen a mano—. **Está entero**: backend, las tres pantallas y esas tres acciones. La última, también del 2026-09-25: la pregunta al directorio **la contesta `auth`**, que es quien tiene la configuración guardada |
| `interfaz-y-experiencia.md` | **as-built** (36 enmiendas) | Principios, forma de la aplicación, los cuatro enfoques por papel, la vista doble, el lenguaje visual (Tailwind v4 y 20 componentes propios), los ocho temas, multi-dispositivo y accesibilidad. Su registro tiene treinta y seis entradas: dos del **2026-09-23**, seis del **2026-09-24**, tres del **2026-09-25**, cuatro del **2026-09-26**, seis del **2026-09-27**, seis del **2026-09-28**, ocho del **2026-09-29** y una del **2026-09-30** |
| `primer-arranque.md` | **as-built** (1 enmienda) | **La vista de instalación**: qué pregunta —la instalación, cómo se entra, la región y la dirección, y el correo—, cómo sabe la aplicación que no está instalada y por qué **no** pregunta el modelo de IA ni la contraseña de fábrica. **Hecho y verificado**. Enmendado el 2026-09-30: **el asistente prueba lo que pide** —los dos endpoints públicos `POST /api/setup/entry/test` y `POST /api/setup/mail/test`, con el mismo candado del sello, y los botones del paso 2 y del paso 4—, y la prueba del correo **comprueba la conexión y la autenticación sin mandar ningún correo** (sección 3.1) |
| `ambientes.md` | **as-built** (18 enmiendas) | El runbook de despliegue, las migraciones —`v1.0.0.sql` para los dos entornos y `v1.0.0_dev.sql` con los datos de ejemplo, sólo en desarrollo—, **las copias de seguridad** (`scripts/backup-db.sh`, con catorce días de retención y **las dos bases en el `cron`**) y las tres capas de pruebas. Sus diecisiete enmiendas incluyen: el 2026-09-25, **el directorio y Keycloak salen del entorno** y los datos de ejemplo dejan los dos caminos configurados, y ese mismo día, al hacer **el primer despliegue de producción**, el guion `scripts/prod-build.sh`, los contenedores verificados por dentro y **el dominio todavía en 503**; el 2026-09-26, el dominio de las direcciones de ejemplo pasa a `@demo.com` y se corrige **la condición de la puesta al día de la migración** (sección 5); el 2026-09-30, **el despliegue queda aparcado** y la sección 4 lleva una nota de que su flujo cambió (rutas relativas, publicación del compose y puertos sin `127.0.0.1`), a la espera de ponerse al día cuando se retome; y el 2026-10-01, los **tres elementos opcionales** con su comando (sección 9.3) y el directorio y Keycloak **levantables por separado** sin romper el comando de siempre; y ese mismo día, **la sección 3.3 deja de repetir las once cuentas de ejemplo** y apunta a la tabla del `README.md`. Enmendado otra vez el 2026-10-01: el directorio de pruebas y Keycloak **dejan `dev.yml`**, pasan a `active-directory.yml` y `keycloak.yml` con su propio comando y **el perfil `auth` desaparece** (secciones 3.3 y 9.3). Y enmendado de nuevo el 2026-10-01: las secciones 3.1, 3.3 y 5 dicen que **en una base nueva es `./scripts/dev-seed.sh` lo que crea el esquema** —paso 1, `v1.0.0.sql`— y que se pasa **antes de abrir la aplicación**, porque la aplicación **no crea tablas al arrancar**. Y otra vez el 2026-10-01: la sección 3.2 estrena **los dos comandos de Docker que valen igual en PowerShell, CMD y bash** (`cp` + `psql -f`) y las secciones 3.1, 3.3 y 5 separan **las dos vías** —el guion, de Linux y macOS, y los comandos de Docker, para Windows—, con lo comprobado sobre una base nueva. Y otra vez el 2026-10-01: la sección 3.1 dice que **también se entra por `http://127.0.0.1:11001` sin nginx** (el servidor de desarrollo reenvía `/api`) y la 9.3 que `BASE_URL` puede apuntar ahí para probar el proxy. Y otra vez el 2026-10-01: **los datos de ejemplo se siembran con `docker compose -f dev.yml run --rm seed`** —contenedor de un solo uso, con perfil propio para que `up -d` no lo arranque, corriendo el guion `scripts/dev-seed.sh`—, igual en Linux, macOS y Windows, con los ejemplos **opcionales** y el esquema obligatorio; quedan al día las secciones 3.1, 3.3, 5 y 9.3 |
| `modules/ai.md` | **as-built** (2 enmiendas) | El módulo `ai`: los dos resúmenes del ticket —**«Motivo»** y **«Última acción»**— que redacta un motor de inteligencia artificial **en un contenedor aparte**, en español y en inglés. El motor y su modelo, qué texto se le manda, la tabla `ai_insights`, el contrato entre módulos, los reintentos y los estados, y lo que **no** hace. **Escrito el 2026-09-27**; **enmendado el 2026-09-30** con su dirección y su modelo **configurables desde Configuración**, y **el 2026-10-01** para corregir los recursos: **1500m de memoria**, **sin tope de CPU** y **~1,44 GiB medidos** |
| `modules/tickets.md` | **as-built** (12 enmiendas) | Modelo de datos, numeración, transiciones de los dos ciclos de vida, adjuntos y la lista cerrada de endpoints. Aprobado tras tres repasos, y **enmendado el 2026-09-24** al empezar a implementarlo: dos erratas, entre quién se reparte un interno, quién reabre el interno, el chip de tipo de la bandeja y las cuatro decisiones que el modelo no tenía resueltas; y otra vez ese día, al construir las pantallas. **Está entero**: backend y las pantallas. El 2026-09-25 llegó **el guion de copias** y el 2026-09-26 **los adjuntos dentro del texto** (sección 2.3): el cuerpo con formato, la lista blanca de etiquetas, las cuatro extensiones de vídeo y las decisiones 45 a 49 |

La cadena de producto **está completa**: siete documentos que cubren qué se construye, cómo se
comporta, cómo se ve y cómo se despliega. Además, **cada módulo tiene su documento**, que se escribe
justo antes de implementarlo, en este orden:

1. `modules/mail.md` — **entero**: backend y el editor.
2. `modules/settings.md` — **la pantalla de Configuración está hecha entera**: la marca, el color, el
   prefijo, el reparto y el idioma de la instalación.
3. `modules/auth.md` — **entero**: los tres caminos de entrada, hechos y verificados.
4. `modules/users.md` — **entero**: backend, las tres pantallas y las tres acciones del directorio.
5. `modules/tickets.md` — **entero**: backend y pantallas, con los adjuntos de los comentarios.

## Convenciones

### Cabecera obligatoria

Todo documento de `docs/` empieza así:

```markdown
# Título

> **Estado:** propuesta | aprobado | as-built
> **Última actualización:** AAAA-MM-DD
```

- `propuesta`: escrito, pendiente de aprobación. No se toca código de esa área.
- `aprobado`: aprobado por el responsable del proyecto. Habilita implementar.
- `as-built`: el código existe y el documento describe lo que hace hoy.

### El registro de enmiendas y cómo se cuenta

Debajo de la cabecera, cada documento lleva **su registro de enmiendas**, de la más reciente a la más
antigua: **una entrada por cambio**, con su fecha y qué cambió y por qué.

**Una enmienda es una entrada del registro que empieza por `Enmendado el …`.** Eso es lo que dice el
número que aparece en la tabla de arriba y en `AGENTS.md`, y se puede comprobar contando esas
entradas en la cabecera del documento. **`Pasa a as-built el …` no cuenta**: no cambia nada del
documento, cambia su estado.

Dos cosas se escriben aparte, y no son enmiendas:

- **La corrección del responsable**, cuando cambió lo propuesto, se anota **dentro de la entrada**
  («es una corrección del responsable», «yo proponía… y decidió…»).
- **Lo que se descubrió al implementar** también va dentro de su entrada, diciendo qué se encontró.

El registro es memoria de **por qué** un documento está como está. No se tachan ni se borran
entradas: si algo se corrige, se añade la entrada nueva y se dice qué queda corregido.

### Nombres y carpetas

- **Los documentos de producto** viven en `docs/`, y se llaman en español, en minúsculas, sin acentos
  ni espacios: `usuarios-y-permisos.md`.
- **Los documentos de módulo** viven en `docs/modules/` y se llaman **como el módulo**:
  `modules/auth.md`, `modules/users.md`. El nombre del módulo manda ahí, aunque sea en inglés, porque
  es el nombre que tiene su carpeta en el código.
- Un documento de módulo cuenta **cómo se construye** ese módulo. Lo que es decisión de producto —qué
  puede cada papel, cómo se comportan los tickets— vive en los documentos de `docs/`, y el de módulo
  no lo repite: lo da por escrito y apunta a él.

### Correspondencia con el código

La tabla documento ↔ código vive en `AGENTS.md`. Cuando un documento pase a `as-built`, hay que
añadir o actualizar su fila allí en el mismo cambio.

### Reglas de escritura

- Español, frases cortas, sin relleno y sin adornos de marketing.
- Cada documento dice **qué está decidido**, **qué está pendiente** y **qué queda fuera de alcance**.
- Lo que no esté decidido se marca como pendiente; no se rellena con suposiciones.
- Cuando un documento describa algo ya implementado, describe lo que el código hace de verdad,
  no lo que se pretendía.
