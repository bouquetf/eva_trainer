# Tâche #001 - Story #014 : API de duplication de template public

## Objectif

Ajouter la méthode `DuplicateToTeam` au `TemplateRepository`, les handlers `ListPublicTemplates` et `DuplicateTemplate`, les routes API associées, et ajouter `MapName` au modèle `Template`.

## Contexte

- Story #014 : `docs/stories/story-014.md` - Template Duplication by Other Teams
- Dépend de : Story #011 (modèle `Template`, `TemplateZone`, `TemplateRepository`, handler `TemplateHandler`)
- Nécessaire pour : Tâche #002 (frontend), Tâche #003 (tests)
- Ordre d'exécution : 1ère
- **Décisions utilisateur** :
  - Ajouter `MapName` au modèle `Template` via JOIN (Q1-a)
  - Déplacer la création de `TemplateHandler` dans `SetupRoutes`, passer à `setupTeamRoutes` (Q2-b)
  - `DuplicateToTeam` retourne `[]*Zone`, handler appelle `FindWithDetailsByTeamAndMap` (Q3-a)
  - Vérification explicite `is_public` dans `DuplicateToTeam` (Q4-a)
  - Réservé aux capitaines/cocapitaines, l'import **remplace** toutes les zones existantes (Q5)
  - Les conflits de noms sont résolus par le remplacement (Q6)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 0 - Ajout de `MapName` au modèle `Template`

Modifier `internal/models/template.go` :

```go
type Template struct {
    ID             int       `json:"id"`
    OriginalTeamID int       `json:"original_team_id"`
    MapID          int       `json:"map_id"`
    MapName        string    `json:"map_name"`
    IsPublic       bool      `json:"is_public"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

Modifier `scanTemplate` et `FindAllPublic` pour JOIN avec `maps` et récupérer `maps.name` :

```go
func scanTemplate(scanner interface {
    Scan(dest ...interface{}) error
}) (*Template, error) {
    t := &Template{}
    err := scanner.Scan(&t.ID, &t.OriginalTeamID, &t.MapID, &t.MapName, &t.CreatedAt, &t.UpdatedAt)
    // ...
}
```

Mettre à jour `CreateFromMap` et `FindByMapID` pour inclure le JOIN et récupérer `maps.name AS map_name`.

`FindAllPublic` doit JOIN avec `maps` :
```sql
SELECT t.id, t.original_team_id, t.map_id, m.name AS map_name, t.created_at, t.updated_at,
       tz.id, tz.polygon_data, tz.name, tz.order_index
FROM templates t
LEFT JOIN maps m ON m.id = t.map_id
LEFT JOIN template_zones tz ON tz.template_id = t.id
ORDER BY t.created_at DESC
```

#### Partie 1 - Nouvelle méthode `DuplicateToTeam` dans `TemplateRepository`

Ajouter dans `internal/models/template.go` la méthode :

```go
func (r *templateRepo) DuplicateToTeam(templateID, targetTeamID int) (mapID int, zones []*Zone, err error)
```

**Comportement :**

1. Récupérer le template via `SELECT map_id, is_public FROM templates WHERE id = $1` :
   - Si aucun template trouvé → retourner `ErrTemplateNotFound`
   - Si `is_public = false` → retourner `ErrTemplateNotPublic` (nouvelle erreur à ajouter dans `errors.go`)
2. Récupérer `mapID` du template
3. **Supprimer toutes les zones existantes** de l'équipe cible pour cette carte :
   `SELECT id FROM zones WHERE team_id = $1 AND map_id = $2` puis pour chaque zone : `zoneRepo.Delete(zoneID)`
   (ou une seule requête `DELETE FROM zones WHERE team_id = $1 AND map_id = $2` si CASCADE gère les dépendances)
4. Récupérer toutes les `template_zones` pour ce template, triées par `order_index`
5. Pour chaque `template_zone` :
   a. Créer une nouvelle zone : `zoneRepo.Create(targetTeamID, mapID)`
   b. Parser `polygon_data` (JSON array) : `[{points: [...]}, ...]`. Pour chaque élément, extraire `points` et appeler `zoneRepo.AddPolygon(zoneID, points, orderIndex)`.
   c. Si le nom n'est pas vide, appeler `zoneRepo.SetName(zoneID, name)`.
6. Retourner `mapID` et la liste des zones créées

