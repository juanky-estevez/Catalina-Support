package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Contraseñas: política y cifrado (docs/modules/auth.md, sección 4).
const (
	// PasswordMinLength es la única regla de la política, además de existir: **mínimo 8 caracteres**,
	// sin caducidad y sin preguntas de seguridad. Cada regla añadida empeora la contraseña que la
	// gente acaba eligiendo, y en una mesa de ayuda interna la contraseña no es lo que protege el
	// sistema: lo que lo protege es quién tiene cuenta (docs/usuarios-y-permisos.md, sección 7,
	// decisión del responsable del 2026-09-25, que bajó el mínimo de 12 a 8).
	PasswordMinLength = 8

	// bcryptCost es 12: el mínimo razonable hoy para que probar contraseñas a lo bruto salga caro.
	bcryptCost = 12
)

// ErrPasswordTooShort se devuelve cuando la contraseña no llega al mínimo.
var ErrPasswordTooShort = errors.New("auth.password.tooShort")

// ErrPasswordWrong se devuelve cuando la contraseña actual no es la que se ha escrito.
var ErrPasswordWrong = errors.New("auth.password.wrong")

// ErrPasswordNotLocal se devuelve al intentar cambiar la contraseña de una cuenta de directorio: su
// contraseña no está aquí.
var ErrPasswordNotLocal = errors.New("auth.password.notLocal")

// HashPassword cifra una contraseña con bcrypt.
//
// Nunca se guarda ni se registra una contraseña en claro: ni en el log, ni en un correo, ni en una
// respuesta de error.
func HashPassword(password string) (string, error) {
	if len([]rune(password)) < PasswordMinLength {
		return "", ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// PasswordMatches dice si esa contraseña es la de ese hash.
//
// Se cuentan las runas y no los bytes al medir: «contraseñá» tiene 8 caracteres aunque ocupe más, y
// contar bytes sería una regla invisible que no se puede explicar en la pantalla.
func PasswordMatches(hash, password string) bool {
	if hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// newToken genera el token de un enlace: 32 bytes aleatorios en base64url.
//
// No se firma ni se deriva de nada: vive en una tabla, así que no hace falta
// (docs/modules/auth.md, sección 2).
func newToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// hashToken calcula la huella del token, que es lo único que se guarda.
//
// Si alguien lee la base no puede reconstruir ningún enlace, y comprobar lo que llega es tan simple
// como calcular su huella y buscarla.
func hashToken(token string) string {
	suma := sha256.Sum256([]byte(token))
	return hex.EncodeToString(suma[:])
}

// TokenIsWellFormed evita ir a la base con cualquier cosa: un token válido son 43 caracteres de
// base64url. Se comprueba la forma antes de calcular su huella, no por seguridad —la búsqueda por
// huella ya la da— sino para no hacer trabajo inútil con basura.
func TokenIsWellFormed(token string) bool {
	token = strings.TrimSpace(token)
	if len(token) != 43 {
		return false
	}

	_, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil
}
