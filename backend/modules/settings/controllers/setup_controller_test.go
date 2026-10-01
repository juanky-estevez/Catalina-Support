package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"catalina-support/backend/modules/settings/repositories"
	"catalina-support/backend/modules/settings/services"
	"catalina-support/backend/shared/config"
)

// repoDePrueba es el doble del repositorio: contesta lo que se le diga y no toca la base.
type repoDePrueba struct {
	instalacion repositories.InstallationSettings
}

func (r *repoDePrueba) Installation() (repositories.InstallationSettings, error) {
	return r.instalacion, nil
}

func (r *repoDePrueba) Tickets() (repositories.TicketSettings, error) {
	return repositories.TicketSettings{}, nil
}

func (r *repoDePrueba) Directory() (repositories.DirectorySettings, error) {
	return repositories.DirectorySettings{}, nil
}

func (r *repoDePrueba) Keycloak() (repositories.KeycloakSettings, error) {
	return repositories.KeycloakSettings{}, nil
}

func (r *repoDePrueba) UpdateInstallation(map[string]any, *int64) error { return nil }
func (r *repoDePrueba) UpdateDirectory(map[string]any, *int64) error    { return nil }
func (r *repoDePrueba) UpdateKeycloak(map[string]any, *int64) error     { return nil }
func (r *repoDePrueba) UpdateTickets(map[string]any, *int64) error      { return nil }

type proberDePrueba struct {
	llamado bool
	err     error
}

func (p *proberDePrueba) ProbeDirectory(config.Directory) error {
	p.llamado = true
	return p.err
}

func (p *proberDePrueba) ProbeKeycloak(config.OIDC) error {
	p.llamado = true
	return p.err
}

type proberDeCorreoDePrueba struct {
	llamado bool
	correo  services.CorreoSaliente
	err     error
}

func (p *proberDeCorreoDePrueba) ProbarCorreo(correo services.CorreoSaliente) error {
	p.llamado = true
	p.correo = correo
	return p.err
}

// sellada devuelve un repositorio con el sello puesto.
func sellada() *repoDePrueba {
	cuando := time.Now()
	return &repoDePrueba{instalacion: repositories.InstallationSettings{InstalledAt: &cuando}}
}

// peticionDePrueba arma la petición de un endpoint del asistente.
func peticionDePrueba(ruta, cuerpo string) *http.Request {
	return httptest.NewRequest(http.MethodPost, ruta, strings.NewReader(cuerpo))
}

// claveDeLaRespuesta lee la clave del cuerpo de error.
func claveDeLaRespuesta(t *testing.T, respuesta *httptest.ResponseRecorder) string {
	t.Helper()

	var cuerpo map[string]string
	if err := json.Unmarshal(respuesta.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("la respuesta no es JSON: %v", err)
	}

	return cuerpo["error"]
}

// Con la instalación sellada, las dos pruebas contestan **409** con su clave, como el resto del
// asistente (docs/primer-arranque.md, secciones 2 y 6).
func TestLasPruebasDelAsistenteEstanSelladas(t *testing.T) {
	casos := []struct {
		nombre  string
		ruta    string
		cuerpo  string
		handler func(*SetupController, http.ResponseWriter, *http.Request)
	}{
		{
			"prueba de la entrada",
			"/api/setup/entry/test",
			`{"entryMethod":"local"}`,
			func(c *SetupController, w http.ResponseWriter, r *http.Request) { c.TestEntry(w, r) },
		},
		{
			"prueba del correo",
			"/api/setup/mail/test",
			`{"mail":{"host":"smtp.empresa.com","port":"587","fromEmail":"mesa@empresa.com"}}`,
			func(c *SetupController, w http.ResponseWriter, r *http.Request) { c.TestMail(w, r) },
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			servicio := services.NewService(sellada(), t.TempDir())
			controlador := NewSetupController(servicio)

			respuesta := httptest.NewRecorder()
			caso.handler(controlador, respuesta, peticionDePrueba(caso.ruta, caso.cuerpo))

			if respuesta.Code != http.StatusConflict {
				t.Fatalf("el código es %d y se esperaba 409", respuesta.Code)
			}
			if clave := claveDeLaRespuesta(t, respuesta); clave != "setup.alreadyInstalled" {
				t.Fatalf("la clave es %q y se esperaba setup.alreadyInstalled", clave)
			}
		})
	}
}

