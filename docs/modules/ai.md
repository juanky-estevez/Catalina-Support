# ai

> **Estado:** as-built
> **Última actualización:** 2026-10-08
>
> **Enmienda implementada y verificada el 2026-10-08: visibilidad por configuración.** La sección 14
> distingue un motor configurado de uno disponible y permite que `tickets` consulte únicamente esa
> condición. El responsable aprobó 1A–3A y confirmó mantener la IA obligatoria.
>
> **Corrección aprobada e implementada el 2026-10-08.** Durante la implementación se confirmó que una instalación
> sellada sin IA configurada no puede abrir tickets porque la IA sigue siendo obligatoria. El
> responsable decidió mantener esa regla. La capacidad falsa se verificará como contrato defensivo,
> sin retirar el bloqueo global ni prometer un recorrido real incompatible con él. El responsable
> aprobó la corrección y las pruebas reflejan esta frontera.
>
> **Enmienda implementada y verificada el 2026-10-08: asistente de redacción.** La sección 13
> describe el motor configurado mejorando borradores de Soporte y Desarrollo, sin guardar prompts ni
> sustituir texto sin revisión. El responsable aprobó las decisiones 1A–21A. Backend, frontend y el
> recorrido real en PC y móvil pasaron; producción permaneció fuera del trabajo.
>
> **Corrección propuesta el 2026-10-08.** Al preparar la implementación se encontró que publicar
> `/api/ai/**` para una pantalla de tickets rompe la frontera modular vigente. El responsable eligió
> 20A: la ruta pública pertenece a `tickets`, que autoriza y delega internamente en `ai`. Esta
> corrección documental fue aprobada explícitamente por el responsable.
>
> **Corrección propuesta el 2026-10-08.** La generación síncrona puede usar los 240 segundos que ya
> admite el cliente del motor, pero el servidor HTTP corta a los 60. El responsable eligió 21A:
> ampliar sólo esta operación a 270 segundos, sin cambiar el resto de rutas. Requiere aprobación
> documental; el responsable la aprobó explícitamente.
>
> **Hallazgo corregido y verificado el 2026-10-08.** La validación real del bloque 5 comprobó que 3B
> no puede sustituir a 1.5B: el administrador exige 4 GiB libres antes de detener el proceso que
> ocupa 1,80 GiB. También encontró que la interfaz anuncia 1.5/2.8/5.5 GB mientras el catálogo exige
> 2/4/6 GiB. La sección 12 incorpora 15A–22A. El responsable aprobó la corrección, que quedó
> implementada y validada con los tres GGUF reales el 2026-10-08.
>
> **Hallazgo y enmienda propuesta el 2026-10-07 (bloque 5).** El catálogo y el administrador local
> están implementados y pasan pruebas con archivos simulados, pero todavía no existe la validación
> real que este documento exige para los GGUF 1.5B, 3B y 7B. La sección 11 define la medición en un
> entorno desechable. El responsable eligió 1A–4A y 5A–9A; el repaso cerró 10A–14A, sin decisiones
> abiertas. El responsable aprobó explícitamente los tres documentos el 2026-10-08.
>
> **Hallazgo y enmienda propuesta el 2026-10-07.** El contrato funcional existente sigue as-built,
> pero faltan cuatro garantías ya prometidas por la arquitectura: no reenviar credenciales mediante
> redirecciones, supervisar el proceso local, calcular recursos efectivos y probar los cuatro
> adaptadores. La sección 10 incorpora las opciones 1A, 2A, 3A y 4A elegidas por el responsable.
> El repaso incorporó 5A, 6A, 7A y 8A y no dejó decisiones abiertas. El responsable aprobó
> explícitamente la enmienda, que se implementó y verificó el 2026-10-07.
>
> **Implementado y verificado el 2026-10-07.** OpenAI, DeepSeek, compatible y Claude pasan pruebas
> de activación y resumen contra servidores falsos; las redirecciones con credenciales se rechazan.
> El motor local se reinicia con espera creciente, expone salud separada, respeta cgroup y recupera
> descargas parciales sin conservar archivos corruptos.
>
> **Enmienda propuesta el 2026-10-06.** Lo existente continúa as-built. La propuesta convierte la IA
> en una parte obligatoria de la instalación, permite elegir un motor local administrado desde la
> interfaz, un servidor propio o un proveedor externo, y genera cada resumen únicamente en el idioma
> global. Sustituye las decisiones 1, 2, 3, 8, 10, 11, 13, 14, 15, 17, 20 y 21 sólo en los puntos que
> esta sección identifica. El responsable cerró las decisiones funcionales en tres tandas y el
> repaso posterior. El responsable aprobó explícitamente el documento y la implementación fue
> verificada por bloques.
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
| 2 | **Cuándo se genera** | **En segundo plano y con reintento**: el alta no espera al motor. Esta decisión originalmente decía que la aplicación funcionaba sin configurarlo; la sección 9 la corrigió al hacer obligatoria la configuración inicial. Una caída posterior del motor no impide guardar tickets: el campo queda pendiente y se recupera cuando vuelve |
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

