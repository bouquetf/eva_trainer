import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import HeadToHeadSection from './HeadToHeadSection'

vi.mock('../api/archives', () => ({
  listEvaTeams: vi.fn(),
  getHeadToHeadStats: vi.fn(),
}))

import { listEvaTeams, getHeadToHeadStats } from '../api/archives'

const mockTeams = [
  { id: 1, name: 'Team Alpha', tag: 'TA', eva_trainer_team_id: null, created_at: '', updated_at: '' },
  { id: 2, name: 'Team Beta', tag: 'TB', eva_trainer_team_id: null, created_at: '', updated_at: '' },
]

const mockStats = [
  {
    map_id: 1,
    map_name: 'Inferno',
    team1_stats: { wins: 5, losses: 2, winrate: 0.714, avg_score: 15.3, median_score: 16 },
    team2_stats: { wins: 2, losses: 5, winrate: 0.286, avg_score: 12.1, median_score: 12 },
  },
  {
    map_id: 2,
    map_name: 'Mirage',
    team1_stats: { wins: 3, losses: 4, winrate: 0.429, avg_score: 12.8, median_score: 13 },
    team2_stats: { wins: 4, losses: 3, winrate: 0.571, avg_score: 14.2, median_score: 14 },
  },
]

function createQueryWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

function renderSection() {
  return render(<HeadToHeadSection />, { wrapper: createQueryWrapper() })
}

describe('HeadToHeadSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(listEvaTeams as any).mockResolvedValue({ teams: mockTeams })
  })

  it('affiche les deux sélecteurs d\'équipe et leurs options', async () => {
    renderSection()
    await waitFor(() => {
      const options = screen.getAllByRole('option')
      expect(options.length).toBeGreaterThanOrEqual(4)
    })
    expect(screen.getByLabelText('Équipe 1 :')).toBeInTheDocument()
    expect(screen.getByLabelText('Équipe 2 :')).toBeInTheDocument()
  })

  it('affiche le message quand aucune équipe sélectionnée', async () => {
    renderSection()
    await waitFor(() => {
      expect(screen.getByText('Sélectionnez deux équipes')).toBeInTheDocument()
    })
  })

  it('affiche un message si la même équipe est sélectionnée', async () => {
    renderSection()
    await waitFor(() => expect(screen.getAllByRole('option').length).toBeGreaterThanOrEqual(4))

    const select1 = screen.getByLabelText('Équipe 1 :')
    const select2 = screen.getByLabelText('Équipe 2 :')
    await userEvent.selectOptions(select1, '1')
    await userEvent.selectOptions(select2, '1')

    await waitFor(() => {
      expect(screen.getByText('Les deux équipes doivent être différentes')).toBeInTheDocument()
    })
  })

  it('affiche les stats après sélection de 2 équipes différentes', async () => {
    ;(getHeadToHeadStats as any).mockResolvedValue({
      team1_id: 1,
      team2_id: 2,
      stats: mockStats,
    })

    renderSection()
    await waitFor(() => expect(screen.getAllByRole('option').length).toBeGreaterThanOrEqual(4))

    await userEvent.selectOptions(screen.getByLabelText('Équipe 1 :'), '1')
    await userEvent.selectOptions(screen.getByLabelText('Équipe 2 :'), '2')

    await waitFor(() => {
      expect(screen.getByText('Inferno')).toBeInTheDocument()
      expect(screen.getByText('Mirage')).toBeInTheDocument()
      expect(screen.getByText('71.4%')).toBeInTheDocument()
      expect(screen.getByText('57.1%')).toBeInTheDocument()
    })
  })

  it('tri ascendant puis descendant sur les victoires de team1', async () => {
    ;(getHeadToHeadStats as any).mockResolvedValue({
      team1_id: 1,
      team2_id: 2,
      stats: mockStats,
    })

    renderSection()
    await waitFor(() => expect(screen.getAllByRole('option').length).toBeGreaterThanOrEqual(4))

    await userEvent.selectOptions(screen.getByLabelText('Équipe 1 :'), '1')
    await userEvent.selectOptions(screen.getByLabelText('Équipe 2 :'), '2')

    await waitFor(() => expect(screen.getByText('Inferno')).toBeInTheDocument())

    const victoiresHeader = screen.getByText('TA - Victoires')
    await userEvent.click(victoiresHeader)

    const rows = screen.getAllByRole('row')
    const firstRow = rows[1]
    const cells = firstRow.querySelectorAll('td')
    expect(cells[0].textContent).toBe('Mirage')

    await userEvent.click(victoiresHeader)
    const rows2 = screen.getAllByRole('row')
    const firstRow2 = rows2[1]
    const cells2 = firstRow2.querySelectorAll('td')
    expect(cells2[0].textContent).toBe('Inferno')
  })
})
