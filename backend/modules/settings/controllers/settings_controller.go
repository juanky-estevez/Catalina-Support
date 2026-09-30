// Package controllers es la capa HTTP del módulo settings. Los permisos los pone el middleware sobre
// la ruta (docs/modules/settings.md, sección 6).
package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/settings/dtos"
	"catalina-support/backend/modules/settings/services"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/httpx"
)

// KeyInternal es lo que se responde cuando el fallo es nuestro.
const KeyInternal = "error interno"

// SettingsController atiende la configuración y la marca.
type SettingsController struct {
	service *services.Service
}

// NewSettingsController construye el controlador.
func NewSettingsController(service *services.Service) *SettingsController {
	return &SettingsController{service: service}
}

// Show devuelve la configuración entera, para la pantalla de administración.
func (c *SettingsController) Show(w http.ResponseWriter, r *http.Request) {
	config, err := c.service.Config()
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewSettingsResponse(config))
}

// Update guarda la configuración.
func (c *SettingsController) Update(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	config, err := c.service.Update(services.UpdateInput{
		Name:                 entrada.Name,
		EntryMethod:          entrada.EntryMethod,
		Directory:            dtos.NewDirectoryInput(entrada.Directory),
		Keycloak:             dtos.NewKeycloakInput(entrada.Keycloak),
		Language:             entrada.Language,
		PrimaryColor:         entrada.PrimaryColor,
		NumberPrefix:         entrada.NumberPrefix,
		MainAssignment:       entrada.MainAssignment,
		MainNotification:     entrada.MainNotification,
		InternalAssignment:   entrada.InternalAssignment,
		InternalNotification: entrada.InternalNotification,
		TimeZone:             entrada.TimeZone,
		PublicAppURL:         entrada.PublicAppURL,
	}, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewSettingsResponse(config))
}

// Brand es lo que la aplicación necesita **antes de que nadie haya entrado**: el color institucional
// y si hay logo propio. Es público a propósito: la pantalla de entrada se pinta antes de entrar.
func (c *SettingsController) Brand(w http.ResponseWriter, r *http.Request) {
	publico, err := c.service.PublicBrand()
	if err != nil {
		c.fail(w, r, err)
		return
	}

	// Con un color, es una **vista previa**: se devuelven los colores resueltos para ese color y no se
	// guarda nada. Es lo que permite ver cómo queda antes de cambiar el color de la casa.
	if color := r.URL.Query().Get("color"); color != "" {
		publico.Colors = c.service.ResolveColors(color)
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewBrandResponse(publico))
}

// UploadLogo recibe el archivo del logo y lo guarda en el hueco que se le pide.
func (c *SettingsController) UploadLogo(w http.ResponseWriter, r *http.Request) {
	variant := r.URL.Query().Get("variant")

	// Se corta el cuerpo antes de leerlo entero: un archivo que pasa del límite se rechaza sin
	// gastar memoria en él.
	r.Body = http.MaxBytesReader(w, r.Body, services.MaxLogoBytes+1024)

	archivo, _, err := r.FormFile("logo")
	if err != nil {
		// Si el cuerpo se pasó del límite, `FormFile` falla con ese motivo: se distingue para poder
		// decir que el archivo es grande y no que está mal formado.
		var demasiado *http.MaxBytesError
		if errors.As(err, &demasiado) {
			httpx.WriteError(w, http.StatusRequestEntityTooLarge, "settings.logo.tooBig")
			return
		}
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.logo.invalid")
		return
	}
	defer archivo.Close()

	info, err := c.service.SaveLogo(variant, archivo, auth.MustFromContext(r.Context()))
	if err != nil {
		c.fail(w, r, err)
		return
	}

	config, err := c.service.Config()
	if err != nil {
		c.fail(w, r, err)
		return
	}

	logs.LogSuccess("logo de la instalación subido (" + variant + ", " + info.FileName + ")")
	httpx.WriteJSON(w, http.StatusOK, dtos.NewSettingsResponse(config))
}

// DeleteLogo quita el logo propio de un hueco y lo devuelve al de fábrica.
func (c *SettingsController) DeleteLogo(w http.ResponseWriter, r *http.Request) {
	variant := r.URL.Query().Get("variant")

	if err := c.service.RemoveLogo(variant, auth.MustFromContext(r.Context())); err != nil {
		c.fail(w, r, err)
		return
	}

	config, err := c.service.Config()
	if err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dtos.NewSettingsResponse(config))
}

