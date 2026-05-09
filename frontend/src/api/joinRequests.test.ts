import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  requestJoinTeam,
  listPendingJoinRequests,
  getMyJoinRequest,
  cancelMyJoinRequest,
  acceptJoinRequest,
  declineJoinRequest,
} from './joinRequests'

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

    async del<T>(path: string, body?: unknown): Promise<T> {
      const token = 'test-token'
      const res = await fetch(`http://localhost:8080${path}`, {
        method: 'DELETE',
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

describe('joinRequests API', () => {
  it('requestJoinTeam calls POST with correct URL', async () => {
    const jr = { id: 1, team_id: 1, user_id: 3, status: 'pending' as const, created_at: '' }
    mockFetch.mockResolvedValue(okResponse(jr))

    const result = await requestJoinTeam(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/join-requests',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.status).toBe('pending')
  })

  it('listPendingJoinRequests calls GET with correct URL', async () => {
    const data = { requests: [{ id: 1, team_id: 1, user_id: 3, status: 'pending' as const, created_at: '' }] }
    mockFetch.mockResolvedValue(okResponse(data))

    const result = await listPendingJoinRequests(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/join-requests',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.requests).toHaveLength(1)
  })

  it('getMyJoinRequest calls GET with /mine suffix', async () => {
    const jr = { id: 1, team_id: 1, user_id: 3, status: 'pending' as const, created_at: '' }
    mockFetch.mockResolvedValue(okResponse(jr))

    const result = await getMyJoinRequest(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/join-requests/mine',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.status).toBe('pending')
  })

  it('cancelMyJoinRequest calls DELETE with /mine suffix', async () => {
    mockFetch.mockResolvedValue(okResponse({ message: 'Demande annulée' }))

    const result = await cancelMyJoinRequest(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/join-requests/mine',
      expect.objectContaining({ method: 'DELETE' }),
    )
    expect(result.message).toBe('Demande annulée')
  })

  it('acceptJoinRequest calls PUT with correct URL', async () => {
    const data = { message: 'Demande acceptée', user_id: 3 }
    mockFetch.mockResolvedValue(okResponse(data))

    const result = await acceptJoinRequest(1, 5)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/join-requests/5/accept',
      expect.objectContaining({ method: 'PUT' }),
    )
    expect(result.user_id).toBe(3)
  })

  it('declineJoinRequest calls PUT with correct URL', async () => {
    const data = { message: 'Demande refusée', user_id: 3 }
    mockFetch.mockResolvedValue(okResponse(data))

    const result = await declineJoinRequest(1, 5)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/join-requests/5/decline',
      expect.objectContaining({ method: 'PUT' }),
    )
    expect(result.user_id).toBe(3)
  })
})
