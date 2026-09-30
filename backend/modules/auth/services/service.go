// Package services contiene las reglas del módulo auth: cómo se entra, cómo se emiten y se gastan
// los enlaces de contraseña y cómo se cambia una contraseña.
package services

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/auth/repositories"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/config"
)

// Las claves de error del módulo (docs/modules/auth.md, sección 10).
var (
	ErrInvalidCredentials = errors.New("auth.invalidCredentials")
	ErrAccountInactive    = errors.New("auth.accountInactive")
	ErrTokenExpired       = errors.New("auth.token.expired")
	ErrTokenUsed          = errors.New("auth.token.used")
)

// Las dos duraciones de un enlace, y el texto con el que se cuentan en el correo.
//
// El texto lo pone este módulo y no `mail`: las plantillas llevan el marcador `{{caducidad}}` y quien
// sabe si son 24 horas o 1 es quien emite el enlace (docs/modules/auth.md, sección 7).
const (
	InvitationDuration = 24 * time.Hour
	RecoveryDuration   = time.Hour
)

var expiryText = map[string]map[time.Duration]string{
	"es": {InvitationDuration: "24 horas", RecoveryDuration: "1 hora"},
	"en": {InvitationDuration: "24 hours", RecoveryDuration: "1 hour"},
}

// Las tres plantillas de correo de este módulo (docs/modules/mail.md, sección 5).
const (
	TemplateInvitation      = "auth.invitation"
	TemplateRecovery        = "auth.recovery"
	TemplatePasswordChanged = "auth.passwordChanged"
)

// Accounts es lo que este módulo necesita del módulo de cuentas.
//
// Se declara aquí, en quien lo usa, y lo cumple el servicio de `users` sin saber que existe: así
// ninguno de los dos módulos importa al otro (docs/arquitectura.md, sección 4).
type Accounts interface {
	ByID(id int64) (auth.Account, error)
	ByEmail(email string) (auth.Account, error)
	SetPasswordHash(id int64, hash string) error
	TouchLastLogin(id int64) error
	// UpsertFromDirectory crea, vincula o actualiza la cuenta de quien entra por el directorio
	// (docs/modules/auth.md, sección 5.4). Vive en `users`, que es quien tiene la tabla.
	UpsertFromDirectory(datos auth.DirectoryAccount) (auth.Account, error)
}

// Mailer es lo que este módulo necesita del módulo de correo: mandar un aviso en segundo plano.
type Mailer interface {
	SendAsync(key, language string, to []string, data map[string]string)
}

// Access es lo que este módulo necesita de la configuración de la instalación: **cómo se entra aquí**.
//
// Se declara aquí, en quien la usa, y la cumple el módulo `settings`, que es quien tiene las tablas:
// el tipo que viaja es de `shared` y ninguno de los dos módulos importa al otro
// (docs/arquitectura.md, sección 4).
// ErrSinDireccionPublica: no hay dirección pública configurada y el enlace de un correo no se puede
// armar. La pantalla lo enseña como un 422 con su clave, no manda un enlace roto.
var ErrSinDireccionPublica = errors.New("auth.publicUrl.missing")

type Access interface {
	Access() (config.Access, error)
}

// Enlaces es lo que el módulo necesita de la configuración para **armar sus enlaces**: la dirección
// pública de la instalación, que es la base de los enlaces de los correos y de la vuelta de
// Keycloak. La declara el propio módulo, así que no depende de quién se la dé
// (docs/modules/settings.md, decisión 14).
type Enlaces interface {
	DireccionPublica() string
}

// Config es lo que el módulo necesita de la configuración de la instalación.
type Config struct {
	// AdminPassword es la contraseña de la cuenta de fábrica. Se compara en cada entrada: la
	// variable de entorno es la fuente de verdad (docs/usuarios-y-permisos.md, sección 8).
	AdminPassword string
	// PublicAppURL es la base de los enlaces que van en los correos, y también la de la vuelta de
	// Keycloak, que deja a la persona en la pantalla de entrada.
	PublicAppURL string
	// TokenSecret firma el `state` de OIDC. Es el mismo secreto que firma el token de sesión: el
	// `state` es un valor que emitimos nosotros y hay que poder reconocer, y no hace falta un segundo
	// secreto para eso (docs/modules/auth.md, sección 5.3).
	TokenSecret string
}

