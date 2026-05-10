# Tâche #001 - Story #010 : Lock Status API avec nom d'éditeur et temps restant

## Objectif

Enrichir les endpoints de verrouillage existants pour exposer le nom d'affichage de l'éditeur et le temps restant avant expiration, et ajouter un endpoint admin de nettoyage forcé.

## Contexte

- Story #010 : `docs/stories/story-010.md` - Map Locking Mechanism for Concurrent Editing
- Dépend de : Story #009 (modèle `EditLock`, handler `EditLockHandler`, routes lock existantes)
- Nécessaire pour : Tâche #002 (affichage frontend du temps restant et du nom d'éditeur)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Ajout du champ `remaining_seconds` via DTO

Le champ `remaining_seconds` est un champ **calculé** (pas stocké en base). Il ne doit PAS être ajouté à la struct `EditLock` du modèle. À la place, créer un DTO de réponse dédié dans `internal/handlers/edit_lock.go` :

```go
// DTO pour les réponses lock, inclut le champ calculé remaining_seconds
type LockResponse struct {
    ID                int       `json:"id"`
    TeamID            int       `json:"team_id"`
    MapID             int       `json:"map_id"`
    EditorID          int       `json:"editor_id"`
    EditorName        string    `json:"editor_name,omitempty"`
    LockedAt          time.Time `json:"locked_at"`
    ExpiresAt         time.Time `json:"expires_at"`
    RemainingSeconds  int       `json:"remaining_seconds"`
}
```

Fonction utilitaire pour convertir un `*EditLock` en `LockResponse` :
```go
func newLockResponse(lock *models.EditLock, editorName string) LockResponse {
    remaining := int(time.Until(lock.ExpiresAt).Seconds())
    if remaining < 0 {
        remaining = 0
    }
    return LockResponse{
        ID:               lock.ID,
        TeamID:           lock.TeamID,
        MapID:            lock.MapID,
        EditorID:         lock.EditorID,
        EditorName:       editorName,
        LockedAt:         lock.LockedAt,
        ExpiresAt:        lock.ExpiresAt,
        RemainingSeconds: remaining,
    }
}
```

**Fichiers à modifier :**

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/edit_lock.go` | Modifier | Ajouter DTO `LockResponse`, utilitaire `newLockResponse`, remplacer les `gin.H{"lock": lock}` par `gin.H{"lock": newLockResponse(...)}` |
| `internal/handlers/edit_lock.go` | Modifier | Idem pour `AcquireLock`, `RefreshLock`, `GetLockStatus` |

#### Partie 2 - Ajout du nom d'éditeur dans GetLockStatus

Actuellement, `GetLockStatus` retourne `{ "locked": bool, "lock": EditLock }` où `EditLock.editor_id` est un entier. Le frontend a besoin du `display_name` de l'utilisateur pour afficher "Verrouillé par [nom]".

**Fichiers à modifier :**

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/edit_lock.go` | Modifier | Dans `GetLockStatus`, faire une requête à la table `users` pour récupérer le `display_name` de l'éditeur |
| `internal/handlers/edit_lock.go` | Modifier | Ajouter `editor_name` dans la réponse JSON de `GetLockStatus` |
| `internal/handlers/edit_lock.go` | Modifier | Ajouter `editor_name` dans la réponse de `AcquireLock` (quand 409 Conflict avec `locked_by`) |

La réponse de `GetLockStatus` doit devenir :
```json
{
  "locked": true,
  "lock": {
    "id": 1,
    "team_id": 1,
    "map_id": 2,
    "editor_id": 5,
    "editor_name": "Player123",
    "locked_at": "2026-05-10T12:00:00Z",
    "expires_at": "2026-05-10T12:15:00Z",
    "remaining_seconds": 840
  }
}
```

La réponse de `AcquireLock` en cas de 409 doit contenir :
```json
{
  "error": "lock_already_acquired",
  "message": "edit lock already acquired by another editor",
  "locked_by": 5,
  "editor_name": "Player123"
}
```

**Contrainte technique :** Injecter `*sql.DB` ou un `UserRepository` dans `EditLockHandler` pour pouvoir requêter le nom d'utilisateur. Actuellement `EditLockHandler` ne prend qu'un `EditLockRepository`. Solution : ajouter un `UserRepository` au constructeur. Voir le pattern existant dans `ZoneHandler` qui reçoit plusieurs repositories.

```go
// Nouveau constructeur
func NewEditLockHandler(lockRepo models.EditLockRepository, userRepo models.UserRepository) *EditLockHandler
```

**Mise à jour des dépendances :** Dans `api/routes.go`, créer un `UserRepository` et le passer à `NewEditLockHandler`.

#### Partie 3 - Endpoint admin de nettoyage des locks

Ajouter un endpoint admin pour forcer le cleanup de **tous** les locks expirés ou de **tous** les locks (forcer la libération).

