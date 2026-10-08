package services

import (
	"errors"
	"testing"

	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/config"
)

// cuentasDeMentira es el módulo de cuentas visto desde aquí, sin base de datos: lo que se prueba es
// **por dónde entra** cada quien, no lo que la tabla guarda
// (docs/modules/auth.md, secciones 5.1 a 5.4).
type cuentasDeMentira struct {
	// porCorreo es lo que hay en la tabla: la clave es el correo en minúsculas.
	porCorreo map[string]auth.Account
	// vinculadas son las veces que el directorio ha dejado una cuenta al día, con lo que traía.
	vinculadas []auth.DirectoryAccount
	// ultimaEntrada cuenta las veces que se ha apuntado la fecha de entrada.
	ultimaEntrada int
}

func (c *cuentasDeMentira) ByID(id int64) (auth.Account, error) {
	for _, cuenta := range c.porCorreo {
		if cuenta.ID == id {
			return cuenta, nil
		}
	}
	return auth.Account{}, auth.ErrAccountNotFound
}

func (c *cuentasDeMentira) ByEmail(email string) (auth.Account, error) {
	cuenta, hay := c.porCorreo[email]
	if !hay {
		return auth.Account{}, auth.ErrAccountNotFound
	}
	return cuenta, nil
}

func (c *cuentasDeMentira) SetPasswordHash(int64, string) error { return nil }

func (c *cuentasDeMentira) TouchLastLogin(int64) error {
	c.ultimaEntrada++
	return nil
}

func (c *cuentasDeMentira) UpsertFromDirectory(datos auth.DirectoryAccount) (auth.Account, error) {
	c.vinculadas = append(c.vinculadas, datos)

	// Lo que hace `users` de verdad, reducido a lo que se mira aquí: la cuenta queda con el correo y
	// el identificador del directorio, sin contraseña local.
	cuenta, hay := c.porCorreo[datos.Email]
	if !hay {
		cuenta = auth.Account{Email: datos.Email, Role: auth.RoleUsuario}
	}
	cuenta.Name = datos.Name
	cuenta.LastName = datos.LastName
	cuenta.Origin = datos.Origin
	cuenta.ExternalID = datos.ExternalID
	cuenta.PasswordHash = ""
	cuenta.IsActive = true
	if cuenta.ID == 0 {
		cuenta.ID = int64(len(c.porCorreo) + 1)
	}
	c.porCorreo[datos.Email] = cuenta

	return cuenta, nil
}

// directorioDeMentira es el directorio: dice a quién conoce y qué contesta.
type directorioDeMentira struct {
	// personas es lo que hay en el directorio, por correo.
	personas map[string]struct {
		password string
		datos    auth.DirectoryAccount
	}
	// err, si no es nulo, es lo que devuelve el directorio: es el directorio caído.
	err error
	// llamadas cuenta las veces que se ha preguntado al directorio.
	llamadas int
}

func (d *directorioDeMentira) Login(email, password string) (auth.DirectoryAccount, error) {
	d.llamadas++
	if d.err != nil {
		return auth.DirectoryAccount{}, d.err
	}

	persona, hay := d.personas[email]
	if !hay || persona.password != password {
		return auth.DirectoryAccount{}, ErrDirectoryRejected
	}
	return persona.datos, nil
}

// Knows dice si el directorio conoce a alguien con ese correo: es lo que pregunta `users` al reactivar
// una cuenta de AD.
func (d *directorioDeMentira) Knows(email string) (bool, error) {
	if d.err != nil {
		return false, d.err
	}

	_, hay := d.personas[email]

	return hay, nil
}

func (d *directorioDeMentira) Configured() bool { return true }

// accesoDeMentira es la configuración de la instalación vista desde aquí: dice **cómo se entra** —el
// método y sus dos configuraciones— sin base de datos de por medio. Lo que se prueba es por dónde
// entra cada quien, no lo que la tabla guarda (docs/modules/settings.md, sección 5.8).
type accesoDeMentira struct {
	metodo     string
	directorio config.Directory
	keycloak   config.OIDC
	// err, si no es nulo, es lo que devuelve la configuración: la base que no responde.
	err error
}

