# mail

> **Estado:** as-built
> **Última actualización:** 2026-09-27
>
> **Enmendado el 2026-09-27**, al entrar **las menciones de los tickets** (`docs/modules/tickets.md`,
> decisión 58): nace la **plantilla número once**, `ticket.mentioned` —«Te han etiquetado en {{numero}}:
> {{asunto}}»—, que va **a cada persona mencionada** cuando se la menciona por primera vez. **Los correos
> siguen sin llevar el texto del ticket ni el del comentario**: número, asunto y enlace, como los demás.
>
> **Pasa a as-built el 2026-09-26**: el módulo **está entero desde el 2026-09-24** —la tabla con sus
> veinte plantillas, los marcadores, el renderizado, el envío, sus cinco endpoints y el editor—, así
> que este documento describe lo que el código hace. La cabecera se quedó en `aprobado` por un
> descuido de forma, no porque faltara nada; lo cuenta el registro de su implementación, más abajo.
>
> Aprobado por el responsable el 2026-09-23, tras repasarlo en forma de preguntas. Fija la tabla de
> plantillas, los marcadores, **los veintidós textos por defecto** (once correos en español e inglés),
> el envío en HTML con versión de texto y el editor de la pantalla de Configuración.
>
> **Es el primer módulo que se implementa**, porque `auth` lo necesita para mandar el correo de alta.
>
> **Estado de la implementación (2026-09-24): el módulo está entero.** Existen la tabla con sus
> veinte plantillas, los marcadores, el renderizado, el envío y **los cinco endpoints**, probados de
> extremo a extremo contra desarrollo, **y el editor**: la pantalla del módulo `mail` a la que enlaza
> Configuración (decisión 15), con sus dos idiomas al lado, los marcadores que se insertan al
> pulsarlos, la vista previa que renderiza el backend y el botón de la prueba.
>
> **Enmendado el 2026-09-23** (al terminar de implementar su backend): el documento seguía con una
> sección 9 de «decisiones que hay que confirmar» cuyas tres decisiones **ya estaban en el cuerpo y ya
> estaban en el código** (el aviso que no impide guardar, la prueba con `[prueba]` y `PUBLIC_APP_URL`),
> así que se han movido al registro del repaso (9, 10 y 11) y la sección ha desaparecido: un documento
> aprobado no puede quedarse con decisiones abiertas.
>
> **Enmendado el 2026-09-24**, al construir el editor: la **vista previa** que el documento pide no
> tenía por dónde hacerse —los cuatro endpoints eran listar, guardar, restaurar y mandar una prueba—,
> así que estrena un **quinto endpoint**, `POST …/{key}/{language}/preview`, que devuelve el correo ya
> renderizado con los mismos datos de ejemplo que la prueba. Es un `POST` y **acepta el borrador**
> —`{subject, body}`—, no sólo lo que está guardado: la vista previa sirve para decidir **antes** de
> guardar, y además avisa del marcador inventado en el mismo momento en que se escribe. Es una decisión del responsable del
> 2026-09-24: la alternativa era que el frontend sustituyera los marcadores por su cuenta, y eso es
> tener dos versiones de la misma regla. Las otras tres decisiones de esa tanda están en el registro
> (17, 18 y 19).
>
> **Enmendado el 2026-09-23**, al escribir los cuatro endpoints: el documento no tenía
> **ninguna clave de error** del módulo, ni decía qué devuelve cada endpoint al ir bien. Se añaden
> como sección 8, con las decisiones 12 a 15 del registro, todas aprobadas por el responsable.

## 1. Alcance de este documento

Cuenta **cómo se construye el módulo `mail`**: la tabla de plantillas, los marcadores, los diez
correos por defecto con sus textos, el envío, los endpoints y el editor de la pantalla de
Configuración.

**No repite**: los permisos (`docs/usuarios-y-permisos.md`), qué correo se manda en cada momento
(`docs/propósito-y-alcance.md` y `docs/flujos.md`) ni el diseño de la pantalla
(`docs/interfaz-y-experiencia.md`).

