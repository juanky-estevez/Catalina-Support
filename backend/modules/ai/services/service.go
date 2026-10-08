// Package services es el módulo ai: los dos resúmenes que redacta el motor de inteligencia
// artificial —**el motivo del ticket** y **su última acción**— en el idioma global
// (docs/modules/ai.md).
//
// El motor vive en un contenedor aparte y **puede estar caído**: por eso todo se pide en segundo
// plano, con una cola en memoria y reintentos, y nada de lo que hace este módulo deja esperando a
// quien abre, comenta o mueve un ticket. **La aplicación funciona entera sin motor**: es la condición
// que manda sobre las demás (decisión 2).
package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"catalina-support/backend/modules/ai/repositories"

	"github.com/juanky-estevez/go-logs"
)

// Tipo es el campo que se le pide al motor. **Uno por llamada**: el «Motivo» y la «Última acción» son
// encargos distintos y son dos peticiones (docs/modules/ai.md, decisión 7).
type Tipo string

const (
	// Motivo es de qué va el ticket.
	Motivo Tipo = "motivo"
	// UltimaAccion es qué fue lo último que pasó.
	UltimaAccion Tipo = "ultima_accion"
)

// Estado es en qué ha quedado un campo.
type Estado string

const (
	EstadoPendiente Estado = "pendiente"
	EstadoListo     Estado = "listo"
	EstadoError     Estado = "error"
	// EstadoSinMotor: hay motor configurado pero no contesta. **No es lo mismo que `error`**: uno es
	// un contenedor caído y el otro un modelo que ha contestado mal (docs/modules/ai.md, decisión 8).
	EstadoSinMotor Estado = "sin_motor"
)

// Las claves de error del módulo, con el código que les toca. Se guardan en la fila del campo para
// que la pantalla pueda decir qué pasó sin inventarse nada.
const (
	// ClaveSinMotor es un motor que no contesta.
	ClaveSinMotor = "ai.unavailable"
	// ClaveInvalida es una respuesta que no vale.
	ClaveInvalida = "ai.invalid"
)

// Los dos fallos, como errores, para poder distinguirlos dentro del módulo.
var (
	ErrSinMotor          = errors.New(ClaveSinMotor)
	ErrRespuestaInvalida = errors.New(ClaveInvalida)
)

// Entrada es el encargo que se le hace al motor: un ticket, un campo y el texto ya armado.
type Entrada struct {
	Numero string
	Tipo   Tipo
	// Texto lo arma `tickets`, que es quien tiene el ticket: asunto, descripción, adjuntos por su
	// nombre y la conversación, ya en texto plano (docs/modules/ai.md, sección 3).
	Texto string
}

// Resumen es lo que hay de un campo. Text y Language son el contrato actual de idioma global; Es y
// En sólo permiten leer filas creadas antes de la actualización.
type Resumen struct {
	Estado     Estado
	Text       string
	Language   string
	Es         string
	En         string
	ErrorKey   string
	GeneradoEn time.Time
}

// DosResumenes son los dos campos de un ticket.
type DosResumenes struct {
	Motivo       Resumen
	UltimaAccion Resumen
}

// Ajustes es lo que el módulo necesita saber del motor.
type Ajustes struct {
	// URL del motor, por ejemplo `http://catalina_support_ai:8080`. **Vacía es «esta instalación no
	// tiene motor»**, y con eso no se encola nada y todo lo demás funciona igual (decisión 2).
	URL string
	// Modelo con el que se pide, para saber con qué se generó cada texto.
	Modelo string
	// Contexto es la ventana del modelo, en piezas: acota lo que se le puede pedir de respuesta.
	Contexto int
	// Palabras es el tope de cada redacción. Se le pide al motor y, si se pasa, se recorta.
	Palabras int
	// Espera es el tiempo máximo de **una** llamada al motor.
	Espera                                                     time.Duration
	Mode, Provider, AuthType, AuthHeader, Credential, Language string
}

// Configurado dice si hay motor al que preguntar.
//
// No es un fallo que no lo haya: es una instalación que no lo tiene puesto, y esa es la condición
// principal de este módulo (docs/modules/ai.md, decisión 2).
func (a Ajustes) Configurado() bool { return strings.TrimSpace(a.URL) != "" }

