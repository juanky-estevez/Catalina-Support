#!/usr/bin/env bash
#
# Deja el entorno de desarrollo con los datos de ejemplo de Catalina-Support.
#
#   ./scripts/dev-seed.sh
#
# Qué hace:
#
#   1. Aplica `backend/migrations/v1.0.0.sql`, por si la base está recién creada o le falta algo.
#   2. Aplica `backend/migrations/v1.0.0_dev.sql`: **once cuentas y 25 tickets** con su historia,
#      sus escalados, sus comentarios y sus adjuntos.
#   3. Copia los archivos de los adjuntos (`config/seed/`) a `_files/`, en la carpeta de su ticket.
#
# **Es de desarrollo y borra**: los tickets que hubiera —también los que hayas creado probando—, las
# cuentas `@demo.com` y los adjuntos que no sean la marca. Deja el entorno exactamente como dice el
# seeder, que es lo que hace que se pueda volver a él después de trastear o de correr las pruebas.
#
# **Las once cuentas entran con `123123123`** (decisión del responsable, 2026-09-25): en desarrollo se
# entra y se sale muchas veces al día, y una contraseña que se escribe con una mano es lo que hace que
# probar no dé pereza. **Cumple la política del producto** (ocho caracteres como mínimo), así que el
# seeder no se salta ninguna regla. Las pruebas de interfaz **no usan estas cuentas**: crean las suyas
# con la contraseña que fija `tests/e2e/ayudas.ts`.
#
set -euo pipefail

PROYECTO="$(cd "$(dirname "$0")/.." && pwd)"
COMPOSE="$PROYECTO/dev.yml"
ENV="$PROYECTO/config/env/dev.env"

valor() {
    sed -n "s/^$1=//p" "$ENV" | head -1
}

USUARIO="$(valor POSTGRES_USER)"
BASE="$(valor POSTGRES_DB)"
PUERTO="$(valor PGPORT)"

if [ -z "$USUARIO" ] || [ -z "$BASE" ] || [ -z "$PUERTO" ]; then
    echo "Faltan POSTGRES_USER, POSTGRES_DB o PGPORT en $ENV." >&2
    exit 1
fi

if ! docker compose -f "$COMPOSE" ps --status running --format '{{.Service}}' | grep -qx database; then
    echo "La base no está levantada: levántala con «docker compose -f dev.yml up -d»." >&2
    exit 1
fi

# Los avisos de «la tabla ya existe» son lo normal al volver a aplicar el esquema, así que se callan;
# **los errores no**: `ON_ERROR_STOP` corta en el primero y el guion se para con él.
psql() {
    docker compose -f "$COMPOSE" exec -T -e PGOPTIONS='-c client_min_messages=warning' database \
        psql -v ON_ERROR_STOP=1 -U "$USUARIO" -d "$BASE" -p "$PUERTO" "$@"
}

echo "1/3  Aplicando el esquema…"
psql -q -o /dev/null <"$PROYECTO/backend/migrations/v1.0.0.sql"

echo "2/3  Aplicando los datos de ejemplo…"
psql -q -o /dev/null <"$PROYECTO/backend/migrations/v1.0.0_dev.sql"

echo "3/3  Copiando los adjuntos…"

# Los adjuntos de desarrollo que quedaran de antes no sirven para nada: el seeder ha borrado sus
# tickets, así que sus archivos serían huérfanos. **La marca no se toca** (el logo de la instalación
# no es un adjunto) y `.gitkeep` tampoco.
find "$PROYECTO/_files" -mindepth 1 -maxdepth 1 ! -name '.gitkeep' ! -name 'brand' -exec rm -rf {} +

# Y se copian los del ejemplo. Se pregunta a la base **cuáles y dónde**, que es lo que garantiza que
# el archivo y su fila digan lo mismo: el nombre original es el del archivo del repositorio y la ruta
# es la que espera la fila.
copiados=0
while IFS='|' read -r original guardado; do
    origen="$PROYECTO/config/seed/$original"
    destino="$PROYECTO/_files/$guardado"
    if [ ! -f "$origen" ]; then
        echo "Falta $origen para el adjunto «$original» del seeder." >&2
        exit 1
    fi
    mkdir -p "$(dirname "$destino")"
    cp -f "$origen" "$destino"
    copiados=$((copiados + 1))
done < <(psql -At -F'|' -c "SELECT filename, stored_name FROM ticket_attachments ORDER BY id")

echo
# Se cuentan **las cuentas de ejemplo**, no todas: el entorno puede tener además las de las pruebas de
# interfaz (`e2e-…`), que no son parte del juego de datos.
echo "Listo: $(psql -At -c "SELECT count(*) FROM users WHERE lower(email) LIKE '%@demo.com' AND lower(email) NOT LIKE 'e2e-%'") cuentas, $(psql -At -c 'SELECT count(*) FROM tickets') tickets y $copiados adjuntos."
# **No se repiten las cuentas aquí**: la tabla con el nombre, el correo, el rol y la contraseña de las
# once está en el `README.md`, en «Probarlo en local (desarrollo)».
echo "Listo. Las once cuentas de ejemplo —con su rol y su contraseña— están en la tabla del README."
