package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"catalina-support/backend/modules/ai/repositories"
)

// El ticket de las pruebas: los mismos números que se leen en los correos.
const (
	ticketDePrueba  = "ACME-2026-0042"
	internoPrueba   = "INT-ACME-2026-0042"
	modeloDePrueba  = "qwen2.5-1.5b-instruct"
	textoDePrueba   = "Asunto: No puedo entrar\nDescripción: me rechaza la contraseña desde esta mañana\n"
	asuntoDePrueba  = "Asunto: No puedo entrar"
	resumenEsperado = "El usuario no puede entrar desde esta mañana."
	resumenIngles   = "The user cannot sign in since this morning."
)

// ============================================================================
// El motor de mentira
// ============================================================================

// respuestaDeMentira es lo que contesta el motor de mentira a una llamada.
type respuestaDeMentira struct {
	codigo    int
	contenido string
	// colgado deja la petición esperando hasta que el cliente se canse: es un motor que no contesta.
	colgado bool
	// espera retrasa la respuesta: es un motor lento.
	espera time.Duration
	// suelta deja la petición esperando a que la prueba la suelte, para poder mirar lo que pasa con una
	// llamada en vuelo.
	suelta chan struct{}
}

// motorDeMentira es el motor de IA de las pruebas: un servidor HTTP que contesta lo que se le diga y
// **apunta lo que le pidieron**, que es lo que permite comprobar los encargos y las llamadas.
type motorDeMentira struct {
	servidor *httptest.Server
	suelta   chan struct{}

	mu       sync.Mutex
	encargos []encargo
	// ultima es el último cuerpo de petición recibido, tal cual: es lo que permite comprobar que se
	// pide el JSON como tal (`response_format`), la temperatura y el tope de piezas.
	ultima peticion
}

// nuevoMotorDeMentira levanta el servidor. El motor de verdad vive en un contenedor aparte, así que
// aquí se habla con él por HTTP: lo que se prueba es el cliente y la cola, no el modelo.
func nuevoMotorDeMentira(t *testing.T, responder func(veces int) respuestaDeMentira) *motorDeMentira {
	t.Helper()

	mentira := &motorDeMentira{suelta: make(chan struct{})}

	mentira.servidor = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}

		var recibida peticion
		cuerpo, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(cuerpo, &recibida)

		mentira.mu.Lock()
		mentira.ultima = recibida
		mentira.encargos = append(mentira.encargos, encargo{
			sistema: contenidoDe(recibida.Messages, "system"),
			usuario: contenidoDe(recibida.Messages, "user"),
		})
		veces := len(mentira.encargos)
		mentira.mu.Unlock()

		contestar := responder(veces)

		switch {
		case contestar.colgado:
			select {
			case <-r.Context().Done():
			case <-mentira.suelta:
			}
			return
		case contestar.suelta != nil:
			select {
			case <-contestar.suelta:
			case <-r.Context().Done():
				return
			}
		case contestar.espera > 0:
			select {
			case <-time.After(contestar.espera):
			case <-r.Context().Done():
				return
			}
		}

		if contestar.codigo != 0 && contestar.codigo != http.StatusOK {
			w.WriteHeader(contestar.codigo)
			return
		}

		salida, _ := json.Marshal(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": contestar.contenido}},
			},
		})

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(salida)
	}))

	// El servidor se cierra **después** de soltar lo que esté colgado: si no, cerrar esperaría a unas
	// peticiones que no van a contestar nunca.
	t.Cleanup(mentira.servidor.Close)
	t.Cleanup(func() { close(mentira.suelta) })

	return mentira
}

// motorQueContesta es un motor que siempre dice lo mismo.
func motorQueContesta(t *testing.T, contenido string) *motorDeMentira {
	t.Helper()

	return nuevoMotorDeMentira(t, func(int) respuestaDeMentira {
		return respuestaDeMentira{contenido: contenido}
	})
}

// motorQueNoContesta deja las peticiones colgadas: es el contenedor apagado.
func motorQueNoContesta(t *testing.T) *motorDeMentira {
	t.Helper()

	return nuevoMotorDeMentira(t, func(int) respuestaDeMentira {
		return respuestaDeMentira{colgado: true}
	})
}

func (m *motorDeMentira) URL() string { return m.servidor.URL }

// veces dice cuántas veces le han pedido un resumen al motor.
func (m *motorDeMentira) veces() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.encargos)
}

// encargosHechos devuelve los encargos recibidos, en orden.
func (m *motorDeMentira) encargosHechos() []encargo {
	m.mu.Lock()
	defer m.mu.Unlock()

	copiados := make([]encargo, len(m.encargos))
	copy(copiados, m.encargos)

	return copiados
}

// peticionHecha devuelve el último cuerpo de petición recibido.
func (m *motorDeMentira) peticionHecha() peticion {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.ultima
}

// jsonDeDosIdiomas es la respuesta que se le pide al motor: `{"es": "…", "en": "…"}`.
func jsonDeDosIdiomas(es, en string) string {
	salida, _ := json.Marshal(map[string]string{"es": es, "en": en})

	return string(salida)
}

