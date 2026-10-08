package services

import "testing"

type idiomaGlobalDePrueba struct {
	idioma string
	err    error
}

func (i idiomaGlobalDePrueba) Language() (string, error) { return i.idioma, i.err }

func TestIdiomaGlobalDeLosAvisos(t *testing.T) {
	casos := []struct {
		nombre    string
		proveedor GlobalLanguage
		esperado  string
	}{
		{"inglés configurado", idiomaGlobalDePrueba{idioma: "en"}, "en"},
		{"español configurado", idiomaGlobalDePrueba{idioma: "es"}, "es"},
		{"sin configuración", nil, "es"},
		{"valor desconocido", idiomaGlobalDePrueba{idioma: "fr"}, "es"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			servicio := &Service{language: caso.proveedor}
			if obtenido := servicio.idiomaGlobal(); obtenido != caso.esperado {
				t.Fatalf("idiomaGlobal() = %q; se esperaba %q", obtenido, caso.esperado)
			}
		})
	}
}
