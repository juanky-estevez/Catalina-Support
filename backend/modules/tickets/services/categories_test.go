package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// catalogoStub es un `Catalogo` de mentira para probar las reglas del servicio sin base de datos.
//
// Implementa lo que estas pruebas preguntan; lo demás lo hereda del `Catalogo` incrustado, que es
// nulo: si una prueba llama a algo que no está aquí, revienta y se ve que falta.
type catalogoStub struct {
	Catalogo
	porID       map[int64]repositories.TicketCategory
	activas     int64
	creadas     []repositories.TicketCategory
	actualizada *repositories.TicketCategory

	// El catálogo de etiquetas, indexado por el nombre normalizado —que es como se busca—, la cuenta
	// de tickets que lleva cada una y lo que las pruebas quieren mirar después.
	etiquetas        map[string]repositories.TicketTagName
	cuentas          map[int64]int64
	etiquetasCreadas []repositories.TicketTagName
	etiquetaEditada  *repositories.TicketTagName
	etiquetaBorrada  *int64

	// Lo que devuelven las dos listas: las sugerencias de siempre y el catálogo entero.
	sugeridas []repositories.TagConCuenta
	catalogo  []repositories.TagConCuenta
}

func (c *catalogoStub) ListTags(prefijo string, limite int) ([]repositories.TagConCuenta, error) {
	return c.sugeridas, nil
}

func (c *catalogoStub) ListAllTags() ([]repositories.TagConCuenta, error) {
	return c.catalogo, nil
}

func (c *catalogoStub) TagNameByNormalized(normalized string) (repositories.TicketTagName, error) {
	if etiqueta, hay := c.etiquetas[normalized]; hay {
		return etiqueta, nil
	}

	return repositories.TicketTagName{}, repositories.ErrTagNotFound
}

func (c *catalogoStub) CreateTagName(nombre repositories.TicketTagName) (repositories.TicketTagName, error) {
	if c.etiquetas == nil {
		c.etiquetas = map[string]repositories.TicketTagName{}
	}
	if _, hay := c.etiquetas[nombre.Normalized]; hay {
		return repositories.TicketTagName{}, repositories.ErrTagDuplicate
	}

	nombre.ID = 77
	c.etiquetas[nombre.Normalized] = nombre
	c.etiquetasCreadas = append(c.etiquetasCreadas, nombre)

	return nombre, nil
}

func (c *catalogoStub) UpdateTagName(nombre repositories.TicketTagName) error {
	c.etiquetaEditada = &nombre

	// El mapa va por normalizado, así que al renombrar cambia de clave.
	for clave, etiqueta := range c.etiquetas {
		if etiqueta.ID == nombre.ID {
			delete(c.etiquetas, clave)
		}
	}
	if c.etiquetas == nil {
		c.etiquetas = map[string]repositories.TicketTagName{}
	}
	c.etiquetas[nombre.Normalized] = nombre

	return nil
}

func (c *catalogoStub) DeleteTagName(id int64) error {
	c.etiquetaBorrada = &id

	return nil
}

func (c *catalogoStub) CountTicketsWithTag(id int64) (int64, error) { return c.cuentas[id], nil }

func (c *catalogoStub) CategoryByID(id int64) (repositories.TicketCategory, error) {
	if categoria, hay := c.porID[id]; hay {
		return categoria, nil
	}

	return repositories.TicketCategory{}, repositories.ErrCategoryNotFound
}

func (c *catalogoStub) CategoryByNormalized(normalized string) (repositories.TicketCategory, error) {
	for _, categoria := range c.porID {
		if categoria.Normalized == normalized {
			return categoria, nil
		}
	}

	return repositories.TicketCategory{}, repositories.ErrCategoryNotFound
}

func (c *catalogoStub) CountActiveCategories() (int64, error) { return c.activas, nil }

func (c *catalogoStub) UpdateCategory(categoria repositories.TicketCategory) error {
	c.actualizada = &categoria
	if c.porID == nil {
		c.porID = map[int64]repositories.TicketCategory{}
	}
	c.porID[categoria.ID] = categoria

	return nil
}

