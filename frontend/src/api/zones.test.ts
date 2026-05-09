import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listZones, createZone, deleteZone, addPolygon, removePolygon, addPolygonPoint, removePolygonPoint, setZoneName } from './zones'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

function mockApiClient() {
  class MockApiClient {
    async get<T>(path: string): Promise<T> {
      const token = 'test-token'
      const res = await fetch(`http://localhost:8080${path}`, {
        method: 'GET',
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) throw new Error('API Error')
      return res.json()
    }

    async post<T>(path: string, body?: unknown): Promise<T> {
      const token = 'test-token'
      const res = await fetch(`http://localhost:8080${path}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: body ? JSON.stringify(body) : undefined,
      })
      if (!res.ok) throw new Error('API Error')
      return res.json()
    }

    async put<T>(path: string, body?: unknown): Promise<T> {
      const token = 'test-token'
      const res = await fetch(`http://localhost:8080${path}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: body ? JSON.stringify(body) : undefined,
      })
      if (!res.ok) throw new Error('API Error')
      return res.json()
    }

    async del<T>(path: string, _body?: unknown): Promise<T> {
      const token = 'test-token'
      const res = await fetch(`http://localhost:8080${path}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) throw new Error('API Error')
      return res.json()
    }
  }

  return new MockApiClient()
}

vi.mock('./client', () => ({
  getApiClient: () => mockApiClient(),
}))

function okResponse(body: unknown) {
  return {
    ok: true,
    status: 200,
    json: () => Promise.resolve(body),
  }
}

beforeEach(() => {
  mockFetch.mockReset()
})

describe('zones API', () => {
  it('listZones calls GET /api/v1/teams/:id/maps/:mapId/zones', async () => {
    const zones = [
      { id: 1, team_id: 1, map_id: 1, polygons: [], name: 'Zone A', created_at: '', updated_at: '' },
    ]
    mockFetch.mockResolvedValue(okResponse({ zones }))

    const result = await listZones(1, 1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/1/zones',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.zones).toHaveLength(1)
    expect(result.zones[0].name).toBe('Zone A')
  })

  it('createZone calls POST /api/v1/teams/:id/maps/:mapId/zones with payload', async () => {
    const payload = { polygons: [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }], name: 'New Zone' }
    mockFetch.mockResolvedValue(okResponse({ zone: { id: 1, team_id: 1, map_id: 1, polygons: [{ id: 1, zone_id: 1, points: payload.polygons[0].points, order_index: 0, created_at: '' }], name: 'New Zone', created_at: '', updated_at: '' } }))

    const result = await createZone(1, 1, payload)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/1/zones',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify(payload),
      }),
    )
    expect(result.zone.name).toBe('New Zone')
  })

  it('deleteZone calls DELETE /api/v1/teams/:id/zones/:zoneId', async () => {
    mockFetch.mockResolvedValue(okResponse({}))

    await deleteZone(1, 5)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/zones/5',
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('addPolygon calls POST /api/v1/teams/:id/zones/:zoneId/polygons', async () => {
    const points = [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }]
    mockFetch.mockResolvedValue(okResponse({ polygon: { id: 1, zone_id: 1, points, order_index: 1, created_at: '' } }))

    const result = await addPolygon(1, 1, { points })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/zones/1/polygons',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ points }),
      }),
    )
    expect(result.polygon.id).toBe(1)
  })

  it('removePolygon calls DELETE /api/v1/teams/:id/zones/:zoneId/polygons/:polygonId', async () => {
    mockFetch.mockResolvedValue(okResponse({ polygon_deleted: true, zone_deleted: false }))

    const result = await removePolygon(1, 1, 2)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/zones/1/polygons/2',
      expect.objectContaining({ method: 'DELETE' }),
    )
    expect(result.polygon_deleted).toBe(true)
  })

  it('setZoneName calls PUT /api/v1/teams/:id/zones/:zoneId/name', async () => {
    mockFetch.mockResolvedValue(okResponse({ name: { id: 1, zone_id: 1, name: 'updated' } }))

    const result = await setZoneName(1, 1, { name: 'updated' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/zones/1/name',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ name: 'updated' }),
      }),
    )
    expect(result.name.name).toBe('updated')
  })

  it('addPolygonPoint calls POST /api/v1/teams/:id/zones/:zoneId/polygons/:polygonId/points', async () => {
    const points = [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }, { x: 150, y: 150 }]
    mockFetch.mockResolvedValue(okResponse({ polygon: { id: 1, zone_id: 1, points, order_index: 0, created_at: '' } }))

    const result = await addPolygonPoint(1, 1, 1, { point: { x: 150, y: 150 }, index_after: 0 })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/zones/1/polygons/1/points',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ point: { x: 150, y: 150 }, index_after: 0 }),
      }),
    )
    expect(result.polygon.points).toHaveLength(4)
  })

  it('removePolygonPoint calls DELETE /api/v1/teams/:id/zones/:zoneId/polygons/:polygonId/points/:pointIndex', async () => {
    mockFetch.mockResolvedValue(okResponse({ polygon: { id: 1, zone_id: 1, points: [{ x: 0, y: 0 }, { x: 10, y: 0 }], order_index: 0, created_at: '' } }))

    await removePolygonPoint(1, 1, 1, 0)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/zones/1/polygons/1/points/0',
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('handles API error properly', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 409,
      json: () => Promise.resolve({ error: 'zone_name_taken', message: 'zone name already exists' }),
    })

    await expect(createZone(1, 1, { polygons: [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }], name: 'dup' })).rejects.toThrow('API Error')
  })
})
