# Interfaz y experiencia

> **Estado:** as-built
> **Última actualización:** 2026-10-08
>
> **Enmienda implementada y verificada el 2026-10-08: ocultar la ayuda sin IA configurada.** La
> sección 20 hace que el botón y su modal dependan de la capacidad recibida con el detalle del
> ticket. El responsable aprobó 1A–3A y confirmó mantener la IA obligatoria.
>
> **Corrección aprobada e implementada el 2026-10-08.** No existe un recorrido real de ticket sin IA configurada:
> la obligatoriedad vigente lleva antes a Configuración. El responsable decidió conservarla. El
> estado falso se probará en el componente y el recorrido de navegador verificará el estado
> configurado. Las pruebas finales respetan esta frontera.
>
> **Enmienda implementada y verificada el 2026-10-08: modal «Mejorar con IA».** La sección 19
> describe el botón con icono, el editor sencillo, las cinco tonalidades, la revisión y sus estados.
> Pasó pruebas unitarias y el recorrido real en PC y móvil, en el entorno desechable.
>
> **Enmienda implementada y verificada el 2026-10-08 (bloque 6).** El único fleco visual registrado en
> `docs/arquitectura.md` era un repaso de formato sin alcance ni criterio de cierre. La sección 18
> define la revisión completa elegida por el responsable en 1A–14C. Lo existente continúa
> as-built; el responsable aprobó explícitamente esta enmienda el 2026-10-08.
>
> **Hallazgo corregido y verificado el 2026-10-08.** El bloque 5 encontró cifras de RAM distintas entre
> los selectores y el administrador, y un error genérico cuando un modelo no cabe. La sección 17
> alinea las cifras y define el toast elegido en 17A y 22A. La corrección fue aprobada explícitamente
> el 2026-10-08.
>
> **Enmienda propuesta el 2026-10-06.** Lo existente continúa as-built. La sección 16 define el
> idioma global como primer control y la experiencia para configurar, descargar y activar la IA
> obligatoria. El repaso quedó cerrado, la propuesta fue aprobada y la implementación fue verificada.
>
> **Enmendado el 2026-10-06**, con aprobación explícita del responsable. Los resultados de guardar y
> probar conexiones en `/settings` y `/setup` se mostraban en un aviso al principio del documento. En una pantalla
> larga, el resultado queda fuera de la parte visible y parece que la acción no respondió. El
> responsable confirmó un toast compartido, fijo en la ventana, para esos resultados. Los avisos que
> explican una condición de la pantalla permanecen dentro de su sección. Se implementó y verificó en
> PC y móvil.
>
> **Enmendado el 2026-10-05**, con aprobación explícita del responsable. Configuración tenía el mismo
> problema visual que tuvo el paso 3 de `/setup`: una lista larga puede ocultar la opción resaltada.
> El responsable confirmó aplicar en `/settings` el mismo campo de sólo lectura antes del buscador,
> con el identificador IANA exacto, actualización inmediata y la hora y el desfase debajo. Se
> implementó y se verificó en PC y móvil.
>
> **Enmendado el 2026-10-05**, con aprobación explícita del responsable. Al revisar todas las vistas se
> encontró que los controles de navegación de cabecera ya están a la izquierda en el alta y el
> detalle de tickets y en la ficha de usuario, pero `/mail` coloca a la derecha su vuelta a
> Configuración aunque este mismo documento dice desde el 2026-09-29 que debe ir a la izquierda.
> El responsable confirmó que se unifiquen los controles de cabecera como icono antes del título y
> que los enlaces de flujo dentro del contenido conserven su presentación. Se corrigió `/mail` y se
> verificó la posición y el regreso a Configuración en PC y móvil.
>
> **Enmendado el 2026-10-05**, conforme a `docs/primer-arranque.md` aprobado: el paso 3 de `/setup`
> muestra antes del buscador un campo de sólo lectura con la zona seleccionada. Enseña el
> identificador IANA exacto, se actualiza en el acto al elegir en la lista y permite seleccionar y
> copiar el texto; la hora y el desfase continúan debajo. Se verificó en PC y móvil.
>
> **Enmendado el 2026-10-04**, conforme a `docs/prueba-local.md` aprobado: `/setup` es la excepción
> inicial al idioma del navegador. Una instalación nueva se presenta entera en inglés; su selector
> cambia todos los textos, `html[lang]` y el valor recordado al instante. Al reanudar manda el idioma
> ya guardado en la instalación.
>
> **Enmendado el 2026-10-03**, conforme a `docs/prueba-local.md` aprobado: Los formularios de
> Keycloak en Configuración y primer arranque incorporan la dirección interna opcional, separada del
> emisor público, con ayuda ES/EN. El selector de Configuración permite elegir métodos todavía
> vacíos para mostrar sus campos; la selección es un borrador hasta guardar con éxito (enmienda
> §5.1.b aprobada).
>
> **Enmendado el 2026-09-30 (novena vez)**: **las tarjetas de Configuración se ordenaron y cada una
> lleva lo que su título dice**. Estaban descolocadas desde el cambio anterior: «La numeración y el
> reparto» tenía dentro sólo el idioma, y el prefijo y el reparto vivían en «Región horaria y dirección
> pública». Ahora son seis, en este orden: **La instalación** (nombre e idioma) · **La marca** (logo y
> color) · **Método de autenticación** · **La numeración y el reparto** · **Región horaria y dirección
> pública** · **Los correos** (`docs/modules/settings.md`, sección 5.11).
>
> **Enmendado el 2026-09-29 (octava vez)**: en **Configuración** entran **la región de la instalación** y
> **la dirección pública** (decisiones 13 a 15 de `docs/modules/settings.md`). La región se elige de una
> **lista de zonas con buscador** y la pantalla enseña **la hora que es en ella y su desfase**; la
> dirección se escribe con su esquema, su host y —si hace falta— su puerto, con **`localhost` permitido**
> y **un aviso, sin bloquear nada, cuando no es https**.
>
> **Enmendado el 2026-09-29 (séptima vez)**, a petición del responsable, al usar el producto: (1) el
> **número del ticket manda** —más grande, y crece en pantallas anchas— y **se retira el cartel «El ticket
> interno»**; (2) los **botones de sólo icono se achican** (de 44 px a 36, que sigue por encima de lo que
> se puede pulsar); (3) en **Configuración**, «Cómo se entra» pasa a llamarse **«Método de
> autenticación»**; (4) el botón **«Configuración» de la pantalla de correos** se cambia por **el botón de
> atrás** —la flecha, a la izquierda—, igual que en las listas de tickets; y (5) en **Usuarios**, el botón
> de mandar el enlace pasa a llamarse **«Cambiar contraseña»** y **sólo se enseña en las cuentas locales**
> (decisión 90 y `docs/usuarios-y-permisos.md`).
>
> **Enmendado el 2026-09-29 (sexta vez)**, a petición del responsable, al usar el producto: (1) en la
> cabecera del ticket, **el número de tu equipo va primero y con más peso** —Desarrollo ve primero el
> **interno**; Soporte y el Administrador, el **principal**— y **se retiran los enlaces «Ver el ticket
> interno» y «Ver el principal»**, porque para eso está el conmutador (decisión 87); (2) **«En espera»
> sólo se confirma**, sin pedir texto, y **cerrar sigue exigiendo el comentario** (decisión 88); y (3) la
> **lista de a quién etiquetar sale junto al cursor**, donde se está escribiendo, en vez de encima del
> bloque del comentario (decisión 89).
>
> **Enmendado el 2026-09-29 (quinta vez)**, a petición del responsable, al usar el producto: (1) en las
> tablas de resumen, **la columna «Ticket» enseña sólo el número** y **la categoría y las etiquetas se
> agrupan en una columna nueva, «Clasificación»** (decisión 86); (2) **el responsable se reasigna con un
> botón junto a su nombre**, que abre un modal —el campo «Asignar a» suelto desaparece—; (3) **los
> botones de volver van a la izquierda y son sólo un icono** (una flecha del proyecto, cuadrada); (4) el
> **área de trabajo es más ancha** en pantallas grandes —`px-5`, o `px-10` a partir de tablet—, sin llegar
> a ocupar todo el ancho; y (5) **los dos resúmenes del motor se refrescan solos** mientras están
> pendientes, después de **cualquier** cambio (antes sólo tras pulsar «Volver a resumir», y el campo se
> quedaba en «Generando…» hasta recargar: lo reportó el responsable como un cuelgue). **La mención, de
> paso, se ve como una mención**: con su `@` delante y con un color del tema que la distingue.
>
> **Enmendado el 2026-09-29 (cuarta vez)**, al aclararse quién etiqueta (decisión 85): el **modal de
> etiquetas** lo tienen **los dos equipos**, cada uno desde el ticket en el que trabaja —Soporte en el
> principal, Desarrollo en el interno—, y **las etiquetas son del principal** (el interno las hereda):
> cambiarlas desde cualquiera de los dos cambia las mismas. Es lo que quiere decir «clasificar es de los
> dos equipos», y **no** toca la regla de que Desarrollo no escribe el texto del principal.
>
> **Enmendado el 2026-09-29 (tercera vez)**, al comprobarse que dos decisiones se contradecían: **el
> catálogo se cura** (decisión 84). Escribir una etiqueta nueva en un ticket **ya no la crea**: quien
> etiqueta **elige de las que hay** —el modal de la ficha sólo ofrece las del catálogo— y **los nombres
> nuevos los estrena el Administrador** en la pantalla del catálogo. A quien no lo es, poner una
> etiqueta que no existe le responde **422** con su aviso, en vez de estrenarla sin querer.
>
> **Enmendado el 2026-09-29 (segunda vez)**, a petición del responsable: **el usuario no pone etiquetas**
> (decisión 82) —**el alta no lleva ese campo** y **la línea de etiquetas de la ficha no se le enseña**—,
> porque etiquetar es de **Soporte y Desarrollo**; y **se etiqueta en un modal**: **todas las etiquetas
> del catálogo** con cuántos tickets lleva cada una, un **buscador**, **casillas** para marcar y
> desmarcar varias a la vez —las del ticket salen marcadas— y **«Guardar» al final** (decisión 82). El
> **catálogo lo mantiene sólo el Administrador** (decisión 83), así que en la pantalla del catálogo los
> botones de crear, renombrar y retirar **no se le enseñan a Soporte**. Y **salir de la aplicación se
> confirma**: el botón de salir del menú abre su ventana antes de cerrar la sesión.
>
> **Enmendado el 2026-09-29**, **corrección del responsable** sobre la decisión 78: el desplegable del
> ticket **no es una lista de acciones, es para cambiar de estado**, y enseña **la lista de estados** —la
> misma en los principales y en los internos—, con **el estado actual como primera opción**, marcado y
> sin poder elegirlo, y **sólo los estados a los que el ticket puede pasar ahora** (decisión 80). Elegir
> uno abre **una sola ventana** que dice a cuál se pasa y, **si esa transición necesita un texto, lo pide
> ahí**: el motivo al escalar, qué se ha hecho al resolver, la pregunta al ponerse en espera y **el
> comentario al cerrar, que es obligatorio** (decisión 81). Y las acciones que tenían nombre **son esos
> mismos estados**: en un interno, «En espera» devuelve el caso a Soporte sin cerrarlo y «Cerrado» sin
> resolver lo devuelve y lo cierra; en un principal, «En espera» es preguntarle al usuario y «Cerrado»
> desde cerrado es reabrirlo.
>
> **Enmendado el 2026-09-28 (sexta vez)**, a petición del responsable: **la entrada, cuadrada**
> (decisión 79). El recuadro del logo no medía lo mismo que el formulario, y el idioma y el tema medían
> cada uno lo suyo —**medido: logo 136 px, formulario 329, controles 256, idioma 104 y tema 140**—. La
> causa era doble: el logo llevaba **tope de ancho propio** (288 y 320 px) y, sobre todo, **la columna
> de la entrada centra a sus hijos**, así que un componente **encoge hasta lo que mida su contenido**
> aunque por dentro ocupe el 100 %. El ancho se pide ahora **en el elemento** —el logo, la tarjeta y los
> controles—, y los dos desplegables **se reparten a medias**. Medido después: **380 = 380 = 380**, y
> 184 cada desplegable. Hay una prueba que lo mide.
>
> **Enmendado el 2026-09-28 (quinta vez)**, a petición del responsable: **las acciones del ticket, en un
> solo desplegable** (decisión 78). Eran cuatro botones sueltos en la ficha —«Preguntar al usuario»,
> «Escalar», «Resolver», «Cerrar» en el principal; «Pedir algo a Soporte», «No es un cambio de código»,
> «Resolver» y «Cerrar» en el interno— y el responsable no les veía el sentido sueltos. Ahora es **un
> desplegable de acciones**, nombradas **por lo que hacen** —no por el estado al que llevan, que sigue
> en su chip—, y **las que piden algo (el motivo, la pregunta) siguen abriendo su ventana**. Vive **en
> la fila de arriba cuando se ve un solo ticket** y **en la cabecera de cada columna cuando se ven los
> dos**. Y **«Escalar a Desarrollo» pasa a «Escalar»**.
>
> **Enmendado el 2026-09-28 (cuarta vez)**, a petición del responsable: **las ventanas salen centradas**
> (decisión 77). No salían: el `<dialog>` nativo se centra con el `margin: auto` de la hoja del
> navegador, y **la preparación de Tailwind pone `margin: 0` a todo**, así que aparecían en la esquina
> —medido: `x: 0, y: 0`—. Se les devuelve el `auto`, con altura máxima y desplazamiento propio para que
> una ventana alta no se salga. Y **el ✕ de la ventana se llama «Cerrar»**, no «Cancelar»: se llamaba
> igual que el botón de cancelar de dentro, y eran dos botones con el mismo nombre.
>
> **Enmendado el 2026-09-28 (tercera vez)**, a petición del responsable: **los iconos del ticket son los
> del proyecto** (decisión 76). El de copiar el número era el carácter `⧉` —un cuadrado con otro detrás,
> que a 16 px se lee como un cuadrado— y el lápiz de editar, `✏`, que **según la fuente del equipo sale
> como emoji de colores**: los dos se parecían. Ahora son **svg propios** —como los del menú y sin
> bibliotecas—: **dos hojas superpuestas** para copiar, un **visto** para confirmarlo y un **lápiz
> inclinado** para editar. El botón los recibe **por su nombre**, no por su carácter.
>
> **Enmendado el 2026-09-28 (segunda vez)**, a petición del responsable: en las listas, **el nombre del
> módulo, los filtros y «Nuevo ticket» van en una sola fila** (decisión 75), que era lo que faltaba para
> no gastar un renglón de cabecera: la tabla gana ese alto. En un ancho corto se reparten solos —el
> nombre, los filtros y el botón a la derecha si caben, y lo que no quepa, debajo—.
>
> **Enmendado el 2026-09-28**, a petición del responsable, al usar el ticket: **todo se edita donde se
> muestra** (decisión 74). En la ficha, la **categoría** es un desplegable en su propia línea, las
> **etiquetas** son fichas con su aspa y un campo pequeño para añadir, los **observadores** llevan su
> **«Añadir»** con el buscador de personas además del aspa de cada uno, y el **asunto** y la
> **descripción** llevan su lápiz que los convierte **en el sitio** en un campo con «Guardar» y
> «Cancelar». **Se van los formularios aparte** —el de «cambiar la clasificación» y el diálogo de
> editar el ticket—, que era lo que hacía que no se encontrara. Además: el número y el estado del
> ticket **se enseñaban dos veces** —los pintaba la cabecera de la página y la del propio ticket— y
> ahora los pinta sólo la de la página; y en el editor se puede **etiquetar escribiendo `@`**, con la
> lista filtrada **por nombre o por correo** mientras se escribe.
>
> **Enmendado el 2026-09-27 (quinta vez)**, a petición del responsable, al usarlo: **la lista lleva un
> solo botón «Buscar»**, que abre un modal con todos los criterios —texto, vista, tipo, estado,
> categoría y etiqueta—, y en su cabecera queda **el resumen de lo que está filtrado** con su botón de
> quitar; las tres filas de filtros se comían la pantalla y **la de categorías se salía** (sección 3.3).
> El título **«Qué se puede hacer»** de la tarjeta de acciones de la ficha **se retira**, porque se
> malinterpretaba. La **vista del ticket** pasa a ser **la del papel** —principal para Soporte, interno
> para Desarrollo—, «Los dos» enseña **sólo las dos conversaciones** sin las fichas, y el botón
> **«Retirar» de una categoría sólo se le enseña al Administrador**. La pantalla del catálogo pasa a
> tener **dos mitades con el mismo peso**, con las etiquetas pudiéndose **crear, renombrar y retirar**.
>
> **Enmendado el 2026-09-27 (cuarta vez)**, a petición del responsable: en Configuración, **cada
> método de entrada enseña sólo su configuración** —«Cuentas de la aplicación» no enseña ninguna, el
> directorio sale con la organización y el reino con Keycloak—. Antes se veían las dos a la vez
> (sección 3.8).
>
> **Enmendado el 2026-09-27 (tercera vez)**, a petición del responsable: **copiar el número** de un
> ticket es un **botón de sólo icono**, y al copiar **se cambia por un visto verde unos segundos**. El
> inventario de componentes (sección 6.3) recoge las dos cosas que hicieron falta: una **cuarta forma**
> de botón —la verde de la confirmación— y que un botón pueda ser **sólo un icono**, con su nombre
> accesible obligatorio.
>
> **Enmendado el 2026-09-27 (segunda vez)**, al entrar **las categorías y las etiquetas**
> (`docs/modules/tickets.md`, decisiones 64 a 68): el **alta pide la categoría** —es obligatoria: sin
> ella no se crea el ticket— y acepta **etiquetas** en minúsculas con guion medio y sugerencias; la
> **lista** las enseña como **chips en el renglón, debajo del número** (y en las tarjetas del móvil),
> con **chip de categoría y de etiqueta para filtrar**; la **ficha** deja cambiarlas a quien puede
> editar el ticket; y el menú del Administrador y de Soporte estrena la pantalla **«Categorías y
> etiquetas»**, que es donde se mantiene el catálogo.
>
> **Enmendado el 2026-09-27**, al entrar **las menciones y los observadores** (`docs/modules/tickets.md`,
> decisiones 58 a 63): el editor estrena un botón **«Etiquetar»** —y basta escribir `@`— que mete el
> nombre de un técnico o de un desarrollador **dentro del comentario, resaltado**; eso lo convierte en
> **observador** del ticket, que **no es lo mismo que el asignado**. «Mis tickets» lleva un chip nuevo
> —**Asignados · Observo · Todos**—, la ficha tiene su lista de **«Observadores»** con su aspa para
> quitarlos (Soporte y Desarrollo), y quitar a alguien **se cuenta en la línea de tiempo**.
>
> **Enmendado el 2026-09-27**, al pedir el responsable **dos columnas nuevas en la lista de tickets**,
> que redacta un motor de inteligencia artificial (`docs/modules/ai.md`): **«Motivo»** —de qué va el
> ticket— y **«Última acción»** —qué fue lo último que pasó—, las dos **en el idioma de quien mira**.
> Y con ellas, **dos nombres que cambian**: la columna «Número» pasa a llamarse **«Ticket»** y «Última
> actualización» pasa a **«Fecha actualización»** (sección 3.3).
>
> **Enmendado el 2026-09-26**, al pedir el responsable que **la imagen y el vídeo no ocupen todo el
> ancho**: en la descripción y en los comentarios se ven **como mucho a 480 × 360** —una captura a
> pantalla completa se comía la conversación— y **al pulsarlos se abren en el visor grande**, que es
> donde se ven enteros y se descargan. **El vídeo pasa a ser una miniatura con su botón de
> reproducir**, y se reproduce en el visor (sección 6.5). Es la decisión 55 de `tickets.md`.
>
> **Enmendado el 2026-09-26**, al pedir el responsable que **la bandeja sea «lo mío»** y que existan
> **dos módulos nuevos para ver todo**: la Bandeja pasa a ser **Mis tickets** —lo asignado a mí, lo que
> abrí yo y donde he comentado—, y nacen **Tickets principales** y **Tickets internos**, sólo para
> Soporte y Desarrollo (secciones 3.2, 3.3 y 3.7). **«Nuevo ticket» sale del menú** —se confundía con
> un módulo— y se queda donde se crea, en el botón de la bandeja. **«Asignármelo» desaparece de la
> bandeja**: la reasignación se hace dentro del ticket, que es donde se ve a quién se le pasa. Las
> filas de la bandeja dicen **«Abrir»** y no «Abrir el ticket», sobra el texto «Lo que lleva más tiempo
> esperando va primero…» —la regla sigue, callada—, y **el idioma pasa a ser un desplegable como el del
> tema**, con su ancho y sin etiqueta a la vista (secciones 3.1 y 6.3).
>
> **Enmendado el 2026-09-26**, al implementar el editor con adjuntos: **pegar texto pega
> sin formato** —con los saltos de línea—, y **un archivo pegado sí entra como adjunto**, porque pegar
> el HTML de un correo traería etiquetas que el backend rechaza y se perdería el comentario entero por
> algo que la persona no ha escrito (sección 6.5). Y **el hueco de un archivo que no subió se ve
> marcado**, en el sitio donde iba.
>
> **Enmendado el 2026-09-26**, al pedir el responsable que **los adjuntos se vean dentro del texto** y
> que **la vista previa se vea también al escribir**: el cuadro de escribir pasa a ser un **editor con
> formato** (sección 6.5), el **visor** se agranda y gana **Descargar** y **Abrir en una pestaña**, y
> **se retira** la pieza de subida de archivos, porque adjuntar ya no es un recuadro aparte.
>
> **Enmendado el 2026-09-25**, al pedir el responsable que **la versión del sistema se
> vea**: en la fila de salir del menú lateral, alineada a la derecha, y **también en el pie de la
> pantalla de entrada** (secciones 3.1 y 6.4). **Es elección suya y no mi recomendación**: yo proponía
> sólo el menú, y decidió que también se vea antes de entrar, que es donde sirve para decir qué versión
> se está mirando al reportar un fallo.
>
> **Pasa a as-built el 2026-09-25**, y **enmendado el mismo día**: la **tarjeta de cómo se entra** en
> la pantalla de Configuración —el método, las dos configuraciones y sus dos pruebas de conexión— y lo
> que la pantalla de entrada ofrece según el método (sección 4.4). La forma de las pantallas que este
> documento fija está toda construida.
>
> **Enmendado el 2026-09-25**: la instalación tiene **nombre propio**, configurable, y
> **es él quien lo lee todo el mundo** —la pantalla de entrada, el menú lateral y **la pestaña del
> navegador**— en lugar del texto de fábrica (sección 6.4 y `docs/modules/settings.md`, sección 5.7).
> **El responsable corrigió dos cosas al pedirlo**: los correos y su remitente no cambian.
>
> **Enmendado el 2026-09-25**, al implementar las tres acciones que dependían del
> directorio: en las pantallas de usuarios, **reactivar sólo se ofrece donde puede funcionar** —en una
> cuenta de AD sí, y en una de Keycloak no, con lo que pasa contado en su lugar— y **la nota del
> origen** dice ya lo que ocurre de verdad: las cuentas de directorio no se dan de alta desde aquí
> porque entran solas (`docs/modules/users.md`, sección 5, punto 4).
>
> Aprobado por el responsable el 2026-09-22, tras repasarlo en forma de preguntas mientras se
> escribía: **no hay barra superior** —el menú lateral lleva el producto, las opciones del papel y,
> abajo y separados, los controles (perfil, tema y salir)— y el papel `administrador` **ve los
> tickets en sólo lectura**. Fija los principios (intuitivo antes que bonito), la forma de la
> aplicación, la experiencia de los cuatro papeles, la vista doble principal/interno, el lenguaje
> visual con **Tailwind v4 y componentes propios**, los ocho temas y lo que se espera en cada
> dispositivo.
>
> Con este documento, la **documentación está completa**: siete documentos que cubren qué se
> construye, cómo se comporta, cómo se ve y cómo se despliega.
> **Enmendado el 2026-09-23**, al escribir el armazón de la sesión: se añade **de dónde sale el
> idioma de la interfaz** (principio 6: del navegador, y si no pide ninguno de los dos, inglés).
> Es una **corrección del responsable**: yo había propuesto español por defecto
> (`docs/modules/auth.md`, decisión 25).
>
> **Enmendado el 2026-09-24**, al construir el armazón: la sección 3.1 dice que **el
> estado plegado del menú se recuerda** y que **el menú enseña sólo lo que existe** —las entradas de
> la tabla 3.2 llegan con sus pantallas—, y la 6.4 recoge el logo en el menú.
>
> **Enmendado el 2026-09-24**, al entrar el logo de la institución: la sección 6.4 dice **dónde se
> enseña la marca** (entrada, menú lateral y Configuración) y **cómo se enseña cualquier logo**
> —entero, sin recortar y en un recuadro con el borde del tema—, para que sirva igual uno con fondo
> opaco que uno con transparencia. El logo y su reemplazo son del módulo `settings`
> (`docs/modules/settings.md`).
>
> **Enmendado el 2026-09-24**, al construir el editor de los correos: la sección 3.8 fija
> **su pantalla**, con los dos idiomas al lado, los marcadores que se insertan al pulsarlos y una vista
> previa que **renderiza el backend** con los datos de ejemplo de la prueba, para que lo que se ve sea
> lo que va a salir.
>
> **Enmendado el 2026-09-24**, al construir las pantallas de tickets: la sección 3.7 fija
> **las cuatro rutas** y quién ve cada una, y recoge que **el detalle es la misma pantalla para los
> cuatro papeles** y que **el prefijo y el reparto se configuran en Configuración**, no en una pantalla
> de `tickets`. La tabla de la sección 3.2 estrena **«Nuevo ticket» para Soporte** —tres entradas—:
> puede crear un ticket en nombre de otra persona y ese permiso necesitaba su puerta (**corrección del
> responsable**). El escalado se escribe **en un diálogo**, porque el motivo es obligatorio y la acción
> es deliberada, y los adjuntos se ven **con vista previa** cuando son imágenes o PDF.
>
> **Enmendado el 2026-09-24**, al empezar los tickets: la sección 3.3 añade el **chip de
> tipo** (Principales · Internos · Todos) en la bandeja del Administrador, que es el único papel que
> ve los dos tipos de ticket.
>
> **Enmendado el 2026-09-24**, al construir las pantallas de usuarios: la sección 3.6
> fija **las tres pantallas de `users`** —la lista, la ficha y el perfil propio—, sus direcciones y
> quién ve cada una. Las cinco decisiones son del responsable, del 2026-09-24, y salen de un hallazgo:
> **Soporte puede cambiar el nombre y los apellidos de otra cuenta, pero no ve la ficha de nadie**, así
> que ese permiso se ejerce **en línea, desde la propia lista**. La sección 3.1 recoge además que **el
> menú mide la zona visible y se desplaza por dentro** —medía el alto del documento, y en móvil dejaba
> el botón de salir fuera de la pantalla—, y la 6.3 que la tarjeta **elige el nivel de su título**,
> porque una pantalla con tres `h1` no se navega con un lector de pantalla. Los dos los encontró la capa
> de Playwright, no las pruebas de unidad.
>
> **Enmendado el 2026-09-23**, al montar el tema: la sección 6.2 dice **cómo se implementa**
> (un atributo en `html`, las paletas escritas una vez con `light-dark()`, la elección recordada en
> el navegador y el tema aplicado por código nuestro **porque la CSP de producción no admite scripts
> incrustados**). El conmutador tiene **dos opciones, claro y oscuro**: **«automático» no se muestra
> ni se elige**, porque seguir al sistema es no haber elegido. Las seis paletas fijas quedan para el
> armazón completo, cuando se aprueben sus nombres y sus colores.

