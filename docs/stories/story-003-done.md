# Story #003 : Team Creation and Management

## Description
En tant qu'utilisateur, je veux créer une équipe afin de pouvoir collaborer avec d'autres membres sur l'entraînement Eva.

## Critères d'acceptation
- [ ] Un utilisateur peut créer une équipe (devient Capitaine automatiquement)
- [ ] Le nom d'équipe est unique, max 250 caractères
- [ ] Les caractères autorisés sont validés (éviter injections SQL et caractères dangereux)
- [ ] Le Capitaine peut modifier le nom et les informations de l'équipe
- [ ] Le Capitaine peut supprimer l'équipe (confirmation requise)
- [ ] Lors de la suppression, les zones privées sont supprimées et les zones publiques deviennent des templates
- [ ] Si le Capitaine quitte sans transférer son rôle, l'équipe est supprimée automatiquement

## Tests automatisés
- Test unitaire : Validation des contraintes sur le nom d'équipe (longueur, unicité, caractères)
- Test d'intégration : Création d'équipe → vérification du rôle Capitaine attribué
- Test E2E : Création d'équipe → modification → suppression avec confirmation

## Documentation
- docs/teams.md (à créer)
- Voir specification.md section 4.3

## Valeur utilisateur
Permet aux utilisateurs de former des équipes pour organiser l'entraînement et partager des zones personnalisées.

## Dépendances
- Story #002 (User Authentication via Discord)