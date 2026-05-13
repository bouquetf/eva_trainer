import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import Callback from './Callback'

const mockHandleCallback = vi.fn()
const mockLogout = vi.fn()
const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useNavigate: () => mockNavigate }
})

vi.mock('../store/auth', () => ({
  useAuthStore: vi.fn(),
}))

import { useAuthStore } from '../store/auth'

const mockUseAuthStore = useAuthStore as unknown as ReturnType<typeof vi.fn>

function createMockStore(overrides: any = {}) {
  const state = {
    handleCallback: mockHandleCallback,
    isLoading: false,
    error: null,
    logout: mockLogout,
    ...overrides,
  }
  return (selector?: any) => (selector ? selector(state) : state)
}

describe('Callback', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.location.hash = ''
  })

  it('calls handleCallback on mount', () => {
    mockUseAuthStore.mockImplementation(createMockStore({ isLoading: true }))
    render(
      <MemoryRouter>
        <Callback />
      </MemoryRouter>,
    )
    expect(mockHandleCallback).toHaveBeenCalledTimes(1)
  })

  it('shows loading state while processing', () => {
    mockUseAuthStore.mockImplementation(createMockStore({ isLoading: true }))
    const { container } = render(
      <MemoryRouter>
        <Callback />
      </MemoryRouter>,
    )
    expect(screen.getByText('Connexion en cours...')).toBeInTheDocument()
    expect(container.querySelector('.spinner')).toBeInTheDocument()
  })

  it('navigates to home on success', async () => {
    mockUseAuthStore.mockImplementation(createMockStore())
    render(
      <MemoryRouter>
        <Callback />
      </MemoryRouter>,
    )
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/', { replace: true })
    })
  })

  it('navigates to redirect from hash on success', async () => {
    window.location.hash = '#redirect=%2Fteams%2F1'
    mockUseAuthStore.mockImplementation(createMockStore())
    render(
      <MemoryRouter>
        <Callback />
      </MemoryRouter>,
    )
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/teams/1', { replace: true })
    })
  })

  it('shows error message on error', () => {
    mockUseAuthStore.mockImplementation(createMockStore({ error: 'Authorization failed' }))
    render(
      <MemoryRouter>
        <Callback />
      </MemoryRouter>,
    )
    expect(screen.getByText('Erreur de connexion')).toBeInTheDocument()
    expect(screen.getByText('Authorization failed')).toBeInTheDocument()
  })

  it('calls logout when error logout button clicked', async () => {
    mockUseAuthStore.mockImplementation(createMockStore({ error: 'Authorization failed' }))
    render(
      <MemoryRouter>
        <Callback />
      </MemoryRouter>,
    )
    await userEvent.click(screen.getByText('Se déconnecter'))
    expect(mockLogout).toHaveBeenCalledTimes(1)
  })
})
