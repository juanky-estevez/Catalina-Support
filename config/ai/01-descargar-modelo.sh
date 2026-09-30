#!/bin/sh
# ============================================================================
# Catalina-Support · motor de IA local
# ----------------------------------------------------------------------------
# 01-descargar-modelo.sh — lo primero que se ejecuta en el contenedor `ai`.
#
# Qué hace, en orden, y por qué:
#
#   1. Comprueba si el modelo ya está en el volumen (`/modelos`). El modelo **no
#      está en el repositorio**: pesa ~1,07 GB y no tiene nada que hacer en git.
#      Vive en un volumen con nombre (`ai_modelos`), así que sobrevive a `down`,
#      a `up` y a recrear el contenedor: sólo se descarga la primera vez (o si
#      alguien borra el volumen a propósito). A partir de ahí el motor arranca
#      **sin salida a internet**, que es lo que quiere `docs/modules/ai.md`
#      (sección 2): el texto de los tickets no sale del servidor.
#
#   2. Si falta, lo descarga **a un archivo temporal** y sólo al terminar lo
#      renombra al nombre definitivo. Así una descarga cortada no deja un `.gguf`
#      a medias que el servidor intentaría cargar al siguiente arranque (y
#      fallaría con un error de formato difícil de leer). La comprobación de la
#      cabecera `GGUF` remata la idea: hasta que el archivo no empieza por `GGUF`,
#      no se da por bueno ni se renombra.
#
#   3. Termina con `exec`, sustituyendo el shell por el servidor: el **proceso 1
#      del contenedor es llama-server**, no un shell esperando detrás. Eso es lo
#      que hace que `docker stop` y `restart: always` se comporten como se espera
#      —las señales llegan al servidor, no a un `sh` intermedio—.
#
# Los argumentos del servidor llegan por `command:` en `ai.yml` y se pasan tal
# cual con `"$@"`. Aquí no se decide nada del servidor: eso se lee en `ai.yml`.
#
# `curl` viene en la propia imagen (es lo que usa su comprobación de salud), así
# que no hay que instalar nada dentro del contenedor.
# ============================================================================

set -eu

# El nombre del archivo es el que trae Hugging Face, sin cambiarlo: así se puede
# comprobar a mano contra el repositorio del modelo sin dudar de qué hay ahí.
MODELO="/modelos/qwen2.5-1.5b-instruct-q4_k_m.gguf"
URL="https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf"
PARCIAL="${MODELO}.parcial"

# Un GGUF de verdad empieza por la magia `GGUF`. Con esto se distingue un modelo
# completo de un archivo a medias sin depender de tamaños ni de sumas de control.
es_modelo() {
  [ -f "$1" ] && [ "$(head -c 4 "$1" 2>/dev/null)" = "GGUF" ]
}

if es_modelo "$MODELO"; then
  echo "[ai] El modelo ya está en el volumen: ${MODELO} ($(du -h "$MODELO" | cut -f1)). No se descarga nada."
else
  echo "[ai] El modelo no está en el volumen. Se descarga una sola vez (~1,1 GB) y ya no se vuelve a pedir."
  echo "[ai] Origen: ${URL}"
  # Un archivo a medias de un intento anterior no sirve para nada: fuera.
  rm -f "$MODELO" "$PARCIAL"
  # `--retry` para que un corte de red no obligue a empezar de cero a mano.
  curl -fL --retry 3 --retry-delay 5 --connect-timeout 20 -o "$PARCIAL" "$URL"
  if ! es_modelo "$PARCIAL"; then
    echo "[ai] ERROR: la descarga no ha dejado un GGUF válido en ${PARCIAL}." >&2
    exit 1
  fi
  # Mismo sistema de archivos (el volumen), así que el renombrado es atómico: o
  # está el archivo bueno con su nombre, o no está.
  mv "$PARCIAL" "$MODELO"
  echo "[ai] Modelo descargado: ${MODELO} ($(du -h "$MODELO" | cut -f1))."
fi

echo "[ai] Arrancando el servidor: /app/llama-server $*"
exec /app/llama-server "$@"
