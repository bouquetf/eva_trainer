# Story #024 : Visibilité des rôles dans l'interface éditeur

## Description
En tant que capitaine, je veux que le bouton d'accès à l'éditeur soit masqué pour les membres afin d'éviter la confusion et les tentatives d'accès non autorisées.

## Origine
Feedback du 2025-05-14 : "Ne pas afficher le bouton éditeur pour les membres"

## Critères d'acceptation
- [ ] Le bouton "Éditeur" (ou le lien vers l'éditeur de carte) n'est pas affiché pour les utilisateurs ayant uniquement le rôle Membre (sans rôle Éditeur, Capitaine ou Cocapitaine)
- [ ] Le bouton est visible pour les Éditeurs, Capitaines et Cocapitaines
- [ ] Règle d'affichage : le rôle le plus élevé détermine l'affichage (si un utilisateur est Membre ET Éditeur → le bouton est visible car le rôle Éditeur prévaut)
- [ ] La vérification se fait côté frontend (affichage conditionnel) et côté backend (API sécurisée)
- [ ] Un message informatif peut être affiché pour les membres : "Seuls les éditeurs peuvent modifier les zones"

## Tests automatisés
- Test unitaire : Logique d'affichage conditionnel selon le rôle (y compris cas multi-rôles)
- Test d'intégration : API d'édition retourne 403 pour les membres sans rôle éditeur
- Test E2E : Membre consulte la page d'une carte → bouton éditeur masqué → tente d'accéder à l'URL éditeur → 403

## Documentation
- docs/roles-permissions.md (mise à jour)

## Valeur utilisateur
Réduit la confusion pour les membres qui ne peuvent pas éditer, et clarifie l'interface en n'affichant que les actions disponibles.

## Dépendances
- Story #004 (Team Roles and Permissions)
- Story #009 (Zone Creation and Editing)
