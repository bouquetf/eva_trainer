import { describe, it, expect, vi, beforeEach } from 'vitest'
import { acquireLock, releaseLock, refreshLock, getLockStatus } from './locks'

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

describe('locks API', () => {
  it('acquireLock calls POST /api/v1/teams/:id/maps/:mapId/lock and returns the lock', async () => {
    const lockData = {
      id: 1,
      team_id: 1,
      map_id: 1,
      editor_id: 1,
      locked_at: '2025-01-01T00:00:00Z',
      expires_at: '2025-01-01T00:15:00Z',
    }
    mockFetch.mockResolvedValue(okResponse({ lock: lockData }))

    const result = await acquireLock(1, 1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/1/lock',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.lock).toEqual(lockData)
  })

  it('releaseLock calls DELETE /api/v1/teams/:id/maps/:mapId/lock', async () => {
    mockFetch.mockResolvedValue(okResponse({}))

    await releaseLock(1, 1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/1/lock',
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('refreshLock calls PUT /api/v1/teams/:id/maps/:mapId/lock and returns refreshed lock', async () => {
    const refreshedLock = {
      id: 1,
      team_id: 1,
      map_id: 1,
      editor_id: 1,
      locked_at: '2025-01-01T00:00:00Z',
      expires_at: '2025-01-01T00:30:00Z',
    }
    mockFetch.mockResolvedValue(okResponse({ lock: refreshedLock }))

    const result = await refreshLock(1, 1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/1/lock',
      expect.objectContaining({ method: 'PUT' }),
    )
    expect(result.lock.expires_at).toBe('2025-01-01T00:30:00Z')
  })

  it('getLockStatus calls GET /api/v1/teams/:id/maps/:mapId/lock and returns status', async () => {
    mockFetch.mockResolvedValue(okResponse({ locked: true, lock: { id: 1, team_id: 1, map_id: 1, editor_id: 1, locked_at: '', expires_at: '' } }))

    const result = await getLockStatus(1, 1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/1/lock',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.locked).toBe(true)
  })

  it('getLockStatus returns locked false when no lock', async () => {
    mockFetch.mockResolvedValue(okResponse({ locked: false, lock: null }))

    const result = await getLockStatus(1, 1)

    expect(result.locked).toBe(false)
    expect(result.lock).toBeNull()
  })
})
