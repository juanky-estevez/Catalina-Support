// Package database crea y configura la conexión única a PostgreSQL.
package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"catalina-support/backend/shared/config"
)

// Connect abre la conexión, la configura y comprueba que la base responde.
//
// El esquema NO se crea aquí: la única vía es backend/migrations (docs/arquitectura.md,
// sección 7). Por eso no se usa AutoMigrate en ningún sitio.
func Connect(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		// GORM escribe por su cuenta, así que se limita a los errores: los avisos y la
		// información de la aplicación van por go-logs.
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
