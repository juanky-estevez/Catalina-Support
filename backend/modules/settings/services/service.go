// Package services contiene las reglas del módulo settings: qué es configurable, con qué límites, y
// la marca de la instalación (el logo y el color institucional).
package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/settings/repositories"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/config"
	"catalina-support/backend/shared/version"
)

// Las variantes del logo y los temas que las usan (docs/modules/settings.md, sección 5.1).
const (
	VarianteClaro  = "claro"
	VarianteOscuro = "oscuro"
)

// Las claves de error del módulo (docs/modules/settings.md, sección 7).
var (
	ErrLanguageUnknown     = errors.New("settings.language.unknown")
	ErrPrimaryColorInvalid = errors.New("settings.primaryColor.invalid")
	// ErrTimeZoneUnknown: la zona horaria que se ha escrito no es un nombre que exista. Se guarda el
	// nombre IANA, no un desfase, y se comprueba con la base de datos de zonas del sistema
	// (docs/modules/settings.md, decisión 15).
	ErrTimeZoneUnknown = errors.New("settings.timeZone.unknown")
	// ErrPublicURLInvalid: la dirección pública no es una dirección http o https con su host.
	ErrPublicURLInvalid = errors.New("settings.publicUrl.invalid")
	// ErrAIURLInvalid: la dirección del motor de IA no es una dirección http o https con su host.
	// Vacía sí vale: es «esta instalación no tiene motor» (docs/modules/ai.md).
	ErrAIURLInvalid = errors.New("settings.aiUrl.invalid")
	// ErrAIUnreachable: el motor de IA no ha contestado a la prueba de la conexión, o ha contestado
	// mal. El detalle queda en el log; a la pantalla le llega que esa dirección no sirve.
	ErrAIUnreachable                = errors.New("settings.ai.unreachable")
	ErrNumberPrefixInvalid          = errors.New("settings.numberPrefix.invalid")
	ErrAssignmentUnknown            = errors.New("settings.assignment.unknown")
	ErrNotificationUnknown          = errors.New("settings.notification.unknown")
	ErrNotificationWithoutAssigment = errors.New("settings.notification.withoutAssignment")
	ErrLogoVariantUnknown           = errors.New("settings.logo.variantUnknown")
	// ErrNameTooLong: el nombre de la instalación no cabe donde se enseña, que es el menú lateral y
	// la tarjeta de la entrada (docs/modules/settings.md).
	ErrNameTooLong = errors.New("settings.name.tooLong")
	// ErrMethodUnknown: el método de entrada no es `local`, `ad` ni `keycloak`.
	ErrMethodUnknown = errors.New("settings.method.unknown")
	// ErrMethodNotConfigured: se elige un método que no está configurado, y eso dejaría la
	// instalación sin puerta para todo el mundo menos la cuenta de fábrica.
	ErrMethodNotConfigured = errors.New("settings.method.notConfigured")
	// ErrDirectoryIncomplete: falta lo mínimo del directorio —el servidor o la base de búsqueda—.
	ErrDirectoryIncomplete = errors.New("settings.directory.incomplete")
	// ErrKeycloakIncomplete: falta lo mínimo de Keycloak —el emisor, el cliente o la vuelta—.
	ErrKeycloakIncomplete = errors.New("settings.keycloak.incomplete")
	// ErrDirectoryUnreachable: la prueba de la conexión con el directorio no ha funcionado —no
	// contesta, o la cuenta de servicio no vale—. El detalle queda en el log; a la pantalla le llega
	// que esa configuración, tal y como está escrita, no sirve.
	ErrDirectoryUnreachable = errors.New("settings.directory.unreachable")
	// ErrKeycloakUnreachable: lo mismo con el reino: no contesta o no dice dónde está su pantalla de
	// entrada.
	ErrKeycloakUnreachable = errors.New("settings.keycloak.unreachable")
	ErrLogoNotFound        = errors.New("settings.logo.notFound")
	ErrSettingsNotFound    = errors.New("settings.notFound")
)

// Los valores aceptados, en la lista cerrada del documento.
var (
	languages     = []string{"es", "en"}
	assignments   = []string{"ninguna", "por_turnos"}
	notifications = []string{"a_nadie", "a_todo_el_equipo", "al_asignado"}
)

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var prefixPattern = regexp.MustCompile(`^[A-Z0-9]{2,8}$`)

// La carpeta de la marca dentro de la carpeta de archivos, que ya existe para los adjuntos y entra
// en la copia de seguridad (docs/ambientes.md).
const carpetaMarca = "brand"

// Los tres métodos de entrada viven en `shared/auth`, que es donde los lee también el camino de
// entrada: **son un valor cerrado que viaja**, no un texto de este módulo. Aquí sólo se validan.

// NombreDeFabrica es el nombre que trae la aplicación, y al que se vuelve si se deja el campo vacío:
// así una instalación nunca se queda sin nombre, que es lo que la base tampoco admite.
const NombreDeFabrica = "Catalina Support"

// MaxNameLength es lo que puede medir el nombre. Se cuenta en **caracteres y no en bytes**, que es lo
// que ve quien lo escribe: «Ayuntamiento de Ávila» son 20, con acentos y todo.
const MaxNameLength = 60

// Repositorio es lo que este módulo necesita de la base de datos. Lo declara aquí quien lo usa y lo
// cumple `repositories.SettingsRepository`: así el módulo se puede probar con un doble, sin base.
type Repositorio interface {
	Installation() (repositories.InstallationSettings, error)
	Tickets() (repositories.TicketSettings, error)
	Directory() (repositories.DirectorySettings, error)
	Keycloak() (repositories.KeycloakSettings, error)
	UpdateInstallation(cambios map[string]any, updatedByID *int64) error
	UpdateDirectory(cambios map[string]any, updatedByID *int64) error
	UpdateKeycloak(cambios map[string]any, updatedByID *int64) error
	UpdateTickets(cambios map[string]any, updatedByID *int64) error
}

// Service es el módulo.
type Service struct {
	repo      Repositorio
	filesPath string
	prober    Prober
	now       func() time.Time
	// ia dice si el motor de IA responde, para poder contarlo en la instalación.
	ia ProberDeIA
	// correo prueba el correo saliente **sin mandar ningún correo**: es lo que hace falta en el
	// asistente, donde no hay destinatario (docs/primer-arranque.md, sección 3).
	correo ProberDeCorreo
}

