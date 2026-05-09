# Tâche #003 - Story #008 : Tests d'intégration et documentation de sécurité

## Objectif

Vérifier que les tests du système de sécurité (promotion admin, middleware de vérification, accès API par rôles) existent et sont fonctionnels, et créer la documentation `docs/api-security.md` manquante.

## Contexte

- Story #008 : `docs/stories/story-008.md` - API Security and Role Verification
- Dépend de : Tâche #001 (backend validé), Tâche #002 (frontend validé)
- Nécessaire pour : Aucune (dernière tâche de la story)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois créer la documentation manquante et vérifier les tests existants.

### Description fonctionnelle

**Partie 1 - Tests existants à vérifier :**

Rechercher et évaluer les fichiers de test suivants :

- `internal/handlers/admin_test.go` — si existe, vérifier la couverture des scénarios
- `internal/middleware/permission_test.go` — vérifier les tests d'admin bypass
- `internal/models/user_test.go` — vérifier les tests de `FindAll`, `SetAdmin`, `IsFirstUser`

Scénarios qui doivent être couverts (créer les tests manquants si nécessaire) :

1. **GET /api/v1/admin/users - Admin list** → 200
2. **GET /api/v1/admin/users - Non-admin** → 403
3. **GET /api/v1/admin/users - Sans auth** → 401
4. **PUT /api/v1/admin/users/:id/admin - Promotion** → 200, is_admin = true
5. **PUT /api/v1/admin/users/:id/admin - Non-admin tente** → 403
6. **Admin bypass middleware** → admin accède à route protégée sans être membre → 200
7. **Permission refusée** → member accès à `team:delete` → 403
8. **Premier utilisateur auto-promu** → IsFirstUser + SetAdmin → is_admin = true

**Partie 2 - Documentation :**

Créer le fichier `docs/api-security.md` manquant avec le contenu suivant :

```markdown
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
- PUT `/api/v1/admin/users/:id/admin` avec `{ "is_admin": true }`
- Réservé aux administrateurs existants

## Vérification des permissions

### Backend
Chaque appel API passe par une chaîne de middlewares :

```
RequireAuth → RequireTeamRole (optionnel) → RequirePermission (optionnel) → Handler
```

### Bypass admin
Si `user.IsAdmin == true`, les middlewares `RequireTeamRole` et `RequirePermission` laissent passer.

### Middleware RequireAdmin
Pour les routes d'administration globale. Utilisé après `RequireAuth`.

## Endpoints d'administration

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| GET | `/api/v1/admin/users` | RequireAuth+Admin | Liste tous les utilisateurs |
| GET | `/api/v1/admin/users/:id` | RequireAuth+Admin | Détail utilisateur + équipes |
| PUT | `/api/v1/admin/users/:id/admin` | RequireAuth+Admin | Promouvoir/retirer admin |
| GET | `/api/v1/admin/permissions` | RequireAuth+Admin | Matrice des permissions |
| GET | `/api/v1/admin/permissions/:role` | RequireAuth+Admin | Permissions d'un rôle |
| GET | `/api/v1/admin/teams` | RequireAuth+Admin | Liste toutes les équipes |
| PUT | `/api/v1/admin/users/:id/teams/:teamId/role` | RequireAuth+Admin | Modifier rôle membre |
| POST | `/api/v1/admin/users/:id/teams` | RequireAuth+Admin | Ajouter membre à équipe |
| DELETE | `/api/v1/admin/users/:id/teams/:teamId` | RequireAuth+Admin | Retirer membre d'équipe |

## Gestion des erreurs

| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | invalid_user_id | ID utilisateur invalide |
| 401 | unauthorized | Token manquant ou invalide |
| 403 | forbidden | Accès admin requis |
| 403 | not_team_member | Pas membre de l'équipe |
| 403 | insufficient_permission | Permission insuffisante |
| 404 | user_not_found | Utilisateur introuvable |
| 500 | internal_error | Erreur serveur |

## Exemples curl

```bash
# Lister les utilisateurs (admin requis)
curl http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer <admin_jwt>"

# Promouvoir un utilisateur (admin requis)
curl -X PUT http://localhost:8080/api/v1/admin/users/2/admin \
  -H "Authorization: Bearer <admin_jwt>" \
  -H "Content-Type: application/json" \
  -d '{"is_admin": true}'

# Accès refusé (utilisateur non-admin)
curl http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer <user_jwt>"
# → 403 { "error": "forbidden", "message": "admin access required" }
```

## Dépendances
- Story #002 : Authentification Discord (JWT, refresh token, `is_admin`)
- Story #004 : Rôles d'équipe et permissions (middleware, seed data)
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/api-security.md` | Créer | Documentation de la sécurité API |
| `internal/handlers/admin_test.go` | Vérifier/Créer | Tests des endpoints admin |
| `internal/middleware/permission_test.go` | Vérifier | Tests middleware (admin bypass) |
| `docs/stories/story-008.md` | Modifier | Marquer les critères d'acceptation comme faits |

### Contraintes techniques

- **Migration** : la migration des permissions est `000005_seed_permissions` (pas `000002`)
- **Testcontainers** : utiliser le même setup que les tests existants
- **Format documentation** : Markdown, tableaux, blocs de code bash/JSON
- **Mise à jour story** : marquer tous les critères d'acceptation de `docs/stories/story-008.md` comme `[x]`

### Documentation

- `docs/api-security.md` : Document principal (créé par cette tâche)
- Voir `docs/authentication.md` pour l'auth Discord
- Voir `docs/roles-permissions.md` pour les rôles d'équipe
