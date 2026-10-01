# Ambientes: despliegue y pruebas

> **Estado:** as-built
> **Última actualización:** 2026-09-30
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
| **Dominio** | `dev-catalina-support.calibyou.com` | `catalina-support.calibyou.com` |
| **Puertos** | frontend 11001, backend 11002, base 11003 | frontend 21001, backend 21002, base 21003 |
| **Código** | Montado como volumen: se edita y el contenedor recarga | **Artefactos construidos** en `/root/prod/catalina-support` |
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

Se entra por `https://dev-catalina-support.calibyou.com`. Los puertos 11001 y 11002 siguen publicados
para depurar sin pasar por nginx.

### 3.2 La base de datos

- El esquema se aplica **repetidas veces** sin miedo: `v1.0.0.sql` es transaccional e idempotente
  (`docs/arquitectura.md`, sección 7). **El archivo existe, está aplicado y crea 18 tablas.**
- Aplicarlo, dentro del contenedor (comando probado el 2026-09-22):

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

Y volver a aplicar la migración. No hay nada que conservar: los datos de desarrollo son de prueba.

### 3.3 Los datos de ejemplo

El entorno de desarrollo se puede dejar con contenido en un comando:

```bash
./scripts/dev-seed.sh
```

Aplica `v1.0.0.sql` (por si la base está recién creada), aplica `v1.0.0_dev.sql` y **copia los
archivos de los adjuntos** desde `config/seed/` a `_files/`, en la carpeta de su ticket: un seeder
SQL no puede crear archivos, y un adjunto sin su archivo no se puede abrir.

Deja **once cuentas** —`user1@demo.com`…`user5` para quien pide, `support1`…`support3` y `dev1`…
`dev3`— y **25 tickets** repartidos por todos los estados, con asignaciones, reasignaciones,
comentarios de los tres papeles, **siete con ticket interno** (uno esperando a Desarrollo, otro
devuelto a Soporte, otros resueltos o cerrados), adjuntos de verdad y **las fechas repartidas en los
últimos tres meses**, para que la bandeja no salga toda del mismo día.

**Las direcciones son `@demo.com`** (decisión del responsable, 2026-09-26), igual que las de las
personas del directorio de pruebas y del reino: en desarrollo todo es de mentira, y un dominio solo
—el mismo en los tres sitios— es lo que hace que no haya que pensar cuál toca.

**Las once entran con `123123123`** (decisión del responsable, 2026-09-25): en desarrollo se entra y se
sale muchas veces al día, y una contraseña que se escribe con una mano es lo que hace que probar no dé
pereza.

- **Cumple la política del producto**, que pide **8 caracteres** como mínimo —bajados de 12 el
  2026-09-25—, así que el seeder **no se salta ninguna regla**: escribe el `bcrypt` ya hecho porque un
  archivo SQL no puede llamar a la aplicación, y el resultado es el mismo que si la contraseña se
  hubiera establecido desde la pantalla.
- **Las pruebas de interfaz no usan estas cuentas**: crean las suyas con la contraseña que fija
  `tests/e2e/ayudas.ts`. Cambiar la de las once no toca ninguna prueba.

**Borra para dejar un estado conocido**: los tickets que hubiera —también los creados a mano
probando—, las cuentas `@demo.com` y los adjuntos que no sean la marca. Es un seeder de
desarrollo, no una migración, y por eso se aplica cuando se quiere y no al arrancar.

**Y deja configurados los dos caminos de directorio, con el método de entrada en `local`**
(enmienda del 2026-09-25): escribe la configuración del directorio apuntando al `ldap` del `dev.yml`
y la de Keycloak apuntando al reino de pruebas, **sin cambiar el método**. Así el entorno arranca
entrando con las once cuentas de ejemplo y **cambiar de método es un clic en Configuración**, que es
justo lo que hay que poder probar a mano. Los dos servicios viven detrás del perfil `auth`: con él
levantado, «Probar la conexión» contesta que sí; sin él, contesta que no, que es la verdad.

