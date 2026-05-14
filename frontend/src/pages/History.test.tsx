import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import History from './History'

vi.mock('../api/versions', () => ({
  listMapVersions: vi.fn(),
  restoreVersion: vi.fn(),
}))

vi.mock('../api/maps', () => ({
  listTeamMaps: vi.fn(),
}))

vi.mock('../api/zones', () => ({
  listZones: vi.fn(),
}))

vi.mock('../api/roles', () => ({
  listMembers: vi.fn(),
}))

vi.mock('../api/teams', () => ({
  getTeam: vi.fn(),
}))

vi.mock('../store/auth', () => ({
  useAuthStore: Object.assign(vi.fn(), { getState: vi.fn() }),
}))

vi.mock('../components/RestoreConfirmModal', () => ({
  default: ({ open, versionDate, authorName, onConfirm, onCancel, loading }: any) =>
    open ? (
      <div data-testid="restore-modal">
        <span>{versionDate}</span>
        <span>{authorName}</span>
        <button onClick={onConfirm} disabled={loading}>
          Confirmer la restauration
        </button>
        <button onClick={onCancel}>Annuler</button>
      </div>
    ) : null,
}))

import { listMapVersions, restoreVersion } from '../api/versions'
import { listTeamMaps } from '../api/maps'
import { listZones } from '../api/zones'
import { listMembers } from '../api/roles'
import { getTeam } from '../api/teams'
import { useAuthStore } from '../store/auth'

const mockVersions = [
  {
    id: 1,
    zone_id: 1,
    zone_name: 'Zone A',
    editor_name: 'User 1',
    created_at: '2026-01-15T10:00:00Z',
    version_data: { polygons: [{ id: 1 }] },
  },
  {
    id: 2,
    zone_id: 2,
    zone_name: 'Zone B',
    editor_name: 'User 2',
    created_at: '2026-02-20T14:30:00Z',
    version_data: { polygons: [{ id: 1 }, { id: 2 }] },
  },
]

describe('History', () => {
  const mockGetState = useAuthStore.getState as unknown as ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
    mockGetState.mockReturnValue({ user: { id: 1 } })
    ;(getTeam as any).mockResolvedValue({
      members: [{ user_id: 1, role: 'captain' }],
    })
    ;(listTeamMaps as any).mockResolvedValue({
      maps: [{ id: 1, name: 'Test Map' }],
    })
    ;(listZones as any).mockResolvedValue({
      zones: [
        { id: 1, name: 'Zone A' },
        { id: 2, name: 'Zone B' },
      ],
    })
    ;(listMembers as any).mockResolvedValue({
      members: [{ user_id: 1, display_name: 'User 1', role: 'captain' }],
    })
  })

  function renderHistory() {
    return render(
      <MemoryRouter initialEntries={['/teams/1/history/1']}>
        <Routes>
          <Route path="/teams/:id/history/:mapId" element={<History />} />
        </Routes>
      </MemoryRouter>,
    )
  }

  it('shows loading state initially', () => {
    ;(listMapVersions as any).mockReturnValue(new Promise(() => {}))
    const { container } = renderHistory()
    expect(container.querySelector('.spinner')).toBeInTheDocument()
  })

  it('renders versions after loading', async () => {
    ;(listMapVersions as any).mockResolvedValue({
      versions: mockVersions,
      total: 2,
    })
    renderHistory()
    await waitFor(() => {
      expect(screen.getByText(/2 résultats/)).toBeInTheDocument()
    })
    expect(screen.getByRole('cell', { name: 'User 1' })).toBeInTheDocument()
    expect(screen.getByRole('cell', { name: 'User 2' })).toBeInTheDocument()
  })

  it('handles error state', async () => {
    ;(listMapVersions as any).mockRejectedValue(new Error('Failed'))
    renderHistory()
    await waitFor(() => {
      expect(
        screen.getByText('Erreur lors du chargement des versions.'),
      ).toBeInTheDocument()
    })
  })

  it('shows retry button on error', async () => {
    ;(listMapVersions as any).mockRejectedValue(new Error('Failed'))
    renderHistory()
    await waitFor(() => {
      expect(screen.getByText('Réessayer')).toBeInTheDocument()
    })
  })

  it('opens restore modal when restore button clicked', async () => {
    ;(listMapVersions as any).mockResolvedValue({
      versions: mockVersions,
      total: 2,
    })
    renderHistory()
    await waitFor(() => {
      expect(screen.getByText(/2 résultats/)).toBeInTheDocument()
    })
    const restoreButtons = screen.getAllByText('Restaurer')
    await userEvent.click(restoreButtons[0])
    expect(screen.getByTestId('restore-modal')).toBeInTheDocument()
  })

  it('calls restoreVersion on confirm', async () => {
    ;(restoreVersion as any).mockResolvedValue({})
    ;(listMapVersions as any).mockResolvedValue({
      versions: mockVersions,
      total: 2,
    })
    renderHistory()
    await waitFor(() => {
      expect(screen.getByText(/2 résultats/)).toBeInTheDocument()
    })
    await userEvent.click(screen.getAllByText('Restaurer')[0])
    await userEvent.click(screen.getByText('Confirmer la restauration'))
    expect(restoreVersion).toHaveBeenCalledWith(1, 1, 1)
  })
})
