package services

import (
	"strings"
	"testing"

	"catalina-support/backend/modules/tickets/repositories"
	"catalina-support/backend/shared/auth"
)

// Los cuatro papeles, que es lo que decide casi todo lo de abajo.
var (
	soporte       = auth.Identity{ID: 1, Role: auth.RoleSoporte}
	desarrollo    = auth.Identity{ID: 2, Role: auth.RoleDesarrollo}
	usuario       = auth.Identity{ID: 3, Role: auth.RoleUsuario}
	administrador = auth.Identity{ID: 4, Role: auth.RoleAdministrador}
)

// nuevoTicket arma un principal en el estado que se pida.
func nuevoTicket(estado string, solicitante int64) repositories.Ticket {
	return repositories.Ticket{ID: 10, Number: "CS-2026-0001", State: estado, RequesterID: solicitante}
}

// TestNumeracion comprueba el número del ticket: prefijo, año, cuatro dígitos y sin truncar.
func TestNumeracion(t *testing.T) {
	casos := []struct {
		prefix string
		year   int
		seq    int
		quiere string
	}{
		{"CS", 2026, 1, "CS-2026-0001"},
		{"CS", 2026, 42, "CS-2026-0042"},
		{"ACME", 2027, 9999, "ACME-2027-9999"},
		// A partir del 9999 el número **crece**: no se trunca ni se repite (decisión 6).
		{"CS", 2026, 10000, "CS-2026-10000"},
	}

	for _, caso := range casos {
		obtenido := repositories.ComponerNumero(caso.prefix, caso.year, caso.seq)
		if obtenido != caso.quiere {
			t.Fatalf("ComponerNumero(%q, %d, %d) = %q, se esperaba %q", caso.prefix, caso.year, caso.seq, obtenido, caso.quiere)
		}
	}

	if interno := repositories.NumeroDeInterno("ACME-2026-0042"); interno != "INT-ACME-2026-0042" {
		t.Fatalf("el número del interno es %q", interno)
	}
}

// puedeHacerlo junta las dos comprobaciones que hace `Move`: que el movimiento exista en la tabla y
// que sea de quien lo pide. Es lo que se espera de la API, así que es lo que se prueba.
func puedeHacerlo(servicio *Service, quien auth.Identity, ticket repositories.Ticket, interno bool, desde, hasta string) bool {
	return transicionExiste(interno, desde, hasta) && servicio.puedeMover(quien, ticket, interno, desde, hasta)
}

// TestTransicionesDelPrincipal recorre la tabla de docs/modules/tickets.md, sección 3.1.
func TestTransicionesDelPrincipal(t *testing.T) {
	servicio := &Service{}

	casos := []struct {
		desde    string
		hasta    string
		quien    auth.Identity
		esperado bool
	}{
		// Soporte mueve el ticket nuevo a `en progreso`.
		{repositories.StateNuevo, repositories.StateEnProgreso, soporte, true},
		// Nadie salta de `nuevo` a `resuelto`: si se resolvió, es que alguien lo trabajó.
		{repositories.StateNuevo, repositories.StateResuelto, soporte, false},
		{repositories.StateNuevo, repositories.StateCerrado, soporte, true},
		// El usuario cierra lo suyo, pero no lo mueve por dentro.
		{repositories.StateNuevo, repositories.StateCerrado, usuario, true},
		{repositories.StateNuevo, repositories.StateEnProgreso, usuario, false},
		{repositories.StateEnProgreso, repositories.StateEnEspera, soporte, true},
		{repositories.StateEnProgreso, repositories.StateEnEspera, usuario, false},
		// Se resuelve desde `en espera` sin pasar por `en progreso` (decisión 8).
		{repositories.StateEnEspera, repositories.StateResuelto, soporte, true},
		{repositories.StateEnEspera, repositories.StateEnProgreso, soporte, true},
		{repositories.StateResuelto, repositories.StateCerrado, soporte, true},
		{repositories.StateResuelto, repositories.StateCerrado, usuario, true},
		// Cerrar sin resolver, desde cualquier estado abierto (decisión 7).
		{repositories.StateEnProgreso, repositories.StateCerrado, soporte, true},
		{repositories.StateEnEspera, repositories.StateCerrado, soporte, true},
		{repositories.StateCerrado, repositories.StateEnProgreso, soporte, false},
		// Desarrollo no mueve el principal: lo lee para tener contexto (regla 2 de la matriz).
		{repositories.StateNuevo, repositories.StateEnProgreso, desarrollo, false},
		{repositories.StateNuevo, repositories.StateCerrado, desarrollo, false},
		// El Administrador mira: no mueve nada.
		{repositories.StateNuevo, repositories.StateEnProgreso, administrador, false},
		{repositories.StateNuevo, repositories.StateCerrado, administrador, false},
	}

	for _, caso := range casos {
		ticket := nuevoTicket(caso.desde, usuario.ID)
		obtenido := puedeHacerlo(servicio, caso.quien, ticket, false, caso.desde, caso.hasta)

		if obtenido != caso.esperado {
			t.Fatalf("principal %s → %s con papel %s: se esperaba %v y salió %v",
				caso.desde, caso.hasta, caso.quien.Role, caso.esperado, obtenido)
		}
	}
}

