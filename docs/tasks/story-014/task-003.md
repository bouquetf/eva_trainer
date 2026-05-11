# Tâche #003 - Story #014 : Tests, E2E et documentation

## Objectif

Créer les tests d'intégration backend pour l'API d'import de template, les tests unitaires frontend, le script E2E, et mettre à jour la documentation.

## Contexte

- Story #014 : `docs/stories/story-014.md` - Template Duplication by Other Teams
- Dépend de : Tâche #001 (API import template avec remplacement), Tâche #002 (frontend)
- Ordre d'exécution : 3ème (dernière)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Partie 1 - Tests d'intégration backend : Import de template

**Fichier :** `internal/handlers/template_test.go` (ajouter dans le fichier existant)

Utiliser `setupTemplateTest` pour créer les fixtures de base. Adapter ou créer des helpers pour avoir deux équipes :

1. **Import d'un template avec zones par un capitaine de l'équipe cible — remplacement des zones existantes**
   - Préparation :
     - Équipe A (originale) : 1 capitaine, 1 carte, 2 zones avec polygones et noms ("Zone A1", "Zone A2")
     - Publier la carte (CreateFromMap)
     - Équipe B (cible) : 1 capitaine, la MÊME carte, 1 zone existante ("Zone B existante")
   - Requête : POST `/api/v1/teams/{teamB}/duplicate-template/{templateId}` (en tant que capitaine de B)
   - Vérification : HTTP 201
   - Vérification : `zones` contient 2 zones
   - Vérification : Les zones ont les mêmes noms que les zones du template ("Zone A1", "Zone A2")
   - Vérification : La zone "Zone B existante" a été supprimée
   - Vérification : `zoneRepo.FindWithDetailsByTeamAndMap(teamB, mapID)` retourne 2 zones (pas 3)
   - Vérification : Le template original n'a pas été modifié

2. **Import d'un template vide (0 zones) — supprime les zones existantes**
   - Préparation : Template créé sans zones (carte vide publiée), équipe B a 2 zones existantes
   - Requête : POST duplicate-template
   - Vérification : HTTP 201, `zones` = [], équipe B n'a plus de zones sur cette carte

3. **Import par un membre non-capitaine → 403**
   - Préparation : Template existant, utilisateur membre de l'équipe B (pas captain/cocaptain)
   - Requête : POST duplicate-template (via middleware RequireTeamRole avec editor → doit être bloqué avant)
   - Vérification : HTTP 403

4. **Import d'un template inexistant → 404**
   - Préparation : Template ID invalide (99999)
   - Requête : POST duplicate-template
   - Vérification : HTTP 404

5. **Import d'un template non public → 400**
   - Préparation : Template créé avec is_public=false (via UPDATE SQL direct)
   - Requête : POST duplicate-template
   - Vérification : HTTP 400, erreur template_not_public

Ajouter les tests pour `ListPublicTemplates` :

6. **Liste des templates publics avec map_name**
   - Préparation : 2 templates publics sur des cartes différentes
   - Requête : GET `/api/v1/templates/public`
   - Vérification : HTTP 200, `templates` contient 2 entrées
   - Vérification : Chaque entrée contient `template.map_name` (non vide)

7. **Liste vide quand aucun template public**
   - Requête : GET `/api/v1/templates/public`
   - Vérification : HTTP 200, `templates` = []

**Test de la méthode `DuplicateToTeam` (intégration repository) :**

**Fichier :** `internal/models/template_test.go` (créer)