func (c *catalogoStub) CreateCategory(categoria repositories.TicketCategory) (repositories.TicketCategory, error) {
	categoria.ID = 99
	c.creadas = append(c.creadas, categoria)

	return categoria, nil
}

func (c *catalogoStub) CategoryWithCount(id int64) (repositories.CategoryConCuenta, error) {
	categoria, err := c.CategoryByID(id)
	if err != nil {
		return repositories.CategoryConCuenta{}, err
	}

	return repositories.CategoryConCuenta{
		ID:         categoria.ID,
		Name:       categoria.Name,
		Normalized: categoria.Normalized,
		Active:     categoria.Active,
	}, nil
}

// TestNormalizarEtiqueta: minúsculas, espacios por guiones, sin acentos y fuera todo lo que no sea
// letra, número o guion (docs/modules/tickets.md, decisión 68).
func TestNormalizarEtiqueta(t *testing.T) {
	casos := []struct {
		bruta  string
		quiere string
	}{
		{"Red Wifi", "red-wifi"},
		{"  Licencia Office  ", "licencia-office"},
		{"CONTRASEÑA", "contrasena"},
		{"Aplicación", "aplicacion"},
		{"red--wifi", "red-wifi"},
		{"red_wifi", "redwifi"},
		{"¿qué?", "que"},
		{"2026", "2026"},
		// El guion de dentro se queda, y el de los extremos se cae.
		{"-red-", "red"},
	}

	for _, caso := range casos {
		obtenida, err := NormalizarEtiqueta(caso.bruta)
		if err != nil {
			t.Fatalf("NormalizarEtiqueta(%q) falló: %v", caso.bruta, err)
		}
		if obtenida != caso.quiere {
			t.Fatalf("NormalizarEtiqueta(%q) = %q, se esperaba %q", caso.bruta, obtenida, caso.quiere)
		}
	}

	// **Una etiqueta que se queda vacía se rechaza**: sólo signos o espacios no es una etiqueta.
	for _, bruta := range []string{"", "   ", "!!!", "¿?", "..."} {
		if _, err := NormalizarEtiqueta(bruta); !errors.Is(err, ErrTagRequired) {
			t.Fatalf("%q debería rechazarse con %v y salió %v", bruta, ErrTagRequired, err)
		}
	}

	// Y una que pasa del tope también, en vez de guardar algo que estropea el chip:
	larga := make([]rune, MaxTagLength+1)
	for i := range larga {
		larga[i] = 'a'
	}
	if _, err := NormalizarEtiqueta(string(larga)); !errors.Is(err, ErrTagTooLong) {
		t.Fatalf("una etiqueta de más de %d caracteres debería rechazarse con %v", MaxTagLength, ErrTagTooLong)
	}
}

// TestNormalizarEtiquetas: la misma etiqueta dos veces es una vez, y se conserva el orden en que se
// escriben, que es el de la pantalla.
func TestNormalizarEtiquetas(t *testing.T) {
	obtenidas, err := NormalizarEtiquetas([]string{"Red Wifi", "red-wifi", "Licencias", "  licencias  ", "Impresoras"})
	if err != nil {
		t.Fatalf("NormalizarEtiquetas falló: %v", err)
	}

	quiere := []string{"red-wifi", "licencias", "impresoras"}
	if len(obtenidas) != len(quiere) {
		t.Fatalf("se esperaban %v y salieron %v", quiere, obtenidas)
	}
	for i := range quiere {
		if obtenidas[i] != quiere[i] {
			t.Fatalf("se esperaban %v y salieron %v", quiere, obtenidas)
		}
	}

	if _, err := NormalizarEtiquetas([]string{"red", "!!!"}); !errors.Is(err, ErrTagRequired) {
		t.Fatalf("una etiqueta vacía en la lista debería rechazar la lista entera, y salió %v", err)
	}
}

