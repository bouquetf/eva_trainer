# Gestion des erreurs API

## Format standard

Toutes les erreurs API retournent une réponse JSON avec la structure suivante :

```json
{
    "error": "error_code",
    "message": "Description claire et concise du problème"
}
```

## Codes HTTP

| Statut | Signification |
|--------|---------------|
| 400 | Bad Request — paramètre invalide ou manquant |
| 401 | Unauthorized — authentification requise ou token invalide |
| 403 | Forbidden — permissions insuffisantes pour cette action |
| 404 | Not Found — ressource introuvable |
| 409 | Conflict — conflit avec l'état actuel de la ressource |
| 410 | Gone — ressource expirée (ex: invitation) |
| 422 | Validation Error — données invalides (format, contraintes) |
| 429 | Rate Limited — trop de requêtes |
| 500 | Internal Server Error — erreur serveur inattendue |
| 502 | Bad Gateway — erreur d'un service externe (ex: API Discord) |
| 503 | Service Unavailable — service temporairement indisponible |

## Codes d'erreur

| ErrorCode | Signification |
|-----------|---------------|
| internal_error | Erreur interne inattendue |
| not_found | Ressource demandée introuvable |
| bad_request | Requête mal formée ou paramètre invalide |
| unauthorized | Authentification requise ou token invalide/expiré |
| forbidden | Accès refusé (permissions insuffisantes) |
| conflict | Conflit avec l'état actuel (doublon, verrou) |
| validation_error | Données ne respectant pas les contraintes |
| discord_api_error | Erreur de l'API Discord (indisponible, timeout) |
| bad_gateway | Erreur d'une dépendance externe |
| rate_limited | Trop de requêtes (à implémenter) |

## Utilisation dans le code

```go
import (
    "eva_trainer/pkg/apierror"
    "eva_trainer/internal/middleware"
)

// Réponse simple
apierror.RespondWithError(c, apierror.NewAPIError(
    apierror.ErrCodeNotFound, "team not found", 404,
))

// À partir d'une erreur modèle
apierror.RespondWithError(c, middleware.MapError(models.ErrTeamNotFound))

// Avec message personnalisé
apierror.RespondWithError(c, middleware.MapErrorWithMessage(
    models.ErrTeamNotFound, "cette équipe n'existe pas",
))

// Succès
apierror.RespondWithSuccess(c, 200, gin.H{"id": 42})
```

## Architecture

- `pkg/apierror/errors.go` — Struct APIError, ErrorCode, helpers (NewAPIError, RespondWithError, RespondWithSuccess)
- `internal/middleware/errors.go` — MapError, MapErrorWithMessage (mapping erreurs Go → APIError)
