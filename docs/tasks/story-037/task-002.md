# Tâche #002 - Story #037 : Backend — Modèle Div + Repository + Classement

## Objectif
Implémenter le modèle Go `Div` et son repository (CRUD, assignation/désassignation d'équipes, et requête de classement).

## Contexte
- Story #037 : `docs/stories/story-037.md`
- Spécification : `docs/specification/features/feedback-archives-matchs.md` (sections 4, 7)
- Dépend de : Tâche #001 (migration `000031` appliquée)
- Nécessaire pour : Tâche #003, Tâche #004, Tâche #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : Modèle Div

Struct `Div` représentant une division de ligue.

**Champs :**
| Champ | Type Go | JSON | Description |
|-------|---------|------|-------------|
| ID | `int` | `id` | Identifiant unique |
| LigueID | `int` | `ligue_id` | FK vers la ligue |
| Nom | `string` | `nom` | Nom personnalisable |

#### Partie 2 : Repository Div

Interface `DivRepository` avec les méthodes suivantes :

- `Create(ligueID int, nom string) (*Div, error)` — Créer une div
- `FindByID(id int) (*Div, error)` — Trouver par ID
- `FindByLigueID(ligueID int) ([]*Div, error)` — Lister les divs d'une ligue
- `Update(id int, nom string) (*Div, error)` — Modifier le nom d'une div
- `Delete(id int) error` — Supprimer une div (cascade les assignations)
- `AssignTeam(divID, equipeEvaID int) error` — Assigner une équipe eva à une div (la contrainte UNIQUE DB garantit l'unicité par ligue)
- `RemoveTeam(divID, equipeEvaID int) error` — Retirer une équipe eva d'une div
- `FindTeamsByDivID(divID int) ([]*EvaTeam, error)` — Lister les équipes eva d'une div
- `GetStandings(divID int) ([]*StandingEntry, error)` — Calculer le classement

**Cas nominaux :**
- Création d'une div avec nom personnalisable
- Liste des divs d'une ligue
- Modification du nom d'une div
- Suppression d'une div (cascade vers `equipe_eva_div`)
- Assignation d'une équipe eva à une div
- Désassignation d'une équipe eva d'une div
- Consultation des équipes d'une div
- Calcul du classement à partir des BOs

**Cas limites :**
- Nom de div vide → rejeté
- Nom en doublon dans une même ligue → rejeté (contrainte unique DB)
- Double assignation de la même équipe à la même div → rejeté (PK composite)
- Assignation d'une équipe déjà dans une autre div de la même ligue → rejeté (contrainte UNIQUE DB sur `equipe_eva_div(equipe_eva_id, ligue_id)`, erreur 23505)
- Div inexistante → `ErrDivNotFound`
- Équipe eva inexistante → `ErrEvaTeamNotFound`
- Aucun BO saisi pour une div → classement vide (slice vide, pas d'erreur)
- `vainqueur_id` NULL → considéré comme nul (draw)

**Gestion d'erreurs :**
| Erreur | Condition | Code HTTP |
|--------|-----------|-----------|
| `ErrDivNameRequired` | Nom vide | 400 |
| `ErrDivNotFound` | ID inconnu | 404 |
| `ErrDivNameExists` | Nom en doublon dans la ligue | 409 |
| `ErrDivTeamAlreadyAssigned` | Équipe déjà dans une autre div de la même ligue | 409 |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/div.go` | Créer | Struct Div + repository |
| `internal/models/errors.go` | Modifier | Ajouter les erreurs Div |
| `internal/models/repository.go` | Modifier | Ajouter interface `DivRepository` |

### Signatures

```go
// internal/models/div.go

type Div struct {
    ID      int    `json:"id"`
    LigueID int    `json:"ligue_id"`
    Nom     string `json:"nom"`
}

type StandingEntry struct {
    TeamID        int    `json:"team_id"`
    TeamName      string `json:"team_name"`
    TeamTag       string `json:"team_tag"`
    BOsPlayed     int    `json:"bos_played"`
    BOWon         int    `json:"bo_won"`
    BODrawn       int    `json:"bo_drawn"`
    BOLost        int    `json:"bo_lost"`
    BOPoints      int    `json:"bo_points"`
    MatchesPlayed int    `json:"matches_played"`
    MatchWon      int    `json:"match_won"`
    MatchDrawn    int    `json:"match_drawn"`
    MatchLost     int    `json:"match_lost"`
    MatchPoints   int    `json:"match_points"`
}

func NewDivRepository(db *sql.DB) DivRepository
```

```go
// internal/models/repository.go — ajouter

type DivRepository interface {
    Create(ligueID int, nom string) (*Div, error)
    FindByID(id int) (*Div, error)
    FindByLigueID(ligueID int) ([]*Div, error)
    Update(id int, nom string) (*Div, error)
    Delete(id int) error
    AssignTeam(divID, equipeEvaID int) error
    RemoveTeam(divID, equipeEvaID int) error
    FindTeamsByDivID(divID int) ([]*EvaTeam, error)
    GetStandings(divID int) ([]*StandingEntry, error)
}
```

```go
// internal/models/errors.go — ajouter

var (
    ErrDivNameRequired        = errors.New("div name is required")
    ErrDivNotFound            = errors.New("div not found")
    ErrDivNameExists          = errors.New("a div with this name already exists in this league")
    ErrDivTeamAlreadyAssigned = errors.New("team is already assigned to another div in this league")
)
```

### Détail d'implémentation

#### `Create`
```sql
INSERT INTO divs (ligue_id, nom) VALUES ($1, $2)
RETURNING id, ligue_id, nom
```
Vérifier `nom != ""`. En cas d'erreur 23505 (unique violation sur `idx_divs_ligue_nom`), retourner `ErrDivNameExists`.

#### `FindByID`
```sql
SELECT id, ligue_id, nom FROM divs WHERE id = $1
```
`sql.ErrNoRows` → `ErrDivNotFound`

#### `FindByLigueID`
```sql
SELECT id, ligue_id, nom FROM divs WHERE ligue_id = $1 ORDER BY nom ASC
```

#### `Update`
```sql
UPDATE divs SET nom = $1 WHERE id = $2
RETURNING id, ligue_id, nom
```
Vérifier `nom != ""`. `sql.ErrNoRows` → `ErrDivNotFound`. Violation 23505 → `ErrDivNameExists`.

#### `Delete`
```sql
DELETE FROM divs WHERE id = $1
```
RowsAffected == 0 → `ErrDivNotFound`

#### `AssignTeam`

La contrainte `UNIQUE (equipe_eva_id, ligue_id)` en base garantit qu'une équipe ne peut être que dans une seule div par ligue.

```sql
INSERT INTO equipe_eva_div (equipe_eva_id, div_id, ligue_id)
VALUES ($1, $2, (SELECT ligue_id FROM divs WHERE id = $2))
```
- La sous-requête récupère `ligue_id` de la div cible pour remplir la colonne dénormalisée
- En cas de violation UNIQUE (équipe déjà dans une autre div de cette ligue), PostgreSQL renvoie l'erreur 23505 → retourner `ErrDivTeamAlreadyAssigned`
- En cas de violation PK (même équipe, même div), 23505 aussi → idempotent, retourner succès (déjà assignée à cette div)
- Pour distinguer les deux cas, on peut vérifier si la div est déjà dans la table, mais l'idempotence est un comportement acceptable

#### `RemoveTeam`
```sql
DELETE FROM equipe_eva_div WHERE equipe_eva_id = $1 AND div_id = $2
```
RowsAffected == 0 → erreur "team not in this div" (retourner `ErrEvaTeamNotFound`).

#### `FindTeamsByDivID`
```sql
SELECT e.id, e.name, e.tag, e.eva_trainer_team_id, e.created_at, e.updated_at
FROM eva_teams e
JOIN equipe_eva_div d ON d.equipe_eva_id = e.id
WHERE d.div_id = $1
ORDER BY e.name ASC
```

#### `GetStandings`

Algorithme en Go (3 requêtes) :

```go
func (r *divRepo) GetStandings(divID int) ([]*StandingEntry, error) {
    // 1. Récupérer les équipes de la div
    teams, err := r.FindTeamsByDivID(divID)
    if err != nil {
        return nil, err
    }

    // 2. Récupérer la ligue de la div
    div, err := r.FindByID(divID)
    if err != nil {
        return nil, err
    }

    // 3. Récupérer tous les BOs de la ligue
    rows, err := r.db.Query(
        `SELECT id, equipe1_id, equipe2_id, vainqueur_id FROM bos WHERE ligue_id = $1`,
        div.LigueID,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query bos: %w", err)
    }
    defer rows.Close()

    type boInfo struct {
        ID          int
        Equipe1ID   int
        Equipe2ID   int
        VainqueurID *int
    }
    var bos []boInfo
    for rows.Next() {
        var b boInfo
        if err := rows.Scan(&b.ID, &b.Equipe1ID, &b.Equipe2ID, &b.VainqueurID); err != nil {
            return nil, fmt.Errorf("failed to scan bo: %w", err)
        }
        bos = append(bos, b)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }

    // 4. Récupérer tous les matchs de ces BOs
    // (un seul query avec IN)
    boIDs := make([]int, len(bos))
    for i, b := range bos {
        boIDs[i] = b.ID
    }

    type matchInfo struct {
        BoID        int
        VainqueurID *int
    }
    matchMap := make(map[int][]matchInfo) // boID -> matchs

    if len(boIDs) > 0 {
        // Construire la requête IN dynamiquement
        // Utiliser fmt avec des placeholders PostgreSQL
        // Ou utiliser un tableau PostgreSQL avec ANY($1)
    }

    // 5. Pour chaque équipe de la div, calculer les stats
    entries := make([]*StandingEntry, 0, len(teams))
    for _, team := range teams {
        entry := &StandingEntry{
            TeamID:   team.ID,
            TeamName: team.Name,
            TeamTag:  team.Tag,
        }

        for _, b := range bos {
            if b.Equipe1ID != team.ID && b.Equipe2ID != team.ID {
                continue
            }
            entry.BOsPlayed++

            if b.VainqueurID != nil && *b.VainqueurID == team.ID {
                entry.BOWon++
                entry.BOPoints++
            } else if b.VainqueurID != nil {
                entry.BOLost++
            } else {
                entry.BODrawn++
            }
        }

        // Match stats
        for _, m := range matchMap {
            // filter by bo in which this team plays
            // ...
        }

        entry.MatchPoints = entry.MatchWon

        entries = append(entries, entry)
    }

    // Trier par BO points desc, puis match points desc
    sort.Slice(entries, func(i, j int) bool {
        if entries[i].BOPoints != entries[j].BOPoints {
            return entries[i].BOPoints > entries[j].BOPoints
        }
        return entries[i].MatchPoints > entries[j].MatchPoints
    })

    return entries, nil
}
```

**Note sur la requête matchs** : Pour éviter la complexité des requêtes IN paramétrées, on peut utiliser `ANY($1)` avec un tableau PostgreSQL :

```go
// pgx supporte nativement les tableaux
// Avec database/sql + pgx driver :
// Passer un []int{} qui sera automatiquement converti en ARRAY[1,2,3]
matchRows, err := r.db.Query(
    `SELECT bo_id, vainqueur_id FROM matchs WHERE bo_id = ANY($1)`,
    pq.Array(boIDs),  // Pour lib/pq
)
```

Ou plus simplement, exécuter une requête par div (pas de IN du tout) :

```go
// Alternative : requête unique avec JOIN
rows, err := r.db.Query(`
    SELECT m.bo_id, m.vainqueur_id
    FROM matchs m
    JOIN bos b ON b.id = m.bo_id
    WHERE b.ligue_id = $1
`, div.LigueID)
```

**Approche recommandée** : Utiliser le JOIN pour une seule requête, plus simple et sans paramètres dynamiques.

```go
func (r *divRepo) GetStandings(divID int) ([]*StandingEntry, error) {
    teams, err := r.FindTeamsByDivID(divID)
    if err != nil {
        return nil, err
    }
    if len(teams) == 0 {
        return []*StandingEntry{}, nil
    }

    div, err := r.FindByID(divID)
    if err != nil {
        return nil, err
    }

    // Tous les BOs de la ligue
    boRows, err := r.db.Query(
        `SELECT id, equipe1_id, equipe2_id, vainqueur_id FROM bos WHERE ligue_id = $1`,
        div.LigueID,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query bos: %w", err)
    }
    defer boRows.Close()

    type boData struct {
        ID          int
        Equipe1ID   int
        Equipe2ID   int
        VainqueurID *int
    }
    var allBos []boData
    boIDSet := make(map[int]bool)
    for boRows.Next() {
        var b boData
        if err := boRows.Scan(&b.ID, &b.Equipe1ID, &b.Equipe2ID, &b.VainqueurID); err != nil {
            return nil, fmt.Errorf("failed to scan bo: %w", err)
        }
        allBos = append(allBos, b)
        boIDSet[b.ID] = true
    }
    if err := boRows.Err(); err != nil {
        return nil, err
    }

    // Tous les matchs de ces BOs
    matchRows, err := r.db.Query(
        `SELECT m.bo_id, m.vainqueur_id
         FROM matchs m
         JOIN bos b ON b.id = m.bo_id
         WHERE b.ligue_id = $1`,
        div.LigueID,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query matchs: %w", err)
    }
    defer matchRows.Close()

    // matchs par BO
    matchCountByBO := make(map[int]int)    // boID -> total matchs
    matchWinsByTeamBO := make(map[int]map[int]int) // teamID -> boID -> wins
    for matchRows.Next() {
        var boID int
        var vainqueurID *int
        if err := matchRows.Scan(&boID, &vainqueurID); err != nil {
            return nil, fmt.Errorf("failed to scan match: %w", err)
        }
        matchCountByBO[boID]++
        if vainqueurID != nil {
            if matchWinsByTeamBO[*vainqueurID] == nil {
                matchWinsByTeamBO[*vainqueurID] = make(map[int]int)
            }
            matchWinsByTeamBO[*vainqueurID][boID]++
        }
    }
    if err := matchRows.Err(); err != nil {
        return nil, err
    }

    // Calcul par équipe
    teamIDs := make(map[int]bool)
    for _, t := range teams {
        teamIDs[t.ID] = true
    }

    entries := make([]*StandingEntry, 0, len(teams))
    for _, team := range teams {
        entry := &StandingEntry{
            TeamID:   team.ID,
            TeamName: team.Name,
            TeamTag:  team.Tag,
        }

        for _, b := range allBos {
            if b.Equipe1ID != team.ID && b.Equipe2ID != team.ID {
                continue
            }
            entry.BOsPlayed++
            if b.VainqueurID != nil && *b.VainqueurID == team.ID {
                entry.BOWon++
            } else if b.VainqueurID != nil {
                entry.BOLost++
            } else {
                entry.BODrawn++
            }

            // Matchs dans ce BO
            totalMatchs := matchCountByBO[b.ID]
            entry.MatchesPlayed += totalMatchs

            wins := matchWinsByTeamBO[team.ID][b.ID]
            entry.MatchWon += wins
        }

        entry.BOPoints = entry.BOWon
        entry.MatchPoints = entry.MatchWon
        entry.MatchLost = entry.MatchesPlayed - entry.MatchWon - entry.MatchDrawn

        entries = append(entries, entry)
    }

    sort.Slice(entries, func(i, j int) bool {
        if entries[i].BOPoints != entries[j].BOPoints {
            return entries[i].BOPoints > entries[j].BOPoints
        }
        if entries[i].MatchPoints != entries[j].MatchPoints {
            return entries[i].MatchPoints > entries[j].MatchPoints
        }
        return entries[i].TeamName < entries[j].TeamName
    })

    return entries, nil
}
```

### Contraintes techniques
- **Framework** : Go standard library + database/sql + pgx
- **Pattern** : Suivre exactement le pattern de `internal/models/eva_team.go` (constructeur `NewXRepository`, méthodes avec `scanX`, requêtes SQL avec `$1`, `$2`)
- **Validation** : Vérifier `nom != ""` dans le repository (comme `ValidateEvaTeamName`)
- **Scan helper** : `scanDiv(scanner)` et `scanDivs(rows)` suivant le pattern `scanEvaTeam`
- **Logging** : Aucun logger nécessaire dans le repository
- **sort** : Utiliser `sort.Slice` du package standard

### Tests à implémenter

Les tests d'intégration pour le repository Div sont décrits dans la Tâche #003. Cette tâche se concentre sur le code fonctionnel uniquement.

**Vérification manuelle :**
```bash
go build ./...
go vet ./internal/models/
```
