# Arquitectura

> **Estado:** as-built
> **Última actualización:** 2026-09-30
>
> **Enmendado el 2026-09-30**: **la sección 13 se pone al día**. Lo que ha cambiado este mes y no estaba
> contado: **la pantalla de Configuración está entera** —con **el motor de IA** y su prueba de conexión,
> la región horaria y la dirección pública, y **el correo saliente**—; hay **vista de primer arranque**
> (`/setup`, cuatro pasos y el sello en `installation_settings.installed_at`, con su API en **409**
> después); **el motor de IA vive aparte y se configura desde Configuración**; y **los recuentos** quedan
> en Go **14 paquetes**, frontend **215 pruebas** (20 ficheros) e interfaz **214 casos** (196 en verde, 18
> saltados, 0 rojos). **El despliegue a producción queda aparcado** hasta que el producto esté terminado
> (decisión del responsable, 2026-09-30), así que no entra en esta puesta al día. Lo que falta está dicho
> abajo.
>
> Aprobado por el responsable del proyecto el 2026-09-22 y **puesto en pie el mismo día**: arrancó el
> esqueleto de contenedores con los dos proyectos. Lo que existe hoy está en la sección 13, y cada
> módulo tiene su propio documento.

## 1. Alcance de este documento

Fija el **stack**, la **forma del repositorio**, los **contenedores** y las **reglas de
modularidad** de Catalina-Support. No define producto ni flujos: eso vive en los documentos de
cada área (propósito y alcance, usuarios y permisos, tickets, flujos).

Este documento está **as-built**: describe lo que el código hace hoy, y si el código cambia, el
documento cambia en el mismo trabajo (Regla 0 de `AGENTS.md`).

## 2. Stack

Versiones comprobadas el **2026-09-22** contra su fuente oficial. Se fijan las versiones
mayores; las menores se actualizan cuando toque sin pasar por una propuesta nueva.

| Capa | Tecnología | Versión | Fuente de la comprobación |
| --- | --- | --- | --- |
| Frontend | Angular | **22.1.7** (CLI 22.1.8) | `npm view @angular/core dist-tags` → `latest` |
| Frontend | Node.js | **24 LTS** (24.21.0) | `nodejs.org/dist/index.json` (Angular 22 exige `^22.22.3 \|\| ^24.15.0 \|\| >=26.0.0`) |
| Backend | Go | **1.27.1** | `go.dev/dl/?mode=json` |
| Backend | `net/http` (biblioteca estándar) | Go 1.27.1 | decisión del responsable: sin framework HTTP |
| Backend | GORM | **v1.31.2** + driver `gorm.io/driver/postgres` **v1.6.3** | `proxy.golang.org` |
| Backend | air (recarga en desarrollo) | **v1.67.4** (`github.com/air-verse/air`) | `proxy.golang.org` |
| Backend | go-logs | **v1.0.0** (`github.com/juanky-estevez/go-logs`) | `proxy.golang.org` |
| Correo | `net/smtp` (biblioteca estándar) | Go 1.27.1 | igual que Calibyou: sin dependencias nuevas |
| Datos | PostgreSQL | **18** (última menor 18.6) | `postgresql.org/versions.json` → `current: true` |
| Ejecución | Docker + Docker Compose | el de la máquina (29.5.2 comprobado) | `docker --version` |

Notas de las versiones:

- **Angular 22** es la última estable y la única que se instala (`latest`); no se fija una
  versión anterior. El `next` publicado es 22.2.0-rc.0 y **no** se usa.
- **Node 24 LTS** ("Krypton") es la línea LTS vigente y la que recomienda Angular 22. Node 26
  funciona pero todavía no es LTS: no se usa.
- **PostgreSQL 18** es la mayor marcada como `current` por el proyecto PostgreSQL, y se usa la
  imagen oficial `postgres:18-alpine`.
- El paquete `go-logs` declara `go 1.23` como mínimo, así que es compatible con Go 1.27.
- Existe un paquete hermano `github.com/juanky-estevez/logs` (`@jcestevez/logs`) que es un
  **middleware de NestJS** (escribe en el sistema de archivos mediante Node). **No sirve para el
  frontend Angular**, que corre en el navegador: no hay paquete de logs para el navegador
  decidido, y queda como pendiente (sección 12).

## 3. Arquitectura general

```text
navegador
  -> nginx (vhost; enruta /api/ al backend y el resto al frontend)
      -> contenedor frontend: Angular 22 (SPA)
          -> /api/<módulo>/...
      -> contenedor backend: Go (net/http) + GORM
          -> contenedor database: PostgreSQL 18
```

Un solo backend, una sola base de datos y un frontend, cada uno en su contenedor. No hay
microservicios, ni colas, ni caché en esta versión: si hicieran falta, se proponen en su
documento correspondiente.

## 4. Modularidad (regla central)

El proyecto se organiza **por módulo**, y el módulo es la unidad de todo: carpeta, endpoints,
tablas y pantallas. La regla es explícita y verificable:

> **Un módulo del frontend consume únicamente los endpoints de su módulo homónimo del backend.**

De ahí salen cinco reglas concretas:

1. **Un nombre, tres sitios.** Un módulo se llama igual en los tres sitios: carpeta del backend,
   carpeta del frontend y prefijo de la API. Ejemplo: `users` →
   `backend/modules/users`, `frontend/src/app/modules/users` y `/api/users/...`.
2. **El frontend no cruza módulos.** `frontend/src/app/modules/users` sólo llama a
   `/api/users/**`. Si una pantalla necesita datos de otro módulo, se llama al módulo dueño de
   esos datos desde el backend, no se pide desde otra pantalla a un endpoint ajeno.
3. **El backend no cruza tablas.** Un módulo sólo escribe en sus propias tablas. Lo que necesite
   de otro módulo lo pide por una función de servicio expuesta, nunca leyendo su tabla.
4. **`shared` no depende de `modules`.** Lo transversal (configuración, conexión a base de
   datos, middlewares, utilidades) vive en `shared` y no conoce ningún módulo concreto. Un
   módulo sí puede usar `shared`.
5. **`core` es la excepción del frontend.** El armazón —`frontend/src/app/core`— puede llamar a los
   endpoints de la **sesión** y a los de **configuración de la instalación** (el prefijo de
   numeración y las plantillas de correo), y a nada más. Es una sola regla con su motivo: la sesión y
   la configuración de la instalación no son de ningún módulo, y repartirlas en tres excepciones
   sueltas es lo que hace que una regla se olvide.
   **El permiso no obliga a usarlo**: la pantalla que edita las plantillas de correo es del módulo
   `mail` del frontend, no de `core` (decidido el 2026-09-23, `docs/modules/mail.md`, decisión 15),
   porque sus datos son suyos y así la excepción no se convierte en la norma.

Correspondencia visual:

```text
frontend/src/app/modules/users  --HTTP-->  /api/users/**  -->  backend/modules/users  -->  tablas users*
```

Las llamadas del frontend son **siempre relativas** (`/api/users`), nunca a un host absoluto: el
mismo código funciona en desarrollo y en producción porque es nginx quien enruta, y así no hay
CORS que configurar.

Qué se gana: encontrar cualquier cosa de un módulo es mirar su carpeta, y borrar o cambiar un
módulo no obliga a rastrear el resto del repositorio. Qué se paga: hay que respetar la regla a
mano, porque no hay nada que la imponga automáticamente (ver sección 12).

La lista cerrada de módulos **no** se decide aquí: cada documento de área define el suyo.
`users` se usa en este documento sólo como ejemplo de la convención.

## 5. Backend (Go + `net/http` + GORM)

### Forma del repositorio

```text
backend
├── go.mod
├── main.go                  # arranque: configuración, GORM, rutas, servidor
├── .air.toml                # recarga en caliente (sólo desarrollo)
├── migrations
│   └── v1.0.0.sql
├── modules
│   └── <módulo>
│       ├── controllers      # HTTP: leer petición, validar forma, responder
│       ├── services         # reglas del módulo
│       ├── repositories     # acceso a datos
│       └── dtos             # lo que entra y sale por la API
└── shared
    ├── auth                 # el token de sesión (firma y validación) y quién es quien llama
    ├── authz                # permisos: si el papel llega para lo que se pide
    ├── config               # variables de entorno
    ├── database             # conexión y configuración de GORM
    ├── httpx                # la forma única de las respuestas y de los errores
    ├── middleware           # autenticación, permisos, recuperación, logs
    └── version              # la versión del sistema, que sale en la marca pública
```

Se mantiene el estilo ya usado en el proyecto hermano Calibyou (`modules/<m>/controllers`,
`services`, `repositories`, `dtos`), porque es el que el responsable ya conoce y mantiene.

Hoy el repositorio tiene `main.go`, `.air.toml`, `modules/` con sus **seis módulos** (`ai`, `auth`,
`mail`, `settings`, `tickets` y `users`), `migrations/` con `v1.0.0.sql` y `v1.0.0_dev.sql`, y
`shared/{auth,authz,config,database,httpx,middleware,version}`. `shared/httpx` se añadió al crear el esqueleto para que la forma de las respuestas HTTP
tenga un solo dueño desde el principio: hoy escribe `{"error": "clave"}`, con una **clave** que el
frontend traduce al idioma de quien lee (`docs/interfaz-y-experiencia.md`).

### `net/http` sin framework

