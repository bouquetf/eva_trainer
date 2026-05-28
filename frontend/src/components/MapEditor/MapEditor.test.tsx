import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, waitFor, act, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MapEditor } from './MapEditor'
import { acquireLock, releaseLock, releaseLockSync, getLockStatus } from '../../api/locks'
import { listZones } from '../../api/zones'
import { listTeamMaps } from '../../api/maps'
import { useAuthStore } from '../../store/auth'
import { BrowserRouter } from 'react-router-dom'

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
})

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
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <MapEditor teamId={1} mapId={1} />
        </BrowserRouter>
      </QueryClientProvider>,
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
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <MapEditor teamId={1} mapId={1} />
        </BrowserRouter>
      </QueryClientProvider>,
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
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <MapEditor teamId={1} mapId={1} />
        </BrowserRouter>
      </QueryClientProvider>,
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
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <MapEditor teamId={1} mapId={1} />
        </BrowserRouter>
      </QueryClientProvider>,
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
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <MapEditor teamId={1} mapId={1} />
        </BrowserRouter>
      </QueryClientProvider>,
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
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <MapEditor teamId={1} mapId={1} />
        </BrowserRouter>
      </QueryClientProvider>,
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

describe('MapEditor copy/paste', () => {
  beforeEach(() => {
    class MockImage {
      onload: (() => void) | null = null;
      onerror: (() => void) | null = null;
      naturalWidth = 2000;
      naturalHeight = 1500;
      set src(_val: string) {
        Promise.resolve().then(() => this.onload?.());
      }
    }
    vi.stubGlobal('Image', MockImage);
  })

  afterEach(() => {
    vi.unstubAllGlobals();
  })

  it('should handle Ctrl+C and Ctrl+V keyboard shortcuts without error', { timeout: 20000 }, async () => {
    const zone = {
      id: 1,
      team_id: 1,
      map_id: 1,
      polygons: [{
        id: 1,
        zone_id: 1,
        points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }],
        order_index: 0,
        created_at: '2024-01-01T00:00:00Z',
      }],
      name: 'Test Zone',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    vi.mocked(listZones).mockResolvedValue({ zones: [zone] })
    vi.mocked(listTeamMaps).mockResolvedValue({ maps: [] })
    vi.mocked(acquireLock).mockResolvedValue({
      lock: { id: 1, team_id: 1, map_id: 1, editor_id: 1, editor_name: 'Test', locked_at: '', expires_at: new Date(Date.now() + 600000).toISOString() },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <MapEditor teamId={1} mapId={1} />
        </BrowserRouter>
      </QueryClientProvider>,
    )

    await waitFor(() => {
      expect(screen.getByText('Test Zone')).toBeInTheDocument()
    }, { timeout: 15000 })

    // Click zone to start editing (selects polygon)
    await act(async () => {
      screen.getByText('Test Zone').click()
    })

    // Dispatch Ctrl+C
    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', {
        key: 'c',
        ctrlKey: true,
        bubbles: true,
      }))
    })

    // Dispatch Ctrl+V
    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', {
        key: 'v',
        ctrlKey: true,
        bubbles: true,
      }))
    })

    // No crash = handlers work. Verify zone name still visible.
    expect(screen.getByText('Test Zone')).toBeInTheDocument()
  })

  it('should do nothing on Ctrl+C when no polygon is selected', { timeout: 20000 }, async () => {
    const zone = {
      id: 1,
      team_id: 1,
      map_id: 1,
      polygons: [],
      name: 'Empty Zone',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    vi.mocked(listZones).mockResolvedValue({ zones: [zone] })
    vi.mocked(listTeamMaps).mockResolvedValue({ maps: [] })

    render(
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <MapEditor teamId={1} mapId={1} />
        </BrowserRouter>
      </QueryClientProvider>,
    )

    await waitFor(() => {
      expect(screen.getByText('Empty Zone')).toBeInTheDocument()
    }, { timeout: 10000 })

    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', {
        key: 'c',
        ctrlKey: true,
        bubbles: true,
      }))
    })

    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', {
        key: 'v',
        ctrlKey: true,
        bubbles: true,
      }))
    })

    expect(screen.getByText('Empty Zone')).toBeInTheDocument()
  })
})
