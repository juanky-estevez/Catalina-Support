# Interfaz y experiencia

> **Estado:** aprobado
> **Última actualización:** 2026-09-22
>
> Aprobado por el responsable el 2026-09-22, tras repasarlo en forma de preguntas mientras se
> escribía. Fija los principios (intuitivo antes que bonito), la forma de la aplicación, la
> experiencia de los cuatro papeles, la vista doble principal/interno, el lenguaje visual con
> **Tailwind v4 y componentes propios**, los ocho temas y lo que se espera en cada dispositivo.
>
> Con este documento, la **documentación está completa**: siete documentos que cubren qué se
> construye, cómo se comporta, cómo se ve y cómo se despliega.

## 1. Alcance de este documento

Fija **cómo se ve, cómo se navega y qué siente cada papel**: la estructura de las pantallas, el
lenguaje visual, el comportamiento en cada dispositivo y la experiencia de los cuatro papeles.

**No repite** lo que cada papel puede hacer (`docs/usuarios-y-permisos.md`), ni los recorridos
(`docs/flujos.md`), ni el modelo de datos (`docs/tickets.md`). Cuando este documento y aquellos no
coincidan, mandan ellos: aquí sólo se decide cómo se presenta.

Es una **propuesta**: no habilita escribir código hasta que esté aprobada (Regla 0). La sección 11
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
6. **Lo que falla, se explica.** Un error dice qué pasó y qué hacer, en el idioma de quien lo lee.

## 3. La forma de la aplicación

### 3.1 El armazón

```text
┌───────────────────────────────────────────────────────────┐
│  cabecera: producto · papel · sesión                      │
├───────────┬───────────────────────────────────────────────┤
│  menú     │  contenido                                    │
│  lateral  │                                               │
│           │                                               │
└───────────┴───────────────────────────────────────────────┘
```

- **Menú lateral** en PC, plegable; se pliega solo en tablet y desaparece en móvil, donde vive en un
  botón de la cabecera. Ya está previsto en el armazón que existe hoy (`frontend/src/app/core`).
- El menú **sólo enseña lo que el papel puede usar**: cuatro entradas como mucho.
- La cabecera lleva el nombre del producto, quién eres, **el selector de tema** (sección 6.2) y el
  botón de salir. Nada más.

### 3.2 Las entradas del menú, por papel

| Papel | Menú |
| --- | --- |
| **Usuario** | Mis tickets · Nuevo ticket |
| **Soporte** | Bandeja · Usuarios |
| **Desarrollo** | Mi bandeja |
| **Administrador** | Usuarios · Numeración · Apariencia |

Cuatro entradas como máximo en el caso más cargado. Si algún papel llegara a necesitar más, es señal
de que algo se ha complicado en otro sitio.

### 3.3 Las bandejas

**Una sola pantalla de bandeja**, con filtros, que cambia de contenido según el papel — ya está
decidido así en `docs/tickets.md`, sección 8: la de Soporte y la de Desarrollo serían la misma
pantalla con distinto filtro, y separarlas sería duplicarlo todo para cambiar una condición.

- La lista es una **tabla** en PC (número, asunto, estado, solicitante, responsable y última
  actualización) y **tarjetas** en móvil, donde una tabla de seis columnas no se lee.
- Los filtros son **chips visibles** encima de la lista, no un formulario escondido: estado, texto y
  nada más.
- El **estado va con color y con texto**: el color ayuda a barrer la lista, el texto dice qué es.
  Nunca sólo color (sección 8).
- Orden por defecto: **lo que lleva más tiempo esperando, primero**, dentro de lo que no está
  cerrado. Lo cerrado no estorba.

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
- Un ticket **cerrado no admite comentarios** (`docs/flujos.md`): en su lugar aparece el botón de
  **reabrir**, explicando por qué.

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

## 4. Los cuatro enfoques

### 4.1 Usuario: pedir y seguir

Lo que hace un usuario es **crear un ticket y responder**. Todo lo demás sobra.

- **Al entrar**: sus tickets, los abiertos primero. Un botón grande y visible de **Nuevo ticket**.
- **Crear**: asunto, descripción y adjuntos. Tres campos, ninguno con jerga. Si se puede arrastrar
  una captura sobre la ventana, mejor que buscarla.
- **Seguir**: la conversación, con sus comentarios y los de Soporte. Cuando Soporte le pide algo, el
  estado lo dice —«Esperamos tu respuesta»— y el correo se lo ha avisado.
