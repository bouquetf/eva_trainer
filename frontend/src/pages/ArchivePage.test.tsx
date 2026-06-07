import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import ArchivePage from './ArchivePage'

vi.mock('../hooks/useArchiveAccess', () => ({
  useArchiveAccess: vi.fn(),
}))

vi.mock('../api/archives', () => ({
  listLiguesLocales: vi.fn(),
  createLigueLocale: vi.fn(),
  updateLigueLocale: vi.fn(),
  deleteLigueLocale: vi.fn(),
  listLigues: vi.fn(),
  listDivs: vi.fn(),
  getDiv: vi.fn(),
  createDiv: vi.fn(),
  updateDiv: vi.fn(),
  deleteDiv: vi.fn(),
  assignTeamToDiv: vi.fn(),
  removeTeamFromDiv: vi.fn(),
  getDivStandings: vi.fn(),
  listEvaTeams: vi.fn(),
}))

import { useArchiveAccess } from '../hooks/useArchiveAccess'
import { listLiguesLocales, createLigueLocale, updateLigueLocale, deleteLigueLocale } from '../api/archives'

const mockUseArchiveAccess = useArchiveAccess as unknown as ReturnType<typeof vi.fn>

const mockLocales = [
  {
    id: 1,
    nom: 'Ligue Grenoble',
    emplacement: 'Grenoble',
    ligue_id: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  },
  {
    id: 2,
    nom: 'Ligue Lyon',
    emplacement: 'Lyon Est',
    ligue_id: null,
    created_at: '2026-02-01T00:00:00Z',
    updated_at: '2026-02-01T00:00:00Z',
  },
]

function createQueryWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/archives']}>{children}</MemoryRouter>
      </QueryClientProvider>
    )
  }
}

async function renderPage() {
  render(<ArchivePage />, { wrapper: createQueryWrapper() })
  await waitFor(() => {
    expect(screen.queryByText('Chargement...')).not.toBeInTheDocument()
  })
}

