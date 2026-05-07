# Tâche #002 - Story #000 : Base de données PostgreSQL et migrations

## Objectif

Configurer la connexion à PostgreSQL, installer le driver, et mettre en place les migrations avec `golang-migrate/migrate` en se basant sur le schéma de la spécification.

## Contexte

- Story #000 : `docs/stories/story-000.md` - Project Initialization and Scaffolding
- Dépend de : Tâche #001 (structure backend Go)
- Nécessaire pour : Tâche #003 (Docker), Tâche #005 (Tests), Tâche #006 (README)
- Ordre d'exécution : 2ème (après #001)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Mettre en place la couche d'accès à la base de données PostgreSQL avec les migrations gérées par `golang-migrate/migrate`.

**Cas nominaux :**
- Package `pkg/database` exporte une fonction `Connect` qui retourne un pool de connexions `*sql.DB`
- Les paramètres de connexion sont lus depuis les variables d'environnement (ou `.env`)
- Les migrations sont gérées avec `github.com/golang-migrate/migrate/v4`
- Les fichiers de migration (up/down) sont dans `db/migrations/` et créent toutes les tables définies dans `docs/specification/database_schema.md`
- Une fonction `database.RunMigrations()` applique les migrations au démarrage (appelée depuis le code ou depuis une commande séparée)

**Cas limites :**
- Si `DATABASE_URL` est définie, l'utiliser préférentiellement (format PostgreSQL standard : `postgres://user:pass@host:port/dbname?sslmode=disable`)
- Sinon, utiliser les variables individuelles : `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- Si aucune variable de connexion n'est trouvée, `Connect` doit retourner une erreur descriptive
- La connexion doit supporter SSL mode configurable via `DB_SSLMODE` (défaut `disable` pour le dev local)

**Gestion d'erreurs :**
- Échec de connexion → retourner une erreur avec le message du driver PostgreSQL
- Timeout de connexion → timeout configurable via `DB_TIMEOUT` (défaut 10 secondes)
- Migration déjà appliquée → ignorée silencieusement (golang-migrate gère cela via sa table `schema_migrations`)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `go.mod` | Modifier | Ajouter `github.com/jackc/pgx/v5`, `github.com/golang-migrate/migrate/v4` |
| `pkg/database/database.go` | Créer | Fonction de connexion et configuration |
| `pkg/database/config.go` | Créer | Types et fonctions de configuration |
| `pkg/database/migrations.go` | Créer | Fonction `RunMigrations` utilisant golang-migrate |
| `db/migrations/000001_initial_schema.up.sql` | Créer | Migration up : création des tables |
| `db/migrations/000001_initial_schema.down.sql` | Créer | Migration down : suppression des tables |
| `go.sum` | Généré | Verrouillage des dépendances |

### Signatures

```go
// pkg/database/database.go
package database

type Config struct {
    DatabaseURL string
    Host        string
    Port        string
    User        string
    Password    string
    DBName      string
    SSLMode     string
    Timeout     time.Duration
}

// LoadConfigFromEnv lit les variables d'environnement et retourne Config.
// Priorité : DATABASE_URL > variables individuelles (DB_HOST, etc.).
// Valeurs par défaut : Host=localhost, Port=5432, User=postgres, Password=postgres,
// DBName=eva_trainer, SSLMode=disable, Timeout=10s.
func LoadConfigFromEnv() Config

// Connect établit une connexion à PostgreSQL avec la config donnée.
// Retourne un pool de connexions *sql.DB prêt à l'emploi.
// L'appelant doit fermer le pool avec db.Close().
func Connect(cfg Config) (*sql.DB, error)

// HealthCheck vérifie que la base de données est joignable avec un ping.
func HealthCheck(db *sql.DB) error
```

```go
// pkg/database/migrations.go
package database

// RunMigrations applique les migrations golang-migrate depuis db/migrations/.
// Utilise la source "file://db/migrations" et la base de données configurée.
// Retourne une erreur si les migrations échouent.
func RunMigrations(db *sql.DB) error
```

### Contraintes techniques

- **Version Go** : 1.26
- **Driver PostgreSQL** : Utiliser `github.com/jackc/pgx/v5` avec le wrapper `pgx/stdlib` pour compatibilité `database/sql`
- **Migrations** : Utiliser `github.com/golang-migrate/migrate/v4` :
  - Driver source : `migrate/file` (fichiers SQL)
  - Driver database : `migrate/pgx` ou `migrate/postgres`
  - Format des fichiers : `{version}_{title}.up.sql` et `{version}_{title}.down.sql`
  - La première migration est `000001_initial_schema`
- **Connexion** : Utiliser `database/sql` standard + pgx comme driver
- **Pool** : Configurer un pool avec max 25 connexions ouvertes, max 5 connexions inactives, durée de vie max 5 minutes
- **Ne PAS** connecter la base de données au démarrage du serveur dans cette tâche. La connexion sera intégrée dans une story ultérieure (qui fera évoluer le endpoint `/health` en conséquence).
- **Ne PAS** modifier `cmd/server/main.go` ou `api/routes.go`
- **Note :** Les migrations sont exécutables via `cmd/migrate/main.go` (créé dans la tâche #005) ou via la CLI golang-migrate

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `pkg/database/config_test.go`
- Scénario 1 : `LoadConfigFromEnv` utilise DATABASE_URL si présente
  - Données : `DATABASE_URL=postgres://user:pass@host:5432/db`
  - Résultat attendu : `cfg.DatabaseURL == "postgres://user:pass@host:5432/db"`