- **La limpieza de las cuentas va en minúsculas** (`lower(email) LIKE '%@demo.com'`): el correo se
  guarda como lo escribe quien lo escribe, y sin eso una cuenta de ejemplo con mayúsculas sobrevivía
  a la limpieza y el entorno dejaba de estar en el estado conocido que promete este guion.

### 3.4 El motor de IA: su propio compose, compartido por los dos entornos

El motor que redacta el **motivo** y la **última acción** de los tickets (`docs/modules/ai.md`) **no
vive en `dev.yml` ni en `prod.yml`**, sino en **`ai.yml`**, y hay que levantarlo aparte:

```bash
docker compose -f ai.yml up -d      # la primera vez descarga el modelo (~1,1 GB) y tarda
docker compose -f ai.yml ps
docker compose -f ai.yml logs -f ai
docker compose -f ai.yml down       # el volumen del modelo se queda
```

**Por qué aparte**: es **uno solo para los dos entornos**, porque no caben dos —el modelo ocupa
~1,1 GB y la máquina tiene 1,8 GB libres, sin GPU—. Los backends de desarrollo y de producción **entran
en su red** (`catalina-support-ai`): **`dev.yml` y `prod.yml` la crean** —con su nombre fijo y sin
`external`— y **`ai.yml` la declara externa**, para que el motor no intente recrearla. Le hablan por su
nombre, `http://catalina_support_ai:8080`. **No publica ningún puerto**: el motor no se alcanza desde la
máquina ni desde fuera.

- **El modelo se descarga una vez** a un volumen con nombre (`ai_modelos`), así que sobrevive a `down`
  y a recrear el contenedor; sólo `down -v` lo borra. A partir de ahí el motor funciona **sin salida a
  internet**.
- **Es opcional, y a propósito**: **su dirección y su modelo se configuran en Configuración** —con su
  botón de probar la conexión— y `AI_URL`/`AI_MODEL` quedan **como respaldo**; sin motor configurado
  —o con el contenedor parado— los dos campos se quedan sin texto y **todo lo demás funciona igual**.
  Un motor caído no puede parar la mesa de ayuda (`docs/modules/settings.md`, decisión 16, y
  `docs/modules/ai.md`, decisión 2).
- **La red la crea el entorno**, así que **el motor se puede levantar después, con el entorno en
  marcha**: `docker compose -f dev.yml up -d` funciona en una máquina nueva sin crear nada a mano.
- Antes de desplegar a producción, `docker compose -f ai.yml up -d` en el servidor: los dos entornos lo
  comparten.

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
`/root/prod/catalina-support` (frontend, backend, `_logs`, `_files`). Los `builder_*` de `prod.yml`
existen para construir y publicar, y **no se levantan con `up`**: están bajo el perfil `build`.

### 4.2 El despliegue, paso a paso

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
5. **Escribe `BUILD_INFO`** en `/root/prod/catalina-support` con el commit, la rama, si el árbol
   estaba sucio, la fecha y qué artefactos se construyeron. Es la respuesta a «¿qué hay desplegado
   ahora mismo?» sin adivinar.
6. **Despliega**: `docker compose -f prod.yml up -d` y **reinicia el backend**. El frontend no se
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
  `/root/prod/catalina-support` es el código construido. Es lo que monta `prod.yml`, y es lo que
  tiene que saber cualquier guion que hable con producción (se corrigió así en
  `scripts/backup-db.sh`, sección 6).

### 4.3 Después de desplegar

1. **Aplicar las migraciones** que toquen (sección 5).
2. **Comprobar que arrancó**, en este orden y sin saltarse ninguno:

```bash
docker compose -f prod.yml ps                                  # los tres arriba y la base sana
curl -s https://catalina-support.calibyou.com/api/health       # {"database":"ok","status":"ok"}
curl -sI https://catalina-support.calibyou.com/ | head -1      # 200 y la CSP de producción
```

3. **Mirar los logs** del backend, que es donde `go-logs` cuenta lo que ha pasado:

```bash
docker compose -f prod.yml logs --tail=50 backend
tail -f /root/prod/catalina-support/_logs/catalina-support_prod_$(date +%Y%m%d).log
```

