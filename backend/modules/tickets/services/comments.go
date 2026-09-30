package services

import (
	"strings"

	"gorm.io/gorm"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// Comment escribe un comentario en la conversación de un ticket.
//
// **En un ticket cerrado no se comenta** (decisión 19): para seguir hablando hay que reabrirlo. Y
// quien escribe es quien puede: el solicitante y Soporte en el principal, Soporte y Desarrollo en el
// interno (docs/modules/tickets.md, sección 4).
//
// **El texto puede ir vacío**: un comentario puede ser sólo un adjunto —«mira esto» con la captura—, y
// es una decisión del responsable del 2026-09-25. Aquí no habría forma de comprobarlo aunque se
// quisiera: **los archivos de un comentario suben después**, con su `commentId`, así que en este
// momento no se sabe si van a llegar. La regla «texto **o** archivos» la sostiene la pantalla, que es
// quien lo sabe, y editar un comentario sí sigue exigiendo texto (decisiones 41 y 42).
func (s *Service) Comment(number, body string, actor auth.Identity) (Comment, error) {
	texto := strings.TrimSpace(body)

	// El texto de un comentario es **texto con formato** (docs/modules/tickets.md, sección 2.3): si
	// trae HTML que no está en la lista blanca, se dice y no se guarda nada.
	if texto != "" {
		saneado, err := SanearCuerpo(texto)
		if err != nil {
			return Comment{}, err
		}
		texto = saneado
	}

	principal, esInterno, _, err := s.porNumero(number)
	if err != nil {
		return Comment{}, err
	}

	destino, convertido, err := s.destinoDe(principal, esInterno, actor)
	if err != nil {
		return Comment{}, err
	}

	if principal.State == repositories.StateCerrado {
		return Comment{}, ErrClosed
	}
	if !puedeComentar(actor, convertido) {
		return Comment{}, ErrForbidden
	}

	// El interno se cierra aparte del principal: cada uno tiene su propio estado que respetar.
	if esInterno {
		interno, _, err := s.tickets.InternalByTicket(principal.ID)
		if err != nil {
			return Comment{}, err
		}
		if interno.State == repositories.StateCerrado {
			return Comment{}, ErrClosed
		}
	}

	// **La mención vale en cualquier cuerpo del ticket** (docs/modules/tickets.md, decisión 58): el
	// editor sólo ofrece etiquetar en los comentarios, pero el backend no puede fiarse de eso —una
	// llamada directa a la API mete la mención donde sea—, así que se comprueba aquí, una a una, y una
	// que no valga rechaza el comentario entero.
	menciones := MencionesDelCuerpo(texto)
	if err := s.validarMenciones(menciones, nil, actor); err != nil {
		return Comment{}, err
	}

	comentario, err := s.escribirComentario(destino, texto, actor, principal.Description, nil, menciones)
	if err != nil {
		return Comment{}, err
	}

	// Quien se acaba de etiquetar recibe su aviso, **una vez por persona** y sin el texto de nadie.
	s.avisarEtiquetado(number, principal.Subject, menciones)

	// Comentar cambia lo último que ha pasado, así que **la última acción se vuelve a redactar**
	// (docs/modules/ai.md, decisión 4). El motivo no: de qué va el ticket no cambia porque alguien
	// conteste.
	s.pedirAlMotor(number, TipoUltimaAccion)

	return comentario, nil
}

// escribirComentario guarda el comentario en el destino que se le diga y lo devuelve ya traducido,
// **reajustando los observadores en la misma transacción** (docs/modules/tickets.md, decisión 63).
//
// `anteriores` son las menciones que tenía ese cuerpo antes —vacío en un comentario nuevo— y `nuevos`
// las que tiene ahora; `descripcion` es la descripción del principal, que cuenta como un cuerpo más de
// su hilo. El motivo de un re-escalado —texto plano y sin menciones— llega con todo vacío: no añade a
// nadie, y el reajuste no quita a nadie que siga nombrado.
//
// **Aquí el texto llega ya listo**: el que escribe por la API lo ha saneado, y el único camino de
// dentro —el motivo de un re-escalado, que es texto plano— lo escapa antes de llamar. Sanea quien tiene
// algo que decirle a alguien: aquí no hay a quién contestarle una clave de error.
func (s *Service) escribirComentario(destino repositories.Destino, texto string, actor auth.Identity, descripcion string, anteriores, nuevos []int64) (Comment, error) {
	var creado repositories.Comment
	err := s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		var err error
		creado, err = s.escribirComentarioEnTx(tx, destino, texto, actor, descripcion, anteriores, nuevos)

		return err
	})
	if err != nil {
		return Comment{}, err
	}

	return Comment{
		ID:        creado.ID,
		Author:    cuentaDe(actor),
		Body:      creado.Body,
		CreatedAt: creado.CreatedAt,
	}, nil
}

