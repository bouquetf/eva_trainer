# Tâche #003 - Story #010 : Tests d'intégration, E2E et documentation

## Objectif

Créer les tests d'intégration backend pour le verrouillage concurrent, un test E2E pour le cycle de vie complet du verrou, et la documentation utilisateur du mécanisme de verrouillage.

## Contexte

- Story #010 : `docs/stories/story-010.md` - Map Locking Mechanism for Concurrent Editing
- Dépend de : Tâche #001 (API lock enrichie), Tâche #002 (frontend lock UI)
- Ordre d'exécution : 3ème (dernière tâche)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Partie 1 - Tests d'intégration backend : Accès concurrent

**Fichier :** `internal/handlers/edit_lock_test.go` (créer ou enrichir)

Tester le scénario de deux éditeurs tentant de modifier la même carte :

1. **Deux éditeurs : le premier acquiert le lock, le second reçoit 409**
   - Préparation :
     - Créer un éditeur A (user ID 1) et un éditeur B (user ID 2) dans la table users
     - Créer une équipe et ajouter les deux utilisateurs comme "editor"
     - Créer une carte
   - Étape 1 : Éditeur A acquiert le lock via POST `/api/v1/teams/1/maps/1/lock` → HTTP 200, `lock.editor_id = 1`
   - Étape 2 : Éditeur B tente d'acquérir le lock → HTTP 409, `error = "lock_already_acquired"`, `locked_by = 1`, `editor_name` non vide
   - Étape 3 : Éditeur A essaie de modifier une zone → doit réussir (le handler vérifie `IsLockedByEditor`)

2. **Deux éditeurs : le premier libère le lock, le second peut acquérir**
   - Étape 1 : Éditeur A acquiert le lock → HTTP 200
   - Étape 2 : Éditeur A libère le lock via DELETE → HTTP 200
   - Étape 3 : Éditeur B acquiert le lock → HTTP 200, `lock.editor_id = 2`

3. **Lock timeout : le lock expire, un autre éditeur peut acquérir**
   - Étape 1 : Éditeur A acquiert le lock → HTTP 200
   - Étape 2 : Modifier manuellement la base de données pour que `expires_at = NOW() - INTERVAL '1 minute'` (simuler le timeout)
   - Étape 3 : Éditeur B acquiert le lock → HTTP 200 (le lock expiré est remplacé automatiquement)

4. **GetLockStatus avec editor_name et remaining_seconds**
   - Préparation : Éditeur A acquiert le lock
   - Requête : GET `/api/v1/teams/1/maps/1/lock`
   - Vérification : `locked = true`, `lock.editor_id = 1`, `lock.editor_name` non vide, `lock.remaining_seconds > 0`

5. **RefreshLock met à jour remaining_seconds**
   - Préparation : Éditeur A acquiert le lock, attendre 1 seconde
   - Requête : PUT refresh
   - Vérification : `remaining_seconds` est ~900 (15 minutes), pas ~899 (a été refreshé)

6. **Modification sans lock → 423 Locked**
   - Préparation : Pas de lock acquis
   - Requête : POST `/api/v1/teams/1/maps/1/zones` (create zone)
   - Vérification : HTTP 423, `error = "resource_locked"`

