# Arquitectura

> **Estado:** as-built
> **Última actualización:** 2026-10-08
>
> **Bloque 7 implementado y verificado, 2026-10-08.** La sección 18 cierra las dos deudas técnicas
> declaradas en la sección 12: automatizar las fronteras de los módulos del frontend y decidir que
> el navegador no tendrá un sistema de logging de aplicación. Incluye una auditoría de todos los
> documentos `as-built`, porque la revisión encontró afirmaciones vigentes que contradicen el código
> actual. El responsable eligió 1A–15A y aprobó explícitamente el documento. El comprobador, sus
> pruebas, la integración con los comandos normales y la auditoría documental quedaron terminados.
>
> **Hallazgo de implementación del bloque 7, 2026-10-08.** El primer recorrido del comprobador
> encontró que la prueba del interceptor de `core` importaba la tabla raíz de rutas sólo para crear
> el `Router`, aunque ninguno de sus casos usa esas rutas. Es una dependencia `core → app` contraria
> a la regla aprobada. Se corrige proporcionando un router vacío en esa prueba; no cambia código de
> producto ni el comportamiento comprobado.
>
> **Enmienda implementada y verificada el 2026-10-08 (bloque 6).** El “repaso de formato” de la sección 13 era el
> único fleco técnico sin definición. La sección 17 lo coordina con interfaz y pruebas conforme a
> 1A–14C. Lo existente continúa as-built; el responsable aprobó explícitamente la propuesta el
> 2026-10-08.
>
> **Hallazgo corregido y verificado el 2026-10-08.** El bloque 5 encontró que la activación local no
> es transaccional al cambiar entre modelos con memoria ajustada. La sección 16 define preflight,
> segunda medición y restauración del modelo sano conforme a 15A–22A. Lo existente continúa
> as-built; la corrección fue aprobada explícitamente el 2026-10-08.
>
> **Hallazgo y enmienda propuesta el 2026-10-07.** Al revisar el bloque 4 se comprobó que los
> clientes de IA todavía siguen redirecciones automáticamente, el administrador local no supervisa
> un `llama-server` que muera después de arrancar y sus cálculos de disco y RAM no consideran bien
> una descarga parcial ni el límite del contenedor. Lo existente continúa as-built. La sección 15
> propone corregir esas diferencias; el responsable eligió 1A, 2A, 3A y 4A. El repaso cerró las
> decisiones 5A, 6A, 7A y 8A y no dejó asuntos abiertos. El responsable aprobó explícitamente la
> enmienda, que se implementó y verificó el 2026-10-07.
>
> **Implementado y verificado el 2026-10-07.** Los clientes de generación rechazan 3xx, limitan el
> cuerpo y fueron comprobados contra los cuatro contratos falsos. El administrador local supervisa
> el hijo, publica `active` y `healthy`, reanuda según bytes pendientes, elimina parciales corruptos
> y toma el menor límite de memoria entre máquina y cgroup. Pasaron todo el backend, `ai-manager`,
> 218 pruebas del frontend y Playwright completo (180 aprobadas y 38 omitidas).
>
> **Enmienda propuesta el 2026-10-06.** Lo existente continúa as-built. La sección 14 separa el
> motor local por entorno, añade un administrador sin socket Docker, adaptadores externos, secretos
> cifrados e idioma global. El repaso quedó cerrado, la propuesta fue aprobada y la implementación
> fue verificada.
>
> **Enmendado el 2026-10-06**, aprobado explícitamente por el responsable, implementado y verificado.
> La instancia desechable conserva el bind de código de sólo lectura, retira el volumen hijo de
> `/app/tmp` y arranca el backend de pruebas con `go run .`. La pasada canónica sin servicios
> opcionales terminó con 177 casos aprobados y 39 omitidos, y la limpieza no dejó recursos.
>
> **Enmendado el 2026-10-06**, conforme a la enmienda aprobada de
> `docs/interfaz-y-experiencia.md`: el frontend incorpora el toast compartido de `/settings` y
> `/setup`. En ese cierre pasaron las suites completas de frontend y de interfaz; las cifras
> históricas quedan en el registro del bloque correspondiente, no como inventario vigente.
>
> **Enmendado el 2026-10-04**, conforme a `docs/prueba-local.md` aprobado: en desarrollo,
> `backend` depende de `migrate` con `service_completed_successfully`; así nunca consulta
> `installation_settings` ni `ai_insights` antes de que existan. `migrate` ya no tiene perfil y el
> seeder conserva el suyo. Las sugerencias `SETUP_MAIL_*` sólo rellenan `/setup` vacío.
>
> **Enmendado el 2026-10-03**, conforme a `docs/prueba-local.md` aprobado: Arranque local aprobado
> en `prueba-local.md`: `dev.yml` incorpora `migrate` de un solo uso y retira el runner e2e.
> `tests.yml` posee una instancia desechable sin puertos publicados ni datos compartidos; Angular
> admite el host `frontend`. Keycloak anuncia localhost y conecta por una dirección interna
> independiente.
>
> **Enmendado el 2026-10-02**: revisión del recorrido del README. Se retira de §13 el pendiente
> de los botones del asistente: existen las pruebas públicas de entrada y correo, protegidas por el
> sello; la disponibilidad de IA se consulta automáticamente. El recorrido completo del asistente
> vacío sigue sin estar automatizado en Playwright. Se reporta además un hallazgo por lectura del
> despliegue actual: `prod-build.sh` construye desde la carpeta de artefactos, pero los Dockerfiles
> requieren las fuentes del repositorio. No se modifica código ni se retoma producción.
>
> **Enmendado el 2026-10-01 (quinta vez)**: **los datos de ejemplo se siembran con un contenedor de un
> solo uso**, `docker compose -f dev.yml run --rm seed`, **igual en Linux, macOS y Windows**, y con
> **perfil propio** para que **`up -d` no lo arranque**: los ejemplos son **opcionales** y se piden a
> propósito. El contenedor corre **el guion de siempre, `scripts/dev-seed.sh`** (una sola verdad), que
> también sigue valiendo en Linux y macOS fuera del contenedor. La sección 10.3 cambia su bloque de
> ejemplos por ese comando (docs/ambientes.md, sección 3.3).
>
> **Enmendado el 2026-10-01 (cuarta vez)**: **el servidor de desarrollo reenvía `/api` al backend**, así
> que **la 11001 basta para entrar sin nginx**. Se añade `frontend/proxy.conf.json` —una entrada para
> `/api` con destino **`http://backend:11002`**, el servicio del backend **dentro de la red del
> entorno**, porque quien hace la petición es el contenedor del frontend, no el navegador— y se declara
> en `serve.options.proxyConfig` de `frontend/angular.json`. **Es sólo de desarrollo**: vive en
> `serve`, el `build` no lo ve, y **en producción sigue reenviando nginx**. La entrada del frontend en
> `http://127.0.0.1:11001` deja de depender de nginx; en la tabla de requisitos, nginx pasa a **no ser
> necesario para entrar** (sí para probar el dominio y el certificado). Comprobado el 2026-10-01 por la
> 11001: `GET /api/auth/methods` y `GET /api/health` → **200 JSON**, `POST /api/auth/login` con
> `admin`/`admin` → **200** con token, y el recorrido de entrada y salida con Playwright, **14 casos en
> verde** contra `http://127.0.0.1:11001`. La regla del frontend no cambia: sigue llamando a rutas
> **relativas**.
>
> **Enmendado el 2026-10-01 (tercera vez)**: **la sección 10.3 («Comandos») estrena el paso del
> esquema**. Su lista empezaba por `docker compose -f dev.yml up -d` y quien la siguiera caía en el
> error de siempre: la aplicación **no crea las tablas al arrancar** y las migraciones **no se aplican
> solas**, así que en una base nueva el backend responde `relación "installation_settings" does not
> exist`. Quedan añadidos los **dos comandos de Docker que valen igual en PowerShell, CMD y bash**
> —copiar `v1.0.0.sql` al contenedor y dárselo a `psql` con `-f`—, con los ejemplos (`_dev`) a
> continuación y el aviso de que `./scripts/dev-seed.sh` **es de Linux y macOS** (los pasos completos,
> en el `README.md`).
>
> **Enmendado el 2026-10-01 (segunda vez)**: **el directorio de pruebas y Keycloak salen de `dev.yml`**
> y pasan a **`active-directory.yml`** y **`keycloak.yml`**, cada uno con **su propio comando**, y **el
> perfil `auth` desaparece**. Los dos entran en la red **`catalina-support-dev`**, que **posee
> `dev.yml`** (nombre fijo, sin `external`), así que **primero se levanta el entorno** y después el
> servicio; el backend los alcanza por su nombre de servicio (`ldap`, `keycloak`). Se ponen al día la
> estructura, la tabla de servicios de desarrollo, los comandos y la sección 13.
>
> **Enmendado el 2026-10-01**: **la fila del motor de IA de la sección 13 se corrige** para que diga lo
> mismo que el README y `docs/modules/ai.md`, con lo medido: **tope de 1500 MiB** (`mem_limit: 1500m`
> en `ai.yml`), **~1,44 GiB en marcha —el 98% del tope—** y **sin tope de CPU** (el compose no limita
> CPU; el «2 CPU» que decía la fila no salía de ningún sitio del compose). Se quita el «1,09 GB con una
> entrada de 2 034 piezas», que no se ha podido comprobar.
>
> **Enmendado el 2026-09-30**: **la sección 13 se pone al día**. Lo que ha cambiado este mes y no estaba
> contado: **la pantalla de Configuración está entera** —con **el motor de IA** y su prueba de conexión,
> la región horaria y la dirección pública, y **el correo saliente**—; hay **vista de primer arranque**
> (`/setup`, entonces cuatro pasos y hoy cinco tras la enmienda del 2026-10-06, y el sello en
> `installation_settings.installed_at`, con su API en **409**
> después); **el motor de IA vive aparte y se configura desde Configuración**; y **los recuentos** quedan
> en Go **14 paquetes**; los recuentos del frontend de esa fecha quedaron superados por la enmienda
> del 2026-10-06. **El despliegue a producción queda aparcado** hasta que el producto esté terminado
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

