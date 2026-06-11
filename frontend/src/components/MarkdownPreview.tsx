import { useMemo } from 'react'
import { marked } from 'marked'

marked.setOptions({
  gfm: true,
})

export default function MarkdownPreview({ content }: { content: string }) {
  const html = useMemo(() => {
    if (!content) return '<p style="color:#888;font-style:italic">Aucun contenu</p>'
    return marked.parse(content) as string
  }, [content])

  return (
    <div
      className="markdown-preview"
      style={{ lineHeight: 1.6, overflowWrap: 'break-word' }}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}
