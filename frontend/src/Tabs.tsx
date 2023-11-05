import './Tabs.css'

export type Tab =
  | { name: 'home' }
  | { name: 'soundboard' }
  | { name: 'settings' }
  | { name: 'credits' }

export interface TabProps {
  tab: Tab
  active: boolean
  onClick: (t: Tab) => void
}

function TabButton({ active, onClick, tab }: TabProps) {
  return (
    <button className={`tab ${active ? 'active' : ''}`} onClick={() => onClick(tab)}>
      {tab.name}
    </button>
  )
}

export interface TabsProps {
  tabs: Tab[]
  active: Tab
  setTab: (t: Tab) => void
}

export default function Tabs({ tabs, active, setTab }: TabsProps) {
  return (
    <div className="tabs-ctr">
      {tabs.map((t) => (
        <TabButton tab={t} key={t.name} active={t.name === active.name} onClick={setTab} />
      ))}
    </div>
  )
}
