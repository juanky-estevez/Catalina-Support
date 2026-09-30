package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/config"
)

// Keycloak es el camino de OIDC: entrar con la cuenta de la organización a través de Keycloak.
//
// La declaración vive en el consumidor —aquí— y la cumple quien habla con Keycloak, como con el
// directorio: así el servicio se puede probar con un doble y nadie importa a nadie
// (docs/arquitectura.md, sección 4).
type Keycloak interface {
	// Configured dice si hay Keycloak al que ir.
	Configured() bool
	// AuthURL arma la dirección a la que se manda al navegador, con su `state`.
	AuthURL(state string) (string, error)
	// Exchange canjea el código de la vuelta y devuelve la persona, ya como cuenta de directorio.
	Exchange(code string) (auth.DirectoryAccount, error)
}

// Los dos fallos del camino de OIDC. Se separan porque se contestan distinto: uno es «no se pudo
// completar la entrada» y el otro, «Keycloak no responde», que es del servidor
// (docs/modules/auth.md, sección 10).
var (
	ErrOIDCRejected    = errors.New("keycloak no ha aceptado la vuelta")
	ErrOIDCUnavailable = errors.New("keycloak no responde")
	// ErrOIDCNoEmail: la cuenta del directorio no trae correo, y sin correo no hay cuenta aquí: el
	// correo es con lo que se entra y con lo que se identifican los tickets.
	ErrOIDCNoEmail = errors.New("la cuenta de keycloak no tiene correo")
	// ErrOIDCNotConfigured: esta instalación no tiene Keycloak. La pantalla no ofrece el botón, así
	// que llegar aquí es escribir la dirección a mano.
	ErrOIDCNotConfigured = errors.New("esta instalación no tiene keycloak")
)

// tiempoDeEspera es lo que se espera a Keycloak. Generoso para el canje —va con la vuelta de una
// persona de por medio— pero acotado: sin límite, un Keycloak colgado deja la petición colgada.
const tiempoDeEspera = 10 * time.Second

// OIDC habla con Keycloak con lo que trae la biblioteca estándar: descubrimiento, canje del código y
// lectura de la persona. **No se añade ninguna dependencia** para esto.
type OIDC struct {
	cfg     config.OIDC
	cliente *http.Client

	// El documento de descubrimiento se pide una vez y se recuerda: cambia cuando cambia el reino, y
	// pedirlo en cada entrada sería una ida y vuelta de red de más en el camino de entrar.
	mu             sync.Mutex
	descubrimiento *descubrimiento
}

// NewOIDC construye el camino de Keycloak.
func NewOIDC(cfg config.OIDC) *OIDC {
	return &OIDC{
		cfg:     cfg,
		cliente: &http.Client{Timeout: tiempoDeEspera},
	}
}

// Configured dice si hay Keycloak al que ir.
func (o *OIDC) Configured() bool { return o.cfg.Configured() }

// Probe prueba la configuración **sin guardarla**: lee el documento del reino y comprueba que dice
// dónde está su pantalla de entrada, que es por donde empieza a entrarse. Es lo que hace el botón
// «Probar la conexión» de Configuración.
//
// No se pide ningún token ni se entra con nadie: lo que se prueba es que el reino está donde dice la
// configuración y que la dice.
func (o *OIDC) Probe() error {
	if !o.cfg.Configured() {
		return ErrOIDCNotConfigured
	}

	documento, err := o.documento()
	if err != nil {
		return err
	}
	if documento.AuthorizationEndpoint == "" {
		return fmt.Errorf("%w: el reino no dice dónde está su pantalla de entrada", ErrOIDCUnavailable)
	}

	return nil
}

// descubrimiento es lo que se lee del documento de descubrimiento del reino, reducido a lo que se
// usa. Se leen tres direcciones y nada más: ni el emisor ni las capacidades, que aquí no deciden
// nada.
type descubrimiento struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
}

