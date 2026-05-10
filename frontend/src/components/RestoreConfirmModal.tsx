import { useEffect, useRef } from 'react'

interface RestoreConfirmModalProps {
  open: boolean
  versionDate: string
  authorName: string
  onConfirm: () => void
  onCancel: () => void
  loading?: boolean
}

function RestoreConfirmModal({ open, versionDate, authorName, onConfirm, onCancel, loading }: RestoreConfirmModalProps) {
  const confirmRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (open) confirmRef.current?.focus()
  }, [open])

  useEffect(() => {
    if (!open) return
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCancel()
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [open, onCancel])

  if (!open) return null

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <h3>Restaurer une version</h3>
        <p>
          Restaurer la version du <strong>{versionDate}</strong> créée par <strong>{authorName}</strong>&nbsp;?
        </p>
        <p style={{ color: '#ff9800', fontSize: 14 }}>
          Les modifications non sauvegardées seront perdues.
        </p>
        <div className="modal-actions">
          <button className="btn-secondary" onClick={onCancel} disabled={loading}>Annuler</button>
          <button ref={confirmRef} className="btn-primary" onClick={onConfirm} disabled={loading}>
            {loading ? 'Restauration...' : 'Confirmer'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default RestoreConfirmModal
