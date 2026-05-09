# Tâche #002 - Story #009 : API Endpoints Zones + Verrou d'édition (back-end)

## Objectif

Créer les endpoints REST pour la gestion complète des zones (CRUD zones, polygones, points, nom) et le système de verrouillage d'édition.

## Contexte

- Story #009 : `docs/stories/story-009.md` - Zone Creation and Editing with Polygons
- Dépend de : Tâche #001 (modèles et repositories Zone, EditLock, EditHistory)
- Nécessaire pour : Tâche #004 (intégration frontend)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - ZoneHandler

Créer `internal/handlers/zone.go` avec un `ZoneHandler` qui expose les endpoints suivants :

**Endpoints Zones :**

1. **`ListZones(c)`** : GET `/api/v1/teams/:id/maps/:mapId/zones`
   - Retourne la liste de toutes les zones d'une équipe sur une carte (avec polygones et nom)
   - Middleware : `RequireAuth` + `RequireTeamRole(member, editor, cocaptain, captain)` + `RequirePermission(zone:view)`
   - Réponse : `{ "zones": [...] }` où chaque zone est au format `ZoneWithDetails`

2. **`CreateZone(c)`** : POST `/api/v1/teams/:id/maps/:mapId/zones`
   - Body : `{ "polygons": [{ "points": [{x, y}, ...] }], "name": "string optionnel" }`
   - Si name est fourni : doit être unique par team+map, max 250 chars
   - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:create)`
   - Vérifier le verrou d'édition (lock requis pour créer)
   - Réponse : 201 + `{ "zone": ZoneWithDetails }`

3. **`DeleteZone(c)`** : DELETE `/api/v1/teams/:id/zones/:zoneId`
   - Supprime une zone, ses polygones et son nom (CASCADE en DB)
   - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:delete)`
   - Vérifier le verrou d'édition
   - Réponse : 200 + `{ "message": "zone deleted" }`
   - Logguer dans edit_history avec action='delete'

4. **`GetZone(c)`** : GET `/api/v1/teams/:id/zones/:zoneId`
   - Détail d'une zone avec ses polygones et son nom
   - Middleware : `RequireAuth` + `RequireTeamRole(member, editor, cocaptain, captain)` + `RequirePermission(zone:view)`
   - Réponse : `{ "zone": ZoneWithDetails }`

**Endpoints Polygones :**

5. **`AddPolygon(c)`** : POST `/api/v1/teams/:id/zones/:zoneId/polygons`
   - Body : `{ "points": [{x, y}, ...] }`
   - Si la zone a déjà des polygones en base → sauvegarde immédiate
   - Si la zone n'a pas encore de polygone en base → ne sauvegarde que si la zone a aussi un nom (vérifier nom existant)
   - Points minimum : 3
   - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:edit)`
   - Vérifier le verrou d'édition
   - Réponse : 201 + `{ "polygon": ZonePolygon }`

6. **`RemovePolygon(c)`** : DELETE `/api/v1/teams/:id/zones/:zoneId/polygons/:polygonId`
   - Supprime un polygone
   - Si le polygone supprimé était le dernier → supprime la zone entière (et retourne `{ "zone_deleted": true }`)
   - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:edit)`
   - Vérifier le verrou d'édition
   - Réponse : 200 + `{ "polygon_deleted": true, "zone_deleted": true|false }`

7. **`UpdatePolygonPoints(c)`** : PUT `/api/v1/teams/:id/zones/:zoneId/polygons/:polygonId/points`
   - Body : `{ "points": [{x, y}, ...] }`
   - Met à jour les points d'un polygone existant
   - Logguer dans edit_history avec action='update' + old_data/new_data
   - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:edit)`
   - Vérifier le verrou d'édition
   - Réponse : 200 + `{ "polygon": ZonePolygon }`

8. **`AddPolygonPoint(c)`** : POST `/api/v1/teams/:id/zones/:zoneId/polygons/:polygonId/points`
   - Body : `{ "point": {x, y}, "index_after": int }`
   - Ajoute un point au polygone après l'index spécifié
   - Sauvegarde seulement si le polygone existe déjà en DB (a un ID)
   - Logguer dans edit_history avec action='add_point'
   - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:edit)`
   - Vérifier le verrou d'édition
   - Réponse : 200 + `{ "polygon": ZonePolygon }`

