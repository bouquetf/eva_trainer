# Tâche #002 - Story #013 : Service de Logging Structuré

## Objectif

Créer un service de logging structuré avec niveaux (debug, info, warn, error), configuration par variable d'environnement et compatibilité Clever Cloud (sortie stdout, format JSON en production).

## Contexte

- Story #013 : `docs/stories/story-013.md` — Error Handling and Logging
- Dépend de : Aucune (indépendant de l'infrastructure d'erreurs)
- Nécessaire pour : Tâche #003 (middleware), Tâche #004 (intégration handlers)
- Ordre d'exécution : 2ème (parallélisable avec Tâche #001)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 — Logger interface et implémentation

Créer un service de logging structuré dans `internal/services/logger.go` :

```go
type LogLevel int

const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelWarn
    LogLevelError
)

type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    With(args ...any) Logger
    SetLevel(level LogLevel)
    Level() LogLevel
}
```

**Cas nominaux :**
- `Debug` : message uniquement si le niveau configuré ≤ DEBUG
- `Info` : message si le niveau configuré ≤ INFO
- `Warn` : message si le niveau configuré ≤ WARN
- `Error` : message si le niveau configuré ≤ ERROR
- `With` : retourne un nouveau Logger avec des champs contextuels pré-attachés (ex: `logger.With("user_id", 42).Info("action", "login")`)
- Format production (GIN_MODE=release) : JSON structuré (un objet JSON par ligne) — compatible avec les outils d'analyse Clever Cloud
- Format développement : texte lisible par un humain (timestamp [LEVEL] message key=value)
- Sortie : stdout pour tout niveau (Clever Cloud capture stdout)

**Format JSON (production) :**
```json
{"time":"2026-05-10T12:00:00Z","level":"info","message":"server starting","port":"8080"}
{"time":"2026-05-10T12:00:01Z","level":"error","message":"failed to connect","error":"connection refused","retry":3}
```

**Format texte (développement) :**
```
2026/05/10 12:00:00 [INFO] server starting port=8080
2026/05/10 12:00:01 [ERROR] failed to connect error="connection refused" retry=3
```

#### Partie 2 — RequestLogger (middleware Gin)

Créer un middleware `RequestLogger` qui **remplace** `gin.Logger()` :

```go
func RequestLogger(logger Logger) gin.HandlerFunc
```

**Comportement :**
- Loggue au niveau INFO chaque requête entrante : méthode, path, status, durée
- Utilise les champs structurés : `"method"`, `"path"`, `"status"`, `"duration"`, `"ip"`
- Remplace complètement `gin.Logger()` dans `api/routes.go`

**Format log :**
```
INFO request method=GET path=/api/v1/teams status=200 duration=12ms client_ip=127.0.0.1
```

**Implémentation :**
```go
func RequestLogger(logger Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        clientIP := c.ClientIP()

        c.Next()

        status := c.Writer.Status()
        duration := time.Since(start)

        logger.With(
            "method", c.Request.Method,
            "path", path,
            "status", status,
            "duration", duration.String(),
            "client_ip", clientIP,
        ).Info("request")
    }
}
```

#### Partie 3 — Configuration

- Ajouter une variable d'environnement `LOG_LEVEL` dans la config (valeurs : `debug`, `info`, `warn`, `error`, défaut : `info`)
- Ajouter une variable d'environnement `LOG_FORMAT` dans la config (valeurs : `auto`, `json`, `text`, défaut : `auto` → `json` quand `GIN_MODE=release`, `text` sinon)

#### Partie 4 — Constructeur et singleton

```go
func NewLogger(level LogLevel, format LogFormat) Logger
func NewLoggerFromConfig(cfg *config.AppConfig) Logger
```

- `NewLoggerFromConfig` lit la config et crée le logger approprié.
- Le logger est initialisé dans `main.go` et passé via `newCoreDeps` à `SetupRoutes`.

**Cas limites :**
- Niveau config invalide → fallback sur `info`
- Format config invalide → fallback sur `auto`
- Args pairs (clé/valeur) : si une clé n'a pas de valeur, la ignorer avec un warning silencieux
- Args impairs : logger le fait mais ne pas paniquer

**Gestion d'erreurs :**
- Pas de retour d'erreur pour les méthodes de log (le logging ne doit jamais bloquer l'application)
- En cas d'échec d'écriture (stdout), ignorer silencieusement

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/services/logger.go` | Créer | Logger interface + implémentation + RequestLogger |
| `pkg/config/config.go` | Modifier | Ajouter `LogConfig` avec Level et Format |
| `.env.example` | Modifier | Ajouter `LOG_LEVEL` et `LOG_FORMAT` |

### Signatures

```go
package services

type LogLevel int

const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelWarn
    LogLevelError
)

type LogFormat int

const (
    LogFormatAuto LogFormat = iota
    LogFormatText
    LogFormatJSON
)

type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    With(args ...any) Logger
    SetLevel(level LogLevel)
    Level() LogLevel
}

type loggerImpl struct { ... }

func NewLogger(level LogLevel, format LogFormat) Logger
func NewLoggerFromConfig(cfg *config.AppConfig) Logger
func ParseLogLevel(s string) (LogLevel, error)
func ParseLogFormat(s string) (LogFormat, error)

func RequestLogger(logger Logger) gin.HandlerFunc

// Méthodes privées
func (l *loggerImpl) log(level LogLevel, msg string, args ...any)
func (l *loggerImpl) isEnabled(level LogLevel) bool
func (l *loggerImpl) formatText(level LogLevel, msg string, args ...any) string
func (l *loggerImpl) formatJSON(level LogLevel, msg string, args ...any) string
func (l *loggerImpl) write(line string)
```

```go
// Ajouts dans pkg/config/config.go

type AppConfig struct {
    // ... existing fields
    Log      LogConfig
}

type LogConfig struct {
    Level  string
    Format string
}

// Dans Load()
Log: LogConfig{
    Level:  getEnv("LOG_LEVEL", "info"),
    Format: getEnv("LOG_FORMAT", "auto"),
},
```

### Contraintes techniques

- **Ne pas utiliser** de bibliothèque de logging externe (stdlib uniquement).
- **Thread-safe** : Utiliser `sync.Mutex` pour les accès concurrents au niveau de log.
- **Performance** : Les appels à `Debug` en dessous du niveau configuré ne doivent pas formater le message. Vérifier `isEnabled` avant de construire la ligne.
- **Format JSON** : Utiliser `encoding/json.Marshal` pour sérialiser. Ne pas pretty-print (un objet par ligne).
- **Time format** : Utiliser `time.RFC3339Nano` pour les timestamps.
- **RequestLogger** : Remplace `gin.Logger()` — à intégrer dans `api/routes.go` via `r.Use(middleware.RequestLogger(logger))` (ne pas créer un deuxième middleware de log, un seul suffit).
- **Clés réservées** : `time`, `level`, `message` — les args écrasent les valeurs par défaut si ces clés sont utilisées.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
