package services

import (
	"errors"
	"strings"
	"unicode"

	"gorm.io/gorm"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// MaxTagLength es el tope de una etiqueta ya normalizada.
//
// **32 caracteres**, y elegido, no heredado: una etiqueta es un matiz que se lee en un chip al lado de
// la categoría, y `licencia-office-2026` cabe de sobra. Un tope corto mantiene los chips legibles y el
// índice pequeño; sin tope, una etiqueta de mil caracteres se guarda igual y estropea la lista. La
// columna es `text`, así que el tope lo pone el servicio: es una regla de producto, no del esquema.
const MaxTagLength = 32

// MaxEtiquetasSugeridas es cuántas se ofrecen mientras se escribe. Es un desplegable, no una lista:
// con diez se elige sin buscar.
const MaxEtiquetasSugeridas = 10

// Catalogo es lo que este módulo necesita del catálogo de categorías y de las etiquetas.
//
// Se declara aquí, en quien lo usa, como `Accounts` y `Mailer`: así las pruebas del servicio pueden
// doblarlo sin base de datos y el repositorio no sabe que existe esta interfaz.
type Catalogo interface {
	ListCategories(soloActivas bool) ([]repositories.CategoryConCuenta, error)
	CategoryWithCount(id int64) (repositories.CategoryConCuenta, error)
	CategoryByID(id int64) (repositories.TicketCategory, error)
	CategoryByNormalized(normalized string) (repositories.TicketCategory, error)
	CategoriesByIDs(ids []int64) (map[int64]repositories.TicketCategory, error)
	CreateCategory(categoria repositories.TicketCategory) (repositories.TicketCategory, error)
	UpdateCategory(categoria repositories.TicketCategory) error
	CountActiveCategories() (int64, error)
	ListTags(prefijo string, limite int) ([]repositories.TagConCuenta, error)
	// ListAllTags es el catálogo entero, sin tope y por nombre: es el listado de mantenimiento, no el
	// de sugerencias (docs/modules/tickets.md, decisión 72).
	ListAllTags() ([]repositories.TagConCuenta, error)
	TagsByTicket(ticketID int64) ([]string, error)
	TagsByTickets(ticketIDs []int64) (map[int64][]string, error)
	ReplaceTags(tx *gorm.DB, ticketID int64, tags []string, createdByID int64) error
	// El catálogo de etiquetas: se mantiene como el de categorías, pero una etiqueta puede existir sin
	// que ningún ticket la lleve (docs/modules/tickets.md, decisión 72).
	TagNameByNormalized(normalized string) (repositories.TicketTagName, error)
	CreateTagName(nombre repositories.TicketTagName) (repositories.TicketTagName, error)
	UpdateTagName(nombre repositories.TicketTagName) error
	DeleteTagName(id int64) error
	CountTicketsWithTag(id int64) (int64, error)
}

// Category es la categoría tal y como viaja dentro de un ticket.
type Category struct {
	ID     int64
	Name   string
	Active bool
}

// CategoryWithCount es una categoría del catálogo con cuántos tickets la usan.
type CategoryWithCount struct {
	Category
	Tickets int64
}

// TagCount es una etiqueta con cuántos tickets la llevan.
type TagCount struct {
	Tag     string
	Tickets int64
}

// acentos quita los acentos y la eñe. Se hace con una tabla y no con normalización Unicode —que
// pediría una dependencia más para esto— porque los acentos que llegan aquí son los del español, que
// son los que escribe quien rellena un ticket.
var acentos = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
)

// NormalizarCategoria deja un nombre de categoría en lo que se compara: minúsculas, sin acentos y sin
// espacios de sobra. **Es lo que impide «Red» y «red» en el mismo catálogo** (decisión 65).
//
// No se tocan los espacios de dentro a propósito: el nombre se lee tal cual se escribió, y «Software
// interno» y «Software  interno» no son la misma cosa escrita de dos maneras, son dos nombres.
func NormalizarCategoria(nombre string) string {
	return acentos.Replace(strings.ToLower(strings.TrimSpace(nombre)))
}

