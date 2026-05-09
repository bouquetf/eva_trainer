# Tâche #008 - Story #006 : Frontend UI - Demandes d'adhésion et découverte d'équipes

## Objectif
Ajouter dans TeamSettings la section "Demandes d'adhésion" pour les capitaines/cocapitaines (voir, accepter, refuser), et sur la page de découverte des équipes permettre à un utilisateur de demander à rejoindre une équipe, voir le statut de sa demande, et l'annuler.

## Contexte
- Story #006 : Team Invitation and Join System
- Dépend de : Tâche #006 (Frontend API), Tâche #007 (Section invitations)
- Nécessaire pour : Aucune

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie A - Section "Demandes en attente" dans TeamSettings

Ajouter une nouvelle section dans `TeamSettings.tsx` (visible uniquement pour captain/cocaptain) qui affiche la liste des demandes d'adhésion en attente. Pour chaque demande, afficher le display_name du demandeur, la date de la demande, et des boutons "Accepter" / "Refuser".

#### Partie B - Interface "Rejoindre une équipe" sur la page des équipes

Sur la page `TeamsList.tsx`, ajouter une section qui liste les équipes que l'utilisateur ne rejoint PAS encore. Pour chaque équipe non-membre, afficher un bouton "Demander à rejoindre". Si l'utilisateur a déjà une demande en attente, afficher "Demande envoyée" avec un bouton "Annuler".

### Cas nominaux
- Capitaine voit les demandes en attente et peut accepter/refuser
- Utilisateur voit les équipes disponibles et peut demander à rejoindre
- Après acceptation, l'utilisateur voit l'équipe dans sa liste
- Après refus, la demande disparaît de la liste du capitaine

### Cas limites
- Aucune demande en attente → message "Aucune demande en attente"
- Utilisateur déjà membre → pas de bouton "Rejoindre" visible
- Demande déjà envoyée → afficher "Demande en attente" avec possibilité d'annuler
- Erreur réseau → message d'erreur

### Gestion d'erreurs
- Mutation `acceptJoinRequest` : en cas d'erreur, afficher le message d'erreur
- Mutation `declineJoinRequest` : idem
- Mutation `requestJoinTeam` : si déjà membre → message spécifique
- Mutation `cancelMyJoinRequest` : idem

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/TeamSettings.tsx` | Modifier | Ajouter section demandes en attente |
| `frontend/src/pages/TeamsList.tsx` | Modifier | Ajouter section découverte d'équipes |
| `frontend/src/pages/Home.tsx` | Modifier | Ajouter section invitations en attente (accepter/refuser) |
| `frontend/src/pages/InviteAccept.tsx` | Créer | Nouvelle page `/invite/:token` |
| `frontend/src/routes/index.tsx` | Modifier | Ajouter route `/invite/:token` |
| `frontend/src/styles/teams.css` | Modifier | Styles additionnels |

### Partie A - Section dans TeamSettings

Ajouter après la section "Invitations" dans `TeamSettings.tsx` :

```tsx
{canManageRoles && (
  <div className="section">
    <h2 className="section-title">Demandes d'adhésion ({pendingRequests.length})</h2>
    {pendingRequests.length === 0 ? (
      <p style={{ color: '#888', fontSize: 14 }}>Aucune demande en attente.</p>
    ) : (
      <div>
        {pendingRequests.map(req => (
          <div key={req.id} className="join-request-item">
            <div className="join-request-info">
              <span className="join-request-user">
                {req.display_name || `#${req.user_id}`}
              </span>
              <span className="join-request-date">
                {new Date(req.created_at).toLocaleDateString()}
              </span>
            </div>
            <div className="join-request-actions">
              <button
                className="btn-primary btn-small"
                onClick={() => acceptRequestMutation.mutate(req.id)}
                disabled={acceptRequestMutation.isPending}
              >
                Accepter
              </button>
              <button
                className="btn-danger btn-small"
                onClick={() => declineRequestMutation.mutate(req.id)}
                disabled={declineRequestMutation.isPending}
              >
                Refuser
              </button>
            </div>
          </div>
        ))}
      </div>
    )}
  </div>
)}
```

#### États et mutations à ajouter dans TeamSettings

```typescript
// Join requests query
const { data: joinRequestsData } = useQuery({
  queryKey: ['pending-join-requests', teamId],
  queryFn: () => listPendingJoinRequests(teamId),
  enabled: canManageRoles && !isNaN(teamId),
})
const pendingRequests = joinRequestsData?.requests ?? []

// Accept request mutation
const acceptRequestMutation = useMutation({
  mutationFn: (requestId: number) => acceptJoinRequest(teamId, requestId),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['pending-join-requests', teamId] })
    queryClient.invalidateQueries({ queryKey: ['team-members', teamId] })
    queryClient.invalidateQueries({ queryKey: ['team', teamId] })
  },
  onError: () => {
    queryClient.invalidateQueries({ queryKey: ['pending-join-requests', teamId] })
  },
})

