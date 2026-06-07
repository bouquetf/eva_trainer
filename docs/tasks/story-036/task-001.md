# Tâche #001 - Story #036 : Migration DB + Model EvaTeam

## Objectif
Créer la table `eva_teams` en base de données et implémenter le modèle Go (struct, repository, interface, erreurs) pour le CRUD des équipes eva.

## Contexte
- Story #036 : `docs/stories/story-036.md`
- Spécification : `docs/specification/features/feedback-archives-matchs.md` (section 2)
- Dépend de : Aucune (migration indépendante)
- Nécessaire pour : Tâche #002, Tâche #003, Tâche #004

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : Migration de base de données

Créer la migration `000030_add_eva_teams` avec la table `eva_teams` :

**Structure de la table `eva_teams` :**
| Champ | Type | Contraintes |
|-------|------|-------------|
| `id` | `SERIAL` | `PRIMARY KEY` |
| `name` | `VARCHAR(255)` | `NOT NULL` |
| `tag` | `VARCHAR(50)` | `NOT NULL` |
| `eva_trainer_team_id` | `INT` | NULL, `REFERENCES teams(id) ON DELETE SET NULL`, `UNIQUE` |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` |

**Contraintes :**
- La contrainte `UNIQUE` sur `eva_trainer_team_id` garantit la relation 1:1 optionnelle
- `ON DELETE SET NULL` : si l'équipe eva-trainer associée est supprimée, le lien devient NULL plutôt que de supprimer l'équipe eva
- Index sur `name` pour la recherche

**Cas nominaux :**
- Migration up : création de la table réussie
- Migration down : suppression de la table réussie (DROP TABLE IF EXISTS CASCADE)

**Cas limites :**
- Ré-exécution de la migration up : idempotence via `IF NOT EXISTS`
- Ré-exécution de la migration down : idempotence via `IF EXISTS`

#### Partie 2 : Modèle Go (EvaTeam)

**Struct `EvaTeam`** dans `internal/models/eva_team.go` :
```go
type EvaTeam struct {
    ID               int       `json:"id"`
    Name             string    `json:"name"`
    Tag              string    `json:"tag"`
    EvaTrainerTeamID *int      `json:"eva_trainer_team_id,omitempty"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

**Interface `EvaTeamRepository`** à ajouter dans `internal/models/repository.go` :
```go
type EvaTeamRepository interface {
    Create(name, tag string, evaTrainerTeamID *int) (*EvaTeam, error)
    FindByID(id int) (*EvaTeam, error)
    FindAll() ([]*EvaTeam, error)
    Update(id int, name, tag string, evaTrainerTeamID *int) (*EvaTeam, error)
    Delete(id int) error
    HasReferences(id int) (bool, error)
}
```

**Implémentation `evaTeamRepo`** dans `internal/models/eva_team.go` :
- Fonction `NewEvaTeamRepository(db *sql.DB) EvaTeamRepository`
- Méthode `scanEvaTeam(scanner) (*EvaTeam, error)` — helper de scan
- Méthode `scanEvaTeams(rows) ([]*EvaTeam, error)` — helper de scan multiple
- `Create` : INSERT ... RETURNING id, created_at, updated_at. Validation : name et tag requis, tag max 50 chars
- `FindByID` : SELECT ... WHERE id = $1. Retourne `ErrEvaTeamNotFound` si 0 rows
- `FindAll` : SELECT ... ORDER BY name ASC
- `Update` : UPDATE ... SET name=$1, tag=$2, eva_trainer_team_id=$3, updated_at=NOW() WHERE id=$4 RETURNING ...
- `Delete` : DELETE FROM eva_teams WHERE id = $1. Vérifie `RowsAffected` → `ErrEvaTeamNotFound` si 0. NOTE : la vérification d'intégrité référentielle (HasReferences) se fait au niveau handler avant d'appeler Delete.
- `HasReferences(id int) (bool, error)` : Vérifie via `information_schema.tables` si des tables avec des FK pointant vers `eva_teams(id)` existent, puis exécute COUNT(*) sur chaque table trouvée. Retourne true si une table référençante contient des lignes liées à l'ID donné.

**Erreurs** à ajouter dans `internal/models/errors.go` :
```go
var (
    ErrEvaTeamNameRequired     = errors.New("eva team name is required")
    ErrEvaTeamTagRequired      = errors.New("eva team tag is required")
    ErrEvaTeamTagTooLong       = errors.New("eva team tag exceeds maximum length of 50 characters")
    ErrEvaTeamNotFound         = errors.New("eva team not found")
    ErrEvaTeamHasReferences    = errors.New("cannot delete eva team: it has associated leagues or matches")
)
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000030_add_eva_teams.up.sql` | Créer | Migration up |
| `db/migrations/000030_add_eva_teams.down.sql` | Créer | Migration down |
| `internal/models/eva_team.go` | Créer | Struct + Repository |
| `internal/models/repository.go` | Modifier | Ajouter interface `EvaTeamRepository` |
| `internal/models/errors.go` | Modifier | Ajouter les erreurs EvaTeam |

### Signatures

```go
// internal/models/eva_team.go

type EvaTeam struct {
    ID               int       `json:"id"`
    Name             string    `json:"name"`
    Tag              string    `json:"tag"`
    EvaTrainerTeamID *int      `json:"eva_trainer_team_id,omitempty"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

func NewEvaTeamRepository(db *sql.DB) EvaTeamRepository

// EvaTeamRepository interface methods
func (r *evaTeamRepo) Create(name, tag string, evaTrainerTeamID *int) (*EvaTeam, error)
func (r *evaTeamRepo) FindByID(id int) (*EvaTeam, error)
func (r *evaTeamRepo) FindAll() ([]*EvaTeam, error)
func (r *evaTeamRepo) Update(id int, name, tag string, evaTrainerTeamID *int) (*EvaTeam, error)
func (r *evaTeamRepo) Delete(id int) error
func (r *evaTeamRepo) HasReferences(id int) (bool, error)
```

```go
// internal/models/errors.go — nouveaux vars
ErrEvaTeamNameRequired
ErrEvaTeamTagRequired
ErrEvaTeamTagTooLong
ErrEvaTeamNotFound
ErrEvaTeamHasReferences
```

```sql
-- 000030_add_eva_teams.up.sql
CREATE TABLE IF NOT EXISTS eva_teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    tag VARCHAR(50) NOT NULL,
    eva_trainer_team_id INT REFERENCES teams(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(eva_trainer_team_id)
);
CREATE INDEX IF NOT EXISTS idx_eva_teams_name ON eva_teams(name);
```

```sql
-- 000030_add_eva_teams.down.sql
DROP TABLE IF EXISTS eva_teams CASCADE;
```

### Contraintes techniques
- **Convention migrations** : Suivre le pattern existant (fichiers `.up.sql` / `.down.sql`, numéro à 6 chiffres `000030`)
- **Idempotence** : Toutes les instructions DDL doivent être idempotentes
- **Validation** : Vérifier `name` non vide, `tag` non vide et ≤ 50 caractères dans le repository (comme `ValidateTeamName`)
- **`ON DELETE SET NULL`** : Si une équipe eva-trainer est supprimée, le lien devient NULL (pas de cascade)
- **Unique sur `eva_trainer_team_id`** : Garantit la relation 1:1
- **Scan helper** : Suivre le pattern `scanLigue()`, `scanLigues()` dans `internal/models/ligue.go`
- **Logger** : Aucun logger nécessaire dans le repository (les erreurs sont propagées)
- **HasReferences** : Implémentation via `information_schema.tables` pour détecter les tables FK référençant `eva_teams(id)`. Pour chaque table existante avec une FK valide, exécuter `SELECT COUNT(*) FROM table WHERE fk_column = $1`. Si COUNT > 0 pour une table, retourner `true`. Tables concernées futures : `equipe_eva_div`, `bos`, `matchs`.

### Tests à implémenter

Les tests sont détaillés dans la Tâche #004. Cette tâche se concentre sur la création du code fonctionnel uniquement.

**Vérification manuelle recommandée :**
```bash
go build ./...
go vet ./internal/models/
```

### Documentation

Mise à jour de `docs/archives-eva-teams.md` (créé dans la Tâche #004) avec :
- Schéma de la table `eva_teams`
- Structure du modèle EvaTeam
