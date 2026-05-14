# Story #009 : Zone Creation and Editing with Polygons

## Description
En tant qu'éditeur, je veux créer et modifier des zones polygonales sur une carte afin de définir les zones d'entraînement pour mon équipe.

## Critères d'acceptation
- [ ] Une zone est constituée d'une ou plusieurs formes polygonales (React-Konva)
- [ ] Pour une équipe donnée, une zone ne peut avoir qu'un seul nom par carte
- [ ] Le nom de zone a une longueur maximale de 250 caractères
- [ ] Ajout d'une forme : sauvegardé immédiatement si une forme existait déjà
- [ ] Suppression d'une forme : si plus de forme, la zone est supprimée
- [ ] Ajouter/supprimer un point à une forme ne sauvegarde que si la forme existe déjà ou est complète
- [ ] L'éditeur peut nommer la zone lors de la création dans l'éditeur
- [ ] Le verrouillage est strictement par équipe (indépendant) : si l'équipe A verrouille la carte, l'équipe B peut toujours modifier SES zones sur cette carte

## Tests automatisés
- Test unitaire : Validation des polygones (formes, points) et contraintes sur les noms
- Test d'intégration : Création zone avec polygones → sauvegarde en base → récupération
- Test E2E : Éditeur dessine polygone → nomme zone → sauvegarde automatique → zone visible sur la carte

## Documentation
- docs/zone-editor.md (à créer)
- Voir specification.md section 2

## Valeur utilisateur
Permet aux équipes de définir visuellement les zones d'entraînement sur les cartes du jeu Eva.

## Dépendances
- Story #005 (Map Management)
- Story #004 (Team Roles and Permissions)