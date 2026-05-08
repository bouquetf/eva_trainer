# Tâche #001 - Story #002 : Modèle utilisateur et stockage des tokens

## Objectif

Créer le modèle de données utilisateur, la table `refresh_tokens` en base de données, et les méthodes d'accès aux données nécessaires à l'authentification Discord.

## Contexte

- Story #002 : `docs/stories/story-002.md` - User Authentication via Discord
- Dépend de : Story #001 (système de configuration, `pkg/config`, base de données)
- Nécessaire pour : Tâche #002 (backend auth), Tâche #003 (frontend auth)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Migration refresh_tokens :**
La table `users` existe déjà dans `db/migrations/000001_initial_schema.up.sql`. Créer une nouvelle migration `000002` pour ajouter la table `refresh_tokens` qui stocke les refresh tokens avec rotation.

**Partie 2 - Modèle User :**
Créer le struct `User` dans `internal/models/user.go` avec les méthodes CRUD de base :
- `FindOrCreateByDiscordID` : Cherche un utilisateur par son Discord ID, le crée s'il n'existe pas
- `FindByID` : Récupère un utilisateur par son ID
- `SetAdmin` : Promeut un utilisateur administrateur (premier utilisateur auto-promu)

**Partie 3 - Modèle RefreshToken :**
Créer le struct `RefreshToken` dans `internal/models/refresh_token.go` avec les méthodes :
- `CreateRefreshToken` : Stocke un nouveau refresh token (hashé)
- `FindRefreshTokenByHash` : Cherche un refresh token par son hash
- `RotateRefreshToken` : Révoque l'ancien token et en crée un nouveau dans une transaction atomique
- `RevokeRefreshToken` : Révoque un refresh token

**Partie 4 - Interface Repository (pour les tests) :**
Définir une interface `UserRepository` et `RefreshTokenRepository` pour permettre les tests avec mocks ET avec la vraie DB :
- L'interface est définie dans `internal/models/repository.go`
- L'implémentation concrète utilise `*sql.DB` directement
- Les tests d'intégration utilisent **testcontainers-go** pour lancer PostgreSQL

**Décisions techniques (issues des questions) :**
- **Q5** : Le state OAuth est stocké via JWT éphémère (pas dans la base) — donc pas de table supplémentaire ici
- **Q6** : Les tests DB utilisent `testcontainers-go` pour lancer PostgreSQL dans un conteneur Docker

**Cas nominaux :**
- Un utilisateur se connecte via Discord → création d'un enregistrement `users` avec son `discord_id`
- Un refresh token est stocké avec une date d'expiration configurable
- La rotation génère un nouveau token et révoque l'ancien dans la même transaction

**Cas limites :**
- Discord ID déjà existant → récupération de l'utilisateur existant (pas de doublon)
- Refresh token expiré → retourne une erreur
- Refresh token déjà révoqué → retourne une erreur
- Plusieurs tokens valides pour un même utilisateur (sessions multiples autorisées)

**Gestion d'erreurs :**
- Discord ID vide → erreur `ErrInvalidDiscordID`
- Token expiré → erreur `ErrRefreshTokenExpired`
- Token introuvable → erreur `ErrRefreshTokenNotFound`
- Token révoqué → erreur `ErrRefreshTokenRevoked`
- Connexion DB impossible → erreur remontée depuis `pkg/database`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000002_add_refresh_tokens.up.sql` | Créer | Migration pour la table refresh_tokens |
| `db/migrations/000002_add_refresh_tokens.down.sql` | Créer | Rollback de la migration |
| `internal/models/user.go` | Créer | Struct User + implémentation DB |
| `internal/models/refresh_token.go` | Créer | Struct RefreshToken + implémentation DB |
| `internal/models/repository.go` | Créer | Interfaces UserRepository / RefreshTokenRepository |
| `internal/models/user_test.go` | Créer | Tests d'intégration via testcontainers |
| `internal/models/refresh_token_test.go` | Créer | Tests d'intégration via testcontainers |
| `go.mod` | Modifier | Ajouter `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto`, `github.com/testcontainers/testcontainers-go` |

### Signatures

```go
// internal/models/repository.go
package models

type UserRepository interface {
    FindOrCreateByDiscordID(discordID string) (*User, bool, error)
    FindByID(userID int) (*User, error)
    SetAdmin(userID int) error
    IsFirstUser() (bool, error)
}

type RefreshTokenRepository interface {
    CreateRefreshToken(userID int, expiresAt time.Time) (tokenPlain string, err error)
    FindRefreshTokenByHash(tokenHash string) (*RefreshToken, error)
    RotateRefreshToken(oldTokenHash string, newExpiresAt time.Time) (newTokenPlain string, err error)
    RevokeRefreshToken(tokenHash string) error
    RevokeAllUserTokens(userID int) error
}
```

```go
// internal/models/user.go
package models

import "time"

