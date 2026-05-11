# Tâche #001 - Story #016 : Backend - Inclure template zones et cartes obsolètes dans l'endpoint révision

## Objectif
Mettre à jour l'endpoint `GET /api/v1/teams/:id/maps/:mapId/revision-zones` pour inclure les zones des templates publics ET les zones des cartes obsolètes (toutes les versions d'une même carte).

## Contexte
- Story #016 : `docs/stories/story-016.md` - Revision Mode with Weighted Probability
- Story #015 : `docs/stories/story-015.md` - Revision Mode with Validation (dépendance)
- Dépend de : Story #015 (endpoint `ListRevisionZones` existant), Story #011 (modèle `Template`, `TemplateRepository`)
- Nécessaire pour : Tâche #002 (frontend affichage), Tâche #003 (tests)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Comportement attendu (modification de `ListRevisionZones`)

Actuellement, l'endpoint ne retourne que les zones de l'équipe pour le `mapId` donné. Il doit maintenant :

1. Résoudre le `mapId` en nom de carte via `mapRepo.FindByID(mapID)`
2. Trouver TOUTES les cartes portant le même nom (versions actuelle + obsolètes, non supprimées) via `mapRepo.FindByName(name)`
3. Pour CHAQUE version de la carte (y compris les obsolètes) :
   - Récupérer les zones de l'équipe via `zoneRepo.FindWithDetailsByTeamAndMap(teamID, mapVersion.ID)`
   - Récupérer les zones du template public via `templateRepo.FindTemplateZonesByMapID(mapVersion.ID)`
4. Combiner toutes les zones dans une seule liste `RevisionZone[]`
5. Retourner `{"zones": [...]}`

**Cas nominaux :**
- Carte "Atlantis" a 2 versions (v1 active, v0 obsolète) → les zones des 2 versions sont incluses
- Carte avec template public → zones équipe + zones template sont incluses
- Carte sans template public → seules les zones équipe sont retournées

**Cas limites :**
- Carte sans aucune zone → `{"zones": []}`
- Carte avec template public mais sans zones équipe → zones template uniquement
- Carte obsolète supprimée (hard delete) → ignorée (déjà filtrée par `FindByName`)
- Toutes les cartes du même nom ont des zones qui se chevauchent → pas de déduplication (garder toutes les zones)

**Gestion d'erreurs :**
- Carte introuvable → 404 Not Found (inchangé)
- Erreur DB → 500 Internal Server Error (inchangé)
- Utilisateur non membre → 403 Forbidden (inchangé, géré par middleware)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/revision.go` | Modifier | Ajouter `templateRepo` au handler, modifier `ListRevisionZones` |
| `api/routes.go` | Modifier | Passer `templateRepo` à `NewRevisionHandler` |

### Signatures

#### RevisionHandler modifié

```go
// internal/handlers/revision.go
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

#### Logique de ListRevisionZones modifiée

```go
func (h *RevisionHandler) ListRevisionZones(c *gin.Context) {
    // 1. Extraire mapID et teamID des paramètres
    // 2. Trouver la carte par ID → récupérer son nom
    // 3. Trouver toutes les cartes avec ce nom (FindByName)
    // 4. Pour chaque carte :
    //    a. zones équipe
    //    b. zones template (si template public existe)
    // 5. Retourner toutes les zones combinées
}
```

#### Route mise à jour

```go
// api/routes.go
revisionHandler := handlers.NewRevisionHandler(zoneRepo, templateRepo, mapRepo, logger)
```

### Contraintes techniques

- **`templateRepo` déjà disponible** : `templateRepo` existe déjà dans `setupTeamRoutes` (ligne 125). Il suffit de le passer au `NewRevisionHandler`
- **Dédoublonnage** : Ne PAS dédupliquer les zones. Si deux versions de carte ont une zone nommée "top", les deux doivent apparaître (elles ont des `zone_id` différents). Le poids est géré côté frontend
- **Source field** : Chaque `RevisionZone` doit conserver son `source` ("team" ou "template") correct
- **MapID field** : Chaque `RevisionZone` doit avoir le `map_id` de la version de carte dont elle provient, pas le `mapID` original. Le frontend utilise ce champ pour charger l'image correcte
- **Ordre des zones** : Les zones sont retournées dans l'ordre des versions de carte (de la plus récente à la plus ancienne), puis zones équipe avant zones template pour une même version
- **Template zones** : Utiliser la méthode existante `templateRepo.FindTemplateZonesByMapID(mapID)` qui retourne `[]TemplateZone`. Convertir `polygon_data` (JSONB) en `[]models.ZonePolygon` comme déjà fait côté handler

### Tests à implémenter

Les tests existants dans `internal/handlers/revision_test.go` doivent continuer à passer sans modification. Les nouveaux scénarios doivent être ajoutés dans la Tâche #003.

### Documentation

Aucune documentation à créer pour cette tâche. La documentation `docs/revision-mode.md` mentionne déjà que les cartes obsolètes sont incluses.
