# Tâche #005 - Story #013 : Tests et Documentation

## Objectif

Implémenter les tests automatisés (unitaires, intégration) et créer la documentation pour la gestion d'erreurs et le logging.

## Contexte

- Story #013 : `docs/stories/story-013.md` — Error Handling and Logging
- Dépend de : Tâche #001 (pkg/apierror, MapError), Tâche #002 (Logger), Tâche #003 (middleware), Tâche #004 (intégration)
- Ordre d'exécution : 5ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 — Tests unitaires

##### 1.1 Tests pour `pkg/apierror/errors.go` (`pkg/apierror/errors_test.go`)

```go
func TestNewAPIError(t *testing.T) {
    // Vérifie les champs Code, Message, HTTPCode, Err
}

func TestNewAPIErrorWithCause(t *testing.T) {
    // Vérifie que Unwrap() retourne la cause
}

func TestAPIError_Error(t *testing.T) {
    // Vérifie le message retourné
}

func TestAPIError_Unwrap(t *testing.T) {
    // nil si pas de cause, sinon la cause
}

func TestRespondWithError(t *testing.T) {
    // Crée un contexte Gin de test, appelle RespondWithError
    // Vérifie le status code et le body JSON
}

func TestRespondWithSuccess(t *testing.T) {
    // Même principe
}
```

##### 1.2 Tests pour `internal/middleware/errors.go` (`internal/middleware/errors_test.go`)

```go
func TestMapError(t *testing.T) {
    tests := []struct {
        name     string
        err      error
        wantCode apierror.ErrorCode
        wantHTTP int
    }{
        {"user not found", models.ErrUserNotFound, apierror.ErrCodeNotFound, 404},
        {"team not found", models.ErrTeamNotFound, apierror.ErrCodeNotFound, 404},
        {"map not found", models.ErrMapNotFound, apierror.ErrCodeNotFound, 404},
        {"zone name taken", models.ErrZoneNameTaken, apierror.ErrCodeConflict, 409},
        {"discord unavailable", models.ErrDiscordUnavailable, apierror.ErrCodeDiscordAPI, 502},
        {"rate limited", models.ErrRateLimited, apierror.ErrCodeRateLimited, 429},
        {"forbidden", models.ErrForbidden, apierror.ErrCodeForbidden, 403},
        {"unauthorized", models.ErrUnauthorized, apierror.ErrCodeUnauthorized, 401},
        {"validation", models.ErrValidation, apierror.ErrCodeValidation, 422},
        {"bad request", models.ErrBadRequest, apierror.ErrCodeBadRequest, 400},
        {"conflict", models.ErrConflict, apierror.ErrCodeConflict, 409},
        {"internal error", models.ErrInternalServer, apierror.ErrCodeInternal, 500},
        {"database unavailable", models.ErrDatabaseUnavailable, apierror.ErrCodeInternal, 503},
        {"unknown error", errors.New("unknown"), apierror.ErrCodeInternal, 500},
        {"nil error", nil, "", 0}, // MapError(nil) retourne nil
    }
    // ...
}

func TestMapErrorWithMessage(t *testing.T) {
    // Vérifie que le message personnalisé surcharge le message par défaut
}
```

##### 1.3 Tests pour le Logger (`internal/services/logger_test.go`)

```go
func TestLogLevels(t *testing.T) {
    // Capturer la sortie, vérifier que seul le bon niveau est écrit
}

func TestLogLevelFiltering(t *testing.T) {
    // Niveau WARN → Debug et Info ne sont pas écrits
}

func TestLogWithContext(t *testing.T) {
    // With() ajoute des champs contextuels
}

func TestLogFormatText(t *testing.T) {
    // Vérifier le format texte (timestamp, niveau, message, key=value)
}

func TestLogFormatJSON(t *testing.T) {
    // Vérifier le format JSON (parser chaque ligne)
}

func TestParseLogLevel(t *testing.T) {
    // "debug" → LogLevelDebug, "info" → LogLevelInfo, "invalid" → erreur
}

func TestParseLogFormat(t *testing.T) {
    // "auto" → LogFormatAuto, "json" → LogFormatJSON, "invalid" → erreur
}

func TestLoggerConcurrency(t *testing.T) {
    // Appels concurrents au logger via goroutines (pas de data race)
}

func TestRequestLogger(t *testing.T) {
    // Créer engine Gin, ajouter RequestLogger, faire requête
    // Vérifier que la sortie contient method, path, status, duration
}
```

##### 1.4 Tests pour le Recovery Middleware (`internal/middleware/recovery_test.go`)

```go
func TestRecoveryMiddleware(t *testing.T) {
    // Créer un engine Gin, ajouter RecoveryMiddleware avec un buffer logger
    // Ajouter un handler qui panique
    // Faire une requête HTTP de test
    // Vérifier : status 500, body JSON { "error": "internal_error", "message": "..." }
    // Vérifier : le log contient la stack trace
}
```

#### Partie 2 — Tests d'intégration

##### 2.1 Test : Simulation d'erreur → vérification format réponse

Dans `api/routes_test.go` ou un fichier dédié (marqué avec `//go:build integration`) :

