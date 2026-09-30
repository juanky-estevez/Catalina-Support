// Package controllers es la capa HTTP del módulo tickets. Los permisos gruesos los pone el middleware
// sobre la ruta y los finos —los que dependen de quién es cada ticket— los comprueba el servicio, que
// es quien ve el ticket y a quien lo pide (docs/modules/tickets.md, sección 5).
package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/tickets/dtos"
	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/modules/tickets/services"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/httpx"
)

// KeyInternal es lo que se responde cuando el fallo es nuestro.
const KeyInternal = "error interno"

// TicketController atiende los endpoints de tickets.
type TicketController struct {
	service *services.Service
}

// NewTicketController construye el controlador.
func NewTicketController(service *services.Service) *TicketController {
	return &TicketController{service: service}
}

// List devuelve la lista de tickets —la bandeja o una de las dos listas del todo—, con sus filtros y
// su paginación.
func (c *TicketController) List(w http.ResponseWriter, r *http.Request) {
	consulta := r.URL.Query()

	filtros := repositories.Filtros{
		Type:  strings.TrimSpace(consulta.Get("type")),
		State: strings.TrimSpace(consulta.Get("state")),
		Query: consulta.Get("q"),
		// `view` es el chip de «Mis tickets»: `assigned`, `watching` o vacío (todo lo mío). Sólo
		// matiza «lo mío», así que por sí solo también lo pide (docs/modules/tickets.md, decisión 62).
		View: strings.TrimSpace(consulta.Get("view")),
		// `category` es el identificador de la categoría y `tag` una etiqueta: los dos chips de filtro
		// (decisión 68). La etiqueta se normaliza en el servicio, que es quien conoce la regla.
		Tag: strings.TrimSpace(consulta.Get("tag")),
	}

	// `category` se filtra por identificador, que es lo que manda el chip. Un valor que no es un
	// número no filtra: no hay categoría que buscar, y la lista sigue siendo la de siempre.
	if valor := strings.TrimSpace(consulta.Get("category")); valor != "" {
		if id, err := strconv.ParseInt(valor, 10, 64); err == nil {
			filtros.CategoryID = &id
		}
	}

	pagina, _ := strconv.Atoi(consulta.Get("page"))
	porPagina, _ := strconv.Atoi(consulta.Get("perPage"))
	filtros.Page = pagina
	filtros.PerPage = porPagina

	// `mine=1` es **«lo mío»**: lo asignado a quien mira, lo que abrió, aquello donde ha comentado y
	// donde le han etiquetado. Lo pide la bandeja de los tres papeles (docs/modules/tickets.md,
	// decisiones 51 y 62); `view` lo pide por sí solo, porque sin «lo mío» no matiza nada.
	if consulta.Get("mine") == "1" || filtros.View != "" {
		yo := auth.MustFromContext(r.Context()).ID
		filtros.Mine = &yo
	}

	resultado, err := c.service.List(filtros, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewTicketsResponse(resultado))
}

// Insights vuelve a pedirle al motor de IA el motivo y la última acción de un ticket.
//
// Es el botón «Regenerar» de la ficha, y el endpoint es de **este** módulo y no de `ai`: la pantalla de
// tickets sólo puede hablar con su propia API, y es este módulo quien decide a quién se lo pide
// (docs/modules/ai.md, sección 5).
func (c *TicketController) Insights(w http.ResponseWriter, r *http.Request) {
	ticket, err := c.service.PedirResumenes(r.PathValue("number"), auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.TicketEnvelope{Ticket: dtos.NewTicketResponse(ticket)})
}

// Create da de alta un ticket principal.
func (c *TicketController) Create(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	ticket, err := c.service.Create(services.CreateInput{
		Subject:        entrada.Subject,
		Description:    entrada.Description,
		CategoryID:     entrada.CategoryID,
		Tags:           entrada.Tags,
		RequesterID:    entrada.RequesterID,
		RequesterEmail: entrada.RequesterEmail,
	}, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, dtos.TicketEnvelope{Ticket: dtos.NewTicketResponse(ticket)})
}

