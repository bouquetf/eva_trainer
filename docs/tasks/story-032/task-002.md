# Tâche #002 - Story #032 : Endpoint unpublish + re-publication avec remplacement

## Objectif
Ajouter un endpoint `POST /teams/:id/maps/:mapId/unpublish` pour dépublier une carte, et modifier l'endpoint publish pour permettre la re-publication avec remplacement du template existant.

## Contexte
- Story #032 : `docs/stories/story-032.md`
- Dépend de : Tâche #001 (migration + refactoring `sync_active`)
- Nécessaire pour : Tâche #004 (frontend)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### A — Nouvel endpoint "unpublish"

Créer un endpoint qui permet au capitaine/cocapitaine de dépublier une carte :

- `POST /api/v1/teams/:id/maps/:mapId/unpublish`
- Auth : `RequireTeamRole(cocaptain, captain)` + permission `zone:publish`
- Comportement : passe `sync_active = false` sur le template existant
- Garde le template inchangé (mêmes template_zones), seule la synchro est désactivée
- Si la carte n'est pas publiée (pas de template) → retourner 404 (template not found)
- Retourner le template mis à jour

#### B — Modification de l'endpoint publish existant

L'endpoint `POST /api/v1/teams/:id/maps/:mapId/publish` doit être modifié :

- Si aucun template n'existe pour cette map → comportement actuel : `CreateFromMap`
- Si un template existe déjà → remplacer ses zones via `UpdateFromMap` ET passer `sync_active = true`
- Le handler doit pouvoir distinguer entre "création" et "remplacement" pour que le frontend sache quelle popup afficher
- Retourner un flag `{ created: true/false, template: {...} }` dans la réponse JSON

**Cas nominaux :**
- Carte jamais publiée → publier → template créé, `sync_active = true`, retour `{ created: true, template }`
- Carte déjà publiée → publier → template remplacé, `sync_active = true`, retour `{ created: false, template }`
- Carte publiée → dépublier → `sync_active = false`, template conservé, retour template mis à jour
- Carte non publiée → dépublier → 404
- Utilisateur non autorisé → 403

**Cas limites :**
- Dépublier une carte déjà dépubliée → retourner 404 (pas de template)
- Publier avec un template existant qui a `sync_active = false` → remplacer et passer `sync_active = true`

**Gestion d'erreurs :**
- Template non trouvé (unpublish) → `ErrTemplateNotFound` → 404
- Permissions insuffisantes → 403 via middleware

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/template.go` | Modifier | Ajouter `UnpublishMap()`, modifier `PublishMap()` |
| `internal/models/template.go` | Modifier | Ajouter `SetSyncStatus(templateID int, active bool) error` |
| `internal/models/repository.go` | Modifier | Ajouter `SetSyncStatus(templateID int, active bool) error` à l'interface |
| `api/routes.go` | Modifier | Ajouter la route POST unpublish |
| `internal/handlers/template_test.go` | Modifier | Ajouter tests unpublish + re-publication |
| `frontend/src/api/templates.ts` | Modifier | Ajouter fonction `unpublishMap()` |

### Signatures

```go
// internal/models/repository.go — ajouter à TemplateRepository
SetSyncStatus(templateID int, active bool) error

// internal/models/template.go — nouvelle méthode
func (r *templateRepo) SetSyncStatus(templateID int, active bool) error {
    _, err := r.db.Exec(
        `UPDATE templates SET sync_active = $1, updated_at = NOW() WHERE id = $2`,
        active, templateID,
    )
    return err
}

// internal/handlers/template.go
func (h *TemplateHandler) UnpublishMap(c *gin.Context)

// PublishMap modifié — retourne un flag created
func (h *TemplateHandler) PublishMap(c *gin.Context)
// Réponse: { "created": true/false, "template": {...} }
```

```typescript
// frontend/src/api/templates.ts — ajouter
export async function unpublishMap(teamId: number, mapId: number): Promise<Template> {
  const api = getApiClient()
  return api.post(`/api/v1/teams/${teamId}/maps/${mapId}/unpublish`)
}

// Modifier le type de retour de publishMap
export async function publishMap(teamId: number, mapId: number): Promise<{ created: boolean; template: Template }> {
  const api = getApiClient()
  return api.post(`/api/v1/teams/${teamId}/maps/${mapId}/publish`)
}
```

### Contraintes techniques
- **Pattern handler** : Suivre le même pattern que les handlers existants dans `template.go` (utiliser `apierror.RespondWithSuccess`, `apierror.RespondWithError`)
- **Transaction** : `PublishMap` n'a pas besoin de transaction pour le remplacement car `UpdateFromMap` gère déjà son propre DELETE + INSERT
- **Routes** : Ajouter la route dans `setupTeamRoutes` dans `api/routes.go`, à côté de la route publish existante
- **Permissions** : Utiliser la permission `zone:publish` déjà existante (identique à publish)

### Tests à implémenter

#### Tests d'intégration handler (template_test.go)
- Tester que `PublishMap` crée un template sur une carte non publiée → `created: true`
- Tester que `PublishMap` remplace le template sur une carte déjà publiée → `created: false`, zones remplacées
- Tester que `UnpublishMap` passe `sync_active = false`
- Tester que `UnpublishMap` sur une carte non publiée retourne 404
- Tester qu'un utilisateur non autorisé reçoit 403

### Exemples d'API

**Publication initiale :**
```
POST /api/v1/teams/1/maps/5/publish
→ 201 { "created": true, "template": { "id": 1, "sync_active": true, ... } }
```

**Re-publication :**
```
POST /api/v1/teams/1/maps/5/publish
→ 200 { "created": false, "template": { "id": 1, "sync_active": true, ... } }
```

**Unpublish :**
```
POST /api/v1/teams/1/maps/5/unpublish
→ 200 { "template": { "id": 1, "sync_active": false, ... } }
```

**Unpublish sans template :**
```
POST /api/v1/teams/1/maps/5/unpublish
→ 404 { "error": "template_not_found", "message": "template not found" }
```