// AIDatos es **el motor de IA tal y como lo cuenta la configuración de la instalación**: su
// dirección, su modelo y si hay alguno puesto.
//
// Es un tipo propio de este módulo a propósito: la configuración la declara `settings` y aquí no se
// importa, porque un módulo no importa a otro. El cableado traduce de uno a otro
// (docs/arquitectura.md, sección 4).
type AIDatos struct {
	URL    string
	Modelo string
	// Hay en falso quiere decir «esta instalación no tiene motor»: se usa el respaldo del entorno.
	Hay                                                        bool
	Mode, Provider, AuthType, AuthHeader, Credential, Language string
}

// Configuracion es lo que este módulo necesita de la configuración de la instalación: **el motor**.
//
// Se declara aquí, en quien la usa, y la cumple el módulo `settings`; se conecta en `main.go`. Se lee
// **en cada petición**, no al arrancar: cambiar la dirección o el modelo desde Configuración tiene
// que valer sin reiniciar nada (docs/modules/ai.md).
type Configuracion interface {
	AI() (AIDatos, error)
}

// Los valores de fábrica de lo que no venga puesto, para que una configuración incompleta no deje el
// módulo sin saber cuánto esperar.
const (
	palabrasDeFabrica = 40
	contextoDeFabrica = 4096
	// El tiempo de espera de una llamada al motor, medido contra el motor de verdad sin GPU: lee a
	// ~24 piezas por segundo, así que **un ticket de 6 000 caracteres tarda él solo unos 82 segundos**,
	// y redacta a 5-9 piezas por segundo: una llamada larga se va a los 95-100 segundos. Con 30 segundos
	// se daría por caído un motor que está trabajando, y el campo quedaría en `sin_motor` sin motivo.
	esperaDeFabrica = 240 * time.Second

	// Los intentos de un motor que no contesta, y los de una respuesta que no vale: **el motor caído
	// se reintenta tres veces con espera creciente y una respuesta mala sólo una** (decisión 8).
	intentosMotor = 3
	// **Tres**, no dos: medido con el motor de verdad, una de cada tantas respuestas de un modelo de
	// 1,5B llega a medias —sólo un idioma, o cortada por el tope de piezas— y el reintento es inmediato,
	// así que insistir sale barato. Lo que **no** hay es bucle: a la tercera, el campo queda en `error`
	// con su clave (docs/modules/ai.md, decisión 8).
	intentosRespuesta = 3

	// La espera entre intentos del motor caído, que crece con cada uno: 2 s, 4 s.
	esperaBaseDeFabrica = 2 * time.Second

	// El tiempo del chequeo de «¿está el motor?»: preguntarlo no puede dejar colgada una pantalla.
	esperaDelChequeo = 3 * time.Second

	// El tiempo de la prueba de la conexión de Configuración: lo mismo, y más corto que una llamada
	// de verdad, porque quien la pide está mirando una pantalla.
	esperaDeLaPrueba = 3 * time.Second

	// **Dos trabajadores**, porque el motor es lo caro: sin cola, diez comentarios seguidos pedirían
	// diez resúmenes a la vez a un modelo que atiende de uno en uno (docs/modules/ai.md, sección 6).
	trabajadores = 2
)

// Fuente es quien sabe **volver a armar el texto de un ticket**.
//
// Hace falta para la puesta al día del arranque: `Pedir` recibe el texto ya hecho y no lo guarda —no
// se guarda el texto de un ticket en la tabla de los resúmenes—, así que al reiniciar hay que
// volver a preguntárselo a quien lo armó, que es `tickets`. Se declara aquí, en quien lo usa, y
// `tickets` lo cumple sin saber que existe: la misma forma que tienen `Accounts`, `Mailer` e
// `Insights` (docs/arquitectura.md, sección 4).
type Fuente interface {
	TextoParaElMotor(numero string) (string, error)
}

