# Tâche #001 - Story #012 : Modèle ZoneVersion, Repository et Migration (Backend)

## Objectif

Créer le modèle `ZoneVersion`, son repository, et la migration de base de données pour stocker les snapshots complets de zone. Étendre `EditHistoryRepository` avec la méthode de pagination.

## Contexte

- Story #012 : `docs/stories/story-012.md` - Versions de zone, historique et rollback
- Dépend de : Story #009 (modèles Zone, ZonePolygon, ZoneName, EditLock existants)
- Nécessaire pour : Tâche #002 (API endpoints)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Migration DB

Créer une migration `000013_add_zone_versions.up.sql` qui crée la table `zone_versions` :

```sql
CREATE TABLE IF NOT EXISTS zone_versions (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    editor_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    version_data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zone_versions_zone ON zone_versions(zone_id);
CREATE INDEX IF NOT EXISTS idx_zone_versions_editor ON zone_versions(editor_id);
CREATE INDEX IF NOT EXISTS idx_zone_versions_created ON zone_versions(created_at);
```

**Format de `version_data` (JSONB) :**
```json
{
  "polygons": [
    {
      "points": [{"x": 100, "y": 200}, {"x": 300, "y": 400}, {"x": 500, "y": 100}],
      "order_index": 0
    }
  ],
  "name": "top"
}
```

#### Partie 2 - Modèle ZoneVersion

Créer `internal/models/zone_version.go` avec :

```go
type ZoneVersion struct {
	ID          int              `json:"id"`
	ZoneID      int              `json:"zone_id"`
	EditorID    *int             `json:"editor_id,omitempty"`
	VersionData VersionSnapshot  `json:"version_data"`
	CreatedAt   time.Time        `json:"created_at"`
}

type VersionSnapshot struct {
	Polygons []SnapshotPolygon `json:"polygons"`
	Name     string            `json:"name"`
}

type SnapshotPolygon struct {
	Points     []Point `json:"points"`
	OrderIndex int     `json:"order_index"`
}
```

#### Partie 3 - ZoneVersionRepository

Créer les méthodes suivantes dans une interface `ZoneVersionRepository` :

