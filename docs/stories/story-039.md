# Story #039 : Statistiques cartes

## Description
En tant que lecteur, je veux consulter les statistiques globales par carte (victoires, défaites, winrate, scores moyen et médian) afin d'analyser les performances sur chaque carte.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (section 6, 8 étape 5).

## Critères d'acceptation
- [x] Consultation : liste des cartes par équipe, avec pour chaque carte :
  - Nombre de victoires
  - Nombre de défaites
  - Nombre moyen de victoires = winrate (victoires / (victoires + défaites))
  - Score moyen et médian sur cette carte
  - Score moyen par équipe adverse sur cette carte
  - Score médian par équipe adverse sur cette carte
- [x] Tri possible sur chacune des colonnes (ascendant et descendant)
- [x] Les données sont calculées à partir de tous les matchs enregistrés (toutes les cartes sont affichées, même sans match)
- [x] L'API REST expose les endpoints
- [x] Les utilisateurs sans droits ne voient pas la fonctionnalité

## Tests automatisés
- [x] Test unitaire : calcul des statistiques (winrate, moyenne, médiane)
- [x] Test d'intégration : API stats cartes
- [x] Test E2E : consultation de la page stats → vérification des valeurs → tri

## Documentation
- [x] docs/archives-card-stats.md

## Valeur utilisateur
Offre une vue analytique des performances par carte, utile pour préparer les stratégies de picks/bans.

## Dépendances
- Story #038 (BO, matchs, bans et picks)
