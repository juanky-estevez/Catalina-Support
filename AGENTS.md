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
| `docs/propósito-y-alcance.md` | **as-built**: define el producto (los dos equipos, el modelo de tickets, la numeración y los estados), que es lo que hace hoy `backend/modules/tickets/**` |
| `docs/usuarios-y-permisos.md` | **as-built**: permisos, acceso, **un método de entrada a la vez**, ciclo de vida de las cuentas y la cuenta de fábrica. Lo implementan `auth`, `users`, `shared/authz` y `shared/middleware` |
| `docs/modules/tickets.md` | Cubre `backend/modules/tickets/**`, las diez tablas de `tickets` en `backend/migrations/v1.0.0.sql`, **las pantallas de `frontend/src/app/modules/tickets/**`** y los aumentos que necesitaron `users` (la lista de técnicos activos y la lectura de cuentas en bloque) y `settings` (la configuración del reparto). **Está entero**, backend y pantallas |
| `docs/flujos.md` | **as-built**: los recorridos paso a paso de los tickets y sus correos, que es lo que hacen `backend/modules/tickets/**` y sus pantallas |
| `docs/ambientes.md` | El runbook: **el despliegue a producción** —`scripts/prod-build.sh`, hecho y usado en el primer despliegue—, las migraciones, **las copias de seguridad** —`scripts/backup-db.sh`, con las dos bases en el `cron`— y las tres capas de pruebas (`tests/e2e/`) |
| `docs/interfaz-y-experiencia.md` | Cubre `frontend/src/app/core/**` (el armazón, los controles, las pantallas de la sesión y Configuración), `frontend/src/app/shared/components/**` y **la forma de las pantallas de producto**: las de usuarios (3.6), las de tickets (3.7) y el editor de los correos (3.8), todas hechas |
| `docs/modules/settings.md` | Cubre `backend/modules/settings/**`, `backend/shared/version/**`, **las cuatro tablas de configuración de una fila** (`installation_settings`, `directory_settings`, `keycloak_settings` y `ticket_settings`) y `frontend/src/app/core/pages/settings-page.*` (el nombre de la instalación, **cómo se entra y las dos configuraciones de directorio con sus pruebas de conexión**, la marca, el color, el idioma, el prefijo y el reparto). **Está entero** |
| `docs/primer-arranque.md` | **as-built**: la vista de primer arranque, que pide lo necesario para dejar la instalación en marcha en cuatro pasos. Es el **sello** de `installation_settings`, `backend/modules/settings/**` (la API `/api/setup/**`), `frontend/src/app/core/pages/setup-page.*` y el correo saliente del módulo `mail` |
| `docs/modules/ai.md` | **aprobado**: los dos resúmenes del ticket —**motivo** y **última acción**— que redacta el **motor de IA de un contenedor aparte**, en español y en inglés. Cubre `backend/modules/ai/**`, la tabla `ai_insights` y el `ai.yml` del motor. Lo consume `tickets` a través de una interfaz que él mismo declara |
| `docs/modules/mail.md` | Cubre `backend/modules/mail/**`, las plantillas de `backend/migrations/v1.0.0.sql` y **el editor de `frontend/src/app/modules/mail/**`**. **Está entero**, backend y editor |
| `docs/modules/auth.md` | Cubre `backend/modules/auth/**`, el token de sesión y el `state` de OIDC de `shared/auth`, `shared/auth/methods.go`, el middleware y `frontend/src/app/core/**` (sesión, interceptor, guarda y las seis pantallas). **Está entero**: los tres caminos de entrada —local, AD y Keycloak— hechos y verificados, **con la instalación entrando por uno a la vez y la cuenta de fábrica siempre dentro** |
| `docs/modules/users.md` | Cubre `backend/modules/users/**` y la tabla `users` de `backend/migrations/v1.0.0.sql`. **Está entero**, backend y pantallas —las tres pantallas viven en `frontend/src/app/modules/users/**` y su forma la fija `docs/interfaz-y-experiencia.md`, sección 3.6—, incluidas las tres acciones que preguntan al directorio |
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

Esqueleto en pie y verificado (2026-09-22), y **cuatro módulos en marcha** (2026-09-24): los cuatro
contenedores de desarrollo levantan (incluido el buzón de pruebas, y **el directorio de pruebas y
Keycloak** detrás del perfil `auth`), nginx publica el proyecto en `https://dev.catalina-support.example.com` con
TLS, el backend Go conecta con PostgreSQL y responde `GET /api/health` (también a través de nginx), y
el frontend Angular 22 se sirve en desarrollo y pasa sus pruebas.

