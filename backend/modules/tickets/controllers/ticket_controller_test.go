package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalina-support/backend/modules/tickets/dtos"
	"catalina-support/backend/modules/tickets/services"
)

// TestFailDeCierreSinComentario: cerrar sin decir por qué tiene que salir **422 con
// `tickets.cierre.sinComentario`**, que es la clave que el frontend ya traduce (decisión 81). Es el
// contrato con la pantalla, y por eso se prueba la traducción, no el servicio.
func TestFailDeCierreSinComentario(t *testing.T) {
	controlador := &TicketController{}
	respuesta := httptest.NewRecorder()

	controlador.fail(respuesta, services.ErrCierreSinComentario)

	if respuesta.Code != http.StatusUnprocessableEntity {
		t.Fatalf("el código es %d y se esperaba 422", respuesta.Code)
	}

	var cuerpo map[string]string
	if err := json.Unmarshal(respuesta.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("la respuesta no es JSON: %v", err)
	}
	if cuerpo["error"] != "tickets.cierre.sinComentario" {
		t.Fatalf("la clave es %q y se esperaba \"tickets.cierre.sinComentario\"", cuerpo["error"])
	}
}

// TestEstadoSolicitado: el estado llega en `state` —el nombre que manda la pantalla— y también en
// `to`, que es el que fija la tabla de endpoints del documento (docs/modules/tickets.md, sección 5).
// Si vienen los dos, manda `state`.
func TestEstadoSolicitado(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  dtos.StateRequest
		esperado string
	}{
		{"state solo", dtos.StateRequest{State: "resuelto"}, "resuelto"},
		{"to solo", dtos.StateRequest{To: "cerrado"}, "cerrado"},
		{"state con espacios", dtos.StateRequest{State: " en espera "}, "en espera"},
		{"los dos", dtos.StateRequest{State: "resuelto", To: "cerrado"}, "resuelto"},
		{"ninguno", dtos.StateRequest{Comment: "sólo comentario"}, ""},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := estadoSolicitado(caso.entrada); obtenido != caso.esperado {
				t.Fatalf("estadoSolicitado() = %q y se esperaba %q", obtenido, caso.esperado)
			}
		})
	}
}

// TestEsPATCHDeSoloEtiquetas: el `PATCH` que trae **sólo** las etiquetas va al camino que Desarrollo
// puede usar desde el interno (decisión 85); en cuanto trae asunto, descripción o categoría —o
// etiquetas junto a otra cosa— vuelve al camino de siempre, que es del solicitante y de Soporte.
func TestEsPATCHDeSoloEtiquetas(t *testing.T) {
	texto := func(valor string) *string { return &valor }
	categoria := func(valor int64) *int64 { return &valor }
	lista := func(valor ...string) *[]string { return &valor }

	casos := []struct {
		nombre   string
		entrada  dtos.UpdateTicketRequest
		esperado bool
	}{
		{"sólo etiquetas", dtos.UpdateTicketRequest{Tags: lista("red-wifi", "vpn")}, true},
		{"etiquetas vacías", dtos.UpdateTicketRequest{Tags: lista()}, true},
		{"también el asunto", dtos.UpdateTicketRequest{Subject: texto("otro"), Tags: lista("vpn")}, false},
		{"también la descripción", dtos.UpdateTicketRequest{Description: texto("<p>x</p>"), Tags: lista("vpn")}, false},
		{"también la categoría", dtos.UpdateTicketRequest{CategoryID: categoria(2), Tags: lista("vpn")}, false},
		{"sin etiquetas", dtos.UpdateTicketRequest{Subject: texto("otro")}, false},
		{"cuerpo vacío", dtos.UpdateTicketRequest{}, false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := esPATCHDeSoloEtiquetas(caso.entrada); obtenido != caso.esperado {
				t.Fatalf("esPATCHDeSoloEtiquetas() = %v y se esperaba %v", obtenido, caso.esperado)
			}
		})
	}
}
