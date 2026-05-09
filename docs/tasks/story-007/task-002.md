# Tâche #002 - Story #007 : Split API Handlers and Routes

## Objectif

Créer les endpoints API REST pour la gestion des splits (admin CRUD), la résolution des maps d'un split, et la consultation publique des splits, ainsi que les routes associées.

## Contexte

- Story #007 : `docs/stories/story-007.md` - Split Management and Filtering
- Dépend de : Tâche #001 (Split model et repository), Tâche #005 Map Management (MapRepository existe déjà)
- Nécessaire pour : Tâche #004 (tests handler)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - SplitHandler :**
Créer le handler `SplitHandler` dans `internal/handlers/split.go` avec les méthodes suivantes :

**Endpoints admin (sous `/api/v1/admin/splits`) :**
- `CreateSplit` : Crée un nouveau split. Corps JSON `{ "name": "...", "map_names": [...], "start_date": "YYYY-MM-DD" (optionnel), "end_date": "YYYY-MM-DD" (optionnel) }`. Valide les entrées et retourne 201 avec le split créé. Les dates sont optionnelles et doivent être au format ISO 8601 date (YYYY-MM-DD).
- `ListSplits` : Liste tous les splits. Retourne 200 avec `{ "splits": [...] }`.
- `GetSplit` : Récupère un split par son ID. Paramètre `:id`. Retourne 200 avec le split. 404 si inexistant.
- `UpdateSplit` : Met à jour un split. Corps JSON identique à Create. Paramètre `:id`. Retourne 200 avec le split mis à jour. 404 si inexistant.
- `DeleteSplit` : Supprime un split. Paramètre `:id`. Retourne 200 avec `{ "message": "split deleted" }`. 404 si inexistant.
- `GetSplitMaps` : Résout les maps d'un split. Paramètre `:id`. Retourne 200 avec `{ "maps": [...] }` (liste des maps dont le nom est dans le split). 404 si split inexistant.

**Endpoints publics (nécessitant authentification, sous `/api/v1/splits`) :**
- `ListSplitsPublic` : Liste tous les splits (identique à ListSplits mais sans privilège admin, juste authentifié). Retourne 200 avec `{ "splits": [...] }`. Destiné à être utilisé par le frontend de révision pour afficher la liste des splits disponibles.
- `GetSplitMapsPublic` : Résout les maps d'un split (identique à GetSplitMaps mais juste authentifié). Retourne 200 avec `{ "maps": [...] }`. Destiné au filtrage par split pendant la révision.

**Partie 2 - Modification de ListMaps dans MapHandler :**
Modifier le handler `ListMaps` existant dans `internal/handlers/map.go` pour accepter un query parameter optionnel `?name=Atlantis`. Quand `name` est fourni, il appelle `mapRepo.FindByName(name)` au lieu de `mapRepo.FindAll()`. Cela permet à l'admin de lister les cartes disponibles pour un nom "eva" donné sans créer de nouvelle route.

```go
func (h *MapHandler) ListMaps(c *gin.Context) {
    name := c.Query("name")
    var maps []*models.Map
    var err error
    if name != "" {
        maps, err = h.mapRepo.FindByName(name)
    } else {
        maps, err = h.mapRepo.FindAll()
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "internal_error",
            "message": "failed to list maps",
        })
        return
    }
    c.JSON(http.StatusOK, gin.H{"maps": maps})
}
```

**Cas nominaux :**
- Admin crée un split → 201 + split créé
- Admin liste les splits → 200 + liste
- Admin récupère un split → 200 + split
- Admin met à jour un split → 200 + split mis à jour
- Admin supprime un split → 200 + message
- Admin résout les maps d'un split → 200 + maps correspondantes
- Utilisateur authentifié liste les splits publics → 200 + liste
- Utilisateur authentifié résout les maps d'un split → 200 + maps correspondantes
- Admin récupère les maps par nom → 200 + maps

