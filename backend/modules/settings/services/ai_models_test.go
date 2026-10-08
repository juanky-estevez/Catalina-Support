package services

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAIManagerPreservesStructuredMemoryError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatal("missing manager token")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"key":"settings.ai.insufficientMemory","requiredBytes":4294967296,"availableBytes":3758096384}`))
	}))
	defer server.Close()

	service := NewService(nil, "")
	service.SetAIManager(server.URL, "secret")
	err := service.ActivateAIModel("qwen2.5-3b-instruct")
	var memory *AIInsufficientMemoryError
	if !errors.As(err, &memory) {
		t.Fatalf("expected memory error, got %v", err)
	}
	if memory.RequiredBytes != 4294967296 || memory.AvailableBytes != 3758096384 {
		t.Fatalf("unexpected details: %+v", memory)
	}
}
