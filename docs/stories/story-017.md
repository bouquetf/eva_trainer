# Story #017 : Test Automation Strategy

## Description
En tant que développeur, je veux une stratégie de tests automatisés complète afin d'assurer la qualité et la stabilité de l'application.

## Critères d'acceptation
- [ ] Tests unitaires pour la logique métier (Go avec framework de test standard)
- [ ] Tests d'intégration pour les API REST (Gin + PostgreSQL)
- [ ] Couverture fonctionnelle à 100%
- [ ] Tests End-to-End (E2E) avec **Playwright** pour les happy path et schémas classiques d'utilisateurs
- [ ] Tests E2E pour l'authentification Discord
- [ ] Tests E2E pour le mode Édition (création zones, historique, rollback)
- [ ] Tests E2E pour le mode Révision (validation, poids, score)
- [ ] Tests E2E pour la gestion des équipes (création, rôles, invitations)
- [ ] Intégration continue configurée pour exécuter les tests automatiquement

## Tests automatisés
- Test unitaire : Framework de test Go pour les services et repositories
- Test d'intégration : Test des endpoints API avec base de données de test
- Test E2E : Utilisation d'un framework E2E (Cypress, Playwright, ou Selenium) pour les parcours utilisateur

## Documentation
- docs/testing-strategy.md (à créer)
- Voir specification.md section 5 (Tests)

## Valeur utilisateur
Assure la fiabilité de l'application et réduit les régressions lors des évolutions.

## Dépendances
- Toutes les autres stories (les tests s'appuient sur les fonctionnalités implémentées)
