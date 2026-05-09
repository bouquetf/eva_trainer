# Tâche #007 - Story #006 : Frontend UI - Section invitations dans TeamSettings

## Objectif
Ajouter une section "Invitations" dans la page de paramètres d'équipe (`TeamSettings.tsx`) permettant au capitaine/cocapitaine d'inviter des membres par display_name, de générer des liens d'invitation, et de voir les invitations en attente.

## Contexte
- Story #006 : Team Invitation and Join System
- Dépend de : Tâche #006 (Frontend API)
- Nécessaire pour : Aucune (dernière tâche UI)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Modifier `TeamSettings.tsx` pour ajouter une section "Invitations" avec :

1. **Inviter par pseudo avec autocomplétion** : Un champ de texte avec autocomplétion qui recherche les utilisateurs via `searchUsers()` au fur et à mesure de la saisie (debounce 300ms). Affiche une dropdown avec les résultats (display_name + discord_username). Le capitaine/cocapitaine sélectionne un utilisateur dans la liste et clique sur Inviter.
2. **Générer un lien d'invitation** : Un bouton "Générer un lien". Le lien généré est affiché dans un champ de texte en lecture seule avec un bouton "Copier".
3. **Invitations en attente** : Un tableau/liste des invitations en attente pour cette équipe, avec le statut et la date d'expiration.

### Cas nominaux
- Capitaine saisit un pseudo valide → message "Invitation envoyée à X"
- Capitaine clique "Générer un lien" → le lien apparaît, cliquable/copiable
- La liste des invitations en attente s'affiche

### Cas limites
- Pseudo inexistant → message d'erreur rouge "Utilisateur introuvable"
- Pseudo déjà membre → message "Déjà membre"
- Lien expiré dans la liste → affiché avec un badge "Expiré"
- Erreur réseau → message d'erreur générique