La migración **`v1.0.0.sql`** está aplicada —`mail_templates` y sus veintidós plantillas, `users`,
`password_tokens`, **las cuatro tablas de configuración de una fila** (la instalación con **el nombre
y el método de entrada**, el directorio, Keycloak y los tickets) y **las diez tablas de los
tickets**—, y en desarrollo se le suma **`v1.0.0_dev.sql`**, los datos de ejemplo, con su guion
(`./scripts/dev-seed.sh`). Hay **cuatro módulos en marcha**:

- **`mail` entero**: tabla, marcadores, renderizado, envío, sus cinco endpoints **y el editor** —los once correos en sus dos idiomas, con vista previa y prueba—.
- **`auth` entero**: los **tres caminos de entrada**, con **la instalación entrando por uno a la
  vez** —el que está puesto en Configuración es el único que se atiende—. El local —correo y
  contraseña, y la cuenta de fábrica—, `me`, salir, olvidar la contraseña, establecerla desde el
  enlace y cambiarla desde dentro, con sus tres correos de cuenta y **las seis pantallas del
  armazón**; el de **AD**, contra un directorio de verdad (OpenLDAP, en `dev.yml` detrás del perfil
  `auth`); y el de **Keycloak**, contra un reino que vive en el repositorio, con el token de vuelta en
  el fragmento de la dirección. En los dos caminos de directorio **nadie da de alta a nadie**: la
  cuenta se crea, se vincula o se pone al día en el primer acceso. **La cuenta de fábrica entra
  siempre**, sea cual sea el método: es la puerta que permite volver a cambiarlo.
- **`tickets` entero, backend y pantallas**: las diez tablas, la numeración sin carreras, el reparto
  por turnos —configurable por tipo—, sus endpoints, los ocho avisos por correo y los permisos de los
  cuatro papeles, **y sus pantallas**: las listas —**Mis tickets**, **Tickets principales** y **Tickets
  internos**, más la bandeja del Administrador—, el alta, el detalle con la **vista doble** y el
  escalado. Son las segundas pantallas de producto.
- **`users` entero, backend y pantallas**: el alta, el reenvío del enlace, la lista con filtros, la
  ficha, los cambios (con el límite de Soporte), el perfil propio, desactivar, reactivar y el cambio
  de origen, **y sus tres pantallas** —la lista, la ficha y el perfil propio—, que además traen las
  **primeras pantallas de producto**.

En desarrollo hay además un **buzón de pruebas** (`mail` en `dev.yml`, Mailpit) al que va todo el
correo: sin él no se podría completar el alta de una cuenta local. Se lee en
`http://127.0.0.1:11004`.

**El armazón de la sesión del frontend está hecho**: se entra desde el navegador con las seis
pantallas (entrada, olvido, establecer contraseña, cambiar contraseña, sin permiso y servidor
caído), con Tailwind v4, los colores en variables de tema y **los textos en dos idiomas que no
pueden quedarse a medias**: si una clave falta en un idioma, no compila.

**El tema funciona**, con **los ocho temas** (dos de fábrica con el color institucional y seis fijos
con su paleta y su acento), **siguiendo al navegador hasta que alguien elige**, con la elección
recordada y cambiando en vivo si el sistema cambia. **El contraste de los ocho está medido**, y hay
una prueba que lo vuelve a medir sobre lo que el navegador pinta. **«Automático» no se muestra ni se
elige**: seguir al sistema es no haber elegido, y mientras nadie elija el conmutador enseña marcado
lo que se está viendo. Los dos temas de fábrica se escriben **una sola vez** con `light-dark()`, y
cada uno de los seis fijos es **un bloque de variables**; el tema se elige con un atributo en `html`. **El tema lo aplica `main.ts`, nunca un script incrustado en el
`index.html`**: la CSP de producción es `script-src 'self'`.

**La marca está puesta**: la instalación tiene **nombre propio** —configurable desde Configuración, y
es lo que se lee en la pantalla de entrada, en el menú lateral y en la pestaña del navegador, en lugar
del texto de fábrica— y **logo**, el que trae la aplicación —**dos, uno por tema**— o el que suba un
administrador, en sus dos huecos claro y oscuro. El **color institucional** se aplica a los dos temas de fábrica. El nombre, el
logo y el color se guardan en la instalación (módulo `settings`), con su endpoint público porque la
entrada se pinta antes de entrar.

**La instalación dice qué versión es**: una constante del backend (`shared/version`, hoy `1.0.0`) que viaja en **la marca pública** y se lee **en la fila de salir del menú lateral, a la derecha, y en el pie de la pantalla de entrada** —decisión del responsable, 2026-09-25, que eligió el pie de la entrada además del menú—.

