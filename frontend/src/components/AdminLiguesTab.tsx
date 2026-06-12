import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listLigues, createLigue, refreshLigue, updateLigue, deleteLigue } from '../api/archives'
import { listSplits } from '../api/splits'
import type { Split } from '../api/splits'
import type { LigueListItem } from '../api/archives'

type AdminLiguesTabProps = {
  onSuccess: (message: string) => void
  onError: (message: string) => void
}

function AdminLiguesTab({ onSuccess, onError }: AdminLiguesTabProps) {
  const queryClient = useQueryClient()
  const [liguesPage, setLiguesPage] = useState(1)
  const [showCreateLigueForm, setShowCreateLigueForm] = useState(false)
  const [createLigueNom, setCreateLigueNom] = useState('')
  const [createLigueDebut, setCreateLigueDebut] = useState('')
  const [createLigueFin, setCreateLigueFin] = useState('')
  const [createLigueSplitId, setCreateLigueSplitId] = useState<number | ''>('')

  const [editingLigue, setEditingLigue] = useState<LigueListItem | null>(null)
  const [editNom, setEditNom] = useState('')
  const [editDebut, setEditDebut] = useState('')
  const [editFin, setEditFin] = useState('')
  const [editSplitId, setEditSplitId] = useState<number | ''>('')

  const [deletingLigueId, setDeletingLigueId] = useState<number | null>(null)

  const { data: liguesData, isLoading: liguesLoading } = useQuery({
    queryKey: ['admin-ligues', liguesPage],
    queryFn: () => listLigues(liguesPage, 5),
    staleTime: 10_000,
  })

  const { data: splitsData } = useQuery({
    queryKey: ['admin-splits'],
    queryFn: () => listSplits(),
    staleTime: 60_000,
  })
  const splits: Split[] = splitsData?.splits ?? []

  const createLigueMutation = useMutation({
    mutationFn: (data: { nom: string; date_debut: string; date_fin: string; split_id: number }) => createLigue(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-ligues'] })
      queryClient.invalidateQueries({ queryKey: ['archive-ligues-import'] })
      setShowCreateLigueForm(false)
      setCreateLigueNom('')
      setCreateLigueDebut('')
      setCreateLigueFin('')
      setCreateLigueSplitId('')
      onSuccess('Ligue créée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      onError(apiErr?.body?.message ?? "Erreur lors de la création")
    },
  })

  const updateLigueMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { nom: string; date_debut: string; date_fin: string; split_id: number } }) => updateLigue(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-ligues'] })
      queryClient.invalidateQueries({ queryKey: ['archive-ligues-import'] })
      setEditingLigue(null)
      onSuccess('Ligue modifiée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      onError(apiErr?.body?.message ?? "Erreur lors de la modification")
    },
  })

  const deleteLigueMutation = useMutation({
    mutationFn: (id: number) => deleteLigue(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-ligues'] })
      queryClient.invalidateQueries({ queryKey: ['archive-ligues-import'] })
      setDeletingLigueId(null)
      onSuccess('Ligue supprimée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      onError(apiErr?.body?.message ?? "Erreur lors de la suppression")
      setDeletingLigueId(null)
    },
  })

  const refreshLigueMutation = useMutation({
    mutationFn: (id: number) => refreshLigue(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['admin-ligues'] })
      queryClient.invalidateQueries({ queryKey: ['archive-ligues-import'] })
      onSuccess(`Ligue rafraîchie : ${data.associated} locales associées`)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      onError(apiErr?.body?.message ?? "Erreur lors du rafraîchissement")
    },
  })

  const liguesTotalPages = liguesData ? Math.ceil(liguesData.total / liguesData.per_page) : 0

  function handleCreateLigue(e: React.FormEvent) {
    e.preventDefault()
    if (createLigueSplitId === '') return
    createLigueMutation.mutate({
      nom: createLigueNom,
      date_debut: createLigueDebut,
      date_fin: createLigueFin,
      split_id: createLigueSplitId,
    })
  }

  function handleEditLigue(e: React.FormEvent) {
    e.preventDefault()
    if (!editingLigue || editSplitId === '') return
    updateLigueMutation.mutate({
      id: editingLigue.id,
      data: {
        nom: editNom,
        date_debut: editDebut,
        date_fin: editFin,
        split_id: editSplitId,
      },
    })
  }

  function openEditForm(l: LigueListItem) {
    setEditingLigue(l)
    setEditNom(l.nom)
    setEditDebut(l.date_debut)
    setEditFin(l.date_fin)
    setEditSplitId(l.split_id)
  }

  return (
    <>
      <div className="admin-header">
        <button className="btn-primary" onClick={() => setShowCreateLigueForm(true)}>
          Nouvelle ligue
        </button>
      </div>

      {liguesLoading ? (
        <p>Chargement...</p>
      ) : liguesData && liguesData.items.length > 0 ? (
        <>
          <table className="admin-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Nom</th>
                <th>Date début</th>
                <th>Date fin</th>
                <th>Split</th>
                <th>Locales</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {liguesData.items.map(l => (
                <tr key={l.id}>
                  <td>{l.id}</td>
                  <td>{l.nom}</td>
                  <td>{l.date_debut}</td>
                  <td>{l.date_fin}</td>
                  <td>{l.split_name}</td>
                  <td>{l.nb_locales}</td>
                  <td>
                    <button className="btn-secondary btn-small" onClick={() => openEditForm(l)} style={{ marginRight: 4 }}>
                      Modifier
                    </button>
                    <button className="btn-secondary btn-small" onClick={() => refreshLigueMutation.mutate(l.id)} disabled={refreshLigueMutation.isPending} style={{ marginRight: 4 }}>
                      {refreshLigueMutation.isPending ? '...' : 'Refresh'}
                    </button>
                    <button className="btn-danger btn-small" onClick={() => setDeletingLigueId(l.id)}>
                      Supprimer
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {liguesTotalPages > 1 && (
            <div className="pagination">
              <button className="btn-secondary btn-small" disabled={liguesPage <= 1} onClick={() => setLiguesPage(p => p - 1)}>Précédent</button>
              <span className="pagination-info">Page {liguesPage} / {liguesTotalPages}</span>
              <button className="btn-secondary btn-small" disabled={liguesPage >= liguesTotalPages} onClick={() => setLiguesPage(p => p + 1)}>Suivant</button>
            </div>
          )}
        </>
      ) : (
        <div className="empty-state"><p>Aucune ligue</p></div>
      )}

      {showCreateLigueForm && (
        <div className="modal-overlay" onClick={() => setShowCreateLigueForm(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>Nouvelle ligue</h3>
            <form onSubmit={handleCreateLigue}>
              <div className="form-group">
                <label htmlFor="ligue-nom">Nom</label>
                <input id="ligue-nom" type="text" value={createLigueNom} onChange={e => setCreateLigueNom(e.target.value)} required />
              </div>
              <div className="form-group">
                <label htmlFor="ligue-split">Split</label>
                <select id="ligue-split" className="role-select" value={createLigueSplitId} onChange={e => setCreateLigueSplitId(e.target.value === '' ? '' : Number(e.target.value))} required>
                  <option value="">Sélectionner un split</option>
                  {splits.map(s => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label htmlFor="ligue-debut">Date début</label>
                <input id="ligue-debut" type="date" value={createLigueDebut} onChange={e => setCreateLigueDebut(e.target.value)} required />
              </div>
              <div className="form-group">
                <label htmlFor="ligue-fin">Date fin</label>
                <input id="ligue-fin" type="date" value={createLigueFin} onChange={e => setCreateLigueFin(e.target.value)} required />
              </div>
              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => setShowCreateLigueForm(false)}>Annuler</button>
                <button type="submit" className="btn-primary" disabled={createLigueMutation.isPending}>
                  {createLigueMutation.isPending ? 'Création...' : 'Créer'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {editingLigue && (
        <div className="modal-overlay" onClick={() => setEditingLigue(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>Modifier la ligue #{editingLigue.id}</h3>
            <form onSubmit={handleEditLigue}>
              <div className="form-group">
                <label htmlFor="ligue-edit-nom">Nom</label>
                <input id="ligue-edit-nom" type="text" value={editNom} onChange={e => setEditNom(e.target.value)} required />
              </div>
              <div className="form-group">
                <label htmlFor="ligue-edit-split">Split</label>
                <select id="ligue-edit-split" className="role-select" value={editSplitId} onChange={e => setEditSplitId(e.target.value === '' ? '' : Number(e.target.value))} required>
                  <option value="">Sélectionner un split</option>
                  {splits.map(s => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label htmlFor="ligue-edit-debut">Date début</label>
                <input id="ligue-edit-debut" type="date" value={editDebut} onChange={e => setEditDebut(e.target.value)} required />
              </div>
              <div className="form-group">
                <label htmlFor="ligue-edit-fin">Date fin</label>
                <input id="ligue-edit-fin" type="date" value={editFin} onChange={e => setEditFin(e.target.value)} required />
              </div>
              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => setEditingLigue(null)}>Annuler</button>
                <button type="submit" className="btn-primary" disabled={updateLigueMutation.isPending}>
                  {updateLigueMutation.isPending ? 'Modification...' : 'Modifier'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {deletingLigueId !== null && (
        <div className="modal-overlay" onClick={() => setDeletingLigueId(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>Confirmer la suppression</h3>
            <p>Êtes-vous sûr de vouloir supprimer cette ligue ? Cela supprimera également les ligues locales associées.</p>
            <div className="modal-actions">
              <button className="btn-secondary" onClick={() => setDeletingLigueId(null)}>Annuler</button>
              <button className="btn-danger" onClick={() => deleteLigueMutation.mutate(deletingLigueId)} disabled={deleteLigueMutation.isPending}>
                {deleteLigueMutation.isPending ? 'Suppression...' : 'Supprimer'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

export default AdminLiguesTab
