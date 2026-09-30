// Package controllers es la capa HTTP del módulo auth: lee la petición, llama al servicio y
// responde. Los permisos los pone el middleware sobre la ruta
// (docs/usuarios-y-permisos.md, sección 10).
package controllers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/auth/dtos"
	"catalina-support/backend/modules/auth/services"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/httpx"
)

// KeyInternal es lo que se responde cuando el fallo es nuestro: sin detalle, y con el detalle en el
// log.
const KeyInternal = "error interno"

// AuthController atiende los endpoints de entrada y de contraseñas.
type AuthController struct {
	service *services.Service
}

// NewAuthController construye el controlador.
func NewAuthController(service *services.Service) *AuthController {
	return &AuthController{service: service}
}

// Methods dice cómo se entra en esta instalación.
//
// Es **público**: lo lee la pantalla de entrada antes de que nadie haya entrado, y no lleva ningún
// dato de nadie, sólo el método que está puesto (docs/modules/auth.md, decisión 27).
func (c *AuthController) Methods(w http.ResponseWriter, r *http.Request) {
	caminos, err := c.service.Methods()
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.MethodsResponse{
		Method:   caminos.Method,
		Local:    caminos.Local,
		AD:       caminos.AD,
		Keycloak: caminos.Keycloak,
	})
}

// Login entra con correo y contraseña.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	sesion, err := c.service.Login(entrada.Email, entrada.Password)
	if err != nil {
		// El intento se registra —cuenta, resultado, IP y fecha— y **nunca la contraseña**, ni
		// siquiera la equivocada (docs/modules/auth.md, decisión 14).
		logs.LogWarning("intento de entrada fallido (" + describeForLog(err) + ") desde " + ClientIP(r))
		c.fail(w, r, err)
		return
	}

	logs.LogSuccess("entrada correcta " + labelOf(sesion))
	httpx.WriteJSON(w, http.StatusOK, dtos.NewLoginResponse(sesion.Token, sesion.ExpiresAt, identityOf(sesion)))
}

// Me dice quién soy, que es lo que el frontend pregunta al arrancar para saber si su sesión sigue
// viva y con qué papel.
func (c *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	identity := auth.MustFromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, dtos.MeResponse{User: dtos.NewUserResponse(identity)})
}

// Logout no invalida nada: no hay tabla de sesiones. El frontend borra su token y aquí queda
// constancia en el log.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	c.service.Logout(auth.MustFromContext(r.Context()))
	httpx.WriteJSON(w, http.StatusOK, dtos.Ok())
}

// Forgot pide el enlace de recuperación. **Responde igual exista o no la cuenta.**
func (c *AuthController) Forgot(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.ForgotRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	if err := c.service.ForgotPassword(entrada.Email); err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.Ok())
}

// Reset establece la contraseña con el token del enlace, sin saber la anterior.
func (c *AuthController) Reset(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.ResetRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	if err := c.service.ResetPassword(entrada.Token, entrada.Password, ClientIP(r)); err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.Ok())
}

// Change cambia la contraseña propia, con la actual.
func (c *AuthController) Change(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.ChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	identity := auth.MustFromContext(r.Context())
	if err := c.service.ChangePassword(identity, entrada.CurrentPassword, entrada.Password, ClientIP(r)); err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.Ok())
}

// KeycloakStart manda al navegador a la pantalla de entrada de Keycloak.
//
// **Es una redirección y no una respuesta con datos**, a propósito: quien entra por Keycloak se
// autentica en Keycloak, y lo que hace falta es que el navegador vaya allí
// (docs/modules/auth.md, sección 5.3).
func (c *AuthController) KeycloakStart(w http.ResponseWriter, r *http.Request) {
	direccion, err := c.service.StartKeycloak()
	if err != nil {
		// Sin Keycloak configurado el camino no existe, y esto sólo se alcanza escribiendo la
		// dirección a mano: se dice que no hay camino, no que algo haya fallado.
		c.fail(w, r, err)
		return
	}

	http.Redirect(w, r, direccion, http.StatusFound)
}

// KeycloakCallback recibe la vuelta de Keycloak y deja a la persona en la pantalla de entrada, con su
// token o con lo que ha pasado en el fragmento de la dirección.
//
// **El fragmento no viaja al servidor**: no queda en los registros de nginx, y el frontend lo borra
// en cuanto lo usa (docs/modules/auth.md, decisiones 13 y 30).
func (c *AuthController) KeycloakCallback(w http.ResponseWriter, r *http.Request) {
	parametros := r.URL.Query()

	// Keycloak cuenta por aquí lo que ha pasado cuando no hay código: alguien que cancela, un cliente
	// mal configurado… Se registra el motivo —que no es un dato de nadie— y se sigue: la pantalla de
	// entrada dirá que la entrada no se ha completado.
	if fallo := parametros.Get("error"); fallo != "" {
		logs.LogWarning("Keycloak no ha devuelto código: " + fallo)
	}

	http.Redirect(w, r, c.service.KeycloakCallback(parametros.Get("state"), parametros.Get("code")), http.StatusFound)
}

