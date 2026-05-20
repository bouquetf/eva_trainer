# Story #029 : Indicateur d'étage sur les formes (Haut/Bas)

## Description
En tant qu'éditeur, je veux pouvoir définir un indicateur d'étage (Haut/Bas) sur chaque forme d'une zone via un clic droit, afin de préciser la position verticale (étage) de la forme sur la carte.

## Origine
Feedback du 2026-05-19 : "Ajouter un indicateur d'étage sur une zone (Haut/Bas). Click droit -> étage -> Haut/Bas/Effacer. Flèche vers le haut et trait au dessus pour Haut, Flèche vers le bas et trait dessous pour bas."

## Critères d'acceptation
- [ ] Un clic droit sur une forme dans l'éditeur affiche un menu contextuel avec une entrée "Étage"
- [ ] Le sous-menu "Étage" propose trois options : Haut, Bas, Effacer (si un étage est déjà défini)
- [ ] Les options Haut et Bas sont mutuellement exclusives sur une forme (une forme ne peut être que Haut, Bas, ou aucun)
- [ ] Sélectionner "Haut" ajoute un indicateur visuel : flèche vers le haut + trait au-dessus de la forme (pictogramme "à l’étage")
- [ ] Sélectionner "Bas" ajoute un indicateur visuel : flèche vers le bas + trait en dessous de la forme (pictogramme "au sol")
- [ ] Sélectionner "Effacer" supprime l'indicateur d'étage de la forme
- [ ] L'indicateur est stocké comme propriété de la forme et persiste entre les sessions
- [ ] L'indicateur est visible par tous les membres de l'équipe dans l'éditeur
- [ ] L'indicateur est également visible pendant la révision

## Tests automatisés
- Test unitaire : Validation de la mise à jour du champ étage sur une forme
- Test d'intégration : API met à jour l'indicateur d'étage d'une forme → vérification en base
- Test E2E : Clic droit sur forme → sélection "Haut" → indicateur visuel affiché → rechargement → indicateur toujours présent
- Test E2E : Sélection "Bas" sur une forme déjà en "Haut" → l'indicateur passe à "Bas" (exclusivité)
- Test E2E : Template public → l'indicateur d'étage est copié lors de la duplication

## Documentation
- docs/zone-editor.md (mise à jour)

## Valeur utilisateur
Permet aux équipes de distinguer visuellement les formes situées à différents étages sur une carte, améliorant la compréhension de la disposition verticale pendant l'édition et la révision.

## Dépendances
- Story #009 (Zone Creation and Editing with Polygons)
