# AGENTS.md

## Regla 0 — Documentación antes que código (obligatoria)

**En este proyecto no se escribe código sin un documento aprobado antes.** La documentación va
primero; el código viene después, y sólo después de una aprobación explícita. No hay excepciones
por prisa: si algo merece código, merece antes unas líneas de documento.

El repositorio arranca vacío, así que hoy los documentos son de **diseño** (describen lo que se
quiere construir). Cuando exista código para un área, el documento de esa área pasa a ser
**as-built**: describe lo que el código hace hoy, y si el código cambia, el documento cambia en
el mismo trabajo.

### Estados de un documento

Cada documento de `docs/` empieza con una cabecera con su estado:

| Estado | Significado |
| --- | --- |
| `propuesta` | Escrito, pendiente de aprobación. **No se toca código de esta área.** |
| `aprobado` | Aprobado por el responsable del proyecto. Habilita implementar. |
| `as-built` | El código ya existe y el documento describe lo que hace hoy. |

Cabecera mínima (obligatoria en todos los documentos de `docs/`):

```markdown
> **Estado:** propuesta | aprobado | as-built
> **Última actualización:** AAAA-MM-DD
```

### Ciclo obligatorio

1. **Identificar** el documento de `docs/` que cubre el área a intervenir (tabla de
   correspondencia más abajo). Si el área no tiene documento, hay que crearlo.
2. **Escribir o actualizar el documento** y presentarlo como propuesta. La propuesta dice qué
   documentos se crean o se modifican, qué decide cada uno y qué queda fuera de alcance.
3. **Pedir aprobación explícita.** No se escribe ni una línea de código hasta recibirla. La
   aprobación se da sobre el documento, no sobre una idea en el chat.
4. **Implementar el código** sólo de lo aprobado, sin colar cambios de alcance.
5. **Cerrar el ciclo en el mismo trabajo**: el documento queda `as-built` y refleja lo que el
   código hace de verdad. Si la implementación se desvió de lo aprobado, se explica la
   desviación en el documento antes de darlo por cerrado.

Si al leer un documento se detecta que ya no coincide con el código (o con la realidad), se
**reporta como hallazgo** en la propuesta; no se corrige en silencio.

### Cómo se decide antes de escribir: preguntas

Un documento no se escribe de una vez: se escribe **después de acordar lo que va dentro**. Este es el
procedimiento, y es el que ha funcionado en este proyecto:

1. **Preguntar antes de escribir.** Cada decisión que cambie el diseño se lleva al responsable como
   una **pregunta con opciones**, no como un párrafo ya redactado. Cada opción dice qué se gana y qué
   se pierde, y **la recomendada va primera y marcada como tal**.
2. **Preguntas cortas y en tandas**: de tres a cinco por vez, numeradas, y que se entiendan sin abrir
   ningún archivo. Una pregunta que obliga a leer dos documentos para contestarla está mal hecha.
3. **Nada se da por decidido hasta que lo dice el responsable.** Lo que propone el agente es
   **propuesta**, y va marcado como propuesta también cuando parece evidente.
4. **El responsable corrige, no sólo elige.** Puede cambiar una opción, mezclarlas o decir otra cosa.
   Cuando corrige, se aplica lo que dice y se anota **que corrigió** en el registro del documento.
5. **Los hallazgos se reportan antes de tocar nada.** Si al escribir aparece que un documento
   aprobado no cuadra —una regla que falta, una fila que no existe, un límite que no se puede
   cumplir—, se cuenta primero y se propone la enmienda. La enmienda se anota en la cabecera del
   documento afectado, con la fecha.
6. **Repaso en modo preguntas antes de cerrar.** Antes de aprobar un documento se repasa preguntando
   por lo que puede faltar, no releyéndolo en silencio. Ahí han aparecido los huecos de verdad: el
   aviso que faltaba, la fila que no estaba en la matriz, el límite de subida de nginx y la
   «devolución a Soporte» que un documento prometía y otro no tenía por dónde hacer.
