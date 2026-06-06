import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AdminMenu from './AdminMenu'

describe('AdminMenu', () => {
  it('renders all tab buttons', () => {
    render(<AdminMenu activeTab="users" onTabChange={() => {}} />)
    expect(screen.getByText('Utilisateurs')).toBeInTheDocument()
    expect(screen.getByText('Permissions')).toBeInTheDocument()
    expect(screen.getByText('Cartes')).toBeInTheDocument()
    expect(screen.getByText('Splits')).toBeInTheDocument()
    expect(screen.getByText('Catalogue')).toBeInTheDocument()
    expect(screen.getByText('Import/Export')).toBeInTheDocument()
    expect(screen.getByText('Ligues')).toBeInTheDocument()
    expect(screen.getByText('Archives')).toBeInTheDocument()
  })

  it('marks the active tab with active class', () => {
    const { rerender } = render(<AdminMenu activeTab="maps" onTabChange={() => {}} />)
    const mapsBtn = screen.getByText('Cartes')
    const usersBtn = screen.getByText('Utilisateurs')
    expect(mapsBtn.className).toBe('active')
    expect(usersBtn.className).not.toBe('active')

    rerender(<AdminMenu activeTab="users" onTabChange={() => {}} />)
    expect(mapsBtn.className).not.toBe('active')
    expect(usersBtn.className).toBe('active')
  })

  it('calls onTabChange with the correct tab key', async () => {
    const onTabChange = vi.fn()
    render(<AdminMenu activeTab="users" onTabChange={onTabChange} />)
    await userEvent.click(screen.getByText('Ligues'))
    expect(onTabChange).toHaveBeenCalledWith('ligues')
  })

  it('shows counts when provided', () => {
    render(
      <AdminMenu
        activeTab="users"
        onTabChange={() => {}}
        usersCount={12}
        mapsCount={5}
        splitsCount={3}
        catalogCount={8}
      />,
    )
    expect(screen.getByText('Utilisateurs (12)')).toBeInTheDocument()
    expect(screen.getByText('Cartes (5)')).toBeInTheDocument()
    expect(screen.getByText('Splits (3)')).toBeInTheDocument()
    expect(screen.getByText('Catalogue (8)')).toBeInTheDocument()
    expect(screen.getByText('Permissions')).toBeInTheDocument()
    expect(screen.getByText('Import/Export')).toBeInTheDocument()
    expect(screen.getByText('Ligues')).toBeInTheDocument()
    expect(screen.getByText('Archives')).toBeInTheDocument()
  })

  it('does not show count for tabs without a countKey', () => {
    render(<AdminMenu activeTab="users" onTabChange={() => {}} usersCount={99} />)
    expect(screen.getByText('Permissions')).toBeInTheDocument()
    expect(screen.getByText('Import/Export')).toBeInTheDocument()
    expect(screen.queryByText(/Permissions \(\d+\)/)).not.toBeInTheDocument()
  })
})
