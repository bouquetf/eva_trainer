# Tâche #002 - Story #038 : Handlers CRUD BO + routes API

## Objectif
Créer les handlers HTTP pour le CRUD des BO (Best Of) avec bans, picks et matchs, et enregistrer les routes API correspondantes.

## Contexte
- Story #038 : BO, matchs, bans et picks
- Dépend de : Tâche #001 (Modèles Go Match, Ban, Pick + BORepository étendu)
- Nécessaire pour : Tâche #003 (Frontend)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Handlers à ajouter à `ArchiveHandler` :**

1. **CreateBO** — `POST /api/v1/archive/bos`
   - Reçoit un JSON avec : `ligue_id`, `equipe1_id`, `equipe2_id`, `format` ("BO3" ou "BO5"), `bans` (array), `picks` (array), `matchs` (array)
   - Crée le BO en base
   - Crée tous les bans en batch
   - Crée tous les picks en batch
   - Crée tous les matchs en batch (avec calcul automatique des vainqueurs)
   - Recalcule le vainqueur du BO
   - Retourne le BO complet avec bans, picks, matchs

2. **GetBO** — `GET /api/v1/archive/bos/:id`
   - Retourne le BO détaillé avec bans, picks, matchs

3. **ListBOs** — `GET /api/v1/archive/bos?ligue_id=`
   - Retourne la liste des BOs pour une ligue
   - Chaque BO inclut les noms des équipes (pas seulement les IDs) pour l'affichage
   - Optionnellement inclure bans/picks/matchs si `?details=true`

4. **UpdateBO** — `PUT /api/v1/archive/bos/:id`
   - Remplace l'ensemble des bans/picks/matchs du BO
   - Supprime les anciens bans/picks/matchs, insère les nouveaux
   - Met à jour le format si fourni
   - Recalcule le vainqueur du BO
   - Retourne le BO complet mis à jour

5. **DeleteBO** — `DELETE /api/v1/archive/bos/:id`
   - Supprime le BO (cascade sur bans/picks/matchs)
   - Vérifie que le BO existe

6. **GetLigueMaps** — `GET /api/v1/archive/ligues/:id/maps`
   - Retourne les cartes (maps) du split associé à la ligue
   - Utile pour le frontend (dropdown de sélection de carte filtré par split)
   - Utilise `splitRepo.FindMapsBySplitID(splitID)` (déjà existant)
   - `splitRepo` est déjà dans `ArchiveHandler`

**Dépendances supplémentaires à injecter dans `ArchiveHandler` :**
- `boRepo models.BORepository`
- `matchRepo models.MatchRepository`
- `banRepo models.BanRepository`
- `pickRepo models.PickRepository`
- `mapRepo models.MapRepository` (pour vérifier que les maps existent dans bans/picks/matchs)

**Cas nominaux :**
- Création d'un BO3 avec 2 bans, 3 picks, 3 matchs → retourne le BO complet avec vainqueur calculé
- Récupération d'un BO par ID → retourne bans, picks, matchs
- Liste des BOs d'une ligue → retourne tous les BOs avec noms des équipes (via `FindByLigueIDWithTeams` — JOIN eva_teams)
- Mise à jour d'un BO → remplace bans/picks/matchs, recalcule vainqueur
- Récupération des maps d'une ligue → retourne les maps du split associé

**Cas limites :**
- BO3 avec seulement 1 match → autorisé (déclaratif)
- BO avec 0 bans → autorisé
- BO avec 0 picks → autorisé
- BO avec 0 matchs → autorisé (vainqueur = NULL)
- Format invalide (ni BO3 ni BO5) → erreur 400
- Équipes identiques pour equipe1 et equipe2 → erreur 400

**Gestion d'erreurs :**
- `ligue_id` invalide → 400
- `equipe1_id` ou `equipe2_id` invalide → 400
- Format invalide → 400
- BO not found → 404
- Erreur interne → 500
- Map invalide dans bans/picks/matchs → 400

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Modifier | Ajouter CreateBO, GetBO, ListBOs, UpdateBO, DeleteBO, GetLigueMaps + modifier constructeur |
| `api/routes.go` | Modifier | Ajouter routes BO sous `/api/v1/archive/bos` + route `/api/v1/archive/ligues/:id/maps` + injecter nouveaux repositories |