// NormalizarEtiqueta deja una etiqueta como se guarda (decisión 68): minúsculas, sin acentos, los
// espacios vueltos guiones y fuera todo lo que no sea letra, número o guion.
//
// **Una etiqueta que se queda vacía se rechaza** con `tickets.tag.required`: sólo signos, o espacios,
// no es una etiqueta, y guardar la cadena vacía sería guardar un hueco que nadie puede leer. Los
// guiones repetidos se juntan en uno y los de los extremos se caen, para que `Red  wifi` y `red-wifi`
// sean la misma etiqueta.
func NormalizarEtiqueta(bruta string) (string, error) {
	texto := acentos.Replace(strings.ToLower(strings.TrimSpace(bruta)))

	var (
		constructor    strings.Builder
		guionPendiente bool
	)

	for _, letra := range texto {
		switch {
		case letra >= 'a' && letra <= 'z', letra >= '0' && letra <= '9':
			if guionPendiente && constructor.Len() > 0 {
				constructor.WriteByte('-')
			}
			guionPendiente = false
			constructor.WriteRune(letra)

		case letra == '-' || unicode.IsSpace(letra):
			guionPendiente = true

		default:
			// Todo lo demás se cae: ni puntos, ni comas, ni barras, ni signos de interrogación.
		}
	}

	normalizada := constructor.String()
	if normalizada == "" {
		return "", ErrTagRequired
	}
	if len([]rune(normalizada)) > MaxTagLength {
		return "", ErrTagTooLong
	}

	return normalizada, nil
}

// NormalizarEtiquetas normaliza una lista entera, **sin repetidas** y en el orden en que se escriben:
// el orden es el de la pantalla, y la misma etiqueta dos veces es una vez.
func NormalizarEtiquetas(brutas []string) ([]string, error) {
	normalizadas := make([]string, 0, len(brutas))
	vistas := map[string]bool{}

	for _, bruta := range brutas {
		normalizada, err := NormalizarEtiqueta(bruta)
		if err != nil {
			return nil, err
		}
		if vistas[normalizada] {
			continue
		}

		vistas[normalizada] = true
		normalizadas = append(normalizadas, normalizada)
	}

	return normalizadas, nil
}

// creadorDelCatalogo es el autor que se guarda al crear una categoría o una etiqueta, o **nada** si
// quien crea es la cuenta de fábrica.
//
// **La cuenta de fábrica no está en `users`: lo que se guarda es que no la creó nadie, y no un cero
// que no existe.** Las dos columnas del catálogo (`ticket_categories.created_by_id` y
// `ticket_tag_names.created_by_id`) son anulables justo para esto; guardar el cero de la cuenta de
// fábrica revienta la clave ajena contra `users` y deja al Administrador sin poder crear nada
// (docs/modules/tickets.md, decisiones 64, 72 y 73).
func creadorDelCatalogo(actor auth.Identity) *int64 {
	if actor.Factory || actor.ID == 0 {
		return nil
	}

	return &actor.ID
}

// puedeMantenerElCatalogo: **el catálogo lo mantiene sólo el Administrador** —crear, renombrar y
// retirar, categorías y etiquetas— (docs/modules/tickets.md, decisión 83, que corrige la 73). **Soporte
// y Desarrollo lo usan** —clasifican tickets y filtran con él— pero **no lo cambian**, y Desarrollo
// tampoco lo mantiene.
//
// De aquí sale también **qué se enseña en el catálogo**: quien no lo mantiene **sólo ve las categorías y
// las etiquetas activas** (`soloActivas`), que es lo coherente con no poder cambiarlas.
func puedeMantenerElCatalogo(actor auth.Identity) bool {
	return actor.Role == auth.RoleAdministrador
}

// Categories devuelve el catálogo con su cuenta de tickets.
//
// **Lo ve cualquiera que haya entrado** —el alta de un ticket necesita las categorías—, pero lo
// retirado sólo lo ven quienes lo mantienen: Soporte y el Administrador. Así la pantalla de alta no
// ofrece una categoría que ya no se puede elegir.
func (s *Service) Categories(actor auth.Identity) ([]CategoryWithCount, error) {
	soloActivas := !puedeMantenerElCatalogo(actor)

	filas, err := s.categories.ListCategories(soloActivas)
	if err != nil {
		return nil, err
	}

	categorias := make([]CategoryWithCount, 0, len(filas))
	for _, fila := range filas {
		categorias = append(categorias, aCategoryWithCount(fila))
	}

	return categorias, nil
}