## 1. Alcance de este documento

Fija **cómo se ve, cómo se navega y qué siente cada papel**: la estructura de las pantallas, el
lenguaje visual, el comportamiento en cada dispositivo y la experiencia de los cuatro papeles.

**No repite** lo que cada papel puede hacer (`docs/usuarios-y-permisos.md`), ni los recorridos
(`docs/flujos.md`), ni el modelo de datos (`docs/modules/tickets.md`). Cuando este documento y aquellos no
coincidan, mandan ellos: aquí sólo se decide cómo se presenta.

Está **aprobado**, así que habilita escribir código (Regla 0). La sección 11
recoge lo que he propuesto yo.

## 2. Los principios, en orden

El orden importa: cuando dos choquen, gana el de más arriba.

1. **Intuitivo antes que bonito.** Si hay que elegir entre algo que se entiende sin explicación y
   algo que se ve mejor, gana lo primero. Un producto que hay que explicar es un producto que se
   usa mal.
2. **Sin ruido** (el criterio del proyecto, `docs/propósito-y-alcance.md`): lo que no aporta, no
   está. Ni animaciones decorativas, ni contadores que nadie mira, ni avisos que se cierran sin
   leer.
3. **Cada uno ve su trabajo, no la aplicación entera.** El usuario no necesita saber que existe un
   ticket interno, ni un desarrollo necesita ver la administración de usuarios.
4. **El estado se lee de un vistazo.** Mirar la lista y saber qué está esperando a quién.
5. **Nada importante se hace por accidente.** Cerrar, escalar o borrar se confirman; lo demás, no.
6. **El idioma es global.** Una instalación nueva abre `/setup` en inglés y su primer control puede
   cambiar toda la pantalla. Después manda siempre el idioma guardado por la instalación, también
   antes de entrar y en sesiones abiertas; no hay selector ni preferencia por cuenta.
7. **Lo que falla, se explica.** Un error dice qué pasó y qué hacer, en el idioma de quien lo lee.

## 3. La forma de la aplicación

### 3.1 El armazón

