# Story #020 : Vue liste des zones en lecture

## Description
En tant qu'utilisateur, je veux un écran affichant la liste des zones d'une carte afin de pouvoir lire les noms des zones sans entrer dans l'éditeur.

## Origine
Feedback du 2025-05-14 : "Avoir un écran avec la liste des zones pour pouvoir lire"

## Critères d'acceptation
- [ ] Nouvel écran accessible depuis la page d'une carte : affiche la même vue que l'éditeur (carte + zones)
- [ ] Toutes les fonctionnalités d'édition sont désactivées (pas de dessin, pas de modification, pas de suppression)
- [ ] Les zones sont affichées avec leurs noms sur la carte (comme dans l'éditeur)
- [ ] Aucune modification possible depuis cet écran (lecture seule)
- [ ] Les zones privées et publiques de l'équipe sont visibles
- [ ] L'écran est accessible à tous les membres de l'équipe (pas besoin du rôle Éditeur)

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
