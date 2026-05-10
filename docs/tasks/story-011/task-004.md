# Tâche #004 - Story #011 : Tests d'intégration, E2E et documentation

## Objectif

Créer les tests d'intégration backend pour le système de publication et mise à jour automatique des templates, le script E2E pour le cycle de vie complet, et la documentation utilisateur du système de templates.

## Contexte

- Story #011 : `docs/stories/story-011.md` - Zone Publishing and Template System
- Dépend de : Tâche #001 (API publication), Tâche #002 (auto-update), Tâche #003 (frontend)
- Ordre d'exécution : 4ème (dernière tâche)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Partie 1 - Tests d'intégration backend : Publication et template

**Fichier :** `internal/handlers/template_test.go` (créer)

Tester le scénario de publication de carte et de template :

1. **Publication d'une carte avec zones**
   - Préparation :
     - Créer une équipe, un capitaine, une carte
     - Créer 2 zones avec polygones et noms
   - Requête : POST `/api/v1/teams/1/maps/1/publish` (en tant que capitaine)
   - Vérification : HTTP 201, `template.map_id = 1`, `template.is_public = true`
   - Vérification : Le template contient 2 zones avec les mêmes polygones et noms

2. **Publication sans zones (carte vide)**
   - Préparation : Carte sans zone
   - Requête : POST publish (en tant que capitaine)
   - Vérification : HTTP 201, template créé avec 0 zones

3. **Double publication → 409 Conflict**
   - Préparation : Carte déjà publiée
   - Requête : POST publish
   - Vérification : HTTP 409, `error = "already_published"`

4. **Publication par un non-capitaine → 403 Forbidden**
   - Préparation : Membre ou éditeur tente de publier
   - Requête : POST publish
   - Vérification : HTTP 403

5. **GetPublishStatus : carte publiée**
   - Préparation : Carte publiée
   - Requête : GET publish-status
   - Vérification : `published = true`, `template.id` présent

6. **GetPublishStatus : carte non publiée**
   - Préparation : Carte sans publication
   - Requête : GET publish-status
   - Vérification : `published = false`

**Fichier :** `internal/services/template_sync_test.go` (créer)

Tester la synchro automatique :

7. **Création de zone sur carte publiée → template mis à jour**
   - Préparation : Carte publiée avec 2 zones
   - Action : Créer une 3ème zone
   - Vérification : Le template contient 3 zones

8. **Suppression de zone sur carte publiée → template mis à jour**
   - Préparation : Carte publiée avec 2 zones
   - Action : Supprimer 1 zone
   - Vérification : Le template contient 1 zone

9. **Modification de zone sur carte NON publiée → aucune action**
   - Préparation : Carte non publiée avec 1 zone
   - Action : Créer une 2ème zone
   - Vérification : Pas de template créé (FindByMapID retourne nil)

10. **Modification de polygone → template mis à jour**
    - Préparation : Carte publiée avec 1 zone ayant 1 polygone de 3 points
    - Action : Modifier les points du polygone
    - Vérification : `polygon_data` du template mis à jour

11. **Renommage de zone → template mis à jour**
    - Préparation : Carte publiée avec zone nommée "top"
    - Action : Renommer en "secret"
    - Vérification : `name` du template = "secret"

### Partie 2 - Tests d'intégration frontend

**Fichier :** `frontend/src/components/MapEditor/PublishButton.test.tsx`

