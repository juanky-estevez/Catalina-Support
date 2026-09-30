# Tickets: modelo de datos, transiciones y módulos

> **Estado:** as-built
> **Última actualización:** 2026-09-27
>
> **Enmendado el 2026-09-27 (tercera vez)**, a petición del responsable, al usarlo: **las etiquetas
> pasan a poder mantenerse** —se crean, se **renombran** (y el cambio vale para **todos** los tickets
> que la llevan) y se **retiran** (y se quitan de esos tickets, preguntando antes)—, porque en la
> pantalla no había forma de crearlas ni de corregirlas y **las categorías se llevaban todo el
> protagonismo**. Siguen **naciendo solas** al escribirlas en un ticket: no hay que darlas de alta para
> usarlas. Y las **dos mitades de la pantalla tienen el mismo peso** (decisión 72).
>
> **Enmendado el 2026-09-27 (segunda vez)**, a petición del responsable, al usarlo: la lista lleva **un
> solo botón de búsqueda** que abre un modal con todos los criterios, con un resumen de lo filtrado
> (sección 5 y `docs/interfaz-y-experiencia.md`, sección 3.3) —las tres filas de filtros se comían la
> pantalla y **la de categorías se salía**—; y la **vista del ticket al abrirlo** pasa a ser la del
> papel de quien mira, con «Los dos» enseñando **sólo las dos conversaciones** (sección 5).
>
> **Enmendado el 2026-09-27**, al pedir el responsable **categorías y etiquetas** para agrupar y
> buscar: nace **`ticket_categories`** —«Red», «Software x», «Licencias»…— con la que **todo ticket
> nace clasificado** (nunca hay uno sin categoría, y lo garantiza la propia base) y que se puede
> corregir después, y **`ticket_tag_names`**

**El catálogo de etiquetas** (decisión 72): el nombre de cada etiqueta, para poder crearla sin que la
lleve ningún ticket, renombrarla de una sola vez y contarla.

| Columna | Para qué |
| --- | --- |
| `id` | Clave |
| `tag` | El nombre, ya normalizado: minúsculas, con guiones, sin acentos. **La forma normalizada es la que se guarda**, a diferencia del nombre de una categoría —que es un nombre que se lee—: una etiqueta es una clave con la que se agrupa |
| `normalized` | El mismo nombre normalizado, **único**: es lo que impide `red-wifi` y `Red-Wifi` a la vez |
| `created_by_id` | Quién la creó, **nulo** si nació de una base puesta al día o de una etiqueta de antes |
| `created_at` | Cuándo |

**`ticket_tags`**

**Qué ticket lleva qué etiqueta**: la tabla puente entre el ticket y el catálogo.

| Columna | Para qué |
| --- | --- |
| `id` | Clave |
| `ticket_id` | El ticket. Clave ajena con **borrado en cascada**, como los comentarios |
| `tag_id` | La etiqueta, **clave ajena al catálogo con borrado en cascada**: retirar una etiqueta es borrar su fila del catálogo y sus enlaces se van con ella, en un solo movimiento |
| `created_by_id`, `created_at` | Quién la puso en ese ticket y cuándo | —opcionales, **en minúsculas y con guion medio** (`red-wifi`),
> con sugerencias de las que ya existen—. **La categoría y las etiquetas son del principal** y el
> interno las hereda al leerse. Las dos sirven para **filtrar** (chips) y para **buscar** (entran en la
> misma caja de texto). El catálogo lo mantienen **Soporte y el Administrador**, y **retirar** una
> categoría es cosa **sólo del Administrador**. Decisiones 64 a 68.
>
> **Enmendado el 2026-09-27**, al pedir el responsable **poder etiquetar a otros técnicos y
> desarrolladores** para que sean partícipes del ticket: nace el concepto de **observador**, que **no es
> lo mismo que el asignado** —el ticket tiene **como mucho un asignado** (que sigue siendo opcional,
> como estaba) **y muchos observadores**—. Se etiqueta **escribiendo en el comentario** (`@` o el botón
> del editor), el nombre queda **resaltado como una mención**, y quien la lee ve a quién se llama. Los
> observadores **salen en su «Mis tickets» con un chip propio** —Asignados · Observo · Todos—, se ven
> en la ficha como **«Observadores»** y **reciben los mismos avisos que el responsable**. Etiquetar es de
> **Soporte y Desarrollo**. Decisiones 58 a 63, y una **plantilla de correo nueva** (la undécima).
>
> **Y el responsable añadió, en la misma tanda**, que **se pueda quitar a un observador** y que **lo
> pueda hacer cualquier técnico o desarrollador**: eso obliga a **guardar los observadores en su propia
> tabla** —quien quita no es el autor del comentario y no puede reescribirlo— y deja una regla que hay
> que decir en voz alta: **quitar manda sobre lo ya escrito**, así que la mención se queda en el
> comentario pero **no vuelve a añadirlo**; vuelve sólo si alguien lo menciona **de nuevo** (decisión
> 63). Quitar a alguien **queda en el historial** del ticket.
>
> **Enmendado el 2026-09-27**, al pedir el responsable **dos columnas nuevas en la lista** que redacta
> el módulo `ai` —**«Motivo»** y **«Última acción»**, en español y en inglés—, y **dos nombres que
> cambian**: «Número» pasa a ser **«Ticket»** y «Última actualización», **«Fecha actualización»**
> (decisiones 56 y 57). Este módulo **no cambia de esquema**: es `ai` quien guarda sus textos en su
> propia tabla, y aquí sólo se **piden** al mover el ticket y se **leen en bloque** para pintar la
> lista. El único endpoint nuevo es el botón de regenerar, y es de este módulo.
>
> **Enmendado el 2026-09-26**, al pedir el responsable **dos cosas de los adjuntos**: que se puedan
> mandar **archivos de texto y código** —un `.sql` no entraba, y es lo que se manda cuando el caso es
> una consulta que falla o un despliegue que no arranca: se añaden **cuarenta y dos extensiones**
> (decisión 54)—, y que **la imagen y el vídeo no ocupen todo el ancho** de la conversación, sino un
> **tope de 480 × 360** que al pulsarlo se abre **en el visor grande** para verlo entero y descargarlo,
> con el vídeo como **miniatura con su botón de reproducir** (decisión 55).
>
> **Enmendado el 2026-09-26**, al pedir el responsable que **la bandeja sea «lo mío»** y que se pueda
> ver **todo** desde su propio sitio: `GET /api/tickets` estrena el filtro **`mine`** —lo asignado a
> mí, lo que abrí yo y aquello donde he comentado (decisión 51)—, **la Bandeja pasa a ser Mis
> tickets** con chip de tipo para Soporte y Desarrollo, y el módulo `tickets` gana **dos listas**:
> **Tickets principales** y **Tickets internos**, para los dos equipos (decisión 52). **«Asignármelo»
> sale de la bandeja** —la reasignación se hace dentro del ticket, que es donde se ve a quién se le
> pasa, y con las reglas de siempre— y las filas dicen **«Abrir»** (decisión 53). **Sin cambios en el
> modelo de datos, en las transiciones ni en los permisos**: lo único nuevo del backend es el filtro.
>
> **Enmendado el 2026-09-26**, al pedir el responsable que **los adjuntos se vean dentro del texto**,
> en el sitio donde se escriben y en el orden en que se cargan: el texto de un ticket y el de un
> comentario **dejan de ser texto plano y pasan a ser HTML con formato**, y **dentro va todo lo que se
> adjunta** —las imágenes y los vídeos como elementos que se ven, el PDF y los demás archivos como
> enlaces—. **Se ve igual al leer que al escribir**: la vista previa está también mientras se edita,
> que es lo que se pidió. Se añaden **cuatro extensiones de vídeo** (`mp4`, `webm`, `mov`, `avi`) y
> **cinco decisiones** (45 a 49). **Los correos no se tocan**: ninguno lleva el texto del ticket ni de
> los comentarios.
>
> **Es la decisión más cara de este documento**, y está tomada a sabiendas: un texto con formato abre
> la puerta a HTML escrito a mano, y por eso entra con **lista blanca de etiquetas**, con **ninguna
> dirección guardada en el texto** —los adjuntos se referencian por su nombre y la dirección se
> resuelve al pintar— y con la búsqueda de la bandeja **quitando las etiquetas antes de buscar**.
>
> **Pasa a as-built el 2026-09-26**: lo que describe este documento es lo que hace el código, backend
> y pantallas.
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
>
> **Enmendado el 2026-09-25**, al escribir el script de copias, con una **decisión del responsable**:
> los adjuntos dejan de vivir sueltos en la raíz de `_files/` y pasan a la **carpeta de su ticket**
> (`_files/2026/CS-2026-0042/`), con el interno en la suya. La sección 2.3 cuenta cómo queda, y el
> nombre del archivo sigue siendo generado: lo que cambia es dónde se guarda, no cómo se llama.
>
> **Enmendado el 2026-09-24**, al empezar a implementarlo, con el repaso en forma de preguntas previo
> al código. Cinco cosas:
>
> 1. La sección 0 decía que a la matriz de `docs/usuarios-y-permisos.md` le faltaba la fila de
>    **cerrar el ticket interno**. **Ya estaba** —la añadió la enmienda del 2026-09-22, como cuenta
>    la cabecera de este mismo documento—. Es una errata y queda corregida; no cambia nada.
> 2. La tabla de endpoints de la sección 5 llevaba `GET` y `PUT /api/settings`, que **no son de este
>    módulo**: viven en `docs/modules/settings.md`, que además ya los lista. Se quitan de aquí.
> 3. La sección 3.3 no decía **entre quién** se reparte un ticket interno cuando el Administrador
>    activa su reparto. Se fija: entre los usuarios activos con papel `desarrollo`, con el mismo
>    criterio del turno. **Es una corrección del responsable**: los internos nacen **sin repartir** y
>    el reparto es **configurable por el Administrador**, que es el valor que ya trae la base.
> 4. La matriz estrena la fila de **reabrir el ticket interno** (Soporte): la transición existía en
>    la sección 3.2 y en la matriz no se veía.
> 5. La bandeja del Administrador lleva **chip de tipo** (Principales · Internos · Todos), porque ve
>    los dos y necesita poder separarlos (`docs/interfaz-y-experiencia.md`, sección 3.3).
>
> Y otras cuatro que salieron **al escribir el servicio** y que el modelo no tenía resueltas: el motivo
> de un **re-escalado** va como comentario en el interno (30), un ticket **cerrado no se edita** (31),
> los dos avisos de dentro van a una persona y no a todo el equipo (32), y el responsable de un interno
> sale de Desarrollo (33). Todas son del responsable, del 2026-09-24, y están en el cuarto repaso.
>
> **Enmendado el 2026-09-24**, al construir las pantallas: la sección 8 deja de listar
> «la pantalla del prefijo de numeración» como pantalla de este módulo —se configura en
> **Configuración**, con el resto de la instalación, y sus endpoints son de `settings`—, y el documento
> recoge las cuatro decisiones que salieron al hacerlas (el repaso quinto): **«Nuevo ticket» para
> Soporte** —su menú estrena la entrada, porque puede crear en nombre de otra persona y ese permiso no
> tenía puerta; corrección del responsable—, el escalado **en un diálogo**, los adjuntos **con vista
> previa**, y que **el detalle es una sola pantalla para los cuatro papeles**.
>
> **Estado de la implementación (2026-09-24): el módulo está entero**, backend y pantallas. El backend
> —las ocho tablas, la numeración, el reparto, los endpoints y los ocho correos— está probado de
> extremo a extremo; las pantallas —la bandeja, el alta y el detalle con la vista doble— están hechas y
> probadas en un navegador, en PC y en móvil. Las siete tablas en
> `v1.0.0.sql`, la numeración sin carreras, el reparto por turnos, los trece endpoints, los siete
> avisos por correo y **78 comprobaciones de extremo a extremo** contra el entorno de desarrollo,
> leyendo los correos del buzón de pruebas: el alta con su reparto, el triaje, el escalado, las dos
> reglas de sincronización, la devolución, el re-escalado, el cierre, la reapertura, los comentarios
> (editar, borrar), los adjuntos (subir, descargar, la extensión que no vale, el svg, los 25 MB) y los
> permisos de los cuatro papeles. **Faltan las pantallas.**