// AuthURL arma la dirección de la pantalla de entrada de Keycloak.
func (o *OIDC) AuthURL(state string) (string, error) {
	documento, err := o.documento()
	if err != nil {
		return "", err
	}
	if documento.AuthorizationEndpoint == "" {
		return "", fmt.Errorf("%w: el reino no dice dónde está su pantalla de entrada", ErrOIDCUnavailable)
	}

	parametros := url.Values{
		"response_type": {"code"},
		"client_id":     {o.cfg.ClientID},
		"redirect_uri":  {o.cfg.RedirectURI},
		// `openid` es lo que hace que sea OIDC; el correo y el nombre son lo que se lee después.
		"scope": {"openid email profile"},
		"state": {state},
	}

	direccion, err := url.Parse(documento.AuthorizationEndpoint)
	if err != nil {
		return "", fmt.Errorf("%w: la dirección de entrada del reino no se entiende", ErrOIDCUnavailable)
	}
	direccion.RawQuery = parametros.Encode()

	return direccion.String(), nil
}

// Exchange canjea el código por un token y lee quién es la persona.
func (o *OIDC) Exchange(code string) (auth.DirectoryAccount, error) {
	if strings.TrimSpace(code) == "" {
		// Sin código no hay nada que canjear: es lo que llega cuando alguien cancela en la pantalla de
		// Keycloak, y se cuenta como una vuelta que no se ha completado.
		return auth.DirectoryAccount{}, ErrOIDCRejected
	}

	documento, err := o.documento()
	if err != nil {
		return auth.DirectoryAccount{}, err
	}

	token, err := o.pedirToken(documento.TokenEndpoint, code)
	if err != nil {
		return auth.DirectoryAccount{}, err
	}

	return o.leerPersona(documento.UserInfoEndpoint, token)
}

// pedirToken hace el canje del código.
func (o *OIDC) pedirToken(endpoint, code string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("%w: el reino no dice dónde canjear el código", ErrOIDCUnavailable)
	}

	cuerpo := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {o.cfg.RedirectURI},
		"client_id":     {o.cfg.ClientID},
		"client_secret": {o.cfg.ClientSecret},
	}

	peticion, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(cuerpo.Encode()))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOIDCUnavailable, err)
	}
	peticion.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	peticion.Header.Set("Accept", "application/json")

	respuesta, err := o.cliente.Do(peticion)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOIDCUnavailable, err)
	}
	defer respuesta.Body.Close()

	datos, err := io.ReadAll(io.LimitReader(respuesta.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOIDCUnavailable, err)
	}

	if respuesta.StatusCode != http.StatusOK {
		// Un 4xx aquí es «ese código no vale» —caducado, ya usado o de otra instalación—, y no un
		// fallo de Keycloak. Se distingue para poder decirlo.
		if respuesta.StatusCode >= 400 && respuesta.StatusCode < 500 {
			return "", fmt.Errorf("%w: %s", ErrOIDCRejected, claveDeError(datos))
		}
		return "", fmt.Errorf("%w: el canje contestó %d", ErrOIDCUnavailable, respuesta.StatusCode)
	}

	var canje struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(datos, &canje); err != nil {
		return "", fmt.Errorf("%w: el canje no devolvió lo que se esperaba", ErrOIDCUnavailable)
	}
	if canje.AccessToken == "" {
		return "", fmt.Errorf("%w: el canje no trajo token de acceso", ErrOIDCUnavailable)
	}

	return canje.AccessToken, nil
}

// leerPersona pregunta a Keycloak quién es quien acaba de entrar.
func (o *OIDC) leerPersona(endpoint, token string) (auth.DirectoryAccount, error) {
	if endpoint == "" {
		return auth.DirectoryAccount{}, fmt.Errorf("%w: el reino no dice dónde preguntar quién es", ErrOIDCUnavailable)
	}

	peticion, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return auth.DirectoryAccount{}, fmt.Errorf("%w: %v", ErrOIDCUnavailable, err)
	}
	peticion.Header.Set("Authorization", "Bearer "+token)
	peticion.Header.Set("Accept", "application/json")

	respuesta, err := o.cliente.Do(peticion)
	if err != nil {
		return auth.DirectoryAccount{}, fmt.Errorf("%w: %v", ErrOIDCUnavailable, err)
	}
	defer respuesta.Body.Close()

	datos, err := io.ReadAll(io.LimitReader(respuesta.Body, 1<<20))
	if err != nil {
		return auth.DirectoryAccount{}, fmt.Errorf("%w: %v", ErrOIDCUnavailable, err)
	}
	if respuesta.StatusCode != http.StatusOK {
		if respuesta.StatusCode >= 400 && respuesta.StatusCode < 500 {
			return auth.DirectoryAccount{}, fmt.Errorf("%w: no se pudo leer quién es", ErrOIDCRejected)
		}
		return auth.DirectoryAccount{}, fmt.Errorf("%w: la lectura contestó %d", ErrOIDCUnavailable, respuesta.StatusCode)
	}

	return personaDeKeycloak(datos)
}

