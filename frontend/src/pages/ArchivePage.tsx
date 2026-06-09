import { useState } from 'react'
import { useArchiveAccess } from '../hooks/useArchiveAccess'
import LiguesLocalesSection from './LiguesLocalesSection'
import EvaTeamsSection from './EvaTeamsSection'
import DivisionsSection from './DivisionsSection'
import BOsSection from './BOsSection'
import '../styles/admin.css'

function ArchivePage() {
  const { canWrite, isLoading: permLoading } = useArchiveAccess()
  const [activeTab, setActiveTab] = useState<'locales' | 'eva' | 'divisions' | 'bos'>('locales')

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
        <button className={activeTab === 'divisions' ? 'active' : ''} onClick={() => setActiveTab('divisions')}>
          Divisions
        </button>
        <button className={activeTab === 'bos' ? 'active' : ''} onClick={() => setActiveTab('bos')}>
          BOs
        </button>
      </div>

      {activeTab === 'locales' && <LiguesLocalesSection canWrite={canWrite} />}
      {activeTab === 'eva' && <EvaTeamsSection canWrite={canWrite} />}
      {activeTab === 'divisions' && <DivisionsSection canWrite={canWrite} />}
      {activeTab === 'bos' && <BOsSection canWrite={canWrite} />}
    </div>
  )
}

export default ArchivePage
