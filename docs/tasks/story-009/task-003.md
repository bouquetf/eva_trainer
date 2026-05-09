# Tâche #003 - Story #009 : Permissions zones + Migration seed

## Objectif

Ajouter les permissions nécessaires pour la gestion des zones dans la base de données, et créer le cleanup automatique des locks expirés.

## Contexte

- Story #009 : `docs/stories/story-009.md` - Zone Creation and Editing with Polygons
- Dépend de : Story #004 (permissions existantes, seed 000005)
- Nécessaire pour : Tâche #002 (middleware RequirePermission)
- Ordre d'exécution : 1ère (parallélisable avec Tâche #001)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Migration des permissions zones

Créer une migration `000012_add_zone_permissions` qui ajoute les permissions liées aux zones dans la table `permissions` et les assigne aux rôles appropriés dans `role_permissions`.

Permissions à créer :

| Feature | Description | Rôles |
|---------|-------------|-------|
| `zone:view` | Voir les zones d'une équipe | member, editor, cocaptain, captain |
| `zone:create` | Créer une nouvelle zone | editor, cocaptain, captain |
| `zone:edit` | Modifier une zone existante (polygones, points, nom) | editor, cocaptain, captain |
| `zone:delete` | Supprimer une zone | editor, cocaptain, captain |

**Important** : L'admin a implicitement toutes les permissions via le middleware `RequireAdmin`. Pas besoin d'assigner à l'admin dans role_permissions.

#### Partie 2 - Cleanup automatique des locks expirés

Ajouter un mécanisme de cleanup des locks expirés :

1. Créer `internal/services/lock_cleanup.go` avec un service qui :
   - Nettoie périodiquement les locks expirés via `EditLockRepository.CleanupExpired()`
   - Utilise un ticker (intervalle : 5 minutes)
   - Peut être démarré/arrêté proprement (Start/Stop)

2. Intégrer ce service dans `cmd/server/main.go` (au démarrage du serveur).

**Cas nominaux :**
- Lock avec `expires_at < NOW()` → supprimé au prochain cleanup
- Cleanup s'exécute toutes les 5 minutes
- Peut aussi être appelé à chaque tentative d'acquisition de lock

**Cas limites :**
- Aucun lock expiré → cleanup sans effet (0 lignes supprimées)
- Le cleanup ne doit pas bloquer le serveur (goroutine séparée)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000012_add_zone_permissions.up.sql` | Créer | INSERT des permissions zones dans permissions et role_permissions |
| `db/migrations/000012_add_zone_permissions.down.sql` | Créer | DELETE des permissions zones |
| `internal/services/lock_cleanup.go` | Créer | Service de nettoyage périodique des locks expirés |
| `cmd/server/main.go` | Modifier | Démarrer le service de cleanup au boot |

### Signatures

```go
// internal/services/lock_cleanup.go
package services

type LockCleanupService struct {
	lockRepo models.EditLockRepository
	interval time.Duration
	done     chan struct{}
}

func NewLockCleanupService(lockRepo models.EditLockRepository, interval time.Duration) *LockCleanupService

// Start démarre la goroutine de nettoyage périodique
func (s *LockCleanupService) Start()

// Stop arrête proprement la goroutine de nettoyage
func (s *LockCleanupService) Stop()

// CleanupNow exécute un nettoyage immédiat (bloquant)
func (s *LockCleanupService) CleanupNow() (int64, error)
```

```sql
-- db/migrations/000012_add_zone_permissions.up.sql
-- Insérer les permissions de zone
INSERT INTO permissions (feature, description) VALUES
    ('zone:view', 'View zones of a team'),
    ('zone:create', 'Create new zones'),
    ('zone:edit', 'Modify existing zones (polygons, points, name)'),
    ('zone:delete', 'Delete zones');

-- Assigner aux rôles
INSERT INTO role_permissions (role, permission_id)
SELECT 'member', id FROM permissions WHERE feature = 'zone:view';

INSERT INTO role_permissions (role, permission_id)
SELECT 'editor', id FROM permissions WHERE feature IN ('zone:view', 'zone:create', 'zone:edit', 'zone:delete');

INSERT INTO role_permissions (role, permission_id)
SELECT 'cocaptain', id FROM permissions WHERE feature IN ('zone:view', 'zone:create', 'zone:edit', 'zone:delete');

INSERT INTO role_permissions (role, permission_id)
SELECT 'captain', id FROM permissions WHERE feature IN ('zone:view', 'zone:create', 'zone:edit', 'zone:delete');
```

### Contraintes techniques

- **Migration** : Vérifier la structure exacte de `000005_seed_permissions.up.sql` pour suivre le même format. Utiliser des sous-requêtes pour lier les permission_ids.
- **LockCleanupService** : Utiliser `time.NewTicker` pour le déclenchement périodique. La goroutine doit gérer le signal de arrêt via le channel `done`.
- **Intégration serveur** : Dans `cmd/server/main.go`, créer le service après l'initialisation de la DB et avant le démarrage du serveur HTTP. Stocker le service pour pouvoir l'arrêter proprement (ex: via signal handling).

### Tests à implémenter

Les tests seront créés dans la Tâche #006.

### Documentation

Aucune documentation à créer pour cette tâche.
