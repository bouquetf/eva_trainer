# Story #038 : BO, matchs, bans et picks

## Description
En tant qu'archiviste, je veux saisir les résultats des BO (Best Of) entre 2 équipes avec les bans, picks et scores de matchs afin d'enregistrer le déroulement complet des rencontres. Les lecteurs peuvent consulter ces données.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (sections 5, 8 étape 4).

## Critères d'acceptation
- [ ] L'archiviste peut créer un BO entre 2 équipes avec :
  - Format (BO3 ou BO5)
  - Bans : sélection d'une équipe + carte bannie, numérotés par équipe
  - Picks : sélection d'une carte, équipe facultative, numérotés
  - Matchs : carte, score équipe 1 (0-100), score équipe 2 (0-100), commentaires optionnels
- [ ] L'ordre des matchs dans le BO correspond à l'ordre des picks
- [ ] Le vainqueur de chaque match est l'équipe avec le plus grand score
- [ ] Le vainqueur du BO est l'équipe avec le plus de victoires de match
- [ ] L'archiviste peut modifier et supprimer un BO après création
- [ ] Pas de contrainte d'organisation sur bans/picks (déclaratif)
- [ ] Consultation : liste des BOs par ligue avec équipes, score, détails (bans, picks, matchs)
- [ ] L'API REST expose les endpoints
- [ ] Les utilisateurs sans droits ne voient pas la fonctionnalité

## Tests automatisés
- Test unitaire : création BO avec bans/picks/matchs, calcul vainqueur BO et match
- Test d'intégration : API BO (CRUD complet)
- Test E2E : saisie d'un BO complet (bans → picks → matchs) → consultation → modification

## Documentation
- docs/archives-bo-matches.md (à créer)

## Valeur utilisateur
Permet l'enregistrement et la consultation de l'historique complet des rencontres, incluant bans, picks et scores détaillés.

## Dépendances
- Story #035 (Gestion des ligues locales) — entité `Ligue` (périodes de compétition), pas `LigueLocale`
- Story #036 (Gestion des équipes eva)
- Story #037 (Divs, assignation et classement)
