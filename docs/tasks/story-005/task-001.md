# Tâche #001 - Story #005 : Modèle Map et Repository (back-end)

## Objectif

Créer le modèle de données `Map`, son repository complet avec les méthodes CRUD et de gestion d'obsolescence, ainsi que les erreurs associées.

## Contexte

- Story #005 : `docs/stories/story-005.md` - Map Management (Admin)
- Dépend de : Story #000 (migration DB : table `maps` existe déjà dans `000001_initial_schema.up.sql`)
- Nécessaire pour : Tâche #002 (endpoints API), Tâche #003 (seed), Tâche #005 (tests)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Modèle Map :**
Créer le struct `Map` dans `internal/models/map.go`. La table `maps` existe déjà dans la migration `000001_initial_schema.up.sql` avec les colonnes :
- `id SERIAL PRIMARY KEY`
- `name VARCHAR(255) NOT NULL`
- `webp_blob BYTEA`
- `added_at TIMESTAMP DEFAULT NOW()`
- `obsolete_of INTEGER REFERENCES maps(id) ON DELETE SET NULL`
- `is_obsolete BOOLEAN DEFAULT FALSE`

**Partie 2 - MapRepository :**
Définir l'interface `MapRepository` et son implémentation `mapRepo` avec les méthodes suivantes :

- `Create(name string, blobData []byte) (*Map, error)` : Crée une nouvelle carte avec son blob webp. Vérifie qu'un nom est fourni. Si une carte du même nom existe déjà et n'est pas obsolète, marque l'ancienne comme obsolète avant d'insérer la nouvelle (dans une même transaction).
- `FindByID(id int) (*Map, error)` : Récupère une carte par son ID. Retourne `ErrMapNotFound` si inexistante.
- `FindAll() ([]*Map, error)` : Récupère toutes les cartes, triées par date d'ajout DESC.
- `FindByName(name string) ([]*Map, error)` : Récupère toutes les cartes d'un nom donné (utile pour trouver les doublons). Triées par date d'ajout DESC.
- `FindByNameLatest(name string) (*Map, error)` : Récupère la version la plus récente (non obsolète si possible) d'une carte par son nom. Retourne `ErrMapNotFound` si aucune carte trouvée.
- `MarkObsolete(mapID int, replacedByMapID int) error` : Marque une carte comme obsolète. Définit `obsolete_of = replacedByMapID` et `is_obsolete = TRUE`.
- `IsMapObsolete(mapID int) (bool, error)` : Vérifie si une carte est marquée obsolète. Retourne `(false, nil)` si la carte n'existe pas (pas d'erreur).
- `FindObsoleteWithoutZones() ([]*Map, error)` : Récupère toutes les cartes obsolètes qui n'ont aucune zone associée dans la table `zones`. Requête : `SELECT m.* FROM maps m LEFT JOIN zones z ON m.id = z.map_id WHERE m.is_obsolete = TRUE AND z.id IS NULL`.
- `Delete(mapID int) error` : Supprime une carte par son ID. Retourne `ErrMapNotFound` si inexistante.
- `GetBlob(mapID int) ([]byte, error)` : Récupère uniquement le blob webp d'une carte. Retourne `ErrMapNotFound` si inexistante.

**Partie 3 - Implémentation de `Create` avec obsolescence automatique (mécanisme central) :**
La méthode `Create` doit :
1. Valider que `name` n'est pas vide
2. Si `name` est valide, chercher la version non-obsolète existante via `FindByNameLatest`
3. Démarrer une transaction
4. Insérer la nouvelle carte
5. Si une ancienne version non-obsolète existe :
   - Marquer l'ancienne comme obsolète (`obsolete_of = newID`, `is_obsolete = TRUE`)
6. Commit
7. Retourner la nouvelle carte

**Cas nominaux :**
- Création d'une carte avec un nom unique → carte créée, `is_obsolete = FALSE`
- Création d'une carte avec un nom déjà existant → ancienne carte marquée `is_obsolete = TRUE`, nouvelle carte créée
- Récupération de toutes les cartes → liste complète
- Suppression d'une carte existante → suppression réussie