// contenidoDe saca el contenido de un papel de la conversación.
func contenidoDe(mensajes []mensaje, papel string) string {
	for _, m := range mensajes {
		if m.Role == papel {
			return m.Content
		}
	}

	return ""
}

// ============================================================================
// La base de mentira
// ============================================================================

// almacenDeMentira es la base de datos de las pruebas: guarda las filas en memoria y **cuenta lo que
// le piden**, que es lo que permite comprobar que la lista se lee en una sola consulta.
type almacenDeMentira struct {
	mu    sync.Mutex
	filas map[string]repositories.AIInsight

	lecturas   int
	intentos   int
	resultados int
}

func nuevoAlmacenDeMentira() *almacenDeMentira {
	return &almacenDeMentira{filas: make(map[string]repositories.AIInsight)}
}

func (a *almacenDeMentira) MarcarPendiente(numero, tipo string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.filas == nil {
		a.filas = make(map[string]repositories.AIInsight)
	}

	a.filas[claveDePrueba(numero, tipo)] = repositories.AIInsight{
		TicketNumber: numero,
		Kind:         tipo,
		State:        repositories.StatePendiente,
		RequestedAt:  time.Now(),
	}

	return nil
}

func (a *almacenDeMentira) GuardarResultado(numero, tipo, es, en, modelo string, generadoEn time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	fila := a.filas[claveDePrueba(numero, tipo)]
	fila.TicketNumber = numero
	fila.Kind = tipo
	fila.State = repositories.StateListo
	fila.TextEs = &es
	fila.TextEn = &en
	fila.Model = &modelo
	fila.ErrorKey = nil
	fila.GeneratedAt = &generadoEn

	a.filas[claveDePrueba(numero, tipo)] = fila
	a.resultados++

	return nil
}

func (a *almacenDeMentira) MarcarError(numero, tipo, clave string) error {
	return a.marcarFallo(numero, tipo, repositories.StateError, clave)
}

func (a *almacenDeMentira) MarcarSinMotor(numero, tipo, clave string) error {
	return a.marcarFallo(numero, tipo, repositories.StateSinMotor, clave)
}

func (a *almacenDeMentira) marcarFallo(numero, tipo, estado, clave string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	fila := a.filas[claveDePrueba(numero, tipo)]
	fila.TicketNumber = numero
	fila.Kind = tipo
	fila.State = estado
	fila.ErrorKey = &clave
	fila.TextEs = nil
	fila.TextEn = nil
	fila.Model = nil
	fila.GeneratedAt = nil

	a.filas[claveDePrueba(numero, tipo)] = fila

	return nil
}

func (a *almacenDeMentira) SubirIntentos(numero, tipo string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	clave := claveDePrueba(numero, tipo)

	fila, hay := a.filas[clave]
	if !hay {
		// Como el `UPDATE` de verdad: si no hay fila, no hay nada que contar.
		return nil
	}

	fila.Attempts++
	a.filas[clave] = fila
	a.intentos++

	return nil
}

func (a *almacenDeMentira) PorNumeros(numeros []string) ([]repositories.AIInsight, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// **Una consulta**, se pidan los tickets que se pidan: es lo que se está comprobando.
	a.lecturas++

	pedidos := make(map[string]bool, len(numeros))
	for _, numero := range numeros {
		pedidos[numero] = true
	}

	var filas []repositories.AIInsight
	for _, fila := range a.filas {
		if pedidos[fila.TicketNumber] {
			filas = append(filas, fila)
		}
	}

	return filas, nil
}

func (a *almacenDeMentira) Pendientes(maxIntentos int) ([]repositories.AIInsight, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// La misma regla que el repositorio de verdad: lo que quedó a medias y lo que no agotó sus
	// intentos. Lo que está en `error` **no** sale de aquí.
	var filas []repositories.AIInsight
	for _, fila := range a.filas {
		if fila.State == repositories.StatePendiente ||
			(fila.State == repositories.StateSinMotor && fila.Attempts < maxIntentos) {
			filas = append(filas, fila)
		}
	}

	return filas, nil
}

// Las filas que se dejan preparadas antes de empezar una prueba.

func (a *almacenDeMentira) pendiente(numero, tipo string) {
	a.poner(repositories.AIInsight{
		TicketNumber: numero, Kind: tipo, State: repositories.StatePendiente, RequestedAt: time.Now(),
	})
}

func (a *almacenDeMentira) caido(numero, tipo string, intentos int) {
	a.poner(repositories.AIInsight{
		TicketNumber: numero, Kind: tipo, State: repositories.StateSinMotor,
		ErrorKey: puntero(ClaveSinMotor), Attempts: intentos, RequestedAt: time.Now(),
	})
}

func (a *almacenDeMentira) enError(numero, tipo string) {
	a.poner(repositories.AIInsight{
		TicketNumber: numero, Kind: tipo, State: repositories.StateError,
		ErrorKey: puntero(ClaveInvalida), RequestedAt: time.Now(),
	})
}

