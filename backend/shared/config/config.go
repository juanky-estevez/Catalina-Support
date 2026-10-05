// Package config lee la configuración del backend desde el entorno.
//
// Sólo se leen variables que cambian entre entornos. Lo que vale lo mismo en
// desarrollo y en producción se escribe como constante en el código, no como variable
// (docs/arquitectura.md, sección 8).
//
// **Y guarda los tipos que viajan entre módulos sin que ninguno importe a otro**: el directorio de la
// organización, el reino de Keycloak y el `Access` que los junta con el método de entrada. No salen del
// entorno —viven en la base y se cambian desde la pantalla—, pero sí son el contrato entre `settings`
// y `auth`, y ese contrato tiene que estar en `shared` (docs/modules/settings.md, sección 5.8).
package config

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// Directory es la configuración del directorio de la organización (AD por LDAP).
//
// Va junta en un tipo y no suelta por el `Config` porque se usa entera o no se usa: quien la lee es
// el camino de AD, y lo primero que mira es si hay host.
type Directory struct {
	Host     string
	Port     string
	UseTLS   bool
	BindDN   string
	BindPass string
	// SearchBase es dónde se busca, y UserFilter el filtro con `%s` donde va el identificador.
	SearchBase string
	UserFilter string
	// Los atributos de los que salen el correo, el nombre, los apellidos y el identificador.
	AttrEmail    string
	AttrName     string
	AttrLastName string
	AttrID       string
}

// Configured dice si hay directorio al que preguntar.
func (d Directory) Configured() bool { return d.Host != "" }

// OIDC es la configuración del camino de Keycloak.
//
// Los cuatro valores van juntos porque sin los cuatro no hay camino: un emisor sin cliente no sirve
// de nada, y un cliente sin su secreto tampoco. **Sin emisor configurado no hay camino de Keycloak**, y
// como con AD, no hay un interruptor aparte (docs/modules/auth.md, sección 5.3).
type OIDC struct {
	// Issuer es la base del reino: de ahí sale el documento de descubrimiento. **El mismo para el
	// navegador y la identidad del reino**: si se ven por direcciones distintas, manda la que ve
	// el navegador, que es quien firma la vuelta.
	Issuer string
	// InternalIssuer permite conectar al mismo reino desde la red interna. Vacío usa Issuer.
	InternalIssuer string
	// ClientID y ClientSecret son del cliente confidencial que se da de alta en el reino. Su
	// secreto no se escribe en la documentación ni en el código: **vive en la base**, se guarda desde
	// Configuración y no se devuelve nunca por la API (docs/modules/settings.md, sección 5.8).
	ClientID     string
	ClientSecret string
	// RedirectURI es a donde vuelve Keycloak: la ruta de vuelta de este módulo.
	RedirectURI string
}

// Configured dice si hay Keycloak al que ir.
func (o OIDC) Configured() bool {
	return o.Issuer != "" && o.ClientID != "" && o.RedirectURI != ""
}

// Access es **cómo se entra en esta instalación**: el método que está puesto y las dos
// configuraciones de los caminos de directorio, con sus secretos.
//
// Es lo que `settings` guarda y lo que `auth` lee para entrar. Vive aquí, en `shared`, porque lo usan
// los dos y ninguno puede importar al otro (docs/arquitectura.md, sección 4). **No se enseña nunca
// tal cual**: lo que sale por la API va sin los secretos (docs/modules/settings.md, sección 5.8).
type Access struct {
	// Method es `local`, `ad` o `keycloak`.
	Method    string
	Directory Directory
	OIDC      OIDC
	// TimeZone es **la zona horaria de la instalación** (nombre IANA). Va aquí porque es un dato de
	// cómo es esta instalación, como el método con el que se entra, y quien tiene que escribir una
	// fecha —los correos de cuenta— necesita saberlo (docs/modules/settings.md, decisión 15).
	TimeZone string
}

// Config es la configuración del backend ya validada.
type Config struct {
	// Environment lo consume go-logs para nombrar el archivo y cada línea.
	Environment string

	// AppPort es el puerto en el que escucha el servidor HTTP.
	AppPort string

	// LogsFolder es la carpeta donde go-logs escribe los archivos del día.
	LogsFolder string

	// FilesPath es la carpeta donde viven los archivos que sube la gente: los adjuntos de los
	// tickets y el logo de la instalación. Es obligatoria en los dos entornos: una instalación sin
	// carpeta de archivos no puede guardar nada, y es mejor no arrancar que arrancar a medias.
	FilesPath string

	// TokenSecret firma el token de sesión. Obligatorio en producción: sin él no se puede
	// validar ninguna sesión.
	TokenSecret string

	// AdminPassword es la contraseña de la cuenta de fábrica `admin`, que no está en la base
	// (docs/usuarios-y-permisos.md, sección 8).
	AdminPassword string

	// PublicAppURL es la base de los enlaces que van en los correos. Obligatoria en producción:
	// un enlace escrito con un host a mano no puede servir en los dos entornos a la vez
	// (docs/modules/mail.md, sección 4).
	PublicAppURL string

	// AI es el motor de IA que redacta el motivo y la última acción del ticket
	// (docs/modules/ai.md). **Puede faltar**: sin motor, los dos campos se quedan sin texto y la
	// mesa de ayuda funciona entera; es la condición que manda sobre este módulo.
	AIURL      string
	AIModel    string
	AIPalabras int
	AIEspera   time.Duration

	// SetupMail son sugerencias para el correo del asistente de primer arranque. Sólo desarrollo
	// las declara; no configuran el envío ni sustituyen valores guardados en la instalación.
	SetupMailHost      string
	SetupMailPort      string
	SetupMailSecure    bool
	SetupMailUser      string
	SetupMailFromName  string
	SetupMailFromEmail string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresName     string
}

