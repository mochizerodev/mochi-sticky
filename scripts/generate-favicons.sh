#!/usr/bin/env bash
set -euo pipefail

# generate-favicons.sh
# Generate PNG favicons and a favicon.ico from the project's SVG assets.
# Requires either `rsvg-convert` (from librsvg) or ImageMagick `convert`.
# Optionally requires `icotool` (from icoutils) to build multi-size .ico files.

SRC_SVG="assets/small-sticky.svg"
OUT_DIR="assets/brand/favicons"
mkdir -p "$OUT_DIR"

if [ ! -f "$SRC_SVG" ]; then
  echo "Source SVG not found: $SRC_SVG"
  echo "Available SVGs: assets/*.svg"
  ls -1 assets/*.svg || true
  exit 1
fi

# Prefer rsvg-convert if available for sharp results
if command -v rsvg-convert >/dev/null 2>&1; then
  echo "Using rsvg-convert to render PNGs"
  rsvg-convert -w 16 -h 16 "$SRC_SVG" -o "$OUT_DIR/favicon-16.png"
  rsvg-convert -w 32 -h 32 "$SRC_SVG" -o "$OUT_DIR/favicon-32.png"
  rsvg-convert -w 180 -h 180 "$SRC_SVG" -o "$OUT_DIR/apple-touch-icon-180.png"
  rsvg-convert -w 1200 -h 630 "$SRC_SVG" -o "$OUT_DIR/social-preview.png"
elif command -v convert >/dev/null 2>&1; then
  echo "Using ImageMagick convert to render PNGs"
  convert -background none "$SRC_SVG" -resize 16x16 "$OUT_DIR/favicon-16.png"
  convert -background none "$SRC_SVG" -resize 32x32 "$OUT_DIR/favicon-32.png"
  convert -background none "$SRC_SVG" -resize 180x180 "$OUT_DIR/apple-touch-icon-180.png"
  convert -background none "$SRC_SVG" -resize 1200x630 "$OUT_DIR/social-preview.png"
else
  echo "No SVG rendering tool found. Install librsvg (`rsvg-convert`) or ImageMagick (`convert`)."
  exit 2
fi

# Create favicon.ico: prefer icotool if available, else use convert if it supports ICO
if command -v icotool >/dev/null 2>&1; then
  echo "Using icotool to build favicon.ico"
  icotool -c -o "$OUT_DIR/favicon.ico" "$OUT_DIR/favicon-16.png" "$OUT_DIR/favicon-32.png"
elif command -v convert >/dev/null 2>&1; then
  echo "Using convert to build favicon.ico"
  convert "$OUT_DIR/favicon-16.png" "$OUT_DIR/favicon-32.png" "$OUT_DIR/favicon.ico"
else
  echo "No tool to build .ico found. Install icoutils (icotool) or ImageMagick for .ico generation."
  echo "You still have PNGs in $OUT_DIR"
fi

cat <<EOF
Generated favicons in: $OUT_DIR
- favicon-16.png
- favicon-32.png
- apple-touch-icon-180.png
- social-preview.png
- favicon.ico (if tool available)

To publish:
- Commit these files and push to GitHub.
- For GitHub repo social preview: go to Repository Settings -> Social preview -> Upload "$OUT_DIR/social-preview.png".
EOF
