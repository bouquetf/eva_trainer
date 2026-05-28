import { useState, useEffect } from 'react'
import type { LineOfFireVersion } from '../lib/api/types'
import { listLineOfFireVersions } from '../api/versions'

interface Props {
  teamId: number
  lineOfFireId: number
  refreshKey?: number
  browsingVersionId?: number | null
  onBrowse: (version: LineOfFireVersion | null) => void
}

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('fr-FR', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso))
}

const containerStyle: React.CSSProperties = { marginTop: 16 }
const loadingStyle: React.CSSProperties = { color: '#888', fontSize: 13 }
const emptyStyle: React.CSSProperties = { color: '#888', fontSize: 13, fontStyle: 'italic' }

function LineOfFireVersionBrowser({ teamId, lineOfFireId, refreshKey, browsingVersionId, onBrowse }: Props) {
  const [versions, setVersions] = useState<LineOfFireVersion[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!lineOfFireId) return
    setLoading(true)
    listLineOfFireVersions(teamId, lineOfFireId)
      .then(res => setVersions(res.versions))
      .catch(() => setVersions([]))
      .finally(() => setLoading(false))
  }, [teamId, lineOfFireId, refreshKey])

  const handleVersionClick = (version: LineOfFireVersion) => {
    if (browsingVersionId === version.id) {
      onBrowse(null)
    } else {
      onBrowse(version)
    }
  }

  if (!lineOfFireId) return null

  return (
    <div style={containerStyle}>
      <h3 style={{ fontSize: 14, margin: '0 0 8px' }}>Versions de la ligne de tir</h3>

      {loading && <div style={loadingStyle}>Chargement...</div>}

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
        <div style={emptyStyle}>Aucune version sauvegardée</div>
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
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

export default LineOfFireVersionBrowser
