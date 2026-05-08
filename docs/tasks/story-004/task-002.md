# Tâche #002 - Story #004 : Endpoints API de gestion des rôles (back-end)

## Objectif

Implémenter les endpoints API REST de gestion des rôles d'équipe : assignation de rôle, transfert de capitainerie, et liste des membres avec leurs rôles.

## Contexte

- Story #004 : `docs/stories/story-004.md` - Team Roles and Permissions
- Dépend de : Tâche #001 (modèle Permission, RolePermission), Story #003 (modèles Team/TeamMember, middleware RequireAuth, RequireTeamRole)
- Nécessaire pour : Tâche #003 (middleware permission), Tâche #004 (frontend), Tâche #005 (tests)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Handler TeamRoles :**
Créer `internal/handlers/team_roles.go` avec les endpoints suivants :

| Méthode | Path | Auth | Rôle requis | Description |
|---------|------|------|-------------|-------------|
| GET | `/api/v1/teams/:id/members` | RequireAuth | member | Lister les membres de l'équipe avec leurs rôles |
| PUT | `/api/v1/teams/:id/members/:userId/role` | RequireAuth | captain | Assigner un rôle à un membre |
| POST | `/api/v1/teams/:id/transfer-captaincy` | RequireAuth | captain | Transférer le rôle de capitaine |

**Partie 2 - Détail des endpoints :**

**GET `/api/v1/teams/:id/members`** - Liste des membres
- Vérifie que l'utilisateur est membre de l'équipe (via `RequireTeamRole` avec `"member"`)
- Appelle `teamMemberRepo.FindByTeamID(teamID)` qui retourne tous les membres
- Pour chaque membre, enrichit avec le nom d'utilisateur (via `userRepo.FindByID`)
- Réponse 200 : `{ "members": [{ "user_id", "username", "role", "joined_at" }, ...] }`
- L'utilisateur connecté voit son propre userId et peut visualiser les rôles de tous
- Erreur 403 : pas membre de l'équipe

**PUT `/api/v1/teams/:id/members/:userId/role`** - Assigner un rôle
- Vérifie que l'utilisateur connecté a le bon rôle :
  - Pour assigner `cocaptain` : `RequireTeamRole("captain")` (seul le capitaine peut promouvoir en cocapitaine)
  - Pour assigner `editor` ou `member` : `RequireTeamRole("captain", "cocaptain")` (le capitaine ET le cocapitaine peuvent assigner ces rôles)
- Le choix du middleware se fait au niveau de la route via deux routes distinctes ou via une vérification en dur dans le handler
- Body : `{ "role": "member" | "editor" | "cocaptain" }`
- **Validations strictes (dans cet ordre) :**
  1. Le `userId` cible doit être membre de l'équipe → si non, 404 `member_not_found`
  2. Le rôle cible doit être valide → si non, 400 `invalid_role`
  3. L'utilisateur cible ne peut pas être rétrogradé s'il est admin global (vérifier `user.IsAdmin`) → si admin, 403 `cannot_downgrade_admin`
  4. Le capitaine ne peut pas être rétrogradé (le capitaine ne peut être rétrogradé que via transfert) → si userId = captainID, 403 `cannot_downgrade_captain`
  5. Un seul cocapitaine maximum → si nouveau rôle = `cocaptain`, vérifier `CountByRole(teamID, "cocaptain") < 1`, sinon 400 `team_already_has_cocaptain`
  6. Le cocapitaine ne peut pas promouvoir quelqu'un en cocapitaine → si l'utilisateur connecté est cocapitaine ET que le nouveau rôle est `cocaptain`, 403 `insufficient_role`
  7. Assignation à "cocaptain" : vérifier que l'utilisateur connecté est capitaine (déjà garanti par le middleware)
  8. Assignation à "editor" ou "member" : vérifier que l'utilisateur connecté a le rôle captain ou cocaptain (déjà garanti par le middleware)
- Appelle `teamMemberRepo.UpdateRole(teamID, targetUserID, role)`
- Réponse 200 : `{ "message": "role updated", "user_id": int, "role": "string" }`
- Erreur 400 : rôle invalide / déjà un cocapitaine
- Erreur 403 : pas les permissions / admin non rétrogradeable / capitaine non rétrogradeable
- Erreur 404 : membre cible introuvable

Note : Dans cette tâche, la vérification se fait via `RequireTeamRole`. La migration vers `RequirePermission` sera faite dans la Tâche #003.

**POST `/api/v1/teams/:id/transfer-captaincy`** - Transférer la capitainerie
- Vérifie que l'utilisateur connecté est capitaine de l'équipe (via `RequireTeamRole("captain")`)
- Body : `{ "user_id": int }` — l'ID du membre qui devient capitaine
- **Validations (dans cet ordre) :**
  1. Le `user_id` cible doit être membre de l'équipe → 404 si non
  2. Le `user_id` cible ne peut pas être le capitaine actuel → 400 si identique
  3. L'utilisateur cible ne doit pas être admin global (un admin ne peut pas être promu capitaine d'équipe) → 403 si admin (car un admin ne peut pas être rétrogradé et deviendrait "non-rétrogradeable" en tant que simple membre)
- Si validation OK :
  - L'ancien capitaine devient `cocaptain` (via `teamMemberRepo.UpdateRole`)
  - Le nouveau capitaine devient `captain` (via `teamMemberRepo.UpdateRole`)
  - Mise à jour de `teams.captain_id` (via `teamRepo.UpdateCaptain(teamID, newCaptainID)`)