**La pantalla de Configuración está entera**: **el nombre de la instalación**, **cómo se entra** —el
método de entrada, el directorio de la organización y Keycloak, con sus dos botones de «Probar la
conexión» y **sus secretos guardados y sin salir nunca por la API**—, **el logo** (los dos huecos,
subir y volver al de fábrica), **el color institucional** con vista previa, **el idioma de la
instalación** y **el prefijo y el reparto** con el aviso de que cambiar el prefijo no cambia los
números ya emitidos. **El directorio y Keycloak ya no son variables de entorno**: viven en la base y
se cambian desde aquí, así que configurar una instalación no pide editar un archivo en el servidor ni
reiniciar nada.

**La instalación tiene región horaria y dirección pública** (2026-09-29): en **Configuración** se
eligen **la zona horaria** —de una lista de zonas con buscador, y se ve la hora que es en ella y su
desfase— y **la dirección pública** —esquema, host y puerto, `localhost` incluido—. **La zona decide
cómo se leen todas las fechas**, en la interfaz y en los correos, y **las guardadas siguen en UTC**:
cambiarla **no mueve ningún ticket**. **La dirección es la base de los enlaces que salen en los
correos** (alta, restablecer) **y de la vuelta de Keycloak**; antes era la variable `PUBLIC_APP_URL`,
que queda como respaldo. **Si la dirección no es https, Configuración lo avisa** —la sesión y la
contraseña viajan sin cifrar— sin bloquear nada, que es lo que permite probar la instalación en local
y usarla en serio (`docs/modules/settings.md`, decisiones 13 a 15). Va **en la propia migración `v1.0.0.sql`**: la versión sigue abierta.

**La instalación tiene vista de primer arranque** (2026-09-30): en una instalación **sin sellar** la
aplicación lleva a **`/setup`**, que pide en cuatro pasos **la instalación** (nombre e idioma), **cómo
se entra** (con su prueba), **dónde está** (región y dirección) y **el correo saliente** (que vive en la base:
las variables `SMTP_*` del entorno se retiraron), y al terminar **sella** la instalación: la vista no vuelve y su API
contesta **409** (`docs/primer-arranque.md`). El sello vive en `installation_settings.installed_at`, y
**una instalación que ya estaba configurada queda sellada al actualizar**, así que nadie ve el
asistente en producción por este cambio. **La contraseña de la cuenta de fábrica sigue en el entorno**
(`ADMIN_PASSWORD`): el asistente no la pide.

**El `README.md` presenta el proyecto a quien llega de fuera**: qué es, que es **software libre (MIT)**,
cómo **participar**, y **cómo levantarlo** —en local con `dev.yml` y en serio con `prod-build.sh`, la
migración, nginx y lo que se configura desde la propia interfaz—. Queda **un hueco marcado** para las
donaciones y la invitación a participar: el responsable lo rellenará con los datos que quiera publicar.

**El armazón con el menú lateral está hecho**, con sus tres zonas: el producto y su logo arriba, las
opciones del papel en medio y los controles abajo (el nombre, el papel, el tema, el idioma y salir).
En PC se pliega y lo recuerda; en móvil es un cajón. **El menú enseña sólo lo que existe**, y lo que
le toca a cada papel: un usuario no ve Configuración, y si escribe la dirección a mano el backend se
lo rechaza.

**El menú ya es el del documento**, y **cambió el 2026-09-26** al repartir las listas: **Mis tickets**
(usuario), **Mis tickets · Tickets principales · Tickets internos · Usuarios** (Soporte), **Mis tickets
· Tickets principales · Tickets internos** (Desarrollo) y **Bandeja · Usuarios · Configuración**
(Administrador). **«Nuevo ticket» salió del menú** —se confundía con un módulo— y vive en el botón de la
bandeja, que es donde se crea. El **inicio provisional desapareció** con las pantallas de tickets: ya no
le faltaba pantalla a nadie, y la raíz es ahora un reparto —el Administrador a los usuarios, los demás a
su bandeja—.

**Cada ticket se clasifica con una categoría y, si hace falta, con etiquetas** (2026-09-27): la
**categoría** viene de un catálogo —«Red», «Software», «Licencias»…— y **es obligatoria: no hay ticket
sin categoría**, lo garantiza la propia base, no la pantalla—, y las **etiquetas** son libres,
opcionales, **en minúsculas y con guion medio** (`red-wifi`) y con sugerencias de las que ya existen.
**La elige quien abre el ticket y Soporte la puede corregir**; el catálogo lo mantienen **Soporte y el
Administrador**, y **retirar una categoría es sólo del Administrador** y es **desactivarla** —deja de
ofrecerse y los tickets que la tienen la conservan—. **La categoría y las etiquetas son del principal**
y el interno las hereda, y **las dos filtran y buscan** (`docs/modules/tickets.md`, decisiones 64 a 68).

