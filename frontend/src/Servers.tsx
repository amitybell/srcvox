import './Servers.css'
import { ReactElement, ReactNode, useEffect, useRef, useState } from 'react'
import Menu from './Menu'
import { GameInfo, Presence, ServerInfo } from './appstate'
import { useEnv, useMapImage, usePresence, useServerInfos, useServers } from './hooks/query'
import { openURL } from './api'
import {
  PiLockKeyOpenFill as PrivateServerIcon,
  PiArrowFatLinesRightFill as ConnectIcon,
  PiArrowFatRight as RightArrow,
  PiArrowFatDown as DownArrow,
  PiCameraDuotone as ScreenshotIcon,
} from 'react-icons/pi'
import Flag from 'react-world-flags'
import { toast } from 'react-toastify'
import Avatar from './Avatar'
import { domToBlob } from 'modern-screenshot'

interface GameProps {
  p: GameInfo
  subtitle?: ReactElement
  onClick?: () => void
}

function Game({ p: { iconURI, title }, onClick, subtitle }: GameProps) {
  return (
    <div className="game-ctr" title={title} onClick={onClick}>
      <img className="game-icon" src={iconURI} />
      <span className="game-title">{title}</span>
      &nbsp;
      <span className="game-subtitle">{subtitle}</span>
    </div>
  )
}

function ServerListInfoIconCell({
  p,
  game,
  src,
  onLoad,
}: ServerListInfoProps & { src?: string; onLoad?: () => void }) {
  return (
    <div className="servers-info-image-ctr">
      {p.restricted ? (
        <PrivateServerIcon className="servers-info-icon" />
      ) : (
        <Flag className="servers-info-icon" code={p.country} />
      )}
      <img
        className="servers-info-map-image"
        alt=""
        src={src || `${game.mapImageURL}&map=${p.map}`}
        onLoad={onLoad}
      />
    </div>
  )
}

async function screenshot(elem: HTMLElement | null) {
  if (!elem) {
    return
  }

  navigator.clipboard
    .write([
      new ClipboardItem({
        'image/png': domToBlob(elem).then((blob) => {
          toast.success('Badge copied to clipbaord')
          return blob
        }),
      }),
    ])
    .catch((e) => {
      toast.error(`Failed to copy badge to clipbaord: ${e}`)
    })
}

function ServerInfoBadge({ p, game, presence: pr }: ServerListInfoProps) {
  const badgeRef = useRef<HTMLDivElement | null>(null)
  const mapImg = useMapImage(game.id, p.map)

  if (mapImg.type !== 'ok') {
    return null
  }

  const players = p.players
  const humans = pr.server === p.addr ? pr.humans : []

  return (
    <div className="servers-info-badge-ctr">
      <div className="servers-info-badge" ref={badgeRef}>
        <img className="servers-info-badge-bg" src={mapImg.v} alt="" />
        <div className="servers-info-badge-content">
          <div className="servers-info-badge-title">
            <span className="servers-info-badge-player-count">{players}</span>{' '}
            <span>{players === 1 ? 'player' : 'players'} on</span>{' '}
            <span className="servers-info-badge-server-name">{p.name}</span>{' '}
            {p.restricted ? (
              <PrivateServerIcon className="servers-info-badge-icon" />
            ) : (
              <Flag className="servers-info-badge-icon" code={p.country} />
            )}
          </div>
          {humans.length ? (
            <div className="servers-info-badge-players-ctr">
              {humans.map((p, i) => (
                <Avatar key={p.id || i} {...p} />
              ))}
            </div>
          ) : null}
        </div>
      </div>
      <button className="servers-info-badge-btn" onClick={() => screenshot(badgeRef.current)}>
        <ScreenshotIcon />
      </button>
    </div>
  )
}

interface ServerListInfoProps {
  p: ServerInfo
  game: GameInfo
  presence: Presence
}

