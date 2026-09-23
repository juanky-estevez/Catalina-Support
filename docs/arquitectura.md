# Arquitectura

> **Estado:** as-built
> **Última actualización:** 2026-09-22
>
> Aprobado por el responsable del proyecto el 2026-09-22 y **puesto en pie el mismo día**:
> existen el esqueleto de contenedores y los dos proyectos vacíos arrancando. Lo que todavía no
> existe está listado en la sección 13; cada módulo necesita además su propio documento antes de
> escribirse.

## 1. Alcance de este documento

Fija el **stack**, la **forma del repositorio**, los **contenedores** y las **reglas de
modularidad** de Catalina-Support. No define producto ni flujos: eso vive en los documentos de
cada área (propósito y alcance, usuarios y permisos, tickets, flujos).

Este documento es una **propuesta**: no habilita escribir código hasta que esté aprobado
(Regla 0 de `AGENTS.md`).

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

De ahí salen cuatro reglas concretas:

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
    ├── config               # variables de entorno
    ├── database             # conexión y configuración de GORM
    └── middleware           # autenticación, permisos, recuperación, logs
```

Se mantiene el estilo ya usado en el proyecto hermano Calibyou (`modules/<m>/controllers`,
`services`, `repositories`, `dtos`), porque es el que el responsable ya conoce y mantiene.

Hoy el repositorio tiene `main.go`, `.air.toml` y `shared/{config,database,httpx,middleware}`.
`modules/` y `migrations/` **todavía no existen**: nacen con el primer módulo documentado
(sección 13). `shared/httpx` se añadió al crear el esqueleto para que la forma de las respuestas HTTP
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
├── src
│   ├── main.ts
│   ├── index.html
│   └── app
│       ├── app.config.ts     # providers globales (router, http, ...)
│       ├── app.routes.ts     # rutas raíz; cada módulo aporta las suyas
│       ├── core              # armazón: layout, sesión, interceptores, guardas
│       ├── shared            # UI y utilidades reutilizables, sin lógica de negocio
│       └── modules
│           └── <módulo>      # pages, components, services, models, <módulo>.routes.ts
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

Pendiente y por decidir en su documento, no aquí: librería de estilos, gestión de sesión y
permisos en el cliente, y si se usa SSR (Angular lo trae, pero una mesa de ayuda interna no lo
necesita).

Hoy el repositorio tiene el armazón (`app.ts`, `app.html`, `app.css`, `app.config.ts`,
`app.routes.ts`), `core/services/health.service.ts` y las carpetas `core`, `shared` y `modules`
vacías. El armazón muestra el título y el estado del servicio, y el pie avisa de que es
provisional: esa pantalla desaparece cuando exista el primer módulo. `app.routes.ts` está vacío
porque las rutas las aporta cada módulo.

## 7. Base de datos (PostgreSQL 18)

- Una base de datos y un esquema; el aislamiento por cliente, si llega a existir, se decide en
  el documento de usuarios y permisos.
- **Una única vía para el esquema**: los archivos de `backend/migrations/` (hoy sólo
  `v1.0.0.sql`), transaccionales (`BEGIN` … `COMMIT`) e idempotentes (`CREATE ... IF NOT EXISTS`,
  `INSERT ... WHERE NOT EXISTS`). Se aplican con `psql -v ON_ERROR_STOP=1` dentro del contenedor de
  base de datos. Nada de `ALTER` sueltos ni pasos manuales al desplegar.
- Mientras el proyecto no tenga datos reales, se actualiza ese mismo archivo en vez de añadir
  migraciones nuevas. **A partir de la 1.0.0 se pasa a un archivo por versión**
  (`v1.1.0.sql`, `v1.2.0.sql`…), aplicados en orden: la política completa está en
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

Todo el correo de la aplicación sale por un único sitio: **`backend/shared/mail`**, con la
**biblioteca estándar** (`net/smtp`), igual que Calibyou. No se añade ninguna dependencia para esto.

| Quién manda | Qué manda |
| --- | --- |
| Módulo `tickets` | Los **siete avisos** de ticket (`docs/propósito-y-alcance.md`) |
| Módulo `auth` | Alta con el enlace de contraseña, cambio y recuperación (`docs/usuarios-y-permisos.md`) |

**Variables de entorno** (por entorno, nunca fijas en el código):

| Variable | Para qué |
| --- | --- |
| `SMTP_HOST`, `SMTP_PORT` | El servidor de correo |
| `SMTP_USER`, `SMTP_PASSWORD` | Las credenciales |
| `SMTP_SECURE` | Si la conexión va cifrada |
| `SMTP_FROM_NAME`, `SMTP_FROM_EMAIL` | Quién firma el correo |

Se llaman `SMTP_*` y no `SYSTEM_SMTP_*` como en Calibyou: allí el prefijo distingue el correo del
sistema del de cada laboratorio, y aquí sólo hay uno. Si faltan, el envío falla con un error claro y
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

## 10. Contenedores y ejecución

Todo corre en contenedores, con la misma forma que Calibyou: **un contenedor para el frontend,
otro para el backend y otro para la base de datos**, y nginx en la máquina como puerta de
entrada. No hace falta Go ni Node instalados en la máquina: sólo Docker y nginx.

### Estructura

```text
.
├── dev.yml                          # compose de desarrollo (los tres servicios)
├── prod.yml                         # compose de producción (imágenes construidas)
├── config
│   ├── dockerfiles                  # una imagen por servicio y por entorno
│   ├── env
│   │   ├── dev.env
│   │   ├── prod.env.example         # plantilla; prod.env NO se versiona
│   │   └── (prod.env)               # NO se versiona
│   └── nginx
│       ├── catalina-support-dev.conf    # vhost de desarrollo
│       ├── catalina-support-prod.conf   # vhost de producción
│       └── frontend.prod.conf           # nginx del contenedor de frontend
├── backend
├── frontend
├── _logs                            # logs de go-logs (la carpeta se versiona, los .log no)
├── _files                           # adjuntos de los tickets (la carpeta se versiona, los archivos no)
└── docs
```

### Servicios de desarrollo (`dev.yml`)

| Servicio | Imagen base | Comando | Puerto (sólo `127.0.0.1`) |
| --- | --- | --- | --- |
| `frontend` | `node:24-alpine` | `npm start -- --host 0.0.0.0 --port 11001` | `11001` |
| `backend` | `golang:1.27-alpine` + air | `air -c .air.toml` | `11002` |
| `database` | `postgres:18-alpine` | `postgres` (`PGPORT=11003`) | `11003` |

Detalles que importan:

- **Código montado como volumen** (`./frontend:/app`, `./backend:/app`): se edita en la máquina y
  el contenedor recarga (Angular con su servidor de desarrollo, Go con `air`). No se reconstruye
  la imagen para cada cambio.
- **Caches en volúmenes con nombre** (`node_modules` del frontend, módulos y caché de compilación
  de Go): ni se pierden al recrear el contenedor ni se mezclan con archivos de la máquina.
- **`_logs` montado** en el backend (`./_logs:/logs`) para que los archivos de `go-logs` queden a
  la vista en la máquina.
- **`_files` montado** en el backend (`./_files:/files`, variable `FILES_PATH`) para los adjuntos de
  los tickets: son archivos, no filas, así que sobreviven a los reinicios y se copian desde la
  máquina como cualquier carpeta.
- **El backend espera a la base de datos**: `depends_on` con `condition: service_healthy` y un
  `healthcheck` de `pg_isready`, para que no arranque contra una base que aún no acepta
  conexiones.
- **Puertos propios**: se usan `11001`-`11003`, distintos de los de Calibyou (`10001`-`10004`)
  para que los dos proyectos puedan estar levantados a la vez. Se publican **sólo en
  `127.0.0.1`**: nada de la base de datos ni del backend queda expuesto a la red.
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
  límite de 25 MB de los adjuntos (`docs/tickets.md`) no se podría cumplir. Los 30 dejan margen.
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
- Cuando exista autenticación habrá que añadir la zona `limit_req` del login, como en Calibyou.
  Hoy no hay ningún endpoint público que reciba credenciales, así que no se inventa.

### Producción (`prod.yml`)

- Imágenes **construidas** con Dockerfiles de dos etapas (compilar y luego ejecutar), no con el
  código montado:
  - backend: `golang:1.27-alpine` compila un binario estático y la imagen final es `alpine` con
    el binario (sin Go, sin `air`, sin código fuente).
  - frontend: `node:24-alpine` compila con `npm ci && npm run build` y el resultado estático lo
    sirve nginx dentro del contenedor `frontend` (puerto 21001).
- El vhost de producción ya está instalado y con certificado, pero **detrás no hay nada
  todavía**. Mientras dure el aplazamiento el dominio responde **503 con un aviso** («esta
  versión todavía está en desarrollo») en lugar del 502 que da nginx cuando no hay nada detrás:
  un 502 parece una web rota. La respuesta conserva las cabeceras de seguridad y la CSP estricta
  porque el bloque no declara `add_header` propio (si lo hiciera, nginx dejaría de heredar las
  del servidor). `/api/` sigue apuntando al backend que no existe, así que devuelve 502: nadie lo
  consume durante la pausa y así el día del despliegue sólo hay que restaurar un bloque, que está
  comentado en el propio vhost.
- **El despliegue a producción está aplazado hasta cerrar la versión 1.0.0** (decisión del
  responsable, 2026-09-22). Hasta entonces `prod.yml`, sus Dockerfiles y `frontend.prod.conf`
  quedan escritos pero **sin construir y sin probar**, y el runbook de despliegue (construir los
  artefactos, publicarlos y levantar `prod.yml`) se escribirá en `ambientes.md` antes de ese
  despliegue. Lo que sí está hecho y no se toca es la parte de nginx y TLS de la sección
  anterior: los certificados se renuevan solos y el vhost ya está en su sitio.

### Comandos

Todos los comandos se ejecutan **en la máquina**, contra los contenedores: no hay Go ni Node
instalados. Están también en `AGENTS.md`.

```bash
docker compose -f dev.yml up -d                 # levantar los tres servicios
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
Los puertos 11001 y 11002 siguen publicados en `127.0.0.1` para depurar sin pasar por nginx.

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

