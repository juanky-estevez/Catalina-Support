# Catalina-Support

**Mesa de ayuda de software libre**, de dos niveles y sin ruido:

```text
usuario (reporta)  ->  soporte técnico (nivel 1)  ->  desarrollo (nivel 2)
```

Nació para una institución que necesitaba lo esencial —que un usuario reporte, que Soporte atienda,
que lo que Soporte no puede resolver llegue a Desarrollo ya filtrado y con contexto— y **sin el
ruido de las herramientas grandes**: pocos estados, pocos campos obligatorios, ninguna notificación
que no aporte y ningún paso que exista sólo porque otro producto lo tiene.

Es **software libre** con licencia **MIT** (`LICENSE`): se puede usar, estudiar, modificar y
distribuir, en local para probarlo o en producción para usarlo de verdad. **Todo corre en
contenedores** y **la configuración se hace desde la propia interfaz**: no hay que editar archivos
en el servidor para poner el nombre de la instalación, el método con el que entra la gente, su
región horaria o su dirección pública.

## Donaciones y cómo participar

> **Hueco pendiente.** Aquí van la invitación a participar y las vías para donar. El responsable del
> proyecto rellenará este apartado con los datos que quiera publicar (una página del proyecto, un
> enlace de donación, un correo de contacto, un repositorio…). **No se ha inventado ninguna
> dirección ni ninguna cuenta.** Cuando estén, se escriben aquí y se quitan estas dos líneas.

Mientras tanto, si quieres **participar**: el proyecto se desarrolla con documentación antes que
código (`AGENTS.md`, **Regla 0**), así que la mejor puerta es leer `docs/` —empieza por
`docs/README.md`— y proponer el cambio en el documento del área que toque antes de escribir una
línea de código.

## Qué hace

**Un flujo de dos niveles, con el contexto viajando con el ticket.** El usuario reporta; Soporte
atiende, resuelve y cierra lo que puede; y lo que no, **se escala a Desarrollo**, que recibe un
ticket **interno** enlazado al principal: el usuario nunca ve el interno, ni la palabra «escalado»,
y las dos conversaciones no se mezclan.

- **Tickets** con numeración propia (`CS-2026-0001`), categoría obligatoria y etiquetas, adjuntos
  (imágenes, vídeo, PDF, Office, texto y código), comentarios con sus archivos, una sola línea de
  tiempo y **vista doble** —principal e interno— cuando hay escalado.
- **Estados** con sus transiciones: nuevo, en progreso, en espera, resuelto, cerrado y escalado. Cada
  cambio **explica lo que va a pasar** antes de hacerlo, y **cerrar exige decir por qué**.
- **Dos resúmenes por ticket los redacta un motor de IA propio** —«Motivo» y «Última acción»—, en
  español y en inglés. **El motor se levanta aparte** —como el directorio y Keycloak— y es **opcional**:
  sin él la mesa de ayuda **funciona entera**, sin los dos resúmenes, y se añade después desde
  Configuración.
- **Personas y permisos**: usuario, Soporte, Desarrollo y Administrador, cada uno con lo suyo.
  Soporte no escribe en el interno y Desarrollo no escribe en el principal: **la interfaz no ofrece
  lo que no se puede hacer**, y el servidor tampoco lo acepta.
- **Usuarios** con alta por enlace, cambio de contraseña, ficha, desactivar y reactivar.
- **Tres formas de entrar**, **una a la vez**: cuenta local (con la cuenta de fábrica siempre
  dentro), **Active Directory** o **Keycloak**.
- **Once correos** editables desde la interfaz, en español y en inglés, con vista previa.
- **Configuración de la instalación**: nombre y logo propios, color institucional, idioma, prefijo
  y reparto de tickets, región horaria y dirección pública.
- **Ocho temas**, que siguen al sistema hasta que alguien elige.
- **Todo en español y en inglés**, y **toda la interfaz con texto traducido**: una clave que falte en
  un idioma no compila.

Lo que existe, lo que está verificado y lo que falta están en `docs/README.md` y en la sección 13 de
`docs/arquitectura.md`.

## Cómo levantarlo

La máquina **sólo necesita Docker** (con el plugin `docker compose`). No hay que instalar Go ni Node
—todo se construye y se ejecuta dentro de contenedores— y **no hacen falta imágenes previas**: los
contenedores se compilan **desde el código del repositorio**.

