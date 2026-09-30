package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// Los dos campos que redacta el motor de IA (`docs/modules/ai.md`). Son los mismos nombres que usa
// su tabla, y se declaran aquí para no escribirlos de dos maneras.
const (
	TipoMotivo       = "motivo"
	TipoUltimaAccion = "ultima_accion"
)

// Resumen es un texto del motor, con lo que hace falta para pintarlo: su estado, los dos idiomas y la
// clave del error cuando no se pudo escribir.
//
// **Los dos idiomas viajan juntos** porque el motor los devuelve en la misma respuesta (decisión 3) y
// porque la lista se pinta en el idioma de quien mira: traducir al vuelo sería pedir dos veces lo
// mismo.
type Resumen struct {
	Estado   string
	Es       string
	En       string
	ErrorKey string
}

// Resumenes son los dos campos de un ticket.
type Resumenes struct {
	Motivo       Resumen
	UltimaAccion Resumen
}

// EntradaDeResumen es el encargo que se le hace al motor: un ticket, un campo y el texto ya armado.
type EntradaDeResumen struct {
	Numero string
	// Tipo es `motivo` o `ultima_accion`.
	Tipo string
	// Texto lo arma este módulo, que es quien tiene el ticket: asunto, descripción, adjuntos y
	// conversación, ya en texto plano (`textoParaElMotor`).
	Texto string
}

// Insights es lo que este módulo necesita del motor de IA.
//
// Se declara aquí, en quien lo usa, y lo cumple el módulo `ai` sin saber que existe: así ninguno de
// los dos importa al otro y quien los une es `main.go` (`docs/arquitectura.md`, sección 4). La misma
// forma que tienen `Accounts` y `Mailer`.
type Insights interface {
	// Pedir apunta un resumen y **vuelve enseguida**: no espera al motor (decisión 2).
	Pedir(EntradaDeResumen)
	// De devuelve lo que hay de esos tickets, **en bloque**: una consulta por página, no una por fila
	// (decisión 9).
	De(numeros []string) (map[string]Resumenes, error)
}

// SetInsights engancha el motor de IA. Si no se llama —o se llama con nada—, los dos campos se quedan
// sin texto y **todo lo demás funciona igual**: la mesa de ayuda no depende del motor.
func (s *Service) SetInsights(insights Insights) {
	s.insights = insights
}

// pedirAlMotor le pide al motor los resúmenes de ese ticket, con el texto ya armado.
//
// **No devuelve error a propósito**: quien crea un ticket, comenta o lo mueve no puede quedarse
// esperando a un contenedor que puede estar caído. Si el texto no se puede armar, o el motor no está,
// se deja dicho en el log y el campo se queda como estaba (decisión 2).
func (s *Service) pedirAlMotor(numero string, tipos ...string) {
	if s.insights == nil {
		return
	}

	texto, err := s.textoParaElMotor(numero)
	if err != nil {
		logs.LogWarning("no se pudo armar el texto para el resumen del ticket " + numero)
		return
	}

	for _, tipo := range tipos {
		s.insights.Pedir(EntradaDeResumen{Numero: numero, Tipo: tipo, Texto: texto})
	}
}

// TextoParaElMotor arma el texto de un ticket, y lo usa el módulo `ai` para **volver a pedir** un
// resumen que quedó a medias cuando el backend se reinició: la cola vive en memoria y el texto no se
// guarda, así que hay que volver a construirlo (docs/modules/ai.md, decisión 10).
//
// Se declara aquí y `ai` lo cumple por su cuenta: ninguno de los dos módulos importa al otro.
func (s *Service) TextoParaElMotor(numero string) (string, error) {
	return s.textoParaElMotor(numero)
}