7. **Admin cleanup des locks expirés**
   - Préparation : Éditeur A acquiert le lock, modifier `expires_at` pour le faire expirer
   - Requête : POST `/api/v1/admin/locks/cleanup` (en tant qu'admin)
   - Vérification : HTTP 200, `cleaned = 1`
   - Vérification : GET lock status → `locked = false`

8. **Force clear all locks**
   - Préparation : Deux locks actifs
   - Requête : DELETE `/api/v1/admin/locks` (en tant qu'admin)
   - Vérification : HTTP 200, `count = 2`

**Fichier :** `internal/handlers/edit_lock_test.go`

**Contraintes techniques :**
- Utiliser `gin.CreateTestContext` ou `httptest.NewServer` comme les autres tests handlers
- Créer les utilisateurs, équipes et cartes nécessaires dans la base de test
- Suivre le pattern de `internal/handlers/zone_test.go` pour la configuration du test
- Pour simuler le timeout, exécuter une requête SQL directe : `UPDATE edit_locks SET expires_at = NOW() - INTERVAL '1 minute' WHERE team_id = $1 AND map_id = $2`
- Nettoyer les données de test après chaque test (dans `t.Cleanup` ou `defer`)

### Partie 2 - Tests d'intégration frontend : cycle de vie du verrou

**Fichier :** `frontend/src/components/MapEditor/MapEditor.test.tsx`

Tester l'intégration du verrouillage dans l'éditeur :

1. **Acquisition du lock au chargement**
   - Mock :
     - `acquireLock` retourne un lock valide avec `remaining_seconds: 900`
     - `listZones` retourne une liste vide
     - `listTeamMaps` retourne une liste de maps
   - Rendre `MapEditor` avec `teamId=1, mapId=1`
   - Vérifier que `acquireLock` a été appelé avec `(1, 1)`
   - Vérifier que `isLocked` et `lockedByMe` sont true
   - Vérifier que LockIndicator affiche "Édition en cours"

2. **Lock déjà pris par un autre → mode lecture seule**
   - Mock :
     - `acquireLock` rejette avec `ApiError(409, { locked_by: 2, editor_name: "Player456" })`
   - Rendre `MapEditor` avec `teamId=1, mapId=1`
   - Vérifier que `canEdit` est false
   - Vérifier que LockIndicator affiche "Verrouillé par Player456"

3. **Lock refresh périodique**
   - Mock :
     - `acquireLock` retourne un lock valide
     - `refreshLock` retourne un lock refreshé
   - Rendre `MapEditor`
   - Avancer le temps de 10 minutes (vi.advanceTimersByTime)
   - Vérifier que `refreshLock` a été appelé

4. **beforeunload libère le lock**
   - Mock :
     - `acquireLock` retourne un lock valide
     - `getApiClient().post` pour la requête keepalive
   - Rendre `MapEditor`
   - Simuler un événement `beforeunload` sur `window`
   - Vérifier que `POST /api/v1/teams/1/maps/1/lock/release` a été appelé

5. **Lock perdu → ré-acquisition via bouton**
   - Mock :
     - `acquireLock` retourne un lock valide (premier appel)
     - `refreshLock` rejette (simuler perte)
   - Rendre `MapEditor`
   - Attendre que le refresh échoue → `lockedByMe` devient false
   - Vérifier que le bouton "Ré-acquérir le verrou" est affiché
   - Cliquer sur le bouton → `acquireLock` est rappelé

### Partie 3 - Tests E2E

**Fichier :** `scripts/test-e2e-lock.sh` ou ajout au fichier E2E existant

Scénario E2E complet (via l'API, pas de navigateur réel) :

```bash
#!/bin/bash
# Test E2E : Map Locking Lifecycle

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN_A="<token_editor_a>"
TOKEN_B="<token_editor_b>"
TEAM_ID=1
MAP_ID=1
ZONE_ID=""

echo "=== Test E2E : Cycle de vie du verrouillage ==="

# 1. Éditeur A acquiert le lock
echo "1. Éditeur A acquiert le lock..."
RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json")
echo "$RESP" | grep -q '"id"' && echo "  ✅ Lock acquis par A" || echo "  ❌ Échec acquisition A"

# 2. Éditeur B tente d'acquérir → 409
echo "2. Éditeur B tente d'acquérir..."
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json")
[ "$RESP" = "409" ] && echo "  ✅ 409 Conflict reçu par B" || echo "  ❌ Attendu 409, reçu $RESP"

# Vérifier que le body contient editor_name
RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json")
echo "$RESP" | grep -q '"editor_name"' && echo "  ✅ editor_name présent dans 409" || echo "  ❌ editor_name manquant"

# 3. Éditeur A crée une zone (succès)
echo "3. Éditeur A crée une zone..."
RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/zones" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"polygons":[{"points":[{"x":100,"y":100},{"x":200,"y":200},{"x":300,"y":100}]}],"name":"top"}')
echo "$RESP" | grep -q '"id"' && echo "  ✅ Zone créée par A" || echo "  ❌ Échec création zone"

# Extraire zone_id
ZONE_ID=$(echo "$RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)

# 4. Éditeur B essaie de créer une zone → 423
echo "4. Éditeur B essaie de créer une zone (sans lock)..."
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/zones" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{"polygons":[{"points":[{"x":50,"y":50},{"x":150,"y":150},{"x":250,"y":50}]}],"name":"secret"}')
[ "$RESP" = "423" ] && echo "  ✅ 423 Locked reçu par B" || echo "  ❌ Attendu 423, reçu $RESP"

# 5. Éditeur A libère le lock
echo "5. Éditeur A libère le lock..."
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_A")
[ "$RESP" = "200" ] && echo "  ✅ Lock libéré par A" || echo "  ❌ Échec libération"

# 6. Éditeur B acquiert le lock
echo "6. Éditeur B acquiert le lock..."
RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json")
echo "$RESP" | grep -q '"id"' && echo "  ✅ Lock acquis par B" || echo "  ❌ Échec acquisition B"

# 7. Vérifier le statut du lock
echo "7. Vérification du statut du lock..."
RESP=$(curl -s -X GET "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B")
echo "$RESP" | grep -q '"editor_name"' && echo "  ✅ editor_name présent dans status" || echo "  ❌ editor_name manquant"
echo "$RESP" | grep -q '"remaining_seconds"' && echo "  ✅ remaining_seconds présent" || echo "  ❌ remaining_seconds manquant"

# 8. Nettoyage : supprimer la zone créée
echo "8. Nettoyage..."
curl -s -X DELETE "$BASE_URL/api/v1/teams/$TEAM_ID/zones/$ZONE_ID" \
  -H "Authorization: Bearer $TOKEN_A" > /dev/null
curl -s -X DELETE "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/lock" \
  -H "Authorization: Bearer $TOKEN_B" > /dev/null

echo ""
echo "=== Résultat : Test E2E terminé ==="
```

Ce script doit être placé dans `scripts/test-e2e-lock.sh` et exécutable.

**Fichier à créer :**
- `scripts/test-e2e-lock.sh` — Script E2E pour le cycle de vie complet du verrouillage

### Partie 4 - Documentation utilisateur

**Fichier à créer :** `docs/map-locking.md`

Documenter le système de verrouillage pour l'utilisateur final :

```markdown
# Verrouillage des cartes

## Principe

Pour éviter les conflits de modification simultanée, une carte ne peut être
modifiée que par un seul éditeur à la fois au sein d'une équipe.

## Comment ça fonctionne

1. Quand un éditeur ouvre l'éditeur de carte, un **verrou d'édition** est
   automatiquement acquis.
2. Tant que le verrou est actif, l'éditeur peut librement modifier les zones
   de la carte.
3. Le verrou expire après **15 minutes d'inactivité**.
4. Quand l'éditeur quitte l'éditeur (fermeture de l'onglet, navigation), le
   verrou est libéré automatiquement.

## Si un autre éditeur tente de modifier la carte

- Un message d'erreur lui indique que la carte est en cours d'édition.
- Il voit le nom de l'éditeur qui a le verrou.
- Il doit attendre la libération du verrou (soit manuelle, soit après 15
  minutes d'inactivité).
- Il peut réessayer d'acquérir le verrou via le bouton "Ré-acquérir".

## Indicateur visuel

L'éditeur affiche en haut de l'interface :

| Statut | Affichage |
|--------|-----------|
| ✅ Verrou acquis (vous) | "Édition en cours" avec temps restant |
| 🔴 Verrou pris (autre) | "Verrouillé par [nom]" avec temps restant |
| 🟡 Verrou perdu | "Verrou d'édition perdu" + bouton ré-acquisition |

## Dépannage

- **Verrou perdu pendant l'édition** : Cliquez sur "Ré-acquérir le verrou".
  Si un autre éditeur a pris le verrou entre-temps, vous devrez attendre.
- **Fermeture brutale du navigateur** : Le verrou est automatiquement libéré
  après 15 minutes d'inactivité. Aucune action requise.
- **Impossible d'acquérir le verrou** : Vérifiez que vous avez le rôle
  "Éditeur" ou supérieur dans l'équipe.

## Administration

Les administrateurs peuvent forcer la libération de tous les verrous via
l'interface d'administration du site.

## Cas particuliers

- Navigation vers une autre carte : le verrou de la carte précédente est
  libéré et un nouveau verrou est acquis pour la nouvelle carte.
- Session expirée : le verrou reste actif jusqu'à son expiration naturelle
  (15 min max).
- Plusieurs onglets : chaque onglet tente d'acquérir son propre verrou.
  Seul le premier onglet obtient le verrou.
```

**Fichier à créer :**
- `docs/map-locking.md` — Documentation utilisateur du verrouillage

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/edit_lock_test.go` | Créer/Modifier | Tests d'intégration lock concurrent |
| `frontend/src/components/MapEditor/MapEditor.test.tsx` | Modifier | Tests d'intégration frontend lock |
| `scripts/test-e2e-lock.sh` | Créer | Script E2E cycle de vie du verrou |
| `docs/map-locking.md` | Créer | Documentation utilisateur |

### Contraintes techniques

- **Tests backend** : Utiliser la base de test PostgreSQL existante. Voir `internal/models/setup_test.go` pour le pattern.
- **Tests frontend** : Utiliser `vi.mock` pour mocker les appels API. Utiliser des timers virtuels (`vi.useFakeTimers()`) pour tester le refresh périodique.
- **Script E2E** : Nécessite un serveur en fonctionnement et des tokens JWT valides. Le script est fourni à titre d'exemple et peut nécessiter des ajustements selon l'environnement de test.
- **Documentation** : En français (comme le reste de la documentation utilisateur). Format Markdown. Ajouter un lien depuis le README.md si pertinent.
