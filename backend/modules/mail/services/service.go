package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/mail/repositories"
)

// Service es lo que el resto del sistema usa para mandar correo.
//
// Nadie de fuera toca las plantillas ni el remitente: se pide un envío con una clave, un idioma,
// unos destinatarios y unos datos. Quién recibe el correo lo decide quien pide, no este módulo
// (docs/modules/mail.md, sección 2).
// Instalacion es lo que este módulo necesita saber de la configuración de la instalación: **la
// dirección pública**, para los enlaces de la vista previa, y **la zona horaria**, para escribir las
// fechas. Lo declara quien lo usa y lo cumple el módulo `settings`
// (docs/modules/settings.md, decisiones 14 y 15).
type Instalacion interface {
	DireccionPublica() string
	ZonaHoraria() string
	// SMTP es **el correo saliente configurado en la instalación**. El segundo valor dice si hay
	// alguno puesto: en falso, el remitente usa el del entorno, que es como funcionaba antes
	// (docs/primer-arranque.md, sección 5).
	SMTP() (CorreoSaliente, bool)
}

// CorreoSaliente es el correo saliente tal y como lo necesita el remitente. Lo declara este
// módulo y lo cumple el módulo de configuración, traducido por el cableado.
type CorreoSaliente struct {
	Host      string
	Port      string
	Secure    bool
	User      string
	Password  string
	FromName  string
	FromEmail string
}

type Service struct {
	repo         *repositories.TemplateRepository
	sender       *Sender
	publicAppURL string
	// instalacion es de dónde salen la dirección y la zona de verdad; lo que hay en `publicAppURL`
	// es el respaldo del entorno, para una instalación que ya lo tuviera puesto.
	instalacion Instalacion
}

// NewService construye el servicio.
func NewService(repo *repositories.TemplateRepository, sender *Sender, publicAppURL string) *Service {
	return &Service{repo: repo, sender: sender, publicAppURL: publicAppURL}
}

// SetInstalacion fija de dónde salen la dirección pública y la zona horaria. Lo llama el cableado,
// con el módulo de configuración: así este módulo no sabe quién se las da (docs/arquitectura.md, 4).
func (s *Service) SetInstalacion(instalacion Instalacion) {
	s.instalacion = instalacion
	// **El remitente resuelve en cada envío**, así que se le da el mismo proveedor: cambiarlo en la
	// pantalla vale sin reiniciar nada (docs/primer-arranque.md, sección 5).
	if s.sender != nil {
		s.sender.SetInstalacion(instalacion)
	}
}

// baseDeLosEnlaces es la dirección pública: **la de la configuración** y, si no hay, la del entorno.
func (s *Service) baseDeLosEnlaces() string {
	if s.instalacion != nil {
		if base := strings.TrimRight(strings.TrimSpace(s.instalacion.DireccionPublica()), "/"); base != "" {
			return base
		}
	}

	return strings.TrimRight(strings.TrimSpace(s.publicAppURL), "/")
}

// zonaHoraria es la de la instalación, o **UTC** si no se puede leer: la vista previa nunca se queda
// sin fecha.
func (s *Service) zonaHoraria() *time.Location {
	if s.instalacion != nil {
		if zona, err := time.LoadLocation(s.instalacion.ZonaHoraria()); err == nil {
			return zona
		}
	}

	return time.UTC
}

// Templates devuelve las veinte plantillas, para el editor.
func (s *Service) Templates() ([]repositories.Template, error) {
	return s.repo.FindAll()
}

// Template devuelve una plantilla concreta.
func (s *Service) Template(key, language string) (repositories.Template, error) {
	return s.repo.Find(key, language)
}

// Update guarda el texto de una plantilla, si sus marcadores son los de ese correo.
func (s *Service) Update(key, language, subject, body string, updatedByID *int64) error {
	if err := validateTarget(key, language); err != nil {
		return err
	}
	if strings.TrimSpace(subject) == "" {
		return fmt.Errorf("mail.subject.required")
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("mail.body.required")
	}
	if err := Validate(key, subject, body); err != nil {
		return err
	}

	return s.repo.Update(key, language, subject, body, updatedByID)
}

