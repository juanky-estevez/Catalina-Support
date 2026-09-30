package repositories

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Los estados, que son un valor cerrado y viven aquí para que la base y el código digan lo mismo.
const (
	StateNuevo      = "nuevo"
	StateEnProgreso = "en progreso"
	StateEnEspera   = "en espera"
	StateEscalado   = "escalado"
	StateResuelto   = "resuelto"
	StateCerrado    = "cerrado"
)

// Los estados que admite cada tipo de ticket (docs/modules/tickets.md, sección 2.4).
var (
	EstadosDelPrincipal = []string{StateNuevo, StateEnProgreso, StateEnEspera, StateEscalado, StateResuelto, StateCerrado}
	EstadosDelInterno   = []string{StateNuevo, StateEnProgreso, StateEnEspera, StateResuelto, StateCerrado}
)

// Filtros de la lista de tickets.
type Filtros struct {
	// Type: `principal`, `interno` o vacío para los dos (sólo el Administrador ve los dos).
	Type string
	// State vacío es «todos los estados».
	State string
	// Query busca por número y por texto del asunto, la descripción o el motivo del escalado.
	Query string
	// RequesterID limita a los tickets de una persona: es lo que hace que un usuario vea «los suyos».
	RequesterID *int64
	// Mine limita a **lo mío**: lo que tengo asignado, lo que abrí yo, aquello donde he comentado y
	// aquello donde me han etiquetado.
	//
	// Es lo que hace que la bandeja de Soporte y la de Desarrollo sean **su** bandeja y no la lista de
	// todo (docs/modules/tickets.md, decisión 51). La lista de todo son las otras dos pantallas, que se
	// piden sin este filtro y con su tipo fijo.
	Mine *int64
	// View matiza **lo mío** cuando la petición lo pide (decisión 62): `assigned` (asignados a mí),
	// `watching` (los que observo) o vacío (todo lo mío). Sin `Mine` no hace nada, porque «Tickets
	// principales» e «Tickets internos» son la lista de todo y no una vista de la bandeja.
	View string
	// CategoryID limita a una categoría del catálogo, por su identificador: es el chip de categoría
	// (docs/modules/tickets.md, decisión 68). En un interno se filtra por la de su principal, que es
	// la que hereda al leerse.
	CategoryID *int64
	// Tag limita a los tickets que llevan esa etiqueta, ya normalizada. También en un interno la
	// etiqueta es la del principal (decisión 67).
	Tag string
	// Page y PerPage son la paginación.
	Page    int
	PerPage int
}

// Las vistas del chip de «Mis tickets» (docs/modules/tickets.md, decisión 62).
const (
	ViewAssigned = "assigned"
	ViewWatching = "watching"
)

// CreateTicketInput es lo que hace falta para insertar un principal, ya numerado.
type CreateTicketInput struct {
	Number      string
	NumberYear  int
	NumberSeq   int
	Subject     string
	Description string
	State       string
	CategoryID  int64
	RequesterID int64
	CreatedByID int64
	AssigneeID  *int64
}

// CreateInternalInput es lo que hace falta para insertar un interno, ya numerado.
type CreateInternalInput struct {
	TicketID         int64
	Number           string
	State            string
	EscalationReason string
	CreatedByID      int64
	AssigneeID       *int64
}

// SiguienteNumero consume el secuencial del año **en la misma transacción** que crea el ticket.
//
// Es una sola sentencia atómica a propósito: con `MAX(number) + 1` dos tickets creados a la vez
// sacarían el mismo número (docs/modules/tickets.md, sección 2.2). Devuelve el secuencial y el
// número ya compuesto con el prefijo.
func (r *TicketRepository) SiguienteNumero(tx *gorm.DB, year int, prefix string) (int, string, error) {
	var secuencia int

	err := tx.Raw(`
		INSERT INTO ticket_number_counters (year, last_number) VALUES (?, 1)
		ON CONFLICT (year) DO UPDATE SET last_number = ticket_number_counters.last_number + 1
		RETURNING last_number`, year).Scan(&secuencia).Error
	if err != nil {
		return 0, "", err
	}

	return secuencia, ComponerNumero(prefix, year, secuencia), nil
}

