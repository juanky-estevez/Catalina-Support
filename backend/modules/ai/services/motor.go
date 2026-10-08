package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// motor es el cliente del motor de inteligencia artificial: se le habla **por HTTP y en el formato de
// OpenAI** (`POST {URL}/v1/chat/completions`), que es el que sirve llama.cpp en su contenedor
// (docs/modules/ai.md, sección 2).
type motor struct {
	cliente *http.Client
}

// nuevoMotor construye el cliente. El tiempo límite es del código: un motor que no responde no puede
// dejar colgada una tarea de fondo.
func nuevoMotor(espera time.Duration) *motor {
	return &motor{cliente: &http.Client{
		Timeout: espera,
		// Una URL de proveedor debe ser la definitiva. Seguir un 3xx podría reenviar Bearer, Basic o
		// una cabecera privada a un destino que el Administrador nunca aprobó.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// peticion es lo que se le manda al motor, en el formato de OpenAI.
type peticion struct {
	Model    string    `json:"model,omitempty"`
	Messages []mensaje `json:"messages"`
	// Temperatura baja: se busca un resumen, no literatura (docs/modules/ai.md, sección 3).
	Temperature float64 `json:"temperature"`
	// Tope de piezas de la respuesta, sacado del tope de palabras.
	MaxTokens int `json:"max_tokens"`
	// Se le pide el JSON como tal, que es lo que hace que un modelo pequeño se salga menos del formato.
	ResponseFormat formatoDeRespuesta `json:"response_format"`
}

// mensaje es un turno de la conversación.
type mensaje struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// formatoDeRespuesta es el `response_format` de OpenAI.
type formatoDeRespuesta struct {
	Type string `json:"type"`
}

// respuesta es lo que devuelve el motor.
type respuesta struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// probarConfiguracion ejecuta una generación artificial con el mismo protocolo y autenticación que
// se usarán después. No contiene datos de tickets.
func (m *motor) probarConfiguracion(ctx context.Context, baseURL, model, provider, authType, authHeader, credential, language string) error {
	prompt := "Reply only with JSON: {\"text\":\"connection ok\"}."
	endpoint := direccionDe(baseURL, "/v1/chat/completions")
	var body any = map[string]any{"model": model, "messages": []map[string]string{{"role": "user", "content": prompt}}, "temperature": 0, "max_tokens": 64, "response_format": map[string]string{"type": "json_object"}}
	if provider == "claude" {
		endpoint = direccionDe(baseURL, "/v1/messages")
		body = map[string]any{"model": model, "messages": []map[string]string{{"role": "user", "content": prompt}}, "temperature": 0, "max_tokens": 64}
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return ErrRespuestaInvalida
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return ErrSinMotor
	}
	req.Header.Set("Content-Type", "application/json")
	switch authType {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+credential)
	case "header":
		req.Header.Set(authHeader, credential)
	case "basic":
		parts := strings.SplitN(credential, ":", 2)
		if len(parts) != 2 {
			return ErrSinMotor
		}
		req.SetBasicAuth(parts[0], parts[1])
	}
	if provider == "claude" {
		req.Header.Set("x-api-key", credential)
		req.Header.Set("anthropic-version", "2023-06-01")
	}
	resp, err := m.cliente.Do(req)
	if err != nil {
		return ErrSinMotor
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		vaciar(resp.Body)
		return ErrSinMotor
	}
	var content string
	if provider == "claude" {
		var out struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		data, readErr := leerRespuestaLimitada(resp.Body)
		if readErr != nil || json.Unmarshal(data, &out) != nil || len(out.Content) == 0 {
			return ErrRespuestaInvalida
		}
		content = out.Content[0].Text
	} else {
		var out respuesta
		data, readErr := leerRespuestaLimitada(resp.Body)
		if readErr != nil || json.Unmarshal(data, &out) != nil || len(out.Choices) == 0 {
			return ErrRespuestaInvalida
		}
		content = out.Choices[0].Message.Content
	}
	obj, ok := primerObjetoJSON(content)
	if !ok {
		return ErrRespuestaInvalida
	}
	var parsed struct {
		Text string `json:"text"`
	}
	if json.Unmarshal([]byte(obj), &parsed) != nil || strings.TrimSpace(parsed.Text) == "" {
		return ErrRespuestaInvalida
	}
	return nil
}

// pedir hace **una** llamada al motor y devuelve las dos redacciones.
//
// Los dos fallos salen distintos porque el módulo los trata distinto: `ErrSinMotor` cuando no
// contesta —que se reintenta— y `ErrRespuestaInvalida` cuando contesta algo que no vale —que se
// reintenta una sola vez— (docs/modules/ai.md, decisión 8).
func (m *motor) pedir(ctx context.Context, ajustes Ajustes, encargo encargo) (string, string, error) {
	if ajustes.Provider != "" {
		return m.pedirGlobal(ctx, ajustes, encargo)
	}
	cuerpo, err := json.Marshal(peticion{
		Model: ajustes.Modelo,
		Messages: []mensaje{
			{Role: "system", Content: encargo.sistema},
			{Role: "user", Content: encargo.usuario},
		},
		Temperature:    temperatura,
		MaxTokens:      topeDePiezas(ajustes),
		ResponseFormat: formatoDeRespuesta{Type: "json_object"},
	})
	if err != nil {
		return "", "", &respuestaInvalida{forma: "no se pudo armar la petición"}
	}

	peticionHTTP, err := http.NewRequestWithContext(
		ctx, http.MethodPost, direccionDe(ajustes.URL, "/v1/chat/completions"), bytes.NewReader(cuerpo),
	)
	if err != nil {
		return "", "", ErrSinMotor
	}
	peticionHTTP.Header.Set("Content-Type", "application/json")

	respuestaHTTP, err := m.cliente.Do(peticionHTTP)
	if err != nil {
		// **Aquí no se registra nada**: el error puede llevar la dirección, y del texto del ticket no
		// se registra ni una letra (docs/modules/ai.md, sección 3).
		return "", "", ErrSinMotor
	}
	defer respuestaHTTP.Body.Close()

	if respuestaHTTP.StatusCode < 200 || respuestaHTTP.StatusCode > 299 {
		// El motor está, pero no atiende. Para el módulo es lo mismo que no contestar: se reintenta y,
		// si sigue, el campo queda en `sin_motor`. El cuerpo se drena —sin leerlo— para que la conexión
		// se pueda reutilizar.
		vaciar(respuestaHTTP.Body)
		return "", "", ErrSinMotor
	}

	var salida respuesta
	if err := json.NewDecoder(io.LimitReader(respuestaHTTP.Body, maxRespuestaBytes)).Decode(&salida); err != nil {
		return "", "", &respuestaInvalida{forma: "el motor devolvió algo que no es JSON"}
	}
	if len(salida.Choices) == 0 {
		return "", "", &respuestaInvalida{forma: "el motor no devolvió ninguna respuesta"}
	}

	es, en, hay := dosIdiomas(salida.Choices[0].Message.Content)
	if !hay {
		// Se devuelve **con la forma**, para que el aviso del log diga qué faltó.
		return "", "", formaDeRespuesta(salida.Choices[0].Message.Content)
	}

	return recortarPalabras(es, ajustes.Palabras), recortarPalabras(en, ajustes.Palabras), nil
}

// pedirGlobal genera una sola redacción en el idioma de la instalación y habla tanto con APIs
// compatibles con OpenAI como con Claude. Se devuelve en los dos huecos transitorios para conservar
// el contrato interno mientras la migración 1.0.0 retira las columnas bilingües.
func (m *motor) pedirGlobal(ctx context.Context, ajustes Ajustes, encargo encargo) (string, string, error) {
	idioma := "English"
	if ajustes.Language == "es" {
		idioma = "Spanish"
	}
	prompt := encargo.texto + "\n\nWrite only in " + idioma +
		`. Return exactly {"text":"..."} and no other keys or surrounding text.`
	endpoint := direccionDe(ajustes.URL, "/v1/chat/completions")
	body := map[string]any{
		"model": ajustes.Modelo, "messages": []map[string]string{
			{"role": "system", "content": encargo.sistema}, {"role": "user", "content": prompt},
		}, "temperature": temperatura, "max_tokens": topeDePiezas(ajustes),
		"response_format": map[string]string{"type": "json_object"},
	}
	if ajustes.Provider == "claude" {
		endpoint = direccionDe(ajustes.URL, "/v1/messages")
		body = map[string]any{"model": ajustes.Modelo, "system": encargo.sistema,
			"messages":    []map[string]string{{"role": "user", "content": prompt}},
			"temperature": temperatura, "max_tokens": topeDePiezas(ajustes)}
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return "", "", ErrRespuestaInvalida
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return "", "", ErrSinMotor
	}
	req.Header.Set("Content-Type", "application/json")
	aplicarCredencial(req, ajustes.Provider, ajustes.AuthType, ajustes.AuthHeader, ajustes.Credential)
	resp, err := m.cliente.Do(req)
	if err != nil {
		return "", "", ErrSinMotor
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		vaciar(resp.Body)
		return "", "", ErrSinMotor
	}
	content, err := contenidoDeRespuesta(resp.Body, ajustes.Provider)
	if err != nil {
		return "", "", err
	}
	obj, ok := primerObjetoJSON(content)
	if !ok {
		return "", "", ErrRespuestaInvalida
	}
	var parsed struct {
		Text string `json:"text"`
	}
	if json.Unmarshal([]byte(obj), &parsed) != nil || strings.TrimSpace(parsed.Text) == "" {
		return "", "", ErrRespuestaInvalida
	}
	text := recortarPalabras(strings.TrimSpace(parsed.Text), ajustes.Palabras)
	return text, text, nil
}

func aplicarCredencial(req *http.Request, provider, authType, authHeader, credential string) {
	switch authType {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+credential)
	case "header":
		req.Header.Set(authHeader, credential)
	case "basic":
		parts := strings.SplitN(credential, ":", 2)
		if len(parts) == 2 {
			req.SetBasicAuth(parts[0], parts[1])
		}
	}
	if provider == "claude" {
		req.Header.Set("x-api-key", credential)
		req.Header.Set("anthropic-version", "2023-06-01")
	}
}

func contenidoDeRespuesta(body io.Reader, provider string) (string, error) {
	data, err := leerRespuestaLimitada(body)
	if err != nil {
		return "", ErrRespuestaInvalida
	}
	if provider == "claude" {
		var out struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		if json.Unmarshal(data, &out) != nil || len(out.Content) == 0 {
			return "", ErrRespuestaInvalida
		}
		return out.Content[0].Text, nil
	}
	var out respuesta
	if json.Unmarshal(data, &out) != nil || len(out.Choices) == 0 {
		return "", ErrRespuestaInvalida
	}
	return out.Choices[0].Message.Content, nil
}

func leerRespuestaLimitada(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, maxRespuestaBytes+1))
	if err != nil || len(data) > maxRespuestaBytes {
		return nil, ErrRespuestaInvalida
	}
	return data, nil
}

