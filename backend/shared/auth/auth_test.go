package auth

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// El camino feliz: lo que se firma se puede leer, y el `sub` vuelve intacto.
func TestSignAndVerifyDevuelvenElSujeto(t *testing.T) {
	manager, err := NewTokenManager("un-secreto-de-pruebas")
	if err != nil {
		t.Fatalf("no se pudo construir el firmante: %v", err)
	}

	now := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)
	token, err := manager.Sign("42", now)
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	subject, err := manager.Verify(token, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("no se pudo validar: %v", err)
	}
	if subject != "42" {
		t.Fatalf("el sujeto es %q y se esperaba %q", subject, "42")
	}
}

// El token de la cuenta de fábrica también: su `sub` no es un identificador de la base.
func TestSignAceptaElSujetoDeFabrica(t *testing.T) {
	manager, _ := NewTokenManager("un-secreto-de-pruebas")
	now := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)

	token, err := manager.Sign(FactorySubject, now)
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	subject, err := manager.Verify(token, now)
	if err != nil {
		t.Fatalf("no se pudo validar: %v", err)
	}
	if subject != FactorySubject {
		t.Fatalf("el sujeto es %q y se esperaba %q", subject, FactorySubject)
	}
}

// A las diez horas justas el token ya no vale: es la única puerta que se cierra sola, porque no hay
// tabla de sesiones que revoque (docs/modules/auth.md, sección 3).
func TestVerifyCaducaALasDiezHoras(t *testing.T) {
	manager, _ := NewTokenManager("un-secreto-de-pruebas")
	now := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)

	token, err := manager.Sign("42", now)
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	if _, err := manager.Verify(token, now.Add(SessionDuration-time.Minute)); err != nil {
		t.Fatalf("todavía debería valer: %v", err)
	}

	if _, err := manager.Verify(token, now.Add(SessionDuration)); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("se esperaba ErrTokenExpired y llegó %v", err)
	}
}

// Un token firmado con otro secreto no entra, aunque su forma sea perfecta.
func TestVerifyRechazaOtroSecreto(t *testing.T) {
	ajeno, _ := NewTokenManager("el-secreto-de-otro")
	nuestro, _ := NewTokenManager("un-secreto-de-pruebas")
	now := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)

	token, err := ajeno.Sign("42", now)
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	if _, err := nuestro.Verify(token, now); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("se esperaba ErrTokenInvalid y llegó %v", err)
	}
}

// El ataque de toda la vida: cambiar el algoritmo a "none" y quitar la firma. No debe colar.
func TestVerifyRechazaElAlgoritmoNone(t *testing.T) {
	manager, _ := NewTokenManager("un-secreto-de-pruebas")
	now := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)

	// {"alg":"none","typ":"JWT"} . {"sub":"1","exp":9999999999} . sin firma
	sinFirma := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxIiwiZXhwIjo5OTk5OTk5OTk5fQ."

	if _, err := manager.Verify(sinFirma, now); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("se esperaba ErrTokenInvalid y llegó %v", err)
	}
}

// Un token sin `sub` no identifica a nadie: no vale, aunque la firma sea buena.
func TestVerifyRechazaUnTokenSinSujeto(t *testing.T) {
	manager, _ := NewTokenManager("un-secreto-de-pruebas")
	now := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)

	token, err := manager.Sign("", now)
	if err != nil {
		t.Fatalf("no se pudo firmar: %v", err)
	}

	if _, err := manager.Verify(token, now); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("se esperaba ErrTokenInvalid y llegó %v", err)
	}
}

// Sin secreto no se firma: en producción esto es lo que impide arrancar sin `TOKEN_SECRET`.
func TestNewTokenManagerExigeSecreto(t *testing.T) {
	for _, vacio := range []string{"", "   "} {
		if _, err := NewTokenManager(vacio); !errors.Is(err, ErrNoSecret) {
			t.Fatalf("con %q se esperaba ErrNoSecret y llegó %v", vacio, err)
		}
	}
}