// escribirComentarioEnTx escribe el comentario **dentro de una transacción que ya está abierta** y deja
// reajustados los observadores en la misma.
//
// Existe porque hay acciones que no pueden ir en dos transacciones: **cerrar un ticket es el cambio de
// estado, el comentario, el historial y lo que se sincroniza con el principal, o nada** (decisión 81),
// y así el movimiento lo llama desde dentro de la suya. `escribirComentario` es el mismo trabajo con su
// propia transacción, para el comentario suelto.
//
// Aquí el texto llega ya saneado y con las menciones validadas: es lo mismo que exige un comentario
// normal, hecho por quien puede rechazarlo con una clave de error antes de abrir la transacción.
func (s *Service) escribirComentarioEnTx(tx *gorm.DB, destino repositories.Destino, texto string, actor auth.Identity, descripcion string, anteriores, nuevos []int64) (repositories.Comment, error) {
	comentario := repositories.Comment{
		TicketID:         destino.TicketID,
		InternalTicketID: destino.InternalTicketID,
		AuthorID:         actor.ID,
		Body:             texto,
	}

	creado, err := s.conversation.CreateComment(tx, comentario)
	if err != nil {
		return repositories.Comment{}, err
	}

	if err := s.sincronizarObservadores(tx, destino, descripcion, anteriores, nuevos, actor.ID); err != nil {
		return repositories.Comment{}, err
	}

	return creado, nil
}

// EditComment cambia el texto de un comentario. **Sólo su autor**, y el comentario queda marcado.
//
// **Aquí el texto sí es obligatorio** (decisión 42): dejar en blanco un comentario que ya se escribió
// no es lo mismo que nacer sin texto —eso es un comentario que sólo lleva un adjunto—, y lo que se
// quiere en ese caso es borrarlo, que para eso está el borrado.
func (s *Service) EditComment(number string, commentID int64, body string, actor auth.Identity) (Comment, error) {
	texto := strings.TrimSpace(body)
	if texto == "" {
		return Comment{}, ErrCommentRequired
	}

	// Y aquí también es texto con formato: editarlo no puede ser la puerta de atrás para colar HTML que
	// al escribirlo se habría rechazado.
	saneado, err := SanearCuerpo(texto)
	if err != nil {
		return Comment{}, err
	}
	texto = saneado

	principal, esInterno, _, err := s.porNumero(number)
	if err != nil {
		return Comment{}, err
	}

	_, convertido, err := s.destinoDe(principal, esInterno, actor)
	if err != nil {
		return Comment{}, err
	}

	destino, err := s.destinoPorTicket(principal.ID, esInterno)
	if err != nil {
		return Comment{}, err
	}

	comentario, err := s.comentarioDelTicket(commentID, destino)
	if err != nil {
		return Comment{}, err
	}

	if comentario.AuthorID != actor.ID {
		return Comment{}, ErrCommentNotYours
	}
	if comentario.DeletedAt != nil {
		// Lo que está borrado, borrado está: no se puede resucitar editándolo.
		return Comment{}, ErrCommentNotFound
	}
	if !puedeComentar(actor, convertido) {
		return Comment{}, ErrForbidden
	}

	// Editar un comentario también etiqueta: se miran las menciones **de antes y de ahora**, porque el
	// reajuste de la decisión 63 necesita las dos —lo que se añade es lo nuevo, y lo que sobra es lo
	// que ya no nombra ningún cuerpo del hilo—.
	anteriores := MencionesDelCuerpo(comentario.Body)
	menciones := MencionesDelCuerpo(texto)
	if err := s.validarMenciones(menciones, anteriores, actor); err != nil {
		return Comment{}, err
	}

	ahora := s.now()
	comentario.Body = texto
	comentario.EditedAt = &ahora

	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.conversation.UpdateComment(tx, comentario); err != nil {
			return err
		}

		return s.sincronizarObservadores(tx, destino, principal.Description, anteriores, menciones, actor.ID)
	})
	if err != nil {
		return Comment{}, err
	}

	// Sólo se avisa de lo que se menciona **de nuevo**: quitar una mención no manda ningún correo.
	s.avisarEtiquetado(number, principal.Subject, diferencia(menciones, anteriores))

	return Comment{
		ID:        comentario.ID,
		Author:    cuentaDe(actor),
		Body:      comentario.Body,
		Edited:    true,
		CreatedAt: comentario.CreatedAt,
	}, nil
}

