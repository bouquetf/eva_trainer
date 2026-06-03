import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useArchiveAccess } from '../hooks/useArchiveAccess'

interface ArchiveGuardProps {
  children: ReactNode
}

function ArchiveGuard({ children }: ArchiveGuardProps) {
  const { canRead, isLoading } = useArchiveAccess()

  if (isLoading) {
    return (
      <div className="auth-page">
        <div className="spinner" />
      </div>
    )
  }

  if (!canRead) {
    return <Navigate to="/" replace />
  }

  return <>{children}</>
}

export default ArchiveGuard