// disponible pregunta si el motor responde.
//
// Se pregunta a su chequeo de salud, que es lo que contesta aunque no haya nada que resumir, y con un
// tiempo corto: «¿está el motor?» lo pregunta una pantalla, y no puede dejarla esperando.
func (m *motor) disponible(ctx context.Context, ajustes Ajustes) bool {
	ctx, cancelar := context.WithTimeout(ctx, esperaDelChequeo)
	defer cancelar()

	peticionHTTP, err := http.NewRequestWithContext(ctx, http.MethodGet, direccionDe(ajustes.URL, "/health"), nil)
	if err != nil {
		return false
	}

	respuestaHTTP, err := m.cliente.Do(peticionHTTP)
	if err != nil {
		return false
	}
	defer respuestaHTTP.Body.Close()

	vaciar(respuestaHTTP.Body)

	return respuestaHTTP.StatusCode >= 200 && respuestaHTTP.StatusCode <= 299
}

// probar pregunta por la salud del motor de esa dirección y **devuelve el error** en vez de un
// booleano: es lo que necesita el botón de probar la conexión, que tiene que decir qué ha pasado.
//
// El tiempo límite lo pone quien llama (un contexto corto): probar una conexión no puede dejar la
// pantalla esperando a un contenedor que no está.
func (m *motor) probar(ctx context.Context, url string) error {
	peticionHTTP, err := http.NewRequestWithContext(ctx, http.MethodGet, direccionDe(url, "/health"), nil)
	if err != nil {
		return ErrSinMotor
	}

	respuestaHTTP, err := m.cliente.Do(peticionHTTP)
	if err != nil {
		return ErrSinMotor
	}
	defer respuestaHTTP.Body.Close()

	vaciar(respuestaHTTP.Body)

	if respuestaHTTP.StatusCode < 200 || respuestaHTTP.StatusCode > 299 {
		return ErrSinMotor
	}

	return nil
}

