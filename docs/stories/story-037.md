# Story #037 : Divs, assignation d'équipes et classement

## Description
En tant qu'archiviste, je veux créer des divs pour une ligue, y assigner des équipes eva et consulter le classement automatique afin d'organiser la phase régulière d'une compétition.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (sections 4, 8 étape 3).

## Critères d'acceptation
- [ ] L'archiviste peut créer des divs pour une ligue avec un nom personnalisable
- [ ] L'archiviste peut assigner une équipe eva à une div (une équipe = une seule div par ligue)
- [ ] L'archiviste peut retirer une équipe eva d'une div
- [ ] Consultation : tableau de classement par div avec équipe, parties jouées, victoires, nuls, défaites
- [ ] Le classement est calculé automatiquement à partir des résultats de BO saisis :
  - 1 point par BO gagné, 0 par BO perdu
  - 1 point par match gagné, 0 par match perdu
  - Les deux classements (BO et match) sont affichés séparément
- [ ] La phase par div correspond à un round-robin (chaque équipe rencontre toutes les autres)
- [ ] L'API REST expose les endpoints
- [ ] Les utilisateurs sans droits ne voient pas la fonctionnalité

## Tests automatisés
- Test unitaire : calcul de classement automatique (scoring BO et match)
- Test d'intégration : API divs et assignation
- Test E2E : création d'une ligue → divs → assignation équipes → classement après saisie de BO

## Documentation
- docs/archives-divisions.md (à créer)

## Valeur utilisateur
Organise les équipes en divisions avec un classement automatique, offrant une vue structurée de la saison régulière.

## Dépendances
- Story #035 (Gestion des ligues locales) — entité `Ligue` (périodes de compétition), pas `LigueLocale`
- Story #036 (Gestion des équipes eva)
