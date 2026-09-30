package services

import (
	"errors"
	"testing"
)

// servicioConDireccion arma un servicio con la configuración mínima: sólo lo que hace falta para
// comprobar de dónde salen los enlaces.
func servicioConDireccion(delEntorno string, deLaInstalacion string) *Service {
	servicio := NewService(nil, nil, nil, nil, Config{PublicAppURL: delEntorno})
	if deLaInstalacion != "" {
		servicio.SetEnlaces(enlacesDeMentira{direccion: deLaInstalacion})
	}

	return servicio
}

type enlacesDeMentira struct{ direccion string }

func (e enlacesDeMentira) DireccionPublica() string { return e.direccion }

// **La dirección de la configuración manda sobre la del entorno** (docs/modules/settings.md,
// decisión 14): es la que se cambia desde la pantalla, y la del entorno queda de respaldo.
func TestLaDireccionDeLaConfiguracionMandaSobreLaDelEntorno(t *testing.T) {
	servicio := servicioConDireccion("https://del-entorno.example.org", "https://de-la-instalacion.example.org")

	enlace, err := servicio.link("abc123")
	if err != nil {
		t.Fatalf("no debería fallar: %v", err)
	}
	if esperado := "https://de-la-instalacion.example.org/set-password#token=abc123"; enlace != esperado {
		t.Fatalf("enlace = %q, se esperaba %q", enlace, esperado)
	}
}

// Y sin configuración se usa la del entorno: es lo que hace que una instalación que ya la tuviera
// puesta siga funcionando igual.
func TestSinConfiguracionSeUsaLaDelEntorno(t *testing.T) {
	servicio := servicioConDireccion("https://del-entorno.example.org/", "")

	enlace, err := servicio.link("abc123")
	if err != nil {
		t.Fatalf("no debería fallar: %v", err)
	}
	if esperado := "https://del-entorno.example.org/set-password#token=abc123"; enlace != esperado {
		t.Fatalf("enlace = %q, se esperaba %q", enlace, esperado)
	}
}

// **Sin ninguna de las dos no hay enlace, y se dice**: es mejor un error con su clave que un correo
// con un enlace roto que nadie va a poder usar (decisión 14).
func TestSinDireccionPublicaNoHayEnlace(t *testing.T) {
	servicio := servicioConDireccion("", "")

	enlace, err := servicio.link("abc123")
	if !errors.Is(err, ErrSinDireccionPublica) {
		t.Fatalf("err = %v, se esperaba ErrSinDireccionPublica", err)
	}
	if enlace != "" {
		t.Fatalf("no debería haber enlace, y hay %q", enlace)
	}
}

// La vuelta de Keycloak también sale de la configuración, y sin ella se queda en una ruta relativa:
// el navegador la resuelve contra el mismo sitio del que vino, que es mejor que un host inventado.
func TestLaVueltaDeKeycloakSaleDeLaConfiguracion(t *testing.T) {
	conDireccion := servicioConDireccion("", "https://de-la-instalacion.example.org")
	if vuelta := conDireccion.direccionDeVuelta(); vuelta != "https://de-la-instalacion.example.org/login" {
		t.Fatalf("vuelta = %q", vuelta)
	}

	sinDireccion := servicioConDireccion("", "")
	if vuelta := sinDireccion.direccionDeVuelta(); vuelta != "/login" {
		t.Fatalf("vuelta = %q, se esperaba la ruta relativa", vuelta)
	}
}
