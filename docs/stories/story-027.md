# Story #027 : Copier/coller de formes dans l'éditeur

## Description
En tant qu'éditeur de carte, je veux pouvoir copier et coller une forme (polygone) dans l'éditeur afin de dupliquer rapidement des zones similaires.

## Origine
Feedback du 2026-05-17 : "Le Copier/coller d'une forme n'est plus accessible dans l'éditeur" — régression : la fonctionnalité existait et a été cassée par une modification récente.

## Critères d'acceptation
- [ ] **Priorité** : régression à corriger en priorité
- [ ] Un raccourci clavier (Ctrl+C / Cmd+C) permet de copier la forme sélectionnée
- [ ] Un raccourci clavier (Ctrl+V / Cmd+V) permet de coller la forme copiée à la position du curseur
- [ ] Le collage crée une nouvelle forme décalée de quelques pixels de l'originale pour éviter la superposition parfaite
- [ ] La forme collée appartient à la **même zone** que l'originale (pas de création de nouvelle zone)
- [ ] Les actions de copier/coller sont visibles dans le menu contextuel de la forme sélectionnée
- [ ] Le copier/coller fonctionne uniquement quand une forme est sélectionnée (état d'édition actif)
- [ ] Raccourci documenté dans la popup d'aide de l'éditeur (Story #023)

## Tests automatisés
- Test unitaire : Logique de duplication d'une forme (coordonnées, propriétés, décalage)
- Test d'intégration : API de création de zone après copier/coller → vérification des données persistées (même zone)
- Test E2E : Sélection d'une forme → Ctrl+C → Ctrl+V → vérification qu'une nouvelle forme apparaît dans la même zone avec un décalé à la position attendue

## Documentation
- Mise à jour de la popup d'aide de l'éditeur (story-023)
- Raccourci ajouté à la liste des commandes clavier

## Valeur utilisateur
Accélère considérablement la création de zones complexes où plusieurs formes similaires sont nécessaires (ex : bâtiments symétriques, zones adjacentes).

## Dépendances
- Story #009 (Zone Creation and Editing)
- Story #023 (Editor Help Popup)
