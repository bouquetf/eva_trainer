# Tâche #002 - Story #036 : Handler CRUD EvaTeam + Routes API

## Objectif
Implémenter les endpoints CRUD complets pour les équipes eva dans le handler `ArchiveHandler` et ajouter les routes manquantes (GET/:id, PUT/:id, DELETE/:id).

## Contexte
- Story #036 : `docs/stories/story-036.md`
- Dépend de : Tâche #001 (modèle EvaTeamRepository et erreurs)
- Nécessaire pour : Tâche #003 (frontend API), Tâche #004 (tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : Handler CRUD dans ArchiveHandler

Remplacer les stubs `ListEvaTeams` et `CreateEvaTeam` dans `internal/handlers/archive.go` par des implémentations complètes. Ajouter `GetEvaTeam`, `UpdateEvaTeam`, `DeleteEvaTeam`.

**Endpoints à implémenter :**

| Méthode | Path | Handler | Auth | Permission |
|---------|------|---------|------|------------|
| GET | `/api/v1/archive/teams` | `ListEvaTeams` | RequireAuth | read |
| GET | `/api/v1/archive/teams/:id` | `GetEvaTeam` | RequireAuth | read |
| POST | `/api/v1/archive/teams` | `CreateEvaTeam` | RequireAuth | write |
| PUT | `/api/v1/archive/teams/:id` | `UpdateEvaTeam` | RequireAuth | write |
| DELETE | `/api/v1/archive/teams/:id` | `DeleteEvaTeam` | RequireAuth | write |

**Comportement attendu :**

- `ListEvaTeams` : Retourne la liste de toutes les équipes eva triées par nom. Réponse : `{ "teams": [...] }`
- `GetEvaTeam` : Retourne une équipe eva par son ID. 404 si inexistante. Réponse : `{ "team": {...} }`
- `CreateEvaTeam` : Crée une équipe eva avec name (obligatoire), tag (obligatoire), eva_trainer_team_id (optionnel). 400 si validation échoue. 409 si eva_trainer_team_id déjà utilisé. 404 si eva_trainer_team_id fourni mais l'équipe eva-trainer n'existe pas. Réponse : `{ "team": {...} }` avec status 201
- `UpdateEvaTeam` : Met à jour une équipe eva. Mêmes règles de validation que la création. 404 si inexistante. 409 si eva_trainer_team_id déjà utilisé par une autre équipe. 404 si eva_trainer_team_id fourni mais l'équipe eva-trainer n'existe pas. Réponse : `{ "team": {...} }`
- `DeleteEvaTeam` : Supprime une équipe eva. 404 si inexistante. Appeler `evaTeamRepo.HasReferences(id)` AVANT la suppression. Si `HasReferences` retourne true, répondre 409. Réponse : `{ "message": "Équipe eva supprimée" }`

**Gestion d'erreurs :**
| Condition | Code HTTP | Message |
|-----------|-----------|---------|
| name vide | 400 | "Le nom est requis" |
| tag vide | 400 | "Le tag est requis" |
| tag > 50 caractères | 400 | "Le tag ne peut pas dépasser 50 caractères" |
| Équipe eva introuvable | 404 | "Équipe eva non trouvée" |
| eva_trainer_team_id fourni mais équipe inexistante | 404 | "Équipe eva-trainer non trouvée" |
| eva_trainer_team_id déjà utilisé | 409 | "Cette équipe eva-trainer est déjà associée à une autre équipe eva" |
| Suppression bloquée (références) | 409 | "Impossible de supprimer cette équipe eva : elle est référencée par des ligues ou matchs" |
| Erreur interne | 500 | "Erreur interne du serveur" |

**Détail de l'association eva_trainer_team_id :**
- Le champ `eva_trainer_team_id` est optionnel (nullable)
- Si fourni, vérifier que l'équipe eva-trainer existe avec `teamRepo.FindByID(id)` avant d'associer. Si l'équipe n'existe pas → 404
- La contrainte UNIQUE sur la colonne garantit la relation 1:1 — si violée, la DB retourne une erreur que le handler doit détecter et traduire en 409
- L'API GET /api/v1/teams (existante) est utilisée par le frontend pour lister les équipes eva-trainer disponibles dans le selecteur d'association

#### Partie 2 : Modification du constructeur ArchiveHandler

Modifier `NewArchiveHandler` pour accepter `EvaTeamRepository` en paramètre.

```go
type ArchiveHandler struct {
    ligueLocaleRepo models.LigueLocaleRepository
    ligueRepo       models.LigueRepository
    splitRepo       models.SplitRepository
    evaTeamRepo     models.EvaTeamRepository   // NOUVEAU
    logger          services.Logger
}

func NewArchiveHandler(
    ligueLocaleRepo models.LigueLocaleRepository,
    ligueRepo models.LigueRepository,
    splitRepo models.SplitRepository,
    evaTeamRepo models.EvaTeamRepository,   // NOUVEAU paramètre
    logger services.Logger,
) *ArchiveHandler
```