// TestNormalizarCategoria: el nombre normalizado es minúsculas y sin acentos, y es lo que impide
// «Red» y «red» en el mismo catálogo (decisión 65).
func TestNormalizarCategoria(t *testing.T) {
	casos := map[string]string{
		"Red":              "red",
		"  Software  ":     "software",
		"LICENCIAS":        "licencias",
		"Aplicación":       "aplicacion",
		"Impresoras":       "impresoras",
		"Software interno": "software interno",
	}

	for nombre, quiere := range casos {
		if obtenido := NormalizarCategoria(nombre); obtenido != quiere {
			t.Fatalf("NormalizarCategoria(%q) = %q, se esperaba %q", nombre, obtenido, quiere)
		}
	}
}

// TestCategoriaDelTicket: las tres claves de la categoría de un ticket —no la has puesto, no existe o
// está retirada—, que son tres cosas distintas (decisiones 65 y 66).
func TestCategoriaDelTicket(t *testing.T) {
	servicio := &Service{categories: &catalogoStub{porID: map[int64]repositories.TicketCategory{
		1: {ID: 1, Name: "Red", Normalized: "red", Active: true},
		2: {ID: 2, Name: "Retirada", Normalized: "retirada", Active: false},
	}}}

	if _, err := servicio.categoriaParaTicket(0); !errors.Is(err, ErrCategoryRequired) {
		t.Fatalf("sin categoría debería salir %v y salió %v", ErrCategoryRequired, err)
	}
	if _, err := servicio.categoriaParaTicket(99); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("una categoría que no existe debería salir %v y salió %v", ErrCategoryNotFound, err)
	}
	if _, err := servicio.categoriaParaTicket(2); !errors.Is(err, ErrCategoryInactive) {
		t.Fatalf("una categoría retirada debería salir %v y salió %v", ErrCategoryInactive, err)
	}

	categoria, err := servicio.categoriaParaTicket(1)
	if err != nil || categoria.ID != 1 {
		t.Fatalf("una categoría activa debería valer, y salió %v (%v)", categoria, err)
	}
}

// TestRetirarLaUltimaActiva: el catálogo no se puede quedar sin ninguna categoría activa, porque el
// alta no tendría qué ofrecer (decisión 65). Y retirar es **sólo del Administrador** (decisión 64).
func TestRetirarLaUltimaActiva(t *testing.T) {
	nueva := func(activas int64) *catalogoStub {
		return &catalogoStub{
			porID: map[int64]repositories.TicketCategory{
				1: {ID: 1, Name: "General", Normalized: "general", Active: true},
				2: {ID: 2, Name: "Red", Normalized: "red", Active: true},
			},
			activas: activas,
		}
	}

	// Con dos activas, retirar una se puede.
	servicio := &Service{categories: nueva(2), now: time.Now}
	categoria, err := servicio.CategoryState(2, false, administrador)
	if err != nil {
		t.Fatalf("con dos activas debería poder retirarse una, y salió %v", err)
	}
	if categoria.Active {
		t.Fatal("la categoría tenía que quedar retirada")
	}

	// Con una sola, no: es la última que le queda al alta.
	servicio = &Service{categories: nueva(1), now: time.Now}
	if _, err := servicio.CategoryState(1, false, administrador); !errors.Is(err, ErrCategoryLastActive) {
		t.Fatalf("retirar la última activa debería salir %v y salió %v", ErrCategoryLastActive, err)
	}

	// Reactivar no mira el número de activas.
	if _, err := servicio.CategoryState(2, true, administrador); err != nil {
		t.Fatalf("reactivar no debería fallar, y salió %v", err)
	}

	// Y quien no es Administrador no retira ni reactiva.
	for _, quien := range []auth.Identity{soporte, desarrollo, usuario} {
		if _, err := servicio.CategoryState(1, false, quien); !errors.Is(err, ErrCategoryForbidden) {
			t.Fatalf("el papel %s no puede retirar, y salió %v", quien.Role, err)
		}
	}
}

