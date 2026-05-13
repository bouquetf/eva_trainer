import { describe, it, expect, vi, beforeEach } from 'vitest'
import { publishMap, getPublishStatus, listPublicTemplates, duplicateTemplate } from './templates'

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

describe('templates API', () => {
  it('publishMap calls POST /api/v1/teams/:id/maps/:mapId/publish', async () => {
    const template = { id: 1, original_team_id: 1, original_team_name: 'Team A', map_id: 5, is_public: true, map_name: 'Test Map', created_at: '', updated_at: '' }
    mockFetch.mockResolvedValue(okResponse(template))

    const result = await publishMap(1, 5)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/5/publish',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.id).toBe(1)
  })

  it('getPublishStatus calls GET /api/v1/teams/:id/maps/:mapId/publish-status', async () => {
    const status = { published: false }
    mockFetch.mockResolvedValue(okResponse(status))

    const result = await getPublishStatus(1, 5)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/maps/5/publish-status',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.published).toBe(false)
  })

  it('getPublishStatus returns published: true with a template', async () => {
    const status = { published: true, template: { id: 1, original_team_id: 1, original_team_name: 'Team A', map_id: 5, is_public: true, map_name: 'Test Map', created_at: '', updated_at: '' } }
    mockFetch.mockResolvedValue(okResponse(status))

    const result = await getPublishStatus(1, 5)

    expect(result.published).toBe(true)
    expect(result.template).toBeDefined()
  })

  it('listPublicTemplates calls GET /api/v1/templates/public', async () => {
    const response = {
      templates: [
        { template: { id: 1, original_team_id: 1, original_team_name: 'Team A', map_id: 5, is_public: true, map_name: 'Map A', created_at: '', updated_at: '' }, zones: [] },
      ],
    }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await listPublicTemplates()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/templates/public',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.templates).toHaveLength(1)
  })

  it('duplicateTemplate calls POST /api/v1/teams/:id/templates/:templateId/duplicate', async () => {
    const response = { zones: [] }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await duplicateTemplate(1, 3)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/teams/1/templates/3/duplicate',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(result.zones).toEqual([])
  })

  it('throws on API error', async () => {
    mockFetch.mockResolvedValue({ ok: false, status: 500, json: () => Promise.resolve({}) })

    await expect(publishMap(1, 1)).rejects.toThrow('API Error')
  })
})