**Dentro del ticket, todo se edita donde se muestra** (2026-09-28): la **categoría** con su desplegable en
su propia línea, las **etiquetas** con su aspa y su campo para añadir, los **observadores** con su
«Añadir» y su buscador —además de llegar por una mención—, y el **asunto** y la **descripción** con su
lápiz, que los convierte en el sitio en un campo con «Guardar» y «Cancelar». **No hay formularios
aparte**: eran los que hacían que no se encontrara. Y en el editor se **etiqueta escribiendo `@`**, con
la lista filtrada por **nombre o correo** mientras se escribe (`docs/modules/tickets.md`, decisión 74).

**Los resúmenes del ticket los redacta un motor de IA propio** (2026-09-27): **«Motivo»** —de qué va el
ticket— y **«Última acción»** —qué fue lo último que pasó—, los dos **en español y en inglés**, en la
lista y en la ficha. El motor vive **en su propio contenedor** (`ai.yml`, llama.cpp con un modelo de
1,5B, compartido por los dos entornos por la red `catalina-support-ai`), **nada del ticket sale del
servidor** y **es opcional a propósito**: si no está, los campos se quedan sin texto y la mesa de ayuda
funciona entera. Se piden **en segundo plano**, con reintentos, y se recalculan con cada movimiento del
ticket (`docs/modules/ai.md`).

**Se puede llamar a otro técnico o desarrollador y quedarse siguiendo el ticket** (2026-09-27): en un
comentario, el botón **«Etiquetar»** (o escribir `@`) mete el nombre **dentro del texto, resaltado**, y
quien lo lee ve a quién se está llamando. Esa persona pasa a ser **observadora** del ticket, que **no
es lo mismo que el responsable**: el ticket tiene **como mucho un responsable** —que sigue siendo
opcional— y **muchos observadores**. **Observar no da permisos**: sólo hace que el ticket esté en tu
bandeja, en el chip **«Observo»** de «Mis tickets», y que recibas **los mismos correos de personal que
el responsable**, además del aviso de que te han etiquetado. Etiquetan **Soporte y Desarrollo**;
**quitar a un observador lo puede hacer cualquiera de los dos**, queda en el historial y **manda sobre
lo ya escrito** —la mención se queda en el comentario, pero no vuelve a añadirlo— (`docs/modules/tickets.md`,
decisiones 58 a 63).

**Copiar el número de un ticket es un botón de sólo icono**, y al pulsarlo **se cambia por un visto
verde dos segundos** (decisión del responsable, 2026-09-27). El botón del inventario estrena con eso
**una cuarta forma** —la verde de la confirmación— y la posibilidad de ser **sólo un icono**, con su
nombre accesible obligatorio.

**Los adjuntos son de todo lo que se puede mandar**: imágenes, vídeo, PDF, Office, **texto y código** —`sql`, `json`, `log`, `sh`, `py`…, que entraron el 2026-09-26 al no poder mandar un `.sql`— y comprimidos, con **una lista cerrada de extensiones** (decisión 54). **La imagen y el vídeo se ven con un tope de 480 × 360** dentro de la descripción y de los comentarios, y **al pulsarlos se abren en el visor grande**, que es donde se ven enteros y se descargan; **el vídeo se reproduce ahí** (decisión 55).

**Los comentarios llevan sus adjuntos**, con **tres puertas y una sola regla**: **pegar**, arrastrar y soltar, y el botón de elegir el archivo (decisión del responsable, 2026-09-25). Un comentario **puede ser sólo un archivo** —y entonces el cuerpo va vacío, que es lo que la decisión cambió en la tabla— y **si una subida falla después de publicarlo**, el comentario se queda, se dice qué archivo falta y el botón lo sube al mismo comentario sin volver a escribirlo.

**Las pantallas de tickets están hechas**: las **listas** —**Mis tickets** (lo asignado, lo abierto por
uno y donde ha comentado, con chip de tipo para Soporte y Desarrollo), **Tickets principales** y
**Tickets internos** para los dos equipos, y la bandeja del Administrador con los dos tipos de sólo
lectura—, todas la misma pantalla con distinto filtro, en tabla en PC y tarjetas en móvil, con chips de
estado y búsqueda; el **alta** con arrastrar y soltar, y el **detalle**, que es la pieza central: una sola línea de tiempo con los comentarios y lo que hizo el
sistema, la ficha con las acciones a la derecha en PC y debajo en móvil, **la vista doble** cuando hay
interno, y los adjuntos **con vista previa** en imágenes y PDF. **La interfaz no ofrece lo que no se
puede hacer**: el usuario no ve el interno, Desarrollo no escribe en el principal y el Administrador
mira sin botones.

**Configuración tiene ya el prefijo, el reparto y el idioma de la instalación**, con el aviso de que
cambiar el prefijo **no cambia los números ya emitidos**, y el enlace al **editor de los correos**,
que es una pantalla del módulo `mail` a propósito: llamar a `/api/mail/**` desde `settings` rompería
la regla dura de modularidad.