8. **`DuplicateToTeam` copie correctement les polygones et noms, supprime les anciennes zones**
   - Préparation :
     - Équipe A : 2 zones avec polygones et noms sur map M
     - Template créé via CreateFromMap
     - Équipe B : 1 zone existante sur map M
   - Action : `templateRepo.DuplicateToTeam(templateID, teamB.ID)`
   - Vérification : Retourne mapID = M.ID
   - Vérification : `zoneRepo.FindWithDetailsByTeamAndMap(teamB.ID, M.ID)` retourne 2 zones
   - Vérification : Les noms et polygones correspondent au template (pas à l'ancienne zone)
   - Vérification : `zoneRepo.FindWithDetailsByTeamAndMap(teamA.ID, M.ID)` retourne toujours 2 zones (inchangé)

9. **`DuplicateToTeam` sur template non public → erreur**
   - Préparation : Template créé puis SET is_public = false
   - Action : `templateRepo.DuplicateToTeam(templateID, teamID)`
   - Vérification : Erreur `ErrTemplateNotPublic`

10. **`DuplicateToTeam` sur template inexistant → `ErrTemplateNotFound`**
    - Action : `templateRepo.DuplicateToTeam(99999, teamID)`
    - Vérification : Erreur `ErrTemplateNotFound`

### Partie 2 - Tests d'intégration frontend

**Fichier :** `frontend/src/pages/ImportTemplate.test.tsx`

(Compléter les tests unitaires décrits dans la Tâche #002)

### Partie 3 - Tests E2E

**Fichier :** `scripts/test-e2e-import-template.sh` (créer)

```bash
#!/bin/bash
# Test E2E : Template Import by Captain/Cocaptain

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN_CAPTAIN_A="<token_captain_teamA>"
TOKEN_CAPTAIN_B="<token_captain_teamB>"
TOKEN_MEMBER_B="<token_member_teamB>"
TEAM_A_ID=1
TEAM_B_ID=2
MAP_ID=1

echo "=== Test E2E : Import de template ==="

# 1. Équipe A crée des zones et publie
echo "1. Équipe A crée des zones et publie..."
curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_A_ID/maps/$MAP_ID/zones" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN_A" \
  -H "Content-Type: application/json" \
  -d '{"polygons":[{"points":[{"x":0,"y":0},{"x":10,"y":0},{"x":5,"y":10}]}],"name":"Zone Publique A"}' > /dev/null

RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_A_ID/maps/$MAP_ID/publish" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN_A" \
  -H "Content-Type: application/json")
TEMPLATE_ID=$(echo "$RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
echo "  ✅ Template créé (id=$TEMPLATE_ID)"

# 2. Équipe B crée une zone existante sur la même carte
echo "2. Équipe B a une zone existante..."
curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_B_ID/maps/$MAP_ID/zones" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN_B" \
  -H "Content-Type: application/json" \
  -d '{"polygons":[{"points":[{"x":50,"y":50},{"x":60,"y":50},{"x":55,"y":60}]}],"name":"Zone Existante B"}' > /dev/null
echo "  ✅ Zone existante créée"

# 3. Équipe B liste les templates publics
echo "3. Équipe B liste les templates publics..."
RESP=$(curl -s -X GET "$BASE_URL/api/v1/templates/public" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN_B")
echo "$RESP" | grep -q '"map_name"' && echo "  ✅ Templates listés avec map_name" || echo "  ❌ map_name manquant"

# 4. Capitaine B importe le template (remplacement)
echo "4. Capitaine B importe le template (remplacement)..."
IMPORT_RESP=$(curl -s -X POST "$BASE_URL/api/v1/teams/$TEAM_B_ID/duplicate-template/$TEMPLATE_ID" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN_B" \
  -H "Content-Type: application/json")
echo "$IMPORT_RESP" | grep -q '"zones"' && echo "  ✅ Template importé" || echo "  ❌ Échec import"

# 5. Vérifier que la zone existante a été remplacée
echo "5. Vérification remplacement..."
ZONES_RESP=$(curl -s -X GET "$BASE_URL/api/v1/teams/$TEAM_B_ID/maps/$MAP_ID/zones" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN_B")
ZONE_COUNT=$(echo "$ZONES_RESP" | grep -o '"id":' | wc -l)
echo "  Nombre de zones pour équipe B : $ZONE_COUNT"
[ "$ZONE_COUNT" -ge 1 ] && echo "  ✅ Zones remplacées (plus de 'Zone Existante B')" || echo "  ❌ Pas de zones"

# 6. Vérifier que le template original est inchangé
echo "6. Vérification template original inchangé..."
TEMPLATE_CHECK=$(curl -s -X GET "$BASE_URL/api/v1/teams/$TEAM_A_ID/maps/$MAP_ID/publish-status" \
  -H "Authorization: Bearer $TOKEN_CAPTAIN_A")
echo "$TEMPLATE_CHECK" | grep -q '"is_public":true' && echo "  ✅ Template original intact" || echo "  ❌ Template original modifié"

# 7. Membre B tente d'importer → 403
echo "7. Membre B tente d'importer (doit échouer)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/teams/$TEAM_B_ID/duplicate-template/$TEMPLATE_ID" \
  -H "Authorization: Bearer $TOKEN_MEMBER_B" \
  -H "Content-Type: application/json")
[ "$HTTP_CODE" = "403" ] && echo "  ✅ 403 Forbidden (membre non autorisé)" || echo "  ❌ Attendu 403, reçu $HTTP_CODE"

echo ""
echo "=== Résultat : Test E2E terminé ==="
```

### Partie 4 - Documentation utilisateur

**Fichier à modifier :** `docs/templates.md`

Remplacer la section "Duplication par d'autres équipes" existante par :

```markdown
## Import d'un template par une autre équipe

Une fois qu'un template est créé (carte rendue publique), les capitaines et
co-capitaines d'autres équipes peuvent importer ce template pour remplacer
les zones de leur équipe sur la même carte.

### Comment importer un template

1. Depuis la page d'accueil de votre équipe, cliquez sur
   **"Importer un template"** (réservé aux capitaines et co-capitaines).
2. Parcourez la liste des templates publics disponibles.
3. Sélectionnez le template souhaité et cliquez sur **"Importer dans ma carte"**.
4. Un message de confirmation vous avertit que **toutes les zones existantes
   de votre équipe pour cette carte seront remplacées**.
5. Confirmez l'action.
6. Les zones du template sont copiées dans votre équipe et vous êtes redirigé
   vers l'éditeur de carte.

### Ce qui est copié

- ✅ Les polygones (formes) de chaque zone
- ✅ Les noms des zones
- ❌ L'historique des modifications (non copié)
- ❌ La référence à l'équipe d'origine (non conservée)

⚠️ **Important** : L'import **remplace** toutes les zones existantes de votre
équipe pour cette carte. Les anciennes zones sont définitivement perdues.

### Indépendance

Une fois importées, les zones vous appartiennent totalement :

- ✏️ Vous pouvez modifier les polygones
- 🔄 Vous pouvez renommer les zones
- 🗑️ Vous pouvez supprimer des zones
- Vos modifications n'affectent pas le template original
- Les modifications du template original n'affectent pas votre copie

### Confidentialité

L'équipe qui a créé le template original ne peut pas savoir que vous
l'avez importé. Aucune information d'import n'est conservée.
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/template_test.go` | Modifier | Ajouter tests `DuplicateTemplate` (remplacement) et `ListPublicTemplates` |
| `internal/models/template_test.go` | Créer | Tests unitaires `DuplicateToTeam` |
| `frontend/src/pages/ImportTemplate.test.tsx` | Créer | Tests frontend import |
| `scripts/test-e2e-import-template.sh` | Créer | Script E2E cycle de vie import |
| `docs/templates.md` | Modifier | Section import template |

### Contraintes techniques

- **Tests backend** : Utiliser le pattern existant (`setupTemplateTestDB`, `setupTemplateTest`). Pour les tests avec deux équipes, créer un helper `setupDuplicateTest` qui configure deux équipes complètes.
- **Helper de test** : Créer `createPublicTemplate(t, ...)` qui crée une équipe + carte + zones + publie, pour réutilisation.
- **Vérification remplacement** : Dans les tests, vérifier que `SELECT COUNT(*) FROM zones WHERE team_id = $1 AND map_id = $2` retourne le bon nombre après import (pas l'ancien + le nouveau).
- **Script E2E** : Nécessite serveur en fonctionnement et tokens JWT valides.
- **Documentation** : En français. Format Markdown.