| Méthode | Route | Description |
|---------|-------|-------------|
| POST | `/api/v1/admin/locks/cleanup` | Nettoie uniquement les locks expirés (via `CleanupExpired()`) |
| DELETE | `/api/v1/admin/locks` | Supprime TOUS les locks (peut être utile en urgence) |

**Cas nominaux :**
- Admin appelle POST `/api/v1/admin/locks/cleanup` → tous les locks expirés sont supprimés → réponse `{ "cleaned": 3 }`
- Admin appelle DELETE `/api/v1/admin/locks` → tous les locks sont supprimés → réponse `{ "cleared": true, "count": 5 }`

**Cas limites :**
- Aucun lock expiré → `{ "cleaned": 0 }`
- Aucun lock du tout → `{ "cleared": true, "count": 0 }`
- Non-admin essaie d'appeler → 403 Forbidden (géré par `middleware.RequireAdmin()`)

**Fichier à créer :** Ajouter les méthodes à `internal/handlers/admin.go` ou créer une méthode `CleanupLocks` et `ForceClearLocks` dans un handler existant (recommandé : les ajouter à `internal/handlers/admin.go`).

**Routes à ajouter dans `api/routes.go`** section admin :

```go
admin.POST("/locks/cleanup", adminHandler.CleanupExpiredLocks)
admin.DELETE("/locks", adminHandler.ForceClearLocks)
```

Pour le DELETE ALL, il faut ajouter une méthode au `EditLockRepository` :
```go
// Dans internal/models/repository.go, interface EditLockRepository
ForceClearAll() (int64, error)

// Dans internal/models/edit_lock.go
func (r *editLockRepo) ForceClearAll() (int64, error) {
    result, err := r.db.Exec(`DELETE FROM edit_locks`)
    if err != nil {
        return 0, fmt.Errorf("failed to force clear locks: %w", err)
    }
    return result.RowsAffected()
}
```

#### Partie 4 - Route POST `/lock/release` pour beforeunload

Ajouter un endpoint POST dédié que le `beforeunload` du navigateur peut appeler avec `keepalive: true` pour libérer le verrou sans attendre la réponse. Cet endpoint est tolérant : il retourne 200 même si le lock n'existe pas.

**Méthode `ReleaseLockForEditor` dans `internal/handlers/edit_lock.go` :**
```go
func (h *EditLockHandler) ReleaseLockForEditor(c *gin.Context) {
    mapID, err := strconv.Atoi(c.Param("mapId"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_map_id"})
        return
    }
    teamID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id"})
        return
    }
    user, err := middleware.MustGetUser(c)
    if err != nil {
        return
    }

    // Tolérant : ne pas retourner d'erreur si lock_not_found
    err = h.lockRepo.ReleaseByEditor(teamID, mapID, user.ID)
    if err != nil && !errors.Is(err, models.ErrLockNotFound) {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "lock released"})
}
```

**Route à ajouter dans `api/routes.go` (dans le groupe teams :id) :**
```go
teams.POST("/:id/maps/:mapId/lock/release",
    middleware.RequireAuth(userRepo, jwtService),
    middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
    lockHandler.ReleaseLockForEditor,
)
```

#### Partie 5 - Mise à jour des routes lock dans api/routes.go

Actuellement, les routes lock n'ont que `zone:lock` permission sur POST. Vérifier que le middleware `RequirePermission` existe bien pour `zone:lock`. Les routes actuelles sont correctes mais il faut :

1. Mettre à jour le constructeur de `EditLockHandler` pour passer le `UserRepository` :

```go
// Dans setupTeamRoutes, remplacer :
lockHandler := handlers.NewEditLockHandler(lockRepo)
// Par :
userRepoForLock := models.NewUserRepository(db)
lockHandler := handlers.NewEditLockHandler(lockRepo, userRepoForLock)
```

2. Ajouter les routes admin :

```go
// Dans setupAdminRoutes, après les routes map existantes :
lockRepo := models.NewEditLockRepository(db)
lockHandler := handlers.NewEditLockHandler(lockRepo, userRepo)

admin.POST("/locks/cleanup", lockHandler.CleanupExpiredLocks)
admin.DELETE("/locks", lockHandler.ForceClearLocks)
```

**Cas nominaux :**
- Requête GET lock status → retourne `locked`, `lock` avec `editor_name`, `remaining_seconds`
- Requête POST acquire lock (succès) → retourne `lock` avec `remaining_seconds`
- Requête POST acquire lock (conflit 409) → retourne `locked_by` + `editor_name`
- Requête PUT refresh lock → retourne `lock` avec `remaining_seconds` mis à jour
- Requête POST .../lock/release (beforeunload) → 200 OK (même si lock inexistant)
- Admin POST cleanup locks → retourne nombre de locks nettoyés
- Admin DELETE all locks → retourne confirmation

