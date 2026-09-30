package services

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"

	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/config"
)

// Los dos fallos que puede tener preguntarle al directorio, y que **no son lo mismo que una
// contraseña equivocada** (docs/modules/auth.md, secciones 5.2 y 10).
var (
	// ErrDirectoryUnavailable: el directorio no responde. Es un 503, porque el fallo es del servidor y
	// no de quien escribe.
	ErrDirectoryUnavailable = errors.New("auth.directory.unavailable")
	// ErrDirectoryRejected: el directorio ha rechazado esas credenciales.
	ErrDirectoryRejected = errors.New("auth.directory.rejected")
)

// Directorio es el camino de AD: validar a una persona contra el directorio de la organización.
//
// Es una interfaz para poder probar el servicio de `auth` sin un directorio de verdad, y la cumple
// `LDAP`.
type Directorio interface {
	// Login busca a la persona y **valida su contraseña contra el directorio**. Si no está o la
	// contraseña no vale, devuelve `ErrDirectoryRejected`.
	Login(email, password string) (auth.DirectoryAccount, error)
	// Knows dice si el directorio sigue conociendo a alguien con ese correo. Lo usa `users` para
	// reactivar una cuenta de AD, y se pregunta igual que se entra: la misma cuenta de servicio y el
	// mismo filtro (docs/modules/users.md, sección 5, punto 4).
	Knows(email string) (bool, error)
}

// LDAP habla con el directorio por LDAP.
//
// **Se valida con una cuenta de servicio**, y no construyendo el nombre completo a mano: en un
// directorio real las unidades organizativas no siguen un patrón predecible, y armar el DN falla el
// día que alguien cambia de departamento (docs/modules/auth.md, sección 5.2).
type LDAP struct {
	cfg config.Directory
	// timeout acota lo que se espera al directorio: sin él, un directorio que no responde deja a
	// quien intenta entrar mirando una pantalla que no dice nada.
	timeout time.Duration
}

// NewLDAP construye el cliente del directorio.
func NewLDAP(cfg config.Directory) *LDAP {
	return &LDAP{cfg: cfg, timeout: 8 * time.Second}
}

// Login busca a la persona por su correo, valida su contraseña y devuelve sus datos.
func (l *LDAP) Login(email, password string) (auth.DirectoryAccount, error) {
	if !l.cfg.Configured() {
		return auth.DirectoryAccount{}, ErrDirectoryUnavailable
	}

	conexion, err := l.conectar()
	if err != nil {
		return auth.DirectoryAccount{}, err
	}
	defer conexion.Close()

	// 1. Se busca a la persona **con la cuenta de servicio**.
	entrada, err := l.buscar(conexion, email)
	if err != nil {
		return auth.DirectoryAccount{}, err
	}
	if entrada == nil {
		// No está en el directorio: para quien lo intenta, es lo mismo que una contraseña equivocada.
		return auth.DirectoryAccount{}, ErrDirectoryRejected
	}

	// 2. Se conecta **con sus credenciales**. Ese intento es la validación: si el directorio lo
	// acepta, es quien dice ser.
	if err := l.validar(entrada.DN, password); err != nil {
		return auth.DirectoryAccount{}, err
	}

	datos := auth.DirectoryAccount{
		Origin:     auth.OriginAD,
		ExternalID: primerValor(entrada, l.cfg.AttrID),
		Email:      primerValor(entrada, l.cfg.AttrEmail),
		Name:       primerValor(entrada, l.cfg.AttrName),
		LastName:   primerValor(entrada, l.cfg.AttrLastName),
	}

	// Sin correo no hay cuenta: el correo es el identificador de las personas en este producto.
	if strings.TrimSpace(datos.Email) == "" {
		datos.Email = email
	}
	// Y sin identificador del directorio tampoco: es lo que manda cuando alguien cambia de correo. Se
	// usa el DN, que en un directorio es único.
	if strings.TrimSpace(datos.ExternalID) == "" {
		datos.ExternalID = entrada.DN
	}

	return datos, nil
}

