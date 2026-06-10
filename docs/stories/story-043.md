# Story #043 : Notes personnelles des lecteurs sur les divs

## Description
En tant que lecteur, je veux ajouter une note personnelle avec un éditeur texte riche minimaliste sur une div afin de prendre des notes sur les matchs de la division.

## Origine
Feedback du 2026-06-10 : "Un joueur qui a accès aux divs peut ajouter des commentaires à l'aide d'un éditeur texte riche minimaliste. Le résultat est stocké en base de données en markdown."

## Critères d'acceptation
- [ ] Un lecteur peut ajouter une note sur une div via un éditeur texte riche minimaliste
- [ ] L'éditeur permet : gras, italique, listes, liens, titres
- [ ] Le contenu est converti et stocké en markdown
- [ ] Un seul commentaire par personne et par div (un nouveau remplace l'ancien)
- [ ] Seul l'auteur voit son commentaire (pas de visibilité partagée)
- [ ] L'auteur peut modifier ou supprimer son propre commentaire
- [ ] Pas de notification liée aux commentaires

## Tests automatisés
- Test unitaire : conversion riche → markdown
- Test d'intégration : API — création, modification, suppression d'une note sur une div
- Test E2E : lecteur ajoute une note → vérification que seul l'auteur la voit (un autre utilisateur ne la voit pas)

## Documentation
- docs/archives-notes.md (à créer)

## Valeur utilisateur
Permet aux lecteurs de prendre des notes personnelles sur les divisions sans impact sur les autres utilisateurs.

## Dépendances
- Story #037 (Divs, assignation d'équipes et classement)
- Story #034 (Rôles Archiviste et Lecteur pour les archives matchs)
