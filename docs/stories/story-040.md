# Story #040 : Historique des matchs entre 2 équipes

## Description
En tant que lecteur, je veux consulter les statistiques de cartes filtrées par l'historique des matchs entre 2 équipes spécifiques afin d'analyser leurs confrontations directes.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (section 6, 8 étape 6).

## Critères d'acceptation
- [ ] Consultation : sélection de 2 équipes eva
- [ ] Affichage des statistiques cartes limitées aux matchs où ces 2 équipes se sont affrontées :
  - Nombre de victoires par équipe sur chaque carte
  - Nombre de défaites par équipe sur chaque carte
  - Winrate par équipe sur chaque carte
  - Score moyen et médian par équipe sur chaque carte
- [ ] Tri possible sur chaque colonne
- [ ] Pas de filtres supplémentaires (date, ligue, etc.)
- [ ] L'API REST expose les endpoints
- [ ] Les utilisateurs sans droits ne voient pas la fonctionnalité

## Tests automatisés
- Test unitaire : calcul des stats filtrées par paire d'équipes
- Test d'intégration : API historique entre 2 équipes
- Test E2E : sélection de 2 équipes → consultation de l'historique → tri

## Documentation
- docs/archives-head-to-head.md (à créer)

## Valeur utilisateur
Permet d'analyser les confrontations directes entre 2 équipes, utile pour préparer les matchs à venir.

## Dépendances
- Story #039 (Statistiques cartes)