// fail traduce el error del servicio a la clave y el código que le tocan
// (docs/modules/auth.md, sección 10).
func (c *AuthController) fail(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, services.ErrSinDireccionPublica):
		// **Sin dirección pública no hay enlace que mandar** (decisión 14): es un dato que falta
		// para poder hacerlo, así que 422 con su clave.
		httpx.WriteError(w, http.StatusUnprocessableEntity, "auth.publicUrl.missing")
	case errors.Is(err, services.ErrOIDCNotConfigured):
		httpx.WriteError(w, http.StatusNotFound, "auth.oidc.notConfigured")
	case errors.Is(err, services.ErrOIDCUnavailable):
		httpx.WriteError(w, http.StatusServiceUnavailable, "auth.oidc.unavailable")
	case errors.Is(err, services.ErrDirectoryUnavailable):
		// El directorio no responde: es un fallo del servidor y no de quien escribe, y se dice así para
		// que nadie se ponga a buscar una contraseña que no ha perdido
		// (docs/modules/auth.md, sección 5.2).
		httpx.WriteError(w, http.StatusServiceUnavailable, "auth.directory.unavailable")
	case errors.Is(err, services.ErrDirectoryRejected):
		httpx.WriteError(w, http.StatusUnauthorized, "auth.directory.rejected")
	case errors.Is(err, services.ErrInvalidCredentials):
		httpx.WriteError(w, http.StatusUnauthorized, "auth.invalidCredentials")
	case errors.Is(err, services.ErrAccountInactive):
		httpx.WriteError(w, http.StatusUnauthorized, "auth.accountInactive")
	case errors.Is(err, services.ErrTokenExpired):
		httpx.WriteError(w, http.StatusUnauthorized, "auth.token.expired")
	case errors.Is(err, services.ErrTokenUsed):
		httpx.WriteError(w, http.StatusUnauthorized, "auth.token.used")
	case errors.Is(err, services.ErrPasswordTooShort):
		httpx.WriteError(w, http.StatusUnauthorized, "auth.password.tooShort")
	case errors.Is(err, services.ErrPasswordWrong):
		httpx.WriteError(w, http.StatusUnauthorized, "auth.password.wrong")
	case errors.Is(err, services.ErrPasswordNotLocal):
		httpx.WriteError(w, http.StatusUnauthorized, "auth.password.notLocal")
	default:
		// Ni un detalle de dentro: al log y 500.
		logs.LogError("error inesperado en auth: " + err.Error())
		httpx.WriteError(w, http.StatusInternalServerError, KeyInternal)
	}
}

// ClientIP es la dirección de quien llama.
//
// Detrás de nginx, `RemoteAddr` es la del proxy, así que se usa la cabecera que nginx pone con la
// dirección real (docs/modules/auth.md, sección 7). Se coge la primera de `X-Forwarded-For` y no la
// última: la pone el proxy de delante, que es el único en el que se confía.
func ClientIP(r *http.Request) string {
	if real := strings.TrimSpace(r.Header.Get("X-Real-IP")); real != "" {
		return real
	}

	if reenviada := r.Header.Get("X-Forwarded-For"); reenviada != "" {
		if primera, _, hay := strings.Cut(reenviada, ","); hay {
			return strings.TrimSpace(primera)
		}
		return strings.TrimSpace(reenviada)
	}

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}

// identityOf traduce la sesión a la identidad que se enseña.
func identityOf(sesion services.Session) auth.Identity {
	if sesion.Factory {
		return auth.FactoryIdentity()
	}
	return sesion.Account.Identity()
}

// labelOf describe la cuenta para el log **sin datos personales**: ni el nombre ni el correo, sólo
// el identificador.
func labelOf(sesion services.Session) string {
	if sesion.Factory {
		return "de la cuenta de fábrica"
	}
	return "de la cuenta #" + strconv.FormatInt(sesion.Account.ID, 10)
}

// describeForLog da lo que se escribe en el log cuando falla una entrada: la clave del error, y el
// detalle si lo hubiera. La contraseña no aparece por ningún lado, ni la correcta ni la equivocada.
func describeForLog(err error) string { return err.Error() }
