import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listMapVersions, listZoneVersions, restoreVersion } from './versions'

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

describe('versions API', () => {
  it('listMapVersions without filters calls GET with page=1 by default', async () => {
    mockFetch.mockResolvedValue(okResponse({
      versions: [],
      total: 0,
      page: 1,
      limit: 20,
    }))

    const result = await listMapVersions(1, 2)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/2/versions',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.versions).toEqual([])
    expect(result.total).toBe(0)
  })

  it('listMapVersions with filters includes query params', async () => {
    mockFetch.mockResolvedValue(okResponse({
      versions: [],
      total: 0,
      page: 1,
      limit: 10,
    }))

    await listMapVersions(1, 2, { zone_id: 5, page: 1, limit: 10 })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/2/versions?page=1&limit=10&zone_id=5',
      expect.objectContaining({ method: 'GET' }),
    )
  })

  it('listZoneVersions calls the correct endpoint with zoneId', async () => {
    mockFetch.mockResolvedValue(okResponse({
      versions: [
        {
          id: 42,
          zone_id: 5,
          zone_name: 'top',
          editor_id: 3,
          editor_name: 'Player',
          version_data: { polygons: [], name: 'top' },
          created_at: '2026-05-10T12:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      limit: 20,
    }))

    const result = await listZoneVersions(1, 5)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/zones/5/versions',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.versions).toHaveLength(1)
    expect(result.versions[0].zone_name).toBe('top')
  })

  it('listZoneVersions with page and limit includes query params', async () => {
    mockFetch.mockResolvedValue(okResponse({
      versions: [],
      total: 0,
      page: 2,
      limit: 10,
    }))

    await listZoneVersions(1, 5, 2, 10)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/zones/5/versions?page=2&limit=10',
      expect.objectContaining({ method: 'GET' }),
    )
  })

  it('restoreVersion calls POST with the correct versionId', async () => {
    const restoredZone = {
      id: 5,
      team_id: 1,
      map_id: 2,
      polygons: [],
      name: 'top',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-05-10T12:00:00Z',
    }
    mockFetch.mockResolvedValue(okResponse({ zone: restoredZone }))

    const result = await restoreVersion(1, 5, 42)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/zones/5/versions/42/restore',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.zone.id).toBe(5)
    expect(result.zone.name).toBe('top')
  })

  it('listMapVersions handles editor_id filter', async () => {
    mockFetch.mockResolvedValue(okResponse({
      versions: [],
      total: 0,
      page: 1,
      limit: 20,
    }))

    await listMapVersions(1, 2, { editor_id: 3 })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/2/versions?editor_id=3',
      expect.objectContaining({ method: 'GET' }),
    )
  })

  it('handles API 403 error', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 403,
      json: () => Promise.resolve({ error: 'not_version_author', message: 'you are not the author of this version' }),
    })

    await expect(restoreVersion(1, 5, 42)).rejects.toThrow('API Error')
  })
})
