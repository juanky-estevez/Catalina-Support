# Propósito y alcance

> **Estado:** aprobado
> **Última actualización:** 2026-09-22
>
> Aprobado por el responsable el 2026-09-22 y **enmendado tres veces el mismo día**: al detallar el
> acceso (el cuarto papel `administrador`, el alta de usuarios desde la aplicación, los tickets
> creados en nombre de otro y las tres formas de entrar) y al repasarlo después (el **sexto aviso**
> por correo, el aviso repetido al re-escalar, el límite de **25 MB**, las marcas de editado y
> eliminado, y la cuenta `admin` con la que nace una instalación). La tercera llegó al escribir
> `docs/flujos.md`: la **regla 5** cambia —cuando Desarrollo resuelve el interno, el principal
> **vuelve a Soporte** en lugar de quedar resuelto— y se añade la **regla 8**, la devolución a
> Soporte, y el **séptimo aviso** (el ticket vuelve a la bandeja de Soporte), que evita que un caso
> resuelto por Desarrollo se quede sin explicar al usuario.
>
> Es el primer documento de la cadena y **desbloquea** `usuarios-y-permisos.md`, `tickets.md` y
> `flujos.md`. **No habilita escribir código**: cada área necesita su documento aprobado antes
> (Regla 0 de `AGENTS.md`).

## 1. Qué problema resuelve

Hoy una necesidad llega por donde caiga (WhatsApp, correo, una llamada, un comentario al pasar) y
no queda en ningún sitio. Dos consecuencias:

1. **Se pierde el rastro**: no hay un lugar único donde ver qué se pidió, qué se contestó y en qué
   quedó.
2. **El desarrollo recibe ruido**: quien escribe código se entera de los problemas sin contexto y
   sin saber si ya se intentó algo.

Catalina-Support existe para eso y sólo para eso: **que cada necesidad tenga un sitio único, un
responsable y un historial**, y que al segundo nivel sólo llegue lo que el primero no puede
resolver.

## 2. Para quién

Los tickets los crean **los usuarios que necesitan atención**: quien pide soporte técnico o quien
pide un desarrollo de software. Cuatro papeles, y ninguno más:

| Papel | Qué hace | Contacto con el usuario |
| --- | --- | --- |
| **Usuario** | Crea el ticket y responde cuando Soporte le pregunta. Ve **sólo sus tickets** y puede cerrarlos y reabrirlos. | Es él mismo |
| **Soporte Técnico** (nivel 1) | Primer contacto. Intenta resolver la necesidad y escala lo que no puede. Además da de alta usuarios y **puede crear tickets en nombre de otro usuario**, para quien no tiene acceso a la plataforma. | **Sí**: es la cara visible del sistema |
| **Desarrollo** (nivel 2, escalación) | Actúa cuando Soporte no puede resolver la necesidad o cuando el usuario pide un desarrollo de software. Resuelve y devuelve; **no cierra tickets**. | **No**: no habla con el usuario nunca |
| **Administrador** | No atiende tickets: gestiona cuentas (alta y desactivación de usuarios). | — |

De la tercera columna sale una regla que atraviesa todo el diseño:

> **Desarrollo no tiene contacto con el usuario.** Todo lo que el usuario tenga que saber pasa por
> Soporte Técnico; todo lo que Desarrollo necesite preguntar pasa por Soporte Técnico también.

Desarrollo **sí ve quién es el usuario** (nombre y correo) para tener contexto al resolver; lo que
no hace es escribirle. «Sin contacto» significa sin interlocución, no a ciegas.

**Quién crea un ticket no es siempre quien lo necesita.** Cuando Soporte lo crea en nombre de un
usuario, el ticket guarda las dos cosas por separado: el **solicitante** (el usuario afectado) y
**quién lo creó**. El solicitante es quien recibe los avisos y quien puede cerrarlo.

## 3. Los dos niveles, en una frase

> **Soporte resuelve todo lo que no requiera tocar código; lo que requiera un cambio de código se
> escala a Desarrollo.**

De ahí salen dos reglas que se aplicarán en los documentos de tickets y de flujos:

- Escalar **no** es "pasar la pelota": el ticket llega al nivel 2 con el motivo del escalado, lo
  que ya se probó y el contexto mínimo para no empezar de cero.
- Nada se pierde entre niveles: el historial completo se conserva y se puede leer entero.

## 4. El modelo de tickets

Esta es la parte que más consecuencias tiene, y por eso queda escrita aquí aunque su detalle
(tabla de transiciones, endpoints y pantallas) pertenezca a `tickets.md` y a `flujos.md`.

### Dos tipos de ticket, un solo número

| Tipo | Numeración | Quién lo ve |
| --- | --- | --- |
| **Ticket principal** | `XXX-YYYY-NNNN` | El usuario, Soporte y Desarrollo |
| **Ticket interno** (el de escalación) | `INT-XXX-YYYY-NNNN` — **el mismo número** del principal, con el prefijo `INT-` delante | Soporte y Desarrollo. **El usuario no lo ve** |

