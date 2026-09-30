package services

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"io"
	"path/filepath"
	"strings"

	// Los decodificadores se registran solos: son los que permiten leer el tamaño de la imagen sin
	// dependencias externas más allá del WebP, que no viene en la biblioteca estándar.
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

// Los límites del logo (docs/modules/settings.md, sección 5.2). Un logo se enseña a 96 px de alto
// como mucho: más de esto no aporta nada y engorda cada copia de seguridad.
const (
	MaxLogoBytes = 1 << 20 // 1 MB
	MaxLogoLado  = 2000    // píxeles
)

// Los formatos que se aceptan, con su extensión y su tipo.
var formatos = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
	// El SVG no es un mapa de bits: va aparte y se mide de otra forma.
	"image/svg+xml": ".svg",
}

// ErrLogoFormat se devuelve cuando el archivo no es de ninguno de los formatos aceptados.
var ErrLogoFormat = errors.New("settings.logo.format")

// ErrLogoInvalid se devuelve cuando el contenido no es una imagen válida o pasa del lado máximo.
var ErrLogoInvalid = errors.New("settings.logo.invalid")

// ErrLogoTooBig se devuelve cuando el archivo pesa más de lo permitido.
var ErrLogoTooBig = errors.New("settings.logo.tooBig")

// Logo es un archivo de logo ya validado, listo para guardar.
type Logo struct {
	// Bytes son los primeros bytes del archivo, que es lo que decide el formato de verdad.
	Bytes []byte
	// ContentType es el tipo real, deducido del contenido y no de la extensión.
	ContentType string
	// Extension es la que se le pone al archivo guardado, según el tipo real.
	Extension string
	// Width y Height son las medidas. En un SVG son las de su `viewBox` o sus atributos.
	Width  int
	Height int
}

// InspectLogo lee el principio del archivo y dice si vale, qué es y de qué tamaño.
//
// **Se mira el contenido, no la extensión**: un `.png` que por dentro es otra cosa se rechaza. Y se
// mira el tamaño: un archivo que pasa del límite se corta antes de leerlo entero, para no gastar
// memoria en algo que se va a rechazar.
func InspectLogo(entrada io.Reader) (Logo, error) {
	// Un byte de más que el límite: si llega hasta ahí, es que se pasa.
	cabeza, err := io.ReadAll(io.LimitReader(entrada, MaxLogoBytes+1))
	if err != nil {
		return Logo{}, ErrLogoInvalid
	}

	if len(cabeza) > MaxLogoBytes {
		return Logo{}, ErrLogoTooBig
	}
	if len(cabeza) == 0 {
		return Logo{}, ErrLogoInvalid
	}

	tipo := tipoDeContenido(cabeza)

	extension, aceptado := formatos[tipo]
	if !aceptado {
		return Logo{}, ErrLogoFormat
	}

	ancho, alto, err := medidas(tipo, cabeza)
	if err != nil {
		return Logo{}, err
	}

	if ancho > MaxLogoLado || alto > MaxLogoLado {
		return Logo{}, fmt.Errorf("%w: %dx%d pasa del lado máximo (%d)", ErrLogoInvalid, ancho, alto, MaxLogoLado)
	}

	return Logo{
		Bytes:       cabeza,
		ContentType: tipo,
		Extension:   extension,
		Width:       ancho,
		Height:      alto,
	}, nil
}

// tipoDeContenido mira los primeros bytes, que es lo único que no se puede falsear.
func tipoDeContenido(datos []byte) string {
	switch {
	case bytes.HasPrefix(datos, []byte("\x89PNG\r\n\x1a\n")):
		return "image/png"
	case bytes.HasPrefix(datos, []byte("\xff\xd8\xff")):
		return "image/jpeg"
	case len(datos) > 12 && bytes.Equal(datos[0:4], []byte("RIFF")) && bytes.Equal(datos[8:12], []byte("WEBP")):
		return "image/webp"
	case pareceSVG(datos):
		return "image/svg+xml"
	default:
		return ""
	}
}