// personaDeKeycloak traduce lo que devuelve Keycloak a la persona que entiende el resto del sistema.
//
// Las claves son las estándar de OIDC: `sub`, `email`, `given_name` y `family_name`. **El `sub` manda
// sobre el correo**, igual que el identificador externo en AD: es lo que hace que un cambio de correo
// no convierta a nadie en otra persona (docs/modules/auth.md, sección 5.4).
func personaDeKeycloak(datos []byte) (auth.DirectoryAccount, error) {
	var persona struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Name          string `json:"name"`
		PreferredUser string `json:"preferred_username"`
	}
	if err := json.Unmarshal(datos, &persona); err != nil {
		return auth.DirectoryAccount{}, fmt.Errorf("%w: no se entiende quién es", ErrOIDCUnavailable)
	}

	correo := strings.TrimSpace(persona.Email)
	if correo == "" {
		return auth.DirectoryAccount{}, ErrOIDCNoEmail
	}

	nombre := strings.TrimSpace(persona.GivenName)
	if nombre == "" {
		nombre = strings.TrimSpace(persona.Name)
	}
	if nombre == "" {
		// Un reino que no rellena el nombre deja al menos su nombre de usuario, que es mejor que una
		// cuenta sin nombre con la que nadie sabe con quién habla.
		nombre = strings.TrimSpace(persona.PreferredUser)
	}

	return auth.DirectoryAccount{
		Origin:     auth.OriginKeycloak,
		ExternalID: strings.TrimSpace(persona.Sub),
		Email:      correo,
		Name:       nombre,
		LastName:   strings.TrimSpace(persona.FamilyName),
	}, nil
}

// documento pide el documento de descubrimiento del reino, y lo recuerda.
func (o *OIDC) documento() (*descubrimiento, error) {
	o.mu.Lock()
	if o.descubrimiento != nil {
		documento := o.descubrimiento
		o.mu.Unlock()
		return documento, nil
	}
	o.mu.Unlock()

	direccion := strings.TrimRight(o.cfg.Issuer, "/") + "/.well-known/openid-configuration"

	peticion, err := http.NewRequest(http.MethodGet, direccion, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOIDCUnavailable, err)
	}
	peticion.Header.Set("Accept", "application/json")

	respuesta, err := o.cliente.Do(peticion)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOIDCUnavailable, err)
	}
	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: el descubrimiento del reino contestó %d", ErrOIDCUnavailable, respuesta.StatusCode)
	}

	var documento descubrimiento
	if err := json.NewDecoder(io.LimitReader(respuesta.Body, 1<<20)).Decode(&documento); err != nil {
		return nil, fmt.Errorf("%w: no se entiende el descubrimiento del reino", ErrOIDCUnavailable)
	}

	o.mu.Lock()
	o.descubrimiento = &documento
	o.mu.Unlock()

	return &documento, nil
}

// claveDeError saca la clave del error que devuelve Keycloak cuando no acepta el canje, para el log.
// **Nunca se registra el código ni el token**, sólo la clave y la descripción corta.
func claveDeError(datos []byte) string {
	var fallo struct {
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.Unmarshal(datos, &fallo); err != nil || fallo.Error == "" {
		return "sin detalle"
	}
	if fallo.Description == "" {
		return fallo.Error
	}

	return fallo.Error + ": " + fallo.Description
}
