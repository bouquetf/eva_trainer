# Story #002 : User Authentication via Discord

## Description
En tant qu'utilisateur, je veux me connecter via Discord OAuth2 afin d'accéder à l'application Eva Trainer.

## Critères d'acceptation
- [ ] L'utilisateur peut se connecter via un bouton "Connexion avec Discord"
- [ ] Après authentification Discord, un JWT est généré (validité 24h) et stocké
- [ ] Un refresh token est implémenté pour renouveler la session
- [ ] Rotation des refresh tokens : un nouveau refresh token est généré à chaque utilisation
- [ ] Si l'API Discord est indisponible, un message d'erreur "connexion impossible" s'affiche
- [ ] L'ID Discord unique est stocké en base (le nom Discord n'est pas utilisé)
- [ ] Les sessions sont gérées côté serveur avec JWT

## Tests automatisés
- Test unitaire : Validation du flux OAuth2 et génération du JWT
- Test d'intégration : Simulation de l'authentification Discord et vérification du stockage du token
- Test E2E : Clic sur "Connexion avec Discord" → redirection → retour avec session active

## Documentation
- docs/authentication.md (à créer)
- Voir specification.md section 4.1

## Valeur utilisateur
Permet aux utilisateurs d'accéder à l'application de manière sécurisée via leur compte Discord existant.

## Dépendances
- Story #001 (Configuration and Deployment)
- Configuration des variables d'environnement Discord (CLIENT_ID, CLIENT_SECRET)