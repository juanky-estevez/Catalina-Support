// Package dtos define lo que entra y lo que sale por la API del módulo tickets.
package dtos

import (
	"time"

	"catalina-support/backend/modules/tickets/services"
	"catalina-support/backend/shared/auth"
)

// CreateTicketRequest es lo que llega para dar de alta un ticket.
//
// `requesterId` sólo lo usa Soporte: es la forma de crear un ticket en nombre de quien llama por
// teléfono (docs/flujos.md, sección 3.2).
type CreateTicketRequest struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
	// **Todo ticket nace con una categoría** (docs/modules/tickets.md, decisión 65): `categoryId` es
	// obligatoria. `tags` es la lista de etiquetas tal y como se escriben: el backend las normaliza.
	CategoryID int64    `json:"categoryId"`
	Tags       []string `json:"tags"`
	// RequesterID y RequesterEmail son de Soporte: crear un ticket en nombre de otra persona. El
	// correo es lo que tiene a mano cuando alguien llama.
	RequesterID    *int64 `json:"requesterId"`
	RequesterEmail string `json:"requesterEmail"`
}

// UpdateTicketRequest cambia el asunto, la descripción, la categoría o las etiquetas. Los cuatro van
// con puntero para distinguir «no lo toques» de «déjalo así»: es un `PATCH`.
type UpdateTicketRequest struct {
	Subject     *string   `json:"subject"`
	Description *string   `json:"description"`
	CategoryID  *int64    `json:"categoryId"`
	Tags        *[]string `json:"tags"`
}

// AssignRequest pone responsable.
type AssignRequest struct {
	AssigneeID int64 `json:"assigneeId"`
}

// StateRequest mueve el estado.
//
// El estado llega en `state`, que es lo que manda la pantalla, y **también se admite `to`**, que es el
// nombre que fija la tabla de endpoints del documento (docs/modules/tickets.md, sección 5): así valen
// los dos y ninguno se rompe. `comment` es el comentario del movimiento: **obligatorio al cerrar**
// (decisión 81) y opcional en los demás estados, y queda como un comentario de la conversación.
type StateRequest struct {
	State   string `json:"state"`
	To      string `json:"to"`
	Comment string `json:"comment"`
}

// EscalateRequest escala el ticket: el motivo es obligatorio.
type EscalateRequest struct {
	Reason string `json:"reason"`
}

// CommentRequest escribe un comentario.
type CommentRequest struct {
	Body string `json:"body"`
}

type ImproveWritingRequest struct {
	Editor string `json:"editor"`
	Draft  string `json:"draft"`
	Tone   string `json:"tone"`
}

type ImproveWritingResponse struct {
	Text string `json:"text"`
}

// AssigneesResponse son las personas que pueden ser responsables, por tipo de ticket.
type AssigneesResponse struct {
	Main     []PersonResponse `json:"main"`
	Internal []PersonResponse `json:"internal"`
}

// NewAssigneesResponse traduce las dos listas.
func NewAssigneesResponse(porTipo map[string][]auth.Account) AssigneesResponse {
	convertir := func(cuentas []auth.Account) []PersonResponse {
		personas := make([]PersonResponse, 0, len(cuentas))
		for i := range cuentas {
			personas = append(personas, *NewPersonResponse(&cuentas[i]))
		}

		return personas
	}

	return AssigneesResponse{
		Main:     convertir(porTipo["principal"]),
		Internal: convertir(porTipo["interno"]),
	}
}

// PersonResponse es una cuenta reducida a lo que se enseña dentro de un ticket.
//
// No lleva el correo de nadie más que del solicitante y el responsable: quien escribe en un ticket no
// tiene por qué repartir su dirección por ahí (docs/arquitectura.md, regla de logs y datos).
type PersonResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role"`
}

