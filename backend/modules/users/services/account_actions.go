package services

import (
	"errors"
	"strconv"
	"strings"

	"github.com/juanky-estevez/go-logs"

	"catalina-support/backend/modules/users/repositories"
	"catalina-support/backend/shared/auth"
)

// Las claves de las acciones que se añaden al módulo (docs/modules/users.md, sección 9).
var (
	ErrSelfDeactivation = errors.New("users.selfDeactivation")
	ErrOriginInUse      = errors.New("users.origin.inUse")
	// ErrStateHasItsOwnAction: el estado no se cambia por el `PATCH` general.
	//
	// El documento lo dice con su motivo (docs/modules/users.md, sección 4): desactivar tiene su
	// propia ruta «para que no se pueda desactivar a alguien por accidente al editar su nombre». Aquí
	// se cumple esa intención, en vez de dejar el campo suelto.
	ErrStateHasItsOwnAction = errors.New("users.stateHasItsOwnAction")
)

// Los avisos que acompañan a una acción que sí se hace, pero conviene saber. No son errores: la
// acción se lleva a cabo.
const (
	// AvisoDirectorioVuelve: a una cuenta de directorio no se le quita el acceso desde aquí; si allí
	// sigue activa, volverá a entrar sola (docs/modules/users.md, sección 7).
	AvisoDirectorioVuelve = "users.directoryMayReturn"
)

// Filtros de la lista de cuentas.
type Filtros struct {
	Role     string
	Origin   string
	IsActive *bool
	Query    string
}

// Pagina es una página de cuentas, con lo que necesita la pantalla para pintar la paginación.
type Pagina struct {
	Accounts []auth.Account
	Total    int64
	Page     int
	PerPage  int
}

// UpsertFromDirectory crea, vincula o actualiza la cuenta de alguien que ha entrado por el
// directorio (docs/modules/auth.md, sección 5.4).
//
// Las cuatro situaciones, y en este orden:
//
//  1. **Ya hay una cuenta de ese directorio**: se busca por su identificador externo, que es lo que
//     manda cuando alguien cambia de correo en la empresa. Se actualizan nombre, apellidos y correo.
//  2. **Hay una cuenta con ese correo**: se **vincula** al directorio. Su contraseña local deja de
//     servir —la base no admite contraseña en una cuenta de directorio— y quien manda a partir de
//     ahí es el directorio. El papel no se toca: la persona es la misma, y lo que cambia es de dónde
//     viene su contraseña.
//  3. **No hay ninguna cuenta**: se crea con papel `usuario`, que es el alta automática del primer
//     acceso.
//  4. **Estaba desactivada y el directorio la tiene activa**: se reactiva. La regla de convivencia lo
//     dice: si allí está activa, vuelve a entrar sola (regla 6).
func (s *Service) UpsertFromDirectory(datos auth.DirectoryAccount) (auth.Account, error) {
	correo := strings.TrimSpace(datos.Email)
	externo := strings.TrimSpace(datos.ExternalID)

	if !EmailIsValid(correo) {
		return auth.Account{}, ErrEmailInvalid
	}
	if !auth.OriginIsValid(datos.Origin) || datos.Origin == auth.OriginLocal {
		return auth.Account{}, ErrOriginUnknown
	}

	// 1. Por el identificador del directorio.
	if externo != "" {
		user, err := s.repo.FindByExternalID(datos.Origin, externo)
		if err == nil {
			return s.actualizarDesdeElDirectorio(user, datos)
		}
		if !errors.Is(err, repositories.ErrUserNotFound) {
			return auth.Account{}, err
		}
	}

	// 2. Por el correo.
	user, err := s.repo.FindByEmail(correo)
	if err == nil {
		return s.actualizarDesdeElDirectorio(user, datos)
	}
	if !errors.Is(err, repositories.ErrUserNotFound) {
		return auth.Account{}, err
	}

	// 3. No hay ninguna: se crea, con el idioma de la instalación.
	id := externo
	nombre := strings.TrimSpace(datos.Name)
	apellidos := strings.TrimSpace(datos.LastName)

	nueva := &repositories.User{
		Name:     nombre,
		LastName: apellidos,
		Email:    correo,
		Role:     auth.RoleUsuario,
		Origin:   datos.Origin,
		IsActive: true,
	}
	if id != "" {
		nueva.ExternalID = &id
	}

	if err := s.repo.Create(nueva); err != nil {
		return auth.Account{}, err
	}

	logs.LogSuccess("cuenta creada al entrar por el directorio: la #" + strconv.FormatInt(nueva.ID, 10))

	return nueva.Account(), nil
}

