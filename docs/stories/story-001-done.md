# Story #001 : Configuration and Deployment

## Description
En tant qu'administrateur système, je veux configurer l'application et la déployer afin de la rendre disponible en production.

## Critères d'acceptation
- [ ] Support des variables d'environnement (.env et variables système)
- [ ] Les variables système sont prioritaires sur le .env
- [ ] Si un .env n'existe pas, utilisation des variables système
- [ ] Si des paramètres de configuration manquent, une erreur est levée et le serveur s'arrête
- [ ] Déploiement local : Docker Compose (PostgreSQL seul ou application complète)
- [ ] En dev, frontend servi via npm, backend Go redémarrable indépendamment
- [ ] En prod, un seul conteneur sert le frontend et le backend (Go sert le frontend)
- [ ] HTTPS géré par Clever Cloud en production
- [ ] En prod, support juste du serveur local pour CORS
- [ ] En dev, configuration CORS surchargée dans le .env pour accepter toutes les origines

## Tests automatisés
- Test unitaire : Chargement et priorité des variables de configuration
- Test d'intégration : Démarrage application avec .env vs variables système
- Test E2E : Déploiement Docker Compose → vérification accessibilité → test configuration CORS

## Documentation
- docs/configuration.md (à créer)
- docs/deployment.md (à créer)
- Voir specification.md section 5 (Technique)

## Valeur utilisateur
Permet un déploiement fiable et configurable de l'application en développement et production.

## Dépendances
- Aucune