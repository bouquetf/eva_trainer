# Tâche #005 - Story #004 : Tests et documentation des rôles et permissions

## Objectif

Créer les tests automatisés (back-end) et la documentation utilisateur pour la gestion des rôles et permissions d'équipe.

## Contexte

- Story #004 : `docs/stories/story-004.md` - Team Roles and Permissions
- Dépend de : Tâche #001 (modèles permissions), Tâche #002 (handlers rôles), Tâche #003 (middleware permissions), Tâche #004 (frontend rôles)
- Nécessaire pour : Stories #006 (Invitations), #008 (Zones), #009 (Visibilité), #011 (Templates), #012 (Révisions)
- Ordre d'exécution : 5ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Tests d'intégration du système de permissions :**

#### Fichier `internal/models/permission_test.go`

Tests des repositories Permission et RolePermission via testcontainers (même pattern que `internal/models/user_test.go`).

Scénarios :

1. **Création d'une permission**
   - Données : feature = "zone:test", description = "Test permission"
   - Résultat : Permission créée avec ID, feature et description corrects

2. **Récupération par feature**
   - Données : Feature "team:create" (seed)
   - Résultat : Permission retournée avec les bons champs

3. **Récupération par ID**
   - Données : ID existant
   - Résultat : Permission retournée

4. **Récupération de toutes les permissions**
   - Résultat : Liste contenant au moins toutes les permissions seedées

5. **Feature inexistante**
   - Données : feature = "nonexistent:feature"
   - Résultat : Erreur `ErrPermissionNotFound`

6. **Assignation d'une permission à un rôle**
   - Données : role = "editor", permissionID existant
   - Résultat : `RoleHasPermission("editor", permissionID)` retourne `true`

7. **Vérification `RoleHasPermission`**
   - Données : role = "captain", feature "team:delete" (seed)
   - Résultat : `true`
   - Données : role = "editor", feature "team:delete" (seed)
   - Résultat : `false`

8. **Récupération des permissions d'un rôle**
   - Données : role = "captain"
   - Résultat : Liste contenant toutes les features auxquelles le capitaine a accès

9. **Retrait d'une permission**
   - Données : Assigner puis retirer une permission à "editor"
   - Résultat : `RoleHasPermission` retourne `false` après retrait

10. **Admin bypass**
    - Données : Appel `HasPermission("admin", "map:add", ...)` (sans seed pour cette feature)
    - Résultat : `true` (admin bypass sans requête DB)
    - Note : ne pas utiliser le repository réel, juste tester la logique de la fonction