- `XXX` es un **prefijo configurable**, `YYYY` el año y `NNNN` un **secuencial por año**.
- El ticket interno **nace cuando Soporte escala** el principal. Es la cara que ve Desarrollo: el
  usuario no tiene que enterarse de que existe.
- Los dos comparten número a propósito: mirando `INT-ACME-2026-0042` cualquiera sabe que su
  principal es `ACME-2026-0042`, sin buscar nada.
- **Hay un solo ticket interno por principal**: se crea en el primer escalado y, en los escalados
  siguientes, se reutiliza (vuelve a `en progreso`). No hay `INT-…-1` ni `INT-…-2`: el historial
  de todos los escalados vive en el mismo ticket.
- El prefijo `XXX` es **uno por instalación** y lo configura quien administra.
- Cada ticket guarda **solicitante** y **creador**, que pueden ser personas distintas: Soporte puede
  abrir un ticket en nombre de un usuario sin acceso. El solicitante es quien recibe los correos.

### Estados

| Tipo | Estados |
| --- | --- |
| Ticket principal | `nuevo`, `en progreso`, `en espera`, `escalado`, `resuelto`, `cerrado` |
| Ticket interno | `nuevo`, `en progreso`, `en espera`, `resuelto`, `cerrado` |

Seis estados para el principal y cinco para el interno, y **ni uno más**: cada estado de más hay
que explicarlo, mantenerlo en las pantallas y decidir quién puede ponerlo.

`en espera` significa esperar a alguien de fuera, y ese alguien no es el mismo en cada tipo:

| Ticket | `en espera` = esperando a |
| --- | --- |
| Principal | El **usuario**: Soporte le pidió información o una prueba |
| Interno | **Soporte**, que es el único interlocutor de Desarrollo |

### Cómo se mueven los dos a la vez

Las reglas que obligan a que los dos tickets no se contradigan:

1. Al **escalar** un principal, se crea el ticket interno (o se reabre el que ya existía) y el
   principal pasa a `escalado`.
2. Si el **interno pasa a `en espera`**, el principal pasa a `en progreso`: lo que se está
   esperando es cosa de Desarrollo, no un problema abierto para el usuario.
3. Si el **principal se vuelve a escalar**, el ticket interno vuelve a `en progreso`.
4. **Soporte puede reabrir el ticket interno** volviendo a escalar: el mismo ticket interno, no uno
   nuevo.
5. Cuando el interno se resuelve, **el principal vuelve a `en progreso`**. Desarrollo ha terminado,
   pero quien cierra el caso con el usuario es Soporte: escribe en el principal qué se hizo y lo
   resuelve.
6. Si el **principal pasa a `en espera`** (Soporte pidió información al usuario), el **interno no se
   toca**: sigue donde esté. Lo que se espera es del usuario, y Desarrollo no tiene por qué pararse
   por eso.
7. **Cierran y reabren el usuario y Soporte.** Desarrollo no cierra tickets: los resuelve y los
   devuelve.
8. **Cerrar el interno sin resolverlo devuelve el principal a `en progreso`**: es la devolución a
   Soporte, para cuando Desarrollo ve que no es un cambio de código.

Regla de lectura: **el usuario nunca ve el ticket interno.** Lo que ve es su principal, con los
avisos que Soporte decida escribirle.

## 5. Qué entra en la 1.0.0

- **Crear un ticket** con asunto, descripción y **adjuntos** (capturas incluidas). Sólo Soporte puede
  crearlo **en nombre de otro usuario**.
- **Numeración automática** `XXX-YYYY-NNNN` con prefijo configurable y secuencial por año.
- **Tres bandejas**: mis tickets (usuario), bandeja de Soporte y bandeja de Desarrollo (que ve los
  tickets internos).
- **Conversación con historial** en los dos tipos de ticket, con autor, fecha y **marcas de editado
  y de eliminado**.
- **Asignación**: un ticket tiene un responsable (persona, no sólo equipo).
- **Escalado** de principal a interno con motivo, reapertura del interno y devolución a Soporte.
- **Los dos ciclos de vida** de la sección 4, con sus reglas de sincronización.
- **Avisos por correo** de los siete eventos de la tabla de abajo, y de ninguno más.
- **Búsqueda y filtros básicos**: por estado, por texto en asunto o descripción y por número.
- **Tres formas de entrar**: correo y contraseña (con el alta que hacen Administrador o Soporte),
  **Active Directory** (alta en el primer acceso) y **Keycloak** (igual que AD). El **correo** es el
  identificador único de una persona, y es con lo que se casa la cuenta local con la del directorio.
- **Contraseñas, sólo con correo y contraseña**: al dar de alta a alguien se le envía un correo para
  que establezca la suya, y quien la olvide la restablece por el mismo camino. Con Active Directory o
  Keycloak **no hay contraseña local**: la gestiona el dominio y la aplicación no interviene.
