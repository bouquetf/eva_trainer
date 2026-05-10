# Système de templates

## Définition

Un **template** est une copie publique des zones d'une carte, créée par le
Capitaine ou le Cocapitaine d'une équipe. Il permet à d'autres équipes de
dupliquer ces zones pour leur propre usage.

## Publication d'une carte

1. Ouvrez l'éditeur de carte.
2. Cliquez sur le bouton **"Rendre publique"** dans la barre d'outils.
3. Confirmez l'action (attention : **irréversible**).
4. Toutes les zones de la carte deviennent publiques sous forme de template.

⚠️ **Important** : Une fois qu'une carte est rendue publique, elle ne peut
plus être rendue privée. C'est une contrainte définitive pour garantir la
stabilité des templates utilisés par d'autres équipes.

## Mise à jour automatique

Après publication, votre équipe continue de modifier ses zones normalement.
Le template public est automatiquement mis à jour à chaque modification :

- ✏️ Création de zone → ajoutée au template
- 📝 Modification de polygone → mis à jour dans le template
- 🔄 Renommage → reflété dans le template
- 🗑️ Suppression de zone → retirée du template

Les nouvelles zones que vous créez sont automatiquement incluses dans le
template si elles sont visibles.

## Duplication par d'autres équipes

Une fois le template créé, d'autres équipes peuvent le dupliquer pour
créer leurs propres zones indépendantes. Voir `docs/duplication.md`
(Story #014) pour plus de détails.

## Administration

Les administrateurs peuvent supprimer tout template via l'interface
d'administration.
