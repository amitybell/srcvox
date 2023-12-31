import 'modern-normalize/modern-normalize.css'
import '@mantine/core/styles.css'
import './App.css'
import 'react-toastify/dist/ReactToastify.css'

import { ReactElement, useState } from 'react'
import Soundboard from './Soundboard'
import Error from './Error'
import Header from './Header'
import { Tab } from './Tabs'
import Servers from './Servers'
import Settings from './Settings'
import { useAppError, useEnv, useGames } from './hooks/query'
import BgVideo from './BgVideo'
import { ToastContainer } from 'react-toastify'
import { MantineProvider } from '@mantine/core'
import { theme } from './theme'

const tabs: Tab[] = [{ name: 'servers' }, { name: 'soundboard' }]

function AppBody() {
  const err = useAppError()
  const env = useEnv()
  const [activeGameIdx, setAciveGameIdx] = useState(0)
  const games = useGames()
  const [selectedTab, setSelectedTab] = useState<Tab | null>(null)

  if (err.type !== 'ok') {
    return err.alt
  }
  if (env.type !== 'ok') {
    return env.alt
  }
  if (games.type !== 'ok') {
    return games.alt
  }

  if (err.v.message) {
    return (
      <Error fatal={err.v.fatal}>
        {typeof err.v.message}: {err.v.message}
      </Error>
    )
  }

  const tab = selectedTab || tabs.find((t) => t.name === env.v.initTab) || tabs[0]
  const game = games.v[activeGameIdx]

  return (
    <>
      <BgVideo src={game.bgVideoURL} />
      <div className="app-ctr">
        <div className="page-content">
          <Header tabs={tabs} setTab={setSelectedTab} activeTab={tab} />
          <main className="app-body">
            {((): ReactElement => {
              switch (tab.name) {
                case 'servers':
                  return (
                    <Servers
                      games={games.v}
                      gameIdx={activeGameIdx}
                      onGameSelect={(i) => {
                        setAciveGameIdx(i)
                      }}
                    />
                  )
                case 'soundboard':
                  return <Soundboard />
              }
            })()}
          </main>
        </div>
      </div>
      <ToastContainer />
    </>
  )
}

export default function App() {
  return (
    <div id="App">
      <MantineProvider theme={theme} defaultColorScheme="dark">
        <AppBody />
      </MantineProvider>
    </div>
  )
}
