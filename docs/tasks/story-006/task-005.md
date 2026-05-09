# Tâche #005 - Story #006 : Routes backend

## Objectif
Enregistrer toutes les nouvelles routes d'invitation et de demande d'adhésion dans le routeur Gin, en configurant les middlewares d'authentification, de rôle et de permission appropriés.

## Contexte
- Story #006 : Team Invitation and Join System
- Dépend de : Tâche #003 (Handlers invitations), Tâche #004 (Handlers demandes)
- Nécessaire pour : Tâche #006 (Frontend API)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Mettre à jour `api/routes.go` pour :
1. Créer les nouvelles instances de handler (`TeamInvitationHandler`, `TeamJoinRequestHandler`)
2. Créer le nouveau middleware `RequireAuthOrMember`
3. Modifier `GET /api/v1/teams` → retourner TOUTES les équipes (avec pagination/recherche)
4. Ajouter `GET /api/v1/myteams` → retourne les équipes de l'utilisateur connecté (remplace l'ancien comportement de GET /teams)
5. Ajouter les routes d'invitation sous `/api/v1/teams/:id/invitations/` et `/api/v1/invite-links/`
6. Ajouter les routes de demande d'adhésion sous `/api/v1/teams/:id/join-requests/`
7. Ajouter la route de recherche d'utilisateurs `/api/v1/users/search`
8. Configurer les middlewares appropriés (auth, team role, permission)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `api/routes.go` | Modifier | Ajouter les nouvelles routes et modifier les existantes |
| `internal/middleware/team.go` | Modifier | Ajouter `RequireAuthOrMember` |

### Nouvelles routes

#### Routes équipes modifiées

| Méthode | Path | Changement |
|---------|------|------------|
| GET | `/api/v1/teams` | **MODIFIÉ** : retourne TOUTES les équipes avec pagination (`?page=1&per_page=20`) et recherche (`?q=nom`) |
| GET | `/api/v1/myteams` | **NOUVEAU** : retourne les équipes de l'utilisateur connecté (ancien comportement de GET /teams) |

#### Routes d'invitation

| Méthode | Path | Auth | Rôle | Permission | Handler |
|---------|------|------|------|------------|---------|
| POST | `/api/v1/teams/:id/invitations` | Oui | captain/cocaptain | `team:invite` | `teamInvitationHandler.InviteByDisplayName` |
| POST | `/api/v1/teams/:id/invitations/link` | Oui | captain/cocaptain | `team:invite` | `teamInvitationHandler.GenerateInviteLink` |
| GET | `/api/v1/teams/:id/invitations/pending` | Oui | captain/cocaptain | `team:requests:view` | `teamInvitationHandler.ListPendingByTeam` |

| Méthode | Path | Auth | Rôle | Permission | Handler |
|---------|------|------|------|------------|---------|
| GET | `/api/v1/invitations` | Oui | - | - | `teamInvitationHandler.ListMyInvitations` |
| POST | `/api/v1/invitations/:invitationId/accept` | Oui | - | - | `teamInvitationHandler.AcceptInvitation` |
| POST | `/api/v1/invitations/:invitationId/decline` | Oui | - | - | `teamInvitationHandler.DeclineInvitation` |
| GET | `/api/v1/invite-links/:token` | Non | - | - | `teamInvitationHandler.GetInviteLinkInfo` |
| POST | `/api/v1/invite-links/:token/accept` | Oui | - | - | `teamInvitationHandler.AcceptInviteLink` |

#### Routes de demande d'adhésion

| Méthode | Path | Auth | Rôle | Permission | Handler |
|---------|------|------|------|------------|---------|
| POST | `/api/v1/teams/:id/join-requests` | Oui | `RequireAuthOrMember` | - | `joinRequestHandler.RequestJoin` |
| GET | `/api/v1/teams/:id/join-requests` | Oui | captain/cocaptain | `team:requests:view` | `joinRequestHandler.ListPendingRequests` |
| PUT | `/api/v1/teams/:id/join-requests/:requestId/accept` | Oui | captain/cocaptain | `team:requests:manage` | `joinRequestHandler.AcceptRequest` |
| PUT | `/api/v1/teams/:id/join-requests/:requestId/decline` | Oui | captain/cocaptain | `team:requests:manage` | `joinRequestHandler.DeclineRequest` |
| GET | `/api/v1/teams/:id/join-requests/mine` | Oui | `RequireAuthOrMember` | - | `joinRequestHandler.GetMyRequest` |
| DELETE | `/api/v1/teams/:id/join-requests/mine` | Oui | `RequireAuthOrMember` | - | `joinRequestHandler.CancelMyRequest` |

