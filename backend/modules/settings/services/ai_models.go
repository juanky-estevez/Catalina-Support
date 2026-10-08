package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"catalina-support/backend/shared/auth"
)

var (
	ErrAIManagerUnavailable = errors.New("settings.ai.managerUnavailable")
	ErrAILicenseRequired    = errors.New("settings.ai.licenseRequired")
)

type AIInsufficientMemoryError struct {
	RequiredBytes  int64
	AvailableBytes int64
}

func (e *AIInsufficientMemoryError) Error() string { return "settings.ai.insufficientMemory" }

type AIModel struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Version         string `json:"version"`
	License         string `json:"license"`
	Source          string `json:"source"`
	Checksum        string `json:"checksum"`
	DownloadBytes   int64  `json:"downloadBytes"`
	RAMBytes        int64  `json:"ramBytes"`
	Installed       bool   `json:"installed"`
	Active          bool   `json:"active"`
	Healthy         bool   `json:"healthy"`
	Downloading     bool   `json:"downloading"`
	DownloadedBytes int64  `json:"downloadedBytes"`
	Error           string `json:"error,omitempty"`
}
type AIModelCatalog struct {
	Models             []AIModel `json:"models"`
	DiskAvailableBytes int64     `json:"diskAvailableBytes"`
}
type aiModelAcceptanceRepository interface {
	AcceptAIModel(string, string, string, *int64, bool) error
}

func (s *Service) SetAIManager(baseURL, token string) {
	s.aiManagerURL = strings.TrimRight(baseURL, "/")
	s.aiManagerToken = token
}
func (s *Service) AIModels() (AIModelCatalog, error) {
	var out AIModelCatalog
	err := s.aiManager(http.MethodGet, "/models", nil, &out)
	return out, err
}
func (s *Service) DownloadAIModel(id string, accepted bool, actor *auth.Identity, setup bool) error {
	if !accepted {
		return ErrAILicenseRequired
	}
	catalog, err := s.AIModels()
	if err != nil {
		return err
	}
	var chosen *AIModel
	for i := range catalog.Models {
		if catalog.Models[i].ID == id {
			chosen = &catalog.Models[i]
			break
		}
	}
	if chosen == nil {
		return ErrAIModelRequired
	}
	var actorIDValue *int64
	if actor != nil {
		actorIDValue = actorID(*actor)
	}
	if repo, ok := s.repo.(aiModelAcceptanceRepository); ok {
		if err := repo.AcceptAIModel(chosen.ID, chosen.Version, chosen.Checksum, actorIDValue, setup); err != nil {
			return traducir(err)
		}
	}
	return s.aiManager(http.MethodPost, "/models/"+url.PathEscape(id)+"/download?acceptLicense=true", bytes.NewReader(nil), nil)
}
func (s *Service) ActivateAIModel(id string) error {
	return s.aiManager(http.MethodPost, "/models/"+url.PathEscape(id)+"/activate", bytes.NewReader(nil), nil)
}
func (s *Service) DeleteAIModel(id string) error {
	return s.aiManager(http.MethodDelete, "/models/"+url.PathEscape(id), nil, nil)
}
func (s *Service) aiManager(method, path string, body io.Reader, out any) error {
	if s.aiManagerURL == "" || s.aiManagerToken == "" {
		return ErrAIManagerUnavailable
	}
	req, err := http.NewRequest(method, s.aiManagerURL+path, body)
	if err != nil {
		return ErrAIManagerUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+s.aiManagerToken)
	timeout := 10 * time.Second
	if strings.HasSuffix(path, "/activate") {
		timeout = 310 * time.Second
	}
	client := http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return ErrAIManagerUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if resp.StatusCode == http.StatusUnprocessableEntity {
			var failure struct {
				Key            string `json:"key"`
				RequiredBytes  int64  `json:"requiredBytes"`
				AvailableBytes int64  `json:"availableBytes"`
			}
			if json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&failure) == nil &&
				failure.Key == "settings.ai.insufficientMemory" && failure.RequiredBytes >= 0 && failure.AvailableBytes >= 0 {
				return &AIInsufficientMemoryError{RequiredBytes: failure.RequiredBytes, AvailableBytes: failure.AvailableBytes}
			}
		}
		return ErrAIManagerUnavailable
	}
	if out != nil && json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out) != nil {
		return ErrAIManagerUnavailable
	}
	return nil
}
