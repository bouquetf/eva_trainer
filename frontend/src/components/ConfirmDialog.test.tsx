import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ConfirmDialog from './ConfirmDialog'

describe('ConfirmDialog', () => {
  const defaultProps = {
    open: true,
    title: 'Supprimer',
    message: 'Voulez-vous vraiment supprimer ?',
    onConfirm: vi.fn(),
    onCancel: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when closed', () => {
    const { container } = render(<ConfirmDialog {...defaultProps} open={false} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders title and message when open', () => {
    render(<ConfirmDialog {...defaultProps} />)
    expect(screen.getByText('Supprimer')).toBeInTheDocument()
    expect(screen.getByText('Voulez-vous vraiment supprimer ?')).toBeInTheDocument()
  })

  it('calls onConfirm when confirm button clicked', async () => {
    const onConfirm = vi.fn()
    render(<ConfirmDialog {...defaultProps} onConfirm={onConfirm} />)
    await userEvent.click(screen.getByText('Confirmer'))
    expect(onConfirm).toHaveBeenCalledTimes(1)
  })

  it('calls onCancel when cancel button clicked', async () => {
    const onCancel = vi.fn()
    render(<ConfirmDialog {...defaultProps} cancelLabel="Annuler" onCancel={onCancel} />)
    await userEvent.click(screen.getByText('Annuler'))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('calls onCancel on Escape key', async () => {
    const onCancel = vi.fn()
    render(<ConfirmDialog {...defaultProps} onCancel={onCancel} />)
    await userEvent.keyboard('{Escape}')
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('calls onCancel when overlay clicked', async () => {
    const onCancel = vi.fn()
    const { container } = render(<ConfirmDialog {...defaultProps} onCancel={onCancel} />)
    const overlay = container.querySelector('.modal-overlay')!
    await userEvent.click(overlay)
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('does not call onCancel when modal content clicked', async () => {
    const onCancel = vi.fn()
    render(<ConfirmDialog {...defaultProps} onCancel={onCancel} />)
    const heading = screen.getByText('Supprimer')
    await userEvent.click(heading)
    expect(onCancel).not.toHaveBeenCalled()
  })

  it('renders danger class when danger prop is true', () => {
    render(<ConfirmDialog {...defaultProps} danger />)
    const confirmBtn = screen.getByText('Confirmer')
    expect(confirmBtn.className).toContain('btn-danger')
  })

  it('renders altButton when provided', () => {
    const altButton = { label: 'Voir details', onClick: vi.fn() }
    render(<ConfirmDialog {...defaultProps} altButton={altButton} />)
    expect(screen.getByText('Voir details')).toBeInTheDocument()
  })
})
