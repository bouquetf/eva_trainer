import { useState, useRef, useId } from 'react'
import styles from './CollapsibleSection.module.css'

interface CollapsibleSectionProps {
  title: string
  defaultOpen?: boolean
  children: React.ReactNode
  className?: string
  count?: number
}

function CollapsibleSection({
  title,
  defaultOpen = false,
  children,
  className,
  count,
}: CollapsibleSectionProps) {
  const [open, setOpen] = useState(defaultOpen)
  const [maxHeight, setMaxHeight] = useState<string | undefined>(
    defaultOpen ? undefined : '0px'
  )
  const contentRef = useRef<HTMLDivElement>(null)
  const id = useId()
  const panelId = `collapsible-panel-${id}`

  const handleToggle = () => {
    const el = contentRef.current
    if (!el) return

    if (open) {
      setMaxHeight(`${el.scrollHeight}px`)
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          setMaxHeight('0px')
        })
      })
      setOpen(false)
    } else {
      setMaxHeight(`${el.scrollHeight}px`)
      setOpen(true)
    }
  }

  const handleTransitionEnd = () => {
    if (open) {
      setMaxHeight(undefined)
    }
  }

  const displayTitle = count !== undefined ? `${title} (${count})` : title

  return (
    <div className={`${styles.collapsibleSection} ${className ?? ''}`}>
      <button
        className={styles.header}
        onClick={handleToggle}
        aria-expanded={open}
        aria-controls={panelId}
        type="button"
      >
        <h2 className={styles.title}>{displayTitle}</h2>
        <span
          className={`${styles.chevron} ${open ? styles.chevronOpen : ''}`}
        >
          ▶
        </span>
      </button>
      <div
        id={panelId}
        ref={contentRef}
        className={styles.content}
        style={{ maxHeight }}
        onTransitionEnd={handleTransitionEnd}
      >
        <div className={styles.contentInner}>
          {children}
        </div>
      </div>
    </div>
  )
}

export default CollapsibleSection
