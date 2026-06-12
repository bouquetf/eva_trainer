import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listLigues, listBOs, getBO, createBO, updateBO, deleteBO, listEvaTeams, getLigueMaps,
} from '../api/archives'
import type { BOListItem, BO, BODetail } from '../api/archives'

interface BanFormItem { equipe_id: number | ''; map_id: number | '' }
interface PickFormItem { equipe_id: number | '' | null; map_id: number | '' }
interface MatchFormItem { map_id: number | ''; score_equipe1: number | ''; score_equipe2: number | ''; commentaires: string; forfait: 0 | 1 | 2 }

function emptyBan(): BanFormItem { return { equipe_id: '', map_id: '' } }
function emptyPick(): PickFormItem { return { equipe_id: null, map_id: '' } }
function emptyMatch(): MatchFormItem { return { map_id: '', score_equipe1: '', score_equipe2: '', commentaires: '', forfait: 0 } }

export default function BOsSection({ canWrite }: { canWrite: boolean }) {
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const [selectedLigueId, setSelectedLigueId] = useState<number | ''>('')
  const [selectedBOId, setSelectedBOId] = useState<number | null>(null)

  const [showBOForm, setShowBOForm] = useState(false)
  const [editingBO, setEditingBO] = useState<BO | null>(null)
  const [formEquipe1, setFormEquipe1] = useState<number | ''>('')
  const [formEquipe2, setFormEquipe2] = useState<number | ''>('')
  const [formFormat, setFormFormat] = useState<string>('BO3')
  const [formForfait, setFormForfait] = useState<0 | 1 | 2>(0)
  const [formBans, setFormBans] = useState<BanFormItem[]>([])
  const [formPicks, setFormPicks] = useState<PickFormItem[]>([])
  const [formMatchs, setFormMatchs] = useState<MatchFormItem[]>([])

  const [deletingBOId, setDeletingBOId] = useState<number | null>(null)

  const { data: liguesData } = useQuery({
    queryKey: ['archive-ligues', 1],
    queryFn: () => listLigues(1, 100),
    staleTime: 60_000,
  })
  const ligues = liguesData?.items ?? []
  const selectedLigue = ligues.find(l => l.id === selectedLigueId)

  const { data: bosData, isLoading: bosLoading } = useQuery({
    queryKey: ['archive-bos', selectedLigueId],
    queryFn: () => listBOs(Number(selectedLigueId)),
    enabled: selectedLigueId !== '',
    staleTime: 30_000,
  })
  const bosItems: BOListItem[] = (bosData as { bos: BOListItem[]; total: number })?.bos ?? []

  const { data: boDetailData } = useQuery({
    queryKey: ['archive-bo-detail', selectedBOId],
    queryFn: () => getBO(selectedBOId!),
    enabled: selectedBOId !== null,
    staleTime: 30_000,
  })
  const boDetail: BODetail | null = boDetailData?.bo ?? null

  const { data: evaTeamsData } = useQuery({
    queryKey: ['archive-eva-teams'],
    queryFn: () => listEvaTeams(),
    staleTime: 60_000,
  })
  const evaTeams = evaTeamsData?.teams ?? []
  const equipe1Name = evaTeams.find(t => t.id === formEquipe1)?.name ?? ''
  const equipe2Name = evaTeams.find(t => t.id === formEquipe2)?.name ?? ''

  const { data: ligueMapsData } = useQuery({
    queryKey: ['archive-ligue-maps', selectedLigueId],
    queryFn: () => getLigueMaps(Number(selectedLigueId)),
    enabled: selectedLigueId !== '',
    staleTime: 60_000,
  })
  const ligueMaps = ligueMapsData?.maps ?? []

  const usedMapIds = new Set<number>()
  for (const b of formBans) { if (b.map_id !== '') usedMapIds.add(Number(b.map_id)) }
  for (const p of formPicks) { if (p.map_id !== '') usedMapIds.add(Number(p.map_id)) }

  function availableMaps(excludeId: number | ''): { id: number; name: string }[] {
    const excluded = excludeId === '' ? [] : [Number(excludeId)]
    const taken = new Set(usedMapIds)
    for (const e of excluded) taken.delete(e)
    return ligueMaps.filter(m => !taken.has(m.id))
  }

  function resetForm() {
    setFormEquipe1('')
    setFormEquipe2('')
    setFormFormat('BO3')
    setFormForfait(0)
    setFormBans([])
    setFormPicks([])
    setFormMatchs([])
  }

  function openCreateForm() {
    setEditingBO(null)
    resetForm()
    setShowBOForm(true)
  }

  function openEditForm(bo: BOListItem) {
    setEditingBO(bo)
    setFormEquipe1(bo.equipe1_id)
    setFormEquipe2(bo.equipe2_id)
    setFormFormat(bo.format)
    setFormForfait(bo.forfait as 0 | 1 | 2)
    setFormBans([])
    setFormPicks([])
    setFormMatchs([])
    setShowBOForm(true)

    if (bo.forfait !== 0) return

    getBO(bo.id).then(resp => {
      const d = resp.bo
      setFormBans(d.bans.map(b => ({ equipe_id: b.equipe_id, map_id: b.map_id })))
      setFormPicks(d.picks.map(p => ({ equipe_id: p.equipe_id, map_id: p.map_id })))
      setFormMatchs(d.matchs.map(m => ({ map_id: m.map_id, score_equipe1: m.score_equipe1, score_equipe2: m.score_equipe2, commentaires: m.commentaires ?? '', forfait: m.forfait as 0 | 1 | 2 })))
    }).catch(() => {})
  }

  function addPick() {
    setFormPicks(prev => [...prev, emptyPick()])
    setFormMatchs(prev => [...prev, emptyMatch()])
  }

  function removePick(idx: number) {
    setFormPicks(prev => prev.filter((_, i) => i !== idx))
    setFormMatchs(prev => prev.filter((_, i) => i !== idx))
  }

  function updatePickMap(idx: number, mapId: number | '') {
    setFormPicks(prev => {
      const np = [...prev]
      np[idx] = { ...np[idx], map_id: mapId }
      return np
    })
    setFormMatchs(prev => {
      const nm = [...prev]
      nm[idx] = { ...nm[idx], map_id: mapId }
      return nm
    })
  }

  const createBOMutation = useMutation({
    mutationFn: () => createBO({
      ligue_id: Number(selectedLigueId),
      equipe1_id: Number(formEquipe1),
      equipe2_id: Number(formEquipe2),
      format: formFormat,
      forfait: formForfait || undefined,
      bans: formForfait === 0 && formBans.length > 0 ? formBans.map((b, idx) => ({ equipe_id: Number(b.equipe_id), map_id: Number(b.map_id), ordre: idx + 1 })) : undefined,
      picks: formForfait === 0 && formPicks.length > 0 ? formPicks.map((p, idx) => ({ equipe_id: p.equipe_id === '' ? null : Number(p.equipe_id), map_id: Number(p.map_id), ordre: idx + 1 })) : undefined,
      matchs: formForfait === 0 && formMatchs.length > 0 ? formMatchs.map((m, i) => {
        const mapId = formPicks[i]?.map_id !== '' ? formPicks[i]?.map_id : (m.map_id !== '' ? m.map_id : '')
        return { map_id: Number(mapId || 0), score_equipe1: Number(m.score_equipe1), score_equipe2: Number(m.score_equipe2), forfait: m.forfait || undefined, commentaires: m.commentaires || null }
      }) : undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-bos'] })
      setShowBOForm(false)
      setEditingBO(null)
      resetForm()
      setSuccess('BO créé')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la création")
    },
  })

  const updateBOMutation = useMutation({
    mutationFn: () => {
      if (!editingBO) throw new Error('no bo')
      return updateBO(editingBO.id, {
        format: formFormat,
        forfait: formForfait || undefined,
        bans: formForfait === 0 && formBans.length > 0 ? formBans.map((b, idx) => ({ equipe_id: Number(b.equipe_id), map_id: Number(b.map_id), ordre: idx + 1 })) : undefined,
        picks: formForfait === 0 && formPicks.length > 0 ? formPicks.map((p, idx) => ({ equipe_id: p.equipe_id === '' ? null : Number(p.equipe_id), map_id: Number(p.map_id), ordre: idx + 1 })) : undefined,
        matchs: formForfait === 0 && formMatchs.length > 0 ? formMatchs.map((m, i) => {
          const mapId = formPicks[i]?.map_id !== '' ? formPicks[i]?.map_id : (m.map_id !== '' ? m.map_id : '')
          return { map_id: Number(mapId || 0), score_equipe1: Number(m.score_equipe1), score_equipe2: Number(m.score_equipe2), forfait: m.forfait || undefined, commentaires: m.commentaires || null }
        }) : undefined,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-bos'] })
      queryClient.invalidateQueries({ queryKey: ['archive-bo-detail'] })
      setShowBOForm(false)
      setEditingBO(null)
      resetForm()
      setSuccess('BO modifié')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la modification")
    },
  })

  const deleteBOMutation = useMutation({
    mutationFn: (id: number) => deleteBO(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-bos'] })
      setDeletingBOId(null)
      setSelectedBOId(null)
      setSuccess('BO supprimé')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la suppression")
      setDeletingBOId(null)
    },
  })

  function handleFormSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (editingBO) {
      updateBOMutation.mutate()
    } else {
      createBOMutation.mutate()
    }
  }

  function handleDeleteConfirm() {
    if (deletingBOId !== null) {
      deleteBOMutation.mutate(deletingBOId)
    }
  }

  function selectBO(id: number) {
    setSelectedBOId(prev => prev === id ? null : id)
  }

  function formatDate(s: string): string {
    return s ? new Date(s).toLocaleDateString('fr-FR') : ''
  }

  return (
    <>
      {error && <div className="error-message" onClick={() => setError(null)}>{error}</div>}
      {success && <div className="success-message" onClick={() => setSuccess(null)}>{success}</div>}

      <div className="form-group">
        <label htmlFor="ligue-selector">Période</label>
        <select
          id="ligue-selector"
          className="role-select"
          value={selectedLigueId}
          onChange={e => { setSelectedLigueId(e.target.value === '' ? '' : Number(e.target.value)); setSelectedBOId(null) }}
        >
          <option value="">Sélectionner une période</option>
          {ligues.map(l => (
            <option key={l.id} value={l.id}>
              {l.nom} ({l.split_name}, {l.date_debut} → {l.date_fin})
            </option>
          ))}
        </select>
      </div>

      {selectedLigueId !== '' && (
        <>
          <div className="admin-header" style={{ marginTop: 16 }}>
            {canWrite && (
              <button className="btn-primary" onClick={openCreateForm}>
                Nouveau BO
              </button>
            )}
          </div>

          {bosLoading ? (
            <p>Chargement...</p>
          ) : bosItems.length === 0 ? (
            <div className="empty-state"><p>Aucun BO</p></div>
          ) : (
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Équipe 1</th>
                  <th>Équipe 2</th>
                  <th>Format</th>
                  <th>Vainqueur</th>
                  <th>Date</th>
                  {canWrite && <th>Actions</th>}
                </tr>
              </thead>
              <tbody>
                {bosItems.map(bo => (
                  <>
                    <tr key={bo.id} onClick={() => selectBO(bo.id)} style={{ cursor: 'pointer' }}>
                      <td><strong>{bo.equipe1_name}</strong></td>
                      <td><strong>{bo.equipe2_name}</strong></td>
                      <td>{bo.format}{bo.forfait !== 0 && ' (Forfait)'}</td>
                      <td>
                        {bo.forfait !== 0 ? (
                          <span style={{ color: '#e74c3c' }}>
                            Forfait {bo.forfait === 1 ? bo.equipe1_name : bo.equipe2_name}
                          </span>
                        ) : bo.vainqueur_id != null ? (
                          bo.vainqueur_id === bo.equipe1_id ? bo.equipe1_name : bo.equipe2_name
                        ) : (
                          <span className="text-muted">—</span>
                        )}
                      </td>
                      <td>{formatDate(bo.created_at)}</td>
                      {canWrite && (
                        <td onClick={e => e.stopPropagation()}>
                          <button className="btn-secondary btn-small" onClick={() => openEditForm(bo)}>Modifier</button>
                          <button className="btn-danger btn-small" style={{ marginLeft: 8 }} onClick={() => setDeletingBOId(bo.id)}>Supprimer</button>
                        </td>
                      )}
                    </tr>
                    {selectedBOId === bo.id && boDetail && (
                      <tr key={`${bo.id}-detail`} className="row-expanded">
                        <td colSpan={canWrite ? 6 : 5} style={{ padding: 0 }}>
                          <div className="detail-cell">
                            <div className="user-detail">
                              <h4>Détails du BO</h4>
                              <p>
                                {boDetail.BO.forfait !== 0 ? (
                                  <span style={{ color: '#e74c3c' }}>
                                    Forfait {boDetail.BO.forfait === 1 ? bo.equipe1_name : bo.equipe2_name}
                                  </span>
                                ) : (
                                  <>
                                    {boDetail.BO.equipe1_id === boDetail.BO.vainqueur_id ? '🏆' : ''} {bo.equipe1_name}
                                    {' vs '}
                                    {bo.equipe2_name} {boDetail.BO.vainqueur_id === boDetail.BO.equipe2_id ? '🏆' : ''}
                                    {' — '}{boDetail.BO.format}
                                  </>
                                )}
                              </p>

                              {boDetail.bans.length > 0 && (
                                <>
                                  <h5 style={{ marginTop: 12 }}>Bans</h5>
                                  <table className="admin-table">
                                    <thead><tr><th>#</th><th>Équipe</th><th>Map</th></tr></thead>
                                    <tbody>
                                      {boDetail.bans.sort((a, b) => a.id - b.id).map(b => (
                                        <tr key={b.id}>
                                          <td>{boDetail.bans.indexOf(b) + 1}</td>
                                          <td>{b.equipe_id === boDetail.BO.equipe1_id ? bo.equipe1_name : bo.equipe2_name}</td>
                                          <td>{ligueMaps.find(m => m.id === b.map_id)?.name ?? `#${b.map_id}`}</td>
                                        </tr>
                                      ))}
                                    </tbody>
                                  </table>
                                </>
                              )}

                              {boDetail.picks.length > 0 && (
                                <>
                                  <h5 style={{ marginTop: 12 }}>Picks</h5>
                                  <table className="admin-table">
                                    <thead><tr><th>#</th><th>Équipe</th><th>Map</th></tr></thead>
                                    <tbody>
                                      {boDetail.picks.sort((a, b) => a.id - b.id).map(p => (
                                        <tr key={p.id}>
                                          <td>{boDetail.picks.indexOf(p) + 1}</td>
                                          <td>{p.equipe_id === 0 ? '—' : p.equipe_id === boDetail.BO.equipe1_id ? bo.equipe1_name : bo.equipe2_name}</td>
                                          <td>{ligueMaps.find(m => m.id === p.map_id)?.name ?? `#${p.map_id}`}</td>
                                        </tr>
                                      ))}
                                    </tbody>
                                  </table>
                                </>
                              )}

                              {boDetail.matchs.length > 0 && (
                                <>
                                  <h5 style={{ marginTop: 12 }}>Matchs</h5>
                                  <table className="admin-table">
                                    <thead><tr><th>Map</th><th>Score E1</th><th>Score E2</th><th>Statut</th><th>Vainqueur</th><th>Commentaires</th></tr></thead>
                                    <tbody>
                                      {boDetail.matchs.map(m => (
                                        <tr key={m.id}>
                                          <td>{ligueMaps.find(mm => mm.id === m.map_id)?.name ?? `#${m.map_id}`}</td>
                                          <td>{m.forfait ? '—' : m.score_equipe1}</td>
                                          <td>{m.forfait ? '—' : m.score_equipe2}</td>
                                          <td>{m.forfait ? <span style={{ color: '#e74c3c' }}>Forfait</span> : 'Joué'}</td>
                                          <td>
                                            {m.vainqueur_id != null
                                              ? (m.vainqueur_id === boDetail.BO.equipe1_id ? bo.equipe1_name : bo.equipe2_name)
                                              : <span className="text-muted">—</span>
                                            }
                                          </td>
                                          <td>{m.commentaires ?? ''}</td>
                                        </tr>
                                      ))}
                                    </tbody>
                                  </table>
                                </>
                              )}
                            </div>
                          </div>
                        </td>
                      </tr>
                    )}
                  </>
                ))}
              </tbody>
            </table>
          )}
        </>
      )}

      {showBOForm && (
        <div className="modal-overlay" onClick={() => { setShowBOForm(false); setEditingBO(null) }}>
          <div className="modal-content modal-wide" onClick={e => e.stopPropagation()}>
            <h3>{editingBO ? 'Modifier le BO' : 'Nouveau BO'}</h3>
            {selectedLigue && <p className="text-muted">Ligue : {selectedLigue.nom} ({selectedLigue.split_name})</p>}
            <form onSubmit={handleFormSubmit}>
              <div className="form-group">
                <label htmlFor="bo-equipe1">Équipe 1</label>
                <select id="bo-equipe1" className="role-select" value={formEquipe1} onChange={e => setFormEquipe1(e.target.value === '' ? '' : Number(e.target.value))} required disabled={!!editingBO}>
                  <option value="">Sélectionner</option>
                  {evaTeams.map(t => (
                    <option key={t.id} value={t.id} disabled={t.id === formEquipe2}>{t.name} ({t.tag})</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label htmlFor="bo-equipe2">Équipe 2</label>
                <select id="bo-equipe2" className="role-select" value={formEquipe2} onChange={e => setFormEquipe2(e.target.value === '' ? '' : Number(e.target.value))} required disabled={!!editingBO}>
                  <option value="">Sélectionner</option>
                  {evaTeams.map(t => (
                    <option key={t.id} value={t.id} disabled={t.id === formEquipe1}>{t.name} ({t.tag})</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label htmlFor="bo-format">Format</label>
                <select id="bo-format" className="role-select" value={formFormat} onChange={e => setFormFormat(e.target.value)} required>
                  <option value="BO3">BO3</option>
                  <option value="BO5">BO5</option>
                </select>
              </div>

              <div className="form-group" style={{ marginTop: 16 }}>
                <label>Statut</label>
                <div className="add-membership-row" style={{ gap: 8 }}>
                  <button type="button" className={`btn-small ${formForfait === 0 ? 'btn-primary' : 'btn-secondary'}`} onClick={() => setFormForfait(0)}>
                    Match normal
                  </button>
                  <button type="button" className={`btn-small ${formForfait === 1 ? 'btn-danger' : 'btn-secondary'}`} onClick={() => setFormForfait(1)}>
                    Forfait {equipe1Name || 'E1'}
                  </button>
                  <button type="button" className={`btn-small ${formForfait === 2 ? 'btn-danger' : 'btn-secondary'}`} onClick={() => setFormForfait(2)}>
                    Forfait {equipe2Name || 'E2'}
                  </button>
                </div>
              </div>

              {formForfait === 0 && (
                <>
                  <div className="form-group" style={{ marginTop: 16 }}>
                    <label>Bans</label>
                    {formBans.map((ban, i) => (
                      <div key={i} className="add-membership-row" style={{ marginTop: 4 }}>
                        <select className="role-select" value={ban.equipe_id} onChange={e => { const nb = [...formBans]; nb[i] = { ...nb[i], equipe_id: e.target.value === '' ? '' : Number(e.target.value) }; setFormBans(nb) }} required>
                          <option value="">Équipe</option>
                          {evaTeams.filter(t => t.id === formEquipe1 || t.id === formEquipe2).map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                        </select>
                        <select className="role-select" value={ban.map_id} onChange={e => { const nb = [...formBans]; nb[i] = { ...nb[i], map_id: e.target.value === '' ? '' : Number(e.target.value) }; setFormBans(nb) }} required>
                          <option value="">Map</option>
                          {availableMaps(ban.map_id).map(m => <option key={m.id} value={m.id}>{m.name}</option>)}
                        </select>
                        <button type="button" className="btn-danger btn-small" onClick={() => setFormBans(formBans.filter((_, j) => j !== i))}>×</button>
                      </div>
                    ))}
                    <button type="button" className="btn-secondary btn-small" style={{ marginTop: 4 }} onClick={() => setFormBans([...formBans, emptyBan()])}>+ Ajouter un ban</button>
                  </div>

                  <div className="form-group" style={{ marginTop: 16 }}>
                    <label>Picks</label>
                    {formPicks.map((pick, i) => (
                      <div key={i} className="add-membership-row" style={{ marginTop: 4 }}>
                        <select className="role-select" value={pick.equipe_id ?? ''} onChange={e => { const np = [...formPicks]; np[i] = { ...np[i], equipe_id: e.target.value === '' ? null : Number(e.target.value) }; setFormPicks(np) }}>
                          <option value="">—</option>
                          {evaTeams.filter(t => t.id === formEquipe1 || t.id === formEquipe2).map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                        </select>
                        <select className="role-select" value={pick.map_id} onChange={e => updatePickMap(i, e.target.value === '' ? '' : Number(e.target.value))} required>
                          <option value="">Map</option>
                          {availableMaps(pick.map_id).map(m => <option key={m.id} value={m.id}>{m.name}</option>)}
                        </select>
                        <button type="button" className="btn-danger btn-small" onClick={() => removePick(i)}>×</button>
                      </div>
                    ))}
                    <button type="button" className="btn-secondary btn-small" style={{ marginTop: 4 }} onClick={addPick}>+ Ajouter un pick</button>
                  </div>

                  <div className="form-group" style={{ marginTop: 16 }}>
                    <label>Matchs</label>
                    {formMatchs.map((match, i) => (
                      <div key={i} style={{ marginTop: 8, padding: 8, border: '1px solid var(--border-color)', borderRadius: 6 }}>
                        <div className="add-membership-row" style={{ marginBottom: 6 }}>
                          <span style={{ fontSize: 14, fontWeight: 600 }}>
                            Map : {ligueMaps.find(m => m.id === (formPicks[i]?.map_id !== '' && formPicks[i]?.map_id !== undefined ? Number(formPicks[i].map_id) : 0))?.name ?? `#${match.map_id || '?'}`}
                          </span>
                          <button type="button" className="btn-danger btn-small" onClick={() => removePick(i)}>×</button>
                        </div>
                        <div className="add-membership-row" style={{ gap: 8 }}>
                          {match.forfait === 0 ? (
                            <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                              <span style={{ fontSize: 13, fontWeight: 600, minWidth: 60, textAlign: 'right' }}>{equipe1Name || 'E1'}</span>
                              <input type="number" className="role-select" style={{ width: 70 }} min={0} max={100} value={match.score_equipe1} onChange={e => { const nm = [...formMatchs]; nm[i] = { ...nm[i], score_equipe1: e.target.value === '' ? '' : Number(e.target.value) }; setFormMatchs(nm) }} required />
                              <span style={{ fontSize: 13 }}>—</span>
                              <input type="number" className="role-select" style={{ width: 70 }} min={0} max={100} value={match.score_equipe2} onChange={e => { const nm = [...formMatchs]; nm[i] = { ...nm[i], score_equipe2: e.target.value === '' ? '' : Number(e.target.value) }; setFormMatchs(nm) }} required />
                              <span style={{ fontSize: 13, fontWeight: 600, minWidth: 60 }}>{equipe2Name || 'E2'}</span>
                            </div>
                          ) : (
                            <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                              <span style={{ fontSize: 13, fontWeight: 600, color: '#e74c3c' }}>
                                Forfait {match.forfait === 1 ? equipe1Name || 'E1' : equipe2Name || 'E2'}
                              </span>
                            </div>
                          )}
                          <button type="button" className={`btn-small ${match.forfait === 1 ? 'btn-danger' : 'btn-secondary'}`}
                            onClick={() => { const nm = [...formMatchs]; nm[i] = { ...nm[i], commentaires: nm[i].forfait ? nm[i].commentaires : (nm[i].commentaires || 'Forfait'), forfait: nm[i].forfait === 1 ? 0 : 1 }; setFormMatchs(nm) }}
                            title="Forfait équipe 1">
                            Forfait {equipe1Name || 'E1'}
                          </button>
                          <button type="button" className={`btn-small ${match.forfait === 2 ? 'btn-danger' : 'btn-secondary'}`}
                            onClick={() => { const nm = [...formMatchs]; nm[i] = { ...nm[i], commentaires: nm[i].forfait ? nm[i].commentaires : (nm[i].commentaires || 'Forfait'), forfait: nm[i].forfait === 2 ? 0 : 2 }; setFormMatchs(nm) }}
                            title="Forfait équipe 2">
                            Forfait {equipe2Name || 'E2'}
                          </button>
                        </div>
                        <div className="add-membership-row" style={{ marginTop: 4 }}>
                          <input type="text" className="role-select" value={match.commentaires} onChange={e => { const nm = [...formMatchs]; nm[i] = { ...nm[i], commentaires: e.target.value }; setFormMatchs(nm) }} placeholder="Commentaire" style={{ flex: 1 }} />
                        </div>
                      </div>
                    ))}
                  </div>
                </>
              )}

              <div className="modal-actions" style={{ marginTop: 16 }}>
                <button type="button" className="btn-secondary" onClick={() => { setShowBOForm(false); setEditingBO(null) }}>Annuler</button>
                <button type="submit" className="btn-primary" disabled={createBOMutation.isPending || updateBOMutation.isPending}>
                  {createBOMutation.isPending || updateBOMutation.isPending
                    ? 'En cours...'
                    : editingBO ? 'Modifier' : 'Créer'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {deletingBOId !== null && (
        <div className="modal-overlay" onClick={() => setDeletingBOId(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>Confirmer la suppression</h3>
            <p>Êtes-vous sûr de vouloir supprimer ce BO ?</p>
            <div className="modal-actions">
              <button className="btn-secondary" onClick={() => setDeletingBOId(null)}>Annuler</button>
              <button className="btn-danger" onClick={handleDeleteConfirm} disabled={deleteBOMutation.isPending}>
                {deleteBOMutation.isPending ? 'Suppression...' : 'Supprimer'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