// actualizarDesdeElDirectorio deja la cuenta con lo que dice el directorio.
func (s *Service) actualizarDesdeElDirectorio(user repositories.User, datos auth.DirectoryAccount) (auth.Account, error) {
	correo := strings.TrimSpace(datos.Email)

	// Si el correo nuevo ya es de otra cuenta, no se puede seguir: la base no admite dos cuentas con
	// el mismo correo, y elegir una sería decidir por su dueño. Se dice, y queda en el log.
	if !strings.EqualFold(correo, user.Email) {
		repetido, err := s.repo.EmailExists(correo, user.ID)
		if err != nil {
			return auth.Account{}, err
		}
		if repetido {
			logs.LogWarning("el directorio trae un correo que ya es de otra cuenta: la #" + strconv.FormatInt(user.ID, 10) + " no se puede actualizar con él")
			return auth.Account{}, ErrEmailDuplicated
		}
	}

	externo := strings.TrimSpace(datos.ExternalID)
	nombre := strings.TrimSpace(datos.Name)
	apellidos := strings.TrimSpace(datos.LastName)

	cambios := map[string]any{
		"origin":    datos.Origin,
		"is_active": true,
	}
	if nombre != "" {
		cambios["name"] = nombre
	}
	if apellidos != "" {
		cambios["last_name"] = apellidos
	}
	if correo != "" {
		cambios["email"] = correo
	}
	if externo != "" {
		cambios["external_id"] = externo
	}
	if user.Origin == auth.OriginLocal {
		// Al vincular, la contraseña local deja de servir: es la regla de convivencia, y la base exige
		// que una cuenta de directorio no tenga contraseña propia.
		cambios["password_hash"] = nil
	}

	if err := s.repo.UpdateFields(user.ID, cambios); err != nil {
		return auth.Account{}, err
	}

	actualizado, err := s.ByID(user.ID)
	if err != nil {
		return auth.Account{}, err
	}

	return actualizado, nil
}

// List devuelve la lista de cuentas, con sus filtros y su paginación.
//
// La lista la ven Soporte y Administrador: Soporte la necesita para buscar a quien reporta, y por eso
// **no lleva la ficha de nadie más** que lo que enseña la propia lista (docs/modules/users.md,
// sección 4).
func (s *Service) List(filtros Filtros, pagina, porPagina int) (Pagina, error) {
	if pagina < 1 {
		pagina = 1
	}
	// Un tope: sin él, una petición con `porPagina=100000` se lleva la tabla entera.
	if porPagina < 1 || porPagina > 100 {
		porPagina = 25
	}

	// Un filtro con un valor que no existe no es un error: devuelve la lista vacía, que es lo que
	// espera quien acaba de escribir mal una búsqueda.
	delRepositorio := repositories.Filtros{
		Role:     strings.TrimSpace(filtros.Role),
		Origin:   strings.TrimSpace(filtros.Origin),
		IsActive: filtros.IsActive,
		Query:    strings.TrimSpace(filtros.Query),
	}

	total, err := s.repo.Count(delRepositorio)
	if err != nil {
		return Pagina{}, err
	}

	usuarios, err := s.repo.List(delRepositorio, pagina, porPagina)
	if err != nil {
		return Pagina{}, err
	}

	cuentas := make([]auth.Account, 0, len(usuarios))
	for _, usuario := range usuarios {
		cuentas = append(cuentas, usuario.Account())
	}

	return Pagina{Accounts: cuentas, Total: total, Page: pagina, PerPage: porPagina}, nil
}

// PatchInput es lo que se puede cambiar de una cuenta.
//
// Los campos van con puntero porque hay que distinguir «no lo toques» de «déjalo vacío»: un `PATCH`
// que no manda el papel no está diciendo que la cuenta se quede sin papel.
type PatchInput struct {
	Name     *string
	LastName *string
	Email    *string
	Role     *string
	IsActive *bool
}

// Patch cambia una cuenta, con las reglas de quién puede qué.
//
// **Soporte sólo cambia el nombre y los apellidos**: el correo, el papel y el estado son decisiones
// de otro (docs/modules/users.md, sección 4). Es una regla que depende de los datos —qué campos
// vienen— y por eso se comprueba aquí y no en la ruta.
func (s *Service) Patch(id int64, input PatchInput, actor auth.Identity) (auth.Account, error) {
	return s.patch(id, input, actor, false)
}

