# Tâche #010 - Story #033 : Documentation

## Objectif
Créer le document `docs/lines-of-fire.md` décrivant le fonctionnement des lignes de tir.

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Tâche #003 (API), Tâche #007 (éditeur)
- Nécessaire pour : Aucune

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer `docs/lines-of-fire.md` avec le contenu suivant :

```markdown
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

## Révision
Les lignes de tir peuvent être révisées :
- "Zones uniquement" : révision classique (zones seules)
- "Lignes uniquement" : uniquement les lignes de tir
- "Les deux" : zones et lignes mélangées aléatoirement

## Historique
Toutes les modifications sur les lignes de tir sont tracées dans l'historique (`edit_history` avec `line_of_fire_id`). Les capitaines et admins peuvent effectuer un rollback complet.

## API

### Endpoints
```

Voir `docs/tasks/story-033/task-003.md` pour la liste complète des endpoints.

### Exemples
```

...
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/lines-of-fire.md` | Créer | Documentation complète |
