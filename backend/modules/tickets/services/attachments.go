package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// AttachmentFile es un adjunto listo para descargar: sus datos y cómo leerlo.
//
// El archivo **no se carga en memoria**: se devuelve abierto y quien responde lo copia a la respuesta,
// así un archivo de 25 MB no ocupa 25 MB de memoria por descarga.
type AttachmentFile struct {
	Attachment
	// Inline dice si se puede previsualizar dentro del ticket. Lo que no, se descarga siempre.
	Inline bool
	// Reader es el archivo abierto. Quien lo pide **tiene que cerrarlo**.
	Reader io.ReadCloser
}

// Attach guarda un adjunto en un ticket.
//
// Los adjuntos siguen la misma regla que los comentarios: **si puedes comentar en un ticket, puedes
// adjuntar en él** (docs/modules/tickets.md, sección 4). No es un permiso nuevo.
func (s *Service) Attach(number string, file multipart.File, cabecera *multipart.FileHeader, commentID *int64, actor auth.Identity) (Attachment, error) {
	principal, esInterno, _, err := s.porNumero(number)
	if err != nil {
		return Attachment{}, err
	}

	destino, convertido, err := s.destinoDe(principal, esInterno, actor)
	if err != nil {
		return Attachment{}, err
	}

	if !puedeComentar(actor, convertido) {
		return Attachment{}, ErrForbidden
	}

	estado := principal.State
	// El número con el que se nombra la carpeta es **el del ticket de verdad**, no lo que haya escrito
	// quien sube el archivo: el interno tiene el suyo (`INT-…`) y el principal el suyo.
	numero := principal.Number
	if esInterno {
		interno, _, err := s.tickets.InternalByTicket(principal.ID)
		if err != nil {
			return Attachment{}, err
		}
		estado = interno.State
		numero = interno.Number
	}
	if estado == repositories.StateCerrado {
		return Attachment{}, ErrClosed
	}

	// **Cada ticket tiene su carpeta**, y el interno la suya con su propio número: así, mirando el
	// disco, se sabe de qué ticket es cada archivo y qué se subió en la conversación interna
	// (docs/modules/tickets.md, sección 2.3).
	carpeta := carpetasDelTicket(principal.NumberYear, numero)

	nombre := filepath.Base(cabecera.Filename)
	if !extensionAdmitida(nombre) {
		return Attachment{}, ErrAttachmentFormat
	}

	// El tamaño se comprueba antes de escribir nada: nginx corta en 30 MB, y aquí el tope son 25.
	if cabecera.Size > MaxAttachmentBytes {
		return Attachment{}, ErrAttachmentTooBig
	}

	// El comentario, si lo hay, tiene que ser de ese mismo ticket.
	if commentID != nil {
		if _, err := s.comentarioDelTicket(*commentID, destino); err != nil {
			return Attachment{}, err
		}
	}

	guardado, tamano, err := s.guardarArchivo(file, nombre, carpeta)
	if err != nil {
		return Attachment{}, err
	}

	adjunto := repositories.Attachment{
		TicketID:         destino.TicketID,
		InternalTicketID: destino.InternalTicketID,
		CommentID:        commentID,
		UploadedByID:     actor.ID,
		Filename:         nombre,
		StoredName:       guardado,
		ContentType:      tipoDeArchivo(nombre),
		SizeBytes:        tamano,
	}

	var creado repositories.Attachment
	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		var err error
		creado, err = s.conversation.CreateAttachment(tx, adjunto)
		return err
	})
	if err != nil {
		// Si la fila no se ha podido guardar, el archivo tampoco se queda: un archivo sin dueño es
		// basura que nadie va a limpiar.
		s.borrarArchivo(guardado)
		return Attachment{}, err
	}

	return Attachment{
		ID:          creado.ID,
		Filename:    creado.Filename,
		ContentType: creado.ContentType,
		Size:        creado.SizeBytes,
		CommentID:   creado.CommentID,
		UploadedBy:  cuentaDe(actor),
		CreatedAt:   creado.CreatedAt,
	}, nil
}

