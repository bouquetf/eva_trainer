# Tâche #004 - Story #006 : Handlers de demande d'adhésion

## Objectif
Créer le handler HTTP pour gérer les demandes d'adhésion : demander à rejoindre une équipe, lister les demandes en attente, accepter/refuser une demande, consulter le statut et annuler sa propre demande.

## Contexte
- Story #006 : Team Invitation and Join System
- Dépend de : Tâche #002 (Modèles)
- Nécessaire pour : Tâche #005 (Routes)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un handler `TeamJoinRequestHandler` avec les méthodes suivantes :

1. **Demander à rejoindre une équipe** : Un utilisateur POSTe sur `/teams/:id/join-requests` (body vide). L'API vérifie que l'utilisateur n'est pas déjà membre, puis crée une demande avec status='pending'. Si l'utilisateur fait partie d'aucune équipe, il peut demander à rejoindre n'importe quelle équipe.

2. **Lister les demandes en attente (capitaine/cocapitaine)** : GET `/teams/:id/join-requests` retourne toutes les demandes pending avec les infos utilisateur (display_name).

3. **Accepter une demande** : PUT `/teams/:id/join-requests/:requestId/accept`. L'utilisateur devient membre de l'équipe (rôle 'member') et la demande est marquée 'accepted'.

4. **Refuser une demande** : PUT `/teams/:id/join-requests/:requestId/decline`. La demande est marquée 'declined'.

5. **Voir le statut de ma demande** : GET `/teams/:id/join-requests/mine` retourne la demande si elle existe, ou `null`/`404` s'il n'y en a pas.

6. **Annuler ma demande d'adhésion** : DELETE `/teams/:id/join-requests/mine` annule la demande (status='cancelled').

### Cas nominaux
- Utilisateur non-membre demande à rejoindre → demande créée (pending)
- Capitaine voit la liste des demandes en attente
- Capitaine accepte → utilisateur devient membre
- Capitaine refuse → demande marquée 'declined'
- Utilisateur voit "Votre demande est en attente" sur la page de l'équipe
- Utilisateur annule sa demande → status='cancelled'

### Cas limites
- Utilisateur déjà membre → erreur "already_member"
- Demande déjà existante → erreur "request_already_exists"
- Demande introuvable → 404 "request_not_found"
- Capitaine tente de modifier une demande non-pending → erreur "request_already_processed"
- Utilisateur qui n'est pas l'auteur tente d'annuler → 403 "not_your_request"
- Capitaine/cocapitaine seulement peut lister/accepter/refuser

### Gestion d'erreurs

| Condition | Code HTTP | `error` | `message` |
|-----------|-----------|---------|-----------|
| Déjà membre | 400 | `already_member` | "Vous êtes déjà membre de cette équipe" |
| Demande déjà existante | 400 | `request_already_exists` | "Vous avez déjà une demande en attente pour cette équipe" |
| Demande introuvable | 404 | `request_not_found` | "Demande d'adhésion introuvable" |
| Demande déjà traitée | 400 | `request_already_processed` | "Cette demande a déjà été traitée" |
| Pas l'auteur de la demande | 403 | `not_your_request` | "Vous n'êtes pas l'auteur de cette demande" |
| Pas les droits | 403 | `insufficient_permission` | "Vous n'avez pas les droits pour gérer les demandes" |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/team_join_request.go` | Créer | Handler complet des demandes d'adhésion |

### Signatures

```go
type TeamJoinRequestHandler struct {
    joinRequestRepo models.TeamJoinRequestRepository
    teamMemberRepo  models.TeamMemberRepository
    teamRepo        models.TeamRepository
}

func NewTeamJoinRequestHandler(
    joinRequestRepo models.TeamJoinRequestRepository,
    teamMemberRepo  models.TeamMemberRepository,
    teamRepo        models.TeamRepository,
) *TeamJoinRequestHandler

func (h *TeamJoinRequestHandler) RequestJoin(c *gin.Context)
func (h *TeamJoinRequestHandler) ListPendingRequests(c *gin.Context)
func (h *TeamJoinRequestHandler) AcceptRequest(c *gin.Context)
func (h *TeamJoinRequestHandler) DeclineRequest(c *gin.Context)
func (h *TeamJoinRequestHandler) GetMyRequest(c *gin.Context)
func (h *TeamJoinRequestHandler) CancelMyRequest(c *gin.Context)
```

### Détails des endpoints

#### `RequestJoin` - POST `/teams/:id/join-requests`
```json
// Request body: vide (ou ignoré)

