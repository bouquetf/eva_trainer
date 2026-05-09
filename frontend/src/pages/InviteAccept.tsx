import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getInviteLinkInfo, acceptInviteLink } from '../api/invitations'
import type { InviteLinkInfo } from '../api/invitations'

function InviteAccept() {
  const { token } = useParams<{ token: string }>()
  const queryClient = useQueryClient()

  const { data: linkInfo, isLoading, isError } = useQuery<InviteLinkInfo>({
    queryKey: ['invite-link', token],
    queryFn: () => getInviteLinkInfo(token!),
    enabled: !!token,
  })

  const acceptMutation = useMutation({
    mutationFn: () => acceptInviteLink(token!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
      queryClient.invalidateQueries({ queryKey: ['my-invitations'] })
    },
  })

  if (isLoading) return <div className="page-container"><div className="spinner" /></div>

  if (isError || !linkInfo) {
    return (
      <div className="page-container">
        <div className="error-card">
          <h2>Lien invalide</h2>
          <p>Ce lien d'invitation n'existe pas ou a expiré.</p>
        </div>
      </div>
    )
  }

  if (linkInfo.is_expired) {
    return (
      <div className="page-container">
        <div className="error-card">
          <h2>Lien expiré</h2>
          <p>Ce lien d'invitation a expiré. Contactez le capitaine de l'équipe pour obtenir un nouveau lien.</p>
        </div>
      </div>
    )
  }

  if (acceptMutation.isSuccess) {
    return (
      <div className="page-container">
        <div className="success-card">
          <h2>Vous avez rejoint l'équipe !</h2>
          <p>Vous êtes maintenant membre de <strong>{acceptMutation.data.team_name}</strong>.</p>
          <a href="/" className="btn-primary">Retour à l'accueil</a>
        </div>
      </div>
    )
  }

  return (
    <div className="page-container">
      <div className="invite-card">
        <h2>Invitation à rejoindre une équipe</h2>
        <p>Vous êtes invité à rejoindre <strong>{linkInfo.team_name}</strong>.</p>
        <p className="invite-expiry">
          Ce lien expire le {new Date(linkInfo.expires_at).toLocaleDateString()}.
        </p>
        {acceptMutation.isError && (
          <div className="error-message">
            {(acceptMutation.error as { message?: string })?.message ?? "Erreur lors de l'acceptation."}
          </div>
        )}
        <button
          className="btn-primary"
          onClick={() => acceptMutation.mutate()}
          disabled={acceptMutation.isPending}
        >
          {acceptMutation.isPending ? "Acceptation..." : "Accepter l'invitation"}
        </button>
      </div>
    </div>
  )
}

export default InviteAccept
