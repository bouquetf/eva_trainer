# Tâche #005 - Story #000 : Tests d'intégration et E2E

## Objectif

Implémenter les tests d'intégration (connexion DB, migrations) et les tests E2E (docker compose, endpoint health) qui valident le fonctionnement de l'ensemble de l'ossature du projet.

## Contexte

- Story #000 : `docs/stories/story-000.md` - Project Initialization and Scaffolding
- Dépend de : Tâche #001 (Backend Go+Gin), Tâche #002 (Database+Migration), Tâche #003 (Docker), Tâche #004 (Frontend)
- Nécessaire pour : Tâche #006 (Documentation), Tâche #007 (Validation)
- Ordre d'exécution : 5ème (après #004)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer les tests qui valident l'ensemble de l'infrastructure : connexion à PostgreSQL, exécution des migrations, démarrage du serveur, endpoint de santé, et vérification E2E via Docker Compose.

**Cas nominaux :**
- Test d'intégration DB : connexion réussie à PostgreSQL avec `database.Connect()` et `database.HealthCheck()`
- Test de migration : `database.RunMigrations()` exécute les fichiers SQL sans erreur
- Test E2E : `docker compose up` → `curl http://localhost:8080/health` → `{"status": "healthy"}`
- Commande Go `cmd/migrate/main.go` exécute les migrations en ligne de commande

**Cas limites :**
- Connexion avec des identifiants invalides → erreur retournée proprement
- Migration déjà appliquée → ignorée silencieusement
- Exécution du down migration → tables supprimées dans le bon ordre

**Gestion d'erreurs :**
- Les tests d'intégration doivent être identifiables (tag `//go:build integration` ou fichier séparé) pour ne pas s'exécuter avec les tests unitaires
- Les tests E2E doivent être documentés (script bash ou fichier séparé) car ils nécessitent Docker

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `cmd/migrate/main.go` | Créer | Commande Go pour exécuter les migrations |
| `pkg/database/database_test.go` | Modifier | Ajouter test d'intégration DB (tag `integration`) |
| `pkg/database/migrations_test.go` | Créer | Tests de validation des migrations (tag `integration`) |
| `scripts/test-e2e.sh` | Créer | Script bash de test E2E via Docker Compose |
| `scripts/test.sh` | Créer | Script de lancement de tous les tests |

### Signatures

```go
// cmd/migrate/main.go
package main
func main()
// Charge la config depuis l'environnement, connecte la DB, exécute les migrations.
// Usage : go run cmd/migrate/main.go
// En cas d'erreur, log.Fatalf avec le message d'erreur.
```

### Contraintes techniques

- **Tags de build** : Les fichiers de test d'intégration doivent utiliser `//go:build integration` pour ne pas s'exécuter avec `go test ./...` simple. L'utilisateur devra faire `go test --tags=integration ./...` pour les inclure.
- **Script test.sh** :
  ```bash
  #!/bin/bash
  set -e
  echo "=== Unit Tests ==="
  go test ./...
  echo "=== Integration Tests ==="
  go test --tags=integration ./...
  echo "=== All tests passed ==="
  ```
- **Script test-e2e.sh** :
  ```bash
  #!/bin/bash
  set -e
  echo "=== E2E Tests ==="
  docker compose up -d
  sleep 5
  curl -f http://localhost:8080/health || (docker compose logs app && exit 1)
  echo "=== E2E Tests Passed ==="
  docker compose down
  ```
- **Commande migrate** : `cmd/migrate/main.go` est un binaire séparé, pas une modification de `cmd/server/main.go`. Il appelle `database.LoadConfigFromEnv()`, `database.Connect()`, puis `database.RunMigrations()`.
- **Ne PAS** modifier `cmd/server/main.go` ou `api/routes.go`
- **Ne PAS** inclure de tests de performance ou de charge

### Tests à implémenter

#### Tests d'intégration (tag `integration`)

- **Fichier** : `pkg/database/database_test.go` (ajout)
- Scénario 1 : Connexion DB avec DATABASE_URL valide
  - Données : DATABASE_URL pointant vers PostgreSQL
  - Résultat attendu : `*sql.DB` non nul, `HealthCheck` retourne nil
- Scénario 2 : Connexion avec URL invalide
  - Données : `DATABASE_URL=postgres://invalid:invalid@localhost:9999/nonexistent`
  - Résultat attendu : Erreur retournée

- **Fichier** : `pkg/database/migrations_test.go` (nouveau)
- Scénario 1 : Exécution de `RunMigrations` sur une DB vide
  - Résultat attendu : Toutes les tables créées, pas d'erreur
- Scénario 2 : Exécution de `RunMigrations` une seconde fois (idempotence)
  - Résultat attendu : Pas d'erreur, les tables existent déjà
- Scénario 3 : Exécution du down puis du up
  - Résultat attendu : Tables supprimées puis recréées

#### Tests E2E (script shell)

- **Fichier** : `scripts/test-e2e.sh`
- Scénario 1 : Démarrage complet via docker compose
  - Résultat attendu : `curl http://localhost:8080/health` retourne 200
- Scénario 2 : Frontend accessible
  - Résultat attendu : `curl http://localhost:5173` retourne 200

### Exemples d'utilisation

```bash
# Tests unitaires uniquement
go test ./...

# Tests unitaires + intégration
go test --tags=integration ./...

# Tout les tests via script
bash scripts/test.sh

# Tests E2E (nécessite Docker)
bash scripts/test-e2e.sh

# Exécuter les migrations manuellement
go run cmd/migrate/main.go
```
