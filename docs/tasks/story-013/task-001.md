# Tâche #001 - Story #013 : Error Types, API Error Format et Helpers

## Objectif

Créer l'infrastructure centralisée de gestion d'erreurs API : types d'erreur, codes standardisés, helpers de réponse et mapping erreurs Go → statuts HTTP.

## Contexte

- Story #013 : `docs/stories/story-013.md` — Error Handling and Logging
- Dépend de : Story #008 (middleware d'authentification et permissions existant)
- Nécessaire pour : Tâche #003 (middleware error handling), Tâche #004 (intégration handlers)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 — APIError struct et ErrorCode dans `pkg/apierror/errors.go`

Créer le package `pkg/apierror/` avec la struct `APIError` et les codes d'erreur standardisés.

**Raison** : Package dédié, réutilisable par toutes les couches (middleware, handlers). `pkg/` peut importer `internal/` sans problème dans le même module Go.

```go
package apierror

type ErrorCode string

const (
    ErrCodeInternal    ErrorCode = "internal_error"
    ErrCodeNotFound    ErrorCode = "not_found"
    ErrCodeBadRequest  ErrorCode = "bad_request"
    ErrCodeUnauthorized ErrorCode = "unauthorized"
    ErrCodeForbidden   ErrorCode = "forbidden"
    ErrCodeConflict    ErrorCode = "conflict"
    ErrCodeValidation  ErrorCode = "validation_error"
    ErrCodeDiscordAPI  ErrorCode = "discord_api_error"
    ErrCodeBadGateway  ErrorCode = "bad_gateway"
    ErrCodeRateLimited ErrorCode = "rate_limited"
)

type APIError struct {
    Code     ErrorCode `json:"error"`
    Message  string    `json:"message"`
    HTTPCode int       `json:"-"`
    Err      error     `json:"-"`
}

func (e *APIError) Error() string
func (e *APIError) Unwrap() error
```

**Cas nominaux :**
- Création d'une APIError via constructeurs : `NewAPIError(code, message, httpCode)`, `NewAPIErrorWithCause(code, message, httpCode, cause)`
- Helper `RespondWithError(c *gin.Context, apiErr *APIError)` qui appelle `c.JSON(apiErr.HTTPCode, apiErr)` puis `c.Abort()`
- Helper `RespondWithSuccess(c *gin.Context, status int, data gin.H)` pour wrapper les réponses success

**Cas limites :**
- `APIError` avec `Err` nil → `Unwrap()` retourne nil
- `RespondWithSuccess` avec data nil → `c.JSON(status, gin.H{})`

#### Partie 2 — Ajout des erreurs manquantes dans `internal/models/errors.go`

Ajouter les sentinelles suivantes :

```go
var ErrDiscordUnavailable   = errors.New("discord API is unavailable")
var ErrDatabaseUnavailable  = errors.New("database is unavailable")
var ErrInternalServer       = errors.New("internal server error")
var ErrForbidden            = errors.New("forbidden")
var ErrUnauthorized         = errors.New("unauthorized")
var ErrConflict             = errors.New("resource conflict")
var ErrValidation           = errors.New("validation error")
var ErrBadRequest           = errors.New("bad request")
var ErrRateLimited          = errors.New("too many requests")
```

#### Partie 3 — Mapping Erreur → Statut HTTP + Code API dans `internal/middleware/errors.go`

Créer `internal/middleware/errors.go` avec `MapError` qui importe `pkg/apierror` et `internal/models`.

Table de mapping complète :

