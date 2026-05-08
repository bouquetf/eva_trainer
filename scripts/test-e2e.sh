#!/bin/bash
set -e

echo "=== E2E Tests ==="

docker compose up -d
sleep 5

echo "--- Health Check ---"
curl -f http://localhost:8080/health || (docker compose logs app && exit 1)

echo ""
echo "--- Setting up admin user ---"
docker compose exec -T db psql -U eva_user -d eva_trainer -c "
  INSERT INTO users (discord_id, is_admin) VALUES ('e2e-admin', true)
  ON CONFLICT (discord_id) DO UPDATE SET is_admin = true;
" 2>/dev/null || echo "Warning: could not create admin user"

USER_ID=$(docker compose exec -T db psql -U eva_user -d eva_trainer -t -A -c "
  SELECT id FROM users WHERE discord_id = 'e2e-admin';
" 2>/dev/null | tr -d '[:space:]')

if [ -z "$USER_ID" ]; then
  echo "Failed to get admin user ID"
  exit 1
fi
echo "Admin user ID: $USER_ID"

# Generate JWT using openssl (no python dependency)
JWT_SECRET="test-secret-key-for-manual-validation-32chars"

b64url() {
  echo -n "$1" | base64 -w0 | tr '+/' '-_' | tr -d '='
}

header='{"alg":"HS256","typ":"JWT"}'
payload="{\"sub\":$USER_ID,\"exp\":9999999999,\"iat\":1700000000}"

b64header=$(b64url "$header")
b64payload=$(b64url "$payload")

signature=$(echo -n "${b64header}.${b64payload}" | openssl dgst -sha256 -hmac "$JWT_SECRET" -binary | base64 -w0 | tr '+/' '-_' | tr -d '=')

TOKEN="${b64header}.${b64payload}.${signature}"

echo ""
echo "--- 1. Upload a map ---"
UPLOAD_RESP=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/admin/maps \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "name=E2E-Atlantis" \
  -F "file=@maps/atlantis-map.webp")
UPLOAD_CODE=$(echo "$UPLOAD_RESP" | tail -1)
UPLOAD_BODY=$(echo "$UPLOAD_RESP" | head -n -1)
echo "HTTP $UPLOAD_CODE"

if [ "$UPLOAD_CODE" != "201" ]; then
  echo "FAIL: Expected 201, got $UPLOAD_CODE"
  echo "$UPLOAD_BODY"
  exit 1
fi

