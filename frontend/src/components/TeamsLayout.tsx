import { NavLink, Outlet } from 'react-router-dom'
import AuthGuard from './AuthGuard'
import '../styles/teams.css'

function TeamsLayout() {
  return (
    <AuthGuard>
      <div className="teams-layout">
        <aside className="teams-sidebar">
          <nav>
            <NavLink to="/teams" end>Mes équipes</NavLink>
            <NavLink to="/teams/create">Créer une équipe</NavLink>
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

export default TeamsLayout
