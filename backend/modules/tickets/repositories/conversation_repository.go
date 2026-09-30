package repositories

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ConversationRepository es lo que cuelga de un ticket: los comentarios, los adjuntos y el historial.
//
// Las tres tablas apuntan a **un principal o a un interno**, así que todas empiezan igual: se dice
// cuál de los dos destinos se quiere y se busca por esa columna. Es el precio de tener dos tablas de
// tickets, y se paga aquí y sólo aquí (docs/modules/tickets.md, sección 2.3).
type ConversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository construye el repositorio.
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Destino dice a qué ticket pertenece algo: uno de los dos, nunca los dos.
type Destino struct {
	TicketID         *int64
	InternalTicketID *int64
}

// DestinoDePrincipal es lo que cuelga de un ticket principal.
func DestinoDePrincipal(ticketID int64) Destino {
	return Destino{TicketID: &ticketID}
}

// DestinoDeInterno es lo que cuelga de un ticket interno.
func DestinoDeInterno(internalID int64) Destino {
	return Destino{InternalTicketID: &internalID}
}

// donde aplica el destino a una consulta.
func (d Destino) donde(consulta *gorm.DB) *gorm.DB {
	if d.TicketID != nil {
		return consulta.Where("ticket_id = ?", *d.TicketID)
	}

	return consulta.Where("internal_ticket_id = ?", *d.InternalTicketID)
}

// Comments devuelve la conversación en orden, con los borrados incluidos: el que se borró sigue
// contando que estuvo ahí.
func (r *ConversationRepository) Comments(destino Destino) ([]Comment, error) {
	return r.CommentsIn(r.db, destino)
}

// CommentsIn es lo mismo, dentro de una transacción: quien guarda un cuerpo y reajusta los
// observadores necesita leer el hilo **con el cuerpo nuevo ya escrito** (decisión 63).
func (r *ConversationRepository) CommentsIn(tx *gorm.DB, destino Destino) ([]Comment, error) {
	var comentarios []Comment

	err := destino.donde(tx.Model(&Comment{})).Order("created_at ASC").Find(&comentarios).Error
	if err != nil {
		return nil, err
	}

	return comentarios, nil
}

// CommentByID busca un comentario.
func (r *ConversationRepository) CommentByID(id int64) (Comment, error) {
	var comentario Comment

	err := r.db.Where("id = ?", id).First(&comentario).Error
	if err != nil {
		return Comment{}, traducir(err, ErrCommentNotFound)
	}

	return comentario, nil
}

// CreateComment escribe un comentario.
func (r *ConversationRepository) CreateComment(tx *gorm.DB, comentario Comment) (Comment, error) {
	if err := tx.Create(&comentario).Error; err != nil {
		return Comment{}, err
	}

	return comentario, nil
}

// UpdateComment guarda los cambios de un comentario: el texto editado o el borrado.
func (r *ConversationRepository) UpdateComment(tx *gorm.DB, comentario Comment) error {
	return tx.Save(&comentario).Error
}

// Attachments devuelve los adjuntos de un ticket en orden.
func (r *ConversationRepository) Attachments(destino Destino) ([]Attachment, error) {
	var adjuntos []Attachment

	err := destino.donde(r.db.Model(&Attachment{})).Order("created_at ASC").Find(&adjuntos).Error
	if err != nil {
		return nil, err
	}

	return adjuntos, nil
}

// AttachmentByID busca un adjunto.
func (r *ConversationRepository) AttachmentByID(id int64) (Attachment, error) {
	var adjunto Attachment

	err := r.db.Where("id = ?", id).First(&adjunto).Error
	if err != nil {
		return Attachment{}, traducir(err, ErrAttachmentNotFound)
	}

	return adjunto, nil
}

// CreateAttachment guarda un adjunto.
func (r *ConversationRepository) CreateAttachment(tx *gorm.DB, adjunto Attachment) (Attachment, error) {
	if err := tx.Create(&adjunto).Error; err != nil {
		return Attachment{}, err
	}

	return adjunto, nil
}

// History devuelve lo que hizo el sistema, en orden.
func (r *ConversationRepository) History(destino Destino) ([]HistoryEntry, error) {
	var historial []HistoryEntry

	err := destino.donde(r.db.Model(&HistoryEntry{})).Order("created_at ASC").Find(&historial).Error
	if err != nil {
		return nil, err
	}

	return historial, nil
}

// AddHistory apunta lo que acaba de pasar.
//
// Va siempre dentro de la transacción del cambio: un cambio de estado sin su entrada en el historial
// es una historia que no se puede contar, y no puede quedar a medias.
func (r *ConversationRepository) AddHistory(tx *gorm.DB, entrada HistoryEntry) error {
	return tx.Create(&entrada).Error
}

// Anotar es el atajo de AddHistory con la mitad de los campos.
func Anotar(destino Destino, actorID *int64, event, from, to, detail string) HistoryEntry {
	entrada := HistoryEntry{
		TicketID:         destino.TicketID,
		InternalTicketID: destino.InternalTicketID,
		ActorID:          actorID,
		Event:            event,
		CreatedAt:        time.Now(),
	}

	if from != "" {
		entrada.FromState = &from
	}
	if to != "" {
		entrada.ToState = &to
	}
	if detail != "" {
		entrada.Detail = &detail
	}

	return entrada
}

// Observers devuelve los observadores de un hilo, en orden de llegada.
func (r *ConversationRepository) Observers(tx *gorm.DB, destino Destino) ([]TicketObserver, error) {
	var observadores []TicketObserver

	err := destino.donde(tx.Model(&TicketObserver{})).Order("created_at ASC").Find(&observadores).Error
	if err != nil {
		return nil, err
	}

	return observadores, nil
}

// AddObserver añade un observador **y no falla si ya estaba**: mencionar dos veces a la misma
// persona no la pone dos veces (docs/modules/tickets.md, sección 2.3.1).
//
// **Devuelve si de verdad lo añadió**: `false` quiere decir que ya observaba el hilo. Quien llama lo
// necesita para no dejar dos veces la misma entrada en el historial —el historial cuenta lo que pasó,
// y no pasó nada—.
//
// El `ON CONFLICT DO NOTHING` sin destino cubre los **dos índices únicos parciales** —el del
// principal y el del interno— sin tener que decir cuál, que es lo que hace que la idempotencia no
// dependa de acertar con la columna.
func (r *ConversationRepository) AddObserver(tx *gorm.DB, observador TicketObserver) (bool, error) {
	resultado := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&observador)
	if resultado.Error != nil {
		return false, resultado.Error
	}

	// Cero filas tocadas es que el conflicto único lo dejó pasar: ya estaba.
	return resultado.RowsAffected > 0, nil
}

// RemoveObserver quita a una persona de los observadores de un hilo y dice cuántas filas había.
//
// **Se borra de verdad**: no hay marca de borrado. Quitar a un observador es quitarlo, y lo que queda
// —que se le quitó, quién y cuándo— está en el historial (docs/modules/tickets.md, sección 2.3.1).
func (r *ConversationRepository) RemoveObserver(tx *gorm.DB, accountID int64, destino Destino) (int64, error) {
	consulta := destino.donde(tx.Model(&TicketObserver{})).Where("account_id = ?", accountID)

	resultado := consulta.Delete(&TicketObserver{})
	if resultado.Error != nil {
		return 0, resultado.Error
	}

	return resultado.RowsAffected, nil
}
