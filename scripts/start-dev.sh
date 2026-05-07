#!/bin/sh
set -e

echo "Running migrations..."
go run ./cmd/migrate/

echo "Starting development server..."
exec air -c .air.toml