// Assignees devuelve quién puede ser responsable de cada tipo de ticket.
//
// Es para la pantalla de asignar: un módulo del frontend sólo habla con su propia API, así que la
// lista de personas la da este módulo y no `users`.
func (c *TicketController) Assignees(w http.ResponseWriter, r *http.Request) {
	porTipo, err := c.service.Assignees(auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewAssigneesResponse(porTipo))
}

// ListCategories devuelve el catálogo de categorías con cuántos tickets las usan.
//
// **Lo ve cualquiera que haya entrado** —el alta de un ticket necesita las categorías—, y lo retirado
// sólo lo ven quienes lo mantienen, que es un filtro del servicio y no de la ruta.
func (c *TicketController) ListCategories(w http.ResponseWriter, r *http.Request) {
	categorias, err := c.service.Categories(auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewCategoriesResponse(categorias))
}

// CreateCategory da de alta una categoría: Soporte y el Administrador.
func (c *TicketController) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	categoria, err := c.service.CreateCategory(entrada.Name, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, dtos.CategoryEnvelope{
		Category: dtos.NewCategoryCatalogResponse(categoria),
	})
}

// RenameCategory cambia el nombre de una categoría: Soporte y el Administrador.
func (c *TicketController) RenameCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeLaRuta(w, r, "id")
	if !ok {
		return
	}

	var entrada dtos.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	categoria, err := c.service.RenameCategory(id, entrada.Name, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.CategoryEnvelope{
		Category: dtos.NewCategoryCatalogResponse(categoria),
	})
}

// CategoryState retira o vuelve a poner una categoría: **sólo el Administrador**, y **retirar es
// desactivar** (docs/modules/tickets.md, decisiones 64 y 66).
func (c *TicketController) CategoryState(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeLaRuta(w, r, "id")
	if !ok {
		return
	}

	var entrada dtos.CategoryStateRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	categoria, err := c.service.CategoryState(id, entrada.Active, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.CategoryEnvelope{
		Category: dtos.NewCategoryCatalogResponse(categoria),
	})
}

// Tags devuelve las etiquetas del catálogo. Sin `q`, las más usadas; con `q`, las que empiezan por lo
// que se escribe; y con **`all=1`, el catálogo entero por nombre**, que es lo que necesita la pantalla
// de mantenimiento para ver también las que no lleva ningún ticket. **Incluye las que no lleva nadie**,
// con `tickets: 0` (docs/modules/tickets.md, decisión 72).
func (c *TicketController) Tags(w http.ResponseWriter, r *http.Request) {
	catalogoCompleto := r.URL.Query().Get("all") == "1"

	etiquetas, err := c.service.Tags(r.URL.Query().Get("q"), catalogoCompleto)
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewTagsResponse(etiquetas))
}

// CreateTag da de alta una etiqueta en el catálogo: Soporte y el Administrador.
//
// **No hace falta para usarla** —escribirla en un ticket la crea—, pero es lo que permite dejarla
// puesta y corregirla después (docs/modules/tickets.md, decisión 72).
func (c *TicketController) CreateTag(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.TagRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	etiqueta, err := c.service.CreateTag(entrada.Tag, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, dtos.TagEnvelope{Tag: dtos.NewTagResponse(etiqueta)})
}

// RenameTag cambia el nombre de una etiqueta **en todos los tickets que la llevan**: Soporte y el
// Administrador (docs/modules/tickets.md, decisiones 72 y 73).
func (c *TicketController) RenameTag(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.TagRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	etiqueta, err := c.service.RenameTag(
		r.PathValue("tag"),
		entrada.Tag,
		auth.MustFromContext(r.Context()),
	)
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.TagEnvelope{Tag: dtos.NewTagResponse(etiqueta)})
}

