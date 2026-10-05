package services

import (
	"errors"
	"strings"
	"time"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/shared/auth"
)

// ErrYaInstalada: la instalación ya tiene su sello, así que **el asistente no se puede volver a
// recorrer** ni por la pantalla ni por la API. Es lo que impide que alguien reescriba la configuración
// de una instalación en marcha (docs/primer-arranque.md, sección 2).
var ErrYaInstalada = errors.New("setup.alreadyInstalled")

// ErrPasoIncompleto: al paso le falta algo de lo suyo. Cada paso comprueba **sólo lo suyo**, porque el
// resto de la configuración todavía no está puesta: el asistente va de vacío a completo.
var ErrPasoIncompleto = errors.New("setup.step.incomplete")

// ErrCorreoIncompleto: los datos del correo saliente no están completos o no son válidos.
var ErrCorreoIncompleto = errors.New("setup.mail.incomplete")

// ErrCorreoInalcanzable: el servidor de correo no ha contestado, o no ha aceptado la autenticación.
// El detalle queda en el log; a la pantalla le llega que esa configuración no sirve.
var ErrCorreoInalcanzable = errors.New("setup.mail.unreachable")

// ProberDeIA es lo que este módulo necesita para poder decir si el motor de IA responde, y para
// probarlo desde Configuración. Lo declara quien lo usa y lo cumple el módulo `ai`, que es el que
// sabe hablar con el motor.
type ProberDeIA interface {
	Disponible() bool
	// Probar comprueba que el motor de esa dirección contesta. Se le pasa la dirección a propósito:
	// la prueba es de **lo que hay en pantalla**, y no de lo que esté guardado.
	Probar(url string) error
}

// ProberDeCorreo es lo que este módulo necesita para **probar el correo saliente sin mandar ningún
// correo**: abre la conexión y valida la autenticación —host, puerto, TLS, usuario y contraseña— y
// no entrega ningún mensaje. Lo declara quien lo usa y lo cumple el módulo `mail`, que es el que sabe
// hablar con un servidor de correo (docs/primer-arranque.md, sección 3).
type ProberDeCorreo interface {
	ProbarCorreo(correo CorreoSaliente) error
}

// CorreoSaliente es el correo saliente **con sus secretos**, para probarlo: lo que se le manda al
// módulo de correo es lo que hay en pantalla, no lo que esté guardado. Declarado aquí y traducido en
// el cableado, como todo lo que cruza de un módulo a otro.
type CorreoSaliente struct {
	Host      string
	Port      string
	Secure    bool
	User      string
	Password  string
	FromName  string
	FromEmail string
}

// EstadoDeInstalacion es lo que la vista de instalación necesita para empezar: si ya está instalada
// —y entonces no se enseña— y qué hay puesto, para poder **seguir donde se dejó**.
//
// **Sin secretos**: ni la contraseña del directorio ni la de Keycloak ni la del correo salen de aquí;
// sólo se dice **si** hay una puesta, que es lo que la pantalla necesita para no pedirla de nuevo.
type EstadoDeInstalacion struct {
	Installed    bool
	Name         string
	Language     string
	EntryMethod  string
	TimeZone     string
	PublicAppURL string

	Directory DirectoryView
	Keycloak  KeycloakView

	// Mail dice si el correo saliente ya está puesto, y con qué remitente.
	MailSet       bool
	MailHost      string
	MailPort      string
	MailSecure    bool
	MailUser      string
	MailFromName  string
	MailFromEmail string

	// AiAvailable lo pone el cableado: si el motor de IA responde. Es opcional, y la instalación
	// funciona entera sin él (docs/modules/ai.md).
	AiAvailable bool
}

