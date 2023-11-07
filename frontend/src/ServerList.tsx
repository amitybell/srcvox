import './ServerList.css'
import {
  PiShieldStarFill as PublicServevrIcon,
  PiLockKeyOpenFill as PrivateServerIcon,
  PiArrowFatLinesRightFill as ConnectIcon,
} from 'react-icons/pi'
import { ServerInfo } from './appstate'
import { useServerInfos } from './hooks/query'
import { openURL } from './api'

export interface ServerListProps {
  gameID: number
}

function ServerListInfo({ p: { name, players, addr, restricted, ping } }: { p: ServerInfo }) {
  return (
    <tr
      className={`serverlist-info ${players > 0 ? '' : 'empty'} ${restricted ? 'restricted' : ''}`}
      title={`${name} ${addr}`}
    >
      <td>
        <span className="serverlist-info-icon">
          {restricted ? <PrivateServerIcon /> : <PublicServevrIcon />}
        </span>
      </td>
      <td className="serverlist-info-name-ctr">
        <span className="serverlist-info-name">{name}</span>
      </td>
      <td>
        <span className="serverlist-info-players">{players}</span>
      </td>
      <td>
        <span className="serverlist-info-ping">{ping}</span>
      </td>
      <td>
        <button
          className="serverlist-info-join button"
          onClick={() => openURL(`steam://connect/${addr}`)}
        >
          <span>JOIN</span>
          <ConnectIcon />
        </button>
      </td>
    </tr>
  )
}

export default function ServerList({ gameID }: ServerListProps) {
  const servers = useServerInfos(gameID, 60000)

  if (servers.type !== 'ok') {
    return servers.alt
  }

  const playerCount = servers.v.reduce((n, p) => n + p.players, 0)

  return (
    <table className="serverlist-ctr">
      <thead>
        <tr className="serverlist-info">
          <th colSpan={2}>
            <span className="serverlist-info-name">Servers ({servers.v.length})</span>
          </th>
          <th>
            <span className="serverlist-info-players">Players ({playerCount})</span>
          </th>
          <th>
            <span className="serverlist-info-ping">Ping</span>
          </th>
          <th>
            <span className="serverlist-info-join"></span>
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
function sum(arg0: number[]) {
  throw new Error('Function not implemented.')
}
