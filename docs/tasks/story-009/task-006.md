# Tâche #006 - Story #009 : Tests automatisés

## Objectif

Créer les tests unitaires, d'intégration et E2E pour valider l'ensemble du système de création et modification de zones polygonales.

## Contexte

- Story #009 : `docs/stories/story-009.md` - Zone Creation and Editing with Polygons
- Dépend de : Tâche #001, #002, #003, #004, #005 (tout doit être implémenté)
- Ordre d'exécution : 5ème (dernière tâche)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Tests unitaires (Backend)

#### Fichier : `internal/models/zone_test.go`

Tester le `ZoneRepository` :

1. **Création d'une zone**
   - Données : team_id=1, map_id=1
   - Attendu : zone créée avec ID > 0, created_at et updated_at définis

2. **Création d'une zone avec polygone et nom**
   - Données : zone créée, puis AddPolygon avec 3 points, puis SetName "top"
   - Attendu : FindWithDetailsByTeamAndMap retourne la zone avec 1 polygone et le nom "top"

3. **Ajout de polygones multiples**
   - Données : une zone, 3 polygones ajoutés
   - Attendu : GetPolygons retourne 3 polygones avec order_index 0, 1, 2

4. **Suppression du dernier polygone → la zone est supprimée**
   - Données : une zone avec 1 polygone
   - Attendu : RemovePolygon → PolygonCount = 0 → la zone n'est plus dans FindByTeamAndMap

5. **Contrainte de nom unique par team+map**
   - Données : zone A nommée "top", zone B nommée "top" sur même team+map
   - Attendu : IsNameTaken = true

6. **Nom trop long**
   - Données : SetName avec nom de 300 caractères
   - Attendu : erreur `ErrZoneNameTooLong`

7. **Polygone avec < 3 points**
   - Données : AddPolygon avec 2 points
   - Attendu : erreur `ErrPolygonHasTooFewPoints`

8. **Zone introuvable**
   - Données : FindByID(99999)
   - Attendu : `ErrZoneNotFound`

#### Fichier : `internal/models/edit_lock_test.go`

Tester le `EditLockRepository` :

1. **Acquisition d'un verrou**
   - Données : team_id=1, map_id=1, editor_id=1
   - Attendu : lock créé avec expires_at > locked_at (différence de 15 min)

2. **Double acquisition sur même team+map**
   - Données : Acquire par editor 1, puis Acquire par editor 2
   - Attendu : seconde acquisition → `ErrLockAlreadyAcquired`

3. **Refresh du verrou**
   - Données : Acquire puis Refresh
   - Attendu : expires_at mis à jour (plus tard)

4. **Libération du verrou**
   - Données : Acquire puis Release, puis Get
   - Attendu : Get retourne nil

5. **Cleanup des locks expirés**
   - Données : Acquire avec expiration passée (via mock ou manipulation SQL)
   - Attendu : CleanupExpired supprime le lock expiré

### Tests d'intégration (Backend)

#### Fichier : `internal/handlers/zone_test.go`

Tester les endpoints API avec un serveur de test Gin :

1. **GET zones (liste)**
   - Préparation : insérer 2 zones avec polygones et noms en DB
   - Requête : GET `/api/v1/teams/1/maps/1/zones`
   - Attendu : HTTP 200, body contient `zones` avec 2 entrées, chaque entrée a `name` et `polygons`

2. **POST create zone (succès)**
   - Requête : POST `/api/v1/teams/1/maps/1/zones` avec `{ "polygons": [{ "points": [{x:100,y:100}, {x:200,y:200}, {x:300,y:100}] }], "name": "top" }`
   - Attendu : HTTP 201, body contient `zone` avec `id`, `name: "top"`, 1 polygone

3. **POST create zone sans nom**
   - Requête : POST avec `{ "polygons": [{ "points": [...] }] }` (sans name)
   - Attendu : HTTP 201, zone créée mais pas de nom (name = "")

4. **POST create zone nom en double**
   - Requête : POST avec name déjà existant
   - Attendu : HTTP 409

5. **DELETE zone**
   - Requête : DELETE `/api/v1/teams/1/zones/1`
   - Attendu : HTTP 200

6. **DELETE zone inexistante**
   - Requête : DELETE zone 99999
   - Attendu : HTTP 404

7. **POST add polygon à zone existante**
   - Requête : POST `/api/v1/teams/1/zones/1/polygons` avec `{ "points": [...] }`
   - Attendu : HTTP 201, polygon_id > 0

