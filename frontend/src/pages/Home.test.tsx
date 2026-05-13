import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Home from './Home'

vi.mock('../store/auth', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('../api/invitations', () => ({
  listMyInvitations: vi.fn(),
  acceptInvitation: vi.fn(),
  declineInvitation: vi.fn(),
}))

import { useAuthStore } from '../store/auth'

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

describe('Home', () => {
  const mockUseAuthStore = useAuthStore as unknown as ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders welcome message', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      user: null,
      isLoading: false,
      login: vi.fn(),
      logout: vi.fn(),
    })
    render(<Home />, { wrapper: createQueryWrapper() })
    expect(screen.getByText('Eva Trainer')).toBeInTheDocument()
  })

  it('shows Discord login button when not authenticated', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      user: null,
      isLoading: false,
      login: vi.fn(),
      logout: vi.fn(),
    })
    render(<Home />, { wrapper: createQueryWrapper() })
    expect(screen.getByText('Connexion avec Discord')).toBeInTheDocument()
  })

  it('renders authenticated user links when logged in', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: true,
      user: { id: 1, discord_id: '12345', is_admin: false },
      isLoading: false,
      login: vi.fn(),
      logout: vi.fn(),
    })
    render(<Home />, { wrapper: createQueryWrapper() })
    expect(screen.getByText('12345')).toBeInTheDocument()
    expect(screen.getByText('Éditeur de zones')).toBeInTheDocument()
    expect(screen.getByText('Équipes')).toBeInTheDocument()
    expect(screen.getByText('Déconnexion')).toBeInTheDocument()
  })

  it('displays loading state', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      user: null,
      isLoading: true,
      login: vi.fn(),
      logout: vi.fn(),
    })
    const { container } = render(<Home />, { wrapper: createQueryWrapper() })
    expect(container.querySelector('.spinner')).toBeInTheDocument()
  })
})
