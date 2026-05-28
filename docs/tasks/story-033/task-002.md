# Tâche #002 - Story #033 : Modèle Go LineOfFire (types, repository, errors)

## Objectif
Créer les types Go pour les lignes de tir, le repository complet avec CRUD, les erreurs sentinelles, et mettre à jour l'interface `Repository`.

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Tâche #001 (migration DB)
- Nécessaire pour : Tâche #003 (handler)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Types Go (`internal/models/line_of_fire.go`)**

Définir les types suivants :

```go
type Cone struct {
    Origin   Point `json:"origin"`
    LeftEnd  Point `json:"left_end"`
    RightEnd Point `json:"right_end"`
}

type ConeData struct {
    Cones []Cone `json:"cones"`
}

type LineOfFire struct {
    ID          int       `json:"id"`
    TeamID      int       `json:"team_id"`
    MapID       int       `json:"map_id"`
    Name        string    `json:"name"`
    PolygonData ConeData  `json:"polygon_data"`
    OrderIndex  int       `json:"order_index"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type TemplateLineOfFire struct {
    ID          int       `json:"id"`
    TemplateID  int       `json:"template_id"`
    PolygonData ConeData  `json:"polygon_data"`
    Name        string    `json:"name"`
    OrderIndex  int       `json:"order_index"`
}

type LineOfFireVersion struct {
    ID            int             `json:"id"`
    LineOfFireID  int             `json:"line_of_fire_id"`
    EditorID      *int            `json:"editor_id,omitempty"`
    VersionData   json.RawMessage `json:"version_data"`
    CreatedAt     time.Time       `json:"created_at"`
}
```

**Interface `LineOfFireRepository` (`internal/models/repository.go`)**

Ajouter l'interface :

```go
type LineOfFireRepository interface {
    Create(teamID, mapID int, name string, data ConeData) (*LineOfFire, error)
    FindByID(id int) (*LineOfFire, error)
    FindByTeamAndMap(teamID, mapID int) ([]*LineOfFire, error)
    UpdateName(id int, name string) error
    UpdateOrder(id int, orderIndex int) error
    UpdatePolygonData(id int, data ConeData) error
    Delete(id int) error
    IsNameTaken(teamID, mapID int, name string) (bool, error)
}
```

**Interface `LineOfFireVersionRepository` (`internal/models/repository.go`)**

```go
type LineOfFireVersionRepository interface {
    Save(lineOfFireID int, editorID *int, snapshot json.RawMessage) (*LineOfFireVersion, error)
    FindByLineOfFire(lineOfFireID int) ([]*LineOfFireVersion, error)
    GetLatestForLineOfFire(lineOfFireID int) (*LineOfFireVersion, error)
    DeleteNewerThan(lineOfFireID int, olderThan time.Time) error
}
```

**Erreurs sentinelles (`internal/models/errors.go`)**

Ajouter :
```go
var (
    ErrLineOfFireNotFound    = errors.New("line of fire not found")
    ErrLineOfFireNameTooLong = errors.New("line of fire name exceeds maximum length of 250 characters")
    ErrLineOfFireNameTaken   = errors.New("line of fire name already exists for this team on this map")
    ErrLineOfFireNameRequired = errors.New("line of fire name is required")
    ErrConeOriginRequired    = errors.New("cone must have an origin point")
)
```

**Repository `lineOfFireRepo` (`internal/models/line_of_fire.go`)**

Implémenter `LineOfFireRepository` :

- `Create(teamID, mapID int, name string, data ConeData) (*LineOfFire, error)` : INSERT INTO lines_of_fire. Valider que name n'est pas vide (max 250), que data contient au moins 1 cône, que chaque cône a origin non nul. Si le nom existe déjà → ErrLineOfFireNameTaken.
- `FindByID(id int) (*LineOfFire, error)` : SELECT avec geomapping. If sql.ErrNoRows → ErrLineOfFireNotFound.
- `FindByTeamAndMap(teamID, mapID int) ([]*LineOfFire, error)` : SELECT ... ORDER BY order_index. Retourner slice vide si aucune.
- `UpdateName(id int, name string) error` : UPDATE ... name. Valider longueur.
- `UpdateOrder(id int, orderIndex int) error` : UPDATE ... order_index.
- `UpdatePolygonData(id int, data ConeData) error` : UPDATE ... polygon_data (maréchal en JSON).
- `Delete(id int) error` : DELETE. 0 rows → ErrLineOfFireNotFound.
- `IsNameTaken(teamID, mapID int, name string) (bool, error)` : COUNT avec JOIN sur team_id et map_id.

Pour la sérialisation JSON de `ConeData` en base : stocker en tant que `json.RawMessage` (JSONB). Le scan/unmarshal se fait via `json.Unmarshal` dans `scanPolygonData`-like helper.

**Repository `lineOfFireVersionRepo`**

Implémenter `LineOfFireVersionRepository` avec des requêtes SQL directes sur `line_of_fire_versions`.

**Mise à jour du fichier de setup_test.go**

Ajouter les migrations :
- `"000020_create_lines_of_fire.up.sql"`
- `"000021_create_template_lines_of_fire.up.sql"`
- `"000022_add_line_of_fire_id_to_edit_history.up.sql"`
- `"000023_create_line_of_fire_versions.up.sql"`
- `"000024_add_line_of_fire_permissions.up.sql"`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/line_of_fire.go` | Créer | Types Go + repository |
| `internal/models/repository.go` | Modifier | Ajouter interfaces LineOfFireRepository + LineOfFireVersionRepository |
| `internal/models/errors.go` | Modifier | Ajouter erreurs sentinelles |
| `internal/models/setup_test.go` | Modifier | Ajouter migrations |

### Signatures

```go
func NewLineOfFireRepository(db *sql.DB) LineOfFireRepository
func NewLineOfFireVersionRepository(db *sql.DB) LineOfFireVersionRepository

// Helpers
func marshalConeData(data ConeData) ([]byte, error)
func unmarshalConeData(data []byte) (*ConeData, error)
```

### Contraintes techniques
- **SQL natif** : Utiliser `database/sql` directement, pas d'ORM. Toutes les requêtes SQL sont en dur.
- **Validation** : Le nom est requis (non vide), max 250 caractères. Chaque cône doit avoir origin non nul.
- **Transaction** : `Delete` doit supprimer les versions et l'edit_history associé avant de supprimer la ligne.
- **Tests** : Voir Tâche #009 pour les tests (tests unitaires et intégration).

### Tests à implémenter

Voir Tâche #009 pour les tests complets. Pour cette tâche, les fonctions doivent être testables.

### Documentation

- `docs/lines-of-fire.md` sera créé dans la Tâche #010.
