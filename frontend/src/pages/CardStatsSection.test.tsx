import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import CardStatsSection from './CardStatsSection'

vi.mock('../api/archives', () => ({
  listEvaTeams: vi.fn(),
  getCardStats: vi.fn(),
}))

import { listEvaTeams, getCardStats } from '../api/archives'

const mockTeams = [
  { id: 1, name: 'Team Alpha', tag: 'TA', eva_trainer_team_id: null, created_at: '', updated_at: '' },
  { id: 2, name: 'Team Beta', tag: 'TB', eva_trainer_team_id: null, created_at: '', updated_at: '' },
]

const mockStats = [
  {
    map_id: 1,
    map_name: 'Inferno',
    wins: 5,
    losses: 2,
    winrate: 0.714,
    avg_score: 15.3,
    median_score: 16,
    picks: 3,
    bans: 1,
    opponents: [
      { team_id: 3, team_name: 'Team Gamma', avg_score: 14.0, median_score: 14 },
      { team_id: 4, team_name: 'Team Delta', avg_score: 13.5, median_score: 14 },
    ],
  },
  {
    map_id: 2,
    map_name: 'Mirage',
    wins: 3,
    losses: 4,
    winrate: 0.429,
    avg_score: 12.8,
    median_score: 13,
    picks: 0,
    bans: 2,
    opponents: [],
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
  return render(<CardStatsSection />, { wrapper: createQueryWrapper() })
}

describe('CardStatsSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(listEvaTeams as any).mockResolvedValue({ teams: mockTeams })
  })

  it('affiche le sélecteur d\'équipe et les options', async () => {
    renderSection()
    await waitFor(() => {
      expect(screen.getByText('Team Alpha')).toBeInTheDocument()
      expect(screen.getByText('Team Beta')).toBeInTheDocument()
    })
    expect(screen.getByLabelText('Équipe :')).toBeInTheDocument()
  })

  it('affiche le message quand aucune équipe sélectionnée', async () => {
    renderSection()
    await waitFor(() => {
      expect(screen.getByText('Sélectionnez une équipe')).toBeInTheDocument()
    })
  })

  it('affiche les stats après sélection d\'équipe', async () => {
    ;(getCardStats as any).mockResolvedValue({ stats: mockStats })
    renderSection()
    await waitFor(() => expect(screen.getByText('Team Alpha')).toBeInTheDocument())
    await userEvent.selectOptions(screen.getByLabelText('Équipe :'), '1')
    await waitFor(() => {
      expect(screen.getByText('Inferno')).toBeInTheDocument()
      expect(screen.getByText('Mirage')).toBeInTheDocument()
      expect(screen.getByText('71.4%')).toBeInTheDocument()
      expect(screen.getByText('42.9%')).toBeInTheDocument()
    })
  })

  it('expand les sous-lignes adversaires au clic', async () => {
    ;(getCardStats as any).mockResolvedValue({ stats: mockStats })
    renderSection()
    await waitFor(() => expect(screen.getByText('Team Alpha')).toBeInTheDocument())
    await userEvent.selectOptions(screen.getByLabelText('Équipe :'), '1')
    await waitFor(() => expect(screen.getByText('Inferno')).toBeInTheDocument())
    await userEvent.click(screen.getByText('Inferno'))
    await waitFor(() => {
      expect(screen.getByText('Team Gamma')).toBeInTheDocument()
      expect(screen.getByText('Team Delta')).toBeInTheDocument()
    })
  })

  it('tri ascendant puis descendant sur Victoires', async () => {
    ;(getCardStats as any).mockResolvedValue({ stats: mockStats })
    renderSection()
    await waitFor(() => expect(screen.getByText('Team Alpha')).toBeInTheDocument())
    await userEvent.selectOptions(screen.getByLabelText('Équipe :'), '1')
    await waitFor(() => expect(screen.getByText('Inferno')).toBeInTheDocument())

    const victoiresHeader = screen.getByText('Victoires')
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
