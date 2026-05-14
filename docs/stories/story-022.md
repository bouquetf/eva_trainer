# Story #022 : Suppression de la page d'accueil

## Description
En tant qu'utilisateur, je veux être redirigé directement vers "Mes équipes" après connexion afin d'accéder plus rapidement à mes fonctionnalités.

## Origine
Feedback du 2025-05-14 : "Ne pas afficher d'accueil : on tombe direct sur 'Mes équipes'"

## Critères d'acceptation
- [ ] La page d'accueil (`/`) redirige automatiquement vers la page "Mes équipes"
- [ ] Après connexion, l'utilisateur arrive directement sur "Mes équipes"
- [ ] Si l'utilisateur n'est pas connecté, redirection vers la page de connexion Discord, puis vers "Mes équipes" après authentification
- [ ] La route `/` n'affiche plus de contenu statique ou de dashboard

## Tests automatisés
- Test unitaire : Logique de redirection selon l'état d'authentification
- Test d'intégration : Vérification des routes et redirections
- Test E2E : Utilisateur connecté accède à `/` → redirigé vers "Mes équipes"

## Documentation
- Mise à jour du routage frontend

## Valeur utilisateur
Réduit le nombre de clics pour accéder à la fonctionnalité principale (mes équipes) et simplifie le parcours utilisateur.

## Dépendances
- Story #003 (Team Creation and Management)
