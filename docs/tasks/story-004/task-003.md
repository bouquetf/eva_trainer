# Tâche #003 - Story #004 : Middleware de vérification des permissions

## Objectif

Créer un système de middleware qui vérifie les permissions lors des appels API, et l'intégrer aux handlers existants et futurs. Chaque appel API vérifie que l'utilisateur a le droit d'effectuer l'action demandée et renvoie une erreur spécifique si accès non autorisé.

## Contexte

- Story #004 : `docs/stories/story-004.md` - Team Roles and Permissions
- Dépend de : Tâche #001 (modèle HasPermission), Tâche #002 (endpoints rôles), Story #003 (RequireTeamRole middleware)
- Nécessaire pour : Tâche #005 (tests), toutes les stories suivantes utilisant les permissions
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Middleware `RequirePermission` (approche par composition) :**
Créer `internal/middleware/permission.go` comme un middleware **séparé et complémentaire** à `RequireTeamRole`. Les deux middlewares sont composés sur les routes : `RequireTeamRole` vérifie le rôle minimum, `RequirePermission` vérifie la permission spécifique.

```go
// RequirePermission vérifie que l'utilisateur connecté a une permission spécifique
// dans l'équipe. Ce middleware est conçu pour être utilisé EN PLUS de RequireTeamRole.
// Il présuppose que le contexte Gin contient déjà "user" (via RequireAuth) et que
// l'équipe :id est dans les paramètres.
// Paramètres :
//   - rolePermRepo: pour la vérification des permissions par rôle
//   - permRepo: pour la recherche de permission par feature
//   - feature: l'identifiant de la fonctionnalité (ex: "member:role:assign")
func RequirePermission(
    rolePermRepo models.RolePermissionRepository,
    permRepo models.PermissionRepository,
    feature string,
) gin.HandlerFunc
```

Fonctionnement :
1. Extrait `:id` (teamID) des paramètres Gin
2. Extrait l'utilisateur du contexte (via `RequireAuth`, clé `"user"`)
3. Vérifie si l'utilisateur est admin global (`user.IsAdmin`) → si oui, passe au handler suivant (l'admin a toutes les permissions sans vérification supplémentaire)
4. Extrait le `teamMember` du contexte (déjà placé par `RequireTeamRole`) → si absent, 500 `missing_team_member`
5. Vérifie la permission via `models.HasPermission(member.Role, feature, rolePermRepo, permRepo)` → si pas de permission, 403 avec le code d'erreur :
   ```json
   { "error": "insufficient_permission", "message": "you don't have the required permission for this action" }
   ```