// TestPermisosDelCatalogo: **el catálogo lo mantiene sólo el Administrador** (decisión 83, que corrige
// la 73): el usuario, Soporte y Desarrollo no lo tocan, y el Administrador sí.
func TestPermisosDelCatalogo(t *testing.T) {
	servicio := &Service{categories: &catalogoStub{porID: map[int64]repositories.TicketCategory{
		1: {ID: 1, Name: "Red", Normalized: "red", Active: true},
	}}}

	for _, quien := range []auth.Identity{usuario, soporte, desarrollo} {
		if _, err := servicio.CreateCategory("Nueva", quien); !errors.Is(err, ErrCategoryForbidden) {
			t.Fatalf("el papel %s no puede crear categorías, y salió %v", quien.Role, err)
		}
		if _, err := servicio.RenameCategory(1, "Otra", quien); !errors.Is(err, ErrCategoryForbidden) {
			t.Fatalf("el papel %s no puede renombrar, y salió %v", quien.Role, err)
		}
	}

	// El Administrador sí. **Soporte no**: lo usa, no lo cambia.
	for _, quien := range []auth.Identity{administrador} {
		if _, err := servicio.CreateCategory("Nueva", quien); err != nil {
			t.Fatalf("el papel %s debería poder crear, y salió %v", quien.Role, err)
		}
	}

	// El nombre vacío se rechaza antes de escribir nada.
	if _, err := servicio.CreateCategory("   ", administrador); !errors.Is(err, ErrCategoryNameRequired) {
		t.Fatalf("un nombre vacío debería salir %v y salió %v", ErrCategoryNameRequired, err)
	}
}

// TestRenombrarCategoriaRepetida: renombrar a un nombre que ya existe se rechaza comparando el
// **normalizado**, así que «RED» choca con «red»; renombrarse a sí misma, no.
func TestRenombrarCategoriaRepetida(t *testing.T) {
	servicio := &Service{categories: &catalogoStub{porID: map[int64]repositories.TicketCategory{
		1: {ID: 1, Name: "Red", Normalized: "red", Active: true},
		2: {ID: 2, Name: "Software", Normalized: "software", Active: true},
	}}}

	if _, err := servicio.RenameCategory(1, "SOFTWARE", administrador); !errors.Is(err, ErrCategoryDuplicate) {
		t.Fatalf("renombrar a un nombre ya existente debería salir %v y salió %v", ErrCategoryDuplicate, err)
	}

	categoria, err := servicio.RenameCategory(1, "Red Wifi", administrador)
	if err != nil {
		t.Fatalf("renombrar a un nombre libre debería valer, y salió %v", err)
	}
	if categoria.Name != "Red Wifi" {
		t.Fatalf("el nombre quedó en %q", categoria.Name)
	}

	// A sí misma, cambiando sólo las mayúsculas: no es un choque.
	if _, err := servicio.RenameCategory(1, "RED WIFI", administrador); err != nil {
		t.Fatalf("renombrarse a sí misma no debería chocar, y salió %v", err)
	}
}

// TestInternoHeredaCategoriaYEtiquetas: la categoría y las etiquetas son **del caso**, y el interno las
// hereda de su principal al leerse (docs/modules/tickets.md, decisión 67).
func TestInternoHeredaCategoriaYEtiquetas(t *testing.T) {
	principal := repositories.Ticket{ID: 10, Number: "CS-2026-0001", Subject: "Asunto", CategoryID: 5}
	interno := repositories.InternalTicket{ID: 20, TicketID: 10, Number: "INT-CS-2026-0001", State: "nuevo"}

	categoria := &Category{ID: 5, Name: "Red", Active: true}
	etiquetas := []string{"red-wifi", "manana"}

	servicio := &Service{}
	convertido := servicio.aTicketInterno(interno, principal, map[int64]auth.Account{}, categoria, etiquetas)

	if convertido.Number != "INT-CS-2026-0001" || !convertido.Internal {
		t.Fatalf("el interno se convirtió mal: %+v", convertido)
	}
	if convertido.Subject != "Asunto" {
		t.Fatal("el interno tiene que heredar el asunto de su principal")
	}
	if convertido.Category == nil || convertido.Category.ID != 5 || convertido.Category.Name != "Red" {
		t.Fatalf("el interno tiene que heredar la categoría de su principal, y salió %+v", convertido.Category)
	}
	if len(convertido.Tags) != 2 || convertido.Tags[0] != "red-wifi" || convertido.Tags[1] != "manana" {
		t.Fatalf("el interno tiene que heredar las etiquetas de su principal, y salieron %v", convertido.Tags)
	}
}

