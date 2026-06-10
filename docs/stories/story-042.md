# Story #042 : Lien de replay stream sur une div

## Description
En tant qu'archiviste, je veux ajouter un lien vidéo de replay sur une div afin que les spectateurs puissent retrouver le stream de la compétition.

## Origine
Feedback du 2026-06-10 : "Ajouter à une div un lien vers une vidéo pour le replay du stream"

## Critères d'acceptation
- [ ] L'archiviste peut ajouter un lien YouTube ou Twitch sur l'édition d'une div (un seul lien par div)
- [ ] Le lien est affiché dans la vue détaillée de la div
- [ ] Un lien YouTube s'affiche avec un lecteur embedded intégré
- [ ] Un lien Twitch s'affiche sous forme de lien cliquable
- [ ] Un lien invalide (pas YouTube ni Twitch) est rejeté avec un message d'erreur
- [ ] L'archiviste peut modifier ou supprimer le lien existant

## Tests automatisés
- Test unitaire : validation des formats d'URL (YouTube, Twitch, rejet des autres)
- Test d'intégration : API — ajout, modification, suppression du lien sur une div
- Test E2E : archiviste ajoute un lien YouTube → vérification du lecteur embedded dans la div

## Documentation
- Mise à jour de docs/archives-divisions.md

## Valeur utilisateur
Permet aux équipes et spectateurs de retrouver et visionner le replay de la compétition depuis la div concernée.

## Dépendances
- Story #037 (Divs, assignation d'équipes et classement)
