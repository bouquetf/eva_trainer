import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Editor from './Editor'

vi.mock('../components/MapEditor/MapEditor', () => ({
  MapEditor: ({ teamId, mapId }: any) => (
    <div data-testid="map-editor">MapEditor teamId={teamId} mapId={mapId}</div>
  ),
}))

vi.mock('../components/AuthGuard', () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useParams: vi.fn(), useNavigate: vi.fn() }
})

vi.mock('../store/auth', () => {
  const mockUser = { id: 1, discord_id: 'test', discord_username: '', display_name: '', is_admin: false, created_at: '2024-01-01T00:00:00Z' }
  const state = { user: mockUser }
  const useAuthStore = vi.fn((selector?: (s: typeof state) => unknown) =>
    selector ? selector(state) : state,
  )
  useAuthStore.getState = () => state
  return { useAuthStore }
})

vi.mock('../api/teams', () => ({
  getTeam: vi.fn().mockResolvedValue({
    team: { id: 42, name: 'Test', captain_id: 1 },
    members: [{ user_id: 1, role: 'editor', display_name: 'Test', joined_at: '2024-01-01T00:00:00Z' }],
  }),
}))

import { useParams } from 'react-router-dom'

const mockUseParams = useParams as unknown as ReturnType<typeof vi.fn>

function createWrapper() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={qc}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  )
}

describe('Editor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders error when no params', () => {
    mockUseParams.mockReturnValue({ id: undefined, mapId: undefined })
    render(<Editor />, { wrapper: createWrapper() })
    expect(screen.getByText(/Paramètres d'URL invalides/)).toBeInTheDocument()
  })

  it('renders MapEditor with correct teamId and mapId', async () => {
    mockUseParams.mockReturnValue({ id: '42', mapId: '7' })
    render(<Editor />, { wrapper: createWrapper() })
    expect(await screen.findByTestId('map-editor')).toHaveTextContent('MapEditor teamId=42 mapId=7')
  })

  it('wraps in AuthGuard', async () => {
    mockUseParams.mockReturnValue({ id: '42', mapId: '7' })
    render(<Editor />, { wrapper: createWrapper() })
    expect(await screen.findByTestId('map-editor')).toBeInTheDocument()
  })
})
