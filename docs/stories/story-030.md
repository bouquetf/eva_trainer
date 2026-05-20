# Story #030 : Tests E2E pour le rejoindre d'équipe et l'acceptation de demandes

## Description
En tant que développeur, je veux des tests E2E robustes couvrant le parcours de demande de rejoindre une équipe et d'acceptation/refus par un capitaine, afin de garantir la fiabilité de ces fonctionnalités.

## Origine
Feedback du 2026-05-19 : "Il semble y avoir des bugs lorsqu'un utilisateur tente de rejoindre une équipe et lorsqu'un admin tente d'accepter une demande. Faudrait retester ces fonctionnalités en e2e."
Story #006 (Team Invitation and Join System) — fonctionnalité existante à sécuriser.

## Scénarios E2E

### Parcours : Demande de rejoindre une équipe
- [ ] Un utilisateur non membre consulte la liste des équipes → clique sur "Rejoindre" sur une équipe → la demande est envoyée → le statut "Demande en attente" s'affiche pour cet utilisateur
- [ ] Un utilisateur annule sa demande en attente → la demande disparaît → le capitaine ne voit plus la demande
- [ ] Un utilisateur tente de rejoindre une équipe déjà rejointe (double demande) → message d'erreur approprié

### Parcours : Acceptation / Refus par le capitaine
- [ ] Le capitaine voit la demande entrante dans le tableau dédié → clique "Accepter" → l'utilisateur devient membre → l'utilisateur voit l'équipe dans sa liste d'équipes
- [ ] Le capitaine clique "Refuser" → la demande est rejetée → l'utilisateur voit que sa demande a été refusée
- [ ] Le co-capitaine peut aussi accepter/refuser une demande

### Parcours : Invitation par lien
- [ ] Le capitaine génère un lien d'invitation → l'utilisateur clique sur le lien → rejoint l'équipe
- [ ] L'utilisateur clique sur un lien d'invitation expiré → message d'erreur "lien expiré"

### Parcours : Invitation par pseudo
- [ ] Le capitaine invite un utilisateur par son pseudo → l'utilisateur reçoit une notification → rejoint l'équipe
- [ ] Le capitaine invite un pseudo inexistant → erreur 404

## Questions pour le relecteur
- Faut-il ajouter un test pour le cas où le capitaine quitte l'équipe sans transférer le rôle (suppression de l'équipe) en rapport avec des demandes en attente ?

non

- Faut-il tester le cas d'un admin qui accepte une demande à la place du capitaine ?

non

- Faut-il tester la limite de 8 membres officiels en lien avec les demandes d'adhésion ?

non, une équipe peut contenir autant de membres qu’on le souhaite

- Faut-il tester le cas d'un utilisateur qui tente de rejoindre une équipe alors qu'il est déjà membre d'une autre équipe ?

non, un utilisateur peut appartenir à plusieurs équipes

## Documentation
- docs/team-management.md (mise à jour si nécessaire)

## Valeur utilisateur
Garantit la fiabilité des fonctionnalités de gestion d'équipe, qui sont critiques pour l'organisation des utilisateurs.

## Dépendances
- Story #006 (Team Invitation and Join System)
