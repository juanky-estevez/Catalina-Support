package services

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// UpdateInput es lo que se puede cambiar de un ticket: el asunto, la descripción, la categoría y las
// etiquetas. El motivo del escalado **no se edita**, y el estado va por sus propias acciones
// (docs/modules/tickets.md, sección 2.1).
//
// La categoría y las etiquetas van con puntero porque este es el `PATCH` que ya existía: nulo es «no
// lo toques», y eso distingue corregir la categoría de dejarla como estaba.
type UpdateInput struct {
	Subject     string
	Description string
	CategoryID  *int64
	Tags        *[]string
}

// Update cambia el asunto o la descripción, y deja la marca del campo que se tocó.
//
// **Un ticket cerrado no se edita**: primero se reabre (decisión 31). Es la misma regla que ya estaba
// decidida para los comentarios, y no deja dos maneras de decir lo mismo.
func (s *Service) Update(number string, entrada UpdateInput, actor auth.Identity) (Ticket, error) {
	ticket, err := s.tickets.ByNumber(number)
	if err != nil {
		return Ticket{}, traducirTicket(err)
	}

	detalle, err := s.detalleDePrincipal(ticket, actor)
	if err != nil {
		return Ticket{}, err
	}

	if !puedeEditar(actor, detalle.Ticket) {
		return Ticket{}, ErrForbidden
	}
	if ticket.State == repositories.StateCerrado {
		return Ticket{}, ErrClosed
	}

	// El asunto es texto plano y la descripción es **texto con formato**: el saneado va sólo en la
	// descripción, y si no pasa no se guarda nada (docs/modules/tickets.md, sección 2.3).
	asunto := strings.TrimSpace(entrada.Subject)
	descripcion := strings.TrimSpace(entrada.Description)
	if asunto == "" {
		return Ticket{}, ErrSubjectRequired
	}
	if descripcion == "" {
		return Ticket{}, ErrBodyRequired
	}
	if descripcion, err = SanearCuerpo(descripcion); err != nil {
		return Ticket{}, err
	}

	// **La mención vale también en la descripción** (docs/modules/tickets.md, decisión 58), al crearla
	// y al editarla. Sólo se valida lo que de verdad se está escribiendo: si el `PATCH` cambia el
	// asunto y deja la descripción como estaba, no se le pide cuentas a un texto que no se toca.
	anteriores := MencionesDelCuerpo(ticket.Description)
	menciones := MencionesDelCuerpo(descripcion)
	descripcionCambia := ticket.Description != descripcion

	if descripcionCambia {
		if err := s.validarMenciones(menciones, anteriores, actor); err != nil {
			return Ticket{}, err
		}
	}

	// **La categoría y las etiquetas se cambian con las mismas reglas que al crear** (decisión 65), y
	// las cambia quien puede editar el ticket: el solicitante en el suyo y Soporte. Desarrollo no
	// escribe en el principal (docs/usuarios-y-permisos.md, regla 2 de la sección 3).
	categoria := detalle.Ticket.Category
	categoriaCambia := false

	if entrada.CategoryID != nil {
		cambiada, err := s.categoriaParaTicket(*entrada.CategoryID)
		if err != nil {
			return Ticket{}, err
		}

		categoria = aCategoria(cambiada)
		categoriaCambia = ticket.CategoryID != cambiada.ID
	}

	etiquetas := detalle.Ticket.Tags
	etiquetasCambian := entrada.Tags != nil

	if etiquetasCambian {
		etiquetas, err = s.etiquetasDelCatalogo(*entrada.Tags, actor)
		if err != nil {
			return Ticket{}, err
		}
	}

	ahora := s.now()
	cambio := ticket.Subject != asunto || descripcionCambia || categoriaCambia || etiquetasCambian
	if !cambio {
		return detalle.Ticket, nil
	}

	if ticket.Subject != asunto {
		ticket.Subject = asunto
		ticket.SubjectEditedAt = &ahora
	}
	if ticket.Description != descripcion {
		ticket.Description = descripcion
		ticket.DescriptionEditedAt = &ahora
	}
	if categoriaCambia {
		ticket.CategoryID = categoria.ID
	}

	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.tickets.UpdateTicket(tx, ticket); err != nil {
			return err
		}

		// Las etiquetas se reemplazan enteras, como el texto: quitar una es quitarla de la lista.
		if etiquetasCambian {
			if err := s.categories.ReplaceTags(tx, ticket.ID, etiquetas, actor.ID); err != nil {
				return err
			}
		}

		if err := s.conversation.AddHistory(tx, repositories.Anotar(
			repositories.DestinoDePrincipal(ticket.ID),
			&actor.ID,
			"editado",
			"",
			"",
			"",
		)); err != nil {
			return err
		}

		// Editar la descripción reajusta los observadores igual que un comentario: se añaden las
		// menciones nuevas y se quitan los que ya no menciona ningún cuerpo del hilo (decisión 63).
		return s.sincronizarObservadores(
			tx,
			repositories.DestinoDePrincipal(ticket.ID),
			ticket.Description,
			anteriores,
			menciones,
			actor.ID,
		)
	})
	if err != nil {
		return Ticket{}, err
	}

	// El aviso de etiquetado va sólo a quien se menciona **de nuevo**, y después de guardar: el
	// correo es un aviso, no la acción.
	s.avisarEtiquetado(number, ticket.Subject, diferencia(menciones, anteriores))

	// Al editar el texto cambian las dos cosas que redacta el motor: de qué va el ticket —el motivo— y
	// qué fue lo último que pasó —la edición— (docs/modules/ai.md, decisiones 4 y 7).
	s.pedirAlMotor(number, TipoMotivo, TipoUltimaAccion)

	actualizado, err := s.detalleDePrincipal(ticket, actor)
	if err != nil {
		return Ticket{}, err
	}

	return actualizado.Ticket, nil
}

