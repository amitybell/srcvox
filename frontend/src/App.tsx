import 'modern-normalize/modern-normalize.css'
import './App.css'
import { ReactElement, useState } from 'react'
import Soundboard from './Soundboard'
import Error from './Error'
import Header from './Header'
import { Tab } from './Tabs'
import Home from './Home'
import Settings from './Settings'
import { useAppError } from './hooks/query'
import Credits from './Credits'

const tabs: Tab[] = [{ name: 'home' }, { name: 'soundboard' }, { name: 'credits' }]

function AppBody() {
  const err = useAppError()
  const [tab, setTab] = useState(tabs[0])

  if (err.type !== 'ok') {
    return err.alt
  }

  if (err.v.message) {
    return <Error fatal={err.v.fatal}>{err.v.message}</Error>
  }

  return (
    <>
      <Header tabs={tabs} setTab={setTab} activeTab={tab} />

      {((): ReactElement => {
        switch (tab.name) {
          case 'home':
            return <Home />
          case 'soundboard':
            return <Soundboard />
          case 'settings':
            return <Settings />
          case 'credits':
            return <Credits />
        }
      })()}
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
