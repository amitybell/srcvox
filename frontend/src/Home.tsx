import { useRef, useState } from 'react'
import './Home.css'
import { GameInfo } from './appstate'
import { useGames, useInGame } from './hooks/query'

function Game({ p: { iconURI, title } }: { p: GameInfo }) {
  return (
    <div className="game-ctr" title={title}>
      <img className="game-icon" src={iconURI} />
      <span className="game-title">{title}</span>
    </div>
  )
}

function PlayerOnline({ gameID }: { gameID: number }) {
  const r = useInGame({ gameID, refresh: 60000 })
  return <span>&nbsp;&mdash; {r.type === 'ok' ? <> {r.v.count} players online</> : r.alt}</span>
}

function Games() {
  const detailsRef = useRef<HTMLDetailsElement | null>(null)
  const [activeIdx, setAciveIdx] = useState(0)

  const r = useGames()
  if (r.type !== 'ok') {
    return r.alt
  }

  const games = r.v
  const active = games[activeIdx]

  return (
    <details ref={detailsRef} className="games-ctr">
      <summary>
        <Game p={active} />
        <PlayerOnline gameID={active.id} />
      </summary>
      <ul>
        {games.map((p, i) => (
          <li
            key={p.id}
            onClick={() => {
              setAciveIdx(i)
              if (detailsRef.current) {
                detailsRef.current.open = false
              }
            }}
          >
            <Game p={p} />
          </li>
        ))}
      </ul>
    </details>
  )
}

export default function Home() {
  return (
    <div className="home-ctr">
      <Games />
    </div>
  )
}
