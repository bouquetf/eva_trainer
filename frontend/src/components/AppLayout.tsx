import { NavLink, Outlet } from 'react-router-dom'
import AuthGuard from './AuthGuard'
import { useAuthStore } from '../store/auth'
import '../styles/teams.css'

function AppLayout() {
  const user = useAuthStore((s) => s.user)

  return (
    <AuthGuard>
      <div className="teams-layout">
        <aside className="teams-sidebar">
          <nav>
            <NavLink to="/" end>Accueil</NavLink>
            <NavLink to="/teams" end>Mes équipes</NavLink>
            <NavLink to="/teams/create">Créer une équipe</NavLink>
            {user?.is_admin && <NavLink to="/admin">Administration</NavLink>}
          </nav>
          <nav className="sidebar-secondary">
            <NavLink to="/settings">Paramètres</NavLink>
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