- Scénario 2 : `LoadConfigFromEnv` utilise les vars individuelles si DATABASE_URL absente
  - Données : `DB_HOST=myhost`, `DB_PORT=5433`, `DB_USER=admin`
  - Résultat attendu : `cfg.Host == "myhost"`, `cfg.Port == "5433"`, `cfg.User == "admin"`
- Scénario 3 : Valeurs par défaut quand aucune variable n'est définie
  - Données : Aucune variable d'environnement
  - Résultat attendu : `cfg.Host == "localhost"`, `cfg.Port == "5432"`, `cfg.DBName == "eva_trainer"`

#### Tests d'intégration
- **Fichier** : `pkg/database/database_test.go`
- Scénario 1 : Connexion avec DATABASE_URL valide
  - Données : DATABASE_URL pointant vers une instance PostgreSQL de test
  - Résultat attendu : `*sql.DB` non nul, `HealthCheck` retourne nil
- Scénario 2 : Connexion avec une URL invalide
  - Données : `DATABASE_URL=postgres://invalid:invalid@localhost:9999/nonexistent`
  - Résultat attendu : Erreur retournée (connexion impossible)

**Note :** Les tests d'intégration (connexion DB réelle, validation des migrations) et les tests E2E (docker compose) sont traités dans la tâche #005.

### Fichiers de migration

**`db/migrations/000001_initial_schema.up.sql` :**

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    discord_id VARCHAR(255) UNIQUE NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(250) UNIQUE NOT NULL,
    captain_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS team_members (
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('captain', 'cocaptain', 'editor', 'member')),
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);

CREATE TABLE IF NOT EXISTS maps (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    webp_blob BYTEA,
    added_at TIMESTAMP DEFAULT NOW(),
    obsolete_of INTEGER REFERENCES maps(id) ON DELETE SET NULL,
    is_obsolete BOOLEAN DEFAULT FALSE,
    current_version_id INTEGER REFERENCES maps(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS splits (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    map_names TEXT[] NOT NULL,
    start_date DATE,
    end_date DATE
);

CREATE TABLE IF NOT EXISTS zones (
    id SERIAL PRIMARY KEY,
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
    map_id INTEGER REFERENCES maps(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS zone_polygons (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER REFERENCES zones(id) ON DELETE CASCADE,
    polygon_data JSONB NOT NULL,
    order_index INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS zone_names (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER REFERENCES zones(id) ON DELETE CASCADE,
    name VARCHAR(250) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS templates (
    id SERIAL PRIMARY KEY,
    original_team_id INTEGER REFERENCES teams(id) ON DELETE SET NULL,
    map_id INTEGER REFERENCES maps(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS template_zones (
    id SERIAL PRIMARY KEY,
    template_id INTEGER REFERENCES templates(id) ON DELETE CASCADE,
    polygon_data JSONB NOT NULL,
    name VARCHAR(250) NOT NULL,
    order_index INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS edit_locks (
    id SERIAL PRIMARY KEY,
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
    map_id INTEGER REFERENCES maps(id) ON DELETE CASCADE,
    editor_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    locked_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP DEFAULT NOW() + INTERVAL '15 minutes',
    UNIQUE(team_id, map_id)
);

CREATE TABLE IF NOT EXISTS edit_history (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER REFERENCES zones(id) ON DELETE CASCADE,
    editor_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    old_data JSONB,
    new_data JSONB,
    timestamp TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    feature VARCHAR(255) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role VARCHAR(50) NOT NULL,
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role, permission_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_zones_team_map ON zones(team_id, map_id);
CREATE INDEX IF NOT EXISTS idx_zone_polygons_zone ON zone_polygons(zone_id);
CREATE INDEX IF NOT EXISTS idx_zone_names_zone ON zone_names(zone_id);
CREATE INDEX IF NOT EXISTS idx_edit_locks_team_map ON edit_locks(team_id, map_id);
CREATE INDEX IF NOT EXISTS idx_edit_history_zone ON edit_history(zone_id);
CREATE INDEX IF NOT EXISTS idx_template_zones_template ON template_zones(template_id);
```

**`db/migrations/000001_initial_schema.down.sql` :**

```sql
DROP INDEX IF EXISTS idx_template_zones_template;
DROP INDEX IF EXISTS idx_edit_history_zone;
DROP INDEX IF EXISTS idx_edit_locks_team_map;
DROP INDEX IF EXISTS idx_zone_names_zone;
DROP INDEX IF EXISTS idx_zone_polygons_zone;
DROP INDEX IF EXISTS idx_zones_team_map;

DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS edit_history;
DROP TABLE IF EXISTS edit_locks;
DROP TABLE IF EXISTS template_zones;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS zone_names;
DROP TABLE IF EXISTS zone_polygons;
DROP TABLE IF EXISTS zones;
DROP TABLE IF EXISTS splits;
DROP TABLE IF EXISTS maps;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;
```