**El editor de los correos está hecho**: los once correos con sus dos idiomas al lado, el cuerpo con
sus botones de formato sobre lo seleccionado, los marcadores que se insertan al pulsarlos y una **vista
previa que renderiza el backend**, para que lo que se ve sea lo que va a salir.

**Y el idioma de la instalación es el que se le pone a una cuenta nueva** cuando quien la da de alta no
elige otro.

**Los dos caminos de directorio están hechos y verificados**: el de AD contra el directorio de pruebas
(OpenLDAP) y el de Keycloak contra un reino que vive en el repositorio, los dos detrás del perfil
`auth` de `dev.yml`. Una persona con cuenta allí **entra directamente, sin que nadie le dé de alta
nada** —es la corrección del responsable del 2026-09-25—, y su cuenta se crea, se vincula o se pone al
día sola con lo que dice el directorio. **Con el método en AD la pantalla de entrada sigue siendo el
mismo formulario** —lo que cambia es quién contesta, y se dice— y **con el de Keycloak no hay
formulario**: un botón que lleva allí, y una vuelta con el token **en el fragmento** de la dirección,
que la pantalla lee, guarda y borra. **El precio de «un método a la vez» está dicho y aceptado**: con
el método en `ad` o en `keycloak`, las cuentas locales —Soporte y Desarrollo incluidas— no entran
hasta que el método vuelva a `local`.

**Y las tres acciones que dependían del directorio también están hechas**: reactivar una cuenta de AD
pregunta al directorio si sigue conociendo a esa persona, la de Keycloak no se reactiva a mano —se
cuenta que vuelve sola al entrar, porque preguntarlo pediría una cuenta de servicio en el reino—, y el
alta y el cambio de origen hacia el directorio no se hacen a mano.

**Las copias de la base están hechas**: `scripts/backup-db.sh` vuelca la base con `pg_dump -Fc`,
comprueba que el volcado se puede leer, guarda **catorce días** y está puesto en el `cron` de esta
máquina. **No copia `_files`**, y es a propósito: los adjuntos no se borran nunca y copiarlos es cosa
de soporte manual (decisión del responsable, 2026-09-25).

**El despliegue a producción está hecho y verificado por dentro** (2026-09-25): `scripts/prod-build.sh` construye y publica los artefactos en `/srv/catalina-support` y levanta los tres contenedores de `prod.yml`, con el esquema de `v1.0.0.sql` aplicado **una sola vez** en la base de producción y **la cuenta de fábrica entrando** con la contraseña de `config/env/prod.env` —que no se versiona y **no aparece en la documentación**—. La base de producción **se copia a diario** desde el mismo día, con su línea en el `cron`.

**Lo que falta**: **el correo saliente de producción**, y con él **el alta de Soporte y Desarrollo** —sin correo no llega el enlace para establecer la contraseña, así que hoy sólo entra la cuenta de fábrica—; **abrir el dominio**, que sigue respondiendo 503 a propósito porque **la 1.0.0 no está cerrada**; y **la etiqueta `v1.0.0`**, que el responsable no quiere crear todavía porque puede haber cambios sobre la versión.

**Las tres capas de pruebas están montadas**, incluidas las de interfaz: `docker compose -f dev.yml
run --rm e2e` ejecuta Playwright **en un contenedor** (no en la máquina) contra el entorno de
desarrollo, en PC y en móvil, con **212 casos** (194 en verde y 18 que se saltan: los de un
dispositivo o de las herramientas y siete de los caminos de directorio, que se prueban una sola vez
porque no dependen del ancho). Leen los correos del buzón de pruebas, así que prueban el enlace de
verdad y no una suposición. **Esa capa ya ha encontrado nueve fallos que las pruebas de unidad no
veían** —los dos últimos, que la cabecera del detalle le enseñaba al usuario el estado interno
(«Nuevo» donde su pantalla dice «Recibido») y que **la imagen con tope de 480 px se salía de la
pantalla en un móvil de 412**, porque `max-w-[480px]` limita pero no encoge—, y los nueve están
corregidos con su prueba.

El detalle de lo que existe, lo que está sin verificar y lo que falta está en
`docs/arquitectura.md`, sección 13. En resumen:

