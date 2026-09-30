package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/config"
)

// reino tiene dentro lo que necesita el camino de OIDC: el descubrimiento, el canje y la lectura de
// quién es. Se levanta un servidor de verdad en lugar de doblar el cliente HTTP porque lo que hay que
// probar es **el protocolo**: qué se pide, con qué y qué se hace con lo que llega.
type reino struct {
	servidor *httptest.Server
	// peticiones cuenta las veces que se ha pedido el descubrimiento: tiene que ser una.
	peticiones int

	// lo que contesta cada paso; se cambia desde cada prueba.
	persona   map[string]any
	canje     int
	sinCorreo bool
}

func nuevoReino(t *testing.T) *reino {
	t.Helper()

	r := &reino{
		persona: map[string]any{
			"sub":                "9f1c",
			"email":              "ana.keycloak@ejemplo.com",
			"given_name":         "Ana",
			"family_name":        "Pérez",
			"preferred_username": "ana",
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		r.peticiones++
		json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 r.servidor.URL,
			"authorization_endpoint": r.servidor.URL + "/auth",
			"token_endpoint":         r.servidor.URL + "/token",
			"userinfo_endpoint":      r.servidor.URL + "/userinfo",
		})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, peticion *http.Request) {
		_ = peticion.ParseForm()
		if r.canje != 0 {
			w.WriteHeader(r.canje)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
			return
		}
		if peticion.Form.Get("grant_type") != "authorization_code" || peticion.Form.Get("code") == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"access_token": "el-token-de-keycloak", "token_type": "Bearer"})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, peticion *http.Request) {
		if peticion.Header.Get("Authorization") != "Bearer el-token-de-keycloak" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.sinCorreo {
			delete(r.persona, "email")
		}
		json.NewEncoder(w).Encode(r.persona)
	})

	r.servidor = httptest.NewServer(mux)
	t.Cleanup(r.servidor.Close)

	return r
}

// El camino configurado contra ese reino.
func (r *reino) oidc() *OIDC {
	return NewOIDC(config.OIDC{
		Issuer:       r.servidor.URL,
		ClientID:     "catalina-support",
		ClientSecret: "el-secreto-del-cliente",
		RedirectURI:  "https://dev.catalina-support.example.com/api/auth/keycloak/callback",
	})
}

// La dirección de entrada lleva lo que Keycloak necesita, y el `state` va dentro.
func TestAuthURL(t *testing.T) {
	r := nuevoReino(t)

	direccion, err := r.oidc().AuthURL("el-state-firmado")
	if err != nil {
		t.Fatalf("no se pudo armar la dirección: %v", err)
	}

	partes, err := url.Parse(direccion)
	if err != nil {
		t.Fatalf("la dirección no se entiende: %v", err)
	}
	if partes.Path != "/auth" {
		t.Fatalf("debería ir a la pantalla de entrada del reino y va a %q", partes.Path)
	}

	parametros := partes.Query()
	esperados := map[string]string{
		"response_type": "code",
		"client_id":     "catalina-support",
		"redirect_uri":  "https://dev.catalina-support.example.com/api/auth/keycloak/callback",
		"scope":         "openid email profile",
		"state":         "el-state-firmado",
	}
	for clave, esperado := range esperados {
		if obtenido := parametros.Get(clave); obtenido != esperado {
			t.Fatalf("%s debería ser %q y es %q", clave, esperado, obtenido)
		}
	}
}

// El canje devuelve la persona con lo que entiende el resto del sistema: origen `keycloak` y el `sub`
// como identificador externo, que es lo que manda cuando alguien cambia de correo.
func TestExchangeLeeLaPersona(t *testing.T) {
	r := nuevoReino(t)
	oidc := r.oidc()

	persona, err := oidc.Exchange("el-codigo")
	if err != nil {
		t.Fatalf("debería haber canjeado: %v", err)
	}

	if persona.Origin != auth.OriginKeycloak {
		t.Fatalf("el origen debería ser `keycloak` y es %q", persona.Origin)
	}
	if persona.ExternalID != "9f1c" {
		t.Fatalf("el identificador externo debería ser el `sub` y es %q", persona.ExternalID)
	}
	if persona.Email != "ana.keycloak@ejemplo.com" || persona.Name != "Ana" || persona.LastName != "Pérez" {
		t.Fatalf("la persona no se leyó bien: %+v", persona)
	}

	// El descubrimiento se pide una vez, aunque se use tres veces: es una ida y vuelta de red de
	// menos en el camino de entrar.
	if _, err := oidc.AuthURL("otro-state"); err != nil {
		t.Fatalf("no se pudo volver a armar la dirección: %v", err)
	}
	if r.peticiones != 1 {
		t.Fatalf("el descubrimiento se pidió %d veces y debería pedirse una", r.peticiones)
	}
}