// DeleteTag retira una etiqueta **y la quita de todos sus tickets**: **sólo el Administrador**
// (docs/modules/tickets.md, decisiones 72 y 73).
func (c *TicketController) DeleteTag(w http.ResponseWriter, r *http.Request) {
	err := c.service.DeleteTag(r.PathValue("tag"), auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Show devuelve la ficha de un ticket, con su conversación, sus adjuntos y su historial.
func (c *TicketController) Show(w http.ResponseWriter, r *http.Request) {
	detalle, err := c.service.ByNumber(r.PathValue("number"), auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewTicketDetailResponse(detalle))
}

// Update cambia el asunto o la descripción.
func (c *TicketController) Update(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	actor := auth.MustFromContext(r.Context())
	numero := r.PathValue("number")

	// **Un `PATCH` de sólo etiquetas** tiene su propio camino (decisión 85): es lo único que Desarrollo
	// puede cambiar de un ticket, y desde el número de un interno va al principal, que es donde viven
	// las etiquetas. Un `PATCH` que traiga asunto, descripción o categoría sigue siendo del solicitante
	// y de Soporte, y a Desarrollo se le rechaza como siempre.
	if esPATCHDeSoloEtiquetas(entrada) {
		ticket, err := c.service.UpdateTags(numero, *entrada.Tags, actor)
		if err != nil {
			c.fail(w, err)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, dtos.TicketEnvelope{Ticket: dtos.NewTicketResponse(ticket)})
		return
	}

	// El `PATCH` cambia lo que venga: lo que no llega no se toca, y por eso se lee antes el ticket.
	actual, err := c.service.ByNumber(numero, actor)
	if err != nil {
		c.fail(w, err)
		return
	}

	asunto := actual.Ticket.Subject
	if entrada.Subject != nil {
		asunto = *entrada.Subject
	}

	descripcion := actual.Ticket.Description
	if entrada.Description != nil {
		descripcion = *entrada.Description
	}

	ticket, err := c.service.Update(numero, services.UpdateInput{
		Subject:     asunto,
		Description: descripcion,
		CategoryID:  entrada.CategoryID,
		Tags:        entrada.Tags,
	}, actor)
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.TicketEnvelope{Ticket: dtos.NewTicketResponse(ticket)})
}

// esPATCHDeSoloEtiquetas dice si el cuerpo que llega trae **sólo etiquetas**. Es la puerta que la
// decisión 85 abre a Desarrollo desde el interno: las etiquetas van al principal y no arrastran nada
// del texto. Un `PATCH` con asunto, descripción o categoría —o con etiquetas **y** algo más— sigue
// siendo del solicitante y de Soporte, y por eso va por el camino de siempre.
func esPATCHDeSoloEtiquetas(entrada dtos.UpdateTicketRequest) bool {
	return entrada.Tags != nil && entrada.Subject == nil && entrada.Description == nil &&
		entrada.CategoryID == nil
}

// Assign pone responsable.
func (c *TicketController) Assign(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.AssignRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	ticket, err := c.service.Assign(r.PathValue("number"), entrada.AssigneeID, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.TicketEnvelope{Ticket: dtos.NewTicketResponse(ticket)})
}

// State mueve el estado del ticket, con el comentario que lo acompaña.
func (c *TicketController) State(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.StateRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	ticket, err := c.service.Move(
		r.PathValue("number"),
		estadoSolicitado(entrada),
		entrada.Comment,
		auth.MustFromContext(r.Context()),
	)
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.TicketEnvelope{Ticket: dtos.NewTicketResponse(ticket)})
}

// estadoSolicitado lee el estado del cuerpo admitiendo los dos nombres: `state`, que es el que manda la
// pantalla, y `to`, que es el que fija la tabla de endpoints del documento (docs/modules/tickets.md,
// sección 5). Si vienen los dos, manda `state`, que es el que el frontend conoce.
func estadoSolicitado(entrada dtos.StateRequest) string {
	if estado := strings.TrimSpace(entrada.State); estado != "" {
		return estado
	}

	return strings.TrimSpace(entrada.To)
}