Se usa el enrutador de la biblioteca estándar, que desde Go 1.22 admite **método y patrón con
parámetros**:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /api/users", h.Listar)
mux.HandleFunc("GET /api/users/{id}", h.Obtener)   // r.PathValue("id")
```

Consecuencias que hay que asumir:

- Los middlewares se encadenan a mano (una función que envuelve `http.Handler`); no hay
  `Use()`. Se resuelve con una cadena explícita en el arranque.
- No hay binding ni validación automática de JSON: cada controlador decodifica, valida y
  responde; los errores se escriben con un helper común de `shared` (respuesta de error única).
- A cambio: cero dependencias de framework y actualizaciones sin sorpresas.

### GORM

- Una conexión única creada en el arranque y compartida (`*gorm.DB` en `shared/database`).
- Los modelos del módulo viven en su `repositories`/`dtos`; no se comparten modelos entre
  módulos.
- Se prefiere **SQL explícito con migraciones** para el esquema (sección 7) y GORM para
  consultas y escritura; `AutoMigrate` **no** se usa, para que el esquema tenga una única
  fuente de verdad.
- Soft delete (`gorm.DeletedAt`) cuando el módulo lo necesite, decidido en su documento.

### Desarrollo con air

- `air` (v1.67.4) recarga el backend al guardar. Va **instalado en la imagen de desarrollo** y es
  el comando del contenedor: no se instala Go en la máquina.
- Su configuración vive en `backend/.air.toml` (vigila `**/*.go`, construye a un binario temporal
  e ignora `tmp/`).
- `air` es **sólo para desarrollo**; el contenedor de producción ejecuta el binario compilado.

## 6. Frontend (Angular 22)

### Forma del repositorio

```text
frontend
├── angular.json
├── package.json
├── .postcssrc.json          # Tailwind v4 (el plugin de PostCSS)
├── src
│   ├── main.ts
│   ├── index.html
│   ├── styles.css           # Tailwind y las variables de los temas
│   └── app
│       ├── app.config.ts    # providers globales (router, http, interceptor)
│       ├── app.routes.ts    # rutas del armazón; cada módulo aporta las suyas
│       ├── core             # armazón: sesión, i18n, interceptores, guardas, pantallas de la sesión
│       │   ├── components   # piezas del armazón (menú lateral, logo, controles, aviso de servidor caído)
│       │   ├── guards       # sin sesión no se entra
│       │   ├── i18n         # los dos diccionarios de textos y el servicio que los sirve
│       │   ├── interceptors # la cabecera de la sesión, el 401 y el «servidor caído»
│       │   ├── navigation.ts # las entradas del menú, por papel
│       │   ├── pages        # las seis pantallas de la sesión y la de Configuración
│       │   └── services     # sesión, salud del backend, disponibilidad
│       ├── shared           # UI y utilidades reutilizables, sin lógica de negocio
│       │   └── components   # los componentes propios (botón, campo, aviso, tarjeta…)
│       └── modules
│           └── <módulo>     # pages, components, services, models, <módulo>.routes.ts
```

### Convenciones (heredadas del Angular moderno, no del Angular antiguo)

- **Componentes standalone** (sin `NgModule`), que es el modo por defecto del CLI desde
  Angular 19. No se crean módulos de Angular; "módulo" aquí significa **carpeta de
  funcionalidad**, no `NgModule`.
- **Señales** (`signal`, `computed`) para el estado de la pantalla y `provideHttpClient` con
  `HttpClient` para las llamadas.
- **Control de flujo nuevo** (`@if`, `@for`, `@switch`) en lugar de `*ngIf`/`*ngFor`.
- **Rutas perezosas por módulo**: cada módulo expone su `<módulo>.routes.ts` y la ruta raíz lo
  carga con `loadChildren`. Una pantalla que no se visita no se descarga.
- **Un servicio por módulo** (`modules/<m>/services/<m>.service.ts`) que centraliza las
  llamadas a `/api/<m>/**`. Ningún componente llama a `HttpClient` directamente, y ningún
  servicio apunta a un módulo ajeno (sección 4).
- Nombres de archivo y de carpeta en **inglés** (es lo que genera el CLI: `user-list.ts`); los
  textos de interfaz y la documentación, en español.
- **Los textos no se escriben dentro de los componentes**: viven en `core/i18n/es.ts` y
  `core/i18n/en.ts`, con el mismo tipo compartido, así que **una clave que falte en un idioma no
  compila**. Los textos de error del backend van en el mismo diccionario
  (`docs/modules/auth.md`, decisión 26).
- **Los colores salen de las variables del tema** de `src/styles.css`, expuestas a Tailwind con
  `@theme inline`: un componente nunca lleva un color escrito a mano, y cambiar el tema no toca
  ningún componente (`docs/interfaz-y-experiencia.md`, sección 6.2). Las paletas se escriben **una
  sola vez** con `light-dark()`, y el tema se elige con un atributo en `html` (`data-theme`), o
  ninguno para seguir al sistema. **El tema lo aplica `main.ts`, no un script incrustado en el
  `index.html`**: la CSP de producción es `script-src 'self'`.
- **Las rutas también en inglés** (`/login`, `/set-password`), en minúsculas y con guiones: decidido
  por el responsable el 2026-09-23 (`docs/modules/auth.md`, decisión 21), porque una dirección se
  copia, se pega y se comparte, y en el código no conviene mezclar idiomas.

Lo que en su día quedó por decidir aquí **ya está decidido en otros documentos**: los estilos y los
componentes propios, en `docs/interfaz-y-experiencia.md` (Tailwind v4, inventario cerrado de veinte
piezas); la gestión de sesión y permisos en el cliente, en `docs/modules/auth.md`, sección 9 (la
sesión vive en `core`: servicio, interceptor, guarda y seis pantallas); y **SSR no se usa**, que es
una decisión y no una ausencia.

Hoy el repositorio tiene **el armazón entero**: el servicio de sesión, el interceptor que pone la
cabecera y atiende el 401, la guarda de rutas, los dos diccionarios de textos, los **componentes
propios** (botón, campo, aviso, tarjeta, conmutador, selector y logo), las **seis pantallas de la
sesión** (`/login`, `/forgot-password`, `/set-password`, `/change-password`, `/forbidden` y el aviso
de servidor caído, que no tiene ruta) y **el menú lateral con sus tres zonas**.

Las rutas están en **dos grupos**: las pantallas de la sesión van a pantalla completa y todo lo demás
va dentro del armazón, con su menú. **El menú enseña sólo lo que existe**, y lo que le toca a cada
papel: las listas de tickets, Usuarios y, para un Administrador, su bandeja y Configuración.

**Lo que falta del frontend**: un repaso de formato. **Las pantallas de producto —usuarios, tickets y
el editor de los correos— están hechas**, y **el inicio provisional desapareció**: la raíz es un
reparto —el Administrador a los usuarios, los demás a su bandeja—.

## 7. Base de datos (PostgreSQL 18)

- Una base de datos y un esquema; el aislamiento por cliente, si llega a existir, se decide en
  el documento de usuarios y permisos.
- **Una única vía para el esquema**: los archivos de `backend/migrations/` (hoy sólo
  `v1.0.0.sql`), transaccionales (`BEGIN` … `COMMIT`) e idempotentes (`CREATE ... IF NOT EXISTS`,
  `INSERT ... WHERE NOT EXISTS`). Se aplican con `psql -v ON_ERROR_STOP=1` dentro del contenedor de
  base de datos. Nada de `ALTER` sueltos ni pasos manuales al desplegar.
- Mientras el proyecto no tenga datos reales, se actualiza ese mismo archivo en vez de añadir
  migraciones nuevas. **A partir de la 1.0.0 se pasa a un archivo por versión**
  (`v1.1.0.sql`, `v1.2.0.sql`…) **más un archivo de datos de ejemplo por versión con el sufijo
  `_dev`** (`v1.0.0_dev.sql`), que **nunca se aplica en producción**, aplicados en orden: la política
  completa está en
  `docs/ambientes.md`, sección 5.
- Nombres de tabla en plural y en inglés, coherentes con el nombre del módulo (`users`,
  `tickets`); claves primarias `id`; marcas de tiempo `created_at`, `updated_at`.
- Los datos viven en un **volumen de Docker**, nunca dentro del contenedor.

## 8. Logs (go-logs)

El backend usa `github.com/juanky-estevez/go-logs` v1.0.0 en lugar de `log`/`slog` y en lugar de
`fmt.Println`. Su API tiene exactamente cuatro funciones:

```go
logs.LogInfo("servidor escuchando en :11002")
logs.LogSuccess("migración aplicada")
logs.LogWarning("intento de acceso sin token")
logs.LogError("no se pudo conectar a la base de datos")
```

Cada llamada imprime en terminal con color **y** agrega una línea al archivo del día:

```text
[info] dev 2026-09-22 10:30:00 servidor escuchando en :11002
```

Variables de entorno que consume:

| Variable | Para qué | Valor |
| --- | --- | --- |
| `PROJECT_NAME` | primera parte del nombre del archivo | `catalina-support` |
| `ENVIRONMENT` | segunda parte del nombre y el texto de cada línea | `dev` / `prod` |
| `LOGS_FOLDER` | carpeta de los archivos | `/logs` (volumen `./_logs` montado en el contenedor) |
| `TIMEZONE` | desplazamiento usado al fechar | `UTC-5` |

El archivo se llama `PROJECT_NAME_ENVIRONMENT_AAAAMMDD.log` (un archivo por día) y se escribe en
`_logs/`, que es una carpeta del repositorio **fuera de git** y montada en el contenedor. Al ser
un volumen, los logs sobreviven a reinicios y se leen desde la máquina sin entrar al contenedor.

**Límites conocidos de la librería** (se aceptan, y por eso quedan escritos):

- No hay niveles configurables ni formato estructurado (no es JSON) ni campos: el mensaje es una
  cadena y nada más. Lo que se quiera consultar después hay que redactarlo dentro del mensaje.
- No hay `Debug` ni trazas: cuatro niveles, cuatro funciones.
- No hay rotación ni borrado: un archivo por día, y crecen. Limpiar `_logs/` es tarea manual.
- Si no puede escribir el archivo, **no falla**: imprime un aviso en rojo y escribe sólo en
  terminal. Un `LOGS_FOLDER` mal configurado pierde el log sin detener la aplicación, así que hay
  que comprobar en el arranque que la carpeta existe.
- No hay identificador de petición ni contexto: correlacionar una traza completa exige incluirlo
  a mano en cada mensaje.

Reglas de uso: nada de `fmt.Println` ni de `log` en el código de la aplicación; el arranque
registra configuración cargada y conexión establecida; los errores que el usuario ve también se
registran; **nunca** se registran contraseñas, tokens ni datos personales.

## 9. El correo

Todo el correo de la aplicación sale por el **módulo `mail`**, con la **biblioteca estándar**
(`net/smtp`), igual que Calibyou. `auth` y `tickets` **no escriben ni envían correos**: le piden a
`mail` que envíe, con los datos dentro, y `mail` pone el texto.

| Quién pide | Qué pide |
| --- | --- |
| Módulo `tickets` | Los **ocho avisos** de ticket (`docs/propósito-y-alcance.md`) |
| Módulo `auth` | Los **tres correos de cuenta**: alta, recuperación y aviso de cambio (`docs/modules/auth.md`) |

**Los textos no están en el código**: son **plantillas editables** que viven en la base de datos, y
un administrador las cambia desde la pantalla de Configuración, con vista previa y una prueba a su
propio correo. El detalle está en `docs/modules/mail.md`.

**El correo saliente vive en la base, no en variables de entorno**: los campos son
`installation_settings.smtp_*` (`smtp_host`, `smtp_port`, `smtp_secure`, `smtp_user`,
`smtp_password`, `smtp_from_name`, `smtp_from_email`), se piden en la vista de primer arranque y se
cambian desde Configuración. Las variables `SMTP_*` del entorno **se retiraron**.

Se llaman `smtp_*` y no `SYSTEM_SMTP_*` como en Calibyou: allí el prefijo distingue el correo del
sistema del de cada laboratorio, y aquí sólo hay uno.

**En desarrollo apuntan al buzón de pruebas** de `dev.yml`: el archivo de datos de ejemplo
(`backend/migrations/v1.0.0_dev.sql`) deja puestos `smtp_host='mail'` y `smtp_port='1025'`, así que
todo lo que sale se lee en `http://127.0.0.1:11004` y **nada sale a internet** (sección 10). Sin él, el
alta de una cuenta local no se podría completar en desarrollo. Si falta la configuración, el envío falla con un error claro y
**no se intenta por otra vía**: no hay correo «de reserva».

**Un fallo de correo no tumba la acción que lo provocó.** Si el ticket se crea y el aviso no sale,
el ticket existe: la acción ya está hecha y lo que falla es una consecuencia. El fallo se registra
con `go-logs` y la petición responde bien. Lo contrario —deshacer un ticket porque el correo no
salió— deja al usuario mirando un error por algo que sí ocurrió.

**El envío no bloquea la petición**: va en segundo plano con su tiempo límite, para que crear un
ticket no dependa de lo que tarde el servidor de correo. El precio es que el fallo se conoce por el
log y no por la pantalla, y es un precio razonable para un aviso.

**Los correos llevan lo justo**: número y asunto del ticket, y el enlace cuando toca. Nunca
contraseñas, ni tokens, ni datos personales de más.

**Los enlaces se construyen con la dirección pública de la instalación**, que se configura en la
pantalla de Configuración: un host escrito a mano en una plantilla no puede servir en desarrollo y en
producción a la vez. `PUBLIC_APP_URL` queda **como respaldo** para una instalación que ya la tuviera
puesta (`docs/modules/settings.md`, decisión 14).

## 10. Contenedores y ejecución

Todo corre en contenedores, con la misma forma que Calibyou: **un contenedor para el frontend,
otro para el backend y otro para la base de datos**, y nginx en la máquina como puerta de
entrada. No hace falta Go ni Node instalados en la máquina: sólo Docker y nginx.

### Estructura

```text
.
├── dev.yml                          # compose de desarrollo (los servicios, con el directorio de
│                                    # pruebas y las de interfaz detrás de un perfil)
├── prod.yml                         # compose de producción (imágenes construidas)
├── config
│   ├── dockerfiles                  # una imagen por servicio y por entorno
│   ├── seed                         # los archivos de los adjuntos del seeder de desarrollo
│   ├── env
│   │   ├── dev.env
│   │   ├── prod.env.example         # plantilla; prod.env NO se versiona
│   │   └── (prod.env)               # NO se versiona
│   ├── keycloak
│   │   └── realm-catalina-support.json  # el reino de pruebas, con su cliente y tres personas
│   ├── ldap
│   │   └── 01-personas.ldif         # las tres personas del directorio de pruebas
│   └── nginx
│       ├── catalina-support-dev.conf    # vhost de desarrollo
│       ├── catalina-support-prod.conf   # vhost de producción
│       └── frontend.prod.conf           # nginx del contenedor de frontend
├── backend
├── frontend
├── scripts                          # lo que se ejecuta en el servidor: las copias de la base
├── _logs                            # logs de go-logs (la carpeta se versiona, los .log no)
├── _files                           # adjuntos, por año y número de ticket (se versiona la carpeta,
│                                    # no los archivos)
└── docs
```

### Servicios de desarrollo (`dev.yml`)

| Servicio | Imagen base | Comando | Puerto |
| --- | --- | --- | --- |
| `frontend` | `node:24-alpine` | `npm start -- --host 0.0.0.0 --port 11001` | `11001` |
| `backend` | `golang:1.27-alpine` + air | `air -c .air.toml` | `11002` |
| `database` | `postgres:18-alpine` | `postgres` (`PGPORT=11003`) | `11003` |
| `mail` | `axllent/mailpit:latest` | `mailpit` | `11004` (su web; el SMTP es el `1025` interno) |
| `ldap` | `osixia/openldap:1.5.0` | El directorio, con las personas de `config/ldap/01-personas.ldif` sembradas al arrancar | `11005` (LDAP en claro: **es de desarrollo**). **No se levanta con `up -d`** (`profiles: ["auth"]`) |
| `keycloak` | `quay.io/keycloak/keycloak:26.0` | `start-dev --import-realm`, con el reino de `config/keycloak/realm-catalina-support.json` | `11006` (HTTP: **es de desarrollo**). **No se levanta con `up -d`** (`profiles: ["auth"]`) |
| `ai` (**en `ai.yml`**, no en `dev.yml`) | llama.cpp con servidor HTTP | Sirve el modelo **Qwen2.5-1.5B-Instruct Q4_K_M**, que se descarga una vez al volumen la primera vez | **Ninguno**: el motor no se publica en la máquina, sólo se habla por la red `catalina-support-ai`. Lo comparten **desarrollo y producción** (`docs/modules/ai.md`, decisión 1) |
| `e2e` | `mcr.microsoft.com/playwright:v1.63.0-noble` | `npx playwright test` | Ninguno: abre un navegador contra la dirección de desarrollo. **No se levanta con `up -d`** (`profiles: ["e2e"]`) |

**El motor de IA no está en la tabla de arriba a propósito**: vive en `ai.yml`, **compartido por los dos
entornos** —no caben dos modelos en esta máquina— y los dos backends entran en su red
(`catalina-support-ai`). Su porqué, sus argumentos y su memoria están en `docs/modules/ai.md`, sección
2, y cómo se levanta en `docs/ambientes.md`, sección 3.4.

Detalles que importan:

- **Código montado como volumen** (`./frontend:/app`, `./backend:/app`): se edita en la máquina y
  el contenedor recarga (Angular con su servidor de desarrollo, Go con `air`). No se reconstruye
  la imagen para cada cambio.
- **Caches en volúmenes con nombre** (`node_modules` del frontend, módulos y caché de compilación
  de Go): ni se pierden al recrear el contenedor ni se mezclan con archivos de la máquina.
- **`_logs` montado** en el backend (`./_logs:/logs`) para que los archivos de `go-logs` queden a
  la vista en la máquina.
- **`_files` montado** en el backend (`./_files:/files`, variable `FILES_PATH`), con **una carpeta
  por ticket** (`_files/2026/CS-2026-0042/`) y la de la marca aparte (`_files/brand/`), para los adjuntos de
  los tickets: son archivos, no filas, así que sobreviven a los reinicios y se copian desde la
  máquina como cualquier carpeta.
- **El backend espera a la base de datos**: `depends_on` con `condition: service_healthy` y un
  `healthcheck` de `pg_isready`, para que no arranque contra una base que aún no acepta
  conexiones.
- **Puertos propios**: se usan `11001`-`11006`, distintos de los de Calibyou (`10001`-`10004`)
  para que los dos proyectos puedan estar levantados a la vez. Se publican **sin dirección de
  escucha**, tal como están en `dev.yml` y en `prod.yml`.
- **`ldap` es el directorio de pruebas**: OpenLDAP con **tres personas** en un LDIF del repositorio
  (`config/ldap/01-personas.ldif`), una por cada caso del camino de AD. **No guarda nada en un
  volumen**: se siembra del repositorio al arrancar, así que siempre está como dice el repositorio, y
  **no se para y se arranca** —sus scripts de arranque no son idempotentes y muere—: se levanta otra
  vez recreándolo (`docs/modules/auth.md`, sección 11). Y **no es Active Directory**: se habla con él
  igual, así que el código se prueba de verdad, pero lo específico de AD no queda cubierto.
- **`keycloak` es el Keycloak de pruebas**: el reino vive **en el repositorio**
  (`config/keycloak/realm-catalina-support.json`), con su cliente confidencial y tres personas, y se
  importa al arrancar: recrear el contenedor lo deja como está escrito. **El navegador llega a él por
  el mismo dominio que a la aplicación y por un camino** (`/sso/`, un `location` del vhost de
  desarrollo): así el emisor de OIDC es uno solo para el navegador y para el backend, y no hay dos
  direcciones que puedan discrepar (`docs/modules/auth.md`, sección 11).
- **`mail` es el buzón de pruebas** (Mailpit), decidido por el responsable el 2026-09-23: **sin él no
  se puede completar el alta de una cuenta local en desarrollo**, porque la contraseña se establece
  desde el enlace del correo. Es **un solo buzón para todo**: acepta cualquier destinatario y de
  cualquier dominio, no tiene cuentas ni autenticación que configurar, y **no reenvía nada a
  internet**, así que en desarrollo ningún correo sale de la máquina. Los correos se leen en
  `http://127.0.0.1:11004`, con su HTML y su versión de texto, y se filtran por destinatario
  (`to:`), por asunto (`subject:`) o por texto, que es lo que hace de «bandeja por cuenta». Guarda los
  últimos 200 mensajes. En desarrollo lo deja apuntado el archivo de datos de ejemplo
  (`smtp_host='mail'`, `smtp_port='1025'`); en producción se pone el servidor de correo de verdad
  desde Configuración.
- **Variables en `config/env/dev.env`**, compartido por los servicios, con los nombres de
  `POSTGRES_USER`, `POSTGRES_PASSWORD` y `POSTGRES_DB` para que la imagen oficial de PostgreSQL
  los tome sola. `config/env/prod.env` **no se versiona** (sólo un `prod.env.example`).

### nginx

nginx es la puerta de entrada en los dos entornos, con **un vhost por entorno**, y se instala
copiando el archivo del repositorio (no hay enlaces simbólicos, igual que en Calibyou):

```bash
sudo cp config/nginx/catalina-support-dev.conf  /etc/nginx/conf.d/
sudo cp config/nginx/catalina-support-prod.conf /etc/nginx/conf.d/
sudo nginx -t && sudo systemctl reload nginx
```

| Entorno | Dominio | Vhost | Contenedores detrás |
| --- | --- | --- | --- |
| Desarrollo | `dev-catalina-support.calibyou.com` | `catalina-support-dev.conf` | frontend `11001`, backend `11002` |
| Producción | `catalina-support.calibyou.com` | `catalina-support-prod.conf` | frontend `21001`, backend `21002` |

- El frontend llama a `/api/**` **en relativo** y nunca a un host: el mismo código sirve en
  desarrollo y en producción sin cambios y sin CORS.
- **El prefijo `/api` no se reescribe**: las rutas del backend ya lo llevan
  (`/api/<módulo>/...`). Calibyou sí lo reescribe porque su backend las expone sin prefijo; aquí
  no, y esa diferencia está comentada en los dos vhosts para que nadie la "arregle".
- **TLS** con Let's Encrypt y el plugin `webroot` (`/var/www/certbot`), un certificado por
  dominio, renovado por la tarea de certbot que ya trae el paquete (`certbot.timer` y
  `/etc/cron.d/certbot`). Los dos vhosts sirven `/.well-known/acme-challenge/` en el puerto 80
  **antes** de la redirección a HTTPS, así que la renovación no necesita tocar nada.
- **`client_max_body_size 30m;`** en los dos vhosts. Sin esta línea nginx usa **1 MB** por defecto y
  cualquier adjunto de más de un mega se rechaza con un 413 **antes de llegar al backend**: el
  límite de 25 MB de los adjuntos (`docs/modules/tickets.md`) no se podría cumplir. Los 30 dejan margen.
- **Zona `limit_req` para el login**, como en Calibyou: es el único endpoint público que recibe
  credenciales, y sin límite se le puede probar contraseñas a gusto. Sólo esa ruta, porque el resto
  del tráfico no lo necesita.
- **El vhost de desarrollo NO incluye** `snippets/common-security.conf`, por el mismo motivo que
  el de Calibyou: ese snippet devuelve 444 a las rutas con un segmento que empieza por punto y
  aplica `limit_req`/`limit_conn`, y eso deja en blanco el servidor de desarrollo de Angular
  (que por dentro es Vite). El de producción **sí** lo incluye.
- La CSP de desarrollo es permisiva (`script-src 'self' 'unsafe-inline'`, `connect-src` con
  `ws:`) porque el servidor de desarrollo inyecta un script y abre un websocket para el recambio
  en caliente; la de producción es estricta (`script-src 'self'`).
- El servidor de desarrollo de Angular **rechaza con 403** cualquier `Host` que no conozca: por
  eso `dev-catalina-support.calibyou.com` está declarado en `serve.options.allowedHosts` de
  `angular.json`. Sin esa línea, la página responde 403 aunque nginx esté bien.
- La zona `limit_req` del login **ya está puesta**, como en Calibyou, porque `auth` está hecho y es
  el único endpoint público que recibe credenciales (unas líneas más arriba).

### Producción (`prod.yml`)

- Imágenes **construidas** con Dockerfiles de dos etapas (compilar y luego ejecutar), no con el
  código montado:
  - backend: `golang:1.27-alpine` compila un binario estático y la imagen final es `alpine` con
    el binario (sin Go, sin `air`, sin código fuente).
  - frontend: `node:24-alpine` compila con `npm ci && npm run build` y el resultado estático lo
    sirve nginx dentro del contenedor `frontend` (puerto 21001).
- El vhost de producción ya está instalado y con certificado. Mientras dure el aplazamiento el
  dominio responde **503 con un aviso** («esta versión todavía está en desarrollo») en lugar del 502
  que da nginx cuando no hay nada detrás:
  un 502 parece una web rota. La respuesta conserva las cabeceras de seguridad y la CSP estricta
  porque el bloque no declara `add_header` propio (si lo hiciera, nginx dejaría de heredar las
  del servidor). `/api/` sigue apuntando al backend que no existe, así que devuelve 502: nadie lo
  consume durante la pausa y así el día del despliegue sólo hay que restaurar un bloque, que está
  comentado en el propio vhost.
- **El despliegue a producción se hizo el 2026-09-25** con `prod.yml`, sus Dockerfiles y
  `frontend.prod.conf`, y quedó **verificado por dentro**; hoy queda **aparcado** hasta que el
  producto esté terminado (decisión del responsable, 2026-09-30). El runbook del despliegue está
  en `docs/ambientes.md`, sección 4. Lo que no se toca es la parte de nginx y TLS de la sección
  anterior: los certificados se renuevan solos y el vhost ya está en su sitio.

### Comandos

Todos los comandos se ejecutan **en la máquina**, contra los contenedores: no hay Go ni Node
instalados. Están también en `AGENTS.md`.

```bash
docker compose -f dev.yml up -d                 # levantar los servicios de desarrollo
docker compose -f dev.yml ps                    # ver el estado
docker compose -f dev.yml logs -f backend        # seguir los logs (salida de go-logs)
docker compose -f dev.yml down                   # parar (los volúmenes se conservan)

docker compose -f dev.yml exec backend go test ./...      # pruebas del backend
docker compose -f dev.yml exec backend go vet ./...       # análisis estático
docker compose -f dev.yml exec backend go mod tidy        # tras añadir dependencias

docker compose -f dev.yml exec frontend npm test -- --watch=false
docker compose -f dev.yml exec frontend npm run build

docker compose -f dev.yml exec database psql -U catalina_support -d catalina_support -p 11003
```

En desarrollo se entra por **https://dev-catalina-support.calibyou.com** (nginx → contenedores).
Los puertos 11001 y 11002 siguen publicados para depurar sin pasar por nginx.

## 11. Entorno de desarrollo

Requisitos en la máquina de desarrollo (comprobados el 2026-09-22):

| Herramienta | Estado | ¿Obligatoria? |
| --- | --- | --- |
| Docker | 29.5.2 instalado | **Sí**: todo corre en contenedores |
| Docker Compose | incluido en Docker | **Sí** |
| nginx en la máquina | 1.24 instalado, **con los dos vhosts instalados y certificados** | **Sí**, como puerta de entrada |
| Node.js | v24.19.0 instalado | No: el frontend compila en su contenedor (v24.21.0 en la imagen) |
| Go | **no instalado** | No: el backend compila y corre en su contenedor (1.27.1 en la imagen) |
| Cliente `psql` | 18.4 instalado | No: para consultar se entra al contenedor de base de datos |

Que Go y Node no hagan falta en la máquina es deliberado: la máquina sólo necesita Docker y
nginx, y las versiones quedan fijadas en las imágenes (Go 1.27, Node 24), no en lo que cada
máquina tenga instalado.

## 12. Decidido, pendiente y fuera de alcance

**Decidido** (lo pidió el responsable y este documento lo recoge):

- Angular en su última versión para el frontend.
- PostgreSQL en su última versión como base de datos.
- Backend Go con `net/http`, GORM y `air` en desarrollo.
- `github.com/juanky-estevez/go-logs` para los logs del backend.
- Organización modular, con el módulo del frontend consumiendo los endpoints de su módulo
  homónimo del backend.
- **Todo en contenedores**: uno para el frontend, otro para el backend y otro para la base de
  datos, con la estructura de Calibyou (`dev.yml` / `prod.yml`, `config/{dockerfiles,env,nginx}`).
- **Node 24 LTS** y **sin Go ni Node instalados en la máquina**.
- **Dominios**: `dev-catalina-support.calibyou.com` en desarrollo y
  `catalina-support.calibyou.com` en producción, los dos con TLS de Let's Encrypt y servidos
  por nginx en la máquina (decidido por el responsable el 2026-09-22).
- **El despliegue a producción se aplaza hasta cerrar la versión 1.0.0** (decidido por el
  responsable el 2026-09-22). Durante el desarrollo sólo se trabaja contra el entorno de
  desarrollo.

**Lo que estaba aquí y ya se decidió** (cada uno en su documento, que es donde vive la decisión):

| Qué | Dónde se decidió |
| --- | --- |
| Estilos, librería de componentes y pantallas | `docs/interfaz-y-experiencia.md` (Tailwind v4 y veinte componentes propios) |
| Autenticación, sesión, roles y permisos | `docs/usuarios-y-permisos.md` y `docs/modules/auth.md` |
| Modelo de datos y estados del ticket | `docs/modules/tickets.md` |
| Lista cerrada de módulos y sus endpoints | `docs/modules/tickets.md`: seis módulos, en el orden `mail` → `auth` → `users` → `tickets`, y `settings` |
| Notificaciones y correo | `docs/modules/mail.md` y, para qué se manda en cada momento, `docs/flujos.md` |
| Runbook de despliegue, migraciones y copias | `docs/ambientes.md` |
| Qué se prueba y con qué | `docs/ambientes.md`, sección 9: tres capas (`go test`, `npm test` y Playwright) |
**Pendiente de decidir** (queda poco, y ninguna de las dos bloquea la 1.0.0):

- **Logs del frontend**: no hay paquete decidido y el hermano es de NestJS, no sirve en navegador.
  Hoy el frontend no registra nada, y hasta que se decida, así se queda.
- **Una prueba que falle si un módulo del frontend llama a un prefijo ajeno**: la regla de la
  sección 4 se sostiene hoy en la revisión, y **ningún documento ha decidido todavía cómo se
  automatiza**. No es un olvido de esta sección: es que no está decidido en ningún sitio.

**Fuera de alcance de esta versión**: microservicios, colas, caché distribuida, SSR,
multi-tenant y aplicación móvil.

## 13. Estado de implementación

Lo que existe hoy en el repositorio y lo que se ha comprobado de verdad. La última pasada es del
**2026-09-30**: **la pantalla de Configuración entera** —con el motor de IA, la región horaria, la
dirección pública y el correo saliente—, **la vista de primer arranque** y **las rutas relativas de
producción**.

### Existe y está verificado

| Elemento | Comprobación |
| --- | --- |
| `dev.yml` con sus servicios | Los contenedores levantan y quedan sanos (`database` con `healthy`), y el del directorio de pruebas también (`docker compose -f dev.yml --profile auth up -d`) |
| Backend Go en contenedor con `air` | Compila, arranca, conecta con PostgreSQL y escucha en 11002 |
| Recarga en caliente de `air` | Al cambiar un archivo reconstruye y reinicia, con cierre ordenado del proceso anterior |
| `GET /api/health` | `200 {"database":"ok","status":"ok"}`, y 503 si la base no responde |
| `go-logs` | Escribe `_logs/catalina-support_dev_20260922.log` con `TIMEZONE` aplicado (`UTC-5`) |
| `go mod tidy` | Genera `go.sum`; el `go.mod` sólo lleva las dependencias reales |
| Frontend Angular 22 | El servidor de desarrollo sirve `index.html` y el bundle en 11001 |
| El armazón de la sesión en el navegador | Las seis rutas se sirven por nginx (200) y el CSS de Tailwind llega compilado. Y **probado en un navegador de verdad** con Playwright, en PC y en móvil: se entra con la cuenta de fábrica, se la reconoce, se cambia de pantalla y se sale |
| Pruebas de interfaz con Playwright | `docker compose -f dev.yml --profile auth run --rm e2e` → **214 casos: 196 en verde, 18 saltados y 0 rojos** (los saltados son las herramientas de diagnóstico, lo que es de un dispositivo concreto y **los de los caminos de directorio que se prueban una sola vez**, porque no dependen del ancho, más alguna prueba que necesita un servicio que no esté levantado): la aplicación abre, el CSS se aplica, los campos tienen nombre accesible, el idioma, los ocho temas con su contraste medido, entrar y salir, el enlace del correo leído del buzón, el armazón con su menú, los permisos del menú, la pantalla de Configuración —**con el método de entrada y las dos pruebas de conexión**—, **las tres pantallas de usuarios** —la lista con sus filtros, el alta con su correo, desactivar y reactivar, la edición en línea de Soporte, el perfil propio y que nadie se desactive a sí mismo—, **las de tickets** —el recorrido entero de un ticket con su adjunto, Soporte preguntando y escalando, Desarrollo devolviendo, el aspecto medido y el Administrador leyendo sin botones—, **el camino de AD** —una persona del directorio **entra sin que nadie le dé de alta nada** y su cuenta aparece con origen `ad`, quien ya es del directorio vuelve a entrar, la contraseña equivocada la rechaza el directorio, y una cuenta local **se vincula** al entrar por su camino y su contraseña local deja de servir— y **el de Keycloak**: el botón está en la pantalla de entrada y lleva a Keycloak, una persona entra por el reino **sin que nadie le dé de alta nada**, quien ya es de Keycloak vuelve a entrar, una cuenta local **se vincula** al entrar por allí, y **el fragmento con el token se borra de la dirección** en cuanto se usa. Y **las tres acciones del directorio en `users`**: la ficha de una cuenta de AD desactivada ofrece reactivarla y la reactivación pregunta al directorio, y la de una cuenta de Keycloak no ofrece el botón y cuenta que vuelve sola al entrar. Y **las listas de tickets** (2026-09-26): que «Mis tickets» sea **lo mío** —el que no lo tiene asignado no lo ve, y el que lo tiene sí—, que las dos listas del «todo» enseñen lo que hay, que **la reasignación se haga dentro del ticket** y mueva el ticket de una bandeja a la otra (un técnico se lo pasa a otro, que es el ejemplo del responsable), que el chip de tipo lleve a los internos, que desde las listas del «todo» no se cree un ticket, y que **el desplegable del idioma mida lo mismo que el del tema** (medido). Y **los adjuntos con tope y con visor** (2026-09-26): que una captura de **900 × 700** se pinte **dentro de 480 × 360 sin deformarse** —se mide, y con la captura de las otras pruebas, que mide justo 480 × 300, la comprobación pasaría sin que hubiera tope—, que al pulsarla se abra el visor con la imagen entera, que **el vídeo se vea como miniatura sin controles** y que al pulsarlo se abra el visor **con su reproductor y Descargar**, y que **un `.sql` se adjunte, se guarde y se descargue** mientras el `.svg` **se sigue rechazando**. Y **la vista de primer arranque** (el candado: una instalación ya terminada no la enseña y `/setup` lleva a la entrada) y **la tarjeta del motor de IA en Configuración, con su prueba de conexión** |
| Pruebas del frontend | `npm test` → **215 pruebas en verde, en 20 ficheros**: el idioma de arranque (incluidas las variantes como `es-MX` y el caso de un idioma que no es ninguno de los dos), el servicio de sesión, el interceptor (cabecera, 401 con y sin sesión, servidor caído), **el tema** (los ocho, el sistema en vivo, lo elegido manda), el armazón, las entradas del menú por papel, **el módulo `users`** —su servicio y sus etiquetas— y **el módulo `tickets`**: su servicio (cada acción a su ruta, la lista de responsables pedida a su propia API) y sus etiquetas (los estados del usuario sin jerga, lo que se previsualiza, las frases del historial). Y **los caminos de entrada**: que la sesión pregunte cuáles hay, que sin respuesta se quede con el local —y no ofrezca un botón que no puede comprobar— y que adopte el token que trae la vuelta de Keycloak. Y de las pantallas de usuarios, **cuáles se reactivan solas**: una cuenta de Keycloak desactivada no ofrece el botón y las demás sí. Y **la marca**: que el nombre de la instalación sea el configurado, que sin backend quede el de fábrica —y la pestaña no se quede sin nombre—, que **la versión del sistema se enseñe con su `v`** y que **sin versión no se enseñe ningún número** —inventarse uno sería peor que no decir ninguno—, y que el título de la pestaña cambie al guardarlo. Y **las listas de tickets**, desde el 2026-09-26: que cada papel tenga **sus entradas del menú** —Mis tickets para el usuario, y las dos del «todo» sólo para Soporte y Desarrollo—, que **crear un ticket no esté en el menú para nadie** y que «lo mío» viaje como `mine=1` y sólo cuando se pide. Y **el editor con adjuntos** (`shared/components/editor-con-adjuntos.spec.ts`, **24 pruebas**): los cinco botones de formato sobre lo seleccionado, **la lista cerrada de lo que se puede guardar** —el saneador del editor quita `src`, `class`, `style` y `on…`, y no deja etiquetas vacías— y **qué extensiones se admiten**: el texto y el código que entraron el 2026-09-26 —`sql`, `json`, `xml`, `yml`, `sh`, `py`, `htaccess`, `tar`…—, y que lo que no se admite sigue sin admitirse (`svg`, `exe`, `html`). Y **la configuración y el primer arranque**, desde el 2026-09-30: el servicio de `settings` y **la guarda de instalación** (sin sellar lleva a `/setup`; sellada, no) |
| Tailwind v4 y los temas | Instalado y compilando: el CSS servido lleva **las utilidades generadas** y las variables del tema, y **ningún componente escribe un color a mano**. Las fuentes se declaran a mano en `styles.css` (`@source './app'`), y hay un caso de Playwright que **falla si la hoja llega sin utilidades** |
| El tema | **Ocho temas** (dos de fábrica con `light-dark()` y el color institucional, y seis fijos con su paleta y su acento), elegidos con un atributo en `html` y el tema elegido con un atributo en `html`. **De fábrica sigue al sistema** —que es no haber elegido, y por eso **«automático» no se muestra ni se elige**—, el sistema manda en vivo mientras nadie haya elegido, la elección se recuerda en el navegador y se aplica **antes de arrancar** desde `main.ts` (no con un script incrustado: la CSP de producción no lo admite). Probado en un navegador **midiendo el color de fondo**, en claro, en oscuro y con la elección ganando al sistema |
| Build de producción del frontend | `npm run build` → `dist/catalina-support/browser`, que es la ruta que espera `prod.yml` |
| Vhost de desarrollo | `https://dev-catalina-support.calibyou.com` sirve la aplicación (200) y `GET /api/health` devuelve `{"database":"ok","status":"ok"}` **a través de nginx** |
| TLS de los dos dominios | Certificados de Let's Encrypt emitidos el 2026-09-22 (caducan el 2026-12-21), con renovación automática configurada por `webroot` |
| Cabeceras de seguridad y CSP | Presentes en la respuesta de desarrollo (`nosniff`, `SAMEORIGIN`, `Referrer-Policy` y CSP de desarrollo) |
| Migración `v1.0.0.sql` | Aplicada con `psql -v ON_ERROR_STOP=1` **dos veces seguidas** y sin error: crea **18 tablas** —`mail_templates` (con sus **22 plantillas** sembradas: once correos en dos idiomas), `users`, `password_tokens`, las **cuatro tablas de configuración de una fila** —`installation_settings` (con el nombre, el método de entrada, **la región horaria, la dirección pública, el motor de IA, el correo saliente `smtp_*` y el sello `installed_at`**), `directory_settings`, `keycloak_settings` y `ticket_settings`—, **las diez de los tickets** y `ai_insights`—, con sus índices y sus restricciones. Lleva además su **puesta al día** (`ADD COLUMN IF NOT EXISTS`, `DROP CONSTRAINT IF EXISTS` + `ADD CONSTRAINT`), que es lo que permite aplicarla sobre una base que ya existía, y **la conversión del texto plano a HTML** de los tickets y los comentarios que ya existían: **aplicada dos veces seguidas, las dos conversiones dan `UPDATE 0`** y un cuerpo con un adjunto dentro —lo que escribe el editor— **no se toca** (la condición se corrigió el 2026-09-26, porque con la anterior el segundo pase estropeaba lo que el editor escribe; `docs/ambientes.md`, sección 5) |
| Los datos de ejemplo, con su guion | `./scripts/dev-seed.sh` aplica el esquema, aplica `v1.0.0_dev.sql` y **copia los archivos de los adjuntos** a `_files/`, en la carpeta de su ticket. Deja **once cuentas** (`user1`…`user5`, `support1`…`support3`, `dev1`…`dev3`, todas con la contraseña de las pruebas) y **25 tickets** con su historia: **123 filas de historial y 27 comentarios**, siete con **ticket interno** (uno esperando a Desarrollo, otro resuelto, otro devuelto a Soporte), **cinco adjuntos que se descargan idénticos** a los del repositorio, y el contador de la numeración en 25. Comprobado entrando por la API con `user1` (ve sus cinco tickets) y con `support1` (la bandeja entera) |
| `shared/auth` | Firma y valida el token de sesión (HS256, 10 horas, `sub`), saca el `Bearer` de la cabecera y compara secretos en tiempo constante. **20 pruebas en verde**, incluidas la caducidad, otro secreto, el algoritmo `none`, los dos secretos vacíos y las cinco del `state` de OIDC |
| `shared/authz` | Comprueba el papel y responde 403 con `auth.forbidden`. **5 pruebas en verde**, incluido el caso de una ruta montada sin autenticación |
| Middleware de autenticación | Lee la cuenta **en cada petición**, deja la identidad en el contexto y responde 401 (`auth.session.invalid`/`auth.session.expired`) o 503 si la base no responde. **8 pruebas en verde** |
| `go test ./...` | Todo en verde, en **14 paquetes** y **209 pruebas**: **20 en `shared/auth`** —cinco del `state` de OIDC—, 5 en `shared/authz`, 8 en `shared/middleware`, 12 en `mail/services`, 6 en `mail/controllers`, **39 en `auth/services`** —doce del camino de AD, once del de Keycloak y cuatro del método que está puesto, con un directorio, un reino y un módulo de cuentas de mentira—, **13 en `users/services`**, 15 en `settings/services`, **2 en `settings/dtos`** —la versión que sale en la marca pública—, **6 en `tickets/controllers`**, **6 en `tickets/repositories`** y **46 en `tickets/services`** —incluidos los cuatro del saneador del cuerpo: lo que se admite tal cual, el texto plano que sale del HTML, el texto convertido a HTML y las extensiones de vídeo—, y el módulo `ai`: **6 en `ai/repositories`** y **28 en `ai/services`**, con `gofmt` limpio y `go vet` sin quejas |
| Los cuatro endpoints de `mail` | Probados de extremo a extremo contra desarrollo, con un token de fábrica firmado a mano (todavía no hay endpoint de entrada): sin cabecera, con un token con basura y con un token de otro secreto → **401 `auth.session.invalid`**; con token válido → **200** y las 22 plantillas; marcador inventado → **422 `mail.marker.unknown`**; asunto vacío → **422 `mail.subject.required`**; guardar sin el enlace → **200** con `missing: ["enlace"]`; restaurar → vuelve el texto de fábrica y `edited` pasa a `false`; prueba desde la cuenta de fábrica → **422 `mail.test.noEmail`**; clave o idioma que no existen → **404** |
| La cuenta se lee en cada petición | El middleware la resuelve con el cargador de `main.go`: la cuenta de fábrica tiene su identidad propia y cualquier otra se lee **de la tabla de cuentas en cada petición**, así que desactivar o cambiar un papel valen al instante |
| El log tras las pruebas | Ni una coincidencia de `Bearer`, `eyJ` ni la palabra `token`: los tokens no se registran |
| El camino local de `auth`, de extremo a extremo | Entrar con la cuenta de fábrica y con una cuenta local, `me`, `salir`, olvidar la contraseña, establecerla desde el enlace del correo y cambiarla desde dentro. **Probado con curl contra desarrollo**, con el enlace leído del buzón de pruebas |
| El camino de AD, de extremo a extremo | **Los cinco casos de la sección 5.4 de `docs/modules/auth.md`, comprobados uno a uno por curl** contra el OpenLDAP de `dev.yml` y con la fila que queda en la base: el **alta automática** (una persona del directorio sin cuenta entra y su cuenta nace con papel `usuario`, origen `ad`, su identificador externo y **sin contraseña**); el **vínculo** (una cuenta local con el correo de alguien del directorio entra con la contraseña de allí, pasa a `ad` **conservando su papel** y su contraseña local deja de servir); la **actualización** (los apellidos del directorio sustituyen a los de aquí); **el correo que cambia en el directorio** (cambié el `mail` de una persona con `ldapmodify` y entró **en la misma cuenta**, porque manda el identificador externo); la **reactivación** (desactivada aquí, el directorio la devuelve al entrar, con el aviso `users.directoryMayReturn` ya ejercitado de verdad); y el **directorio caído** (una cuenta de AD → 503 `auth.directory.unavailable`; un correo desconocido o una cuenta local → 401 `auth.invalidCredentials`, y una cuenta local con su contraseña buena entra igual) |
| Los enlaces de contraseña | Un enlace, un uso: el mismo token dos veces → 401 `auth.token.used`; un token inventado → 401; el de alta caduca en 24 horas y el de recuperación en 1, y lo dicen en el correo |
| El camino de Keycloak, por curl y en el navegador | **Por curl**, contra el Keycloak de desarrollo: `GET /api/auth/keycloak/start` → **302** a su pantalla de entrada con `client_id`, `redirect_uri`, `response_type`, `scope` y `state`; una vuelta con un `state` inventado → **302 a `/login#error=auth.oidc.state`**; y una vuelta con un `state` bueno y un código que no vale → **302 a `/login#error=auth.oidc.rejected`**, que es la prueba de que el canje se hace de verdad contra Keycloak. **En un navegador de verdad** (Playwright): el botón, la pantalla de Keycloak, la vuelta, la sesión y **el fragmento borrado**. Y **sin Keycloak configurado**: `methods` dice `keycloak: false` y su ruta responde **404**, así que esa instalación no ofrece el botón ni tiene el camino |
| Los adjuntos en los comentarios | **Tres puertas y una sola regla**: pegar, arrastrar y elegir con el botón, con la misma lista cerrada de extensiones y el mismo aviso. Comprobado en el navegador con **un evento de pegado de verdad**: el archivo aparece en la lista, se publica el comentario con sus archivos, y **un comentario sin texto** —cuerpo vacío en la base— se lee por sus adjuntos **sin dejar un párrafo vacío**. Y **el camino del fallo**, cortando la subida en el navegador: el aviso dice qué archivo falta, el comentario queda publicado **una sola vez** y el botón de reintentar lo sube al mismo comentario |
| Los adjuntos dentro del texto | El editor con formato: **pegar, arrastrar y el botón meten el archivo donde está el cursor**, la imagen **se ve mientras se escribe** y **después de guardar** con su tamaño, un PDF sale como **enlace que abre el visor** (más grande, con **Descargar** y **Abrir en una pestaña**) y un Word **descarga**. Comprobado además: **lo que se guarda es el texto con sus referencias y nada más** (ni `class`, ni `src`, ni `blob:`), **un adjunto sin nombrar sigue al final**, **la búsqueda no encuentra por una etiqueta**, y **editar** un comentario con una imagen **enseña la vista previa en el editor**. En el backend, `go test` cubre el saneador: seis formas de HTML prohibido dan **422 `tickets.body.notAllowed`** sin dejar fila |
| Los dos logos de fábrica | Que la aplicación enseñe **el que toca al tema que se está viendo**: con un tema claro carga `logo-catalina-support-light.png` y con uno oscuro el `dark`, comprobando que **la imagen carga de verdad** (`naturalWidth`) y que son dos archivos distintos. Los dos salen del repositorio (`frontend/public/`), con fondo transparente, y el caso quita primero cualquier logo propio y deja la instalación como estaba |
| La política de contraseñas y el enlace | **Mínimo 8 caracteres** (bajado de 12 el 2026-09-25), comprobado en el navegador: una contraseña corta se rechaza con su clave **y no gasta el enlace** —se vuelve a intentar en la misma pantalla y entra—, y ocho caracteres valen. El caso destapó que el enlace se gastaba antes de comprobar la contraseña, y está corregido |
| El despliegue a producción | **`scripts/prod-build.sh`** construye y publica los dos artefactos en `/root/prod/catalina-support` (rota el anterior **por copia**, verifica que el artefacto existe y corta si no, y escribe `BUILD_INFO` con el commit, la rama, si el árbol estaba sucio y la fecha) y levanta los contenedores con `prod.yml` reiniciando **sólo el backend**. Los tres contenedores de producción están levantados y sanos, y el `admin` entra con la contraseña de `config/env/prod.env`. **El dominio sigue respondiendo 503**: la 1.0.0 no está cerrada. **Este despliegue queda aparcado** hasta que el producto esté terminado (decisión del responsable, 2026-09-30) |
| La versión del sistema | Publicada en la marca pública (`GET /api/settings/brand` → `"version": "1.0.0"`) y leída en los dos sitios: **en la fila de salir del menú lateral, alineada a la derecha**, y **en el pie de la pantalla de entrada**. Comprobado en el navegador en PC y en móvil, contra lo que dice la API —no contra un número escrito en la prueba—, que es texto y no un enlace, que se lee sobre el fondo del menú, y que **plegado no se enseña**. Y que sin marca no se enseña ningún número: la pestaña no se queda sin nombre y el pie sin versión, en vez de inventarse uno |
| El nombre de la instalación, por curl y en el navegador | Cambiado por la API y **leído antes de entrar** (`GET /api/settings/brand`): un nombre de 61 caracteres → **422 `settings.name.tooLong`**, y uno en blanco → vuelve `Catalina Support`. Y en el navegador: el menú lateral, la pestaña y la pantalla de entrada lo enseñan, y al vaciar el campo vuelve el de fábrica |
| Las copias de la base | `scripts/backup-db.sh dev`, ejecutado de verdad: deja el volcado en `/root/backups/catalina-support/`, **comprueba que se puede leer** (`pg_restore --list`) y **se restauró en una base nueva**: 11 tablas y los mismos registros que la de verdad (tickets, cuentas, plantillas y adjuntos, contados uno a uno). Y la retención, probada con una copia de hace veinte días: se borra, y no se toca la de tres días ni la de hoy |
| Los adjuntos, en su carpeta, por curl | Un ticket creado de verdad y **su adjunto subido por la API**: el archivo aparece en `_files/2026/CS-2026-0001/<nombre generado>.png` y en la base queda esa **ruta relativa**; el adjunto que se sube **al interno** va a la suya, `_files/2026/INT-CS-2026-0001/`; y **los dos se descargan idénticos** al que se subió, comparados byte a byte (`cmp`) |
| Las tres acciones del directorio, por curl | Contra el directorio y el reino de verdad: dar de alta una cuenta `ad` a mano y cambiar el origen de una cuenta local a `ad` y a `keycloak` → **422 `users.origin.byDirectory`** (y el cambio a `local` sigue funcionando, **200**); reactivar una cuenta de **AD** desactivada → **200 y activa**, porque el directorio la encuentra; reactivar una cuenta **cuyo correo el directorio ya no conoce** → **422 `users.directory.notFound`**; y reactivar a mano una cuenta de **Keycloak** → **422 `users.directory.activatesItself`** |
| Las cuentas y sus permisos | Alta de cuenta local → 201 y correo; correo repetido → 409; correo mal escrito → 422; origen inventado → 422; alta con origen `ad` o `keycloak` → 422 `users.directory.notFound`; Soporte creando `soporte` → 403 `users.role.notAllowed`; un usuario cualquiera dando de alta → 403 `auth.forbidden` |
| Desactivar bloquea al instante | Con la cuenta desactivada en la base, **el token que ya tenía deja de valer** (401 `auth.session.invalid`) y al entrar se le dice que está desactivada (`auth.accountInactive`) |
| El buzón de pruebas | Todo el correo de desarrollo llega a él, venga la dirección de donde venga, y se filtra por destinatario o asunto. **Ni un correo sale a internet** |
| El log | Registra entradas correctas, intentos fallidos con su clave y su IP, y cada correo con su destinatario. **Ni una contraseña, ni un token**: comprobado buscando en los archivos de `_logs` |
| La marca de la instalación | El logo se sube, se sirve y se quita: un JPEG de 1254×1254 de verdad (203 KB) subido por la API, servido **idéntico** al original, con su sello de versión; un SVG servido con `Content-Security-Policy: sandbox`; un archivo que no es imagen → 422; uno de 1,6 MB → 413; y al quitarlo, la instalación vuelve al logo de fábrica |
| El nombre de la instalación | Configurado desde la pantalla de Configuración y **leído en los tres sitios**: el menú lateral, la pestaña del navegador y la pantalla de entrada —que se pinta **antes de entrar**, y por eso el nombre viaja en la marca pública—. Con sus límites: pasado de 60 caracteres → **422 `settings.name.tooLong`**, y **vacío devuelve el de fábrica**. Comprobado en un navegador de verdad |
| El color institucional | Probado con un color que no se lee (amarillo): el backend deriva el del tema claro oscurecido y deja el del oscuro, y el texto del botón sigue contrastando. Probado en el navegador |
| El logo en la pantalla de entrada | Playwright comprueba que **la imagen carga de verdad** (`naturalWidth`), que va en el recuadro con su borde, que mide entre 120 y 160 px de alto, que un logo propio sustituye al de fábrica y que al quitarlo se vuelve |
| El armazón con el menú lateral | Probado en PC y en móvil: las tres zonas con el logo arriba y los controles abajo, el plegado **que se recuerda** al recargar, el cajón del móvil que se abre y se cierra, y que **las pantallas de la sesión no llevan menú** |
| El menú por papel | Un usuario **no ve Configuración** en el menú, ve su propio papel, y si escribe la dirección a mano la guarda lo lleva a «sin permiso». Probado creando la cuenta de verdad y entrando con ella |
| La pantalla de Configuración | Se sube un logo **por el formulario**, se ve en la vista previa y en el menú sin recargar, y se vuelve al de fábrica. El color institucional tiene **vista previa pedida al backend** —un amarillo puro se enseña oscurecido para los temas claros— y se puede descartar |
| Método de autenticación, desde la pantalla | El **método de entrada** se cambia en Configuración y **vale sin reiniciar nada** (el camino de entrada lee la base en cada intento): con `ad` la pantalla de entrada enseña el aviso de que la contraseña es la de la organización y quita el enlace de recuperarla, y con `keycloak` no hay formulario y sólo queda el botón, más la **puerta de la cuenta de fábrica**. Comprobado en el navegador, y **la suite deja el método donde estaba**. **Los secretos no salen nunca**: `GET /api/settings` devuelve `passwordSet` y `secretSet` y ningún valor, y guardar con el campo vacío **conserva** el que había |
| Las dos pruebas de conexión | El botón «Probar la conexión» de las dos tarjetas, contra el directorio y el reino de verdad: contesta que sí cuando el servicio está y **con su clave cuando no** —probado apuntando el servidor a uno que no existe—. **No guarda nada**: al recargar, la configuración sigue como estaba. Se salta sola si el perfil `auth` no está levantado, y la suite entera corre sin él |
| La región horaria y la dirección pública | En **Configuración**, desde el 2026-09-29: **la zona horaria** se elige de una lista de zonas (IANA) con buscador —y se enseña la hora que es en ella y su desfase— y **la dirección pública** (esquema, host y puerto, `localhost` incluido), con el aviso de que **si no es https la sesión y la contraseña viajan sin cifrar**. **La zona decide cómo se leen todas las fechas**, en la interfaz y en los correos, y **las guardadas siguen en UTC**: cambiarla no mueve ningún ticket. **La dirección es la base de los enlaces de los correos y de la vuelta de Keycloak**; `PUBLIC_APP_URL` queda como respaldo (`docs/modules/settings.md`, decisiones 14 y 15) |
| El correo saliente, en la base | Las variables `SMTP_*` **se retiraron del entorno**: el SMTP vive en `installation_settings.smtp_*` y se configura en **el paso 4 de la vista de primer arranque** o, después, en **Configuración**. Comprobado en desarrollo: el archivo de datos de ejemplo deja puesto el buzón de pruebas (`smtp_host='mail'`, `smtp_port='1025'`) y el correo sale por él (`docs/primer-arranque.md`, sección 5) |
| El motor de IA, configurable desde la pantalla | **Su dirección y su modelo se configuran en Configuración**, con **su tarjeta y su botón de «Probar la conexión»** (prueba `<dirección>/health` con un tiempo corto y **no guarda nada**): el módulo `ai` los lee **en cada petición**, y `AI_URL`/`AI_MODEL` quedan **como respaldo**. La prueba se salta sola si el motor no está levantado (`docs/modules/settings.md`, decisión 16, sección 5.13) |
| La vista de primer arranque (`/setup`) y el sello | En una instalación **sin sellar** —`installation_settings.installed_at` nulo— la aplicación lleva a **`/setup`**, que pide **en cuatro pasos** la instalación, cómo se entra, dónde está (región y dirección) y el correo saliente, y al terminar **sella** la instalación: la vista no vuelve y su API contesta **409 `setup.alreadyInstalled`**. **Una instalación que ya estaba configurada queda sellada al actualizar**, y en desarrollo **el archivo de ejemplos sella**, para que el asistente no salga en cada arranque. Comprobado en el navegador (el candado: una instalación terminada lleva `/setup` a la entrada) y por la API a mano (los cuatro pasos y el 409). `docs/primer-arranque.md` |
| Los endpoints de `users` | Probados de extremo a extremo: la lista con filtros y búsqueda, la ficha, los cambios, el perfil propio, desactivar y reactivar. Y los límites: Soporte no cambia el papel ni el correo de otro (**403**) ni ve la ficha de nadie, nadie se desactiva a sí mismo (**403 `users.selfDeactivation`**), el estado no se cambia por el `PATCH` (**422**), el origen inventado (**422**) y pasar a directorio se rechaza mientras no exista su consulta (**422**) |
| El editor de los correos | La pantalla del módulo `mail`: los once correos a la izquierda, **los dos idiomas al lado** —apilados en móvil, medido por la prueba—, el cuerpo con sus botones de formato **envolviendo lo seleccionado**, los marcadores que **se insertan al pulsarlos**, la **vista previa que renderiza el backend** con los datos de ejemplo de la prueba, y los botones de guardar, volver al de fábrica y **enviarme una prueba**, que se lee en el buzón. Probado de extremo a extremo |
| El idioma de la instalación | Se configura en Configuración y **es el que se le pone a una cuenta nueva cuando quien la da de alta no elige otro**: probado con la instalación en inglés y un alta sin idioma |
| Las pantallas de tickets | La **bandeja** —una sola con tres nombres, tabla en PC y tarjetas en móvil, con chips de estado, búsqueda, paginación y el chip de tipo del Administrador—, el **alta** con arrastrar y soltar, y el **detalle**: una línea de tiempo con los comentarios y lo que hizo el sistema, la ficha con las acciones —a la derecha en PC y debajo en móvil, medido por la prueba—, **la vista doble** cuando hay interno y los adjuntos **con vista previa**. Probado en PC y en móvil, incluido que **el usuario no ve el interno**, que **Desarrollo no escribe en el principal** y que **al Administrador no se le ofrece ningún botón** |
| El reparto y el prefijo, desde Configuración | La pantalla de Configuración tiene ya el **prefijo** —con el aviso de que no cambia los números ya emitidos— y el **reparto y el aviso** de cada tipo de ticket, con la opción «al asignado» desactivada cuando no se reparte |
| Las tres pantallas de usuarios | La **lista** (tabla en PC y tarjetas en móvil, con los chips de papel, origen y estado, la búsqueda y la paginación), **la ficha** y **el perfil propio**, probados en PC y en móvil. Soporte **edita en línea** desde la lista y no tiene ficha de nadie; no se le ofrece desactivarse a sí mismo; la cuenta de fábrica no tiene perfil ni enlace a él, y su nombre en el menú no es un enlace |
| El menú llega hasta abajo | El cajón **mide la zona visible y se desplaza por dentro**: en un móvil con barra del navegador (839 px visibles de 890) el botón de salir quedaba fuera de la pantalla y no se podía pulsar. Lo encontró Playwright, y con él se corrigió además que en PC la barra se estirara a todo el contenido |
| Un solo `h1` por pantalla | Las tarjetas de las pantallas de producto titulan con `h2`: la de usuarios tenía tres `h1` y el perfil otros tantos. Lo dijo la prueba al encontrar dos encabezados con el mismo nombre donde esperaba uno |

| Las diez tablas de `tickets` | Creadas en `v1.0.0.sql` y aplicadas **dos veces seguidas** sin error: los dos tipos de ticket, la conversación, los adjuntos, el historial y el contador de la numeración, con sus restricciones (el interno no puede estar `escalado`, exactamente un destino en lo que cuelga de un ticket, el motivo del escalado no puede estar vacío) |
| Las extensiones del texto y el código | Comprobado de punta a punta por la pantalla: un `.sql` **se coge** —sin el aviso de rechazo—, **se guarda** su referencia en el comentario, **queda entre los adjuntos del ticket** y **al pulsarlo se descarga** con su nombre; y un `.svg` **se sigue rechazando** antes de subirlo. Las **cuarenta y dos** extensiones nuevas son las mismas en los dos sitios —el editor y el backend—, y el editor lo comprueba por su lado con sus pruebas de unidad |
| El módulo `ai`, con el motor de verdad | **Comprobado de punta a punta contra el contenedor**: al crear un ticket, `Pedir` vuelve en milisegundos y los dos campos quedan **`pendiente`**; entre 20 y 40 segundos después están **`listo`**, en español y en inglés, y con contenido razonable (`El usuario está experimentando problemas al intentar iniciar sesión…` / `The user is experiencing issues trying to log in…`). En la pantalla: la lista enseña el motivo y la última acción, la ficha los enseña enteros con su botón, y **el interno no pide motivo** —el suyo es el del escalado—. Y **sin motor configurado** los campos quedan `sin_motor` y la aplicación funciona entera, que es la condición de la decisión 2 |
| El motor de IA, medido | El contenedor `catalina_support_ai` (llama.cpp `server-b11206` + Qwen2.5-1.5B-Instruct Q4_K_M) gasta **~950 MB en reposo y 1,09 GB** con una entrada de 2 034 piezas, dentro de su techo de 1500m y 2 CPU, **sin publicar ningún puerto**. **Tarda entre 12 y 24 segundos por campo** y unos 95 s en un ticket de 6 000 caracteres; la segunda llamada del mismo ticket es ~7 veces más rápida (caché de prefijo). Su comprobación de salud contesta en 325 ms mientras está generando |
| El JSON del modelo, sin creérselo | **Con `response_format` y la exigencia escrita al final del mensaje, 4 de 4 respuestas válidas** con un ticket de 6 000 caracteres; con sólo una de las dos vías, el modelo envuelve la respuesta en un bloque de código, se inventa las claves o contesta en un solo idioma. El módulo rescata el primer objeto JSON y reintenta; cuando no vale, el log dice **la forma** de lo que llegó (claves y longitudes) y nunca su contenido |
| Las categorías y las etiquetas | **Comprobado por las tres capas**: la migración crea «General» y se la pone a todo lo que existe —simulando una base vieja, el ticket sin categoría acaba en «General» y la columna queda `NOT NULL`—; el alta **sin categoría se rechaza** (`tickets.category.required`) y **un usuario no puede crear ni retirar** categorías; las etiquetas sucias se guardan normalizadas (`["Red Wifi","  Licencias  ","red-wifi","Ñoño_Ütil"]` → `['red-wifi','licencias','nonoutil']`); un ticket con 40 caracteres de etiqueta se rechaza; retirar la última activa se rechaza; `?category=` y `?tag=` filtran; y **buscar «red» encuentra la categoría «Red» y la etiqueta `red`**. En la pantalla: el alta pide la categoría, las fichas se normalizan al escribirlas, los chips salen en el renglón, hay filtros de las dos y el catálogo lo mantiene quien toca |
| Las menciones y los observadores | **Comprobado por las tres capas**: el saneador admite `<span data-mencion="12">` y rechaza lo que no es un identificador; al comentar con una mención se crea el observador, el ticket **entra en el «Observo» de esa persona y no en sus «Asignados»**, quitar al observador deja su entrada en el historial y **no vuelve** al reenviar el mismo comentario, un usuario que lo intenta recibe `tickets.mention.notAllowed`, y el aviso `ticket.mentioned` llega al buzón de pruebas. En la pantalla: el buscador de personas, el nombre resaltado dentro del comentario, la línea «Observadores» de la ficha con su aspa y el chip de vista |
| El botón de copiar el número | De sólo icono, y al pulsarlo un visto verde dos segundos. La prueba **mide el color pintado** y lo compara con el `--exito` del tema, no con el nombre de la clase —y ahí salió que el botón tiene transición, así que la primera lectura coge el color a medio camino— |
| El filtro «lo mío» | `GET /api/tickets?mine=1` devuelve **exactamente lo que dice la consulta**: comprobado contra la base con `support1` (**16 tickets**: 9 principales asignados, abiertos o comentados por él y 7 internos escalados o comentados), con `dev1` (**3 internos**) y con `user1` (**5**, todos suyos: el usuario sigue viendo sólo lo suyo con o sin `mine`) |
| El backend de `tickets`, de extremo a extremo | **78 comprobaciones contra desarrollo**, leyendo los correos del buzón de pruebas: el alta con su **número** (`CS-2026-0001`, y el 9999 que crece a cinco dígitos), el **reparto por turnos** que alterna entre los técnicos activos, el triaje, el **escalado** con su motivo obligatorio, las reglas de sincronización (el interno en espera y el interno resuelto devuelven el principal a `en progreso`), la **devolución** al cerrar el interno sin resolverlo, el **re-escalado** con el mismo interno, el cierre, la reapertura con las fechas limpias, los comentarios (editar y borrar vaciando el texto), los adjuntos (subir, descargar, la extensión fuera de la lista, el `svg` y los 25 MB) y **los permisos de los cuatro papeles**, incluido que el usuario no alcanza un ticket interno ni por su número ni por la ruta de un adjunto |
| Los siete avisos de `tickets` | Salen cuando la transición que los provoca lo dice, y se han leído en el buzón: al asignado, al usuario (en espera, resuelto, cerrado), a Desarrollo (escalado, también al re-escalar) y a Soporte (Desarrollo necesita algo, y el ticket ha vuelto a su bandeja) |
| Los sitios que ya existían | `calibyou.com`, `dev.calibyou.com` y `dev-ramona-fit.calibyou.com` siguen respondiendo 200 después de instalar los vhosts |

### Existe pero NO está verificado

- **El camino público de producción, aparcado**: el contenedor de frontend está levantado con su nginx,
  pero **el dominio sigue respondiendo 503 a propósito** y **el despliegue no se retoma hasta que el
  producto esté terminado** (decisión del responsable, 2026-09-30), así que el vhost público y
  `config/nginx/frontend.prod.conf` sirviendo la aplicación no se han probado.

### Todavía no existe

- **Los caminos de AD y de Keycloak ya no están aquí**: los dos existen y están verificados, cada uno
  con su servicio de pruebas detrás del perfil `auth` de `dev.yml` (OpenLDAP y Keycloak), con sus
  personas en el repositorio (`config/ldap/`, `config/keycloak/`) y con pruebas de interfaz.
  **`docs/modules/auth.md` no tiene nada pendiente.**
- **`scripts/prod-build.sh` y el script de copias ya están** (`scripts/backup-db.sh`, en el `cron` de
  esta máquina), y lo que no está probado de este último es su rama de producción
  (`docs/ambientes.md`, sección 6).
- **Lo que falta**, sólo flecos técnicos:
  - **Los botones de prueba del asistente**: la API `/api/setup/**` **no tiene endpoints de prueba** —y
    todavía no hay sesión con la que probar—, así que el asistente **guarda y valida** pero no prueba el
    directorio, el reino, el correo ni el motor; eso se hace después desde Configuración
    (`docs/primer-arranque.md`).
  - **Un repaso de formato en el frontend**, pendiente.
- **Aparcado a propósito** (decisión del responsable, 2026-09-30): **el despliegue a producción y abrir el
  dominio no se retoman hasta que el producto esté terminado**. El despliegue del 2026-09-25 está hecho y
  verificado por dentro, y el dominio sigue en 503 a propósito (`docs/ambientes.md`, sección 4).
- **Un Active Directory y un Keycloak de verdad**: lo que se prueba es OpenLDAP —que se habla igual—
  y un Keycloak de desarrollo, así que el código se prueba de verdad, pero lo específico de un AD
  —sus referencias entre dominios, sus atributos particulares, sus reglas de contraseña— no queda
  cubierto, ni las particularidades del Keycloak de una organización (`docs/modules/auth.md`,
  sección 11).
- **PKCE en el camino de Keycloak**, propuesto y no hecho: el cliente es confidencial y el secreto ya
  protege el canje, así que no estaba en el documento aprobado. Si se hace, el `code_verifier` viaja
  dentro del `state` firmado (`docs/modules/auth.md`, decisión 40).
- **Logs del frontend**: sin paquete decidido (el hermano es de NestJS y no sirve en navegador).

### Pendiente de decidir

La lista vive en un solo sitio: **sección 12**, para que no haya dos listas que puedan divergir.
