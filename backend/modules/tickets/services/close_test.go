package services

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"gorm.io/gorm"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// Estas pruebas cubren la decisión 81: **un ticket no se cierra sin decir por qué**. Primero, la
// preparación del comentario —que es donde se rechaza— sin base de datos; después, el movimiento
// entero contra la base **de verdad**, dentro de una transacción que se deshace al terminar, para
// comprobar lo que de verdad importa: que el comentario queda en la conversación, que el cierre del
// solicitante pasa por el mismo camino y que un cierre rechazado no deja el ticket a medias.

// correoDePrueba es un `Mailer` que no manda nada. Hace falta porque mover un ticket avisa por correo y
// el servicio no admite un correo nulo —a diferencia del etiquetado, que sí lo comprueba—: sin él, la
// prueba no llega ni a la transacción.
type correoDePrueba struct{}

func (correoDePrueba) SendAsync(string, string, []string, map[string]string) {}

func (correoDePrueba) Link(ruta string) string { return ruta }

// TestPrepararComentarioDelMovimiento es la regla del encargo, sin base de datos: cerrar exige un
// comentario que no sea sólo espacios, los demás estados lo admiten opcional, y un cuerpo que no pasa
// el saneado o una mención que no vale se rechazan como en un comentario normal.
func TestPrepararComentarioDelMovimiento(t *testing.T) {
	servicio := &Service{accounts: cuentasStub{
		2: {ID: 2, Role: auth.RoleDesarrollo, IsActive: true},
	}}

	mencion := `<p>Cerrado porque <span data-mencion="2">Ana</span> lo confirmó.</p>`

	casos := []struct {
		nombre     string
		quien      auth.Identity
		estado     string
		comentario string
		quiereErr  error
		quiereMsj  bool
	}{
		{"cerrar sin comentario", soporte, repositories.StateCerrado, "", ErrCierreSinComentario, false},
		{"cerrar con espacios", soporte, repositories.StateCerrado, "   \n ", ErrCierreSinComentario, false},
		{"cerrar con comentario", soporte, repositories.StateCerrado, "<p>Ya funciona.</p>", nil, true},
		{"resolver sin comentario", soporte, repositories.StateResuelto, "", nil, false},
		{"resolver con comentario", soporte, repositories.StateResuelto, "<p>Hecho.</p>", nil, true},
		// Un texto que no pasa la lista blanca se rechaza, cierre o no.
		{"comentario con HTML que no vale", soporte, repositories.StateResuelto, "<script>x</script>", ErrBodyNotAllowed, false},
		// Etiquetar sigue las mismas reglas: Soporte puede, el solicitante no estrena menciones.
		{"cierre con mención de Soporte", soporte, repositories.StateCerrado, mencion, nil, true},
		{"cierre con mención del solicitante", usuario, repositories.StateCerrado, mencion, ErrMentionNotAllowed, false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			texto, menciones, err := servicio.prepararComentarioDelMovimiento(caso.estado, caso.comentario, caso.quien)

			if !errors.Is(err, caso.quiereErr) {
				t.Fatalf("se esperaba %v y salió %v", caso.quiereErr, err)
			}
			if caso.quiereErr != nil {
				return
			}
			if caso.quiereMsj && strings.TrimSpace(texto) == "" {
				t.Fatal("se esperaba un comentario y salió vacío")
			}
			if !caso.quiereMsj && texto != "" {
				t.Fatalf("no se esperaba comentario y salió %q", texto)
			}
			if caso.nombre == "cierre con mención de Soporte" && len(menciones) != 1 {
				t.Fatalf("se esperaba una mención y salieron %v", menciones)
			}
		})
	}
}

// TestCerrarConComentarioLoGuardaEnLaConversacion es el caso central: un cierre con comentario deja el
// ticket cerrado **y el comentario en la conversación**, escrito por quien cierra y con el texto que
// mandó (decisión 81).
func TestCerrarConComentarioLoGuardaEnLaConversacion(t *testing.T) {
	servicio, tx, _ := servicioDeCierre(t)

	numero, _, _ := unPrincipalCerrable(t, tx)
	soporte1 := cuentaDeLaBase(t, tx, auth.RoleSoporte, true)
	servicio.SetAccounts(cuentasStub{soporte1.ID: soporte1})

	actor := auth.Identity{ID: soporte1.ID, Role: auth.RoleSoporte}
	cuerpo := "<p>Se cierra porque el usuario confirmó que ya le funciona.</p>"

	cerrado, err := servicio.Move(numero, repositories.StateCerrado, cuerpo, actor)
	if err != nil {
		t.Fatalf("el cierre falló: %v", err)
	}
	if cerrado.State != repositories.StateCerrado {
		t.Fatalf("el ticket quedó en %q y se esperaba cerrado", cerrado.State)
	}

	// Se lee por la API, que es lo que verá la pantalla: el comentario está en la conversación y su
	// autor es quien cerró.
	ficha, err := servicio.ByNumber(numero, actor)
	if err != nil {
		t.Fatalf("no se pudo leer la ficha: %v", err)
	}

	comentario := buscarComentario(ficha.Comments, soporte1.ID, cuerpo)
	if comentario == nil {
		t.Fatalf("el comentario del cierre no está en la conversación: %+v", ficha.Comments)
	}
	if comentario.Author == nil || comentario.Author.ID != soporte1.ID {
		t.Fatalf("el comentario quedó sin autor o con otro: %+v", comentario.Author)
	}
}

