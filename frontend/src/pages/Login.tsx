import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import '../styles/auth.css'

function DiscordLogo() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 127.14 96.36" width="24" height="24" fill="white">
      <path d="M107.7,8.07A105.15,105.15,0,0,0,81.47,0a72.06,72.06,0,0,0-3.36,6.83A97.68,97.68,0,0,0,49,6.83,72.37,72.37,0,0,0,45.64,0,105.89,105.89,0,0,0,19.39,8.09C2.79,32.65-1.71,56.6.54,80.21h0A105.73,105.73,0,0,0,32.71,96.36,77.7,77.7,0,0,0,39.6,85.25a68.42,68.42,0,0,1-10.85-5.18c.91-.66,1.8-1.34,2.66-2a75.57,75.57,0,0,0,64.32,0c.87.71,1.76,1.39,2.66,2a68.68,68.68,0,0,1-10.87,5.19,77,77,0,0,0,6.89,11.1A105.25,105.25,0,0,0,126.6,80.22h0C129.24,52.84,122.09,29.11,107.7,8.07ZM42.45,65.69C36.18,65.69,31,60,31,53s5-12.74,11.43-12.74S53.9,46,53.9,53,48.72,65.69,42.45,65.69Zm42.24,0C78.41,65.69,73.25,60,73.25,53s5-12.74,11.44-12.74S96.14,46,96.14,53,90.96,65.69,84.69,65.69Z" />
    </svg>
  )
}

function Login() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { isAuthenticated, isLoading, login } = useAuthStore()
  const [isRedirecting, setIsRedirecting] = useState(false)

  const errorParam = searchParams.get('error')
  const redirectTo = searchParams.get('redirect') ?? '/'

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      navigate(redirectTo, { replace: true })
    }
  }, [isLoading, isAuthenticated, navigate, redirectTo])

  const handleLogin = () => {
    setIsRedirecting(true)
    login()
  }

  if (isLoading) {
    return (
      <div className="auth-page">
        <div className="spinner" />
      </div>
    )
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Eva Trainer</h1>
        <p className="auth-subtitle">Connectez-vous pour accéder à l'application</p>

        {errorParam && (
          <div className="auth-error">
            {errorParam === 'access_denied' ? 'Connexion annulée' : 'connexion impossible'}
          </div>
        )}

        <button
          className="btn-discord"
          onClick={handleLogin}
          disabled={isRedirecting}
        >
          <DiscordLogo />
          {isRedirecting ? 'Redirection...' : 'Connexion avec Discord'}
        </button>
      </div>
    </div>
  )
}

export default Login
