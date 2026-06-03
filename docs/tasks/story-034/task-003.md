# Tâche #003 - Story #034 : Admin handler pour la gestion des archivistes et lecteurs

## Objectif
Étendre le panneau d'administration avec des endpoints API pour gérer les archivistes (rôle global) et les équipes avec rôle lecteur.

## Contexte
- Story #034 : `docs/stories/story-034.md`
- Dépend de : Tâche #001, Tâche #002 (modèles + middleware)
- Nécessaire pour : Tâche #005 (frontend admin UI)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter 6 nouveaux endpoints dans le groupe admin pour gérer les rôles d'archives. Tous sous le middleware `RequireAdmin`.

**Endpoints :**

1. **GET /admin/archive/archivists** — Liste tous les utilisateurs ayant le rôle "archivist" dans `global_roles`
   - Retourne : `{ archivists: [{ user_id, display_name }] }`

2. **POST /admin/archive/archivists** — Ajoute le rôle "archivist" à un utilisateur
   - Body : `{ user_id: number }`
   - Retourne : `{ message: "Archivist added" }`
   - Erreur si user_id inexistant (FK contrainte) → 400

3. **DELETE /admin/archive/archivists/:userId** — Retire le rôle "archivist" d'un utilisateur
   - Retourne : `{ message: "Archivist removed" }`
   - Erreur si l'utilisateur n'a pas le rôle → 404

4. **GET /admin/archive/reader-teams** — Liste toutes les équipes ayant le rôle "reader" dans `team_roles`
   - Retourne : `{ teams: [{ team_id, team_name }] }`

5. **POST /admin/archive/reader-teams** — Ajoute le rôle "reader" à une équipe
   - Body : `{ team_id: number }`
   - Retourne : `{ message: "Reader role added to team" }`
   - Erreur si team_id inexistant → 400

6. **DELETE /admin/archive/reader-teams/:teamId** — Retire le rôle "reader" d'une équipe
   - Retourne : `{ message: "Reader role removed from team" }`
   - Erreur si team_id n'a pas le rôle → 404

**Modifications du handler :**
- Ajouter `globalRoleRepo` et `teamRoleRepo` à la struct `AdminHandler`
- Mettre à jour `NewAdminHandler` pour accepter les deux nouveaux repositories
- Ajouter les 6 nouvelles méthodes

**Modifications des routes :**
- Mettre à jour `setupAdminRoutes` pour créer les nouveaux repositories et les passer à `NewAdminHandler`
- Ajouter les 6 routes sous le groupe admin

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/admin.go` | Modifier | Ajouter les nouveaux champs et méthodes |
| `api/routes.go` | Modifier | Ajouter les nouvelles routes admin + mettre à jour l'appel à NewAdminHandler |

### Signatures

```go
// Nouvelles méthodes sur AdminHandler

func (h *AdminHandler) ListArchivists(c *gin.Context)
func (h *AdminHandler) AddArchivist(c *gin.Context)
func (h *AdminHandler) RemoveArchivist(c *gin.Context)
func (h *AdminHandler) ListReaderTeams(c *gin.Context)
func (h *AdminHandler) AddReaderTeam(c *gin.Context)
func (h *AdminHandler) RemoveReaderTeam(c *gin.Context)
```

```go
// AdminHandler modifié
type AdminHandler struct {
    userRepo       models.UserRepository
    permRepo       models.PermissionRepository
    rolePermRepo   models.RolePermissionRepository
    teamRepo       models.TeamRepository
    teamMemberRepo models.TeamMemberRepository
    globalRoleRepo models.GlobalRoleRepository   // NOUVEAU
    teamRoleRepo   models.TeamRoleRepository      // NOUVEAU
    logger         services.Logger
}

func NewAdminHandler(
    userRepo models.UserRepository,
    permRepo models.PermissionRepository,
    rolePermRepo models.RolePermissionRepository,
    teamRepo models.TeamRepository,
    teamMemberRepo models.TeamMemberRepository,
    globalRoleRepo models.GlobalRoleRepository,   // NOUVEAU
    teamRoleRepo models.TeamRoleRepository,        // NOUVEAU
    logger services.Logger,
) *AdminHandler
```

```go
// Routes à ajouter dans setupAdminRoutes (api/routes.go)
archiveGroup := admin.Group("/archive")
{
    archiveGroup.GET("/archivists", adminHandler.ListArchivists)
    archiveGroup.POST("/archivists", adminHandler.AddArchivist)
    archiveGroup.DELETE("/archivists/:userId", adminHandler.RemoveArchivist)
    archiveGroup.GET("/reader-teams", adminHandler.ListReaderTeams)
    archiveGroup.POST("/reader-teams", adminHandler.AddReaderTeam)
    archiveGroup.DELETE("/reader-teams/:teamId", adminHandler.RemoveReaderTeam)
}
```

### Contraintes techniques
- **Pattern** : Suivre exactement le même pattern que les handlers existants (mêmes codes erreur, même structure de réponse JSON)
- **Validation** : Vérifier que `user_id` / `team_id` sont des entiers valides depuis le body ou le paramètre URL
- **Réponses** : Utiliser `c.JSON(http.StatusOK, ...)` pour les succès, `apierror.RespondWithError` pour les erreurs
- **Idempotence** : DELETE doit être idempotent — si le rôle n'existe pas, retourner 200 ou 404 selon le choix (choisir 404 pour être explicite)

### Tests à implémenter

#### Tests d'intégration

**Fichier** : `internal/handlers/admin_archive_test.go` (ou ajouter au fichier existant)

| Scénario | Méthode | Résultat attendu |
|----------|---------|------------------|
| Lister archivistes (aucun) | GET /admin/archive/archivists | `[]` |
| Ajouter un archiviste | POST /admin/archive/archivists avec `{user_id: 1}` | 200 |
| Lister archivistes (1) | GET /admin/archive/archivists | `[{user_id: 1, display_name: ...}]` |
| Ajouter un archiviste sans user_id | POST avec body vide | 400 |
| Supprimer un archiviste | DELETE /admin/archive/archivists/1 | 200 |
| Supprimer un archiviste inexistant | DELETE /admin/archive/archivists/999 | 404 |
| Ajouter une équipe lectrice | POST /admin/archive/reader-teams avec `{team_id: 1}` | 200 |
| Lister équipes lectrices | GET /admin/archive/reader-teams | `[{team_id: 1, team_name: ...}]` |
| Supprimer une équipe lectrice | DELETE /admin/archive/reader-teams/1 | 200 |
| Accès non-admin | Sans token admin | 403 |

### Documentation
- Mettre à jour `docs/roles-archives.md` avec la documentation des endpoints admin
