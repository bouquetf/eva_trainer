import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listCatalog, createCatalogEntry, updateCatalogEntry, deleteCatalogEntry } from './map-catalog'

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

describe('map catalog API', () => {
  it('listCatalog calls GET /api/v1/admin/map-catalog', async () => {
    const response = { entries: [{ id: 1, name: 'Entry A', created_at: '2025-01-01T00:00:00Z' }] }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await listCatalog()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/map-catalog',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.entries).toHaveLength(1)
  })

  it('createCatalogEntry calls POST /api/v1/admin/map-catalog', async () => {
    const response = { entry: { id: 1, name: 'New Entry', created_at: '2025-01-01T00:00:00Z' } }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await createCatalogEntry('New Entry')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/map-catalog',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'New Entry' }),
      }),
    )
    expect(result.entry.name).toBe('New Entry')
  })

  it('updateCatalogEntry calls PUT /api/v1/admin/map-catalog/:id', async () => {
    const response = { entry: { id: 1, name: 'Updated Entry', created_at: '2025-01-01T00:00:00Z' } }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await updateCatalogEntry(1, 'Updated Entry')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/map-catalog/1',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ name: 'Updated Entry' }),
      }),
    )
    expect(result.entry.name).toBe('Updated Entry')
  })

  it('deleteCatalogEntry calls DELETE /api/v1/admin/map-catalog/:id', async () => {
    const response = { message: 'Deleted' }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await deleteCatalogEntry(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/map-catalog/1',
      expect.objectContaining({ method: 'DELETE' }),
    )
    expect(result.message).toBe('Deleted')
  })

  it('throws on API error', async () => {
    mockFetch.mockResolvedValue({ ok: false, status: 500, json: () => Promise.resolve({}) })

    await expect(listCatalog()).rejects.toThrow('API Error')
  })
})
