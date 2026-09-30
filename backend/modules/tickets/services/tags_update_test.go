package services

import (
	"errors"
	"testing"

	"gorm.io/gorm"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// Estas pruebas cubren la decisión 85: **las etiquetas las ponen los dos equipos, y son del principal**.
// Se comprueban contra la base de datos **de verdad** —dentro de una transacción que se deshace al
// terminar— porque lo que importa es dónde acaban: en el principal, aunque el número que llegue sea el
// del interno. Sin base a mano, se saltan como las demás de este estilo.

// TestDesarrolloCambiaLasEtiquetasDesdeElInterno es el encargo: Desarrollo manda las etiquetas **al
// número del interno** y quedan en **el principal**, que es de quien son (decisiones 67 y 85). También
// se lee la ficha con la que contesta, para que lo que se enseña sea lo que de verdad hay.
func TestDesarrolloCambiaLasEtiquetasDesdeElInterno(t *testing.T) {
	servicio, tx, _ := servicioDeObservadores(t)

	interno, principal := internoAbierto(t, tx)
	desarrollo1 := cuentaDeLaBase(t, tx, auth.RoleDesarrollo, true)
	antes := etiquetasDe(t, servicio, principal.ID)
	catalogo := etiquetasDelCatalogoQueNoLleva(t, tx, antes, 2)

	// La ficha del interno trae lo que cuelga de él: el dueño del ticket y poco más hacen falta.
	servicio.SetAccounts(cuentasStub{desarrollo1.ID: desarrollo1})

	actor := auth.Identity{ID: desarrollo1.ID, Role: auth.RoleDesarrollo}

	detalle, err := servicio.UpdateTags(interno.Number, catalogo, actor)
	if err != nil {
		t.Fatalf("Desarrollo debería poder etiquetar desde el interno: %v", err)
	}
	if !detalle.Internal || detalle.Number != interno.Number {
		t.Fatalf("la ficha devuelta debería ser el interno %q: %+v", interno.Number, detalle)
	}

	// **En el principal**, que es donde viven las etiquetas: el interno no tiene tabla propia.
	enElPrincipal := etiquetasDe(t, servicio, principal.ID)
	if !mismasEtiquetas(enElPrincipal, catalogo) {
		t.Fatalf("el principal debería llevar %v y lleva %v", catalogo, enElPrincipal)
	}
	if !mismasEtiquetas(detalle.Tags, catalogo) {
		t.Fatalf("la ficha del interno debería heredar %v y enseña %v", catalogo, detalle.Tags)
	}
}

// TestDesarrolloNoTocaElPrincipal: Desarrollo no escribe el texto del principal —ni el asunto, ni la
// descripción, ni la categoría— y **tampoco sus etiquetas por el número del principal**: la 85 le deja
// etiquetar *desde el interno*, no cambiar el principal. Las dos puertas responden `403`.
func TestDesarrolloNoTocaElPrincipal(t *testing.T) {
	servicio, tx, _ := servicioDeObservadores(t)

	_, principal := internoAbierto(t, tx)
	desarrollo1 := cuentaDeLaBase(t, tx, auth.RoleDesarrollo, true)
	servicio.SetAccounts(cuentasStub{desarrollo1.ID: desarrollo1})

	actor := auth.Identity{ID: desarrollo1.ID, Role: auth.RoleDesarrollo}

	// El `PATCH` de texto y categoría, como siempre.
	_, err := servicio.Update(principal.Number, UpdateInput{
		Subject:     "otro asunto",
		Description: "<p>otra descripción</p>",
	}, actor)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Desarrollo debería recibir %v al editar el principal y recibió %v", ErrForbidden, err)
	}

	// Y el `PATCH` de sólo etiquetas contra el número del principal: tampoco.
	_, err = servicio.UpdateTags(principal.Number, etiquetasDelCatalogo(t, tx, 1), actor)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Desarrollo debería recibir %v al etiquetar el principal y recibió %v", ErrForbidden, err)
	}
}