// TestCrearEtiqueta: el nombre se normaliza al guardarlo, una repetida choca por el nombre
// normalizado y una que no vale se rechaza con las claves de siempre (decisión 72).
func TestCrearEtiqueta(t *testing.T) {
	stub := &catalogoStub{}
	servicio := &Service{categories: stub}

	// Se normaliza: lo que se guarda en el catálogo es la forma con guiones y sin mayúsculas.
	creada, err := servicio.CreateTag("  Red Wifi  ", administrador)
	if err != nil {
		t.Fatalf("crear una etiqueta debería valer, y salió %v", err)
	}
	if creada.Tag != "red-wifi" || creada.Tickets != 0 {
		t.Fatalf("la etiqueta salió %+v: se esperaba red-wifi con 0 tickets", creada)
	}
	if len(stub.etiquetasCreadas) != 1 || stub.etiquetasCreadas[0].Normalized != "red-wifi" {
		t.Fatalf("el catálogo no recibió la etiqueta normalizada: %+v", stub.etiquetasCreadas)
	}

	// Repetida, aunque cambien mayúsculas y espacios: es un duplicado, y se compara por el normalizado.
	if _, err := servicio.CreateTag("RED-WIFI", administrador); !errors.Is(err, ErrTagDuplicate) {
		t.Fatalf("una etiqueta repetida debería salir %v y salió %v", ErrTagDuplicate, err)
	}

	// Una que se queda vacía al normalizar: `tickets.tag.required`.
	if _, err := servicio.CreateTag("!!!", administrador); !errors.Is(err, ErrTagRequired) {
		t.Fatalf("una etiqueta sin nada debería salir %v y salió %v", ErrTagRequired, err)
	}

	// Y una que pasa del tope: `tickets.tag.tooLong`.
	if _, err := servicio.CreateTag(strings.Repeat("a", MaxTagLength+1), administrador); !errors.Is(err, ErrTagTooLong) {
		t.Fatalf("una etiqueta demasiado larga debería salir %v y salió %v", ErrTagTooLong, err)
	}
}

// TestRenombrarEtiqueta: se cambia la fila del catálogo —y el cambio vale para todos los tickets que
// la llevan—, la ruta se normaliza antes de buscarla y renombrarse a sí misma no es un duplicado.
func TestRenombrarEtiqueta(t *testing.T) {
	// Un catálogo nuevo por escenario: renombrar mueve las claves del mapa y encadenar escenarios sobre
	// el mismo dejaría de probar lo que dice el nombre de cada uno.
	base := func() *catalogoStub {
		return &catalogoStub{
			etiquetas: map[string]repositories.TicketTagName{
				"red-wifi": {ID: 1, Tag: "red-wifi", Normalized: "red-wifi"},
				"software": {ID: 2, Tag: "software", Normalized: "software"},
			},
			cuentas: map[int64]int64{1: 3, 2: 0},
		}
	}

	// Se cambia la fila del catálogo, que es lo que alcanza a todos los tickets que la llevan.
	stub := base()
	servicio := &Service{categories: stub}
	renombrada, err := servicio.RenameTag("red-wifi", "wifi-casa", administrador)
	if err != nil {
		t.Fatalf("renombrar debería valer, y salió %v", err)
	}
	if renombrada.Tag != "wifi-casa" || renombrada.Tickets != 3 {
		t.Fatalf("la etiqueta renombrada salió %+v", renombrada)
	}
	if stub.etiquetaEditada == nil || stub.etiquetaEditada.Normalized != "wifi-casa" {
		t.Fatalf("el catálogo no guardó el nombre nuevo: %+v", stub.etiquetaEditada)
	}

	// La ruta se normaliza antes de buscarla: `Red Wifi` encuentra `red-wifi`.
	if _, err := (&Service{categories: base()}).RenameTag("Red Wifi", "wifi-casa", administrador); err != nil {
		t.Fatalf("la etiqueta de la ruta tendría que normalizarse, y salió %v", err)
	}

	// A un nombre que ya existe: choque, comparado por el normalizado.
	if _, err := (&Service{categories: base()}).RenameTag("software", "RED-WIFI", administrador); !errors.Is(err, ErrTagDuplicate) {
		t.Fatalf("renombrar a una etiqueta que ya existe debería salir %v y salió %v", ErrTagDuplicate, err)
	}

	// Renombrarse a sí misma cambiando sólo las mayúsculas: vale y no es un duplicado.
	if _, err := (&Service{categories: base()}).RenameTag("software", "SOFTWARE", administrador); err != nil {
		t.Fatalf("renombrarse a sí misma no debería chocar, y salió %v", err)
	}

	// La que no existe: `tickets.tag.notFound`.
	if _, err := (&Service{categories: base()}).RenameTag("no-existe", "otra", administrador); !errors.Is(err, ErrTagNotFound) {
		t.Fatalf("una etiqueta que no existe debería salir %v y salió %v", ErrTagNotFound, err)
	}
}

