import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  inviteByDisplayName,
  generateInviteLink,
  listPendingInvitations,
  listMyInvitations,
  acceptInvitation,
  declineInvitation,
  getInviteLinkInfo,
  acceptInviteLink,
  searchUsers,
} from './invitations'

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

describe('invitations API', () => {
  it('inviteByDisplayName calls POST with correct URL and body', async () => {
    const invitation = { id: 1, team_id: 1, status: 'pending' as const, expires_at: '', created_at: '' }
    mockFetch.mockResolvedValue(okResponse(invitation))

    const result = await inviteByDisplayName(1, { display_name: 'Alice' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/invitations',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ display_name: 'Alice' }),
      }),
    )
    expect(result.id).toBe(1)
  })

  it('generateInviteLink calls POST with correct URL', async () => {
    const invitation = { id: 1, team_id: 1, token: 'abc', invite_url: 'http://...', status: 'pending' as const, expires_at: '', created_at: '' }
    mockFetch.mockResolvedValue(okResponse(invitation))

    const result = await generateInviteLink(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/invitations/link',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.token).toBe('abc')
  })

  it('listPendingInvitations calls GET with correct URL', async () => {
    const data = { invitations: [] }
    mockFetch.mockResolvedValue(okResponse(data))

    const result = await listPendingInvitations(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/invitations/pending',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.invitations).toEqual([])
  })

  it('listMyInvitations calls GET /api/v1/invitations', async () => {
    const invitations = [{ id: 1, team_id: 1, status: 'pending' as const, expires_at: '', created_at: '' }]
    mockFetch.mockResolvedValue(okResponse(invitations))

    const result = await listMyInvitations()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/invitations',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result).toHaveLength(1)
  })

  it('acceptInvitation calls POST with correct URL', async () => {
    const data = { message: 'accepted', team_id: 1, team_name: 'Team A' }
    mockFetch.mockResolvedValue(okResponse(data))

    const result = await acceptInvitation(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/invitations/1/accept',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.team_name).toBe('Team A')
  })

  it('declineInvitation calls POST with correct URL', async () => {
    mockFetch.mockResolvedValue(okResponse({ message: 'declined' }))

    const result = await declineInvitation(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/invitations/1/decline',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.message).toBe('declined')
  })

  it('getInviteLinkInfo calls GET with correct URL', async () => {
    const info = { team_id: 1, team_name: 'Team A', expires_at: '', is_expired: false }
    mockFetch.mockResolvedValue(okResponse(info))

    const result = await getInviteLinkInfo('abc123')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/invite-links/abc123',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.team_name).toBe('Team A')
  })

  it('acceptInviteLink calls POST with correct URL', async () => {
    const data = { message: 'accepted', team_id: 1, team_name: 'Team A' }
    mockFetch.mockResolvedValue(okResponse(data))

    const result = await acceptInviteLink('abc123')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/invite-links/abc123/accept',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.team_name).toBe('Team A')
  })

  it('searchUsers calls GET with query param', async () => {
    const data = { users: [{ id: 1, display_name: 'Alice', discord_username: 'alice#1234' }] }
    mockFetch.mockResolvedValue(okResponse(data))

    const result = await searchUsers('Ali')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/users/search?q=Ali',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.users).toHaveLength(1)
    expect(result.users[0].display_name).toBe('Alice')
  })
})
