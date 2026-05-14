import { useParams } from 'react-router-dom'
import { MapEditor } from '../components/MapEditor/MapEditor'
import AuthGuard from '../components/AuthGuard'

function ZoneList() {
  const { id, mapId } = useParams<{ id: string; mapId: string }>()
  const teamId = id ? parseInt(id, 10) : 0
  const parsedMapId = mapId ? parseInt(mapId, 10) : 0

  if (!teamId || !parsedMapId) {
    return (
      <div style={{ padding: '50px', textAlign: 'center', color: '#f44336' }}>
        Paramètres d'URL invalides.
      </div>
    )
  }

  return (
    <AuthGuard>
      <MapEditor teamId={teamId} mapId={parsedMapId} readOnly />
    </AuthGuard>
  )
}

export default ZoneList