// Response 201
{
    "id": 1,
    "team_id": 1,
    "user_id": 3,
    "status": "pending",
    "created_at": "2026-05-09T01:00:00Z"
}

// Response 400 si déjà membre
{ "error": "already_member", "message": "Vous êtes déjà membre de cette équipe" }
```

#### `ListPendingRequests` - GET `/teams/:id/join-requests`
```json
// Response 200
{
    "requests": [
        {
            "id": 1,
            "user_id": 3,
            "display_name": "User3",
            "status": "pending",
            "created_at": "2026-05-09T01:00:00Z"
        }
    ]
}
```

#### `AcceptRequest` - PUT `/teams/:id/join-requests/:requestId/accept`
```json
// Response 200
{ "message": "Demande acceptée", "user_id": 3 }
```

#### `DeclineRequest` - PUT `/teams/:id/join-requests/:requestId/decline`
```json
// Response 200
{ "message": "Demande refusée", "user_id": 3 }
```

#### `GetMyRequest` - GET `/teams/:id/join-requests/mine`
```json
// Response 200 (si une demande existe)
{
    "id": 1,
    "team_id": 1,
    "user_id": 3,
    "status": "pending",
    "created_at": "2026-05-09T01:00:00Z"
}

// Response 404 si pas de demande
{ "error": "request_not_found", "message": "Aucune demande d'adhésion trouvée" }
```

#### `CancelMyRequest` - DELETE `/teams/:id/join-requests/mine`
```json
// Response 200
{ "message": "Demande annulée" }

// Response 403 si l'utilisateur n'est pas l'auteur de la demande (vérification par user ID)
```

### Contraintes techniques
- **Pattern** : Même pattern que `team_invitation.go`
- **Vérification** : `RequestJoin` doit vérifier que l'utilisateur n'est pas déjà membre AVANT de créer la demande
- **Middleware** : Les endpoints de listing/accept/decline utilisent le middleware `RequireTeamRole` avec les rôles `captain`, `cocaptain`
- **RequireAuthOrMember** : Pour les endpoints `RequestJoin`, `GetMyRequest` et `CancelMyRequest`, l'utilisateur n'est pas encore membre de l'équipe. Utiliser le nouveau middleware `middleware.RequireAuthOrMember(teamMemberRepo, userRepo, jwtService)` disponible après la **Tâche #005**. Ce middleware :
  - Authentifie l'utilisateur (comme `RequireAuth`)
  - Tente de trouver l'utilisateur comme membre de l'équipe (via `c.Param("id")`)
  - Si trouvé, stocke `team_member` dans le contexte (comme `RequireTeamRole`)
  - Si NON trouvé (pas membre), continue sans erreur (contrairement à `RequireTeamRole` qui rejette)
  - L'utilisateur est toujours accessible via `middleware.MustGetUser(c)`
  - Vérifier l'appartenance avec `middleware.GetTeamMember(c)` (retourne nil si pas membre)
- **Route pour `GetMyRequest` / `CancelMyRequest`** : Ces routes fixes (`/mine`) doivent être déclarées AVANT les routes paramétrées (`/:requestId/accept`) dans le groupe Gin. Gin priorise les routes fixes, donc l'ordre n'est pas critique mais par clarté les mettre en premier.

### Tests à implémenter

#### Tests d'intégration
- **Fichier** : `internal/handlers/team_join_request_test.go`
- Scénario 1 : Demander à rejoindre (non-membre) → 201
  - Setup : Créer user A (capitaine) + équipe, user B (non-membre)
  - POST `/api/v1/teams/1/join-requests` par user B
  - Vérifier : 201, body a `status: "pending"`
- Scénario 2 : Demander à rejoindre (déjà membre) → 400 + "already_member"
- Scénario 3 : Capitaine liste les demandes → 200
  - Vérifier : la liste contient la demande de user B
- Scénario 4 : Capitaine accepte → 200 + B devient membre
  - Vérifier : B peut maintenant accéder à `GET /teams/:id` (normalement nécessite d'être membre)
- Scénario 5 : Capitaine refuse → 200 + demande 'declined'
- Scénario 6 : Utilisateur voit sa demande → GET `/teams/:id/join-requests/mine` → 200
- Scénario 7 : Utilisateur annule sa demande → DELETE `/teams/:id/join-requests/mine` → 200
- Scénario 8 : Demander en doublon → 400 + "request_already_exists"
- Scénario 9 : Membre (pas captain) tente de lister/accept/decline → 403

### Documentation
- Aucune documentation spécifique pour cette tâche
