# Tâche #001 - Story #039 : Modèle et repository CardStats

## Objectif
Créer le modèle de données et le repository pour le calcul des statistiques par carte (victoires, défaites, winrate, scores moyen et médian, stats par adversaire).

## Contexte
- Story #039 : `docs/stories/story-039.md`
- Dépend de : Story #038 (BO, matchs) — déjà livrée
- Nécessaire pour : Tâche #002 (Handler API)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un nouveau fichier `internal/models/card_stats.go` contenant :

1. **Structures de données** :
   - `CardStats` : statistiques d'une équipe sur une carte
   - `OpponentCardStats` : statistiques contre un adversaire spécifique sur une carte

2. **Interface `CardStatsRepository`** avec une méthode :
   - `Compute(teamID int) ([]*CardStats, error)` : calcule les stats pour une équipe eva donnée

3. **Implémentation `cardStatsRepo`** qui :
   - Récupère tous les matchs où l'équipe a participé (via BO)
   - Calcule pour chaque carte : victoires, défaites, winrate, score moyen, score médian
   - Calcule pour chaque carte + adversaire : score moyen et médian
   - Gère les matchs sans vainqueur (draw) — ils ne comptent ni en victoire ni en défaite

**Algorithme de calcul :**

Pour une équipe eva donnée (teamID) :
1. Requête SQL : `SELECT m.map_id, mp.name AS map_name, m.score_equipe1, m.score_equipe2, m.vainqueur_id, b.equipe1_id, b.equipe2_id` depuis `matchs m JOIN bos b ON b.id = m.bo_id JOIN maps mp ON mp.id = m.map_id WHERE (b.equipe1_id = $1 OR b.equipe2_id = $1)`
2. Pour chaque match, déterminer :
   - `teamScore` = score_equipe1 si equipe1_id == teamID, sinon score_equipe2
   - `opponentID` = equipe2_id si equipe1_id == teamID, sinon equipe1_id
   - `isWin` = vainqueur_id == teamID (true si oui, false si perdu, nil si draw)
3. Grouper par `map_id` et agréger dans CardStats
4. Grouper par `(map_id, opponent_id)` pour les stats par adversaire
5. Winrate = wins / (wins + losses) — 0 si wins + losses == 0
6. Médiane : trier les scores, prendre la valeur centrale (ou moyenne des 2 centrales si pair)

**Cas nominaux :**
- Équipe avec des matchs gagnés et perdus sur plusieurs cartes → stats calculées correctement
- Équipe avec matchs contre plusieurs adversaires sur la même carte → stats par adversaire calculées

**Cas limites :**
- Équipe sans aucun match → retourner slice vide
- Équipe avec seulement des draws → wins=0, losses=0, winrate=0
- Un seul match sur une carte → avg_score = median_score = ce score
- Score à 0 ou 100 → inclus normalement dans les calculs
- Équipe ID inexistant → ErrEvaTeamNotFound

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/card_stats.go` | Créer | Modèle + repository |
| `internal/models/card_stats_test.go` | Créer | Tests unitaires |
| `internal/models/repository.go` | Modifier | Ajouter l'interface `CardStatsRepository` |
| `internal/models/errors.go` | Modifier | Ajouter `ErrCardStatsTeamNotFound` si besoin (utiliser `ErrEvaTeamNotFound` existant) |

### Signatures

```go
package models

type OpponentCardStats struct {
    TeamID      int     `json:"team_id"`
    TeamName    string  `json:"team_name"`
    AvgScore    float64 `json:"avg_score"`
    MedianScore float64 `json:"median_score"`
}

type CardStats struct {
    MapID       int                 `json:"map_id"`
    MapName     string              `json:"map_name"`
    Wins        int                 `json:"wins"`
    Losses      int                 `json:"losses"`
    Winrate     float64             `json:"winrate"`
    AvgScore    float64             `json:"avg_score"`
    MedianScore float64             `json:"median_score"`
    Opponents   []OpponentCardStats `json:"opponents"`
}

type CardStatsRepository interface {
    Compute(teamID int) ([]*CardStats, error)
}

type cardStatsRepo struct {
    db *sql.DB
}

func NewCardStatsRepository(db *sql.DB) CardStatsRepository
func (r *cardStatsRepo) Compute(teamID int) ([]*CardStats, error)

// Helpers exportées pour les tests
func ComputeMedian(scores []int) float64
func ComputeWinrate(wins, losses int) float64
```

### Structure de la requête SQL

Utiliser une seule requête pour récupérer les données brutes, puis agréger en Go :

```sql
SELECT m.map_id, mp.name AS map_name,
       m.score_equipe1, m.score_equipe2,
       m.vainqueur_id,
       b.equipe1_id, b.equipe2_id
FROM matchs m
JOIN bos b ON b.id = m.bo_id
JOIN maps mp ON mp.id = m.map_id
WHERE (b.equipe1_id = $1 OR b.equipe2_id = $1)
  AND m.forfait = 0
ORDER BY mp.name ASC
```

Ne pas inclure les matchs avec forfait (m.forfait != 0).

### Contraintes techniques
- **Pattern** : Suivre exactement le pattern des repositories existants (ex: `bo.go`, `div.go`) : struct `cardStatsRepo`, constructeur `NewCardStatsRepository`, fonction privée `scan…`, pattern `QueryRow`/`Query`
- **Tests** : Coverage minimum 80% pour le nouveau fichier
- **Convention** : Utiliser `sql.DB` comme dépendance, pas de transaction pour les lectures

### Tests à implémenter

#### Tests unitaires (`internal/models/card_stats_test.go`)

- **Scénario 1 : ComputeMedian avec liste impaire**
  - Données : `[]int{10, 20, 30}`
  - Attendu : `20.0`

- **Scénario 2 : ComputeMedian avec liste paire**
  - Données : `[]int{10, 20, 30, 40}`
  - Attendu : `25.0`

- **Scénario 3 : ComputeMedian avec liste vide**
  - Données : `[]int{}`
  - Attendu : `0.0`

- **Scénario 4 : ComputeMedian avec un seul élément**
  - Données : `[]int{42}`
  - Attendu : `42.0`

- **Scénario 5 : ComputeWinrate normal**
  - Données : wins=3, losses=1
  - Attendu : `0.75`

- **Scénario 6 : ComputeWinrate zéro total**
  - Données : wins=0, losses=0
  - Attendu : `0.0`

- **Scénario 7 : ComputeWinrate 100%**
  - Données : wins=5, losses=0
  - Attendu : `1.0`

- **Scénario 8 : Compute intégration (avec DB réelle, tag `integration`)**
  - Créer : split, ligue, 2 eva_teams, BO avec matchs sur 2 cartes différentes, avec différentes combinaisons gagnant/perdant
  - Vérifier les stats calculées pour chaque équipe

**Note sur les tests :** Le tag `integration` est nécessaire car le test fait des appels DB réels. Utiliser `testutil.SetupTestDB` avec les mêmes migrations que `archive_test.go`.

### Exemples d'utilisation

```go
repo := NewCardStatsRepository(db)
stats, err := repo.Compute(1)
// stats = [
//   {
//     MapID: 1,
//     MapName: "cp_metal",
//     Wins: 3,
//     Losses: 1,
//     Winrate: 0.75,
//     AvgScore: 75.5,
//     MedianScore: 78.0,
//     Opponents: [
//       {TeamID: 2, TeamName: "Team B", AvgScore: 72.0, MedianScore: 72.0}
//     ]
//   }
// ]
```