func (a *almacenDeMentira) conTexto(numero, tipo, es, en string) {
	a.poner(repositories.AIInsight{
		TicketNumber: numero, Kind: tipo, State: repositories.StateListo,
		TextEs: &es, TextEn: &en, RequestedAt: time.Now(),
	})
}

func (a *almacenDeMentira) poner(fila repositories.AIInsight) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.filas[claveDePrueba(fila.TicketNumber, fila.Kind)] = fila
}

// Las lecturas que hacen las pruebas.

func (a *almacenDeMentira) estado(numero, tipo string) string {
	return a.fila(numero, tipo).State
}

func (a *almacenDeMentira) fila(numero, tipo string) repositories.AIInsight {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.filas[claveDePrueba(numero, tipo)]
}

func (a *almacenDeMentira) consultas() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.lecturas
}

func (a *almacenDeMentira) intentosHechos() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.intentos
}

func (a *almacenDeMentira) resumenesGuardados() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.resultados
}

// fuenteDeMentira hace de `tickets` en la puesta al día: es quien vuelve a armar el texto.
type fuenteDeMentira struct {
	mu      sync.Mutex
	numeros []string
}

func (f *fuenteDeMentira) TextoParaElMotor(numero string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.numeros = append(f.numeros, numero)

	return "Asunto: ticket " + numero, nil
}

func (f *fuenteDeMentira) pedidos() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	copiados := make([]string, len(f.numeros))
	copy(copiados, f.numeros)

	return copiados
}

// ============================================================================
// Utilidades
// ============================================================================

// ajustesDePrueba son los ajustes de una instalación con motor.
func ajustesDePrueba(url string) Ajustes {
	return Ajustes{
		URL:      url,
		Modelo:   modeloDePrueba,
		Contexto: 4096,
		Palabras: 40,
		Espera:   2 * time.Second,
	}
}

// servicioDePrueba monta el servicio con un motor y un almacén de mentira. Sin motor —`motor == nil`—
// es una instalación que no lo tiene puesto.
func servicioDePrueba(t *testing.T, motor *motorDeMentira, almacen *almacenDeMentira) *Service {
	t.Helper()

	url := ""
	if motor != nil {
		url = motor.URL()
	}

	ajustes := ajustesDePrueba(url)

	servicio := nuevoServicio(almacen, ajustes)
	// Las esperas de verdad son de segundos: en una prueba se acortan, que lo que se comprueba es que
	// se reintenta, no cuánto se espera.
	servicio.esperaBase = 5 * time.Millisecond

	t.Cleanup(servicio.Parar)

	return servicio
}

// esperarA espera a que pase algo, en vez de dormir un rato fijo: la cola trabaja en segundo plano y
// una prueba que duerme de más es una prueba que un día falla.
func esperarA(t *testing.T, quePasa func() bool) {
	t.Helper()

	limite := time.Now().Add(5 * time.Second)
	for time.Now().Before(limite) {
		if quePasa() {
			return
		}

		time.Sleep(2 * time.Millisecond)
	}

	t.Fatal("se agotó el tiempo esperando a que el módulo hiciera su trabajo")
}

func claveDePrueba(numero, tipo string) string {
	return trabajo{numero: numero, tipo: Tipo(tipo)}.clave()
}

func puntero(texto string) *string { return &texto }

// ============================================================================
// Las pruebas
// ============================================================================

func TestElJSONValidoSeGuardaEnLosDosIdiomas(t *testing.T) {
	motor := motorQueContesta(t, jsonDeDosIdiomas(resumenEsperado, resumenIngles))
	almacen := nuevoAlmacenDeMentira()
	servicio := servicioDePrueba(t, motor, almacen)

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: textoDePrueba})

	esperarA(t, func() bool { return almacen.estado(ticketDePrueba, "motivo") == repositories.StateListo })

	fila := almacen.fila(ticketDePrueba, "motivo")
	if textoDe(fila.TextEs) != resumenEsperado {
		t.Errorf("el español guardado es %q y se esperaba %q", textoDe(fila.TextEs), resumenEsperado)
	}
	if textoDe(fila.TextEn) != resumenIngles {
		t.Errorf("el inglés guardado es %q y se esperaba %q", textoDe(fila.TextEn), resumenIngles)
	}
	if textoDe(fila.Model) != modeloDePrueba {
		t.Errorf("el modelo guardado es %q y se esperaba %q", textoDe(fila.Model), modeloDePrueba)
	}
	if fila.GeneratedAt == nil {
		t.Error("el resumen quedó sin fecha de generación")
	}
	if textoDe(fila.ErrorKey) != "" {
		t.Errorf("un resumen listo no lleva clave de error, y lleva %q", textoDe(fila.ErrorKey))
	}

	// Al motor le llegó el texto del ticket, que es lo único que puede resumir.
	encargos := motor.encargosHechos()
	if len(encargos) != 1 {
		t.Fatalf("el motor recibió %d encargos y se esperaba uno", len(encargos))
	}
	if !strings.Contains(encargos[0].usuario, asuntoDePrueba) {
		t.Errorf("el texto que recibió el motor no lleva el asunto del ticket: %q", encargos[0].usuario)
	}
}

