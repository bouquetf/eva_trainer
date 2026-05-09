import { useEffect, useRef } from 'react'

interface ConfirmDialogAltButton {
  label: string
  onClick: () => void
  className?: string
}

interface ConfirmDialogProps {
  open: boolean
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
  onConfirm: () => void
  onCancel: () => void
  altButton?: ConfirmDialogAltButton
}

function ConfirmDialog({ open, title, message, confirmLabel = 'Confirmer', cancelLabel, danger, onConfirm, onCancel, altButton }: ConfirmDialogProps) {
  const confirmRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (open) {
      confirmRef.current?.focus()
    }
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
        <h3>{title}</h3>
        <p>{message}</p>
        <div className="modal-actions">
          {cancelLabel && (
            <button className="btn-secondary" onClick={onCancel}>{cancelLabel}</button>
          )}
          {altButton && (
            <button className={altButton.className ?? 'btn-secondary'} onClick={altButton.onClick}>{altButton.label}</button>
          )}
          <button ref={confirmRef} className={danger ? 'btn-danger' : 'btn-primary'} onClick={onConfirm}>{confirmLabel}</button>
        </div>
      </div>
    </div>
  )
}

export default ConfirmDialog
