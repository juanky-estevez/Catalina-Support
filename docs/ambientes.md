# Ambientes: despliegue y pruebas

> **Estado:** as-built
> **Última actualización:** 2026-10-08
>
> **Enmienda implementada y verificada el 2026-10-08 (bloque 6).** La sección 17 define el entorno y la matriz de
> verificación del repaso integral de formato. Lo existente continúa as-built y producción permanece
> aparcada. El responsable aprobó explícitamente esta enmienda el 2026-10-08.
>
> **Hallazgo corregido y verificado el 2026-10-08.** La instancia real quedó detenida conservando sólo
> sus volúmenes aislados después de reproducir el rechazo 1.5B→3B. La sección 16 define las pruebas
> de la corrección 15A–22A y la reanudación con los GGUF ya verificados. No se toca desarrollo ni
> producción; la corrección fue aprobada explícitamente el 2026-10-08.
>
> **Hallazgo y enmienda propuesta el 2026-10-07 (bloque 5).** Las pruebas aisladas cubren el motor
> con procesos y GGUF falsos, pero falta validar el catálogo real sin alterar desarrollo. La sección
> 15 define un proyecto Compose desechable, mediciones reproducibles y limpieza completa. Las
> decisiones 1A–14A quedaron cerradas por el responsable. La propuesta fue aprobada explícitamente
> el 2026-10-08; lo existente continúa as-built mientras se ejecuta la validación.
>
> **Hallazgo y enmienda propuesta el 2026-10-07.** El bloque 4 necesita completar las pruebas
> aisladas prometidas para adaptadores, redirecciones, límites del contenedor y supervisión del
> motor local. Lo existente continúa as-built. La sección 14 recoge la corrección elegida por el
> responsable (4A). El repaso terminó con 5A, 6A, 7A y 8A, sin decisiones abiertas. El responsable
> aprobó explícitamente la enmienda, que se implementó y verificó el 2026-10-07.
>
> **Implementado y verificado el 2026-10-07.** Los proveedores y el administrador se probaron con
> servidores, procesos, archivos y límites falsos, sin proveedores comerciales ni GGUF reales. La
> pasada canónica terminó con 180 casos E2E aprobados y 38 omisiones previstas, en PC y móvil; la
> instancia se eliminó con sus volúmenes al cerrar.
>
> **Enmienda propuesta el 2026-10-06.** Lo existente continúa as-built. La sección 13 incorpora la
> IA como requisito de instalación, separa su motor local por entorno y añade las variables de
> cifrado y administración. El repaso quedó cerrado, la propuesta fue aprobada y la implementación
> fue verificada.
>
> **Enmendado el 2026-10-06**, aprobado explícitamente por el responsable, implementado y verificado:
> la suite aislada ejecuta su backend con `go run .`, conserva `./backend:/app:ro` y retira el montaje
> anidado `backend_tmp:/app/tmp`. El recorrido canónico terminó con salida 0 y la limpieza eliminó
> todos los recursos de la instancia.
>
> **Enmendado el 2026-10-04**, conforme a la sección 12 aprobada de `docs/prueba-local.md`:
> `dev.yml` ejecuta la migración automáticamente y `backend` espera a que termine correctamente.
> Una instalación local nueva queda lista con un solo `up -d --build`, sin datos de ejemplo; el
> seeder continúa separado y opcional. El paso de correo recibe las sugerencias editables de Mailpit.
>
> **Enmendado el 2026-10-03**, conforme a `docs/prueba-local.md` aprobado: El recorrido vigente es
> instalación vacía con `docker compose -f dev.yml up -d --build` y `docker compose -f dev.yml run
> --rm migrate`, asistente y admin/admin. El seeder es opcional: conserva las cinco tablas de
> ajustes de una instalación sellada; una instalación sin sellar recibe la demo local y queda
> sellada. Las pruebas usan exclusivamente `tests.yml`; los comandos anteriores sobre el entorno de
> desarrollo quedan sustituidos por el procedimiento de abajo.
>
> **Enmendado el 2026-10-02**: el README prioriza la instalación vacía, por corrección del
> responsable, con los ejemplos opcionales. Sólo el esquema no sella una base nueva: primero se
> completa `/setup` y después se entra con `admin`/`admin` en local; en producción la contraseña es
> `ADMIN_PASSWORD`. Los ejemplos restauran la dirección pública del dominio de desarrollo, que
> debe ajustarse para usar los enlaces en localhost. Se reporta el contexto de construcción del
> script de producción como bloqueo para reproducirlo desde cero; no se cambia código ni se despliega.
> Comprobado en una copia aislada de Linux: esquema sobre base nueva (0 cuentas y 0 tickets),
> los cinco pasos del asistente en navegador, prueba SMTP e IA, entrada con `admin`/`admin`,
> bloqueo 409 tras sellar, alta de cuenta y apertura del enlace del buzón. Después, el seeder dejó
> 11 cuentas de ejemplo, 25 tickets y 5 adjuntos; las entradas en navegador de fábrica, Usuario,
> Soporte y Desarrollo funcionaron. No se han ejecutado estas comprobaciones en
> Windows o macOS. Para el navegador del contenedor se usó la IP interna del frontend: el host
> `frontend` es rechazado por Angular. La suite predeterminada sigue usando el dominio de desarrollo.
>
> **Enmendado el 2026-10-01 (séptima vez)**: **los datos de ejemplo se siembran con un contenedor de
> un solo uso**, `docker compose -f dev.yml run --rm seed`, **igual en Linux, macOS y Windows** —es lo
> que deja fuera el guion de bash, que no se puede ejecutar en Windows, y con él la copia de los
> adjuntos—. El servicio vive en `dev.yml` **con perfil propio**, así que **`up -d` no lo arranca**:
> se pide a propósito, y **los ejemplos son opcionales** (el esquema sigue siendo obligatorio). El
> contenedor corre **el guion de siempre, `scripts/dev-seed.sh`** —una sola verdad, y el guion sigue
> valiendo en Linux y macOS fuera del contenedor—: el guion distingue con la marca `SEED_EN_CONTENEDOR`
> si habla con la base por la red del entorno con el `psql` de la imagen o entrando al contenedor con
> `docker compose exec`. Quedan al día las secciones 3.1, 3.3, 5 y 9.3. Comprobado el 2026-10-01 con
> **dos pasadas seguidas** del comando («Listo: 11 cuentas, 25 tickets y 5 adjuntos» las dos veces),
> con las once cuentas y los 25 tickets en la base, **los 5 adjuntos en `_files/`** y `up -d` sin
> arrancar el servicio.
>
> **Enmendado el 2026-10-01 (sexta vez)**: **el servidor de desarrollo de Angular reenvía `/api` al
> backend** (`frontend/proxy.conf.json`, declarado en `serve.options.proxyConfig` de
> `frontend/angular.json`), con destino **`http://backend:11002`** —el nombre de servicio del backend
> dentro de la red del entorno— y **sin reescribir la ruta**. Es una opción de `serve`: **el `build` de
> producción no cambia** y **en producción sigue reenviando nginx**. La sección 3.1 lo dice: **con el
> entorno levantado y el esquema aplicado se entra en `http://127.0.0.1:11001` con `admin`/`admin`**,
> sin nginx delante. Comprobado el 2026-10-01 por la 11001 (`GET /api/auth/methods` y `GET /api/health`
> → **200 JSON**; `POST /api/auth/login` con `admin`/`admin` → **200** con token) y con Playwright
> contra `http://127.0.0.1:11001`: **el recorrido de entrar y salir, 14 casos en verde**.
>
> **Enmendado el 2026-10-01 (quinta vez)**: **el esquema y los ejemplos se pueden aplicar sin `bash`**,
> con comandos de Docker que funcionan **igual en PowerShell, CMD y bash**: se copia el archivo al
> contenedor con `docker compose -f dev.yml cp` y se le da a `psql` con `-f` y `-v ON_ERROR_STOP=1`
> (sección 3.2, que era el único sitio que solo enseñaba la redirección `<`, de intérprete). El guion
> `./scripts/dev-seed.sh` **queda dicho como lo que es: de Linux y macOS**, y quien no tenga `bash`
> aplica sus dos archivos con esos dos comandos (secciones 3.1, 3.3 y 5). Los pasos del arranque en
> local, con el porqué de cada uno, viven en el `README.md`. Comprobado el 2026-10-01: en una base
> nueva (`catalina_support_prueba`) los dos comandos de la sección 3.2 dejan **18 tablas** y la
> aplicación responde (`GET /api/auth/methods` y `POST /api/auth/login` con `admin`/`admin`, **200**);
> el mismo archivo aplicado **dos veces seguidas** da `ERROR lines: 0` y termina en `COMMIT`, y la
> base de desarrollo, tras reaplicarlo, sigue con `12 usuarios, 25 tickets y 22 plantillas`.
>
> **Enmendado el 2026-10-01 (cuarta vez)**: queda dicho **qué aplica el esquema en desarrollo y
> cuándo**. La sección 3.3 dice ya lo que el guion siempre hizo y la documentación no contaba: **en
> una base nueva `./scripts/dev-seed.sh` es lo que crea el esquema** —su paso 1, `Aplicando el
> esquema…`, aplica `v1.0.0.sql`—, **no sólo los datos de ejemplo**, y se pasa **después de levantar
> el entorno y antes de abrir la aplicación**: la aplicación **no crea tablas al arrancar** y las
> migraciones no se aplican solas, así que abrirla antes deja `relación "installation_settings" does
> not exist` y `relación "ai_insights" does not exist`. La 3.1 lo apunta en una línea, la 5 añade el
> momento en su tabla y nombra el síntoma, y **los pasos del arranque en local no se repiten aquí**:
> viven en el `README.md`, «Probar en local: instalación vacía». Comprobado en la base de desarrollo: el
> seeder termina con `Listo: 11 cuentas, 25 tickets y 5 adjuntos` sobre una base que ya tiene el
> esquema, y `v1.0.0.sql` aplicado **dos veces seguidas** no da ningún error (`ERROR lines: 0` en las
> dos pasadas).
>
> **Enmendado el 2026-10-01 (tercera vez)**: el directorio de pruebas y Keycloak **dejan `dev.yml`** y
> pasan a **`active-directory.yml`** y **`keycloak.yml`**, cada uno con **su propio comando**, y **el
> perfil `auth` desaparece**. Los dos entran en la red **`catalina-support-dev`**, que **posee
> `dev.yml`** (nombre fijo, sin `external`), y por eso **primero se levanta el entorno** y después el
> servicio. Quedan al día la sección 3.3 (los dos servicios ya no viven en `dev.yml`), la 9.3 (los
> tres comandos y el comando nuevo de la suite, en dos pasos) y la tabla de casos (necesitan su
> servicio levantado, no un perfil).
>
> **Enmendado el 2026-10-01 (segunda vez)**: la sección 3.3 **deja de enumerar las once cuentas de
> ejemplo** —y de repetir su contraseña— y **apunta a la tabla del `README.md`**, que es donde se
> detallan con su **nombre, su correo, su rol y su contraseña** («Probar en local: instalación vacía»):
> el mismo dato en dos documentos se separa solo con el tiempo, así que vive en uno y el otro lo
> referencia en vez de copiarlo.
>
> **Enmendado el 2026-10-01**: los **tres elementos opcionales** se levantan cada uno con su propio
> comando y así queda dicho (sección 9.3): el motor de IA (`docker compose -f ai.yml up -d`, sección
> 3.4), el directorio de pruebas (`docker compose -f dev.yml up -d ldap`) y Keycloak
> (`docker compose -f dev.yml up -d keycloak`). El directorio y Keycloak **comparten el perfil `auth`**
> pero **pueden levantarse por separado** nombrando el servicio, **sin activar el perfil** —comprobado
> levantando sólo uno y viendo en `docker compose -f dev.yml ps` que el otro no aparece—; el comando de
> siempre `docker compose -f dev.yml --profile auth up -d` **sigue levantando los dos a la vez** y es el
> que usan la suite con el método en AD o Keycloak y las pruebas de interfaz. Los tres son opcionales.
>
> **Enmendado el 2026-09-30 (sexta vez)**: **el despliegue a producción queda aparcado** hasta que el
> producto esté terminado, así que **este runbook no se reescribe ahora**: la sección 4 lleva una **nota**
> diciendo que su flujo **cambió el 2026-09-30** —las rutas de `prod.yml` pasan a ser relativas al propio
> archivo, `scripts/prod-build.sh` publica el compose y `config/` en la carpeta de despliegue y ejecuta el
> compose desde ahí, y los puertos dejan de ir atados a `127.0.0.1`— y que **se pondrá al día cuando se
> retome el despliegue**. Lo que hay en la sección describe el despliegue **ya hecho** el 2026-09-25.
>
> **Enmendado el 2026-09-30 (quinta vez, y corregida el mismo día)**: **la red compartida con el motor
> de IA la crea el entorno**, y el motor **entra en ella**. Los contenedores de `dev.yml`, `prod.yml` y
> `ai.yml` entran en la red `catalina-support-ai`. Al principio estaba declarada **externa en los tres**
> —«alguien tiene que crearla»— y eso **rompía el arranque en una máquina nueva**: `docker compose -f
> dev.yml up -d` fallaba con «network catalina-support-ai declares as external, but could not be found»,
> y el README no lo decía. Se arregló con un guion, y **el responsable lo corrigió**: un `.sh` deja fuera
> a Windows.
>
> **Ahora no hay guion**: **`dev.yml` y `prod.yml` poseen la red** —con su nombre fijo y sin `external`—
> y **`ai.yml` la declara externa**, porque el motor se levanta **después**, con el entorno en marcha, y
> **no puede intentar recrearla** (si los tres se creyeran dueños, arrancar el motor con el entorno
> levantado falla: se probó y falló). Así `docker compose -f dev.yml up -d` **a secas funciona en una
> máquina nueva** —verificado bajando todo y levantando sólo el de desarrollo: la crea— y el motor se
> puede añadir después —verificado con el entorno en marcha: arranca sano y el backend lo ve—. Vale
> igual en Linux, macOS y Windows.
> **Enmendado el 2026-09-30 (cuarta vez)**: la migración **`v1.0.0.sql`** trae además **el sello de
> instalación** (`installation_settings.installed_at`) y **el correo saliente** (`smtp_*`), que pasa a
> vivir en la base **y deja de estar en el entorno**: las variables `SMTP_*` se retiran del entorno y de los
> archivos de ejemplo (corrección del responsable, 2026-09-30). **Una instalación que las tuviera puestas
> pone el correo una vez** en la vista de instalación o en Configuración. **El sello se rellena solo en las
> instalaciones que ya existían** —y sólo la primera vez que la columna nace—, así que una instalación
> en marcha **no ve el asistente** por actualizar, y una base recién creada **nace sin sellar**, que es
> «sin instalar»: al abrirla aparece `/setup` (`docs/primer-arranque.md`). En desarrollo **el archivo de
> ejemplos sella la instalación**, para que el asistente no salga en cada arranque ni en las pruebas.
>
> **Enmendado el 2026-09-29 (tercera vez)**: **`PUBLIC_APP_URL` deja de ser obligatoria**: la dirección
> pública de la instalación pasa a **Configuración** (`docs/modules/settings.md`, decisión 14) y la
> variable queda **como respaldo** para una instalación que ya la tuviera puesta. Una instalación nueva
> ya no necesita esa variable para arrancar; sin dirección configurada, **los correos con enlace fallan
> con su clave** en vez de mandar un enlace roto. Los dos campos —**la zona horaria** y **la dirección pública**— van **en la propia `v1.0.0.sql`**: la
> versión sigue abierta en la **1.0.0** y no se crea un archivo nuevo por esto.
>
> **Enmendado el 2026-09-28 (segunda vez)** — **decisión del responsable** al ver el entorno lleno de
> datos de prueba: **después de cada pasada de la capa de interfaz se ejecuta `./scripts/dev-seed.sh`**,
> para que desarrollo quede con las once cuentas y los 25 tickets de ejemplo y **nadie se encuentre
> tickets de prueba en las listas**. Los tickets que crea la suite **no se pueden borrar desde ella** —un
> ticket no se borra, y el contenedor de las pruebas no habla con la base—, así que el reinicio lo hace
> el seeder. **Se avisa antes de cada pasada**: el seeder **borra también lo que haya a mano en
> desarrollo**, y eso el responsable lo ha aceptado con el coste dicho.
>
> **Enmendado el 2026-09-28**: la capa de interfaz **apaga sus cuentas al terminar**
> (`tests/e2e/limpiar.ts`, el cierre de `playwright.config.ts`, sección 9.3). Cada caso da de alta las
> cuentas que necesita, y **se quedaban encendidas en desarrollo**: el responsable se encontró
> **seiscientas personas activas de soporte y desarrollo, todas llamadas «Prueba Automática»**, en el
> desplegable de «Asignar a» y en el buscador de a quién etiquetar. Ahora se apagan —**no se borran**,
> que es la regla de las cuentas— pidiendo **por la API**, que es lo que usa la aplicación: el
> contenedor de las pruebas no habla con la base. Los **tickets** que deja la pasada no se tocan —un
> ticket no se borra—: el entorno se deja como nuevo con `./scripts/dev-seed.sh`.
>
> **Enmendado el 2026-09-27 (segunda vez)**: la capa de interfaz **devuelve la instalación a cuentas
> locales antes de empezar** (`tests/e2e/preparar.ts`, la preparación previa de `playwright.config.ts`),
> para que una pasada cortada a medias con el método en AD o en Keycloak no haga fallar la siguiente por
> «credenciales incorrectas» (sección 9.3). Lo hace la **cuenta de fábrica**, que entra siempre.
>
> **Enmendado el 2026-09-27**, al entrar **el motor de IA** (`docs/modules/ai.md`): un contenedor nuevo,
> **en su propio compose (`ai.yml`)** y **compartido por desarrollo y producción**, con la red
> `catalina-support-ai` a la que entran los dos backends (sección 3.4). **No es obligatorio**: sin él,
> los dos campos que redacta se quedan sin texto y todo lo demás funciona igual. Y el archivo de datos
> de ejemplo **borra también los resúmenes**, que van por número de ticket y se pegarían a los tickets
> nuevos (sección 3.3).
>
> **Enmendado el 2026-09-26**, al comprobar la puesta al día de la migración con los
> adjuntos dentro del texto: la condición que la protege **no era la que hacía falta**. Decía «no
> empieza por `<`», y desde que el editor escribe un cuerpo **que empieza por su texto y lleva
> etiquetas dentro**, volver a aplicar el archivo **escapaba las referencias a los adjuntos y las
> dejaba como texto literal** (sección 5). Corregida a **«no tiene ni `<` ni `&`»**, que es lo que
> sólo cumple el texto plano de antes, y **comprobado aplicando el archivo entero dos veces** sobre
> desarrollo: `UPDATE 0` en las dos conversiones.
>
> **Enmendado el 2026-09-26**, al cambiar el responsable **el dominio de las direcciones de
> ejemplo**: las cuentas del seeder, las personas del directorio de pruebas y las del reino pasan de
> `@ejemplo.com` a **`@demo.com`** (sección 3.3). **Es un cambio de desarrollo y de nada más**: el
> seeder no se aplica en producción, las personas del directorio y del reino son de los servicios de
> pruebas, y **el remitente del correo no cambia** (en producción tiene que ser un dominio de la casa,
> con su SPF y su DKIM, o el correo no llega). Las direcciones que rellenan las pruebas de unidad no se
> tocan: ahí el dominio da igual.
>
> **Enmendado el 2026-09-25**, al hacer **el primer despliegue de producción**:
> `scripts/prod-build.sh` **está escrito** y es el que construye y publica (sección 4.2), los
> **contenedores de producción están levantados y verificados por dentro** —salud, marca, esquema
> aplicado **una sola vez** y la cuenta de fábrica entrando con la contraseña de `config/env/prod.env`—,
> y **el dominio sigue respondiendo 503** hasta que se abra (sección 4.5). **El correo saliente queda
> pendiente**, y con él el alta de cuentas: es lo que más falta. Además, la copia de la base de
> **producción** entra en el `cron` (sección 6) y se corrige **un fallo del guion de copias**: su rama
> de producción buscaba el `prod.yml` y el entorno entre los artefactos, y viven en el repositorio.
>
> **Pasa a as-built el 2026-09-25**, y **enmendado el mismo día**: el **directorio de la organización
> y Keycloak dejan de ser variables de entorno** y se configuran desde la pantalla —viven en la base
> y sus secretos van dentro de la copia de la base, sección 3.2—, y **los datos de ejemplo dejan
> puestos los dos caminos con el método en `local`** (sección 3.3). El despliegue a producción sigue
> siendo lo que falta de este documento.
>
> Aprobado por el responsable el 2026-09-22, tras repasarlo en forma de preguntas mientras se
> escribía. Fija el despliegue (un script con la misma forma que el de Calibyou), las migraciones
> —un archivo por versión, con tres dígitos—, las copias de seguridad y las tres capas de pruebas.
>
> Con este documento, **la cadena de documentación queda completa**: no hay más documentos
> previstos, y lo que falte se decide en el documento del área que corresponda.

