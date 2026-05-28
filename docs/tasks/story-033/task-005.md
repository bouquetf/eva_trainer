# Tâche #005 - Story #033 : Endpoint POST de révision filtrée

## Objectif
Créer un nouvel endpoint POST `/api/v1/teams/{teamId}/maps/{mapId}/revision-items` avec un body `{zones: boolean, lines: boolean}` permettant de filtrer entre zones uniquement, lignes de tir uniquement, ou les deux. Mettre à jour le store Zustand et l'UI de révision.

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Tâche #002 (modèle line_of_fire), Tâche #003 (handler line_of_fire)
- Nécessaire pour : Tâche #008 (frontend révision)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Backend — Nouvel endpoint POST**

Créer un endpoint POST `/api/v1/teams/{teamId}/maps/{mapId}/revision-items` avec body JSON :

```json
{
  "zones": true,
  "lines": false
}
```

- `zones: true` → inclure les zones dans la révision
- `lines: true` → inclure les lignes de tir
- Les deux peuvent être `true` simultanément → mélanger les deux types
- Si les deux sont `false` → erreur 400 (`revision_filter_empty`)

Ajouter un champ `type` dans chaque élément retourné (zone ou ligne de tir) :

```go
type RevisionItem struct {
    ID       int                  `json:"id"`
    Source   string               `json:"source"`
    MapID    int                  `json:"map_id"`
    Polygons []models.ZonePolygon `json:"polygons"`
    Name     string               `json:"name"`
    Type     string               `json:"type"` // "zone" ou "line_of_fire"
}
```

**Logique de récupération**

- Si `zones: true` : récupérer les zones de l'équipe et des templates publics, les convertir en `RevisionItem` avec `type: "zone"`
- Si `lines: true` : récupérer les lignes de tir via `lineOfFireRepo.FindByTeamAndMap(teamID, mapID)` et `templateRepo.FindTemplateLinesOfFireByMapID(mapID)`, convertir chaque `LineOfFire` en `RevisionItem` avec `type: "line_of_fire"`
- Pour l'affichage, les `polygons` des lignes de tir doivent être convertis depuis les données de cône : chaque cône devient un polygone à 3 points (origin → left_end → right_end → origin), ou pour un segment droit (left_end == right_end), le polygone est juste origin → left_end → origin.
- Si les deux sont `true`, concaténer les listes et mélanger aléatoirement.

**Backend — RevisionRequest + handler**

Dans `internal/handlers/revision.go` :

```go
type RevisionFilterRequest struct {
    Zones bool `json:"zones"`
    Lines bool `json:"lines"`
}

type RevisionItem struct {
    ID       int                  `json:"id"`
    Source   string               `json:"source"`
    MapID    int                  `json:"map_id"`
    Polygons []models.ZonePolygon `json:"polygons"`
    Name     string               `json:"name"`
    Type     string               `json:"type"`
}
```

Ajouter `ListRevisionItems` handler :
- Lire `RevisionFilterRequest` depuis le body
- Valider qu'au moins un des deux champs est `true`
- Récupérer et retourner les items filtrés

**Backend — Route**

Ajouter dans `api/routes.go` :

```go
teams.POST("/:id/maps/:mapId/revision-items",
    middleware.RequireTeamRole(logger, teamMemberRepo, "member", "editor", "cocaptain", "captain"),
    middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "zone:view"),
    revisionHandler.ListRevisionItems,
)
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/revision.go` | Modifier | Ajouter structs + handler `ListRevisionItems` |
| `api/routes.go` | Modifier | Ajouter la route POST |

### Contraintes techniques
- Rétrocompatibilité : l'ancien endpoint GET `/revision-zones` reste inchangé.
- Les lignes de tir des templates publics doivent être incluses quand `lines: true`.
- La conversion des cônes en polygones pour l'affichage doit produire des polygones valides (3+ points).

### Exemples API

**Requête :**
```
POST /api/v1/teams/1/maps/5/revision-items
Content-Type: application/json

{"zones": true, "lines": true}
```

**Réponse :**
```json
{
  "items": [
    {
      "id": 1,
      "source": "team",
      "map_id": 5,
      "type": "zone",
      "polygons": [...],
      "name": "top"
    },
    {
      "id": 10,
      "source": "team",
      "map_id": 5,
      "type": "line_of_fire",
      "polygons": [
        {
          "id": 0,
          "zone_id": 10,
          "points": [{"x": 100, "y": 200}, {"x": 150, "y": 300}, {"x": 150, "y": 100}],
          "order_index": 0,
          "created_at": ""
        }
      ],
      "name": "couloir A"
    }
  ]
}
```

### Tests

- Tester avec `{"zones": true, "lines": false}` → seulement les zones
- Tester avec `{"zones": false, "lines": true}` → seulement les lignes de tir
- Tester avec `{"zones": true, "lines": true}` → zones + lignes de tir
- Tester avec `{"zones": false, "lines": false}` → erreur 400