// TestCerrarComoSolicitante: **el propio solicitante cierra su ticket por el mismo camino** y con la
// misma exigencia: cierra con comentario y el comentario queda en la conversación (decisión 81).
func TestCerrarComoSolicitante(t *testing.T) {
	servicio, tx, _ := servicioDeCierre(t)

	numero, _, solicitanteID := unPrincipalCerrableDeUsuario(t, tx)
	solicitante := cuentaDeLaBasePorID(t, tx, solicitanteID)
	if solicitante.Role != auth.RoleUsuario {
		t.Skip("el ticket más antiguo cerrable no lo abrió un usuario")
	}

	servicio.SetAccounts(cuentasStub{solicitante.ID: solicitante})

	actor := auth.Identity{ID: solicitante.ID, Role: auth.RoleUsuario}
	cuerpo := "<p>Ya no me hace falta, lo cierro.</p>"

	// Y **sin comentario no cierra**, también para el solicitante: la regla es de quien cierra.
	if _, err := servicio.Move(numero, repositories.StateCerrado, "  ", actor); !errors.Is(err, ErrCierreSinComentario) {
		t.Fatalf("el solicitante debería recibir %v y recibió %v", ErrCierreSinComentario, err)
	}

	cerrado, err := servicio.Move(numero, repositories.StateCerrado, cuerpo, actor)
	if err != nil {
		t.Fatalf("el cierre del solicitante falló: %v", err)
	}
	if cerrado.State != repositories.StateCerrado {
		t.Fatalf("el ticket quedó en %q y se esperaba cerrado", cerrado.State)
	}

	ficha, err := servicio.ByNumber(numero, actor)
	if err != nil {
		t.Fatalf("no se pudo leer la ficha: %v", err)
	}
	if comentario := buscarComentario(ficha.Comments, solicitante.ID, cuerpo); comentario == nil {
		t.Fatalf("el comentario del solicitante no está en la conversación: %+v", ficha.Comments)
	}
}

// TestCerrarConMencionAvisaYObserva: un comentario de cierre con mención **se comporta como cualquier
// comentario** (decisión 81): se guarda con su etiqueta y quien se nombra pasa a observar el ticket.
// Es la decisión que se llevó al informe: se mantiene la regla del comentario normal, no se inventa una
// puerta distinta para el cierre.
func TestCerrarConMencionAvisaYObserva(t *testing.T) {
	servicio, tx, _ := servicioDeCierre(t)

	numero, principalID, _ := unPrincipalCerrable(t, tx)
	soporte1 := cuentaDeLaBase(t, tx, auth.RoleSoporte, true)
	dev1 := cuentaDeLaBase(t, tx, auth.RoleDesarrollo, true)
	servicio.SetAccounts(cuentasStub{soporte1.ID: soporte1, dev1.ID: dev1})

	// La prueba empieza de cero para ese ticket: se le quitan los observadores que trajera.
	limpiarObservadores(t, tx, "ticket_id", principalID)

	nombre := strings.TrimSpace(dev1.FullName())
	cuerpo := `<p>Cerrado porque <span data-mencion="` + strconv.FormatInt(dev1.ID, 10) + `">` + nombre + `</span> ya lo revisó.</p>`

	cerrado, err := servicio.Move(numero, repositories.StateCerrado, cuerpo, auth.Identity{ID: soporte1.ID, Role: auth.RoleSoporte})
	if err != nil {
		t.Fatalf("el cierre con mención falló: %v", err)
	}
	if cerrado.State != repositories.StateCerrado {
		t.Fatalf("el ticket quedó en %q y se esperaba cerrado", cerrado.State)
	}

	// La mención se lee en el cuerpo y quien la lleva queda como observador, como en un comentario.
	ficha, err := servicio.ByNumber(numero, auth.Identity{ID: soporte1.ID, Role: auth.RoleSoporte})
	if err != nil {
		t.Fatalf("no se pudo leer la ficha: %v", err)
	}
	if comentario := buscarComentario(ficha.Comments, soporte1.ID, cuerpo); comentario == nil {
		t.Fatalf("el comentario con la mención no está en la conversación: %+v", ficha.Comments)
	}
	if !tieneObservador(ficha, dev1.ID) {
		t.Fatalf("quien se mencionó al cerrar no quedó como observador: %+v", ficha.Observers)
	}
}

