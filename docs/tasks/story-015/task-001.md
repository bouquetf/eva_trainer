# Tâche #001 - Story #015 : Backend - API endpoint de récupération des zones de révision

## Objectif

Créer l'endpoint API backend permettant de récupérer toutes les zones disponibles pour une session de révision : les zones de l'équipe + les zones des templates publics pour la même carte.

## Contexte

- Story #015 : `docs/stories/story-015.md` - Revision Mode with Validation
- Dépend de : Story #009 (modèle `Zone`, `ZonePolygon`, `ZoneName`), Story #011 (modèle `Template`, `TemplateZone`)
- Nécessaire pour : Tâche #002 (API client frontend), Tâche #003 (store), Tâche #004 (page)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Nouvelle méthode sur TemplateRepository

Ajouter une méthode `FindTemplateZonesByMapID` dans l'interface `TemplateRepository` (dans `internal/models/repository.go`) et son implémentation (dans `internal/models/template.go`).

**Comportement :**
- Reçoit un `mapID`
- Trouve le template public associé à cette carte via `SELECT * FROM templates WHERE map_id = $1 AND is_public = true`
- Si un template existe, récupère toutes ses `template_zones`
- Chaque `TemplateZone` contient `polygon_data` (JSONB) et `name`
- Retourne une liste de `TemplateZone`

**Signature :**

```go
// Dans repository.go, ajouter à l'interface TemplateRepository :
FindTemplateZonesByMapID(mapID int) ([]TemplateZone, error)
```

**Note :** Un template est lié à une map_id. Si la carte n'a pas de template public, retourner une slice vide (pas d'erreur).

#### Partie 2 - Revision zones endpoint

Créer un nouveau handler `RevisionHandler` dans `internal/handlers/revision.go` avec :

```go
type RevisionHandler struct {
    zoneRepo     models.ZoneRepository
    templateRepo models.TemplateRepository
    mapRepo      models.MapRepository
    logger       services.Logger
}

func NewRevisionHandler(
    zoneRepo models.ZoneRepository,
    templateRepo models.TemplateRepository,
    mapRepo models.MapRepository,
    logger services.Logger,
) *RevisionHandler
```

**Structure de réponse :**
```go
type RevisionZone struct {
    ID       int              `json:"id"`        // Zone ID (pour les zones équipe) ou TemplateZone ID (pour les zones template)
    Source   string           `json:"source"`    // "team" ou "template"
    MapID    int              `json:"map_id"`     // map_id de la zone
    Polygons []models.ZonePolygon  `json:"polygons"`  // polygones à afficher sur la carte
    Name     string           `json:"name"`      // nom attendu (réponse correcte)
}
```

**Méthode `ListRevisionZones` (GET /api/v1/teams/:id/maps/:mapId/revision-zones) :**

1. Extraire `teamID` et `mapID` des paramètres
2. Récupérer l'utilisateur connecté via `middleware.MustGetUser(c)`
3. Récupérer les zones de l'équipe pour cette carte via `zoneRepo.FindWithDetailsByTeamAndMap(teamID, mapID)`
4. Récupérer les zones du template public pour cette carte via `templateRepo.FindTemplateZonesByMapID(mapID)`
5. Pour chaque zone template, construire un `RevisionZone` :
   - Convertir `polygon_data` (JSON array) en `[]ZonePolygon` : chaque élément du JSON array devient un `ZonePolygon` avec l'ordre comme `OrderIndex` et les points parsés depuis le JSON
   - `ID` = ID de la template_zone, `Source` = "template", `Name` = name, `MapID` = mapID
6. Pour chaque zone équipe, construire un `RevisionZone` :
   - `ID` = zone.ID, `Source` = "team", `Polygons` = zone.Polygons, `Name` = zone.Name, `MapID` = zone.MapID
7. Retourner `{"zones": [...]}`

**Gestion d'erreurs :**
- Utilisateur non membre de l'équipe → 403 Forbidden (géré par middleware)
- Carte introuvable → 404 Not Found
- Erreur DB → 500 Internal Server Error

#### Partie 3 - Routes

Dans `api/routes.go`, ajouter dans `setupTeamRoutes` :