1. **`Save(zoneID int, editorID *int, snapshot VersionSnapshot) (*ZoneVersion, error)`**
   - Crée une nouvelle version avec le snapshot complet de la zone
   - Appelé après chaque batch auto-save réussi (hook dans ZoneHandler.Task #002)

2. **`FindByMap(teamID, mapID int, page, limit int) ([]*ZoneVersion, int, error)`**
   - Retourne les versions de toutes les zones d'une carte, paginées
   - Jointure avec zones pour filtrer par team_id et map_id
   - Tri chronologique inverse (plus récent d'abord)
   - Retourne les zone_versions + le total

3. **`FindByZone(zoneID int, page, limit int) ([]*ZoneVersion, int, error)`**
   - Retourne les versions d'une zone spécifique, paginées
   - Tri chronologique inverse

4. **`FindByID(versionID int) (*ZoneVersion, error)`**
   - Retourne une version par son ID

5. **`GetLatestForZone(zoneID int) (*ZoneVersion, error)`**
   - Retourne la version la plus récente pour une zone

6. **`FindFirstOtherEditorVersion(zoneID, editorID int) (*int, error)`**
   - Retourne l'ID de la première version (la plus ancienne dans le temps) créée par un AUTRE éditeur
   - Utilisé pour déterminer la limite de visibilité de l'éditeur
   - Si l'éditeur a toutes les versions récentes sans interruption → retourne nil
   - Retourne nil si toutes les versions sont de cet éditeur ou s'il n'y a qu'un seul éditeur

7. **`CountByZone(zoneID int) (int, error)`**

#### Partie 4 - Extension de EditHistoryRepository

Ajouter à `EditHistoryRepository` :

1. **`FindByTeamAndMapWithFilters(teamID, mapID int, page, limit int) ([]*EditHistory, int, error)`**
   - Pagination avec tri chronologique inverse
   - Méthode utilitaire pour l'audit trail (utilisé par l'admin)

2. **`FindByZoneWithFilters(zoneID int, page, limit int) ([]*EditHistory, int, error)`**
   - Pagination pour les actions d'une zone spécifique

**Cas nominaux :**
- Sauvegarde de zone (batch) → snapshot créé dans zone_versions
- Liste des versions d'une carte → paginée, triée du plus récent au plus ancien
- Calcul de la limite éditeur → première version d'un autre éditeur identifiée

**Cas limites :**
- Zone sans version → liste vide
- Toutes les versions du même éditeur → FindFirstOtherEditorVersion retourne nil
- Éditeur supprimé → editor_id = NULL, "Utilisateur supprimé"
- Snapshot avec zone sans nom → version_data.name = ""
- Snapshot avec zone sans polygone → version_data.polygons = []

**Gestion d'erreurs :**
- `ErrVersionNotFound` → version introuvable
- `ErrZoneNotFound` → zone introuvable
- Erreur DB → propagation avec contexte

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/zone_version.go` | Créer | Struct ZoneVersion, VersionSnapshot, SnapshotPolygon + ZoneVersionRepository |
| `internal/models/repository.go` | Modifier | Ajouter l'interface ZoneVersionRepository |
| `internal/models/errors.go` | Modifier | Ajouter ErrVersionNotFound |
| `internal/models/edit_history.go` | Modifier | Ajouter FindByTeamAndMapWithFilters, FindByZoneWithFilters |
| `db/migrations/000013_add_zone_versions.up.sql` | Créer | CREATE TABLE zone_versions |
| `db/migrations/000013_add_zone_versions.down.sql` | Créer | DROP TABLE zone_versions |

### Signatures détaillées

```go
// internal/models/zone_version.go
package models

import "time"

type ZoneVersion struct {
	ID          int             `json:"id"`
	ZoneID      int             `json:"zone_id"`
	EditorID    *int            `json:"editor_id,omitempty"`
	VersionData VersionSnapshot `json:"version_data"`
	CreatedAt   time.Time       `json:"created_at"`
}

type VersionSnapshot struct {
	Polygons []SnapshotPolygon `json:"polygons"`
	Name     string            `json:"name"`
}

type SnapshotPolygon struct {
	Points     []Point `json:"points"`
	OrderIndex int     `json:"order_index"`
}

type ZoneVersionRepository interface {
	Save(zoneID int, editorID *int, snapshot VersionSnapshot) (*ZoneVersion, error)
	FindByMap(teamID, mapID int, page, limit int) ([]*ZoneVersion, int, error)
	FindByZone(zoneID int, page, limit int) ([]*ZoneVersion, int, error)
	FindByID(versionID int) (*ZoneVersion, error)
	GetLatestForZone(zoneID int) (*ZoneVersion, error)
	FindFirstOtherEditorVersion(zoneID, editorID int) (*int, error)
	CountByZone(zoneID int) (int, error)
}
```

```go
// Ajouts dans internal/models/edit_history.go
func (r *editHistoryRepo) FindByTeamAndMapWithFilters(teamID, mapID int, page, limit int) ([]*EditHistory, int, error)
func (r *editHistoryRepo) FindByZoneWithFilters(zoneID int, page, limit int) ([]*EditHistory, int, error)
```

```go
// Ajout dans internal/models/errors.go
var ErrVersionNotFound = errors.New("version not found")
```

### Contraintes techniques

- **VersionSnapshot** : Struct Go dédiée, sérialisée en JSONB. Les champs `polygons` et `name` correspondent exactement à l'état de la zone au moment de la sauvegarde.
- **Save** : Appelé après chaque batch auto-save qui modifie la zone. Le snapshot est pris après l'exécution de toutes les actions du batch.
- **FindFirstOtherEditorVersion** : Requête SQL :
  ```sql
  SELECT id FROM zone_versions
  WHERE zone_id = $1 AND editor_id IS NOT NULL AND editor_id != $2
  ORDER BY created_at DESC
  LIMIT 1
  ```
  Retourne l'ID de la version la plus récente (dans le temps) créée par un autre éditeur. Comme l'historique est trié du plus récent au plus ancien, cela donne la "frontière" : l'éditeur voit les versions plus récentes que cette limite.
- **Calcul visibilité éditeur** : Le frontend demande `FindByZone(zoneID, page, limit)` et `FindFirstOtherEditorVersion(zoneID, editorID)`. Si `FindFirstOtherEditorVersion` retourne un ID, l'éditeur ne voit que les versions avec `created_at` après cette version-limite.
- **Pattern repository** : Suivre exactement le pattern des repositories existants : interface publique, struct privée, constructeur public `NewZoneVersionRepository(db *sql.DB)`.
- **Join pour FindByMap** :
  ```sql
  SELECT zv.id, zv.zone_id, zv.editor_id, zv.version_data, zv.created_at
  FROM zone_versions zv
  JOIN zones z ON z.id = zv.zone_id
  WHERE z.team_id = $1 AND z.map_id = $2
  ORDER BY zv.created_at DESC
  LIMIT $3 OFFSET $4
  ```
- **Count pour FindByMap** : `SELECT COUNT(*) FROM zone_versions zv JOIN zones z ON z.id = zv.zone_id WHERE z.team_id = $1 AND z.map_id = $2`
- **VersionData en Go** : Utiliser `json.Marshal(snapshot)` pour sérialiser et `json.Unmarshal` pour désérialiser le JSONB.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
