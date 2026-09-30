package repositories

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"catalina-support/backend/shared/config"
	"catalina-support/backend/shared/database"
)

// Las pruebas del repositorio del catálogo van **contra la base de datos de verdad**, porque lo que
// hay que comprobar aquí es el SQL: que el nombre normalizado es único, que las etiquetas de un
// ticket se reemplazan enteras y que los filtros por categoría, etiqueta y búsqueda encuentran lo que
// tienen que encontrar.
//
// Todo se hace **dentro de una transacción que se deshace al terminar**, así que no queda nada en la
// base, y sin base de datos a mano las pruebas **se saltan**: `go test ./...` tiene que seguir
// sirviendo para lo que no necesita contenedores.

func TestCatalogoDeCategorias(t *testing.T) {
	categorias, _, tx := repositorioDePrueba(t)

	antes, err := categorias.CountActiveCategories()
	if err != nil {
		t.Fatalf("CountActiveCategories falló: %v", err)
	}

	creada, err := categorias.CreateCategory(TicketCategory{
		Name: "Prueba Categoría", Normalized: "prueba-categoria", Active: true,
	})
	if err != nil {
		t.Fatalf("CreateCategory falló: %v", err)
	}
	if creada.ID == 0 {
		t.Fatal("la categoría creada no tiene identificador")
	}

	// El nombre normalizado es único: la misma categoría otra vez no entra.
	if _, err := categorias.CreateCategory(TicketCategory{
		Name: "PRUEBA CATEGORIA", Normalized: "prueba-categoria", Active: true,
	}); !errors.Is(err, ErrCategoryDuplicate) {
		t.Fatalf("repetir el nombre normalizado debería salir %v y salió %v", ErrCategoryDuplicate, err)
	}

	// Y se puede leer por él.
	porNombre, err := categorias.CategoryByNormalized("prueba-categoria")
	if err != nil {
		t.Fatalf("CategoryByNormalized falló: %v", err)
	}
	if porNombre.ID != creada.ID {
		t.Fatalf("se leyó la categoría %d y se esperaba la %d", porNombre.ID, creada.ID)
	}

	// El catálogo la enseña con su cuenta de tickets, que de momento es cero.
	lista, err := categorias.ListCategories(false)
	if err != nil {
		t.Fatalf("ListCategories falló: %v", err)
	}
	encontrada := false
	for _, fila := range lista {
		if fila.ID == creada.ID {
			encontrada = true
			if fila.Tickets != 0 {
				t.Fatalf("una categoría recién creada tiene %d tickets", fila.Tickets)
			}
			if fila.Name != "Prueba Categoría" || !fila.Active {
				t.Fatalf("la categoría salió mal: %+v", fila)
			}
		}
	}
	if !encontrada {
		t.Fatal("la categoría creada no aparece en el catálogo")
	}

	// El filtro de «sólo activas» no la deja fuera mientras esté activa.
	activas, err := categorias.ListCategories(true)
	if err != nil {
		t.Fatalf("ListCategories(soloActivas) falló: %v", err)
	}
	for _, fila := range activas {
		if fila.ID == creada.ID && !fila.Active {
			t.Fatal("el filtro de activas devolvió una retirada")
		}
	}

	// Retirarla es desactivarla, y cuenta una activa menos.
	creada.Active = false
	if err := categorias.UpdateCategory(creada); err != nil {
		t.Fatalf("UpdateCategory falló: %v", err)
	}

	despues, err := categorias.CountActiveCategories()
	if err != nil {
		t.Fatalf("CountActiveCategories falló: %v", err)
	}
	if despues != antes {
		t.Fatalf("se esperaban %d activas después de retirar la creada y hay %d", antes, despues)
	}

	// Y con su cuenta sigue apareciendo: retirar no la borra (decisión 66).
	conCuenta, err := categorias.CategoryWithCount(creada.ID)
	if err != nil {
		t.Fatalf("CategoryWithCount falló: %v", err)
	}
	if conCuenta.Active {
		t.Fatal("la categoría tenía que quedar retirada")
	}

	// El guardado se apoya en la transacción: nada de esto queda al terminar.
	if err := tx.Rollback().Error; err != nil {
		t.Fatalf("no se pudo deshacer la transacción: %v", err)
	}
}