// TestTransicionesDelInterno es la tabla de la sección 3.2.
func TestTransicionesDelInterno(t *testing.T) {
	servicio := &Service{}

	casos := []struct {
		desde    string
		hasta    string
		quien    auth.Identity
		esperado bool
	}{
		{repositories.StateNuevo, repositories.StateEnProgreso, desarrollo, true},
		{repositories.StateNuevo, repositories.StateEnProgreso, soporte, false},
		{repositories.StateEnProgreso, repositories.StateEnEspera, desarrollo, true},
		{repositories.StateEnProgreso, repositories.StateEnEspera, soporte, false},
		// Soporte responde a lo que Desarrollo preguntó: `en espera → en progreso` es de los dos.
		{repositories.StateEnEspera, repositories.StateEnProgreso, soporte, true},
		{repositories.StateEnEspera, repositories.StateEnProgreso, desarrollo, true},
		{repositories.StateEnProgreso, repositories.StateResuelto, desarrollo, true},
		{repositories.StateEnEspera, repositories.StateResuelto, desarrollo, true},
		// Resolver el interno no es de Soporte: el interno lo trabaja Desarrollo.
		{repositories.StateEnProgreso, repositories.StateResuelto, soporte, false},
		// El interno se cierra a mano, y lo pueden cerrar los dos.
		{repositories.StateResuelto, repositories.StateCerrado, desarrollo, true},
		{repositories.StateResuelto, repositories.StateCerrado, soporte, true},
		{repositories.StateEnProgreso, repositories.StateCerrado, soporte, true},
		// El usuario no toca un interno: ni siquiera lo ve.
		{repositories.StateEnProgreso, repositories.StateEnProgreso, usuario, false},
		{repositories.StateEnProgreso, repositories.StateCerrado, usuario, false},
	}

	for _, caso := range casos {
		obtenido := puedeHacerlo(servicio, caso.quien, repositories.Ticket{State: caso.desde}, true, caso.desde, caso.hasta)

		if obtenido != caso.esperado {
			t.Fatalf("interno %s → %s con papel %s: se esperaba %v y salió %v",
				caso.desde, caso.hasta, caso.quien.Role, caso.esperado, obtenido)
		}
	}
}

