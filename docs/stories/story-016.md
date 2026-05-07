# Story #016 : Revision Mode with Weighted Probability

## Description
En tant qu'utilisateur, je veux que les zones sur lesquelles je fais des erreurs soient proposées plus fréquemment afin d'optimiser mon entraînement.

## Critères d'acceptation
- [ ] Chaque zone a un poids initial de 1
- [ ] En cas d'erreur, le poids augmente (+1)
- [ ] En cas de succès, le poids diminue (-0.5, minimum 1)
- [ ] La zone est choisie aléatoirement avec une probabilité proportionnelle au poids
- [ ] Les poids sont uniquement en mémoire pour la session en cours (non persistants)
- [ ] Si l'utilisateur ne fait aucune erreur sur une question, sur deux tentatives, la zone est retirée de la session
- [ ] Après 2 succès consécutifs sur une zone, celle-ci n'apparaît plus dans la session en cours
- [ ] Toutes les zones définies par l'équipe (ou publiques) sont incluses
- [ ] Les zones des cartes obsolètes sont incluses (sauf si carte supprimée)

## Tests automatisés
- Test unitaire : Algorithme de poids (calcul, mise à jour, sélection aléatoire)
- Test d'intégration : Simulation session révision → vérification fréquence zones selon erreurs
- Test E2E : Utilisateur fait erreur sur zone A → zone A revient plus souvent → utilisateur réussit zone A → poids diminue

## Documentation
- docs/revision-mode.md (voir Story #015)
- Voir specification.md section 3.2 (Algorithme de poids)

## Valeur utilisateur
Maximise l'efficacité de l'entraînement en se concentrant sur les zones problématiques.

## Dépendances
- Story #015 (Revision Mode with Validation)