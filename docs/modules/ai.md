# ai

> **Estado:** as-built
> **Última actualización:** 2026-10-01
>
> **Enmendado el 2026-10-01**: **se corrigen los recursos del motor** para que digan lo que de verdad
> tiene `ai.yml`, que es la fuente: el **tope de memoria son 1500m** (`mem_limit: 1500m`) y **no hay
> tope de CPU** —la afirmación de «2 CPU» que arrastraban la sección 2 y la decisión 11 no salía de
> ningún sitio del compose—; y **lo medido es ~1,44 GiB en marcha** —`docker stats` da 1.441-1.448 GiB
> y el `VmRSS` del proceso 1,51 GB—, no los «~950 MB en reposo y 1,09 GB con una entrada larga» que se
> habían anotado antes. Los dos números viejos se quedan corregidos en la sección 2 y en la decisión 11,
> y en la sección 2 se explica además el **aviso esperado del volumen del modelo** al levantar el motor
> (que el volumen `ai_modelos` figure con otro nombre de proyecto: no rompe nada y no hay que
> «arreglarlo»).
>
> **Enmendado el 2026-09-30**: el motor **se levanta aparte** (`ai.yml`, como el directorio de pruebas y
> Keycloak) y **es opcional** —decisión y corrección del responsable—: sin él la mesa de ayuda
> funciona entera, sin los dos resúmenes, y el asistente de primer arranque **lo avisa sin
> bloquear**, diciendo qué comprobar (su contenedor, y el modelo la primera vez). **Su dirección y
> su modelo ya se configuran desde Configuración** (decisión 16 de `docs/modules/settings.md`), con
> su tarjeta y su botón de «Probar la conexión»: este módulo los lee **en cada petición** por la
> interfaz `Configuracion` que él declara (sección 5), y si no hay nada puesto usa `AI_URL` y
> `AI_MODEL` **como respaldo** (`AI_PALABRAS` y `AI_ESPERA_SEGUNDOS` siguen siendo del entorno).
>
> **Pasa a as-built el 2026-09-27**, el mismo día en que se escribió: el módulo, el contenedor del
> motor, la tabla `ai_insights`, los dos campos en la lista y en la ficha, y sus pruebas **están
> hechos y verificados**. Lo que se aprendió al ponerlo en marcha —el motor tarda lo que tarda, el
> modelo no devuelve el JSON si no se le exige, y una de cada tantas respuestas llega a medias— está
> recogido en las secciones 2, 3 y 6 y en las decisiones 11 a 19.
>
> **Escrito el 2026-09-27**, al pedir el responsable **dos campos que resuman el ticket**: **«Motivo»**
> —de qué va el ticket— y **«Última acción»** —qué fue lo último que pasó—, **los dos redactados por un
> motor de inteligencia artificial** que corre **en un contenedor aparte**, y **los dos en español y en
> inglés**. Las cinco decisiones de diseño las eligió él sobre mi propuesta, y **todas las recomendadas
> salieron adelante**: motor local con un modelo pequeño, generación en segundo plano con reintento, una
> sola llamada que devuelve los dos idiomas, recálculo con cada movimiento del ticket, y los dos campos
> en los dos tipos de ticket (decisiones 1 a 5).
>
> Es el **sexto módulo** del backend y el único que **no tiene pantalla propia**: se lee dentro de las
> pantallas de `tickets`, que es de quien son los tickets.
>
> **Aprobado por el responsable el 2026-09-27.** Habilita implementar.

## 1. Qué hace, en una frase

Cuando un ticket nace, se edita o se mueve —un comentario, un cambio de estado, un escalado—, este
módulo **le manda el texto del ticket a un motor de IA propio** y guarda **dos redacciones cortas**:

| Campo | Qué cuenta | Ejemplo |
| --- | --- | --- |
| **Motivo** | **De qué va el ticket**, en una línea | «El usuario no puede entrar: la aplicación rechaza su contraseña desde esta mañana» |
| **Última acción** | **Qué fue lo último que pasó**, en una línea | «Se pidió al usuario una captura del error; queda esperando su respuesta» |

**Los dos se guardan en español y en inglés**, y la pantalla enseña el del idioma de quien mira.

**No sustituyen a nada**: son un resumen para leer la lista de un vistazo. El texto del ticket sigue
siendo el ticket, y ninguna decisión del sistema se toma a partir de estos campos.