// ComponerNumero arma el número de un principal: `ACME-2026-0042`.
//
// Se rellena a cuatro dígitos pero **no se trunca**: a partir del 9999 crece a cinco y sigue siendo
// único y ordenable (decisión 6 del primer repaso).
func ComponerNumero(prefix string, year, seq int) string {
	return fmt.Sprintf("%s-%d-%04d", prefix, year, seq)
}

// NumeroDeInterno arma el número de un interno a partir del de su principal: `INT-ACME-2026-0042`.
func NumeroDeInterno(numeroDelPrincipal string) string { return "INT-" + numeroDelPrincipal }

// CreateTicket inserta un principal y deja su primera entrada del historial. Va en la transacción
// que abre quien numera, para que un ticket sin número no pueda existir.
func (r *TicketRepository) CreateTicket(tx *gorm.DB, entrada CreateTicketInput) (Ticket, error) {
	ticket := Ticket{
		Number:      entrada.Number,
		NumberYear:  entrada.NumberYear,
		NumberSeq:   entrada.NumberSeq,
		Subject:     entrada.Subject,
		Description: entrada.Description,
		State:       entrada.State,
		CategoryID:  entrada.CategoryID,
		RequesterID: entrada.RequesterID,
		CreatedByID: entrada.CreatedByID,
		AssigneeID:  entrada.AssigneeID,
	}

	if err := tx.Create(&ticket).Error; err != nil {
		return Ticket{}, err
	}

	return ticket, nil
}

// CreateInternal inserta el interno de un principal.
func (r *TicketRepository) CreateInternal(tx *gorm.DB, entrada CreateInternalInput) (InternalTicket, error) {
	interno := InternalTicket{
		TicketID:         entrada.TicketID,
		Number:           entrada.Number,
		State:            entrada.State,
		EscalationReason: entrada.EscalationReason,
		CreatedByID:      entrada.CreatedByID,
		AssigneeID:       entrada.AssigneeID,
	}

	if err := tx.Create(&interno).Error; err != nil {
		return InternalTicket{}, err
	}

	return interno, nil
}

// ByNumber busca un principal por su número. `number` es lo que la gente dice en voz alta.
func (r *TicketRepository) ByNumber(number string) (Ticket, error) {
	var ticket Ticket

	err := r.db.Where("number = ?", number).First(&ticket).Error
	if err != nil {
		return Ticket{}, traducir(err, ErrTicketNotFound)
	}

	return ticket, nil
}

// ByID busca un principal por su identificador interno.
func (r *TicketRepository) ByID(id int64) (Ticket, error) {
	var ticket Ticket

	err := r.db.Where("id = ?", id).First(&ticket).Error
	if err != nil {
		return Ticket{}, traducir(err, ErrTicketNotFound)
	}

	return ticket, nil
}

// InternalByNumber busca un interno por su número (`INT-ACME-2026-0042`).
func (r *TicketRepository) InternalByNumber(number string) (InternalTicket, error) {
	var interno InternalTicket

	err := r.db.Where("number = ?", number).First(&interno).Error
	if err != nil {
		return InternalTicket{}, traducir(err, ErrTicketNotFound)
	}

	return interno, nil
}

// InternalByTicket busca el interno de un principal. No existir **no es un error**: es un principal
// que todavía no se ha escalado.
func (r *TicketRepository) InternalByTicket(ticketID int64) (InternalTicket, bool, error) {
	var interno InternalTicket

	err := r.db.Where("ticket_id = ?", ticketID).First(&interno).Error
	if err == gorm.ErrRecordNotFound {
		return InternalTicket{}, false, nil
	}
	if err != nil {
		return InternalTicket{}, false, err
	}

	return interno, true, nil
}

