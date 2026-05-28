# Tâche #004 - Story #033 : Intégration templates (publication, sync, duplication)

## Objectif
Mettre à jour le système de templates pour inclure les lignes de tir : création snapshot lors de la publication, synchronisation automatique, et duplication vers une autre équipe.

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Tâche #001 (migration DB) et Tâche #002 (modèle)
- Nécessaire pour : Tâche #003 (handler - sync), Tâche #005 (révision)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**1. Mettre à jour `TemplateRepository.CreateFromMap`**

Dans `internal/models/template.go`, étendre `CreateFromMap` pour :
- Après avoir inséré les `template_zones`, récupérer aussi les lignes de tir via `SELECT id, polygon_data, name, order_index FROM lines_of_fire WHERE team_id = $1 AND map_id = $2`
- Insérer chaque ligne dans `template_lines_of_fire` avec `INSERT INTO template_lines_of_fire (template_id, polygon_data, name, order_index)`

**2. Mettre à jour `TemplateRepository.UpdateFromMap`**

Dans `internal/models/template.go`, étendre `UpdateFromMap` pour :
- Supprimer les `template_lines_of_fire` existants (`DELETE FROM template_lines_of_fire WHERE template_id = $1`)
- Réinsérer les lignes de tir actuelles (même logique que CreateFromMap)

**3. Ajouter `SyncAfterLineOfFireChange` à `TemplateSyncService`**

Dans `internal/services/template_sync.go`, ajouter une méthode :
```go
func (s *TemplateSyncService) SyncAfterLineOfFireChange(teamID, mapID int) error
```
Cette méthode suit le même pattern que `SyncOnZoneChange` :
- Vérifier si un template existe pour cette map
- Si oui et si `sync_active = true`, appeler `UpdateFromMap`

**4. Mettre à jour `TemplateRepository.DuplicateToTeam`**

Étendre `DuplicateToTeam` pour dupliquer aussi les `template_lines_of_fire` :
- Après avoir créé les zones dupliquées, récupérer les `template_lines_of_fire`
- Pour chaque template line, créer une nouvelle `lines_of_fire` avec les mêmes données

**5. Ajouter `FindTemplateLinesOfFireByMapID`**

Ajouter une méthode à `TemplateRepository` :
```go
FindTemplateLinesOfFireByMapID(mapID int) ([]TemplateLineOfFire, error)
```
Requête sur `template_lines_of_fire` via JOIN avec `templates` sur `map_id`.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/template.go` | Modifier | CreateFromMap, UpdateFromMap, DuplicateToTeam, FindTemplateLinesOfFireByMapID |
| `internal/models/repository.go` | Modifier | Ajouter `FindTemplateLinesOfFireByMapID` à l'interface |
| `internal/services/template_sync.go` | Modifier | Ajouter `SyncAfterLineOfFireChange` |
| `internal/handlers/line_of_fire.go` | Modifier | Appeler `SyncAfterLineOfFireChange` dans les handlers |

### Signatures

```go
// TemplateRepository interface - nouvelles méthodes
FindTemplateLinesOfFireByMapID(mapID int) ([]TemplateLineOfFire, error)

// TemplateSyncService - nouvelle méthode
func (s *TemplateSyncService) SyncAfterLineOfFireChange(teamID, mapID int) error
```

### Tests

- Tester que `CreateFromMap` insère aussi les lignes de tir dans `template_lines_of_fire`
- Tester que `UpdateFromMap` remplace les lignes de tir du template
- Tester que `DuplicateToTeam` crée les lignes de tir chez l'équipe cible
- Tests d'intégration Go avec base PostgreSQL (tag `//go:build integration`)