6. Si OK, stocke `"permissions"` dans le contexte Gin (liste des features autorisées pour l'utilisateur) et passe au handler suivant

Utilisation typique sur une route :
```go
teams.PUT("/:id", 
    RequireTeamRole(teamMemberRepo, "captain"),
    RequirePermission(rolePermRepo, permRepo, "team:edit"),
    teamHandler.UpdateTeam,
)
```

**Partie 2 - Helper de récupération des permissions :**
Ajouter une fonction utilitaire :

```go
// GetUserPermissions retourne toutes les permissions d'un utilisateur dans une équipe.
// Pour un admin, retourne toutes les permissions (sans requêter la base).
// Pour un membre, retourne la liste des features auxquelles son rôle donne accès.
func GetUserPermissions(
    userID int,
    teamID int,
    userRepo models.UserRepository,
    teamMemberRepo models.TeamMemberRepository,
    rolePermRepo models.RolePermissionRepository,
    permRepo models.PermissionRepository,
) ([]string, error)
```

**Partie 3 - Aucune modification de RequireTeamRole :**
`RequireTeamRole` reste inchangé. Il continue de vérifier uniquement l'appartenance à l'équipe et le rôle minimum. `RequirePermission` s'ajoute en complément.

**Partie 4 - Application aux routes existantes (refactor) :**
Mettre à jour `api/routes.go` pour ajouter `RequirePermission` en complément de `RequireTeamRole` sur toutes les routes d'équipe.

Routes à mettre à jour :

| Route | Middleware existant | Nouveaux middlewares |
|-------|-------------------|---------------------|
| GET /api/v1/teams | RequireAuth | (inchangé : pas de permission d'équipe, pas de teamId) |
| GET /api/v1/teams/:id | RequireTeamRole("member") | RequireTeamRole("member") + RequirePermission(..., "member:view") |
| PUT /api/v1/teams/:id | RequireTeamRole("captain") | RequireTeamRole("captain") + RequirePermission(..., "team:edit") |
| DELETE /api/v1/teams/:id | RequireTeamRole("captain") | RequireTeamRole("captain") + RequirePermission(..., "team:delete") |
| POST /api/v1/teams/:id/leave | RequireTeamRole("member") | RequireTeamRole("member") + RequirePermission(..., "team:leave") |
| GET /api/v1/teams/:id/members | (depuis Task #002) RequireTeamRole("member") | RequireTeamRole("member") + RequirePermission(..., "member:view") |
| PUT /api/v1/teams/:id/members/:userId/role | (depuis Task #002) RequireTeamRole("captain") | RequireTeamRole("captain") + RequirePermission(..., "member:role:assign") |
| POST /api/v1/teams/:id/transfer-captaincy | (depuis Task #002) RequireTeamRole("captain") | RequireTeamRole("captain") + RequirePermission(..., "member:role:transfer-captain") |

Important : Pour `PUT /api/v1/teams/:id/members/:userId/role` avec rôle `cocaptain`, la permission `member:role:promote-cocaptain` est vérifiée. Il faut soit :
- Une deuxième route avec `RequirePermission(..., "member:role:promote-cocaptain")` pour les promotions en cocaptain
- Ou une vérification dans le handler si le body contient `role: "cocaptain"`

**Partie 5 - Helper `GetRequiredRole` :**
Pour les futures stories qui auront besoin de connaître le rôle minimum requis pour une permission, ajouter :

```go
// GetRequiredRole retourne le rôle minimum requis pour une feature donnée.
// Parcourt role_permissions et retourne le rôle le plus permissif.
// Utilisé pour le frontend (cacher/afficher des éléments selon le rôle).
func GetRequiredRole(feature string, rolePermRepo RolePermissionRepository) (string, error)
```

**Partie 6 - Admin global bypass :**
Le bypass admin est intégré dans `RequirePermission` : si `user.IsAdmin == true`, le middleware passe directement sans vérifier la permission. Les admins ne passent pas par `RequireTeamRole` (ils peuvent ne pas être membres d'une équipe). Donc pour les routes qui ont `RequireTeamRole` + `RequirePermission`, il faut que l'admin bypass dans `RequirePermission` gère aussi le cas où `RequireTeamRole` n'a pas placé `teamMember` dans le contexte.

**Cas nominaux :**
- Capitaine accède à une route `team:edit` → autorisé (passe)
- Éditeur accède à une route `team:edit` → refusé (403)
- Admin global accède à n'importe quelle route → autorisé (bypass)
- Membre accède à une route `member:view` → autorisé

**Cas limites :**
- Utilisateur non membre de l'équipe → 403 `not_team_member`
- Membre sans la permission spécifique → 403 avec le code d'erreur `insufficient_permission`
- Feature inexistante en base → 403 (considéré comme non autorisé)
- TeamID invalide (non numérique) → 400 `invalid_team_id`

**Gestion d'erreurs :**
| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | invalid_team_id | L'ID de l'équipe n'est pas un nombre valide |
| 403 | not_team_member | L'utilisateur n'est pas membre de l'équipe |
| 403 | insufficient_permission | Le rôle n'a pas la permission requise |
| 500 | internal_error | Erreur serveur |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/middleware/permission.go` | Créer | Middleware RequirePermission, helpers GetUserPermissions, GetRequiredRole |
| `internal/middleware/team.go` | Modifier | Ajouter helper GetTeamMemberFromContext si nécessaire |
| `api/routes.go` | Modifier | Ajouter RequirePermission en complément de RequireTeamRole sur toutes les routes |

### Ordre d'exécution

Cette tâche est exécutée APRÈS la Tâche #002. Les routes créées dans #002 avec `RequireTeamRole` sont refactorisées ici pour ajouter `RequirePermission` en complément.

### Signatures

```go
// internal/middleware/permission.go
package middleware

// RequirePermission vérifie que l'utilisateur connecté a la permission
// d'effectuer l'action sur l'équipe spécifiée.
func RequirePermission(
    teamMemberRepo models.TeamMemberRepository,
    rolePermRepo models.RolePermissionRepository,
    permRepo models.PermissionRepository,
    feature string,
    allowedRoles ...string,
) gin.HandlerFunc

// GetUserPermissions retourne toutes les permissions d'un utilisateur
// dans une équipe. Pour un admin global, retourne la liste complète.
func GetUserPermissions(
    userID int,
    teamID int,
    userRepo models.UserRepository,
    teamMemberRepo models.TeamMemberRepository,
    rolePermRepo models.RolePermissionRepository,
    permRepo models.PermissionRepository,
) ([]string, error)

// GetRequiredRole retourne le rôle minimum requis pour une feature.
func GetRequiredRole(
    feature string,
    rolePermRepo models.RolePermissionRepository,
) (string, error)
```

### Contraintes techniques

- **Composition, pas remplacement** : `RequirePermission` est un middleware indépendant qui s'ajoute PAR-DESSUS `RequireTeamRole`. Les deux sont appliqués sur les routes concernées.
- **Ordre des middlewares** : `RequireAuth` → `RequireTeamRole` → `RequirePermission` → Handler.
- **Dépendance RequireTeamRole** : `RequirePermission` utilise `teamMember` depuis le contexte Gin (déjà placé par `RequireTeamRole`). Si absent, c'est une erreur 500.
- **Admin bypass intégré** : Si `user.IsAdmin`, `RequirePermission` passe directement (même sans `teamMember` dans le contexte). Cela permet à un admin d'accéder à une route même s'il n'est pas membre de l'équipe.
- **Contexte Gin** : Stocker les permissions dans le contexte :
  ```go
  c.Set("permissions", permissions)  // []string des features autorisées
  ```
- **Gestion d'erreur feature inconnue** : Si `permRepo.FindByFeature(feature)` retourne `ErrPermissionNotFound`, le middleware renvoie 403 (feature inconnue = personne n'y a accès).
- **Aucune modification de RequireTeamRole** : Le middleware existant reste strictement inchangé pour garantir la rétrocompatibilité avec les handlers de Story #003.
- **Idempotence** : Si `RequirePermission` est appelée sur une route sans `RequireTeamRole` préalable et que l'utilisateur n'est pas admin, le middleware doit lui-même vérifier l'appartenance à l'équipe (via `teamMemberRepo`) avant de vérifier la permission.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
