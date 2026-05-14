# Story #018 : Refinements du verrouillage d'édition

## Description
En tant qu'éditeur, je veux que le verrouillage soit plus explicite et réactif afin d'éviter les conflits de modification et de savoir qui édite la carte.

## Origine
Feedbacks du 2025-05-14 :
- "Afficher un message d'erreur lorsque l'on tente d'ouvrir l'éditeur et que quelqu'un d'autre a le lock (afficher le nom du joueur qui a le lock)"
- "Rendre le lock lorsque l'on quitte l'éditeur"
- "Afficher un message d'erreur lorsque l'on reprend le focus sur l'éditeur et qu'un autre joueur prend le lock"

## Critères d'acceptation
- [ ] Quand un éditeur tente d'ouvrir l'éditeur et qu'un autre utilisateur a le lock : message d'erreur affichant le nom du joueur qui détient le lock (pas de moyen de contact pour l'instant)
- [ ] Quand l'éditeur quitte l'éditeur (navigation, fermeture d'onglet, fermeture du navigateur) : le lock est rendu immédiatement
- [ ] Quand l'éditeur reprend le focus sur l'éditeur (retour d'arrière-plan, réactivation de l'onglet) : vérification que le lock est toujours détenu
- [ ] Si le lock a été pris par un autre joueur pendant l'absence : message d'erreur avec le nom du nouveau détenteur
- [ ] Conservation du timeout de 15 minutes comme filet de sécurité en cas de fermeture brutale

## Tests automatisés
- Test unitaire : Logique de libération du lock à la fermeture (beforeunload)
- Test d'intégration : Simulation de reprise de focus → vérification du lock
- Test E2E : Éditeur A ouvre éditeur → Éditeur B tente → message avec nom de A → A quitte → B peut ouvrir

## Documentation
- docs/map-locking.md (mise à jour)

## Valeur utilisateur
Évite la confusion lors des conflits d'édition en affichant clairement qui édite, et libère le lock plus rapidement quand l'éditeur quitte.

## Dépendances
- Story #010 (Map Locking Mechanism for Concurrent Editing)
