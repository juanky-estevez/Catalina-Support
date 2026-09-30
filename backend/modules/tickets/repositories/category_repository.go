package repositories

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CategoryConCuenta es una categoría con **cuántos tickets la usan**: es lo que la pantalla del
// catálogo enseña al lado del nombre, y lo que deja ver que retirar una categoría no la vacía
// (docs/modules/tickets.md, decisión 66).
type CategoryConCuenta struct {
	ID            int64
	Name          string
	Normalized    string
	Active        bool
	CreatedByID   *int64
	CreatedAt     time.Time
	DeactivatedAt *time.Time
	// Tickets es cuántos lo usan. Cuenta los dos tipos de ticket: una etiqueta o una categoría del
	// principal se ven en su interno.
	Tickets int64
}

// TagConCuenta es una etiqueta con **cuántos tickets la llevan**. Es lo que se sugiere mientras se
// escribe, para que no acaben siendo cinco maneras de decir lo mismo (decisión 68).
type TagConCuenta struct {
	Tag     string
	Tickets int64
}

// CategoryRepository lee y escribe el catálogo de categorías y las etiquetas de los tickets.
//
// Las dos cosas viven en este repositorio porque son una sola función —clasificar—: el catálogo es
// cerrado y lo mantiene Soporte y el Administrador, y las etiquetas son texto libre que nace la
// primera vez que alguien las escribe.
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository construye el repositorio.
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// DB deja la conexión a quien necesita una transacción que abarca más de una tabla.
func (r *CategoryRepository) DB() *gorm.DB { return r.db }

// ListCategories devuelve el catálogo con cuántos tickets tiene cada categoría.
//
// `soloActivas` es el filtro del papel: **el alta de un ticket necesita las activas**, y las retiradas
// sólo las ven quienes mantienen el catálogo, que son Soporte y el Administrador
// (docs/modules/tickets.md, sección 5).
func (r *CategoryRepository) ListCategories(soloActivas bool) ([]CategoryConCuenta, error) {
	consulta := `
		SELECT c.id, c.name, c.normalized, c.active, c.created_by_id, c.created_at, c.deactivated_at,
		       COUNT(t.id) AS tickets
		FROM ticket_categories c
		LEFT JOIN tickets t ON t.category_id = c.id`

	if soloActivas {
		consulta += ` WHERE c.active = true`
	}

	// **Las activas primero y por nombre**: lo que se ofrece al crear un ticket va delante, y lo
	// retirado se queda abajo y a la vista, que es como se vuelve a poner.
	consulta += ` GROUP BY c.id ORDER BY c.active DESC, c.name ASC`

	var filas []CategoryConCuenta
	if err := r.db.Raw(consulta).Scan(&filas).Error; err != nil {
		return nil, err
	}

	return filas, nil
}

// CategoryWithCount devuelve una categoría con su cuenta de tickets.
func (r *CategoryRepository) CategoryWithCount(id int64) (CategoryConCuenta, error) {
	var fila CategoryConCuenta

	err := r.db.Raw(`
		SELECT c.id, c.name, c.normalized, c.active, c.created_by_id, c.created_at, c.deactivated_at,
		       COUNT(t.id) AS tickets
		FROM ticket_categories c
		LEFT JOIN tickets t ON t.category_id = c.id
		WHERE c.id = ?
		GROUP BY c.id`, id).Scan(&fila).Error
	if err != nil {
		return CategoryConCuenta{}, err
	}
	if fila.ID == 0 {
		return CategoryConCuenta{}, ErrCategoryNotFound
	}

	return fila, nil
}

// CategoryByID busca una categoría por su identificador.
func (r *CategoryRepository) CategoryByID(id int64) (TicketCategory, error) {
	var categoria TicketCategory

	err := r.db.Where("id = ?", id).First(&categoria).Error
	if err != nil {
		return TicketCategory{}, traducir(err, ErrCategoryNotFound)
	}

	return categoria, nil
}

// CategoryByNormalized busca una categoría por su nombre normalizado: es como se comprueba que
// «Red» y «red» no conviven en el mismo catálogo.
func (r *CategoryRepository) CategoryByNormalized(normalized string) (TicketCategory, error) {
	var categoria TicketCategory

	err := r.db.Where("normalized = ?", normalized).First(&categoria).Error
	if err != nil {
		return TicketCategory{}, traducir(err, ErrCategoryNotFound)
	}

	return categoria, nil
}

// CategoriesByIDs trae varias categorías de una vez: es lo que evita preguntar una a una al pintar
// una lista paginada.
func (r *CategoryRepository) CategoriesByIDs(ids []int64) (map[int64]TicketCategory, error) {
	porID := map[int64]TicketCategory{}
	if len(ids) == 0 {
		return porID, nil
	}

	var categorias []TicketCategory
	if err := r.db.Where("id IN ?", ids).Find(&categorias).Error; err != nil {
		return nil, err
	}

	for _, categoria := range categorias {
		porID[categoria.ID] = categoria
	}

	return porID, nil
}

