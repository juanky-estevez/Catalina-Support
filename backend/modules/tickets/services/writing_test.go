package services

import (
	"errors"
	"testing"

	"catalina-support/backend/shared/auth"
)

func TestImproveDraftRechazaPapelesSinPermisoAntesDeLeerElTicket(t *testing.T) {
	servicio := &Service{}
	for _, actor := range []auth.Identity{
		{Role: auth.RoleUsuario},
		{Role: auth.RoleAdministrador},
	} {
		_, err := servicio.ImproveDraft("CS-2026-0001", "comment", "borrador", "professional", actor)
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("el papel %s debía recibir forbidden y obtuvo %v", actor.Role, err)
		}
	}
}

func TestImproveDraftValidaBorradorEditorYTonoAntesDelMotor(t *testing.T) {
	servicio := &Service{}
	actor := auth.Identity{Role: auth.RoleSoporte}
	casos := []struct {
		editor, borrador, tono string
		esperado               error
	}{
		{"comment", "   ", "professional", ErrWritingDraftRequired},
		{"subject", "texto", "professional", ErrWritingEditorInvalid},
		{"comment", "texto", "aggressive", ErrWritingToneInvalid},
	}
	for _, caso := range casos {
		_, err := servicio.ImproveDraft("CS-2026-0001", caso.editor, caso.borrador, caso.tono, actor)
		if !errors.Is(err, caso.esperado) {
			t.Fatalf("editor=%q borrador=%q tono=%q: se esperaba %v y llegó %v",
				caso.editor, caso.borrador, caso.tono, caso.esperado, err)
		}
	}
}
