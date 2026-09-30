package services

import (
	"errors"
	"strings"
	"testing"

	"catalina-support/backend/shared/auth"
)

// El color institucional: se elige uno, pero la aplicación necesita dos (el del tema claro y el del
// oscuro) y el texto que va encima. Esto comprueba que **siempre** se lee, elija lo que elija el
// administrador.
func TestColorsFromSeLeeSiempre(t *testing.T) {
	// Colores difíciles a propósito: uno casi blanco, uno casi negro, uno chillón y uno apagado.
	colores := []string{
		"#1d4ed8", "#ffffff", "#000000", "#ffff00", "#7f7f7f",
		"#ff0000", "#00ff00", "#0a0a0a", "#f5f5f5", "#8a2be2",
	}

	for _, elegido := range colores {
		colors := colorsFrom(elegido)

		if razon := contraste(colors.Light, "#ffffff"); razon < 4.5 {
			t.Fatalf("con %s, el color del tema claro se lee a %.2f sobre blanco", elegido, razon)
		}
		if razon := contraste(colors.Dark, "#14181d"); razon < 4.5 {
			t.Fatalf("con %s, el color del tema oscuro se lee a %.2f sobre el fondo oscuro", elegido, razon)
		}
		// Y el texto del botón, sobre su color.
		if razon := contraste(colors.OnLight, colors.Light); razon < 4.5 {
			t.Fatalf("con %s, el texto del botón claro se lee a %.2f", elegido, razon)
		}
		if razon := contraste(colors.OnDark, colors.Dark); razon < 4.5 {
			t.Fatalf("con %s, el texto del botón oscuro se lee a %.2f", elegido, razon)
		}
	}
}

// El color de fábrica no se toca: es el que trae la aplicación y ya se lee.
func TestColorsFromNoTocaElDeFabrica(t *testing.T) {
	colors := colorsFrom(colorInstitucionalDeFabrica)

	if !colors.IsDefault {
		t.Fatal("el color de fábrica debería reconocerse como tal")
	}
	if colors.Light != colorInstitucionalDeFabrica {
		t.Fatalf("el color claro es %q y debería ser el elegido tal cual", colors.Light)
	}
}

// Un color que no es un hexadecimal vale lo mismo que el de fábrica: la marca no puede romper el
// arranque de la aplicación.
func TestColorsFromConBasura(t *testing.T) {
	for _, basura := range []string{"", "azul", "#12345", "#gggggg", "rgb(1,2,3)"} {
		colors := colorsFrom(basura)
		if colors.Light == "" || colors.Dark == "" {
			t.Fatalf("con %q debería caer al color de fábrica", basura)
		}
	}
}

// El texto que va encima es el que más contrasta, no siempre blanco.
func TestTextoEncima(t *testing.T) {
	if textoEncima("#ffffff") != "#0d0f12" {
		t.Fatal("sobre blanco el texto tiene que ser oscuro")
	}
	if textoEncima("#000000") != "#ffffff" {
		t.Fatal("sobre negro el texto tiene que ser claro")
	}
	if textoEncima("#ffff00") != "#0d0f12" {
		t.Fatal("sobre amarillo el texto tiene que ser oscuro")
	}
}

// La combinación que no tiene sentido: avisar al asignado sin repartir.
func TestSoloAvisaAlAsignado(t *testing.T) {
	if !soloAvisaAlAsignado("ninguna", "al_asignado") {
		t.Fatal("avisar al asignado sin repartir debería rechazarse")
	}
	if soloAvisaAlAsignado("por_turnos", "al_asignado") {
		t.Fatal("repartir y avisar al asignado es lo normal")
	}
	if soloAvisaAlAsignado("ninguna", "a_nadie") {
		t.Fatal("no repartir y no avisar es válido")
	}
}