// CreateCategory inserta una categoría.
//
// Lleva `ON CONFLICT DO NOTHING` y comprueba las filas afectadas: si el nombre normalizado ya estaba,
// no se escribe nada y se dice que está repetido, en vez de dejar que la base conteste con su error
// —que el servicio tendría que adivinar—. Es la red por debajo de la comprobación del servicio,
// porque dos peticiones a la vez también pueden chocar.
func (r *CategoryRepository) CreateCategory(categoria TicketCategory) (TicketCategory, error) {
	resultado := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&categoria)
	if resultado.Error != nil {
		return TicketCategory{}, resultado.Error
	}
	if resultado.RowsAffected == 0 {
		return TicketCategory{}, ErrCategoryDuplicate
	}

	return categoria, nil
}

// UpdateCategory guarda el nombre, el estado o la marca de retirada de una categoría.
func (r *CategoryRepository) UpdateCategory(categoria TicketCategory) error {
	return r.db.Save(&categoria).Error
}

// CountActiveCategories cuenta cuántas siguen activas: es lo que impide retirar la última, porque un
// catálogo sin ninguna categoría activa dejaría el alta de tickets sin nada que ofrecer.
func (r *CategoryRepository) CountActiveCategories() (int64, error) {
	var total int64
	if err := r.db.Model(&TicketCategory{}).Where("active = true").Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

// ListTags devuelve **el catálogo de etiquetas con cuántos tickets lleva cada una**.
//
// `prefijo` vacío son **las más usadas**, que es lo que se enseña cuando nadie ha escrito nada;
// con prefijo son **las que empiezan por lo que se escribe**, para sugerirlas mientras se teclea. El
// orden es por uso y, a igualdad, por nombre, para que la lista no baile.
//
// **Salen también las que no lleva nadie, con `tickets: 0`** (docs/modules/tickets.md, decisión 72):
// una etiqueta creada a mano tiene que poder verse para volver a usarla, y es el `LEFT JOIN` quien las
// trae. El prefijo se compara contra el nombre normalizado, que es como llega: ya en minúsculas.
func (r *CategoryRepository) ListTags(prefijo string, limite int) ([]TagConCuenta, error) {
	consulta := `
		SELECT n.tag AS tag, COUNT(DISTINCT tt.ticket_id) AS tickets
		FROM ticket_tag_names n
		LEFT JOIN ticket_tags tt ON tt.tag_id = n.id`

	argumentos := make([]any, 0, 2)
	if prefijo != "" {
		consulta += ` WHERE n.normalized LIKE ?`
		argumentos = append(argumentos, prefijo+"%")
	}

	consulta += ` GROUP BY n.id, n.tag ORDER BY tickets DESC, n.tag ASC LIMIT ?`
	argumentos = append(argumentos, limite)

	var filas []TagConCuenta
	if err := r.db.Raw(consulta, argumentos...).Scan(&filas).Error; err != nil {
		return nil, err
	}

	return filas, nil
}

// ListAllTags devuelve **el catálogo de etiquetas entero, sin tope y por nombre**, con cuántos tickets
// lleva cada una.
//
// Es el listado de mantenimiento de la pantalla «Categorías y etiquetas», y por eso **no es el mismo
// que `ListTags`**: las sugerencias son un desplegable de diez por uso, y por uso una etiqueta recién
// creada que no lleva ningún ticket se queda fuera y no se puede mantener (docs/modules/tickets.md,
// decisión 72). El orden es por nombre para que el listado no baile entre peticiones.
func (r *CategoryRepository) ListAllTags() ([]TagConCuenta, error) {
	var filas []TagConCuenta
	if err := r.db.Raw(`
		SELECT n.tag AS tag, COUNT(DISTINCT tt.ticket_id) AS tickets
		FROM ticket_tag_names n
		LEFT JOIN ticket_tags tt ON tt.tag_id = n.id
		GROUP BY n.id, n.tag
		ORDER BY n.tag ASC`).Scan(&filas).Error; err != nil {
		return nil, err
	}

	return filas, nil
}

// TagsByTicket devuelve las etiquetas de un ticket, ordenadas. El nombre se lee del catálogo, que es
// donde vive (docs/modules/tickets.md, decisión 72).
func (r *CategoryRepository) TagsByTicket(ticketID int64) ([]string, error) {
	var etiquetas []string
	if err := r.db.Raw(`
		SELECT n.tag
		FROM ticket_tags tt
		JOIN ticket_tag_names n ON n.id = tt.tag_id
		WHERE tt.ticket_id = ?
		ORDER BY n.tag ASC`, ticketID).Scan(&etiquetas).Error; err != nil {
		return nil, err
	}

	return etiquetas, nil
}

// TagsByTickets trae las etiquetas de varios tickets de una vez: es lo que evita preguntar uno a uno
// al pintar una lista paginada.
func (r *CategoryRepository) TagsByTickets(ticketIDs []int64) (map[int64][]string, error) {
	porTicket := map[int64][]string{}
	if len(ticketIDs) == 0 {
		return porTicket, nil
	}

	// La pareja se lee con una forma propia porque aquí no hay una tabla que la represente: es la
	// unión de las dos.
	var filas []struct {
		TicketID int64
		Tag      string
	}
	if err := r.db.Raw(`
		SELECT tt.ticket_id, n.tag
		FROM ticket_tags tt
		JOIN ticket_tag_names n ON n.id = tt.tag_id
		WHERE tt.ticket_id IN ?
		ORDER BY n.tag ASC`, ticketIDs).Scan(&filas).Error; err != nil {
		return nil, err
	}

	for _, fila := range filas {
		porTicket[fila.TicketID] = append(porTicket[fila.TicketID], fila.Tag)
	}

	return porTicket, nil
}

// ReplaceTags deja en un ticket **exactamente** las etiquetas que se piden: borra las que hubiera y
// escribe las nuevas, todo dentro de la transacción que guarda el ticket.
//
// Se reemplaza en bloque y no se suman unas a otras a propósito: quitar una etiqueta del ticket se
// hace quitándola de la lista, igual que se hace con el texto. **Quitar una etiqueta de un ticket
// borra su fila, no la del catálogo**: la etiqueta sigue existiendo para volver a usarla, que es lo
// que hace que se pueda mantener (docs/modules/tickets.md, decisión 72).
//
// **Cada nombre se asegura antes en el catálogo**: escribir una etiqueta nueva en un ticket la crea,
// que es lo que la hacía ligera. `ON CONFLICT DO NOTHING` no pisa el nombre ni el autor de la que ya
// estaba.
func (r *CategoryRepository) ReplaceTags(tx *gorm.DB, ticketID int64, tags []string, createdByID int64) error {
	if err := tx.Where("ticket_id = ?", ticketID).Delete(&TicketTag{}).Error; err != nil {
		return err
	}

	if len(tags) == 0 {
		return nil
	}

	for _, tag := range tags {
		nueva := TicketTagName{Tag: tag, Normalized: tag, CreatedByID: &createdByID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&nueva).Error; err != nil {
			return err
		}
	}

	// Y se leen sus claves de una vez, que es como se enganchan las filas del ticket.
	var nombres []TicketTagName
	if err := tx.Where("normalized IN ?", tags).Find(&nombres).Error; err != nil {
		return err
	}

	porNormalizado := make(map[string]int64, len(nombres))
	for _, nombre := range nombres {
		porNormalizado[nombre.Normalized] = nombre.ID
	}

	filas := make([]TicketTag, 0, len(tags))
	for _, tag := range tags {
		id, hay := porNormalizado[tag]
		if !hay {
			continue
		}

		filas = append(filas, TicketTag{TicketID: ticketID, TagID: id, CreatedByID: createdByID})
	}

	if len(filas) == 0 {
		return nil
	}

	return tx.Create(&filas).Error
}

// TagNameByNormalized busca una etiqueta del catálogo por su nombre normalizado: es como se comprueba
// que `red-wifi` y `Red-Wifi` no conviven en el mismo catálogo.
func (r *CategoryRepository) TagNameByNormalized(normalized string) (TicketTagName, error) {
	var nombre TicketTagName

	err := r.db.Where("normalized = ?", normalized).First(&nombre).Error
	if err != nil {
		return TicketTagName{}, traducir(err, ErrTagNotFound)
	}

	return nombre, nil
}

// CreateTagName da de alta una etiqueta en el catálogo, **aunque no la lleve ningún ticket**.
//
// Lleva `ON CONFLICT DO NOTHING` y comprueba las filas afectadas, como las categorías: si el nombre
// normalizado ya estaba, no se escribe nada y se dice que está repetido, en vez de dejar que la base
// conteste con su error. Es la red por debajo de la comprobación del servicio.
func (r *CategoryRepository) CreateTagName(nombre TicketTagName) (TicketTagName, error) {
	resultado := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&nombre)
	if resultado.Error != nil {
		return TicketTagName{}, resultado.Error
	}
	if resultado.RowsAffected == 0 {
		return TicketTagName{}, ErrTagDuplicate
	}

	return nombre, nil
}

// UpdateTagName guarda el nombre de una etiqueta del catálogo. **El cambio vale para todos los
// tickets que la llevan**, porque ellos sólo guardan su clave ajena (docs/modules/tickets.md,
// decisión 72): renombrar es cambiar esta fila y nada más.
func (r *CategoryRepository) UpdateTagName(nombre TicketTagName) error {
	return r.db.Save(&nombre).Error
}

// DeleteTagName retira una etiqueta del catálogo. Las filas de `ticket_tags` que la usaban se van con
// ella por el `ON DELETE CASCADE` de la clave ajena: **quitarla de todos sus tickets es el mismo
// movimiento**, y por eso la decisión 72 la hace del Administrador.
func (r *CategoryRepository) DeleteTagName(id int64) error {
	return r.db.Delete(&TicketTagName{}, id).Error
}

// CountTicketsWithTag cuenta cuántos tickets llevan una etiqueta: es el número que se enseña al lado
// del nombre y lo que dice a cuántos afecta retirarla (docs/modules/tickets.md, decisión 72).
func (r *CategoryRepository) CountTicketsWithTag(id int64) (int64, error) {
	var total int64
	if err := r.db.Model(&TicketTag{}).Where("tag_id = ?", id).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}