## 2. El motor: un contenedor propio, y nada sale del servidor

| | |
| --- | --- |
| **Dónde** | Un contenedor **`catalina_support_ai`**, aparte del backend, con **llama.cpp** sirviendo un modelo por HTTP |
| **Modelo** | **Qwen2.5-1.5B-Instruct**, cuantizado **Q4_K_M** (unos 1,1 GB), elegido porque **cabe en la memoria de esta máquina**: no hay GPU y quedaban unos 1,8 GB libres |
| **Imagen** | `ghcr.io/ggml-org/llama.cpp:server-b11206`, **con la compilación fijada** (y no la etiqueta móvil `:server`): un `pull` no puede cambiar el motor por debajo sin que nadie lo haya decidido. El tope de memoria son **1500m** (`mem_limit: 1500m` en `ai.yml`) —**`ai.yml` no limita las CPU**— y es lo que impide que se coma la máquina. Lo que gasta, medido en marcha: **~1,44 GiB** —`docker stats` da 1.441-1.448 GiB, el **98% del tope**, y el `VmRSS` del proceso 1,51 GB—, justo por debajo de él |
| **Arranque** | `ai.yml` con un guion (`config/ai/01-descargar-modelo.sh`) que **descarga el modelo sólo si falta** —a un archivo temporal, y comprobando la cabecera `GGUF` antes de renombrarlo— y después hace `exec` al servidor. El `.gguf` vive en un volumen con nombre, así que sobrevive a `down`/`up` y **no se vuelve a descargar** |
| **Red** | Una red propia, **`catalina-support-ai`**, a la que entran los backends de desarrollo y de producción. **Un solo contenedor sirve a los dos**: no caben dos (decisión 1) |
| **Puerto** | **Ninguno publicado en la máquina**: el motor no se alcanza desde fuera, sólo por la red interna de docker |
| **Internet** | El modelo **se descarga una vez** al levantar el contenedor, y a partir de ahí **el motor funciona sin salida a internet** |
| **Datos** | El texto del ticket **no sale del servidor**. Es la razón de que el motor sea local y no una API de terceros (decisión 1) |

**Y un aviso esperado al levantar el motor**: `docker compose -f ai.yml up -d` puede avisar de que el
volumen `ai_modelos` «was created for project "catalina-support-ai" (expected "catalina-support")».
**No rompe nada y no hay que «arreglarlo»**: el volumen tiene **nombre fijo** (`name: ai_modelos` en
`ai.yml`), así que es el mismo aunque cambie el nombre del proyecto con el que se creó; por eso el motor
entra en él y **el modelo no se vuelve a descargar**.

**Y tarda, y eso se dice**: sin GPU, el motor lee el mensaje a ~24 piezas por segundo y redacta a
5-9, así que **cada campo tarda entre 12 y 24 segundos** y un ticket largo —6 000 caracteres— ronda el
minuto y medio entre los dos. La segunda llamada del mismo ticket es unas **siete veces más rápida**
(llama.cpp reaprovecha el prefijo en caché). De ahí salen dos cosas que no son adorno:

- **El tiempo de espera del cliente son 240 segundos** (`AI_ESPERA_SEGUNDOS`), no los 30 de fábrica de
  cualquier cliente HTTP: con menos, fallarían justo los tickets largos, que son los que más lo
  necesitan.
- **La cola tiene dos trabajadores y el motor atiende de uno en uno** (`--parallel 1`): el segundo
  trabajador hace cola dentro del motor, y lo que no puede pasar es que una espera al motor bloquee
  el alta de un ticket.

**«El motor está caído» no puede ser «la mesa de ayuda está caída»**: si el contenedor no está, los
dos campos quedan **sin texto y con su aviso**, y todo lo demás —abrir tickets, comentar, escalar,
cerrar— funciona igual. Es la decisión 2 y es la condición que manda sobre todas las demás.

## 3. Qué se le manda y qué se le pide

**El texto se arma en `tickets`** —que es quien tiene el ticket— y se le pasa a este módulo ya hecho:

```text
Asunto: No puedo entrar a la aplicación
Descripción: Desde esta mañana me sale «usuario o contraseña incorrectos»…
Adjuntos: captura.png, informe.pdf

Conversación (de lo más antiguo a lo más reciente):
- user1 (comentario): Sigue fallando después de reiniciar.
- support1 (estado): en progreso
- support1 (comentario): ¿Puedes probar desde otro navegador?
```