// CreateCategory da de alta una categoría nueva: Soporte y el Administrador.
func (s *Service) CreateCategory(nombre string, actor auth.Identity) (CategoryWithCount, error) {
	if !puedeMantenerElCatalogo(actor) {
		return CategoryWithCount{}, ErrCategoryForbidden
	}

	limpio := strings.TrimSpace(nombre)
	if limpio == "" {
		return CategoryWithCount{}, ErrCategoryNameRequired
	}

	normalizado := NormalizarCategoria(limpio)

	// Se comprueba antes de escribir para poder decir **«ya existe»** y no el error de la base; la
	// restricción única se queda detrás, porque dos peticiones a la vez también pueden chocar.
	_, err := s.categories.CategoryByNormalized(normalizado)
	if err == nil {
		return CategoryWithCount{}, ErrCategoryDuplicate
	}
	if !errors.Is(err, repositories.ErrCategoryNotFound) {
		return CategoryWithCount{}, err
	}

	creada, err := s.categories.CreateCategory(repositories.TicketCategory{
		Name:        limpio,
		Normalized:  normalizado,
		Active:      true,
		CreatedByID: creadorDelCatalogo(actor),
	})
	if errors.Is(err, repositories.ErrCategoryDuplicate) {
		return CategoryWithCount{}, ErrCategoryDuplicate
	}
	if err != nil {
		return CategoryWithCount{}, err
	}

	return CategoryWithCount{Category: Category{
		ID:     creada.ID,
		Name:   creada.Name,
		Active: creada.Active,
	}}, nil
}

// RenameCategory cambia el nombre de una categoría. La comparación para el choque es **por el nombre
// normalizado**: renombrar «red» a «Red» es renombrarse a sí misma, y a «RED» también.
func (s *Service) RenameCategory(id int64, nombre string, actor auth.Identity) (CategoryWithCount, error) {
	if !puedeMantenerElCatalogo(actor) {
		return CategoryWithCount{}, ErrCategoryForbidden
	}

	limpio := strings.TrimSpace(nombre)
	if limpio == "" {
		return CategoryWithCount{}, ErrCategoryNameRequired
	}

	normalizado := NormalizarCategoria(limpio)

	existente, err := s.categories.CategoryByNormalized(normalizado)
	if err == nil && existente.ID != id {
		return CategoryWithCount{}, ErrCategoryDuplicate
	}
	if err != nil && !errors.Is(err, repositories.ErrCategoryNotFound) {
		return CategoryWithCount{}, err
	}

	categoria, err := s.categories.CategoryByID(id)
	if err != nil {
		return CategoryWithCount{}, traducirCategoria(err)
	}

	categoria.Name = limpio
	categoria.Normalized = normalizado
	if err := s.categories.UpdateCategory(categoria); err != nil {
		return CategoryWithCount{}, err
	}

	return s.categoriaConCuenta(id)
}

// CategoryState retira o vuelve a poner una categoría. **Retirar es desactivarla**: deja de ofrecerse
// al crear un ticket y los tickets que la tienen la conservan (decisión 66).
//
// Es **sólo del Administrador**, y **no se puede retirar la última activa**: un catálogo sin ninguna
// categoría activa dejaría el alta de tickets sin nada que ofrecer, y la categoría es obligatoria.
func (s *Service) CategoryState(id int64, active bool, actor auth.Identity) (CategoryWithCount, error) {
	if actor.Role != auth.RoleAdministrador {
		return CategoryWithCount{}, ErrCategoryForbidden
	}

	categoria, err := s.categories.CategoryByID(id)
	if err != nil {
		return CategoryWithCount{}, traducirCategoria(err)
	}

	if categoria.Active != active {
		if active {
			categoria.DeactivatedAt = nil
		} else {
			total, err := s.categories.CountActiveCategories()
			if err != nil {
				return CategoryWithCount{}, err
			}
			if total <= 1 {
				return CategoryWithCount{}, ErrCategoryLastActive
			}

			ahora := s.now()
			categoria.DeactivatedAt = &ahora
		}

		categoria.Active = active
		if err := s.categories.UpdateCategory(categoria); err != nil {
			return CategoryWithCount{}, err
		}
	}

	return s.categoriaConCuenta(id)
}

// Tags devuelve las etiquetas del catálogo. Hay dos preguntas distintas y por eso dos caminos:
//
// **Sin `catalogoCompleto` son las sugerencias**: como mucho diez, por uso, para ofrecerlas mientras se
// escribe. Es un desplegable, no una lista, y con diez se elige sin buscar.
//
// **Con `catalogoCompleto` es el listado de mantenimiento** (la pantalla «Categorías y etiquetas»): el
// catálogo entero, **sin tope y por nombre**, porque una etiqueta recién creada que no lleva ningún
// ticket tiene que poder verse para mantenerla, y por uso se quedaría fuera de las diez primeras
// (docs/modules/tickets.md, decisión 72). Salen también las que no lleva nadie, con `tickets: 0`.
func (s *Service) Tags(q string, catalogoCompleto bool) ([]TagCount, error) {
	filas, err := s.filasDeEtiquetas(q, catalogoCompleto)
	if err != nil {
		return nil, err
	}

	etiquetas := make([]TagCount, 0, len(filas))
	for _, fila := range filas {
		etiquetas = append(etiquetas, TagCount{Tag: fila.Tag, Tickets: fila.Tickets})
	}

	return etiquetas, nil
}