func TestElJSONEnvueltoEnUnBloqueDeCodigoSeRescata(t *testing.T) {
	envuelto := "Aquí tienes el resumen:\n```json\n" + jsonDeDosIdiomas(resumenEsperado, resumenIngles) +
		"\n```\nEspero que sirva."

	motor := motorQueContesta(t, envuelto)
	almacen := nuevoAlmacenDeMentira()
	servicio := servicioDePrueba(t, motor, almacen)

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: UltimaAccion, Texto: textoDePrueba})

	esperarA(t, func() bool { return almacen.estado(ticketDePrueba, "ultima_accion") == repositories.StateListo })

	fila := almacen.fila(ticketDePrueba, "ultima_accion")
	if textoDe(fila.TextEs) != resumenEsperado || textoDe(fila.TextEn) != resumenIngles {
		t.Errorf("el JSON envuelto no se rescató: es=%q en=%q", textoDe(fila.TextEs), textoDe(fila.TextEn))
	}
}

func TestUnaRespuestaQueNoValeDejaErrorConSuClave(t *testing.T) {
	motor := motorQueContesta(t, "Pues no sé, depende del ticket.")
	almacen := nuevoAlmacenDeMentira()
	servicio := servicioDePrueba(t, motor, almacen)

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: textoDePrueba})

	esperarA(t, func() bool { return almacen.estado(ticketDePrueba, "motivo") == repositories.StateError })

	fila := almacen.fila(ticketDePrueba, "motivo")
	if textoDe(fila.ErrorKey) != ClaveInvalida {
		t.Errorf("la clave del error es %q y se esperaba %q", textoDe(fila.ErrorKey), ClaveInvalida)
	}
	if textoDe(fila.TextEs) != "" || textoDe(fila.TextEn) != "" {
		t.Error("un campo en error no puede quedarse con texto")
	}

	// **Se reintenta, pero no para siempre**: volver a preguntarle a un modelo que ya contestó mal tiene
	// sentido porque es estocástico, y dejarlo en bucle no (decisión 8).
	if veces := motor.veces(); veces != intentosRespuesta {
		t.Errorf("el motor recibió %d peticiones y se esperaban %d", veces, intentosRespuesta)
	}
	if fila.Attempts != intentosRespuesta {
		t.Errorf("el contador de intentos quedó en %d y se esperaba %d", fila.Attempts, intentosRespuesta)
	}
}

func TestElMotorQueNoContestaDejaSinMotorYNoBloquea(t *testing.T) {
	motor := motorQueNoContesta(t)
	almacen := nuevoAlmacenDeMentira()

	ajustes := ajustesDePrueba(motor.URL())
	ajustes.Espera = 150 * time.Millisecond

	servicio := nuevoServicio(almacen, ajustes)
	servicio.esperaBase = 5 * time.Millisecond
	t.Cleanup(servicio.Parar)

	empezado := time.Now()
	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: textoDePrueba})

	if tardanza := time.Since(empezado); tardanza > 100*time.Millisecond {
		t.Fatalf("Pedir tardó %s: no puede esperar al motor", tardanza)
	}

	esperarA(t, func() bool { return almacen.estado(ticketDePrueba, "motivo") == repositories.StateSinMotor })

	fila := almacen.fila(ticketDePrueba, "motivo")
	if textoDe(fila.ErrorKey) != ClaveSinMotor {
		t.Errorf("la clave del error es %q y se esperaba %q", textoDe(fila.ErrorKey), ClaveSinMotor)
	}
	if fila.Attempts != intentosMotor || motor.veces() != intentosMotor {
		t.Errorf("se intentó %d veces y el motor recibió %d peticiones: se esperaban %d",
			fila.Attempts, motor.veces(), intentosMotor)
	}

	// Lo que estaba `pendiente` estuvo dicho desde el primer momento: la pantalla pudo decir que se
	// estaba escribiendo sin esperar a nadie.
	if fila.RequestedAt.IsZero() {
		t.Error("el campo no quedó apuntado como pedido")
	}
}

func TestPedirNoBloqueaAunqueElMotorTarde(t *testing.T) {
	motor := nuevoMotorDeMentira(t, func(int) respuestaDeMentira {
		return respuestaDeMentira{
			contenido: jsonDeDosIdiomas(resumenEsperado, resumenIngles),
			espera:    2 * time.Second,
		}
	})
	almacen := nuevoAlmacenDeMentira()
	servicio := servicioDePrueba(t, motor, almacen)

	empezado := time.Now()
	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: textoDePrueba})
	tardanza := time.Since(empezado)

	if tardanza > 250*time.Millisecond {
		t.Fatalf("Pedir tardó %s: quien crea un ticket no puede esperar al motor", tardanza)
	}

	// Y el campo quedó `pendiente` en la base desde el primer momento.
	if estado := almacen.estado(ticketDePrueba, "motivo"); estado != repositories.StatePendiente {
		t.Errorf("el campo quedó en %q y se esperaba %q", estado, repositories.StatePendiente)
	}
}