## 0. El cierre del interno, y un hueco en un documento aprobado

Las ocho reglas de sincronización de `docs/propósito-y-alcance.md` **no dicen quién cierra el
ticket interno**. Se decidió que **se cierra a mano**, y no como consecuencia de cerrar el
principal:

- Lo cierran **Desarrollo o Soporte**, cuando lo dan por terminado.
- **Cerrar el interno no cambia el estado del principal.**
- **Re-escalar el principal lo reabre**, aunque estuviera cerrado (regla 3 de
  `docs/propósito-y-alcance.md`).

Eso **no contradice** el documento aprobado, así que no hay que enmendarlo. Y el permiso que faltaba
—**cerrar el ticket interno**— **ya está en la matriz** desde la enmienda del 2026-09-22
(`docs/usuarios-y-permisos.md`, sección 3): esta sección decía lo contrario y queda corregida
(2026-09-24). Al implementarlo se añadió además la fila de **reabrir el interno**, que es la otra
mitad de lo mismo.

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
| `assignee_id` | `bigint` | Responsable. **Se pone al repartir el ticket**, y queda nulo sólo si no había ningún técnico activo |
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

**El prefijo lo cambia un Administrador desde la aplicación**, no es configuración del servidor, y
vive con el resto de la configuración de la instalación: **el módulo `settings`**
(`docs/modules/settings.md`). `tickets` lo lee de ahí por su servicio cuando numera un ticket, igual
que lee la configuración del reparto.

Es una tabla de **una sola fila** y no una de «ajustes» con clave y valor: una tabla así acaba siendo
el cajón donde entra todo. Con una fila se ve de un vistazo **qué** es configurable, y añadir otro
ajuste exige tocar el esquema, que es justo la conversación que conviene tener.

> **Enmendado el 2026-09-24, al escribir `docs/modules/settings.md`** (aprobado por el responsable el
> mismo día): esta tabla se llamaba `ticket_number_settings` y **sólo recogía el prefijo**, cuando el
> reparto y los avisos de la sección 3.3 no cabían en ninguna parte. Pasa a llamarse
> **`ticket_settings`** y recoge las cinco cosas: el prefijo, la asignación y el aviso de los tickets
> principales y los de los internos.

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
| `body` | `text` | **HTML con formato, y puede ir vacío cuando el comentario trae adjuntos.** Lleva dentro los adjuntos que se colocaron (enmienda del 2026-09-26) |
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
- **Un comentario puede ser sólo un adjunto** (decisión 41, del responsable): «mira esto» con la
  captura y sin escribir nada es algo que pasa todos los días, y obligar a escribir una frase para
  mandar una captura es ruido. **Es lo que obliga a que `body` deje de ser obligatorio**, y el motivo
  está dicho entero en la decisión: cuando el comentario se crea, sus archivos **todavía no han
  subido** —van después, con su `commentId`—, así que el backend no puede saber si van a llegar.
  **La regla que sí se sostiene se sostiene en la pantalla**: no se envía un comentario que no tenga
  **ni texto ni archivos**, y eso se dice antes de enviarlo.
- **Y al revés**: editar un comentario **sí** sigue exigiendo texto (decisión 42). Dejar un comentario
  en blanco por la puerta de atrás, después de haberse escrito, no es lo mismo que nacer sin texto: lo
  que se quiere en ese caso es borrarlo, y para eso está el borrado.
- **Un comentario sin texto se lee por sus adjuntos**: en la conversación aparece con su autor, su
  fecha y sus archivos, y **sin un hueco en blanco** que parezca un fallo de la pantalla.

**`ticket_attachments`**

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | `bigserial` | |
| `ticket_id` | `bigint` | El principal, o nulo |
| `internal_ticket_id` | `bigint` | El interno, o nulo. Nunca los dos |
| `comment_id` | `bigint` | Nulo si el adjunto va en la descripción inicial |
| `uploaded_by_id` | `bigint` | |
| `filename` | `text` | Nombre original, tal como lo subió la persona |
| `stored_name` | `text` | **Ruta en disco**, relativa a la carpeta de archivos y generada: `2026/CS-2026-0042/ab12….png` |
| `content_type` | `text` | |
| `size_bytes` | `bigint` | |
| `created_at` | `timestamptz` | |

- Los archivos **no se guardan en la base de datos**, sino en disco, en la carpeta que indique
  `FILES_PATH`. En los contenedores es un volumen montado (`_files`), igual que `_logs`: los adjuntos
  sobreviven a los reinicios y se pueden copiar desde la máquina. `dev.yml` y `prod.yml` lo montan y
  `config/env/*.env` define la variable.
- **Cada ticket tiene su carpeta**: `_files/<año>/<número>/`, con el año del ticket y su número tal y
  como se enseña (`_files/2026/CS-2026-0042/ab12….png`). **El ticket interno tiene la suya**, porque su
  número es suyo: `_files/2026/INT-CS-2026-0042/`. Así, mirando el disco, se sabe de qué ticket es cada
  archivo y qué se subió en la conversación interna, sin abrir la aplicación (decisión del responsable,
  2026-09-25: antes eran archivos sueltos en la raíz de `_files/`, y no había forma de saber de quién
  era cada uno).
- **El nombre del archivo sigue siendo generado** (32 caracteres y su extensión): el original puede
  repetirse, traer acentos o intentar salirse de la carpeta. **El original se queda en la base** para
  enseñarlo, y el número de la carpeta también se limpia antes de usarse: nada que venga de fuera entra
  en una ruta sin pasar por ahí.
- **Los adjuntos subidos antes de las carpetas** viven sueltos en `_files/` y **se siguen viendo
  igual**: `stored_name` es una ruta relativa, y un nombre suelto resuelve a la raíz de la carpeta de
  archivos. **No se movió ninguno**: en desarrollo se borraron y se volvió a empezar, como pidió el
  responsable, y en producción no hay nada todavía.
- Si al borrar un archivo se queda vacía la carpeta de su ticket, se quita también. Sólo pasa en los
  caminos de fallo —un archivo que se subió de más, una fila que no se pudo guardar—, porque **borrar
  adjuntos ya subidos no entra en la 1.0.0** (sección 7).
- **Límite de subida en nginx**: los vhosts llevan `client_max_body_size 30m;`. Sin él, nginx usa
  **1 MB** por defecto y cualquier adjunto de más de un mega se rechazaría con un 413 antes de llegar
  al backend. El 30 deja margen sobre los 25 MB del archivo.
- **Se descargan por la API, nunca por una ruta estática**: el adjunto de un ticket interno no puede
  ser alcanzable por el solicitante, y eso sólo se garantiza si cada descarga pasa por la
  comprobación de permisos.
- Límites: **25 MB por archivo** y **sin límite de cantidad**. Lo que sí es una lista cerrada es
  la **extensión**: `pdf`; imágenes `png`, `jpg`, `jpeg`, `gif`, `webp`; **vídeo `mp4`, `webm`, `mov`,
  `avi`** (añadidos el 2026-09-26, decisión 47); Office `doc`, `docx`, `xls`, `xlsx`, `ppt`, `pptx`,
  `odt`, `ods`, `odp`; **texto y código** `txt`, `csv`, `log`, `md`, `css`, **`sql`, `json`, `xml`,
  `yml`, `yaml`, `ini`, `conf`, `cnf`, `sh`, `bash`, `bat`, `ps1`, `py`, `js`, `ts`, `java`, `php`,
  `go`, `cs`, `rb`, `pl`, `kt`, `rs`, `swift`, `c`, `h`, `cpp`, `hpp`, `vue`, `jsx`, `tsx`,
  `htaccess`, `env`, `properties`, `diff`, `patch`, `bak`, `old`** (añadidos el 2026-09-26,
  decisión 54); y comprimidos y copias `zip`, `rar`, `tar`, `gz`, `tgz`, `7z`. Lo que no esté en la
  lista se rechaza: no se acepta «cualquier cosa» y luego se mira.
- **El texto y el código se guardan con su tipo y se descargan siempre**: no se pintan nunca en línea,
  así que da igual lo que lleven dentro. Es lo mismo que ya valía para el Office y el `zip`, y es lo
  que permite admitir un `.sh` o un `.sql` sin abrir un agujero: **lo único que se enseña dentro del
  ticket es la imagen, el vídeo y el PDF** (`sePuedePrevisualizar`), y todo lo demás baja como
  `application/octet-stream` con `attachment`, que es lo que impide que un archivo se abra solo.
- **El texto también puede nombrar a una persona, y lo hace por su identificador**: una mención es
  `<span data-mencion="12">María Pérez</span>`, y **nunca por el nombre escrito a mano**. Es la misma
  idea que el `data-adjunto`: el nombre que se enseña es un adorno y el identificador es el dato, así
  que si alguien se cambia los apellidos —o hay dos personas que se llaman igual— la mención sigue
  apuntando a quien apuntaba (decisión 58).
- **Al usuario no se le rechaza un texto que no ha escrito él** (2026-09-27): si edita la descripción
  de un ticket que abrió Soporte en su nombre y esa descripción **ya llevaba una mención**, el cuerpo se
  guarda —no ha etiquetado a nadie, sólo ha conservado lo que había—. Lo que se rechaza es que
  **aparezca una mención nueva** en un cuerpo suyo.
- **Sólo se puede etiquetar a técnicos y desarrolladores activos** y **sólo Soporte y Desarrollo pueden
  hacerlo**: al guardar el comentario, el módulo **comprueba una a una** las menciones contra las
  cuentas activas de esos dos papeles, y si alguna no vale **rechaza el comentario entero** con
  `tickets.mention.notAllowed` —igual que el saneador rechaza el HTML que no admite: un texto que se
  guarda a medias es peor que uno que se rechaza—. El usuario no etiqueta: no ve la lista de técnicos ni
  tiene por qué (decisión 59).
- **`svg` queda fuera a propósito**, aunque sea una imagen: un SVG puede llevar código dentro y se
  abre en el navegador. Los demás formatos de imagen no.
- **Sin antivirus** en la 1.0.0: queda dicho para que nadie lo dé por hecho.

**`ticket_categories`**

El catálogo de categorías de la instalación. **Es de `tickets`** y no de `settings`: son datos de los
tickets, y su pantalla es una pantalla de este módulo (decisión 64).

| Columna | Para qué |
| --- | --- |
| `id` | Clave |
| `name` | El nombre, tal y como se lee: «Red», «Software x» |
| `normalized` | El nombre en minúsculas y sin acentos, **único**: es lo que impide que el catálogo acabe con «Red» y «red» |
| `active` | Si se ofrece al crear un ticket. **Retirar es desactivar** (decisión 66) |
| `created_by_id`, `created_at` | Quién la creó y cuándo |
| `deactivated_at` | Cuándo se retiró, si se retiró |

**`tickets.category_id`**: columna nueva, **`NOT NULL`** con clave ajena a `ticket_categories`
(decisión 65). La trae la migración con la categoría **«General»** ya puesta a lo que exista.

**`ticket_tags`**

Las etiquetas de un ticket: opcionales, muchas por ticket, y **sólo en el principal** (decisión 67).

| Columna | Para qué |
| --- | --- |
| `id` | Clave |
| `ticket_id` | El ticket al que pertenece. **Clave ajena con borrado en cascada**, como los comentarios |
| `tag` | La etiqueta, ya normalizada: minúsculas, con guiones, sin acentos |
| `created_by_id`, `created_at` | Quién la puso y cuándo |

- **`UNIQUE (ticket_id, tag)`**: la misma etiqueta no se repite en el mismo ticket.
- **Índice por `tag`**: es como se filtran los tickets de una etiqueta, y como se cuentan para
  sugerirlas.

**`ticket_observers`**

