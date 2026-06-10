# Tâche #002 - Story #039 : Handler API CardStats + routes

## Objectif
Ajouter l'endpoint API REST `GET /api/v1/archive/card-stats` et l'intégration dans les routes.

## Contexte
- Story #039 : `docs/stories/story-039.md`
- Dépend de : Tâche #001 (Modèle CardStats)
- Nécessaire pour : Tâche #003 (Frontend)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

1. Ajouter une méthode `GetCardStats` au `ArchiveHandler` existant
2. Ajouter `cardStatsRepo models.CardStatsRepository` au struct `ArchiveHandler`
3. Mettre à jour le constructeur `NewArchiveHandler`
4. Enregistrer la route `GET /api/v1/archive/card-stats` dans les routes

**Endpoint :**
```
GET /api/v1/archive/card-stats?team_id=1
```

**Query parameters :**
- `team_id` (requis) : ID de l'équipe eva

**Réponse succès (200) :**
```json
{
  "stats": [
    {
      "map_id": 1,
      "map_name": "cp_metal",
      "wins": 3,
      "losses": 1,
      "winrate": 0.75,
      "avg_score": 75.5,
      "median_score": 78.0,
      "opponents": [
        {
          "team_id": 2,
          "team_name": "Team B",
          "avg_score": 72.0,
          "median_score": 72.0
        }
      ]
    }
  ]
}
```

**Gestion d'erreurs :**
- `team_id` manquant ou invalide → 400 "team_id query parameter is required/invalid"
- `ErrEvaTeamNotFound` → 404 "Équipe eva non trouvée"
- Erreur interne → 500 "internal server error"

**Permissions :** Protéger la route avec le middleware `middleware.RequireArchiveRead()` (déjà en place pour les autres routes du groupe `/archive`)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Modifier | Ajouter `cardStatsRepo` au struct, constructeur, méthode `GetCardStats` |
| `api/routes.go` | Modifier | Ajouter la route `GET /archive/card-stats` |
| `internal/handlers/archive_test.go` | Modifier | Ajouter test d'intégration pour CardStats |

### Modifications dans `archive.go`

**Struct :**
```go
type ArchiveHandler struct {
    // ... champs existants ...
    boRepo          models.BORepository
    cardStatsRepo   models.CardStatsRepository    // AJOUTER ICI (entre boRepo et divRepo)
    matchRepo       models.MatchRepository
    banRepo         models.BanRepository
    pickRepo        models.PickRepository
    mapRepo         models.MapRepository
    // ... autres champs ...
}
```

**Constructeur (ajouter `cardStatsRepo` entre `boRepo` et `divRepo`) :**
```go
func NewArchiveHandler(
    ligueLocaleRepo models.LigueLocaleRepository,
    ligueRepo models.LigueRepository,
    splitRepo models.SplitRepository,
    evaTeamRepo models.EvaTeamRepository,
    teamRepo models.TeamRepository,
    boRepo models.BORepository,
    cardStatsRepo models.CardStatsRepository,   // AJOUTER ICI
    divRepo models.DivRepository,
    matchRepo models.MatchRepository,
    banRepo models.BanRepository,
    pickRepo models.PickRepository,
    mapRepo models.MapRepository,
    logger services.Logger,
) *ArchiveHandler {
    return &ArchiveHandler{
        ligueLocaleRepo: ligueLocaleRepo,
        ligueRepo:       ligueRepo,
        splitRepo:       splitRepo,
        evaTeamRepo:     evaTeamRepo,
        teamRepo:        teamRepo,
        boRepo:          boRepo,
        cardStatsRepo:   cardStatsRepo,          // AJOUTER ICI
        divRepo:         divRepo,
        matchRepo:       matchRepo,
        banRepo:         banRepo,
        pickRepo:        pickRepo,
        mapRepo:         mapRepo,
        logger:          logger,
    }
}
```

**Mise à jour de l'appel dans `api/routes.go` :**
```go
archiveHandler := handlers.NewArchiveHandler(
    ligueLocaleRepo, ligueRepo, splitRepo, evaTeamRepo, teamRepo,
    boRepo, cardStatsRepo, divRepo, matchRepo, banRepo, pickRepo, mapRepo, logger,
)
// où cardStatsRepo := models.NewCardStatsRepository(db)
```

**Méthode :**
```go
func (h *ArchiveHandler) GetCardStats(c *gin.Context) {
    teamIDStr := c.Query("team_id")
    if teamIDStr == "" {
        apierror.RespondWithError(c, apierror.NewAPIError("missing_team_id", "team_id query parameter is required", http.StatusBadRequest))
        return
    }
    teamID, err := strconv.Atoi(teamIDStr)
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_team_id", "invalid team_id", http.StatusBadRequest))
        return
    }

    // Vérifier que l'équipe existe
    _, err = h.evaTeamRepo.FindByID(teamID)
    if errors.Is(err, models.ErrEvaTeamNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("eva_team_not_found", "Équipe eva non trouvée", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to find eva team", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    stats, err := h.cardStatsRepo.Compute(teamID)
    if err != nil {
        h.logger.Error("failed to compute card stats", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"stats": stats})
}
```

### Modification dans `api/routes.go`

Dans la fonction de setup des routes, après les routes BO existantes, ajouter :

```go
archive.GET("/card-stats",
    middleware.RequireArchiveRead(),
    archiveHandler.GetCardStats,
)
```

### Tests à implémenter

#### Test d'intégration (dans `internal/handlers/archive_test.go`)

Utiliser les helpers existants `createTestLigue`, `createTestEvaTeam` et créer les données suivantes :

- 2 équipes eva (A et B)
- 1 ligue
- 1 split avec des maps
- 1 BO entre A et B avec 3 matchs :
  - Match 1 : map "cp_metal", A gagne 80-60
  - Match 2 : map "cp_metal", B gagne 55-70 (B gagne)
  - Match 3 : map "mistral", A gagne 90-50

- **Test :** `GET /api/v1/archive/card-stats?team_id=<A_ID>` → 200
  - Vérifier 2 stats (cp_metal et mistral)
  - cp_metal : wins=1, losses=1, winrate=0.5, avg_score=67.5, median_score=67.5
  - mistral : wins=1, losses=0, winrate=1.0, avg_score=90.0, median_score=90.0

- **Test :** `GET /api/v1/archive/card-stats` (sans team_id) → 400

- **Test :** `GET /api/v1/archive/card-stats?team_id=99999` → 404

- **Test :** `GET /api/v1/archive/card-stats?team_id=<team_sans_match>` → 200 avec `stats: []`

**Note :** Utiliser le pattern de test existant avec `setupArchiveHandlerTest`, `gin.Context`, `httptest.NewRecorder`. Les tests doivent avoir le tag `//go:build integration`.

### Contraintes techniques
- **Pattern** : Suivre exactement le pattern des autres handlers d'archive (ex: `ListBOs`, `GetDivStandings`)
- **Middleware** : La route doit être protégée par `middleware.RequireArchiveRead()` (déjà utilisé pour les autres routes du groupe)
- **Logger** : Logger les erreurs avec `h.logger.Error()` comme dans les autres handlers
