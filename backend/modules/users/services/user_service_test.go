package services

import (
	"errors"
	"testing"

	"catalina-support/backend/shared/auth"
)

func TestEmailIsValid(t *testing.T) {
	validos := []string{
		"ana@ejemplo.com",
		"ana.perez@ejemplo.com",
		"ana+avisos@ejemplo.co",
		"a@b.co",
	}

	for _, correo := range validos {
		if !EmailIsValid(correo) {
			t.Fatalf("%q debería ser un correo válido", correo)
		}
	}

	invalidos := []string{
		"",
		"   ",
		"ana",
		"ana@",
		"@ejemplo.com",
		"ana@ejemplo",
		"ana arroba ejemplo.com",
		"ana@ejemplo .com",
		"Ana Pérez <ana@ejemplo.com>",
		"ana@ejemplo.com\n",
	}

	for _, correo := range invalidos {
		if EmailIsValid(correo) {
			t.Fatalf("%q no debería valer como correo", correo)
		}
	}
}

// Soporte sólo reparte `usuario`; el Administrador reparte los cuatro.
func TestCanAssignRole(t *testing.T) {
	administrador := auth.Identity{Role: auth.RoleAdministrador}
	soporte := auth.Identity{Role: auth.RoleSoporte}
	usuario := auth.Identity{Role: auth.RoleUsuario}

	for _, role := range auth.Roles() {
		if !CanAssignRole(administrador, role) {
			t.Fatalf("el administrador debería poder repartir %q", role)
		}
	}

	if !CanAssignRole(soporte, auth.RoleUsuario) {
		t.Fatal("soporte debería poder crear usuarios")
	}
	for _, role := range []string{auth.RoleSoporte, auth.RoleDesarrollo, auth.RoleAdministrador} {
		if CanAssignRole(soporte, role) {
			t.Fatalf("soporte no debería poder repartir %q", role)
		}
	}

	if CanAssignRole(usuario, auth.RoleUsuario) {
		t.Fatal("un usuario no da de alta a nadie")
	}
}

// Quién provoca la acción: la cuenta de fábrica no deja identificador, porque no está en la tabla.
func TestActorID(t *testing.T) {
	if obtenido := actorID(auth.Identity{ID: 5}); obtenido == nil || *obtenido != 5 {
		t.Fatalf("una cuenta normal deja su identificador; llegó %v", obtenido)
	}
	if obtenido := actorID(auth.FactoryIdentity()); obtenido != nil {
		t.Fatalf("la cuenta de fábrica no deja identificador; llegó %v", *obtenido)
	}
}

// Un alta sin nombre, con papel que no toca o con origen inventado no llega a la base: se corta antes
// de tocar nada. El repositorio es nil a propósito: si el servicio intentara usarlo, el test caería.
func TestCreateSeCortaAntesDeLaBase(t *testing.T) {
	servicio := NewService(nil)
	administrador := auth.Identity{ID: 1, Role: auth.RoleAdministrador}

	casos := []struct {
		nombre string
		input  CreateInput
		clave  error
	}{
		{
			"sin apellidos",
			CreateInput{Name: "Ana", Email: "ana@ejemplo.com", Role: auth.RoleUsuario, Origin: auth.OriginLocal},
			ErrNameRequired,
		},
		{
			"correo sin forma de correo",
			CreateInput{Name: "Ana", LastName: "Pérez", Email: "ana", Role: auth.RoleUsuario, Origin: auth.OriginLocal},
			ErrEmailInvalid,
		},
		{
			"papel inventado",
			CreateInput{Name: "Ana", LastName: "Pérez", Email: "ana@ejemplo.com", Role: "jefe", Origin: auth.OriginLocal},
			ErrRoleNotAllowed,
		},
		{
			"origen inventado",
			CreateInput{Name: "Ana", LastName: "Pérez", Email: "ana@ejemplo.com", Role: auth.RoleUsuario, Origin: "ldap"},
			ErrOriginUnknown,
		},
		{
			// Las cuentas de directorio **no se dan de alta a mano**: quien está en el directorio entra
			// solo y su cuenta nace en ese primer acceso (docs/modules/users.md, sección 5, punto 4).
			"alta de una cuenta de AD",
			CreateInput{Name: "Ana", LastName: "Pérez", Email: "ana@ejemplo.com", Role: auth.RoleUsuario, Origin: auth.OriginAD},
			ErrOriginByDirectory,
		},
		{
			"alta de una cuenta de Keycloak",
			CreateInput{Name: "Ana", LastName: "Pérez", Email: "ana@ejemplo.com", Role: auth.RoleUsuario, Origin: auth.OriginKeycloak},
			ErrOriginByDirectory,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			_, err := servicio.Create(caso.input, administrador)
			if err != caso.clave {
				t.Fatalf("se esperaba %v y llegó %v", caso.clave, err)
			}
		})
	}
}