Está **aprobado**, así que habilita escribir código (Regla 0). La sección 9
recoge lo que he propuesto yo.

## 2. Qué hace y qué no hace el módulo

**Hace**: guardar los textos, validarlos, rellenarlos con datos, enviarlos por SMTP y dejar que un
administrador los edite, los previsualice y se mande una prueba.

**No hace**: decidir **a quién** se avisa. Quien pide el envío —`auth` o `tickets`— resuelve los
destinatarios y se los pasa. Es lo que mantiene a `mail` sin saber nada de papeles, de repartos ni de
tickets: **recibe una plantilla, un idioma, unos destinatarios y unos datos, y envía**.

## 3. La tabla de plantillas

**`mail_templates`**

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `bigserial` | |
| `key` | `text` | Cuál de los once correos: `ticket.created`, `auth.recovery`… |
| `language` | `text` | `es` o `en` |
| `subject` | `text` | El asunto actual, con sus marcadores |
| `body` | `text` | El **HTML** actual del cuerpo, con sus marcadores |
| `default_subject` | `text` | El asunto **de fábrica**. No se edita nunca |
| `default_body` | `text` | El cuerpo **de fábrica**. No se edita nunca |
| `updated_at` | `timestamptz` | Cuándo se editó por última vez |
| `updated_by_id` | `bigint` | Quién lo editó. Nulo mientras siga el texto de fábrica, y también cuando edita la cuenta de fábrica `admin`, que no está en la tabla de cuentas |

**Único por (`key`, `language`)**: veintidós filas, ni una más —once correos en dos idiomas—.

- **Los veintidós textos se siembran en la migración**, con `INSERT ... ON CONFLICT DO NOTHING`. Eso
  importa: al ser idempotente, **volver a aplicar la migración no pisa lo que un administrador haya
  editado**. Si algún día hay que cambiar un texto de fábrica, será con una migración nueva que lo
  diga, no reescribiendo la siembra.
- **Cada fila guarda las dos versiones**: la que se usa (`subject`, `body`) y la de fábrica
  (`default_subject`, `default_body`). Es lo que hace que **restaurar sea copiar dos columnas** en
  lugar de tener los textos por defecto en dos sitios —el SQL y el código— que acabarían diciendo
  cosas distintas. Cuestan unos kilobytes y quitan un problema.
- **`updated_by_id` dice si un texto lo ha tocado alguien**: nulo es de fábrica. Es la diferencia
  entre «restaurar» y «no hay nada que restaurar».
- **Restaurar** copia las columnas de fábrica sobre las actuales y deja `updated_by_id` en nulo.
- Si algún día hay que cambiar un texto de fábrica, se hace con una migración que actualice
  `default_*` y, **sólo si nadie lo ha tocado**, también el actual.

**Las columnas de fábrica no se pueden escribir desde la API**: el `PUT` de una plantilla sólo toca
`subject`, `body` y `updated_by_id`.

## 4. Los marcadores

Se escriben con llaves dobles y en minúsculas: `{{numero}}`, `{{enlace}}`, `{{nombre}}`.

- **Cada correo tiene su lista**, y el editor la enseña al lado del texto.
- **Se valida al guardar**: un marcador que no esté en la lista de ese correo **no se guarda**, y el
  aviso dice cuál. Es la mitad del trabajo del módulo: un marcador inventado se descubre al guardar,
  no cuando el correo ya ha salido.
- **No se exige que estén todos** (decisión del responsable, sección 8). El editor **avisa** si falta
  uno de los imprescindibles —el enlace, el número— pero deja guardar: hay casos legítimos para
  escribir un aviso sin enlace, y quien edita sabe lo que hace.
- **Los datos se escapan siempre** antes de entrar en el HTML, porque el asunto de un ticket lo
  escribe una persona y puede llevar `<`, `>` o `&`. El escapado es del código, no del texto: no hay
  forma de escribir una plantilla que se salte esto.
- **Los enlaces se construyen con `PUBLIC_APP_URL`**, nunca con un host escrito en el texto: el mismo
  texto tiene que servir en desarrollo y en producción.

## 5. Los once correos por defecto

