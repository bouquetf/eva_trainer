# Import/Export de zones

## Rôle requis
Administrateur uniquement. Les endpoints sont protégés par le middleware `RequireAdmin`.

## Export

### Format
L'export d'une carte unique produit un objet JSON contenant les zones avec leurs polygones.
L'export de toutes les cartes d'une équipe produit un tableau JSON de ces objets.

### Endpoints API

#### Exporter une carte spécifique
```
GET /api/v1/admin/teams/{team_id}/export/{map_id}
```

#### Exporter toutes les cartes d'une équipe
```
GET /api/v1/admin/teams/{team_id}/export
```

### Utilisation via l'interface admin
1. Aller dans l'onglet "Import/Export" du tableau de bord Admin
2. Sélectionner une équipe
3. Choisir une carte spécifique ou "Toutes les cartes"
4. Cliquer sur "Exporter"
5. Le fichier JSON est téléchargé automatiquement

## Import

### Format
Le format d'import est un tableau JSON d'objets carte. Chaque objet contient `map_name` (nom eva de la carte) et `zones` (liste des zones à importer). Les champs supplémentaires (comme `team` ou `export_date`) sont ignorés, ce qui permet de réutiliser directement un fichier d'export.

### Règles d'import
- La carte doit exister dans la base (uploadée au préalable via l'onglet "Cartes")
- Si la carte existe déjà pour l'équipe : les zones importées sont AJOUTÉES aux zones existantes (pas d'écrasement)
- L'import est traité carte par carte : chaque carte est importée dans sa propre transaction
- Si une zone échoue pour une carte (nom en conflit, polygone invalide), toutes les zones de cette carte sont annulées, mais les autres cartes sont traitées normalement
- Les noms de zones doivent être uniques par carte et par équipe
- Un polygone doit avoir au moins 3 points

### Endpoint API
```
POST /api/v1/admin/teams/{team_id}/import
```

Le body est toujours un tableau JSON, même pour une seule carte :
```json
[
  {
    "map_name": "Atlantis",
    "zones": [...]
  }
]
```

### Réponse
```json
{
  "imported_count": 5,
  "maps_processed": 2,
  "errors": []
}
```

- `imported_count` : nombre total de zones créées (toutes cartes confondues)
- `maps_processed` : nombre de cartes pour lesquelles l'import a réussi
- `errors` : tableau optionnel des erreurs par carte

### Utilisation via l'interface admin
1. Aller dans l'onglet "Import/Export" du tableau de bord Admin
2. Sélectionner l'équipe cible
3. Choisir un fichier JSON au format d'export
4. Cliquer sur "Importer"
5. Un message de confirmation indique le nombre de zones importées
6. Si des erreurs se produisent, un tableau détaillé est affiché

## Formats JSON

### Export d'une carte unique
```json
{
  "map_name": "Atlantis",
  "team": "Équipe Alpha",
  "export_date": "2026-05-15T12:00:00Z",
  "zones": [
    {
      "name": "Porte Nord",
      "polygons": [
        {
          "points": [
            { "x": 100, "y": 50 },
            { "x": 200, "y": 50 },
            { "x": 200, "y": 150 },
            { "x": 100, "y": 150 }
          ]
        }
      ]
    },
    {
      "name": "Porte Sud",
      "polygons": [
        {
          "points": [
            { "x": 100, "y": 300 },
            { "x": 200, "y": 300 },
            { "x": 150, "y": 400 }
          ]
        }
      ]
    },
    {
      "name": "Couloir Est",
      "polygons": [
        {
          "points": [
            { "x": 350, "y": 50 },
            { "x": 450, "y": 50 },
            { "x": 450, "y": 200 },
            { "x": 350, "y": 200 }
          ]
        },
        {
          "points": [
            { "x": 350, "y": 250 },
            { "x": 450, "y": 250 },
            { "x": 450, "y": 400 },
            { "x": 350, "y": 400 }
          ]
        }
      ]
    }
  ]
}
```

### Export de toutes les cartes d'une équipe
```json
[
  {
    "map_name": "Atlantis",
    "team": "Équipe Alpha",
    "export_date": "2026-05-15T12:00:00Z",
    "zones": [
      {
        "name": "Porte Nord",
        "polygons": [
          {
            "points": [
              { "x": 100, "y": 50 },
              { "x": 200, "y": 50 },
              { "x": 200, "y": 150 },
              { "x": 100, "y": 150 }
            ]
          }
        ]
      }
    ]
  },
  {
    "map_name": "Nouvelle-Calédonie",
    "team": "Équipe Alpha",
    "export_date": "2026-05-15T12:00:00Z",
    "zones": [
      {
        "name": "Baie de Prony",
        "polygons": [
          {
            "points": [
              { "x": 50, "y": 50 },
              { "x": 150, "y": 50 },
              { "x": 100, "y": 150 }
            ]
          }
        ]
      }
    ]
  }
]
```

### Import d'une carte unique
```json
[
  {
    "map_name": "Atlantis",
    "zones": [
      {
        "name": "Porte Nord",
        "polygons": [
          {
            "points": [
              { "x": 100, "y": 50 },
              { "x": 200, "y": 50 },
              { "x": 200, "y": 150 },
              { "x": 100, "y": 150 }
            ]
          }
        ]
      }
    ]
  }
]
```

### Import de plusieurs cartes
```json
[
  {
    "map_name": "Atlantis",
    "zones": [
      {
        "name": "Porte Nord",
        "polygons": [
          {
            "points": [
              { "x": 100, "y": 50 },
              { "x": 200, "y": 50 },
              { "x": 200, "y": 150 },
              { "x": 100, "y": 150 }
            ]
          }
        ]
      }
    ]
  },
  {
    "map_name": "Nouvelle-Calédonie",
    "zones": [
      {
        "name": "Baie de Prony",
        "polygons": [
          {
            "points": [
              { "x": 50, "y": 50 },
              { "x": 150, "y": 50 },
              { "x": 100, "y": 150 }
            ]
          }
        ]
      }
    ]
  }
]
```

## Codes d'erreur

### Erreurs HTTP globales

| Code HTTP | Code API | Description |
|-----------|----------|-------------|
| 400 | `invalid_json` | Le JSON est mal formé ou la structure est invalide |
| 400 | `invalid_team_id` | L'ID d'équipe est invalide |
| 400 | `invalid_map_id` | L'ID de carte est invalide (export uniquement) |
| 404 | `team_not_found` | L'équipe n'existe pas |
| 404 | `map_not_found` | La carte n'existe pas dans la base |
| 500 | `internal_error` | Erreur interne du serveur |

### Erreurs par carte (retournées dans `errors` lors de l'import)

| Message | Description |
|---------|-------------|
| `map not found` | La carte (nom eva) n'existe pas dans la base — uploader la carte d'abord |
| `zone name already exists` | Un nom de zone existe déjà sur cette carte pour cette équipe |
| `polygon must have at least 3 points` | Un polygone a moins de 3 points |