// NewService construye el servicio.
func NewService(repo Repositorio, filesPath string) *Service {
	return &Service{repo: repo, filesPath: filesPath, now: time.Now}
}

// UpdateInput es lo que llega para guardar la configuración. Va todo: es un `PUT`, y quien guarda
// manda la configuración entera, no un trozo.
type UpdateInput struct {
	Name string
	// EntryMethod es el método de entrada que se elige: `local`, `ad` o `keycloak`.
	EntryMethod string
	// Directory y Keycloak son las dos configuraciones de los caminos de directorio. Van **siempre**,
	// esté puesto el método que esté: así se puede dejar configurado lo que no se usa todavía.
	Directory            DirectoryInput
	Keycloak             KeycloakInput
	Language             string
	PrimaryColor         string
	NumberPrefix         string
	MainAssignment       string
	MainNotification     string
	InternalAssignment   string
	InternalNotification string
	// TimeZone es la zona horaria que se elige (nombre IANA) y PublicAppURL la dirección pública.
	TimeZone     string
	PublicAppURL string
	// AIURL y AIModel son **el motor de IA**: su dirección y el modelo con el que redacta. Vacíos
	// quieren decir «esta instalación no tiene motor», que es lo normal y no es un error
	// (docs/modules/ai.md, decisión 2).
	AIURL   string
	AIModel string
}

// DirectoryInput es la configuración del directorio tal y como llega de la pantalla.
//
// `BindPassword` vacío quiere decir **«no la cambies»**: la contraseña guardada no se devuelve nunca,
// así que la pantalla no puede mandarla de vuelta, y vacío es lo único que puede mandar cuando no la
// toca.
type DirectoryInput struct {
	Host         string
	Port         string
	UseTLS       bool
	BindDN       string
	BindPassword string
	SearchBase   string
	UserFilter   string
	AttrEmail    string
	AttrName     string
	AttrLastName string
	AttrID       string
}

// KeycloakInput es la configuración de Keycloak tal y como llega de la pantalla. El secreto vacío
// quiere decir «no lo cambies», igual que la contraseña del directorio.
type KeycloakInput struct {
	InternalIssuer string
	Issuer         string
	ClientID       string
	ClientSecret   string
	RedirectURI    string
}

// DirectoryView es la configuración del directorio **sin su contraseña**, que es lo único que sale
// por la API: en su lugar se dice si hay una puesta.
type DirectoryView struct {
	DirectoryInput
	// PasswordSet dice si hay contraseña guardada. La contraseña, nunca.
	PasswordSet bool
}

// KeycloakView es la configuración de Keycloak **sin su secreto**.
type KeycloakView struct {
	KeycloakInput
	// SecretSet dice si hay secreto guardado. El secreto, nunca.
	SecretSet bool
}

// Config es la configuración tal y como la ve un Administrador.
type Config struct {
	// Name es el nombre de la instalación, ya resuelto: el que hay guardado, o el de fábrica.
	Name string
	// EntryMethod es el método de entrada que está puesto.
	EntryMethod string
	// Directory y Keycloak son las dos configuraciones, sin sus secretos.
	Directory            DirectoryView
	Keycloak             KeycloakView
	Language             string
	PrimaryColor         string
	NumberPrefix         string
	MainAssignment       string
	MainNotification     string
	InternalAssignment   string
	InternalNotification string
	UpdatedAt            time.Time
	// TimeZone es la zona horaria de la instalación, y PublicAppURL su dirección pública.
	TimeZone     string
	PublicAppURL string
	// AIURL y AIModel son **el motor de IA** de la instalación. `AIURL` vacía es «no integrado»
	// (docs/modules/ai.md).
	AIURL   string
	AIModel string
	Brand   Brand
}

// Brand es el estado de la marca.
type Brand struct {
	// Light y Dark dicen si hay logo propio en cada hueco, y con qué datos.
	Light LogoInfo
	Dark  LogoInfo
}

// LogoInfo es un hueco del logo.
type LogoInfo struct {
	Filled    bool
	FileName  string
	Size      int64
	UpdatedAt time.Time
	Width     int
	Height    int
}

// Colors es el color institucional ya resuelto para los dos temas de fábrica.
//
// El administrador elige **un** color; de ahí salen los cuatro valores que la aplicación necesita:
// el del tema claro, el del oscuro —aclarado si hace falta que se lea sobre fondo oscuro—, y el
// texto que va encima del botón en cada uno.
type Colors struct {
	Light     string
	Dark      string
	OnLight   string
	OnDark    string
	IsDefault bool
}

// Access devuelve el método de entrada y las dos configuraciones, con sus secretos, porque son para
// hablar con el directorio y no para enseñarlas.
//
// **Es lo que lee el camino de entrada en cada intento**, así que se lee de la base y no se recuerda:
// cambiar el método o el directorio tiene que valer en el intento siguiente, sin reiniciar nada.
//
// Devuelve el tipo de `shared` y no uno de este módulo a propósito: quien lo consume es `auth`, que no
// puede importar `settings`, y ese tipo es el contrato entre los dos
// (docs/modules/settings.md, sección 5.8).
func (s *Service) Access() (config.Access, error) {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return config.Access{}, traducir(err)
	}
	directorio, err := s.repo.Directory()
	if err != nil {
		return config.Access{}, traducir(err)
	}
	keycloak, err := s.repo.Keycloak()
	if err != nil {
		return config.Access{}, traducir(err)
	}

	return config.Access{
		Method:    instalacion.EntryMethod,
		Directory: directorioDe(directorio),
		OIDC:      keycloakDe(keycloak),
		TimeZone:  zonaHorariaDe(instalacion.TimeZone),
	}, nil
}

// Prober es lo que este módulo necesita para **probar** una configuración antes de guardarla.
//
// Se declara aquí, en quien lo usa, y lo cumple el módulo `auth`, que es el que sabe hablar con un
// directorio y con Keycloak: probar una conexión es hablar el protocolo, y eso no se duplica
// (docs/arquitectura.md, sección 4).
type Prober interface {
	// ProbeDirectory abre la conexión y valida la cuenta de servicio.
	ProbeDirectory(config.Directory) error
	// ProbeKeycloak lee el documento de descubrimiento del reino.
	ProbeKeycloak(config.OIDC) error
}