MAP_ID=$(echo "$UPLOAD_BODY" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
MAP_NAME=$(echo "$UPLOAD_BODY" | grep -o '"name":"[^"]*"' | head -1 | cut -d\" -f4)
echo "Created map: $MAP_NAME (ID: $MAP_ID)"

PREV_OBSOLETED=$(echo "$UPLOAD_BODY" | grep -o '"previous_version_obsoleted":[^,]*' | head -1)
if echo "$UPLOAD_BODY" | grep -q '"previous_version_obsoleted":null'; then
  echo "OK: No previous version (first upload)"
elif echo "$UPLOAD_BODY" | grep -q '"previous_version_obsoleted"'; then
  echo "OK: Previous version obsoleted"
else
  echo "FAIL: unexpected previous_version_obsoleted"
  echo "$UPLOAD_BODY"
  exit 1
fi

echo ""
echo "--- 2. List maps ---"
LIST_RESP=$(curl -s -w "\n%{http_code}" http://localhost:8080/api/v1/admin/maps \
  -H "Authorization: Bearer ${TOKEN}")
LIST_CODE=$(echo "$LIST_RESP" | tail -1)
LIST_BODY=$(echo "$LIST_RESP" | head -n -1)
echo "HTTP $LIST_CODE"

if [ "$LIST_CODE" != "200" ]; then
  echo "FAIL: Expected 200, got $LIST_CODE"
  echo "$LIST_BODY"
  exit 1
fi

MAP_COUNT=$(echo "$LIST_BODY" | grep -o '"name":"[^"]*"' | wc -l)
echo "Maps in list: $MAP_COUNT"

echo ""
echo "--- 3. Upload same map again (should obsolete first) ---"
UPLOAD2_RESP=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/admin/maps \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "name=E2E-Atlantis" \
  -F "file=@maps/atlantis-map.webp")
UPLOAD2_CODE=$(echo "$UPLOAD2_RESP" | tail -1)
UPLOAD2_BODY=$(echo "$UPLOAD2_RESP" | head -n -1)
echo "HTTP $UPLOAD2_CODE"

if [ "$UPLOAD2_CODE" != "201" ]; then
  echo "FAIL: Expected 201, got $UPLOAD2_CODE"
  echo "$UPLOAD2_BODY"
  exit 1
fi

if echo "$UPLOAD2_BODY" | grep -q '"previous_version_obsoleted":null'; then
  echo "FAIL: Expected previous_version_obsoleted to be present"
  echo "$UPLOAD2_BODY"
  exit 1
fi

PREV_ID=$(echo "$UPLOAD2_BODY" | grep -o '"previous_version_obsoleted":{"id":[0-9]*' | grep -o '[0-9]*$')
echo "Previous version obsoleted: ID $PREV_ID"

echo ""
echo "--- 4. Check old map is obsolete ---"
DETAIL_RESP=$(curl -s -w "\n%{http_code}" "http://localhost:8080/api/v1/admin/maps/${PREV_ID}" \
  -H "Authorization: Bearer ${TOKEN}")
DETAIL_CODE=$(echo "$DETAIL_RESP" | tail -1)
DETAIL_BODY=$(echo "$DETAIL_RESP" | head -n -1)
echo "HTTP $DETAIL_CODE"

if echo "$DETAIL_BODY" | grep -q '"is_obsolete":true'; then
  echo "OK: Old map is obsolete"
else
  echo "FAIL: Expected is_obsolete true"
  echo "$DETAIL_BODY"
  exit 1
fi

echo ""
echo "--- 5. Purge obsolete maps ---"
PURGE_RESP=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/admin/maps/purge \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json")
PURGE_CODE=$(echo "$PURGE_RESP" | tail -1)
PURGE_BODY=$(echo "$PURGE_RESP" | head -n -1)
echo "HTTP $PURGE_CODE"

if [ "$PURGE_CODE" != "200" ]; then
  echo "FAIL: Expected 200, got $PURGE_CODE"
  echo "$PURGE_BODY"
  exit 1
fi

PURGED_COUNT=$(echo "$PURGE_BODY" | grep -o '"purged_ids":\[[^]]*\]' | grep -o '[0-9]' | wc -w)
echo "Purged: $PURGED_COUNT map(s)"

echo ""
echo "--- 6. Public image endpoint ---"
IMAGE_RESP=$(curl -s -w "\n%{http_code}" "http://localhost:8080/api/v1/maps/${MAP_ID}/image" -o /dev/null)
IMAGE_CODE=$(echo "$IMAGE_RESP" | tail -1)
echo "HTTP $IMAGE_CODE"

if [ "$IMAGE_CODE" != "200" ]; then
  echo "FAIL: Expected 200, got $IMAGE_CODE"
  exit 1
fi

IMAGE_CT=$(curl -s -I "http://localhost:8080/api/v1/maps/${MAP_ID}/image" | grep -i content-type | tr -d '\r')
echo "Content-Type: $IMAGE_CT"

echo ""
echo "--- 7. Image of non-existent map (404) ---"
NOTFOUND_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/api/v1/maps/99999/image")
echo "HTTP $NOTFOUND_CODE"

if [ "$NOTFOUND_CODE" != "404" ]; then
  echo "FAIL: Expected 404, got $NOTFOUND_CODE"
  exit 1
fi

echo "OK: 404 for non-existent map"

echo ""
echo "=== E2E Tests Passed ==="

docker compose down
