# Tâche #001 - Story #037 : Migration BDD — Divs + BOs/Matchs

## Objectif
Créer les tables `divs`, `equipe_eva_div`, `bos` et `matchs` en base de données, permettant de gérer les divisions de ligue, l'assignation d'équipes, et le calcul de classement à partir des BOs.

## Contexte
- Story #037 : `docs/stories/story-037.md`
- Spécification : `docs/specification/features/feedback-archives-matchs.md` (sections 4, 5, 7)
- Dépend de : Story #035, Story #036 (migrations déjà appliquées : `ligue_locale`, `ligue_locale_association`, `eva_teams`)
- Nécessaire pour : Tâche #002, Tâche #003, Tâche #004, Tâche #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer la migration `000031_add_archive_divs_and_bos` avec 4 tables.

**Cas nominaux :**
- Migration up : création réussie des 4 tables
- Migration down : suppression des 4 tables dans l'ordre inverse des dépendances

**Cas limites :**
- Idempotence via `IF NOT EXISTS` pour les CREATE TABLE
- Idempotence via `IF EXISTS` pour le DROP TABLE

### Structure des tables

#### Table `divs`
| Champ | Type | Contraintes |
|-------|------|-------------|
| `id` | `SERIAL` | `PRIMARY KEY` |
| `ligue_id` | `INT` | `NOT NULL`, `REFERENCES ligue(id) ON DELETE CASCADE` |
| `nom` | `VARCHAR(255)` | `NOT NULL` |

- Index sur `(ligue_id, nom)` pour unicité + recherche (pas de doublon de nom de div dans une même ligue)

#### Table `equipe_eva_div`
| Champ | Type | Contraintes |
|-------|------|-------------|
| `equipe_eva_id` | `INT` | `NOT NULL`, `REFERENCES eva_teams(id) ON DELETE CASCADE` |
| `div_id` | `INT` | `NOT NULL`, `REFERENCES divs(id) ON DELETE CASCADE` |
| `ligue_id` | `INT` | `NOT NULL`, `REFERENCES ligue(id) ON DELETE CASCADE` |

- `PRIMARY KEY (equipe_eva_id, div_id)`
- `UNIQUE (equipe_eva_id, ligue_id)` — garantit au niveau DB qu'une équipe ne peut être que dans une seule div par ligue
- Index sur `div_id` pour les requêtes de listing par div
- La colonne `ligue_id` est dénormalisée (provient de `divs.ligue_id`) pour permettre la contrainte UNIQUE. L'application est responsable de passer la bonne valeur à l'INSERT.

#### Table `bos`
| Champ | Type | Contraintes |
|-------|------|-------------|
| `id` | `SERIAL` | `PRIMARY KEY` |
| `ligue_id` | `INT` | `NOT NULL`, `REFERENCES ligue(id) ON DELETE CASCADE` |
| `equipe1_id` | `INT` | `NOT NULL`, `REFERENCES eva_teams(id) ON DELETE RESTRICT` |
| `equipe2_id` | `INT` | `NOT NULL`, `REFERENCES eva_teams(id) ON DELETE RESTRICT` |
| `vainqueur_id` | `INT` | `NULL`, `REFERENCES eva_teams(id) ON DELETE SET NULL` |
| `format` | `VARCHAR(3)` | `NOT NULL DEFAULT 'BO3'`, `CHECK (format IN ('BO3', 'BO5'))` |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` |

- Index sur `ligue_id` pour les requêtes de classement
- `ON DELETE RESTRICT` pour `equipe1_id` / `equipe2_id` : on ne peut pas supprimer une équipe eva qui a des BOs
- `ON DELETE SET NULL` pour `vainqueur_id` : si l'équipe gagnante est supprimée, le résultat devient indéterminé

#### Table `matchs`
| Champ | Type | Contraintes |
|-------|------|-------------|
| `id` | `SERIAL` | `PRIMARY KEY` |
| `bo_id` | `INT` | `NOT NULL`, `REFERENCES bos(id) ON DELETE CASCADE` |
| `map_id` | `INT` | `NOT NULL`, `REFERENCES maps(id) ON DELETE RESTRICT` |
| `score_equipe1` | `INT` | `NOT NULL`, `CHECK (score_equipe1 >= 0 AND score_equipe1 <= 100)` |
| `score_equipe2` | `INT` | `NOT NULL`, `CHECK (score_equipe2 >= 0 AND score_equipe2 <= 100)` |
| `vainqueur_id` | `INT` | `NULL`, `REFERENCES eva_teams(id) ON DELETE SET NULL` |
| `commentaires` | `TEXT` | `NULL` |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` |