Estos son los textos que se siembran. **Se revisan aquí, que es donde se pueden corregir sin tocar
código.**

### 5.1 Los ocho avisos de ticket

| Clave | Quién lo recibe |
| --- | --- |
| `ticket.created` | **A quien decida el reparto** de tickets nuevos (el técnico asignado, o nadie, según la configuración) |
| `ticket.escalated` | A quien decida el reparto del ticket interno |
| `ticket.waitingUser` | Al solicitante |
| `ticket.waitingSupport` | Al técnico que lleva el ticket |
| `ticket.resolved` | Al solicitante |
| `ticket.closed` | Al solicitante |
| `ticket.backToSupport` | Al técnico que lleva el ticket |
| `ticket.mentioned` | **A quien acaban de etiquetar** en un comentario (`docs/modules/tickets.md`, decisión 58) |

**Y los observadores reciben los mismos correos de personal que el asignado** (decisión 61 de
`tickets.md`): el del reparto, el del escalado, el de «Desarrollo necesita algo» y el de «el ticket
vuelve a Soporte». Los que van **al solicitante** no se les duplica.

**`ticket.created`** — marcadores: `{{numero}}`, `{{asunto}}`, `{{solicitante}}`, `{{enlace}}`

> **Asunto (ES)**: Ticket nuevo: {{numero}}
> **Asunto (EN)**: New ticket: {{numero}}
> **Cuerpo (ES)**: `<p>Se ha creado el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>Lo ha pedido {{solicitante}}.</p><p><a href="{{enlace}}">Abrir el ticket</a></p>`
> **Cuerpo (EN)**: `<p>Ticket <strong>{{numero}}</strong> has been created.</p><p>{{asunto}}</p><p>Requested by {{solicitante}}.</p><p><a href="{{enlace}}">Open the ticket</a></p>`

**`ticket.escalated`** — marcadores: `{{numero}}`, `{{asunto}}`, `{{motivo}}`, `{{enlace}}`

> **Asunto (ES)**: Escalado: {{numero}}
> **Asunto (EN)**: Escalated: {{numero}}
> **Cuerpo (ES)**: `<p>El ticket <strong>{{numero}}</strong> se ha escalado a Desarrollo.</p><p>{{asunto}}</p><p><strong>Motivo:</strong> {{motivo}}</p><p><a href="{{enlace}}">Abrir el ticket interno</a></p>`
> **Cuerpo (EN)**: `<p>Ticket <strong>{{numero}}</strong> has been escalated to Development.</p><p>{{asunto}}</p><p><strong>Reason:</strong> {{motivo}}</p><p><a href="{{enlace}}">Open the internal ticket</a></p>`

**`ticket.waitingUser`** — marcadores: `{{nombre}}`, `{{numero}}`, `{{asunto}}`, `{{enlace}}`

> **Asunto (ES)**: Necesitamos algo de ti: {{numero}}
> **Asunto (EN)**: We need something from you: {{numero}}
> **Cuerpo (ES)**: `<p>Hola {{nombre}}:</p><p>Para seguir con tu ticket <strong>{{numero}}</strong> ({{asunto}}) necesitamos que nos contestes.</p><p><a href="{{enlace}}">Responder en el ticket</a></p>`
> **Cuerpo (EN)**: `<p>Hi {{nombre}},</p><p>To continue with ticket <strong>{{numero}}</strong> ({{asunto}}) we need you to get back to us.</p><p><a href="{{enlace}}">Reply in the ticket</a></p>`

**`ticket.waitingSupport`** — marcadores: `{{numero}}`, `{{asunto}}`, `{{enlace}}`

> **Asunto (ES)**: Desarrollo necesita algo: {{numero}}
> **Asunto (EN)**: Development needs something: {{numero}}
> **Cuerpo (ES)**: `<p>Desarrollo necesita algo de Soporte para seguir con el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Abrir el ticket interno</a></p>`
> **Cuerpo (EN)**: `<p>Development needs something from Support to continue with ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Open the internal ticket</a></p>`

**`ticket.resolved`** — marcadores: `{{nombre}}`, `{{numero}}`, `{{asunto}}`, `{{enlace}}`