## 9. IA implementada: obligatoria, local o externa, en el idioma global

### 9.1 Hallazgos que obligan a enmendar lo existente

1. La IA se presenta como diferenciador del producto, pero hoy una instalación puede terminar sin
   configurarla y `AI_URL`/`AI_MODEL` vacíos significan «no integrada».
2. `ai.yml` descarga y sirve un único modelo fijo; `/settings` sólo cambia un texto con la URL y el
   nombre, y no puede descargar, activar ni borrar modelos.
3. El límite fijo de 1500 MB impide ejecutar los modelos locales de 3B y 7B elegidos para el
   catálogo. Además, un único motor compartido por desarrollo y producción permitiría que una
   instalación cambiara el modelo de la otra.
4. El cliente sólo conoce Chat Completions sin autenticación. OpenAI y DeepSeek pueden usar ese
   contrato con credenciales, pero Claude necesita un adaptador para Messages API.
5. Cada petición genera español e inglés porque el idioma pertenece hoy a la cuenta. Al convertirlo
   en una propiedad global, la segunda salida sobra y alarga la generación.
6. Los trabajos que alcanzan `sin_motor` no se recuperan solos. Una IA obligatoria y tolerante a
   caídas necesita reanudarlos cuando vuelva la conexión, sin bloquear la mesa de ayuda.

### 9.2 Qué significa «obligatoria»

- Una instalación nueva **no se sella** hasta que haya una configuración de IA probada mediante una
  generación real válida.
- Una instalación ya sellada que se actualiza sin una IA válida entra en estado **«IA requiere
  configuración»**. La cuenta de fábrica o un Administrador llega directamente a la configuración
  de IA; las demás cuentas ven esa condición y no entran al producto. No se reabre la API pública de
  `/setup` ni se toca el método de entrada.
- Después de configurarla, una caída del motor **no bloquea** crear, comentar, mover ni cerrar
  tickets. Los resúmenes quedan pendientes y la aplicación muestra el estado al Administrador.
- Cuando la conexión vuelve, un monitor detecta y cuantifica lo pendiente y lo que quedó en
  `sin_motor`, y lo reencola gradualmente en cualquiera de las tres modalidades. Respeta un máximo
  de intentos por ciclo, espera creciente y un solo
  encargo por ticket y campo; no crea un bucle sin límite.

### 9.3 Las tres modalidades

| Modalidad | Configuración | Adaptador |
| --- | --- | --- |
| **Local** | Modelo del catálogo instalado en el volumen propio del entorno | Chat Completions compatible con OpenAI, servido por `llama.cpp` |
| **Servidor propio** | URL base, modelo y autenticación: ninguna, Bearer, clave en una cabecera o usuario/contraseña | Chat Completions compatible con OpenAI |
| **Proveedor** | OpenAI, Anthropic Claude, DeepSeek o «compatible con OpenAI»; modelo sugerido o identificador escrito; credencial | OpenAI y DeepSeek: Chat Completions. Claude: Messages API. El genérico: Chat Completions |

Las listas de modelos externos son **sugerencias**, no una lista cerrada. Siempre se puede escribir
un identificador porque los proveedores cambian sus catálogos sin publicar Catalina Support. La URL
de OpenAI, Claude y DeepSeek viene de fábrica; el servidor propio y el proveedor compatible piden la
suya. No se desactiva la comprobación TLS.

