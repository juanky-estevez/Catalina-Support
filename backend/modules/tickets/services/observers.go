package services

import (
	"errors"
	"strconv"
	"strings"

	"github.com/juanky-estevez/go-logs"
	"golang.org/x/net/html"
	"gorm.io/gorm"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// MencionesDelCuerpo devuelve los identificadores de las personas etiquetadas en un cuerpo HTML, **sin
// repetir y en el orden en que aparecen**.
//
// La mención es un `span` con su `data-mencion` (docs/modules/tickets.md, decisión 58), y se extrae con
// el mismo analizador que el saneador y no con una expresión regular: aquí el texto ya pasó la lista
// blanca, pero el HTML tiene sus reglas y adivinar dónde acaba cada atributo con patrones es
// exactamente lo que deja pasar lo que no debe.
//
// Va aparte del saneado a propósito: el saneador dice **si el texto se puede guardar**, y esto dice **a
// quién nombra**, que es lo que decide quién pasa a observar el ticket.
func MencionesDelCuerpo(cuerpo string) []int64 {
	var (
		ids   []int64
		visto = map[int64]bool{}
	)

	analizador := html.NewTokenizer(strings.NewReader(cuerpo))

	for {
		switch analizador.Next() {
		case html.ErrorToken:
			return ids

		case html.StartTagToken, html.SelfClosingTagToken:
			nombre, hayAtributos := analizador.TagName()
			esMencion := string(nombre) == "span"

			// Los atributos se recorren **siempre**, aunque la etiqueta no sea un `span`: dejar el
			// analizador a mitad de una etiqueta es lo que haría que la siguiente lectura empezara donde
			// no toca.
			for hayAtributos {
				var clave, valor []byte
				clave, valor, hayAtributos = analizador.TagAttr()

				if !esMencion || string(clave) != atributoDeLaMencion {
					continue
				}

				identificador := string(valor)
				if !mencionValida.MatchString(identificador) {
					continue
				}

				id, err := strconv.ParseInt(identificador, 10, 64)
				if err != nil || visto[id] {
					continue
				}

				visto[id] = true
				ids = append(ids, id)
			}
		}
	}
}

// diferencia devuelve los identificadores de `nuevos` que **no estaban** en `anteriores`: son las
// menciones nuevas de un cuerpo, y sólo esas hacen que alguien vuelva a observar
// (docs/modules/tickets.md, decisión 63). Los dos vienen sin repetir.
func diferencia(nuevos, anteriores []int64) []int64 {
	antes := make(map[int64]bool, len(anteriores))
	for _, id := range anteriores {
		antes[id] = true
	}

	var nuevosIds []int64
	for _, id := range nuevos {
		if !antes[id] {
			nuevosIds = append(nuevosIds, id)
		}
	}

	return nuevosIds
}

// validarMenciones aplica la decisión 59: **Soporte y Desarrollo etiquetan, y sólo a técnicos y
// desarrolladores activos**, y una que no valga rechaza el cuerpo entero con `tickets.mention.notAllowed`
// —igual que el saneador rechaza el HTML que no admite—.
//
// El usuario no etiqueta: no ve la lista de técnicos ni tiene por qué. **Pero no se le rechaza un texto
// que no ha escrito él**: si edita la descripción de un ticket que abrió Soporte en su nombre, y esa
// descripción **ya llevaba una mención**, el cuerpo se guarda —no ha etiquetado a nadie, ha conservado
// lo que había—. Lo que se rechaza es que **aparezca una mención nueva** en un cuerpo suyo.
func (s *Service) validarMenciones(menciones, anteriores []int64, actor auth.Identity) error {
	if len(menciones) == 0 {
		return nil
	}

	if actor.Role != auth.RoleSoporte && actor.Role != auth.RoleDesarrollo {
		if len(diferencia(menciones, anteriores)) > 0 {
			return ErrMentionNotAllowed
		}

		return nil
	}

	cuentas, err := s.accounts.ByIDs(menciones)
	if err != nil {
		return err
	}

	for _, id := range menciones {
		cuenta, hay := cuentas[id]
		if !hay || !cuenta.IsActive {
			return ErrMentionNotAllowed
		}
		if cuenta.Role != auth.RoleSoporte && cuenta.Role != auth.RoleDesarrollo {
			return ErrMentionNotAllowed
		}
	}

	return nil
}

// sincronizarObservadores aplica el reparto de la decisión 63, **dentro de la transacción que guarda
// el cuerpo**:
//
//  1. se añaden las menciones nuevas —las que no estaban en el cuerpo anterior—;
//  2. se quitan los observadores que **ya no menciona ningún cuerpo del hilo**, leyendo la
//     descripción del principal y todos sus comentarios con el cuerpo nuevo ya escrito.
//
// Lo segundo es lo que hace que valgan las dos reglas a la vez: **borrar la mención editando el
// comentario quita al observador**, y **quitar a mano manda** —a quien se quitó no lo devuelve un
// guardado posterior, porque su mención vieja no es nueva; sólo vuelve con una mención nueva, que es
// una llamada deliberada—.
//
// `anteriores` se ignora para el reajuste: lo que decide quién sobra es el texto que hay **ahora** en
// el hilo, no lo que había antes en ese cuerpo.
func (s *Service) sincronizarObservadores(tx *gorm.DB, destino repositories.Destino, descripcion string, anteriores, nuevos []int64, actorID int64) error {
	for _, id := range diferencia(nuevos, anteriores) {
		if _, err := s.conversation.AddObserver(tx, repositories.TicketObserver{
			TicketID:         destino.TicketID,
			InternalTicketID: destino.InternalTicketID,
			AccountID:        id,
			AddedByID:        actorID,
		}); err != nil {
			return err
		}
	}

	vigentes := map[int64]bool{}
	if destino.TicketID != nil {
		// La descripción es un cuerpo del hilo del principal: una mención en ella mantiene al
		// observador aunque **ningún comentario** la vuelva a nombrar.
		for _, id := range MencionesDelCuerpo(descripcion) {
			vigentes[id] = true
		}
	}

	comentarios, err := s.conversation.CommentsIn(tx, destino)
	if err != nil {
		return err
	}
	for _, comentario := range comentarios {
		for _, id := range MencionesDelCuerpo(comentario.Body) {
			vigentes[id] = true
		}
	}

	observadores, err := s.conversation.Observers(tx, destino)
	if err != nil {
		return err
	}

	for _, observador := range observadores {
		if vigentes[observador.AccountID] {
			continue
		}

		if _, err := s.conversation.RemoveObserver(tx, observador.AccountID, destino); err != nil {
			return err
		}
	}

	return nil
}

// AddObserver añade **a mano** a una persona como observadora de un ticket.
//
// Hasta el 2026-09-28 sólo se llegaba por una mención en un comentario; con esto, la ficha tiene su
// botón de añadir, y **quien llega así queda igual que quien llegó por una mención**: todos son
// observadores y la lista no distingue de dónde vino cada uno (docs/modules/tickets.md, sección 2.3.1
// y decisión 74). Lo puede hacer **cualquier técnico o desarrollador**, la misma regla que quitarlos
// (decisión 63), y **la cuenta que se añade tiene que ser un técnico o un desarrollador activo**,
// como para etiquetar: es la misma regla y por eso se rechaza con la misma clave
// (`tickets.mention.notAllowed`, decisión 59).
//
// **Es idempotente**: añadir a quien ya observa no es un error, la lista queda igual y **no deja una
// segunda entrada en el historial**. Sólo se anota cuando de verdad se añade, con el evento
// `observador_anadido` y el nombre de quien entra en el `detail` —el mismo estilo que quitar, que usa
// «quién quitó a quién»—.
//
// El ticket puede ser un principal o un interno, y los observadores son de ese hilo: los del interno
// son del interno y los del principal, del principal (sección 2.3.1).
func (s *Service) AddObserver(number string, accountID int64, actor auth.Identity) (Detail, error) {
	if actor.Role != auth.RoleSoporte && actor.Role != auth.RoleDesarrollo {
		return Detail{}, ErrForbidden
	}

	principal, esInterno, _, err := s.porNumero(number)
	if err != nil {
		return Detail{}, err
	}

	destino, err := s.destinoPorTicket(principal.ID, esInterno)
	if err != nil {
		return Detail{}, err
	}

	// La misma comprobación que el etiquetado (decisión 59): cuenta que existe, activa y de uno de
	// los dos papeles que observan. Todo lo que no valga devuelve la misma clave, porque es la misma
	// regla —no se distinguen «no existe» de «no vale» para quien lo pide—.
	cuenta, err := s.accounts.ByID(accountID)
	if err != nil || !cuenta.IsActive ||
		(cuenta.Role != auth.RoleSoporte && cuenta.Role != auth.RoleDesarrollo) {
		return Detail{}, ErrMentionNotAllowed
	}

	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		nuevo, err := s.conversation.AddObserver(tx, repositories.TicketObserver{
			TicketID:         destino.TicketID,
			InternalTicketID: destino.InternalTicketID,
			AccountID:        accountID,
			AddedByID:        actor.ID,
		})
		if err != nil {
			return err
		}

		// Ya observaba: no pasó nada, así que no se anota nada.
		if !nuevo {
			return nil
		}

		return s.conversation.AddHistory(tx, repositories.Anotar(
			destino,
			&actor.ID,
			"observador_anadido",
			"",
			"",
			cuenta.FullName(),
		))
	})
	if err != nil {
		return Detail{}, err
	}

	// Se devuelve la ficha leída —con la lista de observadores ya puesta— para que el controlador
	// arme la respuesta sin volver a leer nada.
	return s.ByNumber(number, actor)
}