> **Asunto (ES)**: Tu ticket {{numero}} está resuelto
> **Asunto (EN)**: Your ticket {{numero}} is resolved
> **Cuerpo (ES)**: `<p>Hola {{nombre}}:</p><p>El ticket <strong>{{numero}}</strong> ({{asunto}}) está resuelto.</p><p>Si no es así, respóndenos en el ticket o ábrelo de nuevo.</p><p><a href="{{enlace}}">Ver el ticket</a></p>`
> **Cuerpo (EN)**: `<p>Hi {{nombre}},</p><p>Ticket <strong>{{numero}}</strong> ({{asunto}}) is resolved.</p><p>If that is not the case, reply in the ticket or reopen it.</p><p><a href="{{enlace}}">View the ticket</a></p>`

**`ticket.closed`** — marcadores: `{{nombre}}`, `{{numero}}`, `{{enlace}}`

> **Asunto (ES)**: Tu ticket {{numero}} se ha cerrado
> **Asunto (EN)**: Your ticket {{numero}} has been closed
> **Cuerpo (ES)**: `<p>Hola {{nombre}}:</p><p>El ticket <strong>{{numero}}</strong> está cerrado.</p><p>Si vuelve a pasar, puedes reabrirlo.</p><p><a href="{{enlace}}">Ver el ticket</a></p>`
> **Cuerpo (EN)**: `<p>Hi {{nombre}},</p><p>Ticket <strong>{{numero}}</strong> is closed.</p><p>If it happens again, you can reopen it.</p><p><a href="{{enlace}}">View the ticket</a></p>`

**`ticket.backToSupport`** — marcadores: `{{numero}}`, `{{asunto}}`, `{{enlace}}`

> **Asunto (ES)**: {{numero}} ha vuelto a tu bandeja
> **Asunto (EN)**: {{numero}} is back in your queue
> **Cuerpo (ES)**: `<p>El ticket <strong>{{numero}}</strong> ha vuelto a Soporte.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Abrir el ticket</a></p>`
> **Cuerpo (EN)**: `<p>Ticket <strong>{{numero}}</strong> is back with Support.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Open the ticket</a></p>`

**`ticket.mentioned`** — marcadores: `{{numero}}`, `{{asunto}}`, `{{enlace}}`

> **Asunto (ES)**: Te han etiquetado en {{numero}}: {{asunto}}
> **Asunto (EN)**: You were mentioned in {{numero}}: {{asunto}}
> **Cuerpo (ES)**: `<p>Te han etiquetado en el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>Lo sigues desde tu bandeja, en «Observo».</p><p><a href="{{enlace}}">Abrir el ticket</a></p>`
> **Cuerpo (EN)**: `<p>You were mentioned in ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>You are following it from your inbox, under “Watching”.</p><p><a href="{{enlace}}">Open the ticket</a></p>`

### 5.2 Los tres correos de cuenta

| Clave | Quién lo recibe |
| --- | --- |
| `auth.invitation` | La persona a la que se acaba de dar de alta |
| `auth.recovery` | La persona que ha pedido —o a la que le han lanzado— un restablecimiento |
| `auth.passwordChanged` | La persona cuya contraseña ha cambiado |

**`auth.invitation`** — marcadores: `{{nombre}}`, `{{enlace}}`, `{{caducidad}}`

> **Asunto (ES)**: Establece tu contraseña de Catalina Support
> **Asunto (EN)**: Set your Catalina Support password
> **Cuerpo (ES)**: `<p>Hola {{nombre}}:</p><p>Ya tienes cuenta. Establece tu contraseña desde este enlace, que caduca en {{caducidad}}.</p><p><a href="{{enlace}}">Establecer mi contraseña</a></p>`
> **Cuerpo (EN)**: `<p>Hi {{nombre}},</p><p>Your account is ready. Set your password from this link; it expires in {{caducidad}}.</p><p><a href="{{enlace}}">Set my password</a></p>`

**`auth.recovery`** — marcadores: `{{nombre}}`, `{{enlace}}`, `{{caducidad}}`

