# Tâche #001 - Story #003 : Modèles Team et TeamMember (back-end)

## Objectif

Créer le modèle de données Team et TeamMember, leurs repositories (CRUD), et les règles de validation du nom d'équipe.

## Contexte

- Story #003 : `docs/stories/story-003.md` - Team Creation and Management
- Dépend de : Story #002 (authentification utilisateur, le captain est un `User`)
- Nécessaire pour : Tâche #002 (handlers équipe), Tâche #004 (tests)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Modèle Team :**
Créer le struct `Team` dans `internal/models/team.go`. La table `teams` existe déjà dans la migration `000001_initial_schema.up.sql`. Définir le repository avec les méthodes suivantes :

- `Create(name string, captainID int) (*Team, error)` : Crée une équipe et ajoute le créateur comme `captain` dans `team_members` dans une **même transaction**. Vérifie l'unicité du nom via la contrainte UNIQUE de la DB.
- `FindByID(id int) (*Team, error)` : Récupère une équipe par son ID.
- `FindByUserID(userID int) ([]*Team, error)` : Récupère toutes les équipes dont l'utilisateur est membre (jointure avec `team_members`).
- `UpdateName(id int, name string) error` : Met à jour le nom de l'équipe.
- `Delete(id int) error` : Supprime l'équipe. La contrainte ON DELETE CASCADE supprime les membres et les zones privées.

**Partie 2 - Modèle TeamMember :**
Créer le struct `TeamMember` dans `internal/models/team_member.go`. La table `team_members` existe déjà dans la migration `000001_initial_schema.up.sql`. Définir le repository avec :

- `AddMember(teamID, userID int, role string) error` : Ajoute un membre avec le rôle spécifié. Vérifie que le rôle est valide (`captain`, `cocaptain`, `editor`, `member`).
- `RemoveMember(teamID, userID int) error` : Retire un membre.
- `UpdateRole(teamID, userID int, role string) error` : Change le rôle d'un membre.
- `FindByTeamID(teamID int) ([]*TeamMember, error)` : Liste les membres d'une équipe.
- `FindByUserAndTeam(userID, teamID int) (*TeamMember, error)` : Récupère le rôle d'un utilisateur dans une équipe.
- `GetCaptainID(teamID int) (int, error)` : Récupère l'ID du capitaine de l'équipe.
- `CountByRole(teamID int, role string) (int, error)` : Compte les membres avec un rôle donné (utilisé pour vérifier la limite de 1 co-capitaine).

**Partie 3 - Validation du nom d'équipe :**
Créer une fonction de validation `ValidateTeamName(name string) error` dans `internal/models/team.go` :

- Longueur : entre 1 et 250 caractères.
- Caractères autorisés : lettres Unicode (avec accents : é, è, ê, ñ, ü, etc.), chiffres (0-9), espaces, tirets (`-`), underscores (`_`), apostrophes (`'`), points (`.`).
- La regex de validation est : `^[\p{L}\p{M}0-9\-_' .]+$`
- Si invalide → retourne `ErrInvalidTeamName` (à définir dans `internal/models/errors.go`).
- Si vide → retourne `ErrTeamNameRequired`.

**Partie 4 - Interface Repository :**
Étendre le fichier `internal/models/repository.go` pour ajouter les nouvelles interfaces `TeamRepository` et `TeamMemberRepository` :

- L'interface permet les tests avec mocks.
- L'implémentation concrète utilise `*sql.DB`.
- Les constructeurs : `NewTeamRepository(db *sql.DB) TeamRepository`, `NewTeamMemberRepository(db *sql.DB) TeamMemberRepository`.

**Cas nominaux :**
- Création d'une équipe → table `teams` reçoit un enregistrement, `team_members` reçoit le capitaine avec role='captain'.
- Récupération d'une équipe par ID → retourne le Team avec tous les champs.
- Mise à jour du nom → changé en base, `updated_at` mis à jour.
- Suppression → l'équipe et ses membres sont supprimés (CASCADE).

