# Story #015 : Revision Mode with Validation

## Description
En tant qu'utilisateur, je veux réviser les zones d'une carte afin de m'entraîner à mémoriser les noms des zones.

## Critères d'acceptation
- [ ] Sélection d'une seule carte à la fois (pas de sélection multiple)
- [ ] Lors de l'affichage de la liste des cartes, on peut filtrer par split pour n'afficher que les cartes d'un split
- [ ] L'option "cartes du split" permet de filtrer la liste d'affichage (pas de révision multi-cartes simultanée)
- [ ] Affichage aléatoire d'une zone sur la carte sélectionnée
- [ ] L'utilisateur saisit le nom dans un champ de texte
- [ ] Validation insensible à la casse et aux accents (aucune autre tolérance)
- [ ] Si le nom correspond : coche verte affichée 1 seconde, puis question suivante
- [ ] Si le nom ne correspond pas : affichage du bon nom, l'utilisateur clique pour continuer
- [ ] Score affiché : nombre de bonnes réponses / nombre total de questions
- [ ] En cas d'actualisation (F5), la session en cours est perdue
- [ ] Pas de possibilité de mettre en pause une révision

## Tests automatisés
- Test unitaire : Validation de la réponse (insensible casse/accents)
- Test d'intégration : Lancement révision → récupération zones aléatoires → validation réponse
- Test E2E : Utilisateur lance révision → zone affichée → saisit nom correct → coche verte → question suivante

## Documentation
- docs/revision-mode.md (à créer)
- Voir specification.md section 3.2

## Valeur utilisateur
Outil d'entraînement efficace pour mémoriser les noms des zones sur les cartes du jeu Eva.

## Dépendances
- Story #009 (Zone Creation and Editing)
- Story #011 (Zone Publishing - zones publiques incluses)