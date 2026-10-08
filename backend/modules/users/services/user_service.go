// Package services contiene las reglas del módulo users: qué es un alta válida, quién puede dar de
// alta qué, cómo se cambia el papel y cómo se desactiva una cuenta.
package services

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"catalina-support/backend/modules/users/repositories"
	"catalina-support/backend/shared/auth"
)

// Las claves de error del módulo, tal y como viajan al frontend (docs/modules/users.md, sección 9).
var (
	ErrNameRequired      = errors.New("users.name.required")
	ErrAccountInactive   = errors.New("users.accountInactive")
	ErrEmailInvalid      = errors.New("users.email.invalid")
	ErrEmailDuplicated   = errors.New("users.email.duplicated")
	ErrRoleNotAllowed    = errors.New("users.role.notAllowed")
	ErrOriginUnknown     = errors.New("users.origin.unknown")
	ErrDirectoryNotFound = errors.New("users.directory.notFound")
	// ErrPasswordNotLocal: se ha pedido cambiar la contraseña de una cuenta que **no es local**. Con
	// Active Directory o Keycloak la contraseña la comprueba el directorio, así que un enlace nuestro no
	// serviría de nada (decisión del responsable, 2026-09-29).
	ErrPasswordNotLocal = errors.New("users.password.notLocal")
	// ErrOriginByDirectory: las cuentas de directorio no se crean ni se cambian a mano. No es un
	// rechazo por no poder comprobarlo —eso ya se puede—, es que **no hace falta y no se debe**:
	// quien está en el directorio entra solo y su cuenta se crea, se vincula o se pone al día en ese
	// primer acceso (docs/modules/users.md, sección 5, punto 4).
	ErrOriginByDirectory = errors.New("users.origin.byDirectory")
	// ErrDirectoryActivatesItself: reactivar una cuenta de Keycloak a mano. Su camino **se reactiva
	// solo** al entrar, y preguntarle a Keycloak si esa persona sigue allí pediría una cuenta de
	// servicio en el reino para un botón que no hace falta (docs/modules/users.md, sección 5, punto 4).
	ErrDirectoryActivatesItself = errors.New("users.directory.activatesItself")
	// ErrDirectoryUnavailable: no se puede comprobar con el directorio ahora mismo, o esta instalación
	// no tiene ninguno conectado. Es un 503: el fallo no es de quien pide la reactivación.
	ErrDirectoryUnavailable = errors.New("users.directory.unavailable")
)

// PasswordLinks es lo que este módulo necesita de `auth`: emitir el enlace y mandar el correo.
//
// `users` crea la cuenta; el enlace y el correo son de `auth`, que es quien sabe de contraseñas y de
// tokens, y `auth` se los pide a `mail`, que tiene el texto (docs/modules/users.md, sección 5).
//
// Se declara **aquí, en quien lo usa**, y `auth` lo cumple sin saber que existe esta interfaz: así
// ninguno de los dos módulos importa al otro y no hay círculo que no compile. Quien los une es
// `main.go` (docs/arquitectura.md, sección 4).
type PasswordLinks interface {
	// Invite emite el enlace de alta (24 horas) y manda el correo.
	Invite(account auth.Account, requestedBy *int64) error
	// Recovery emite el enlace de recuperación (1 hora) y manda el correo.
	Recovery(account auth.Account, requestedBy *int64) error
}

// Directory es lo que este módulo necesita del directorio de la organización, y **se declara aquí**,
// en quien lo usa, como `PasswordLinks`: así `users` no importa a `auth` y es `auth` quien cumple
// esta interfaz sin saber que existe (docs/arquitectura.md, sección 4).
type Directory interface {
	// Knows dice si el directorio sigue conociendo a alguien con ese correo. Es lo que hace falta para
	// reactivar una cuenta de AD sin devolverle un acceso que no podría usar
	// (docs/modules/users.md, sección 5, punto 4).
	Knows(email string) (bool, error)
}

// Service son las cuentas.
type Service struct {
	repo      *repositories.UserRepository
	links     PasswordLinks
	directory Directory
}

// NewService construye el servicio.
//
// Los enlaces se conectan después con `SetPasswordLinks`, porque `auth` necesita a su vez las cuentas
// para entrar: los dos módulos se necesitan y alguien tiene que romper el círculo. Se rompe aquí, en
// el único sitio donde no hay más remedio (docs/arquitectura.md, sección 4).
func NewService(repo *repositories.UserRepository) *Service {
	return &Service{repo: repo}
}

// SetPasswordLinks conecta el módulo de contraseñas. Se llama una vez, al arrancar.
func (s *Service) SetPasswordLinks(links PasswordLinks) { s.links = links }

// SetDirectory conecta el directorio de la organización, que es lo que hace falta para comprobar que
// una cuenta de AD sigue estando allí antes de reactivarla. Sin él, ese camino no existe y la
// reactivación de una cuenta de AD se rechaza diciendo que no se puede comprobar.
func (s *Service) SetDirectory(directory Directory) { s.directory = directory }

