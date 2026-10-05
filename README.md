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

## Probar en local: instalación vacía

Este es el recorrido recomendado: configurar una instalación vacía y crear tus propias cuentas.
Los datos de ejemplo son opcionales y se explican después. Si ya ejecutaste `seed`, la instalación está sellada
y `/setup` no volverá a aparecer.

Necesitas **Docker con Docker Compose**, Git y un navegador. No necesitas instalar Go, Node,
PostgreSQL ni nginx. Docker debe estar en marcha y la primera ejecución necesita internet para
descargar imágenes y dependencias; puede tardar varios minutos.

Los siguientes comandos sirven en PowerShell, CMD y bash. Trae el código:

```bash
git clone https://github.com/juanky-estevez/Catalina-Support.git catalina-support
cd catalina-support
```

Levanta el entorno:

```bash
docker compose -f dev.yml up -d --build
docker compose -f dev.yml ps
```

Ese único arranque espera a PostgreSQL, aplica el esquema y sólo entonces inicia el backend. No
carga datos de ejemplo. La migración es transaccional e idempotente; si falla, el backend permanece
detenido y el error se ve en los registros de `migrate`. Espera también a que frontend y backend
terminen de compilar; puedes seguir sus registros con:

```bash
docker compose -f dev.yml logs -f frontend backend
```

`Ctrl+C` deja de mostrar los registros y los servicios siguen funcionando.

