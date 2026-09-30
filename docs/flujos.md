# Flujos

> **Estado:** as-built
> **Última actualización:** 2026-09-26
>
> **Enmendado el 2026-09-26**, al repartir las listas de tickets: donde estos recorridos decían «la
> bandeja de Soporte» o «la bandeja de Desarrollo» ahora hay **dos sitios** —**Mis tickets** (lo mío) y
> las listas del todo, **Tickets principales** y **Tickets internos**—, y **asignarse un ticket se hace
> dentro del ticket** y no desde la lista (`docs/modules/tickets.md`, decisiones 50 a 53). **Los
> recorridos y los correos no cambian**: cambia dónde se ve cada cosa.
>
> **Pasa a as-built el 2026-09-25**: los recorridos paso a paso y sus correos **están implementados y
> verificados** —el recorrido entero de un ticket, con el escalado, la devolución a Soporte y los
> siete avisos, tiene su caso de interfaz—. Se cambia el estado, no el contenido.
>
> Aprobado por el responsable el 2026-09-22, tras repasarlo en forma de preguntas mientras se
> escribía. El repaso destapó un hueco que no era de este documento sino de los dos anteriores —la
> **devolución a Soporte** que `propósito-y-alcance.md` prometía y no tenía puerta en
> `modules/tickets.md`— y de ahí salieron la **regla 8**, el cambio de la **regla 5** y el **séptimo aviso**
> por correo.
>
> Con este documento, **la cadena de producto está completa**: `auth`, `users` y `tickets` pueden
> implementarse en ese orden. Queda `ambientes.md` para el despliegue y las pruebas.

## 1. Alcance de este documento

Cuenta **paso a paso** qué pasa desde que alguien pide algo hasta que el ticket se cierra, quién
hace cada cosa, qué se ve en la pantalla, qué correo sale y qué queda en el historial. Es el
documento que se lee para construir las pantallas y los correos.

Da por escritos los permisos (`docs/usuarios-y-permisos.md`), los estados, las transiciones y los
endpoints (`docs/modules/tickets.md`) y el producto (`docs/propósito-y-alcance.md`). **No los repite**: si
algo se contradice, mandan esos tres.

Está **aprobado**, así que habilita escribir código (Regla 0). La sección 9
recoge lo que he propuesto yo y conviene que confirmes.

## 2. Cómo leerlo

Cada flujo lleva lo mismo, en el mismo orden:

| | |
| --- | --- |
| **Quién** | Los papeles que intervienen |
| **Paso a paso** | Lo que ocurre, en orden |
| **Se ve** | Lo que cambia en la pantalla |
| **Correo** | Qué aviso sale y a quién |
| **Historial** | Qué queda registrado |

Una regla que vale para todos: **el correo es un aviso, no la acción**. Si un aviso no sale, lo que
el flujo cuenta ya ha pasado (docs/arquitectura.md, sección 9).

## 3. Alta de un ticket

**Quién**: el usuario, o Soporte en su nombre.

### 3.1 Lo crea el usuario

1. Entra y pulsa «Nuevo ticket».
2. Escribe **asunto** y **descripción**, **elige la categoría** —es obligatoria: sin ella no hay ticket
   (`docs/modules/tickets.md`, decisión 65)—, pone las **etiquetas** que quiera y adjunta lo que tenga
   (capturas, sobre todo). La categoría la puede corregir Soporte después; las etiquetas, también.
3. Guarda.

**Se ve**: el ticket aparece en «Mis tickets» del solicitante con su número (`ACME-2026-0042`), en
estado `nuevo`, y **ya asignado al técnico al que le tocaba el turno** —así que ese técnico lo tiene
también en **su** «Mis tickets», y los demás lo ven en **Tickets principales**, que es la lista de
todo lo que hay (2026-09-26).

**Correo**: **al técnico asignado**, «Ticket nuevo: ACME-2026-0042», con el asunto y el enlace. Si no
hay ningún técnico activo, el ticket se queda sin responsable y **no se avisa a nadie**: aparece en
**Tickets principales** para que alguien lo coja, y queda en el log.

**El reparto es por turnos** entre los técnicos activos, y vale también para los tickets que crea
Soporte en nombre de otro (`docs/modules/tickets.md`, sección 3.3). Se puede reasignar a mano
después.

**Historial**: `creado`, con el usuario como actor.

### 3.2 Lo crea Soporte en nombre de otro

Igual, con dos diferencias: Soporte **elige al solicitante** (la persona a la que está ayudando) y
el ticket queda con **creador** y **solicitante** distintos.

**Se ve**: el ticket aparece en **Tickets principales** —la lista de todo lo que hay—, en el «Mis
tickets» del técnico al que le tocó el turno y **también en «Mis tickets» del solicitante**, porque es
su ticket: lo verá y podrá responder cuando entre.

**Correo**: a Soporte, igual que en 3.1. **Al solicitante no se le avisa de que se ha creado**: no lo
creó él, y un correo diciendo «has creado un ticket» sería ruido.

## 4. Triaje y atención en Soporte

