# Tâche #001 - Story #040 : Modèle et repository HeadToHeadStats

## Objectif
Créer le modèle de données et le repository pour le calcul des statistiques par carte filtrées par paire d'équipes (head-to-head : victoires, défaites, winrate, scores moyen et médian pour les 2 équipes sur chaque carte).

## Contexte
- Story #040 : `docs/stories/story-040.md`
- Dépend de : Story #039 (CardStats) — déjà livrée
- Nécessaire pour : Tâche #002 (Handler API)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un nouveau fichier `internal/models/head_to_head_stats.go` contenant :

1. **Structures de données** :
   - `HeadToHeadTeamStats` : statistiques d'une équipe sur une carte dans le contexte d'une confrontation (ne contient pas d'opponents, car l'adversaire est fixe : c'est l'autre équipe de la paire)
   - `HeadToHeadMapStats` : statistiques d'une carte pour les 2 équipes qui s'affrontent

2. **Interface `HeadToHeadStatsRepository`** avec une méthode :
   - `Compute(team1ID, team2ID int) ([]*HeadToHeadMapStats, error)` : calcule les stats des matchs entre team1 et team2, groupées par carte

3. **Implémentation `headToHeadStatsRepo`** qui :
   - Récupère tous les matchs où les 2 équipes spécifiées se sont affrontées (via BO)
   - Calcule pour chaque carte : victoires, défaites, winrate, score moyen, score médian pour **les 2 équipes**
   - Gère les matchs sans vainqueur (draw) — ils ne comptent ni en victoire ni en défaite

**Algorithme de calcul :**

Pour une paire d'équipes eva (team1ID, team2ID) :
1. Requête SQL : récupérer tous les matchs des BOs où les 2 équipes s'affrontent (peu importe qui est equipe1/equipe2 dans le BO)
2. Pour chaque match, déterminer :
   - Si `equipe1_id == team1ID` : `team1Score = score_equipe1`, `team2Score = score_equipe2`
   - Si `equipe1_id == team2ID` : `team1Score = score_equipe2`, `team2Score = score_equipe1`
   - `team1Win = (vainqueur_id == team1ID)`, `team2Win = (vainqueur_id == team2ID)`
   - Draw : vainqueur_id est NULL
3. Grouper par `map_id` et agréger dans HeadToHeadMapStats
4. Winrate = wins / (wins + losses) — 0 si wins + losses == 0
5. Médiane : trier les scores, prendre la valeur centrale (ou moyenne des 2 centrales si pair)
6. Ne PAS inclure les cartes sans match entre les 2 équipes

**Cas nominaux :**
- 2 équipes qui se sont affrontées sur plusieurs cartes → stats calculées correctement pour les 2 équipes
- Plusieurs matchs sur la même carte → scores agrégés correctement

**Cas limites :**
- 2 équipes qui ne se sont jamais affrontées → retourner slice vide
- Matchs avec forfait (forfait != 0) → exclus du calcul
- Draws → ni victoire ni défaite, mais scores comptent dans les moyennes/médianes
- team1ID == team2ID → retourner ErrInvalidTeamPair (à créer dans errors.go)
- Équipe ID inexistant → ErrEvaTeamNotFound (existant)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/head_to_head_stats.go` | Créer | Modèle + repository |
| `internal/models/head_to_head_stats_test.go` | Créer | Tests unitaires |
| `internal/models/repository.go` | Modifier | Ajouter l'interface `HeadToHeadStatsRepository` |
| `internal/models/errors.go` | Modifier | Ajouter `ErrInvalidTeamPair` |

### Signatures

```go
package models

type HeadToHeadTeamStats struct {
    Wins        int     `json:"wins"`
    Losses      int     `json:"losses"`
    Winrate     float64 `json:"winrate"`
    AvgScore    float64 `json:"avg_score"`
    MedianScore float64 `json:"median_score"`
}

type HeadToHeadMapStats struct {
    MapID       int                 `json:"map_id"`
    MapName     string              `json:"map_name"`
    Team1Stats  HeadToHeadTeamStats `json:"team1_stats"`
    Team2Stats  HeadToHeadTeamStats `json:"team2_stats"`
}

