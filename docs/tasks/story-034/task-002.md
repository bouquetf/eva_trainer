# Tâche #002 - Story #034 : Modèles Go, repositories et middleware RequireArchiveAccess

## Objectif
Implémenter les modèles Go, les repositories et le middleware `RequireArchiveAccess` pour les rôles "archiviste" (global) et "lecteur" (par équipe).

## Contexte
- Story #034 : `docs/stories/story-034.md`
- Dépend de : Tâche #001 (migration DB)
- Nécessaire pour : Tâche #003, Tâche #004

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : Modèles et repositories

Créer deux nouveaux modèles avec leurs repositories :

**GlobalRole** : représente un rôle global attribué à un utilisateur.
- `NewGlobalRoleRepository(db)` avec méthodes :
  - `Add(userID int, role string) error`
  - `Remove(userID int, role string) error`
  - `UserHasRole(userID int, role string) (bool, error)`
  - `FindUsersByRole(role string) ([]*UserGlobalRole, error)` — retourne les utilisateurs avec leur display_name

**TeamRole** : représente un rôle attribué à une équipe.
- `NewTeamRoleRepository(db)` avec méthodes :
  - `Add(teamID int, role string) error`
  - `Remove(teamID int, role string) error`
  - `TeamHasRole(teamID int, role string) (bool, error)`
  - `FindTeamsByRole(role string) ([]*TeamWithRole, error)` — retourne les équipes avec leur nom
  - `UserIsMemberOfTeamWithRole(userID int, role string) (bool, error)` — vérifie si l'utilisateur est membre d'au moins une équipe ayant le rôle donné

**Définitions de constantes** :
- `RoleArchivist = "archivist"`
- `RoleReader = "reader"`

#### Partie 2 : Middleware RequireArchiveAccess

Créer un nouveau middleware `RequireArchiveAccess` qui vérifie les droits d'accès aux archives :
- **read** : l'utilisateur a accès en lecture si (admin) OU (archivist dans global_roles) OU (membre d'une équipe avec reader dans team_roles)
- **write** : l'utilisateur a accès en écriture si (admin) OU (archivist dans global_roles)

**Cas nominaux :**
- Admin → accès complet (read + write)
- Archiviste → accès complet (read + write)
- Membre d'une équipe avec rôle reader → accès read uniquement
- Utilisateur sans aucun rôle → 403 Forbidden

**Cas limites :**
- Archiviste sans équipe reader peut quand même lire → OK
- Utilisateur membre de plusieurs équipes mais aucune avec reader → 403
- Le middleware vérifie le paramètre `accessType` ("read" ou "write")

**Gestion d'erreurs :**
- Utilisateur non authentifié → 401 (laisse passer via RequireAuth en amont)
- `accessType` invalide → panique ou erreur 500 (valeur codée en dur dans les routes)
- Erreur de base de données → 500

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/global_role.go` | Créer | Modèle GlobalRole + repository |
| `internal/models/team_role.go` | Créer | Modèle TeamRole + repository |
| `internal/models/repository.go` | Modifier | Ajouter les interfaces GlobalRoleRepository et TeamRoleRepository |
| `internal/middleware/archive.go` | Créer | Middleware RequireArchiveAccess |

### Signatures

```go
// internal/models/global_role.go

type UserGlobalRole struct {
    UserID      int    `json:"user_id"`
    Role        string `json:"role"`
    DisplayName string `json:"display_name,omitempty"`
}

type GlobalRoleRepository interface {
    Add(userID int, role string) error
    Remove(userID int, role string) error
    UserHasRole(userID int, role string) (bool, error)
    FindUsersByRole(role string) ([]*UserGlobalRole, error)
}

type globalRoleRepo struct {
    db *sql.DB
}

func NewGlobalRoleRepository(db *sql.DB) GlobalRoleRepository
```

```go
// internal/models/team_role.go

type TeamWithRole struct {
    TeamID   int    `json:"team_id"`
    Role     string `json:"role"`
    TeamName string `json:"team_name,omitempty"`
}

type TeamRoleRepository interface {
    Add(teamID int, role string) error
    Remove(teamID int, role string) error
    TeamHasRole(teamID int, role string) (bool, error)
    FindTeamsByRole(role string) ([]*TeamWithRole, error)
    UserIsMemberOfTeamWithRole(userID int, role string) (bool, error)
}

