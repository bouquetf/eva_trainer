# Tâche #003 - Story #006 : Handlers d'invitation (par pseudo et par lien)

## Objectif
Créer le handler HTTP pour gérer les invitations : inviter un utilisateur par display_name, générer un lien d'invitation, lister les invitations en attente, accepter/refuser une invitation, et traiter les invitations par lien.

## Contexte
- Story #006 : Team Invitation and Join System
- Dépend de : Tâche #002 (Modèles)
- Nécessaire pour : Tâche #005 (Routes)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un handler `TeamInvitationHandler` avec les méthodes suivantes :

1. **Inviter par pseudo** : Le capitaine/cocapitaine POSTe sur `/teams/:id/invitations` avec `{"display_name": "..."}`. L'API cherche l'utilisateur par display_name OU discord_username (insensible à la casse). Si trouvé et pas déjà membre, crée une invitation. Si introuvable, retourne 404.

2. **Générer un lien d'invitation** : Le capitaine/cocapitaine POSTe sur `/teams/:id/invitations/link`. L'API génère un token unique, crée une invitation avec une validité de 24h, et retourne le lien complet : `{frontend_url}/invite/{token}`.

3. **Accepter une invitation (par pseudo)** : L'utilisateur POSTe sur `/invitations/:invitationId/accept`. L'invitation est marquée 'accepted' et l'utilisateur est ajouté à l'équipe comme 'member'.

4. **Refuser une invitation (par pseudo)** : L'utilisateur POSTe sur `/invitations/:invitationId/decline`. L'invitation est marquée 'declined'.

5. **Accepter une invitation par lien** : L'utilisateur POSTe sur `/api/v1/invite-links/:token/accept`. L'API vérifie le token, vérifie que le lien n'est pas expiré, et ajoute l'utilisateur à l'équipe. Si expiré → message spécifique "Ce lien d'invitation a expiré."

6. **Voir les informations d'une invitation par lien** : GET `/api/v1/invite-links/:token` retourne les infos de base (team_id, team_name, expires_at) sans authentification requise. Utilisé par la page frontend `/invite/:token` pour pré-afficher les infos.

7. **Lister mes invitations en attente** : GET `/invitations` retourne les invitations en attente pour l'utilisateur connecté (affiché sur la page d'accueil).

8. **Lister les invitations en attente d'une équipe** : GET `/teams/:id/invitations/pending` retourne les invitations en attente (capitaine/cocapitaine seulement).

9. **Rechercher des utilisateurs (autocomplétion)** : GET `/api/v1/users/search?q=...` retourne une liste d'utilisateurs dont le display_name ou discord_username correspond à la query (limité à 10 résultats). Utilisé par l'autocomplétion du champ d'invitation.

### Cas nominaux
- Capitaine invite "MonPseudo" → invitation créée, statut pending
- Capitaine génère un lien → token + URL retournée
- Utilisateur invité par pseudo voit l'invitation dans sa liste → peut accepter
- Utilisateur clique sur un lien valide → accepté et devient membre
- Capitaine voit la liste des invitations en attente

### Cas limites
- Pseudo inexistant (display_name ET discord_username) → 404 "user_not_found"
- Utilisateur déjà membre → 400 "already_member" avec message "Vous êtes déjà membre de cette équipe"
- Lien expiré → 410 "invitation_expired" avec message spécifique
- Lien déjà utilisé → 400 "invitation_already_used"
- Invitation déjà acceptée/refusée → 400 "invitation_already_used"
- S'inviter soi-même → 400 "cannot_invite_self"
- Recherche autocomplétion vide → liste vide (200 OK)