**Cas limites :**
- Nom d'équipe déjà existant → erreur remontée par la contrainte UNIQUE PostgreSQL (`pq: duplicate key value`).
- Nom d'équipe avec caractères non autorisés (`<script>`, `--`, `;`) → erreur `ErrInvalidTeamName`.
- Nom d'équipe > 250 caractères → erreur `ErrInvalidTeamName`.
- Suppression d'une équipe inexistante → erreur `ErrTeamNotFound`.
- Récupération d'une équipe inexistante → `ErrTeamNotFound`.
- Ajout d'un co-capitaine alors qu'il y en a déjà un → erreur `ErrTeamAlreadyHasCocaptain`.

**Gestion d'erreurs (ajouter dans `internal/models/errors.go`) :**

```go
var (
    ErrTeamNotFound          = errors.New("team not found")
    ErrTeamNameRequired      = errors.New("team name is required")
    ErrInvalidTeamName       = errors.New("team name contains invalid characters")
    ErrTeamNameTooLong       = errors.New("team name exceeds maximum length of 250 characters")
    ErrMemberNotFound        = errors.New("member not found in team")
    ErrInvalidRole           = errors.New("invalid team role")
    ErrTeamAlreadyHasCocaptain = errors.New("team already has a cocaptain")
    ErrNotCaptain            = errors.New("user is not the team captain")
    ErrNotMember             = errors.New("user is not a team member")
    ErrDuplicateTeamMember   = errors.New("user is already a member of this team")
)
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/team.go` | Créer | Struct Team + TeamRepository + ValidateTeamName |
| `internal/models/team_member.go` | Créer | Struct TeamMember + TeamMemberRepository |
| `internal/models/errors.go` | Modifier | Ajouter les nouvelles erreurs d'équipe |
| `internal/models/repository.go` | Modifier | Ajouter les interfaces TeamRepository et TeamMemberRepository |

### Signatures

```go
// internal/models/team.go
package models

import "time"

type Team struct {
    ID         int       `json:"id"`
    Name       string    `json:"name"`
    CaptainID  int       `json:"captain_id"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type TeamRepository interface {
    Create(name string, captainID int) (*Team, error)
    FindByID(id int) (*Team, error)
    FindByUserID(userID int) ([]*Team, error)
    UpdateName(id int, name string) error
    Delete(id int) error
}

type teamRepo struct {
    db *sql.DB
}

func NewTeamRepository(db *sql.DB) TeamRepository {
    return &teamRepo{db: db}
}

// ValidateTeamName valide le nom d'une équipe.
func ValidateTeamName(name string) error {
    if name == "" {
        return ErrTeamNameRequired
    }
    if len(name) > 250 {
        return ErrTeamNameTooLong
    }
    // Regex : lettres, chiffres, espaces, tirets, underscores, apostrophes, points
    re := regexp.MustCompile(`^[\p{L}\p{M}0-9\-_' .]+$`)
    if !re.MatchString(name) {
        return ErrInvalidTeamName
    }
    return nil
}
```

```go
// internal/models/team_member.go
package models

import "time"

type TeamMember struct {
    TeamID   int       `json:"team_id"`
    UserID   int       `json:"user_id"`
    Role     string    `json:"role"`
    JoinedAt time.Time `json:"joined_at"`
}

const (
    RoleCaptain   = "captain"
    RoleCocaptain = "cocaptain"
    RoleEditor    = "editor"
    RoleMember    = "member"
)

var ValidRoles = map[string]bool{
    RoleCaptain:   true,
    RoleCocaptain: true,
    RoleEditor:    true,
    RoleMember:    true,
}

type TeamMemberRepository interface {
    AddMember(teamID, userID int, role string) error
    RemoveMember(teamID, userID int) error
    UpdateRole(teamID, userID int, role string) error
    FindByTeamID(teamID int) ([]*TeamMember, error)
    FindByUserAndTeam(userID, teamID int) (*TeamMember, error)
    GetCaptainID(teamID int) (int, error)
    CountByRole(teamID int, role string) (int, error)
}

type teamMemberRepo struct {
    db *sql.DB
}

