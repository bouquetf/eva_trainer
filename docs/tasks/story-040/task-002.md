# Tâche #002 - Story #040 : Handler API HeadToHeadStats + routes

## Objectif
Ajouter l'endpoint API REST `GET /api/v1/archive/head-to-head-stats` au handler ArchiveHandler, injecter le nouveau repository, et enregistrer la route.

## Contexte
- Story #040 : `docs/stories/story-040.md`
- Dépend de : Tâche #001 (Modèle HeadToHeadStats)
- Nécessaire pour : Tâche #003 (Frontend)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

1. Ajouter une méthode `GetHeadToHeadStats` à `ArchiveHandler`
2. Ajouter `headToHeadStatsRepo models.HeadToHeadStatsRepository` au struct `ArchiveHandler`
3. Mettre à jour le constructeur `NewArchiveHandler`
4. Enregistrer la route `GET /api/v1/archive/head-to-head-stats` dans les routes

**Endpoint :**
```
GET /api/v1/archive/head-to-head-stats?team1_id=1&team2_id=2
```

**Query parameters :**
- `team1_id` (requis) : ID de la première équipe eva
- `team2_id` (requis) : ID de la deuxième équipe eva

**Réponse succès (200) :**
```json
{
  "team1_id": 1,
  "team2_id": 2,
  "stats": [
    {
      "map_id": 1,
      "map_name": "cp_metal",
      "team1_stats": {
        "wins": 1,
        "losses": 1,
        "winrate": 0.5,
        "avg_score": 67.5,
        "median_score": 67.5
      },
      "team2_stats": {
        "wins": 1,
        "losses": 1,
        "winrate": 0.5,
        "avg_score": 65.0,
        "median_score": 65.0
      }
    }
  ]
}
```

**Gestion d'erreurs :**
- `team1_id` manquant ou invalide → 400 "team1_id query parameter is required/invalid"
- `team2_id` manquant ou invalide → 400 "team2_id query parameter is required/invalid"
- `team1_id == team2_id` → 400 "team1 and team2 must be different"
- `ErrEvaTeamNotFound` → 404 "Équipe eva non trouvée"
- Erreur interne → 500

**Permissions :** Protéger la route avec `middleware.RequireArchiveAccess(..., "read")`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Modifier | Ajouter `headToHeadStatsRepo` au struct, constructeur, méthode `GetHeadToHeadStats` |
| `api/routes.go` | Modifier | Ajouter la route `GET /archive/head-to-head-stats` et l'instanciation du repository |
| `internal/handlers/archive_test.go` | Modifier | Ajouter tests d'intégration pour HeadToHeadStats |

### Modifications dans `archive.go`

**Struct :**
```go
type ArchiveHandler struct {
    ligueLocaleRepo     models.LigueLocaleRepository
    ligueRepo           models.LigueRepository
    splitRepo           models.SplitRepository
    evaTeamRepo         models.EvaTeamRepository
    teamRepo            models.TeamRepository
    divRepo             models.DivRepository
    boRepo              models.BORepository
    cardStatsRepo       models.CardStatsRepository
    headToHeadStatsRepo models.HeadToHeadStatsRepository  // AJOUTER (après cardStatsRepo)
    matchRepo           models.MatchRepository
    banRepo             models.BanRepository
    pickRepo            models.PickRepository
    mapRepo             models.MapRepository
    logger              services.Logger
}
```

**Constructeur (ajouter `headToHeadStatsRepo` entre `cardStatsRepo` et `matchRepo`) :**
```go
func NewArchiveHandler(
    ligueLocaleRepo models.LigueLocaleRepository,
    ligueRepo models.LigueRepository,
    splitRepo models.SplitRepository,
    evaTeamRepo models.EvaTeamRepository,
    teamRepo models.TeamRepository,
    divRepo models.DivRepository,
    boRepo models.BORepository,
    cardStatsRepo models.CardStatsRepository,
    headToHeadStatsRepo models.HeadToHeadStatsRepository,  // AJOUTER
    matchRepo models.MatchRepository,
    banRepo models.BanRepository,
    pickRepo models.PickRepository,
    mapRepo models.MapRepository,
    logger services.Logger,
) *ArchiveHandler {
    return &ArchiveHandler{
        // ... existants ...
        cardStatsRepo:       cardStatsRepo,
        headToHeadStatsRepo: headToHeadStatsRepo,  // AJOUTER
        matchRepo:           matchRepo,
        // ... existants ...
    }
}
```