## 1. Alcance de este documento

Cuenta **cómo se despliega y cómo se prueba**: el runbook de producción, las migraciones, las copias
de seguridad y las tres capas de pruebas.

**No describe la arquitectura ni los contenedores**: eso está en `docs/arquitectura.md` (secciones
2, 10 y 11, con los puertos, las imágenes y los vhosts). Aquí no se repite, y cuando algo cambie en
un sitio se cambia allí.

Está **aprobado**, así que habilita escribir código (Regla 0). La sección 10
recoge lo que he propuesto yo.

## 2. Los dos entornos de un vistazo

| | Desarrollo | Producción |
| --- | --- | --- |
| **Dominio** | `dev.catalina-support.example.com` | `support.example.com` |
| **Puertos** | frontend 11001, backend 11002, base 11003 | frontend 21001, backend 21002, base 21003 |
| **Código** | Montado como volumen: se edita y el contenedor recarga | **Artefactos construidos** en `/srv/catalina-support` |
| **Recarga** | Angular y `air` recargan al guardar | Hay que desplegar |
| **Base de datos** | Volumen de Compose, se puede vaciar sin pensar | Datos reales: **la migración se aplica una sola vez**, al cerrar la 1.0.0 |
| **Estado hoy** | **En pie y verificado** | **Aplazado** hasta cerrar la 1.0.0 (ver `docs/arquitectura.md`, sección 13) |

