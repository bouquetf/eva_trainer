# Tâche #001 - Story #029 : Backend — Migration, modèle, repository, batch action pour l'indicateur d'étage

## Objectif
Ajouter le champ `floor` (Haut/Bas/null) à la table `zone_polygons`, mettre à jour les modèles Go, le repository, le batch handler, et la duplication de template pour supporter l'indicateur d'étage.

## Contexte
- Story #029 : `docs/stories/story-029.md`
- Dépend de : Aucune (première tâche)
- Nécessaire pour : Tâche #002 (Frontend contexte menu), Tâche #003 (Frontend indicateurs visuels), Tâche #004 (Tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Stockage :**
- Chaque `zone_polygon` peut avoir un indicateur d'étage : `NULL` (aucun), `'haut'`, ou `'bas'`
- La contrainte est gérée en base via CHECK et en Go via validation
- L'indicateur est copié lors de la duplication de template (inclure `floor` dans le JSON de `polygon_data` de `template_zones`)
- L'indicateur est inclus dans les snapshots de version (`zone_versions.version_data`)

**API :**
- Un nouvel action type `set_floor` est ajouté au batch endpoint
- L'action prend un `polygon_id` et un `floor` (string: `"haut"`, `"bas"`, ou `null` pour effacer)
- Validation : si `floor` est fourni et non-null, il doit valoir `"haut"` ou `"bas"` (case-sensitive)

**Cas nominaux :**
- Définir `floor = "haut"` sur un polygone existant → succès, valeur persistée
- Définir `floor = "bas"` sur un polygone existant → succès, valeur persistée
- Définir `floor = null` (effacer) → succès, valeur remise à NULL
- Passer de `"haut"` à `"bas"` (ou inverse) → succès, mise à jour immédiate

**Cas limites :**
- Polygone inexistant → erreur `ErrPolygonNotFound`
- Valeur de `floor` invalide (ex: `"middle"`) → erreur `ErrInvalidFloorValue`
- Rien ne change dans la sauvegarde de version si seul le floor change (la snapshot est sauvée normalement comme pour toute action batch)

**Gestion d'erreurs :**
- `polygon_id` manquant → erreur bad request
- `floor` invalide → `ErrInvalidFloorValue`
- Polygone introuvable → `ErrPolygonNotFound`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000017_add_polygon_floor.up.sql` | Créer | Migration up : ajouter `floor TEXT` à `zone_polygons` avec CHECK |
| `db/migrations/000017_add_polygon_floor.down.sql` | Créer | Migration down : DROP COLUMN `floor` |
| `internal/models/errors.go` | Modifier | Ajouter `ErrInvalidFloorValue` |
| `internal/models/zone.go` | Modifier | Ajouter `Floor` à `ZonePolygon` + méthode `SetPolygonFloor` + helper `validateFloor` |
| `internal/models/repository.go` | Modifier | Ajouter `SetPolygonFloor` à l'interface `ZoneRepository` |
| `internal/models/zone_version.go` | Modifier | Ajouter `Floor` à la structure polygon dans `VersionSnapshot` |
| `internal/models/template.go` | Modifier | Mettre à jour `collectZonePolygons` et `DuplicateToTeam` pour inclure/restaurer floor |
| `internal/handlers/zone.go` | Modifier | Ajouter `"set_floor"` au switch batch + champ `Floor` dans `batchAction` + méthode `batchSetFloor` |

### Signatures

```go
// internal/models/errors.go — AJOUTER
ErrInvalidFloorValue = errors.New("invalid floor value: must be 'haut', 'bas', or empty")

// internal/models/zone.go — MODIFIER ZonePolygon
type ZonePolygon struct {
    ID         int       `json:"id"`
    ZoneID     int       `json:"zone_id"`
    Points     []Point   `json:"points"`
    OrderIndex int       `json:"order_index"`
    Floor      *string   `json:"floor"`
    CreatedAt  time.Time `json:"created_at"`
}

// internal/models/zone.go — AJOUTER (helper)
// validateFloor vérifie que floor est nil, "haut" ou "bas"
func validateFloor(floor *string) error

// internal/models/zone.go — AJOUTER
func (r *zoneRepo) SetPolygonFloor(polygonID int, floor *string) error

// internal/models/zone_version.go — MODIFIER VersionSnapshot
type VersionSnapshot struct {
    Polygons []struct {
        Points     []Point `json:"points"`
        OrderIndex int     `json:"order_index"`
        Floor      *string `json:"floor,omitempty"`
    } `json:"polygons"`
    Name string `json:"name"`
}

// internal/handlers/zone.go — MODIFIER batchAction
type batchAction struct {
    // ... champs existants ...
    Floor *string `json:"floor,omitempty"`
}

