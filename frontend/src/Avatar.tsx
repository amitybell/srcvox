import { Popover, PopoverProps } from '@mantine/core'
import cls from './Avatar.module.css'
import { openBrowser } from './api'
import { Profile } from './appstate'
import { useFetchDataURI } from './hooks/query'

function AvatarInfo(p: AvatarProps) {
  const id64 = BigInt('76561197960265728') + BigInt(p.userID)
  const scURL = `https://steamcommunity.com/profiles/${id64}`
  return (
    <div className={cls.info}>
      {p.userID > 0 ? (
        <a href={scURL} onClick={(e) => openBrowser(scURL, e)}>
          <Avatar {...p} />
        </a>
      ) : (
        <Avatar {...p} />
      )}
      <table className={cls.infoTable}>
        <tbody>
          <tr>
            <th>User Name:</th>
            <td>{p.username}</td>
          </tr>
          {p.userID > 0 ? (
            <>
              <tr>
                <th>Steam ID32:</th>
                <td>{p.userID > 0 ? p.userID : 'Unknown'}</td>
              </tr>
              <tr>
                <th>Steam ID64:</th>
                <td>{`${id64}`}</td>
              </tr>
            </>
          ) : null}
        </tbody>
      </table>
    </div>
  )
}

export interface AvatarProps extends Profile {
  className?: string
  embedded?: boolean
  title?: string
}

function Avatar({
  userID,
  title,
  username,
  avatarURI,
  avatarAlt,
  className,
  embedded,
}: AvatarProps) {
  const img = useFetchDataURI(`avatar(${avatarURI})`, embedded ? avatarURI : '')
  const src = img.type === 'ok' ? img.v : avatarURI
  return (
    <span key={userID} title={title} className={`${cls.avatar} ${className ?? ''}`}>
      {src ? <img alt={username} src={src} /> : <span>{avatarAlt}</span>}
    </span>
  )
}

export default function AvatarContainer({
  popover,
  ...props
}: AvatarProps & { popover?: PopoverProps }) {
  const body = (
    <span className={cls.root}>
      <Avatar {...props} />
    </span>
  )

  if (!popover) {
    return body
  }

  return (
    <Popover {...popover} withArrow={popover.withArrow ?? true}>
      <Popover.Target>{body}</Popover.Target>
      <Popover.Dropdown>
        <AvatarInfo {...props} />
      </Popover.Dropdown>
    </Popover>
  )
}