func TestLosDosEncargosSonDistintosYSonDosPeticiones(t *testing.T) {
	motor := motorQueContesta(t, jsonDeDosIdiomas(resumenEsperado, resumenIngles))
	almacen := nuevoAlmacenDeMentira()
	servicio := servicioDePrueba(t, motor, almacen)

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: textoDePrueba})
	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: UltimaAccion, Texto: textoDePrueba})

	esperarA(t, func() bool {
		return almacen.estado(ticketDePrueba, "motivo") == repositories.StateListo &&
			almacen.estado(ticketDePrueba, "ultima_accion") == repositories.StateListo
	})

	encargos := motor.encargosHechos()
	if len(encargos) != 2 {
		t.Fatalf("el motor recibió %d encargos y se esperaban dos, uno por campo", len(encargos))
	}
	if encargos[0].sistema == encargos[1].sistema {
		t.Error("el encargo del motivo y el de la última acción no pueden ser el mismo texto")
	}
}

func TestUnEncargoPorTicketYCampoSustituyeAlQueEsperaba(t *testing.T) {
	suelta := make(chan struct{})

	motor := nuevoMotorDeMentira(t, func(veces int) respuestaDeMentira {
		contestar := respuestaDeMentira{contenido: jsonDeDosIdiomas(resumenEsperado, resumenIngles)}
		if veces == 1 {
			// La primera llamada se queda en vuelo: lo que se pide mientras tanto tiene que sustituir al
			// encargo que espera, y no encolarse cinco veces.
			contestar.suelta = suelta
		}

		return contestar
	})

	almacen := nuevoAlmacenDeMentira()

	ajustes := ajustesDePrueba(motor.URL())
	ajustes.Espera = 5 * time.Second

	servicio := nuevoServicio(almacen, ajustes)
	servicio.esperaBase = 5 * time.Millisecond
	t.Cleanup(servicio.Parar)

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: "Texto de la primera vez"})
	esperarA(t, func() bool { return motor.veces() == 1 })

	for i := 0; i < 5; i++ {
		servicio.Pedir(Entrada{
			Numero: ticketDePrueba,
			Tipo:   Motivo,
			Texto:  fmt.Sprintf("Texto que ya no sirve, %d", i),
		})
	}

	close(suelta)

	esperarA(t, func() bool { return almacen.estado(ticketDePrueba, "motivo") == repositories.StateListo })
	esperarA(t, func() bool { return motor.veces() == 2 })

	if veces := motor.veces(); veces != 2 {
		t.Errorf("el motor recibió %d peticiones y sólo pueden ser dos: la que estaba en vuelo y la que sustituye a las demás", veces)
	}
}

func TestDeDevuelveLosDosCamposDeVariosTicketsEnUnaConsulta(t *testing.T) {
	almacen := nuevoAlmacenDeMentira()
	almacen.conTexto(ticketDePrueba, "motivo", "No puede entrar.", "Cannot sign in.")
	almacen.conTexto(ticketDePrueba, "ultima_accion", "Se pidió una captura.", "A screenshot was requested.")
	almacen.conTexto(internoPrueba, "ultima_accion", "Desarrollo lo está mirando.", "Development is on it.")
	almacen.pendiente("ACME-2026-0043", "motivo")

	servicio := servicioDePrueba(t, nil, almacen)

	resumenes, err := servicio.De([]string{ticketDePrueba, internoPrueba, "ACME-2026-0043", "ACME-2026-9999"})
	if err != nil {
		t.Fatalf("De devolvió un error: %v", err)
	}

	if len(resumenes) != 3 {
		t.Fatalf("el mapa trae %d tickets y se esperaban 3", len(resumenes))
	}
	if resumenes[ticketDePrueba].Motivo.Es != "No puede entrar." {
		t.Errorf("el motivo es %q", resumenes[ticketDePrueba].Motivo.Es)
	}
	if resumenes[ticketDePrueba].UltimaAccion.En != "A screenshot was requested." {
		t.Errorf("la última acción en inglés es %q", resumenes[ticketDePrueba].UltimaAccion.En)
	}
	if resumenes[ticketDePrueba].Motivo.Estado != EstadoListo {
		t.Errorf("el estado del motivo es %q", resumenes[ticketDePrueba].Motivo.Estado)
	}
	if resumenes[internoPrueba].UltimaAccion.Es != "Desarrollo lo está mirando." {
		t.Errorf("la última acción del interno es %q", resumenes[internoPrueba].UltimaAccion.Es)
	}
	if resumenes["ACME-2026-0043"].Motivo.Estado != EstadoPendiente {
		t.Error("un campo pendiente se lee con su estado")
	}

	// **Una sola consulta** para toda la lista, no una por fila (decisión 9).
	if consultas := almacen.consultas(); consultas != 1 {
		t.Errorf("la lista se leyó en %d consultas y tiene que ser una", consultas)
	}

	// Un ticket sin resumen no aparece: no hay nada que enseñar y no se inventa.
	if _, hay := resumenes["ACME-2026-9999"]; hay {
		t.Error("un ticket sin resumen no puede salir en el mapa")
	}
}

