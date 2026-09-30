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
  español y en inglés. Es **opcional a propósito**: si el motor no está, la mesa de ayuda funciona
  entera.
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

La máquina **sólo necesita Docker** (con el plugin `docker compose`). No hay que instalar Go ni
Node: todo se ejecuta dentro de los contenedores.

### Probarlo en local (desarrollo)

```bash
git clone <el-repositorio> catalina-support
cd catalina-support
docker compose -f dev.yml up -d
docker compose -f dev.yml ps
```

- **La interfaz, en `http://127.0.0.1:11001`.** Así se prueba sin nginx y sin certificados, que es
  lo más cómodo para un primer contacto. El backend queda en `http://127.0.0.1:11002`
  (`GET /api/health` responde).
- **El correo de desarrollo no sale a internet**: va a un buzón de pruebas (Mailpit) que se lee en
  `http://127.0.0.1:11004`. Ahí se ve el enlace de alta de una cuenta sin configurar ningún SMTP.
- **Con datos de ejemplo**, para no entrar a un sistema vacío:

  ```bash
  ./scripts/dev-seed.sh
  ```

  Deja **once cuentas** (`user1@demo.com` … `dev3@demo.com`), **25 tickets** con su historia y sus
  adjuntos. Todas las cuentas entran con la contraseña **`123123123`**. **Cuidado: el guion borra
  los tickets que hubiera** y deja la base en un estado conocido.
- **El motor de IA es opcional** (la primera vez descarga un modelo de ~1,1 GB al volumen):

  ```bash
  docker compose -f ai.yml up -d
  ```

En desarrollo también hay **nginx delante** (para probar con el dominio y el certificado de verdad)
y, detrás de un perfil, un **directorio de pruebas (OpenLDAP)** y un **Keycloak** para probar los dos
caminos de entrada:

```bash
docker compose -f dev.yml --profile auth up -d
```

### Ponerlo a funcionar de verdad (producción)

1. **La configuración de la instalación**, en `config/env/prod.env`, partiendo del ejemplo
   `config/env/prod.env.example`. **Ese archivo no se versiona** y lo que se ponga ahí no aparece en
   la interfaz. Lo imprescindible:
   - `ADMIN_PASSWORD`: la contraseña de **la cuenta de fábrica** (`admin`), que es la puerta para
     entrar la primera vez y la única que entra siempre, sea cual sea el método configurado.
   - `TOKEN_SECRET`: el secreto con el que se firman las sesiones.
   - `POSTGRES_PASSWORD` y los datos de la base.
   - `SMTP_*`: el servidor de correo saliente. **Sin él no salen los correos** y, por tanto, nadie
     puede completar un alta —hoy eso deja como única entrada la cuenta de fábrica—.
   - `PUBLIC_APP_URL` es **opcional**: la dirección pública se configura desde la interfaz. Aquí
     queda como respaldo para una instalación que ya la tuviera puesta.
2. **Construir, publicar y levantar**:

   ```bash
   ./scripts/prod-build.sh                 # construye, publica y levanta los contenedores
   ./scripts/prod-build.sh --no-deploy     # sólo construir y publicar
   ./scripts/prod-build.sh --only backend  # un solo componente
   ```

3. **El esquema de la base**, aplicado en orden hasta la versión que se despliegue. Es
   **transaccional e idempotente** (se puede volver a aplicar sin miedo):

   ```bash
   docker compose -f prod.yml exec -T database \
     psql -v ON_ERROR_STOP=1 -U catalina_support -d catalina_support < backend/migrations/v1.0.0.sql
   ```

   Y las versiones siguientes, en orden (`v1.1.0.sql`, `v1.2.0.sql`…). **Los archivos `_dev` no se
   aplican nunca en producción**: son los datos de ejemplo.
4. **nginx y el certificado**, con el vhost de `config/nginx/`. Los vhosts del repositorio son
   **copias** de `/etc/nginx/conf.d/` (no enlaces): si se cambia uno, hay que copiarlo y recargar.
   ```bash
   sudo cp config/nginx/catalina-support-prod.conf /etc/nginx/conf.d/
   sudo nginx -t && sudo systemctl reload nginx
   ```
5. **Entrar y configurar la instalación** con la cuenta de fábrica, desde **Configuración**:
   - **el nombre y el logo** de la institución, y su **color**;
   - **el método de autenticación** —local, Active Directory o Keycloak—, con sus datos y su botón
     de «Probar la conexión»;
   - **la región horaria**: la zona en la que se leen todas las fechas, aquí y en los correos. Se
     guardan en UTC, así que cambiarla no mueve ningún ticket;
   - **la dirección pública**: la base de los enlaces que van en los correos y de la vuelta de
     Keycloak. Sirve http o https, con puerto si hace falta, y también `localhost`. **Si no es https,
     la pantalla lo avisa** —la contraseña y la sesión viajan sin cifrar— y no bloquea nada: para
     probarlo en local está bien, y para usarlo en serio conviene un certificado;
   - **el idioma de la instalación**, el prefijo de los tickets y su reparto.
6. **Las copias de la base**, que es lo que va en el `cron` del servidor:

   ```bash
   ./scripts/backup-db.sh          # el entorno de desarrollo
   ./scripts/backup-db.sh prod     # producción
   ```

## Las pruebas

Las tres capas están montadas y se ejecutan **en contenedores**:

```bash
docker compose -f dev.yml exec backend  go test ./...            # el backend
docker compose -f dev.yml exec frontend npm test -- --watch=false # la interfaz, unitaria
docker compose -f dev.yml run --rm e2e                            # los recorridos, en un navegador
./scripts/dev-seed.sh                                             # reiniciar el entorno después
```

La capa de interfaz (Playwright) prueba los recorridos de verdad —entrando como cada papel, leyendo
los correos del buzón de pruebas— **en PC y en móvil**.

## Antes de tocar el código

Leer `AGENTS.md`. La **Regla 0** es obligatoria: **primero el documento aprobado, después el
código.** El mapa de qué documento cubre qué está en `AGENTS.md`, y el índice con el estado de cada
documento en `docs/README.md`.

## Licencia

**MIT** (`LICENSE`). Úsalo, estúdialo, modifícalo y distribúyelo.