> **Asunto (ES)**: Restablecer tu contraseña
> **Asunto (EN)**: Reset your password
> **Cuerpo (ES)**: `<p>Hola {{nombre}}:</p><p>Alguien ha pedido restablecer tu contraseña. Si has sido tú, hazlo desde este enlace; caduca en {{caducidad}}.</p><p>Si no has sido tú, no hagas nada: sin abrir el enlace, tu contraseña no cambia.</p><p><a href="{{enlace}}">Restablecer mi contraseña</a></p>`
> **Cuerpo (EN)**: `<p>Hi {{nombre}},</p><p>Someone asked to reset your password. If it was you, use this link; it expires in {{caducidad}}.</p><p>If it was not you, do nothing: your password does not change unless the link is opened.</p><p><a href="{{enlace}}">Reset my password</a></p>`

**`auth.passwordChanged`** — marcadores: `{{nombre}}`, `{{cuando}}`, `{{ip}}`

> **Asunto (ES)**: Tu contraseña ha cambiado
> **Asunto (EN)**: Your password has changed
> **Cuerpo (ES)**: `<p>Hola {{nombre}}:</p><p>Tu contraseña ha cambiado el {{cuando}}, desde la dirección {{ip}}.</p><p>Si no has sido tú, avisa a Soporte cuanto antes.</p>`
> **Cuerpo (EN)**: `<p>Hi {{nombre}},</p><p>Your password changed on {{cuando}}, from the address {{ip}}.</p><p>If it was not you, tell Support as soon as you can.</p>`

## 6. El envío

- **`net/smtp` de la biblioteca estándar**, sin dependencias (igual que Calibyou), con las variables
  `SMTP_*` de `docs/arquitectura.md`, sección 9.
- **HTML con una versión de texto automática**: se manda un correo de dos partes, y la de texto se
  saca del HTML quitando las etiquetas. Un cliente que no pinte HTML enseña algo legible en vez de
  nada.
- **HTML sencillo**: etiquetas de texto —párrafos, negritas, listas, enlaces— y **estilos en línea**.
  Sin imágenes externas, sin fuentes de terceros y sin hojas de estilo: los clientes de correo no
  cargan nada de fuera, y cargarlo dejaría rastro de quién abre el correo.
- **En segundo plano**: el envío no bloquea la petición que lo provocó
  (`docs/arquitectura.md`, sección 9). Un correo tarda lo que tarde el servidor de SMTP, y quien crea
  un ticket no tiene por qué esperarlo.
- **Un fallo de correo no tumba la acción**: si el ticket se crea y el aviso no sale, el ticket
  existe. El fallo va al log con `go-logs`, con la clave de la plantilla y el destinatario.
- **Cada envío se registra en el log** —clave, idioma, destinatario y resultado— y **en ninguna
  tabla**; la dirección del destinatario sí se escribe, por decisión del responsable del 2026-09-23
  (`AGENTS.md`, reglas para agentes). Nunca el contenido del correo ni el enlace que lleva: igual que los intentos de acceso (`docs/modules/auth.md`, sección 12).
- **El idioma lo decide el destinatario**, no quien provoca el correo: si a un equipo se le avisa a
  tres personas y una lee en inglés, esa recibe el texto en inglés. Son tres envíos, no uno.

## 7. Los endpoints

| Método y ruta | Qué hace | Quién |
| --- | --- | --- |
| `GET /api/mail/templates` | Las veinte plantillas: clave, idioma, asunto y cuerpo | Administrador |
| `POST /api/mail/templates/{key}/{language}/preview` | El asunto y el cuerpo **ya renderizados**, con datos de ejemplo. Acepta el **borrador** que se está escribiendo | Administrador |
| `PUT /api/mail/templates/{key}/{language}` | Guarda una plantilla. Valida los marcadores | Administrador |
| `POST /api/mail/templates/{key}/{language}/reset` | Restaura el texto de fábrica | Administrador |
| `POST /api/mail/templates/{key}/{language}/test` | Envía una prueba **al correo de quien la pide** | Administrador |