```text
┌───────────┬───────────────────────────────────────────────┐
│  producto │                                               │
│           │                                               │
│  opciones │  contenido                                    │
│  del papel│                                               │
│           │                                               │
│  ───────  │                                               │
│  controles│                                               │
└───────────┴───────────────────────────────────────────────┘
```

- **No hay barra superior.** El menú lateral lo lleva todo, y el contenido gana todo el ancho: en
  una mesa de ayuda la pantalla se llena de listas, conversaciones y adjuntos, y una franja
  superior es espacio que se le quita a lo que importa.
- **El menú lateral tiene tres zonas**, separadas visualmente:
  1. **Arriba**: el producto —**el logo de la instalación y el nombre**—. Es lo único que no se mueve.
  2. **En medio**: las **opciones del papel** (sección 3.2). Es lo que cambia de una persona a otra.
  3. **Abajo, separado por una línea**: los **controles**, que son los mismos para todos —**tu
     nombre (y tu perfil)**, **el idioma**, **el tema** y **salir**—. Es el mismo reparto que usa
     Calibyou, y evita mezclar «lo que hago aquí» con «quién soy».
- **El idioma y el tema son el mismo control, uno debajo del otro** (decisión del responsable,
  2026-09-26): los dos son **desplegables**, ocupan **el mismo ancho** y **ninguno lleva la etiqueta a
  la vista** —se lee «Español» o «Claro»—, porque el valor ya dice qué es. El idioma se cambia
  **eligiendo cualquier valor de la lista**, y el nombre accesible sigue ahí, oculto, para quien navega
  con lector de pantalla. Antes el idioma eran dos botones segmentados y el tema un desplegable: dos
  formas para lo mismo, y dos anchos distintos en la misma columna.
- **La versión del sistema se lee en la fila de salir, alineada a la derecha** (2026-09-25): es un
  dato de la instalación, no una opción, así que **no es un botón ni ocupa una fila propia**: va con el
  control que ya está en esa fila. En el menú plegado se pierde con el texto, que es lo correcto: un
  número suelto al lado de un icono no dice nada. **En móvil se lee dentro del cajón**, que es donde
  vive esa fila.
- En PC el menú se puede **plegar** y queda una tira de iconos; en tablet se plegará solo. **El
  estado se recuerda** en el navegador, como el tema: quien trabaja con el menú plegado no tiene que
  plegarlo cada mañana.
- **Mide 15 rem (240 px)** y **4 rem (64 px) plegado** (ajustado el 2026-09-26: eran 16 rem). Es un
  ancho que **enseña entero lo que hay dentro** —las opciones, tu nombre, salir y la versión— y que le
  deja 16 px más al contenido, que es donde se leen los tickets. **Un nombre largo se corta con puntos
  suspensivos**, no ensancha el menú ni parte la línea: lo que no puede pasar es que el menú crezca
  según lo que alguien se llame.
- **El menú enseña lo que existe**: cada entrada llega **con su pantalla**, no antes —un menú que
  lleva a ninguna parte es peor que un menú corto—, y no se enseña a quien no le toca. El usuario ve
  una sola opción, **Mis tickets**, y eso es todo lo que necesita: pedir y seguir lo suyo.
- **El menú mide la zona visible, no el documento, y se desplaza por dentro si su contenido no cabe**
  (decisión del responsable, 2026-09-24, al construir las pantallas de usuarios). El cajón medía el
  alto del documento: en un móvil con barra de navegador lo visible son 839 px de 890, así que **el
  botón de salir quedaba por debajo de lo que se ve**, imposible de pulsar; y en PC, con una lista
  larga, la barra se estiraba a todo el contenido y sus controles acababan al final de la página. En
  PC la barra va **pegada arriba** mientras el contenido se desplaza, que es lo que la sección 7 llama
  «fijo».
- **En móvil sí hay una tira superior mínima**, con el botón de menú y el nombre del producto: sin
  ella no habría dónde poner el disparador. Todo lo demás vive en el menú, que se abre como cajón.

### 3.2 Las entradas del menú, por papel

| Papel | Opciones del menú |
| --- | --- |
| **Usuario** | Mis tickets |
| **Soporte** | Mis tickets · Tickets principales · Tickets internos · **Categorías y etiquetas** · Usuarios |
| **Desarrollo** | Mis tickets · Tickets principales · Tickets internos |
| **Administrador** | Bandeja (sólo lectura) · **Categorías y etiquetas** · Usuarios · Configuración |

**Cinco entradas como máximo** en el caso más cargado —Soporte— y cuatro el Administrador, y **los
controles no cuentan**: el perfil, **el idioma**, el tema y salir van abajo y son iguales para todos.
**Eran cuatro hasta el 2026-09-27**, que entró **«Categorías y etiquetas»** con las categorías: es una
pantalla de tickets —el catálogo son datos de los tickets, no configuración de la instalación— y por eso
vive en el menú y no dentro de Configuración (`docs/modules/tickets.md`, decisión 64). Si algún papel
llegara a necesitar más opciones, es señal de que algo se ha complicado en otro sitio.

**«Nuevo ticket» ya no está en el menú** (decisión del responsable, 2026-09-26): se confundía con un
módulo, y crear un ticket es una acción de la bandeja, no un sitio al que ir. El botón vive **dentro
de la bandeja** —arriba y también cuando está vacía—, así que **no se pierde nada**: la ruta
`/tickets/new` sigue existiendo y sigue siendo la misma pantalla.

**Soporte tiene cinco** porque, además de lo suyo, es el único papel —con Desarrollo— que ve **todo
lo que hay**, y ahora eso se ve en dos sitios de una vez: los principales y los internos. Antes estaba
todo junto en la bandeja, que es lo que dejaba de funcionar cuando la bandeja pasó a ser «lo mío».

### 3.3 Las bandejas

**Tres bandejas y una pantalla.** La misma pantalla de lista sirve para las tres —cambia el contenido
y el título—, porque separarlas en tres pantallas sería triplicar la tabla, los filtros y la paginación
para cambiar una condición (`docs/modules/tickets.md`, sección 8).

| Bandeja | Quién | Qué enseña |
| --- | --- | --- |
| **Mis tickets** (`/tickets`) | Usuario, Soporte y Desarrollo | **Lo mío**: lo que tengo asignado, lo que abrí yo y aquello en lo que he comentado |
| **Tickets principales** (`/tickets/main`) | Soporte y Desarrollo | **Todos** los principales |
| **Tickets internos** (`/tickets/internal`) | Soporte y Desarrollo | **Todos** los internos |
| **Bandeja del Administrador** (`/tickets`) | Administrador | Los dos tipos, con su chip y de sólo lectura |

- **«Lo mío» es lo que toca de verdad** (decisión del responsable, 2026-09-26): lo asignado a mí, lo
  que abrí yo y **aquello donde he comentado**, porque un ticket en el que ya se ha hablado se sigue
  hasta el final. Participar es **haber comentado en el hilo de ese ticket** —en un principal, el
  principal; en un interno, el interno—, sin mirar el papel de quien comenta.
- **La lista es una tabla** en PC y **tarjetas** en móvil, donde una tabla de ocho columnas no se lee.
  Las columnas son: **Ticket** · **Asunto** · **Motivo** · **Última acción** · **Clasificación** ·
  **Estado** · **Solicitante** · **Responsable** · **Fecha actualización** (los dos nombres de los
  extremos y las dos columnas del medio son del 2026-09-27: «Número» era lo que se lee como un ticket, y
  «Última actualización» es una fecha, no un suceso; **«Clasificación» es del 2026-09-29** y lleva la
  categoría y las etiquetas, decisión 86 — y **«Ticket» se queda con el número solo**—).
- **Todo se edita donde se muestra** (decisión del responsable, 2026-09-28): la categoría, las
  **etiquetas**, los observadores, el asunto y la descripción **se cambian en el sitio donde se leen**,
  no en un formulario aparte. El reparto, para **el texto y la categoría**, es el de siempre: **el
  solicitante en su ticket y Soporte**; Desarrollo no escribe el principal y el Administrador no toca
  nada, así que **a ellos no se les enseña el lápiz**. **Lo único que sí hace Desarrollo en un principal es seguir a quien lo sigue**: la
  lista de observadores —añadir y quitar— es de **los dos equipos** (decisión 63), porque seguir un
  ticket no es escribirlo. **Y las etiquetas también son de los dos** (decisión 85): se ponen, se cambian
  y se quitan **desde el ticket en el que trabaja cada uno** —Soporte en el principal, Desarrollo en el
  interno—, y como **son del principal** (el interno las hereda, decisión 67), cambiarlas desde cualquiera
  de los dos cambia las mismas. Lo encontró la prueba de interfaz al comprobar que no le dejábamos nada —la interfaz no ofrece lo que el backend rechazaría—. Cada control dice lo que
  hace y **el cambio se guarda al confirmarlo**, nunca al escribir.
- **Etiquetar escribiendo `@`** (decisión del responsable, 2026-09-28): en el editor, escribir `@` abre
  la lista **filtrada por lo que se sigue escribiendo**, y vale **el nombre o el correo** —el buscador
  enseña **los dos**, que es lo que permite distinguir a dos personas que se llaman igual—. La mención
  **se lee con el nombre**, como hasta ahora, y el botón «Etiquetar» sigue estando para quien no se
  acuerde del atajo.
- **Mandar el enlace pregunta antes** (decisión del responsable, 2026-09-28): el enlace es el de alta
  para quien nunca ha entrado y el de **recuperación** para quien ya tiene contraseña —la que tuviera
  deja de valer—, así que la pantalla **pregunta antes con una ventana**, y el aviso dice **cuál de los
  dos es** y cuánto caduca: 24 horas el de alta, 1 hora el de recuperación. No se usa un `confirm()` del
  navegador, como en el resto de las acciones que preguntan.
- **En la cabecera del ticket, el número del interno** (decisión del responsable, 2026-09-28): abriendo
  un interno, esa fila enseñaba sólo el número del principal y no se sabía cuál de los dos se estaba
  mirando. Va en apagado, porque es la referencia y no el título, y **el botón de volver es un icono con su nombre accesible «Volver»**
  —decía «Volver a la bandeja», y a la bandeja se vuelve desde muchos sitios—.
- **Los observadores se añaden también a mano** (decisión del responsable, 2026-09-28): además de
  llegar por una mención, se pueden **añadir desde su lista** con el buscador de personas y **quitar**
  con su aspa. Quien se añade así **queda igual** que quien llegó por una mención: todos son
  observadores, y la lista no distingue de dónde vino cada uno.
- **La tarjeta de acciones de la ficha no lleva título** (decisión del responsable, 2026-09-27): decía
  «Qué se puede hacer», que **se malinterpreta** —parece una promesa de lo que el sistema hará, y es
  sólo dónde están los botones—. Los botones siguen diciendo lo que hacen, que es lo que hacía falta.
- **La vista del ticket, por papel** (decisión del responsable, 2026-09-27): al abrir un ticket,
  **Soporte ve el principal** y **Desarrollo el interno**; los dos pueden cambiar a **Principal · Los
  dos · Interno** cuando quieran, y **abrir el número de un interno enseña el interno** aunque seas
  Soporte. **«Los dos» enseña sólo las dos conversaciones** —la columna de cada ticket—: con las cuatro
  columnas (dos conversaciones y dos fichas) «se ve muy saturado». Para actuar sobre uno, se elige ese.
- **Copiar el número es un botón de sólo icono** (decisión del responsable, 2026-09-27): el
  portapapeles, y **al copiar se cambia por un visto verde durante dos segundos** y vuelve solo. **El
  dibujo es el del proyecto** —dos hojas superpuestas—, y no se confunde con el lápiz de editar
  (2026-09-28). El
  nombre accesible **cambia con el estado** («Copiar el número» → «Copiado»), así que quien no ve el
  color sí oye la confirmación; el verde sale del tema (`--exito`), no escrito a mano.
- **El renglón lleva sus chips de clasificación** (2026-09-27, **movidos el 2026-09-29**, decisión 86):
  **un chip de categoría** —el *qué es* del ticket— y los de sus etiquetas. En la **tabla** viven en su
  propia columna, **«Clasificación»**; en las **tarjetas del móvil**, donde no hay columnas, siguen
  debajo del número. Un chip ocupa lo que mide su nombre y se lee de un vistazo. **La categoría no se
  recorta**: es una palabra o dos.
- **Un solo botón de búsqueda, y todo dentro** (decisión del responsable, 2026-09-27): la lista lleva
  **«Buscar»**, y el modal que abre tiene **todos los criterios** —el texto, la vista, el tipo, el
  estado, la categoría y la etiqueta— con sus botones de **buscar** y **limpiar** dentro. En su lugar
  no queda nada más que **el resumen de lo que está filtrado** —«3 filtros»— con su botón de quitar:
  la lista se ve limpia y **se entiende por qué salen pocos tickets**. Antes eran **tres filas de
  filtros** para cuatro criterios, y la de categorías **se salía de la pantalla** cuando el catálogo
  crecía; un modal no crece, se llena.
- **Y las dos (categoría y etiqueta) entran en la búsqueda**: buscar «red» encuentra la categoría «Red»
  y la etiqueta `red`, que es para lo que se pidieron.
- **«Motivo» y «Última acción» son redacciones del motor de IA** (`docs/modules/ai.md`), y se enseñan
  **recortadas a dos líneas**, porque son un resumen para barrer la lista y no el texto del ticket: el
  completo se lee en la ficha. Mientras el motor no ha contestado, dicen **«Generando…»**; si no hay
  motor o falló, lo dicen y **no se inventan nada** (decisiones 2 y 8 de `docs/modules/ai.md`).
- **En las tarjetas del móvil van las dos**, como dos líneas más: en móvil es donde más se agradece
  saber de qué va un ticket sin abrirlo.
- **La fila lleva «Abrir»**, no «Abrir el ticket» (2026-09-26): en una columna de acciones la palabra
  «ticket» sobra, y la fila entera ya habla de un ticket.
- **No hay «Asignármelo»** (2026-09-26). Se quitó de la tabla y de las tarjetas: **la reasignación se
  hace dentro del ticket**, en su ficha, donde se ve a quién se le pasa, y con las reglas de siempre
  (`docs/modules/tickets.md`, decisión 50). El atajo de la lista asignaba sin más, y eso es lo que
  dejaba sin hacer la reasignación a otro compañero.
- Los filtros son **chips visibles** encima de la lista, no un formulario escondido: estado, texto y
  nada más.
- **«Mis tickets» lleva además un chip de vista —Asignados · Observo · Todos—** para Soporte y
  Desarrollo (2026-09-27): **arranca en «Asignados»**, que es lo que hay que atender, y **«Observo»** es
  dónde se mira lo que uno sigue sin atenderlo. Mezclarlos sería una bandeja que crece sola en cuanto
  alguien te etiqueta.
