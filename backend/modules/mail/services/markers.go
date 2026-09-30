// Package services contiene las reglas del módulo mail: qué marcadores admite cada correo,
// cómo se rellenan sin romper el HTML y cómo se construye y se envía el mensaje.
package services

import "fmt"

// Marker es un dato que se puede insertar en una plantilla.
//
// Essential marca los que sin ellos el correo no sirve de nada —el enlace, el número—: el editor lo
// advierte al guardar, aunque deja guardar (docs/modules/mail.md, sección 4).
type Marker struct {
	Name      string
	Essential bool
}

// ErrUnknownMarker se devuelve cuando una plantilla usa un marcador que ese correo no admite.
type ErrUnknownMarker struct {
	Key    string
	Marker string
}

func (e ErrUnknownMarker) Error() string {
	return fmt.Sprintf("mail.marker.unknown: %s no admite %s", e.Key, e.Marker)
}

// markersFor es la lista cerrada de marcadores de cada correo. Es la única fuente: el editor la
// enseña, la validación la usa y el rellenado no admite nada que no esté aquí.
var markersFor = map[string][]Marker{
	"ticket.created": {
		{Name: "numero", Essential: true},
		{Name: "asunto"},
		{Name: "solicitante"},
		{Name: "enlace", Essential: true},
	},
	"ticket.escalated": {
		{Name: "numero", Essential: true},
		{Name: "asunto"},
		{Name: "motivo"},
		{Name: "enlace", Essential: true},
	},
	"ticket.waitingUser": {
		{Name: "nombre"},
		{Name: "numero", Essential: true},
		{Name: "asunto"},
		{Name: "enlace", Essential: true},
	},
	"ticket.waitingSupport": {
		{Name: "numero", Essential: true},
		{Name: "asunto"},
		{Name: "enlace", Essential: true},
	},
	"ticket.resolved": {
		{Name: "nombre"},
		{Name: "numero", Essential: true},
		{Name: "asunto"},
		{Name: "enlace", Essential: true},
	},
	"ticket.closed": {
		{Name: "nombre"},
		{Name: "numero", Essential: true},
		{Name: "enlace", Essential: true},
	},
	"ticket.backToSupport": {
		{Name: "numero", Essential: true},
		{Name: "asunto"},
		{Name: "enlace", Essential: true},
	},
	// El aviso de que te han etiquetado: la plantilla número once (docs/modules/tickets.md, sección 5).
	"ticket.mentioned": {
		{Name: "numero", Essential: true},
		{Name: "asunto"},
		{Name: "enlace", Essential: true},
	},
	"auth.invitation": {
		{Name: "nombre"},
		{Name: "enlace", Essential: true},
		{Name: "caducidad"},
	},
	"auth.recovery": {
		{Name: "nombre"},
		{Name: "enlace", Essential: true},
		{Name: "caducidad"},
	},
	"auth.passwordChanged": {
		{Name: "nombre"},
		{Name: "cuando"},
		{Name: "ip"},
	},
}

// Languages son los idiomas en los que existe cada plantilla.
var Languages = []string{"es", "en"}

// Keys son las once claves, en el orden en que se enseñan en el editor.
var Keys = []string{
	"ticket.created",
	"ticket.escalated",
	"ticket.waitingUser",
	"ticket.waitingSupport",
	"ticket.resolved",
	"ticket.closed",
	"ticket.backToSupport",
	"ticket.mentioned",
	"auth.invitation",
	"auth.recovery",
	"auth.passwordChanged",
}

// Markers devuelve los marcadores de un correo, y si la clave existe.
func Markers(key string) ([]Marker, bool) {
	list, ok := markersFor[key]
	return list, ok
}

// LanguageIsValid dice si el idioma es uno de los dos que existen.
func LanguageIsValid(language string) bool {
	for _, l := range Languages {
		if l == language {
			return true
		}
	}
	return false
}