type HeadToHeadStatsRepository interface {
    Compute(team1ID, team2ID int) ([]*HeadToHeadMapStats, error)
}

type headToHeadStatsRepo struct {
    db *sql.DB
}

func NewHeadToHeadStatsRepository(db *sql.DB) HeadToHeadStatsRepository
func (r *headToHeadStatsRepo) Compute(team1ID, team2ID int) ([]*HeadToHeadMapStats, error)
```

### Structure de la requête SQL

```sql
SELECT m.map_id, mp.name AS map_name,
       m.score_equipe1, m.score_equipe2,
       m.vainqueur_id,
       b.equipe1_id, b.equipe2_id
FROM matchs m
JOIN bos b ON b.id = m.bo_id
JOIN maps mp ON mp.id = m.map_id
WHERE ((b.equipe1_id = $1 AND b.equipe2_id = $2) OR (b.equipe1_id = $2 AND b.equipe2_id = $1))
  AND m.forfait = 0
ORDER BY mp.name ASC, m.id ASC
```

Ne pas inclure les matchs avec forfait (m.forfait != 0).

### Contraintes techniques
- **Pattern** : Suivre exactement le pattern des repositories existants : struct `headToHeadStatsRepo`, constructeur `NewHeadToHeadStatsRepository`, dépendance `sql.DB`
- **Réutilisation** : Utiliser les helpers exportés `ComputeMedian`, `ComputeWinrate` et `average` de `card_stats.go` (dans le même package `models`, accessibles directement)
- **Erreur** : `ErrInvalidTeamPair` = `errors.New("team1 and team2 must be different")` dans `errors.go`
- **Tests** : Coverage minimum 80% pour le nouveau fichier

### Tests à implémenter

#### Tests unitaires (`internal/models/head_to_head_stats_test.go`)

- **Scénario 1 : Compute entre 2 équipes avec matchs sur plusieurs cartes**
  - Données : créer 2 eva teams, 1 ligue, 1 BO avec 3 matchs sur 2 cartes
  - Match 1 : map1, team1 gagne 80-60
  - Match 2 : map1, team2 gagne 55-70
  - Match 3 : map2, team1 gagne 90-50
  - Vérifier 2 stats retournées
  - map1 : team1 wins=1, losses=1, winrate=0.5, avg=67.5, median=67.5 ; team2 wins=1, losses=1, winrate=0.5, avg=65.0, median=65.0
  - map2 : team1 wins=1, losses=0, winrate=1.0, avg=90.0, median=90.0 ; team2 wins=0, losses=1, winrate=0.0, avg=50.0, median=50.0

- **Scénario 2 : Compute avec équipes jamais affrontées**
  - Données : 2 eva teams sans BO commun
  - Attendu : slice vide

- **Scénario 3 : Compute avec team1 == team2**
  - Attendu : erreur `ErrInvalidTeamPair`

- **Scénario 4 : Compute avec match nul (draw)**
  - Données : 1 match draw 60-60 sur map1 (vainqueur_id = NULL)
  - Attendu : wins=0, losses=0, winrate=0.0 pour les 2 équipes, avg=60.0, median=60.0

- **Scénario 5 : Compute avec matchs forfait exclus**
  - Données : 1 match normal + 1 match forfait (forfait != 0) sur la même carte
  - Attendu : le match forfait n'est pas compté

**Note sur les tests :** Le tag `//go:build integration` est nécessaire pour les tests qui font des appels DB réels. Utiliser `testutil.SetupTestDB` avec les mêmes migrations que `archive_test.go`.

### Exemples d'utilisation

```go
repo := NewHeadToHeadStatsRepository(db)
stats, err := repo.Compute(1, 2)
// stats = [
//   {
//     MapID: 1,
//     MapName: "cp_metal",
//     Team1Stats: {Wins: 1, Losses: 1, Winrate: 0.5, AvgScore: 67.5, MedianScore: 67.5},
//     Team2Stats: {Wins: 1, Losses: 1, Winrate: 0.5, AvgScore: 65.0, MedianScore: 65.0}
//   }
// ]
```