// Sin correo no hay cuenta: el correo es con lo que se entra y con lo que se identifican los tickets.
func TestExchangeSinCorreo(t *testing.T) {
	r := nuevoReino(t)
	r.sinCorreo = true

	if _, err := r.oidc().Exchange("el-codigo"); !errors.Is(err, ErrOIDCNoEmail) {
		t.Fatalf("se esperaba ErrOIDCNoEmail y llegó %v", err)
	}
}

// Un código que Keycloak no acepta no es un fallo de Keycloak: es una vuelta que no vale.
func TestExchangeConCodigoQueNoVale(t *testing.T) {
	r := nuevoReino(t)
	r.canje = http.StatusBadRequest

	if _, err := r.oidc().Exchange("un-codigo-caducado"); !errors.Is(err, ErrOIDCRejected) {
		t.Fatalf("se esperaba ErrOIDCRejected y llegó %v", err)
	}

	// Y sin código no se llama a nadie: es lo que llega cuando alguien cancela en Keycloak.
	if _, err := r.oidc().Exchange("  "); !errors.Is(err, ErrOIDCRejected) {
		t.Fatalf("se esperaba ErrOIDCRejected y llegó %v", err)
	}
	if r.canje != http.StatusBadRequest {
		t.Fatal("sin código no debería haberse canjeado nada")
	}
}

// Con el reino caído, el fallo es del servidor y se dice como tal: no es que la persona haya hecho
// nada mal.
func TestExchangeConElReinoCaido(t *testing.T) {
	r := nuevoReino(t)
	oidc := r.oidc()
	r.servidor.Close()

	if _, err := oidc.AuthURL("el-state"); !errors.Is(err, ErrOIDCUnavailable) {
		t.Fatalf("se esperaba ErrOIDCUnavailable y llegó %v", err)
	}

	// Y también cuando el reino responde pero el canje no contesta.
	r2 := nuevoReino(t)
	oidc2 := r2.oidc()
	if _, err := oidc2.documento(); err != nil {
		t.Fatalf("el descubrimiento debería funcionar: %v", err)
	}
	r2.servidor.Close()
	if _, err := oidc2.Exchange("el-codigo"); !errors.Is(err, ErrOIDCUnavailable) {
		t.Fatalf("se esperaba ErrOIDCUnavailable y llegó %v", err)
	}
}

// keycloakDeMentira es el camino de OIDC visto desde el servicio, sin red.
type keycloakDeMentira struct {
	datos auth.DirectoryAccount
	err   error
	// estados son los `state` que ha recibido, para comprobar que se firma uno nuevo en cada entrada.
	estados []string
}

func (k *keycloakDeMentira) Configured() bool { return true }

func (k *keycloakDeMentira) AuthURL(state string) (string, error) {
	k.estados = append(k.estados, state)

	return "https://sso.ejemplo.com/auth?state=" + url.QueryEscape(state), nil
}

func (k *keycloakDeMentira) Exchange(string) (auth.DirectoryAccount, error) {
	if k.err != nil {
		return auth.DirectoryAccount{}, k.err
	}

	return k.datos, nil
}