**Lo que necesita la máquina**, para elegir dónde probarlo:

| | |
| --- | --- |
| **Docker** con `docker compose` | Es lo único que hay que tener instalado |
| **Memoria** | **1 GB** para probar en local **sin** el motor de IA; **2 GB** con él, y **4 GB** para ir cómodo. El motor es lo que más pide: **reserva 1,5 GB** |
| **Disco** | Unos **2 GB** para las imágenes y la base de datos, **más 1,1 GB** si se levanta el motor de IA (el modelo, en su volumen) |
| **Procesador** | Cualquiera. **No hace falta GPU**: el modelo va por CPU |
| **Sistema** | Linux, macOS o Windows con Docker; en Linux, el usuario tiene que poder ejecutar `docker` |

### Probarlo en local (desarrollo)

Este es el recorrido mínimo para **ver la mesa de ayuda funcionando**: clonar, levantar, aplicar el
esquema y entrar. Son cuatro pasos, y **los mismos comandos valen en Linux, macOS y Windows**
—PowerShell, CMD o bash: todo es `docker compose`, sin guiones de shell ni redirecciones del
intérprete—.

**1. Traer el código y entrar en la carpeta** (sólo la primera vez):

```bash
git clone <el-repositorio> catalina-support
cd catalina-support
```

**2. Levantar el entorno** —frontend, backend, PostgreSQL y el buzón de pruebas—:

```bash
docker compose -f dev.yml up -d
docker compose -f dev.yml ps
```

**Es un solo comando de Docker, igual en Linux, macOS y Windows**: la red del entorno la crea
`dev.yml`, así que no hay que crear nada a mano. El directorio de pruebas, Keycloak y el motor de IA
**no se levantan aquí**: son opcionales y van aparte, en **«Los tres elementos opcionales»**, más
abajo. Para este recorrido **no hace falta ninguno de los tres**.

**3. Aplicar el esquema de la base.** Este paso hace falta porque **la aplicación no crea las tablas
al arrancar** y **las migraciones no se aplican solas**: el esquema vive en `backend/migrations/` y
se aplica a propósito. **Si se salta**, la aplicación contesta **«Something failed on our side.
Please try again in a moment»** al abrirla, y en los registros del backend (`docker compose -f dev.yml
logs backend`) se lee esto:

```text
ERROR: relación "installation_settings" does not exist
ERROR: relación "ai_insights" does not exist
```

Estos dos comandos de Docker, **iguales en los tres sistemas**, copian el archivo al contenedor y se
lo dan a `psql` con `-f`. La migración es **transaccional e idempotente**, así que se puede volver a
pasar sin miedo (en una base que ya la tenga, avisa de lo que ya existe y termina en `COMMIT`):

```bash
docker compose -f dev.yml cp backend/migrations/v1.0.0.sql database:/tmp/esquema.sql
docker compose -f dev.yml exec -T database psql -U catalina_support -d catalina_support -p 11003 -v ON_ERROR_STOP=1 -f /tmp/esquema.sql
```

**En Linux y macOS, además, se puede hacer sin copiar**, con la redirección del intérprete —que **no
existe en PowerShell**, y por eso no es la forma de arriba—:

```bash
docker compose -f dev.yml exec -T database psql -U catalina_support -d catalina_support -p 11003 -v ON_ERROR_STOP=1 < backend/migrations/v1.0.0.sql
```

**4. Abrir la aplicación y entrar.** La interfaz queda en **`http://127.0.0.1:11001`** y el backend en
`http://127.0.0.1:11002` (`GET /api/health` responde). **El servidor de desarrollo reenvía `/api` al
backend**: su configuración de proxy (`frontend/proxy.conf.json`, declarada en `serve.options` de
`frontend/angular.json`) manda las peticiones del SPA a **`http://backend:11002`**, el servicio del
backend **dentro de la red del entorno** —quien hace la petición es el contenedor del frontend, no el
navegador—. Como el frontend llama a `/api/**` **en relativo**, abrir `http://127.0.0.1:11001` deja la
pantalla funcionando y hablando con la API **sin nginx y sin nada más**: es el recorrido de quien se
descarga el proyecto y sólo tiene Docker. Y **con el esquema ya aplicado se entra con la cuenta de
fábrica**: usuario **`admin`** y contraseña **`admin`** —la de `config/env/dev.env`, que es un archivo
versionado porque es una credencial de contenedor local—.

