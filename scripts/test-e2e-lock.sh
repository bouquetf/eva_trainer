#!/bin/bash
# Test E2E : Map Locking Lifecycle
# Requires a running server on BASE_URL and valid JWT tokens

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN_A="${TOKEN_A:-}"
TOKEN_B="${TOKEN_B:-}"
TEAM_ID="${TEAM_ID:-1}"
MAP_ID="${MAP_ID:-1}"
ZONE_ID=""

if [ -z "$TOKEN_A" ] || [ -z "$TOKEN_B" ]; then
  echo "Usage: TOKEN_A=<token> TOKEN_B=<token> [TEAM_ID=1] [MAP_ID=1] $0"
  echo "Requires two valid JWT tokens for two different editors."
  exit 1
fi

PASS=0
FAIL=0

check() {
  local desc="$1"
  local cond="$2"
  if $cond; then
    echo "  [PASS] $desc"
    PASS=$((PASS + 1))
  else
    echo "  [FAIL] $desc"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== Test E2E : Cycle de vie du verrouillage ==="
echo ""

# 1. Editor A acquires lock
echo "1. Editor A acquiert le lock..."
RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json")
check "Lock acquis par A" "echo '$RESP' | grep -q '\"id\"'"

# 2. Editor B tries to acquire → 409
echo "2. Editor B tente d'acquérir..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json")
check "HTTP 409 pour B" "test '$HTTP_CODE' = '409'"

RESP_B=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json")
check "editor_name dans 409" "echo '$RESP_B' | grep -q '\"editor_name\"'"

# 3. Editor A creates a zone (should succeed)
echo "3. Editor A crée une zone..."
RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/zones" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"polygons":[{"points":[{"x":100,"y":100},{"x":200,"y":200},{"x":300,"y":100}]}],"name":"e2e-top"}')
check "Zone créée par A" "echo '$RESP' | grep -q '\"id\"'"

ZONE_ID=$(echo "$RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)

# 4. Editor B tries to create a zone → 423
echo "4. Editor B essaie de créer une zone (sans lock)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/zones" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{"polygons":[{"points":[{"x":50,"y":50},{"x":150,"y":150},{"x":250,"y":50}]}],"name":"e2e-secret"}')
check "HTTP 423 pour B (resource locked)" "test '$HTTP_CODE' = '423'"

# 5. Editor A releases the lock
echo "5. Editor A libère le lock..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_A")
check "Lock libéré par A (HTTP 200)" "test '$HTTP_CODE' = '200'"

# 6. Editor B acquires the lock
echo "6. Editor B acquiert le lock..."
RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json")
check "Lock acquis par B" "echo '$RESP' | grep -q '\"id\"'"

# 7. Check lock status
echo "7. Vérification du statut du lock..."
RESP=$(curl -s -X GET "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B")
check "editor_name présent dans status" "echo '$RESP' | grep -q '\"editor_name\"'"
check "remaining_seconds présent" "echo '$RESP' | grep -q '\"remaining_seconds\"'"

# 8. Cleanup
echo "8. Nettoyage..."
if [ -n "$ZONE_ID" ]; then
  curl -s -X DELETE "$BASE_URL/api/v1/teams/$TEAM_ID/zones/$ZONE_ID" \
    -H "Authorization: Bearer $TOKEN_A" > /dev/null
fi
curl -s -X DELETE "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B" > /dev/null

echo ""
echo "=== Résultat : $PASS passed, $FAIL failed ==="
exit $FAIL
