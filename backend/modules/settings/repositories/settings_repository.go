// Package repositories es el acceso a datos del módulo settings: la configuración de la instalación.
package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrSettingsNotFound se devuelve cuando no se encuentra la fila de configuración, que es un fallo
// del que nadie puede salir: la migración la siembra y el módulo no la borra nunca.
var ErrSettingsNotFound = errors.New("settings.notFound")

// InstallationSettings es la fila única de la configuración propia de la instalación.
type InstallationSettings struct {
	ID       int64  `gorm:"primaryKey"`
	Language string `gorm:"column:language"`
	// EntryMethod es **el método de entrada: uno a la vez** (`local`, `ad` o `keycloak`). Los otros
	// dos quedan apagados, y se puede cambiar cuando haga falta (docs/modules/settings.md, 5.8).
	EntryMethod  string `gorm:"column:entry_method"`
	PrimaryColor string `gorm:"column:primary_color"`
	// InstallationName es **el nombre de la instalación**: la institución, la empresa o el equipo.
	// Es lo que se enseña donde antes decía «Catalina Support» (docs/modules/settings.md).
	InstallationName string `gorm:"column:installation_name"`
	// TimeZone es **la zona horaria de la instalación** (nombre IANA, `America/Guayaquil`): decide cómo
	// se leen las fechas, en la interfaz y en los correos. Las guardadas siguen en UTC
	// (docs/modules/settings.md, decisión 15).
	TimeZone string `gorm:"column:time_zone"`
	// PublicAppURL es **la dirección pública** de la instalación: la base de los enlaces de los correos
	// y la vuelta de Keycloak. Vacía es «no configurada» y se usa la variable de entorno
	// (docs/modules/settings.md, decisión 14).
	PublicAppURL string `gorm:"column:public_app_url"`
	// **El motor de IA**: su dirección y su modelo. Vacíos es «no integrado».
	AIURL   string `gorm:"column:ai_url"`
	AIModel string `gorm:"column:ai_model"`
	// InstalledAt es **el sello de instalación**: nulo mientras nadie haya terminado el asistente de
	// primer arranque. Nulo es «sin instalar»: sólo entonces se enseña la vista de instalación y su API
	// acepta configurar (docs/primer-arranque.md, sección 2).
	InstalledAt *time.Time `gorm:"column:installed_at"`
	// **El correo saliente** (docs/primer-arranque.md, sección 5): vive aquí, como el directorio y
	// Keycloak, y la variable de entorno queda de respaldo. La contraseña **no sale nunca por la API**.
	SMTPHost      string `gorm:"column:smtp_host"`
	SMTPPort      string `gorm:"column:smtp_port"`
	SMTPSecure    bool   `gorm:"column:smtp_secure"`
	SMTPUser      string `gorm:"column:smtp_user"`
	SMTPPassword  string `gorm:"column:smtp_password"`
	SMTPFromName  string `gorm:"column:smtp_from_name"`
	SMTPFromEmail string `gorm:"column:smtp_from_email"`
	// LogoLight y LogoDark son el **nombre del archivo** en el disco, no el archivo. Nulo es «no hay
	// logo propio» y se usa el de fábrica (docs/modules/settings.md, sección 5.3).
	LogoLight   *string   `gorm:"column:logo_light"`
	LogoDark    *string   `gorm:"column:logo_dark"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	UpdatedByID *int64    `gorm:"column:updated_by_id"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (InstallationSettings) TableName() string { return "installation_settings" }

// TicketSettings es la fila única de lo que configura el comportamiento de los tickets.
type TicketSettings struct {
	ID                   int64     `gorm:"primaryKey"`
	NumberPrefix         string    `gorm:"column:number_prefix"`
	MainAssignment       string    `gorm:"column:main_assignment"`
	MainNotification     string    `gorm:"column:main_notification"`
	InternalAssignment   string    `gorm:"column:internal_assignment"`
	InternalNotification string    `gorm:"column:internal_notification"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
	UpdatedByID          *int64    `gorm:"column:updated_by_id"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (TicketSettings) TableName() string { return "ticket_settings" }

// DirectorySettings es la fila única de la configuración del directorio de la organización (AD).
//
// **La contraseña de la cuenta de servicio vive aquí** y no sale nunca por la API: quien la lee es el
// camino de AD, y a la pantalla se le dice sólo si está puesta.
type DirectorySettings struct {
	ID           int64     `gorm:"primaryKey"`
	Host         string    `gorm:"column:host"`
	Port         string    `gorm:"column:port"`
	UseTLS       bool      `gorm:"column:use_tls"`
	BindDN       string    `gorm:"column:bind_dn"`
	BindPassword string    `gorm:"column:bind_password"`
	SearchBase   string    `gorm:"column:search_base"`
	UserFilter   string    `gorm:"column:user_filter"`
	AttrEmail    string    `gorm:"column:attr_email"`
	AttrName     string    `gorm:"column:attr_name"`
	AttrLastName string    `gorm:"column:attr_last_name"`
	AttrID       string    `gorm:"column:attr_id"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
	UpdatedByID  *int64    `gorm:"column:updated_by_id"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (DirectorySettings) TableName() string { return "directory_settings" }

// KeycloakSettings es la fila única de la configuración del camino de Keycloak. El secreto del
// cliente se guarda aquí y no sale nunca por la API.
type KeycloakSettings struct {
	ID           int64     `gorm:"primaryKey"`
	Issuer       string    `gorm:"column:issuer"`
	ClientID     string    `gorm:"column:client_id"`
	ClientSecret string    `gorm:"column:client_secret"`
	RedirectURI  string    `gorm:"column:redirect_uri"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
	UpdatedByID  *int64    `gorm:"column:updated_by_id"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (KeycloakSettings) TableName() string { return "keycloak_settings" }

// SettingsRepository lee y escribe la configuración.
type SettingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository construye el repositorio.
func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Installation devuelve la configuración de la instalación.
func (r *SettingsRepository) Installation() (InstallationSettings, error) {
	var ajustes InstallationSettings

	err := r.db.Where("id = 1").First(&ajustes).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return InstallationSettings{}, ErrSettingsNotFound
	}

	return ajustes, err
}

// Tickets devuelve la configuración de los tickets.
func (r *SettingsRepository) Tickets() (TicketSettings, error) {
	var ajustes TicketSettings

	err := r.db.Where("id = 1").First(&ajustes).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return TicketSettings{}, ErrSettingsNotFound
	}

	return ajustes, err
}

// Directory devuelve la configuración del directorio.
func (r *SettingsRepository) Directory() (DirectorySettings, error) {
	var ajustes DirectorySettings

	err := r.db.Where("id = 1").First(&ajustes).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DirectorySettings{}, ErrSettingsNotFound
	}

	return ajustes, err
}

// Keycloak devuelve la configuración del camino de Keycloak.
func (r *SettingsRepository) Keycloak() (KeycloakSettings, error) {
	var ajustes KeycloakSettings

	err := r.db.Where("id = 1").First(&ajustes).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return KeycloakSettings{}, ErrSettingsNotFound
	}

	return ajustes, err
}

