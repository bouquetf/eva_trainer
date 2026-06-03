# Tâche #004 - Story #034 : Endpoint de vérification des permissions archives + routes placeholder

## Objectif
Créer un endpoint API permettant au frontend de vérifier si l'utilisateur courant a accès aux archives, et ajouter les routes placeholder pour les futures fonctionnalités d'archives.

## Contexte
- Story #034 : `docs/stories/story-034.md`
- Dépend de : Tâche #002 (middleware)
- Nécessaire pour : Tâche #005 (frontend)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Endpoint GET /api/v1/archive/permissions

Retourne un objet JSON indiquant les droits de l'utilisateur authentifié sur les archives :

```json
{
  "can_read": true,
  "can_write": false
}
```

Logique :
- `can_read` = admin OU (archivist dans global_roles) OU (membre d'une équipe avec reader dans team_roles)
- `can_write` = admin OU (archivist dans global_roles)

#### Routes placeholder

Ajouter quelques routes d'archives qui serviront de base pour les futures stories. Ces routes sont protégées par `RequireArchiveAccess` et retournent pour l'instant un simple 200 OK avec un message "not implemented yet".

Routes à créer :
- **GET /api/v1/archive/leagues** — Liste des ligues (read)
- **POST /api/v1/archive/leagues** — Créer une ligue (write)
- **GET /api/v1/archive/teams** — Liste des équipes eva (read)
- **POST /api/v1/archive/teams** — Créer une équipe eva (write)

**Handler** : Créer `internal/handlers/archive.go` avec un `ArchiveHandler` vide pour l'instant (méthodes retournant "not implemented").

**Router** : Créer une fonction `setupArchiveRoutes` dans `api/routes.go` (ou ajouter dans `SetupRoutes`).

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Créer | ArchiveHandler avec méthodes placeholder |
| `internal/handlers/archive_permissions.go` | Créer | Handler pour GET /archive/permissions |
| `api/routes.go` | Modifier | Ajouter setupArchiveRoutes |

### Signatures

```go
// internal/handlers/archive_permissions.go
type ArchivePermissionsHandler struct {
    globalRoleRepo models.GlobalRoleRepository
    teamRoleRepo   models.TeamRoleRepository
    logger         services.Logger
}

func NewArchivePermissionsHandler(
    globalRoleRepo models.GlobalRoleRepository,
    teamRoleRepo models.TeamRoleRepository,
    logger services.Logger,
) *ArchivePermissionsHandler

func (h *ArchivePermissionsHandler) GetPermissions(c *gin.Context)
```

```go
// internal/handlers/archive.go
type ArchiveHandler struct {
    logger services.Logger
}

func NewArchiveHandler(logger services.Logger) *ArchiveHandler

func (h *ArchiveHandler) ListLeagues(c *gin.Context)
func (h *ArchiveHandler) CreateLeague(c *gin.Context)
func (h *ArchiveHandler) ListEvaTeams(c *gin.Context)
func (h *ArchiveHandler) CreateEvaTeam(c *gin.Context)
```

### Contraintes techniques
- **Middleware** : La route `/archive/permissions` utilise `RequireAuth` mais PAS `RequireArchiveAccess` (elle doit être accessible à tous les utilisateurs authentifiés)
- **Routes placeholder** : Les routes sous `/archive/leagues` et `/archive/teams` utilisent `RequireAuth` + `RequireArchiveAccess("read" ou "write")`
- **Handler** : Les méthodes placeholder retournent `c.JSON(http.StatusOK, gin.H{"message": "not implemented yet"})`
- **Permissions** : Le handler GetPermissions doit réutiliser la même logique que le middleware (factoriser dans une fonction utilitaire si pertinent)

### Tests à implémenter

#### Tests d'intégration

**Fichier** : `internal/handlers/archive_permissions_test.go`

| Scénario | Input | Résultat attendu |
|----------|-------|------------------|
| Admin accède aux permissions | Token admin | `{can_read: true, can_write: true}` |
| Archiviste accède aux permissions | Token archiviste | `{can_read: true, can_write: true}` |
| Membre d'équipe reader | Token membre reader | `{can_read: true, can_write: false}` |
| Utilisateur sans droits | Token normal | `{can_read: false, can_write: false}` |
| Non authentifié | Pas de token | 401 |
| Accès à une route write sans droit write | Token normal vers POST /leagues | 403 |

**Fichier** : `internal/handlers/archive_test.go`

| Scénario | Input | Résultat attendu |
|----------|-------|------------------|
| GET /archive/leagues (lecteur) | Token avec droit read | `{message: "not implemented yet"}` |
| POST /archive/leagues (archiviste) | Token avec droit write | `{message: "not implemented yet"}` |
| POST /archive/leagues (lecteur seul) | Token sans droit write | 403 |

### Documentation
- Ajouter la section "Permissions API" dans `docs/roles-archives.md`
