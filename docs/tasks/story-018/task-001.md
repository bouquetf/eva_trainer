# Tâche #001 - Story #018 : Ajouter le nom de l'éditeur dans la réponse 409 du lock

## Objectif
Ajouter le `editor_name` (display name) du détenteur actuel du verrou dans la réponse HTTP 409 lorsque `AcquireLock` échoue car un autre éditeur détient déjà le verrou.

## Contexte
- Story #018 : Refinements du verrouillage d'édition
- Dépend de : Story #010 (Map Locking)
- Le handler `AcquireLock` renvoie actuellement une erreur 409 générique sans le nom de l'éditeur
- Le frontend (MapEditor.tsx) essaie déjà de lire `body.editor_name` dans le catch 409, mais le backend ne l'envoie pas
- Le composant `LockIndicator` supporte déjà l'affichage de `editorName` via la prop `editorName`

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Quand un éditeur B tente d'acquérir le verrou (`POST /api/v1/teams/:id/maps/:mapId/lock`) et que l'éditeur A détient déjà un verrou actif, la réponse 409 doit inclure le `display_name` de l'éditeur A et son `editor_id`.

**Cas nominal :**
- Éditeur A acquiert le verrou → 200 OK avec `{ "lock": { ..., "editor_name": "Alice", "editor_id": 1 } }`
- Éditeur B tente d'acquérir → 409 Conflict avec `{ "error": "lock_already_acquired", "message": "...", "editor_name": "Alice", "editor_id": 1 }`
- Si le verrou a expiré, l'éditeur B peut acquérir normalement (pas de 409)

**Cas limites :**
- L'éditeur A n'a pas de `display_name` (utilisateur sans pseudo défini) → renvoyer `editor_name: ""` (chaîne vide)
- Le verrou a expiré entre l'appel et la vérification → laisser le mécanisme existant gérer (Acquire expire puis réinsère)
- L'éditeur A a été supprimé entre-temps → `editor_name: "Utilisateur supprimé"` comme dans `getLockStatus`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/edit_lock.go` | Modifier | Ajouter `editor_name` et `editor_id` dans la réponse 409 |
| `internal/handlers/edit_lock_test.go` | Modifier | Vérifier que la réponse 409 contient `editor_name` |

### Détail d'implémentation

**Étape 1 : Étendre `APIError` pour supporter des champs supplémentaires**

Dans `pkg/apierror/errors.go`, ajouter un champ `Extra` à la struct `APIError` et modifier `RespondWithError` pour l'inclure dans la réponse JSON :

```go
type APIError struct {
    Code     ErrorCode            `json:"error"`
    Message  string               `json:"message"`
    HTTPCode int                  `json:"-"`
    Err      error                `json:"-"`
    Extra    map[string]any       `json:"-"` // non sérialisé par défaut
}
```

Puis créer une nouvelle méthode `WithExtra` sur `APIError` :

```go
func (e *APIError) WithExtra(key string, value any) *APIError {
    if e.Extra == nil {
        e.Extra = make(map[string]any)
    }
    e.Extra[key] = value
    return e
}
```

Modifier `RespondWithError` pour inclure `Extra` dans la réponse JSON :

```go
func RespondWithError(c *gin.Context, apiErr *APIError) {
    if apiErr == nil {
        return
    }
    resp := gin.H{
        "error":   apiErr.Code,
        "message": apiErr.Message,
    }
    for k, v := range apiErr.Extra {
        resp[k] = v
    }
    c.JSON(apiErr.HTTPCode, resp)
    c.Abort()
}
```

**Étape 2 : Utiliser l'Extra dans `AcquireLock`**

Dans `internal/handlers/edit_lock.go`, modifier le bloc 409 de `AcquireLock` :

```go
if errors.Is(err, models.ErrLockAlreadyAcquired) {
    existingLock, getErr := h.lockRepo.Get(teamID, mapID)
    if getErr == nil && existingLock != nil && existingLock.ExpiresAt.After(time.Now()) {
        editorName := h.getUserDisplayName(existingLock.EditorID)
        apiErr := apierror.NewAPIError("lock_already_acquired", "Cette carte est en cours d'édition par un autre éditeur.", http.StatusConflict)
        apiErr.WithExtra("editor_name", editorName).WithExtra("editor_id", existingLock.EditorID)
        apierror.RespondWithError(c, apiErr)
        return
    }
    apierror.RespondWithError(c, apierror.NewAPIError("lock_already_acquired", "edit lock already acquired by another editor", http.StatusConflict))
    return
}
```

Ne pas oublier d'importer `"time"` si ce n'est pas déjà fait (vérifier les imports existants — `"time"` est déjà importé dans `edit_lock.go`).

### Contraintes techniques
- **Framework** : Go + Gin
- **Pattern** : Utiliser `gin.H{}` pour construire la réponse JSON (déjà utilisé ailleurs dans le handler comme `gin.H{"lock": ...}`)
- **Style** : Respecter le format de réponse existant (snake_case)
- **Ne pas casser** les autres cas d'erreur du handler (403, 500)

### Tests à implémenter

#### Tests d'intégration (backend)
- **Fichier** : `internal/handlers/edit_lock_test.go`
- Dans `TestEditLockHandler_AcquireAndRejectSecond` (déjà existant), ajouter la vérification que la réponse 409 contient `editor_name` et `editor_id` :
  - Ligne ~72 : après avoir vérifié `errResp["error"]`, ajouter :
    ```go
    if editorName, ok := errResp["editor_name"]; !ok || editorName == "" {
        t.Errorf("expected editor_name in 409 response, got '%v'", editorName)
    }
    if editorID, ok := errResp["editor_id"]; !ok || editorID == nil {
        t.Errorf("expected editor_id in 409 response, got '%v'", editorID)
    }
    ```
  - Vérifier que `editor_id` est un float64 (JSON unmarshaling) et correspond à l'ID du premier éditeur (`f.user.ID`)
  - Vérifier que `editor_name` est une string non vide

### Documentation
- `docs/map-locking.md` : Ajouter une note sur le format de la réponse 409 incluant `editor_name` et `editor_id`
