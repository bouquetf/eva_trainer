import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AuthGuard from './AuthGuard'

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

vi.mock('../store/auth', () => ({
  useAuthStore: vi.fn(),
}))

import { useAuthStore } from '../store/auth'

describe('AuthGuard', () => {
  const mockUseAuthStore = useAuthStore as unknown as ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('redirects to /login when not authenticated', () => {
    mockUseAuthStore.mockReturnValue({ isAuthenticated: false, isLoading: false })
    render(
      <MemoryRouter initialEntries={['/editor']}>
        <AuthGuard>
          <div>Protected content</div>
        </AuthGuard>
      </MemoryRouter>,
    )
    expect(mockNavigate).toHaveBeenCalledWith(
      expect.objectContaining({ to: '/login?redirect=%2Feditor', replace: true }),
    )
  })

  it('renders children when authenticated', () => {
    mockUseAuthStore.mockReturnValue({ isAuthenticated: true, isLoading: false })
    render(
      <MemoryRouter>
        <AuthGuard>
          <div>Protected content</div>
        </AuthGuard>
      </MemoryRouter>,
    )
    expect(screen.getByText('Protected content')).toBeInTheDocument()
  })

  it('shows loading state while checking auth', () => {
    mockUseAuthStore.mockReturnValue({ isAuthenticated: false, isLoading: true })
    const { container } = render(
      <MemoryRouter>
        <AuthGuard>
          <div>Protected content</div>
        </AuthGuard>
      </MemoryRouter>,
    )
    expect(container.querySelector('.spinner')).toBeInTheDocument()
    expect(screen.queryByText('Protected content')).not.toBeInTheDocument()
  })
})