- **La prueba va a tu propio correo**, no a una dirección que se escriba: así el endpoint no sirve
  para mandar correos a terceros con la plantilla de la institución.
- La prueba lleva datos de ejemplo y lo dice en el asunto (`[prueba]`), para que nadie confunda un
  correo de prueba con uno de verdad.
- **No hay endpoint para enviar en general**: los correos los manda el sistema, no una persona.
- **La vista previa la renderiza el backend**, y no la sustituye el frontend: la sustitución de
  marcadores y el escapado viven en un solo sitio (sección 4), y el día que cambien tiene que cambiar
  también lo que se ve antes de guardar. Es **el mismo motor** que el correo de prueba, con los mismos
  datos de ejemplo: lo que se ve en la pantalla es lo que va a salir.

## 8. Las claves de error y lo que devuelve cada endpoint

Las claves, con el código HTTP con el que viajan. El frontend las traduce; ninguna necesita más
contexto que su propia clave:

| Clave | Cuándo | Código |
| --- | --- | --- |
| `mail.template.notFound` | La plantilla (clave + idioma) no está en la tabla | **404** |
| `mail.template.unknownKey` | Esa clave no es ninguno de los diez correos | **404** |
| `mail.language.unknown` | El idioma no es `es` ni `en` | **404** |
| `mail.subject.required` | El asunto viene vacío | **422** |
| `mail.body.required` | El cuerpo viene vacío | **422** |
| `mail.marker.unknown` | El texto usa un marcador que ese correo no admite | **422** |
| `mail.marker.missing` | Al mandar falta el valor de un marcador que el texto usa | **422**, y en el envío de verdad sólo va al log |
| `mail.test.noEmail` | Se pide una prueba desde una cuenta sin correo: es el caso de la cuenta de fábrica `admin`, la única sin dirección | **422** |
| `mail.smtp.notConfigured` | No hay servidor de correo configurado | **503** |
| `mail.smtp.*` | El servidor de correo no responde o rechaza el envío | **503**, con el detalle en el log y nunca en la respuesta |

- **Un fallo de SMTP no se cuenta por dentro**: `mail.smtp.notConfigured` y `mail.smtp.*` se ven tal
  cual, porque son configuración de la instalación y quien las ve es un administrador. Lo que no
  sale nunca es la respuesta del servidor, que va al log.

Lo que devuelve cada endpoint cuando sale bien. **Siempre devuelve la plantilla afectada**, para que
el editor no tenga que volver a pedirla ni duplicar reglas:

| Endpoint | Devuelve |
| --- | --- |
| `GET /api/mail/templates` | `{"templates": [ … ]}` con las veinte |
| `PUT /api/mail/templates/{key}/{language}` | `{"template": { … }}` y, si falta algún marcador imprescindible, `"missing": ["enlace"]` junto a ella |
| `POST /api/mail/templates/{key}/{language}/reset` | `{"template": { … }}` con el texto de fábrica ya puesto |
| `POST /api/mail/templates/{key}/{language}/test` | `{"sentTo": "quien@demo.com"}` |
| `POST /api/mail/templates/{key}/{language}/preview` | `{"subject": …, "body": …, "text": …}` con los datos de ejemplo ya puestos |

Y **qué lleva cada plantilla**, que es lo que el editor necesita para pintarse con una sola llamada:

| Campo | Qué es |
| --- | --- |
| `key`, `language` | Cuál es |
| `subject`, `body` | Su texto de hoy |
| `edited` | Si su texto **es distinto** del de fábrica. Se compara el texto, no se mira quién lo tocó: así una edición de la cuenta de fábrica —que no deja identificador— también cuenta como editada |
| `updatedAt` | Cuándo se editó por última vez |
| `markers` | Los marcadores que ese correo admite, cada uno con `name` y `essential` |
| `missing` | Los imprescindibles que **no** aparecen hoy en su texto: es el aviso del editor |

- **El aviso lo calcula el backend**, no el frontend: la lista de qué es imprescindible en cada
  correo vive en un solo sitio (sección 4), y duplicarla en el frontend es la forma segura de que
  un día digan cosas distintas.

