import { useEffect, useRef, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listLigues, listDivs, getDiv, createDiv, updateDiv, deleteDiv,
  assignTeamToDiv, removeTeamFromDiv, getDivStandings, listEvaTeams, generateRoundRobin,
  setDivReplayURL, deleteDivReplayURL,
  getDivNote, upsertDivNote,
} from '../api/archives'
import type { Div, DivDetailResponse, StandingEntry } from '../api/archives'
import { detectReplayType, getYouTubeEmbedUrl } from '../lib/replay-url'
import MarkdownEditor from '../components/MarkdownEditor'

export default function DivisionsSection({ canWrite }: { canWrite: boolean }) {
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const [selectedLigueId, setSelectedLigueId] = useState<number | ''>('')
  const [selectedDivId, setSelectedDivId] = useState<number | null>(null)

  const [showDivForm, setShowDivForm] = useState(false)
  const [editingDiv, setEditingDiv] = useState<Div | null>(null)
  const [divFormName, setDivFormName] = useState('')
  const [deletingDivId, setDeletingDivId] = useState<number | null>(null)

  const [showAssignForm, setShowAssignForm] = useState(false)
  const [selectedTeamId, setSelectedTeamId] = useState<number | ''>('')

  const [replayFormVisible, setReplayFormVisible] = useState(false)
  const [replayFormUrl, setReplayFormUrl] = useState('')
  const [deletingReplay, setDeletingReplay] = useState(false)

  const [noteContent, setNoteContent] = useState('')
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const initialContentRef = useRef('')
  const noteLoadedRef = useRef(false)

  const { data: liguesData } = useQuery({
    queryKey: ['archive-ligues', 1],
    queryFn: () => listLigues(1, 100),
    staleTime: 60_000,
  })
  const ligues = liguesData?.items ?? []

  const selectedLigue = ligues.find(l => l.id === selectedLigueId)

  const { data: divsData, isLoading: divsLoading } = useQuery({
    queryKey: ['archive-divs', selectedLigueId],
    queryFn: () => listDivs(Number(selectedLigueId)),
    enabled: selectedLigueId !== '',
    staleTime: 30_000,
  })
  const divs = divsData?.divs ?? []

  const { data: divDetailData } = useQuery({
    queryKey: ['archive-div-detail', selectedDivId],
    queryFn: () => getDiv(selectedDivId!),
    enabled: selectedDivId !== null,
    staleTime: 30_000,
  })
  const divTeams = divDetailData?.teams ?? []

  const { data: standingsData } = useQuery({
    queryKey: ['archive-standings', selectedDivId],
    queryFn: () => getDivStandings(selectedDivId!),
    enabled: selectedDivId !== null,
    staleTime: 30_000,
  })
  const standings: StandingEntry[] = standingsData?.standings ?? []

  const { data: evaTeamsData } = useQuery({
    queryKey: ['archive-eva-teams'],
    queryFn: () => listEvaTeams(),
    staleTime: 60_000,
  })
  const evaTeams = evaTeamsData?.teams ?? []

  const unassignedEvaTeams = evaTeams.filter(
    et => !divTeams.some(dt => dt.id === et.id)
  )

  const createDivMutation = useMutation({
    mutationFn: (data: { ligue_id: number; nom: string }) => createDiv(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
      setShowDivForm(false)
      setEditingDiv(null)
      setSuccess('Division créée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la création")
    },
  })

  const updateDivMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { nom: string } }) => updateDiv(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
      setShowDivForm(false)
      setEditingDiv(null)
      setSuccess('Division modifiée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la modification")
    },
  })

  const deleteDivMutation = useMutation({
    mutationFn: (id: number) => deleteDiv(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
      setDeletingDivId(null)
      setSelectedDivId(null)
      setSuccess('Division supprimée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la suppression")
      setDeletingDivId(null)
    },
  })

  const assignMutation = useMutation({
    mutationFn: ({ divId, teamId }: { divId: number; teamId: number }) => assignTeamToDiv(divId, teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-div-detail'] })
      queryClient.invalidateQueries({ queryKey: ['archive-standings'] })
      queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
      setShowAssignForm(false)
      setSelectedTeamId('')
      setSuccess('Équipe assignée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de l'assignation")
    },
  })

  const removeTeamMutation = useMutation({
    mutationFn: ({ divId, teamId }: { divId: number; teamId: number }) => removeTeamFromDiv(divId, teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['archive-div-detail'] })
      queryClient.invalidateQueries({ queryKey: ['archive-standings'] })
      queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
      setSuccess('Équipe retirée')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors du retrait")
    },
  })

  const generateRoundRobinMutation = useMutation({
    mutationFn: (divId: number) => generateRoundRobin(divId),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['archive-standings'] })
      queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
      setSuccess(`Calendrier généré : ${data.bos.length} rencontre(s)`)
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la génération du calendrier")
    },
  })

  const setReplayURLMutation = useMutation({
    mutationFn: ({ divId, replayUrl }: { divId: number; replayUrl: string }) => setDivReplayURL(divId, replayUrl),
    onSuccess: (data) => {
      queryClient.setQueryData(['archive-div-detail', selectedDivId], (old: DivDetailResponse | undefined) => {
        if (!old) return old
        return { ...old, div: data.div }
      })
      queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
      setReplayFormVisible(false)
      setReplayFormUrl('')
      setSuccess('Lien replay mis à jour')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la mise à jour du lien replay")
    },
  })

  const deleteReplayURLMutation = useMutation({
    mutationFn: (divId: number) => deleteDivReplayURL(divId),
    onSuccess: (data) => {
      queryClient.setQueryData(['archive-div-detail', selectedDivId], (old: DivDetailResponse | undefined) => {
        if (!old) return old
        return { ...old, div: data.div }
      })
      queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
      setSuccess('Lien replay supprimé')
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de la suppression du lien replay")
    },
  })

  const { data: noteData, isLoading: noteLoading } = useQuery({
    queryKey: ['archive-div-note', selectedDivId],
    queryFn: () => getDivNote(selectedDivId!).then(r => r.note.content_md).catch(() => ''),
    enabled: selectedDivId !== null,
    staleTime: 30_000,
  })

  const upsertNoteMutation = useMutation({
    mutationFn: ({ divId, contentMd }: { divId: number; contentMd: string }) => upsertDivNote(divId, contentMd),
    onError: (err: unknown) => {
      const apiErr = err as { body?: { message?: string } }
      setError(apiErr?.body?.message ?? "Erreur lors de l'enregistrement de la note")
    },
  })

  function openCreateDivForm() {
    setEditingDiv(null)
    setDivFormName('')
    setShowDivForm(true)
  }

  function openEditDivForm(div: Div) {
    setEditingDiv(div)
    setDivFormName(div.nom)
    setShowDivForm(true)
  }

  function handleDivSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (editingDiv) {
      updateDivMutation.mutate({ id: editingDiv.id, data: { nom: divFormName } })
    } else {
      createDivMutation.mutate({ ligue_id: Number(selectedLigueId), nom: divFormName })
    }
  }

  function handleDeleteDivConfirm() {
    if (deletingDivId !== null) {
      deleteDivMutation.mutate(deletingDivId)
    }
  }

  function handleAssignSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (selectedDivId !== null && selectedTeamId !== '') {
      assignMutation.mutate({ divId: selectedDivId, teamId: Number(selectedTeamId) })
    }
  }

  function openAddReplayForm() {
    setReplayFormUrl('')
    setReplayFormVisible(true)
  }

  function openEditReplayForm(currentUrl: string) {
    setReplayFormUrl(currentUrl)
    setReplayFormVisible(true)
  }

  function handleReplaySubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (selectedDivId !== null) {
      setReplayURLMutation.mutate({ divId: selectedDivId, replayUrl: replayFormUrl })
    }
  }

  function selectDiv(divId: number) {
    if (selectedDivId !== null && noteLoadedRef.current && noteContent !== initialContentRef.current) {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current)
      upsertNoteMutation.mutate({ divId: selectedDivId, contentMd: noteContent })
    }
    setSelectedDivId(prev => prev === divId ? null : divId)
    setNoteContent('')
    noteLoadedRef.current = false
  }

  useEffect(() => {
    if (noteData !== undefined) {
      initialContentRef.current = noteData
      setNoteContent(noteData)
      noteLoadedRef.current = true
    }
  }, [noteData])

  useEffect(() => {
    if (!noteLoadedRef.current) return
    if (noteContent === initialContentRef.current) return
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current)
    saveTimerRef.current = setTimeout(() => {
      if (selectedDivId !== null) {
        upsertNoteMutation.mutate({ divId: selectedDivId, contentMd: noteContent })
        initialContentRef.current = noteContent
      }
    }, 5000)
    return () => {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current)
    }
  }, [noteContent, selectedDivId])

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
          onChange={e => { setSelectedLigueId(e.target.value === '' ? '' : Number(e.target.value)); setSelectedDivId(null) }}
        >
          <option value="">Sélectionner une période</option>
          {ligues.map(l => (
            <option key={l.id} value={l.id}>
              {l.split_name} ({l.date_debut} → {l.date_fin})
            </option>
          ))}
        </select>
      </div>

      {selectedLigueId !== '' && (
        <>
          <div className="admin-header" style={{ marginTop: 16 }}>
            {canWrite && (
              <button className="btn-primary" onClick={openCreateDivForm}>
                Nouvelle division
              </button>
            )}
          </div>

          {divsLoading ? (
            <p>Chargement...</p>
          ) : divs.length === 0 ? (
            <div className="empty-state"><p>Aucune division</p></div>
          ) : (
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Nom</th>
                  <th>Équipes</th>
                  {canWrite && <th>Actions</th>}
                </tr>
              </thead>
              <tbody>
                {divs.map(div => (
                  <>
                    <tr key={div.id} onClick={() => selectDiv(div.id)} style={{ cursor: 'pointer' }}>
                      <td>{div.nom}</td>
                      <td>{div.teams && div.teams.length > 0 ? div.teams.map(t => `${t.name} (${t.tag})`).join(', ') : (
                        <span className="text-muted">—</span>
                      )}</td>
                      {canWrite && (
                        <td onClick={e => e.stopPropagation()}>
                          <button className="btn-secondary btn-small" onClick={() => openEditDivForm(div)}>Modifier</button>
                          <button className="btn-danger btn-small" style={{ marginLeft: 8 }} onClick={() => setDeletingDivId(div.id)}>Supprimer</button>
                        </td>
                      )}
                    </tr>
                    {selectedDivId === div.id && (
                      <tr key={`${div.id}-detail`} className="row-expanded">
                        <td colSpan={canWrite ? 3 : 2} style={{ padding: 0 }}>
                          <div className="detail-cell">
                            <div className="user-detail">
                              <h4>Équipes dans {div.nom}</h4>
                              {divTeams.length === 0 ? (
                                <p className="text-muted">Aucune équipe assignée</p>
                              ) : (
                                <table className="admin-table memberships-table">
                                  <thead>
                                    <tr>
                                      <th>Nom</th>
                                      <th>Tag</th>
                                      {canWrite && <th>Actions</th>}
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {divTeams.map(team => (
                                      <tr key={team.id}>
                                        <td>{team.name}</td>
                                        <td><code>{team.tag}</code></td>
                                        {canWrite && (
                                          <td>
                                            <button
                                              className="btn-danger btn-small"
                                              onClick={() => removeTeamMutation.mutate({ divId: div.id, teamId: team.id })}
                                              disabled={removeTeamMutation.isPending}
                                            >
                                              Retirer
                                            </button>
                                          </td>
                                        )}
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                              )}

                              {canWrite && (
                                <>
                                  {showAssignForm ? (
                                    <div className="add-membership-form">
                                      <h4>Assigner une équipe</h4>
                                      <form onSubmit={handleAssignSubmit}>
                                        <div className="add-membership-row">
                                          <select
                                            className="role-select"
                                            value={selectedTeamId}
                                            onChange={e => setSelectedTeamId(e.target.value === '' ? '' : Number(e.target.value))}
                                          >
                                            <option value="">Sélectionner une équipe</option>
                                            {unassignedEvaTeams.map(et => (
                                              <option key={et.id} value={et.id}>
                                                {et.name} ({et.tag})
                                              </option>
                                            ))}
                                          </select>
                                          <button type="submit" className="btn-primary btn-small" disabled={selectedTeamId === '' || assignMutation.isPending}>
                                            {assignMutation.isPending ? 'En cours...' : 'Assigner'}
                                          </button>
                                          <button type="button" className="btn-secondary btn-small" onClick={() => { setShowAssignForm(false); setSelectedTeamId('') }}>
                                            Annuler
                                          </button>
                                        </div>
                                      </form>
                                    </div>
                                  ) : (
                                    <button className="btn-primary btn-small" style={{ marginTop: 8 }} onClick={() => setShowAssignForm(true)}>
                                      Assigner une équipe
                                    </button>
                                  )}
                                </>
                              )}

                              {canWrite && (
                                <button
                                  className="btn-secondary btn-small"
                                  style={{ marginTop: 12 }}
                                  onClick={() => generateRoundRobinMutation.mutate(div.id)}
                                  disabled={generateRoundRobinMutation.isPending}
                                >
                                  {generateRoundRobinMutation.isPending ? 'Génération...' : 'Générer le calendrier'}
                                </button>
                              )}

                              <h4 style={{ marginTop: 16 }}>Replay stream</h4>
                              {(() => {
                                const divReplayUrl = divDetailData?.div?.replay_url
                                const replayType = divReplayUrl ? detectReplayType(divReplayUrl) : null
                                const embedUrl = replayType === 'youtube' && divReplayUrl ? getYouTubeEmbedUrl(divReplayUrl) : null

                                if (divReplayUrl) {
                                  return (
                                    <>
                                      {replayType === 'youtube' && embedUrl && (
                                        <div style={{ margin: '8px 0', maxWidth: 560 }}>
                                          <iframe
                                            src={embedUrl}
                                            title="Replay YouTube"
                                            allowFullScreen
                                            style={{ width: '100%', height: 'auto', aspectRatio: '16 / 9', border: 'none', borderRadius: 8 }}
                                          />
                                        </div>
                                      )}
                                      {replayType === 'twitch' && (
                                        <p style={{ margin: '8px 0' }}>
                                          <a href={divReplayUrl} target="_blank" rel="noopener noreferrer">
                                            Regarder sur Twitch
                                          </a>
                                        </p>
                                      )}
                                      {canWrite && (
                                        <div style={{ marginTop: 8 }}>
                                          <button className="btn-secondary btn-small" onClick={() => openEditReplayForm(divReplayUrl)}>Modifier</button>
                                          <button className="btn-danger btn-small" style={{ marginLeft: 8 }} onClick={() => setDeletingReplay(true)}>Supprimer</button>
                                        </div>
                                      )}

                                      {deletingReplay && (
                                        <div className="modal-overlay" onClick={() => setDeletingReplay(false)}>
                                          <div className="modal-content" onClick={e => e.stopPropagation()}>
                                            <h3>Confirmer la suppression</h3>
                                            <p>Êtes-vous sûr de vouloir supprimer le lien replay ?</p>
                                            <div className="modal-actions">
                                              <button className="btn-secondary" onClick={() => setDeletingReplay(false)}>Annuler</button>
                                              <button className="btn-danger" onClick={() => { deleteReplayURLMutation.mutate(div.id); setDeletingReplay(false) }} disabled={deleteReplayURLMutation.isPending}>
                                                Supprimer
                                              </button>
                                            </div>
                                          </div>
                                        </div>
                                      )}
                                    </>
                                  )
                                }

                                return (
                                  <>
                                    <p className="text-muted">Aucun lien replay</p>
                                    {canWrite && (
                                      <button className="btn-primary btn-small" onClick={openAddReplayForm}>Ajouter un lien replay</button>
                                    )}
                                  </>
                                )
                              })()}

                              {replayFormVisible && (
                                <div className="add-membership-form" style={{ marginTop: 8 }}>
                                  <form onSubmit={handleReplaySubmit}>
                                    <div className="add-membership-row">
                                      <input
                                        type="text"
                                        className="role-select"
                                        style={{ flex: 1 }}
                                        value={replayFormUrl}
                                        onChange={e => setReplayFormUrl(e.target.value)}
                                        placeholder="https://www.youtube.com/watch?v=..."
                                      />
                                      <button type="submit" className="btn-primary btn-small" disabled={setReplayURLMutation.isPending}>
                                        {setReplayURLMutation.isPending ? 'En cours...' : 'Valider'}
                                      </button>
                                      <button type="button" className="btn-secondary btn-small" onClick={() => { setReplayFormVisible(false); setReplayFormUrl('') }}>
                                        Annuler
                                      </button>
                                    </div>
                                  </form>
                                </div>
                              )}

                              <h4 style={{ marginTop: 16 }}>Classement {div.nom}</h4>
                              {standings.length === 0 ? (
                                <p className="text-muted">Aucun classement disponible</p>
                              ) : (
                                <table className="admin-table">
                                  <thead>
                                    <tr>
                                      <th>Équipe</th>
                                      <th colSpan={5}>BO</th>
                                      <th colSpan={5}>Matchs</th>
                                    </tr>
                                    <tr>
                                      <th></th>
                                      <th title="BO joués">J</th>
                                      <th title="BO gagnés">G</th>
                                      <th title="BO nuls">N</th>
                                      <th title="BO perdus">P</th>
                                      <th title="Points de BO">Pts</th>
                                      <th title="Matchs joués">J</th>
                                      <th title="Matchs gagnés">G</th>
                                      <th title="Matchs nuls">N</th>
                                      <th title="Matchs perdus">P</th>
                                      <th title="Points de match">Pts</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {standings.map(s => (
                                      <tr key={s.team_id}>
                                        <td><strong>{s.team_name}</strong> <code>{s.team_tag}</code></td>
                                        <td>{s.bos_played}</td>
                                        <td>{s.bo_won}</td>
                                        <td>{s.bo_drawn}</td>
                                        <td>{s.bo_lost}</td>
                                        <td><strong>{s.bo_points}</strong></td>
                                        <td>{s.matches_played}</td>
                                        <td>{s.match_won}</td>
                                        <td>{s.match_drawn}</td>
                                        <td>{s.match_lost}</td>
                                        <td><strong>{s.match_points}</strong></td>
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                              )}

                              <hr style={{ margin: '16px 0', border: 'none', borderTop: '1px solid #ddd' }} />

                              <h4 style={{ marginTop: 16 }}>Notes personnelles</h4>
                              {noteLoading ? (
                                <p className="text-muted">Chargement de la note…</p>
                              ) : (
                                <div className="add-membership-form" style={{ marginTop: 8 }}>
                                  <MarkdownEditor value={noteContent} onChange={setNoteContent} />
                                </div>
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

      {showDivForm && (
        <div className="modal-overlay" onClick={() => { setShowDivForm(false); setEditingDiv(null) }}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>{editingDiv ? 'Modifier la division' : 'Nouvelle division'}</h3>
            {selectedLigue && <p className="text-muted">Ligue : {selectedLigue.split_name}</p>}
            <form onSubmit={handleDivSubmit}>
              <div className="form-group">
                <label htmlFor="div-name">Nom</label>
                <input id="div-name" type="text" value={divFormName} onChange={e => setDivFormName(e.target.value)} required />
              </div>
              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => { setShowDivForm(false); setEditingDiv(null) }}>Annuler</button>
                <button type="submit" className="btn-primary" disabled={createDivMutation.isPending || updateDivMutation.isPending}>
                  {createDivMutation.isPending || updateDivMutation.isPending
                    ? 'En cours...'
                    : editingDiv ? 'Modifier' : 'Créer'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {deletingDivId !== null && (
        <div className="modal-overlay" onClick={() => setDeletingDivId(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h3>Confirmer la suppression</h3>
            <p>Êtes-vous sûr de vouloir supprimer cette division ?</p>
            <div className="modal-actions">
              <button className="btn-secondary" onClick={() => setDeletingDivId(null)}>Annuler</button>
              <button className="btn-danger" onClick={handleDeleteDivConfirm} disabled={deleteDivMutation.isPending}>
                {deleteDivMutation.isPending ? 'Suppression...' : 'Supprimer'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