### Signatures

```go
// Dans ArchiveHandler — nouvelles dépendances
type ArchiveHandler struct {
    // ... existantes
    boRepo    models.BORepository
    matchRepo models.MatchRepository
    banRepo   models.BanRepository
    pickRepo  models.PickRepository
    mapRepo   models.MapRepository
}

// Constructeur modifié
func NewArchiveHandler(
    // ... existants
    boRepo models.BORepository,
    matchRepo models.MatchRepository,
    banRepo models.BanRepository,
    pickRepo models.PickRepository,
    mapRepo models.MapRepository,
) *ArchiveHandler

// Request/Response types
type createBORequest struct {
    LigueID   int                `json:"ligue_id" binding:"required"`
    Equipe1ID int                `json:"equipe1_id" binding:"required"`
    Equipe2ID int                `json:"equipe2_id" binding:"required"`
    Format    string             `json:"format" binding:"required"` // "BO3" ou "BO5"
    Bans      []createBanItem    `json:"bans"`
    Picks     []createPickItem   `json:"picks"`
    Matchs    []createMatchItem  `json:"matchs"`
}

type createBanItem struct {
    EquipeID int `json:"equipe_id" binding:"required"`
    MapID    int `json:"map_id" binding:"required"`
    Ordre    int `json:"ordre" binding:"required"`
}

type createPickItem struct {
    EquipeID *int `json:"equipe_id"`
    MapID    int  `json:"map_id" binding:"required"`
    Ordre    int  `json:"ordre" binding:"required"`
}

type createMatchItem struct {
    MapID        int     `json:"map_id" binding:"required"`
    ScoreEquipe1 int     `json:"score_equipe1" binding:"required"`
    ScoreEquipe2 int     `json:"score_equipe2" binding:"required"`
    Commentaires *string `json:"commentaires"`
}

type updatedBORequest struct {
    Format string            `json:"format" binding:"required"`
    Bans   []createBanItem   `json:"bans"`
    Picks  []createPickItem  `json:"picks"`
    Matchs []createMatchItem `json:"matchs"`
}

// Handlers
func (h *ArchiveHandler) CreateBO(c *gin.Context)
func (h *ArchiveHandler) GetBO(c *gin.Context)
func (h *ArchiveHandler) ListBOs(c *gin.Context)
func (h *ArchiveHandler) UpdateBO(c *gin.Context)
func (h *ArchiveHandler) DeleteBO(c *gin.Context)
func (h *ArchiveHandler) GetLigueMaps(c *gin.Context)
```

### Routes à ajouter dans `api/routes.go`

```go
archive.POST("/bos",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.CreateBO,
)
archive.GET("/bos/:id",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.GetBO,
)
archive.GET("/bos",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.ListBOs,
)
archive.PUT("/bos/:id",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.UpdateBO,
)
archive.DELETE("/bos/:id",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.DeleteBO,
)

// Maps du split d'une ligue (pour le dropdown de sélection de carte dans le formulaire BO)
archive.GET("/ligues/:id/maps",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"),
    archiveHandler.GetLigueMaps,
)
```

Injecter les nouveaux repositories dans `setupArchiveRoutes` :
```go
boRepo := models.NewBORepository(db)
matchRepo := models.NewMatchRepository(db)
banRepo := models.NewBanRepository(db)
pickRepo := models.NewPickRepository(db)
mapRepo := models.NewMapRepository(db)

archiveHandler := handlers.NewArchiveHandler(
    ligueLocaleRepo, ligueRepo, splitRepo, evaTeamRepo, teamRepo, divRepo,
    boRepo, matchRepo, banRepo, pickRepo, mapRepo,
    logger,
)
```

