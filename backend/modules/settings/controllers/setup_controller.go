package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"catalina-support/backend/modules/settings/dtos"
	"catalina-support/backend/modules/settings/services"
	"catalina-support/backend/shared/httpx"
)

// SetupController atiende **la vista de primer arranque**: lo que se pide para dejar la instalación en
// marcha, paso a paso, antes de que exista ninguna cuenta con la que entrar.
//
// **No lleva sesión, y no puede llevarla**: aquí se configura la puerta, así que todavía no hay puerta
// por la que entrar. Lo que protege estas rutas no es un permiso, es **el sello**: en cuanto la
// instalación está sellada, todas contestan **409** y no hay forma de reescribirla desde aquí
// (docs/primer-arranque.md, secciones 2 y 6).
type SetupController struct {
	service *services.Service
}

// NewSetupController construye el controlador del asistente.
func NewSetupController(service *services.Service) *SetupController {
	return &SetupController{service: service}
}

// Show dice si la instalación está sin instalar y qué hay puesto, para poder seguir donde se dejó.
func (c *SetupController) Show(w http.ResponseWriter, r *http.Request) {
	estado, err := c.service.EstadoDeInstalacion()
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewSetupResponse(estado))
}

// SaveInstallation guarda el primer paso: el nombre y el idioma.
func (c *SetupController) SaveInstallation(w http.ResponseWriter, r *http.Request) {
	c.save(w, r, 1)
}

// SaveEntry guarda el segundo: el método de entrada y sus datos.
func (c *SetupController) SaveEntry(w http.ResponseWriter, r *http.Request) {
	c.save(w, r, 2)
}

// SaveLocation guarda el tercero: la región horaria y la dirección pública.
func (c *SetupController) SaveLocation(w http.ResponseWriter, r *http.Request) {
	c.save(w, r, 3)
}

// SaveMail guarda el cuarto: el correo saliente.
func (c *SetupController) SaveMail(w http.ResponseWriter, r *http.Request) {
	c.save(w, r, 4)
}

// Finish sella la instalación: **es lo que hace que el asistente no vuelva a aparecer**.
func (c *SetupController) Finish(w http.ResponseWriter, r *http.Request) {
	estado, err := c.service.TerminarInstalacion()
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewSetupResponse(estado))
}

// save lee un paso y lo guarda. Los cuatro pasos comparten el mismo cuerpo porque comparten la misma
// forma: el paso que llega trae **lo suyo** y lo demás viene vacío a propósito.
func (c *SetupController) save(w http.ResponseWriter, r *http.Request, paso int) {
	var entrada dtos.SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	estado, err := c.service.GuardarPasoDeInstalacion(paso, dtos.NewSetupStep(entrada))
	if err != nil {
		c.fail(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewSetupResponse(estado))
}

// fail traduce el error del servicio a la clave y el código que le tocan
// (docs/primer-arranque.md, sección 6).
func (c *SetupController) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrYaInstalada):
		// **La instalación ya está sellada**: no se configura dos veces.
		httpx.WriteError(w, http.StatusConflict, "setup.alreadyInstalled")
	case errors.Is(err, services.ErrPasoIncompleto):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "setup.step.incomplete")
	case errors.Is(err, services.ErrCorreoIncompleto):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "setup.mail.incomplete")
	case errors.Is(err, services.ErrNameTooLong):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.name.tooLong")
	case errors.Is(err, services.ErrLanguageUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.language.unknown")
	case errors.Is(err, services.ErrMethodUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.method.unknown")
	case errors.Is(err, services.ErrMethodNotConfigured):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.method.notConfigured")
	case errors.Is(err, services.ErrDirectoryIncomplete):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.directory.incomplete")
	case errors.Is(err, services.ErrKeycloakIncomplete):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.keycloak.incomplete")
	case errors.Is(err, services.ErrTimeZoneUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.timeZone.unknown")
	case errors.Is(err, services.ErrPublicURLInvalid):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.publicUrl.invalid")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, KeyInternal)
	}
}
