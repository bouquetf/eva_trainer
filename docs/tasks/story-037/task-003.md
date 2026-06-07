# Tâche #003 - Story #037 : Backend — API Divs + Routes + Tests Intégration

## Objectif
Implémenter les handlers API REST pour les divs (CRUD, assignation/désassignation d'équipes, classement), les routes, et les tests d'intégration.

## Contexte
- Story #037 : `docs/stories/story-037.md`
- Dépend de : Tâche #001 (migration), Tâche #002 (modèle + repository)
- Nécessaire pour : Tâche #004 (frontend), Tâche #005 (E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : Handlers API

Ajouter les méthodes suivantes à `ArchiveHandler` dans `internal/handlers/archive.go` :

**Endpoints CRUD Divs :**

| Méthode | Route | Handler | Description |
|---------|-------|---------|-------------|
| `GET` | `/archive/divisions?ligue_id=` | `ListDivs` | Liste les divisions d'une ligue |
| `GET` | `/archive/divisions/:id` | `GetDiv` | Détail d'une division (avec équipes) |
| `POST` | `/archive/divisions` | `CreateDiv` | Créer une division |
| `PUT` | `/archive/divisions/:id` | `UpdateDiv` | Modifier le nom d'une division |
| `DELETE` | `/archive/divisions/:id` | `DeleteDiv` | Supprimer une division |
| `POST` | `/archive/divisions/:id/assign` | `AssignTeamToDiv` | Assigner une équipe à une division |
| `DELETE` | `/archive/divisions/:id/teams/:teamId` | `RemoveTeamFromDiv` | Retirer une équipe d'une division |
| `GET` | `/archive/divisions/:id/standings` | `GetDivStandings` | Classement d'une division |

**Cas nominaux :**
- Création d'une div avec nom + ligue_id
- Liste des divs filtrée par ligue_id (requis)
- Détail d'une div avec la liste des équipes
- Modification du nom d'une div
- Suppression d'une div (cascade)
- Assignation d'une équipe à une div
- Désassignation d'une équipe d'une div
- Consultation du classement d'une div

**Cas limites :**
- `ligue_id` manquant dans ListDivs → 400
- ID invalide (non numérique) → 400
- Div inexistante → 404
- Nom vide → 400
- Nom en doublon → 409
- Équipe déjà assignée à une autre div de la même ligue → 409
- Équipe inexistante → 404
- Aucun BO saisi → classement vide (200 avec `[]`)

**Gestion d'erreurs :**
| Erreur | Code | Message |
|--------|------|---------|
| `ligue_id` manquant | 400 | "ligue_id query parameter is required" |
| ID invalide | 400 | "invalid div id" |
| `ErrDivNameRequired` | 400 | "Le nom de la div est requis" |
| `ErrDivNotFound` | 404 | "Div non trouvée" |
| `ErrDivNameExists` | 409 | "Une div avec ce nom existe déjà dans cette ligue" |
| `ErrEvaTeamNotFound` | 404 | "Équipe non trouvée" |
| `ErrDivTeamAlreadyAssigned` | 409 | "Cette équipe est déjà assignée à une autre div dans cette ligue" |

#### Partie 2 : Routes

Ajouter les routes dans `api/routes.go` dans le bloc `archive` existant, après les routes des équipes eva.

Toutes les routes en écriture nécessitent `middleware.RequireArchiveAccess(..., "write")`.
Les routes en lecture nécessitent `middleware.RequireArchiveAccess(..., "read")`.

#### Partie 3 : Intégration ArchiveHandler

Ajouter `divRepo models.DivRepository` au `ArchiveHandler` struct et au constructeur `NewArchiveHandler`.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Modifier | Ajouter divRepo, constructeur, handlers Div |
| `api/routes.go` | Modifier | Ajouter routes Div |
| `internal/handlers/archive_test.go` | Modifier | Ajouter tests intégration Div |

### Signatures

```go
// Ajouter au ArchiveHandler struct
type ArchiveHandler struct {
    // ... champs existants ...
    divRepo models.DivRepository
}
```

```go
// Nouveaux types de requête
type createDivRequest struct {
    LigueID int    `json:"ligue_id" binding:"required"`
    Nom     string `json:"nom" binding:"required"`
}

type updateDivRequest struct {
    Nom string `json:"nom" binding:"required"`
}

type assignTeamRequest struct {
    EquipeEvaID int `json:"equipe_eva_id" binding:"required"`
}
```

```go
// Nouvelles méthodes handler
func (h *ArchiveHandler) ListDivs(c *gin.Context)
func (h *ArchiveHandler) GetDiv(c *gin.Context)
func (h *ArchiveHandler) CreateDiv(c *gin.Context)
func (h *ArchiveHandler) UpdateDiv(c *gin.Context)
func (h *ArchiveHandler) DeleteDiv(c *gin.Context)
func (h *ArchiveHandler) AssignTeamToDiv(c *gin.Context)
func (h *ArchiveHandler) RemoveTeamFromDiv(c *gin.Context)
func (h *ArchiveHandler) GetDivStandings(c *gin.Context)
```

### Détail d'implémentation

#### `ListDivs`
```go
func (h *ArchiveHandler) ListDivs(c *gin.Context) {
    ligueIDStr := c.Query("ligue_id")
    if ligueIDStr == "" {
        apierror.RespondWithError(c, apierror.NewAPIError("missing_ligue_id", "ligue_id query parameter is required", http.StatusBadRequest))
        return
    }
    ligueID, err := strconv.Atoi(ligueIDStr)
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_ligue_id", "invalid ligue_id", http.StatusBadRequest))
        return
    }

    divs, err := h.divRepo.FindByLigueID(ligueID)
    if err != nil {
        h.logger.Error("failed to list divs", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"divs": divs})
}
```

#### `GetDiv`
```go
func (h *ArchiveHandler) GetDiv(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
        return
    }

    div, err := h.divRepo.FindByID(id)
    if errors.Is(err, models.ErrDivNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to find div", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    teams, err := h.divRepo.FindTeamsByDivID(id)
    if err != nil {
        h.logger.Error("failed to find teams for div", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
        "div":   div,
        "teams": teams,
    })
}
```

#### `CreateDiv`
```go
func (h *ArchiveHandler) CreateDiv(c *gin.Context) {
    var req createDivRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
        return
    }

    div, err := h.divRepo.Create(req.LigueID, req.Nom)
    if err != nil {
        switch {
        case errors.Is(err, models.ErrDivNameRequired):
            apierror.RespondWithError(c, apierror.NewAPIError("div_name_required", "Le nom de la div est requis", http.StatusBadRequest))
        case errors.Is(err, models.ErrDivNameExists):
            apierror.RespondWithError(c, apierror.NewAPIError("div_name_exists", "Une div avec ce nom existe déjà dans cette ligue", http.StatusConflict))
        default:
            h.logger.Error("failed to create div", "error", err)
            apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        }
        return
    }

    apierror.RespondWithSuccess(c, http.StatusCreated, gin.H{"div": div})
}
```

#### `UpdateDiv`, `DeleteDiv`
Pattern similaire : parse ID, appeler repo, gérer les erreurs avec switch.

#### `AssignTeamToDiv`
```go
func (h *ArchiveHandler) AssignTeamToDiv(c *gin.Context) {
    divID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
        return
    }

    var req assignTeamRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
        return
    }

    // Vérifier que la div existe (retournera 404 si inexistante)
    _, err = h.divRepo.FindByID(divID)
    if errors.Is(err, models.ErrDivNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to find div", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    // Vérifier que l'équipe eva existe
    _, err = h.evaTeamRepo.FindByID(req.EquipeEvaID)
    if errors.Is(err, models.ErrEvaTeamNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva non trouvée", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to find eva team", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    err = h.divRepo.AssignTeam(divID, req.EquipeEvaID)
    if errors.Is(err, models.ErrDivTeamAlreadyAssigned) {
        apierror.RespondWithError(c, apierror.NewAPIError("team_already_assigned", "Cette équipe est déjà assignée à une autre div dans cette ligue", http.StatusConflict))
        return
    }
    if err != nil {
        h.logger.Error("failed to assign team to div", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Équipe assignée à la div"})
}
```

#### `RemoveTeamFromDiv`
```go
func (h *ArchiveHandler) RemoveTeamFromDiv(c *gin.Context) {
    divID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
        return
    }

    teamID, err := strconv.Atoi(c.Param("teamId"))
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team id", http.StatusBadRequest))
        return
    }

    err = h.divRepo.RemoveTeam(divID, teamID)
    if errors.Is(err, models.ErrEvaTeamNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("team_not_in_div", "Cette équipe n'est pas dans cette div", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to remove team from div", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Équipe retirée de la div"})
}
```

#### `GetDivStandings`
```go
func (h *ArchiveHandler) GetDivStandings(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
        return
    }

    // Vérifier que la div existe
    _, err = h.divRepo.FindByID(id)
    if errors.Is(err, models.ErrDivNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to find div", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    standings, err := h.divRepo.GetStandings(id)
    if err != nil {
        h.logger.Error("failed to get standings", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"standings": standings})
}
```

### Routes (à ajouter dans `api/routes.go`)

```go
archive.GET("/divisions",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.ListDivs,
)
archive.GET("/divisions/:id",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.GetDiv,
)
archive.POST("/divisions",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.CreateDiv,
)
archive.PUT("/divisions/:id",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.UpdateDiv,
)
archive.DELETE("/divisions/:id",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.DeleteDiv,
)
archive.POST("/divisions/:id/assign",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.AssignTeamToDiv,
)
archive.DELETE("/divisions/:id/teams/:teamId",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.RemoveTeamFromDiv,
)
archive.GET("/divisions/:id/standings",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.GetDivStandings,
)
```

#### Partie 4 : Refactoring partagé `testutil`

Avant d'ajouter les nouveaux tests, créer un package `internal/testutil/` partagé pour mutualiser le setup des containers PostgreSQL entre `internal/models/setup_test.go` et `internal/handlers/archive_test.go`.

**Fichier** : `internal/testutil/db.go`
```go
package testutil

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "testing"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

const MigrationsDir = "../../db/migrations"

var DefaultMigrationFiles = []string{
    "000001_initial_schema.up.sql",
    // ... toutes les migrations existantes (mêmes que dans setup_test.go) ...
    "000030_add_eva_teams.up.sql",
    "000031_add_archive_divs_and_bos.up.sql",
}

func SetupTestDB(t *testing.T, migrationFiles []string) (*sql.DB, func()) {
    t.Helper()

    if migrationFiles == nil {
        migrationFiles = DefaultMigrationFiles
    }

    ctx := context.Background()
    // ... même logique que setupTestDB dans models/setup_test.go ...
    // Container PostgreSQL, port mapping, exécution des migrations, return db + cleanup
}

func RunMigrations(db *sql.DB, files []string, t *testing.T) {
    t.Helper()
    for _, mf := range files {
        path := filepath.Join(MigrationsDir, mf)
        content, err := os.ReadFile(path)
        if err != nil {
            t.Fatalf("failed to read migration %s: %v", mf, err)
        }
        statements := strings.Split(string(content), ";")
        for _, stmt := range statements {
            stmt = strings.TrimSpace(stmt)
            if stmt == "" {
                continue
            }
            if _, err := db.Exec(stmt); err != nil {
                t.Fatalf("failed to execute migration %s: %v\nSQL: %s", mf, err, stmt)
            }
        }
    }
}
```

**Mise à jour de `internal/models/setup_test.go`** :
- Remplacer la logique de création de container par un appel à `testutil.SetupTestDB`
- Garder la liste des migrations spécifiques aux models si différente

**Mise à jour de `internal/handlers/archive_test.go`** :
- Remplacer `setupArchiveHandlerTest` par un appel à `testutil.SetupTestDB`
- Inclure `000031_add_archive_divs_and_bos.up.sql` dans la liste des migrations

### Tests d'intégration

Ajouter dans `internal/handlers/archive_test.go` (fichier existant, build tag `integration`).

**Configuration :**
- Ajouter `divRepo := models.NewDivRepository(db)` dans le setup
- Passer `divRepo` à `NewArchiveHandler`
- Utiliser `testutil.SetupTestDB` pour le container PG

**Tests à ajouter :**

#### 1. `TestArchiveHandler_ListDivs_Empty`
- Créer une ligue
- GET `/archive/divisions?ligue_id=1`
- Vérifier `divs` = `[]`

#### 2. `TestArchiveHandler_ListDivs_WithData`
- Créer une ligue
- Créer 2 divs
- GET `/archive/divisions?ligue_id=<ligueID>`
- Vérifier 2 divs dans la réponse

#### 3. `TestArchiveHandler_ListDivs_MissingLigueID`
- GET `/archive/divisions` (sans ligue_id)
- Vérifier 400

#### 4. `TestArchiveHandler_CreateDiv_Success`
- Créer une ligue
- POST `/archive/divisions` avec `{ ligue_id, nom }`
- Vérifier 201 et div retournée avec ID > 0

#### 5. `TestArchiveHandler_CreateDiv_DuplicateName`
- Créer une ligue
- Créer une div "Groupe A"
- POST avec même nom et même ligue_id
- Vérifier 409

#### 6. `TestArchiveHandler_GetDiv_NotFound`
- GET `/archive/divisions/99999`
- Vérifier 404

#### 7. `TestArchiveHandler_UpdateDiv_Success`
- Créer une div
- PUT avec nouveau nom
- Vérifier 200 et nom mis à jour

#### 8. `TestArchiveHandler_DeleteDiv_Success`
- Créer une div
- DELETE → 200
- GET → 404

#### 9. `TestArchiveHandler_AssignTeamToDiv_Success`
- Créer une ligue, une div, une équipe eva
- POST `/archive/divisions/<id>/assign` avec `{ equipe_eva_id }`
- Vérifier 200
- GET `/archive/divisions/<id>` → vérifier teams contient l'équipe

#### 10. `TestArchiveHandler_RemoveTeamFromDiv`
- Assigner puis retirer
- Vérifier la liste des teams est vide après

#### 11. `TestArchiveHandler_GetDivStandings_Empty`
- Créer une ligue, une div, assigner une équipe
- GET `/archive/divisions/<id>/standings`
- Vérifier 200 et `standings` non null, entries avec l'équipe et 0 partout

#### 12. `TestArchiveHandler_GetDivStandings_WithBOs`
- Créer une ligue, 2 divs, 2 équipes eva
- Assigner les 2 équipes à la div
- Créer un BO entre les 2 équipes avec des matchs
- GET standings → vérifier les scores calculés

### Contraintes techniques
- **Pattern handler** : Suivre exactement le pattern des handlers existants dans `archive.go` (gin.Context, switch d'erreurs, apierror.RespondWithSuccess/Error)
- **Logger** : Utiliser `h.logger.Error` pour les erreurs internes
- **Tests** : Utiliser `testutil.SetupTestDB` (nouveau package partagé) avec la migration `000031`. Refactorer `internal/models/setup_test.go` et `internal/handlers/archive_test.go` pour utiliser ce package commun.
- **Build tag** : Les tests doivent avoir le build tag `//go:build integration` (déjà présent)

### Documentation

Documentation mise à jour dans `docs/archives-divisions.md` (créé dans la Tâche #004) :
- Liste des endpoints API REST pour les divs
- Exemples de requêtes/réponses

### Exemples d'API

Créer un fichier `docs/tasks/story-037/task-003-examples.http` :

```http
### Créer une division
POST /api/v1/archive/divisions
Content-Type: application/json

{
  "ligue_id": 1,
  "nom": "Groupe A"
}

### Lister les divisions d'une ligue
GET /api/v1/archive/divisions?ligue_id=1

### Détail d'une division
GET /api/v1/archive/divisions/1

### Modifier une division
PUT /api/v1/archive/divisions/1
Content-Type: application/json

{
  "nom": "Poule Haute"
}

### Supprimer une division
DELETE /api/v1/archive/divisions/1

### Assigner une équipe à une division
POST /api/v1/archive/divisions/1/assign
Content-Type: application/json

{
  "equipe_eva_id": 1
}

### Retirer une équipe d'une division
DELETE /api/v1/archive/divisions/1/teams/1

### Classement d'une division
GET /api/v1/archive/divisions/1/standings
```