// TestMovimientoQueNoExiste: lo que no está en la tabla no se hace, y no es lo mismo que un permiso.
func TestMovimientoQueNoExiste(t *testing.T) {
	casos := []struct {
		interno bool
		desde   string
		hasta   string
		existe  bool
	}{
		// Existen: son movimientos de la tabla, aunque no los pueda hacer cualquiera.
		{false, repositories.StateNuevo, repositories.StateEnProgreso, true},
		{false, repositories.StateEnEspera, repositories.StateResuelto, true},
		{false, repositories.StateResuelto, repositories.StateCerrado, true},
		{true, repositories.StateEnProgreso, repositories.StateEnEspera, true},
		{true, repositories.StateResuelto, repositories.StateCerrado, true},
		// No existen: nadie salta de `nuevo` a `resuelto`, ni se sale de `cerrado` sin reabrir.
		{false, repositories.StateNuevo, repositories.StateResuelto, false},
		{false, repositories.StateResuelto, repositories.StateEnProgreso, false},
		{false, repositories.StateCerrado, repositories.StateResuelto, false},
		// `escalado → en progreso` lo provoca el interno, no una mano.
		{false, repositories.StateEscalado, repositories.StateEnProgreso, false},
		// Y `cerrado → en progreso` es reabrir, que tiene su propia acción.
		{false, repositories.StateCerrado, repositories.StateEnProgreso, false},
		// Reabrir un interno es volver a escalar, y tiene su propia acción.
		{true, repositories.StateCerrado, repositories.StateEnProgreso, false},
		{true, repositories.StateNuevo, repositories.StateResuelto, false},
	}

	for _, caso := range casos {
		if obtenido := transicionExiste(caso.interno, caso.desde, caso.hasta); obtenido != caso.existe {
			t.Fatalf("transicionExiste(%v, %s → %s) = %v, se esperaba %v",
				caso.interno, caso.desde, caso.hasta, obtenido, caso.existe)
		}
	}
}

// TestQuienVeQue: el usuario sólo ve lo suyo, y **nunca** un interno.
func TestQuienVeQue(t *testing.T) {
	propio := Ticket{Number: "CS-2026-0001", Requester: &auth.Account{ID: usuario.ID}}
	ajeno := Ticket{Number: "CS-2026-0002", Requester: &auth.Account{ID: 99}}
	interno := Ticket{Number: "INT-CS-2026-0002", Internal: true, Requester: &auth.Account{ID: usuario.ID}}

	if !puedeVer(usuario, propio) {
		t.Fatal("un usuario tiene que ver su propio ticket")
	}
	if puedeVer(usuario, ajeno) {
		t.Fatal("un usuario no puede ver el ticket de otro")
	}
	if puedeVer(usuario, interno) {
		t.Fatal("un usuario no puede ver un ticket interno, ni siquiera el de su incidencia")
	}

	// Soporte, Desarrollo y Administrador ven todo, incluidos los internos.
	for _, quien := range []auth.Identity{soporte, desarrollo, administrador} {
		if !puedeVer(quien, propio) || !puedeVer(quien, ajeno) || !puedeVer(quien, interno) {
			t.Fatalf("el papel %s debería verlo todo", quien.Role)
		}
	}
}

// TestQuienComenta: Desarrollo no escribe en el principal; es Soporte quien traslada al usuario.
func TestQuienComenta(t *testing.T) {
	principal := Ticket{Number: "CS-2026-0001", Requester: &auth.Account{ID: usuario.ID}}
	ajeno := Ticket{Number: "CS-2026-0002", Requester: &auth.Account{ID: 99}}
	interno := Ticket{Number: "INT-CS-2026-0001", Internal: true}

	casos := []struct {
		quien    auth.Identity
		ticket   Ticket
		esperado bool
	}{
		{soporte, principal, true},
		{desarrollo, principal, false},
		{administrador, principal, false},
		{usuario, principal, true},
		{usuario, ajeno, false},
		{soporte, interno, true},
		{desarrollo, interno, true},
		{usuario, interno, false},
		{administrador, interno, false},
	}

	for _, caso := range casos {
		if obtenido := puedeComentar(caso.quien, caso.ticket); obtenido != caso.esperado {
			t.Fatalf("comentar %s en %s: se esperaba %v y salió %v",
				caso.quien.Role, caso.ticket.Number, caso.esperado, obtenido)
		}
	}
}

