import './Presence.css'
import cls from './Presence.module.css'

import { PopoverProps, Tooltip, TooltipProps, rem } from '@mantine/core'
import { IconFaceIdError as IconError } from '@tabler/icons-react'
import { useState } from 'react'
import Avatar from './Avatar'
import Menu from './Menu'
import { Presence as Pr, Profile } from './appstate'
import { usePresence } from './hooks/query'

export interface PresenceAvatarProps {
  className?: string
  tooltip?: Partial<TooltipProps>
  popover?: Partial<PopoverProps>
}

export function PresenceAvatar({ className, tooltip, popover }: PresenceAvatarProps) {
  const presence = usePresence()
  const pr = presence.type === 'ok' ? presence.v : new Pr({})
  const { avatarURL, gameIconURI, inGame, humans } = pr
  const playing = inGame && humans.length !== 0
  const avatarURI = playing ? gameIconURI || avatarURL : avatarURL || gameIconURI

  const avatar = pr.error ? (
    <IconError color="red" size={rem(32)} />
  ) : (
    <Avatar popover={popover} {...new Profile(pr)} avatarURI={avatarURI} />
  )
  className = `${className || ''} ${cls.root} ${pr.inGame ? cls.inGame : ''} ${pr.error ? 'error' : ''}`

  if (!tooltip) {
    return <span className={className}>{avatar}</span>
  }

  const tprops = {
    ...tooltip,
    className,
    label: pr.error ? `Error: ${pr.error}` : pr.username,
  }
  return (
    <Tooltip {...tprops}>
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
          key: p.userID,
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