Mientras el despliegue esté aplazado, el dominio de producción responde 503 con un aviso de que está
en desarrollo. No es un fallo: es lo decidido.

## 3. Desarrollo

### 3.1 Levantar, mirar y parar

```bash
docker compose -f dev.yml up -d
docker compose -f dev.yml ps
docker compose -f dev.yml logs -f backend
docker compose -f dev.yml down          # los volúmenes se conservan
```

Se entra por `https://dev.catalina-support.example.com`. **Y también, sin nginx, por
`http://127.0.0.1:11001`**: el servidor de desarrollo reenvía `/api` al backend
(`frontend/proxy.conf.json` → `http://backend:11002`), así que con el entorno levantado **se abre la
11001**. `migrate` espera a PostgreSQL, aplica el esquema y `backend` espera su terminación correcta.
En una base nueva aparece `/setup`; después de completar el
asistente se entra con **`admin`/`admin`**. Los ejemplos son opcionales y sellan la instalación
(`README.md`, «Probar en local: instalación vacía»).
Los puertos 11001 y 11002 siguen publicados para depurar sin pasar por nginx.

**Levantar el entorno deja la base lista**: la aplicación no crea tablas por su cuenta; el servicio
de un solo uso `migrate` las aplica antes de que Compose permita iniciar `backend`. Si el SQL falla,
el backend queda detenido y el diagnóstico vive en `docker compose -f dev.yml logs migrate`. Los
datos de ejemplo también aplican el esquema en su primer paso, pero siguen siendo opcionales.

### 3.2 La base de datos

- El esquema se aplica **repetidas veces** sin miedo: `v1.0.0.sql` es transaccional e idempotente
  (`docs/arquitectura.md`, sección 7). **El archivo existe, está aplicado y crea 18 tablas.**
- `docker compose -f dev.yml up -d --build` ejecuta el servicio de un solo uso automáticamente:
  espera la salud de PostgreSQL, usa su cliente en Docker y termina con error si falla el SQL.
  `backend` depende de su finalización correcta.
- También se puede repetir manualmente; no borra datos, no carga ejemplos y no reabre el asistente:

```bash
docker compose -f dev.yml run --rm migrate
```

- En Linux y macOS, además, se puede hacer sin copiar, con la redirección del intérprete —que **no
  existe en PowerShell** y por eso no es la forma de arriba—:

```bash
docker compose -f dev.yml exec -T database \
  psql -v ON_ERROR_STOP=1 -U catalina_support -d catalina_support -p 11003 \
  < backend/migrations/v1.0.0.sql
```

- **Vaciar la base de desarrollo** es tan simple como borrar el volumen:

```bash
docker compose -f dev.yml down
docker volume rm catalina_support_dev_database
docker compose -f dev.yml up -d
```

El siguiente `up -d` vuelve a aplicar la migración automáticamente. Esto borra los datos y la
configuración local: hacerlo sólo si se quiere empezar otra instalación.

### 3.3 Los datos de ejemplo

**Los datos de ejemplo se aplican a propósito**, después de levantar el entorno. **No son un
requisito para entrar** —`up` ya deja el esquema listo, se completa el asistente y después se entra
con `admin`—, pero sí son lo que deja el entorno con contenido. La aplicación no crea las tablas: el
servicio `migrate` de
Compose lo hace y termina antes de que arranque el backend (sección 3.2). Los pasos del arranque en
local están en el `README.md`, «Probar en local: instalación vacía».

**La forma de los tres sistemas —Linux, macOS y Windows— es un comando de Docker**: un contenedor de
un solo uso que arranca, siembra y se va.

```bash
docker compose -f dev.yml run --rm seed
```

**No se levanta con `up -d`**, porque el servicio `seed` de `dev.yml` tiene **perfil propio**: se pide
a propósito, y **quien quiera el proyecto vacío —sólo con el esquema— no ejecuta este comando**.
**Los ejemplos son una elección; el esquema es obligatorio.**

El contenedor corre **el guion de siempre, `scripts/dev-seed.sh`** —una sola verdad sobre cómo se
siembra—: aplica el esquema, aplica `v1.0.0_dev.sql` y **copia los archivos de los adjuntos** a
`_files/`, que es el paso que un seeder SQL no puede hacer. Sus montajes son de **sólo lectura** salvo
`_files/`, que es lo único que el guion escribe en la máquina.

**En Linux y macOS el mismo guion sigue existiendo fuera del contenedor**, y hace lo mismo:

```bash
./scripts/dev-seed.sh
```

El guion habla con la base **entrando al contenedor con `docker compose exec`** cuando corre en la
máquina, y **con el `psql` de la imagen por la red del entorno** cuando corre dentro del contenedor:
lo distingue su marca `SEED_EN_CONTENEDOR`, y por eso no hay dos formas de sembrar que puedan
separarse.

Deja **once cuentas, 25 tickets y 5 adjuntos** —probado el 2026-10-01 con **dos pasadas seguidas** del
comando de Docker («Listo: 11 cuentas, 25 tickets y 5 adjuntos» las dos veces), comprobando además en
la base las once cuentas y los 25 tickets y en `_files/` los cinco archivos—. El servicio `migrate` de
la sección 3.2 aplica el esquema, pero **no copian los adjuntos**: para eso está el
contenedor, que es el paso 3 del guion.

Su **paso 1** (`Aplicando el esquema…`) aplica `backend/migrations/v1.0.0.sql` —transaccional e
idempotente, así que sobre una base que ya lo tiene se puede volver a pasar sin daño—, el **paso 2**
aplica `v1.0.0_dev.sql` y el **paso 3** **copia los archivos de los adjuntos** desde `config/seed/` a
`_files/`, en la carpeta de su ticket: un seeder SQL no puede crear archivos, y un adjunto sin su
archivo no se puede abrir.

