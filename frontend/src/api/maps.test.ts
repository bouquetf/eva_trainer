import { describe, it, expect, vi, beforeEach } from 'vitest'
import { uploadMap, listAdminMaps, listTeamMaps, getMapDetail, purgeObsoleteMaps, getMapImageUrl } from './maps'

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

    async upload<T>(path: string, formData: FormData): Promise<T> {
      const token = 'test-token'
      const res = await fetch(`http://localhost:8080${path}`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: formData,
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
  return { ok: true, status: 200, json: () => Promise.resolve(body) }
}

beforeEach(() => {
  mockFetch.mockReset()
})

describe('maps API', () => {
  it('uploadMap sends POST with FormData', async () => {
    const response = {
      map: { id: 1, name: 'test-map', added_at: '2025-01-01T00:00:00Z', is_obsolete: false, obsolete_of: null },
      previous_version_obsoleted: null,
    }
    mockFetch.mockResolvedValue(okResponse(response))

    const file = new File([''], 'test.w3x', { type: 'application/octet-stream' })
    const result = await uploadMap('test-map', file)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/maps',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.map.name).toBe('test-map')
  })

  it('listAdminMaps calls GET /api/v1/admin/maps', async () => {
    const response = {
      maps: [{ id: 1, name: 'map1', added_at: '2025-01-01T00:00:00Z', is_obsolete: false, obsolete_of: null }],
    }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await listAdminMaps()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/maps',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.maps).toHaveLength(1)
  })

  it('listTeamMaps calls GET /api/v1/teams/:id/maps', async () => {
    const response = { maps: [] }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await listTeamMaps(5)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/5/maps',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.maps).toEqual([])
  })

  it('getMapDetail calls GET /api/v1/admin/maps/:id', async () => {
    const response = {
      map: { id: 42, name: 'detail-map', added_at: '2025-01-01T00:00:00Z', is_obsolete: false, obsolete_of: null },
    }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await getMapDetail(42)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/maps/42',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.map.id).toBe(42)
  })

  it('purgeObsoleteMaps calls POST /api/v1/admin/maps/purge', async () => {
    const response = { message: 'Purged', purged_ids: [1, 2, 3] }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await purgeObsoleteMaps()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/maps/purge',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.purged_ids).toEqual([1, 2, 3])
  })

  it('getMapImageUrl returns the image URL', () => {
    vi.stubEnv('VITE_API_URL', 'http://localhost:8080')
    const url = getMapImageUrl(7)
    expect(url).toBe('http://localhost:8080/api/v1/maps/7/image')
    vi.unstubAllEnvs()
  })

  it('throws on API error', async () => {
    mockFetch.mockResolvedValue({ ok: false, status: 500, json: () => Promise.resolve({}) })

    await expect(listAdminMaps()).rejects.toThrow('API Error')
  })
})
