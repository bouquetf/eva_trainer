# Tâche #004 - Story #013 : Intégration du Logging et des Helpers d'Erreur

## Objectif

Intégrer le logging structuré et les helpers d'erreur dans l'ensemble du codebase : injecter le logger dans tous les handlers et middlewares, remplacer `log.Printf`, et uniformiser les réponses d'erreur avec `RespondWithError`.

## Contexte

- Story #013 : `docs/stories/story-013.md` — Error Handling and Logging
- Dépend de : Tâche #001 (pkg/apierror, MapError), Tâche #002 (Logger, RequestLogger), Tâche #003 (middleware, routes)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 — Injection du Logger dans TOUS les handlers et middlewares

Ajouter un champ `Logger services.Logger` dans **tous** les structs Handler ET Middleware :

**Handlers concernés :**
- `AuthHandler` (`internal/handlers/auth.go`)
- `AdminHandler` (`internal/handlers/admin.go`)
- `MapHandler` (`internal/handlers/map.go`)
- `SplitHandler` (`internal/handlers/split.go`)
- `TeamHandler` (`internal/handlers/team.go`)
- `ZoneHandler` (`internal/handlers/zone.go`)
- `TemplateHandler` (`internal/handlers/template.go`)
- `EditLockHandler` (`internal/handlers/edit_lock.go`)
- `HistoryHandler` (`internal/handlers/history.go`)
- `MapCatalogHandler` (`internal/handlers/map_catalog.go`)
- `TeamInvitationHandler` (`internal/handlers/team_invitation.go`)
- `TeamJoinRequestHandler` (`internal/handlers/team_join_request.go`)
- `TeamRolesHandler` (`internal/handlers/team_roles.go`)

**Middlewares concernés :**
- `RequireAuth` (`internal/middleware/auth.go`)
- `RequireAdmin` (`internal/middleware/admin.go`)
- `RequirePermission` (`internal/middleware/permission.go`)
- `RequireGlobalPermission` (`internal/middleware/permission.go`)

**Modifications par handler/middleware :**
1. Ajouter `logger services.Logger` comme dernier champ de la struct
2. Ajouter `logger services.Logger` comme dernier paramètre du constructeur `NewXxx`
3. Dans le constructeur : `Xxx: logger` ou `logger: logger`
4. Mettre à jour TOUS les appels aux constructeurs dans `api/routes.go`

**Ne pas modifier :**
- Les repositories dans `internal/models/` (pas de logger dans la couche données)
- Les services dans `internal/services/` sauf pour remplacer `log.Printf`

#### Partie 2 — Uniformisation des réponses d'erreur avec `RespondWithError`

Remplacer toutes les réponses d'erreur `c.JSON(http.StatusXxx, gin.H{"error": "...", "message": "..."})` par des appels à `apierror.RespondWithError`.

**Règles de conversion :**

| Actuel | Devenir |
|--------|---------|
| `c.JSON(400, gin.H{"error":"x","message":"y"})` | `apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeBadRequest, "y", 400))` |
| `c.JSON(401, gin.H{"error":"x","message":"y"})` | `apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeUnauthorized, "y", 401))` |
| `c.JSON(403, gin.H{"error":"x","message":"y"})` | `apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeForbidden, "y", 403))` |
| `c.JSON(404, ...)` | `apierror.RespondWithError(c, middleware.MapError(models.ErrXxx))` |
| `c.JSON(409, ...)` | `apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeConflict, "y", 409))` |
| `c.JSON(500, ...)` | `apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "y", 500))` |
| `c.JSON(502, ...)` | `apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeDiscordAPI, "y", 502))` |

