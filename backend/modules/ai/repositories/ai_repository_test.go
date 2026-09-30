package repositories

import (
	"testing"
	"time"

	"catalina-support/backend/shared/config"
	"catalina-support/backend/shared/database"
)

// Las pruebas del repositorio van **contra la base de datos de verdad**, porque lo que hay que
// comprobar aquí es el SQL: que volver a pedir un campo reescribe su misma fila en vez de fallar o
// acumular, que la lista se lee en bloque y que la puesta al día deja fuera lo que está en `error`.
//
// Todo se hace **dentro de una transacción que se deshace al terminar**, así que no queda nada en la
// base, y sin base de datos a mano las pruebas **se saltan**: `go test ./...` tiene que seguir
// sirviendo para lo que no necesita contenedores.

const (
	numeroDePrueba = "PRUEBA-AI-0001"
	otroNumero     = "PRUEBA-AI-0002"
	vecinoNumero   = "PRUEBA-AI-0003"
)

func TestMarcarPendienteCreaLaFilaYLaReescribeSiSeVuelveAPedir(t *testing.T) {
	repositorio := repositorioDePrueba(t)

	if err := repositorio.MarcarPendiente(numeroDePrueba, KindMotivo); err != nil {
		t.Fatalf("MarcarPendiente falló: %v", err)
	}

	fila := leerFila(t, repositorio, numeroDePrueba, KindMotivo)
	if fila.State != StatePendiente {
		t.Errorf("el estado es %q y se esperaba %q", fila.State, StatePendiente)
	}
	if fila.Attempts != 0 {
		t.Errorf("los intentos son %d y una petición nueva empieza en cero", fila.Attempts)
	}
	if fila.TextEs != nil || fila.TextEn != nil {
		t.Error("un campo pendiente no tiene texto que enseñar")
	}
	if fila.RequestedAt.IsZero() {
		t.Error("una petición tiene que quedar fechada")
	}

	// Se escribe el resumen y se le suman intentos, y después se vuelve a pedir: la petición nueva
	// **reescribe la misma fila** y deja el texto viejo fuera (docs/modules/ai.md, sección 4).
	if err := repositorio.GuardarResultado(numeroDePrueba, KindMotivo, "Uno", "One", "modelo", time.Now()); err != nil {
		t.Fatalf("GuardarResultado falló: %v", err)
	}
	if err := repositorio.SubirIntentos(numeroDePrueba, KindMotivo); err != nil {
		t.Fatalf("SubirIntentos falló: %v", err)
	}
	if err := repositorio.MarcarPendiente(numeroDePrueba, KindMotivo); err != nil {
		t.Fatalf("MarcarPendiente falló: %v", err)
	}

	fila = leerFila(t, repositorio, numeroDePrueba, KindMotivo)
	if fila.State != StatePendiente {
		t.Errorf("el estado es %q y se esperaba %q", fila.State, StatePendiente)
	}
	if fila.Attempts != 0 {
		t.Errorf("los intentos son %d y volver a pedirlo los pone a cero", fila.Attempts)
	}
	if fila.TextEs != nil || fila.TextEn != nil || fila.Model != nil || fila.GeneratedAt != nil {
		t.Error("el resumen viejo tiene que quedar fuera: cuenta una historia que ya no es la del ticket")
	}

	if filas := contarFilas(t, repositorio, numeroDePrueba); filas != 1 {
		t.Errorf("hay %d filas de ese ticket y tiene que haber una por campo: el campo se reescribe", filas)
	}
}

func TestGuardarResultadoGuardaLasDosRedaccionesYElModelo(t *testing.T) {
	repositorio := repositorioDePrueba(t)

	generado := time.Now().Truncate(time.Millisecond)

	if err := repositorio.GuardarResultado(numeroDePrueba, KindUltimaAccion, "Se pidió una captura.", "A screenshot was requested.", "qwen2.5-1.5b-instruct", generado); err != nil {
		t.Fatalf("GuardarResultado falló: %v", err)
	}

	fila := leerFila(t, repositorio, numeroDePrueba, KindUltimaAccion)
	if fila.State != StateListo {
		t.Errorf("el estado es %q y se esperaba %q", fila.State, StateListo)
	}
	if texto(fila.TextEs) != "Se pidió una captura." || texto(fila.TextEn) != "A screenshot was requested." {
		t.Errorf("las dos redacciones tienen que estar en la misma fila: es=%q en=%q", texto(fila.TextEs), texto(fila.TextEn))
	}
	if texto(fila.Model) != "qwen2.5-1.5b-instruct" {
		t.Errorf("el modelo guardado es %q", texto(fila.Model))
	}
	if fila.GeneratedAt == nil || !fila.GeneratedAt.Equal(generado) {
		t.Error("la fecha de generación no se guardó")
	}
	if fila.ErrorKey != nil {
		t.Errorf("un resumen listo no lleva clave de error, y lleva %q", texto(fila.ErrorKey))
	}
}

