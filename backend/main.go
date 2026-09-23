// Catalina-Support: backend de una mesa de ayuda de dos niveles.
//
// El arranque hace, en este orden: leer configuración, comprobar que se puede escribir
// el log, conectar con la base de datos, montar las rutas y escuchar. Cualquier fallo
// de esos pasos detiene el arranque: es preferible no arrancar a arrancar a medias.
//
// Las rutas de los módulos se registran con el prefijo /api/<módulo>/... y nginx las
// pasa tal cual (no las reescribe), así que el backend es el dueño de ese prefijo.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juanky-estevez/go-logs"
	"gorm.io/gorm"

	"catalina-support/backend/shared/config"
	"catalina-support/backend/shared/database"
	"catalina-support/backend/shared/httpx"
	"catalina-support/backend/shared/middleware"
)

const shutdownTimeout = 15 * time.Second

func main() {
	if err := run(); err != nil {
		logs.LogError(err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// go-logs crea la carpeta por su cuenta y, si no puede escribir, NO falla: avisa en
	// rojo y sigue escribiendo sólo en terminal. Un LOGS_FOLDER mal puesto perdería el
	// log sin detener nada, así que aquí se comprueba de verdad.
	if err := os.MkdirAll(cfg.LogsFolder, 0o755); err != nil {
		return fmt.Errorf("no se puede escribir en LOGS_FOLDER (%s): %w", cfg.LogsFolder, err)
	}
	if probe, err := os.CreateTemp(cfg.LogsFolder, ".probe-*"); err != nil {
		return fmt.Errorf("no se puede escribir en LOGS_FOLDER (%s): %w", cfg.LogsFolder, err)
	} else {
		probe.Close()
		os.Remove(probe.Name())
	}

	logs.LogInfo("arrancando el backend en el entorno " + cfg.Environment)

	db, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("no se pudo conectar a la base de datos: %w", err)
	}
	logs.LogSuccess("conexión a la base de datos establecida")

	// Todavía no hay módulos: la lista de rutas sólo tiene la comprobación de salud.
	// Cada módulo añade las suyas cuando su documento esté aprobado.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health(db))

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           middleware.Chain(mux, middleware.Recover),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// El servidor escucha en una goroutine para que main pueda esperar la señal de parada.
	serverErrors := make(chan error, 1)
	go func() {
		logs.LogSuccess("escuchando en el puerto " + cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("el servidor se detuvo: %w", err)
	case signalReceived := <-stop:
		logs.LogWarning("señal recibida (" + signalReceived.String() + "): cerrando")
	}

	// Se deja terminar las peticiones en curso antes de salir.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("no se pudo cerrar el servidor con orden: %w", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("no se pudo cerrar la conexión a la base de datos: %w", err)
		}
	}

	logs.LogSuccess("backend detenido")
	return nil
}

// health comprueba que el proceso responde y que la base de datos contesta de verdad.
func health(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := db.DB()
		if err != nil {
			logs.LogError("health: no se pudo obtener la conexión a la base de datos: " + err.Error())
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "database": "error"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			logs.LogError("health: la base de datos no responde: " + err.Error())
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "database": "error"})
			return
		}

		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
	}
}
