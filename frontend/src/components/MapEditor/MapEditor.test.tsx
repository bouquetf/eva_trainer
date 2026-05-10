import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, waitFor, act } from '@testing-library/react'
import { MapEditor } from './MapEditor'
import { releaseLock, releaseLockSync } from '../../api/locks'
import { listZones } from '../../api/zones'
import { listTeamMaps } from '../../api/maps'
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
})
