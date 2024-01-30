import '@mantine/core/styles.css'
import '@mantine/notifications/styles.css'
import 'modern-normalize/modern-normalize.css'
import './App.css'

import { MantineProvider } from '@mantine/core'
import { ModalsProvider } from '@mantine/modals'
import { Notifications } from '@mantine/notifications'
import Shell from './Shell'
import { theme } from './theme'

export default function App() {
  return (
    <div id="App">
      <MantineProvider theme={theme} defaultColorScheme="dark">
        <ModalsProvider>
          <Notifications />
          <Shell />
        </ModalsProvider>
      </MantineProvider>
    </div>
  )
}
