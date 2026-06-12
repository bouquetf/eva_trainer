import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import AdminLiguesTab from './AdminLiguesTab'

const mockListLigues = vi.fn()
const mockListSplits = vi.fn()
vi.mock('../api/archives', () => ({
  listLigues: (...args: unknown[]) => mockListLigues(...args),
  createLigue: vi.fn(),
  updateLigue: vi.fn(),
  deleteLigue: vi.fn(),
  refreshLigue: vi.fn(),
}))
vi.mock('../api/splits', () => ({
  listSplits: (...args: unknown[]) => mockListSplits(...args),
}))

function renderWithQuery(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('AdminLiguesTab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockListSplits.mockResolvedValue({ splits: [] })
  })

  it('renders the create button', () => {
    mockListLigues.mockResolvedValue({ items: [], total: 0, page: 1, per_page: 5 })
    renderWithQuery(<AdminLiguesTab onSuccess={() => {}} onError={() => {}} />)
    expect(screen.getByText('Nouvelle ligue')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    mockListLigues.mockReturnValue(new Promise(() => {}))
    renderWithQuery(<AdminLiguesTab onSuccess={() => {}} onError={() => {}} />)
    expect(screen.getByText('Chargement...')).toBeInTheDocument()
  })

  it('shows empty state when no ligues', async () => {
    mockListLigues.mockResolvedValue({ items: [], total: 0, page: 1, per_page: 5 })
    renderWithQuery(<AdminLiguesTab onSuccess={() => {}} onError={() => {}} />)
    expect(await screen.findByText('Aucune ligue')).toBeInTheDocument()
  })
})
