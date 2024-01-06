import 'modern-normalize/modern-normalize.css'
import '@mantine/core/styles.css'
import '@mantine/notifications/styles.css'
import './App.css'

import { MantineProvider } from '@mantine/core'
import { Notifications } from '@mantine/notifications'
import { ModalsProvider } from '@mantine/modals'
import { theme } from './theme'
import Shell from './Shell'

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
