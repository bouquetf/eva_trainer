# Tâche #001 - Story #046 : Backend `hors_competition` — migration, modèle, logique

## Objectif
Ajouter le champ `hors_competition` sur les matchs côté backend : migration DB, mise à jour du modèle `Match`, correction de `ComputeBOWinner` et `GetStandings` pour exclure les matchs hors compétition du résultat du BO, et création d'un endpoint pour toggler le flag.

## Contexte
- Story #046 : `docs/stories/story-046.md`
- Dépend de : aucune
- Nécessaire pour : Tâche #003 (Frontend)

## Critères d'acceptation concernés
- `[ ]` L'API expose le champ `hors_competition` (boolean) sur les endpoints match (ligne 27)
- `[ ]` Un match "hors compétition" est compté dans les statistiques (stats cartes, winrate, scores) (ligne 25)
- `[ ]` Un match "hors compétition" n'est PAS compté dans le résultat du BO (victoires/défaites du BO) (ligne 26)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter la possibilité de marquer un match comme "hors compétition" (exhibition, warm-up). Ces matchs sont comptés dans les statistiques (card stats, head-to-head) mais ne comptent pas dans le résultat du BO (victoires/défaites du BO, classement).

**Cas nominaux :**
- Créer un match avec `hors_competition = false` (comportement existant)
- Créer un match avec `hors_competition = true`
- Toggler `hors_competition` sur un match existant
- Un match `hors_competition = true` n'est pas compté dans `ComputeBOWinner`
- Un match `hors_competition = true` est compté dans les card stats et head-to-head stats

**Cas limites :**
- `hors_competition` par défaut à `false` pour les matchs existants (migration avec DEFAULT)
- Forfait + hors_competition : un match forfait peut aussi être hors compétition
- Un BO dont tous les matchs sont "hors compétition" → pas de vainqueur (nil)

