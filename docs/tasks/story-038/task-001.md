# Tâche #001 - Story #038 : Modèles Go Match, Ban, Pick + extension BORepository

## Objectif
Créer les modèles Go pour `Match`, `Ban` et `Pick` (bans et picks d'un BO), étendre le repository `BORepository` avec le CRUD complet, et ajouter les méthodes de calcul de vainqueur automatique.

## Contexte
- Story #038 : BO, matchs, bans et picks
- Dépend de : Story #037 (struct `BO` partagée, déjà implémentée dans `internal/models/bo.go`)
- Nécessaire pour : Tâche #002 (Handlers CRUD BO)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Modèles à créer :**

1. **Match** (table `matchs`) : représentation d'un match individuel dans un BO
   - Champs : `id`, `bo_id`, `map_id`, `score_equipe1` (0-100), `score_equipe2` (0-100), `vainqueur_id` (calculé automatiquement, nullable), `commentaires` (nullable), `created_at`, `updated_at`
   - Le vainqueur est déterminé automatiquement : l'équipe avec le plus grand score
   - Si scores égaux → `vainqueur_id` = NULL (match nul, pas de vainqueur)

2. **Ban** (table `bo_bans`) : ban d'une carte par une équipe
   - Champs : `id`, `bo_id`, `equipe_id` → `eva_teams(id)`, `map_id` → `maps(id)`, `ordre` (INT)

3. **Pick** (table `bo_picks`) : pick d'une carte (équipe facultative)
   - Champs : `id`, `bo_id`, `equipe_id` (nullable) → `eva_teams(id)`, `map_id` → `maps(id)`, `ordre` (INT)

**Extension de BORepository :**
- `FindByID(id int) (*BO, error)` — récupère un BO sans ses sous-ressources
- `FindWithDetails(id int) (*BOWithDetails, error)` — récupère un BO avec bans, picks, matchs
- `FindByLigueID(ligueID int) ([]*BO, error)` — déjà existant, ne pas modifier la signature mais remplacer l'implémentation pour inclure le format dans l'INSERT (actuellement toujours 'BO3')
- `FindByLigueIDWithTeams(ligueID int) ([]*BOListItem, error)` — NOUVEAU : jointure avec eva_teams pour retourner equipe1_nom, equipe1_tag, equipe2_nom, equipe2_tag
- `Update(id int, format string) (*BO, error)` — met à jour le format du BO
- `Delete(id int) error` — déjà existant, supprime le BO + cascade sur bans/picks/matchs via FK CASCADE
- `Create(ligueID, equipe1ID, equipe2ID int, format string) (*BO, error)` — MODIFIER la signature pour accepter `format string` au lieu de toujours 'BO3'

**Nouveaux repositories :**
- `MatchRepository` : CRUD pour les matchs d'un BO
  - `Create(boID, mapID, scoreEquipe1, scoreEquipe2 int, commentaires *string) (*Match, error)` — avec calcul automatique de vainqueur_id
  - `FindByBOID(boID int) ([]*Match, error)`
  - `DeleteByBOID(boID int) error` — supprime tous les matchs d'un BO
  - `Delete(id int) error`
  - `Update(id, mapID, scoreEquipe1, scoreEquipe2 int, commentaires *string) (*Match, error)` — recalcule vainqueur_id
- `BanRepository` : CRUD pour les bans d'un BO
  - `Create(boID, equipeID, mapID, ordre int) (*Ban, error)`
  - `FindByBOID(boID int) ([]*Ban, error)`
  - `DeleteByBOID(boID int) error`
- `PickRepository` : CRUD pour les picks d'un BO
  - `Create(boID, mapID, ordre int, equipeID *int) (*Pick, error)`
  - `FindByBOID(boID int) ([]*Pick, error)`
  - `DeleteByBOID(boID int) error`

**Structure `BOWithDetails` :**
```go
type BOWithDetails struct {
    BO
    Bans    []*Ban   `json:"bans"`
    Picks   []*Pick  `json:"picks"`
    Matchs  []*Match `json:"matchs"`
}

// BOListItem pour la liste avec noms d'équipes (via JOIN eva_teams)
type BOListItem struct {
    ID          int    `json:"id"`
    LigueID     int    `json:"ligue_id"`
    Equipe1ID   int    `json:"equipe1_id"`
    Equipe2ID   int    `json:"equipe2_id"`
    Equipe1Nom  string `json:"equipe1_nom"`
    Equipe1Tag  string `json:"equipe1_tag"`
    Equipe2Nom  string `json:"equipe2_nom"`
    Equipe2Tag  string `json:"equipe2_tag"`
    VainqueurID *int   `json:"vainqueur_id,omitempty"`
    Format      string `json:"format"`
}
```

**Calcul automatique du vainqueur du BO :**
- À chaque création/mise à jour des matchs, recalculer le vainqueur du BO
- Méthode utilitaire `func ComputeBOWinner(matchs []*Match, equipe1ID, equipe2ID int) *int`
- Logique : l'équipe avec le plus de matchs gagnés est vainqueur du BO
- Si égalité → vainqueur_id = NULL (pas de vainqueur)

**Cas nominaux :**
- Création d'un match avec scores 100 et 85 → vainqueur_id = equipe1_id
- Création d'un match avec scores 70 et 100 → vainqueur_id = equipe2_id
- Création d'un match avec scores 50 et 50 → vainqueur_id = NULL

**Cas limites :**
- BO3 avec 2 matchs gagnés par equipe1 → vainqueur_id du BO = equipe1_id
- BO5 avec 3 matchs gagnés par equipe2 → vainqueur_id du BO = equipe2_id
- BO avec autant de matchs gagnés de chaque côté → vainqueur_id = NULL
- Match avec score > 100 ou < 0 → erreur de validation (contrainte CHECK en base)
- Commentaire vide/null → OK, pas d'erreur

**Gestion d'erreurs :**
- Match invalide (mauvaises données) → erreur validation
- BO not found → `ErrBONotFound`
- Match not found → nouveau `ErrMatchNotFound`
- Ban not found → nouveau `ErrBanNotFound`
- Pick not found → nouveau `ErrPickNotFound`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/match.go` | Créer | Modèle Match + MatchRepository |
| `internal/models/bo_ban.go` | Créer | Modèle Ban + BanRepository |
| `internal/models/bo_pick.go` | Créer | Modèle Pick + PickRepository |
| `internal/models/bo.go` | Modifier | Ajouter Update, FindByID, FindWithDetails, ComputeBOWinner + modifier signature Create pour accepter format |
| `internal/models/repository.go` | Modifier | Étendre l'interface BORepository + ajouter interfaces MatchRepository, BanRepository, PickRepository |
| `internal/models/errors.go` | Modifier | Ajouter ErrMatchNotFound, ErrBanNotFound, ErrPickNotFound |

### Signatures

```go
// match.go
type Match struct {
    ID           int        `json:"id"`
    BOID         int        `json:"bo_id"`
    MapID        int        `json:"map_id"`
    ScoreEquipe1 int        `json:"score_equipe1"`
    ScoreEquipe2 int        `json:"score_equipe2"`
    VainqueurID  *int       `json:"vainqueur_id,omitempty"`
    Commentaires *string    `json:"commentaires,omitempty"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}

type matchRepo struct {
    db *sql.DB
}

func NewMatchRepository(db *sql.DB) MatchRepository

func scanMatch(scanner interface{ Scan(dest ...interface{}) error }) (*Match, error)
func scanMatchs(rows *sql.Rows) ([]*Match, error)

func (r *matchRepo) Create(boID, mapID, scoreEquipe1, scoreEquipe2 int, commentaires *string) (*Match, error)
func (r *matchRepo) FindByBOID(boID int) ([]*Match, error)
func (r *matchRepo) Update(id, mapID, scoreEquipe1, scoreEquipe2 int, commentaires *string) (*Match, error)
func (r *matchRepo) DeleteByBOID(boID int) error
func (r *matchRepo) Delete(id int) error

// bo_ban.go
type Ban struct {
    ID       int `json:"id"`
    BOID     int `json:"bo_id"`
    EquipeID int `json:"equipe_id"`
    MapID    int `json:"map_id"`
    Ordre    int `json:"ordre"`
}

type banRepo struct {
    db *sql.DB
}

func NewBanRepository(db *sql.DB) BanRepository

func (r *banRepo) Create(boID, equipeID, mapID, ordre int) (*Ban, error)
func (r *banRepo) FindByBOID(boID int) ([]*Ban, error)
func (r *banRepo) DeleteByBOID(boID int) error

// bo_pick.go
type Pick struct {
    ID       int  `json:"id"`
    BOID     int  `json:"bo_id"`
    EquipeID *int `json:"equipe_id,omitempty"`
    MapID    int  `json:"map_id"`
    Ordre    int  `json:"ordre"`
}

type pickRepo struct {
    db *sql.DB
}

func NewPickRepository(db *sql.DB) PickRepository

func (r *pickRepo) Create(boID, mapID, ordre int, equipeID *int) (*Pick, error)
func (r *pickRepo) FindByBOID(boID int) ([]*Pick, error)
func (r *pickRepo) DeleteByBOID(boID int) error

// bo.go — ajouter/modifier
type BOWithDetails struct {
    BO
    Bans   []*Ban   `json:"bans"`
    Picks  []*Pick  `json:"picks"`
    Matchs []*Match `json:"matchs"`
}

// Modifier l'interface BORepository
type BORepository interface {
    Create(ligueID, equipe1ID, equipe2ID int, format string) (*BO, error) // format ajouté (BREAKING CHANGE — mettre à jour div.go)
    Delete(id int) error
    FindByLigueID(ligueID int) ([]*BO, error)
    FindByLigueIDWithTeams(ligueID int) ([]*BOListItem, error)            // NOUVEAU — jointure eva_teams
    FindByID(id int) (*BO, error)                                          // NOUVEAU
    FindWithDetails(id int) (*BOWithDetails, error)                        // NOUVEAU
    Update(id int, format string) (*BO, error)                             // NOUVEAU
    RecomputeWinner(id int) error                                          // NOUVEAU
}

// Repository interfaces (repository.go)
type MatchRepository interface {
    Create(boID, mapID, scoreEquipe1, scoreEquipe2 int, commentaires *string) (*Match, error)
    FindByBOID(boID int) ([]*Match, error)
    Update(id, mapID, scoreEquipe1, scoreEquipe2 int, commentaires *string) (*Match, error)
    DeleteByBOID(boID int) error
    Delete(id int) error
}

type BanRepository interface {
    Create(boID, equipeID, mapID, ordre int) (*Ban, error)
    FindByBOID(boID int) ([]*Ban, error)
    DeleteByBOID(boID int) error
}

type PickRepository interface {
    Create(boID, mapID, ordre int, equipeID *int) (*Pick, error)
    FindByBOID(boID int) ([]*Pick, error)
    DeleteByBOID(boID int) error
}
```

### Contraintes techniques
- **Framework** : Go 1.22+, `database/sql`, PostgreSQL (`github.com/lib/pq` ou `pgx`)
- **Pattern** : Suivre le pattern exact des repositories existants (e.g. `internal/models/bo.go`, `internal/models/eva_team.go`) — avec `scanXxx`, `boRepo` struct, etc.
- **Style** : Respecter la convention d'erreurs du projet (`errors.Is`, variables `var ErrXxx = errors.New(...)`)
- **Calcul vainqueur** : Implémenter `computeWinner(score1, score2 int) *int` qui retourne nil si scores égaux
- **Calcul vainqueur BO** : `func ComputeBOWinner(matchs []*Match, equipe1ID, equipe2ID int) *int` — fonction libre dans `bo.go` (pas une méthode)
- **Ordre des bans/picks** : Les bans et picks sont retournés triés par `ordre ASC`
- **Validation format** : Vérifier que format est "BO3" ou "BO5" dans le handler (pas dans le repository)

### Tests à implémenter

#### Tests unitaires (créer `internal/models/bo_test.go` si pas existant, sinon modifier)
- **Scénario 1** : `ComputeBOWinner` — equipe1 gagne 3-2 → retourne equipe1_id
  - Données : 5 matchs, equipe1 gagne 3, equipe2 gagne 2, equipe1ID=1, equipe2ID=2
  - Résultat attendu : `*int` pointant vers 1
- **Scénario 2** : `ComputeBOWinner` — égalité 2-2 → retourne nil
  - Données : 4 matchs, equipe1 gagne 2, equipe2 gagne 2
  - Résultat attendu : nil
- **Scénario 3** : `ComputeBOWinner` — equipe2 gagne 3-0 → retourne equipe2_id
  - Données : 3 matchs, equipe2 gagne les 3
  - Résultat attendu : `*int` pointant vers 2
- **Scénario 4** : `ComputeBOWinner` — 0 match → retourne nil
  - Données : tableau vide
  - Résultat attendu : nil

- **Scénario 5** : Création d'un match avec score 100 vs 50 → vainqueur = equipe1 (intégration DB)
- **Scénario 6** : Création d'un match avec score 40 vs 80 → vainqueur = equipe2 (intégration DB)
- **Scénario 7** : Création d'un match avec score 50 vs 50 → vainqueur = NULL (intégration DB)

### Documentation
- Mettre à jour `docs/archives-bo-matches.md` (créé dans la tâche documentation) avec les détails des types

### Exemples d'utilisation

```go
// Création d'un match
match, err := matchRepo.Create(boID, mapID, 100, 85, nil)
// match.VainqueurID == &equipe1ID (car 100 > 85)

// Calcul vainqueur BO
matchs := []*Match{
    {VainqueurID: &e1ID}, {VainqueurID: &e2ID}, {VainqueurID: &e1ID},
}
winner := ComputeBOWinner(matchs, e1ID, e2ID)
// winner == &e1ID (car equipe1 a 2 victoires, equipe2 a 1)
```
