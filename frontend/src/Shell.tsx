import '@mantine/core/styles.css'
import '@mantine/notifications/styles.css'
import 'modern-normalize/modern-normalize.css'
import './App.css'
import cls from './Shell.module.css'

import { ActionIcon, Tooltip, TooltipProps, rem } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import {
  IconCamera,
  IconHome,
  IconInfoCircle,
  IconSettings,
  IconSpeakerphone,
  TablerIconsProps,
} from '@tabler/icons-react'
import { ComponentType, ReactElement, useEffect, useState } from 'react'
import Credits from './Credits'
import Error from './Error'
import Presence, { PresenceAvatar } from './Presence'
import Servers from './Servers'
import Settings from './Settings'
import Snapshots from './Snapshots'
import Soundboard from './Soundboard'
import Wallpaper from './Wallpaper'
import { GameInfo } from './appstate'
import { useAppError, useGames } from './hooks/query'

type PageName =
  | { name: 'home' }
  | { name: 'snap' }
  | { name: 'soundboard' }
  | { name: 'credits' }
  | { name: 'settings' }

type Page = PageName & {
  title: string
  Icon: ComponentType<TablerIconsProps>
}

const pages: Page[] = [
  { name: 'home', title: 'Home', Icon: IconHome },
  { name: 'soundboard', title: 'Soundboard', Icon: IconSpeakerphone },
  { name: 'snap', title: 'Snapshots', Icon: IconCamera },
  { name: 'credits', title: 'Credits', Icon: IconInfoCircle },
  { name: 'settings', title: 'Settings', Icon: IconSettings },
]

function Head() {
  return (
    <header className={cls.head}>
      <Presence />
    </header>
  )
}

interface SideProps {
  pages: Page[]
  active: Page
  goto: (p: Page) => void
}

function Side({ pages, active, goto }: SideProps) {
  const iconSize = rem(32)
  const tprops: Partial<TooltipProps> = {
    position: 'right',
    withArrow: true,
  }

  return (
    <nav className={cls.side}>
      {pages.map((p) => (
        <ActionIcon
          key={p.name}
          aria-label={p.title}
          variant={p.name === active.name ? 'filled' : 'light'}
          onClick={() => goto(p)}
          size={iconSize}
          color="dark"
        >
          <Tooltip {...tprops} label={p.title}>
            <p.Icon size={iconSize} stroke={p.name === active.name ? 2 : 1.5} />
          </Tooltip>
        </ActionIcon>
      ))}

      <PresenceAvatar className={cls.presenceAvatar} popover={{}} />
    </nav>
  )
}

interface ContentProps {
  page: Page
  games: GameInfo[]
  activeGameIdx: number
  setActiveGameIdx: (idx: number) => void
}

function Content({ page, games, activeGameIdx, setActiveGameIdx }: ContentProps) {
  const err = useAppError()

  useEffect(() => {
    if (err.type === 'ok' && !err.v.fatal && err.v.message) {
      notifications.show({
        id: 'sv.appErr',
        autoClose: false,
        color: 'red',
        message: err.v.message,
      })
    }
  })

  if (err.type !== 'ok') {
    return err.alt
  }

  if (err.v.message && err.v.fatal) {
    return <Error fatal={err.v.fatal}>{err.v.message}</Error>
  }

  return ((): ReactElement => {
    switch (page.name) {
      case 'home':
        return (
          <Servers
            games={games}
            gameIdx={activeGameIdx}
            onGameSelect={(i) => {
              setActiveGameIdx(i)
            }}
          />
        )
      case 'soundboard':
        return <Soundboard />
      case 'credits':
        return <Credits />
      case 'snap':
        return <Snapshots game={games[activeGameIdx]} />
      case 'settings':
        return <Settings game={games[activeGameIdx]} />
    }
  })()
}

function Body(p: ContentProps) {
  return (
    <main className={cls.body}>
      <div className={cls.content}>
        <Content {...p} />
      </div>
      <Foot />
    </main>
  )
}

function Foot() {
  return <footer className={cls.foot}></footer>
}

export default function Shell() {
  const savedPage = localStorage.getItem('app.page')
  const [page, setPage] = useState(pages.find((p) => p.name === savedPage) ?? pages[0])
  const [activeGameIdx, setAciveGameIdx] = useState(0)
  const games = useGames()

  if (games.type !== 'ok') {
    return games.alt
  }

  const game = games.v[activeGameIdx]

  function gotoPage(p: Page) {
    setPage(p)
    localStorage.setItem('app.page', p.name)
  }

  return (
    <Wallpaper urls={game.mapImageURLs} interval={30000}>
      <div className={cls.shell}>
        <Side pages={pages} active={page} goto={gotoPage} />
        <Body
          page={page}
          games={games.v}
          activeGameIdx={activeGameIdx}
          setActiveGameIdx={setAciveGameIdx}
        />
      </div>
    </Wallpaper>
  )
}
