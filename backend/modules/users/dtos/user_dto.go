// Package dtos define lo que entra y lo que sale por la API del módulo users.
package dtos

import (
	"time"

	"catalina-support/backend/modules/users/services"
	"catalina-support/backend/shared/auth"
)

// CreateUserRequest es lo que llega para dar de alta una cuenta.
//
// El idioma es opcional: si no se elige, se le escriben los correos en español.
type CreateUserRequest struct {
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Origin   string `json:"origin"`
	Language string `json:"language"`
	// ExternalID sólo tiene sentido en las cuentas de directorio.
	ExternalID string `json:"externalId"`
}

// UserResponse es una cuenta tal y como la ve el frontend.
//
// Es **el mismo objeto** que devuelve `auth` en la entrada y en `me`, más los datos de la ficha que
// aquí sí importan (docs/modules/auth.md, decisión 20).
type UserResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	Origin      string `json:"origin"`
	Language    string `json:"language"`
	IsActive    bool   `json:"isActive"`
	HasPassword bool   `json:"hasPassword"`
	ExternalID  string `json:"externalId,omitempty"`
	LastLoginAt string `json:"lastLoginAt,omitempty"`
}

// UserEnvelope es la respuesta del alta, del reenvío del enlace y de las acciones sobre una cuenta.
type UserEnvelope struct {
	User UserResponse `json:"user"`
	// Invited dice si se le ha lanzado el correo del enlace. Es información útil para quien da de
	// alta: si es `false` y la cuenta es local, hay que volver a lanzarlo.
	Invited bool `json:"invited,omitempty"`
	// Warnings son avisos sobre algo que **sí se ha hecho**, pero conviene saber. Van como claves,
	// como los errores, y el frontend los traduce.
	Warnings []string `json:"warnings,omitempty"`
}

// UsersResponse es la lista de cuentas, con lo que necesita la paginación.
type UsersResponse struct {
	Users   []UserResponse `json:"users"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	PerPage int            `json:"perPage"`
}

// UpdateUserRequest es lo que llega al cambiar una cuenta.
//
// Los campos van con puntero porque hay que distinguir «no lo toques» de «déjalo vacío»: un `PATCH`
// que no manda el papel no está diciendo que la cuenta se quede sin papel.
type UpdateUserRequest struct {
	Name     *string `json:"name"`
	LastName *string `json:"lastName"`
	Email    *string `json:"email"`
	Role     *string `json:"role"`
	Language *string `json:"language"`
	IsActive *bool   `json:"isActive"`
}

// UpdateProfileRequest es lo que puede cambiar cualquiera de su propio perfil: su nombre, sus
// apellidos y su idioma, y nada más (docs/modules/users.md, sección 8).
type UpdateProfileRequest struct {
	Name     *string `json:"name"`
	LastName *string `json:"lastName"`
	Language *string `json:"language"`
}

// OriginRequest es el cambio de origen de una cuenta.
type OriginRequest struct {
	Origin string `json:"origin"`
	// ExternalID es el identificador en el directorio, cuando el origen nuevo es de directorio.
	ExternalID string `json:"externalId"`
}

// NewUsersResponse arma la lista que se enseña.
func NewUsersResponse(pagina services.Pagina) UsersResponse {
	usuarios := make([]UserResponse, 0, len(pagina.Accounts))
	for _, cuenta := range pagina.Accounts {
		usuarios = append(usuarios, NewUserResponse(cuenta))
	}

	return UsersResponse{
		Users:   usuarios,
		Total:   pagina.Total,
		Page:    pagina.Page,
		PerPage: pagina.PerPage,
	}
}

// NewUserResponse arma la cuenta que se enseña.
func NewUserResponse(account auth.Account) UserResponse {
	respuesta := UserResponse{
		ID:          account.ID,
		Name:        account.Name,
		LastName:    account.LastName,
		Email:       account.Email,
		Role:        account.Role,
		Origin:      account.Origin,
		Language:    account.Language,
		IsActive:    account.IsActive,
		HasPassword: account.HasPassword(),
		ExternalID:  account.ExternalID,
	}

	if account.LastLoginAt != nil {
		respuesta.LastLoginAt = account.LastLoginAt.UTC().Format(time.RFC3339)
	}

	return respuesta
}
