package services

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// La contraseña se guarda cifrada y se comprueba sin verla nunca.
func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("una-contraseña-larga")
	if err != nil {
		t.Fatalf("no se pudo cifrar: %v", err)
	}

	if strings.Contains(hash, "una-contraseña-larga") {
		t.Fatal("el hash no puede llevar la contraseña dentro")
	}
	if !strings.HasPrefix(hash, "$2a$12$") && !strings.HasPrefix(hash, "$2b$12$") {
		t.Fatalf("el hash no parece de bcrypt con coste 12: %q", hash[:7])
	}
	if !PasswordMatches(hash, "una-contraseña-larga") {
		t.Fatal("la contraseña correcta debería coincidir")
	}
	if PasswordMatches(hash, "una-contraseña-larg") {
		t.Fatal("una contraseña distinta no debería coincidir")
	}
}

// El mismo texto cifrado dos veces da hashes distintos: si no, mirar la base diría quién tiene la
// misma contraseña que quién.
func TestHashPasswordLlevaSal(t *testing.T) {
	primero, _ := HashPassword("una-contraseña-larga")
	segundo, _ := HashPassword("una-contraseña-larga")

	if primero == segundo {
		t.Fatal("dos hashes de la misma contraseña no pueden ser iguales")
	}
}

// La única regla de la política: **ocho caracteres** (decisión del responsable, 2026-09-25).
func TestPasswordTooShort(t *testing.T) {
	if _, err := HashPassword("1234567"); !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("se esperaba ErrPasswordTooShort y llegó %v", err)
	}

	// Ocho justos valen, y se cuentan caracteres y no bytes: con acentos, ocho letras ocupan más de
	// ocho bytes y tienen que valer igual.
	if _, err := HashPassword("12345678"); err != nil {
		t.Fatalf("ocho caracteres deberían valer: %v", err)
	}
	if _, err := HashPassword(strings.Repeat("á", 8)); err != nil {
		t.Fatalf("ocho caracteres acentuados deberían valer: %v", err)
	}
	if _, err := HashPassword(strings.Repeat("á", 7)); !errors.Is(err, ErrPasswordTooShort) {
		t.Fatal("siete caracteres acentuados no deberían valer")
	}
}

// Una cuenta sin contraseña no coincide con nada, ni con la cadena vacía.
func TestPasswordMatchesSinHash(t *testing.T) {
	if PasswordMatches("", "") {
		t.Fatal("sin hash no puede coincidir nada")
	}
	if PasswordMatches("", "loquesea") {
		t.Fatal("sin hash no puede coincidir nada")
	}
}

// El token son 43 caracteres de base64url, y dos seguidos nunca son iguales.
func TestNewToken(t *testing.T) {
	uno, err := newToken()
	if err != nil {
		t.Fatalf("no se pudo generar: %v", err)
	}
	otro, _ := newToken()

	if len(uno) != 43 {
		t.Fatalf("el token mide %d caracteres y debería medir 43", len(uno))
	}
	if uno == otro {
		t.Fatal("dos tokens seguidos no pueden ser iguales")
	}
	if !TokenIsWellFormed(uno) {
		t.Fatal("el token que generamos debe pasar su propia comprobación de forma")
	}
}

func TestTokenIsWellFormed(t *testing.T) {
	casos := map[string]bool{
		"":                            false,
		"corto":                       false,
		strings.Repeat("a", 43):       true,
		strings.Repeat("a", 42):       false,
		strings.Repeat("a", 44):       false,
		strings.Repeat("!", 43):       false,
		strings.Repeat("a", 42) + " ": false,
	}

	for token, esperado := range casos {
		if obtenido := TokenIsWellFormed(token); obtenido != esperado {
			t.Fatalf("TokenIsWellFormed(%q) = %v y se esperaba %v", token, obtenido, esperado)
		}
	}

	// Se admiten espacios alrededor: vienen de copiar y pegar un enlace.
	if !TokenIsWellFormed("  " + strings.Repeat("a", 43) + " ") {
		t.Fatal("un token copiado con espacios debería valer")
	}
}

// La huella es estable y no se parece al token.
func TestHashToken(t *testing.T) {
	token := strings.Repeat("a", 43)
	huella := hashToken(token)

	if huella == token {
		t.Fatal("la huella no puede ser el token")
	}
	if len(huella) != 64 {
		t.Fatalf("la huella mide %d caracteres y debería medir 64", len(huella))
	}
	if hashToken(token) != huella {
		t.Fatal("la misma entrada debe dar la misma huella")
	}
	if hashToken(strings.Repeat("b", 43)) == huella {
		t.Fatal("dos tokens distintos no pueden dar la misma huella")
	}
}

// El texto de la caducidad que va en el correo: en el idioma global y con la duración que toca.
func TestExpiryTextFor(t *testing.T) {
	casos := []struct {
		language string
		duration time.Duration
		esperado string
	}{
		{"es", InvitationDuration, "24 horas"},
		{"es", RecoveryDuration, "1 hora"},
		{"en", InvitationDuration, "24 hours"},
		{"en", RecoveryDuration, "1 hour"},
		// Un idioma que no conocemos: se cae al español, que es el de la instalación por defecto.
		{"fr", InvitationDuration, "24 horas"},
		{"", RecoveryDuration, "1 hora"},
	}

	for _, caso := range casos {
		if obtenido := expiryTextFor(caso.language, caso.duration); obtenido != caso.esperado {
			t.Fatalf("expiryTextFor(%q, %v) = %q y se esperaba %q", caso.language, caso.duration, obtenido, caso.esperado)
		}
	}
}