// patch es el cambio de una cuenta, con `propio` en verdadero cuando es la de quien pide el cambio.
//
// La diferencia importa: **el límite de Soporte es para editar a otros**. Soporte sólo cambia el
// nombre y los apellidos de otra persona, pero de la suya cambia lo que cualquiera: su nombre, sus
// apellidos y su idioma. Aplicarle ese límite a su propio perfil le impediría cambiar su idioma.
func (s *Service) patch(id int64, input PatchInput, actor auth.Identity, propio bool) (auth.Account, error) {
	esAdministrador := actor.Can(auth.RoleAdministrador)

	if !propio && !esAdministrador {
		// Soporte editando a otra persona: sólo el nombre y los apellidos.
		if input.Email != nil || input.Role != nil || input.IsActive != nil {
			return auth.Account{}, ErrRoleNotAllowed
		}
	}

	// El estado no se cambia por aquí, y se rechaza **antes de leer nada**: desactivar tiene su
	// propia acción, que es la que impide desactivarse a uno mismo.
	if input.IsActive != nil {
		return auth.Account{}, ErrStateHasItsOwnAction
	}

	actual, err := s.ByID(id)
	if err != nil {
		return auth.Account{}, err
	}

	cambios := map[string]any{}

	if input.Name != nil {
		nombre := strings.TrimSpace(*input.Name)
		if nombre == "" {
			return auth.Account{}, ErrNameRequired
		}
		cambios["name"] = nombre
	}

	if input.LastName != nil {
		apellidos := strings.TrimSpace(*input.LastName)
		if apellidos == "" {
			return auth.Account{}, ErrNameRequired
		}
		cambios["last_name"] = apellidos
	}

	if input.Email != nil {
		correo := strings.TrimSpace(*input.Email)
		if !EmailIsValid(correo) {
			return auth.Account{}, ErrEmailInvalid
		}

		repetido, err := s.repo.EmailExists(correo, id)
		if err != nil {
			return auth.Account{}, err
		}
		if repetido {
			return auth.Account{}, ErrEmailDuplicated
		}

		cambios["email"] = correo
	}

	if input.Role != nil {
		papel := strings.TrimSpace(*input.Role)
		if !auth.RoleIsValid(papel) || !CanAssignRole(actor, papel) {
			return auth.Account{}, ErrRoleNotAllowed
		}
		cambios["role"] = papel
	}

	if len(cambios) == 0 {
		return actual, nil
	}

	if err := s.repo.UpdateFields(id, cambios); err != nil {
		return auth.Account{}, traducirUsuario(err)
	}

	return s.ByID(id)
}

// PatchOwn cambia el perfil propio: nombre, apellidos e idioma, y nada más.
func (s *Service) PatchOwn(input PatchInput, actor auth.Identity) (auth.Account, error) {
	if input.Role != nil || input.Email != nil {
		// El correo y el papel son decisiones de otro (sección 8).
		return auth.Account{}, ErrRoleNotAllowed
	}

	if input.IsActive != nil {
		// Y el estado se cambia con su propia acción, nunca editando la cuenta.
		return auth.Account{}, ErrStateHasItsOwnAction
	}

	if actor.Factory {
		// La cuenta de fábrica no está en la tabla de cuentas: no hay fila que cambiar. Su nombre y su
		// contraseña viven en la configuración de la instalación.
		return auth.Account{}, auth.ErrAccountNotFound
	}

	return s.patch(actor.ID, input, actor, true)
}

// Deactivate desactiva una cuenta.
//
// Desactivar no borra nada: la cuenta deja de poder entrar y los tickets que tenga siguen contando su
// historia con su nombre (docs/modules/users.md, sección 7).
func (s *Service) Deactivate(id int64, actor auth.Identity) (auth.Account, []string, error) {
	if !actor.Factory && actor.ID == id {
		// Desactivarse es quedarse fuera al instante, y quien lo hace no puede volver: hace falta que
		// otro le reactive.
		return auth.Account{}, nil, ErrSelfDeactivation
	}

	cuenta, err := s.ByID(id)
	if err != nil {
		return auth.Account{}, nil, err
	}

	if err := s.repo.UpdateFields(id, map[string]any{"is_active": false}); err != nil {
		return auth.Account{}, nil, traducirUsuario(err)
	}

	cuenta.IsActive = false

	// A una cuenta de directorio no se le quita el acceso desde aquí: se avisa, y la interfaz lo dice.
	var avisos []string
	if cuenta.IsDirectory() {
		avisos = append(avisos, AvisoDirectorioVuelve)
	}

	return cuenta, avisos, nil
}

// Activate reactiva una cuenta.
func (s *Service) Activate(id int64, actor auth.Identity) (auth.Account, error) {
	cuenta, err := s.ByID(id)
	if err != nil {
		return auth.Account{}, err
	}

	if cuenta.IsDirectory() {
		return s.reactivarDeDirectorio(cuenta)
	}

	if err := s.repo.UpdateFields(id, map[string]any{"is_active": true}); err != nil {
		return auth.Account{}, traducirUsuario(err)
	}

	cuenta.IsActive = true
	return cuenta, nil
}

