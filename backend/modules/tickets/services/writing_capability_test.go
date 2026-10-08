package services

import (
	"testing"

	"catalina-support/backend/shared/auth"
)

type asistenteDeRedaccionDePrueba struct{ configurado bool }

func (a asistenteDeRedaccionDePrueba) Configured() bool { return a.configurado }
func (a asistenteDeRedaccionDePrueba) ImproveDraft(string, string, string, string, string) (string, error) {
	return "", nil
}

func TestPuedeMejorarRedaccionRespetaPapelEditorYCierre(t *testing.T) {
	principal := Ticket{State: "en progreso"}
	interno := Ticket{Internal: true, State: "en progreso"}

	casos := []struct {
		nombre string
		actor  auth.Identity
		ticket Ticket
		quiere bool
	}{
		{"Soporte en principal", auth.Identity{Role: auth.RoleSoporte}, principal, true},
		{"Desarrollo en principal", auth.Identity{Role: auth.RoleDesarrollo}, principal, false},
		{"Soporte en interno", auth.Identity{Role: auth.RoleSoporte}, interno, true},
		{"Desarrollo en interno", auth.Identity{Role: auth.RoleDesarrollo}, interno, true},
		{"Usuario", auth.Identity{Role: auth.RoleUsuario}, principal, false},
		{"Administrador", auth.Identity{Role: auth.RoleAdministrador}, principal, false},
		{"Cerrado", auth.Identity{Role: auth.RoleSoporte}, Ticket{State: "cerrado"}, false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := puedeMejorarRedaccion(caso.actor, caso.ticket); obtenido != caso.quiere {
				t.Fatalf("puedeMejorarRedaccion() = %v; se esperaba %v", obtenido, caso.quiere)
			}
		})
	}
}

func TestCapacidadDeRedaccionExigeConfiguracion(t *testing.T) {
	actor := auth.Identity{Role: auth.RoleSoporte}
	ticket := Ticket{State: "en progreso"}
	for _, caso := range []struct {
		nombre    string
		asistente WritingAssistant
		quiere    bool
	}{
		{"sin adaptador", nil, false},
		{"sin configuración", asistenteDeRedaccionDePrueba{configurado: false}, false},
		{"configurada", asistenteDeRedaccionDePrueba{configurado: true}, true},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			servicio := &Service{writing: caso.asistente}
			if obtenido := servicio.capacidadDeRedaccion(actor, ticket); obtenido != caso.quiere {
				t.Fatalf("capacidadDeRedaccion() = %v; se esperaba %v", obtenido, caso.quiere)
			}
		})
	}
}
