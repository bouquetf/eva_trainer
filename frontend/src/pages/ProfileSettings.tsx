import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { updateDisplayName, fetchMe } from '../api/auth'
import { useAuthStore } from '../store/auth'

function ProfileSettings() {
  const queryClient = useQueryClient()
  const [editingDisplayName, setEditingDisplayName] = useState(false)
  const [displayNameValue, setDisplayNameValue] = useState('')
  const [success, setSuccess] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const { data: user } = useQuery({
    queryKey: ['me'],
    queryFn: fetchMe,
  })

  const displayNameMutation = useMutation({
    mutationFn: () => updateDisplayName(displayNameValue.trim()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me'] })
      useAuthStore.setState(state => ({
        user: state.user ? { ...state.user, display_name: displayNameValue.trim() } : null,
      }))
      setEditingDisplayName(false)
      setSuccess('Nom d\'affichage mis à jour.')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la mise à jour.')
    },
  })

  return (
    <div>
      <h1>Paramètres du profil</h1>
      {success && <div className="success-message">{success}</div>}
      {error && <div className="error-message">{error}</div>}

      <div className="section">
        <h2 className="section-title">Mon nom d'affichage</h2>
        {editingDisplayName ? (
          <form onSubmit={e => { e.preventDefault(); displayNameMutation.mutate() }}>
            <div className="form-group">
              <label htmlFor="display-name">Nom affiché sur le site</label>
              <input
                id="display-name"
                type="text"
                value={displayNameValue}
                onChange={e => setDisplayNameValue(e.target.value)}
                disabled={displayNameMutation.isPending}
                autoFocus
              />
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button type="submit" className="btn-primary" disabled={!displayNameValue.trim() || displayNameMutation.isPending}>
                {displayNameMutation.isPending ? 'Enregistrement...' : 'Enregistrer'}
              </button>
              <button type="button" className="btn-secondary" onClick={() => setEditingDisplayName(false)}>Annuler</button>
            </div>
          </form>
        ) : (
          <div>
            <p style={{ margin: '0 0 12px', color: '#aaa' }}>
              {user?.display_name || `#${user?.id ?? ''}`}
            </p>
            <button className="btn-secondary" onClick={() => { setDisplayNameValue(user?.display_name || ''); setEditingDisplayName(true) }}>
              Modifier
            </button>
          </div>
        )}
      </div>

      <div className="section">
        <h2 className="section-title">Compte Discord</h2>
        <p style={{ margin: '0 0 12px', color: '#aaa' }}>
          Pseudo Discord : <strong>{user?.discord_username || user?.discord_id || '—'}</strong>
        </p>
      </div>
    </div>
  )
}

export default ProfileSettings