// Soporte no puede dar de alta a otro soporte: se corta antes de tocar la base.
func TestSoporteNoRepartePapeles(t *testing.T) {
	servicio := NewService(nil)
	soporte := auth.Identity{ID: 2, Role: auth.RoleSoporte}

	_, err := servicio.Create(CreateInput{
		Name:     "Ana",
		LastName: "Pérez",
		Email:    "ana@ejemplo.com",
		Role:     auth.RoleSoporte,
		Origin:   auth.OriginLocal,
	}, soporte)

	if err != ErrRoleNotAllowed {
		t.Fatalf("se esperaba ErrRoleNotAllowed y llegó %v", err)
	}
}

// ---------------------------------------------------------------------------------------------
// Las acciones sobre una cuenta: lo que se corta antes de tocar la base.
// ---------------------------------------------------------------------------------------------

func TestNadieSeDesactivaASiMismo(t *testing.T) {
	// El repositorio es nil a propósito: la comprobación tiene que ocurrir antes.
	servicio := NewService(nil)

	_, _, err := servicio.Deactivate(7, auth.Identity{ID: 7, Role: auth.RoleSoporte})
	if !errors.Is(err, ErrSelfDeactivation) {
		t.Fatalf("se esperaba ErrSelfDeactivation y llegó %v", err)
	}

	// Desactivar a otra persona **sí pasa** esa comprobación. Con el repositorio en nil lo que viene
	// después es un pánico, y eso es justo lo que lo demuestra: si se hubiera cortado antes, no
	// habría llegado hasta ahí.
	defer func() {
		if recover() == nil {
			t.Fatal("desactivar a otra cuenta debería seguir adelante, no cortarse como si fuera la propia")
		}
	}()

	_, _, _ = servicio.Deactivate(8, auth.Identity{ID: 7, Role: auth.RoleSoporte})
}

func TestPatchDeSoporteSoloCambiaElNombre(t *testing.T) {
	servicio := NewService(nil)
	soporte := auth.Identity{ID: 2, Role: auth.RoleSoporte}

	// Lo que sí puede: el nombre. Se sabe porque **pasa la comprobación** y sigue adelante; con el
	// repositorio en nil, eso es un pánico, así que se recoge.
	func() {
		defer func() { _ = recover() }()

		nombre := "Ana"
		if _, err := servicio.Patch(7, PatchInput{Name: &nombre}, soporte); errors.Is(err, ErrRoleNotAllowed) {
			t.Fatal("Soporte sí puede cambiar el nombre y los apellidos")
		}
	}()

	// Lo que no puede: el correo, el papel ni el estado. El idioma personal ya no existe como
	// preferencia: los clientes antiguos pueden enviarlo, pero el servicio lo ignora.
	correo := "ana@ejemplo.com"
	papel := auth.RoleAdministrador
	activo := true

	for nombreDelCaso, entrada := range map[string]PatchInput{
		"el correo": {Email: &correo},
		"el papel":  {Role: &papel},
		"el estado": {IsActive: &activo},
	} {
		if _, err := servicio.Patch(7, entrada, soporte); !errors.Is(err, ErrRoleNotAllowed) {
			t.Fatalf("Soporte no debería poder cambiar %s, y llegó %v", nombreDelCaso, err)
		}
	}
}

func TestPerfilPropioNoCambiaLoAjeno(t *testing.T) {
	servicio := NewService(nil)
	usuario := auth.Identity{ID: 9, Role: auth.RoleUsuario}

	correo := "otro@ejemplo.com"
	if _, err := servicio.PatchOwn(PatchInput{Email: &correo}, usuario); !errors.Is(err, ErrRoleNotAllowed) {
		t.Fatalf("nadie cambia su propio correo y llegó %v", err)
	}

	papel := auth.RoleAdministrador
	if _, err := servicio.PatchOwn(PatchInput{Role: &papel}, usuario); !errors.Is(err, ErrRoleNotAllowed) {
		t.Fatal("nadie se asciende a sí mismo")
	}

	desactivar := false
	if _, err := servicio.PatchOwn(PatchInput{IsActive: &desactivar}, usuario); !errors.Is(err, ErrStateHasItsOwnAction) {
		t.Fatal("el estado se cambia con su propia acción, tampoco desde el perfil")
	}
}

// La cuenta de fábrica no está en la tabla: su perfil no existe.
func TestElPerfilDeLaCuentaDeFabricaNoExiste(t *testing.T) {
	servicio := NewService(nil)

	nombre := "Otro nombre"
	_, err := servicio.PatchOwn(PatchInput{Name: &nombre}, auth.FactoryIdentity())
	if !errors.Is(err, auth.ErrAccountNotFound) {
		t.Fatalf("se esperaba ErrAccountNotFound y llegó %v", err)
	}
}

// El estado no se cambia por el `PATCH` general: tiene su propia acción, para que no se desactive a
// nadie por accidente al editar su nombre. Y se rechaza **sin leer la cuenta**, que es lo que
// demuestra que el repositorio sea nil.
func TestPatchNoCambiaElEstado(t *testing.T) {
	servicio := NewService(nil)
	administrador := auth.Identity{ID: 1, Role: auth.RoleAdministrador}

	for nombre, activar := range map[string]bool{"activar": true, "desactivar": false} {
		valor := activar
		if _, err := servicio.Patch(7, PatchInput{IsActive: &valor}, administrador); !errors.Is(err, ErrStateHasItsOwnAction) {
			t.Fatalf("el PATCH no debería %s: se esperaba ErrStateHasItsOwnAction y llegó %v", nombre, err)
		}
	}
}