- **Y un chip de tipo —Principales · Internos · Todos—** en **Mis tickets** para Soporte y Desarrollo,
  y en la bandeja del Administrador. **Cada uno entra por el suyo**: Soporte por **Principales** y
  Desarrollo por **Internos**, que es lo que mira la mayor parte del tiempo; y puede cambiar o ver los
  dos, porque un técnico de Soporte tiene internos donde ha comentado y uno de Desarrollo puede haber
  abierto un principal. **El usuario no lleva chip**: sus tickets son principales, y un filtro que sólo
  tiene una opción es ruido. Lo mismo en las dos bandejas del «todo», donde el tipo es la pantalla
  misma.
- **Las dos bandejas del «todo» no llevan chip de tipo** por lo de arriba: «Tickets principales»
  enseña principales y «Tickets internos», internos. Sí llevan el chip de **estado** y la búsqueda.
- El **estado va con color y con texto**: el color ayuda a barrer la lista, el texto dice qué es.
  Nunca sólo color (sección 8).
- Orden por defecto: **lo que lleva más tiempo esperando, primero**, dentro de lo que no está
  cerrado. Lo cerrado no estorba. **La regla sigue escrita aquí y ya no se escribe en la pantalla**
  (2026-09-26): el orden se ve en la lista, y una frase que explica lo que ya se ve es ruido.

### 3.4 El detalle del ticket: la pieza central

Es la pantalla donde se pasa el tiempo, así que se lleva la mayor parte del diseño.

```text
┌──────────────────────────────────────────────────────────────┐
│  ACME-2026-0042 · asunto                    [estado]         │
├──────────────────────────────┬───────────────────────────────┤
│  conversación                │  ficha: solicitante,          │
│  (comentarios, adjuntos,     │  responsable, fechas,         │
│   historial intercalado)     │  adjuntos                     │
│                              │  acciones                     │
├──────────────────────────────┴───────────────────────────────┤
│  escribir un comentario…                       [Enviar]      │
└──────────────────────────────────────────────────────────────┘
```

- **Una sola línea de tiempo**: los comentarios de las personas y lo que hizo el sistema, en orden.
  El historial no es una pestaña aparte: si cambió el estado, se ve ahí, entre los mensajes.
- **La ficha y las acciones van a la derecha** en PC: quién lo pidió, quién lo lleva, desde cuándo, y
  los botones de estado, asignación y escalado. En móvil bajan debajo, plegadas.
- **El número del ticket siempre visible y con un botón de copiar**: es lo que se dice en voz alta y
  lo que se pega en un correo.
- La caja de comentario está **siempre a la vista** al final de la conversación. Escribir es la
  acción más frecuente del sistema; esconderla detrás de un botón es un error de principiante.
- Un ticket **cerrado no admite comentarios** (`docs/flujos.md`): en su lugar aparece la explicación de
  por qué, y **reabrir se elige en el desplegable de acciones** (decisión 78), que es donde viven todas
  las acciones del ticket; no hay dos controles para lo mismo.

### 3.5 El usuario no lee jerga interna

El usuario **nunca** ve el ticket interno ni la palabra `escalado`. Y sus estados se le cuentan en
lenguaje llano:

| Estado interno | Lo que ve el usuario |
| --- | --- |
| `nuevo` | Recibido |
| `en progreso` | En curso |
| `en espera` | Esperamos tu respuesta |
| `escalado` | En curso *(el usuario no tiene por qué saber que ha pasado a Desarrollo)* |
| `resuelto` | Resuelto |
| `cerrado` | Cerrado |

### 3.6 Las pantallas de usuarios: la lista, la ficha y el perfil

Son **las primeras pantallas de producto**, y son tres, cada una con su dirección:

| Pantalla | Ruta | Quién la ve |
| --- | --- | --- |
| La lista de cuentas | `/users` | Soporte y Administrador |
| La ficha de una cuenta | `/users/{id}` | Administrador |
| El perfil propio | `/profile` | Cualquiera que haya entrado, menos la cuenta de fábrica |

- **La lista es una tabla en PC** —nombre, correo, papel, origen, estado y última entrada, más las
  acciones de la fila— y **tarjetas en móvil**: una tabla de seis columnas no se lee en un teléfono, y
  no se arrastra de lado (sección 7).
- **Los filtros son chips visibles encima de la lista**, como los de las bandejas (sección 3.3): papel,
  origen y estado, más la búsqueda por nombre o correo. Nada escondido detrás de un botón.
- **El orden es fijo —apellidos y nombre— y la lista se pagina**, con el total a la vista. **Ordenar
  por columna no se ofrece**: eso es de la tabla de las bandejas, que es donde hace falta.
- **El alta es un diálogo sobre la lista**, no una pantalla: seis campos y un botón. Al guardar, la
  lista se refresca con la cuenta nueva en su sitio, sin salir de donde se estaba.
- **Soporte edita el nombre y los apellidos en línea, desde la propia lista** (decisión del
  responsable, 2026-09-24). Soporte **no ve la ficha de nadie** (`docs/modules/users.md`, sección 4,
  decisión 2), y ese permiso tiene que poder ejercerse en algún sitio: el hallazgo fue justamente que,
  sin ficha, no había dónde. El Administrador cambia lo mismo desde la ficha.
- **La ficha** enseña la cuenta —nombre, correo, papel, origen, estado y cuándo entró por última
  vez— y **sus acciones**: cambiar el nombre, el correo, el papel y el idioma, desactivar o reactivar,
  cambiar el origen y **lanzar otra vez el enlace de contraseña**. Toda acción sensible dice lo que va
  a pasar antes de hacerlo (sección 4.4).
- **Reactivar se ofrece sólo donde puede funcionar**, y no es un detalle de forma: en una cuenta de
  **AD** el botón pregunta al directorio antes de devolver el acceso, y en una de **Keycloak** no se
  ofrece —se cuenta que **vuelve sola al entrar**, que es lo que pasa de verdad— porque su camino no
  tiene forma de comprobar quién sigue allí sin una cuenta de servicio en el reino
  (`docs/modules/users.md`, sección 5, punto 4). **La interfaz no ofrece lo que no se puede hacer.**
- **El botón de desactivar no se ofrece sobre uno mismo** (`docs/modules/users.md`, sección 7):
  desactivarse es quedarse fuera al instante y sin poder volver.
- **El origen se ofrece con sus tres valores, y `ad` y `keycloak` salen desactivados**, con una nota
  que dice lo que pasa: **las cuentas de directorio no se dan de alta desde aquí**, porque quien está
  en el directorio entra con su cuenta y la suya se crea sola en ese primer acceso
  (`docs/modules/users.md`, sección 5, punto 4).
- **El perfil propio se abre desde tu nombre**, en la zona de controles del menú (sección 3.1), y
  lleva **tu nombre, tus apellidos y tu idioma**, más el enlace a cambiar la contraseña. El idioma se
  cambia aquí y en los controles: es el mismo ajuste, y es el que decide en qué idioma se te escriben
  los correos (`docs/modules/users.md`, sección 8).
- **La cuenta de fábrica no tiene perfil**: no está en la tabla de cuentas, así que **su nombre en el
  menú no lleva a ninguna parte** y la pantalla no se le enseña.

### 3.7 Las pantallas de tickets: la bandeja, el alta y el detalle

Cuatro rutas para las listas, y **el detalle es la misma pantalla para los cuatro papeles**: lo que
cambia es lo que cada uno puede hacer dentro, no la pantalla.

| Pantalla | Ruta | Quién la ve | Cómo se llama en el menú |
| --- | --- | --- | --- |
| **Mis tickets** | `/tickets` | Usuario, Soporte y Desarrollo | **Mis tickets** |
| **Tickets principales** | `/tickets/main` | Soporte y Desarrollo | **Tickets principales** |
| **Tickets internos** | `/tickets/internal` | Soporte y Desarrollo | **Tickets internos** |
| **La bandeja del Administrador** | `/tickets` | Administrador | **Bandeja** |
| El alta | `/tickets/new` | Usuario, Soporte y Desarrollo | — (se entra por el botón de la bandeja) |
| El detalle | `/tickets/{numero}` | Quien pueda verlo | — |
| **Categorías y etiquetas** | `/tickets/categories` | Soporte (mantiene el catálogo) y Administrador (además lo retira) | **Categorías y etiquetas** |
| El prefijo y el reparto | `/settings` | Administrador | **Configuración** |

- **Una sola pantalla de lista con cuatro nombres**: es el papel el que decide qué se ve dentro, y el
  documento ya lo llama distinto por papel. El nombre es lo único que cambia.
- **Las cuatro direcciones son del módulo `tickets`** y ninguna pide nada nuevo al backend: las listas
  ya se distinguen por su tipo y por de quién es el ticket (`docs/modules/tickets.md`, sección 5).
- **La pantalla del catálogo tiene dos mitades con el mismo peso** (decisión del responsable,
  2026-09-27): **categorías** y **etiquetas**, cada una con su lista, su cuenta de tickets y sus
  botones. Las etiquetas se pueden **crear**, **renombrar** —lo que vale para todos los tickets que la
  llevan— y **retirar** —lo que las quita de ellos, avisando antes—; el botón de retirar **sólo se le
  enseña al Administrador** y el de crear y renombrar, a Soporte y al Administrador.
- **La pantalla del catálogo es de `tickets` y no de Configuración** (2026-09-27): lista las
  categorías con su nombre y **cuántos tickets tiene cada una**, deja **crear y renombrar** a Soporte y
  al Administrador, y **retirar o volver a poner** sólo al Administrador —retirar una categoría en uso
  se permite y **no toca los tickets que la tienen**: deja de ofrecerse al crear—. Debajo, **la lista de
  etiquetas que se están usando**, con su cuenta, que es lo que hace que la gente vea que `red-wifi` ya
  existe y no escriba `wifi-red`.
- **El Administrador se queda con la bandeja de siempre**, con su chip de tipo y de sólo lectura
  (decisión del responsable, 2026-09-26): **no tiene tickets propios**, así que «Mis tickets» no
  significa nada para él, y las dos bandejas del «todo» le darían lo mismo que ya tiene, repartido en
  dos.
- **El alta es una pantalla y no un diálogo**, al revés que el alta de una cuenta: aquí se escribe un
  texto largo y se adjuntan archivos, y eso necesita sitio. En móvil es una pantalla completa.
- **Las direcciones las fijó el backend antes que la pantalla**: los correos que ya salen llevan
  `/tickets/ACME-2026-0042`, así que esa ruta no se elige aquí, se respeta.
- **El detalle de un ticket interno es el mismo detalle**, con su número (`INT-…`) y su conversación:
  no hay una pantalla aparte para los internos.
- **El prefijo de la numeración y el reparto se configuran en Configuración**, no en una pantalla de
  `tickets`: viven en la instalación y sus endpoints son de `settings`. Una pantalla del módulo
  `tickets` que llamara a `/api/settings` rompería la regla dura de modularidad. `tickets.md`, sección
  8, los listaba como pantalla suya y queda corregido (2026-09-24).

### 3.8 El editor de los correos

Vive en el módulo `mail`, en `/mail`, y **Configuración lleva hasta él**: es una pantalla de otro
módulo a propósito, porque llamar a `/api/mail/**` desde `settings` rompería la regla dura de
modularidad (sección 4.4).

| Pantalla | Ruta | Quién la ve |
| --- | --- | --- |
| El editor de los correos | `/mail` | Administrador |

- **A la izquierda, los once correos**; a la derecha, el que se está editando: su **asunto**, su
  **cuerpo** y sus **marcadores**, con los imprescindibles marcados.
- **Los dos idiomas van al lado** —el mismo correo en español y en inglés, para ver que dicen lo
  mismo— **y en móvil se apilan**: en un teléfono no caben dos columnas.
- **El cuerpo se edita con formato, no escribiendo etiquetas**: negrita, cursiva, lista y enlace
  **sobre el texto seleccionado**, y el HTML lo escribe la aplicación. También **acepta HTML escrito a
  mano**, porque quien administra la instalación sabe lo que hace.
- **Los marcadores se insertan al pulsarlos**, en el punto donde esté el cursor.
- **La vista previa va al lado, y la renderiza el backend** con datos de ejemplo: no es una
  aproximación, es el correo que va a salir. Debajo, el botón de **«enviarme una prueba»**, que manda
  el correo de verdad a la dirección de quien lo pide.
- **Restaurar el texto de fábrica** está en cada plantilla, y el aviso de que falta un marcador
  imprescindible **no impide guardar**: hay avisos legítimos sin enlace.

## 4. Los cuatro enfoques

### 4.1 Usuario: pedir y seguir

Lo que hace un usuario es **crear un ticket y responder**. Todo lo demás sobra.

- **Al entrar**: sus tickets, los abiertos primero. Un botón grande y visible de **Nuevo ticket**.
- **Crear**: asunto, descripción y adjuntos. Tres campos, ninguno con jerga. Si se puede arrastrar
  una captura sobre la ventana, mejor que buscarla.
- **Seguir**: la conversación, con sus comentarios y los de Soporte. Cuando Soporte le pide algo, el
  estado lo dice —«Esperamos tu respuesta»— y el correo se lo ha avisado.
- **Cerrar y reabrir**: las dos, en el **desplegable de acciones** del ticket, y la de cerrar pregunta
  antes (`docs/usuarios-y-permisos.md`: cierra también el usuario).
- **Lo que no ve**: notas internas (no existen), el ticket interno, la lista de usuarios, nada de
  administración. Su mundo son sus tickets.

### 4.2 Soporte: atender y escalar

Su problema no es entender la aplicación, es **no perder de vista qué le toca**.

- **Al entrar**: la **bandeja**, y dentro de ella lo que reclama atención primero — `nuevo`,
  `en espera` y `escalado`. Un vistazo y se sabe por dónde empezar.
- **Asignarse**: un clic desde la lista, sin abrir el ticket.
- **Atender**: comentar, mover el estado, pedir información al usuario (un botón que dice
  **«Preguntar al usuario»**, no «pasar a `en espera`»), resolver o **escalar**.
- **Escalar**: un formulario con **el motivo como campo principal** y obligatorio. Es lo que
  convierte una escalación en trabajo y no en un «pásalo tú»; la interfaz tiene que dejar claro que
  sin ese texto no se escala.
- **La vista doble** (sección 5) es suya: es quien traduce entre el usuario y Desarrollo.
- **Usuarios**: alta, desactivación y reseteo de contraseña, sin salir de la aplicación.

### 4.3 Desarrollo: entender y resolver

Desarrollo llega a un ticket que ya viene filtrado, y **no habla con el usuario**. Su pantalla tiene
que dejar eso claro.

- **Al entrar**: su bandeja de **tickets internos**.
- **Contexto**: lee el **principal** (la conversación con el usuario) y el **interno** donde trabaja.
  La vista doble es exactamente esto.
- **Acciones**: asignarse, **pedir algo a Soporte** («Necesito algo de Soporte»), y **resolver** con
  la explicación de qué se hizo.
- **Devolver**: un botón que dice **«No es un cambio de código»** y devuelve el caso a Soporte con un
  comentario. La interfaz no lo llama «cerrar sin resolver»: dice lo que significa.
- **Lo que no hace**: escribir en el ticket principal, comentar como si fuera Soporte, ni escribirle
  al usuario. La interfaz **no le ofrece** esas acciones, en vez de ofrecérselas y negárselas
  después.

