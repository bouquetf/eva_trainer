import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import TransferCaptaincyDialog from './TransferCaptaincyDialog'
import type { TeamMemberWithUser } from '../api/roles'

vi.mock('../store/auth', () => ({
  useAuthStore: { getState: () => ({ user: { id: 1 } }) },
}))

const transferCaptaincyMock = vi.fn()
vi.mock('../api/roles', () => ({
  transferCaptaincy: (...args: unknown[]) => transferCaptaincyMock(...args),
}))

function createQueryWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

const mockMembers: TeamMemberWithUser[] = [
  { user_id: 1, display_name: 'Alice', role: 'captain', joined_at: '2024-01-01T00:00:00Z' },
  { user_id: 2, display_name: 'Bob', role: 'cocaptain', joined_at: '2024-01-02T00:00:00Z' },
  { user_id: 3, display_name: 'Charlie', role: 'editor', joined_at: '2024-01-03T00:00:00Z' },
]

const defaultProps = {
  isOpen: true,
  teamId: 1,
  members: mockMembers,
  onClose: vi.fn(),
  onTransferred: vi.fn(),
}

describe('TransferCaptaincyDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when closed', () => {
    const { container } = render(
      <TransferCaptaincyDialog {...defaultProps} isOpen={false} />,
      { wrapper: createQueryWrapper() },
    )
    expect(container.innerHTML).toBe('')
  })

  it('renders dialog when open', () => {
    render(<TransferCaptaincyDialog {...defaultProps} />, { wrapper: createQueryWrapper() })
    expect(screen.getByText('Transférer la capitainerie')).toBeInTheDocument()
  })

  it('shows member options in select', () => {
    render(<TransferCaptaincyDialog {...defaultProps} />, { wrapper: createQueryWrapper() })
    expect(screen.getByText('Bob (cocaptain)')).toBeInTheDocument()
    expect(screen.getByText('Charlie (editor)')).toBeInTheDocument()
  })

  it('excludes current user from options', () => {
    render(<TransferCaptaincyDialog {...defaultProps} />, { wrapper: createQueryWrapper() })
    expect(screen.queryByText('Alice (captain)')).not.toBeInTheDocument()
  })

  it('shows warning when member selected', async () => {
    const user = userEvent.setup()
    render(<TransferCaptaincyDialog {...defaultProps} />, { wrapper: createQueryWrapper() })
    const select = screen.getByRole('combobox')
    await user.selectOptions(select, '2')
    expect(screen.getByText(/Vous allez transférer votre rôle de capitaine à Bob/)).toBeInTheDocument()
  })

  it('confirm button disabled when no member selected', () => {
    render(<TransferCaptaincyDialog {...defaultProps} />, { wrapper: createQueryWrapper() })
    const confirmBtn = screen.getByText('Confirmer le transfert')
    expect(confirmBtn).toBeDisabled()
  })

  it('calls transferCaptaincy on confirm', async () => {
    const user = userEvent.setup()
    transferCaptaincyMock.mockResolvedValueOnce(undefined)
    render(<TransferCaptaincyDialog {...defaultProps} />, { wrapper: createQueryWrapper() })
    const select = screen.getByRole('combobox')
    await user.selectOptions(select, '2')
    await user.click(screen.getByText('Confirmer le transfert'))
    expect(transferCaptaincyMock).toHaveBeenCalledWith(1, { user_id: 2 })
  })

  it('shows loading state during transfer', async () => {
    const user = userEvent.setup()
    transferCaptaincyMock.mockImplementationOnce(() => new Promise(() => {}))
    render(<TransferCaptaincyDialog {...defaultProps} />, { wrapper: createQueryWrapper() })
    const select = screen.getByRole('combobox')
    await user.selectOptions(select, '2')
    await user.click(screen.getByText('Confirmer le transfert'))
    expect(screen.getByText('Transfert...')).toBeInTheDocument()
  })
})