// Con datos malos, cada prueba contesta la misma clave que contestaría al guardar.
func TestLasPruebasDelAsistenteValidanComoAlGuardar(t *testing.T) {
	casos := []struct {
		nombre   string
		cuerpo   string
		esperado string
		handler  func(*SetupController, http.ResponseWriter, *http.Request)
	}{
		{
			"método inventado",
			`{"entryMethod":"sso"}`,
			"settings.method.unknown",
			func(c *SetupController, w http.ResponseWriter, r *http.Request) { c.TestEntry(w, r) },
		},
		{
			"método sin configurar",
			`{"entryMethod":"ad"}`,
			"settings.method.notConfigured",
			func(c *SetupController, w http.ResponseWriter, r *http.Request) { c.TestEntry(w, r) },
		},
		{
			"directorio a medias",
			`{"entryMethod":"ad","directory":{"host":"ldap","port":"389"}}`,
			"settings.directory.incomplete",
			func(c *SetupController, w http.ResponseWriter, r *http.Request) { c.TestEntry(w, r) },
		},
		{
			"keycloak a medias",
			`{"entryMethod":"keycloak","keycloak":{"issuer":"https://sso.empresa.com/realms/x"}}`,
			"settings.keycloak.incomplete",
			func(c *SetupController, w http.ResponseWriter, r *http.Request) { c.TestEntry(w, r) },
		},
		{
			"correo a medias",
			`{"mail":{"host":"smtp.empresa.com","fromEmail":"mesa@empresa.com"}}`,
			"setup.mail.incomplete",
			func(c *SetupController, w http.ResponseWriter, r *http.Request) { c.TestMail(w, r) },
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			servicio := services.NewService(&repoDePrueba{}, t.TempDir())
			controlador := NewSetupController(servicio)

			respuesta := httptest.NewRecorder()
			caso.handler(controlador, respuesta, peticionDePrueba("/api/setup/x", caso.cuerpo))

			if respuesta.Code != http.StatusUnprocessableEntity {
				t.Fatalf("el código es %d y se esperaba 422", respuesta.Code)
			}
			if clave := claveDeLaRespuesta(t, respuesta); clave != caso.esperado {
				t.Fatalf("la clave es %q y se esperaba %q", clave, caso.esperado)
			}
		})
	}
}

// Con datos buenos, la prueba llama **a la prueba del módulo** y contesta lo mismo que Configuración.
func TestLasPruebasDelAsistenteLlamanALaPrueba(t *testing.T) {
	t.Run("la entrada", func(t *testing.T) {
		prober := &proberDePrueba{}
		servicio := services.NewService(&repoDePrueba{}, t.TempDir())
		servicio.SetProber(prober)
		controlador := NewSetupController(servicio)

		respuesta := httptest.NewRecorder()
		controlador.TestEntry(respuesta, peticionDePrueba("/api/setup/entry/test",
			`{"entryMethod":"ad","directory":{"host":"ldap","port":"389","searchBase":"ou=personas,dc=empresa,dc=local"}}`))

		if respuesta.Code != http.StatusOK {
			t.Fatalf("el código es %d y se esperaba 200: %s", respuesta.Code, respuesta.Body.String())
		}
		if !prober.llamado {
			t.Fatal("no se llamó a la prueba del directorio")
		}

		var cuerpo map[string]string
		_ = json.Unmarshal(respuesta.Body.Bytes(), &cuerpo)
		if cuerpo["status"] != "ok" {
			t.Fatalf("la respuesta es %q y se esperaba status ok", respuesta.Body.String())
		}
	})

	t.Run("el correo", func(t *testing.T) {
		prober := &proberDeCorreoDePrueba{}
		servicio := services.NewService(&repoDePrueba{}, t.TempDir())
		servicio.SetProberDeCorreo(prober)
		controlador := NewSetupController(servicio)

		respuesta := httptest.NewRecorder()
		controlador.TestMail(respuesta, peticionDePrueba("/api/setup/mail/test",
			`{"mail":{"host":"smtp.empresa.com","port":"587","user":"mesa","password":"secreta","fromEmail":"mesa@empresa.com"}}`))

		if respuesta.Code != http.StatusOK {
			t.Fatalf("el código es %d y se esperaba 200: %s", respuesta.Code, respuesta.Body.String())
		}
		if !prober.llamado {
			t.Fatal("no se llamó a la prueba del correo")
		}
		if prober.correo.Host != "smtp.empresa.com" || prober.correo.Password != "secreta" {
			t.Fatalf("el correo que llegó a la prueba no es el esperado: %+v", prober.correo)
		}
	})
}

// Una prueba que no contesta deja su clave, no el detalle técnico.
func TestLasPruebasDelAsistenteCuentanElFallo(t *testing.T) {
	t.Run("el directorio", func(t *testing.T) {
		servicio := services.NewService(&repoDePrueba{}, t.TempDir())
		servicio.SetProber(&proberDePrueba{err: errors.New("dial tcp: refused")})
		controlador := NewSetupController(servicio)

		respuesta := httptest.NewRecorder()
		controlador.TestEntry(respuesta, peticionDePrueba("/api/setup/entry/test",
			`{"entryMethod":"ad","directory":{"host":"ldap","port":"389","searchBase":"ou=personas,dc=empresa,dc=local"}}`))

		if respuesta.Code != http.StatusUnprocessableEntity {
			t.Fatalf("el código es %d y se esperaba 422", respuesta.Code)
		}
		if clave := claveDeLaRespuesta(t, respuesta); clave != "settings.directory.unreachable" {
			t.Fatalf("la clave es %q y se esperaba settings.directory.unreachable", clave)
		}
	})

	t.Run("el correo", func(t *testing.T) {
		servicio := services.NewService(&repoDePrueba{}, t.TempDir())
		servicio.SetProberDeCorreo(&proberDeCorreoDePrueba{err: errors.New("mail.smtp.connect: refused")})
		controlador := NewSetupController(servicio)

		respuesta := httptest.NewRecorder()
		controlador.TestMail(respuesta, peticionDePrueba("/api/setup/mail/test",
			`{"mail":{"host":"smtp.empresa.com","port":"587","fromEmail":"mesa@empresa.com"}}`))

		if respuesta.Code != http.StatusUnprocessableEntity {
			t.Fatalf("el código es %d y se esperaba 422", respuesta.Code)
		}
		if clave := claveDeLaRespuesta(t, respuesta); clave != "setup.mail.unreachable" {
			t.Fatalf("la clave es %q y se esperaba setup.mail.unreachable", clave)
		}
	})
}
