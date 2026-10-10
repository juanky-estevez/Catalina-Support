package dtos

import (
	"encoding/json"
	"testing"

	"catalina-support/backend/modules/settings/services"
	"catalina-support/backend/shared/version"
)

// La marca pública lleva **la versión del sistema**, y llega al JSON con ese nombre
// (docs/modules/settings.md, sección 5.9).
//
// Se prueba aquí, en la traducción, porque es donde puede romperse en silencio: el dato sale del
// servicio y se queda por el camino si nadie lo copia al DTO. Lo que la instalación tenga en su base
// da igual para esto, y por eso no hace falta un repositorio.
func TestLaMarcaPublicaLlevaLaVersion(t *testing.T) {
	respuesta := NewBrandResponse(services.Public{
		Name:        "Mesa de ayuda de Acme",
		Version:     version.Version,
		License:     version.License,
		SourceURL:   version.SourceURL,
		Development: version.Development,
	})

	if respuesta.Version != version.Version {
		t.Fatalf("la marca debería llevar la versión %q y llevó %q", version.Version, respuesta.Version)
	}

	// Y en el JSON va como `version`, que es lo que lee el frontend.
	datos, err := json.Marshal(respuesta)
	if err != nil {
		t.Fatalf("no se pudo serializar la marca: %v", err)
	}

	var suelto map[string]any
	if err := json.Unmarshal(datos, &suelto); err != nil {
		t.Fatalf("no se pudo leer lo serializado: %v", err)
	}
	if suelto["version"] != version.Version {
		t.Fatalf("el JSON debería llevar `version` y llevó %v", suelto["version"])
	}
	if suelto["license"] != "AGPL-3.0-only" || suelto["sourceUrl"] != version.SourceURL {
		t.Fatalf("la marca debería identificar licencia y fuente, y llevó %v", suelto)
	}
	if suelto["development"] != true {
		t.Fatalf("el árbol abierto debería identificarse como desarrollo, y llevó %v", suelto["development"])
	}
}

// Y sin versión —una marca que no la trae— el campo va vacío y no se inventa nada: la pantalla
// entonces no enseña ningún número.
func TestLaMarcaPublicaSinVersionNoInventaNada(t *testing.T) {
	respuesta := NewBrandResponse(services.Public{Name: "Catalina Support"})

	if respuesta.Version != "" {
		t.Fatalf("sin versión el campo debería ir vacío y llevó %q", respuesta.Version)
	}
}
