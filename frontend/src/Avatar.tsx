import './Avatar.css'
import { Profile } from './appstate'

export default function Avatar({ id, name, avatarURI, avatarAlt }: Profile) {
  return (
    <span key={id} title={name} className="avatar-ctr">
      {avatarURI ? <img alt={name} src={avatarURI} /> : <span>{avatarAlt}</span>}
    </span>
  )
}