// SetProber conecta quien sabe probar las dos configuraciones. Se llama una vez, al arrancar.
func (s *Service) SetProber(prober Prober) { s.prober = prober }

// TestDirectory prueba una configuración de directorio **sin guardarla**.
//
// La contraseña vacía quiere decir «usa la que hay guardada», que es lo que permite probar después de
// cambiar el servidor sin volver a escribirla.
func (s *Service) TestDirectory(input DirectoryInput) error {
	if s.prober == nil {
		return errors.New("no hay quien pruebe el directorio")
	}

	cfg, err := s.directoryDeInput(input, true)
	if err != nil {
		return err
	}
	if !cfg.Configured() {
		return ErrDirectoryIncomplete
	}

	if err := s.prober.ProbeDirectory(cfg); err != nil {
		// El motivo —no contesta, el bind no vale— va al log, que es donde lo ve quien administra. A la
		// pantalla le llega una clave suya y no la del camino de entrada: la prueba es de este módulo y
		// quien la lee está en Configuración, no entrando.
		logs.LogWarning("la prueba del directorio ha fallado: " + err.Error())
		return ErrDirectoryUnreachable
	}

	return nil
}

// TestKeycloak prueba una configuración de Keycloak **sin guardarla**.
func (s *Service) TestKeycloak(input KeycloakInput) error {
	if s.prober == nil {
		return errors.New("no hay quien pruebe Keycloak")
	}

	cfg, err := s.keycloakDeInput(input, true)
	if err != nil {
		return err
	}
	if !cfg.Configured() {
		return ErrKeycloakIncomplete
	}

	if err := s.prober.ProbeKeycloak(cfg); err != nil {
		logs.LogWarning("la prueba de Keycloak ha fallado: " + err.Error())
		return ErrKeycloakUnreachable
	}

	return nil
}

// TestAI prueba el motor de IA **sin guardarlo**.
//
// Se prueba **la dirección que llega** —lo que hay en pantalla— y **nunca la guardada**: al abrir la
// pantalla el campo viene relleno con lo que hay en la base, así que vaciarlo es borrarlo a
// propósito, y decir que no hay nada que probar es más útil que probar una dirección que la persona
// acaba de quitar. No toca la base.
func (s *Service) TestAI(url string) error {
	direccion := direccionPublicaDe(url)

	// La misma comprobación que la dirección pública, con la clave de este módulo. Vacía tampoco vale
	// aquí: no hay nada que probar.
	if direccion == "" {
		return ErrAIURLInvalid
	}
	if err := validarDireccionPublica(direccion); err != nil {
		return ErrAIURLInvalid
	}

	if s.ia == nil {
		return errors.New("no hay quien pruebe el motor de IA")
	}

	if err := s.ia.Probar(direccion); err != nil {
		// El motivo —no contesta, o contesta mal— va al log. A la pantalla le llega que esa dirección
		// no sirve, que es lo que tiene que hacer quien la está configurando.
		logs.LogWarning("la prueba del motor de IA ha fallado: " + err.Error())
		return ErrAIUnreachable
	}

	return nil
}

// Language es el idioma de la instalación.
//
// Lo lee el alta de cuentas para saber en qué idioma escribirle los correos a quien no se le ha
// elegido uno: es el único ajuste de la instalación que necesita otro módulo en su día a día.
func (s *Service) Language() (string, error) {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return "", traducir(err)
	}

	return instalacion.Language, nil
}

// Config lee la configuración entera, para la pantalla de administración.
func (s *Service) Config() (Config, error) {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return Config{}, traducir(err)
	}
	tickets, err := s.repo.Tickets()
	if err != nil {
		return Config{}, traducir(err)
	}

	directorio, err := s.repo.Directory()
	if err != nil {
		return Config{}, traducir(err)
	}
	keycloak, err := s.repo.Keycloak()
	if err != nil {
		return Config{}, traducir(err)
	}

	return Config{
		Name:                 nombreDeLaInstalacion(instalacion.InstallationName),
		EntryMethod:          instalacion.EntryMethod,
		TimeZone:             zonaHorariaDe(instalacion.TimeZone),
		PublicAppURL:         direccionPublicaDe(instalacion.PublicAppURL),
		Directory:            vistaDeDirectorio(directorio),
		Keycloak:             vistaDeKeycloak(keycloak),
		Language:             instalacion.Language,
		PrimaryColor:         instalacion.PrimaryColor,
		NumberPrefix:         tickets.NumberPrefix,
		MainAssignment:       tickets.MainAssignment,
		MainNotification:     tickets.MainNotification,
		InternalAssignment:   tickets.InternalAssignment,
		InternalNotification: tickets.InternalNotification,
		UpdatedAt:            instalacion.UpdatedAt,
		AIURL:                direccionPublicaDe(instalacion.AIURL),
		AIModel:              strings.TrimSpace(instalacion.AIModel),
		Brand: Brand{
			Light: s.logoInfo(VarianteClaro, instalacion.LogoLight),
			Dark:  s.logoInfo(VarianteOscuro, instalacion.LogoDark),
		},
	}, nil
}

// Update valida y guarda la configuración.
// ZonaHoraria es el nombre IANA de la zona de la instalación, y **nunca vacío**: quien la use
// formatea una fecha sin tener que preguntarse si hay zona. Es lo que necesitan los correos
// para escribir `{{cuando}}` en la hora de la instalación (docs/modules/settings.md, 15).
func (s *Service) ZonaHoraria() string {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return "UTC"
	}

	return zonaHorariaDe(instalacion.TimeZone)
}

// DireccionPublica es **la base de los enlaces** de esta instalación: la que sale en los correos y
// la vuelta de Keycloak. La lee el módulo de autenticación a través de la interfaz que él declara,
// así que aquí no se sabe quién la usa (docs/modules/settings.md, decisión 14).
//
// Si la fila no la tiene puesta, devuelve vacío y **quien la use decide qué hacer**: la variable de
// entorno queda como respaldo en el cableado, no aquí.
func (s *Service) DireccionPublica() string {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return ""
	}

	return direccionPublicaDe(instalacion.PublicAppURL)
}

