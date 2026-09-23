# Tickets: modelo de datos, transiciones y módulos

> **Estado:** aprobado
> **Última actualización:** 2026-09-22
>
> Aprobado por el responsable el 2026-09-22 **después de tres repasos en forma de preguntas**: el
> primero fijó las seis decisiones de modelo y de API (las dos tablas, los identificadores, el
> `svg`, el re-escalado, la pantalla del prefijo y el 9999); el segundo, las transiciones que
> faltaban, la edición del ticket, las tres formas de tocar una contraseña y seis huecos técnicos
> —el más serio, que los adjuntos de 25 MB no habrían llegado nunca por el límite de nginx—; y el
> tercero, al escribir `flujos.md`, la **devolución a Soporte** (regla 8) y quién resuelve el
> principal (la regla 5, que cambió).
>
> Es el tercer documento de la cadena. Con él, **`auth`, `users` y `tickets` quedan listos para
> implementarse**, en ese orden. `flujos.md` va después y no bloquea empezar por `auth`.

## 0. El cierre del interno, y un hueco en un documento aprobado

Las ocho reglas de sincronización de `docs/propósito-y-alcance.md` **no dicen quién cierra el
ticket interno**. Se decidió que **se cierra a mano**, y no como consecuencia de cerrar el
principal:

- Lo cierran **Desarrollo o Soporte**, cuando lo dan por terminado.
- **Cerrar el interno no cambia el estado del principal.**
- **Re-escalar el principal lo reabre**, aunque estuviera cerrado (regla 3 de
  `docs/propósito-y-alcance.md`).

Eso **no contradice** el documento aprobado, así que no hay que enmendarlo. Lo que sí falta es un
permiso: la matriz de `docs/usuarios-y-permisos.md` no tiene la acción «**cerrar el ticket
interno**», porque cuando se escribió el interno se cerraba solo. Habría que añadir esa fila.

## 1. Alcance de este documento

Fija **el modelo de datos**, **las transiciones de los dos ciclos de vida** con quién puede cada
movimiento y sus efectos, **la lista cerrada de módulos y endpoints** y **los adjuntos**. No define
los flujos paso a paso con sus correos (`flujos.md`), ni repite los permisos: los da por escritos en
`docs/usuarios-y-permisos.md`.

Está **aprobado**: con él y los dos documentos anteriores, `auth`, `users` y `tickets` se pueden
implementar en ese orden (Regla 0 de `AGENTS.md`). La sección 9 recoge lo que se decidió en los dos
repasos, incluidas las decisiones que corrigieron lo que yo proponía.

## 2. El modelo de datos

### 2.1 Dos tablas: `tickets` e `internal_tickets`

Los principales y los internos viven en **tablas separadas**. Tiene una consecuencia que hay que
asumir a cambio y que se ve en la sección 2.3: **los comentarios, los adjuntos y el historial
apuntan a una tabla o a la otra**, así que cada una de esas tres tablas lleva **dos columnas** —una
por tipo— y una restricción que obliga a que exactamente una esté rellena. A cambio, cada tabla
tiene sólo sus columnas, sin ninguna de relleno.

**`tickets`** — los principales

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `bigserial` | Clave interna. La API no la expone: expone el `number` |
| `number` | `text` | `ACME-2026-0042`. **Único** |
| `number_year` | `int` | Año del número, para ordenar y filtrar |
| `number_seq` | `int` | Secuencial **por año** |
| `subject` | `text` | Asunto. Obligatorio |
| `description` | `text` | Descripción. Obligatoria |
| `state` | `text` | `nuevo`, `en progreso`, `en espera`, `escalado`, `resuelto`, `cerrado` |
| `requester_id` | `bigint` | Quien necesita la atención |
| `created_by_id` | `bigint` | Quien lo creó. Puede no ser el solicitante |
| `assignee_id` | `bigint` | Responsable. Nulo mientras nadie lo coja |
| `subject_edited_at` | `timestamptz` | Nulo mientras nadie haya editado el asunto |
| `description_edited_at` | `timestamptz` | Nulo mientras nadie haya editado la descripción |
| `resolved_at` | `timestamptz` | Cuándo quedó resuelto. Se limpia al reabrir |
| `closed_at` | `timestamptz` | Cuándo se cerró. Se limpia al reabrir |
| `created_at`, `updated_at` | `timestamptz` | |

