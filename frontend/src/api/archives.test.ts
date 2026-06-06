import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listLiguesLocales, createLigueLocale, updateLigueLocale, deleteLigueLocale, getLigueLocale } from './archives'

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

describe('archives API', () => {
  it('listLiguesLocales calls GET /api/v1/archive/leagues-locales sans filtre', async () => {
    const expected = { ligues_locales: [{ id: 1, nom: 'Test', emplacement: 'Grenoble' }] }
    mockFetch.mockResolvedValue(okResponse(expected))

    const result = await listLiguesLocales()

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/archive/leagues-locales',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result).toEqual(expected)
  })

  it('listLiguesLocales avec ?emplacement=', async () => {
    mockFetch.mockResolvedValue(okResponse({ ligues_locales: [] }))

    await listLiguesLocales('Grenoble')

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/archive/leagues-locales?emplacement=Grenoble',
      expect.objectContaining({ method: 'GET' }),
    )
  })

  it('createLigueLocale calls POST /api/v1/archive/leagues-locales', async () => {
    const payload = { nom: 'Ligue', emplacement: 'Grenoble' }
    const expected = { ligue_locale: { id: 1, ...payload, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' } }
    mockFetch.mockResolvedValue(okResponse(expected))

    const result = await createLigueLocale(payload)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/archive/leagues-locales',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify(payload),
      }),
    )
    expect(result).toEqual(expected)
  })

  it('updateLigueLocale calls PUT /api/v1/archive/leagues-locales/:id', async () => {
    const payload = { nom: 'Updated', emplacement: 'Lyon' }
    const expected = { ligue_locale: { id: 42, ...payload, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' } }
    mockFetch.mockResolvedValue(okResponse(expected))

    const result = await updateLigueLocale(42, payload)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/archive/leagues-locales/42',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify(payload),
      }),
    )
    expect(result).toEqual(expected)
  })

  it('deleteLigueLocale calls DELETE /api/v1/archive/leagues-locales/:id', async () => {
    mockFetch.mockResolvedValue(okResponse({ message: 'ligue locale deleted' }))

    const result = await deleteLigueLocale(7)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/archive/leagues-locales/7',
      expect.objectContaining({ method: 'DELETE' }),
    )
    expect(result).toEqual({ message: 'ligue locale deleted' })
  })

  it('getLigueLocale calls GET /api/v1/archive/leagues-locales/:id', async () => {
    const expected = {
      ligue_locale: { id: 3, nom: 'Detail', emplacement: 'Grenoble', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
      ligues: [{ id: 1, date_debut: '2026-01-01', date_fin: '2026-02-01', split_name: 'Split 1' }],
    }
    mockFetch.mockResolvedValue(okResponse(expected))

    const result = await getLigueLocale(3)

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/archive/leagues-locales/3',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(result).toEqual(expected)
  })

  it('throws on API error', async () => {
    mockFetch.mockResolvedValue({ ok: false, status: 500, json: () => Promise.resolve({}) })

    await expect(listLiguesLocales()).rejects.toThrow('API Error')
  })
})
