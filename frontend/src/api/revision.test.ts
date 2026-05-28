import { describe, it, expect, vi, beforeEach } from 'vitest'
import { fetchRevisionZones, fetchRevisionItems } from './revision'

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
        headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: body ? JSON.stringify(body) : undefined,
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

describe('fetchRevisionZones', () => {
  it('returns zones array from API response', async () => {
    const zones = [
      {
        id: 1,
        source: 'team',
        map_id: 1,
        polygons: [],
        name: 'top',
      },
      {
        id: 5,
        source: 'template',
        map_id: 1,
        polygons: [],
        name: 'secret',
      },
    ]
    mockFetch.mockResolvedValue(okResponse({ zones }))

    const result = await fetchRevisionZones(1, 1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/1/revision-zones',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result).toHaveLength(2)
    expect(result[0].name).toBe('top')
    expect(result[0].source).toBe('team')
    expect(result[1].source).toBe('template')
  })

  it('throws on invalid token', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: 'unauthorized' }),
    })

    await expect(fetchRevisionZones(1, 1)).rejects.toThrow('API Error')
  })
})

describe('fetchRevisionItems', () => {
  it('returns zones and lines from POST endpoint', async () => {
    const items = {
      zones: [{ id: 1, source: 'team', map_id: 1, polygons: [], name: 'Zone A' }],
      lines: [{ id: 10, source: 'team', map_id: 1, polygon_data: { cones: [] }, name: 'Line A' }],
    }
    mockFetch.mockResolvedValue(okResponse(items))

    const result = await fetchRevisionItems(1, 1, { zones: true, lines: true })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/1/revision-items',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ zones: true, lines: true }),
      }),
    )
    expect(result.zones).toHaveLength(1)
    expect(result.lines).toHaveLength(1)
  })

  it('respects filter with zones only', async () => {
    mockFetch.mockResolvedValue(okResponse({ zones: [{ id: 1, source: 'team', map_id: 1, polygons: [], name: 'Zone A' }], lines: [] }))

    const result = await fetchRevisionItems(1, 1, { zones: true, lines: false })

    expect(result.zones).toHaveLength(1)
    expect(result.lines).toHaveLength(0)
  })

  it('returns empty arrays when no items', async () => {
    mockFetch.mockResolvedValue(okResponse({ zones: [], lines: [] }))

    const result = await fetchRevisionItems(1, 1)

    expect(result.zones).toEqual([])
    expect(result.lines).toEqual([])
  })
})