// DeleteComment borra un comentario: **vacía el texto** y deja la marca.
//
// La fila no desaparece —la conversación sigue contando que ahí hubo un comentario— pero el texto se
// vacía de verdad: si alguien pegó algo que no debía, borrar tiene que borrar
// (docs/modules/tickets.md, sección 2.3). **Los adjuntos del comentario no se borran**: pueden ser la
// prueba de algo, y quien borra su texto no está pidiendo que desaparezca el archivo.
func (s *Service) DeleteComment(number string, commentID int64, actor auth.Identity) error {
	principal, esInterno, _, err := s.porNumero(number)
	if err != nil {
		return err
	}

	destino, err := s.destinoPorTicket(principal.ID, esInterno)
	if err != nil {
		return err
	}

	comentario, err := s.comentarioDelTicket(commentID, destino)
	if err != nil {
		return err
	}

	if comentario.AuthorID != actor.ID {
		return ErrCommentNotYours
	}

	if comentario.DeletedAt != nil {
		return nil
	}

	ahora := s.now()
	comentario.Body = ""
	comentario.DeletedAt = &ahora

	return s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		return s.conversation.UpdateComment(tx, comentario)
	})
}

// comentarioDelTicket busca el comentario y comprueba que es de ese ticket: un identificador suelto no
// puede servir para tocar la conversación de otro.
func (s *Service) comentarioDelTicket(commentID int64, destino repositories.Destino) (repositories.Comment, error) {
	comentario, err := s.conversation.CommentByID(commentID)
	if err != nil {
		if err == repositories.ErrCommentNotFound {
			return repositories.Comment{}, ErrCommentNotFound
		}

		return repositories.Comment{}, err
	}

	esDeEseTicket := (destino.TicketID != nil && comentario.TicketID != nil && *comentario.TicketID == *destino.TicketID) ||
		(destino.InternalTicketID != nil && comentario.InternalTicketID != nil && *comentario.InternalTicketID == *destino.InternalTicketID)
	if !esDeEseTicket {
		return repositories.Comment{}, ErrCommentNotFound
	}

	return comentario, nil
}

// destinoDe deja el destino del ticket y su versión de lectura, comprobando que se puede ver.
func (s *Service) destinoDe(principal repositories.Ticket, esInterno bool, actor auth.Identity) (repositories.Destino, Ticket, error) {
	if esInterno {
		interno, _, err := s.tickets.InternalByTicket(principal.ID)
		if err != nil {
			return repositories.Destino{}, Ticket{}, err
		}

		if actor.Role == auth.RoleUsuario {
			return repositories.Destino{}, Ticket{}, ErrTicketNotFound
		}

		porID, err := s.accounts.ByIDs([]int64{principal.RequesterID, principal.CreatedByID, interno.CreatedByID})
		if err != nil {
			return repositories.Destino{}, Ticket{}, err
		}

		// La categoría y las etiquetas no hacen falta para decidir un permiso: aquí sólo se comprueba
		// quién es y de quién es el ticket (`puedeVer` y `puedeComentar`), así que no se leen.
		return repositories.DestinoDeInterno(interno.ID), s.aTicketInterno(interno, principal, porID, nil, nil), nil
	}

	porID, err := s.accounts.ByIDs([]int64{principal.RequesterID, principal.CreatedByID})
	if err != nil {
		return repositories.Destino{}, Ticket{}, err
	}

	convertido := s.aTicket(principal, porID, nil, nil)

	if !puedeVer(actor, convertido) {
		return repositories.Destino{}, Ticket{}, ErrTicketNotFound
	}

	return repositories.DestinoDePrincipal(principal.ID), convertido, nil
}

// destinoPorTicket arma el destino a partir del principal y de si se trata de su interno.
func (s *Service) destinoPorTicket(ticketID int64, esInterno bool) (repositories.Destino, error) {
	if !esInterno {
		return repositories.DestinoDePrincipal(ticketID), nil
	}

	interno, hay, err := s.tickets.InternalByTicket(ticketID)
	if err != nil {
		return repositories.Destino{}, err
	}
	if !hay {
		return repositories.Destino{}, ErrTicketNotFound
	}

	return repositories.DestinoDeInterno(interno.ID), nil
}
