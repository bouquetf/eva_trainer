import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getLoginUrl, fetchMe, refreshToken, updateDisplayName } from './auth'

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
  vi.stubEnv('VITE_API_URL', 'http://localhost:8080')
})

afterEach(() => {
  vi.unstubAllEnvs()
})

describe('auth API', () => {
  it('getLoginUrl returns the Discord OAuth URL with redirect to /login', () => {
    const url = getLoginUrl()
    expect(url).toContain('/api/v1/auth/discord/login?redirect=')
  })

  it('getLoginUrl redirects to / when current path is /login', () => {
    window.location.pathname = '/login'
    const url = getLoginUrl()
    expect(url).toContain('redirect=%2F')
  })

  it('fetchMe calls GET /api/v1/auth/me', async () => {
    const user = { id: 1, discord_id: '123', discord_username: 'testuser', display_name: 'Test', is_admin: false, created_at: '2025-01-01T00:00:00Z' }
    mockFetch.mockResolvedValue(okResponse(user))

    const result = await fetchMe()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/auth/me',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result.discord_username).toBe('testuser')
  })

  it('refreshToken calls POST /api/v1/auth/refresh', async () => {
    const response = { access_token: 'new-access', refresh_token: 'new-refresh', expires_in: 3600 }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await refreshToken('old-refresh')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/auth/refresh',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ refresh_token: 'old-refresh' }),
      }),
    )
    expect(result.access_token).toBe('new-access')
  })

  it('updateDisplayName calls PUT /api/v1/auth/display-name', async () => {
    const response = { display_name: 'NewName' }
    mockFetch.mockResolvedValue(okResponse(response))

    const result = await updateDisplayName('NewName')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/auth/display-name',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ display_name: 'NewName' }),
      }),
    )
    expect(result.display_name).toBe('NewName')
  })

  it('throws on API error', async () => {
    mockFetch.mockResolvedValue({ ok: false, status: 401, json: () => Promise.resolve({}) })

    await expect(fetchMe()).rejects.toThrow('API Error')
  })
})