**Gestion d'erreurs :**
- Template introuvable → `ErrTemplateNotFound`
- Template non public → `ErrTemplateNotPublic`
- Erreur lors de la création → rollback des zones déjà créées pour cette opération, retourner l'erreur
- `polygon_data` incorrect (JSON invalide) → idem

**Note sur le parsing de `polygon_data` :**
Le champ `polygon_data` dans `template_zones` stocke un JSON array représentant tous les polygones d'une zone, au format :
```json
[
  {"points": [{"x": 0, "y": 0}, {"x": 10, "y": 0}, {"x": 5, "y": 10}]},
  {"points": [{"x": 100, "y": 100}, {"x": 200, "y": 100}, {"x": 150, "y": 200}]}
]
```
Chaque élément correspond à un polygone. Pour chaque élément, extraire le tableau `points` et le mapper en `[]models.Point`, puis appeler `zoneRepo.AddPolygon`.

#### Partie 2 - Handler `DuplicateTemplate`

Ajouter dans `internal/handlers/template.go` la méthode :

```go
func (h *TemplateHandler) DuplicateTemplate(c *gin.Context)
```

**Route :** `POST /api/v1/teams/:id/duplicate-template/:templateId`

**Comportement :**
1. Extraire `teamID` (param `id`) et `templateID` (param `templateId`) des paramètres
2. Récupérer l'utilisateur connecté via `middleware.MustGetUser(c)`
3. Appeler `templateRepo.DuplicateToTeam(templateID, teamID)` → reçoit `mapID` et `[]*Zone`
4. Appeler `zoneRepo.FindWithDetailsByTeamAndMap(teamID, mapID)` pour récupérer les zones détaillées (avec polygones et noms)
5. Retourner 201 Created :

```json
{
  "zones": [
    {
      "id": 1,
      "team_id": 2,
      "map_id": 1,
      "polygons": [...],
      "name": "Zone A",
      "created_at": "...",
      "updated_at": "..."
    },
    ...
  ],
  "template_id": 1,
  "map_id": 1
}
```

**Cas nominaux :**
- Capitaine/Cocapitaine importe un template → 201 + zones créées, les anciennes zones sont remplacées
- Template avec 0 zones → 201 + liste vide (les zones existantes sont supprimées)

**Gestion d'erreurs :**
- Template introuvable → 404 Not Found
- Template non public → 400 Bad Request
- Erreur DB → 500 Internal Server Error
- Paramètres invalides → 400 Bad Request

#### Partie 3 - Handler `ListPublicTemplates`

Ajouter dans `internal/handlers/template.go` la méthode :

```go
func (h *TemplateHandler) ListPublicTemplates(c *gin.Context)
```

**Route :** `GET /api/v1/templates/public`

**Comportement :**
1. Appeler `templateRepo.FindAllPublic()`
2. Retourner 200 OK avec la liste des templates (inclut `map_name` maintenant)

```json
{
  "templates": [
    {
      "template": { "id": 1, "original_team_id": 1, "map_id": 1, "map_name": "Atlantis", "is_public": true, "created_at": "...", "updated_at": "..." },
      "zones": [
        { "id": 1, "template_id": 1, "polygon_data": [...], "name": "Zone A", "order_index": 0 }
      ]
    }
  ]
}
```

**Cas nominaux :**
- Aucun template public → `{ "templates": [] }`
- Un ou plusieurs templates → liste complète avec `map_name`

**Gestion d'erreurs :**
- Erreur DB → 500 Internal Server Error

#### Partie 4 - Routes API

Modifier `api/routes.go` :

**Dans `SetupRoutes`** : Créer `TemplateHandler` et ajouter la route publique.

```go
func SetupRoutes(r *gin.Engine, cfg *config.AppConfig, logger services.Logger, frontendFS http.FileSystem, assetsFS http.FileSystem) {
    // ... existing code ...
    
    deps := newCoreDeps(cfg)
    
    db := database.DB()
    templateRepo := models.NewTemplateRepository(db)
    zoneRepo := models.NewZoneRepository(db)
    teamMemberRepo := models.NewTeamMemberRepository(db)
    templateHandler := handlers.NewTemplateHandler(templateRepo, zoneRepo, teamMemberRepo, logger)
    
    // Route publique des templates
    templatesPublic := apiV1.Group("/templates")
    templatesPublic.Use(middleware.RequireAuth(logger, deps.userRepo, deps.jwtService))
    {
        templatesPublic.GET("/public", templateHandler.ListPublicTemplates)
    }
    
    // Passer templateHandler à setupTeamRoutes
    setupTeamRoutes(apiV1, cfg, logger, templateHandler)
    
    // ... rest of SetupRoutes ...
}
```