function ServerListInfo({ p, game, presence }: ServerListInfoProps) {
  const [open, setOpen] = useState(false)
  const details: [string, ReactNode][] = [
    ['Name:', p.name],
    [
      'Address:',
      <a
        key={p.addr}
        href={`steam://connect/${p.addr}`}
        target="_blank"
        rel="noreferrer"
        onClick={() => openURL(`steam://connect/${p.addr}`)}
      >
        {p.addr}
      </a>,
    ],
    ['Ping:', p.ping],
    ['Country:', p.country],
    ['Players:', `${p.players} / ${p.maxPlayers} (${p.bots} bots)`],
    ['Game:', p.game],
    ['Map:', p.map],
    ['Protected:', p.restricted ? 'yes' : 'no'],
  ]

  return (
    <>
      <tr
        className={`servers-info ${p.players <= 0 ? 'empty' : ''} ${
          p.restricted ? 'restricted' : ''
        }`}
        title={`${p.name} ${p.addr}`}
      >
        <td className="servers-info-image-ctr-ctr">
          <ServerListInfoIconCell p={p} game={game} presence={presence} />
        </td>
        <td className="servers-info-name-ctr" onClick={() => setOpen(!open)}>
          <span className="servers-info-name">
            {p.name} {open ? <DownArrow /> : <RightArrow />}
          </span>
        </td>
        <td>
          <span className="servers-info-players">{p.players}</span>
        </td>
        <td>
          <span className="servers-info-ping">{p.ping}</span>
        </td>
        <td>
          <button
            className="servers-info-join"
            onClick={() => openURL(`steam://connect/${p.addr}`)}
          >
            <span>JOIN</span>
            <ConnectIcon />
          </button>
        </td>
      </tr>
      {open ? (
        <tr className="servers-info-details-ctr">
          <td></td>
          <td colSpan={4}>
            <table className="servers-info-details">
              <tbody>
                {details.map(([k, v]) => (
                  <tr key={k}>
                    <td className="servers-info-details-name">{k}</td>
                    <td className="servers-info-details-value">
                      {typeof v === 'boolean' ? (v ? 'yes' : 'no') : v}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            <ServerInfoBadge p={p} game={game} presence={presence} />
          </td>
        </tr>
      ) : null}
    </>
  )
}

type OrderBy = 'players'

function orderBy(o: OrderBy): (a: ServerInfo, b: ServerInfo) => boolean {
  switch (o) {
    case 'players':
      return ServerInfo.orderByPlayers
  }
}

export interface ServerListProps {
  gameID: number
  gameIdx: number
  games: GameInfo[]
  onGameSelect: (gameId: number) => void
}

function ServerList({ gameID, gameIdx, games, onGameSelect }: ServerListProps) {
  const [refresh, setRefresh] = useState(60000)
  const [order, _serOrder] = useState<OrderBy>('players')
  const addrs = useServers(gameID, refresh)
  const pr = usePresence()
  const servers = useServerInfos(addrs.type === 'ok' ? addrs.v : {}, refresh / 2, orderBy(order))
  const env = useEnv()
  useEffect(() => {
    if (env.type !== 'ok' || !env.v.demo) {
      return
    }
    const n = 3000
    if (refresh !== n) {
      setRefresh(n)
    }
  }, [refresh, env])

  if (addrs.type !== 'ok') {
    return addrs.alt
  }
  if (env.type !== 'ok') {
    return env.alt
  }
  if (pr.type !== 'ok') {
    return pr.alt
  }

  const game = games[gameIdx]
  const playerCount = servers.reduce((n, p) => n + p.players, 0)
  const serverList = env.v.demo ? servers.filter((p) => !p.restricted) : servers

  return (
    <table className="servers-ctr">
      <thead>
        <tr className="servers-info">
          <th className="servers-info-name-ctr" colSpan={2}>
            <div className="servers-info-name">
              <GamesMenu
                games={games}
                activeIdx={gameIdx}
                onSelect={onGameSelect}
                title={<span>Servers ({serverList.length})</span>}
              />
            </div>
          </th>
          <th>
            <span className="servers-info-players">Players ({playerCount})</span>
          </th>
          <th>
            <span className="servers-info-ping">Ping</span>
          </th>
          <th>
            <span className="servers-info-join"></span>
          </th>
        </tr>
      </thead>
      <tbody>
        {serverList.map((p) => (
          <ServerListInfo key={p.addr} game={game} p={p} presence={pr.v} />
        ))}
      </tbody>
    </table>
  )
}

interface GamesListProps {
  games: GameInfo[]
  activeIdx: number
  onSelect: (idx: number) => void
  title: ReactElement
}

function GamesMenu({ games, onSelect, activeIdx, title }: GamesListProps) {
  const active = games[activeIdx]
  const [menuOpen, setMenuOpen] = useState(false)

  return (
    <Menu
      className="games-list-menu"
      open={menuOpen}
      onToggle={setMenuOpen}
      hover={false}
      title={
        <div className="games-menu-title">
          <Game p={active} subtitle={title} />
        </div>
      }
      items={games.map((p, i) => ({
        body: (
          <Game
            p={p}
            onClick={() => {
              onSelect(i)
              setMenuOpen(false)
            }}
          />
        ),
      }))}
    />
  )
}

export interface ServersProps {
  gameIdx: number
  games: GameInfo[]
  onGameSelect: (gameId: number) => void
}

export default function Servers({ gameIdx, games, onGameSelect }: ServersProps) {
  const game = games[gameIdx]
  return (
    <div className="servers-ctr">
      <ServerList games={games} onGameSelect={onGameSelect} gameIdx={gameIdx} gameID={game.id} />
    </div>
  )
}