**`internal_tickets`** — los internos

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `bigserial` | |
| `ticket_id` | `bigint` | Su principal. **Único**: es lo que garantiza un solo interno por principal |
| `number` | `text` | `INT-ACME-2026-0042`. **Único** |
| `state` | `text` | `nuevo`, `en progreso`, `en espera`, `resuelto`, `cerrado` |
| `escalation_reason` | `text` | Por qué se escaló. Obligatorio, y no se edita |
| `created_by_id` | `bigint` | Quien escaló |
| `assignee_id` | `bigint` | Responsable en Desarrollo |
| `resolved_at` | `timestamptz` | |
| `closed_at` | `timestamptz` | |
| `created_at`, `updated_at` | `timestamptz` | |

Decisiones que van con las tablas:

- **El interno no repite nada del principal.** Ni el solicitante, ni el asunto, ni la descripción:
  su contenido es **el motivo del escalado**, y todo lo demás se lee de su principal. Aquí es donde
  las dos tablas se portan mejor que una: sin columnas vacías ni de relleno.
- **El número es inmutable**: se guarda entero en `number`, y no se recompone desde el prefijo
  actual. Si mañana cambia el prefijo, los tickets viejos conservan el suyo.
- **No hay `deleted_at`** en ninguna de las dos: no existe la acción de borrar un ticket.
- El interno guarda **su propio número** (`INT-…`) en vez de calcularlo al vuelo, porque aparece en
  todas las pantallas y en los correos.
- **El asunto y la descripción se pueden editar** (el solicitante y Soporte, en el principal), y cada
  uno deja su marca: `subject_edited_at` y `description_edited_at`, una por campo para saber **qué**
  se cambió sin tener que leer el historial. No se guarda el texto anterior, igual que en los
  comentarios.
- **El motivo del escalado del interno no se edita.** Es la justificación de por qué se escaló, y
  reescribirla es reescribir la historia.

### 2.2 La numeración, sin carreras

El secuencial por año se lleva en una tabla de una fila por año:

```sql
ticket_number_counters (year int primary key, last_number int not null)
```

Y se incrementa **en la misma transacción** que crea el ticket, con una sola sentencia atómica:

```sql
INSERT INTO ticket_number_counters (year, last_number) VALUES (2026, 1)
ON CONFLICT (year) DO UPDATE SET last_number = ticket_number_counters.last_number + 1
RETURNING last_number;
```

Por qué así y no con `MAX(number) + 1`: dos tickets creados a la vez leerían el mismo máximo y
sacarían el mismo número. Por qué no una `SEQUENCE` por año: obliga a crear objetos en la base cada
1 de enero, y eso es un paso manual que se olvida.

`number_seq` se rellena a **cuatro dígitos** (`0042`), pero **no se trunca**: a partir del 9999 el
número crece a cinco dígitos y sigue siendo único y ordenable.

**Sólo los tickets principales consumen el secuencial.** El interno no gasta número: el suyo es el
de su principal con `INT-` delante.

El prefijo admite **mayúsculas y dígitos, de 2 a 8 caracteres** (`^[A-Z0-9]{2,8}$`): `CS` y `ACME`
valen, `acme-2026` no. Se valida al guardarlo, no al usarlo.

**El prefijo lo cambia un Administrador desde la aplicación**, no es configuración del servidor.
Vive en una tabla de **una sola fila**:

```sql
ticket_number_settings (id boolean primary key default true check (id), prefix text not null,
                        updated_at timestamptz, updated_by_id bigint)
```

Una fila suelta y no una tabla de «ajustes» con clave y valor: una tabla así acaba siendo el cajón
donde entra todo. Con una fila se ve de un vistazo **qué** es configurable, y añadir otro ajuste
exige tocar el esquema, que es justo la conversación que conviene tener. La fila guarda **quién y
cuándo** la cambió por última vez; no se guarda un historial de cambios, y queda dicho para que
nadie lo dé por hecho.

### 2.3 Comentarios, adjuntos e historial