// Store es lo que el servicio necesita de la base: los mismos métodos del repositorio del módulo,
// detrás de una interfaz para poder probar la cola, los reintentos y los estados sin una base de
// datos de verdad.
type Store interface {
	MarcarPendiente(numero, tipo string) error
	GuardarResultado(numero, tipo, es, en, modelo string, generadoEn time.Time) error
	MarcarError(numero, tipo, clave string) error
	MarcarSinMotor(numero, tipo, clave string) error
	SubirIntentos(numero, tipo string) error
	PorNumeros(numeros []string) ([]repositories.AIInsight, error)
	Pendientes(maxIntentos int) ([]repositories.AIInsight, error)
}

// Service es el módulo: la cola, los trabajadores y el motor.
type Service struct {
	store   Store
	ajustes Ajustes
	motor   *motor

	// configuracion da **el motor que está puesto en la instalación**. Es opcional: sin ella se usa
	// lo del entorno, que es lo que había antes. Se lee en cada petición, no al arrancar.
	configuracion Configuracion

	// La cola vive **en memoria** y se pierde al reiniciar: por eso existe la puesta al día
	// (docs/modules/ai.md, decisión 10). Se guarda por (ticket, campo) para que pedir dos veces lo
	// mismo no encole dos veces.
	mu       sync.Mutex
	cola     map[string]trabajo
	aviso    chan struct{}
	parar    chan struct{}
	cerrada  bool
	enMarcha sync.WaitGroup

	// El contexto de las llamadas al motor: apagarlo corta la que esté en vuelo, para que `Parar` no
	// tenga que esperar a que un motor lento conteste.
	ctx    context.Context
	apagar context.CancelFunc

	// fuente puede ser nula: sin ella la puesta al día no puede rehacer el encargo.
	fuente Fuente

	// esperaBase es la espera entre intentos del motor caído. Está aquí, y no en una constante, para
	// que las pruebas no tengan que esperar segundos de verdad.
	esperaBase time.Duration

	now func() time.Time
}

// NewService construye el servicio y **enciende la cola**: los dos trabajadores empiezan a escuchar
// desde aquí, y `Parar` los apaga.
func NewService(repositorio *repositories.AIRepository, ajustes Ajustes) *Service {
	return nuevoServicio(repositorio, ajustes)
}

// nuevoServicio es el mismo constructor con el almacén detrás de una interfaz: es el que usan las
// pruebas.
func nuevoServicio(repositorio Store, ajustes Ajustes) *Service {
	ajustes = conValoresDeFabrica(ajustes)

	ctx, apagar := context.WithCancel(context.Background())

	servicio := &Service{
		store:      repositorio,
		ajustes:    ajustes,
		motor:      nuevoMotor(ajustes.Espera),
		cola:       make(map[string]trabajo),
		aviso:      make(chan struct{}, 1),
		parar:      make(chan struct{}),
		ctx:        ctx,
		apagar:     apagar,
		esperaBase: esperaBaseDeFabrica,
		now:        time.Now,
	}

	for i := 0; i < trabajadores; i++ {
		servicio.enMarcha.Add(1)
		go servicio.trabajador()
	}

	return servicio
}

// SetFuente engancha quién vuelve a armar el texto de un ticket. Se llama al arrancar, antes de
// `Retomar`, y es opcional: sin fuente, la puesta al día sólo puede decir cuántos campos se quedaron
// a medias.
func (s *Service) SetFuente(fuente Fuente) { s.fuente = fuente }

// SetConfiguracion engancha de dónde sale el motor: **la configuración de la instalación**, con el
// respaldo del entorno que ya trae `Ajustes`. Se llama al arrancar y no hace falta volver a llamarlo:
// el módulo la relee en cada petición (docs/modules/ai.md).
func (s *Service) SetConfiguracion(configuracion Configuracion) { s.configuracion = configuracion }

// ajustesDeAhora resuelve el motor con lo que diga la configuración **en este momento**.
//
// Si no hay configuración, o no tiene motor puesto, o no se puede leer, devuelve los ajustes del
// entorno: son el respaldo para una instalación que ya lo tuviera configurado así. Una lectura que
// falla no puede dejar sin motor a quien lo tenía.
func (s *Service) ajustesDeAhora() Ajustes {
	ajustes := s.ajustes
	if s.configuracion == nil {
		return ajustes
	}

	datos, err := s.configuracion.AI()
	if err != nil || !datos.Hay {
		return ajustes
	}

	ajustes.URL = strings.TrimSpace(datos.URL)
	if modelo := strings.TrimSpace(datos.Modelo); modelo != "" {
		ajustes.Modelo = modelo
	}
	ajustes.Provider = datos.Provider
	ajustes.Mode = datos.Mode
	ajustes.AuthType = datos.AuthType
	ajustes.AuthHeader = datos.AuthHeader
	ajustes.Credential = datos.Credential
	ajustes.Language = datos.Language

	return ajustes
}

