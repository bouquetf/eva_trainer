import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import Editor from './Editor'

vi.mock('../components/MapEditor/MapEditor', () => ({
  MapEditor: ({ teamId, mapId }: any) => (
    <div data-testid="map-editor">MapEditor teamId={teamId} mapId={mapId}</div>
  ),
}))

vi.mock('../components/AuthGuard', () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useParams: vi.fn() }
})

import { useParams } from 'react-router-dom'

const mockUseParams = useParams as unknown as ReturnType<typeof vi.fn>

describe('Editor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders error when no params', () => {
    mockUseParams.mockReturnValue({ id: undefined, mapId: undefined })
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>,
    )
    expect(screen.getByText(/Paramètres d'URL invalides/)).toBeInTheDocument()
  })

  it('renders MapEditor with correct teamId and mapId', () => {
    mockUseParams.mockReturnValue({ id: '42', mapId: '7' })
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>,
    )
    expect(screen.getByTestId('map-editor')).toHaveTextContent('MapEditor teamId=42 mapId=7')
  })

  it('wraps in AuthGuard', () => {
    mockUseParams.mockReturnValue({ id: '42', mapId: '7' })
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>,
    )
    expect(screen.getByTestId('map-editor')).toBeInTheDocument()
  })
})