```go
func TestErrorResponseFormat(t *testing.T) {
    // Appeler un endpoint avec token invalide
    // Vérifier que la réponse est JSON avec les champs "error" et "message"
    // Vérifier le code HTTP 401
}

func TestDiscordAPIErrorResponse(t *testing.T) {
    // Simuler une erreur Discord API
    // Vérifier 502 Bad Gateway
    // Vérifier le format JSON { "error": "discord_api_error", "message": "..." }
}

func TestForbiddenResponseFormat(t *testing.T) {
    // Appeler un endpoint sans permissions
    // Vérifier 403
    // Vérifier { "error": "forbidden", "message": "..." }
}

func TestNotFoundResponseFormat(t *testing.T) {
    // Appeler un endpoint avec ID inexistant
    // Vérifier 404
    // Vérifier { "error": "not_found", "message": "..." }
}
```

##### 2.2 Test : Logs générés (niveau approprié)

```go
func TestLogLevelGeneration(t *testing.T) {
    // Capturer la sortie du logger (via un bufio.Writer)
    // Déclencher une erreur connue (ex: 404)
    // Vérifier qu'un log de niveau ERROR ou WARN est émis
}
```

#### Partie 3 — Documentation

##### `docs/error-handling.md`

```markdown
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
```

##### `docs/logging.md`

```markdown
# Logging

## Niveaux de log

| Niveau | Usage |
|--------|-------|
| DEBUG | Informations détaillées pour le débogage en développement |
| INFO  | Événements normaux de l'application (requêtes, démarrage) |
| WARN  | Situations anormales mais non bloquantes |
| ERROR | Erreurs nécessitant une intervention |

## Configuration

Variables d'environnement :
- `LOG_LEVEL` : `debug`, `info`, `warn`, `error` (défaut: `info`)
- `LOG_FORMAT` : `auto`, `json`, `text` (défaut: `auto`)
  - `auto` → `json` en production (`GIN_MODE=release`), `text` en développement

En production (Clever Cloud), les logs sont écrits sur **stdout** en format JSON. Clever Cloud capture automatiquement stdout et le rend disponible dans l'interface de logs.

## Structure des logs

### Développement (format texte)
```
2026/05/10 12:00:00 [INFO] server starting port=8080
2026/05/10 12:00:01 [ERROR] failed to connect error="connection refused" retry=3
```

### Production (format JSON)
```json
{"time":"2026-05-10T12:00:00Z","level":"info","message":"server starting","port":"8080"}
{"time":"2026-05-10T12:00:01Z","level":"error","message":"failed to connect","error":"connection refused","retry":3}
```

### Logs de requêtes HTTP
Les requêtes HTTP sont loggées automatiquement par le middleware `RequestLogger` :
```
INFO request method=GET path=/api/v1/teams status=200 duration=12ms client_ip=127.0.0.1
```

## Utilisation dans le code

```go
import "eva_trainer/internal/services"

type MyHandler struct {
    // ...
    logger services.Logger
}

func (h *MyHandler) DoSomething() {
    h.logger.Info("processing request", "user_id", 42)
    
    h.logger.With("component", "myhandler").Debug("details", "key", "value")
    
    if err != nil {
        h.logger.Error("operation failed", "error", err)
    }
}
```

## Utilisation des champs contextuels

```go
// Avec contexte (recommendé pour les logs métier)
logger.With("user_id", user.ID, "team_id", teamID).Info("user joined team")

// Simple
logger.Warn("rate limit approaching", "current", 95, "limit", 100)
```

## Sauvegarde de la base de données

La stratégie de sauvegarde de la base de données PostgreSQL est gérée par **Clever Cloud** via leur service add-on PostgreSQL.
Consulter la documentation Clever Cloud pour les détails :
- Fréquence des sauvegardes automatiques
- Politique de rétention
- Procédure de restauration en cas d'incident
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `pkg/apierror/errors_test.go` | Créer | Tests unitaires APIError, NewAPIError, RespondWithError |
| `internal/middleware/errors_test.go` | Créer | Tests unitaires MapError, MapErrorWithMessage |
| `internal/services/logger_test.go` | Créer | Tests unitaires Logger + RequestLogger |
| `internal/middleware/recovery_test.go` | Créer | Tests du recovery middleware |
| `api/routes_test.go` | Modifier | Tests d'intégration format réponse erreur |
| `docs/error-handling.md` | Créer | Documentation gestion d'erreurs |
| `docs/logging.md` | Créer | Documentation logging |

### Contraintes techniques

- **Tests unitaires** purs Go, sans base de données.
- **Capture de logger dans les tests** : Utiliser un `bytes.Buffer` comme output writer pour le logger de test.
- **Recovery middleware test** : Créer un engine Gin minimal, ajouter le middleware, ajouter un handler qui panique, faire une requête HTTP de test avec `httptest.NewRecorder()`.
- **Mock RequestLogger** : Utiliser `bytes.Buffer` et vérifier la présence des champs `method`, `path`, `status`.
- **Documentation en français** (comme le reste du projet).
- **Vérifier que tous les tests passent** : `go test ./...` avant de finaliser.

### Exemples de vérification

```bash
go test ./pkg/apierror/ -v
go test ./internal/middleware/ -run "TestMapError|TestRecovery" -v
go test ./internal/services/ -run "TestLog" -v
go test ./... -count=1
```