// Reset devuelve una plantilla a su texto de fábrica.
func (s *Service) Reset(key, language string) error {
	if err := validateTarget(key, language); err != nil {
		return err
	}
	return s.repo.Reset(key, language)
}

// Link construye un enlace con la dirección pública de la aplicación.
//
// Los enlaces no se escriben a mano en las plantillas: el mismo texto tiene que servir en
// desarrollo y en producción (docs/modules/mail.md, sección 4).
func (s *Service) Link(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// **La dirección de la configuración manda** (decisión 14): la vista previa tiene que enseñar el
	// enlace que va a salir de verdad, no el del entorno.
	return s.baseDeLosEnlaces() + path
}

// MissingEssential dice qué marcadores imprescindibles faltan en una plantilla, para que el editor
// pueda avisar sin impedir guardar.
func (s *Service) MissingEssential(key, subject, body string) []string {
	return MissingEssential(key, subject, body)
}

// Send manda un correo y devuelve el error si algo falla.
//
// Es síncrono a propósito: lo usan la prueba del editor y quien quiera saber si salió. Para el resto
// de los avisos está SendAsync.
func (s *Service) Send(key, language string, to []string, data map[string]string) error {
	if err := validateTarget(key, language); err != nil {
		return err
	}
	if len(to) == 0 {
		return fmt.Errorf("mail.to.empty")
	}

	template, err := s.repo.Find(key, language)
	if err != nil {
		return err
	}

	subject, err := RenderSubject(key, template.Subject, data)
	if err != nil {
		return err
	}
	body, err := RenderBody(key, template.Body, data)
	if err != nil {
		return err
	}

	if err := s.sender.Send(to, subject, body); err != nil {
		logs.LogError(fmt.Sprintf("correo %s (%s) a %s: %v", key, language, strings.Join(to, ", "), err))
		return err
	}

	logs.LogSuccess(fmt.Sprintf("correo %s (%s) enviado a %s", key, language, strings.Join(to, ", ")))

	return nil
}

// Probar comprueba **la conexión y la autenticación** del correo saliente, con los datos que se le
// mandan, **sin mandar ningún correo**: lo usa el asistente de primer arranque, donde todavía no hay
// destinatario y sólo se quiere saber si el servidor contesta y si las credenciales valen
// (docs/primer-arranque.md, sección 3).
//
// Es la prueba que el módulo ofrece hacia fuera; el protocolo vive en el remitente, que es quien sabe
// hablar con un servidor de correo.
func (s *Service) Probar(correo CorreoSaliente) error {
	if s.sender == nil {
		return ErrNotConfigured
	}

	return s.sender.Probar(correo)
}

// SendAsync manda un correo en segundo plano y no devuelve nada.
//
// Es lo que usan los avisos: quien crea un ticket no tiene por qué esperar a que responda el
// servidor de correo, y un fallo de correo no puede tumbar la acción que lo provocó
// (docs/arquitectura.md, sección 9). **El fallo queda en el log**, y eso es parte del contrato: un
// correo que no sale y no se cuenta es un correo que nadie echa de menos.
func (s *Service) SendAsync(key, language string, to []string, data map[string]string) {
	if len(to) == 0 {
		logs.LogWarning(fmt.Sprintf("correo %s (%s) sin destinatarios: no se envía", key, language))
		return
	}

	go func() {
		// Aquí sólo se evita que el pánico de una goroutine tumbe el proceso entero.
		defer func() {
			if recovered := recover(); recovered != nil {
				logs.LogError(fmt.Sprintf("correo %s (%s): pánico al enviar: %v", key, language, recovered))
			}
		}()

		// **El error se registra aquí, y no se pierde**: nadie más lo mira —el aviso sale de una
		// goroutine y la acción que lo provocó ya ha terminado—, así que sin esta línea una plantilla
		// mal puesta o un servidor de correo caído **no avisan a nadie, ni al log**. Es exactamente lo
		// que pasó el 2026-09-27 con el aviso del etiquetado: la plantilla existía en la migración,
		// `markers.go` no la conocía, y el correo no salió ni quedó rastro
		// (docs/modules/mail.md, decisión 20).
		if err := s.Send(key, language, to, data); err != nil {
			// La dirección del destinatario **sí** se registra, que es lo que permite responder a «no
			// me ha llegado nada»; el contenido del correo, nunca (AGENTS.md).
			logs.LogWarning(fmt.Sprintf(
				"no se ha podido enviar el correo %s (%s) a %s: %s",
				key,
				language,
				strings.Join(to, ", "),
				err.Error(),
			))
		}
	}()
}

