# Système de rôles et permissions d'équipe

## Vue d'ensemble

Le système de rôles d'équipe permet de gérer finement les permissions des membres au sein d'une équipe. Quatre rôles sont définis, auxquels s'ajoute un rôle global d'administrateur.

## Rôles d'équipe

| Rôle | Description |
|------|-------------|
| **Capitaine** (`captain`) | Tous les droits. Peut transférer son rôle. Ne peut pas être rétrogradé. |
| **Co-capitaine** (`cocaptain`) | Mêmes droits que le Capitaine sauf suppression d'équipe, promotion en co-capitaine et transfert de capitainerie. |
| **Éditeur** (`editor`) | Modification des zones, lancement de révisions, duplication de templates. |
| **Membre** (`member`) | Accès aux zones, révision, duplication de templates. |

## Rôle global Admin

- Accès complet à toutes les fonctionnalités
- Ne peut pas être rétrogradé depuis une équipe
- Bypasse toutes les vérifications de permissions d'équipe
- Identifié par `is_admin = true` dans la table `users`

## Contraintes

- 1 co-capitaine maximum par équipe
- Un admin ne peut pas être rétrogradé
- Le capitaine ne peut être rétrogradé que par transfert de capitainerie
- Par défaut, un nouveau membre reçoit le rôle `member`

## Assignation des rôles

| Action | Captain | Co-captain | Editor | Member |
|--------|---------|------------|--------|--------|
| Définir comme Éditeur | ✅ | ✅ | ❌ | ❌ |
| Définir comme Membre | ✅ | ✅ | ❌ | ❌ |
| Promouvoir Co-capitaine | ✅ | ❌ | ❌ | ❌ |
| Transférer Capitainerie | ✅ | ❌ | ❌ | ❌ |

## Transfert de capitainerie

- Le capitaine transfère son rôle à un autre membre
- L'ancien capitaine devient co-capitaine
- Nécessite une confirmation explicite
- Les deux membres doivent être dans l'équipe
- Le nouveau capitaine ne peut pas être un admin global

## Visibilité de l'éditeur dans l'interface

- Le bouton "Éditeur" est visible pour les rôles **editor**, **cocaptain**, **captain** uniquement
- Les membres (role **member**) ne voient pas le bouton
- L'accès direct à l'URL `/teams/:id/editor/:mapId` redirige les membres vers la page d'équipe
- Le bouton "Restaurer" dans l'historique est également masqué pour les membres
- Les liens "Retour à l'éditeur" dans l'historique sont masqués pour les membres

La vérification est effectuée côté frontend via le hook `useTeamRole(teamId)` et côté backend par le middleware `RequireTeamRole`.

## Système de permissions

Chaque fonctionnalité de l'application est identifiée en base de données (table `permissions`). Chaque rôle a des permissions associées (table `role_permissions`). La vérification se fait à chaque appel API via middleware.

### Features et rôles autorisés

| Feature | Captain | Co-captain | Editor | Member |
|---------|---------|------------|--------|--------|
| `team:create` | - | - | - | ✅ |
| `team:edit` | ✅ | ✅ | ❌ | ❌ |
| `team:delete` | ✅ | ❌ | ❌ | ❌ |
| `team:leave` | ✅ | ✅ | ✅ | ✅ |
| `team:invite` | ✅ | ✅ | ❌ | ❌ |
| `team:requests:view` | ✅ | ✅ | ❌ | ❌ |
| `team:requests:manage` | ✅ | ✅ | ❌ | ❌ |
| `member:view` | ✅ | ✅ | ✅ | ✅ |
| `member:kick` | ✅ | ❌ | ❌ | ❌ |
| `member:role:assign` | ✅ | ✅ | ❌ | ❌ |
| `member:role:promote-cocaptain` | ✅ | ❌ | ❌ | ❌ |
| `member:role:transfer-captain` | ✅ | ❌ | ❌ | ❌ |
| `zone:view` | ✅ | ✅ | ✅ | ✅ |
| `zone:create` | ✅ | ✅ | ✅ | ❌ |
| `zone:edit` | ✅ | ✅ | ✅ | ❌ |
| `zone:delete` | ✅ | ✅ | ✅ | ❌ |
| `zone:publish` | ✅ | ✅ | ❌ | ❌ |
| `zone:lock` | ✅ | ✅ | ❌ | ❌ |
| `review:start` | ✅ | ✅ | ✅ | ✅ |
| `template:duplicate` | ✅ | ✅ | ✅ | ✅ |
| `edit_history:view` | ✅ | ✅ | ❌ | ✅ |
| `edit_history:rollback` | ✅ | ❌ | ❌ | ❌ |

## Endpoints API

| Méthode | Path | Auth | Rôle min. | Permission | Description |
|---------|------|------|-----------|------------|-------------|
| GET | `/api/v1/teams/:id/members` | Oui | member | `member:view` | Liste des membres |
| PUT | `/api/v1/teams/:id/members/:userId/role` | Oui | captain/cocaptain | `member:role:assign` | Assigner un rôle |
| POST | `/api/v1/teams/:id/transfer-captaincy` | Oui | captain | `member:role:transfer-captain` | Transférer la capitainerie |

### Exemples curl

```bash
# Lister les membres
curl http://localhost:8080/api/v1/teams/1/members \
  -H "Authorization: Bearer <jwt>"

# Assigner le rôle Éditeur
curl -X PUT http://localhost:8080/api/v1/teams/1/members/2/role \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"role": "editor"}'

# Promouvoir Co-capitaine
curl -X PUT http://localhost:8080/api/v1/teams/1/members/3/role \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"role": "cocaptain"}'

# Transférer la capitainerie
curl -X POST http://localhost:8080/api/v1/teams/1/transfer-captaincy \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 4}'
```

## Gestion des erreurs

| Code HTTP | `error` | Cause |
|-----------|---------|-------|
| 400 | `invalid_role` | Rôle invalide |
| 400 | `team_already_has_cocaptain` | Déjà un cocapitaine |
| 400 | `same_user` | Transfert vers soi-même |
| 400 | `confirmation_name_mismatch` | Nom de confirmation incorrect |
| 403 | `not_team_member` | Pas membre de l'équipe |
| 403 | `insufficient_role` | Rôle insuffisant |
| 403 | `insufficient_permission` | Permission insuffisante |
| 403 | `cannot_downgrade_admin` | Tentative de rétrograder un admin |
| 403 | `cannot_downgrade_captain` | Tentative de rétrograder le capitaine |
| 404 | `member_not_found` | Membre introuvable |
| 500 | `internal_error` | Erreur serveur |