Los observadores de un ticket: **quién sigue el ticket sin atenderlo**. Va aparte de la tabla de los
tickets porque son **muchos por ticket** y porque la lista tiene que poder cambiar sin tocar el texto
de nadie.

| Columna | Para qué |
| --- | --- |
| `id` | Clave |
| `ticket_id` / `internal_ticket_id` | **Uno de los dos**, y sólo uno: cada hilo tiene sus observadores, igual que tiene sus comentarios |
| `account_id` | Quién observa. **Clave ajena a `users`** |
| `added_by_id` | Quién lo añadió —el autor de la mención—, o quién lo puso si algún día se añade a mano |
| `created_at` | Cuándo |

- **`UNIQUE (ticket_id, account_id)` y `UNIQUE (internal_ticket_id, account_id)`**: mencionar dos veces a
  la misma persona no la pone dos veces. Con `NULL` de por medio —una columna nula siempre es distinta de
  otra— hacen falta los dos índices únicos parciales, uno por cada lado, como en los comentarios.
- **Se borra de verdad**: no hay marca de borrado. Quitar a un observador es quitarlo, y lo que queda
  —que se le quitó, quién y cuándo— está en el historial.

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

#### El texto de un ticket y de un comentario es HTML, y los adjuntos van dentro

**Lo que alguien escribe en la descripción de un ticket o en un comentario es un texto con formato**, y
**los archivos que adjunta viven dentro de ese texto**, en el sitio donde los puso, en el orden en que
los cargó (decisión 45, del responsable, 2026-09-26).

```text
Hola mi PC no funciona, me aparece estos errores:

[captura1.png]          <- esto es una imagen, y se ve la imagen

[captura2.png]          <- y esta otra

Solicito su ayuda por favor
Gracias
```

**Lo que se ve en cada sitio** (decisión 46):

| Qué se adjunta | Dentro del texto | En la lista del final |
| --- | --- | --- |
| **Imagen** (`png`, `jpg`, `jpeg`, `gif`, `webp`) | **La imagen en línea**, con el ancho de la conversación | Igual que las de dentro, con sus botones |
| **Vídeo** (`mp4`, `webm`, `mov`, `avi`) | **Un reproductor en línea**: el texto guarda `<video data-adjunto="grabacion.mp4"></video>` y **los controles los pone la pantalla al pintarlo** —el texto no los lleva, porque un atributo de más se rechaza— | |
| **PDF** | Un **enlace** con el nombre del archivo: al pulsarlo se abre **el visor** | |
| **Todo lo demás** (Word, Excel, texto, comprimidos…) | Un **enlace**: al pulsarlo **se descarga** | |

**Y se ve igual al escribir que al leer** (lo pidió así el responsable): mientras se redacta, cada
imagen y cada vídeo **se ven de verdad** en su sitio, no como un texto que los representa. Un cuadro de
texto normal no puede llevar una imagen dentro, así que **el cuadro de escribir es un editor con
formato** (`docs/interfaz-y-experiencia.md`, sección 6.5), con su barra de botones —negrita, cursiva,
lista, enlace— **y el botón de adjuntar**, que mete el archivo donde está el cursor.

**Las cuatro reglas que hacen que esto no se convierta en un agujero**:

1. **Lista blanca de etiquetas.** El backend **no guarda cualquier HTML**: se aceptan párrafos, saltos,
   negrita, cursiva, subrayado, tachado, listas, enlaces, imágenes y vídeos, y nada más. Un `<script>`,
   un `onerror`, un `<iframe>` o un `<style>` **se rechazan** —la clave es `tickets.body.notAllowed`—,
   y lo que llega fuera de la lista no se limpia en silencio: se dice, porque un texto que se guarda a
   medias es peor que uno que se rechaza. **La referencia de un adjunto es obligatoria**: un `<img>`,
   un `<video>` o un `<a>` sin ella se rechazan, porque un hueco no es nada. **Los comentarios de HTML
   sí pasan**, que no son etiquetas y no se ejecutan: el navegador mete un `<!--StartFragment-->` al
   pegar texto con formato, y rechazarlo rompería pegar.
2. **Ninguna dirección se guarda dentro del texto.** Un adjunto se referencia **por su nombre**
   —`<img data-adjunto="captura1.png">`— y **la dirección se resuelve al pintar**, con la lista de
   adjuntos del ticket que ya se ha leído. Así **no hay forma de meter una imagen de fuera** (un
   contador que avisa a un tercero de que alguien abrió el ticket), ni un `data:` gigante, ni una
   dirección que caduque. Un `src` que no sea el de un adjunto de **ese** ticket se rechaza.
3. **El texto no se guarda con las direcciones provisionales.** Mientras se escribe, el editor enseña
   el archivo desde el navegador (`blob:`); eso **no llega a la base**: lo que se guarda es la
   referencia por nombre, y el `blob:` se cae al guardar.
4. **La búsqueda de la bandeja quita las etiquetas antes de buscar.** La bandeja busca por número,
   asunto y texto (`ILIKE`), y buscar «p» o «img» sobre el HTML crudo encontraría **todos** los tickets
   con una imagen. La consulta **quita las etiquetas** antes de comparar. Lo que queda fuera, y se dice:
   un `&amp;` escrito a mano no se encuentra buscando «&», y **la búsqueda no mira dentro de los
   comentarios**, sólo el asunto y la descripción, como hasta ahora.

**Los adjuntos que el texto no nombra se siguen enseñando al final**, en su lista, como hasta ahora
(decisión 49). Es lo que hace que **los tickets y comentarios que ya existen se vean igual** y que no
se pierda nada de lo que se colgó antes de esta decisión. También es donde caen los que se suben sin
poder colocarse.

**Y un adjunto de dentro del texto no se puede borrar solo él**: quitarlo del texto **no borra el
archivo** —sigue en la lista del final, porque un adjunto puede ser la prueba de algo (sección 2.3)—.
Para que desaparezca haría falta poder borrar adjuntos, que sigue **fuera de la 1.0.0** (sección 7).

**Y hay un sitio donde el texto se escribe solo**: el **motivo de un re-escalado**, que la decisión 30
manda como comentario del interno. Ese motivo es texto plano —el diálogo de escalar no tiene editor—,
así que **se escapa al escribirlo** (`&`, `<`, `>`), o un motivo con un `<` se leería como una
etiqueta. El `escalation_reason` del interno sigue siendo texto plano y se queda como está.

**El orden de la carga**, que es lo que hay detrás de todo esto:

1. Al adjuntar —pegando, arrastrando o con el botón— **se sube el archivo** y el editor lo enseña
   mientras se sigue escribiendo.
2. Al enviar, **el texto se guarda con la referencia por nombre** ya en su sitio.
3. Si un archivo **no sube**, la regla es la de la decisión 44: el comentario se queda publicado, se
   dice qué archivo falta y el botón lo sube **al mismo comentario**, sin volver a escribirlo. Lo que
   cambia es que ahora el hueco del archivo que falta **se ve también en el texto**, como un recuadro
   que dice que ese archivo no subió.

**El alta de un ticket no necesita nada especial**, y esa es la ventaja de referenciar por nombre: el
ticket se crea **con las referencias ya puestas** y sus archivos se suben después, colgando del
ticket; al pintarlo, cada referencia se resuelve con el adjunto que tiene ese nombre. La persona ve un
solo botón de crear, y la aplicación no vuelve a guardar el texto por el camino.

### 2.3.1 Los observadores

**Un ticket tiene como mucho un asignado y muchos observadores**, y no son lo mismo: el asignado **es
quien lo atiende** —y sigue siendo **opcional**, como estaba: un ticket nace sin responsable cuando no
hay ningún técnico activo al que repartirlo—, y los observadores **siguen el ticket sin atenderlo**
(decisión 60, corrección del responsable).

| | Asignado | Observador |
| --- | --- | --- |
| Cuántos | **Como mucho uno** | **Los que hagan falta** |
| Quién | Un técnico (principal) o un desarrollador (interno), por turno o a mano | Cualquier técnico o desarrollador activo, **etiquetado** |
| Qué hace | Atiende: comenta, mueve el estado, cierra | Mira: **lo mismo que cualquier otro que ve el ticket**, no gana permisos |
| De dónde sale | Del reparto, o de la ficha | **Del texto** —de las menciones de los comentarios de ese hilo— **o a mano, desde la ficha del ticket** (decisión 74) |
| Correo | El del reparto y los avisos del ticket | **Los mismos avisos de personal que el asignado**, y el de que le etiquetaron |
| Dónde lo ve | En «Asignados a mí» | En **«Observo»**, dentro de «Mis tickets» |

- **Ser observador no da permisos**: un observador ve el ticket porque ya podía verlo —los dos equipos
  ven los principales y los internos—, y sigue sin poder comentar donde no le toca (Desarrollo no
  escribe en el principal, `docs/usuarios-y-permisos.md`, regla 3).
- **Los observadores de un interno son del interno** y los del principal, del principal: cada hilo tiene
  los suyos, porque cada uno se lee aparte (sección 2.1).
- **Se guardan en su tabla** (`ticket_observers`), y no se deducen del texto al vuelo: **quitar a un
  observador lo puede hacer cualquier técnico o desarrollador**, y quien quita no es el autor del
  comentario que lo nombró, así que la lista tiene que poder vivir sin el texto (decisión 63).
- **Se añaden al escribir el comentario**: al guardar un comentario —nuevo o editado— **se añaden los
  que se mencionan de nuevo** (en un comentario nuevo, todos; en uno editado, los que no estaban antes
  en ese mismo comentario) y **se quitan los que ya no menciona ningún comentario del hilo**. Con eso
  valen las dos cosas que se decidieron: **borrar la mención editando el comentario quita al
  observador**, y **quitar a mano manda** —a quien se quitó no lo devuelve un guardado posterior, sólo
  una mención nueva, que es una llamada deliberada—.
- **Añadirlo a mano deja su entrada en el historial** (`observador_anadido`, con el nombre de quien
  entra), como quitarlo (`observador`). **Y sólo cuando de verdad se añade**: repetirlo no deja otra.
- **Quitarlo es una acción de la ficha**, con la misma regla que etiquetar: Soporte y Desarrollo. Y
  **queda en el historial** («María quitó a Juan de los observadores»), como todo lo que pasa en el
  ticket.
- **«Mención nueva» quiere decir nueva de verdad**: un **comentario nuevo** que vuelve a nombrar a
  quien se quitó **sí lo vuelve a añadir** —es una llamada deliberada, y por eso vale—; lo que no lo
  devuelve es **volver a guardar el mismo texto** (una edición del comentario que ya la llevaba).
- **A quien se quita sigue nombrado en el comentario**: no se reescribe el texto de nadie. La mención
  se lee como lo que es —una llamada que se hizo— y la lista de observadores dice quién lo sigue
  ahora.

### 2.3.2 Categorías y etiquetas

**Dos cosas distintas, con dos trabajos distintos** (decisiones 64 a 68):

| | **Categoría** | **Etiquetas** |
| --- | --- | --- |
| Cuántas | **Exactamente una**: no hay ticket sin categoría | **Las que hagan falta**, o ninguna |
| De dónde salen | De un **catálogo** cerrado que mantiene **sólo el Administrador** (decisión 83) | De un **catálogo** que mantiene **sólo el Administrador** (decisiones 83 y 84): **ya no nacen al escribirlas** |
| Quién la pone | **Quien abre el ticket** —el usuario incluido—, y **Soporte la puede corregir** después. **Desarrollo no**: no escribe en el principal | **Soporte y Desarrollo**, cada uno desde el ticket en el que trabaja: **Soporte en el principal y Desarrollo en el interno** (decisión 85). Son del principal, y el interno las hereda |
| Para qué sirve | **Agrupar**: «Red», «Software x», «Licencias»… es el *qué es* del ticket | **Matizar**: lo que no merece una categoría propia |
| Cómo se escribe | Como se llame en el catálogo | **En minúsculas y con guion medio en lugar de espacios**: `red-wifi`, `licencia-office`. Sin acentos ni mayúsculas, y el editor lo va dejando así mientras se escribe |
| Dónde vive | Columna del ticket, con clave ajena al catálogo | Su tabla, muchas por ticket |
| Dónde se ve | **Un chip en el renglón** del ticket, debajo del número, y en la ficha | Los mismos chips, al lado |

