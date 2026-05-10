# Tâche #002 - Story #012 : API Endpoints Versions, Restauration et Historique (Backend)

## Objectif

Créer les endpoints REST pour la consultation des versions de zone, la restauration d'une version, et la sauvegarde automatique des snapshots lors du batch auto-save. Corriger la permission `edit_history:view` dans les seeds et migrations.

## Contexte

- Story #012 : `docs/stories/story-012.md` - Versions de zone, historique et rollback
- Dépend de : Tâche #001 (modèle ZoneVersion, ZoneVersionRepository, migration)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 0 - Correction de permission

Créer une migration `000014_fix_edit_history_permissions.up.sql` qui :
- Supprime `edit_history:view` du rôle `member` dans `role_permissions`
- Ajoute `edit_history:view` au rôle `editor` dans `role_permissions`
- Les permissions `edit_history:view` et `edit_history:rollback` existent déjà dans la table `permissions`

**Important** : Cette migration doit être EXÉCUTÉE APRÈS la migration `000013_add_zone_versions`. Le numéro de migration doit être `000014`.

#### Partie 1 - HistoryHandler

Créer `internal/handlers/history.go` avec un `HistoryHandler` qui expose les endpoints suivants :

**1. `ListVersions(c)`** : GET `/api/v1/teams/:id/maps/:mapId/versions`
- Middleware : `RequireAuth` + `RequireTeamRole(member, editor, cocaptain, captain)` + `RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "edit_history:view")`
- **Règle de visibilité :**
  - Si l'utilisateur est `captain` ou `cocaptain` → voir toutes les versions de toutes les zones
  - Si l'utilisateur est `editor` → voir toutes les versions (le filtrage par zone/éditeur sera fait frontend et via le handler spécifique zone)
  - Si l'utilisateur est `admin` → voir toutes les versions
- **Paramètres query :**
  - `page` : numéro de page (défaut: 1)
  - `limit` : éléments par page (défaut: 20, max: 100)
  - `zone_id` : filtrer par zone (optionnel)
  - `editor_id` : filtrer par éditeur (optionnel, uniquement pour captain/cocaptain/admin)
- **Réponse :**
  ```json
  {
    "versions": [
      {
        "id": 1,
        "zone_id": 5,
        "zone_name": "top",
        "editor_id": 3,
        "editor_name": "PlayerName",
        "version_data": { "polygons": [...], "name": "top" },
        "created_at": "2026-05-10T12:00:00Z"
      }
    ],
    "total": 42,
    "page": 1,
    "limit": 20
  }
  ```

**2. `ListZoneVersions(c)`** : GET `/api/v1/teams/:id/zones/:zoneId/versions`
- Middleware : `RequireAuth` + `RequireTeamRole(member, editor, cocaptain, captain)` + `RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "edit_history:view")`
- **Règle de visibilité éditeur :**
  - Si l'utilisateur est `captain` ou `cocaptain` → toutes les versions
  - Si l'utilisateur est `editor` → versions jusqu'à la première version créée par un autre éditeur
    - Appeler `versionRepo.FindFirstOtherEditorVersion(zoneID, user.ID)`
    - Si une version limite est trouvée, ne retourner que les versions avec `created_at >=` cette date
    - Le calcul est fait backend pour éviter de fuiter des données d'autres éditeurs
- **Correction** : Le `member` ne devrait pas avoir accès à `edit_history:view` (corrigé dans la migration 000014). Mais pour la sécurité, le handler doit vérifier le rôle. Si le rôle est `member` et que la permission n'est pas présente, le middleware renverra 403.

**3. `RestoreVersion(c)`** : POST `/api/v1/teams/:id/zones/:zoneId/versions/:versionId/restore`
- Middleware : `RequireAuth` + `RequireTeamRole(editor, cocaptain, captain)` + `RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "edit_history:rollback")`
  - **Note** : Pour les éditeurs sans `edit_history:rollback`, la permission est `edit_history:view` suffisante (car un éditeur peut restaurer ses propres versions). Mais le handler doit vérifier lui-même les droits.
  - En pratique : middleware `RequireTeamRole(editor, cocaptain, captain)` + middleware `RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "edit_history:view")` (les éditeurs ont cette permission après la correction).
