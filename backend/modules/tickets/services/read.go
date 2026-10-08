package services

import (
	"sort"
	"strings"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// Los tipos de ticket, tal y como los entiende el filtro de la bandeja.
const (
	TypePrincipal = "principal"
	TypeInternal  = "interno"
)

// List devuelve la bandeja: lo que le toca ver a quien la pide.
//
// El filtro de tipo llega de la pantalla —el Administrador tiene su chip, y Soporte y Desarrollo ven
// el suyo— y **para un usuario no hay elección**: ve sus principales y nada más
// (docs/usuarios-y-permisos.md, regla 1 de la sección 3).
//
// `View` es el chip de «Mis tickets» (docs/modules/tickets.md, decisión 62): `assigned`, `watching` o
// vacío. **Sólo matiza «lo mío»**, así que sin `Mine` no hace nada: «Tickets principales» y «Tickets
// internos» son la lista de todo y se piden sin él.
func (s *Service) List(filtros repositories.Filtros, actor auth.Identity) (Page, error) {
	filtros = NormalizarFiltros(filtros)

	// **La etiqueta del filtro se normaliza como se normaliza al guardar** (decisión 68): el chip
	// manda lo que se escribió, y `Red Wifi` tiene que encontrar `red-wifi`. Si no queda nada —sólo
	// signos—, no hay etiqueta que buscar y el filtro no limita.
	if etiqueta := strings.TrimSpace(filtros.Tag); etiqueta != "" {
		if normalizada, err := NormalizarEtiqueta(etiqueta); err == nil {
			filtros.Tag = normalizada
		} else {
			filtros.Tag = ""
		}
	}

	if actor.Role == auth.RoleUsuario {
		filtros.Type = TypePrincipal
		yo := actor.ID
		filtros.RequesterID = &yo
		// Para un usuario, «lo mío» es exactamente lo que solicitó: el chip de asignados no cambia nada
		// y no observa nada —no etiqueta—, así que la vista no tiene por dónde matizar. Lo que sí manda
		// es el `RequesterID` de arriba, que le deja ver sólo lo suyo.
		filtros.Mine = nil
		filtros.View = ""
	}

	switch filtros.Type {
	case TypePrincipal:
		return s.listaDePrincipales(filtros)
	case TypeInternal:
		return s.listaDeInternos(filtros)
	default:
		return s.listaDeLosDos(filtros)
	}
}

// ByNumber devuelve la ficha de un ticket por su número, con su conversación, sus adjuntos y su
// historial.
//
// **El número manda**: `INT-` delante quiere decir interno, y cualquier otro, principal. El interno se
// lee con este mismo endpoint porque es un ticket (docs/modules/tickets.md, sección 5).
func (s *Service) ByNumber(number string, actor auth.Identity) (Detail, error) {
	numero := strings.TrimSpace(number)
	var detalle Detail
	var err error
	if strings.HasPrefix(numero, "INT-") {
		detalle, err = s.fichaDeInterno(numero, actor)
	} else {
		detalle, err = s.fichaDePrincipal(numero, actor)
	}
	if err != nil {
		return Detail{}, err
	}
	detalle.Capabilities.AIWriting = s.capacidadDeRedaccion(actor, detalle.Ticket)
	return detalle, nil
}

// listaDePrincipales arma una página de principales.
func (s *Service) listaDePrincipales(filtros repositories.Filtros) (Page, error) {
	tickets, total, err := s.tickets.List(filtros)
	if err != nil {
		return Page{}, err
	}

	cuentas := make([]Ticket, 0, len(tickets))
	ids := make([]int64, 0, len(tickets))

	internos, err := s.tickets.InternalsByTicket(idsDe(tickets))
	if err != nil {
		return Page{}, err
	}

	// **La categoría y las etiquetas viajan en el renglón** (decisión 68): se leen en bloque para toda
	// la página, no una consulta por fila.
	categorias, err := s.categoriasDe(idsDeCategorias(tickets))
	if err != nil {
		return Page{}, err
	}

	etiquetas, err := s.categories.TagsByTickets(idsDe(tickets))
	if err != nil {
		return Page{}, err
	}

	for _, ticket := range tickets {
		ids = append(ids, ticket.RequesterID, ticket.CreatedByID)
		if ticket.AssigneeID != nil {
			ids = append(ids, *ticket.AssigneeID)
		}
	}

	porID, err := s.accounts.ByIDs(ids)
	if err != nil {
		return Page{}, err
	}

	for _, ticket := range tickets {
		convertido := s.aTicket(ticket, porID, categorias[ticket.CategoryID], etiquetas[ticket.ID])
		if interno, hay := internos[ticket.ID]; hay {
			convertido.Child = interno.Number
		}
		cuentas = append(cuentas, convertido)
	}

	// Los dos campos del motor de IA se pegan aquí, **de una vez para toda la página**: una consulta
	// para la lista entera, no una por fila (docs/modules/ai.md, decisión 9).
	s.pegarResumenes(cuentas)

	return Page{Tickets: cuentas, Total: total, Page: filtros.Page, PerPage: filtros.PerPage}, nil
}

// listaDeInternos arma una página de internos, cada uno con la referencia de su principal.
func (s *Service) listaDeInternos(filtros repositories.Filtros) (Page, error) {
	internos, total, err := s.tickets.ListInternals(filtros)
	if err != nil {
		return Page{}, err
	}

	principales, err := s.principalesDe(internos)
	if err != nil {
		return Page{}, err
	}

	// **El interno hereda la categoría y las etiquetas de su principal** (decisión 67), así que se leen
	// las del principal y se pegan igual que el asunto.
	categorias, err := s.categoriasDe(idsDeCategoriasDePrincipales(internos, principales))
	if err != nil {
		return Page{}, err
	}

	etiquetas, err := s.categories.TagsByTickets(idsDePrincipales(internos))
	if err != nil {
		return Page{}, err
	}

	ids := make([]int64, 0, len(internos))
	for _, interno := range internos {
		ids = append(ids, interno.CreatedByID)
		if interno.AssigneeID != nil {
			ids = append(ids, *interno.AssigneeID)
		}
	}

	porID, err := s.accounts.ByIDs(ids)
	if err != nil {
		return Page{}, err
	}

	tickets := make([]Ticket, 0, len(internos))
	for _, interno := range internos {
		principal := principales[interno.TicketID]
		tickets = append(tickets, s.aTicketInterno(
			interno,
			principal,
			porID,
			categorias[principal.CategoryID],
			etiquetas[principal.ID],
		))
	}

	// **Los internos también llevan sus dos campos**: el «Motivo» de un interno es el del escalado —lo
	// pinta la pantalla, no viene de aquí— y la «Última acción» la redacta el motor para él, con su
	// propio número. Sin esta línea, las dos columnas salían vacías en la lista de internos.
	s.pegarResumenes(tickets)

	return Page{Tickets: tickets, Total: total, Page: filtros.Page, PerPage: filtros.PerPage}, nil
}

// listaDeLosDos mezcla las dos listas y pagina el resultado.
//
// Para servir la página N basta con pedir N páginas de cada lado: lo que salga en la mezcla está, como
// mucho, en esas N páginas de su propia lista. Así «Todos» es correcto sin traerse las dos tablas
// enteras.
func (s *Service) listaDeLosDos(filtros repositories.Filtros) (Page, error) {
	hasta := filtros.Page * filtros.PerPage

	dePrincipales := filtros
	dePrincipales.Type = TypePrincipal
	dePrincipales.Page = 1
	dePrincipales.PerPage = hasta

	deInternos := filtros
	deInternos.Type = TypeInternal
	deInternos.Page = 1
	deInternos.PerPage = hasta

	principales, err := s.listaDePrincipales(dePrincipales)
	if err != nil {
		return Page{}, err
	}

	internos, err := s.listaDeInternos(deInternos)
	if err != nil {
		return Page{}, err
	}

	todos := append(append([]Ticket{}, principales.Tickets...), internos.Tickets...)

	// El mismo orden que la base: lo cerrado al final, y dentro de eso lo que lleva más tiempo
	// esperando, primero.
	sort.SliceStable(todos, func(i, j int) bool {
		cerradoI := todos[i].State == repositories.StateCerrado
		cerradoJ := todos[j].State == repositories.StateCerrado

		if cerradoI != cerradoJ {
			return !cerradoI
		}

		return todos[i].UpdatedAt.Before(todos[j].UpdatedAt)
	})

	desde := (filtros.Page - 1) * filtros.PerPage
	siHasta := desde + filtros.PerPage
	if desde > len(todos) {
		desde = len(todos)
	}
	if siHasta > len(todos) {
		siHasta = len(todos)
	}

	return Page{
		Tickets: todos[desde:siHasta],
		Total:   principales.Total + internos.Total,
		Page:    filtros.Page,
		PerPage: filtros.PerPage,
	}, nil
}

// fichaDePrincipal arma la ficha de un principal, con la referencia de su interno si lo tiene.
func (s *Service) fichaDePrincipal(number string, actor auth.Identity) (Detail, error) {
	ticket, err := s.tickets.ByNumber(number)
	if err != nil {
		return Detail{}, traducirTicket(err)
	}

	detalle, err := s.detalleDePrincipal(ticket, actor)
	if err != nil {
		return Detail{}, err
	}

	return detalle, nil
}

// fichaDeInterno arma la ficha de un interno, con la referencia de su principal.
func (s *Service) fichaDeInterno(number string, actor auth.Identity) (Detail, error) {
	interno, err := s.tickets.InternalByNumber(number)
	if err != nil {
		return Detail{}, traducirTicket(err)
	}

	// **El usuario nunca ve el interno**: para él, ese ticket no existe
	// (docs/usuarios-y-permisos.md, regla 1 de la sección 3).
	if actor.Role == auth.RoleUsuario {
		return Detail{}, ErrTicketNotFound
	}

	principal, err := s.tickets.ByID(interno.TicketID)
	if err != nil {
		return Detail{}, traducirTicket(err)
	}

	detalle, err := s.detalleDeInterno(interno, principal, actor)
	if err != nil {
		return Detail{}, err
	}

	return detalle, nil
}

// detalleDePrincipal arma un principal con todo lo que cuelga de él.
func (s *Service) detalleDePrincipal(ticket repositories.Ticket, actor auth.Identity) (Detail, error) {
	destino := repositories.DestinoDePrincipal(ticket.ID)

	comentarios, adjuntos, historial, err := s.conversacionDe(destino)
	if err != nil {
		return Detail{}, err
	}

	// **La lista de observadores viaja en la ficha**, no en las listas paginadas: ahí sería una
	// consulta por página para un dato que no se usa (docs/modules/tickets.md, sección 5).
	observadores, err := s.conversation.Observers(s.tickets.DB(), destino)
	if err != nil {
		return Detail{}, err
	}

	ids := []int64{ticket.RequesterID, ticket.CreatedByID}
	if ticket.AssigneeID != nil {
		ids = append(ids, *ticket.AssigneeID)
	}
	ids = append(ids, idsDeComentarios(comentarios)...)
	ids = append(ids, idsDeAdjuntos(adjuntos)...)
	ids = append(ids, idsDeHistorial(historial)...)
	ids = append(ids, idsDeObservadores(observadores)...)

	porID, err := s.accounts.ByIDs(ids)
	if err != nil {
		return Detail{}, err
	}

	categorias, err := s.categoriasDe([]int64{ticket.CategoryID})
	if err != nil {
		return Detail{}, err
	}

	etiquetas, err := s.categories.TagsByTicket(ticket.ID)
	if err != nil {
		return Detail{}, err
	}

	convertido := s.aTicket(ticket, porID, categorias[ticket.CategoryID], etiquetas)

	// Sólo se enseña lo que esa persona puede ver: para un usuario, el principal es todo lo que hay.
	if !puedeVer(actor, convertido) {
		return Detail{}, ErrTicketNotFound
	}

	detalle := Detail{
		Ticket:      convertido,
		Comments:    s.aComentarios(comentarios, porID),
		Attachments: s.aAdjuntos(adjuntos, porID),
		History:     s.aHistorial(historial, porID),
		Observers:   s.aObservadores(observadores, porID),
	}

	if interno, hay, err := s.tickets.InternalByTicket(ticket.ID); err != nil {
		return Detail{}, err
	} else if hay && actor.Role != auth.RoleUsuario {
		detalle.Ticket.Child = interno.Number
	}

	// Los dos campos del motor de IA, para la ficha: la misma consulta en bloque que usan las listas,
	// con un solo número.
	temporal := []Ticket{detalle.Ticket}
	s.pegarResumenes(temporal)
	detalle.Ticket.Resumen = temporal[0].Resumen

	return detalle, nil
}

// detalleDeInterno arma un interno con su conversación y la referencia de su principal.
func (s *Service) detalleDeInterno(interno repositories.InternalTicket, principal repositories.Ticket, actor auth.Identity) (Detail, error) {
	destino := repositories.DestinoDeInterno(interno.ID)

	comentarios, adjuntos, historial, err := s.conversacionDe(destino)
	if err != nil {
		return Detail{}, err
	}

	observadores, err := s.conversation.Observers(s.tickets.DB(), destino)
	if err != nil {
		return Detail{}, err
	}

	ids := []int64{principal.RequesterID, principal.CreatedByID, interno.CreatedByID}
	if interno.AssigneeID != nil {
		ids = append(ids, *interno.AssigneeID)
	}
	ids = append(ids, idsDeComentarios(comentarios)...)
	ids = append(ids, idsDeAdjuntos(adjuntos)...)
	ids = append(ids, idsDeHistorial(historial)...)
	ids = append(ids, idsDeObservadores(observadores)...)

	porID, err := s.accounts.ByIDs(ids)
	if err != nil {
		return Detail{}, err
	}

	// **La categoría y las etiquetas del interno son las del principal** (decisión 67): clasificar el
	// mismo caso dos veces sería pedir trabajo de más.
	categorias, err := s.categoriasDe([]int64{principal.CategoryID})
	if err != nil {
		return Detail{}, err
	}

	etiquetas, err := s.categories.TagsByTicket(principal.ID)
	if err != nil {
		return Detail{}, err
	}

	convertido := s.aTicketInterno(interno, principal, porID, categorias[principal.CategoryID], etiquetas)

	// Y la ficha del interno, lo mismo: la última acción va con él (decisión 5).
	temporal := []Ticket{convertido}
	s.pegarResumenes(temporal)
	convertido = temporal[0]

	return Detail{
		Ticket:      convertido,
		Comments:    s.aComentarios(comentarios, porID),
		Attachments: s.aAdjuntos(adjuntos, porID),
		History:     s.aHistorial(historial, porID),
		Observers:   s.aObservadores(observadores, porID),
	}, nil
}

// conversacionDe lee las tres cosas que cuelgan de un ticket de una vez.
func (s *Service) conversacionDe(destino repositories.Destino) ([]repositories.Comment, []repositories.Attachment, []repositories.HistoryEntry, error) {
	comentarios, err := s.conversation.Comments(destino)
	if err != nil {
		return nil, nil, nil, err
	}

	adjuntos, err := s.conversation.Attachments(destino)
	if err != nil {
		return nil, nil, nil, err
	}

	historial, err := s.conversation.History(destino)
	if err != nil {
		return nil, nil, nil, err
	}

	return comentarios, adjuntos, historial, nil
}

// principalesDe trae los principales de varios internos, para no preguntar uno a uno.
func (s *Service) principalesDe(internos []repositories.InternalTicket) (map[int64]repositories.Ticket, error) {
	if len(internos) == 0 {
		return map[int64]repositories.Ticket{}, nil
	}

	ids := make([]int64, 0, len(internos))
	for _, interno := range internos {
		ids = append(ids, interno.TicketID)
	}

	var tickets []repositories.Ticket
	if err := s.tickets.DB().Where("id IN ?", ids).Find(&tickets).Error; err != nil {
		return nil, err
	}

	porID := make(map[int64]repositories.Ticket, len(tickets))
	for _, ticket := range tickets {
		porID[ticket.ID] = ticket
	}

	return porID, nil
}

// aTicket traduce una fila a lo que ve el resto del sistema.
func (s *Service) aTicket(ticket repositories.Ticket, porID map[int64]auth.Account, categoria *Category, etiquetas []string) Ticket {
	convertido := Ticket{
		ID:                  ticket.ID,
		Number:              ticket.Number,
		Subject:             ticket.Subject,
		Description:         ticket.Description,
		State:               ticket.State,
		Category:            categoria,
		Tags:                etiquetas,
		Parent:              "",
		Requester:           cuentaDeID(porID, ticket.RequesterID),
		CreatedBy:           cuentaDeID(porID, ticket.CreatedByID),
		Assignee:            cuentaDeIDOpcional(porID, ticket.AssigneeID),
		SubjectEditedAt:     ticket.SubjectEditedAt,
		DescriptionEditedAt: ticket.DescriptionEditedAt,
		ResolvedAt:          ticket.ResolvedAt,
		ClosedAt:            ticket.ClosedAt,
		CreatedAt:           ticket.CreatedAt,
		UpdatedAt:           ticket.UpdatedAt,
	}

	return convertido
}

// aTicketInterno traduce un interno, que **hereda del principal** el asunto, la descripción, la
// categoría y las etiquetas (decisión 67).
func (s *Service) aTicketInterno(interno repositories.InternalTicket, principal repositories.Ticket, porID map[int64]auth.Account, categoria *Category, etiquetas []string) Ticket {
	return Ticket{
		ID:          interno.ID,
		Number:      interno.Number,
		Internal:    true,
		Subject:     principal.Subject,
		Description: principal.Description,
		State:       interno.State,
		Reason:      interno.EscalationReason,
		Parent:      principal.Number,
		Category:    categoria,
		Tags:        etiquetas,
		Requester:   cuentaDeID(porID, principal.RequesterID),
		CreatedBy:   cuentaDeID(porID, interno.CreatedByID),
		Assignee:    cuentaDeIDOpcional(porID, interno.AssigneeID),
		ResolvedAt:  interno.ResolvedAt,
		ClosedAt:    interno.ClosedAt,
		CreatedAt:   interno.CreatedAt,
		UpdatedAt:   interno.UpdatedAt,
	}
}

func (s *Service) aComentarios(comentarios []repositories.Comment, porID map[int64]auth.Account) []Comment {
	convertidos := make([]Comment, 0, len(comentarios))

	for _, comentario := range comentarios {
		convertidos = append(convertidos, Comment{
			ID:        comentario.ID,
			Author:    cuentaDeID(porID, comentario.AuthorID),
			Body:      comentario.Body,
			Edited:    comentario.EditedAt != nil,
			Deleted:   comentario.DeletedAt != nil,
			CreatedAt: comentario.CreatedAt,
		})
	}

	return convertidos
}

func (s *Service) aAdjuntos(adjuntos []repositories.Attachment, porID map[int64]auth.Account) []Attachment {
	convertidos := make([]Attachment, 0, len(adjuntos))

	for _, adjunto := range adjuntos {
		convertidos = append(convertidos, Attachment{
			ID:          adjunto.ID,
			Filename:    adjunto.Filename,
			ContentType: adjunto.ContentType,
			Size:        adjunto.SizeBytes,
			CommentID:   adjunto.CommentID,
			UploadedBy:  cuentaDeID(porID, adjunto.UploadedByID),
			CreatedAt:   adjunto.CreatedAt,
		})
	}

	return convertidos
}

func (s *Service) aHistorial(historial []repositories.HistoryEntry, porID map[int64]auth.Account) []HistoryEntry {
	convertidos := make([]HistoryEntry, 0, len(historial))

	for _, entrada := range historial {
		convertidos = append(convertidos, HistoryEntry{
			ID:        entrada.ID,
			Event:     entrada.Event,
			FromState: textoDe(entrada.FromState),
			ToState:   textoDe(entrada.ToState),
			Detail:    textoDe(entrada.Detail),
			Actor:     cuentaDeIDOpcional(porID, entrada.ActorID),
			CreatedAt: entrada.CreatedAt,
		})
	}

	return convertidos
}

// cuentaDeID saca la cuenta del mapa. Si no está —una cuenta borrada, por ejemplo— no se inventa
// nada: el hueco se queda vacío y la pantalla enseña lo que hay.
func cuentaDeID(porID map[int64]auth.Account, id int64) *auth.Account {
	cuenta, hay := porID[id]
	if !hay {
		return nil
	}

	return &cuenta
}

// cuentaDeIDOpcional es lo mismo para una columna que puede estar vacía: sin identificador, no hay
// cuenta que enseñar.
func cuentaDeIDOpcional(porID map[int64]auth.Account, id *int64) *auth.Account {
	if id == nil {
		return nil
	}

	return cuentaDeID(porID, *id)
}

func textoDe(valor *string) string {
	if valor == nil {
		return ""
	}

	return *valor
}

func idsDe(tickets []repositories.Ticket) []int64 {
	ids := make([]int64, 0, len(tickets))
	for _, ticket := range tickets {
		ids = append(ids, ticket.ID)
	}

	return ids
}

// idsDeCategorias junta las categorías de una lista de tickets, **sin repetir y sin ceros**: es lo
// que se lee de una vez para pintar la página.
func idsDeCategorias(tickets []repositories.Ticket) []int64 {
	ids := make([]int64, 0, len(tickets))
	vistas := map[int64]bool{}

	for _, ticket := range tickets {
		if ticket.CategoryID == 0 || vistas[ticket.CategoryID] {
			continue
		}

		vistas[ticket.CategoryID] = true
		ids = append(ids, ticket.CategoryID)
	}

	return ids
}

// idsDePrincipales junta los identificadores de los principales de una lista de internos, que es de
// donde el interno hereda su categoría y sus etiquetas.
func idsDePrincipales(internos []repositories.InternalTicket) []int64 {
	ids := make([]int64, 0, len(internos))
	vistas := map[int64]bool{}

	for _, interno := range internos {
		if vistas[interno.TicketID] {
			continue
		}

		vistas[interno.TicketID] = true
		ids = append(ids, interno.TicketID)
	}

	return ids
}

// idsDeCategoriasDePrincipales junta las categorías de los principales que se han leído para una
// lista de internos.
func idsDeCategoriasDePrincipales(internos []repositories.InternalTicket, principales map[int64]repositories.Ticket) []int64 {
	ids := make([]int64, 0, len(internos))
	vistas := map[int64]bool{}

	for _, interno := range internos {
		principal, hay := principales[interno.TicketID]
		if !hay || principal.CategoryID == 0 || vistas[principal.CategoryID] {
			continue
		}

		vistas[principal.CategoryID] = true
		ids = append(ids, principal.CategoryID)
	}

	return ids
}

func idsDeComentarios(comentarios []repositories.Comment) []int64 {
	ids := make([]int64, 0, len(comentarios))
	for _, comentario := range comentarios {
		ids = append(ids, comentario.AuthorID)
	}

	return ids
}

func idsDeAdjuntos(adjuntos []repositories.Attachment) []int64 {
	ids := make([]int64, 0, len(adjuntos))
	for _, adjunto := range adjuntos {
		ids = append(ids, adjunto.UploadedByID)
	}

	return ids
}

func idsDeHistorial(historial []repositories.HistoryEntry) []int64 {
	ids := make([]int64, 0, len(historial))
	for _, entrada := range historial {
		if entrada.ActorID != nil {
			ids = append(ids, *entrada.ActorID)
		}
	}

	return ids
}

// idsDeObservadores junta los identificadores que hay que leer para pintar la lista: quien observa y
// quien lo añadió.
func idsDeObservadores(observadores []repositories.TicketObserver) []int64 {
	ids := make([]int64, 0, len(observadores)*2)
	for _, observador := range observadores {
		ids = append(ids, observador.AccountID, observador.AddedByID)
	}

	return ids
}

// aObservadores traduce los observadores de un hilo a lo que ve la ficha.
func (s *Service) aObservadores(observadores []repositories.TicketObserver, porID map[int64]auth.Account) []Observer {
	convertidos := make([]Observer, 0, len(observadores))

	for _, observador := range observadores {
		convertidos = append(convertidos, Observer{
			ID:        observador.ID,
			Account:   cuentaDeID(porID, observador.AccountID),
			AddedBy:   cuentaDeID(porID, observador.AddedByID),
			CreatedAt: observador.CreatedAt,
		})
	}

	return convertidos
}

// traducirTicket deja el error del repositorio en el del servicio, que es el que viaja a la API.
func traducirTicket(err error) error {
	if err == repositories.ErrTicketNotFound {
		return ErrTicketNotFound
	}

	return err
}