- Ces opérations doivent être atomiques → utiliser une **transaction**
- Réponse 200 : `{ "message": "captaincy transferred", "new_captain_id": int }`
- Erreur 400 : même utilisateur
- Erreur 403 : admin cible / pas capitaine
- Erreur 404 : cible introuvable
- **Note** : Après transfert, l'ancien capitaine devient un cocapitaine normal. Il peut être rétrogradé au rang d'éditeur ou membre par le nouveau capitaine (aucune protection spéciale après transfert).

**Partie 3 - Nouvelle méthode `UpdateCaptain` dans TeamRepository :**
Ajouter à `internal/models/team.go` et à l'interface `TeamRepository` :

```go
// UpdateCaptain met à jour le capitaine de l'équipe.
// UPDATE teams SET captain_id = $1, updated_at = NOW() WHERE id = $2
func (r *teamRepo) UpdateCaptain(teamID, newCaptainID int) error
```

**Partie 4 - Intégration des routes :**
Modifier `api/routes.go` pour ajouter les nouvelles routes dans `setupTeamRoutes` :

```go
// Dans setupTeamRoutes, après les routes existantes
teams.GET("/:id/members", teamRolesHandler.ListMembers)
teams.PUT("/:id/members/:userId/role", teamRolesHandler.UpdateRole)
teams.POST("/:id/transfer-captaincy", teamRolesHandler.TransferCaptaincy)
```

Le `TeamRolesHandler` est instancié avec les repositories nécessaires :

```go
teamRolesHandler := handlers.NewTeamRolesHandler(teamRepo, teamMemberRepo, userRepo, rolePermRepo, permRepo)
```

**Cas nominaux :**
- Capitaine assigne le rôle Éditeur à un membre → HTTP 200, rôle mis à jour en base
- Capitaine promet un membre en Cocapitaine → HTTP 200, vérifier limite d'1 cocapitaine
- Capitaine transfère son rôle à un autre membre → HTTP 200, l'ancien devient cocapitaine, le nouveau devient capitaine
- Membre liste les membres de son équipe → HTTP 200 avec la liste complète

**Cas limites :**
- Tentative d'assigner un rôle inexistant (`"superadmin"`) → 400 `invalid_role`
- Tentative d'assigner un 2ème cocapitaine → 400 `team_already_has_cocapitain`
- Tentative de rétrograder un admin → 403 `cannot_downgrade_admin`
- Tentative de rétrograder le capitaine (sans transfert) → 403 `cannot_downgrade_captain`
- Transfert vers soi-même → 400
- Transfert vers un non-membre → 404
- Membre non-autorisé tente d'assigner un rôle → 403

**Gestion d'erreurs :**
Format JSON standard (même format que les handlers d'équipe) :
```json
{
  "error": "error_code",
  "message": "Description lisible"
}
```

| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | invalid_role | Rôle invalide |
| 400 | team_already_has_cocaptain | Déjà un cocapitaine |
| 400 | same_user | Transfert vers soi-même |
| 403 | not_team_member | Pas membre de l'équipe |
| 403 | insufficient_role | Rôle insuffisant pour l'action |
| 403 | cannot_downgrade_admin | Tentative de rétrograder un admin |
| 403 | cannot_downgrade_captain | Tentative de rétrograder le capitaine sans transfert |
| 404 | member_not_found | Membre cible introuvable |
| 404 | team_not_found | Équipe introuvable |
| 500 | internal_error | Erreur serveur |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/team_roles.go` | Créer | TeamRolesHandler avec les 3 endpoints |
| `internal/models/team.go` | Modifier | Ajouter UpdateCaptain à TeamRepository |
| `api/routes.go` | Modifier | Ajouter les nouvelles routes rôle |

### Signatures

```go
// internal/handlers/team_roles.go
package handlers

type TeamRolesHandler struct {
    teamRepo       models.TeamRepository
    teamMemberRepo models.TeamMemberRepository
    userRepo       models.UserRepository
    rolePermRepo   models.RolePermissionRepository
    permRepo       models.PermissionRepository
}

func NewTeamRolesHandler(
    teamRepo models.TeamRepository,
    teamMemberRepo models.TeamMemberRepository,
    userRepo models.UserRepository,
    rolePermRepo models.RolePermissionRepository,
    permRepo models.PermissionRepository,
) *TeamRolesHandler

func (h *TeamRolesHandler) ListMembers(c *gin.Context)
func (h *TeamRolesHandler) UpdateRole(c *gin.Context)
func (h *TeamRolesHandler) TransferCaptaincy(c *gin.Context)
```

```go
// Ajout dans internal/models/team.go (interface et implémentation)
// TeamRepository interface - ajouter :
UpdateCaptain(teamID, newCaptainID int) error
```

### Contraintes techniques

- **Gin** : Utiliser `c.ShouldBindJSON` pour les body, `c.Param("id")` et `c.Param("userId")` pour les paramètres d'URL.
- **Conversion ID** : Utiliser `strconv.Atoi` pour convertir les paramètres en entiers.
- **Pattern handler** : Suivre le pattern des handlers existants (validation → action → réponse JSON).
- **Transaction pour transfert** : Le transfert de capitainerie doit utiliser une transaction SQL (BEGIN, 2x UPDATE team_members, UPDATE teams, COMMIT). Utiliser le même pattern que `TeamRepository.Create`.
- **Récupération utilisateur** : Le `userRepo.FindByID(userID)` doit être disponible (créé dans Story #002).
- **Vérification de permission** : Utiliser `models.HasPermission(role, feature, rolePermRepo, permRepo)` défini dans la Tâche #001. Le rôle de l'utilisateur connecté est extrait du contexte Gin via `middleware.GetTeamMember(c)`.
- **Rôle dans le contexte** : Le middleware `RequireTeamRole` stocke déjà `*models.TeamMember` dans le contexte sous la clé `"team_member"`.
- **Validation des body** : Les body doivent être validés avec `c.ShouldBindJSON` et les champs obligatoires vérifiés.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
