import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import CollapsibleSection from './CollapsibleSection'

describe('CollapsibleSection', () => {
  it('renders title and children when defaultOpen is true', () => {
    render(
      <CollapsibleSection title="Test" defaultOpen={true}>
        <p>Content</p>
      </CollapsibleSection>
    )
    expect(screen.getByText('Test')).toBeInTheDocument()
    expect(screen.getByText('Content')).toBeInTheDocument()
  })

  it('renders title but children are hidden when defaultOpen is false', () => {
    render(
      <CollapsibleSection title="Test">
        <p>Content</p>
      </CollapsibleSection>
    )
    expect(screen.getByText('Test')).toBeInTheDocument()
    expect(screen.getByText('Content')).toBeInTheDocument()
    const header = screen.getByRole('button', { name: /Test/ })
    expect(header).toHaveAttribute('aria-expanded', 'false')
  })

  it('opens on header click', async () => {
    const user = userEvent.setup()
    render(
      <CollapsibleSection title="Test">
        <p>Content</p>
      </CollapsibleSection>
    )
    const header = screen.getByRole('button', { name: /Test/ })
    expect(header).toHaveAttribute('aria-expanded', 'false')

    await user.click(header)
    expect(header).toHaveAttribute('aria-expanded', 'true')
  })

  it('closes on header click when already open', async () => {
    const user = userEvent.setup()
    render(
      <CollapsibleSection title="Test" defaultOpen={true}>
        <p>Content</p>
      </CollapsibleSection>
    )
    const header = screen.getByRole('button', { name: /Test/ })
    expect(header).toHaveAttribute('aria-expanded', 'true')

    await user.click(header)
    expect(header).toHaveAttribute('aria-expanded', 'false')
  })

  it('toggles aria-expanded between false and true on consecutive clicks', async () => {
    const user = userEvent.setup()
    render(
      <CollapsibleSection title="Test">
        <p>Content</p>
      </CollapsibleSection>
    )
    const header = screen.getByRole('button', { name: /Test/ })

    await user.click(header)
    expect(header).toHaveAttribute('aria-expanded', 'true')

    await user.click(header)
    expect(header).toHaveAttribute('aria-expanded', 'false')
  })

  it('shows count in title when count prop is provided', () => {
    render(
      <CollapsibleSection title="Membres" count={5}>
        <p>Content</p>
      </CollapsibleSection>
    )
    expect(screen.getByText('Membres (5)')).toBeInTheDocument()
  })

  it('shows title without count when count is not provided', () => {
    render(
      <CollapsibleSection title="Membres">
        <p>Content</p>
      </CollapsibleSection>
    )
    expect(screen.getByText('Membres')).toBeInTheDocument()
  })

  it('applies className prop to the container', () => {
    const { container } = render(
      <CollapsibleSection title="Test" className="danger">
        <p>Content</p>
      </CollapsibleSection>
    )
    const root = container.firstElementChild
    expect(root?.className).toContain('danger')
  })

  it('header has aria-controls attribute pointing to the panel', () => {
    render(
      <CollapsibleSection title="Test">
        <p>Content</p>
      </CollapsibleSection>
    )
    const header = screen.getByRole('button', { name: /Test/ })
    const panelId = header.getAttribute('aria-controls')
    expect(panelId).toBeTruthy()
    const panel = document.getElementById(panelId!)
    expect(panel).toBeInTheDocument()
  })
})