// AIDatos es el motor de IA resuelto: su dirección, su modelo y **si hay alguno puesto**.
type AIDatos struct {
	// URL es la dirección del motor, sin la barra del final.
	URL string
	// Modelo es el modelo con el que redacta. Puede venir vacío y quien lo use decide su valor de
	// fábrica.
	Modelo string
	// Hay en falso quiere decir «esta instalación no tiene motor»: el módulo de IA usa entonces el
	// respaldo del entorno (docs/modules/ai.md, decisión 2).
	Hay bool
}

// AI devuelve **el motor de IA ya resuelto**, para el módulo de IA: lo de la configuración y, si no
// hay nada, el respaldo del entorno, que lo pone el cableado. Tiene la misma forma que `SMTP()`
// (docs/modules/ai.md).
//
// Se lee de la base en cada petición, no se recuerda: cambiar la dirección o el modelo desde
// Configuración tiene que valer sin reiniciar nada.
func (s *Service) AI() (AIDatos, error) {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return AIDatos{}, traducir(err)
	}

	direccion := direccionPublicaDe(instalacion.AIURL)
	if direccion == "" {
		return AIDatos{}, nil
	}

	return AIDatos{
		URL:    direccion,
		Modelo: strings.TrimSpace(instalacion.AIModel),
		Hay:    true,
	}, nil
}
func (s *Service) Update(input UpdateInput, actor auth.Identity) (Config, error) {
	language := strings.TrimSpace(input.Language)
	if !contiene(languages, language) {
		return Config{}, ErrLanguageUnknown
	}

	color := strings.ToLower(strings.TrimSpace(input.PrimaryColor))
	if !colorPattern.MatchString(color) {
		return Config{}, ErrPrimaryColorInvalid
	}

	prefix := strings.TrimSpace(input.NumberPrefix)
	if !prefixPattern.MatchString(prefix) {
		return Config{}, ErrNumberPrefixInvalid
	}

	if !contiene(assignments, input.MainAssignment) || !contiene(assignments, input.InternalAssignment) {
		return Config{}, ErrAssignmentUnknown
	}
	if !contiene(notifications, input.MainNotification) || !contiene(notifications, input.InternalNotification) {
		return Config{}, ErrNotificationUnknown
	}

	// Avisar «al asignado» sin repartir no tiene sentido: no hay a quién avisar. Lo comprueba también
	// la base, pero se avisa aquí para poder decirlo con su clave.
	if soloAvisaAlAsignado(input.MainAssignment, input.MainNotification) ||
		soloAvisaAlAsignado(input.InternalAssignment, input.InternalNotification) {
		return Config{}, ErrNotificationWithoutAssigment
	}

	updatedByID := actorID(actor)

	// El nombre se limpia y, si se queda vacío, vuelve el de fábrica: **una instalación sin nombre no
	// se puede ni nombrar**, y la base tampoco lo admite.
	nombre := strings.TrimSpace(input.Name)
	if nombre == "" {
		nombre = NombreDeFabrica
	}
	if len([]rune(nombre)) > MaxNameLength {
		return Config{}, ErrNameTooLong
	}

	// El método de entrada: uno de los tres, y **tiene que estar configurado**. Elegir un método sin
	// configurarlo dejaría la instalación sin puerta para todo el mundo menos la cuenta de fábrica, y
	// eso se avisa aquí y no el día que alguien no puede entrar.
	metodo := strings.TrimSpace(input.EntryMethod)
	if !metodoValido(metodo) {
		return Config{}, ErrMethodUnknown
	}

	directorio, err := s.directoryDeInput(input.Directory, false)
	if err != nil {
		return Config{}, err
	}
	keycloak, err := s.keycloakDeInput(input.Keycloak, false)
	if err != nil {
		return Config{}, err
	}
	if err := s.metodoEstaConfigurado(metodo, directorio, keycloak); err != nil {
		return Config{}, err
	}

	// **La zona horaria y la dirección pública** (decisiones 14 y 15): la zona tiene que ser un
	// nombre que exista —se comprueba contra la base de zonas, que va incrustada en el binario— y la
	// dirección tiene que ser http o https con su host. Ni una cosa ni la otra se puede dar por
	// buena: la primera decide qué hora lee todo el mundo, y la segunda a dónde apuntan los enlaces
	// que salen en los correos.
	zona := strings.TrimSpace(input.TimeZone)
	if _, err := time.LoadLocation(zona); err != nil || zona == "" {
		return Config{}, ErrTimeZoneUnknown
	}

	direccion := direccionPublicaDe(input.PublicAppURL)
	if err := validarDireccionPublica(direccion); err != nil {
		return Config{}, err
	}

	// **El motor de IA** (docs/modules/ai.md): la dirección, vacía o http/https con su host. Vacía es
	// «no integrado» y se acepta; con algo escrito, se le exige lo mismo que a la dirección pública
	// —se reutiliza su comprobación, que es la misma regla— pero con su propia clave, porque quien la
	// lee está en la tarjeta del motor.
	motor := direccionPublicaDe(input.AIURL)
	if err := validarDireccionPublica(motor); err != nil {
		return Config{}, ErrAIURLInvalid
	}

	err = s.repo.UpdateInstallation(map[string]any{
		"installation_name": nombre,
		"entry_method":      metodo,
		"language":          language,
		"primary_color":     color,
		"time_zone":         zona,
		"public_app_url":    direccion,
		"ai_url":            motor,
		"ai_model":          strings.TrimSpace(input.AIModel),
	}, updatedByID)
	if err != nil {
		return Config{}, traducir(err)
	}

	if err := s.repo.UpdateDirectory(s.cambiosDeDirectorio(directorio), updatedByID); err != nil {
		return Config{}, traducir(err)
	}
	if err := s.repo.UpdateKeycloak(s.cambiosDeKeycloak(keycloak), updatedByID); err != nil {
		return Config{}, traducir(err)
	}

	err = s.repo.UpdateTickets(map[string]any{
		"number_prefix":         prefix,
		"main_assignment":       input.MainAssignment,
		"main_notification":     input.MainNotification,
		"internal_assignment":   input.InternalAssignment,
		"internal_notification": input.InternalNotification,
	}, updatedByID)
	if err != nil {
		return Config{}, traducir(err)
	}

	return s.Config()
}

