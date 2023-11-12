import './Menu.css'
import React, { ReactElement } from 'react'

export interface MenuItemProps {
  body: ReactElement
}

function MenuItem({ body }: MenuItemProps) {
  return (
    <li>
      <div className="menu-item">{body}</div>
    </li>
  )
}

export interface MenuProps {
  open: boolean
  onToggle: (open: boolean) => void
  hover: boolean
  title: ReactElement
  items: Array<MenuItemProps> | Array<MenuItemProps & { key: React.Key }>
  className?: string
}

export default function Menu({ title, items, open, onToggle, className, hover }: MenuProps) {
  return (
    <div
      className={`menu-ctr ${className || ''}`}
      onMouseEnter={hover ? () => onToggle(true) : undefined}
      onMouseLeave={hover ? () => onToggle(false) : undefined}
    >
      <div className="menu-title" onClick={hover ? undefined : () => onToggle(!open)}>
        {title}
      </div>
      {open ? (
        <ul className="menu-list">
          {items.map((p, i) => (
            <MenuItem key={'key' in p ? p.key : i} body={p.body} />
          ))}
        </ul>
      ) : null}
    </div>
  )
}