// pareceSVG busca la etiqueta `<svg` en el principio del archivo, saltándose la declaración XML, los
// comentarios y los espacios. No se valida el XML entero: no hace falta, porque el SVG **no se
// interpreta en el servidor** —se sirve aislado y el navegador lo usa como imagen—.
func pareceSVG(datos []byte) bool {
	texto := strings.ToLower(string(datos))
	indice := strings.Index(texto, "<svg")
	if indice < 0 {
		return false
	}

	// Sólo vale si antes de la etiqueta no hay nada más que la declaración, comentarios o espacios.
	prefijo := strings.TrimSpace(texto[:indice])
	if prefijo == "" {
		return true
	}

	if strings.HasPrefix(prefijo, "<?xml") && strings.HasSuffix(prefijo, "?>") {
		return true
	}

	return strings.HasPrefix(prefijo, "<!--") && strings.HasSuffix(prefijo, "-->")
}

// medidas saca el tamaño de la imagen. Para los mapas de bits lo lee el decodificador; para el SVG,
// sus atributos o su `viewBox`.
func medidas(tipo string, cabeza []byte) (int, int, error) {
	if tipo == "image/svg+xml" {
		return medidasDelSVG(cabeza)
	}

	configuracion, _, err := image.DecodeConfig(bytes.NewReader(cabeza))
	if err != nil {
		return 0, 0, fmt.Errorf("%w: no se pudo leer la imagen", ErrLogoInvalid)
	}

	return configuracion.Width, configuracion.Height, nil
}

// svgRaiz son los atributos del `<svg>` que dicen el tamaño.
type svgRaiz struct {
	Width   string `xml:"width,attr"`
	Height  string `xml:"height,attr"`
	ViewBox string `xml:"viewBox,attr"`
}

// medidasDelSVG lee el ancho y el alto del `<svg>`. Si no los trae, usa el `viewBox`, que es lo que
// define el sistema de coordenadas.
func medidasDelSVG(datos []byte) (int, int, error) {
	var raiz svgRaiz
	decodificador := xml.NewDecoder(bytes.NewReader(datos))

	for {
		token, err := decodificador.Token()
		if err != nil {
			return 0, 0, fmt.Errorf("%w: el SVG no se pudo leer", ErrLogoInvalid)
		}

		if inicio, ok := token.(xml.StartElement); ok {
			if strings.EqualFold(inicio.Name.Local, "svg") {
				if err := decodificador.DecodeElement(&raiz, &inicio); err != nil {
					return 0, 0, fmt.Errorf("%w: el SVG no se pudo leer", ErrLogoInvalid)
				}
				break
			}
		}
	}

	ancho, alto := numero(raiz.Width), numero(raiz.Height)
	if ancho > 0 && alto > 0 {
		return ancho, alto, nil
	}

	// `viewBox="0 0 120 40"`: el tercero y el cuarto son el ancho y el alto.
	campos := strings.Fields(raiz.ViewBox)
	if len(campos) == 4 {
		ancho, alto = numero(campos[2]), numero(campos[3])
		if ancho > 0 && alto > 0 {
			return ancho, alto, nil
		}
	}

	// Un SVG sin medidas ni `viewBox` es un SVG al que no se le puede poner un límite de tamaño: se
	// acepta, porque se enseña a un alto máximo y no rompe nada.
	return 1, 1, nil
}

// numero lee un número que puede venir con unidades («120px», «12.5»), y devuelve 0 si no lo es.
func numero(valor string) int {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		return 0
	}

	corte := len(valor)
	for i, letra := range valor {
		if (letra < '0' || letra > '9') && letra != '.' {
			corte = i
			break
		}
	}

	var resultado float64
	if _, err := fmt.Sscanf(valor[:corte], "%f", &resultado); err != nil {
		return 0
	}

	return int(resultado)
}

// ExtensionDe es la extensión que le toca a ese tipo, o cadena vacía si no lo conocemos.
func ExtensionDe(tipo string) string {
	return formatos[tipo]
}

// NombreDeLogo arma el nombre del archivo guardado. **Lo pone la aplicación, no quien sube**:
// así no hay dos archivos con el mismo nombre, no se puede colar una ruta y se sabe de cuándo es.
func NombreDeLogo(variante, marca string, extension string) string {
	return fmt.Sprintf("logo-%s-%s%s", variante, marca, extension)
}

// TipoDeLogo deduce el tipo por la extensión del archivo guardado, que es de fiar porque **la puso
// la aplicación después de validar el contenido**.
func TipoDeLogo(nombre string) string {
	switch strings.ToLower(filepath.Ext(nombre)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