### 9.4 Probar y activar es una operación atómica

La prueba deja de consultar sólo `/health`. Usa un texto artificial sin datos de personas y exige la
misma respuesta estructurada que un resumen real, en el idioma global y dentro del límite de
palabras. También comprueba autenticación, modelo y formato. Sólo una respuesta válida permite
activar.

Al cambiar de proveedor o modelo, la configuración anterior sigue atendiendo mientras se prueba la
nueva. Si descarga, autenticación, arranque o generación falla, no se reemplaza. El secreto escrito
en una prueba fallida no se persiste. El cambio exitoso guarda proveedor y modelo y desde ese momento
los trabajos nuevos los usan; los textos ya listos conservan en la tabla qué proveedor, modelo e
idioma los produjeron.

### 9.5 Administrador local de modelos

`ai.yml` deja de ejecutar directamente el guion fijo y arranca un administrador interno como proceso
principal. Ese proceso supervisa `llama-server`, pero **no usa el socket de Docker**. Su API sólo vive
en la red interna y exige `AI_MANAGER_TOKEN`, compartido con el backend y ausente del navegador.

Desde `/setup` y `/settings` se puede:

- ver el catálogo, cuál está instalado y cuál está activo;
- comprobar espacio y memoria antes de descargar;
- descargar a un archivo parcial reanudable, validar el tamaño, la suma publicada y la cabecera GGUF,
  y mostrar progreso;
- activar el archivo validado, reiniciar sólo `llama-server`, esperar salud y ejecutar la generación
  de prueba;
- conservar varios modelos, uno activo, y borrar sólo los inactivos.

Cada entorno tiene administrador, volumen y red propios. Desarrollo y producción no comparten modelo
ni configuración; en una máquina pequeña se levanta sólo el entorno con el que se trabaja. El
contenedor tiene un techo de 8 GB —no una reserva— y el administrador rechaza una activación si la
memoria o el disco disponibles no cumplen el perfil elegido.

### 9.6 Catálogo local inicial

Los tres archivos son oficiales, de la misma familia y cuantización Q4_K_M. El actual permanece como
predeterminado.

| Modelo | Descarga | Memoria del motor | Recomendación visible |
| --- | ---: | ---: | --- |
| **Qwen2.5-1.5B-Instruct** | 1.117.320.736 bytes | 1.927.028 KiB RSS medidos | 2 GiB disponibles |
| **Qwen2.5-3B-Instruct** | 2.104.932.768 bytes | 3.566.068 KiB RSS medidos | 4 GiB disponibles |
| **Qwen2.5-7B-Instruct** | 4.683.073.632 bytes en dos partes | no se activó: la máquina no cumplía el mínimo | 6 GiB disponibles |

Las cifras de proceso se midieron con el contexto de 4096 tokens usado por Catalina Support. El
requisito de disco se calcula en vivo con lo ya instalado, el archivo parcial y un margen, en vez de
prometer una cifra fija.

