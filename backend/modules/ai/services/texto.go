package services

import "strings"

// MaxTexto es lo que se le manda al motor como mucho, en caracteres (docs/modules/ai.md, decisión 6).
//
// El tope es del módulo `ai` y no de `tickets` a propósito: quien manda el texto manda lo que hay, y
// quien conoce el motor sabe cuánto le cabe.
const MaxTexto = 6000

// recortar deja el texto en el tope **quitando por el principio**.
//
// Lo último que pasó es justo lo que hay que leer: si un ticket tiene doscientos comentarios, el
// resumen se escribe con los de ahora y no con el primero (docs/modules/ai.md, decisión 6).
//
// El corte se mide en caracteres y no en bytes —el texto español lleva acentos, y cada uno ocupa dos
// bytes: cortar por bytes partiría una letra por la mitad— y, cuando hay un salto de línea cerca del
// corte, se empieza en la línea siguiente para no arrancar a media palabra.
func recortar(texto string) string {
	texto = strings.TrimSpace(texto)
	if len(texto) <= MaxTexto {
		return texto
	}

	letras := []rune(texto)
	corte := string(letras[len(letras)-MaxTexto:])

	if salto := strings.IndexByte(corte, '\n'); salto >= 0 && salto <= maxAjusteDeLinea {
		corte = corte[salto+1:]
	}

	return corte
}

// recortarPalabras corta una redacción al tope de palabras.
//
// El tope se le pide al motor en el encargo, pero **no se fía**: un modelo pequeño se pasa, y un
// campo de la lista que ocupe media pantalla deja de ser un resumen. Lo que se corta se dice con unos
// puntos suspensivos, para que no parezca que el texto está entero.
func recortarPalabras(texto string, palabras int) string {
	texto = strings.TrimSpace(texto)
	if palabras <= 0 {
		return texto
	}

	partes := strings.Fields(texto)
	if len(partes) <= palabras {
		return texto
	}

	return strings.Join(partes[:palabras], " ") + "…"
}

// maxAjusteDeLinea es hasta dónde se busca un salto de línea para empezar el recorte en una línea
// entera. Más allá, se corta donde toque: quitar cien caracteres más por no empezar a media palabra
// sería perder texto que sí hace falta.
const maxAjusteDeLinea = 200