**Méthode :**
```go
func (h *ArchiveHandler) GetHeadToHeadStats(c *gin.Context) {
    team1IDStr := c.Query("team1_id")
    team2IDStr := c.Query("team2_id")

    if team1IDStr == "" {
        apierror.RespondWithError(c, apierror.NewAPIError("missing_team1_id", "team1_id query parameter is required", http.StatusBadRequest))
        return
    }
    if team2IDStr == "" {
        apierror.RespondWithError(c, apierror.NewAPIError("missing_team2_id", "team2_id query parameter is required", http.StatusBadRequest))
        return
    }

    team1ID, err := strconv.Atoi(team1IDStr)
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_team1_id", "invalid team1_id", http.StatusBadRequest))
        return
    }
    team2ID, err := strconv.Atoi(team2IDStr)
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_team2_id", "invalid team2_id", http.StatusBadRequest))
        return
    }

    if team1ID == team2ID {
        apierror.RespondWithError(c, apierror.NewAPIError("same_team", "team1 and team2 must be different", http.StatusBadRequest))
        return
    }

    // Vérifier que les 2 équipes existent
    _, err = h.evaTeamRepo.FindByID(team1ID)
    if errors.Is(err, models.ErrEvaTeamNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva team1 non trouvée", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to find eva team1", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    _, err = h.evaTeamRepo.FindByID(team2ID)
    if errors.Is(err, models.ErrEvaTeamNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva team2 non trouvée", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to find eva team2", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    stats, err := h.headToHeadStatsRepo.Compute(team1ID, team2ID)
    if err != nil {
        h.logger.Error("failed to compute head-to-head stats", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
        "team1_id": team1ID,
        "team2_id": team2ID,
        "stats":    stats,
    })
}
```

### Modification dans `api/routes.go`

1. Ajouter l'instanciation du repository dans `setupArchiveRoutes` :
```go
headToHeadStatsRepo := models.NewHeadToHeadStatsRepository(db)
```

2. Ajouter le paramètre au constructeur de l'handler (entre `cardStatsRepo` et `matchRepo`) :
```go
archiveHandler := handlers.NewArchiveHandler(
    ligueLocaleRepo, ligueRepo, splitRepo, evaTeamRepo, teamRepo,
    divRepo, boRepo, cardStatsRepo, headToHeadStatsRepo,
    matchRepo, banRepo, pickRepo, mapRepo, logger,
)
```

3. Ajouter la route après la route `/card-stats` existante :
```go
archive.GET("/head-to-head-stats",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.GetHeadToHeadStats,
)
```

### Tests à implémenter

#### Tests d'intégration (dans `internal/handlers/archive_test.go`)

Ajouter les tests suivants :

- **Test : GetHeadToHeadStats_Success**
  - Créer 2 maps, 1 ligue, 2 eva teams
  - Créer 1 BO avec 3 matchs :
    - Match 1 : map1, team1 gagne 80-60
    - Match 2 : map1, team2 gagne 55-70
    - Match 3 : map2, team1 gagne 90-50
  - `GET /archive/head-to-head-stats?team1_id=<team1>&team2_id=<team2>` → 200
  - Vérifier `team1_id`, `team2_id` dans la réponse
  - Vérifier 2 stats (map1, map2)
  - map1 : team1 wins=1, losses=1, winrate=0.5, avg_score=67.5 ; team2 wins=1, losses=1, winrate=0.5, avg_score=65.0
  - map2 : team1 wins=1, losses=0, winrate=1.0, avg_score=90.0 ; team2 wins=0, losses=1, winrate=0.0, avg_score=50.0

- **Test : GetHeadToHeadStats_MissingTeam1ID** → 400

- **Test : GetHeadToHeadStats_MissingTeam2ID** → 400

- **Test : GetHeadToHeadStats_SameTeam** → 400

- **Test : GetHeadToHeadStats_TeamNotFound** → 404

- **Test : GetHeadToHeadStats_NoMatches** → 200 avec `stats: []`

### Contraintes techniques
- **Pattern** : Suivre exactement le pattern de `GetCardStats` dans le même fichier
- **Middleware** : La route doit être protégée par `middleware.RequireArchiveAccess(..., "read")`
- **Logger** : Logger les erreurs avec `h.logger.Error()`
- **Tests** : Utiliser `setupArchiveHandlerTest` existant et les helpers `createTestEvaTeam`, `createTestLigue`
