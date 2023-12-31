import './Menu.css'
import React, { ReactElement } from 'react'
import { PiArrowFatRight as RightArrow, PiArrowFatDown as DownArrow } from 'react-icons/pi'
import { Menu as MMenu, PopoverWidth } from '@mantine/core'

export interface MenuItemProps {
  onClick: () => void
  body: ReactElement
}

export interface MenuProps {
  open: boolean
  onToggle: (open: boolean) => void
  title: ReactElement
  items: Array<MenuItemProps> | Array<MenuItemProps & { key: React.Key }>
  width?: PopoverWidth
  className?: string
  indicator?: boolean
}

export default function Menu({
  title,
  items,
  open,
  onToggle,
  width,
  className,
  indicator,
}: MenuProps) {
  return (
    <div className={`menu-ctr ${className ?? ''}`}>
      <MMenu
        onChange={onToggle}
        position="bottom"
        withArrow
        shadow="md"
        opened={open}
        width={width}
      >
        <MMenu.Target>
          <div className="menu-title">
            {title}
            {indicator && items.length !== 0 ? open ? <DownArrow /> : <RightArrow /> : null}
          </div>
        </MMenu.Target>
        {open && items.length !== 0 ? (
          <MMenu.Dropdown component="div" className="menu-list">
            {items.map((p, i) => (
              <MMenu.Item
                onClick={p.onClick}
                component="div"
                className="menu-item"
                key={'key' in p ? p.key : i}
              >
                {p.body}
              </MMenu.Item>
            ))}
          </MMenu.Dropdown>
        ) : null}
      </MMenu>
    </div>
  )
}
