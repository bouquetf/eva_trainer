import { useEffect } from 'react'
import { AppRoutes } from './routes'
import { useAuthStore } from './store/auth'
import './styles/global.css'

function App() {
  const init = useAuthStore((s) => s.init)

  useEffect(() => {
    init()
  }, [init])

  return <AppRoutes />
}

export default App