func TestEtiquetasDeUnTicket(t *testing.T) {
	categorias, _, tx := repositorioDePrueba(t)

	usuario := primerUsuario(t, tx)
	ticket := primerTicket(t, tx)

	// Un prefijo propio de esta ejecución: la base es la de desarrollo y el catálogo de etiquetas es
	// compartido, así que un nombre fijo se encontraría con lo que dejaran otras pruebas.
	prefijo := "prueba-" + strconv.FormatInt(time.Now().UnixNano(), 36) + "-"
	uno := prefijo + "uno"
	dos := prefijo + "dos"

	if err := categorias.ReplaceTags(tx, ticket, []string{dos, uno}, usuario); err != nil {
		t.Fatalf("ReplaceTags falló: %v", err)
	}

	etiquetas, err := categorias.TagsByTicket(ticket)
	if err != nil {
		t.Fatalf("TagsByTicket falló: %v", err)
	}
	if len(etiquetas) != 2 || etiquetas[0] != dos || etiquetas[1] != uno {
		t.Fatalf("las etiquetas salieron %v", etiquetas)
	}

	// La lista para sugerir: con prefijo, las que empiezan por él; sin prefijo, las más usadas.
	sugeridas, err := categorias.ListTags(prefijo, 10)
	if err != nil {
		t.Fatalf("ListTags falló: %v", err)
	}
	if len(sugeridas) != 2 {
		t.Fatalf("se esperaban 2 sugerencias y salieron %v", sugeridas)
	}
	for _, sugerida := range sugeridas {
		if sugerida.Tickets != 1 {
			t.Fatalf("la etiqueta %s dice tener %d tickets", sugerida.Tag, sugerida.Tickets)
		}
	}

	// Y en bloque, que es como se pinta una lista.
	porTicket, err := categorias.TagsByTickets([]int64{ticket})
	if err != nil {
		t.Fatalf("TagsByTickets falló: %v", err)
	}
	if len(porTicket[ticket]) != 2 {
		t.Fatalf("el mapa de etiquetas salió %v", porTicket)
	}

	// Reemplazar deja **exactamente** las que se piden: la que no está, se va del ticket —pero su
	// etiqueta sigue en el catálogo, para volver a usarla—.
	if err := categorias.ReplaceTags(tx, ticket, []string{uno}, usuario); err != nil {
		t.Fatalf("ReplaceTags falló: %v", err)
	}

	etiquetas, err = categorias.TagsByTicket(ticket)
	if err != nil {
		t.Fatalf("TagsByTicket falló: %v", err)
	}
	if len(etiquetas) != 1 || etiquetas[0] != uno {
		t.Fatalf("después de reemplazar quedaron %v", etiquetas)
	}

	// La que se quitó del ticket sigue en el catálogo, con cero tickets.
	sugeridas, err = categorias.ListTags(dos, 10)
	if err != nil {
		t.Fatalf("ListTags falló: %v", err)
	}
	if len(sugeridas) != 1 || sugeridas[0].Tickets != 0 {
		t.Fatalf("la etiqueta quitada del ticket tenía que seguir en el catálogo con 0 tickets: %v", sugeridas)
	}

	// Y vaciar la lista las quita todas.
	if err := categorias.ReplaceTags(tx, ticket, nil, usuario); err != nil {
		t.Fatalf("ReplaceTags con la lista vacía falló: %v", err)
	}
	etiquetas, err = categorias.TagsByTicket(ticket)
	if err != nil {
		t.Fatalf("TagsByTicket falló: %v", err)
	}
	if len(etiquetas) != 0 {
		t.Fatalf("quedaron etiquetas: %v", etiquetas)
	}
}