func TestBearerToken(t *testing.T) {
	casos := []struct {
		nombre   string
		cabecera string
		esperado string
		hay      bool
	}{
		{"normal", "Bearer abc.def.ghi", "abc.def.ghi", true},
		{"en minúsculas", "bearer abc.def.ghi", "abc.def.ghi", true},
		{"con espacios de más", "Bearer   abc.def.ghi  ", "abc.def.ghi", true},
		{"vacía", "", "", false},
		{"sin esquema", "abc.def.ghi", "", false},
		{"otro esquema", "Basic abc.def.ghi", "", false},
		{"sólo el esquema", "Bearer ", "", false},
		{"sólo espacios", "Bearer    ", "", false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			token, ok := BearerToken(caso.cabecera)
			if ok != caso.hay {
				t.Fatalf("se esperaba hay=%v y llegó %v", caso.hay, ok)
			}
			if token != caso.esperado {
				t.Fatalf("el token es %q y se esperaba %q", token, caso.esperado)
			}
		})
	}
}

func TestSameSecret(t *testing.T) {
	if !SameSecret("C4t4l1n4", "C4t4l1n4") {
		t.Fatal("dos secretos iguales deben coincidir")
	}
	if SameSecret("C4t4l1n4", "C4t4l1n5") {
		t.Fatal("dos secretos distintos no deben coincidir")
	}
	if SameSecret("C4t4l1n4", "C4t4l1n4 ") {
		t.Fatal("un espacio de más cambia el secreto")
	}
	if SameSecret("", "") {
		t.Fatal("dos secretos vacíos no deben darse por buenos")
	}
}

func TestIdentityCan(t *testing.T) {
	soporte := Identity{Role: RoleSoporte}

	if !soporte.Can(RoleSoporte) {
		t.Fatal("soporte debe poder lo suyo")
	}
	if !soporte.Can(RoleAdministrador, RoleSoporte) {
		t.Fatal("basta con que uno de los papeles coincida")
	}
	if soporte.Can(RoleAdministrador) {
		t.Fatal("soporte no debe poder lo del administrador")
	}
}

func TestFullName(t *testing.T) {
	casos := []struct {
		identity Identity
		esperado string
	}{
		{Identity{Name: "Ana", LastName: "Pérez"}, "Ana Pérez"},
		{Identity{Name: "Ana"}, "Ana"},
		{Identity{LastName: "Pérez"}, "Pérez"},
		{Identity{}, ""},
	}

	for _, caso := range casos {
		if obtenido := caso.identity.FullName(); obtenido != caso.esperado {
			t.Fatalf("FullName() = %q y se esperaba %q", obtenido, caso.esperado)
		}
	}
}

func TestRolesCerrados(t *testing.T) {
	for _, role := range Roles() {
		if !RoleIsValid(role) {
			t.Fatalf("%q debería ser un papel válido", role)
		}
	}

	for _, inventado := range []string{"", "admin", "Administrador", "superusuario"} {
		if RoleIsValid(inventado) {
			t.Fatalf("%q no es uno de los cuatro papeles", inventado)
		}
	}

	if len(Roles()) != 4 {
		t.Fatalf("hay %d papeles y deben ser cuatro", len(Roles()))
	}
}

// La identidad viaja en el contexto y se recupera entera.
func TestIdentidadEnElContexto(t *testing.T) {
	original := Identity{ID: 7, Name: "Ana", Role: RoleSoporte}
	ctx := WithIdentity(t.Context(), original)

	recuperada, ok := FromContext(ctx)
	if !ok {
		t.Fatal("debería haber identidad en el contexto")
	}
	if recuperada != original {
		t.Fatalf("la identidad volvió distinta: %+v", recuperada)
	}

	if _, ok := FromContext(t.Context()); ok {
		t.Fatal("en un contexto limpio no debe haber identidad")
	}
}

// MustFromContext es para rutas protegidas: sin middleware delante, avisa en vez de seguir con una
// identidad vacía que daría permisos a nadie.
func TestMustFromContextAvisaSinIdentidad(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("debería haber avisado con un pánico")
		}
	}()

	_ = MustFromContext(t.Context())
}

// Los nombres de rol son cadenas cerradas que también viajan a la base y al frontend: si alguien los
// cambia aquí sin cambiarlos allí, esto lo enseña.
func TestNombresDeRol(t *testing.T) {
	esperados := []string{"usuario", "soporte", "desarrollo", "administrador"}
	for i, rol := range Roles() {
		if rol != esperados[i] {
			t.Fatalf("el papel %d es %q y debería ser %q", i, rol, esperados[i])
		}
		if strings.TrimSpace(rol) != rol {
			t.Fatalf("el papel %q lleva espacios", rol)
		}
	}
}
