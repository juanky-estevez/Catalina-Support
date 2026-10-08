package services

import (
	"strings"

	"catalina-support/backend/shared/auth"
)

var writingTones = map[string]bool{
	"professional": true,
	"friendly":     true,
	"brief":        true,
	"empathetic":   true,
	"technical":    true,
}

// ImproveDraft autoriza el editor, obtiene el destinatario desde la base y delega únicamente el
// contexto mínimo. No persiste ni el borrador ni el resultado.
func (s *Service) ImproveDraft(number, editor, draft, tone string, actor auth.Identity) (string, error) {
	if !actor.Can(auth.RoleSoporte, auth.RoleDesarrollo) {
		return "", ErrForbidden
	}
	borrador := strings.TrimSpace(draft)
	if borrador == "" {
		return "", ErrWritingDraftRequired
	}
	if !writingTones[tone] {
		return "", ErrWritingToneInvalid
	}
	if editor != "description" && editor != "comment" {
		return "", ErrWritingEditorInvalid
	}
	if s.writing == nil || s.accounts == nil || s.language == nil {
		return "", ErrWritingUnavailable
	}

	principal, interno, _, err := s.porNumero(number)
	if err != nil {
		return "", err
	}
	detalle, err := s.ticketDe(number, actor)
	if err != nil {
		return "", err
	}
	if editor == "description" {
		if interno || !puedeEditar(actor, detalle) {
			return "", ErrForbidden
		}
	} else if !puedeComentar(actor, detalle) {
		return "", ErrForbidden
	}

	destinatarioID := principal.RequesterID
	tipo := "main"
	if interno {
		internoDB, err := s.tickets.InternalByNumber(strings.TrimSpace(number))
		if err != nil {
			return "", traducirTicket(err)
		}
		destinatarioID = internoDB.CreatedByID
		tipo = "internal"
	}
	destinatario, err := s.accounts.ByID(destinatarioID)
	if err != nil {
		return "", err
	}
	idioma, err := s.language.Language()
	if err != nil {
		return "", err
	}

	return s.writing.ImproveDraft(borrador, tone, destinatario.FullName(), tipo, idioma)
}
