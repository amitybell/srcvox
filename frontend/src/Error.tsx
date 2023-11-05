import './Error.css'
import { ReactNode } from 'react'
import Logo from './Logo'

export interface ErrorProps {
  title?: string
  fatal?: boolean
  children: ReactNode
}

export default function Error({ title, fatal, children }: ErrorProps) {
  title = title || (fatal ? 'Fatal Error' : 'Error')
  return (
    <div className={`error-ctr ${fatal ? ' fatal' : ''}`}>
      <Logo />
      {fatal ? <h1>{title}</h1> : <h2>{title}</h2>}
      <div className="error-bdy">{children}</div>
    </div>
  )
}
