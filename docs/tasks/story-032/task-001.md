# Tâche #001 - Story #032 : Migration base + refactoring modèle template

## Objectif
Appliquer les changements de schéma et de modèle Go : renommer `is_public` en `sync_active`, changer la FK `original_team_id` en ON DELETE CASCADE, supprimer les vérifications obsolètes, et renommer `FindAllPublic`.

## Contexte
- Story #032 : `docs/stories/story-032.md`
- Dépend de : Aucune (tâche racine)
- Nécessaire pour : Tâche #002, Tâche #003, Tâche #004

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Appliquer les 5 changements suivants sur le backend :

**M4 — Renommer la colonne `is_public` en `sync_active`**
- Créer une migration SQL `db/migrations/000018_rename_template_is_public.up.sql` qui renomme la colonne
- Mettre à jour le type `Template` dans `internal/models/template.go` : `IsPublic bool` → `SyncActive bool`
- Mettre à jour tous les appels à `is_public` dans les requêtes SQL du fichier `template.go`
- Mettre à jour l'interface `TemplateRepository` dans `internal/models/repository.go` si nécessaire
- Mettre à jour la création du template : `CreateFromMap` doit explicitement insérer `sync_active = true`
- Mettre à jour `FindByMapID` pour scanner `sync_active` au lieu de `is_public`
- Mettre à jour `FindAllTemplates` (ex `FindAllPublic`) pour scanner `sync_active`
- Mettre à jour tout référentiel à `is_public` dans les handlers, services, et les tests
- Mettre à jour l'interface TypeScript `Template` dans `frontend/src/api/templates.ts` : `is_public` → `sync_active`

