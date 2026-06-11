# Tâche #002 - Story #043 : Handlers API pour les notes personnelles sur les divs

## Objectif
Créer les trois endpoints REST pour gérer les notes personnelles sur une div, avec middleware d'authentification, et les intégrer dans les routes.

## Contexte
- Story #043 : `docs/stories/story-043.md`
- Dépend de : Tâche #001 (modèle DivNote + repository)
- Nécessaire pour : Tâche #004 (intégration frontend)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Trois endpoints pour les notes personnelles : récupération, création/mise à jour (upsert), suppression.

L'utilisateur est identifié par le token JWT (via `middleware.MustGetUser(c)`). La note est toujours personnelle : on utilise `user.ID` pour toutes les opérations, pas d'ID utilisateur dans l'URL.

**Cas nominaux :**
- GET `/divisions/:id/note` → retourne la note de l'utilisateur connecté pour cette div (ou 404 si pas de note)
- PUT `/divisions/:id/note` → crée ou met à jour la note (body : `{ "content_md": "..." }`)
- DELETE `/divisions/:id/note` → supprime la note de l'utilisateur connecté

**Cas limites :**
- PUT avec `content_md` vide → autorisé (stocke une chaîne vide)
- GET sans note existante → 404 (note pas encore créée) — le frontend gère l'affichage "vide"
- DELETE sans note existante → 404

**Gestion d'erreurs :**
- Div inexistante → 404 `ErrDivNotFound`
- Non authentifié → 401 (géré par `RequireAuth`)
- Permission insuffisante → 403 (géré par `RequireArchiveAccess`)
- Erreur interne → 500

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Modifier | Ajouter 3 méthodes à ArchiveHandler |
| `api/routes.go` | Modifier | Enregistrer 3 routes |
| `internal/handlers/archive_test.go` | Modifier | Tests d'intégration |

### Signatures

```go
// Dans ArchiveHandler — nouvelles méthodes
func (h *ArchiveHandler) GetMyDivNote(c *gin.Context)
func (h *ArchiveHandler) UpsertMyDivNote(c *gin.Context)
func (h *ArchiveHandler) DeleteMyDivNote(c *gin.Context)

// Request body pour Upsert
type upsertDivNoteRequest struct {
    ContentMD string `json:"content_md"`
}
```

### Comportement détaillé des handlers

#### `GetMyDivNote`
1. Extraire `:id` de l'URL → divID
2. Vérifier que la div existe (optionnel, `divRepo.FindByID` — simple vérification)
3. Récupérer l'utilisateur connecté via `middleware.MustGetUser(c)`
4. Appeler `h.divNoteRepo.FindByDivAndUser(divID, user.ID)`
5. Si `ErrDivNoteNotFound` → 404
6. Sinon → 200 avec `{ "note": { div_id, user_id, content_md, created_at, updated_at } }`

#### `UpsertMyDivNote`
1. Extraire `:id` → divID, binder le body JSON → `upsertDivNoteRequest`
2. Récupérer user connecté
3. Appeler `h.divNoteRepo.Upsert(divID, user.ID, req.ContentMD)`
4. → 200 avec `{ "note": { ... } }`

#### `DeleteMyDivNote`
1. Extraire `:id` → divID
2. Récupérer user connecté
3. Appeler `h.divNoteRepo.DeleteByDivAndUser(divID, user.ID)`
4. Si `ErrDivNoteNotFound` → 404
5. Sinon → 200 avec `{ "message": "Note supprimée" }`

### Routes (dans `api/routes.go`)

Ajouter après les routes `/divisions/:id/replay` (après ligne 673) :

```go
archive.GET("/divisions/:id/note",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.GetMyDivNote,
)
archive.PUT("/divisions/:id/note",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.UpsertMyDivNote,
)
archive.DELETE("/divisions/:id/note",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.DeleteMyDivNote,
)
```

⚠️ **Note** : Les endpoints utilisent `RequireArchiveAccess("read")` car l'utilisateur écrit sa propre note personnelle — il n'a pas besoin du droit d'écriture sur les archives.

### Contraintes techniques
- **Handler struct modifié** : Ajouter `divNoteRepo models.DivNoteRepository` à `ArchiveHandler`
- **Constructor modifié** : Ajouter le paramètre `divNoteRepo models.DivNoteRepository` à `NewArchiveHandler` et l'assigner
- **Appels au constructeur** : Mettre à jour les deux appels :
  - `api/routes.go` ligne 529 — ajouter le nouveau repo
  - `internal/handlers/archive_test.go` ligne 69 — ajouter le mock

### Tests à implémenter

#### Tests d'intégration
- **Fichier** : `internal/handlers/archive_test.go`
- Scénario 1 : `GET /divisions/:id/note` — pas de note → 404
- Scénario 2 : `PUT /divisions/:id/note` — upsert d'une note → 200 + `note.content_md`
- Scénario 3 : `GET /divisions/:id/note` — note existante → 200 + `note.content_md`
- Scénario 4 : `PUT /divisions/:id/note` — mise à jour d'une note existante → 200 + nouveau contenu
- Scénario 5 : `DELETE /divisions/:id/note` → 200 + `message`
- Scénario 6 : `DELETE /divisions/:id/note` — note inexistante → 404
- Scénario 7 : `PUT /divisions/:id/note` — sans auth → 401
- Scénario 8 : `PUT /divisions/:id/note` — sans permission lecture archive → 403

Utiliser le même pattern que les tests existants dans `archive_test.go` :
- `setupArchiveHandlerTest` ou équivalent
- `mockDivNoteRepo` avec `gomock` si le projet utilise gomock, ou implémentation manuelle