// PasoDeInstalacion es lo que trae **un paso** del asistente. Cada paso usa lo suyo y deja lo demás
// como está.
type PasoDeInstalacion struct {
	// Paso 1: la instalación.
	Name     string
	Language string
	// Paso 2: cómo se entra.
	EntryMethod string
	Directory   DirectoryInput
	Keycloak    KeycloakInput
	// Paso 3: dónde está.
	TimeZone     string
	PublicAppURL string
	// Paso 4: el correo saliente.
	Mail MailInput
}

// MailInput es el correo saliente tal y como llega de la pantalla. `Password` vacío quiere decir
// **«no la cambies»**, como en el directorio: la contraseña guardada no sale nunca por la API.
type MailInput struct {
	Host      string
	Port      string
	Secure    bool
	User      string
	Password  string
	FromName  string
	FromEmail string
}

// EstáInstalada dice si el asistente ya se terminó.
func (s *Service) EstáInstalada() (bool, error) {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return false, traducir(err)
	}

	return instalacion.InstalledAt != nil, nil
}

// EstadoDeInstalacion lee lo que hay para que el asistente pueda continuar donde se quedó.
func (s *Service) EstadoDeInstalacion() (EstadoDeInstalacion, error) {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return EstadoDeInstalacion{}, traducir(err)
	}
	directorio, err := s.repo.Directory()
	if err != nil {
		return EstadoDeInstalacion{}, traducir(err)
	}
	keycloak, err := s.repo.Keycloak()
	if err != nil {
		return EstadoDeInstalacion{}, traducir(err)
	}

	estado := EstadoDeInstalacion{
		Installed:     instalacion.InstalledAt != nil,
		Name:          nombreDeLaInstalacion(instalacion.InstallationName),
		Language:      instalacion.Language,
		EntryMethod:   instalacion.EntryMethod,
		TimeZone:      zonaHorariaDe(instalacion.TimeZone),
		PublicAppURL:  direccionPublicaDe(instalacion.PublicAppURL),
		Directory:     vistaDeDirectorio(directorio),
		Keycloak:      vistaDeKeycloak(keycloak),
		MailHost:      instalacion.SMTPHost,
		MailPort:      instalacion.SMTPPort,
		MailSecure:    instalacion.SMTPSecure,
		MailUser:      instalacion.SMTPUser,
		MailFromName:  instalacion.SMTPFromName,
		MailFromEmail: instalacion.SMTPFromEmail,
		MailSet:       strings.TrimSpace(instalacion.SMTPHost) != "",
	}
	if !estado.Installed && !estado.MailSet && strings.TrimSpace(s.correoInicial.Host) != "" {
		estado.MailHost = s.correoInicial.Host
		estado.MailPort = s.correoInicial.Port
		estado.MailSecure = s.correoInicial.Secure
		estado.MailUser = s.correoInicial.User
		estado.MailFromName = s.correoInicial.FromName
		estado.MailFromEmail = s.correoInicial.FromEmail
	}
	if s.ia != nil {
		estado.AiAvailable = s.ia.Disponible()
	}

	return estado, nil
}

// SetProberDeIA fija quién dice si el motor de IA responde. Lo llama el cableado.
func (s *Service) SetProberDeIA(ia ProberDeIA) { s.ia = ia }

// SetProberDeCorreo fija quién prueba el correo saliente. Lo llama el cableado.
func (s *Service) SetProberDeCorreo(correo ProberDeCorreo) { s.correo = correo }