// filasDeEtiquetas elige de dónde salen las etiquetas: el catálogo entero o las diez sugeridas.
func (s *Service) filasDeEtiquetas(q string, catalogoCompleto bool) ([]repositories.TagConCuenta, error) {
	// El listado completo no filtra por lo que se escribe: es el catálogo, no una búsqueda.
	if catalogoCompleto {
		return s.categories.ListAllTags()
	}

	var prefijo string

	if strings.TrimSpace(q) != "" {
		normalizada, err := NormalizarEtiqueta(q)
		if err != nil {
			// Lo que se escribe no puede ser el principio de ninguna etiqueta guardada —se normaliza
			// igual al guardar—, así que no hay sugerencias. No es un error de quien escribe.
			return []repositories.TagConCuenta{}, nil
		}

		prefijo = normalizada
	}

	return s.categories.ListTags(prefijo, MaxEtiquetasSugeridas)
}

// CreateTag da de alta una etiqueta en el catálogo, **aunque no la lleve ningún ticket**
// (docs/modules/tickets.md, decisión 72).
//
// La crean Soporte y el Administrador, como las categorías; Desarrollo tampoco, porque el catálogo lo
// mantiene quien mantiene el de categorías (decisión 73). El nombre se normaliza, y de ahí salen las
// dos claves que ya existían: una que queda vacía es `tickets.tag.required`, y una que pasa del tope,
// `tickets.tag.tooLong`.
func (s *Service) CreateTag(nombre string, actor auth.Identity) (TagCount, error) {
	if !puedeMantenerElCatalogo(actor) {
		return TagCount{}, ErrTagForbidden
	}

	normalizada, err := NormalizarEtiqueta(nombre)
	if err != nil {
		return TagCount{}, err
	}

	// Se comprueba antes de escribir para poder decir **«ya existe»** y no el error de la base; la
	// restricción única del catálogo se queda detrás, porque dos peticiones a la vez también chocan.
	_, err = s.categories.TagNameByNormalized(normalizada)
	if err == nil {
		return TagCount{}, ErrTagDuplicate
	}
	if !errors.Is(err, repositories.ErrTagNotFound) {
		return TagCount{}, err
	}

	creada, err := s.categories.CreateTagName(repositories.TicketTagName{
		Tag:         normalizada,
		Normalized:  normalizada,
		CreatedByID: creadorDelCatalogo(actor),
	})
	if errors.Is(err, repositories.ErrTagDuplicate) {
		return TagCount{}, ErrTagDuplicate
	}
	if err != nil {
		return TagCount{}, err
	}

	// Nace sin tickets: puede que nadie la use todavía.
	return TagCount{Tag: creada.Tag, Tickets: 0}, nil
}

// RenameTag cambia el nombre de una etiqueta **en el catálogo**, y el cambio vale para **todos los
// tickets que la llevan** (docs/modules/tickets.md, decisión 72): ellos sólo guardan la clave ajena,
// así que renombrar es cambiar esta fila y nada más.
//
// La etiqueta de la ruta se normaliza antes de buscarla, así que `Red-Wifi` encuentra `red-wifi`. Si
// el nombre nuevo normaliza al mismo —`red-wifi` → `Red-Wifi`— **vale y no es un duplicado**: es
// renombrarse a sí misma, y la comparación es por el nombre normalizado. Si choca con otra etiqueta,
// `tickets.tag.duplicate`.
func (s *Service) RenameTag(ruta string, nuevo string, actor auth.Identity) (TagCount, error) {
	if !puedeMantenerElCatalogo(actor) {
		return TagCount{}, ErrTagForbidden
	}

	actual, err := s.etiquetaDeLaRuta(ruta)
	if err != nil {
		return TagCount{}, err
	}

	normalizada, err := NormalizarEtiqueta(nuevo)
	if err != nil {
		return TagCount{}, err
	}

	// Sólo se comprueba el choque si de verdad cambia la etiqueta: renombrarse a sí misma, aunque
	// cambien las mayúsculas, no es un duplicado.
	if normalizada != actual.Normalized {
		existente, err := s.categories.TagNameByNormalized(normalizada)
		if err == nil && existente.ID != actual.ID {
			return TagCount{}, ErrTagDuplicate
		}
		if err != nil && !errors.Is(err, repositories.ErrTagNotFound) {
			return TagCount{}, err
		}
	}

	actual.Tag = normalizada
	actual.Normalized = normalizada
	if err := s.categories.UpdateTagName(actual); err != nil {
		return TagCount{}, err
	}

	total, err := s.categories.CountTicketsWithTag(actual.ID)
	if err != nil {
		return TagCount{}, err
	}

	return TagCount{Tag: actual.Tag, Tickets: total}, nil
}

