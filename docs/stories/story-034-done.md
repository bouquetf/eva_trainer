# Story #034 : Rôles Archiviste et Lecteur pour les archives matchs

## Description
En tant qu'administrateur, je veux pouvoir attribuer les rôles "archiviste" (global) et "lecteur" (par équipe) afin de gérer les droits d'accès aux archives des matchs, avec masquage des fonctionnalités pour les utilisateurs non autorisés.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (section 1).

## Critères d'acceptation
- [x] Le rôle global "archiviste" est ajouté au système de rôles (attribuable par l'admin uniquement)
- [x] Le rôle "lecteur" est ajouté, attribuable par l'admin à une équipe eva-trainer
- [x] Un archiviste a automatiquement le rôle lecteur
- [x] Les utilisateurs sans rôle archiviste ni rôle lecteur ne voient pas les fonctionnalités d'archives dans l'interface
- [x] Les admins du site ont automatiquement accès à toutes les archives
- [x] Les endpoints API des archives vérifient les permissions (archiviste pour écriture, lecteur ou archiviste ou admin pour lecture)

## Tests automatisés
- Test unitaire : vérification des permissions archiviste/lecteur dans le système de rôles
- Test d'intégration : API refuse les accès non autorisés aux endpoints archives
- Test E2E : un utilisateur sans droits ne voit pas les entrées d'archives dans l'UI

## Documentation
- docs/roles-archives.md (à créer)

## Valeur utilisateur
Permet un contrôle d'accès granulaire aux archives : seuls les archivistes et les équipes avec le rôle lecteur peuvent accéder aux données.

## Dépendances
- Story #004 (Team Roles and Permissions)