11. **Feature inconnue avec `HasPermission`**
    - Données : feature = "does:not:exist"
    - Résultat : `false, nil` (pas d'erreur, pas de permission)

#### Fichier `internal/models/permission_test.go` (suite) - Tests de `SeedPermissions`

12. **Idempotence de la migration SQL**
    - Données : Exécuter `000002_seed_permissions.up.sql` deux fois
    - Résultat : Pas d'erreur, pas de doublons (ON CONFLICT DO NOTHING), toutes les permissions présentes une seule fois

13. **Idempotence de SeedRolePermissions**
    - Données : Appeler `SeedRolePermissions(db)` deux fois
    - Résultat : Pas d'erreur, pas de doublons dans `role_permissions`

14. **Vérification complète des permissions seedées**
    - Données : La liste de toutes les features définies dans la tâche #001
    - Résultat : Chaque feature existe en base (via migration), chaque association rôle-permission est correcte (via seed Go)

#### Fichier `internal/middleware/permission_test.go`

Tests du middleware de permission via httptest.

Scénarios :

1. **Admin global bypass**
   - Setup : Utilisateur avec `is_admin = true`
   - Appel à une route protégée par `RequirePermission`
   - Résultat : 200 (accès autorisé sans vérification de rôle)
   - (Note : dépend de la configuration du test — si l'admin bypass est dans RequirePermission)

2. **Membre avec permission**
   - Setup : Utilisateur membre avec role="captain", feature "team:delete"
   - Appel à la route
   - Résultat : 200 (permission accordée)

3. **Membre sans permission**
   - Setup : Utilisateur membre avec role="member", feature "team:delete"
   - Résultat : 403 `insufficient_permission`

4. **Non membre**
   - Setup : Utilisateur non membre de l'équipe
   - Résultat : 403 `not_team_member`

5. **Composition RequireTeamRole + RequirePermission - permission OK mais rôle insuffisant**
   - Setup : Utilisateur membre role="editor" mais `RequireTeamRole` avec `allowedRoles = ["captain"]`
   - Résultat : 403 de `RequireTeamRole` (bloqué avant même d'arriver à `RequirePermission`)

#### Fichier `internal/handlers/team_roles_test.go`

Tests des handlers de gestion des rôles via httptest.

Scénarios :

1. **GET /api/v1/teams/:id/members - Liste des membres**
   - Setup : Équipe avec 3 membres (1 captain, 1 editor, 1 member)
   - Appel par un des membres
   - Résultat : 200 + 3 membres avec leurs rôles
   - Vérifier les champs : user_id, role, joined_at

2. **GET /api/v1/teams/:id/members - Non membre**
   - Setup : Utilisateur non membre
   - Résultat : 403

3. **PUT /api/v1/teams/:id/members/:userId/role - Capitaine assigne Éditeur**
   - Setup : Capitaine connecté, cible = membre
   - Body : `{ "role": "editor" }`
   - Résultat : 200 + rôle changé
   - Vérifier que `FindByUserAndTeam` retourne "editor" après l'appel

4. **PUT /api/v1/teams/:id/members/:userId/role - Capitaine promet Cocapitaine**
   - Setup : Capitaine connecté, cible = editor, pas encore de cocapitaine
   - Body : `{ "role": "cocaptain" }`
   - Résultat : 200 + rôle changé

5. **PUT /api/v1/teams/:id/members/:userId/role - Déjà un cocapitaine**
   - Setup : Équipe a déjà un cocapitaine
   - Body : `{ "role": "cocaptain" }`
   - Résultat : 400 `team_already_has_cocaptain`

6. **PUT /api/v1/teams/:id/members/:userId/role - Non capitaine**
   - Setup : Utilisateur avec role="editor"
   - Résultat : 403 `insufficient_role`

7. **PUT /api/v1/teams/:id/members/:userId/role - Rétrograder un admin**
   - Setup : Cible est admin global (is_admin = true)
   - Résultat : 403 `cannot_downgrade_admin`

8. **PUT /api/v1/teams/:id/members/:userId/role - Rôle invalide**
   - Body : `{ "role": "superadmin" }`
   - Résultat : 400 `invalid_role`

9. **PUT /api/v1/teams/:id/members/:userId/role - Cible introuvable**
   - Body : `{ "role": "member" }` avec userId inexistant dans l'équipe
   - Résultat : 404 `member_not_found`

10. **POST /api/v1/teams/:id/transfer-captaincy - Transfert réussi**
    - Setup : Capitaine connecté, cible = editor
    - Body : `{ "user_id": targetID }`
    - Résultat : 200 + message, l'ancien capitaine devient cocapitaine, le nouveau devient capitaine
    - Vérifier `captain_id` de l'équipe mis à jour

11. **POST /api/v1/teams/:id/transfer-captaincy - Même utilisateur**
    - Body : `{ "user_id": currentCaptainID }`
    - Résultat : 400 `same_user`

12. **POST /api/v1/teams/:id/transfer-captaincy - Cible admin**
    - Setup : Cible a is_admin = true
    - Résultat : 403

13. **POST /api/v1/teams/:id/transfer-captaincy - Non capitaine**
    - Setup : Utilisateur avec role="cocaptain"
    - Résultat : 403 `insufficient_role`

#### Fichier `api/routes_test.go`

Ajouter des tests de base pour vérifier que les nouvelles routes sont enregistrées :
- `GET /api/v1/teams/1/members` sans auth → 401
- `PUT /api/v1/teams/1/members/2/role` sans auth → 401
- `POST /api/v1/teams/1/transfer-captaincy` sans auth → 401

#### Fichier `frontend/src/api/roles.test.ts`

Tests unitaires de l'API roles (mocking fetch).

Scénarios :
1. `listMembers` appelle GET avec le bon teamId
2. `updateMemberRole` appelle PUT avec le bon body et IDs
3. `transferCaptaincy` appelle POST avec le bon body

#### Fichier `frontend/src/components/MemberList.test.tsx` (optionnel)

Si les outils de test frontend sont configurés :
1. Affiche la liste des membres avec leurs badges de rôle
2. Le menu d'actions est visible uniquement pour le capitaine
3. Le bouton "Promouvoir cocapitaine" est absent si déjà un cocapitaine
4. Le bouton "Promouvoir cocapitaine" est absent pour les admins

**Partie 2 - Documentation :**

#### Fichier `docs/roles-permissions.md`

Documentation complète du système de rôles et permissions, couvrant :

1. **Vue d'ensemble** : Description du système de rôles d'équipe (Capitaine, Co-capitaine, Éditeur, Membre) et du rôle global Admin
2. **Rôles d'équipe** :
   - **Capitaine** : Tous les droits. Peut transférer son rôle. Ne peut pas être rétrogradé.
   - **Co-capitaine** : Mêmes droits que le Capitaine sauf suppression d'équipe. Ne peut pas promouvoir en co-capitaine ou capitaine.
   - **Éditeur** : Modification des zones. Modifications annulables par le Capitaine.
   - **Membre** : Accès aux zones privées, révision, duplication de templates.
3. **Rôle global Admin** :
   - Accès complet à toutes les fonctionnalités
   - Ne peut pas être rétrogradé
   - Bypasse toutes les vérifications de permissions d'équipe
4. **Contraintes** :
   - 1 co-capitaine maximum par équipe
   - Un admin ne peut pas être rétrogradé
   - Le capitaine ne peut être rétrogradé que par transfert
   - Par défaut, un nouveau membre est Membre
5. **Assignation des rôles** :
   - Le Capitaine peut assigner tous les rôles
   - Le Co-capitaine peut assigner les rôles Éditeur et Membre
   - Tableau des actions possibles par rôle
6. **Transfert de capitainerie** :
   - Le capitaine transfère son rôle → devient co-capitaine
   - Nécessite une confirmation
   - Les deux membres doivent être dans l'équipe
7. **Système de permissions** :
   - Chaque fonctionnalité de l'application est identifiée en base de données (table `permissions`)
   - Chaque rôle a des permissions associées (table `role_permissions`)
   - La vérification se fait à chaque appel API via middleware
   - Tableau de toutes les features et leurs rôles autorisés
8. **Endpoints API** :
   | Méthode | Path | Auth | Rôle | Permission | Description |
   |---------|------|------|------|------------|-------------|
   | GET | `/api/v1/teams/:id/members` | Oui | member | member:view | Liste des membres |
   | PUT | `/api/v1/teams/:id/members/:userId/role` | Oui | captain | member:role:assign | Assigner un rôle |
   | POST | `/api/v1/teams/:id/transfer-captaincy` | Oui | captain | member:role:transfer-captain | Transférer la capitainerie |
9. **Exemples curl** :
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
10. **Gestion des erreurs** : Tableau des codes d'erreur

| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | invalid_role | Rôle invalide |
| 400 | team_already_has_cocaptain | Déjà un cocapitaine |
| 400 | same_user | Transfert vers soi-même |
| 403 | not_team_member | Pas membre de l'équipe |
| 403 | insufficient_role | Rôle insuffisant |
| 403 | insufficient_permission | Permission insuffisante |
| 403 | cannot_downgrade_admin | Tentative de rétrograder un admin |
| 403 | cannot_downgrade_captain | Tentative de rétrograder le capitaine |
| 404 | member_not_found | Membre introuvable |
| 500 | internal_error | Erreur serveur |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/permission_test.go` | Créer | Tests d'intégration des repositories permissions |
| `internal/middleware/permission_test.go` | Créer | Tests du middleware de permission |
| `internal/handlers/team_roles_test.go` | Créer | Tests des handlers de gestion des rôles |
| `api/routes_test.go` | Modifier | Ajouter tests de base des nouvelles routes |
| `frontend/src/api/roles.test.ts` | Créer | Tests unitaires de l'API roles frontend |
| `frontend/src/components/MemberList.test.tsx` | Créer | Tests composant (optionnel) |
| `docs/roles-permissions.md` | Créer | Documentation du système de rôles et permissions |

### Contraintes techniques

- **Testcontainers** : Utiliser `testcontainers-go` avec l'image `postgres:16-alpine`. Même setup que `internal/models/setup_test.go`.
- **Migration** : Exécuter les migrations (`000001_initial_schema.up.sql` puis `000002_seed_permissions.up.sql`) avant les tests.
- **Seed dans les tests** : Les tests de permission ont besoin des données seedées. Exécuter les migrations SQL puis appeler `seed.SeedRolePermissions(db)` dans `TestMain` ou dans un setup de test.
- **Mock Auth** : Pour les tests des handlers, utiliser `gin.SetMode(gin.TestMode)` et `httptest.NewRecorder()`. Pour l'auth, utiliser un middleware mocké.
- **Isolation des tests** : Chaque test doit créer ses propres données (via les repositories). Utiliser `t.Run` pour les sous-tests.
- **Vitest** : Pour les tests frontend, utiliser `vitest`. Mock `fetch` global avec `vi.fn()`.
- **Conventions de test** : Suivre les patterns existants :
  - `t.Run("name", func(t *testing.T) { ... })` pour les sous-tests
  - `t.Helper()` pour les fonctions de setup
  - `t.Cleanup()` pour le nettoyage des ressources testcontainers

### Documentation

- `docs/roles-permissions.md` : Guide complet du système de rôles et permissions (créé par cette tâche)
- Mettre à jour `docs/stories/story-004.md` pour marquer les critères d'acceptation comme faits après exécution de cette tâche
