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
