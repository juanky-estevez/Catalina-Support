// Package services es la lógica del módulo tickets: la numeración, el reparto, las transiciones de
// los dos ciclos de vida y los avisos por correo (docs/modules/tickets.md).
package services

import (
	"errors"
	"time"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// Las claves de error del módulo, con el código que les toca en el controlador
// (docs/interfaz-y-experiencia.md, sección 8: los errores viajan como claves).
var (
	ErrTicketNotFound  = errors.New("tickets.notFound")
	ErrSubjectRequired = errors.New("tickets.subject.required")
	ErrBodyRequired    = errors.New("tickets.description.required")
	ErrReasonRequired  = errors.New("tickets.reason.required")
	// ErrCommentRequired: un comentario **editado** se queda sin texto. Al escribir uno nuevo el texto
	// puede ir vacío —un comentario puede ser sólo un adjunto—, pero dejar en blanco uno que ya se
	// escribió no es lo mismo, y para eso está borrarlo (docs/modules/tickets.md, decisión 42).
	ErrCommentRequired = errors.New("tickets.comment.required")
	ErrStateUnknown    = errors.New("tickets.state.unknown")
	// ErrCierreSinComentario: **un ticket no se cierra sin decir por qué** (decisión 81). Cerrar —lo
	// haga Soporte, Desarrollo o el propio solicitante— exige un comentario, y lo exige el servidor, no
	// la pantalla: lo que no puede faltar lo garantiza el backend, como la categoría obligatoria.
	ErrCierreSinComentario = errors.New("tickets.cierre.sinComentario")
	// ErrTransitionNotAllowed: el movimiento no existe en la tabla de transiciones.
	ErrTransitionNotAllowed = errors.New("tickets.transition.notAllowed")
	// ErrForbidden: el movimiento existe, pero no es de quien lo pide. Es un permiso y no una errata.
	ErrForbidden = errors.New("tickets.forbidden")
	// ErrClosed: el ticket está cerrado. Para escribir, primero se reabre (decisión 19 y 31).
	ErrClosed = errors.New("tickets.closed")
	// ErrBodyNotAllowed: el texto de un ticket o de un comentario trae HTML que no está en la lista
	// blanca. **No se limpia en silencio**: se dice, porque un texto guardado a medias es peor que uno
	// que se rechaza (docs/modules/tickets.md, sección 2.3).
	ErrBodyNotAllowed    = errors.New("tickets.body.notAllowed")
	ErrCommentNotYours   = errors.New("tickets.comment.notYours")
	ErrCommentNotFound   = errors.New("tickets.comment.notFound")
	ErrAssigneeUnknown   = errors.New("tickets.assignee.unknown")
	ErrRequesterUnknown  = errors.New("tickets.requester.unknown")
	ErrRequesterNotYours = errors.New("tickets.requester.notAllowed")
	// ErrMentionNotAllowed: el cuerpo etiqueta a alguien que no vale —no es Soporte ni Desarrollo
	// quien escribe, o el identificador no es una cuenta activa de esos dos papeles—. **Se rechaza el
	// cuerpo entero**, como el saneador: un texto guardado a medias es peor que uno que se rechaza
	// (docs/modules/tickets.md, decisión 59).
	ErrMentionNotAllowed = errors.New("tickets.mention.notAllowed")
	// ErrObserverNotFound: se intenta quitar a alguien que no observa ese hilo.
	ErrObserverNotFound     = errors.New("tickets.observer.notFound")
	ErrAttachmentFormat     = errors.New("tickets.attachment.extension")
	ErrAttachmentTooBig     = errors.New("tickets.attachment.tooBig")
	ErrAttachmentInvalid    = errors.New("tickets.attachment.invalid")
	ErrWritingDraftRequired = errors.New("tickets.writing.draft.required")
	ErrWritingEditorInvalid = errors.New("tickets.writing.editor.invalid")
	ErrWritingToneInvalid   = errors.New("tickets.writing.tone.invalid")
	ErrWritingUnavailable   = errors.New("tickets.writing.unavailable")
	ErrWritingInvalid       = errors.New("tickets.writing.invalid")
	ErrAttachmentMissing    = errors.New("tickets.attachment.notFound")
	// Las categorías y las etiquetas (docs/modules/tickets.md, sección 2.3.2, decisiones 64 a 68).
	// `required` es un ticket que llega sin categoría —no puede existir—, `notFound` una categoría que
	// no está en el catálogo y `inactive` una retirada, que los tickets ya creados conservan pero el
	// alta no puede volver a elegir.
	ErrCategoryRequired     = errors.New("tickets.category.required")
	ErrCategoryNotFound     = errors.New("tickets.category.notFound")
	ErrCategoryInactive     = errors.New("tickets.category.inactive")
	ErrCategoryNameRequired = errors.New("tickets.category.name.required")
	ErrCategoryDuplicate    = errors.New("tickets.category.duplicate")
	ErrCategoryLastActive   = errors.New("tickets.category.lastActive")
	ErrCategoryForbidden    = errors.New("tickets.category.forbidden")
	// ErrTagRequired: una etiqueta que se queda vacía al normalizarla —sólo signos, o espacios— no es
	// una etiqueta. Se rechaza en vez de guardar la cadena vacía, que no diría nada.
	ErrTagRequired = errors.New("tickets.tag.required")
	ErrTagTooLong  = errors.New("tickets.tag.tooLong")
	// Las etiquetas del catálogo (docs/modules/tickets.md, decisión 72): `notFound` es una etiqueta
	// que no está en el catálogo, `duplicate` un nombre nuevo que ya existe y `forbidden` quien no
	// mantiene el catálogo. Retirar es **sólo del Administrador** (decisión 73).
	ErrTagNotFound = errors.New("tickets.tag.notFound")
	// ErrTagDesconocida: se ha intentado poner en un ticket una etiqueta **que no está en el catálogo**
	// y quien lo intenta no es el Administrador (decisión 84): el catálogo se cura, y quien etiqueta
	// elige de las que hay.
	ErrTagDesconocida = errors.New("tickets.etiqueta.desconocida")
	ErrTagDuplicate   = errors.New("tickets.tag.duplicate")
	ErrTagForbidden   = errors.New("tickets.tag.forbidden")
)

// Las plantillas de los avisos. Son las claves de `mail_templates` y no se escriben en dos
// sitios: si una cambia de nombre, se cambia aquí (docs/modules/mail.md).
const (
	CorreoTicketCreado         = "ticket.created"
	CorreoTicketEscalado       = "ticket.escalated"
	CorreoTicketEsperaUsuario  = "ticket.waitingUser"
	CorreoTicketEsperaSoporte  = "ticket.waitingSupport"
	CorreoTicketVuelveASoporte = "ticket.backToSupport"
	CorreoTicketResuelto       = "ticket.resolved"
	CorreoTicketCerrado        = "ticket.closed"
	// CorreoTicketEtiquetado es el aviso de que te han etiquetado: la plantilla número once
	// (docs/modules/tickets.md, sección 5).
	CorreoTicketEtiquetado = "ticket.mentioned"
)

// Máximo de adjuntos y extensiones admitidas. La lista cerrada está en
// docs/modules/tickets.md, sección 2.3: lo que no esté, se rechaza **antes** de guardarlo.
const MaxAttachmentBytes = 25 << 20

var extensionesAdmitidas = map[string]bool{
	"pdf": true,
	"png": true, "jpg": true, "jpeg": true, "gif": true, "webp": true,
	// Los cuatro de vídeo, añadidos el 2026-09-26 (decisión 47). `mov` y `avi` son formatos de cámara y
	// muchos navegadores **no saben reproducirlos**: se admiten igual, y si el navegador no puede, el
	// reproductor no arranca y queda el botón de descargar.
	"mp4": true, "webm": true, "mov": true, "avi": true,
	"doc": true, "docx": true, "xls": true, "xlsx": true, "ppt": true, "pptx": true,
	"odt": true, "ods": true, "odp": true,
	"txt": true, "csv": true, "log": true, "md": true, "css": true,
	// **Texto y código**, añadidos el 2026-09-26 (decisión 54): es lo que se adjunta cuando el caso es
	// una consulta que falla o un despliegue que no arranca, y sin esto no había forma de mandarlo.
	// **No se pintan nunca en línea** —sólo la imagen, el vídeo y el PDF se previsualizan—, así que da
	// igual lo que lleven dentro: se descargan.
	"sql": true, "json": true, "xml": true, "yml": true, "yaml": true,
	"ini": true, "conf": true, "cnf": true, "sh": true, "bash": true, "bat": true, "ps1": true,
	"py": true, "js": true, "ts": true, "java": true, "php": true, "go": true, "cs": true,
	"rb": true, "pl": true, "kt": true, "rs": true, "swift": true,
	"c": true, "h": true, "cpp": true, "hpp": true, "vue": true, "jsx": true, "tsx": true,
	"htaccess": true, "env": true, "properties": true, "diff": true, "patch": true,
	"bak": true, "old": true,
	"zip": true, "rar": true, "tar": true, "gz": true, "tgz": true, "7z": true,
	// `svg` **no** está: puede llevar código dentro y el navegador lo ejecuta al abrirlo en línea.
}

// Accounts es lo que este módulo necesita de `users`: las cuentas que aparecen en un ticket.
//
// Se declara aquí, en quien lo usa, y `users` lo cumple sin saber que existe: así ninguno de los dos
// importa al otro y no hay círculo que no compile. Quien los une es `main.go`
// (docs/arquitectura.md, sección 4).
type Accounts interface {
	ByID(id int64) (auth.Account, error)
	// ByEmail es como se identifica a quien Soporte pone como solicitante: el correo es el
	// identificador de las personas en este producto.
	ByEmail(email string) (auth.Account, error)
	// ByIDs trae varias de una vez: es lo que evita preguntar una a una al pintar una lista.
	ByIDs(ids []int64) (map[int64]auth.Account, error)
	// ActiveByRole son las cuentas activas con ese papel: los técnicos del turno y los responsables
	// que se pueden elegir.
	ActiveByRole(role string) ([]auth.Account, error)
}

// Mailer es lo que este módulo necesita del correo: mandar un aviso —en segundo plano— y armar la
// dirección que va dentro, que es la misma base que la de los correos de cuenta.
type Mailer interface {
	SendAsync(key, language string, to []string, data map[string]string)
	Link(path string) string
}

// Configuration es lo que este módulo necesita de `settings`: el prefijo y el reparto.
//
// Se lee **en cada ticket nuevo** y no al arrancar: el Administrador puede cambiar el prefijo o
// apagar el reparto sin reiniciar nada.
type Configuration interface {
	TicketSettings() (TicketSettings, error)
}

type GlobalLanguage interface {
	Language() (string, error)
}

// WritingAssistant es el generador del módulo ai visto desde tickets. Tickets conserva el permiso y
// los datos del caso; el motor sólo recibe el contexto mínimo ya autorizado.
type WritingAssistant interface {
	Configured() bool
	ImproveDraft(draft, tone, recipient, ticketType, language string) (string, error)
}

type Capabilities struct {
	AIWriting bool
}

// TicketSettings es la configuración de los tickets, tal y como la necesita este módulo.
type TicketSettings struct {
	NumberPrefix         string
	MainAssignment       string
	MainNotification     string
	InternalAssignment   string
	InternalNotification string
}

// Los valores de la configuración, que son los mismos que acepta la base.
const (
	AssignmentNone       = "ninguna"
	AssignmentRoundRobin = "por_turnos"

	NotifyNobody    = "a_nadie"
	NotifyWholeTeam = "a_todo_el_equipo"
	NotifyAssigned  = "al_asignado"
)

// Ticket es un ticket —principal o interno— con lo que hace falta para pintarlo.
//
// Los internos **no repiten nada del principal** (docs/modules/tickets.md, sección 2.1): su asunto y
// su descripción son los del principal, y aquí viajan ya resueltos para que la pantalla no tenga que
// cruzar nada.
type Ticket struct {
	ID       int64
	Number   string
	Internal bool

	Subject     string
	Description string
	State       string

	// **La categoría y las etiquetas son del caso** (docs/modules/tickets.md, decisión 67): se guardan
	// en el principal y **el interno las hereda al leerse**, como hereda el asunto.
	Category *Category
	Tags     []string

	// Reason es el motivo del escalado. Sólo en los internos.
	Reason string
	// Parent es el número del principal, en los internos.
	Parent string
	// Child es el número del interno, en los principales que lo tienen.
	Child string

	Requester *auth.Account
	CreatedBy *auth.Account
	Assignee  *auth.Account

	SubjectEditedAt     *time.Time
	DescriptionEditedAt *time.Time
	ResolvedAt          *time.Time
	ClosedAt            *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time

	// Resumen son los dos campos que redacta el motor de IA: el **motivo** y la **última acción**
	// (docs/modules/ai.md). Si el motor no está, van vacíos y con su estado.
	Resumen Resumenes
}

// Comment es un comentario de la conversación.
type Comment struct {
	ID        int64
	Author    *auth.Account
	Body      string
	Edited    bool
	Deleted   bool
	CreatedAt time.Time
}

// Attachment es un adjunto. El archivo no viaja aquí: esto es lo que la pantalla enseña.
type Attachment struct {
	ID          int64
	Filename    string
	ContentType string
	Size        int64
	// CommentID dice si va con un comentario o con la descripción inicial.
	CommentID  *int64
	UploadedBy *auth.Account
	CreatedAt  time.Time
}

// HistoryEntry es lo que hizo el sistema, para la línea de tiempo.
type HistoryEntry struct {
	ID        int64
	Event     string
	FromState string
	ToState   string
	Detail    string
	// Actor nulo es «lo hizo el sistema».
	Actor     *auth.Account
	CreatedAt time.Time
}

// Observer es un observador del ticket: la persona y quién la añadió.
//
// **Ser observador no da permisos** (docs/modules/tickets.md, decisión 60): sólo hace que el ticket
// esté en su bandeja. Por eso aquí no hay nada del ticket, sólo quién sigue y desde cuándo.
type Observer struct {
	ID        int64
	Account   *auth.Account
	AddedBy   *auth.Account
	CreatedAt time.Time
}

// Detail es un ticket con todo lo que cuelga de él.
type Detail struct {
	Ticket      Ticket
	Comments    []Comment
	Attachments []Attachment
	History     []HistoryEntry
	// Observers sólo viaja en la ficha: en las listas paginadas sería una consulta por página para un
	// dato que ahí no se usa (docs/modules/tickets.md, sección 5).
	Observers    []Observer
	Capabilities Capabilities
}

// Page es una página de la bandeja.
type Page struct {
	Tickets []Ticket
	Total   int64
	Page    int
	PerPage int
}

// Service es el módulo.
type Service struct {
	tickets      *repositories.TicketRepository
	conversation *repositories.ConversationRepository
	// categories es el catálogo de categorías y las etiquetas: se declara por interfaz para que las
	// pruebas del servicio puedan doblarlo sin base de datos, como `Accounts`.
	categories Catalogo
	accounts   Accounts
	mailer     Mailer
	config     Configuration
	language   GlobalLanguage
	// insights es el motor de IA, y **puede ser nulo**: sin él los dos resúmenes se quedan sin texto
	// y todo lo demás funciona igual (docs/modules/ai.md, decisión 2).
	insights  Insights
	writing   WritingAssistant
	filesPath string
	now       func() time.Time
}

// NewService construye el servicio.
func NewService(
	tickets *repositories.TicketRepository,
	conversation *repositories.ConversationRepository,
	categories Catalogo,
	filesPath string,
) *Service {
	return &Service{
		tickets:      tickets,
		conversation: conversation,
		categories:   categories,
		filesPath:    filesPath,
		now:          time.Now,
	}
}

// SetAccounts conecta el módulo de cuentas, de donde salen los técnicos del turno y los nombres.
func (s *Service) SetAccounts(accounts Accounts) { s.accounts = accounts }

// SetMailer conecta el correo: los siete avisos salen por aquí.
func (s *Service) SetMailer(mailer Mailer) { s.mailer = mailer }

// SetConfiguration conecta la configuración de la instalación: el prefijo y el reparto.
func (s *Service) SetConfiguration(config Configuration) { s.config = config }

func (s *Service) SetGlobalLanguage(language GlobalLanguage) { s.language = language }

func (s *Service) SetWritingAssistant(writing WritingAssistant) { s.writing = writing }

// esSuyo dice si el ticket es de esa persona: es lo que hace que un usuario vea «los suyos».
func esSuyo(actor auth.Identity, ticket Ticket) bool {
	return ticket.Requester != nil && ticket.Requester.ID == actor.ID
}

// puedeVer aplica la matriz (docs/usuarios-y-permisos.md, sección 3).
//
// **El usuario sólo ve lo suyo y nunca un interno**: esa es la regla de lectura, y por eso vive aquí
// y no en la ruta, que no sabe de quién es cada ticket.
func puedeVer(actor auth.Identity, ticket Ticket) bool {
	switch actor.Role {
	case auth.RoleAdministrador, auth.RoleSoporte, auth.RoleDesarrollo:
		return true
	default:
		return !ticket.Internal && esSuyo(actor, ticket)
	}
}

// puedeComentar: «si puedes comentar en un ticket, puedes adjuntar en él».
func puedeComentar(actor auth.Identity, ticket Ticket) bool {
	if ticket.Internal {
		return actor.Role == auth.RoleSoporte || actor.Role == auth.RoleDesarrollo
	}

	switch actor.Role {
	case auth.RoleSoporte:
		return true
	case auth.RoleUsuario:
		return esSuyo(actor, ticket)
	default:
		// Desarrollo lee el principal para tener contexto, y el Administrador lo mira: ninguno de los
		// dos escribe en él (regla 2 de la matriz).
		return false
	}
}

// puedeEditar: el asunto y la descripción los cambian el solicitante y Soporte.
func puedeEditar(actor auth.Identity, ticket Ticket) bool {
	if ticket.Internal {
		return false
	}

	switch actor.Role {
	case auth.RoleSoporte:
		return true
	case auth.RoleUsuario:
		return esSuyo(actor, ticket)
	default:
		return false
	}
}

func puedeMejorarRedaccion(actor auth.Identity, ticket Ticket) bool {
	if ticket.State == "cerrado" || !actor.Can(auth.RoleSoporte, auth.RoleDesarrollo) {
		return false
	}
	return puedeEditar(actor, ticket) || puedeComentar(actor, ticket)
}

func (s *Service) capacidadDeRedaccion(actor auth.Identity, ticket Ticket) bool {
	return s.writing != nil && s.writing.Configured() && puedeMejorarRedaccion(actor, ticket)
}

// NormalizarFiltros deja los filtros en algo que la base entiende: página desde 1 y un tope de
// página, para que una petición con `perPage=100000` no se lleve la tabla entera.
func NormalizarFiltros(filtros repositories.Filtros) repositories.Filtros {
	if filtros.Page < 1 {
		filtros.Page = 1
	}
	if filtros.PerPage < 1 || filtros.PerPage > 100 {
		filtros.PerPage = 25
	}

	return filtros
}
