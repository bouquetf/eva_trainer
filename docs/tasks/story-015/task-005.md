# Tâche #005 - Story #015 : Documentation du mode révision

## Objectif

Créer la documentation utilisateur et technique du mode révision.

## Contexte

- Story #015 : `docs/stories/story-015.md` - Revision Mode with Validation
- Dépend de : Tâche #001, #002, #003, #004 (toutes les tâches précédentes)
- Ordre d'exécution : 5ème (dernière)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - `docs/revision-mode.md`

Créer le fichier `docs/revision-mode.md` contenant :

```markdown
# Mode Révision

## Vue d'ensemble
Le mode révision permet aux membres d'une équipe de s'entraîner à mémoriser les noms des zones sur les cartes du jeu Eva.

## Prérequis
- Être membre d'une équipe
- L'équipe doit avoir défini des zones sur au moins une carte (ou la carte doit avoir un template public)

## Accès
- Naviguer vers `Mon équipe > Révision` ou directement `/teams/{id}/revision`

## Fonctionnement

### Sélection de carte
1. La liste des cartes disponibles s'affiche
2. Optionnellement, filtrer par split pour ne voir que les cartes d'un split spécifique
3. Sélectionner une **carte unique** (pas de sélection multiple)

### Configuration
1. Définir le nombre de questions (défaut : 20)
2. Cliquer sur "Lancer la révision"

### Déroulement
1. Une zone aléatoire est mise en évidence sur la carte
2. L'utilisateur saisit le nom de la zone dans un champ de texte
3. Validation :
   - **Correct** : coche verte pendant 1 seconde, passage automatique à la question suivante
   - **Incorrect** : affichage du bon nom, l'utilisateur clique pour continuer
4. Le score est mis à jour en temps réel

### Algorithme de sélection
- Chaque zone a un poids initial de 1
- En cas d'erreur : le poids augmente (+1)
- En cas de succès : le poids diminue (-0.5, minimum 1)
- La zone est choisie aléatoirement : probabilité proportionnelle au poids
- Après 2 succès consécutifs sur une zone, elle peut être retirée du pool

### Fin de session
- La session se termine automatiquement après le nombre de questions configuré
- L'utilisateur peut terminer la session prématurément
- Le score final s'affiche : bonnes réponses / total questions
- En cas d'actualisation (F5), la session en cours est perdue

## Notes techniques

### Zones disponibles
La révision inclut :
- Toutes les zones définies par l'équipe sur la carte sélectionnée
- Les zones des templates publics sur cette carte (publiés par d'autres équipes)
- Les zones des cartes obsolètes (même nom de carte)

### Validation
- Insensible à la casse : "TOP" = "top"
- Insensible aux accents : "Forêt" = "Foret"
- Espaces superflus ignorés (trim)
- Aucune autre tolérance (ponctuation, caractères spéciaux comptent)

### Session
- État uniquement en mémoire (Zustand store)
- Perdue en cas d'actualisation du navigateur
- Pas de fonctionnalité de pause

## API

### Endpoints

#### GET /api/v1/teams/{teamId}/maps/{mapId}/revision-zones

Retourne toutes les zones disponibles pour la révision.

**Réponse :**
```json
{
  "zones": [
    {
      "id": 1,
      "source": "team",
      "map_id": 1,
      "polygons": [...],
      "name": "top"
    },
    {
      "id": 5,
      "source": "template",
      "map_id": 1,
      "polygons": [...],
      "name": "secret"
    }
  ]
}
```

- `source`: `"team"` (zone de l'équipe) ou `"template"` (zone d'un template public)
- `polygons`: tableau de polygones pour l'affichage sur la carte
- `name`: nom attendu (réponse correcte)
```

#### Partie 2 - Mise à jour de `docs/specification/specification.md`

Aucune mise à jour nécessaire : la spécification contient déjà la section 3.2 complète.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/revision-mode.md` | Créer | Documentation complète du mode révision |

### Contraintes techniques

- **Format** : Markdown standard, compatible GitHub
- **Style** : Suivre le style des autres fichiers de documentation dans `docs/`
- **Contenu** : Couvrir l'utilisation utilisateur ET les notes techniques pour les développeurs
- **Pas de redondance** : Ne pas dupliquer la spécification, seulement la documenter sous forme de guide

### Documentation

Cette tâche EST la documentation. Aucune autre documentation à créer.

### Tests

Aucun test nécessaire pour cette tâche.
