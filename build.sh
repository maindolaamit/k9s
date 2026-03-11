#!/bin/bash
set -e

VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')

echo "=================================================="
echo "  Building k9s-custom"
echo "=================================================="
echo "  Version: $VERSION"
echo "  Commit:  $COMMIT"
echo "  Date:    $DATE"
echo "  Branch:  $(git branch --show-current)"
echo "=================================================="
echo ""

# Clean previous build
rm -f execs/k9s

# Build
echo "Building..."
CGO_ENABLED=0 go build \
  -ldflags "-w -s \
    -X github.com/derailed/k9s/cmd.version=$VERSION-custom \
    -X github.com/derailed/k9s/cmd.commit=$COMMIT \
    -X github.com/derailed/k9s/cmd.date=$DATE" \
  -a -tags=netgo -o execs/k9s main.go

if [ -f execs/k9s ]; then
    SIZE=$(ls -lh execs/k9s | awk '{print $5}')
    echo ""
    echo "✓ Build successful!"
    echo ""
    echo "  Binary: execs/k9s ($SIZE)"
    echo ""
    echo "Installation options:"
    echo ""
    echo "  1) Test version (recommended):"
    echo "     cp execs/k9s /usr/local/bin/k9s-custom"
    echo "     k9s-custom"
    echo ""
    echo "  2) Replace current k9s:"
    echo "     cp execs/k9s /opt/homebrew/bin/k9s"
    echo ""
    echo "  3) Create symlink:"
    echo "     ln -sf $(pwd)/execs/k9s /usr/local/bin/k9s"
    echo ""
else
    echo "❌ Build failed!"
    exit 1
fi
