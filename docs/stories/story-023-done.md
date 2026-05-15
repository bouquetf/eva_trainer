# Story #023 : Popup d'aide dans l'éditeur

## Description
En tant qu'éditeur, je veux une popup d'aide affichant la liste des commandes et raccourcis clavier afin de connaître les fonctionnalités disponibles dans l'éditeur.

## Origine
Feedback du 2025-05-14 : "Popup d'aide sur l'éditeur : affiche la liste de commandes et les raccourcis clavier associés"

## Critères d'acceptation
- [ ] Un bouton "Aide" ou "?" dans l'éditeur ouvre une popup modale
- [ ] La popup liste les commandes disponibles et leurs raccourcis clavier (basée sur les raccourcis réellement implémentés dans le code)
- [ ] Les raccourcis documentés dans la popup :

| Commande | Raccourci | Contexte |
|----------|-----------|----------|
| Sauvegarder la zone | Ctrl+S / Cmd+S | Zone en cours d'édition |
| Annuler (undo) | Ctrl+Z / Cmd+Z | Maximum 50 états |
| Supprimer polygone(s) sélectionné(s) | Delete / Suppr | Au moins 1 polygone sélectionné |
| Confirmer le nom de la zone | Entrée | Champ de saisie du nom |
| Valider le renommage | Entrée | Renommage d'une zone |
| Annuler le renommage | Échap | Champ de renommage |
| Fermer la boîte de dialogue | Échap | Dialogue de nom de zone |
| Ajouter un point | Clic gauche sur le canvas | Mode dessin de polygone |
| Supprimer le dernier point | Ctrl+Click sur le canvas | Mode dessin de polygone |
| Sélection multiple de polygones | Ctrl+Click / Cmd+Click | Sur un polygone |
| Supprimer un point | Ctrl+Click / Cmd+Click | Sur une poignée de point |
| Ajouter/insérer un point | Double-clic | Sur le canvas ou un segment |
| Annuler dernière action (menu contextuel) | Clic droit | Mode dessin |
| Terminer la forme (menu contextuel) | Clic droit | Mode dessin |

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