**La categoría es del caso, no de cada mitad**: se guarda en el principal y **el interno la hereda al
leerse**, como hereda el asunto (decisión 67). Lo mismo las etiquetas.

**«No hay ticket sin categoría» lo garantiza la base, no la pantalla** (decisión 65): la columna es
`NOT NULL` con clave ajena, así que ni una consulta a mano puede dejar un ticket sin clasificar. Para
que eso valga desde el primer día, **la migración crea la categoría «General» y se la pone a todo lo
que ya exista** —en producción no hay tickets, en desarrollo sí—, y **el catálogo no se puede quedar
sin ninguna categoría activa**: retirar la última se rechaza con una clave.

**Retirar una categoría es desactivarla** (decisión 66), como se desactiva una cuenta: deja de
ofrecerse al crear un ticket, y **los tickets que ya la tienen la conservan y la siguen enseñando**.
No se borra nada, y por eso **no puede existir un ticket sin categoría** —borrar la categoría de un
ticket en uso lo dejaría sin ella—.

**El catálogo de etiquetas lo mantiene sólo el Administrador** (decisiones **83** y **84**, que
corrigen la 72 y la 73): se pueden **crear** en la pantalla, **renombrar** —y el renombrado **cambia la
etiqueta en todos los tickets que la llevan**, no en uno— y **retirar** —y retirarla **la quita de esos
tickets**, con un aviso por delante de a cuántos afecta—, y **las tres cosas son del Administrador**.
**Ya no nacen al escribirlas**: poner en un ticket una etiqueta que no está en el catálogo se **rechaza**
(`tickets.etiqueta.desconocida`), así que **quien etiqueta elige de las que hay** —que es lo que hace el
modal de la ficha— y **los nombres nuevos los estrena el Administrador** en esa pantalla.

**Y quién pone las etiquetas en un ticket son los dos equipos** (decisión **85**): **Soporte en el
principal y Desarrollo en el interno**, cada uno desde el ticket en el que trabaja. Las etiquetas **son
del principal** y el interno las **hereda** (decisión 67), así que cambiarlas desde cualquiera de los dos
cambia **las mismas**.

**Las etiquetas se normalizan al guardar**: minúsculas, los espacios se vuelven guiones, y lo que no
sea letra sin acento, número o guion se cae. **Se sugieren las que ya existen** mientras se escriben,
para que no acaben siendo cinco maneras de decir lo mismo. **Nacen solas** —una etiqueta existe la
primera vez que alguien la escribe en un ticket— y desde el **2026-09-27 también tienen catálogo**
(decisión 72): su nombre vive en **`ticket_tag_names`** y `ticket_tags` dice **qué ticket la lleva**.
Así se pueden **crear sin que las lleve nadie**, **renombrar** —y el cambio vale para todos los tickets
que la llevan, porque el nombre está en un solo sitio— y **retirar** —que las quita de esos tickets—.
**Una sola fuente de verdad**: el nombre en el catálogo, y la tabla puente para lo demás.

**Y las dos sirven para lo mismo: encontrar** (decisión 68). La categoría se filtra con su chip, las
etiquetas con los suyos, y **las dos entran en la caja de búsqueda**: buscar «red» encuentra los
tickets de la categoría «Red» y los que llevan la etiqueta `red`.

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
| **`ticket_tags(tag_id)`** | **Filtrar por etiqueta** y contar cuántos tickets la llevan (sección 2.3.2) |
| **`ticket_tags(ticket_id, tag_id)`** único | La misma etiqueta no se repite en el mismo ticket |
| **`ticket_tag_names(normalized)`** único | Que el catálogo no acabe con `red-wifi` y `Red-Wifi` |
| **`ticket_categories(normalized)`** único | Que el catálogo no acabe con «Red» y «red» |
| **`tickets(category_id)`** | Agrupar y contar los tickets de una categoría |

La búsqueda por texto usa `ILIKE` sobre asunto y descripción, **y desde el 2026-09-27 también sobre el
nombre de la categoría y las etiquetas** del ticket (decisión 68). Es suficiente para el volumen de una
mesa de ayuda interna; cuando deje de serlo, se pasa a índices de texto completo con su propio
cambio.

## 3. Las transiciones

### 3.1 Ticket principal

| Desde | Hasta | Quién | Efectos |
| --- | --- | --- | --- |
| — | `nuevo` | Usuario (el suyo) o Soporte (en nombre de otro) | Se asigna el número; **se reparte por turnos**; historial `creado`; correo **al técnico asignado** |
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

### 3.3 El reparto de los tickets nuevos

Un ticket no nace en el aire: **se reparte por turnos** entre los técnicos de Soporte activos.

- **Sólo entran los técnicos activos** con papel `soporte`. Un técnico desactivado no recibe tickets.
- **El turno lo decide quién hace más tiempo que no recibe uno**: se ordenan los técnicos por la fecha
  de su último ticket asignado y elige el primero. Quien nunca ha recibido uno va delante.
- **Así no hace falta un puntero de turno** que guardar y que puede desincronizarse: la lista de
  tickets ya dice a quién le tocó la última vez. Si mañana entra un técnico nuevo, recibe el
  siguiente; si uno se desactiva, el turno sigue sin saltarse a nadie para siempre.
- **La cuenta la lleva `tickets`**, que es quien tiene la tabla; la lista de técnicos se la pide a
  `users` por su servicio, que es como se cruzan los módulos (`docs/arquitectura.md`, sección 4).
- **Si no hay ningún técnico activo**, el ticket queda **sin responsable** y se queda en la bandeja de
  Soporte. No hay a quién avisar, así que no se envía el aviso y se registra en el log.
- **Los tickets que crea Soporte también pasan por el turno**, aunque los cree quien los vaya a
  atender.
- **Reasignar siempre se puede**: el reparto decide el punto de partida, no el destino.

**Y todo esto es configurable**, porque no hay una única forma sensata de repartir. El administrador
elige, **por cada tipo de ticket**, dos cosas por separado:

| Ajuste | Valores | Qué significa |
| --- | --- | --- |
| **Asignación** | `ninguna` o `por turnos` | Si el ticket nace sin responsable o repartido |
| **Aviso** | `a nadie`, `a todo el equipo` o `al asignado` | A quién se le manda el aviso de ticket nuevo |

Los dos son independientes a propósito: se puede repartir sin avisar —los técnicos miran su bandeja—
o avisar a todo el equipo sin repartir, que es como funciona un ticket que nadie ha cogido todavía.
Con `asignación = ninguna`, el aviso `al asignado` no tiene sentido y la pantalla no lo ofrece.

Los valores por defecto, que son los que se han decidido aquí:

| | Asignación | Aviso |
| --- | --- | --- |
| **Tickets principales** | `por turnos` | `al asignado` |
| **Tickets internos** | `ninguna` | `a nadie` |

Los internos nacen sin repartir y sin aviso: la comunicación del equipo de Desarrollo es interna y
son ellos los que están atentos a su bandeja.

**Entre quién se reparte cada tipo** (aclarado el 2026-09-24, al implementarlo): el turno de los
tickets **principales** se lleva entre los usuarios activos con papel `soporte`, y el de los
**internos**, si el Administrador lo activa, entre los usuarios activos con papel `desarrollo`. Cada
tipo lleva **su propia cuenta**: quien hace más tiempo que no recibe uno de ese tipo va delante. El
reparto de los internos **viene apagado** (`ninguna`) y es el Administrador quien lo enciende
(corrección del responsable, 2026-09-24); la lista de gente de cada papel se la da `users` a
`tickets` por su servicio, igual que el resto.

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

**Tres formas de adjuntar**, y las tres llevan a lo mismo (decisión 43): **pegar** (`Ctrl+V`), con el
cursor **en la caja de escribir o en el recuadro de los archivos**; **arrastrar y soltar** sobre el
recuadro; y **elegir el archivo** con el botón. En una mesa de ayuda la mitad de los adjuntos son
capturas de pantalla, y pegarlas es lo que hace todo el mundo sin pensarlo.

- **Al pegar se acepta lo que venga en el pegado**, sea una captura o un archivo copiado del
  explorador, **si está en la lista cerrada de extensiones**; lo que no esté se rechaza **con el mismo
  aviso** que al elegirlo con el botón. Una sola regla para las tres formas.
- **El pegado solo funciona con el foco donde se escribe**: en la caja de comentario y en el recuadro
  de los archivos. Pegar en cualquier otra parte de la pantalla **no** añade nada, para que pegar no
  sorprenda a nadie.
- **Los archivos van con el comentario**: primero se publica el comentario y luego se suben sus
  archivos, uno a uno, con su `commentId`. **Si uno falla** (decisión 44), el comentario **se queda**
  —está escrito y lo ha visto quien lo escriba, y borrarlo dejaría un hueco en la conversación—, la
  caja se vacía, y se avisa de **qué archivo falta** con un botón para **reintentarlo** sin volver a
  escribir el comentario. Es lo que evita el comentario duplicado, que es lo que pasaba cuando el
  fallo se contaba como si no hubiera pasado nada.

**Cómo se sirven**: los tipos que se pueden previsualizar —las **imágenes** (`png`, `jpg`, `jpeg`,
`gif`, `webp`) y los **vídeos** (`mp4`, `webm`, `mov`, `avi`)— se sirven **en línea** para verlos
dentro del ticket, y el **PDF** también, porque lo que lo enseña es el visor. **Todo lo demás se sirve
siempre como descarga**, con `application/octet-stream` y `Content-Disposition: attachment`.

El caso que obliga a esta regla es el **`svg`**, que **no se admite**: puede llevar código dentro y el
navegador lo ejecuta al abrirlo, y desde el 2026-09-26 las imágenes se enseñan **en línea** dentro del
texto, que es justo lo que no se puede hacer con un archivo que se ejecuta. Lo mismo vale para
cualquier formato que el navegador interprete en vez de enseñarlo: lo que no está en la lista cerrada
se rechaza.

## 5. Los módulos y sus endpoints

Seis módulos, y no más:

| Módulo | Qué cubre |
| --- | --- |
| `auth` | Entrar, salir, quién soy, contraseñas |
| `users` | Cuentas, papeles, desactivación, origen |
| `tickets` | Principales, internos, comentarios, adjuntos y su historial |
| `mail` | Las plantillas de los correos, su edición y el envío |
| `settings` | La configuración de la instalación: idioma, apariencia, numeración y reparto |
| `ai` | Los dos resúmenes del ticket: **motivo** y **última acción**, en español y en inglés (`docs/modules/ai.md`). **Sin endpoints propios**: se llega por `tickets` |

Los dos tipos de ticket viven en **un solo módulo** (`tickets`): son la misma cosa contada de dos
maneras, y separarlos obligaría a que dos módulos compartieran tablas, que es justo lo que la regla
de modularidad prohíbe.

### Los de los demás módulos, en su sitio

Los endpoints de `auth`, `users` y `mail` **no se listan aquí**: cada uno los detalla en su documento
(`docs/modules/auth.md`, `docs/modules/users.md` y `docs/modules/mail.md`), y esta sección se queda
con los de `tickets`. Una lista por módulo, y ninguna repetida: dos listas del mismo contrato acaban
diciendo cosas distintas.

### `tickets`

