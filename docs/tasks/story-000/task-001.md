# Tâche #001 - Story #000 : Backend Go + Gin avec endpoint de santé

## Objectif

Créer la structure du backend Go avec le framework Gin, les répertoires du projet, et un endpoint de santé `/health` fonctionnel.

## Contexte

- Story #000 : `docs/stories/story-000.md` - Project Initialization and Scaffolding
- Dépend de : Aucune
- Nécessaire pour : Tâche #002 (Database), Tâche #003 (Docker), Tâche #005 (Tests), Tâche #006 (README)
- Ordre d'exécution : 1ère (séquentiel)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Mettre en place l'ossature du backend Go avec Gin, prête à accueillir les futures fonctionnalités.

**Cas nominaux :**
- L'application démarre sur le port configuré (défaut 8080) via `go run cmd/server/main.go`
- Un appel GET à `/health` retourne `{"status": "healthy"}` avec un code HTTP 200
- La structure des répertoires suit les conventions du projet
- Le routeur Gin prépare un groupe `/api/v1/` pour les futurs endpoints métier (vide pour l'instant)
- Le fichier `go.mod` est mis à jour avec la version Go 1.26

**Cas limites :**
- Si la variable d'environnement `PORT` n'est pas définie, utiliser le port 8080 par défaut
- Gin doit fonctionner en mode `release` quand `GIN_MODE=release` est défini, sinon en mode debug

**Gestion d'erreurs :**
- Port déjà utilisé → Gin doit loguer l'erreur et terminer le processus avec un code non-zéro
- Erreur fatale au démarrage → `log.Fatalf` avec message explicite

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `go.mod` | Modifier | Mettre à jour `go 1.25` → `go 1.26`, ajouter `github.com/gin-gonic/gin` |
| `cmd/server/main.go` | Créer | Point d'entrée du serveur |
| `api/routes.go` | Créer | Configuration des routes (groupe `/api/v1/` préparé) |
| `internal/handlers/health.go` | Créer | Handler de santé |
| `internal/models/` | Créer | Répertoire pour les modèles de données |
| `internal/middleware/` | Créer | Répertoire pour les middlewares |
| `pkg/database/` | Créer | Répertoire pour le package database |
| `go.sum` | Généré | Verrouillage des dépendances |

### Signatures

```go
// cmd/server/main.go
package main
func main()

// api/routes.go
package api
func SetupRoutes(r *gin.Engine)

// internal/handlers/health.go
package handlers
func HealthCheck(c *gin.Context)
```

### Contraintes techniques

- **Version Go** : 1.26 (mettre à jour `go.mod` de `1.25` à `1.26`)
- **Framework** : Utiliser Gin (`github.com/gin-gonic/gin`) dernière version compatible Go 1.26
- **Module** : Le module Go s'appelle `eva_trainer` (déjà dans go.mod)
- **Routes** : `SetupRoutes` doit :
  - Définir `GET /health` → HealthCheck (endpoint d'infrastructure, pas préfixé `/api/v1/`)
  - Créer un groupe `apiV1 := r.Group("/api/v1")` vide, prêt à accueillir les routes métier futures
- **Style** : Respecter les conventions Go (gofmt, pas de warnings)
- **Gestion des erreurs** : Utiliser `log.Fatalf` pour les erreurs fatales, `c.JSON` pour les réponses API
- **Ne PAS créer** : Aucun handler autre que `/health`, aucune route métier
- **Ne PAS connecter** la base de données (ce sera fait dans une tâche ultérieure)

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/handlers/health_test.go`
- Scénario 1 : Appel à HealthCheck retourne status 200
  - Données : Requête GET simulée via Gin test context
  - Résultat attendu : Code HTTP 200, body JSON `{"status": "healthy"}`
- Scénario 2 : Vérification du Content-Type
  - Résultat attendu : `Content-Type: application/json; charset=utf-8`

#### Test de démarrage
- **Fichier** : `cmd/server/main_test.go` (optionnel)
- Scénario : Vérifier que `SetupRoutes` ne panique pas
  - Résultat attendu : Les routes sont enregistrées sans erreur

**Note :** Les tests d'intégration (connexion DB, E2E docker compose) sont traités dans la tâche #005.

### Exemples d'utilisation

```bash
# Mettre à jour la version Go dans go.mod
# (déjà fait manuellement ou via go mod edit)

# Installer Gin
go get github.com/gin-gonic/gin

# Démarrer le serveur
go run cmd/server/main.go

# Tester l'endpoint de santé
curl http://localhost:8080/health
# Réponse : {"status":"healthy"}

# Avec un port personnalisé
PORT=3000 go run cmd/server/main.go
```
