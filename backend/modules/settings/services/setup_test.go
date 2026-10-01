package services

import (
	"errors"
	"testing"
	"time"

	"catalina-support/backend/modules/settings/repositories"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/config"
)

// repoDePrueba es el doble del repositorio: devuelve lo que se le diga y no toca ninguna base.
type repoDePrueba struct {
	instalacion repositories.InstallationSettings
	directorio  repositories.DirectorySettings
	keycloak    repositories.KeycloakSettings
}

func (r *repoDePrueba) Installation() (repositories.InstallationSettings, error) {
	return r.instalacion, nil
}

func (r *repoDePrueba) Tickets() (repositories.TicketSettings, error) {
	return repositories.TicketSettings{}, nil
}

func (r *repoDePrueba) Directory() (repositories.DirectorySettings, error) {
	return r.directorio, nil
}

func (r *repoDePrueba) Keycloak() (repositories.KeycloakSettings, error) {
	return r.keycloak, nil
}

func (r *repoDePrueba) UpdateInstallation(map[string]any, *int64) error { return nil }
func (r *repoDePrueba) UpdateDirectory(map[string]any, *int64) error    { return nil }
func (r *repoDePrueba) UpdateKeycloak(map[string]any, *int64) error     { return nil }
func (r *repoDePrueba) UpdateTickets(map[string]any, *int64) error      { return nil }

// proberDePrueba apunta si le han llamado y con qué, y **no sale a la red**.
type proberDePrueba struct {
	directorioLlamado bool
	keycloakLlamado   bool
	directorio        config.Directory
	keycloak          config.OIDC
	err               error
}

func (p *proberDePrueba) ProbeDirectory(directorio config.Directory) error {
	p.directorioLlamado = true
	p.directorio = directorio
	return p.err
}

func (p *proberDePrueba) ProbeKeycloak(keycloak config.OIDC) error {
	p.keycloakLlamado = true
	p.keycloak = keycloak
	return p.err
}

// proberDeCorreoDePrueba recoge el correo con el que se le prueba, sin hablar con nadie.
type proberDeCorreoDePrueba struct {
	llamado bool
	correo  CorreoSaliente
	err     error
}

func (p *proberDeCorreoDePrueba) ProbarCorreo(correo CorreoSaliente) error {
	p.llamado = true
	p.correo = correo
	return p.err
}

// servicioDePrueba arma el servicio con los dobles y la instalación sin sellar.
func servicioDePrueba(t *testing.T) (*Service, *repoDePrueba) {
	t.Helper()

	repo := &repoDePrueba{}
	return NewService(repo, t.TempDir()), repo
}

func sellada() *repoDePrueba {
	cuando := time.Now()
	return &repoDePrueba{instalacion: repositories.InstallationSettings{InstalledAt: &cuando}}
}

// --- la prueba del paso 2 ---------------------------------------------------------------------

// Con la instalación sellada, la prueba del paso 2 no se hace: el asistente ya no existe.
func TestProbarPasoDeLaEntradaSellada(t *testing.T) {
	prober := &proberDePrueba{}
	servicio := NewService(sellada(), t.TempDir())
	servicio.SetProber(prober)

	err := servicio.ProbarPasoDeLaEntrada(PasoDeInstalacion{EntryMethod: auth.MethodAD})

	if !errors.Is(err, ErrYaInstalada) {
		t.Fatalf("se esperaba %v y llegó %v", ErrYaInstalada, err)
	}
	if prober.directorioLlamado || prober.keycloakLlamado {
		t.Fatal("no se puede probar nada con la instalación sellada")
	}
}

// Un método que no es uno de los tres: la misma clave que al guardar.
func TestProbarPasoDeLaEntradaMetodoDesconocido(t *testing.T) {
	servicio, _ := servicioDePrueba(t)

	if err := servicio.ProbarPasoDeLaEntrada(PasoDeInstalacion{EntryMethod: "sso"}); !errors.Is(err, ErrMethodUnknown) {
		t.Fatalf("se esperaba %v y llegó %v", ErrMethodUnknown, err)
	}
}

// AD con servidor pero sin base de búsqueda: la misma clave que al guardar.
func TestProbarPasoDeLaEntradaDirectorioIncompleto(t *testing.T) {
	servicio, _ := servicioDePrueba(t)

	err := servicio.ProbarPasoDeLaEntrada(PasoDeInstalacion{
		EntryMethod: auth.MethodAD,
		Directory:   DirectoryInput{Host: "ldap", Port: "389"},
	})

	if !errors.Is(err, ErrDirectoryIncomplete) {
		t.Fatalf("se esperaba %v y llegó %v", ErrDirectoryIncomplete, err)
	}
}

