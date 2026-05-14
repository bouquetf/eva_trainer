# Story #021 : Import/Export de zones (Admin)

## Description
En tant qu'administrateur, je veux exporter et importer les zones d'une carte ou de toutes les cartes d'une équipe afin de sauvegarder et transférer des configurations entre instances.

## Origine
Feedback du 2025-05-14 : "Import/Export de zones"

## Critères d'acceptation

### Export
- [ ] Export d'une carte et de toutes ses zones associées pour une équipe au format JSON
- [ ] Export de toutes les cartes d'une équipe (cartes + zones associées) au format JSON
- [ ] Les zones sont exportées avec le lien vers la carte identifiée par son nom eva
- [ ] L'export est accessible aux administrateurs uniquement
- [ ] Le JSON contient : polygones, noms de zones, relation carte (nom eva), métadonnées d'équipe
- [ ] Format JSON d'export :

```json
{
  "map_name": "Atlantis",
  "team": "...",
  "export_date": "...",
  "zones": [
    {
      "name": "top",
      "polygons": [
        { "points": [0, 0, 100, 0, 100, 100, 0, 100] }
      ]
    }
  ]
}
```

### Import
- [ ] Import d'une carte et de ses zones depuis un fichier JSON au même format que l'export
- [ ] Import de toutes les cartes d'une équipe depuis un fichier JSON
- [ ] Validation du fichier importé (structure, types de données, intégrité des références)
- [ ] Si la carte (nom eva) existe déjà : les zones importées sont ajoutées aux existantes
- [ ] Si la carte (nom eva) n'existe pas : l'import crée la carte dans l'équipe cible
- [ ] L'import est accessible aux administrateurs uniquement
- [ ] En cas d'import, les zones existantes ne sont pas écrasées (l'import ajoute les nouvelles zones)

## Tests automatisés
- Test unitaire : Validation du format JSON d'export/import
- Test d'intégration : Export d'une carte avec zones → réimport dans une autre équipe → vérification
- Test E2E : Admin exporte toutes les cartes d'une équipe → admin importe dans une autre équipe → zones visibles dans l'éditeur

## Documentation
- docs/import-export.md (à créer)
- Mettre à jour docs/specification/specification.md section 3.1 (supprimer "aucun format d'import/export")

## Valeur utilisateur
Permet la sauvegarde locale des configurations de zones et le transfert entre instances par les administrateurs.

## Dépendances
- Story #009 (Zone Creation and Editing)
- Story #005 (Map Management)
