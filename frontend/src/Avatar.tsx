import cls from './Avatar.module.css'
import { Profile } from './appstate'

export default function Avatar({ id, name, avatarURI, avatarAlt }: Profile) {
  return (
    <span key={id} title={name} className={cls.avatar}>
      {avatarURI ? <img alt={name} src={avatarURI} /> : <span>{avatarAlt}</span>}
    </span>
  )
}
