import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import InviteAccept from './InviteAccept'

vi.mock('../api/invitations', () => ({
  getInviteLinkInfo: vi.fn(),
  acceptInviteLink: vi.fn(),
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useParams: () => ({ token: 'test-token' }) }
})

import { getInviteLinkInfo, acceptInviteLink } from '../api/invitations'

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

describe('InviteAccept', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading spinner initially', () => {
    ;(getInviteLinkInfo as any).mockReturnValue(new Promise(() => {}))
    const { container } = render(<InviteAccept />, { wrapper: createQueryWrapper() })
    expect(container.querySelector('.spinner')).toBeInTheDocument()
  })

  it('shows error card when link is not found', async () => {
    ;(getInviteLinkInfo as any).mockRejectedValue(new Error('Not found'))
    render(<InviteAccept />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText('Lien invalide')).toBeInTheDocument()
    })
  })

  it('shows expired message when link is expired', async () => {
    ;(getInviteLinkInfo as any).mockResolvedValue({
      is_expired: true,
      team_name: 'Test Team',
      expires_at: '2026-01-01T00:00:00Z',
    })
    render(<InviteAccept />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText('Lien expiré')).toBeInTheDocument()
    })
  })

  it('shows invite card with team name', async () => {
    ;(getInviteLinkInfo as any).mockResolvedValue({
      is_expired: false,
      team_name: 'Test Team',
      expires_at: '2026-12-31T00:00:00Z',
    })
    render(<InviteAccept />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText(/Test Team/)).toBeInTheDocument()
    })
  })

  it('calls accept on button click', async () => {
    ;(getInviteLinkInfo as any).mockResolvedValue({
      is_expired: false,
      team_name: 'Test Team',
      expires_at: '2026-12-31T00:00:00Z',
    })
    ;(acceptInviteLink as any).mockResolvedValue({ team_name: 'Test Team' })
    render(<InviteAccept />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText("Accepter l'invitation")).toBeInTheDocument()
    })
    await userEvent.click(screen.getByText("Accepter l'invitation"))
    expect(acceptInviteLink).toHaveBeenCalledWith('test-token')
  })

  it('shows success message after accepting', async () => {
    ;(getInviteLinkInfo as any).mockResolvedValue({
      is_expired: false,
      team_name: 'Test Team',
      expires_at: '2026-12-31T00:00:00Z',
    })
    ;(acceptInviteLink as any).mockResolvedValue({ team_name: 'Test Team' })
    render(<InviteAccept />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText("Accepter l'invitation")).toBeInTheDocument()
    })
    await userEvent.click(screen.getByText("Accepter l'invitation"))
    await waitFor(() => {
      expect(screen.getByText(/Vous avez rejoint/)).toBeInTheDocument()
      expect(screen.getByText(/Test Team/)).toBeInTheDocument()
    })
  })

  it('shows error message when accept fails', async () => {
    ;(getInviteLinkInfo as any).mockResolvedValue({
      is_expired: false,
      team_name: 'Test Team',
      expires_at: '2026-12-31T00:00:00Z',
    })
    ;(acceptInviteLink as any).mockRejectedValue({
      message: "Vous n'êtes pas autorisé",
    })
    render(<InviteAccept />, { wrapper: createQueryWrapper() })
    await waitFor(() => {
      expect(screen.getByText("Accepter l'invitation")).toBeInTheDocument()
    })
    await userEvent.click(screen.getByText("Accepter l'invitation"))
    await waitFor(() => {
      expect(screen.getByText("Vous n'êtes pas autorisé")).toBeInTheDocument()
    })
  })
})