**Cas limites :**
- Corps JSON invalide → 400 `invalid_request_body`
- Nom manquant → 400 `split_name_required`
- map_names vide → 400 `split_map_names_required`
- Dates invalides (start > end) → 400 `invalid_date_range`
- Format de date invalide → 400 `invalid_date_format`
- Split inexistant → 404 `split_not_found`
- Aucune carte pour un nom donné → 200 avec `{ "maps": [] }`
- Aucun split en base → 200 avec `{ "splits": [] }`

**Gestion d'erreurs :**
- Erreur de parsing JSON → 400 avec `{ "error": "invalid_request_body", "message": "..." }`
- Erreur de validation → 400 avec code d'erreur spécifique
- Split introuvable → 404 avec `{ "error": "split_not_found", "message": "..." }`
- Erreur interne → 500 avec `{ "error": "internal_error", "message": "..." }`

**Partie 3 - Permission `split:view` :**
Ajouter une permission `split:view` pour la consultation des splits. Elle est assignée à tous les rôles.

Deux actions nécessaires :

**A. Migration SQL** : Créer un nouveau fichier `db/migrations/000008_add_split_permission.up.sql` :
```sql
INSERT INTO permissions (feature, description) VALUES
    ('split:view', 'Voir les splits et leurs cartes')
ON CONFLICT (feature) DO NOTHING;
```
Créer aussi le fichier `db/migrations/000008_add_split_permission.down.sql` :
```sql
DELETE FROM permissions WHERE feature = 'split:view';
```

**B. Mise à jour du seed Go** : Ajouter `split:view` dans `db/seed/permissions.go`, dans la liste `assignments`, pour tous les rôles :
```go
// À ajouter dans la liste assignments :
{role: "captain", feature: "split:view"},
{role: "cocaptain", feature: "split:view"},
{role: "editor", feature: "split:view"},
{role: "member", feature: "split:view"},
```

**Note :** Les endpoints d'administration (CRUD splits) restent protégés par `RequireAdmin` uniquement, pas besoin de permission `split:manage`.

**Partie 4 - Middleware `RequireGlobalPermission` :**
Créer un nouveau middleware dans `internal/middleware/permission.go` :
```go
func RequireGlobalPermission(
    permRepo models.PermissionRepository,
    rolePermRepo models.RolePermissionRepository,
    feature string,
) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, err := MustGetUser(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "unauthorized",
                "message": "unauthorized",
            })
            c.Abort()
            return
        }

        if user.IsAdmin {
            c.Next()
            return
        }

        perm, err := permRepo.FindByFeature(feature)
        if err != nil {
            c.JSON(http.StatusForbidden, gin.H{
                "error":   "insufficient_permission",
                "message": "you don't have the required permission for this action",
            })
            c.Abort()
            return
        }

        // Vérifie que le rôle 'member' (le plus bas niveau) a cette permission
        // Si member y a accès, tout utilisateur authentifié y a accès
        hasPerm, err := rolePermRepo.RoleHasPermission("member", perm.ID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "internal_error",
                "message": "internal server error",
            })
            c.Abort()
            return
        }

        if !hasPerm {
            c.JSON(http.StatusForbidden, gin.H{
                "error":   "insufficient_permission",
                "message": "you don't have the required permission for this action",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

Ce middleware est utilisé sur les endpoints splits publics à la place d'un simple `RequireAuth` :
```go
splitsPublic.Use(middleware.RequireAuth(deps.userRepo, deps.jwtService))
splitsPublic.Use(middleware.RequireGlobalPermission(permRepo, rolePermRepo, "split:view"))
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/split.go` | Créer | SplitHandler avec toutes les méthodes CRUD + résolution |
| `internal/handlers/map.go` | Modifier | Modifier `ListMaps` pour accepter `?name=` |
| `internal/middleware/permission.go` | Modifier | Ajouter `RequireGlobalPermission` middleware |
| `db/migrations/000008_add_split_permission.up.sql` | Créer | Nouvelle migration pour la permission `split:view` |
| `db/migrations/000008_add_split_permission.down.sql` | Créer | Rollback de la migration |
| `db/seed/permissions.go` | Modifier | Ajouter l'assignation de `split:view` à tous les rôles |
| `api/routes.go` | Modifier | Ajouter les routes admin splits + routes publiques splits |

### Signatures

```go
// internal/handlers/split.go
package handlers