Fuentes verificadas el 2026-10-06: repositorios oficiales de
[1.5B](https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/tree/main),
[3B](https://huggingface.co/Qwen/Qwen2.5-3B-Instruct-GGUF/tree/main) y
[7B](https://huggingface.co/Qwen/Qwen2.5-7B-Instruct-GGUF/tree/main). La compatibilidad del servidor
se apoya en la [documentación oficial de llama.cpp](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md).

### 9.7 Un idioma por instalación

El encargo pide un único objeto `{"text":"…"}` en el idioma global. Ya no genera `es` y `en` en la
misma respuesta. La tabla guarda el texto y el idioma que lo produjo; las respuestas de tickets
devuelven un solo texto. Al cambiar el idioma, los textos ya generados se conservan y continúan
visibles con su idioma original. Sólo un movimiento del ticket o la acción individual de volver a
resumir los reemplaza, ya en el nuevo idioma global. No existe un lote de traducción o regeneración.

Esto reduce los tokens de salida y elimina el fallo «llegó un idioma y faltó el otro». No reduce el
texto de entrada: el motor todavía lee el mismo contexto del ticket.

### 9.8 Privacidad y secretos

Antes de activar un servidor propio o proveedor externo, el Administrador confirma que saldrán del
servidor **asunto, descripción, comentarios, historial y nombres de adjuntos**; al cambiar el idioma,
también los asuntos y cuerpos personalizados de las plantillas que pida traducir. No salen archivos,
destinatarios, contraseñas, tokens ni binarios. Se registra quién confirmó y cuándo; durante el
asistente se registra el acto del instalador sin inventar una cuenta, y en Configuración se guarda la
cuenta administradora. No se pregunta en cada ticket.

Las credenciales se cifran de forma autenticada antes de guardarse, con una clave maestra propia del
entorno. La API sólo devuelve si existe una credencial y nunca su valor. Tampoco se registran la
credencial, el cuerpo enviado ni la respuesta del proveedor. Perder o cambiar la clave maestra exige
volver a escribir la credencial; no se intenta recuperar texto ilegible.

`AI_CREDENTIAL_KEY` es obligatoria al arrancar el backend en todos los entornos y contiene 32 bytes
aleatorios codificados en base64. Desarrollo lleva un valor local versionado; cada instalación de
producción genera y conserva uno independiente. Si falta o no tiene el formato esperado, el backend
no arranca. Si una clave válida fue reemplazada y ya no puede descifrar una credencial existente, el
backend sí arranca en estado «IA requiere configuración» para que la cuenta de fábrica la sustituya.

### 9.9 Recuperación y costo

La recuperación de resúmenes pendientes o `sin_motor` empieza automáticamente en las tres
modalidades, incluido un proveedor externo: no crea un lote nuevo por cambiar el idioma y respeta los
topes normales de la cola. La confirmación económica se reserva exclusivamente para traducir las
once plantillas de correo. Antes de esas veintidós solicitudes se muestran tokens y costo estimados;
si se rechaza, no se llama al proveedor y las plantillas se editan manualmente.

### 9.10 Licencia e integridad del catálogo local

Antes de descargar, cada ficha muestra fuente oficial, licencia, tamaño y checksum del archivo. La
primera descarga de cada versión exige aceptar su licencia y registra modelo, versión, checksum,
fecha y actor: «instalador» en `/setup`, o la cuenta administradora en `/settings`. Si cambia la
versión, licencia o checksum, se vuelve a pedir aceptación. La aceptación no sustituye la
comprobación del archivo descargado.

### 9.11 Alcance de implementación y pruebas

Cambian `backend/modules/ai/**`, el contrato que `tickets` consume, la persistencia de
`ai_insights`, la configuración que entrega `settings`, `ai.yml`, `config/ai/**` y las pantallas de
`/setup` y `/settings`. Las pruebas cubren los cuatro adaptadores con servidores falsos, cifrado y
ocultación de secretos, cambio atómico, descarga interrumpida y reanudada, activación, borrado,
recuperación tras caída, confirmación previa de la traducción comercial de plantillas, aceptación de
licencia y conservación de resúmenes al cambiar el idioma. Ninguna prueba automática consume APIs
comerciales.

Quedan fuera facturación, cuotas y compra de créditos de proveedores; GPU y selección de capas;
modelos locales personalizados fuera del catálogo; adjuntos con visión; streaming; traducción de
tickets; y elegir un modelo por usuario, equipo o ticket.

### 9.12 Registro de decisiones y repaso de la propuesta

El responsable decidió el 2026-10-06: configuración obligatoria con operación tolerante a fallos;
paso propio en `/setup`; tres modalidades; OpenAI, Claude, DeepSeek y compatible con OpenAI;
catálogo local 1.5B/3B/7B; administración completa desde la interfaz; varios modelos instalados y uno
activo; secretos cifrados; sugerencias externas con identificador libre; prueba mediante generación
real y cambio atómico; actualización existente bloqueada hasta configurar; confirmación de
privacidad; cuatro autenticaciones para servidor propio; recuperación automática local y confirmada
por el Administrador cuando puede generar cargos; techo de 8 GB;
aislamiento por entorno; y API interna protegida con token.

El repaso del 2026-10-06 añadió: clave de cifrado obligatoria y distinta por producción; comparación
de una traducción con el destino personalizado; recuperación comercial detenida hasta que el
Administrador vea el volumen y acepte el posible costo; licencia, fuente, tamaño y checksum antes de
descargar; y actualización de sesiones abiertas mediante la versión del idioma. No quedan
decisiones abiertas. La propuesta fue aprobada explícitamente.

## 10. Robustez implementada de adaptadores y motor local

Los adaptadores externos rechazan cualquier 3xx sin seguirlo. La regla vale en la activación y en
cada resumen. Los cuatro proveedores se cubren por nombre, aunque OpenAI, DeepSeek y compatible
compartan protocolo; Claude conserva su cuerpo y cabeceras propios. Servidores falsos comprueban que
un token nunca llega a un segundo servidor.

El administrador supervisa `llama-server`: tras una salida inesperada reinicia indefinidamente el
modelo activo con esperas de 1, 2, 4… hasta 60 segundos. Activar otro modelo cancela ese ciclo y
espera al proceso anterior antes de iniciar el nuevo. El catálogo expone `active` y `healthy` por
separado y conserva el último error operativo.

Una reanudación reserva sólo lo pendiente y 256 MiB de margen. Un parcial imposible se descarta; si
su tamaño es correcto pero formato o checksum fallan, también se elimina para reiniciar limpio. La
RAM se decide con el menor valor entre la máquina y el cgroup efectivo. Estas reglas se prueban con
archivos y límites simulados, sin descargar modelos reales ni depender de la RAM del equipo.

Las descargas públicas siguen como máximo cinco redirecciones y sólo entre direcciones HTTP/HTTPS.
Las llamadas de generación con credenciales no siguen ninguna.

No cambian catálogo, modos, autenticación, prompts, cola, costos ni interfaz pública. No se añade
socket Docker ni configuración avanzada del supervisor.

El responsable cerró el repaso el 2026-10-07 con las decisiones 5A a 8A. No quedan decisiones
abiertas en esta enmienda.

## 11. Validación real del catálogo local

### 11.1 Catálogo y capacidad de la máquina

Antes de transferir un GGUF se comparan con la publicación oficial su URL, licencia, nombre,
tamaño y checksum. Cualquier diferencia detiene ese modelo: se registra como hallazgo y requiere
una enmienda aprobada del catálogo; no se confía en un archivo sólo porque coincida con los datos
locales ni se actualizan metadatos automáticamente.

La máquina de validación tiene unos 5 GiB de RAM disponible y 68 GiB de disco libre al abrir el
bloque. Se descargan y verifican los tres modelos. Se activan 1.5B y 3B; el 7B no se activa porque su
perfil exige 6 GiB. Después de recoger la evidencia se eliminan 3B y 7B y no se conserva ningún
binario del entorno desechable.

### 11.2 Mediciones

Para 1.5B y 3B se registran:

1. bytes y tiempo de descarga, checksum final y resultado de la validación GGUF;
2. tiempo desde solicitar la activación hasta que `active` y `healthy` sean verdaderos;
3. memoria del proceso `llama-server` y del contenedor una vez sano;
4. tres generaciones por modelo, con tiempo individual, mínimo, promedio y máximo;
5. muerte controlada únicamente de `llama-server`, secuencia de reintentos y tiempo hasta recuperar
   salud, con un máximo de cinco minutos.

Las entradas son dos tickets ficticios controlados, uno en español y otro en inglés. Una generación
aprueba si devuelve JSON válido, respeta el idioma global, completa `reason` y `lastAction` y cumple
los límites de longitud. La apreciación editorial se anota, pero no cambia ese resultado objetivo.

### 11.3 Resultado y defectos

El informe conserva comandos reproducibles y una tabla por modelo, sin guardar los GGUF. Un fallo
de catálogo, descarga, activación, generación, recursos, interfaz o recuperación se documenta antes
de corregirse. Si exige cambiar comportamiento o código, se añade una enmienda dentro de este bloque
y se pide otra aprobación explícita. La validación no autoriza correcciones anticipadas.

Quedan fuera GPU, proveedor comercial, servidor propio, ajuste de prompts, modelos personalizados,
comparación subjetiva de calidad y producción. El responsable eligió 1A–4A y 5A–9A; en el repaso
eligió 10A–14A. No quedan decisiones abiertas.

### 11.4 Resultado medido el 2026-10-08

Los metadatos y SHA-256 de los tres modelos coincidieron con Hugging Face. Las descargas tardaron
aproximadamente 58 s (1.5B), 113 s (3B) y 219 s (7B). El 1.5B quedó sano en unos 5 s y el cambio
1.5B→3B, incluida la liberación del proceso anterior, tardó 28,46 s.

Cada modelo activado produjo doce respuestas: español e inglés, `reason` y `lastAction`, tres veces
por combinación. Las 24 fueron JSON válido, respetaron idioma, campo y límite. En 1.5B las peticiones
tardaron entre 1,69 y 5,86 s, con promedio de 3,62 s; en 3B, entre 2,88 y 11,01 s, con promedio de
5,22 s. Dos respuestas iniciales de `lastAction` del 1.5B fueron editorialmente imprecisas, aunque
cumplieron el contrato objetivo; no se cambió el prompt porque esa comparación quedó fuera.

Al terminar sólo el hijo, el 1.5B recuperó salud en unos 4 s y el 3B en 9 s, dentro del máximo de
cinco minutos. El 7B fue verificado pero no arrancado: exigía 6 GiB y la capacidad recuperable
observada con 3B activo fue de aproximadamente 5,07 GiB.

## 12. Cambio seguro entre modelos implementado

Antes de detener el modelo activo, el administrador suma la memoria disponible efectiva y el RSS
real del proceso `llama-server`. Si esa capacidad recuperable no alcanza el requisito del destino,
rechaza la activación sin interrumpir el modelo sano. Así se comporta 7B en la máquina del bloque 5.

Si el cálculo permite intentarlo, guarda el modelo anterior, detiene su proceso y vuelve a medir la
memoria efectiva. Sólo entonces arranca el destino. Si falta memoria, falla el arranque o el nuevo
proceso no alcanza salud, restaura el anterior y espera a que vuelva sano antes de responder. El
marcador persistente `active` cambia únicamente cuando el destino está sano; una activación fallida
no convierte al destino en modelo activo.

La falta de memoria se devuelve con el requisito y la disponibilidad en bytes, además de una clave
estable. Backend y frontend la traducen a GiB para explicar en el toast cuánto requiere el modelo y
cuánto hay disponible. El catálogo y los selectores muestran los mismos mínimos aplicados por el
administrador: 2, 4 y 6 GiB.

Las pruebas cubren: cambio 1.5B→3B usando RSS recuperable; segunda medición insuficiente y
restauración; fallo de salud y restauración; rechazo de 7B sin terminar 1.5B; y error estructurado con
memoria requerida/disponible. Se reutilizan los GGUF oficiales ya verificados y se retoma la
validación desde la activación de 3B. Producción, GPU, swap forzado y cambios de catálogo quedan
fuera. El responsable eligió 15A–18A y cerró el repaso con 19A–22A.

## 13. Implementación: asistente de redacción

### 13.1 Contrato

El módulo no publica una ruta propia para esta pantalla. `tickets` le entrega internamente el
borrador ya autorizado, el nombre contextual, el tipo de ticket, el idioma y una tonalidad cerrada:
`professional`,
`friendly`, `brief`, `empathetic` o `technical`. El backend obtiene de la base el ticket y el nombre
contextual: la persona que requiere soporte en el principal y la persona de Soporte en el interno.
El cliente no envía ni puede sustituir ese nombre.

La entrada al motor contiene únicamente borrador, tonalidad, idioma global, nombre contextual y si
el ticket es principal o interno. No incluye asunto, descripción previa, conversación ni adjuntos.
El prompt pide claridad, ortografía y el tono elegido; prohíbe inventar información y exige conservar
idioma, nombres, números y detalles técnicos. El nombre se incorpora sólo cuando resulte natural.

La respuesta es un único texto plano con saltos de línea. No genera HTML, formato enriquecido,
menciones ni adjuntos. Un resultado vacío, que exceda el límite del campo o que no sea texto válido
se rechaza con una clave estable y nunca modifica el editor.

### 13.2 Autorización y datos

Sólo Soporte y Desarrollo pueden llamar al endpoint. Además del papel, el backend exige acceso al
ticket y permiso para escribir en ese editor concreto; ocultar el botón no sustituye esta regla. Un
borrador vacío o que exceda el límite vigente del campo se rechaza antes de llamar al motor.

No se guardan el borrador, el prompt ni el resultado. Los logs técnicos excluyen contenido. El texto
sólo se persiste después, mediante el guardado normal de una descripción o la publicación normal de
un comentario. Para servidor propio o proveedor externo se reutiliza la aceptación administrativa
dada al configurar la IA; el modal avisa brevemente qué motor procesará el texto y no pide una
confirmación por solicitud.

### 13.3 Ejecución y errores

El mismo adaptador configurado para los resúmenes atiende la mejora, sea motor local, servidor propio
o proveedor. Una persona mantiene una sola generación activa en la interfaz: el botón se deshabilita
hasta recibir respuesta. Cerrar el modal descarta cualquier respuesta posterior, pero no promete
cancelar una operación que el motor ya está procesando ni introduce estado compartido entre pestañas.

Los errores conservan el borrador dentro del modal y permiten reintentar. «Volver a generar» usa el
texto visible en ese momento y la tonalidad elegida, incluidas correcciones manuales. No se añaden
cuotas, historial, métricas de uso ni tablas.

La ruta que espera esta generación amplía su fecha límite de escritura a **270 segundos**: cubre los
240 segundos configurados para el motor y el margen de serialización y red. La excepción se aplica
con `ResponseController` únicamente a esta respuesta; los demás endpoints conservan el límite global
de 60 segundos. Si el motor agota su propio plazo, se devuelve `tickets.writing.unavailable` mientras
la conexión HTTP todavía puede entregar el error.

### 13.4 Alcance implementado

El responsable eligió 1A–19A y corrigió la frontera con 20A y el plazo con 21A. Quedan fuera redactar desde un campo vacío, enviar el contexto completo
del ticket, generar formato, permitir la función al Usuario o Administrador, almacenar solicitudes,
configurar tonalidades, cancelar en el proveedor y cambiar configuración o catálogo de IA. No quedan
decisiones abiertas.

### 13.5 Verificación

La suite completa del backend y del frontend, y el recorrido Playwright en PC y móvil, pasan el
2026-10-08. El recorrido usa el adaptador compatible del entorno desechable,
comprueba que la propuesta permanece en el modal, que aplicarla sólo cambia el borrador y que el
Usuario no ve el botón. No se levantó ni modificó producción.

## 14. Implementación: condición de configuración para la ayuda de redacción

El módulo ofrece internamente a `tickets` una respuesta booleana que indica si existe una
configuración efectiva de IA. Relee la configuración persistida, incluida la resolución interna de
credenciales, y conserva el respaldo del entorno usado por la propia generación. La condición final
es que los ajustes efectivos tengan una dirección de motor. No devuelve dirección, proveedor,
modelo, credenciales ni estado de salud.

**Configurada no significa disponible.** Una IA configurada conserva visible la ayuda aunque el
motor esté temporalmente caído; si se intenta generar, el modal muestra el error vigente y conserva
el borrador. No se añade una comprobación de salud, sondeo periódico ni llamada al proveedor para
decidir si se pinta un botón.

La condición se evalúa al construir el detalle del ticket. Cambiar o retirar la configuración se
refleja la próxima vez que ese detalle se solicite. Quedan fuera actualización en vivo, un endpoint
público de IA y añadir información de configuración a la sesión o a la marca pública. El responsable
eligió 1A, 2A y 3A; no quedan decisiones abiertas.

La IA continúa siendo obligatoria: una instalación sellada que requiera configuración no accede a
los tickets y permanece dirigida a Configuración. `false` es una garantía del contrato ante ausencia
del adaptador, configuración efectiva inválida, respuestas antiguas o pruebas aisladas; no abre un
modo operativo sin IA.
