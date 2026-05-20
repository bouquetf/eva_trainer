# Story #031 : Réorganisation de l'interface d'équipe

## Description
En tant qu'utilisateur, je veux une interface d'équipe réorganisée avec des sections repliables afin de naviguer plus facilement dans les paramètres.

## Origine
Feedback du 2026-05-19 : "Dans l'interface d'équipe : le paragraphe 'gestion des rôles' est inutile, supprime le bloc. Demandes d'adhésions doit être dans le bloc 'Invitations'. La zone membres doit être sous 'cartes'. Possibilité de réduire sous forme d'accordéon les blocs : Membres, Invitations, Zone de danger."

## Critères d'acceptation
- [ ] Le bloc "Gestion des rôles" est supprimé de l'interface d'équipe
- [ ] Le bloc "Demandes d'adhésion" est déplacé à l'intérieur du bloc "Invitations"
- [ ] Le bloc "Membres" est repositionné après le bloc "Cartes" (simple changement d'ordre, pas d'imbrication)
- [ ] L'ordre final des sections est : Informations, Cartes, Membres, Invitations (contenant Demandes d'adhésion), Quitter l'équipe, Zone de danger
- [ ] Les blocs "Membres" et "Zone de danger" sont repliables via un accordéon, fermés par défaut
- [ ] Le bloc "Invitations" est repliable via un accordéon, ouvert par défaut s'il y a des demandes en attente

## Tests automatisés
- Test E2E : Vérifier que le bloc "Gestion des rôles" n'est plus affiché
- Test E2E : Vérifier que "Demandes d'adhésion" apparaît dans le bloc "Invitations"
- Test E2E : Vérifier que "Membres" apparaît après "Cartes"
- Test E2E : Cliquer sur l'en-tête d'un bloc accordéon → le contenu se replie/affiche

## Documentation
- docs/team-management.md (mise à jour si nécessaire)

## Valeur utilisateur
Simplifie l'interface d'équipe en supprimant les éléments redondants et en regroupant logiquement les sections, rendant la navigation plus intuitive.

## Dépendances
- Story #003 (Team Creation)
- Story #004 (Team Roles and Permissions)
- Story #006 (Team Invitation and Join System)
