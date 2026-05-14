# Tâche #001 - Story #021 : Export des zones — Backend (Go)

## Objectif
Implémenter l'API backend d'export des zones : endpoints administrateur pour exporter une carte ou toutes les cartes d'une équipe au format JSON.

## Contexte
- Story #021 : [Import/Export de zones (Admin)](../../stories/story-021-done.md)
- Dépend de : *aucune* (Story #009 et #005 déjà implémentées)
- Nécessaire pour : Tâche #002 (Import API), Tâche #003 (Frontend Admin UI), Tâche #004 (Documentation)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter deux endpoints administrateur pour exporter les zones d'une équipe au format JSON.

**Format JSON d'export (points en format structuré `{x, y}` — cohérent avec le modèle existant) :**

```json
// Export d'une seule carte
{
  "map_name": "Atlantis",
  "team": "Mon équipe",
  "export_date": "2026-05-14T22:00:00Z",
  "zones": [
    {
      "name": "top",
      "polygons": [
        {
          "points": [
            { "x": 0, "y": 0 },
            { "x": 100, "y": 0 },
            { "x": 100, "y": 100 },
            { "x": 0, "y": 100 }
          ]
        }
      ]
    }
  ]
}

// Export de toutes les cartes
[
  {
    "map_name": "Atlantis",
    "team": "Mon équipe",
    "export_date": "2026-05-14T22:00:00Z",
    "zones": [...]
  },
  {
    "map_name": "Lunar",
    "team": "Mon équipe",
    "export_date": "2026-05-14T22:00:00Z",
    "zones": [...]
  }
]
```

**Cas nominaux :**
- Export d'une carte spécifique d'une équipe → retourne le JSON formaté avec les zones et polygones
- Export de toutes les cartes d'une équipe → retourne un tableau JSON des cartes avec leurs zones
- Les zones sont exportées avec leurs polygones (liste de points {x, y}) et leur nom
- Le JSON contient : `map_name` (nom eva de la carte), `team` (nom de l'équipe), `export_date` (timestamp ISO8601), `zones` (tableau)

**Cas limites :**
- Une carte sans zone → `zones` est un tableau vide `[]`
- Une équipe sans aucune carte → tableau vide `[]` pour l'export multi-cartes
- Carte introuvable → 404
- Équipe introuvable → 404
- L'export ne doit pas inclure le `webp_blob` des cartes

**Gestion d'erreurs :**
- `invalid_team_id` → 400
- `team_not_found` → 404
- `invalid_map_id` → 400
- `map_not_found` → 404
- `internal_error` → 500

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/repository.go` | Modifier | Ajouter `FindWithDetailsByTeam` à `ZoneRepository` interface |
| `internal/models/zone.go` | Modifier | Ajouter l'implémentation de `FindWithDetailsByTeam` |
| `internal/handlers/import_export.go` | Créer | Nouveau handler `ImportExportHandler` avec méthodes d'export |
| `internal/models/team.go` | Aucune modification nécessaire | Utiliser `FindByID` existant |
| `internal/models/map.go` | Aucune modification nécessaire | Utiliser `FindByID` et `FindByNameLatest` existants |
| `api/routes.go` | Modifier | Ajouter les routes d'export dans `setupAdminRoutes` |

### Signatures

```go
// internal/models/repository.go — à ajouter à ZoneRepository
FindWithDetailsByTeam(teamID int) (map[int][]*ZoneWithDetails, error)

// La clé de la map est le mapID, la valeur est la liste des zones avec détails
```

```go
// internal/models/zone.go — nouvelle méthode sur zoneRepo
func (r *zoneRepo) FindWithDetailsByTeam(teamID int) (map[int][]*ZoneWithDetails, error) {
    // 1. Récupérer tous les maps_id distincts pour cette équipe
    // 2. Pour chaque map, appeler FindWithDetailsByTeamAndMap (existant)
    // 3. Retourner une map[mapID][]*ZoneWithDetails
}
```

```go
// internal/handlers/import_export.go
import (
    "net/http"
    "strconv"
    "time"

    "eva_trainer/internal/models"
    "eva_trainer/internal/services"
    "eva_trainer/pkg/apierror"
    "github.com/gin-gonic/gin"
)

type ImportExportHandler struct {
    zoneRepo models.ZoneRepository
    mapRepo  models.MapRepository
    teamRepo models.TeamRepository
    logger   services.Logger
}

func NewImportExportHandler(
    zoneRepo models.ZoneRepository,
    mapRepo models.MapRepository,
    teamRepo models.TeamRepository,
    logger services.Logger,
) *ImportExportHandler {
    return &ImportExportHandler{
        zoneRepo: zoneRepo,
        mapRepo:  mapRepo,
        teamRepo: teamRepo,
        logger:   logger,
    }
}

// Structures pour le format d'export
type ZoneExport struct {
    Name     string          `json:"name"`
    Polygons []PolygonExport `json:"polygons"`
}

type PolygonExport struct {
    Points []models.Point `json:"points"`
}

type MapExport struct {
    MapName    string       `json:"map_name"`
    Team       string       `json:"team"`
    ExportDate string       `json:"export_date"`
    Zones      []ZoneExport `json:"zones"`
}

// ExportTeamMap exporte les zones d'une équipe pour une carte spécifique
// GET /api/v1/admin/teams/:id/export/:mapId
func (h *ImportExportHandler) ExportTeamMap(c *gin.Context)

// ExportTeamMaps exporte les zones de toutes les cartes d'une équipe
// GET /api/v1/admin/teams/:id/export
func (h *ImportExportHandler) ExportTeamMaps(c *gin.Context)
```

```go
// api/routes.go — dans setupAdminRoutes, après la création du handler
zoneRepoAdmin := models.NewZoneRepository(db)
mapRepoAdmin := models.NewMapRepository(db)
teamRepoAdmin := models.NewTeamRepository(db)
importExportHandler := handlers.NewImportExportHandler(zoneRepoAdmin, mapRepoAdmin, teamRepoAdmin, logger)

admin.GET("/teams/:id/export", importExportHandler.ExportTeamMaps)
admin.GET("/teams/:id/export/:mapId", importExportHandler.ExportTeamMap)
```

### Contraintes techniques
- **Framework** : Go avec Gin web framework
- **Pattern handler** : suivre le pattern exact de `AdminHandler` (struct + constructeur + méthodes)
- **Pattern route** : suivre le pattern exact de `setupAdminRoutes` — création des repos, handler, puis enregistrement des routes sur le group `admin`
- **Accès admin uniquement** : les routes sont sous le group `admin` qui utilise déjà `RequireAuth` + `RequireAdmin`
- **Performance** : pas de limite spécifique, les exports concernent des données de taille modérée
- **Timesamp** : utiliser `time.Now().UTC().Format(time.RFC3339)` pour `export_date`
- **Nom de carte** : utiliser `map.Name` (nom eva) et non l'ID
- **Nom d'équipe** : utiliser `team.Name`

### Tests à implémenter

#### Tests unitaires — `internal/handlers/export_test.go`
- **Scénario 1** : Export d'une carte avec zones → retourne le JSON formaté correctement
  - Mock zoneRepo.FindWithDetailsByTeamAndMap() retourne 2 zones avec polygones
  - Mock mapRepo.FindByID() retourne la carte
  - Mock teamRepo.FindByID() retourne l'équipe
  - Vérifier : statut 200, body JSON valide avec map_name, team, export_date, zones
- **Scénario 2** : Export d'une carte sans zone → retourne zones: []
  - Mock zoneRepo.FindWithDetailsByTeamAndMap() retourne [] (vide)
  - Vérifier : statut 200, `"zones": []`
- **Scénario 3** : Export de toutes les cartes d'une équipe → retourne un tableau
  - Mock zoneRepo.FindWithDetailsByTeam() retourne 2 maps avec leurs zones
  - Vérifier : statut 200, body est un tableau de MapExport
- **Scénario 4** : Export avec team_id invalide → 400
- **Scénario 5** : Export avec map_id invalide → 400
- **Scénario 6** : Équipe inexistante → 404

#### Tests d'intégration — `internal/handlers/admin_test.go` ou nouveau fichier
- Créer une équipe, ajouter des zones sur une carte, exporter → vérifier le format JSON
- Exporter toutes les cartes → vérifier que toutes les cartes sont présentes

### Documentation

Ne pas créer de documentation ici. Voir Tâche #004.

### Exemples d'utilisation

```bash
# Export d'une carte spécifique
curl -H "Authorization: Bearer <admin-token>" \
  http://localhost:8080/api/v1/admin/teams/1/export/5

# Export de toutes les cartes d'une équipe
curl -H "Authorization: Bearer <admin-token>" \
  http://localhost:8080/api/v1/admin/teams/1/export
```