#### Route de recherche d'utilisateurs (autocomplétion)

| Méthode | Path | Auth | Rôle | Permission | Handler |
|---------|------|------|------|------------|---------|
| GET | `/api/v1/users/search` | Oui | - | - | `teamInvitationHandler.SearchUsers` |

### Nouveau middleware `RequireAuthOrMember`

Ajouter dans `internal/middleware/team.go` :

```go
// RequireAuthOrMember authentifie l'utilisateur et tente de le trouver comme membre de l'équipe.
// Contrairement à RequireTeamRole, il ne rejette PAS les non-membres.
// Si l'utilisateur est membre, team_member est stocké dans le contexte.
// Si l'utilisateur n'est pas membre, la requête continue sans team_member.
func RequireAuthOrMember(
    teamMemberRepo models.TeamMemberRepository,
    userRepo models.UserRepository,
    jwtService *services.JWTService,
) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Authentifier l'utilisateur (comme RequireAuth)
        user, err := MustGetUser(c)
        if err != nil {
            // Tente d'extraire le token et de valider
            authHeader := c.GetHeader("Authorization")
            if authHeader == "" {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "unauthorized"})
                c.Abort()
                return
            }
            // Valider token et récupérer user...
            // (même logique que RequireAuth)
        }

        // 2. Essayer de trouver le membre (ne pas échouer si pas trouvé)
        teamID, err := strconv.Atoi(c.Param("id"))
        if err == nil {
            member, err := teamMemberRepo.FindByUserAndTeam(user.ID, teamID)
            if err == nil {
                c.Set("team_member", member)
            }
            // Si err != nil (pas membre), on continue sans set
        }

        c.Set("team_id", teamID)
        c.Next()
    }
}
```

### Détail du code à ajouter dans `setupTeamRoutes`

