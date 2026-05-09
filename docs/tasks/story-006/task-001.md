# Tâche #001 - Story #006 : Migration base de données et nouvelles erreurs

## Objectif
Créer la migration SQL pour les tables `team_invitations` et `team_join_requests`, ajouter l'interface `FindByDisplayName` au `UserRepository`, et définir les nouveaux codes d'erreur.

## Contexte
- Story #006 : Team Invitation and Join System
- Dépend de : Aucune (tâche fondation)
- Nécessaire pour : Tâche #002 (Modèles)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer les deux nouvelles tables suivantes dans une migration SQL :

**Table `team_invitations` :** Stocke les invitations créées par les capitaines/cocapitaines, que ce soit par pseudo ou par lien d'invitation.

**Table `team_join_requests` :** Stocke les demandes d'adhésion faites par les utilisateurs.

Ajouter la méthode `FindByDisplayNameOrDiscordUsername` au `UserRepository` (recherche insensible à la casse sur `display_name` ET `discord_username` avec `OR`, retourne le premier utilisateur correspondant). Ajouter aussi `SearchUsers(query string, limit int)` pour l'autocomplétion (retourne une liste d'utilisateurs dont le `display_name` ou `discord_username` correspond à la query, limité à `limit` résultats).

Ajouter les nouvelles erreurs dans `models/errors.go`.

### Cas nominaux
- Migration exécutée : les deux tables sont créées avec leurs contraintes
- `FindByDisplayNameOrDiscordUsername("MonPseudo")` retourne l'utilisateur correspondant (match sur display_name)
- `FindByDisplayNameOrDiscordUsername("DiscordUser#1234")` retourne l'utilisateur (match sur discord_username)
- `SearchUsers("Mon", 5)` retourne jusqu'à 5 utilisateurs dont le nom contient "Mon"
- `FindByDisplayNameOrDiscordUsername("inexistant")` retourne `ErrUserNotFound`

### Cas limites
- `FindByDisplayNameOrDiscordUsername` avec chaîne vide → `ErrUserNotFound`
- `SearchUsers` avec chaîne vide → liste vide
- Unicité sur `(team_id, user_id)` dans `team_join_requests` empêche les doublons (UNIQUE sur toutes les demandes, pas seulement pending)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000007_add_team_invitations_and_join_requests.up.sql` | Créer | Migration up pour les nouvelles tables |
| `db/migrations/000007_add_team_invitations_and_join_requests.down.sql` | Créer | Migration down |
| `internal/models/errors.go` | Modifier | Ajouter les nouvelles erreurs |
| `internal/models/repository.go` | Modifier | Ajouter `FindByDisplayNameOrDiscordUsername` et `SearchUsers` à `UserRepository` |
| `internal/models/user.go` | Modifier | Implémenter `FindByDisplayNameOrDiscordUsername` et `SearchUsers` |

### Détail de la migration SQL

```sql
-- 000007_add_team_invitations_and_join_requests.up.sql

CREATE TABLE IF NOT EXISTS team_invitations (
    id SERIAL PRIMARY KEY,
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE NOT NULL,
    invited_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    invited_by_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    token VARCHAR(255) UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'declined', 'expired', 'cancelled')),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS team_join_requests (
    id SERIAL PRIMARY KEY,
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled')),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_invitations_token ON team_invitations(token);
CREATE INDEX IF NOT EXISTS idx_invitations_team ON team_invitations(team_id);
CREATE INDEX IF NOT EXISTS idx_invitations_invited_user ON team_invitations(invited_user_id);
CREATE INDEX IF NOT EXISTS idx_join_requests_team ON team_join_requests(team_id);
CREATE INDEX IF NOT EXISTS idx_join_requests_user ON team_join_requests(user_id);
```

```sql
-- 000007_add_team_invitations_and_join_requests.down.sql

DROP TABLE IF EXISTS team_join_requests;
DROP TABLE IF EXISTS team_invitations;
```

### Nouvelles erreurs à ajouter dans `errors.go`

```go
// Invitation errors
var (
    ErrInvitationNotFound = errors.New("invitation not found")
    ErrInvitationExpired  = errors.New("invitation has expired")
    ErrInvitationAlreadyUsed = errors.New("invitation has already been used")
    ErrCannotInviteSelf  = errors.New("cannot invite yourself")
)

// Join request errors
var (
    ErrJoinRequestNotFound    = errors.New("join request not found")
    ErrJoinRequestAlreadyExists = errors.New("you already have a pending join request for this team")
    ErrJoinRequestAlreadyMember = errors.New("you are already a member of this team")
)
```

### Signatures à ajouter dans `repository.go` (interface `UserRepository`)

```go
FindByDisplayNameOrDiscordUsername(query string) (*User, error)
SearchUsers(query string, limit int) ([]*User, error)
```

### Implémentations dans `user.go`

```go
func (r *userRepo) FindByDisplayNameOrDiscordUsername(query string) (*User, error) {
    // Recherche ILIKE sur display_name OR discord_username
    // Si query est vide, retourner ErrUserNotFound
    // Si aucun utilisateur trouvé, retourner ErrUserNotFound
    // Si plusieurs correspondent, retourner le premier trouvé
}

func (r *userRepo) SearchUsers(query string, limit int) ([]*User, error) {
    // Recherche ILIKE sur display_name OR discord_username avec LIMIT
    // Si query est vide, retourner liste vide
    // Retourner []*User{} (liste vide non-nil) si aucun résultat
    // Limiter à `limit` résultats (max 10 recommandé)
}
```

### Contraintes techniques
- **Framework** : Go standard + lib/pq pour PostgreSQL
- **Pattern** : Suivre le pattern exact des migrations existantes (fichiers .up.sql et .down.sql)
- **Style** : Même style que les autres repos (QueryRow, sql.ErrNoRows, etc.)
- **Convention** : Les migrations sont numérotées séquentiellement (000007_...)

### Tests à implémenter

#### Tests d'intégration
- **Fichier** : `internal/models/user_test.go`
- Scénario 1 : Recherche par display_name existant
  - Données : Créer un user avec display_name "TestUser"
  - Résultat attendu : `FindByDisplayNameOrDiscordUsername("testuser")` retourne l'utilisateur (ILIKE ignore la casse)
- Scénario 2 : Recherche par discord_username existant
  - Données : Créer un user avec discord_username "MonDiscord#1234"
  - Résultat attendu : `FindByDisplayNameOrDiscordUsername("MonDiscord#1234")` retourne l'utilisateur
- Scénario 3 : Recherche inexistante
  - Résultat attendu : `ErrUserNotFound`
- Scénario 4 : Recherche avec chaîne vide
  - Résultat attendu : `ErrUserNotFound`
- Scénario 5 : `SearchUsers("test", 5)` retourne jusqu'à 5 utilisateurs
- Scénario 6 : `SearchUsers("", 5)` retourne liste vide

### Documentation
- Aucune documentation spécifique pour cette tâche (tâche technique)