| Erreur Go | HTTP Code | ErrorCode |
|-----------|-----------|-----------|
| ErrUserNotFound | 404 | not_found |
| ErrTeamNotFound | 404 | not_found |
| ErrMapNotFound | 404 | not_found |
| ErrSplitNotFound | 404 | not_found |
| ErrZoneNotFound | 404 | not_found |
| ErrInvitationNotFound | 404 | not_found |
| ErrJoinRequestNotFound | 404 | not_found |
| ErrTemplateNotFound | 404 | not_found |
| ErrVersionNotFound | 404 | not_found |
| ErrMemberNotFound | 404 | not_found |
| ErrPermissionNotFound | 404 | not_found |
| ErrLockNotFound | 404 | not_found |
| ErrPolygonNotFound | 404 | not_found |
| ErrInvalidDiscordID | 400 | bad_request |
| ErrTeamNameRequired | 400 | bad_request |
| ErrInvalidTeamName | 400 | bad_request |
| ErrTeamNameTooLong | 400 | bad_request |
| ErrMapNameRequired | 400 | bad_request |
| ErrZoneNameRequired | 400 | bad_request |
| ErrZoneNameTooLong | 400 | bad_request |
| ErrZoneNameTaken | 409 | conflict |
| ErrPolygonHasTooFewPoints | 400 | bad_request |
| ErrNotTeamMember | 403 | forbidden |
| ErrNotCaptain | 403 | forbidden |
| ErrInvalidRole | 400 | bad_request |
| ErrTeamAlreadyHasCocaptain | 409 | conflict |
| ErrDuplicateTeamMember | 409 | conflict |
| ErrNotMember | 403 | forbidden |
| ErrInvitationExpired | 410 | bad_request |
| ErrInvitationAlreadyUsed | 409 | conflict |
| ErrCannotInviteSelf | 400 | bad_request |
| ErrJoinRequestAlreadyExists | 409 | conflict |
| ErrJoinRequestAlreadyMember | 409 | conflict |
| ErrRefreshTokenNotFound | 401 | unauthorized |
| ErrRefreshTokenExpired | 401 | unauthorized |
| ErrRefreshTokenRevoked | 401 | unauthorized |
| ErrLockAlreadyAcquired | 409 | conflict |
| ErrLockNotOwned | 403 | forbidden |
| ErrTemplateAlreadyExists | 409 | conflict |
| ErrSplitMapNamesRequired | 400 | bad_request |
| ErrSplitNameRequired | 400 | bad_request |
| ErrDiscordUnavailable | 502 | discord_api_error |
| ErrDatabaseUnavailable | 503 | internal_error |
| ErrInternalServer | 500 | internal_error |
| ErrForbidden | 403 | forbidden |
| ErrUnauthorized | 401 | unauthorized |
| ErrConflict | 409 | conflict |
| ErrValidation | 422 | validation_error |
| ErrBadRequest | 400 | bad_request |
| ErrRateLimited | 429 | rate_limited |

**Gestion d'erreurs :**
- Erreur inconnue (non mappée) → 500, `internal_error`
- Nil → retourne nil

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `pkg/apierror/errors.go` | Créer | APIError struct, ErrorCode const, helpers (NewAPIError, RespondWithError, RespondWithSuccess) |
| `internal/models/errors.go` | Modifier | Ajouter les erreurs sentinelles manquantes |
| `internal/middleware/errors.go` | Créer | MapError, MapErrorWithMessage (importe pkg/apierror et internal/models) |

### Signatures

```go
package apierror

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type ErrorCode string

const (
    ErrCodeInternal    ErrorCode = "internal_error"
    ErrCodeNotFound    ErrorCode = "not_found"
    ErrCodeBadRequest  ErrorCode = "bad_request"
    ErrCodeUnauthorized ErrorCode = "unauthorized"
    ErrCodeForbidden   ErrorCode = "forbidden"
    ErrCodeConflict    ErrorCode = "conflict"
    ErrCodeValidation  ErrorCode = "validation_error"
    ErrCodeDiscordAPI  ErrorCode = "discord_api_error"
    ErrCodeBadGateway  ErrorCode = "bad_gateway"
    ErrCodeRateLimited ErrorCode = "rate_limited"
)

type APIError struct {
    Code     ErrorCode `json:"error"`
    Message  string    `json:"message"`
    HTTPCode int       `json:"-"`
    Err      error     `json:"-"`
}

func NewAPIError(code ErrorCode, message string, httpCode int) *APIError
func NewAPIErrorWithCause(code ErrorCode, message string, httpCode int, cause error) *APIError
func (e *APIError) Error() string
func (e *APIError) Unwrap() error
func RespondWithError(c *gin.Context, apiErr *APIError)
func RespondWithSuccess(c *gin.Context, status int, data gin.H)
```

```go
package middleware

import (
    "eva_trainer/pkg/apierror"
    "eva_trainer/internal/models"
)

func MapError(err error) *apierror.APIError
func MapErrorWithMessage(err error, message string) *apierror.APIError
```

### Contraintes techniques

- **`pkg/apierror/`** : Ne doit PAS importer `internal/` (respect des conventions Go pour `pkg/`). Les helpers sont purement techniques : struct + sérialisation JSON.
- **`internal/middleware/errors.go`** : C'est ici que vit `MapError` car il a besoin d'importer les sentinelles de `internal/models/`.
- **Pattern** : Suivre les conventions Go pour `pkg/` : package nommé `apierror`, pas de sous-packages inutiles.
- **MapError** : Utiliser `errors.Is()` pour comparer les sentinelles (supporte le wrapping).
- **RespondWithError** : Appelle `c.JSON(apiErr.HTTPCode, apiErr)` puis `c.Abort()`.
- **Compatibilité** : Le format `{"error": "...", "message": "..."}` doit correspondre exactement au format déjà utilisé dans le code existant pour éviter une rupture API.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