import (
    "net/http"
    "strconv"
    "time"

    "eva_trainer/internal/models"

    "github.com/gin-gonic/gin"
)

type SplitHandler struct {
    splitRepo models.SplitRepository
}

func NewSplitHandler(splitRepo models.SplitRepository) *SplitHandler {
    return &SplitHandler{splitRepo: splitRepo}
}

// Admin endpoints
func (h *SplitHandler) CreateSplit(c *gin.Context)
func (h *SplitHandler) ListSplits(c *gin.Context)
func (h *SplitHandler) GetSplit(c *gin.Context)
func (h *SplitHandler) UpdateSplit(c *gin.Context)
func (h *SplitHandler) DeleteSplit(c *gin.Context)
func (h *SplitHandler) GetSplitMaps(c *gin.Context)

// Public endpoints (authenticated, no admin required)
func (h *SplitHandler) ListSplitsPublic(c *gin.Context)
func (h *SplitHandler) GetSplitMapsPublic(c *gin.Context)
```

```go
// DTO pour les requêtes de création/mise à jour de split
type createSplitRequest struct {
    Name      string   `json:"name" binding:"required"`
    MapNames  []string `json:"map_names" binding:"required"`
    StartDate *string  `json:"start_date"`
    EndDate   *string  `json:"end_date"`
}
```

// Note : Plus de nouvelle méthode GetMapsByName. La modification se fait sur ListMaps existant.

### Implémentation détaillée des méthodes

#### CreateSplit
1. Parser le body JSON dans `createSplitRequest`
2. Si `binding` échoue → 400 `invalid_request_body`
3. Parser les dates optionnelles : si une date string est fournie, parser avec `time.Parse("2006-01-02", dateStr)`. Si le parsing échoue → 400 `invalid_date_format`
4. Convertir les dates parsées en `*time.Time`
5. Appeler `h.splitRepo.Create(name, mapNames, startDate, endDate)`
6. En cas d'erreur de validation (`ErrSplitNameRequired`, `ErrSplitMapNamesRequired`, `ErrInvalidDateRange`) → 400 avec le code d'erreur correspondant
7. En cas d'erreur inattendue → 500
8. Retourner 201 avec `{ "split": split }`

#### ListSplits / ListSplitsPublic
1. Appeler `h.splitRepo.FindAll()`
2. En cas d'erreur → 500
3. Retourner 200 avec `{ "splits": splits }`

#### GetSplit
1. Parser `c.Param("id")` en int
2. Appeler `h.splitRepo.FindByID(id)`
3. Si `ErrSplitNotFound` → 404
4. En cas d'erreur → 500
5. Retourner 200 avec `{ "split": split }`

#### UpdateSplit
1. Parser `c.Param("id")` en int
2. Parser le body JSON
3. Parser les dates optionnelles (comme CreateSplit)
4. Appeler `h.splitRepo.Update(id, name, mapNames, startDate, endDate)`
5. Si `ErrSplitNotFound` → 404
6. Si erreur de validation → 400
7. Retourner 200 avec `{ "split": split }`

#### DeleteSplit
1. Parser `c.Param("id")` en int
2. Appeler `h.splitRepo.Delete(id)`
3. Si `ErrSplitNotFound` → 404
4. Retourner 200 avec `{ "message": "split deleted" }`

#### GetSplitMaps / GetSplitMapsPublic
1. Parser `c.Param("id")` en int
2. Appeler `h.splitRepo.FindMapsBySplitID(id)`
3. Si `ErrSplitNotFound` → 404
4. Retourner 200 avec `{ "maps": maps }`

### Routes à ajouter dans `api/routes.go`

#### 1. Créer une fonction partagée pour les dépendances communes

Ajouter en haut de `api/routes.go`, avant les fonctions de setup :

```go
type coreDeps struct {
    userRepo      models.UserRepository
    jwtService    *services.JWTService
    permRepo      models.PermissionRepository
    rolePermRepo  models.RolePermissionRepository
}

