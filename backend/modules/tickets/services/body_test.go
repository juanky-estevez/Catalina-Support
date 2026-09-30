package services

import (
	"errors"
	"strings"
	"testing"
)

// TestSaneadoDelCuerpo recorre la lista blanca de docs/modules/tickets.md, sección 2.3: lo que entra se
// guarda **igual** —el saneador no reescribe nada, sólo deja pasar o rechaza— y lo que no, se rechaza
// con su clave en vez de limpiarse en silencio.
func TestSaneadoDelCuerpo(t *testing.T) {
	casos := []struct {
		nombre   string
		cuerpo   string
		esperado bool
	}{
		// Las etiquetas de la lista, que son las del editor: párrafo, salto, negrita, cursiva,
		// subrayado, tachado, las dos listas y el enlace.
		{"un párrafo", `<p>Hola</p>`, true},
		{"un salto de línea", `<p>Hola<br>qué tal</p>`, true},
		{"la negrita", `<p><strong>Hola</strong> y <b>adiós</b></p>`, true},
		{"la cursiva", `<p><em>Hola</em> y <i>adiós</i></p>`, true},
		{"el subrayado", `<p><u>Hola</u></p>`, true},
		{"el tachado, con sus dos nombres", `<p><s>Hola</s> y <strike>adiós</strike></p>`, true},
		{"una lista con viñetas", `<ul><li>uno</li><li>dos</li></ul>`, true},
		{"una lista numerada", `<ol><li>uno</li><li>dos</li></ol>`, true},
		{"un enlace de fuera", `<p><a href="https://ejemplo.com">la web</a></p>`, true},
		{"un enlace http", `<p><a href="http://ejemplo.com">la web</a></p>`, true},
		{"un correo", `<p><a href="mailto:alguien@ejemplo.com">escríbeme</a></p>`, true},
		{"texto a secas, que es HTML válido", `Hola, mi PC no funciona`, true},
		{"un texto vacío, que es un comentario sin escribir", ``, true},
		{"un menor suelto, que no es una etiqueta", `<p>Compara: 5 < 3</p>`, true},
		// Un HTML que no cierra bien no es peligroso: lo que entra son etiquetas de la lista.
		{"sin cerrar la etiqueta", `<p>Hola`, true},

		// El adjunto va **por su nombre**, y lo llevan la imagen, el vídeo y el enlace.
		{`una imagen adjunta`, `<img data-adjunto="captura.png">`, true},
		{`un vídeo adjunto`, `<video data-adjunto="grabacion.mp4"></video>`, true},
		{`un enlace a un adjunto`, `<a data-adjunto="informe.pdf">el informe</a>`, true},

		// **La mención**: un `span` con su `data-mencion` por identificador, que es lo que etiqueta a
		// una persona y la hace observadora del ticket (docs/modules/tickets.md, decisión 58). El
		// nombre que se enseña es un adorno; el dato es el número.
		{`una mención`, `<p><span data-mencion="12">María Pérez</span>, ¿lo ves?</p>`, true},
		{`una mención con el identificador más largo`, `<span data-mencion="123456789012345678">x</span>`, true},

		// Y esto es lo que se rechaza. Se nombra uno a uno porque es lo que hay que poder decirle a
		// quien escribe: qué es lo que no entra.
		{"un script", `<p>Hola</p><script>alert(1)</script>`, false},
		{"un style", `<style>p{color:red}</style>`, false},
		{"un iframe", `<iframe src="https://ejemplo.com"></iframe>`, false},
		{"un object", `<object data="x.swf"></object>`, false},
		{"un embed", `<embed src="x.swf">`, false},
		{"un formulario", `<form action="/x"><input></form>`, false},
		{"una imagen con onerror", `<img data-adjunto="captura.png" onerror="alert(1)">`, false},
		{"un enlace con onclick", `<a href="https://ejemplo.com" onclick="alert(1)">x</a>`, false},
		{"una etiqueta con style", `<p style="position:fixed">Hola</p>`, false},
		// **Ninguna `src`**: la dirección se resuelve al pintar, con la lista de adjuntos del ticket.
		{"una imagen con src", `<img src="https://ejemplo.com/x.png">`, false},
		{"una imagen con src de datos", `<img src="data:image/png;base64,AAAA">`, false},
		{"un vídeo con src", `<video src="https://ejemplo.com/x.mp4"></video>`, false},
		{"un src junto al adjunto", `<img data-adjunto="captura.png" src="https://ejemplo.com/x.png">`, false},
		{"un enlace con javascript", `<a href="javascript:alert(1)">x</a>`, false},
		{"un enlace con javascript escondido en una entidad", `<a href="java&#115;cript:alert(1)">x</a>`, false},
		{"un enlace con data", `<a href="data:text/html,<script>alert(1)</script>">x</a>`, false},
		{"un enlace con blob", `<a href="blob:https://ejemplo.com/1234">x</a>`, false},
		{"un enlace con un espacio antes del esquema", `<a href=" javascript:alert(1)">x</a>`, false},
		{"un enlace relativo, que no es de los tres esquemas", `<a href="/tickets/1">x</a>`, false},
		// Una etiqueta que no está en la lista tampoco, y sin que importe cuál sea.
		{"una tabla", `<table><tr><td>x</td></tr></table>`, false},
		{"un div", `<div>Hola</div>`, false},
		{"una imagen sin su adjunto", `<img>`, false},
		{"un adjunto sin nombre", `<img data-adjunto="">`, false},
		{"un atributo de más en un párrafo", `<p class="grande">Hola</p>`, false},
		{"una etiqueta en mayúsculas", `<SCRIPT>alert(1)</SCRIPT>`, false},
		// Y lo mismo con la mención: un `data-mencion` que no sea un identificador de cuenta —ni el
		// cero, ni ceros a la izquierda, ni letras, ni vacío— se rechaza, y con él el cuerpo entero.
		{"una mención al cero", `<span data-mencion="0">x</span>`, false},
		{"una mención con ceros a la izquierda", `<span data-mencion="007">x</span>`, false},
		{"una mención que no es un número", `<span data-mencion="abc">x</span>`, false},
		{"una mención vacía", `<span data-mencion="">x</span>`, false},
		{"una mención de más de dieciocho cifras", `<span data-mencion="1234567890123456789">x</span>`, false},
		{"un span con un atributo de más", `<span data-mencion="12" class="grande">x</span>`, false},
		{"un span sin su mención", `<span>x</span>`, false},
		{"una mención en otra etiqueta", `<p data-mencion="12">x</p>`, false},
	}

	for _, caso := range casos {
		obtenido, err := SanearCuerpo(caso.cuerpo)

		if !caso.esperado {
			if !errors.Is(err, ErrBodyNotAllowed) {
				t.Fatalf("%s: se esperaba %v y salió %v", caso.nombre, ErrBodyNotAllowed, err)
			}

			continue
		}

		if err != nil {
			t.Fatalf("%s: se esperaba que pasara y se rechazó con %v", caso.nombre, err)
		}

		// Lo que entra se guarda **tal cual**: ni se completa el HTML ni se le quitan etiquetas.
		if obtenido != caso.cuerpo {
			t.Fatalf("%s: se esperaba guardar %q y se guardó %q", caso.nombre, caso.cuerpo, obtenido)
		}
	}
}