func (a accesoDeMentira) Access() (config.Access, error) {
	if a.err != nil {
		return config.Access{}, a.err
	}

	return config.Access{Method: a.metodo, Directory: a.directorio, OIDC: a.keycloak}, nil
}

// servicioDePrueba arma el módulo con las cuentas, el método que está puesto y, si se pasa, un
// directorio de mentira.
func servicioDePrueba(t *testing.T, cuentas *cuentasDeMentira, directorio Directorio, metodo string) *Service {
	t.Helper()

	firmante, err := auth.NewTokenManager("un-secreto-de-prueba-largo")
	if err != nil {
		t.Fatalf("no se pudo construir el firmante: %v", err)
	}

	servicio := NewService(cuentas, nil, nil, firmante, Config{AdminPassword: "la-de-fabrica-larga"})
	servicio.SetAccess(accesoDeMentira{
		metodo:     metodo,
		directorio: config.Directory{Host: "ldap", Port: "389", SearchBase: "dc=ejemplo,dc=com", UserFilter: "(mail=%s)"},
	})
	if directorio != nil {
		servicio.SetDirectory(directorio)
	}

	return servicio
}

// El alta automática: quien no tiene cuenta y el directorio conoce, entra con papel `usuario`.
func TestEntraPorElDirectorioYSeDaDeAlta(t *testing.T) {
	directorio := &directorioDeMentira{personas: map[string]struct {
		password string
		datos    auth.DirectoryAccount
	}{
		"ana.directorio@ejemplo.com": {"una-contraseña-larga", auth.DirectoryAccount{
			Origin: auth.OriginAD, ExternalID: "ana.directorio",
			Email: "ana.directorio@ejemplo.com", Name: "Ana", LastName: "Pérez",
		}},
	}}

	servicio := servicioDePrueba(t, &cuentasDeMentira{porCorreo: map[string]auth.Account{}}, directorio, auth.MethodAD)

	sesion, err := servicio.Login("ana.directorio@ejemplo.com", "una-contraseña-larga")
	if err != nil {
		t.Fatalf("debería haber entrado: %v", err)
	}
	if sesion.Token == "" {
		t.Fatal("la sesión tiene que llevar token")
	}
	if sesion.Account.Role != auth.RoleUsuario {
		t.Fatalf("el alta automática es con papel `usuario` y llegó %q", sesion.Account.Role)
	}
	if sesion.Account.Origin != auth.OriginAD {
		t.Fatalf("la cuenta debería quedar como del directorio y quedó %q", sesion.Account.Origin)
	}
}