// servicioConKeycloak arma el módulo con el camino de Keycloak conectado **y el método puesto en
// él**, que es lo que quiere decir «el camino existe» desde que la instalación entra por uno solo.
func servicioConKeycloak(t *testing.T, cuentas *cuentasDeMentira, keycloak Keycloak) *Service {
	t.Helper()

	firmante, err := auth.NewTokenManager("un-secreto-de-prueba-largo")
	if err != nil {
		t.Fatalf("no se pudo construir el firmante: %v", err)
	}

	servicio := NewService(cuentas, nil, nil, firmante, Config{
		AdminPassword: "la-de-fabrica-larga",
		PublicAppURL:  "https://dev.catalina-support.example.com",
		TokenSecret:   "un-secreto-de-prueba-largo",
	})
	servicio.SetAccess(accesoDeMentira{
		metodo:   auth.MethodKeycloak,
		keycloak: config.OIDC{Issuer: "https://sso.ejemplo.com", ClientID: "catalina-support", RedirectURI: "https://x/callback"},
	})
	servicio.SetKeycloak(keycloak)

	return servicio
}

// La entrada empieza con un `state` firmado y distinto cada vez, y el navegador acaba en Keycloak.
func TestStartKeycloak(t *testing.T) {
	keycloak := &keycloakDeMentira{}
	servicio := servicioConKeycloak(t, &cuentasDeMentira{}, keycloak)

	for i := 0; i < 2; i++ {
		direccion, err := servicio.StartKeycloak()
		if err != nil {
			t.Fatalf("debería haber dirección: %v", err)
		}
		if !strings.HasPrefix(direccion, "https://sso.ejemplo.com/auth?state=") {
			t.Fatalf("la dirección no es la del reino: %q", direccion)
		}
	}

	if len(keycloak.estados) != 2 || keycloak.estados[0] == keycloak.estados[1] {
		t.Fatalf("cada entrada debería llevar su propio state: %v", keycloak.estados)
	}
	for _, state := range keycloak.estados {
		if err := auth.VerifyState("un-secreto-de-prueba-largo", state, time.Now()); err != nil {
			t.Fatalf("el state debería estar firmado por nosotros: %v", err)
		}
	}
}

// Sin Keycloak configurado no hay dirección: es una instalación que no tiene ese camino.
func TestStartKeycloakSinCamino(t *testing.T) {
	firmante, _ := auth.NewTokenManager("un-secreto-de-prueba-largo")
	servicio := NewService(&cuentasDeMentira{}, nil, nil, firmante, Config{PublicAppURL: "https://ejemplo.com"})

	if _, err := servicio.StartKeycloak(); !errors.Is(err, ErrOIDCNotConfigured) {
		t.Fatalf("se esperaba ErrOIDCNotConfigured y llegó %v", err)
	}
}

// La vuelta buena deja a la persona en la pantalla de entrada **con el token en el fragmento**, que es
// lo que el navegador no manda al servidor (docs/modules/auth.md, decisiones 13 y 30).
func TestKeycloakCallbackDejaLaSesion(t *testing.T) {
	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{}}
	keycloak := &keycloakDeMentira{datos: auth.DirectoryAccount{
		Origin: auth.OriginKeycloak, ExternalID: "9f1c",
		Email: "ana.keycloak@ejemplo.com", Name: "Ana", LastName: "Pérez",
	}}
	servicio := servicioConKeycloak(t, cuentas, keycloak)

	state, err := auth.SignState("un-secreto-de-prueba-largo", servicio.now())
	if err != nil {
		t.Fatalf("no se pudo firmar el state: %v", err)
	}

	vuelta := servicio.KeycloakCallback(state, "el-codigo")

	if !strings.HasPrefix(vuelta, "https://dev.catalina-support.example.com/login#token=") {
		t.Fatalf("la vuelta no lleva el token al fragmento: %q", vuelta)
	}
	token := strings.TrimPrefix(vuelta, "https://dev.catalina-support.example.com/login#token=")
	if _, err := servicio.session.Verify(token, servicio.now()); err != nil {
		t.Fatalf("el token de la vuelta debería valer: %v", err)
	}

	// Y la cuenta quedó creada con lo que dijo Keycloak.
	if len(cuentas.vinculadas) != 1 || cuentas.vinculadas[0].Origin != auth.OriginKeycloak {
		t.Fatalf("la cuenta no se dejó al día: %+v", cuentas.vinculadas)
	}
}