**Où appliquer :**
1. `internal/handlers/auth.go` — Tous les retours d'erreur
2. `internal/handlers/admin.go` — Tous les retours d'erreur
3. `internal/middleware/auth.go` — Retours unauthorized/invalid token
4. `internal/middleware/admin.go` — Retours forbidden
5. `internal/middleware/permission.go` — Retours forbidden/unauthorized
6. Tous les autres handlers (créés dans les stories #003 à #012)

**Important :** Le format JSON de sortie `{"error": "...", "message": "..."}` reste IDENTIQUE. C'est juste la manière de construire la réponse qui change.

#### Partie 3 — Remplacement de `log.Printf` par le Logger structuré

Remplacer TOUS les appels à `log.Printf` par des appels au logger structuré.

**Fichiers avec `log.Printf` :**
- `internal/handlers/zone.go` (~15 appels, préfixe `[batch]`)
- `internal/handlers/history.go` (~4 appels, préfixe `warning:`)
- `internal/services/template_sync.go` (~1 appel)
- `internal/services/lock_cleanup.go` (~2 appels)

**Règles de mapping :**

| `log.Printf` actuel | Niveau logger | Arguments |
|---------------------|---------------|-----------|
| `log.Printf("warning: ...")` | `Warn` | `msg` = message sans préfixe `warning: ` |
| `log.Printf("[batch] ...")` | `Error` | `msg` = message sans préfixe `[batch] `, ajouter `"component", "batch"` |
| `log.Printf("LockCleanupService: ...")` | `Info` ou `Warn` | `msg` = message sans préfixe, ajouter `"component", "lock_cleanup"` |
| `log.Printf("failed to ...")` | `Error` | `msg` = le message d'erreur, inclure l'erreur dans les args |

**Exemples de remplacement :**
```go
// Avant
log.Printf("warning: failed to sync template after zone change: %v", err)

// Après
logger.With("component", "template_sync").Warn("failed to sync template after zone change", "error", err)
```

```go
// Avant
log.Printf("[batch] lock check failed (user=%d, team=%d, map=%d): %v", user.ID, teamID, mapID, err)

// Après
logger.With("component", "batch", "user_id", user.ID, "team_id", teamID, "map_id", mapID).
    Error("lock check failed", "error", err)
```

**Ne pas remplacer** `log.Fatalf` dans `main.go` (les erreurs de démarrage fatales doivent utiliser `log.Fatalf` standard).

#### Partie 4 — Mise à jour de `api/routes.go`

1. Ajouter `logger services.Logger` dans `coreDeps`
2. Modifier `newCoreDeps` pour accepter le logger
3. Passer le logger à chaque constructeur de handler et middleware
4. Ajouter `services.RequestLogger(logger)` dans le setup Gin (déjà référencé dans Tâche #003)

#### Partie 5 — Initialisation dans `main.go`

```go
logger := services.NewLoggerFromConfig(cfg)
```

Placer cette ligne APRÈS `config.Load()` mais AVANT la connexion à la base de données, pour pouvoir logger les erreurs de connexion DB.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/auth.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/admin.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/zone.go` | Modifier | Injecter logger + remplacer log.Printf |
| `internal/handlers/history.go` | Modifier | Injecter logger + remplacer log.Printf |
| `internal/handlers/team.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/map.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/split.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/template.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/edit_lock.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/health.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/map_catalog.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/team_invitation.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/team_join_request.go` | Modifier | Injecter logger + RespondWithError |
| `internal/handlers/team_roles.go` | Modifier | Injecter logger + RespondWithError |
| `internal/services/template_sync.go` | Modifier | Remplacer log.Printf par logger.With |
| `internal/services/lock_cleanup.go` | Modifier | Remplacer log.Printf par logger.With |
| `internal/middleware/auth.go` | Modifier | Injecter logger + RespondWithError |
| `internal/middleware/admin.go` | Modifier | Injecter logger + RespondWithError |
| `internal/middleware/permission.go` | Modifier | Injecter logger + RespondWithError |
| `internal/middleware/team.go` | Modifier | Injecter logger + RespondWithError (si applicable) |
| `api/routes.go` | Modifier | Injecter logger dans coreDeps + tous les constructeurs |
| `main.go` | Modifier | Initialiser logger, passer à SetupRoutes |

### Contraintes techniques

- **Format d'erreur inchangé** : `{"error": "...", "message": "..."}` doit rester exactement le même pour ne pas casser l'API.
- **Ne pas casser les tests** : Les tests existants qui vérifient les réponses JSON doivent continuer à passer.
- **Logger dernier paramètre** : Ajouter `logger services.Logger` comme dernier paramètre des constructeurs pour minimiser les changements de signature.
- **Logger inutilisé** : Même si un handler n'utilise pas encore le logger activement, l'injecter quand même (cohérence, prêt pour le futur).
- **Pas de logger dans `internal/models/`** : Les repositories retournent des erreurs, c'est la couche appelante qui logge.
- **Ordre d'initialisation dans main.go** : `config.Load()` → `logger := NewLoggerFromConfig(cfg)` → connexion DB → migrations → seed → serveur.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