// TestCatalogoDeEtiquetas es el corazón de la decisión 72: una etiqueta puede existir **sin que ningún
// ticket la lleve**, renombrarla alcanza a **todos** los tickets que la llevan y retirarla la quita de
// ellos. Todo eso sale de una sola fuente de verdad: el nombre vive en el catálogo.
func TestCatalogoDeEtiquetas(t *testing.T) {
	categorias, _, tx := repositorioDePrueba(t)

	usuario := primerUsuario(t, tx)
	ticket := primerTicket(t, tx)
	otro := otroTicket(t, tx, ticket)

	// Un nombre propio de esta ejecución, por lo mismo que en la prueba de arriba.
	nombre := "prueba-catalogo-" + strconv.FormatInt(time.Now().UnixNano(), 36)

	creada, err := categorias.CreateTagName(TicketTagName{
		Tag: nombre, Normalized: nombre, CreatedByID: &usuario,
	})
	if err != nil {
		t.Fatalf("CreateTagName falló: %v", err)
	}
	if creada.ID == 0 {
		t.Fatal("la etiqueta creada no tiene identificador")
	}

	// El nombre normalizado es único: la misma etiqueta otra vez no entra.
	if _, err := categorias.CreateTagName(TicketTagName{Tag: "OTRA", Normalized: nombre}); !errors.Is(err, ErrTagDuplicate) {
		t.Fatalf("repetir el nombre normalizado debería salir %v y salió %v", ErrTagDuplicate, err)
	}

	// Y sale en el catálogo con **cero tickets**: es lo que permite volver a usarla.
	filas, err := categorias.ListTags(nombre, 10)
	if err != nil {
		t.Fatalf("ListTags falló: %v", err)
	}
	if len(filas) != 1 || filas[0].Tickets != 0 || filas[0].Tag != nombre {
		t.Fatalf("una etiqueta recién creada tenía que salir sola y con 0 tickets: %v", filas)
	}

	// Ponerla en dos tickets la cuenta dos veces.
	for _, id := range []int64{ticket, otro} {
		if err := categorias.ReplaceTags(tx, id, []string{nombre}, usuario); err != nil {
			t.Fatalf("ReplaceTags falló: %v", err)
		}
	}

	total, err := categorias.CountTicketsWithTag(creada.ID)
	if err != nil {
		t.Fatalf("CountTicketsWithTag falló: %v", err)
	}
	if total != 2 {
		t.Fatalf("la etiqueta dice tener %d tickets y son 2", total)
	}

	// **Renombrarla alcanza a todos los tickets que la llevan**, porque ellos sólo guardan su clave.
	nuevo := nombre + "-nuevo"
	creada.Tag = nuevo
	creada.Normalized = nuevo
	if err := categorias.UpdateTagName(creada); err != nil {
		t.Fatalf("UpdateTagName falló: %v", err)
	}

	for _, id := range []int64{ticket, otro} {
		etiquetas, err := categorias.TagsByTicket(id)
		if err != nil {
			t.Fatalf("TagsByTicket falló: %v", err)
		}
		if len(etiquetas) != 1 || etiquetas[0] != nuevo {
			t.Fatalf("el ticket %d tendría que llevar %q y lleva %v", id, nuevo, etiquetas)
		}
	}

	// **Retirarla la quita de sus tickets**, por el borrado en cascada de la clave ajena.
	if err := categorias.DeleteTagName(creada.ID); err != nil {
		t.Fatalf("DeleteTagName falló: %v", err)
	}

	for _, id := range []int64{ticket, otro} {
		etiquetas, err := categorias.TagsByTicket(id)
		if err != nil {
			t.Fatalf("TagsByTicket falló: %v", err)
		}
		if len(etiquetas) != 0 {
			t.Fatalf("el ticket %d tendría que quedarse sin la etiqueta y lleva %v", id, etiquetas)
		}
	}

	// El guardado se apoya en la transacción: nada de esto queda al terminar.
	if err := tx.Rollback().Error; err != nil {
		t.Fatalf("no se pudo deshacer la transacción: %v", err)
	}
}