### 4.4 Volver atrás

Si el despliegue sale mal, **el artefacto anterior sigue ahí**: se copia `<nombre>_prev` sobre
`<nombre>`, se reinicia el backend y se vuelve a la versión de antes. Está automatizado en la
sección 4.2 (la rotación) y a mano en el peor caso:

```bash
cp -a /root/prod/catalina-support/backend_prev/. /root/prod/catalina-support/backend/
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
| `v1.0.0_dev.sql` | Los **datos de ejemplo**: once cuentas (cinco que piden, tres de Soporte y tres de Desarrollo), 25 tickets con su historia —asignaciones, reasignaciones, comentarios, escalados con su interno, adjuntos— y el contador de la numeración | **Sólo desarrollo**, con `scripts/dev-seed.sh` |

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
enseguida, porque la aplicación pedirá una tabla que no existe.

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
   nombre: `/root/backups/catalina-support/catalina_support_20260925_033000.dump`.
3. **Comprueba que el volcado se puede leer** (`pg_restore --list`). Una copia que no se puede
   restaurar no es una copia, y eso hay que saberlo hoy y no el día que haga falta.
4. **Borra las copias de más de catorce días** y dice cuántas quedan.

**No copia `_files`**, y es a propósito (corrección del responsable, 2026-09-25): **los adjuntos de
los tickets y los documentos que sube la gente no se borran nunca**, y copiarlos es cosa de soporte
manual cuando haga falta. El script se ocupa de la base y de nada más. Cuando haya que llevarse los
archivos —a otro servidor, a un disco, a una revisión—, se copian desde la máquina como cualquier
carpeta, respetando lo que hay dentro:

```bash
cp -a /root/dev/Catalina-Support/_files/. /root/backups/catalina-support/_files/
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

**Dónde viven las copias**: `/root/backups/catalina-support/`, fuera del repositorio y con las
copias de los otros proyectos de la máquina. El script escribe lo que hace por la salida estándar, y
el `cron` lo recoge en `copias.log`, ahí mismo:

```cron
30 3 * * * /root/dev/Catalina-Support/scripts/backup-db.sh dev  >> /root/backups/catalina-support/copias.log 2>&1
45 3 * * * /root/dev/Catalina-Support/scripts/backup-db.sh prod >> /root/backups/catalina-support/copias.log 2>&1
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
  --no-owner < /root/backups/catalina-support/catalina_support_20260925_033000.dump
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
echo | openssl s_client -connect catalina-support.calibyou.com:443 \
  -servername catalina-support.calibyou.com 2>/dev/null | openssl x509 -noout -dates
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

### 9.3 Interfaz: Playwright contra desarrollo

Como en Calibyou: `tests/e2e/` con Playwright, **contra el entorno de desarrollo**, con un caso por
flujo crítico. Un recorrido completo vale más que veinte pruebas de detalle.

**Se ejecutan en un contenedor**, no en la máquina: la imagen de Playwright ya trae los navegadores
instalados, y en la máquina no se instala Node ni nada (AGENTS.md). El servicio es `e2e` en
`dev.yml`, detrás de un perfil, así que **no se levanta con el resto**:

```bash
docker compose -f dev.yml up -d                  # el entorno tiene que estar arriba
docker compose -f dev.yml run --rm e2e           # instala lo que falte y ejecuta las pruebas
```

- **Por defecto corre en PC y en móvil** (`pc` y `movil`, con un Pixel 7): la interfaz tiene que
  funcionar en los tres anchos, no sólo caber (`docs/interfaz-y-experiencia.md`, sección 7).
- **Lee la contraseña de fábrica y el buzón de pruebas** del entorno de desarrollo: `BASE_URL`,
  `ADMIN_PASSWORD` y `MAILPIT_URL`, los tres por variables. Sin ellos, los casos que necesitan entrar
  o leer un correo se saltan en vez de fallar con un mensaje que no explica nada.
- **Después de cada pasada se reinicia el entorno** (`./scripts/dev-seed.sh`, decisión del
  responsable, 2026-09-28): los tickets de prueba que deja la suite **no se borran desde ella** —un
  ticket no se borra—, así que el entorno se deja como nuevo con el seeder, **avisando antes** porque
  el seeder borra también lo que haya a mano.
- **Cada caso monta sus datos y los deja apagados al terminar**: da de alta sus cuentas —lo que prueba
  el camino de verdad, con su correo y su enlace—, y **al acabar la pasada el cierre
  (`tests/e2e/limpiar.ts`) las apaga** por la API. No se borran: las cuentas no se borran en este
  producto. Y **los tickets que crea la pasada tampoco se borran** —un ticket no se borra—: si se
  quiere el entorno como nuevo, `./scripts/dev-seed.sh` borra los tickets y las cuentas de prueba y
  vuelve a poner las once de ejemplo.
- **Los correos se leen del buzón**, no se dan por hechos: el caso del enlace del correo saca el
  token de la pantalla de Mailpit y lo abre en el navegador, que es lo que haría una persona.
- Cuando algo no se ve como debería, hay una herramienta que vuelca el HTML pintado, los errores de
  la consola y una captura:

```bash
docker compose -f dev.yml run --rm -e DIAGNOSTICO=1 e2e
```

**Las cuentas que crean las pruebas se quedan**, porque en este producto **no existe dar de baja
cuentas** (`docs/modules/users.md`): dar de alta una es para siempre. Se llaman `e2e-…@demo.com`,
así que se reconocen, y en desarrollo se limpian cuando estorban:

```bash
docker compose -f dev.yml exec database psql -U catalina_support -d catalina_support -p 11003 \
  -c "DELETE FROM ticket_comments; DELETE FROM ticket_attachments; DELETE FROM ticket_history; \
      DELETE FROM internal_tickets; DELETE FROM tickets; DELETE FROM ticket_number_counters; \
      DELETE FROM password_tokens WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'e2e-%@demo.com' OR email LIKE '%directorio@demo.com' OR email LIKE '%keycloak@demo.com'); \
      DELETE FROM users WHERE email LIKE 'e2e-%@demo.com' OR email LIKE '%directorio@demo.com' OR email LIKE '%keycloak@demo.com';"