// TestRetirarEtiqueta: retirarla la quita del catálogo —y de sus tickets, por la clave ajena—, es
// **sólo del Administrador** y una que no existe es un «no existe» (decisiones 72 y 73).
func TestRetirarEtiqueta(t *testing.T) {
	stub := &catalogoStub{
		etiquetas: map[string]repositories.TicketTagName{
			"red-wifi": {ID: 1, Tag: "red-wifi", Normalized: "red-wifi"},
		},
	}
	servicio := &Service{categories: stub}

	for _, quien := range []auth.Identity{soporte, desarrollo, usuario} {
		if err := servicio.DeleteTag("red-wifi", quien); !errors.Is(err, ErrTagForbidden) {
			t.Fatalf("el papel %s no puede retirar etiquetas, y salió %v", quien.Role, err)
		}
	}
	if stub.etiquetaBorrada != nil {
		t.Fatal("un papel que no puede retirar no debería haber borrado nada")
	}

	if err := servicio.DeleteTag("no-existe", administrador); !errors.Is(err, ErrTagNotFound) {
		t.Fatalf("retirar una etiqueta que no existe debería salir %v y salió %v", ErrTagNotFound, err)
	}

	if err := servicio.DeleteTag("Red Wifi", administrador); err != nil {
		t.Fatalf("el Administrador debería poder retirarla, y salió %v", err)
	}
	if stub.etiquetaBorrada == nil || *stub.etiquetaBorrada != 1 {
		t.Fatalf("no se borró la etiqueta del catálogo: %+v", stub.etiquetaBorrada)
	}
}

// TestAutorDeFabricaEnElCatalogo: la cuenta de fábrica no está en `users`, así que el autor que se
// guarda es **nulo** y no el cero de su identificador, que no existe y reventaría la clave ajena. Y
// crear **no falla** (docs/modules/tickets.md, decisiones 64, 72 y 73).
func TestAutorDeFabricaEnElCatalogo(t *testing.T) {
	fabrica := auth.Identity{ID: 0, Factory: true, Role: auth.RoleAdministrador}

	stub := &catalogoStub{}
	servicio := &Service{categories: stub}

	if _, err := servicio.CreateTag("sin-tickets", fabrica); err != nil {
		t.Fatalf("la cuenta de fábrica debería poder crear una etiqueta, y salió %v", err)
	}
	if len(stub.etiquetasCreadas) != 1 || stub.etiquetasCreadas[0].CreatedByID != nil {
		t.Fatalf("la etiqueta de la cuenta de fábrica tenía que quedar sin autor: %+v", stub.etiquetasCreadas)
	}

	if _, err := servicio.CreateCategory("Sin autor", fabrica); err != nil {
		t.Fatalf("la cuenta de fábrica debería poder crear una categoría, y salió %v", err)
	}
	if len(stub.creadas) != 1 || stub.creadas[0].CreatedByID != nil {
		t.Fatalf("la categoría de la cuenta de fábrica tenía que quedar sin autor: %+v", stub.creadas)
	}

	// Y una cuenta que sí está en `users` sigue dejando su identificador: el arreglo no le quita el
	// autor a nadie.
	conAutor := &catalogoStub{}
	if _, err := (&Service{categories: conAutor}).CreateTag("con-autor", administrador); err != nil {
		t.Fatalf("el Administrador debería poder crear una etiqueta, y salió %v", err)
	}
	if len(conAutor.etiquetasCreadas) != 1 || conAutor.etiquetasCreadas[0].CreatedByID == nil ||
		*conAutor.etiquetasCreadas[0].CreatedByID != administrador.ID {
		t.Fatalf("la etiqueta del Administrador tenía que guardar su autor: %+v", conAutor.etiquetasCreadas)
	}
}

