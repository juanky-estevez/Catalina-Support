// Package repositories es el acceso a datos del módulo users: la tabla de cuentas.
package repositories

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"catalina-support/backend/shared/auth"
)

// ErrUserNotFound se devuelve cuando la cuenta pedida no existe.
var ErrUserNotFound = errors.New("users.notFound")

// User es una cuenta de la tabla `users`.
//
// Los valores nulos de la base se leen como cadenas vacías o punteros nulos, y el modelo los
// conserva tal cual: quien decide qué significa cada uno es el servicio, no el acceso a datos
// (docs/modules/users.md, sección 2).
type User struct {
	ID           int64      `gorm:"primaryKey"`
	Name         string     `gorm:"column:name"`
	LastName     string     `gorm:"column:last_name"`
	Email        string     `gorm:"column:email"`
	PasswordHash *string    `gorm:"column:password_hash"`
	Role         string     `gorm:"column:role"`
	Origin       string     `gorm:"column:origin"`
	ExternalID   *string    `gorm:"column:external_id"`
	Language     string     `gorm:"column:language"`
	IsActive     bool       `gorm:"column:is_active"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

// TableName fija el nombre de la tabla, que no se deduce del tipo.
func (User) TableName() string { return "users" }

// Account traduce la fila a lo que el resto del sistema entiende por cuenta.
//
// Es la frontera del módulo: fuera no se ve `*string` ni el nombre de una columna
// (docs/arquitectura.md, sección 4).
func (u User) Account() auth.Account {
	account := auth.Account{
		ID:          u.ID,
		Name:        u.Name,
		LastName:    u.LastName,
		Email:       u.Email,
		Role:        u.Role,
		Origin:      u.Origin,
		Language:    u.Language,
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
	}

	if u.PasswordHash != nil {
		account.PasswordHash = *u.PasswordHash
	}
	if u.ExternalID != nil {
		account.ExternalID = *u.ExternalID
	}

	return account
}

// textOrNil convierte una cadena vacía en un nulo, que es como la base guarda «no hay».
func textOrNil(valor string) *string {
	if strings.TrimSpace(valor) == "" {
		return nil
	}
	return &valor
}

// UserRepository lee y escribe las cuentas.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository construye el repositorio.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID devuelve una cuenta por su identificador.
func (r *UserRepository) FindByID(id int64) (User, error) {
	var user User

	err := r.db.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, ErrUserNotFound
	}

	return user, err
}

// FindByEmail busca por correo **sin distinguir mayúsculas**: `Ana@…` y `ana@…` son la misma persona.
//
// Se compara con `lower()` y no con `ILIKE` porque el índice único está sobre `lower(email)` y así
// la búsqueda lo usa.
func (r *UserRepository) FindByEmail(email string) (User, error) {
	var user User

	err := r.db.Where("lower(email) = lower(?)", strings.TrimSpace(email)).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, ErrUserNotFound
	}

	return user, err
}

// FindByExternalID busca una cuenta de directorio por su identificador allí.
func (r *UserRepository) FindByExternalID(origin, externalID string) (User, error) {
	var user User

	err := r.db.Where("origin = ? AND external_id = ?", origin, externalID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, ErrUserNotFound
	}

	return user, err
}

// EmailExists dice si ese correo ya lo usa **otra** cuenta.
//
// `exceptID` es la cuenta que se está editando, que no cuenta como duplicado de sí misma.
func (r *UserRepository) EmailExists(email string, exceptID int64) (bool, error) {
	var total int64

	consulta := r.db.Model(&User{}).Where("lower(email) = lower(?)", strings.TrimSpace(email))
	if exceptID > 0 {
		consulta = consulta.Where("id <> ?", exceptID)
	}

	if err := consulta.Count(&total).Error; err != nil {
		return false, err
	}

	return total > 0, nil
}

// Filtros de la lista de cuentas (docs/modules/users.md, sección 4).
type Filtros struct {
	// Role, Origin e IsActive vacíos o nulos quieren decir «sin filtrar por eso».
	Role     string
	Origin   string
	IsActive *bool
	// Query busca en el nombre, los apellidos y el correo, sin distinguir mayúsculas.
	Query string
}

// FindByIDs devuelve las cuentas de esos identificadores, sin importar el orden.
//
// Es lo que evita preguntar una a una cuando una pantalla enseña veinticinco tickets con su
// solicitante y su responsable.
func (r *UserRepository) FindByIDs(ids []int64) ([]User, error) {
	if len(ids) == 0 {
		return []User{}, nil
	}

	var users []User
	if err := r.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// FindActiveByRole devuelve las cuentas activas con ese papel, por apellidos y nombre.
//
// Lo usan dos cosas que no son de este módulo: el reparto por turnos de `tickets`, que necesita saber
// quién está de servicio, y su lista de responsables posibles.
func (r *UserRepository) FindActiveByRole(role string) ([]User, error) {
	var users []User

	err := r.db.
		Where("role = ? AND is_active = true", role).
		Order("last_name, name").
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

// List devuelve una página de cuentas, ordenadas por apellidos y nombre.
//
// El orden es el de una lista de personas, no el de inserción: buscar a alguien en una lista que
// cambia de orden cada vez que se da de alta a otro es lo que hace que nadie la use.
func (r *UserRepository) List(filtros Filtros, pagina, porPagina int) ([]User, error) {
	var usuarios []User

	consulta := r.conFiltros(r.db.Model(&User{}), filtros).
		Order("last_name, name").
		Limit(porPagina).
		Offset((pagina - 1) * porPagina)

	err := consulta.Find(&usuarios).Error
	return usuarios, err
}

// Count dice cuántas cuentas hay con esos filtros, que es lo que necesita la paginación.
func (r *UserRepository) Count(filtros Filtros) (int64, error) {
	var total int64

	err := r.conFiltros(r.db.Model(&User{}), filtros).Count(&total).Error
	return total, err
}

// conFiltros aplica los filtros que se hayan pedido.
func (r *UserRepository) conFiltros(consulta *gorm.DB, filtros Filtros) *gorm.DB {
	if filtros.Role != "" {
		consulta = consulta.Where("role = ?", filtros.Role)
	}
	if filtros.Origin != "" {
		consulta = consulta.Where("origin = ?", filtros.Origin)
	}
	if filtros.IsActive != nil {
		consulta = consulta.Where("is_active = ?", *filtros.IsActive)
	}
	if texto := strings.TrimSpace(filtros.Query); texto != "" {
		// Se busca en los tres sitios a la vez: quien busca «ana» espera encontrarla escriba el
		// nombre o el correo.
		patron := "%" + strings.ToLower(texto) + "%"
		consulta = consulta.Where(
			"lower(name) LIKE ? OR lower(last_name) LIKE ? OR lower(email) LIKE ?",
			patron, patron, patron,
		)
	}

	return consulta
}

// Create guarda una cuenta nueva y devuelve su identificador.
func (r *UserRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// UpdatePasswordHash deja la cuenta con esa contraseña, o sin ninguna si llega vacío.
func (r *UserRepository) UpdatePasswordHash(id int64, hash string) error {
	result := r.db.Model(&User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"password_hash": textOrNil(hash),
			"updated_at":    time.Now(),
		})

	return r.afterWrite(result)
}

// TouchLastLogin apunta cuándo entró por última vez.
//
// Se hace al entrar y no en cada petición: es información para la ficha de la cuenta, no para
// vigilar a nadie (docs/modules/users.md, sección 2).
func (r *UserRepository) TouchLastLogin(id int64, cuando time.Time) error {
	return r.afterWrite(r.db.Model(&User{}).Where("id = ?", id).Update("last_login_at", cuando))
}

// UpdateFields cambia las columnas indicadas de una cuenta. El mapa lo arma el servicio, que es
// quien decide qué se puede cambiar (docs/modules/users.md, sección 6).
func (r *UserRepository) UpdateFields(id int64, cambios map[string]any) error {
	cambios["updated_at"] = time.Now()
	return r.afterWrite(r.db.Model(&User{}).Where("id = ?", id).Updates(cambios))
}

// afterWrite convierte «no se ha tocado ninguna fila» en «la cuenta no existe», que es lo que quiere
// saber quien pregunta: una escritura sobre algo que no está tiene que fallar, no pasar en silencio.
func (r *UserRepository) afterWrite(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
