package auth

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionDuration es lo que dura el token de sesión: 10 horas, sin renovación deslizante. El `exp`
// se pone al firmar y no se toca (docs/modules/auth.md, sección 3).
//
// Son horas y no días a propósito: como no hay tabla de sesiones, un token emitido no se puede
// revocar, así que su caducidad es la única puerta que se cierra sola.
const SessionDuration = 10 * time.Hour

// Los dos motivos por los que un token no vale. Se separan porque el frontend hace lo mismo con
// ambos —borrar el token y llevar a la pantalla de entrada— pero el log sí quiere distinguirlos.
var (
	ErrTokenInvalid = errors.New("el token de sesión no vale")
	ErrTokenExpired = errors.New("el token de sesión ha caducado")
)

// ErrNoSecret avisa de que no hay secreto con el que firmar: en producción el backend no arranca sin
// `TOKEN_SECRET`, y esto es la red que lo detiene.
var ErrNoSecret = errors.New("falta TOKEN_SECRET")

// TokenManager firma y valida los tokens de sesión.
type TokenManager struct {
	secret []byte
}

// NewTokenManager construye el firmante con el secreto de la configuración.
func NewTokenManager(secret string) (*TokenManager, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, ErrNoSecret
	}
	return &TokenManager{secret: []byte(secret)}, nil
}

// Sign emite un token para esa cuenta. El `sub` es el identificador de la cuenta, o `admin` para la
// de fábrica. No se meten el papel ni el nombre: se leen de la cuenta en cada petición, y meterlos
// aquí sería guardar una copia que envejece.
func (m *TokenManager) Sign(subject string, issuedAt time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(issuedAt.Add(SessionDuration)),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Verify comprueba la firma y la fecha, y devuelve el `sub`.
//
// `now` se recibe en lugar de mirarlo aquí dentro para que las pruebas puedan colocarse en el
// futuro sin dormir.
func (m *TokenManager) Verify(token string, now time.Time) (string, error) {
	claims := &jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return m.secret, nil
	},
		// Sólo HS256: sin esto, un token podría declarar "alg: none" o pedir otra firma y colarse.
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrTokenExpired
		}
		return "", ErrTokenInvalid
	}

	if claims.Subject == "" {
		return "", ErrTokenInvalid
	}

	return claims.Subject, nil
}

// BearerToken saca el token de la cabecera `Authorization`. El segundo valor dice si había uno con
// la forma esperada; el middleware no distingue ese caso del token que no vale, porque la respuesta
// es la misma.
func BearerToken(header string) (string, bool) {
	const prefix = "Bearer "

	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}

	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}

	return token, true
}

// SameSecret compara dos secretos en tiempo constante.
//
// Se usa con la contraseña de fábrica: no hay hash que comparar —la variable de entorno es la fuente
// de verdad— así que la comparación se hace aquí, sin que el tiempo de respuesta delate cuántos
// caracteres iniciales coincidían (docs/modules/auth.md, sección 4).
//
// Dos secretos vacíos **no** coinciden: `ConstantTimeCompare` da por iguales dos rebanadas vacías, y
// eso convertiría una variable sin definir en una puerta abierta.
func SameSecret(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
