# Tâche #003 - Story #044 : Frontend — Suppression membre + Annulation invitation

## Objectif
Ajouter dans l'interface d'équipe (TeamSettings) la possibilité de supprimer un membre et d'annuler une invitation en attente.

## Contexte
- Story #044 : `docs/stories/story-044.md`
- Dépend de : Tâche #001 (endpoint RemoveMember), Tâche #002 (endpoint CancelInvitation)
- Nécessaire pour : Tâche #004 (tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. Suppression d'un membre
Ajouter un bouton "Supprimer" pour chaque membre (sauf soi-même, le capitaine, et les admins globaux) dans la liste des membres. Ce bouton ouvre un `ConfirmDialog` de confirmation.

**Cas nominaux :**
- Capitaine voit le bouton "Supprimer" pour tous les membres non-capitaine/non-admin/non-soi
- Cocapitaine voit le bouton "Supprimer" pour les membres (pas pour un autre cocapitaine, pas pour le capitaine, pas pour lui-même)
- Clic sur "Supprimer" → ConfirmDialog : "Confirmer la suppression de <username> de l'équipe ?"
- Confirmation → `DELETE /api/v1/teams/:id/members/:userId` → refresh de la liste

**Cas limites :**
- Auto-suppression : le bouton n'apparaît pas pour soi-même (condition `!isSelf` dans le JSX)
- Suppression du capitaine : impossible (bouton pas affiché pour le capitaine, condition `m.role !== 'captain'`)
- Suppression admin : impossible (bouton pas affiché pour les admins, condition `!isGlobalAdmin`)
- Cocapitaine ne peut pas supprimer un autre cocapitaine : le bouton "Supprimer" doit être dans le même bloc conditionnel que les boutons de changement de rôle : `(currentUserRole === 'captain' || m.role !== 'cocaptain')`
- Erreur API : afficher le message d'erreur dans un `error-message`

#### 2. Annulation d'une invitation en attente
Ajouter une colonne "Actions" dans le tableau des invitations en attente avec un bouton "Annuler".

**Cas nominaux :**
- Capitaine/cocapitaine voit un bouton "Annuler" pour chaque invitation en attente
- Clic sur "Annuler" → `DELETE /api/v1/teams/:id/invitations/:invitationId` → refresh de la liste des invitations

**Cas limites :**
- Invitation déjà expirée : pas de bouton (statut `expired` au lieu de `pending`)
- Erreur API : afficher le message d'erreur

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/roles.ts` | Modifier | Ajouter `removeMember(teamId, userId)` |
| `frontend/src/api/invitations.ts` | Modifier | Ajouter `cancelInvitation(teamId, invitationId)` |
| `frontend/src/components/MemberList.tsx` | Modifier | Ajouter bouton "Supprimer" + ConfirmDialog |
| `frontend/src/pages/TeamSettings.tsx` | Modifier | Ajouter colonne "Actions" + bouton "Annuler" dans le tableau des invitations |

### Signatures

```typescript
// Dans frontend/src/api/roles.ts
export async function removeMember(
  teamId: number,
  userId: number,
): Promise<{ message: string }>

// Dans frontend/src/api/invitations.ts
export async function cancelInvitation(
  teamId: number,
  invitationId: number,
): Promise<{ message: string }>
```

### Contraintes techniques
- **Framework** : React / TypeScript, React Query (@tanstack/react-query)
- **Pattern** : Suivre exactement le pattern existant dans `MemberList.tsx` :
  - Utiliser `useMutation` avec invalidation de `['team-members', teamId]` et `['team', teamId]`
  - Utiliser `ConfirmDialog` pour la confirmation (comme pour les changements de rôle)
  - Style : classes CSS existantes `btn-danger btn-small` pour le bouton supprimer
- **Pattern pour invitations** : Suivre le pattern existant dans `TeamSettings.tsx` :
  - Les mutations existantes (`acceptRequestMutation`, `declineRequestMutation`) servent de modèle
  - Utiliser `useMutation` avec invalidation de `['pending-invitations', teamId]`
- **Localisation** : Messages en français (comme le reste de l'interface)

### Tests à implémenter

#### Tests unitaires (si infrastructure existante)
#### Pas de tests unitaires frontend spécifiques requis — les tests E2E couvriront la fonctionnalité.