### Contraintes techniques
- **Pattern handler** : Suivre exactement le pattern des handlers existants dans `archive.go` (validation `ShouldBindJSON`, utilisation de `apierror.RespondWithSuccess` / `RespondWithError`)
- **Validation** : Vérifier que `equipe1_id != equipe2_id`, que le format est "BO3" ou "BO5", que les maps existent
- **Transactions** : Utiliser une transaction SQL pour CreateBO et UpdateBO (tout ou rien) — via `db.Begin()` / `tx.Commit()` / `tx.Rollback()`. Les méthodes des repositories doivent accepter un `*sql.Tx` optionnel. Alternativement, passer le `*sql.DB` directement et utiliser `db.Begin()` dans le handler, avec des méthodes dédiées qui prennent un `*sql.Tx`.
- **ListBOs** : Utiliser `FindByLigueIDWithTeams()` avec JOIN sur `eva_teams` pour retourner equipe1_nom, equipe1_tag, equipe2_nom, equipe2_tag directement dans la requête SQL
- **GetLigueMaps** : Récupérer la ligue (vérifier son existence), puis appeler `splitRepo.FindMapsBySplitID(ligue.SplitID)`. Le `splitRepo` est déjà disponible dans le handler.
- **Calcul vainqueur BO** : Appeler `models.ComputeBOWinner()` après chaque création/mise à jour des matchs
- **Format des réponses** :
  - Création : `{"bo": {...}}` avec détails complets
  - Liste : `{"bos": [...], "total": N}` (les éléments sont des BO sans détails par défaut, sauf `?details=true`)
  - Détail : `{"bo": {...}}` avec bans/picks/matchs
  - Suppression : `{"message": "BO supprimé"}`

### Tests à implémenter

#### Tests d'intégration (dans `internal/handlers/archive_test.go`)
- **Scénario 1** : Création d'un BO complet avec bans, picks, matchs → vérifier statut 201, vérifier vainqueur_id calculé
- **Scénario 2** : Création d'un BO sans bans/picks/matchs (minimum) → vérifier statut 201, vainqueur_id = nil
- **Scénario 3** : Récupération d'un BO par ID → vérifier statut 200, vérifier bans/picks/matchs
- **Scénario 4** : Liste des BOs par ligue → vérifier statut 200, vérifier que le BO créé est dans la liste
- **Scénario 5** : Mise à jour d'un BO (format, nouveaux bans/picks/matchs) → vérifier statut 200, vérifier nouvelles données
- **Scénario 6** : Suppression d'un BO → vérifier statut 200, vérifier que GET retourne 404
- **Scénario 7** : Création avec format invalide → vérifier statut 400
- **Scénario 8** : Création avec equipe1_id = equipe2_id → vérifier statut 400

### Documentation
- Documenter les endpoints dans `docs/archives-bo-matches.md` (fichier créé dans une autre tâche)

### Exemples d'API

```http
### Créer un BO
POST /api/v1/archive/bos
Content-Type: application/json

{
  "ligue_id": 1,
  "equipe1_id": 1,
  "equipe2_id": 2,
  "format": "BO3",
  "bans": [
    {"equipe_id": 1, "map_id": 5, "ordre": 1},
    {"equipe_id": 2, "map_id": 3, "ordre": 2}
  ],
  "picks": [
    {"equipe_id": 1, "map_id": 7, "ordre": 1},
    {"equipe_id": null, "map_id": 2, "ordre": 2},
    {"equipe_id": 2, "map_id": 9, "ordre": 3}
  ],
  "matchs": [
    {"map_id": 7, "score_equipe1": 100, "score_equipe2": 85},
    {"map_id": 2, "score_equipe1": 70, "score_equipe2": 100},
    {"map_id": 9, "score_equipe1": 90, "score_equipe2": 90}
  ]
}

### Récupérer un BO
GET /api/v1/archive/bos/1

### Lister les BOs d'une ligue
GET /api/v1/archive/bos?ligue_id=1

### Lister avec détails
GET /api/v1/archive/bos?ligue_id=1&details=true

### Mettre à jour un BO
PUT /api/v1/archive/bos/1
Content-Type: application/json

{
  "format": "BO5",
  "bans": [...],
  "picks": [...],
  "matchs": [...]
}

### Supprimer un BO
DELETE /api/v1/archive/bos/1
```