// **El vínculo**: una cuenta local con ese correo y el directorio conoce a esa persona. Entra con la
// contraseña del directorio, la cuenta pasa a ser del directorio y su contraseña local deja de servir
// (docs/modules/auth.md, sección 5.4).
func TestVinculaUnaCuentaLocalAlEntrarPorElDirectorio(t *testing.T) {
	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"marta.directorio@ejemplo.com": {
			ID: 7, Name: "Marta", LastName: "Local", Email: "marta.directorio@ejemplo.com",
			PasswordHash: "$2a$12$loquesea", Role: auth.RoleSoporte, Origin: auth.OriginLocal,
			IsActive: true,
		},
	}}
	directorio := &directorioDeMentira{personas: map[string]struct {
		password string
		datos    auth.DirectoryAccount
	}{
		"marta.directorio@ejemplo.com": {"la-del-directorio", auth.DirectoryAccount{
			Origin: auth.OriginAD, ExternalID: "marta.directorio",
			Email: "marta.directorio@ejemplo.com", Name: "Marta", LastName: "Ruiz",
		}},
	}}

	servicio := servicioDePrueba(t, cuentas, directorio, auth.MethodAD)

	// La contraseña local de esa cuenta no es la que llega, así que se pregunta al directorio.
	sesion, err := servicio.Login("marta.directorio@ejemplo.com", "la-del-directorio")
	if err != nil {
		t.Fatalf("debería haber entrado y vinculado la cuenta: %v", err)
	}

	if len(cuentas.vinculadas) != 1 {
		t.Fatalf("se esperaba un vínculo y hubo %d", len(cuentas.vinculadas))
	}
	if sesion.Account.Origin != auth.OriginAD {
		t.Fatalf("la cuenta debería quedar como del directorio y quedó %q", sesion.Account.Origin)
	}
	if sesion.Account.PasswordHash != "" {
		t.Fatal("al vincular, la contraseña local no puede seguir guardada")
	}
	// El papel no se toca: la persona es la misma y lo que cambia es de dónde viene su contraseña.
	if sesion.Account.Role != auth.RoleSoporte {
		t.Fatalf("el papel debería seguir siendo `soporte` y quedó %q", sesion.Account.Role)
	}
	// Y los datos del directorio mandan.
	if sesion.Account.LastName != "Ruiz" {
		t.Fatalf("los apellidos deberían venir del directorio y quedaron %q", sesion.Account.LastName)
	}
}

// Una cuenta local con su contraseña buena entra **sin preguntar al directorio**: sería mandar a la
// red, en cada entrada, algo que ya se sabe aquí.
func TestUnaCuentaLocalNoPreguntaAlDirectorio(t *testing.T) {
	hash, err := HashPassword("la-local-larga")
	if err != nil {
		t.Fatalf("no se pudo cifrar: %v", err)
	}

	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"ana@ejemplo.com": {
			ID: 1, Name: "Ana", Email: "ana@ejemplo.com", PasswordHash: hash,
			Role: auth.RoleUsuario, Origin: auth.OriginLocal, IsActive: true,
		},
	}}
	directorio := &directorioDeMentira{personas: map[string]struct {
		password string
		datos    auth.DirectoryAccount
	}{}}

	servicio := servicioDePrueba(t, cuentas, directorio, auth.MethodLocal)

	if _, err := servicio.Login("ana@ejemplo.com", "la-local-larga"); err != nil {
		t.Fatalf("debería haber entrado: %v", err)
	}
	if directorio.llamadas != 0 {
		t.Fatalf("no se debería haber preguntado al directorio y se preguntó %d veces", directorio.llamadas)
	}
}

// Y con la contraseña mala, sin directorio al que preguntar, es la contraseña equivocada de siempre.
func TestSinDirectorioLaContrasenaMalaNoEntra(t *testing.T) {
	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"ana@ejemplo.com": {
			ID: 1, Name: "Ana", Email: "ana@ejemplo.com", PasswordHash: "$2a$12$loquesea",
			Role: auth.RoleUsuario, Origin: auth.OriginLocal, IsActive: true,
		},
	}}

	servicio := servicioDePrueba(t, cuentas, nil, auth.MethodLocal)

	if _, err := servicio.Login("ana@ejemplo.com", "la-que-no-es"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials y llegó %v", err)
	}
	if _, err := servicio.Login("quien.sea@ejemplo.com", "lo-que-sea"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials y llegó %v", err)
	}
}

// **Una cuenta de AD con el directorio caído lo dice tal cual**: se sabe de dónde es esa persona, así
// que el mensaje es «el directorio no responde», no «la contraseña está mal». Confundirlos manda a
// alguien a buscar una contraseña que no ha perdido (docs/modules/auth.md, sección 5.2).
func TestCuentaDeADConElDirectorioCaido(t *testing.T) {
	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"ana.directorio@ejemplo.com": {
			ID: 2, Name: "Ana", Email: "ana.directorio@ejemplo.com",
			Role: auth.RoleUsuario, Origin: auth.OriginAD, ExternalID: "ana.directorio",
			IsActive: true,
		},
	}}
	directorio := &directorioDeMentira{err: ErrDirectoryUnavailable}

	servicio := servicioDePrueba(t, cuentas, directorio, auth.MethodAD)

	if _, err := servicio.Login("ana.directorio@ejemplo.com", "una-contraseña-larga"); !errors.Is(err, ErrDirectoryUnavailable) {
		t.Fatalf("se esperaba ErrDirectoryUnavailable y llegó %v", err)
	}
}