Las tres tablas que siguen apuntan a **un principal o a un interno**: llevan `ticket_id` y
`internal_ticket_id`, y la restricción de que **exactamente una** esté rellena. Es el precio de
tener dos tablas de tickets, y se paga aquí y sólo aquí.

**`ticket_comments`**

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `bigserial` | |
| `ticket_id` | `bigint` | El principal, o nulo |
| `internal_ticket_id` | `bigint` | El interno, o nulo. Nunca los dos |
| `author_id` | `bigint` | Quien escribe |
| `body` | `text` | Obligatorio |
| `edited_at` | `timestamptz` | Nulo mientras nadie lo haya editado |
| `deleted_at` | `timestamptz` | Nulo mientras siga vivo. Borrado **lógico** |
| `created_at` | `timestamptz` | |

- **Se pueden editar, pero sólo por quien lo escribió**, y el comentario queda marcado como
  editado (`edited_at`). Nadie edita lo que dijo otro.
- **No se guarda la versión anterior.** La marca dice *que* se editó, no *qué* decía antes; si
  hiciera falta el texto previo, sería otro documento.
- **El autor puede borrar el suyo, y queda la marca.** La fila no desaparece: la conversación
  muestra que ahí hubo un comentario, con su autor y su fecha, y **el texto se vacía**. Se vacía a
  propósito: si alguien pegó una contraseña o un dato que no debía, borrar tiene que borrar de
  verdad, no esconder.
- **Borrar un comentario no borra sus adjuntos**: siguen colgando del ticket y se descargan como
  cualquier otro. Un adjunto puede ser la prueba de algo, y quien borra su texto no está pidiendo
  que desaparezca el archivo.
- **No hay notas internas dentro del principal.** Lo que el usuario no debe leer vive en el ticket
  interno, que es exactamente para lo que existe. Añadir notas internas sería un segundo mecanismo
  para lo mismo.

**`ticket_attachments`**

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `bigserial` | |
| `ticket_id` | `bigint` | El principal, o nulo |
| `internal_ticket_id` | `bigint` | El interno, o nulo. Nunca los dos |
| `comment_id` | `bigint` | Nulo si el adjunto va en la descripción inicial |
| `uploaded_by_id` | `bigint` | |
| `filename` | `text` | Nombre original, tal como lo subió la persona |
| `stored_name` | `text` | Nombre en disco (generado, único) |
| `content_type` | `text` | |
| `size_bytes` | `bigint` | |
| `created_at` | `timestamptz` | |

- Los archivos **no se guardan en la base de datos**, sino en disco, en la carpeta que indique
  `FILES_PATH`. En los contenedores es un volumen montado (`_files`), igual que `_logs`: los adjuntos
  sobreviven a los reinicios y se pueden copiar desde la máquina. `dev.yml` y `prod.yml` lo montan y
  `config/env/*.env` define la variable.
- **Límite de subida en nginx**: los vhosts llevan `client_max_body_size 30m;`. Sin él, nginx usa
  **1 MB** por defecto y cualquier adjunto de más de un mega se rechazaría con un 413 antes de llegar
  al backend. El 30 deja margen sobre los 25 MB del archivo.
- **Se descargan por la API, nunca por una ruta estática**: el adjunto de un ticket interno no puede
  ser alcanzable por el solicitante, y eso sólo se garantiza si cada descarga pasa por la
  comprobación de permisos.
- Límites: **25 MB por archivo** y **sin límite de cantidad**. Lo que sí es una lista cerrada es
  la **extensión**: `pdf`; imágenes `png`, `jpg`, `jpeg`, `gif`, `webp`; Office `doc`, `docx`,
  `xls`, `xlsx`, `ppt`, `pptx`, `odt`, `ods`, `odp`; texto `txt`, `csv`, `log`, `md`, `css`; y
  `zip`, `rar`. Lo que no esté en la lista se rechaza: no se acepta «cualquier cosa» y luego se
  mira.
- **`svg` queda fuera a propósito**, aunque sea una imagen: un SVG puede llevar código dentro y se
  abre en el navegador. Los demás formatos de imagen no.
- **Sin antivirus** en la 1.0.0: queda dicho para que nadie lo dé por hecho.