func TestDeSinNumerosNoPreguntaNadaALaBase(t *testing.T) {
	almacen := nuevoAlmacenDeMentira()
	servicio := servicioDePrueba(t, nil, almacen)

	resumenes, err := servicio.De(nil)
	if err != nil {
		t.Fatalf("De devolvió un error: %v", err)
	}
	if len(resumenes) != 0 {
		t.Errorf("el mapa trae %d tickets y no se pidió ninguno", len(resumenes))
	}
	if consultas := almacen.consultas(); consultas != 0 {
		t.Errorf("se preguntó %d veces a la base sin tener nada que preguntar", consultas)
	}
}

func TestSinMotorConfiguradoNoSeEncolaNadaYQuedaSinMotor(t *testing.T) {
	almacen := nuevoAlmacenDeMentira()
	servicio := servicioDePrueba(t, nil, almacen)

	if servicio.Disponible() {
		t.Error("sin URL configurada no hay motor disponible")
	}

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: UltimaAccion, Texto: textoDePrueba})

	fila := almacen.fila(ticketDePrueba, "ultima_accion")
	if fila.State != repositories.StateSinMotor {
		t.Errorf("el campo quedó en %q y se esperaba %q", fila.State, repositories.StateSinMotor)
	}
	if textoDe(fila.ErrorKey) != ClaveSinMotor {
		t.Errorf("la clave del error es %q y se esperaba %q", textoDe(fila.ErrorKey), ClaveSinMotor)
	}

	// Ni se encoló nada ni se le pidió nada a nadie: la aplicación entera funciona sin motor.
	servicio.mu.Lock()
	esperando := len(servicio.cola)
	servicio.mu.Unlock()

	if esperando != 0 {
		t.Errorf("la cola tiene %d trabajos y sin motor no puede haber ninguno", esperando)
	}
	if intentos := almacen.intentosHechos(); intentos != 0 {
		t.Errorf("se contaron %d intentos y sin motor no se intenta nada", intentos)
	}
}

func TestSinMotorLaListaSeSigueLeyendoYLosTicketsSiguenSuVida(t *testing.T) {
	almacen := nuevoAlmacenDeMentira()
	almacen.caido(ticketDePrueba, "motivo", 1)

	servicio := servicioDePrueba(t, nil, almacen)

	// Lo que quedó de cuando sí había motor se lee igual: la lista no depende de que el contenedor
	// esté levantado (decisión 2).
	resumenes, err := servicio.De([]string{ticketDePrueba})
	if err != nil {
		t.Fatalf("De devolvió un error sin motor: %v", err)
	}
	if resumenes[ticketDePrueba].Motivo.Estado != EstadoSinMotor {
		t.Errorf("el estado leído es %q y se esperaba %q", resumenes[ticketDePrueba].Motivo.Estado, EstadoSinMotor)
	}

	// Y sin motor no hay nada que retomar.
	servicio.Retomar()
	if almacen.resumenesGuardados() != 0 {
		t.Error("sin motor no se puede haber escrito ningún resumen")
	}
}

func TestRetomarVuelveAPedirLoSuyoYOlvidaLoQueEstaEnError(t *testing.T) {
	motor := motorQueContesta(t, jsonDeDosIdiomas(resumenEsperado, resumenIngles))
	almacen := nuevoAlmacenDeMentira()

	// Lo que quedó a medias cuando se apagó el backend, y lo que ya está decidido.
	almacen.pendiente("ACME-2026-0001", "motivo")
	almacen.pendiente("ACME-2026-0002", "ultima_accion")
	almacen.caido("ACME-2026-0003", "motivo", 1)
	almacen.enError("ACME-2026-0004", "motivo")
	almacen.caido("ACME-2026-0005", "motivo", intentosMotor)

	fuente := &fuenteDeMentira{}

	servicio := servicioDePrueba(t, motor, almacen)
	servicio.SetFuente(fuente)

	servicio.Retomar()

	esperarA(t, func() bool { return almacen.resumenesGuardados() == 3 })

	if veces := motor.veces(); veces != 3 {
		t.Errorf("el motor recibió %d peticiones y se esperaban 3", veces)
	}
	if pedidos := fuente.pedidos(); len(pedidos) != 3 {
		t.Errorf("se volvió a armar el texto de %d tickets y se esperaban 3", len(pedidos))
	}

	// **Lo que está en `error` no se reintenta solo**, y un campo que agotó sus intentos tampoco:
	// reiniciar el backend no puede ser la forma de saltarse el tope (decisión 8 y 10).
	if estado := almacen.estado("ACME-2026-0004", "motivo"); estado != repositories.StateError {
		t.Errorf("el campo en error quedó en %q", estado)
	}
	if estado := almacen.estado("ACME-2026-0005", "motivo"); estado != repositories.StateSinMotor {
		t.Errorf("el campo que agotó sus intentos quedó en %q", estado)
	}
}

