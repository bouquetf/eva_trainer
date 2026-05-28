# Tâche #003 - Story #033 : Handler HTTP + routes pour les lignes de tir

## Objectif
Créer le handler HTTP `LineOfFireHandler` avec les endpoints CRUD, l'endpoint batch, l'historique, et l'enregistrement des routes.

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Tâche #002 (modèle/repository)
- Nécessaire pour : Tâche #006 (frontend API)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Handler `LineOfFireHandler`**

Créer un nouveau fichier `internal/handlers/line_of_fire.go` avec un handler suivant le pattern exact de `ZoneHandler` (voir `internal/handlers/zone.go`). Le handler doit :

1. **ListLinesOfFire** — GET `/api/v1/teams/:id/maps/:mapId/lines-of-fire`
   - Retourne toutes les lignes de tir pour une équipe et carte
   - Vérifie permission `line_of_fire:view`
   - Réponse : `{"lines_of_fire": [...]}`

2. **CreateLineOfFire** — POST `/api/v1/teams/:id/maps/:mapId/lines-of-fire`
   - Vérifie verrouillage (lock), carte obsolète, permission `line_of_fire:create`
   - Body : `{"name": "...", "polygon_data": {"cones": [...]}}`
   - Valide que le nom n'est pas vide et ne fait pas doublon
   - Crée la ligne, log edit_history, save version snapshot, sync template
   - Réponse : `{"line_of_fire": {...}}` HTTP 201

3. **GetLineOfFire** — GET `/api/v1/teams/:id/lines-of-fire/:lineOfFireId`
   - Vérifie permission `line_of_fire:view`
   - Retourne la ligne avec ses données détaillées
   - Réponse : `{"line_of_fire": {...}}`

4. **DeleteLineOfFire** — DELETE `/api/v1/teams/:id/lines-of-fire/:lineOfFireId`
   - Vérifie verrouillage, permission `line_of_fire:delete`
   - Log edit_history, sync template
   - Réponse : `{"message": "line of fire deleted"}`

5. **SetLineOfFireName** — PUT `/api/v1/teams/:id/lines-of-fire/:lineOfFireId/name`
   - Vérifie verrouillage, permission `line_of_fire:edit`
   - Body : `{"name": "..."}`
   - Valide nom (non vide, max 250, unique)
   - Log edit_history, save snapshot

6. **UpdateLineOfFireCone** — PUT `/api/v1/teams/:id/lines-of-fire/:lineOfFireId/cones/:coneIndex`
   - Vérifie verrouillage, permission `line_of_fire:edit`
   - Body : `{"origin": {...}, "left_end": {...}, "right_end": {...}}`
   - Met à jour un cône à un index donné dans polygon_data.cones
   - Log edit_history, save snapshot, sync template

7. **AddLineOfFireCone** — POST `/api/v1/teams/:id/lines-of-fire/:lineOfFireId/cones`
   - Vérifie verrouillage, permission `line_of_fire:edit`
   - Body : `{"origin": {...}, "left_end": {...}, "right_end": {...}}`
   - Ajoute un nouveau cône à polygon_data.cones
   - Log edit_history, save snapshot, sync template

8. **RemoveLineOfFireCone** — DELETE `/api/v1/teams/:id/lines-of-fire/:lineOfFireId/cones/:coneIndex`
   - Vérifie verrouillage, permission `line_of_fire:edit`
   - Supprime le cône à l'index donné. Si c'est le dernier, supprime la ligne entière.
   - Log edit_history, save snapshot, sync template

9. **BatchActions** — POST `/api/v1/teams/:id/maps/:mapId/lines-of-fire/batch`
   - Similaire à `ZoneHandler.BatchActions` mais pour les lignes de tir
   - Actions supportées : `create_line_of_fire`, `add_cone`, `update_cone`, `remove_cone`, `delete_line_of_fire`, `rename_line_of_fire`

10. **ImportLinesOfFire** — POST `/api/v1/teams/:id/maps/:mapId/lines-of-fire/import`
    - Vérifie verrouillage, permission `line_of_fire:create`
    - Body : `{"lines_of_fire": [{"name": "...", "polygon_data": {"cones": [...]}}]}`
    - Importe plusieurs lignes de tir en une requête (atomique)
    - Log edit_history pour chaque ligne créée
    - Sync template une fois après tout l'import
    - Réponse : `{"lines_of_fire": [...]}` HTTP 201
    - Suivre le pattern de `ZoneHandler.ImportZones` si existant

