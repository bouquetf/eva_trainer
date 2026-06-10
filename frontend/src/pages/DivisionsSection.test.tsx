import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import DivisionsSection from './DivisionsSection'

vi.mock('../api/archives', () => ({
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
  generateRoundRobin: vi.fn(),
  setDivReplayURL: vi.fn(),
  deleteDivReplayURL: vi.fn(),
}))

import {
  listLigues, listDivs, getDiv, createDiv, updateDiv, deleteDiv,
  getDivStandings, listEvaTeams, generateRoundRobin,
  setDivReplayURL, deleteDivReplayURL,
} from '../api/archives'

const mockLigues = [
  { id: 1, date_debut: '2026-01-01', date_fin: '2026-06-30', split_name: 'Split 1 2026', nb_locales: 2 },
  { id: 2, date_debut: '2026-07-01', date_fin: '2026-12-31', split_name: 'Split 2 2026', nb_locales: 1 },
]

const mockTeams = [
  { id: 1, name: 'Team Alpha', tag: 'TA', eva_trainer_team_id: null, created_at: '', updated_at: '' },
  { id: 2, name: 'Team Beta', tag: 'TB', eva_trainer_team_id: null, created_at: '', updated_at: '' },
]

const mockDivs = [
  { id: 1, ligue_id: 1, nom: 'Groupe A', replay_url: '', teams: [mockTeams[0]] },
  { id: 2, ligue_id: 1, nom: 'Groupe B', replay_url: '', teams: [] },
]

const mockStandings = [
  { team_id: 1, team_name: 'Team Alpha', team_tag: 'TA', bos_played: 2, bo_won: 2, bo_drawn: 0, bo_lost: 0, bo_points: 2, matches_played: 6, match_won: 5, match_drawn: 1, match_lost: 0, match_points: 5 },
  { team_id: 2, team_name: 'Team Beta', team_tag: 'TB', bos_played: 2, bo_won: 0, bo_drawn: 0, bo_lost: 2, bo_points: 0, matches_played: 6, match_won: 0, match_drawn: 1, match_lost: 5, match_points: 0 },
]

function createQueryWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    )
  }
}

async function selectLigue() {
  await waitFor(() => {
    expect(screen.getByText(/Split 1 2026/)).toBeInTheDocument()
  })
  await userEvent.selectOptions(screen.getByLabelText('Période'), '1')
}