// CreateInput es lo que llega para dar de alta una cuenta.
type CreateInput struct {
	Name     string
	LastName string
	Email    string
	Role     string
	Origin   string
	// ExternalID sólo se usa en las cuentas de directorio, y lo resuelve quien las da de alta.
	ExternalID string
}

// Create da de alta una cuenta y, si es local, le manda el enlace para establecer la contraseña.
//
// Las reglas del alta están en `docs/modules/users.md`, sección 5.
func (s *Service) Create(input CreateInput, actor auth.Identity) (auth.Account, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = strings.TrimSpace(input.Email)
	input.Role = strings.TrimSpace(input.Role)
	input.Origin = strings.TrimSpace(input.Origin)

	if input.Name == "" || input.LastName == "" {
		return auth.Account{}, ErrNameRequired
	}
	if !EmailIsValid(input.Email) {
		return auth.Account{}, ErrEmailInvalid
	}
	// El papel tiene que existir **y** estar entre los que ese papel puede repartir.
	if !auth.RoleIsValid(input.Role) || !CanAssignRole(actor, input.Role) {
		return auth.Account{}, ErrRoleNotAllowed
	}
	if !auth.OriginIsValid(input.Origin) {
		return auth.Account{}, ErrOriginUnknown
	}
	// Y ahora que se sabe que el origen existe: **las cuentas de directorio no se dan de alta a mano**.
	// Quien está en el directorio entra solo y su cuenta nace en ese primer acceso, con lo que dice el
	// directorio; crearla aquí antes de tiempo sería una cuenta con la contraseña de otro sitio y unos
	// datos que envejecen (docs/modules/users.md, sección 5, punto 4). Lo decidió el responsable el
	// 2026-09-25: nadie da de alta a nadie que esté en el directorio.
	if input.Origin != auth.OriginLocal {
		return auth.Account{}, ErrOriginByDirectory
	}
	// El idioma es global. Se ignora el valor personal de clientes anteriores y se conserva en esta
	// columna sólo mientras la migración 1.0.0 termina de retirarla.
	repetido, err := s.repo.EmailExists(input.Email, 0)
	if err != nil {
		return auth.Account{}, err
	}
	if repetido {
		return auth.Account{}, ErrEmailDuplicated
	}

	user := &repositories.User{
		Name:     input.Name,
		LastName: input.LastName,
		Email:    input.Email,
		Role:     input.Role,
		Origin:   input.Origin,
		IsActive: true,
	}
	if input.ExternalID != "" {
		user.ExternalID = &input.ExternalID
	}

	if err := s.repo.Create(user); err != nil {
		return auth.Account{}, err
	}

	account := user.Account()

	// Una cuenta de directorio no recibe correo: su contraseña es la del dominio. Una local no puede
	// entrar hasta que establece la suya, así que el enlace es parte del alta
	// (docs/modules/users.md, sección 5).
	if account.Origin == auth.OriginLocal {
		if s.links == nil {
			// Sin con quién emitir el enlace, la cuenta se queda sin poder entrar y nadie se entera.
			// Se falla a la vista en lugar de dejar una cuenta muda.
			return account, errors.New("no hay módulo de contraseñas conectado: no se puede mandar el enlace de alta")
		}
		if err := s.links.Invite(account, actorID(actor)); err != nil {
			// La cuenta ya existe: el correo se puede volver a lanzar desde el listado
			// (docs/modules/users.md, sección 5, punto 6).
			return account, err
		}
	}

	return account, nil
}

// SendInviteAgain lanza otra vez el enlace de la cuenta: el alta y la recuperación son el mismo
// enlace, y es la salida para quien no recibió el primer correo.
func (s *Service) SendInviteAgain(id int64, actor auth.Identity) (auth.Account, error) {
	account, err := s.ByID(id)
	if err != nil {
		return auth.Account{}, err
	}

	if !account.IsActive {
		// Un enlace para quien no puede entrar es una promesa falsa: establecería la contraseña y
		// seguiría fuera, sin entender por qué (docs/modules/users.md, sección 5, punto 5).
		return auth.Account{}, ErrAccountInactive
	}

	if account.Origin != auth.OriginLocal {
		// **La contraseña sólo es nuestra en las cuentas locales** (decisión del responsable,
		// 2026-09-29): con el directorio o Keycloak la comprueba otro, y cambiarla desde aquí sería una
		// promesa falsa. La pantalla ya no ofrece el botón; esto lo garantiza también por la API.
		return auth.Account{}, ErrPasswordNotLocal
	}

	if s.links == nil {
		return auth.Account{}, errors.New("no hay módulo de contraseñas conectado: no se puede mandar el enlace")
	}

	// Si nunca ha entrado, el enlace es el de alta, que dura 24 horas; si ya tiene contraseña, es el
	// de recuperación, que dura 1 hora. Eso lo decide `auth`, aquí sólo se pide.
	if account.HasPassword() {
		err = s.links.Recovery(account, actorID(actor))
	} else {
		err = s.links.Invite(account, actorID(actor))
	}
	if err != nil {
		return auth.Account{}, err
	}

	return account, nil
}

