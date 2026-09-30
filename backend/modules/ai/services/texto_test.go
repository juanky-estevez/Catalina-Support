package services

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestElRecorteQuitaPorElPrincipioYConservaElFinal(t *testing.T) {
	principio := "PRINCIPIO: esto ya no importa\n" + strings.Repeat("relleno que hace bulto\n", 400)
	texto := principio + "Última acción: se pidió una captura"

	recortado := recortar(texto)

	if utf8.RuneCountInString(recortado) > MaxTexto {
		t.Errorf("el texto recortado tiene %d caracteres y el tope es %d", utf8.RuneCountInString(recortado), MaxTexto)
	}
	if strings.Contains(recortado, "PRINCIPIO") {
		t.Error("el recorte tiene que quitar por el principio")
	}
	if !strings.HasSuffix(recortado, "Última acción: se pidió una captura") {
		t.Error("el recorte tiene que conservar el final del texto, que es lo último que pasó")
	}
}

func TestElRecorteDejaElTextoQueYaCabe(t *testing.T) {
	texto := "Asunto: No puedo entrar\n\nConversación (de lo más antiguo a lo más reciente):\n- user1 (comentario): Sigue fallando.\n"

	if recortado := recortar(texto); recortado != strings.TrimSpace(texto) {
		t.Errorf("un texto que cabe no se toca, y quedó %q", recortado)
	}
}

func TestElRecorteNoParteUnaLetraPorLaMitad(t *testing.T) {
	// Cada «á» ocupa dos bytes: cortando por bytes, el texto saldría con una letra rota.
	texto := strings.Repeat("á", MaxTexto*2)

	recortado := recortar(texto)

	if !utf8.ValidString(recortado) {
		t.Error("el recorte tiene que dejar un texto válido, sin media letra al principio")
	}
	if letras := utf8.RuneCountInString(recortado); letras > MaxTexto {
		t.Errorf("el texto recortado tiene %d caracteres y el tope es %d", letras, MaxTexto)
	}
}

func TestElRecorteEmpiezaEnUnaLineaEnteraCuandoPuede(t *testing.T) {
	// El corte cae cuatro caracteres antes del final de una línea y hay un salto cerca, así que se
	// empieza en la línea siguiente en vez de a media palabra. Las cuentas: la cabeza mide 5 941
	// caracteres, la cola 5 996, y el corte empieza en el 5 937 —cuatro «a» antes del salto—.
	cabeza := strings.Repeat("a", 5940) + "\n"
	cola := "EL FINAL" + strings.Repeat("z", 5988)
	texto := cabeza + cola

	recortado := recortar(texto)

	if !strings.HasPrefix(recortado, "EL FINAL") {
		t.Errorf("el recorte tendría que empezar en la línea siguiente, y empieza por %q", recortado[:20])
	}
}

func TestElTopeDePalabrasRecortaYLoDice(t *testing.T) {
	recortado := recortarPalabras("una dos tres cuatro cinco seis", 3)

	if recortado != "una dos tres…" {
		t.Errorf("el recorte de palabras quedó %q", recortado)
	}

	if corto := recortarPalabras("una dos", 3); corto != "una dos" {
		t.Errorf("un texto que cabe no se toca, y quedó %q", corto)
	}

	if sinTope := recortarPalabras("una dos tres", 0); sinTope != "una dos tres" {
		t.Errorf("sin tope no se recorta nada, y quedó %q", sinTope)
	}
}