- `docs/arquitectura.md` (as-built): stack, contenedores, dominios y regla de modularidad.
- `docs/propósito-y-alcance.md` (**as-built**, **3 enmiendas**): producto, los dos equipos y el modelo de tickets.
- `docs/usuarios-y-permisos.md` (**as-built**, **5 enmiendas**): permisos, acceso —**un método a la vez**— y cuentas.
- `docs/modules/tickets.md` (**as-built**, **6 enmiendas**): modelo de datos, transiciones, endpoints y pantallas. **Terminado**, backend y pantallas, con **los adjuntos dentro del texto**.
- `docs/flujos.md` (**as-built**, **1 enmienda**): los recorridos paso a paso y sus correos.
- `docs/ambientes.md` (**as-built**, **3 enmiendas**): despliegue, migraciones, **copias de seguridad** y pruebas. **El despliegue está hecho y verificado por dentro**; falta el correo de producción y abrir el dominio.
- `docs/interfaz-y-experiencia.md` (**as-built**, **14 enmiendas**): la parte visual y de experiencia, con **la tarjeta de cómo se entra**, **la versión del sistema** y **el editor con los adjuntos dentro**. **Las pantallas de usuarios (3.6), de tickets (3.7) y el editor de los correos (3.8) están hechos**.
- `docs/modules/mail.md` (**as-built**, **3 enmiendas**): el módulo de correo. **Terminado**, backend y editor.
- `docs/modules/settings.md` (**as-built**, **4 enmiendas**): la configuración de la instalación y la marca. **Terminado**: la pantalla de Configuración está entera, con el nombre de la instalación, con cómo se entra y con la versión del sistema.
- `docs/modules/auth.md` (**as-built**, **6 enmiendas**): el módulo de autenticación. **Terminado**: los tres caminos de entrada, con **un método a la vez**.
- `docs/modules/users.md` (**as-built**, **5 enmiendas**): el módulo de usuarios. **Terminado**, backend y pantallas, incluidas **las tres acciones que dependían del directorio**.
- `docs/README.md`: índice de documentación.

- `docs/primer-arranque.md` (**as-built**): la vista de primer arranque y su sello.
- `docs/modules/ai.md` (**aprobado**): el motor de IA y los dos campos que redacta —«Motivo» y «Última acción»—, en español y en inglés. **Es el sexto módulo**, y el único sin pantalla propia.

La cadena de producto **está completa**. Además, **cada módulo tiene su documento**, escrito justo
antes de implementarlo. El orden es **`mail` → `auth` → `users` → `tickets`**, porque `auth` no
puede mandar el correo de alta sin el módulo de correo.

## Estructura de Carpetas

```text
.
├── dev.yml                 # compose de desarrollo (frontend, backend, database, mail; y el
│                           # directorio de pruebas y las de interfaz detrás de un perfil)
├── prod.yml                # compose de producción (sin verificar todavía)
├── ai.yml                  # el motor de IA: un contenedor, compartido por los dos entornos
│                           # (el modelo vive en un volumen, no en el repositorio)
├── backend                 # Go: main.go, .air.toml, shared/, modules/ y migrations/
├── frontend                # Angular 22 + Tailwind v4: src/app/{core,shared,modules}, y public/ con el logo
├── config
│   ├── dockerfiles         # una imagen por servicio y por entorno
│   ├── env                 # dev.env y prod.env.example (versionados); prod.env no se versiona
│   ├── keycloak            # el reino de pruebas: el cliente y sus personas
│   ├── ldap                # el directorio de pruebas: las personas del LDIF
│   ├── seed                # los archivos de los adjuntos del seeder de desarrollo
│   └── nginx               # vhost de desarrollo, vhost de producción y nginx del contenedor
├── _logs                   # archivos de go-logs (la carpeta se versiona, los .log no)
├── _files                  # adjuntos, por año y número de ticket (se versiona la carpeta,
│                           # no los archivos)
├── scripts                 # lo que se ejecuta en el servidor: el despliegue y las copias de la base
├── tests
│   └── e2e                 # Playwright: los recorridos de interfaz contra desarrollo
├── docs                    # documentos de producto
│   └── modules             # un documento por módulo (auth.md, users.md)
├── AGENTS.md
├── README.md
└── LICENSE
```

No crear carpetas de código "por si acaso": cada una nace con el documento que la justifica.

## Comandos

Todo se ejecuta **dentro de los contenedores**: la máquina no tiene Go ni Node instalados (y no
debe tenerlos). `docs/arquitectura.md`, sección 10.