// Public es lo que necesita la aplicación **antes de que nadie haya entrado**: el color
// institucional y si hay logo propio, con su versión para no enseñar el viejo.
type Public struct {
	// Name es el nombre de la instalación: la pantalla de entrada y el menú lateral lo enseñan **antes
	// de que nadie haya entrado**, así que viaja con la marca.
	Name string
	// Version es la versión del software, **sin la `v`**: la pone la interfaz al enseñarla. No es
	// configuración de la instalación —es un dato del software— y por eso es una constante del código
	// y no una columna (docs/modules/settings.md, sección 5.9).
	Version string
	Colors  Colors
	// LogoVersion cambia cada vez que se reemplaza un logo; vacío si no hay ninguno propio.
	LogoVersion string
	// HasLight y HasDark dicen qué huecos están puestos, para que el frontend sepa si pedir el logo.
	HasLight bool
	HasDark  bool
	// **TimeZone es la zona horaria de la instalacion** (nombre IANA): con ella se leen todas las
	// fechas de la interfaz. Viaja en la marca porque la entrada se pinta antes de entrar y porque
	// cualquier pantalla necesita formatear una fecha (docs/modules/settings.md, decision 15).
	TimeZone string
}

// ResolveColors resuelve un color **sin guardarlo**, para la vista previa de la pantalla.
//
// Se hace aquí y no en el frontend a propósito: la regla de qué se lee y qué no es una sola, y tener
// una copia en el navegador es la forma segura de que un día digan cosas distintas.
func (s *Service) ResolveColors(color string) Colors {
	return colorsFrom(color)
}

// PublicBrand lee la marca para el arranque de la aplicación.
func (s *Service) PublicBrand() (Public, error) {
	instalacion, err := s.repo.Installation()
	if err != nil {
		return Public{}, traducir(err)
	}

	publico := Public{
		Name:        nombreDeLaInstalacion(instalacion.InstallationName),
		Version:     version.Version,
		Colors:      colorsFrom(instalacion.PrimaryColor),
		HasLight:    instalacion.LogoLight != nil,
		HasDark:     instalacion.LogoDark != nil,
		LogoVersion: s.versionDe(instalacion),
		TimeZone:    zonaHorariaDe(instalacion.TimeZone),
	}

	return publico, nil
}

// SaveLogo guarda el logo de una variante.
//
// **El nuevo se escribe antes de borrar el viejo** (docs/modules/settings.md, sección 5.3): si algo
// falla a mitad, se queda el logo anterior y no una instalación sin logo.
func (s *Service) SaveLogo(variant string, entrada io.Reader, actor auth.Identity) (LogoInfo, error) {
	if variant != VarianteClaro && variant != VarianteOscuro {
		return LogoInfo{}, ErrLogoVariantUnknown
	}

	logo, err := InspectLogo(entrada)
	if err != nil {
		return LogoInfo{}, err
	}

	if err := os.MkdirAll(s.carpeta(), 0o755); err != nil {
		return LogoInfo{}, fmt.Errorf("no se pudo preparar la carpeta de la marca: %w", err)
	}

	nombre := NombreDeLogo(variant, s.now().Format("20060102T150405"), logo.Extension)
	if err := os.WriteFile(filepath.Join(s.carpeta(), nombre), logo.Bytes, 0o644); err != nil {
		return LogoInfo{}, fmt.Errorf("no se pudo guardar el logo: %w", err)
	}

	anterior, err := s.repo.Installation()
	if err != nil {
		s.borrarArchivo(&nombre)
		return LogoInfo{}, traducir(err)
	}

	columna := "logo_light"
	if variant == VarianteOscuro {
		columna = "logo_dark"
	}

	if err := s.repo.UpdateInstallation(map[string]any{columna: nombre}, actorID(actor)); err != nil {
		// La base no aceptó el cambio: se quita el archivo que se acababa de escribir para no dejar
		// basura, y se queda el logo que había.
		s.borrarArchivo(&nombre)
		return LogoInfo{}, traducir(err)
	}

	// Ya está guardado el nuevo: ahora sí se borra el viejo.
	s.borrarArchivo(apuntado(anterior.LogoLight, anterior.LogoDark, variant))

	logs.LogSuccess("logo de la instalación reemplazado (" + variant + ")")

	return s.logoInfo(variant, &nombre), nil
}

// RemoveLogo quita el logo propio de una variante y devuelve ese hueco al de fábrica.
func (s *Service) RemoveLogo(variant string, actor auth.Identity) error {
	if variant != VarianteClaro && variant != VarianteOscuro {
		return ErrLogoVariantUnknown
	}

	actual, err := s.repo.Installation()
	if err != nil {
		return traducir(err)
	}

	columna := "logo_light"
	if variant == VarianteOscuro {
		columna = "logo_dark"
	}

	if err := s.repo.UpdateInstallation(map[string]any{columna: nil}, actorID(actor)); err != nil {
		return traducir(err)
	}

	s.borrarArchivo(apuntado(actual.LogoLight, actual.LogoDark, variant))

	return nil
}

// LogoFile es el archivo que hay que servir para un tema.
type LogoFile struct {
	Path        string
	ContentType string
	Version     string
}

// LogoFor devuelve el logo que toca a ese tema, con la caída al otro hueco y, si no hay ninguno, el
// aviso de que no hay logo propio (docs/modules/settings.md, sección 5.1).
func (s *Service) LogoFor(theme string) (LogoFile, error) {
	if theme != VarianteClaro && theme != VarianteOscuro {
		theme = VarianteClaro
	}

	instalacion, err := s.repo.Installation()
	if err != nil {
		return LogoFile{}, traducir(err)
	}

	// El del tema que toca; si está vacío, el del otro hueco; si tampoco, no hay logo propio.
	candidatos := []*string{apuntado(instalacion.LogoLight, instalacion.LogoDark, theme)}
	if theme == VarianteClaro {
		candidatos = append(candidatos, instalacion.LogoDark)
	} else {
		candidatos = append(candidatos, instalacion.LogoLight)
	}

	for _, nombre := range candidatos {
		if nombre == nil || *nombre == "" {
			continue
		}

		ruta := filepath.Join(s.carpeta(), *nombre)
		info, err := os.Stat(ruta)
		if err != nil || info.IsDir() {
			// El archivo ya no está —alguien lo borró a mano—: se comporta como si el hueco estuviera
			// vacío. La marca no es una función crítica.
			continue
		}

		return LogoFile{
			Path:        ruta,
			ContentType: TipoDeLogo(*nombre),
			Version:     info.ModTime().UTC().Format("20060102T150405"),
		}, nil
	}

	return LogoFile{}, ErrLogoNotFound
}

