# Story #004 : Team Roles and Permissions

## Description
En tant qu'administrateur d'équipe, je veux gérer les rôles des membres afin de contrôler qui peut modifier les zones et gérer l'équipe.

## Critères d'acceptation
- [ ] Le Capitaine peut assigner les rôles : Membre, Éditeur, Cocapitaine
- [ ] Un seul Cocapitaine par équipe maximum
- [ ] Le Cocapitaine a les mêmes droits que le Capitaine sauf la suppression d'équipe
- [ ] Le Capitaine peut transférer son rôle à un autre membre (devient Cocapitaine)
- [ ] Un admin ne peut pas être rétrogradé
- [ ] Chaque appel API vérifie les rôles et renvoie une erreur spécifique si accès non autorisé
- [ ] Par défaut, un nouvel utilisateur rejoignant l'équipe devient Membre

## Tests automatisés
- Test unitaire : Vérification de la logique d'assignation des rôles et contraintes (1 cocapitaine max)
- Test d'intégration : Test des permissions API selon les rôles (Capitaine, Cocapitaine, Éditeur, Membre)
- Test E2E : Capitaine assigne rôle Éditeur → Éditeur modifie une zone → vérification réussie

## Documentation
- docs/roles-permissions.md (à créer)
- Voir specification.md section 4.4

## Valeur utilisateur
Assure une gestion fine des permissions au sein de l'équipe pour maintenir l'intégrité des données.

## Dépendances
- Story #003 (Team Creation and Management)