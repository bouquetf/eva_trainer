# Tâche #001 - Story #009 : Modèles Zone, Repository et Migration (back-end)

## Objectif

Créer les modèles de données `Zone`, `ZonePolygon`, `ZoneName`, `EditLock`, `EditHistory` ainsi que leurs repositories complets, avec une migration de base de données si nécessaire.

## Contexte

- Story #009 : `docs/stories/story-009.md` - Zone Creation and Editing with Polygons
- Dépend de : Story #005 (table `maps` existante), Story #004 (rôles et permissions), Story #003 (table `teams` et `team_members`)
- Nécessaire pour : Tâche #002 (API endpoints), Tâche #003 (edit lock), Tâche #004 (frontend)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Modèles de données

Créer les 5 modèles suivants dans `internal/models/` :

1. **`Zone`** : représente une zone appartenant à une équipe sur une carte.
2. **`ZonePolygon`** : représente une forme polygonale au sein d'une zone (1+ polygones par zone).
3. **`ZoneName`** : représente le nom d'une zone (1 seul nom par zone).
4. **`EditLock`** : représente le verrou d'édition (1 verrou par équipe x carte).
5. **`EditHistory`** : représente une entrée dans l'historique des modifications.

#### Partie 2 - Repositories

Créer 3 repositories avec leurs interfaces :

1. **`ZoneRepository`** : CRUD zones + polygones + noms.
2. **`EditLockRepository`** : acquisition, libération, vérification, refresh du verrou.
3. **`EditHistoryRepository`** : écriture de l'historique, lecture par zone.

#### Partie 3 - Migration DB

Créer une migration `000011_add_zone_polygon_constraints.up.sql` si des contraintes ou index supplémentaires sont nécessaires par rapport au schéma initial.

**Cas nominaux :**
- Création d'une zone → retourne la zone avec ID
- Ajout d'un polygone à une zone existante → polygone créé avec polygon_data JSONB
- Ajout d'un nom à une zone → nom enregistré
- Récupération de toutes les zones d'une équipe sur une carte → liste complète avec polygones et noms
- Acquisition d'un verrou → lock créé avec expiration à NOW() + 15 minutes
- Libération d'un verrou → lock supprimé
- Refresh d'un verrou → expiration repoussée de 15 minutes
- Historique : ajout d'une entrée → entrée créée avec action, old_data, new_data

**Cas limites :**
- Zone sans polygone → liste de polygones vide (pas nil)
- Zone sans nom → name = ""
- Création d'un doublon de verrou (team_id, map_id) → erreur `ErrLockAlreadyAcquired`
- Vérification d'un verrou inexistant → retourne nil (pas d'erreur)
- Verrou expiré → considéré comme libre
- Aucune zone en base → liste vide
- Nom de zone vide → erreur `ErrZoneNameRequired`
- Nom de zone > 250 caractères → erreur `ErrZoneNameTooLong`