**`ticket_history`**

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `bigserial` | |
| `ticket_id` | `bigint` | El principal, o nulo |
| `internal_ticket_id` | `bigint` | El interno, o nulo. Nunca los dos |
| `actor_id` | `bigint` | Quién provocó el cambio. Nulo si lo hizo el sistema |
| `event` | `text` | `creado`, `editado`, `asignado`, `estado`, `escalado`, `resuelto`, `cerrado`, `reabierto` |
| `from_state`, `to_state` | `text` | Nulos cuando el cambio no es de estado |
| `detail` | `text` | Texto libre: a quién se asignó, por qué se escaló… |
| `created_at` | `timestamptz` | |

El historial es **lo que hizo el sistema**; los comentarios son **lo que dijeron las personas**. La
línea de tiempo de la pantalla es la mezcla de los dos ordenada por fecha, y el historial no lo
escribe nadie a mano: lo genera cada transición.

### 2.4 Restricciones

- `tickets.state` sólo admite los seis estados del principal; `internal_tickets.state`, los cinco
  del interno. **El interno no puede estar `escalado`**: el interno *es* la escalación.
- `internal_tickets.ticket_id` es **único**: un solo interno por principal, garantizado por la base
  y no por el código.
- En `ticket_comments`, `ticket_attachments` y `ticket_history`, **exactamente una** de las dos
  columnas de destino está rellena.
- `internal_tickets.escalation_reason` no puede estar vacío: escalar sin decir por qué es lo que
  convierte una escalación en un «pásalo tú».

### 2.5 Índices

| Índice | Para qué |
| --- | --- |
| `tickets(number)` único | Buscar por número, que es como la gente habla de un ticket |
| `internal_tickets(number)` único | Lo mismo, para los internos |
| `internal_tickets(ticket_id)` único | El interno de cada principal, y sólo uno |
| `tickets(state)` | Las bandejas de Soporte |
| `internal_tickets(state)` | La bandeja de Desarrollo |
| `tickets(requester_id)` | «Mis tickets» |
| `tickets(assignee_id)` | Los tickets de una persona |
| `ticket_comments(ticket_id, created_at)` y `(internal_ticket_id, created_at)` | La conversación en orden |
| `ticket_history(ticket_id, created_at)` y `(internal_ticket_id, created_at)` | La línea de tiempo en orden |
| `ticket_attachments(ticket_id)` y `(internal_ticket_id)` | Los adjuntos de un ticket |

La búsqueda por texto usa `ILIKE` sobre asunto y descripción. Es suficiente para el volumen de una
mesa de ayuda interna; cuando deje de serlo, se pasa a índices de texto completo con su propio
cambio.

## 3. Las transiciones

### 3.1 Ticket principal

| Desde | Hasta | Quién | Efectos |
| --- | --- | --- | --- |
| — | `nuevo` | Usuario (el suyo) o Soporte (en nombre de otro) | Se asigna el número; historial `creado`; correo a Soporte |
| `nuevo` | `en progreso` | Soporte | Al asignárselo a alguien o al empezar |
| `nuevo` | `escalado` | Soporte | Crea el interno. Exige motivo. Correo a Desarrollo |
| `en progreso` | `en espera` | Soporte | Soporte pide información al usuario. **Correo al usuario** |
| `en espera` | `en progreso` | Soporte o el usuario (al responder) | |
| `en progreso` | `escalado` | Soporte | Crea el interno. Exige motivo. Correo a Desarrollo |
| `en espera` | `escalado` | Soporte | Igual que el anterior |
| `en progreso` | `resuelto` | Soporte | Resuelto **sin escalar**, o después de que Desarrollo resolviera el interno. Correo al usuario |
| `en espera` | `resuelto` | Soporte | Llegó lo que se esperaba. **No hace falta pasar por `en progreso`** |
| `nuevo`, `en progreso` o `en espera` | `cerrado` | Soporte o el usuario (el suyo) | **Cerrar sin resolver**: duplicados, pruebas, tickets creados por error |
| `escalado` | `en progreso` | Sistema | Lo provoca el interno al pasar a `en espera` (regla 2) |
| `escalado` | `en progreso` | Sistema | Lo provoca el interno al resolverse: **el caso vuelve a Soporte** (regla 5), que explica al usuario qué se hizo y lo resuelve |
| `resuelto` | `escalado` | Soporte | Re-escalar sin reabrir: el interno vuelve a `en progreso` |
| `resuelto` | `cerrado` | Soporte o el usuario | **No toca el interno**: ese se cierra aparte (sección 3.2). **Correo al usuario** |
| `cerrado` | `en progreso` | Soporte o el usuario | Reabrir. Volver a escalar exige pasar por aquí |