```bash
# Los servicios de desarrollo (frontend, backend, database y el buzón de pruebas).
# **Un solo comando de Docker, igual en Linux, macOS y Windows**: la red compartida con el motor de IA
# (`catalina-support-ai`) la crea el primero que arranca, así que no hay que crear nada a mano. Para el
# motor, que es opcional: `docker compose -f ai.yml up -d`
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

# El buzón de pruebas: todo el correo de desarrollo, con su HTML y su versión de texto
open http://127.0.0.1:11004

# Interfaz (Playwright, en un contenedor y contra desarrollo: PC y móvil).
# **Después de cada pasada, reiniciar el entorno**: la suite apaga sus cuentas al terminar, pero los
# tickets que crea no se borran —un ticket no se borra—, y el seeder es lo que deja desarrollo con las
# once cuentas y los 25 tickets de ejemplo (decisión del responsable, 2026-09-28). **Se avisa antes**:
# el seeder borra también lo que haya a mano en desarrollo
docker compose -f dev.yml run --rm e2e
./scripts/dev-seed.sh
docker compose -f dev.yml --profile auth run --rm e2e      # con los dos caminos de directorio
docker compose -f dev.yml run --rm -e DIAGNOSTICO=1 e2e    # volcado de lo que hay en pantalla

# El motor de IA: su propio compose, compartido por desarrollo y producción (docs/modules/ai.md).
# La primera vez descarga el modelo (~1,1 GB) al volumen; sin él, la aplicación funciona igual y los
# dos campos que redacta se quedan sin texto
docker compose -f ai.yml up -d
docker compose -f ai.yml ps
docker compose -f ai.yml logs -f ai

# Los dos servicios de pruebas de los caminos de directorio (OpenLDAP y Keycloak), con sus personas
docker compose -f dev.yml --profile auth up -d

# Los datos de ejemplo de desarrollo: once cuentas, 25 tickets con su historia y sus adjuntos.
# Borra los tickets que hubiera y deja el entorno en un estado conocido (docs/ambientes.md, sección 3.3)
./scripts/dev-seed.sh

# El despliegue a producción: construye y publica los artefactos en /srv/catalina-support y
# levanta los contenedores de prod.yml (docs/ambientes.md, sección 4.2). **No abre el dominio**: eso
# es el vhost de nginx, y hoy sigue respondiendo 503
./scripts/prod-build.sh                 # construir, publicar y desplegar
./scripts/prod-build.sh --no-deploy     # sólo construir y publicar
./scripts/prod-build.sh --only backend  # un solo componente

# La copia de la base de datos (se ejecuta en la máquina, no dentro de un contenedor: es lo que va
# en el cron del servidor). Deja el volcado en /srv/catalina-support/backups/ y borra lo que pasa
# de 14 días (docs/ambientes.md, sección 6)
./scripts/backup-db.sh            # el entorno de desarrollo
./scripts/backup-db.sh prod       # producción
```

En desarrollo se entra por **https://dev.catalina-support.example.com**, que es nginx (en la
máquina) delante de los contenedores. Los puertos 11001 y 11002 siguen publicados en
`127.0.0.1` para depurar sin pasar por nginx.

Los dos vhosts son **copias** de `config/nginx/` en `/etc/nginx/conf.d/` (no enlaces), igual que
en Calibyou. Si se cambia un vhost en el repositorio, hay que volver a copiarlo y recargar:

```bash
sudo cp config/nginx/catalina-support-dev.conf  /etc/nginx/conf.d/
sudo nginx -t && sudo systemctl reload nginx
```

El vhost de producción (`https://support.example.com`) ya está instalado y con
certificado, y **el despliegue a producción ya está hecho** (2026-09-25): los contenedores de
`prod.yml` están levantados y verificados **por dentro** —salud, esquema aplicado y la cuenta de
fábrica entrando—. **El dominio sigue respondiendo 503** a propósito, porque **la 1.0.0 no está
cerrada**: abrirlo es sustituir el bloque del aviso por el `proxy_pass` al frontend —que está
comentado justo debajo, en `config/nginx/catalina-support-prod.conf`—, copiar el vhost y recargar
nginx. Mientras el dominio está cerrado responde 503 con un aviso, y no un 502 de nginx, que
parecería una web rota.

## Reglas para Agentes

- **Regla 0 primero**: documento aprobado antes de escribir código. Sin aprobación no hay código.
- No inventar decisiones de producto, stack ni estructura. Si falta un dato, se pregunta o se
  deja explícitamente como pendiente en el documento.
- La documentación y la comunicación con el responsable del proyecto se escriben en **español**.
- Un documento por área, con su cabecera de estado: **de producto en `docs/`, de módulo en
  `docs/modules/`** (y con el nombre del módulo: `modules/auth.md`). Nada de documentos de diseño
  sueltos en la raíz.
- Mantener `docs/README.md` (índice) y la tabla de correspondencia de este archivo al día en el
  mismo cambio que cree, renombre o jubile un documento.
- **Nada se ejecuta en la máquina**: los comandos van por `docker compose -f dev.yml exec …` (o
  `run --rm` para lo que se pide a propósito, como las pruebas de interfaz). No instalar Go ni Node
  en la máquina ni añadir pasos que los necesiten.
- **Un cambio que toca la interfaz se prueba en un navegador antes de darlo por terminado**
  (`docker compose -f dev.yml run --rm e2e`), y **mirando el aspecto, no sólo el texto**: que la hoja
  de estilos llegue con sus utilidades y que los colores y bordes se apliquen. Una pantalla sin
  estilos pasa todas las pruebas de estructura. Si el cambio afecta a un recorrido que no tiene caso,
  **se añade el caso en el mismo trabajo** (`docs/ambientes.md`, sección 9.3).