// colorsFrom resuelve el color institucional para los dos temas de fábrica.
//
// El administrador elige uno; aquí se decide con cuál se lee bien en cada uno y de qué color va el
// texto encima. Es la única forma de que el color de la casa no rompa el contraste al cambiarlo.
func colorsFrom(elegido string) Colors {
	color := strings.ToLower(strings.TrimSpace(elegido))
	if !colorPattern.MatchString(color) {
		return colorsFrom(colorInstitucionalDeFabrica)
	}

	// Sobre el fondo claro tiene que leerse; si no llega, se oscurece hasta que llegue.
	claro := ajustarHasta(color, "#ffffff", 4.5, false)
	// Y sobre el fondo oscuro: si no llega, se aclara.
	oscuro := ajustarHasta(color, "#14181d", 4.5, true)

	return Colors{
		Light:     claro,
		Dark:      oscuro,
		OnLight:   textoEncima(claro),
		OnDark:    textoEncima(oscuro),
		IsDefault: color == colorInstitucionalDeFabrica,
	}
}

// colorInstitucionalDeFabrica es el azul que trae la aplicación, y el mismo que está en la
// migración: la fuente de verdad es la base, y esto es sólo la red por si llega algo raro.
const colorInstitucionalDeFabrica = "#1d4ed8"

// textoEncima elige el color del texto que va sobre ese color: blanco o casi negro, el que más
// contraste dé.
func textoEncima(color string) string {
	if contraste(color, "#ffffff") >= contraste(color, "#0d0f12") {
		return "#ffffff"
	}
	return "#0d0f12"
}

// ajustarHasta mueve la claridad del color hasta que se lea sobre ese fondo, sin cambiarle el tono.
//
// Se mueve poco a poco y se para en el primer valor que cumple: así el color del administrador se
// parece lo más posible a lo que eligió.
func ajustarHasta(color, fondo string, minimo float64, aclarar bool) string {
	if contraste(color, fondo) >= minimo {
		return color
	}

	r, g, b := aRGB(color)
	for paso := 1; paso <= 100; paso++ {
		factor := float64(paso) / 100

		candidato := mezclar(r, g, b, "#ffffff", factor)
		if !aclarar {
			candidato = mezclar(r, g, b, "#000000", factor)
		}

		if contraste(candidato, fondo) >= minimo {
			return candidato
		}
	}

	// Ni con el extremo se llega: se devuelve blanco o negro, que es lo más legible que hay.
	if aclarar {
		return "#ffffff"
	}
	return "#000000"
}

// mezclar acerca un color a otro en la proporción indicada (0 = el primero, 1 = el segundo).
func mezclar(desdeR, desdeG, desdeB int, hacia string, proporcion float64) string {
	hr, hg, hb := aRGB(hacia)

	r := int(math.Round(float64(desdeR) + (float64(hr)-float64(desdeR))*proporcion))
	g := int(math.Round(float64(desdeG) + (float64(hg)-float64(desdeG))*proporcion))
	b := int(math.Round(float64(desdeB) + (float64(hb)-float64(desdeB))*proporcion))

	return aHex(r, g, b)
}

// contraste calcula la razón de contraste entre dos colores, como la define la norma (WCAG 2.1).
func contraste(uno, otro string) float64 {
	a := luminancia(uno)
	b := luminancia(otro)

	if a < b {
		a, b = b, a
	}

	return (a + 0.05) / (b + 0.05)
}

func luminancia(color string) float64 {
	r, g, b := aRGB(color)

	canal := func(valor int) float64 {
		v := float64(valor) / 255
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}

	return 0.2126*canal(r) + 0.7152*canal(g) + 0.0722*canal(b)
}

// aRGB convierte un color hexadecimal en sus tres números.
func aRGB(color string) (int, int, int) {
	color = strings.TrimPrefix(strings.TrimSpace(color), "#")

	var r, g, b int
	_, _ = fmt.Sscanf(color, "%02x%02x%02x", &r, &g, &b)

	return r, g, b
}

// aHex vuelve a armar el color hexadecimal.
func aHex(r, g, b int) string {
	return fmt.Sprintf("#%02x%02x%02x", limitar(r), limitar(g), limitar(b))
}

func limitar(valor int) int {
	if valor < 0 {
		return 0
	}
	if valor > 255 {
		return 255
	}
	return valor
}

// carpeta es la carpeta de la marca, dentro de la carpeta de archivos.
func (s *Service) carpeta() string {
	return filepath.Join(s.filesPath, carpetaMarca)
}

// logoInfo describe un hueco del logo, con los datos de su archivo si lo tiene.
func (s *Service) logoInfo(variant string, nombre *string) LogoInfo {
	if nombre == nil || *nombre == "" {
		return LogoInfo{}
	}

	info := LogoInfo{Filled: true, FileName: *nombre}

	if datos, err := os.Stat(filepath.Join(s.carpeta(), *nombre)); err == nil {
		info.Size = datos.Size()
		info.UpdatedAt = datos.ModTime().UTC()

		if ancho, alto := medidasDeArchivo(filepath.Join(s.carpeta(), *nombre)); ancho > 0 {
			info.Width, info.Height = ancho, alto
		}
	}

	return info
}

// medidasDeArchivo lee el tamaño de la imagen guardada, para poder enseñarlo en la pantalla.
func medidasDeArchivo(ruta string) (int, int) {
	archivo, err := os.Open(ruta)
	if err != nil {
		return 0, 0
	}
	defer archivo.Close()

	cabeza := make([]byte, MaxLogoBytes)
	leidos, err := archivo.Read(cabeza)
	if err != nil && leidos == 0 {
		return 0, 0
	}

	logo, err := InspectLogo(bytes.NewReader(cabeza[:leidos]))
	if err != nil {
		return 0, 0
	}

	return logo.Width, logo.Height
}

