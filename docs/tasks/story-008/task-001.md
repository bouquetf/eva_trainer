# Tâche #001 - Story #008 : API de promotion administrateur (back-end)

## Objectif

Vérifier et valider que les endpoints d'administration (promotion admin, liste utilisateurs) sont correctement implémentés et conformes à la spécification.

## Contexte

- Story #008 : `docs/stories/story-008.md` - API Security and Role Verification
- Dépend de : Story #002 (modèle User, `SetAdmin`, `IsFirstUser`), Story #004 (permissions, middleware)
- Nécessaire pour : Tâche #002 (frontend admin), Tâche #003 (tests et documentation)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois vérifier que les éléments suivants sont déjà implémentés et fonctionnels. Ne rien créer de nouveau — valider l'existant.

### Description fonctionnelle

#### État actuel (déjà implémenté)

Les éléments suivants sont déjà en place et doivent être validés :

**Partie 1 - UserRepository :**
- `FindAll()` déjà implémenté dans `internal/models/user.go` (ligne 95-111)
- `SetAdmin()` déjà implémenté (ligne 75-93)
- `AdminCount()` déjà implémenté (ligne 179-186)
- `IsFirstUser()` déjà implémenté (ligne 188-195)
- `RemoveAdmin()` déjà implémenté (ligne 113-126) — **attention** : la spec dit qu'un admin ne peut pas être rétrogradé. Vérifier si `ToggleAdmin` dans le handler admin utilise cette méthode et si c'est conforme.

**Partie 2 - Middleware RequireAdmin :**
- Fichier : `internal/middleware/admin.go` (ligne 9-32)
- Vérifie `user.IsAdmin` après `RequireAuth`
- Renvoie 403 `forbidden` avec message `"admin access required"` si pas admin

**Partie 3 - AdminHandler :**
- Fichier : `internal/handlers/admin.go`
- `ListUsers(c)` : GET `/api/v1/admin/users` — retourne tous les utilisateurs
- `GetUserDetail(c)` : GET `/api/v1/admin/users/:id` — détail utilisateur + membreships
- `ToggleAdmin(c)` : PUT `/api/v1/admin/users/:id/admin` — **toggle** admin (promote AND demote !)
  - Vérifier la conformité avec la spec : "Un admin ne peut pas être rétrogradé"
  - Actuellement, `ToggleAdmin` permet de retirer le statut admin → potentiellement non conforme
- `ListAllTeams(c)` : GET `/api/v1/admin/teams`
- `UpdateUserRole(c)`, `AddUserToTeam(c)`, `RemoveUserFromTeam(c)` : gestion des membres
- `ListPermissions(c)` : GET `/api/v1/admin/permissions` — matrice permissions
- `GetRolePermissions(c)` : GET `/api/v1/admin/permissions/:role`

**Partie 4 - Routes :**
- Fichier : `api/routes.go` — fonction `setupAdminRoutes` (ligne 193+)
- Groupe `/api/v1/admin` protégé par `RequireAuth` + `RequireAdmin`
- Routes déjà configurées

**Cas à valider :**
- **Conformité spec section 4.2** : "Un admin ne peut pas être rétrogradé." Le `ToggleAdmin` actuel permet de retirer le statut admin. Est-ce souhaité ?
  - Si non conforme → supprimer le endpoint de demotion, ne garder que la promotion
  - Si conforme → documenter la décision
- Migration : les permissions sont seedées via `000005_seed_permissions` (pas `000002`)
- Le modèle User a déjà `DiscordUsername` et `DisplayName`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/admin.go` | Vérifier | Valider la conformité de ToggleAdmin avec la spec |
| `internal/middleware/admin.go` | Vérifier | RequireAdmin middleware |
| `api/routes.go` | Vérifier | Routes admin déjà configurées |

### Points de vigilance

1. **`ToggleAdmin` permet la démotion** : la spec section 4.2 dit "Un admin ne peut pas être rétrogradé". Pourtant `ToggleAdmin` avec `is_admin: false` appelle `RemoveAdmin`. Vérifier avec le produit si c'est un comportement voulu ou s'il faut le supprimer.
2. **Migration** : utiliser `000005_seed_permissions` et non `000002` comme référencé dans les anciens documents.
3. **User model** : les champs `discord_username` et `display_name` sont déjà présents (migrations 000003 et 000004).

### Documentation

Aucune documentation à créer.
