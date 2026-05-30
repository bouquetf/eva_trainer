import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { createRef } from 'react'
import VersionManager, { type VersionManagerHandle } from './VersionManager'
import type { Zone, LineOfFire } from '../../lib/api/types'

vi.mock('../../api/versions', () => ({
  listZoneVersions: vi.fn().mockResolvedValue({ versions: [], total: 0, page: 1, limit: 50 }),
  listLineOfFireVersions: vi.fn().mockResolvedValue({ versions: [], total: 0 }),
  restoreVersion: vi.fn(),
}))

vi.mock('../../api/zones', () => ({
  listZones: vi.fn().mockResolvedValue({ zones: [] }),
}))

vi.mock('../../store/auth', () => ({
  useAuthStore: {
    getState: () => ({
      user: { id: 1, discord_id: 'test', discord_username: '', display_name: '', is_admin: false, created_at: '2024-01-01T00:00:00Z' },
    }),
  },
}))

const mockZone: Zone = {
  id: 1,
  team_id: 1,
  map_id: 1,
  name: 'Zone A',
  polygons: [],
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockLoF: LineOfFire = {
  id: 1,
  team_id: 1,
  map_id: 1,
  name: 'LoF A',
  polygon_data: { cones: [] },
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

describe('VersionManager', () => {
  const defaultProps = {
    teamId: 1,
    mapId: 1,
    editingZone: null,
    setEditingZone: vi.fn(),
    setZones: vi.fn(),
    setOriginalSnapshot: vi.fn(),
    linesOfFire: [],
    setLinesOfFire: vi.fn(),
    selectedLineOfFireId: null,
    dispatch: vi.fn(),
    ensureLock: vi.fn().mockResolvedValue(true),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when readOnly', () => {
    const { container } = render(<VersionManager {...defaultProps} readOnly />)
    expect(container.innerHTML).toBe('')
  })

  it('renders zone version browser when editingZone.id > 0', () => {
    render(<VersionManager {...defaultProps} editingZone={mockZone} />)
    expect(screen.getByText('Versions de la zone')).toBeInTheDocument()
  })

  it('renders LoF version browser when selectedLineOfFireId is set', () => {
    render(
      <VersionManager
        {...defaultProps}
        selectedLineOfFireId={1}
        linesOfFire={[mockLoF]}
      />,
    )
    expect(screen.getByText('Versions de la ligne de tir')).toBeInTheDocument()
  })

  it('renders both LoF and zone version browsers when both are available', () => {
    render(
      <VersionManager
        {...defaultProps}
        selectedLineOfFireId={1}
        linesOfFire={[mockLoF]}
        editingZone={mockZone}
      />,
    )
    expect(screen.getByText('Versions de la ligne de tir')).toBeInTheDocument()
    expect(screen.getByText('Versions de la zone')).toBeInTheDocument()
  })

  it('does not render zone version browser when zone.id is 0', () => {
    render(
      <VersionManager
        {...defaultProps}
        editingZone={{ ...mockZone, id: 0 }}
      />,
    )
    expect(screen.queryByText('Versions de la zone')).not.toBeInTheDocument()
  })

  it('exposes clearBrowsing via ref', () => {
    const ref = createRef<VersionManagerHandle>()
    render(<VersionManager {...defaultProps} ref={ref} />)
    expect(ref.current).not.toBeNull()
    expect(typeof ref.current!.clearBrowsing).toBe('function')
  })

  it('calls onBrowsingChange when browsing state changes', () => {
    const onBrowsingChange = vi.fn()
    render(
      <VersionManager
        {...defaultProps}
        editingZone={mockZone}
        onBrowsingChange={onBrowsingChange}
      />,
    )
    expect(onBrowsingChange).toHaveBeenCalledWith(false, false)
  })
})