// Las validaciones se cortan antes de tocar la base: el repositorio es nil a propósito, así que si
// alguna dejara pasar el valor, la prueba caería.
func TestUpdateValidaAntesDeLaBase(t *testing.T) {
	servicio := NewService(nil, t.TempDir())
	administrador := auth.Identity{ID: 1, Role: auth.RoleAdministrador}

	valido := UpdateInput{
		Name:                 NombreDeFabrica,
		EntryMethod:          auth.MethodLocal,
		Language:             "es",
		PrimaryColor:         "#1d4ed8",
		NumberPrefix:         "CS",
		MainAssignment:       "por_turnos",
		MainNotification:     "al_asignado",
		InternalAssignment:   "ninguna",
		InternalNotification: "a_nadie",
	}

	casos := []struct {
		nombre string
		cambia func(*UpdateInput)
		clave  error
	}{
		{"idioma inventado", func(i *UpdateInput) { i.Language = "fr" }, ErrLanguageUnknown},
		{
			// El tope se cuenta en caracteres, no en bytes: «á» es uno.
			"nombre de sesenta y uno",
			func(i *UpdateInput) { i.Name = strings.Repeat("a", MaxNameLength+1) },
			ErrNameTooLong,
		},
		{
			"nombre de sesenta y uno con acentos",
			func(i *UpdateInput) { i.Name = strings.Repeat("á", MaxNameLength+1) },
			ErrNameTooLong,
		},
		{"color que no es un color", func(i *UpdateInput) { i.PrimaryColor = "azul" }, ErrPrimaryColorInvalid},
		{"prefijo en minúsculas", func(i *UpdateInput) { i.NumberPrefix = "cs" }, ErrNumberPrefixInvalid},
		{"prefijo de una letra", func(i *UpdateInput) { i.NumberPrefix = "C" }, ErrNumberPrefixInvalid},
		{"prefijo de nueve", func(i *UpdateInput) { i.NumberPrefix = "ABCDEFGHI" }, ErrNumberPrefixInvalid},
		{"asignación inventada", func(i *UpdateInput) { i.MainAssignment = "a_mano" }, ErrAssignmentUnknown},
		{"aviso inventado", func(i *UpdateInput) { i.MainNotification = "por_correo" }, ErrNotificationUnknown},
		{
			"avisar al asignado sin repartir",
			func(i *UpdateInput) { i.MainAssignment = "ninguna"; i.MainNotification = "al_asignado" },
			ErrNotificationWithoutAssigment,
		},
		{
			"lo mismo en los internos",
			func(i *UpdateInput) { i.InternalAssignment = "ninguna"; i.InternalNotification = "al_asignado" },
			ErrNotificationWithoutAssigment,
		},
		{"método inventado", func(i *UpdateInput) { i.EntryMethod = "saml" }, ErrMethodUnknown},
		{"método vacío", func(i *UpdateInput) { i.EntryMethod = "" }, ErrMethodUnknown},
		{
			// **No se elige un método a medias**: con `ad` sin directorio, la instalación se quedaría sin
			// puerta para todo el mundo menos la cuenta de fábrica.
			"AD sin directorio configurado",
			func(i *UpdateInput) { i.EntryMethod = auth.MethodAD },
			ErrMethodNotConfigured,
		},
		{
			"Keycloak sin reino configurado",
			func(i *UpdateInput) { i.EntryMethod = auth.MethodKeycloak },
			ErrMethodNotConfigured,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			entrada := valido
			caso.cambia(&entrada)

			_, err := servicio.Update(entrada, administrador)
			if !errors.Is(err, caso.clave) {
				t.Fatalf("se esperaba %v y llegó %v", caso.clave, err)
			}
		})
	}
}

// Y con datos buenos, el corte ya no es la validación: se llega a la base. Con el repositorio en nil
// eso es un pánico, así que se envuelve para comprobar que pasó de la validación.
func TestUpdateConDatosBuenosLlegaALaBase(t *testing.T) {
	servicio := NewService(nil, t.TempDir())

	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("con datos válidos debería intentar guardar, y el repositorio es nil")
		}
	}()

	_, _ = servicio.Update(UpdateInput{
		// Sesenta justos: el tope está donde dice, ni uno menos.
		Name:                 strings.Repeat("á", MaxNameLength),
		EntryMethod:          auth.MethodLocal,
		Language:             "es",
		PrimaryColor:         "#1d4ed8",
		NumberPrefix:         "ACME",
		MainAssignment:       "ninguna",
		MainNotification:     "a_todo_el_equipo",
		InternalAssignment:   "ninguna",
		InternalNotification: "a_nadie",
		// **La zona horaria y la dirección son obligatorias en la validación** (decisiones 14 y 15):
		// sin ellas no se llega a la base, que es lo que esta prueba comprueba.
		TimeZone:     "America/Guayaquil",
		PublicAppURL: "https://soporte.example.org",
	}, auth.Identity{ID: 1})
}