### 4.4 Administrador: dejar el sistema en marcha

Su trabajo es de una vez al principio, no de todos los días.

- **La primera vez, aterriza en la puesta en marcha**: una pantalla que agrupa el **idioma de la
  instalación**, la **apariencia** (tema y color primario) y el **prefijo de numeración**, para
  dejarlo listo de una vez en lugar de ir a buscarlo por tres sitios. Después entra donde quiere, y
  esa misma pantalla sigue en el menú como **Configuración**.
- **Al entrar**: la lista de usuarios, que es lo que va a tocar.
- **Usuarios**: alta (con el rol y el origen), desactivar, reactivar, cambiar el origen y **lanzar un
  reseteo de contraseña**. Toda acción sensible dice lo que va a pasar antes de hacerlo.
- **Numeración**: el prefijo, con un aviso de que **no cambia los números ya emitidos** — es la
  duda que cualquiera tiene antes de tocarlo.
- **Y ahora también cómo se entra** (2026-09-25): la tarjeta del **método de entrada** —los tres en un
  desplegable, con la explicación del que está elegido— y **la configuración que toca según el método
  elegido**, con su botón de **«Probar la conexión»**. **Cada panel sale con su método y ninguno más**
  (decisión del responsable, 2026-09-27): con **«Cuentas de la aplicación»** no se enseña ninguna de
  las dos —no hay nada que configurar—, con la **organización** sale el del directorio, y con
  **Keycloak** el del reino. **Lo que no se está usando no es trabajo que toque hacer**, y las dos
  configuraciones a la vez invitaban a tocar la que no era. **Las opciones de un
  método sin configurar se enseñan pero no se pueden elegir**, y eso vale también para lo que se está
  escribiendo: en cuanto se rellenan el servidor y la base de búsqueda, la opción se enciende sin
  guardar antes. Los campos de los dos secretos **llegan siempre vacíos** y su ayuda dice si hay uno
  puesto: la pantalla nunca enseña un secreto, porque no lo recibe (`docs/modules/settings.md`,
  sección 5.8).
- **Configuración tiene ya todo lo suyo** (2026-09-24 y 2026-09-25): **el nombre de la instalación**
  —el campo, con su ayuda y su botón de guardar; se ve en toda la aplicación en cuanto se guarda—,
  **la marca** —los dos huecos del logo, con su vista previa, subir y volver al de fábrica—, **el color institucional** con **vista previa de cómo
  queda** antes de guardar, el **idioma global de la instalación** —primer control de la pantalla— y el
  **prefijo de numeración** con el aviso de que **no cambia los números ya emitidos**, más el
  **reparto y los avisos** de cada tipo de ticket. Todo lo de esta pantalla viene de un solo endpoint,
  el del módulo `settings`, más los del logo.
- **Los correos se editan en una pantalla del módulo `mail`**, a la que Configuración enlaza: cada
  plantilla con su asunto y su cuerpo, el idioma al lado —se editan los dos—, **los marcadores
  disponibles a la vista** (`{{nombre}}`, `{{enlace}}`, `{{numero}}`…), una **vista previa** y un
  botón de **«enviarme una prueba»**. Un marcador que no exista se avisa **al guardar**, no cuando el
  correo ha salido. `docs/modules/mail.md`, decisión 15.
- **Es una pantalla de otro módulo a propósito**: llamar a `/api/mail/**` desde `settings` rompería
  la regla dura de modularidad (`docs/arquitectura.md`, sección 4), así que la pantalla vive donde
  viven sus datos y Configuración sólo lleva hasta ella.
- **El cuerpo se edita con formato**, no escribiendo etiquetas: negrita, cursiva, lista y enlace
  sobre el texto seleccionado, y el HTML lo genera la aplicación. Ver la decisión 1 de `mail.md`.
- **Ve todo, sin tocar nada**: puede mirar cualquier ticket —también los internos y la vista
  doble— pero **no comenta, no mueve estados, no asigna y no cierra**. Es el precio de ser el papel
  que lo puede todo en configuración: donde no trabaja, mira.

## 5. La vista doble: principal e interno

Pedida a propósito, y es la pieza que hace usable el modelo de dos tickets. **Sólo existe para
Soporte y Desarrollo**, porque el usuario no ve el interno.

- **Cuándo aparece**: cuando el ticket tiene interno y el papel puede verlo. Si no hay interno, no
  hay nada que comparar y no aparece nada.
- **Cómo se controla**: un conmutador con tres posiciones —**Principal**, **Los dos**, **Interno**—
  en la cabecera del detalle. «Los dos» es el valor por defecto cuando existe el interno, porque
  quien está ahí quiere las dos cosas.
- **En PC**: dos columnas, la conversación de cada ticket y su ficha. Es comparar, y comparar es leer
  en paralelo.
- **En móvil y tablet vertical**: no caben dos columnas, así que el mismo conmutador cambia entre
  uno y otro. **Nunca dos columnas estrechas**: es peor que verlas por separado.
- **Cómo no confundirlos**: el interno lleva fondo distinto y su número `INT-…` bien visible. Si los
  dos se ven iguales, alguien va a escribir en el ticket equivocado.
- **Las acciones son del ticket que se está tocando**: en «Los dos» cada columna tiene su ficha y sus
  botones, y la caja de comentario es la de la columna donde estás. Escribir en el principal desde
  Desarrollo no es una opción que se ofrezca (sólo lectura).

## 6. El lenguaje visual

### 6.1 La base

**Tailwind CSS v4 con componentes propios.** Se escribe el aspecto con utilidades y **los componentes
se construyen aquí**, no se traen de una biblioteca.

Lo que eso significa, dicho claro:

- **El aspecto es exactamente el que se decida**, sin pelearse con los estilos de una biblioteca
  ajena ni con su sistema de temas.
- **Los componentes complejos son nuestros**: tabla, diálogo, menú, pestañas, subida de archivos,
  avisos. Cada uno hay que construirlo **y hacerlo accesible** (sección 8). Es el precio, y es un
  precio real: no es lo mismo escribir un botón que un diálogo que atrapa el foco y lo devuelve al
  cerrarse.
- **Inventario cerrado de componentes** (sección 6.3): lo que no esté en esa lista, se discute antes
  de construirlo, en lugar de aparecer por sorpresa.

Calibyou usa Tailwind 3; aquí se usa la **4** (la actual), que define los colores como variables CSS
y encaja mejor con el sistema de temas de la sección 6.2.

**Moderno, sobrio y sin adornos.** La tendencia de estos años —interfaces planas, tipografía grande,
mucho espacio en blanco, bordes suaves, sombras casi inexistentes— coincide con lo que hace falta
aquí: nada que compita con los datos.

| | |
| --- | --- |
| **Tipografía** | Una sola familia de sistema (la que ya usa la máquina), con escala de tamaños fija. Nada de fuentes descargadas: cargan tarde y en una mesa de ayuda no aportan |
| **Espaciado** | Escala de 4 y 8 píxeles, sin valores sueltos |
| **Bordes** | Radio pequeño y bordes finos; el borde separa mejor que la sombra |
| **Sombras** | Sólo en lo que flota de verdad: diálogos y menús |
| **Color** | Neutro para casi todo y **un color de acento** para lo accionable. El color se usa para señalar, no para decorar. Los colores no se escriben a mano en los componentes: salen de los temas (sección 6.2) |
| **Iconos** | Una sola librería, siempre con texto al lado en las acciones importantes |
| **Movimiento** | Casi ninguno: transiciones cortas al abrir un diálogo o plegar el menú, y nada más |

**El color de los estados** (siempre acompañado de su texto, nunca solo):

| Estado | Intención |
| --- | --- |
| `nuevo` | Llama la atención: acaba de llegar |
| `en progreso` | Neutro: alguien lo tiene |
| `en espera` | Ámbar: bloqueado, esperando a alguien |
| `escalado` | Distinto de todo lo demás: está en otro equipo |
| `resuelto` | Verde suave: hecho, pendiente de cerrar |
| `cerrado` | Gris apagado: fuera de la vista |

### 6.2 Los temas: ocho, y sólo cambian los colores

La forma de los controles **no cambia nunca**: un botón es el mismo botón en todos los temas. Lo que
cambia es la paleta. Eso es lo que hace que ocho temas no sean ocho interfaces distintas.

| | Cuántos | Qué son |
| --- | --- | --- |
| **Claros** | 3 | Tres paletas claras, fijas |
| **Oscuros** | 3 | Tres paletas oscuras, fijas |
| **Personalizables** | 1 claro y 1 oscuro | Se les cambia **el color primario**: el institucional, al gusto de quien administra |

- **Cómo se implementa**: los colores son **variables CSS** (`--color-primario`, `--color-fondo`,
  `--color-texto`, los colores de estado…). Un tema es un conjunto de valores de esas variables, y
  ningún componente sabe qué tema está activo. Añadir un tema noveno es escribir nueve líneas.
- **Los dos personalizables son los de por defecto.** Una instalación nueva entra con el tema
  personalizable **claro u oscuro según lo que pida el sistema operativo**, siempre con el color
  institucional: así lo primero que ve todo el mundo es la cara de la casa, sin que nadie tenga que
  configurar nada.
- **El color primario lo fija el administrador**, y sólo él: es el color de la institución, y si cada
  uno lo cambiara dejaría de serlo. Lo que cada persona puede cambiar es **el tema**, eligiendo
  entre los ocho desde la cabecera.
- **Elegir es opcional**: quien no toque nada se queda con el de la instalación. La elección de cada
  persona se recuerda **en su navegador**, que es lo simple: no hace falta tocar la base de datos ni
  añadir una pantalla de perfil.
- La paleta de cada tema tiene que **pasar el contraste** de la sección 8, incluidos los colores de
  estado. Un tema bonito que no se lee no entra.

**Cómo se implementa hoy** (2026-09-23, con lo que hay construido):

- **Un atributo en el elemento `html`**: `data-theme="claro"`, `data-theme="oscuro"` o **ninguno**,
  que es «seguir al sistema». Las paletas se escriben **una sola vez** con la función `light-dark()`
  de CSS y el atributo sólo cambia `color-scheme`: así no hay dos listas de colores que se puedan
  desincronizar, y añadir una paleta nueva son unas líneas, como dice la tabla de arriba.
- **La elección se recuerda en el navegador** (`localStorage`), que es lo simple y lo que ya está
  decidido: no toca la base de datos ni hace falta una pantalla de perfil.
- **El tema lo aplica código nuestro, nunca un `script` escrito en el `index.html`**: la CSP de
  producción es `script-src 'self'`, sin `'unsafe-inline'`, así que un script incrustado **no
  funcionaría en producción** y ensanchar la CSP por esto sería cambiar seguridad por comodidad. Se
  aplica antes de arrancar la aplicación, desde `main.ts`.
- **El conmutador enseña dos opciones: claro y oscuro. Nada más.** **«Automático» no se muestra ni se
  elige** (decisión del responsable, 2026-09-23): seguir al sistema es **no haber elegido**, que es el
  estado de fábrica, y por eso no hace falta una tercera opción que además sería la única forma de
  volver atrás. Mientras nadie toque nada, el conmutador **enseña marcado lo que se está viendo**
  —claro u oscuro según el sistema—, no una opción que no existe; en cuanto alguien elige, se queda
  con lo que eligió y el sistema deja de mandar. Si el sistema cambia con la página abierta y nadie
  ha elegido, la aplicación cambia con él, sin recargar.
- **Las seis paletas fijas** llegan con el armazón completo, cuando se aprueben sus nombres y sus
  colores. Mientras no exista la cabecera del armazón, el conmutador vive **al lado del de idioma**.
- **La paleta de fábrica sigue al sistema operativo** hasta que alguien elija otra cosa, que es la
  decisión 5 del registro.

### 6.4 La marca: el logo y el nombre de la instalación

La marca **es de la instalación**, no de cada persona, y la cambia un Administrador desde
Configuración (`docs/modules/settings.md`, secciones 5 y 5.7). Son **el logo y el nombre**, y se
enseñan juntos y siempre igual:

| Dónde | El logo | El nombre |
| --- | --- | --- |
| **Pantalla de entrada** | Arriba del formulario, centrado, **hasta 128 px en móvil y 160 px de tablet para arriba**: es lo primero que se ve | Debajo del título de la tarjeta, donde está el formulario: es lo primero que se lee |
| **Primer arranque** | Junto a la cabecera, **80 px de alto**: ocupa el mismo bloque visual que el título y su explicación sin competir con el contenido del paso | El nombre todavía no se enseña aquí: se está configurando |
| **Menú lateral** | Arriba, junto al nombre, **hasta 32 px de alto** | Arriba, junto al logo |
| **Pestaña del navegador** | — | Es el título de la pestaña, en todas las pantallas |
| **Configuración** | Los dos huecos con su vista previa y los botones de subir y volver al de fábrica | El campo, con su ayuda: si se deja vacío vuelve el nombre de fábrica |

**Y con la marca va la versión del sistema**, que no se configura: la escribe el código
(`docs/modules/settings.md`, sección 5.9).

| Dónde | La versión |
| --- | --- |
| **Menú lateral** | En la fila de salir, **alineada a la derecha**, en texto pequeño y apagado |
| **Pantalla de entrada** | **Debajo de los controles**, centrada y en texto pequeño y apagado: es el pie de la pantalla |
| **En ningún otro sitio** | No va en los correos, ni en el título de la pestaña, ni en las pantallas de producto |

- **Se escribe `v1.0.0`, con la `v`**: es como se lee una versión, y sin ella un `1.0.0` suelto
  parece cualquier otra cosa.
- **No es un enlace ni un botón**: no lleva a ningún sitio. Se lee y ya.
- **En la pantalla de entrada sí tiene sentido** (decisión del responsable, 2026-09-25): es donde
  alguien que no puede entrar mira para decir qué versión está usando, y la pantalla de entrada es lo
  único que ve.
- **No se enseña la fecha de compilación**: cambiaría en cada compilación sin decir nada nuevo, y dos
  instalaciones con el mismo código dirían cosas distintas.

- **El nombre no lleva el de la pantalla**: la pestaña se llama como la instalación y nada más (todas
  las rutas traían el mismo texto, así que no decían dónde estabas).
- **Se cuenta en caracteres y no pasa de 60**, que es lo que cabe en el menú y en la tarjeta de la
  entrada; el campo lo corta al escribir y el backend lo comprueba igual.

**El sistema tiene que servir para cualquier logo**, y eso se resuelve en cómo se enseña, no pidiendo
un archivo concreto:

- **No se recorta y no se estira**: se enseña entero, con su proporción, dentro del alto máximo.
- **Va en un recuadro de esquinas redondeadas y con un borde del tema.** Así un logo con **fondo
  opaco** —el de fábrica, por ejemplo, es un cuadrado negro— se ve como una pieza deliberada, y uno
  **con transparencia** se integra en el fondo. Ninguno de los dos se da por supuesto.
