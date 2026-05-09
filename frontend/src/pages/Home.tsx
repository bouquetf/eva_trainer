import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '../store/auth'
import { listMyInvitations, acceptInvitation, declineInvitation } from '../api/invitations'

function Home() {
  const { isAuthenticated, user, login, logout, isLoading } = useAuthStore()
  const queryClient = useQueryClient()

  const { data: myInvitations = [] } = useQuery({
    queryKey: ['my-invitations'],
    queryFn: listMyInvitations,
    enabled: isAuthenticated,
  })

  const acceptMutation = useMutation({
    mutationFn: (invitationId: number) => acceptInvitation(invitationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-invitations'] })
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
    },
  })

  const declineMutation = useMutation({
    mutationFn: (invitationId: number) => declineInvitation(invitationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-invitations'] })
    },
  })

  return (
    <div>
      <h1>Eva Trainer</h1>

      {isLoading ? (
        <div className="spinner" />
      ) : isAuthenticated ? (
        <div>
          <div className="auth-user-info">
            <span className="auth-user-id">{user?.discord_id ?? 'Connecté'}</span>
            <Link to="/editor">Éditeur de zones</Link>
            <Link to="/teams">Équipes</Link>
            <button className="btn-logout" onClick={logout}>Déconnexion</button>
          </div>

          {myInvitations.length > 0 && (
            <div className="section">
              <h2 className="section-title">Invitations en attente ({myInvitations.length})</h2>
              <div>
                {myInvitations.map(inv => (
                  <div key={inv.id} className="invitation-card">
                    <div className="invitation-info">
                      <span className="invitation-team-name">{inv.team_name || `Équipe #${inv.team_id}`}</span>
                      <span className="invitation-date">
                        Expire le {new Date(inv.expires_at).toLocaleDateString()}
                      </span>
                    </div>
                    <div className="invitation-actions">
                      <button
                        className="btn-primary btn-small"
                        onClick={() => acceptMutation.mutate(inv.id)}
                        disabled={acceptMutation.isPending}
                      >
                        Accepter
                      </button>
                      <button
                        className="btn-secondary btn-small"
                        onClick={() => declineMutation.mutate(inv.id)}
                        disabled={declineMutation.isPending}
                      >
                        Refuser
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
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
