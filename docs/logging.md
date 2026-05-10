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