**M6 — Changer FK `original_team_id` en ON DELETE CASCADE**
- Créer une migration SQL `db/migrations/000019_template_original_team_cascade.up.sql` qui :
  - Supprime la FK existante sur `original_team_id`
  - La recrée avec `ON DELETE CASCADE`
  - Change aussi `original_team_id` en `NOT NULL` (un template a toujours une équipe d'origine)

**M8 — Supprimer la vérification `is_public` dans `DuplicateToTeam`**
- Dans `DuplicateToTeam`, supprimer les lignes qui vérifient `!t.IsPublic` et retournent `ErrTemplateNotPublic`
- Tous les templates sont toujours duplicables
- Supprimer la sentinelle `ErrTemplateNotPublic` de `internal/models/errors.go` (si elle n'est plus utilisée ailleurs)

**M9 — Retirer le filtre `is_public = true` dans `FindTemplateZonesByMapID`**
- Modifier la requête SQL pour supprimer `AND is_public = true` (ou `AND sync_active = true`)
- Tous les templates sont trouvables quel que soit leur état de synchronisation

**M10 — Renommer `FindAllPublic` en `FindAllTemplates`**
- Renommer la méthode dans `template.go` et dans l'interface `TemplateRepository`
- Renommer le handler associé `ListPublicTemplates` en `ListAllTemplates`
- Mettre à jour la route `/api/v1/templates/public` (garder le même path)
- Mettre à jour l'appel frontend `listPublicTemplates` dans `frontend/src/api/templates.ts`
- Mettre à jour les tests et E2E

**Cas nominaux :**
- Un template créé a `sync_active = true`
- `FindAllTemplates()` liste tous les templates sans filtre
- `DuplicateToTeam` duplique tous les templates sans restriction
- `FindTemplateZonesByMapID` trouve les zones de tous les templates
- La suppression d'une équipe supprime ses templates (CASCADE)
- Les templates créés avant la migration ont `sync_active = true`

**Gestion d'erreurs :**
- Supprimer `ErrTemplateNotPublic` de `internal/models/errors.go` si non référencé ailleurs
- Si `ErrTemplateNotPublic` est référencé dans le handler, supprimer la branche de gestion correspondante

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/migrations/000018_rename_template_is_public.up.sql` | Créer | Migration : `ALTER TABLE templates RENAME COLUMN is_public TO sync_active` |
| `db/migrations/000019_template_original_team_cascade.up.sql` | Créer | Migration : changer FK ON DELETE SET NULL → ON DELETE CASCADE + NOT NULL |
| `internal/models/template.go` | Modifier | Renommer `IsPublic` → `SyncActive`, update all SQL, remove gate, remove filter, rename method |
| `internal/models/repository.go` | Modifier | Mettre à jour l'interface `TemplateRepository` |
| `internal/models/errors.go` | Modifier | Supprimer `ErrTemplateNotPublic` si inutilisé |
| `internal/handlers/template.go` | Modifier | Renommer `ListPublicTemplates` → `ListAllTemplates` |
| `api/routes.go` | Modifier | Mettre à jour le nom du handler si changé |
| `frontend/src/api/templates.ts` | Modifier | Renommer `is_public` → `sync_active` dans l'interface `Template` |
| `frontend/src/api/templates.test.ts` | Modifier | Mettre à jour les références à `is_public` |
| `internal/models/template_test.go` | Modifier | Mettre à jour `IsPublic` → `SyncActive`, `FindAllPublic` → `FindAllTemplates` |
| `internal/handlers/template_test.go` | Modifier | Mettre à jour les références à `is_public` et `FindAllPublic` |
| `frontend/e2e/templates/publish.spec.ts` | Modifier | Remplacer `is_public` par `sync_active` si référencé |
| `frontend/e2e/templates/duplication.spec.ts` | Modifier | Mettre à jour si nécessaire |

### Signatures

```go
// Template struct (internal/models/template.go)
type Template struct {
    ID               int       `json:"id"`
    OriginalTeamID   int       `json:"original_team_id"`
    OriginalTeamName string    `json:"original_team_name"`
    MapID            int       `json:"map_id"`
    SyncActive       bool      `json:"sync_active"`  // renommé
    MapName          string    `json:"map_name"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

// Interface (internal/models/repository.go)
type TemplateRepository interface {
    CreateFromMap(teamID, mapID int) (*Template, error)
    FindByMapID(mapID int) (*Template, error)
    FindAllTemplates() ([]*TemplateWithZones, error)      // renommé
    UpdateFromMap(templateID, teamID, mapID int) error
    Delete(templateID int) error
    DuplicateToTeam(teamID, templateID int) ([]*Zone, int, error)
    FindTemplateZonesByMapID(mapID int) ([]TemplateZone, error)
}
```

```typescript
// frontend/src/api/templates.ts
export interface Template {
  id: number
  original_team_id: number
  original_team_name: string
  map_id: number
  sync_active: boolean  // renommé
  map_name: string
  created_at: string
  updated_at: string
}
```

### Contraintes techniques
- **Base de données** : PostgreSQL 16 via migrations versionnées. Chaque fichier `.up.sql` doit être idempotent (`IF NOT EXISTS`, `IF EXISTS`).
- **Go** : Le projet utilise `database/sql` avec scanning manuel. Toutes les requêtes SQL sont en dur dans les méthodes du repository.
- **Tests** : Les tests d'intégration sont taggés `//go:build integration`. Les tests unitaires frontend utilisent Vitest.
- **Retro-compatibilité** : La migration `000018` doit être un pur `RENAME COLUMN` pour préserver les données existantes. Les templates existants auront `sync_active = true` (valeur de l'ancien `is_public`).

### Tests à implémenter / mettre à jour

#### Tests d'intégration Go
- **Fichier** : `internal/models/template_test.go`
- Vérifier que `CreateFromMap` retourne un template avec `SyncActive == true`
- Vérifier que `DuplicateToTeam` ne bloque aucun template (suppression du gate)
- Vérifier que `FindAllTemplates` liste tous les templates
- Vérifier que `FindTemplateZonesByMapID` trouve les zones même pour un template avec `sync_active = false`

#### Tests unitaires frontend
- **Fichier** : `frontend/src/api/templates.test.ts`
- Mettre à jour les `is_public` → `sync_active` dans les données mockées

### Documentation
- `docs/templates.md` : mentionner le renommage (sera complété dans la tâche documentation)

### Exemples

**Migration 000018 :**
```sql
ALTER TABLE templates RENAME COLUMN is_public TO sync_active;
```

**Migration 000019 :**
```sql
ALTER TABLE templates DROP CONSTRAINT IF EXISTS templates_original_team_id_fkey;
ALTER TABLE templates ADD CONSTRAINT templates_original_team_id_fkey
    FOREIGN KEY (original_team_id) REFERENCES teams(id) ON DELETE CASCADE;
ALTER TABLE templates ALTER COLUMN original_team_id SET NOT NULL;
```

**CreateFromMap modifié (INSERT explicite) :**
```sql
INSERT INTO templates (original_team_id, map_id, sync_active, created_at, updated_at)
VALUES ($1, $2, true, NOW(), NOW())
RETURNING id, original_team_id, map_id, sync_active, created_at, updated_at,
    (SELECT name FROM maps WHERE id = templates.map_id),
    (SELECT name FROM teams WHERE id = templates.original_team_id)
```
