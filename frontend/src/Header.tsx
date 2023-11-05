import './Header.css'
import Logo from './Logo'
import Presence from './Presence'
import Tabs, { Tab } from './Tabs'

export interface HeaderProps {
  tabs: Tab[]
  setTab: (p: Tab) => void
  activeTab: Tab
}

export default function Header({ tabs, setTab, activeTab }: HeaderProps) {
  return (
    <header className="header-ctr">
      <div className="logo-ctr">
        <Logo />
      </div>
      <Tabs tabs={tabs} active={activeTab} setTab={setTab} />
      <Presence />
    </header>
  )
}