// ByID devuelve una cuenta por su identificador.
func (s *Service) ByID(id int64) (auth.Account, error) {
	user, err := s.repo.FindByID(id)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return auth.Account{}, auth.ErrAccountNotFound
	}
	if err != nil {
		return auth.Account{}, err
	}

	return user.Account(), nil
}

// ByIDs devuelve varias cuentas de una vez, cada una con su identificador.
//
// Es lo que evita preguntar una a una cuando una pantalla enseña veinticinco tickets con su
// solicitante y su responsable: una consulta en vez de cincuenta.
func (s *Service) ByIDs(ids []int64) (map[int64]auth.Account, error) {
	if len(ids) == 0 {
		return map[int64]auth.Account{}, nil
	}

	// Los identificadores repetidos se piden una sola vez: en una lista de tickets el mismo técnico
	// aparece en muchas filas.
	unicos := make([]int64, 0, len(ids))
	vistos := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if id == 0 || vistos[id] {
			continue
		}
		vistos[id] = true
		unicos = append(unicos, id)
	}

	users, err := s.repo.FindByIDs(unicos)
	if err != nil {
		return nil, err
	}

	cuentas := make(map[int64]auth.Account, len(users))
	for _, user := range users {
		cuentas[user.ID] = user.Account()
	}

	return cuentas, nil
}

// ActiveByRole son las cuentas activas con ese papel.
//
// Lo usan dos cosas que no son de este módulo: el reparto por turnos de `tickets`, que necesita saber
// quién está de servicio, y su lista de responsables posibles.
func (s *Service) ActiveByRole(role string) ([]auth.Account, error) {
	if !auth.RoleIsValid(role) {
		return nil, ErrRoleNotAllowed
	}

	users, err := s.repo.FindActiveByRole(role)
	if err != nil {
		return nil, err
	}

	cuentas := make([]auth.Account, 0, len(users))
	for _, user := range users {
		cuentas = append(cuentas, user.Account())
	}

	return cuentas, nil
}

// ByEmail devuelve una cuenta por su correo, sin distinguir mayúsculas.
func (s *Service) ByEmail(email string) (auth.Account, error) {
	user, err := s.repo.FindByEmail(email)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return auth.Account{}, auth.ErrAccountNotFound
	}
	if err != nil {
		return auth.Account{}, err
	}

	return user.Account(), nil
}

// SetPasswordHash deja la cuenta con esa contraseña ya cifrada.
func (s *Service) SetPasswordHash(id int64, hash string) error {
	err := s.repo.UpdatePasswordHash(id, hash)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return auth.ErrAccountNotFound
	}
	return err
}

// TouchLastLogin apunta la última entrada.
func (s *Service) TouchLastLogin(id int64) error {
	err := s.repo.TouchLastLogin(id, time.Now())
	if errors.Is(err, repositories.ErrUserNotFound) {
		return auth.ErrAccountNotFound
	}
	return err
}

// CanAssignRole dice si quien da de alta puede repartir ese papel: Soporte sólo crea usuarios, y
// repartir soporte, desarrollo o administrador es cosa de un Administrador
// (docs/usuarios-y-permisos.md, sección 3).
func CanAssignRole(actor auth.Identity, role string) bool {
	if actor.Can(auth.RoleAdministrador) {
		return true
	}
	return actor.Can(auth.RoleSoporte) && role == auth.RoleUsuario
}

// EmailIsValid comprueba la forma del correo, sin más: una dirección que existe de verdad sólo se
// sabe mandándole algo, y eso es justo lo que hace el alta.
//
// **Juzga, no limpia**: si le llega con espacios o saltos de línea alrededor, dice que no. Quien
// recorta es el alta, antes de llamar aquí, así que pegar un correo con un espacio detrás sigue
// funcionando; lo que no puede pasar es que un salto de línea acabe guardado en la tabla.
func EmailIsValid(email string) bool {
	if email == "" || email != strings.TrimSpace(email) {
		return false
	}
	if strings.ContainsAny(email, " \t\n\r") {
		return false
	}

	direccion, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	// `ParseAddress` acepta `Nombre <a@b>`, y aquí sólo vale la dirección pelada.
	if direccion.Address != email {
		return false
	}

	arroba := strings.LastIndex(email, "@")
	if arroba <= 0 || arroba == len(email)-1 {
		return false
	}

	// Con punto en el dominio: `ana@localhost` no es un correo al que se le pueda escribir.
	return strings.Contains(email[arroba+1:], ".")
}

// actorID es quién provoca la acción, o nada si es la cuenta de fábrica, que no está en la tabla.
func actorID(actor auth.Identity) *int64 {
	if actor.Factory || actor.ID == 0 {
		return nil
	}
	return &actor.ID
}