// ProbarPasoDeLaEntrada prueba **el directorio o Keycloak con los datos que llegan**, según el método
// elegido, **sin guardar nada**: es lo que hace el botón «Probar la conexión» del paso 2 del
// asistente. Con el método local no hay nada que probar y no se sale a ninguna parte.
//
// Los datos se validan **igual que al guardar** —mismo método, mismas configuraciones y mismas
// claves—, y después se llama **a la misma prueba que usa Configuración**: el módulo `auth` es el
// único que sabe hablar con un directorio y con un reino, y aquí no se duplica nada.
func (s *Service) ProbarPasoDeLaEntrada(entrada PasoDeInstalacion) error {
	if instalada, err := s.EstáInstalada(); err != nil {
		return err
	} else if instalada {
		return ErrYaInstalada
	}

	metodo := strings.TrimSpace(entrada.EntryMethod)
	if !metodoValido(metodo) {
		return ErrMethodUnknown
	}

	// La contraseña vacía usa la guardada, como en Configuración: el secreto no sale nunca por la API,
	// así que la pantalla no puede mandarlo de vuelta.
	directorio, err := s.directoryDeInput(entrada.Directory, true)
	if err != nil {
		return err
	}
	keycloak, err := s.keycloakDeInput(entrada.Keycloak, true)
	if err != nil {
		return err
	}
	if err := s.metodoEstaConfigurado(metodo, directorio, keycloak); err != nil {
		return err
	}

	// El método local no tiene a qué conectarse: no hay nada que probar.
	if metodo == auth.MethodLocal {
		return nil
	}
	if s.prober == nil {
		return errors.New("no hay quien pruebe la entrada")
	}

	switch metodo {
	case auth.MethodAD:
		if err := s.prober.ProbeDirectory(directorio); err != nil {
			logs.LogWarning("la prueba del directorio del asistente ha fallado: " + err.Error())
			return ErrDirectoryUnreachable
		}
	case auth.MethodKeycloak:
		if err := s.prober.ProbeKeycloak(keycloak); err != nil {
			logs.LogWarning("la prueba de Keycloak del asistente ha fallado: " + err.Error())
			return ErrKeycloakUnreachable
		}
	}

	return nil
}

// ProbarCorreoDeInstalacion prueba **el correo saliente con los datos que llegan**, sin guardarlo y
// **sin mandar ningún correo**: comprueba la conexión y la autenticación —host, puerto, TLS, usuario y
// contraseña—, que es lo que hace falta en el paso 4, donde no hay destinatario.
//
// Los datos se validan **igual que al guardar**: sin servidor, sin puerto o sin remitente con arroba
// se contesta la misma clave que guardaría.
func (s *Service) ProbarCorreoDeInstalacion(entrada MailInput) error {
	if instalada, err := s.EstáInstalada(); err != nil {
		return err
	} else if instalada {
		return ErrYaInstalada
	}

	correo, err := s.correoDeInput(entrada, true)
	if err != nil {
		return err
	}
	if s.correo == nil {
		return errors.New("no hay quien pruebe el correo")
	}

	if err := s.correo.ProbarCorreo(correo); err != nil {
		// El motivo —no contesta, o no acepta las credenciales— va al log, que es donde lo ve quien
		// administra. A la pantalla le llega una clave, nunca un secreto.
		logs.LogWarning("la prueba del correo del asistente ha fallado: " + err.Error())
		return ErrCorreoInalcanzable
	}

	return nil
}

// correoDeInput arma el correo saliente a partir de lo que llega de la pantalla.
//
// Con `usarLaGuardada`, la contraseña vacía se rellena con la que hay en la base: es lo que permite
// probar una configuración sin volver a escribir el secreto. Al guardar, en cambio, vacío quiere
// decir «no la cambies» y se resuelve en el propio paso.
func (s *Service) correoDeInput(input MailInput, usarLaGuardada bool) (CorreoSaliente, error) {
	correo := CorreoSaliente{
		Host:      strings.TrimSpace(input.Host),
		Port:      strings.TrimSpace(input.Port),
		Secure:    input.Secure,
		User:      strings.TrimSpace(input.User),
		Password:  strings.TrimSpace(input.Password),
		FromName:  strings.TrimSpace(input.FromName),
		FromEmail: strings.TrimSpace(input.FromEmail),
	}

	if usarLaGuardada && correo.Password == "" {
		guardada, err := s.repo.Installation()
		if err != nil {
			return CorreoSaliente{}, traducir(err)
		}
		correo.Password = guardada.SMTPPassword
	}

	// Sin servidor, sin puerto y sin remitente con arroba no hay nada que probar: es la misma
	// condición que exige el guardado.
	if correo.Host == "" || correo.Port == "" || !strings.Contains(correo.FromEmail, "@") {
		return CorreoSaliente{}, ErrCorreoIncompleto
	}

	return correo, nil
}

