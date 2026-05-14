# Story #006 : Team Invitation and Join System

## Description
En tant qu'utilisateur, je veux inviter des membres ou rejoindre une équipe afin de collaborer avec d'autres personnes.

## Critères d'acceptation
- [ ] Le Capitaine/Cocapitaine peut inviter un utilisateur par son pseudo (erreur 404 si inexistant)
- [ ] Le Capitaine/Cocapitaine peut générer un lien d'invitation (validité 1 jour)
- [ ] L'utilisation d'un lien expiré affiche un message d'erreur spécifique
- [ ] Un utilisateur peut demander à rejoindre une équipe via la liste des équipes
- [ ] Si un utilisateur reçoit une invitation pour une équipe dont il est déjà membre : affichage du message "Vous êtes déjà membre de cette équipe"
- [ ] Les demandes en attente sont affichées aux capitaines/cocapitaines
- [ ] Le Capitaine/Cocapitaine peut accepter ou refuser une demande
- [ ] L'utilisateur peut annuler sa demande d'adhésion et voir son statut

## Tests automatisés
- Test unitaire : Validation de la génération du lien d'invitation et expiration
- Test d'intégration : Invitation par pseudo → acceptation → vérification du rôle Membre attribué
- Test E2E : Utilisateur demande à rejoindre → Capitaine accepte → Utilisateur voit l'équipe dans sa liste

## Documentation
- docs/invitation-system.md (à créer)
- Voir specification.md section 4.3

## Valeur utilisateur
Facilite l'expansion des équipes et permet aux utilisateurs de rejoindre des équipes existantes.

## Dépendances
- Story #003 (Team Creation and Management)
- Story #004 (Team Roles and Permissions)