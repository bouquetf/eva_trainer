import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listEvaTeams, createEvaTeam, updateEvaTeam, deleteEvaTeam, listTrainerTeams,
} from '../api/archives'
import type { EvaTeam, CreateEvaTeamPayload } from '../api/archives'

export default function EvaTeamsSection({ canWrite }: { canWrite: boolean }) {
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const [showEvaForm, setShowEvaForm] = useState(false)
  const [editingEvaTeam, setEditingEvaTeam] = useState<EvaTeam | null>(null)
  const [evaFormName, setEvaFormName] = useState('')
  const [evaFormTag, setEvaFormTag] = useState('')
  const [evaFormTrainerTeamId, setEvaFormTrainerTeamId] = useState<number | ''>('')
  const [deletingEvaTeamId, setDeletingEvaTeamId] = useState<number | null>(null)

  const { data: evaTeamsData, isLoading: evaTeamsLoading } = useQuery({
    queryKey: ['archive-eva-teams'],
    queryFn: () => listEvaTeams(),
    staleTime: 30_000,
  })

  const evaTeams = evaTeamsData?.teams ?? []

  const { data: trainerTeamsData } = useQuery({
    queryKey: ['trainer-teams'],
    queryFn: () => listTrainerTeams(),
    staleTime: 60_000,
  })

  const trainerTeams = trainerTeamsData?.teams ?? []

  const createEvaMutation = useMutation({
    mutationFn: (data: CreateEvaTeamPayload) => createEvaTeam(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-eva-teams'] })
      setShowEvaForm(false)
      setEditingEvaTeam(null)
      setSuccess('Équipe eva créée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la création")
    },
  })

  const updateEvaMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: CreateEvaTeamPayload }) => updateEvaTeam(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-eva-teams'] })
      setShowEvaForm(false)
      setEditingEvaTeam(null)
      setSuccess('Équipe eva modifiée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la modification")
    },
  })

  const deleteEvaMutation = useMutation({
    mutationFn: (id: number) => deleteEvaTeam(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-eva-teams'] })
      setDeletingEvaTeamId(null)
      setSuccess('Équipe eva supprimée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la suppression")
      setDeletingEvaTeamId(null)
    },
  })

  function openCreateEvaForm() {
    setEditingEvaTeam(null)
    setEvaFormName('')
    setEvaFormTag('')
    setEvaFormTrainerTeamId('')
    setShowEvaForm(true)
  }

  function openEditEvaForm(team: EvaTeam) {
    setEditingEvaTeam(team)
    setEvaFormName(team.name)
    setEvaFormTag(team.tag)
    setEvaFormTrainerTeamId(team.eva_trainer_team_id ?? '')
    setShowEvaForm(true)
  }

  function handleEvaSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    const payload: CreateEvaTeamPayload = {
      name: evaFormName,
      tag: evaFormTag,
      eva_trainer_team_id: evaFormTrainerTeamId === '' ? null : Number(evaFormTrainerTeamId),
    }
    if (editingEvaTeam) {
      updateEvaMutation.mutate({ id: editingEvaTeam.id, data: payload })
    } else {
      createEvaMutation.mutate(payload)
    }
  }

  function handleDeleteEvaConfirm() {
    if (deletingEvaTeamId !== null) {
      deleteEvaMutation.mutate(deletingEvaTeamId)
    }
  }

  function findTrainerTeamName(trainerTeamId: number | null): string {
    if (trainerTeamId === null) return 'Aucune'
    const found = trainerTeams.find(t => t.team.id === trainerTeamId)
    return found ? found.team.name : `#${trainerTeamId}`
  }

  if (evaTeamsLoading) {
    return <p>Chargement...</p>
  }

  return (
    <>
      {error && <div className="error-message" onClick={() => setError(null)}>{error}</div>}
      {success && <div className="success-message" onClick={() => setSuccess(null)}>{success}</div>}

      <div className="admin-header" style={{ marginTop: 0 }}>
        {canWrite && (
          <button className="btn-primary" onClick={openCreateEvaForm}>
            Nouvelle équipe eva
          </button>
        )}
      </div>

      {evaTeams.length === 0 ? (
        <div className="empty-state"><p>Aucune équipe eva</p></div>
      ) : (
        <table className="admin-table">
          <thead>
            <tr>
              <th>Nom</th>
              <th>Tag</th>
              <th>Équipe eva-trainer</th>
              {canWrite && <th>Actions</th>}
            </tr>
          </thead>
          <tbody>
            {evaTeams.map(team => (
              <tr key={team.id}>
                <td>{team.name}</td>
                <td><code>{team.tag}</code></td>
                <td>{findTrainerTeamName(team.eva_trainer_team_id)}</td>
                {canWrite && (
                  <td>
                    <button className="btn-secondary btn-small" onClick={() => openEditEvaForm(team)}>Modifier</button>
                    <button className="btn-danger btn-small" style={{ marginLeft: 8 }} onClick={() => setDeletingEvaTeamId(team.id)}>Supprimer</button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {showEvaForm && (
        <div className="modal-overlay" onClick={() => { setShowEvaForm(false); setEditingEvaTeam(null) }}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>{editingEvaTeam ? 'Modifier l\'équipe eva' : 'Nouvelle équipe eva'}</h3>
            <form onSubmit={handleEvaSubmit}>
              <div className="form-group">
                <label htmlFor="eva-team-name">Nom</label>
                <input id="eva-team-name" type="text" value={evaFormName} onChange={e => setEvaFormName(e.target.value)} required />
              </div>
              <div className="form-group">
                <label htmlFor="eva-team-tag">Tag</label>
                <input id="eva-team-tag" type="text" value={evaFormTag} onChange={e => setEvaFormTag(e.target.value)} required maxLength={50} />
              </div>
              <div className="form-group">
                <label htmlFor="eva-team-trainer">Équipe eva-trainer associée</label>
                <select id="eva-team-trainer" className="role-select" value={evaFormTrainerTeamId} onChange={e => setEvaFormTrainerTeamId(e.target.value === '' ? '' : Number(e.target.value))}>
                  <option value="">Aucune</option>
                  {trainerTeams.map(item => (
                    <option key={item.team.id} value={item.team.id}>
                      {item.team.name} (ID: {item.team.id})
                    </option>
                  ))}
                </select>
              </div>
              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => { setShowEvaForm(false); setEditingEvaTeam(null) }}>Annuler</button>
                <button type="submit" className="btn-primary" disabled={createEvaMutation.isPending || updateEvaMutation.isPending}>
                  {createEvaMutation.isPending || updateEvaMutation.isPending
                    ? 'En cours...'
                    : editingEvaTeam ? 'Modifier' : 'Créer'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {deletingEvaTeamId !== null && (
        <div className="modal-overlay" onClick={() => setDeletingEvaTeamId(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>Confirmer la suppression</h3>
            <p>Êtes-vous sûr de vouloir supprimer cette équipe eva ?</p>
            <div className="modal-actions">
              <button className="btn-secondary" onClick={() => setDeletingEvaTeamId(null)}>Annuler</button>
              <button className="btn-danger" onClick={handleDeleteEvaConfirm} disabled={deleteEvaMutation.isPending}>
                {deleteEvaMutation.isPending ? 'Suppression...' : 'Supprimer'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