// Elegir un método sin configurarlo tampoco pasa: es lo que dejaría la instalación sin puerta.
func TestProbarPasoDeLaEntradaMetodoSinConfigurar(t *testing.T) {
	servicio, _ := servicioDePrueba(t)

	if err := servicio.ProbarPasoDeLaEntrada(PasoDeInstalacion{EntryMethod: auth.MethodAD}); !errors.Is(err, ErrMethodNotConfigured) {
		t.Fatalf("se esperaba %v y llegó %v", ErrMethodNotConfigured, err)
	}
}

// Con datos buenos se llama **a la prueba del directorio** y a ninguna otra.
func TestProbarPasoDeLaEntradaDirectorioLlamaAlProber(t *testing.T) {
	prober := &proberDePrueba{}
	servicio, _ := servicioDePrueba(t)
	servicio.SetProber(prober)

	err := servicio.ProbarPasoDeLaEntrada(PasoDeInstalacion{
		EntryMethod: auth.MethodAD,
		Directory: DirectoryInput{
			Host:         "ldap",
			BindDN:       "cn=consulta,dc=empresa,dc=local",
			BindPassword: "secreta",
			SearchBase:   "ou=personas,dc=empresa,dc=local",
		},
	})

	if err != nil {
		t.Fatalf("la prueba no debería fallar: %v", err)
	}
	if !prober.directorioLlamado || prober.keycloakLlamado {
		t.Fatal("se debía probar sólo el directorio")
	}
	if prober.directorio.Host != "ldap" || prober.directorio.Port != "389" || prober.directorio.BindPass != "secreta" {
		t.Fatalf("la configuración que llegó a la prueba no es la esperada: %+v", prober.directorio)
	}
}

// Con Keycloak elegido, la prueba que se llama es la del reino.
func TestProbarPasoDeLaEntradaKeycloakLlamaAlProber(t *testing.T) {
	prober := &proberDePrueba{}
	servicio, _ := servicioDePrueba(t)
	servicio.SetProber(prober)

	err := servicio.ProbarPasoDeLaEntrada(PasoDeInstalacion{
		EntryMethod: auth.MethodKeycloak,
		Keycloak: KeycloakInput{
			Issuer:      "https://sso.empresa.com/realms/empresa",
			ClientID:    "catalina",
			RedirectURI: "https://mesa.empresa.com/login",
		},
	})

	if err != nil {
		t.Fatalf("la prueba no debería fallar: %v", err)
	}
	if prober.directorioLlamado || !prober.keycloakLlamado {
		t.Fatal("se debía probar sólo el reino")
	}
}

// Si la prueba falla, la pantalla recibe la clave del módulo y no el detalle.
func TestProbarPasoDeLaEntradaQueFalla(t *testing.T) {
	servicio, _ := servicioDePrueba(t)
	servicio.SetProber(&proberDePrueba{err: errors.New("dial tcp: connection refused")})

	err := servicio.ProbarPasoDeLaEntrada(PasoDeInstalacion{
		EntryMethod: auth.MethodKeycloak,
		Keycloak: KeycloakInput{
			Issuer:      "https://sso.empresa.com/realms/empresa",
			ClientID:    "catalina",
			RedirectURI: "https://mesa.empresa.com/login",
		},
	})

	if !errors.Is(err, ErrKeycloakUnreachable) {
		t.Fatalf("se esperaba %v y llegó %v", ErrKeycloakUnreachable, err)
	}
	if err.Error() == "dial tcp: connection refused" {
		t.Fatal("el detalle de la prueba no puede salir en la clave")
	}
}

// El método local no tiene a qué conectarse: no hay nada que probar y no se sale a la red.
func TestProbarPasoDeLaEntradaLocalNoPruebaNada(t *testing.T) {
	prober := &proberDePrueba{}
	servicio, _ := servicioDePrueba(t)
	servicio.SetProber(prober)

	if err := servicio.ProbarPasoDeLaEntrada(PasoDeInstalacion{EntryMethod: auth.MethodLocal}); err != nil {
		t.Fatalf("el método local no debería dar error: %v", err)
	}
	if prober.directorioLlamado || prober.keycloakLlamado {
		t.Fatal("el método local no tiene nada que probar")
	}
}

