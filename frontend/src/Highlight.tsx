import './Highlight.css'
import { ReactNode } from 'react'

export interface HighlightProps {
  text: string
  pat: RegExp
  className?: string
}

export default function Highlight({ text, pat, className }: HighlightProps) {
  // copy to avoid changing its state
  // set g flag so pat.lastIndex is updated for each match
  const re = new RegExp(pat.source, pat.flags.includes('g') ? pat.flags : pat.flags + 'g')

  const cl: ReactNode[] = []
  for (let prevIdx = re.lastIndex; ; prevIdx = Math.max(prevIdx, re.lastIndex)) {
    const mat = re.exec(text)?.[0]
    if (mat == null) {
      // text after the last match
      if (prevIdx < text.length) {
        cl.push(
          <span key={cl.length} className={className || ''}>
            {text.slice(prevIdx)}
          </span>,
        )
      }
      break
    }
    const currIdx = re.lastIndex - mat.length
    if (prevIdx < currIdx) {
      // text between previous match and the current one
      cl.push(
        <span className={className || ''} key={cl.length}>
          {text.slice(prevIdx, currIdx)}
        </span>,
      )
    }
    cl.push(
      <span key={cl.length} className={`${className || ''} highlight`}>
        {mat}
      </span>,
    )
  }

  return <>{cl.length !== 0 ? cl : <span>text</span>}</>
}
