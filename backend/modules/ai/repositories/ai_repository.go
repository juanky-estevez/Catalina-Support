// Package repositories es el acceso a datos del módulo ai: la tabla de los dos resúmenes que redacta
// el motor de IA (docs/modules/ai.md, sección 4).
//
// **Sólo se escribe en `ai_insights`**: este módulo no escribe en las tablas de nadie más y nadie
// escribe en la suya (docs/modules/ai.md, sección 5).
package repositories

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Los dos campos que se le piden al motor. Son un valor cerrado y viven aquí para que la base y el
// código digan lo mismo, como los estados de un ticket.
const (
	KindMotivo       = "motivo"
	KindUltimaAccion = "ultima_accion"
)

// Los cuatro estados de un campo.
//
// **Los dos fallos están separados a propósito**: `error` es una respuesta que no valía —culpa del
// modelo— y `sin_motor` es que hay motor configurado pero no contesta. No significan lo mismo ni se
// arreglan igual, así que no pueden compartir estado (docs/modules/ai.md, decisión 8).
const (
	StatePendiente = "pendiente"
	StateListo     = "listo"
	StateError     = "error"
	StateSinMotor  = "sin_motor"
)

// AIInsight es una fila de `ai_insights`: **un campo de un ticket**, con sus dos redacciones.
//
// Las columnas que la base admite nulas van con puntero, como en el resto del proyecto: así se
// distingue «no hay texto» de «el texto está vacío», y un campo `pendiente` se lee como lo que es.
type AIInsight struct {
	ID           int64  `gorm:"primaryKey"`
	TicketNumber string `gorm:"column:ticket_number"`
	Kind         string `gorm:"column:kind"`
	State        string `gorm:"column:state"`
	// Las dos redacciones, en la misma fila: el motor las devuelve juntas (decisión 3).
	TextEs *string `gorm:"column:text_es"`
	TextEn *string `gorm:"column:text_en"`
	// La clave del error (`ai.unavailable` o `ai.invalid`) cuando el estado no es `listo`.
	ErrorKey *string `gorm:"column:error_key"`
	// El modelo que lo escribió: dice con qué se generó.
	Model       *string    `gorm:"column:model"`
	Attempts    int        `gorm:"column:attempts"`
	RequestedAt time.Time  `gorm:"column:requested_at"`
	GeneratedAt *time.Time `gorm:"column:generated_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (AIInsight) TableName() string { return "ai_insights" }

// AIRepository lee y escribe los resúmenes del motor.
type AIRepository struct {
	db *gorm.DB
}

// NewAIRepository construye el repositorio.
func NewAIRepository(db *gorm.DB) *AIRepository {
	return &AIRepository{db: db}
}

// MarcarPendiente deja el campo apuntado como `pendiente` antes de encolarlo.
//
// Es un `INSERT ... ON CONFLICT`, y no un `UPDATE`: la fila del campo puede no existir todavía —es la
// primera vez que se pide para ese ticket— y volver a pedirlo **reescribe la misma fila** en lugar de
// acumular historial, que es lo que fija la clave única `(ticket_number, kind)` (sección 4).
//
// **El texto viejo se borra al volver a pedirlo**, y los intentos vuelven a cero: el resumen de un
// ticket que ya se ha movido cuenta una historia que no es la de ahora, enseñarlo mientras se
// escribe sería enseñar algo falso, y una petición nueva es una cuenta nueva.
func (r *AIRepository) MarcarPendiente(numero, tipo string) error {
	ahora := time.Now()

	fila := AIInsight{
		TicketNumber: numero,
		Kind:         tipo,
		State:        StatePendiente,
		RequestedAt:  ahora,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "ticket_number"}, {Name: "kind"}},
		DoUpdates: clause.Assignments(map[string]any{
			"state":        StatePendiente,
			"text_es":      nil,
			"text_en":      nil,
			"error_key":    nil,
			"model":        nil,
			"attempts":     0,
			"requested_at": ahora,
			"generated_at": nil,
		}),
	}).Create(&fila).Error
}

// GuardarResultado guarda las dos redacciones y deja el campo en `listo`.
//
// Los intentos **no se tocan** aquí: los lleva `SubirIntentos`, que es quien cuenta las llamadas al
// motor, y dejarlos a cero al guardar borraría esa cuenta.
func (r *AIRepository) GuardarResultado(numero, tipo, es, en, modelo string, generadoEn time.Time) error {
	fila := AIInsight{
		TicketNumber: numero,
		Kind:         tipo,
		State:        StateListo,
		TextEs:       &es,
		TextEn:       &en,
		Model:        &modelo,
		RequestedAt:  generadoEn,
		GeneratedAt:  &generadoEn,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "ticket_number"}, {Name: "kind"}},
		DoUpdates: clause.Assignments(map[string]any{
			"state":        StateListo,
			"text_es":      es,
			"text_en":      en,
			"error_key":    nil,
			"model":        modelo,
			"generated_at": generadoEn,
		}),
	}).Create(&fila).Error
}

// MarcarError deja el campo en `error`: el motor contestó, pero lo que contestó no vale.
func (r *AIRepository) MarcarError(numero, tipo, clave string) error {
	return r.marcarFallo(numero, tipo, StateError, clave)
}

// MarcarSinMotor deja el campo en `sin_motor`: hay motor configurado pero no contesta. Es el estado
// de una instalación sin motor y el de un contenedor caído, que para quien lee son lo mismo.
func (r *AIRepository) MarcarSinMotor(numero, tipo, clave string) error {
	return r.marcarFallo(numero, tipo, StateSinMotor, clave)
}

// SubirIntentos suma uno al contador del campo.
//
// Es lo que corta el reintento infinito: sin este número, un contenedor caído se reintentaría para
// siempre cada vez que el backend arranca (decisión 8).
func (r *AIRepository) SubirIntentos(numero, tipo string) error {
	return r.db.Model(&AIInsight{}).
		Where("ticket_number = ? AND kind = ?", numero, tipo).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
}

// PorNumeros trae los resúmenes de varios tickets **de una vez**: es lo que hace que la lista se
// pinte con una sola consulta y no con una por fila (decisión 9).
func (r *AIRepository) PorNumeros(numeros []string) ([]AIInsight, error) {
	if len(numeros) == 0 {
		return nil, nil
	}

	var filas []AIInsight
	if err := r.db.Where("ticket_number IN ?", numeros).Find(&filas).Error; err != nil {
		return nil, err
	}

	return filas, nil
}

// Pendientes son los campos que quedaron a medias y se pueden volver a pedir: los `pendiente` de una
// cola que se perdió al reiniciar el backend y los `sin_motor` que **no han agotado sus intentos**
// (decisión 10).
//
// **Los `error` no salen de aquí**, y es a propósito: una respuesta que no valía no se reintenta sola
// —para eso está el botón de regenerar—, porque sería volver a preguntarle lo mismo a un modelo que
// ya contestó mal.
//
// El orden es por `requested_at`: lo que lleva más tiempo esperando se retoma primero.
func (r *AIRepository) Pendientes(maxIntentos int) ([]AIInsight, error) {
	var filas []AIInsight

	err := r.db.
		Where(
			"state = ? OR (state = ? AND attempts < ?)",
			StatePendiente, StateSinMotor, maxIntentos,
		).
		Order("requested_at ASC").
		Find(&filas).Error
	if err != nil {
		return nil, err
	}

	return filas, nil
}

// marcarFallo apunta el estado y su clave. Las dos redacciones se vacían: un campo que no está
// `listo` no tiene texto que enseñar, y dejar el viejo sería enseñar un resumen que ya no es el del
// ticket.
func (r *AIRepository) marcarFallo(numero, tipo, estado, clave string) error {
	fila := AIInsight{
		TicketNumber: numero,
		Kind:         tipo,
		State:        estado,
		ErrorKey:     &clave,
		RequestedAt:  time.Now(),
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "ticket_number"}, {Name: "kind"}},
		DoUpdates: clause.Assignments(map[string]any{
			"state":        estado,
			"error_key":    clave,
			"text_es":      nil,
			"text_en":      nil,
			"model":        nil,
			"generated_at": nil,
		}),
	}).Create(&fila).Error
}
