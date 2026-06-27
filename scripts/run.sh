#!/bin/bash
set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

# Detect binary name
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
BINARY="bin/github-fast-dns-${GOOS}-${GOARCH}"

echo "=========================================="
echo "  GitHub Fast DNS - One-Click Run"
echo "=========================================="

# Check if running as root (required for port 53)
if [ "$EUID" -ne 0 ]; then 
    echo "This script requires root privileges to bind port 53."
    echo "Please run with sudo: sudo $0"
    exit 1
fi

if [ ! -f "$BINARY" ]; then
    echo "Binary not found, building first..."
    bash scripts/build.sh
fi

echo ""
echo "Starting GitHub Fast DNS service..."
echo "Press Ctrl+C to stop"
echo ""

"$BINARY"