- **Cerrar y reabrir**: dos botones, y el de cerrar pregunta antes (`docs/usuarios-y-permisos.md`:
  cierra también el usuario).
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

- **Al entrar**: la lista de usuarios, que es lo que va a tocar.
- **Usuarios**: alta (con el rol y el origen), desactivar, reactivar, cambiar el origen y **lanzar un
  reseteo de contraseña**. Toda acción sensible dice lo que va a pasar antes de hacerlo.
- **Numeración**: el prefijo, con un aviso de que **no cambia los números ya emitidos** — es la
  duda que cualquiera tiene antes de tocarlo.
- **Apariencia**: el tema y el color primario de la instalación, con una **vista previa antes de
  guardar**. Cambiar el color institucional no puede ser a ciegas.
- **Lo que no ve**: tickets. Ni bandeja ni detalle: no atiende.

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

### 6.3 Inventario de componentes propios

Lo que hay que construir, y nada más:

`botón` · `campo de texto` · `área de texto` · `selector` · `casilla` · `etiqueta de estado` ·
`tabla` (con orden y paginación) · `tarjeta` (la lista en móvil) · `diálogo` · `menú` ·
`conmutador segmentado` (la vista doble y los filtros) · `subida de archivos` con arrastrar y soltar ·
`adjunto` (con vista previa) · `comentario` (con marcas de editado y eliminado) · `aviso` (éxito,
error, información) · `estado vacío` · `indicador de carga` · `menú lateral` · `cabecera`.

Diecinueve piezas. Cada una se construye **una vez**, con sus estados (normal, hover, foco,
deshabilitado, cargando) y su comportamiento de teclado.

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
- **Lo que se hace en móvil** (decisión abierta, sección 11): mi propuesta es que se pueda
  **consultar y comentar todo**, y que lo que se deja para el PC sea administrar usuarios y comparar
  tickets.

## 8. Accesibilidad y confianza

- **Contraste suficiente** en texto y en los colores de estado, en claro y en oscuro.
- **Todo se puede usar con teclado**: foco visible siempre, orden de tabulación coherente y nada que
  sólo funcione con el ratón.
- **Nunca sólo color** para decir algo: el estado lleva su texto y los avisos llevan su icono.
- **Etiquetas de verdad** en los campos, no textos de ejemplo que desaparecen al escribir.
- **Los errores dicen qué pasó y qué hacer**, en el idioma de quien los lee. Un «Error 422» no es un
  mensaje.
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
- **Sin refrescos automáticos ni contadores en vivo** (no hay tiempo real, `docs/tickets.md`): la
  lista se actualiza al entrar y con un botón de refrescar. Un contador que no se actualiza solo es
  peor que no tenerlo.
- **Los adjuntos de imagen y PDF se previsualizan** dentro del ticket; el resto se descarga
  (`docs/tickets.md`, sección 4).
- **Los correos son texto plano** (ya decidido): mismo tono que la interfaz y sin jerga.

## 11. Lo que se decidió al repasar este documento

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **La base visual** | **Tailwind CSS v4 con componentes propios** (19 piezas, sección 6.3), no una biblioteca |
| 2 | **La vista doble** | Dos columnas en PC con **«Los dos» por defecto** cuando existe el interno; conmutador en móvil |
| 3 | **Qué se hace desde el móvil** | **Todo menos administración y vista doble** |
| 4 | **Los temas** | **Ocho: 3 claros + 3 oscuros fijos, más 1 claro y 1 oscuro personalizables**, que son los de por defecto. El **color primario lo fija el administrador**; cada persona elige tema. Sólo cambian colores, nunca la forma |
| 5 | **Modo claro y oscuro** | Sí desde la 1.0.0, siguiendo al sistema hasta que alguien elija |
| 6 | **Accesibilidad** | **Requisito**: contraste, teclado y foco en todas las pantallas, con AA como referencia, y comprobado al construir cada componente |

## 12. Qué habilita este documento

Con `docs/interfaz-y-experiencia.md` aprobado, la documentación cubre **qué se construye, cómo se
comporta y cómo se ve**. El código puede empezar por `auth` (`docs/tickets.md`), y sus pantallas
—entrar, recuperar contraseña— se diseñan ya con estas reglas, en lugar de retocarlas después.