- Index sur `bo_id` pour les requêtes de détail BO
- Pas de CHECK supplémentaire sur `vainqueur_id` (calculé applicativement)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000031_add_archive_divs_and_bos.up.sql` | Créer | Migration up |
| `db/migrations/000031_add_archive_divs_and_bos.down.sql` | Créer | Migration down |

### SQL

```sql
-- 000031_add_archive_divs_and_bos.up.sql

CREATE TABLE IF NOT EXISTS divs (
    id SERIAL PRIMARY KEY,
    ligue_id INT NOT NULL REFERENCES ligue(id) ON DELETE CASCADE,
    nom VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_divs_ligue_nom ON divs(ligue_id, nom);

CREATE TABLE IF NOT EXISTS equipe_eva_div (
    equipe_eva_id INT NOT NULL REFERENCES eva_teams(id) ON DELETE CASCADE,
    div_id INT NOT NULL REFERENCES divs(id) ON DELETE CASCADE,
    ligue_id INT NOT NULL REFERENCES ligue(id) ON DELETE CASCADE,
    PRIMARY KEY (equipe_eva_id, div_id),
    UNIQUE (equipe_eva_id, ligue_id)
);

CREATE INDEX IF NOT EXISTS idx_equipe_eva_div_div_id ON equipe_eva_div(div_id);
CREATE INDEX IF NOT EXISTS idx_equipe_eva_div_ligue_id ON equipe_eva_div(ligue_id);

CREATE TABLE IF NOT EXISTS bos (
    id SERIAL PRIMARY KEY,
    ligue_id INT NOT NULL REFERENCES ligue(id) ON DELETE CASCADE,
    equipe1_id INT NOT NULL REFERENCES eva_teams(id) ON DELETE RESTRICT,
    equipe2_id INT NOT NULL REFERENCES eva_teams(id) ON DELETE RESTRICT,
    vainqueur_id INT NULL REFERENCES eva_teams(id) ON DELETE SET NULL,
    format VARCHAR(3) NOT NULL DEFAULT 'BO3' CHECK (format IN ('BO3', 'BO5')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bos_ligue_id ON bos(ligue_id);

CREATE TABLE IF NOT EXISTS matchs (
    id SERIAL PRIMARY KEY,
    bo_id INT NOT NULL REFERENCES bos(id) ON DELETE CASCADE,
    map_id INT NOT NULL REFERENCES maps(id) ON DELETE RESTRICT,
    score_equipe1 INT NOT NULL CHECK (score_equipe1 >= 0 AND score_equipe1 <= 100),
    score_equipe2 INT NOT NULL CHECK (score_equipe2 >= 0 AND score_equipe2 <= 100),
    vainqueur_id INT NULL REFERENCES eva_teams(id) ON DELETE SET NULL,
    commentaires TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_matchs_bo_id ON matchs(bo_id);
```

```sql
-- 000031_add_archive_divs_and_bos.down.sql
DROP TABLE IF EXISTS matchs CASCADE;
DROP TABLE IF EXISTS bos CASCADE;
DROP TABLE IF EXISTS equipe_eva_div CASCADE;
DROP INDEX IF EXISTS idx_equipe_eva_div_ligue_id;
DROP TABLE IF EXISTS divs CASCADE;
```

**L'ordre du down est important** : on doit supprimer les tables dans l'ordre inverse des dépendances FK (matchs → bos, bos → divs/equipe_eva_div, equipe_eva_div → divs, divs).

### Contraintes techniques
- **Convention migrations** : Suivre le pattern existant (fichiers `.up.sql` / `.down.sql`, numéro `000031`)
- **Idempotence** : Toutes les instructions DDL doivent être idempotentes (`IF NOT EXISTS`, `IF EXISTS`)
- **`ON DELETE CASCADE`** sur `divs` et `equipe_eva_div` : si on supprime une div, les assignations sont automatiquement supprimées
- **`ON DELETE RESTRICT`** sur `bos.equipe1_id` / `bos.equipe2_id` : empêche la suppression d'une équipe eva référencée
- **`ON DELETE SET NULL`** sur `vainqueur_id` : si l'équipe gagnante est supprimée, le vainqueur devient NULL
- **UNIQUE(equipe_eva_id, ligue_id)** : garantie DB qu'une équipe ne peut être que dans une seule div par ligue. La colonne `ligue_id` est dénormalisée — l'application doit passer la bonne valeur, et la FK `REFERENCES ligue(id)` assure l'intégrité référentielle.

### Tests

Aucun test automatisé pour cette tâche (purement DDL).

**Vérification manuelle :**
```bash
cd db/migrations
# Vérifier la syntaxe SQL (optionnel si pas d'outil dédié)
```

### Documentation

Documentation mise à jour dans `docs/archives-divisions.md` (créé dans la Tâche #004) :
- Schéma des tables `divs`, `equipe_eva_div`
- Schéma des tables `bos`, `matchs` (partagé avec Story #038)