```go
revisionHandler := handlers.NewRevisionHandler(zoneRepo, templateRepo, mapRepo, logger)

teams.GET("/:id/maps/:mapId/revision-zones",
    middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
    revisionHandler.ListRevisionZones,
)
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/revision.go` | Créer | Nouveau handler `RevisionHandler` avec `ListRevisionZones` |
| `internal/models/repository.go` | Modifier | Ajouter `FindTemplateZonesByMapID` à `TemplateRepository` |
| `internal/models/template.go` | Modifier | Implémenter `FindTemplateZonesByMapID` |
| `api/routes.go` | Modifier | Ajouter la route revision-zones |

### Signatures

```go
// TemplateRepository (repository.go)
type TemplateRepository interface {
    // ... existing methods ...
    FindTemplateZonesByMapID(mapID int) ([]TemplateZone, error)
}

// Nouvelle méthode dans template.go
func (r *templateRepo) FindTemplateZonesByMapID(mapID int) ([]TemplateZone, error) {
    // 1. Find template by mapID
    // 2. Query template_zones
    // 3. Return zones
}

// RevisionHandler (revision.go)
type RevisionHandler struct {
    zoneRepo     models.ZoneRepository
    templateRepo models.TemplateRepository
    mapRepo      models.MapRepository
    logger       services.Logger
}

func NewRevisionHandler(zoneRepo, templateRepo, mapRepo, logger) *RevisionHandler

func (h *RevisionHandler) ListRevisionZones(c *gin.Context)

// RevisionZone (revision.go)
type RevisionZone struct {
    ID       int                    `json:"id"`
    Source   string                 `json:"source"`
    MapID    int                    `json:"map_id"`
    Polygons []models.ZonePolygon   `json:"polygons"`
    Name     string                 `json:"name"`
}
```

### Contraintes techniques

- **Pattern handler** : Suivre exactement le pattern des handlers existants (constructeur public avec injection de repositories)
- **Polygon data conversion** : `TemplateZone.PolygonData` est un `json.RawMessage` contenant un JSON array d'objets avec des `points`. Chaque élément correspond à un polygone. Convertir en `[]ZonePolygon` complet (avec `ZoneID` = -1 pour les templates, `ID` négatif unique basé sur l'index, `OrderIndex`, `Points` parsés). Le format doit être identique à `ZonePolygon` pour que le frontend puisse afficher les polygones de manière uniforme
- **Permissions** : Seul `RequireTeamRole` avec les rôles "member", "editor", "cocaptain", "captain" est nécessaire. Pas de `RequirePermission` supplémentaire — l'appartenance à l'équipe donne accès aux zones privées et aux templates publics
- **Cartes obsolètes** : Ne pas inclure les zones des cartes obsolètes pour l'instant. Seules les zones de la carte sélectionnée (mapID) sont retournées
- **Slice vide** : Retourner `[]RevisionZone{}` si aucune zone n'est trouvée, pas `null`
- **Logger** : Utiliser `services.Logger` pour le logging comme dans les autres handlers

### Tests à implémenter

#### Tests d'intégration (dans `internal/handlers/revision_test.go`)

- **Fichier** : `internal/handlers/revision_test.go`
- Build tag : `//go:build integration`

**Scénario 1 : Récupération des zones équipe uniquement**
- Créer une équipe, une carte, des zones pour cette équipe
- Aucun template public
- Appeler `GET /api/v1/teams/:id/maps/:mapId/revision-zones`
- Vérifier que la réponse contient uniquement les zones de l'équipe avec `source: "team"`
- Vérifier les polygones et noms

**Scénario 2 : Récupération des zones équipe + template**
- Créer une équipe A, une carte, des zones pour l'équipe A
- Publier la carte (template public)
- Créer une équipe B (autre utilisateur)
- Appeler l'endpoint avec teamID = équipe B, mapID = la carte
- Vérifier que les zones du template public sont incluses avec `source: "template"`
- Vérifier que les zones sont dédoublonnées si l'équipe B a aussi ses propres zones

**Scénario 3 : Aucune zone**
- Aucune zone pour l'équipe, aucun template public
- Vérifier `{"zones": []}`

**Scénario 4 : 401 non authentifié**
- Appeler sans token JWT → 401

#### Tests unitaires (dans `internal/models/template_test.go` si existant)

- **Fichier** : `internal/models/template_test.go`
- Tester `FindTemplateZonesByMapID` avec un template existant et un template absent

### Documentation

Aucune documentation à créer pour cette tâche. La documentation sera créée dans la Tâche #005.