// textoParaElMotor arma lo que lee el motor: **el ticket en texto plano**.
//
// Lo que se manda y por qué está en `docs/modules/ai.md`, sección 3: el asunto, la descripción y la
// conversación —comentarios e historial, de lo más antiguo a lo más reciente—, **el HTML convertido a
// texto** y **los adjuntos por su nombre**, nunca su contenido. El recorte, si el texto es largo, lo
// hace el módulo `ai` (decisión 6), que es quien conoce el tope: aquí se manda lo que hay.
func (s *Service) textoParaElMotor(numero string) (string, error) {
	principal, esInterno, hay, err := s.porNumero(numero)
	if err != nil {
		return "", err
	}

	destino := repositories.DestinoDePrincipal(principal.ID)
	asunto := principal.Subject
	descripcion := principal.Description
	motivo := ""

	if esInterno {
		interno, _, err := s.tickets.InternalByTicket(principal.ID)
		if err != nil {
			return "", err
		}

		destino = repositories.DestinoDeInterno(interno.ID)
		// Un interno no tiene asunto ni descripción propios: lo que cuenta es el motivo del escalado.
		asunto = principal.Subject
		descripcion = interno.EscalationReason
		motivo = interno.EscalationReason
	}

	comentarios, adjuntos, historial, err := s.conversacionDe(destino)
	if err != nil {
		return "", err
	}

	// Los nombres de las personas que aparecen: para el motor, «Soporte pidió una captura» dice más que
	// «la cuenta 42 pidió una captura». Son los mismos datos que ya se ven en el ticket.
	ids := []int64{principal.RequesterID, principal.CreatedByID}
	if principal.AssigneeID != nil {
		ids = append(ids, *principal.AssigneeID)
	}
	for i := range comentarios {
		ids = append(ids, comentarios[i].AuthorID)
	}
	for i := range historial {
		if historial[i].ActorID != nil {
			ids = append(ids, *historial[i].ActorID)
		}
	}

	porID, err := s.accounts.ByIDs(ids)
	if err != nil {
		return "", err
	}

	var texto strings.Builder
	fmt.Fprintf(&texto, "Asunto: %s\n", asunto)
	fmt.Fprintf(&texto, "Descripción: %s\n", TextoPlano(descripcion))

	if hay {
		fmt.Fprintf(&texto, "Estado: %s\n", principal.State)
	}
	if motivo != "" {
		fmt.Fprintf(&texto, "Motivo del escalado: %s\n", TextoPlano(motivo))
	}

	if nombres := nombresDe(adjuntos); nombres != "" {
		fmt.Fprintf(&texto, "Adjuntos: %s\n", nombres)
	}

	texto.WriteString("\nConversación (de lo más antiguo a lo más reciente):\n")

	// **Comentarios e historial en una sola línea de tiempo**, en orden: es lo que hace que la última
	// línea del texto sea, de verdad, la última acción.
	for _, renglon := range lineasDe(comentarios, historial, porID) {
		texto.WriteString("- ")
		texto.WriteString(renglon)
		texto.WriteString("\n")
	}

	return texto.String(), nil
}

// nombresDe junta los nombres de los adjuntos, sin repetir y en el orden en que se subieron.
func nombresDe(adjuntos []repositories.Attachment) string {
	vistos := make(map[string]bool, len(adjuntos))
	nombres := make([]string, 0, len(adjuntos))

	for _, adjunto := range adjuntos {
		if adjunto.Filename == "" || vistos[adjunto.Filename] {
			continue
		}

		vistos[adjunto.Filename] = true
		nombres = append(nombres, adjunto.Filename)
	}

	return strings.Join(nombres, ", ")
}

