# Tâche #001 - Story #004 : Modèle de permissions et repository (back-end)

## Objectif

Créer le modèle de données Permission et RolePermission, leurs repositories, et les données de seed qui définissent l'ensemble des fonctionnalités de l'application et leur association aux rôles.

## Contexte

- Story #004 : `docs/stories/story-004.md` - Team Roles and Permissions
- Dépend de : Story #000 (migration DB : tables `permissions` et `role_permissions` existantes), Story #003 (modèle TeamMember avec constantes de rôles)
- Nécessaire pour : Tâche #002 (endpoints de gestion des rôles), Tâche #003 (middleware de vérification)
- Ordre d'exécution : 1ère

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Modèle Permission :**
Créer le struct `Permission` dans `internal/models/permission.go`. La table `permissions` existe déjà dans la migration `000001_initial_schema.up.sql`. Définir le repository avec les méthodes suivantes :

- `FindAll() ([]*Permission, error)` : Récupère toutes les permissions.
- `FindByID(id int) (*Permission, error)` : Récupère une permission par son ID.
- `FindByFeature(feature string) (*Permission, error)` : Récupère une permission par son identifiant de fonctionnalité.
- `Create(feature string, description string) (*Permission, error)` : Crée une nouvelle permission.

**Partie 2 - Modèle RolePermission :**
Créer le struct `RolePermission` dans `internal/models/role_permission.go`. La table `role_permissions` existe déjà dans la migration. Définir le repository avec :

- `FindByRole(role string) ([]*Permission, error)` : Récupère toutes les permissions d'un rôle donné.
- `FindByPermissionID(permissionID int) ([]*RolePermission, error)` : Récupère tous les rôles associés à une permission.
- `AssignPermissionToRole(role string, permissionID int) error` : Associe une permission à un rôle.
- `RemovePermissionFromRole(role string, permissionID int) error` : Retire l'association.
- `RoleHasPermission(role string, permissionID int) (bool, error)` : Vérifie si un rôle a une permission spécifique.

**Partie 3 - Helper `HasPermission` :**
Créer une fonction utilitaire `HasPermission(role string, feature string, rolePermRepo RolePermissionRepository, permRepo PermissionRepository) (bool, error)` dans `internal/models/permission.go` qui :
1. Cherche la permission par `feature` dans `permRepo.FindByFeature`
2. Si trouvée, vérifie via `rolePermRepo.RoleHasPermission(role, permissionID)`
3. Retourne `true`/`false`

**Partie 4 - Admin bypass :**
La fonction `HasPermission` doit retourner `true` si le rôle est `admin` (les admins ont toutes les permissions). L'admin est un rôle global, pas un rôle d'équipe.

**Partie 5 - Seed data (approche mixte) :**
Les données de seed sont réparties en deux parties :

**5a - Migration SQL** (`db/migrations/000002_seed_permissions.up.sql`) : Contient les INSERT des permissions (la table `permissions`). Chaque feature est une ligne avec son ID, feature et description. Le down supprime ces insertions.

```sql
-- 000002_seed_permissions.up.sql
INSERT INTO permissions (feature, description) VALUES
    ('team:create', 'Créer une équipe'),
    ('team:edit', 'Modifier le nom/infos de l''équipe'),
    ('team:delete', 'Supprimer l''équipe'),
    ('team:leave', 'Quitter l''équipe'),
    ('team:invite', 'Générer des invitations'),
    ('team:requests:view', 'Voir les demandes d''adhésion'),
    ('team:requests:manage', 'Accepter/refuser les demandes'),
    ('member:view', 'Voir la liste des membres'),
    ('member:kick', 'Retirer un membre'),
    ('member:role:assign', 'Assigner les rôles Éditeur et Membre'),
    ('member:role:promote-cocaptain', 'Promouvoir un membre en cocapitaine'),
    ('member:role:transfer-captain', 'Transférer le rôle de capitaine'),
    ('zone:view', 'Voir les zones de l''équipe'),
    ('zone:create', 'Créer des zones'),
    ('zone:edit', 'Modifier des zones'),
    ('zone:delete', 'Supprimer des zones'),
    ('zone:publish', 'Rendre des zones publiques'),
    ('zone:lock', 'Verrouiller/déverrouiller une carte'),
    ('review:start', 'Lancer une révision'),
    ('template:duplicate', 'Dupliquer un template public'),
    ('edit_history:view', 'Voir l''historique des modifications'),
    ('edit_history:rollback', 'Effectuer un rollback'),
    ('map:add', 'Ajouter des cartes (admin uniquement)')
ON CONFLICT (feature) DO NOTHING;
```