// etiquetasDelCatalogo normaliza la lista que llega y comprueba que existe.
//
// **El catálogo se cura** (decisión 84): **sólo el Administrador introduce nombres de etiqueta**. Quien
// no lo es **elige de las que hay**, así que una etiqueta que no está en el catálogo se rechaza en vez
// de nacer sola al escribirla. Es lo que pidió el responsable: etiquetar es de Soporte y Desarrollo,
// pero el catálogo lo mantiene el Administrador. Vive aquí porque lo usan las dos puertas que cambian
// etiquetas —la edición entera y el `PATCH` de sólo etiquetas—, y las dos tienen que aplicar la misma
// regla.
func (s *Service) etiquetasDelCatalogo(tags []string, actor auth.Identity) ([]string, error) {
	normalizadas, err := NormalizarEtiquetas(tags)
	if err != nil {
		return nil, err
	}

	if actor.Role == auth.RoleAdministrador {
		return normalizadas, nil
	}

	for _, etiqueta := range normalizadas {
		if _, err := s.categories.TagNameByNormalized(etiqueta); err != nil {
			if errors.Is(err, repositories.ErrTagNotFound) {
				return nil, ErrTagDesconocida
			}

			return nil, err
		}
	}

	return normalizadas, nil
}

// etiquetasIguales compara dos listas de etiquetas **sin importar el orden**: `TagsByTicket` las
// devuelve por nombre, y lo que importa es el conjunto, no cómo venga.
func etiquetasIguales(una, otra []string) bool {
	if len(una) != len(otra) {
		return false
	}

	vistas := make(map[string]int, len(una))
	for _, etiqueta := range una {
		vistas[etiqueta]++
	}
	for _, etiqueta := range otra {
		vistas[etiqueta]--
		if vistas[etiqueta] < 0 {
			return false
		}
	}

	return true
}