// TestSoporteSigueEtiquetandoElPrincipal: la regla de siempre no se toca. Soporte etiqueta el
// principal —por su número— y el solicitante sigue pudiendo su ticket.
func TestSoporteSigueEtiquetandoElPrincipal(t *testing.T) {
	servicio, tx, _ := servicioDeObservadores(t)

	_, principal := internoAbierto(t, tx)
	soporte1 := cuentaDeLaBase(t, tx, auth.RoleSoporte, true)
	solicitante := cuentaPorID(t, tx, principal.RequesterID)
	catalogo := etiquetasDelCatalogo(t, tx, 1)
	servicio.SetAccounts(cuentasStub{soporte1.ID: soporte1, solicitante.ID: solicitante})

	deSoporte := auth.Identity{ID: soporte1.ID, Role: auth.RoleSoporte}
	if _, err := servicio.UpdateTags(principal.Number, catalogo, deSoporte); err != nil {
		t.Fatalf("Soporte debería poder etiquetar el principal: %v", err)
	}
	if enElPrincipal := etiquetasDe(t, servicio, principal.ID); !mismasEtiquetas(enElPrincipal, catalogo) {
		t.Fatalf("Soporte dejó %v y debería dejar %v", enElPrincipal, catalogo)
	}

	// Y el texto, que es lo que ya podía: el asunto y la descripción siguen siendo suyos.
	if _, err := servicio.Update(principal.Number, UpdateInput{
		Subject:     "asunto corregido",
		Description: "<p>descripción corregida</p>",
	}, deSoporte); err != nil {
		t.Fatalf("Soporte debería poder editar el texto del principal: %v", err)
	}

	// Y el solicitante, con su propio ticket.
	delSolicitante := auth.Identity{ID: solicitante.ID, Role: auth.RoleUsuario}
	if _, err := servicio.UpdateTags(principal.Number, nil, delSolicitante); err != nil {
		t.Fatalf("el solicitante debería poder cambiar las etiquetas de su ticket: %v", err)
	}
	if enElPrincipal := etiquetasDe(t, servicio, principal.ID); len(enElPrincipal) != 0 {
		t.Fatalf("quitar todas las etiquetas debería dejarlas en cero y quedaron %v", enElPrincipal)
	}
}

// TestEtiquetaQueNoEstaEnElCatalogo: el catálogo se cura (decisión 84), así que **quien etiqueta elige
// de las que hay** —Desarrollo desde el interno igual que Soporte— y una etiqueta que no existe
// responde `422` (`tickets.etiqueta.desconocida`).
func TestEtiquetaQueNoEstaEnElCatalogo(t *testing.T) {
	servicio, tx, _ := servicioDeObservadores(t)

	interno, principal := internoAbierto(t, tx)
	desarrollo1 := cuentaDeLaBase(t, tx, auth.RoleDesarrollo, true)
	soporte1 := cuentaDeLaBase(t, tx, auth.RoleSoporte, true)
	servicio.SetAccounts(cuentasStub{desarrollo1.ID: desarrollo1, soporte1.ID: soporte1})

	const inventada = "etiqueta-que-no-existe-catalina"

	var existe int64
	if err := tx.Raw(`SELECT count(*) FROM ticket_tag_names WHERE normalized = ?`, inventada).Scan(&existe).Error; err != nil {
		t.Fatalf("no se pudo comprobar el catálogo: %v", err)
	}
	if existe != 0 {
		t.Skipf("la etiqueta %q ya existe en el catálogo de desarrollo", inventada)
	}

	if _, err := servicio.UpdateTags(interno.Number, []string{inventada}, auth.Identity{ID: desarrollo1.ID, Role: auth.RoleDesarrollo}); !errors.Is(err, ErrTagDesconocida) {
		t.Fatalf("Desarrollo debería recibir %v y recibió %v", ErrTagDesconocida, err)
	}
	if _, err := servicio.UpdateTags(principal.Number, []string{inventada}, auth.Identity{ID: soporte1.ID, Role: auth.RoleSoporte}); !errors.Is(err, ErrTagDesconocida) {
		t.Fatalf("Soporte debería recibir %v y recibió %v", ErrTagDesconocida, err)
	}
}

// ============================================================================
// Utilidades
// ============================================================================

// internoAbierto devuelve un interno que no está cerrado con su principal, que tampoco lo está: es lo
// que hace falta para probar un cambio de etiquetas sin chocar con la puerta del ticket cerrado.
func internoAbierto(t *testing.T, tx *gorm.DB) (repositories.InternalTicket, repositories.Ticket) {
	t.Helper()

	var interno repositories.InternalTicket
	if err := tx.Raw(`
		SELECT i.id, i.ticket_id, i.number, i.state
		FROM internal_tickets i
		JOIN tickets t ON t.id = i.ticket_id
		WHERE i.state <> 'cerrado' AND t.state <> 'cerrado'
		ORDER BY i.id LIMIT 1`).Scan(&interno).Error; err != nil {
		t.Fatalf("no se pudo leer un interno: %v", err)
	}
	if interno.ID == 0 {
		t.Skip("no hay ningún interno abierto: aplica backend/migrations/v1.0.0_dev.sql")
	}

	return interno, principalDeID(t, tx, interno.TicketID)
}

