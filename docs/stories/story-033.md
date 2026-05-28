# Story #033 : Lignes de tir — Édition et révision

## Description
En tant qu'éditeur, je veux ajouter et modifier des lignes de tir sur les cartes afin de visualiser les cônes de tir. En tant que membre, je veux pouvoir réviser les lignes de tir en plus des zones.

## Origine
Feedback du 2026-05-27 : "Éditeur : ajouter/éditer des lignes de tir, révision sur les lignes. C'est une autre table que les zones. Pour la révision, offrir la possibilité de réviser sur les lignes ou les zones, ou les deux."

## Critères d'acceptation
- [ ] L'éditeur permet de créer une ligne de tir via clic droit → "Créer une ligne de tir"
- [ ] Une ligne de tir est définie par 3 points : origine (départ), extrémité gauche du cône, extrémité droite du cône
- [ ] Si les extrémités gauche et droite sont identiques, la ligne est un segment de droite (pas de cône)
- [ ] Un double-clic sur l'origine ajoute un cône supplémentaire à la même ligne (même centre)
- [ ] Les lignes de tir sont stockées dans une table dédiée (distincte des zones) avec la même structure qu'une zone (id, équipe, carte, polygon_data, nom, ordre), la polygon_data stockant les points du/des cônes
- [ ] Chaque ligne de tir a un nom (comme les zones)
- [ ] Le verrouillage existant (carte entière) s'applique aussi aux lignes de tir (A)
- [ ] Les lignes de tir sont incluses dans la publication et la duplication des templates
- [ ] La révision propose un filtre au lancement : "Zones uniquement", "Lignes uniquement", "Les deux" (pioche aléatoire dans les deux)
- [ ] Les lignes de tir sont soumises à l'historique des modifications (rollback possible)

## Tests automatisés
- Test unitaire : CRUD des lignes de tir et de leurs cônes en base de données
- Test d'intégration : API endpoints pour les lignes de tir
- Test E2E : Création d'une ligne de tir avec 3 points, modification, suppression
- Test E2E : Double-clic sur origine pour ajouter un cône supplémentaire
- Test E2E : Segment de droite (extrémités identiques)
- Test E2E : Révision avec filtre "Lignes uniquement" et "Les deux"

## Documentation
- docs/lines-of-fire.md (à créer)

## Valeur utilisateur
Permet de cartographier les lignes et cônes de tir sur chaque carte en complément des zones, et de les réviser séparément ou conjointement avec les zones.

## Dépendances
- Story #002 (Map Editor)
- Story #005 (Revision Mode)