**Pendiente de decidir** (no se asume nada hasta su documento):

- Estilos del frontend, librería de componentes y diseño de pantallas.
- Autenticación, sesión, roles y permisos (documento de usuarios y permisos).
- Modelo de datos y estados del ticket (documento de tickets).
- Lista cerrada de módulos y qué endpoints expone cada uno.
- Notificaciones y correo.
- Runbook de despliegue a producción (documento de ambientes): cómo se construyen y publican los
  artefactos y cómo se levanta `prod.yml`. Los dominios, el TLS y los vhosts ya están hechos.
- Logs del frontend: no hay paquete decidido y el hermano no sirve en navegador.
- Pruebas: qué se prueba, con qué herramientas y en qué contenedor se ejecutan.
- Cómo se hace cumplir la regla de modularidad de la sección 4 (revisión manual o una prueba
  automática que falle si un módulo del frontend llama a un prefijo ajeno).

**Fuera de alcance de esta versión**: microservicios, colas, caché distribuida, SSR,
multi-tenant y aplicación móvil.

## 13. Estado de implementación

Lo que existe hoy en el repositorio y lo que se ha comprobado de verdad el 2026-09-22.

### Existe y está verificado

| Elemento | Comprobación |
| --- | --- |
| `dev.yml` con los tres servicios | Los tres contenedores levantan y quedan sanos (`database` con `healthy`) |
| Backend Go en contenedor con `air` | Compila, arranca, conecta con PostgreSQL y escucha en 11002 |
| Recarga en caliente de `air` | Al cambiar un archivo reconstruye y reinicia, con cierre ordenado del proceso anterior |
| `GET /api/health` | `200 {"database":"ok","status":"ok"}`, y 503 si la base no responde |
| `go-logs` | Escribe `_logs/catalina-support_dev_20260922.log` con `TIMEZONE` aplicado (`UTC-5`) |
| `go mod tidy` | Genera `go.sum`; el `go.mod` sólo lleva las dependencias reales |
| Frontend Angular 22 | El servidor de desarrollo sirve `index.html` y el bundle en 11001 |
| Pruebas del frontend | `npm test` → 2 pruebas en verde (armazón + estado del servicio) |
| Build de producción del frontend | `npm run build` → `dist/catalina-support/browser`, que es la ruta que espera `prod.yml` |
| Vhost de desarrollo | `https://dev-catalina-support.calibyou.com` sirve la aplicación (200) y `GET /api/health` devuelve `{"database":"ok","status":"ok"}` **a través de nginx** |
| TLS de los dos dominios | Certificados de Let's Encrypt emitidos el 2026-09-22 (caducan el 2026-12-21), con renovación automática configurada por `webroot` |
| Cabeceras de seguridad y CSP | Presentes en la respuesta de desarrollo (`nosniff`, `SAMEORIGIN`, `Referrer-Policy` y CSP de desarrollo) |
| Los sitios que ya existían | `calibyou.com`, `dev.calibyou.com` y `dev-ramona-fit.calibyou.com` siguen respondiendo 200 después de instalar los vhosts |