```go
// Nouveaux repositories
invitationRepo := models.NewTeamInvitationRepository(db, teamMemberRepo)
joinRequestRepo := models.NewTeamJoinRequestRepository(db, teamMemberRepo)

// Nouveaux handlers
teamInvitationHandler := handlers.NewTeamInvitationHandler(
    invitationRepo, teamMemberRepo, userRepo, teamRepo,
)
joinRequestHandler := handlers.NewTeamJoinRequestHandler(
    joinRequestRepo, teamMemberRepo, teamRepo,
)

// MODIFIER la route GET /teams existante pour retourner toutes les équipes
// Remplacer le handler existant :
// teams.GET("", teamHandler.ListMyTeams) → SUPPRIMER
// Ajouter :
// teams.GET("", teamInvitationHandler.ListAllTeams) → NOUVEAU handler
// Note : Le handler ListAllTeams doit être ajouté sur TeamHandler ou sur un handler dédié

// AJOUTER la route /myteams (hors du groupe teams)
myTeams := rg.Group("/myteams")
myTeams.Use(middleware.RequireAuth(userRepo, jwtService))
{
    myTeams.GET("", teamHandler.ListMyTeams)
}

// Routes d'invitation (sous /teams/:id)
teams.POST("/:id/invitations",
    middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
    middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:invite"),
    teamInvitationHandler.InviteByDisplayName,
)
teams.POST("/:id/invitations/link",
    middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
    middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:invite"),
    teamInvitationHandler.GenerateInviteLink,
)
teams.GET("/:id/invitations/pending",
    middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
    middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:requests:view"),
    teamInvitationHandler.ListPendingByTeam,
)

// Routes de demande d'adhésion (sous /teams/:id)
teams.POST("/:id/join-requests",
    middleware.RequireAuthOrMember(teamMemberRepo, userRepo, jwtService),
    joinRequestHandler.RequestJoin,
)
teams.GET("/:id/join-requests/mine",
    middleware.RequireAuthOrMember(teamMemberRepo, userRepo, jwtService),
    joinRequestHandler.GetMyRequest,
)
teams.DELETE("/:id/join-requests/mine",
    middleware.RequireAuthOrMember(teamMemberRepo, userRepo, jwtService),
    joinRequestHandler.CancelMyRequest,
)
teams.GET("/:id/join-requests",
    middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
    middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:requests:view"),
    joinRequestHandler.ListPendingRequests,
)
teams.PUT("/:id/join-requests/:requestId/accept",
    middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
    middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:requests:manage"),
    joinRequestHandler.AcceptRequest,
)
teams.PUT("/:id/join-requests/:requestId/decline",
    middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain"),
    middleware.RequirePermission(teamMemberRepo, rolePermRepo, permRepo, "team:requests:manage"),
    joinRequestHandler.DeclineRequest,
)

// Routes d'invitation globales (hors /teams)
invitations := rg.Group("/invitations")
invitations.Use(middleware.RequireAuth(userRepo, jwtService))
{
    invitations.GET("", teamInvitationHandler.ListMyInvitations)
    invitations.POST("/:invitationId/accept", teamInvitationHandler.AcceptInvitation)
    invitations.POST("/:invitationId/decline", teamInvitationHandler.DeclineInvitation)
}

// Routes pour les liens d'invitation (préfixe séparé invite-links)
inviteLinks := rg.Group("/invite-links")
{
    // Public (sans auth) - voir les infos du lien
    inviteLinks.GET("/:token", teamInvitationHandler.GetInviteLinkInfo)

    // Avec auth - accepter un lien
    inviteLinks.POST("/:token/accept",
        middleware.RequireAuth(userRepo, jwtService),
        teamInvitationHandler.AcceptInviteLink,
    )
}

// Route de recherche d'utilisateurs (autocomplétion)
users := rg.Group("/users")
users.Use(middleware.RequireAuth(userRepo, jwtService))
{
    users.GET("/search", teamInvitationHandler.SearchUsers)
}
```

### Contraintes techniques
- **Middleware** : Les routes qui nécessitent captain/cocaptain doivent utiliser `middleware.RequireTeamRole(teamMemberRepo, "captain", "cocaptain")`
- **Middleware** : Les routes sans rôle spécifique mais avec auth utilisent `middleware.RequireAuth(userRepo, jwtService)`
- **Middleware RequireAuthOrMember** : Pour les routes accessibles aux non-membres mais nécessitant auth (join-requests). Stocké dans `internal/middleware/team.go`
- **Ordre Gin** : Les routes fixes (`/mine`) doivent être déclarées AVANT les routes paramétrées (`/:requestId/accept`) dans le groupe
- **Constructeur JWT** : Le `jwtService` est créé dans `setupTeamRoutes` actuellement via `services.NewJWTService(...)`. Vérifier la portée. Si `setupTeamRoutes` n'a pas accès à `jwtService`, le passer en paramètre ou le recréer.
- **Permissions** : Les permissions `team:invite`, `team:requests:view`, `team:requests:manage` existent déjà dans la base de données (seed)
- **GET /teams** : Le handler actuel `teamHandler.ListMyTeams` est déplacé sur `GET /myteams`. Le nouveau `GET /teams` doit utiliser `teamRepo.FindAll()` avec filtrage optionnel par query string `?q=...`. Ajouter une méthode `FindAllPaginated(q string, page, perPage int)` au TeamRepository si nécessaire.
- **GET /myteams** : Utilise l'implémentation existante de `teamHandler.ListMyTeams` (inchangé)
- **Préfixe invite-links** : Utiliser `/api/v1/invite-links/` comme préfixe pour toutes les routes liées aux liens d'invitation (séparé de `/api/v1/invitations/` pour éviter les conflits)

### Tests

Aucun test spécifique pour cette tâche (les tests de routage sont couverts par les tests d'intégration des handlers).

### Documentation
- Aucune documentation spécifique pour cette tâche
