package services

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"catalina-support/backend/shared/config"
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
}

// NewSender construye el remitente a partir de la configuración del entorno.
func NewSender(cfg config.Config) *Sender {
	return &Sender{
		host:      cfg.SMTPHost,
		port:      cfg.SMTPPort,
		user:      cfg.SMTPUser,
		password:  cfg.SMTPPassword,
		fromName:  cfg.SMTPFromName,
		fromEmail: cfg.SMTPFromEmail,
		secure:    cfg.SMTPSecure,
		timeout:   15 * time.Second,
	}
}

// Configured dice si hay con qué enviar. Sin esto, el envío falla con un error claro en lugar de
// intentar una conexión a ninguna parte.
func (s *Sender) Configured() bool {
	return s.host != "" && s.port != "" && s.fromEmail != ""
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

	message, err := Build(s.fromName, s.fromEmail, to, subject, htmlBody)
	if err != nil {
		return err
	}

	address := net.JoinHostPort(s.host, s.port)

	// Conexión cifrada desde el principio (el puerto 465 de siempre).
	if s.secure {
		return s.sendSecure(address, to, message)
	}

	// Sin cifrado explícito se usa el camino de la biblioteca estándar, que negocia STARTTLS si el
	// servidor lo ofrece.
	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	return smtp.SendMail(address, auth, s.fromEmail, to, message)
}

func (s *Sender) sendSecure(address string, to []string, message []byte) error {
	connection, err := tls.DialWithDialer(
		&net.Dialer{Timeout: s.timeout},
		"tcp",
		address,
		&tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12},
	)
	if err != nil {
		return fmt.Errorf("mail.smtp.connect: %w", err)
	}

	client, err := smtp.NewClient(connection, s.host)
	if err != nil {
		return fmt.Errorf("mail.smtp.handshake: %w", err)
	}
	defer client.Close()

	if s.user != "" {
		if err := client.Auth(smtp.PlainAuth("", s.user, s.password, s.host)); err != nil {
			return fmt.Errorf("mail.smtp.auth: %w", err)
		}
	}

	if err := client.Mail(s.fromEmail); err != nil {
		return fmt.Errorf("mail.smtp.from: %w", err)
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("mail.smtp.recipient: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("mail.smtp.data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		return fmt.Errorf("mail.smtp.write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("mail.smtp.close: %w", err)
	}

	return client.Quit()
}
