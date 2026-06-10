# Story #041 : Masquer l'import de ligue pour les lecteurs

## Description
En tant que lecteur, je ne dois pas voir le formulaire "Ajouter une ligue" dans la vue détaillée d'une ligue locale, car cette fonctionnalité est réservée aux archivistes.

## Origine
Feedback du 2026-06-07 : "Lecteur ne doit pas voir : ligues locales -> importer"

## Critères d'acceptation
- [ ] Le formulaire "Ajouter une ligue" n'est pas affiché pour les utilisateurs ayant le rôle "lecteur"
- [ ] Le formulaire "Ajouter une ligue" reste visible pour les archivistes et admins

## Tests automatisés
- Test E2E : connexion en tant que lecteur → navigation dans Archives → clic sur une ligue locale → vérification que "Ajouter une ligue" est absent

## Documentation
- Aucune

## Valeur utilisateur
Empêche les lecteurs de voir une action qu'ils ne peuvent pas effectuer, clarifiant l'interface.

## Dépendances
- Aucune
