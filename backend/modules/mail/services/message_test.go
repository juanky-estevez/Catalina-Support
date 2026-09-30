package services

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuildLlevaLasDosVersiones(t *testing.T) {
	message, err := Build("Catalina Support", "soporte@ejemplo.com", []string{"ana@ejemplo.com"}, "Asunto", "<p>Hola</p>")
	if err != nil {
		t.Fatalf("no debería fallar: %v", err)
	}

	text := string(message)

	for _, esperado := range []string{
		"multipart/alternative",
		"text/plain; charset=UTF-8",
		"text/html; charset=UTF-8",
		"Content-Transfer-Encoding: quoted-printable",
		"To: ana@ejemplo.com",
	} {
		if !strings.Contains(text, esperado) {
			t.Fatalf("falta %q en el mensaje:\n%s", esperado, text)
		}
	}

	// Las líneas de SMTP van con retorno de carro, no sólo con salto.
	if !bytes.Contains(message, []byte("\r\n")) {
		t.Fatal("el mensaje debería usar CRLF")
	}
}

// El asunto lleva acentos: una cabecera no admite UTF-8 crudo y hay que codificarla.
func TestBuildCodificaElAsunto(t *testing.T) {
	message, err := Build("Catalina Support", "soporte@ejemplo.com", []string{"ana@ejemplo.com"}, "Restablecer tu contraseña", "<p>Hola</p>")
	if err != nil {
		t.Fatalf("no debería fallar: %v", err)
	}

	text := string(message)
	if strings.Contains(text, "contraseña") && !strings.Contains(text, "=?utf-8?") {
		t.Fatalf("el asunto debería ir codificado:\n%s", text)
	}
	if !strings.Contains(text, "Subject: =?utf-8?") {
		t.Fatalf("el asunto debería empezar codificado:\n%s", text)
	}
}

func TestBuildRechazaDireccionesRaras(t *testing.T) {
	if _, err := Build("Catalina", "no-es-un-correo", []string{"ana@ejemplo.com"}, "A", "<p>B</p>"); err == nil {
		t.Fatal("un remitente sin forma de correo debería fallar")
	}
	if _, err := Build("Catalina", "soporte@ejemplo.com", []string{"tampoco"}, "A", "<p>B</p>"); err == nil {
		t.Fatal("un destinatario sin forma de correo debería fallar")
	}
}

func TestHTMLToTextDejaAlgoLegible(t *testing.T) {
	html := `<p>Hola {{nombre}}:</p><p>El ticket <strong>ACME-2026-0042</strong> está resuelto.</p><ul><li>Uno</li><li>Dos</li></ul><p><a href="https://ejemplo.com">Ver el ticket</a></p>`
	html = strings.ReplaceAll(html, "{{nombre}}", "Ana")

	text := HTMLToText(html)

	if strings.Contains(text, "<") || strings.Contains(text, ">") {
		t.Fatalf("no debería quedar ninguna etiqueta: %s", text)
	}
	for _, esperado := range []string{"Hola Ana:", "ACME-2026-0042", "- Uno", "- Dos", "Ver el ticket"} {
		if !strings.Contains(text, esperado) {
			t.Fatalf("falta %q en el texto:\n%s", esperado, text)
		}
	}
	if strings.Contains(text, "\n\n\n") {
		t.Fatalf("no debería haber tres saltos seguidos:\n%s", text)
	}
}

func TestHTMLToTextDeshaceEntidades(t *testing.T) {
	text := HTMLToText(`<p>Ana &amp; Cía &lt;3</p>`)
	if text != "Ana & Cía <3" {
		t.Fatalf("se esperaba «Ana & Cía <3» y llegó: %s", text)
	}
}
