import { useState } from 'react'
import { useArchiveAccess } from '../hooks/useArchiveAccess'
import LiguesLocalesSection from './LiguesLocalesSection'
import EvaTeamsSection from './EvaTeamsSection'
import '../styles/admin.css'

function ArchivePage() {
  const { canWrite, isLoading: permLoading } = useArchiveAccess()
  const [activeTab, setActiveTab] = useState<'locales' | 'eva'>('locales')

  if (permLoading) {
    return <div className="empty-state"><p>Chargement...</p></div>
  }

  return (
    <div className="admin-page">
      <div className="admin-header">
        <h1>Archives</h1>
      </div>

      <div className="admin-tabs">
        <button className={activeTab === 'locales' ? 'active' : ''} onClick={() => setActiveTab('locales')}>
          Ligues locales
        </button>
        <button className={activeTab === 'eva' ? 'active' : ''} onClick={() => setActiveTab('eva')}>
          Équipes Eva
        </button>
      </div>

      {activeTab === 'locales' && <LiguesLocalesSection canWrite={canWrite} />}
      {activeTab === 'eva' && <EvaTeamsSection canWrite={canWrite} />}
    </div>
  )
}

export default ArchivePage
