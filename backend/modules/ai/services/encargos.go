package services

import (
	"strconv"
)

// encargo es lo que se le pide al motor: el oficio que hace y lo que se le manda de verdad.
type encargo struct {
	sistema string
	usuario string
}

// encargoDe arma el encargo de uno de los dos campos.
//
// **Son dos encargos distintos, y por eso son dos peticiones** (docs/modules/ai.md, decisión 7): «de
// qué va el ticket» y «qué fue lo último que pasó» no se resumen igual, y un solo texto para las dos
// cosas saldría peor en las dos.
//
// El reparto es este, y está medido contra el motor de verdad (llama.cpp con Qwen2.5-1.5B-Instruct,
// 2026-09-27):
//
//   - el **mensaje de sistema** dice el oficio: de qué va el ticket, o qué fue lo último;
//   - el **mensaje de usuario** lleva el texto del ticket y, **al final**, la exigencia del formato:
//     el objeto JSON exacto con sus dos claves y nada más, con la plantilla delante de los ojos del
//     modelo. Va al final a propósito, porque es lo último que lee antes de escribir y es lo que más
//     se le pega a un modelo pequeño;
//   - y en el cuerpo de la petición va `response_format: {"type": "json_object"}`, que es la otra
//     mitad de lo que hace que el JSON salga limpio.
//
// El texto del ticket va recortado por el principio (decisión 6).
func encargoDe(tipo Tipo, texto string, palabras int) encargo {
	return encargo{
		sistema: oficioDe(tipo),
		usuario: recortar(texto) + "\n\n" + exigenciaDelFormato(palabras),
	}
}

// oficioDe es lo que se le pide a cada campo, que es lo único que cambia entre los dos encargos.
func oficioDe(tipo Tipo) string {
	switch tipo {
	case Motivo:
		return "Eres quien resume tickets en una mesa de ayuda. Escribe DE QUÉ VA EL TICKET en una sola " +
			"frase, en presente y sin rodeos, como se lo contarías a un compañero que va a atenderlo: " +
			"qué le pasa a quien lo abrió y qué necesita. Fíjate en el asunto, en la descripción y en " +
			"lo que pide el usuario, no en lo que se ha hecho después."
	case UltimaAccion:
		return "Eres quien resume tickets en una mesa de ayuda. Escribe QUÉ FUE LO ÚLTIMO QUE PASÓ en " +
			"este ticket, en una sola frase, en pasado y sin rodeos. Lo último está al final de la " +
			"conversación: fíjate en el último comentario y en el último cambio de estado, y di si el " +
			"ticket está esperando a alguien (al usuario, a soporte o a desarrollo)."
	}

	return ""
}

// exigenciaDelFormato es el contrato del módulo, y va en el mensaje del usuario: **un objeto JSON con
// las dos redacciones y ningún campo más** (docs/modules/ai.md, decisión 3), con el inglés escrito a
// la vez que el español y no traducido después (sección 7).
//
// La plantilla se le enseña tal cual —`{"es": "…", "en": "…"}`— porque es lo que hace que el modelo
// conteste ese objeto y no el suyo: se midió que sin ella se inventa las claves, y con ella sale bien
// incluso con un ticket de 6 000 caracteres.
func exigenciaDelFormato(palabras int) string {
	return "Devuelve EXACTAMENTE este objeto JSON, con esas dos claves y ningún campo más: el español " +
		"en \"es\" y el inglés en \"en\". El inglés no es una traducción del español, es el mismo " +
		"resumen escrito en inglés.\n" +
		"No inventes nada que no esté en el ticket: si un dato no aparece, no lo pongas.\n" +
		"No pases de " + strconv.Itoa(palabras) + " palabras en cada idioma.\n" +
		"Devuelve sólo el objeto, sin texto alrededor y sin bloques de código:\n" +
		`{"es": "…", "en": "…"}`
}