7. **Al cerrar no queda nada abierto.** Lo decidido se mueve al cuerpo del documento; el registro
   del repaso se queda como memoria de **por qué** está así; y la sección de decisiones abiertas
   desaparece. Un documento aprobado con decisiones pendientes es una contradicción.

### Qué no requiere el ciclo completo

No necesitan aprobación previa, aunque si afectan a un documento lo actualizan en el mismo
cambio: corregir una errata, un comentario o un texto de interfaz; reformatear sin cambiar
comportamiento; ajustar documentación interna del propio documento. Cualquier cambio de
**comportamiento, alcance, interfaz, modelo de datos, permisos o flujo** sí pasa por el ciclo
completo.

### Correspondencia entre documentación y código

| Documento | Código que cubre |
| --- | --- |
| `AGENTS.md`, `README.md` | Todo el repositorio |
| `docs/README.md` | Índice y convenciones de `docs/`; no cubre código |
| `docs/propósito-y-alcance.md` | Todavía no cubre código: define el producto (los dos equipos, el modelo de tickets, la numeración y los estados). Se convertirá en as-built cuando exista el módulo de tickets |
| `docs/usuarios-y-permisos.md` | Todavía no cubre código: permisos, acceso, convivencia de los tres métodos de entrada, ciclo de vida de las cuentas y la cuenta de fábrica. Se implementará con los módulos `auth` y `users` |
| `docs/tickets.md` | Todavía no cubre código: modelo de datos, transiciones y los módulos `auth`, `users` y `tickets` con sus endpoints. Se implementarán en ese orden |
| `docs/flujos.md` | Todavía no cubre código: los recorridos paso a paso de los tickets y sus correos. Se implementará con el módulo `tickets` y su parte del frontend |
| `docs/ambientes.md` | Todavía no cubre código: el despliegue, las migraciones, las copias y las pruebas. Se implementará con `scripts/prod-build.sh` y `tests/e2e/` |
| `docs/interfaz-y-experiencia.md` | Todavía no cubre código: la estructura de pantallas, el lenguaje visual y la experiencia de los cuatro papeles. Se implementará con `frontend/src/app/core` y los módulos del frontend |
| `docs/usuarios.md` | Todavía no cubre código: el módulo `users`, su tabla de cuentas y sus endpoints. Se implementará junto con `auth` |
| `docs/arquitectura.md` | La forma del repositorio entero: `backend/**`, `frontend/**`, contenedores (`dev.yml`, `prod.yml`, `config/**`), esquema y migraciones, convenciones de módulo y logs. |
| _(por definir)_ | _(se añade una fila por área cuando exista su documento)_ |

La tabla se mantiene en `AGENTS.md`; el índice con el estado de cada documento vive en
`docs/README.md`. Los nombres de documento se escriben en español y en minúsculas.

## Proyecto

**Catalina-Support** es una mesa de ayuda de dos niveles, simple y sin ruido.

Flujo de dos niveles:

```text
usuario (reporta)  ->  soporte técnico (nivel 1)  ->  desarrollo (nivel 2)
```

- **Soporte técnico** atiende, resuelve y cierra lo que puede resolver.
- **Desarrollo** recibe sólo lo que soporte no puede resolver, ya filtrado y con contexto.

"Sin ruido" es un requisito, no un eslogan: pocos estados, pocos campos obligatorios, nada de
notificaciones ni automatismos que no aporten, y ningún paso que exista sólo porque un producto
grande lo tiene.

## Estado del proyecto

Esqueleto en pie y verificado (2026-09-22): los tres contenedores levantan, nginx publica el
proyecto en `https://dev-catalina-support.calibyou.com` con TLS, el backend Go conecta con
PostgreSQL y responde `GET /api/health` (también a través de nginx), y el frontend Angular 22 se
sirve en desarrollo y pasa sus pruebas. **Todavía no hay ningún módulo ni ninguna tabla.**

El detalle de lo que existe, lo que está sin verificar y lo que falta está en
`docs/arquitectura.md`, sección 13. En resumen:

- `docs/arquitectura.md` (as-built): stack, contenedores, dominios y regla de modularidad.
- `docs/propósito-y-alcance.md` (aprobado): producto, los dos equipos y el modelo de tickets.
- `docs/usuarios-y-permisos.md` (aprobado): permisos, acceso y cuentas.
- `docs/tickets.md` (aprobado): modelo de datos, transiciones y módulos.
- `docs/flujos.md` (aprobado): los recorridos paso a paso y sus correos.
- `docs/ambientes.md` (aprobado): despliegue, migraciones y pruebas.
- `docs/interfaz-y-experiencia.md` (aprobado): la parte visual y de experiencia.
- `docs/usuarios.md` (propuesta): el módulo de usuarios.
- `docs/README.md`: índice de documentación.

La cadena de producto **está completa**. Además, **cada módulo tiene su documento**, escrito justo
antes de implementarlo: `docs/usuarios.md` (propuesta) y `docs/autenticación.md`, que es el
siguiente. Después, `auth` y `users` se implementan juntos. **El código puede empezar** por ahí
(`docs/tickets.md`).

## Estructura de Carpetas

```text
.
├── dev.yml                 # compose de desarrollo (frontend, backend, database)
├── prod.yml                # compose de producción (sin verificar todavía)
├── backend                 # Go: main.go, .air.toml, shared/ (modules/ y migrations/ aún no existen)
├── frontend                # Angular 22: src/app/{core,shared,modules}
├── config
│   ├── dockerfiles         # una imagen por servicio y por entorno
│   ├── env                 # dev.env y prod.env.example (versionados); prod.env no se versiona
│   └── nginx               # vhost de desarrollo, vhost de producción y nginx del contenedor
├── _logs                   # archivos de go-logs (la carpeta se versiona, los .log no)
├── _files                  # adjuntos de los tickets (la carpeta se versiona, los archivos no)
├── docs
├── AGENTS.md
├── README.md
└── LICENSE
```

No crear carpetas de código "por si acaso": cada una nace con el documento que la justifica.

## Comandos

Todo se ejecuta **dentro de los contenedores**: la máquina no tiene Go ni Node instalados (y no
debe tenerlos). `docs/arquitectura.md`, sección 10.

```bash
# Los tres servicios
docker compose -f dev.yml up -d
docker compose -f dev.yml ps
docker compose -f dev.yml logs -f backend
docker compose -f dev.yml down

# Backend (Go 1.27 dentro del contenedor, con air recargando al guardar)
docker compose -f dev.yml exec backend go test ./...
docker compose -f dev.yml exec backend go vet ./...
docker compose -f dev.yml exec backend go mod tidy        # tras añadir dependencias

# Frontend (Node 24 dentro del contenedor)
docker compose -f dev.yml exec frontend npm test -- --watch=false
docker compose -f dev.yml exec frontend npm run build

# Base de datos
docker compose -f dev.yml exec database psql -U catalina_support -d catalina_support -p 11003
```

En desarrollo se entra por **https://dev-catalina-support.calibyou.com**, que es nginx (en la
máquina) delante de los contenedores. Los puertos 11001 y 11002 siguen publicados en
`127.0.0.1` para depurar sin pasar por nginx.

Los dos vhosts son **copias** de `config/nginx/` en `/etc/nginx/conf.d/` (no enlaces), igual que
en Calibyou. Si se cambia un vhost en el repositorio, hay que volver a copiarlo y recargar:

```bash
sudo cp config/nginx/catalina-support-dev.conf  /etc/nginx/conf.d/
sudo nginx -t && sudo systemctl reload nginx
```

El vhost de producción (`https://catalina-support.calibyou.com`) ya está instalado y con
certificado, pero **el despliegue a producción está aplazado hasta cerrar la versión 1.0.0**
(decisión del responsable, 2026-09-22): hasta entonces sólo se trabaja contra el entorno de
desarrollo. Ese dominio responde 503 con un aviso de que está en desarrollo (no un 502 de
nginx, que parecería una web rota), y es lo decidido.

## Reglas para Agentes

