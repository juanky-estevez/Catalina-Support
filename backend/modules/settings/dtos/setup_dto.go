package dtos

import "catalina-support/backend/modules/settings/services"

// SetupRequest es lo que llega en **un paso** del asistente de primer arranque. Los cuatro pasos
// comparten cuerpo porque comparten forma: cada uno trae lo suyo y deja lo demás vacío
// (docs/primer-arranque.md, sección 3).
type SetupRequest struct {
	// Paso 1: la instalación.
	Name     string `json:"name"`
	Language string `json:"language"`

	// Paso 2: cómo se entra.
	EntryMethod string       `json:"entryMethod"`
	Directory   DirectoryDto `json:"directory"`
	Keycloak    KeycloakDto  `json:"keycloak"`

	// Paso 3: dónde está.
	TimeZone     string `json:"timeZone"`
	PublicAppURL string `json:"publicAppUrl"`

	// Paso 4: el correo saliente.
	Mail MailDto `json:"mail"`
}

// MailDto es el correo saliente tal y como llega de la pantalla. `Password` vacío quiere decir **«no la
// cambies»**: la contraseña guardada no sale nunca por la API, así que la pantalla no puede mandarla de
// vuelta. Es la misma regla que el directorio y Keycloak.
type MailDto struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	Secure    bool   `json:"secure"`
	User      string `json:"user"`
	Password  string `json:"password"`
	FromName  string `json:"fromName"`
	FromEmail string `json:"fromEmail"`
}

// NewSetupStep traduce el cuerpo que llega al paso que entiende el servicio.
func NewSetupStep(entrada SetupRequest) services.PasoDeInstalacion {
	return services.PasoDeInstalacion{
		Name:         entrada.Name,
		Language:     entrada.Language,
		EntryMethod:  entrada.EntryMethod,
		Directory:    NewDirectoryInput(entrada.Directory),
		Keycloak:     NewKeycloakInput(entrada.Keycloak),
		TimeZone:     entrada.TimeZone,
		PublicAppURL: entrada.PublicAppURL,
		Mail: services.MailInput{
			Host:      entrada.Mail.Host,
			Port:      entrada.Mail.Port,
			Secure:    entrada.Mail.Secure,
			User:      entrada.Mail.User,
			Password:  entrada.Mail.Password,
			FromName:  entrada.Mail.FromName,
			FromEmail: entrada.Mail.FromEmail,
		},
	}
}

// SetupResponse es el estado de la instalación tal y como lo ve el asistente: si ya está sellada —y
// entonces no se enseña— y qué hay puesto, para **seguir donde se dejó**.
//
// **Sin secretos**: de la contraseña del correo sólo se dice si hay una puesta.
type SetupResponse struct {
	Installed    bool   `json:"installed"`
	Name         string `json:"name"`
	Language     string `json:"language"`
	EntryMethod  string `json:"entryMethod"`
	TimeZone     string `json:"timeZone"`
	PublicAppURL string `json:"publicAppUrl"`

	Directory DirectoryDto `json:"directory"`
	Keycloak  KeycloakDto  `json:"keycloak"`
	Mail      MailViewDto  `json:"mail"`

	// AiAvailable dice si el motor de IA responde. **Es opcional**: sin él, la instalación funciona
	// entera y sólo se queda sin los dos resúmenes (docs/modules/ai.md).
	AiAvailable bool `json:"aiAvailable"`
}

// MailViewDto es el correo saliente **sin su contraseña**: en su lugar dice si hay una guardada.
type MailViewDto struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	Secure      bool   `json:"secure"`
	User        string `json:"user"`
	FromName    string `json:"fromName"`
	FromEmail   string `json:"fromEmail"`
	PasswordSet bool   `json:"passwordSet"`
}

// NewSetupResponse arma la respuesta del asistente a partir del estado del servicio.
func NewSetupResponse(estado services.EstadoDeInstalacion) SetupResponse {
	return SetupResponse{
		Installed:    estado.Installed,
		Name:         estado.Name,
		Language:     estado.Language,
		EntryMethod:  estado.EntryMethod,
		TimeZone:     estado.TimeZone,
		PublicAppURL: estado.PublicAppURL,
		Directory: DirectoryDto{
			Host:         estado.Directory.Host,
			Port:         estado.Directory.Port,
			UseTLS:       estado.Directory.UseTLS,
			BindDN:       estado.Directory.BindDN,
			SearchBase:   estado.Directory.SearchBase,
			UserFilter:   estado.Directory.UserFilter,
			AttrEmail:    estado.Directory.AttrEmail,
			AttrName:     estado.Directory.AttrName,
			AttrLastName: estado.Directory.AttrLastName,
			AttrID:       estado.Directory.AttrID,
			PasswordSet:  estado.Directory.PasswordSet,
		},
		Keycloak: KeycloakDto{
			Issuer:         estado.Keycloak.Issuer,
			InternalIssuer: estado.Keycloak.InternalIssuer,
			ClientID:       estado.Keycloak.ClientID,
			RedirectURI:    estado.Keycloak.RedirectURI,
			SecretSet:      estado.Keycloak.SecretSet,
		},
		Mail: MailViewDto{
			Host:        estado.MailHost,
			Port:        estado.MailPort,
			Secure:      estado.MailSecure,
			User:        estado.MailUser,
			FromName:    estado.MailFromName,
			FromEmail:   estado.MailFromEmail,
			PasswordSet: estado.MailSet,
		},
		AiAvailable: estado.AiAvailable,
	}
}
