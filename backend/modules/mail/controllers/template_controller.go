// Package controllers es la capa HTTP del módulo mail: lee la petición, llama al servicio y
// responde. No decide reglas ni permisos: los permisos los pone el middleware sobre la ruta y las
// reglas viven en el servicio (docs/modules/mail.md, sección 7).
package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/mail/dtos"
	"catalina-support/backend/modules/mail/repositories"
	"catalina-support/backend/modules/mail/services"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/httpx"
)

// Las claves que este módulo puede añadir a las que ya devuelve el servicio.
const (
	KeyTestNoEmail = "mail.test.noEmail"
	KeyInternal    = "error interno"
)

// TemplateController atiende los cuatro endpoints del editor de plantillas.
type TemplateController struct {
	service *services.Service
	translator services.Translator
}

func (c *TemplateController) SetTranslator(translator services.Translator) { c.translator = translator }

// LanguageDrafts generates the eleven reviewable translations without saving any of them.
func (c *TemplateController) LanguageDrafts(w http.ResponseWriter, r *http.Request) {
	var input struct { Language string `json:"language"`; ConfirmCommercialCost bool `json:"confirmCommercialCost"`; Manual bool `json:"manual"` }
	if json.NewDecoder(r.Body).Decode(&input) != nil { httpx.WriteError(w, http.StatusBadRequest, KeyInternal); return }
	// Cost confirmation is required only for the provider mode; the settings workflow supplies that
	// fact before exposing this action. The boolean is still explicit in the request for auditability.
	var drafts []services.LanguageDraft
	var err error
	if input.Manual { drafts, err = c.service.ManualLanguageDrafts(input.Language) } else { drafts, err = c.service.LanguageDrafts(input.Language, c.translator, input.ConfirmCommercialCost) }
	if err != nil {
		if strings.Contains(err.Error(), "mail.translation.costConfirmationRequired") {
			httpx.WriteJSON(w, http.StatusConflict, map[string]any{"error": "mail.translation.costConfirmationRequired", "requests": 22, "estimatedInputTokens": 12000, "estimatedOutputTokens": 8000})
			return
		}
		c.fail(w, r, err); return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"drafts": drafts})
}

func (c *TemplateController) ApplyLanguage(w http.ResponseWriter, r *http.Request) {
	var input struct { Language string `json:"language"`; Templates []services.LanguageChoice `json:"templates"` }
	if json.NewDecoder(r.Body).Decode(&input) != nil { httpx.WriteError(w, http.StatusBadRequest, KeyInternal); return }
	if err := c.service.ApplyLanguage(input.Language, input.Templates, editorID(r)); err != nil { c.fail(w, r, err); return }
	w.WriteHeader(http.StatusNoContent)
}

// NewTemplateController construye el controlador.
func NewTemplateController(service *services.Service) *TemplateController {
	return &TemplateController{service: service}
}

// List devuelve las veinte plantillas con todo lo que el editor necesita para pintarse.
func (c *TemplateController) List(w http.ResponseWriter, r *http.Request) {
	templates, err := c.service.Templates()
	if err != nil {
		c.fail(w, r, err)
		return
	}

	// La lista se arma vacía y no nula: `{"templates":[]}` se pinta distinto de `{"templates":null}`.
	respuesta := dtos.TemplatesResponse{Templates: make([]dtos.TemplateResponse, 0, len(templates))}
	for _, template := range templates {
		respuesta.Templates = append(respuesta.Templates, c.toResponse(template))
	}

	httpx.WriteJSON(w, http.StatusOK, respuesta)
}

// Update guarda el texto de una plantilla y devuelve el aviso de los marcadores que falten.
func (c *TemplateController) Update(w http.ResponseWriter, r *http.Request) {
	key, language := r.PathValue("key"), r.PathValue("language")

	var entrada dtos.UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "error interno")
		return
	}

	// Quién edita sale de su sesión, no del cuerpo: la cuenta de fábrica no está en la tabla de
	// cuentas, así que su edición se guarda sin identificador (docs/modules/mail.md, sección 3).
	updatedByID := editorID(r)

	if err := c.service.Update(key, language, entrada.Subject, entrada.Body, updatedByID); err != nil {
		c.fail(w, r, err)
		return
	}

	template, err := c.service.Template(key, language)
	if err != nil {
		c.fail(w, r, err)
		return
	}

	c.writeTemplate(w, http.StatusOK, template)
}

// Reset devuelve una plantilla a su texto de fábrica.
func (c *TemplateController) Reset(w http.ResponseWriter, r *http.Request) {
	key, language := r.PathValue("key"), r.PathValue("language")

	if err := c.service.Reset(key, language); err != nil {
		c.fail(w, r, err)
		return
	}

	// Se relee para devolver el texto de fábrica ya puesto, no lo que había antes.
	template, err := c.service.Template(key, language)
	if err != nil {
		c.fail(w, r, err)
		return
	}

	c.writeTemplate(w, http.StatusOK, template)
}

// Preview devuelve el correo ya renderizado, con datos de ejemplo.
//
// Es lo que pinta el editor antes de guardar: **lo mismo que saldría**, porque lo renderiza el mismo
// motor que el envío (docs/modules/mail.md, decisión 17).
func (c *TemplateController) Preview(w http.ResponseWriter, r *http.Request) {
	// El cuerpo es opcional: sin él se previsualiza lo que hay guardado.
	var entrada dtos.UpdateTemplateRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&entrada)
	}

	vista, err := c.service.Preview(r.PathValue("key"), r.PathValue("language"), entrada.Subject, entrada.Body)
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.PreviewResponse{
		Subject: vista.Subject,
		Body:    vista.Body,
		Text:    vista.Text,
	})
}