9. **`RemovePolygonPoint(c)`** : DELETE `/api/v1/teams/:id/zones/:zoneId/polygons/:polygonId/points/:pointIndex`
   - Supprime un point d'un polygone (par son index dans le tableau)
   - Sauvegarde seulement si le polygone existe déjà en DB
   - Si le polygone n'a plus que 2 points → erreur (minimum 3 points)
   - Logguer dans edit_history avec action='remove_point'
   - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:edit)`
   - Vérifier le verrou d'édition
   - Réponse : 200 + `{ "polygon": ZonePolygon }`

**Endpoints Nom :**

10. **`SetZoneName(c)`** : PUT `/api/v1/teams/:id/zones/:zoneId/name`
    - Body : `{ "name": "string" }`
    - Max 250 caractères, unique par team+map
    - Logguer dans edit_history avec action='rename' + old_name/new_name
    - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:edit)`
    - Vérifier le verrou d'édition
    - Réponse : 200 + `{ "name": ZoneName }`

**Endpoints Batch :**

11. **`BatchActions(c)`** : POST `/api/v1/teams/:id/maps/:mapId/zones/batch`
    - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:edit)`
    - Vérifier le verrou d'édition (une seule vérification pour tout le batch)
    - Traite un tableau d'actions dans une **transaction unique**. Si une action échoue, tout le batch est annulé (rollback).
    - Body :
      ```json
      {
        "actions": [
          { "type": "create_zone", "temp_id": "zone-abc", "polygons": [...], "name": "top" },
          { "type": "add_polygon", "zone_id": 5, "points": [...] },
          { "type": "add_point", "polygon_id": 10, "point": {x:1,y:2}, "index_after": 0 },
          { "type": "remove_point", "polygon_id": 10, "point_index": 2 },
          { "type": "remove_polygon", "polygon_id": 12 },
          { "type": "rename", "zone_id": 5, "name": "secret" },
          { "type": "delete_zone", "zone_id": 3 },
          { "type": "update_points", "polygon_id": 10, "points": [...] }
        ]
      }
      ```
    - Réponse :
      ```json
      {
        "results": [
          { "temp_id": "zone-abc", "zone_id": 10, "polygons": [...], "name": "top" },
          { "polygon": { "id": 15, "zone_id": 5, "points": [...], "order_index": 2 } },
          { "polygon": { "id": 15, "zone_id": 5, "points": [...] } },
          { "polygon": { "id": 15, "zone_id": 5, "points": [...] } },
          { "zone_deleted": true, "deleted_polygon_id": 12 },
          { "name": { "id": 3, "zone_id": 5, "name": "secret" } },
          { "deleted": true },
          { "polygon": { "id": 10, "points": [...] } }
        ]
      }
      ```
    - Les `temp_id` permettent au frontend de faire le mapping entre les IDs temporaires clients et les IDs réels serveur.
    - Actions supportées : `create_zone`, `add_polygon`, `remove_polygon`, `update_points`, `add_point`, `remove_point`, `rename`, `delete_zone`.
    - Logger chaque action dans edit_history individuellement.

#### Partie 2 - EditLockHandler

Créer `internal/handlers/edit_lock.go` avec un `EditLockHandler` :

11. **`AcquireLock(c)`** : POST `/api/v1/teams/:id/maps/:mapId/lock`
    - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(zone:edit)`
    - Si lock déjà pris par un autre → 409 Conflict avec `{ "error": "lock_already_acquired", "locked_by": editorID }`
    - Si lock déjà pris par le même éditeur → refresh + 200 OK
    - Si lock expiré → ré-acquisition automatique
    - Réponse : 200 + `{ "lock": EditLock }`