**El proxy es sólo de desarrollo**: vive en `serve.options`, no toca el `build`, y **en producción
sigue siendo nginx quien reenvía `/api`** (`docs/arquitectura.md`, sección 9). La regla no cambia: el
frontend llama a rutas **relativas** y nunca a un host.

**La contraseña de la cuenta de fábrica no está en la base y no hace falta ninguna cuenta más**: vive
en la configuración del entorno (`ADMIN_PASSWORD`) y se comprueba en cada entrada. En desarrollo el
valor es `admin`; **en producción es el que ponga `config/env/prod.env`** —que no se versiona y **no
se escribe en la documentación**—, y **no hay que copiar el de desarrollo a un servidor**
(`docs/usuarios-y-permisos.md`, sección 8).

**Los datos de ejemplo son un extra, no un requisito para entrar.** Dejan **once cuentas** y **25
tickets** con su historia y sus adjuntos, para no entrar a un sistema vacío. Se aplican **después**
del esquema, **con un comando de Docker igual en Linux, macOS y Windows** —es un contenedor de un solo
uso, que arranca, siembra y se va—:

```bash
docker compose -f dev.yml run --rm seed
```

**Quien quiera probar el proyecto vacío no ejecuta este comando**: el esquema es **obligatorio** y los
ejemplos son **una elección**. El comando tiene **perfil propio** (`seed`), así que **no se levanta con
`docker compose -f dev.yml up -d`**: se pide a propósito.

**Ese comando corre el guion de siempre, `scripts/dev-seed.sh`, dentro de un contenedor**: por eso vale
en los tres sistemas y hace los tres pasos, incluida **la copia de los adjuntos**. Aplica el esquema
—por si la base está recién creada—, aplica `v1.0.0_dev.sql` y copia los archivos de los adjuntos de
ejemplo desde `config/seed/` a `_files/`. **En Linux y macOS sigue existiendo `./scripts/dev-seed.sh`**
y hace lo mismo fuera del contenedor; el comando de Docker es la forma de los tres sistemas.

**Cuidado: los ejemplos borran los tickets que hubiera en desarrollo** —también los creados probando—
y dejan la base en un estado conocido. **No se usan en producción** (el archivo `_dev` no se aplica
allí).

Las **once cuentas de ejemplo** entran todas con la misma contraseña, **`123123123`**, que **no se usa
en producción**:

| Nombre | Correo | Rol | Contraseña |
| --- | --- | --- | --- |
| Usuario Uno | `user1@demo.com` | `usuario` | `123123123` |
| Usuario Dos | `user2@demo.com` | `usuario` | `123123123` |
| Usuario Tres | `user3@demo.com` | `usuario` | `123123123` |
| Usuario Cuatro | `user4@demo.com` | `usuario` | `123123123` |
| Usuario Cinco | `user5@demo.com` | `usuario` | `123123123` |
| Soporte Uno | `support1@demo.com` | `soporte` | `123123123` |
| Soporte Dos | `support2@demo.com` | `soporte` | `123123123` |
| Soporte Tres | `support3@demo.com` | `soporte` | `123123123` |
| Desarrollador 1 | `dev1@demo.com` | `desarrollo` | `123123123` |
| Desarrollador 2 | `dev2@demo.com` | `desarrollo` | `123123123` |
| Desarrollador 3 | `dev3@demo.com` | `desarrollo` | `123123123` |

**La cuenta de fábrica (`admin`) no es una de estas** y entra sin los ejemplos: **los ejemplos son
para tener contenido, no para poder entrar**.

- **El correo de desarrollo no sale a internet**: va a un buzón de pruebas (Mailpit) que se lee en
  `http://127.0.0.1:11004`. Ahí se ve el enlace de alta de una cuenta sin configurar ningún SMTP.
- **El motor de IA, el directorio de pruebas y Keycloak son aparte y son opcionales**: los tres se
  levantan con su propio comando y **ninguno hace falta** para que la mesa de ayuda funcione. Están
  juntos, con lo que es cada uno y su comando, en **«Los tres elementos opcionales»**, justo abajo.

