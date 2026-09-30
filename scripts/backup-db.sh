#!/usr/bin/env bash
#
# Copia de seguridad de la base de datos de Catalina-Support.
#
#   ./scripts/backup-db.sh            # el entorno de desarrollo
#   ./scripts/backup-db.sh prod       # producción
#
# Qué hace, en este orden:
#
#   1. Comprueba que la base está levantada. Sin ella no hay nada que copiar, y fallar aquí es mejor
#      que dejar un archivo vacío que parece una copia.
#   2. Vuelca la base con `pg_dump -Fc` **dentro del contenedor** —el formato comprimido de PostgreSQL,
#      que se restaura con `pg_restore`— y lo deja en la carpeta de copias con la fecha en el nombre.
#   3. **Comprueba que el volcado se puede leer** (`pg_restore --list`). Una copia que no se puede
#      restaurar no es una copia, y es lo que hay que saber hoy y no el día que haga falta.
#   4. Borra las copias de más de 14 días (docs/ambientes.md, decisión 3).
#
# **No copia `_files`**, y es a propósito (decisión del responsable, 2026-09-25): los adjuntos de los
# tickets y los documentos que sube la gente se copian a mano cuando Soporte lo pida. Este script se
# ocupa de la base y de nada más, que es lo que no se puede reconstruir.
#
# Lo que cuenta va saliendo por la salida estándar, para que lo recoja el `cron`:
#
#   30 3 * * * /srv/catalina-support/source/scripts/backup-db.sh dev >> /srv/catalina-support/backups/copias.log 2>&1
#
set -euo pipefail

# Los días que se guardan las copias de la base. Lo dice el documento, y está aquí a la vista para
# que cambiarlo sea cambiar un número y volver a leer esta línea.
DIAS_DE_RETENCION=14

# Dónde se dejan: fuera del repositorio, con las copias de los otros proyectos de la máquina.
CARPETA_DE_COPIAS="${CARPETA_DE_COPIAS:-/srv/catalina-support/backups}"

ENTORNO="${1:-dev}"

case "$ENTORNO" in
    dev)
        PROYECTO="$(cd "$(dirname "$0")/.." && pwd)"
        COMPOSE="$PROYECTO/dev.yml"
        ENV="$PROYECTO/config/env/dev.env"
        ;;
    prod)
        # **El `prod.yml` y su entorno viven en el repositorio**, no con los artefactos: lo que está
        # en /srv/catalina-support es el código construido (frontend, backend, `_logs` y
        # `_files`), y el compose los monta desde ahí (docs/ambientes.md, secciones 4.1 y 4.2).
        PROYECTO="$(cd "$(dirname "$0")/.." && pwd)"
        COMPOSE="$PROYECTO/prod.yml"
        ENV="$PROYECTO/config/env/prod.env"
        ;;
    *)
        echo "Entorno desconocido: $ENTORNO. Se usa «dev» o «prod»." >&2
        exit 2
        ;;
esac

# Se leen del archivo de entorno **sin ejecutarlo**, que es un archivo de configuración y no un
# programa: aquí no se le da de comer a `source`.
valor() {
    sed -n "s/^$1=//p" "$ENV" | head -1
}

USUARIO="$(valor POSTGRES_USER)"
BASE="$(valor POSTGRES_DB)"
PUERTO="$(valor PGPORT)"

if [ -z "$USUARIO" ] || [ -z "$BASE" ] || [ -z "$PUERTO" ]; then
    echo "Faltan POSTGRES_USER, POSTGRES_DB o PGPORT en $ENV: no se puede copiar nada." >&2
    exit 1
fi

if [ ! -f "$COMPOSE" ]; then
    echo "No está el compose del entorno $ENTORNO: $COMPOSE" >&2
    exit 1
fi

ahora="$(date +%Y%m%d_%H%M%S)"
destino="$CARPETA_DE_COPIAS/${BASE}_${ahora}.dump"

mkdir -p "$CARPETA_DE_COPIAS"

# El contenedor tiene que estar levantado: `docker compose exec` no arranca nada.
if ! docker compose -f "$COMPOSE" ps --status running --format '{{.Service}}' | grep -qx database; then
    echo "$(date '+%F %T')  La base del entorno $ENTORNO no está levantada: no se copia nada." >&2
    exit 1
fi

echo "$(date '+%F %T')  Copiando la base «$BASE» del entorno $ENTORNO…"

# El volcado sale por la salida estándar del contenedor y se escribe aquí: así no hay que copiar
# archivos de dentro a fuera, y `-T` quita el terminal para que la salida sea el volcado y nada más.
if ! docker compose -f "$COMPOSE" exec -T database \
    pg_dump -U "$USUARIO" -d "$BASE" -p "$PUERTO" -Fc >"$destino"; then
    rm -f "$destino"
    echo "$(date '+%F %T')  La copia ha fallado: no se deja ningún archivo a medias." >&2
    exit 1
fi

# Un volcado de cero bytes es un fallo que no se ve hasta que hace falta.
if [ ! -s "$destino" ]; then
    rm -f "$destino"
    echo "$(date '+%F %T')  La copia ha salido vacía: se borra." >&2
    exit 1
fi

# Y que se pueda leer, que es lo que separa una copia de un archivo.
if ! docker compose -f "$COMPOSE" exec -T database pg_restore --list <"$destino" >/dev/null; then
    echo "$(date '+%F %T')  La copia no se puede leer con pg_restore: se queda para poder mirarla, pero esto hay que arreglarlo." >&2
    exit 1
fi

echo "$(date '+%F %T')  Copia hecha: $destino ($(du -h "$destino" | cut -f1))"

# Y se limpia lo viejo. Sólo las copias de la base: los adjuntos no los toca nadie.
borradas="$(find "$CARPETA_DE_COPIAS" -maxdepth 1 -type f -name "${BASE}_*.dump" -mtime "+$DIAS_DE_RETENCION" -print -delete | wc -l)"
if [ "$borradas" -gt 0 ]; then
    echo "$(date '+%F %T')  Se han borrado $borradas copias de más de $DIAS_DE_RETENCION días."
fi

echo "$(date '+%F %T')  Quedan $(find "$CARPETA_DE_COPIAS" -maxdepth 1 -type f -name "${BASE}_*.dump" | wc -l) copias de «$BASE»."