12. **`ReleaseLock(c)`** : DELETE `/api/v1/teams/:id/maps/:mapId/lock`
    - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)`
    - Seul le propriétaire du lock peut le libérer
    - Réponse : 200 + `{ "message": "lock released" }`

13. **`RefreshLock(c)`** : PUT `/api/v1/teams/:id/maps/:mapId/lock`
    - Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)`
    - Prolonge l'expiration de 15 minutes
    - Réponse : 200 + `{ "lock": EditLock }`

14. **`GetLockStatus(c)`** : GET `/api/v1/teams/:id/maps/:mapId/lock`
    - Middleware : `RequireAuth` + `RequireTeamRole(member, editor, cocaptain, captain)`
    - Retourne le statut du lock (verrouillé ou non, par qui, expiration)
    - Réponse : 200 + `{ "locked": bool, "lock": EditLock|null }`

#### Partie 3 - Routes

Ajouter les routes dans `api/routes.go` dans un groupe protégé par équipe.

**Cas nominaux :**
- Création d'une zone avec polygone + nom → zone créée, réponse 201
- Liste des zones d'une équipe sur une carte → toutes les zones avec polygones et noms
- Ajout d'un polygone à une zone existante → polygone sauvegardé
- Suppression du dernier polygone → zone supprimée
- Acquisition d'un lock → lock créé, éditeur peut modifier
- Refresh du lock → expiration repoussée
- Libération du lock → lock supprimé

**Cas limites :**
- Création d'une zone avec nom déjà pris → 409 Conflict
- Création d'un polygone avec < 3 points → 400 Bad Request
- Suppression d'un point qui laisse < 3 points → 400 Bad Request
- Lock déjà pris par un autre éditeur → 409 Conflict
- Tentative de modification sans lock → 423 Locked
- Carte obsolète → 400 Bad Request (vérifier via MapRepository.IsMapObsolete)

**Gestion d'erreurs :**
- 400 : validation échouée (nom vide, nom trop long, points insuffisants)
- 404 : zone/polygone/lock introuvable
- 409 : conflit (nom déjà pris, lock déjà acquis)
- 423 : ressource verrouillée par un autre éditeur
- 403 : permission insuffisante (géré par middleware)
- 500 : erreur interne

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/zone.go` | Créer | ZoneHandler : tous les endpoints zones/polygones/points/nom |
| `internal/handlers/edit_lock.go` | Créer | EditLockHandler : endpoints de verrouillage |
| `api/routes.go` | Modifier | Ajouter les routes zones et locks |
| `internal/handlers/zone_test.go` | Créer | Tests handler (préparé pour Tâche #006) |

### Signatures

```go
// internal/handlers/zone.go
package handlers

type ZoneHandler struct {
	zoneRepo      models.ZoneRepository
	mapRepo       models.MapRepository
	lockRepo      models.EditLockRepository
	historyRepo   models.EditHistoryRepository
	teamMemberRepo models.TeamMemberRepository
}

func NewZoneHandler(
	zoneRepo models.ZoneRepository,
	mapRepo models.MapRepository,
	lockRepo models.EditLockRepository,
	historyRepo models.EditHistoryRepository,
	teamMemberRepo models.TeamMemberRepository,
) *ZoneHandler

func (h *ZoneHandler) ListZones(c *gin.Context)
func (h *ZoneHandler) CreateZone(c *gin.Context)
func (h *ZoneHandler) GetZone(c *gin.Context)
func (h *ZoneHandler) DeleteZone(c *gin.Context)
func (h *ZoneHandler) AddPolygon(c *gin.Context)
func (h *ZoneHandler) RemovePolygon(c *gin.Context)
func (h *ZoneHandler) UpdatePolygonPoints(c *gin.Context)
func (h *ZoneHandler) AddPolygonPoint(c *gin.Context)
func (h *ZoneHandler) RemovePolygonPoint(c *gin.Context)
func (h *ZoneHandler) SetZoneName(c *gin.Context)
func (h *ZoneHandler) BatchActions(c *gin.Context)
```

```go
// internal/handlers/edit_lock.go
package handlers

type EditLockHandler struct {
	lockRepo models.EditLockRepository
}

func NewEditLockHandler(lockRepo models.EditLockRepository) *EditLockHandler

