#!/usr/bin/env bash
set -euo pipefail

# Construye y publica los artefactos de producción de Catalina-Support en
# /root/prod/catalina-support y, si no se le dice lo contrario, levanta los contenedores
# que los consumen (docs/ambientes.md, sección 4.2).
#
#   ./scripts/prod-build.sh                 # los dos artefactos + desplegar
#   ./scripts/prod-build.sh --no-deploy     # sólo construir y publicar
#   ./scripts/prod-build.sh --only backend  # frontend | backend
#
# Antes de construir rota el artefacto anterior a <nombre>_prev. La rotación es por COPIA
# y no por movimiento: el directorio está montado dentro de los contenedores de producción
# y renombrarlo los dejaría apuntando al inodo viejo.

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="/root/prod/catalina-support"
# **Se despliega desde la carpeta de despliegue**, no desde el repositorio: `prod.yml` lleva las rutas
# de los volúmenes **relativas a sí mismo**, así que el archivo (y lo que necesita: `config/`) tienen
# que estar al lado de los artefactos. Así el despliegue funciona desde cualquier carpeta y en
# cualquier sistema, y `./frontend` es `$OUT_DIR/frontend`, que es lo que era con las rutas absolutas.
COMPOSE=(docker compose -f "$OUT_DIR/prod.yml")

ONLY=""
DEPLOY=1
while [ $# -gt 0 ]; do
  case "$1" in
    --no-deploy) DEPLOY=0 ;;
    --only) ONLY="${2:-}"; shift ;;
    -h|--help) sed -n '3,13p' "$0"; exit 0 ;;
    *) echo "Opción desconocida: $1" >&2; exit 2 ;;
  esac
  shift
done

case "$ONLY" in
  ""|frontend|backend) ;;
  *) echo "--only acepta frontend o backend" >&2; exit 2 ;;
esac

# Los tres directorios que montan los contenedores, y el de las copias de la base que ya
# usa scripts/backup-db.sh prod (docs/ambientes.md, sección 6).
mkdir -p "$OUT_DIR/frontend" "$OUT_DIR/backend" "$OUT_DIR/_files" "$OUT_DIR/_logs"

# **El compose y lo que necesita van a la carpeta de despliegue, antes de construir**: las etapas de
# construcción escriben en `./frontend` y `./backend`, que son relativas al propio `prod.yml`, así que
# el archivo tiene que estar ya ahí para que apunten a los artefactos y no al repositorio.
cp "$REPO_DIR/prod.yml" "$OUT_DIR/prod.yml"
rm -rf "$OUT_DIR/config"
cp -a "$REPO_DIR/config" "$OUT_DIR/config"


# ---------------------------------------------------------------------------------------
# Qué se está construyendo
#
# **Un artefacto construido desde un árbol sucio no corresponde a ningún commit**, y eso
# hay que saberlo antes y no después: queda escrito en BUILD_INFO.
# ---------------------------------------------------------------------------------------
COMMIT="$(git -C "$REPO_DIR" rev-parse --short HEAD)"
BRANCH="$(git -C "$REPO_DIR" rev-parse --abbrev-ref HEAD)"
if [ -n "$(git -C "$REPO_DIR" status --porcelain)" ]; then
  DIRTY="con cambios locales: el artefacto NO corresponde a un commit limpio"
  echo "AVISO: el árbol de trabajo tiene cambios sin commitear." >&2
else
  DIRTY="limpio"
fi

rotate() {
  local target="$1"
  [ -d "$OUT_DIR/$target" ] || return 0

  # Se rota a un temporal y luego se renombra, para que un fallo a mitad no deje el
  # artefacto anterior perdido: hasta que el temporal está completo, el bueno sigue ahí.
  rm -rf "$OUT_DIR/${target}_prev.tmp"
  cp -a "$OUT_DIR/$target" "$OUT_DIR/${target}_prev.tmp"
  rm -rf "$OUT_DIR/${target}_prev"
  mv "$OUT_DIR/${target}_prev.tmp" "$OUT_DIR/${target}_prev"
  echo "rotado: $target -> ${target}_prev"
}

build() {
  local name="$1"
  rotate "$name"
  echo "== construyendo y publicando $name =="

  # **Se construye la imagen del constructor antes de usarla**, y no es un paso de más: `run` sólo
  # construye si la imagen **no existe**, así que a partir del segundo despliegue reutilizaría la de
  # la vez anterior y **publicaría el artefacto viejo sin decir nada**. Se descubrió el 2026-09-26, al
  # reemplazar los logos: el artefacto publicado seguía trayendo los de antes.
  "${COMPOSE[@]}" --profile build build "builder_$name"
  "${COMPOSE[@]}" --profile build run --rm "builder_$name"
}

# Un artefacto que no está **corta el despliegue**: desplegar sin él dejaría la aplicación
# sin ese componente, que es peor que no desplegar.
verify() {
  case "$1" in
    frontend)
      [ -f "$OUT_DIR/frontend/index.html" ] || { echo "FALTA $OUT_DIR/frontend/index.html" >&2; exit 1; } ;;
    backend)
      [ -x "$OUT_DIR/backend/catalina-support" ] || { echo "FALTA $OUT_DIR/backend/catalina-support (o no es ejecutable)" >&2; exit 1; } ;;
  esac
}

if [ -n "$ONLY" ]; then
  TARGETS=("$ONLY")
else
  TARGETS=(frontend backend)
fi

for t in "${TARGETS[@]}"; do
  build "$t"
  verify "$t"
done

# La respuesta a «¿qué hay desplegado ahora mismo?» sin adivinar.
{
  echo "commit=$COMMIT"
  echo "rama=$BRANCH"
  echo "arbol=$DIRTY"
  echo "fecha=$(date -Is)"
  echo "artefactos=${TARGETS[*]}"
} > "$OUT_DIR/BUILD_INFO"
echo "--- $OUT_DIR/BUILD_INFO ---"
cat "$OUT_DIR/BUILD_INFO"

if [ "$DEPLOY" = "1" ]; then
  echo "== levantando/actualizando los contenedores de producción =="
  "${COMPOSE[@]}" up -d

  for t in "${TARGETS[@]}"; do
    case "$t" in
      # **El frontend no se reinicia**: su contenedor es nginx y lee los estáticos del
      # disco en cada petición, así que el artefacto nuevo se sirve solo.
      backend)
        "${COMPOSE[@]}" restart backend >/dev/null
        echo "reiniciado: backend"
        ;;
    esac
  done

  "${COMPOSE[@]}" ps
fi
