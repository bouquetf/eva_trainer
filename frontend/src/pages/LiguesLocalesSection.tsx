import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listLiguesLocales, getLigueLocale, createLigueLocale, updateLigueLocale, deleteLigueLocale,
  listLigues, importLigue,
} from '../api/archives'
import type { LigueLocale, CreateLigueLocalePayload, LigueListItem } from '../api/archives'

function LigueSelectField({ ligues, onImport }: { ligues: LigueListItem[]; onImport: (ligueId: number) => void }) {
  const [selectedId, setSelectedId] = useState<number | ''>('')
  return (
    <div className="add-membership-row">
      <select className="role-select" value={selectedId} onChange={e => setSelectedId(Number(e.target.value) || '')}>
        <option value="">Choisir une ligue...</option>
        {ligues.map(l => (
          <option key={l.id} value={l.id}>
            #{l.id} — {l.nom} ({l.split_name}, {l.date_debut} → {l.date_fin})
          </option>
        ))}
      </select>
      <button className="btn-primary btn-small" disabled={selectedId === ''} onClick={() => onImport(Number(selectedId))}>
        Importer
      </button>
    </div>
  )
}

export default function LiguesLocalesSection({ canWrite }: { canWrite: boolean }) {
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const [showForm, setShowForm] = useState(false)
  const [editingLigueLocale, setEditingLigueLocale] = useState<LigueLocale | null>(null)
  const [formNom, setFormNom] = useState('')
  const [formEmplacement, setFormEmplacement] = useState('')
  const [deletingLigueLocaleId, setDeletingLigueLocaleId] = useState<number | null>(null)
  const [selectedLocaleId, setSelectedLocaleId] = useState<number | null>(null)
  const [importTarget, setImportTarget] = useState<{ localeId: number; ligueId: number } | null>(null)

  const { data: localesData, isLoading: localesLoading } = useQuery({
    queryKey: ['archive-ligues-locales'],
    queryFn: () => listLiguesLocales(),
    staleTime: 30_000,
  })

  const locales = localesData?.ligues_locales ?? []

  const { data: detailData, isLoading: detailLoading } = useQuery({
    queryKey: ['archive-ligue-locale-detail', selectedLocaleId],
    queryFn: () => getLigueLocale(selectedLocaleId!),
    enabled: selectedLocaleId !== null,
    staleTime: 10_000,
  })

  const { data: liguesForImport } = useQuery({
    queryKey: ['archive-ligues-import'],
    queryFn: () => listLigues(1, 100),
    enabled: selectedLocaleId !== null,
    staleTime: 30_000,
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateLigueLocalePayload) => createLigueLocale(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-ligues-locales'] })
      setShowForm(false)
      setEditingLigueLocale(null)
      setSuccess('Ligue locale créée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la création")
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: CreateLigueLocalePayload }) => updateLigueLocale(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-ligues-locales'] })
      setShowForm(false)
      setEditingLigueLocale(null)
      setSuccess('Ligue locale modifiée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la modification")
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteLigueLocale(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-ligues-locales'] })
      if (selectedLocaleId === deletingLigueLocaleId) {
        setSelectedLocaleId(null)
      }
      setDeletingLigueLocaleId(null)
      setSuccess('Ligue locale supprimée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la suppression")
      setDeletingLigueLocaleId(null)
    },
  })

  const importLigueMutation = useMutation({
    mutationFn: ({ localeId, ligueId }: { localeId: number; ligueId: number }) => importLigue(localeId, ligueId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-ligue-locale-detail', selectedLocaleId] })
      queryClient.invalidateQueries({ queryKey: ['archive-ligues-import'] })
      setImportTarget(null)
      setSuccess('Ligue importée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de l'import")
      setImportTarget(null)
    },
  })

  function openCreateForm() {
    setEditingLigueLocale(null)
    setFormNom('')
    setFormEmplacement('')
    setShowForm(true)
  }

  function openEditForm(locale: LigueLocale) {
    setEditingLigueLocale(locale)
    setFormNom(locale.nom)
    setFormEmplacement(locale.emplacement)
    setShowForm(true)
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    const payload: CreateLigueLocalePayload = { nom: formNom, emplacement: formEmplacement }
    if (editingLigueLocale) {
      updateMutation.mutate({ id: editingLigueLocale.id, data: payload })
    } else {
      createMutation.mutate(payload)
    }
  }

  function handleDeleteConfirm() {
    if (deletingLigueLocaleId !== null) {
      deleteMutation.mutate(deletingLigueLocaleId)
    }
  }

  function openLigues(locale: LigueLocale) {
    setSelectedLocaleId(locale.id)
  }

  function backToList() {
    setSelectedLocaleId(null)
  }

  if (localesLoading) {
    return <p>Chargement...</p>
  }

  const selectedLocaleName = selectedLocaleId !== null
    ? locales.find(l => l.id === selectedLocaleId)?.nom ?? ''
    : null

  const associatedLigues = detailData?.ligues ?? []

  return (
    <>
      {error && <div className="error-message" onClick={() => setError(null)}>{error}</div>}
      {success && <div className="success-message" onClick={() => setSuccess(null)}>{success}</div>}

      <div className="admin-header" style={{ marginTop: 0 }}>
        {selectedLocaleId !== null ? (
          <span className="header-right">
            <a href="#back" className="back-arrow" onClick={e => { e.preventDefault(); backToList() }}>&larr;</a>
            {selectedLocaleName}
          </span>
        ) : canWrite ? (
          <button className="btn-primary" onClick={openCreateForm}>
            Nouvelle ligue locale
          </button>
        ) : null}
      </div>

      {selectedLocaleId === null ? (
        <>
          {locales.length === 0 ? (
            <div className="empty-state"><p>Aucune ligue locale</p></div>
          ) : (
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Nom</th>
                  <th>Emplacement</th>
                  {canWrite && <th>Actions</th>}
                </tr>
              </thead>
              <tbody>
                {locales.map(locale => (
                  <tr key={locale.id}>
                    <td><a href="#ligues" className="link" onClick={e => { e.preventDefault(); openLigues(locale) }}>{locale.nom}</a></td>
                    <td>{locale.emplacement}</td>
                    {canWrite && (
                      <td>
                        <button className="btn-secondary btn-small" onClick={() => openEditForm(locale)}>Modifier</button>
                        <button className="btn-danger btn-small" style={{ marginLeft: 8 }} onClick={() => setDeletingLigueLocaleId(locale.id)}>Supprimer</button>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      ) : (
        <>
          {detailLoading ? (
            <p>Chargement...</p>
          ) : detailData ? (
            <>
              {associatedLigues.length > 0 ? (
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>Nom</th>
                      <th>Split</th>
                      <th>Date début</th>
                      <th>Date fin</th>
                    </tr>
                  </thead>
                  <tbody>
                    {associatedLigues.map(l => (
                      <tr key={l.id}>
                        <td>{l.nom}</td>
                        <td>{l.split_name}</td>
                        <td>{l.date_debut}</td>
                        <td>{l.date_fin}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <p className="text-muted">Aucune ligue associée</p>
              )}

              {canWrite && (
                <div className="add-membership-form">
                  <h4>Ajouter une ligue</h4>
                  {liguesForImport && (
                    <LigueSelectField
                      ligues={liguesForImport.items}
                      onImport={(ligueId) => setImportTarget({ localeId: detailData.ligue_locale.id, ligueId })}
                    />
                  )}
                </div>
              )}
            </>
          ) : null}
        </>
      )}

      {showForm && (
        <div className="modal-overlay" onClick={() => { setShowForm(false); setEditingLigueLocale(null) }}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>{editingLigueLocale ? 'Modifier la ligue locale' : 'Nouvelle ligue locale'}</h3>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label htmlFor="ligue-locale-nom">Nom</label>
                <input id="ligue-locale-nom" type="text" value={formNom} onChange={e => setFormNom(e.target.value)} required />
              </div>
              <div className="form-group">
                <label htmlFor="ligue-locale-emplacement">Emplacement</label>
                <input id="ligue-locale-emplacement" type="text" value={formEmplacement} onChange={e => setFormEmplacement(e.target.value)} required />
              </div>
              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => { setShowForm(false); setEditingLigueLocale(null) }}>Annuler</button>
                <button type="submit" className="btn-primary" disabled={createMutation.isPending || updateMutation.isPending}>
                  {createMutation.isPending || updateMutation.isPending
                    ? 'En cours...'
                    : editingLigueLocale ? 'Modifier' : 'Créer'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {importTarget !== null && (
        <div className="modal-overlay" onClick={() => setImportTarget(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>Confirmer l'import</h3>
            <p>Associer cette ligue locale à la ligue {importTarget.ligueId} ?</p>
            <div className="modal-actions">
              <button className="btn-secondary" onClick={() => setImportTarget(null)}>Annuler</button>
              <button className="btn-primary" onClick={() => importLigueMutation.mutate(importTarget)} disabled={importLigueMutation.isPending}>
                {importLigueMutation.isPending ? 'Import...' : 'Importer'}
              </button>
            </div>
          </div>
        </div>
      )}

      {deletingLigueLocaleId !== null && (
        <div className="modal-overlay" onClick={() => setDeletingLigueLocaleId(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>Confirmer la suppression</h3>
            <p>Êtes-vous sûr de vouloir supprimer cette ligue locale ?</p>
            <div className="modal-actions">
              <button className="btn-secondary" onClick={() => setDeletingLigueLocaleId(null)}>Annuler</button>
              <button className="btn-danger" onClick={handleDeleteConfirm} disabled={deleteMutation.isPending}>
                {deleteMutation.isPending ? 'Suppression...' : 'Supprimer'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
