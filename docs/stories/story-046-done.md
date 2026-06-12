# Story #046 : Améliorations BO/match (regroupement, création, hors compétition)

## Description
En tant qu'archiviste et lecteur, je veux que les BO soient regroupés par division, que l'ajout d'un pick ne crée pas un match automatiquement, et pouvoir marquer un match "hors compétition" afin d'améliorer la navigation, la saisie, et le suivi des matchs.

## Origine
Feedbacks du 2026-06-12 :
- "Dans archives -> BO, il faut les regrouper par division"
- "Bug : dans les détails d'un BO, ajouter un pick ne doit pas créer un match automatiquement. Lorsque l'on ajoute un match, on l'ajoute en prenant automatiquement le pick suivant."
- "Il doit être possible dans les BO de marquer un match 'hors compétition'. Le résultat est enregistré dans les stats mais ne compte pas pour le résultat du BO."

## Critères d'acceptation

### Regroupement des BO par division
- [x] La liste des BO dans les archives est regroupée/affichée par division
- [x] Chaque division affiche ses BOs avec les équipes et scores

### Création de match depuis les picks
- [x] Ajouter un pick dans un BO ne crée pas de match automatiquement
- [x] Lorsque l'archiviste ajoute un match, celui-ci prend automatiquement le pick suivant (carte et équipe) comme valeur par défaut
- [x] L'archiviste peut modifier la carte et les scores du match ainsi créé

### Match "hors compétition"
- [x] L'archiviste peut cocher/décocher "hors compétition" sur un match
- [x] Un match "hors compétition" est compté dans les statistiques (stats cartes, winrate, scores)
- [x] Un match "hors compétition" n'est PAS compté dans le résultat du BO (victoires/défaites du BO)
- [x] L'API expose le champ `hors_competition` (boolean) sur les endpoints match

## Tests automatisés
- [x] Test unitaire : création de match avec auto-pick du pick suivant (`TestComputeBOWinner` avec `HorsCompetition`)
- [x] Test unitaire : match hors compétition compté dans les stats mais pas dans le résultat BO (3 cas dans `TestComputeBOWinner`)
- [x] Test d'intégration : API — ajout pick sans match, ajout match avec auto-pick (`TestArchiveHandler_CreateBO_PicksWithoutMatchs`)
- [x] Test d'intégration : API — toggle hors compétition sur un match (`TestArchiveHandler_ToggleMatchHorsCompetition`)
- [x] Test d'intégration : API — BO avec match hors compétition (`TestArchiveHandler_CreateBO_WithHorsCompetitionMatch`)
- [x] Test d'intégration : API — liste BO avec div_id/div_name (`TestArchiveHandler_ListBOs_WithDiv`)

## Documentation
- docs/archives-bo-matches.md (mise à jour)

## Valeur utilisateur
Améliore la saisie des résultats (pas de matchs automatiques intempestifs, auto-pick facilitant la création) et permet de gérer des matchs hors compétition (exhibition, warm-up) sans fausser le classement BO.

## Dépendances
- Story #037 (Divs, assignation et classement) — pour le regroupement par division
- Story #038 (BO, matchs, bans et picks)
