import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AppLayout from './AppLayout'

function createMockAuthStore(overrides: { user?: any }) {
  const state = { user: overrides.user ?? null }
  return (selector: (s: typeof state) => any) => selector(state)
}

vi.mock('../store/auth', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('./AuthGuard', () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

import { useAuthStore } from '../store/auth'

describe('AppLayout', () => {
  const mockUseAuthStore = useAuthStore as unknown as ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders sidebar navigation', () => {
    mockUseAuthStore.mockImplementation(createMockAuthStore({ user: null }))
    render(
      <MemoryRouter>
        <AppLayout />
      </MemoryRouter>,
    )
    expect(screen.getByText('Accueil')).toBeInTheDocument()
    expect(screen.getByText('Mes équipes')).toBeInTheDocument()
    expect(screen.getByText('Créer une équipe')).toBeInTheDocument()
  })

  it('shows admin link when user is admin', () => {
    mockUseAuthStore.mockImplementation(createMockAuthStore({ user: { id: 1, discord_id: 'test', is_admin: true } }))
    render(
      <MemoryRouter>
        <AppLayout />
      </MemoryRouter>,
    )
    expect(screen.getByText('Administration')).toBeInTheDocument()
  })

  it('hides admin link when user is not admin', () => {
    mockUseAuthStore.mockImplementation(createMockAuthStore({ user: { id: 1, discord_id: 'test', is_admin: false } }))
    render(
      <MemoryRouter>
        <AppLayout />
      </MemoryRouter>,
    )
    expect(screen.queryByText('Administration')).not.toBeInTheDocument()
  })

  it('renders settings link', () => {
    mockUseAuthStore.mockImplementation(createMockAuthStore({ user: null }))
    render(
      <MemoryRouter>
        <AppLayout />
      </MemoryRouter>,
    )
    expect(screen.getByText('Paramètres')).toBeInTheDocument()
  })
})