type teamRoleRepo struct {
    db *sql.DB
}

func NewTeamRoleRepository(db *sql.DB) TeamRoleRepository
```

```go
// internal/models/constants.go (ou dans team_member.go)

const (
    RoleArchivist = "archivist"
    RoleReader    = "reader"
)
```

```go
// internal/middleware/archive.go

func RequireArchiveAccess(
    logger services.Logger,
    globalRoleRepo models.GlobalRoleRepository,
    teamRoleRepo models.TeamRoleRepository,
    accessType string, // "read" ou "write"
) gin.HandlerFunc
```

Renvoyer les erreurs via `apierror.RespondWithError` avec les codes suivants :
- `apierror.ErrCodeForbidden` (HTTP 403) pour accès refusé
- `apierror.ErrCodeInternalError` (HTTP 500) pour erreur de base de données

### Contraintes techniques
- **Pattern** : Suivre exactement le même pattern que les repositories existants (e.g. `team_member.go`, `permission.go`)
- **SQL** : Utiliser `database/sql` avec des requêtes paramétrées (`$1`, `$2`)
- **DB** : Accès via `database.DB()` ou réception de `*sql.DB`
- **Transaction** : Les méthodes Add/Remove n'ont pas besoin de transaction (simple INSERT/DELETE)
- **Logger** : Utiliser `services.Logger` comme dans les autres middlewares

### Tests à implémenter

#### Tests unitaires des repositories

**Fichier** : `internal/models/global_role_test.go`

| Scénario | Input | Résultat attendu |
|----------|-------|------------------|
| Ajouter un rôle global | userID valide, role "archivist" | Pas d'erreur |
| Vérifier qu'un utilisateur a un rôle | Même userID, role "archivist" | `true`, pas d'erreur |
| Vérifier qu'un utilisateur n'a PAS un rôle | userID sans le rôle, role "reader" | `false`, pas d'erreur |
| Supprimer un rôle global | userID, role "archivist" | Pas d'erreur |
| Vérifier après suppression | Même userID, role "archivist" | `false`, pas d'erreur |
| Lister les utilisateurs par rôle | role "archivist" | Liste avec le bon userID |
| Ajouter un rôle inexistant (userID inconnu) | userID invalide | Erreur (clé étrangère) |

**Fichier** : `internal/models/team_role_test.go`

| Scénario | Input | Résultat attendu |
|----------|-------|------------------|
| Ajouter un rôle d'équipe | teamID valide, role "reader" | Pas d'erreur |
| Vérifier qu'une équipe a le rôle | Même teamID, role "reader" | `true`, pas d'erreur |
| Vérifier qu'une équipe n'a PAS le rôle | teamID, role "archivist" | `false`, pas d'erreur |
| Supprimer un rôle d'équipe | teamID, role "reader" | Pas d'erreur |
| Lister les équipes par rôle | role "reader" | Liste avec le bon teamID |
| Vérifier si l'utilisateur est membre d'une équipe avec le rôle | userID, role "reader" | `true` si membre d'une telle équipe |
| Ajouter un rôle sur une équipe inexistante | teamID invalide | Erreur (clé étrangère) |

#### Tests d'intégration du middleware

**Fichier** : `internal/middleware/archive_test.go`

| Scénario | Contexte | Résultat attendu |
|----------|----------|------------------|
| Admin accède en write | user.IsAdmin = true | `c.Next()` appelé |
| Archiviste accède en write | globalRoleRepo.UserHasRole("archivist") = true | `c.Next()` appelé |
| Archiviste accède en read | globalRoleRepo.UserHasRole("archivist") = true | `c.Next()` appelé |
| Membre d'équipe reader accède en read | UserIsMemberOfTeamWithRole("reader") = true | `c.Next()` appelé |
| Membre d'équipe reader accède en write | UserIsMemberOfTeamWithRole("reader") = true | 403 refusé |
| Utilisateur sans rôle accède en read | Aucun droit | 403 refusé |

### Documentation
- Ajouter la section "Middleware d'accès aux archives" dans `docs/roles-archives.md`
