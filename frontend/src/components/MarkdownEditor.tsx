import { useRef, useEffect } from 'react'
import TurndownService from 'turndown'
import { marked } from 'marked'

const td = new TurndownService({
  headingStyle: 'atx',
  hr: '---',
  bulletListMarker: '-',
  codeBlockStyle: 'fenced',
  emDelimiter: '*',
  strongDelimiter: '**',
  linkStyle: 'inlined',
  linkReferenceStyle: 'full',
})

function handleWysiwygLink() {
  const url = window.prompt('URL du lien :')
  if (url) {
    document.execCommand('createLink', false, url)
  }
}

const BUTTONS = [
  { label: 'H₁', title: 'Titre 1', action: () => document.execCommand('formatBlock', false, '<h1>') },
  { label: 'H₂', title: 'Titre 2', action: () => document.execCommand('formatBlock', false, '<h2>') },
  { label: 'H₃', title: 'Titre 3', action: () => document.execCommand('formatBlock', false, '<h3>') },
  { label: '𝐁', title: 'Gras', action: () => document.execCommand('bold') },
  { label: '𝑰', title: 'Italique', action: () => document.execCommand('italic') },
  { label: '•', title: 'Liste à puces', action: () => document.execCommand('insertUnorderedList') },
  { label: '1.', title: 'Liste numérotée', action: () => document.execCommand('insertOrderedList') },
  { label: '🔗', title: 'Lien', action: () => handleWysiwygLink() },
]

export default function MarkdownEditor({
  value,
  onChange,
}: {
  value: string
  onChange: (v: string) => void
}) {
  const wysiwygRef = useRef<HTMLDivElement>(null)
  const userChangeRef = useRef(false)

  useEffect(() => {
    if (!wysiwygRef.current) return
    if (userChangeRef.current) {
      userChangeRef.current = false
      return
    }
    const html = marked.parse(value || '') as string
    if (wysiwygRef.current.innerHTML !== html) {
      wysiwygRef.current.innerHTML = html
    }
  }, [value])

  function onWysiwygInput() {
    if (!wysiwygRef.current) return
    userChangeRef.current = true
    const html = wysiwygRef.current.innerHTML
    if (html === '<br>') {
      onChange('')
      return
    }
    const md = td.turndown(html).trim()
    onChange(md)
  }

  return (
    <div style={{ border: '1px solid #333', borderRadius: 6, overflow: 'hidden', background: '#1e1e1e' }}>
      <div style={{
        display: 'flex', gap: 1, padding: '4px 6px', background: '#2a2a2a',
        borderBottom: '1px solid #333', flexWrap: 'wrap', alignItems: 'center',
      }}>
        {BUTTONS.map((btn) => (
          <button
            key={btn.title}
            type="button"
            title={btn.title}
            onClick={() => { btn.action(); wysiwygRef.current?.dispatchEvent(new Event('input', { bubbles: true })) }}
            style={{
              padding: '3px 8px', fontSize: 13, cursor: 'pointer',
              border: '1px solid #444', borderRadius: 3,
              background: '#333', color: '#e0e0e0',
              fontWeight: btn.label === '𝐁' ? 700 : 400,
              fontStyle: btn.label === '𝑰' ? 'italic' : 'normal',
              minWidth: 28, textAlign: 'center',
            }}
          >
            {btn.label}
          </button>
        ))}
      </div>
      <div
        ref={wysiwygRef}
        contentEditable
        suppressContentEditableWarning
        onInput={onWysiwygInput}
        style={{
          width: '100%', minHeight: 120, padding: '8px 12px',
          outline: 'none', fontSize: 14, lineHeight: 1.6,
          color: '#e0e0e0',
        }}
        data-placeholder="Écrire une note…"
      />
      <style>{`
        [contenteditable]:empty:before {
          content: attr(data-placeholder);
          color: #666;
          pointer-events: none;
        }
        [contenteditable] h1 { font-size: 1.5em; margin: 0.5em 0; }
        [contenteditable] h2 { font-size: 1.3em; margin: 0.4em 0; }
        [contenteditable] h3 { font-size: 1.15em; margin: 0.3em 0; }
        [contenteditable] ul, [contenteditable] ol { padding-left: 1.5em; }
        [contenteditable] p { margin: 0.3em 0; }
      `}</style>
    </div>
  )
}
