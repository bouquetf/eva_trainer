import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import ProfileSettings from './ProfileSettings'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

vi.mock('../api/client', () => {
  class MockApiClient {
    private baseURL = 'http://localhost:8080'

    private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
      const token = 'test-token'
      const headers: Record<string, string> = {
        ...(options.headers as Record<string, string>),
        Authorization: `Bearer ${token}`,
      }
      const res = await fetch(`${this.baseURL}${path}`, { ...options, headers })
      if (!res.ok) throw { status: res.status, body: await res.json().catch(() => ({})) }
      return res.json()
    }

    async get<T>(path: string): Promise<T> { return this.request<T>(path, { method: 'GET' }) }
    async post<T>(path: string, body?: unknown): Promise<T> { return this.request<T>(path, { method: 'POST', headers: body ? { 'Content-Type': 'application/json' } : undefined, body: body ? JSON.stringify(body) : undefined }) }
    async put<T>(path: string, body?: unknown): Promise<T> { return this.request<T>(path, { method: 'PUT', headers: body ? { 'Content-Type': 'application/json' } : undefined, body: body ? JSON.stringify(body) : undefined }) }
    async del<T>(path: string, body?: unknown): Promise<T> { return this.request<T>(path, { method: 'DELETE', headers: body ? { 'Content-Type': 'application/json' } : undefined, body: body ? JSON.stringify(body) : undefined }) }
  }

  return { getApiClient: () => new MockApiClient() }
})

function createQueryWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

function okResponse(body: unknown) {
  return {
    ok: true,
    status: 200,
    json: () => Promise.resolve(body),
  }
}

const defaultUser = {
  id: 1,
  discord_id: 'discord123',
  discord_username: 'monpseudo#1234',
  display_name: 'MonPseudo',
  is_admin: false,
  created_at: '2024-01-01T00:00:00Z',
}

vi.mock('../store/auth', () => {
  let currentUser = {
    id: 1,
    discord_id: 'discord123',
    discord_username: 'monpseudo#1234',
    display_name: 'MonPseudo',
    is_admin: false,
    created_at: '2024-01-01T00:00:00Z',
  }

  return {
    useAuthStore: Object.assign(
      (selector?: (s: { user: typeof currentUser }) => unknown) => {
        const state = { user: currentUser }
        return selector ? selector(state) : state
      },
      {
        getState: () => ({ user: currentUser }),
        setState: (partial: { user?: typeof currentUser } | ((state: { user: typeof currentUser }) => { user?: typeof currentUser })) => {
          const updater = typeof partial === 'function' ? partial({ user: currentUser }) : partial
          if (updater.user) currentUser = updater.user
        },
      },
    ),
  }
})

describe('ProfileSettings', () => {
  beforeEach(() => {
    mockFetch.mockReset()
  })

  it('affiche le titre de la page', async () => {
    mockFetch.mockResolvedValue(okResponse(defaultUser))
    render(<ProfileSettings />, { wrapper: createQueryWrapper() })
    expect(await screen.findByText('Paramètres du profil')).toBeInTheDocument()
  })

  it('affiche le pseudo Discord', async () => {
    mockFetch.mockResolvedValue(okResponse(defaultUser))
    render(<ProfileSettings />, { wrapper: createQueryWrapper() })
    expect(await screen.findByText(/monpseudo#1234/)).toBeInTheDocument()
  })

  it('affiche le display_name depuis la base de données', async () => {
    mockFetch.mockResolvedValue(okResponse(defaultUser))
    render(<ProfileSettings />, { wrapper: createQueryWrapper() })
    expect(await screen.findByText('MonPseudo')).toBeInTheDocument()
  })

  it('affiche #id quand display_name est vide', async () => {
    const userNoDisplay = { ...defaultUser, display_name: '' }
    mockFetch.mockResolvedValue(okResponse(userNoDisplay))
    render(<ProfileSettings />, { wrapper: createQueryWrapper() })
    expect(await screen.findByText('#1')).toBeInTheDocument()
  })
})