Cuatro reglas que salen de la tabla:

- **Un ticket cerrado no se escala.** Para volver a escalarlo, primero se reabre.
- **Nadie salta de `nuevo` a `resuelto`**: si se resolvió, es que alguien lo trabajó.
- **Cerrar sin resolver sí se puede**, desde cualquier estado abierto, y la diferencia queda en el
  historial: no hay un estado `cancelado` porque «seis estados y ni uno más», y porque un ticket
  cerrado sin resolución no se comporta distinto de uno cerrado después de resolverlo.
- **Al reabrir se limpian `resolved_at` y `closed_at`**, para que la pantalla no diga «resuelto el»
  con una fecha vieja. El historial conserva cuándo pasó cada cosa.

### 3.2 Ticket interno

| Desde | Hasta | Quién | Efectos |
| --- | --- | --- | --- |
| — | `nuevo` | Sistema | Lo crea el escalado del principal, con su motivo |
| `nuevo` | `en progreso` | Desarrollo | Al asignárselo o al empezar |
| `en progreso` | `en espera` | Desarrollo | Desarrollo necesita algo de Soporte. El principal pasa a `en progreso` (regla 2). **Correo a Soporte** |
| `en espera` | `en progreso` | Soporte o Desarrollo | Soporte responde |
| `en progreso` | `resuelto` | Desarrollo | **El principal vuelve a `en progreso`**, a manos de Soporte (regla 5). **Aviso a Soporte** |
| `en espera` | `resuelto` | Desarrollo | Llegó lo que se esperaba. **No hace falta pasar por `en progreso`** |
| `nuevo`, `en progreso` o `en espera` | `cerrado` | Desarrollo o Soporte | **Cerrar sin resolver** (por ejemplo, escalado por error). **Devuelve el principal a `en progreso`** (regla 8). **Aviso a Soporte** |
| `resuelto` | `cerrado` | Desarrollo o Soporte | **A mano**, cuando lo dan por terminado. No cambia nada del principal |
| `cerrado` | `en progreso` | Soporte | Reabrir el interno: **es volver a escalar**. El principal pasa a `escalado` (regla 3) |

El interno **no tiene `escalado`**: el interno *es* la escalación. Y su cierre es **manual**, no
consecuencia del principal: se cierra cuando Desarrollo o Soporte lo dan por terminado, y re-escalar
lo devuelve a `en progreso` aunque estuviera cerrado.

**Cerrar el interno sin resolverlo devuelve el principal a `en progreso`** (regla 8): es la
devolución a Soporte. Si lo dejara en `escalado`, el ticket se quedaría en el estado que significa
«lo tiene Desarrollo» sin que nadie de Desarrollo lo esté mirando.

**Y resolver el interno no devuelve el caso al usuario por sí solo**: el principal pasa a
`en progreso` para que Soporte escriba qué se hizo. La explicación que lee el usuario no puede ser
el motivo del escalado ni las notas de Desarrollo.

## 4. Adjuntos: de quién son

Los adjuntos siguen la misma regla que los comentarios, que es lo que dice la matriz aprobada: **si
puedes comentar en un ticket, puedes adjuntar en él**. No es un permiso nuevo, es la aplicación del
que ya existe.

| Ticket | Quién adjunta | Quién descarga |
| --- | --- | --- |
| Principal | El solicitante y Soporte | El solicitante, Soporte y Desarrollo (Desarrollo en sólo lectura) |
| Interno | Soporte y Desarrollo | Soporte y Desarrollo |

El solicitante **nunca** alcanza un adjunto del interno: ni por la API ni por una ruta de archivos,
porque no existe tal ruta.

**Cómo se sirven**: los tipos que se pueden previsualizar (`png`, `jpg`, `jpeg`, `gif`, `webp` y
`pdf`) se pueden servir en línea para verlos dentro del ticket. **Todo lo demás se sirve siempre
como descarga**, con `application/octet-stream` y `Content-Disposition: attachment`.