### Gestion d'erreurs
| Condition | Code HTTP | `error` | `message` |
|-----------|-----------|---------|-----------|
| Display name introuvable | 404 | `user_not_found` | "User not found" |
| Déjà membre | 400 | `already_member` | "Vous êtes déjà membre de cette équipe" |
| Lien expiré | 410 | `invitation_expired` | "Ce lien d'invitation a expiré." |
| Lien déjà utilisé | 400 | `invitation_already_used` | "Cette invitation a déjà été utilisée." |
| Invitation introuvable | 404 | `invitation_not_found` | "Invitation not found" |
| Auto-invitation | 400 | `cannot_invite_self` | "Vous ne pouvez pas vous inviter vous-même" |
| Pas les droits | 403 | `insufficient_role` | "Vous n'avez pas les droits pour inviter des membres" |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/team_invitation.go` | Créer | Handler complet des invitations |

### Signatures

```go
type TeamInvitationHandler struct {
    teamInvitationRepo models.TeamInvitationRepository
    teamMemberRepo     models.TeamMemberRepository
    userRepo           models.UserRepository
    teamRepo           models.TeamRepository
}

func NewTeamInvitationHandler(
    teamInvitationRepo models.TeamInvitationRepository,
    teamMemberRepo     models.TeamMemberRepository,
    userRepo           models.UserRepository,
    teamRepo           models.TeamRepository,
) *TeamInvitationHandler

func (h *TeamInvitationHandler) InviteByDisplayName(c *gin.Context)
func (h *TeamInvitationHandler) GenerateInviteLink(c *gin.Context)
func (h *TeamInvitationHandler) AcceptInvitation(c *gin.Context)
func (h *TeamInvitationHandler) DeclineInvitation(c *gin.Context)
func (h *TeamInvitationHandler) AcceptInviteLink(c *gin.Context)
func (h *TeamInvitationHandler) GetInviteLinkInfo(c *gin.Context)
func (h *TeamInvitationHandler) ListMyInvitations(c *gin.Context)
func (h *TeamInvitationHandler) ListPendingByTeam(c *gin.Context)

// NOUVEAU - Recherche d'utilisateurs pour autocomplétion
func (h *TeamInvitationHandler) SearchUsers(c *gin.Context)
```

### Détails des endpoints

#### `InviteByDisplayName` - POST `/teams/:id/invitations`
```json
// Request
{ "display_name": "MonPseudo" }

// Response 201
{
    "id": 1,
    "team_id": 1,
    "invited_user_id": 2,
    "status": "pending",
    "expires_at": "2026-05-10T01:00:00Z",
    "created_at": "2026-05-09T01:00:00Z"
}
```

#### `GenerateInviteLink` - POST `/teams/:id/invitations/link`
```json
// Request body: vide

// Response 201
{
    "id": 2,
    "team_id": 1,
    "token": "a1b2c3d4e5f6...",
    "invite_url": "https://example.com/invite/a1b2c3d4e5f6...",
    "expires_at": "2026-05-10T01:00:00Z",
    "created_at": "2026-05-09T01:00:00Z"
}
```

#### `AcceptInvitation` - POST `/invitations/:invitationId/accept`
```json
// Response 200
{ "message": "Vous avez rejoint l'équipe", "team_id": 1, "team_name": "Mon Équipe" }
```

#### `DeclineInvitation` - POST `/invitations/:invitationId/decline`
```json
// Response 200
{ "message": "Invitation refusée" }
```

#### `AcceptInviteLink` - POST `/api/v1/invite-links/:token/accept`
```json
// Response 200
{ "message": "Vous avez rejoint l'équipe", "team_id": 1, "team_name": "Mon Équipe" }

// Si expiré → 410
{ "error": "invitation_expired", "message": "Ce lien d'invitation a expiré." }
```

#### `GetInviteLinkInfo` - GET `/api/v1/invite-links/:token`
```json
// Response 200 (pas d'auth requise)
{
    "team_id": 1,
    "team_name": "Mon Équipe",
    "expires_at": "2026-05-10T01:00:00Z",
    "is_expired": false
}
```

#### `SearchUsers` - GET `/api/v1/users/search?q=...`
```json
// Response 200
{
    "users": [
        { "id": 1, "display_name": "MonPseudo", "discord_username": "MonDiscord#1234" },
        { "id": 2, "display_name": "MonAutre", "discord_username": "Autre#5678" }
    ]
}