// GuardarPasoDeInstalacion guarda **un paso** del asistente. Se niega si la instalación ya está
// sellada, y cada paso comprueba **sólo lo suyo**: el resto todavía no está puesto.
func (s *Service) GuardarPasoDeInstalacion(paso int, entrada PasoDeInstalacion) (EstadoDeInstalacion, error) {
	if instalada, err := s.EstáInstalada(); err != nil {
		return EstadoDeInstalacion{}, err
	} else if instalada {
		return EstadoDeInstalacion{}, ErrYaInstalada
	}

	var err error
	switch paso {
	case 1:
		err = s.guardarPasoDeLaInstalacion(entrada)
	case 2:
		err = s.guardarPasoDeLaEntrada(entrada)
	case 3:
		err = s.guardarPasoDeDondeEsta(entrada)
	case 4:
		err = s.guardarPasoDelCorreo(entrada)
	default:
		err = ErrPasoIncompleto
	}
	if err != nil {
		return EstadoDeInstalacion{}, err
	}

	return s.EstadoDeInstalacion()
}

// TerminarInstalacion sella la instalación: **es el paso que hace que el asistente no vuelva a
// aparecer**. Comprueba lo imprescindible antes de sellar, para no dejar una instalación sin puerta.
func (s *Service) TerminarInstalacion() (EstadoDeInstalacion, error) {
	if instalada, err := s.EstáInstalada(); err != nil {
		return EstadoDeInstalacion{}, err
	} else if instalada {
		return EstadoDeInstalacion{}, ErrYaInstalada
	}

	instalacion, err := s.repo.Installation()
	if err != nil {
		return EstadoDeInstalacion{}, traducir(err)
	}
	directorio, err := s.repo.Directory()
	if err != nil {
		return EstadoDeInstalacion{}, traducir(err)
	}
	keycloak, err := s.repo.Keycloak()
	if err != nil {
		return EstadoDeInstalacion{}, traducir(err)
	}

	if err := s.metodoEstaConfigurado(instalacion.EntryMethod, directorioDe(directorio), keycloakDe(keycloak)); err != nil {
		return EstadoDeInstalacion{}, err
	}

	if err := s.repo.UpdateInstallation(map[string]any{"installed_at": time.Now()}, nil); err != nil {
		return EstadoDeInstalacion{}, traducir(err)
	}

	return s.EstadoDeInstalacion()
}

// --- los pasos -------------------------------------------------------------------------------

func (s *Service) guardarPasoDeLaInstalacion(entrada PasoDeInstalacion) error {
	nombre := strings.TrimSpace(entrada.Name)
	if nombre == "" {
		nombre = NombreDeFabrica
	}
	if len([]rune(nombre)) > MaxNameLength {
		return ErrNameTooLong
	}
	if !contiene(languages, entrada.Language) {
		return ErrLanguageUnknown
	}

	return traducir(s.repo.UpdateInstallation(map[string]any{
		"installation_name": nombre,
		"language":          entrada.Language,
	}, nil))
}

func (s *Service) guardarPasoDeLaEntrada(entrada PasoDeInstalacion) error {
	metodo := strings.TrimSpace(entrada.EntryMethod)
	if !metodoValido(metodo) {
		return ErrMethodUnknown
	}

	directorio, err := s.directoryDeInput(entrada.Directory, false)
	if err != nil {
		return err
	}
	keycloak, err := s.keycloakDeInput(entrada.Keycloak, false)
	if err != nil {
		return err
	}
	if err := s.metodoEstaConfigurado(metodo, directorio, keycloak); err != nil {
		return err
	}

	if err := s.repo.UpdateInstallation(map[string]any{"entry_method": metodo}, nil); err != nil {
		return traducir(err)
	}
	if err := s.repo.UpdateDirectory(s.cambiosDeDirectorio(directorio), nil); err != nil {
		return traducir(err)
	}

	return traducir(s.repo.UpdateKeycloak(s.cambiosDeKeycloak(keycloak), nil))
}

