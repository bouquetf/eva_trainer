import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, within, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import BOsSection from './BOsSection'

vi.mock('../api/archives', () => ({
  listLigues: vi.fn(),
  listBOs: vi.fn(),
  getBO: vi.fn(),
  createBO: vi.fn(),
  updateBO: vi.fn(),
  deleteBO: vi.fn(),
  listEvaTeams: vi.fn(),
  getLigueMaps: vi.fn(),
}))

import { listLigues, listBOs, getBO, createBO, listEvaTeams, getLigueMaps } from '../api/archives'

const mockLigues = [
  { id: 1, nom: 'Saison 2026 S1', date_debut: '2026-01-01', date_fin: '2026-06-30', split_id: 1, split_name: 'Split 1 2026', nb_locales: 0 },
]
const mockTeams = [
  { id: 1, name: 'Team Alpha', tag: 'TA', eva_trainer_team_id: null, created_at: '', updated_at: '' },
  { id: 2, name: 'Team Beta', tag: 'TB', eva_trainer_team_id: null, created_at: '', updated_at: '' },
]
const mockMaps = [
  { id: 1, name: 'Map Alpha', image: '', created_at: '', updated_at: '' },
  { id: 2, name: 'Map Beta', image: '', created_at: '', updated_at: '' },
]

function createQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
}

function Wrapper({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={createQueryClient()}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  )
}

