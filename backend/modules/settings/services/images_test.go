package services

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
)

// Un PNG de verdad, del tamaño que se le pida, para probar la validación con algo real.
func pngDe(ancho, alto int) []byte {
	imagen := image.NewRGBA(image.Rect(0, 0, ancho, alto))
	imagen.Set(0, 0, color.RGBA{R: 255, A: 255})

	var salida bytes.Buffer
	if err := png.Encode(&salida, imagen); err != nil {
		panic(err)
	}

	return salida.Bytes()
}

// Un JPEG de verdad, del tamaño que se le pida: el contenido se valida de verdad, no sólo su firma.
func jpegDe(ancho, alto int) []byte {
	imagen := image.NewRGBA(image.Rect(0, 0, ancho, alto))
	imagen.Set(0, 0, color.RGBA{G: 255, A: 255})

	var salida bytes.Buffer
	if err := jpeg.Encode(&salida, imagen, nil); err != nil {
		panic(err)
	}

	return salida.Bytes()
}

// Un WebP de verdad, de 1x1: no hay codificador en la biblioteca estándar, así que sus bytes van
// aquí. Un WebP inventado no valdría: lo que se valida es que la imagen se pueda leer.
var webpMinimo, _ = base64.StdEncoding.DecodeString(
	"UklGRiIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEAAUAmJaQAA3AA/vuUAAA=",
)

func TestInspectLogoAceptaLosCuatroFormatos(t *testing.T) {
	casos := []struct {
		nombre      string
		datos       []byte
		contentType string
		extension   string
	}{
		{"png", pngDe(10, 10), "image/png", ".png"},
		{"jpeg", jpegDe(10, 10), "image/jpeg", ".jpg"},
		{"webp", webpMinimo, "image/webp", ".webp"},
		{
			"svg",
			[]byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" width="120" height="40"><rect/></svg>`),
			"image/svg+xml",
			".svg",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			logo, err := InspectLogo(bytes.NewReader(caso.datos))
			if err != nil {
				t.Fatalf("debería aceptarlo: %v", err)
			}
			if logo.ContentType != caso.contentType {
				t.Fatalf("el tipo es %q y se esperaba %q", logo.ContentType, caso.contentType)
			}
			if logo.Extension != caso.extension {
				t.Fatalf("la extensión es %q y se esperaba %q", logo.Extension, caso.extension)
			}
		})
	}
}

// Se mira el contenido, no la extensión: un archivo con nombre de imagen y contenido que no lo es se
// rechaza. Y al revés: un PNG con nombre raro se acepta, porque el nombre lo pone la aplicación.
func TestInspectLogoMiraElContenido(t *testing.T) {
	_, err := InspectLogo(strings.NewReader("<?php echo 'hola'; ?>"))
	if !errors.Is(err, ErrLogoFormat) {
		t.Fatalf("se esperaba ErrLogoFormat y llegó %v", err)
	}

	if _, err := InspectLogo(bytes.NewReader(pngDe(4, 4))); err != nil {
		t.Fatalf("un PNG válido debería valer: %v", err)
	}
}

func TestInspectLogoRechazaLoQuePasa(t *testing.T) {
	// Peso: un byte más del límite.
	grande := make([]byte, MaxLogoBytes+1)
	copy(grande, pngDe(2, 2))

	if _, err := InspectLogo(bytes.NewReader(grande)); !errors.Is(err, ErrLogoTooBig) {
		t.Fatalf("se esperaba ErrLogoTooBig y llegó %v", err)
	}

	// Lado: una imagen válida pero más ancha de lo permitido.
	if _, err := InspectLogo(bytes.NewReader(pngDe(MaxLogoLado+1, 10))); !errors.Is(err, ErrLogoInvalid) {
		t.Fatalf("se esperaba ErrLogoInvalid y llegó %v", err)
	}

	// Y un archivo vacío no es una imagen.
	if _, err := InspectLogo(bytes.NewReader(nil)); !errors.Is(err, ErrLogoInvalid) {
		t.Fatalf("se esperaba ErrLogoInvalid y llegó %v", err)
	}
}

func TestInspectLogoLeeElTamanio(t *testing.T) {
	logo, err := InspectLogo(bytes.NewReader(pngDe(300, 120)))
	if err != nil {
		t.Fatalf("no se pudo leer: %v", err)
	}

	if logo.Width != 300 || logo.Height != 120 {
		t.Fatalf("el tamaño es %dx%d y se esperaba 300x120", logo.Width, logo.Height)
	}
}

// El SVG se mide por sus atributos o por su `viewBox`, y uno sin ninguno de los dos se acepta.
func TestMedidasDelSVG(t *testing.T) {
	casos := []struct {
		nombre string
		svg    string
		ancho  int
		alto   int
	}{
		{"con width y height", `<svg width="120" height="40"></svg>`, 120, 40},
		{"con unidades", `<svg width="120px" height="40px"></svg>`, 120, 40},
		{"sólo viewBox", `<svg viewBox="0 0 120 40"></svg>`, 120, 40},
		{"con declaración y comentario", "<?xml version=\"1.0\"?><!-- hecho a mano --><svg width=\"64\" height=\"64\"></svg>", 64, 64},
		{"sin medidas", `<svg></svg>`, 1, 1},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			ancho, alto, err := medidasDelSVG([]byte(caso.svg))
			if err != nil {
				t.Fatalf("no se pudo medir: %v", err)
			}
			if ancho != caso.ancho || alto != caso.alto {
				t.Fatalf("mide %dx%d y se esperaba %dx%d", ancho, alto, caso.ancho, caso.alto)
			}
		})
	}
}

// Un SVG con XML roto se rechaza: no se puede medir, y no se acepta lo que no se entiende.
func TestMedidasDelSVGFallaConXMLRoto(t *testing.T) {
	if _, _, err := medidasDelSVG([]byte("<svg width=\"10\"")); err == nil {
		t.Fatal("un SVG sin cerrar no debería medirse")
	}
}

func TestNombreDeLogo(t *testing.T) {
	nombre := NombreDeLogo(VarianteClaro, "20260924T181500", ".png")

	if nombre != "logo-claro-20260924T181500.png" {
		t.Fatalf("el nombre es %q", nombre)
	}
	// El nombre lo pone la aplicación: no lleva nada de lo que mandó quien subió el archivo.
	if strings.Contains(nombre, "/") || strings.Contains(nombre, "..") {
		t.Fatalf("el nombre no puede llevar rutas: %q", nombre)
	}
}

func TestTipoDeLogo(t *testing.T) {
	casos := map[string]string{
		"logo-claro-1.png":  "image/png",
		"logo-claro-1.jpg":  "image/jpeg",
		"logo-claro-1.webp": "image/webp",
		"logo-claro-1.svg":  "image/svg+xml",
		"logo-claro-1.bin":  "application/octet-stream",
	}

	for nombre, esperado := range casos {
		if obtenido := TipoDeLogo(nombre); obtenido != esperado {
			t.Fatalf("TipoDeLogo(%q) = %q y se esperaba %q", nombre, obtenido, esperado)
		}
	}
}
