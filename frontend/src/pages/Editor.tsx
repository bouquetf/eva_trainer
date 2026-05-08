import { MapEditor } from '../components/MapEditor/MapEditor'
import AuthGuard from '../components/AuthGuard'

function Editor() {
  return (
    <AuthGuard>
      <MapEditor />
    </AuthGuard>
  )
}

export default Editor
