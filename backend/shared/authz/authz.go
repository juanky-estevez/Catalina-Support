// Package authz decide si quien llama puede hacer lo que pide.
//
// La autorización vive aquí y no en cada módulo: los controladores no deciden permisos por su cuenta,
// se aplican como middleware sobre las rutas, y la misma comprobación sirve dentro de los servicios
// para lo que no se puede expresar como una ruta
// (docs/usuarios-y-permisos.md, sección 10).
package authz

import (
	"net/http"

	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/httpx"
)

// KeyForbidden es la clave única del 403. La usan todos los módulos: el frontend la traduce una vez
// (docs/modules/auth.md, sección 10).
const KeyForbidden = "auth.forbidden"

// RequireRole deja pasar sólo a los papeles indicados. Va **detrás** de `middleware.Auth`, que es
// quien deja la identidad en el contexto.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := auth.FromContext(r.Context())
			if !ok {
				// Sin identidad no se puede decidir nada. Pasa sólo si la ruta se montó mal (sin
				// `middleware.Auth` delante), y responder 403 es lo prudente: no se deja pasar.
				httpx.WriteError(w, http.StatusForbidden, KeyForbidden)
				return
			}

			if !identity.Can(roles...) {
				httpx.WriteError(w, http.StatusForbidden, KeyForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Allowed dice si esa identidad puede, para comprobaciones dentro de un servicio —las que dependen
// de los datos y no sólo de la ruta, como «Soporte sólo puede dar de alta usuarios»—.
func Allowed(identity auth.Identity, roles ...string) bool {
	return identity.Can(roles...)
}
