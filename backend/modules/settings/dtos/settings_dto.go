// Package dtos define lo que entra y lo que sale por la API del módulo settings.
package dtos

import (
	"time"

	"catalina-support/backend/modules/settings/services"
)

// SettingsResponse es la configuración entera, tal y como la ve un Administrador.
type SettingsResponse struct {
	Name string `json:"name"`
	// EntryMethod es el método de entrada que está puesto: `local`, `ad` o `keycloak`.
	EntryMethod string `json:"entryMethod"`
	// TimeZone es la zona horaria de la instalación (nombre IANA) y PublicAppURL su dirección
	// pública (docs/modules/settings.md, decisiones 14 y 15).
	TimeZone     string `json:"timeZone"`
	PublicAppURL string `json:"publicAppUrl"`
	// Directory y Keycloak van **sin sus secretos**: en su lugar dicen si hay uno puesto.
	Directory            DirectoryDto `json:"directory"`
	Keycloak             KeycloakDto  `json:"keycloak"`
	Language             string       `json:"language"`
	PrimaryColor         string       `json:"primaryColor"`
	NumberPrefix         string       `json:"numberPrefix"`
	MainAssignment       string       `json:"mainAssignment"`
	MainNotification     string       `json:"mainNotification"`
	InternalAssignment   string       `json:"internalAssignment"`
	InternalNotification string       `json:"internalNotification"`
	UpdatedAt            string       `json:"updatedAt"`
	// AIURL y AIModel son **el motor de IA**: su dirección y su modelo. Vacíos es «no integrado».
	AIURL   string   `json:"aiUrl"`
	AIModel string   `json:"aiModel"`
	Brand   BrandDto `json:"brand"`
}

// BrandDto es el estado de la marca: los dos huecos.
type BrandDto struct {
	Light LogoDto `json:"light"`
	Dark  LogoDto `json:"dark"`
}

