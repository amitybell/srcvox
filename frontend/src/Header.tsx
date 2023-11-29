import { ReactElement, useState } from 'react'
import './Header.css'
import Logo from './Logo'
import Menu, { MenuItemProps } from './Menu'
import Presence from './Presence'
import Tabs, { Tab } from './Tabs'
import { CiMenuBurger as MenuIconClosed, CiMenuFries as MenuIconOpened } from 'react-icons/ci'
import Modal from './Modal'
import Credits from './Credits'

export interface HeaderProps {
  tabs: Tab[]
  setTab: (p: Tab) => void
  activeTab: Tab
}

type OverlayPage = 'credits'

interface OverlayProps {
  page: OverlayPage
  onClose: () => void
}

function Overlay({ page, onClose }: OverlayProps) {
  return (
    <Modal onClose={onClose}>
      {((): ReactElement => {
        switch (page) {
          case 'credits':
            return <Credits />
        }
      })()}
    </Modal>
  )
}

type OverlayKind = OverlayPage | 'menu'

export default function Header({ tabs, setTab, activeTab }: HeaderProps) {
  const [overlay, setOverlay] = useState<OverlayKind | null>(null)
  const menuOpen = overlay === 'menu'
  const menuItems: MenuItemProps[] = [
    { body: <button onClick={() => setOverlay('credits')}>Credits</button> },
  ]

  return (
    <div className="header-ctr">
      <header className="header">
        <div className="logo-ctr">
          <Logo />
        </div>
        <Tabs tabs={tabs} active={activeTab} setTab={setTab} />
        <Presence />
        <Menu
          title={
            menuOpen ? (
              <MenuIconOpened className="menu-icon" />
            ) : (
              <MenuIconClosed className="menu-icon" />
            )
          }
          hover={false}
          open={menuOpen}
          onToggle={() => setOverlay(overlay === 'menu' ? null : 'menu')}
          items={menuItems}
        />
      </header>
      {overlay && overlay !== 'menu' ? (
        <Overlay page={overlay} onClose={() => setOverlay(null)} />
      ) : null}
    </div>
  )
}