- **El HTML se convierte a texto**: al motor no le llegan etiquetas, ni direcciones, ni `data-adjunto`.
- **Los adjuntos van por su nombre**, nunca su contenido: decir «adjuntos: captura.png» da el contexto
  que hace falta, y mandar el binario pediría un modelo con visión y multiplicaría el tamaño.
- **El historial va en orden, lo más reciente al final**, y **si hay que recortar se recorta por el
  principio**: lo último que pasó es justo lo que hay que leer. El tope es de **6 000 caracteres**
  (decisión 6).
- **Al motor se le pide un JSON con los dos idiomas** y nada más (decisión 3):

```json
{ "es": "…", "en": "…" }
```

- **Y hay que exigírselo de dos maneras a la vez** (medido, y no es un detalle): con
  `response_format: {"type": "json_object"}` en el cuerpo **y** con la exigencia escrita **al final del
  mensaje del usuario** —«Devuelve EXACTAMENTE este objeto JSON, con esas dos claves y ningún campo
  más», seguido de la plantilla—. Con sólo una de las dos, un modelo de 1,5B envuelve la respuesta en
  un bloque de código, se inventa sus propias claves o contesta en un solo idioma. Aun así, el cliente
  **rescata el primer objeto JSON** del texto (por si viene envuelto) y **reintenta** si no vale.
- **El tope de piezas de la respuesta deja margen**: palabras pedidas × 4 + 160, y nunca más de media
  ventana del modelo: una respuesta cortada a mitad de JSON es una respuesta que no vale.

- **Lo que se pide se pide dos veces, con la misma llamada y distinto encargo**: el encargo del
  «Motivo» (de qué va) y el de la «Última acción» (qué pasó al final). Son **dos peticiones**, una por
  campo, porque piden cosas distintas (decisión 7).
- **La temperatura es baja y la respuesta tiene un tope de palabras**: se busca un resumen, no
  literatura. Si el modelo devuelve algo que no es el JSON pedido, **se reintenta una vez** y, si
  vuelve a fallar, el campo queda en error (decisión 8).

## 4. La tabla: `ai_insights`

Una tabla, en `v1.0.0.sql` con su **puesta al día** (la 1.0.0 no está cerrada, así que no se abre una
versión nueva por esto):

| Columna | Para qué |
| --- | --- |
| `id` | Clave |
| `ticket_number` | El número del ticket, `ACME-2026-0042` o `INT-…`, **tal y como se lee en los correos** |
| `kind` | `motivo` o `ultima_accion` |
| `state` | `pendiente`, `listo`, `error` o **`sin_motor`** (hay motor configurado y no contesta, o no hay motor: los dos se dicen distinto en la pantalla) |
| `text_es`, `text_en` | Las dos redacciones, **en la misma fila** |
| `error_key` | La clave del error, cuando el estado es `error` |
| `model` | El modelo que lo escribió, para saber con qué se generó |
| `requested_at`, `generated_at` | Cuándo se pidió y cuándo se escribió |
| `attempts` | Cuántos intentos lleva, que es lo que corta el reintento infinito. **Se cuenta también cuando acierta**: queda dicho cuántos hicieron falta |

- **Una fila por ticket y por campo** (`UNIQUE (ticket_number, kind)`): el campo se **reescribe** en
  cada movimiento, no se acumula historial. Lo que interesa es lo último, y guardar cada versión sería
  llenar la base de textos que nadie lee.
- **Se indexa por `ticket_number`**, que es como se piden: la lista de tickets pide los suyos **en
  bloque**, no uno a uno (decisión 9). El índice es redundante con el `UNIQUE (ticket_number, kind)`
  —la primera columna ya sirve— y se queda porque lo pide el documento y no estorba: una tabla con dos
  filas por ticket no va a crecer por esto.
- **La tabla lleva `CHECK` de verdad**: `kind` sólo admite los dos campos, `state` los cuatro estados y
  `attempts` no puede ser negativo. Lo que la aplicación no debe escribir, la base tampoco lo acepta.

## 5. El contrato entre módulos

**`tickets` declara lo que necesita y `ai` lo cumple sin conocerlo** —el mismo patrón que usa con
`users`—, y quien los une es `main.go`:

