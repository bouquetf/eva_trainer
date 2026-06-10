# Tâche #002 - Story #042 : Handler API replay URL + routes

## Objectif
Ajouter les endpoints REST pour modifier et supprimer le lien replay d'une div, mettre à jour les endpoints existants pour inclure `replay_url` dans leurs réponses, et intégrer les nouvelles routes.

## Contexte
- Story #042 : `docs/stories/story-042.md`
- Dépend de : Tâche #001 (Migration + Modèle + Repository)
- Nécessaire pour : Tâche #003 (Frontend)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. Nouvelles méthodes handler

Ajouter deux méthodes à `ArchiveHandler` dans `internal/handlers/archive.go` :

**`SetDivReplayURL`** — Définit ou supprime le lien replay d'une div :

```
PUT /api/v1/archive/divisions/:id/replay
Content-Type: application/json

{
  "replay_url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
}
```

Pour supprimer le lien, envoyer `replay_url` à `""` (chaîne vide).

**Réponse succès (200) :**
```json
{
  "div": {
    "id": 1,
    "ligue_id": 1,
    "nom": "Groupe A",
    "replay_url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  }
}
```

**Gestion d'erreurs :**
- `replay_url` manquant dans le body → 400 "replay_url is required"
- URL invalide (après validation `IsValidReplayURL`) → 400 "invalid replay URL: only YouTube and Twitch URLs are allowed" + code `invalid_replay_url`
- `ErrDivNotFound` → 404 "Div non trouvée"
- Erreur interne → 500

**`DeleteDivReplayURL`** — Supprime le lien replay d'une div (endpoint dédié) :

```
DELETE /api/v1/archive/divisions/:id/replay
```

**Réponse succès (200) :**
```json
{
  "div": {
    "id": 1,
    "ligue_id": 1,
    "nom": "Groupe A",
    "replay_url": ""
  }
}
```

**Gestion d'erreurs :**
- `ErrDivNotFound` → 404 "Div non trouvée"
- Erreur interne → 500

#### 2. Mise à jour des handlers existants

Mettre à jour la structure `divWithTeams` pour inclure `replay_url` :

```go
type divWithTeams struct {
    ID        int               `json:"id"`
    LigueID   int               `json:"ligue_id"`
    Nom       string            `json:"nom"`
    ReplayURL string            `json:"replay_url"`
    Teams     []*models.EvaTeam `json:"teams"`
}
```

Mettre à jour `ListDivs` pour remplir `ReplayURL` depuis les divs :

Dans la boucle `for _, d := range divs`, ajouter `ReplayURL: d.ReplayURL`.

**Note :** `GetDiv` renvoie déjà le `div` directement (qui contient maintenant `replay_url` grâce à la mise à jour du scanDiv dans la tâche #001), donc aucune modification nécessaire pour `GetDiv`.

#### 3. Routes

Ajouter dans `api/routes.go` après la route `generate-round-robin` :

```go
archive.PUT("/divisions/:id/replay",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.SetDivReplayURL,
)
archive.DELETE("/divisions/:id/replay",
    middleware.RequireAuth(logger, userRepo, jwtService),
    middleware.RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"),
    archiveHandler.DeleteDivReplayURL,
)
```

#### 4. Validation d'URL dans le handler

Dans `SetDivReplayURL`, valider l'URL avec `models.IsValidReplayURL` :

```go
func (h *ArchiveHandler) SetDivReplayURL(c *gin.Context) {
    divID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
        return
    }

    var req struct {
        ReplayURL string `json:"replay_url"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_request_body", "invalid request body", http.StatusBadRequest))
        return
    }

    // replay_url must be present in body even if empty
    // (use pointer or check if field was sent — but simpler: require field in json)

    if !models.IsValidReplayURL(req.ReplayURL) {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_replay_url", "invalid replay URL: only YouTube and Twitch URLs are allowed", http.StatusBadRequest))
        return
    }

    div, err := h.divRepo.SetReplayURL(divID, req.ReplayURL)
    if errors.Is(err, models.ErrDivNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to set replay url", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"div": div})
}
```

```go
func (h *ArchiveHandler) DeleteDivReplayURL(c *gin.Context) {
    divID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError("invalid_div_id", "invalid div id", http.StatusBadRequest))
        return
    }

    div, err := h.divRepo.SetReplayURL(divID, "")
    if errors.Is(err, models.ErrDivNotFound) {
        apierror.RespondWithError(c, apierror.NewAPIError("div_not_found", "Div non trouvée", http.StatusNotFound))
        return
    }
    if err != nil {
        h.logger.Error("failed to delete replay url", "error", err)
        apierror.RespondWithError(c, apierror.NewAPIError(apierror.ErrCodeInternal, "internal server error", http.StatusInternalServerError))
        return
    }

    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{"div": div})
}
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive.go` | Modifier | Ajouter SetDivReplayURL, DeleteDivReplayURL, update divWithTeams, update ListDivs |
| `api/routes.go` | Modifier | Ajouter les 2 routes PUT/DELETE /divisions/:id/replay |
| `internal/handlers/archive_test.go` | Modifier | Ajouter tests d'intégration |

### Tests à implémenter

#### Tests d'intégration (dans `internal/handlers/archive_test.go`, tag `integration`)

Utiliser le pattern existant avec `setupArchiveHandlerTest`, `gin.Context`, `httptest.NewRecorder`.

- **Scénario 1 : SetReplayURL YouTube valide**
  - Setup : créer ligue + div
  - Requête : `PUT /divisions/:id/replay` avec `{"replay_url": "https://www.youtube.com/watch?v=test123"}`
  - Vérifier : status 200, response div.replay_url == "https://www.youtube.com/watch?v=test123"

- **Scénario 2 : SetReplayURL Twitch valide**
  - Setup : créer ligue + div
  - Requête : `PUT /divisions/:id/replay` avec `{"replay_url": "https://www.twitch.tv/videos/123456"}`
  - Vérifier : status 200

- **Scénario 3 : SetReplayURL invalide**
  - Setup : créer ligue + div
  - Requête : `PUT /divisions/:id/replay` avec `{"replay_url": "https://vimeo.com/123"}`
  - Vérifier : status 400, body code == "invalid_replay_url"

- **Scénario 4 : SetReplayURL avec chaîne vide (suppression)**
  - Setup : créer ligue + div avec replay_url défini
  - Requête : `PUT /divisions/:id/replay` avec `{"replay_url": ""}`
  - Vérifier : status 200, response div.replay_url == ""

- **Scénario 5 : SetReplayURL sur div inexistante**
  - Requête : `PUT /divisions/99999/replay` avec `{"replay_url": "https://www.youtube.com/watch?v=test"}`
  - Vérifier : status 404

- **Scénario 6 : DeleteDivReplayURL**
  - Setup : créer ligue + div avec replay_url défini
  - Requête : `DELETE /divisions/:id/replay`
  - Vérifier : status 200, response div.replay_url == ""

- **Scénario 7 : DeleteDivReplayURL sur div inexistante**
  - Requête : `DELETE /divisions/99999/replay`
  - Vérifier : status 404

- **Scénario 8 : ListDivs inclut replay_url**
  - Setup : créer ligue + 2 divs, une avec replay_url, une sans
  - Requête : `GET /divisions?ligue_id=X`
  - Vérifier : status 200, la div avec URL a replay_url non vide, l'autre a replay_url == ""

- **Scénario 9 : GetDiv inclut replay_url**
  - Setup : créer ligue + div avec replay_url
  - Requête : `GET /divisions/:id`
  - Vérifier : status 200, response div.replay_url == valeur attendue