Deja **once cuentas** —**la tabla con su nombre, su correo, su rol y su contraseña está en el
`README.md`**, en «Probar en local: instalación vacía», y **aquí no se repite**— y **25 tickets**
repartidos por todos los estados, con asignaciones, reasignaciones, comentarios de los tres papeles,
**siete con ticket interno** (uno esperando a Desarrollo, otro devuelto a Soporte, otros resueltos o
cerrados), adjuntos de verdad y **las fechas repartidas en los últimos tres meses**, para que la
bandeja no salga toda del mismo día.

**Las direcciones son `@demo.com`** (decisión del responsable, 2026-09-26), igual que las de las
personas del directorio de pruebas y del reino: en desarrollo todo es de mentira, y un dominio solo
—el mismo en los tres sitios— es lo que hace que no haya que pensar cuál toca.

**Las once entran con la misma contraseña**, la que dice la tabla del `README.md` (decisión del
responsable, 2026-09-25): en desarrollo se entra y se sale muchas veces al día, y una contraseña que
se escribe con una mano es lo que hace que probar no dé pereza.

- **Cumple la política del producto**, que pide **8 caracteres** como mínimo —bajados de 12 el
  2026-09-25—, así que el seeder **no se salta ninguna regla**: escribe el `bcrypt` ya hecho porque un
  archivo SQL no puede llamar a la aplicación, y el resultado es el mismo que si la contraseña se
  hubiera establecido desde la pantalla.
- **Las pruebas de interfaz no usan estas cuentas**: crean las suyas con la contraseña que fija
  `tests/e2e/ayudas.ts`. Cambiar la de las once no toca ninguna prueba.

**Borra para dejar un estado conocido**: los tickets que hubiera —también los creados a mano
probando—, las cuentas `@demo.com` y los adjuntos que no sean la marca. Es un seeder de
desarrollo, no una migración, y por eso se aplica cuando se quiere y no al arrancar.

**Sólo cuando la instalación no está sellada**, deja configurados los dos caminos de directorio
y activa el método `local` (enmienda aprobada del 2026-10-03): escribe la configuración del directorio apuntando al `ldap` que levanta
`active-directory.yml` y la de Keycloak apuntando al reino de pruebas, **dejando el método local**. Una instalación sellada conserva sus ajustes. Así
el entorno arranca entrando con las once cuentas de ejemplo y **cambiar de método es un clic en
Configuración**, que es justo lo que hay que poder probar a mano. Los dos servicios de pruebas viven
cada uno en su archivo (`active-directory.yml` y `keycloak.yml`): con ellos levantados, «Probar la
conexión» contesta que sí; sin ellos, contesta que no, que es la verdad.

- **La limpieza de las cuentas va en minúsculas** (`lower(email) LIKE '%@demo.com'`): el correo se
  guarda como lo escribe quien lo escribe, y sin eso una cuenta de ejemplo con mayúsculas sobrevivía
  a la limpieza y el entorno dejaba de estar en el estado conocido que promete este guion.

### 3.4 El motor local de IA: su propio compose y recursos por entorno

El motor que redacta el **motivo** y la **última acción** de los tickets (`docs/modules/ai.md`) **no
vive en `dev.yml` ni en `prod.yml`**, sino en **`ai.yml`**. Sólo se levanta cuando se elige la
modalidad local:

```bash
docker compose -f ai.yml up -d --build
docker compose -f ai.yml ps
docker compose -f ai.yml logs -f ai
docker compose -f ai.yml down       # el volumen del modelo se queda
```

**Por qué aparte**: el administrador descarga, activa y supervisa `llama-server` sin acceder al
socket de Docker. `dev.yml` crea `catalina-support-ai-dev` y `prod.yml` crea
`catalina-support-ai-prod`; `ai.yml` entra en la red correspondiente mediante `AI_NETWORK_NAME`.
Sus proyectos y volúmenes también llevan nombre por entorno (`AI_PROJECT_NAME` y
`AI_MODELS_VOLUME`), por lo que desarrollo y producción no comparten modelos ni secretos. El
backend lo alcanza por el nombre de servicio y el administrador no publica puertos hacia la
máquina.

- **El modelo se descarga desde `/setup` o `/settings`** al volumen del entorno; la descarga se puede
  reanudar y sólo `down -v` borra el volumen. La interfaz ofrece los modelos 1.5B, 3B y 7B con sus
  estimaciones de disco y RAM.
- **La IA es obligatoria para instalar**, pero la modalidad local no lo es: también se puede elegir
  un servidor propio o un proveedor externo. Si el motor configurado cae después, los resúmenes
  quedan pendientes y la mesa de ayuda continúa funcionando.
- **La red la crea el entorno**, así que **el motor se puede levantar después, con el entorno en
  marcha**: `docker compose -f dev.yml up -d` funciona en una máquina nueva sin crear nada a mano.
- La operación de producción sigue aparcada hasta el cierre de la versión 1.0.0. Cuando se retome,
  Compose se ejecutará con los nombres y secretos de producción documentados en la sección 13.

### 3.5 Dónde queda lo que no es código

| Carpeta | Qué guarda | ¿En git? |
| --- | --- | --- |
| `_logs` | Los archivos de `go-logs`, un archivo por día | La carpeta sí, los `.log` no |
| `_files` | Los adjuntos de los tickets | La carpeta sí, los archivos no |

Las dos son volúmenes montados en el backend: sobreviven a `down` y se copian desde la máquina sin
entrar al contenedor.

## 4. Producción

> **Nota (2026-09-30):** este flujo **cambió** —las rutas de los volúmenes de `prod.yml` pasaron a ser
> relativas al propio archivo, `scripts/prod-build.sh` publica el compose y `config/` en la carpeta de
> despliegue y el despliegue se ejecuta desde ahí, y los puertos ya no van atados a `127.0.0.1`—, pero
> **la sección no se pone al día todavía**: el despliegue a producción **queda aparcado** hasta que el
> producto esté terminado. Lo que se lee debajo describe el despliegue **ya hecho** el 2026-09-25, no el
> estado de hoy. Se pondrá al día cuando se retome el despliegue.

### 4.1 Cómo está montada

Contenedores «finos»: el runtime vive en imágenes base y **el código son artefactos construidos** en
`/srv/catalina-support` (frontend, backend, `_logs`, `_files`). Los `builder_*` de `prod.yml`
existen para construir y publicar, y **no se levantan con `up`**: están bajo el perfil `build`.

### 4.2 El despliegue, paso a paso

> **Hallazgo del 2026-10-02:** los pasos siguientes conservan el registro del despliegue histórico;
> no son hoy un recorrido reproducible desde cero. `prod-build.sh` usa el compose publicado en la
> carpeta de artefactos para construir, pero allí no copia las fuentes que piden los Dockerfiles.
> La ruta de salida sigue fija en `/srv/catalina-support`. La corrección de comportamiento
> requiere una propuesta aprobada y producción sigue aparcada.


Un script, `scripts/prod-build.sh`, con el mismo guion que el de Calibyou. **Está escrito y probado**
(el 2026-09-25 se construyeron y publicaron los dos artefactos con él). Lo que hace:

1. **Avisa si el árbol de trabajo tiene cambios sin commitear**: un artefacto construido desde un
   árbol sucio no corresponde a ningún commit, y eso hay que saberlo antes y no después.
2. **Rota el artefacto anterior** de cada componente a `<nombre>_prev`, **por copia y no por
   movimiento**, porque el directorio está montado en los contenedores: renombrarlo los dejaría
   apuntando al inodo viejo.
3. **Construye la imagen del constructor y publica** con
   `docker compose -f prod.yml --profile build build builder_<nombre>` y después
   `… --profile build run --rm builder_<nombre>`.
   - **Los dos pasos son obligatorios, y el primero es el que se olvida**: `run` sólo construye la
     imagen si **no existe**, así que a partir del segundo despliegue reutilizaría la de la vez
     anterior y **publicaría el artefacto viejo sin decir nada**. Se descubrió el 2026-09-26, al
     reemplazar los logos de fábrica: el artefacto publicado seguía trayendo los de antes, y el
     dominio cerrado (503) fue lo que evitó que se notara en producción.
4. **Verifica que el artefacto existe** antes de seguir: `frontend/index.html` y
   `backend/catalina-support` ejecutable. Si falta, corta sin desplegar.
5. **Copia los avisos de terceros**: `THIRD_PARTY_NOTICES.md` y `third_party_licenses/` quedan
   dentro de cada artefacto construido. Esta copia está implementada y su comprobación real queda
   para el repaso final de producción, que continúa aparcado.
6. **Escribe `BUILD_INFO`** en `/srv/catalina-support` con el commit, la rama, si el árbol
   estaba sucio, la fecha y qué artefactos se construyeron. Es la respuesta a «¿qué hay desplegado
   ahora mismo?» sin adivinar.
7. **Despliega**: `docker compose -f prod.yml up -d` y **reinicia el backend**. El frontend no se
   reinicia: el contenedor de nginx lee los estáticos del disco en cada petición.

```bash
./scripts/prod-build.sh                 # construye, publica y despliega
./scripts/prod-build.sh --no-deploy     # sólo construir y publicar
./scripts/prod-build.sh --only backend  # un solo componente
```

