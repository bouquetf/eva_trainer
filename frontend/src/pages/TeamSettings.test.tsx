import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import TeamSettings from './TeamSettings'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

// Mock the client module so getApiClient returns a client that uses the mocked fetch
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

    async get<T>(path: string): Promise<T> {
      return this.request<T>(path, { method: 'GET' })
    }

    async post<T>(path: string, body?: unknown): Promise<T> {
      return this.request<T>(path, {
        method: 'POST',
        headers: body ? { 'Content-Type': 'application/json' } : undefined,
        body: body ? JSON.stringify(body) : undefined,
      })
    }

    async put<T>(path: string, body?: unknown): Promise<T> {
      return this.request<T>(path, {
        method: 'PUT',
        headers: body ? { 'Content-Type': 'application/json' } : undefined,
        body: body ? JSON.stringify(body) : undefined,
      })
    }

    async del<T>(path: string, body?: unknown): Promise<T> {
      return this.request<T>(path, {
        method: 'DELETE',
        headers: body ? { 'Content-Type': 'application/json' } : undefined,
        body: body ? JSON.stringify(body) : undefined,
      })
    }
  }

  return { getApiClient: () => new MockApiClient() }
})

// Mock the auth store so user is available
vi.mock('../store/auth', () => ({
  useAuthStore: {
    getState: () => ({
      user: { id: 1, discord_id: 'test', discord_username: '', display_name: '', is_admin: false, created_at: '2024-01-01T00:00:00Z' },
    }),
  },
}))

function createQueryWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/teams/1']}>
          <Routes>
            <Route path="/teams/:id" element={children} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    )
  }
}

function mockTeamDetail(isCaptain: boolean) {
  const captainId = isCaptain ? 1 : 2
  return {
    team: { id: 1, name: 'Test Team', captain_id: captainId, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
    members: [
      { user_id: 1, role: 'captain', display_name: null, joined_at: '2024-01-01T00:00:00Z' },
      { user_id: 2, role: 'editor', display_name: null, joined_at: '2024-01-01T00:00:00Z' },
    ],
  }
}

function okResponse(body: unknown) {
  return {
    ok: true,
    status: 200,
    json: () => Promise.resolve(body),
  }
}

function notFoundResponse() {
  return {
    ok: false,
    status: 404,
    json: () => Promise.resolve({ error: 'team_not_found', message: 'team not found' }),
  }
}

describe('TeamSettings', () => {
  beforeEach(() => {
    mockFetch.mockReset()
  })

  it('shows loading spinner initially', () => {
    mockFetch.mockResolvedValue(new Promise(() => {}))
    render(<TeamSettings />, { wrapper: createQueryWrapper() })
    expect(document.querySelector('.spinner')).toBeTruthy()
  })

  it('displays team details when loaded', async () => {
    mockFetch.mockResolvedValue(okResponse(mockTeamDetail(true)))

    render(<TeamSettings />, { wrapper: createQueryWrapper() })

    expect(await screen.findAllByText('Test Team')).toHaveLength(3)
  })

  it('shows members section', async () => {
    mockFetch.mockResolvedValue(okResponse(mockTeamDetail(true)))

    render(<TeamSettings />, { wrapper: createQueryWrapper() })

    expect(await screen.findByText(/Membres/)).toBeInTheDocument()
  })

  it('shows danger zone for captain', async () => {
    mockFetch.mockResolvedValue(okResponse(mockTeamDetail(true)))

    render(<TeamSettings />, { wrapper: createQueryWrapper() })

    expect(await screen.findByText('Zone de danger')).toBeInTheDocument()
  })

  it('does not show danger zone for non-captain', async () => {
    mockFetch.mockResolvedValue(okResponse(mockTeamDetail(false)))

    render(<TeamSettings />, { wrapper: createQueryWrapper() })

    await waitFor(() => {
      expect(screen.queryByText('Zone de danger')).not.toBeInTheDocument()
    })
  })

  it('shows leave button for non-captain', async () => {
    mockFetch.mockResolvedValue(okResponse(mockTeamDetail(false)))

    render(<TeamSettings />, { wrapper: createQueryWrapper() })

    expect(await screen.findAllByText("Quitter l'équipe")).toHaveLength(2)
  })

  it('does not show leave button for captain', async () => {
    mockFetch.mockResolvedValue(okResponse(mockTeamDetail(true)))

    render(<TeamSettings />, { wrapper: createQueryWrapper() })

    await waitFor(() => {
      expect(screen.queryByText("Quitter l'équipe")).not.toBeInTheDocument()
    })
  })

  it('shows error message when team not found', async () => {
    mockFetch.mockResolvedValue(notFoundResponse())

    render(<TeamSettings />, { wrapper: createQueryWrapper() })

    expect(await screen.findByText('Équipe introuvable.')).toBeInTheDocument()
  })
})