En desarrollo también hay **nginx delante**, para probar con el dominio y el certificado de verdad;
**no hace falta para el primer contacto**, porque la 11001 ya reenvía `/api`.

### Los tres elementos opcionales

El comando de arriba (`docker compose -f dev.yml up -d`), **con la base ya preparada** (arriba), levanta
todo lo que hace falta para usar la mesa de ayuda. Estos tres van aparte, **son opcionales** y **cada
uno tiene su propio comando**: se
puede levantar cualquiera de ellos sin levantar los otros, y **sin ninguno la aplicación funciona
entera**.

**1. El motor de IA**, que redacta los dos resúmenes del ticket —**«Motivo»** y **«Última acción»**—
en español y en inglés:

```bash
docker compose -f ai.yml up -d
```

Es **llama.cpp** sirviendo un modelo **Qwen2.5-1.5B-Instruct** en cuantización **Q4_K_M**
(`qwen2.5-1.5b-instruct-q4_k_m.gguf`). La primera vez **descarga el modelo (~1,1 GB)** al volumen; el
contenedor **reserva 1,5 GB de memoria** (`mem_limit: 1500m` en `ai.yml`) y, medido en marcha, gasta
**~1,44 GiB**. **Va por CPU** —no hace falta GPU— y **tarda entre 12 y 24 segundos por campo**. Sin
él, los dos resúmenes se quedan sin texto y todo lo demás funciona igual. Más abajo, en **«El motor
de IA, y lo que consume»**, está el detalle y cómo cambiarlo por otro modelo.

**2. El directorio de pruebas (OpenLDAP)**, para probar el camino de entrada por **Active Directory**:

```bash
docker compose -f active-directory.yml up -d
```

Trae **las personas de prueba del directorio** (`config/ldap/`) y publica el puerto `11005`. Con el
método de entrada en `ad`, la aplicación entra contra él como entraría contra el directorio real de la
institución. **Va en su propio archivo** (`active-directory.yml`) y entra en la red del entorno de
desarrollo, que la crea `dev.yml`: por eso **primero se levanta el entorno y después el servicio**.

**3. Keycloak**, para probar el camino de entrada por **OIDC**:

```bash
docker compose -f keycloak.yml up -d
```

Trae **un reino de pruebas que vive en el repositorio** (`config/keycloak/`) y se importa al
arrancar; se publica por `/sso/`, en el puerto `11006`. Como el directorio, **va en su propio archivo**
(`keycloak.yml`) y entra en la red del entorno, así que se levanta **después de `dev.yml`**.

**Cada uno se levanta con su comando y con su archivo**: el directorio y Keycloak **ya no viven en
`dev.yml`** y ya no hay perfil `auth`. Para probar **los dos caminos de directorio** en la suite hay
que levantar los dos y después correr las pruebas de siempre (ver **«Las pruebas»**).

**En producción no se levanta ninguno de los dos**: se configura el directorio o el reino que ya
exista —su servidor, sus credenciales— en la pantalla de **Configuración**, y la instalación entra
contra él. Lo que trae el proyecto es el camino para hablar con ellos, no el servidor.

### Ponerlo a funcionar de verdad (producción)

1. **Traer el código** a la máquina, con Docker ya instalado:

   ```bash
   git clone <el-repositorio> catalina-support
   cd catalina-support
   ```

2. **La configuración de la instalación**, en `config/env/prod.env`, partiendo del ejemplo
   `config/env/prod.env.example`. **Ese archivo no se versiona** y lo que se ponga ahí no aparece en
   la interfaz. Lo que hay que tocar:

   | Variable | Qué es |
   | --- | --- |
   | `ADMIN_PASSWORD` | La contraseña de **la cuenta de fábrica** (`admin`): la puerta para entrar la primera vez, y la única que entra **siempre**, sea cual sea el método configurado |
   | `TOKEN_SECRET` | El secreto con el que se firman las sesiones. **Largo y distinto en cada instalación** |
   | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | Los datos de la base. La contraseña, puesta aquí, es la que se le da al contenedor al crearla |
   | `AI_URL`, `AI_MODEL` | **Opcional**: dónde está el motor de IA y con qué modelo. **La fuente es Configuración** —con su botón de «Probar la conexión»—, y esto queda como respaldo para una instalación que ya lo tuviera puesto aquí |
   | `TIMEZONE` | La zona del **contenedor** (sus registros). **No** es la región de la instalación: esa se elige en la pantalla, y es la que decide cómo se leen las fechas |
   | `PUBLIC_APP_URL` | **Opcional**: la dirección pública se configura en la pantalla y esto queda como respaldo |

   **Las demás vienen puestas** y no hace falta tocarlas: los puertos internos, las carpetas de
   registros y adjuntos, y el idioma de los contenedores.