describe('ArchivePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseArchiveAccess.mockReturnValue({
      canWrite: true,
      canRead: true,
      isLoading: false,
      permissions: { can_read: true, can_write: true },
    })
    ;(listLiguesLocales as any).mockResolvedValue({ ligues_locales: mockLocales })
  })

  it('shows loading state', () => {
    mockUseArchiveAccess.mockReturnValue({
      canWrite: false,
      canRead: false,
      isLoading: true,
      permissions: undefined,
    })
    render(<ArchivePage />, { wrapper: createQueryWrapper() })
    expect(screen.getByText('Chargement...')).toBeInTheDocument()
  })

  it('shows empty state when no ligues locales', async () => {
    ;(listLiguesLocales as any).mockResolvedValue({ ligues_locales: [] })
    render(<ArchivePage />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText('Aucune ligue locale')).toBeInTheDocument()
    })
  })

  it('renders ligue locale table', async () => {
    await renderPage()
    expect(screen.getByText('Archives')).toBeInTheDocument()
    expect(screen.getByText('Ligue Grenoble')).toBeInTheDocument()
    expect(screen.getByText('Ligue Lyon')).toBeInTheDocument()
  })

  it('shows Nouvelle ligue locale button when canWrite is true', async () => {
    await renderPage()
    expect(screen.getByText('Nouvelle ligue locale')).toBeInTheDocument()
  })

  it('hides Nouvelle ligue locale button when canWrite is false', async () => {
    mockUseArchiveAccess.mockReturnValue({
      canWrite: false,
      canRead: true,
      isLoading: false,
      permissions: { can_read: true, can_write: false },
    })
    await renderPage()
    expect(screen.queryByText('Nouvelle ligue locale')).not.toBeInTheDocument()
  })

  it('opens create form modal', async () => {
    await renderPage()
    await userEvent.click(screen.getByRole('button', { name: 'Nouvelle ligue locale' }))
    expect(screen.getByText('Créer')).toBeInTheDocument()
    expect(screen.getByLabelText('Nom')).toBeInTheDocument()
    expect(screen.getByLabelText('Emplacement')).toBeInTheDocument()
  })

  it('calls createLigueLocale on form submit', async () => {
    (createLigueLocale as any).mockResolvedValue({ ligue_locale: mockLocales[0] })
    await renderPage()
    await userEvent.click(screen.getByRole('button', { name: 'Nouvelle ligue locale' }))
    await userEvent.type(screen.getByLabelText('Nom'), 'Nouvelle ligue locale test')
    await userEvent.type(screen.getByLabelText('Emplacement'), 'Test')
    await userEvent.click(screen.getByText('Créer'))
    await waitFor(() => {
      expect(createLigueLocale).toHaveBeenCalledWith({
        nom: 'Nouvelle ligue locale test',
        emplacement: 'Test',
      })
    })
  })

  it('opens edit form modal with pre-filled values', async () => {
    await renderPage()
    await userEvent.click(screen.getAllByText('Modifier')[0])
    expect(screen.getByText('Modifier la ligue locale')).toBeInTheDocument()
    const nomInput = screen.getByLabelText('Nom') as HTMLInputElement
    expect(nomInput.value).toBe('Ligue Grenoble')
  })

  it('calls updateLigueLocale on edit submit', async () => {
    (updateLigueLocale as any).mockResolvedValue({ ligue_locale: mockLocales[0] })
    await renderPage()
    await userEvent.click(screen.getAllByText('Modifier')[0])
    const nomInput = screen.getByLabelText('Nom')
    await userEvent.clear(nomInput)
    await userEvent.type(nomInput, 'Ligue locale modifiée')
    await userEvent.click(screen.getAllByText('Modifier')[2])
    await waitFor(() => {
      expect(updateLigueLocale).toHaveBeenCalledWith(1, {
        nom: 'Ligue locale modifiée',
        emplacement: 'Grenoble',
      })
    })
  })

  it('shows delete confirmation modal', async () => {
    await renderPage()
    await userEvent.click(screen.getAllByText('Supprimer')[0])
    expect(screen.getByText('Confirmer la suppression')).toBeInTheDocument()
    expect(screen.getByText('Êtes-vous sûr de vouloir supprimer cette ligue locale ?')).toBeInTheDocument()
  })

  it('calls deleteLigueLocale on delete confirm', async () => {
    (deleteLigueLocale as any).mockResolvedValue({ message: 'ok' })
    await renderPage()
    await userEvent.click(screen.getAllByText('Supprimer')[0])
    await userEvent.click(screen.getAllByText('Supprimer')[2])
    await waitFor(() => {
      expect(deleteLigueLocale).toHaveBeenCalledWith(1)
    })
  })

  it('hides action buttons when canWrite is false', async () => {
    mockUseArchiveAccess.mockReturnValue({
      canWrite: false,
      canRead: true,
      isLoading: false,
      permissions: { can_read: true, can_write: false },
    })
    await renderPage()
    expect(screen.queryByText('Modifier')).not.toBeInTheDocument()
    expect(screen.queryByText('Supprimer')).not.toBeInTheDocument()
  })

  it('displays error message', async () => {
    (createLigueLocale as any).mockRejectedValue({ body: { message: 'Erreur test' } })
    await renderPage()
    await userEvent.click(screen.getByRole('button', { name: 'Nouvelle ligue locale' }))
    await userEvent.type(screen.getByLabelText('Nom'), 'Test')
    await userEvent.type(screen.getByLabelText('Emplacement'), 'Test')
    await userEvent.click(screen.getByRole('button', { name: 'Créer' }))
    await waitFor(() => {
      expect(screen.getByText('Erreur test')).toBeInTheDocument()
    })
  })

  it('displays success message after create', async () => {
    (createLigueLocale as any).mockResolvedValue({ ligue_locale: mockLocales[0] })
    await renderPage()
    await userEvent.click(screen.getByRole('button', { name: 'Nouvelle ligue locale' }))
    await userEvent.type(screen.getByLabelText('Nom'), 'Test')
    await userEvent.type(screen.getByLabelText('Emplacement'), 'Test')
    await userEvent.click(screen.getByRole('button', { name: 'Créer' }))
    await waitFor(() => {
      expect(screen.getByText('Ligue locale créée')).toBeInTheDocument()
    })
  })

  it('closes modal on cancel', async () => {
    await renderPage()
    await userEvent.click(screen.getByRole('button', { name: 'Nouvelle ligue locale' }))
    expect(screen.getByRole('button', { name: 'Nouvelle ligue locale' })).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Annuler' }))
    await waitFor(() => {
      expect(screen.queryByText('Annuler')).not.toBeInTheDocument()
    })
  })
})