// Pedir apunta un encargo y **vuelve enseguida**: no espera al motor (docs/modules/ai.md, decisión 2).
//
// Puede llamarse desde cualquier sitio y cuantas veces haga falta. Un encargo nuevo del mismo ticket y
// campo **sustituye al que estaba esperando**: el texto viejo ya no sirve, y el campo se vuelve a
// escribir en su misma fila (decisión 4).
func (s *Service) Pedir(entrada Entrada) {
	if entrada.Numero == "" {
		return
	}
	if entrada.Tipo != Motivo && entrada.Tipo != UltimaAccion {
		logs.LogWarning("el módulo ai recibió un campo que no conoce: " + string(entrada.Tipo))
		return
	}
	if strings.TrimSpace(entrada.Texto) == "" {
		// Sin texto no hay nada que resumir: pedirlo sería pedirle al modelo que se invente el ticket.
		logs.LogWarning("no hay texto que resumir para el ticket " + entrada.Numero + " (" + string(entrada.Tipo) + ")")
		return
	}

	// **Sin motor configurado no se encola nada**: el campo queda dicho en la base y la aplicación
	// entera sigue funcionando. Es una escritura local, no una espera al motor (decisión 2).
	if !s.ajustesDeAhora().Configurado() {
		s.apuntarFallo(entrada.Numero, entrada.Tipo, EstadoSinMotor, ClaveSinMotor)
		return
	}

	// El `pendiente` se apunta **antes** de encolar: si el backend se apaga entre las dos cosas, la
	// fila queda dicha para que la puesta al día del arranque la recoja (decisión 10).
	if err := s.store.MarcarPendiente(entrada.Numero, string(entrada.Tipo)); err != nil {
		logs.LogError("no se pudo apuntar el resumen pendiente del ticket " + entrada.Numero + ": " + err.Error())
	}

	s.encolar(trabajo{numero: entrada.Numero, tipo: entrada.Tipo, texto: entrada.Texto})
}

// De devuelve lo que hay de esos tickets, **en bloque**: una consulta por página y no una por fila
// (docs/modules/ai.md, decisión 9).
//
// Sólo salen los campos que tienen fila. Un ticket que nunca ha pedido resumen —uno de antes de que
// existiera el motor— no aparece en el mapa, y quien lo lea sabrá que no hay nada.
func (s *Service) De(numeros []string) (map[string]DosResumenes, error) {
	resumenes := make(map[string]DosResumenes, len(numeros))
	if len(numeros) == 0 {
		return resumenes, nil
	}

	filas, err := s.store.PorNumeros(numeros)
	if err != nil {
		return nil, err
	}

	for _, fila := range filas {
		dos := resumenes[fila.TicketNumber]

		resumen := Resumen{
			Estado:   Estado(fila.State),
			Es:       textoDe(fila.TextEs),
			En:       textoDe(fila.TextEn),
			ErrorKey: textoDe(fila.ErrorKey),
		}
		if fila.Text != nil {
			resumen.Text = *fila.Text
			resumen.Language = textoDe(fila.Language)
		} else if resumen.Es != "" || resumen.En != "" {
			// Compatibilidad de lectura durante la actualización: el idioma global decide cuál de las
			// dos columnas anteriores sale por el contrato nuevo.
			resumen.Language = s.ajustesDeAhora().Language
			if resumen.Language == "en" {
				resumen.Text = resumen.En
			} else {
				resumen.Text = resumen.Es
			}
		}
		if fila.GeneratedAt != nil {
			resumen.GeneradoEn = *fila.GeneratedAt
		}

		switch Tipo(fila.Kind) {
		case Motivo:
			dos.Motivo = resumen
		case UltimaAccion:
			dos.UltimaAccion = resumen
		}

		resumenes[fila.TicketNumber] = dos
	}

	return resumenes, nil
}