#### Partie 3 : Routes API

Ajouter les routes manquantes dans `api/routes.go` (dans `setupArchiveRoutes`) :

```go
archive.GET("/teams",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.ListEvaTeams,
)
archive.GET("/teams/:id",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.GetEvaTeam,
)
archive.POST("/teams",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.CreateEvaTeam,
)
archive.PUT("/teams/:id",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.UpdateEvaTeam,
)
archive.DELETE("/teams/:id",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.DeleteEvaTeam,
)
```

NOTA : les routes GET et POST existent déjà (stubs). Il faut :
1. Conserver GET /teams et POST /teams (le handler sera remplacé)
2. Ajouter GET /teams/:id, PUT /teams/:id, DELETE /teams/:id

Mettre à jour l'instanciation de `ArchiveHandler` dans `setupArchiveRoutes` pour passer `evaTeamRepo`.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Modifier | Remplacer stubs, ajouter CRUD complet, modifier constructeur |
| `api/routes.go` | Modifier | Ajouter routes GET/:id, PUT/:id, DELETE/:id + instanciation evaTeamRepo |

### Signatures

```go
// Nouvelles méthodes sur ArchiveHandler
func (h *ArchiveHandler) GetEvaTeam(c *gin.Context)
func (h *ArchiveHandler) UpdateEvaTeam(c *gin.Context)
func (h *ArchiveHandler) DeleteEvaTeam(c *gin.Context)

// Modifications des méthodes existantes
func (h *ArchiveHandler) ListEvaTeams(c *gin.Context)  // stub → implémentation
func (h *ArchiveHandler) CreateEvaTeam(c *gin.Context)  // stub → implémentation
```

```go
// Requêtes
type createEvaTeamRequest struct {
    Name             string `json:"name" binding:"required"`
    Tag              string `json:"tag" binding:"required"`
    EvaTrainerTeamID *int   `json:"eva_trainer_team_id,omitempty"`
}

type updateEvaTeamRequest struct {
    Name             string `json:"name" binding:"required"`
    Tag              string `json:"tag" binding:"required"`
    EvaTrainerTeamID *int   `json:"eva_trainer_team_id,omitempty"`
}
```

```go
// Réponses
type evaTeamResponse struct {
    Team *models.EvaTeam `json:"team"`
}

type evaTeamListResponse struct {
    Teams []*models.EvaTeam `json:"teams"`
}
```

### Contraintes techniques
- **Pattern handler** : Suivre le pattern des handlers existants (ex: `CreateLigueLocale` dans `archive.go`) : Bind JSON → Validation → Appel repo → Réponse
- **Gestion erreur unique** : Pour la contrainte UNIQUE sur `eva_trainer_team_id`, le repository retourne une erreur PostgreSQL. Le handler doit détecter le code d'erreur `23505` (unique_violation) avec `strings.Contains(err.Error(), "23505")` ou via `errors.As` si un type dédié existe
- **Vérification existence équipe eva-trainer** : Avant d'associer, utiliser `teamRepo.FindByID(id)` (le TeamRepository existe déjà dans le handler — non, il n'est pas dans ArchiveHandler). Il faut soit ajouter teamRepo au constructeur d'ArchiveHandler, soit faire une requête directe avec `db.QueryRow("SELECT 1 FROM teams WHERE id = $1", id)`.
- **Vérification références avant suppression** : Avant DELETE, appeler `evaTeamRepo.HasReferences(id)`. Cette méthode interroge `information_schema.tables` pour détecter les tables FK référençant `eva_teams(id)` et exécute COUNT(*) sur chaque table trouvée. Si un COUNT > 0, retourne 409.
- **Logger** : `h.logger.Error("msg", "error", err)` pour les erreurs internes
- **Réponse paginée** : Non nécessaire (FindAll retourne tout, le nombre d'équipes eva reste faible)
- **Constructeur** : Mettre à jour `NewArchiveHandler` — attention à l'appel existant dans `api/routes.go`
- **Pas de logo_url** : Le champ `logo_url` n'est pas inclus (décision utilisateur)

### Tests

Les tests d'intégration sont détaillés dans la Tâche #004.

**Vérification manuelle :**
```bash
go build ./...
go vet ./internal/handlers/ ./api/
```

### Documentation

Mise à jour de `docs/archives-eva-teams.md` (créé dans la Tâche #004) avec :
- Documentation des endpoints API (méthode, path, body, réponse, codes d'erreur)
- Exemples de requêtes curl