El caso que obliga a esta regla es el **`svg`**: se admite como adjunto, pero **sólo como
descarga**, porque un SVG puede llevar código dentro y el navegador lo ejecuta al abrirlo en línea.
Descargado, es un archivo más.

## 5. Los módulos y sus endpoints

Tres módulos, y no más:

| Módulo | Qué cubre |
| --- | --- |
| `auth` | Entrar, salir, quién soy, contraseñas |
| `users` | Cuentas, papeles, desactivación, origen |
| `tickets` | Principales, internos, comentarios, adjuntos y su historial |

Los dos tipos de ticket viven en **un solo módulo** (`tickets`): son la misma cosa contada de dos
maneras, y separarlos obligaría a que dos módulos compartieran tablas, que es justo lo que la regla
de modularidad prohíbe.

### `auth`

| Método y ruta | Qué hace | Quién |
| --- | --- | --- |
| `POST /api/auth/login` | Entra. Acepta **correo o usuario** (por la cuenta de fábrica `admin`). Resuelve el origen: `local` valida en la base, `ad` contra el directorio | Cualquiera |
| `GET /api/auth/keycloak/start` | Empieza el acceso por Keycloak (OIDC) | Cualquiera |
| `GET /api/auth/keycloak/callback` | Vuelve de Keycloak, vincula o crea la cuenta y abre sesión | Keycloak |
| `POST /api/auth/logout` | Cierra sesión y borra la cookie | Autenticado |
| `GET /api/auth/me` | Quién soy: papel, nombre y correo | Autenticado |
| `POST /api/auth/password/forgot` | Pide el correo de recuperación | Cualquiera |
| `POST /api/auth/password/reset` | Establece la contraseña con el enlace | Cualquiera con enlace válido |
| `POST /api/auth/password/change` | Cambia la contraseña **estando dentro**. Sólo cuentas `local`: con AD o Keycloak la gestiona el dominio | Autenticado |

### `users`

| Método y ruta | Qué hace | Quién |
| --- | --- | --- |
| `GET /api/users` | Lista de usuarios (nombre y correo) | Soporte y Administrador |
| `POST /api/users` | Alta. Acepta `origin`; si es de directorio, valida que exista antes de crear | Soporte (sólo `usuario`) y Administrador |
| `GET /api/users/{id}` | Detalle de una cuenta | Administrador |
| `POST /api/users/{id}/deactivate` | Desactiva | Soporte y Administrador |
| `POST /api/users/{id}/activate` | Reactiva | Soporte y Administrador |
| `POST /api/users/{id}/origin` | Cambia el origen de la cuenta | Administrador |
| `POST /api/users/{id}/reset-password` | Lanza un reseteo: el usuario recibe el correo de recuperación. Nadie ve ni asigna la contraseña | Soporte y Administrador |

### `tickets`

| Método y ruta | Qué hace | Quién |
| --- | --- | --- |
| `GET /api/tickets` | Bandeja. Filtros: `type`, `state`, `q` (texto o número) y paginación | Según la matriz: cada uno ve lo que le toca |
| `POST /api/tickets` | Crea un principal. `requesterId` opcional, para crearlo en nombre de otro | Usuario, Soporte, Desarrollo |
| `GET /api/tickets/{number}` | Detalle, con comentarios, adjuntos e historial | Quien pueda verlo |
| `POST /api/tickets/{number}/comments` | Comenta | Quien pueda comentar |
| `PATCH /api/tickets/{number}/comments/{id}` | Edita un comentario. Sólo su autor; deja la marca de editado | El autor del comentario |
| `DELETE /api/tickets/{number}/comments/{id}` | Borra un comentario: vacía el texto y deja la marca de eliminado | El autor del comentario |
| `POST /api/tickets/{number}/attachments` | Sube un adjunto (`multipart`) | Quien pueda comentar |
| `GET /api/tickets/{number}/attachments/{id}` | Descarga, con comprobación de permisos | Quien pueda ver el ticket |
| `PATCH /api/tickets/{number}` | Edita el asunto o la descripción, con su marca | El solicitante y Soporte en el principal |
| `POST /api/tickets/{number}/assign` | Asigna responsable | Soporte; Desarrollo en los internos |
| `POST /api/tickets/{number}/state` | Mueve el estado: `en progreso`, `en espera`, `resuelto`, `cerrado` | Quien pueda ese movimiento (sección 3) |
| `POST /api/tickets/{number}/escalate` | Escala. Exige motivo. Crea el interno o lo reabre | Soporte |
| `POST /api/tickets/{number}/reopen` | Reabre un principal cerrado | Soporte o el solicitante |
| `GET /api/tickets/number-prefix` | Lee el prefijo de la numeración y quién lo cambió | Administrador |
| `PUT /api/tickets/number-prefix` | Cambia el prefijo. No toca los números ya emitidos | Administrador |