- **En la barra de controles, el selector de tema va sin etiqueta a la vista** (decisión del
  responsable, 2026-09-24): el desplegable ya enseña el tema que está puesto, así que la palabra
  «Tema» sólo ocupaba sitio. **El nombre accesible sigue ahí**, oculto a la vista: un lector de
  pantalla tiene que poder decir de qué es el control (sección 8).
- **Hay dos huecos, claro y oscuro, y los dos son opcionales** (sección 6.2 y
  `docs/modules/settings.md`, sección 5.1): si sólo hay uno, se usa en los ocho temas. **Y los dos
  logos de fábrica también son dos, uno por tema** (2026-09-26), así que una instalación sin logo
  propio se lee igual de bien en claro que en oscuro.
- **En los correos no va**: los correos son HTML sin imágenes (`docs/modules/mail.md`).

### 6.3 Inventario de componentes propios

Lo que hay que construir, y nada más:

`botón` (**cuatro formas** —principal, secundaria, peligrosa y la verde de la confirmación— **y puede
ser sólo un icono**, que entonces necesita su nombre accesible; **el icono es uno de los del proyecto**
—`app-icono`, `frontend/src/app/shared/components/icono.ts`, svg propios, nunca un carácter ni una
biblioteca—) · `campo de texto` · `área de texto` · `selector` · `casilla` · `etiqueta de estado` ·
`tabla` (con orden y paginación) · `tarjeta` (la lista en móvil) · `diálogo` · `menú` ·
`conmutador segmentado` (la vista doble y los filtros) · **`editor de texto con formato`, con los
adjuntos dentro** · `adjunto` (con vista previa) · `comentario` (con marcas de editado y eliminado) ·
`aviso` (éxito, error, información) · **`toast`** (resultado fijo en la ventana) · `estado vacío` ·
`indicador de carga` · `menú lateral` (con sus tres zonas) · `barra de móvil`.

Veinte piezas. Eran veinte antes de que **`subida de archivos con arrastrar y soltar` se retirara**
(2026-09-26); al integrarla en el editor quedaron diecinueve, y el toast aprobado el 2026-10-06 lleva
el inventario de nuevo a veinte. Adjuntar dejó de ser un recuadro aparte —ahora **se adjunta dentro
del texto**, con el editor— y tener dos sitios donde se hace lo mismo era el ruido que este producto
evita. Lo que hacía —arrastrar, elegir y filtrar por extensión con su aviso— **lo hace el editor**,
con las tres puertas de la sección 6.5.

Cada una se construye **una vez**, con sus estados (normal, hover, foco, deshabilitado, cargando) y su
comportamiento de teclado.

**Hechas hoy**: el botón, el campo de texto, el **área de texto**, el selector, el conmutador
segmentado, la tarjeta, el **aviso**, el **toast**, el **diálogo**, la **etiqueta de estado**, el **adjunto con su
vista previa**, el **editor con formato —con los adjuntos dentro, sección 6.5—**, el menú lateral con
sus tres zonas y la barra de móvil. Lo que falta —el comentario como pieza suelta y el indicador de
carga— llega con las pantallas que lo usan.

- **El diálogo es un `<dialog>` nativo** (2026-09-24): se cierra con **Escape**, atrapa el foco y lo
  **devuelve a donde estaba** sin que haya que programarlo. Es lo que hace el alta de una cuenta.
- **La tarjeta elige el nivel de su título** (2026-09-24): `h1` cuando la tarjeta *es* la pantalla —las
  de la sesión— y `h2` cuando va debajo del título de una pantalla de producto. Sin eso, la pantalla
  de usuarios tenía tres `h1`, y una página así no se navega con un lector de pantalla (sección 8).
- **Adjuntar tiene tres puertas y una sola regla** (2026-09-25, y desde el 2026-09-26 con el editor):
  **pegar**, **arrastrar** y **el botón**, y las tres **meten el archivo donde está el cursor**. Se
  comprueba lo mismo —la lista cerrada de extensiones— y se avisa igual de lo que no vale
  (`docs/modules/tickets.md`, secciones 2.3 y 4).
- **El selector puede enseñar una opción que no se puede elegir** (2026-09-24), con su texto de ayuda
  debajo: es como se enseñan hoy los orígenes de directorio, que están previstos y todavía no
  funcionan (sección 3.6).

### 6.5 El cuadro de escribir: un editor con formato y los adjuntos dentro

**Escribir un ticket o un comentario no es rellenar un campo de texto**: lo que se escribe lleva
formato y **los archivos que se adjuntan se ven dentro del texto**, en el sitio donde se ponen
(`docs/modules/tickets.md`, sección 2.3). Un cuadro de texto normal no puede llevar una imagen dentro,
así que **el cuadro de escribir es un editor con formato**.

- **La barra del editor**: negrita, cursiva, subrayado, tachado, lista con viñetas, lista numerada,
  enlace y **Adjuntar**. Los botones actúan **sobre lo que está seleccionado** —es la misma regla del
  editor de los correos, sección 3.8— y **Adjuntar mete el archivo donde está el cursor**.
- **Tres puertas para adjuntar, y las tres meten el archivo en su sitio**: **pegar** (`Ctrl+V`),
  **arrastrarlo encima del editor** y **el botón**. Se acepta lo que esté en la lista cerrada de
  extensiones, y lo que no está se rechaza con el mismo aviso de siempre.
- **Al pegar texto, el texto entra sin formato** (2026-09-26): lo que se pega se queda como texto plano
  y **se respetan los saltos de línea**. Pegar el HTML de un correo o de Word traería etiquetas que no
  están en la lista blanca, y el comentario se rechazaría entero por algo que la persona no ha escrito
  —un `style`, un `<font>`—, que es la peor forma de perder un texto. Un **archivo** pegado (una
  captura, lo que se copia del explorador) sí entra, y entra como adjunto. Es la misma regla de
  `docs/modules/tickets.md`, sección 2.3.
- **Un enlace se pide con su dirección y se comprueba antes de ponerlo**: tiene que empezar por `http`,
  `https` o `mailto`. Lo que no, no se pone y se dice por qué, en vez de rechazar el comentario al
  guardarlo (es la misma regla que aplica el backend).
- **Se ve lo que va a salir**: la imagen y el vídeo **se ven de verdad** mientras se escribe, no como
  un texto que los representa. Un PDF o un Word se ven como **su enlace**, con su nombre.
- **Llamar a alguien se escribe en el comentario**: el editor tiene un botón **«Etiquetar»** —y
  escribir `@` abre lo mismo—, que enseña **la lista de técnicos y desarrolladores activos** y mete el
  nombre elegido **donde está el cursor**, resaltado como una **mención**. Quien lo lee ve a quién se
  está llamando, y esa persona pasa a **observar** el ticket: aparece en su «Observo» y en la lista de
  observadores de la ficha. **Sólo Soporte y Desarrollo pueden etiquetar** (el usuario no ve la lista de
  técnicos), y a quien se etiqueta **se le avisa por correo**. En la ficha, cada observador lleva su
  aspa para quitarlo, y quitarlo **queda en la línea de tiempo**.
- **Con tope, y el clic los abre grandes**: en el texto —escribiendo y leyendo— la imagen y el vídeo
  se ven **como mucho a 480 × 360** (decisión 55, del 2026-09-26), sin deformarse: lo que no puede es
  que una captura a pantalla completa se coma la conversación. **El tope es `min(100%, 480px)` y no
  `480px`**: un `max-width` sólo limita, así que con el valor fijo una imagen de 480 **se salía de la
  pantalla de un móvil de 412 px** y ensanchaba el documento entero —lo encontró la capa de
  Playwright, porque el clic en un botón de la misma tarjeta empezó a fallar con el scroll
  horizontal—. Ahora se encoge con la pantalla en vez de desbordarla. **Al pulsarlos se abren en el visor**,
  más grande, con **Descargar** y **Abrir en una pestaña**. Para eso el marco de lo que se ve es **un
  `button` de verdad** —se llega con el tabulador, se pulsa con Enter y dice qué es—, y el vídeo lleva
  encima **su botón de reproducir**: una imagen que se pulsa y no dice que se pulsa es una imagen que
  nadie descubre con un lector de pantalla (sección 8).
- **El vídeo se reproduce en el visor**, no en el texto: en la conversación se ve **una miniatura**
  —sin controles— y al pulsarla se abre el visor, que es donde se ve grande, se reproduce y se
  descarga. **Mientras se escribe sí lleva controles**, porque ahí no hay visor y lo que se está
  haciendo es comprobar lo que se acaba de adjuntar.
- **Los controles del reproductor los pone la pantalla**: lo que se guarda es
  `<video data-adjunto="grabacion.mp4">`, sin `controls`, porque **ese atributo no está en la lista
  blanca del backend** (`backend/modules/tickets/services/body.go`): con él, el texto entero se
  rechazaría. **El tamaño tampoco se guarda**: es cómo se pinta. Queda dicho aquí porque **el hallazgo
  fue justo el contrario**: el ejemplo de `docs/modules/tickets.md`, sección 2.3, escribía
  `<video data-adjunto="x.mp4" controls>`, que es un texto que el backend habría rechazado; el
  documento se corrigió en el mismo trabajo.
- **Un archivo que se nombró y no subió deja su hueco marcado**: si una subida falla después de
  publicar el comentario (decisión 44 de `tickets.md`), el sitio donde iba el archivo **se ve**, con su
  nombre y un aviso, en vez de quedar un hueco mudo que parezca un fallo de la pantalla. Cuando el
  archivo sube, el hueco se convierte en la imagen.
- **Quitar un adjunto es quitarlo del texto**: se borra como se borra cualquier cosa que se ha
  pegado. Y **un archivo que no está en el texto no se sube**, así que no hay que acordarse de
  quitarlo en dos sitios: hay uno.
- **El recuadro de archivos aparte desaparece** de estos formularios (`app-archivos` se retira, y el
  inventario de la sección 6.3 lo refleja): tener dos sitios donde se adjunta lo mismo era justo el
  ruido que este producto evita.
- **Lo que se escribe se cuenta en texto, no en HTML**: las etiquetas y las direcciones de los
  adjuntos son cosa del editor y de la pantalla, y **nadie las ve**. Eso incluye la búsqueda de la
  bandeja, que **quita las etiquetas antes de buscar** para que buscar «p» no encuentre todos los
  tickets con una imagen.
- **Y lo que se manda lleva sólo lo que el backend admite**: un adjunto sale como su `data-adjunto` y
  **nada más** —ni la dirección `blob:` de la vista previa, ni su `class`, ni un `style`—, porque el
  saneador del backend **rechaza el texto entero** con `422 tickets.body.notAllowed` en cuanto aparece
  algo que no está en su lista. Cuando eso pasa, la pantalla lo dice en castellano («El texto lleva algo
  que no se admite…»), no enseña la clave. **Los controles del vídeo tampoco se guardan**: los pone la
  pantalla al pintarlo.
- **Al leer, lo mismo**: la conversación pinta el texto con formato y, en su sitio, la imagen o el
  reproductor del vídeo; el PDF y los demás archivos, como enlaces. **Pulsar una imagen o un PDF abre
  el visor** —el modal, más grande, con **Descargar** y **Abrir en una pestaña**, decisión 48 de
  `tickets.md`—, y pulsar un Word o un Excel **lo descarga**.

## 7. Multi-dispositivo

**La prioridad es el PC**: es donde se trabaja, donde hay dos columnas y donde están las bandejas
completas. Tablet y móvil tienen que **funcionar de verdad**, no ser una promesa.

| | Móvil (< 768) | Tablet (768–1200) | PC (> 1200) |
| --- | --- | --- | --- |
| **Menú** | Botón en la cabecera | Plegado, se abre si se quiere | Fijo, plegable |
| **Listas** | Tarjetas | Tabla reducida | Tabla completa |
| **Detalle** | Una columna; la ficha baja debajo | Una columna más ancha | Conversación + ficha a la derecha |
| **Vista doble** | Conmutador, de uno en uno | Conmutador, de uno en uno | Las dos columnas |
| **Administración** | Consultar y acciones simples | Todo | Todo |

- **Objetivos táctiles de 44 px** como mínimo en móvil: un botón que no se acierta con el dedo es un
  botón que no existe.
- **Nada de desplazamiento horizontal**: si una tabla no cabe, se convierte en tarjetas, no se
  arrastra de lado.
- **Lo que se hace en móvil** (decidido, sección 11): **todo menos la administración y la vista
  doble**, que son lo que se deja para el PC.

## 8. Accesibilidad y confianza

- **Contraste suficiente** en texto y en los colores de estado, en claro y en oscuro.
- **Todo se puede usar con teclado**: foco visible siempre, orden de tabulación coherente y nada que
  sólo funcione con el ratón.
- **Nunca sólo color** para decir algo: el estado lleva su texto y los avisos llevan su icono.
- **Etiquetas de verdad** en los campos, no textos de ejemplo que desaparecen al escribir.
- **Los errores dicen qué pasó y qué hacer**, en el idioma de quien los lee. Un «Error 422» no es un
  mensaje. El backend manda una **clave** (`auth.invalidCredentials`) y la interfaz la traduce: por eso
  no basta con tener un diccionario de pantallas —**cada clave de error necesita su texto en los dos
  idiomas**, y una clave sin texto es un error sin mensaje.
- **Los diálogos cierran con Escape** y no roban el foco sin devolverlo.
- **Los adjuntos se anuncian** con su nombre y su tamaño antes de descargarlos.

**Como los componentes son nuestros, esto no viene de serie.** Lo que hay que hacer bien una vez y
reutilizar:

| Componente | Lo que no se puede olvidar |
| --- | --- |
| **Diálogo** | Foco dentro al abrirse, devuelto a donde estaba al cerrarse, Escape para cerrar y el fondo sin poder tocarse |
| **Menú** | Flechas para moverse, Enter para elegir, Escape para cerrar |
| **Conmutador segmentado** | Se anuncia como grupo de opciones, no como botones sueltos |
| **Tabla** | Encabezados de verdad, orden anunciado y navegación con teclado por las filas |
| **Subida de archivos** | Se puede hacer sin ratón, y cada archivo dice qué pasó con él |
| **Avisos** | Se anuncian solos a quien usa lector de pantalla, y no roban el foco |

## 9. Los estados que no son pantallas

Casi nadie diseña esto y es lo que más se sufre:

- **Lista vacía**: dice por qué está vacía y qué hacer. «No tienes tickets todavía» con el botón de
  crear al lado, no una tabla en blanco.
- **Bandeja vacía**: «No hay nada pendiente» — y que sea verdad, sin inventarse un mensaje motivador.
- **Cargando**: un indicador discreto y sólo si tarda; nada de pantallas de carga para medio segundo.
- **Sin permiso**: si alguien llega a una pantalla que no le toca, se le dice, no se le enseña en
  blanco.
- **Algo falló**: qué se intentaba, qué pasó y qué puede hacer. Y el camino de vuelta siempre visible.
- **Sin conexión o servidor caído**: un aviso claro, sin tecnicismos.

## 10. Detalles pequeños que se notan mucho

