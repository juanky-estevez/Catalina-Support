package services

import (
	"fmt"
	"html"
	"regexp"
)

// markerPattern encuentra un marcador en un texto: {{numero}}, {{ numero }}.
var markerPattern = regexp.MustCompile(`\{\{\s*([a-z0-9_]+)\s*\}\}`)

// ErrMissingMarker se devuelve cuando la plantilla usa un marcador que quien envía no ha dado.
//
// Es preferible fallar a mandar un correo con un hueco: un enlace vacío no se nota hasta que
// alguien pincha.
type ErrMissingMarker struct {
	Key    string
	Marker string
}

func (e ErrMissingMarker) Error() string {
	return fmt.Sprintf("mail.marker.missing: %s usa %s y no se ha dado su valor", e.Key, e.Marker)
}

// Validate comprueba que un texto sólo use marcadores que ese correo admite.
//
// Se llama al guardar, que es cuando se puede avisar a quien lo está escribiendo. No exige que
// estén todos: eso es sólo un aviso del editor (docs/modules/mail.md, sección 4).
func Validate(key, subject, body string) error {
	if _, ok := Markers(key); !ok {
		return fmt.Errorf("mail.template.unknownKey: %s", key)
	}

	allowed := make(map[string]bool)
	for _, marker := range mustMarkers(key) {
		allowed[marker.Name] = true
	}

	for _, text := range []string{subject, body} {
		for _, match := range markerPattern.FindAllStringSubmatch(text, -1) {
			if !allowed[match[1]] {
				return ErrUnknownMarker{Key: key, Marker: match[1]}
			}
		}
	}

	return nil
}

// MissingEssential dice qué marcadores imprescindibles no aparecen en el texto. No es un error: es
// el aviso que el editor enseña al guardar.
func MissingEssential(key, subject, body string) []string {
	// Se buscan los marcadores **con sus llaves**, no la palabra suelta: un texto que dice «Sin
	// enlace» no lleva el enlace, y comprobarlo buscando la palabra lo daría por bueno.
	present := make(map[string]bool)
	for _, text := range []string{subject, body} {
		for _, match := range markerPattern.FindAllStringSubmatch(text, -1) {
			present[match[1]] = true
		}
	}

	var missing []string
	for _, marker := range mustMarkers(key) {
		if marker.Essential && !present[marker.Name] {
			missing = append(missing, marker.Name)
		}
	}

	return missing
}

// RenderSubject rellena el asunto.
//
// Los valores **no** se escapan como HTML: el asunto no es HTML, y escapar ahí convertiría una
// comilla en `&quot;` a la vista de quien lee. De la codificación de la cabecera se encarga quien
// construye el mensaje.
func RenderSubject(key, subject string, data map[string]string) (string, error) {
	return render(key, subject, data, false)
}

// RenderBody rellena el cuerpo.
//
// Aquí los valores **sí** se escapan: el asunto de un ticket lo escribe una persona y puede llevar
// `<`, `>` o `&`, y el cuerpo es HTML. El escapado es del código, no del texto de la plantilla.
func RenderBody(key, body string, data map[string]string) (string, error) {
	return render(key, body, data, true)
}

func render(key, text string, data map[string]string, escape bool) (string, error) {
	allowed := make(map[string]bool)
	for _, marker := range mustMarkers(key) {
		allowed[marker.Name] = true
	}

	var failure error
	result := markerPattern.ReplaceAllStringFunc(text, func(match string) string {
		name := markerPattern.FindStringSubmatch(match)[1]

		if !allowed[name] {
			failure = ErrUnknownMarker{Key: key, Marker: name}
			return match
		}

		value, ok := data[name]
		if !ok {
			failure = ErrMissingMarker{Key: key, Marker: name}
			return match
		}

		if escape {
			return html.EscapeString(value)
		}
		return value
	})

	return result, failure
}

func mustMarkers(key string) []Marker {
	list, _ := Markers(key)
	return list
}
