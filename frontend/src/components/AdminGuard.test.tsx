import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AdminGuard from './AdminGuard'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    Navigate: (props: any) => {
      mockNavigate(props)
      return null
    },
  }
})

function createMockAuthStore(overrides: { user?: any; isLoading?: boolean }) {
  const state = { user: overrides.user ?? null, isLoading: overrides.isLoading ?? false }
  return (selector: (s: typeof state) => any) => selector(state)
}

vi.mock('../store/auth', () => ({
  useAuthStore: vi.fn(),
}))

import { useAuthStore } from '../store/auth'

describe('AdminGuard', () => {
  const mockUseAuthStore = useAuthStore as unknown as ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('redirects to / when not authenticated', () => {
    mockUseAuthStore.mockImplementation(createMockAuthStore({ user: null }))
    render(
      <MemoryRouter>
        <AdminGuard>
          <div>Admin content</div>
        </AdminGuard>
      </MemoryRouter>,
    )
    expect(mockNavigate).toHaveBeenCalledWith(
      expect.objectContaining({ to: '/', replace: true }),
    )
  })

  it('redirects to / when not admin', () => {
    mockUseAuthStore.mockImplementation(createMockAuthStore({ user: { id: 1, discord_id: 'test', is_admin: false } }))
    render(
      <MemoryRouter>
        <AdminGuard>
          <div>Admin content</div>
        </AdminGuard>
      </MemoryRouter>,
    )
    expect(mockNavigate).toHaveBeenCalledWith(
      expect.objectContaining({ to: '/', replace: true }),
    )
  })

  it('renders children when admin', () => {
    mockUseAuthStore.mockImplementation(createMockAuthStore({ user: { id: 1, discord_id: 'test', is_admin: true } }))
    render(
      <MemoryRouter>
        <AdminGuard>
          <div>Admin content</div>
        </AdminGuard>
      </MemoryRouter>,
    )
    expect(screen.getByText('Admin content')).toBeInTheDocument()
  })

  it('shows loading state while checking auth', () => {
    mockUseAuthStore.mockImplementation(createMockAuthStore({ user: null, isLoading: true }))
    const { container } = render(
      <MemoryRouter>
        <AdminGuard>
          <div>Admin content</div>
        </AdminGuard>
      </MemoryRouter>,
    )
    expect(container.querySelector('.spinner')).toBeInTheDocument()
    expect(screen.queryByText('Admin content')).not.toBeInTheDocument()
  })
})