- **El número del ticket, siempre visible y copiable.**
- **La fecha de última actualización** en la lista y en el detalle: lo que lleva días parado se ve.
- **Sin refrescos automáticos ni contadores en vivo** (no hay tiempo real, `docs/modules/tickets.md`): la
  lista se actualiza al entrar y con un botón de refrescar. Un contador que no se actualiza solo es
  peor que no tenerlo.
- **Los adjuntos de imagen y PDF se previsualizan** dentro del ticket; el resto se descarga
  (`docs/modules/tickets.md`, sección 4).
- **Los correos son HTML, con una versión de texto automática** (`docs/modules/mail.md`): mismo tono
  que la interfaz, sin jerga y con estilos en línea, sin imágenes externas ni fuentes de terceros.

## 11. Lo que se decidió al repasar este documento

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **La base visual** | **Tailwind CSS v4 con componentes propios** (20 piezas, sección 6.3), no una biblioteca |
| 2 | **La vista doble** | Dos columnas en PC con **«Los dos» por defecto** cuando existe el interno; conmutador en móvil |
| 3 | **Qué se hace desde el móvil** | **Todo menos administración y vista doble** |
| 4 | **Los temas** | **Ocho: 3 claros + 3 oscuros fijos, más 1 claro y 1 oscuro personalizables**, que son los de por defecto. El **color primario lo fija el administrador**; cada persona elige tema. Sólo cambian colores, nunca la forma |
| 5 | **Modo claro y oscuro** | Sí desde la 1.0.0, siguiendo al sistema hasta que alguien elija |
| 6 | **Accesibilidad** | **Requisito**: contraste, teclado y foco en todas las pantallas, con AA como referencia, y comprobado al construir cada componente |
| 7 | **El tema, hoy** | Claro y oscuro (los dos personalizables de fábrica), con el mecanismo listo para las seis paletas fijas. De fábrica sigue al sistema, que es lo mismo que **no haber elegido** |
| 10 | **«Automático» no se elige ni se muestra** | Decisión del responsable (2026-09-23): el conmutador tiene dos opciones y mientras nadie elija **enseña marcado lo que se está viendo**. Seguir al sistema es el estado de fábrica, no una tercera opción |
| 11 | **El sistema manda en vivo** | Si cambia de claro a oscuro con la página abierta y nadie ha elegido, la aplicación cambia con él sin recargar |
| 8 | **Dónde vive el conmutador mientras no hay cabecera** | Al lado del de idioma, en la entrada y en el inicio: se muda a los controles de la cabecera cuando exista el armazón |
| 9 | **El tema no se aplica con un script incrustado** | La CSP de producción es `script-src 'self'`: se aplica desde `main.ts`, antes de arrancar |
| 13 | **Un comentario sin texto se lee por sus adjuntos** | **Elección del responsable, 2026-09-25**, al decidir que un comentario pueda ser sólo un archivo: en la conversación aparece con su autor, su fecha y sus archivos, **sin un hueco en blanco** que parezca un fallo de la pantalla. Y **si un archivo no sube después de publicar el comentario**, se avisa de cuál falta y se puede reintentar sin volver a escribirlo (`docs/modules/tickets.md`, decisiones 41 y 44) |
| 12 | **La versión del sistema** | **Elección del responsable, 2026-09-25**, y **no la que yo recomendaba**: se lee en la fila de salir del menú lateral, **alineada a la derecha**, y **también en el pie de la pantalla de entrada**. Es un dato, no un botón ni una fila propia, y **plegado no se enseña**. Yo proponía sólo el menú; él decidió que también se vea antes de entrar, que es donde sirve para decir qué versión se está mirando |

## 12. Qué habilita este documento

Con `docs/interfaz-y-experiencia.md` aprobado, la documentación cubre **qué se construye, cómo se
comporta y cómo se ve**. El código puede empezar por `auth` (`docs/modules/tickets.md`), y sus pantallas
—entrar, recuperar contraseña— se diseñan ya con estas reglas, en lugar de retocarlas después.

## 13. Navegación de regreso a la izquierda

### 13.1 Hallazgo e inventario

La regla aprobada el 2026-09-29 ya dice que los botones de volver van a la izquierda y son una
flecha de sólo icono con nombre accesible. El inventario de las vistas con ese control de cabecera
encontró:

| Vista | Destino | Posición encontrada | Cumplía la regla |
| --- | --- | --- | --- |
| `/mail` | Configuración | A la derecha del título | No |
| `/tickets/new` | Lista de tickets | A la izquierda del título | Sí |
| `/tickets/:número` | Lista de tickets | A la izquierda del número | Sí |
| `/users/:id` | Lista de usuarios | A la izquierda del nombre | Sí |

El comentario del propio `correos-page.html` también afirmaba que la flecha estaba a la izquierda
del título, pero la estructura `justify-between` la colocaba después del título y al extremo
derecho. Era una incoherencia entre el documento, el comentario y lo que se pintaba.

### 13.2 Regla

Todo control que **regresa desde una vista secundaria y vive en su cabecera** aparece como primer
elemento de esa cabecera, inmediatamente antes del título o identificador principal. Es una flecha
de sólo icono, conserva un nombre accesible y mantiene su destino actual.

Esta regla se aplica a los cuatro controles inventariados. Tres ya cumplían; la implementación
reordena la cabecera de `/mail` para colocar su enlace a Configuración antes del bloque de título y
descripción. No cambia el destino, el texto accesible, el icono ni el tamaño del control.

Quedan fuera:

- los enlaces que forman parte del flujo de una pantalla, como **Volver a iniciar sesión**;
- el botón **Volver a mi trabajo** de la vista sin permiso;
- acciones con otra finalidad que usan la palabra «volver», como restaurar un logo o un correo;
- añadir botones nuevos a vistas que hoy no tienen navegación de regreso;
- el backend, las rutas y los permisos.

### 13.3 Documentos y código

- Este documento define la regla general y registra el hallazgo.
- `docs/README.md` y `AGENTS.md` reflejan el estado final de la enmienda.
- `frontend/src/app/modules/mail/correos-page.html` coloca la flecha antes del bloque de título y
  descripción.
- La prueba E2E del editor de correos comprueba que la flecha es el primer elemento de la cabecera,
  antes del título, y que sigue regresando a Configuración. La suite la ejecuta en PC y móvil.

`docs/modules/mail.md` no cambia: describe las funciones del editor y no fija la composición de su
cabecera. Además, su comportamiento funcional no cambia.

### 13.4 Criterios de aceptación

1. En `/mail`, la flecha aparece inmediatamente a la izquierda del título **Los correos**.
2. En anchos de PC y móvil, el control permanece antes del título cuando la cabecera se reparte en
   más de una línea.
3. Su nombre accesible sigue siendo **Volver / Back** y su destino sigue siendo `/settings`.
4. Los tres controles que ya cumplen conservan su posición y su destino.
5. Los enlaces y acciones excluidos no cambian.

### 13.5 Decisión registrada

El responsable confirmó el 2026-10-05 la recomendación: unificar sólo los controles de navegación
de cabecera como icono a la izquierda del título. Los enlaces de flujo dentro del contenido quedan
fuera.

### 13.6 Registro del repaso

El responsable confirmó el 2026-10-05 que:

1. la flecha quedaría como primer elemento de la cabecera;
2. el título y su descripción permanecerían juntos, en un bloque a la derecha de la flecha;
3. no se añade ninguna excepción adicional a la regla ni se amplía el alcance.

El repaso queda cerrado sin decisiones abiertas. El responsable aprobó explícitamente la propuesta
con «Sí». La implementación respeta el alcance aprobado y el documento vuelve a `as-built`.

### 13.7 Verificación

- Las 216 pruebas unitarias del frontend pasan.
- Los cuatro casos E2E del editor de correos pasan: dos en PC y dos en móvil.
- La prueba comprueba que la flecha es el primer elemento de la cabecera, que su posición horizontal
  queda antes del título y que al pulsarla se regresa a `/settings`.

## 14. Región seleccionada visible en Configuración

### 14.1 Hallazgo

La tarjeta **Región horaria y dirección pública** de `/settings` ofrece el mismo buscador y la misma
lista desplazable que `/setup`. La opción elegida sólo se reconoce por el fondo y el peso de texto
dentro de esa lista; si queda fuera de la parte visible, la pantalla no nombra de forma estable la
zona que se guardará. La hora y el desfase permiten comprobarla, pero no dicen cuál es.

### 14.2 Comportamiento

Se reutiliza exactamente el patrón aprobado para `/setup`:

- campo de sólo lectura antes del buscador;
- etiqueta **Selected time zone / Región horaria seleccionada** según el idioma de la interfaz;
- identificador IANA exacto como valor;
- zona guardada al cargar la pantalla y actualización inmediata al elegir otra;
- texto seleccionable y copiable, sin edición directa;
- hora y desfase conservados debajo de la lista.

El valor del campo es el borrador actual del formulario. Elegir una zona lo cambia y habilita el
guardado existente; guardar, recibir un error o volver a cargar siguen las reglas actuales de
Configuración. El buscador, la lista, el resaltado, la dirección pública y el botón de guardar no
cambian.

### 14.3 Alcance

- `docs/modules/settings.md` define el comportamiento dentro de la tarjeta de región.
- Este documento define su presentación coherente con `/setup`.
- `docs/README.md` y `AGENTS.md` reflejan el estado final de la enmienda.
- `frontend/src/app/core/pages/settings-page.html` muestra el campo y las pruebas existentes de
  Configuración comprueban su comportamiento.
- Se reutilizan los textos ES/EN actuales; no se añaden claves de traducción.

Quedan fuera la API, la base de datos, la validación de zonas, `/setup`, la dirección pública y las
demás tarjetas de Configuración.

### 14.4 Criterios de aceptación

1. Al cargar `/settings`, el campo muestra la zona guardada.
2. Elegir `America/Guayaquil` actualiza el campo inmediatamente antes de guardar.
3. El campo es de sólo lectura y permite seleccionar y copiar el identificador.
4. La búsqueda, la lista, el resaltado, la hora y el desfase conservan su comportamiento.
5. Las pruebas unitarias comprueban el valor cargado, el cambio inmediato y el atributo de sólo
   lectura; Playwright comprueba la presencia del campo en PC y móvil.

### 14.5 Decisión registrada

El responsable confirmó el 2026-10-05 reutilizar exactamente el patrón de `/setup`, sin cambios en
el guardado, la API ni la base de datos.

### 14.6 Registro del repaso

El responsable confirmó el 2026-10-05 que `/settings` usará el mismo identificador técnico, la
misma apariencia y el mismo comportamiento de `/setup`, sin diferencias adicionales.

El repaso queda cerrado sin decisiones abiertas. El responsable aprobó explícitamente las dos
propuestas con «Sí». La implementación respeta el alcance aprobado y ambos documentos vuelven a
`as-built`.

### 14.7 Verificación

- Las 216 pruebas unitarias del frontend pasan; las de Configuración comprueban el valor inicial,
  el cambio inmediato y el atributo de sólo lectura.
- Los cuatro recorridos E2E afectados pasan: dos en PC y dos en móvil.
- Playwright comprueba el valor guardado, `America/Guayaquil` después de elegir, el atributo de sólo
  lectura y que el campo pertenece a la tarjeta de región y dirección.

## 15. Resultados de acciones en mensajes toast

### 15.1 Hallazgo

`/settings` y `/setup` tienen un único mensaje de pantalla colocado antes de su contenido. Guardar
una tarjeta o probar una conexión desde una sección inferior actualiza ese mensaje, pero la ventana
no vuelve al principio. La acción sí responde y el texto sí existe; queda fuera de la parte visible.
Esto explica por qué una prueba de conexión fallida parecía no mostrar ningún resultado.

### 15.2 Componente y comportamiento

Se añade un componente compartido de **toast** para los resultados transitorios de una acción:

- queda fijo en la ventana, arriba a la derecha en PC y con el ancho disponible, dejando margen, en
  móvil;
- sólo se muestra uno a la vez y el resultado más reciente sustituye al anterior;
- éxito e información desaparecen a los **5 segundos**;
- error y atención permanecen hasta que la persona los cierre o una acción posterior los sustituya;
- todos incluyen un botón de cierre con nombre accesible traducido;
- éxito e información se anuncian como estado; error y atención se anuncian como alerta, sin mover
  el foco ni desplazar el contenido;
- mantiene las cuatro formas visuales del aviso existente —éxito, error, información y atención— y
  usa los colores, iconos y contraste del tema activo.

El temporizador empieza de nuevo cuando llega un resultado distinto. Destruir la pantalla cancela
el temporizador pendiente, para que un mensaje antiguo no actúe después de navegar.

### 15.3 Qué es toast y qué sigue dentro de la página

Usan toast los resultados de acciones asíncronas que hoy alimentan el mensaje general de
`/settings` o `/setup`: guardar, subir o restaurar, avanzar en el asistente y probar una conexión.
También lo usan sus fallos de carga o de validación devueltos por la API cuando no pertenecen a un
campo concreto.

Permanecen dentro de su sección los avisos que describen un estado o una consecuencia mientras esa
condición exista, entre ellos la dirección pública sin HTTPS, la explicación de la cuenta de fábrica
y los avisos informativos propios de un paso. Los errores asociados a un campo conservan su lugar
junto al campo si una pantalla los presenta así.

### 15.4 Alcance

- Este documento define el componente compartido y su accesibilidad.
- `docs/modules/settings.md` define su uso en `/settings`.
- `docs/primer-arranque.md` define su uso en `/setup`.
- `docs/README.md` y `AGENTS.md` reflejan el estado de la enmienda.
- La implementación vive en `frontend/src/app/shared/components/**` y en las dos páginas afectadas.

Quedan fuera el backend, sus respuestas y códigos, la persistencia, las reglas de validación, los
mensajes de las demás pantallas y un servicio global de notificaciones para toda la aplicación.

### 15.5 Criterios de aceptación

1. Probar una conexión desde una tarjeta o paso inferior muestra su éxito o error dentro de la parte
   visible de la ventana, sin desplazarla al principio.
2. Guardar, subir, restaurar y avanzar muestran sus resultados por el mismo componente.
3. Éxito e información desaparecen después de 5 segundos; error y atención continúan hasta cerrarse
   o ser sustituidos.
4. Un segundo resultado sustituye al primero y reinicia el temporizador que corresponda.
5. El botón de cierre funciona con teclado, tiene nombre accesible en español e inglés y el mensaje
   conserva el anuncio semántico adecuado.
6. Los avisos persistentes de contexto siguen dentro de su sección.
7. Las pruebas unitarias cubren sustitución, cierre y duración; Playwright comprueba la posición
   visible de los resultados de conexión en `/settings` y `/setup`, en PC y móvil.

### 15.6 Decisión registrada

El responsable confirmó el 2026-10-06 proponer el componente compartido para `/settings` y
`/setup`: fijo arriba a la derecha en PC y adaptado al ancho móvil, un mensaje a la vez, éxito con
cierre automático a los 5 segundos, errores persistentes, cierre manual y avisos de contexto todavía
integrados en la página.