// UpdateTags cambia **sólo las etiquetas** de un ticket, y es lo que permite que Desarrollo etiquete
// desde el interno (decisión 85): las etiquetas son **del principal** y el interno las hereda
// (decisión 67), así que se aplican al principal aunque el número que llegue sea el del interno.
//
// La puerta es la misma que la del resto de la clasificación: quien puede editar el ticket. Desarrollo
// no escribe en el principal —con el número del principal recibe `403`, como con el asunto—, pero **en
// el interno sí puede etiquetar**, que es lo que la 85 le reconoce. El número del interno se resuelve
// con `porNumero`, el mismo mecanismo que usan asignar y mover, y no con uno nuevo.
//
// El catálogo se cura igual que en `Update`: quien no es el Administrador elige de las que hay, y una
// etiqueta que no existe responde `422` (`tickets.etiqueta.desconocida`).
func (s *Service) UpdateTags(number string, tags []string, actor auth.Identity) (Ticket, error) {
	principal, esInterno, _, err := s.porNumero(number)
	if err != nil {
		return Ticket{}, err
	}

	if esInterno {
		// **Desde el interno etiqueta Desarrollo**, que es quien trabaja ahí; Soporte etiqueta el
		// principal, que es su sitio (decisión 85). Ni el usuario ni el Administrador tocan etiquetas.
		if actor.Role != auth.RoleDesarrollo {
			return Ticket{}, ErrForbidden
		}

		interno, _, err := s.tickets.InternalByTicket(principal.ID)
		if err != nil {
			return Ticket{}, err
		}
		if interno.State == repositories.StateCerrado {
			return Ticket{}, ErrClosed
		}
	} else {
		detalle, err := s.detalleDePrincipal(principal, actor)
		if err != nil {
			return Ticket{}, err
		}

		// En el principal manda la regla de siempre: editan el solicitante y Soporte. Desarrollo —y el
		// Administrador— reciben `403`.
		if !puedeEditar(actor, detalle.Ticket) {
			return Ticket{}, ErrForbidden
		}
		if principal.State == repositories.StateCerrado {
			return Ticket{}, ErrClosed
		}
	}

	etiquetas, err := s.etiquetasDelCatalogo(tags, actor)
	if err != nil {
		return Ticket{}, err
	}

	// Si la lista es la misma, no hay nada que guardar: no se apunta un «editado» que no ocurrió ni se
	// vuelve a pedir el resumen al motor.
	actuales, err := s.categories.TagsByTicket(principal.ID)
	if err != nil {
		return Ticket{}, err
	}
	if etiquetasIguales(actuales, etiquetas) {
		return s.ticketDe(number, actor)
	}

	// **Las etiquetas viven en el principal** (decisión 67): se reemplazan por su identificador, que es
	// el que resuelve `porNumero` tanto si el número era el del principal como el del interno.
	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.categories.ReplaceTags(tx, principal.ID, etiquetas, actor.ID); err != nil {
			return err
		}

		return s.conversation.AddHistory(tx, repositories.Anotar(
			repositories.DestinoDePrincipal(principal.ID),
			&actor.ID,
			"editado",
			"",
			"",
			"",
		))
	})
	if err != nil {
		return Ticket{}, err
	}

	// Etiquetar es un movimiento más, y el motor lo cuenta en la última acción (docs/modules/ai.md,
	// decisión 4).
	s.pedirAlMotor(number, TipoUltimaAccion)

	return s.ticketDe(number, actor)
}

// Assign pone responsable. El reparto decide el punto de partida, no el destino: reasignar siempre se
// puede (docs/modules/tickets.md, sección 3.3).
func (s *Service) Assign(number string, assigneeID int64, actor auth.Identity) (Ticket, error) {
	ticket, interno, _, err := s.porNumero(number)
	if err != nil {
		return Ticket{}, err
	}

	if !s.puedeAsignar(actor, interno) {
		return Ticket{}, ErrForbidden
	}

	// Quien puede ser responsable: en el principal, un técnico de Soporte; en el interno, alguien de
	// Desarrollo (decisión 33).
	papel := auth.RoleSoporte
	if interno {
		papel = auth.RoleDesarrollo
	}

	candidatos, err := s.accounts.ActiveByRole(papel)
	if err != nil {
		return Ticket{}, err
	}

	var elegido *auth.Account
	for i := range candidatos {
		if candidatos[i].ID == assigneeID {
			elegido = &candidatos[i]
			break
		}
	}
	if elegido == nil {
		return Ticket{}, ErrAssigneeUnknown
	}

	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		if interno {
			ticketInterno, _, err := s.tickets.InternalByTicket(ticket.ID)
			if err != nil {
				return err
			}

			ticketInterno.AssigneeID = &elegido.ID
			if err := s.tickets.UpdateInternal(tx, ticketInterno); err != nil {
				return err
			}

			return s.conversation.AddHistory(tx, repositories.Anotar(
				repositories.DestinoDeInterno(ticketInterno.ID),
				&actor.ID,
				"asignado",
				"",
				"",
				elegido.FullName(),
			))
		}

		ticket.AssigneeID = &elegido.ID
		if err := s.tickets.UpdateTicket(tx, ticket); err != nil {
			return err
		}

		return s.conversation.AddHistory(tx, repositories.Anotar(
			repositories.DestinoDePrincipal(ticket.ID),
			&actor.ID,
			"asignado",
			"",
			"",
			elegido.FullName(),
		))
	})
	if err != nil {
		return Ticket{}, err
	}

	// Asignar es un movimiento más, y el motor lo cuenta en la última acción (decisión 4).
	s.pedirAlMotor(number, TipoUltimaAccion)

	return s.ticketDe(number, actor)
}

