import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import MarkdownEditor from './MarkdownEditor'

describe('MarkdownEditor', () => {
  it('renders contentEditable div', () => {
    render(<MarkdownEditor value="hello" onChange={() => {}} />)
    const editable = document.querySelector('[contenteditable]')
    expect(editable).toBeInTheDocument()
  })

  it('renders all toolbar buttons', () => {
    render(<MarkdownEditor value="" onChange={() => {}} />)
    expect(screen.getByTitle('Titre 1')).toBeInTheDocument()
    expect(screen.getByTitle('Titre 2')).toBeInTheDocument()
    expect(screen.getByTitle('Titre 3')).toBeInTheDocument()
    expect(screen.getByTitle('Gras')).toBeInTheDocument()
    expect(screen.getByTitle('Italique')).toBeInTheDocument()
    expect(screen.getByTitle('Liste à puces')).toBeInTheDocument()
    expect(screen.getByTitle('Liste numérotée')).toBeInTheDocument()
    expect(screen.getByTitle('Lien')).toBeInTheDocument()
  })

  it('renders with dark theme container and toolbar', () => {
    render(<MarkdownEditor value="" onChange={() => {}} />)
    expect(screen.getByText('H₁')).toBeInTheDocument()
    expect(screen.getByText('𝐁')).toBeInTheDocument()
    expect(screen.getByText('🔗')).toBeInTheDocument()
  })
})