// dosIdiomas saca las dos redacciones de lo que haya contestado el modelo, **con tolerancia**: un
// modelo pequeño contesta con el JSON pedido, o con el JSON envuelto en ```json … ```, o con una
// frase delante. Nada de eso es motivo para tirar una respuesta que sirve.
func dosIdiomas(contenido string) (string, string, bool) {
	objeto, hay := primerObjetoJSON(contenido)
	if !hay {
		return "", "", false
	}

	var resumen struct {
		Es string `json:"es"`
		En string `json:"en"`
	}
	if err := json.Unmarshal([]byte(objeto), &resumen); err != nil {
		return "", "", false
	}

	es := strings.TrimSpace(resumen.Es)
	en := strings.TrimSpace(resumen.En)

	// Las dos tienen que venir: media respuesta no es una respuesta, y **el inglés no se traduce del
	// español** a propósito (docs/modules/ai.md, sección 7).
	if es == "" || en == "" {
		return "", "", false
	}

	return es, en, true
}

// respuestaInvalida es el error de una respuesta que no vale, **con su forma y sin su contenido**.
//
// La forma es lo que se registra: qué claves llegaron y cuánto midieron. El texto que escribió el
// modelo no se copia en ningún log, por la misma razón por la que no se registra el texto del ticket.
type respuestaInvalida struct {
	forma string
}