// RemoveObserver quita a una persona de los observadores de un ticket.
//
// Lo puede hacer **cualquier técnico o desarrollador**, no sólo quien la etiquetó, y deja su entrada
// en el historial («quién quitó a quién», con el nombre del quitado en el `detail`), como `Assign` con
// «asignado» (docs/modules/tickets.md, decisión 63). El `{id}` es el de la **cuenta**, no el de la
// fila: quien se quita del ticket es una persona.
//
// A quien se quita **sigue nombrado en el comentario**: no se reescribe el texto de nadie. La mención
// se lee como lo que es —una llamada que se hizo— y la lista de observadores dice quién lo sigue ahora.
func (s *Service) RemoveObserver(number string, accountID int64, actor auth.Identity) error {
	if actor.Role != auth.RoleSoporte && actor.Role != auth.RoleDesarrollo {
		return ErrForbidden
	}

	principal, esInterno, _, err := s.porNumero(number)
	if err != nil {
		return err
	}

	destino, err := s.destinoPorTicket(principal.ID, esInterno)
	if err != nil {
		return err
	}

	// El nombre se lee antes de la transacción: es lo que va en el historial, y una cuenta que no
	// existe no observa nada, así que el identificador no vale.
	cuenta, err := s.accounts.ByID(accountID)
	if err != nil {
		return ErrObserverNotFound
	}

	err = s.tickets.DB().Transaction(func(tx *gorm.DB) error {
		filas, err := s.conversation.RemoveObserver(tx, accountID, destino)
		if err != nil {
			return err
		}
		if filas == 0 {
			return repositories.ErrObserverNotFound
		}

		return s.conversation.AddHistory(tx, repositories.Anotar(
			destino,
			&actor.ID,
			"observador",
			"",
			"",
			cuenta.FullName(),
		))
	})
	if errors.Is(err, repositories.ErrObserverNotFound) {
		return ErrObserverNotFound
	}

	return err
}