- **Comportement :**
  1. Récupérer l'utilisateur connecté
  2. Récupérer la version demandée
  3. Récupérer la zone (pour vérifier l'équipe)
  4. **Vérification lock** : `lockRepo.IsLockedByEditor(teamID, mapID, user.ID)` → 423 si pas de lock
  5. **Si éditeur** (pas captain/cocaptain) :
     - Vérifier que l'éditeur est l'auteur de la version → 403 "vous n'êtes pas l'auteur de cette version"
     - Vérifier que la zone n'a pas été modifiée par un autre éditeur depuis la version :
       - `versionRepo.GetLatestForZone(zoneID)` → version la plus récente
       - Si la version la plus récente a un `editor_id` différent et est plus récente que la version ciblée → 403 "la zone a été modifiée par un autre éditeur depuis cette version"
  6. Remplacer les polygones de la zone par ceux du snapshot :
     - Supprimer tous les polygones existants (`RemovePolygon` pour chacun ou requête directe)
     - Ajouter les polygones du snapshot via `AddPolygon`
     - Mettre à jour le nom via `SetName`
  7. Logger `version_restore` dans edit_history (audit trail)
  8. Retourner la zone restaurée
- **Réponse :** 200 + `{ "zone": ZoneWithDetails }`

#### Partie 2 - Sauvegarde automatique des snapshots (hook dans ZoneHandler)

Modifier `internal/handlers/zone.go` :

1. Injecter le `ZoneVersionRepository` dans `ZoneHandler`
2. Après chaque batch auto-save réussi (fin de `BatchActions`), pour chaque zone modifiée :
   - Récupérer les polygones actuels : `zoneRepo.GetPolygons(zoneID)`
   - Récupérer le nom actuel : `zoneRepo.GetName(zoneID)`
   - Créer un `VersionSnapshot`
   - Appeler `versionRepo.Save(zoneID, &user.ID, snapshot)`
3. Faire de même dans les handlers individuels : `CreateZone`, `RemovePolygon` (si zone_deleted), `SetZoneName`, etc.
   - Une helper `saveZoneSnapshot(zoneID, editorID)` mutualisée pour éviter la duplication

**Helper à ajouter dans `ZoneHandler` :**
```go
func (h *ZoneHandler) saveZoneSnapshot(zoneID, editorID int) {
    if h.versionRepo == nil {
        return
    }
    polygons, err := h.zoneRepo.GetPolygons(zoneID)
    if err != nil {
        log.Printf("warning: failed to get polygons for snapshot (zone=%d): %v", zoneID, err)
        return
    }
    zoneName := ""
    zn, err := h.zoneRepo.GetName(zoneID)
    if err == nil && zn != nil {
        zoneName = zn.Name
    }

    snapshot := models.VersionSnapshot{
        Name: zoneName,
    }
    for _, p := range polygons {
        snapshot.Polygons = append(snapshot.Polygons, models.SnapshotPolygon{
            Points:     p.Points,
            OrderIndex: p.OrderIndex,
        })
    }

    var editorIDPtr *int
    editorIDPtr = &editorID
    if _, err := h.versionRepo.Save(zoneID, editorIDPtr, snapshot); err != nil {
        log.Printf("warning: failed to save zone snapshot (zone=%d): %v", zoneID, err)
    }
}
```

**Points d'appel de `saveZoneSnapshot` :**
- Fin de `BatchActions` (appeler pour chaque zone modifiée)
- Fin de `CreateZone` (après création réussie)
- Fin de `SetZoneName` (après rename)
- Fin de `AddPolygon` (après ajout)
- Fin de `RemovePolygon` (après suppression, surtout si zone_deleted — dans ce cas, prendre le snapshot AVANT la suppression)
- Fin de `UpdatePolygonPoints`
- Fin de `AddPolygonPoint`
- Fin de `RemovePolygonPoint`

#### Partie 3 - Routes

Ajouter les routes dans `api/routes.go` :

```go
// Dans setupTeamRoutes, après les routes zones existantes

// Versions d'une carte (accessible à tous les membres avec edit_history:view)
teams.GET("/:id/maps/:mapId/versions",
    middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "edit_history:view"),
    historyHandler.ListVersions,
)

// Versions d'une zone
teams.GET("/:id/zones/:zoneId/versions",
    middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "edit_history:view"),
    historyHandler.ListZoneVersions,
)

// Restauration d'une version
teams.POST("/:id/zones/:zoneId/versions/:versionId/restore",
    middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "edit_history:view"),
    historyHandler.RestoreVersion,
)
```

#### Partie 4 - Injection des dépendances

Dans `api/routes.go` → `setupTeamRoutes` :
```go
versionRepo := models.NewZoneVersionRepository(db)
// Après la création de zoneHandler...
historyHandler := handlers.NewHistoryHandler(versionRepo, zoneRepo, lockRepo, teamMemberRepo, userRepo, historyRepo)
// Modifier la création de zoneHandler pour inclure versionRepo :
zoneHandler := handlers.NewZoneHandler(zoneRepo, mapRepo, lockRepo, historyRepo, teamMemberRepo, templateSyncSvc, versionRepo)
```

Mettre à jour le constructeur de `ZoneHandler` pour accepter `*models.ZoneVersionRepository` (optionnel, peut être nil pour les tests).

Mettre à jour le constructeur de `NewHistoryHandler` :
```go
func NewHistoryHandler(
    versionRepo models.ZoneVersionRepository,
    zoneRepo models.ZoneRepository,
    lockRepo models.EditLockRepository,
    teamMemberRepo models.TeamMemberRepository,
    userRepo models.UserRepository,
    historyRepo models.EditHistoryRepository,
) *HistoryHandler
```

**Cas nominaux :**
- Consultation des versions d'une carte par le capitaine → toutes les versions
- Consultation des versions d'une zone par un éditeur → versions jusqu'à la limite (première version d'un autre)
- Restauration d'une version par l'éditeur auteur (zone non modifiée par d'autres) → succès
- Restauration d'une version par le capitaine → succès (n'importe quelle version)
- Snapshot créé automatiquement après chaque batch save

**Cas limites :**
- Éditeur consulte les versions d'une zone → la liste s'arrête à la première version d'un autre éditeur
- Éditeur tente de restaurer une version d'un autre éditeur → 403
- Éditeur tente de restaurer une version après modification par un autre → 403
- Restauration sans lock d'édition → 423
- Zone sans version → liste vide
- Snapshot d'une zone supprimée → pas de snapshot (la zone n'existe plus)
- Version introuvable → 404

**Gestion d'erreurs :**
- 400 : paramètres invalides
- 403 : pas l'auteur de la version, zone modifiée par un autre
- 404 : version introuvable, zone introuvable
- 423 : lock requis
- 500 : erreur interne

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/history.go` | Créer | HistoryHandler : ListVersions, ListZoneVersions, RestoreVersion |
| `internal/handlers/zone.go` | Modifier | Ajouter versionRepo, helper saveZoneSnapshot, appels dans BatchActions et handlers individuels |
| `api/routes.go` | Modifier | Ajouter routes versions + restore + injection dépendances |
| `api/routes.go` | Modifier | Mettre à jour constructeur ZoneHandler (ajouter versionRepo) |
| `db/migrations/000014_fix_edit_history_permissions.up.sql` | Créer | Correction permission : editor ✅, member ❌ |
| `db/migrations/000014_fix_edit_history_permissions.down.sql` | Créer | Rollback |

### Exemples d'API

```http
# Lister les versions d'une carte
GET /api/v1/teams/1/maps/3/versions?page=1&limit=20

# Lister les versions d'une carte filtrées par zone
GET /api/v1/teams/1/maps/3/versions?zone_id=5

# Lister les versions d'une zone spécifique
GET /api/v1/teams/1/zones/5/versions

# Restaurer une version (éditeur ou capitaine)
POST /api/v1/teams/1/zones/5/versions/42/restore
```

### Contraintes techniques

- **Handler pattern** : Suivre exactement le pattern des handlers existants (constructeur public, méthodes Gin).
- **Vérification restauration éditeur** : Dans `RestoreVersion`, vérifier `teamMemberRepo.FindByUserAndTeam(user.ID, teamID)` pour déterminer le rôle. Si `editor`, appliquer les restrictions (auteur + pas de modification par un autre depuis).
- **Lock restauration** : Avant la restauration, vérifier le lock via `lockRepo.IsLockedByEditor(teamID, mapID, user.ID)`. Si pas locké → 423.
- **Atomicité restauration** : La restauration doit supprimer tous les polygones existants et les recréer. Utiliser DELETE + INSERT dans une transaction :
  1. `BEGIN`
  2. Supprimer les zone_names pour la zone (ON CONFLICT dans SetName)
  3. Supprimer les zone_polygons pour la zone
  4. Insérer les nouveaux polygones
  5. Mettre à jour le nom
  6. `COMMIT`
- **Logging** : Logger `version_restore` dans edit_history avec old_data = état avant, new_data = snapshot restauré.
- **Stratégie snapshot** : Les snapshots sont créés pour chaque zone modifiée dans le batch. Dans `BatchActions`, après le traitement, appeler `saveZoneSnapshot` pour chaque zone affectée (traquer les zoneIDs modifiés dans une map).
- **Ne pas casser l'existant** : L'ajout du paramètre `versionRepo` à `ZoneHandler` doit être optionnel (nil autorisé pour compatibilité).

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
