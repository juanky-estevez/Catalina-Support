// Package repositories es el acceso a datos del módulo auth: los tokens de los enlaces de
// contraseña.
package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Los dos propósitos de un enlace, que es un valor cerrado en la base.
const (
	PurposeInvitation = "alta"
	PurposeRecovery   = "recuperacion"
)

// ErrTokenNotFound se devuelve cuando el token no está: o nunca existió, o lo borró uno más nuevo.
var ErrTokenNotFound = errors.New("auth.token.used")

// PasswordToken es un enlace de contraseña.
type PasswordToken struct {
	ID        int64      `gorm:"primaryKey"`
	UserID    int64      `gorm:"column:user_id"`
	Purpose   string     `gorm:"column:purpose"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedBy *int64     `gorm:"column:created_by_id"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (PasswordToken) TableName() string { return "password_tokens" }

// Expired dice si el enlace ya pasó de su hora.
func (t PasswordToken) Expired(now time.Time) bool { return now.After(t.ExpiresAt) }

// Used dice si el enlace ya se usó, que es lo mismo que no valer.
func (t PasswordToken) Used() bool { return t.UsedAt != nil }

// PasswordTokenRepository lee y escribe los enlaces.
type PasswordTokenRepository struct {
	db *gorm.DB
}

// NewPasswordTokenRepository construye el repositorio.
func NewPasswordTokenRepository(db *gorm.DB) *PasswordTokenRepository {
	return &PasswordTokenRepository{db: db}
}

// Create borra los enlaces anteriores de esa cuenta y ese propósito, y guarda el nuevo.
//
// Las dos cosas van en una transacción: pedir un enlace nuevo y quedarse sin ninguno porque se borró
// el viejo y falló el insert sería la peor forma de fallar. Y se borran los anteriores a propósito:
// si alguien pide tres correos seguidos, sólo vale el último (docs/modules/auth.md, sección 2).
func (r *PasswordTokenRepository) Create(token *PasswordToken) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND purpose = ?", token.UserID, token.Purpose).
			Delete(&PasswordToken{}).Error; err != nil {
			return err
		}

		return tx.Create(token).Error
	})
}

// FindByHash busca el enlace por la huella del token que llegó.
//
// Se busca por la huella y no por el token porque el token no está guardado en ningún sitio: es lo
// que hace que leer la base no sirva para entrar en ninguna cuenta.
func (r *PasswordTokenRepository) FindByHash(hash string) (PasswordToken, error) {
	var token PasswordToken

	err := r.db.Where("token_hash = ?", hash).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return PasswordToken{}, ErrTokenNotFound
	}

	return token, err
}

// MarkUsed marca el enlace como usado: un enlace, un uso.
//
// Sólo marca si seguía sin usar, así que dos peticiones con el mismo token a la vez no pueden usarlo
// las dos: la segunda no toca ninguna fila y se entera.
func (r *PasswordTokenRepository) MarkUsed(id int64, now time.Time) (bool, error) {
	result := r.db.Model(&PasswordToken{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now)

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}

// DeleteFor borra los enlaces de una cuenta y un propósito. Se usa cuando la contraseña ya se ha
// establecido: los enlaces que quedaran abiertos dejan de tener sentido.
func (r *PasswordTokenRepository) DeleteFor(userID int64, purpose string) error {
	return r.db.Where("user_id = ? AND purpose = ?", userID, purpose).
		Delete(&PasswordToken{}).Error
}
