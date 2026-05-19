# Story #028 : Mode "Navigation en tant que" pour administrateurs

## Description
En tant qu'administrateur, je veux pouvoir naviguer sur le site en tant qu'utilisateur, membre d'une équipe, éditeur, cocapitaine et capitaine, avec les restrictions d'accès correspondant à chaque rôle, afin de tester et valider le comportement du site depuis la perspective de chaque rôle.

## Origine
Feedback du 2026-05-17 : "En tant qu'administrateur, je veux avoir la possibilité de naviguer sur le site en tant qu'utilisateur, membre d'une équipe donnée, cocapitanie de l'équipe et capitaine. Cela cache les autres fonctionnalités auxquelles ce rôle n'a pas accès à l'exception du panneau d'administration. La fonctionnalité s'active/configure et se désactive dans le panel d'administration."

## Critères d'acceptation
- [ ] L'administrateur peut activer le mode "Navigation en tant que" depuis le panneau d'administration
- [ ] L'administrateur peut sélectionner un **rôle générique** (utilisateur sans équipe, membre, éditeur, cocapitaine, capitaine)
- [ ] L'administrateur peut sélectionner un **utilisateur spécifique** de la base pour voir exactement ce que cet utilisateur voit
- [ ] L'administrateur peut sélectionner une équipe spécifique pour les rôles liés à une équipe (membre, éditeur, cocapitaine, capitaine)
- [ ] En mode navigation, l'interface se comporte exactement comme pour le rôle cible : seules les fonctionnalités autorisées pour ce rôle sont visibles
- [ ] Le panneau d'administration reste accessible même en mode navigation
- [ ] L'administrateur peut désactiver le mode navigation depuis n'importe quelle page
- [ ] Un indicateur visuel clair (ex : bandeau en haut de l'écran) signale que le mode navigation est actif, avec le rôle et l'équipe simulée
- [ ] Une checkbox "Mode sandbox" est disponible lors de l'activation : quand elle est cochée, les actions effectuées en mode navigation sont en lecture seule (aucune modification réelle)

## Tests automatisés
- Test unitaire : Logique de restriction d'affichage selon le rôle simulé, y compris l'accès prioritaire au panneau d'administration
- Test d'intégration : API vérifie que l'administrateur en mode navigation reçoit les mêmes permissions que le rôle cible
- Test d'intégration : API en mode sandbox rejette les modifications (lecture seule)
- Test E2E : Activation du mode navigation depuis l'admin → sélection d'un rôle générique membre → vérification que l'éditeur est masqué → désactivation → retour à la vue administrateur complète
- Test E2E : Activation du mode navigation avec sélection d'un utilisateur spécifique → vérification que l'interface correspond aux permissions de cet utilisateur
- Test E2E : Activation du mode sandbox → tentative de modification → vérification que la modification est rejetée

## Documentation
- docs/admin-impersonation.md

## Valeur utilisateur
Permet aux administrateurs de valider visuellement et fonctionnellement l'expérience de chaque rôle, facilitant le débogage et la vérification des permissions sans avoir à créer plusieurs comptes de test.

## Dépendances
- Story #004 (Team Roles and Permissions)
- Story #024 (Visibilité des rôles dans l'interface éditeur)