// Move lleva un ticket a otro estado, con lo que ese movimiento arrastra: el historial, lo que se
// sincronice con el principal y, si viene, **un comentario**.
//
// La tabla de transiciones de `docs/modules/tickets.md`, sección 3, es la que manda: lo que no esté
// ahí no se puede hacer, y quien no sea de los que aparecen en la fila, tampoco.
//
// **Un ticket no se cierra sin decir por qué** (decisión 81): al cerrar, `comment` es obligatorio —lo
// mueve Soporte, Desarrollo en el interno o el propio solicitante, y los tres por el mismo camino— y
// queda **como un comentario de la conversación** escrito por quien cierra. Sin él se responde
// `tickets.cierre.sinComentario`. En los demás estados el comentario es opcional, y si viene se guarda
// igual; si no viene, nada cambia respecto a antes.
func (s *Service) Move(number, to, comment string, actor auth.Identity) (Ticket, error) {
	ticket, interno, _, err := s.porNumero(number)
	if err != nil {
		return Ticket{}, err
	}

	if !estadoValido(to, interno) {
		return Ticket{}, ErrStateUnknown
	}

	// El estado del que se parte es **el del ticket que se mueve**: en un interno, el suyo, que no
	// tiene por qué ser el del principal.
	internoActual, _, err := s.tickets.InternalByTicket(ticket.ID)
	if err != nil {
		return Ticket{}, err
	}

	desde := ticket.State
	if interno {
		desde = internoActual.State
	}

	// Primero, si ese movimiento existe: lo que no está en la tabla no se hace, y da igual quién lo
	// pida. Después, si es de quien lo pide: eso ya es un permiso, y se contesta distinto. El comentario
	// se pide **después del permiso**: a quien no puede cerrar el ticket no se le corrige el texto.
	if !transicionExiste(interno, desde, to) {
		return Ticket{}, ErrTransitionNotAllowed
	}
	if !s.puedeMover(actor, ticket, interno, desde, to) {
		return Ticket{}, ErrForbidden
	}

	texto, menciones, err := s.prepararComentarioDelMovimiento(to, comment, actor)
	if err != nil {
		return Ticket{}, err
	}

	err = s.mover(ticket, internoActual, interno, to, actor, texto, menciones)
	if err != nil {
		return Ticket{}, err
	}

	s.avisarMovimiento(ticket, internoActual, interno, desde, to)

	// Quien se acaba de etiquetar en el comentario del movimiento recibe el mismo aviso que en un
	// comentario: es un comentario más (decisión 81).
	s.avisarEtiquetado(number, ticket.Subject, menciones)

	// **La última acción** se vuelve a redactar con cada movimiento (docs/modules/ai.md, decisión 4), y
	// en un interno que devuelve el principal hay **dos** tickets que han cambiado: el interno y su
	// principal. Se piden los dos.
	s.pedirAlMotor(number, TipoUltimaAccion)
	if interno && ticket.State != desde {
		s.pedirAlMotor(ticket.Number, TipoUltimaAccion)
	}

	return s.ticketDe(number, actor)
}

// ticketDe vuelve a leer el ticket después de un cambio: se contesta con lo que hay de verdad, no con
// lo que creíamos que iba a quedar.
func (s *Service) ticketDe(number string, actor auth.Identity) (Ticket, error) {
	detalle, err := s.ByNumber(number, actor)
	if err != nil {
		return Ticket{}, err
	}

	return detalle.Ticket, nil
}

// prepararComentarioDelMovimiento deja listo el comentario que acompaña a un cambio de estado.
//
// **Cerrar exige decirlo** (decisión 81): sin comentario —ni vacío, ni sólo espacios— se rechaza con
// `tickets.cierre.sinComentario`, y lo rechaza el servidor, no la pantalla. En los demás estados el
// comentario es opcional y, sin él, no hay nada que preparar.
//
// Cuando viene, pasa por **lo mismo que un comentario normal**: se sanea como el cuerpo de cualquier
// comentario (la lista blanca de la sección 2.3) y sus menciones se validan como las de cualquier
// cuerpo (decisión 59). Por eso quien cierra **puede seguir etiquetando** con su aviso si es Soporte o
// Desarrollo, y el solicitante conserva la misma regla que al comentar: no estrena menciones.
func (s *Service) prepararComentarioDelMovimiento(to, comment string, actor auth.Identity) (string, []int64, error) {
	texto := strings.TrimSpace(comment)

	if to == repositories.StateCerrado && texto == "" {
		return "", nil, ErrCierreSinComentario
	}
	if texto == "" {
		return "", nil, nil
	}

	saneado, err := SanearCuerpo(texto)
	if err != nil {
		return "", nil, err
	}

	menciones := MencionesDelCuerpo(saneado)
	if err := s.validarMenciones(menciones, nil, actor); err != nil {
		return "", nil, err
	}

	return saneado, menciones, nil
}

