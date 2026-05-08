import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'

interface AdminGuardProps {
  children: ReactNode
}

function AdminGuard({ children }: AdminGuardProps) {
  const user = useAuthStore((s) => s.user)
  const isLoading = useAuthStore((s) => s.isLoading)

  if (isLoading) {
    return (
      <div className="auth-page">
        <div className="spinner" />
      </div>
    )
  }

  if (!user?.is_admin) {
    return <Navigate to="/" replace />
  }

  return <>{children}</>
}

export default AdminGuard