```sql
-- 000002_seed_permissions.down.sql
DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions);
DELETE FROM permissions;
```

**5b - Fonction Go** (`db/seed/permissions.go`) : Contient l'association des permissions aux rôles (table `role_permissions`). Appelée au démarrage de l'application dans `cmd/server/main.go` après les migrations.

```go
// db/seed/permissions.go
package seed

// SeedRolePermissions insère les associations rôle ↔ permission.
// Utilise ON CONFLICT DO NOTHING pour être idempotent.
// Les permissions doivent déjà exister en base (via migration).
func SeedRolePermissions(db *sql.DB) error
```

Règles d'association :

| Feature | admin | captain | cocaptain | editor | member |
|---------|-------|---------|-----------|--------|--------|
| `team:create` | ✓ | - | - | - | ✓ |
| `team:edit` | ✓ | ✓ | ✓ | - | - |
| `team:delete` | ✓ | ✓ | - | - | - |
| `team:leave` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `team:invite` | ✓ | ✓ | ✓ | - | - |
| `team:requests:view` | ✓ | ✓ | ✓ | - | - |
| `team:requests:manage` | ✓ | ✓ | ✓ | - | - |
| `member:view` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `member:kick` | ✓ | ✓ | - | - | - |
| `member:role:assign` | ✓ | ✓ | ✓ | - | - |
| `member:role:promote-cocaptain` | ✓ | ✓ | - | - | - |
| `member:role:transfer-captain` | ✓ | ✓ | - | - | - |
| `zone:view` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `zone:create` | ✓ | ✓ | ✓ | ✓ | - |
| `zone:edit` | ✓ | ✓ | ✓ | ✓ | - |
| `zone:delete` | ✓ | ✓ | ✓ | ✓ | - |
| `zone:publish` | ✓ | ✓ | ✓ | - | - |
| `zone:lock` | ✓ | ✓ | ✓ | - | - |
| `review:start` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `template:duplicate` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `edit_history:view` | ✓ | ✓ | ✓ | - | ✓ |
| `edit_history:rollback` | ✓ | ✓ | - | - | - |
| `map:add` | ✓ | - | - | - | - |

Note : Le rôle `admin` n'est pas stocké dans `role_permissions` (l'admin bypass est géré dans `HasPermission`). Les associations pour `admin` ne sont pas insérées.

**Cas nominaux :**
- Récupération de toutes les permissions → liste complète retournée
- Vérification `HasPermission("captain", "team:delete", ...)` → `true`
- Vérification `HasPermission("editor", "team:delete", ...)` → `false`
- Admin bypass : `HasPermission("admin", "map:add", ...)` → `true`

