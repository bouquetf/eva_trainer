#!/bin/bash
set -e

echo "=== Unit Tests ==="
go test ./...

echo "=== Integration Tests ==="
go test --tags=integration ./...

echo "=== Frontend Unit Tests ==="
cd frontend && npx vitest run && cd ..

echo "=== All tests passed ==="
