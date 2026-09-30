package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/httpx"
)

// Las claves con las que se rechaza una petición sin sesión válida. Están aquí y no sueltas por el
// código porque el frontend las traduce (docs/modules/auth.md, sección 10).
const (
	KeySessionInvalid = "auth.session.invalid"
	KeySessionExpired = "auth.session.expired"
)

// Auth exige una sesión válida y deja en el contexto quién es la persona que llama.
//
// En cada petición: valida la firma y la fecha del token, saca el `sub`, **lee la cuenta** y guarda
// su identidad. Leer la cuenta cada vez es lo que hace que desactivar a alguien o cambiarle el papel
// surta efecto al instante, sin esperar a que caduque su token
// (docs/usuarios-y-permisos.md, sección 6).
//
// El cargador llega como argumento y no se importa: el middleware vive en `shared` y `shared` no
// puede importar de `modules`, así que es `main.go` quien lo conecta con el servicio de `users`
// (docs/modules/auth.md, sección 3).
func Auth(tokens *auth.TokenManager, load auth.IdentityLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := auth.BearerToken(r.Header.Get("Authorization"))
			if !ok {
				// Sin cabecera, sin `Bearer` o con el token vacío. Es el mismo caso que un token que
				// no vale: quien lo lee hace lo mismo en los tres.
				httpx.WriteError(w, http.StatusUnauthorized, KeySessionInvalid)
				return
			}

			subject, err := tokens.Verify(token, time.Now())
			if err != nil {
				if errors.Is(err, auth.ErrTokenExpired) {
					httpx.WriteError(w, http.StatusUnauthorized, KeySessionExpired)
					return
				}
				httpx.WriteError(w, http.StatusUnauthorized, KeySessionInvalid)
				return
			}

			identity, err := load(r.Context(), subject)
			if err != nil {
				if errors.Is(err, auth.ErrIdentityNotFound) {
					// La cuenta se borró o se desactivó después de emitirse el token: la sesión ya no vale.
					httpx.WriteError(w, http.StatusUnauthorized, KeySessionInvalid)
					return
				}

				// Un fallo del servidor no es una sesión inválida: se registra y se responde 503, para
				// no echar a nadie a la pantalla de entrada por un problema de la base de datos.
				// Nunca se registran datos de la cuenta: sólo el motivo (AGENTS.md, reglas para agentes).
				logs.LogError("no se pudo leer la cuenta de la sesión: " + err.Error())
				httpx.WriteError(w, http.StatusServiceUnavailable, "error interno")
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), identity)))
		})
	}
}
