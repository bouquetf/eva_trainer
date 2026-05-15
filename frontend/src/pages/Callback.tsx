import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import '../styles/auth.css'

function Callback() {
  const navigate = useNavigate()
  const { handleCallback, isLoading, error, logout } = useAuthStore()

  useEffect(() => {
    handleCallback()
  }, [handleCallback])

  useEffect(() => {
    if (!isLoading && !error) {
      const redirect = new URLSearchParams(window.location.hash.replace(/^#/, '')).get('redirect') ?? '/teams'
      navigate(redirect, { replace: true })
    }
  }, [isLoading, error, navigate])

  if (error) {
    return (
      <div className="auth-page">
        <div className="auth-card">
          <h1>Erreur de connexion</h1>
          <p className="auth-error">{error}</p>
          <a href="/login" className="auth-link">Retour à la connexion</a>
          <button className="btn-logout" onClick={logout}>Se déconnecter</button>
        </div>
      </div>
    )
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Eva Trainer</h1>
        <p>Connexion en cours...</p>
        <div className="spinner" />
      </div>
    </div>
  )
}

export default Callback