// TestResolverSinComentarioSigueIgual: en los demás estados el comentario es opcional, y sin él **nada
// cambia respecto a hoy** (decisión 81).
func TestResolverSinComentarioSigueIgual(t *testing.T) {
	servicio, tx, _ := servicioDeCierre(t)

	numero, principalID, _ := unPrincipalResoluble(t, tx)
	soporte1 := cuentaDeLaBase(t, tx, auth.RoleSoporte, true)
	servicio.SetAccounts(cuentasStub{soporte1.ID: soporte1})

	comentariosAntes := contarComentarios(t, tx, "ticket_id", principalID)

	resuelto, err := servicio.Move(numero, repositories.StateResuelto, "", auth.Identity{ID: soporte1.ID, Role: auth.RoleSoporte})
	if err != nil {
		t.Fatalf("resolver sin comentario falló: %v", err)
	}
	if resuelto.State != repositories.StateResuelto {
		t.Fatalf("el ticket quedó en %q y se esperaba resuelto", resuelto.State)
	}
	if despues := contarComentarios(t, tx, "ticket_id", principalID); despues != comentariosAntes {
		t.Fatalf("resolver sin comentario escribió un comentario: %d antes y %d después", comentariosAntes, despues)
	}
}

// TestCierreRechazadoNoDejaNadaAMedias: cerrar sin comentario se rechaza **sin tocar nada** —ni el
// estado, ni el historial, ni la conversación— (decisión 81).
func TestCierreRechazadoNoDejaNadaAMedias(t *testing.T) {
	servicio, tx, _ := servicioDeCierre(t)

	numero, principalID, _ := unPrincipalCerrable(t, tx)
	soporte1 := cuentaDeLaBase(t, tx, auth.RoleSoporte, true)
	servicio.SetAccounts(cuentasStub{soporte1.ID: soporte1})

	estadoAntes := estadoDelPrincipal(t, tx, principalID)
	comentariosAntes := contarComentarios(t, tx, "ticket_id", principalID)
	historialAntes := contarHistorialDeEstado(t, tx, "ticket_id", principalID)

	_, err := servicio.Move(numero, repositories.StateCerrado, "", auth.Identity{ID: soporte1.ID, Role: auth.RoleSoporte})
	if !errors.Is(err, ErrCierreSinComentario) {
		t.Fatalf("se esperaba %v y salió %v", ErrCierreSinComentario, err)
	}

	if estado := estadoDelPrincipal(t, tx, principalID); estado != estadoAntes {
		t.Fatalf("el estado cambió con un cierre rechazado: %q y antes %q", estado, estadoAntes)
	}
	if despues := contarComentarios(t, tx, "ticket_id", principalID); despues != comentariosAntes {
		t.Fatalf("un cierre rechazado dejó comentario: %d antes y %d después", comentariosAntes, despues)
	}
	if despues := contarHistorialDeEstado(t, tx, "ticket_id", principalID); despues != historialAntes {
		t.Fatalf("un cierre rechazado dejó historial: %d antes y %d después", historialAntes, despues)
	}
}

// ============================================================================
// Utilidades
// ============================================================================

// servicioDeCierre monta el servicio como las pruebas de observadores, pero **con correo**: mover un
// ticket avisa, y sin un `Mailer` el servicio no llega a la transacción.
func servicioDeCierre(t *testing.T) (*Service, *gorm.DB, cuentasStub) {
	t.Helper()

	servicio, tx, cuentas := servicioDeObservadores(t)
	servicio.SetMailer(correoDePrueba{})

	return servicio, tx, cuentas
}

// unPrincipalCerrable devuelve el primer ticket principal al que la tabla le permite cerrarse (sección
// 3.1): su número, su identificador y el de su solicitante.
func unPrincipalCerrable(t *testing.T, tx *gorm.DB) (string, int64, int64) {
	t.Helper()

	var fila struct {
		ID          int64
		Number      string
		RequesterID int64
	}

	err := tx.Raw(
		`SELECT id, number, requester_id FROM tickets
		 WHERE state IN ('nuevo', 'en progreso', 'en espera', 'resuelto')
		 ORDER BY id LIMIT 1`,
	).Scan(&fila).Error
	if err != nil {
		t.Fatalf("no se pudo leer un ticket cerrable: %v", err)
	}
	if fila.ID == 0 {
		t.Skip("no hay ningún ticket principal cerrable: aplica backend/migrations/v1.0.0_dev.sql")
	}

	return fila.Number, fila.ID, fila.RequesterID
}