8. **DELETE remove polygon (dernier → zone supprimée)**
   - Préparation : zone avec 1 seul polygone
   - Requête : DELETE polygon
   - Attendu : HTTP 200, `{ "zone_deleted": true }`

9. **PUT set name**
   - Requête : PUT `/api/v1/teams/1/zones/1/name` avec `{ "name": "secret" }`
   - Attendu : HTTP 200

10. **POST add polygon point**
    - Requête : POST `/api/v1/teams/1/zones/1/polygons/1/points` avec `{ "point": {x:150,y:150}, "index_after": 0 }`
    - Attendu : HTTP 200, polygone maintenant avec 4 points

11. **DELETE remove polygon point (< 3 restants)**
    - Préparation : polygone à 3 points
    - Requête : DELETE `/api/v1/teams/1/zones/1/polygons/1/points/0`
    - Attendu : HTTP 400

12. **POST acquire lock**
    - Requête : POST `/api/v1/teams/1/maps/1/lock`
    - Attendu : HTTP 200, `lock.editor_id` = ID de l'éditeur

13. **POST acquire lock (déjà pris)**
    - Préparation : lock pris par un autre éditeur
    - Requête : POST lock
    - Attendu : HTTP 409

14. **PUT refresh lock**
    - Préparation : lock existant
    - Requête : PUT `/api/v1/teams/1/maps/1/lock`
    - Attendu : HTTP 200, `lock.expires_at` repoussée

15. **Modification sans lock**
    - Préparation : pas de lock ou lock expiré
    - Requête : POST polygon
    - Attendu : HTTP 423

### Tests unitaires (Frontend)

#### Fichier : `frontend/src/api/zones.test.ts`

1. **listZones** : appelle GET avec les bons paramètres, retourne la liste
2. **createZone** : appelle POST avec le bon body, retourne la zone créée
3. **deleteZone** : appelle DELETE avec le bon ID
4. **addPolygon** : appelle POST avec les bons paramètres
5. **removePolygon** : appelle DELETE avec les bons IDs
6. **setZoneName** : appelle PUT avec le bon body
7. **addPolygonPoint** : appelle POST avec le bon body
8. **removePolygonPoint** : appelle DELETE avec le bon pointIndex
9. **Erreur API** : le client rejette correctement les erreurs ApiError

#### Fichier : `frontend/src/api/locks.test.ts`

1. **acquireLock** : appelle POST, retourne le lock
2. **releaseLock** : appelle DELETE
3. **refreshLock** : appelle PUT, retourne le lock refreshé
4. **getLockStatus** : appelle GET, retourne le statut

### Test E2E (optionnel, selon la config du projet)

#### Fichier : `scripts/test-e2e.sh` (mettre à jour)

Scénario E2E :
1. Se connecter en tant qu'éditeur
2. Charger l'éditeur de carte
3. Dessiner un polygone sur la carte
4. Nommer la zone "top"
5. Vérifier que la zone apparaît dans la liste
6. Recharger la page
7. Vérifier que la zone est toujours présente
8. Supprimer la zone
9. Vérifier que la zone a disparu

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/zone_test.go` | Créer | Tests unitaires ZoneRepository |
| `internal/models/edit_lock_test.go` | Créer | Tests unitaires EditLockRepository |
| `internal/handlers/zone_test.go` | Créer | Tests d'intégration API zones |
| `frontend/src/api/zones.test.ts` | Créer | Tests API client zones |
| `frontend/src/api/locks.test.ts` | Créer | Tests API client locks |

### Contraintes techniques

- **Tests backend** : Utiliser une base de données de test PostgreSQL (via `database.DB()`). Voir le pattern dans `internal/models/setup_test.go` pour la configuration du test DB.
- **Test d'intégration Gin** : Utiliser `httptest.NewServer` ou `gin.CreateTestContext` pour tester les handlers. Voir les tests existants dans `internal/handlers/team_test.go` pour le pattern.
- **Mock API client** : Pour les tests frontend, mocker `getApiClient()` pour éviter les vrais appels réseau. Utiliser `viest.mock` si disponible.
- **Couverture** : Viser une couverture fonctionnelle de 100% pour les cas décrits dans les critères d'acceptation de la story.
- **Base de test** : Vérifier que la migration de test inclut les tables zones, zone_polygons, zone_names, edit_locks, edit_history.

### Documentation

Mettre à jour `docs/zone-editor.md` si des informations de test sont pertinentes.
