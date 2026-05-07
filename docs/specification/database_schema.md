# Proposition de schéma de base de données - Eva Trainer

## Tables principales

### users
- `id` SERIAL PRIMARY KEY
- `discord_id` VARCHAR(255) UNIQUE NOT NULL
- `is_admin` BOOLEAN DEFAULT FALSE
- `created_at` TIMESTAMP DEFAULT NOW()
- `updated_at` TIMESTAMP DEFAULT NOW()

### teams
- `id` SERIAL PRIMARY KEY
- `name` VARCHAR(250) UNIQUE NOT NULL
- `captain_id` INTEGER REFERENCES users(id) ON DELETE SET NULL
- `created_at` TIMESTAMP DEFAULT NOW()
- `updated_at` TIMESTAMP DEFAULT NOW()

### team_members
- `team_id` INTEGER REFERENCES teams(id) ON DELETE CASCADE
- `user_id` INTEGER REFERENCES users(id) ON DELETE CASCADE
- `role` VARCHAR(50) NOT NULL CHECK (role IN ('captain', 'cocaptain', 'editor', 'member'))
- `joined_at` TIMESTAMP DEFAULT NOW()
- PRIMARY KEY (team_id, user_id)

### maps
- `id` SERIAL PRIMARY KEY
- `name` VARCHAR(255) NOT NULL
- `webp_blob` BYTEA
- `added_at` TIMESTAMP DEFAULT NOW()
- `obsolete_of` INTEGER REFERENCES maps(id) ON DELETE SET NULL
- `is_obsolete` BOOLEAN DEFAULT FALSE
- `current_version_id` INTEGER REFERENCES maps(id) ON DELETE SET NULL

### splits
- `id` SERIAL PRIMARY KEY
- `name` VARCHAR(255) NOT NULL
- `map_names` TEXT[] NOT NULL -- Liste des noms de cartes "eva" jouables dans ce split (pas d'IDs, les noms correspondent aux noms dans la table maps)
- `start_date` DATE
- `end_date` DATE

### zones
- `id` SERIAL PRIMARY KEY
- `team_id` INTEGER REFERENCES teams(id) ON DELETE CASCADE
- `map_id` INTEGER REFERENCES maps(id) ON DELETE CASCADE
- `created_at` TIMESTAMP DEFAULT NOW()
- `updated_at` TIMESTAMP DEFAULT NOW()

### zone_polygons
- `id` SERIAL PRIMARY KEY
- `zone_id` INTEGER REFERENCES zones(id) ON DELETE CASCADE
- `polygon_data` JSONB NOT NULL -- Stocke les coordonnées du polygone
- `order_index` INTEGER DEFAULT 0
- `created_at` TIMESTAMP DEFAULT NOW()

### zone_names
- `id` SERIAL PRIMARY KEY
- `zone_id` INTEGER REFERENCES zones(id) ON DELETE CASCADE
- `name` VARCHAR(250) NOT NULL
- `created_at` TIMESTAMP DEFAULT NOW()

### templates
- `id` SERIAL PRIMARY KEY
- `original_team_id` INTEGER REFERENCES teams(id) ON DELETE SET NULL
- `map_id` INTEGER REFERENCES maps(id) ON DELETE CASCADE
- `created_at` TIMESTAMP DEFAULT NOW()
- `updated_at` TIMESTAMP DEFAULT NOW()

### template_zones
- `id` SERIAL PRIMARY KEY
- `template_id` INTEGER REFERENCES templates(id) ON DELETE CASCADE
- `polygon_data` JSONB NOT NULL
- `name` VARCHAR(250) NOT NULL
- `order_index` INTEGER DEFAULT 0

### edit_locks
- `id` SERIAL PRIMARY KEY
- `team_id` INTEGER REFERENCES teams(id) ON DELETE CASCADE
- `map_id` INTEGER REFERENCES maps(id) ON DELETE CASCADE
- `editor_id` INTEGER REFERENCES users(id) ON DELETE CASCADE
- `locked_at` TIMESTAMP DEFAULT NOW()
- `expires_at` TIMESTAMP DEFAULT NOW() + INTERVAL '15 minutes'
- UNIQUE(team_id, map_id)

### edit_history
- `id` SERIAL PRIMARY KEY
- `zone_id` INTEGER REFERENCES zones(id) ON DELETE CASCADE
- `editor_id` INTEGER REFERENCES users(id) ON DELETE SET NULL
- `action` VARCHAR(50) NOT NULL -- 'create', 'update', 'delete', 'add_point', 'remove_point', 'rename'
- `old_data` JSONB -- Pour 'rename' : {old_name: '...'}, pour 'update' : {old_polygon: ...}
- `new_data` JSONB -- Pour 'rename' : {new_name: '...'}, pour 'update' : {new_polygon: ...}
- `timestamp` TIMESTAMP DEFAULT NOW()

-- Note : Les changements de nom de zone sont trackés avec action='rename'.

### permissions
- `id` SERIAL PRIMARY KEY
- `feature` VARCHAR(255) UNIQUE NOT NULL -- identifiant de la fonctionnalité
- `description` TEXT

### role_permissions
- `role` VARCHAR(50) NOT NULL -- 'admin', 'captain', 'cocaptain', 'editor', 'member'
- `permission_id` INTEGER REFERENCES permissions(id) ON DELETE CASCADE
- PRIMARY KEY (role, permission_id)

## Index recommandés
- `idx_zones_team_map` ON zones(team_id, map_id)
- `idx_zone_polygons_zone` ON zone_polygons(zone_id)
- `idx_zone_names_zone` ON zone_names(zone_id)
- `idx_edit_locks_team_map` ON edit_locks(team_id, map_id)
- `idx_edit_history_zone` ON edit_history(zone_id)
- `idx_template_zones_template` ON template_zones(template_id)

## Notes
- Les zones privées sont supprimées si l'équipe est supprimée (ON DELETE CASCADE)
- Si une équipe est supprimée sans transfert de capitaine, les zones deviennent des templates (original_team_id reste pour l'admin)
- Le verrou expire automatiquement après 15 minutes (à gérer via tâche cron ou vérification à l'accès)
- L'historique est conservé même si l'utilisateur est supprimé (editor_id peut être NULL)