// **Y con un correo que no es de nadie, el directorio caído no se cuenta**: no se sabe si esa persona
// es del directorio, y decirlo sería contarle a un desconocido cómo está la red de la casa.
func TestCorreoDesconocidoConElDirectorioCaido(t *testing.T) {
	directorio := &directorioDeMentira{err: ErrDirectoryUnavailable}
	servicio := servicioDePrueba(t, &cuentasDeMentira{porCorreo: map[string]auth.Account{}}, directorio, auth.MethodAD)

	if _, err := servicio.Login("quien.sea@ejemplo.com", "lo-que-sea"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials y llegó %v", err)
	}
}

// Lo mismo para una cuenta **local** cuya contraseña no vale y que además está en el directorio: si el
// directorio no responde, la entrada local sigue diciendo lo de siempre.
func TestCuentaLocalConElDirectorioCaidoNoCambiaSuMensaje(t *testing.T) {
	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"ana@ejemplo.com": {
			ID: 1, Name: "Ana", Email: "ana@ejemplo.com", PasswordHash: "$2a$12$loquesea",
			Role: auth.RoleUsuario, Origin: auth.OriginLocal, IsActive: true,
		},
	}}
	directorio := &directorioDeMentira{err: ErrDirectoryUnavailable}
	servicio := servicioDePrueba(t, cuentas, directorio, auth.MethodAD)

	if _, err := servicio.Login("ana@ejemplo.com", "la-que-no-es"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials y llegó %v", err)
	}
}

// **Con el método en local, una cuenta de directorio no tiene puerta**: su contraseña no está aquí, y
// no hay a quién preguntar. Es la otra cara de «un método a la vez».
func TestConElMetodoEnLocalUnaCuentaDeDirectorioNoEntra(t *testing.T) {
	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"ana.directorio@ejemplo.com": {
			ID: 2, Name: "Ana", Email: "ana.directorio@ejemplo.com",
			Role: auth.RoleUsuario, Origin: auth.OriginAD, IsActive: true,
		},
	}}
	servicio := servicioDePrueba(t, cuentas, nil, auth.MethodLocal)

	if _, err := servicio.Login("ana.directorio@ejemplo.com", "una-contraseña-larga"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials y llegó %v", err)
	}
}

// Una cuenta de Keycloak no entra por la pantalla de correo y contraseña: su camino es la vuelta por
// el navegador (docs/modules/auth.md, sección 5.3).
func TestCuentaDeKeycloakNoEntraPorLaPantalla(t *testing.T) {
	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"ana.directorio@ejemplo.com": {
			ID: 3, Name: "Ana", Email: "ana.directorio@ejemplo.com",
			Role: auth.RoleUsuario, Origin: auth.OriginKeycloak, IsActive: true,
		},
	}}
	servicio := servicioDePrueba(t, cuentas, &directorioDeMentira{}, auth.MethodAD)

	if _, err := servicio.Login("ana.directorio@ejemplo.com", "una-contraseña-larga"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials y llegó %v", err)
	}
}

