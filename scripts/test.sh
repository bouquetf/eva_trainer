#!/bin/bash
set -e

echo "=== Unit Tests ==="
go test ./...

echo "=== Integration Tests ==="
go test --tags=integration ./...

echo "=== All tests passed ==="
