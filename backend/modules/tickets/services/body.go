package services

import (
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// El texto de un ticket y el de un comentario es HTML con formato, y **la lista blanca es la única
// puerta**: lo que no esté aquí no se guarda (docs/modules/tickets.md, sección 2.3).
//
// No hay lista negra de etiquetas a propósito. Nombrando las que entran, `script`, `style`, `iframe`,
// `object`, `embed` y `form` quedan fuera por no estar —y se rechazan, que es distinto de limpiarlos en
// silencio: un texto que se guarda a medias es peor que uno que se rechaza—, y una etiqueta que
// invente mañana un navegador tampoco entra sin que alguien la añada aquí, que es la revisión que se
// quiere.
var etiquetasAdmitidas = map[string]bool{
	"p":  true,
	"br": true,
	// La negrita, la cursiva, el subrayado y el tachado, cada uno con sus dos nombres: el editor escribe
	// `strong` y `em`, y el mismo formato copiado de otro sitio llega como `b` e `i`.
	"strong": true, "b": true,
	"em": true, "i": true,
	"u": true,
	"s": true, "strike": true,
	"ul": true, "ol": true, "li": true,
	"a": true, "img": true, "video": true,
	// El `span` sólo existe para **etiquetar a una persona**: nombra por su identificador, no por el
	// nombre escrito a mano (docs/modules/tickets.md, decisión 58).
	"span": true,
}

// atributoDelAdjunto es la referencia a un adjunto **por su nombre**, y la única forma que tiene el
// texto de nombrar un archivo: la dirección se resuelve al pintar, con la lista de adjuntos del ticket
// que ya se ha leído.
//
// Es lo que hace que no haya ninguna dirección guardada dentro del texto, y con ella desaparecen tres
// problemas de golpe: una imagen de fuera —un contador que avisa a un tercero de que alguien abrió el
// ticket—, un `data:` que crece sin fin dentro de una fila, y una dirección firmada que caduca.
const atributoDelAdjunto = "data-adjunto"

// atributoDeLaMencion es la forma que tiene el texto de **etiquetar a una persona**: el
// identificador de su cuenta, nunca su nombre escrito a mano (docs/modules/tickets.md, decisión 58).
//
// Es la misma idea que `data-adjunto`: lo que se enseña —el nombre— es un adorno y el dato es el
// identificador, así que si alguien se cambia los apellidos, o hay dos personas que se llaman igual,
// la mención sigue apuntando a quien apuntaba.
const atributoDeLaMencion = "data-mencion"

// mencionValida es el formato del identificador de una mención: **sólo dígitos, sin ceros a la
// izquierda y de 1 a 18 cifras** (docs/modules/tickets.md, decisión 58). Un `0`, un `007`, un `abc`
// o un `12a` no son un identificador de cuenta y se rechazan como todo lo demás: aquí no se limpia
// en silencio.
//
// El tope de 18 cifras es el de un `bigint` con signo: cabe cualquier identificador real y no deja
// entrar un número que no cabría en la columna.
var mencionValida = regexp.MustCompile(`^[1-9][0-9]{0,17}$`)

// Los esquemas que sí puede llevar un enlace.
var esquemasDeEnlace = []string{"http://", "https://", "mailto:"}

// Y los que no entran nunca. La lista de arriba ya los deja fuera —una dirección que empieza por
// `http://` no puede empezar por `javascript:`—, pero se nombran porque son justo lo que un lector
// busca cuando quiere saber qué se rechaza: `javascript:` ejecuta código, `data:` mete el archivo
// dentro del texto, y `blob:` es una dirección provisional del navegador que guardada no significa
// nada.
var esquemasProhibidos = []string{"javascript:", "data:", "blob:"}

// SanearCuerpo comprueba que el texto de un ticket o de un comentario es HTML de la lista blanca, y lo
// devuelve **tal cual**: lo que pasa la puerta se guarda como llegó.
//
// Cuando no pasa se dice, con `ErrBodyNotAllowed`, y **no se guarda nada**: ni el texto a medias ni el
// ticket sin su descripción.
//
// **Los comentarios de HTML y el `<!DOCTYPE …>` se dejan pasar**: no son etiquetas y no se pueden
// ejecutar, y el navegador mete un `<!--StartFragment-->` cuando alguien pega texto con formato, así que
// rechazarlos rompería pegar. Lo que sí se mira, una a una, es cada etiqueta y cada atributo.
//
// Va con el analizador de HTML del propio Go y no con una expresión regular. El HTML tiene sus reglas
// —una etiqueta se cierra, un atributo puede venir sin comillas, `&` puede ser una entidad que esconde
// una letra— y adivinar dónde acaba cada cosa con patrones es exactamente lo que deja pasar lo que no
// debe.
func SanearCuerpo(cuerpo string) (string, error) {
	analizador := html.NewTokenizer(strings.NewReader(cuerpo))

	for {
		switch analizador.Next() {
		case html.ErrorToken:
			// Llegar al final no es un fallo: es que se ha acabado el texto, y todo lo que había pasó.
			if analizador.Err() == io.EOF {
				return cuerpo, nil
			}

			return "", ErrBodyNotAllowed

		case html.StartTagToken, html.SelfClosingTagToken:
			if err := revisarEtiqueta(analizador); err != nil {
				return "", err
			}

		case html.EndTagToken:
			nombre, _ := analizador.TagName()
			if !etiquetaAdmitida(string(nombre)) {
				return "", ErrBodyNotAllowed
			}
		}
	}
}

// revisarEtiqueta mira una etiqueta abierta y sus atributos: la etiqueta tiene que estar en la lista, y
// cada atributo tiene que ser de los que esa etiqueta admite.
//
// Además, `img`, `video` y `a` **exigen el suyo**: una imagen o un vídeo existen para enseñar su
// adjunto y un enlace para llevar a algún sitio, así que sin su referencia no son nada, y guardarlos
// así sería guardar un hueco.
func revisarEtiqueta(analizador *html.Tokenizer) error {
	nombre, hayAtributos := analizador.TagName()

	etiqueta := string(nombre)
	if !etiquetaAdmitida(etiqueta) {
		return ErrBodyNotAllowed
	}

	referencia := false
	enlace := false
	mencion := false

	for hayAtributos {
		var clave, valor []byte
		clave, valor, hayAtributos = analizador.TagAttr()

		atributo, contenido := string(clave), string(valor)
		if !atributoAdmitido(etiqueta, atributo, contenido) {
			return ErrBodyNotAllowed
		}

		switch atributo {
		case atributoDelAdjunto:
			referencia = true
		case atributoDeLaMencion:
			mencion = true
		case "href":
			enlace = true
		}
	}

	switch etiqueta {
	case "img", "video":
		if !referencia {
			return ErrBodyNotAllowed
		}

	case "a":
		if !referencia && !enlace {
			return ErrBodyNotAllowed
		}

	case "span":
		// El `span` **sólo** vale con su mención, y sin ningún otro atributo: uno suelto no dice a
		// quién se etiqueta, y no es una etiqueta que se guarde por su cara bonita.
		if !mencion {
			return ErrBodyNotAllowed
		}
	}

	return nil
}

// etiquetaAdmitida dice si la etiqueta está en la lista blanca. El analizador ya la devuelve en
// minúsculas, así que `<SCRIPT>` y `<script>` son la misma cosa y no hay forma de colarla por la caja.
func etiquetaAdmitida(etiqueta string) bool { return etiquetasAdmitidas[etiqueta] }

// atributoAdmitido es la lista blanca de atributos, que va **por etiqueta**: `img` y `video` llevan
// sólo la referencia a su adjunto y `a` lleva la referencia o su dirección, y ninguna otra etiqueta
// lleva ninguno.
//
// Los nombres prohibidos se comprueban aparte, aunque no estar en la lista ya los rechace, para que se
// lea qué se está dejando fuera: una `src` metería una imagen o un vídeo de fuera, un `style` es una
// dirección dentro del texto —y además pisa el ancho de la conversación—, y un `on…` es código que se
// ejecuta solo, con el `onerror` de una imagen que no carga como el caso de siempre.
func atributoAdmitido(etiqueta, clave, valor string) bool {
	if atributoProhibido(clave) {
		return false
	}

	switch etiqueta {
	case "img", "video":
		return clave == atributoDelAdjunto && hayReferenciaDeAdjunto(valor)

	case "a":
		if clave == atributoDelAdjunto {
			return hayReferenciaDeAdjunto(valor)
		}

		return clave == "href" && enlaceAdmitido(valor)

	case "span":
		// Un `span` **sólo** lleva su mención: cualquier otro atributo —un `class`, un `style` que ya
		// está prohibido, un `data-` inventado— se rechaza, y con él el cuerpo entero.
		return clave == atributoDeLaMencion && mencionValida.MatchString(valor)
	}

	return false
}

// atributoProhibido son los atributos que no entran ni en una etiqueta permitida.
func atributoProhibido(clave string) bool {
	if strings.HasPrefix(clave, "on") {
		return true
	}

	switch clave {
	case "src", "style":
		return true
	}

	return false
}

// hayReferenciaDeAdjunto dice si el nombre del archivo viene relleno. El nombre se compara al pintar
// con los adjuntos del ticket, así que lo que no exista se queda sin dirección y no se ve: aquí sólo se
// comprueba que la referencia diga algo.
func hayReferenciaDeAdjunto(nombre string) bool { return strings.TrimSpace(nombre) != "" }

// enlaceAdmitido dice si la dirección de un enlace se puede guardar.
//
// Se exige que empiece por uno de los esquemas permitidos, y sobre el valor **ya desescapado** —el
// analizador convierte `&#115;` en `s` antes de que lo veamos—: así `java&#115;cript:` se mira como lo
// que de verdad es. Los espacios de delante se quitan porque el navegador también los ignora, y una
// dirección con un salto de línea dentro no empieza por ninguno de los tres y se queda fuera.
func enlaceAdmitido(destino string) bool {
	valor := strings.ToLower(strings.TrimSpace(destino))
	if valor == "" {
		return false
	}

	for _, prohibido := range esquemasProhibidos {
		if strings.HasPrefix(valor, prohibido) {
			return false
		}
	}

	for _, admitido := range esquemasDeEnlace {
		if strings.HasPrefix(valor, admitido) {
			return true
		}
	}

	return false
}

// TextoPlano devuelve lo que dice un texto con formato, sin sus etiquetas: los trozos de texto, uno
// detrás de otro, sin inventar separaciones.
//
// **La búsqueda de la bandeja no usa esto**: quita las etiquetas en SQL, que es donde tiene que
// hacerlo para mirar todas las filas de una vez (docs/modules/tickets.md, sección 2.3). Está para quien
// necesite el texto desde Go, que hoy son las pruebas: así una prueba puede decir que lo guardado se
// lee igual sin volver a escribir aquí un `strings.ReplaceAll` que no sería el mismo que el del
// navegador.
func TextoPlano(cuerpo string) string {
	var texto strings.Builder

	analizador := html.NewTokenizer(strings.NewReader(cuerpo))

	for {
		switch analizador.Next() {
		case html.ErrorToken:
			return texto.String()

		case html.TextToken:
			texto.Write(analizador.Text())
		}
	}
}

// TextoComoHtml convierte texto plano en HTML escapando lo que el HTML se come —`&`, `<` y `>`—, y
// dejando los saltos de línea como están.
//
// Es lo que necesita el **motivo de un re-escalado**, que se guarda como comentario del interno
// (docs/modules/tickets.md, decisión 30) siendo texto plano y sin formato: el motivo no se sanea —no se
// le puede decir a Soporte que su explicación lleva un `<` de más— pero el comentario donde cae sí es
// HTML, y sin escapar, un motivo con un `<` se leería como una etiqueta. El `&` va primero: si no, se
// escaparían los `&` de los `&lt;` que se acaban de escribir.
func TextoComoHtml(textoPlano string) string {
	escapado := strings.ReplaceAll(textoPlano, "&", "&amp;")
	escapado = strings.ReplaceAll(escapado, "<", "&lt;")

	return strings.ReplaceAll(escapado, ">", "&gt;")
}
