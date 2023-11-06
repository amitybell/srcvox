import 'modern-normalize/modern-normalize.css'
import './App.css'
import { ReactElement, useState } from 'react'
import Soundboard from './Soundboard'
import Error from './Error'
import Header from './Header'
import { Tab } from './Tabs'
import Home from './Home'
import Settings from './Settings'
import { useAppError, useGames } from './hooks/query'
import Credits from './Credits'

const tabs: Tab[] = [{ name: 'home' }, { name: 'soundboard' }, { name: 'credits' }]

function AppBody() {
  const err = useAppError()
  const [tab, setTab] = useState(tabs[0])
  const [activeGameIdx, setAciveGameIdx] = useState(0)
  const games = useGames()

  if (err.type !== 'ok') {
    return err.alt
  }
  if (games.type !== 'ok') {
    return games.alt
  }

  if (err.v.message) {
    return <Error fatal={err.v.fatal}>{err.v.message}</Error>
  }

  const game = games.v[activeGameIdx]

  return (
    <>
      <Header tabs={tabs} setTab={setTab} activeTab={tab} game={game} />
      <main className="app-body">
        {((): ReactElement => {
          switch (tab.name) {
            case 'home':
              return <Home games={games.v} gameIdx={activeGameIdx} onGameSelect={setAciveGameIdx} />
            case 'soundboard':
              return <Soundboard />
            case 'settings':
              return <Settings />
            case 'credits':
              return <Credits />
          }
        })()}
      </main>
    </>
  )
}

export default function App() {
  return (
    <div id="App">
      <AppBody />
    </div>
  )
}