3. **Construir, publicar y levantar**:

   ```bash
   ./scripts/prod-build.sh                 # construye, publica y levanta los contenedores
   ./scripts/prod-build.sh --no-deploy     # sólo construir y publicar
   ./scripts/prod-build.sh --only backend  # un solo componente
   ```

4. **El esquema de la base**, aplicado en orden hasta la versión que se despliegue. Es
   **transaccional e idempotente** (se puede volver a aplicar sin miedo):

   ```bash
   docker compose -f prod.yml exec -T database \
     psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d catalina_support < backend/migrations/v1.0.0.sql
   ```

   Y las versiones siguientes, en orden (`v1.1.0.sql`, `v1.2.0.sql`…). **Los archivos `_dev` no se
   aplican nunca en producción**: son los datos de ejemplo.
5. **nginx y el certificado**, con el vhost de `config/nginx/`. Los vhosts del repositorio son
   **copias** de `/etc/nginx/conf.d/` (no enlaces): si se cambia uno, hay que copiarlo y recargar.
   ```bash
   sudo cp config/nginx/catalina-support-prod.conf /etc/nginx/conf.d/
   sudo nginx -t && sudo systemctl reload nginx
   ```

   **Lo que hay que abrir en el cortafuegos es nginx**: el 80 y el 443. Los puertos de los
   contenedores se publican **sólo en `127.0.0.1`**, para depurar desde la propia máquina; no se
   exponen a internet.
6. **Abrir la instalación y recorrer sus cuatro pasos.** Con el esquema aplicado, al entrar la
   aplicación lleva a **`/setup`**: el nombre y el idioma, cómo se entra (con su prueba de conexión),
   la región horaria y la dirección pública, y el correo saliente. Al terminar, **la instalación queda
   sellada** y se entra con **la cuenta de fábrica** (`admin`, con la contraseña del archivo de
   entorno), que es la puerta que entra siempre. Lo demás se cambia después en **Configuración**:
   - **el nombre y el logo** de la institución, y su **color**;
   - **el método de autenticación** —local, Active Directory o Keycloak—, con sus datos y su botón
     de «Probar la conexión»;
   - **la región horaria**: la zona en la que se leen todas las fechas, aquí y en los correos. Se
     guardan en UTC, así que cambiarla no mueve ningún ticket;
   - **la dirección pública**: la base de los enlaces que van en los correos y de la vuelta de
     Keycloak. Sirve http o https, con puerto si hace falta, y también `localhost`. **Si no es https,
     la pantalla lo avisa** —la contraseña y la sesión viajan sin cifrar— y no bloquea nada: para
     probarlo en local está bien, y para usarlo en serio conviene un certificado;
   - **el idioma de la instalación**, el prefijo de los tickets y su reparto;
   - **el motor de IA**: su dirección y su modelo, con su botón de «Probar la conexión». Es opcional:
     sin él la mesa de ayuda funciona entera, sin los dos resúmenes del ticket.
> **La primera vez, la aplicación te lleva a su vista de instalación.** Al abrirla en una instalación
> nueva —la base recién creada, sin el sello de instalación— aparece **`/setup`**: cuatro pasos que
> piden **el nombre y el idioma, cómo se entra, la región y la dirección, y el correo saliente**, con
> un resumen al final. **Al terminar queda sellada**: esa vista no vuelve a aparecer y la API rechaza
> reconfigurarla (`docs/primer-arranque.md`). Lo que se ponga después se cambia en **Configuración**,
> y **la contraseña de la cuenta de fábrica sigue viniendo del archivo de entorno** (`ADMIN_PASSWORD`):
> el asistente no la pide, sólo dice dónde está.

7. **Las copias de la base**, que es lo que va en el `cron` del servidor:

   ```bash
   ./scripts/backup-db.sh          # el entorno de desarrollo
   ./scripts/backup-db.sh prod     # producción
   ```

