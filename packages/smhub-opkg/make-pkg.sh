#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

PKG_NAME="go-tydom"
VERSION="$1"
BIN_PATH="$2"
OUT_DIR="$3"
OUTPUT_PKG="$4"


if [ -z "$VERSION" ] || [ -z "$OUT_DIR" ]; then
    echo "Usage: $0 <version> <output_directory>"
    exit 1
fi

if [ -d "$OUT_DIR" ]; then
    echo "output directory already exists"
    exit 1
fi

if [ ! -f "$BIN_PATH" ]; then
    echo "go-tydom binary not found. Please build it first."
    exit 1
fi

mkdir -p "$OUT_DIR/data/opt/bin"
echo "2.0" > "$OUT_DIR/debian-binary"
cp -r "$SCRIPT_DIR/control" "$OUT_DIR/control"
cat "$SCRIPT_DIR/control.tmpl" | VERSION=$VERSION envsubst > "$OUT_DIR/control/control"

cp "$BIN_PATH" "$OUT_DIR/data/opt/bin/go-tydom"

tar czf "$OUT_DIR/control.tar.gz" -C "$OUT_DIR/control" .
tar cJf "$OUT_DIR/data.tar.xz" -C "$OUT_DIR/data" .

# Assemblage du paquet IPK
ar r "$OUTPUT_PKG" "$OUT_DIR/debian-binary" "$OUT_DIR/control.tar.gz" "$OUT_DIR/data.tar.xz"

echo "Paquet IPK généré : $OUTPUT_PKG"