| Método y ruta | Qué hace | Quién |
| --- | --- | --- |
| `GET /api/tickets` | Bandeja y listas. Filtros: `type`, `state`, `q` (texto o número), **`mine`** (sólo lo mío) y paginación | Según la matriz: cada uno ve lo que le toca |
| `POST /api/tickets` | Crea un principal. `requesterEmail` opcional, para crearlo en nombre de otro | Usuario, Soporte, Desarrollo |
| `GET /api/tickets/assignees` | Quién puede ser el responsable de cada tipo de ticket | Soporte, Desarrollo y Administrador |
| `GET /api/tickets/{number}` | Detalle, con comentarios, adjuntos e historial | Quien pueda verlo |
| `POST /api/tickets/{number}/comments` | Comenta. **El cuerpo puede ir vacío**: sus archivos suben después, así que aquí no se sabe si van a llegar (decisión 41) | Quien pueda comentar |
| `PATCH /api/tickets/{number}/comments/{id}` | Edita un comentario. Sólo su autor; deja la marca de editado | El autor del comentario |
| `DELETE /api/tickets/{number}/comments/{id}` | Borra un comentario: vacía el texto y deja la marca de eliminado | El autor del comentario |
| `POST /api/tickets/{number}/attachments` | Sube un adjunto (`multipart`) | Quien pueda comentar |
| `GET /api/tickets/{number}/attachments/{id}` | Descarga, con comprobación de permisos | Quien pueda ver el ticket |
| `PATCH /api/tickets/{number}` | Edita el asunto, la descripción o la categoría, con su marca. **Si el cuerpo trae sólo `tags`**, es la puerta de las etiquetas: **Desarrollo puede usarla desde el número de un interno**, y el cambio va **al ticket principal**, que es de quien son las etiquetas (decisión 85) | El solicitante y Soporte, en el principal. **Sólo etiquetas**: también Desarrollo, desde el interno |
| `POST /api/tickets/{number}/assign` | Asigna responsable | Soporte; Desarrollo en los internos |
| `POST /api/tickets/{number}/state` | Mueve el estado: `en progreso`, `en espera`, `resuelto`, `cerrado`, con `{ "state": "…", "comment": "…" }` —**el estado también se admite en `to`**, que es el nombre de esta tabla, y valen los dos—. **Al cerrar, el comentario es obligatorio** (decisión 81): sin él responde **422** con `tickets.cierre.sinComentario`, y con él **queda como un comentario de la conversación** | Quien pueda ese movimiento (sección 3) |
| `POST /api/tickets/{number}/escalate` | Escala. Exige motivo. Crea el interno o lo reabre | Soporte |
| `POST /api/tickets/{number}/reopen` | Reabre un principal cerrado | Soporte o el solicitante |
| `GET /api/tickets?view=…` | Además de los filtros de siempre, **`view`**: `assigned` (asignados a mí), `watching` (los que observo) o vacío (todo lo mío: asignados, abiertos, comentados y observados). Es el chip de «Mis tickets». **`view` matiza «lo mío», no lo sustituye**: `mine=1` sigue siendo el interruptor, y `assigned`/`watching` lo implican por sí solos. Las listas de «todo» —Tickets principales e internos— no mandan ni `mine` ni `view` y siguen devolviendo todo (decisión 52) |
| `GET /api/tickets/categories` | El catálogo: las activas para todo el mundo —el alta las necesita— y las retiradas para quien las mantiene |
| `POST /api/tickets/categories` | Crea una categoría | Soporte y Administrador |
| `PATCH /api/tickets/categories/{id}` | La renombra | Soporte y Administrador |
| `POST /api/tickets/categories/{id}/state` | La retira o la vuelve a poner | **Sólo el Administrador** |
| `POST /api/tickets/tags` | Crea una etiqueta | Soporte y Administrador |
| `PATCH /api/tickets/tags/{tag}` | **La renombra en todos los tickets que la llevan** | Soporte y Administrador |
| `DELETE /api/tickets/tags/{tag}` | **La retira y la quita de los tickets** | **Sólo el Administrador** |
| `GET /api/tickets/tags?q=…` | Las etiquetas que ya existen —**como mucho diez**, por uso— para **sugerirlas** mientras se escriben. **Incluye las del catálogo que no lleva ningún ticket**, con `tickets: 0` | Cualquiera que haya entrado |
| `GET /api/tickets/tags?all=1` | **El catálogo entero**, sin tope y por nombre: es el listado de la pantalla de mantenimiento, y sin él **una etiqueta recién creada que no lleva ningún ticket no se veía** (lo encontró la prueba del que hizo el catálogo, con diez etiquetas ya en uso). Con `all` se ignora `q`: es el catálogo, no una búsqueda | Cualquiera que haya entrado |
| `POST /api/tickets/{number}/observers` | **Añade a un observador a mano**, con `{ "accountId": 12 }` en el cuerpo. Contesta **201** con `{ "ticket": { …, "observers": [ … ] } }`, y **es idempotente**: añadir a quien ya observa no es un error ni deja otra entrada en el historial | Soporte y Desarrollo |
| `DELETE /api/tickets/{number}/observers/{id}` | Quita a un observador. Contesta `{ "status": "ok" }` —**no** la ficha—, y la pantalla recarga por su cuenta | Soporte y Desarrollo |
| `POST /api/tickets/{number}/insights` | Vuelve a pedir al motor el **motivo** y la **última acción** (`docs/modules/ai.md`) | Soporte y Desarrollo |

- **La API se dirige por número, no por `id`.** `ACME-2026-0042` es lo que la gente dice en voz alta
  y lo que aparece en los correos; el `id` interno no sale de la base.
- **`mine` es «lo mío», y se pregunta en el backend, no en la pantalla** (2026-09-26). Un ticket es
  mío si **me lo han asignado**, si **lo abrí yo** (como solicitante) o si **he comentado en su
  hilo**; en un principal el hilo es el principal y en un interno, el interno, y **el papel de quien
  comenta no cambia nada**: si Soporte comentó en un interno, ese interno aparece en sus tickets. El
  filtro se aplica a la consulta —con un `EXISTS` sobre los comentarios— y no trayendo todo para
  descartarlo después: la lista va paginada, y filtrar en la pantalla daría páginas de distinto
  tamaño. **Un usuario sigue viendo sólo los suyos**, con o sin `mine` (sección 3 de
  `docs/usuarios-y-permisos.md`).
- **El interno se lee con el mismo endpoint**, pasando su número (`INT-ACME-2026-0042`): es un
  ticket, y no necesita una ruta propia.
- **El único borrado que existe es el de un comentario, y es lógico**: la fila se queda con su
  marca. Un ticket no se borra, y un adjunto tampoco.
- **El solicitante de un ticket creado en nombre de otro se identifica por su correo**
  (`requesterEmail`), y no por su identificador: el correo **es** el identificador de las personas en
  este producto (`docs/usuarios-y-permisos.md`, sección 3), y es lo que Soporte tiene a mano cuando
  alguien llama. Así su pantalla no necesita leer la lista de cuentas, que es de otro módulo.
- **`GET /api/tickets/assignees` existe para que la pantalla no tenga que llamar a `/api/users`**:
  un módulo del frontend sólo habla con su propia API (`docs/arquitectura.md`, sección 4), así que la
  lista de responsables posibles la da este módulo, que ya le pide a `users` los técnicos activos para
  el turno. Devuelve las dos listas —Soporte y Desarrollo— y cada pantalla usa la suya.
- **`GET` y `PUT /api/settings` no están en esta lista a propósito**: son de `settings` y viven en
  `docs/modules/settings.md`, sección 7, que es donde se leen. Aquí estaban repetidos y se han
  quitado (2026-09-24): dos listas del mismo contrato acaban diciendo cosas distintas.
- **La configuración vive en `settings`**, y la pantalla de Configuración del frontend vive en
  `core` porque es de la instalación y no de un módulo: es la excepción que recoge
  `docs/arquitectura.md`, sección 4. La ve sólo un Administrador.
- Las listas van paginadas. El formato de error es el único de la API: **`{"error": "clave"}`**, con
  una **clave** y no un texto montado, definido en `shared/httpx`. El frontend decide cómo se dice y
  en qué idioma, porque la interfaz está en español y en inglés
  (`docs/interfaz-y-experiencia.md`). Es el mismo criterio que en Calibyou, donde los errores viajan
  como `reception.documents.template.unknownField`.

### El aviso de que te han etiquetado

Es la **plantilla número once** (`ticket.mentioned`), la única que estrena esta función, y va **a cada
persona mencionada**, una vez por comentario que la menciona:

| | |
| --- | --- |
| **Cuándo** | Al guardar un comentario con una mención **nueva** |
| **A quién** | A quien se acaba de mencionar, y **no** a quien ya estaba mencionado antes (mencionar dos veces lo mismo no manda dos correos) |
| **Qué lleva** | El número, el asunto y el enlace —como los demás avisos: **nunca el texto del ticket ni el del comentario** |
| **Asunto** | «Te han etiquetado en {{numero}}: {{asunto}}» y su inglés |

Y **los observadores reciben además los mismos correos de personal que el asignado** (decisión 61): el
del reparto, el del escalado, el de «Desarrollo necesita algo» y el de «el ticket vuelve a Soporte».

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
| `ticket_settings` | Una fila: el prefijo de la numeración **y el reparto y los avisos por tipo de ticket** |
| **`ticket_categories`** | **El catálogo de categorías** de la instalación, con su nombre, su nombre normalizado (único) y si está activa (sección 2.3.2) |
| **`ticket_tag_names`** | **El catálogo de etiquetas**: el nombre normalizado y único de cada una |
| **`ticket_tags`** | **Qué ticket lleva qué etiqueta**: la tabla puente con el catálogo, única por ticket y etiqueta, con quién la puso |

La puesta al día **también mete en `ticket_tag_names` las etiquetas que ya existían** —las que estaban
como texto en `ticket_tags`— y engancha los tickets con su `tag_id` antes de quitar la columna vieja, así
que una base con etiquetas de antes se pone al día sola (decisión 72).

Y **`tickets` gana `category_id`**, `NOT NULL` con clave ajena: **todo ticket nace clasificado**. La
migración **crea «General» y se la pone a todo lo que ya exista** antes de poner el `NOT NULL`, así que
una base con tickets a medias se pone al día sola (decisión 65).

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
- **Campos personalizados** y prioridad. **Las categorías salieron de esta lista el 2026-09-27**: entraron
  al aprobarse (sección 2.3.2), y esta línea se quedó diciendo lo contrario.
- **Tiempo real** (websockets) y contadores en vivo.
- **SLA, métricas e informes.**
- **Fusión y duplicado de tickets.**

## 8. Módulos de frontend

| Módulo | Pantallas |
| --- | --- |
| `tickets` | **Mis tickets** (lo mío, con chip de tipo para Soporte y Desarrollo), **Tickets principales**, **Tickets internos**, la bandeja del Administrador, el detalle del ticket (conversación, adjuntos, historial), el alta de ticket y el formulario de escalado |
| `users` | Lista de usuarios, alta, desactivación y cambio de origen |
| — (`core`) | Entrar, salir, recuperar contraseña y las guardas de sesión, como manda `docs/usuarios-y-permisos.md` |

**Las cuatro listas son la misma pantalla** con distinto filtro y distinto nombre: separarlas sería
duplicar la tabla, los filtros y la paginación para cambiar una condición. **Sus rutas son cuatro**
—`/tickets`, `/tickets/main`, `/tickets/internal` y la del Administrador, que es `/tickets`—, y las
dos del «todo» **sólo se ofrecen a Soporte y Desarrollo** (2026-09-26).

**El prefijo de la numeración no es una pantalla de `tickets`**, aunque aquí estuviera listado: se
configura en **Configuración**, con el resto de la instalación, y sus endpoints son de `settings`. Una
pantalla de este módulo que llamara a `/api/settings` rompería la regla dura de modularidad
(`docs/arquitectura.md`, sección 4). Corregido el 2026-09-24.

## 9. Lo que se decidió en los repasos

