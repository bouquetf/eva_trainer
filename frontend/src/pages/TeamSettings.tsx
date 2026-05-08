import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getTeam, updateTeam, deleteTeam, leaveTeam } from '../api/teams'
import { useAuthStore } from '../store/auth'
import ConfirmDialog from '../components/ConfirmDialog'

function TeamSettings() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const teamId = Number(id)

  const [editName, setEditName] = useState('')
  const [editing, setEditing] = useState(false)
  const [success, setSuccess] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [confirmLeave, setConfirmLeave] = useState(false)
  const [deleteName, setDeleteName] = useState('')

  const { data, isLoading, error: fetchError } = useQuery({
    queryKey: ['team', teamId],
    queryFn: () => getTeam(teamId),
    enabled: !isNaN(teamId),
  })

  const updateMutation = useMutation({
    mutationFn: () => updateTeam(teamId, { name: editName.trim() }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team', teamId] })
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
      setEditing(false)
      setSuccess('Nom de l\'équipe mis à jour.')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la mise à jour.')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteTeam(teamId, { confirm_with_name: deleteName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
      navigate('/teams')
    },
    onError: (err: unknown) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la suppression.')
      setConfirmDelete(false)
    },
  })

  const leaveMutation = useMutation({
    mutationFn: () => leaveTeam(teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
      navigate('/teams')
    },
    onError: (err: unknown) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur.')
      setConfirmLeave(false)
    },
  })

  if (isLoading) {
    return <div className="spinner" />
  }

  if (fetchError || !data) {
    return <div className="error-message">Équipe introuvable.</div>
  }

  const { team, members } = data
  const currentUser = useAuthStore.getState().user
  const isCaptain = currentUser !== null && team.captain_id === currentUser.id

  const handleStartEdit = () => {
    setEditName(team.name)
    setEditing(true)
  }

  const handleCancelEdit = () => {
    setEditing(false)
    setEditName('')
  }

  const handleDeleteClick = () => {
    setDeleteName('')
    setConfirmDelete(true)
  }

  return (
    <div>
      <h1>{team.name}</h1>
      {success && <div className="success-message">{success}</div>}
      {error && <div className="error-message">{error}</div>}

      <div className="section">
        <h2 className="section-title">Informations</h2>
        {editing ? (
          <form onSubmit={e => { e.preventDefault(); updateMutation.mutate() }}>
            <div className="form-group">
              <label htmlFor="edit-name">Nom de l'équipe</label>
              <input
                id="edit-name"
                type="text"
                value={editName}
                onChange={e => setEditName(e.target.value)}
                disabled={updateMutation.isPending}
                autoFocus
              />
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button type="submit" className="btn-primary" disabled={!editName.trim() || updateMutation.isPending}>
                {updateMutation.isPending ? 'Enregistrement...' : 'Enregistrer'}
              </button>
              <button type="button" className="btn-secondary" onClick={handleCancelEdit}>Annuler</button>
            </div>
          </form>
        ) : (
          <div>
            <p style={{ margin: '0 0 12px', color: '#aaa' }}>{team.name}</p>
            {isCaptain && (
              <button className="btn-secondary" onClick={handleStartEdit}>Modifier</button>
            )}
          </div>
        )}
      </div>

      <div className="section">
        <h2 className="section-title">Membres ({members.length})</h2>
        {members.map(m => (
          <div key={m.user_id} className="member-item">
            <span className="member-user-id">{m.display_name || `#${m.user_id}`}</span>
            <span className={`role-badge ${m.role}`}>{m.role}</span>
          </div>
        ))}
      </div>

      {!isCaptain && (
        <div className="section">
          <h2 className="section-title">Quitter l'équipe</h2>
          <p style={{ color: '#888', fontSize: 14, marginBottom: 12 }}>
            Vous allez quitter définitivement cette équipe.
          </p>
          <button className="btn-danger" onClick={() => setConfirmLeave(true)}>
            Quitter l'équipe
          </button>
        </div>
      )}

      {isCaptain && (
        <div className="section danger">
          <h2 className="section-title">Zone de danger</h2>
          <p style={{ color: '#888', fontSize: 14, marginBottom: 12 }}>
            Supprimer l'équipe effacera toutes les zones associées. Cette action est irréversible.
          </p>
          <div className="form-group">
            <label htmlFor="delete-confirm">Tapez le nom de l'équipe <strong>{team.name}</strong> pour confirmer</label>
            <input
              id="delete-confirm"
              type="text"
              value={deleteName}
              onChange={e => setDeleteName(e.target.value)}
              placeholder={team.name}
            />
          </div>
          <button
            className="btn-danger"
            disabled={deleteName !== team.name || deleteMutation.isPending}
            onClick={handleDeleteClick}
          >
            {deleteMutation.isPending ? 'Suppression...' : 'Supprimer l\'équipe'}
          </button>
        </div>
      )}

      <ConfirmDialog
        open={confirmDelete}
        title="Supprimer l'équipe"
        message="Êtes-vous sûr de vouloir supprimer cette équipe ? Cette action est irréversible."
        confirmLabel="Supprimer"
        danger
        onConfirm={() => deleteMutation.mutate()}
        onCancel={() => setConfirmDelete(false)}
      />

      <ConfirmDialog
        open={confirmLeave}
        title="Quitter l'équipe"
        message="Êtes-vous sûr de vouloir quitter cette équipe ?"
        confirmLabel="Quitter"
        danger
        onConfirm={() => leaveMutation.mutate()}
        onCancel={() => setConfirmLeave(false)}
      />
    </div>
  )
}

export default TeamSettings