// observadoresDe lee los observadores de un hilo para los avisos: es la lista de personas que reciben
// los mismos correos de personal que el asignado (docs/modules/tickets.md, decisión 61).
func (s *Service) observadoresDe(destino repositories.Destino) []auth.Account {
	filas, err := s.conversation.Observers(s.tickets.DB(), destino)
	if err != nil {
		logs.LogWarning("no se han podido leer los observadores del ticket: " + err.Error())
		return nil
	}
	if len(filas) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(filas))
	for _, fila := range filas {
		ids = append(ids, fila.AccountID)
	}

	porID, err := s.accounts.ByIDs(ids)
	if err != nil {
		logs.LogWarning("no se han podido leer las cuentas de los observadores: " + err.Error())
		return nil
	}

	cuentas := make([]auth.Account, 0, len(filas))
	for _, fila := range filas {
		if cuenta, hay := porID[fila.AccountID]; hay {
			cuentas = append(cuentas, cuenta)
		}
	}

	return cuentas
}

// avisarEtiquetado manda el correo `ticket.mentioned` **una vez por persona** a quien se acaba de
// etiquetar.
//
// Lleva el número, el asunto y el enlace —como los demás avisos— y **nunca el texto del ticket ni el
// del comentario** (docs/modules/tickets.md, sección 5). El asunto sí va: la plantilla lo lleva en su
// asunto y en su cuerpo, y no es el texto de nadie.
func (s *Service) avisarEtiquetado(numero, asunto string, ids []int64) {
	if len(ids) == 0 || s.mailer == nil {
		return
	}

	cuentas, err := s.accounts.ByIDs(ids)
	if err != nil {
		logs.LogWarning("no se ha podido avisar del etiquetado del ticket " + numero + ": " + err.Error())
		return
	}

	datos := map[string]string{
		"numero": numero,
		"asunto": asunto,
		"enlace": s.enlaceDelTicket(numero),
	}

	for _, id := range ids {
		cuenta, hay := cuentas[id]
		if !hay || !cuenta.IsActive {
			continue
		}

		s.mailer.SendAsync(CorreoTicketEtiquetado, s.idiomaGlobal(), []string{cuenta.Email}, datos)
	}
}