### Existe pero NO está verificado

- **`prod.yml` y los Dockerfiles de producción**: siguen la forma de Calibyou, pero no se han
  construido ni desplegado, y **no se hará hasta cerrar la versión 1.0.0**. Los puertos
  21001-21003 y las rutas `/root/prod/catalina-support` son una propuesta que hay que confirmar
  en `ambientes.md` antes de ese despliegue. El vhost de producción ya está instalado y con
  certificado, así que hoy **`https://catalina-support.calibyou.com` responde 503 con un aviso
  de que está en desarrollo**: es el comportamiento decidido, no un fallo.
- **`config/nginx/frontend.prod.conf`**: no se ha probado con nginx (depende del despliegue).

### Todavía no existe

- **Ningún módulo**, ni en el backend ni en el frontend: `backend/modules` no se ha creado y
  `frontend/src/app/modules` está vacío. Cada módulo necesita su documento aprobado antes de
  escribirse (Regla 0).
- **Ninguna migración**: `backend/migrations/v1.0.0.sql` se creará con el primer modelo de datos,
  cuando el documento de tickets lo defina. Hasta entonces no hay ninguna tabla.
- **Autenticación, sesión y permisos**: no hay nada, ni en el backend ni en el frontend.
- **Logs del frontend**: sin paquete decidido (el hermano es de NestJS y no sirve en navegador).

### Pendiente de decidir

La lista vive en un solo sitio: **sección 12**, para que no haya dos listas que puedan divergir.
