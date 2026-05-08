import { Link } from 'react-router-dom'
import { useAuthStore } from '../store/auth'

function Home() {
  const { isAuthenticated, user, login, logout, isLoading } = useAuthStore()

  return (
    <div>
      <h1>Eva Trainer</h1>

      {isLoading ? (
        <div className="spinner" />
      ) : isAuthenticated ? (
        <div className="auth-user-info">
          <span className="auth-user-id">{user?.discord_id ?? 'Connecté'}</span>
          <Link to="/editor">Éditeur de zones</Link>
          <Link to="/teams">Équipes</Link>
          <button className="btn-logout" onClick={logout}>Déconnexion</button>
        </div>
      ) : (
        <nav>
          <button className="btn-discord" onClick={login}>
            Connexion avec Discord
          </button>
        </nav>
      )}
    </div>
  )
}

export default Home
