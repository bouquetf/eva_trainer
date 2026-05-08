import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MemberList from './MemberList'
import type { TeamMemberWithUser } from '../api/roles'

vi.mock('../api/client', () => ({
  getApiClient: () => ({
    put: vi.fn(),
  }),
}))

vi.mock('../store/auth', () => ({
  useAuthStore: {
    getState: () => ({
      user: { id: 1, discord_id: 'test', discord_username: '', display_name: '', is_admin: false, created_at: '2024-01-01T00:00:00Z' },
    }),
  },
}))

function createQueryWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }
}

const mockMembers: TeamMemberWithUser[] = [
  { user_id: 1, display_name: 'Alice', role: 'captain', joined_at: '2024-01-01T00:00:00Z' },
  { user_id: 2, display_name: 'Bob', role: 'editor', joined_at: '2024-01-02T00:00:00Z' },
  { user_id: 3, display_name: 'Charlie', role: 'member', joined_at: '2024-01-03T00:00:00Z' },
]

describe('MemberList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('displays members with their role badges', () => {
    render(
      <MemberList
        teamId={1}
        members={mockMembers}
        currentUserId={1}
        currentUserRole="captain"
        onRoleChanged={() => {}}
      />,
      { wrapper: createQueryWrapper() },
    )

    expect(screen.getByText('Capitaine')).toBeInTheDocument()
    expect(screen.getByText('Éditeur')).toBeInTheDocument()
    expect(screen.getByText('Membre')).toBeInTheDocument()
  })

  it('shows action buttons for captain', () => {
    render(
      <MemberList
        teamId={1}
        members={mockMembers}
        currentUserId={1}
        currentUserRole="captain"
        onRoleChanged={() => {}}
      />,
      { wrapper: createQueryWrapper() },
    )

    expect(screen.getByText('👤 Membre')).toBeInTheDocument()
    expect(screen.getAllByText('⚓ Cocapitaine')).toHaveLength(2)
  })

  it('does not show promote cocaptain for cocaptain viewer', () => {
    render(
      <MemberList
        teamId={1}
        members={mockMembers}
        currentUserId={4}
        currentUserRole="cocaptain"
        onRoleChanged={() => {}}
      />,
      { wrapper: createQueryWrapper() },
    )

    expect(screen.queryByText('⚓ Cocapitaine')).not.toBeInTheDocument()
  })
})
