import { ButtonHTMLAttributes, ReactNode, useState } from 'react'
import { PiPlayCircle as PlayIcon, PiStopCircleBold as StopIcon } from 'react-icons/pi'
import Highlight from './Highlight'
import './Soundboard.css'
import Spinner from './Spinner'
import { cmpLess, cmpLessPat } from './appstate'
import { useSounds } from './hooks/query'
import { SoundPlayerStatus, useSoundPlayer } from './hooks/useSoundPlayer'

interface PlayButtonProps {
  status: SoundPlayerStatus
  text: string
  onClick?: (text: string) => void
  type?: ButtonHTMLAttributes<HTMLButtonElement>['type']
  children?: ReactNode
  className?: string
  title?: string
  duration?: number
}

function PlayButton({
  type,
  status,
  onClick,
  text,
  children,
  className,
  title,
  duration,
}: PlayButtonProps) {
  // account for some delay between clicking the button and the sound starting
  // which can cause the progress to stop before it reaches 100%
  const soundStartDelay = 0.1
  const active = status === 'playing' && duration
  return (
    <button
      title={active ? `${title} (${duration.toFixed(1)}s)` : title}
      type={type}
      className={`play-btn ${status} ${className || ''}`}
      onClick={() => onClick?.(text)}
      style={
        active
          ? { backgroundSize: `100%`, transitionDuration: `${duration - soundStartDelay}s` }
          : undefined
      }
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
  status: SoundPlayerStatus
  name: string
  onClick: (text: string) => void
  pat: RegExp | null
  duration?: number
}

function SoundButton({ name, onClick, status, pat, duration }: SoundButtonProps) {
  return (
    <PlayButton
      title={name}
      className="sound-ctr"
      text={name}
      onClick={onClick}
      status={status}
      duration={duration}
    >
      <span className={`sound-title-ctr ${name.length > 8 ? 'long' : ''}`}>
        {pat ? (
          <Highlight className="sound-title" text={name} pat={pat} />
        ) : (
          <span className="sound-title">{name}</span>
        )}
      </span>
    </PlayButton>
  )
}

interface SoundboardMessageFormProps {
  message: string
  status: SoundPlayerStatus
  onChange: (message: string) => void
  onSubmit: (message: string) => void
  duration?: number
}

function SoundboardMessageForm({
  message,
  onChange,
  onSubmit,
  status,
  duration,
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
        placeholder="Sound name or text-to-speech message..."
        value={message}
        onChange={(e) => onChange(e.target.value)}
        type="text"
      />
      <PlayButton type="submit" text={message} status={status} duration={duration} />
    </form>
  )
}

export default function Soundboard() {
  const sp = useSoundPlayer()
  const sounds = useSounds()
  const [message, setMessage] = useState<string | null>(null)

  if (sounds.type !== 'ok') {
    return sounds.alt
  }

  const msg = message ?? ''
  const pat = msg.trim().toLowerCase()
  const filter = pat ? new RegExp(`${pat.split('').join('.*')}`, 'i') : null
  let soundsList = filter
    ? sounds.v.filter((p) => filter.test(p.sortName)).sort(cmpLessPat(pat))
    : sounds.v.sort(cmpLess)

  return (
    <div className="soundboard-ctr">
      {sp.embed}
      <SoundboardMessageForm
        message={msg}
        onChange={setMessage}
        onSubmit={sp.toggle}
        status={!!msg && msg === sp.text ? sp.status : 'stopped'}
        duration={sp.duration}
      />
      <div className="soundboard-list-ctr">
        {soundsList.map((p) => (
          <SoundButton
            key={p.name}
            status={p.name === sp.text ? sp.status : 'stopped'}
            onClick={sp.toggle}
            name={p.name}
            pat={filter}
            duration={sp.duration}
          />
        ))}
      </div>
    </div>
  )
}