```go
// En tickets: lo que este módulo necesita del motor.
type Insights interface {
    // Pedir apunta un resumen y vuelve enseguida: no espera al motor.
    Pedir(entrada EntradaDeResumen)
    // De devuelve lo que hay, en bloque, para pintar la lista.
    De(numeros []string) (map[string]Resumen, error)
}

// En ai: el servicio que lo cumple.
func (s *Service) Pedir(entrada tickets.EntradaDeResumen)
func (s *Service) De(numeros []string) (map[string]tickets.Resumen, error)
```

**Y del otro lado, el motor**: `ai` declara lo que necesita de la configuración de la instalación y
`settings` lo cumple, **sin que ninguno de los dos importe al otro**:

```go
// En ai: de dónde sale el motor que está puesto.
type Configuracion interface {
    AI() (AIDatos, error)
}

// El cableado lo conecta y traduce al tipo de cada módulo (main.go).
aiService.SetConfiguracion(instalacionesDeLaConfiguracion{settings: settingsService})
```

- **La configuración se lee en cada petición**, no al arrancar: la dirección y el modelo de la base
  son los que mandan cada vez que se encola un resumen, se resuelve un trabajo o se pregunta por la
  salud del motor. Es lo que hace que cambiarlos en Configuración valga sin reiniciar nada.
- **Si la configuración no tiene motor, el entorno hace de respaldo** (`AI_URL`, `AI_MODEL`), y lo
  resuelve el módulo con sus propios ajustes: la configuración dice **si hay** motor, no cuál es el de
  reserva (decisión 21).
- **La prueba de la conexión** (`Probar(url)`) es la que pide el botón de Configuración: pregunta a
  `<url>/health` con un contexto de tres segundos y devuelve el error en vez de un booleano. La
  declara `settings` en su interfaz `ProberDeIA` y la cumple este módulo —igual que `Disponible()`,
  que es la que usa el asistente de primer arranque—.

- **El texto no se guarda**: la tabla no tiene columna de texto a propósito —sería copiar el ticket
  entero en otra tabla, y con él los datos de las personas—. Eso obliga a que, para volver a pedir un
  resumen que quedó a medias al reiniciar, alguien **vuelva a armar el texto**: `ai` declara también
  `Fuente` (`TextoParaElMotor(numero)`) y **`tickets` lo cumple** con su método del mismo nombre. Es el
  camino de vuelta del mismo adaptador de `main.go`.
- **`ai` no importa `tickets` y `tickets` no importa `ai`**: cada uno declara su lado y `main.go` los
  junta (`docs/arquitectura.md`, sección 4). Es la regla dura de modularidad, y aquí es lo que permite
  que el motor se pueda apagar sin tocar el módulo de tickets.
- **El módulo `ai` no tiene endpoints propios.** Lo que el usuario toca —pedir que se regenere un
  campo— entra por **`POST /api/tickets/{number}/insights`**, que es de `tickets`: la pantalla de
  tickets sólo puede hablar con su propia API, y el módulo de tickets es quien decide a quién se lo
  pide. Sin esta regla, el frontend tendría que llamar a `/api/ai/**` desde una pantalla de tickets.
- **Este módulo no escribe en las tablas de nadie más**, y nadie escribe en la suya.

## 6. Cuándo se pide, y qué pasa cuando falla

| Cuándo | Qué se pide |
| --- | --- |
| Se crea el ticket | Motivo y última acción |
| Se edita el asunto o la descripción | Motivo y última acción |
| Se comenta | Última acción |
| Cambia de estado, se escala, se reabre, se cierra | Última acción |
| Alguien pulsa **«Regenerar»** en la ficha | Los dos |

- **Todo va en segundo plano** (decisión 2): quien crea un ticket **no espera al motor**. El alta
  responde cuando el ticket está guardado, con los campos en **`pendiente`**.
- **Una cola en memoria con pocos trabajadores** (dos), porque **el motor es lo caro**: sin cola, diez
  comentarios seguidos pedirían diez resúmenes a la vez a un modelo que atiende de uno en uno.
- **Si el motor no contesta, se reintenta tres veces** con espera creciente (2 s y 4 s); después el
  campo queda en **`sin_motor`** con su clave y **no se reintenta solo** —un bucle de reintentos contra
  un contenedor caído es ruido—. Se puede volver a pedir con «Regenerar».