// TestQuienEdita: el asunto y la descripción, el solicitante y Soporte; el interno, nadie.
func TestQuienEdita(t *testing.T) {
	principal := Ticket{Number: "CS-2026-0001", Requester: &auth.Account{ID: usuario.ID}}

	if !puedeEditar(soporte, principal) || !puedeEditar(usuario, principal) {
		t.Fatal("el solicitante y Soporte tienen que poder editar el principal")
	}
	if puedeEditar(desarrollo, principal) || puedeEditar(administrador, principal) {
		t.Fatal("Desarrollo y el Administrador leen el principal, no lo editan")
	}
	if puedeEditar(soporte, Ticket{Internal: true}) {
		t.Fatal("el motivo del escalado no se edita")
	}
}

// TestEstadosValidos: el interno no puede estar `escalado`, porque el interno **es** la escalación.
func TestEstadosValidos(t *testing.T) {
	for _, estado := range repositories.EstadosDelPrincipal {
		if !estadoValido(estado, false) {
			t.Fatalf("el principal debería admitir %q", estado)
		}
	}

	if !estadoValido(repositories.StateEscalado, false) {
		t.Fatal("el principal sí puede estar escalado")
	}
	if estadoValido(repositories.StateEscalado, true) {
		t.Fatal("el interno no puede estar escalado")
	}
	if estadoValido("cancelado", false) {
		t.Fatal("no hay estado `cancelado`: son seis y ni uno más")
	}
}

// TestAdjuntos: la lista cerrada de extensiones y lo que se puede ver en línea.
func TestAdjuntos(t *testing.T) {
	admitidas := []string{
		"informe.pdf", "captura.PNG", "hoja.xlsx", "notas.txt", "todo.zip", "log.log",
		// **Texto y código** (2026-09-26, decisión 54): es lo que se manda cuando el caso es una
		// consulta que falla o un despliegue que no arranca.
		"consulta.sql", "datos.json", "config.xml", "docker-compose.yml", "nginx.conf",
		"script.sh", "tarea.py", "app.js", "Consulta.java", ".htaccess", "copia.bak",
		"paquete.tar", "paquete.gz", "paquete.7z",
	}
	for _, nombre := range admitidas {
		if !extensionAdmitida(nombre) {
			t.Fatalf("%q debería admitirse", nombre)
		}
	}

	// Lo que no está en la lista cerrada se rechaza, y el `html` sigue fuera: no se admite «cualquier
	// cosa» y luego se mira.
	rechazadas := []string{"binario.exe", "sin-extension", "hoja.html", "pagina.htm", "app.apk"}
	for _, nombre := range rechazadas {
		if extensionAdmitida(nombre) {
			t.Fatalf("%q no debería admitirse", nombre)
		}
	}

	// **El `svg` no entra**: puede llevar código dentro y el navegador lo ejecuta al abrirlo.
	if extensionAdmitida("dibujo.svg") {
		t.Fatal("el svg no puede admitirse como adjunto")
	}

	// **El texto y el código se guardan con su tipo** —es lo que la API dice que son— y **se descargan
	// como todo lo que no se previsualiza**: `octet-stream` y `attachment`, que es lo que impide que un
	// archivo se abra solo en el navegador. No se pintan nunca en línea.
	if tipo := tipoDeArchivo("consulta.sql"); tipo != "text/plain; charset=utf-8" {
		t.Fatalf("un .sql se guarda como texto plano, y salió %q", tipo)
	}
	if tipo, disposicion := ServirCon("consulta.sql", true); disposicion != "attachment" || tipo != "application/octet-stream" {
		t.Fatalf("un .sql se descarga como los demás, y salió %q %q", tipo, disposicion)
	}
	if _, disposicion := ServirCon("script.sh", true); disposicion != "attachment" {
		t.Fatal("un archivo de código se descarga: nunca se enseña en línea")
	}

	// Lo que se puede previsualizar se sirve en línea; lo demás, siempre como descarga.
	if tipo, disposicion := ServirCon("captura.png", true); disposicion != "inline" || tipo != "image/png" {
		t.Fatalf("una imagen debería servirse en línea, y salió %q %q", tipo, disposicion)
	}
	if _, disposicion := ServirCon("informe.pdf", false); disposicion != "attachment" {
		t.Fatal("lo que no se previsualiza se descarga")
	}
	if tipo, disposicion := ServirCon("hoja.xlsx", true); disposicion != "attachment" || tipo != "application/octet-stream" {
		t.Fatalf("una hoja de cálculo se descarga, y salió %q %q", tipo, disposicion)
	}
}

