package services

import (
	"strings"
	"testing"
)

type translatorFake struct{ commercial bool; alter bool }
func (f translatorFake) EsProveedorComercial() bool { return f.commercial }
func (f translatorFake) Traducir(text, language string) (string, error) {
	if f.alter { return strings.Replace(text, "[[CS_000]]", "", 1), nil }
	return "translated " + text, nil
}

func TestTranslateProtectedRestoresHTMLAndMarkers(t *testing.T) {
	in := `<p>Hello {{nombre}}</p>`
	out, err := translateProtected(in, "es", translatorFake{})
	if err != nil { t.Fatal(err) }
	if !strings.Contains(out, "<p>") || !strings.Contains(out, "{{nombre}}") { t.Fatalf("protected pieces were not restored: %s", out) }
}

func TestTranslateProtectedRejectsChangedToken(t *testing.T) {
	if _, err := translateProtected(`<p>Hello</p>`, "es", translatorFake{alter:true}); err == nil {
		t.Fatal("expected changed protected token to be rejected")
	}
}

func TestCommercialDraftRequiresExplicitConfirmation(t *testing.T) {
	service := &Service{}
	if _, err := service.LanguageDrafts("en", translatorFake{commercial:true}, false); err == nil || !strings.Contains(err.Error(), "costConfirmationRequired") {
		t.Fatalf("expected cost confirmation error, got %v", err)
	}
}