- **Si contesta algo que no vale, se reintenta tres veces también, al momento** (sin espera: no es un
  motor lento, es una respuesta mal explicada), y después el campo queda en **`error`** con
  `ai.invalid`. **Tres y no una** porque un modelo de 1,5B es estocástico: medido, una de cada tantas
  respuestas llega a medias —sólo un idioma, o cortada por el tope de piezas— y volver a preguntar sale
  casi gratis. Lo que no hay es bucle infinito.
- **Cuando una respuesta no vale, el log dice su forma y nunca su contenido**: qué claves llegaron y
  cuántos caracteres tenía cada una («claves: es=142 caracteres, en=falta»). Sin eso, «contestó algo que
  no vale» no se puede investigar; con eso, se sabe si faltó el inglés, si llegó vacío o si se cortó a
  mitad. **El texto del ticket no se registra nunca.**
- **Un trabajo por ticket y campo**: pedir dos veces lo mismo no encola dos veces; el nuevo sustituye
  al que estaba esperando, porque el texto viejo ya no sirve.
- **Al reiniciar el backend, la cola se pierde** y los campos que estaban `pendiente` se quedan así:
  hay una **pasada de puesta al día al arrancar** que vuelve a pedir lo que quedó a medias (decisión
  10).

## 7. Límites, y lo que NO hace

- **No se le manda el contenido de los adjuntos** —ni el binario de un PDF ni los píxeles de una
  captura—: sólo sus nombres.
- **No escribe en el ticket, ni comenta, ni mueve estados, ni manda correos.** Sólo redacta.
- **No decide nada**: si el resumen dice una cosa y el ticket dice otra, manda el ticket.
- **No se usa en los correos**: ningún aviso lleva estos textos (los avisos siguen usando el número, el
  asunto y el enlace, `docs/flujos.md`).
- **El inglés no se traduce del español**: se piden las dos redacciones a la vez, y el texto inglés
  puede decir lo mismo con otras palabras. Es un resumen, no un documento oficial.
- **La calidad depende del modelo**: un modelo de 1,5B redacta resúmenes cortos correctamente y **se
  equivoca en los matices**. Es la consecuencia dicha y aceptada de que quepa en esta máquina
  (decisión 1), y por eso los campos **no sustituyen a la lectura del ticket**.
- **Un modelo por instalación, no uno por usuario ni por idioma**: el motor no sabe quién mira.