### El motor de IA, y lo que consume

**Va aparte y es opcional**, como el directorio y Keycloak: se levanta con su propio comando y **un solo
motor sirve a desarrollo y a producción** (el modelo ocupa ~1,1 GB, en un volumen). **La primera vez baja
su modelo**, y se ve en sus registros:

```bash
docker compose -f ai.yml up -d         # la primera vez descarga el modelo (~1,1 GB)
docker compose -f ai.yml logs -f ai    # la descarga y, después, el servidor
```

**Sin él, la mesa de ayuda funciona entera**: los dos campos que redacta se quedan sin texto y la
pantalla lo dice, para que nadie se quede sin trabajar mientras se arregla. Y se **añade cuando se
quiera**, sin reinstalar nada.

**Su dirección y su modelo se configuran en Configuración**, con el botón de «Probar la conexión», que
pregunta a la comprobación de salud del motor. La instalación los lee **en cada petición**, así que
cambiarlos vale sin reiniciar nada. Las variables `AI_URL` y `AI_MODEL` del archivo de entorno quedan
**como respaldo** para una instalación que ya las tuviera puestas; `AI_PALABRAS` y
`AI_ESPERA_SEGUNDOS` siguen siendo ajuste del entorno.

**El modelo por defecto** es `qwen2.5-1.5b-instruct`, en cuantización `q4_k_m` —el archivo
`qwen2.5-1.5b-instruct-q4_k_m.gguf`—:

| | |
| --- | --- |
| **Descarga la primera vez** | **1,1 GB** (1.117.320.736 bytes), al volumen del modelo (no está en el repositorio). Sale del repositorio **`Qwen/Qwen2.5-1.5B-Instruct-GGUF`** de Hugging Face, con el nombre exacto del archivo (`config/ai/01-descargar-modelo.sh`) |
| **Memoria que reserva el contenedor** | **1,5 GB** (`mem_limit: 1500m` en `ai.yml`) —el modelo, su caché y el servidor—. Medido en marcha: **~1,44 GiB** |
| **Tiempo por resumen** | **entre 12 y 24 segundos** por campo, en CPU |
| **Procesador** | **no hace falta GPU**: va por CPU |

**Cambiarlo por otro modelo** es dejar su `.gguf` en el volumen y poner su nombre en el servicio `ai`
de `ai.yml` (`--model /modelos/<archivo>.gguf`), **ajustando el tope de memoria** del contenedor a lo
que pida el nuevo: **un modelo más grande pide más memoria y responde más rápido**, y el tope está
puesto justo por encima de lo que ocupa el de por defecto. El nombre del modelo que se le pide al
motor se elige en **Configuración**, y la instalación **comprueba** que responde.

## Las pruebas

Las tres capas están montadas y se ejecutan **en contenedores**:

```bash
docker compose -f dev.yml exec backend  go test ./...            # el backend
docker compose -f dev.yml exec frontend npm test -- --watch=false # la interfaz, unitaria
docker compose -f dev.yml run --rm e2e                            # los recorridos, en un navegador
docker compose -f dev.yml run --rm seed                           # reiniciar el entorno después

# Los recorridos de los dos caminos de directorio, en dos pasos: se levanta cada servicio con su
# archivo y se corre la suite de siempre. El directorio y Keycloak ya no viven en `dev.yml`
docker compose -f active-directory.yml up -d
docker compose -f keycloak.yml up -d
docker compose -f dev.yml run --rm e2e
```

La capa de interfaz (Playwright) prueba los recorridos de verdad —entrando como cada papel, leyendo
los correos del buzón de pruebas— **en PC y en móvil**.

**Para reiniciar el entorno después de una pasada**, el comando es el mismo en los tres sistemas:
`docker compose -f dev.yml run --rm seed`, en **«Probarlo en local (desarrollo)»**. En Linux y macOS
también se puede correr el guion directamente, con `./scripts/dev-seed.sh`.

## Antes de tocar el código

Leer `AGENTS.md`. La **Regla 0** es obligatoria: **primero el documento aprobado, después el
código.** El mapa de qué documento cubre qué está en `AGENTS.md`, y el índice con el estado de cada
documento en `docs/README.md`.

## Licencia

**MIT** (`LICENSE`). Úsalo, estúdialo, modifícalo y distribúyelo.