func TestRetomarSinQuienArmeElTextoNoPierdeNada(t *testing.T) {
	motor := motorQueContesta(t, jsonDeDosIdiomas(resumenEsperado, resumenIngles))
	almacen := nuevoAlmacenDeMentira()
	almacen.pendiente(ticketDePrueba, "motivo")

	servicio := servicioDePrueba(t, motor, almacen)

	// Sin fuente no se puede rehacer el encargo, y lo que había se queda como estaba: `pendiente`, para
	// que la puesta al día del próximo arranque lo recoja.
	servicio.Retomar()

	if veces := motor.veces(); veces != 0 {
		t.Errorf("se le pidieron %d resúmenes al motor sin poder armar su texto", veces)
	}
	if estado := almacen.estado(ticketDePrueba, "motivo"); estado != repositories.StatePendiente {
		t.Errorf("el campo quedó en %q y se esperaba %q", estado, repositories.StatePendiente)
	}
}

func TestDisponibleDiceSiElMotorResponde(t *testing.T) {
	almacen := nuevoAlmacenDeMentira()

	conMotor := servicioDePrueba(t, motorQueContesta(t, jsonDeDosIdiomas("Uno", "One")), almacen)
	if !conMotor.Disponible() {
		t.Error("un motor que responde tiene que estar disponible")
	}

	sinMotor := servicioDePrueba(t, nil, almacen)
	if sinMotor.Disponible() {
		t.Error("sin URL configurada no hay motor disponible")
	}

	// Un motor que no responde no está disponible, aunque haya una URL puesta: «disponible» es lo que
	// pasa ahora y no lo que dice la configuración.
	apagado := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := apagado.URL
	apagado.Close()

	ajustes := ajustesDePrueba(url)
	ajustes.Espera = time.Second

	caido := nuevoServicio(almacen, ajustes)
	t.Cleanup(caido.Parar)

	if caido.Disponible() {
		t.Error("un motor que no responde no puede estar disponible")
	}
}

func TestPararCierraLaColaSinEsperarAlMotor(t *testing.T) {
	motor := motorQueNoContesta(t)
	almacen := nuevoAlmacenDeMentira()

	ajustes := ajustesDePrueba(motor.URL())
	// Un motor que se queda colgado para siempre: apagar no puede quedarse esperándolo.
	ajustes.Espera = 30 * time.Second

	servicio := nuevoServicio(almacen, ajustes)
	servicio.esperaBase = 5 * time.Millisecond

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: textoDePrueba})
	esperarA(t, func() bool { return motor.veces() >= 1 })

	empezado := time.Now()
	servicio.Parar()

	if tardanza := time.Since(empezado); tardanza > time.Second {
		t.Errorf("Parar tardó %s: tiene que cortar la llamada en vuelo", tardanza)
	}

	// Parar dos veces no revienta, y pedir algo después de parar tampoco: el backend se apaga sin que
	// nada se caiga por el camino.
	servicio.Parar()
	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: UltimaAccion, Texto: textoDePrueba})
}

func TestElTopeDePalabrasSeRespetaAunqueElModeloSePase(t *testing.T) {
	largo := strings.TrimSpace(strings.Repeat("palabra ", 200))

	motor := motorQueContesta(t, jsonDeDosIdiomas(largo, largo))
	almacen := nuevoAlmacenDeMentira()

	ajustes := ajustesDePrueba(motor.URL())
	ajustes.Palabras = 20

	servicio := nuevoServicio(almacen, ajustes)
	servicio.esperaBase = 5 * time.Millisecond
	t.Cleanup(servicio.Parar)

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: textoDePrueba})

	esperarA(t, func() bool { return almacen.estado(ticketDePrueba, "motivo") == repositories.StateListo })

	fila := almacen.fila(ticketDePrueba, "motivo")
	if palabras := len(strings.Fields(textoDe(fila.TextEs))); palabras > 20 {
		t.Errorf("el resumen guardado tiene %d palabras y el tope es 20", palabras)
	}
	if !strings.HasSuffix(textoDe(fila.TextEs), "…") {
		t.Error("lo que se corta tiene que decirse con unos puntos suspensivos")
	}
}