func newCoreDeps(cfg *config.AppConfig) coreDeps {
    db := database.DB()
    userRepo := models.NewUserRepository(db)
    jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)
    permRepo := models.NewPermissionRepository(db)
    rolePermRepo := models.NewRolePermissionRepository(db)
    return coreDeps{userRepo: userRepo, jwtService: jwtService, permRepo: permRepo, rolePermRepo: rolePermRepo}
}
```

#### 2. Ajouter les routes splits dans `setupAdminRoutes`

```go
// Dans setupAdminRoutes, après la création de mapHandler :
splitRepo := models.NewSplitRepository(db)
splitHandler := handlers.NewSplitHandler(splitRepo)

admin.POST("/splits", splitHandler.CreateSplit)
admin.GET("/splits", splitHandler.ListSplits)
admin.GET("/splits/:id", splitHandler.GetSplit)
admin.PUT("/splits/:id", splitHandler.UpdateSplit)
admin.DELETE("/splits/:id", splitHandler.DeleteSplit)
admin.GET("/splits/:id/maps", splitHandler.GetSplitMaps)
```

#### 3. Ajouter les routes splits publiques dans `SetupRoutes`

```go
// Dans SetupRoutes, créer les dépendances partagées :
deps := newCoreDeps(cfg)

// Puis après apiV1.GET("/maps/:id/image", ...) :
splitRepo := models.NewSplitRepository(db)
splitHandler := handlers.NewSplitHandler(splitRepo)

splitsPublic := apiV1.Group("/splits")
splitsPublic.Use(middleware.RequireAuth(deps.userRepo, deps.jwtService))
splitsPublic.Use(middleware.RequireGlobalPermission(deps.permRepo, deps.rolePermRepo, "split:view"))
{
    splitsPublic.GET("", splitHandler.ListSplitsPublic)
    splitsPublic.GET("/:id/maps", splitHandler.GetSplitMapsPublic)
}
```

### Contraintes techniques

- **Format handler** : Suivre exactement le pattern des handlers existants (MapHandler, TeamHandler, etc.) : struct avec repository, constructeur public `New*Handler`.
- **Format réponse erreur** : Utiliser `gin.H{"error": "error_code", "message": "human message"}` comme dans les handlers existants.
- **Format dates** : Les dates sont au format `YYYY-MM-DD` (ISO 8601) en JSON. Utiliser `time.Parse("2006-01-02", ...)` pour le parsing.
- **Binding JSON** : Utiliser `c.ShouldBindJSON(&req)` pour le parsing. Le binding standard de Gin avec `binding:"required"` gère la validation des champs obligatoires.
- **Middleware** : Les routes admin utilisent déjà `middleware.RequireAuth` + `middleware.RequireAdmin()` (c'est le groupe existant). Les routes publiques splits doivent utiliser `middleware.RequireAuth` + `middleware.RequireGlobalPermission(permRepo, rolePermRepo, "split:view")`.
- **Permission seed** : Ajouter `split:view` dans le fichier de seed des permissions (`db/seed/permissions.go`). Vérifie comment la seed est faite et suit le même pattern (insertion dans `permissions` puis dans `role_permissions` pour tous les rôles).
- **Import du package handlers** : Vérifier que les nouveaux handlers sont accessibles depuis `api/routes.go`. Le package `api` importe déjà `eva_trainer/internal/handlers`.

### Tests à implémenter

Les tests seront créés dans la Tâche #004.

### Documentation

Voir Tâche #005 pour la documentation utilisateur.