- **Un artefacto construido desde un árbol sucio no corresponde a ningún commit**, y por eso el
  script lo avisa **y lo escribe en `BUILD_INFO`**: el primer despliegue se hizo así, con el árbol
  sin commitear, y eso está dicho ahí para que nadie lo confunda con la versión publicada.
- **El `prod.yml` y el entorno de producción viven en el repositorio**, y lo que está en
  `/srv/catalina-support` es el código construido. Es lo que monta `prod.yml`, y es lo que
  tiene que saber cualquier guion que hable con producción (se corrigió así en
  `scripts/backup-db.sh`, sección 6).

### 4.3 Después de desplegar

1. **Aplicar las migraciones** que toquen (sección 5).
2. **Comprobar que arrancó**, en este orden y sin saltarse ninguno:

```bash
docker compose -f prod.yml ps                                  # los tres arriba y la base sana
curl -s https://support.example.com/api/health       # {"database":"ok","status":"ok"}
curl -sI https://support.example.com/ | head -1      # 200 y la CSP de producción
```

3. **Mirar los logs** del backend, que es donde `go-logs` cuenta lo que ha pasado:

```bash
docker compose -f prod.yml logs --tail=50 backend
tail -f /srv/catalina-support/_logs/catalina-support_prod_$(date +%Y%m%d).log
```

### 4.4 Volver atrás

Si el despliegue sale mal, **el artefacto anterior sigue ahí**: se copia `<nombre>_prev` sobre
`<nombre>`, se reinicia el backend y se vuelve a la versión de antes. Está automatizado en la
sección 4.2 (la rotación) y a mano en el peor caso:

```bash
cp -a /srv/catalina-support/backend_prev/. /srv/catalina-support/backend/
docker compose -f prod.yml restart backend
```

**Lo que no se puede deshacer con esto es una migración**: una vez aplicada, volver atrás exige
escribir otra que lo revierta. Por eso la migración se aplica **después** de comprobar que el código
nuevo arranca, y no antes.

### 4.5 La primera vez: hecho el 2026-09-25, y lo que falta

Es un despliegue normal, con tres cosas que sólo pasan una vez. **Dos están hechas y verificadas**:

| Paso | Estado |
| --- | --- |
| **1. La migración `v1.0.0.sql`, una sola vez** | **Hecho**: 13 tablas, las 20 plantillas de correo y la fila de configuración con el método de entrada en `local`, y **cero cuentas** —en producción no se siembra ninguna—. En desarrollo se ha aplicado muchas veces; en producción, esa y ninguna más |
| **2. La cuenta `admin` entra** con la contraseña de `config/env/prod.env` | **Hecho y comprobado**: entra con papel `administrador`, y con la contraseña equivocada contesta `auth.invalidCredentials`. La contraseña vive **sólo** en ese archivo, que no se versiona (`docs/usuarios-y-permisos.md`, sección 8) |
| **3. Dar de alta a Soporte y a Desarrollo** desde la aplicación | **Pendiente, y depende del correo**: el alta manda un enlace por correo, y **el correo de producción está sin configurar**. Mientras siga así, **nadie puede establecer su contraseña** y la única cuenta que entra es la de fábrica |

**El correo saliente es lo primero que hay que rellenar** en `config/env/prod.env`: el servidor, el
usuario, su contraseña y la dirección de remitente; después, reiniciar el backend. En cuanto haya
correo, el paso 3 se hace desde la aplicación en un minuto.

**El vhost y el certificado ya están puestos** desde el 2026-09-22, y **el dominio sigue respondiendo
503** a propósito: la 1.0.0 todavía no está cerrada y el responsable puede cambiar cosas antes de
cerrarla. **Abrirlo es un cambio de dos líneas** en `config/nginx/catalina-support-prod.conf` —se
sustituye el bloque del aviso por el `proxy_pass` al frontend, que está comentado justo debajo—,
copiarlo a `/etc/nginx/conf.d/` y recargar. Lo que se gana y lo que se pierde está claro: el dominio
pasa a enseñar la aplicación de verdad, con el correo todavía pendiente.

## 5. Las migraciones

**Un archivo por versión**, y el nombre es la versión que se publica: **tres dígitos**,
`v1.0.0.sql`, `v1.1.0.sql`, `v1.2.0.sql`… El archivo y la etiqueta de git que marca esa versión
llevan el mismo número, así que mirando el tag se sabe qué migración le toca.

**Y un archivo de datos de ejemplo por versión, con `_dev` detrás del nombre** (`v1.0.0_dev.sql`):
dónde van las cuentas y los tickets con los que se trabaja en desarrollo. **Ese archivo no se aplica
nunca en producción**, y por eso lleva el sufijo bien visible.

| Momento | Qué se aplica |
| --- | --- |
| **En desarrollo, al empezar** | `docker compose -f dev.yml up -d --build` aplica `v1.0.0.sql` automáticamente antes del backend. `run --rm migrate` permite repetirlo. Los ejemplos son **aparte** (sección 3.3) |
| **Hasta la 1.0.0** | `backend/migrations/v1.0.0.sql`, todas las veces que haga falta en desarrollo |
| **Al cerrar la 1.0.0** | El mismo archivo, **una sola vez** en producción |
| **Después de la 1.0.0** | Un archivo por versión, aplicados **en orden**, sin saltarse ninguno |

**Un archivo por versión y no uno por cambio**: el archivo cuenta la historia completa de esa
versión, se revisa de una vez y se aplica de una vez. Un archivo por cada cambio menudea el
despliegue y llena `migrations/` de retales.

**Cada versión trae su archivo de esquema y, si hace falta, su archivo de ejemplos**, y cada uno dice para qué es:

| Archivo | Qué lleva | Dónde se aplica |
| --- | --- | --- |
| `v1.0.0.sql` | Las tablas, los índices y sus restricciones **y lo que la aplicación necesita para funcionar**: las dos filas de configuración —la de la instalación y la de los tickets— y las veinte plantillas de correo. Sin la fila de configuración no se puede leer ni el idioma ni el nombre; sin plantilla no sale ningún correo | **Desarrollo y producción** |
| `v1.0.0_dev.sql` | Los **datos de ejemplo**: once cuentas (cinco que piden, tres de Soporte y tres de Desarrollo), 25 tickets con su historia —asignaciones, reasignaciones, comentarios, escalados con su interno, adjuntos— y el contador de la numeración | **Sólo desarrollo**, con `docker compose -f dev.yml run --rm seed` (`scripts/dev-seed.sh` en Linux y macOS) |

**El guion borra también los resúmenes del motor de IA** (`ai_insights`): la tabla se lleva **por número
de ticket**, y los datos de ejemplo **vuelven a emitir los mismos números**, así que una fila vieja se
pegaría a un ticket nuevo y enseñaría el motivo de otro (`docs/modules/ai.md`, sección 4).

**La separación no es de forma**: lo que hay en el `_dev` **borra y vuelve a crear** el contenido de
trabajo —tickets incluidos—, así que no puede ser algo que se aplique solo por desplegar. En
desarrollo se aplica cuando se quiere volver a un entorno conocido.

**Una migración también puede cambiar datos, y esta los cambia**: el 2026-09-26 el texto de los
tickets y de los comentarios **pasó de texto plano a HTML** (`docs/modules/tickets.md`, sección 2.3),
así que `v1.0.0.sql` **convierte lo que ya existe** en su sección de puesta al día: escapa los
caracteres que el HTML se come (`&`, `<`, `>`), cambia los saltos de línea por `<br>` y **envuelve el
resultado en un párrafo** (`<p>…</p>`).
   - **La condición es «no tiene ni `<` ni `&`», y no «no empieza por `<`»** (corregido el
     2026-09-26, en el mismo trabajo que metió los adjuntos dentro del texto). Con la condición
     vieja —«este texto no es HTML todavía», preguntado por su primera letra—, un cuerpo escrito por
     el editor **empezaba por su texto y llevaba etiquetas dentro** («Mira lo que me
     sale.<img data-adjunto="captura.png">»), así que volvía a entrar, se escapaba entero y las
     referencias a los adjuntos quedaban como texto literal: **aplicar el archivo dos veces rompía
     lo que el editor escribe**, que es lo contrario de lo que esta sección promete. Ahora sólo se
     convierte lo que puede ser texto plano de antes, porque un texto del editor **siempre lleva `&`
     o `<`** —escapa el `&` y el `<` al guardarlos, y mete etiquetas en cuanto hay formato o un
     adjunto— y un texto ya convertido empieza por `<p>`.
     - Lo que la condición deja fuera, y se dice: un texto antiguo que **llevara un `<` o un `&` a
       mano** no se toca y se lee sin formato. **El `<p>` sigue puesto** —es cómo se lee un texto como
       un párrafo—, pero ya no es lo que sostiene la idempotencia.
     - **Comprobado aplicando el archivo entero dos veces** sobre desarrollo: la segunda pasada deja
       las dos conversiones en **`UPDATE 0`**, y un cuerpo con una imagen dentro —el que escribe el
       editor— **no se toca**. Es la comprobación que hay que repetir si esta sección se vuelve a
       tocar.
   Es una
conversión **de una sola dirección** —el texto se lee igual que antes, ahora con formato—, se aplica
sola al aplicar el archivo y **en producción no toca nada**, porque allí no hay todavía ningún ticket.
Lo que hay que tener presente es que **las copias anteriores a ese día tienen el texto plano**: una
restauración vieja se vería igual, sin formato, y hay que volver a aplicar la migración para ponerla al
día.

