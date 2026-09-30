package services

import (
	"strings"

	"github.com/juanky-estevez/go-logs"
	"gorm.io/gorm"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// CreateInput es lo que llega para dar de alta un ticket.
type CreateInput struct {
	Subject     string
	Description string
	// **Todo ticket nace con una categoría** (docs/modules/tickets.md, decisión 65): es obligatoria, y
	// el alta la exige antes de nada. Las etiquetas son opcionales y vienen como se escriben: el
	// servicio las normaliza al guardarlas.
	CategoryID int64
	Tags       []string
	// RequesterID y RequesterEmail sólo los usa Soporte, para crear un ticket en nombre de otra
	// persona. El correo es lo que tiene a mano cuando alguien llama, y es el identificador de las
	// personas en este producto.
	RequesterID    *int64
	RequesterEmail string
}

// Create da de alta un ticket principal: le pone número, lo reparte y avisa.
//
// El número y el ticket nacen **en la misma transacción**: un ticket sin número no puede existir, y
// dos tickets creados a la vez no pueden llevarse el mismo (docs/modules/tickets.md, sección 2.2).
func (s *Service) Create(entrada CreateInput, actor auth.Identity) (Ticket, error) {
	// El Administrador mira los tickets, no los crea (docs/usuarios-y-permisos.md, sección 3).
	if actor.Role == auth.RoleAdministrador {
		return Ticket{}, ErrForbidden
	}

	asunto := strings.TrimSpace(entrada.Subject)
	if asunto == "" {
		return Ticket{}, ErrSubjectRequired
	}

	// La descripción es **texto con formato** (docs/modules/tickets.md, sección 2.3): si trae HTML que
	// no está en la lista blanca, no se guarda nada. El asunto, en cambio, sigue siendo texto plano.
	descripcion := strings.TrimSpace(entrada.Description)
	if descripcion == "" {
		return Ticket{}, ErrBodyRequired
	}

	descripcion, err := SanearCuerpo(descripcion)
	if err != nil {
		return Ticket{}, err
	}

	// **La mención vale en cualquier cuerpo del ticket** (docs/modules/tickets.md, decisión 58): la
	// descripción del alta se trata igual que un comentario, porque el backend no puede fiarse de que
	// el editor sólo ofrezca etiquetar en los comentarios —un pegado o una llamada directa a la API
	// meten la mención donde sea—. Se comprueban una a una y, si alguna no vale, no se guarda nada.
	menciones := MencionesDelCuerpo(descripcion)
	if err := s.validarMenciones(menciones, nil, actor); err != nil {
		return Ticket{}, err
	}

	// **La categoría se comprueba antes de abrir la transacción y antes de repartir**: un ticket sin
	// categoría no puede existir, así que no se numera ni se avisa a nadie por él. Y las etiquetas se
	// normalizan aquí, para que un texto que no vale no llegue a la base.
	categoria, err := s.categoriaParaTicket(entrada.CategoryID)
	if err != nil {
		return Ticket{}, err
	}

	etiquetas, err := NormalizarEtiquetas(entrada.Tags)
	if err != nil {
		return Ticket{}, err
	}

	solicitante, err := s.solicitante(entrada.RequesterID, entrada.RequesterEmail, actor)
	if err != nil {
		return Ticket{}, err
	}

	config, err := s.config.TicketSettings()
	if err != nil {
		return Ticket{}, err
	}

	// El turno se decide **antes** de abrir la transacción: preguntar a `users` con una transacción
	// abierta es dejar una conexión esperando a otra.
	responsable, err := s.repartir(config.MainAssignment, auth.RoleSoporte, false)
	if err != nil {
		return Ticket{}, err
	}

	var (
		ticket repositories.Ticket
		numero string
	)

	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		year := s.now().Year()

		secuencia, compuesto, err := s.tickets.SiguienteNumero(tx, year, config.NumberPrefix)
		if err != nil {
			return err
		}
		numero = compuesto

		ticket, err = s.tickets.CreateTicket(tx, repositories.CreateTicketInput{
			Number:      compuesto,
			NumberYear:  year,
			NumberSeq:   secuencia,
			Subject:     asunto,
			Description: descripcion,
			State:       repositories.StateNuevo,
			CategoryID:  categoria.ID,
			RequesterID: solicitante.ID,
			CreatedByID: actor.ID,
			AssigneeID:  idDe(responsable),
		})
		if err != nil {
			return err
		}

		// **Las etiquetas se guardan con el ticket**, en la misma transacción: un ticket con una
		// etiqueta y sin su fila sería dos verdades.
		if err := s.categories.ReplaceTags(tx, ticket.ID, etiquetas, actor.ID); err != nil {
			return err
		}

		if err := s.conversation.AddHistory(tx, repositories.Anotar(
			repositories.DestinoDePrincipal(ticket.ID),
			&actor.ID,
			"creado",
			"",
			repositories.StateNuevo,
			"",
		)); err != nil {
			return err
		}

		// Los etiquetados en la descripción **nacen observadores del ticket**, en la misma transacción
		// que el ticket: un ticket creado con una mención y sin su observador sería dos verdades.
		return s.sincronizarObservadores(
			tx,
			repositories.DestinoDePrincipal(ticket.ID),
			ticket.Description,
			nil,
			menciones,
			actor.ID,
		)
	})
	if err != nil {
		return Ticket{}, err
	}

	// El aviso sale **después** de que el ticket exista: el correo es un aviso, no la acción
	// (docs/flujos.md, sección 2).
	s.avisarCreacion(config, numero, asunto, solicitante, responsable, ticket.ID)
	s.avisarEtiquetado(numero, asunto, menciones)

	// Y el motor de IA empieza a redactar **sin que nadie espere**: quien abre un ticket no puede
	// quedarse mirando un contenedor (docs/modules/ai.md, decisión 2). Los dos campos nacen aquí.
	s.pedirAlMotor(numero, TipoMotivo, TipoUltimaAccion)

	return Ticket{
		ID:          ticket.ID,
		Number:      ticket.Number,
		Subject:     ticket.Subject,
		Description: ticket.Description,
		State:       ticket.State,
		Category:    aCategoria(categoria),
		Tags:        etiquetas,
		Requester:   &solicitante,
		CreatedBy:   cuentaDe(actor),
		Assignee:    responsable,
		CreatedAt:   ticket.CreatedAt,
		UpdatedAt:   ticket.UpdatedAt,
	}, nil
}