**Gestion d'erreurs :**
- Erreur DB → propagation avec contexte (fmt.Errorf avec %w)
- Zone introuvable → `ErrZoneNotFound`
- Polygone introuvable → `ErrPolygonNotFound`
- Verrou inexistant → `ErrLockNotFound`
- Verrou déjà acquis → `ErrLockAlreadyAcquired`
- Nom invalide → `ErrZoneNameRequired` / `ErrZoneNameTooLong`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/zone.go` | Créer | Structs Zone, ZonePolygon, ZoneName + ZoneRepository interface + implémentation |
| `internal/models/edit_lock.go` | Créer | Struct EditLock + EditLockRepository interface + implémentation |
| `internal/models/edit_history.go` | Créer | Struct EditHistory + EditHistoryRepository |
| `internal/models/repository.go` | Modifier | Ajouter les interfaces ZoneRepository, EditLockRepository, EditHistoryRepository |
| `internal/models/errors.go` | Modifier | Ajouter les erreurs liées aux zones, polygones, locks |
| `db/migrations/000011_add_zone_polygon_constraints.up.sql` | Créer | Migration complémentaire si besoin |
| `db/migrations/000011_add_zone_polygon_constraints.down.sql` | Créer | Rollback |

### Signatures

```go
// internal/models/zone.go
package models

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Zone struct {
	ID        int       `json:"id"`
	TeamID    int       `json:"team_id"`
	MapID     int       `json:"map_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ZonePolygon struct {
	ID         int       `json:"id"`
	ZoneID     int       `json:"zone_id"`
	Points     []Point   `json:"points"`       // stocké en JSONB
	OrderIndex int       `json:"order_index"`
	CreatedAt  time.Time `json:"created_at"`
}

type ZoneName struct {
	ID        int       `json:"id"`
	ZoneID    int       `json:"zone_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ZoneWithDetails agrège une zone avec ses polygones et son nom
type ZoneWithDetails struct {
	ID        int            `json:"id"`
	TeamID    int            `json:"team_id"`
	MapID     int            `json:"map_id"`
	Polygons  []ZonePolygon  `json:"polygons"`
	Name      string         `json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type ZoneRepository interface {
	// Zones
	Create(teamID, mapID int) (*Zone, error)
	FindByID(zoneID int) (*Zone, error)
	FindByTeamAndMap(teamID, mapID int) ([]*Zone, error)
	FindWithDetailsByTeamAndMap(teamID, mapID int) ([]*ZoneWithDetails, error)
	Delete(zoneID int) error

	// Polygones
	AddPolygon(zoneID int, points []Point, orderIndex int) (*ZonePolygon, error)
	GetPolygons(zoneID int) ([]*ZonePolygon, error)
	UpdatePolygonPoints(polygonID int, points []Point) error
	RemovePolygon(polygonID int) error
	PolygonCount(zoneID int) (int, error)

	// Noms
	SetName(zoneID int, name string) error
	GetName(zoneID int) (*ZoneName, error)
	IsNameTaken(teamID, mapID int, name string) (bool, error)
}
```

```go
// internal/models/edit_lock.go
package models

type EditLock struct {
	ID        int       `json:"id"`
	TeamID    int       `json:"team_id"`
	MapID     int       `json:"map_id"`
	EditorID  int       `json:"editor_id"`
	LockedAt  time.Time `json:"locked_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type EditLockRepository interface {
	Acquire(teamID, mapID, editorID int) (*EditLock, error)
	Release(teamID, mapID int) error
	ReleaseByEditor(teamID, mapID, editorID int) error
	Get(teamID, mapID int) (*EditLock, error)
	Refresh(teamID, mapID int) (*EditLock, error)
	IsLocked(teamID, mapID int) (bool, error)
	IsLockedByEditor(teamID, mapID, editorID int) (bool, error)
	CleanupExpired() (int64, error) // nettoie les locks expirés, retourne le nombre
}
```

```go
// internal/models/edit_history.go
package models

type EditHistory struct {
	ID        int             `json:"id"`
	ZoneID    int             `json:"zone_id"`
	EditorID  *int            `json:"editor_id,omitempty"`
	Action    string          `json:"action"`
	OldData   json.RawMessage `json:"old_data,omitempty"`
	NewData   json.RawMessage `json:"new_data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

type EditHistoryRepository interface {
	Log(zoneID int, editorID *int, action string, oldData, newData json.RawMessage) (*EditHistory, error)
	FindByZone(zoneID int) ([]*EditHistory, error)
	FindByTeamAndMap(teamID, mapID int) ([]*EditHistory, error)
}
```

```go
// Ajouts dans internal/models/errors.go
var (
	ErrZoneNotFound          = errors.New("zone not found")
	ErrZoneNameRequired      = errors.New("zone name is required")
	ErrZoneNameTooLong       = errors.New("zone name exceeds maximum length of 250 characters")
	ErrZoneNameTaken         = errors.New("zone name already exists for this team on this map")
	ErrPolygonNotFound       = errors.New("polygon not found")
	ErrPolygonHasTooFewPoints = errors.New("polygon must have at least 3 points")
	ErrLockNotFound          = errors.New("edit lock not found")
	ErrLockAlreadyAcquired   = errors.New("edit lock already acquired by another editor")
	ErrLockNotOwned          = errors.New("edit lock is owned by another editor")
)
```

### Contraintes techniques

- **Package naming** : `models`. Import path : `eva_trainer/internal/models`.
- **Connexion DB** : Prendre `*sql.DB` dans le constructeur. La connexion est disponible via `database.DB()` dans `pkg/database`.
- **Polymorphisme JSONB** : Utiliser `encoding/json` pour marshal/unmarshal des `[]Point` dans le champ `polygon_data`. Créer une helper function `scanPolygonData(data []byte) ([]Point, error)`.
- **Nullable EditorID** : `*int` pour `editor_id` dans EditHistory (ON DELETE SET NULL en DB).
- **Point** : Struct avec `X` et `Y` en `float64`. Compatible avec le format attendu par le frontend.
- **Pattern repository** : Suivre exactement le pattern des repositories existants (TeamRepository, UserRepository, etc.) : interface publique, struct privée, constructeur public `New*Repository`.
- **Transactions** : La méthode `Create` de ZoneRepository doit être simple (INSERT seule). La gestion de l'atomicité zone+polygones+nom sera faite au niveau handler.
- **EditLock.Acquire** : Doit être atomique (INSERT avec vérification UNIQUE ou UPDATE). Utiliser `INSERT ... ON CONFLICT DO NOTHING` ou vérifier d'abord si un verrou existe et n'est pas expiré.
- **EditLock.Refresh** : Met à jour `expires_at = NOW() + INTERVAL '15 minutes'`.
- **EditLock.CleanupExpired** : `DELETE FROM edit_locks WHERE expires_at < NOW()`.
- **ZoneName contrainte** : Vérifier avec `IsNameTaken` avant d'assigner un nom. Unique par (team_id, map_id) — utiliser une query dédiée.
- **Ordre des polygones** : `order_index` pour préserver l'ordre d'affichage des polygones.
- **ZoneWithDetails** : Doit charger une zone avec tous ses polygones et son nom en 3 requêtes max (une pour la zone, une pour les polygones, une pour le nom) ou via une jointure plus complexe.

### Tests à implémenter

Les tests seront créés dans la Tâche #006.

### Documentation

Aucune documentation à créer pour cette tâche.
