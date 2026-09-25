#!/usr/bin/env bash
# Builds a single self-contained Windows .exe (and a Linux twin for local
# testing) that hosts the React UI + Go API + SQLite DB, ready to be
# double-clicked on a Windows machine with nothing installed.
#
# Windows resources (icon + file properties + manifest) are generated with
# go-winres from backend/winres/winres.json and embedded into the exe, which
# gives the file a proper version and icon and reduces antivirus/SmartScreen
# false positives.
#
# Output:
#   dist/historial.exe                (Windows 64-bit, the file to copy/share)
#   dist/historial-windows-v0.1.0.zip (release zip: Historial/ + scripts + README + sha256)
#   dist/historial-linux              (Linux binary to verify locally)
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

VERSION="0.1.0"
ZIP="$OUT/historial-windows-v$VERSION.zip"

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

# 3) Generate Windows resources (icon, version info, manifest) into a .syso
#    object. go build embeds it in the exe for windows/amd64 only; the linux
#    build ignores it thanks to the "_windows_amd64" suffix.
echo "==> Generating Windows resources (icon/version/manifest) with go-winres"
(
  cd "$BACKEND"
  go run github.com/tc-hib/go-winres@v0.3.3 make --in winres/winres.json --arch amd64
)

# 4) Cross-compile. CGO is disabled: the SQLite driver is pure Go, which is what
#    makes a portable .exe possible.
echo "==> Compiling Windows .exe"
(
  cd "$BACKEND"
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
    go build -tags web_embed -trimpath -ldflags "-s -w" \
    -o "$OUT/historial.exe" .
)

echo "==> Compiling Linux test binary"
(
  cd "$BACKEND"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags web_embed -trimpath -ldflags "-s -w" \
    -o "$OUT/historial-linux" .
)

# The .syso object is only needed while compiling; remove it so a bare
# `go build` from this repo never picks up stale resources.
rm -f "$BACKEND"/rsrc_windows_*.syso

# 5) Assemble the release zip: everything the user needs inside a Historial/
#    folder, plus the exe's sha256 checksum.
echo "==> Assembling release zip ($ZIP)"
STAGE="$OUT/Historial"
rm -rf "$STAGE" "$ZIP"
mkdir -p "$STAGE"
cp "$OUT/historial.exe" "$STAGE/"
( cd "$STAGE" && sha256sum historial.exe > historial.exe.sha256 )
cp "$ROOT/scripts/windows/"*.cmd "$STAGE/"
cp "$ROOT/scripts/windows/README-windows.txt" "$STAGE/"
( cd "$OUT" && zip -qr "$ZIP" "Historial" )
rm -rf "$STAGE"

echo ""
echo "Done."
echo "  $ZIP  <- esto es lo que se comparte/instala"
echo "  dist/historial.exe   -> raw exe (same file as inside the zip)"
echo "  dist/historial-linux -> local test binary (this machine)."
echo ""
echo "First run creates data/ and photos/ next to the binary."