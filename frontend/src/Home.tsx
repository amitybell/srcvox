import { useRef, useState } from 'react'
import './Home.css'
import { GameInfo } from './appstate'
import { useGames } from './hooks/query'
import ServerList from './ServerList'

function Game({ p: { iconURI, title } }: { p: GameInfo }) {
  return (
    <div className="game-ctr" title={title}>
      <img className="game-icon" src={iconURI} />
      <span className="game-title">{title}</span>
    </div>
  )
}

interface GamesListProps {
  games: GameInfo[]
  activeIdx: number
  onSelect: (idx: number) => void
}

function GamesList({ games, onSelect, activeIdx }: GamesListProps) {
  const detailsRef = useRef<HTMLDetailsElement | null>(null)
  const active = games[activeIdx]

  return (
    <details ref={detailsRef} className="games-ctr">
      <summary>
        <Game p={active} />
      </summary>
      <ul>
        {games.map((p, i) => (
          <li
            key={p.id}
            onClick={() => {
              if (detailsRef.current) {
                detailsRef.current.open = false
              }
              onSelect(i)
            }}
          >
            <Game p={p} />
          </li>
        ))}
      </ul>
    </details>
  )
}

export interface HomeProps {
  gameIdx: number
  games: GameInfo[]
  onGameSelect: (gameId: number) => void
}

export default function Home({ gameIdx, games, onGameSelect }: HomeProps) {
  const game = games[gameIdx]
  return (
    <div className="home-ctr">
      <GamesList games={games} activeIdx={gameIdx} onSelect={onGameSelect} />
      <ServerList gameID={game.id} />
    </div>
  )
}
