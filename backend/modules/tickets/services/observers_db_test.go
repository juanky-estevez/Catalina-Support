package services

import (
	"errors"
	"testing"

	"gorm.io/gorm"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/config"
	"catalina-support/backend/shared/database"
)

// Estas pruebas van **contra la base de datos de verdad**, porque añadir un observador escribe dos
// tablas —`ticket_observers` y `ticket_history`— y lo que hay que comprobar es justo eso: que la
// segunda vez no deja fila ni historial nuevos, y que el interno guarda los suyos aparte. Todo se
// hace **dentro de una transacción que se deshace al terminar**, así que no queda nada en la base, y
// sin base de datos a mano las pruebas **se saltan**: `go test ./...` tiene que seguir sirviendo para
// lo que no necesita contenedores.

// TestAnadirObservador es la prueba del encargo: se añade a un desarrollador a un principal, queda en
// la lista, y **añadirlo otra vez no es un error ni deja rastro de más** —ni fila duplicada ni una
// segunda entrada en el historial— (docs/modules/tickets.md, secciones 2.3.1 y 5, decisión 74).
func TestAnadirObservador(t *testing.T) {
	servicio, tx, _ := servicioDeObservadores(t)

	principal := primerNumeroDePrincipal(t, tx)
	desarrollo1 := cuentaDeLaBase(t, tx, auth.RoleDesarrollo, true)
	soporte1 := cuentaDeLaBase(t, tx, auth.RoleSoporte, true)

	// El servicio sólo ve cuentas por su interfaz: el doble lleva las dos que hacen falta.
	servicio.SetAccounts(cuentasStub{desarrollo1.ID: desarrollo1, soporte1.ID: soporte1})

	// La prueba empieza de cero para ese ticket: se le quitan los observadores que trajera.
	limpiarObservadores(t, tx, "ticket_id", idDelPrincipal(t, tx, principal))

	actor := auth.Identity{ID: soporte1.ID, Role: auth.RoleSoporte}

	detalle, err := servicio.AddObserver(principal, desarrollo1.ID, actor)
	if err != nil {
		t.Fatalf("AddObserver falló: %v", err)
	}
	if detalle.Ticket.Number != principal || detalle.Ticket.Internal {
		t.Fatalf("la ficha devuelta no es el principal %q: %+v", principal, detalle.Ticket)
	}
	if !tieneObservador(detalle, desarrollo1.ID) {
		t.Fatalf("el desarrollador %d no aparece en los observadores: %+v", desarrollo1.ID, detalle.Observers)
	}

	// **Otra vez**: contesta igual, la lista queda igual y no se apunta nada de más.
	repetido, err := servicio.AddObserver(principal, desarrollo1.ID, actor)
	if err != nil {
		t.Fatalf("añadir a quien ya observa no debería fallar: %v", err)
	}
	if !tieneObservador(repetido, desarrollo1.ID) {
		t.Fatal("la lista perdió al observador al añadirlo otra vez")
	}
	if filas := contarObservadores(t, tx, "ticket_id", idDelPrincipal(t, tx, principal)); filas != 1 {
		t.Fatalf("se esperaba una sola fila de observador y hay %d", filas)
	}
	if entradas := contarHistorial(t, tx, "ticket_id", idDelPrincipal(t, tx, principal)); entradas != 1 {
		t.Fatalf("se esperaba una sola entrada de historial y hay %d", entradas)
	}
}

// TestAnadirObservadorPermisos: añadir es de **Soporte y Desarrollo**, como quitar; el usuario y el
// Administrador reciben `tickets.forbidden` (docs/modules/tickets.md, decisión 63).
//
// Estas dos se comprueban sin base de datos: el permiso se decide **antes** de tocar el ticket, así
// que no hay nada que leer.
func TestAnadirObservadorPermisos(t *testing.T) {
	servicio := &Service{}

	if _, err := servicio.AddObserver("CS-2026-0001", 2, usuario); !errors.Is(err, ErrForbidden) {
		t.Fatalf("el usuario debería recibir %v y recibió %v", ErrForbidden, err)
	}
	if _, err := servicio.AddObserver("CS-2026-0001", 2, administrador); !errors.Is(err, ErrForbidden) {
		t.Fatalf("el administrador debería recibir %v y recibió %v", ErrForbidden, err)
	}
}

