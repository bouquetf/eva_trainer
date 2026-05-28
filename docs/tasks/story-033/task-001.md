# Tâche #001 - Story #033 : Migration base de données + permissions

## Objectif
Créer les tables `lines_of_fire`, `template_lines_of_fire`, `line_of_fire_edit_history`, `line_of_fire_versions`, et les permissions associées.

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Aucune (tâche racine)
- Nécessaire pour : Tâche #002, Tâche #004, Tâche #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**M1 — Créer la table `lines_of_fire`**
- Créer une migration SQL `db/migrations/000020_create_lines_of_fire.up.sql`
- Colonnes :
  - `id SERIAL PRIMARY KEY`
  - `team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE NOT NULL`
  - `map_id INTEGER REFERENCES maps(id) ON DELETE CASCADE NOT NULL`
  - `name VARCHAR(250) NOT NULL`
  - `polygon_data JSONB NOT NULL` — stocke tous les cônes de cette ligne : `{"cones": [{"origin": {"x": 1, "y": 2}, "left_end": {"x": 3, "y": 4}, "right_end": {"x": 5, "y": 6}}]}`
  - `order_index INTEGER DEFAULT 0`
  - `created_at TIMESTAMP DEFAULT NOW()`
  - `updated_at TIMESTAMP DEFAULT NOW()`
- Index : `CREATE INDEX idx_lines_of_fire_team_map ON lines_of_fire(team_id, map_id)`

**M2 — Créer la table `template_lines_of_fire`**
- Créer une migration SQL `db/migrations/000021_create_template_lines_of_fire.up.sql`
- Colonnes :
  - `id SERIAL PRIMARY KEY`
  - `template_id INTEGER REFERENCES templates(id) ON DELETE CASCADE NOT NULL`
  - `polygon_data JSONB NOT NULL`
  - `name VARCHAR(250) NOT NULL`
  - `order_index INTEGER DEFAULT 0`
- Index : `CREATE INDEX idx_template_lines_of_fire_template ON template_lines_of_fire(template_id)`

**M3 — Créer la table `line_of_fire_edit_history`**
- Créer une migration SQL `db/migrations/000022_create_line_of_fire_edit_history.up.sql`
- Table parallèle à `edit_history` mais pour les lignes de tir :
  - `id SERIAL PRIMARY KEY`
  - `line_of_fire_id INTEGER REFERENCES lines_of_fire(id) ON DELETE CASCADE NOT NULL`
  - `editor_id INTEGER REFERENCES users(id) ON DELETE SET NULL`
  - `action VARCHAR(50) NOT NULL` — 'create', 'update', 'delete', 'add_cone', 'remove_cone', 'rename'
  - `old_data JSONB`
  - `new_data JSONB`
  - `timestamp TIMESTAMP DEFAULT NOW()`
- Index : `CREATE INDEX idx_lof_edit_history_line ON line_of_fire_edit_history(line_of_fire_id)`

**M4 — Créer la table `line_of_fire_versions`**
- Créer une migration SQL `db/migrations/000023_create_line_of_fire_versions.up.sql`
- Colonnes :
  - `id SERIAL PRIMARY KEY`
  - `line_of_fire_id INTEGER REFERENCES lines_of_fire(id) ON DELETE CASCADE NOT NULL`
  - `editor_id INTEGER REFERENCES users(id) ON DELETE SET NULL`
  - `version_data JSONB NOT NULL` — snapshot complet : `{"name": "...", "polygon_data": ...}`
  - `created_at TIMESTAMP DEFAULT NOW()`
- Index : `CREATE INDEX idx_lof_versions_line ON line_of_fire_versions(line_of_fire_id)`

**M5 — Ajouter les permissions `line_of_fire:*`**
- Créer une migration SQL `db/migrations/000024_add_line_of_fire_permissions.up.sql`
- Ajouter dans `permissions` les features suivantes :
  - `line_of_fire:view`
  - `line_of_fire:create`
  - `line_of_fire:edit`
  - `line_of_fire:delete`
- Ajouter dans `role_permissions` les assignments suivants :
  - Rôles `captain`, `cocaptain`, `editor` : toutes les 4 permissions
  - Rôle `member` : uniquement `line_of_fire:view`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000020_create_lines_of_fire.up.sql` | Créer | Table lines_of_fire |
| `db/migrations/000021_create_template_lines_of_fire.up.sql` | Créer | Table template_lines_of_fire |
| `db/migrations/000022_create_line_of_fire_edit_history.up.sql` | Créer | Table line_of_fire_edit_history |
| `db/migrations/000023_create_line_of_fire_versions.up.sql` | Créer | Table line_of_fire_versions |
| `db/migrations/000024_add_line_of_fire_permissions.up.sql` | Créer | Permissions pour les lignes de tir |
| `db/migrations/000020_create_lines_of_fire.down.sql` | Créer | Rollback |
| `db/migrations/000021_create_template_lines_of_fire.down.sql` | Créer | Rollback |
| `db/migrations/000022_create_line_of_fire_edit_history.down.sql` | Créer | Rollback |
| `db/migrations/000023_create_line_of_fire_versions.down.sql` | Créer | Rollback |
| `db/migrations/000024_add_line_of_fire_permissions.down.sql` | Créer | Rollback |
| `internal/models/setup_test.go` | Modifier | Ajouter les nouvelles migrations dans `migrationFiles` |

### Contraintes techniques
- **Base de données** : PostgreSQL 16. Chaque fichier `.up.sql` / `.down.sql` doit être présent.
- **Nommage** : Suivre le pattern `000XXX_description.up.sql` / `000XXX_description.down.sql`.
- **Rollback** : Chaque migration `.down.sql` doit annuler proprement les changements.
- **Tests** : Ajouter les nouvelles migrations dans `internal/models/setup_test.go` pour que les tests d'intégration puissent les utiliser.
- **Permissions** : Suivre le pattern exact des migrations existantes (`000012_add_zone_permissions.up.sql` pour référence).