**Sin tabla de control**: no hace falta, porque los archivos son idempotentes. Aplicar dos veces el
mismo no rompe nada, así que la forma de trabajar es sencilla: **se aplican todos los archivos hasta
la versión publicada**, en orden, y los que ya estaban aplicados no hacen nada.

Y para saber qué versión está publicada no hay que adivinar: lo dicen **la etiqueta de git y el
`BUILD_INFO`** del despliegue (sección 4.2). Aplicar de más es inofensivo; aplicar de menos se ve
enseguida, porque la aplicación pedirá una tabla que no existe —`installation_settings`, `ai_insights`…—
y la base responderá `relación "…" does not exist` (sección 3.3).

Nunca `AutoMigrate`, nunca un `ALTER` a mano, nunca un paso que no esté en un archivo del
repositorio.

## 6. Las copias de seguridad

**Lo que no se puede reconstruir es la base de datos**, así que es lo que se copia solo: un volcado
al día, con **catorce días de retención**, en el `cron` del servidor.

```bash
./scripts/backup-db.sh            # el entorno de desarrollo
./scripts/backup-db.sh prod       # producción
```

Lo que hace, en este orden (y si algo falla, **no deja una copia a medias**: borra lo que haya
escrito y sale con error):

1. **Comprueba que la base está levantada.** Sin ella no hay nada que copiar.
2. **Vuelca la base** con `pg_dump -Fc` **dentro del contenedor** —el formato comprimido de
   PostgreSQL, que se restaura con `pg_restore`— y lo deja en la carpeta de copias con la fecha en el
   nombre: `/srv/catalina-support/backups/catalina_support_20260925_033000.dump`.
3. **Comprueba que el volcado se puede leer** (`pg_restore --list`). Una copia que no se puede
   restaurar no es una copia, y eso hay que saberlo hoy y no el día que haga falta.
4. **Borra las copias de más de catorce días** y dice cuántas quedan.

**No copia `_files`**, y es a propósito (corrección del responsable, 2026-09-25): **los adjuntos de
los tickets y los documentos que sube la gente no se borran nunca**, y copiarlos es cosa de soporte
manual cuando haga falta. El script se ocupa de la base y de nada más. Cuando haya que llevarse los
archivos —a otro servidor, a un disco, a una revisión—, se copian desde la máquina como cualquier
carpeta, respetando lo que hay dentro:

```bash
cp -a /srv/catalina-support/source/_files/. /srv/catalina-support/backups/_files/
```

**Y un aviso que conviene tener delante el día de una restauración de verdad**: en la base van las
**filas** de los adjuntos, y los archivos van aparte. Restaurar sólo la base deja una instalación con
adjuntos que **no se pueden abrir**, porque su archivo no está. Cuando haya que volver atrás de
verdad, se restauran las dos cosas: la base **y** `_files`.

**Y el segundo aviso, confirmado por el responsable el 2026-09-25**: desde que el directorio de la
organización y Keycloak se configuran desde la pantalla, **sus secretos —la contraseña de la cuenta de
servicio y el secreto del cliente— viven en la base**, así que **una copia de la base es un secreto
más**: quien la tenga puede hablar con el directorio y con el reino. Es el precio de poder cambiarlos
sin entrar por SSH, y por eso las copias viven donde viven, con los permisos de la máquina y sin salir
de ella.

**Dónde viven las copias**: `/srv/catalina-support/backups/`, fuera del repositorio y con las
copias de los otros proyectos de la máquina. El script escribe lo que hace por la salida estándar, y
el `cron` lo recoge en `copias.log`, ahí mismo:

```cron
30 3 * * * /srv/catalina-support/source/scripts/backup-db.sh dev  >> /srv/catalina-support/backups/copias.log 2>&1
45 3 * * * /srv/catalina-support/source/scripts/backup-db.sh prod >> /srv/catalina-support/backups/copias.log 2>&1
```

**Las dos bases se copian**, y a horas distintas para que no compitan: la de desarrollo a las 3:30 y
la de **producción a las 3:45**. La de producción se copió y se comprobó el 2026-09-25 —el volcado se
puede leer y lleva sus trece tablas—, y desde ese día está en el `cron`.

**Las copias de las dos bases se llaman igual**, porque la base se llama igual en los dos entornos
(`catalina_support`): lo que las distingue es la fecha y la hora del nombre. Si algún día hay que
saber de cuál es una copia, se mira el contenido —la de producción no tiene cuentas y la de desarrollo
tiene las once de ejemplo—. Queda dicho aquí **en vez de inventar un nombre distinto**: renombrar el
archivo obligaría a tocar también la restauración, y no aporta nada que el contenido no diga.

**Restaurar**, cuando haga falta: se crea una base y se vuelca dentro. Con `--no-owner`, porque el
dueño de las tablas de una copia no tiene por qué ser el de la base donde se restaura.

```bash
docker compose -f dev.yml exec database psql -U catalina_support -d postgres -p 11003 \
  -c "CREATE DATABASE restauracion"
docker compose -f dev.yml exec database pg_restore -U catalina_support -d restauracion -p 11003 \
  --no-owner < /srv/catalina-support/backups/catalina_support_20260925_033000.dump
```

**Lo que está probado y lo que no.** El 2026-09-25 se comprobó, contra el entorno de desarrollo: que
el script deja su copia, que la copia **se restaura** en una base nueva —**11 tablas y los mismos
registros** que la de verdad, contados uno a uno— y que la retención borra lo que pasa de catorce
días sin tocar lo demás. **Lo que no está probado es en producción**, porque no hay producción: el
script tiene su rama para `prod` escrita y sin estrenar, y el día del despliegue se prueba allí.

## 7. Las variables de entorno

Cada grupo lo documenta quien lo usa, y aquí sólo se dice dónde vive:

| Grupo | Quién lo documenta |
| --- | --- |
| `ENVIRONMENT`, `PROJECT_NAME`, `LOGS_FOLDER`, `TIMEZONE` (los de `go-logs`) | `docs/arquitectura.md`, sección 8 |
| `APP_PORT`, `POSTGRES_HOST`, `PGPORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | `docs/arquitectura.md`, sección 10 |
| `FILES_PATH` | `docs/modules/tickets.md`, sección 2.3 |
| `ADMIN_PASSWORD` | `docs/usuarios-y-permisos.md`, sección 8 |
| `TOKEN_SECRET` | `docs/usuarios-y-permisos.md`, sección 6 |
| `KEYCLOAK_ADMIN_PASSWORD` | La contraseña del administrador del **Keycloak de desarrollo** (sólo lo lee `dev.yml`, no el backend) |

**Y dos cosas que dejaron de ser variables de entorno el 2026-09-25**: el **directorio de la
organización** y **Keycloak** (`LDAP_*` y `OIDC_*`). Ahora viven en la base —`directory_settings` y
`keycloak_settings`— y se configuran desde la pantalla de Configuración, con su botón de «Probar la
conexión» (`docs/modules/settings.md`, sección 5.8). **No queda ninguna por entorno para esos dos
caminos**: si se ven `LDAP_*` u `OIDC_*` en un archivo de entorno, son restos de una versión anterior
y no las lee nadie. Y una consecuencia que hay que tener presente: **los secretos de los dos caminos
van dentro de la copia de la base**, así que esa copia es un secreto más (sección 6).

**Y el correo saliente tampoco es una variable de entorno**: vive en `installation_settings.smtp_*` y
se configura en la vista de primer arranque o en Configuración; las variables `SMTP_*` se retiraron
(`docs/arquitectura.md`, sección 9). En desarrollo lo deja puesto
`backend/migrations/v1.0.0_dev.sql`, apuntando al buzón de pruebas del `dev.yml` (`mail`, web en
`127.0.0.1:11004`).

Reglas que valen para los dos entornos:

- **Sólo lo que cambia por entorno es variable.** Lo que vale lo mismo en los dos se escribe como
  constante en el código.
- **`config/env/dev.env` se versiona** porque no tiene secretos reales: son credenciales de
  contenedores locales.
- **`config/env/prod.env` no se versiona**, y ninguno de sus valores aparece en la documentación ni
  en el código. Hay una plantilla (`prod.env.example`) con la lista de lo que hace falta.
- **Si una variable deja de tener consumidor, se borra**: del código, de los env, de los compose y de
  la documentación, en el mismo cambio.
- **Añadir una variable obligatoria no basta con escribirla en el env**: los contenedores leen su
  entorno **al crearse**, así que uno que lleve horas levantado no la tiene y el backend se negará a
  arrancar —que es lo que debe hacer—. Hay que recrear el servicio, no sólo reiniciarlo:

```bash
docker compose -f dev.yml up -d backend      # recrea el contenedor con el entorno nuevo
```

  Comprobado el 2026-09-23 al añadir `TOKEN_SECRET`: con el contenedor viejo, el backend respondía
  «no se puede firmar la sesión: falta TOKEN_SECRET» y nginx devolvía 502 hasta recrearlo.
- **Lo mismo vale para las dependencias del frontend**: el servidor de desarrollo de Angular
  **cachea el resultado de Tailwind**, así que uno que estuviera levantado antes de instalar
  `tailwindcss` sigue sirviendo la hoja **sin ninguna utilidad generada** —la página se ve como
  texto sin estilos— hasta que se reinicia. Pasó el 2026-09-23, con el servidor llevando un día en
  pie. Si la interfaz se ve sin estilos después de tocar dependencias o la configuración de
  Tailwind:

```bash
docker compose -f dev.yml restart frontend
```

## 8. nginx y los certificados

- Los vhosts son **copias** de `config/nginx/` en `/etc/nginx/conf.d/`: si se cambia uno en el
  repositorio, hay que volver a copiarlo y recargar (los comandos están en `AGENTS.md`).
- Los certificados se **renuevan solos** (certbot con el plugin `webroot`, un certificado por
  dominio). La renovación no necesita intervención porque los dos vhosts sirven
  `/.well-known/acme-challenge/` en el puerto 80.
- **Comprobar que sigue vivo**, de vez en cuando y siempre antes de un despliegue:

```bash
sudo certbot certificates | grep -A2 catalina
echo | openssl s_client -connect support.example.com:443 \
  -servername support.example.com 2>/dev/null | openssl x509 -noout -dates
