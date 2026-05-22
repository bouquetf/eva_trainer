import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PublishButton } from './PublishButton'

const mockGetPublishStatus = vi.fn()
const mockPublishMap = vi.fn()
const mockUnpublishMap = vi.fn()
const mockListMyTeams = vi.fn()

vi.mock('../../api/templates', () => ({
  getPublishStatus: () => mockGetPublishStatus(),
  publishMap: () => mockPublishMap(),
  unpublishMap: () => mockUnpublishMap(),
}))

vi.mock('../../api/teams', () => ({
  listMyTeams: () => mockListMyTeams(),
}))

vi.mock('../../store/auth', () => ({
  useAuthStore: (selector: any) => selector({ user: { id: 1 } }),
}))

beforeEach(() => {
  vi.clearAllMocks()
})

describe('PublishButton', () => {
  it('affiche le badge "Publique" et bouton "Rendre privée" si publié (sync_active=true) et captain', async () => {
    mockGetPublishStatus.mockResolvedValue({
      published: true,
      template: {
        id: 1,
        original_team_id: 1,
        map_id: 1,
        sync_active: true,
        created_at: '2025-01-15T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      },
    })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'captain', member_count: 2 }])

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Publique')).toBeInTheDocument()
    })
    expect(screen.getByText('Rendre privée')).toBeInTheDocument()
  })

  it('affiche badge "Publique" sans bouton si publié et rôle editor', async () => {
    mockGetPublishStatus.mockResolvedValue({
      published: true,
      template: {
        id: 1,
        original_team_id: 1,
        map_id: 1,
        sync_active: true,
        created_at: '2025-01-15T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      },
    })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'editor', member_count: 2 }])

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Publique')).toBeInTheDocument()
    })
    expect(screen.queryByText('Rendre privée')).not.toBeInTheDocument()
  })

  it('affiche badge "Privée" et bouton "Rendre publique" si dépublié (sync_active=false) et captain', async () => {
    mockGetPublishStatus.mockResolvedValue({
      published: true,
      template: {
        id: 1,
        original_team_id: 1,
        map_id: 1,
        sync_active: false,
        created_at: '2025-01-15T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      },
    })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'captain', member_count: 2 }])

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Privée')).toBeInTheDocument()
    })
    expect(screen.getByText('Rendre publique')).toBeInTheDocument()
  })

  it('affiche badge "Privée" si dépublié et rôle member', async () => {
    mockGetPublishStatus.mockResolvedValue({
      published: true,
      template: {
        id: 1,
        original_team_id: 1,
        map_id: 1,
        sync_active: false,
        created_at: '2025-01-15T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      },
    })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'member', member_count: 2 }])

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Privée')).toBeInTheDocument()
    })
    expect(screen.queryByText('Rendre publique')).not.toBeInTheDocument()
  })

  it('affiche le bouton "Rendre publique" si non publié et rôle captain', async () => {
    mockGetPublishStatus.mockResolvedValue({ published: false })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'captain', member_count: 2 }])

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Rendre publique')).toBeInTheDocument()
    })
  })

  it('affiche le badge "Privée" si non publié et rôle editor', async () => {
    mockGetPublishStatus.mockResolvedValue({ published: false })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'editor', member_count: 2 }])

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Privée')).toBeInTheDocument()
    })
    expect(screen.queryByText('Rendre publique')).not.toBeInTheDocument()
  })

  it('confirmation → publication → succès', async () => {
    const user = userEvent.setup()
    mockGetPublishStatus.mockResolvedValue({ published: false })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'captain', member_count: 2 }])
    mockPublishMap.mockResolvedValue({
      created: true,
      template: {
        id: 1,
        original_team_id: 1,
        map_id: 1,
        sync_active: true,
        created_at: '2025-01-15T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      },
    })

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Rendre publique')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Rendre publique'))

    await waitFor(() => {
      expect(screen.getByText('Rendre la carte publique ?')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Rendre publique', { selector: 'button.btn-danger' }))

    await waitFor(() => {
      expect(mockPublishMap).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(screen.getByText('Carte publiée avec succès !')).toBeInTheDocument()
    })
  })

  it('confirmation → unpublish → succès', async () => {
    const user = userEvent.setup()
    mockGetPublishStatus.mockResolvedValue({
      published: true,
      template: {
        id: 1,
        original_team_id: 1,
        map_id: 1,
        sync_active: true,
        created_at: '2025-01-15T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      },
    })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'captain', member_count: 2 }])

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Rendre privée')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Rendre privée'))

    await waitFor(() => {
      expect(screen.getByText('Rendre la carte privée ?')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Rendre privée', { selector: 'button.btn-danger' }))

    await waitFor(() => {
      expect(mockUnpublishMap).toHaveBeenCalled()
    })
  })

  it('confirmation → annulation → pas d action', async () => {
    const user = userEvent.setup()
    mockGetPublishStatus.mockResolvedValue({ published: false })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'captain', member_count: 2 }])

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Rendre publique')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Rendre publique'))

    await waitFor(() => {
      expect(screen.getByText('Annuler')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Annuler'))

    expect(mockPublishMap).not.toHaveBeenCalled()

    await waitFor(() => {
      expect(screen.getByText('Rendre publique')).toBeInTheDocument()
    })
  })

  it('re-publication avec created=false affiche message de remplacement', async () => {
    const user = userEvent.setup()
    mockGetPublishStatus.mockResolvedValue({
      published: true,
      template: {
        id: 1,
        original_team_id: 1,
        map_id: 1,
        sync_active: false,
        created_at: '2025-01-15T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      },
    })
    mockListMyTeams.mockResolvedValue([{ team: { id: 1 }, my_role: 'captain', member_count: 2 }])
    mockPublishMap.mockResolvedValue({
      created: false,
      template: {
        id: 1,
        original_team_id: 1,
        map_id: 1,
        sync_active: true,
        created_at: '2025-01-15T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      },
    })

    render(<PublishButton teamId={1} mapId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Rendre publique')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Rendre publique'))

    await waitFor(() => {
      expect(screen.getByText('Rendre la carte publique ?')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Rendre publique', { selector: 'button.btn-danger' }))

    await waitFor(() => {
      expect(screen.getByText('Le template a été remplacé avec les zones actuelles.')).toBeInTheDocument()
    })
  })
})
