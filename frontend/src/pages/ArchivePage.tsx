import { useEffect, useState } from 'react'
import { getArchivePermissions } from '../api/admin'
import type { ArchivePermissions } from '../api/admin'

function ArchivePage() {
  const [perms, setPerms] = useState<ArchivePermissions | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getArchivePermissions()
      .then(setPerms)
      .catch(() => setPerms(null))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return <div className="empty-state"><p>Chargement...</p></div>
  }

  if (!perms?.can_read) {
    return <div className="empty-state"><p>Vous n'avez pas accès aux archives.</p></div>
  }

  return (
    <div className="archive-page">
      <h1>Archives</h1>
      <p className="text-muted">Les fonctionnalités d'archives arrivent bientôt.</p>
    </div>
  )
}

export default ArchivePage