// TestAnadirObservadorQueNoVale: la cuenta que se añade tiene que ser un **técnico o un desarrollador
// activo**, que es la regla del etiquetado (decisión 59). Una cuenta de usuario, una desactivada y
// una que no existe se rechazan con la **misma clave** (`tickets.mention.notAllowed`), porque es la
// misma regla.
func TestAnadirObservadorQueNoVale(t *testing.T) {
	servicio, tx, _ := servicioDeObservadores(t)

	principal := primerNumeroDePrincipal(t, tx)
	soporte1 := cuentaDeLaBase(t, tx, auth.RoleSoporte, true)
	unUsuario := cuentaDeLaBase(t, tx, auth.RoleUsuario, true)
	unDesactivado := cuentaDeLaBase(t, tx, auth.RoleDesarrollo, true)

	// Se desactiva dentro de la transacción —y se deshace al terminar—: así la prueba distingue «no
	// vale por el papel» de «no vale porque está apagada».
	if err := tx.Exec(`UPDATE users SET is_active = false WHERE id = ?`, unDesactivado.ID).Error; err != nil {
		t.Fatalf("no se pudo desactivar la cuenta de prueba: %v", err)
	}
	unDesactivado.IsActive = false

	// El doble de cuentas devuelve lo que hay en la base, ya desactivada la tercera.
	servicio.SetAccounts(cuentasStub{
		soporte1.ID:      soporte1,
		unUsuario.ID:     unUsuario,
		unDesactivado.ID: unDesactivado,
	})

	actor := auth.Identity{ID: soporte1.ID, Role: auth.RoleSoporte}

	casos := []struct {
		nombre    string
		accountID int64
	}{
		{"una cuenta de usuario", unUsuario.ID},
		{"una cuenta desactivada", unDesactivado.ID},
		{"una cuenta que no existe", 999999999},
	}

	for _, caso := range casos {
		if _, err := servicio.AddObserver(principal, caso.accountID, actor); !errors.Is(err, ErrMentionNotAllowed) {
			t.Fatalf("%s debería rechazarse con %v y salió %v", caso.nombre, ErrMentionNotAllowed, err)
		}
	}
}

// TestAnadirObservadorAUnInterno: el ticket puede ser un **interno**, y sus observadores son los del
// interno —no los del principal— (docs/modules/tickets.md, sección 2.3.1). Lo puede hacer también
// Desarrollo, que es el otro papel que observa.
func TestAnadirObservadorAUnInterno(t *testing.T) {
	servicio, tx, _ := servicioDeObservadores(t)

	interno, principal := unInterno(t, tx)
	dev1 := cuentaDeLaBase(t, tx, auth.RoleDesarrollo, true)
	dev2 := cuentaDeLaBase(t, tx, auth.RoleDesarrollo, true)
	soporte1 := cuentaDeLaBase(t, tx, auth.RoleSoporte, true)

	servicio.SetAccounts(cuentasStub{dev2.ID: dev2, soporte1.ID: soporte1})

	limpiarObservadores(t, tx, "internal_ticket_id", interno.ID)
	antesDelPrincipal := contarObservadores(t, tx, "ticket_id", principal.ID)

	detalle, err := servicio.AddObserver(interno.Number, dev2.ID, auth.Identity{ID: dev1.ID, Role: auth.RoleDesarrollo})
	if err != nil {
		t.Fatalf("AddObserver sobre un interno falló: %v", err)
	}
	if !detalle.Ticket.Internal || detalle.Ticket.Number != interno.Number {
		t.Fatalf("la ficha devuelta no es el interno %q: %+v", interno.Number, detalle.Ticket)
	}
	if !tieneObservador(detalle, dev2.ID) {
		t.Fatalf("el observador %d no aparece en la ficha del interno: %+v", dev2.ID, detalle.Observers)
	}

	// La fila es del interno, no del principal, y la lista del principal no se ha tocado.
	var delInterno int64
	if err := tx.Raw(`SELECT count(*) FROM ticket_observers WHERE internal_ticket_id = ? AND account_id = ?`,
		interno.ID, dev2.ID).Scan(&delInterno).Error; err != nil {
		t.Fatalf("no se pudo leer la fila del interno: %v", err)
	}
	if delInterno != 1 {
		t.Fatalf("se esperaba una fila del observador en el interno y hay %d", delInterno)
	}
	if despues := contarObservadores(t, tx, "ticket_id", principal.ID); despues != antesDelPrincipal {
		t.Fatalf("los observadores del principal cambiaron: %d antes y %d después", antesDelPrincipal, despues)
	}
}

// ============================================================================
// Utilidades
// ============================================================================

