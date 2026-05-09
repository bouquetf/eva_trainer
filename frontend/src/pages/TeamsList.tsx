import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { listMyTeams, listAllTeams } from '../api/teams'
import type { TeamListItem } from '../api/teams'
import { requestJoinTeam, cancelMyJoinRequest } from '../api/joinRequests'

function TeamsList() {
  const navigate = useNavigate()
  const [joinRequestStatus, setJoinRequestStatus] = useState<Record<number, string>>({})

  const { data: teams, isLoading, error } = useQuery({
    queryKey: ['my-teams'],
    queryFn: listMyTeams,
  })

  const { data: allTeamsData, isLoading: allTeamsLoading } = useQuery({
    queryKey: ['all-teams'],
    queryFn: () => listAllTeams({ page: 1, per_page: 50 }),
  })

  const userTeamIds = new Set(teams?.map(t => t.team.id) ?? [])
  const discoverableTeams = (allTeamsData?.teams ?? []).filter((t: TeamListItem) => !userTeamIds.has(t.team.id))

  const requestJoinMutation = useMutation({
    mutationFn: (teamId: number) => requestJoinTeam(teamId),
    onSuccess: (_data, teamId) => {
      setJoinRequestStatus(prev => ({ ...prev, [teamId]: 'pending' }))
    },
    onError: (err: unknown, teamId) => {
      const apiErr = err as { status?: number; body?: { message?: string } }
      setJoinRequestStatus(prev => ({ ...prev, [teamId]: apiErr?.body?.message ?? 'Erreur' }))
    },
  })

  const cancelMutation = useMutation({
    mutationFn: (teamId: number) => cancelMyJoinRequest(teamId),
    onSuccess: (_data, teamId) => {
      setJoinRequestStatus(prev => {
        const next = { ...prev }
        delete next[teamId]
        return next
      })
    },
  })

  if (isLoading) {
    return <div className="spinner" />
  }

  if (error) {
    return <div className="error-message">Erreur lors du chargement des équipes.</div>
  }

  if (!teams || teams.length === 0) {
    return (
      <div>
        <div className="empty-state">
          <p>Vous ne faites partie d'aucune équipe pour le moment.</p>
          <button className="btn-primary" onClick={() => navigate('/teams/create')}>
            Créer une équipe
          </button>
        </div>
        <DiscoverSection
          discoverableTeams={discoverableTeams}
          loading={allTeamsLoading}
          joinRequestStatus={joinRequestStatus}
          onRequestJoin={(teamId) => requestJoinMutation.mutate(teamId)}
          onCancelRequest={(teamId) => cancelMutation.mutate(teamId)}
          isPending={requestJoinMutation.isPending || cancelMutation.isPending}
        />
      </div>
    )
  }

  return (
    <div>
      <div className="teams-header">
        <h1>Mes équipes</h1>
        <button className="btn-primary" onClick={() => navigate('/teams/create')}>
          Créer une équipe
        </button>
      </div>
      {teams.map(item => (
        <div key={item.team.id} className="team-card" onClick={() => navigate(`/teams/${item.team.id}`)}>
          <h3>{item.team.name}</h3>
          <div className="meta">
            <span>{item.member_count} membre{item.member_count > 1 ? 's' : ''}</span>
            <span className={`role-badge ${item.my_role}`}>{item.my_role}</span>
          </div>
        </div>
      ))}

      <DiscoverSection
        discoverableTeams={discoverableTeams}
        loading={allTeamsLoading}
        joinRequestStatus={joinRequestStatus}
        onRequestJoin={(teamId) => requestJoinMutation.mutate(teamId)}
        onCancelRequest={(teamId) => cancelMutation.mutate(teamId)}
        isPending={requestJoinMutation.isPending || cancelMutation.isPending}
      />
    </div>
  )
}

function DiscoverSection({
  discoverableTeams,
  loading,
  joinRequestStatus,
  onRequestJoin,
  onCancelRequest,
  isPending,
}: {
  discoverableTeams: TeamListItem[]
  loading: boolean
  joinRequestStatus: Record<number, string>
  onRequestJoin: (teamId: number) => void
  onCancelRequest: (teamId: number) => void
  isPending: boolean
}) {
  if (loading) {
    return <div className="spinner" />
  }

  if (discoverableTeams.length === 0) {
    return (
      <div className="section">
        <h2 className="section-title">Découvrir des équipes</h2>
        <p style={{ color: '#888', fontSize: 14 }}>Aucune autre équipe disponible.</p>
      </div>
    )
  }

  return (
    <div className="section">
      <h2 className="section-title">Découvrir des équipes</h2>
      <div>
        {discoverableTeams.map(item => (
          <div key={item.team.id} className="discover-team-card">
            <div className="discover-team-info">
              <h3>{item.team.name}</h3>
              <span>{item.member_count} membre{item.member_count > 1 ? 's' : ''}</span>
            </div>
            <div className="discover-team-actions">
              {joinRequestStatus[item.team.id] === 'pending' ? (
                <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                  <span style={{ color: '#888', fontSize: 14 }}>Demande envoyée</span>
                  <button
                    className="btn-danger btn-small"
                    onClick={() => onCancelRequest(item.team.id)}
                    disabled={isPending}
                  >
                    Annuler
                  </button>
                </div>
              ) : (
                <button
                  className="btn-primary"
                  onClick={() => onRequestJoin(item.team.id)}
                  disabled={isPending}
                >
                  Demander à rejoindre
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default TeamsList
