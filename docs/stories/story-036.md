# Story #036 : Gestion des équipes eva

## Description
En tant qu'archiviste, je veux créer, modifier et consulter les équipes "eva" et les associer optionnellement à une équipe eva-trainer afin de référencer les équipes réelles participantes.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (sections 2, 8 étape 2).

## Critères d'acceptation
- [ ] L'archiviste peut créer une équipe eva avec : nom (obligatoire), tag (optionnel), logo_url (optionnel)
- [ ] L'archiviste peut associer une équipe eva à 0 ou 1 équipe eva-trainer (relation 1:1 optionnelle)
- [ ] L'archiviste peut modifier les informations d'une équipe eva
- [ ] L'archiviste peut supprimer une équipe eva (bloqué si ligues/matchs associés)
- [ ] Consultation : fiche détaillée d'une équipe eva avec ses informations
- [ ] L'API REST expose les endpoints CRUD
- [ ] Les utilisateurs sans droits ne voient pas la fonctionnalité

## Tests automatisés
- Test unitaire : CRUD équipe eva + contrainte association 1:1
- Test d'intégration : API équipes eva
- Test E2E : création d'une équipe eva → association → consultation

## Documentation
- docs/archives-eva-teams.md (à créer)

## Valeur utilisateur
Permet de référencer les équipes réelles du jeu Eva dans le système d'archives, indépendamment ou en lien avec les équipes du site.

## Dépendances
- Story #034 (Rôles Archiviste et Lecteur)
