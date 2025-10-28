#!/bin/bash

# Build script for syncenv
# This script handles Go version conflicts by unsetting GOROOT and GOPATH

set -e

echo "Building syncenv..."
echo "Using Go version: $(unset GOROOT && unset GOPATH && go version)"

# Unset conflicting environment variables
unset GOROOT
unset GOPATH

# Build the binary
go build -o syncenv -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" ./cmd/syncenv

echo "Build successful!"
echo "Binary location: ./syncenv"
echo "Size: $(ls -lh syncenv | awk '{print $5}')"