// internal/handlers/zone.go — AJOUTER
func (h *ZoneHandler) batchSetFloor(editorID int, action batchAction) (batchActionResult, error)
```

### Contraintes techniques
- **Migration** : Le numéro de migration est `000017` (après `000016_add_user_settings`)
- **Validation** : `validateFloor` accepte `nil` (effacer), `string pointer` valant `"haut"` ou `"bas"`. Tout autre string → erreur
- **Template** : Dans `collectZonePolygons`, après avoir lu `polygon_data` et construit l'objet `{"points": points}`, ajouter `"floor": floor` si non-null. Dans `DuplicateToTeam`, lire le champ `floor` depuis le JSON et l'insérer dans la nouvelle `zone_polygons`
- **Version snapshot** : Dans `saveZoneSnapshot` (dans `zone.go`), inclure `floor` dans les données du polygone. Déjà fait automatiquement si `VersionSnapshot.Polygons` a un `Floor` field — le snapshot est construit à partir des `GetPolygons()` qui retournera maintenant le floor.
- **Lock requis** : La batch action doit vérifier le lock (déjà fait dans `BatchActions` générique)
- **GetPolygons** : Mettre à jour le scan de chaque polygone pour lire la nouvelle colonne `floor` (ajouter `floor` dans la SELECT et le Scan)

#### Détail des modifications sur GetPolygons

Dans `internal/models/zone.go`, la méthode `GetPolygons` doit être modifiée :
- Requête SQL : `SELECT id, zone_id, polygon_data, order_index, floor, created_at FROM zone_polygons ...`
- Scan : ajouter `&p.Floor` dans `rows.Scan()`
- Attention : `floor` est nullable → utiliser `sql.NullString` puis convertir en `*string`

#### Détail des modifications sur AddPolygon

La méthode `AddPolygon` doit accepter un paramètre `floor *string` optionnel :
```go
func (r *zoneRepo) AddPolygon(zoneID int, points []Point, orderIndex int, floor *string) (*ZonePolygon, error)
```
- L'INSERT doit inclure la colonne `floor`
- Mettre à jour l'interface `ZoneRepository` en conséquence
- Mettre à jour tous les appelants (notamment `batchAddPolygon` et `ImportZones`)

#### Détail duplication template

`collectZonePolygons` (template.go) : après `sql.Rows.Scan`, inclure floor dans chaque objet du tableau JSON :
```go
polyData := map[string]interface{}{
    "points": pointsData,
}
if floor.Valid {
    polyData["floor"] = floor.String
}
polygons = append(polygons, polyData)
```

`DuplicateToTeam` (template.go) : lors de l'insertion des nouveaux polygones, lire le floor depuis le JSON et l'insérer :
```go
floor, _ := poly["floor"].(string)
var floorPtr *string
if floor != "" {
    floorPtr = &floor
}
// dans l'INSERT : ajouter floor
```

### Tests

#### Tests unitaires
- **Fichier** : `internal/models/zone_test.go`
- Scénario 1 : `validateFloor(nil)` → pas d'erreur
- Scénario 2 : `validateFloor(stringPointer("haut"))` → pas d'erreur
- Scénario 3 : `validateFloor(stringPointer("bas"))` → pas d'erreur
- Scénario 4 : `validateFloor(stringPointer(""))` → pas d'erreur (considéré comme effacer)
- Scénario 5 : `validateFloor(stringPointer("middle"))` → `ErrInvalidFloorValue`
- Scénario 6 : `validateFloor(stringPointer("HAUT"))` → `ErrInvalidFloorValue` (case-sensitive)

#### Tests d'intégration
- **Fichier** : `internal/handlers/zone_test.go` (ou créer `internal/handlers/floor_test.go`)
- Scénario 1 : Batch action `set_floor` avec `floor = "haut"` → polygone mis à jour en base
- Scénario 2 : Batch action `set_floor` avec `floor = null` → colonne remise à NULL
- Scénario 3 : Batch action `set_floor` sur polygone inexistant → erreur
- Scénario 4 : Batch action `set_floor` avec `floor = "invalid"` → erreur
- Scénario 5 : Template duplication → floor préservé dans les nouveaux polygones

### Documentation
- Aucune documentation à créer dans cette tâche

### Exemples d'API

```http
POST /api/v1/teams/1/maps/1/zones/batch
Content-Type: application/json

{
  "actions": [
    {
      "type": "set_floor",
      "polygon_id": 42,
      "floor": "haut"
    },
    {
      "type": "set_floor",
      "polygon_id": 43,
      "floor": "bas"
    },
    {
      "type": "set_floor",
      "polygon_id": 44,
      "floor": null
    }
  ]
}
```
