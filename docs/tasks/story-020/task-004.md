# Tâche #004 - Story #020 : Documentation de la vue liste des zones

## Objectif

Créer la documentation `docs/zone-list.md` décrivant la vue lecture seule des zones d'une carte.

## Contexte

- Story #020 : `docs/stories/story-020.md` - Vue liste des zones en lecture
- Dépend de : Tâche #001, Tâche #002 (implémentation terminée)
- Ordre d'exécution : 4ème (après implémentation)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un fichier de documentation `docs/zone-list.md` décrivant la fonctionnalité de liste des zones en lecture seule.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/zone-list.md` | Créer | Documentation de la vue liste des zones |

### Contenu attendu

La documentation doit inclure :

1. **Titre** : "Liste des zones (lecture seule)"
2. **Accès** : URL `/teams/:id/zones/:mapId`, bouton "Voir les zones" dans les paramètres d'équipe
3. **Utilisation** :
   - Consultation des zones avec leurs noms sur la carte
   - Sélection d'une zone pour la surligner
   - Affichage du nombre de polygones et points par zone
4. **Permissions** : accessible à tous les membres de l'équipe (member, editor, cocaptain, captain)
5. **Limitations** :
   - Aucune édition possible (pas de dessin, modification, suppression)
   - Pas de verrouillage d'édition
   - Toutes les zones de la carte sont visibles (pas de filtrage par split)
6. **Cas particuliers** :
   - Carte sans zones → message "Aucune zone."
   - Paramètres d'URL invalides → message d'erreur

### Contraintes techniques
- **Format** : Markdown
- **Style** : Suivre le style de `docs/zone-editor.md` (même ton, même structure)
- **Langue** : Français

### Documentation à créer

- `docs/zone-list.md` : Documentation complète de la fonctionnalité

### Documentation à mettre à jour

Aucune mise àjour nécessaire.

### Exemple

```markdown
# Liste des zones (lecture seule)

## Accès

La liste des zones est accessible via l'URL :

```
/teams/:id/zones/:mapId
```

- `:id` = ID numérique de l'équipe
- `:mapId` = ID numérique de la carte

Ou depuis les paramètres d'équipe en cliquant sur **"Voir les zones"** dans la colonne Actions de la liste des cartes.

## Utilisation

### Consultation des zones

1. Ouvrir la vue liste des zones
2. La carte affiche toutes les zones avec leurs polygones et leurs noms
3. Cliquer sur une zone dans la liste pour la surligner sur la carte
4. La zone sélectionnée est mise en évidence (bordure et fond colorés)

### Informations affichées

Pour chaque zone, la liste affiche :
- Le **nom** de la zone
- Le **nombre de polygones**
- Le **nombre total de points**

## Permissions

Cette vue est accessible à tous les membres d'une équipe, quel que soit leur rôle :
- Membre
- Éditeur
- Co-capitaine
- Capitaine

## Limitations

- **Aucune modification possible** : Il n'est pas possible de créer, modifier, renommer ou supprimer des zones depuis cet écran
- **Pas de verrou d'édition** : Aucun verrou n'est acquis lors de la consultation
- **Toutes les zones sont visibles** : Le filtrage par split ne s'applique pas à cette vue

## Cas particuliers

- **Carte sans zones** : Le message "Aucune zone." s'affiche dans la liste
- **URL invalide** : Si les paramètres de l'URL sont incorrects, un message d'erreur s'affiche
```
