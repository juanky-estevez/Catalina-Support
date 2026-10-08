package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"catalina-support/backend/modules/settings/repositories"
	"catalina-support/backend/shared/auth"
)

var (
	ErrAIModeUnknown       = errors.New("settings.ai.modeUnknown")
	ErrAIProviderUnknown   = errors.New("settings.ai.providerUnknown")
	ErrAIAuthUnknown       = errors.New("settings.ai.authUnknown")
	ErrAIModelRequired     = errors.New("settings.ai.modelRequired")
	ErrAIPrivacyRequired   = errors.New("settings.ai.privacyRequired")
	ErrAICredentialInvalid = errors.New("settings.ai.credentialInvalid")
)

type AIConfigurationInput struct {
	Mode, Provider, BaseURL, Model, AuthType, AuthHeader, Credential, Language string
	PrivacyConfirmed                                                           bool
}

type AIConfigurationView struct {
	Mode, Provider, BaseURL, Model, AuthType, AuthHeader string
	CredentialSet, PrivacyConfirmed, Tested              bool
}

type aiRepository interface {
	AI() (repositories.AISettings, error)
	UpdateAI(map[string]any, *int64) error
}

func (s *Service) SetAICredentialKey(key []byte) { s.aiCredentialKey = append([]byte(nil), key...) }

func (s *Service) AIConfiguration() (AIConfigurationView, error) {
	repo, ok := s.repo.(aiRepository)
	if !ok {
		return AIConfigurationView{}, ErrSettingsNotFound
	}
	row, err := repo.AI()
	if err != nil {
		return AIConfigurationView{}, traducir(err)
	}
	return aiView(row), nil
}