### Primer repaso: el modelo y la API

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **Tablas** | **Dos tablas** (`tickets` e `internal_tickets`), con dos columnas de destino en comentarios, adjuntos e historial |
| 2 | **Identificadores** | `bigserial` por dentro; la API se dirige por el número del ticket |
| 3 | **`svg`** | **Se admite**, pero siempre como descarga, nunca en línea. **Corregido el 2026-09-26**: el código nunca lo admitió —no está en la lista cerrada de extensiones, ni en el backend ni en la pantalla—, y la sección 2.3 lo dice así desde su primera enmienda. Esta fila y la sección 4 decían lo contrario, y quedan corregidas: **`svg` no se admite**, y menos ahora que las imágenes se enseñan en línea dentro del texto, que es exactamente lo que no se puede hacer con un archivo que el navegador ejecuta |
| 4 | **Re-escalar desde `resuelto`** | Directo, sin obligar a reabrir. Desde `cerrado`, hay que reabrir primero |
| 5 | **El prefijo del número** | **Pantalla de administración**, con una fila en `ticket_number_settings` |
| 22 | **La tabla del prefijo, renombrada** | Enmendado el 2026-09-24: pasa a `ticket_settings` y recoge también el reparto y los avisos, que no tenían tabla |
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

### Cuarto repaso: lo que faltaba al empezar a implementarlo (2026-09-24)

| # | Decisión o hallazgo | Quedó así |
| --- | --- | --- |
| 23 | **La fila que ya existía** | La sección 0 decía que a la matriz le faltaba «cerrar el ticket interno»: llevaba ahí desde el 2026-09-22. Corregida la sección |
| 24 | **Los endpoints repetidos** | `GET` y `PUT /api/settings` salen de la tabla de `tickets`: son de `settings`, que ya los lista |
| 25 | **El turno de los internos** | Entre los usuarios activos con papel `desarrollo`, con su propia cuenta por tipo. **Corrección del responsable**: los internos nacen **sin repartir**, y el reparto lo enciende el Administrador |
| 26 | **Reabrir el interno** | Fila nueva en la matriz: **Soporte**. Era la mitad que faltaba de «cerrar el interno» |
| 27 | **La bandeja del Administrador** | **Chip de tipo** (Principales · Internos · Todos): ve los dos y necesita separarlos |
| 28 | **El troceo del trabajo** | **El backend entero primero** —tablas, numeración, reparto, endpoints y los siete correos— y las pantallas después, contra una API que ya no se mueve |
| 29 | **Los correos** | Salen **en la misma tanda que el backend**: el aviso es parte de la transición que lo provoca, y se prueban leyendo el buzón |
| 30 | **El motivo de un re-escalado** | Va **como comentario en el ticket interno**, escrito por quien escala: el motivo original no se reescribe —es la justificación de la primera vez— y el nuevo queda donde Desarrollo lo lee |
| 31 | **Editar un ticket cerrado** | **No se edita**: primero se reabre. Es la misma regla que ya estaba decidida para los comentarios, y no deja dos maneras de decir lo mismo |
| 32 | **A quién van los dos avisos de dentro** | «Ha vuelto a tu bandeja» al **responsable del principal**, y «Desarrollo necesita algo» a **quien escaló**. Si esa cuenta ya no está activa, se avisa a todo el equipo de Soporte activo en su lugar: es lo que hace que «tu bandeja» signifique algo y no un correo para todos |
| 33 | **El responsable de un interno** | Sale de **Desarrollo**: usuarios activos con ese papel, que son también entre quienes reparte el turno si el Administrador lo enciende |

### Quinto repaso: las pantallas (2026-09-24)