// La carpeta de los adjuntos: **el año y el número del ticket**, que es lo que hace que el disco se
// pueda mirar y entender (docs/modules/tickets.md, sección 2.3).
func TestCarpetaDelTicket(t *testing.T) {
	casos := map[string]string{
		"CS-2026-0042":     "2026/CS-2026-0042",
		"INT-CS-2026-0042": "2026/INT-CS-2026-0042",
		"ACME-2025-0007":   "2025/ACME-2025-0007",
		// Un número de cinco dígitos, que es lo que pasa a partir del 9999.
		"CS-2026-10000": "2026/CS-2026-10000",
	}

	for numero, esperado := range casos {
		año := 2026
		if strings.HasPrefix(numero, "ACME") {
			año = 2025
		}

		if obtenido := carpetasDelTicket(año, numero); obtenido != esperado {
			t.Fatalf("la carpeta de %q debería ser %q y es %q", numero, esperado, obtenido)
		}
	}

	// **Nada que venga de fuera entra en una ruta sin limpiarse**: ni una barra, ni dos puntos, ni un
	// `..`, ni un acento. Es la misma razón por la que el nombre del archivo se genera.
	sucios := map[string]string{
		"../../etc/passwd":           "2026/etcpasswd",
		"CS-2026-0042/../../secreto": "2026/CS-2026-0042secreto",
		"CS 2026 0042":               "2026/CS20260042",
		"cs-2026-0042":               "2026/cs-2026-0042",
		"CS-2026-0042/otra":          "2026/CS-2026-0042otra",
	}
	for sucio, esperado := range sucios {
		obtenido := carpetasDelTicket(2026, sucio)
		if obtenido != esperado {
			t.Fatalf("la carpeta de %q debería ser %q y es %q", sucio, esperado, obtenido)
		}
		if strings.Contains(obtenido, "..") || strings.Count(obtenido, "/") != 1 {
			t.Fatalf("%q no puede acabar en una ruta que se salga: %q", sucio, obtenido)
		}
	}
}

// TestFiltros: página desde 1 y un tope, para que nadie se lleve la tabla entera.
func TestFiltros(t *testing.T) {
	normalizados := NormalizarFiltros(repositories.Filtros{Page: 0, PerPage: 0})
	if normalizados.Page != 1 || normalizados.PerPage != 25 {
		t.Fatalf("los filtros vacíos quedan en %d/%d", normalizados.Page, normalizados.PerPage)
	}

	if grande := NormalizarFiltros(repositories.Filtros{Page: 2, PerPage: 100000}); grande.PerPage != 25 {
		t.Fatalf("un `perPage` enorme tiene que quedar en 25, y quedó en %d", grande.PerPage)
	}

	if valido := NormalizarFiltros(repositories.Filtros{Page: 3, PerPage: 50}); valido.PerPage != 50 || valido.Page != 3 {
		t.Fatal("una página razonable se respeta tal cual")
	}
}
