# Tâche #001 - Story #044 : Endpoint RemoveMember (Backend)

## Objectif
Créer l'endpoint `DELETE /api/v1/teams/:id/members/:userId` permettant au capitaine/cocapitaine de supprimer un membre de l'équipe.

## Contexte
- Story #044 : `docs/stories/story-044.md`
- Dépend de : Story #003 (système d'équipe existant), Story #006 (système d'invitation)
- Nécessaire pour : Tâche #003 (frontend), Tâche #004 (tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Un nouvel endpoint `DELETE /api/v1/teams/:id/members/:userId` qui retire un membre de l'équipe.

**Cas nominaux :**
- Capitaine supprime un membre → le membre est retiré de `team_members`
- Cocapitaine supprime un membre → le membre est retiré de `team_members`

**Cas limites :**
- Auto-suppression : le capitaine ou cocapitaine ne peut pas se supprimer lui-même
- Suppression du capitaine : impossible (doit transfert la captaincy d'abord)
- Suppression d'un admin global : impossible (cohérent avec `UpdateRole`)
- Cocapitaine supprime un autre cocapitaine : impossible (cohérent avec `UpdateRole`)
- Membre inexistant : erreur 404
- Équipe inexistante : erreur 404 (détecté par le middleware)

**Gestion d'erreurs :**
- Auto-suppression → `cannot_remove_self` (400)
- Suppression du capitaine → `cannot_remove_captain` (403)
- Suppression d'un admin → `cannot_remove_admin` (403)
- Cocapitaine supprime un autre cocapitaine → `insufficient_role` (403)
- Membre non trouvé → `member_not_found` (404)
- Pas capitaine/cocapitaine → `insufficient_role` (403, géré par le middleware)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/team_roles.go` | Modifier | Ajouter méthode `RemoveMember` |
| `api/routes.go` | Modifier | Ajouter route `DELETE /:id/members/:userId` |
*(aucune modification de permission nécessaire — réutilisation de `member:role:assign` déjà présente pour captain et cocaptain)*

### Signature

```go
// Dans internal/handlers/team_roles.go
func (h *TeamRolesHandler) RemoveMember(c *gin.Context) {
    // ...
}
```

### Contraintes techniques
- **Framework** : Go / Gin (comme le handler existant)
- **Pattern** : Suivre exactement le pattern de `UpdateRole` et `TransferCaptaincy` dans `team_roles.go`
- **Validation** : Utiliser `middleware.GetTeamMember(c)` pour récupérer le membre courant (mis en place par le middleware `RequireTeamRole`)
- **Repository** : Utiliser `h.teamMemberRepo.RemoveMember(teamID, userID)` qui existe déjà
- **Permission** : Utiliser `member:role:assign` (déjà présente pour captain et cocaptain)
- **Route role** : `middleware.RequireTeamRole(logger, teamMemberRepo, "captain", "cocaptain")`
- **Route permission** : `middleware.RequirePermission(logger, teamMemberRepo, rolePermRepo, permRepo, "member:role:assign")`

### Tests à implémenter

#### Tests d'intégration
- **Fichier** : `internal/handlers/team_roles_test.go` (existant)
- Scénario 1 : Capitaine supprime un membre → 200, membre retiré
- Scénario 2 : Cocapitaine supprime un membre → 200
- Scénario 3 : Auto-suppression → 400 `cannot_remove_self`
- Scénario 4 : Suppression du capitaine → 403 `cannot_remove_captain`
- Scénario 5 : Suppression d'un admin global → 403 `cannot_remove_admin`
- Scénario 6 : Cocapitaine supprime un autre cocapitaine → 403 `insufficient_role`
- Scénario 7 : Membre inexistant → 404 `member_not_found`
- Scénario 8 : Membre (pas capitaine/cocapitaine) tente de supprimer → 403 `insufficient_role`