| # | Decisión o hallazgo | Quedó así |
| --- | --- | --- |
| 34 | **«Nuevo ticket» para Soporte** | Su menú estrena la entrada, y son tres: Soporte puede crear un ticket en nombre de otra persona y ese permiso no tenía puerta. **Corrección del responsable** |
| 35 | **El escalado** | **Un diálogo**, con el motivo como campo principal y obligatorio, como el alta de cuentas |
| 36 | **Los adjuntos** | **Con vista previa** las imágenes y los PDF; lo demás se descarga siempre, que es lo que ya hacía el backend |
| 37 | **El prefijo y el reparto** | Se configuran en **Configuración**, no en una pantalla de `tickets`: sus endpoints son de `settings` y la regla dura de modularidad lo impide. La sección 8 de este documento los listaba como pantalla suya y queda corregido |
| 38 | **El detalle** | **Una sola pantalla para los cuatro papeles**: lo que cambia es lo que cada uno puede hacer dentro, no la pantalla |
| 39 | **Crear en nombre de otro** | El solicitante se identifica **por su correo** (`requesterEmail`), no por su identificador: es lo que Soporte tiene a mano y evita que su pantalla lea la lista de cuentas |
| 45 | **Los adjuntos van dentro del texto** | **Decisión del responsable, 2026-09-26**: lo que se escribe en la descripción de un ticket y en un comentario **es un texto con formato**, y **los archivos que se adjuntan viven dentro de él**, en el sitio donde se pusieron y en el orden en que se cargaron. **Es la decisión más cara del documento**: el texto deja de ser texto plano, y con él conviven la lista blanca de etiquetas, la búsqueda que quita las etiquetas y la migración que convierte lo que ya existe |
| 46 | **Qué se ve de cada adjunto** | **Decisión del responsable, 2026-09-26**: **imágenes y vídeos en línea** —con el ancho de la conversación y un reproductor con controles—; **el PDF como enlace**, que al pulsarlo abre el visor; **todo lo demás como enlace**, que al pulsarlo descarga. **La vista previa se ve también mientras se edita**, que es lo que se pidió: el cuadro de escribir es un editor con formato, no un cuadro de texto |
| 47 | **Cuatro extensiones de vídeo** | **Elección del responsable, 2026-09-26**, y **más ancha que mi recomendación**: se admiten `mp4` y `webm` —que cualquier navegador reproduce— **y también `mov` y `avi`**, que son formatos de cámara y que muchos navegadores **no saben reproducir**. Consecuencia dicha y aceptada: si el navegador no puede, el reproductor no arranca y queda el botón de descargar del visor. **No crece el tope de 25 MB**, así que son unos segundos de vídeo |
| 48 | **El visor** | **Decisión del responsable, 2026-09-26**: el modal que ya existía —imagen o PDF dentro— pasa a ser **más grande** y lleva **Descargar** y **Abrir en una pestaña**. Es también la forma de ver una imagen a tamaño completo, porque en línea va con el ancho de la conversación |
| 49 | **Lo que el texto no nombra, al final** | **Decisión del responsable, 2026-09-26**: un adjunto que el texto no referencia **se sigue enseñando en su lista, al final**, como hasta ahora. Es lo que hace que **los tickets que ya existen se vean igual** y que no se pierda nada de lo colgado antes de este cambio |
| 41 | **Un comentario puede ser sólo un adjunto** | **Elección del responsable, 2026-09-25**, y **no la que yo recomendaba**: yo proponía seguir exigiendo texto; él decidió que se pueda mandar sólo el archivo. **Consecuencia dicha y aceptada**: `body` **deja de ser obligatorio** en la tabla, porque cuando el comentario se crea sus archivos todavía no han subido y el backend no puede saber si van a llegar; la regla «texto **o** archivos» se sostiene en la pantalla, que es quien lo sabe |
| 42 | **Editar un comentario sigue exigiendo texto** | **Decisión del responsable, 2026-09-25**: la puerta de atrás se cierra. Dejar en blanco un comentario que ya se escribió no es lo mismo que nacer sin texto, y lo que se quiere en ese caso es borrarlo |
| 43 | **Pegar, arrastrar y elegir el archivo** | **Decisión del responsable, 2026-09-25**: las tres formas llevan a lo mismo y comparten la lista cerrada de extensiones y su aviso. **Al pegar se acepta lo que venga** —captura o archivo copiado— si es de un tipo admitido, y **el pegado sólo funciona con el foco en la caja de escribir o en el recuadro**, para que no añada archivos sin querer |
| 44 | **Si una subida falla, el comentario se queda** | **Decisión del responsable, 2026-09-25**: el comentario ya está publicado, así que no se borra; se avisa de **qué archivo** no subió y se puede **reintentar** al mismo comentario. Antes, un fallo ahí dejaba la caja con el texto otra vez y lo más probable era acabar con dos comentarios iguales |
| 50 | **Quién reasigna, y desde dónde** | **Decisión del responsable, 2026-09-26**: **la reasignación se hace dentro del ticket** —en su ficha, con el selector de responsables y su botón— y **«Asignármelo» desaparece de la bandeja**, en la tabla y en las tarjetas. Las reglas no cambian: en el principal reasigna **Soporte**, a **cualquier técnico activo**; en el interno reasignan **Soporte y Desarrollo**, a **cualquier desarrollador activo**. Así `support1` pasa su ticket a `support2` desde el ticket, que es donde se ve a quién se le pasa; el atajo de la lista asignaba sin más, y era lo que dejaba la reasignación sin hacer |
| 56 | **Dos columnas que redacta el motor de IA** | **Decisión del responsable, 2026-09-27**: la lista estrena **«Motivo»** —de qué va el ticket— y **«Última acción»** —qué fue lo último que pasó—, **redactadas por el módulo `ai`** en español y en inglés (`docs/modules/ai.md`). **Los textos viven en la tabla de `ai`**, no en la de tickets: aquí sólo se piden y se leen, y el módulo de tickets **no cambia de esquema** |
| 57 | **Los dos nombres de la cabecera** | **Decisión del responsable, 2026-09-27**: **«Número» pasa a «Ticket»** —es lo que la columna enseña y como se llama lo que hay dentro— y **«Última actualización» pasa a «Fecha actualización»**, que dice que es una fecha y no un suceso. La columna de la derecha pasa a ser la fecha, y el suceso lo cuenta **«Última acción»** |
| 54 | **El texto y el código también se adjuntan** | **Decisión del responsable, 2026-09-26**, al no poder mandar un `.sql`: la lista cerrada se queda **cerrada** y se le añaden **texto y código** —`sql`, `json`, `xml`, `yml`/`yaml`, `ini`, `conf`, `cnf`, `sh`, `bash`, `bat`, `ps1`, `py`, `js`, `ts`, `java`, `php`, `go`, `cs`, `rb`, `pl`, `kt`, `rs`, `swift`, `c`, `h`, `cpp`, `hpp`, `vue`, `jsx`, `tsx`, `htaccess`, `env`, `properties`, `diff`, `patch`, `bak`, `old`— **y comprimidos** (`tar`, `gz`, `tgz`, `7z`). **No se pintan nunca en línea**: se sirven como `text/plain` y se descargan, que es lo que permite admitir un `sh` sin abrir un agujero. **La alternativa que él descartó** era abrir la lista entera con una lista negra: más cómodo, y deja entrar cualquier cosa que nadie ha pensado |
| 55 | **La imagen y el vídeo, con tope y al visor** | **Decisión del responsable, 2026-09-26**: en la descripción y en los comentarios la imagen y el vídeo se ven **como mucho a 480 × 360**, sin deformarse, y **al pulsarlos se abren en el visor grande** —con **Descargar** y **Abrir en una pestaña**—, que es donde se ven enteros. **El vídeo pasa a ser una miniatura con su botón de reproducir** y se reproduce **en el visor**: antes mandaba su reproductor dentro de la conversación y el visor sólo salía cuando el navegador no sabía reproducirlo. **Lo que se guarda no cambia** —la referencia sin `controls`—, y el ancho tampoco es un dato guardado: es cómo se pinta. **El tope se escribe `min(100%, 480px)`**, y no `480px`: un máximo sólo limita, así que con el valor fijo una imagen de 480 px **se salía de la pantalla en un móvil de 412** y el documento se ensanchaba a 530 —lo encontró la capa de Playwright, porque el clic en un botón de la misma tarjeta dejó de funcionar— |
| 64 | **Categorías y etiquetas son de `tickets`** | **Decisión del responsable, 2026-09-27**: el catálogo de categorías vive en este módulo y su pantalla es **una pantalla de tickets** —«Categorías y etiquetas»—, no una sección de Configuración: son datos de los tickets, y un módulo del frontend sólo habla con su propia API (`docs/arquitectura.md`, sección 4). **Lo mantienen Soporte y el Administrador**, y **retirar** una categoría es **sólo del Administrador** |
| 65 | **No hay ticket sin categoría, y lo garantiza la base** | **Decisión del responsable, 2026-09-27**: **una categoría por ticket, obligatoria**. Se implementa con la columna **`NOT NULL`** y clave ajena, no con una comprobación en la pantalla, y la migración **crea «General» y se la pone a todo lo que ya exista**. El catálogo **no se puede quedar sin ninguna categoría activa**: retirar la última se rechaza |
| 66 | **Retirar una categoría es desactivarla** | **Decisión del responsable, 2026-09-27**, sobre la mía de prohibir borrar la que esté en uso: **retirar es desactivar** —deja de ofrecerse al crear un ticket y **los tickets que la tienen la conservan**—, como se desactiva una cuenta. **No se borra ninguna fila**, y es lo que hace que «no hay ticket sin categoría» se pueda cumplir siempre |
| 67 | **De la categoría, del principal; el interno hereda** | **Decisión del responsable, 2026-09-27**: la categoría y las etiquetas son **del caso**, se guardan en el principal y **el interno las hereda al leerse**, como hereda el asunto. Clasificar el mismo caso dos veces —una por cada mitad— sería pedir trabajo de más |
| 68-quater | **El recuento del catálogo cuenta principales, y el filtro respeta la lista** | **Aclarado el 2026-09-27**, porque parecía un desajuste: el número que enseña el catálogo al lado de cada categoría son **los tickets principales** que la tienen, y **filtrar por esa categoría en la lista de internos devuelve los internos** —que la heredan de su principal—. Es decir, **no es la misma pregunta**: «cuántos tickets tienen esta categoría» (los principales: la categoría es suya) y «qué tickets salen al filtrar por ella en esta lista» (los de esa lista, y el interno la hereda para poder buscarse). Se deja así, y **se dice aquí** para que nadie lo lea como un fallo |
| 68-ter | **Las claves del catálogo y de las etiquetas** | `tickets.category.required` (422), `tickets.category.notFound` (404), `tickets.category.inactive` (422), `tickets.category.name.required` (422), `tickets.category.duplicate` (409), `tickets.category.lastActive` (409), `tickets.category.forbidden` (403), `tickets.tag.required` (422) y `tickets.tag.tooLong` (422) —**no hay `tag.invalid`**: la normalización convierte en vacío lo que no vale, y eso es `tag.required`, así que una clave que no puede salir no se pone—. Todas con su texto en español y en inglés (`docs/interfaz-y-experiencia.md`, sección 8) |
| 69 | **Un solo botón de búsqueda, con modal** | **Corrección del responsable, 2026-09-27**, al usarlo: «ya tengo tres filas para cuatro criterios, y puedo tener más». Los criterios —texto, vista, tipo, estado, categoría y etiqueta— viven en **un modal** que abre el botón «Buscar», y **en la cabecera de la lista queda un resumen** de lo que está filtrado con su botón de quitar: una lista limpia y, a la vez, se entiende por qué salen pocos tickets. El desbordamiento que lo provocó —la fila de categorías saliéndose de la pantalla— **también se arregla** (sección 5) |
| 70 | **La vista del ticket, por papel** | **Corrección del responsable, 2026-09-27**: al abrir un ticket, **Soporte ve el principal** y **Desarrollo el interno**, y los dos pueden cambiar a Principal · Los dos · Interno cuando quieran. **Si abres el número de un interno, se ve el interno** aunque seas Soporte: es lo que espera quien pulsa ese enlace. Y **«Los dos» enseña sólo las conversaciones** —las dos columnas de la izquierda—, sin las fichas: con las cuatro columnas «se ve muy saturado» |
| 71 | **«Qué se puede hacer» se va** | **Corrección del responsable, 2026-09-27**: el título de la tarjeta de acciones de la ficha **se malinterpretaba** y se retira. Los botones siguen diciendo lo que hacen, que es lo que hacía falta |
| 72 | **Las etiquetas se mantienen, con el mismo peso que las categorías** | **Corrección del responsable, 2026-09-27**: en la pantalla no había forma de **crear ni de corregir** una etiqueta y **las categorías se llevaban todo el protagonismo**. Se elige **una pantalla con dos mitades iguales** —no dos módulos—, y las etiquetas se pueden **crear**, **renombrar** y **retirar**. Renombrar **vale para todos los tickets** que la lleven, y retirar **la quita de ellos**, avisando antes de a cuántos afecta: una etiqueta vive en muchos tickets, y cambiar una no puede cambiar sólo uno. **Siguen naciendo solas** al escribirlas en un ticket |
| 73 | **Quién mantiene las etiquetas** —**corregida por la 83**: lo mantiene sólo el Administrador** — | **Propuesta mía, 2026-09-27**, por simetría con las categorías (decisión 64), **implementada así el 2026-09-28** por el encargo del responsable —él pidió poder crearlas y corregirlas, y no dijo quién—, y **pendiente de que la corrija si no es lo que quiere**: crear y renombrar, **Soporte y el Administrador**; **retirar, sólo el Administrador**, porque retirar una etiqueta toca todos los tickets que la llevan. **Desarrollo tampoco las mantiene**: es el catálogo, y lo mantiene quien mantiene el de categorías |
| 80 | **El control de estados, y el comentario obligatorio al cerrar** | **Corrección del responsable, 2026-09-29**, sobre la 78: el desplegable **no es una lista de acciones**, es **para cambiar de estado el ticket**, y se presenta **la lista de estados** en los principales y en los internos. Y **cerrar siempre se confirma en una ventana y pide un comentario obligatorio**, también desde el backend (ver la 81). La lista enseña **los estados a los que el ticket puede pasar ahora**, con **el actual como primera opción** (marcado, sin poder elegirlo), y **el que necesita un texto lo pide en esa misma ventana**: el motivo al escalar, qué se ha hecho al resolver, la pregunta al ponerse en espera, y **el comentario al cerrar**. **Las acciones con nombre no desaparecen: son estados** —lo comprobó quien lo implementó en el propio código—: en un interno, **«En espera» es devolver el caso a Soporte sin cerrarlo** (regla 2 de la sección 3.3) y **«Cerrado» sin resolver es devolverlo y cerrarlo** (regla 8), que es lo que hacían «Necesito algo de Soporte» y «No es un cambio de código»; y en un principal, **«En espera» es preguntarle al usuario** y **«Cerrado» desde cerrado es reabrirlo**, que es lo que hacían «Preguntar al usuario» y «Reabrir» |
| 82 | **Las etiquetas son de Soporte y Desarrollo, y el alta no las pide** | **Decisión del responsable, 2026-09-29**: **el usuario no pone etiquetas** —ni las ve en su ticket—: son una forma de **clasificar y filtrar** que usan **Soporte y Desarrollo**. Así que **el alta no lleva campo de etiquetas** (la categoría sí, que es obligatoria y la elige quien abre el ticket) y **la línea de etiquetas de la ficha no se le enseña al usuario**. Y se etiqueta **mejor**: en la ficha, un botón abre **un modal con todas las etiquetas del catálogo** —con cuántos tickets lleva cada una—, un **buscador**, y **casillas para marcar y desmarcar varias a la vez**; las que el ticket ya lleva salen marcadas y **«Guardar» aplica todo de una vez** (y «Cancelar» no cambia nada) |
| 90 | **El número manda, y el cartel de «interno» sobra** | **Decisión del responsable, 2026-09-29**: en la cabecera del ticket, **el número se ve más** —es el texto más grande de la fila, y crece en pantallas anchas— y **desaparece el cartel «El ticket interno»** que iba al lado: con el número del interno y su estado ya se sabe cuál se está mirando, y el cartel era ruido. **Y los botones de sólo icono se achican**: medían 44 px y los lápices de la ficha pesaban demasiado; ahora 36, que sigue por encima de lo que se puede pulsar |
| 89 | **La lista de a quién etiquetar, junto al cursor** | **Decisión del responsable, 2026-09-29**: al escribir `@` en un comentario, la lista de personas **sale junto al cursor** —donde se está escribiendo— y no encima del bloque del comentario, que es donde estaba y donde el ojo no la encuentra. La lista se coloca **bajo la línea que se escribe**, y si no cabe ahí, **encima**, siempre dentro de la pantalla. El **botón «Etiquetar»** sigue abriendo la misma lista en su sitio de siempre, para quien no use el atajo |
| 87 | **La cabecera enseña primero el ticket de tu equipo** | **Decisión del responsable, 2026-09-29**: en la cabecera del ticket, **el número más relevante es el del ticket con el que trabaja quien mira**: para **Desarrollo, el interno** —y el principal queda como referencia, en apagado—; para **Soporte y el Administrador, el principal** —y el interno, como referencia—. **Y desaparecen los enlaces «Ver el ticket interno» y «Ver el principal»** de la cabecera de cada columna: para pasar de uno a otro ya está **el conmutador** «Principal · Los dos · Interno», que es el sitio donde se elige qué se ve |
| 88 | **«En espera» sólo se confirma; cerrar sigue pidiendo el porqué** | **Decisión del responsable, 2026-09-29**, que corrige parte de la 80: pasar el ticket a **«En espera»** —en el principal, que era «Preguntar al usuario»; en el interno, que era «Necesito algo de Soporte»— **se confirma en su ventana y ya no pide ningún texto**. **Cerrar sigue exigiendo el comentario** (decisión 81): es el único estado que lo pide. El motivo al escalar y qué se ha hecho al resolver **siguen pidiéndose**, que son los que de verdad hacen falta para entender el ticket después |
| 86 | **Las listas: el número solo, y una columna de clasificación** | **Decisión del responsable, 2026-09-29**, al mirar la bandeja: en las tablas de resumen, la columna **«Ticket» enseña sólo el número** (los chips de la clasificación la ensuciaban), el **«Estado»** se queda con el estado, y **la categoría y las etiquetas se agrupan en una columna nueva: «Clasificación»**. Es donde se leen juntas, que es lo que son: cómo está clasificado el ticket |
| 85 | **Las etiquetas las ponen los dos equipos, y son del principal** | **Decisión del responsable, 2026-09-29**, al aclararse la 82: las etiquetas **se aplican al ticket principal** y **el interno las hereda** (decisión 67), y **tanto Soporte como Desarrollo pueden ponerlas, cambiarlas y quitarlas** en los tickets. En la pantalla, cada uno desde **el ticket en el que trabaja** —Soporte en el principal, Desarrollo en el interno—, y **cambiarlas desde cualquiera de los dos cambia las del principal**, porque son las mismas: no hay dos juegos de etiquetas. **Lo que no cambia**: Desarrollo no escribe el texto del principal (asunto, descripción o categoría), y el catálogo lo sigue curando el Administrador (decisión 84) |
| 84 | **El catálogo se cura: las etiquetas no nacen al escribirlas** | **Decisión del responsable, 2026-09-29**, al comprobarse que se contradecían la 72 y la 83: hasta hoy, escribir una etiqueta nueva en un ticket **la creaba en el catálogo** —medido: Soporte ponía `etiqueta-nueva-de-soporte` en un ticket, respondía **200** y aparecía en el catálogo—, así que «sólo el Administrador crea las etiquetas» no se cumplía. Ahora **etiquetar un ticket con un nombre que no está en el catálogo se rechaza** (`422`, `tickets.etiqueta.desconocida`) a quien no sea Administrador: **quien etiqueta elige de las que hay** —que es lo que hace el modal de la ficha— y **los nombres nuevos los estrena el Administrador en el catálogo**. El Administrador **no toca tickets** (mira sin botones), así que su camino es la pantalla del catálogo, y desde ahí la etiqueta queda disponible para todos |
| 83 | **El catálogo lo mantiene sólo el Administrador** | **Decisión del responsable, 2026-09-29**, que **cierra la duda que quedaba abierta en la 73**: **crear, renombrar y retirar** —tanto **categorías** como **etiquetas**— es **sólo del Administrador**. Soporte y Desarrollo **usan** el catálogo —clasifican tickets y filtran con él— pero **no lo cambian**. Es la 73 corregida: allí se propuso que Soporte pudiera crear y renombrar, y el responsable lo ha decidido al revés |
| 81 | **Un ticket no se cierra sin decir por qué** | **Decisión del responsable, 2026-09-29**: cerrar —lo haga Soporte, Desarrollo o el propio solicitante— **exige un comentario**, y lo exige **el backend**, no sólo la pantalla: se pide en la ventana de confirmación y se guarda **como un comentario de la conversación**, así que quien lo lea ve por qué se cerró. Por la API, cerrar sin comentario responde **422** con `tickets.cierre.sinComentario`. Es la misma regla que la categoría obligatoria: lo que no puede faltar lo garantiza el servidor |
| 78 | **Un solo control para mover el ticket, en vez de cuatro botones** | **Decisión del responsable, 2026-09-28**: los botones sueltos de la ficha —«Preguntar al usuario», «Escalar», «Resolver», «Cerrar» en el principal, y «Necesito algo de Soporte», «No es un cambio de código», «Resolver» y «Cerrar» en el interno— **se juntan en un desplegable**, para que la ficha no sea un muro de botones. Se decide **por las acciones y no por los estados** (`docs/interfaz-y-experiencia.md`, sección 8): el desplegable enseña **lo que hace cada una** —«Empezar», «Preguntar al usuario», «Escalar», «Resolver», «Cerrar», «Reabrir»—, y **el estado sigue viéndose en su chip**. Elegir una la ejecuta, y **si esa acción pide algo** —el motivo, la pregunta, qué se ha hecho— **sale su ventana como hasta ahora**; las que preguntan siguen preguntando. **Dónde vive**: con un solo ticket a la vista, **en la fila de arriba**, junto a «Principal · Los dos · Interno»; con **los dos** a la vista, **en la cabecera de cada columna**, porque cada ticket tiene su estado y desde un solo control no se sabría a cuál se le cambia. **Y «Escalar a Desarrollo» pasa a «Escalar»**: la palabra —y el estado «Escalado»— ya dicen de qué equipo es |
| 77 | **La cabecera del interno, y las ventanas centradas** | **Corrección del responsable, 2026-09-28**: (1) en la cabecera del ticket, **también el número del interno** —abriendo un interno sólo se veía el del principal— y **el botón dice «Volver»**, no «Volver a la bandeja»; (2) **las ventanas salían en la esquina**, porque la preparación de Tailwind pone `margin: 0` a todo y el `<dialog>` nativo se centra con `margin: auto`: se les devuelve, con altura máxima y desplazamiento propio; y (3) **el ✕ de la ventana se llama «Cerrar»**, no «Cancelar», que era el mismo nombre que el botón de dentro |
| 76 | **Los iconos del ticket son los del proyecto** | **Corrección del responsable, 2026-09-28**: el botón de copiar el número y el lápiz de editar «no se distinguen bien». Eran dos caracteres —`⧉` y `✏`—, y el segundo **sale como emoji de colores** según la fuente del equipo. Ahora son **svg propios** de `app-icono` —dos hojas superpuestas para copiar, un visto para confirmarlo y un lápiz inclinado para editar—, y **`app-boton` recibe el icono por su nombre**, no por su carácter. Es lo que ya hacía el menú: dibujos propios, sin bibliotecas de iconos |
| 75 | **La cabecera de una lista, en una sola fila** | **Decisión del responsable, 2026-09-28**: el **nombre del módulo**, los **filtros** y **«Nuevo ticket»** van **en la misma fila** —eran dos renglones—, para ganarle ese alto a la tabla de tickets. Vale para las tres listas (`Mis tickets`, `Tickets principales` y `Tickets internos`); en un ancho corto se reparten solos |
| 74 | **Se edita donde se muestra, y los observadores se añaden a mano** | **Corrección del responsable, 2026-09-28**, al usar el ticket: no encontraba dónde cambiar la **categoría** ni las **etiquetas** —estaban en un formulario aparte, dentro de la tarjeta de acciones—, y el asunto y la descripción, en un diálogo. A partir de aquí, **cada dato lleva su control donde se lee**: la categoría con su desplegable, las etiquetas con sus fichas y su campo para añadir, los observadores con su «Añadir» —**nuevo**: hasta ahora sólo se llegaba por una mención—, y el asunto y la descripción con su lápiz en el sitio. **Los formularios aparte se retiran.** Quien puede editarlo no cambia: el solicitante en su ticket y Soporte |
| 68-quinquies | **El catálogo de etiquetas se pide entero, y el de sugerencias sigue capado** | **Corrección mía, 2026-09-27**, al encontrar el fallo: la misma ruta servía para dos cosas con necesidades opuestas —sugerir mientras se escribe (diez, por uso) y mantener el catálogo (todo)—, y con diez etiquetas ya en uso **una recién creada no se veía en la pantalla de mantenimiento**, que es justo para lo que se pidió. La ruta acepta **`all=1`**, y el listado completo va **por nombre** (orden estable: agrupar por uso haría bailar la lista al usar o renombrar) |
| 68-bis | **Lo que la implementación fijó, y el documento no** | **Anotado el 2026-09-27**, al cerrar la parte del backend: **una etiqueta no pasa de 32 caracteres** (`tickets.tag.tooLong`) —cabe `licencia-office-2026`, se lee en un chip y no estropea la lista—; **las sugerencias son como mucho 10**, por uso y luego por nombre; el catálogo **enseña las retiradas sólo a Soporte y Administrador** (a los demás les llegan sólo las activas, que es lo que el alta necesita); **la «General» de fábrica no la creó nadie** (`created_by_id` nulo) y las etiquetas sí llevan quién las puso; la clave ajena de la categoría es **`RESTRICT`** —no se borra una categoría que tenga tickets, y por eso **retirar es desactivar**— y las etiquetas van con **`CASCADE`**; la normalización de la **categoría** sólo baja a minúsculas, quita acentos y recorta (los espacios de dentro se respetan, porque es un nombre que se lee) mientras que la de la **etiqueta** además convierte espacios en guiones, junta los repetidos y tira lo que no sea letra, número o guion —por eso `red_wifi` es `redwifi`—; y el filtro de los **internos** une la tabla de los principales, porque la categoría y las etiquetas son suyas |
| 68 | **Las dos sirven para filtrar y para buscar** | **Decisión del responsable, 2026-09-27**: categoría y etiquetas tienen **sus chips de filtro** en la lista y, además, **entran en la caja de búsqueda**: buscar «red» encuentra los tickets de la categoría «Red» y los de la etiqueta `red`. **Las etiquetas son texto libre en minúsculas y con guion medio** (`red-wifi`) —**corrección del responsable** sobre «en minúsculas» a secas: los espacios se vuelven guiones—, con **sugerencias** de las que ya existen, y sin catálogo que mantener |
| 58 | **La mención se guarda por identificador** | **Decisión del responsable, 2026-09-27**: el texto nombra a una persona como `<span data-mencion="12">María Pérez</span>`, **nunca por el nombre escrito a mano**: los nombres se repiten y cambian, y una mención guardada por nombre se rompe sola. Es la misma idea que el `data-adjunto`. La etiqueta y el atributo entran en la **lista blanca del saneador**, con la misma regla que lo demás |
| 59 | **Quién etiqueta, y a quién** | **Decisión del responsable, 2026-09-27**: **Soporte y Desarrollo**, y **sólo a técnicos y desarrolladores activos**. Al guardar el comentario se comprueban una a una, y una que no valga **rechaza el comentario entero** con `tickets.mention.notAllowed`. **El usuario no etiqueta**: no ve la lista de técnicos ni tiene por qué |
| 60 | **Observador no es asignado** | **Corrección del responsable, 2026-09-27**, sobre mi propuesta, que hablaba sólo de «partícipes»: el ticket tiene **como mucho un asignado —el que atiende— y muchos observadores —los que lo siguen—**, y no son lo mismo. **El asignado sigue siendo opcional**, como estaba: el responsable corrigió su propia corrección al ver que un ticket sin técnicos activos no tiene a quién repartirse. **Ser observador no da permisos**: sólo hace que el ticket esté en tu bandeja |
| 61 | **Los observadores reciben los mismos avisos que el responsable** | **Decisión del responsable, 2026-09-27**, y **no la que yo recomendaba** —yo proponía sólo el aviso del etiquetado—: quien observa recibe **los mismos correos de personal** que el asignado, además del de que le etiquetaron. **Consecuencia dicha y aceptada**: un ticket con tres observadores manda cuatro correos por movimiento. Los avisos que van **al usuario** (lo resolvimos, lo cerramos, te preguntamos algo) **no se duplican** a los observadores: son de quien abrió el ticket |
| 62 | **Cómo se ven los observados** | **Decisión del responsable, 2026-09-27**: «Mis tickets» estrena un **chip —Asignados · Observo · Todos—**, como el de tipo de los técnicos, y el chip arranca en **Asignados**. La ficha del ticket lleva su lista de **«Observadores»** |
| 63 | **Quitar a un observador** | **Decisión del responsable, 2026-09-27**: **se puede quitar**, y **cualquier técnico o desarrollador** puede hacerlo —no sólo quien lo etiquetó—. Eso obliga a **guardar los observadores en una tabla** (`ticket_observers`), porque quien quita no es el autor del comentario y no puede reescribirlo, y deja la regla que hay que decir en voz alta: **quitar manda sobre lo ya escrito** —la mención se queda en el comentario, como prueba de la llamada que se hizo, pero **no vuelve a añadirlo**; vuelve sólo si alguien lo menciona **de nuevo**—. Quitar a alguien **queda en el historial** |
| 51 | **Qué es «lo mío»** | **Decisión del responsable, 2026-09-26**, **ampliada el 2026-09-27**: un ticket es mío si **me lo han asignado**, si **lo abrí yo**, si **he comentado en su hilo** o si **me han etiquetado** en él (observar es una forma de ser parte). El responsable **corrigió** la propuesta —que era «asignado o abierto»— añadiendo **donde he participado**, porque un ticket en el que ya se ha hablado se sigue hasta el final. **Participar es comentar en el hilo de ese ticket**, y el papel de quien comenta no cambia nada (un interno donde comentó Soporte es un ticket de Soporte) |
| 52 | **Tres listas: lo mío, todo lo principal y todo lo interno** | **Decisión del responsable, 2026-09-26**: la **Bandeja** es **Mis tickets** —lo mío— y nacen **dos pantallas más**, **Tickets principales** y **Tickets internos**, para **Soporte y Desarrollo**, que son los dos papeles que pueden verlo todo. **Cada uno entra por el suyo**: Soporte por principales y Desarrollo por internos, con el chip de tipo para cambiar, porque un técnico de Soporte tiene internos donde ha comentado y uno de Desarrollo puede haber abierto un principal. **El Administrador se queda con la bandeja de siempre** —los dos tipos, con su chip y de sólo lectura—: no tiene tickets propios, así que «Mis tickets» no le dice nada |
| 53 | **Los textos y el menú que se van** | **Decisión del responsable, 2026-09-26**: las filas de la lista dicen **«Abrir»** (`Open`) y no «Abrir el ticket» —en una columna de acciones, la palabra «ticket» sobra—; **fuera el texto** «Lo que lleva más tiempo esperando va primero. Lo cerrado no estorba.» —la regla sigue, callada: el orden se ve en la lista—; **«Nuevo ticket» sale del menú lateral**, porque se confundía con un módulo, y se queda **dentro de la bandeja**, que es donde se crea; y **el idioma pasa a ser un desplegable** como el del tema, con su mismo ancho y sin la etiqueta a la vista |
| 40 | **Quién puede ser el responsable** | Lo dice **este módulo** (`GET /api/tickets/assignees`), no `users`: la pantalla del frontend sólo puede hablar con su propia API |

## 10. Qué habilita este documento

Aprobar `docs/modules/tickets.md` cierra la lista de módulos y deja **`auth`, `users` y `tickets`** listos
para implementarse, en ese orden:

1. `auth` — sin sesión no hay nada que proteger.
2. `users` — Soporte necesita dar de alta a la gente, y el prefijo se configura.
3. `tickets` — el producto.

`flujos.md` queda después: describe el recorrido paso a paso de cada caso (alta, triaje, escalado,
resolución, cierre y reapertura) con sus correos, y no bloquea empezar por `auth`.