// InternalsByTicket trae los internos de varios principales de una vez: es lo que evita preguntar
// uno a uno al pintar una lista.
func (r *TicketRepository) InternalsByTicket(ticketIDs []int64) (map[int64]InternalTicket, error) {
	if len(ticketIDs) == 0 {
		return map[int64]InternalTicket{}, nil
	}

	var internos []InternalTicket
	if err := r.db.Where("ticket_id IN ?", ticketIDs).Find(&internos).Error; err != nil {
		return nil, err
	}

	porPrincipal := make(map[int64]InternalTicket, len(internos))
	for _, interno := range internos {
		porPrincipal[interno.TicketID] = interno
	}

	return porPrincipal, nil
}

// UpdateTicket guarda los cambios de un principal.
func (r *TicketRepository) UpdateTicket(tx *gorm.DB, ticket Ticket) error {
	ticket.UpdatedAt = time.Now()
	return tx.Save(&ticket).Error
}

// UpdateInternal guarda los cambios de un interno.
func (r *TicketRepository) UpdateInternal(tx *gorm.DB, interno InternalTicket) error {
	interno.UpdatedAt = time.Now()
	return tx.Save(&interno).Error
}

// List devuelve una página de principales y cuántos hay en total.
func (r *TicketRepository) List(filtros Filtros) ([]Ticket, int64, error) {
	var total int64
	if err := conFiltrosDeTicket(r.db.Model(&Ticket{}), filtros).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// **Lo que lleva más tiempo esperando, primero**, y lo cerrado al final: un ticket cerrado no
	// estorba a lo que hay que atender (docs/interfaz-y-experiencia.md, sección 3.3).
	var tickets []Ticket
	err := conFiltrosDeTicket(r.db.Model(&Ticket{}), filtros).
		Order("CASE WHEN state = 'cerrado' THEN 1 ELSE 0 END ASC").
		Order("updated_at ASC").
		Limit(filtros.PerPage).
		Offset((filtros.Page - 1) * filtros.PerPage).
		Find(&tickets).Error
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// ListInternals devuelve una página de internos y su total.
func (r *TicketRepository) ListInternals(filtros Filtros) ([]InternalTicket, int64, error) {
	var total int64
	if err := conFiltrosDeInterno(r.db.Model(&InternalTicket{}), filtros).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var internos []InternalTicket
	err := conFiltrosDeInterno(r.db.Model(&InternalTicket{}), filtros).
		Select("internal_tickets.*").
		Order("CASE WHEN internal_tickets.state = 'cerrado' THEN 1 ELSE 0 END ASC").
		Order("internal_tickets.updated_at ASC").
		Limit(filtros.PerPage).
		Offset((filtros.Page - 1) * filtros.PerPage).
		Find(&internos).Error
	if err != nil {
		return nil, 0, err
	}

	return internos, total, nil
}

// UltimaAsignacion dice, por técnico, cuándo recibió su último ticket **de ese tipo**.
//
// Es el turno: quien hace más tiempo que no recibe uno —o no ha recibido ninguno, que no aparece en
// el mapa— va delante. Así no hay puntero de turno que guardar ni que desincronizar
// (docs/modules/tickets.md, sección 3.3).
func (r *TicketRepository) UltimaAsignacion(internos bool) (map[int64]time.Time, error) {
	tabla := "tickets"
	if internos {
		tabla = "internal_tickets"
	}

	type fila struct {
		AssigneeID int64
		Ultima     time.Time
	}

	var filas []fila
	err := r.db.Raw(`
		SELECT assignee_id, MAX(created_at) AS ultima
		FROM ` + tabla + `
		WHERE assignee_id IS NOT NULL
		GROUP BY assignee_id`).Scan(&filas).Error
	if err != nil {
		return nil, err
	}

	ultimas := make(map[int64]time.Time, len(filas))
	for _, f := range filas {
		ultimas[f.AssigneeID] = f.Ultima
	}

	return ultimas, nil
}

// conFiltrosDeTicket aplica lo que la lista de principales entiende por filtrar.
func conFiltrosDeTicket(consulta *gorm.DB, filtros Filtros) *gorm.DB {
	if filtros.State != "" {
		consulta = consulta.Where("state = ?", filtros.State)
	}

	if filtros.RequesterID != nil {
		consulta = consulta.Where("requester_id = ?", *filtros.RequesterID)
	}

	// **La categoría se filtra con su chip** (docs/modules/tickets.md, decisión 68), por identificador:
	// es lo que manda la pantalla al pulsarlo.
	if filtros.CategoryID != nil {
		consulta = consulta.Where("category_id = ?", *filtros.CategoryID)
	}

	// Y las etiquetas con los suyos. La comparación es exacta sobre el nombre normalizado del
	// catálogo: `red` y `red` son la misma, y `red-wifi` no es `red` (docs/modules/tickets.md,
	// decisión 72: el nombre vive en `ticket_tag_names`).
	if etiqueta := strings.TrimSpace(filtros.Tag); etiqueta != "" {
		consulta = consulta.Where(
			`EXISTS (SELECT 1 FROM ticket_tags
			          JOIN ticket_tag_names n ON n.id = ticket_tags.tag_id
			          WHERE ticket_tags.ticket_id = tickets.id
			            AND n.normalized = ?)`,
			etiqueta,
		)
	}

	// **«Lo mío»** (docs/modules/tickets.md, decisión 51): asignado, abierto por mí, comentado por mí
	// o **donde me han etiquetado** (ampliado el 2026-09-27). El `EXISTS` mira el hilo del principal,
	// que es la tabla de comentarios con `ticket_id`.
	//
	// `View` matiza ese «lo mío» sin salirse de él (decisión 62): el chip de «Mis tickets» puede
	// pedir sólo lo asignado o sólo lo que observo, y sin chip es todo. **Se resuelve en la consulta**,
	// que es donde se puede paginar de verdad: filtrar en memoria daría páginas de distinto tamaño.
	if filtros.Mine != nil {
		yo := *filtros.Mine

		switch filtros.View {
		case ViewAssigned:
			consulta = consulta.Where("assignee_id = ?", yo)

		case ViewWatching:
			consulta = consulta.Where(
				`EXISTS (SELECT 1 FROM ticket_observers
				          WHERE ticket_observers.ticket_id = tickets.id
				            AND ticket_observers.account_id = ?)`,
				yo,
			)

		default:
			consulta = consulta.Where(
				`requester_id = ? OR assignee_id = ?
				 OR EXISTS (SELECT 1 FROM ticket_comments
				            WHERE ticket_comments.ticket_id = tickets.id
				              AND ticket_comments.author_id = ?)
				 OR EXISTS (SELECT 1 FROM ticket_observers
				            WHERE ticket_observers.ticket_id = tickets.id
				              AND ticket_observers.account_id = ?)`,
				yo, yo, yo, yo,
			)
		}
	}

	if texto := strings.TrimSpace(filtros.Query); texto != "" {
		// El número se busca tal cual, el asunto se busca tal cual —es texto plano— y **la descripción
		// quitando antes sus etiquetas**: desde el 2026-09-26 la descripción es HTML, y buscar «p» o
		// «img» sobre el HTML crudo encontraría todos los tickets con una imagen
		// (docs/modules/tickets.md, sección 2.3). Se quitan con `regexp_replace` en la propia consulta,
		// que es donde se puede mirar todas las filas de una vez.
		//
		// **Y la búsqueda encuentra también la categoría y las etiquetas** (decisión 68): buscar «red»
		// saca los tickets de la categoría «Red» y los que llevan la etiqueta `red`. Son dos `EXISTS`
		// sobre las tablas del catálogo, que es como se pregunta sin duplicar filas.
		como := "%" + texto + "%"
		consulta = consulta.Where(
			`number ILIKE ? OR subject ILIKE ?
			 OR regexp_replace(description, '<[^>]*>', ' ', 'g') ILIKE ?
			 OR EXISTS (SELECT 1 FROM ticket_categories c
			            WHERE c.id = tickets.category_id AND c.name ILIKE ?)
			 OR EXISTS (SELECT 1 FROM ticket_tags tt
			            JOIN ticket_tag_names n ON n.id = tt.tag_id
			            WHERE tt.ticket_id = tickets.id AND n.tag ILIKE ?)`,
			como, como, como, como, como,
		)
	}

	return consulta
}

// conFiltrosDeInterno aplica lo mismo a los internos, que no tienen asunto ni descripción propios:
// su texto es el motivo del escalado, y lo demás se busca en su principal.
//
// **El principal va siempre en la consulta** porque la categoría y las etiquetas son suyas y el
// interno las hereda al leerse (docs/modules/tickets.md, decisión 67): buscar o filtrar por ellas es
// preguntar por lo del principal. La unión es uno a uno —la clave ajena lo garantiza—, así que no
// duplica filas.
//
// **Aquí no hay nada que quitar**: el motivo del escalado es texto plano y sigue siendo texto plano
// (docs/modules/tickets.md, sección 2.3), y la descripción del principal no se mira en esta consulta.
func conFiltrosDeInterno(consulta *gorm.DB, filtros Filtros) *gorm.DB {
	consulta = consulta.Joins("JOIN tickets ON tickets.id = internal_tickets.ticket_id")

	if filtros.State != "" {
		consulta = consulta.Where("internal_tickets.state = ?", filtros.State)
	}

	// **«Lo mío» en un interno** (docs/modules/tickets.md, decisión 51): asignado, **escalado por mí**
	// —un interno no tiene solicitante, quien lo abre es quien escala—, comentado por mí en su hilo o
	// **donde me han etiquetado** (ampliado el 2026-09-27). `View` matiza igual que en el principal.
	if filtros.Mine != nil {
		yo := *filtros.Mine

		switch filtros.View {
		case ViewAssigned:
			consulta = consulta.Where("internal_tickets.assignee_id = ?", yo)

		case ViewWatching:
			consulta = consulta.Where(
				`EXISTS (SELECT 1 FROM ticket_observers
				          WHERE ticket_observers.internal_ticket_id = internal_tickets.id
				            AND ticket_observers.account_id = ?)`,
				yo,
			)

		default:
			consulta = consulta.Where(
				`internal_tickets.created_by_id = ? OR internal_tickets.assignee_id = ?
				 OR EXISTS (SELECT 1 FROM ticket_comments
				            WHERE ticket_comments.internal_ticket_id = internal_tickets.id
				              AND ticket_comments.author_id = ?)
				 OR EXISTS (SELECT 1 FROM ticket_observers
				            WHERE ticket_observers.internal_ticket_id = internal_tickets.id
				              AND ticket_observers.account_id = ?)`,
				yo, yo, yo, yo,
			)
		}
	}

	// La categoría y la etiqueta del interno son **las del principal** (decisión 67).
	if filtros.CategoryID != nil {
		consulta = consulta.Where("tickets.category_id = ?", *filtros.CategoryID)
	}

	if etiqueta := strings.TrimSpace(filtros.Tag); etiqueta != "" {
		consulta = consulta.Where(
			`EXISTS (SELECT 1 FROM ticket_tags
			          JOIN ticket_tag_names n ON n.id = ticket_tags.tag_id
			          WHERE ticket_tags.ticket_id = internal_tickets.ticket_id
			            AND n.normalized = ?)`,
			etiqueta,
		)
	}

	if texto := strings.TrimSpace(filtros.Query); texto != "" {
		// Y la búsqueda mira también la categoría y las etiquetas del principal, igual que en la lista
		// de principales (decisión 68).
		como := "%" + texto + "%"
		consulta = consulta.Where(
			`internal_tickets.number ILIKE ? OR internal_tickets.escalation_reason ILIKE ?
			 OR tickets.number ILIKE ? OR tickets.subject ILIKE ?
			 OR EXISTS (SELECT 1 FROM ticket_categories c
			            WHERE c.id = tickets.category_id AND c.name ILIKE ?)
			 OR EXISTS (SELECT 1 FROM ticket_tags tt
			            JOIN ticket_tag_names n ON n.id = tt.tag_id
			            WHERE tt.ticket_id = internal_tickets.ticket_id AND n.tag ILIKE ?)`,
			como, como, como, como, como, como,
		)
	}

	return consulta
}

// traducir cambia «no hay filas» por el error del módulo, y deja pasar los demás tal cual.
func traducir(err error, cuandoNoHay error) error {
	if err == gorm.ErrRecordNotFound {
		return cuandoNoHay
	}

	return err
}