Modifier la signature de `setupTeamRoutes` pour recevoir `templateHandler` :

```go
func setupTeamRoutes(rg *gin.RouterGroup, cfg *config.AppConfig, logger services.Logger, templateHandler *handlers.TemplateHandler)
```

**Dans `setupTeamRoutes`** : Ne plus créer `templateHandler`, utiliser le paramètre. Ajouter la route de duplication.

```go
// Dans le groupe teams, ajouter :
teams.POST("/:id/duplicate-template/:templateId",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireTeamRole(logger, teamMemberRepo, "cocaptain", "captain"),
    templateHandler.DuplicateTemplate,
)
```

**Note :** La route est protégée par `RequireTeamRole("cocaptain", "captain")`. Pas besoin de `RequirePermission` supplémentaire car le rôle captain/cocaptain a déjà `template:duplicate` par définition, et l'action est destructive (remplace toutes les zones).

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/template.go` | Modifier | Ajouter `MapName` à `Template`, modifier `scanTemplate` et requêtes `FindAllPublic`, `CreateFromMap`, `FindByMapID` pour JOIN avec `maps` |
| `internal/models/template.go` | Modifier | Ajouter implémentation de `DuplicateToTeam` |
| `internal/models/repository.go` | Modifier | Ajouter `DuplicateToTeam(templateID, teamID int) (int, []*Zone, error)` à `TemplateRepository` |
| `internal/models/errors.go` | Modifier | Ajouter `ErrTemplateNotPublic = errors.New("template is not public")` |
| `internal/handlers/template.go` | Modifier | Ajouter `DuplicateTemplate` et `ListPublicTemplates` |
| `api/routes.go` | Modifier | Créer `TemplateHandler` dans `SetupRoutes`, ajouter route `/templates/public`, passer `TemplateHandler` à `setupTeamRoutes`, ajouter route `POST /:id/duplicate-template/:templateId` |

### Signatures

```go
// Dans models/repository.go — ajouter à TemplateRepository
DuplicateToTeam(templateID, teamID int) (mapID int, zones []*Zone, err error)

// Dans models/template.go — nouveau champ
type Template struct {
    MapName string `json:"map_name"`
    // ... autres champs inchangés ...
}

// Dans handlers/template.go — nouvelles méthodes
func (h *TemplateHandler) ListPublicTemplates(c *gin.Context)
func (h *TemplateHandler) DuplicateTemplate(c *gin.Context)

// Dans api/routes.go — nouvelle signature
func setupTeamRoutes(rg *gin.RouterGroup, cfg *config.AppConfig, logger services.Logger, templateHandler *handlers.TemplateHandler)
```

### Contraintes techniques

- **Remplacement total** : L'import remplace TOUTES les zones existantes de l'équipe cible pour la carte. Les anciennes zones sont supprimées avant la création des nouvelles.
- **Template non modifié** : La duplication ne modifie pas le template original. Les `template_zones` ne sont pas touchées.
- **Anonymat** : Aucune information sur l'équipe qui duplique n'est stockée.
- **Vérification `is_public`** : `DuplicateToTeam` vérifie explicitement `is_public = true` et retourne `ErrTemplateNotPublic` sinon.
- **Permissions** : Réservé aux capitaines et cocapitaines (action destructive).
- **Consistance des données** : En cas d'erreur au milieu de la création des zones, les zones déjà créées doivent être supprimées.
- **Suppression des anciennes zones** : Utiliser `zoneRepo.Delete(zoneID)` pour chaque zone existante, ou un `DELETE ... WHERE team_id = $1 AND map_id = $2` direct si le CASCADE gère les dépendances (zone_polygons, zone_names, zone_versions, edit_history).
- **Transaction** : Idéalement, toute l'opération (suppression des anciennes zones + création des nouvelles) devrait être dans une transaction. Mais comme `DuplicateToTeam` utilise `zoneRepo` qui a ses propres transactions, soit :
  - Utiliser une transaction au niveau `templateRepo` avec `db.Begin()` et passer la transaction aux appels zoneRepo (nécessite que zoneRepo supporte une transaction externe)
  - Ou faire au mieux : supprimer d'abord, puis créer, et en cas d'échec de création, les zones supprimées sont perdues (acceptable pour une opération d'import)

### Tests à implémenter

Les tests seront créés dans la Tâche #003.

### Documentation

Aucune documentation à créer pour cette tâche. Voir Tâche #003.