(Compléter les tests unitaires décrits dans la Tâche #003)

### Partie 3 - Tests E2E

**Fichier :** `scripts/test-e2e-publish.sh` (créer)

Scénario E2E complet (via l'API) :

```bash
#!/bin/bash
# Test E2E : Zone Publishing and Template Auto-update

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN_CAPTAIN="<token_captain>"
TOKEN_CAPTAIN2="<token_captain_team2>"
TEAM_ID=1
TEAM2_ID=2
MAP_ID=1

echo "=== Test E2E : Publication de carte et template ==="

# 1. Capitaine publie la carte
echo "1. Capitaine publie la carte..."
RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/publish" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN" \
  -H "Content-Type: application/json")
echo "$RESP" | grep -q '"is_public":true' && echo "  ✅ Carte publiée" || echo "  ❌ Échec publication"

TEMPLATE_ID=$(echo "$RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)

# 2. Vérifier le statut de publication
echo "2. Vérification du statut..."
RESP=$(curl -s -X GET "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/publish-status" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN")
echo "$RESP" | grep -q '"published":true' && echo "  ✅ Statut = publiée" || echo "  ❌ Statut incorrect"

# 3. Tentative de double publication → 409
echo "3. Double publication..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/publish" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN" \
  -H "Content-Type: application/json")
[ "$HTTP_CODE" = "409" ] && echo "  ✅ 409 Conflict reçu" || echo "  ❌ Attendu 409, reçu $HTTP_CODE"

# 4. Créer une zone → template mis à jour automatiquement
echo "4. Création de zone après publication..."
ZONE_RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/zones" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN" \
  -H "Content-Type: application/json" \
  -d '{"polygons":[{"points":[{"x":400,"y":400},{"x":500,"y":500},{"x":600,"y":400}]}],"name":"new-zone"}')
echo "$ZONE_RESP" | grep -q '"id"' && echo "  ✅ Nouvelle zone créée" || echo "  ❌ Échec création zone"

# Vérifier que le template a été mis à jour
echo "   Vérification mise à jour template..."
TEMPLATE_CHECK=$(curl -s -X GET "$BASE_URL/api/v1/teams/$TEAM_ID/maps/$MAP_ID/publish-status" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN")
echo "$TEMPLATE_CHECK" | grep -q '"is_public":true' && echo "  ✅ Template toujours public" || echo "  ❌ Template perdu"

# 5. Rendre la carte publique par un non-capitaine → 403
echo "5. Publication par non-capitaine..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/teams/$TEAM2_ID/maps/$MAP_ID/publish" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN2" \
  -H "Content-Type: application/json")
echo "   Code reçu : $HTTP_CODE (attendu 201, peut dépendre du rôle)"

# 6. Nettoyage (admin uniquement)
echo "6. Nettoyage..."
TOKEN_ADMIN="<token_admin>"
curl -s -X DELETE "$BASE_URL/api/v1/admin/templates/$TEMPLATE_ID" \
  -H "Authorization: Bearer $TOKEN_ADMIN" > /dev/null 2>&1 || true

echo ""
echo "=== Résultat : Test E2E terminé ==="
```

**Fichier à créer :**
- `scripts/test-e2e-publish.sh` — Script E2E pour le cycle de vie publication + auto-update

### Partie 4 - Documentation utilisateur

**Fichier à créer :** `docs/templates.md`

```markdown
# Système de templates

## Définition

Un **template** est une copie publique des zones d'une carte, créée par le
Capitaine ou le Cocapitaine d'une équipe. Il permet à d'autres équipes de
dupliquer ces zones pour leur propre usage.

## Publication d'une carte

1. Ouvrez l'éditeur de carte.
2. Cliquez sur le bouton **"Rendre publique"** dans la barre d'outils.
3. Confirmez l'action (attention : **irréversible**).
4. Toutes les zones de la carte deviennent publiques sous forme de template.

⚠️ **Important** : Une fois qu'une carte est rendue publique, elle ne peut
plus être rendue privée. C'est une contrainte définitive pour garantir la
stabilité des templates utilisés par d'autres équipes.

## Mise à jour automatique

Après publication, votre équipe continue de modifier ses zones normalement.
Le template public est automatiquement mis à jour à chaque modification :

- ✏️ Création de zone → ajoutée au template
- 📝 Modification de polygone → mis à jour dans le template
- 🔄 Renommage → reflété dans le template
- 🗑️ Suppression de zone → retirée du template

Les nouvelles zones que vous créez sont automatiquement incluses dans le
template si elles sont visibles.

## Duplication par d'autres équipes

Une fois le template créé, d'autres équipes peuvent le dupliquer pour
créer leurs propres zones indépendantes. Voir `docs/duplication.md`
(Story #014) pour plus de détails.

## Administration

Les administrateurs peuvent supprimer tout template via l'interface
d'administration.
```

**Fichier à créer :**
- `docs/templates.md` — Documentation utilisateur du système de templates

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/template_test.go` | Créer | Tests d'intégration publication |
| `internal/services/template_sync_test.go` | Créer | Tests d'intégration synchro auto |
| `frontend/src/components/MapEditor/PublishButton.test.tsx` | Créer | Tests frontend publication |
| `scripts/test-e2e-publish.sh` | Créer | Script E2E cycle de vie |
| `docs/templates.md` | Créer | Documentation utilisateur |

### Contraintes techniques

- **Tests backend** : Utiliser la base de test PostgreSQL existante. Voir `internal/models/setup_test.go` pour le pattern.
  - Créer les fixtures nécessaires (team, captain user, map, zones)
  - Nettoyer après chaque test avec `t.Cleanup`
- **Test synchro** : Tester via le `TemplateSyncService` directement (pas via HTTP). Mocker le repository ou utiliser une vraie base de test.
- **Script E2E** : Nécessite un serveur en fonctionnement et des tokens JWT valides. Le script est fourni à titre d'exemple et peut nécessiter des ajustements selon l'environnement de test.
- **Documentation** : En français (comme le reste de la documentation utilisateur). Format Markdown.