func NewTeamMemberRepository(db *sql.DB) TeamMemberRepository {
    return &teamMemberRepo{db: db}
}
```

### Méthodes détaillées pour TeamRepository

```go
// Create crée une équipe et ajoute le créateur comme capitaine dans une transaction.
// Retourne l'équipe créée.
func (r *teamRepo) Create(name string, captainID int) (*Team, error) {
    // 1. Valider le nom
    // 2. BEGIN transaction
    // 3. INSERT INTO teams (name, captain_id) VALUES ($1, $2) RETURNING id, name, captain_id, created_at, updated_at
    // 4. INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'captain')
    // 5. COMMIT
    // 6. Retourner le Team créé
    // Gérer la violation UNIQUE (pq: duplicate key) → wrapper en erreur métier
}
```

```go
// FindByID récupère une équipe par son ID.
func (r *teamRepo) FindByID(id int) (*Team, error)

// FindByUserID récupère toutes les équipes dont l'utilisateur est membre.
// Jointure : teams JOIN team_members ON teams.id = team_members.team_id WHERE team_members.user_id = $1
func (r *teamRepo) FindByUserID(userID int) ([]*Team, error)

// UpdateName met à jour le nom.
// UPDATE teams SET name = $1, updated_at = NOW() WHERE id = $2
func (r *teamRepo) UpdateName(id int, name string) error

// Delete supprime l'équipe.
// DELETE FROM teams WHERE id = $1
func (r *teamRepo) Delete(id int) error
```

### Méthodes détaillées pour TeamMemberRepository

```go
// AddMember ajoute un membre. Vérifie que le rôle est valide.
func (r *teamMemberRepo) AddMember(teamID, userID int, role string) error

// RemoveMember retire un membre.
func (r *teamMemberRepo) RemoveMember(teamID, userID int) error

// UpdateRole change le rôle. Vérifie limite de 1 co-capitaine.
func (r *teamMemberRepo) UpdateRole(teamID, userID int, role string) error

// FindByTeamID liste tous les membres d'une équipe avec leurs infos utilisateur.
// SELECT tm.team_id, tm.user_id, tm.role, tm.joined_at FROM team_members tm WHERE tm.team_id = $1
func (r *teamMemberRepo) FindByTeamID(teamID int) ([]*TeamMember, error)

// FindByUserAndTeam récupère le rôle d'un utilisateur dans une équipe.
func (r *teamMemberRepo) FindByUserAndTeam(userID, teamID int) (*TeamMember, error)

// GetCaptainID retourne l'ID du capitaine.
// SELECT user_id FROM team_members WHERE team_id = $1 AND role = 'captain'
func (r *teamMemberRepo) GetCaptainID(teamID int) (int, error)

// CountByRole compte les membres avec un rôle donné.
func (r *teamMemberRepo) CountByRole(teamID int, role string) (int, error)
```

### Contraintes techniques

- **Package naming** : `models`. Import path : `eva_trainer/internal/models`.
- **Contexte DB** : Les repositories prennent `*sql.DB` dans leur constructeur. Le package `pkg/database` fournit la connexion via `database.DB()`.
- **Transactions** : `TeamRepository.Create` doit utiliser une transaction SQL pour garantir l'atomicité (BEGIN, INSERT team, INSERT member, COMMIT). Suivre le pattern de `refreshTokenRepo.RotateRefreshToken` dans `internal/models/refresh_token.go`.
- **Validation des rôles** : Utiliser la constante `ValidRoles` pour vérifier qu'un rôle est valide avant insertion.
- **Regex** : Utiliser `regexp.MustCompile` au niveau du package (compilée une fois).
- **PostgreSQL error handling** : Pour détecter la violation UNIQUE, utiliser `strings.Contains(err.Error(), "duplicate key")` ou `errors.As` avec un `*pq.Error` (mais préférer la détection par chaîne pour éviter l'import de `lib/pq`).

### Tests à implémenter

Les tests d'intégration seront créés dans la Tâche #004. Aucun test spécifique à cette tâche — les tests des repositories seront intégrés avec les tests des handlers.

### Documentation

Aucune documentation à créer pour cette tâche. La documentation sera créée dans la Tâche #004.
