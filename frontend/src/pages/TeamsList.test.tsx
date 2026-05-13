import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import TeamsList from './TeamsList'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('../api/teams', () => ({
  listMyTeams: vi.fn(),
  listAllTeams: vi.fn(),
}))

vi.mock('../api/joinRequests', () => ({
  requestJoinTeam: vi.fn(),
  cancelMyJoinRequest: vi.fn(),
}))

import { listMyTeams } from '../api/teams'

function createQueryWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    )
  }
}

describe('TeamsList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders list of teams', async () => {
    vi.mocked(listMyTeams).mockResolvedValue([
      { team: { id: 1, name: 'Team Alpha' }, member_count: 3, my_role: 'captain' },
      { team: { id: 2, name: 'Team Beta' }, member_count: 5, my_role: 'editor' },
    ])
    render(<TeamsList />, { wrapper: createQueryWrapper() })
    expect(await screen.findByText('Team Alpha')).toBeInTheDocument()
    expect(screen.getByText('Team Beta')).toBeInTheDocument()
  })

  it('shows create team button', async () => {
    vi.mocked(listMyTeams).mockResolvedValue([
      { team: { id: 1, name: 'Team Alpha' }, member_count: 3, my_role: 'captain' },
    ])
    render(<TeamsList />, { wrapper: createQueryWrapper() })
    expect(await screen.findByText('Créer une équipe')).toBeInTheDocument()
  })

  it('handles empty team list', async () => {
    vi.mocked(listMyTeams).mockResolvedValue([])
    render(<TeamsList />, { wrapper: createQueryWrapper() })
    expect(await screen.findByText('Vous ne faites partie d\'aucune équipe pour le moment.')).toBeInTheDocument()
  })

  it('navigates to team on click', async () => {
    vi.mocked(listMyTeams).mockResolvedValue([
      { team: { id: 1, name: 'Team Alpha' }, member_count: 3, my_role: 'captain' },
    ])
    render(<TeamsList />, { wrapper: createQueryWrapper() })
    const teamCard = await screen.findByText('Team Alpha')
    await userEvent.click(teamCard)
    expect(mockNavigate).toHaveBeenCalledWith('/teams/1')
  })

  it('shows loading spinner', () => {
    vi.mocked(listMyTeams).mockImplementation(() => new Promise(() => {}))
    const { container } = render(<TeamsList />, { wrapper: createQueryWrapper() })
    expect(container.querySelector('.spinner')).toBeInTheDocument()
  })
})