**Cas limites :**
- Permission inexistante (feature inconnue) → `false` (pas d'erreur)
- Rôle inexistant → `false` (pas d'erreur)
- Seed exécuté deux fois → ne pas dupliquer les données (`INSERT ... ON CONFLICT DO NOTHING` ou vérification préalable)

**Gestion d'erreurs :**
- Erreur DB → propagation de l'erreur au caller
- Feature `""` (vide) → retourne `false, nil`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/permission.go` | Créer | Struct Permission, PermissionRepository, HasPermission |
| `internal/models/role_permission.go` | Créer | Struct RolePermission, RolePermissionRepository |
| `internal/models/repository.go` | Modifier | Ajouter les interfaces PermissionRepository et RolePermissionRepository |
| `internal/models/errors.go` | Modifier | Ajouter les erreurs liées aux permissions |
| `db/migrations/000002_seed_permissions.up.sql` | Créer | Migration SQL : insertion des permissions |
| `db/migrations/000002_seed_permissions.down.sql` | Créer | Migration SQL : rollback des permissions |
| `db/seed/permissions.go` | Créer | Fonction SeedRolePermissions (associations rôle-permission) |
| `cmd/server/main.go` | Modifier | Appeler SeedRolePermissions au démarrage (après les migrations) |

### Signatures

```go
// internal/models/permission.go
package models

type Permission struct {
    ID          int       `json:"id"`
    Feature     string    `json:"feature"`
    Description string    `json:"description"`
}

type PermissionRepository interface {
    FindAll() ([]*Permission, error)
    FindByID(id int) (*Permission, error)
    FindByFeature(feature string) (*Permission, error)
    Create(feature, description string) (*Permission, error)
}

type permissionRepo struct {
    db *sql.DB
}

func NewPermissionRepository(db *sql.DB) PermissionRepository {
    return &permissionRepo{db: db}
}

// HasPermission vérifie si un rôle a accès à une fonctionnalité.
// Le rôle "admin" a toujours toutes les permissions.
// Retourne false (sans erreur) si la feature n'existe pas en base.
func HasPermission(
    role string,
    feature string,
    rolePermRepo RolePermissionRepository,
    permRepo PermissionRepository,
) (bool, error)
```

```go
// internal/models/role_permission.go
package models

type RolePermission struct {
    Role         string `json:"role"`
    PermissionID int    `json:"permission_id"`
}

type RolePermissionRepository interface {
    FindByRole(role string) ([]*Permission, error)
    FindByPermissionID(permissionID int) ([]*RolePermission, error)
    AssignPermissionToRole(role string, permissionID int) error
    RemovePermissionFromRole(role string, permissionID int) error
    RoleHasPermission(role string, permissionID int) (bool, error)
}

type rolePermissionRepo struct {
    db *sql.DB
}

func NewRolePermissionRepository(db *sql.DB) RolePermissionRepository {
    return &rolePermissionRepo{db: db}
}
```

```go
// db/seed/permissions.go
package seed

import "database/sql"

// SeedRolePermissions insère les associations rôle ↔ permission.
// Les permissions doivent déjà exister en base (migration 000002).
// Utilise INSERT ... ON CONFLICT DO NOTHING pour être idempotent.
func SeedRolePermissions(db *sql.DB) error
```

### Contraintes techniques

- **Package naming** : `models`. Import path : `eva_trainer/internal/models`.
- **Migration SQL** : Le fichier `db/migrations/000002_seed_permissions.up.sql` insère les lignes dans `permissions`. Le `.down.sql` les supprime. Les migrations sont exécutées automatiquement par le système de migration existant.
- **Fonction Go seed** : `SeedRolePermissions` dans `db/seed/permissions.go` insère les associations dans `role_permissions`. Idempotent via `ON CONFLICT (role, permission_id) DO NOTHING`. Appelée dans `cmd/server/main.go` après les migrations.
- **Admin bypass** : L'admin étant un rôle global (colonne `is_admin` de la table `users`), le helper `HasPermission` traite `"admin"` comme ayant toutes les permissions sans requêter la base. Les associations admin ne sont pas stockées dans `role_permissions`.
- **Contexte DB** : Les repositories prennent `*sql.DB` dans leur constructeur. Le package `pkg/database` fournit la connexion via `database.DB()`.
- **Seed au démarrage** : Dans `cmd/server/main.go`, appeler `seed.SeedRolePermissions(db)` après les migrations et avant le démarrage du serveur Gin. La migration SQL est déjà exécutée par le système de migration.
- **Transactions** : `SeedRolePermissions` utilise une transaction pour garantir l'atomicité des associations.
- **Erreurs** : Ajouter dans `internal/models/errors.go` :
  ```go
  var (
      ErrPermissionNotFound = errors.New("permission not found")
      // Les erreurs existantes sont conservées
  )
  ```

### Tests à implémenter

Les tests seront créés dans la Tâche #005. Aucun test spécifique à cette tâche — les tests des repositories seront intégrés avec les tests fonctionnels.

### Documentation

Aucune documentation à créer pour cette tâche.
