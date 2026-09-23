// Package middleware reúne los middlewares transversales del servidor HTTP.
package middleware

import (
	"fmt"
	"net/http"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/shared/httpx"
)

// Chain aplica los middlewares en orden, de modo que el primero de la lista es el más
// externo (el primero en ver la petición).
//
// net/http no tiene "Use()": la cadena se arma a mano aquí, en un solo sitio, en lugar
// de envolver handlers sueltos por el código (docs/arquitectura.md, sección 5).
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// Recover evita que un pánico en un controlador tumbe todo el servidor: registra el
// pánico con go-logs y responde 500 al cliente.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logs.LogError(fmt.Sprintf("pánico en %s %s: %v", r.Method, r.URL.Path, recovered))
				httpx.WriteError(w, http.StatusInternalServerError, "error interno")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
