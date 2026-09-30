// Package dtos define lo que entra y lo que sale por la API del módulo `mail`.
//
// Son tipos propios y no los del repositorio: así, cambiar una columna de la tabla no cambia lo que
// ve el frontend sin decirlo aquí (docs/arquitectura.md, sección 5).
package dtos

import (
	"time"

	"catalina-support/backend/modules/mail/repositories"
	"catalina-support/backend/modules/mail/services"
)

// El nombre de los campos va en inglés, como las columnas y como los modelos del frontend.
const timeLayout = time.RFC3339

// MarkerResponse es un marcador que la plantilla admite.
type MarkerResponse struct {
	Name string `json:"name"`
	// Essential marca los que no deberían faltar (el enlace, el número). Falta no es error: es el
	// aviso que el editor enseña (docs/modules/mail.md, sección 4).
	Essential bool `json:"essential"`
}

// TemplateResponse es una plantilla tal y como la ve el editor: su texto, si está editada, qué
// marcadores admite y cuáles le faltan.
type TemplateResponse struct {
	Key       string           `json:"key"`
	Language  string           `json:"language"`
	Subject   string           `json:"subject"`
	Body      string           `json:"body"`
	Edited    bool             `json:"edited"`
	UpdatedAt string           `json:"updatedAt"`
	Markers   []MarkerResponse `json:"markers"`
	// Missing son los marcadores imprescindibles que no aparecen en el texto de hoy. Lo calcula el
	// backend: la regla vive en un solo sitio (docs/modules/mail.md, sección 8).
	Missing []string `json:"missing"`
}

// TemplatesResponse es la respuesta del listado.
type TemplatesResponse struct {
	Templates []TemplateResponse `json:"templates"`
}

// TemplateEnvelope es la respuesta de guardar y de restaurar.
type TemplateEnvelope struct {
	Template TemplateResponse `json:"template"`
	// Missing va también aquí, fuera de la plantilla, para que el aviso al guardar no obligue a
	// rebuscar dentro (docs/modules/mail.md, sección 8).
	Missing []string `json:"missing"`
}

// UpdateTemplateRequest es lo que llega al guardar. Sólo el texto: la clave y el idioma vienen en la
// dirección y quién edita sale de su sesión, no del cuerpo.
type UpdateTemplateRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// PreviewResponse es el correo ya renderizado, con datos de ejemplo.
type PreviewResponse struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Text    string `json:"text"`
}

// TestResponse dice a qué correo salió la prueba, que es siempre el de quien la pide.
type TestResponse struct {
	SentTo string `json:"sentTo"`
}

// NewTemplateResponse arma lo que ve el editor a partir de la plantilla guardada.
func NewTemplateResponse(template repositories.Template, missing []string) TemplateResponse {
	markers := make([]MarkerResponse, 0, 8)
	if declarados, ok := services.Markers(template.Key); ok {
		for _, marker := range declarados {
			markers = append(markers, MarkerResponse{Name: marker.Name, Essential: marker.Essential})
		}
	}

	// Nunca nil: `null` y `[]` se pintan distinto en el frontend, y una lista vacía es una lista.
	if missing == nil {
		missing = []string{}
	}

	return TemplateResponse{
		Key:       template.Key,
		Language:  template.Language,
		Subject:   template.Subject,
		Body:      template.Body,
		Edited:    template.Edited(),
		UpdatedAt: template.UpdatedAt.UTC().Format(timeLayout),
		Markers:   markers,
		Missing:   missing,
	}
}