// Escalate escala el ticket a Desarrollo: crea el interno o lo reabre.
func (c *TicketController) Escalate(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.EscalateRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	ticket, err := c.service.Escalate(r.PathValue("number"), entrada.Reason, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.TicketEnvelope{Ticket: dtos.NewTicketResponse(ticket)})
}

// Reopen reabre un ticket cerrado.
func (c *TicketController) Reopen(w http.ResponseWriter, r *http.Request) {
	ticket, err := c.service.Reopen(r.PathValue("number"), auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.TicketEnvelope{Ticket: dtos.NewTicketResponse(ticket)})
}

// Comment escribe un comentario en la conversación.
func (c *TicketController) Comment(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	comentario, err := c.service.Comment(r.PathValue("number"), entrada.Body, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, dtos.CommentEnvelope{Comment: dtos.NewCommentResponse(comentario)})
}

// EditComment cambia el texto de un comentario: sólo su autor.
func (c *TicketController) EditComment(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeLaRuta(w, r, "id")
	if !ok {
		return
	}

	var entrada dtos.CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	comentario, err := c.service.EditComment(
		r.PathValue("number"),
		id,
		entrada.Body,
		auth.MustFromContext(r.Context()),
	)
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.CommentEnvelope{Comment: dtos.NewCommentResponse(comentario)})
}

// DeleteComment borra un comentario: vacía el texto y deja la marca.
func (c *TicketController) DeleteComment(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeLaRuta(w, r, "id")
	if !ok {
		return
	}

	err := c.service.DeleteComment(r.PathValue("number"), id, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// AddObserver añade **a mano** a una persona como observadora de un ticket.
//
// La cuenta llega en el cuerpo (`accountId`) y tiene que ser un técnico o un desarrollador activo,
// como para etiquetar (docs/modules/tickets.md, decisión 59), y lo puede hacer **cualquier técnico o
// desarrollador**, como quitarlos (decisión 63). Contesta `201` con **el ticket actualizado en la
// misma forma que la ficha, con su lista de observadores**; es idempotente, así que añadir a quien ya
// observa contesta igual y la lista queda como estaba (decisión 74).
func (c *TicketController) AddObserver(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.AddObserverRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	detalle, err := c.service.AddObserver(
		r.PathValue("number"),
		entrada.AccountID,
		auth.MustFromContext(r.Context()),
	)
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, dtos.ObserverEnvelope{
		Ticket: dtos.NewObserverTicketResponse(detalle.Ticket, detalle.Observers),
	})
}

// RemoveObserver quita a una persona de los observadores de un ticket.
//
// Lo puede hacer **cualquier técnico o desarrollador**, no sólo quien la etiquetó, y queda en el
// historial (docs/modules/tickets.md, decisión 63). El `{id}` de la ruta es el de la **cuenta**, no el
// de la fila de `ticket_observers`: quien se quita del ticket es una persona.
func (c *TicketController) RemoveObserver(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeLaRuta(w, r, "id")
	if !ok {
		return
	}

	err := c.service.RemoveObserver(r.PathValue("number"), id, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Upload guarda un adjunto.
//
// El archivo llega por `multipart`, y nginx admite hasta 30 MB para que quepan los 25 del tope.
func (c *TicketController) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(services.MaxAttachmentBytes); err != nil {
		httpx.WriteError(w, http.StatusRequestEntityTooLarge, "tickets.attachment.tooBig")
		return
	}

	archivo, cabecera, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.attachment.invalid")
		return
	}
	defer archivo.Close()

	// El comentario es opcional: un adjunto puede ir en la descripción inicial.
	var commentID *int64
	if valor := strings.TrimSpace(r.FormValue("commentId")); valor != "" {
		id, err := strconv.ParseInt(valor, 10, 64)
		if err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.attachment.invalid")
			return
		}
		commentID = &id
	}

	adjunto, err := c.service.Attach(
		r.PathValue("number"),
		archivo,
		cabecera,
		commentID,
		auth.MustFromContext(r.Context()),
	)
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, dtos.AttachmentEnvelope{Attachment: dtos.NewAttachmentResponse(adjunto)})
}