```

## 9. Las pruebas

Tres capas, de la más barata a la más lenta. Las tres se ejecutan **en los contenedores**.

### 9.1 Backend: `go test ./...`

Lo que se prueba, por orden de importancia:

1. **La máquina de estados de los tickets**: qué transiciones existen, quién puede cada una y, sobre
   todo, **las ocho reglas de sincronización** (`docs/propósito-y-alcance.md`). Es la lógica más
   delicada del sistema y la que peor se ve a simple vista: un ticket que se queda en `escalado` o
   un principal que no vuelve a Soporte no salta hasta que alguien lo mira.
2. **La numeración**: que dos tickets creados a la vez no saquen el mismo número, y que el número sea
   inmutable aunque cambie el prefijo.
3. **Los permisos**: que cada papel sólo pueda lo suyo, probado en los servicios y no sólo en la
   interfaz.
4. **La resolución del origen de una cuenta** al entrar: `local` contra la base, `ad` contra el
   directorio, y las reglas de convivencia (`docs/usuarios-y-permisos.md`, sección 5).

```bash
docker compose -f dev.yml exec backend go test ./...
docker compose -f dev.yml exec backend go vet ./...
```

### 9.2 Frontend: `npm test`

Vitest, que es lo que trae Angular 22. Para lo que tiene lógica: servicios, guardas, y los
componentes que calculan algo. **No se prueba el aspecto**: para eso están las pruebas de 8.3.

```bash
docker compose -f dev.yml exec frontend npm test -- --watch=false
```

### 9.3 Interfaz: Playwright en una instalación desechable

Desde la raíz, sin arrancar ni modificar el entorno de desarrollo:

```text
docker compose -f tests.yml --profile directory down -v
docker compose -f tests.yml up -d --build database backend frontend mail
docker compose -f tests.yml run --rm e2e
docker compose -f tests.yml --profile directory down -v
```

Para incluir AD y Keycloak, sustituir el segundo comando por:

```text
docker compose -f tests.yml --profile directory up -d --build database backend frontend mail ldap keycloak
```

El runner conserva el código de salida de Playwright. Tras un fallo también se ejecuta la limpieza
final; las capturas, trazas e informe permanecen en `tests/e2e/resultados/`. Cada pasada comienza
retirando exclusivamente los volúmenes de este proyecto de pruebas, también tras una interrupción.
No se cargan seeders: las cuentas y fixtures se crean por las APIs normales de la aplicación.

El origen de Chromium es `http://frontend.localhost:11001`: Docker declara el alias y el navegador
lo resuelve al servicio `frontend` con `--host-resolver-rules`. El sufijo localhost habilita el
portapapeles real sin modificar la seguridad de la aplicación. La IA permanece ausente por diseño.

La suite cubre cuentas locales, correo, configuración, permisos y tickets en PC y móvil.
Los directorios se prueban una sola vez porque el protocolo no depende del ancho. Los casos de
IA requieren un motor y se omiten en este entorno sin IA. La preparación verifica además el
asistente antes de sellar, con capturas de ambos anchos. El historial de los hallazgos anteriores
queda en las enmiendas de la cabecera; sus comandos contra desarrollo ya no son el procedimiento.

Verificado en Linux el 2026-10-03: suite con directorios, 193 aprobados, 20 omitidos y un
acceso cancelado durante la recarga del frontend (salida 1, informe conservado). En una base
nueva, sin directorios y con los archivos estables, ese caso pasó en PC y móvil: 2 aprobados
y 24 omitidos por servicios ausentes, salida 0. La detección de LDAP ausente se corrigió al
encontrar que su función de comprobación no se llamaba. Go y vet pasan; frontend: 216 pruebas.
No se han ejecutado estas comprobaciones en Windows o macOS.

### 9.4 Arranque del backend aislado sin montaje anidado

La enmienda as-built de `docs/prueba-local.md`, sección 13, conserva los comandos de esta sección.
Sólo cambia cómo arranca internamente el servicio `backend` de `tests.yml`: `go run .` en lugar del
comando de Air de la imagen, sin `backend_tmp`. Esto evita escribir dentro del bind de código de sólo
lectura y no cambia la preparación, los perfiles, los informes ni la limpieza.

La verificación parte siempre de `down -v`, reconstruye los servicios, espera la salud del backend,
ejecuta al menos los recorridos afectados y termina con otro `down -v`. El compose final debe
funcionar directamente, sin archivos de sobreescritura temporales.

El repaso confirmó que no habrá recarga, volúmenes temporales ni cachés adicionales. No quedan
decisiones abiertas propias del runbook.

Verificado en Linux el 2026-10-06 con el compose definitivo: desde una limpieza completa, los
servicios `database`, `migrate`, `backend`, `frontend` y `mail` arrancaron sin el error OCI; la suite
sin perfiles opcionales terminó con 177 casos aprobados y 39 omitidos, salida 0, y el `down -v` final
eliminó contenedores, red y volúmenes. `docker compose -f tests.yml ps -a` quedó vacío.

## 10. Lo que se decidió al repasar este documento

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **Las tres capas de pruebas** | Sí: unitarias de backend, unitarias de frontend y los cinco recorridos con **Playwright** contra desarrollo |
| 2 | **La base de las pruebas del backend** | Una base aparte, `catalina_support_test`, en el mismo contenedor: las pruebas la vacían y no tocan los datos de desarrollo |
| 3 | **Copias de seguridad** | `pg_dump` diario, con retención (14 días) y un script en `scripts/`, programado en el `cron` del servidor. **Corrección del responsable (2026-09-25)**: el script **no copia `_files`** —los adjuntos y los documentos se copian a mano cuando Soporte lo pida—, y **los adjuntos no se borran nunca**. Está en la sección 6 |
| 4 | **Integración continua** | **Ninguna** en la 1.0.0: las pruebas se ejecutan a mano en los contenedores, como parte del trabajo |
| 5 | **Las migraciones** | **Un archivo por versión**, con tres dígitos (`v1.0.0.sql`, `v1.1.0.sql`…), aplicados **en orden** y **sin tabla de control**: son idempotentes y se aplican todos hasta la versión publicada |
| 6 | **Los datos de ejemplo, aparte** | **Decisión del responsable, 2026-09-25**: la 1.0.0 se cierra con **dos archivos** —`v1.0.0.sql`, con las tablas y lo que la aplicación necesita para funcionar en los dos entornos, y `v1.0.0_dev.sql`, con once cuentas y 25 tickets que **sólo se aplican en desarrollo**—. Lo que la aplicación necesita para arrancar (las dos filas de configuración y las veinte plantillas de correo) **no es un dato de ejemplo** y se queda en el primero, porque en producción no hay quien lo siembre por otra vía |
| 7 | **La limpieza de los logs** | **Manual**: `go-logs` no rota y no se le añade un script |

## 11. Qué NO entra en la 1.0.0

- **Integración continua y despliegue automático**: todo a mano, a propósito.
- **Un entorno de staging**: desarrollo y producción, y nada en medio.
- **Monitorización y alertas**: nadie avisa de que la web se ha caído. Enterarse es cosa de mirar.
- **Alta disponibilidad**: un servidor, sin réplicas ni balanceo.
- **Despliegue sin cortar el servicio**: reiniciar el backend corta lo que esté en curso, y se acepta.
- **Copias de seguridad fuera del servidor**: el script las deja en el servidor; sacarlas es aparte.
- **Rotación automática de logs** (ni por tamaño ni por antigüedad): la limpieza es manual, y la
  carpeta `_logs` crece un archivo por día. Está escrito para que no sorprenda.
- **Envío de logs a un servidor centralizado.**

## 12. Qué habilita este documento

Con `ambientes.md` aprobado, **la cadena de documentos está completa** y se puede empezar a
implementar, en el orden que fija `docs/modules/tickets.md`: `auth`, `users` y `tickets`, con sus pruebas. El
`scripts/prod-build.sh` y el de copias se escriben entonces, no antes.

## 13. Operación implementada para la IA obligatoria

### 13.1 Variables y secretos

- `AI_CREDENTIAL_KEY`: 32 bytes aleatorios codificados en base64, distintos de `TOKEN_SECRET` y
  obligatorios al arrancar el backend. Desarrollo tiene un valor local versionado. Cada producción
  genera el suyo, por ejemplo con `openssl rand -base64 32`, y lo guarda en `config/env/prod.env`.
- `AI_MANAGER_TOKEN`: secreto compartido entre backend y administrador local. Es obligatorio cuando
  se usa la modalidad local y no llega al navegador.
