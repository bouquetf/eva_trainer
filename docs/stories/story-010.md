# Story #010 : Map Locking Mechanism for Concurrent Editing

## Description
En tant qu'éditeur, je veux être averti si une carte est déjà en cours d'édition afin d'éviter les conflits de modification.

## Critères d'acceptation
- [ ] Verrouillage au niveau de la CARTE entière, mais uniquement au sein d'une équipe
- [ ] Un seul éditeur peut modifier une carte à la fois
- [ ] Le verrou devient inactif après 15 minutes d'inactivité
- [ ] Si l'éditeur tente de modifier une carte verrouillée par un autre, il reçoit un message d'erreur
- [ ] En cas de fermeture brutale du navigateur, le timeout d'inactivité (15 min) libère le verrou
- [ ] L'éditeur doit tenter d'obtenir à nouveau le verrou s'il veut modifier après erreur

## Tests automatisés
- Test unitaire : Logique de verrouillage/déverrouillage avec timeout
- Test d'intégration : Deux éditeurs tentent de modifier la même carte → seul le premier réussit
- Test E2E : Éditeur A verrouille carte → Éditeur B tente modification → message d'erreur → Éditeur A inactif 15 min → Éditeur B peut modifier

## Documentation
- docs/map-locking.md (à créer)
- Voir specification.md section 2 (Verrouillage)

## Valeur utilisateur
Évite les conflits de modification simultanée et assure l'intégrité des données.

## Dépendances
- Story #009 (Zone Creation and Editing)