**Quién**: Soporte.

1. Soporte ve el ticket nuevo —en **Tickets principales** si no es suyo, o en **Mis tickets** si le
tocó el turno— y **lo lee**.
2. **Se lo asigna desde el propio ticket**, con el selector de responsables de su ficha (o lo comenta
y pregunta al usuario antes de nada).
3. Pasa el ticket a **`en progreso`** mientras lo trabaja.

De ahí salen tres caminos, y **Soporte elige**:

### 4.1 Se resuelve sin escalar

4. Soporte pasa el ticket a **`resuelto`** y escribe qué se hizo.
5. Soporte —o el usuario— lo **cierra**.

**Correo**: al usuario, «Tu ticket ACME-2026-0042 está resuelto» y después «…se ha cerrado».

### 4.2 Hace falta algo del usuario

4. Soporte pasa el ticket a **`en espera`** y pregunta en la conversación.
5. El usuario responde cuando puede.
6. Soporte **resuelve desde `en espera`** (no hace falta pasar por `en progreso`) o **escala**.

**Correo**: al usuario, el cambio a `en espera` («necesitamos algo de ti»). Es el único aviso que
existe para pedirle algo: no hay recordatorios ni plazos.

**El estado no cambia solo cuando el usuario escribe**: el ticket sigue en `en espera` hasta que
Soporte lo mueve. Un comentario no siempre es una respuesta —a veces es «gracias» o una pregunta
nueva—, y dar por respondido algo que no lo está es peor que moverlo a mano.

### 4.3 No se puede resolver: se escala

Ver la sección 5.

### 4.4 El ticket no debía existir

Si es un duplicado, una prueba o algo creado por error, Soporte —o el usuario, si es suyo— **lo
cierra sin resolverlo**. No hay estado `cancelado`: el historial dice que se cerró sin pasar por
`resuelto`, y eso es toda la diferencia que hace falta.

## 5. Escalado a Desarrollo

**Quién**: Soporte escala. Desarrollo recibe.

1. Soporte **escala** el ticket y **escribe el motivo**: qué se probó, qué se descartó y qué hace
   falta. **El motivo es obligatorio**: sin él, el escalado es «pásalo tú».
2. El sistema **crea el ticket interno** (`INT-ACME-2026-0042`) con ese motivo, en estado `nuevo`.
3. El principal pasa a **`escalado`**.

**Se ve**: el principal sigue en las listas de Soporte, ahora en `escalado`; el interno aparece en
**Tickets internos** —y en el «Mis tickets» de Desarrollo cuando tiene responsable—. Al ser la
primera escalación, el interno nace.

**Correo**: a Desarrollo, «Escalado: INT-ACME-2026-0042», con el motivo.

**Historial**: `escalado` en los dos tickets, con Soporte como actor.

**Cómo se lee el interno**: Desarrollo lee el principal (la conversación con el usuario incluida)
pero **no escribe en él**. Todo lo que tenga que decir va en el interno, y es Soporte quien decide
qué traslada al usuario.

## 6. Trabajo en Desarrollo

**Quién**: Desarrollo.

1. Ve el interno —en **Tickets internos** o en **Mis tickets** si ya es suyo—, **se lo asigna desde el
   propio ticket** y lo pasa a **`en progreso`**.
2. Si al mirarlo cree que **no es un cambio de código** —es algo que Soporte puede resolver—, lo
   **cierra sin resolver** con un comentario que lo explique. **El principal vuelve a manos de
   Soporte** (regla 8), y Soporte recibe el mismo aviso de «ha vuelto a tu bandeja».
3. Si necesita algo de Soporte (una prueba, un dato, una confirmación), pasa el interno a
   **`en espera`** y lo pide en la conversación.

**Se ve**: el principal pasa a **`en progreso`** mientras el interno está `en espera`: lo que se
espera es cosa de Desarrollo, y el usuario no tiene por qué ver el ticket como «escalado» para
siempre.

**Correo**: **a Soporte**, «Desarrollo necesita algo en INT-ACME-2026-0042». Es el único aviso que
avisa hacia dentro del nivel 2, y sin él el ticket puede quedarse parado esperando a alguien que no
sabe que le toca.

4. Cuando Soporte responde, el interno vuelve a `en progreso`.
5. Desarrollo **resuelve** el interno, con un comentario que explica qué se hizo. Ese comentario lo
   lee Soporte, no el usuario: vive en el ticket interno.

**Se ve**: el principal vuelve a **`en progreso`**, en la bandeja de Soporte. El caso no se cierra
desde Desarrollo: **vuelve al nivel 1**, que es quien habla con el usuario.

**Correo**: **a Soporte**, «El ticket ACME-2026-0042 ha vuelto a tu bandeja». Sin este aviso, un caso
que Desarrollo ya terminó puede quedarse sin explicar al usuario.

6. **Soporte lee el interno**, escribe en el principal qué se hizo —en palabras del usuario, no notas
   técnicas— y **resuelve** el principal.

