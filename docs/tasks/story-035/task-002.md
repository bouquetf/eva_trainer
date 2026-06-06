# Tâche #002 - Story #035 : Handlers API LigueLocale et Ligue

## Objectif
Implémenter les handlers HTTP pour le CRUD des ligues locales (`/api/v1/archive/leagues-locales`) et des ligues (`/api/v1/archive/ligues`), avec vérification des permissions (archiviste/lecteur).

## Contexte
- Story #035 : `docs/stories/story-035.md`
- Dépend de : Tâche #001 (modèles LigueLocale + Ligue)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Endpoints ligues locales :**

| Méthode | Path | Permission | Description |
|---------|------|-----------|-------------|
| GET | `/api/v1/archive/leagues-locales` | `read` | Liste filtrée par `?emplacement=` |
| GET | `/api/v1/archive/leagues-locales/:id` | `read` | Détail avec ligues associées |
| POST | `/api/v1/archive/leagues-locales` | `write` | Créer une ligue locale |
| PUT | `/api/v1/archive/leagues-locales/:id` | `write` | Modifier une ligue locale |
| DELETE | `/api/v1/archive/leagues-locales/:id` | `write` | Supprimer une ligue locale |
| POST | `/api/v1/archive/leagues-locales/:id/import-ligue` | `write` | Associer une ligue à la locale |

**Endpoints ligues (périodes de compétition) :**

| Méthode | Path | Permission | Description |
|---------|------|-----------|-------------|
| GET | `/api/v1/archive/ligues` | `read` | Liste paginée (`?page=&per_page=`) |
| GET | `/api/v1/archive/ligues/:id` | `read` | Détail d'une ligue |
| POST | `/api/v1/archive/ligues` | `write` | Créer une ligue (optionnel : associer toutes les locales) |
| PUT | `/api/v1/archive/ligues/:id` | `write` | Modifier une ligue |
| DELETE | `/api/v1/archive/ligues/:id` | `write` | Supprimer une ligue (bloqué si BO associés) |
| POST | `/api/v1/archive/ligues/:id/refresh` | `write` | Associer les nouvelles locales à la ligue |

**Endpoints admin ligues :**

| Méthode | Path | Permission | Description |
|---------|------|-----------|-------------|
| GET | `/admin/ligues` | `read` | Liste paginée |
| POST | `/admin/ligues` | `write` | Créer une ligue (associe toutes les locales) |
| PUT | `/admin/ligues/:id` | `write` | Modifier une ligue |
| DELETE | `/admin/ligues/:id` | `write` | Supprimer une ligue |

**Cas nominaux :**
- GET `/api/v1/archive/leagues-locales` → `200` avec `{ "ligues_locales": [...] }`
- GET `/api/v1/archive/leagues-locales/:id` → `200` avec `{ "ligue_locale": {...}, "ligues": [...] }`
- POST `/api/v1/archive/leagues-locales` → `201` avec `{ "ligue_locale": {...} }`
- POST `/api/v1/archive/leagues-locales/:id/import-ligue` → `200` avec `{ "message": "ligue associée" }`
- GET `/api/v1/archive/ligues?page=1&per_page=20` → `200` avec `{ "items": [...], "total": N, "page": 1, "per_page": 20 }`
- POST `/api/v1/archive/ligues` → `201` avec `{ "ligue": {...}, "locales_associated": N }`
- DELETE `/api/v1/archive/ligues/:id` → `200` avec `{ "message": "ligue deleted" }`

**Cas limites :**
- Modification d'une ligue : toujours autorisée (pas de garde "déjà commencée")
- Suppression d'une ligue avec BO associés → `409` avec message "cannot delete league with associated BOs"
- Import d'une ligue déjà associée → succès silencieux (ON CONFLICT DO NOTHING côté repo)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Créer/Modifier | Handlers ligues locales + ligues |
| `internal/handlers/admin_ligue.go` | Créer | Handlers admin ligues |
| `api/routes.go` | Modifier | Ajouter les routes |
| `internal/handlers/archive_test.go` | Créer | Tests d'intégration |

### Signatures

