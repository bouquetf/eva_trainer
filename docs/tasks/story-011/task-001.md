# Tâche #001 - Story #011 : Modèle Template, Repository et API de publication

## Objectif

Créer les modèles de données `Template` et `TemplateZone`, leur repository, et l'endpoint API de publication d'une carte (all-or-nothing par carte) pour les Capitaine/Cocapitaine.

## Contexte

- Story #011 : `docs/stories/story-011.md` - Zone Publishing and Template System
- Dépend de : Story #009 (modèle `Zone`, `ZonePolygon`, `ZoneName` + repositories existants), Story #004 (rôles et permissions, rôle Captain/Cocaptain)
- Nécessaire pour : Tâche #002 (auto-update template), Tâche #003 (frontend publication), Tâche #004 (tests)
- Ordre d'exécution : 1ère
- Les tables `templates` et `template_zones` existent déjà dans la migration `000001_initial_schema.up.sql`

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Modèles de données

Créer le fichier `internal/models/template.go` avec les structures suivantes :

```go
type Template struct {
    ID             int       `json:"id"`
    OriginalTeamID int       `json:"original_team_id"`
    MapID          int       `json:"map_id"`
    IsPublic       bool      `json:"is_public"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}

type TemplateZone struct {
    ID           int       `json:"id"`
    TemplateID   int       `json:"template_id"`
    PolygonData  json.RawMessage `json:"polygon_data"`  // JSONB stockant les coordonnées
    Name         string    `json:"name"`
    OrderIndex   int       `json:"order_index"`
}

type TemplateWithZones struct {
    Template Template      `json:"template"`
    Zones    []TemplateZone `json:"zones"`
}
```

Ajouter dans `internal/models/repository.go` l'interface `TemplateRepository` :

```go
type TemplateRepository interface {
    CreateFromMap(teamID, mapID int) (*Template, error)
    FindByMapID(mapID int) (*Template, error)
    FindAllPublic() ([]*TemplateWithZones, error)
    UpdateFromMap(templateID, teamID, mapID int) error
    Delete(templateID int) error
}
```

Implémenter le repository dans `internal/models/template.go` :

```go
type templateRepo struct {
    db *sql.DB
}

func NewTemplateRepository(db *sql.DB) TemplateRepository {
    return &templateRepo{db: db}
}
```

**Méthode `CreateFromMap`** :
1. Vérifier qu'aucun template n'existe déjà pour cette `map_id`
2. Récupérer toutes les zones de l'équipe pour cette carte via `zoneRepo.FindWithDetailsByTeamAndMap(teamID, mapID)`
3. Insérer dans `templates` : `INSERT INTO templates (original_team_id, map_id, is_public, created_at, updated_at) VALUES ($1, $2, true, NOW(), NOW()) RETURNING id, ...`
4. Pour chaque zone, insérer dans `template_zones` : `INSERT INTO template_zones (template_id, polygon_data, name, order_index) VALUES ($1, $2::jsonb, $3, $4)`
5. Retourner le Template créé

**Méthode `FindByMapID`** :
- `SELECT * FROM templates WHERE map_id = $1 AND is_public = true`

**Méthode `FindAllPublic`** :
- `SELECT t.*, tz.id, tz.polygon_data, tz.name, tz.order_index FROM templates t LEFT JOIN template_zones tz ON tz.template_id = t.id WHERE t.is_public = true ORDER BY t.created_at DESC`

**Méthode `UpdateFromMap`** :
- Utilisée par la Tâche #002 pour la mise à jour automatique
- Supprimer toutes les `template_zones` existantes pour ce template
- Réinsérer les zones actuelles (même logique que CreateFromMap)
- Mettre à jour `updated_at`

**Méthode `Delete`** :
- `DELETE FROM templates WHERE id = $1` (CASCADE supprime les template_zones)

#### Partie 2 - Handler de publication

Créer `internal/handlers/template.go` :

```go
type TemplateHandler struct {
    templateRepo   models.TemplateRepository
    zoneRepo       models.ZoneRepository
    teamMemberRepo models.TeamMemberRepository
}

func NewTemplateHandler(
    templateRepo models.TemplateRepository,
    zoneRepo models.ZoneRepository,
    teamMemberRepo models.TeamMemberRepository,
) *TemplateHandler
```

**Méthode `PublishMap` (POST /api/v1/teams/:id/maps/:mapId/publish) :**

1. Extraire `teamID` et `mapID` des paramètres
2. Récupérer l'utilisateur connecté via `middleware.MustGetUser(c)`
3. Vérifier que l'utilisateur est Capitaine ou Cocapitaine de l'équipe (via `teamMemberRepo`)
4. Vérifier que la carte n'est pas déjà publiée (via `templateRepo.FindByMapID`) → si oui, retourner 409 Conflict
5. Appeler `templateRepo.CreateFromMap(teamID, mapID)`
6. Retourner 201 Created avec le template créé

**Cas nominaux :**
- Capitaine publie une carte → 201 + template complet (`{ "template": { "id": 1, "map_id": 1, "is_public": true, ... } }`)
- La carte n'a pas de zones → 201 + template avec 0 zones (cas valide)

**Gestion d'erreurs :**
- Capitaine/Cocapitaine manquant → 403 Forbidden
- Template existe déjà pour cette carte → 409 Conflict, `error: "already_published"`
- Utilisateur non membre de l'équipe → 403 Forbidden
- Erreur DB → 500 Internal Server Error

**Méthode `GetPublishStatus` (GET /api/v1/teams/:id/maps/:mapId/publish-status) :**

1. Extraire `teamID` et `mapID`
2. Appeler `templateRepo.FindByMapID(mapID)`
3. Retourner `{ "published": true/false, "template": {...} }` ou `{ "published": false }`

**Méthode `AdminDeleteTemplate` (DELETE /api/v1/admin/templates/:templateId) :**

1. Extraire `templateID` des paramètres
2. Récupérer l'utilisateur connecté — le middleware `RequireAdmin` garantit que seul un admin peut appeler
3. Appeler `templateRepo.Delete(templateID)`
4. Retourner 200 OK `{ "message": "template deleted" }`

**Cas nominaux :**
- Admin supprime un template → 200 OK
- Template inexistant → 404 Not Found (`templateRepo.Delete` ne fait rien ou on ajoute une vérification préalable)

**Gestion d'erreurs :**
- Non-admin → 403 Forbidden (géré par middleware)
- Erreur DB → 500 Internal Server Error

**Réponse :**
```json
// Publiée
{
  "published": true,
  "template": {
    "id": 1,
    "map_id": 1,
    "original_team_id": 1,
    "is_public": true,
    "created_at": "2026-05-10T12:00:00Z",
    "updated_at": "2026-05-10T12:00:00Z"
  }
}

