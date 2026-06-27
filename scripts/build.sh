#!/bin/bash
set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

mkdir -p bin

echo "=========================================="
echo "  GitHub Fast DNS - Build Script"
echo "=========================================="

echo ""
echo "[1/3] Downloading dependencies..."
go mod tidy

echo ""
echo "[2/3] Building for current platform..."
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
OUTPUT="bin/github-fast-dns-${GOOS}-${GOARCH}"

go build -ldflags "-s -w" -o "$OUTPUT" src/cmd/main.go

echo ""
echo "[3/3] Build completed successfully!"
echo "  Output: $OUTPUT"
echo ""
