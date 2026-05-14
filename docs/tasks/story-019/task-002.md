# Tâche #002 - Story #019 : User Settings — API Endpoints

## Objectif
Créer les endpoints API REST pour lire et modifier les préférences utilisateur, avec validation et tests d'intégration.

## Contexte
- Story #019 : [Paramètres de révision et timers configurables](../../stories/story-019-done.md)
- Dépend de : Tâche #001 (DB + Repository)
- Nécessaire pour : Tâche #003 (Frontend Settings UI), Tâche #004 (Revision timers)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter deux endpoints protégés par authentification sur le groupe `/api/v1/auth/` :

**GET /api/v1/auth/settings**
Retourne les settings de l'utilisateur connecté (fusionnés avec les défauts si settings NULL en base).

**PUT /api/v1/auth/settings**
Met à jour les settings de l'utilisateur connecté. Le body est un objet JSON partiel ; les champs non fournis conservent leur valeur actuelle (merge côté serveur).

**Cas nominaux :**
- GET sans settings préexistants → retourne `{"revision_correct_timer": 3, "revision_incorrect_timer": 5, "revision_question_count": 20}`
- GET avec settings partiels → fusion avec les défauts
- PUT avec body `{"revision_correct_timer": 2}` → seul ce champ est modifié, les autres inchangés
- PUT avec body complet → remplacement complet

**Gestion d'erreurs :**
- PUT avec `revision_correct_timer: 0` → 400 Bad Request (`"invalid_settings"`)
- PUT avec `revision_question_count: 0` → 400 Bad Request
- PUT avec body vide → 400 Bad Request
- PUT avec des champs inconnus → ignorés (pas d'erreur)
- Non authentifié → 401 (géré par le middleware RequireAuth existant)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/auth.go` | Modifier | Ajouter `GetSettings` et `UpdateSettings` à `AuthHandler` |
| `api/routes.go` | Modifier | Ajouter les routes dans le groupe `protected` de `setupAuthRoutes` |
| `internal/handlers/auth_test.go` | Modifier | Ajouter les tests d'intégration pour les settings (voir ci-dessous) |

### Signatures

```go
// À ajouter dans internal/handlers/auth.go

func (h *AuthHandler) GetSettings(c *gin.Context) {
    user, err := middleware.MustGetUser(c)
    if err != nil { return }

    settings, err := h.userRepo.GetUserSettings(user.ID)
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError(
            "internal_error", "failed to load settings", http.StatusInternalServerError,
        ))
        return
    }

    // Mapping explicite — pattern cohérent avec Me() et RespondWithSuccess
    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
        "revision_correct_timer":   settings.RevisionCorrectTimer,
        "revision_incorrect_timer": settings.RevisionIncorrectTimer,
        "revision_question_count":  settings.RevisionQuestionCount,
    })
}

