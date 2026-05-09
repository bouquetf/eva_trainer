import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createTeam, listMyTeams, listAllTeams, getTeam, updateTeam, deleteTeam, leaveTeam } from './teams'

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

describe('teams API', () => {
  it('createTeam calls POST /api/v1/teams with correct body', async () => {
    mockFetch.mockResolvedValue(okResponse({ id: 1, name: 'New Team', captain_id: 1, created_at: '', updated_at: '' }))

    const result = await createTeam({ name: 'New Team' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'New Team' }),
      }),
    )
    expect(result.name).toBe('New Team')
  })

  it('listMyTeams calls GET /api/v1/myteams and returns list', async () => {
    const teams = [{ team: { id: 1, name: 'Team A', captain_id: 1 }, member_count: 2, my_role: 'captain' }]
    mockFetch.mockResolvedValue(okResponse(teams))

    const result = await listMyTeams()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/myteams',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result).toHaveLength(1)
    expect(result[0].team.name).toBe('Team A')
  })

  it('listAllTeams calls GET /api/v1/teams with query params', async () => {
    const data = { teams: [{ team: { id: 1, name: 'Team A', captain_id: 1 }, member_count: 2, my_role: '' }], total: 1, page: 1, per_page: 20 }
    mockFetch.mockResolvedValue(okResponse(data))

    const result = await listAllTeams({ page: 1, per_page: 20, q: 'Team' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams?page=1&per_page=20&q=Team',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.teams).toHaveLength(1)
    expect(result.total).toBe(1)
  })

  it('listAllTeams calls GET /api/v1/teams without params', async () => {
    const data = { teams: [], total: 0, page: 1, per_page: 20 }
    mockFetch.mockResolvedValue(okResponse(data))

    const result = await listAllTeams()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.teams).toHaveLength(0)
  })

  it('getTeam calls GET /api/v1/teams/:id', async () => {
    const detail = { team: { id: 1, name: 'Team A', captain_id: 1, created_at: '', updated_at: '' }, members: [] }
    mockFetch.mockResolvedValue(okResponse(detail))

    const result = await getTeam(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.team.name).toBe('Team A')
  })

  it('updateTeam calls PUT /api/v1/teams/:id with correct body', async () => {
    mockFetch.mockResolvedValue(okResponse({ id: 1, name: 'Updated', captain_id: 1, created_at: '', updated_at: '' }))

    const result = await updateTeam(1, { name: 'Updated' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ name: 'Updated' }),
      }),
    )
    expect(result.name).toBe('Updated')
  })

  it('deleteTeam calls DELETE /api/v1/teams/:id with confirmation body', async () => {
    mockFetch.mockResolvedValue(okResponse({ message: 'team deleted' }))

    const result = await deleteTeam(1, { confirm_with_name: 'My Team' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1',
      expect.objectContaining({
        method: 'DELETE',
        body: JSON.stringify({ confirm_with_name: 'My Team' }),
      }),
    )
    expect(result.message).toBe('team deleted')
  })

  it('leaveTeam calls POST /api/v1/teams/:id/leave', async () => {
    mockFetch.mockResolvedValue(okResponse({ message: 'left team' }))

    const result = await leaveTeam(1)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/leave',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.message).toBe('left team')
  })
})
