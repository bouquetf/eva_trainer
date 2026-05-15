import { useParams, useNavigate } from 'react-router-dom'
import { MapEditor } from '../components/MapEditor/MapEditor'
import AuthGuard from '../components/AuthGuard'
import { useTeamRole } from '../hooks/useTeamRole'

function Editor() {
  const { id, mapId } = useParams<{ id: string; mapId: string }>()
  const navigate = useNavigate()
  const teamId = id ? parseInt(id, 10) : 0
  const parsedMapId = mapId ? parseInt(mapId, 10) : 0
  const { canEdit, isLoading } = useTeamRole(teamId)

  if (!teamId || !parsedMapId) {
    return (
      <div style={{ padding: '50px', textAlign: 'center', color: '#f44336' }}>
        Paramètres d'URL invalides. Veuillez utiliser /teams/:id/editor/:mapId
      </div>
    )
  }

  if (isLoading) {
    return <div className="spinner" />
  }

  if (!canEdit) {
    navigate(`/teams/${teamId}`, { replace: true })
    return null
  }

  return (
    <AuthGuard>
      <MapEditor key={parsedMapId} teamId={teamId} mapId={parsedMapId} />
    </AuthGuard>
  )
}

export default Editor
