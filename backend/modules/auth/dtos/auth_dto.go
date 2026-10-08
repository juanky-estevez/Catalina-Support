// Package dtos define lo que entra y lo que sale por la API del módulo auth.
package dtos

import (
	"time"

	"catalina-support/backend/shared/auth"
)

// UserResponse es la cuenta tal y como la ve el frontend.
//
// **Es el mismo objeto en la entrada y en `me`**: tener dos formas de describir a alguien es la
// manera segura de que un día digan cosas distintas (docs/modules/auth.md, decisión 20).
type UserResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Origin   string `json:"origin"`
	// Factory distingue la cuenta de fábrica, que no está en la base de datos y no se puede tocar.
	Factory bool `json:"factory"`
}

// NewUserResponse arma la cuenta que se enseña.
func NewUserResponse(identity auth.Identity) UserResponse {
	return UserResponse{
		ID:       identity.ID,
		Name:     identity.Name,
		LastName: identity.LastName,
		Email:    identity.Email,
		Role:     identity.Role,
		Origin:   identity.Origin,
		Factory:  identity.Factory,
	}
}

// MethodsResponse dice cómo se entra en la instalación.
//
// Va el nombre del método —`local`, `ad` o `keycloak`— y además cuál de los tres caminos está abierto,
// porque la pantalla no tiene por qué conocer los nombres: con el método en AD se sigue entrando por el
// mismo formulario de correo y contraseña, y con el de Keycloak no hay formulario.
type MethodsResponse struct {
	Method   string `json:"method"`
	Local    bool   `json:"local"`
	AD       bool   `json:"ad"`
	Keycloak bool   `json:"keycloak"`
}

// LoginRequest es lo que llega al entrar. **Un solo campo**, que además acepta `admin` para la
// cuenta de fábrica (docs/modules/auth.md, decisión 19).
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse es el token, cuándo caduca y quién es.
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt string       `json:"expiresAt"`
	User      UserResponse `json:"user"`
}

// MeResponse es la sesión que sigue viva.
type MeResponse struct {
	User UserResponse `json:"user"`
}

// StatusResponse es la respuesta de lo que no devuelve nada más.
type StatusResponse struct {
	Status string `json:"status"`
}

// Ok es el «todo ha ido bien».
func Ok() StatusResponse { return StatusResponse{Status: "ok"} }

// ForgotRequest pide el enlace de recuperación.
type ForgotRequest struct {
	Email string `json:"email"`
}

// ResetRequest establece la contraseña con el token del enlace.
type ResetRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// ChangeRequest cambia la contraseña propia, con la actual.
type ChangeRequest struct {
	CurrentPassword string `json:"currentPassword"`
	Password        string `json:"password"`
}

// NewLoginResponse arma la respuesta de la entrada.
func NewLoginResponse(token string, expiresAt time.Time, identity auth.Identity) LoginResponse {
	return LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
		User:      NewUserResponse(identity),
	}
}