Abre [http://127.0.0.1:11001](http://127.0.0.1:11001): **aparece `/setup`, todavía no el login**.
Completa sus pasos:

| Paso | Valores para probar en local |
| --- | --- |
| Instalación | El nombre que quieras. El asistente empieza en inglés; elegir Español traduce toda la pantalla inmediatamente |
| Cómo se entra | **Local**; no necesita un directorio externo |
| Dónde está | Tu zona horaria y **`http://127.0.0.1:11001`** como dirección pública |
| Correo saliente | Ya aparece **`mail`**, puerto **`1025`**, sin TLS ni credenciales, remitente **Catalina Support** y **`no-responder@catalina-support.local`**. Todos los valores son editables |

En el paso del correo pulsa **Probar la conexión**. Comprueba conexión y autenticación; no envía un
mensaje. El aviso de motor de IA ausente no impide terminar. Al finalizar queda sellada la instalación:
entra con **`admin` / `admin`**, crea cuentas desde Usuarios y abre sus enlaces en
[Mailpit](http://127.0.0.1:11004). Para atender y escalar tickets necesitarás cuentas de Soporte y Desarrollo.

**Cuenta de fábrica:** en local es **`admin` / `admin`**, definida en `config/env/dev.env`.
En producción, su contraseña es el valor de **`ADMIN_PASSWORD`** en el entorno, no `admin`.
El asistente no cambia ni pide esa contraseña. Después de terminar el asistente, esa cuenta permite
configurar la aplicación; entra siempre, incluso si eliges AD o Keycloak.

| Qué abrir | Dirección |
| --- | --- |
| Aplicación | [http://127.0.0.1:11001](http://127.0.0.1:11001) |
| Buzón de pruebas | [http://127.0.0.1:11004](http://127.0.0.1:11004) |
| Salud del backend | [http://127.0.0.1:11002/api/health](http://127.0.0.1:11002/api/health) |

## Datos de ejemplo (opcionales)

Si quieres contenido y cuentas listas para probar los cuatro papeles, desde la carpeta del proyecto:

```bash
docker compose -f dev.yml run --rm seed
```

**Este comando prepara el esquema y los ejemplos**: 11 cuentas, 25 tickets y 5 adjuntos.
No hace falta aplicar la migración por separado para esta demo.
**Este comando borra todos los tickets, categorías, etiquetas, cuentas `@demo.com` y adjuntos
de desarrollo (excepto la marca)**. Úsalo sólo cuando quieras volver a los ejemplos.

Si ya terminaste el asistente, **conserva la configuración**: nombre, marca, idioma, dirección,
correo, método de entrada, directorios, IA, prefijo y reparto. También conserva el sello.
Si todavía no lo terminaste, **completa y sella una demo local**: reemplaza los ajustes incompletos
con los de la demo y configura Mailpit y `http://127.0.0.1:11001`.

La cuenta de fábrica sigue siendo `admin` / `admin`. Si conservaste AD o Keycloak como método de
entrada, las cuentas locales de ejemplo no entran: usa `admin` para volver a **Local** si quieres
probarlas. Si tu dirección pública conservada apunta a otro sitio, ajústala desde Configuración.

### Cuentas y recorrido para probarlo

Todas estas cuentas usan la contraseña **`123123123`**. `admin` usa **`admin`** y no está en esta tabla.
Son credenciales para pruebas locales; no las uses en producción.

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

1. Entra como `user1@demo.com`, abre **Mis tickets** y crea un ticket con categoría y descripción.
2. Sal y entra como `support1@demo.com`. Busca el ticket en **Tickets principales**, comenta y
   prueba sus cambios de estado. Puedes resolverlo o escalarlo indicando el contexto.
3. Si lo escalas, entra como `dev1@demo.com` y busca el interno en **Tickets internos**.
   Prueba la conversación interna y la resolución; el usuario sólo ve el principal.
4. Abre el buzón de pruebas para leer los avisos. Con `admin` puedes crear otra cuenta local:
   abre su correo de alta en el buzón y sigue el enlace para establecer su contraseña.

No hace falta levantar el motor de IA, AD ni Keycloak para completar este recorrido.

### Parar y volver a arrancar

Desde la carpeta del repositorio:

```bash
docker compose -f dev.yml stop
docker compose -f dev.yml up -d
```

La base y los adjuntos se conservan. **No ejecutes `seed` para arrancar de nuevo**: ese comando
reinicia los datos de prueba. `docker compose -f dev.yml down` elimina los contenedores y la red,
pero conserva los volúmenes; añadir `-v` elimina también los volúmenes, incluida la base.

## Si algo falla

| Síntoma | Qué comprobar |
| --- | --- |
| `docker compose` no existe o Docker no responde | Docker debe estar instalado y en marcha; en Linux tu usuario necesita acceso al daemon. Comprueba `docker compose version`. |
| Un puerto está ocupado | El entorno usa 11001–11004. Detén el servicio que los ocupa antes de arrancar; una segunda copia del repositorio comparte nombres de contenedores y volúmenes. |
| La web aún no abre | Mira `docker compose -f dev.yml logs -f frontend backend`: `up -d` no espera a que Angular y Go terminen de compilar. |
| `backend` no arranca y `migrate` terminó con error | Mira `docker compose -f dev.yml logs migrate`. Corrige el error y repite `docker compose -f dev.yml run --rm migrate`; después ejecuta `docker compose -f dev.yml up -d backend`. |
| Aparece el asistente en vez del login | Es normal en una base nueva sin ejemplos: completa los cuatro pasos. |
| El enlace del correo abre el dominio del proyecto | Corrige la dirección pública en Configuración a `http://127.0.0.1:11001`. El seeder conserva esta dirección si la instalación ya está sellada. |
| Las cuentas de ejemplo no entran | Comprueba que cargaste los ejemplos y que el método de entrada es **Local**. `admin` entra siempre y permite cambiarlo. |
| No llegan correos al buzón | Comprueba que `mail` está en marcha y que el SMTP de la instalación apunta a `mail:1025`, sin TLS ni credenciales. |
| No hay resúmenes de IA | Son opcionales. Levanta y configura el motor si quieres probarlos. |

Los datos persisten en volúmenes de Docker y los adjuntos en `_files/`. No borres los volúmenes
si quieres conservar el trabajo.

## Requisitos y alcance de la prueba local

Linux, macOS o Windows con Docker Compose. En macOS y Windows, Docker Desktop debe estar arrancado.
Los comandos anteriores no necesitan Bash. Los scripts `./scripts/*.sh` sí lo necesitan.

Como orientación, asigna **4 GB de RAM o más** a Docker para construir y probar con margen.
No hay un mínimo de memoria o disco validado para todos los sistemas. Las imágenes, dependencias,
cachés y datos necesitan espacio adicional; el modelo de IA añade unos **1,1 GB** de descarga y su
contenedor tiene un **tope de 1500 MiB**, con unos **1,44 GiB medidos** en esta máquina.
No requiere GPU. Los tiempos dependen del procesador y de la conexión.

`dev.yml` publica sus puertos en todas las interfaces. Usa este entorno para pruebas en una máquina
y red de confianza; no lo publiques como instalación de producción.

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

## Servicios opcionales

Arranca primero el entorno local. Cada servicio tiene su compose y se puede levantar por separado.

### Resúmenes con IA

```bash
docker compose -f ai.yml up -d
docker compose -f ai.yml logs -f ai
```

La primera vez descarga **Qwen2.5-1.5B-Instruct Q4_K_M** (~1,1 GB) y después carga el modelo.
En **Configuración → motor de IA**, usa la dirección **`http://catalina_support_ai:8080`** y el modelo
**`qwen2.5-1.5b-instruct`**. Prueba la conexión y guarda. La dirección es la que alcanza el backend
por la red de Docker; no es una dirección que tengas que abrir en el navegador.

Los resúmenes se generan en segundo plano. En esta máquina se midieron 12–24 segundos por campo;
no es un tiempo garantizado. Sin motor, el resto del producto funciona.
Para detenerlo: `docker compose -f ai.yml stop`.
Más detalles en [el documento del módulo](docs/modules/ai.md).

### Active Directory de pruebas

```bash
docker compose -f active-directory.yml up -d
```

Es OpenLDAP con personas de ejemplo, no un Active Directory de una organización.
Los ejemplos dejan su configuración preparada. Desde `admin`, prueba la conexión del directorio
y elige **AD** como método de entrada. Si partiste de una instalación vacía, configura:

| Campo | Valor |
| --- | --- |
| Servidor / puerto | `ldap` / `389`, sin TLS |
| Cuenta de servicio / contraseña | `cn=admin,dc=ejemplo,dc=com` / `admin-directorio` |
| Base de búsqueda / filtro | `ou=personas,dc=ejemplo,dc=com` / `(mail=%s)` |
| Atributos de correo, nombre, apellido e identificador | `mail`, `givenName`, `sn`, `uid` |

Puedes entrar como **`ana.directorio@demo.com`**, contraseña **`una-contraseña-larga`**.
La cuenta se crea en su primer acceso. Mientras AD sea el método elegido, las cuentas locales de
ejemplo no entran; `admin` sí, y permite volver a **Local**.
Para apagarlo: `docker compose -f active-directory.yml down`. Más detalles en
[autenticación](docs/modules/auth.md).

### Keycloak de pruebas

```bash
docker compose -f keycloak.yml up -d
```

Espera a que termine de importar el reino; puedes verlo con
`docker compose -f keycloak.yml logs -f keycloak`. Entra con `admin` y, en Configuración, elige
Keycloak y completa:

| Campo | Valor local |
| --- | --- |
| Emisor del reino | `http://127.0.0.1:11006/sso/realms/catalina-support` |
| Dirección interna del reino | `http://keycloak:8080/sso/realms/catalina-support` |
| Cliente | `catalina-support` |
| Secreto del cliente de pruebas | `el-secreto-de-desarrollo` |
| Dirección de vuelta | `http://127.0.0.1:11001/api/auth/keycloak/callback` |

Mantén `http://127.0.0.1:11001` como dirección pública de la aplicación. **Prueba la conexión**
y guarda el método. Sal y pulsa **Entrar con Keycloak**: usa `sara.keycloak@demo.com`, contraseña
`la-de-keycloak-larga`. Su cuenta se crea en el primer acceso. No necesitas nginx, TLS ni editar
el archivo de hosts. El navegador usa la dirección pública y el backend la interna; las dos
identifican el mismo reino.

Mientras Keycloak sea el método elegido, las cuentas locales de ejemplo no entran. La pantalla
ofrece la puerta de `admin`, que permite volver a **Local**. Para apagar el servicio:
`docker compose -f keycloak.yml down`. Los datos del reino son de prueba; al recrearlo se importa
el archivo del repositorio. Las credenciales anteriores no se usan en producción.

## Usarlo en producción

**La versión 1.0.0 sigue abierta.** La prueba local anterior es el recorrido disponible para evaluar
el producto. El despliegue público de este proyecto sigue aparcado por decisión del responsable.

La configuración habitual se hace desde el asistente y después desde Configuración, pero el
despliegue actual todavía requiere trabajo del servidor:

- Docker Compose, Bash, nginx y un certificado TLS para tu dominio.
- Crear `config/env/prod.env` a partir de [prod.env.example](config/env/prod.env.example), con valores
  propios de `ADMIN_PASSWORD`, `TOKEN_SECRET` y PostgreSQL. Ese archivo no se versiona.
- Construir los artefactos, aplicar las migraciones de versión (nunca los archivos `_dev`) y
  completar el asistente antes de exponer la instalación.
- Configurar el SMTP real y crear las cuentas; mantener copias de la base **y de los adjuntos**.
  El script de copias sólo incluye la base y conserva 14 días.

**Limitaciones detectadas en la revisión del 2026-10-02:**

- `scripts/prod-build.sh` publica en **`/srv/catalina-support`**. Construye desde esa carpeta,
  pero sus Dockerfiles necesitan fuentes que están en el repositorio: el recorrido necesita corregirse
  antes de presentarlo como un despliegue reproducible desde cero.
- El vhost incluido usa el dominio y los certificados de esta máquina, un snippet externo de nginx
  y un aviso **503**. Copiarlo no publica automáticamente tu instalación.
- `prod.yml` publica 21001–21003 en todas las interfaces, no sólo en localhost. La exposición debe
  resolverse antes de abrir el servidor.

El [runbook](docs/ambientes.md) conserva los detalles del despliegue realizado y sus pendientes.
La revisión del README no modifica scripts ni abre producción.

## Pruebas para quienes desarrollan

Las pruebas de unidad se ejecutan en contenedores:

```bash
docker compose -f dev.yml exec backend go test ./...
docker compose -f dev.yml exec backend go vet ./...
docker compose -f dev.yml exec frontend npm test -- --watch=false
```

**Playwright usa su propia instalación desechable**, sin puertos publicados, sin nginx ni dominio.
No usa la base, los correos, los adjuntos ni la configuración de tu instalación local. No necesita
que hayas levantado `dev.yml` ni cargado ejemplos.

Antes de cada pasada, elimina sólo los datos de la instancia de pruebas:

```bash
docker compose -f tests.yml --profile directory down -v
docker compose -f tests.yml up -d --build database backend frontend mail
docker compose -f tests.yml run --rm e2e
```

El runner instala sus dependencias, recorre el asistente vacío en PC y móvil y prepara sus propios
fixtures por la API. Después ejecuta la suite. Un fallo devuelve un código de salida distinto de
cero; no lo ignores. Los casos de directorio se omiten si sus servicios no están presentes.

Para incluir **AD y Keycloak**, el segundo comando es:

```bash
docker compose -f tests.yml --profile directory up -d --build database backend frontend mail ldap keycloak
```

Después usa el mismo `run --rm e2e`. Sus directorios también son exclusivos de pruebas.

Al terminar, incluso si falla una prueba, retira la instancia:

```bash
docker compose -f tests.yml --profile directory down -v
```

Los informes y capturas quedan en **`tests/e2e/resultados/`** y se conservan al retirar los
contenedores. Si interrumpes una pasada, ejecuta la limpieza antes de comenzar otra. No ejecutes
`seed` para limpiar las pruebas: ese comando modifica tu instalación de desarrollo.
Las pasadas no se ejecutan simultáneamente y la instancia de pruebas debe empezar sin sello.

## Participar y donar

El proyecto sigue **documentación antes que código**: lee [AGENTS.md](AGENTS.md) y
[el índice de documentos](docs/README.md). Propón el cambio en el documento del área y espera su
aprobación antes de programarlo.

> **Pendiente del responsable:** vías de contacto, invitación a participar y donaciones.

## Licencia

[MIT](LICENSE). Puedes usarlo, estudiarlo, modificarlo y distribuirlo.