// DeleteTag retira una etiqueta del catálogo **y la quita de todos sus tickets** (docs/modules/
// tickets.md, decisión 72). Es **sólo del Administrador** (decisión 73), porque toca todos los
// tickets que la llevan.
func (s *Service) DeleteTag(ruta string, actor auth.Identity) error {
	if actor.Role != auth.RoleAdministrador {
		return ErrTagForbidden
	}

	etiqueta, err := s.etiquetaDeLaRuta(ruta)
	if err != nil {
		return err
	}

	// Las filas de los tickets se van con ella por el `ON DELETE CASCADE` de la clave ajena: no hay
	// que borrarlas una a una ni puede quedar una etiqueta apuntando a algo que ya no existe.
	return s.categories.DeleteTagName(etiqueta.ID)
}

// etiquetaDeLaRuta busca la etiqueta que trae la ruta, **normalizándola antes**: el nombre viaja como
// se escribe y lo que se guarda es su forma normalizada. Una que no normaliza a nada no existe.
func (s *Service) etiquetaDeLaRuta(ruta string) (repositories.TicketTagName, error) {
	normalizada, err := NormalizarEtiqueta(ruta)
	if err != nil {
		return repositories.TicketTagName{}, ErrTagNotFound
	}

	etiqueta, err := s.categories.TagNameByNormalized(normalizada)
	if errors.Is(err, repositories.ErrTagNotFound) {
		return repositories.TicketTagName{}, ErrTagNotFound
	}
	if err != nil {
		return repositories.TicketTagName{}, err
	}

	return etiqueta, nil
}

// categoriaParaTicket valida la categoría que llega al crear o editar un ticket: **obligatoria**, del
// catálogo y activa. Las tres claves son distintas a propósito, porque son tres cosas distintas: no
// la has puesto, no existe, o está retirada (decisiones 65 y 66).
func (s *Service) categoriaParaTicket(id int64) (repositories.TicketCategory, error) {
	if id <= 0 {
		return repositories.TicketCategory{}, ErrCategoryRequired
	}

	categoria, err := s.categories.CategoryByID(id)
	if err != nil {
		return repositories.TicketCategory{}, traducirCategoria(err)
	}
	if !categoria.Active {
		return repositories.TicketCategory{}, ErrCategoryInactive
	}

	return categoria, nil
}

// categoriaConCuenta lee una categoría con su cuenta de tickets después de cambiarla.
func (s *Service) categoriaConCuenta(id int64) (CategoryWithCount, error) {
	fila, err := s.categories.CategoryWithCount(id)
	if err != nil {
		return CategoryWithCount{}, traducirCategoria(err)
	}

	return aCategoryWithCount(fila), nil
}

// aCategoryWithCount traduce la fila del repositorio a lo que viaja por la API.
func aCategoryWithCount(fila repositories.CategoryConCuenta) CategoryWithCount {
	return CategoryWithCount{
		Category: Category{ID: fila.ID, Name: fila.Name, Active: fila.Active},
		Tickets:  fila.Tickets,
	}
}

// aCategoria traduce una categoría del repositorio a la que viaja dentro de un ticket.
func aCategoria(categoria repositories.TicketCategory) *Category {
	return &Category{ID: categoria.ID, Name: categoria.Name, Active: categoria.Active}
}

// categoriasDe lee varias categorías de una vez y las deja listas para pegar en una lista de tickets.
// Es lo que evita preguntar una a una al pintar una página.
func (s *Service) categoriasDe(ids []int64) (map[int64]*Category, error) {
	filas, err := s.categories.CategoriesByIDs(ids)
	if err != nil {
		return nil, err
	}

	porID := make(map[int64]*Category, len(filas))
	for id, fila := range filas {
		porID[id] = aCategoria(fila)
	}

	return porID, nil
}

// traducirCategoria deja el error del repositorio en el del servicio, que es el que viaja a la API.
func traducirCategoria(err error) error {
	if errors.Is(err, repositories.ErrCategoryNotFound) {
		return ErrCategoryNotFound
	}

	return err
}