// UpdateDirectory cambia las columnas indicadas de la configuración del directorio.
func (r *SettingsRepository) UpdateDirectory(cambios map[string]any, updatedByID *int64) error {
	return r.update("directory_settings", cambios, updatedByID)
}

// UpdateKeycloak cambia las columnas indicadas de la configuración de Keycloak.
func (r *SettingsRepository) UpdateKeycloak(cambios map[string]any, updatedByID *int64) error {
	return r.update("keycloak_settings", cambios, updatedByID)
}

// UpdateInstallation cambia las columnas indicadas. El mapa lo arma el servicio, que es quien decide
// qué se puede cambiar y con qué reglas.
func (r *SettingsRepository) UpdateInstallation(cambios map[string]any, updatedByID *int64) error {
	return r.update("installation_settings", cambios, updatedByID)
}

// UpdateTickets cambia las columnas indicadas de la configuración de los tickets.
func (r *SettingsRepository) UpdateTickets(cambios map[string]any, updatedByID *int64) error {
	return r.update("ticket_settings", cambios, updatedByID)
}

// update escribe en la fila única y avisa si no había ninguna, que sería un fallo del que nadie
// puede salir: la migración la siembra y aquí no se crea.
func (r *SettingsRepository) update(tabla string, cambios map[string]any, updatedByID *int64) error {
	cambios["updated_at"] = time.Now()
	cambios["updated_by_id"] = updatedByID

	result := r.db.Table(tabla).Where("id = 1").Updates(cambios)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSettingsNotFound
	}

	return nil
}