func TestAlMotorLeLlegaElFinalDelTextoYNoElPrincipio(t *testing.T) {
	motor := motorQueContesta(t, jsonDeDosIdiomas(resumenEsperado, resumenIngles))
	almacen := nuevoAlmacenDeMentira()
	servicio := servicioDePrueba(t, motor, almacen)

	// Un ticket con doscientos comentarios: lo que hay que leer es lo último, y lo primero se queda
	// fuera del encargo (decisión 6).
	texto := "PRINCIPIO: esto ya no importa\n" + strings.Repeat("relleno que hace bulto\n", 400) +
		"Última acción: se pidió una captura"

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: UltimaAccion, Texto: texto})

	esperarA(t, func() bool { return almacen.estado(ticketDePrueba, "ultima_accion") == repositories.StateListo })

	enviado := motor.encargosHechos()[0].usuario

	if strings.Contains(enviado, "PRINCIPIO") {
		t.Error("al motor le llegó el principio del texto, y el recorte quita por el principio")
	}
	if !strings.Contains(enviado, "Última acción: se pidió una captura") {
		t.Errorf("al motor no le llegó el final del texto: %q", enviado[len(enviado)-80:])
	}
}

// El cuerpo de la petición es lo que hace que el modelo conteste el JSON pedido, y está medido contra
// el motor de verdad: `response_format` **más** la exigencia en el mensaje del usuario. Si alguna de
// las dos se cae, el modelo se inventa las claves.
func TestLaPeticionLlevaResponseFormatLaTemperaturaYElTopeDePiezas(t *testing.T) {
	motor := motorQueContesta(t, jsonDeDosIdiomas(resumenEsperado, resumenIngles))
	almacen := nuevoAlmacenDeMentira()

	ajustes := ajustesDePrueba(motor.URL())
	ajustes.Palabras = 60

	servicio := nuevoServicio(almacen, ajustes)
	servicio.esperaBase = 5 * time.Millisecond
	t.Cleanup(servicio.Parar)

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: textoDePrueba})
	esperarA(t, func() bool { return almacen.estado(ticketDePrueba, "motivo") == repositories.StateListo })

	peticion := motor.peticionHecha()

	if peticion.ResponseFormat.Type != "json_object" {
		t.Errorf("no se está pidiendo el JSON como tal: response_format=%q", peticion.ResponseFormat.Type)
	}
	if peticion.Temperature != temperatura {
		t.Errorf("la temperatura es %v y tiene que ser %v", peticion.Temperature, temperatura)
	}
	// Con 60 palabras, 400 piezas: margen de sobra, pero sin pasarse de media ventana.
	if peticion.MaxTokens < 400 {
		t.Errorf("el tope de piezas es %d y con 60 palabras no puede bajar de 400", peticion.MaxTokens)
	}
	if peticion.MaxTokens > ajustes.Contexto/2 {
		t.Errorf("el tope de piezas es %d y no puede pasar de media ventana (%d)", peticion.MaxTokens, ajustes.Contexto/2)
	}
	if peticion.Model != modeloDePrueba {
		t.Errorf("el modelo pedido es %q y se esperaba %q", peticion.Model, modeloDePrueba)
	}
	if len(peticion.Messages) != 2 || peticion.Messages[0].Role != "system" || peticion.Messages[1].Role != "user" {
		t.Errorf("la conversación no es la de dos papeles: %+v", peticion.Messages)
	}
}

// El encargo lleva el tope de palabras y **la plantilla exacta del objeto**, y la lleva en el mensaje
// del usuario, al final: es lo último que lee el modelo antes de escribir.
func TestElEncargoLlevaElTopeDePalabrasYElFormatoPedido(t *testing.T) {
	motor := motorQueContesta(t, jsonDeDosIdiomas(resumenEsperado, resumenIngles))
	almacen := nuevoAlmacenDeMentira()

	ajustes := ajustesDePrueba(motor.URL())
	ajustes.Palabras = 25

	servicio := nuevoServicio(almacen, ajustes)
	servicio.esperaBase = 5 * time.Millisecond
	t.Cleanup(servicio.Parar)

	servicio.Pedir(Entrada{Numero: ticketDePrueba, Tipo: Motivo, Texto: textoDePrueba})
	esperarA(t, func() bool { return almacen.estado(ticketDePrueba, "motivo") == repositories.StateListo })

	encargo := motor.encargosHechos()[0]

	if !strings.Contains(encargo.usuario, "25 palabras") {
		t.Errorf("el encargo no lleva el tope de palabras: %q", encargo.usuario)
	}
	if !strings.Contains(encargo.usuario, "EXACTAMENTE") {
		t.Errorf("el encargo no exige el objeto exacto: %q", encargo.usuario)
	}
	if !strings.HasSuffix(encargo.usuario, `{"es": "…", "en": "…"}`) {
		t.Errorf("la plantilla del objeto tiene que ir al final del encargo: %q", encargo.usuario)
	}
	// El texto del ticket va antes de la exigencia: primero se le da el ticket, después se le pide el
	// formato.
	if !strings.HasPrefix(encargo.usuario, asuntoDePrueba) {
		t.Errorf("el encargo tendría que empezar por el texto del ticket: %q", encargo.usuario[:60])
	}
	if !strings.Contains(encargo.sistema, "DE QUÉ VA EL TICKET") {
		t.Errorf("el mensaje de sistema no dice el oficio: %q", encargo.sistema)
	}
}
