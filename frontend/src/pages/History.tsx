import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { listTeamMaps } from '../api/maps'
import { listMapVersions, restoreVersion } from '../api/versions'
import { listMembers } from '../api/roles'
import { listZones } from '../api/zones'
import { useTeamRole } from '../hooks/useTeamRole'
import { ApiError } from '../api/client'
import RestoreConfirmModal from '../components/RestoreConfirmModal'
import type { ZoneVersion, VersionPage } from '../lib/api/types'
import type { TeamMemberWithUser } from '../api/roles'

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('fr-FR', { dateStyle: 'long', timeStyle: 'short' }).format(new Date(iso))
}

function History() {
  const { id, mapId } = useParams<{ id: string; mapId: string }>()
  const teamId = Number(id)
  const parsedMapId = Number(mapId)

  const [mapName, setMapName] = useState('')
  const [zones, setZones] = useState<{ id: number; name: string }[]>([])
  const [members, setMembers] = useState<TeamMemberWithUser[]>([])

  const { role: currentUserRole, canEdit } = useTeamRole(teamId)
  const isCaptainOrCocaptain = currentUserRole === 'captain' || currentUserRole === 'cocaptain'

  const [versions, setVersions] = useState<ZoneVersion[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [limit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [filterZone, setFilterZone] = useState<number | undefined>()
  const [filterEditor, setFilterEditor] = useState<number | undefined>()

  const [restoreTarget, setRestoreTarget] = useState<ZoneVersion | null>(null)
  const [restoring, setRestoring] = useState(false)
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null)

  useEffect(() => {
    if (isNaN(teamId) || isNaN(parsedMapId)) return

    const init = async () => {
      try {
        const mapsData = await listTeamMaps(teamId)
        const map = mapsData.maps.find(m => m.id === parsedMapId)
        if (map) setMapName(map.name)
      } catch { /* ignore */ }

      try {
        const zonesData = await listZones(teamId, parsedMapId)
        setZones(zonesData.zones.map(z => ({ id: z.id, name: z.name || `Zone #${z.id}` })))
      } catch { /* ignore */ }

      if (isCaptainOrCocaptain) {
        try {
          const membersData = await listMembers(teamId)
          setMembers(membersData.members)
        } catch { /* ignore */ }
      }
    }
    init()
  }, [teamId, parsedMapId, isCaptainOrCocaptain])

  useEffect(() => {
    loadVersions()
  }, [teamId, parsedMapId, page, filterZone, filterEditor])

  const loadVersions = async () => {
    setLoading(true)
    setError(null)
    try {
      const data: VersionPage = await listMapVersions(teamId, parsedMapId, {
        page,
        limit,
        zone_id: filterZone,
        editor_id: filterEditor,
      })
      setVersions(data.versions)
      setTotal(data.total)
    } catch (err) {
      setError('Erreur lors du chargement des versions.')
      setVersions([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  const handleRestore = async () => {
    if (!restoreTarget) return
    setRestoring(true)
    try {
      await restoreVersion(teamId, restoreTarget.zone_id, restoreTarget.id)
      setNotification({ type: 'success', message: `Version du ${formatDate(restoreTarget.created_at)} restaurée avec succès.` })
      setRestoreTarget(null)
      loadVersions()
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 403) {
          setNotification({ type: 'error', message: 'Vous n\'êtes pas autorisé à restaurer cette version.' })
        } else if (err.status === 423) {
          setNotification({ type: 'error', message: 'Vous devez avoir un verrou d\'édition actif.' })
        } else {
          setNotification({ type: 'error', message: err.body?.message as string || 'Erreur lors de la restauration.' })
        }
      } else {
        setNotification({ type: 'error', message: 'Erreur lors de la restauration.' })
      }
      setRestoreTarget(null)
    } finally {
      setRestoring(false)
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / limit))

  if (isNaN(teamId) || isNaN(parsedMapId)) {
    return <div style={{ padding: 50, textAlign: 'center', color: '#f44336' }}>Paramètres d'URL invalides.</div>
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
        <h1 style={{ margin: 0 }}>Historique des versions - {mapName || `Carte #${parsedMapId}`}</h1>
        {canEdit && (
          <Link to={`/teams/${teamId}/editor/${parsedMapId}`} className="btn-secondary" style={{ textDecoration: 'none' }}>
            Retour à l'éditeur
          </Link>
        )}
      </div>

      {notification && (
        <div className={notification.type === 'success' ? 'success-message' : 'error-message'} style={{ marginBottom: 16 }}>
          {notification.message}
          <button
            onClick={() => setNotification(null)}
            style={{ marginLeft: 12, background: 'none', border: 'none', cursor: 'pointer', color: 'inherit' }}
          >
            ×
          </button>
        </div>
      )}

      <div style={{ display: 'flex', gap: 12, marginBottom: 16, alignItems: 'flex-end' }}>
        <div>
          <label style={{ display: 'block', fontSize: 12, marginBottom: 4, color: '#aaa' }}>Zone</label>
          <select
            value={filterZone ?? ''}
            onChange={e => { setFilterZone(e.target.value ? Number(e.target.value) : undefined); setPage(1) }}
            style={{ padding: '6px 8px' }}
          >
            <option value="">Toutes les zones</option>
            {zones.map(z => (
              <option key={z.id} value={z.id}>{z.name}</option>
            ))}
          </select>
        </div>
        {isCaptainOrCocaptain && (
          <div>
            <label style={{ display: 'block', fontSize: 12, marginBottom: 4, color: '#aaa' }}>Éditeur</label>
            <select
              value={filterEditor ?? ''}
              onChange={e => { setFilterEditor(e.target.value ? Number(e.target.value) : undefined); setPage(1) }}
              style={{ padding: '6px 8px' }}
            >
              <option value="">Tous les éditeurs</option>
              {members.map(m => (
                <option key={m.user_id} value={m.user_id}>{m.display_name}</option>
              ))}
            </select>
          </div>
        )}
      </div>

      {loading && <div className="spinner" />}

      {error && (
        <div className="error-message">
          {error}
          <button className="btn-secondary btn-small" onClick={loadVersions} style={{ marginLeft: 12 }}>Réessayer</button>
        </div>
      )}

      {!loading && !error && versions.length === 0 && (
        <p style={{ color: '#888', fontStyle: 'italic' }}>Aucune version sauvegardée sur cette carte.</p>
      )}

      {!loading && !error && versions.length > 0 && (
        <>
          <div style={{ overflowX: 'auto' }}>
            <table className="invitations-table" style={{ width: '100%' }}>
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Zone</th>
                  <th>Auteur</th>
                  <th>Polygones</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {versions.map(v => (
                  <tr key={v.id}>
                    <td>{formatDate(v.created_at)}</td>
                    <td>
                      {canEdit ? (
                        <Link to={`/teams/${teamId}/editor/${parsedMapId}`} style={{ color: '#90caf9' }}>
                          {v.zone_name || `Zone #${v.zone_id}`}
                        </Link>
                      ) : (
                        <span>{v.zone_name || `Zone #${v.zone_id}`}</span>
                      )}
                    </td>
                    <td>{v.editor_name || 'Utilisateur supprimé'}</td>
                    <td>{v.version_data.polygons.length}</td>
                    <td>
                      {canEdit && (
                        <button
                          className="btn-primary btn-small"
                          onClick={() => setRestoreTarget(v)}
                        >
                          Restaurer
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 16, alignItems: 'center' }}>
            <button
              className="btn-secondary btn-small"
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page <= 1}
            >
              Précédent
            </button>
            <span style={{ fontSize: 13, color: '#aaa' }}>
              Page {page} / {totalPages} ({total} résultats)
            </span>
            <button
              className="btn-secondary btn-small"
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
            >
              Suivant
            </button>
          </div>
        </>
      )}

      <RestoreConfirmModal
        open={restoreTarget !== null}
        versionDate={restoreTarget ? formatDate(restoreTarget.created_at) : ''}
        authorName={restoreTarget?.editor_name || 'Utilisateur supprimé'}
        onConfirm={handleRestore}
        onCancel={() => setRestoreTarget(null)}
        loading={restoring}
      />
    </div>
  )
}

export default History
