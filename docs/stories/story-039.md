# Story #039 : Statistiques cartes

## Description
En tant que lecteur, je veux consulter les statistiques globales par carte (victoires, défaites, winrate, scores moyen et médian) afin d'analyser les performances sur chaque carte.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (section 6, 8 étape 5).

## Critères d'acceptation
- [ ] Consultation : liste des cartes avec pour chaque carte :
  - Nombre de victoires
  - Nombre de défaites
  - Nombre moyen de victoires = winrate (victoires / (victoires + défaites))
  - Score moyen par équipe sur cette carte
  - Score médian par équipe sur cette carte
- [ ] Tri possible sur chacune des colonnes (ascendant et descendant)
- [ ] Les données sont calculées à partir de tous les matchs enregistrés
- [ ] L'API REST expose les endpoints
- [ ] Les utilisateurs sans droits ne voient pas la fonctionnalité

## Tests automatisés
- Test unitaire : calcul des statistiques (winrate, moyenne, médiane)
- Test d'intégration : API stats cartes
- Test E2E : consultation de la page stats → vérification des valeurs → tri

## Documentation
- docs/archives-card-stats.md (à créer)

## Valeur utilisateur
Offre une vue analytique des performances par carte, utile pour préparer les stratégies de picks/bans.

## Dépendances
- Story #038 (BO, matchs, bans et picks)
