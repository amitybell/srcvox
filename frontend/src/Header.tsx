import './Header.css'
import Logo from './Logo'
import Presence from './Presence'
import Tabs, { Tab } from './Tabs'
import { GameInfo } from './appstate'

export interface HeaderProps {
  tabs: Tab[]
  setTab: (p: Tab) => void
  activeTab: Tab
  game: GameInfo
}

export default function Header({ tabs, setTab, activeTab, game }: HeaderProps) {
  return (
    <header
      className="header-ctr"
      style={{
        background: `linear-gradient(to bottom, transparent, var(--background)), url(${game.heroURI})`,
        backgroundPosition: 'top left',
        backgroundRepeat: 'no-repeat',
        backgroundSize: 'cover',
      }}
    >
      <div className="logo-ctr">
        <Logo />
      </div>
      <Tabs tabs={tabs} active={activeTab} setTab={setTab} />
      <Presence />
    </header>
  )
}
