package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalina-support/backend/modules/mail/repositories"
	"catalina-support/backend/modules/mail/services"
	"catalina-support/backend/shared/auth"
)

// La traducción de error a clave es el contrato con el frontend: si una clave sale mal, la pantalla
// enseña una clave en vez de un mensaje (docs/modules/mail.md, sección 8).
func TestKeyOf(t *testing.T) {
	casos := []struct {
		nombre   string
		err      error
		esperado string
	}{
		{"plantilla que no está", repositories.ErrTemplateNotFound, "mail.template.notFound"},
		{"clave inventada", fmt.Errorf("mail.template.unknownKey: ticket.inventado"), "mail.template.unknownKey"},
		{"idioma inventado", fmt.Errorf("mail.language.unknown: fr"), "mail.language.unknown"},
		{"asunto vacío", fmt.Errorf("mail.subject.required"), "mail.subject.required"},
		{"cuerpo vacío", fmt.Errorf("mail.body.required"), "mail.body.required"},
		{"sin destinatarios", fmt.Errorf("mail.to.empty"), "mail.to.empty"},
		{"marcador inventado", services.ErrUnknownMarker{Key: "ticket.created", Marker: "inventado"}, "mail.marker.unknown"},
		{"falta un marcador", services.ErrMissingMarker{Key: "ticket.created", Marker: "numero"}, "mail.marker.missing"},
		{"SMTP sin configurar", services.ErrNotConfigured, "mail.smtp.notConfigured"},
		{"SMTP que no responde", fmt.Errorf("mail.smtp.connect: %w", errors.New("dial tcp 127.0.0.1:587: connection refused")), "mail.smtp.connect"},
		{"fallo de base de datos", errors.New("ERROR: relation \"mail_templates\" does not exist"), "ERROR"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := keyOf(caso.err); obtenido != caso.esperado {
				t.Fatalf("keyOf() = %q y se esperaba %q", obtenido, caso.esperado)
			}
		})
	}
}

// El detalle técnico del SMTP no puede salir en la respuesta: la clave sí, la respuesta del servidor
// no (docs/modules/mail.md, sección 8).
func TestElDetalleDeSmtpNoSaleEnLaRespuesta(t *testing.T) {
	detalle := "dial tcp 10.0.0.5:587: connection refused"
	err := fmt.Errorf("mail.smtp.connect: %s", detalle)

	controlador := &TemplateController{}
	respuesta := httptest.NewRecorder()
	peticion := httptest.NewRequest(http.MethodGet, "/api/mail/templates", nil)

	controlador.fail(respuesta, peticion, err)

	if respuesta.Code != http.StatusServiceUnavailable {
		t.Fatalf("el código es %d y se esperaba 503", respuesta.Code)
	}

	var cuerpo map[string]string
	if err := json.Unmarshal(respuesta.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("la respuesta no es JSON: %v", err)
	}

	if cuerpo["error"] != "mail.smtp.connect" {
		t.Fatalf("la clave es %q y se esperaba \"mail.smtp.connect\"", cuerpo["error"])
	}
	if cuerpo["error"] == detalle {
		t.Fatal("el detalle del servidor de correo no debe salir en la respuesta")
	}
}

