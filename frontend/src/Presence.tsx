import { useState } from 'react'
import Menu from './Menu'
import './Presence.css'
import { usePresence } from './hooks/query'

export default function Presence() {
  const [open, setOpen] = useState(false)
  const pr = usePresence()
  if (pr.type !== 'ok') {
    return <div className="presence-ctr">{pr.alt}</div>
  }

  const { clan, name, avatarURL, gameIconURI, inGame, username } = pr.v
  const playing = inGame && pr.v.humans.length !== 0
  const avatarSrc = playing ? gameIconURI || avatarURL : avatarURL || gameIconURI
  const humans = pr.v.humans.filter((name) => name !== username)

  if (!name) {
    return <div className="presence-ctr"></div>
  }

  return (
    <Menu
      onToggle={setOpen}
      hover={false}
      open={open}
      indicator
      title={
        <div className={`presence-ctr ${inGame ? 'in-game' : ''}`}>
          {avatarSrc ? <img alt="" className="presence-icon" src={avatarSrc} /> : null}
          <span className="presence-clan">{clan}</span>
          <span className="presence-name">{name}</span>
        </div>
      }
      items={humans
        .sort((a, b) => a.toLowerCase().localeCompare(b.toLowerCase()))
        .map((name) => ({ key: name, body: <span>{name}</span> }))}
    />
  )
}