```

**Los tickets van primero, y no es un detalle de orden**: un ticket apunta a quien lo pidió, y la
base no deja borrar una cuenta que tiene tickets detrás (`ON DELETE RESTRICT`), que es justo lo que
hace que la historia de un ticket sobreviva a la baja de una cuenta.

**Y después de limpiar, se vuelve a los datos de ejemplo con un comando**: esa limpieza se lleva por
delante también los 25 tickets del seeder —borra **todos** los tickets, a propósito—, así que lo
normal es terminar con `./scripts/dev-seed.sh`, que deja el entorno otra vez conocido (sección 3.3).

**Las cuentas de las personas del directorio y del reino de Keycloak se borran a propósito**, y no es
por limpieza: los casos del alta automática de los dos caminos de directorio comprueban que la cuenta
**no existía** antes de entrar, y con una cuenta de una ejecución anterior lo que probarían es otra
cosa. Si se dejan, esos casos se saltan solos diciendo qué hay que limpiar
(`docs/modules/auth.md`, sección 11).

**Los casos de los caminos de directorio —AD y Keycloak— necesitan sus servicios levantados**, que en
desarrollo viven detrás del perfil `auth`:

```bash
docker compose -f dev.yml --profile auth up -d          # el directorio y Keycloak, con sus personas
docker compose -f dev.yml --profile auth run --rm e2e   # la suite entera, con los dos caminos
```

Sin él, esos casos **se saltan** en vez de fallar: la suite tiene que poder correr en una instalación
que no tenga AD (`docs/modules/auth.md`, decisión 28).

**Y cinco de ellos —el alta automática y el vínculo de cada camino, que son los que crean la
cuenta— corren sólo en PC**: los dos proyectos comparten la misma base de datos, así que el segundo
vería la cuenta que dejó el primero y ya no probaría lo mismo. Se saltan diciéndolo, no en silencio.

**Keycloak se levanta con su reino ya importado**, y el navegador llega a él por el mismo dominio que
a la aplicación (`/sso/`, que es un `location` del vhost de desarrollo): por eso, si se cambia algo de
Keycloak en el repositorio, se recrea el contenedor y **el reino vuelve a como está escrito**.

**Si la suite falla con un `vite-error-overlay` que intercepta el clic**, no es la prueba: es que el
**servidor de desarrollo está sirviendo un error de compilación** de un cambio anterior. Se ve en sus
logs (`docker compose -f dev.yml logs --tail 40 frontend`) y se arregla tocando el archivo que dice el
error —o reiniciando el contenedor—, porque el servidor sólo vuelve a intentar lo que ve cambiar.
Pasó al añadir un texto que faltaba: la suite siguió corriendo contra una pantalla que no compilaba y
el fallo que contaba hablaba de otra cosa.

Los casos, con lo que hay hoy:

| Caso | Qué recorre | Estado |
| --- | --- | --- |
| La aplicación abre | La raíz lleva a la entrada, los campos tienen nombre accesible, se llega con el tabulador y **ninguna pantalla repite un `id`** | **Hecho** |
| La marca | Que **el logo de fábrica cambie con el tema** —el claro y el oscuro, cada uno cargando de verdad—, que el logo **cargue de verdad** en la pantalla de entrada, que un logo propio sustituya al de fábrica y que al quitarlo se vuelva, y que el color institucional se aplique a los dos temas de fábrica sin romper el contraste. Y **la versión del sistema**: que se lea en el pie de la entrada y en la fila de salir del menú —comparada con lo que dice la API—, que sea texto y no un enlace, que se lea sobre el fondo del menú, y que plegado no se enseñe | **Hecho** |
| El armazón | Las tres zonas del menú, el plegado que se recuerda, el cajón del móvil y que las pantallas de la sesión no lo lleven | **Hecho** |
| Los permisos del menú | Que un papel no vea lo que no le toca **y** que el backend se lo rechace si lo pide a mano | **Hecho** |
| La pantalla de Configuración | Subir un logo por el formulario, verlo en la vista previa y en el menú, volver al de fábrica, el color con su vista previa antes de guardar, **el nombre de la instalación en los tres sitios donde se lee**, **el método de entrada cambiándolo desde la pantalla** y **las dos pruebas de conexión** —la del directorio y la del reino, con lo que hay en pantalla y sin guardar nada— | **Hecho** |
| La lista de cuentas | Que un Administrador **entre directamente en ella**, los filtros por papel, origen y estado, la búsqueda por nombre o correo, el estado vacío y el filtro que esconde lo desactivado | **Hecho** |
| Dar de alta una cuenta | El diálogo, el papel que se elige, el aviso de lo que ha pasado, **el enlace leído del buzón de pruebas** y que la persona nueva entre con él | **Hecho** |
| Desactivar y reactivar | Que desactivar **pregunte antes**, que la persona deje de poder entrar de verdad y que reactivarla la devuelva | **Hecho** |
| Soporte edita en la lista | Que Soporte cambie nombre y apellidos **en línea**, que el cambio esté guardado de verdad al recargar, y que **no tenga ficha de nadie** (403 si la escribe a mano) | **Hecho** |
| El perfil propio | Cambiar el nombre desde tu nombre en el menú, que el menú lo enseñe al momento, que el correo no se pueda tocar, y que el idioma de la cuenta cambie la interfaz (y sus correos) | **Hecho** |
| Nadie se desactiva a sí mismo | Que tu propia ficha **no ofrezca** el botón, y que la cuenta de fábrica no tenga perfil ni enlace a él | **Hecho** |
| El idioma de arranque | Español, inglés y **un idioma que no es ninguno de los dos** (inglés), más el conmutador | **Hecho** |
| El tema | Que siga al sistema, que el selector cambie **de verdad** —se mide el color de fondo—, que se recuerde al recargar, que estén los ocho con sus grupos y **que los ocho se lean**: el contraste se mide en el navegador, sobre lo que se pinta, para el texto, el texto apagado, el acento, el botón, el borde del campo y los tres colores de estado | **Hecho** |
| Entrar y salir | Sesión, cabecera `Authorization`, `GET /api/auth/me`, salir y que sin sesión no se entra | **Hecho** |
| La política de contraseñas y el enlace | Que una contraseña de **siete** caracteres se rechace con su clave, que **el enlace siga sirviendo** después de ese intento fallido, que **ocho** caracteres valgan, y que con esa contraseña se entre de verdad. Nació al bajar el mínimo de 12 a 8 y encontró que el enlace se gastaba antes de comprobar la contraseña | **Hecho** |
| El enlace del correo | Alta, correo en el buzón, enlace en el fragmento, contraseña puesta y entrada con ella, y que **un enlace no vale dos veces** | **Hecho** |
| Los adjuntos en los comentarios | Que un comentario lleve sus archivos **pegados** (con un evento de pegado de verdad), **elegidos** con el botón y arrastrados; que **un comentario pueda ser sólo un archivo** —con el cuerpo vacío en la base y **sin dejar un párrafo vacío** en la conversación—; que sin texto **ni** archivos no se envíe; y que **si una subida falla** —cortándola en el navegador— el comentario se quede publicado **una sola vez**, se diga qué archivo falta y el botón de reintentar lo suba al mismo comentario. Y el pegado también en el alta | **Hecho** |
| El recorrido de un ticket | El usuario lo abre con su adjunto, comenta, lo cierra y lo reabre; Soporte pregunta y **escala**; Desarrollo **lo devuelve**; y el principal vuelve a la bandeja de Soporte | **Hecho** |
| El aspecto del ticket | El estado con **su color y su borde** medidos, las tarjetas con el borde del tema, la ficha **a la derecha en PC y debajo en móvil**, y que en móvil **no se pinte la tabla** | **Hecho** |
| Los permisos de los tickets | El usuario no ve el interno, Desarrollo no escribe en el principal y al Administrador no se le ofrece ningún botón | **Hecho** |
| El editor de los correos | Llegar desde Configuración, cambiar el asunto y el cuerpo, **insertar un marcador pulsándolo**, ver la vista previa que renderiza el backend, guardarlo, **mandarse una prueba y leerla en el buzón**, y volver al texto de fábrica. Y que **un marcador inventado se avise antes de guardar** | **Hecho** |
| El camino de Keycloak | Que el botón **esté sólo si la instalación tiene ese camino** y que lleve a la pantalla de Keycloak, que una persona entre por el reino **sin que nadie le dé de alta nada** y su cuenta aparezca con origen `keycloak`, que quien ya es de Keycloak vuelva a entrar, que una cuenta local con el correo de alguien del reino **se vincule** (y su contraseña local deje de servir), **que el fragmento con el token se borre de la dirección**, y que una vuelta que no vale se cuente en la pantalla de entrada | **Hecho** (necesita el perfil `auth`) |
| Las tres acciones del directorio en `users` | Que la ficha de una cuenta de **AD** desactivada **sí ofrezca reactivarla** y que la reactivación pregunte al directorio de verdad, y que la de una cuenta de **Keycloak** no ofrezca el botón y **cuente lo que pasa** —que vuelve sola al entrar— en vez de llevar a un error | **Hecho** (necesita el perfil `auth`) |
| El camino de AD | Que una persona del directorio **entre sin que nadie le dé de alta nada** y que su cuenta aparezca con origen `ad` y sin contraseña local, que quien ya es del directorio vuelva a entrar, que la contraseña equivocada **la rechace el directorio** y que una cuenta local con el correo de alguien del directorio **se vincule** al entrar por su camino (y su contraseña local deje de servir) | **Hecho** (necesita el perfil `auth`) |

**Y desde el 2026-09-25, las dos suites de los caminos de directorio ponen el método y lo devuelven.**
Como la instalación entra por uno solo, probar el de AD o el de Keycloak exige elegirlo primero: cada
suite lo pone al empezar y **lo deja en `local` al terminar, pase lo que pase**, con la cuenta de
fábrica, que entra siempre. Si el servicio no responde, la suite **se salta** en vez de fallar, y lo
pregunta **por el botón de «Probar la conexión»**, que es lo que pregunta una persona y no depende del
método puesto.

**El backend se prueba con `curl`, y los correos se leen del buzón.** Mientras una pantalla no
existe, lo que se comprueba de extremo a extremo es la API: se entra con cuentas de verdad, se
recorre el caso completo y se miran los avisos en Mailpit. Así se probó el módulo de `tickets` el
2026-09-24: **78 comprobaciones** que van del alta con su reparto hasta el adjunto que el solicitante
no puede alcanzar, pasando por las dos reglas de sincronización y el re-escalado. Cuando exista la
pantalla, esos recorridos pasan a `tests/e2e/` y esto queda como lo que es: la prueba de la API.

**Cuándo se ejecutan**: cuando el cambio toca la interfaz, antes de darlo por terminado. Si el cambio
afecta a un flujo que todavía no tiene prueba, **se añade el caso en el mismo trabajo**.

**Lo que encontró esta capa el 2026-09-23**, para que se vea para qué sirve: tres fallos que las
pruebas de unidad no podían ver.

1. Un `id` repetido en el componente de campo dejaba los campos **sin nombre accesible**: la etiqueta
   apuntaba al elemento anfitrión y no al campo.
2. Un 401 del enlace de contraseña **echaba de su sesión** a quien ya estaba dentro, en vez de decir
   que el enlace había caducado.
3. `hayToken` era un valor calculado que leía el almacenamiento **sin depender de nada**, así que se
   quedaba con el primer «no» para siempre: después de entrar bien, la guarda devolvía a la pantalla
   de entrada.
4. **El CSS llegaba sin estilos**: la hoja se servía con el tema de Tailwind pero **sin ninguna
   utilidad generada**, así que la pantalla se veía como texto suelto. Lo encontró el responsable al
   abrir la aplicación, no una prueba: las de estructura daban todo por bueno. Ahora hay un caso que
   mira el **aspecto** —que la hoja traiga utilidades, que el botón tenga color, que el campo tenga
   borde y que no falle ninguna petición—, y es el que impide que esto vuelva sin que nadie se
   entere.

Los cuatro están corregidos y con su prueba. El segundo, el tercero y el cuarto sólo se ven **usando
la aplicación de verdad**, que es justo lo que aporta esta capa.

**Y el 2026-09-24, con las pantallas de usuarios, otros dos**:

5. **La tarjeta titulaba con `h1`**, así que la pantalla de usuarios tenía tres `h1` y el perfil otros
   tantos: una página así no se navega con un lector de pantalla. Lo dijo la prueba al encontrar dos
   encabezados con el mismo nombre donde esperaba uno. La tarjeta ahora **elige su nivel**, y en las
   pantallas de producto es un `h2`.
6. **El cajón del menú medía el alto del documento y no el de la zona visible** (`inset-y-0`): en un
   móvil con barra del navegador lo visible son 839 px de 890, y **el botón de salir quedaba por debajo
   de lo que se ve**, imposible de pulsar. En PC, con una lista larga, los controles del menú quedaban
   al final de la página. Ahora mide la zona visible (`h-dvh`), se desplaza por dentro si su contenido
   no cabe y en PC va pegado arriba, que es lo que el documento de interfaz llama «fijo». La prueba
   decía que «otro elemento interceptaba el clic»: ese mensaje no cuenta el problema de verdad.

Los seis están corregidos y con su prueba.

**Antes de la primera prueba, la instalación vuelve a entrar por cuentas de la aplicación**
(`tests/e2e/preparar.ts`, la preparación previa de `playwright.config.ts`). Hace falta porque **la
instalación entra por un método a la vez**: si una pasada se corta con el método en `ad` o en
`keycloak`, la siguiente empieza con decenas de casos fallando por «credenciales incorrectas», que no
dicen nada de lo que pasó de verdad —pasó el 2026-09-27—. Lo hace la **cuenta de fábrica**, que entra
siempre sea cual sea el método.

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