// unPrincipalCerrableDeUsuario es lo mismo, pero **de un ticket que abrió un usuario**: es el que hace
// falta para probar que el solicitante cierra el suyo por el mismo camino.
func unPrincipalCerrableDeUsuario(t *testing.T, tx *gorm.DB) (string, int64, int64) {
	t.Helper()

	var fila struct {
		ID          int64
		Number      string
		RequesterID int64
	}

	err := tx.Raw(
		`SELECT t.id, t.number, t.requester_id FROM tickets t
		 JOIN users u ON u.id = t.requester_id
		 WHERE t.state IN ('nuevo', 'en progreso', 'en espera', 'resuelto') AND u.role = 'usuario'
		 ORDER BY t.id LIMIT 1`,
	).Scan(&fila).Error
	if err != nil {
		t.Fatalf("no se pudo leer un ticket cerrable de un usuario: %v", err)
	}
	if fila.ID == 0 {
		t.Skip("no hay ningún ticket de usuario cerrable: aplica backend/migrations/v1.0.0_dev.sql")
	}

	return fila.Number, fila.ID, fila.RequesterID
}

// unPrincipalResoluble devuelve un principal que se puede resolver sin pasar antes por otro estado.
func unPrincipalResoluble(t *testing.T, tx *gorm.DB) (string, int64, int64) {
	t.Helper()

	var fila struct {
		ID          int64
		Number      string
		RequesterID int64
	}

	err := tx.Raw(
		`SELECT id, number, requester_id FROM tickets
		 WHERE state IN ('en progreso', 'en espera')
		 ORDER BY id LIMIT 1`,
	).Scan(&fila).Error
	if err != nil {
		t.Fatalf("no se pudo leer un ticket resoluble: %v", err)
	}
	if fila.ID == 0 {
		t.Skip("no hay ningún ticket principal resoluble: aplica backend/migrations/v1.0.0_dev.sql")
	}

	return fila.Number, fila.ID, fila.RequesterID
}

// cuentaDeLaBasePorID trae una cuenta concreta, sea del papel que sea.
func cuentaDeLaBasePorID(t *testing.T, tx *gorm.DB, id int64) auth.Account {
	t.Helper()

	var cuenta auth.Account
	if err := tx.Raw(
		`SELECT id, name, last_name, email, role, is_active FROM users WHERE id = ?`, id,
	).Scan(&cuenta).Error; err != nil {
		t.Fatalf("no se pudo leer la cuenta %d: %v", id, err)
	}
	if cuenta.ID == 0 {
		t.Skipf("la cuenta %d no está: aplica backend/migrations/v1.0.0_dev.sql", id)
	}

	return cuenta
}

// buscarComentario encuentra, en la conversación leída por la API, el comentario de esa persona con
// ese texto. Devuelve nil si no está.
func buscarComentario(comentarios []Comment, authorID int64, cuerpo string) *Comment {
	for i := range comentarios {
		if comentarios[i].Author == nil || comentarios[i].Author.ID != authorID {
			continue
		}
		if comentarios[i].Body == cuerpo {
			return &comentarios[i]
		}
	}

	return nil
}

// estadoDelPrincipal lee el estado de la fila, **de la base**, no de lo que devolvió el servicio.
func estadoDelPrincipal(t *testing.T, tx *gorm.DB, id int64) string {
	t.Helper()

	var estado string
	if err := tx.Raw(`SELECT state FROM tickets WHERE id = ?`, id).Scan(&estado).Error; err != nil {
		t.Fatalf("no se pudo leer el estado: %v", err)
	}

	return estado
}

// contarComentarios cuenta los comentarios de un hilo.
func contarComentarios(t *testing.T, tx *gorm.DB, columna string, id int64) int64 {
	t.Helper()

	var total int64
	if err := tx.Raw("SELECT count(*) FROM ticket_comments WHERE "+columna+" = ?", id).Scan(&total).Error; err != nil {
		t.Fatalf("no se pudieron contar los comentarios: %v", err)
	}

	return total
}

// contarHistorialDeEstado cuenta las entradas de cambio de estado de un hilo.
func contarHistorialDeEstado(t *testing.T, tx *gorm.DB, columna string, id int64) int64 {
	t.Helper()

	var total int64
	consulta := "SELECT count(*) FROM ticket_history WHERE " + columna + " = ? AND event IN ('estado', 'resuelto', 'cerrado', 'reabierto')"
	if err := tx.Raw(consulta, id).Scan(&total).Error; err != nil {
		t.Fatalf("no se pudo contar el historial: %v", err)
	}

	return total
}