// TicketResponse es un ticket tal y como lo ve el frontend.
type TicketResponse struct {
	// Number es lo que se dice en voz alta y lo que va en los correos. El `id` interno no sale de la
	// base (docs/modules/tickets.md, sección 5).
	Number   string `json:"number"`
	Internal bool   `json:"internal"`

	Subject     string `json:"subject"`
	Description string `json:"description"`
	State       string `json:"state"`

	// Reason es el motivo del escalado, en los internos. Parent es el número del principal, y Child el
	// del interno: es lo que permite saltar de uno a otro en la vista doble.
	Reason string `json:"reason,omitempty"`
	Parent string `json:"parent,omitempty"`
	Child  string `json:"child,omitempty"`

	// **La categoría y las etiquetas viajan en el ticket** (docs/modules/tickets.md, decisión 67): en
	// las listas y en la ficha, y en un interno son las de su principal, que es lo que hereda.
	Category *CategoryResponse `json:"category,omitempty"`
	Tags     []string          `json:"tags"`

	Requester *PersonResponse `json:"requester,omitempty"`
	CreatedBy *PersonResponse `json:"createdBy,omitempty"`
	Assignee  *PersonResponse `json:"assignee,omitempty"`

	SubjectEditedAt     string `json:"subjectEditedAt,omitempty"`
	DescriptionEditedAt string `json:"descriptionEditedAt,omitempty"`
	ResolvedAt          string `json:"resolvedAt,omitempty"`
	ClosedAt            string `json:"closedAt,omitempty"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`

	// Insights son los dos campos que redacta el motor de IA: el motivo y la última acción, con sus
	// dos idiomas y su estado (docs/modules/ai.md). Viajan siempre, también vacíos: la pantalla necesita
	// saber si están «pendientes», si el motor no está o si fallaron, y eso no se distingue de un campo
	// que no viene.
	Insights InsightsResponse `json:"insights"`
}

// CommentResponse es un comentario de la conversación.
type CommentResponse struct {
	ID        int64           `json:"id"`
	Author    *PersonResponse `json:"author,omitempty"`
	Body      string          `json:"body"`
	Edited    bool            `json:"edited"`
	Deleted   bool            `json:"deleted"`
	CreatedAt string          `json:"createdAt"`
}

// AttachmentResponse es un adjunto: lo que hace falta para enseñarlo y para descargarlo.
type AttachmentResponse struct {
	ID          int64           `json:"id"`
	Filename    string          `json:"filename"`
	ContentType string          `json:"contentType"`
	Size        int64           `json:"size"`
	CommentID   *int64          `json:"commentId,omitempty"`
	UploadedBy  *PersonResponse `json:"uploadedBy,omitempty"`
	CreatedAt   string          `json:"createdAt"`
}

// HistoryResponse es lo que hizo el sistema, para la línea de tiempo.
type HistoryResponse struct {
	ID        int64  `json:"id"`
	Event     string `json:"event"`
	FromState string `json:"fromState,omitempty"`
	ToState   string `json:"toState,omitempty"`
	Detail    string `json:"detail,omitempty"`
	// Actor nulo es «lo hizo el sistema».
	Actor     *PersonResponse `json:"actor,omitempty"`
	CreatedAt string          `json:"createdAt"`
}

// ObserverResponse es un observador del ticket: quién lo sigue y quién lo añadió.
//
// **El `id` es el de la cuenta**, no el de la fila de `ticket_observers`: quien se quita del ticket es
// una persona, y la ruta `DELETE /api/tickets/{number}/observers/{id}` habla de esa misma persona. La
// fila es cosa de dentro y no sale de la base.
type ObserverResponse struct {
	ID        int64           `json:"id"`
	Account   *PersonResponse `json:"account,omitempty"`
	AddedBy   *PersonResponse `json:"addedBy,omitempty"`
	CreatedAt string          `json:"createdAt"`
}

// AddObserverRequest es lo que llega para **añadir a mano** un observador: la cuenta que pasa a seguir
// el ticket (docs/modules/tickets.md, sección 2.3.1 y decisión 74). Va por identificador y no por
// correo, que es lo que manda la ficha, y la cuenta tiene que ser un técnico o un desarrollador
// activo: la misma regla que etiquetar (decisión 59).
type AddObserverRequest struct {
	AccountID int64 `json:"accountId"`
}

// ObserverEnvelope es la respuesta de **añadir un observador a mano**: el ticket actualizado con su
// lista de observadores, que es lo que la pantalla vuelve a pintar sin pedir nada más
// (docs/modules/tickets.md, sección 5 y decisión 74).
type ObserverEnvelope struct {
	Ticket ObserverTicketResponse `json:"ticket"`
}

// ObserverTicketResponse es el ticket con sus observadores: **los mismos campos que en la ficha** más
// la lista. No se le añade el campo a `TicketResponse` para no ensuciar las listas paginadas, que no
// llevan observadores (docs/modules/tickets.md, sección 5).
type ObserverTicketResponse struct {
	TicketResponse
	Observers []ObserverResponse `json:"observers"`
}

