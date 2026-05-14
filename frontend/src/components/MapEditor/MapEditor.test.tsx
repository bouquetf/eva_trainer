import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, waitFor, act } from '@testing-library/react'
import { MapEditor } from './MapEditor'
import { releaseLock, releaseLockSync, getLockStatus } from '../../api/locks'
import { listZones } from '../../api/zones'
import { listTeamMaps } from '../../api/maps'
import { useAuthStore } from '../../store/auth'
import { BrowserRouter } from 'react-router-dom'

vi.mock('../../api/locks', () => ({
  acquireLock: vi.fn(),
  releaseLock: vi.fn().mockResolvedValue(undefined),
  refreshLock: vi.fn(),
  releaseLockSync: vi.fn(),
  getLockStatus: vi.fn(),
}))

vi.mock('../../api/zones', () => ({
  listZones: vi.fn(),
  batchActions: vi.fn(),
}))

vi.mock('../../api/maps', () => ({
  getMapImageUrl: vi.fn().mockReturnValue(''),
  listTeamMaps: vi.fn(),
}))

beforeEach(() => {
  vi.clearAllMocks()
  useAuthStore.setState({
    user: { id: 1, discord_id: 'test', discord_username: 'tester', display_name: 'Test', is_admin: false, created_at: '2024-01-01' },
    accessToken: 'test-token',
    refreshToken: 'test-refresh',
    isAuthenticated: true,
    isLoading: false,
    error: null,
  })
})

describe('MapEditor lock integration', () => {
  it('charge les zones et les maps au montage', async () => {
    vi.mocked(listZones).mockResolvedValue({ zones: [] })
    vi.mocked(listTeamMaps).mockResolvedValue({ maps: [] })

    render(
      <BrowserRouter>
        <MapEditor teamId={1} mapId={1} />
      </BrowserRouter>,
    )

    await waitFor(() => {
      expect(listZones).toHaveBeenCalledWith(1, 1)
    }, { timeout: 10000 })

    await waitFor(() => {
      expect(listTeamMaps).toHaveBeenCalledWith(1)
    }, { timeout: 10000 })
  })

  it('appelle releaseLock lors du démontage', async () => {
    vi.mocked(listZones).mockResolvedValue({ zones: [] })
    vi.mocked(listTeamMaps).mockResolvedValue({ maps: [] })

    const { unmount } = render(
      <BrowserRouter>
        <MapEditor teamId={1} mapId={1} />
      </BrowserRouter>,
    )

    await waitFor(() => {
      expect(listZones).toHaveBeenCalledWith(1, 1)
    }, { timeout: 10000 })

    unmount()

    await waitFor(() => {
      expect(releaseLock).toHaveBeenCalledWith(1, 1)
    }, { timeout: 10000 })
  })

  it('appelle releaseLockSync lors de beforeunload', async () => {
    vi.mocked(listZones).mockResolvedValue({ zones: [] })
    vi.mocked(listTeamMaps).mockResolvedValue({ maps: [] })

    render(
      <BrowserRouter>
        <MapEditor teamId={1} mapId={1} />
      </BrowserRouter>,
    )

    await waitFor(() => {
      expect(listZones).toHaveBeenCalledWith(1, 1)
    }, { timeout: 10000 })

    act(() => {
      window.dispatchEvent(new Event('beforeunload'))
    })

    expect(releaseLockSync).toHaveBeenCalledWith(1, 1)
  })

  it('appelle getLockStatus lors du retour de visibilité', async () => {
    vi.mocked(listZones).mockResolvedValue({ zones: [] })
    vi.mocked(listTeamMaps).mockResolvedValue({ maps: [] })

    render(
      <BrowserRouter>
        <MapEditor teamId={1} mapId={1} />
      </BrowserRouter>,
    )

    await waitFor(() => {
      expect(listZones).toHaveBeenCalledWith(1, 1)
    }, { timeout: 10000 })

    vi.mocked(getLockStatus).mockResolvedValue({
      locked: true,
      lock: { id: 1, team_id: 1, map_id: 1, editor_id: 1, editor_name: 'Test', locked_at: '', expires_at: '', remaining_seconds: 800 },
    })

    Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true })
    act(() => {
      document.dispatchEvent(new Event('visibilitychange'))
    })

    await vi.waitFor(() => {
      expect(getLockStatus).toHaveBeenCalledWith(1, 1)
    }, { timeout: 5000 })
  })

  it('affiche une erreur si le lock a été volé au retour de visibilité', async () => {
    vi.mocked(listZones).mockResolvedValue({ zones: [] })
    vi.mocked(listTeamMaps).mockResolvedValue({ maps: [] })
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})

    render(
      <BrowserRouter>
        <MapEditor teamId={1} mapId={1} />
      </BrowserRouter>,
    )

    await waitFor(() => {
      expect(listZones).toHaveBeenCalledWith(1, 1)
    }, { timeout: 10000 })

    vi.mocked(getLockStatus).mockResolvedValue({
      locked: true,
      lock: { id: 1, team_id: 1, map_id: 1, editor_id: 999, editor_name: 'Intrus', locked_at: '', expires_at: '', remaining_seconds: 800 },
    })

    Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true })
    act(() => {
      document.dispatchEvent(new Event('visibilitychange'))
    })

    await vi.waitFor(() => {
      expect(alertSpy).toHaveBeenCalledWith(expect.stringContaining('Intrus'))
    }, { timeout: 5000 })

    alertSpy.mockRestore()
  })

  it('ne fait rien si le lock nous appartient toujours au retour de visibilité', async () => {
    vi.mocked(listZones).mockResolvedValue({ zones: [] })
    vi.mocked(listTeamMaps).mockResolvedValue({ maps: [] })

    render(
      <BrowserRouter>
        <MapEditor teamId={1} mapId={1} />
      </BrowserRouter>,
    )

    await waitFor(() => {
      expect(listZones).toHaveBeenCalledWith(1, 1)
    }, { timeout: 10000 })

    vi.mocked(getLockStatus).mockResolvedValue({
      locked: true,
      lock: { id: 1, team_id: 1, map_id: 1, editor_id: 1, editor_name: 'Test', locked_at: '', expires_at: new Date(Date.now() + 600000).toISOString(), remaining_seconds: 600 },
    })

    Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true })
    act(() => {
      document.dispatchEvent(new Event('visibilitychange'))
    })

    await vi.waitFor(() => {
      expect(getLockStatus).toHaveBeenCalledWith(1, 1)
    }, { timeout: 5000 })
  })
})