**El repaso integral de formato del frontend está cerrado**. **Las pantallas de producto —usuarios, tickets y
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
├── dev.yml                          # compose de desarrollo (los servicios; la red del entorno es suya
│                                    # y las de interfaz van detrás de un perfil)
├── prod.yml                         # compose de producción (imágenes construidas)
├── ai.yml                           # el motor local de IA, requerido sólo para la modalidad local
├── active-directory.yml             # el directorio de pruebas (OpenLDAP; opcional)
├── keycloak.yml                     # Keycloak de pruebas (opcional)
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
| `ldap` (**en `active-directory.yml`**, no en `dev.yml`) | `osixia/openldap:1.5.0` | El directorio, con las personas de `config/ldap/01-personas.ldif` sembradas al arrancar | `11005` (LDAP en claro: **es de desarrollo**). Opcional: **no se levanta con `dev.yml`**, se levanta con `docker compose -f active-directory.yml up -d` |
| `keycloak` (**en `keycloak.yml`**, no en `dev.yml`) | `quay.io/keycloak/keycloak:26.0` | `start-dev --import-realm`, con el reino de `config/keycloak/realm-catalina-support.json` | `11006` (HTTP: **es de desarrollo**). Opcional: **no se levanta con `dev.yml`**, se levanta con `docker compose -f keycloak.yml up -d` |
| `ai` (**en `ai.yml`**, no en `dev.yml`) | llama.cpp con servidor HTTP | Sirve el modelo **Qwen2.5-1.5B-Instruct Q4_K_M**, que se descarga una vez al volumen la primera vez | **Ninguno**: el motor no se publica en la máquina, sólo se habla por la red `catalina-support-ai`. Lo comparten **desarrollo y producción** (`docs/modules/ai.md`, decisión 1) |
| `migrate` | `postgres:18-alpine` | Aplica sólo `v1.0.0.sql`, con `ON_ERROR_STOP=1`; `backend` espera su finalización correcta | Ninguno; se ejecuta automáticamente con `up` y también admite `run --rm migrate` |
| `seed` | `postgres:18-alpine` | Ejecuta `scripts/dev-seed.sh` | Ninguno; ejemplos opcionales, ejecución explícita `run --rm seed` |

**El motor de IA no está en la tabla de arriba a propósito**: vive en `ai.yml`, **compartido por los dos
entornos** —no caben dos modelos en esta máquina— y los dos backends entran en su red
(`catalina-support-ai`). Su porqué, sus argumentos y su memoria están en `docs/modules/ai.md`, sección
2, y cómo se levanta en `docs/ambientes.md`, sección 3.4.

**El directorio de pruebas y Keycloak tampoco están en `dev.yml`**: cada uno vive en su archivo
(`active-directory.yml` y `keycloak.yml`), se levantan con su comando **después** del entorno y entran
en la red **`catalina-support-dev`**, que **posee `dev.yml`** con un nombre fijo (sin `external`): así
el backend los alcanza por su nombre de servicio (`ldap`, `keycloak`) sin que estén en el compose del
entorno. Su porqué y cómo se levantan, en `docs/ambientes.md`, sección 9.3.

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
- **`ldap` es el directorio de pruebas** (`active-directory.yml`): OpenLDAP con **tres personas** en un LDIF del repositorio
  (`config/ldap/01-personas.ldif`), una por cada caso del camino de AD. **No guarda nada en un
  volumen**: se siembra del repositorio al arrancar, así que siempre está como dice el repositorio, y
  **no se para y se arranca** —sus scripts de arranque no son idempotentes y muere—: se levanta otra
  vez recreándolo (`docs/modules/auth.md`, sección 11). Y **no es Active Directory**: se habla con él
  igual, así que el código se prueba de verdad, pero lo específico de AD no queda cubierto.
- **`keycloak` es el Keycloak de pruebas** (`keycloak.yml`): el reino vive **en el repositorio**
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
| Desarrollo | `dev.catalina-support.example.com` | `catalina-support-dev.conf` | frontend `11001`, backend `11002` |
| Producción | `support.example.com` | `catalina-support-prod.conf` | frontend `21001`, backend `21002` |

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
  eso `dev.catalina-support.example.com` está declarado en `serve.options.allowedHosts` de
  `angular.json`. Sin esa línea, la página responde 403 aunque nginx esté bien.
- **El servidor de desarrollo reenvía `/api` al backend** (`frontend/proxy.conf.json`, declarado en
  `serve.options.proxyConfig` de `frontend/angular.json`), con destino **`http://backend:11002`** —el
  nombre de servicio del backend **dentro de la red del entorno**, porque quien hace la petición es
  **el contenedor del frontend**, no el navegador— y **sin reescribir la ruta** (las rutas del backend
  ya llevan `/api`, igual que en nginx). Es **sólo de desarrollo**: es una opción de `serve`, el
  `build` no la ve y **en producción reenvía nginx**. Es lo que permite entrar en
  `http://127.0.0.1:11001` con sólo Docker y sin nginx delante.
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
docker compose -f dev.yml up -d --build         # esquema y servicios; no carga ejemplos
docker compose -f dev.yml ps                    # ver el estado
docker compose -f dev.yml logs -f backend        # seguir los logs (salida de go-logs)
docker compose -f dev.yml down                   # parar (los volúmenes se conservan)

# Repetición manual opcional del esquema idempotente; el `up` anterior ya lo ejecuta
docker compose -f dev.yml run --rm migrate
# Los datos de ejemplo, que son **opcionales** y **un extra para tener contenido, no un requisito
# para entrar**. **El comando de los tres sistemas** es un contenedor de un solo uso que corre el
# guion `scripts/dev-seed.sh` (esquema, ejemplos y adjuntos); **no lo arranca `up -d`** (perfil
# propio). En Linux y macOS el mismo guion se puede correr fuera del contenedor
docker compose -f dev.yml run --rm seed
# ./scripts/dev-seed.sh

docker compose -f dev.yml exec backend go test ./...      # pruebas del backend
docker compose -f dev.yml exec backend go vet ./...       # análisis estático
docker compose -f dev.yml exec backend go mod tidy        # tras añadir dependencias

