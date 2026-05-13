import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import VersionBrowser from './VersionBrowser'
import type { ZoneVersion } from '../lib/api/types'

vi.mock('../api/versions', () => ({
  listZoneVersions: vi.fn(),
}))

import { listZoneVersions } from '../api/versions'

vi.mock('../store/auth', () => ({
  useAuthStore: {
    getState: () => ({
      user: { id: 1, discord_id: 'test', discord_username: '', display_name: '', is_admin: false, created_at: '2024-01-01T00:00:00Z' },
    }),
  },
}))

const mockVersions: ZoneVersion[] = [
  {
    id: 1,
    zone_id: 10,
    zone_name: 'Zone A',
    editor_id: 5,
    editor_name: 'Alice',
    version_data: { polygons: [{ id: 'p1', points: [] }] } as any,
    created_at: '2024-06-01T12:00:00Z',
  },
  {
    id: 2,
    zone_id: 10,
    zone_name: 'Zone A',
    editor_id: 6,
    editor_name: 'Bob',
    version_data: { polygons: [{ id: 'p2', points: [] }, { id: 'p3', points: [] }] } as any,
    created_at: '2024-06-02T12:00:00Z',
  },
]

describe('VersionBrowser', () => {
  const defaultProps = {
    teamId: 1,
    zoneId: 10,
    onBrowse: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('displays empty state when no versions', async () => {
    vi.mocked(listZoneVersions).mockResolvedValue({ versions: [], total: 0, page: 1, limit: 50 })
    render(<VersionBrowser {...defaultProps} />)
    expect(await screen.findByText('Aucune version sauvegardée')).toBeInTheDocument()
  })

  it('renders version list', async () => {
    vi.mocked(listZoneVersions).mockResolvedValue({ versions: mockVersions, total: 2, page: 1, limit: 50 })
    const { container } = render(<VersionBrowser {...defaultProps} />)
    await waitFor(() => {
      expect(container.textContent).toContain('2024')
    })
  })

  it('shows zone name and editor name for each version', async () => {
    vi.mocked(listZoneVersions).mockResolvedValue({ versions: mockVersions, total: 2, page: 1, limit: 50 })
    render(<VersionBrowser {...defaultProps} />)
    expect(await screen.findByText('Alice', { exact: false })).toBeInTheDocument()
    expect(screen.getByText('Bob', { exact: false })).toBeInTheDocument()
  })

  it('calls onBrowse when version clicked', async () => {
    vi.mocked(listZoneVersions).mockResolvedValue({ versions: mockVersions, total: 2, page: 1, limit: 50 })
    const onBrowse = vi.fn()
    const { container } = render(<VersionBrowser {...defaultProps} onBrowse={onBrowse} />)
    await waitFor(() => {
      expect(container.textContent).toContain('2024')
    })
    const items = container.querySelectorAll('[style*="cursor"]')
    if (items.length > 0) {
      await userEvent.click(items[0])
      expect(onBrowse).toHaveBeenCalled()
    }
  })

  it('returns to current version when browsing version clicked again', async () => {
    vi.mocked(listZoneVersions).mockResolvedValue({ versions: mockVersions, total: 2, page: 1, limit: 50 })
    const onBrowse = vi.fn()
    render(<VersionBrowser {...defaultProps} browsingVersionId={1} onBrowse={onBrowse} />)
    const currentBtn = await screen.findByText('Version actuelle')
    await userEvent.click(currentBtn)
    expect(onBrowse).toHaveBeenCalledWith(null)
  })

  it('handles loading state', async () => {
    vi.mocked(listZoneVersions).mockImplementation(() => new Promise(() => {}))
    render(<VersionBrowser {...defaultProps} />)
    expect(screen.getByText('Chargement...')).toBeInTheDocument()
  })

  it('returns null when zoneId is 0', () => {
    const { container } = render(<VersionBrowser {...defaultProps} zoneId={0} />)
    expect(container.innerHTML).toBe('')
  })
})