## 8. Lo que se decidió

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **Dónde corre el motor** | **Contenedor local**, un **modelo pequeño** (Qwen2.5-1.5B-Instruct Q4_K_M) **compartido por desarrollo y producción** por la red `catalina-support-ai`. **Decisión del responsable, 2026-09-27**, y la recomendada: **nada del ticket sale del servidor**, al precio de un resumen más humilde. **No caben dos contenedores** —quedaban 1,8 GB libres y el modelo ocupa ~1,1 GB— y **no hay GPU**. La alternativa que descartó era una API externa, mejor y más cara en privacidad |
| 2 | **Cuándo se genera** | **En segundo plano y con reintento**: el alta no espera al motor, y **la aplicación funciona entera sin él**. Un campo sin generar dice «pendiente» y se reintenta; los datos del ticket nunca dependen de un contenedor que puede estar caído |
| 3 | **Los dos idiomas** | **Una sola llamada** que pide `{"es": …, "en": …}`: más rápido que dos y sin traducción de por medio. Si el JSON no llega, se reintenta una vez |
| 4 | **Cuándo se recalcula la última acción** | **Con cada movimiento**: comentario, estado, escalado, reapertura, cierre, asignación y edición. Es lo que la hace útil: un texto que no se actualiza deja de ser «la última acción» en cuanto alguien escribe |
| 5 | **Alcance** | **Los dos tipos de ticket**, y en sus dos listas. En un **interno**, «Motivo» es **el del escalado, que ya existe** y no se genera; «Última acción» sí. **El Administrador los ve** como ve todo lo demás, de sólo lectura, y **no puede regenerarlos** (no escribe) |
| 6 | **Cuánto texto se manda** | Asunto, descripción, **la conversación y el historial**, con el HTML en texto y **los adjuntos por su nombre**, recortado a **6 000 caracteres** y **quitando por el principio**: lo último que pasó es lo que hay que leer |
| 7 | **Dos peticiones, no una** | «Motivo» y «Última acción» se piden por separado, porque son encargos distintos: de qué va el ticket, y qué fue lo último. Un solo texto para las dos cosas saldría peor en las dos |
| 8 | **Qué pasa si el motor contesta mal** | Un reintento, y después el campo queda en **`error`** con su clave: **no hay bucle**. Los reintentos automáticos de la cola (tres, con espera creciente) son para el motor caído o lento, no para una respuesta que no vale |
| 9 | **Cómo se leen para la lista** | **En bloque** (`De(numeros)`), una consulta por página y no una por fila: la lista de tickets se pinta con una sola pregunta, como las cuentas de las personas |
| 11 | **La compilación del motor se fija** | `server-b11206` y no la etiqueta móvil: subir de motor es una decisión, no algo que pase un día al hacer `pull`. Y sus argumentos quedan escritos: 4096 de contexto, dos hilos, una conversación a la vez, 512 piezas de respuesta, sin interfaz web, **1500m de memoria de techo** —`ai.yml` no limita las CPU, corregido el 2026-10-01— y una comprobación de salud con quince minutos de margen por la primera descarga |
| 12 | **El tope de palabras son 40** | Una columna de una tabla no es un párrafo: con 60 palabras el «Motivo» llenaba la celda. La pantalla, además, recorta a dos líneas en la lista y lo enseña entero en la ficha |
| 13 | **El tiempo de espera del motor son 240 segundos** | Medido, un ticket largo tarda ~95 s en los dos campos. Con 30 s —lo de fábrica del cliente HTTP— fallarían justo los tickets largos. Está en `AI_ESPERA_SEGUNDOS` |
| 14 | **Se le exige el JSON por dos vías** | `response_format` en el cuerpo **y** la exigencia escrita al final del mensaje del usuario. Con una sola, el modelo envuelve la respuesta, se inventa las claves o contesta en un solo idioma. El rescate del primer objeto JSON y el reintento se quedan como red |
| 15 | **La respuesta inválida y el motor ausente se tratan distinto** | `error` con `ai.invalid` y reintento inmediato para lo primero; `sin_motor` con `ai.unavailable` y espera creciente para lo segundo. **Y se comparan con `errors.Is`**, no por el texto: es un hallazgo de las pruebas —al enriquecer el error con su forma, la comparación dejó de valer y una respuesta inválida se habría tratado como un motor caído— |
| 16 | **El texto no se guarda, y hay un camino de vuelta** | `ai` declara `Fuente` y `tickets` lo cumple con `TextoParaElMotor`: es lo que permite retomar al arrancar sin copiar el ticket en otra tabla |
| 17 | **La puesta al día no toca el contador de intentos** | Si lo reiniciara, reiniciar el backend sería la forma de saltarse el tope de un motor caído |
| 18 | **Las pruebas del repositorio van contra PostgreSQL de verdad**, dentro de una transacción que se deshace | Es lo único que comprueba de verdad el `ON CONFLICT` de la tabla de una fila por ticket y campo. Si no hay base, se saltan |
| 19 | **El módulo depende de una interfaz de almacén** | La cumple el repositorio, y es lo que permite probar la cola, los reintentos y los estados sin base de datos |
| 20 | **La cola tiene dos trabajadores, y el motor atiende de uno en uno** | El motor tiene una sola ranura (`--parallel 1`), así que el segundo trabajador hace cola dentro de él. Se quedan dos a propósito: es lo que evita que un ticket se quede esperando detrás de la cola de otro en la aplicación |
| 10 | **Qué pasa al reiniciar el backend** | La cola vive en memoria y **se pierde**, así que al arrancar **se vuelve a pedir** lo que quedó en `pendiente` o con pocos intentos. Lo que estaba en `error` **no se reintenta solo**: para eso está el botón |
| 21 | **El motor sale de la configuración** | **Decisión del responsable, 2026-09-30**: la dirección y el modelo se guardan en Configuración (decisión 16 de `docs/modules/settings.md`) y este módulo los lee **en cada petición** por la interfaz `Configuracion` que él declara; el entorno (`AI_URL`, `AI_MODEL`) queda **como respaldo**. La prueba de la conexión —`Probar(url)` contra `<url>/health`, con tres segundos de tope— la consume `settings` por su interfaz `ProberDeIA` |
