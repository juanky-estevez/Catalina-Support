#!/usr/bin/env sh
set -eu

REPO_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
MODE=${1:-write}

case "$MODE" in
  write) TOOL_ARGS="" ;;
  check) TOOL_ARGS="--check" ;;
  *) echo "Uso: $0 [write|check]" >&2; exit 2 ;;
esac

# Instala exactamente el lockfile en el volumen aislado de desarrollo. La máquina no necesita
# Node ni Go: el generador usa los mismos entornos en contenedor que el resto del proyecto.
docker compose -f "$REPO_DIR/dev.yml" run --rm --no-deps frontend npm ci

docker run --rm \
  -v "$REPO_DIR:/repo" \
  -v catalina_support_dev_frontend_node_modules:/repo/frontend/node_modules:ro \
  -v catalina_support_dev_backend_go_mod:/go/pkg/mod \
  -w /repo/backend \
  golang:1.27-alpine \
  go run ../scripts/third-party-notices.go $TOOL_ARGS