- **Gestión de usuarios**: alta por Administrador o Soporte. Los roles no se editan desde la
  aplicación.
- **Autenticación y permisos** de los **cuatro** papeles, comprobados en el backend.
- **Una instalación nueva nace con la cuenta `admin`**, que es la que deja el sistema en marcha: dar
  de alta a Soporte y a Desarrollo. Su contraseña viene de la configuración, así que no hay que
  crear la primera cuenta a mano. El detalle está en `docs/usuarios-y-permisos.md`.

Los avisos por correo son estos, y ninguno más:

| Evento | A quién |
| --- | --- |
| Se crea el ticket | A Soporte Técnico |
| Soporte escala a Desarrollo | A Desarrollo |
| Soporte pide información al usuario (el principal pasa a `en espera`) | Al usuario |
| **Desarrollo necesita algo de Soporte** (el interno pasa a `en espera`) | **A Soporte Técnico** |
| **El ticket vuelve a Soporte** (Desarrollo resolvió el interno, o lo cerró sin resolver) | **A Soporte Técnico** |
| El ticket queda resuelto | Al usuario |
| El ticket se cierra | Al usuario |

El aviso de escalado se manda **también cuando se re-escala** un ticket que ya tenía interno: el
trabajo ha vuelto a aparecer, y eso es lo que el aviso significa.

El **cuarto** y el **séptimo** avisan hacia dentro, y son los que evitan que un ticket se quede
parado esperando a alguien que no sabe que le toca: el cuarto cuando Desarrollo necesita algo de
Soporte, y el séptimo cuando el ticket vuelve a la bandeja de Soporte —porque Desarrollo lo
resolvió y hay que explicárselo al usuario, o porque lo devolvió sin resolver—.

Todo lo anterior es el mínimo: si algo se cae, la 1.0.0 sigue siendo útil; si se cae la creación,
la conversación, la escalación o los estados, no hay producto.

## 6. Qué NO entra en la 1.0.0

- **SLA, plazos, métricas e informes.** Nada de tiempos de respuesta ni cuadros de mando.
- **Base de conocimiento** y artículos de ayuda.
- **Crear tickets por correo entrante.** El correo sale, no entra.
- **Chat en vivo** y cualquier cosa en tiempo real más allá de recargar la pantalla.
- **Campos personalizados** y formularios configurables: los campos son los que decida el
  documento de tickets, y punto.
- **Integraciones** (Slack, Teams, GitHub, hilos de conversación en otros sitios).
- **Multi-tenant**, subdominio por cliente y aislamiento por producto.
- **Aplicación móvil.** La web tiene que funcionar en el móvil, eso sí.
- **Adjuntos por encima de 25 MB** o almacenamiento externo: los archivos van a disco en el servidor.
- **Roles a medida y edición de permisos**: los cuatro papeles son fijos, no se crean roles nuevos ni
  se cambian permisos desde la aplicación. La gestión de usuarios se limita al alta.
- **Categoría y prioridad**: no existen. El asunto y la descripción bastan para triar; si más
  adelante hacen falta, entrarán con su documento.

## 7. El criterio para decidir lo dudoso: «sin ruido»

Cuando en los documentos siguientes haya que elegir y este documento no lo resuelva, se aplica esta
regla, que es la del responsable y vale más que cualquier preferencia estética:

- **Pocos estados y pocos campos obligatorios.** Si un campo se puede rellenar mal o deducir, no
  es obligatorio.
- **Nada de automatismos que no ahorren trabajo real**: ni recordatorios, ni campos calculados que
  nadie consulta.
- **Ningún paso que exista sólo porque un producto grande lo tiene.** Si no se puede explicar en
  una frase por qué hace falta, no se hace.

## 8. Cómo sabremos que la 1.0.0 está terminada

Cuando una necesidad real recorra el camino completo sin salirse de la aplicación: un usuario crea
el ticket, Soporte lo atiende, lo escala si hace falta, Desarrollo lo resuelve, Soporte lo cierra y
el historial se puede leer entero meses después. Eso, y no una lista de tareas, es el criterio para
desplegar a producción (aplazado hasta ese momento, ver `docs/arquitectura.md`).

## 9. Qué habilita este documento

Este documento, ya aprobado, fija **qué problema se resuelve, para quién, cómo se numeran y se
mueven los tickets y qué queda fuera**, y desbloquea los tres que dependen de él:

1. `usuarios-y-permisos.md` — los cuatro papeles, los dos equipos y qué puede cada uno.
2. `tickets.md` — modelo de datos, tabla de transiciones de los dos ciclos de vida y la lista
   cerrada de módulos.
3. `flujos.md` — alta, triaje, escalado, resolución, cierre y reapertura, con los avisos por correo.

Ninguno de los tres habilita escribir código por sí solo: cada área necesita su documento aprobado
antes de que se implemente nada (Regla 0 de `AGENTS.md`).