**Correo**: al usuario, «Tu ticket ACME-2026-0042 está resuelto». Y desde ahí, lo cierra Soporte o el
propio usuario.

**Por qué el último paso es de Soporte**: Desarrollo no tiene contacto con el usuario, así que la
explicación que él lee no puede salir del ticket interno. Es la regla 5, y es la razón de que
Desarrollo no resuelva el principal.

## 7. Cierre y reapertura

### 7.1 Cierre

**Quién**: Soporte o el solicitante.

1. El principal está `resuelto`.
2. Soporte —o el usuario— **lo cierra**.

**Correo**: al usuario, «Tu ticket se ha cerrado».

**El interno no se cierra solo.** Si sigue abierto en `resuelto`, Soporte o Desarrollo **lo cierran a
mano**. Es una tarea de orden, no un efecto del principal, y por eso puede quedar pendiente sin que
nada se rompa.

### 7.2 Reapertura

**Quién**: Soporte o el solicitante, sobre un ticket `cerrado`.

1. Se reabre: el principal vuelve a **`en progreso`**.
2. Se limpian `resolved_at` y `closed_at`: la pantalla no puede decir «resuelto el…» con una fecha
   de la vez anterior. El historial conserva cuándo pasó cada cosa.

**Para volver a escalar hay que reabrir primero**: un ticket cerrado no se escala.

## 8. Re-escalado

**Quién**: Soporte, sobre un principal en `resuelto` (o ya reabierto).

1. Soporte **vuelve a escalar**, con su motivo otra vez.
2. El **mismo ticket interno** —no uno nuevo— vuelve a **`en progreso`**, aunque estuviera `resuelto`
   o `cerrado`.
3. El principal pasa a `escalado`.

**Correo**: a Desarrollo, el aviso de escalado **otra vez**: el trabajo ha vuelto a aparecer, y eso
es lo que el aviso significa.

## 9. Lo que se decidió al repasar este documento

Al escribirlo salió un hueco que no era de este documento, sino de los dos anteriores: las reglas de
sincronización **no tenían puerta para la «devolución a Soporte»** que `docs/propósito-y-alcance.md`
promete —el caso en que Desarrollo mira el interno y concluye que no es un cambio de código—, y
tampoco decían quién explica al usuario qué se hizo. Las dos cosas obligaron a enmendar
`docs/propósito-y-alcance.md` y `docs/modules/tickets.md`.

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **La devolución a Soporte** | Cerrar el interno sin resolverlo **devuelve el principal a `en progreso`**. Es la **regla 8**, nueva |
| 2 | **Quién resuelve el principal** | **Soporte, siempre**, y con su explicación para el usuario. La **regla 5** cambió: el interno resuelto devuelve el principal a `en progreso`, no a `resuelto` |
| 3 | **Comentar en un ticket cerrado** | No se puede: para seguir hablando hay que reabrirlo |
| 4 | **Plazo para reabrir** | Ninguno: se puede reabrir un ticket cerrado haga el tiempo que haga |
| 5 | **El ticket creado en nombre de otro** | Lo ve el solicitante en «Mis tickets» y no recibe aviso de su creación: no lo creó él |
| 6 | **El formato de los correos** | Texto plano, con el número, el asunto y el enlace |
| 7 | **El aviso a Soporte cuando el ticket vuelve** | **Séptimo aviso**: cubre los dos caminos por los que el principal vuelve a Soporte (Desarrollo resolvió el interno, o lo devolvió sin resolver) |
| 8 | **Asignarse el interno** | Desarrollo puede asignarse los internos, y Soporte puede asignárselos a alguien |
| 9 | **Dónde se asignan los tickets** | **Dentro del ticket**, en su ficha: en el principal reasigna Soporte a cualquier técnico, y en el interno Soporte y Desarrollo a cualquier desarrollador. **La bandeja no asigna** (decisión del responsable, 2026-09-26) |

## 10. Lo que NO se automatiza

Nada de lo que sigue existe, y está escrito para que no se dé por hecho:

- **Recordatorios por plazo**: si un ticket lleva días `en espera`, nadie lo recuerda por correo.
- **Escalado automático por tiempo**: un ticket no sube a Desarrollo porque se haya tardado.
- **Cierre automático**: un ticket `resuelto` no se cierra solo pasados unos días.
- **Aviso a Desarrollo por cada comentario de Soporte**: sólo por el escalado.
- **Reapertura automática** cuando el usuario responde: si el ticket está cerrado, hay que reabrirlo.

## 11. Flujos que no están aquí

- **Cuentas**: alta, primer acceso, olvido y reseteo de contraseña, y la convivencia de los tres
  métodos de entrada. Están en `docs/usuarios-y-permisos.md`, secciones 4 a 9.
- **Despliegue y pruebas**: van en `docs/ambientes.md`.

## 12. Qué habilita este documento

Con este documento aprobado, **la cadena de producto está completa**: `auth`, `users` y `tickets`
pueden implementarse en ese orden, y las pantallas y los correos tienen su recorrido escrito. Queda
`ambientes.md` para el despliegue y las pruebas, que es el último.
