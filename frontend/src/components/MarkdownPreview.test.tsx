import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { marked } from 'marked'
import MarkdownPreview from './MarkdownPreview'

describe('marked basic behavior', () => {
  it('parses unordered list correctly', () => {
    const result = marked.parse('- item1\n- item2') as string
    expect(result).toContain('<li>item1</li>')
    expect(result).toContain('<li>item2</li>')
  })

  it('parses ordered list correctly', () => {
    const result = marked.parse('1. first\n2. second') as string
    expect(result).toContain('<li>first</li>')
    expect(result).toContain('<li>second</li>')
  })

  it('parses headings with newlines', () => {
    const result = marked.parse('# H1\n## H2\n### H3') as string
    expect(result).toContain('<h1>H1</h1>')
    expect(result).toContain('<h2>H2</h2>')
    expect(result).toContain('<h3>H3</h3>')
  })
})

describe('MarkdownPreview', () => {
  it('renders empty content placeholder', () => {
    render(<MarkdownPreview content="" />)
    expect(screen.getByText('Aucun contenu')).toBeInTheDocument()
  })

  it('renders bold markdown', () => {
    const { container } = render(<MarkdownPreview content="**bold** text" />)
    expect(container.innerHTML).toContain('<strong>bold</strong>')
  })

  it('renders italic markdown', () => {
    const { container } = render(<MarkdownPreview content="*italic* text" />)
    expect(container.innerHTML).toContain('<em>italic</em>')
  })

  it('renders heading', () => {
    const { container } = render(<MarkdownPreview content="# Heading 1" />)
    expect(container.innerHTML).toContain('<h1>Heading 1</h1>')
  })

  it('renders links', () => {
    const { container } = render(<MarkdownPreview content="[click](https://example.com)" />)
    expect(container.innerHTML).toContain('href="https://example.com"')
    expect(container.innerHTML).toContain('>click<')
  })

  it('renders list items from markdown', async () => {
    render(<MarkdownPreview content={'- item1\n- item2'} />)
    const items = screen.getAllByRole('listitem')
    expect(items.length).toBe(2)
  })

  it('renders ordered list items', async () => {
    render(<MarkdownPreview content={'1. first\n2. second'} />)
    const items = screen.getAllByRole('listitem')
    expect(items.length).toBe(2)
  })

  it('renders multi-level headings with newlines', async () => {
    const content = ['# H1', '## H2', '### H3'].join('\n')
    const { container } = render(<MarkdownPreview content={content} />)
    expect(container.innerHTML).toContain('<h1>H1</h1>')
    expect(container.innerHTML).toContain('<h2>H2</h2>')
    expect(container.innerHTML).toContain('<h3>H3</h3>')
  })
})
