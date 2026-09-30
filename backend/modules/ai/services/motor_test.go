package services

import (
	"strings"
	"testing"
)

// El motor de verdad es un modelo pequeño: se le pide un JSON y contesta con el JSON, o con el JSON
// dentro de un bloque de código, o con una frase delante. Estas pruebas fijan qué se acepta y qué no.

func TestPrimerObjetoJSON(t *testing.T) {
	casos := []struct {
		nombre   string
		texto    string
		esperado string
		hay      bool
	}{
		{
			nombre:   "el JSON a secas",
			texto:    `{"es": "Uno", "en": "One"}`,
			esperado: `{"es": "Uno", "en": "One"}`,
			hay:      true,
		},
		{
			nombre:   "envuelto en un bloque de código",
			texto:    "```json\n{\"es\": \"Uno\", \"en\": \"One\"}\n```",
			esperado: `{"es": "Uno", "en": "One"}`,
			hay:      true,
		},
		{
			nombre:   "con una frase delante y otra detrás",
			texto:    "Claro, aquí lo tienes: {\"es\": \"Uno\", \"en\": \"One\"} Espero que sirva.",
			esperado: `{"es": "Uno", "en": "One"}`,
			hay:      true,
		},
		{
			nombre:   "con llaves dentro de una cadena",
			texto:    `{"es": "Se ve {esto} raro", "en": "It looks {weird}"}`,
			esperado: `{"es": "Se ve {esto} raro", "en": "It looks {weird}"}`,
			hay:      true,
		},
		{
			nombre:   "con una llave escapada dentro de una cadena",
			texto:    `{"es": "Usa \" así", "en": "Use \" like this"}`,
			esperado: `{"es": "Usa \" así", "en": "Use \" like this"}`,
			hay:      true,
		},
		{
			nombre:   "un objeto sin cerrar",
			texto:    `{"es": "Uno", "en": "One"`,
			esperado: "",
			hay:      false,
		},
		{
			nombre:   "sin ninguna llave",
			texto:    "No sé qué decirte.",
			esperado: "",
			hay:      false,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			objeto, hay := primerObjetoJSON(caso.texto)

			if hay != caso.hay {
				t.Fatalf("hay=%v y se esperaba %v", hay, caso.hay)
			}
			if objeto != caso.esperado {
				t.Errorf("el objeto es %q y se esperaba %q", objeto, caso.esperado)
			}
		})
	}
}

func TestDosIdiomas(t *testing.T) {
	casos := []struct {
		nombre string
		texto  string
		hay    bool
	}{
		{
			nombre: "las dos redacciones",
			texto:  `{"es": "El usuario no puede entrar.", "en": "The user cannot sign in."}`,
			hay:    true,
		},
		{
			nombre: "con espacios de más",
			texto:  `{"es": "  El usuario no puede entrar.  ", "en": "  The user cannot sign in.  "}`,
			hay:    true,
		},
		{
			nombre: "sólo el español: media respuesta no es una respuesta",
			texto:  `{"es": "El usuario no puede entrar."}`,
			hay:    false,
		},
		{
			nombre: "el inglés vacío",
			texto:  `{"es": "El usuario no puede entrar.", "en": "   "}`,
			hay:    false,
		},
		{
			nombre: "un JSON que no es el pedido",
			texto:  `{"resumen": "El usuario no puede entrar."}`,
			hay:    false,
		},
		{
			nombre: "texto que no es JSON",
			texto:  "El usuario no puede entrar, seguramente sea la contraseña.",
			hay:    false,
		},
		{
			nombre: "un JSON roto",
			texto:  `{"es": "El usuario no puede entrar.", "en": }`,
			hay:    false,
		},
		{
			nombre: "vacío",
			texto:  "",
			hay:    false,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			es, en, hay := dosIdiomas(caso.texto)

			if hay != caso.hay {
				t.Fatalf("hay=%v y se esperaba %v (es=%q en=%q)", hay, caso.hay, es, en)
			}
			if hay && (es == "" || en == "") {
				t.Errorf("cuando vale, las dos redacciones tienen que venir: es=%q en=%q", es, en)
			}
			if hay && (es != strings.TrimSpace(es) || en != strings.TrimSpace(en)) {
				t.Errorf("las redacciones tienen que venir sin espacios de más: es=%q en=%q", es, en)
			}
		})
	}
}

func TestTopeDePiezasNuncaPasaDeMediaVentana(t *testing.T) {
	casos := []struct {
		nombre   string
		ajustes  Ajustes
		esperado int
	}{
		{"la ventana manda cuando el tope de palabras es grande", Ajustes{Palabras: 1000, Contexto: 4096}, 2048},
		{"el tope de palabras manda cuando la ventana es grande", Ajustes{Palabras: 40, Contexto: 32768}, 320},
		{"con 60 palabras, 400 piezas de margen", Ajustes{Palabras: 60, Contexto: 4096}, 400},
		{"con pocas palabras queda el mínimo", Ajustes{Palabras: 1, Contexto: 4096}, 164},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if piezas := topeDePiezas(caso.ajustes); piezas != caso.esperado {
				t.Errorf("el tope es %d y se esperaba %d", piezas, caso.esperado)
			}
		})
	}
}

func TestDireccionDeJuntaElCaminoSinBarrasDeMas(t *testing.T) {
	casos := map[string]string{
		"http://catalina_support_ai:8080":     "http://catalina_support_ai:8080/health",
		"http://catalina_support_ai:8080/":    "http://catalina_support_ai:8080/health",
		" http://catalina_support_ai:8080// ": "http://catalina_support_ai:8080/health",
	}

	for base, esperado := range casos {
		if direccion := direccionDe(base, "/health"); direccion != esperado {
			t.Errorf("la dirección de %q es %q y se esperaba %q", base, direccion, esperado)
		}
	}
}