**Gestion d'erreurs :**
- Match introuvable → 404
- BO introuvable (pour le toggle) → 404
- Requête invalide → 400

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000035_add_matchs_hors_competition.up.sql` | Créer | Migration pour ajouter `hors_competition` |
| `db/migrations/000035_add_matchs_hors_competition.down.sql` | Créer | Rollback |
| `internal/models/match.go` | Modifier | Ajouter `HorsCompetition` au struct, `scanMatch`, Create, Update |
| `internal/models/repository.go` | Modifier | Ajouter `hors_competition` dans l'interface `MatchRepository` (Create, Update) |
| `internal/models/bo.go` | Modifier | `ComputeBOWinner` ignore les matchs `hors_competition` |
| `internal/models/div.go` | Modifier | `GetStandings` ignore les matchs `hors_competition` pour le résultat BO |
| `internal/handlers/archive.go` | Modifier | Nouvel endpoint `UpdateMatch` (PATCH générique) |
| `api/routes.go` | Modifier | Ajouter la route PATCH `/archive/matchs/:id` |

### Signatures

#### Models

```go
// internal/models/match.go — Match struct
type Match struct {
    ID              int       `json:"id"`
    BoID            int       `json:"bo_id"`
    MapID           int       `json:"map_id"`
    ScoreEquipe1    int       `json:"score_equipe1"`
    ScoreEquipe2    int       `json:"score_equipe2"`
    Forfait         int       `json:"forfait"`
    VainqueurID     *int      `json:"vainqueur_id,omitempty"`
    Commentaires    *string   `json:"commentaires,omitempty"`
    HorsCompetition bool      `json:"hors_competition"`    // NOUVEAU
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

#### Repository interface

```go
// internal/models/repository.go — MatchRepository
type MatchRepository interface {
    Create(boID, mapID, scoreEquipe1, scoreEquipe2, forfait int, commentaires *string, horsCompetition bool) (*Match, error)
    FindByID(id int) (*Match, error)
    FindByBOID(boID int) ([]*Match, error)
    Update(id, mapID, scoreEquipe1, scoreEquipe2, forfait int, commentaires *string, horsCompetition bool) (*Match, error)
    Delete(id int) error
    DeleteByBOID(boID int) error
    SetHorsCompetition(id int, horsCompetition bool) (*Match, error)  // pour le PATCH générique
}
```

#### Handlers

```go
// internal/handlers/archive.go
// Nouveau type pour PATCH générique
type updateMatchRequest struct {
    HorsCompetition *bool `json:"hors_competition"`
}

// Nouveau handler
func (h *ArchiveHandler) UpdateMatch(c *gin.Context)
// PATCH /archive/matchs/:id
// Body : {"hors_competition": true}
// Pour l'instant ne gère que le champ hors_competition (extensible)
```

#### Route

```go
// api/routes.go — dans le bloc archive
archive.PATCH("/matchs/:id", middleware.RequireWritePermission([]string{"archive"}), archiveHandler.UpdateMatch)
```

### Détails d'implémentation

#### Migration SQL

```sql
-- 000035_add_matchs_hors_competition.up.sql
ALTER TABLE matchs ADD COLUMN hors_competition BOOLEAN NOT NULL DEFAULT FALSE;
```

```sql
-- 000035_add_matchs_hors_competition.down.sql
ALTER TABLE matchs DROP COLUMN hors_competition;
```

#### Match model

1. Ajouter `HorsCompetition bool` au struct `Match` avec le json tag `hors_competition`
2. Mettre à jour `scanMatch` pour inclure `hors_competition` (nouveau champ dans SELECT)
3. Ajouter `hors_competition` dans les INSERT/UPDATE de `Create`, `Update`, `FindByID`, `FindByBOID`
4. Ajouter méthode `SetHorsCompetition(id int, val bool) (*Match, error)` qui UPDATE `matchs SET hors_competition = $1`

⚠️ **Important** : Les SELECTs existants doivent être mis à jour pour inclure `hors_competition`. Vérifie toutes les occurrences de `SELECT ... FROM matchs` et `RETURNING` dans `match.go`.

#### `ComputeBOWinner` (bo.go:231-257)

Ajouter un filtre dans la boucle `for _, m := range matchs` pour ignorer les matchs où `m.HorsCompetition == true`.

```go
for _, m := range matchs {
    if m.HorsCompetition {
        continue  // ne compte pas dans le résultat du BO
    }
    if m.VainqueurID != nil {
        // ...
    }
}
```

#### `GetStandings` (div.go:268-387)

La requête existante compte tous les matchs (`SELECT m.bo_id, m.vainqueur_id FROM matchs m JOIN bos b ON b.id = m.bo_id WHERE b.ligue_id = $1`). Ajouter `AND m.hors_competition = false` pour ne pas compter les matchs hors compétition dans les résultats BO.

⚠️ **Les card stats et head-to-head stats ne doivent PAS être modifiées** — les matchs hors compétition doivent être comptés dans les stats.

#### Handler — `UpdateMatch`

```go
func (h *ArchiveHandler) UpdateMatch(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_match_id", "invalid match id", http.StatusBadRequest))
        return
    }

    var req updateMatchRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
        return
    }

    var match *models.Match

    if req.HorsCompetition != nil {
        // Utiliser SetHorsCompetition pour mettre à jour uniquement ce champ
        // Le repo a une méthode dédiée : matchRepo.SetHorsCompetition(id, *req.HorsCompetition)
        // OU : charger le match existant, modifier, puis appeler Update
        // Privilégier l'approche directe : SetHorsCompetition sur le repo.
    }

    if match == nil {
        apierror.RespondWithError(c, apierror.NewAPIError("nothing_to_update", "no fields to update", http.StatusBadRequest))
        return
    }

    // Recompute BO winner since hors_competition affects BO result
    _, recErr := h.boRepo.RecomputeWinner(match.BoID)
    if recErr != nil {
        h.logger.Error("failed to recompute BO winner after match update", "error", recErr)
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"match": match})
}
```

Implémentation recommandée pour `SetHorsCompetition` dans `matchRepo` :

```go
func (r *matchRepo) SetHorsCompetition(id int, horsCompetition bool) (*Match, error) {
    m, err := scanMatch(r.db.QueryRow(
        `UPDATE matchs SET hors_competition = $1, updated_at = NOW()
         WHERE id = $2
         RETURNING id, bo_id, map_id, score_equipe1, score_equipe2, forfait, vainqueur_id, commentaires, hors_competition, created_at, updated_at`,
        horsCompetition, id,
    ))
    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrMatchNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to set hors_competition: %w", err)
    }
    return m, nil
}
```

Ajouter `SetHorsCompetition` à l'interface `MatchRepository`.

⚠️ **Important** : Après la modification, appeler `h.boRepo.RecomputeWinner(match.BoID)` car changer le statut hors_compétition affecte le résultat du BO.

#### `recomputeWinner` dans la migration

Ajouter le fichier de migration à la liste `archiveMigrationFiles` dans `archive_test.go` ligne 42 (après `000034_add_ligue_nom.up.sql`).

### Contraintes techniques
- **Framework** : Go standard + Gin
- **Pattern** : Suivre le pattern existant des autres handlers/repos dans `archive.go` et `match.go`
- **Base de données** : PostgreSQL via `database/sql`
- **Migration** : Ajouter le nouveau fichier à la liste dans `archive_test.go` pour les tests d'intégration
- **Tests** : Les tests d'intégration existants `archive_test.go` doivent continuer à passer

### Tests à implémenter (intégration)

Voir Tâche #004 pour la liste complète. Cette tâche doit au minimum s'assurer que les tests existants passent.

### Exemples d'API

#### Toggle hors_competition

```http
PATCH /api/v1/archive/matchs/42/hors-competition
Content-Type: application/json

{"hors_competition": true}

Response 200:
{
  "match": {
    "id": 42,
    "bo_id": 5,
    "map_id": 3,
    "score_equipe1": 13,
    "score_equipe2": 7,
    "forfait": 0,
    "vainqueur_id": 1,
    "commentaires": null,
    "hors_competition": true,
    "created_at": "2026-06-12T10:00:00Z",
    "updated_at": "2026-06-12T12:00:00Z"
  }
}
```

#### Création BO avec match hors_competition

```http
POST /api/v1/archive/bos
Content-Type: application/json

{
  "ligue_id": 1,
  "equipe1_id": 1,
  "equipe2_id": 2,
  "format": "BO3",
  "picks": [
    {"equipe_id": 1, "map_id": 1, "ordre": 1},
    {"equipe_id": 2, "map_id": 2, "ordre": 2},
    {"equipe_id": null, "map_id": 3, "ordre": 3}
  ],
  "matchs": [
    {"map_id": 1, "score_equipe1": 13, "score_equipe2": 7, "hors_competition": true},
    {"map_id": 2, "score_equipe1": 5, "score_equipe2": 13},
    {"map_id": 3, "score_equipe1": 13, "score_equipe2": 10}
  ]
}
```

⚠️ **Note** : le champ `hors_competition` dans `createMatchItem` doit être ajouté.