// versionDe es el sello de la marca: la fecha del logo más reciente. Es lo que se le pone a la
// dirección del logo para que un navegador con el viejo en su caché lo vuelva a pedir.
func (s *Service) versionDe(instalacion repositories.InstallationSettings) string {
	var ultima time.Time

	for _, nombre := range []*string{instalacion.LogoLight, instalacion.LogoDark} {
		if nombre == nil || *nombre == "" {
			continue
		}

		if info, err := os.Stat(filepath.Join(s.carpeta(), *nombre)); err == nil && info.ModTime().After(ultima) {
			ultima = info.ModTime()
		}
	}

	if ultima.IsZero() {
		return ""
	}

	return ultima.UTC().Format("20060102T150405")
}

// borrarArchivo quita un archivo de la marca, y no se queja si ya no estaba.
func (s *Service) borrarArchivo(nombre *string) {
	if nombre == nil || *nombre == "" {
		return
	}

	if err := os.Remove(filepath.Join(s.carpeta(), *nombre)); err != nil && !os.IsNotExist(err) {
		logs.LogWarning("no se pudo borrar un logo antiguo: " + err.Error())
	}
}

// apuntado devuelve el nombre del archivo de esa variante.
func apuntado(claro, oscuro *string, variant string) *string {
	if variant == VarianteOscuro {
		return oscuro
	}
	return claro
}

// soloAvisaAlAsignado dice si esa combinación es la que no tiene sentido.
func soloAvisaAlAsignado(asignacion, aviso string) bool {
	return asignacion == "ninguna" && aviso == "al_asignado"
}

func contiene(valores []string, valor string) bool {
	for _, candidato := range valores {
		if candidato == valor {
			return true
		}
	}
	return false
}

// actorID es quién cambia la configuración, o nada si es la cuenta de fábrica, que no está en la
// tabla de cuentas.
func actorID(actor auth.Identity) *int64 {
	if actor.Factory || actor.ID == 0 {
		return nil
	}
	return &actor.ID
}

// traducir convierte el error del repositorio en el del módulo.
func traducir(err error) error {
	if errors.Is(err, repositories.ErrSettingsNotFound) {
		return ErrSettingsNotFound
	}
	return err
}

// nombreDeLaInstalacion devuelve el nombre guardado, o el de fábrica si no hay ninguno.
//
// La columna no admite vacío, así que esto es la red que evita que una instalación se quede sin
// nombre si algún día se escribe en la base por otra vía.
// zonaHorariaDe devuelve la zona horaria guardada, o UTC si la fila no la trae. Nunca
// devuelve vacio: la marca publica no puede llevar una zona que la interfaz no sepa usar.
func zonaHorariaDe(guardada string) string {
	zona := strings.TrimSpace(guardada)
	if zona == "" {
		return "UTC"
	}

	return zona
}

// direccionPublicaDe deja la direccion publica como se guarda y como se usa en un enlace: sin
// espacios y sin la barra del final, que es lo que se le pega a /set-password/...
// validarDireccionPublica comprueba que una dirección sirva para un enlace: http o https, con su
// host. Vacía es válida: es «no configurada». La usan la pantalla de Configuración y el paso de la
// dirección del asistente, para que las dos digan lo mismo (docs/primer-arranque.md, sección 3).
func validarDireccionPublica(direccion string) error {
	if direccion == "" {
		return nil
	}

	u, err := url.Parse(direccion)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ErrPublicURLInvalid
	}

	return nil
}

func direccionPublicaDe(guardada string) string {
	return strings.TrimRight(strings.TrimSpace(guardada), "/")
}
func nombreDeLaInstalacion(nombre string) string {
	limpio := strings.TrimSpace(nombre)
	if limpio == "" {
		return NombreDeFabrica
	}

	return limpio
}

// metodoValido dice si el método es uno de los tres.
func metodoValido(metodo string) bool { return auth.MethodIsValid(metodo) }

// metodoEstaConfigurado comprueba que el método elegido tiene con qué funcionar.
//
// **No se elige un método a medias**: si se guardara `ad` sin directorio, la pantalla de entrada no
// tendría por dónde dejar entrar a nadie, y la única puerta sería la cuenta de fábrica.
func (s *Service) metodoEstaConfigurado(metodo string, directorio config.Directory, keycloak config.OIDC) error {
	switch metodo {
	case auth.MethodAD:
		if !directorio.Configured() {
			return ErrMethodNotConfigured
		}
	case auth.MethodKeycloak:
		if !keycloak.Configured() {
			return ErrMethodNotConfigured
		}
	}

	return nil
}

// directoryDeInput arma la configuración del directorio a partir de lo que llega de la pantalla.
//
// Con `usarLaGuardada`, la contraseña vacía se rellena con la que hay en la base: es lo que permite
// **probar** una configuración sin volver a escribir el secreto. Al guardar, en cambio, vacío quiere
// decir «no la cambies» y se resuelve en `cambiosDeDirectorio`.
func (s *Service) directoryDeInput(input DirectoryInput, usarLaGuardada bool) (config.Directory, error) {
	directorio := config.Directory{
		Host:         strings.TrimSpace(input.Host),
		Port:         strings.TrimSpace(input.Port),
		UseTLS:       input.UseTLS,
		BindDN:       strings.TrimSpace(input.BindDN),
		BindPass:     input.BindPassword,
		SearchBase:   strings.TrimSpace(input.SearchBase),
		UserFilter:   strings.TrimSpace(input.UserFilter),
		AttrEmail:    strings.TrimSpace(input.AttrEmail),
		AttrName:     strings.TrimSpace(input.AttrName),
		AttrLastName: strings.TrimSpace(input.AttrLastName),
		AttrID:       strings.TrimSpace(input.AttrID),
	}

	// Los valores de fábrica, para no obligar a escribir siete cosas para empezar.
	if directorio.Port == "" {
		directorio.Port = "389"
	}
	if directorio.UserFilter == "" {
		directorio.UserFilter = "(mail=%s)"
	}
	if directorio.AttrEmail == "" {
		directorio.AttrEmail = "mail"
	}
	if directorio.AttrName == "" {
		directorio.AttrName = "givenName"
	}
	if directorio.AttrLastName == "" {
		directorio.AttrLastName = "sn"
	}
	if directorio.AttrID == "" {
		directorio.AttrID = "objectGUID"
	}

	if usarLaGuardada && directorio.BindPass == "" {
		guardada, err := s.repo.Directory()
		if err != nil {
			return config.Directory{}, traducir(err)
		}
		directorio.BindPass = guardada.BindPassword
	}

	// Un directorio configurado a medias se rechaza: sin base de búsqueda no se busca a nadie.
	if directorio.Host != "" && directorio.SearchBase == "" {
		return config.Directory{}, ErrDirectoryIncomplete
	}

	return directorio, nil
}