// Retomar es la puesta al día del arranque: la cola vive en memoria y se pierde al reiniciar, así que
// lo que quedó a medias se vuelve a pedir (docs/modules/ai.md, decisión 10).
//
// **No toca el estado ni el contador de intentos**: lo que estaba en `error` no vuelve a entrar —para
// eso está «Regenerar»—, y reiniciar el backend no puede ser la forma de saltarse el tope de
// intentos de un motor que no contesta.
func (s *Service) Retomar() {
	ajustes := s.ajustesDeAhora()
	if !ajustes.Configurado() {
		logs.LogInfo("el motor de IA no está configurado: no hay nada que retomar")
		return
	}
	pendientes, err := s.store.Pendientes(intentosMotor)
	if err != nil {
		logs.LogError("no se pudieron leer los resúmenes que quedaron a medias: " + err.Error())
		return
	}
	if len(pendientes) == 0 {
		logs.LogInfo("no hay resúmenes a medias que retomar")
		return
	}

	if s.fuente == nil {
		logs.LogWarning(
			"quedaron resúmenes a medias y no hay quién vuelva a armar su texto: se quedarán " +
				"pendientes hasta el próximo movimiento del ticket",
		)
		return
	}

	pedidos := 0
	for _, fila := range pendientes {
		texto, err := s.fuente.TextoParaElMotor(fila.TicketNumber)
		if err != nil {
			logs.LogWarning("no se pudo armar el texto del ticket " + fila.TicketNumber + " para retomar su resumen")
			continue
		}

		s.encolar(trabajo{numero: fila.TicketNumber, tipo: Tipo(fila.Kind), texto: texto})
		pedidos++
	}

	logs.LogInfo("puesta al día del motor de IA: se vuelven a pedir " + strconv.Itoa(pedidos) + " resúmenes")
}

// Disponible dice si hay motor configurado **y responde**.
//
// Se pregunta de verdad, con un tiempo corto: un motor caído no se nota hasta que se le pide algo, y
// «disponible» tiene que ser lo que pasa ahora y no lo que dice la configuración. La dirección sale
// de la configuración **en este momento**, así que cambiarla en Configuración vale al instante.
func (s *Service) Disponible() bool {
	ajustes := s.ajustesDeAhora()
	if !ajustes.Configurado() {
		return false
	}

	return s.motor.disponible(s.ctx, ajustes)
}

// Configurado indica si hay un motor efectivo sin hacer ninguna petición de red. Una caída temporal
// no cambia esta respuesta: la interfaz usa esta condición para ofrecer la ayuda de redacción y deja
// que el intento muestre el error dentro de su modal.
func (s *Service) Configurado() bool { return s.ajustesDeAhora().Configurado() }

// Traducir traduce un texto administrativo, como una plantilla de correo, con la configuración
// activa. Los marcadores protegidos los prepara y valida el módulo que conoce la plantilla.
func (s *Service) Traducir(texto, idioma string) (string, error) {
	ajustes := s.ajustesDeAhora()
	if !ajustes.Configurado() {
		return "", ErrSinMotor
	}
	ajustes.Language = idioma
	// Una plantilla puede ser bastante más larga que un resumen. El contexto sigue poniendo el
	// límite real; este tope evita recortarla con las cuarenta palabras de los tickets.
	ajustes.Palabras = 2000
	ctx, cancelar := context.WithTimeout(context.Background(), ajustes.Espera)
	defer cancelar()
	traducido, _, err := s.motor.pedirGlobal(ctx, ajustes, encargo{
		sistema: "Translate the supplied email template faithfully. Preserve every protected token exactly, including its position. Return no explanations.",
		texto:   texto,
	})
	return traducido, err
}

