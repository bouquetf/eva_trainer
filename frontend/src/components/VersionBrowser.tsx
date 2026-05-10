import { useState, useEffect } from 'react'
import type { ZoneVersion } from '../lib/api/types'
import { listZoneVersions } from '../api/versions'

interface VersionBrowserProps {
  teamId: number
  zoneId: number
  refreshKey?: number
  browsingVersionId?: number | null
  onBrowse: (version: ZoneVersion | null) => void
}

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('fr-FR', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso))
}

function VersionBrowser({ teamId, zoneId, refreshKey, browsingVersionId, onBrowse }: VersionBrowserProps) {
  const [versions, setVersions] = useState<ZoneVersion[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!zoneId) return
    setLoading(true)
    listZoneVersions(teamId, zoneId, 1, 50)
      .then(page => setVersions(page.versions))
      .catch(() => setVersions([]))
      .finally(() => setLoading(false))
  }, [teamId, zoneId, refreshKey])

  const handleVersionClick = (version: ZoneVersion) => {
    if (browsingVersionId === version.id) {
      onBrowse(null)
    } else {
      onBrowse(version)
    }
  }

  if (!zoneId) return null

  return (
    <div className="version-browser" style={{ marginTop: 16 }}>
      <h3>Versions de la zone</h3>

      {loading && <div style={{ color: '#888', fontSize: 13 }}>Chargement...</div>}

      {!loading && browsingVersionId != null && (
        <div
          onClick={() => onBrowse(null)}
          style={{
            padding: '6px 8px',
            borderRadius: 4,
            cursor: 'pointer',
            background: '#2e7d32',
            border: '1px solid #00e676',
            fontSize: 12,
            marginBottom: 6,
          }}
        >
          <div style={{ fontWeight: 'bold' }}>
            <span style={{ marginRight: 4 }}>◄</span>
            Version actuelle
          </div>
          <div style={{ color: '#aaa', fontSize: 11 }}>
            Retour à l'état d'édition courant
          </div>
        </div>
      )}

      {!loading && versions.length === 0 && (
        <div style={{ color: '#888', fontSize: 13, fontStyle: 'italic' }}>
          Aucune version sauvegardée
        </div>
      )}

      {versions.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          {versions.map((v) => {
            const isBrowsing = browsingVersionId === v.id

            return (
              <div
                key={v.id}
                onClick={() => handleVersionClick(v)}
                style={{
                  padding: '6px 8px',
                  borderRadius: 4,
                  cursor: isBrowsing ? 'default' : 'pointer',
                  background: isBrowsing ? '#1565c0' : '#333',
                  border: isBrowsing ? '1px solid #42a5f5' : '1px solid transparent',
                  fontSize: 12,
                }}
              >
                <div style={{ fontWeight: isBrowsing ? 'bold' : 'normal' }}>
                  {formatDate(v.created_at)}
                </div>
                <div style={{ color: '#aaa', fontSize: 11 }}>
                  {v.editor_name || 'Utilisateur supprimé'}
                  {' '}({v.version_data.polygons.length} poly.)
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

export default VersionBrowser
