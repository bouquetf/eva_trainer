# API Security - Eva Trainer

## Architecture de sécurité

Eva Trainer utilise un système de sécurité à plusieurs niveaux :

1. **Authentification** : JWT (24h) + Refresh Token avec rotation (voir `docs/authentication.md`)
2. **Rôles globaux** : Admin / Utilisateur
3. **Rôles d'équipe** : Capitaine, Co-capitaine, Éditeur, Membre (voir `docs/roles-permissions.md`)
4. **Permissions** : Vérification par fonctionnalité en base de données

## Rôle global Admin

### Privilèges
- Accès complet (lecture/écriture) à toutes les données
- Équivalent à appartenir à toutes les équipes
- Bypasse toutes les vérifications de permissions d'équipe
- Peut promouvoir d'autres utilisateurs en administrateur

### Promotion automatique
- Le premier utilisateur à se connecter via Discord est automatiquement promu administrateur
- Logique dans le callback OAuth (`internal/handlers/auth.go`)

### Promotion manuelle
- `POST /api/v1/admin/users/:id/promote` (body vide)
- Réservé aux administrateurs existants
- Retourne `400 already_admin` si l'utilisateur est déjà admin
- La démotion admin n'est pas supportée (conforme spec §4.2)

## Middleware

### Chaîne de middlewares

Chaque appel API passe par une chaîne de middlewares :

```
RequireAuth → RequireAdmin (optionnel) → RequireTeamRole (optionnel) → RequirePermission (optionnel) → Handler
```

### Bypass admin
Si `user.IsAdmin == true`, les middlewares `RequireTeamRole` et `RequirePermission` laissent passer (testé dans `internal/middleware/permission_test.go`).

### Middleware RequireAdmin
Pour les routes d'administration globale. Utilisé après `RequireAuth`. Retourne `403 { "error": "forbidden", "message": "admin access required" }` si l'utilisateur n'est pas admin.

## Endpoints d'administration

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/v1/admin/users` | RequireAuth+Admin | Liste tous les utilisateurs |
| GET | `/api/v1/admin/users/:id` | RequireAuth+Admin | Détail utilisateur + équipes |
| POST | `/api/v1/admin/users/:id/promote` | RequireAuth+Admin | Promouvoir un utilisateur admin |
| GET | `/api/v1/admin/permissions` | RequireAuth+Admin | Matrice des permissions |
| GET | `/api/v1/admin/permissions/:role` | RequireAuth+Admin | Permissions d'un rôle |
| GET | `/api/v1/admin/teams` | RequireAuth+Admin | Liste toutes les équipes |
| PUT | `/api/v1/admin/users/:id/teams/:teamId/role` | RequireAuth+Admin | Modifier rôle membre |
| POST | `/api/v1/admin/users/:id/teams` | RequireAuth+Admin | Ajouter membre à équipe |
| DELETE | `/api/v1/admin/users/:id/teams/:teamId` | RequireAuth+Admin | Retirer membre d'équipe |

## Gestion des erreurs

| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | `invalid_user_id` | ID utilisateur invalide |
| 400 | `already_admin` | L'utilisateur est déjà admin |
| 401 | `unauthorized` | Token manquant ou invalide |
| 403 | `forbidden` | Accès admin requis |
| 403 | `not_team_member` | Pas membre de l'équipe |
| 403 | `insufficient_permission` | Permission insuffisante |
| 404 | `user_not_found` | Utilisateur introuvable |
| 500 | `internal_error` | Erreur serveur |

## Exemples curl

```bash
# Lister les utilisateurs (admin requis)
curl http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer <admin_jwt>"

# Promouvoir un utilisateur (admin requis)
curl -X POST http://localhost:8080/api/v1/admin/users/2/promote \
  -H "Authorization: Bearer <admin_jwt>"

# Promouvoir un admin existant → 400 already_admin
curl -X POST http://localhost:8080/api/v1/admin/users/1/promote \
  -H "Authorization: Bearer <admin_jwt>"
# → 400 { "error": "already_admin", "message": "User is already an admin" }

# Accès refusé (utilisateur non-admin)
curl http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer <user_jwt>"
# → 403 { "error": "forbidden", "message": "admin access required" }
```

## Tests

### Fichiers de test

| Fichier | Type | Description |
|---------|------|-------------|
| `internal/handlers/admin_test.go` | Intégration | Tests des endpoints admin (ListUsers, PromoteToAdmin, etc.) |
| `internal/middleware/permission_test.go` | Unitaire | Tests middleware permission et admin bypass |
| `internal/models/user_test.go` | Unitaire | Tests SetAdmin, IsFirstUser, FindAll |

### Scénarios couverts

- **GET /admin/users** → 200 pour admin, 403 pour non-admin
- **POST /admin/users/:id/promote** → 200 si succès, 400 si déjà admin
- **Admin bypass** → admin passe outre les vérifications de permission
- **Permission refusée** → member sans permission → 403
- **IsFirstUser** → détection du premier utilisateur
- **SetAdmin** → promotion en admin, user not found

## Dépendances
- Story #002 : Authentification Discord (JWT, refresh token, `is_admin`)
- Story #004 : Rôles d'équipe et permissions (middleware, seed data)
