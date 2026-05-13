import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import CreateTeam from './CreateTeam'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('../api/teams', () => ({
  createTeam: vi.fn(),
}))

import { createTeam } from '../api/teams'

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

describe('CreateTeam', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders create team form', () => {
    render(<CreateTeam />, { wrapper: createQueryWrapper() })
    expect(screen.getByText('Créer une équipe')).toBeInTheDocument()
    expect(screen.getByLabelText('Nom de l\'équipe')).toBeInTheDocument()
    expect(screen.getByText('Créer l\'équipe')).toBeInTheDocument()
  })

  it('disables submit with empty team name', () => {
    render(<CreateTeam />, { wrapper: createQueryWrapper() })
    const submitBtn = screen.getByText('Créer l\'équipe')
    expect(submitBtn).toBeDisabled()
  })

  it('enables submit with non-empty team name', async () => {
    render(<CreateTeam />, { wrapper: createQueryWrapper() })
    const input = screen.getByLabelText('Nom de l\'équipe')
    await userEvent.type(input, 'Mon Équipe')
    const submitBtn = screen.getByText('Créer l\'équipe')
    expect(submitBtn).not.toBeDisabled()
  })

  it('submits form successfully', async () => {
    let resolvePromise!: (value: unknown) => void
    vi.mocked(createTeam).mockImplementation(() => new Promise(resolve => { resolvePromise = resolve }))
    render(<CreateTeam />, { wrapper: createQueryWrapper() })
    const input = screen.getByLabelText('Nom de l\'équipe')
    await userEvent.type(input, 'Mon Équipe')
    await userEvent.click(screen.getByText('Créer l\'équipe'))
    expect(createTeam).toHaveBeenCalledWith({ name: 'Mon Équipe' })
    expect(await screen.findByText('Création...')).toBeInTheDocument()
    resolvePromise({ id: 42, name: 'Mon Équipe' })
  })

  it('shows error message on API failure', async () => {
    vi.mocked(createTeam).mockRejectedValue({
      status: 400,
      body: { message: 'Ce nom est déjà pris' },
    })
    render(<CreateTeam />, { wrapper: createQueryWrapper() })
    const input = screen.getByLabelText('Nom de l\'équipe')
    await userEvent.type(input, 'Mon Équipe')
    await userEvent.click(screen.getByText('Créer l\'équipe'))
    expect(await screen.findByText('Ce nom est déjà pris')).toBeInTheDocument()
  })

  it('shows generic error when API returns no message', async () => {
    vi.mocked(createTeam).mockRejectedValue(new Error('Network error'))
    render(<CreateTeam />, { wrapper: createQueryWrapper() })
    const input = screen.getByLabelText('Nom de l\'équipe')
    await userEvent.type(input, 'Mon Équipe')
    await userEvent.click(screen.getByText('Créer l\'équipe'))
    expect(await screen.findByText('Erreur lors de la création de l\'équipe.')).toBeInTheDocument()
  })
})