11. **ExportLinesOfFire** — GET `/api/v1/teams/:id/maps/:mapId/lines-of-fire/export`
    - Vérifie permission `line_of_fire:view`
    - Retourne toutes les lignes de tir de la carte dans un format exportable
    - Réponse : `{"lines_of_fire": [...]}` avec structure complète
    - Encodage JSON indenté pour lisibilité

**Routes (`api/routes.go`)**

Dans `setupTeamRoutes`, ajouter sous le groupe `teams` :

```go
teams.GET("/:id/maps/:mapId/lines-of-fire",
    middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:view"),
    lineOfFireHandler.ListLinesOfFire,
)
teams.POST("/:id/maps/:mapId/lines-of-fire",
    middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:create"),
    lineOfFireHandler.CreateLineOfFire,
)
teams.POST("/:id/maps/:mapId/lines-of-fire/batch",
    middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
    lineOfFireHandler.BatchActions,
)
teams.GET("/:id/lines-of-fire/:lineOfFireId",
    middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:view"),
    lineOfFireHandler.GetLineOfFire,
)
teams.DELETE("/:id/lines-of-fire/:lineOfFireId",
    middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:delete"),
    lineOfFireHandler.DeleteLineOfFire,
)
teams.PUT("/:id/lines-of-fire/:lineOfFireId/name",
    middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
    lineOfFireHandler.SetLineOfFireName,
)
teams.POST("/:id/lines-of-fire/:lineOfFireId/cones",
    middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
    lineOfFireHandler.AddLineOfFireCone,
)
teams.PUT("/:id/lines-of-fire/:lineOfFireId/cones/:coneIndex",
    middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
    lineOfFireHandler.UpdateLineOfFireCone,
)
teams.DELETE("/:id/lines-of-fire/:lineOfFireId/cones/:coneIndex",
    middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:edit"),
    lineOfFireHandler.RemoveLineOfFireCone,
)
teams.POST("/:id/maps/:mapId/lines-of-fire/import",
    middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:create"),
    lineOfFireHandler.ImportLinesOfFire,
)
teams.GET("/:id/maps/:mapId/lines-of-fire/export",
    middleware.RequireTeamRole(logger, teamMemberRepo, "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "line_of_fire:view"),
    lineOfFireHandler.ExportLinesOfFire,
)
```

Créer l'instance `lineOfFireHandler` dans `setupTeamRoutes` :
```go
lineOfFireRepo := models.NewLineOfFireRepository(db)
lofVersionRepo := models.NewLineOfFireVersionRepository(db)
lineOfFireHandler := handlers.NewLineOfFireHandler(lineOfFireRepo, mapRepo, lockRepo, historyRepo, teamMemberRepo, templateSyncSvc, lofVersionRepo, logger)
```

### Gestion d'erreurs

| Condition | Code HTTP | Error code |
|-----------|-----------|------------|
| Nom vide | 400 | `line_of_fire_name_required` |
| Nom trop long (>250) | 400 | `line_of_fire_name_too_long` |
| Nom déjà pris | 409 | `line_of_fire_name_taken` |
| Ligne introuvable | 404 | `line_of_fire_not_found` |
| Carte obsolète | 400 | `map_obsolete` |
| Verrou requis | 423 | `lock_required` |
| Index de cône invalide | 400 | `invalid_cone_index` |
| Permission insuffisante | 403 | `forbidden` |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/line_of_fire.go` | Créer | Handler complet |
| `api/routes.go` | Modifier | Ajouter les routes et l'instance |

### Contraintes techniques
- **Pattern** : Suivre exactement le pattern de `ZoneHandler` (verrouillage, snapshot, sync template, edit history).
- **Lock** : Tous les endpoints de modification doivent vérifier `lockRepo.IsLockedByEditor(teamID, mapID, user.ID)`.
- **Snapshot** : Après chaque modification, sauvegarder une version dans `line_of_fire_versions`.
- **Template sync** : Appeler `templateSyncSvc.SyncAfterLineOfFireChange(teamID, mapID)` après chaque modification (à implémenter dans Tâche #004).
- **Historique** : Logger dans `edit_history` avec `line_of_fire_id` (nullable).

### Tests à implémenter

Voir Tâche #009 pour les tests complets.
