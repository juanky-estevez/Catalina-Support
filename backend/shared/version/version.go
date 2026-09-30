// Package version dice qué versión es este software.
//
// **Es una constante escrita a mano y no la fecha de compilación**: una fecha cambia en cada
// compilación sin decir nada nuevo, y dos instalaciones con el mismo código dirían cosas distintas.
// Lo que hace falta saber —qué versión está puesta— lo dice el número.
//
// **Se cambia al cerrar una versión**, en el mismo trabajo en que se etiqueta el repositorio: la
// versión de la aplicación y la del esquema de la base son la misma cosa, y `1.0.0` es la que nombra
// `backend/migrations/v1.0.0.sql` (docs/ambientes.md, sección 5).
//
// Vive en `shared` porque lo lee el módulo que publica la marca y lo puede leer cualquiera: no depende
// de nadie y nadie depende de él (docs/arquitectura.md, sección 4).
package version

// Version es la versión de esta instalación, sin la `v`: la `v` la pone la interfaz al enseñarla
// (docs/modules/settings.md, sección 5.9).
const Version = "1.0.0"