// Download sirve un adjunto, **con la comprobación de permisos delante**.
//
// Lo que se puede ver en línea va en línea; todo lo demás se descarga, con su nombre original y como
// flujo de bytes. Nunca por una ruta estática: el adjunto de un ticket interno no puede ser alcanzable
// por el solicitante (docs/modules/tickets.md, sección 4).
func (c *TicketController) Download(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeLaRuta(w, r, "id")
	if !ok {
		return
	}

	archivo, err := c.service.Attachment(r.PathValue("number"), id, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, err)
		return
	}
	defer archivo.Reader.Close()

	tipo, disposicion := services.ServirCon(archivo.Filename, archivo.Inline)

	w.Header().Set("Content-Type", tipo)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", disposicion+`; filename="`+nombreSeguro(archivo.Filename)+`"`)
	w.Header().Set("Content-Length", strconv.FormatInt(archivo.Size, 10))

	if _, err := io.Copy(w, archivo.Reader); err != nil {
		// La cabecera ya ha salido: sólo queda dejarlo en el log.
		logs.LogWarning("se ha cortado la descarga del adjunto " + strconv.FormatInt(id, 10))
	}
}

// idDeLaRuta lee un identificador de la dirección. Si no es un número, no existe.
func idDeLaRuta(w http.ResponseWriter, r *http.Request, nombre string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(nombre), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "tickets.notFound")
		return 0, false
	}

	return id, true
}

// nombreSeguro quita de un nombre lo que rompería la cabecera: comillas, saltos de línea y rutas.
func nombreSeguro(nombre string) string {
	limpio := strings.ReplaceAll(nombre, `"`, "")
	limpio = strings.ReplaceAll(limpio, "\n", "")
	limpio = strings.ReplaceAll(limpio, "\r", "")

	return strings.ReplaceAll(limpio, "\\", "")
}