- **La API se dirige por número, no por `id`.** `ACME-2026-0042` es lo que la gente dice en voz alta
  y lo que aparece en los correos; el `id` interno no sale de la base.
- **El interno se lee con el mismo endpoint**, pasando su número (`INT-ACME-2026-0042`): es un
  ticket, y no necesita una ruta propia.
- **El único borrado que existe es el de un comentario, y es lógico**: la fila se queda con su
  marca. Un ticket no se borra, y un adjunto tampoco.
- **El prefijo vive en el módulo `tickets`**, no en `users`: es la numeración de los tickets, y así
  la pantalla que lo cambia está en el módulo de tickets del frontend y la regla de modularidad no
  se rompe. La ve sólo un Administrador.
- Las listas van paginadas. El formato de error es el único de la API (`{"error": "..."}`), ya
  definido en `shared/httpx`.

## 6. La migración

Todo el esquema entra en **`backend/migrations/v1.0.0.sql`**, el archivo único, transaccional e
idempotente que manda `docs/arquitectura.md`. Tablas nuevas:

| Tabla | Qué guarda |
| --- | --- |
| `tickets` | Los tickets principales |
| `internal_tickets` | Los internos, uno por principal |
| `ticket_comments` | La conversación de los dos tipos |
| `ticket_attachments` | Los adjuntos de los dos tipos |
| `ticket_history` | Lo que hizo el sistema en los dos tipos |
| `ticket_number_counters` | El secuencial de cada año |
| `ticket_number_settings` | Una fila: el prefijo de la numeración |

La migración **siembra el prefijo con `CS`**, para que se pueda crear el primer ticket sin pasar
por la pantalla de configuración. Si no gusta, se cambia antes de crear el primero.

La migración **no siembra ninguna cuenta**: la del administrador de fábrica (`admin`) la crea el
**arranque del backend** desde `ADMIN_PASSWORD`, porque la migración no puede leer variables de
entorno (`docs/usuarios-y-permisos.md`, sección 8). Y no hay más datos de ejemplo: los tickets los
crea la gente.

## 7. Qué NO entra en la 1.0.0

- **Borrar un comentario de otro** y **recuperar uno borrado**: el texto vaciado no se guarda.
- **Editar o borrar adjuntos** ya subidos.
- **Historial de versiones**: no se guarda el texto anterior de un comentario ni de un asunto editado.
- **Editar el motivo del escalado** de un ticket interno.
- **Borrar tickets.**
- **Antivirus** y análisis del contenido de los adjuntos.
- **Almacenamiento externo** (S3 y similares).
- **Búsqueda de texto completo** con índices dedicados: se usa `ILIKE` hasta que duela.
- **Exportar** tickets a CSV o PDF.
- **Respuestas predefinidas** y plantillas.
- **Campos personalizados**, categoría y prioridad.
- **Tiempo real** (websockets) y contadores en vivo.
- **SLA, métricas e informes.**
- **Fusión y duplicado de tickets.**

## 8. Módulos de frontend

| Módulo | Pantallas |
| --- | --- |
| `tickets` | Mis tickets, bandeja de Soporte, bandeja de Desarrollo, detalle del ticket (conversación, adjuntos, historial), alta de ticket, el formulario de escalado y la pantalla del prefijo de numeración (sólo Administrador) |
| `users` | Lista de usuarios, alta, desactivación y cambio de origen |
| — (`core`) | Entrar, salir, recuperar contraseña y las guardas de sesión, como manda `docs/usuarios-y-permisos.md` |