// Test manda una prueba de la plantilla **al correo de quien la pide**.
//
// Nunca a una dirección que venga en la petición: si se pudiera elegir el destino, el endpoint
// serviría para mandar correos con la plantilla de la institución a quien se quisiera
// (docs/modules/mail.md, sección 7).
func (c *TemplateController) Test(w http.ResponseWriter, r *http.Request) {
	key, language := r.PathValue("key"), r.PathValue("language")
	identity := auth.MustFromContext(r.Context())

	if strings.TrimSpace(identity.Email) == "" {
		// La cuenta de fábrica `admin` es la única sin correo: no hay dónde mandar la prueba.
		httpx.WriteError(w, http.StatusUnprocessableEntity, KeyTestNoEmail)
		return
	}

	if err := c.service.SendTest(key, language, identity.Email); err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.TestResponse{SentTo: identity.Email})
}

// writeTemplate responde con la plantilla y, aparte, con los marcadores imprescindibles que le
// faltan: es el aviso del editor, que avisa sin impedir guardar (docs/modules/mail.md, sección 4).
func (c *TemplateController) writeTemplate(w http.ResponseWriter, status int, template repositories.Template) {
	respuesta := c.toResponse(template)

	httpx.WriteJSON(w, status, dtos.TemplateEnvelope{
		Template: respuesta,
		Missing:  respuesta.Missing,
	})
}

func (c *TemplateController) toResponse(template repositories.Template) dtos.TemplateResponse {
	missing := c.service.MissingEssential(template.Key, template.Subject, template.Body)
	return dtos.NewTemplateResponse(template, missing)
}

// editorID es el identificador de quien edita, o nil si es la cuenta de fábrica.
func editorID(r *http.Request) *int64 {
	identity := auth.MustFromContext(r.Context())
	if identity.Factory || identity.ID == 0 {
		return nil
	}
	return &identity.ID
}

// fail traduce el error del servicio a la clave y el código que le tocan.
//
// El contrato es que los errores del servicio lleven delante su clave: `mail.subject.required` o
// `mail.smtp.connect: dial tcp …`. Lo que va detrás de los dos puntos es detalle para el log y no
// sale en la respuesta (docs/modules/mail.md, sección 8).
func (c *TemplateController) fail(w http.ResponseWriter, r *http.Request, err error) {
	clave := keyOf(err)

	if codigo, ok := statuses[clave]; ok {
		httpx.WriteError(w, codigo, clave)
		return
	}

	// Todo lo que empieza por `mail.smtp.` es el servidor de correo, y siempre es un 503: no es algo
	// que quien edita pueda arreglar escribiendo mejor.
	if strings.HasPrefix(clave, "mail.smtp.") {
		logs.LogError("fallo al mandar correo: " + err.Error())
		httpx.WriteError(w, http.StatusServiceUnavailable, clave)
		return
	}

	// Lo que no lleva clave es un fallo nuestro —la base de datos, por ejemplo—: al log y 500.
	logs.LogError("error inesperado en mail: " + err.Error())
	httpx.WriteError(w, http.StatusInternalServerError, KeyInternal)
}

// statuses es el código de cada clave conocida (docs/modules/mail.md, sección 8).
var statuses = map[string]int{
	// No existe: la clave del correo, el idioma o la fila.
	"mail.template.notFound":   http.StatusNotFound,
	"mail.template.unknownKey": http.StatusNotFound,
	"mail.language.unknown":    http.StatusNotFound,

	// Existe, pero lo que llega no vale.
	"mail.subject.required": http.StatusUnprocessableEntity,
	"mail.body.required":    http.StatusUnprocessableEntity,
	"mail.marker.unknown":   http.StatusUnprocessableEntity,
	"mail.marker.missing":   http.StatusUnprocessableEntity,
	"mail.translation.incomplete":       http.StatusUnprocessableEntity,
	"mail.translation.protectedChanged": http.StatusUnprocessableEntity,
	"ai.invalid":                         http.StatusUnprocessableEntity,
	"ai.unavailable":                     http.StatusServiceUnavailable,
	"mail.test.noEmail":     http.StatusUnprocessableEntity,
	"mail.to.empty":         http.StatusUnprocessableEntity,
	"mail.to.invalid":       http.StatusUnprocessableEntity,
	"mail.from.invalid":     http.StatusUnprocessableEntity,

	// Configuración del servidor de correo.
	"mail.smtp.notConfigured": http.StatusServiceUnavailable,
}

// keyOf saca la clave de un error del servicio: lo que va antes de los dos puntos.
//
// Se apoya en los tipos de error para los dos casos que importan y, para el resto, en la forma de la
// clave, que es la misma que ya usa el módulo en todos sus errores.
func keyOf(err error) string {
	var unknown services.ErrUnknownMarker
	if errors.As(err, &unknown) {
		return "mail.marker.unknown"
	}

	var missing services.ErrMissingMarker
	if errors.As(err, &missing) {
		return "mail.marker.missing"
	}

	if errors.Is(err, repositories.ErrTemplateNotFound) {
		return "mail.template.notFound"
	}

	if errors.Is(err, services.ErrNotConfigured) {
		return "mail.smtp.notConfigured"
	}

	text := err.Error()
	if i := strings.Index(text, ":"); i > 0 {
		return strings.TrimSpace(text[:i])
	}
	return strings.TrimSpace(text)
}