// reactivarDeDirectorio devuelve el acceso a una cuenta de directorio, y **no a ciegas**: comprueba
// antes con el directorio, porque devolvérselo a quien ya no está allí sería devolverle un acceso que
// no puede usar —su contraseña es de allí— (docs/modules/users.md, sección 5, punto 4).
//
// Los dos caminos no se comprueban igual, y es a propósito:
//
//   - **AD**: se pregunta al directorio si sigue conociendo a esa persona, con la misma búsqueda con
//     la que se entra. Si no aparece, `users.directory.notFound`.
//   - **Keycloak**: no se pregunta. Hacerlo pediría una **cuenta de servicio con permiso de lectura
//     dentro del reino**, y eso no se le pide a quien administra Keycloak por un botón que se arregla
//     solo: esa cuenta **entra por su camino y al entrar se reactiva** (regla 6 de la convivencia). Se
//     le dice así, que es más útil que un «no se puede».
func (s *Service) reactivarDeDirectorio(cuenta auth.Account) (auth.Account, error) {
	if cuenta.Origin == auth.OriginKeycloak {
		return auth.Account{}, ErrDirectoryActivatesItself
	}

	if s.directory == nil {
		// Un camino de AD sin directorio conectado: no hay a quién preguntar.
		return auth.Account{}, ErrDirectoryUnavailable
	}

	loConoce, err := s.directory.Knows(cuenta.Email)
	if err != nil {
		// El motivo va al log —el directorio caído, la cuenta de servicio mal puesta— y a la pantalla
		// le llega que ahora mismo no se puede comprobar, que es lo que entiende quien lo pide.
		logs.LogWarning("no se pudo comprobar con el directorio si la cuenta #" +
			strconv.FormatInt(cuenta.ID, 10) + " sigue allí: " + err.Error())
		return auth.Account{}, ErrDirectoryUnavailable
	}
	if !loConoce {
		return auth.Account{}, ErrDirectoryNotFound
	}

	if err := s.repo.UpdateFields(cuenta.ID, map[string]any{"is_active": true}); err != nil {
		return auth.Account{}, traducirUsuario(err)
	}

	cuenta.IsActive = true
	return cuenta, nil
}

// ChangeOrigin cambia el origen de una cuenta.
//
// Es de las acciones peligrosas: es una de las dos que pueden dejar a alguien sin poder entrar, y por
// eso tiene su propia ruta y su permiso (docs/modules/users.md, sección 4).
func (s *Service) ChangeOrigin(id int64, origen string, externalID string, actor auth.Identity) (auth.Account, error) {
	origen = strings.TrimSpace(origen)
	if !auth.OriginIsValid(origen) {
		return auth.Account{}, ErrOriginUnknown
	}

	cuenta, err := s.ByID(id)
	if err != nil {
		return auth.Account{}, err
	}

	if cuenta.Origin == origen {
		return cuenta, nil
	}

	// De una cuenta de directorio **activa** no se cambia el origen: allí sigue mandando el directorio,
	// y volvería a entrar por su camino (sección 5).
	if cuenta.IsDirectory() && cuenta.IsActive {
		return auth.Account{}, ErrOriginInUse
	}

	if origen != auth.OriginLocal {
		// Pasar a `ad` o `keycloak` tampoco se hace a mano, por lo mismo que el alta: **el origen de una
		// cuenta de directorio lo pone el directorio**, al entrar esa persona por su camino
		// (docs/modules/users.md, sección 5, punto 4).
		return auth.Account{}, ErrOriginByDirectory
	}

	if err := s.repo.UpdateFields(id, map[string]any{
		"origin":      origen,
		"external_id": nil,
	}); err != nil {
		return auth.Account{}, traducirUsuario(err)
	}

	// Al pasar a local, la cuenta se queda sin contraseña: se le manda el enlace para que establezca
	// una, que es la salida para no dejarla fuera.
	cuenta.Origin = auth.OriginLocal
	cuenta.ExternalID = ""

	if !cuenta.HasPassword() {
		if s.links == nil {
			return cuenta, errors.New("no hay módulo de contraseñas conectado: no se puede mandar el enlace")
		}
		if err := s.links.Invite(cuenta, actorID(actor)); err != nil {
			return cuenta, err
		}
	}

	return cuenta, nil
}

// traducirUsuario convierte los errores del repositorio en los del módulo.
func traducirUsuario(err error) error {
	if errors.Is(err, repositories.ErrUserNotFound) {
		return auth.ErrAccountNotFound
	}
	return err
}