// TestListadoCompletoDeEtiquetas: con el catálogo completo salen **todas** —incluida la etiqueta que
// no lleva ningún ticket, que las diez sugeridas dejarían fuera— y sin él se comporta como siempre
// (docs/modules/tickets.md, decisión 72).
func TestListadoCompletoDeEtiquetas(t *testing.T) {
	servicio := &Service{categories: &catalogoStub{
		sugeridas: []repositories.TagConCuenta{{Tag: "usada", Tickets: 3}},
		catalogo: []repositories.TagConCuenta{
			{Tag: "sin-tickets", Tickets: 0},
			{Tag: "usada", Tickets: 3},
		},
	}}

	// Sin el catálogo completo: las diez sugeridas de siempre, que es el desplegable.
	sugeridas, err := servicio.Tags("", false)
	if err != nil {
		t.Fatalf("Tags sin catálogo completo falló: %v", err)
	}
	if len(sugeridas) != 1 || sugeridas[0].Tag != "usada" {
		t.Fatalf("sin catálogo completo tenían que salir sólo las sugeridas: %+v", sugeridas)
	}

	// Con el catálogo completo: todo, y la que no lleva ningún ticket está dentro.
	completo, err := servicio.Tags("", true)
	if err != nil {
		t.Fatalf("Tags con catálogo completo falló: %v", err)
	}
	if len(completo) != 2 {
		t.Fatalf("el catálogo completo tenía que traer las dos etiquetas y trajo %+v", completo)
	}

	encontrada := false
	for _, etiqueta := range completo {
		if etiqueta.Tag == "sin-tickets" {
			encontrada = true
			if etiqueta.Tickets != 0 {
				t.Fatalf("la etiqueta sin tickets dice llevar %d", etiqueta.Tickets)
			}
		}
	}
	if !encontrada {
		t.Fatalf("el listado completo no incluye la etiqueta que no lleva ningún ticket: %+v", completo)
	}
}

// TestPermisosDeLasEtiquetas: **el catálogo lo mantiene sólo el Administrador** (decisión 83, que
// corrige la 73): ni Desarrollo, ni Soporte, ni el usuario lo tocan.
func TestPermisosDeLasEtiquetas(t *testing.T) {
	nuevo := func() *catalogoStub {
		return &catalogoStub{
			etiquetas: map[string]repositories.TicketTagName{
				"red-wifi": {ID: 1, Tag: "red-wifi", Normalized: "red-wifi"},
			},
		}
	}

	for _, quien := range []auth.Identity{soporte, desarrollo, usuario} {
		servicio := &Service{categories: nuevo()}
		if _, err := servicio.CreateTag("nueva", quien); !errors.Is(err, ErrTagForbidden) {
			t.Fatalf("el papel %s no puede crear etiquetas, y salió %v", quien.Role, err)
		}
		if _, err := servicio.RenameTag("red-wifi", "otra", quien); !errors.Is(err, ErrTagForbidden) {
			t.Fatalf("el papel %s no puede renombrar etiquetas, y salió %v", quien.Role, err)
		}
	}

	for _, quien := range []auth.Identity{administrador} {
		servicio := &Service{categories: nuevo()}
		if _, err := servicio.CreateTag("nueva", quien); err != nil {
			t.Fatalf("el papel %s debería poder crear, y salió %v", quien.Role, err)
		}
		if _, err := servicio.RenameTag("red-wifi", "otra", quien); err != nil {
			t.Fatalf("el papel %s debería poder renombrar, y salió %v", quien.Role, err)
		}
	}
}