// LogoDto es un hueco del logo. `filled` en falso quiere decir que ese hueco usa el de fábrica.
type LogoDto struct {
	Filled    bool   `json:"filled"`
	FileName  string `json:"fileName,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// DirectoryDto es la configuración del directorio, **sin la contraseña**: `passwordSet` dice si hay
// una puesta, y la contraseña no sale nunca por aquí.
type DirectoryDto struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	UseTLS       bool   `json:"useTls"`
	BindDN       string `json:"bindDn"`
	SearchBase   string `json:"searchBase"`
	UserFilter   string `json:"userFilter"`
	AttrEmail    string `json:"attrEmail"`
	AttrName     string `json:"attrName"`
	AttrLastName string `json:"attrLastName"`
	AttrID       string `json:"attrId"`
	PasswordSet  bool   `json:"passwordSet"`
	// BindPassword sólo va **de entrada**: vacío quiere decir «no la cambies». Lo que se devuelve
	// lleva `passwordSet` y nunca la contraseña.
	BindPassword string `json:"bindPassword,omitempty"`
}

// KeycloakDto es la configuración de Keycloak, **sin el secreto**.
type KeycloakDto struct {
	Issuer         string `json:"issuer"`
	InternalIssuer string `json:"internalIssuer"`
	ClientID       string `json:"clientId"`
	RedirectURI    string `json:"redirectUri"`
	SecretSet      bool   `json:"secretSet"`
	// ClientSecret sólo va de entrada, con la misma regla que la contraseña del directorio.
	ClientSecret string `json:"clientSecret,omitempty"`
}

// UpdateSettingsRequest es lo que llega al guardar. Va todo: es un `PUT`.
type UpdateSettingsRequest struct {
	Name                 string       `json:"name"`
	EntryMethod          string       `json:"entryMethod"`
	Directory            DirectoryDto `json:"directory"`
	Keycloak             KeycloakDto  `json:"keycloak"`
	Language             string       `json:"language"`
	PrimaryColor         string       `json:"primaryColor"`
	NumberPrefix         string       `json:"numberPrefix"`
	MainAssignment       string       `json:"mainAssignment"`
	MainNotification     string       `json:"mainNotification"`
	InternalAssignment   string       `json:"internalAssignment"`
	InternalNotification string       `json:"internalNotification"`
	// TimeZone es la zona horaria de la instalación (nombre IANA, `America/Guayaquil`). Decide
	// cómo se leen las fechas; las guardadas siguen en UTC (docs/modules/settings.md, decisión 15).
	TimeZone string `json:"timeZone"`
	// PublicAppURL es la dirección pública: la base de los enlaces de los correos y de la vuelta
	// de Keycloak (docs/modules/settings.md, decisión 14).
	PublicAppURL string `json:"publicAppUrl"`
	// AIURL y AIModel son **el motor de IA**: su dirección y su modelo. Vacíos es «no integrado»
	// (docs/modules/ai.md).
	AIURL   string `json:"aiUrl"`
	AIModel string `json:"aiModel"`
}

// AITestRequest es lo que llega para probar el motor de IA. `URL` vacía quiere decir «prueba la que
// hay guardada», que es lo que permite comprobar el motor ya configurado sin volver a escribirlo.
type AITestRequest struct {
	URL string `json:"url"`
}

// BrandResponse es lo que la aplicación necesita **antes de que nadie haya entrado**: el color
// institucional ya resuelto para los dos temas de fábrica, y si hay logo propio.
type BrandResponse struct {
	// Name es el nombre de la instalación: es lo que la pantalla de entrada y el menú lateral enseñan.
	Name string `json:"name"`
	// TimeZone es la zona horaria de la instalación (nombre IANA): con ella se leen todas las fechas
	// de la interfaz. Viaja en la marca porque la entrada se pinta antes de entrar (decisión 15).
	TimeZone string `json:"timeZone"`
	// Version es la versión del software, **sin la `v`**: la `v` la pone la interfaz. Va aquí porque
	// la marca es lo que la aplicación pide al arrancar, y así el menú y la pantalla de entrada la
	// tienen sin una segunda llamada (docs/modules/settings.md, sección 5.9).
	Version      string    `json:"version"`
	PrimaryColor string    `json:"primaryColor"`
	Colors       ColorsDto `json:"colors"`
	Logo         BrandDto  `json:"logo"`
	LogoVersion  string    `json:"logoVersion"`
}

// ColorsDto son los cuatro valores del color institucional, ya resueltos.
type ColorsDto struct {
	Light     string `json:"light"`
	Dark      string `json:"dark"`
	OnLight   string `json:"onLight"`
	OnDark    string `json:"onDark"`
	IsDefault bool   `json:"isDefault"`
}

// NewSettingsResponse arma la configuración que se enseña.
func NewSettingsResponse(config services.Config) SettingsResponse {
	return SettingsResponse{
		Name:         config.Name,
		EntryMethod:  config.EntryMethod,
		TimeZone:     config.TimeZone,
		PublicAppURL: config.PublicAppURL,
		Directory: DirectoryDto{
			Host:         config.Directory.Host,
			Port:         config.Directory.Port,
			UseTLS:       config.Directory.UseTLS,
			BindDN:       config.Directory.BindDN,
			SearchBase:   config.Directory.SearchBase,
			UserFilter:   config.Directory.UserFilter,
			AttrEmail:    config.Directory.AttrEmail,
			AttrName:     config.Directory.AttrName,
			AttrLastName: config.Directory.AttrLastName,
			AttrID:       config.Directory.AttrID,
			PasswordSet:  config.Directory.PasswordSet,
		},
		Keycloak: KeycloakDto{
			Issuer:         config.Keycloak.Issuer,
			InternalIssuer: config.Keycloak.InternalIssuer,
			ClientID:       config.Keycloak.ClientID,
			RedirectURI:    config.Keycloak.RedirectURI,
			SecretSet:      config.Keycloak.SecretSet,
		},
		Language:             config.Language,
		PrimaryColor:         config.PrimaryColor,
		NumberPrefix:         config.NumberPrefix,
		MainAssignment:       config.MainAssignment,
		MainNotification:     config.MainNotification,
		InternalAssignment:   config.InternalAssignment,
		InternalNotification: config.InternalNotification,
		UpdatedAt:            config.UpdatedAt.UTC().Format(time.RFC3339),
		AIURL:                config.AIURL,
		AIModel:              config.AIModel,
		Brand: BrandDto{
			Light: newLogoDto(config.Brand.Light),
			Dark:  newLogoDto(config.Brand.Dark),
		},
	}
}

// NewBrandResponse arma lo que se enseña en la pantalla de entrada.
func NewBrandResponse(public services.Public) BrandResponse {
	return BrandResponse{
		Name:         public.Name,
		Version:      public.Version,
		TimeZone:     public.TimeZone,
		PrimaryColor: public.Colors.Light,
		Colors: ColorsDto{
			Light:     public.Colors.Light,
			Dark:      public.Colors.Dark,
			OnLight:   public.Colors.OnLight,
			OnDark:    public.Colors.OnDark,
			IsDefault: public.Colors.IsDefault,
		},
		Logo: BrandDto{
			Light: LogoDto{Filled: public.HasLight},
			Dark:  LogoDto{Filled: public.HasDark},
		},
		LogoVersion: public.LogoVersion,
	}
}

func newLogoDto(info services.LogoInfo) LogoDto {
	logo := LogoDto{
		Filled:   info.Filled,
		FileName: info.FileName,
		Size:     info.Size,
		Width:    info.Width,
		Height:   info.Height,
	}

	if !info.UpdatedAt.IsZero() {
		logo.UpdatedAt = info.UpdatedAt.Format(time.RFC3339)
	}

	return logo
}

// NewDirectoryInput traduce lo que llega de la pantalla a lo que entiende el servicio.
func NewDirectoryInput(dto DirectoryDto) services.DirectoryInput {
	return services.DirectoryInput{
		Host:         dto.Host,
		Port:         dto.Port,
		UseTLS:       dto.UseTLS,
		BindDN:       dto.BindDN,
		BindPassword: dto.BindPassword,
		SearchBase:   dto.SearchBase,
		UserFilter:   dto.UserFilter,
		AttrEmail:    dto.AttrEmail,
		AttrName:     dto.AttrName,
		AttrLastName: dto.AttrLastName,
		AttrID:       dto.AttrID,
	}
}

// NewKeycloakInput traduce lo que llega de la pantalla a lo que entiende el servicio.
func NewKeycloakInput(dto KeycloakDto) services.KeycloakInput {
	return services.KeycloakInput{
		Issuer:         dto.Issuer,
		InternalIssuer: dto.InternalIssuer,
		ClientID:       dto.ClientID,
		ClientSecret:   dto.ClientSecret,
		RedirectURI:    dto.RedirectURI,
	}
}
