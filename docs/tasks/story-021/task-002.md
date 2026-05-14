# Tâche #002 - Story #021 : Import des zones — Backend (Go)

## Objectif
Implémenter l'API backend d'import des zones : endpoint administrateur pour importer des zones depuis un fichier JSON, avec validation du format et création transactionnelle.

## Contexte
- Story #021 : [Import/Export de zones (Admin)](../../stories/story-021-done.md)
- Dépend de : Tâche #001 (Export backend — réutilisation du format JSON et des structures)
- Nécessaire pour : Tâche #003 (Frontend Admin UI), Tâche #004 (Documentation)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter un endpoint administrateur pour importer des zones depuis un fichier JSON au même format que l'export.

**Format JSON d'import attendu (toujours un tableau, même pour une seule carte) :**

```json
// Import d'une seule carte (tableau avec 1 élément)
[
  {
    "map_name": "Atlantis",
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
]

// Import de plusieurs cartes
[
  { "map_name": "Atlantis", "zones": [...] },
  { "map_name": "Lunar", "zones": [...] }
]
```

**Cas nominaux :**
- Import d'une carte (nom eva) qui existe dans la base `maps` → les zones importées sont ajoutées aux zones existantes de cette carte pour l'équipe (pas d'écrasement)
- Import de plusieurs cartes → chaque carte est traitée indépendamment (carte par carte)
- L'input est toujours un tableau JSON `[{ map_name, zones }]` même pour une seule carte

**Cas limites :**
- JSON mal formé → erreur de parsing avec message clair
- Carte (nom eva) introuvable dans la table `maps` → erreur 404 "map not found" (la carte doit être uploadée au préalable par un admin)
- Zone avec un nom déjà existant sur la carte → l'import échoue pour CETTE carte avec 409 (les cartes précédentes sont déjà importées)
- Points en nombre < 3 pour un polygone → erreur de validation pour la carte concernée (les autres cartes ne sont pas affectées)
- Tableau vide → retourne 200 avec 0 zones importées

**Gestion d'erreurs :**
- `invalid_json` → 400 (JSON mal formé ou structure invalide)
- `map_not_found` → 404 (nom eva de carte introuvable dans la base maps)
- `zone_name_taken` → 409 (nom de zone déjà existant — l'utilisateur doit renommer avant réimport)
- `invalid_polygon` → 400 (polygone avec moins de 3 points)
- `invalid_body` → 400 (body n'est pas un tableau)
- `internal_error` → 500

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/import_export.go` | Modifier | Ajouter les méthodes d'import et validation |
| `internal/models/repository.go` | Modifier | Ajouter `ImportZones` à `ZoneRepository` interface |
| `internal/models/zone.go` | Modifier | Ajouter l'implémentation de `ImportZones` (transactions) |
| `internal/models/errors.go` | Modifier | Ajouter les nouvelles erreurs si nécessaire |
| `api/routes.go` | Modifier | Ajouter la route POST d'import dans `setupAdminRoutes` |
| `internal/handlers/import_export_test.go` | Créer | Tests d'import |

### Signatures

```go
// internal/models/repository.go — à ajouter à ZoneRepository

// ImportZoneData représente une zone à importer avec ses polygones et son nom
type ImportZoneData struct {
    Name     string              `json:"name"`
    Polygons []ImportPolygonData `json:"polygons"`
}

type ImportPolygonData struct {
    Points []Point `json:"points"`
}

// ImportZones importe des zones sur une carte pour une équipe.
// Toute l'opération est transactionnelle : tout réussit ou tout échoue.
// Si une zone avec le même nom existe déjà, ErrZoneNameTaken est retourné.
ImportZones(teamID, mapID int, zones []ImportZoneData) ([]*Zone, error)
```

```go
// internal/handlers/import_export.go — à ajouter

// ImportMapRequest représente une carte avec ses zones à importer
type ImportMapRequest struct {
    MapName string                  `json:"map_name"`
    Zones   []models.ImportZoneData `json:"zones"`
}

// ImportTeamZones importe des zones pour une équipe à partir d'un fichier JSON.
// Le body est TOUJOURS un tableau d'ImportMapRequest, même pour une seule carte.
// POST /api/v1/admin/teams/:id/import
// Body: [{ "map_name": "Atlantis", "zones": [...] }]
func (h *ImportExportHandler) ImportTeamZones(c *gin.Context)
```

```go
// internal/models/zone.go — nouvelle méthode sur zoneRepo
func (r *zoneRepo) ImportZones(teamID, mapID int, zones []ImportZoneData) ([]*Zone, error) {
    tx, err := r.db.Begin()
    if err != nil { return nil, fmt.Errorf("...") }
    defer tx.Rollback()
    
    var createdZones []*Zone
    for _, z := range zones {
        // Vérifier si le nom existe déjà
        taken, err := checkNameTakenInTx(tx, teamID, mapID, z.Name)
        if err != nil { return nil, err }
        if taken { return nil, ErrZoneNameTaken }
        
        // Créer la zone
        zone, err := createZoneInTx(tx, teamID, mapID)
        if err != nil { return nil, err }
        
        // Ajouter les polygones
        for i, p := range z.Polygons {
            if len(p.Points) < 3 { return nil, ErrPolygonHasTooFewPoints }
            if err := addPolygonInTx(tx, zone.ID, p.Points, i); err != nil { return nil, err }
        }
        
        // Définir le nom
        if err := setZoneNameInTx(tx, zone.ID, z.Name, teamID, mapID); err != nil { return nil, err }
        
        createdZones = append(createdZones, zone)
    }
    
    if err := tx.Commit(); err != nil { return nil, err }
    return createdZones, nil
}
```

```go
// api/routes.go — dans setupAdminRoutes
admin.POST("/teams/:id/import", importExportHandler.ImportTeamZones)
```

### Contraintes techniques
- **Format d'input** : le body est TOUJOURS un tableau JSON, même pour une seule carte. Valider avec `c.ShouldBindJSON(&[]ImportMapRequest{})`.
- **Transaction** : chaque carte est traitée dans sa propre transaction (carte par carte). Si une zone échoue pour une carte, SEULE cette carte est annulée. Les cartes précédentes sont déjà importées.
- **Recherche de carte par nom** : utiliser `mapRepo.FindByNameLatest(name)` pour trouver la carte la plus récente par son nom eva. Si introuvable → retourner 404 (la carte doit exister en base, uploadée au préalable).
- **Pas d'écrasement** : les zones existantes ne sont JAMAIS supprimées ou modifiées
- **Gestion des noms de zones** : si un nom de zone existe déjà, l'import de CETTE carte échoue avec `ErrZoneNameTaken` (status 409), mais les cartes précédentes restent importées
- **Réutilisation de code** : s'inspirer de `batchCreateZone` dans `zone.go` pour la logique de création (mais en transactionnel)
- **Structure de retour** : retourner un objet avec `imported_count` (nombre total de zones importées), `maps_processed` (nombre de cartes traitées), et optionnellement `errors` pour les cartes qui ont échoué

### Tests à implémenter

#### Tests unitaires — `internal/handlers/import_export_test.go`
- **Scénario 1** : Import d'une carte avec des zones valides → création réussie
  - Mock zoneRepo.ImportZones() retourne les zones créées
  - Vérifier : statut 200, body contient les zones créées
- **Scénario 2** : Import avec JSON mal formé → 400
  - Body: `{invalid}`
  - Vérifier : statut 400, erreur "invalid_json"
- **Scénario 3** : Import d'une carte inexistante → 404
  - Mock mapRepo.FindByNameLatest() retourne ErrMapNotFound
  - Vérifier : statut 404
- **Scénario 4** : Import avec zone déjà existante → 409
  - Mock zoneRepo.ImportZones() retourne ErrZoneNameTaken
  - Vérifier : statut 409
- **Scénario 5** : Import multiple (tableau de cartes) → traitement de toutes les cartes
- **Scénario 6** : Import d'une carte qui existe déjà → zones ajoutées aux existantes

#### Tests d'intégration
- Créer une équipe + carte + zone → exporter → importer dans une autre équipe → vérifier les zones créées
- Importer les mêmes zones deux fois → la seconde tentative échoue avec 409 (conflit de nom)

### Documentation

Ne pas créer de documentation ici. Voir Tâche #004.

### Exemples d'utilisation

```bash
# Import d'une carte avec ses zones (toujours un tableau)
curl -X POST -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '[{"map_name":"Atlantis","zones":[{"name":"top","polygons":[{"points":[{"x":0,"y":0},{"x":100,"y":0},{"x":100,"y":100},{"x":0,"y":100}]}]}]}]' \
  http://localhost:8080/api/v1/admin/teams/1/import

# Import multiple
curl -X POST -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '[{"map_name":"Atlantis","zones":[...]},{"map_name":"Lunar","zones":[...]}]' \
  http://localhost:8080/api/v1/admin/teams/1/import
```