```go
// Dans internal/handlers/archive.go

type ArchiveHandler struct {
    ligueLocaleRepo models.LigueLocaleRepository
    ligueRepo       models.LigueRepository
    splitRepo       models.SplitRepository
    logger          services.Logger
}

type createLigueLocaleRequest struct {
    Nom         string `json:"nom" binding:"required"`
    Emplacement string `json:"emplacement" binding:"required"`
}

type createLigueRequest struct {
    DateDebut           string `json:"date_debut" binding:"required"`
    DateFin             string `json:"date_fin" binding:"required"`
    SplitID             int    `json:"split_id" binding:"required"`
    AssociateAllLocales bool   `json:"associate_all_locales"`
}

func (h *ArchiveHandler) ListLiguesLocales(c *gin.Context)
func (h *ArchiveHandler) GetLigueLocale(c *gin.Context)
func (h *ArchiveHandler) CreateLigueLocale(c *gin.Context)
func (h *ArchiveHandler) UpdateLigueLocale(c *gin.Context)
func (h *ArchiveHandler) DeleteLigueLocale(c *gin.Context)
func (h *ArchiveHandler) ImportLigue(c *gin.Context)
func (h *ArchiveHandler) ListLigues(c *gin.Context)
func (h *ArchiveHandler) GetLigue(c *gin.Context)
func (h *ArchiveHandler) CreateLigue(c *gin.Context)
func (h *ArchiveHandler) UpdateLigue(c *gin.Context)
func (h *ArchiveHandler) DeleteLigue(c *gin.Context)
func (h *ArchiveHandler) RefreshLigue(c *gin.Context)
```

### Réponse GetLigueLocale

```json
{
  "ligue_locale": { "id": 1, "nom": "Grenoble", "emplacement": "Grenoble", "created_at": "...", "updated_at": "..." },
  "ligues": [
    { "id": 1, "split_name": "Split 1", "date_debut": "2026-01-01", "date_fin": "2026-06-30" }
  ]
}
```

### Contraintes techniques
- **Pattern handler** : Suivre le pattern de `internal/handlers/split.go` (binding JSON, error handling avec `errors.Is`, réponses avec `apierror`)
- **Validation des dates** : Parser les strings `"YYYY-MM-DD"` avec `time.Parse("2006-01-02", ...)`. Si le parsing échoue → 400.
- **Pagination** : `page` et `per_page` en query params avec defaults (page=1, per_page=20). `per_page` max 100.
- **ImportLigue** : Appelle `AssociateLigue(localeID, ligueID)` via `ligueLocaleRepo`. Pas de vérification préalable (ON CONFLICT DO NOTHING gère les doublons).
- **RefreshLigue** : Appelle `AssociateAllLocales(ligueID)` — n'ajoute que les locales pas encore associées.
- **Middleware existant** : Utiliser `middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"|"write")` déjà en place.

### Tests à implémenter

#### Tests d'intégration archive
- **Fichier** : `internal/handlers/archive_test.go`

1. **ListLiguesLocales_Empty** → `200` avec `{ "ligues_locales": [] }`
2. **ListLiguesLocales_WithData** → Créer 2 locales → list → 2 résultats
3. **ListLiguesLocales_FilterByEmplacement** → Filtrer → 1 résultat
4. **ListLiguesLocales_NoMatch** → Filtrer inexistant → 0 résultats
5. **GetLigueLocale_Success** → Créer → GET /:id → 200
6. **GetLigueLocale_NotFound** → ID inexistant → 404
7. **GetLigueLocale_WithLigue** → Créer locale + ligue + association → détail inclut la ligue
8. **CreateLigueLocale_Success** → POST → 201
9. **CreateLigueLocale_NomVide** → POST nom vide → 400
10. **UpdateLigueLocale_Success** → Créer → PUT → 200
11. **UpdateLigueLocale_NotFound** → PUT ID 99999 → 404
12. **DeleteLigueLocale_Success** → Créer → DELETE → 200, GET → 404
13. **DeleteLigueLocale_NotFound** → DELETE ID 99999 → 404
14. **ImportLigue_Success** → Créer locale + ligue → import → 200
15. **ImportLigue_AlreadyAssociated** → Importer 2 fois → 200 (idempotent)
16. **ListLigues_Empty** → 200 avec pagination vide
17. **ListLigues_WithPagination** → Créer ligues → pagination correcte
18. **CreateLigue_Success** → POST → 201 avec `locales_associated`
19. **CreateLigue_WithAssociateAll** → POST avec associate_all_locales=true → locales associées
20. **UpdateLigue_Success** → Créer → PUT → 200
21. **DeleteLigue_Success** → Créer → DELETE → 200
22. **DeleteLigue_NotFound** → DELETE ID 99999 → 404
23. **RefreshLigue_AddsNewLocales** → Créer ligue → créer locale → refresh → locale associée
