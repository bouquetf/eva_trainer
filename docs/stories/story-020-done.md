# Story #020 : Vue liste des zones en lecture

## Description
En tant qu'utilisateur, je veux un écran affichant la liste des zones d'une carte afin de pouvoir lire les noms des zones sans entrer dans l'éditeur.

## Origine
Feedback du 2025-05-14 : "Avoir un écran avec la liste des zones pour pouvoir lire"

## Critères d'acceptation
- [ ] Mode "lecture seule" activable dans le composant éditeur existant (même rendu carte + zones, mais fonctionnalités d'édition désactivées)
- [ ] Toutes les fonctionnalités d'édition sont désactivées dans ce mode : pas de dessin, pas de modification, pas de suppression de polygones
- [ ] Les zones sont affichées avec leurs noms sur la carte (comme dans l'éditeur)
- [ ] Aucune modification possible depuis cet écran (lecture seule)
- [ ] Les zones privées et publiques de l'équipe sont visibles
- [ ] L'écran est accessible à tous les membres de l'équipe (pas besoin du rôle Éditeur)
- [ ] Le filtrage par split ne s'applique pas à cette vue : toutes les zones de la carte pour l'équipe sont affichées

## Tests automatisés
- Test unitaire : Récupération de la liste des zones en lecture seule
- Test d'intégration : API retournant la liste des zones sans droits d'édition
- Test E2E : Membre consulte la liste des zones → voit les noms → pas de bouton d'édition

## Documentation
- docs/zone-list.md (à créer)

## Valeur utilisateur
Permet aux membres de consulter les noms des zones pour apprendre et réviser sans avoir besoin d'accéder à l'éditeur.

## Dépendances
- Story #009 (Zone Creation and Editing)