docker compose -f dev.yml exec frontend npm test -- --watch=false
docker compose -f dev.yml exec frontend npm run build

docker compose -f dev.yml exec database psql -U catalina_support -d catalina_support -p 11003

docker compose -f active-directory.yml up -d    # el directorio de pruebas (opcional; tras el entorno)
docker compose -f keycloak.yml up -d            # Keycloak (opcional; tras el entorno)
docker compose -f ai.yml up -d                  # motor requerido si se eligió IA local
```

En desarrollo se entra por **https://dev.catalina-support.example.com** (nginx → contenedores).
Los puertos 11001 y 11002 siguen publicados para depurar sin pasar por nginx, y **la 11001 ya reenvía
`/api` al backend** con el proxy de `serve` (`frontend/proxy.conf.json`), así que **también se entra
en `http://127.0.0.1:11001` sin nginx**; la 11002 es el backend directo.

## 11. Entorno de desarrollo

Requisitos en la máquina de desarrollo (comprobados el 2026-09-22):

| Herramienta | Estado | ¿Obligatoria? |
| --- | --- | --- |
| Docker | 29.5.2 instalado | **Sí**: todo corre en contenedores |
| Docker Compose | incluido en Docker | **Sí** |
| nginx en la máquina | 1.24 instalado, **con los dos vhosts instalados y certificados** | **No para entrar**: la 11001 reenvía `/api` con su proxy. **Sí**, para probar el dominio y el certificado |
| Node.js | v24.19.0 instalado | No: el frontend compila en su contenedor (v24.21.0 en la imagen) |
| Go | **no instalado** | No: el backend compila y corre en su contenedor (1.27.1 en la imagen) |
| Cliente `psql` | 18.4 instalado | No: para consultar se entra al contenedor de base de datos |

Que Go y Node no hagan falta en la máquina es deliberado: **para entrar basta Docker** —nginx sólo
hace falta para probar el dominio y el certificado—, y las versiones quedan fijadas en las imágenes
(Go 1.27, Node 24), no en lo que cada
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
- **Dominios**: `dev.catalina-support.example.com` en desarrollo y
  `support.example.com` en producción, los dos con TLS de Let's Encrypt y servidos
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
**Cerrado en el bloque 7** (sección 18):

- **Logs del frontend**: no se incorpora un sistema de logging. El navegador sólo conserva el
  reporte del fallo fatal durante el arranque; la aplicación no registra actividad, cuerpos,
  credenciales ni errores funcionales.
- **Fronteras del frontend**: un comprobador propio, sin dependencias nuevas, hará fallar pruebas y
  compilación ante importaciones entre módulos, dependencias en sentido contrario entre capas o
  llamadas de un módulo a un prefijo API ajeno.

**Fuera de alcance de esta versión**: microservicios, colas, caché distribuida, SSR,
multi-tenant y aplicación móvil.

## 13. Estado de implementación

Lo que existe hoy en el repositorio y lo que se ha comprobado de verdad. La última pasada es del
**2026-10-08**: suites completas de backend y frontend, compilación de producción del frontend y
fronteras modulares automáticas. La última matriz visual completa también es del 2026-10-08. La
operación de producción continúa aparcada y no forma parte de esta evidencia.

### Existe y está verificado

