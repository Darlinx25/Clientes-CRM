#!/usr/bin/env bash
#
# seed_demo.sh - Vacía la base de datos y carga una demostración realista de
# Historial: 100 clientes (personas y sociedades) con entre 3 y 12 notas
# "lógicas" cada uno (apertura de empresa, bajas de padrones, CUD, expedientes,
# domicilio fiscal, WhatsApp, cambio de SRL a SA, inactivos, servicio
# doméstico, etc.), escritas por 4 usuarios distintos.
#
# Requisitos: tener el backend buildeable en el host (go), docker y la app
# levantada con docker compose desde la raíz del repo.
#
# Uso:
#   ./scripts/seed_demo.sh
#
# Usuarios demo (contraseña "demo1234" para los 4):
#   admin / admin   (admin real de la app, no cambia)
#   julia, rober, marcela, andres / demo1234
#
set -euo pipefail

cd "$(dirname "$0")/.."
ROOT="$(pwd)"

if ! command -v go >/dev/null 2>&1; then
  echo "ERROR: se necesita 'go' en el host para compilar el seeder." >&2
  exit 1
fi

if ! docker info >/dev/null 2>&1; then
  echo "ERROR: docker no está disponible." >&2
  exit 1
fi

DB_FILE="$ROOT/data/meerkat.db"
if [ ! -f "$DB_FILE" ]; then
  echo "ERROR: no existe $DB_FILE. Levantá la app una vez antes de seedear." >&2
  exit 1
fi

echo "==> 1/4 Deteniendo la app (para liberar la base de datos)..."
docker compose down >/dev/null 2>&1 || true

echo "==> 2/4 Compilando el seeder..."
BIN="$(mktemp -u /tmp/seeddemo-XXXXXX)"
trap 'rm -f "$BIN"' EXIT
( cd "$ROOT/backend" && CGO_ENABLED=0 go build -o "$BIN" ./cmd/seeddemo )

echo "==> 3/4 Vaciando y cargando datos de demostración..."
if ! docker image inspect alpine:3.20 >/dev/null 2>&1; then
  docker pull alpine:3.20 >/dev/null 2>&1 || true
fi
docker run --rm \
  -v "$ROOT/data":/data \
  -v "$BIN":/seeddemo:ro \
  -e DB_PATH=/data/meerkat.db \
  alpine:3.20 /seeddemo

# Si se corrió como root dentro del container, los archivos nuevos quedan de
# root; devolvemos la propiedad al mismo uid/gid que usa la app (PUID/PGID).
docker run --rm -v "$ROOT/data":/data alpine:3.20 chown -R "${PUID:-1001}:${PGID:-1001}" /data

echo "==> 4/4 Arrancando la app..."
docker compose up -d >/dev/null 2>&1
sleep 5

echo ""
echo "Listo. Datos de demostración cargados:"
echo "  - 100 clientes, 3-12 notas cada uno, 4 autores distintos."
echo "  - Usuarios: julia / rober / marcela / andres (clave: demo1234)"
echo "  - Admin (sin cambios): admin / admin"