// TestFiltrosDeCategoriaYEtiqueta es el corazón de la decisión 68: los dos chips filtran y la caja de
// búsqueda encuentra por categoría y por etiqueta.
func TestFiltrosDeCategoriaYEtiqueta(t *testing.T) {
	categorias, tickets, tx := repositorioDePrueba(t)

	usuario := primerUsuario(t, tx)
	ticket := primerTicket(t, tx)

	creada, err := categorias.CreateCategory(TicketCategory{
		Name: "Prueba Filtro", Normalized: "prueba-filtro", Active: true,
	})
	if err != nil {
		t.Fatalf("CreateCategory falló: %v", err)
	}

	if err := tx.Model(&Ticket{}).Where("id = ?", ticket).Update("category_id", creada.ID).Error; err != nil {
		t.Fatalf("no se pudo poner la categoría al ticket: %v", err)
	}
	if err := categorias.ReplaceTags(tx, ticket, []string{"prueba-filtro"}, usuario); err != nil {
		t.Fatalf("ReplaceTags falló: %v", err)
	}

	// Por categoría, por su identificador.
	encontrados, _, err := tickets.List(Filtros{CategoryID: &creada.ID, Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("List por categoría falló: %v", err)
	}
	if !contieneTicket(encontrados, ticket) {
		t.Fatalf("el filtro por categoría no encontró el ticket: %v", idsDeTickets(encontrados))
	}

	// Por etiqueta.
	encontrados, _, err = tickets.List(Filtros{Tag: "prueba-filtro", Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("List por etiqueta falló: %v", err)
	}
	if !contieneTicket(encontrados, ticket) {
		t.Fatalf("el filtro por etiqueta no encontró el ticket: %v", idsDeTickets(encontrados))
	}

	// Y la búsqueda: «filtro» sale por el nombre de la categoría «Prueba Filtro» y por la etiqueta
	// `prueba-filtro`, sin mirar el asunto ni la descripción del ticket.
	encontrados, _, err = tickets.List(Filtros{Query: "filtro", Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("List por texto falló: %v", err)
	}
	if !contieneTicket(encontrados, ticket) {
		t.Fatalf("la búsqueda no encontró el ticket por su categoría ni por su etiqueta: %v", idsDeTickets(encontrados))
	}
}

// TestListadoCompletoDeEtiquetas: el catálogo entero, **sin tope y por nombre**, incluye la etiqueta
// que no lleva ningún ticket —justo la que las diez sugeridas dejan fuera— (decisión 72).
func TestListadoCompletoDeEtiquetas(t *testing.T) {
	categorias, _, tx := repositorioDePrueba(t)

	// Un prefijo propio de esta ejecución: el catálogo es compartido y un nombre fijo se encontraría
	// con lo que dejaran otras pruebas.
	prefijo := "prueba-completo-" + strconv.FormatInt(time.Now().UnixNano(), 36) + "-"

	// Doce etiquetas sin tickets: con el tope de las sugerencias, dos se quedan fuera.
	for i := 0; i < 12; i++ {
		nombre := prefijo + fmt.Sprintf("%02d", i)
		if _, err := categorias.CreateTagName(TicketTagName{Tag: nombre, Normalized: nombre}); err != nil {
			t.Fatalf("CreateTagName(%q) falló: %v", nombre, err)
		}
	}

	// Las sugerencias se quedan en diez: por eso una etiqueta sin tickets no se ve en esa lista.
	sugeridas, err := categorias.ListTags(prefijo, 10)
	if err != nil {
		t.Fatalf("ListTags falló: %v", err)
	}
	if len(sugeridas) != 10 {
		t.Fatalf("las sugerencias tenían que ser diez y salieron %d", len(sugeridas))
	}

	// Y el listado completo las trae todas, por nombre.
	todas, err := categorias.ListAllTags()
	if err != nil {
		t.Fatalf("ListAllTags falló: %v", err)
	}

	nuestras := make([]string, 0, 12)
	for _, fila := range todas {
		if !strings.HasPrefix(fila.Tag, prefijo) {
			continue
		}
		nuestras = append(nuestras, fila.Tag)
		if fila.Tickets != 0 {
			t.Fatalf("la etiqueta %s dice llevar %d tickets", fila.Tag, fila.Tickets)
		}
	}
	if len(nuestras) != 12 {
		t.Fatalf("el listado completo tenía que traer las 12 y trajo %v", nuestras)
	}
	for i := 1; i < len(nuestras); i++ {
		if nuestras[i-1] > nuestras[i] {
			t.Fatalf("el listado completo no viene por nombre: %v", nuestras)
		}
	}

	if err := tx.Rollback().Error; err != nil {
		t.Fatalf("no se pudo deshacer la transacción: %v", err)
	}
}

// ============================================================================
// Utilidades
// ============================================================================

// repositorioDePrueba abre la base del proyecto y devuelve los repositorios trabajando dentro de una
// transacción que se deshace al acabar la prueba.
func repositorioDePrueba(t *testing.T) (*CategoryRepository, *TicketRepository, *gorm.DB) {
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

	if !tx.Migrator().HasTable(&TicketCategory{}) || !tx.Migrator().HasTable(&TicketTag{}) ||
		!tx.Migrator().HasTable(&TicketTagName{}) {
		_ = tx.Rollback()
		t.Skip("faltan ticket_categories, ticket_tags o ticket_tag_names: aplica backend/migrations/v1.0.0.sql")
	}

	t.Cleanup(func() {
		_ = tx.Rollback()

		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	return NewCategoryRepository(tx), NewTicketRepository(tx), tx
}

// primerUsuario devuelve una cuenta cualquiera de la base: las etiquetas tienen autor y esta prueba no
// está para crear cuentas. Sin ninguna cuenta a mano —una base recién creada— la prueba se salta.
func primerUsuario(t *testing.T, tx *gorm.DB) int64 {
	t.Helper()

	var id int64
	if err := tx.Raw(`SELECT id FROM users ORDER BY id LIMIT 1`).Scan(&id).Error; err != nil {
		t.Fatalf("no se pudo leer una cuenta: %v", err)
	}
	if id == 0 {
		t.Skip("no hay ninguna cuenta en la base: aplica backend/migrations/v1.0.0_dev.sql")
	}

	return id
}

// primerTicket devuelve un ticket cualquiera: los filtros y las etiquetas se prueban sobre uno que ya
// exista. Sin tickets, la prueba se salta.
func primerTicket(t *testing.T, tx *gorm.DB) int64 {
	t.Helper()

	var id int64
	if err := tx.Raw(`SELECT id FROM tickets ORDER BY id LIMIT 1`).Scan(&id).Error; err != nil {
		t.Fatalf("no se pudo leer un ticket: %v", err)
	}
	if id == 0 {
		t.Skip("no hay ningún ticket en la base: aplica backend/migrations/v1.0.0_dev.sql")
	}

	return id
}

// otroTicket devuelve un ticket cualquiera que no sea el indicado: es lo que hace falta para
// comprobar que renombrar una etiqueta alcanza a **todos** los tickets, no a uno. Sin un segundo
// ticket, la prueba se salta.
func otroTicket(t *testing.T, tx *gorm.DB, distinto int64) int64 {
	t.Helper()

	var id int64
	if err := tx.Raw(`SELECT id FROM tickets WHERE id <> ? ORDER BY id LIMIT 1`, distinto).Scan(&id).Error; err != nil {
		t.Fatalf("no se pudo leer otro ticket: %v", err)
	}
	if id == 0 {
		t.Skip("no hay un segundo ticket en la base: aplica backend/migrations/v1.0.0_dev.sql")
	}

	return id
}

func contieneTicket(tickets []Ticket, id int64) bool {
	for _, ticket := range tickets {
		if ticket.ID == id {
			return true
		}
	}

	return false
}

func idsDeTickets(tickets []Ticket) []int64 {
	ids := make([]int64, 0, len(tickets))
	for _, ticket := range tickets {
		ids = append(ids, ticket.ID)
	}

	return ids
}
