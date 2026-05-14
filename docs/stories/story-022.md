# Story #022 : Suppression de la page d'accueil

## Description
En tant qu'utilisateur, je veux être redirigé directement vers "Mes équipes" après connexion afin d'accéder plus rapidement à mes fonctionnalités.

## Origine
Feedback du 2025-05-14 : "Ne pas afficher d'accueil : on tombe direct sur 'Mes équipes'"

## Critères d'acceptation
- [ ] La page d'accueil (`/`) redirige automatiquement vers la page `/teams` ("Mes équipes")
- [ ] Après connexion, l'utilisateur arrive directement sur `/teams`
- [ ] Si l'utilisateur n'est pas connecté, redirection vers la page de connexion Discord, puis vers `/teams` après authentification
- [ ] Si l'utilisateur connecté n'appartient à aucune équipe : la page `/teams` affiche un état vide avec la possibilité de créer une équipe ou d'en rejoindre une existante
- [ ] L'admin est également redirigé vers `/teams` (même comportement que les utilisateurs)
- [ ] La route `/` n'affiche plus de contenu statique ou de dashboard

## Tests automatisés
- Test unitaire : Logique de redirection selon l'état d'authentification
- Test d'intégration : Vérification des routes et redirections
- Test E2E : Utilisateur connecté accède à `/` → redirigé vers `/teams` ("Mes équipes")

## Documentation
- Mise à jour du routage frontend

## Valeur utilisateur
Réduit le nombre de clics pour accéder à la fonctionnalité principale (mes équipes) et simplifie le parcours utilisateur.

## Dépendances
- Story #003 (Team Creation and Management)