- `AI_URL` y `AI_MODEL` dejan de ser respaldo de una configuración vacía: la base probada es la
  fuente única. `AI_PALABRAS` y el tiempo máximo pueden conservarse como límites operativos.

La ausencia o formato inválido de `AI_CREDENTIAL_KEY` impide arrancar el backend. Una clave válida
que no descifra una credencial existente deja la instalación en «IA requiere configuración» para
que la cuenta de fábrica la reemplace. Las copias de PostgreSQL llevan el cifrado, pero no la clave
maestra; ambas piezas se respaldan por separado.

### 13.2 Motor local por entorno

`ai.yml` se parametriza con el proyecto y la red del entorno, sin nombres globales compartidos.
Desarrollo y producción tienen volúmenes de modelos distintos. Levantarlo no descarga nada hasta que
`/setup` o `/settings` lo solicite; el 1.5B aparece seleccionado por defecto. En una máquina que no
puede mantener ambos motores, el runbook indica apagar uno antes de levantar el otro.

La operación habitual sigue siendo Compose: levantar, ver logs y bajar sin borrar modelos. La
interfaz administra archivos dentro del volumen, no contenedores. Una descarga incompleta queda con
nombre parcial, se reanuda y nunca se activa. Borrar el volumen sigue siendo una acción manual y
destructiva fuera de la aplicación.

### 13.3 Arranque nuevo y actualización

La aplicación principal puede arrancar antes que el motor local para mostrar `/setup`. Si se elige
local, el administrador debe estar accesible y completar descarga/activación antes de sellar. Si se
elige proveedor o servidor propio, `ai.yml` no se levanta.

En una actualización, la migración conserva una configuración existente que supere la generación de
prueba. Una instalación sin ella queda restringida hasta que Administrador o cuenta de fábrica la
configure; no se inventa una credencial ni se presupone que el contenedor local existe.

### 13.4 Pruebas

Las pruebas unitarias y de integración usan servidores HTTP falsos para los proveedores y archivos
GGUF diminutos de prueba para el administrador. Playwright ofrece un servicio falso dentro de
`tests.yml`, de modo que `/setup` pueda completar obligatoriamente el paso de IA sin internet,
credenciales comerciales ni descargas de gigabytes. También prueba caída y recuperación.

La verificación manual del catálogo real descarga y mide cada modelo por separado y registra tamaño,
RAM, tiempo de carga y generación en `docs/modules/ai.md` antes de cerrar como as-built. No se prueba
contra cuentas comerciales del proyecto.

La operación local recupera trabajos acumulados automáticamente. Un proveedor comercial requiere
que el Administrador confirme cada lote de recuperación después de ver su volumen estimado; las
pruebas verifican que ninguna llamada del lote ocurra antes de esa confirmación.

### 13.5 Fuera de alcance

No se automatizan compra de créditos, límites de gasto, GPU, copias del volumen de modelos ni alta
disponibilidad del motor. Los GGUF pueden volver a descargarse; el backup obligatorio sigue centrado
en PostgreSQL y los adjuntos. El responsable decidió este alcance el 2026-10-06. El repaso cerró la
generación y recuperación comercial, la clave obligatoria y la aceptación de licencias; no quedan
decisiones abiertas; propuesta aprobada explícitamente.

## 14. Verificación aislada implementada para el bloque 4

Las pruebas unitarias levantan servidores `httptest` distintos para OpenAI, DeepSeek, compatible con
OpenAI y Claude. Verifican ruta, cuerpo, autenticación, idioma, respuesta válida, respuesta demasiado
grande y rechazo de redirecciones. No usan DNS público, internet, tokens ni cuentas comerciales.

El administrador local admite sustitutos sólo dentro de sus pruebas para ejecutable, reloj/espera,
memoria y descarga. Así se verifican reinicio con espera creciente, cancelación al cambiar de modelo,
cgroup v1/v2, disco restante, parcial inválido, reanudación, checksum y límite de redirecciones. Los
GGUF de prueba son mínimos y temporales.

La prueba del supervisor usa esperas sustituibles y comprueba la serie 1, 2, 4… 60 segundos sin
esperar tiempo real. La descarga falsa permite hasta cinco redirecciones HTTP/HTTPS, rechaza la sexta
y cualquier otro esquema. Un archivo con tamaño correcto y checksum incorrecto se elimina antes del
siguiente intento.

`tests.yml` conserva su servidor compatible falso para `/setup`. Playwright no simula muerte de
procesos ni descarga modelos reales; esas condiciones quedan en pruebas Go deterministas. El cierre
ejecuta backend completo, `ai-manager`, las pruebas del frontend y Playwright en PC y móvil, y termina
con `down -v --remove-orphans`. Los recuentos se anotan con el resultado real. Producción continúa
aparcada hasta cerrar la versión 1.0.0.

El responsable cerró el repaso el 2026-10-07 con 5A, 6A, 7A y 8A. No quedan decisiones abiertas en
esta enmienda.

## 15. Entorno desechable usado para modelos reales

La validación usa nombres propios para proyecto, red y volumen, además de una base desechable para
los recorridos de navegador. No cambia la configuración, el volumen de modelos ni los datos del
desarrollo habitual. La aplicación y el administrador local se ejecutan con el mismo código e
imágenes que se están validando.

El recorrido comprueba `/setup` y `/settings` en PC y móvil. En `/setup` descarga, activa y prueba el
modelo predeterminado antes de sellar. En `/settings` verifica catálogo, estados activo/sano,
cambio entre 1.5B y 3B, prueba de generación y mensajes de error. El 7B se descarga y valida, pero
la interfaz debe impedir su activación al no cumplirse la RAM requerida.

La medición separa memoria de `llama-server` y memoria total del contenedor. Los tiempos se toman
desde la solicitud hasta la salud y alrededor de cada petición de generación. Para probar el
supervisor se termina sólo el proceso hijo; el administrador debe permanecer vivo y recuperar el
modelo en no más de cinco minutos.

Al cerrar, se guardan únicamente el informe y los comandos documentados. Se eliminan contenedores,
red, base, adjuntos, logs y volumen de modelos del proyecto temporal. Un inventario posterior debe
demostrar que no queda ningún recurso con sus nombres. Desarrollo y producción permanecen intactos;
producción continúa aparcada hasta cerrar la versión 1.0.0.

## 16. Reanudación completada después del hallazgo de memoria

Las pruebas unitarias sustituyen lecturas de RSS, memoria efectiva, proceso y salud. Verifican que un
destino imposible no detiene el origen y que cualquier fallo posterior a la parada lo restaura sano.
Backend y frontend prueban el error estructurado y su toast con requerida/disponible.

Después se reconstruye únicamente la imagen del administrador del proyecto
`catalina-support-ai-block5` y se reutiliza `catalina_support_ai_models_block5`. Se activa 3B, se
miden memoria, tiempos, doce generaciones y recuperación. Se comprueba el rechazo de 7B conservando
3B sano y se repiten `/setup` o `/settings` sólo donde la corrección cambie el resultado, en PC y
móvil.

Al terminar se ejecuta la limpieza aprobada sobre los nombres explícitos del bloque 5. No se usa
`down -v` sobre el proyecto habitual de pruebas. Los GGUF no se vuelven a descargar y producción
permanece aparcada.

La reanudación pasó las suites completas de `ai-manager` y backend, 220 pruebas unitarias del
frontend y el recorrido afectado de `/settings` en PC y móvil. El entorno aislado se retiró al
terminar y el inventario final no dejó contenedores, redes ni volúmenes con `block5`.

## 17. Verificación aislada del bloque 6

La revisión usa `tests.yml` y datos desechables. `/setup` se recorre desde una instalación vacía;
después, una instancia reiniciada con los seeders opcionales aporta listas, estados, adjuntos,
usuarios y tickets variados para el resto de pantallas. No se leen ni modifican los datos de
desarrollo y no se usa producción.

La matriz combina español e inglés, claro y oscuro de fábrica, y 1440×900, 768×1024 y 412×915. Los
seis temas fijos pasan sus comprobaciones comunes de variables y contraste. Chromium es el navegador
de ejecución; el código sigue usando HTML y CSS compatibles, pero Firefox y WebKit no forman parte
del criterio de cierre de este bloque.

El inventario se obtiene primero, sin corregir mientras se recorre una pantalla. Después se corrige
por prioridad y se repite la matriz afectada. Las comprobaciones estables entran en las pruebas de
componentes o Playwright; la revisión manual registra lo que no pueda expresarse con una aserción
fiable. El cierre exige frontend unitario, Playwright en la matriz aprobada, `git diff --check` y
cero hallazgos visuales o de accesibilidad abiertos.

La limpieza final usa los nombres explícitos del proyecto desechable y retira contenedores, red,
base, adjuntos, logs y cachés. No se ejecuta despliegue, no se abre el dominio y no se corrige el
hallazgo conocido de `prod-build.sh` en este bloque.

La ejecución real usó el proyecto Compose `catalina-support-block6`. Primero recorrió `/setup` sobre
una base vacía y después aplicó el seeder oficial: 11 cuentas, 25 tickets y 5 adjuntos. La auditoría
reproducible encontró y permitió corregir un único desbordamiento en `/users`; la repetición afectada
y las 220 pruebas unitarias pasaron. La ejecución canónica terminó con **180 casos E2E aprobados y
38 omisiones previstas**. El proyecto aislado se eliminó con sus volúmenes y producción permaneció
sin tocar.
