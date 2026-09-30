// Package services es el módulo ai: los dos resúmenes que redacta el motor de inteligencia
// artificial —**el motivo del ticket** y **su última acción**— en español y en inglés
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

// Resumen es lo que hay de un campo: su estado y sus dos redacciones.
//
// **Los dos idiomas viajan juntos** porque el motor los devuelve en la misma respuesta (decisión 3), y
// porque la pantalla enseña el del idioma de quien mira: traducir al vuelo sería pedir dos veces lo
// mismo.
type Resumen struct {
	Estado     Estado
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
	Espera time.Duration
}

// Configurado dice si hay motor al que preguntar.
//
// No es un fallo que no lo haya: es una instalación que no lo tiene puesto, y esa es la condición
// principal de este módulo (docs/modules/ai.md, decisión 2).
func (a Ajustes) Configurado() bool { return strings.TrimSpace(a.URL) != "" }

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
	if !s.ajustes.Configurado() {
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
	if !s.ajustes.Configurado() {
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
// «disponible» tiene que ser lo que pasa ahora y no lo que dice la configuración.
func (s *Service) Disponible() bool {
	if !s.ajustes.Configurado() {
		return false
	}

	return s.motor.disponible(s.ctx, s.ajustes)
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
