// Package auth reúne la mecánica del token de sesión y quién es quien hace la petición.
//
// Vive en `shared` y no en el módulo `auth` por la regla de modularidad: el middleware que protege
// las rutas vive en `shared/middleware` y `shared` no puede importar de `modules`
// (docs/arquitectura.md, sección 4). El módulo `auth` usa esta mecánica igual que la usa el
// middleware (docs/modules/auth.md, sección 3).
package auth

import (
	"context"
	"errors"
)

// Los cuatro papeles del producto (docs/usuarios-y-permisos.md, sección 3). Son fijos: no se crean
// roles nuevos ni se editan permisos desde la aplicación, así que son constantes y no una tabla.
const (
	RoleUsuario       = "usuario"
	RoleSoporte       = "soporte"
	RoleDesarrollo    = "desarrollo"
	RoleAdministrador = "administrador"
)

// Roles devuelve los cuatro papeles, para validar lo que llega de fuera.
func Roles() []string {
	return []string{RoleUsuario, RoleSoporte, RoleDesarrollo, RoleAdministrador}
}

// RoleIsValid dice si el papel es uno de los cuatro.
func RoleIsValid(role string) bool {
	for _, known := range Roles() {
		if role == known {
			return true
		}
	}
	return false
}

// FactorySubject es el `sub` que llevan los tokens de la cuenta de fábrica. No es un identificador
// de la base, porque esa cuenta no está en la base (docs/usuarios-y-permisos.md, sección 8).
const FactorySubject = "admin"

// Identity es quién hace la petición: lo que el middleware deja en el contexto y leen los
// controladores.
//
// Lleva el papel y el idioma resueltos porque se leen de la cuenta **en cada petición**: así, dar de
// baja a alguien o cambiarle el papel surte efecto en la siguiente llamada, sin esperar a que caduque
// su token (docs/usuarios-y-permisos.md, sección 6).
type Identity struct {
	// ID es el identificador de la cuenta. Cero en la cuenta de fábrica.
	ID int64
	// Factory distingue la cuenta de fábrica, que no está en la base y no se puede tocar.
	Factory  bool
	Name     string
	LastName string
	Email    string
	Role     string
	Origin   string
	Language string
	IsActive bool
	Subject  string
}

// FullName arma el nombre completo, que es como se enseña en las pantallas y en los correos.
func (i Identity) FullName() string {
	if i.Name == "" {
		return i.LastName
	}
	if i.LastName == "" {
		return i.Name
	}
	return i.Name + " " + i.LastName
}

// Can administra el papel. Es la comprobación de permisos: no decide nada por su cuenta más allá de
// «¿su papel está entre estos?».
func (i Identity) Can(roles ...string) bool {
	for _, role := range roles {
		if i.Role == role {
			return true
		}
	}
	return false
}

// ErrIdentityNotFound lo devuelve el cargador cuando el `sub` del token ya no corresponde a ninguna
// cuenta utilizable: no existe, o existe y está desactivada. En los dos casos el middleware responde
// 401 con `auth.session.invalid`: para quien la usa, es lo mismo.
var ErrIdentityNotFound = errors.New("la cuenta no existe o está desactivada")

// IdentityLoader resuelve un `sub` del token en la cuenta que hay detrás.
//
// El middleware recibe esta función en lugar de hablar con la base: `shared` no puede importar de
// `modules`, así que es `main.go` quien la conecta con el servicio de `users`
// (docs/modules/auth.md, sección 3).
type IdentityLoader func(ctx context.Context, subject string) (Identity, error)

// Entrada del contexto. El tipo es propio y sin exportar para que nadie pueda pisar la identidad
// desde fuera de este paquete con una clave que se le ocurra.
type contextKey struct{}

var identityKey contextKey

// WithIdentity deja la identidad en el contexto de la petición.
func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityKey, identity)
}

// FromContext saca la identidad que dejó el middleware. El segundo valor dice si había alguna: en
// una ruta pública no la hay, y eso no es un error.
func FromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityKey).(Identity)
	return identity, ok
}

// MustFromContext es para los controladores que van detrás del middleware: si no hay identidad, es
// un fallo de programación (la ruta está mal montada), no algo que deba tratar el controlador.
func MustFromContext(ctx context.Context) Identity {
	identity, ok := FromContext(ctx)
	if !ok {
		panic("no hay identidad en el contexto: falta el middleware de autenticación en esta ruta")
	}
	return identity
}
