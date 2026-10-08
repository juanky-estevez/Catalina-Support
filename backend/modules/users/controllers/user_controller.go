// Package controllers es la capa HTTP del módulo users. Los permisos los pone el middleware sobre la
// ruta, y las reglas del alta viven en el servicio (docs/modules/users.md, sección 5).
package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/users/dtos"
	"catalina-support/backend/modules/users/services"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/httpx"
)

// KeyInternal es lo que se responde cuando el fallo es nuestro.
const KeyInternal = "error interno"

// UserController atiende los endpoints de cuentas.
type UserController struct {
	service *services.Service
}

// NewUserController construye el controlador.
func NewUserController(service *services.Service) *UserController {
	return &UserController{service: service}
}

// Create da de alta una cuenta.
//
// Si es local, el enlace de alta sale en segundo plano: **la cuenta ya está creada** aunque el correo
// falle, y se puede lanzar otro desde `POST /api/users/{id}/reset-password`
// (docs/modules/users.md, sección 5).
func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	actor := auth.MustFromContext(r.Context())

	account, err := c.service.Create(services.CreateInput{
		Name:       entrada.Name,
		LastName:   entrada.LastName,
		Email:      entrada.Email,
		Role:       entrada.Role,
		Origin:     entrada.Origin,
		ExternalID: entrada.ExternalID,
	}, actor)

	invitado := err == nil && account.Origin == auth.OriginLocal

	if err != nil {
		// Si el fallo fue sólo al mandar el enlace, la cuenta existe: se dice, en vez de dejar a quien
		// da de alta creyendo que no se creó nada y volviendo a intentarlo.
		if account.ID != 0 && account.Origin == auth.OriginLocal {
			logs.LogWarning("la cuenta se creó pero el enlace de alta no salió: " + err.Error())
			httpx.WriteJSON(w, http.StatusCreated, dtos.UserEnvelope{User: dtos.NewUserResponse(account)})
			return
		}

		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, dtos.UserEnvelope{
		User:    dtos.NewUserResponse(account),
		Invited: invitado,
	})
}

// ResetPassword lanza el enlace otra vez: es la salida para quien no recibió el primero, y no hace
// falta borrar y volver a crear nada.
func (c *UserController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "users.notFound")
		return
	}

	actor := auth.MustFromContext(r.Context())

	account, err := c.service.SendInviteAgain(id, actor)
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.UserEnvelope{
		User:    dtos.NewUserResponse(account),
		Invited: true,
	})
}

// List devuelve la lista de cuentas, con sus filtros y su paginación.
func (c *UserController) List(w http.ResponseWriter, r *http.Request) {
	filtros := services.Filtros{
		Role:   r.URL.Query().Get("role"),
		Origin: r.URL.Query().Get("origin"),
		Query:  r.URL.Query().Get("q"),
	}

	// `active` es un filtro de tres estados: sin él, todas; con él, activas o desactivadas.
	if activo := r.URL.Query().Get("active"); activo != "" {
		valor := activo == "true"
		filtros.IsActive = &valor
	}

	pagina, _ := strconv.Atoi(r.URL.Query().Get("page"))
	porPagina, _ := strconv.Atoi(r.URL.Query().Get("perPage"))

	resultado, err := c.service.List(filtros, pagina, porPagina)
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewUsersResponse(resultado))
}

// Show devuelve la ficha de una cuenta. Es de administrador: Soporte necesita la lista para buscar a
// quien reporta, no la ficha de nadie (docs/modules/users.md, sección 4).
func (c *UserController) Show(w http.ResponseWriter, r *http.Request) {
	id, ok := c.idDeLaRuta(w, r)
	if !ok {
		return
	}

	cuenta, err := c.service.ByID(id)
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.UserEnvelope{User: dtos.NewUserResponse(cuenta)})
}

// Update cambia una cuenta. Soporte sólo puede el nombre y los apellidos, y lo comprueba el servicio,
// que es quien ve qué campos vienen.
func (c *UserController) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := c.idDeLaRuta(w, r)
	if !ok {
		return
	}

	var entrada dtos.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	cuenta, err := c.service.Patch(id, services.PatchInput{
		Name:     entrada.Name,
		LastName: entrada.LastName,
		Email:    entrada.Email,
		Role:     entrada.Role,
		IsActive: entrada.IsActive,
	}, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.UserEnvelope{User: dtos.NewUserResponse(cuenta)})
}

