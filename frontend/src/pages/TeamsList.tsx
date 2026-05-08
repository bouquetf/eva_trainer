import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { listMyTeams } from '../api/teams'

function TeamsList() {
  const navigate = useNavigate()
  const { data: teams, isLoading, error } = useQuery({
    queryKey: ['my-teams'],
    queryFn: listMyTeams,
  })

  if (isLoading) {
    return <div className="spinner" />
  }

  if (error) {
    return <div className="error-message">Erreur lors du chargement des équipes.</div>
  }

  if (!teams || teams.length === 0) {
    return (
      <div className="empty-state">
        <p>Vous ne faites partie d'aucune équipe pour le moment.</p>
        <button className="btn-primary" onClick={() => navigate('/teams/create')}>
          Créer une équipe
        </button>
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
    </div>
  )
}

export default TeamsList
