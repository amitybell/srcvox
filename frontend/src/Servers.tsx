import './Servers.css'
import { ReactElement, useState } from 'react'
import Menu from './Menu'
import { GameInfo, ServerInfo } from './appstate'
import { useServerInfos } from './hooks/query'
import { openURL } from './api'
import {
  PiShieldStarFill as PublicServevrIcon,
  PiLockKeyOpenFill as PrivateServerIcon,
  PiArrowFatLinesRightFill as ConnectIcon,
} from 'react-icons/pi'

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

function ServerListInfo({ p: { name, players, addr, restricted, ping } }: { p: ServerInfo }) {
  return (
    <tr
      className={`servers-info ${players > 0 ? '' : 'empty'} ${restricted ? 'restricted' : ''}`}
      title={`${name} ${addr}`}
    >
      <td>
        <span className="servers-info-icon">
          {restricted ? <PrivateServerIcon /> : <PublicServevrIcon />}
        </span>
      </td>
      <td className="servers-info-name-ctr">
        <details>
          <summary className="servers-info-name">{name}</summary>
          <table>
            <tbody>
              <tr>
                <td>Address:</td>
                <td>{addr}</td>
              </tr>
            </tbody>
          </table>
        </details>
      </td>
      <td>
        <span className="servers-info-players">{players}</span>
      </td>
      <td>
        <span className="servers-info-ping">{ping}</span>
      </td>
      <td>
        <button
          className="servers-info-join button"
          onClick={() => openURL(`steam://connect/${addr}`)}
        >
          <span>JOIN</span>
          <ConnectIcon />
        </button>
      </td>
    </tr>
  )
}

export interface ServerListProps {
  gameID: number
  gameIdx: number
  games: GameInfo[]
  onGameSelect: (gameId: number) => void
}

function ServerList({ gameID, gameIdx, games, onGameSelect }: ServerListProps) {
  const servers = useServerInfos(gameID, 30000)

  if (servers.type !== 'ok') {
    return servers.alt
  }

  const playerCount = servers.v.reduce((n, p) => n + p.players, 0)

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
                title={<span>Servers ({servers.v.length})</span>}
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
        {servers.v.map((p) => (
          <ServerListInfo key={p.addr} p={p} />
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
