package services

import (
	"errors"
	"testing"

	"catalina-support/backend/shared/auth"
)

// cuentasStub es un `Accounts` de mentira para probar las menciones sin base de datos: sólo contesta
// lo que estas pruebas preguntan.
type cuentasStub map[int64]auth.Account

func (c cuentasStub) ByID(id int64) (auth.Account, error) {
	if cuenta, hay := c[id]; hay {
		return cuenta, nil
	}

	return auth.Account{}, errors.New("no está")
}

func (c cuentasStub) ByEmail(email string) (auth.Account, error) {
	for _, cuenta := range c {
		if cuenta.Email == email {
			return cuenta, nil
		}
	}

	return auth.Account{}, errors.New("no está")
}

func (c cuentasStub) ByIDs(ids []int64) (map[int64]auth.Account, error) {
	cuentas := map[int64]auth.Account{}
	for _, id := range ids {
		if cuenta, hay := c[id]; hay {
			cuentas[id] = cuenta
		}
	}

	return cuentas, nil
}

func (c cuentasStub) ActiveByRole(role string) ([]auth.Account, error) {
	var cuentas []auth.Account
	for _, cuenta := range c {
		if cuenta.IsActive && cuenta.Role == role {
			cuentas = append(cuentas, cuenta)
		}
	}

	return cuentas, nil
}

// TestMencionesDelCuerpo: se extraen los identificadores de las personas etiquetadas —una vez cada
// una, aunque se repitan, y en el orden en que aparecen—, y un cuerpo sin menciones no devuelve nada.
func TestMencionesDelCuerpo(t *testing.T) {
	sinMenciones := []string{
		``,
		`<p>Hola, mi PC no funciona</p>`,
		`<img data-adjunto="captura.png">`,
		// Un `data-mencion` que no vale no es una mención, aunque el saneador no lo haya dejado pasar.
		`<span data-mencion="0">x</span>`,
	}
	for _, cuerpo := range sinMenciones {
		if ids := MencionesDelCuerpo(cuerpo); len(ids) != 0 {
			t.Fatalf("%q no tiene menciones y salieron %v", cuerpo, ids)
		}
	}

	cuerpo := `<p><span data-mencion="12">María</span> mira esto, y tú <span data-mencion="7">Juan</span> también.</p><p>Otra vez <span data-mencion="12">María</span>.</p>`

	ids := MencionesDelCuerpo(cuerpo)
	quiere := []int64{12, 7}

	if len(ids) != len(quiere) {
		t.Fatalf("se esperaban %v y salieron %v", quiere, ids)
	}
	for i := range quiere {
		if ids[i] != quiere[i] {
			t.Fatalf("se esperaban %v y salieron %v", quiere, ids)
		}
	}
}

// TestDiferencia: las menciones **nuevas** son las que no estaban antes en ese mismo cuerpo, y son
// las únicas que vuelven a poner a alguien como observador (docs/modules/tickets.md, decisión 63).
func TestDiferencia(t *testing.T) {
	casos := []struct {
		nuevos     []int64
		anteriores []int64
		quiere     []int64
	}{
		// Un comentario nuevo: todas sus menciones son nuevas.
		{[]int64{12, 7}, nil, []int64{12, 7}},
		// Editarlo sin tocar la mención no la vuelve a añadir.
		{[]int64{12}, []int64{12}, nil},
		// Quitar una mención no añade nada.
		{[]int64{12}, []int64{12, 7}, nil},
		// Añadir una sí.
		{[]int64{12, 7}, []int64{12}, []int64{7}},
	}

	for _, caso := range casos {
		salida := diferencia(caso.nuevos, caso.anteriores)
		if len(salida) != len(caso.quiere) {
			t.Fatalf("diferencia(%v, %v) = %v, se esperaba %v", caso.nuevos, caso.anteriores, salida, caso.quiere)
		}
		for i := range caso.quiere {
			if salida[i] != caso.quiere[i] {
				t.Fatalf("diferencia(%v, %v) = %v, se esperaba %v", caso.nuevos, caso.anteriores, salida, caso.quiere)
			}
		}
	}
}

// TestValidarMenciones: etiquetan Soporte y Desarrollo, y sólo a técnicos y desarrolladores activos;
// una mención que no vale rechaza el cuerpo entero con `tickets.mention.notAllowed` (decisión 59).
func TestValidarMenciones(t *testing.T) {
	servicio := &Service{accounts: cuentasStub{
		2: {ID: 2, Role: auth.RoleDesarrollo, IsActive: true},
		5: {ID: 5, Role: auth.RoleSoporte, IsActive: false},
		6: {ID: 6, Role: auth.RoleUsuario, IsActive: true},
	}}

	casos := []struct {
		quien      auth.Identity
		menciones  []int64
		anteriores []int64
		esperado   bool
	}{
		// Soporte y Desarrollo etiquetan a un compañero activo.
		{soporte, []int64{2}, nil, true},
		{desarrollo, []int64{2}, nil, true},
		// Un usuario y el Administrador no etiquetan.
		{usuario, []int64{2}, nil, false},
		{administrador, []int64{2}, nil, false},
		// A una cuenta inactiva, a un usuario o a quien no existe, tampoco.
		{soporte, []int64{5}, nil, false},
		{soporte, []int64{6}, nil, false},
		{soporte, []int64{99}, nil, false},
		// Y un cuerpo sin menciones no comprueba nada: un usuario comenta como siempre.
		{usuario, nil, nil, true},
		// **Un usuario puede conservar una mención que no puso él** (2026-09-27): si el texto ya la
		// llevaba, guardarlo no es etiquetar a nadie. Lo que no puede es estrenar una.
		{usuario, []int64{2}, []int64{2}, true},
		{usuario, []int64{2, 7}, []int64{2}, false},
	}

	for _, caso := range casos {
		err := servicio.validarMenciones(caso.menciones, caso.anteriores, caso.quien)

		if caso.esperado && err != nil {
			t.Fatalf("el papel %s con %v (antes %v) debería valer y salió %v", caso.quien.Role, caso.menciones, caso.anteriores, err)
		}
		if !caso.esperado && !errors.Is(err, ErrMentionNotAllowed) {
			t.Fatalf("el papel %s con %v debería rechazarse con %v y salió %v", caso.quien.Role, caso.menciones, ErrMentionNotAllowed, err)
		}
	}
}