func (h *EditLockHandler) AcquireLock(c *gin.Context)
func (h *EditLockHandler) ReleaseLock(c *gin.Context)
func (h *EditLockHandler) RefreshLock(c *gin.Context)
func (h *EditLockHandler) GetLockStatus(c *gin.Context)
```

```go
// Routes à ajouter dans api/routes.go (dans setupTeamRoutes)
// Groupe zones et locks
teamsID := rg.Group("/teams/:id")
teamsID.Use(middleware.RequireTeamRole(teamMemberRepo, "member", "editor", "cocaptain", "captain"))
{
    // Locks
    teamsID.POST("/maps/:mapId/lock", ...)
    teamsID.DELETE("/maps/:mapId/lock", ...)
    teamsID.PUT("/maps/:mapId/lock", ...)
    teamsID.GET("/maps/:mapId/lock", ...)

    // Zones
    teamsID.GET("/maps/:mapId/zones", ...)
    teamsID.POST("/maps/:mapId/zones", ...)
    teamsID.POST("/maps/:mapId/zones/batch", ...)  // batch d'actions

    // Zone detail (sous-groupe zone)
    zones := teamsID.Group("/zones/:zoneId")
    zones.GET("", ...)
    zones.DELETE("", ...)
    zones.PUT("/name", ...)
    zones.POST("/polygons", ...)
    zones.DELETE("/polygons/:polygonId", ...)
    zones.PUT("/polygons/:polygonId/points", ...)
    zones.POST("/polygons/:polygonId/points", ...)
    zones.DELETE("/polygons/:polygonId/points/:pointIndex", ...)
}
```

### Contraintes techniques

- **Handler pattern** : Suivre exactement le pattern des handlers existants (MapHandler, TeamHandler). Constructeur public `New*Handler`, méthodes sur le handler avec signature `func (h *Handler) Method(c *gin.Context)`.
- **Middleware** : Utiliser `middleware.RequireAuth`, `middleware.RequireTeamRole` et `middleware.RequirePermission` existants. Vérifier les permissions spécifiques (ex: `"zone:create"`, `"zone:edit"`, `"zone:delete"`, `"zone:view"`).
  - Note : Il faudra peut-être ajouter ces features dans la table `permissions` (via migration ou seed). Vérifier la migration `000005_seed_permissions.up.sql` pour voir les permissions existantes.
- **Vérification lock** : Avant chaque modification (create, delete, add/remove polygon, add/remove point, rename), vérifier `lockRepo.IsLockedByEditor(teamID, mapID, editorID)`. Si pas locké par l'éditeur, retourner 423 Locked.
  - Exception : l'acquisition du lock elle-même n'a pas besoin de lock préalable.
- **Validation nom** : `SetZoneName` doit vérifier longueur <= 250, non vide, et unicité via `zoneRepo.IsNameTaken`.
- **Migration permissions** : Si les permissions `zone:create`, `zone:edit`, `zone:delete`, `zone:view` n'existent pas, ajouter une migration `000012_add_zone_permissions` qui les crée et les assigne aux rôles concernés (editor, cocaptain, captain, member pour `zone:view`).
- **Vérification carte obsolète** : Dans `CreateZone`, vérifier avec `mapRepo.IsMapObsolete(mapID)`. Si obsolète, retourner 400.
- **Gestion du dernier polygone** : Dans `RemovePolygon`, après suppression, vérifier `PolygonCount`. Si 0, supprimer la zone.
- **Format erreur** : Utiliser le même format que les handlers existants : `{"error": "error_code", "message": "human readable"}`.
- **Logging historique** : Logger les actions `create`, `delete`, `add_polygon`, `remove_polygon`, `add_point`, `remove_point`, `update`, `rename` dans edit_history avec les données avant/après quand approprié.
- **Dépendances routes** : Créer les repositories Zone, EditLock, EditHistory dans `setupTeamRoutes` et les injecter dans les handlers.

### Tests à implémenter

Les tests seront créés dans la Tâche #006.

### Documentation

Aucune documentation à créer pour cette tâche.
