package auth

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// El secreto de las pruebas, largo y distinto del de nadie.
const secretoDePrueba = "un-secreto-de-pruebas-largo-y-aburrido"

// Un `state` que acaba de salir vale, y el mismo sirve una sola vez porque cada uno es distinto.
func TestStateRedondo(t *testing.T) {
	ahora := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)

	uno, err := SignState(secretoDePrueba, ahora)
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}
	otro, _ := SignState(secretoDePrueba, ahora)

	if uno == otro {
		t.Fatal("dos states seguidos no pueden ser iguales")
	}
	if err := VerifyState(secretoDePrueba, uno, ahora.Add(2*time.Minute)); err != nil {
		t.Fatalf("debería valer: %v", err)
	}
	if err := VerifyState(secretoDePrueba, otro, ahora.Add(2*time.Minute)); err != nil {
		t.Fatalf("debería valer: %v", err)
	}
}

// **A los cinco minutos y un segundo ya no vale.** Es lo único que impide que una vuelta vieja siga
// sirviendo, porque no hay tabla de estados pendientes (docs/modules/auth.md, sección 5.3).
func TestStateCaduca(t *testing.T) {
	ahora := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)
	state, _ := SignState(secretoDePrueba, ahora)

	if err := VerifyState(secretoDePrueba, state, ahora.Add(StateDuration)); err != nil {
		t.Fatalf("justo en el límite todavía debería valer: %v", err)
	}
	if err := VerifyState(secretoDePrueba, state, ahora.Add(StateDuration+time.Second)); !errors.Is(err, ErrStateExpired) {
		t.Fatalf("se esperaba ErrStateExpired y llegó %v", err)
	}
}

// Un `state` de otro secreto, uno con la firma retocada y uno inventado no valen: es lo que impide
// que alguien se fabrique la vuelta.
func TestStateFalseado(t *testing.T) {
	ahora := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)
	state, _ := SignState(secretoDePrueba, ahora)

	if err := VerifyState("otro-secreto-distinto", state, ahora); !errors.Is(err, ErrStateInvalid) {
		t.Fatalf("con otro secreto se esperaba ErrStateInvalid y llegó %v", err)
	}

	partes := strings.Split(state, ".")
	partes[2] = strings.Repeat("A", len(partes[2]))
	if err := VerifyState(secretoDePrueba, strings.Join(partes, "."), ahora); !errors.Is(err, ErrStateInvalid) {
		t.Fatalf("con la firma retocada se esperaba ErrStateInvalid y llegó %v", err)
	}

	// Y si se estira la caducidad cambiando la fecha, la firma deja de cuadrar: por eso la fecha va
	// dentro de lo firmado.
	partes = strings.Split(state, ".")
	partes[1] = "0"
	if err := VerifyState(secretoDePrueba, strings.Join(partes, "."), ahora); !errors.Is(err, ErrStateInvalid) {
		t.Fatalf("con la fecha cambiada se esperaba ErrStateInvalid y llegó %v", err)
	}

	for _, inventado := range []string{"", "loquesea", "a.b", "a.b.c.d", ".."} {
		if err := VerifyState(secretoDePrueba, inventado, ahora); err == nil {
			t.Fatalf("%q no debería valer como state", inventado)
		}
	}
}

// Un `state` emitido en el futuro no vale: o el reloj va mal o alguien está probando a estirar la
// caducidad. Con un minuto de margen por los desajustes de reloj entre servidores.
func TestStateDelFuturo(t *testing.T) {
	ahora := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)

	dentroDeUnRato, _ := SignState(secretoDePrueba, ahora.Add(30*time.Second))
	if err := VerifyState(secretoDePrueba, dentroDeUnRato, ahora); err != nil {
		t.Fatalf("medio minuto de desajuste debería tolerarse: %v", err)
	}

	mañana, _ := SignState(secretoDePrueba, ahora.Add(24*time.Hour))
	if err := VerifyState(secretoDePrueba, mañana, ahora); !errors.Is(err, ErrStateInvalid) {
		t.Fatalf("se esperaba ErrStateInvalid y llegó %v", err)
	}
}

// Sin secreto no se firma ni se comprueba nada: es la misma red que protege al token de sesión.
func TestStateSinSecreto(t *testing.T) {
	if _, err := SignState("  ", time.Now()); !errors.Is(err, ErrNoSecret) {
		t.Fatalf("se esperaba ErrNoSecret y llegó %v", err)
	}
	if err := VerifyState("", "a.b.c", time.Now()); !errors.Is(err, ErrNoSecret) {
		t.Fatalf("se esperaba ErrNoSecret y llegó %v", err)
	}
}
