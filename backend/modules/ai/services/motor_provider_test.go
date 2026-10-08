package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestLosCuatroAdaptadoresUsanSuContratoSinSalirALaRed(t *testing.T) {
	for _, provider := range []string{"openai", "deepseek", "openai-compatible", "claude"} {
		t.Run(provider, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if provider == "claude" {
					if r.URL.Path != "/v1/messages" || r.Header.Get("x-api-key") != "secreto" {
						t.Fatalf("contrato Claude incorrecto: %s, cabecera=%q", r.URL.Path, r.Header.Get("x-api-key"))
					}
					fmt.Fprint(w, `{"content":[{"text":"{\"text\":\"ok\"}"}]}`)
					return
				}
				if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer secreto" {
					t.Fatalf("contrato compatible incorrecto: %s, cabecera=%q", r.URL.Path, r.Header.Get("Authorization"))
				}
				fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"text\":\"ok\"}"}}]}`)
			}))
			defer server.Close()

			m := nuevoMotor(time.Second)
			if err := m.probarConfiguracion(context.Background(), server.URL, "modelo", provider, "bearer", "", "secreto", "es"); err != nil {
				t.Fatalf("el adaptador %s rechazó una respuesta válida: %v", provider, err)
			}
			text, _, err := m.pedirGlobal(context.Background(), Ajustes{
				URL: server.URL, Modelo: "modelo", Provider: provider, AuthType: "bearer",
				Credential: "secreto", Language: "es", Palabras: 60, Contexto: 4096,
			}, encargo{sistema: "resume", texto: "ticket falso"})
			if err != nil || text != "ok" {
				t.Fatalf("el resumen de %s no usó su adaptador: texto=%q err=%v", provider, text, err)
			}
		})
	}
}

func TestUnaRedireccionNoReenviaLaCredencial(t *testing.T) {
	var destinoLlamado atomic.Bool
	destino := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		destinoLlamado.Store(true)
	}))
	defer destino.Close()
	origen := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", destino.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer origen.Close()

	m := nuevoMotor(time.Second)
	err := m.probarConfiguracion(context.Background(), origen.URL, "modelo", "openai", "bearer", "", "no-debe-salir", "es")
	if !errors.Is(err, ErrSinMotor) || destinoLlamado.Load() {
		t.Fatalf("la redirección debía rechazarse sin llamar al destino: err=%v llamado=%v", err, destinoLlamado.Load())
	}
}

func TestUnaRespuestaMayorAlLimiteSeRechaza(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"text\":\"ok\"}"}}]}`+strings.Repeat(" ", maxRespuestaBytes))
	}))
	defer server.Close()

	m := nuevoMotor(time.Second)
	err := m.probarConfiguracion(context.Background(), server.URL, "modelo", "openai", "none", "", "", "es")
	if !errors.Is(err, ErrRespuestaInvalida) {
		t.Fatalf("se esperaba respuesta inválida por tamaño y llegó %v", err)
	}
}

func TestMejorarRedaccionUsaContextoMinimoYDevuelveTextoPlano(t *testing.T) {
	var recibido string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		recibido = string(data)
		fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"text\":\"Hola, Ana. Ya puedes ingresar.\"}"}}]}`)
	}))
	defer server.Close()

	servicio := nuevoServicio(nil, Ajustes{
		URL: server.URL, Modelo: "modelo", Provider: "openai-compatible", Language: "es",
		Palabras: 40, Contexto: 4096, Espera: time.Second,
	})
	defer servicio.Parar()

	texto, err := servicio.MejorarRedaccion(
		"hola ana ya puedes ingresar", "friendly", "Ana Pérez", "main", "es",
	)
	if err != nil || texto != "Hola, Ana. Ya puedes ingresar." {
		t.Fatalf("resultado inesperado: texto=%q err=%v", texto, err)
	}
	for _, esperado := range []string{"hola ana ya puedes ingresar", "Ana Pérez", "friendly and approachable", "main support ticket"} {
		if !strings.Contains(recibido, esperado) {
			t.Fatalf("la petición no contiene %q: %s", esperado, recibido)
		}
	}
	for _, prohibido := range []string{"attachment", "conversation", "subject"} {
		if strings.Contains(strings.ToLower(recibido), prohibido) {
			t.Fatalf("la petición incluyó contexto no autorizado %q", prohibido)
		}
	}
}

func TestConfiguradoNoCompruebaLaSaludDelMotor(t *testing.T) {
	servicio := nuevoServicio(nil, Ajustes{URL: "http://127.0.0.1:1", Modelo: "modelo"})
	defer servicio.Parar()
	if !servicio.Configurado() {
		t.Fatal("una dirección configurada debe conservar la capacidad aunque el motor esté caído")
	}
}