// TestTextoDelCuerpo comprueba que el texto de un HTML permitido se lee por lo que dice: es lo que hace
// que una descripción con formato siga diciendo lo mismo que decía sin él.
func TestTextoDelCuerpo(t *testing.T) {
	cuerpo := `<p>Hola mi PC <strong>no funciona</strong>, me aparece esto:</p><img data-adjunto="captura.png"><ul><li>Error 1</li></ul>`

	texto := TextoPlano(cuerpo)

	if !strings.Contains(texto, "no funciona") {
		t.Fatalf("el texto plano no lleva lo que se escribió: %q", texto)
	}
	if strings.Contains(texto, "data-adjunto") || strings.Contains(texto, "<p>") {
		t.Fatalf("el texto plano lleva etiquetas: %q", texto)
	}
}

// TestTextoComoHtml comprueba la conversión del texto plano a HTML, que es la que necesita el motivo de
// un re-escalado: se guarda como comentario del interno, y un motivo con un `<` no puede leerse como
// una etiqueta. El `&` va primero, o los `&lt;` recién escritos se volverían a escapar.
func TestTextoComoHtml(t *testing.T) {
	convertido := TextoComoHtml(`El error <NULL> sale con & y "comillas"`)

	quiere := `El error &lt;NULL&gt; sale con &amp; y "comillas"`
	if convertido != quiere {
		t.Fatalf("TextoComoHtml = %q, se esperaba %q", convertido, quiere)
	}
}

// TestExtensionesDeVideo comprueba las cuatro que se añadieron el 2026-09-26 (decisión 47), que se
// sirven **en línea** para poder verlas dentro del ticket.
func TestExtensionesDeVideo(t *testing.T) {
	tipos := map[string]string{
		"grabacion.mp4":  "video/mp4",
		"grabacion.webm": "video/webm",
		"grabacion.mov":  "video/quicktime",
		"grabacion.avi":  "video/x-msvideo",
	}

	for nombre, tipo := range tipos {
		if !extensionAdmitida(nombre) {
			t.Fatalf("%s debería estar en la lista cerrada de extensiones", nombre)
		}

		// El guion y el reparto del adjunto van en mayúsculas sin cambiar nada: la extensión se compara
		// en minúsculas.
		if !extensionAdmitida(strings.ToUpper(nombre)) {
			t.Fatalf("%s debería admitirse también en mayúsculas", strings.ToUpper(nombre))
		}

		if obtenido := tipoDeArchivo(nombre); obtenido != tipo {
			t.Fatalf("tipoDeArchivo(%s) = %q, se esperaba %q", nombre, obtenido, tipo)
		}

		if !sePuedePrevisualizar(nombre) {
			t.Fatalf("%s debería servirse en línea", nombre)
		}

		if _, disposicion := ServirCon(nombre, true); disposicion != "inline" {
			t.Fatalf("%s debería servirse en línea y se sirve como %q", nombre, disposicion)
		}
	}

	// El `svg` se admite **sólo como descarga**: puede llevar código dentro y el navegador lo ejecuta al
	// abrirlo en línea.
	if _, disposicion := ServirCon("dibujo.svg", true); disposicion != "attachment" {
		t.Fatalf("el svg debería descargarse siempre y se sirve como %q", disposicion)
	}
}
