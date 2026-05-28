# Lignes de tir

## Définition
Une ligne de tir représente une position d'où un joueur peut voir une zone de la carte, matérialisée par un cône de vision. Les lignes de tir sont complémentaires aux zones.

## Stockage
Les lignes de tir sont stockées dans une table dédiée `lines_of_fire`, distincte des zones. Chaque ligne contient :
- Un **nom** (unique par équipe et carte)
- Des **cônes** (un ou plusieurs) stockés dans `polygon_data` au format JSON
- Un **ordre** d'affichage

### Structure d'un cône
Chaque cône est défini par 3 points :
- **origin** : point de départ du cône (position du joueur)
- **left_end** : extrémité gauche du cône
- **right_end** : extrémité droite du cône

Si `left_end` et `right_end` sont identiques, la ligne de tir est un segment de droite (pas de cône visible).

## Création
1. Ouvrir l'éditeur de carte
2. Clic droit sur la carte → "Créer une ligne de tir"
3. Cliquer pour placer l'origine
4. Cliquer pour placer l'extrémité gauche
5. Cliquer pour placer l'extrémité droite
6. Saisir un nom
7. La ligne est automatiquement sauvegardée

## Ajout d'un cône
- Double-cliquer sur l'origine d'un cône existant pour ajouter un cône supplémentaire
- Le nouveau cône partage la même origine

## Modification
- Glisser l'origine : tous les cônes de la ligne suivent
- Glisser une extrémité : seul ce cône est modifié
- Clic droit sur une ligne : menu contextuel (Renommer, Supprimer, Ajouter un cône)

## Verrouillage
Le verrouillage d'édition (carte entière) s'applique aussi aux lignes de tir. Les mêmes règles que pour les zones s'appliquent.

## Templates
Les lignes de tir sont incluses dans :
- La publication d'une carte (création du template)
- La synchronisation automatique du template
- La duplication d'un template
- La révision d'une carte (lignes de tir publiques visibles avec `source: "template"`)

## Révision
Les lignes de tir peuvent être révisées via l'endpoint `POST /api/v1/teams/:id/maps/:mapId/revision-items` :
- `{"zones": true, "lines": false}` : révision classique (zones seules)
- `{"zones": false, "lines": true}` : uniquement les lignes de tir (équipe + templates publics)
- `{"zones": true, "lines": true}` : zones et lignes mélangées

Les lignes de tir des templates publics apparaissent avec `source: "template"`.

## Historique
Toutes les modifications sur les lignes de tir sont tracées dans l'historique (`line_of_fire_edit_history` avec `line_of_fire_id`). Les capitaines et admins peuvent effectuer un rollback complet via `line_of_fire_versions`.

## API

### Endpoints

| Méthode | Chemin | Description |
|---------|--------|-------------|
| GET | `/api/v1/teams/:id/maps/:mapId/lines-of-fire` | Lister les lignes de tir |
| POST | `/api/v1/teams/:id/maps/:mapId/lines-of-fire` | Créer une ligne de tir |
| POST | `/api/v1/teams/:id/maps/:mapId/lines-of-fire/batch` | Actions par lot |
| POST | `/api/v1/teams/:id/maps/:mapId/lines-of-fire/import` | Importer des lignes |
| GET | `/api/v1/teams/:id/maps/:mapId/lines-of-fire/export` | Exporter des lignes |
| GET | `/api/v1/teams/:id/lines-of-fire/:lineOfFireId` | Obtenir une ligne |
| DELETE | `/api/v1/teams/:id/lines-of-fire/:lineOfFireId` | Supprimer une ligne |
| PUT | `/api/v1/teams/:id/lines-of-fire/:lineOfFireId/name` | Renommer une ligne |
| POST | `/api/v1/teams/:id/lines-of-fire/:lineOfFireId/cones` | Ajouter un cône |
| PUT | `/api/v1/teams/:id/lines-of-fire/:lineOfFireId/cones/:coneIndex` | Modifier un cône |
| DELETE | `/api/v1/teams/:id/lines-of-fire/:lineOfFireId/cones/:coneIndex` | Supprimer un cône |
| POST | `/api/v1/teams/:id/maps/:mapId/revision-items` | Révision filtrée |

### Exemples

Création d'une ligne de tir avec 1 cône :

```bash
curl -X POST /api/v1/teams/1/maps/1/lines-of-fire \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Couloir A",
    "polygon_data": {
      "cones": [{
        "origin": {"x": 10, "y": 20},
        "left_end": {"x": 30, "y": 40},
        "right_end": {"x": 50, "y": 60}
      }]
    }
  }'
```

Révision avec lignes uniquement :

```bash
curl -X POST /api/v1/teams/1/maps/1/revision-items \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"zones": false, "lines": true}'
```
