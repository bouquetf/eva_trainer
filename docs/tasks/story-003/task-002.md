# Tâche #002 - Story #003 : Handlers équipe et routes API

## Objectif

Implémenter les endpoints API REST de gestion d'équipe : création, consultation, modification, suppression et départ. Ajouter le middleware de vérification du rôle capitaine.

## Contexte

- Story #003 : `docs/stories/story-003.md` - Team Creation and Management
- Dépend de : Tâche #001 (modèles Team + TeamMember), Story #002 (auth, middleware `RequireAuth`)
- Nécessaire pour : Tâche #003 (frontend), Tâche #004 (tests)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Handler Team :**
Créer `internal/handlers/team.go` avec les endpoints suivants :

| Méthode | Path | Auth | Rôle | Description |
|---------|------|------|------|-------------|
| POST | `/api/v1/teams` | RequireAuth | - | Créer une équipe (le créateur devient capitaine) |
| GET | `/api/v1/teams` | RequireAuth | - | Lister les équipes de l'utilisateur connecté |
| GET | `/api/v1/teams/:id` | RequireAuth | member | Détails d'une équipe |
| PUT | `/api/v1/teams/:id` | RequireAuth | captain | Modifier le nom de l'équipe |
| DELETE | `/api/v1/teams/:id` | RequireAuth | captain | Supprimer l'équipe (saisie du nom requise) |
| POST | `/api/v1/teams/:id/leave` | RequireAuth | member | Quitter l'équipe (si capitaine → équipe supprimée) |

**Partie 2 - Détail des endpoints :**

**POST `/api/v1/teams`** - Création d'équipe
- Body : `{ "name": "string" }`
- Valide le nom via `models.ValidateTeamName`
- Crée l'équipe via `teamRepo.Create(name, userID)`
- Réponse 201 : `{ "id", "name", "captain_id", "created_at", "updated_at" }`
- Erreur 400 : nom invalide ou déjà utilisé
- Erreur 500 : erreur interne

**GET `/api/v1/teams`** - Liste des équipes de l'utilisateur
- Appelle `teamRepo.FindByUserID(userID)`
- Pour chaque équipe, ajoute le nombre de membres et le rôle de l'utilisateur connecté
- Réponse 200 : `[{ "team": {...}, "member_count": int, "my_role": "string" }, ...]`
- Erreur 500 : erreur interne

**GET `/api/v1/teams/:id`** - Détails d'une équipe
- Vérifie que l'utilisateur est membre de l'équipe
- Retourne l'équipe + la liste des membres (avec leurs infos)
- Réponse 200 : `{ "team": {...}, "members": [{ "user_id", "role", "joined_at" }] }`
- Erreur 403 : l'utilisateur n'est pas membre
- Erreur 404 : équipe introuvable

**PUT `/api/v1/teams/:id`** - Modification du nom
- Vérifie que l'utilisateur est capitaine
- Body : `{ "name": "string" }`
- Valide le nom, met à jour
- Réponse 200 : `{ "id", "name", "captain_id", "created_at", "updated_at" }`
- Erreur 400 : nom invalide
- Erreur 403 : pas capitaine

**DELETE `/api/v1/teams/:id`** - Suppression d'équipe
- Vérifie que l'utilisateur est capitaine
- Body : `{ "confirm_with_name": "Team Name" }` — le nom exact de l'équipe doit être fourni
- Si `confirm_with_name` est vide ou ne correspond pas au nom actuel de l'équipe → erreur 400 "team name does not match, confirmation required"
- Appelle `convertTeamZonesToTemplates(teamID)` (fonction squelette avec TODO) avant la suppression
- Supprime l'équipe (la table `zones` a ON DELETE CASCADE)
- Réponse 200 : `{ "message": "team deleted" }`
- Erreur 403 : pas capitaine
- Erreur 400 : nom de confirmation incorrect

**POST `/api/v1/teams/:id/leave`** - Quitter l'équipe
- Vérifie que l'utilisateur est membre
- Si l'utilisateur est le capitaine :
  - Appelle `convertTeamZonesToTemplates(teamID)` (fonction squelette avec TODO) avant la suppression
  - Supprime l'équipe (équivaut à DELETE avec confirmation automatique par le départ)
- Sinon : retire le membre de `team_members`
- Réponse 200 : `{ "message": "left team" }` ou `{ "message": "team deleted (captain left without transfer)" }`
- Erreur 403 : pas membre
- Erreur 500 : erreur interne

**Partie 3 - Middleware `RequireTeamRole` :**
Créer `internal/middleware/team.go` avec un middleware qui vérifie le rôle dans l'équipe :

```go
// RequireTeamRole vérifie que l'utilisateur a au moins l'un des rôles spécifiés dans l'équipe.
// Le paramètre :id (team ID) doit être dans l'URL.
// Utilise teamMemberRepo.FindByUserAndTeam pour vérifier l'appartenance et le rôle.
func RequireTeamRole(teamMemberRepo models.TeamMemberRepository, allowedRoles ...string) gin.HandlerFunc
```

- Extrait `:id` des paramètres Gin
- Récupère l'utilisateur du contexte (via `RequireAuth`)
- Appelle `teamMemberRepo.FindByUserAndTeam(userID, teamID)`
- Si pas membre → 403 `{ "error": "not_team_member", "message": "you are not a member of this team" }`
- Si rôle non autorisé → 403 `{ "error": "insufficient_role", "message": "you don't have the required role" }`
- Si OK → stocke `teamMember` dans le contexte Gin via `c.Set("team_member", member)` et passe au handler suivant

**Partie 4 - Intégration des routes :**
Modifier `api/routes.go` pour ajouter `setupTeamRoutes` :