func TestMarcarErrorYSinMotorGuardanSuClaveYNoLaConfunden(t *testing.T) {
	repositorio := repositorioDePrueba(t)

	// Los dos fallos son distintos y se guardan distintos: uno es culpa del modelo y el otro de que el
	// contenedor no está (docs/modules/ai.md, decisión 8).
	if err := repositorio.MarcarError(numeroDePrueba, KindMotivo, "ai.invalid"); err != nil {
		t.Fatalf("MarcarError falló: %v", err)
	}
	if err := repositorio.MarcarSinMotor(otroNumero, KindMotivo, "ai.unavailable"); err != nil {
		t.Fatalf("MarcarSinMotor falló: %v", err)
	}

	enError := leerFila(t, repositorio, numeroDePrueba, KindMotivo)
	if enError.State != StateError || texto(enError.ErrorKey) != "ai.invalid" {
		t.Errorf("el campo quedó en %q con la clave %q", enError.State, texto(enError.ErrorKey))
	}

	sinMotor := leerFila(t, repositorio, otroNumero, KindMotivo)
	if sinMotor.State != StateSinMotor || texto(sinMotor.ErrorKey) != "ai.unavailable" {
		t.Errorf("el campo quedó en %q con la clave %q", sinMotor.State, texto(sinMotor.ErrorKey))
	}
}

func TestSubirIntentosSumaUnoYNoReVientaSinFila(t *testing.T) {
	repositorio := repositorioDePrueba(t)

	if err := repositorio.MarcarPendiente(numeroDePrueba, KindMotivo); err != nil {
		t.Fatalf("MarcarPendiente falló: %v", err)
	}

	for i := 0; i < 3; i++ {
		if err := repositorio.SubirIntentos(numeroDePrueba, KindMotivo); err != nil {
			t.Fatalf("SubirIntentos falló: %v", err)
		}
	}

	if intentos := leerFila(t, repositorio, numeroDePrueba, KindMotivo).Attempts; intentos != 3 {
		t.Errorf("el contador de intentos quedó en %d y se esperaba 3", intentos)
	}

	// Un campo del que no hay fila no puede inventarse una: el contador va con la fila.
	if err := repositorio.SubirIntentos(vecinoNumero, KindMotivo); err != nil {
		t.Fatalf("SubirIntentos sin fila devolvió un error: %v", err)
	}
	if filas := contarFilas(t, repositorio, vecinoNumero); filas != 0 {
		t.Errorf("se crearon %d filas al contar un intento que no era de nadie", filas)
	}
}

func TestPorNumerosTraeVariosTicketsDeUnaVez(t *testing.T) {
	repositorio := repositorioDePrueba(t)

	if err := repositorio.GuardarResultado(numeroDePrueba, KindMotivo, "Uno", "One", "modelo", time.Now()); err != nil {
		t.Fatalf("GuardarResultado falló: %v", err)
	}
	if err := repositorio.GuardarResultado(numeroDePrueba, KindUltimaAccion, "Dos", "Two", "modelo", time.Now()); err != nil {
		t.Fatalf("GuardarResultado falló: %v", err)
	}
	if err := repositorio.GuardarResultado(otroNumero, KindMotivo, "Tres", "Three", "modelo", time.Now()); err != nil {
		t.Fatalf("GuardarResultado falló: %v", err)
	}
	// El vecino no se pide: no puede salir en la respuesta.
	if err := repositorio.GuardarResultado(vecinoNumero, KindMotivo, "Cuatro", "Four", "modelo", time.Now()); err != nil {
		t.Fatalf("GuardarResultado falló: %v", err)
	}

	filas, err := repositorio.PorNumeros([]string{numeroDePrueba, otroNumero})
	if err != nil {
		t.Fatalf("PorNumeros falló: %v", err)
	}

	if len(filas) != 3 {
		t.Fatalf("se leyeron %d filas y se esperaban 3 (los dos campos de un ticket y uno del otro)", len(filas))
	}

	for _, fila := range filas {
		if fila.TicketNumber == vecinoNumero {
			t.Error("salió un ticket que no se pidió")
		}
	}

	// Sin números no se pregunta nada.
	if vacias, err := repositorio.PorNumeros(nil); err != nil || len(vacias) != 0 {
		t.Errorf("sin números la consulta trae %d filas (err=%v)", len(vacias), err)
	}
}

