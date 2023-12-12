import './Error.css'
import { ReactNode } from 'react'
import Logo from './Logo'

export interface ErrorProps {
  title?: string
  fatal?: boolean
  children: ReactNode
  noLogo?: boolean
}

export default function Error({ title, fatal, noLogo, children }: ErrorProps) {
  title = title || (fatal ? 'Fatal Error' : 'Error')
  return (
    <div className={`error-ctr ${fatal ? ' fatal' : ''}`}>
      {noLogo ? null : <Logo />}
      {fatal ? <h1>{title}</h1> : <h2>{title}</h2>}
      <div className="error-bdy">{children}</div>
    </div>
  )
}
