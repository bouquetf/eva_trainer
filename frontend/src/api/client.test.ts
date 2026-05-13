import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

function okResponse(body: unknown, status = 200) {
  return { ok: status >= 200 && status < 300, status, json: () => Promise.resolve(body) }
}

function errorResponse(status: number, body: Record<string, unknown>) {
  return { ok: false, status, json: () => Promise.resolve(body) }
}

beforeEach(() => {
  mockFetch.mockReset()
})

describe('ApiError', () => {
  it('stores status and body from constructor', async () => {
    const { ApiError } = await import('./client')
    const error = new ApiError(404, { error: 'not_found', message: 'Resource not found' })
    expect(error.status).toBe(404)
    expect(error.body).toEqual({ error: 'not_found', message: 'Resource not found' })
    expect(error.message).toBe('Resource not found')
  })

  it('falls back to error code when message is missing', async () => {
    const { ApiError } = await import('./client')
    const error = new ApiError(403, { error: 'forbidden' })
    expect(error.message).toBe('forbidden')
  })
})

describe('ApiClient', () => {
  async function createClient(overrides?: Record<string, unknown>) {
    const { ApiClient } = await import('./client')
    const store = {
      getAccessToken: vi.fn().mockReturnValue('test-token'),
      getRefreshToken: vi.fn().mockReturnValue(null),
      setTokens: vi.fn(),
      logout: vi.fn(),
    }
    return {
      client: new ApiClient({
        baseURL: overrides?.baseURL ?? 'http://localhost:8080',
        getAccessToken: () => store.getAccessToken(),
        getRefreshToken: () => store.getRefreshToken(),
        onRefresh: (...args: [string, string]) => store.setTokens(...args),
        onAuthError: () => store.logout(),
      }),
      store,
    }
  }

  it('sends GET request with Authorization header', async () => {
    const { client } = await createClient()
    mockFetch.mockResolvedValue(okResponse({ data: 'ok' }))

    await client.get('/api/v1/test')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/test',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer test-token' }),
      }),
    )
  })

  it('sends POST request with JSON body', async () => {
    const { client } = await createClient()
    mockFetch.mockResolvedValue(okResponse({ id: 1 }))

    await client.post('/api/v1/test', { name: 'test' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/test',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
        body: JSON.stringify({ name: 'test' }),
      }),
    )
  })

  it('sends PUT request with JSON body', async () => {
    const { client } = await createClient()
    mockFetch.mockResolvedValue(okResponse({ updated: true }))

    await client.put('/api/v1/test', { value: 'new' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/test',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ value: 'new' }),
      }),
    )
  })

  it('sends DELETE request', async () => {
    const { client } = await createClient()
    mockFetch.mockResolvedValue(okResponse({ deleted: true }))

    await client.del('/api/v1/test')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/test',
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('sends upload with FormData and no Content-Type', async () => {
    const { client } = await createClient()
    mockFetch.mockResolvedValue(okResponse({ id: 1 }))
    const formData = new FormData()
    formData.append('file', new Blob(['data']), 'test.txt')

    await client.upload('/api/v1/admin/maps', formData)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/maps',
      expect.objectContaining({ method: 'POST' }),
    )
  })

  it('handles 401 by refreshing token and retrying', async () => {
    const { ApiClient } = await import('./client')
    let currentToken = 'old-token'
    const getAccessToken = vi.fn(() => currentToken)
    const getRefreshToken = vi.fn(() => 'old-refresh')
    const setTokens = vi.fn((access: string) => { currentToken = access })
    const logout = vi.fn()

    const client = new ApiClient({
      baseURL: 'http://localhost:8080',
      getAccessToken: () => getAccessToken(),
      getRefreshToken: () => getRefreshToken(),
      onRefresh: (a, r) => setTokens(a, r),
      onAuthError: () => logout(),
    })

    mockFetch
      .mockResolvedValueOnce(errorResponse(401, {}))
      .mockResolvedValueOnce(okResponse({ access_token: 'new-token', refresh_token: 'new-refresh' }))
      .mockResolvedValueOnce(okResponse({ data: 'retried' }))

    const result = await client.get('/api/v1/test')

    expect(setTokens).toHaveBeenCalledWith('new-token', 'new-refresh')
    expect(result).toEqual({ data: 'retried' })
    expect(logout).not.toHaveBeenCalled()
  })

  it('calls onAuthError when 401 and refresh fails', async () => {
    const { ApiClient, ApiError } = await import('./client')
    const logout = vi.fn()

    const client = new ApiClient({
      baseURL: 'http://localhost:8080',
      getAccessToken: () => 'token',
      getRefreshToken: () => null,
      onRefresh: () => {},
      onAuthError: () => logout(),
    })

    mockFetch.mockResolvedValue(errorResponse(401, {}))

    await expect(client.get('/api/v1/test')).rejects.toThrow(ApiError)
    expect(logout).toHaveBeenCalled()
  })

  it('throws ApiError on non-ok response', async () => {
    const { ApiClient, ApiError } = await import('./client')

    const client = new ApiClient({
      baseURL: 'http://localhost:8080',
      getAccessToken: () => 'token',
      getRefreshToken: () => null,
      onRefresh: () => {},
      onAuthError: () => {},
    })

    mockFetch.mockResolvedValue(errorResponse(500, { error: 'server_error', message: 'Internal error' }))

    await expect(client.get('/api/v1/test')).rejects.toThrow(ApiError)
  })

  it('handles network errors', async () => {
    const { ApiClient } = await import('./client')

    const client = new ApiClient({
      baseURL: 'http://localhost:8080',
      getAccessToken: () => 'token',
      getRefreshToken: () => null,
      onRefresh: () => {},
      onAuthError: () => {},
    })

    mockFetch.mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(client.get('/api/v1/test')).rejects.toThrow(TypeError)
  })

  it('returns parsed JSON on success', async () => {
    const { client } = await createClient()
    mockFetch.mockResolvedValue(okResponse({ user: { id: 1, name: 'test' } }))

    const result = await client.get('/api/v1/auth/me')

    expect(result).toEqual({ user: { id: 1, name: 'test' } })
  })
})

describe('singleton', () => {
  beforeEach(() => {
    vi.resetModules()
  })

  it('getApiClient throws when not initialized', async () => {
    const { getApiClient } = await import('./client')
    expect(() => getApiClient()).toThrow('ApiClient not initialized')
  })

  it('__initClient creates the singleton', async () => {
    const mod = await import('./client')
    const store = {
      getAccessToken: () => 'token',
      getRefreshToken: () => null,
      setTokens: () => {},
      logout: () => {},
    }
    mod.__initClient(() => store)
    expect(mod.getApiClient()).toBeDefined()
  })

  it('getApiClient returns the same instance', async () => {
    const mod = await import('./client')
    const store = {
      getAccessToken: () => 'token',
      getRefreshToken: () => null,
      setTokens: () => {},
      logout: () => {},
    }
    mod.__initClient(() => store)
    const instance1 = mod.getApiClient()
    const instance2 = mod.getApiClient()
    expect(instance1).toBe(instance2)
  })
})