func TestPendientesDejaFueraLoQueEstaEnErrorYLoQueAgotoSusIntentos(t *testing.T) {
	repositorio := repositorioDePrueba(t)

	// Lo que se retoma: lo que quedó a medias y lo que no agotó sus intentos.
	if err := repositorio.MarcarPendiente(numeroDePrueba, KindMotivo); err != nil {
		t.Fatalf("MarcarPendiente falló: %v", err)
	}
	if err := repositorio.MarcarSinMotor(otroNumero, KindMotivo, "ai.unavailable"); err != nil {
		t.Fatalf("MarcarSinMotor falló: %v", err)
	}
	if err := repositorio.SubirIntentos(otroNumero, KindMotivo); err != nil {
		t.Fatalf("SubirIntentos falló: %v", err)
	}

	// Lo que no se retoma: una respuesta que no valía, un motor caído que ya agotó sus intentos y un
	// resumen ya escrito.
	if err := repositorio.MarcarError(vecinoNumero, KindMotivo, "ai.invalid"); err != nil {
		t.Fatalf("MarcarError falló: %v", err)
	}
	if err := repositorio.MarcarSinMotor("PRUEBA-AI-0004", KindMotivo, "ai.unavailable"); err != nil {
		t.Fatalf("MarcarSinMotor falló: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := repositorio.SubirIntentos("PRUEBA-AI-0004", KindMotivo); err != nil {
			t.Fatalf("SubirIntentos falló: %v", err)
		}
	}
	if err := repositorio.GuardarResultado("PRUEBA-AI-0005", KindMotivo, "Cinco", "Five", "modelo", time.Now()); err != nil {
		t.Fatalf("GuardarResultado falló: %v", err)
	}

	// **El 3 es el tope de intentos, no un tope de filas**: `Pendientes` devuelve los campos a medias de
	// **toda la instalación**, y en la base de desarrollo hay resúmenes de tickets de verdad.
	pendientes, err := repositorio.Pendientes(3)
	if err != nil {
		t.Fatalf("Pendientes falló: %v", err)
	}

	porNumero := make(map[string]string, len(pendientes))
	for _, fila := range pendientes {
		porNumero[fila.TicketNumber] = fila.State
	}

	// **Y se cuentan sólo los números de esta prueba**: los de la aplicación no son asunto suyo, y
	// contarlos todos hacía que fallara según lo que estuviera pasando en el entorno —crear un ticket
	// por la API deja dos campos en `pendiente` unos segundos—, que es un fallo de la prueba y no del
	// código (hallazgo del 2026-09-27).
	mios := []string{numeroDePrueba, otroNumero, vecinoNumero, "PRUEBA-AI-0004", "PRUEBA-AI-0005"}
	retomados := 0
	for _, numero := range mios {
		if _, hay := porNumero[numero]; hay {
			retomados++
		}
	}
	if retomados != 2 {
		t.Fatalf("se retoman %d campos de esta prueba y se esperaban 2: %v", retomados, porNumero)
	}

	if porNumero[numeroDePrueba] != StatePendiente {
		t.Errorf("el campo pendiente no entra en la puesta al día: %v", porNumero)
	}
	if porNumero[otroNumero] != StateSinMotor {
		t.Errorf("el campo con pocos intentos no entra en la puesta al día: %v", porNumero)
	}
	for _, numero := range []string{vecinoNumero, "PRUEBA-AI-0004", "PRUEBA-AI-0005"} {
		if _, hay := porNumero[numero]; hay {
			t.Errorf("el campo %s no puede retomarse solo: %v", numero, porNumero)
		}
	}
}

// ============================================================================
// Utilidades
// ============================================================================

// repositorioDePrueba abre la base de datos del proyecto y devuelve el repositorio trabajando dentro
// de una transacción que se deshace al acabar la prueba.
func repositorioDePrueba(t *testing.T) *AIRepository {
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

	if !tx.Migrator().HasTable(&AIInsight{}) {
		_ = tx.Rollback()
		t.Skip("la tabla ai_insights no está en esta base: aplica backend/migrations/v1.0.0.sql")
	}

	t.Cleanup(func() {
		_ = tx.Rollback()

		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	return NewAIRepository(tx)
}

func leerFila(t *testing.T, repositorio *AIRepository, numero, tipo string) AIInsight {
	t.Helper()

	var filas []AIInsight
	if err := repositorio.db.Where("ticket_number = ? AND kind = ?", numero, tipo).Find(&filas).Error; err != nil {
		t.Fatalf("no se pudo leer la fila: %v", err)
	}
	if len(filas) != 1 {
		t.Fatalf("hay %d filas de %s/%s y tiene que haber una", len(filas), numero, tipo)
	}

	return filas[0]
}

func contarFilas(t *testing.T, repositorio *AIRepository, numero string) int64 {
	t.Helper()

	var total int64
	if err := repositorio.db.Model(&AIInsight{}).Where("ticket_number = ?", numero).Count(&total).Error; err != nil {
		t.Fatalf("no se pudieron contar las filas: %v", err)
	}

	return total
}

func texto(valor *string) string {
	if valor == nil {
		return ""
	}

	return *valor
}
