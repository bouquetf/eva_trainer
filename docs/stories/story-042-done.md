# Story #042 : Lien de replay stream sur une div

## Description
En tant qu'archiviste, je veux ajouter un lien vidéo de replay sur une div afin que les spectateurs puissent retrouver le stream de la compétition.

## Origine
Feedback du 2026-06-10 : "Ajouter à une div un lien vers une vidéo pour le replay du stream"

## Critères d'acceptation
- [x] L'archiviste peut ajouter un lien YouTube ou Twitch sur l'édition d'une div (un seul lien par div)
- [x] Le lien est affiché dans la vue détaillée de la div
- [x] Un lien YouTube s'affiche avec un lecteur embedded intégré
- [x] Un lien Twitch s'affiche sous forme de lien cliquable
- [x] Un lien invalide (pas YouTube ni Twitch) est rejeté avec un message d'erreur
- [x] L'archiviste peut modifier ou supprimer le lien existant

## Tests automatisés
- [x] Test unitaire : validation des formats d'URL (YouTube, Twitch, rejet des autres)
- [x] Test d'intégration : API — ajout, modification, suppression du lien sur une div
- [x] Test E2E : archiviste ajoute un lien YouTube → vérification du lecteur embedded dans la div

## Documentation
- [x] Mise à jour de docs/archives-divisions.md

## Valeur utilisateur
Permet aux équipes et spectateurs de retrouver et visionner le replay de la compétition depuis la div concernée.

## Dépendances
- Story #037 (Divs, assignation d'équipes et classement)