// Preview devuelve el correo ya renderizado, con los mismos datos de ejemplo que la prueba.
//
// **Lo renderiza el backend a propósito**: la sustitución de marcadores y el escapado viven aquí, y
// una segunda versión en el frontend acabaría enseñando una vista previa que miente
// (docs/modules/mail.md, decisión 17).
//
// `subject` y `body` son **el borrador que se está escribiendo**. Si llegan vacíos, se renderiza lo
// que hay guardado: así la vista previa sirve antes de guardar, que es cuando hace falta.
func (s *Service) Preview(key, language, subject, body string) (Preview, error) {
	if err := validateTarget(key, language); err != nil {
		return Preview{}, err
	}

	if strings.TrimSpace(subject) == "" && strings.TrimSpace(body) == "" {
		template, err := s.repo.Find(key, language)
		if err != nil {
			return Preview{}, err
		}

		subject, body = template.Subject, template.Body
	}

	data := s.exampleData(key)

	asunto, err := RenderSubject(key, subject, data)
	if err != nil {
		return Preview{}, err
	}

	cuerpo, err := RenderBody(key, body, data)
	if err != nil {
		return Preview{}, err
	}

	return Preview{Subject: asunto, Body: cuerpo, Text: HTMLToText(cuerpo)}, nil
}

// Preview es un correo ya renderizado, tal y como lo va a recibir alguien.
type Preview struct {
	Subject string
	Body    string
	// Text es la versión de texto, que es la que lee quien no pinta HTML.
	Text string
}

// exampleData son los datos de ejemplo con el enlace ya compuesto.
func (s *Service) exampleData(key string) map[string]string {
	data := ExampleData(key, s.Link("/"))
	// **La fecha de ejemplo se escribe en la zona de la instalación** (decisión 15): la vista previa
	// tiene que enseñar lo mismo que va a salir, y en UTC saldría a otra hora.
	data["cuando"] = time.Now().In(s.zonaHoraria()).Format("2006-01-02 15:04")
	data["enlace"] = s.Link("/tickets/ACME-2026-0042")

	return data
}

// SendTest manda una prueba de una plantilla a quien la está editando.
//
// Lleva datos de ejemplo y lo dice en el asunto, para que nadie confunda una prueba con un correo de
// verdad.
func (s *Service) SendTest(key, language, to string) error {
	return s.SendTestWithData(key, language, to, s.exampleData(key))
}

// SendTestWithData es SendTest con los datos que se quieran, para las pruebas del módulo.
func (s *Service) SendTestWithData(key, language, to string, data map[string]string) error {
	if err := validateTarget(key, language); err != nil {
		return err
	}

	template, err := s.repo.Find(key, language)
	if err != nil {
		return err
	}

	subject, err := RenderSubject(key, template.Subject, data)
	if err != nil {
		return err
	}
	body, err := RenderBody(key, template.Body, data)
	if err != nil {
		return err
	}

	return s.sender.Send([]string{to}, "[prueba] "+subject, body)
}

// ExampleData son los datos de ejemplo que usa la prueba del editor.
func ExampleData(key, link string) map[string]string {
	return map[string]string{
		"numero":      "ACME-2026-0042",
		"asunto":      "No puedo entrar al sistema",
		"solicitante": "Ana Pérez",
		"motivo":      "Es un cambio en el cálculo de la fecha de calibración",
		"nombre":      "Ana",
		"enlace":      link,
		"caducidad":   "24 horas",
		"cuando":      time.Now().Format("2006-01-02 15:04"),
		"ip":          "10.0.0.1",
	}
}

func validateTarget(key, language string) error {
	if _, ok := Markers(key); !ok {
		return fmt.Errorf("mail.template.unknownKey: %s", key)
	}
	if !LanguageIsValid(language) {
		return fmt.Errorf("mail.language.unknown: %s", language)
	}
	return nil
}
