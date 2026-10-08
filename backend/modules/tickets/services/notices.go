package services

import (
	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// Los ocho avisos por correo (docs/flujos.md y docs/propósito-y-alcance.md, sección 5).
//
// **El correo es un aviso, no la acción**: si un aviso no sale, lo que el flujo cuenta ya ha pasado.
// Por eso todos van en segundo plano y ninguno puede tumbar la acción que los provoca
// (docs/arquitectura.md, sección 9).
//
// **Los observadores reciben los mismos correos de personal que el asignado** (decisión 61): el del
// reparto, el del escalado, el de «Desarrollo necesita algo» y el de «el ticket vuelve a Soporte». Los
// avisos que van **al solicitante** no se les duplica: son de quien abrió el ticket.

// avisarCreacion manda el aviso de ticket nuevo, según lo que tenga configurado el Administrador.
//
// Con `al_asignado` y sin responsable —no había ningún técnico activo— **no se avisa a nadie**: no
// hay a quién, y el ticket queda en la bandeja para que alguien lo coja. Queda en el log.
func (s *Service) avisarCreacion(config TicketSettings, numero, asunto string, solicitante auth.Account, responsable *auth.Account, principalID int64) {
	datos := map[string]string{
		"numero":      numero,
		"asunto":      asunto,
		"solicitante": solicitante.FullName(),
		"enlace":      s.enlaceDelTicket(numero),
	}

	// Los observadores del principal —los etiquetados en la descripción— reciben el mismo aviso de
	// personal que el asignado (decisión 61).
	observadores := s.observadoresDe(repositories.DestinoDePrincipal(principalID))

	switch config.MainNotification {
	case NotifyNobody:
		return

	case NotifyAssigned:
		personas := observadores
		if responsable != nil {
			personas = append([]auth.Account{*responsable}, observadores...)
		}
		if len(personas) == 0 {
			logs.LogWarning("el ticket " + numero + " se ha creado sin responsable y sin aviso: no hay ningún técnico activo")
			return
		}

		s.avisarPersonal(CorreoTicketCreado, datos, personas)

	default:
		equipo, err := s.accounts.ActiveByRole(auth.RoleSoporte)
		if err != nil {
			logs.LogWarning("no se ha podido leer el equipo al que avisar: " + err.Error())
			return
		}

		s.avisarPersonal(CorreoTicketCreado, datos, append(equipo, observadores...))
	}
}

// avisarEscalado avisa a Desarrollo de que le ha llegado trabajo. Sale también al re-escalar: el
// trabajo ha vuelto a aparecer, y eso es lo que el aviso significa.
func (s *Service) avisarEscalado(numeroInterno string, principal repositories.Ticket, motivo string, interno repositories.InternalTicket) {
	datos := map[string]string{
		"numero": numeroInterno,
		"asunto": principal.Subject,
		"motivo": motivo,
		"enlace": s.enlaceDelTicket(numeroInterno),
	}

	// Los observadores del interno reciben lo mismo que su asignado.
	observadores := s.observadoresDe(repositories.DestinoDeInterno(interno.ID))

	if interno.AssigneeID != nil {
		if responsable, err := s.accounts.ByID(*interno.AssigneeID); err == nil {
			s.avisarPersonal(CorreoTicketEscalado, datos, append([]auth.Account{responsable}, observadores...))
			return
		}
	}

	equipo, err := s.accounts.ActiveByRole(auth.RoleDesarrollo)
	if err != nil {
		logs.LogWarning("no se ha podido leer el equipo al que avisar: " + err.Error())
		return
	}

	s.avisarPersonal(CorreoTicketEscalado, datos, append(equipo, observadores...))
}

// avisarMovimiento manda el aviso que le toca al movimiento que se acaba de hacer.
//
// `desde` importa: cerrar el interno **después de resolverlo** es una tarea de orden y no avisa a
// nadie, mientras que cerrarlo sin resolverlo es la devolución a Soporte y sí avisa
// (docs/modules/tickets.md, sección 3.2).
func (s *Service) avisarMovimiento(principal repositories.Ticket, interno repositories.InternalTicket, esInterno bool, desde, to string) {
	switch {
	// El principal espera al usuario: es el único aviso que existe para pedirle algo.
	case !esInterno && to == repositories.StateEnEspera:
		s.avisarAlUsuario(principal, CorreoTicketEsperaUsuario)

	case !esInterno && to == repositories.StateResuelto:
		s.avisarAlUsuario(principal, CorreoTicketResuelto)

	case !esInterno && to == repositories.StateCerrado:
		s.avisarAlUsuario(principal, CorreoTicketCerrado)

	// Desarrollo necesita algo de Soporte: se avisa a quien escaló, que es quien lo pidió, y a los
	// observadores del interno, que siguen ese hilo.
	case esInterno && to == repositories.StateEnEspera:
		quienEscalo := interno.CreatedByID
		s.avisarASoporte(CorreoTicketEsperaSoporte, interno.Number, principal, &quienEscalo, s.observadoresDe(repositories.DestinoDeInterno(interno.ID)))

	// El caso vuelve a Soporte: Desarrollo lo resolvió —y hay que explicárselo al usuario— o lo cerró
	// sin resolverlo (la devolución, regla 8). El aviso va al responsable del principal, así que los
	// observadores que lo reciben son **los del principal**.
	case esInterno && to == repositories.StateResuelto:
		s.avisarASoporte(CorreoTicketVuelveASoporte, principal.Number, principal, principal.AssigneeID, s.observadoresDe(repositories.DestinoDePrincipal(principal.ID)))

	case esInterno && to == repositories.StateCerrado && desde != repositories.StateResuelto:
		s.avisarASoporte(CorreoTicketVuelveASoporte, principal.Number, principal, principal.AssigneeID, s.observadoresDe(repositories.DestinoDePrincipal(principal.ID)))
	}
}

// avisarAlUsuario manda al solicitante el aviso, en el idioma global.
//
// **A los observadores no se les manda**: son avisos de quien abrió el ticket, y duplicarlos sería
// ruido (decisión 61).
func (s *Service) avisarAlUsuario(principal repositories.Ticket, plantilla string) {
	solicitante, err := s.accounts.ByID(principal.RequesterID)
	if err != nil {
		logs.LogWarning("no se ha podido avisar del ticket " + principal.Number + ": su solicitante no está")
		return
	}

	s.mailer.SendAsync(plantilla, s.idiomaGlobal(), []string{solicitante.Email}, map[string]string{
		"numero": principal.Number,
		"asunto": principal.Subject,
		"nombre": solicitante.FullName(),
		"enlace": s.enlaceDelTicket(principal.Number),
	})
}

// avisarASoporte avisa hacia dentro del nivel 2.
//
// Va a **una persona** y no a todo el equipo: «ha vuelto a tu bandeja» es para quien la lleva —el
// responsable del principal, o quien escaló— y un correo para todos convierte el aviso en ruido
// (decisión 32). Si esa cuenta ya no está activa, se avisa al equipo entero en su lugar: es peor que
// nadie se entere que recibir un correo de más.
func (s *Service) avisarASoporte(plantilla, numero string, principal repositories.Ticket, personaID *int64, observadores []auth.Account) {
	datos := map[string]string{
		"numero": numero,
		"asunto": principal.Subject,
		"enlace": s.enlaceDelTicket(numero),
	}

	if personaID != nil {
		persona, err := s.accounts.ByID(*personaID)
		if err == nil && persona.IsActive {
			s.avisarPersonal(plantilla, datos, append([]auth.Account{persona}, observadores...))
			return
		}
	}

	equipo, err := s.accounts.ActiveByRole(auth.RoleSoporte)
	if err != nil {
		logs.LogWarning("no se ha podido leer el equipo al que avisar: " + err.Error())
		return
	}

	s.avisarPersonal(plantilla, datos, append(equipo, observadores...))
}

// avisarPersonal manda una plantilla en el idioma global y **sin repetir a nadie**: el asignado puede
// ser también observador, y no puede recibir dos veces el mismo
// correo (docs/modules/tickets.md, decisión 61).
func (s *Service) avisarPersonal(plantilla string, datos map[string]string, personas []auth.Account) {
	direcciones := []string{}
	visto := map[int64]bool{}

	for i := range personas {
		persona := personas[i]
		if visto[persona.ID] || !persona.IsActive {
			continue
		}
		visto[persona.ID] = true

		direcciones = append(direcciones, persona.Email)
	}

	if len(direcciones) > 0 {
		s.mailer.SendAsync(plantilla, s.idiomaGlobal(), direcciones, datos)
	}
}

func (s *Service) idiomaGlobal() string {
	if s.language != nil {
		if idioma, err := s.language.Language(); err == nil && idioma == "en" {
			return "en"
		}
	}
	return "es"
}

// enlaceDelTicket arma la dirección que va en el correo: es lo que se pega para abrir el ticket.
func (s *Service) enlaceDelTicket(numero string) string {
	return s.mailer.Link("/tickets/" + numero)
}