// --- la prueba del correo ---------------------------------------------------------------------

// Sin servidor no hay nada que probar: la misma clave que al guardar.
func TestProbarCorreoIncompleto(t *testing.T) {
	casos := []struct {
		nombre string
		correo MailInput
	}{
		{"sin servidor", MailInput{Port: "587", FromEmail: "mesa@empresa.com"}},
		{"sin puerto", MailInput{Host: "smtp.empresa.com", FromEmail: "mesa@empresa.com"}},
		{"sin remitente", MailInput{Host: "smtp.empresa.com", Port: "587"}},
		{"remitente sin arroba", MailInput{Host: "smtp.empresa.com", Port: "587", FromEmail: "mesa"}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			servicio, _ := servicioDePrueba(t)

			if err := servicio.ProbarCorreoDeInstalacion(caso.correo); !errors.Is(err, ErrCorreoIncompleto) {
				t.Fatalf("se esperaba %v y llegó %v", ErrCorreoIncompleto, err)
			}
		})
	}
}

// Con datos buenos se llama a la prueba del correo, con **lo que hay en pantalla**.
func TestProbarCorreoLlamaAlProber(t *testing.T) {
	prober := &proberDeCorreoDePrueba{}
	servicio, _ := servicioDePrueba(t)
	servicio.SetProberDeCorreo(prober)

	err := servicio.ProbarCorreoDeInstalacion(MailInput{
		Host:      "smtp.empresa.com",
		Port:      "587",
		Secure:    true,
		User:      "mesa",
		Password:  "secreta",
		FromName:  "Mesa de ayuda",
		FromEmail: "mesa@empresa.com",
	})

	if err != nil {
		t.Fatalf("la prueba no debería fallar: %v", err)
	}
	if !prober.llamado {
		t.Fatal("no se llamó a la prueba del correo")
	}
	if prober.correo.Host != "smtp.empresa.com" || prober.correo.Password != "secreta" || !prober.correo.Secure {
		t.Fatalf("el correo que llegó a la prueba no es el esperado: %+v", prober.correo)
	}
}

// La contraseña vacía usa la guardada: es lo que permite probar sin volver a escribir el secreto.
func TestProbarCorreoUsaLaGuardada(t *testing.T) {
	prober := &proberDeCorreoDePrueba{}
	servicio := NewService(&repoDePrueba{instalacion: repositories.InstallationSettings{SMTPPassword: "guardada"}}, t.TempDir())
	servicio.SetProberDeCorreo(prober)

	err := servicio.ProbarCorreoDeInstalacion(MailInput{
		Host:      "smtp.empresa.com",
		Port:      "587",
		FromEmail: "mesa@empresa.com",
	})

	if err != nil {
		t.Fatalf("la prueba no debería fallar: %v", err)
	}
	if prober.correo.Password != "guardada" {
		t.Fatalf("se esperaba la contraseña guardada y llegó %q", prober.correo.Password)
	}
}

// Un servidor que no contesta deja su clave, no el detalle.
func TestProbarCorreoQueFalla(t *testing.T) {
	servicio, _ := servicioDePrueba(t)
	servicio.SetProberDeCorreo(&proberDeCorreoDePrueba{err: errors.New("mail.smtp.connect: dial tcp: refused")})

	err := servicio.ProbarCorreoDeInstalacion(MailInput{
		Host:      "smtp.empresa.com",
		Port:      "587",
		FromEmail: "mesa@empresa.com",
	})

	if !errors.Is(err, ErrCorreoInalcanzable) {
		t.Fatalf("se esperaba %v y llegó %v", ErrCorreoInalcanzable, err)
	}
}

// Con la instalación sellada, la prueba del correo tampoco se hace.
func TestProbarCorreoSellada(t *testing.T) {
	prober := &proberDeCorreoDePrueba{}
	servicio := NewService(sellada(), t.TempDir())
	servicio.SetProberDeCorreo(prober)

	err := servicio.ProbarCorreoDeInstalacion(MailInput{
		Host:      "smtp.empresa.com",
		Port:      "587",
		FromEmail: "mesa@empresa.com",
	})

	if !errors.Is(err, ErrYaInstalada) {
		t.Fatalf("se esperaba %v y llegó %v", ErrYaInstalada, err)
	}
	if prober.llamado {
		t.Fatal("no se puede probar nada con la instalación sellada")
	}
}
