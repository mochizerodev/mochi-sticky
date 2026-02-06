#!/usr/bin/env bash
set -e
ROOT="$(dirname "$(dirname "$0")")"
ASSETS="$ROOT/assets"
OUT="$ASSETS/png"
mkdir -p "$OUT"

# prefer inkscape (modern CLI), fallback to rsvg-convert
for svg in "$ASSETS"/*.svg; do
  name="$(basename "$svg" .svg)"
  out="$OUT/${name}.png"
  if command -v inkscape >/dev/null 2>&1; then
    inkscape "$svg" --export-type=png --export-filename="$out" --export-width=1024
  elif command -v rsvg-convert >/dev/null 2>&1; then
    rsvg-convert -w 1024 -o "$out" "$svg"
  else
    echo "No SVG renderer found (install inkscape or librsvg)."
    exit 2
  fi
  echo "Exported $out"
done
