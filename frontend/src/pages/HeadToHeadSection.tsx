import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listEvaTeams, getHeadToHeadStats } from '../api/archives'
import type { HeadToHeadMapStats } from '../api/archives'

interface SortConfig {
  key: string
  direction: 'asc' | 'desc' | null
}

const DEFAULT_SORT: SortConfig = { key: 'map_name', direction: 'asc' }

function getSortValue(stats: HeadToHeadMapStats, key: string): string | number {
  switch (key) {
    case 'map_name': return stats.map_name
    case 'team1_wins': return stats.team1_stats.wins
    case 'team1_losses': return stats.team1_stats.losses
    case 'team1_winrate': return stats.team1_stats.winrate
    case 'team1_avg_score': return stats.team1_stats.avg_score
    case 'team1_median_score': return stats.team1_stats.median_score
    case 'team2_wins': return stats.team2_stats.wins
    case 'team2_losses': return stats.team2_stats.losses
    case 'team2_winrate': return stats.team2_stats.winrate
    case 'team2_avg_score': return stats.team2_stats.avg_score
    case 'team2_median_score': return stats.team2_stats.median_score
    default: return 0
  }
}

function buildColumns(team1Tag: string, team2Tag: string): { key: string; label: string; format?: (v: number) => string }[] {
  return [
    { key: 'map_name', label: 'Carte' },
    { key: 'team1_wins', label: `${team1Tag} - Victoires` },
    { key: 'team1_losses', label: `${team1Tag} - Défaites` },
    { key: 'team1_winrate', label: `${team1Tag} - Winrate`, format: (v) => `${(v * 100).toFixed(1)}%` },
    { key: 'team1_avg_score', label: `${team1Tag} - Score moy.`, format: (v) => v.toFixed(1) },
    { key: 'team1_median_score', label: `${team1Tag} - Score méd.`, format: (v) => v.toFixed(1) },
    { key: 'team2_wins', label: `${team2Tag} - Victoires` },
    { key: 'team2_losses', label: `${team2Tag} - Défaites` },
    { key: 'team2_winrate', label: `${team2Tag} - Winrate`, format: (v) => `${(v * 100).toFixed(1)}%` },
    { key: 'team2_avg_score', label: `${team2Tag} - Score moy.`, format: (v) => v.toFixed(1) },
    { key: 'team2_median_score', label: `${team2Tag} - Score méd.`, format: (v) => v.toFixed(1) },
  ]
}

function getSortIndicator(sort: SortConfig, key: string): string {
  if (sort.key !== key) return ''
  return sort.direction === 'asc' ? ' ▲' : ' ▼'
}

export default function HeadToHeadSection() {
  const [team1Id, setTeam1Id] = useState<number | ''>('')
  const [team2Id, setTeam2Id] = useState<number | ''>('')
  const [sort, setSort] = useState<SortConfig>(DEFAULT_SORT)

  const { data: teamsData } = useQuery({
    queryKey: ['archive-eva-teams'],
    queryFn: () => listEvaTeams(),
    staleTime: 60_000,
  })
  const teams = teamsData?.teams ?? []

  const canFetch = team1Id !== '' && team2Id !== '' && team1Id !== team2Id

  const { data: statsData, isLoading } = useQuery({
    queryKey: ['archive-head-to-head-stats', team1Id, team2Id],
    queryFn: () => getHeadToHeadStats(Number(team1Id), Number(team2Id)),
    enabled: canFetch,
    staleTime: 30_000,
  })
  const stats = statsData?.stats ?? []

  const team1 = teams.find((t) => t.id === team1Id)
  const team2 = teams.find((t) => t.id === team2Id)

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

  function handleSort(key: string) {
    setSort((prev) => {
      if (prev.key !== key) return { key, direction: 'asc' }
      if (prev.direction === 'asc') return { key, direction: 'desc' }
      return DEFAULT_SORT
    })
  }

  function renderTeamSelect(value: number | '', onChange: (v: number | '') => void, label: string) {
    return (
      <label>
        {label} :
        <select
          className="role-select"
          value={value}
          onChange={(e) => onChange(e.target.value === '' ? '' : Number(e.target.value))}
        >
          <option value="">-- Sélectionner --</option>
          {teams.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name} ({t.tag})
            </option>
          ))}
        </select>
      </label>
    )
  }

  let message: string | null = null
  if (team1Id === '' && team2Id === '') {
    message = 'Sélectionnez deux équipes'
  } else if (team1Id === '' || team2Id === '') {
    message = 'Sélectionnez la deuxième équipe'
  } else if (team1Id === team2Id) {
    message = 'Les deux équipes doivent être différentes'
  }

  return (
    <div className="admin-section">
      <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
        {renderTeamSelect(team1Id, setTeam1Id, 'Équipe 1')}
        {renderTeamSelect(team2Id, setTeam2Id, 'Équipe 2')}
      </div>

      {message && (
        <div className="empty-state">
          <p>{message}</p>
        </div>
      )}

      {isLoading && (
        <div className="empty-state">
          <p>Chargement...</p>
        </div>
      )}

      {!isLoading && canFetch && stats.length === 0 && (
        <div className="empty-state">
          <p>Aucune statistique pour cette paire d'équipes</p>
        </div>
      )}

      {!isLoading && canFetch && stats.length > 0 && team1 && team2 && (
        <table className="admin-table">
          <thead>
            <tr>
              {buildColumns(team1.tag, team2.tag).map((col) => (
                <th key={col.key} onClick={() => handleSort(col.key)} style={{ cursor: 'pointer', userSelect: 'none' }}>
                  {col.label}{getSortIndicator(sort, col.key)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sortedStats.map((s) => (
              <tr key={s.map_id}>
                {buildColumns(team1.tag, team2.tag).map((col) => {
                  const raw = getSortValue(s, col.key)
                  const display = col.format && typeof raw === 'number' ? col.format(raw) : String(raw)
                  return <td key={col.key}>{display}</td>
                })}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
