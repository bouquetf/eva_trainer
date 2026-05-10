# Tâche #005 - Story #012 : Tests automatisés

## Objectif

Créer les tests unitaires, d'intégration et E2E pour valider l'ensemble du système de versions de zone, navigation et restauration.

## Contexte

- Story #012 : `docs/stories/story-012.md` - Versions de zone, historique et rollback
- Dépend de : Tâche #001, #002, #003, #004 (tout doit être implémenté)
- Ordre d'exécution : 5ème (dernière tâche)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Tests unitaires (Backend)

#### Fichier : `internal/models/zone_version_test.go`

Tester le `ZoneVersionRepository` :

1. **Création d'un snapshot**
   - Préparation : créer une zone avec un polygone et un nom
   - Appel : `Save(zoneID, &editorID, snapshot)` avec un snapshot contenant 1 polygone et le nom "top"
   - Attendu : version créée, ID > 0, version_data correspond au snapshot

2. **Création d'un snapshot sans nom**
   - Préparation : zone sans nom
   - Appel : `Save` avec snapshot sans nom
   - Attendu : version créée, version_data.name = ""

3. **FindByMap — pagination**
   - Préparation : insérer 25 versions pour 3 zones différentes sur la même carte
   - Appel : `FindByMap(teamID, mapID, 1, 10)`
   - Attendu : 10 versions, total = 25

4. **FindByMap — page 2**
   - Appel : `FindByMap(teamID, mapID, 2, 10)`
   - Attendu : 10 versions (les 10 suivantes), total = 25

5. **FindByZone**
   - Préparation : insérer 5 versions pour une zone spécifique, 3 pour une autre
   - Appel : `FindByZone(zoneID, 1, 20)`
   - Attendu : 5 versions, tri DESC

6. **FindFirstOtherEditorVersion — pas d'autre éditeur**
   - Préparation : toutes les versions créées par le même éditeur
   - Appel : `FindFirstOtherEditorVersion(zoneID, editorID)`
   - Attendu : nil

7. **FindFirstOtherEditorVersion — avec autre éditeur**
   - Préparation : 3 versions de editor A (les plus récentes), 2 versions de editor B (plus anciennes)
   - Appel : `FindFirstOtherEditorVersion(zoneID, editorA)`
   - Attendu : l'ID de la version la plus récente de editor B (la limite)

8. **FindByID — introuvable**
   - Appel : `FindByID(99999)`
   - Attendu : ErrVersionNotFound

#### Fichier : `internal/models/edit_history_test.go`

Tester les nouvelles méthodes de `EditHistoryRepository` :

1. **FindByTeamAndMapWithFilters — pagination**
   - Préparation : insérer 15 entrées d'historique
   - Appel : `FindByTeamAndMapWithFilters(teamID, mapID, 1, 10)`
   - Attendu : 10 entrées, total = 15, tri DESC

2. **FindByZoneWithFilters**
   - Préparation : 5 entrées pour zone A, 3 pour zone B
   - Appel : `FindByZoneWithFilters(zoneA, 1, 20)`
   - Attendu : 5 entrées

### Tests d'intégration (Backend)

#### Fichier : `internal/handlers/history_test.go`

Tester les endpoints API avec un serveur de test Gin :

1. **GET versions d'une carte (succès)**
   - Préparation : insérer des versions pour plusieurs zones via zone_version repo
   - Requête : GET `/api/v1/teams/1/maps/1/versions`
   - Attendu : HTTP 200, body contient `versions` avec les bonnes données

2. **GET versions d'une zone spécifique**
   - Préparation : versions pour zone 5 et zone 6
   - Requête : GET `/api/v1/teams/1/zones/5/versions`
   - Attendu : HTTP 200, seules les versions de zone 5

3. **GET versions (éditeur filtré)**
   - Préparation : l'utilisateur a le rôle 'editor', versions de editor A et B
   - Requête : GET zone versions
   - Attendu : le backend applique déjà le filtre FindFirstOtherEditorVersion

4. **POST restore version (éditeur, sa version, pas modifiée depuis)**
   - Préparation : zone créée par editor A, version sauvegardée par editor A, lock acquis
   - Requête : POST `/api/v1/teams/1/zones/5/versions/1/restore`
   - Attendu : HTTP 200, zone restaurée avec les données du snapshot

5. **POST restore version (éditeur, version d'un autre)**
   - Préparation : version créée par editor B
   - Requête : POST restore
   - Attendu : HTTP 403

6. **POST restore version (éditeur, zone modifiée par un autre depuis)**
   - Préparation : editor A crée une version, editor B modifie la zone
   - Requête : POST restore par editor A
   - Attendu : HTTP 403

7. **POST restore version (capitaine, n'importe quelle version)**
   - Préparation : version créée par editor A, lock acquis par le capitaine
   - Requête : POST restore par le capitaine
   - Attendu : HTTP 200

8. **POST restore version (sans lock)**
   - Préparation : pas de lock
   - Requête : POST restore
   - Attendu : HTTP 423

9. **POST restore version (version introuvable)**
   - Requête : POST restore avec versionId=99999
   - Attendu : HTTP 404

### Tests d'intégration (Snapshot auto-save)

#### Fichier : `internal/handlers/zone_test.go` (ajouter des tests)

1. **Création de zone → snapshot créé**
   - Requête : POST create zone avec polygone et nom
   - Attendu : HTTP 201 + une entrée zone_versions créée

2. **Batch actions → snapshots créés pour chaque zone modifiée**
   - Requête : POST batch avec 2 actions sur 2 zones différentes
   - Attendu : 2 snapshots créés (un par zone)

### Tests unitaires (Frontend)

#### Fichier : `frontend/src/api/versions.test.ts`

1. **listMapVersions (succès)**
   - Mock : API retourne `{ versions: [...], total: 5, page: 1, limit: 20 }`
   - Résultat attendu : VersionPage avec les bonnes valeurs

2. **listMapVersions avec filtres**
   - Vérifier que les query params sont envoyés correctement

3. **listZoneVersions**
   - Vérifier l'URL appelée contient le bon zoneId

4. **restoreVersion**
   - Vérifier que POST est appelé avec le bon versionId

5. **Erreur API 403**
   - ApiError bien rejetée

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/zone_version_test.go` | Créer | Tests unitaires ZoneVersionRepository |
| `internal/models/edit_history_test.go` | Créer | Tests unitaires EditHistoryRepository (nouvelles méthodes pagination) |
| `internal/handlers/history_test.go` | Créer | Tests d'intégration API versions/restore |
| `internal/handlers/zone_test.go` | Modifier | Ajouter tests vérification snapshot auto-save |
| `frontend/src/api/versions.test.ts` | Créer | Tests API client versions |

### Contraintes techniques

- **Tests backend** : Utiliser une base de données de test PostgreSQL (via `database.DB()`). Voir le pattern dans les tests existants.
- **Test d'intégration Gin** : Utiliser `httptest.NewServer` ou le pattern existant.
- **Mock API client** : Pour les tests frontend, mocker `getApiClient()`. Utiliser `viest.mock` si disponible.
- **Base de test** : Vérifier que les migrations de test incluent `zone_versions`.
- **Seed de test** : Créer des helpers pour insérer des données de test (zones, version snapshots) en base.

### Documentation

Mettre à jour `docs/edit-mode-history.md` avec des informations sur les tests si pertinent.