// Attachment abre un adjunto para descargarlo, **comprobando que esa persona puede ver el ticket**.
//
// Se descargan por la API y nunca por una ruta estática: el adjunto de un ticket interno no puede ser
// alcanzable por el solicitante, y eso sólo se garantiza si cada descarga pasa por aquí.
func (s *Service) Attachment(number string, attachmentID int64, actor auth.Identity) (AttachmentFile, error) {
	principal, esInterno, _, err := s.porNumero(number)
	if err != nil {
		return AttachmentFile{}, err
	}

	destino, convertido, err := s.destinoDe(principal, esInterno, actor)
	if err != nil {
		return AttachmentFile{}, err
	}

	if !puedeVer(actor, convertido) {
		return AttachmentFile{}, ErrTicketNotFound
	}

	adjunto, err := s.conversation.AttachmentByID(attachmentID)
	if err != nil {
		if err == repositories.ErrAttachmentNotFound {
			return AttachmentFile{}, ErrAttachmentMissing
		}

		return AttachmentFile{}, err
	}

	esDeEseTicket := (destino.TicketID != nil && adjunto.TicketID != nil && *adjunto.TicketID == *destino.TicketID) ||
		(destino.InternalTicketID != nil && adjunto.InternalTicketID != nil && *adjunto.InternalTicketID == *destino.InternalTicketID)
	if !esDeEseTicket {
		return AttachmentFile{}, ErrAttachmentMissing
	}

	archivo, err := os.Open(s.rutaDe(adjunto.StoredName))
	if err != nil {
		return AttachmentFile{}, ErrAttachmentMissing
	}

	return AttachmentFile{
		Attachment: Attachment{
			ID:          adjunto.ID,
			Filename:    adjunto.Filename,
			ContentType: adjunto.ContentType,
			Size:        adjunto.SizeBytes,
			CommentID:   adjunto.CommentID,
			CreatedAt:   adjunto.CreatedAt,
		},
		Inline: sePuedePrevisualizar(adjunto.Filename),
		Reader: archivo,
	}, nil
}

// guardarArchivo escribe el archivo en disco, dentro de la carpeta de su ticket y con un nombre
// generado.
//
// El nombre original **no se usa en el disco**: puede repetirse, traer acentos o intentar salirse de
// la carpeta. Se guarda un nombre generado, y el original se queda en la base para enseñarlo.
//
// Lo que devuelve es la **ruta relativa** de la carpeta de archivos (`2026/CS-2026-0042/ab12….png`),
// que es lo que se guarda en `stored_name`.
func (s *Service) guardarArchivo(file multipart.File, nombre, carpeta string) (string, int64, error) {
	carpetaCompleta := filepath.Join(s.filesPath, filepath.FromSlash(carpeta))
	if err := os.MkdirAll(carpetaCompleta, 0o755); err != nil {
		return "", 0, err
	}

	generado, err := nombreGenerado(nombre)
	if err != nil {
		return "", 0, err
	}

	guardado := path.Join(carpeta, generado)

	destino, err := os.Create(s.rutaDe(guardado))
	if err != nil {
		return "", 0, err
	}
	defer destino.Close()

	tamano, err := io.Copy(destino, io.LimitReader(file, MaxAttachmentBytes+1))
	if err != nil {
		s.borrarArchivo(generado)
		return "", 0, err
	}

	// El tope también se comprueba mientras se copia: el `Size` de la cabecera lo dice el cliente, y de
	// un cliente no se fía uno.
	if tamano > MaxAttachmentBytes {
		s.borrarArchivo(guardado)
		return "", 0, ErrAttachmentTooBig
	}
	if tamano == 0 {
		s.borrarArchivo(guardado)
		return "", 0, ErrAttachmentInvalid
	}

	return guardado, tamano, nil
}

// carpetasDelTicket arma la ruta relativa donde viven los adjuntos de un ticket: **el año y su
// número**, que es lo que hace que el disco se pueda mirar y entender (docs/modules/tickets.md,
// sección 2.3).
//
// Va con barras, y no con el separador del sistema, porque es lo que se guarda en la base y lo que
// tiene que valer igual en cualquier sitio; al tocar el disco se traduce.
//
// Y se limpia: el número lo compone el sistema —prefijo, año y secuencia—, pero **nada que venga de
// fuera entra en una ruta sin pasar por aquí**, que es la misma razón por la que el nombre del
// archivo se genera en vez de usarse el original.
func carpetasDelTicket(year int, numero string) string {
	limpio := strings.Map(func(letra rune) rune {
		switch {
		case letra >= 'a' && letra <= 'z', letra >= 'A' && letra <= 'Z', letra >= '0' && letra <= '9':
			return letra
		case letra == '-':
			return letra
		default:
			return -1
		}
	}, numero)

	return path.Join(strconv.Itoa(year), limpio)
}

// rutaDe arma la ruta completa de un archivo guardado.
//
// `stored_name` es una **ruta relativa** desde la carpeta de archivos: `2026/CS-2026-0042/ab12….png`.
// Los adjuntos que se subieron antes de que existieran las carpetas guardan sólo su nombre, y siguen
// resolviendo igual: se quedan donde están y se ven igual (docs/modules/tickets.md, sección 2.3).
func (s *Service) rutaDe(nombre string) string {
	return filepath.Join(s.filesPath, filepath.FromSlash(nombre))
}