// **Un `state` que no vale no entra**, y no se canjea nada: es lo que impide que una vuelta fabricada
// por otro sirva para entrar (docs/modules/auth.md, sección 5.3).
func TestKeycloakCallbackConStateQueNoVale(t *testing.T) {
	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{}}
	keycloak := &keycloakDeMentira{}
	servicio := servicioConKeycloak(t, cuentas, keycloak)

	caducado, _ := auth.SignState("un-secreto-de-prueba-largo", servicio.now().Add(-10*time.Minute))

	casos := map[string]string{
		"caducado": caducado,
		"de otro":  "un.state.inventado",
		"vacío":    "",
	}
	for nombre, state := range casos {
		vuelta := servicio.KeycloakCallback(state, "el-codigo")
		if !strings.HasSuffix(vuelta, "#error=auth.oidc.state") {
			t.Fatalf("con el state %s se esperaba el error de state y llegó %q", nombre, vuelta)
		}
	}

	if len(cuentas.vinculadas) != 0 {
		t.Fatalf("con un state que no vale no se puede tocar ninguna cuenta: %+v", cuentas.vinculadas)
	}
}

// Cada fallo del camino se cuenta con su clave, y todos dejan a la persona en la pantalla de entrada:
// esto lo recibe un navegador, así que hasta lo que sale mal tiene que acabar en algún sitio.
func TestKeycloakCallbackConFallos(t *testing.T) {
	casos := []struct {
		nombre   string
		keycloak Keycloak
		clave    string
	}{
		{"sin correo", &keycloakDeMentira{err: ErrOIDCNoEmail}, "auth.oidc.noEmail"},
		{"rechazado", &keycloakDeMentira{err: ErrOIDCRejected}, "auth.oidc.rejected"},
		{"caído", &keycloakDeMentira{err: ErrOIDCUnavailable}, "auth.oidc.unavailable"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			servicio := servicioConKeycloak(t, &cuentasDeMentira{}, caso.keycloak)
			state, _ := auth.SignState("un-secreto-de-prueba-largo", servicio.now())

			if vuelta := servicio.KeycloakCallback(state, "el-codigo"); !strings.HasSuffix(vuelta, "#error="+caso.clave) {
				t.Fatalf("se esperaba el error %s y llegó %q", caso.clave, vuelta)
			}
		})
	}

	// Y sin el método en Keycloak, el mismo camino dice que no hay camino: con la instalación entrando
	// por otro sitio, esa vuelta no se atiende.
	firmante, _ := auth.NewTokenManager("un-secreto-de-prueba-largo")
	sinCamino := NewService(&cuentasDeMentira{}, nil, nil, firmante, Config{
		PublicAppURL: "https://dev.catalina-support.example.com",
		TokenSecret:  "un-secreto-de-prueba-largo",
	})
	sinCamino.SetAccess(accesoDeMentira{metodo: auth.MethodLocal})
	if vuelta := sinCamino.KeycloakCallback("lo-que-sea", "el-codigo"); !strings.HasSuffix(vuelta, "#error=auth.oidc.notConfigured") {
		t.Fatalf("se esperaba el error de camino no configurado y llegó %q", vuelta)
	}

	// Lo mismo al empezar la entrada: con el método en local, `start` no lleva a ninguna parte.
	if _, err := sinCamino.StartKeycloak(); !errors.Is(err, ErrOIDCNotConfigured) {
		t.Fatalf("se esperaba ErrOIDCNotConfigured y llegó %v", err)
	}
}

// Lo que decide `Methods`: con el método en Keycloak, el camino existe y la pantalla ofrece su botón
// **en lugar del formulario**, no además (docs/modules/auth.md, decisión 27).
func TestMethodsConKeycloak(t *testing.T) {
	servicio := servicioConKeycloak(t, &cuentasDeMentira{}, &keycloakDeMentira{})

	caminos, err := servicio.Methods()
	if err != nil {
		t.Fatalf("no se pudieron leer los caminos: %v", err)
	}
	if caminos.Method != auth.MethodKeycloak || caminos.Local || caminos.AD || !caminos.Keycloak {
		t.Fatalf("con el método en Keycloak sólo debería estar ese camino y llegó %+v", caminos)
	}
}
