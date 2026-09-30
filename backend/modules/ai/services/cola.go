package services

import (
	"errors"
	"time"

	"github.com/juanky-estevez/go-logs"
)

// trabajo es lo que espera en la cola: un ticket, un campo y el texto que hay que leer.
type trabajo struct {
	numero string
	tipo   Tipo
	texto  string
}

// clave identifica un trabajo: **uno por ticket y campo**. Pedir dos veces lo mismo no encola dos
// veces, y el encargo nuevo sustituye al que esperaba porque el texto viejo ya no sirve
// (docs/modules/ai.md, sección 6).
func (t trabajo) clave() string { return t.numero + "|" + string(t.tipo) }

// encolar deja el trabajo esperando y despierta a un trabajador.
//
// **No bloquea nunca**, y por eso el aviso es una señal y no el trabajo: si ya hay uno pendiente, el
// trabajador vaciará la cola entera cuando llegue, y quien pide un resumen no espera a nadie.
func (s *Service) encolar(t trabajo) {
	s.mu.Lock()
	if s.cerrada {
		s.mu.Unlock()
		return
	}
	s.cola[t.clave()] = t
	s.mu.Unlock()

	select {
	case s.aviso <- struct{}{}:
	default:
	}
}

// siguiente saca el trabajo que toque.
//
// El orden no importa —son resúmenes sueltos y ninguno corre más que otro— y por eso no se mantiene
// una lista aparte del mapa: lo que importa es que no haya dos encargos del mismo campo a la vez.
func (s *Service) siguiente() (trabajo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for clave, t := range s.cola {
		delete(s.cola, clave)
		return t, true
	}

	return trabajo{}, false
}

// trabajador vacía la cola mientras haya trabajo.
func (s *Service) trabajador() {
	defer s.enMarcha.Done()

	for {
		select {
		case <-s.parar:
			return
		case <-s.aviso:
		}

		for {
			t, hay := s.siguiente()
			if !hay {
				break
			}

			s.atender(t)
		}
	}
}

// atender le pide un campo al motor, con sus reintentos.
//
// Los dos fallos no se tratan igual, y es la decisión 8 tal cual: **un motor que no contesta se
// reintenta tres veces con espera creciente** —es un contenedor que puede estar arrancando, o lento—
// y **una respuesta que no vale se reintenta una sola vez** y después el campo queda en `error`,
// porque volver a preguntarle lo mismo a un modelo que ya contestó mal es pedirle que se repita.
func (s *Service) atender(t trabajo) {
	for intento := 1; ; intento++ {
		es, en, err := s.motor.pedir(s.ctx, s.ajustes, encargoDe(t.tipo, t.texto, s.ajustes.Palabras))

		if err == nil {
			s.contarIntento(t)
			s.guardar(t, es, en)
			return
		}

		// Un intento cortado por el apagado no es un intento del motor, y no se cuenta ni se apunta:
		// el campo se queda como estaba y la puesta al día del próximo arranque lo recogerá.
		if s.apagando() {
			return
		}

		s.contarIntento(t)

		if errors.Is(err, ErrRespuestaInvalida) {
			if intento >= intentosRespuesta {
				s.apuntarFallo(t.numero, t.tipo, EstadoError, ClaveInvalida)
				// **Se registra la forma, nunca el contenido**: qué claves llegaron y cuánto miden. Sin
				// esto, «contestó algo que no vale» no se puede investigar; con esto, se sabe si faltó el
				// inglés, si vino vacío o si se cortó a mitad.
				logs.LogWarning("el motor de IA contestó algo que no vale para el ticket " +
					t.numero + " (" + string(t.tipo) + "): " + formaDe(err))
				return
			}

			// El reintento es inmediato: no es un motor lento, es un modelo que se ha explicado mal.
			continue
		}

		if intento >= intentosMotor {
			s.apuntarFallo(t.numero, t.tipo, EstadoSinMotor, ClaveSinMotor)
			logs.LogWarning("el motor de IA no contesta para el ticket " + t.numero + " (" + string(t.tipo) + ")")
			return
		}

		// La espera crece con cada intento: al motor que se ha quedado sin memoria hay que darle algo
		// más que al que ha tenido un hipo.
		if !s.esperar(s.esperaBase * time.Duration(intento)) {
			return
		}
	}
}

// guardar escribe las dos redacciones en la fila del campo.
func (s *Service) guardar(t trabajo, es, en string) {
	err := s.store.GuardarResultado(t.numero, string(t.tipo), es, en, s.ajustes.Modelo, s.now())
	if err != nil {
		logs.LogError("no se pudo guardar el resumen del ticket " + t.numero + " (" + string(t.tipo) + "): " + err.Error())
		return
	}

	// Se registran el número, el campo y el estado, y **nunca el texto del ticket** ni el que escribe
	// el motor: los logs son para saber qué pasó, no para leer los tickets de nadie.
	logs.LogSuccess("resumen escrito para el ticket " + t.numero + " (" + string(t.tipo) + ")")
}

// contarIntento suma uno al contador del campo. Es lo que corta el reintento infinito entre reinicios
// del backend (decisión 8).
// formaDe describe **sin el contenido** por qué no valió una respuesta: qué claves llegaron y
// cuánto midió cada una.
//
// Es lo que hace investigable el «contestó algo que no vale»: si falta el inglés, si llegó vacío o si
// el modelo se quedó a medias. El texto del ticket **no** se registra nunca.
func formaDe(err error) string {
	var deRespuesta *respuestaInvalida
	if errors.As(err, &deRespuesta) {
		return deRespuesta.forma
	}

	return "sin detalle"
}

func (s *Service) contarIntento(t trabajo) {
	if err := s.store.SubirIntentos(t.numero, string(t.tipo)); err != nil {
		logs.LogError("no se pudo contar el intento del resumen del ticket " + t.numero + ": " + err.Error())
	}
}

// esperar duerme entre intentos y **se despierta si el backend se está apagando**: apagar no puede
// quedarse esperando a que venza una espera.
func (s *Service) esperar(espera time.Duration) bool {
	temporizador := time.NewTimer(espera)
	defer temporizador.Stop()

	select {
	case <-temporizador.C:
		return true
	case <-s.parar:
		return false
	}
}

// apagando dice si el servicio está cerrando.
func (s *Service) apagando() bool {
	select {
	case <-s.parar:
		return true
	default:
		return false
	}
}