### Gestion d'erreurs
- La mutation TanStack Query gère les erreurs via `onError`
- Les messages d'erreur sont affichés dans une `div.error-message`
- Les succès sont affichés dans une `div.success-message`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/TeamSettings.tsx` | Modifier | Ajouter la section invitations |
| `frontend/src/styles/teams.css` | Modifier | Styles pour les nouveaux composants (si nécessaire) |

### Composants à ajouter dans TeamSettings

#### Sous-section "Inviter un membre"
```tsx
// Ajouter après la section "Membres"
{canManageRoles && (
  <div className="section">
    <h2 className="section-title">Invitations</h2>

    {/* Inviter par pseudo avec autocomplétion */}
    <div style={{ marginBottom: 16 }}>
      <h3>Inviter par pseudo</h3>
      <div style={{ position: 'relative' }}>
        <input
          type="text"
          value={inviteDisplayName}
          onChange={handleInviteInputChange}
          onFocus={() => setShowSuggestions(true)}
          onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
          placeholder="Rechercher un utilisateur..."
          disabled={inviteMutation.isPending}
        />
        {showSuggestions && searchResults.length > 0 && (
          <div className="autocomplete-dropdown">
            {searchResults.map(user => (
              <div
                key={user.id}
                className="autocomplete-item"
                onClick={() => {
                  setInviteDisplayName(user.display_name || user.discord_username)
                  setSelectedUserId(user.id)
                  setShowSuggestions(false)
                }}
              >
                <span className="autocomplete-name">{user.display_name || user.discord_username}</span>
                {user.discord_username && (
                  <span className="autocomplete-discord">{user.discord_username}</span>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
      <div style={{ marginTop: 8 }}>
        <button
          className="btn-primary"
          onClick={handleInvite}
          disabled={!inviteDisplayName.trim() || inviteMutation.isPending}
        >
          Inviter
        </button>
      </div>
      {inviteError && <div className="error-message" style={{ marginTop: 8 }}>{inviteError}</div>}
      {inviteSuccess && <div className="success-message" style={{ marginTop: 8 }}>{inviteSuccess}</div>}
    </div>

    {/* Générer un lien d'invitation */}
    <div style={{ marginBottom: 16 }}>
      <h3>Lien d'invitation</h3>
      <button className="btn-secondary" onClick={handleGenerateLink} disabled={linkMutation.isPending}>
        Générer un lien
      </button>
      {inviteLink && (
        <div style={{ marginTop: 8, display: 'flex', gap: 8 }}>
          <input type="text" value={inviteLink} readOnly style={{ flex: 1 }} />
          <button className="btn-secondary" onClick={() => { navigator.clipboard.writeText(inviteLink) }}>
            Copier
          </button>
        </div>
      )}
      {linkError && <div className="error-message" style={{ marginTop: 8 }}>{linkError}</div>}
    </div>

    {/* Invitations en attente */}
    <div>
      <h3>Invitations en attente</h3>
      {pendingInvitations.length === 0 ? (
        <p style={{ color: '#888', fontSize: 14 }}>Aucune invitation en attente.</p>
      ) : (
        <table className="invitations-table">
          <thead>
            <tr>
              <th>Utilisateur</th>
              <th>Type</th>
              <th>Statut</th>
              <th>Expire le</th>
            </tr>
          </thead>
          <tbody>
            {pendingInvitations.map(inv => (
              <tr key={inv.id}>
                <td>{inv.invited_user_display_name || 'Lien générique'}</td>
                <td>{inv.token ? 'Lien' : 'Pseudo'}</td>
                <td>
                  <span className={`status-badge ${inv.status}`}>
                    {inv.status === 'expired' ? 'Expiré' : 'En attente'}
                  </span>
                </td>
                <td>{new Date(inv.expires_at).toLocaleDateString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  </div>
)}
```

#### États du composant (à ajouter dans TeamSettings)
```typescript
// Imports à ajouter
import { searchUsers } from '../api/invitations'
import type { UserSearchResult } from '../api/invitations'

// Invitation state
const [inviteDisplayName, setInviteDisplayName] = useState('')
const [selectedUserId, setSelectedUserId] = useState<number | null>(null)
const [inviteError, setInviteError] = useState<string | null>(null)
const [inviteSuccess, setInviteSuccess] = useState<string | null>(null)

// Autocomplete state
const [showSuggestions, setShowSuggestions] = useState(false)
const [searchResults, setSearchResults] = useState<UserSearchResult[]>([])

// Invite link state
const [inviteLink, setInviteLink] = useState<string | null>(null)
const [linkError, setLinkError] = useState<string | null>(null)

// Debounce ref for search
const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

// Handler for invite input change with autocomplete
const handleInviteInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
  const value = e.target.value
  setInviteDisplayName(value)
  setSelectedUserId(null)

  // Debounce search
  if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current)
  if (value.trim().length < 2) {
    setSearchResults([])
    return
  }
  searchTimeoutRef.current = setTimeout(async () => {
    try {
      const res = await searchUsers(value.trim())
      setSearchResults(res.users)
      setShowSuggestions(true)
    } catch {
      setSearchResults([])
    }
  }, 300)
}

// Mutations
const inviteMutation = useMutation({
  mutationFn: () => inviteByDisplayName(teamId, { display_name: inviteDisplayName.trim() }),
  onSuccess: () => {
    setInviteSuccess(`Invitation envoyée à ${inviteDisplayName.trim()}`)
    setInviteDisplayName('')
    setInviteError(null)
    setSearchResults([])
    queryClient.invalidateQueries({ queryKey: ['pending-invitations', teamId] })
  },
  onError: (err: unknown) => {
    const apiErr = err as { status?: number; body?: { message?: string } }
    setInviteError(apiErr?.body?.message ?? "Erreur lors de l'invitation.")
    setInviteSuccess(null)
  },
})

const linkMutation = useMutation({
  mutationFn: () => generateInviteLink(teamId),
  onSuccess: (data) => {
    setInviteLink(data.invite_url ?? null)
    setLinkError(null)
    queryClient.invalidateQueries({ queryKey: ['pending-invitations', teamId] })
  },
  onError: () => setLinkError('Erreur lors de la génération du lien.'),
})

// Query pour les invitations en attente
const { data: pendingInvData } = useQuery({
  queryKey: ['pending-invitations', teamId],
  queryFn: () => listPendingInvitations(teamId),
  enabled: canManageRoles && !isNaN(teamId),
})
const pendingInvitations = pendingInvData?.invitations ?? []
```

#### Handlers
```typescript
const handleInvite = () => {
  if (!inviteDisplayName.trim()) return
  inviteMutation.mutate()
}

const handleGenerateLink = () => {
  linkMutation.mutate()
}
```

### Styles à ajouter dans `teams.css`

```css
/* Autocomplete dropdown */
.autocomplete-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: var(--bg-secondary, #1a1a2e);
  border: 1px solid var(--border-color, #333);
  border-radius: 4px;
  max-height: 200px;
  overflow-y: auto;
  z-index: 1000;
}

.autocomplete-item {
  padding: 8px 12px;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.autocomplete-item:hover {
  background: var(--bg-hover, #2a2a3e);
}

.autocomplete-name {
  font-weight: 500;
}

.autocomplete-discord {
  color: #888;
  font-size: 12px;
}

/* Invitations table */
.invitations-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 8px;
}

.invitations-table th,
.invitations-table td {
  padding: 8px 12px;
  text-align: left;
  border-bottom: 1px solid var(--border-color, #333);
}

.invitations-table th {
  color: #888;
  font-size: 12px;
  text-transform: uppercase;
}

.status-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.status-badge.pending {
  background: #2d4a2d;
  color: #4caf50;
}

.status-badge.expired {
  background: #4a2d2d;
  color: #f44336;
}
```

### Contraintes techniques
- **Framework** : React + TanStack Query (React Query) + TypeScript
- **State management** : Utiliser `useState` local (pas besoin de Zustand pour ce composant)
- **Pattern** : Suivre exactement le pattern de `TeamSettings.tsx` : `useMutation`, `useQuery`, `invalidateQueries`
- **Validation** : Désactiver le bouton "Inviter" si le champ est vide ou si la mutation est en cours
- **Copie** : `navigator.clipboard.writeText()` pour copier le lien (pas de dépendance externe)

### Tests

Aucun test spécifique pour cette tâche (les tests E2E seront écrits dans la story).

### Documentation
- Aucune documentation spécifique pour cette tâche
