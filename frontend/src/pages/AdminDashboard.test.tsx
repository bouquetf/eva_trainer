import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import AdminDashboard from './AdminDashboard'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useNavigate: () => mockNavigate }
})

vi.mock('../store/auth', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('../api/admin', () => ({
  listAdminUsers: vi.fn(),
  listAdminPermissions: vi.fn(),
  getUserDetail: vi.fn(),
  getRolePermissions: vi.fn(),
  listAdminTeams: vi.fn(),
  promoteUser: vi.fn(),
  updateUserRole: vi.fn(),
  addUserToTeam: vi.fn(),
  removeUserFromTeam: vi.fn(),
  listArchivists: vi.fn(),
  addArchivist: vi.fn(),
  removeArchivist: vi.fn(),
  listReaderTeams: vi.fn(),
  addReaderTeam: vi.fn(),
  removeReaderTeam: vi.fn(),
}))

vi.mock('../api/maps', () => ({
  uploadMap: vi.fn(),
  listAdminMaps: vi.fn(),
  purgeObsoleteMaps: vi.fn(),
  getMapImageUrl: vi.fn(() => '/maps/1/image'),
}))

vi.mock('../api/splits', () => ({
  createSplit: vi.fn(),
  listSplits: vi.fn(),
  getSplitMaps: vi.fn(),
  updateSplit: vi.fn(),
  deleteSplit: vi.fn(),
}))

vi.mock('../api/map-catalog', () => ({
  listCatalog: vi.fn(),
  createCatalogEntry: vi.fn(),
  updateCatalogEntry: vi.fn(),
  deleteCatalogEntry: vi.fn(),
}))

import { useAuthStore } from '../store/auth'
import { listAdminUsers, listAdminPermissions, listAdminTeams, listArchivists, listReaderTeams } from '../api/admin'
import { listAdminMaps } from '../api/maps'
import { listSplits } from '../api/splits'
import { listCatalog } from '../api/map-catalog'

const mockUseAuthStore = useAuthStore as unknown as ReturnType<typeof vi.fn>

function createQueryWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/admin']}>{children}</MemoryRouter>
      </QueryClientProvider>
    )
  }
}

describe('AdminDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, discord_id: 'admin', is_admin: true },
    })
    ;(listAdminUsers as any).mockResolvedValue({
      users: [
        { id: 1, discord_username: 'admin', display_name: 'Admin', is_admin: true },
      ],
    })
    ;(listAdminPermissions as any).mockResolvedValue({
      permissions: [],
      roles: [],
    })
    ;(listAdminMaps as any).mockResolvedValue({ maps: [] })
    ;(listSplits as any).mockResolvedValue({ splits: [] })
    ;(listCatalog as any).mockResolvedValue({ entries: [] })
    ;(listAdminTeams as any).mockResolvedValue({ teams: [] })
    ;(listArchivists as any).mockResolvedValue({ archivists: [] })
    ;(listReaderTeams as any).mockResolvedValue({ teams: [] })
  })

  it('renders admin dashboard with tabs', async () => {
    render(<AdminDashboard />, { wrapper: createQueryWrapper() })
    expect(screen.getByText('Administration')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText('Utilisateurs (1)')).toBeInTheDocument()
      expect(screen.getByText('Permissions')).toBeInTheDocument()
      expect(screen.getByText('Cartes (0)')).toBeInTheDocument()
      expect(screen.getByText('Splits (0)')).toBeInTheDocument()
      expect(screen.getByText('Catalogue (0)')).toBeInTheDocument()
    })
  })

  it('renders users tab by default', async () => {
    render(<AdminDashboard />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText('Tous les utilisateurs')).toBeInTheDocument()
      expect(screen.getByText('admin')).toBeInTheDocument()
    })
  })

  it('renders without crashing when user is null', async () => {
    mockUseAuthStore.mockReturnValue({ user: null })
    ;(listAdminUsers as any).mockResolvedValue({
      users: [
        { id: 2, discord_username: 'other', display_name: 'Other', is_admin: false },
      ],
    })
    render(<AdminDashboard />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText('Utilisateurs (1)')).toBeInTheDocument()
      expect(screen.getByText('other')).toBeInTheDocument()
    })
  })
})