// comentarioDelMovimiento escribe, dentro de la transacción del movimiento, el comentario que lo
// acompaña. Sin texto no escribe nada: el comentario de un movimiento sólo existe si viene.
//
// `anteriores` va vacío porque es un comentario nuevo: todas sus menciones son nuevas y quien se
// etiqueta pasa a observar el ticket, como en cualquier comentario (decisión 63).
func (s *Service) comentarioDelMovimiento(tx *gorm.DB, destino repositories.Destino, texto string, actor auth.Identity, descripcion string, menciones []int64) error {
	if texto == "" {
		return nil
	}

	_, err := s.escribirComentarioEnTx(tx, destino, texto, actor, descripcion, nil, menciones)

	return err
}

// mover hace el cambio de estado y sus efectos, todo en una transacción: **el estado, el comentario del
// cierre y lo que se sincronice con el principal, o nada** (decisión 81).
func (s *Service) mover(principal repositories.Ticket, internoActual repositories.InternalTicket, esInterno bool, to string, actor auth.Identity, comentario string, menciones []int64) error {
	return s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		ahora := s.now()

		if esInterno {
			interno, _, err := s.tickets.InternalByTicket(principal.ID)
			if err != nil {
				return err
			}

			desde := interno.State
			interno.State = to
			if to == repositories.StateResuelto {
				interno.ResolvedAt = &ahora
			}
			if to == repositories.StateCerrado {
				interno.ClosedAt = &ahora
			}

			if err := s.tickets.UpdateInternal(tx, interno); err != nil {
				return err
			}

			// El comentario va antes que la entrada del historial para que en la línea de tiempo se lea
			// primero el porqué y después el cierre.
			if err := s.comentarioDelMovimiento(tx, repositories.DestinoDeInterno(interno.ID), comentario, actor, "", menciones); err != nil {
				return err
			}

			if err := s.historialDeEstado(tx, repositories.DestinoDeInterno(interno.ID), &actor.ID, desde, to); err != nil {
				return err
			}

			// Las reglas que obligan a que los dos tickets no se contradigan: el interno manda cuando
			// se pone en espera (regla 2), cuando se resuelve (regla 5) y cuando se cierra sin resolver
			// (regla 8, la devolución a Soporte).
			return s.sincronizarConElPrincipal(tx, principal, interno, desde, to, actor)
		}

		desde := principal.State
		principal.State = to
		if to == repositories.StateResuelto {
			principal.ResolvedAt = &ahora
			// Resolver limpia el cierre anterior: un ticket reabierto no puede seguir diciendo que se
			// cerró, y el historial ya cuenta cuándo pasó cada cosa.
			principal.ClosedAt = nil
		}
		if to == repositories.StateCerrado {
			principal.ClosedAt = &ahora
		}

		if err := s.tickets.UpdateTicket(tx, principal); err != nil {
			return err
		}

		// La descripción del principal cuenta como un cuerpo de su hilo, como en un comentario normal:
		// el reajuste de observadores la necesita (decisión 63).
		if err := s.comentarioDelMovimiento(tx, repositories.DestinoDePrincipal(principal.ID), comentario, actor, principal.Description, menciones); err != nil {
			return err
		}

		return s.historialDeEstado(tx, repositories.DestinoDePrincipal(principal.ID), &actor.ID, desde, to)
	})
}