| Elemento | Comprobación |
| --- | --- |
| `dev.yml` con sus servicios | Los contenedores levantan y quedan sanos (`database` con `healthy`); el directorio de pruebas y Keycloak se levantan aparte, con su archivo y su comando (`docker compose -f active-directory.yml up -d`, `docker compose -f keycloak.yml up -d`), **después** del entorno, que es quien crea la red |
| Backend Go en contenedor con `air` | Compila, arranca, conecta con PostgreSQL y escucha en 11002 |
| Recarga en caliente de `air` | Al cambiar un archivo reconstruye y reinicia, con cierre ordenado del proceso anterior |
| `GET /api/health` | `200 {"database":"ok","status":"ok"}`, y 503 si la base no responde |
| `go-logs` | Escribe `_logs/catalina-support_dev_20260922.log` con `TIMEZONE` aplicado (`UTC-5`) |
| `go mod tidy` | Genera `go.sum`; el `go.mod` sólo lleva las dependencias reales |
| Frontend Angular 22 | El servidor de desarrollo sirve `index.html` y el bundle en 11001, y **reenvía `/api` al backend** (`frontend/proxy.conf.json` → `http://backend:11002`): `GET /api/auth/methods`, `GET /api/health` y `POST /api/auth/login` con `admin`/`admin` responden **200 por la 11001** (sin nginx), y el recorrido de entrar y salir pasa **en un navegador de verdad** |
| El armazón de la sesión en el navegador | Las seis rutas se sirven por nginx (200) y el CSS de Tailwind llega compilado. Y **probado en un navegador de verdad** con Playwright, en PC y en móvil: se entra con la cuenta de fábrica, se la reconoce, se cambia de pantalla y se sale |
| Pruebas de interfaz con Playwright | `docker compose -f tests.yml run --rm e2e`, con el perfil `directory` cuando corresponde, recorre la instancia desechable en PC y móvil. La última matriz completa terminó correctamente el **2026-10-08** y limpió sus recursos; `docs/ambientes.md` conserva el alcance y `docs/prueba-local.md`, la evidencia histórica de cada recorrido. Incluye instalación desde una base vacía, sesión, marca, configuración, usuarios, tickets, correo, AD y Keycloak |
| Pruebas del frontend | `npm test -- --watch=false` ejecuta primero el comprobador de fronteras y sus casos aislados, y después toda la suite Angular. Terminó correctamente el **2026-10-08**. Cubre idioma global, sesión, temas, armazón, configuración, primer arranque, usuarios, tickets, correo, adjuntos, toasts y ayuda de redacción |
| Tailwind v4 y los temas | Instalado y compilando: el CSS servido lleva **las utilidades generadas** y las variables del tema, y **ningún componente escribe un color a mano**. Las fuentes se declaran a mano en `styles.css` (`@source './app'`), y hay un caso de Playwright que **falla si la hoja llega sin utilidades** |
| El tema | **Ocho temas** (dos de fábrica con `light-dark()` y el color institucional, y seis fijos con su paleta y su acento), elegidos con un atributo en `html` y el tema elegido con un atributo en `html`. **De fábrica sigue al sistema** —que es no haber elegido, y por eso **«automático» no se muestra ni se elige**—, el sistema manda en vivo mientras nadie haya elegido, la elección se recuerda en el navegador y se aplica **antes de arrancar** desde `main.ts` (no con un script incrustado: la CSP de producción no lo admite). Probado en un navegador **midiendo el color de fondo**, en claro, en oscuro y con la elección ganando al sistema |
| Build de producción del frontend | `npm run build` → `dist/catalina-support/browser`, que es la ruta que espera `prod.yml` |
| Vhost de desarrollo | `https://dev.catalina-support.example.com` sirve la aplicación (200) y `GET /api/health` devuelve `{"database":"ok","status":"ok"}` **a través de nginx** |
| TLS de los dos dominios | Certificados de Let's Encrypt emitidos el 2026-09-22 (caducan el 2026-12-21), con renovación automática configurada por `webroot` |
| Cabeceras de seguridad y CSP | Presentes en la respuesta de desarrollo (`nosniff`, `SAMEORIGIN`, `Referrer-Policy` y CSP de desarrollo) |
| Migración `v1.0.0.sql` | Aplicada con `psql -v ON_ERROR_STOP=1` **dos veces seguidas** y sin error: crea **21 tablas** —plantillas, cuentas y tokens; las **cinco tablas de configuración de una fila** (`installation_settings`, `ai_settings`, `directory_settings`, `keycloak_settings` y `ticket_settings`); dos auxiliares de IA; las diez de tickets; y `ai_insights`— con sus índices y restricciones. Incluye su puesta al día idempotente y la conversión histórica de texto plano a HTML sin tocar un cuerpo que ya contiene adjuntos. |
| Los datos de ejemplo, con su guion | `./scripts/dev-seed.sh` aplica el esquema, aplica `v1.0.0_dev.sql` y **copia los archivos de los adjuntos** a `_files/`, en la carpeta de su ticket. Deja **once cuentas** (`user1`…`user5`, `support1`…`support3`, `dev1`…`dev3`, todas con la contraseña de las pruebas) y **25 tickets** con su historia: **123 filas de historial y 27 comentarios**, siete con **ticket interno** (uno esperando a Desarrollo, otro resuelto, otro devuelto a Soporte), **cinco adjuntos que se descargan idénticos** a los del repositorio, y el contador de la numeración en 25. Comprobado entrando por la API con `user1` (ve sus cinco tickets) y con `support1` (la bandeja entera) |
| `shared/auth` | Firma y valida el token de sesión (HS256, 10 horas, `sub`), saca el `Bearer` de la cabecera y compara secretos en tiempo constante. **20 pruebas en verde**, incluidas la caducidad, otro secreto, el algoritmo `none`, los dos secretos vacíos y las cinco del `state` de OIDC |
| `shared/authz` | Comprueba el papel y responde 403 con `auth.forbidden`. **5 pruebas en verde**, incluido el caso de una ruta montada sin autenticación |
| Middleware de autenticación | Lee la cuenta **en cada petición**, deja la identidad en el contexto y responde 401 (`auth.session.invalid`/`auth.session.expired`) o 503 si la base no responde. **8 pruebas en verde** |
| `go test ./...` | La suite completa del backend terminó correctamente el **2026-10-08**. Cubre autenticación y autorización, correo, configuración, usuarios, tickets, IA, persistencia y middleware; los registros de cada módulo conservan los casos particulares sin duplicar aquí un recuento volátil |
| Los cuatro endpoints de `mail` | Probados de extremo a extremo contra desarrollo, con un token de fábrica firmado a mano (todavía no hay endpoint de entrada): sin cabecera, con un token con basura y con un token de otro secreto → **401 `auth.session.invalid`**; con token válido → **200** y las 22 plantillas; marcador inventado → **422 `mail.marker.unknown`**; asunto vacío → **422 `mail.subject.required`**; guardar sin el enlace → **200** con `missing: ["enlace"]`; restaurar → vuelve el texto de fábrica y `edited` pasa a `false`; prueba desde la cuenta de fábrica → **422 `mail.test.noEmail`**; clave o idioma que no existen → **404** |
| La cuenta se lee en cada petición | El middleware la resuelve con el cargador de `main.go`: la cuenta de fábrica tiene su identidad propia y cualquier otra se lee **de la tabla de cuentas en cada petición**, así que desactivar o cambiar un papel valen al instante |
| El log tras las pruebas | Ni una coincidencia de `Bearer`, `eyJ` ni la palabra `token`: los tokens no se registran |
| El camino local de `auth`, de extremo a extremo | Entrar con la cuenta de fábrica y con una cuenta local, `me`, `salir`, olvidar la contraseña, establecerla desde el enlace del correo y cambiarla desde dentro. **Probado con curl contra desarrollo**, con el enlace leído del buzón de pruebas |
| El camino de AD, de extremo a extremo | **Los cinco casos de la sección 5.4 de `docs/modules/auth.md`, comprobados uno a uno por curl** contra el OpenLDAP de `active-directory.yml` y con la fila que queda en la base: el **alta automática** (una persona del directorio sin cuenta entra y su cuenta nace con papel `usuario`, origen `ad`, su identificador externo y **sin contraseña**); el **vínculo** (una cuenta local con el correo de alguien del directorio entra con la contraseña de allí, pasa a `ad` **conservando su papel** y su contraseña local deja de servir); la **actualización** (los apellidos del directorio sustituyen a los de aquí); **el correo que cambia en el directorio** (cambié el `mail` de una persona con `ldapmodify` y entró **en la misma cuenta**, porque manda el identificador externo); la **reactivación** (desactivada aquí, el directorio la devuelve al entrar, con el aviso `users.directoryMayReturn` ya ejercitado de verdad); y el **directorio caído** (una cuenta de AD → 503 `auth.directory.unavailable`; un correo desconocido o una cuenta local → 401 `auth.invalidCredentials`, y una cuenta local con su contraseña buena entra igual) |
| Los enlaces de contraseña | Un enlace, un uso: el mismo token dos veces → 401 `auth.token.used`; un token inventado → 401; el de alta caduca en 24 horas y el de recuperación en 1, y lo dicen en el correo |
| El camino de Keycloak, por curl y en el navegador | **Por curl**, contra el Keycloak de desarrollo: `GET /api/auth/keycloak/start` → **302** a su pantalla de entrada con `client_id`, `redirect_uri`, `response_type`, `scope` y `state`; una vuelta con un `state` inventado → **302 a `/login#error=auth.oidc.state`**; y una vuelta con un `state` bueno y un código que no vale → **302 a `/login#error=auth.oidc.rejected`**, que es la prueba de que el canje se hace de verdad contra Keycloak. **En un navegador de verdad** (Playwright): el botón, la pantalla de Keycloak, la vuelta, la sesión y **el fragmento borrado**. Y **sin Keycloak configurado**: `methods` dice `keycloak: false` y su ruta responde **404**, así que esa instalación no ofrece el botón ni tiene el camino |
| Los adjuntos en los comentarios | **Tres puertas y una sola regla**: pegar, arrastrar y elegir con el botón, con la misma lista cerrada de extensiones y el mismo aviso. Comprobado en el navegador con **un evento de pegado de verdad**: el archivo aparece en la lista, se publica el comentario con sus archivos, y **un comentario sin texto** —cuerpo vacío en la base— se lee por sus adjuntos **sin dejar un párrafo vacío**. Y **el camino del fallo**, cortando la subida en el navegador: el aviso dice qué archivo falta, el comentario queda publicado **una sola vez** y el botón de reintentar lo sube al mismo comentario |
| Los adjuntos dentro del texto | El editor con formato: **pegar, arrastrar y el botón meten el archivo donde está el cursor**, la imagen **se ve mientras se escribe** y **después de guardar** con su tamaño, un PDF sale como **enlace que abre el visor** (más grande, con **Descargar** y **Abrir en una pestaña**) y un Word **descarga**. Comprobado además: **lo que se guarda es el texto con sus referencias y nada más** (ni `class`, ni `src`, ni `blob:`), **un adjunto sin nombrar sigue al final**, **la búsqueda no encuentra por una etiqueta**, y **editar** un comentario con una imagen **enseña la vista previa en el editor**. En el backend, `go test` cubre el saneador: seis formas de HTML prohibido dan **422 `tickets.body.notAllowed`** sin dejar fila |
| Los dos logos de fábrica | Que la aplicación enseñe **el que toca al tema que se está viendo**: con un tema claro carga `logo-catalina-support-light.png` y con uno oscuro el `dark`, comprobando que **la imagen carga de verdad** (`naturalWidth`) y que son dos archivos distintos. Los dos salen del repositorio (`frontend/public/`), con fondo transparente, y el caso quita primero cualquier logo propio y deja la instalación como estaba |
| La política de contraseñas y el enlace | **Mínimo 8 caracteres** (bajado de 12 el 2026-09-25), comprobado en el navegador: una contraseña corta se rechaza con su clave **y no gasta el enlace** —se vuelve a intentar en la misma pantalla y entra—, y ocho caracteres valen. El caso destapó que el enlace se gastaba antes de comprobar la contraseña, y está corregido |
| El despliegue a producción | **`scripts/prod-build.sh`** construye y publica los dos artefactos en `/srv/catalina-support` (rota el anterior **por copia**, verifica que el artefacto existe y corta si no, y escribe `BUILD_INFO` con el commit, la rama, si el árbol estaba sucio y la fecha) y levanta los contenedores con `prod.yml` reiniciando **sólo el backend**. Los tres contenedores de producción están levantados y sanos, y el `admin` entra con la contraseña de `config/env/prod.env`. **El dominio sigue respondiendo 503**: la 1.0.0 no está cerrada. **Este despliegue queda aparcado** hasta que el producto esté terminado (decisión del responsable, 2026-09-30) |
| La versión del sistema | Publicada en la marca pública (`GET /api/settings/brand` → `"version": "1.0.0"`) y leída en los dos sitios: **en la fila de salir del menú lateral, alineada a la derecha**, y **en el pie de la pantalla de entrada**. Comprobado en el navegador en PC y en móvil, contra lo que dice la API —no contra un número escrito en la prueba—, que es texto y no un enlace, que se lee sobre el fondo del menú, y que **plegado no se enseña**. Y que sin marca no se enseña ningún número: la pestaña no se queda sin nombre y el pie sin versión, en vez de inventarse uno |
| El nombre de la instalación, por curl y en el navegador | Cambiado por la API y **leído antes de entrar** (`GET /api/settings/brand`): un nombre de 61 caracteres → **422 `settings.name.tooLong`**, y uno en blanco → vuelve `Catalina Support`. Y en el navegador: el menú lateral, la pestaña y la pantalla de entrada lo enseñan, y al vaciar el campo vuelve el de fábrica |
| Las copias de la base | `scripts/backup-db.sh dev`, ejecutado de verdad: deja el volcado en `/srv/catalina-support/backups/`, **comprueba que se puede leer** (`pg_restore --list`) y **se restauró en una base nueva**: 11 tablas y los mismos registros que la de verdad (tickets, cuentas, plantillas y adjuntos, contados uno a uno). Y la retención, probada con una copia de hace veinte días: se borra, y no se toca la de tres días ni la de hoy |
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
| Las dos pruebas de conexión | El botón «Probar la conexión» de las dos tarjetas, contra el directorio y el reino de verdad: contesta que sí cuando el servicio está y **con su clave cuando no** —probado apuntando el servidor a uno que no existe—. **No guarda nada**: al recargar, la configuración sigue como estaba. Se salta sola si el servicio no está levantado, y la suite entera corre sin él |
| La región horaria y la dirección pública | En **Configuración**, desde el 2026-09-29: **la zona horaria** se elige de una lista de zonas (IANA) con buscador —y se enseña la hora que es en ella y su desfase— y **la dirección pública** (esquema, host y puerto, `localhost` incluido), con el aviso de que **si no es https la sesión y la contraseña viajan sin cifrar**. **La zona decide cómo se leen todas las fechas**, en la interfaz y en los correos, y **las guardadas siguen en UTC**: cambiarla no mueve ningún ticket. **La dirección es la base de los enlaces de los correos y de la vuelta de Keycloak**; `PUBLIC_APP_URL` queda como respaldo (`docs/modules/settings.md`, decisiones 14 y 15) |
| El correo saliente, en la base | Las variables `SMTP_*` **se retiraron del entorno**: el SMTP vive en `installation_settings.smtp_*` y se configura en **el paso 4 de la vista de primer arranque** o, después, en **Configuración**. Comprobado en desarrollo: el archivo de datos de ejemplo deja puesto el buzón de pruebas (`smtp_host='mail'`, `smtp_port='1025'`) y el correo sale por él (`docs/primer-arranque.md`, sección 5) |
| El motor de IA, configurable desde la pantalla | **Su dirección y su modelo se configuran en Configuración**, con **su tarjeta y su botón de «Probar la conexión»** (prueba `<dirección>/health` con un tiempo corto y **no guarda nada**): el módulo `ai` los lee **en cada petición**, y `AI_URL`/`AI_MODEL` quedan **como respaldo**. La prueba se salta sola si el motor no está levantado (`docs/modules/settings.md`, decisión 16, sección 5.13) |
| La vista de primer arranque (`/setup`) y el sello | En una instalación **sin sellar** —`installation_settings.installed_at` nulo— la aplicación lleva a **`/setup`**, que pide **en cinco pasos** instalación, entrada, ubicación, correo e IA, y al terminar **sella** la instalación: la vista no vuelve y su API contesta **409 `setup.alreadyInstalled`**. **Una instalación que ya estaba configurada queda sellada al actualizar**, y en desarrollo **el archivo de ejemplos sella**, para que el asistente no salga en cada arranque. Comprobado en navegador y por la API, incluido el 409. `docs/primer-arranque.md` |
| Los endpoints de `users` | Probados de extremo a extremo: la lista con filtros y búsqueda, la ficha, los cambios, el perfil propio, desactivar y reactivar. Y los límites: Soporte no cambia el papel ni el correo de otro (**403**) ni ve la ficha de nadie, nadie se desactiva a sí mismo (**403 `users.selfDeactivation`**), el estado no se cambia por el `PATCH` (**422**), el origen inventado (**422**) y pasar a directorio se rechaza mientras no exista su consulta (**422**) |
| El editor de los correos | La pantalla del módulo `mail`: los once correos a la izquierda, **los dos idiomas al lado** —apilados en móvil, medido por la prueba—, el cuerpo con sus botones de formato **envolviendo lo seleccionado**, los marcadores que **se insertan al pulsarlos**, la **vista previa que renderiza el backend** con los datos de ejemplo de la prueba, y los botones de guardar, volver al de fábrica y **enviarme una prueba**, que se lee en el buzón. Probado de extremo a extremo |
| El idioma de la instalación | Es global para interfaz, correos y nuevos resúmenes; las cuentas no guardan idioma propio |
| Las pantallas de tickets | La **bandeja** —una sola con tres nombres, tabla en PC y tarjetas en móvil, con chips de estado, búsqueda, paginación y el chip de tipo del Administrador—, el **alta** con arrastrar y soltar, y el **detalle**: una línea de tiempo con los comentarios y lo que hizo el sistema, la ficha con las acciones —a la derecha en PC y debajo en móvil, medido por la prueba—, **la vista doble** cuando hay interno y los adjuntos **con vista previa**. Probado en PC y en móvil, incluido que **el usuario no ve el interno**, que **Desarrollo no escribe en el principal** y que **al Administrador no se le ofrece ningún botón** |
| El reparto y el prefijo, desde Configuración | La pantalla de Configuración tiene ya el **prefijo** —con el aviso de que no cambia los números ya emitidos— y el **reparto y el aviso** de cada tipo de ticket, con la opción «al asignado» desactivada cuando no se reparte |
| Las tres pantallas de usuarios | La **lista** (tabla en PC y tarjetas en móvil, con los chips de papel, origen y estado, la búsqueda y la paginación), **la ficha** y **el perfil propio**, probados en PC y en móvil. Soporte **edita en línea** desde la lista y no tiene ficha de nadie; no se le ofrece desactivarse a sí mismo; la cuenta de fábrica no tiene perfil ni enlace a él, y su nombre en el menú no es un enlace |
| El menú llega hasta abajo | El cajón **mide la zona visible y se desplaza por dentro**: en un móvil con barra del navegador (839 px visibles de 890) el botón de salir quedaba fuera de la pantalla y no se podía pulsar. Lo encontró Playwright, y con él se corrigió además que en PC la barra se estirara a todo el contenido |
| Un solo `h1` por pantalla | Las tarjetas de las pantallas de producto titulan con `h2`: la de usuarios tenía tres `h1` y el perfil otros tantos. Lo dijo la prueba al encontrar dos encabezados con el mismo nombre donde esperaba uno |

