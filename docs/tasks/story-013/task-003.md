# Tâche #003 - Story #013 : Error Handling Middleware et Discord API Enhancement

## Objectif

Créer les middlewares de gestion d'erreurs (recovery, gestion centralisée) et améliorer la gestion de l'erreur API Discord.

## Contexte

- Story #013 : `docs/stories/story-013.md` — Error Handling and Logging
- Dépend de : Tâche #001 (pkg/apierror, MapError), Tâche #002 (Logger, RequestLogger)
- Nécessaire pour : Tâche #004 (intégration complète dans les handlers)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 — Recovery Middleware

```go
func RecoveryMiddleware(logger services.Logger) gin.HandlerFunc
```

**Comportement :**
- Attrape les panics dans les handlers
- Logge l'erreur avec stack trace (via `debug.Stack()`) au niveau ERROR
- Utilise `apierror.RespondWithError` pour retourner une réponse JSON standard
- HTTP Status : 500, ErrorCode : `internal_error`, Message : `"internal server error"`
- **Ne pas exposer** de détails de la stack trace dans la réponse JSON

**Cas nominaux :**
- Panique avec `error` → 500, log stack trace, JSON
- Panique avec `string` ou `int` → 500, convertir en string pour le log

#### Partie 2 — Error Handler Middleware (optionnel)

```go
func ErrorHandlerMiddleware(logger services.Logger) gin.HandlerFunc
```

**Comportement :**
- Vérifie si `c.Errors` contient des erreurs après l'exécution du handler
- Si oui, loggue chaque erreur au niveau ERROR
- Si aucune réponse n'a été écrite, utilise `MapError` pour générer une réponse standard
- **Stratégie principale** : les handlers utilisent `apierror.RespondWithError` explicitement (Tâche #004)

#### Partie 3 — Route setup adapté

Modifier `SetupRoutes` dans `api/routes.go` :

**Actuellement :**
```go
r := gin.Default()
```

**Devenir :**
```go
r := gin.New()
r.Use(services.RequestLogger(logger))     // remplace gin.Logger()
r.Use(middleware.RecoveryMiddleware(logger))
r.Use(middleware.ErrorHandlerMiddleware(logger))
// CORS et autres middlewares existants inchangés
```

#### Partie 4 — Amélioration de la gestion d'erreur API Discord

Dans `internal/handlers/auth.go`, améliorer `Callback` :

**Actuellement :**
```go
c.JSON(http.StatusBadGateway, gin.H{
    "error":   "discord_api_error",
    "message": "connexion impossible",
})
```

**Devenir :**
```go
apierror.RespondWithError(c, apierror.NewAPIError(
    apierror.ErrCodeDiscordAPI,
    "connexion impossible",
    http.StatusBadGateway,
))
```

Ajouter une distinction pour différents types d'erreur Discord :
- Erreur réseau (timeout, connexion refusée) → `ErrCodeDiscordAPI` + 502 + "Discord est actuellement indisponible, veuillez réessayer"
- Erreur HTTP non-200 de Discord → `ErrCodeDiscordAPI` + 502 + message incluant le status Discord
- Erreur avec données invalides (empty user ID) → `ErrCodeDiscordAPI` + 502 + "réponse invalide de Discord"

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/middleware/recovery.go` | Créer | RecoveryMiddleware |
| `internal/middleware/errors.go` | Modifier | Ajouter ErrorHandlerMiddleware (déjà créé dans Tâche #001) |
| `api/routes.go` | Modifier | Remplacer gin.Default() par gin.New() + middlewares + RequestLogger |
| `internal/handlers/auth.go` | Modifier | Améliorer la gestion d'erreur Discord dans Callback |

### Signatures

```go
package middleware

import (
    "eva_trainer/internal/services"
    "github.com/gin-gonic/gin"
)

func RecoveryMiddleware(logger services.Logger) gin.HandlerFunc
func ErrorHandlerMiddleware(logger services.Logger) gin.HandlerFunc
```

### Contraintes techniques

- **RecoveryMiddleware** : Utiliser `debug.Stack()` pour capturer la stack trace. Logger l'erreur AVANT de répondre.
- **Ne PAS** utiliser `gin.Default()` — utiliser `gin.New()` qui ne crée aucun middleware par défaut, puis ajouter `services.RequestLogger(logger)` et les middlewares de cette tâche.
- **Compatibilité** : Vérifier qu'aucune route ne dépend du recovery HTML de Gin (toutes les routes retournent du JSON).
- **Discord error enhancement** : Ne pas casser le flux OAuth existant. Les erreurs de `GetDiscordUser` doivent continuer à retourner 502.
- **RequestLogger** : Implémenté dans la Tâche #002, simplement référencé ici dans le setup de routes.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
