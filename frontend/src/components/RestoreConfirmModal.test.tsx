import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import RestoreConfirmModal from './RestoreConfirmModal'

describe('RestoreConfirmModal', () => {
  const defaultProps = {
    open: true,
    versionDate: '2024-06-15',
    authorName: 'Jean Dupont',
    onConfirm: vi.fn(),
    onCancel: vi.fn(),
    loading: false,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when closed', () => {
    const { container } = render(<RestoreConfirmModal {...defaultProps} open={false} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders version date and author', () => {
    render(<RestoreConfirmModal {...defaultProps} />)
    expect(screen.getByText('2024-06-15')).toBeInTheDocument()
    expect(screen.getByText('Jean Dupont')).toBeInTheDocument()
  })

  it('calls onConfirm on confirm button click', async () => {
    const onConfirm = vi.fn()
    render(<RestoreConfirmModal {...defaultProps} onConfirm={onConfirm} />)
    await userEvent.click(screen.getByText('Confirmer'))
    expect(onConfirm).toHaveBeenCalledTimes(1)
  })

  it('calls onCancel on cancel button click', async () => {
    const onCancel = vi.fn()
    render(<RestoreConfirmModal {...defaultProps} onCancel={onCancel} />)
    await userEvent.click(screen.getByText('Annuler'))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('calls onCancel on Escape key', async () => {
    const onCancel = vi.fn()
    render(<RestoreConfirmModal {...defaultProps} onCancel={onCancel} />)
    await userEvent.keyboard('{Escape}')
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('disables buttons during loading', () => {
    render(<RestoreConfirmModal {...defaultProps} loading />)
    expect(screen.getByText('Annuler')).toBeDisabled()
    expect(screen.getByText('Restauration...')).toBeDisabled()
  })

  it('shows Restauration... text during loading', () => {
    render(<RestoreConfirmModal {...defaultProps} loading />)
    expect(screen.getByText('Restauration...')).toBeInTheDocument()
  })
})
