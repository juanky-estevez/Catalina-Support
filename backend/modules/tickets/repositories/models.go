// Package repositories es el acceso a datos del módulo tickets: los dos tipos de ticket y lo que
// cuelga de ellos (docs/modules/tickets.md, sección 2).
package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Los errores del acceso a datos. El servicio los traduce a claves de la API.
var (
	ErrTicketNotFound     = errors.New("tickets.notFound")
	ErrCommentNotFound    = errors.New("tickets.comment.notFound")
	ErrAttachmentNotFound = errors.New("tickets.attachment.notFound")
	ErrObserverNotFound   = errors.New("tickets.observer.notFound")
	ErrCategoryNotFound   = errors.New("tickets.category.notFound")
	ErrCategoryDuplicate  = errors.New("tickets.category.duplicate")
	ErrTagNotFound        = errors.New("tickets.tag.notFound")
	ErrTagDuplicate       = errors.New("tickets.tag.duplicate")
)

// Ticket es un ticket principal.
type Ticket struct {
	ID          int64  `gorm:"primaryKey"`
	Number      string `gorm:"column:number"`
	NumberYear  int    `gorm:"column:number_year"`
	NumberSeq   int    `gorm:"column:number_seq"`
	Subject     string `gorm:"column:subject"`
	Description string `gorm:"column:description"`
	State       string `gorm:"column:state"`
	// La categoría del ticket: obligatoria y una sola, y la garantiza la base con `NOT NULL` y clave
	// ajena (docs/modules/tickets.md, decisión 65).
	CategoryID  int64 `gorm:"column:category_id"`
	RequesterID int64 `gorm:"column:requester_id"`
	CreatedByID int64 `gorm:"column:created_by_id"`
	// Nulo sólo si no había ningún técnico activo al repartir.
	AssigneeID *int64 `gorm:"column:assignee_id"`
	// Una marca por campo: dice **qué** se editó, sin tener que leer el historial.
	SubjectEditedAt     *time.Time `gorm:"column:subject_edited_at"`
	DescriptionEditedAt *time.Time `gorm:"column:description_edited_at"`
	ResolvedAt          *time.Time `gorm:"column:resolved_at"`
	ClosedAt            *time.Time `gorm:"column:closed_at"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (Ticket) TableName() string { return "tickets" }

// InternalTicket es un ticket interno: uno por principal, y su contenido es el motivo del escalado.
type InternalTicket struct {
	ID               int64      `gorm:"primaryKey"`
	TicketID         int64      `gorm:"column:ticket_id"`
	Number           string     `gorm:"column:number"`
	State            string     `gorm:"column:state"`
	EscalationReason string     `gorm:"column:escalation_reason"`
	CreatedByID      int64      `gorm:"column:created_by_id"`
	AssigneeID       *int64     `gorm:"column:assignee_id"`
	ResolvedAt       *time.Time `gorm:"column:resolved_at"`
	ClosedAt         *time.Time `gorm:"column:closed_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (InternalTicket) TableName() string { return "internal_tickets" }

// Comment es un comentario, de un principal o de un interno: exactamente uno de los dos.
type Comment struct {
	ID               int64  `gorm:"primaryKey"`
	TicketID         *int64 `gorm:"column:ticket_id"`
	InternalTicketID *int64 `gorm:"column:internal_ticket_id"`
	AuthorID         int64  `gorm:"column:author_id"`
	// Vacío cuando el comentario está borrado: el texto se vacía de verdad.
	Body      string     `gorm:"column:body"`
	EditedAt  *time.Time `gorm:"column:edited_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (Comment) TableName() string { return "ticket_comments" }

// Attachment es un adjunto. El archivo vive en disco; aquí está lo que hace falta para servirlo.
type Attachment struct {
	ID               int64  `gorm:"primaryKey"`
	TicketID         *int64 `gorm:"column:ticket_id"`
	InternalTicketID *int64 `gorm:"column:internal_ticket_id"`
	// Nulo cuando el adjunto va en la descripción inicial y no en un comentario.
	CommentID    *int64    `gorm:"column:comment_id"`
	UploadedByID int64     `gorm:"column:uploaded_by_id"`
	Filename     string    `gorm:"column:filename"`
	StoredName   string    `gorm:"column:stored_name"`
	ContentType  string    `gorm:"column:content_type"`
	SizeBytes    int64     `gorm:"column:size_bytes"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (Attachment) TableName() string { return "ticket_attachments" }

// HistoryEntry es lo que hizo el sistema. No lo escribe nadie a mano: lo genera cada transición.
type HistoryEntry struct {
	ID               int64  `gorm:"primaryKey"`
	TicketID         *int64 `gorm:"column:ticket_id"`
	InternalTicketID *int64 `gorm:"column:internal_ticket_id"`
	// Nulo cuando lo hizo el sistema.
	ActorID   *int64    `gorm:"column:actor_id"`
	Event     string    `gorm:"column:event"`
	FromState *string   `gorm:"column:from_state"`
	ToState   *string   `gorm:"column:to_state"`
	Detail    *string   `gorm:"column:detail"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (HistoryEntry) TableName() string { return "ticket_history" }

// TicketObserver es un observador: **quien sigue el ticket sin atenderlo**, que no es lo mismo que el
// asignado (docs/modules/tickets.md, sección 2.3.1). Vive en su tabla y no se deduce del texto al
// vuelo, porque **quitar a un observador lo puede hacer cualquier técnico o desarrollador** y quien
// quita no es el autor del comentario que lo nombró (decisión 63).
type TicketObserver struct {
	ID int64 `gorm:"primaryKey"`
	// Uno de los dos, y sólo uno: cada hilo tiene sus observadores, como sus comentarios.
	TicketID         *int64 `gorm:"column:ticket_id"`
	InternalTicketID *int64 `gorm:"column:internal_ticket_id"`
	// La cuenta que observa, y quien la añadió —el autor de la mención—.
	AccountID int64     `gorm:"column:account_id"`
	AddedByID int64     `gorm:"column:added_by_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (TicketObserver) TableName() string { return "ticket_observers" }

// TicketCategory es una categoría del catálogo: **el «qué es» de un ticket**
// (docs/modules/tickets.md, sección 2.3.2). Retirar una categoría es desactivarla, no borrarla, y por
// eso la fila se queda con su marca y los tickets que la tienen la conservan (decisión 66).
type TicketCategory struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"column:name"`
	// El nombre en minúsculas y sin acentos: único, y lo que impide «Red» y «red» en el mismo catálogo.
	Normalized string `gorm:"column:normalized"`
	// Retirada deja de ofrecerse; los tickets que la tienen la siguen enseñando.
	Active bool `gorm:"column:active"`
	// Nulo en la categoría de fábrica («General»), que no la crea nadie.
	CreatedByID *int64    `gorm:"column:created_by_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	// Cuándo se retiró. Volver a activarla lo limpia.
	DeactivatedAt *time.Time `gorm:"column:deactivated_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (TicketCategory) TableName() string { return "ticket_categories" }

// TicketTagName es una etiqueta **del catálogo**: el nombre vive aquí y no en los tickets, para que
// una etiqueta pueda existir sin que ninguno la lleve y para que renombrarla valga para todos los que
// la llevan (docs/modules/tickets.md, decisión 72). Es la única fuente de verdad del nombre.
type TicketTagName struct {
	ID int64 `gorm:"primaryKey"`
	// El nombre tal cual se lee. La normalización de una etiqueta ya la deja en minúsculas y con
	// guiones, así que hoy coincide con `normalized`; la columna está para comparar siempre igual.
	Tag string `gorm:"column:tag"`
	// En minúsculas y sin acentos: único, y lo que impide `red-wifi` y `Red-Wifi` en el catálogo.
	Normalized string `gorm:"column:normalized"`
	// **Nulo puede quedar**: una etiqueta puede existir sin que nadie la creara a mano.
	CreatedByID *int64    `gorm:"column:created_by_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (TicketTagName) TableName() string { return "ticket_tag_names" }

// TicketTag es **qué ticket lleva qué etiqueta**: sólo la clave ajena al catálogo, porque el nombre
// vive allí (docs/modules/tickets.md, decisión 72).
type TicketTag struct {
	ID       int64 `gorm:"primaryKey"`
	TicketID int64 `gorm:"column:ticket_id"`
	TagID    int64 `gorm:"column:tag_id"`
	// Quien la puso en el ticket. Es una cuenta real: poner una etiqueta es una decisión de alguien.
	CreatedByID int64     `gorm:"column:created_by_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (TicketTag) TableName() string { return "ticket_tags" }

// NumberCounter es el último número emitido en un año.
type NumberCounter struct {
	Year       int `gorm:"primaryKey"`
	LastNumber int `gorm:"column:last_number"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (NumberCounter) TableName() string { return "ticket_number_counters" }

// TicketRepository lee y escribe los tickets y su numeración.
type TicketRepository struct {
	db *gorm.DB
}

// NewTicketRepository construye el repositorio.
func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// DB deja la conexión a quien necesita una transacción que abarca más de una tabla.
func (r *TicketRepository) DB() *gorm.DB { return r.db }