// Service es el módulo.
type Service struct {
	accounts Accounts
	tokens   *repositories.PasswordTokenRepository
	mailer   Mailer
	session  *auth.TokenManager
	cfg      Config

	// access es de dónde se lee cómo se entra en esta instalación. **Se lee en cada intento y no se
	// recuerda**: un administrador puede cambiar el método o el directorio desde Configuración, y eso
	// tiene que valer en el intento siguiente (docs/modules/settings.md, sección 5.8).
	access Access

	// directory y keycloak son los que hablan el protocolo. El módulo construye los suyos con lo que
	// diga la configuración; estas dos puertas existen para poder probar el servicio **sin un
	// directorio ni un reino de verdad**, y por eso sólo las usan las pruebas.
	directory Directorio
	keycloak  Keycloak

	// El cliente de Keycloak se recuerda **mientras su configuración no cambie**: lleva dentro el
	// documento de descubrimiento del reino, y pedirlo en cada entrada sería una ida y vuelta de red
	// de más en el camino de entrar.
	mu      sync.Mutex
	oidc    Keycloak
	oidcCfg config.OIDC

	now func() time.Time
	// enlaces es de dónde salen la dirección pública y los enlaces (ver `SetEnlaces`).
	enlaces Enlaces
}

// NewService construye el servicio.
func NewService(
	accounts Accounts,
	tokens *repositories.PasswordTokenRepository,
	mailer Mailer,
	tokenManager *auth.TokenManager,
	cfg Config,
) *Service {
	return &Service{
		accounts: accounts,
		tokens:   tokens,
		mailer:   mailer,
		session:  tokenManager,
		cfg:      cfg,
		now:      time.Now,
	}
}

// SetAccess conecta la configuración de la instalación, de donde sale cómo se entra. Se llama una vez,
// al arrancar.
func (s *Service) SetAccess(access Access) { s.access = access }

// SetEnlaces fija de dónde salen los enlaces. Sin él, se usa la variable de entorno del arranque,
// que es como funcionaba antes: así una instalación que ya la tuviera puesta no cambia.
func (s *Service) SetEnlaces(enlaces Enlaces) { s.enlaces = enlaces }

// baseDeLosEnlaces es la dirección pública: **la de la configuración primero** y, si no hay, la del
// entorno. Vacío quiere decir «no hay», y quien la use decide qué hacer.
func (s *Service) baseDeLosEnlaces() string {
	if s.enlaces != nil {
		if base := strings.TrimRight(strings.TrimSpace(s.enlaces.DireccionPublica()), "/"); base != "" {
			return base
		}
	}

	return strings.TrimRight(strings.TrimSpace(s.cfg.PublicAppURL), "/")
}

// SetDirectory fija quién habla con el directorio. **Sólo lo usan las pruebas**: en marcha, el módulo
// construye su cliente LDAP con lo que diga la configuración guardada.
func (s *Service) SetDirectory(directory Directorio) { s.directory = directory }

// SetKeycloak fija quién habla con Keycloak, con la misma regla que el directorio.
func (s *Service) SetKeycloak(keycloak Keycloak) { s.keycloak = keycloak }

// acceso lee cómo se entra en esta instalación.
//
// Sin configuración conectada sólo existe el camino local: es lo que hace que una instalación recién
// puesta —que todavía no tiene fila de configuración— no se quede sin puerta. La cuenta de fábrica
// entra siempre, por su cuenta (docs/usuarios-y-permisos.md, sección 8).
func (s *Service) acceso() (config.Access, error) {
	if s.access == nil {
		return config.Access{Method: auth.MethodLocal}, nil
	}

	return s.access.Access()
}

// directorioDe devuelve el que habla con el directorio para esa configuración.
func (s *Service) directorioDe(cfg config.Directory) Directorio {
	if s.directory != nil {
		return s.directory
	}

	return NewLDAP(cfg)
}