// LogoFile sirve el archivo del logo que toca a ese tema.
//
// **Es público**: la pantalla de entrada lo enseña antes de que nadie haya entrado. Sólo devuelve la
// marca de la institución, y nada más.
func (c *SettingsController) LogoFile(w http.ResponseWriter, r *http.Request) {
	logo, err := c.service.LogoFor(r.URL.Query().Get("theme"))
	if err != nil {
		if errors.Is(err, services.ErrLogoNotFound) {
			// No hay logo propio: es lo normal en una instalación nueva. El frontend enseña el de
			// fábrica, que viene con la aplicación.
			httpx.WriteError(w, http.StatusNotFound, "settings.logo.notFound")
			return
		}
		c.fail(w, r, err)
		return
	}

	// El SVG se sirve **aislado**: si alguien abre su dirección directamente, no puede ejecutar nada.
	// Dentro de la aplicación se usa como imagen, donde un SVG no puede hacer nada de todos modos.
	if logo.ContentType == "image/svg+xml" {
		w.Header().Set("Content-Security-Policy", "sandbox")
	}

	w.Header().Set("Content-Type", logo.ContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// El sello de la versión: si la dirección lo trae y coincide, el navegador puede guardarlo para
	// siempre; si no lo trae, se le dice que pregunte, para no enseñar un logo viejo.
	if r.URL.Query().Get("v") == logo.Version {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}

	archivo, err := os.Open(logo.Path)
	if err != nil {
		// Estaba en la base pero ya no en el disco: para quien pregunta, no hay logo.
		logs.LogWarning("el logo de la instalación no está en el disco: " + err.Error())
		httpx.WriteError(w, http.StatusNotFound, "settings.logo.notFound")
		return
	}
	defer archivo.Close()

	_, _ = io.Copy(w, archivo)
}

// TestDirectory prueba una configuración de directorio **sin guardarla**.
//
// Es lo que hace el botón «Probar la conexión»: conecta y valida la cuenta de servicio. **No valida la
// contraseña de nadie** —eso sólo se sabe cuando alguien entra— y no toca la base.
func (c *SettingsController) TestDirectory(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.DirectoryDto
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	if err := c.service.TestDirectory(dtos.NewDirectoryInput(entrada)); err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// TestKeycloak prueba una configuración de Keycloak sin guardarla: lee su documento de reino.
func (c *SettingsController) TestKeycloak(w http.ResponseWriter, r *http.Request) {
	var entrada dtos.KeycloakDto
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, KeyInternal)
		return
	}

	if err := c.service.TestKeycloak(dtos.NewKeycloakInput(entrada)); err != nil {
		c.fail(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// fail traduce el error del servicio a la clave y el código que le tocan
// (docs/modules/settings.md, sección 7).
func (c *SettingsController) fail(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, services.ErrNameTooLong):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.name.tooLong")
	case errors.Is(err, services.ErrMethodUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.method.unknown")
	case errors.Is(err, services.ErrMethodNotConfigured):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.method.notConfigured")
	case errors.Is(err, services.ErrDirectoryIncomplete):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.directory.incomplete")
	case errors.Is(err, services.ErrKeycloakIncomplete):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.keycloak.incomplete")
	case errors.Is(err, services.ErrDirectoryUnreachable):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.directory.unreachable")
	case errors.Is(err, services.ErrKeycloakUnreachable):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.keycloak.unreachable")
	case errors.Is(err, services.ErrLanguageUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.language.unknown")
	case errors.Is(err, services.ErrPrimaryColorInvalid):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.primaryColor.invalid")
	case errors.Is(err, services.ErrTimeZoneUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.timeZone.unknown")
	case errors.Is(err, services.ErrPublicURLInvalid):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.publicUrl.invalid")
	case errors.Is(err, services.ErrNumberPrefixInvalid):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.numberPrefix.invalid")
	case errors.Is(err, services.ErrAssignmentUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.assignment.unknown")
	case errors.Is(err, services.ErrNotificationUnknown):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.notification.unknown")
	case errors.Is(err, services.ErrNotificationWithoutAssigment):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.notification.withoutAssignment")
	case errors.Is(err, services.ErrLogoFormat):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.logo.format")
	case errors.Is(err, services.ErrLogoInvalid):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "settings.logo.invalid")
	case errors.Is(err, services.ErrLogoTooBig):
		httpx.WriteError(w, http.StatusRequestEntityTooLarge, "settings.logo.tooBig")
	case errors.Is(err, services.ErrLogoVariantUnknown):
		httpx.WriteError(w, http.StatusNotFound, "settings.logo.variantUnknown")
	case errors.Is(err, services.ErrLogoNotFound):
		httpx.WriteError(w, http.StatusNotFound, "settings.logo.notFound")
	default:
		logs.LogError("error inesperado en settings: " + err.Error())
		httpx.WriteError(w, http.StatusInternalServerError, KeyInternal)
	}
}
