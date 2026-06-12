# Tâche #003 - Story #046 : Backend — découplage picks/matchs dans CreateBO/UpdateBO

## Objectif
Découpler les picks et les matchs dans la création et la modification d'un BO : ajouter un pick ne crée pas de match automatiquement, les tableaux picks et matchs sont indépendants côté backend.

## Contexte
- Story #046 : `docs/stories/story-046.md`
- Dépend de : Tâche #001 (pour le champ `hors_competition` dans `createMatchItem`)
- Nécessaire pour : Tâche #004 (Frontend — le frontend envoie désormais des tableaux dépliés)

## Critères d'acceptation concernés
- `[ ]` Ajouter un pick dans un BO ne crée pas de match automatiquement (ligne 19)
- `[ ]` L'archiviste peut modifier la carte et les scores du match ainsi créé (ligne 21)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Actuellement, le frontend envoie toujours autant de matchs que de picks, index par index. La tâche 003 (frontend) va découpler : les tableaux `picks` et `matchs` pourront avoir des longueurs différentes dans les requêtes `CreateBO` et `UpdateBO`. Le backend doit déjà supporter cette possibilité (et c'est déjà le cas — chaque tableau est traité indépendamment).

Cependant, il faut ajouter le champ `hors_competition` dans les structs `createMatchItem` et `createBORequest`/`updateBORequest` pour permettre la création d'un match avec `hors_competition` depuis ces endpoints.

**Changements réels nécessaires :**

#### 1. Ajouter `hors_competition` à `createMatchItem`

```go
// internal/handlers/archive.go
type createMatchItem struct {
    MapID           int     `json:"map_id" binding:"required"`
    ScoreEquipe1    int     `json:"score_equipe1"`
    ScoreEquipe2    int     `json:"score_equipe2"`
    Forfait         int     `json:"forfait"`
    Commentaires    *string `json:"commentaires"`
    HorsCompetition bool    `json:"hors_competition"`  // NOUVEAU
}
```

#### 2. Mettre à jour `matchRepo.Create` et `matchRepo.Update`

Ajouter le paramètre `horsCompetition bool` à la signature de `Create` et `Update` :

```go
// internal/models/match.go
func (r *matchRepo) Create(boID, mapID, scoreEquipe1, scoreEquipe2, forfait int, commentaires *string, horsCompetition bool) (*Match, error)
```

```go
// internal/models/repository.go
type MatchRepository interface {
    Create(boID, mapID, scoreEquipe1, scoreEquipe2, forfait int, commentaires *string, horsCompetition bool) (*Match, error)
    SetHorsCompetition(id int, horsCompetition bool) (*Match, error)
    // ... reste inchangé
}
```

Dans l'INSERT, ajouter `hors_competition` :
```sql
INSERT INTO matchs (bo_id, map_id, score_equipe1, score_equipe2, forfait, vainqueur_id, commentaires, hors_competition)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, bo_id, map_id, score_equipe1, score_equipe2, forfait, vainqueur_id, commentaires, hors_competition, created_at, updated_at
```

De même pour `Update` :
```go
func (r *matchRepo) Update(id, mapID, scoreEquipe1, scoreEquipe2, forfait int, commentaires *string, horsCompetition bool) (*Match, error)
```

#### 3. Mettre à jour les appels à `matchRepo.Create` dans `CreateBO` et `UpdateBO`

Dans `archive.go`:
```go
// CreateBO (ligne ~1215)
for _, m := range req.Matchs {
    _, err := h.matchRepo.Create(bo.ID, m.MapID, m.ScoreEquipe1, m.ScoreEquipe2, m.Forfait, m.Commentaires, m.HorsCompetition)
    // ...
}

// UpdateBO (ligne ~1420)
for _, m := range req.Matchs {
    _, err := h.matchRepo.Create(id, m.MapID, m.ScoreEquipe1, m.ScoreEquipe2, m.Forfait, m.Commentaires, m.HorsCompetition)
    // ...
}
```

#### 4. Vérifier les appels existants à `h.matchRepo.Create` ailleurs

Chercher toutes les occurrences de `h.matchRepo.Create` ou `matchRepo.Create` dans le codebase (tests, seed, etc.) et les mettre à jour avec le nouveau paramètre `false`.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Modifier | Ajouter `HorsCompetition` à `createMatchItem` |
| `internal/models/match.go` | Modifier | Ajouter `horsCompetition` à `Create` et `Update` |
| `internal/models/repository.go` | Modifier | Ajouter `horsCompetition` à l'interface `MatchRepository` |
| `internal/handlers/archive_test.go` | Modifier | Mettre à jour les appels à `matchRepo.Create` dans les tests |

### Contraintes techniques
- **Rétrocompatibilité** : Le champ `hors_competition` est optionnel dans le JSON (booléen avec zero value = false)
- **Tests** : Tous les tests d'intégration existants doivent passer après les modifications de signature

### Tests

Voir Tâche #005 pour les nouveaux tests. S'assurer que les tests existants `TestArchiveHandler_CreateBO_*`, `TestArchiveHandler_UpdateBO_*` passent.