// Non publiée
{
  "published": false
}
```

#### Partie 3 - Routes API

Ajouter dans `api/routes.go` (dans le groupe team routes) :

```go
// Dans setupTeamRoutes, après les routes zones :
templateRepo := models.NewTemplateRepository(db)
templateHandler := handlers.NewTemplateHandler(templateRepo, zoneRepo, teamMemberRepo)

// Routes de publication (Captain/Cocaptain only)
teams.POST("/:id/maps/:mapId/publish",
    middleware.RequireAuth(userRepo, jwtService),
    middleware.RequireTeamRole(teamMemberRepo, "cocaptain", "captain"),
    templateHandler.PublishMap,
)

teams.GET("/:id/maps/:mapId/publish-status",
    middleware.RequireAuth(userRepo, jwtService),
    middleware.RequireTeamRole(teamMemberRepo, "editor", "cocaptain", "captain"),
    templateHandler.GetPublishStatus,
)

// Dans le groupe admin (setupAdminRoutes), ajouter :
admin := r.Group("/api/v1/admin", middleware.RequireAuth(userRepo, jwtService), middleware.RequireAdmin())
// ...
admin.DELETE("/templates/:templateId", templateHandler.AdminDeleteTemplate)
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/template.go` | Créer | Modèles `Template`, `TemplateZone`, `TemplateWithZones` + implémentation `TemplateRepository` |
| `internal/models/repository.go` | Modifier | Ajouter l'interface `TemplateRepository` |
| `internal/handlers/template.go` | Créer | Handler `TemplateHandler` avec `PublishMap`, `GetPublishStatus`, `AdminDeleteTemplate` |
| `api/routes.go` | Modifier | Ajouter les routes publication + admin DELETE /templates/:templateId |

### Signatures

```go
// Modèles
type Template struct {
    ID             int       `json:"id"`
    OriginalTeamID int       `json:"original_team_id"`
    MapID          int       `json:"map_id"`
    IsPublic       bool      `json:"is_public"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}

type TemplateZone struct {
    ID          int             `json:"id"`
    TemplateID  int             `json:"template_id"`
    PolygonData json.RawMessage `json:"polygon_data"`
    Name        string          `json:"name"`
    OrderIndex  int             `json:"order_index"`
}

type TemplateWithZones struct {
    Template Template       `json:"template"`
    Zones    []TemplateZone `json:"zones"`
}

// Repository interface (dans repository.go)
type TemplateRepository interface {
    CreateFromMap(teamID, mapID int) (*Template, error)
    FindByMapID(mapID int) (*Template, error)
    FindAllPublic() ([]*TemplateWithZones, error)
    UpdateFromMap(templateID, teamID, mapID int) error
    Delete(templateID int) error
}

// Handler
type TemplateHandler struct {
    templateRepo   models.TemplateRepository
    zoneRepo       models.ZoneRepository
    teamMemberRepo models.TeamMemberRepository
}

func NewTemplateHandler(templateRepo, zoneRepo, teamMemberRepo) *TemplateHandler
func (h *TemplateHandler) PublishMap(c *gin.Context)
func (h *TemplateHandler) GetPublishStatus(c *gin.Context)
func (h *TemplateHandler) AdminDeleteTemplate(c *gin.Context)
```

### Contraintes techniques

- **Contrainte définitive** : Une fois `is_public = true`, on ne permet pas de repasser à false via l'API publique. Seul un admin peut supprimer un template.
- **Snapshot au moment de la publication** : `CreateFromMap` doit copier les polygones (`polygon_data`) et les noms (`name`) de TOUTES les zones de l'équipe pour cette carte. Ne pas copier les `zone_id` ou autres références internes — le template est une copie indépendante.
- **Format `polygon_data`** : Stocker le JSONB exact des polygones. Pour chaque zone, convertir les `[]ZonePolygon` en un tableau JSON `[{ "points": [...] }, ...]` avec les points existants.
- **Pattern handler** : Suivre le pattern existant des autres handlers (injection de repositories dans le constructeur).
- **Permissions** : Vérifier via `middleware.RequireTeamRole(teamMemberRepo, "cocaptain", "captain")`. Seul le Capitaine ou Cocapitaine peut publier.
- **Tests** : Voir Tâche #004 pour les tests d'intégration.

### Tests à implémenter

Les tests seront créés dans la Tâche #004.

### Documentation

Aucune documentation à créer pour cette tâche. La documentation utilisateur sera créée dans la Tâche #004.