// sincronizarConElPrincipal aplica las reglas 2, 5 y 8: es lo que hace que el principal no se quede
// diciendo «lo tiene Desarrollo» cuando Desarrollo ya no lo tiene.
func (s *Service) sincronizarConElPrincipal(tx *gorm.DB, principal repositories.Ticket, interno repositories.InternalTicket, desde, to string, actor auth.Identity) error {
	devuelve := false

	switch {
	// Regla 2: si el interno espera a Soporte, el principal pasa a `en progreso` —lo que se espera es
	// cosa de Desarrollo—.
	case to == repositories.StateEnEspera:
		devuelve = true

	// Regla 5: cuando el interno se resuelve, el principal vuelve a Soporte, que es quien le explica
	// al usuario qué se hizo.
	case to == repositories.StateResuelto && desde != repositories.StateResuelto:
		devuelve = true

	// Regla 8: cerrar el interno **sin resolverlo** es la devolución a Soporte. Cerrarlo después de
	// resolverlo es sólo orden, y no toca al principal.
	case to == repositories.StateCerrado && desde != repositories.StateResuelto:
		devuelve = true
	}

	// Un principal **cerrado** no se reabre solo: eso sería automatismo, y aquí no lo hay
	// (docs/flujos.md, sección 10). El interno sigue su vida y Soporte decide si reabre.
	if !devuelve || principal.State == repositories.StateEnProgreso || principal.State == repositories.StateCerrado {
		return nil
	}

	anterior := principal.State
	principal.State = repositories.StateEnProgreso

	if err := s.tickets.UpdateTicket(tx, principal); err != nil {
		return err
	}

	// **El actor va nulo**: ese movimiento no lo pidió nadie, lo provoca el interno. En la tabla de
	// transiciones es «Sistema», y en el historial se lee como «el sistema».
	detalle := "lo provocó el ticket interno " + interno.Number

	return s.conversation.AddHistory(tx, repositories.Anotar(
		repositories.DestinoDePrincipal(principal.ID),
		nil,
		"estado",
		anterior,
		repositories.StateEnProgreso,
		detalle,
	))
}

// Escalate escala un principal: crea el interno o lo reabre, y deja el principal en `escalado`.
func (s *Service) Escalate(number, motivo string, actor auth.Identity) (Ticket, error) {
	// Escalar es de Soporte (docs/usuarios-y-permisos.md, sección 3): Desarrollo recibe lo escalado y
	// el Administrador mira.
	if actor.Role != auth.RoleSoporte {
		return Ticket{}, ErrForbidden
	}

	razon := strings.TrimSpace(motivo)
	if razon == "" {
		return Ticket{}, ErrReasonRequired
	}

	principal, err := s.tickets.ByNumber(number)
	if err != nil {
		return Ticket{}, traducirTicket(err)
	}

	// Un ticket cerrado no se escala: para volver a escalarlo, primero se reabre.
	if principal.State == repositories.StateCerrado {
		return Ticket{}, ErrClosed
	}

	interno, hay, err := s.tickets.InternalByTicket(principal.ID)
	if err != nil {
		return Ticket{}, err
	}

	config, err := s.config.TicketSettings()
	if err != nil {
		return Ticket{}, err
	}

	// El interno nuevo se reparte **si el Administrador lo ha encendido**: viene apagado.
	var responsable *auth.Account
	if !hay {
		responsable, err = s.repartir(config.InternalAssignment, auth.RoleDesarrollo, true)
		if err != nil {
			return Ticket{}, err
		}
	}

	numero := repositories.NumeroDeInterno(principal.Number)

	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		if hay {
			// El mismo ticket interno, no uno nuevo: vuelve a `en progreso` aunque estuviera cerrado
			// (regla 4 del modelo).
			interno.State = repositories.StateEnProgreso
			interno.ResolvedAt = nil
			interno.ClosedAt = nil

			if err := s.tickets.UpdateInternal(tx, interno); err != nil {
				return err
			}

			return s.historialDeEstado(tx, repositories.DestinoDeInterno(interno.ID), &actor.ID, "", repositories.StateEnProgreso)
		}

		creado, err := s.tickets.CreateInternal(tx, repositories.CreateInternalInput{
			TicketID:         principal.ID,
			Number:           numero,
			State:            repositories.StateNuevo,
			EscalationReason: razon,
			CreatedByID:      actor.ID,
			AssigneeID:       idDe(responsable),
		})
		if err != nil {
			return err
		}
		interno = creado

		return s.conversation.AddHistory(tx, repositories.Anotar(
			repositories.DestinoDeInterno(interno.ID),
			&actor.ID,
			"creado",
			"",
			repositories.StateNuevo,
			"",
		))
	})
	if err != nil {
		return Ticket{}, err
	}

	// El principal queda en `escalado`, y eso se apunta en **los dos** historiales.
	if principal.State != repositories.StateEscalado {
		anterior := principal.State
		principal.State = repositories.StateEscalado

		err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
			if err := s.tickets.UpdateTicket(tx, principal); err != nil {
				return err
			}

			return s.historialDeEstado(tx, repositories.DestinoDePrincipal(principal.ID), &actor.ID, anterior, repositories.StateEscalado)
		})
		if err != nil {
			return Ticket{}, err
		}
	}

	// **El motivo de un re-escalado va como comentario en el interno** (decisión 30): el motivo
	// original es la justificación de la primera vez y no se reescribe.
	//
	// El motivo es texto plano y no se sanea —a Soporte no se le puede rechazar su explicación por
	// llevar un `<`—, pero el comentario donde cae es HTML: se escapa al escribirlo, que es lo mismo que
	// hace la migración con los comentarios que ya existían.
	if hay {
		if _, err := s.escribirComentario(repositories.DestinoDeInterno(interno.ID), TextoComoHtml(razon), actor, "", nil, nil); err != nil {
			return Ticket{}, err
		}
	}

	s.avisarEscalado(numero, principal, razon, interno)

	// Escalar mueve **dos** tickets: el principal queda escalado y nace —o se reabre— el interno. La
	// última acción se pide para los dos. **El motivo del interno no se pide**: en un interno el motivo
	// es el del escalado, que ya está escrito (docs/modules/ai.md, decisión 5).
	s.pedirAlMotor(principal.Number, TipoUltimaAccion)
	s.pedirAlMotor(interno.Number, TipoUltimaAccion)

	return s.ticketDe(number, actor)
}