func (h *AuthHandler) UpdateSettings(c *gin.Context) {
    user, err := middleware.MustGetUser(c)
    if err != nil { return }

    var req struct {
        RevisionCorrectTimer   *int `json:"revision_correct_timer"`
        RevisionIncorrectTimer *int `json:"revision_incorrect_timer"`
        RevisionQuestionCount  *int `json:"revision_question_count"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError(
            "invalid_settings", "invalid settings format", http.StatusBadRequest,
        ))
        return
    }

    // Valider : timers >= 1, question_count >= 1
    if (req.RevisionCorrectTimer != nil && *req.RevisionCorrectTimer < 1) ||
       (req.RevisionIncorrectTimer != nil && *req.RevisionIncorrectTimer < 1) ||
       (req.RevisionQuestionCount != nil && *req.RevisionQuestionCount < 1) {
        apierror.RespondWithError(c, apierror.NewAPIError(
            "invalid_settings", "values must be >= 1", http.StatusBadRequest,
        ))
        return
    }

    // Fusion avec les settings existants
    current, err := h.userRepo.GetUserSettings(user.ID)
    if err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError(
            "internal_error", "failed to load settings", http.StatusInternalServerError,
        ))
        return
    }

    merged := *current
    if req.RevisionCorrectTimer != nil { merged.RevisionCorrectTimer = *req.RevisionCorrectTimer }
    if req.RevisionIncorrectTimer != nil { merged.RevisionIncorrectTimer = *req.RevisionIncorrectTimer }
    if req.RevisionQuestionCount != nil { merged.RevisionQuestionCount = *req.RevisionQuestionCount }

    if err := h.userRepo.UpdateUserSettings(user.ID, &merged); err != nil {
        apierror.RespondWithError(c, apierror.NewAPIError(
            "internal_error", "failed to save settings", http.StatusInternalServerError,
        ))
        return
    }

    // Réponse : même format que GET
    apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
        "revision_correct_timer":   merged.RevisionCorrectTimer,
        "revision_incorrect_timer": merged.RevisionIncorrectTimer,
        "revision_question_count":  merged.RevisionQuestionCount,
    })
}
```

```go
// Route dans api/routes.go, fonction setupAuthRoutes
protected.GET("/auth/settings", authHandler.GetSettings)
protected.PUT("/auth/settings", authHandler.UpdateSettings)
```

### Contraintes techniques
- **Framework** : Gin 1.12 — suivre le pattern des handlers existants (ex: `UpdateDisplayName`)
- **Validation** : côté handler uniquement (pas de validation base)
- **Fusion partielle** : le handler doit lire les settings actuels, appliquer les overrides du body, puis sauvegarder le résultat complet
- **Format réponse** : `RespondWithSuccess(c, http.StatusOK, gin.H{"settings": settings})`

### Tests à implémenter

#### Tests d'intégration — `internal/handlers/auth_test.go`

Ajouter les fichiers de migration nécessaires au `migrationFiles` dans `setupTestDB` :
```go
var migrationFiles = []string{
    // ... existants ...
    "000015_add_template_is_public.up.sql", // attention : cette migration dépend de la table templates
    // On n'ajoute PAS 000015 si le test n'a pas besoin de templates
}
```

**Stratégie** : pour éviter la dépendance à la migration 000015, créer un setup dédié qui utilise seulement les migrations jusqu'à 000004 + la nouvelle 000016.

Le plus simple : ajouter `"000016_add_user_settings.up.sql"` aux migrations chargées, après avoir vérifié que l'ordre est cohérent.

**Scénario 1** : `TestGetSettings_NoSettings_ReturnsDefaults`
- Créer un utilisateur
- GET /api/v1/auth/settings (via handler, avec RequireAuth middleware)
- Vérifier : status 200, body = `{"revision_correct_timer": 3, "revision_incorrect_timer": 5, "revision_question_count": 20}`

**Scénario 2** : `TestGetSettings_WithPartialSettings_ReturnsMerged`
- Créer un utilisateur, UPDATE users SET settings = '{"revision_correct_timer": 2}' WHERE id = ...
- GET /api/v1/auth/settings
- Vérifier : status 200, revision_correct_timer = 2, autres = défauts

**Scénario 3** : `TestUpdateSettings_StoresAndPersists`
- Créer un utilisateur
- PUT /api/v1/auth/settings avec body `{"revision_correct_timer": 4, "revision_incorrect_timer": 6, "revision_question_count": 15}`
- Vérifier : status 200, réponse contient les nouvelles valeurs
- GET /api/v1/auth/settings → vérifier que les valeurs sont persistées

**Scénario 4** : `TestUpdateSettings_PartialUpdate`
- Créer un utilisateur, pré-remplir settings avec des valeurs
- PUT body `{"revision_correct_timer": 2}`
- Vérifier : seuls les champs fournis sont modifiés, les autres inchangés

**Scénario 5** : `TestUpdateSettings_InvalidValues_Returns400`
- PUT avec `{"revision_correct_timer": 0}` → 400
- PUT avec `{"revision_question_count": 0}` → 400
- PUT avec body `{}` → 400

### Documentation

#### Documentation à mettre à jour
- `docs/user-settings.md` : documenter les 2 endpoints (GET, PUT)

### Exemples d'utilisation (fichier .http)

```http
### GET user settings
GET {{baseUrl}}/api/v1/auth/settings
Authorization: Bearer {{token}}

### Update user settings (partial)
PUT {{baseUrl}}/api/v1/auth/settings
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "revision_correct_timer": 2,
  "revision_incorrect_timer": 7,
  "revision_question_count": 25
}
```