| Las diez tablas de `tickets` | Creadas en `v1.0.0.sql` y aplicadas **dos veces seguidas** sin error: los dos tipos de ticket, la conversación, los adjuntos, el historial y el contador de la numeración, con sus restricciones (el interno no puede estar `escalado`, exactamente un destino en lo que cuelga de un ticket, el motivo del escalado no puede estar vacío) |
| Las extensiones del texto y el código | Comprobado de punta a punta por la pantalla: un `.sql` **se coge** —sin el aviso de rechazo—, **se guarda** su referencia en el comentario, **queda entre los adjuntos del ticket** y **al pulsarlo se descarga** con su nombre; y un `.svg` **se sigue rechazando** antes de subirlo. Las **cuarenta y dos** extensiones nuevas son las mismas en los dos sitios —el editor y el backend—, y el editor lo comprueba por su lado con sus pruebas de unidad |
| El módulo `ai`, con el motor de verdad | **Comprobado de punta a punta contra el contenedor**: al crear un ticket, `Pedir` vuelve sin esperar la generación y los resúmenes pasan de `pendiente` a `listo`. En la pantalla, la lista y la ficha muestran motivo y última acción; el interno conserva como motivo el escalado. La IA es obligatoria para completar una instalación. Si el motor configurado cae después, los trabajos quedan `sin_motor` o pendientes y la mesa de ayuda sigue operando mientras el monitor intenta recuperarlo |
| El motor de IA, medido | El contenedor `catalina_support_ai` (llama.cpp `server-b11206` + Qwen2.5-1.5B-Instruct Q4_K_M) tiene un **tope de 1500 MiB** (`mem_limit: 1500m` en `ai.yml`) y **sin tope de CPU** —el compose no limita CPU—, **sin publicar ningún puerto**. Medido en marcha: **~1,44 GiB**, el **98% del tope** (`docker stats` da 1.441-1.448 GiB; el `VmRSS` del proceso, 1,51 GB). **Tarda entre 12 y 24 segundos por campo** y unos 95 s en un ticket de 6 000 caracteres; la segunda llamada del mismo ticket es ~7 veces más rápida (caché de prefijo). Su comprobación de salud contesta en 325 ms mientras está generando |
| El JSON del modelo, sin creérselo | **Con `response_format` y la exigencia escrita al final del mensaje, 4 de 4 respuestas válidas** con un ticket de 6 000 caracteres; con sólo una de las dos vías, el modelo envuelve la respuesta en un bloque de código, se inventa las claves o contesta en un solo idioma. El módulo rescata el primer objeto JSON y reintenta; cuando no vale, el log dice **la forma** de lo que llegó (claves y longitudes) y nunca su contenido |
| Las categorías y las etiquetas | **Comprobado por las tres capas**: la migración crea «General» y se la pone a todo lo que existe —simulando una base vieja, el ticket sin categoría acaba en «General» y la columna queda `NOT NULL`—; el alta **sin categoría se rechaza** (`tickets.category.required`) y **un usuario no puede crear ni retirar** categorías; las etiquetas sucias se guardan normalizadas (`["Red Wifi","  Licencias  ","red-wifi","Ñoño_Ütil"]` → `['red-wifi','licencias','nonoutil']`); un ticket con 40 caracteres de etiqueta se rechaza; retirar la última activa se rechaza; `?category=` y `?tag=` filtran; y **buscar «red» encuentra la categoría «Red» y la etiqueta `red`**. En la pantalla: el alta pide la categoría, las fichas se normalizan al escribirlas, los chips salen en el renglón, hay filtros de las dos y el catálogo lo mantiene quien toca |
| Las menciones y los observadores | **Comprobado por las tres capas**: el saneador admite `<span data-mencion="12">` y rechaza lo que no es un identificador; al comentar con una mención se crea el observador, el ticket **entra en el «Observo» de esa persona y no en sus «Asignados»**, quitar al observador deja su entrada en el historial y **no vuelve** al reenviar el mismo comentario, un usuario que lo intenta recibe `tickets.mention.notAllowed`, y el aviso `ticket.mentioned` llega al buzón de pruebas. En la pantalla: el buscador de personas, el nombre resaltado dentro del comentario, la línea «Observadores» de la ficha con su aspa y el chip de vista |
| El botón de copiar el número | De sólo icono, y al pulsarlo un visto verde dos segundos. La prueba **mide el color pintado** y lo compara con el `--exito` del tema, no con el nombre de la clase —y ahí salió que el botón tiene transición, así que la primera lectura coge el color a medio camino— |
| El filtro «lo mío» | `GET /api/tickets?mine=1` devuelve **exactamente lo que dice la consulta**: comprobado contra la base con `support1` (**16 tickets**: 9 principales asignados, abiertos o comentados por él y 7 internos escalados o comentados), con `dev1` (**3 internos**) y con `user1` (**5**, todos suyos: el usuario sigue viendo sólo lo suyo con o sin `mine`) |
| El backend de `tickets`, de extremo a extremo | **78 comprobaciones contra desarrollo**, leyendo los correos del buzón de pruebas: el alta con su **número** (`CS-2026-0001`, y el 9999 que crece a cinco dígitos), el **reparto por turnos** que alterna entre los técnicos activos, el triaje, el **escalado** con su motivo obligatorio, las reglas de sincronización (el interno en espera y el interno resuelto devuelven el principal a `en progreso`), la **devolución** al cerrar el interno sin resolverlo, el **re-escalado** con el mismo interno, el cierre, la reapertura con las fechas limpias, los comentarios (editar y borrar vaciando el texto), los adjuntos (subir, descargar, la extensión fuera de la lista, el `svg` y los 25 MB) y **los permisos de los cuatro papeles**, incluido que el usuario no alcanza un ticket interno ni por su número ni por la ruta de un adjunto |
| Los siete avisos de `tickets` | Salen cuando la transición que los provoca lo dice, y se han leído en el buzón: al asignado, al usuario (en espera, resuelto, cerrado), a Desarrollo (escalado, también al re-escalar) y a Soporte (Desarrollo necesita algo, y el ticket ha vuelto a su bandeja) |
| Los sitios que ya existían | `example.com`, `dev.example.com` y `other-app.example.com` siguen respondiendo 200 después de instalar los vhosts |

