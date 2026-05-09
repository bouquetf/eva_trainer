# Tâche #001 - Story #007 : Split Model and Repository

## Objectif

Créer le modèle de données `Split`, son repository complet avec les méthodes CRUD et de résolution des noms de cartes, ainsi que les erreurs associées.

## Contexte

- Story #007 : `docs/stories/story-007.md` - Split Management and Filtering
- Dépend de : Story #000 (la table `splits` existe déjà dans `000001_initial_schema.up.sql`)
- Nécessaire pour : Tâche #002 (endpoints API), Tâche #003 (tests repository), Tâche #004 (tests handler)
- Dépendances indirectes : La `MapRepository` (Story #005) est nécessaire pour `FindMapsBySplitID`
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Modèle Split :**
Créer le struct `Split` dans `internal/models/split.go`. La table `splits` existe déjà dans la migration `000001_initial_schema.up.sql` avec les colonnes :
- `id SERIAL PRIMARY KEY`
- `name VARCHAR(255) NOT NULL`
- `map_names TEXT[] NOT NULL` -- liste des noms de cartes "eva" jouables
- `start_date DATE` (nullable)
- `end_date DATE` (nullable)

**Partie 2 - SplitRepository :**
Définir l'interface `SplitRepository` dans `internal/models/repository.go` et son implémentation `splitRepo` dans `internal/models/split.go` avec les méthodes suivantes :

- `Create(name string, mapNames []string, startDate, endDate *time.Time) (*Split, error)` : Crée un nouveau split. Valide que le nom n'est pas vide, que map_names n'est pas vide, et que les dates sont cohérentes (si les deux sont fournies, start_date <= end_date).
- `FindByID(id int) (*Split, error)` : Récupère un split par son ID. Retourne `ErrSplitNotFound` si inexistant.
- `FindAll() ([]*Split, error)` : Récupère tous les splits, triés par nom ASC.
- `Update(id int, name string, mapNames []string, startDate, endDate *time.Time) (*Split, error)` : Met à jour un split existant. Applique les mêmes validations que Create. Retourne `ErrSplitNotFound` si inexistant.
- `Delete(id int) error` : Supprime un split par son ID. Retourne `ErrSplitNotFound` si inexistant.
- `FindMapsBySplitID(splitID int) ([]*Map, error)` : Résout les noms de cartes du split en cartes réelles. Requiert : trouver le split, puis `SELECT * FROM maps WHERE name = ANY($1)` avec la liste `map_names` du split. Retourne toutes les versions (y compris obsolètes) des cartes dont le nom est dans le split. Retourne une slice vide si le split n'a pas de cartes correspondantes. Retourne `ErrSplitNotFound` si le split n'existe pas.

**Cas nominaux :**
- Création d'un split avec nom, map_names, start_date et end_date valides → split créé
- Création d'un split avec uniquement nom et map_names (sans dates) → split créé, start_date et end_date NULL
- Récupération de tous les splits → liste triée par nom
- Mise à jour du nom d'un split → modification réussie
- Suppression d'un split → suppression réussie
- Résolution des maps d'un split → toutes les maps correspondantes retournées
- Résolution avec des noms sans carte correspondante → slice vide (pas d'erreur)

**Cas limites :**
- Création avec un nom vide → erreur `ErrSplitNameRequired`
- Création avec map_names vide → erreur `ErrSplitMapNamesRequired`
- Création avec start_date > end_date → erreur `ErrInvalidDateRange`
- Recherche d'un split inexistant → `ErrSplitNotFound`
- Mise à jour d'un split inexistant → `ErrSplitNotFound`
- Suppression d'un split inexistant → `ErrSplitNotFound`
- Résolution des maps d'un split inexistant → `ErrSplitNotFound`
- Aucun split en base → `FindAll` retourne slice vide (pas nil)

**Gestion d'erreurs :**
- Erreur DB → propagation au caller avec contexte (`fmt.Errorf` avec `%w`)
- Nom vide → `ErrSplitNameRequired`
- MapNames vide → `ErrSplitMapNamesRequired`
- Dates incohérentes → `ErrInvalidDateRange`
- Split introuvable → `ErrSplitNotFound`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/split.go` | Créer | Struct Split, SplitRepository interface, implémentation splitRepo |
| `internal/models/errors.go` | Modifier | Ajouter les erreurs liées aux splits |
| `internal/models/repository.go` | Modifier | Ajouter l'interface SplitRepository |

### Signatures

```go
// internal/models/split.go
package models

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/lib/pq"
)

type Split struct {
    ID        int        `json:"id"`
    Name      string     `json:"name"`
    MapNames  []string   `json:"map_names"`
    StartDate *time.Time `json:"start_date,omitempty"`
    EndDate   *time.Time `json:"end_date,omitempty"`
}

type splitRepo struct {
    db *sql.DB
}

func NewSplitRepository(db *sql.DB) SplitRepository {
    return &splitRepo{db: db}
}
```

```go
// Ajouts dans internal/models/errors.go
var (
    ErrSplitNameRequired    = errors.New("split name is required")
    ErrSplitMapNamesRequired = errors.New("split must have at least one map name")
    ErrInvalidDateRange     = errors.New("start date must be before or equal to end date")
    ErrSplitNotFound        = errors.New("split not found")
)
```

```go
// Ajouts dans internal/models/repository.go
type SplitRepository interface {
    Create(name string, mapNames []string, startDate, endDate *time.Time) (*Split, error)
    FindByID(id int) (*Split, error)
    FindAll() ([]*Split, error)
    Update(id int, name string, mapNames []string, startDate, endDate *time.Time) (*Split, error)
    Delete(id int) error
    FindMapsBySplitID(splitID int) ([]*Map, error)
}
```

### Méthodes détaillées

#### scanSplit
```go
func scanSplit(scanner interface {
    Scan(dest ...interface{}) error
}) (*Split, error) {
    s := &Split{}
    var startDate, endDate sql.NullTime
    var mapNames []string

    err := scanner.Scan(&s.ID, &s.Name, pq.Array(&mapNames), &startDate, &endDate)
    if err != nil {
        return nil, err
    }

    s.MapNames = mapNames
    if startDate.Valid {
        s.StartDate = &startDate.Time
    }
    if endDate.Valid {
        s.EndDate = &endDate.Time
    }

    return s, nil
}
```

#### Create
```go
func (r *splitRepo) Create(name string, mapNames []string, startDate, endDate *time.Time) (*Split, error)
```
1. Valider que `name` n'est pas vide → `ErrSplitNameRequired`
2. Valider que `mapNames` n'est pas vide → `ErrSplitMapNamesRequired`
3. Valider les dates : si les DEUX sont non-nil et `startDate.After(*endDate)` → `ErrInvalidDateRange`. Si une seule date est fournie (startDate sans endDate ou inversement), c'est accepté sans erreur (l'autre reste NULL en base).
4. Insérer : `INSERT INTO splits (name, map_names, start_date, end_date) VALUES ($1, $2, $3, $4) RETURNING id, name, map_names, start_date, end_date`
5. Utiliser `pq.Array(&mapNames)` pour le paramètre `$2`
6. Retourner le split créé

#### FindByID
```go
func (r *splitRepo) FindByID(id int) (*Split, error)
```
- `SELECT id, name, map_names, start_date, end_date FROM splits WHERE id = $1`
- Si `sql.ErrNoRows` → `ErrSplitNotFound`

#### FindAll
```go
func (r *splitRepo) FindAll() ([]*Split, error)
```
- `SELECT id, name, map_names, start_date, end_date FROM splits ORDER BY name ASC`
- Retourner slice vide (pas nil) si aucun résultat

#### Update
```go
func (r *splitRepo) Update(id int, name string, mapNames []string, startDate, endDate *time.Time) (*Split, error)
```
1. Vérifier que le split existe (ou gérer via RowsAffected)
2. Appliquer les mêmes validations que Create
3. `UPDATE splits SET name = $1, map_names = $2, start_date = $3, end_date = $4 WHERE id = $5 RETURNING id, name, map_names, start_date, end_date`
4. Si aucune ligne affectée → `ErrSplitNotFound`

#### Delete
```go
func (r *splitRepo) Delete(id int) error
```
- `DELETE FROM splits WHERE id = $1`
- Vérifier `RowsAffected`, si 0 → `ErrSplitNotFound`

#### FindMapsBySplitID
```go
func (r *splitRepo) FindMapsBySplitID(splitID int) ([]*Map, error)
```
1. Récupérer le split via `FindByID(splitID)` (retourne `ErrSplitNotFound` si inexistant)
2. Si `len(split.MapNames) == 0` → retourner slice vide
3. Exécuter : `SELECT id, name, added_at, obsolete_of, is_obsolete FROM maps WHERE name = ANY($1)` avec `pq.Array(split.MapNames)`
4. Scanner les résultats avec `scanMap` (déjà existant dans `map.go`)
5. Retourner les maps trouvées (slice vide si aucune, pas nil)

### Contraintes techniques

- **Package naming** : `models`. Import path : `eva_trainer/internal/models`.
- **Connexion DB** : Prendre `*sql.DB` dans le constructeur.
- **Arrays PostgreSQL** : Utiliser `pq.Array()` de `github.com/lib/pq` (déjà en dépendance indirecte) pour l'insertion et la lecture des `TEXT[]`.
- **Nullables** : `start_date` et `end_date` sont des `*time.Time` dans le struct, scannés via `sql.NullTime`.
- **Pattern repository** : Suivre exactement le pattern des repositories existants (MapRepository, TeamRepository, etc.) : interface publique, struct privée, constructeur public `New*Repository`.
- **Réutilisation** : `FindMapsBySplitID` doit réutiliser `scanMap` de `internal/models/map.go` pour scanner les résultats de maps.
- **Validation** : Les validations se limitent aux règles métier (nom non vide, au moins un nom de carte, dates cohérentes).

### Tests à implémenter

Les tests seront créés dans la Tâche #003.

### Documentation

Aucune documentation à créer pour cette tâche.
