# Tâche #009 - Story #033 : Tests automatisés

## Objectif
Couvrir tous les critères d'acceptation de la story avec des tests unitaires (Go + frontend), d'intégration (Go) et E2E (Playwright).

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Toutes les tâches #001 à #008
- Nécessaire pour : Aucune (tâche finale)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Tests à implémenter par couche :**

#### Tests unitaires Go — `internal/models/line_of_fire_test.go`

Build tag : `//go:build integration` (base de données réelle)

| Scénario | Description | Données | Résultat attendu |
|----------|-------------|---------|------------------|
| Création ligne avec 1 cône | Créer une ligne de tir avec 1 cône valide | name="couloir A", 1 cône avec 3 points distincts | Ligne créée avec ID > 0, polygon_data contient le cône |
| Création sans nom | name="" | | ErrLineOfFireNameRequired |
| Création nom trop long | name de 251 caractères | | ErrLineOfFireNameTooLong |
| Création nom en doublon | Même équipe, même carte, même nom | | ErrLineOfFireNameTaken |
| Création ligne sans cône | polygon_data.cones vide | | Erreur (validation échoue) |
| FindByID inexistant | ID 99999 | | ErrLineOfFireNotFound |
| FindByTeamAndMap | 2 lignes sur la même carte | | 2 résultats, ordre croissant par order_index |
| UpdateName | Renommer une ligne | | Nom mis à jour |
| UpdatePolygonData | Modifier les cônes | | polygon_data mis à jour |
| Delete | Supprimer une ligne | | Ligne supprimée, FindByID retourne not found |
| Delete supprime versions | Vérifier CASCADE sur line_of_fire_versions | | Versions supprimées |

#### Tests d'intégration Go — `internal/handlers/line_of_fire_test.go`

Build tag : `//go:build integration`

| Scénario | Endpoint | Résultat attendu |
|----------|----------|------------------|
| Liste lignes de tir | GET /teams/:id/maps/:mapId/lines-of-fire | 200, liste des lignes |
| Création ligne | POST /teams/:id/maps/:mapId/lines-of-fire | 201, ligne créée |
| Création sans verrou | POST avec lock non acquis | 423 Locked |
| Création carte obsolète | POST sur carte obsolète | 400 |
| Get ligne | GET /teams/:id/lines-of-fire/:id | 200, ligne retournée |
| Get ligne inexistante | GET /teams/:id/lines-of-fire/99999 | 404 |
| Delete ligne | DELETE /teams/:id/lines-of-fire/:id | 200 |
| Renommer | PUT /teams/:id/lines-of-fire/:id/name | 200 |
| Ajouter cône | POST /teams/:id/lines-of-fire/:id/cones | 201, cône ajouté |
| Modifier cône | PUT /teams/:id/lines-of-fire/:id/cones/0 | 200, cône modifié |
| Supprimer cône | DELETE /teams/:id/lines-of-fire/:id/cones/0 | 200 |
| Dernier cône supprimé → ligne supprimée | DELETE sur cône unique | 200, ligne aussi supprimée |
| Batch create_line_of_fire | POST batch | 200, ligne créée |
| Révision filter=lines | GET /teams/:id/maps/:mapId/revision-zones?filter=lines | 200, seulement les lignes |
| Révision filter=both | GET /teams/:id/maps/:mapId/revision-zones?filter=both | 200, zones + lignes |

#### Tests E2E — `frontend/e2e/lines-of-fire/`

| Fichier | Scénario | Description |
|---------|----------|-------------|
| `lines-of-fire/creation.spec.ts` | Création d'une ligne de tir | Clic droit → "Créer une ligne de tir", placer 3 points, saisir un nom |
| `lines-of-fire/edition.spec.ts` | Modification d'une ligne de tir | Déplacer les points d'extrémité, renommer |
| `lines-of-fire/deletion.spec.ts` | Suppression d'une ligne | Supprimer via clic droit |
| `lines-of-fire/cone.spec.ts` | Double-clic sur origine | Ajouter un cône, vérifier qu'il apparaît |
| `lines-of-fire/straight-line.spec.ts` | Segment de droite | left_end = right_end, vérifier l'affichage |
| `lines-of-fire/revision.spec.ts` | Révision avec filtre | Lancer révision "Lignes uniquement" et "Les deux" |

Chaque test E2E doit :
- Utiliser les fixtures existantes (`page`, `adminUser`, `editorUser`)
- Se connecter en tant qu'éditeur
- Naviguer vers l'éditeur de carte
- Exécuter le scénario
- Vérifier le résultat

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/line_of_fire_test.go` | Créer | Tests unitaires repository |
| `internal/handlers/line_of_fire_test.go` | Créer | Tests d'intégration handler |
| `frontend/e2e/lines-of-fire/creation.spec.ts` | Créer | E2E création |
| `frontend/e2e/lines-of-fire/edition.spec.ts` | Créer | E2E édition |
| `frontend/e2e/lines-of-fire/deletion.spec.ts` | Créer | E2E suppression |
| `frontend/e2e/lines-of-fire/cone.spec.ts` | Créer | E2E ajout cône |
| `frontend/e2e/lines-of-fire/straight-line.spec.ts` | Créer | E2E segment droit |
| `frontend/e2e/lines-of-fire/revision.spec.ts` | Créer | E2E révision filtrée |

### Contraintes techniques
- **Tests Go** : Utiliser le `setupTestDB()` existant avec testcontainers.
- **Tests E2E** : Utiliser le pattern des tests existants (fixtures, helpers dans `frontend/e2e/helpers/`).
- **Nouveau helper DB** : Ajouter `seedLineOfFire(teamId, mapId, name, coneData)` dans `frontend/e2e/helpers/db.ts` si nécessaire.
- **Exécution** : `go test --tags=integration ./internal/models/ ./internal/handlers/` et `cd frontend && npx vitest run` et `cd frontend && npx playwright test`.