// solicitante decide de quién es el ticket: de quien lo crea, o de quien diga Soporte.
//
// **Crear en nombre de otro es de Soporte** (docs/usuarios-y-permisos.md, sección 3), y el
// Administrador no crea tickets, así que sólo le queda al usuario y a Desarrollo crear los suyos.
func (s *Service) solicitante(pedido *int64, correo string, actor auth.Identity) (auth.Account, error) {
	// Sin nada, o con uno mismo: el ticket es de quien lo crea.
	sinPedido := (pedido == nil || *pedido == actor.ID) && strings.TrimSpace(correo) == ""

	if sinPedido {
		cuenta, err := s.accounts.ByID(actor.ID)
		if err != nil {
			return auth.Account{}, ErrRequesterUnknown
		}

		return cuenta, nil
	}

	// Crear en nombre de otro es de Soporte (docs/usuarios-y-permisos.md, sección 3).
	if actor.Role != auth.RoleSoporte {
		return auth.Account{}, ErrRequesterNotYours
	}

	var (
		cuenta auth.Account
		err    error
	)

	if pedido != nil {
		cuenta, err = s.accounts.ByID(*pedido)
	} else {
		cuenta, err = s.accounts.ByEmail(strings.TrimSpace(correo))
	}

	if err != nil || !cuenta.IsActive {
		return auth.Account{}, ErrRequesterUnknown
	}

	return cuenta, nil
}

// Assignees son las personas que pueden ser responsables de cada tipo de ticket.
//
// Vive aquí, y no en `users`, porque **la pantalla del frontend sólo puede hablar con su propia API**
// (docs/arquitectura.md, sección 4): el módulo ya le pide a `users` los técnicos activos para el
// turno, y esto es la misma pregunta.
func (s *Service) Assignees(actor auth.Identity) (map[string][]auth.Account, error) {
	if actor.Role == auth.RoleUsuario {
		return nil, ErrForbidden
	}

	soporte, err := s.accounts.ActiveByRole(auth.RoleSoporte)
	if err != nil {
		return nil, err
	}

	desarrollo, err := s.accounts.ActiveByRole(auth.RoleDesarrollo)
	if err != nil {
		return nil, err
	}

	return map[string][]auth.Account{
		"principal": soporte,
		"interno":   desarrollo,
	}, nil
}

// repartir elige a quién le toca, o nada si el reparto está apagado o no hay a quién repartir.
//
// El turno es **quien hace más tiempo que no recibe uno de ese tipo**, y quien no ha recibido ninguno
// va delante: así no hay puntero de turno que guardar ni que desincronizarse
// (docs/modules/tickets.md, sección 3.3).
func (s *Service) repartir(asignacion, papel string, internos bool) (*auth.Account, error) {
	if asignacion != AssignmentRoundRobin {
		return nil, nil
	}

	candidatos, err := s.accounts.ActiveByRole(papel)
	if err != nil {
		return nil, err
	}
	if len(candidatos) == 0 {
		// No hay a quién: el ticket queda sin responsable y en la bandeja, para que alguien lo coja.
		logs.LogWarning("no hay ninguna cuenta activa a la que repartir el ticket: se queda sin responsable")
		return nil, nil
	}

	ultimas, err := s.tickets.UltimaAsignacion(internos)
	if err != nil {
		return nil, err
	}

	// El primero de la lista es el que va por delante mientras nadie demuestre lo contrario, y la
	// lista viene ordenada por apellidos y nombre: dos técnicos que nunca han recibido uno no
	// dependen del azar de cómo los devuelva la base.
	mejor := candidatos[0]
	mejorFecha, haRecibido := ultimas[mejor.ID]

	for _, candidato := range candidatos[1:] {
		fecha, haRecibidoEste := ultimas[candidato.ID]

		switch {
		case !haRecibidoEste && haRecibido:
			// Quien nunca ha recibido uno va delante de quien sí.
			mejor, mejorFecha, haRecibido = candidato, fecha, false
		case haRecibidoEste == haRecibido && haRecibido && fecha.Before(mejorFecha):
			// Y entre los que sí han recibido, el que hace más tiempo.
			mejor, mejorFecha = candidato, fecha
		}
	}

	elegido := mejor
	return &elegido, nil
}

// cuentaDe arma la cuenta de quien está haciendo algo, para el historial y la ficha.
func cuentaDe(actor auth.Identity) *auth.Account {
	cuenta := auth.Account{
		ID:       actor.ID,
		Name:     actor.Name,
		LastName: actor.LastName,
		Email:    actor.Email,
		Role:     actor.Role,
		Origin:   actor.Origin,
		Language: actor.Language,
		IsActive: actor.IsActive,
	}

	return &cuenta
}

// idDe devuelve el identificador de una cuenta, o nada: es el `assignee_id` de la base.
func idDe(cuenta *auth.Account) *int64 {
	if cuenta == nil {
		return nil
	}

	return &cuenta.ID
}
