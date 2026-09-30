package services

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

// Build arma el mensaje completo tal como lo espera SMTP: cabeceras y cuerpo, con las dos versiones
// del contenido.
//
// Se manda multipart/alternative —texto y HTML— porque un cliente que no pinte HTML enseñaría un
// correo vacío. La versión de texto se saca del HTML (docs/modules/mail.md, sección 6).
func Build(fromName, fromEmail string, to []string, subject, htmlBody string) ([]byte, error) {
	if _, err := mail.ParseAddress(fromEmail); err != nil {
		return nil, fmt.Errorf("mail.from.invalid: %w", err)
	}
	for _, address := range to {
		if _, err := mail.ParseAddress(address); err != nil {
			return nil, fmt.Errorf("mail.to.invalid: %w", err)
		}
	}

	boundary, err := boundary()
	if err != nil {
		return nil, err
	}

	var message bytes.Buffer

	writeHeader := func(name, value string) {
		fmt.Fprintf(&message, "%s: %s\r\n", name, value)
	}

	// El nombre y el asunto se codifican: llevan acentos, y una cabecera no admite UTF-8 crudo.
	writeHeader("From", fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", fromName), fromEmail))
	writeHeader("To", strings.Join(to, ", "))
	writeHeader("Subject", mime.QEncoding.Encode("utf-8", subject))
	writeHeader("Date", time.Now().Format(time.RFC1123Z))
	writeHeader("MIME-Version", "1.0")
	writeHeader("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%q", boundary))

	writePart := func(contentType, content string) error {
		fmt.Fprintf(&message, "\r\n--%s\r\n", boundary)
		fmt.Fprintf(&message, "Content-Type: %s; charset=UTF-8\r\n", contentType)
		message.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")

		writer := quotedprintable.NewWriter(&message)
		if _, err := writer.Write([]byte(content)); err != nil {
			return err
		}
		return writer.Close()
	}

	if err := writePart("text/plain", HTMLToText(htmlBody)); err != nil {
		return nil, err
	}
	if err := writePart("text/html", htmlBody); err != nil {
		return nil, err
	}

	fmt.Fprintf(&message, "\r\n--%s--\r\n", boundary)

	return message.Bytes(), nil
}

// HTMLToText saca una versión legible del HTML: es la parte de texto del correo.
//
// No es un conversor completo ni pretende serlo: quita etiquetas, respeta los saltos que el HTML
// expresa con `</p>`, `</div>` y `</li>` y deshace las entidades. Para los correos que se escriben
// aquí —párrafos, listas, enlaces— es suficiente; un conversor de verdad sería otra dependencia
// para el mismo resultado.
func HTMLToText(body string) string {
	replacements := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`(?i)<br\s*/?>`), "\n"},
		{regexp.MustCompile(`(?i)</(p|div|h[1-6]|tr)>`), "\n"},
		{regexp.MustCompile(`(?i)<li[^>]*>`), "\n- "},
		{regexp.MustCompile(`(?i)</li>`), ""},
		{regexp.MustCompile(`(?i)</a>`), " "},
		{regexp.MustCompile(`(?s)<[^>]*>`), ""},
	}

	text := body
	for _, replacement := range replacements {
		text = replacement.pattern.ReplaceAllString(text, replacement.replacement)
	}

	text = html.UnescapeString(text)

	// Se normalizan los saltos: ni tres seguidos ni líneas con espacios de más.
	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" && len(cleaned) > 0 && cleaned[len(cleaned)-1] == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func boundary() (string, error) {
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return "=_catalina_" + hex.EncodeToString(random), nil
}