// MejorarRedaccion revisa un borrador ya autorizado por el módulo que conoce el ticket. No guarda
// entrada ni salida: devuelve texto plano para que la persona lo revise antes de decidir si lo usa.
func (s *Service) MejorarRedaccion(borrador, tono, destinatario, tipo, idioma string) (string, error) {
	ajustes := s.ajustesDeAhora()
	if !ajustes.Configurado() {
		return "", ErrSinMotor
	}
	ajustes.Language = idioma
	ajustes.Palabras = 2000

	tonalidades := map[string]string{
		"professional": "professional",
		"friendly":     "friendly and approachable",
		"brief":        "brief and direct",
		"empathetic":   "empathetic and respectful",
		"technical":    "precise and technical",
	}
	estilo, ok := tonalidades[tono]
	if !ok {
		return "", ErrRespuestaInvalida
	}

	clase := "main support ticket"
	if tipo == "internal" {
		clase = "internal ticket between Support and Development"
	}
	prompt := "Draft:\n" + borrador + "\n\nRecipient name: " + destinatario +
		"\nTicket type: " + clase + "\nTone: " + estilo
	ctx, cancelar := context.WithTimeout(context.Background(), ajustes.Espera)
	defer cancelar()
	mejorado, _, err := s.motor.pedirGlobal(ctx, ajustes, encargo{
		sistema: "Improve the supplied support message. Correct spelling and clarity while preserving every fact, name, number and technical detail. Do not invent information. Use the recipient name only when it sounds natural. Return plain text without Markdown, HTML, mentions or explanations.",
		texto:   prompt,
	})
	if err != nil {
		return "", err
	}
	mejorado = strings.TrimSpace(mejorado)
	if mejorado == "" {
		return "", ErrRespuestaInvalida
	}
	return mejorado, nil
}

func (s *Service) EsProveedorComercial() bool { return s.ajustesDeAhora().Mode == "provider" }

// Probar comprueba que el motor de esa dirección contesta, para el botón de «Probar la conexión» de
// Configuración.
//
// Es la misma pregunta que `Disponible` —su comprobación de salud— pero con un error en vez de un
// booleano, y **con la dirección que se le pasa**: la prueba es de lo que hay en pantalla, y no de lo
// que esté guardado (docs/modules/ai.md).
func (s *Service) Probar(url string) error {
	direccion := strings.TrimSpace(url)
	if direccion == "" {
		return ErrSinMotor
	}

	ctx, cancelar := context.WithTimeout(context.Background(), esperaDeLaPrueba)
	defer cancelar()

	return s.motor.probar(ctx, direccion)
}

// ProbarConfiguracion realiza una generación real artificial para validar protocolo, credencial,
// modelo y formato antes de activar una integración.
func (s *Service) ProbarConfiguracion(url, model, provider, authType, authHeader, credential, language string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.ajustes.Espera)
	defer cancel()
	return s.motor.probarConfiguracion(ctx, url, model, provider, authType, authHeader, credential, language)
}

// Parar apaga la cola: los trabajadores terminan lo que están haciendo y se van, y la llamada al
// motor que esté en vuelo se corta. Sin esto, apagar el backend dejaría goroutines hablando con un
// contenedor que puede no estar.
func (s *Service) Parar() {
	s.mu.Lock()
	if s.cerrada {
		s.mu.Unlock()
		return
	}
	s.cerrada = true
	s.mu.Unlock()

	s.apagar()
	close(s.parar)
	s.enMarcha.Wait()
}

// apuntarFallo deja el estado y su clave en la base.
//
// No revienta si la base falla: los dos resúmenes son un adorno y el ticket es el ticket.
func (s *Service) apuntarFallo(numero string, tipo Tipo, estado Estado, clave string) {
	var err error

	switch estado {
	case EstadoError:
		err = s.store.MarcarError(numero, string(tipo), clave)
	default:
		err = s.store.MarcarSinMotor(numero, string(tipo), clave)
	}

	if err != nil {
		logs.LogError("no se pudo apuntar el resumen del ticket " + numero + " (" + string(tipo) + "): " + err.Error())
	}
}

// conValoresDeFabrica rellena lo que no venga puesto.
func conValoresDeFabrica(ajustes Ajustes) Ajustes {
	if ajustes.Palabras <= 0 {
		ajustes.Palabras = palabrasDeFabrica
	}
	if ajustes.Contexto <= 0 {
		ajustes.Contexto = contextoDeFabrica
	}
	if ajustes.Espera <= 0 {
		ajustes.Espera = esperaDeFabrica
	}

	return ajustes
}

// textoDe lee una columna que puede estar nula.
func textoDe(valor *string) string {
	if valor == nil {
		return ""
	}

	return *valor
}