// principalDeID lee el principal con lo que necesitan estas pruebas.
func principalDeID(t *testing.T, tx *gorm.DB, id int64) repositories.Ticket {
	t.Helper()

	var principal repositories.Ticket
	if err := tx.Raw(`SELECT id, number, requester_id, category_id, state FROM tickets WHERE id = ?`, id).
		Scan(&principal).Error; err != nil {
		t.Fatalf("no se pudo leer el principal %d: %v", id, err)
	}

	return principal
}

// etiquetasDelCatalogo trae las primeras etiquetas del catálogo, por nombre. Sin suficientes, se salta.
func etiquetasDelCatalogo(t *testing.T, tx *gorm.DB, cuantas int) []string {
	t.Helper()

	var etiquetas []string
	if err := tx.Raw(`SELECT tag FROM ticket_tag_names ORDER BY tag ASC LIMIT ?`, cuantas).
		Scan(&etiquetas).Error; err != nil {
		t.Fatalf("no se pudo leer el catálogo: %v", err)
	}
	if len(etiquetas) < cuantas {
		t.Skip("el catálogo de desarrollo no tiene etiquetas suficientes: aplica backend/migrations/v1.0.0_dev.sql")
	}

	return etiquetas
}

// etiquetasDelCatalogoQueNoLleva trae del catálogo las primeras que **ese ticket no lleva ya**, para
// que la prueba compruebe un cambio de verdad y no un no-op. Sin suficientes, se salta.
func etiquetasDelCatalogoQueNoLleva(t *testing.T, tx *gorm.DB, lleva []string, cuantas int) []string {
	t.Helper()

	var todas []string
	if err := tx.Raw(`SELECT tag FROM ticket_tag_names ORDER BY tag ASC`).Scan(&todas).Error; err != nil {
		t.Fatalf("no se pudo leer el catálogo: %v", err)
	}

	ya := map[string]bool{}
	for _, etiqueta := range lleva {
		ya[etiqueta] = true
	}

	elegidas := make([]string, 0, cuantas)
	for _, etiqueta := range todas {
		if ya[etiqueta] {
			continue
		}

		elegidas = append(elegidas, etiqueta)
		if len(elegidas) == cuantas {
			return elegidas
		}
	}

	t.Skip("el catálogo no tiene etiquetas libres suficientes: aplica backend/migrations/v1.0.0_dev.sql")

	return nil
}

// cuentaPorID lee la cuenta que hace falta —por ejemplo, el solicitante de un ticket—. Sin ella, se
// salta: `users` es de otro módulo y esta prueba se limita a traer lo que el servicio vería.
func cuentaPorID(t *testing.T, tx *gorm.DB, id int64) auth.Account {
	t.Helper()

	var cuenta auth.Account
	if err := tx.Raw(
		`SELECT id, name, last_name, email, role, is_active FROM users WHERE id = ?`, id,
	).Scan(&cuenta).Error; err != nil {
		t.Fatalf("no se pudo leer la cuenta %d: %v", id, err)
	}
	if cuenta.ID == 0 {
		t.Skipf("la cuenta %d no está en la base: aplica backend/migrations/v1.0.0_dev.sql", id)
	}

	return cuenta
}

// etiquetasDe lee las etiquetas de un ticket por su identificador.
func etiquetasDe(t *testing.T, servicio *Service, ticketID int64) []string {
	t.Helper()

	etiquetas, err := servicio.categories.TagsByTicket(ticketID)
	if err != nil {
		t.Fatalf("no se pudieron leer las etiquetas del ticket %d: %v", ticketID, err)
	}

	return etiquetas
}

// mismasEtiquetas compara dos listas sin importar el orden: `TagsByTicket` las devuelve por nombre.
func mismasEtiquetas(una, otra []string) bool {
	if len(una) != len(otra) {
		return false
	}

	vistas := map[string]int{}
	for _, etiqueta := range una {
		vistas[etiqueta]++
	}
	for _, etiqueta := range otra {
		vistas[etiqueta]--
		if vistas[etiqueta] < 0 {
			return false
		}
	}

	return true
}