### 15.7 Registro del repaso

El responsable confirmó el 2026-10-06 que:

1. todos los resultados transitorios descritos de `/settings` y `/setup` usan el toast;
2. no hay cola: el resultado nuevo sustituye al anterior;
3. sólo éxito e información desaparecen automáticamente después de 5 segundos; error y atención
   permanecen hasta cerrarse o ser sustituidos.

El repaso queda cerrado sin decisiones abiertas. El responsable aprobó explícitamente los tres
documentos con «Apruebo» el 2026-10-06. La implementación respeta el alcance aprobado.

### 15.8 Implementación y verificación

- `app-toast` envuelve el aviso compartido, permanece fijo en la ventana y emite el cierre manual o
  automático; al destruirse cancela su temporizador.
- `app-aviso` admite un botón de cierre opcional. Los avisos integrados no lo reciben y conservan su
  presentación; error y atención usan `alert`, éxito e información usan `status`.
- Las 220 pruebas unitarias del frontend pasan; cuatro casos nuevos comprueban cierre, semántica,
  duración, persistencia y reinicio del temporizador.
- La compilación de producción termina correctamente.
- Los 20 casos Playwright seleccionados pasan en PC y móvil; 4 se omiten porque los servicios
  opcionales de directorio e IA no estaban levantados. La preparación de la instalación vacía
  comprobó también el toast de `/setup` en ambos anchos.

## 16. Implementación: idioma global y configuración guiada de IA

### 16.1 El idioma es el primer control

En `/setup` y `/settings`, Español/English aparece antes que cualquier otro dato. En el asistente
cambia todos los textos al elegirlo. En Configuración previsualiza la pantalla, pero una banda
explica que el cambio global todavía no se ha aplicado mientras se traducen y revisan correos.

Desaparecen todos los selectores personales de idioma: entrada, controles del armazón, perfil, alta
y ficha de usuario. El espacio liberado en el menú no se rellena con otro control. El tema continúa
siendo personal.

### 16.2 Selector de modalidad de IA

La tarjeta y el paso usan tres opciones grandes con nombre y explicación: **En este servidor**,
**En otro servidor** y **Proveedor externo**. Sólo se muestran los campos de la opción elegida. La
modalidad externa incluye un aviso persistente con los datos que saldrán y una casilla de aceptación
antes del botón **Probar y activar**.

Para proveedor se muestran OpenAI, Claude, DeepSeek y Compatible con OpenAI. El modelo es un
desplegable de sugerencias con **Otro modelo…**; al elegirlo aparece el identificador libre. Los
secretos usan campo de contraseña, indican si existe uno guardado y nunca muestran puntos que puedan
confundirse con su valor real.

### 16.3 Catálogo local y progreso

Cada modelo muestra nombre, tamaño de descarga, RAM recomendada y estado: no instalado, descargando,
instalado, activo o error. El actual 1.5B aparece primero y marcado como recomendado. Antes de una
descarga se muestran disco y RAM disponibles y se bloquea con una explicación si no alcanzan.

La ficha enlaza la fuente oficial y muestra licencia y checksum. Antes de la primera descarga de una
versión se abre la licencia y se pide una aceptación explícita; la pantalla dice quién la aceptó y
cuándo. Un cambio de versión, licencia o checksum vuelve a pedirla.

La descarga tiene barra de progreso, bytes descargados/total y acción de cancelar. Recargar conserva
el avance real que devuelve el servidor. Activar enseña las etapas «cargando», «probando» y «activo».
Los modelos inactivos tienen **Eliminar** con confirmación; el activo nunca ofrece esa acción.

En móvil las fichas se apilan, la barra ocupa el ancho y las acciones quedan debajo. Ningún estado
depende sólo del color; progreso y resultado se anuncian de forma accesible. Los resultados breves
usan el toast existente, mientras descarga, requisitos y estado activo permanecen dentro de la
tarjeta.

### 16.4 Revisión del cambio de idioma

La revisión presenta las once plantillas como lista de pendientes. Al abrir una se ven origen y
borrador lado a lado en PC y apilados en móvil, con asunto, cuerpo y marcadores. Si ya existe un
destino personalizado, aparece como tercera versión y el Administrador elige conservarlo, usar el
borrador o combinarlos en el editor. No se reemplaza en silencio. **Aplicar idioma** sólo se habilita
cuando las once están válidas; cancelar descarta los borradores y vuelve al idioma persistido.

Tras confirmar, la interfaz completa adopta el nuevo idioma. Los resúmenes anteriores se conservan y
siguen visibles en el idioma en que fueron redactados. Los nuevos y los que se vuelvan a pedir se
generan en el nuevo idioma global; no existe regeneración masiva al cambiarlo.

La advertencia económica aparece únicamente antes de traducir las plantillas con un proveedor
externo y muestra solicitudes, tokens y costo estimados. Rechazarla abre la misma revisión con los
destinos editables manualmente. La edición manual permanece disponible siempre desde Configuración.

Las sesiones que ya estaban abiertas reciben idioma y versión de configuración en la siguiente
respuesta autenticada. El interceptor actualiza la señal global y repinta textos y fechas sin
cerrar sesión ni consultar periódicamente. Hasta que esa pestaña haga otra petición conserva la
vista que ya tenía.

### 16.5 Criterios de aceptación propuestos

Playwright cubrirá los tres modos, secreto oculto, aviso de privacidad, prueba fallida que conserva
la configuración anterior, descarga/progreso/reanudación, licencia, recursos insuficientes,
activación y borrado, revisión con destino personalizado, confirmación comercial, cambio global,
actualización de una sesión abierta y ausencia de selectores personales, en PC y móvil.

### 16.6 Registro del repaso

El responsable cerró el repaso el 2026-10-06. Eligió una clave de cifrado obligatoria por entorno,
comparación de tres versiones para no pisar plantillas personalizadas, recuperación comercial bajo
decisión del Administrador, licencia e integridad visibles y actualización del idioma en sesiones
abiertas sin salir ni hacer sondeo. No quedan decisiones abiertas; propuesta aprobada.

## 17. Memoria coherente al activar modelos

Los selectores y fichas muestran mínimos de **2 GiB, 4 GiB y 6 GiB** para 1.5B, 3B y 7B. No mezclan
GB decimales con GiB ni presentan una estimación inferior a la puerta que aplica el administrador.

Cuando no cabe un modelo, el toast permanece hasta cerrarse y dice cuál se intentó activar, cuánta
RAM requiere y cuánta estaba disponible, con valores formateados en GiB. El modelo anterior continúa
marcado «Activo · Motor disponible». El mensaje funciona igual en `/setup` y `/settings`, en PC y
móvil, y no recomienda forzar swap.

La comprobación real en PC y Pixel 7 mostró 2,00/4,00/6,00 GiB. Al intentar el 7B, el toast permaneció
visible con 6,00 GiB requeridos y entre 5,06 y 5,08 GiB disponibles, mientras el 3B continuó marcado
«Activo · Motor disponible».

## 18. Repaso integral de formato

### 18.1 Alcance

El bloque 6 revisa todas las pantallas existentes: sesión, `/setup`, armazón, configuración, correo,
usuarios y tickets. Incluye estados vacíos, listas con contenido, formularios, diálogos, toasts,
adjuntos, vistas dobles y errores visibles. Se comprueban español e inglés en Chromium a
**1440×900**, **768×1024** y **412×915**.

La revisión cubre espaciado, alineación, jerarquía, legibilidad, cortes, desbordamientos, adaptación
responsive, foco visible, navegación por teclado, nombres accesibles, semántica, contraste y áreas
táctiles conforme a los criterios aplicables de **WCAG 2.2 AA**. Los ocho temas mantienen su prueba
automatizada de contraste; los temas claro y oscuro de fábrica reciben además el recorrido visual
detallado. Los seis temas fijos se comprueban mediante sus variables y pruebas comunes.

### 18.2 Inventario y prioridad

Cada hallazgo se registra antes de corregirse con pantalla, idioma, tema, resolución, forma de
reproducirlo y evidencia medible. La prioridad significa:

- **crítico**: bloquea una tarea;
- **alto**: impide usarla en un dispositivo o sin ratón;
- **medio**: dificulta lectura, comprensión o consistencia;
- **menor**: defecto cosmético perceptible.

El criterio de cierre elegido es **cero hallazgos abiertos de cualquier nivel**. Si aparece un cambio
que afecta comportamiento, alcance, datos, permisos o flujo, se detiene esa corrección y se abre otro
ciclo documental; no se reclasifica como formato para introducirlo en este bloque.

### 18.3 Regla de corrección y evidencia

Cuando el defecto se repite, se corrige el componente o estilo compartido y se verifican todos sus
usos. Un ajuste local se reserva para una composición propia de una sola pantalla. Se prefieren
mediciones y aserciones reproducibles sobre capturas de referencia: ancho, posición, desbordamiento,
orden de foco, nombre accesible, contraste y tamaño de objetivo se automatizan cuando aportan una
señal estable. El recorrido manual queda documentado con su resultado.

No se introduce una biblioteca de capturas visuales, no se añaden Firefox o WebKit, no se rediseñan
recorridos y no cambian backend, API, persistencia, permisos ni reglas del producto. Producción sigue
aparcada hasta el cierre de la versión 1.0.0.

### 18.4 Decisiones y repaso

El responsable eligió 1A–4A: todas las pantallas; formato, accesibilidad y responsive; PC, tableta y
móvil; inventario completo y corrección por prioridad. Eligió 5A–9A: ocho temas con recorrido
detallado en los dos de fábrica; ambos idiomas; Chromium; evidencia medible; y corrección común para
defectos repetidos. Cerró el repaso con 10A–13A y 14C: WCAG 2.2 AA aplicable, cuatro prioridades,
datos desechables con seeders, sin cambios funcionales y cero hallazgos abiertos. No quedan
decisiones abiertas.

### 18.5 Resultado as-built

El inventario recorrió las pantallas autenticadas de los cuatro papeles y las pantallas públicas en
español e inglés, con los temas claro y oscuro de fábrica y las tres resoluciones aprobadas. `/setup`
se comprobó además vacío en sus cuatro estados navegables: fueron **48 combinaciones** del asistente,
**150 vistas autenticadas y 18 públicas por idioma**. Las excepciones legítimas de WCAG para enlaces
en texto, controles con separación suficiente y contenido del correo no se registraron como defectos.

Se encontró un hallazgo medio: el conmutador de papel de `/users` medía 424 px en una ventana de
412 px cuando estaba en inglés. El componente compartido `Conmutador` ahora limita su ancho y parte
sus opciones en varias líneas; la corrección se verificó en los tres tamaños, ambos temas y con los
papeles Administrador y Soporte. No cambió ningún flujo ni contrato.

El cierre dejó **cero hallazgos abiertos**. Pasaron las 220 pruebas unitarias del frontend, 18 casos
afectados de Configuración con 4 omisiones previstas y la suite completa de Playwright con 180 casos
aprobados y 38 omisiones previstas. Los ocho temas conservaron su contraste medido.

## 19. Implementación: modal «Mejorar con IA»

### 19.1 Entrada y contenido

Los editores de descripción y comentarios permitidos a Soporte y Desarrollo incorporan, junto a sus
controles, un botón con icono y texto **«Mejorar con IA»**. En móvil puede ocupar su propia fila para
no comprimir el editor. No aparece para Usuario ni Administrador.

El botón exige un borrador no vacío y abre un modal con una copia en un `textarea` sencillo: sin
negrita, cursiva, menciones ni adjuntos. Editar esa copia no toca el editor principal. Debajo se elige
una de cinco tonalidades: profesional, cordial, breve, empática o técnica. Una nota indica que el
motor configurado procesará el texto y, cuando no sea local, que saldrá de la instalación conforme a
la aceptación administrativa existente.

### 19.2 Generación y revisión

«Mejorar» inicia una sola solicitud y queda deshabilitado mientras responde. El resultado sustituye
la copia dentro del modal para poder revisarlo y editarlo. Desde allí se puede **«Usar este texto»**,
**«Volver a generar»** o **«Cancelar»**. Regenerar usa el contenido visible y permite cambiar la
tonalidad. Usar el texto reemplaza el borrador del editor principal como texto plano con saltos de
línea; todavía no guarda ni publica nada.

Cancelar o cerrar conserva intacto el editor principal. Si se cierra durante una solicitud, una
respuesta tardía se descarta. Un fallo permanece dentro del modal, conserva su contenido y ofrece
reintentar o cancelar; no se muestra lejos del contexto mediante un toast.

### 19.3 Accesibilidad, adaptación y pruebas

El modal usa las piezas compartidas existentes, tiene título y descripción accesibles, atrapa y
restaura el foco, se cierra con Escape cuando no existe otra confirmación pendiente y anuncia carga,
error y resultado. En 412 px no crea desplazamiento horizontal y sus acciones pueden partirse en
varias filas.

Las pruebas unitarias cubren el contrato de la ruta, permisos por papel, texto vacío, tonalidad y
editor inválidos, contexto mínimo del prompt, respuesta, error conservando el borrador y aplicación
sin guardado automático. El recorrido Playwright comprueba en PC y móvil el modal, la tonalidad, la
revisión, la aplicación local y la ausencia del control para el Usuario. El cierre general previo de
la interfaz mantiene cubiertos ambos idiomas y los temas claro y oscuro de fábrica.

### 19.4 Decisiones y alcance

El responsable eligió 1A–19A y añadió que la IA reciba el nombre de la persona que requiere soporte
en el principal o de la persona de Soporte en el interno, para usarlo únicamente cuando la redacción
lo requiera. No quedan decisiones abiertas. La propuesta no introduce nuevos componentes de diseño,
formatos, adjuntos, persistencia, cuotas ni configuración administrativa.

## 20. Implementación: visibilidad de «Mejorar con IA»

La ficha sólo monta los botones «Mejorar con IA» y el modal compartido cuando el detalle devuelve
`capabilities.aiWriting: true`. Con `false` o con el campo ausente, no deja un control deshabilitado ni
un aviso ocupando espacio: la ayuda sencillamente no aparece. Esta regla se suma al borrador no vacío,
que continúa deshabilitando el botón una vez que la capacidad existe.

La pantalla no consulta configuración, salud ni proveedores por su cuenta. Al recargar el ticket
recibe otra vez la capacidad y refleja cualquier cambio administrativo. Una IA configurada pero
temporalmente caída mantiene el botón; el error sólo aparece si la persona intenta generar, dentro
del modal y conservando el texto.

Las pruebas de componente cubren capacidad verdadera, falsa y ausente. Playwright verifica en PC y
móvil que una instalación configurada muestre el botón y mantenga el recorrido de revisión. El caso
sin configuración queda en backend y componente: un navegador real sería dirigido a Configuración
antes de alcanzar esta vista, conforme a la IA obligatoria. No se modifican el modal, sus
tonalidades, los editores, la configuración administrativa ni el bloqueo global. El responsable
eligió 1A, 2A y 3A y confirmó mantener la obligatoriedad; no quedan decisiones abiertas.
