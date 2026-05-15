# Tâche #002 - Story #022 : Migrer les invitations en attente de Home vers TeamsList

## Objectif
Déplacer l'affichage des invitations en attente (acceptation/refus) de la page Home (qui n'est plus accessible) vers la page TeamsList, afin de ne pas perdre cette fonctionnalité.

## Contexte
- Story #022 : `docs/stories/story-022.md`
- Dépend de : Tâche #001 (la route `/` ne rend plus Home)
- Nécessaire pour : Tâche #003 (les tests E2E doivent passer)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

La page Home (`frontend/src/pages/Home.tsx`) affiche actuellement les invitations en attente de l'utilisateur avec des boutons "Accepter" et "Refuser". Comme la route `/` ne rend plus Home (voir tâche #001), il faut déplacer cette fonctionnalité dans TeamsList (`frontend/src/pages/TeamsList.tsx`).

**Cas nominaux :**
- L'utilisateur connecté a des invitations en attente → elles sont affichées en haut de la page TeamsList, avant la section "Mes équipes" et avant la section "Découvrir des équipes"
- L'utilisateur n'a pas d'invitations → aucune section d'invitations n'est affichée
- L'utilisateur clique "Accepter" → l'invitation disparaît, l'équipe apparaît dans la liste
- L'utilisateur clique "Refuser" → l'invitation disparaît

**Cas limites :**
- L'utilisateur n'est membre d'aucune équipe et a des invitations → les invitations sont affichées au-dessus de l'état vide et de la section Découvrir

**Gestion d'erreurs :**
- Erreur API lors de l'acceptation/refus → identique au comportement actuel dans Home (TanStack Query gère via `onError`)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/TeamsList.tsx` | Modifier | Ajouter le bloc d'invitations en haut du rendu |
| `frontend/src/pages/Home.tsx` | Supprimer | Fichier orphelin — tout le code utile a été migré vers TeamsList |

### Détails des modifications

#### `frontend/src/pages/TeamsList.tsx`

Ajouter dans le composant `TeamsList` :

1. **Imports** à ajouter en haut :
```tsx
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { listMyInvitations, acceptInvitation, declineInvitation } from '../api/invitations'
import type { TeamInvitation } from '../api/invitations'
```

2. **Requête des invitations** (inspirée de Home.tsx) :
```tsx
const { data: myInvitations = [] } = useQuery({
  queryKey: ['my-invitations'],
  queryFn: listMyInvitations,
})
```

(Note: `isAuthenticated` n'est pas directement disponible dans TeamsList. On peut soit importer `useAuthStore`, soit se fier au fait que TeamsList est déjà protégée par AuthGuard. Voir question en fin de fichier. **Solution retenue** : pas besoin de `enabled`, car TeamsList est toujours appelée dans un contexte authentifié.)

3. **Mutations** (inspirée de Home.tsx) :
```tsx
const queryClient = useQueryClient()

const acceptMutation = useMutation({
  mutationFn: (invitationId: number) => acceptInvitation(invitationId),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['my-invitations'] })
    queryClient.invalidateQueries({ queryKey: ['my-teams'] })
  },
})

const declineMutation = useMutation({
  mutationFn: (invitationId: number) => declineInvitation(invitationId),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['my-invitations'] })
  },
})
```

4. **Bloc d'affichage** — Ajouter ce bloc **juste après le `<div>` d'ouverture du rendu principal et avant le `{!teams || teams.length === 0 ? ...}`** :

```tsx
{myInvitations.length > 0 && (
  <div className="section">
    <h2 className="section-title">Invitations en attente ({myInvitations.length})</h2>
    <div>
      {myInvitations.map(inv => (
        <div key={inv.id} className="invitation-card">
          <div className="invitation-info">
            <span className="invitation-team-name">{inv.team_name || `Équipe #${inv.team_id}`}</span>
            <span className="invitation-date">
              Expire le {new Date(inv.expires_at).toLocaleDateString()}
            </span>
          </div>
          <div className="invitation-actions">
            <button
              className="btn-primary btn-small"
              onClick={() => acceptMutation.mutate(inv.id)}
              disabled={acceptMutation.isPending}
            >
              Accepter
            </button>
            <button
              className="btn-secondary btn-small"
              onClick={() => declineMutation.mutate(inv.id)}
              disabled={declineMutation.isPending}
            >
              Refuser
            </button>
          </div>
        </div>
      ))}
    </div>
  </div>
)}
```

#### `frontend/src/pages/Home.tsx`

**Supprimer** le fichier après la migration. Il s'inspire du code de `Home.tsx` pour le copier dans `TeamsList.tsx`, puis supprime le fichier orphelin.

### Signatures (non applicable)

### Contraintes techniques
- **Framework** : React 19 + TanStack Query v5 (React Query)
- **Style** : Utiliser les classes CSS existantes (`.section`, `.section-title`, `.invitation-card`, `.btn-primary`, `.btn-secondary`, `.btn-small`, etc.) qui sont déjà définies dans les styles globaux (`auth.css`, `teams.css`)
- **Pattern** : Suivre le même pattern que Home.tsx pour les mutations et les invalidations de cache
- **Tests** : Vérifier manuellement que TeamsList fonctionne avec et sans invitations

### Tests à implémenter

Aucun test spécifique à cette tâche. Les tests de la tâche #003 couvriront la non-régression.

### Documentation

Aucune documentation à créer.
