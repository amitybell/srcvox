import { useState } from 'react'
import { OrderBy, orderBy } from './Servers'
import { GameInfo } from './appstate'
import { usePresence, useServerInfos, useServers } from './hooks/query'
import Badges from './Badges'

export interface SnapshotsProps {
  game: GameInfo
}

export default function Snapshots({ game }: SnapshotsProps) {
  const refresh = 10000
  const [order, _serOrder] = useState<OrderBy>('players')
  const addrs = useServers(game.id, refresh)
  const pr = usePresence()
  const srvs = useServerInfos(
    addrs.type === 'ok' ? addrs.v : {},
    refresh / 2,
    orderBy(order),
  ).filter((p) => p.players > 0)

  if (addrs.type !== 'ok') {
    return addrs.alt
  }
  if (pr.type !== 'ok') {
    return pr.alt
  }

  return <Badges game={game} pr={pr.v} srvs={srvs} />
}
