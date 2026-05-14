# Story #023 : Popup d'aide dans l'éditeur

## Description
En tant qu'éditeur, je veux une popup d'aide affichant la liste des commandes et raccourcis clavier afin de connaître les fonctionnalités disponibles dans l'éditeur.

## Origine
Feedback du 2025-05-14 : "Popup d'aide sur l'éditeur : affiche la liste de commandes et les raccourcis clavier associés"

## Critères d'acceptation
- [ ] Un bouton "Aide" ou "?" dans l'éditeur ouvre une popup modale
- [ ] La popup liste les commandes disponibles (dessiner un polygone, ajouter un point, supprimer, etc.)
- [ ] Chaque commande est associée à son raccourci clavier
- [ ] La popup peut être fermée avec Échap ou en cliquant en dehors
- [ ] La popup peut être déclenchée par un raccourci clavier (ex: ? ou F1)

## Tests automatisés
- Test unitaire : Logique d'ouverture/fermeture de la popup
- Test d'intégration : Vérification du contenu de la popup (commandes listées)
- Test E2E : Éditeur ouvre éditeur → clique Aide → popup affichée → commandes visibles → ferme la popup

## Documentation
- docs/editor-help.md (à créer)

## Valeur utilisateur
Améliore la prise en main de l'éditeur en rendant les commandes et raccourcis clavier facilement accessibles.

## Dépendances
- Story #009 (Zone Creation and Editing)
