import './Soundboard.css'
import { ButtonHTMLAttributes, ReactNode, useState } from 'react'
import { PiPlayCircle as PlayIcon, PiStopCircleBold as StopIcon } from 'react-icons/pi'
import { SoundStatus, useSoundPlayer } from './hooks/useSoundPlayer'
import { useSounds } from './hooks/query'
import Spinner from './Spinner'

interface PlayButtonProps {
  status: SoundStatus
  text: string
  onClick?: (text: string) => void
  type?: ButtonHTMLAttributes<HTMLButtonElement>['type']
  children?: ReactNode
  className?: string
  title?: string
}

function PlayButton({ type, status, onClick, text, children, className, title }: PlayButtonProps) {
  return (
    <button
      title={title}
      type={type}
      className={`play-btn ${status} ${className || ''}`}
      onClick={() => onClick?.(text)}
    >
      {children}
      <span className="play-icon-ctr">
        {(() => {
          switch (status) {
            case 'stopped':
              return <PlayIcon />
            case 'loading':
              return <Spinner />
            case 'playing':
              return <StopIcon />
          }
        })()}
      </span>
    </button>
  )
}

interface SoundButtonProps {
  status: SoundStatus
  name: string
  onClick: (text: string) => void
}

function SoundButton({ name, onClick, status }: SoundButtonProps) {
  return (
    <PlayButton title={name} className="sound-ctr" text={name} onClick={onClick} status={status}>
      <span className="sound-title-ctr">{name}</span>
    </PlayButton>
  )
}

interface SoundboardMessageFormProps {
  message: string
  status: SoundStatus
  onChange: (message: string) => void
  onSubmit: (message: string) => void
}

function SoundboardMessageForm({
  message,
  onChange,
  onSubmit,
  status,
}: SoundboardMessageFormProps) {
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        onSubmit(message)
        return false
      }}
      className="soundboard-message-ctr"
    >
      <input
        name="text"
        placeholder="File name or message..."
        value={message}
        onChange={(e) => onChange(e.target.value)}
        type="text"
      />
      <PlayButton type="submit" text={message} status={status} />
    </form>
  )
}

export default function Soundboard() {
  const sp = useSoundPlayer()
  const sounds = useSounds()
  const [message, setMessage] = useState('')

  if (sounds.type !== 'ok') {
    return sounds.alt
  }

  const pat = message.trim()
  const filter = pat ? new RegExp(`${pat.split('').join('.*')}`, 'i') : null
  const soundsList = filter ? sounds.v.filter((p) => filter.test(p.name)) : sounds.v

  return (
    <div className="soundboard-ctr">
      {sp.embed}
      <SoundboardMessageForm
        message={message}
        onChange={setMessage}
        onSubmit={sp.toggle}
        status={!!message && message === sp.text ? sp.status : 'stopped'}
      />
      <div className="soundboard-list-ctr">
        {soundsList.map((p) => (
          <SoundButton
            key={p.name}
            status={p.name === sp.text ? sp.status : 'stopped'}
            onClick={sp.toggle}
            name={p.name}
          />
        ))}
      </div>
    </div>
  )
}