type User struct {
    ID         int       `json:"id"`
    DiscordID  string    `json:"discord_id"`
    IsAdmin    bool      `json:"is_admin"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type userRepo struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
    return &userRepo{db: db}
}
```

```go
// internal/models/refresh_token.go
package models

import "time"

type RefreshToken struct {
    ID        int        `json:"id"`
    UserID    int        `json:"user_id"`
    TokenHash string     `json:"-"` // SHA-256 hash, jamais exposé
    ExpiresAt time.Time  `json:"expires_at"`
    CreatedAt time.Time  `json:"created_at"`
    RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type refreshTokenRepo struct {
    db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
    return &refreshTokenRepo{db: db}
}
```

### Migration SQL

```sql
-- db/migrations/000002_add_refresh_tokens.up.sql
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

```sql
-- db/migrations/000002_add_refresh_tokens.down.sql
DROP TABLE IF EXISTS refresh_tokens;
```

### Contraintes techniques

- **JWT** : `github.com/golang-jwt/jwt/v5` — ajouter au go.mod. `go get github.com/golang-jwt/jwt/v5`
- **Hash** : SHA-256 pour les refresh tokens (pas de bcrypt, le token est déjà une chaîne aléatoire). Utiliser `crypto/sha256` de la stdlib.
- **Transactions** : `RotateRefreshToken` doit utiliser une transaction SQL pour garantir l'atomicité (BEGIN, révoquer ancien, créer nouveau, COMMIT).
- **Contexte DB** : Les repositories prennent `*sql.DB` dans leur constructeur (`NewUserRepository(db)` et `NewRefreshTokenRepository(db)`). Le package `pkg/database` fournit la connexion via `database.DB()`.
- **Package naming** : Le package s'appelle `models`. Import path : `eva_trainer/internal/models`.
- **Génération de token** : Utiliser `crypto/rand` pour générer 32 bytes aléatoires → encodage hex (64 chars).
- **Décision Q5** : Le state OAuth n'est pas stocké en DB. Il est encodé dans un JWT éphémère côté service (Task #002). Aucune table supplémentaire ici.
- **Décision Q6** : Les tests d'intégration utilisent `testcontainers-go` (`github.com/testcontainers/testcontainers-go` v0.35+) pour lancer PostgreSQL dans un conteneur Docker. `go get github.com/testcontainers/testcontainers-go`
- **golang.org/x/crypto** : Pas nécessaire pour le hash (sha256 stdlib suffit). Ne pas l'ajouter inutilement.

### Tests à implémenter

#### Tests d'intégration via testcontainers
- **Fichier** : `internal/models/user_test.go`
- Utiliser `testcontainers-go` avec l'image `postgres:16-alpine`
- Exécuter la migration `000001_initial_schema.up.sql` avant les tests
- Scénario 1 : Création d'un nouvel utilisateur
  - Données : `discordID = "123456789"`
  - Résultat attendu : User créé avec `discord_id = "123456789"`, `is_admin = false`, `isNew = true`
- Scénario 2 : Récupération d'un utilisateur existant
  - Données : `discordID = "123456789"` (déjà créé)
  - Résultat attendu : Retourne l'utilisateur existant sans doublon, `isNew = false`
- Scénario 3 : Premier utilisateur détecté
  - Données : Aucun utilisateur en base
  - Résultat attendu : `IsFirstUser()` retourne `true`
- Scénario 4 : Promotion admin
  - Données : Utilisateur avec ID 1
  - Résultat attendu : Après `SetAdmin(1)`, l'utilisateur a `is_admin = true`
- Scénario 5 : Recherche par ID inexistant
  - Données : `userID = 99999`
  - Résultat attendu : Erreur `ErrUserNotFound`

- **Fichier** : `internal/models/refresh_token_test.go`
- Scénario 1 : Création et recherche d'un refresh token
  - Données : `userID = 1`, validité = 7 jours
  - Résultat attendu : Token créé, `FindRefreshTokenByHash` retrouve le token avec les bonnes infos
- Scénario 2 : Rotation atomique d'un refresh token
  - Données : Ancien token valide
  - Résultat attendu : Ancien token a `revoked_at` non null, nouveau token valide et actif
- Scénario 3 : Token expiré (vérification)
  - Données : Token avec expiration dans le passé
  - Résultat attendu : Le token est retourné avec `expires_at` passé ; la validation de l'expiration est faite par le service (Task #002), pas par le repo
- Scénario 4 : Révocation de tous les tokens d'un utilisateur
  - Données : Utilisateur avec 3 tokens valides
  - Résultat attendu : Les 3 tokens sont révoqués (`revoked_at` non null)
- Scénario 5 : Rotation avec token déjà révoqué
  - Données : Token révoqué
  - Résultat attendu : Erreur `ErrRefreshTokenRevoked`

### Documentation

Aucune documentation à créer pour cette tâche. La documentation d'authentification sera créée dans la Tâche #004.
