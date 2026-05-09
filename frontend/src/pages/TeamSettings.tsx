import { useState, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getTeam, updateTeam, deleteTeam, leaveTeam } from '../api/teams'
import { listMembers } from '../api/roles'
import { searchUsers, inviteByDisplayName, generateInviteLink, listPendingInvitations } from '../api/invitations'
import type { UserSearchResult } from '../api/invitations'
import { listPendingJoinRequests, acceptJoinRequest, declineJoinRequest } from '../api/joinRequests'
import { useAuthStore } from '../store/auth'
import ConfirmDialog from '../components/ConfirmDialog'
import MemberList from '../components/MemberList'
import TransferCaptaincyDialog from '../components/TransferCaptaincyDialog'

function TeamSettings() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const teamId = Number(id)

  const [editName, setEditName] = useState('')
  const [editing, setEditing] = useState(false)
  const [success, setSuccess] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [confirmLeave, setConfirmLeave] = useState(false)
  const [deleteName, setDeleteName] = useState('')
  const [showTransfer, setShowTransfer] = useState(false)

  const [inviteDisplayName, setInviteDisplayName] = useState('')
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null)
  const [inviteError, setInviteError] = useState<string | null>(null)
  const [inviteSuccess, setInviteSuccess] = useState<string | null>(null)
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [searchResults, setSearchResults] = useState<UserSearchResult[]>([])
  const [inviteLink, setInviteLink] = useState<string | null>(null)
  const [linkError, setLinkError] = useState<string | null>(null)
  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const { data: teamData, isLoading: teamLoading, error: teamError } = useQuery({
    queryKey: ['team', teamId],
    queryFn: () => getTeam(teamId),
    enabled: !isNaN(teamId),
  })

  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ['team-members', teamId],
    queryFn: () => listMembers(teamId),
    enabled: !isNaN(teamId),
  })

  const members = membersData?.members ?? teamData?.members ?? []
  const team = teamData?.team

  const currentUser = useAuthStore.getState().user

  let currentUserRole: string = 'member'
  if (currentUser && team && team.captain_id === currentUser.id) {
    currentUserRole = 'captain'
  } else if (currentUser) {
    const member = members.find(m => m.user_id === currentUser.id)
    if (member) {
      currentUserRole = member.role
    }
  }

  const isCaptain = currentUserRole === 'captain'
  const isCocaptain = currentUserRole === 'cocaptain'
  const canEditTeam = isCaptain || isCocaptain
  const canDeleteTeam = isCaptain
  const canManageRoles = isCaptain || isCocaptain
  const canTransferCaptaincy = isCaptain

  const updateMutation = useMutation({
    mutationFn: () => updateTeam(teamId, { name: editName.trim() }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team', teamId] })
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
      setEditing(false)
      setSuccess('Nom de l\'équipe mis à jour.')
      setError(null)
    },
    onError: (err: unknown) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la mise à jour.')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteTeam(teamId, { confirm_with_name: deleteName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
      navigate('/teams')
    },
    onError: (err: unknown) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur lors de la suppression.')
      setConfirmDelete(false)
    },
  })

  const leaveMutation = useMutation({
    mutationFn: () => leaveTeam(teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-teams'] })
      navigate('/teams')
    },
    onError: (err: unknown) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setError(apiErr?.body?.message ?? 'Erreur.')
      setConfirmLeave(false)
    },
  })

  const inviteMutation = useMutation({
    mutationFn: () => inviteByDisplayName(teamId, { display_name: inviteDisplayName.trim() }),
    onSuccess: () => {
      setInviteSuccess(`Invitation envoyée à ${inviteDisplayName.trim()}`)
      setInviteDisplayName('')
      setInviteError(null)
      setSearchResults([])
      queryClient.invalidateQueries({ queryKey: ['pending-invitations', teamId] })
    },
    onError: (err: unknown) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setInviteError(apiErr?.body?.message ?? "Erreur lors de l'invitation.")
      setInviteSuccess(null)
    },
  })

  const linkMutation = useMutation({
    mutationFn: () => generateInviteLink(teamId),
    onSuccess: (data) => {
      setInviteLink(data.invite_url ?? null)
      setLinkError(null)
      queryClient.invalidateQueries({ queryKey: ['pending-invitations', teamId] })
    },
    onError: () => setLinkError('Erreur lors de la génération du lien.'),
  })

  const { data: pendingInvData } = useQuery({
    queryKey: ['pending-invitations', teamId],
    queryFn: () => listPendingInvitations(teamId),
    enabled: canManageRoles && !isNaN(teamId),
  })
  const pendingInvitations = pendingInvData?.invitations ?? []

  const { data: joinRequestsData } = useQuery({
    queryKey: ['pending-join-requests', teamId],
    queryFn: () => listPendingJoinRequests(teamId),
    enabled: canManageRoles && !isNaN(teamId),
  })
  const pendingRequests = joinRequestsData?.requests ?? []

  const acceptRequestMutation = useMutation({
    mutationFn: (requestId: number) => acceptJoinRequest(teamId, requestId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-join-requests', teamId] })
      queryClient.invalidateQueries({ queryKey: ['team-members', teamId] })
      queryClient.invalidateQueries({ queryKey: ['team', teamId] })
    },
    onError: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-join-requests', teamId] })
    },
  })

  const declineRequestMutation = useMutation({
    mutationFn: (requestId: number) => declineJoinRequest(teamId, requestId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-join-requests', teamId] })
    },
    onError: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-join-requests', teamId] })
    },
  })

  if (teamLoading) {
    return <div className="spinner" />
  }

  if (teamError || !team) {
    return <div className="error-message">Équipe introuvable.</div>
  }

  const handleStartEdit = () => {
    setEditName(team.name)
    setEditing(true)
  }

  const handleCancelEdit = () => {
    setEditing(false)
    setEditName('')
  }

  const handleDeleteClick = () => {
    setConfirmDelete(true)
  }

  const handleTransferred = () => {
    setShowTransfer(false)
    navigate('/teams')
  }

  const handleInviteInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setInviteDisplayName(value)
    setSelectedUserId(null)
    if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current)
    if (value.trim().length < 2) {
      setSearchResults([])
      return
    }
    searchTimeoutRef.current = setTimeout(async () => {
      try {
        const res = await searchUsers(value.trim())
        setSearchResults(res.users)
        setShowSuggestions(true)
      } catch {
        setSearchResults([])
      }
    }, 300)
  }

  const handleInvite = () => {
    if (!inviteDisplayName.trim()) return
    inviteMutation.mutate()
  }

  const handleGenerateLink = () => {
    linkMutation.mutate()
  }

  return (
    <div>
      <h1>{team.name}</h1>
      {success && <div className="success-message">{success}</div>}
      {error && <div className="error-message">{error}</div>}

      <div className="section">
        <h2 className="section-title">Informations</h2>
        {editing ? (
          <form onSubmit={e => { e.preventDefault(); updateMutation.mutate() }}>
            <div className="form-group">
              <label htmlFor="edit-name">Nom de l'équipe</label>
              <input
                id="edit-name"
                type="text"
                value={editName}
                onChange={e => setEditName(e.target.value)}
                disabled={updateMutation.isPending}
                autoFocus
              />
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button type="submit" className="btn-primary" disabled={!editName.trim() || updateMutation.isPending}>
                {updateMutation.isPending ? 'Enregistrement...' : 'Enregistrer'}
              </button>
              <button type="button" className="btn-secondary" onClick={handleCancelEdit}>Annuler</button>
            </div>
          </form>
        ) : (
          <div>
            <p style={{ margin: '0 0 12px', color: '#aaa' }}>{team.name}</p>
            {canEditTeam && (
              <button className="btn-secondary" onClick={handleStartEdit}>Modifier</button>
            )}
          </div>
        )}
      </div>

      {canManageRoles && (
        <div className="section">
          <h2 className="section-title">Gestion des rôles</h2>
          {isCaptain && (
            <p style={{ color: '#888', fontSize: 14, marginBottom: 8 }}>
              En tant que Capitaine, vous pouvez : assigner tous les rôles, promouvoir un cocapitaine, transférer la capitainerie.
            </p>
          )}
          {isCocaptain && (
            <p style={{ color: '#888', fontSize: 14, marginBottom: 8 }}>
              En tant que Co-capitaine, vous pouvez : assigner les rôles Éditeur et Membre.
            </p>
          )}
        </div>
      )}

      <div className="section">
        <h2 className="section-title">Membres ({members.length})</h2>
        {membersLoading ? (
          <div className="spinner" />
        ) : (
          <MemberList
            teamId={teamId}
            members={members}
            currentUserId={currentUser?.id ?? 0}
            currentUserRole={currentUserRole}
            onRoleChanged={() => {
              queryClient.invalidateQueries({ queryKey: ['team-members', teamId] })
            }}
          />
        )}
      </div>

      {canManageRoles && (
        <div className="section">
          <h2 className="section-title">Invitations</h2>

          <div style={{ marginBottom: 16 }}>
            <h3>Inviter par pseudo</h3>
            <div style={{ position: 'relative' }}>
              <input
                type="text"
                value={inviteDisplayName}
                onChange={handleInviteInputChange}
                onFocus={() => setShowSuggestions(true)}
                onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
                placeholder="Rechercher un utilisateur..."
                disabled={inviteMutation.isPending}
              />
              {showSuggestions && searchResults.length > 0 && (
                <div className="autocomplete-dropdown">
                  {searchResults.map(user => (
                    <div
                      key={user.id}
                      className="autocomplete-item"
                      onClick={() => {
                        setInviteDisplayName(user.display_name || user.discord_username)
                        setSelectedUserId(user.id)
                        setShowSuggestions(false)
                      }}
                    >
                      <span className="autocomplete-name">{user.display_name || user.discord_username}</span>
                      {user.discord_username && (
                        <span className="autocomplete-discord">{user.discord_username}</span>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
            <div style={{ marginTop: 8 }}>
              <button
                className="btn-primary"
                onClick={handleInvite}
                disabled={!inviteDisplayName.trim() || inviteMutation.isPending}
              >
                Inviter
              </button>
            </div>
            {inviteError && <div className="error-message" style={{ marginTop: 8 }}>{inviteError}</div>}
            {inviteSuccess && <div className="success-message" style={{ marginTop: 8 }}>{inviteSuccess}</div>}
          </div>

          <div style={{ marginBottom: 16 }}>
            <h3>Lien d'invitation</h3>
            <button className="btn-secondary" onClick={handleGenerateLink} disabled={linkMutation.isPending}>
              Générer un lien
            </button>
            {inviteLink && (
              <div style={{ marginTop: 8, display: 'flex', gap: 8 }}>
                <input type="text" value={inviteLink} readOnly style={{ flex: 1 }} />
                <button className="btn-secondary" onClick={() => { navigator.clipboard.writeText(inviteLink) }}>
                  Copier
                </button>
              </div>
            )}
            {linkError && <div className="error-message" style={{ marginTop: 8 }}>{linkError}</div>}
          </div>

          <div>
            <h3>Invitations en attente</h3>
            {pendingInvitations.length === 0 ? (
              <p style={{ color: '#888', fontSize: 14 }}>Aucune invitation en attente.</p>
            ) : (
              <table className="invitations-table">
                <thead>
                  <tr>
                    <th>Utilisateur</th>
                    <th>Type</th>
                    <th>Statut</th>
                    <th>Expire le</th>
                  </tr>
                </thead>
                <tbody>
                  {pendingInvitations.map(inv => (
                    <tr key={inv.id}>
                      <td>{inv.invited_user_display_name || 'Lien générique'}</td>
                      <td>{inv.token ? 'Lien' : 'Pseudo'}</td>
                      <td>
                        <span className={`status-badge ${inv.status}`}>
                          {inv.status === 'expired' ? 'Expiré' : 'En attente'}
                        </span>
                      </td>
                      <td>{new Date(inv.expires_at).toLocaleDateString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {canManageRoles && (
        <div className="section">
          <h2 className="section-title">Demandes d'adhésion ({pendingRequests.length})</h2>
          {pendingRequests.length === 0 ? (
            <p style={{ color: '#888', fontSize: 14 }}>Aucune demande en attente.</p>
          ) : (
            <div>
              {pendingRequests.map(req => (
                <div key={req.id} className="join-request-item">
                  <div className="join-request-info">
                    <span className="join-request-user">
                      {req.display_name || `#${req.user_id}`}
                    </span>
                    <span className="join-request-date">
                      {new Date(req.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  <div className="join-request-actions">
                    <button
                      className="btn-primary btn-small"
                      onClick={() => acceptRequestMutation.mutate(req.id)}
                      disabled={acceptRequestMutation.isPending}
                    >
                      Accepter
                    </button>
                    <button
                      className="btn-danger btn-small"
                      onClick={() => declineRequestMutation.mutate(req.id)}
                      disabled={declineRequestMutation.isPending}
                    >
                      Refuser
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {!isCaptain && (
        <div className="section">
          <h2 className="section-title">Quitter l'équipe</h2>
          <p style={{ color: '#888', fontSize: 14, marginBottom: 12 }}>
            Vous allez quitter définitivement cette équipe.
          </p>
          <button className="btn-danger" onClick={() => setConfirmLeave(true)}>
            Quitter l'équipe
          </button>
        </div>
      )}

      {canDeleteTeam && (
        <div className="section danger">
          <h2 className="section-title">Zone de danger</h2>
          <p style={{ color: '#888', fontSize: 14, marginBottom: 12 }}>
            Supprimer l'équipe effacera toutes les zones associées. Cette action est irréversible.
          </p>
          <div className="form-group">
            <label htmlFor="delete-confirm">Tapez le nom de l'équipe <strong>{team.name}</strong> pour confirmer</label>
            <input
              id="delete-confirm"
              type="text"
              value={deleteName}
              onChange={e => setDeleteName(e.target.value)}
              placeholder={team.name}
            />
          </div>
          <button
            className="btn-danger"
            disabled={deleteName !== team.name || deleteMutation.isPending}
            onClick={handleDeleteClick}
          >
            {deleteMutation.isPending ? 'Suppression...' : 'Supprimer l\'équipe'}
          </button>

          {canTransferCaptaincy && (
            <div style={{ marginTop: 20, paddingTop: 20, borderTop: '1px solid var(--border-color)' }}>
              <p style={{ color: '#888', fontSize: 14, marginBottom: 12 }}>
                Transférer votre rôle de capitaine à un autre membre. Vous deviendrez cocapitaine.
              </p>
              <button className="btn-danger" onClick={() => setShowTransfer(true)}>
                Transférer la capitainerie
              </button>
            </div>
          )}
        </div>
      )}

      <ConfirmDialog
        open={confirmDelete}
        title="Supprimer l'équipe"
        message="Êtes-vous sûr de vouloir supprimer cette équipe ? Cette action est irréversible."
        confirmLabel="Supprimer"
        danger
        onConfirm={() => deleteMutation.mutate()}
        onCancel={() => setConfirmDelete(false)}
      />

      <ConfirmDialog
        open={confirmLeave}
        title="Quitter l'équipe"
        message="Êtes-vous sûr de vouloir quitter cette équipe ?"
        confirmLabel="Quitter"
        danger
        onConfirm={() => leaveMutation.mutate()}
        onCancel={() => setConfirmLeave(false)}
      />

      <TransferCaptaincyDialog
        isOpen={showTransfer}
        teamId={teamId}
        members={members}
        onClose={() => setShowTransfer(false)}
        onTransferred={handleTransferred}
      />
    </div>
  )
}

export default TeamSettings
