package services

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateRechazaMarcadoresInventados(t *testing.T) {
	err := Validate("ticket.closed", "Tu ticket {{numero}}", "<p>{{inventado}}</p>")

	var unknown ErrUnknownMarker
	if !errors.As(err, &unknown) {
		t.Fatalf("se esperaba un marcador desconocido y llegó: %v", err)
	}
	if unknown.Marker != "inventado" {
		t.Fatalf("el marcador del error debería ser «inventado» y es %q", unknown.Marker)
	}
}

func TestValidateAdmiteLosMarcadoresDeSuCorreo(t *testing.T) {
	if err := Validate("ticket.closed", "Tu ticket {{numero}}", "<p>Hola {{nombre}}</p><p><a href=\"{{enlace}}\">Ver</a></p>"); err != nil {
		t.Fatalf("no debería fallar: %v", err)
	}
}

// Un marcador de otro correo no vale: {{ip}} es de «tu contraseña ha cambiado», no de un ticket.
func TestValidateNoAdmiteMarcadoresDeOtroCorreo(t *testing.T) {
	if err := Validate("ticket.closed", "{{ip}}", "<p>{{numero}}</p>"); err == nil {
		t.Fatal("{{ip}} no es un marcador de ticket.closed y debería haber fallado")
	}
}

// El cuerpo es HTML: lo que escribe una persona se escapa siempre.
func TestRenderBodyEscapaLosValores(t *testing.T) {
	_, body, err := renderTodo("ticket.closed", "T", "<p>{{nombre}}</p>", map[string]string{
		"nombre": `<script>alert("x")</script> & <b>negrita</b>`,
	})
	if err != nil {
		t.Fatalf("no debería fallar: %v", err)
	}

	if strings.Contains(body, "<script>") {
		t.Fatalf("el valor no se ha escapado: %s", body)
	}
	if !strings.Contains(body, "&lt;script&gt;") || !strings.Contains(body, "&amp;") {
		t.Fatalf("se esperaba el valor escapado y llegó: %s", body)
	}
}

// El asunto no es HTML: escapar ahí enseñaría «&quot;» en lugar de una comilla.
func TestRenderSubjectNoEscapa(t *testing.T) {
	subject, _, err := renderTodo("ticket.closed", "{{nombre}} y \"comillas\"", "<p>{{numero}}</p>", map[string]string{
		"nombre": "Ana & Cía",
		"numero": "ACME-2026-0001",
	})
	if err != nil {
		t.Fatalf("no debería fallar: %v", err)
	}
	if subject != `Ana & Cía y "comillas"` {
		t.Fatalf("el asunto no debería ir escapado y es: %s", subject)
	}
}

// Faltar un dato es un error, no un hueco: un enlace vacío no se nota hasta que alguien pincha.
func TestRenderFallaSiFaltaUnDato(t *testing.T) {
	_, _, err := renderTodo("ticket.closed", "T", "<p>{{numero}}</p>", map[string]string{})

	var missing ErrMissingMarker
	if !errors.As(err, &missing) {
		t.Fatalf("se esperaba un marcador que falta y llegó: %v", err)
	}
	if missing.Marker != "numero" {
		t.Fatalf("el marcador que falta debería ser «numero» y es %q", missing.Marker)
	}
}

func TestMissingEssentialAvisaDelEnlace(t *testing.T) {
	missing := MissingEssential("ticket.closed", "Tu ticket {{numero}}", "<p>Sin enlace</p>")

	if len(missing) != 1 || missing[0] != "enlace" {
		t.Fatalf("se esperaba el aviso por «enlace» y llegó: %v", missing)
	}

	if aviso := MissingEssential("ticket.closed", "{{numero}}", "<p>{{enlace}}</p>"); len(aviso) != 0 {
		t.Fatalf("no debería avisar de nada y avisa de: %v", aviso)
	}
}

func renderTodo(key, subject, body string, data map[string]string) (string, string, error) {
	renderedSubject, err := RenderSubject(key, subject, data)
	if err != nil {
		return "", "", err
	}
	renderedBody, err := RenderBody(key, body, data)
	if err != nil {
		return "", "", err
	}
	return renderedSubject, renderedBody, nil
}
