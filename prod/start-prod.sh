#!/bin/sh
set -e

echo "Running migrations..."
/app/migrate

echo "Starting server..."
exec /app/server
