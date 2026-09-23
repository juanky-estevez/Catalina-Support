// Package httpx centraliza cómo se escribe una respuesta HTTP.
//
// Existe para que todas las respuestas de la API tengan la misma forma: si cada
// controlador escribiera la suya, el frontend tendría que adivinar en cada endpoint.
package httpx

import (
	"encoding/json"
	"net/http"
)

// WriteJSON escribe el cuerpo como JSON con el código de estado indicado.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if body == nil {
		return
	}

	// Si fallar al serializar dejara la respuesta a medias, ya se habría enviado la
	// cabecera: no hay forma de convertirlo en un 500, así que sólo se evita el pánico.
	_ = json.NewEncoder(w).Encode(body)
}

// WriteError escribe el error con la forma única de la API: {"error": "..."}.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}
