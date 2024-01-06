import { useState } from 'react'
import Menu from './Menu'
import './Presence.css'
import { usePresence } from './hooks/query'
import Avatar from './Avatar'
import { Profile } from './appstate'
import { Tooltip, TooltipProps } from '@mantine/core'

export interface PresenceAvatarProps {
  className?: string
  tooltip?: Partial<TooltipProps>
}

export function PresenceAvatar({ className, tooltip }: PresenceAvatarProps) {
  const pr = usePresence()
  const p = new Profile(pr.type === 'ok' ? pr.v : {})
  const avatar = (
    <Avatar
      {...p}
      avatarURI={(() => {
        if (pr.type !== 'ok') {
          return ''
        }

        const { avatarURL, gameIconURI, inGame, humans } = pr.v
        const playing = inGame && humans.length !== 0
        return playing ? gameIconURI || avatarURL : avatarURL || gameIconURI
      })()}
    />
  )

  if (!tooltip) {
    return <span className={className}>avatar</span>
  }
  return (
    <Tooltip {...tooltip} label={p.username} className={className}>
      <span>{avatar}</span>
    </Tooltip>
  )
}

export default function Presence() {
  const [open, setOpen] = useState(false)
  const pr = usePresence()
  if (pr.type !== 'ok') {
    return (
      <div className="presence-ctr">
        {' '}
        {pr.type} {pr.alt}
      </div>
    )
  }

  const { clan, name, avatarURL, gameIconURI, inGame, humans } = pr.v
  const playing = inGame && humans.length !== 0
  const avatarSrc = playing ? gameIconURI || avatarURL : avatarURL || gameIconURI

  return (
    <Menu
      onToggle={setOpen}
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
        .filter(({ username }) => username !== pr.v.username)
        .sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()))
        .map((p) => ({
          key: p.id,
          onClick: () => {},
          body: (
            <div className="presence-human-profile">
              <Avatar {...p} />
              <span className="presence-human-profile-name">{p.username}</span>
            </div>
          ),
        }))}
    />
  )
}