describe('DivisionsSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(listLigues as any).mockResolvedValue({ items: mockLigues, total: 2, page: 1, per_page: 100 })
    ;(listEvaTeams as any).mockResolvedValue({ teams: mockTeams })
    ;(listDivs as any).mockResolvedValue({ divs: mockDivs })
    ;(getDiv as any).mockResolvedValue({ div: mockDivs[0], teams: [mockTeams[0]] })
    ;(getDivStandings as any).mockResolvedValue({ standings: mockStandings })
  })

  it('shows ligue selector with ligue options', async () => {
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText(/Split 1 2026/)).toBeInTheDocument()
    })
  })

  it('does not show divs list when no ligue selected', async () => {
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await waitFor(() => expect(screen.getByText(/Split 1 2026/)).toBeInTheDocument())
    expect(screen.queryByText('Groupe A')).not.toBeInTheDocument()
  })

  it('lists divs when a ligue is selected', async () => {
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => {
      expect(screen.getByText('Groupe A')).toBeInTheDocument()
      expect(screen.getByText('Groupe B')).toBeInTheDocument()
    })
  })

  it('opens create div form', async () => {
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => {
      expect(screen.getByText('Nouvelle division')).toBeInTheDocument()
    })
    await userEvent.click(screen.getByText('Nouvelle division'))
    expect(screen.getByText('Créer')).toBeInTheDocument()
    expect(screen.getByLabelText('Nom')).toBeInTheDocument()
  })

  it('creates a div', async () => {
    (createDiv as any).mockResolvedValue({ div: mockDivs[0] })
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getByText('Nouvelle division')).toBeInTheDocument())
    await userEvent.click(screen.getByText('Nouvelle division'))
    await userEvent.type(screen.getByLabelText('Nom'), 'Groupe C')
    await userEvent.click(screen.getByText('Créer'))
    await waitFor(() => {
      expect(createDiv).toHaveBeenCalledWith({ ligue_id: 1, nom: 'Groupe C' })
    })
  })

  it('edits a div', async () => {
    (updateDiv as any).mockResolvedValue({ div: mockDivs[0] })
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getByText('Groupe A')).toBeInTheDocument())

    const groupeARow = screen.getByText('Groupe A').closest('tr')!
    await userEvent.click(within(groupeARow).getByText('Modifier'))

    await waitFor(() => {
      expect(screen.getByText('Modifier la division')).toBeInTheDocument()
    })
    const nameInput = screen.getByLabelText('Nom') as HTMLInputElement
    expect(nameInput.value).toBe('Groupe A')
    await userEvent.clear(nameInput)
    await userEvent.type(nameInput, 'Groupe A modifié')
    await userEvent.click(screen.getAllByText('Modifier')[2])
    await waitFor(() => {
      expect(updateDiv).toHaveBeenCalledWith(1, { nom: 'Groupe A modifié' })
    })
  })

  it('deletes a div with confirmation', async () => {
    (deleteDiv as any).mockResolvedValue({ message: 'ok' })
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getAllByText('Supprimer')[0]).toBeInTheDocument())
    await userEvent.click(screen.getAllByText('Supprimer')[0])
    expect(screen.getByText('Confirmer la suppression')).toBeInTheDocument()
    await userEvent.click(screen.getAllByText('Supprimer')[2])
    await waitFor(() => {
      expect(deleteDiv).toHaveBeenCalledWith(1)
    })
  })

  it('shows team tags in Équipes column without clicking', async () => {
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getByText('Groupe A')).toBeInTheDocument())
    await waitFor(() => {
      expect(screen.getByText(/Team Alpha \(TA\)/)).toBeInTheDocument()
    })
    const rows = screen.getAllByRole('row')
    const groupeARow = rows.find(r => r.textContent?.includes('Groupe A'))
    expect(groupeARow).toBeDefined()
    expect(groupeARow!.textContent).toContain('Team Alpha (TA)')
    expect(groupeARow!.textContent).not.toContain('Team Beta')
  })

  it('shows teams when clicking a div', async () => {
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getByText('Groupe A')).toBeInTheDocument())
    await userEvent.click(screen.getByText('Groupe A'))
    await waitFor(() => {
      expect(screen.getAllByText('Team Alpha').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('shows assign team form', async () => {
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getByText('Groupe A')).toBeInTheDocument())
    await userEvent.click(screen.getByText('Groupe A'))
    await waitFor(() => {
      expect(screen.getAllByText('Team Alpha').length).toBeGreaterThanOrEqual(1)
    })
    expect(screen.getByText('Assigner une équipe')).toBeInTheDocument()
  })

  it('hides write buttons when canWrite is false', async () => {
    render(<DivisionsSection canWrite={false} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => {
      expect(screen.queryByText('Nouvelle division')).not.toBeInTheDocument()
    })
  })

  it('shows standings when selecting a div', async () => {
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getByText('Groupe A')).toBeInTheDocument())
    await userEvent.click(screen.getByText('Groupe A'))
    await waitFor(() => {
      expect(screen.getAllByText('Team Alpha').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('Team Beta').length).toBeGreaterThanOrEqual(1)
      expect(screen.getByText(/Classement/)).toBeInTheDocument()
    })
  })

  it('shows generate round-robin button when div is expanded and canWrite is true', async () => {
    (generateRoundRobin as any).mockResolvedValue({ bos: [{ id: 1, equipe1_id: 1, equipe2_id: 2 }] })
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getByText('Groupe A')).toBeInTheDocument())
    await userEvent.click(screen.getByText('Groupe A'))
    await waitFor(() => {
      expect(screen.getByText('Générer le calendrier')).toBeInTheDocument()
    })
  })

  it('calls generateRoundRobin when button clicked and shows success', async () => {
    (generateRoundRobin as any).mockResolvedValue({ bos: [{ id: 1, equipe1_id: 1, equipe2_id: 2 }] })
    render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getByText('Groupe A')).toBeInTheDocument())
    await userEvent.click(screen.getByText('Groupe A'))
    await waitFor(() => expect(screen.getByText('Générer le calendrier')).toBeInTheDocument())
    await userEvent.click(screen.getByText('Générer le calendrier'))
    await waitFor(() => {
      expect(generateRoundRobin).toHaveBeenCalledWith(1)
      expect(screen.getByText(/Calendrier généré/)).toBeInTheDocument()
    })
  })

  it('does not show generate round-robin button when canWrite is false', async () => {
    render(<DivisionsSection canWrite={false} />, { wrapper: createQueryWrapper() })
    await selectLigue()
    await waitFor(() => expect(screen.getByText('Groupe A')).toBeInTheDocument())
    await userEvent.click(screen.getByText('Groupe A'))
    expect(screen.queryByText('Générer le calendrier')).not.toBeInTheDocument()
  })

  describe('replay URL', () => {
    beforeEach(() => {
      vi.clearAllMocks()
      ;(listLigues as any).mockResolvedValue({ items: mockLigues, total: 2, page: 1, per_page: 100 })
      ;(listEvaTeams as any).mockResolvedValue({ teams: mockTeams })
      ;(listDivs as any).mockResolvedValue({ divs: mockDivs })
      ;(getDivStandings as any).mockResolvedValue({ standings: mockStandings })
    })

    it('shows "Aucun lien replay" when replay_url is empty', async () => {
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: '' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => {
        expect(screen.getByText('Aucun lien replay')).toBeInTheDocument()
      })
    })

    it('shows "Ajouter un lien replay" button when canWrite=true and replay_url empty', async () => {
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: '' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => {
        expect(screen.getByText('Ajouter un lien replay')).toBeInTheDocument()
      })
    })

    it('does not show "Ajouter un lien replay" when canWrite=false', async () => {
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: '' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={false} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => {
        expect(screen.queryByText('Ajouter un lien replay')).not.toBeInTheDocument()
      })
    })

    it('shows YouTube embed iframe for a YouTube link', async () => {
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: 'https://www.youtube.com/watch?v=test123' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => {
        const iframe = document.querySelector('iframe')
        expect(iframe).toBeInTheDocument()
        expect(iframe?.src).toContain('youtube.com/embed/test123')
      })
    })

    it('shows "Regarder sur Twitch" link for a Twitch link', async () => {
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: 'https://www.twitch.tv/videos/123456' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => {
        const link = screen.getByText('Regarder sur Twitch')
        expect(link).toBeInTheDocument()
        expect(link).toHaveAttribute('href', 'https://www.twitch.tv/videos/123456')
      })
    })

    it('shows Modifier/Supprimer buttons for canWrite=true with existing link', async () => {
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: 'https://www.youtube.com/watch?v=test123' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => {
        expect(screen.getAllByText('Modifier').length).toBeGreaterThanOrEqual(2)
        expect(screen.getAllByText('Supprimer').length).toBeGreaterThanOrEqual(2)
      })
    })

    it('hides Modifier/Supprimer when canWrite=false but still shows iframe', async () => {
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: 'https://www.youtube.com/watch?v=test123' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={false} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => {
        const iframe = document.querySelector('iframe')
        expect(iframe).toBeInTheDocument()
        expect(screen.queryByText('Modifier')).not.toBeInTheDocument()
      })
    })

    it('opens add replay form when clicking "Ajouter un lien replay"', async () => {
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: '' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => expect(screen.getByText('Ajouter un lien replay')).toBeInTheDocument())
      await userEvent.click(screen.getByText('Ajouter un lien replay'))
      expect(screen.getByPlaceholderText('https://www.youtube.com/watch?v=...')).toBeInTheDocument()
      expect(screen.getByText('Valider')).toBeInTheDocument()
    })

    it('calls setDivReplayURL on form submit', async () => {
      ;(setDivReplayURL as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: 'https://www.youtube.com/watch?v=test' } })
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: '' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => expect(screen.getByText('Ajouter un lien replay')).toBeInTheDocument())
      await userEvent.click(screen.getByText('Ajouter un lien replay'))
      await userEvent.type(screen.getByPlaceholderText('https://www.youtube.com/watch?v=...'), 'https://www.youtube.com/watch?v=test')
      await userEvent.click(screen.getByText('Valider'))
      await waitFor(() => {
        expect(setDivReplayURL).toHaveBeenCalledWith(1, 'https://www.youtube.com/watch?v=test')
      })
    })

    it('shows confirmation modal before deletion and confirms', async () => {
      ;(deleteDivReplayURL as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: '' } })
      ;(getDiv as any).mockResolvedValue({ div: { ...mockDivs[0], replay_url: 'https://www.youtube.com/watch?v=test' }, teams: [mockTeams[0]] })
      render(<DivisionsSection canWrite={true} />, { wrapper: createQueryWrapper() })
      await selectLigue()
      await userEvent.click(screen.getByText('Groupe A'))
      await waitFor(() => expect(screen.getByText('Replay stream')).toBeInTheDocument())
      const detailCell = screen.getByText('Replay stream').closest('.detail-cell')!
      const replayDeleteBtn = within(detailCell).getByText('Supprimer')
      await userEvent.click(replayDeleteBtn)
      await waitFor(() => {
        expect(screen.getByText('Confirmer la suppression')).toBeInTheDocument()
      })
      const confirmBtn = screen.getByText('Supprimer', { selector: 'button:not(.btn-small)' })
      await userEvent.click(confirmBtn)
      await waitFor(() => {
        expect(deleteDivReplayURL).toHaveBeenCalledWith(1)
      })
    })
  })
})
