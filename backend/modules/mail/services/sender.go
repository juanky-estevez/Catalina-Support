package services

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

// ErrNotConfigured se devuelve cuando falta la configuración de SMTP. Es un error propio y no un
// fallo genérico: quien lo lea tiene que saber que el problema no es el correo, sino la
// configuración.
var ErrNotConfigured = errors.New("mail.smtp.notConfigured")

// Sender envía correos por SMTP, con la biblioteca estándar y sin dependencias.
type Sender struct {
	host      string
	port      string
	user      string
	password  string
	fromName  string
	fromEmail string
	secure    bool
	timeout   time.Duration
	// instalacion da el correo saliente: **vive en la configuración**, como el directorio y Keycloak.
	// Ya no hay respaldo en el entorno (decisión del responsable, 2026-09-30).
	instalacion Instalacion
}

// NewSender construye el remitente a partir de la configuración del entorno.
func NewSender() *Sender {
	return &Sender{timeout: 15 * time.Second}
}

// SetInstalacion fija de dónde sale el correo saliente cuando la instalación tiene uno puesto. Lo
// llama el servicio, y el respaldo es lo del entorno.
func (s *Sender) SetInstalacion(instalacion Instalacion) { s.instalacion = instalacion }

// datos es el correo saliente de este envío: **el de la configuración**, que es donde vive. Sin él no
// hay con qué enviar, y el envío falla con su error claro en lugar de intentar una conexión a ninguna
// parte (docs/primer-arranque.md, sección 5).
func (s *Sender) datos() CorreoSaliente {
	if s.instalacion == nil {
		return CorreoSaliente{}
	}

	correo, _ := s.instalacion.SMTP()

	return correo
}

// Configured dice si hay con qué enviar. Sin esto, el envío falla con un error claro en lugar de
// intentar una conexión a ninguna parte.
func (s *Sender) Configured() bool {
	correo := s.datos()

	return correo.Host != "" && correo.Port != "" && correo.FromEmail != ""
}

// Send arma el mensaje y lo entrega.
//
// El tiempo límite es del código: un servidor de correo que no responde no puede dejar colgada una
// petición ni una tarea de fondo para siempre.
func (s *Sender) Send(to []string, subject, htmlBody string) error {
	if !s.Configured() {
		return ErrNotConfigured
	}
	if len(to) == 0 {
		return errors.New("mail.to.empty")
	}

	correo := s.datos()

	message, err := Build(correo.FromName, correo.FromEmail, to, subject, htmlBody)
	if err != nil {
		return err
	}

	address := net.JoinHostPort(correo.Host, correo.Port)

	// Conexión cifrada desde el principio (el puerto 465 de siempre).
	if correo.Secure {
		return s.sendSecure(to, message, correo)
	}

	// Sin cifrado explícito se usa el camino de la biblioteca estándar, que negocia STARTTLS si el
	// servidor lo ofrece.
	var auth smtp.Auth
	if correo.User != "" {
		auth = smtp.PlainAuth("", correo.User, correo.Password, correo.Host)
	}

	return smtp.SendMail(address, auth, correo.FromEmail, to, message)
}

// Probar abre la conexión y **valida la autenticación, sin mandar ningún correo**: no hay MAIL FROM,
// ni RCPT TO, ni DATA. Es lo que necesita el asistente, donde no hay destinatario y sólo se quiere
// saber si el servidor contesta y si las credenciales valen (docs/primer-arranque.md, sección 3).
//
// Usa **la misma conexión y la misma autenticación que el envío** —`conectar`—, así que lo que se
// prueba es exactamente lo que se va a usar al mandar.
func (s *Sender) Probar(correo CorreoSaliente) error {
	if correo.Host == "" || correo.Port == "" {
		return ErrNotConfigured
	}

	cliente, err := s.conectar(correo)
	if err != nil {
		return err
	}
	defer cliente.Close()

	return cliente.Quit()
}

// conectar deja un cliente SMTP **ya conectado y autenticado**, listo para entregar un mensaje o para
// comprobar la conexión. El cifrado directo (puerto 465) y STARTTLS (587) se resuelven aquí, con el
// mismo tiempo límite: un servidor que no responde no puede dejar colgada una petición.
func (s *Sender) conectar(correo CorreoSaliente) (*smtp.Client, error) {
	address := net.JoinHostPort(correo.Host, correo.Port)

	espera := s.timeout
	if espera == 0 {
		espera = 15 * time.Second
	}

	var cliente *smtp.Client

	if correo.Secure {
		conexion, err := tls.DialWithDialer(
			&net.Dialer{Timeout: espera},
			"tcp",
			address,
			&tls.Config{ServerName: correo.Host, MinVersion: tls.VersionTLS12},
		)
		if err != nil {
			return nil, fmt.Errorf("mail.smtp.connect: %w", err)
		}

		cliente, err = smtp.NewClient(conexion, correo.Host)
		if err != nil {
			conexion.Close()
			return nil, fmt.Errorf("mail.smtp.handshake: %w", err)
		}
	} else {
		conexion, err := net.DialTimeout("tcp", address, espera)
		if err != nil {
			return nil, fmt.Errorf("mail.smtp.connect: %w", err)
		}

		cliente, err = smtp.NewClient(conexion, correo.Host)
		if err != nil {
			conexion.Close()
			return nil, fmt.Errorf("mail.smtp.handshake: %w", err)
		}

		// Si el servidor ofrece STARTTLS, se sube la conexión antes de mandar credenciales.
		if soportado, _ := cliente.Extension("STARTTLS"); soportado {
			if err := cliente.StartTLS(&tls.Config{ServerName: correo.Host, MinVersion: tls.VersionTLS12}); err != nil {
				cliente.Close()
				return nil, fmt.Errorf("mail.smtp.handshake: %w", err)
			}
		}
	}

	if correo.User != "" {
		if err := cliente.Auth(smtp.PlainAuth("", correo.User, correo.Password, correo.Host)); err != nil {
			cliente.Close()
			return nil, fmt.Errorf("mail.smtp.auth: %w", err)
		}
	}

	return cliente, nil
}

func (s *Sender) sendSecure(to []string, message []byte, correo CorreoSaliente) error {
	cliente, err := s.conectar(correo)
	if err != nil {
		return err
	}
	defer cliente.Close()

	if err := cliente.Mail(correo.FromEmail); err != nil {
		return fmt.Errorf("mail.smtp.from: %w", err)
	}
	for _, recipient := range to {
		if err := cliente.Rcpt(recipient); err != nil {
			return fmt.Errorf("mail.smtp.recipient: %w", err)
		}
	}

	writer, err := cliente.Data()
	if err != nil {
		return fmt.Errorf("mail.smtp.data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		return fmt.Errorf("mail.smtp.write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("mail.smtp.close: %w", err)
	}

	return cliente.Quit()
}
