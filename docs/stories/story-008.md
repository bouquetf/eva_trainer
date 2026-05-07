# Story #008 : API Security and Role Verification

## Description
En tant qu'administrateur système, je veux que chaque appel API vérifie les rôles utilisateur afin de sécuriser l'accès aux fonctionnalités.

## Critères d'acceptation
- [ ] Chaque fonctionnalité de l'application est identifiée en base de données
- [ ] Chaque rôle a accès à certaines de ces fonctionnalités
- [ ] La vérification des rôles se fait lors de chaque appel à l'API
- [ ] En cas d'accès non autorisé : code HTTP + message d'erreur spécifique
- [ ] Le rôle Admin a accès complet (lecture/écriture) à toutes les données (équivalent à toutes les équipes)
- [ ] Le premier utilisateur à se connecter est automatiquement promu administrateur
- [ ] Les autres administrateurs sont promus par un administrateur existant

## Tests automatisés
- Test unitaire : Middleware de vérification des rôles et permissions
- Test d'intégration : Appel API avec différents rôles → vérification codes HTTP (200, 403)
- Test E2E : Utilisateur non autorisé tente accès → erreur 403 → message spécifique

## Documentation
- docs/api-security.md (à créer)
- Voir specification.md section 4.2 (Rôles globaux)

## Valeur utilisateur
Assure que les données sont protégées et que chaque utilisateur n'accède qu'aux fonctionnalités autorisées.

## Dépendances
- Story #002 (User Authentication via Discord)
- Story #004 (Team Roles and Permissions)