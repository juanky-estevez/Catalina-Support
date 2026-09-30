package repositories

import (
	"testing"
)

// TestAddObserverEsIdempotente va contra la base de verdad porque lo que se comprueba es el
// `ON CONFLICT`: añadir dos veces a la misma persona deja **una sola fila**, y el repositorio **dice
// si la añadió** —`false` la segunda vez—, que es lo que permite no repetir la entrada del historial
// (docs/modules/tickets.md, secciones 2.3.1 y 5, decisión 74).
func TestAddObserverEsIdempotente(t *testing.T) {
	_, _, tx := repositorioDePrueba(t)

	conversacion := NewConversationRepository(tx)
	principal := primerTicket(t, tx)
	cuenta := primerUsuario(t, tx)
	destino := DestinoDePrincipal(principal)

	// Se parte de un hilo sin observadores, para contar desde cero.
	if _, err := conversacion.RemoveObserver(tx, cuenta, destino); err != nil {
		t.Fatalf("RemoveObserver falló: %v", err)
	}

	nuevo, err := conversacion.AddObserver(tx, TicketObserver{
		TicketID: destino.TicketID, AccountID: cuenta, AddedByID: cuenta,
	})
	if err != nil {
		t.Fatalf("AddObserver falló: %v", err)
	}
	if !nuevo {
		t.Fatal("la primera vez tenía que decir que lo añadió")
	}

	otraVez, err := conversacion.AddObserver(tx, TicketObserver{
		TicketID: destino.TicketID, AccountID: cuenta, AddedByID: cuenta,
	})
	if err != nil {
		t.Fatalf("repetir AddObserver no debería fallar: %v", err)
	}
	if otraVez {
		t.Fatal("la segunda vez no lo añadió y no puede decir que sí")
	}

	observadores, err := conversacion.Observers(tx, destino)
	if err != nil {
		t.Fatalf("Observers falló: %v", err)
	}

	cuantos := 0
	for _, observador := range observadores {
		if observador.AccountID == cuenta {
			cuantos++
		}
	}
	if cuantos != 1 {
		t.Fatalf("se esperaba una sola fila del observador %d y hay %d", cuenta, cuantos)
	}
}
