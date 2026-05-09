# Éditeur de zones polygonales

## Accès

L'éditeur est accessible via l'URL :

```
/teams/:id/editor/:mapId
```

- `:id` = ID numérique de l'équipe
- `:mapId` = ID numérique de la carte

## Utilisation

### Créer une zone

1. Sélectionner l'outil **Dessiner** dans la barre d'outils
2. Double-cliquer sur la carte pour ajouter des points
3. Terminer la forme via le clic droit → "Terminer la forme"
4. Saisir un nom pour la zone (max 250 caractères)
5. La zone est automatiquement sauvegardée

### Sélectionner des zones

- **Clic simple** : sélectionne une seule zone
- **Ctrl/Clic** ou **Shift/Clic** : ajoute/retire une zone à la sélection
- **Tout sélectionner** : sélectionne toutes les zones

### Modifier une zone

- **Déplacer** : glisser une zone sélectionnée
- **Éditer les points** : outil "Éditer Points", glisser les poignées blanches
- **Ajouter un point** : double-cliquer sur un segment de polygone
- **Supprimer un point** : clic droit sur une poignée
- **Flip H/V** : miroir horizontal/vertical via le menu contextuel

### Supprimer

- **Supprimer la sélection** : bouton dans la barre d'outils
- **Supprimer une zone** : clic droit sur la zone → "Supprimer" ou bouton ✕ dans la liste
- **Reset** : supprime toutes les zones (après confirmation)

### Verrouillage d'édition

- Le verrou est acquis automatiquement au chargement de l'éditeur
- Il est rafraîchi toutes les 10 minutes
- Il est libéré automatiquement en quittant l'éditeur
- Si la carte est verrouillée par un autre éditeur, l'édition est en lecture seule
- Le verrou expire après 15 minutes d'inactivité

### Sauvegarde automatique

- Toutes les actions sont automatiquement sauvegardées via un système de batch
- Un indicateur en haut à droite montre l'état de la sauvegarde :
  - **Vert** : sauvegardé
  - **Orange** : sauvegarde en cours
  - **Rouge** : échec (nouvelle tentative automatique après 5 secondes)
- En cas d'échec persistant après 5 tentatives, une erreur critique est affichée

## Cas particuliers

- **Carte obsolète** : impossible de créer ou modifier des zones
- **Perte de connexion** : les actions sont conservées et retentées automatiquement
- **Fermeture du navigateur** : le verrou expire après 15 minutes
- **Polygone minimum** : 3 points minimum par polygone