// UpdateProfile cambia el perfil propio: el nombre, los apellidos y el idioma.
func (c *UserController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	cuenta, err := c.service.PatchOwn(services.PatchInput{
		Name:     entrada.Name,
		LastName: entrada.LastName,
	}, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.UserEnvelope{User: dtos.NewUserResponse(cuenta)})
}

// Deactivate desactiva una cuenta. Devuelve los avisos de lo que sí se ha hecho pero conviene saber.
func (c *UserController) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, ok := c.idDeLaRuta(w, r)
	if !ok {
		return
	}

	cuenta, avisos, err := c.service.Deactivate(id, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.UserEnvelope{
		User:     dtos.NewUserResponse(cuenta),
		Warnings: avisos,
	})
}

// Activate reactiva una cuenta.
func (c *UserController) Activate(w http.ResponseWriter, r *http.Request) {
	id, ok := c.idDeLaRuta(w, r)
	if !ok {
		return
	}

	cuenta, err := c.service.Activate(id, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.UserEnvelope{User: dtos.NewUserResponse(cuenta)})
}

// ChangeOrigin cambia el origen de una cuenta. Es una de las dos acciones peligrosas del módulo.
func (c *UserController) ChangeOrigin(w http.ResponseWriter, r *http.Request) {
	id, ok := c.idDeLaRuta(w, r)
	if !ok {
		return
	}

	var entrada dtos.OriginRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	cuenta, err := c.service.ChangeOrigin(
		id,
		strings.TrimSpace(entrada.Origin),
		strings.TrimSpace(entrada.ExternalID),
		auth.MustFromContext(r.Context()),
	)
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.UserEnvelope{User: dtos.NewUserResponse(cuenta)})
}

// idDeLaRuta lee el identificador de la dirección. Si no es un número, la cuenta no existe.
func (c *UserController) idDeLaRuta(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "users.notFound")
		return 0, false
	}

	return id, true
}

// fail traduce el error del servicio a la clave y el código que le tocan
// (docs/modules/users.md, sección 9).
func (c *UserController) fail(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, services.ErrNameRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "users.name.required")
	case errors.Is(err, services.ErrEmailInvalid):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "users.email.invalid")
	case errors.Is(err, services.ErrEmailDuplicated):
		httpx.WriteError(w, http.StatusConflict, "users.email.duplicated")
	case errors.Is(err, services.ErrAccountInactive):
		httpx.WriteError(w, http.StatusConflict, "users.accountInactive")
	case errors.Is(err, services.ErrPasswordNotLocal):
		// **La contraseña no es nuestra** (decisión del responsable, 2026-09-29): es un dato que no
		// aplica a esa cuenta, así que 422 con su clave.
		httpx.WriteError(w, http.StatusUnprocessableEntity, "users.password.notLocal")
	case errors.Is(err, services.ErrOriginUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "users.origin.unknown")
	case errors.Is(err, services.ErrDirectoryNotFound):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "users.directory.notFound")
	case errors.Is(err, services.ErrOriginByDirectory):
		// Las cuentas de directorio no se crean ni se cambian a mano: las pone el directorio al entrar
		// esa persona (docs/modules/users.md, sección 5, punto 4).
		httpx.WriteError(w, http.StatusUnprocessableEntity, "users.origin.byDirectory")
	case errors.Is(err, services.ErrDirectoryActivatesItself):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "users.directory.activatesItself")
	case errors.Is(err, services.ErrDirectoryUnavailable):
		// No se pudo comprobar con el directorio: es un fallo del servidor y no de quien lo pide.
		httpx.WriteError(w, http.StatusServiceUnavailable, "users.directory.unavailable")
	case errors.Is(err, services.ErrSelfDeactivation):
		// Desactivarse es quedarse fuera al instante y sin poder volver.
		httpx.WriteError(w, http.StatusForbidden, "users.selfDeactivation")
	case errors.Is(err, services.ErrStateHasItsOwnAction):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "users.stateHasItsOwnAction")
	case errors.Is(err, services.ErrOriginInUse):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "users.origin.inUse")
	case errors.Is(err, services.ErrRoleNotAllowed):
		// Un permiso y no una errata: quien lo recibe ha pedido algo que no le corresponde.
		httpx.WriteError(w, http.StatusForbidden, "users.role.notAllowed")
	case errors.Is(err, auth.ErrAccountNotFound):
		httpx.WriteError(w, http.StatusNotFound, "users.notFound")
	default:
		logs.LogError("error inesperado en users: " + err.Error())
		httpx.WriteError(w, http.StatusInternalServerError, KeyInternal)
	}
}