### Existe pero NO está verificado

- **La reproducción del despliegue actual desde cero queda bloqueada**: el script publica el
  compose en `/srv/catalina-support` y usa esa carpeta como contexto de construcción, pero
  los Dockerfiles necesitan `frontend/package*.json`, las fuentes del frontend y las del backend.
  El script no publica esas fuentes. Hallazgo por lectura el 2026-10-02; no se ha ejecutado contra
  producción. Se requiere propuesta y aprobación antes de corregir su comportamiento.


- **El camino público de producción, aparcado**: el contenedor de frontend está levantado con su nginx,
  pero **el dominio sigue respondiendo 503 a propósito** y **el despliegue no se retoma hasta que el
  producto esté terminado** (decisión del responsable, 2026-09-30), así que el vhost público y
  `config/nginx/frontend.prod.conf` sirviendo la aplicación no se han probado.

### Limitaciones y trabajos aplazados

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

### Pendiente de decidir

No quedan decisiones arquitectónicas abiertas. Los trabajos de producción están aplazados por una
decisión ya tomada, no pendientes de diseño.

## Instancia desechable de pruebas (`tests.yml`, 2026-10-03)

Proyecto `catalina-support-tests`, con PostgreSQL, migración de un solo uso, backend, frontend y
Mailpit propios. El backend espera a la migración completada; el runner espera salud del backend.
Todos los datos, adjuntos, logs y dependencias tienen volúmenes del proyecto; el código entra
en sólo lectura. No hay puertos publicados ni red de IA compartida. LDAP y Keycloak entran sólo
con el perfil `directory`. Los informes de Playwright se escriben en `tests/e2e/resultados/`,
que se conserva después de `docker compose -f tests.yml --profile directory down -v`.
El recorrido de preparación, ejecución y limpieza está en `docs/ambientes.md`, sección 9.3.

### Proceso del backend corregido (2026-10-06)

El código permanece montado en `/app:ro`. La suite sustituye el comando de la imagen —Air, útil para
desarrollo— por `go run .`, apropiado para una ejecución inmutable. Go usa `backend_go_mod` y
`backend_go_build`; su artefacto temporal queda fuera de `/app`. Desaparecen el montaje
`backend_tmp:/app/tmp` y el volumen `backend_tmp`.

No cambia la imagen compartida de desarrollo, `.air.toml`, la dependencia de la migración, las redes,
los datos ni ningún contenedor de desarrollo o producción.

El repaso confirmó una sola compilación por pasada, únicamente las dos cachés de Go ya existentes y
limpieza completa con `down -v`. No quedan decisiones arquitectónicas abiertas.

La implementación se verificó desde volúmenes nuevos: la migración terminó antes del backend, el
servicio compiló con `go run .`, respondió salud y permitió completar los 216 casos de la pasada sin
perfiles opcionales —177 aprobados y 39 omitidos—. El `down -v` posterior retiró todos los recursos y
`docker compose -f tests.yml ps -a` quedó vacío.

## 14. Arquitectura implementada: IA obligatoria y un idioma global

### 14.1 Componentes y dependencias

