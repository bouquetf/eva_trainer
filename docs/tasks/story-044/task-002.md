# Tâche #002 - Story #044 : Endpoint CancelInvitation (Backend)

## Objectif
Créer l'endpoint `DELETE /api/v1/teams/:id/invitations/:invitationId` permettant au capitaine/capitaine d'annuler (cancel) une invitation en attente.

## Contexte
- Story #044 : `docs/stories/story-044.md`
- Dépend de : Story #006, Tâche #001 (structure de routes similaire)
- Nécessaire pour : Tâche #003 (frontend), Tâche #004 (tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Un nouvel endpoint `DELETE /api/v1/teams/:id/invitations/:invitationId` qui annule une invitation en attente en marquant son statut `cancelled`.

**Cas nominaux :**
- Capitaine ou cocapitaine annule une invitation en attente → statut passé à `cancelled`
- L'invitation avec un `token` (lien générique) est aussi annulée → le lien devient inutilisable

**Cas limites :**
- Invitation déjà acceptée/expirée/annulée → erreur (le repo retourne `ErrInvitationNotFound` si `status != 'pending'`)
- Invitation inexistante → erreur 404
- Invitation d'une autre équipe (sécurisé par le fait que la route est sous `/teams/:id/invitations/`)

**Gestion d'erreurs :**
- Invitation non trouvée ou déjà traitée → `invitation_not_found` (404)
- Pas capitaine/cocapitaine → `insufficient_role` (403, géré par le middleware)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/team_invitation.go` | Modifier | Ajouter méthode `CancelInvitation` |
| `api/routes.go` | Modifier | Ajouter route `DELETE /:id/invitations/:invitationId` (dans le bloc `teams`) |

### Signature

```go
// Dans internal/handlers/team_invitation.go
func (h *TeamInvitationHandler) CancelInvitation(c *gin.Context) {
    // ...
}
```

### Contraintes techniques
- **Framework** : Go / Gin
- **Pattern** : Suivre exactement le pattern de `ListPendingByTeam` dans `team_invitation.go` (mêmes validations, même style de réponse)
- **Repository** : Utiliser `h.teamInvitationRepo.CancelInvitation(invitationID)` qui existe déjà
- **Sécurité** : Vérifier que `inv.TeamID == teamID` — renvoyer 404 `invitation_not_found` si mismatch (évite qu'un capitaine annule une invitation d'une autre équipe en devinant l'ID)
- **Route role** : `middleware.RequireTeamRole(logger, teamMemberRepo, "captain", "cocaptain")`
- **Route permission** : `middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "team:invite")` (permission déjà existante, utilisée par les autres endpoints d'invitation)
- **Méthode HTTP** : `DELETE` (cohérent avec le reste de l'API)

### Tests à implémenter

#### Tests d'intégration
- **Fichier** : `internal/handlers/team_invitation_test.go`
- Scénario 1 : Capitaine annule une invitation user-to-user en attente → 200, statut `cancelled`
- Scénario 2 : Capitaine annule un lien d'invitation → 200
- Scénario 3 : Cocapitaine annule une invitation → 200
- Scénario 4 : Annulation d'une invitation déjà acceptée → 404
- Scénario 5 : Annulation d'une invitation inexistante → 404
- Scénario 6 : Invitation d'une autre équipe → 404 (team_id mismatch)