// Query string: ?q=Mon (recherche insensible à la casse sur display_name ET discord_username)
// Limité à 10 résultats max
```

#### `ListMyInvitations` - GET `/invitations`
```json
// Response 200
[
    {
        "id": 1,
        "team_id": 1,
        "team_name": "Mon Équipe",
        "status": "pending",
        "expires_at": "2026-05-10T01:00:00Z",
        "created_at": "2026-05-09T01:00:00Z"
    }
]
```

#### `ListPendingByTeam` - GET `/teams/:id/invitations/pending`
```json
// Response 200
{
    "invitations": [
        {
            "id": 1,
            "invited_user_id": 2,
            "invited_user_display_name": "MonPseudo",
            "invited_by_user_id": 1,
            "invited_by_display_name": "Capitaine",
            "status": "pending",
            "expires_at": "2026-05-10T01:00:00Z",
            "created_at": "2026-05-09T01:00:00Z"
        }
    ]
}
```

### Contraintes techniques
- **Pattern** : Même pattern que `team_roles.go` (struct + constructeur + méthodes)
- **Format erreur** : Utiliser le format `{"error": "code", "message": "..."}` comme partout
- **Génération de token** : 32 bytes aléatoires, encodés en hexadécimal (64 caractères)
- **Frontend URL** : Lire depuis `config.AppConfig.Server.FrontendURL` pour construire le lien d'invitation
- **Injection** : Pour récupérer `FrontendURL` dans le handler, soit le passer dans le constructeur, soit lire depuis `config.AppConfig`
- **Vérification** : Quand l'utilisateur accepte une invitation par lien, vérifier l'expiration AVANT de faire quoi que ce soit d'autre
- **URL du lien** : Construire comme `{frontend_url}/invite/{token}` → ex: `http://localhost:5173/invite/abc123`
- **Recherche d'utilisateurs** : `SearchUsers` doit chercher avec `ILIKE` sur `display_name` ET `discord_username` (OR), limité à 10 résultats. Ne pas retourner l'utilisateur connecté lui-même.
- **`InviteByDisplayName`** : Utiliser `FindByDisplayNameOrDiscordUsername` qui cherche sur les deux champs

### Tests à implémenter

#### Tests d'intégration
- **Fichier** : `internal/handlers/team_invitation_test.go`
- Scénario 1 : Inviter un utilisateur par display_name (capitaine) → 201 + invitation créée
  - Setup : Créer un user avec display_name, créer une équipe, se connecter en tant que capitaine
  - POST `/api/v1/teams/1/invitations` avec `{"display_name": "TestUser"}`
  - Vérifier : status 201, body a `id`, `team_id`, `invited_user_id`, `status: "pending"`
- Scénario 2 : Inviter un utilisateur inexistant → 404
- Scénario 3 : Inviter quelqu'un déjà membre → 400 + "already_member"
- Scénario 4 : Générer un lien d'invitation → 201 + token + invite_url
- Scénario 5 : Accepter une invitation par lien valide → 200 + membre ajouté
- Scénario 6 : Accepter une invitation par lien expiré → 410 + message spécifique
- Scénario 7 : Capitaine liste les invitations en attente → 200 + liste
- Scénario 8 : Membre liste ses invitations → 200 + liste
- Scénario 9 : Utilisateur non connecté voit les infos d'un lien → 200 avec team_name, is_expired
- Scénario 10 : Résister à une invitation sans être l'utilisateur ciblé → 403
- Scénario 11 : Recherche d'utilisateurs (autocomplétion) → GET `/api/v1/users/search?q=Test` → 200 + liste d'utilisateurs
- Scénario 12 : Recherche sans résultats → 200 + `{"users": []}`

### Documentation
- Aucune documentation spécifique pour cette tâche (les endpoints seront documentés dans la tâche #009)
