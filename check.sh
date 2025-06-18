#!/bin/bash
# Run all CI checks locally before committing

set -e

echo "=== Running CI checks locally ==="
echo

echo "1. Checking Go formatting..."
if [ -n "$(gofmt -l .)" ]; then
    echo "❌ Go code is not formatted. Files that need formatting:"
    gofmt -l .
    echo
    echo "Run 'gofmt -w .' to fix formatting"
    exit 1
fi
echo "✅ Go formatting OK"
echo

echo "2. Running go vet..."
go vet ./...
echo "✅ go vet OK"
echo

echo "3. Running tests..."
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
echo "✅ Tests OK"
echo

echo "4. Building binary..."
go build -o bin/ratchet ./cmd/ratchet
echo "✅ Build OK"
echo

echo "5. Testing binary..."
./bin/ratchet --version
echo "✅ Binary test OK"
echo

echo "6. Running golangci-lint (if available)..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run --timeout=5m
    echo "✅ golangci-lint OK"
else
    echo "⚠️  golangci-lint not installed, skipping"
fi
echo

echo "=== All checks passed! ✅ ==="
echo "Your code is ready to commit and push."