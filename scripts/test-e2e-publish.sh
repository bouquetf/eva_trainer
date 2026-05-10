#!/bin/bash
# Test E2E : Zone Publishing and Template Auto-update
# 
# Prérequis :
#   1. Un serveur en fonctionnement (BASE_URL)
#   2. Tokens JWT valides pour un capitaine, un capitaine d'une autre team, et un admin
#   3. curl, grep, cut installés
#
# Usage :
#   BASE_URL="http://localhost:8080" \
#   TOKEN_CAPTAIN="<token>" \
#   TOKEN_CAPTAIN2="<token>" \
#   TOKEN_ADMIN="<token>" \
#   TEAM_ID=1 \
#   TEAM2_ID=2 \
#   MAP_ID=1 \
#   bash scripts/test-e2e-publish.sh

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN_CAPTAIN="${TOKEN_CAPTAIN:?Must set TOKEN_CAPTAIN}"
TOKEN_CAPTAIN2="${TOKEN_CAPTAIN2:?Must set TOKEN_CAPTAIN2}"
TOKEN_ADMIN="${TOKEN_ADMIN:?Must set TOKEN_ADMIN}"
TEAM_ID="${TEAM_ID:?Must set TEAM_ID}"
TEAM2_ID="${TEAM2_ID:?Must set TEAM2_ID}"
MAP_ID="${MAP_ID:?Must set MAP_ID}"

PASS=0
FAIL=0

pass() { PASS=$((PASS+1)); echo "  ✅ $1"; }
fail() { FAIL=$((FAIL+1)); echo "  ❌ $1"; }

echo "============================================"
echo "  Test E2E : Publication de carte et template"
echo "============================================"
echo ""

# ---- 1. Capitaine publie la carte ----
echo "1. Capitaine publie la carte..."
RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/publish" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN" \
  -H "Content-Type: application/json")
if echo "$RESP" | grep -q '"is_public":true'; then
  TEMPLATE_ID=$(echo "$RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
  pass "Carte publiée (template_id=$TEMPLATE_ID)"
else
  fail "Échec publication : $RESP"
fi

# ---- 2. Vérifier le statut ----
echo "2. Vérification du statut..."
RESP=$(curl -s -X GET "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/publish-status" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN")
if echo "$RESP" | grep -q '"published":true'; then
  pass "Statut = publiée"
else
  fail "Statut incorrect : $RESP"
fi

# ---- 3. Double publication → 409 ----
echo "3. Double publication..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/publish" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN" \
  -H "Content-Type: application/json")
if [ "$HTTP_CODE" = "409" ]; then
  pass "409 Conflict reçu"
else
  fail "Attendu 409, reçu $HTTP_CODE"
fi

# ---- 4. Créer une zone → template mis à jour automatiquement ----
echo "4. Création de zone après publication..."
ZONE_RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/zones" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN" \
  -H "Content-Type: application/json" \
  -d '{"polygons":[{"points":[{"x":400,"y":400},{"x":500,"y":500},{"x":600,"y":400}]}],"name":"new-zone"}')
if echo "$ZONE_RESP" | grep -q '"id"'; then
  pass "Nouvelle zone créée"
else
  fail "Échec création zone : $ZONE_RESP"
fi

echo "   Vérification mise à jour template..."
TEMPLATE_CHECK=$(curl -s -X GET "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/publish-status" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN")
if echo "$TEMPLATE_CHECK" | grep -q '"is_public":true'; then
  pass "Template toujours public après modification"
else
  fail "Template perdu après modification"
fi

# ---- 5. Publication par non-capitaine (autre équipe) ----
echo "5. Publication par non-capitaine (autre équipe)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/teams/$TEAM2_ID/maps/$MAP_ID/publish" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN2" \
  -H "Content-Type: application/json")
echo "   Code reçu : $HTTP_CODE (selon permissions)"

# ---- 6. Nettoyage (admin) ----
echo "6. Nettoyage (admin)..."
if [ -n "$TEMPLATE_ID" ]; then
  ADMIN_RESP=$(curl -s -X DELETE "$BASE_URL/api/v1/admin/templates/$TEMPLATE_ID" \
    -H "Authorization: Bearer $TOKEN_ADMIN")
  if [ $? -eq 0 ]; then
    pass "Template #$TEMPLATE_ID supprimé"
  else
    fail "Échec suppression template #$TEMPLATE_ID"
  fi
else
  echo "   (aucun template_id à nettoyer)"
fi

echo ""
echo "============================================"
echo "  Résultat : $PASS passés, $FAIL échoués"
echo "============================================"
exit $FAIL