// Error cumple lo que pide `error`, y devuelve la clave del módulo: es lo que viaja a la pantalla.
func (e *respuestaInvalida) Error() string { return ClaveInvalida }

// Is hace que `errors.Is(err, ErrRespuestaInvalida)` siga siendo verdad aunque el error lleve la forma
// dentro.
//
// **No es un detalle de estilo**: el módulo decide con esa comparación si el fallo es «contestó mal»
// —que se reintenta al momento y acaba en `error`— o «no contesta» —que espera y acaba en
// `sin_motor`—. Sin esto, una respuesta inválida se trataría como un motor caído.
func (e *respuestaInvalida) Is(destino error) bool { return destino == ErrRespuestaInvalida }

// formaDeRespuesta resume lo que llegó, sin copiarlo.
func formaDeRespuesta(contenido string) *respuestaInvalida {
	objeto, hay := primerObjetoJSON(contenido)
	if !hay {
		return &respuestaInvalida{forma: "sin ningún objeto JSON (" + strconv.Itoa(len(contenido)) + " caracteres)"}
	}

	var claves map[string]any
	if err := json.Unmarshal([]byte(objeto), &claves); err != nil {
		return &respuestaInvalida{forma: "un JSON que no se puede leer: " + err.Error()}
	}

	var partes []string
	for _, idioma := range []string{"es", "en"} {
		valor, hay := claves[idioma]
		if !hay {
			partes = append(partes, idioma+"=falta")
			continue
		}

		texto, esTexto := valor.(string)
		if !esTexto {
			partes = append(partes, idioma+"=no es texto")
			continue
		}

		partes = append(partes, idioma+"="+strconv.Itoa(len(strings.TrimSpace(texto)))+" caracteres")
	}

	return &respuestaInvalida{forma: "claves: " + strings.Join(partes, ", ")}
}

