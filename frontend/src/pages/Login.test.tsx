import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import Login from './Login'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('../store/auth', () => ({
  useAuthStore: vi.fn(),
}))

import { useAuthStore } from '../store/auth'

describe('Login', () => {
  const mockUseAuthStore = useAuthStore as unknown as ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders Discord login button', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login: vi.fn(),
    })
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>,
    )
    expect(screen.getByText('Connexion avec Discord')).toBeInTheDocument()
  })

  it('calls login on button click', async () => {
    const login = vi.fn()
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login,
    })
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>,
    )
    await userEvent.click(screen.getByText('Connexion avec Discord'))
    expect(login).toHaveBeenCalledTimes(1)
  })

  it('shows redirecting state after click', async () => {
    const login = vi.fn()
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login,
    })
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>,
    )
    await userEvent.click(screen.getByText('Connexion avec Discord'))
    expect(screen.getByText('Redirection...')).toBeInTheDocument()
  })

  it('shows loading spinner when auth is loading', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      isLoading: true,
      login: vi.fn(),
    })
    const { container } = render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>,
    )
    expect(container.querySelector('.spinner')).toBeInTheDocument()
  })

  it('shows error when error param is present', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login: vi.fn(),
    })
    render(
      <MemoryRouter initialEntries={['/login?error=access_denied']}>
        <Login />
      </MemoryRouter>,
    )
    expect(screen.getByText('Connexion annulée')).toBeInTheDocument()
  })

  it('shows generic error for unknown error param', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login: vi.fn(),
    })
    render(
      <MemoryRouter initialEntries={['/login?error=unknown']}>
        <Login />
      </MemoryRouter>,
    )
    expect(screen.getByText('connexion impossible')).toBeInTheDocument()
  })

  it('renders Discord SVG logo', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login: vi.fn(),
    })
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>,
    )
    const svg = document.querySelector('svg')
    expect(svg).toBeInTheDocument()
  })

  it('navigates to redirect when already authenticated', () => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
      login: vi.fn(),
    })
    render(
      <MemoryRouter initialEntries={['/login?redirect=/teams']}>
        <Login />
      </MemoryRouter>,
    )
    expect(mockNavigate).toHaveBeenCalledWith('/teams', { replace: true })
  })
})
