import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import TeamsLayout from './TeamsLayout'

vi.mock('./AuthGuard', () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

describe('TeamsLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders sidebar with team links', () => {
    render(
      <MemoryRouter>
        <TeamsLayout />
      </MemoryRouter>,
    )
    expect(screen.getByText('Mes équipes')).toBeInTheDocument()
    expect(screen.getByText('Créer une équipe')).toBeInTheDocument()
    expect(screen.getByText('Paramètres')).toBeInTheDocument()
  })

  it('renders Outlet content', () => {
    render(
      <MemoryRouter initialEntries={['/teams']}>
        <TeamsLayout />
      </MemoryRouter>,
    )
    expect(screen.getByText('Mes équipes').closest('a')).toHaveAttribute('href', '/teams')
  })

  it('is wrapped in AuthGuard', () => {
    const { container } = render(
      <MemoryRouter>
        <TeamsLayout />
      </MemoryRouter>,
    )
    expect(container.querySelector('.teams-layout')).toBeInTheDocument()
  })
})
