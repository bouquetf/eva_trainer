# Tâche #001 - Story #034 : Migration DB (global_roles, team_roles, permissions)

## Objectif
Créer la migration de base de données pour le stockage des rôles "archiviste" (global) et "lecteur" (par équipe), ainsi que les permissions `archive:read` et `archive:write`.

## Contexte
- Story #034 : `docs/stories/story-034.md`
- Spécification : `docs/specification/features/feedback-archives-matchs.md` (section 1)
- Dépend de : Aucune
- Nécessaire pour : Tâche #002

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer deux nouvelles tables et enregistrer les permissions dans la base de données.

**Table `global_roles`** : stocke les rôles globaux attribués aux utilisateurs.
- Un utilisateur peut avoir plusieurs rôles globaux.
- Clé primaire composite (user_id, role).
- CASCADE à la suppression de l'utilisateur.

**Table `team_roles`** : stocke les rôles attribués au niveau des équipes.
- Une équipe peut avoir plusieurs rôles.
- Clé primaire composite (team_id, role).
- CASCADE à la suppression de l'équipe.

**Permissions** : Ajouter `archive:read` et `archive:write` dans la table `permissions`.
- Ces permissions sont enregistrées mais la vérification d'accès se fait via le middleware `RequireArchiveAccess` (Tâche #002), pas via le système `HasPermission` existant.
- Ne PAS créer d'entrées dans `role_permissions` pour ces permissions.

**Cas nominaux :**
- Migration up : création des tables et insertion des permissions réussie
- Migration down : suppression des tables et des permissions réussie

**Cas limites :**
- Ré-exécution de la migration up : idempotence via `IF NOT EXISTS` / `ON CONFLICT DO NOTHING`
- Ré-exécution de la migration down : idempotence via `IF EXISTS`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000025_add_archive_roles.up.sql` | Créer | Migration up |
| `db/migrations/000025_add_archive_roles.down.sql` | Créer | Migration down |

### Signatures SQL

```sql
-- global_roles table
CREATE TABLE IF NOT EXISTS global_roles (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    PRIMARY KEY (user_id, role)
);

-- team_roles table
CREATE TABLE IF NOT EXISTS team_roles (
    team_id INT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    PRIMARY KEY (team_id, role)
);

-- permissions
INSERT INTO permissions (feature, description) VALUES
    ('archive:read', 'Consulter les archives'),
    ('archive:write', 'Gérer les archives (création, modification, suppression)')
ON CONFLICT (feature) DO NOTHING;
```

### Contraintes techniques
- **Convention** : Suivre le pattern des migrations existantes (numéro à 6 chiffres, fichiers `.up.sql` / `.down.sql`)
- **Idempotence** : Toutes les instructions DDL/DML doivent être idempotentes
- **Cascade** : Les clés étrangères doivent utiliser `ON DELETE CASCADE`

### Tests

#### Test manuel de vérification
- Exécuter `go run ./cmd/migrate up` et vérifier que les tables sont créées
- Exécuter `go run ./cmd/migrate down` et vérifier que le rollback fonctionne
- Vérifier les permissions dans la table `permissions`

### Documentation
- Mettre à jour `docs/roles-archives.md` avec le schéma des nouvelles tables (à créer dans la Tâche #006)
