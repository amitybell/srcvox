import cls from './Avatar.module.css'
import { Profile } from './appstate'
import { useFetchDataURI } from './hooks/query'

export interface AvatarProps extends Profile {
  className?: string
  embedded?: boolean
  title?: string
}

export default function Avatar({
  id,
  title,
  username,
  avatarURI,
  avatarAlt,
  className,
  embedded,
}: AvatarProps) {
  const img = useFetchDataURI(`avatar(${avatarURI})`, embedded ? avatarURI : '')
  avatarURI = img.type === 'ok' ? img.v : avatarURI
  return (
    <span key={id} title={title} className={`${cls.avatar} ${className ?? ''}`}>
      {avatarURI ? <img alt={username} src={avatarURI} /> : <span>{avatarAlt}</span>}
    </span>
  )
}