- **Modularidad (regla dura)**: un módulo del frontend sólo llama a `/api/<su nombre>/**`; un
  módulo del backend sólo escribe en sus tablas; `shared` no importa nunca de `modules`. Las
  llamadas del frontend son siempre **relativas**, nunca a un host (`docs/arquitectura.md`, sección 4).
- **Logs**: sólo `github.com/juanky-estevez/go-logs` (`logs.LogInfo/LogSuccess/LogWarning/LogError`).
  Nada de `fmt.Println` ni de `log`/`slog` en el código de la aplicación. **Nunca** se registran
  contraseñas —ni la correcta ni la equivocada—, tokens de sesión, tokens de enlace ni el contenido
  de un correo. **La dirección de correo del destinatario sí se registra** (decisión del responsable,
  2026-09-23): es lo que permite responder a «no me ha llegado nada», y sin ella un envío fallido no
  se puede investigar. Fuera de eso, nada de datos personales: para referirse a una cuenta se usa su
  identificador, no su nombre.
- **Base de datos**: el esquema vive en `backend/migrations/`, **un archivo por versión**
  (`v1.0.0.sql`, `v1.1.0.sql`…) **más un archivo de datos de ejemplo por versión con el sufijo
  `_dev`** (`v1.0.0_dev.sql`), transaccionales e idempotentes, aplicados con
  `psql -v ON_ERROR_STOP=1` en orden hasta la versión publicada. **El `_dev` no se aplica nunca en
  producción**, y lo que la aplicación necesita para funcionar —las filas de configuración y las
  plantillas de correo— va en el archivo de la versión, no en el de ejemplos. Nunca `AutoMigrate` ni
  un `ALTER` a mano (`docs/ambientes.md`, sección 5).
- **Cuenta de fábrica**: el usuario `admin` (papel `administrador`) **no está en la base de datos**:
  vive en la configuración y su contraseña es `ADMIN_PASSWORD`, obligatoria en producción y
  comprobada en cada entrada. Es la única cuenta sin correo, la única que no sigue la política de
  contraseñas y la única que no aparece en los tickets. El valor de producción vive sólo en
  `config/env/prod.env` y **nunca se escribe en la documentación ni en el código**
  (`docs/usuarios-y-permisos.md`, sección 8).
- **Angular moderno**: componentes standalone sin `NgModule`, señales, `@if`/`@for`, rutas
  perezosas por módulo y un único servicio por módulo que centraliza las llamadas HTTP. Los
  componentes no usan `HttpClient` directamente.
- **Las rutas del frontend van en inglés** (`/login`, `/set-password`), en minúsculas y con guiones,
  aunque el idioma del producto sea español (`docs/modules/auth.md`, decisión 21).
- **Los textos de la interfaz no se escriben dentro de los componentes**: viven en
  `core/i18n/es.ts` y `core/i18n/en.ts`, con el mismo tipo compartido, así que **una clave que falte
  en un idioma no compila**. Los textos de las claves de error van en el mismo diccionario: una
  pantalla nunca enseña una clave (`docs/interfaz-y-experiencia.md`, sección 8).
- **Los colores salen del tema** (variables CSS de `src/styles.css`, expuestas a Tailwind con
  `@theme inline`). Ningún componente escribe un color a mano, para que los ocho temas sean posibles
  sin tocar componentes. **Un tema es un bloque de variables** (los dos de fábrica, con
  `light-dark(claro, oscuro)`; los seis fijos, con sus valores), y el tema se elige con un atributo en
  `html` (`data-theme`, o ninguno para seguir al sistema). **Cada
  tema nuevo se mide antes de entrar**: texto, texto apagado, acento, el texto del botón, el borde del
  campo y los tres colores de estado; **un tema que no se lee no entra**, y hay una prueba que lo
  vuelve a medir sobre lo que el navegador pinta. **Nunca se aplica un tema con un `script`
  incrustado en el `index.html`**: la CSP de producción es `script-src 'self'` y no lo permitiría
  (`ThemeService`, `main.ts`).
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
  `escalado`. **El control de estados del ticket sí enseña estados** (decisión del responsable,
  2026-09-29, decisión 80 de `docs/modules/tickets.md`): se corrigió la regla anterior —«las acciones se
  llaman por lo que hacen y no por el estado al que llevan»— porque el desplegable del ticket es
  **para cambiar de estado**, y **lo que se lee se adapta a quien mira**: el usuario ve «Recibido»,
  «En curso» y «Esperamos tu respuesta», y Soporte y Desarrollo ven los estados de verdad. Lo que no
  cambia es que **la ventana de cada cambio explica qué va a pasar**.
- Antes de editar, revisar los patrones que ya existen en el repositorio.
- Preferir `rg` para búsquedas.
