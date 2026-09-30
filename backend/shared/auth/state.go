package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"
)

// StateDuration es lo que dura el `state` de OIDC: cinco minutos.
//
// Es el tiempo que se le da a alguien para autenticarse en Keycloak y volver. Ni un segundo más: el
// `state` no se guarda en ningún sitio —no hay cookies y no hay tabla—, así que **su caducidad es lo
// único que impide que una vuelta vieja siga valiendo** (docs/modules/auth.md, sección 5.3).
const StateDuration = 5 * time.Minute

// Los dos motivos por los que un `state` no vale. El frontend los trata igual —no se ha entrado y
// se dice—, pero el log los distingue: uno es un intento de colarse y el otro, alguien que se
// entretuvo.
var (
	ErrStateInvalid = errors.New("el state no está firmado por nosotros")
	ErrStateExpired = errors.New("el state ha caducado")
)

// SignState acuña un `state`: un valor aleatorio, el momento en que se emitió y su firma.
//
// **Se firma con `TOKEN_SECRET` y no se guarda en ningún sitio.** Es lo que permite comprobar en la
// vuelta que la entrada la empezó esta instalación y que no ha pasado demasiado tiempo, sin cookies
// y sin tabla de estados pendientes.
//
// El valor aleatorio no se usa para nada más: está para que **dos `state` seguidos nunca sean
// iguales**, que es lo que impide reutilizar el de otra persona.
func SignState(secret string, now time.Time) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", ErrNoSecret
	}

	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// El cuerpo va en claro —lleva el momento en que se emitió, que hay que poder leer— y lo que lo
	// hace valer es la firma de al lado.
	emitido := strconv.FormatInt(now.Unix(), 10)
	cuerpo := base64.RawURLEncoding.EncodeToString(nonce) + "." + emitido

	return cuerpo + "." + firmaDeState(secret, cuerpo), nil
}

// VerifyState comprueba un `state` que vuelve: que lo firmamos nosotros y que no ha caducado.
//
// `now` se recibe en lugar de mirarlo aquí dentro para que las pruebas puedan colocarse en el
// futuro sin dormir.
func VerifyState(secret, value string, now time.Time) error {
	if strings.TrimSpace(secret) == "" {
		return ErrNoSecret
	}

	partes := strings.Split(value, ".")
	if len(partes) != 3 {
		return ErrStateInvalid
	}

	cuerpo := partes[0] + "." + partes[1]
	// En tiempo constante: comparar con `==` deja medir la firma letra a letra.
	if !hmac.Equal([]byte(partes[2]), []byte(firmaDeState(secret, cuerpo))) {
		return ErrStateInvalid
	}

	emitido, err := strconv.ParseInt(partes[1], 10, 64)
	if err != nil {
		return ErrStateInvalid
	}

	momento := time.Unix(emitido, 0)
	if now.Sub(momento) > StateDuration {
		return ErrStateExpired
	}
	// Un `state` emitido en el futuro tampoco vale: o el reloj de quien lo acuñó va mal, o alguien
	// está probando a estirar la caducidad. Con un minuto de margen por los desajustes de reloj.
	if momento.After(now.Add(time.Minute)) {
		return ErrStateInvalid
	}

	return nil
}

// firmaDeState es el HMAC del cuerpo con el secreto de la instalación, en base64url.
func firmaDeState(secret, cuerpo string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(cuerpo))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