`tickets` mantiene su interfaz de resúmenes y no conoce proveedores. `ai` conserva cola,
normalización y persistencia y añade adaptadores para OpenAI/DeepSeek/OpenAI-compatible y Claude.
`settings` es dueño de la configuración, cifrado, confirmación de privacidad y estado requerido.
`mail` es dueño de plantillas y validación de sus borradores. `main.go` conecta las interfaces sin
importaciones cruzadas nuevas.

El motor local sigue en `ai.yml`, pero su proceso 1 es un administrador pequeño que descarga modelos
y supervisa un proceso hijo `llama-server`. No monta `/var/run/docker.sock`, no publica su API y
exige `AI_MANAGER_TOKEN`. Backend sólo puede pedir catálogo, progreso, descarga, activación, prueba y
borrado mediante la red privada del entorno.

### 14.2 Aislamiento local

Desarrollo y producción usan proyectos, redes y volúmenes distintos. El nombre del contenedor deja de
ser global y la URL interna se resuelve por servicio. El catálogo es el mismo, pero cada volumen
decide qué archivos conserva y cuál está activo. El techo de memoria es 8 GB y el administrador hace
la comprobación de RAM/disco antes de iniciar el modelo.

El modelo local predeterminado continúa siendo Qwen2.5-1.5B-Instruct Q4_K_M. `ai.yml` ya no fija la
ruta del modelo en `command`; el administrador la elige después de validar el catálogo versionado en
el repositorio y el checksum fijado allí.

### 14.3 Datos

La configuración detallada de IA sale de las dos columnas de `installation_settings` y pasa a una
fila propia de `settings`. La credencial se guarda como cifrado autenticado, con versión, nonce y
texto cifrado; la clave maestra `AI_CREDENTIAL_KEY` vive sólo en el entorno. Nunca se usa
`TOKEN_SECRET` para cifrarla. Son 32 bytes aleatorios codificados en base64 y el proceso rechaza el
arranque si falta o tiene formato inválido. Una clave válida que no abre datos existentes pone la
instalación en estado de configuración requerida.

`users.language` desaparece. `installation_settings.language` queda como única fuente. La marca
pública incluye ese idioma para pintar entrada y armazón antes de tener sesión.

La instalación guarda además una versión creciente de su configuración global. Todas las respuestas
autenticadas publican `X-Catalina-Language` y `X-Catalina-Settings-Version`; el interceptor actualiza
la señal de idioma cuando observa una versión nueva. No se invalida el token, no se añade sondeo y
una pestaña adopta el cambio en su siguiente petición.

`ai_insights` deja de guardar dos textos por campo: guarda uno, su idioma, proveedor y modelo. Al
cambiar el idioma, una actualización marca para regenerar los textos de otro idioma sin borrar el
histórico de ticket del que se reconstruye el encargo.

### 14.4 Disponibilidad y actualización

La IA es requisito de configuración, no dependencia síncrona de cada acción. Una instalación
configurada sigue guardando tickets durante una caída. Un monitor de salud con espera creciente
reactiva la cola cuando vuelve. Con motor local o servidor propio procesa el acumulado; con proveedor
deja el lote pausado hasta que el Administrador vea el volumen y confirme el posible costo. Sólo el
estado de configuración inicial bloquea la entrada.

Las instalaciones selladas se migran sin abrir `/setup`. Si no hay configuración probada, el
middleware permite salud, marca, entrada/salida y las rutas mínimas de Configuración para
Administrador/cuenta de fábrica; las demás rutas informan que falta IA.

### 14.5 Seguridad y pruebas

Los clientes HTTP limitan tiempo, cuerpo de respuesta y redirecciones; nunca registran prompts,
respuestas o cabeceras secretas. El servidor propio permite autenticación sin relajar TLS. Los
adaptadores se prueban contra servidores falsos y el administrador contra archivos pequeños de
prueba; CI y la suite local no llaman a proveedores ni descargan los GGUF reales. El catálogo fija
fuente, licencia, tamaño y checksum, y registra la aceptación por versión antes de descargar.

Esta propuesta afecta backend, frontend, migración abierta `v1.0.0.sql`, `ai.yml`, configuración de
entornos y pruebas. No cambia nginx público, los caminos ni tokens de autenticación, ni el contenido
de adjuntos.
Fue decidida por el responsable el 2026-10-06. El repaso cerró claves, plantillas personalizadas,
costos de recuperación, licencias y propagación del idioma a sesiones abiertas. No quedan decisiones
abiertas; propuesta aprobada.

## 15. Cierre implementado del bloque 4

### 15.1 Redirecciones y credenciales

Las generaciones y pruebas que llevan Bearer, cabecera propia, Basic o clave de Claude **no siguen
redirecciones**. Una respuesta 3xx se trata como conexión fallida y obliga a configurar la URL final.
La regla se aplica a OpenAI, DeepSeek, Claude y compatible con OpenAI, al probar y al resumir.

Las descargas públicas de GGUF sí pueden seguir redirecciones porque Hugging Face las necesita, pero
con un máximo finito y sin credenciales de Catalina Support. Cada salto sólo puede usar `http` o
`https`; el cliente limita conexión y cabeceras, y la descarga no supera el tamaño declarado.

### 15.2 Supervisión del proceso local

Después de activar un modelo, el administrador espera al proceso hijo. Si `llama-server` termina
mientras ese modelo sigue activo, lo reinicia indefinidamente con esperas de 1, 2, 4, 8, 16, 32 y
después 60 segundos entre intentos. Sólo
existe un supervisor por modelo activo. Cambiar de modelo o apagar el administrador cancela el
supervisor anterior; una terminación provocada por ese cambio no cuenta como fallo.

El catálogo expone por separado `active` y `healthy` y muestra el último error operativo. El
endpoint de salud del administrador sigue indicando que el administrador vive; la generación de
prueba confirma que el modelo atiende.

### 15.3 Disco y memoria efectivos

Para descargar se suman únicamente los bytes que faltan de cada archivo, contando sus `.partial`,
más 256 MiB de margen. Un parcial mayor que el tamaño declarado se descarta y se reinicia desde cero.
Si alcanza el tamaño esperado pero falla formato o checksum, también se elimina para que el siguiente
intento descargue un archivo limpio.

La memoria disponible es el menor valor entre `MemAvailable` y la capacidad restante del cgroup v2
o v1 cuando haya un límite finito. Si no hay cgroup se usa `MemAvailable`. El modelo no arranca
cuando su estimación supera esa capacidad efectiva.

### 15.4 Criterios de aceptación

1. Los cuatro proveedores se prueban con servidores HTTP falsos y sin red comercial.
2. Una respuesta 3xx no provoca otra solicitud ni reenvía credenciales.
3. Una respuesta mayor al límite se rechaza sin crecer indefinidamente en memoria.
4. La autenticación llega sólo al servidor falso configurado.
5. El administrador reanuda un parcial válido y reserva sólo los bytes pendientes.
6. Los límites simulados de cgroup prevalecen sobre la memoria de la máquina.
7. La muerte inesperada del hijo inicia la espera creciente; activar otro modelo cancela el
   supervisor anterior sin dejar dos procesos.

Quedan fuera redirecciones configurables, socket Docker, GPU, límites de gasto y llamadas a
proveedores reales.

### 15.5 Decisiones y repaso

El responsable eligió el 2026-10-07: rechazar redirecciones en llamadas con credenciales (1A),
supervisar y reiniciar el motor local (2A), medir disco pendiente y memoria efectiva del cgroup (3A)
y cubrir todo con pruebas aisladas (4A). En el repaso eligió reintento indefinido con espera máxima de
60 segundos (5A), estados `active` y `healthy` separados (6A), eliminar parciales corruptos (7A) y
permitir como máximo cinco saltos HTTP/HTTPS en descargas públicas (8A). No quedan decisiones
abiertas. Producción y cualquier descarga real permanecen fuera de este bloque.

## 16. Activación local transaccional implementada

El administrador trata el cambio de modelo como una transacción operativa: origen sano, preflight,
parada, medición real, arranque y confirmación del destino. Conserva identificador y especificación
del origen hasta que el destino está sano. El archivo `active` representa el último modelo sano, no
el último intento.

