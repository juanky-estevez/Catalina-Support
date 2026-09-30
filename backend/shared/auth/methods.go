package auth

// Los métodos de entrada de la instalación. **Uno a la vez**: el que está puesto es el único que se
// ofrece, y los otros dos quedan apagados (docs/modules/settings.md, sección 5.8).
//
// Viven aquí, en `shared`, por la misma razón que los orígenes: los usa `settings` para guardarlos y
// validarlos y los usa `auth` para entrar, y ninguno de los dos puede importar al otro. Son **un valor
// cerrado que viaja**, no un texto suelto.
const (
	// MethodLocal es el de fábrica: las cuentas viven en la base y se entra con correo y contraseña.
	MethodLocal = "local"
	// MethodAD delega en el directorio de la organización: la contraseña es la de la empresa.
	MethodAD = "ad"
	// MethodKeycloak delega en Keycloak: se entra por su pantalla, y aquí no hay contraseña que
	// comparar.
	MethodKeycloak = "keycloak"
)

// Methods devuelve los tres métodos, para validar lo que llega de fuera.
func Methods() []string { return []string{MethodLocal, MethodAD, MethodKeycloak} }

// MethodIsValid dice si el método es uno de los tres.
func MethodIsValid(method string) bool {
	for _, known := range Methods() {
		if method == known {
			return true
		}
	}

	return false
}
