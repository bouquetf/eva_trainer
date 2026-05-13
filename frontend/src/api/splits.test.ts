import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createSplit, listSplits, getSplit, updateSplit, deleteSplit, getSplitMaps } from './splits'

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
  return { ok: true, status: 200, json: () => Promise.resolve(body) }
}

beforeEach(() => {
  mockFetch.mockReset()
})

describe('splits API', () => {
  it('listSplits calls GET /api/v1/admin/splits', async () => {
    const response = { splits: [{ id: 1, name: 'Split A', map_names: ['map1.w3x'] }] }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await listSplits()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/splits',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.splits).toHaveLength(1)
  })

  it('createSplit calls POST /api/v1/admin/splits', async () => {
    const response = { split: { id: 1, name: 'New Split', map_names: ['map1.w3x', 'map2.w3x'] } }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await createSplit({ name: 'New Split', map_names: ['map1.w3x', 'map2.w3x'] })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/splits',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'New Split', map_names: ['map1.w3x', 'map2.w3x'] }),
      }),
    )
    expect(result.split.name).toBe('New Split')
  })

  it('getSplit calls GET /api/v1/admin/splits/:id', async () => {
    const response = { split: { id: 3, name: 'Split C', map_names: ['map3.w3x'] } }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await getSplit(3)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/splits/3',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.split.id).toBe(3)
  })

  it('updateSplit calls PUT /api/v1/admin/splits/:id', async () => {
    const response = { split: { id: 1, name: 'Updated Split', map_names: ['map1.w3x'] } }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await updateSplit(1, { name: 'Updated Split', map_names: ['map1.w3x'] })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/splits/1',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ name: 'Updated Split', map_names: ['map1.w3x'] }),
      }),
    )
    expect(result.split.name).toBe('Updated Split')
  })

  it('deleteSplit calls DELETE /api/v1/admin/splits/:id', async () => {
    const response = { message: 'Split deleted' }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await deleteSplit(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/splits/1',
      expect.objectContaining({ method: 'DELETE' }),
    )
    expect(result.message).toBe('Split deleted')
  })

  it('getSplitMaps calls GET /api/v1/admin/splits/:id/maps', async () => {
    const response = { maps: [{ id: 10, name: 'map1.w3x', added_at: '', is_obsolete: false, obsolete_of: null }] }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await getSplitMaps(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/splits/1/maps',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.maps).toHaveLength(1)
  })

  it('throws on API error', async () => {
    mockFetch.mockResolvedValue({ ok: false, status: 500, json: () => Promise.resolve({}) })

    await expect(listSplits()).rejects.toThrow('API Error')
  })
})