// AIConfigurationRequired detects a sealed installation that still needs a valid AI setup. A
// later engine outage does not make the configuration invalid; an unreadable stored credential does.
func (s *Service) AIConfigurationRequired() (bool, error) {
	installed, err := s.EstáInstalada()
	if err != nil || !installed {
		return false, err
	}
	repo, ok := s.repo.(aiRepository)
	if !ok {
		return true, nil
	}
	row, err := repo.AI()
	if err != nil {
		return false, traducir(err)
	}
	if row.TestedAt == nil || strings.TrimSpace(row.Mode) == "" {
		return true, nil
	}
	if len(row.CredentialCiphertext) > 0 {
		if _, err := s.decryptAICredential(row); err != nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) TestAndActivateAI(input AIConfigurationInput, actor *auth.Identity, setup bool) (AIConfigurationView, error) {
	input = normalizeAI(input)
	if err := validateAI(input); err != nil {
		return AIConfigurationView{}, err
	}
	repo, ok := s.repo.(aiRepository)
	if !ok {
		return AIConfigurationView{}, ErrSettingsNotFound
	}
	current, err := repo.AI()
	if err != nil {
		return AIConfigurationView{}, traducir(err)
	}
	credential := input.Credential
	if credential == "" && len(current.CredentialCiphertext) > 0 {
		credential, err = s.decryptAICredential(current)
		if err != nil {
			return AIConfigurationView{}, ErrAICredentialInvalid
		}
	}
	if input.AuthType != "none" && credential == "" {
		return AIConfigurationView{}, ErrAICredentialInvalid
	}
	if s.ia == nil || s.ia.ProbarConfiguracion(input.BaseURL, input.Model, input.Provider, input.AuthType, input.AuthHeader, credential, input.Language) != nil {
		return AIConfigurationView{}, ErrAIUnreachable
	}
	nonce, ciphertext, err := s.encryptAICredential(credential)
	if err != nil {
		return AIConfigurationView{}, err
	}
	now := s.now()
	changes := map[string]any{"mode": input.Mode, "provider": input.Provider, "base_url": input.BaseURL, "model": input.Model, "auth_type": input.AuthType, "auth_header": input.AuthHeader, "tested_at": now}
	if credential != "" {
		changes["credential_version"] = int16(1)
		changes["credential_nonce"] = nonce
		changes["credential_ciphertext"] = ciphertext
	}
	if input.Mode != "local" {
		changes["privacy_confirmed_at"] = now
		changes["privacy_confirmed_in_setup"] = setup
	}
	var actorIDValue *int64
	if actor != nil {
		actorIDValue = actorID(*actor)
		changes["privacy_confirmed_by_id"] = actorIDValue
	}
	if err := repo.UpdateAI(changes, actorIDValue); err != nil {
		return AIConfigurationView{}, traducir(err)
	}
	return s.AIConfiguration()
}

func normalizeAI(in AIConfigurationInput) AIConfigurationInput {
	in.Mode = strings.TrimSpace(in.Mode)
	in.Provider = strings.TrimSpace(in.Provider)
	in.BaseURL = strings.TrimRight(strings.TrimSpace(in.BaseURL), "/")
	in.Model = strings.TrimSpace(in.Model)
	in.AuthType = strings.TrimSpace(in.AuthType)
	in.AuthHeader = strings.TrimSpace(in.AuthHeader)
	in.Language = strings.TrimSpace(in.Language)
	if in.Mode == "local" {
		in.Provider = "local"
		in.AuthType = "none"
	}
	if in.Mode == "remote" {
		in.Provider = "openai-compatible"
	}
	if in.BaseURL == "" {
		switch in.Provider {
		case "openai":
			in.BaseURL = "https://api.openai.com"
		case "claude":
			in.BaseURL = "https://api.anthropic.com"
		case "deepseek":
			in.BaseURL = "https://api.deepseek.com"
		}
	}
	if in.AuthType == "" {
		in.AuthType = "none"
	}
	return in
}

func validateAI(in AIConfigurationInput) error {
	if in.Mode != "local" && in.Mode != "remote" && in.Mode != "provider" {
		return ErrAIModeUnknown
	}
	if in.Model == "" {
		return ErrAIModelRequired
	}
	if in.Provider != "local" && in.Provider != "openai" && in.Provider != "claude" && in.Provider != "deepseek" && in.Provider != "openai-compatible" {
		return ErrAIProviderUnknown
	}
	if (in.Mode == "local") != (in.Provider == "local") {
		return ErrAIProviderUnknown
	}
	if in.AuthType != "none" && in.AuthType != "bearer" && in.AuthType != "header" && in.AuthType != "basic" {
		return ErrAIAuthUnknown
	}
	if in.AuthType == "header" && in.AuthHeader == "" {
		return ErrAIAuthUnknown
	}
	if in.Mode != "local" && !in.PrivacyConfirmed {
		return ErrAIPrivacyRequired
	}
	u, err := url.Parse(in.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return ErrAIURLInvalid
	}
	return nil
}

func aiView(row repositories.AISettings) AIConfigurationView {
	return AIConfigurationView{Mode: row.Mode, Provider: row.Provider, BaseURL: row.BaseURL, Model: row.Model, AuthType: row.AuthType, AuthHeader: row.AuthHeader, CredentialSet: len(row.CredentialCiphertext) > 0, PrivacyConfirmed: row.PrivacyConfirmedAt != nil, Tested: row.TestedAt != nil}
}

func (s *Service) encryptAICredential(value string) ([]byte, []byte, error) {
	if value == "" {
		return nil, nil, nil
	}
	block, err := aes.NewCipher(s.aiCredentialKey)
	if err != nil {
		return nil, nil, fmt.Errorf("ai credential key: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	return nonce, gcm.Seal(nil, nonce, []byte(value), []byte("catalina-support:ai:v1")), nil
}

func (s *Service) decryptAICredential(row repositories.AISettings) (string, error) {
	block, err := aes.NewCipher(s.aiCredentialKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, row.CredentialNonce, row.CredentialCiphertext, []byte("catalina-support:ai:v1"))
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