**Gestion d'erreurs :**
- Erreur de requête user name dans GetLockStatus → ne pas blocker, retourner `editor_name: "inconnu"` ou `editor_name: ""`
- Erreur DB dans cleanup → 500 Internal Server Error

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/edit_lock.go` | Vérifier | `RemainingSeconds` est un champ calculé dans le DTO handler, PAS dans le modèle |
| `internal/models/repository.go` | Modifier | Ajouter `ForceClearAll() (int64, error)` à l'interface `EditLockRepository` |
| `internal/models/edit_lock.go` | Modifier | Implémenter `ForceClearAll()` |
| `internal/handlers/edit_lock.go` | Modifier | Ajouter `userRepo` au handler, créer DTO `LockResponse` avec `remaining_seconds` calculé, enrichir réponses avec `editor_name`, ajouter méthodes admin |
| `internal/handlers/admin.go` | Modifier | (Optionnel) Si on préfère mettre les méthodes admin ici plutôt que dans edit_lock.go |
| `api/routes.go` | Modifier | Mettre à jour constructeur `NewEditLockHandler`, ajouter routes admin lock, ajouter route `POST .../lock/release` |
| `db/migrations/000005_seed_permissions.up.sql` | Vérifier | Déjà existant, juste vérifier que `zone:lock` est bien seedé |

### Signatures

```go
// DTO dans internal/handlers/edit_lock.go (PAS dans le modèle)

type LockResponse struct {
	ID                int       `json:"id"`
	TeamID            int       `json:"team_id"`
	MapID             int       `json:"map_id"`
	EditorID          int       `json:"editor_id"`
	EditorName        string    `json:"editor_name,omitempty"`
	LockedAt          time.Time `json:"locked_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	RemainingSeconds  int       `json:"remaining_seconds"`
}

func newLockResponse(lock *models.EditLock, editorName string) LockResponse {
	remaining := int(time.Until(lock.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return LockResponse{
		ID:               lock.ID,
		TeamID:           lock.TeamID,
		MapID:            lock.MapID,
		EditorID:         lock.EditorID,
		EditorName:       editorName,
		LockedAt:         lock.LockedAt,
		ExpiresAt:        lock.ExpiresAt,
		RemainingSeconds: remaining,
	}
}

// Nouvelle méthode dans l'interface EditLockRepository
type EditLockRepository interface {
	// ... méthodes existantes ...
	ForceClearAll() (int64, error)
}
```

```go
// Modifications dans internal/handlers/edit_lock.go

type EditLockHandler struct {
	lockRepo models.EditLockRepository
	userRepo models.UserRepository
}

func NewEditLockHandler(lockRepo models.EditLockRepository, userRepo models.UserRepository) *EditLockHandler

// Nouvelles méthodes
func (h *EditLockHandler) CleanupExpiredLocks(c *gin.Context)
func (h *EditLockHandler) ForceClearLocks(c *gin.Context)
```

```go
// Helper pour récupérer le display_name d'un utilisateur
func (h *EditLockHandler) getUserDisplayName(userID int) string {
	user, err := h.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return ""
	}
	return user.DisplayName
}
```

### Contraintes techniques

- **Champ calculé** : `RemainingSeconds` n'est PAS stocké en base. Ne pas modifier la struct `EditLock`. Utiliser le DTO `LockResponse` dans le handler et la fonction `newLockResponse()` pour convertir avant chaque réponse JSON.
- **UserRepository** : Injecter `models.NewUserRepository(db)` dans le constructeur. Utiliser `userRepo.FindByID(editorID)` pour obtenir le `DisplayName`. Ne pas blocker si la requête échoue, passer `editorName = ""` dans `newLockResponse`.
- **Pattern handler** : Conserver le pattern existant. Les méthodes admin peuvent être dans le même fichier `edit_lock.go` ou dans `admin.go`. Préférer `edit_lock.go` pour garder la logique lock centralisée, mais les routes doivent être dans le groupe admin.
- **AcquireLock (409)** : Dans la branche `ErrLockAlreadyAcquired`, récupérer `existing.EditorID` (déjà fait) et ajouter le nom.
- **CleanupExpiredLocks** : Appeler `h.lockRepo.CleanupExpired()`. Retourner le nombre. Utiliser `LockCleanupService.CleanupNow()` si disponible, mais appeler directement le repo est plus simple.
- **ForceClearLocks** : Appeler `h.lockRepo.ForceClearAll()`. Retourner le nombre de locks supprimés. Attention : cette action est destructive et ne devrait être utilisée qu'en cas d'urgence.
- **Permissions admin** : Les routes admin sont déjà protégées par `middleware.RequireAdmin()` dans le groupe `/api/v1/admin`.
- **Tests** : Voir Tâche #003 pour les tests d'intégration.

### Tests à implémenter

Les tests seront créés dans la Tâche #003.

### Documentation

Aucune documentation à créer pour cette tâche. La documentation utilisateur du verrouillage sera créée dans la Tâche #003.
