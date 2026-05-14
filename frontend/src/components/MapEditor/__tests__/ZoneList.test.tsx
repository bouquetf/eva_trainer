import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { Zone } from '../../../lib/api/types'
import { ZoneList } from '../ZoneList'

const mockZones: Zone[] = [
  {
    id: 1,
    team_id: 1,
    map_id: 1,
    name: 'Zone A',
    polygons: [{ id: 10, points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 10, y: 10 }, { x: 0, y: 10 }], order_index: 0 }],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: 2,
    team_id: 1,
    map_id: 1,
    name: 'Zone B',
    polygons: [
      { id: 20, points: [{ x: 5, y: 5 }, { x: 15, y: 5 }, { x: 10, y: 15 }], order_index: 0 },
      { id: 21, points: [{ x: 20, y: 20 }, { x: 30, y: 20 }, { x: 25, y: 30 }], order_index: 1 },
    ],
    created_at: '2025-01-02T00:00:00Z',
    updated_at: '2025-01-02T00:00:00Z',
  },
]

describe('ZoneList readOnly', () => {
  it('cache l\'input "Nouvelle zone" et le bouton "+" en readOnly', () => {
    render(<ZoneList zones={mockZones} editingZoneId={null} onSelectZone={vi.fn()} readOnly />)
    expect(screen.queryByPlaceholderText('Nouvelle zone...')).not.toBeInTheDocument()
    expect(screen.queryByText('+')).not.toBeInTheDocument()
  })

  it('cache les boutons "X" de suppression en readOnly', () => {
    render(<ZoneList zones={mockZones} editingZoneId={null} onSelectZone={vi.fn()} readOnly />)
    expect(screen.queryByText('X')).not.toBeInTheDocument()
  })

  it('affiche les noms de zones en readOnly', () => {
    render(<ZoneList zones={mockZones} editingZoneId={null} onSelectZone={vi.fn()} readOnly />)
    expect(screen.getByText('Zone A')).toBeInTheDocument()
    expect(screen.getByText('Zone B')).toBeInTheDocument()
  })

  it('affiche les statistiques polygones/pts en readOnly', () => {
    render(<ZoneList zones={mockZones} editingZoneId={null} onSelectZone={vi.fn()} readOnly />)
    expect(screen.getByText('1 polygone, 4 pts')).toBeInTheDocument()
    expect(screen.getByText('2 polygones, 6 pts')).toBeInTheDocument()
  })

  it('affiche "Aucune zone." quand la liste est vide en readOnly', () => {
    render(<ZoneList zones={[]} editingZoneId={null} onSelectZone={vi.fn()} readOnly />)
    expect(screen.getByText('Aucune zone.')).toBeInTheDocument()
  })

  it('appelle onSelectZone au clic sur une zone en readOnly', async () => {
    const onSelect = vi.fn()
    const user = userEvent.setup()
    render(<ZoneList zones={mockZones} editingZoneId={null} onSelectZone={onSelect} readOnly />)
    await user.click(screen.getByText('Zone A'))
    expect(onSelect).toHaveBeenCalledWith(1)
  })
})