// Load lee el entorno y falla si falta algo obligatorio: es preferible no arrancar a
// arrancar a medias con una configuración incompleta.
func Load() (Config, error) {
	cfg := Config{
		Environment:   get("ENVIRONMENT", "dev"),
		AppPort:       get("APP_PORT", "11002"),
		LogsFolder:    get("LOGS_FOLDER", "/logs"),
		FilesPath:     get("FILES_PATH", "/files"),
		TokenSecret:   strings.TrimSpace(os.Getenv("TOKEN_SECRET")),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		PublicAppURL:  strings.TrimSpace(os.Getenv("PUBLIC_APP_URL")),
		AIURL:         strings.TrimSpace(os.Getenv("AI_URL")),
		AIModel:       get("AI_MODEL", "qwen2.5-1.5b-instruct"),
		AIPalabras:    getInt("AI_PALABRAS", 40),
		// **Cuatro minutos por defecto**, y no es un número redondo: medido, un ticket de 6 000
		// caracteres tarda ~82 s sólo en leerse y luego redacta a 5-9 palabras por segundo
		// (`docs/modules/ai.md`, sección 2). Un tiempo corto haría fallar justo los tickets largos.
		AIEspera:           time.Duration(getInt("AI_ESPERA_SEGUNDOS", 240)) * time.Second,
		SetupMailHost:      strings.TrimSpace(os.Getenv("SETUP_MAIL_HOST")),
		SetupMailPort:      strings.TrimSpace(os.Getenv("SETUP_MAIL_PORT")),
		SetupMailSecure:    getBool("SETUP_MAIL_SECURE", false),
		SetupMailUser:      strings.TrimSpace(os.Getenv("SETUP_MAIL_USER")),
		SetupMailFromName:  strings.TrimSpace(os.Getenv("SETUP_MAIL_FROM_NAME")),
		SetupMailFromEmail: strings.TrimSpace(os.Getenv("SETUP_MAIL_FROM_EMAIL")),
		PostgresHost:       get("POSTGRES_HOST", "database"),
		PostgresPort:       get("PGPORT", "11003"),
		PostgresUser:       os.Getenv("POSTGRES_USER"),
		PostgresPassword:   os.Getenv("POSTGRES_PASSWORD"),
		PostgresName:       os.Getenv("POSTGRES_DB"),
	}

	var missing []string
	if cfg.PostgresUser == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if cfg.PostgresPassword == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if cfg.PostgresName == "" {
		missing = append(missing, "POSTGRES_DB")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("faltan variables de entorno obligatorias: %s", strings.Join(missing, ", "))
	}

	// La carpeta de archivos no puede quedar vacía: sin ella no se puede guardar un adjunto ni el
	// logo de la instalación, y fallar al subir algo es peor que no arrancar.
	if cfg.FilesPath == "" {
		return Config{}, fmt.Errorf("falta FILES_PATH, la carpeta donde se guardan los archivos")
	}

	// **La dirección pública ya no es obligatoria en el entorno** (docs/modules/settings.md,
	// decisión 14): vive en la configuración de la instalación, que se cambia desde la pantalla, y
	// esta variable queda como respaldo. Sin ninguna de las dos, los correos con enlace **fallan con
	// su clave** en vez de mandar un enlace roto, así que arrancar sin ella es seguro.

	// Ni sin el secreto de firma: sin él no se puede validar ninguna sesión.
	if cfg.Environment == "prod" && cfg.TokenSecret == "" {
		return Config{}, fmt.Errorf("falta TOKEN_SECRET, obligatorio en producción")
	}

	// Ni sin la contraseña de la cuenta de fábrica: es la única forma de entrar en una instalación
	// recién puesta.
	if cfg.Environment == "prod" && cfg.AdminPassword == "" {
		return Config{}, fmt.Errorf("falta ADMIN_PASSWORD, obligatoria en producción")
	}

	return cfg, nil
}

// DSN es la cadena de conexión de GORM.
//
// sslmode=disable va fijo: la base de datos vive en la red interna de Compose y no
// habla TLS. Si algún día sale de ahí, se convierte en variable de entorno.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresName,
	)
}

// getInt lee un número del entorno, con su valor de siempre si falta o no es un número.
func getInt(key string, fallback int) int {
	valor := strings.TrimSpace(os.Getenv(key))
	if valor == "" {
		return fallback
	}

	numero, err := strconv.Atoi(valor)
	if err != nil || numero <= 0 {
		return fallback
	}

	return numero
}

func getBool(key string, fallback bool) bool {
	valor := strings.TrimSpace(os.Getenv(key))
	if valor == "" {
		return fallback
	}

	booleano, err := strconv.ParseBool(valor)
	if err != nil {
		return fallback
	}

	return booleano
}

func get(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

// ValidIssuerURL admite bases de reino HTTP(S), sin credenciales ni componentes ambiguos.
func ValidIssuerURL(value string) bool {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" || u.RawPath != "" {
		return false
	}
	p := strings.TrimRight(u.Path, "/")
	return p == "" || path.Clean(p) == p
}