describe('BOsSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(listLigues as any).mockResolvedValue({ items: mockLigues, total: 1, page: 1, per_page: 100 })
    ;(listEvaTeams as any).mockResolvedValue({ teams: mockTeams })
    ;(getLigueMaps as any).mockResolvedValue({ maps: mockMaps })
    ;(listBOs as any).mockResolvedValue({ bos: [], total: 0 })
  })

  it('renders period selector', async () => {
    render(<BOsSection canWrite={true} />, { wrapper: Wrapper })
    await screen.findByText(/Split 1 2026/)
    expect(screen.getByLabelText('Période')).toBeInTheDocument()
  })

  it('shows Nouveau BO button when canWrite is true', async () => {
    render(<BOsSection canWrite={true} />, { wrapper: Wrapper })
    await screen.findByText(/Split 1 2026/)
    await userEvent.selectOptions(screen.getByLabelText('Période'), '1')
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Nouveau BO' })).toBeInTheDocument()
    })
  })

  it('opens creation form on Nouveau BO click', async () => {
    render(<BOsSection canWrite={true} />, { wrapper: Wrapper })
    await screen.findByText(/Split 1 2026/)
    await userEvent.selectOptions(screen.getByLabelText('Période'), '1')
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Nouveau BO' })).toBeInTheDocument()
    })
    await userEvent.click(screen.getByRole('button', { name: 'Nouveau BO' }))
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Nouveau BO' })).toBeInTheDocument()
    })
  })

  describe('addPick / addMatch découplage', () => {
    beforeEach(async () => {
      render(<BOsSection canWrite={true} />, { wrapper: Wrapper })
      await screen.findByText(/Split 1 2026/)
      await userEvent.selectOptions(screen.getByLabelText('Période'), '1')
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Nouveau BO' })).toBeInTheDocument()
      })
      await userEvent.click(screen.getByRole('button', { name: 'Nouveau BO' }))
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Nouveau BO' })).toBeInTheDocument()
      })
    })

    function getPicksGroup() {
      return screen.getByText('Picks').closest('.form-group') as HTMLElement
    }

    function getMatchsGroup() {
      return screen.getByText('Matchs').closest('.form-group') as HTMLElement
    }

    it('addPick n\'ajoute qu\'un pick sans créer de match', async () => {
      await userEvent.click(within(getPicksGroup()).getByRole('button', { name: '+ Ajouter un pick' }))
      const pickRows = getPicksGroup().querySelectorAll('.add-membership-row')
      expect(pickRows.length).toBe(1)
      const matchCards = getMatchsGroup().querySelectorAll('[style*="border"]')
      expect(matchCards.length).toBe(0)
    })

    it('addMatch auto-remplit la map depuis le pick au même index', async () => {
      await userEvent.click(within(getPicksGroup()).getByRole('button', { name: '+ Ajouter un pick' }))
      const pickSelect = getPicksGroup().querySelectorAll('.add-membership-row select')[1] as HTMLSelectElement
      await userEvent.selectOptions(pickSelect, '2')
      await userEvent.click(within(getMatchsGroup()).getByRole('button', { name: '+ Ajouter un match' }))
      const matchSelect = getMatchsGroup().querySelector('.add-membership-row select') as HTMLSelectElement
      await waitFor(() => {
        expect(matchSelect.value).toBe('2')
      })
    })

    it('addMatch n\'auto-remplit pas si aucun pick au même index', async () => {
      await userEvent.click(within(getMatchsGroup()).getByRole('button', { name: '+ Ajouter un match' }))
      const matchSelect = getMatchsGroup().querySelector('.add-membership-row select') as HTMLSelectElement
      expect(matchSelect.value).toBe('')
    })
  })

  describe('groupement par division', () => {
    it('affiche les BOs groupés par division', async () => {
      ;(listBOs as any).mockResolvedValue({
        bos: [
          { id: 1, ligue_id: 1, equipe1_id: 1, equipe2_id: 2, equipe1_name: 'Team Alpha', equipe2_name: 'Team Beta', format: 'BO3', vainqueur_id: null, created_at: '2026-03-01T00:00:00Z', forfait: 0, div_id: 10, div_name: 'Groupe A' },
          { id: 2, ligue_id: 1, equipe1_id: 2, equipe2_id: 1, equipe1_name: 'Team Beta', equipe2_name: 'Team Alpha', format: 'BO3', vainqueur_id: null, created_at: '2026-03-02T00:00:00Z', forfait: 0, div_id: 20, div_name: 'Groupe B' },
        ],
        total: 2,
      })
      render(<BOsSection canWrite={true} />, { wrapper: Wrapper })
      await screen.findByText(/Split 1 2026/)
      await userEvent.selectOptions(screen.getByLabelText('Période'), '1')
      await waitFor(() => {
        expect(screen.getByText('Groupe A')).toBeInTheDocument()
        expect(screen.getByText('Groupe B')).toBeInTheDocument()
      })
    })

    it('met les BOs sans division dans Non assigné', async () => {
      ;(listBOs as any).mockResolvedValue({
        bos: [{ id: 1, ligue_id: 1, equipe1_id: 1, equipe2_id: 2, equipe1_name: 'Team Alpha', equipe2_name: 'Team Beta', format: 'BO3', vainqueur_id: null, created_at: '2026-03-01T00:00:00Z', forfait: 0, div_id: null, div_name: '' }],
        total: 1,
      })
      render(<BOsSection canWrite={true} />, { wrapper: Wrapper })
      await screen.findByText(/Split 1 2026/)
      await userEvent.selectOptions(screen.getByLabelText('Période'), '1')
      await waitFor(() => {
        expect(screen.getByText('Non assigné')).toBeInTheDocument()
      })
    })
  })

  describe('hors_compétition', () => {
    beforeEach(async () => {
      render(<BOsSection canWrite={true} />, { wrapper: Wrapper })
      await screen.findByText(/Split 1 2026/)
      await userEvent.selectOptions(screen.getByLabelText('Période'), '1')
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Nouveau BO' })).toBeInTheDocument()
      })
      await userEvent.click(screen.getByRole('button', { name: 'Nouveau BO' }))
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Nouveau BO' })).toBeInTheDocument()
      })
    })

    it('affiche la checkbox Hors compétition dans le formulaire match', async () => {
      const matchsGroup = screen.getByText('Matchs').closest('.form-group') as HTMLElement
      await userEvent.click(within(matchsGroup).getByRole('button', { name: '+ Ajouter un match' }))
      expect(screen.getByText('Hors compétition')).toBeInTheDocument()
    })

    it('envoie hors_competition dans le payload createBO', async () => {
      ;(createBO as any).mockResolvedValue({ bo: { id: 1 } })
      const matchsGroup = screen.getByText('Matchs').closest('.form-group') as HTMLElement
      await userEvent.click(within(matchsGroup).getByRole('button', { name: '+ Ajouter un match' }))
      await userEvent.selectOptions(screen.getByLabelText('Équipe 1'), '1')
      await userEvent.selectOptions(screen.getByLabelText('Équipe 2'), '2')
      const matchCard = matchsGroup.querySelector('[style*="border"]') as HTMLElement
      const checkbox = within(matchCard).getByRole('checkbox')
      await userEvent.click(checkbox)
      const mapSelect = matchCard.querySelector('select') as HTMLSelectElement
      await userEvent.selectOptions(mapSelect, '1')
      const score1 = matchCard.querySelector('input[type="number"]') as HTMLInputElement
      await userEvent.clear(score1)
      await userEvent.type(score1, '3')
      const scoreInputs = matchCard.querySelectorAll('input[type="number"]')
      await userEvent.clear(scoreInputs[1])
      await userEvent.type(scoreInputs[1], '1')
      const form = document.querySelector('form')!
      fireEvent.submit(form)
      await waitFor(() => {
        expect(createBO).toHaveBeenCalled()
      })
      const payload = (createBO as any).mock.calls[0][0]
      expect(payload.matchs[0].hors_competition).toBe(true)
    })
  })

  describe('badge hors_competition detail', () => {
    it('affiche le badge Hors compétition dans la vue détail', async () => {
      ;(getBO as any).mockResolvedValue({
        bo: {
          BO: { id: 1, equipe1_id: 1, equipe2_id: 2, format: 'BO3', vainqueur_id: 1, forfait: 0 },
          bans: [], picks: [],
          matchs: [{ id: 1, map_id: 1, score_equipe1: 3, score_equipe2: 1, forfait: 0, commentaires: '', vainqueur_id: 1, hors_competition: true }],
        },
      })
      ;(listBOs as any).mockResolvedValue({
        bos: [{ id: 1, ligue_id: 1, equipe1_id: 1, equipe2_id: 2, equipe1_name: 'Team Alpha', equipe2_name: 'Team Beta', format: 'BO3', vainqueur_id: 1, created_at: '2026-03-01T00:00:00Z', forfait: 0, div_id: null, div_name: '' }],
        total: 1,
      })
      render(<BOsSection canWrite={true} />, { wrapper: Wrapper })
      await screen.findByText(/Split 1 2026/)
      await userEvent.selectOptions(screen.getByLabelText('Période'), '1')
      await waitFor(() => {
        expect(screen.getAllByText('Team Alpha').length).toBeGreaterThan(0)
      })
      await userEvent.click(screen.getAllByText('Team Alpha')[0])
      await waitFor(() => {
        expect(screen.getByText('Détails du BO')).toBeInTheDocument()
      })
      expect(screen.getByText('Hors compétition')).toBeInTheDocument()
    })
  })
})