La bandeja de Desarrollo y la de Soporte son **la misma pantalla** con distinto filtro: separarlas
sería duplicar la tabla, los filtros y la paginación para cambiar una condición.

## 9. Lo que se decidió en los repasos

### Primer repaso: el modelo y la API

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **Tablas** | **Dos tablas** (`tickets` e `internal_tickets`), con dos columnas de destino en comentarios, adjuntos e historial |
| 2 | **Identificadores** | `bigserial` por dentro; la API se dirige por el número del ticket |
| 3 | **`svg`** | **Se admite**, pero siempre como descarga, nunca en línea |
| 4 | **Re-escalar desde `resuelto`** | Directo, sin obligar a reabrir. Desde `cerrado`, hay que reabrir primero |
| 5 | **El prefijo del número** | **Pantalla de administración**, con una fila en `ticket_number_settings` |
| 6 | **El secuencial pasa de 9999** | El número crece a cinco dígitos; nunca se repite ni se trunca |

### Segundo repaso: las transiciones, la edición y los huecos

| # | Decisión o hueco | Quedó así |
| --- | --- | --- |
| 7 | **Cerrar sin resolver** | Se puede, desde cualquier estado abierto (duplicados, pruebas). No hay estado `cancelado` |
| 8 | **Resolver desde `en espera`** | Permitido, en los dos tipos de ticket, sin pasar por `en progreso` |
| 9 | **Editar el ticket** | Asunto **y** descripción, con marca por campo (`subject_edited_at`, `description_edited_at`). El motivo del escalado no se edita |
| 10 | **Contraseñas** | El usuario `local` la cambia cuando quiere; al darse de alta la crea desde el correo; un administrador puede **lanzar un reseteo**, que también va por correo |
| 11 | **Límite de subida de nginx** | `client_max_body_size 30m;` en los dos vhosts: sin él, nginx cortaba en 1 MB |
| 12 | **Los adjuntos en disco** | Variable `FILES_PATH` y volumen `_files` en `dev.yml` y `prod.yml`, como `_logs` |
| 13 | **El correo** | Apartado nuevo en `docs/arquitectura.md` (sección 9): `shared/mail` con `net/smtp`, variables `SMTP_*`, y un fallo de correo **no** tumba la acción |
| 14 | **Formato del prefijo** | Mayúsculas y dígitos, de 2 a 8 (`^[A-Z0-9]{2,8}$`) |
| 15 | **Intentos de login** | Zona `limit_req` para el login y para pedir el correo de recuperación, en los dos vhosts |
| 16 | **Fechas al reabrir** | `resolved_at` y `closed_at` se limpian al reabrir; el historial conserva cuándo pasó cada cosa |

### Tercer repaso: la devolución y el cierre (al escribir `flujos.md`)

| # | Decisión | Quedó así |
| --- | --- | --- |
| 17 | **La devolución a Soporte** | Cerrar el interno sin resolverlo **devuelve el principal a `en progreso`**: es la **regla 8**, nueva en `docs/propósito-y-alcance.md` |
| 18 | **Quién resuelve el principal** | **Soporte**, siempre, y con su explicación para el usuario. La **regla 5** cambia: cuando Desarrollo resuelve el interno, el principal vuelve a `en progreso` en lugar de quedar resuelto |
| 19 | **Comentar en un ticket cerrado** | No se puede: para seguir hablando hay que reabrirlo |
| 20 | **Plazo para reabrir** | Ninguno |
| 21 | **El formato de los correos** | Texto plano, con el número, el asunto y el enlace |

Los huecos del segundo repaso (11 al 15) están aplicados en los vhosts, en `dev.yml`/`prod.yml`, en
`config/env/*.env` y en `docs/arquitectura.md`, no sólo escritos aquí.

## 10. Qué habilita este documento

Aprobar `docs/tickets.md` cierra la lista de módulos y deja **`auth`, `users` y `tickets`** listos
para implementarse, en ese orden:

1. `auth` — sin sesión no hay nada que proteger.
2. `users` — Soporte necesita dar de alta a la gente, y el prefijo se configura.
3. `tickets` — el producto.

`flujos.md` queda después: describe el recorrido paso a paso de cada caso (alta, triaje, escalado,
resolución, cierre y reapertura) con sus correos, y no bloquea empezar por `auth`.