// ---------------------------------------------------------------------------------------------
// La entrada por el directorio (docs/modules/auth.md, sección 5.4).
// ---------------------------------------------------------------------------------------------

// Un dato del directorio que no vale se corta **antes de tocar la base**: el repositorio es nil a
// propósito, así que si el servicio intentara usarlo, el test caería.
func TestUpsertFromDirectorySeCortaAntesDeLaBase(t *testing.T) {
	servicio := NewService(nil)

	casos := []struct {
		nombre string
		datos  auth.DirectoryAccount
		clave  error
	}{
		{
			"sin correo",
			auth.DirectoryAccount{Origin: auth.OriginAD, ExternalID: "ana", Name: "Ana"},
			ErrEmailInvalid,
		},
		{
			"correo sin forma de correo",
			auth.DirectoryAccount{Origin: auth.OriginAD, ExternalID: "ana", Email: "ana"},
			ErrEmailInvalid,
		},
		{
			// Una cuenta del directorio no puede quedar como local: su contraseña es de allí.
			"origen local",
			auth.DirectoryAccount{Origin: auth.OriginLocal, ExternalID: "ana", Email: "ana@ejemplo.com"},
			ErrOriginUnknown,
		},
		{
			"origen inventado",
			auth.DirectoryAccount{Origin: "ldap", ExternalID: "ana", Email: "ana@ejemplo.com"},
			ErrOriginUnknown,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if _, err := servicio.UpsertFromDirectory(caso.datos); !errors.Is(err, caso.clave) {
				t.Fatalf("se esperaba %v y llegó %v", caso.clave, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------------------------
// Reactivar una cuenta de directorio, que es lo que el documento pedía comprobar (sección 5, 4).
// ---------------------------------------------------------------------------------------------

// directorioDeMentira es el directorio visto desde `users`: contesta si sigue conociendo a alguien.
type directorioDeMentira struct {
	loConoce bool
	falla    error
	// preguntas son los correos por los que se le ha preguntado: sirve para comprobar que **no se le
	// pregunta** cuando no hay nada que preguntar.
	preguntas []string
}

func (d *directorioDeMentira) Knows(email string) (bool, error) {
	d.preguntas = append(d.preguntas, email)
	if d.falla != nil {
		return false, d.falla
	}

	return d.loConoce, nil
}

func TestReactivarUnaCuentaDeDirectorio(t *testing.T) {
	casos := []struct {
		nombre     string
		cuenta     auth.Account
		directorio Directory
		clave      error
	}{
		{
			// Keycloak **no se pregunta**: hacerlo pediría una cuenta de servicio en el reino, y esa
			// cuenta se reactiva sola al entrar por su camino.
			"una cuenta de Keycloak se reactiva sola al entrar",
			auth.Account{ID: 3, Email: "ana@ejemplo.com", Origin: auth.OriginKeycloak},
			&directorioDeMentira{loConoce: true},
			ErrDirectoryActivatesItself,
		},
		{
			"una cuenta de AD sin directorio conectado",
			auth.Account{ID: 4, Email: "ana@ejemplo.com", Origin: auth.OriginAD},
			nil,
			ErrDirectoryUnavailable,
		},
		{
			"el directorio no responde",
			auth.Account{ID: 4, Email: "ana@ejemplo.com", Origin: auth.OriginAD},
			&directorioDeMentira{falla: errors.New("el directorio no responde")},
			ErrDirectoryUnavailable,
		},
		{
			// Devolverle el acceso a quien ya no está allí sería devolverle un acceso que no puede usar.
			"el directorio ya no la conoce",
			auth.Account{ID: 4, Email: "ana@ejemplo.com", Origin: auth.OriginAD},
			&directorioDeMentira{loConoce: false},
			ErrDirectoryNotFound,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			servicio := NewService(nil)
			if caso.directorio != nil {
				servicio.SetDirectory(caso.directorio)
			}

			if _, err := servicio.reactivarDeDirectorio(caso.cuenta); !errors.Is(err, caso.clave) {
				t.Fatalf("se esperaba %v y llegó %v", caso.clave, err)
			}
		})
	}

	// Y a una cuenta de Keycloak no se le pregunta al directorio, porque no hay nada que preguntar.
	servicio := NewService(nil)
	directorio := &directorioDeMentira{loConoce: true}
	servicio.SetDirectory(directorio)

	if _, err := servicio.reactivarDeDirectorio(auth.Account{ID: 3, Email: "ana@ejemplo.com", Origin: auth.OriginKeycloak}); !errors.Is(err, ErrDirectoryActivatesItself) {
		t.Fatalf("se esperaba ErrDirectoryActivatesItself y llegó %v", err)
	}
	if len(directorio.preguntas) != 0 {
		t.Fatalf("no se debería haber preguntado al directorio y se preguntó por %v", directorio.preguntas)
	}
}
