package authz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalina-support/backend/shared/auth"
)

func claveDe(t *testing.T, cuerpo *httptest.ResponseRecorder) string {
	t.Helper()

	var respuesta struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(cuerpo.Body.Bytes(), &respuesta); err != nil {
		t.Fatalf("la respuesta no es JSON: %v (%s)", err, cuerpo.Body.String())
	}
	return respuesta.Error
}

// conIdentidad monta la petición como la dejaría el middleware de autenticación.
func conIdentidad(identidad auth.Identity) *http.Request {
	peticion := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	return peticion.WithContext(auth.WithIdentity(peticion.Context(), identidad))
}

func TestRequireRoleDejaPasarAlPapelCorrecto(t *testing.T) {
	llamado := false
	handler := RequireRole(auth.RoleAdministrador)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		llamado = true
		w.WriteHeader(http.StatusOK)
	}))

	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, conIdentidad(auth.Identity{Role: auth.RoleAdministrador}))

	if !llamado || respuesta.Code != http.StatusOK {
		t.Fatalf("el administrador debería pasar; código %d", respuesta.Code)
	}
}

func TestRequireRoleCortaAlPapelQueNoLlega(t *testing.T) {
	handler := RequireRole(auth.RoleAdministrador)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("no debería haber pasado")
	}))

	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, conIdentidad(auth.Identity{Role: auth.RoleUsuario}))

	if respuesta.Code != http.StatusForbidden {
		t.Fatalf("el código es %d y se esperaba 403", respuesta.Code)
	}
	if clave := claveDe(t, respuesta); clave != KeyForbidden {
		t.Fatalf("la clave es %q y se esperaba %q", clave, KeyForbidden)
	}
}

// Varios papeles: basta con uno.
func TestRequireRoleAceptaVariosPapeles(t *testing.T) {
	handler := RequireRole(auth.RoleAdministrador, auth.RoleSoporte)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, conIdentidad(auth.Identity{Role: auth.RoleSoporte}))

	if respuesta.Code != http.StatusOK {
		t.Fatalf("soporte debería pasar; código %d", respuesta.Code)
	}
}

// Sin identidad en el contexto —la ruta se montó sin el middleware de autenticación— no se deja
// pasar. Se falla del lado seguro.
func TestRequireRoleSinIdentidadNoDejaPasar(t *testing.T) {
	handler := RequireRole(auth.RoleUsuario)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("no debería haber pasado")
	}))

	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, httptest.NewRequest(http.MethodGet, "/api/users", nil))

	if respuesta.Code != http.StatusForbidden {
		t.Fatalf("el código es %d y se esperaba 403", respuesta.Code)
	}
}

// La cuenta de fábrica es `administrador`, así que pasa las comprobaciones de administrador.
func TestLaCuentaDeFabricaPasaComoAdministrador(t *testing.T) {
	fabrica := auth.Identity{Factory: true, Subject: auth.FactorySubject, Role: auth.RoleAdministrador}

	if !Allowed(fabrica, auth.RoleAdministrador) {
		t.Fatal("la cuenta de fábrica debería poder lo del administrador")
	}
	if Allowed(auth.Identity{Role: auth.RoleUsuario}, auth.RoleAdministrador) {
		t.Fatal("un usuario no debería poder lo del administrador")
	}
}