// keycloakDeInput arma la configuración de Keycloak a partir de lo que llega de la pantalla, con la
// misma regla para el secreto que el directorio.
func (s *Service) keycloakDeInput(input KeycloakInput, usarLaGuardada bool) (config.OIDC, error) {
	keycloak := config.OIDC{
		Issuer:         strings.TrimRight(strings.TrimSpace(input.Issuer), "/"),
		InternalIssuer: strings.TrimRight(strings.TrimSpace(input.InternalIssuer), "/"),
		ClientID:       strings.TrimSpace(input.ClientID),
		ClientSecret:   input.ClientSecret,
		RedirectURI:    strings.TrimSpace(input.RedirectURI),
	}

	if usarLaGuardada && keycloak.ClientSecret == "" {
		guardada, err := s.repo.Keycloak()
		if err != nil {
			return config.OIDC{}, traducir(err)
		}
		keycloak.ClientSecret = guardada.ClientSecret
	}

	for _, direccion := range []string{keycloak.Issuer, keycloak.InternalIssuer} {
		if direccion != "" && !config.ValidIssuerURL(direccion) {
			return config.OIDC{}, ErrKeycloakIncomplete
		}
	}
	if keycloak.InternalIssuer != "" && keycloak.Issuer == "" {
		return config.OIDC{}, ErrKeycloakIncomplete
	}
	// Un Keycloak configurado a medias tampoco vale: sin vuelta no se puede volver.
	if keycloak.Issuer != "" && (keycloak.ClientID == "" || keycloak.RedirectURI == "") {
		return config.OIDC{}, ErrKeycloakIncomplete
	}

	return keycloak, nil
}

// cambiosDeDirectorio arma lo que se escribe en la base: **la contraseña sólo si viene**.
func (s *Service) cambiosDeDirectorio(directorio config.Directory) map[string]any {
	cambios := map[string]any{
		"host":           directorio.Host,
		"port":           directorio.Port,
		"use_tls":        directorio.UseTLS,
		"bind_dn":        directorio.BindDN,
		"search_base":    directorio.SearchBase,
		"user_filter":    directorio.UserFilter,
		"attr_email":     directorio.AttrEmail,
		"attr_name":      directorio.AttrName,
		"attr_last_name": directorio.AttrLastName,
		"attr_id":        directorio.AttrID,
	}
	if directorio.BindPass != "" {
		cambios["bind_password"] = directorio.BindPass
	}

	return cambios
}

// cambiosDeKeycloak arma lo que se escribe en la base: el secreto sólo si viene.
func (s *Service) cambiosDeKeycloak(keycloak config.OIDC) map[string]any {
	cambios := map[string]any{
		"issuer":          keycloak.Issuer,
		"internal_issuer": keycloak.InternalIssuer,
		"client_id":       keycloak.ClientID,
		"redirect_uri":    keycloak.RedirectURI,
	}
	if keycloak.ClientSecret != "" {
		cambios["client_secret"] = keycloak.ClientSecret
	}

	return cambios
}

// directorioDe traduce la fila a la configuración que usa el camino de AD.
func directorioDe(fila repositories.DirectorySettings) config.Directory {
	return config.Directory{
		Host:         fila.Host,
		Port:         fila.Port,
		UseTLS:       fila.UseTLS,
		BindDN:       fila.BindDN,
		BindPass:     fila.BindPassword,
		SearchBase:   fila.SearchBase,
		UserFilter:   fila.UserFilter,
		AttrEmail:    fila.AttrEmail,
		AttrName:     fila.AttrName,
		AttrLastName: fila.AttrLastName,
		AttrID:       fila.AttrID,
	}
}

// keycloakDe traduce la fila a la configuración que usa el camino de Keycloak.
func keycloakDe(fila repositories.KeycloakSettings) config.OIDC {
	return config.OIDC{
		Issuer:         fila.Issuer,
		InternalIssuer: fila.InternalIssuer,
		ClientID:       fila.ClientID,
		ClientSecret:   fila.ClientSecret,
		RedirectURI:    fila.RedirectURI,
	}
}

// vistaDeDirectorio es lo que se enseña: la configuración **sin la contraseña**, y si hay una puesta.
func vistaDeDirectorio(fila repositories.DirectorySettings) DirectoryView {
	return DirectoryView{
		DirectoryInput: DirectoryInput{
			Host:         fila.Host,
			Port:         fila.Port,
			UseTLS:       fila.UseTLS,
			BindDN:       fila.BindDN,
			SearchBase:   fila.SearchBase,
			UserFilter:   fila.UserFilter,
			AttrEmail:    fila.AttrEmail,
			AttrName:     fila.AttrName,
			AttrLastName: fila.AttrLastName,
			AttrID:       fila.AttrID,
		},
		PasswordSet: fila.BindPassword != "",
	}
}

// vistaDeKeycloak es lo que se enseña: la configuración sin el secreto.
func vistaDeKeycloak(fila repositories.KeycloakSettings) KeycloakView {
	return KeycloakView{
		KeycloakInput: KeycloakInput{
			Issuer:         fila.Issuer,
			InternalIssuer: fila.InternalIssuer,
			ClientID:       fila.ClientID,
			RedirectURI:    fila.RedirectURI,
		},
		SecretSet: fila.ClientSecret != "",
	}
}