// borrarArchivo quita un archivo que se quedó sin fila.
//
// Si con él se queda vacía la carpeta de su ticket, se quita también: una carpeta sin nada dentro no
// dice nada a quien mira el disco.
func (s *Service) borrarArchivo(nombre string) {
	ruta := s.rutaDe(nombre)

	if err := os.Remove(ruta); err != nil && !errors.Is(err, os.ErrNotExist) {
		// No se puede hacer nada más: el archivo se queda huérfano y la copia de seguridad lo incluirá.
		return
	}

	if carpeta := filepath.Dir(ruta); carpeta != s.filesPath {
		// `os.Remove` no borra una carpeta con algo dentro, así que esto no se lleva nada por delante.
		_ = os.Remove(carpeta)
	}
}

// nombreGenerado inventa el nombre del archivo en disco: el original con un prefijo aleatorio, y sin
// nada que se pueda salir de la carpeta.
func nombreGenerado(original string) (string, error) {
	azar := make([]byte, 16)
	if _, err := rand.Read(azar); err != nil {
		return "", err
	}

	extension := strings.ToLower(filepath.Ext(original))

	return hex.EncodeToString(azar) + extension, nil
}

// extensionAdmitida dice si el archivo está en la lista cerrada de extensiones.
//
// Lo que no esté en la lista **se rechaza**: no se acepta «cualquier cosa» y luego se mira.
func extensionAdmitida(nombre string) bool {
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(nombre), "."))
	if extension == "" {
		return false
	}

	return extensionesAdmitidas[extension]
}

// sePuedePrevisualizar dice si ese archivo se puede servir en línea dentro del ticket: las imágenes,
// los vídeos y el PDF, que es lo que se ve dentro del texto y lo que enseña el visor.
//
// **El `svg` nunca**: aunque sea una imagen, puede llevar código dentro y el navegador lo ejecuta al
// abrirlo en línea (docs/modules/tickets.md, sección 4).
func sePuedePrevisualizar(nombre string) bool {
	// Las imágenes y los vídeos se ven dentro del texto, y el PDF lo enseña el visor: los tres van en
	// línea. Todo lo demás se descarga siempre.
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(nombre), ".")) {
	case "png", "jpg", "jpeg", "gif", "webp",
		"mp4", "webm", "mov", "avi",
		"pdf":
		return true
	default:
		return false
	}
}

// tipoDeArchivo adivina el tipo por la extensión, con un valor por defecto que obliga a descargar.
func tipoDeArchivo(nombre string) string {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(nombre), ".")) {
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "pdf":
		return "application/pdf"
	// Los cuatro de vídeo, con su tipo para que el reproductor del navegador sepa qué le está llegando.
	case "mp4":
		return "video/mp4"
	case "webm":
		return "video/webm"
	case "mov":
		return "video/quicktime"
	case "avi":
		return "video/x-msvideo"
	// **El texto y el código se guardan con su tipo de texto**, que es lo que la API dice que son y lo
	// que enseña la ficha del adjunto. **La descarga va igual para todos los que no se previsualizan**
	// —`application/octet-stream` y `attachment`, en `ServirCon`—: aquí no se decide cómo se sirve un
	// archivo, sino qué es.
	case "txt", "log", "md", "css", "csv",
		"sql", "json", "xml", "yml", "yaml", "ini", "conf", "cnf",
		"sh", "bash", "bat", "ps1", "py", "js", "ts", "java", "php", "go", "cs", "rb", "pl",
		"kt", "rs", "swift", "c", "h", "cpp", "hpp", "vue", "jsx", "tsx",
		"htaccess", "env", "properties", "diff", "patch", "bak", "old":
		return "text/plain; charset=utf-8"
	case "svg":
		return "image/svg+xml"
	case "zip":
		return "application/zip"
	case "rar":
		return "application/vnd.rar"
	case "tar":
		return "application/x-tar"
	case "gz", "tgz":
		return "application/gzip"
	case "7z":
		return "application/x-7z-compressed"
	case "doc":
		return "application/msword"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xls":
		return "application/vnd.ms-excel"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "ppt":
		return "application/vnd.ms-powerpoint"
	case "pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case "odt":
		return "application/vnd.oasis.opendocument.text"
	case "ods":
		return "application/vnd.oasis.opendocument.spreadsheet"
	case "odp":
		return "application/vnd.oasis.opendocument.presentation"
	default:
		// Todo lo demás se sirve como descarga, que es lo que impide que un archivo se abra solo.
		return "application/octet-stream"
	}
}

// ServirCon define cómo se sirve un adjunto: en línea lo que se puede ver, y descarga lo demás.
func ServirCon(nombre string, inline bool) (contentType, disposicion string) {
	tipo := tipoDeArchivo(nombre)
	if inline && sePuedePrevisualizar(nombre) {
		return tipo, "inline"
	}

	return "application/octet-stream", "attachment"
}