// servicioDeObservadores monta el servicio sobre una transacción de la base **de verdad** y deja un
// doble de cuentas vacío, que cada prueba rellena con lo que necesita. Sin base de datos a mano, se
// salta.
func servicioDeObservadores(t *testing.T) (*Service, *gorm.DB, cuentasStub) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Skip("sin configuración de base de datos: " + err.Error())
	}

	db, err := database.Connect(cfg)
	if err != nil {
		t.Skip("la base de datos no responde: " + err.Error())
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Skip("no se pudo abrir la transacción: " + tx.Error.Error())
	}

	if !tx.Migrator().HasTable(&repositories.TicketObserver{}) ||
		!tx.Migrator().HasTable(&repositories.HistoryEntry{}) {
		_ = tx.Rollback()
		t.Skip("faltan ticket_observers o ticket_history: aplica backend/migrations/v1.0.0.sql")
	}

	t.Cleanup(func() {
		_ = tx.Rollback()

		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	cuentas := cuentasStub{}
	servicio := NewService(
		repositories.NewTicketRepository(tx),
		repositories.NewConversationRepository(tx),
		repositories.NewCategoryRepository(tx),
		"",
	)
	servicio.SetAccounts(cuentas)

	return servicio, tx, cuentas
}

// cuentaDeLaBase lee una cuenta real con ese papel y ese estado. `users` es de otro módulo y no se
// importa: la prueba se limita a traer los datos que el servicio vería por su interfaz. Sin ninguna
// cuenta que cumpla, se salta.
func cuentaDeLaBase(t *testing.T, tx *gorm.DB, rol string, activa bool) auth.Account {
	t.Helper()

	var cuenta auth.Account
	if err := tx.Raw(
		`SELECT id, name, last_name, email, role, is_active FROM users
		 WHERE role = ? AND is_active = ? ORDER BY id LIMIT 1`, rol, activa,
	).Scan(&cuenta).Error; err != nil {
		t.Fatalf("no se pudo leer una cuenta %s: %v", rol, err)
	}
	if cuenta.ID == 0 {
		t.Skipf("no hay ninguna cuenta %s activa=%v: aplica backend/migrations/v1.0.0_dev.sql", rol, activa)
	}

	return cuenta
}

// primerNumeroDePrincipal devuelve el número del primer ticket principal de la base.
func primerNumeroDePrincipal(t *testing.T, tx *gorm.DB) string {
	t.Helper()

	var numero string
	if err := tx.Raw(`SELECT number FROM tickets ORDER BY id LIMIT 1`).Scan(&numero).Error; err != nil {
		t.Fatalf("no se pudo leer un ticket: %v", err)
	}
	if numero == "" {
		t.Skip("no hay ningún ticket: aplica backend/migrations/v1.0.0_dev.sql")
	}

	return numero
}

// idDelPrincipal traduce el número a su identificador, que es lo que usan las tablas que cuelgan.
func idDelPrincipal(t *testing.T, tx *gorm.DB, numero string) int64 {
	t.Helper()

	var id int64
	if err := tx.Raw(`SELECT id FROM tickets WHERE number = ?`, numero).Scan(&id).Error; err != nil {
		t.Fatalf("no se pudo leer el identificador de %q: %v", numero, err)
	}

	return id
}

// unInterno devuelve un interno de la base con su principal. Sin ninguno, se salta.
func unInterno(t *testing.T, tx *gorm.DB) (repositories.InternalTicket, repositories.Ticket) {
	t.Helper()

	var interno repositories.InternalTicket
	if err := tx.Raw(`SELECT id, ticket_id, number FROM internal_tickets ORDER BY id LIMIT 1`).Scan(&interno).Error; err != nil {
		t.Fatalf("no se pudo leer un interno: %v", err)
	}
	if interno.ID == 0 {
		t.Skip("no hay ningún ticket interno: aplica backend/migrations/v1.0.0_dev.sql")
	}

	var principal repositories.Ticket
	if err := tx.Raw(`SELECT id, number FROM tickets WHERE id = ?`, interno.TicketID).Scan(&principal).Error; err != nil {
		t.Fatalf("no se pudo leer el principal del interno: %v", err)
	}

	return interno, principal
}

// limpiarObservadores deja el hilo sin observadores dentro de la transacción, para que la prueba
// cuente desde cero. La columna es `ticket_id` o `internal_ticket_id`.
func limpiarObservadores(t *testing.T, tx *gorm.DB, columna string, id int64) {
	t.Helper()

	if err := tx.Exec("DELETE FROM ticket_observers WHERE "+columna+" = ?", id).Error; err != nil {
		t.Fatalf("no se pudieron limpiar los observadores: %v", err)
	}
}

// contarObservadores cuenta las filas de un hilo.
func contarObservadores(t *testing.T, tx *gorm.DB, columna string, id int64) int64 {
	t.Helper()

	var total int64
	if err := tx.Raw("SELECT count(*) FROM ticket_observers WHERE "+columna+" = ?", id).Scan(&total).Error; err != nil {
		t.Fatalf("no se pudieron contar los observadores: %v", err)
	}

	return total
}

// contarHistorial cuenta las entradas de «observador añadido» de un hilo: es lo que tiene que salir
// **una sola vez** aunque se añada dos veces.
func contarHistorial(t *testing.T, tx *gorm.DB, columna string, id int64) int64 {
	t.Helper()

	var total int64
	if err := tx.Raw("SELECT count(*) FROM ticket_history WHERE "+columna+" = ? AND event = 'observador_anadido'", id).
		Scan(&total).Error; err != nil {
		t.Fatalf("no se pudo contar el historial: %v", err)
	}

	return total
}

// tieneObservador dice si esa cuenta está en la lista de la ficha.
func tieneObservador(detalle Detail, accountID int64) bool {
	for _, observador := range detalle.Observers {
		if observador.Account != nil && observador.Account.ID == accountID {
			return true
		}
	}

	return false
}