- **Regla 0 primero**: documento aprobado antes de escribir código. Sin aprobación no hay código.
- No inventar decisiones de producto, stack ni estructura. Si falta un dato, se pregunta o se
  deja explícitamente como pendiente en el documento.
- La documentación y la comunicación con el responsable del proyecto se escriben en **español**.
- Un documento por área, en `docs/`, con su cabecera de estado. Nada de documentos de diseño
  sueltos en la raíz.
- Mantener `docs/README.md` (índice) y la tabla de correspondencia de este archivo al día en el
  mismo cambio que cree, renombre o jubile un documento.
- **Nada se ejecuta en la máquina**: los comandos van por `docker compose -f dev.yml exec …`. No
  instalar Go ni Node en la máquina ni añadir pasos que los necesiten.
- **Modularidad (regla dura)**: un módulo del frontend sólo llama a `/api/<su nombre>/**`; un
  módulo del backend sólo escribe en sus tablas; `shared` no importa nunca de `modules`. Las
  llamadas del frontend son siempre **relativas**, nunca a un host (`docs/arquitectura.md`, sección 4).
- **Logs**: sólo `github.com/juanky-estevez/go-logs` (`logs.LogInfo/LogSuccess/LogWarning/LogError`).
  Nada de `fmt.Println` ni de `log`/`slog` en el código de la aplicación. Nunca registrar
  contraseñas, tokens ni datos personales.
- **Base de datos**: el esquema vive en `backend/migrations/`, **un archivo por versión**
  (`v1.0.0.sql`, `v1.1.0.sql`…), transaccional e idempotente, aplicado con
  `psql -v ON_ERROR_STOP=1` en orden hasta la versión publicada. Nunca `AutoMigrate` ni un `ALTER`
  a mano (`docs/ambientes.md`, sección 5).
- **Cuenta de fábrica**: el usuario `admin` (papel `administrador`) **no está en la base de datos**:
  vive en la configuración y su contraseña es `ADMIN_PASSWORD`, obligatoria en producción y
  comprobada en cada entrada. Es la única cuenta sin correo, la única que no sigue la política de
  contraseñas y la única que no aparece en los tickets. El valor de producción vive sólo en
  `config/env/prod.env` y **nunca se escribe en la documentación ni en el código**
  (`docs/usuarios-y-permisos.md`, sección 8).
- **Angular moderno**: componentes standalone sin `NgModule`, señales, `@if`/`@for`, rutas
  perezosas por módulo y un único servicio por módulo que centraliza las llamadas HTTP. Los
  componentes no usan `HttpClient` directamente.
- **Interfaz**: **Tailwind v4 con componentes propios** del inventario de
  `docs/interfaz-y-experiencia.md`, sección 6.3. No se añaden bibliotecas de componentes, y un
  componente que no esté en ese inventario se habla antes de construirlo. **Los colores salen de los
  temas** (variables CSS), nunca escritos a mano dentro de un componente. La **accesibilidad**
  —contraste, teclado, foco visible— es requisito y se comprueba al construir cada componente, no al
  final.
- **La sesión viaja en una cabecera `Authorization: Bearer …`**, con el token en `localStorage`: el
  navegador no manda nada por su cuenta. Va con **una regla que no se puede saltar: no se añaden
  scripts de terceros** (analíticas, chats, fuentes con JavaScript), porque cualquiera de ellos puede
  leer el token. Y como contrapartida, no hay CSRF que proteger
  (`docs/usuarios-y-permisos.md`, sección 6).
- **Los errores viajan como claves**, no como texto: el backend responde `{"error": "modulo.clave"}` y
  el frontend lo traduce al idioma de quien lee. Cada clave nueva necesita su texto en español y en
  inglés en el mismo cambio, o la pantalla enseñará una clave.
- **La interfaz no enseña jerga interna**: el usuario nunca ve el ticket interno ni la palabra
  `escalado`, y las acciones se llaman por lo que hacen («Preguntar al usuario», «No es un cambio de
  código») y no por el estado al que llevan.
- Antes de editar, revisar los patrones que ya existen en el repositorio.
- Preferir `rg` para búsquedas.