El preflight suma memoria efectiva disponible y RSS del hijo actual. No cuenta la memoria del propio
administrador. Después de terminar el hijo vuelve a leer host y cgroup; si el requisito no se cumple,
restaura el origen. Cualquier error de proceso o salud sigue el mismo rollback y espera su salud antes
de responder.

El contrato de error de memoria transporta una clave estable, bytes requeridos y bytes disponibles.
No expone PID, rutas ni detalles del host. La misma regla sirve a `/setup` y `/settings` mediante el
adaptador actual. No se añade concurrencia de modelos, reserva de memoria, GPU ni uso forzado de swap.

La prueba real cambió 1.5B→3B aprovechando el RSS recuperable del origen. El intento de 7B devolvió
6.442.450.944 bytes requeridos y alrededor de 5,07 GiB disponibles sin interrumpir el 3B sano. Las
pruebas deterministas cubren además la segunda medición insuficiente y el fallo de salud, con
restauración del origen antes de responder.

## 17. Bloque 6: repaso integral del frontend

El bloque elimina el fleco “repaso de formato” mediante un inventario completo de todas las pantallas
y su corrección total. La implementación permanece en Angular, Tailwind y los componentes propios;
un patrón repetido se resuelve en la capa compartida y una composición singular en su pantalla. No
se incorporan dependencias visuales ni una herramienta de comparación por capturas.

La frontera arquitectónica es estricta: sólo presentación, semántica, accesibilidad y responsive.
No cambian API, backend, base, permisos, estados ni recorridos. Un hallazgo que necesite uno de esos
cambios se documenta en otro ciclo antes de escribir ese código. La evidencia usa la instancia
desechable, Chromium y pruebas estables; producción y el defecto conocido de reproducción de
`prod-build.sh` permanecen fuera.

El responsable eligió 1A–14C y aprobó la propuesta sin decisiones abiertas. La implementación
encontró un único desbordamiento: el selector compartido de papeles en `/users` no cabía en 412 px
con los textos ingleses. `Conmutador` ahora limita su ancho y distribuye opciones en varias líneas.
La corrección pasó las tres resoluciones, ambos temas de fábrica y los papeles que usan esa lista.

La validación final pasó 220 pruebas unitarias del frontend y Playwright completo: 180 casos
aprobados y 38 omisiones previstas. Los ocho temas conservaron su medición de contraste y no quedan
hallazgos visuales o de accesibilidad abiertos. No cambiaron backend, API, datos, permisos ni flujos.

## 18. Bloque 7 implementado: coherencia documental y fronteras del frontend

### 18.1 Hallazgos que motivan el bloque

El producto y la ayuda de redacción del bloque anterior están implementados, pero el inventario
arquitectónico no describe todo el estado vigente de forma coherente:

- la sección 13 decía que una instalación funcionaba entera sin IA configurada, aunque la IA
  es obligatoria y el middleware limita una instalación sellada que no la tenga;
- presentaba como no automatizado el recorrido completo de `/setup`, pese a que la suite desechable ya
  lo recorre;
- conservaba como actuales cifras y fechas de pruebas anteriores a los bloques 5, 6 y a la ayuda de
  redacción;
- algunos registros históricos llamaban «pendiente» a una decisión reemplazada y cerrada;
- la sección 12 declaraba sin resolver el logging y la automatización de fronteras, aunque el
  responsable ya decidió ambos en el repaso de este bloque.

Estos hallazgos se corrigen de forma visible. No se reescribe la historia: se conservan decisiones y
enmiendas, se marca cuál las sustituyó y se eliminan afirmaciones que aparentan seguir vigentes.

### 18.2 Comprobador de fronteras

El frontend incorpora `scripts/check-boundaries.mjs`, propio y versionado, sin paquete adicional,
que revisa todos los
archivos TypeScript de `frontend/src/app`, incluidas sus pruebas. Ignora dependencias y artefactos
generados. Aplica estas direcciones:

1. la raíz de `app` puede componer `core`, `shared` y cualquier módulo;
2. cada carpeta de `modules/*` puede depender de sí misma, de `core` y de `shared`, pero no de otro
   módulo;
3. `core` puede depender de `shared`, pero no de un módulo;
4. `shared` no puede depender de `core` ni de un módulo.

Además, cada módulo sólo puede consumir su prefijo homónimo bajo `/api`: `mail` usa `/api/mail`;
`tickets`, `/api/tickets`; y `users`, `/api/users`. El comprobador reconoce literales de texto y
plantillas interpoladas. Las rutas transversales de `core` quedan fuera de esa correspondencia
homónima porque allí viven sesión, instalación y configuración.

Existe una lista explícita de excepciones, inicialmente vacía. Una excepción futura debe indicar
origen, destino y motivo, y cambiarla requerirá mantener este documento. Cada infracción informa
archivo, línea, regla incumplida y corrección esperada; se muestran todas en una ejecución y el
proceso termina con un código distinto de cero.

El comando puede ejecutarse por separado y forma parte tanto de `npm test` como de `npm run build`.
`scripts/check-boundaries.test.mjs` demuestra una importación cruzada, una dependencia de
capa invertida, una ruta API ajena, una plantilla interpolada y un conjunto válido.

### 18.3 Logging del navegador

No se añade servicio, dependencia ni proveedor de logging al frontend. Los errores funcionales se
presentan mediante los componentes y mensajes existentes. El único uso técnico de consola permitido
es el fallo fatal que impide arrancar Angular, en `main.ts`. El comprobador de fronteras no se
convierte en una regla general sobre llamadas a consola.

### 18.4 Auditoría documental

Se revisan todos los documentos con estado `as-built`, no sólo este archivo. La corrección:

- alinea la obligatoriedad de IA y el comportamiento de una instalación sin configurar;
- actualiza recorridos que ya están automatizados;
- sustituye inventarios volátiles de pruebas por el comando, la fecha y el resultado global de la
  última verificación pertinente;
- conserva los registros históricos, señalando con claridad las decisiones reemplazadas;
- actualiza `docs/README.md` y la tabla de correspondencia de `AGENTS.md`.

Si la auditoría descubre una diferencia nueva entre documentación y código que implique cambiar
comportamiento, se reportará y abrirá otro ciclo. No se corregirá como parte silenciosa de este
bloque.

### 18.5 Verificación y cierre

Antes de volver el documento a `as-built` deben pasar:

1. el comprobador por separado, incluidas sus pruebas negativas aisladas;
2. `npm test` y `npm run build`, demostrando que ambos ejecutan la frontera;
3. la suite completa del backend, porque el cierre describe el repositorio entero;
4. una búsqueda final de estados documentales y de las contradicciones registradas;
5. `git diff --check`.

La evidencia se registra con fecha, comandos y resultado global, sin mantener desgloses por paquete
o archivo que queden obsoletos con cada caso nuevo.

### 18.6 Alcance y decisiones

El bloque modificó `docs/`, `AGENTS.md`, los scripts del frontend y su `package.json`. No cambió API,
backend, base de datos, permisos, interfaz, flujos de usuario ni dependencias de producción.

Producción queda totalmente fuera: no se configura correo, no se abre el dominio, no se crea la
etiqueta `v1.0.0`, no se despliega y no se corrige todavía la reproducción desde cero de
`prod-build.sh`. Esos trabajos permanecen aparcados hasta el cierre de la versión 1.0.0.

El responsable eligió 1A–5A, 6A–10A y 11A–15A el 2026-10-08 y aprobó explícitamente el documento.
Durante la implementación apareció una dependencia `core → app` en la prueba del interceptor; quedó
registrada en la cabecera y se corrigió usando el router vacío que esa prueba realmente necesita.

La verificación del 2026-10-08 terminó correctamente: el comprobador y sus cuatro casos aislados;
`npm test -- --watch=false`, que ejecutó primero las fronteras y después las **225 pruebas** del
frontend; `npm run build`, también con la frontera previa; y `go test ./...` para todo el backend.
La auditoría final no encontró documentos en `propuesta` o `aprobado`, ni contradicciones vigentes de
las registradas en 18.1. `git diff --check` terminó limpio. No quedan decisiones abiertas.
