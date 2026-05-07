#!/bin/bash
set -e

echo "=== E2E Tests ==="

docker compose up -d
sleep 5

curl -f http://localhost:8080/health || (docker compose logs app && exit 1)

echo "=== E2E Tests Passed ==="

docker compose down
