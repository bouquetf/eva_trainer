# Story #007 : Split Management and Filtering

## Description
En tant qu'utilisateur, je veux filtrer les cartes par split (saison) afin de m'entraîner sur un ensemble de cartes spécifique.

## Critères d'acceptation
- [ ] Un admin peut créer un split avec : nom, liste des noms de cartes, date début, date fin
- [ ] Seuls les Admins peuvent créer, modifier et supprimer des splits
- [ ] Lors de la révision, l'utilisateur peut cocher "cartes du split" pour filtrer
- [ ] Le filtrage par split affiche uniquement les cartes dont le nom est dans le split
- [ ] Réviser les cartes d'un split peut conduire à réviser les cartes obsolètes (le split référence les noms)
- [ ] API fournissant la liste des cartes disponibles pour un nom "eva" donné

## Tests automatisés
- Test unitaire : Création de split et validation des dates
- Test d'intégration : Filtrage par split en révision → vérification des cartes retournées
- Test E2E : Création split → lancement révision avec filtre split → seules les cartes du split sont proposées

## Documentation
- docs/splits.md (à créer)
- Voir specification.md section 1 (Splits)

## Valeur utilisateur
Permet aux utilisateurs de se concentrer sur des ensembles de cartes spécifiques (saisons du jeu Eva).

## Dépendances
- Story #005 (Map Management)