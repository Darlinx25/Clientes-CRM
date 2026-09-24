#!/usr/bin/env bash
# Builds a single self-contained Windows .exe (and a Linux twin for local
# testing) that hosts the React UI + Go API + SQLite DB, ready to be
# double-clicked on a Windows machine with nothing installed.
#
# Output:
#   dist/clientes-crm.exe      (Windows 64-bit, the file to copy/share)
#   dist/clientes-crm-linux    (Linux binary to verify locally)
#
# Usage:
#   ./scripts/build-windows-exe.sh
set -euo pipefail

cd "$(dirname "$0")/.."
ROOT="$(pwd)"
FRONTEND="$ROOT/frontend"
BACKEND="$ROOT/backend"
WEBDIR="$BACKEND/web/dist"
OUT="$ROOT/dist"

# 1) Build the React app with a relative API base so it talks to the same
#    origin (the embedded backend) once served.
echo "==> Building frontend (REACT_APP_API_URL='')"
(
  cd "$FRONTEND"
  REACT_APP_API_URL="" corepack yarn build
)

# 2) Copy the frontend build into the Go package so `go:embed dist` can find
#    it. go:embed cannot reach outside its package, hence the copy.
echo "==> Copying frontend build into backend/web/dist"
rm -rf "$WEBDIR"
mkdir -p "$WEBDIR"
cp -a "$FRONTEND/build/." "$WEBDIR/"
touch "$WEBDIR/index.html"  # guard against an empty build

mkdir -p "$OUT"

# 3) Cross-compile. CGO is disabled: the SQLite driver is pure Go, which is what
#    makes a portable .exe possible.
echo "==> Compiling Windows .exe"
(
  cd "$BACKEND"
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
    go build -tags web_embed -trimpath -ldflags "-s -w" \
    -o "$OUT/clientes-crm.exe" .
)

echo "==> Compiling Linux test binary"
(
  cd "$BACKEND"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags web_embed -trimpath -ldflags "-s -w" \
    -o "$OUT/clientes-crm-linux" .
)

echo ""
echo "Done."
echo "  dist/clientes-crm.exe  -> copy to the Windows PC, double-click."
echo "  dist/clientes-crm-linux-> local test binary (this machine)."
echo ""
echo "First run creates data/ and photos/ next to the binary."