import { describe, it, expect, vi, beforeEach } from 'vitest'
import { fetchRevisionZones } from './revision'

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

  it('returns empty array when no zones', async () => {
    mockFetch.mockResolvedValue(okResponse({ zones: [] }))

    const result = await fetchRevisionZones(1, 1)

    expect(result).toEqual([])
  })
})