// lineasDe intercala los comentarios y lo que hizo el sistema, como la línea de tiempo de la pantalla.
func lineasDe(
	comentarios []repositories.Comment,
	historial []repositories.HistoryEntry,
	porID map[int64]auth.Account,
) []string {
	type conFecha struct {
		cuando time.Time
		texto  string
	}

	renglones := make([]conFecha, 0, len(comentarios)+len(historial))

	for i := range comentarios {
		comentario := comentarios[i]
		cuerpo := TextoPlano(comentario.Body)
		if comentario.DeletedAt != nil {
			cuerpo = "(comentario borrado)"
		}

		renglones = append(renglones, conFecha{
			cuando: comentario.CreatedAt,
			texto:  fmt.Sprintf("%s (comentario): %s", quienEs(&comentario.AuthorID, porID), cuerpo),
		})
	}

	for i := range historial {
		entrada := historial[i]

		detalle := ""
		if entrada.Detail != nil {
			detalle = *entrada.Detail
		}

		renglones = append(renglones, conFecha{
			cuando: entrada.CreatedAt,
			texto:  fmt.Sprintf("%s (%s): %s", quienEs(entrada.ActorID, porID), entrada.Event, detalle),
		})
	}

	sort.SliceStable(renglones, func(i, j int) bool { return renglones[i].cuando.Before(renglones[j].cuando) })

	lineas := make([]string, 0, len(renglones))
	for _, renglon := range renglones {
		lineas = append(lineas, strings.TrimSpace(renglon.texto))
	}

	return lineas
}

// quienEs es el nombre de esa cuenta, o «el sistema» cuando la entrada no tiene autor.
func quienEs(id *int64, porID map[int64]auth.Account) string {
	if id == nil {
		return "el sistema"
	}

	cuenta, hay := porID[*id]
	if !hay {
		return "alguien"
	}

	return cuenta.FullName()
}

// PedirResumenes vuelve a pedir los dos campos, a mano.
//
// Es el botón «Regenerar» de la ficha: lo pueden pulsar **Soporte y Desarrollo**, que son los que
// trabajan el ticket; el Administrador mira sin botones y el usuario no tiene por qué ver esto
// (docs/modules/ai.md, decisión 5). Devuelve el ticket ya leído, para que la pantalla se pinte con lo
// que hay de verdad.
func (s *Service) PedirResumenes(number string, actor auth.Identity) (Ticket, error) {
	if actor.Role != auth.RoleSoporte && actor.Role != auth.RoleDesarrollo {
		return Ticket{}, ErrForbidden
	}

	// Se comprueba que el ticket existe —y que esa persona puede verlo— antes de encolar nada.
	detalle, err := s.ByNumber(number, actor)
	if err != nil {
		return Ticket{}, err
	}

	// **En un interno no se pide el motivo**: su motivo es el del escalado, que ya está escrito
	// (docs/modules/ai.md, decisión 5). Se pide sólo la última acción.
	tipos := []string{TipoMotivo, TipoUltimaAccion}
	if detalle.Ticket.Internal {
		tipos = []string{TipoUltimaAccion}
	}

	s.pedirAlMotor(detalle.Ticket.Number, tipos...)

	return detalle.Ticket, nil
}

// pegarResumenes escribe los dos campos en una lista de tickets, **de una sola consulta**.
//
// Recibe la lista por su sitio en memoria —un `[]Ticket`— y no una copia: es lo que hace que el mismo
// trabajo valga para una página entera y para un solo ticket.
func (s *Service) pegarResumenes(tickets []Ticket) {
	if len(tickets) == 0 {
		return
	}

	numeros := make([]string, 0, len(tickets))
	for i := range tickets {
		numeros = append(numeros, tickets[i].Number)
	}

	resumenes := s.resumenesDe(numeros)
	for i := range tickets {
		if resumen, hay := resumenes[tickets[i].Number]; hay {
			tickets[i].Resumen = resumen
		}
	}
}

// resumenesDe lee los dos campos de una lista de tickets, **en una sola consulta**.
//
// Va aparte de `aTicket` porque la lista se arma en dos pasadas: primero los tickets y después sus
// resúmenes, para no preguntar una vez por fila (decisión 9). Si el motor no está, devuelve el mapa
// vacío y los campos se quedan sin texto.
func (s *Service) resumenesDe(numeros []string) map[string]Resumenes {
	if s.insights == nil || len(numeros) == 0 {
		return nil
	}

	resumenes, err := s.insights.De(numeros)
	if err != nil {
		logs.LogWarning("no se pudieron leer los resúmenes de los tickets")
		return nil
	}

	return resumenes
}
