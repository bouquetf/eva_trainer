import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import Revision from './Revision'
import { useRevisionStore } from '../store/revision'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

vi.mock('../api/client', () => ({
  getApiClient: () => {
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
      async post<T>(path: string, _body?: unknown): Promise<T> {
        const token = 'test-token'
        const res = await fetch(`http://localhost:8080${path}`, {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}` },
        })
        if (!res.ok) throw new Error('API Error')
        return res.json()
      }
    }
    return new MockApiClient()
  },
}))

function okResponse(body: unknown) {
  return {
    ok: true,
    status: 200,
    json: () => Promise.resolve(body),
  }
}

function renderRevision(teamId = '1') {
  return render(
    <MemoryRouter initialEntries={[`/teams/${teamId}/revision`]}>
      <Routes>
        <Route path="/teams/:id/revision" element={<Revision />} />
        <Route path="/teams/:id/revision/:mapId" element={<Revision />} />
      </Routes>
    </MemoryRouter>,
  )
}

beforeEach(() => {
  mockFetch.mockReset()
  useRevisionStore.getState().resetSession()
})

describe('Revision page - map selection', () => {
  it('displays map list', async () => {
    mockFetch
      .mockResolvedValueOnce(okResponse({ maps: [{ id: 1, name: 'Map A', added_at: '', is_obsolete: false, obsolete_of: null }] }))

    renderRevision()

    await waitFor(() => {
      expect(screen.getByText('Map A')).toBeTruthy()
    })
  })

  it('selects a map on click', async () => {
    mockFetch
      .mockResolvedValueOnce(okResponse({ maps: [{ id: 1, name: 'Map A', added_at: '', is_obsolete: false, obsolete_of: null }] }))

    renderRevision()

    await waitFor(() => {
      expect(screen.getByText('Map A')).toBeTruthy()
    })

    await userEvent.click(screen.getByText('Map A'))

    expect(screen.getByText('Nombre de questions')).toBeTruthy()
  })
})

describe('Revision page - active revision', () => {
  it('starts revision from selection screen', async () => {
    mockFetch
      .mockResolvedValueOnce(okResponse({ maps: [{ id: 1, name: 'Map A', added_at: '', is_obsolete: false, obsolete_of: null }] }))
      .mockResolvedValueOnce(okResponse({
        zones: [
          { id: 1, source: 'team', map_id: 1, polygons: [], name: 'top' },
          { id: 2, source: 'team', map_id: 1, polygons: [], name: 'secret' },
        ],
      }))

    renderRevision()

    await waitFor(() => expect(screen.getByText('Map A')).toBeTruthy())

    await userEvent.click(screen.getByText('Map A'))
    await userEvent.click(screen.getByText('Lancer la révision'))

    await waitFor(() => {
      expect(useRevisionStore.getState().isActive).toBe(true)
    })
  })
})

describe('Revision page - finished screen', () => {
  it('navigates to finish after manual finishSession', async () => {
    mockFetch
      .mockResolvedValueOnce(okResponse({ maps: [{ id: 1, name: 'Map A', added_at: '', is_obsolete: false, obsolete_of: null }] }))
      .mockResolvedValueOnce(okResponse({
        zones: [
          { id: 1, source: 'team', map_id: 1, polygons: [], name: 'top' },
        ],
      }))

    renderRevision()

    await waitFor(() => expect(screen.getByText('Map A')).toBeTruthy())
    await userEvent.click(screen.getByText('Map A'))
    await userEvent.click(screen.getByText('Lancer la révision'))

    await waitFor(() => {
      expect(useRevisionStore.getState().isActive).toBe(true)
    })

    useRevisionStore.getState().finishSession()

    await waitFor(() => {
      expect(screen.getByText('Révision terminée !')).toBeTruthy()
    })
  })
})

describe('Revision page - error state', () => {
  it('shows error when fetch fails', async () => {
    mockFetch
      .mockRejectedValueOnce(new Error('Network error'))

    renderRevision()

    await waitFor(() => {
      expect(screen.getByText('Erreur lors du chargement des cartes.')).toBeTruthy()
    })
  })
})