// Reopen reabre un principal cerrado y limpia las fechas de resolución y cierre.
//
// El historial conserva cuándo pasó cada cosa: lo que se limpia son las fechas que la pantalla enseña
// como «resuelto el…», que si no dirían una verdad vieja (docs/modules/tickets.md, sección 3.1).
func (s *Service) Reopen(number string, actor auth.Identity) (Ticket, error) {
	ticket, err := s.tickets.ByNumber(number)
	if err != nil {
		return Ticket{}, traducirTicket(err)
	}

	detalle, err := s.detalleDePrincipal(ticket, actor)
	if err != nil {
		return Ticket{}, err
	}

	// Reabren Soporte y el solicitante; Desarrollo no cierra ni reabre tickets (regla 7).
	puede := actor.Role == auth.RoleSoporte || (actor.Role == auth.RoleUsuario && esSuyo(actor, detalle.Ticket))
	if !puede {
		return Ticket{}, ErrForbidden
	}

	if ticket.State != repositories.StateCerrado {
		return Ticket{}, ErrTransitionNotAllowed
	}

	anterior := ticket.State
	ticket.State = repositories.StateEnProgreso
	ticket.ResolvedAt = nil
	ticket.ClosedAt = nil

	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.tickets.UpdateTicket(tx, ticket); err != nil {
			return err
		}

		return s.historialDeEstado(tx, repositories.DestinoDePrincipal(ticket.ID), &actor.ID, anterior, repositories.StateEnProgreso)
	})
	if err != nil {
		return Ticket{}, err
	}

	s.pedirAlMotor(number, TipoUltimaAccion)

	reabierto, err := s.ByNumber(number, actor)
	if err != nil {
		return Ticket{}, err
	}

	return reabierto.Ticket, nil
}

// puedeAsignar aplica la matriz: en el principal asigna Soporte, y en el interno Soporte y Desarrollo.
func (s *Service) puedeAsignar(actor auth.Identity, interno bool) bool {
	if interno {
		return actor.Role == auth.RoleSoporte || actor.Role == auth.RoleDesarrollo
	}

	return actor.Role == auth.RoleSoporte
}

// transicionExiste dice si ese movimiento está en la tabla, para cualquiera.
//
// Va aparte de los permisos a propósito: un movimiento que no existe es una errata —422—, y uno que
// existe pero no es de quien lo pide es un permiso —403—. Mezclarlos deja al frontend sin saber si
// está enseñando un botón de más o si alguien ha pedido algo que no le toca.
func transicionExiste(esInterno bool, desde, to string) bool {
	if esInterno {
		switch {
		case to == repositories.StateEnProgreso:
			// `cerrado → en progreso` **no** se hace por aquí: reabrir un interno es volver a escalar,
			// y tiene su propia acción.
			return desde == repositories.StateNuevo || desde == repositories.StateEnEspera
		case to == repositories.StateEnEspera:
			return desde == repositories.StateEnProgreso
		case to == repositories.StateResuelto:
			return desde == repositories.StateEnProgreso || desde == repositories.StateEnEspera
		case to == repositories.StateCerrado:
			return desde == repositories.StateNuevo || desde == repositories.StateEnProgreso ||
				desde == repositories.StateEnEspera || desde == repositories.StateResuelto
		}

		return false
	}

	switch {
	case to == repositories.StateEnProgreso:
		// Ni `escalado → en progreso` —lo provoca el interno, no una mano— ni `cerrado → en progreso`
		// —eso es reabrir, y reabrir tiene su propia acción— se hacen por aquí.
		return desde == repositories.StateNuevo || desde == repositories.StateEnEspera
	case to == repositories.StateEnEspera:
		return desde == repositories.StateEnProgreso
	case to == repositories.StateResuelto:
		return desde == repositories.StateEnProgreso || desde == repositories.StateEnEspera
	case to == repositories.StateCerrado:
		return desde == repositories.StateNuevo || desde == repositories.StateEnProgreso ||
			desde == repositories.StateEnEspera || desde == repositories.StateResuelto
	}

	// `escalado` y `reabierto` no se piden por aquí: tienen su propia acción.
	return false
}

