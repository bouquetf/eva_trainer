import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listMembers, updateMemberRole, transferCaptaincy } from './roles'

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

describe('roles API', () => {
  it('listMembers calls GET with the correct teamId', async () => {
    const membersData = {
      members: [
        { user_id: 1, display_name: 'Alice', role: 'captain', joined_at: '2024-01-01T00:00:00Z' },
        { user_id: 2, display_name: 'Bob', role: 'editor', joined_at: '2024-01-02T00:00:00Z' },
      ],
    }
    mockFetch.mockResolvedValue(okResponse(membersData))

    const result = await listMembers(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/members',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.members).toHaveLength(2)
    expect(result.members[0].role).toBe('captain')
    expect(result.members[1].role).toBe('editor')
  })

  it('updateMemberRole calls PUT with correct body and IDs', async () => {
    mockFetch.mockResolvedValue(okResponse({ message: 'role updated', user_id: 2, role: 'editor' }))

    const result = await updateMemberRole(1, 2, { role: 'editor' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/members/2/role',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ role: 'editor' }),
      }),
    )
    expect(result.message).toBe('role updated')
    expect(result.role).toBe('editor')
  })

  it('transferCaptaincy calls POST with correct body', async () => {
    mockFetch.mockResolvedValue(okResponse({ message: 'captaincy transferred', new_captain_id: 3 }))

    const result = await transferCaptaincy(1, { user_id: 3 })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/transfer-captaincy',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ user_id: 3 }),
      }),
    )
    expect(result.message).toBe('captaincy transferred')
    expect(result.new_captain_id).toBe(3)
  })
})