```go
func setupTeamRoutes(rg *gin.RouterGroup, cfg *config.AppConfig) {
    db := database.DB()
    teamRepo := models.NewTeamRepository(db)
    teamMemberRepo := models.NewTeamMemberRepository(db)
    teamHandler := handlers.NewTeamHandler(teamRepo, teamMemberRepo)

    teams := rg.Group("/teams")
    teams.Use(middleware.RequireAuth(
        models.NewUserRepository(db),
        services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry),
    ))
    {
        teams.POST("", teamHandler.Create)
        teams.GET("", teamHandler.ListMyTeams)
        teams.GET("/:id", teamHandler.GetTeam)
        teams.PUT("/:id", teamHandler.UpdateTeam)
        teams.DELETE("/:id", teamHandler.DeleteTeam)
        teams.POST("/:id/leave", teamHandler.LeaveTeam)
    }
}
```

**Cas nominaux :**
- Création d'équipe → HTTP 201 + équipe créée avec le créateur comme capitaine
- Consultation de ses équipes → HTTP 200 + liste des équipes avec rôles
- Modification du nom → HTTP 200 + équipe mise à jour
- Suppression avec confirmation → HTTP 200 + équipe supprimée
- Quitter une équipe (non-capitaine) → HTTP 200 + membre retiré
- Capitaine quitte → HTTP 200 + équipe supprimée

**Cas limites :**
- Création avec nom déjà pris → erreur 400 "team name already exists"
- Création avec nom invalide → erreur 400 "invalid team name"
- Modification par non-capitaine → erreur 403
- Suppression sans confirmation → erreur 400 "confirmation required"
- Suppression par non-capitaine → erreur 403
- Consultation d'une équipe dont on n'est pas membre → erreur 403
- Quitter une équipe dont on n'est pas membre → erreur 403
- ID d'équipe inexistant → erreur 404 "team not found"

**Gestion d'erreurs :**
Format JSON standard (même format que les handlers d'auth) :
```json
{
  "error": "error_code",
  "message": "Description lisible"
}
```

| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | invalid_team_name | Nom invalide (caractères, longueur) |
| 400 | team_name_exists | Nom déjà utilisé |
| 400 | confirmation_name_mismatch | Le nom saisi ne correspond pas au nom de l'équipe |
| 403 | not_team_member | Pas membre de l'équipe |
| 403 | insufficient_role | Rôle insuffisant pour l'action |
| 404 | team_not_found | Équipe inexistante |
| 500 | internal_error | Erreur serveur |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/team.go` | Créer | TeamHandler avec tous les endpoints |
| `internal/middleware/team.go` | Créer | Middleware RequireTeamRole |
| `api/routes.go` | Modifier | Ajouter setupTeamRoutes, appeler dans SetupRoutes |
| `api/routes_test.go` | Modifier | Ajouter tests des nouvelles routes (voir Tâche #004) |

### Signatures

```go
// internal/handlers/team.go
package handlers

import "eva_trainer/internal/models"

type TeamHandler struct {
    teamRepo       models.TeamRepository
    teamMemberRepo models.TeamMemberRepository
}

func NewTeamHandler(
    teamRepo models.TeamRepository,
    teamMemberRepo models.TeamMemberRepository,
) *TeamHandler

func (h *TeamHandler) Create(c *gin.Context)
func (h *TeamHandler) ListMyTeams(c *gin.Context)
func (h *TeamHandler) GetTeam(c *gin.Context)
func (h *TeamHandler) UpdateTeam(c *gin.Context)
func (h *TeamHandler) DeleteTeam(c *gin.Context)
func (h *TeamHandler) LeaveTeam(c *gin.Context)

// convertTeamZonesToTemplates est un squelette pour la conversion des zones en templates
// lors de la suppression d'équipe. L'implémentation réelle sera faite dans Story #011.
// Pour chaque zone de l'équipe, copier les polygones et noms dans les tables
// templates et template_zones.
func convertTeamZonesToTemplates(teamID int) error {
    // TODO: Story #011 - Implémenter la conversion des zones publiques en templates
    return nil
}
```

```go
// internal/middleware/team.go

// RequireTeamRole vérifie que l'utilisateur a le rôle requis dans l'équipe.
func RequireTeamRole(teamMemberRepo models.TeamMemberRepository, allowedRoles ...string) gin.HandlerFunc

// GetTeamMember récupère le TeamMember du contexte Gin.
func GetTeamMember(c *gin.Context) (*models.TeamMember, bool)
```

### Contraintes techniques

- **Gin** : Utiliser `c.ShouldBindJSON` pour les body, `c.Param("id")` pour les paramètres d'URL, `c.Set`/`c.Get` pour le contexte.
- **Conversion ID** : Utiliser `strconv.Atoi` pour convertir `:id` en entier.
- **Pattern handler** : Suivre le pattern de `handlers/auth.go` (validation → action → réponse JSON).
- **Routes** : Ajouter `setupTeamRoutes` dans `SetupRoutes` après `setupAuthRoutes`, dans le groupe `apiV1`. Comme pour l'auth, elle est appelée directement dans `SetupRoutes`.
- **Confirmation** : Le body DELETE doit contenir `{ "confirmation": true }`. Si absent ou false → 400.
- **Team info dans contexte** : Le middleware `RequireTeamRole` stocke `*models.TeamMember` dans le contexte Gin sous la clé `"team_member"`.
- **Pas de service layer dédié** : Les handlers utilisent directement les repositories (même pattern que `AuthHandler`).

### Tests à implémenter

Les tests seront créés dans la Tâche #004.

### Documentation

Aucune documentation à créer pour cette tâche.
