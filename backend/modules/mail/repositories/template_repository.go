// Package repositories es el acceso a datos del módulo mail.
package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type LanguageTemplate struct { Key, Subject, Body string }

// ApplyLanguage updates all destination templates and the global language in one transaction.
func (r *TemplateRepository) ApplyLanguage(language string, templates []LanguageTemplate, actorID *int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, template := range templates {
			result := tx.Model(&Template{}).Where("key = ? AND language = ?", template.Key, language).
				Updates(map[string]any{"subject":template.Subject,"body":template.Body,"updated_at":time.Now(),"updated_by_id":actorID})
			if result.Error != nil { return result.Error }
			if result.RowsAffected != 1 { return ErrTemplateNotFound }
		}
		return tx.Table("installation_settings").Where("id = 1").Updates(map[string]any{
			"language": language, "settings_version": gorm.Expr("settings_version + 1"),
			"updated_at": time.Now(), "updated_by_id": actorID,
		}).Error
	})
}

// ErrTemplateNotFound se devuelve cuando no existe la plantilla pedida.
var ErrTemplateNotFound = errors.New("mail.template.notFound")

// Template es una de las veinte plantillas: un correo, en un idioma.
//
// Cada fila guarda el texto que se usa y el de fábrica. Los de fábrica no se editan nunca: son lo
// que permite restaurar un texto sin tenerlos también en el código, que sería la otra forma de
// acabar con dos versiones que dicen cosas distintas (docs/modules/mail.md, sección 3).
type Template struct {
	ID             int64     `gorm:"primaryKey"`
	Key            string    `gorm:"column:key"`
	Language       string    `gorm:"column:language"`
	Subject        string    `gorm:"column:subject"`
	Body           string    `gorm:"column:body"`
	DefaultSubject string    `gorm:"column:default_subject"`
	DefaultBody    string    `gorm:"column:default_body"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
	UpdatedByID    *int64    `gorm:"column:updated_by_id"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (Template) TableName() string { return "mail_templates" }

// Edited dice si el texto de hoy es distinto del de fábrica, que es la diferencia entre «restaurar»
// y «no hay nada que restaurar».
//
// Se compara el texto y no se mira `updated_by_id`, que sería lo cómodo: la cuenta de fábrica no
// está en la tabla de cuentas, así que sus ediciones se guardan sin identificador y aparecerían como
// «sin editar». El texto no engaña (docs/modules/mail.md, sección 3).
func (t Template) Edited() bool {
	return t.Subject != t.DefaultSubject || t.Body != t.DefaultBody
}

// TemplateRepository lee y escribe las plantillas.
type TemplateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository construye el repositorio.
func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// FindAll devuelve las veinte plantillas, ordenadas por clave y idioma.
func (r *TemplateRepository) FindAll() ([]Template, error) {
	var templates []Template
	err := r.db.Order("key, language").Find(&templates).Error
	return templates, err
}

// Find devuelve una plantilla concreta.
func (r *TemplateRepository) Find(key, language string) (Template, error) {
	var template Template

	err := r.db.Where("key = ? AND language = ?", key, language).First(&template).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Template{}, ErrTemplateNotFound
	}

	return template, err
}

// Update guarda el texto de una plantilla. No toca las columnas de fábrica: el texto de fábrica no
// se puede escribir desde la API.
func (r *TemplateRepository) Update(key, language, subject, body string, updatedByID *int64) error {
	result := r.db.Model(&Template{}).
		Where("key = ? AND language = ?", key, language).
		Updates(map[string]any{
			"subject":       subject,
			"body":          body,
			"updated_at":    time.Now(),
			"updated_by_id": updatedByID,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTemplateNotFound
	}

	return nil
}

// Reset devuelve una plantilla a su texto de fábrica.
func (r *TemplateRepository) Reset(key, language string) error {
	result := r.db.Exec(`
		UPDATE mail_templates
		   SET subject = default_subject,
		       body = default_body,
		       updated_at = now(),
		       updated_by_id = NULL
		 WHERE key = ? AND language = ?`, key, language)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTemplateNotFound
	}

	return nil
}
