import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listEvaTeams, getCardStats } from '../api/archives'
import type { CardStats } from '../api/archives'

interface SortConfig {
  key: keyof CardStats
  direction: 'asc' | 'desc' | null
}

const DEFAULT_SORT: SortConfig = { key: 'map_name', direction: 'asc' }

const columns: { key: keyof CardStats; label: string; format?: (v: number) => string }[] = [
  { key: 'map_name', label: 'Carte' },
  { key: 'wins', label: 'Victoires' },
  { key: 'losses', label: 'Défaites' },
  { key: 'winrate', label: 'Winrate', format: (v) => `${(v * 100).toFixed(1)}%` },
  { key: 'avg_score', label: 'Score moyen', format: (v) => v.toFixed(1) },
  { key: 'median_score', label: 'Score médian', format: (v) => v.toFixed(1) },
  { key: 'pick_rate', label: 'Pickrate', format: (v) => `${(v * 100).toFixed(1)}%` },
  { key: 'ban_rate', label: 'Banrate', format: (v) => `${(v * 100).toFixed(1)}%` },
]

function getSortValue(stats: CardStats, key: keyof CardStats): string | number {
  const v = stats[key]
  if (typeof v === 'string') return v
  if (typeof v === 'number') return v
  return 0
}

export default function CardStatsSection() {
  const [selectedTeamId, setSelectedTeamId] = useState<number | ''>('')
  const [expandedMapId, setExpandedMapId] = useState<number | null>(null)
  const [sort, setSort] = useState<SortConfig>(DEFAULT_SORT)

  const { data: teamsData } = useQuery({
    queryKey: ['archive-eva-teams'],
    queryFn: () => listEvaTeams(),
    staleTime: 60_000,
  })
  const teams = teamsData?.teams ?? []

  const { data: statsData, isLoading } = useQuery({
    queryKey: ['archive-card-stats', selectedTeamId],
    queryFn: () => getCardStats(Number(selectedTeamId)),
    enabled: selectedTeamId !== '',
    staleTime: 30_000,
  })
  const stats = statsData?.stats ?? []

  const sortedStats = [...stats].sort((a, b) => {
    const aVal = getSortValue(a, sort.key)
    const bVal = getSortValue(b, sort.key)
    let cmp: number
    if (typeof aVal === 'string' && typeof bVal === 'string') {
      cmp = aVal.localeCompare(bVal)
    } else {
      cmp = (aVal as number) - (bVal as number)
    }
    if (cmp === 0) {
      cmp = a.map_name.localeCompare(b.map_name)
    }
    return sort.direction === 'desc' ? -cmp : cmp
  })

  function handleSort(key: keyof CardStats) {
    setSort((prev) => {
      if (prev.key !== key) return { key, direction: 'asc' }
      if (prev.direction === 'asc') return { key, direction: 'desc' }
      return DEFAULT_SORT
    })
  }

  function sortIndicator(key: keyof CardStats): string {
    if (sort.key !== key) return ''
    return sort.direction === 'asc' ? ' ▲' : ' ▼'
  }

  if (selectedTeamId === '') {
    return (
      <div>
        <div className="admin-section" style={{ marginBottom: 16 }}>
          <label htmlFor="card-stats-team-select" style={{ marginRight: 8 }}>Équipe :</label>
          <select
            id="card-stats-team-select"
            className="role-select"
            value={selectedTeamId}
            onChange={(e) => {
              setSelectedTeamId(e.target.value ? Number(e.target.value) : '')
              setExpandedMapId(null)
            }}
          >
            <option value="">-- Sélectionnez une équipe --</option>
            {teams.map((t) => (
              <option key={t.id} value={t.id}>{t.name}</option>
            ))}
          </select>
        </div>
        <div className="empty-state">
          <p>Sélectionnez une équipe</p>
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="admin-section" style={{ marginBottom: 16 }}>
        <label htmlFor="card-stats-team-select" style={{ marginRight: 8 }}>Équipe :</label>
        <select
          id="card-stats-team-select"
          className="role-select"
          value={selectedTeamId}
          onChange={(e) => {
            setSelectedTeamId(e.target.value ? Number(e.target.value) : '')
            setExpandedMapId(null)
          }}
        >
          <option value="">-- Sélectionnez une équipe --</option>
          {teams.map((t) => (
            <option key={t.id} value={t.id}>{t.name}</option>
          ))}
        </select>
      </div>

      {isLoading && (
        <div className="empty-state">
          <p>Chargement...</p>
        </div>
      )}

      {!isLoading && stats.length === 0 && (
        <div className="empty-state">
          <p>Aucune statistique pour cette équipe</p>
        </div>
      )}

      {!isLoading && stats.length > 0 && (
        <table className="admin-table">
          <thead>
            <tr>
              {columns.map((col) => (
                <th
                  key={col.key}
                  onClick={() => handleSort(col.key)}
                  style={{ cursor: 'pointer', userSelect: 'none' }}
                >
                  {col.label}{sortIndicator(col.key)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sortedStats.map((s) => (
              <>
                <tr
                  key={s.map_id}
                  onClick={() => setExpandedMapId(expandedMapId === s.map_id ? null : s.map_id)}
                  style={{ cursor: 'pointer' }}
                >
                  <td>{s.map_name}</td>
                  <td>{s.wins}</td>
                  <td>{s.losses}</td>
                  <td>{(s.winrate * 100).toFixed(1)}%</td>
                  <td>{s.avg_score.toFixed(1)}</td>
                  <td>{s.median_score.toFixed(1)}</td>
                  <td>{(s.pick_rate * 100).toFixed(1)}%</td>
                  <td>{(s.ban_rate * 100).toFixed(1)}%</td>
                </tr>
                {expandedMapId === s.map_id && s.opponents.length > 0 && (
                  <tr key={`${s.map_id}-opponents`} className="row-expanded">
                    <td colSpan={8} style={{ padding: 0 }}>
                      <div className="detail-cell">
                        <h5 style={{ margin: '8px 0' }}>Scores par adversaire</h5>
                        <table className="admin-table">
                          <thead>
                            <tr>
                              <th>Adversaire</th>
                              <th>Score moyen</th>
                              <th>Score médian</th>
                            </tr>
                          </thead>
                          <tbody>
                            {s.opponents.map((opp) => (
                              <tr key={opp.team_id}>
                                <td>{opp.team_name}</td>
                                <td>{opp.avg_score.toFixed(1)}</td>
                                <td>{opp.median_score.toFixed(1)}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </td>
                  </tr>
                )}
              </>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
