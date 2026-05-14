import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  listAdminUsers,
  getUserDetail,
  listAdminPermissions,
  getRolePermissions,
  listAdminTeams,
  promoteUser,
  updateUserRole,
  addUserToTeam,
  removeUserFromTeam,
} from './admin'

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

vi.mock('./client', () => {
  class MockApiError extends Error {
    status: number
    body: Record<string, unknown>
    constructor(status: number, body: Record<string, unknown>) {
      super(String(body.message ?? body.error ?? 'API Error'))
      this.status = status
      this.body = body
    }
  }
  return {
    __initClient: vi.fn(),
    getApiClient: () => mockApiClient(),
    ApiError: MockApiError,
  }
})

function okResponse(body: unknown) {
  return { ok: true, status: 200, json: () => Promise.resolve(body) }
}

beforeEach(() => {
  mockFetch.mockReset()
})

describe('admin API', () => {
  it('listAdminUsers calls GET /api/v1/admin/users', async () => {
    const response = {
      users: [{ id: 1, discord_id: '123', discord_username: 'admin', display_name: 'Admin', is_admin: true, created_at: '' }],
    }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await listAdminUsers()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/users',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.users).toHaveLength(1)
  })

  it('getUserDetail calls GET /api/v1/admin/users/:id', async () => {
    const response = {
      user: { id: 1, discord_id: '123', discord_username: 'user1', display_name: 'User1', is_admin: false, created_at: '' },
      memberships: [{ team_id: 1, team_name: 'Team A', role: 'member' }],
    }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await getUserDetail(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/users/1',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.memberships).toHaveLength(1)
  })

  it('listAdminPermissions calls GET /api/v1/admin/permissions', async () => {
    const response = {
      permissions: [{ id: 1, feature: 'maps.upload', description: 'Upload maps', roles: ['admin'] }],
      roles: ['admin'],
    }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await listAdminPermissions()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/permissions',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.permissions).toHaveLength(1)
  })

  it('getRolePermissions calls GET /api/v1/admin/permissions/:role', async () => {
    const response = {
      role: 'admin',
      permissions: [{ id: 1, feature: 'maps.upload', description: 'Upload maps' }],
    }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await getRolePermissions('admin')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/permissions/admin',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.role).toBe('admin')
  })

  it('listAdminTeams calls GET /api/v1/admin/teams', async () => {
    const response = { teams: [{ id: 1, name: 'Team A', captain_id: 1 }] }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await listAdminTeams()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/teams',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.teams).toHaveLength(1)
  })

  it('promoteUser calls POST /api/v1/admin/users/:id/promote', async () => {
    mockFetch.mockResolvedValue(okResponse({}))

    await promoteUser(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/users/1/promote',
      expect.objectContaining({ method: 'POST' }),
    )
  })

  it('updateUserRole calls PUT /api/v1/admin/users/:id/teams/:teamId/role', async () => {
    mockFetch.mockResolvedValue(okResponse({}))

    await updateUserRole(1, 2, 'captain')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/users/1/teams/2/role',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ role: 'captain' }),
      }),
    )
  })

  it('addUserToTeam calls POST /api/v1/admin/users/:id/teams', async () => {
    mockFetch.mockResolvedValue(okResponse({}))

    await addUserToTeam(1, 2, 'member')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/users/1/teams',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ team_id: 2, role: 'member' }),
      }),
    )
  })

  it('removeUserFromTeam calls DELETE /api/v1/admin/users/:id/teams/:teamId', async () => {
    mockFetch.mockResolvedValue(okResponse({}))

    await removeUserFromTeam(1, 2)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/users/1/teams/2',
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('throws 403 error for non-admin', async () => {
    mockFetch.mockResolvedValue({ ok: false, status: 403, json: () => Promise.resolve({ error: 'forbidden', message: 'Forbidden' }) })

    await expect(listAdminUsers()).rejects.toThrow('API Error')
  })
})