// TicketEnvelope es la respuesta de un ticket y de las acciones sobre él.
type TicketEnvelope struct {
	Ticket TicketResponse `json:"ticket"`
}

// TicketDetailResponse es un ticket con todo lo que cuelga de él.
type TicketDetailResponse struct {
	Ticket      TicketResponse       `json:"ticket"`
	Comments    []CommentResponse    `json:"comments"`
	Attachments []AttachmentResponse `json:"attachments"`
	History     []HistoryResponse    `json:"history"`
	// Observers sólo viaja en la ficha, no en las listas paginadas (docs/modules/tickets.md,
	// sección 5).
	Observers    []ObserverResponse   `json:"observers"`
	Capabilities CapabilitiesResponse `json:"capabilities"`
}

type CapabilitiesResponse struct {
	AIWriting bool `json:"aiWriting"`
}

// TicketsResponse es la bandeja, con lo que necesita la paginación.
type TicketsResponse struct {
	Tickets []TicketResponse `json:"tickets"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	PerPage int              `json:"perPage"`
}

// CategoryRequest es el cuerpo de crear y de renombrar una categoría.
type CategoryRequest struct {
	Name string `json:"name"`
}

// CategoryStateRequest retira o vuelve a poner una categoría: `active` a `false` la retira.
type CategoryStateRequest struct {
	Active bool `json:"active"`
}

// CategoryResponse es una categoría tal y como viaja dentro de un ticket.
type CategoryResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// CategoryCatalogResponse es una categoría del catálogo, con cuántos tickets la usan: es la cuenta
// que enseña que retirar una categoría no vacía los tickets que la tienen (decisión 66).
type CategoryCatalogResponse struct {
	CategoryResponse
	Tickets int64 `json:"tickets"`
}

// CategoriesResponse es el catálogo entero.
type CategoriesResponse struct {
	Categories []CategoryCatalogResponse `json:"categories"`
}

// CategoryEnvelope es la respuesta de crear, renombrar o retirar una categoría.
type CategoryEnvelope struct {
	Category CategoryCatalogResponse `json:"category"`
}

// TagResponse es una etiqueta con cuántos tickets la llevan: lo que se sugiere al escribir y lo que
// devuelve el catálogo al crearla o renombrarla. **Cero es un número válido**: una etiqueta puede
// existir sin que ningún ticket la lleve (docs/modules/tickets.md, decisión 72).
type TagResponse struct {
	Tag     string `json:"tag"`
	Tickets int64  `json:"tickets"`
}

// TagRequest es el cuerpo de crear y de renombrar una etiqueta.
type TagRequest struct {
	Tag string `json:"tag"`
}

// TagsResponse son las etiquetas del catálogo.
type TagsResponse struct {
	Tags []TagResponse `json:"tags"`
}

// TagEnvelope es la respuesta de crear o renombrar una etiqueta.
type TagEnvelope struct {
	Tag TagResponse `json:"tag"`
}

// CommentEnvelope es la respuesta de escribir o editar un comentario.
type CommentEnvelope struct {
	Comment CommentResponse `json:"comment"`
}

// AttachmentEnvelope es la respuesta de subir un adjunto.
type AttachmentEnvelope struct {
	Attachment AttachmentResponse `json:"attachment"`
}

// NewPersonResponse traduce una cuenta a lo que se enseña.
func NewPersonResponse(cuenta *auth.Account) *PersonResponse {
	if cuenta == nil {
		return nil
	}

	return &PersonResponse{
		ID:       cuenta.ID,
		Name:     cuenta.Name,
		LastName: cuenta.LastName,
		Email:    cuenta.Email,
		Role:     cuenta.Role,
	}
}

// NewTicketResponse traduce un ticket.
// InsightsResponse son los dos resúmenes del motor de IA, con lo que hace falta para pintarlos.
type InsightsResponse struct {
	Motivo       ResumenResponse `json:"motivo"`
	UltimaAccion ResumenResponse `json:"ultimaAccion"`
}

// ResumenResponse es un texto del motor en el idioma global y la clave del error.
type ResumenResponse struct {
	State    string `json:"state"`
	Text     string `json:"text,omitempty"`
	Language string `json:"language,omitempty"`
	ErrorKey string `json:"errorKey,omitempty"`
}

// NewInsightsResponse traduce los dos campos de un ticket.
func NewInsightsResponse(resumen services.Resumenes) InsightsResponse {
	return InsightsResponse{
		Motivo:       NewResumenResponse(resumen.Motivo),
		UltimaAccion: NewResumenResponse(resumen.UltimaAccion),
	}
}

// NewResumenResponse traduce un resumen.
func NewResumenResponse(resumen services.Resumen) ResumenResponse {
	return ResumenResponse{
		State:    resumen.Estado,
		Text:     resumen.Text,
		Language: resumen.Language,
		ErrorKey: resumen.ErrorKey,
	}
}

func NewTicketResponse(ticket services.Ticket) TicketResponse {
	return TicketResponse{
		Number:              ticket.Number,
		Internal:            ticket.Internal,
		Subject:             ticket.Subject,
		Description:         ticket.Description,
		State:               ticket.State,
		Reason:              ticket.Reason,
		Parent:              ticket.Parent,
		Child:               ticket.Child,
		Category:            NewCategoryResponse(ticket.Category),
		Tags:                etiquetas(ticket.Tags),
		Requester:           NewPersonResponse(ticket.Requester),
		CreatedBy:           NewPersonResponse(ticket.CreatedBy),
		Assignee:            NewPersonResponse(ticket.Assignee),
		SubjectEditedAt:     fecha(ticket.SubjectEditedAt),
		DescriptionEditedAt: fecha(ticket.DescriptionEditedAt),
		ResolvedAt:          fecha(ticket.ResolvedAt),
		ClosedAt:            fecha(ticket.ClosedAt),
		CreatedAt:           ticket.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           ticket.UpdatedAt.UTC().Format(time.RFC3339),
		Insights:            NewInsightsResponse(ticket.Resumen),
	}
}

// NewCommentResponse traduce un comentario.
func NewCommentResponse(comentario services.Comment) CommentResponse {
	return CommentResponse{
		ID:        comentario.ID,
		Author:    NewPersonResponse(comentario.Author),
		Body:      comentario.Body,
		Edited:    comentario.Edited,
		Deleted:   comentario.Deleted,
		CreatedAt: comentario.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// NewAttachmentResponse traduce un adjunto.
func NewAttachmentResponse(adjunto services.Attachment) AttachmentResponse {
	return AttachmentResponse{
		ID:          adjunto.ID,
		Filename:    adjunto.Filename,
		ContentType: adjunto.ContentType,
		Size:        adjunto.Size,
		CommentID:   adjunto.CommentID,
		UploadedBy:  NewPersonResponse(adjunto.UploadedBy),
		CreatedAt:   adjunto.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// NewHistoryResponse traduce una entrada del historial.
func NewHistoryResponse(entrada services.HistoryEntry) HistoryResponse {
	return HistoryResponse{
		ID:        entrada.ID,
		Event:     entrada.Event,
		FromState: entrada.FromState,
		ToState:   entrada.ToState,
		Detail:    entrada.Detail,
		Actor:     NewPersonResponse(entrada.Actor),
		CreatedAt: entrada.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// NewTicketDetailResponse traduce un ticket con todo lo que cuelga de él.
func NewTicketDetailResponse(detalle services.Detail) TicketDetailResponse {
	respuesta := TicketDetailResponse{
		Ticket:       NewTicketResponse(detalle.Ticket),
		Comments:     make([]CommentResponse, 0, len(detalle.Comments)),
		Attachments:  make([]AttachmentResponse, 0, len(detalle.Attachments)),
		History:      make([]HistoryResponse, 0, len(detalle.History)),
		Observers:    make([]ObserverResponse, 0, len(detalle.Observers)),
		Capabilities: CapabilitiesResponse{AIWriting: detalle.Capabilities.AIWriting},
	}

	for _, comentario := range detalle.Comments {
		respuesta.Comments = append(respuesta.Comments, NewCommentResponse(comentario))
	}
	for _, adjunto := range detalle.Attachments {
		respuesta.Attachments = append(respuesta.Attachments, NewAttachmentResponse(adjunto))
	}
	for _, entrada := range detalle.History {
		respuesta.History = append(respuesta.History, NewHistoryResponse(entrada))
	}
	for _, observador := range detalle.Observers {
		respuesta.Observers = append(respuesta.Observers, NewObserverResponse(observador))
	}

	return respuesta
}

// NewObserverTicketResponse traduce el ticket de la respuesta de añadir un observador: el ticket como
// en la ficha y, con él, la lista de observadores ya puesta (docs/modules/tickets.md, sección 5).
func NewObserverTicketResponse(ticket services.Ticket, observadores []services.Observer) ObserverTicketResponse {
	respuesta := ObserverTicketResponse{
		TicketResponse: NewTicketResponse(ticket),
		Observers:      make([]ObserverResponse, 0, len(observadores)),
	}

	for _, observador := range observadores {
		respuesta.Observers = append(respuesta.Observers, NewObserverResponse(observador))
	}

	return respuesta
}

// NewObserverResponse traduce un observador. El `id` que sale es el de la **cuenta**, que es con el
// que se quita (docs/modules/tickets.md, sección 5).
func NewObserverResponse(observador services.Observer) ObserverResponse {
	respuesta := ObserverResponse{
		AddedBy:   newObserverPersonResponse(observador.AddedBy),
		CreatedAt: observador.CreatedAt.UTC().Format(time.RFC3339),
	}

	if observador.Account != nil {
		respuesta.ID = observador.Account.ID
		respuesta.Account = newObserverPersonResponse(observador.Account)
	}

	return respuesta
}

// newObserverPersonResponse es una cuenta sin correo: quien escribe en un ticket no tiene por qué
// repartir su dirección por ahí, y la ficha de un observador sólo necesita el nombre y el papel.
func newObserverPersonResponse(cuenta *auth.Account) *PersonResponse {
	if cuenta == nil {
		return nil
	}

	return &PersonResponse{
		ID:       cuenta.ID,
		Name:     cuenta.Name,
		LastName: cuenta.LastName,
		Role:     cuenta.Role,
	}
}

// NewTicketsResponse traduce una página de la bandeja.
func NewTicketsResponse(pagina services.Page) TicketsResponse {
	tickets := make([]TicketResponse, 0, len(pagina.Tickets))
	for _, ticket := range pagina.Tickets {
		tickets = append(tickets, NewTicketResponse(ticket))
	}

	return TicketsResponse{
		Tickets: tickets,
		Total:   pagina.Total,
		Page:    pagina.Page,
		PerPage: pagina.PerPage,
	}
}

// fecha escribe una fecha en ISO, o nada si no la hay.
func fecha(momento *time.Time) string {
	if momento == nil {
		return ""
	}

	return momento.UTC().Format(time.RFC3339)
}

// NewCategoryResponse traduce una categoría para dentro de un ticket.
func NewCategoryResponse(categoria *services.Category) *CategoryResponse {
	if categoria == nil {
		return nil
	}

	return &CategoryResponse{ID: categoria.ID, Name: categoria.Name, Active: categoria.Active}
}

// NewCategoryCatalogResponse traduce una categoría del catálogo con su cuenta de tickets.
func NewCategoryCatalogResponse(categoria services.CategoryWithCount) CategoryCatalogResponse {
	return CategoryCatalogResponse{
		CategoryResponse: CategoryResponse{
			ID:     categoria.ID,
			Name:   categoria.Name,
			Active: categoria.Active,
		},
		Tickets: categoria.Tickets,
	}
}

// NewCategoriesResponse traduce el catálogo entero.
func NewCategoriesResponse(categorias []services.CategoryWithCount) CategoriesResponse {
	respuesta := CategoriesResponse{Categories: make([]CategoryCatalogResponse, 0, len(categorias))}
	for _, categoria := range categorias {
		respuesta.Categories = append(respuesta.Categories, NewCategoryCatalogResponse(categoria))
	}

	return respuesta
}

// NewTagsResponse traduce las etiquetas del catálogo.
func NewTagsResponse(etiquetas []services.TagCount) TagsResponse {
	respuesta := TagsResponse{Tags: make([]TagResponse, 0, len(etiquetas))}
	for _, etiqueta := range etiquetas {
		respuesta.Tags = append(respuesta.Tags, TagResponse{Tag: etiqueta.Tag, Tickets: etiqueta.Tickets})
	}

	return respuesta
}

// NewTagResponse traduce una etiqueta del catálogo, con su cuenta de tickets.
func NewTagResponse(etiqueta services.TagCount) TagResponse {
	return TagResponse{Tag: etiqueta.Tag, Tickets: etiqueta.Tickets}
}

// etiquetas deja la lista como un array JSON: vacío es `[]` y no `null`, para que la pantalla no
// tenga que distinguir «no hay ninguna» de «no ha llegado».
func etiquetas(tags []string) []string {
	if tags == nil {
		return []string{}
	}

	return tags
}