// Una cuenta desactivada que **sabe su contraseña** se lleva su propio mensaje: no puede arreglarlo
// sola y tiene que llamar a Soporte. Es información para quien ya ha demostrado quién es.
func TestCuentaDesactivadaLoDice(t *testing.T) {
	hash, err := HashPassword("la-local-larga")
	if err != nil {
		t.Fatalf("no se pudo cifrar: %v", err)
	}

	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"ana@ejemplo.com": {
			ID: 1, Name: "Ana", Email: "ana@ejemplo.com", PasswordHash: hash,
			Role: auth.RoleUsuario, Origin: auth.OriginLocal, IsActive: false,
		},
	}}
	servicio := servicioDePrueba(t, cuentas, &directorioDeMentira{}, auth.MethodLocal)

	if _, err := servicio.Login("ana@ejemplo.com", "la-local-larga"); !errors.Is(err, ErrAccountInactive) {
		t.Fatalf("se esperaba ErrAccountInactive y llegó %v", err)
	}
}

// **La cuenta de fábrica entra siempre, y lo primero**: no está en la tabla, su contraseña vive en la
// configuración y es la única puerta que no se puede cerrar. Se prueba con el método en Keycloak, que
// es el caso en el que nadie más entra por esta pantalla: si el administrador elige mal el método,
// sigue habiendo por dónde volver a cambiarlo (docs/usuarios-y-permisos.md, sección 8).
func TestLaCuentaDeFabricaEntraPorLaConfiguracion(t *testing.T) {
	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{}}
	servicio := servicioDePrueba(t, cuentas, nil, auth.MethodKeycloak)

	sesion, err := servicio.Login("admin", "la-de-fabrica-larga")
	if err != nil {
		t.Fatalf("la cuenta de fábrica debería entrar: %v", err)
	}
	if !sesion.Factory {
		t.Fatal("la sesión debería ser la de la cuenta de fábrica")
	}
	if sesion.Account.Role != auth.RoleAdministrador {
		t.Fatalf("el papel debería ser `administrador` y llegó %q", sesion.Account.Role)
	}

	if _, err := servicio.Login("admin", "la-que-no-es"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials y llegó %v", err)
	}
}

// **Con el método en AD no se entra con la contraseña local**, ni siquiera teniéndola buena: el camino
// de la instalación es el directorio, y una cuenta local —Soporte, Desarrollo— se queda fuera hasta que
// el método vuelva a ser el local. Es la consecuencia de «un método a la vez», y la cuenta de fábrica es
// la excepción (docs/modules/settings.md, sección 5.8).
func TestConElMetodoEnADNoSeEntraConLaContrasenaLocal(t *testing.T) {
	hash, err := HashPassword("la-local-larga")
	if err != nil {
		t.Fatalf("no se pudo cifrar: %v", err)
	}

	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"marta@ejemplo.com": {
			ID: 4, Name: "Marta", Email: "marta@ejemplo.com", PasswordHash: hash,
			Role: auth.RoleSoporte, Origin: auth.OriginLocal, IsActive: true,
		},
	}}
	// El directorio no conoce a esa persona: es una cuenta local, y con este método no hay puerta.
	servicio := servicioDePrueba(t, cuentas, &directorioDeMentira{}, auth.MethodAD)

	if _, err := servicio.Login("marta@ejemplo.com", "la-local-larga"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials y llegó %v", err)
	}
}

// **Y con el método en Keycloak tampoco**: ese camino es una vuelta por el navegador y no este
// endpoint, así que aquí no hay contraseña que comparar (docs/modules/auth.md, sección 5.3).
func TestConElMetodoEnKeycloakNoSeEntraConLaContrasenaLocal(t *testing.T) {
	hash, err := HashPassword("la-local-larga")
	if err != nil {
		t.Fatalf("no se pudo cifrar: %v", err)
	}

	cuentas := &cuentasDeMentira{porCorreo: map[string]auth.Account{
		"marta@ejemplo.com": {
			ID: 4, Name: "Marta", Email: "marta@ejemplo.com", PasswordHash: hash,
			Role: auth.RoleSoporte, Origin: auth.OriginLocal, IsActive: true,
		},
	}}
	servicio := servicioDePrueba(t, cuentas, nil, auth.MethodKeycloak)

	if _, err := servicio.Login("marta@ejemplo.com", "la-local-larga"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("se esperaba ErrInvalidCredentials y llegó %v", err)
	}
}