// Decline request mutation
const declineRequestMutation = useMutation({
  mutationFn: (requestId: number) => declineJoinRequest(teamId, requestId),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['pending-join-requests', teamId] })
  },
  onError: () => {
    queryClient.invalidateQueries({ queryKey: ['pending-join-requests', teamId] })
  },
})
```

### Partie B - Modification de TeamsList.tsx

```tsx
// Ajouter les imports
import { listAllTeams } from '../api/teams'
import type { TeamListResponse, TeamListItem } from '../api/teams'
import {
  requestJoinTeam,
  getMyJoinRequest,
  cancelMyJoinRequest,
} from '../api/joinRequests'

// Dans le composant TeamsList, AJOUTER après la liste des équipes de l'utilisateur :

function TeamsList() {
  // ... existing code ...

  // Découverte d'équipes (utilise listAllTeams de task-006)
  const { data: allTeamsData, isLoading: allTeamsLoading } = useQuery({
    queryKey: ['all-teams'],
    queryFn: () => listAllTeams({ page: 1, per_page: 50 }),
  })

  // Équipes dont l'utilisateur n'est PAS membre
  const userTeamIds = new Set(teams?.map(t => t.team.id) ?? [])
  const discoverableTeams = (allTeamsData?.teams ?? []).filter(t => !userTeamIds.has(t.id))

  // Join request state par équipe
  const [joinRequestStatus, setJoinRequestStatus] = useState<Record<number, string>>({})

  const requestJoinMutation = useMutation({
    mutationFn: (teamId: number) => requestJoinTeam(teamId),
    onSuccess: (_data, teamId) => {
      setJoinRequestStatus(prev => ({ ...prev, [teamId]: 'pending' }))
    },
    onError: (err: unknown, teamId) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setJoinRequestStatus(prev => ({ ...prev, [teamId]: apiErr?.body?.message ?? 'Erreur' }))
    },
  })

  const cancelMutation = useMutation({
    mutationFn: (teamId: number) => cancelMyJoinRequest(teamId),
    onSuccess: (_data, teamId) => {
      setJoinRequestStatus(prev => {
        const next = { ...prev }
        delete next[teamId]
        return next
      })
    },
  })

  return (
    <div>
      {/* Section existing: Mes équipes */}
      {/* ... */}

      {/* Section: Découvrir des équipes */}
      <div className="section">
        <h2 className="section-title">Découvrir des équipes</h2>
        {allTeamsLoading ? (
          <div className="spinner" />
        ) : discoverableTeams.length === 0 ? (
          <p style={{ color: '#888', fontSize: 14 }}>Aucune autre équipe disponible.</p>
        ) : (
          <div>
            {discoverableTeams.map(team => (
              <div key={team.id} className="discover-team-card">
                <div className="discover-team-info">
                  <h3>{team.name}</h3>
                  <span>{team.member_count} membre{team.member_count > 1 ? 's' : ''}</span>
                </div>
                <div className="discover-team-actions">
                  {joinRequestStatus[team.id] === 'pending' ? (
                    <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                      <span style={{ color: '#888', fontSize: 14 }}>Demande envoyée</span>
                      <button
                        className="btn-danger btn-small"
                        onClick={() => cancelMutation.mutate(team.id)}
                        disabled={cancelMutation.isPending}
                      >
                        Annuler
                      </button>
                    </div>
                  ) : (
                    <button
                      className="btn-primary"
                      onClick={() => requestJoinMutation.mutate(team.id)}
                      disabled={requestJoinMutation.isPending}
                    >
                      Demander à rejoindre
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
```

#### Composant enfant pour le statut de demande persisté

Pour éviter de perdre le statut de la demande au refresh (si on utilise juste `useState`), idéalement on devrait charger le statut depuis l'API au montage du composant. Mais pour simplifier, on peut utiliser un pattern plus simple :

```tsx
// Au chargement, pour chaque équipe "discoverable", on peut charger le statut
// Mais ce serait trop de requêtes. À la place, on utilise un endpoint qui retourne
// les demandes de l'utilisateur en cours.

const { data: myRequests } = useQuery({
  queryKey: ['my-join-requests'],
  queryFn: async () => {
    // Idéalement: GET /api/v1/join-requests (toutes mes demandes)
    // Mais on n'a pas cet endpoint pour l'instant.
    // Alternative: on utilise le state local et on vérifie au clic.
    return []
  },
  enabled: false, // Désactivé pour l'instant
})
```

**Note pour l'utilisateur** : Le statut de la demande est stocké dans le state local seulement. Si l'utilisateur rafraîchit la page, il perd la trace. Solution proposée dans `questions.md`.

### Partie C - Section invitations en attente dans Home.tsx

Ajouter dans `Home.tsx` une section qui liste les invitations en attente de l'utilisateur connecté, avec des boutons "Accepter" et "Refuser".

```tsx
// Imports à ajouter
import { listMyInvitations, acceptInvitation, declineInvitation } from '../api/invitations'
import type { TeamInvitation } from '../api/invitations'

// Dans le composant Home, ajouter après la section des équipes :
function Home() {
  // ... existing code (teams list, etc.) ...

  // Invitations query
  const { data: myInvitations = [] } = useQuery({
    queryKey: ['my-invitations'],
    queryFn: listMyInvitations,
  })

  const acceptMutation = useMutation({
    mutationFn: (invitationId: number) => acceptInvitation(invitationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-invitations'] })
      queryClient.invalidateQueries({ queryKey: ['teams'] })
    },
  })

  const declineMutation = useMutation({
    mutationFn: (invitationId: number) => declineInvitation(invitationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-invitations'] })
    },
  })

  return (
    <div>
      {/* ... existing content ... */}

      {/* Section: Mes invitations */}
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
    </div>
  )
}
```

### Partie D - Page InviteAccept (`/invite/:token`)

Créer la page frontend dédiée à l'acceptation d'une invitation par lien :

```tsx
// frontend/src/pages/InviteAccept.tsx
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getInviteLinkInfo, acceptInviteLink } from '../api/invitations'
import type { InviteLinkInfo } from '../api/invitations'

function InviteAccept() {
  const { token } = useParams<{ token: string }>()
  const queryClient = useQueryClient()

  // Charger les infos du lien
  const { data: linkInfo, isLoading, isError } = useQuery<InviteLinkInfo>({
    queryKey: ['invite-link', token],
    queryFn: () => getInviteLinkInfo(token!),
    enabled: !!token,
  })

  const acceptMutation = useMutation({
    mutationFn: () => acceptInviteLink(token!),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['teams'] })
      queryClient.invalidateQueries({ queryKey: ['my-invitations'] })
    },
  })

  if (isLoading) return <div className="page-container"><div className="spinner" /></div>

  if (isError || !linkInfo) {
    return (
      <div className="page-container">
        <div className="error-card">
          <h2>Lien invalide</h2>
          <p>Ce lien d'invitation n'existe pas ou a expiré.</p>
        </div>
      </div>
    )
  }

  if (linkInfo.is_expired) {
    return (
      <div className="page-container">
        <div className="error-card">
          <h2>Lien expiré</h2>
          <p>Ce lien d'invitation a expiré. Contactez le capitaine de l'équipe pour obtenir un nouveau lien.</p>
        </div>
      </div>
    )
  }

  if (acceptMutation.isSuccess) {
    return (
      <div className="page-container">
        <div className="success-card">
          <h2>Vous avez rejoint l'équipe !</h2>
          <p>Vous êtes maintenant membre de <strong>{acceptMutation.data.team_name}</strong>.</p>
          <a href="/" className="btn-primary">Retour à l'accueil</a>
        </div>
      </div>
    )
  }

  return (
    <div className="page-container">
      <div className="invite-card">
        <h2>Invitation à rejoindre une équipe</h2>
        <p>Vous êtes invité à rejoindre <strong>{linkInfo.team_name}</strong>.</p>
        <p className="invite-expiry">
          Ce lien expire le {new Date(linkInfo.expires_at).toLocaleDateString()}.
        </p>
        {acceptMutation.isError && (
          <div className="error-message">
            {(acceptMutation.error as { message?: string })?.message ?? "Erreur lors de l'acceptation."}
          </div>
        )}
        <button
          className="btn-primary"
          onClick={() => acceptMutation.mutate()}
          disabled={acceptMutation.isPending}
        >
          {acceptMutation.isPending ? "Acceptation..." : "Accepter l'invitation"}
        </button>
      </div>
    </div>
  )
}
```

#### Route frontend à ajouter dans `frontend/src/routes/index.tsx`

```tsx
import InviteAccept from '../pages/InviteAccept'

// Ajouter dans la définition des routes :
<Route path="/invite/:token" element={<InviteAccept />} />
```

### Contraintes techniques
- **Pattern** : Même pattern que les autres sections de `TeamSettings.tsx`
- **Mutations** : Utiliser `useMutation` avec invalidation des queries
- **Design** : Les écarts de style sont gérés via les classes CSS existantes
- **Boutons** : `btn-small` est une classe à ajouter dans `teams.css` pour les petits boutons dans les listes
- **Requêtes** : `listAllTeams` a besoin d'un nouvel endpoint backend → voir `questions.md`

### Styles à ajouter

```css
.join-request-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-color, #333);
}

.join-request-user {
  font-weight: 500;
}

.join-request-date {
  color: #888;
  font-size: 12px;
  margin-left: 12px;
}

.join-request-actions {
  display: flex;
  gap: 8px;
}

.btn-small {
  padding: 4px 12px;
  font-size: 12px;
}

.discover-team-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color, #333);
}

.discover-team-info h3 {
  margin: 0 0 4px;
  font-size: 16px;
}

.discover-team-info span {
  color: #888;
  font-size: 13px;
}
```

### Tests
- Aucun test spécifique pour cette tâche (les tests E2E couvriront les flows)

### Documentation
- Aucune documentation spécifique pour cette tâche