func (s *Service) guardarPasoDeDondeEsta(entrada PasoDeInstalacion) error {
	zona := strings.TrimSpace(entrada.TimeZone)
	if zona == "" {
		return ErrTimeZoneUnknown
	}
	if _, err := time.LoadLocation(zona); err != nil {
		return ErrTimeZoneUnknown
	}

	direccion := direccionPublicaDe(entrada.PublicAppURL)
	if direccion != "" {
		if err := validarDireccionPublica(direccion); err != nil {
			return err
		}
	}

	return traducir(s.repo.UpdateInstallation(map[string]any{
		"time_zone":      zona,
		"public_app_url": direccion,
	}, nil))
}

func (s *Service) guardarPasoDelCorreo(entrada PasoDeInstalacion) error {
	correo := entrada.Mail
	anfitrion := strings.TrimSpace(correo.Host)
	puesto := strings.TrimSpace(correo.FromEmail)

	if anfitrion == "" {
		// **Se puede dejar el correo sin poner**: hay instalaciones que no mandan correo todavía, y el
		// asistente no puede obligar a tener un servidor a mano. Se dice, y se sigue.
		return nil
	}
	if !strings.Contains(puesto, "@") || strings.TrimSpace(correo.Port) == "" {
		return ErrCorreoIncompleto
	}

	cambios := map[string]any{
		"smtp_host":       anfitrion,
		"smtp_port":       strings.TrimSpace(correo.Port),
		"smtp_secure":     correo.Secure,
		"smtp_user":       strings.TrimSpace(correo.User),
		"smtp_from_name":  strings.TrimSpace(correo.FromName),
		"smtp_from_email": puesto,
	}
	// **Vacío es «no la cambies»**, como en el directorio: la contraseña guardada no sale por la API.
	if contrasena := strings.TrimSpace(correo.Password); contrasena != "" {
		cambios["smtp_password"] = contrasena
	}

	return traducir(s.repo.UpdateInstallation(cambios, nil))
}

// SMTP es el correo saliente ya resuelto, para el módulo de correo: lo de la configuración y, si no
// hay, **el respaldo del entorno**. Declarado aquí y consumido por quien lo necesita
// (docs/primer-arranque.md, sección 5).
func (s *Service) SMTP() (SMTPResuelto, error) {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return SMTPResuelto{}, traducir(err)
	}
	if strings.TrimSpace(instalacion.SMTPHost) != "" {
		return SMTPResuelto{
			Host:      instalacion.SMTPHost,
			Port:      instalacion.SMTPPort,
			Secure:    instalacion.SMTPSecure,
			User:      instalacion.SMTPUser,
			Password:  instalacion.SMTPPassword,
			FromName:  instalacion.SMTPFromName,
			FromEmail: instalacion.SMTPFromEmail,
			Set:       true,
		}, nil
	}

	return SMTPResuelto{}, nil
}

// SMTPResuelto es el correo saliente que se le da al módulo de correo. `Set` en falso quiere decir
// «no hay nada configurado»: el módulo usa entonces lo del entorno.
type SMTPResuelto struct {
	Host      string
	Port      string
	Secure    bool
	User      string
	Password  string
	FromName  string
	FromEmail string
	Set       bool
}

// El sello lo pone el asistente, y **con la identidad de nadie**: no hay sesión todavía. Se marca el
// actor como vacío a propósito, y queda dicho aquí para que no sorprenda: es la única escritura de la
// aplicación que no tiene detrás a una cuenta (docs/primer-arranque.md, sección 6).
var _ = auth.Identity{}
