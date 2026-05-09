# Tâche #005 - Story #007 : Documentation des Splits

## Objectif

Créer la documentation utilisateur pour la fonctionnalité de gestion des splits, expliquant comment créer, gérer et utiliser les splits pour filtrer les révisions.

## Contexte

- Story #007 : `docs/stories/story-007.md` - Split Management and Filtering
- Dépend de : Tâche #001, #002 (fonctionnalité implémentée)
- Ordre d'exécution : 5ème (après l'implémentation)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer le fichier de documentation `docs/splits.md` qui explique :
- Qu'est-ce qu'un split (saison du jeu Eva)
- Comment les splits sont utilisés pour filtrer les cartes pendant la révision
- Que les splits référencent les noms de cartes (pas les IDs), ce qui inclut les cartes obsolètes
- La gestion des splits par les administrateurs uniquement

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/splits.md` | Créer | Documentation utilisateur des splits |

### Contenu attendu

```markdown
# Splits (Saisons)

## Qu'est-ce qu'un split ?

Un split représente une **saison** du jeu Eva. Chaque split contient un sous-ensemble de cartes jouables, identifiées par leur nom.

Les splits permettent de filtrer les cartes pendant la révision pour se concentrer sur un ensemble spécifique de cartes.

## Structure d'un split

Un split est défini par :
- **Nom** : Identifiant du split (ex: "Season 1", "Prerelease")
- **Noms de cartes** : Liste des noms de cartes Eva incluses dans ce split (ex: ["Atlantis", "Lunar Outpost", "Polaris"])
- **Date de début** (optionnelle) : Date de début de la saison
- **Date de fin** (optionnelle) : Date de fin de la saison

## Gestion des splits

La création, la modification et la suppression des splits sont réservées aux **administrateurs**.

### API Admin

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/v1/admin/splits` | Créer un split |
| GET | `/api/v1/admin/splits` | Lister tous les splits |
| GET | `/api/v1/admin/splits/:id` | Détail d'un split |
| PUT | `/api/v1/admin/splits/:id` | Modifier un split |
| DELETE | `/api/v1/admin/splits/:id` | Supprimer un split |
| GET | `/api/v1/admin/splits/:id/maps` | Cartes d'un split |

### API Publique (authentifiée)

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/v1/splits` | Lister les splits |
| GET | `/api/v1/splits/:id/maps` | Cartes d'un split |

## Filtrage par split pendant la révision

Lors de la révision, l'utilisateur peut cocher l'option "cartes du split" pour filtrer l'affichage des cartes. Seules les cartes dont le nom est présent dans le split sélectionné sont proposées.

**Note importante** : Comme le split référence les **noms** de cartes (pas les IDs), réviser les cartes d'un split peut inclure des cartes obsolètes. Par exemple, si "Atlantis" est dans un split et qu'il existe deux versions d'Atlantis (v1 obsolète, v2 actuelle), les deux peuvent apparaître pendant la révision.

## Exemple de création d'un split

```json
POST /api/v1/admin/splits
{
    "name": "Season 1",
    "map_names": ["Atlantis", "Lunar Outpost", "Polaris"],
    "start_date": "2024-01-01",
    "end_date": "2024-06-30"
}
```

```json
{
    "split": {
        "id": 1,
        "name": "Season 1",
        "map_names": ["Atlantis", "Lunar Outpost", "Polaris"],
        "start_date": "2024-01-01T00:00:00Z",
        "end_date": "2024-06-30T00:00:00Z"
    }
}
```

## Résolution des cartes d'un split

```json
GET /api/v1/admin/splits/1/maps
{
    "maps": [
        {"id": 1, "name": "Atlantis", "is_obsolete": false, ...},
        {"id": 2, "name": "Atlantis", "is_obsolete": true, ...},
        {"id": 3, "name": "Lunar Outpost", "is_obsolete": false, ...},
        {"id": 4, "name": "Polaris", "is_obsolete": false, ...}
    ]
}
```

## Consultation des cartes par nom

```json
GET /api/v1/admin/maps/by-name/Atlantis
{
    "maps": [
        {"id": 1, "name": "Atlantis", "is_obsolete": false, ...},
        {"id": 2, "name": "Atlantis", "is_obsolete": true, ...}
    ]
}
```
```

### Contraintes techniques

- **Format** : Markdown standard.
- **Style** : Utiliser le même format que les autres fichiers de documentation existants dans `docs/`.
- **Exemples** : Les exemples d'API doivent être précis et correspondre aux réelles réponses des endpoints.

### Tests à implémenter

Aucun test pour cette tâche de documentation.

### Documentation

Cette tâche consiste à créer la documentation elle-même.