// fail traduce el error del servicio a la clave y el código que le tocan.
func (c *TicketController) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrTicketNotFound), errors.Is(err, repositories.ErrTicketNotFound):
		httpx.WriteError(w, http.StatusNotFound, "tickets.notFound")
	case errors.Is(err, services.ErrCommentNotFound):
		httpx.WriteError(w, http.StatusNotFound, "tickets.comment.notFound")
	case errors.Is(err, services.ErrAttachmentMissing):
		httpx.WriteError(w, http.StatusNotFound, "tickets.attachment.notFound")
	case errors.Is(err, services.ErrObserverNotFound), errors.Is(err, repositories.ErrObserverNotFound):
		httpx.WriteError(w, http.StatusNotFound, "tickets.observer.notFound")
	case errors.Is(err, services.ErrCategoryNotFound), errors.Is(err, repositories.ErrCategoryNotFound):
		httpx.WriteError(w, http.StatusNotFound, "tickets.category.notFound")
	case errors.Is(err, services.ErrTagDesconocida):
		// **Una etiqueta que no está en el catálogo** y quien la pone no es el Administrador (decisión
		// 84): es un dato que no vale, así que 422, como el resto de lo que el catálogo rechaza.
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())

	case errors.Is(err, services.ErrTagNotFound), errors.Is(err, repositories.ErrTagNotFound):
		// La etiqueta de la ruta no está en el catálogo: es un «no existe», no una errata de quien
		// escribe.
		httpx.WriteError(w, http.StatusNotFound, "tickets.tag.notFound")
	case errors.Is(err, services.ErrSubjectRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.subject.required")
	case errors.Is(err, services.ErrBodyRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.description.required")
	case errors.Is(err, services.ErrBodyNotAllowed):
		// El texto trae HTML que no está en la lista blanca: se dice cuál es el problema y no se guarda
		// nada, en vez de limpiarlo por su cuenta (docs/modules/tickets.md, sección 2.3).
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.body.notAllowed")
	case errors.Is(err, services.ErrReasonRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.reason.required")
	case errors.Is(err, services.ErrMentionNotAllowed):
		// Etiquetar es de Soporte y Desarrollo, y sólo a técnicos y desarrolladores activos: una
		// mención que no vale rechaza el cuerpo entero, como el saneador (decisión 59).
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.mention.notAllowed")
	case errors.Is(err, services.ErrCommentRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.comment.required")
	case errors.Is(err, services.ErrStateUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.state.unknown")
	case errors.Is(err, services.ErrCierreSinComentario):
		// **Un ticket no se cierra sin decir por qué** (decisión 81): lo exige el servidor, no la
		// pantalla, y se contesta con la clave que el frontend ya traduce.
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.cierre.sinComentario")
	case errors.Is(err, services.ErrTransitionNotAllowed):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.transition.notAllowed")
	case errors.Is(err, services.ErrAssigneeUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.assignee.unknown")
	case errors.Is(err, services.ErrRequesterUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.requester.unknown")
	case errors.Is(err, services.ErrAttachmentFormat):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.attachment.extension")
	case errors.Is(err, services.ErrAttachmentTooBig):
		httpx.WriteError(w, http.StatusRequestEntityTooLarge, "tickets.attachment.tooBig")
	case errors.Is(err, services.ErrAttachmentInvalid):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.attachment.invalid")
	case errors.Is(err, services.ErrCategoryRequired):
		// Un ticket sin categoría no se acepta: la categoría es obligatoria.
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.category.required")
	case errors.Is(err, services.ErrCategoryInactive):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.category.inactive")
	case errors.Is(err, services.ErrCategoryNameRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.category.name.required")
	case errors.Is(err, services.ErrTagRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.tag.required")
	case errors.Is(err, services.ErrTagTooLong):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "tickets.tag.tooLong")
	case errors.Is(err, services.ErrCategoryDuplicate):
		// Renombrar (o crear) a un nombre que ya existe es un choque, no una errata.
		httpx.WriteError(w, http.StatusConflict, "tickets.category.duplicate")
	case errors.Is(err, services.ErrTagDuplicate), errors.Is(err, repositories.ErrTagDuplicate):
		// Lo mismo en las etiquetas: el choque se compara por el nombre normalizado, así que `Red-Wifi`
		// choca con `red-wifi`.
		httpx.WriteError(w, http.StatusConflict, "tickets.tag.duplicate")
	case errors.Is(err, services.ErrCategoryLastActive):
		// El catálogo no puede quedarse sin ninguna categoría activa: el alta no tendría qué ofrecer.
		httpx.WriteError(w, http.StatusConflict, "tickets.category.lastActive")
	case errors.Is(err, services.ErrCategoryForbidden):
		httpx.WriteError(w, http.StatusForbidden, "tickets.category.forbidden")
	case errors.Is(err, services.ErrTagForbidden):
		// Crear y renombrar una etiqueta es de Soporte y del Administrador, y **retirarla es sólo del
		// Administrador**, porque toca todos los tickets que la llevan (decisión 73).
		httpx.WriteError(w, http.StatusForbidden, "tickets.tag.forbidden")
	case errors.Is(err, services.ErrClosed):
		// El ticket está cerrado: para escribir, primero se reabre.
		httpx.WriteError(w, http.StatusConflict, "tickets.closed")
	case errors.Is(err, services.ErrCommentNotYours), errors.Is(err, services.ErrRequesterNotYours),
		errors.Is(err, services.ErrForbidden):
		// Son permisos y no erratas: quien los recibe ha pedido algo que no le corresponde.
		clave := "tickets.forbidden"
		if errors.Is(err, services.ErrCommentNotYours) {
			clave = "tickets.comment.notYours"
		}
		if errors.Is(err, services.ErrRequesterNotYours) {
			clave = "tickets.requester.notAllowed"
		}

		httpx.WriteError(w, http.StatusForbidden, clave)
	default:
		logs.LogError("error inesperado en tickets: " + err.Error())
		httpx.WriteError(w, http.StatusInternalServerError, KeyInternal)
	}
}
