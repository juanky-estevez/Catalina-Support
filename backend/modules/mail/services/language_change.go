package services

import (
	"fmt"
	"regexp"
	"strings"

	"catalina-support/backend/modules/mail/repositories"
)

// Translator is declared by mail because mail owns the text that is translated.
type Translator interface {
	Traducir(text, language string) (string, error)
	EsProveedorComercial() bool
}

type LanguageDraft struct {
	Key                  string `json:"key"`
	SourceLanguage       string `json:"sourceLanguage"`
	TargetLanguage       string `json:"targetLanguage"`
	SourceSubject        string `json:"sourceSubject"`
	SourceBody           string `json:"sourceBody"`
	DraftSubject         string `json:"draftSubject"`
	DraftBody            string `json:"draftBody"`
	ExistingSubject      string `json:"existingSubject"`
	ExistingBody         string `json:"existingBody"`
	DestinationCustomized bool `json:"destinationCustomized"`
}

type LanguageChoice struct { Key, Subject, Body string }

func (s *Service) ApplyLanguage(target string, choices []LanguageChoice, actorID *int64) error {
	if !LanguageIsValid(target) || len(choices) != len(Keys) { return fmt.Errorf("mail.translation.incomplete") }
	seen := map[string]bool{}
	rows := make([]repositories.LanguageTemplate, 0, len(choices))
	for _, choice := range choices {
		if seen[choice.Key] { return fmt.Errorf("mail.translation.incomplete") }
		if err := validateTarget(choice.Key, target); err != nil { return err }
		if strings.TrimSpace(choice.Subject) == "" || strings.TrimSpace(choice.Body) == "" { return fmt.Errorf("mail.translation.incomplete") }
		if err := Validate(choice.Key, choice.Subject, choice.Body); err != nil { return err }
		seen[choice.Key] = true
		rows = append(rows, repositories.LanguageTemplate{Key:choice.Key, Subject:choice.Subject, Body:choice.Body})
	}
	return s.repo.ApplyLanguage(target, rows, actorID)
}

var protectedPart = regexp.MustCompile(`(?s)</?[^>]+>|\{\{[^{}]+\}\}`)

// LanguageDrafts creates reviewable drafts and never persists them.
func (s *Service) LanguageDrafts(target string, translator Translator, commercialConfirmed bool) ([]LanguageDraft, error) {
	if !LanguageIsValid(target) || translator == nil {
		return nil, fmt.Errorf("mail.language.unknown")
	}
	if translator.EsProveedorComercial() && !commercialConfirmed { return nil, fmt.Errorf("mail.translation.costConfirmationRequired") }
	source := "es"
	if target == "es" { source = "en" }
	all, err := s.repo.FindAll()
	if err != nil { return nil, err }
	by := map[string]repositories.Template{}
	for _, template := range all { by[template.Key+":"+template.Language] = template }
	drafts := make([]LanguageDraft, 0, len(Keys))
	for _, key := range Keys {
		from, okFrom := by[key+":"+source]
		to, okTo := by[key+":"+target]
		if !okFrom || !okTo { return nil, repositories.ErrTemplateNotFound }
		subject, err := translateProtected(from.Subject, target, translator)
		if err != nil { return nil, err }
		body, err := translateProtected(from.Body, target, translator)
		if err != nil { return nil, err }
		if err := Validate(key, subject, body); err != nil { return nil, err }
		drafts = append(drafts, LanguageDraft{Key:key, SourceLanguage:source, TargetLanguage:target,
			SourceSubject:from.Subject, SourceBody:from.Body, DraftSubject:subject, DraftBody:body,
			ExistingSubject:to.Subject, ExistingBody:to.Body, DestinationCustomized:to.Edited()})
	}
	return drafts, nil
}

func (s *Service) ManualLanguageDrafts(target string) ([]LanguageDraft, error) {
	if !LanguageIsValid(target) { return nil, fmt.Errorf("mail.language.unknown") }
	source := "es"; if target == "es" { source = "en" }
	all, err := s.repo.FindAll(); if err != nil { return nil, err }
	by := map[string]repositories.Template{}; for _, t := range all { by[t.Key+":"+t.Language] = t }
	out := make([]LanguageDraft, 0, len(Keys))
	for _, key := range Keys {
		from, okFrom := by[key+":"+source]; to, okTo := by[key+":"+target]
		if !okFrom || !okTo { return nil, repositories.ErrTemplateNotFound }
		out = append(out, LanguageDraft{Key:key, SourceLanguage:source, TargetLanguage:target,
			SourceSubject:from.Subject, SourceBody:from.Body, DraftSubject:to.Subject, DraftBody:to.Body,
			ExistingSubject:to.Subject, ExistingBody:to.Body, DestinationCustomized:to.Edited()})
	}
	return out, nil
}

func translateProtected(value, language string, translator Translator) (string, error) {
	parts := protectedPart.FindAllString(value, -1)
	protected := value
	for i, part := range parts { protected = strings.Replace(protected, part, fmt.Sprintf("[[CS_%03d]]", i), 1) }
	out, err := translator.Traducir(protected, language)
	if err != nil { return "", err }
	for i, part := range parts {
		token := fmt.Sprintf("[[CS_%03d]]", i)
		if strings.Count(out, token) != 1 { return "", fmt.Errorf("mail.translation.protectedChanged") }
		out = strings.Replace(out, token, part, 1)
	}
	if protectedPart.MatchString(out) && len(protectedPart.FindAllString(out, -1)) != len(parts) {
		return "", fmt.Errorf("mail.translation.protectedChanged")
	}
	return out, nil
}
