# Story #044 : Suppression de membre et de lien d'invitation

## Description
En tant que capitaine ou cocapitaine d'équipe, je veux pouvoir supprimer un membre de l'équipe et supprimer un lien d'invitation afin de gérer les effectifs de mon équipe.

## Origine
Feedback du 2026-06-10 : "Ajouter la possibilité pour un capitaine ou cocapitaine d'équipe de supprimer un membre et de supprimer un lien d'invitation"

## Critères d'acceptation
- [ ] Le capitaine/cocapitaine peut supprimer un membre de l'équipe (confirmation requise)
- [ ] La suppression d'un membre retire ses accès à l'équipe et ses zones privées
- [ ] Le capitaine/cocapitaine peut supprimer un lien d'invitation, le rendant immédiatement invalide
- [ ] Un utilisateur ne peut pas supprimer le capitaine (sauf transfert préalable)
- [ ] Un capitaine/cocapitaine ne peut pas se supprimer lui-même

## Tests automatisés
- Test unitaire : validation des règles de suppression (pas d'auto-suppression, pas de suppression du capitaine)
- Test d'intégration : API — suppression d'un membre → vérification des accès révoqués
- Test E2E : capitaine supprime un membre → le membre ne voit plus l'équipe dans sa liste
- Test E2E : capitaine supprime un lien d'invitation → le lien n'est plus utilisable

## Documentation
- Mise à jour de docs/teams.md

## Valeur utilisateur
Donne aux capitaines et cocapitaines le contrôle nécessaire pour gérer les effectifs de leur équipe.

## Dépendances
- Story #003 (Team Creation and Management)
- Story #006 (Team Invitation and Join System)
