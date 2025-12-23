#!/bin/bash

# JSON WASM Build Script
# Compiles WASM binaries using either JSON or stdlib encoding/json

set -e

# Default values
OUTPUT_DIR="web/public"
OUTPUT_FILE="main.wasm"

# Parse arguments
USE_STDLIB=false
if [ "$1" == "stlib" ]; then
    USE_STDLIB=true
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Set source directory based on parameter
if [ "$USE_STDLIB" = true ]; then
    echo "Building with encoding/json (stdlib)..."
    SOURCE_DIR="clients/stdlib"
else
    echo "Building with JSON..."
    SOURCE_DIR="clients/json"
fi

# Build the WASM binary
echo "Compiling $SOURCE_DIR/main.go..."
tinygo build -o "$OUTPUT_DIR/$OUTPUT_FILE" \
    -target wasm \
    -no-debug \
    -opt=z \
    -panic=trap \
    "$SOURCE_DIR/main.go"

# Get file size
FILE_SIZE=$(stat -f%z "$OUTPUT_DIR/$OUTPUT_FILE" 2>/dev/null || stat -c%s "$OUTPUT_DIR/$OUTPUT_FILE" 2>/dev/null)
FILE_SIZE_KB=$((FILE_SIZE / 1024))

echo "✓ Build complete: $OUTPUT_DIR/$OUTPUT_FILE"
echo "  Size: ${FILE_SIZE_KB} KB"

# Create gzipped version for size comparison
gzip -c "$OUTPUT_DIR/$OUTPUT_FILE" > "$OUTPUT_DIR/$OUTPUT_FILE.gz"
GZIP_SIZE=$(stat -f%z "$OUTPUT_DIR/$OUTPUT_FILE.gz" 2>/dev/null || stat -c%s "$OUTPUT_DIR/$OUTPUT_FILE.gz" 2>/dev/null)
GZIP_SIZE_KB=$((GZIP_SIZE / 1024))

echo "  Gzipped: ${GZIP_SIZE_KB} KB"

# Clean up gzip file
rm "$OUTPUT_DIR/$OUTPUT_FILE.gz"

echo ""
echo "To test, serve the web directory and open in a browser:"
echo "  cd web && go run server.go"
