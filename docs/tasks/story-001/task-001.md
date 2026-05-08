# Tâche #001 - Story #001 : Système de configuration centralisé

## Objectif

Créer un système de configuration applicative centralisé qui charge `.env` (via godotenv), fusionne avec les variables système (priorité aux variables système), valide les champs requis, et fournit une config structurée à l'ensemble de l'application.

## Contexte

- Story #001 : `docs/stories/story-001.md` - Configuration and Deployment
- Dépend de : Story #000 (ossature du projet, terminée)
- Nécessaire pour : Tâche #002 de cette story (CORS + static serving), Story #002 (Auth Discord)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un package `pkg/config` qui centralise toute la configuration de l'application. Déplacer le point d'entrée Go de `cmd/server/main.go` vers `main.go` à la racine du projet (pour permettre `//go:embed frontend/dist/*`).

**Cas nominaux :**
- Chargement d'un fichier `.env` s'il existe dans le répertoire de travail
- Les variables d'environnement système prennent priorité sur celles du `.env`
- Si aucun `.env` n'existe, utilisation exclusive des variables système
- Validation des champs requis au démarrage : si un champ requis est manquant, une erreur est retournée et le serveur s'arrête
- Structuration de la configuration en sous-groupes (Server, Database, Discord, JWT, CORS)

**Cas limites :**
- `.env` mal formé → erreur explicative
- Variables d'environnement avec des espaces ou caractères spéciaux → parsing correct
- Valeurs par défaut documentées pour les champs optionnels (PORT, GIN_MODE, DB_TIMEOUT, etc.)
- `DATABASE_URL` prioritaire sur les vars DB individuelles (comportement existant dans `pkg/database/config.go`)

**Gestion d'erreurs :**
- Champs requis manquants (JWT_SECRET, DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET) → erreur avec message listant tous les champs manquants
- Échec de chargement du `.env` → erreur (sauf si le fichier n'existe pas, ce qui est normal)
- Type invalide (ex: timeout négatif) → erreur explicative

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `go.mod` | Modifier | Ajouter `github.com/joho/godotenv` |
| `pkg/config/config.go` | Créer | Package de configuration centralisée |
| `pkg/config/config_test.go` | Créer | Tests unitaires |
| `main.go` | Créer | Nouveau point d'entrée principal à la racine du projet (utilise `pkg/config`, appelle `api.SetupRoutes`) |
| `cmd/server/main.go` | Supprimer | Contenu migré vers `main.go` racine |
| `cmd/migrate/main.go` | Modifier | Utiliser `pkg/config` (via `config.Load()`) |
| `pkg/database/config.go` | Aucun changement | Le type Config reste, ajouter un constructeur `NewConfigFromApp(appCfg)` |

### Signatures

```go
// pkg/config/config.go
package config

type AppConfig struct {
    Server   ServerConfig
    Database database.Config
    Discord  DiscordConfig
    JWT      JWTConfig
    CORS     CORSConfig
}

type ServerConfig struct {
    Port    string  // défaut: "8080"
    GinMode string  // défaut: "debug"
}

type DiscordConfig struct {
    ClientID     string
    ClientSecret string
}

type JWTConfig struct {
    Secret              string
    Expiry              time.Duration  // défaut: 24h
    RefreshTokenExpiry  time.Duration  // défaut: 7d
}

type CORSConfig struct {
    Origins []string
}

func Load() (*AppConfig, error)
// Charge .env si existe, lit les vars système, valide, retourne la config.
// N'est PAS une erreur si .env n'existe pas.

func (c *AppConfig) Validate() error
// Vérifie que JWT_SECRET, DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET sont non vides.
```

```go
// main.go (racine du projet)
package main

import (
    "embed"
    "log"
    "eva_trainer/api"
    "eva_trainer/pkg/config"
    "github.com/gin-gonic/gin"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    if cfg.Server.GinMode == "release" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.Default()
    api.SetupRoutes(r, &cfg.CORS, frontendFS)

    log.Printf("Server starting on port %s", cfg.Server.Port)
    if err := r.Run(":" + cfg.Server.Port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

### Contraintes techniques

- **godotenv** : `github.com/joho/godotenv` v1.5.x. `go get github.com/joho/godotenv`
- **Ordre de priorité** : 1) Variables système 2) `.env` 3) Valeurs par défaut
- **Ne PAS** supprimer `pkg/database/config.go` (le type Config et LoadConfigFromEnv sont utilisés par les tests existants). Ajouter plutôt un constructeur `NewConfigFromApp(cfg *config.AppConfig) database.Config` pour bridge
- **Validation** : Ne pas dupliquer la validation DB (la DB valide déjà sa connexion). Valider uniquement les champs applicatifs (JWT, Discord)
- **main.go racine** : Le fichier `main.go` à la racine du projet devient le point d'entrée unique. Supprimer `cmd/server/main.go`. Le binaire `migrate` reste dans `cmd/migrate/`.
- **Décision Q1** : Solution clean — `main.go` à la racine avec `//go:embed frontend/dist/*`. Pas de `..` dans le chemin embed, pas de copie intermédiaire.
- **Tests** : 100% des cas de chargement, priorité, validation, erreurs

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `pkg/config/config_test.go`
- Scénario 1 : Chargement avec `.env` uniquement
  - Setup : Créer un fichier `.env` temporaire avec PORT=3000
  - Résultat attendu : `cfg.Server.Port == "3000"`
- Scénario 2 : Priorité variable système sur `.env`
  - Setup : `.env` avec PORT=3000, variable système PORT=4000
  - Résultat attendu : `cfg.Server.Port == "4000"`
- Scénario 3 : Absence de `.env`
  - Setup : Pas de fichier `.env`, variable système PORT=5000
  - Résultat attendu : `cfg.Server.Port == "5000"` (pas d'erreur)
- Scénario 4 : Validation champs manquants
  - Setup : `JWT_SECRET`, `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET` non définis
  - Résultat attendu : Erreur listant les 3 champs manquants
- Scénario 5 : Valeurs par défaut
  - Setup : Aucune variable définie (sauf les requis via vars système)
  - Résultat attendu : PORT=8080, GIN_MODE=debug, JWT_EXPIRY=24h, REFRESH_TOKEN_EXPIRY=7d
- Scénario 6 : Désérialisation du timeout
  - Setup : DB_TIMEOUT=15
  - Résultat attendu : `cfg.Database.Timeout == 15 * time.Second`
- Scénario 7 : CORS_ORIGINS parsing
  - Setup : `CORS_ORIGINS=http://localhost:3000,http://example.com`
  - Résultat attendu : `cfg.CORS.Origins == ["http://localhost:3000", "http://example.com"]`
- Scénario 8 : CORS_ORIGINS wildcard
  - Setup : `CORS_ORIGINS=*`
  - Résultat attendu : `cfg.CORS.Origins == ["*"]`

### Documentation

- `README.md` : Mettre à jour la section "Architecture" pour refléter le nouveau `main.go` à la racine
- `.env.example` : Déjà existant, vérifier qu'il correspond à la nouvelle structuration
- `docs/configuration.md` : Créé dans la tâche #003