## 9. Lo que se decidió al repasar este documento

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **El editor** | Campo de texto con botones que envuelven lo seleccionado (negrita, cursiva, lista, enlace), **vista previa** al lado, y **acepta HTML escrito a mano** |
| 2 | **Los marcadores** | Sólo se valida que **no haya marcadores inventados**; no se exige que estén todos |
| 3 | **Restaurar** | Sí, cada plantilla puede volver a su texto de fábrica |
| 4 | **El trato** | De tú, en los dos idiomas |
| 5 | **El formato** | **HTML** con versión de texto automática, sin imágenes ni fuentes externas |
| 6 | **El pie** | Ninguno: el contenido y el enlace |
| 7 | **El registro de envíos** | Sólo en el log, sin tabla |
| 8 | **Los destinatarios** | Los resuelve **quien pide el envío**, no `mail`. El reparto de tickets es cosa de `docs/modules/tickets.md` |
| 9 | **El aviso que falta** | Cuando a una plantilla le falta un marcador imprescindible —el enlace, el número—, **se avisa sin impedir guardar** |
| 10 | **El correo de prueba** | Lleva datos de ejemplo y el asunto empieza por `[prueba]`, para no confundirlo con uno de verdad |
| 11 | **`PUBLIC_APP_URL`** | Es la variable de la que salen todos los enlaces, y va en los env y en `docs/arquitectura.md` (sección 9) |
| 12 | **Las claves de error** | Documentadas en la sección 8, con su código: 404 lo que no existe, 422 lo que no vale y 503 el servidor de correo |
| 13 | **Qué devuelven los endpoints** | Siempre la plantilla afectada, para que el editor no haga una segunda llamada |
| 14 | **El aviso al guardar** | Lo devuelve el backend en `missing`: la regla de qué marcador es imprescindible no se duplica en el frontend |
| 15 | **La pantalla del editor** | Es del módulo `mail` del frontend y Configuración enlaza a ella: así se respeta la regla dura de modularidad sin excepciones |
| 16 | **La prueba desde la cuenta de fábrica** | No se puede: esa cuenta no tiene correo y se responde 422 con `mail.test.noEmail`. El destino sigue siendo siempre el de quien pide la prueba |
| 17 | **La vista previa** | **La renderiza el backend**, con un quinto endpoint y los mismos datos de ejemplo que la prueba: una sola versión de la sustitución de marcadores y del escapado |
| 18 | **Los dos idiomas** | **Dos columnas en PC** —el mismo correo en español y en inglés, uno al lado del otro— **y apilados en móvil**: en un teléfono no caben dos columnas |
| 20 | **Un correo que no sale queda en el log** | **Hallazgo y corrección del 2026-09-27**: los avisos se mandan en segundo plano, y el error de `Send` **no lo miraba nadie** —la acción que lo provocó ya había terminado—, así que una plantilla mal puesta o un servidor de correo caído **no avisaban a nadie, ni al log**. Pasó de verdad: la plantilla del etiquetado existía en la migración, la lista cerrada de `markers.go` no la conocía, `Send` la rechazaba y **el correo no salió sin dejar rastro**. Ahora el fallo se registra, con la clave, el idioma y **los destinatarios** —la dirección sí se registra, el contenido del correo nunca—. Y la lista de `markers.go` tiene las once claves |
| 19 | **Los marcadores** | Se ven los de ese correo, con los imprescindibles marcados, y **se insertan al pulsarlos** en el punto donde esté el cursor: nadie tiene que memorizar `{{solicitante}}` ni escribirlo bien |

## 10. Qué habilita este documento

Con `docs/modules/mail.md` aprobado, **`mail` es el primer módulo que se implementa**, porque `auth`
lo necesita para mandar el correo de alta. El orden de trabajo que propongo:

1. La tabla `mail_templates` y los veinte textos sembrados, en `migrations/v1.0.0.sql`.
2. El rellenado de marcadores y el escapado, con sus pruebas: es la parte que más se rompe.
3. El envío por SMTP con las dos versiones, y el registro en el log.
4. Los cuatro endpoints.
5. El editor en la pantalla de Configuración.
