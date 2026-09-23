// Package config lee la configuración del backend desde el entorno.
//
// Sólo se leen variables que cambian entre entornos. Lo que vale lo mismo en
// desarrollo y en producción se escribe como constante en el código, no como variable
// (docs/arquitectura.md, sección 8).
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config es la configuración del backend ya validada.
type Config struct {
	// Environment lo consume go-logs para nombrar el archivo y cada línea.
	Environment string

	// AppPort es el puerto en el que escucha el servidor HTTP.
	AppPort string

	// LogsFolder es la carpeta donde go-logs escribe los archivos del día.
	LogsFolder string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresName     string
}

// Load lee el entorno y falla si falta algo obligatorio: es preferible no arrancar a
// arrancar a medias con una configuración incompleta.
func Load() (Config, error) {
	cfg := Config{
		Environment:      get("ENVIRONMENT", "dev"),
		AppPort:          get("APP_PORT", "11002"),
		LogsFolder:       get("LOGS_FOLDER", "/logs"),
		PostgresHost:     get("POSTGRES_HOST", "database"),
		PostgresPort:     get("PGPORT", "11003"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresName:     os.Getenv("POSTGRES_DB"),
	}

	var missing []string
	if cfg.PostgresUser == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if cfg.PostgresPassword == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if cfg.PostgresName == "" {
		missing = append(missing, "POSTGRES_DB")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("faltan variables de entorno obligatorias: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

// DSN es la cadena de conexión de GORM.
//
// sslmode=disable va fijo: la base de datos vive en la red interna de Compose y no
// habla TLS. Si algún día sale de ahí, se convierte en variable de entorno.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresName,
	)
}

func get(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