// Lo que la pantalla de entrada pregunta: el método que está puesto y cuál de los tres caminos está
// abierto. **Uno solo**, y con el método en AD sigue habiendo formulario —es el mismo, lo que cambia es
// quién lo contesta— (docs/modules/auth.md, decisión 27).
func TestMethods(t *testing.T) {
	casos := []struct {
		metodo   string
		local    bool
		ad       bool
		keycloak bool
	}{
		{auth.MethodLocal, true, false, false},
		{auth.MethodAD, false, true, false},
		{auth.MethodKeycloak, false, false, true},
	}

	for _, caso := range casos {
		t.Run(caso.metodo, func(t *testing.T) {
			servicio := servicioDePrueba(t, &cuentasDeMentira{}, &directorioDeMentira{}, caso.metodo)

			caminos, err := servicio.Methods()
			if err != nil {
				t.Fatalf("no se pudieron leer los caminos: %v", err)
			}
			if caminos.Method != caso.metodo {
				t.Fatalf("el método debería ser %q y llegó %q", caso.metodo, caminos.Method)
			}
			if caminos.Local != caso.local || caminos.AD != caso.ad || caminos.Keycloak != caso.keycloak {
				t.Fatalf("con el método en %s los caminos no cuadran: %+v", caso.metodo, caminos)
			}
		})
	}
}

// Sin configuración de la instalación —una instalación recién puesta, que todavía no tiene su fila— sólo
// existe el camino local: es lo que hace que no se quede sin puerta.
func TestSinConfiguracionSoloHayCaminoLocal(t *testing.T) {
	firmante, err := auth.NewTokenManager("un-secreto-de-prueba-largo")
	if err != nil {
		t.Fatalf("no se pudo construir el firmante: %v", err)
	}

	servicio := NewService(&cuentasDeMentira{}, nil, nil, firmante, Config{AdminPassword: "la-de-fabrica-larga"})

	caminos, err := servicio.Methods()
	if err != nil {
		t.Fatalf("no se pudieron leer los caminos: %v", err)
	}
	if caminos.Method != auth.MethodLocal || !caminos.Local || caminos.AD || caminos.Keycloak {
		t.Fatalf("sin configuración sólo debería haber camino local y llegó %+v", caminos)
	}
}

// `Knows` es lo que pregunta `users` antes de reactivar una cuenta de AD: se contesta con la
// configuración guardada, y sin directorio configurado no hay a quién preguntar.
func TestKnows(t *testing.T) {
	servicio := servicioDePrueba(t, &cuentasDeMentira{}, &directorioDeMentira{personas: map[string]struct {
		password string
		datos    auth.DirectoryAccount
	}{
		"ana.directorio@ejemplo.com": {"la-del-directorio", auth.DirectoryAccount{}},
	}}, auth.MethodAD)

	if loConoce, err := servicio.Knows("ana.directorio@ejemplo.com"); err != nil || !loConoce {
		t.Fatalf("el directorio debería conocerla y llegó %v, %v", loConoce, err)
	}
	if loConoce, err := servicio.Knows("quien.sea@ejemplo.com"); err != nil || loConoce {
		t.Fatalf("el directorio no la conoce y llegó %v, %v", loConoce, err)
	}

	// Y sin directorio configurado, la respuesta es que no se puede comprobar.
	sinDirectorio := NewService(&cuentasDeMentira{}, nil, nil, servicio.session, Config{})
	sinDirectorio.SetAccess(accesoDeMentira{metodo: auth.MethodLocal})
	if _, err := sinDirectorio.Knows("ana.directorio@ejemplo.com"); !errors.Is(err, ErrDirectoryUnavailable) {
		t.Fatalf("se esperaba ErrDirectoryUnavailable y llegó %v", err)
	}
}