// Los códigos de cada clave, uno a uno.
func TestCodigosDeLasClaves(t *testing.T) {
	casos := map[string]int{
		"mail.template.notFound":   http.StatusNotFound,
		"mail.template.unknownKey": http.StatusNotFound,
		"mail.language.unknown":    http.StatusNotFound,
		"mail.subject.required":    http.StatusUnprocessableEntity,
		"mail.body.required":       http.StatusUnprocessableEntity,
		"mail.marker.unknown":      http.StatusUnprocessableEntity,
		"mail.marker.missing":      http.StatusUnprocessableEntity,
		"mail.to.empty":            http.StatusUnprocessableEntity,
		"mail.to.invalid":          http.StatusUnprocessableEntity,
		"mail.from.invalid":        http.StatusUnprocessableEntity,
		"mail.test.noEmail":        http.StatusUnprocessableEntity,
		"mail.smtp.notConfigured":  http.StatusServiceUnavailable,
	}

	for clave, esperado := range casos {
		if codigo, ok := statuses[clave]; !ok {
			t.Fatalf("%s no tiene código asignado", clave)
		} else if codigo != esperado {
			t.Fatalf("%s es %d y se esperaba %d", clave, codigo, esperado)
		}
	}

	// La clave que el controlador añade también tiene que estar en el mapa, o respondería 500.
	if _, ok := statuses[KeyTestNoEmail]; !ok {
		t.Fatalf("%s no está en el mapa de códigos", KeyTestNoEmail)
	}
}

// Un fallo inesperado no cuenta nada de dentro: al log y 500.
func TestUnFalloInesperadoEsUnErrorInterno(t *testing.T) {
	controlador := &TemplateController{}
	respuesta := httptest.NewRecorder()
	peticion := httptest.NewRequest(http.MethodGet, "/api/mail/templates", nil)

	controlador.fail(respuesta, peticion, errors.New("ERROR: relation does not exist"))

	if respuesta.Code != http.StatusInternalServerError {
		t.Fatalf("el código es %d y se esperaba 500", respuesta.Code)
	}

	var cuerpo map[string]string
	_ = json.Unmarshal(respuesta.Body.Bytes(), &cuerpo)
	if cuerpo["error"] != KeyInternal {
		t.Fatalf("la clave es %q y se esperaba %q", cuerpo["error"], KeyInternal)
	}
}

// La prueba no puede salir hacia una cuenta sin correo, y la cuenta de fábrica es la única así.
func TestTestSinCorreoResponde422(t *testing.T) {
	// El servicio se construye sin repositorio ni remitente a propósito: esta ruta tiene que
	// responder **antes** de tocar ninguno de los dos.
	controlador := NewTemplateController(services.NewService(nil, nil, ""))

	peticion := httptest.NewRequest(http.MethodPost, "/api/mail/templates/ticket.created/es/test", nil)
	peticion.SetPathValue("key", "ticket.created")
	peticion.SetPathValue("language", "es")

	fabrica := auth.Identity{Factory: true, Subject: auth.FactorySubject, Role: auth.RoleAdministrador}
	peticion = peticion.WithContext(auth.WithIdentity(peticion.Context(), fabrica))

	respuesta := httptest.NewRecorder()
	controlador.Test(respuesta, peticion)

	if respuesta.Code != http.StatusUnprocessableEntity {
		t.Fatalf("el código es %d y se esperaba 422", respuesta.Code)
	}

	var cuerpo map[string]string
	_ = json.Unmarshal(respuesta.Body.Bytes(), &cuerpo)
	if cuerpo["error"] != KeyTestNoEmail {
		t.Fatalf("la clave es %q y se esperaba %q", cuerpo["error"], KeyTestNoEmail)
	}
}

// Quién edita sale de la sesión: la cuenta de fábrica no deja identificador, porque no está en la
// tabla de cuentas.
func TestEditorID(t *testing.T) {
	peticionCon := func(identidad auth.Identity) *http.Request {
		peticion := httptest.NewRequest(http.MethodPut, "/api/mail/templates/ticket.created/es", nil)
		return peticion.WithContext(auth.WithIdentity(peticion.Context(), identidad))
	}

	if obtenido := editorID(peticionCon(auth.Identity{ID: 7, Role: auth.RoleAdministrador})); obtenido == nil || *obtenido != 7 {
		t.Fatalf("una cuenta normal debe dejar su identificador; llegó %v", obtenido)
	}

	if obtenido := editorID(peticionCon(auth.Identity{Factory: true, Role: auth.RoleAdministrador})); obtenido != nil {
		t.Fatalf("la cuenta de fábrica no debe dejar identificador; llegó %v", *obtenido)
	}
}
