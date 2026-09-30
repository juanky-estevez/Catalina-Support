package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalina-support/backend/shared/auth"
)

const secretoDePruebas = "un-secreto-de-pruebas"

// La respuesta de error se lee como la lee el frontend: la clave, no el texto.
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

// handlerQueAnotaDevuelve un handler final que dice si llegó a ejecutarse y con qué identidad.
func handlerQueAnota(llamado *bool, identidad *auth.Identity) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*llamado = true
		if quien, ok := auth.FromContext(r.Context()); ok {
			*identidad = quien
		}
		w.WriteHeader(http.StatusOK)
	})
}

// cargador fijo devuelve siempre la misma identidad, o el error indicado.
func cargador(identidad auth.Identity, err error) auth.IdentityLoader {
	return func(context.Context, string) (auth.Identity, error) {
		return identidad, err
	}
}

func TestAuthSinCabeceraResponde401(t *testing.T) {
	manager, _ := auth.NewTokenManager(secretoDePruebas)
	llamado := false
	handler := Auth(manager, cargador(auth.Identity{}, nil))(handlerQueAnota(&llamado, &auth.Identity{}))

	peticion := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("el código es %d y se esperaba 401", respuesta.Code)
	}
	if clave := claveDe(t, respuesta); clave != KeySessionInvalid {
		t.Fatalf("la clave es %q y se esperaba %q", clave, KeySessionInvalid)
	}
	if llamado {
		t.Fatal("el handler no debería haberse ejecutado")
	}
}

func TestAuthConBasuraResponde401(t *testing.T) {
	manager, _ := auth.NewTokenManager(secretoDePruebas)
	handler := Auth(manager, cargador(auth.Identity{}, nil))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("el handler no debería haberse ejecutado")
	}))

	for _, cabecera := range []string{"", "Bearer", "Bearer ", "Basic algo", "Bearer no.es.un.token"} {
		peticion := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		peticion.Header.Set("Authorization", cabecera)
		respuesta := httptest.NewRecorder()
		handler.ServeHTTP(respuesta, peticion)

		if respuesta.Code != http.StatusUnauthorized {
			t.Fatalf("con %q el código es %d y se esperaba 401", cabecera, respuesta.Code)
		}
		if clave := claveDe(t, respuesta); clave != KeySessionInvalid {
			t.Fatalf("con %q la clave es %q y se esperaba %q", cabecera, clave, KeySessionInvalid)
		}
	}
}

// Un token caducado se distingue del que no vale: el frontend puede decir «tu sesión ha caducado».
func TestAuthConTokenCaducadoResponde401YSuClave(t *testing.T) {
	manager, _ := auth.NewTokenManager(secretoDePruebas)
	handler := Auth(manager, cargador(auth.Identity{}, nil))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("el handler no debería haberse ejecutado")
	}))

	// Se firma en el pasado, más allá de las diez horas de vida.
	token, err := manager.Sign("7", time.Now().Add(-auth.SessionDuration-time.Hour))
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	peticion := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	peticion.Header.Set("Authorization", "Bearer "+token)
	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("el código es %d y se esperaba 401", respuesta.Code)
	}
	if clave := claveDe(t, respuesta); clave != KeySessionExpired {
		t.Fatalf("la clave es %q y se esperaba %q", clave, KeySessionExpired)
	}
}

func TestAuthConTokenDeOtroSecretoResponde401(t *testing.T) {
	ajeno, _ := auth.NewTokenManager("el-secreto-de-otro")
	nuestro, _ := auth.NewTokenManager(secretoDePruebas)
	handler := Auth(nuestro, cargador(auth.Identity{}, nil))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("el handler no debería haberse ejecutado")
	}))

	token, _ := ajeno.Sign("7", time.Now())
	peticion := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	peticion.Header.Set("Authorization", "Bearer "+token)
	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("el código es %d y se esperaba 401", respuesta.Code)
	}
	if clave := claveDe(t, respuesta); clave != KeySessionInvalid {
		t.Fatalf("la clave es %q y se esperaba %q", clave, KeySessionInvalid)
	}
}

// La cuenta se desactivó después de emitirse el token: la siguiente petición ya no vale.
func TestAuthConCuentaDesactivadaResponde401(t *testing.T) {
	manager, _ := auth.NewTokenManager(secretoDePruebas)
	handler := Auth(manager, cargador(auth.Identity{}, auth.ErrIdentityNotFound))(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("el handler no debería haberse ejecutado")
		}))

	token, _ := manager.Sign("7", time.Now())
	peticion := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	peticion.Header.Set("Authorization", "Bearer "+token)
	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("el código es %d y se esperaba 401", respuesta.Code)
	}
	if clave := claveDe(t, respuesta); clave != KeySessionInvalid {
		t.Fatalf("la clave es %q y se esperaba %q", clave, KeySessionInvalid)
	}
}

// Si la base de datos no responde, no se echa a nadie: es un 503 del servidor.
func TestAuthConLaBaseCaidaResponde503(t *testing.T) {
	manager, _ := auth.NewTokenManager(secretoDePruebas)
	handler := Auth(manager, cargador(auth.Identity{}, errors.New("la base no responde")))(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("el handler no debería haberse ejecutado")
		}))

	token, _ := manager.Sign("7", time.Now())
	peticion := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	peticion.Header.Set("Authorization", "Bearer "+token)
	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusServiceUnavailable {
		t.Fatalf("el código es %d y se esperaba 503", respuesta.Code)
	}
	if clave := claveDe(t, respuesta); clave != "error interno" {
		t.Fatalf("la clave es %q y se esperaba \"error interno\"", clave)
	}
}

// El camino feliz: la identidad queda en el contexto para los controladores.
func TestAuthDejaLaIdentidadEnElContexto(t *testing.T) {
	manager, _ := auth.NewTokenManager(secretoDePruebas)
	esperada := auth.Identity{ID: 7, Name: "Ana", LastName: "Pérez", Role: auth.RoleSoporte, Language: "es", IsActive: true}

	llamado := false
	var recibida auth.Identity
	handler := Auth(manager, cargador(esperada, nil))(handlerQueAnota(&llamado, &recibida))

	token, _ := manager.Sign("7", time.Now())
	peticion := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	peticion.Header.Set("Authorization", "Bearer "+token)
	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusOK {
		t.Fatalf("el código es %d y se esperaba 200", respuesta.Code)
	}
	if !llamado {
		t.Fatal("el handler debería haberse ejecutado")
	}
	if recibida != esperada {
		t.Fatalf("la identidad que llegó es %+v y se esperaba %+v", recibida, esperada)
	}
}

// El `sub` que se le pasa al cargador es el del token, no otro.
func TestAuthPasaElSujetoDelTokenAlCargador(t *testing.T) {
	manager, _ := auth.NewTokenManager(secretoDePruebas)

	var recibido string
	cargadorEspia := func(_ context.Context, subject string) (auth.Identity, error) {
		recibido = subject
		return auth.Identity{ID: 99}, nil
	}

	handler := Auth(manager, cargadorEspia)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	token, _ := manager.Sign("99", time.Now())
	peticion := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	peticion.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(httptest.NewRecorder(), peticion)

	if recibido != "99" {
		t.Fatalf("el cargador recibió %q y se esperaba \"99\"", recibido)
	}
}
