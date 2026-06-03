import { NavLink, Outlet } from 'react-router-dom'
import AuthGuard from './AuthGuard'
import { useAuthStore } from '../store/auth'
import { useArchiveAccess } from '../hooks/useArchiveAccess'
import '../styles/teams.css'

function AppLayout() {
  const { user, logout } = useAuthStore()
  const { canRead: canArchiveRead } = useArchiveAccess()

  return (
    <AuthGuard>
      <div className="teams-layout">
        <aside className="teams-sidebar">
          <nav>
            <NavLink to="/teams" end>Mes équipes</NavLink>
            {canArchiveRead && <NavLink to="/archives">Archives</NavLink>}
            {user?.is_admin && <NavLink to="/admin">Administration</NavLink>}
          </nav>
          <nav className="sidebar-secondary">
            <NavLink to="/settings">Paramètres</NavLink>
            <button className="sidebar-logout" onClick={logout}>Déconnexion</button>
          </nav>
        </aside>
        <main className="teams-content">
          <Outlet />
        </main>
      </div>
    </AuthGuard>
  )
}

export default AppLayout