// primerObjetoJSON devuelve el primer objeto JSON que haya dentro del texto, contando llaves y
// respetando las que vayan dentro de una cadena.
//
// Es lo que hace que sirva tanto `{"es": "…", "en": "…"}` como el mismo objeto dentro de un bloque de
// código, y que una llave suelta en una frase no despiste al recuento.
func primerObjetoJSON(texto string) (string, bool) {
	inicio := strings.IndexByte(texto, '{')
	if inicio < 0 {
		return "", false
	}

	profundidad := 0
	dentroDeCadena := false
	escapado := false

	for i := inicio; i < len(texto); i++ {
		caracter := texto[i]

		if dentroDeCadena {
			switch {
			case escapado:
				escapado = false
			case caracter == '\\':
				escapado = true
			case caracter == '"':
				dentroDeCadena = false
			}

			continue
		}

		switch caracter {
		case '"':
			dentroDeCadena = true
		case '{':
			profundidad++
		case '}':
			profundidad--
			if profundidad == 0 {
				return texto[inicio : i+1], true
			}
		}
	}

	return "", false
}

// direccionDe junta la base del motor con su camino, sin barras de más.
func direccionDe(base, camino string) string {
	return strings.TrimRight(strings.TrimSpace(base), "/") + camino
}

// vaciar lee y tira lo que quede del cuerpo, que es lo que permite reutilizar la conexión.
func vaciar(cuerpo io.Reader) {
	_, _ = io.Copy(io.Discard, io.LimitReader(cuerpo, maxRespuestaBytes))
}

// topeDePiezas es el tope de la respuesta, en piezas del modelo.
//
// Se saca del tope de palabras con **margen de sobra**, y medido contra el motor de verdad: una
// palabra castellana se lleva una o dos piezas, la respuesta lleva **dos redacciones** (el español y
// su inglés), y encima van las claves del JSON. Con 60 palabras, 400 piezas: el margen no es para que
// escriba más —el tope de palabras es el que manda y se recorta— sino para que **el JSON no se corte
// a medias**, que es el único final que no tiene arreglo.
//
// Y **nunca pasa de media ventana**, para que la petición y su respuesta quepan juntas en el contexto
// del modelo: un ticket de 6 000 caracteres son unas 2 000 piezas de entrada.
func topeDePiezas(ajustes Ajustes) int {
	tope := ajustes.Palabras*piezasPorPalabra + margenDeLaRespuesta

	if media := ajustes.Contexto / 2; tope > media {
		tope = media
	}
	if tope < minPiezas {
		tope = minPiezas
	}

	return tope
}

const (
	// Temperatura de la llamada: baja, porque lo que se quiere es un resumen fiel y no una versión
	// libre del ticket.
	temperatura = 0.2
	// Las piezas que se le dejan por palabra pedida, y el margen fijo del envoltorio del JSON.
	piezasPorPalabra    = 4
	margenDeLaRespuesta = 160
	// Lo que se acepta leer de una respuesta. Es un tope de seguridad: un motor desbocado no puede
	// llenar la memoria del backend.
	maxRespuestaBytes = 1 << 20
	// El mínimo de piezas que se le deja contestar, aunque la configuración diga menos.
	minPiezas = 64
)