// keycloakDe devuelve el que habla con Keycloak, recordado mientras la configuración no cambie: si un
// administrador cambia el reino o el cliente, se construye otro y se vuelve a leer su descubrimiento.
func (s *Service) keycloakDe(cfg config.OIDC) Keycloak {
	if s.keycloak != nil {
		return s.keycloak
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.oidc != nil && s.oidcCfg == cfg {
		return s.oidc
	}

	s.oidc = NewOIDC(cfg)
	s.oidcCfg = cfg

	return s.oidc
}

// Methods dice cómo se entra en esta instalación.
//
// **Es público y no lleva ningún dato de nadie**: dice el método que está puesto y cuál de los tres
// caminos es el que se ofrece, que es lo que la pantalla de entrada necesita para pintarse
// (docs/modules/auth.md, decisión 27).
//
// Los tres caminos se dicen uno a uno, y no sólo el nombre del método, porque la pantalla no tiene por
// qué conocer los nombres: con el método en AD sigue habiendo **formulario de correo y contraseña** —es
// el mismo, lo que cambia es quién lo contesta—, y con el método en Keycloak no lo hay.
func (s *Service) Methods() (Methods, error) {
	acceso, err := s.acceso()
	if err != nil {
		return Methods{}, err
	}

	return Methods{
		Method:   acceso.Method,
		Local:    acceso.Method == auth.MethodLocal,
		AD:       acceso.Method == auth.MethodAD,
		Keycloak: acceso.Method == auth.MethodKeycloak,
	}, nil
}

// Methods es lo que responde el endpoint de los caminos de entrada.
type Methods struct {
	// Method es el método que está puesto: `local`, `ad` o `keycloak`.
	Method   string `json:"method"`
	Local    bool   `json:"local"`
	AD       bool   `json:"ad"`
	Keycloak bool   `json:"keycloak"`
}

// Session es el resultado de una entrada: el token, cuándo caduca y quién es.
type Session struct {
	Token     string
	ExpiresAt time.Time
	Account   auth.Account
	Factory   bool
}

// Login entra con correo y contraseña (docs/modules/auth.md, sección 5.1).
//
// **De qué camino se entra lo decide la instalación, no quien escribe**: el método que está puesto es
// el único que se atiende, y los otros dos quedan apagados aunque estén configurados
// (docs/modules/settings.md, sección 5.8). La cuenta de fábrica es la excepción, y va primero.
func (s *Service) Login(email, password string) (Session, error) {
	email = strings.TrimSpace(email)

	// La cuenta de fábrica no está en la base: su correo es `admin` y su contraseña vive en la
	// configuración. **Entra siempre**, sea cual sea el método, porque es la única puerta que no se
	// puede cerrar: sin ella, elegir mal el método dejaría la instalación sin nadie que lo cambie
	// (docs/usuarios-y-permisos.md, sección 8).
	if strings.EqualFold(email, auth.FactorySubject) {
		if !auth.SameSecret(password, s.cfg.AdminPassword) {
			return Session{}, ErrInvalidCredentials
		}
		return s.factorySession()
	}

	acceso, err := s.acceso()
	if err != nil {
		return Session{}, err
	}

	switch acceso.Method {
	case auth.MethodAD:
		// Con el método en AD se entra **por el directorio y sólo por el directorio**: las
		// contraseñas locales no valen, ni siquiera para Soporte o Desarrollo. Es lo que quiere decir
		// «un método a la vez», y por eso la cuenta de fábrica tiene su puerta aparte.
		return s.entrarPorElDirectorio(email, password, acceso.Directory)

	case auth.MethodKeycloak:
		// El camino de Keycloak es una vuelta por el navegador y no este endpoint: aquí no hay
		// contraseña que comparar (docs/modules/auth.md, sección 5.3).
		return Session{}, ErrInvalidCredentials

	default:
		return s.entrarLocal(email, password)
	}
}

// entrarLocal es la entrada de fábrica: correo y contraseña contra la tabla de cuentas.
func (s *Service) entrarLocal(email, password string) (Session, error) {
	account, err := s.accounts.ByEmail(email)
	if errors.Is(err, auth.ErrAccountNotFound) {
		// Sin cuenta, y con el método en local, no hay a quién más preguntar: decir «esa cuenta no
		// existe» es regalar la lista de quién trabaja aquí.
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, err
	}

	// Una cuenta de directorio **no tiene contraseña aquí**: con el método en local no hay camino para
	// ella, y quien la use tiene que esperar a que la instalación vuelva a su método.
	if account.IsDirectory() {
		return Session{}, ErrInvalidCredentials
	}

	if !account.HasPassword() || !PasswordMatches(account.PasswordHash, password) {
		return Session{}, ErrInvalidCredentials
	}

	// Desactivada, se dice: esa persona no puede arreglarlo sola y necesita llamar a Soporte. Es
	// información que se le da a quien ya ha demostrado saber la contraseña.
	if !account.IsActive {
		return Session{}, ErrAccountInactive
	}

	return s.sesionDeCuenta(account)
}

// entrarPorElDirectorio es la entrada por el directorio de la organización (docs/modules/auth.md,
// sección 5.2).
//
// Aquí no se mira si la cuenta existe: **el directorio es quien sabe quién trabaja en la casa**. Si
// conoce a esa persona y acepta sus credenciales, la cuenta se crea, se vincula o se pone al día sola
// (docs/modules/auth.md, sección 5.4).
func (s *Service) entrarPorElDirectorio(email, password string, cfg config.Directory) (Session, error) {
	// Lo que se sabe de esa persona **antes de preguntar**: si tiene cuenta aquí y de dónde es. Es lo
	// que decide qué se cuenta cuando el directorio no acepta, y lo que impide que una cuenta de un
	// camino entre por el otro.
	cuenta, errCuenta := s.accounts.ByEmail(email)
	esDeAlli := errCuenta == nil && cuenta.IsDirectory()

	if errCuenta == nil && cuenta.Origin == auth.OriginKeycloak {
		// **Una cuenta es de un camino y no de dos**: la de Keycloak entra por su vuelta, no por el
		// directorio, aunque el método puesto sea el de AD.
		return Session{}, ErrInvalidCredentials
	}

	datos, err := s.directorioDe(cfg).Login(email, password)
	if err != nil {
		// **Se dice lo que ha pasado de verdad sólo a quien se sabe que es de allí**, y eso se sabe si
		// hay una cuenta de directorio con ese correo. A cualquier otro se le contesta lo mismo que a
		// una contraseña equivocada: decir «esa cuenta no está» o «el directorio no responde» sería
		// contarle a un desconocido quién trabaja aquí y cómo está la red de la casa
		// (docs/modules/auth.md, sección 5.2).
		if errors.Is(err, ErrDirectoryUnavailable) {
			if esDeAlli {
				return Session{}, ErrDirectoryUnavailable
			}

			logs.LogWarning("el directorio no responde al intentar entrar con ese correo")
			return Session{}, ErrInvalidCredentials
		}

		if esDeAlli {
			return Session{}, ErrDirectoryRejected
		}

		return Session{}, ErrInvalidCredentials
	}

	return s.sesionDeDirectorio(datos)
}

// Knows dice si el directorio sigue conociendo a alguien con ese correo.
//
// Lo necesita `users` para **reactivar una cuenta de AD**, y se pregunta igual que se entra: la misma
// cuenta de servicio y el mismo filtro. Una instalación sin directorio no puede contestar, y eso no es
// un error de nadie: es que no hay a quién preguntar (docs/modules/users.md, sección 5, punto 4).
func (s *Service) Knows(email string) (bool, error) {
	acceso, err := s.acceso()
	if err != nil {
		return false, err
	}

	if !acceso.Directory.Configured() {
		return false, ErrDirectoryUnavailable
	}

	return s.directorioDe(acceso.Directory).Knows(email)
}

// sesionDeCuenta firma la sesión de una cuenta local y apunta la entrada.
func (s *Service) sesionDeCuenta(account auth.Account) (Session, error) {
	now := s.now()
	token, err := s.session.Sign(subjectOf(account), now)
	if err != nil {
		return Session{}, err
	}

	// Si esto falla, la entrada ya ha ocurrido: se registra y se sigue, porque negar la entrada por
	// no poder apuntar la fecha sería absurdo.
	if err := s.accounts.TouchLastLogin(account.ID); err != nil {
		logs.LogError("no se pudo apuntar la última entrada de la cuenta: " + err.Error())
	}

	return Session{
		Token:     token,
		ExpiresAt: now.Add(auth.SessionDuration),
		Account:   account,
	}, nil
}

// sesionDeDirectorio deja la cuenta al día con lo que dice el directorio y firma la sesión.
func (s *Service) sesionDeDirectorio(datos auth.DirectoryAccount) (Session, error) {
	account, err := s.accounts.UpsertFromDirectory(datos)
	if err != nil {
		return Session{}, err
	}

	now := s.now()
	token, err := s.session.Sign(subjectOf(account), now)
	if err != nil {
		return Session{}, err
	}

	if err := s.accounts.TouchLastLogin(account.ID); err != nil {
		logs.LogError("no se pudo apuntar la última entrada de la cuenta: " + err.Error())
	}

	// La cuenta se nombra **por su identificador**, no por su correo: fuera del envío de un correo,
	// que es donde la dirección es el dato que se investiga, no se registran datos personales
	// (AGENTS.md, reglas para agentes).
	logs.LogSuccess("entrada por el directorio (" + account.Origin + ") de la cuenta #" + strconv.FormatInt(account.ID, 10))

	return Session{
		Token:     token,
		ExpiresAt: now.Add(auth.SessionDuration),
		Account:   account,
	}, nil
}

// keycloakDelMetodo devuelve con quién hablar de Keycloak, **si el método puesto es el suyo**.
//
// Con el método en local o en AD ese camino no existe: contestar que no está configurado es lo que ya
// sabe leer la pantalla de entrada, y evita una segunda clave para lo mismo.
func (s *Service) keycloakDelMetodo() (Keycloak, error) {
	acceso, err := s.acceso()
	if err != nil {
		return nil, err
	}

	if acceso.Method != auth.MethodKeycloak {
		return nil, ErrOIDCNotConfigured
	}

	keycloak := s.keycloakDe(acceso.OIDC)
	if !keycloak.Configured() {
		return nil, ErrOIDCNotConfigured
	}

	return keycloak, nil
}

// StartKeycloak devuelve la dirección a la que hay que mandar el navegador para entrar por Keycloak.
//
// **El `state` se acuña aquí y viaja firmado**: es lo que permite comprobar en la vuelta que la
// entrada la empezó esta instalación y que no ha pasado demasiado tiempo, sin cookies y sin guardar
// nada (docs/modules/auth.md, sección 5.3).
func (s *Service) StartKeycloak() (string, error) {
	keycloak, err := s.keycloakDelMetodo()
	if err != nil {
		return "", err
	}

	state, err := auth.SignState(s.cfg.TokenSecret, s.now())
	if err != nil {
		return "", err
	}

	return keycloak.AuthURL(state)
}

// KeycloakCallback cierra la vuelta de Keycloak y devuelve **la dirección a la que mandar al
// navegador**, que es siempre la pantalla de entrada: con el token en el fragmento si se ha entrado, y
// con la clave del fallo si no.
//
// Devuelve una dirección y no un error a propósito: esto lo recibe un navegador en mitad de un
// recorrido, así que hasta cuando sale mal hay que dejar a la persona en un sitio donde se le pueda
// contar qué ha pasado (docs/modules/auth.md, decisión 30).
func (s *Service) KeycloakCallback(state, code string) string {
	vuelta := s.direccionDeVuelta()

	keycloak, err := s.keycloakDelMetodo()
	if err != nil {
		return vuelta + "#error=" + claveDeOIDC(err)
	}

	if err := auth.VerifyState(s.cfg.TokenSecret, state, s.now()); err != nil {
		// Ni el `state` ni el código se registran: sólo que la vuelta no valía. Un `state` que no vale
		// es o alguien que se entretuvo más de cinco minutos, o alguien probando a fabricarse una
		// vuelta; el log lo distingue por el motivo.
		logs.LogWarning("vuelta de Keycloak con un state que no vale: " + err.Error())
		return vuelta + "#error=auth.oidc.state"
	}

	datos, err := keycloak.Exchange(code)
	if err != nil {
		logs.LogWarning("la vuelta de Keycloak no se pudo completar: " + err.Error())
		return vuelta + "#error=" + claveDeOIDC(err)
	}

	sesion, err := s.sesionDeDirectorio(datos)
	if err != nil {
		// La cuenta no se pudo dejar al día. El detalle va al log; a la pantalla, la clave si es de
		// las que ya sabe leer.
		logs.LogWarning("la cuenta que vuelve de Keycloak no se pudo dejar al día: " + err.Error())
		return vuelta + "#error=" + claveDeCuenta(err)
	}

	return vuelta + "#token=" + sesion.Token
}

// direccionDeVuelta es dónde se deja a quien vuelve de Keycloak: **la pantalla de entrada**, que es
// donde empezó todo, con el token o con el fallo en el fragmento (decisión 30).
//
// El fragmento no viaja al servidor: no queda en los registros de nginx ni en el historial de nadie
// como parte de una petición, y el frontend lo borra en cuanto lo usa (decisión 13).
// zonaHoraria es la de la instalación, leída de la configuración en cada uso: cambiarla en
// Configuración vale sin reiniciar nada. Si no se pudiera leer, **UTC**, que es lo que se hacía
// antes de que existiera (docs/modules/settings.md, decisión 15).
func (s *Service) zonaHoraria() *time.Location {
	if s.access != nil {
		if acceso, err := s.access.Access(); err == nil {
			if zona, err := time.LoadLocation(acceso.TimeZone); err == nil {
				return zona
			}
		}
	}

	return time.UTC
}
func (s *Service) direccionDeVuelta() string {
	base := s.baseDeLosEnlaces()
	if base == "" {
		// Sin dirección pública configurada se devuelve una ruta relativa, que el navegador resuelve
		// contra el mismo sitio del que vino: es mejor que mandar a nadie a un host inventado.
		return "/login"
	}

	return base + "/login"
}

// ProbeDirectory prueba una configuración de directorio **sin guardarla**: es lo que hace el botón
// «Probar la conexión» de Configuración, y lo que pide el módulo `settings` por su interfaz.
//
// Abre la conexión con la cuenta de servicio, que es lo mínimo que tiene que funcionar para que
// alguien pueda entrar. **No valida la contraseña de nadie**: eso sólo se sabe cuando alguien entra.
func (s *Service) ProbeDirectory(cfg config.Directory) error {
	return NewLDAP(cfg).Probe()
}

// ProbeKeycloak prueba una configuración de Keycloak sin guardarla: lee el documento del reino y
// comprueba que dice dónde está su pantalla de entrada, que es por donde se empieza a entrar.
func (s *Service) ProbeKeycloak(cfg config.OIDC) error {
	return NewOIDC(cfg).Probe()
}

// claveDeOIDC traduce el fallo del camino de OIDC a la clave que lee la pantalla de entrada
// (docs/modules/auth.md, sección 10).
func claveDeOIDC(err error) string {
	switch {
	case errors.Is(err, ErrOIDCNotConfigured):
		return "auth.oidc.notConfigured"
	case errors.Is(err, ErrOIDCNoEmail):
		return "auth.oidc.noEmail"
	case errors.Is(err, ErrOIDCUnavailable):
		return "auth.oidc.unavailable"
	default:
		return "auth.oidc.rejected"
	}
}

// clavesDeCuenta son las claves del módulo de cuentas que la pantalla de entrada ya sabe leer.
//
// Se comparan **por su texto** porque `auth` no puede importar `users` (regla dura de modularidad):
// la clave es el contrato entre los dos, y ese contrato es público. Cualquier otra cosa que devuelva
// el módulo de cuentas es un fallo de dentro, y a la pantalla le llega el mensaje genérico mientras
// el detalle queda en el log.
var clavesDeCuenta = map[string]bool{
	"users.email.invalid":    true,
	"users.email.duplicated": true,
	"users.origin.unknown":   true,
	"users.accountInactive":  true,
}

// claveDeCuenta devuelve la clave del fallo de la cuenta, si es de las que se pueden contar.
func claveDeCuenta(err error) string {
	if clavesDeCuenta[err.Error()] {
		return err.Error()
	}

	return "auth.oidc.rejected"
}

// factorySession firma la sesión de la cuenta de fábrica.
func (s *Service) factorySession() (Session, error) {
	now := s.now()

	token, err := s.session.Sign(auth.FactorySubject, now)
	if err != nil {
		return Session{}, err
	}

	return Session{
		Token:     token,
		ExpiresAt: now.Add(auth.SessionDuration),
		Factory:   true,
		Account: auth.Account{
			ID:       0,
			Name:     "Administrador",
			Role:     auth.RoleAdministrador,
			Origin:   auth.OriginLocal,
			Language: "es",
			IsActive: true,
		},
	}, nil
}

// Logout no invalida nada, porque no hay tabla de sesiones que tocar: deja constancia en el log, que
// es la forma de saber cuándo alguien dijo que se iba (docs/modules/auth.md, sección 8).
func (s *Service) Logout(identity auth.Identity) {
	logs.LogInfo("cierre de sesión de la cuenta " + labelOf(identity))
}

// Invite emite el enlace de alta (24 horas) y manda el correo. Lo llama el módulo de cuentas.
func (s *Service) Invite(account auth.Account, requestedBy *int64) error {
	return s.sendLink(account, repositories.PurposeInvitation, TemplateInvitation, InvitationDuration, requestedBy)
}

// Recovery emite el enlace de recuperación (1 hora) y manda el correo.
func (s *Service) Recovery(account auth.Account, requestedBy *int64) error {
	return s.sendLink(account, repositories.PurposeRecovery, TemplateRecovery, RecoveryDuration, requestedBy)
}

// ForgotPassword manda el enlace de recuperación si la cuenta existe, y responde igual si no.
//
// Que exista o no es información que no se da: si el endpoint dijera «esa cuenta no existe», serviría
// para averiguar quién trabaja aquí (docs/modules/auth.md, sección 8).
func (s *Service) ForgotPassword(email string) error {
	account, err := s.accounts.ByEmail(strings.TrimSpace(email))
	if errors.Is(err, auth.ErrAccountNotFound) {
		// Se hace como si se hubiera mandado. Y se tarda lo mismo, que es lo que impide deducirlo
		// cronometrando la respuesta.
		return nil
	}
	if err != nil {
		return err
	}

	// A una cuenta de directorio no se le manda nada: su contraseña no está aquí.
	if account.IsDirectory() || !account.IsActive {
		return nil
	}

	if err := s.Recovery(account, nil); err != nil {
		// El fallo se registra, pero la respuesta sigue siendo la misma: quien pregunta no tiene por
		// qué saber si el correo salió.
		logs.LogError("no se pudo mandar el enlace de recuperación: " + err.Error())
	}

	return nil
}

// ResetPassword establece la contraseña con el token de un enlace, sin saber la anterior.
func (s *Service) ResetPassword(token, password, ip string) error {
	// **La contraseña se comprueba antes de gastar el enlace**, y el orden importa: un enlace de un
	// solo uso que se gasta cuando la contraseña no vale deja a la persona sin poder arreglarlo —pide
	// otro y vuelve a empezar—, y el error que acaba de leer ya le decía qué tenía que cambiar
	// (docs/modules/auth.md, sección 4).
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	account, err := s.consumeToken(token, nil)
	if err != nil {
		return err
	}

	if err := s.accounts.SetPasswordHash(account.ID, hash); err != nil {
		return err
	}

	// Y se avisa: es la única forma de que alguien se entere de que han entrado en su cuenta.
	s.notifyPasswordChanged(account, ip)

	return nil
}

// ChangePassword cambia la contraseña propia, y exige la actual.
func (s *Service) ChangePassword(identity auth.Identity, current, password, ip string) error {
	if identity.Factory {
		// La contraseña de la cuenta de fábrica no se cambia desde aquí: vive en la configuración de
		// la instalación y se cambia allí (docs/usuarios-y-permisos.md, sección 8).
		return ErrPasswordNotLocal
	}

	account, err := s.accounts.ByID(identity.ID)
	if err != nil {
		return err
	}

	if account.IsDirectory() {
		return ErrPasswordNotLocal
	}

	if !PasswordMatches(account.PasswordHash, current) {
		return ErrPasswordWrong
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	if err := s.accounts.SetPasswordHash(account.ID, hash); err != nil {
		return err
	}

	s.notifyPasswordChanged(account, ip)

	return nil
}

// sendLink emite el enlace, lo guarda y manda el correo.
//
// El envío va **en segundo plano**: un fallo de correo no puede tumbar el alta ni el reseteo, porque
// la cuenta ya está creada o el enlace ya está emitido, y el correo se puede volver a lanzar
// (docs/modules/users.md, sección 5).
func (s *Service) sendLink(account auth.Account, purpose, template string, duration time.Duration, requestedBy *int64) error {
	token, err := newToken()
	if err != nil {
		return err
	}

	now := s.now()
	registro := &repositories.PasswordToken{
		UserID:    account.ID,
		Purpose:   purpose,
		TokenHash: hashToken(token),
		ExpiresAt: now.Add(duration),
		CreatedBy: requestedBy,
	}

	// Guardar el enlace antes de mandarlo: si el correo sale y el enlace no está guardado, llega un
	// enlace que no vale para nada.
	if err := s.tokens.Create(registro); err != nil {
		return err
	}

	// **El enlace se arma antes de mandar nada** (decisión 14): sin dirección pública no hay enlace, y
	// es mejor fallar aquí —con su clave, que la pantalla traduce— que mandar un correo con un enlace
	// roto que nadie va a poder usar.
	enlace, err := s.link(token)
	if err != nil {
		return err
	}

	s.mailer.SendAsync(template, account.Language, []string{account.Email}, map[string]string{
		"nombre": account.Name,
		"enlace": enlace,
		// El texto de la caducidad, en el idioma de la cuenta, que es el mismo de la plantilla.
		"caducidad": expiryTextFor(account.Language, duration),
	})

	return nil
}

// consumeToken comprueba el token de un enlace y lo gasta.
//
// Devuelve la cuenta dueña del enlace. Los tres motivos por los que un enlace no vale —no existe, ya
// se usó, ha caducado— se distinguen, porque la pantalla puede decir «pide otro» en los tres casos
// pero el log agradece saber cuál fue.
func (s *Service) consumeToken(token string, expectPurpose *string) (auth.Account, error) {
	token = strings.TrimSpace(token)
	if !TokenIsWellFormed(token) {
		return auth.Account{}, ErrTokenUsed
	}

	encontrado, err := s.tokens.FindByHash(hashToken(token))
	if errors.Is(err, repositories.ErrTokenNotFound) {
		// O nunca existió, o lo sustituyó uno más nuevo: los anteriores se borran al emitir otro.
		return auth.Account{}, ErrTokenUsed
	}
	if err != nil {
		return auth.Account{}, err
	}

	if encontrado.Used() {
		return auth.Account{}, ErrTokenUsed
	}

	if expectPurpose != nil && encontrado.Purpose != *expectPurpose {
		return auth.Account{}, ErrTokenUsed
	}

	now := s.now()
	if encontrado.Expired(now) {
		return auth.Account{}, ErrTokenExpired
	}

	account, err := s.accounts.ByID(encontrado.UserID)
	if err != nil {
		return auth.Account{}, err
	}

	// Se gasta antes de tocar la contraseña, y sólo si seguía sin usar: dos peticiones a la vez con el
	// mismo enlace no pueden pasar las dos.
	gastado, err := s.tokens.MarkUsed(encontrado.ID, now)
	if err != nil {
		return auth.Account{}, err
	}
	if !gastado {
		return auth.Account{}, ErrTokenUsed
	}

	return account, nil
}

// notifyPasswordChanged manda el aviso de cambio, en segundo plano.
func (s *Service) notifyPasswordChanged(account auth.Account, ip string) {
	if strings.TrimSpace(ip) == "" {
		ip = "desconocida"
	}

	s.mailer.SendAsync(TemplatePasswordChanged, account.Language, []string{account.Email}, map[string]string{
		"nombre": account.Name,
		// **La fecha del aviso se escribe en la zona de la instalación** (decisión 15): es lo que lee una
		// persona, y en UTC le saldría a una hora que no es la suya.
		"cuando": s.now().In(s.zonaHoraria()).Format("2006-01-02 15:04"),
		"ip":     ip,
	})
}

// link construye el enlace a la pantalla de establecer contraseña.
//
// El token viaja en el **fragmento**, que el navegador no manda al servidor: así no acaba en el
// registro de accesos de nginx ni en ningún log (docs/modules/auth.md, decisión 22).
func (s *Service) link(token string) (string, error) {
	// **Sin dirección pública no hay enlace que mandar**: se devuelve el fallo con su clave en vez de
	// un enlace roto, que es lo que el responsable pidió que no pasara (decisión 14).
	base := s.baseDeLosEnlaces()
	if base == "" {
		return "", ErrSinDireccionPublica
	}

	return base + "/set-password#token=" + token, nil
}

// expiryTextFor devuelve el texto de la caducidad en el idioma de la cuenta.
func expiryTextFor(language string, duration time.Duration) string {
	if textos, ok := expiryText[language]; ok {
		if texto, ok := textos[duration]; ok {
			return texto
		}
	}
	return expiryText["es"][duration]
}

// subjectOf es lo que va en el `sub` del token: el identificador de la cuenta.
func subjectOf(account auth.Account) string {
	return strconv.FormatInt(account.ID, 10)
}

// labelOf describe una cuenta para el log **sin datos personales**: ni el nombre ni el correo, sólo
// el identificador (AGENTS.md, reglas para agentes).
func labelOf(identity auth.Identity) string {
	if identity.Factory {
		return "de la cuenta de fábrica"
	}
	return "de la cuenta #" + strconv.FormatInt(identity.ID, 10)
}
