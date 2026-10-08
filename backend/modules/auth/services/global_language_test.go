package services

import "testing"

type idiomaDeCuentaDePrueba struct {
	idioma string
	err    error
}

func (i idiomaDeCuentaDePrueba) Language() (string, error) { return i.idioma, i.err }

func TestIdiomaGlobalDeLosCorreosDeCuenta(t *testing.T) {
	casos := []struct {
		nombre    string
		proveedor IdiomaGlobal
		esperado  string
	}{
		{"inglés configurado", idiomaDeCuentaDePrueba{idioma: "en"}, "en"},
		{"español configurado", idiomaDeCuentaDePrueba{idioma: "es"}, "es"},
		{"sin configuración", nil, "es"},
		{"valor desconocido", idiomaDeCuentaDePrueba{idioma: "fr"}, "es"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			servicio := &Service{idioma: caso.proveedor}
			if obtenido := servicio.idiomaGlobal(); obtenido != caso.esperado {
				t.Fatalf("idiomaGlobal() = %q; se esperaba %q", obtenido, caso.esperado)
			}
		})
	}
}