// puedeMover es la tabla de transiciones, con quién puede cada fila.
func (s *Service) puedeMover(actor auth.Identity, ticket repositories.Ticket, esInterno bool, desde, to string) bool {
	if esInterno {
		return s.puedeMoverInterno(actor, desde, to)
	}

	return s.puedeMoverPrincipal(actor, ticket, desde, to)
}

func (s *Service) puedeMoverPrincipal(actor auth.Identity, ticket repositories.Ticket, desde, to string) bool {
	esSoporte := actor.Role == auth.RoleSoporte
	esSolicitante := actor.Role == auth.RoleUsuario && ticket.RequesterID == actor.ID

	switch to {
	case repositories.StateEnProgreso, repositories.StateEnEspera, repositories.StateResuelto:
		// Los movimientos de dentro del trabajo son de Soporte: el solicitante cierra y reabre lo suyo,
		// pero no lo pone en progreso ni lo da por resuelto.
		return esSoporte

	case repositories.StateCerrado:
		// Cerrar sin resolver se puede desde cualquier estado abierto, y también desde `resuelto`:
		// dupdos, pruebas o tickets creados por error.
		return esSoporte || esSolicitante
	}

	return false
}

func (s *Service) puedeMoverInterno(actor auth.Identity, desde, to string) bool {
	esDesarrollo := actor.Role == auth.RoleDesarrollo
	esSoporte := actor.Role == auth.RoleSoporte

	switch to {
	case repositories.StateEnProgreso:
		// Soporte responde a lo que Desarrollo preguntó: `en espera → en progreso` es de los dos.
		return esDesarrollo || (esSoporte && desde == repositories.StateEnEspera)

	case repositories.StateEnEspera:
		return esDesarrollo

	case repositories.StateResuelto:
		// Resolver el interno es de Desarrollo: es quien lo trabaja.
		return esDesarrollo

	case repositories.StateCerrado:
		return esDesarrollo || esSoporte
	}

	return false
}

// historialDeEstado apunta un cambio de estado. Si no cambia nada, no apunta nada.
func (s *Service) historialDeEstado(tx *gorm.DB, destino repositories.Destino, actorID *int64, desde, to string) error {
	if desde == to {
		return nil
	}

	evento := "estado"
	switch to {
	case repositories.StateResuelto:
		evento = "resuelto"
	case repositories.StateCerrado:
		evento = "cerrado"
	case repositories.StateEnProgreso:
		if desde == repositories.StateCerrado {
			evento = "reabierto"
		}
	}

	return s.conversation.AddHistory(tx, repositories.Anotar(destino, actorID, evento, desde, to, ""))
}

// estadoValido dice si ese estado existe para ese tipo de ticket.
func estadoValido(estado string, interno bool) bool {
	estados := repositories.EstadosDelPrincipal
	if interno {
		estados = repositories.EstadosDelInterno
	}

	for _, conocido := range estados {
		if conocido == estado {
			return true
		}
	}

	return false
}

// porNumero trae el ticket y, si el número es de un interno, también su principal.
func (s *Service) porNumero(number string) (repositories.Ticket, bool, bool, error) {
	numero := strings.TrimSpace(number)

	if strings.HasPrefix(numero, "INT-") {
		interno, err := s.tickets.InternalByNumber(numero)
		if err != nil {
			return repositories.Ticket{}, true, true, traducirTicket(err)
		}

		principal, err := s.tickets.ByID(interno.TicketID)
		if err != nil {
			return repositories.Ticket{}, true, true, traducirTicket(err)
		}

		return principal, true, true, nil
	}

	principal, err := s.tickets.ByNumber(numero)
	if err != nil {
		return repositories.Ticket{}, false, false, traducirTicket(err)
	}

	_, hay, err := s.tickets.InternalByTicket(principal.ID)
	if err != nil {
		return repositories.Ticket{}, false, false, err
	}

	return principal, false, hay, nil
}
