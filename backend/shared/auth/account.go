package auth

import (
	"errors"
	"time"
)

// Los tres orígenes de una cuenta. Van juntos con los papeles porque son un valor cerrado que viaja
// a la base, al frontend y a los tres caminos de entrada (docs/modules/users.md, sección 2).
const (
	OriginLocal    = "local"
	OriginAD       = "ad"
	OriginKeycloak = "keycloak"
)

// Origins devuelve los tres orígenes, para validar lo que llega de fuera.
func Origins() []string { return []string{OriginLocal, OriginAD, OriginKeycloak} }

// OriginIsValid dice si el origen es uno de los tres.
func OriginIsValid(origin string) bool {
	for _, known := range Origins() {
		if origin == known {
			return true
		}
	}
	return false
}

// Account es una cuenta completa: lo que el módulo `users` guarda y lo que `auth` necesita para
// entrar, para cambiar una contraseña y para escribir un correo.
//
// Vive en `shared` y no en ninguno de los dos módulos por una razón de forma: **los dos la usan, y
// los dos declaran interfaces que hablan de ella** (`Accounts` en `auth`, `PasswordLinks` en
// `users`). Si viviera en uno de los dos, el otro tendría que importarlo y los dos módulos quedarían
// mirándose en círculo, que en Go no compila. `shared` es justo el sitio de lo que no es de nadie
// (docs/arquitectura.md, sección 4).
type Account struct {
	ID       int64
	Name     string
	LastName string
	Email    string
	// PasswordHash está vacío cuando la cuenta no tiene contraseña: en las de directorio nunca la
	// hay, y en las locales no la hay hasta que su dueño la establece desde el enlace.
	PasswordHash string
	Role         string
	Origin       string
	// ExternalID es el identificador en el directorio. Vacío en las cuentas locales.
	ExternalID  string
	Language    string
	IsActive    bool
	LastLoginAt *time.Time
}

// FullName arma el nombre completo, que es como se saluda en los correos y se enseña en las
// pantallas.
func (a Account) FullName() string {
	if a.Name == "" {
		return a.LastName
	}
	if a.LastName == "" {
		return a.Name
	}
	return a.Name + " " + a.LastName
}

// HasPassword dice si la cuenta tiene contraseña con la que entrar. Es la diferencia entre «no has
// acertado» y «todavía no la has establecido», que para quien entra es el mismo mensaje
// (docs/modules/auth.md, sección 5.1).
func (a Account) HasPassword() bool { return a.PasswordHash != "" }

// IsDirectory dice si la cuenta viene de un directorio, y por tanto su contraseña es de allí.
func (a Account) IsDirectory() bool {
	return a.Origin == OriginAD || a.Origin == OriginKeycloak
}

// DirectoryAccount es una persona tal y como la cuenta su directorio.
//
// Vive en `shared` por la misma razón que `Account`: la produce el camino de AD de `auth` y la
// consume `users` para crear o actualizar la cuenta, y ninguno de los dos puede importar al otro
// (docs/arquitectura.md, sección 4). **Los datos del directorio mandan**: nombre, apellidos y correo
// se actualizan en cada entrada (docs/modules/auth.md, sección 5.4).
type DirectoryAccount struct {
	// Origin es `ad` o `keycloak`, según por dónde haya entrado.
	Origin string
	// ExternalID es el identificador allí: el GUID de AD o el `sub` de Keycloak. Es lo que manda
	// cuando alguien cambia de correo.
	ExternalID string
	Email      string
	Name       string
	LastName   string
}

// Identity es la cuenta reducida a lo que se guarda en el contexto de la petición: sin la contraseña
// y sin nada que no haga falta para decidir permisos y para responder.
func (a Account) Identity() Identity {
	return Identity{
		ID:       a.ID,
		Name:     a.Name,
		LastName: a.LastName,
		Email:    a.Email,
		Role:     a.Role,
		Origin:   a.Origin,
		Language: a.Language,
		IsActive: a.IsActive,
	}
}

// FactoryIdentity es la identidad de la cuenta de fábrica, que no está en la base de datos.
func FactoryIdentity() Identity {
	return Identity{
		Factory:  true,
		Subject:  FactorySubject,
		Name:     "Administrador",
		Role:     RoleAdministrador,
		Origin:   OriginLocal,
		Language: "es",
		IsActive: true,
	}
}

// ErrAccountNotFound lo devuelve el módulo de cuentas cuando la cuenta no existe.
//
// Vive aquí, y no en `users`, porque quien lo pregunta es `auth`: así el módulo de cuentas puede
// devolver este error sin que `auth` tenga que conocer los errores de nadie.
var ErrAccountNotFound = errors.New("cuenta no encontrada")
