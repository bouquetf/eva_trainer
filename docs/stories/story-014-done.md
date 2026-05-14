# Story #014 : Template Duplication by Other Teams

## Description
En tant qu'utilisateur, je veux dupliquer un template public afin d'utiliser les zones d'une autre équipe comme base pour mon équipe.

## Critères d'acceptation
- [ ] Une équipe peut dupliquer un template public d'une autre équipe
- [ ] La nouvelle équipe devient propriétaire indépendante de leur duplication
- [ ] La duplication copie les polygones + noms de la carte publique
- [ ] Les modifications sur la copie n'affectent pas l'original
- [ ] Une équipe ne peut pas voir qui a dupliqué son template (information non gardée)
- [ ] Il n'existe pas de template à partir de zone privée

## Tests automatisés
- Test unitaire : Logique de duplication (copie indépendante)
- Test d'intégration : Duplication template → vérification copie indépendante → modification copie → original inchangé
- Test E2E : Équipe A publie carte → Équipe B duplique → Équipe B modifie sa copie → Équipe A voit son template inchangé

## Documentation
- docs/templates.md (voir Story #011)
- Voir specification.md section 4.5

## Valeur utilisateur
Permet aux équipes de gagner du temps en utilisant des bases de zones déjà définies par d'autres.

## Dépendances
- Story #011 (Zone Publishing and Template System)