// Knows dice si el directorio sigue conociendo a alguien con ese correo.
//
// Es lo que necesita `users` para reactivar una cuenta de directorio: **devolver el acceso a quien ya
// no está allí sería devolver un acceso que no puede usar**, porque su contraseña es de allí y sin
// aparecer en el directorio no podría entrar (docs/modules/users.md, sección 5, punto 4).
//
// Se busca con **la misma cuenta de servicio y el mismo filtro** con los que se entra, a propósito:
// la pregunta «¿sigue estando?» tiene que contestarse igual que se contesta «¿puede entrar?».
func (l *LDAP) Knows(email string) (bool, error) {
	if !l.cfg.Configured() {
		return false, ErrDirectoryUnavailable
	}

	conexion, err := l.conectar()
	if err != nil {
		return false, err
	}
	defer conexion.Close()

	entrada, err := l.buscar(conexion, email)
	if err != nil {
		return false, err
	}

	return entrada != nil, nil
}

// Probe prueba la configuración **sin guardarla y sin saber nada de nadie**: abre la conexión con la
// cuenta de servicio y la cierra. Es lo que hace el botón «Probar la conexión» de Configuración, y
// contesta lo que hay que saber antes de guardar: que el directorio está donde dice y que la cuenta de
// servicio entra.
func (l *LDAP) Probe() error {
	if !l.cfg.Configured() {
		return ErrDirectoryUnavailable
	}

	conexion, err := l.conectar()
	if err != nil {
		return err
	}

	return conexion.Close()
}

// conectar abre la conexión con la cuenta de servicio.
func (l *LDAP) conectar() (*ldap.Conn, error) {
	direccion := l.cfg.Host + ":" + l.cfg.Port

	var (
		conexion *ldap.Conn
		err      error
	)

	if l.cfg.UseTLS {
		conexion, err = ldap.DialURL("ldaps://"+direccion, ldap.DialWithTLSConfig(&tls.Config{
			ServerName: l.cfg.Host,
			MinVersion: tls.VersionTLS12,
		}))
	} else {
		conexion, err = ldap.DialURL("ldap://" + direccion)
	}
	if err != nil {
		return nil, ErrDirectoryUnavailable
	}

	conexion.SetTimeout(l.timeout)

	if l.cfg.BindDN != "" {
		if err := conexion.Bind(l.cfg.BindDN, l.cfg.BindPass); err != nil {
			// La cuenta de servicio no vale: es un fallo de configuración, y quien lo ve es un
			// administrador. Se cuenta como «el directorio no responde» para no decirle a quien
			// intenta entrar algo que no puede arreglar.
			conexion.Close()
			return nil, ErrDirectoryUnavailable
		}
	}

	return conexion, nil
}

// buscar encuentra a la persona por su correo con el filtro configurado.
func (l *LDAP) buscar(conexion *ldap.Conn, email string) (*ldap.Entry, error) {
	filtro := fmt.Sprintf(l.cfg.UserFilter, ldap.EscapeFilter(email))

	peticion := ldap.NewSearchRequest(
		l.cfg.SearchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		// Un límite de uno: dos cuentas con el mismo correo en el directorio es un problema de allí, y
		// aquí no se puede decidir cuál es.
		1,
		int(l.timeout.Seconds()),
		false,
		filtro,
		[]string{l.cfg.AttrEmail, l.cfg.AttrName, l.cfg.AttrLastName, l.cfg.AttrID},
		nil,
	)

	resultado, err := conexion.Search(peticion)
	if err != nil {
		return nil, ErrDirectoryUnavailable
	}
	if len(resultado.Entries) == 0 {
		return nil, nil
	}

	return resultado.Entries[0], nil
}

// validar prueba las credenciales de la persona contra el directorio.
func (l *LDAP) validar(dn, password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrDirectoryRejected
	}

	conexion, err := l.conectar()
	if err != nil {
		return err
	}
	defer conexion.Close()

	if err := conexion.Bind(dn, password); err != nil {
		// El directorio no lo acepta: es lo mismo que una contraseña equivocada.
		return ErrDirectoryRejected
	}

	return nil
}

// primerValor devuelve el primer valor de ese atributo, o nada.
func primerValor(entrada *ldap.Entry, atributo string) string {
	if atributo == "" {
		return ""
	}

	return strings.TrimSpace(entrada.GetAttributeValue(atributo))
}