**Cas limites :**
- Création avec un nom vide → erreur `ErrMapNameRequired`
- Recherche d'une carte inexistante → `ErrMapNotFound`
- Suppression d'une carte inexistante → `ErrMapNotFound`
- Aucune carte en base → `FindAll` retourne slice vide (pas nil)
- Aucune carte obsolète sans zone → `FindObsoleteWithoutZones` retourne slice vide

**Gestion d'erreurs :**
- Erreur DB → propagation au caller avec contexte (fmt.Errorf avec %w)
- Nom vide → `ErrMapNameRequired`
- Carte introuvable → `ErrMapNotFound`
- Transaction échouée → rollback + erreur

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/map.go` | Créer | Struct Map, MapRepository interface, implémentation mapRepo |
| `internal/models/repository.go` | Modifier | Ajouter l'interface MapRepository |
| `internal/models/errors.go` | Modifier | Ajouter les erreurs liées aux maps |

### Signatures

```go
// internal/models/map.go
package models

type Map struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	WebpBlob   []byte    `json:"-"`
	AddedAt    time.Time `json:"added_at"`
	ObsoleteOf *int      `json:"obsolete_of,omitempty"`
	IsObsolete bool      `json:"is_obsolete"`
}

type MapRepository interface {
	Create(name string, blobData []byte) (*Map, error)
	FindByID(id int) (*Map, error)
	FindAll() ([]*Map, error)
	FindByName(name string) ([]*Map, error)
	FindByNameLatest(name string) (*Map, error)
	MarkObsolete(mapID int, replacedByMapID int) error
	IsMapObsolete(mapID int) (bool, error)
	FindObsoleteWithoutZones() ([]*Map, error)
	Delete(mapID int) error
	GetBlob(mapID int) ([]byte, error)
}

type mapRepo struct {
	db *sql.DB
}

func NewMapRepository(db *sql.DB) MapRepository {
	return &mapRepo{db: db}
}
```

```go
// Ajouts dans internal/models/errors.go
var (
	ErrMapNameRequired = errors.New("map name is required")
	ErrMapNotFound     = errors.New("map not found")
)
```

```go
// Ajout dans internal/models/repository.go
type MapRepository interface {
	Create(name string, blobData []byte) (*Map, error)
	FindByID(id int) (*Map, error)
	FindAll() ([]*Map, error)
	FindByName(name string) ([]*Map, error)
	FindByNameLatest(name string) (*Map, error)
	MarkObsolete(mapID int, replacedByMapID int) error
	IsMapObsolete(mapID int) (bool, error)
	FindObsoleteWithoutZones() ([]*Map, error)
	Delete(mapID int) error
	GetBlob(mapID int) ([]byte, error)
}
```

### Contraintes techniques

- **Package naming** : `models`. Import path : `eva_trainer/internal/models`.
- **Connexion DB** : Prendre `*sql.DB` dans le constructeur. La connexion est disponible via `database.DB()` dans `pkg/database`.
- **Transactions** : La méthode `Create` doit utiliser une transaction pour garantir l'atomicité de l'insertion + marquage d'obsolescence. Utiliser le même pattern que `teamRepo.Create`.
- **Nullables** : `ObsoleteOf` est un `*int` (nullable en DB). Utiliser `sql.NullInt32` pour le scan et convertir en `*int`.
- **Blob** : `WebpBlob` est un `[]byte` mappé à `BYTEA` en PostgreSQL. Le champ est masqué en JSON avec `json:"-"` car le blob n'est pas renvoyé dans les listes (uniquement via l'endpoint dédié).
- **Pattern repository** : Suivre exactement le pattern des repositories existants (TeamRepository, UserRepository, etc.) : interface publique, struct privée, constructeur public `New*Repository`.
- **Validation** : La validation du nom de carte se limite à vérifier qu'il n'est pas vide (longueur max gérée par la DB).

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
